// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//! The chunk journal, an append-only file of records. A record is its u32 length (counting the whole record), then
//! tagged fields, then the CRC-32C of everything before it. The fields are a kind (1 for a root hash, 2 for a chunk),
//! an address, a u64 timestamp in Unix seconds for root hashes, and for chunks a payload that runs to the checksum and
//! holds a compressed chunk record. Every integer is big-endian.

use std::collections::HashMap;
use std::fs::File;
use std::path::{Path, PathBuf};

use crate::chunk::{CHECKSUM_LEN, Chunk, crc};
use crate::error::{Result, corrupt};
use crate::file::{be_u32, read_at};
use crate::hash::Hash;

/// JOURNAL_FILE is the journal's file name, which is also its name in the manifest.
pub const JOURNAL_FILE: &str = "vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv";
/// MAX_RECORD_LEN is the largest record the journal writer can write.
const MAX_RECORD_LEN: u32 = 5 * 1024 * 1024;
/// ROOT_RECORD_LEN is the length of a root hash record.
const ROOT_RECORD_LEN: usize = 4 + 2 + (1 + Hash::LEN) + (1 + 8) + CHECKSUM_LEN;

/// Record tags and kinds.
const KIND_TAG: u8 = 1;
const ADDRESS_TAG: u8 = 2;
const PAYLOAD_TAG: u8 = 3;
const TIMESTAMP_TAG: u8 = 4;
const ROOT_KIND: u8 = 1;
const CHUNK_KIND: u8 = 2;

/// Record is a decoded journal record.
struct Record<'a> {
    kind: u8,
    address: Hash,
    payload: &'a [u8],
    /// The offset of the payload within the record.
    payload_at: usize,
}

/// valid reports whether the bytes start with a whole record whose length and checksum are right.
fn valid(bytes: &[u8]) -> bool {
    if bytes.len() < 4 + CHECKSUM_LEN {
        return false;
    }
    let len = be_u32(bytes, 0) as usize;
    len >= 4 + CHECKSUM_LEN
        && len <= bytes.len()
        && crc(&bytes[..len - CHECKSUM_LEN]) == be_u32(bytes, len - CHECKSUM_LEN)
}

/// parse decodes a valid record.
fn parse(record: &[u8]) -> Result<Record<'_>> {
    let mut parsed = Record { kind: 0, address: Hash::default(), payload: &[], payload_at: 0 };
    let end = record.len() - CHECKSUM_LEN;
    let mut at = 4;
    let short = || corrupt("journal record field overruns the record");
    while at < end {
        let tag = record[at];
        at += 1;
        match tag {
            KIND_TAG => {
                parsed.kind = *record.get(at).ok_or_else(short)?;
                at += 1;
            }
            ADDRESS_TAG => {
                let bytes = record.get(at..at + Hash::LEN).ok_or_else(short)?;
                parsed.address = Hash(bytes.try_into().unwrap());
                at += Hash::LEN;
            }
            TIMESTAMP_TAG => at += 8,
            PAYLOAD_TAG => {
                parsed.payload = &record[at..end];
                parsed.payload_at = at;
                at = end;
            }
            _ => return Err(corrupt(format!("unknown record field tag: {tag}"))),
        }
    }
    Ok(parsed)
}

/// possible_data_loss reports whether valid records follow a point where reading stopped: a root hash record followed
/// by any other record means the journal lost data that a writer synced.
fn possible_data_loss(rest: &[u8]) -> bool {
    let mut first_root = false;
    let mut at = 0;
    while at + ROOT_RECORD_LEN <= rest.len() {
        let len = be_u32(rest, at);
        if len > 0 && len <= MAX_RECORD_LEN && len as usize <= rest.len() - at && valid(&rest[at..]) {
            let record = &rest[at..at + len as usize];
            if first_root {
                return true;
            }
            if parse(record).is_ok_and(|r| r.kind == ROOT_KIND) {
                first_root = true;
            }
            at += len as usize;
            continue;
        }
        at += 1;
    }
    false
}

/// Journal reads the chunks and root hashes in a chunk journal.
pub struct Journal {
    file: File,
    /// The offset and length of each chunk's compressed record within the file.
    chunks: HashMap<Hash, (u64, u32)>,
    /// The order the chunks were written in.
    order: Vec<Hash>,
    /// The last root hash written, which is empty when none was.
    pub root: Hash,
}

impl Journal {
    /// open reads every record of the journal, stopping at the end of the valid records as the Go journal does, and
    /// failing when valid records follow an invalid one.
    pub fn open(path: &Path) -> Result<Journal> {
        let bytes = std::fs::read(path)?;
        let mut chunks = HashMap::new();
        let mut order = Vec::new();
        let mut root = Hash::default();
        let mut at = 0;
        let mut recovered = false;
        while at + 4 <= bytes.len() {
            let len = be_u32(&bytes, at);
            if len == 0 || len > MAX_RECORD_LEN || at + len as usize > bytes.len() || !valid(&bytes[at..]) {
                recovered = true;
                break;
            }
            let record = parse(&bytes[at..at + len as usize])?;
            match record.kind {
                ROOT_KIND => root = record.address,
                CHUNK_KIND => {
                    let offset = (at + record.payload_at) as u64;
                    if chunks.insert(record.address, (offset, record.payload.len() as u32)).is_none() {
                        order.push(record.address);
                    }
                }
                _ => {}
            }
            at += len as usize;
        }
        if recovered && possible_data_loss(&bytes[at..]) {
            return Err(corrupt(format!(
                "possible data loss detected in journal file {} at offset {at}: corrupted journal\nplease run 'dolt \
                 fsck' to assess the damage and attempt repairs",
                path.display()
            )));
        }
        Ok(Journal { file: File::open(path)?, chunks, order, root })
    }

    /// path returns the journal's path in a noms directory.
    pub fn path(dir: &Path) -> PathBuf {
        dir.join(JOURNAL_FILE)
    }

    /// has reports whether the journal holds the chunk.
    pub fn has(&self, hash: &Hash) -> bool {
        self.chunks.contains_key(hash)
    }

    /// get returns the chunk when the journal holds it.
    pub fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        match self.chunks.get(hash) {
            Some(&(offset, len)) => Chunk::from_record(*hash, &read_at(&self.file, offset, len as usize)?).map(Some),
            None => Ok(None),
        }
    }

    /// for_each_record calls the function with the address and compressed record of every chunk in the order they
    /// were written.
    pub fn for_each_record(&self, f: &mut dyn FnMut(Hash, &[u8]) -> Result<()>) -> Result<()> {
        for hash in &self.order {
            let (offset, len) = self.chunks[hash];
            f(*hash, &read_at(&self.file, offset, len as usize)?)?;
        }
        Ok(())
    }

    /// for_each calls the function with every chunk in the order they were written.
    pub fn for_each(&self, f: &mut dyn FnMut(Chunk) -> Result<()>) -> Result<()> {
        for hash in &self.order {
            f(self.get(hash)?.unwrap())?;
        }
        Ok(())
    }
}
