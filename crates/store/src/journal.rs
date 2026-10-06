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
use crate::error::{Error, Result, corrupt};
use crate::file::{be_u32, read_at};
use crate::hash::Hash;

/// JOURNAL_FILE is the journal's file name, which is also its name in the manifest.
pub const JOURNAL_FILE: &str = "vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv";
/// MAX_RECORD_LEN is the largest record the journal writer can write.
pub(crate) const MAX_RECORD_LEN: u32 = 5 * 1024 * 1024;
/// ROOT_RECORD_LEN is the length of a root hash record.
pub(crate) const ROOT_RECORD_LEN: usize = 4 + 2 + (1 + Hash::LEN) + (1 + 8) + CHECKSUM_LEN;
/// PAYLOAD_OFFSET is the offset of a chunk record's payload within the record.
pub(crate) const PAYLOAD_OFFSET: usize = 4 + 2 + (1 + Hash::LEN) + 1;

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
    timestamp: u64,
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
    let mut parsed = Record { kind: 0, address: Hash::default(), payload: &[], payload_at: 0, timestamp: 0 };
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
            TIMESTAMP_TAG => {
                let bytes = record.get(at..at + 8).ok_or_else(short)?;
                parsed.timestamp = u64::from_be_bytes(bytes.try_into().unwrap());
                at += 8;
            }
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

/// record_len_at returns the length of the record at the offset when a whole, valid record is there, as Dolt's
/// journal reader requires before it stops reading.
pub(crate) fn record_len_at(bytes: &[u8], at: usize) -> Option<usize> {
    let len = be_u32(bytes, at);
    (len != 0 && len <= MAX_RECORD_LEN && at + len as usize <= bytes.len() && valid(&bytes[at..]))
        .then_some(len as usize)
}

/// parse_record decodes a valid chunk or root hash record, returning it with the offset of its payload.
pub(crate) fn parse_record(raw: &[u8]) -> Result<(JournalRecord<'_>, usize)> {
    let record = parse(raw)?;
    match record.kind {
        ROOT_KIND => Ok((JournalRecord::Root { hash: record.address, timestamp: record.timestamp }, 0)),
        CHUNK_KIND => Ok((JournalRecord::Chunk { hash: record.address, record: record.payload }, record.payload_at)),
        kind => Err(corrupt(format!("unknown journal record kind ({kind})"))),
    }
}

/// data_loss_error returns the error for valid records that follow an invalid one at the offset.
pub(crate) fn data_loss_error(path: &Path, at: usize) -> Error {
    corrupt(format!(
        "possible data loss detected in journal file {} at offset {at}: corrupted journal\nplease run 'dolt fsck' to \
         assess the damage and attempt repairs",
        path.display()
    ))
}

/// possible_data_loss reports whether valid records follow a point where reading stopped: a root hash record followed
/// by any other record means the journal lost data that a writer synced.
pub(crate) fn possible_data_loss(rest: &[u8]) -> bool {
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
        while at + 4 <= bytes.len() {
            let Some(len) = record_len_at(&bytes, at) else { break };
            match parse_record(&bytes[at..at + len])? {
                (JournalRecord::Root { hash, .. }, _) => root = hash,
                (JournalRecord::Chunk { hash, record }, payload_at) => {
                    if chunks.insert(hash, ((at + payload_at) as u64, record.len() as u32)).is_none() {
                        order.push(hash);
                    }
                }
            }
            at += len;
        }
        if at < bytes.len() && possible_data_loss(&bytes[at..]) {
            return Err(data_loss_error(path, at));
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

/// JournalRecord is a record of the journal.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum JournalRecord<'a> {
    /// A chunk, with its compressed chunk record.
    Chunk { hash: Hash, record: &'a [u8] },
    /// A root hash, with the Unix time in seconds it was written at.
    Root { hash: Hash, timestamp: u64 },
}

impl JournalRecord<'_> {
    /// encode returns the record's bytes as Dolt's writeChunkRecord and writeRootHashRecord write them.
    pub fn encode(&self) -> Vec<u8> {
        let mut out = vec![0; 4];
        match self {
            JournalRecord::Chunk { hash, record } => {
                out.extend_from_slice(&[KIND_TAG, CHUNK_KIND, ADDRESS_TAG]);
                out.extend_from_slice(&hash.0);
                out.push(PAYLOAD_TAG);
                out.extend_from_slice(record);
            }
            JournalRecord::Root { hash, timestamp } => {
                out.extend_from_slice(&[KIND_TAG, ROOT_KIND, TIMESTAMP_TAG]);
                out.extend_from_slice(&timestamp.to_be_bytes());
                out.push(ADDRESS_TAG);
                out.extend_from_slice(&hash.0);
            }
        }
        let len = (out.len() + CHECKSUM_LEN) as u32;
        out[..4].copy_from_slice(&len.to_be_bytes());
        let checksum = crc(&out);
        out.extend_from_slice(&checksum.to_be_bytes());
        out
    }
}

/// read_records returns the valid records at the start of a journal's bytes, each with the bytes it was read from.
pub fn read_records(bytes: &[u8]) -> Result<Vec<(JournalRecord<'_>, &[u8])>> {
    let mut records = Vec::new();
    let mut at = 0;
    while at + 4 <= bytes.len() {
        let Some(len) = record_len_at(bytes, at) else { break };
        let raw = &bytes[at..at + len];
        records.push((parse_record(raw)?.0, raw));
        at += len;
    }
    Ok(records)
}
