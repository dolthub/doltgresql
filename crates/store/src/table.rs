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

//! NBS table files. A table file is a sequence of chunk records followed by an index and a footer:
//!
//! - A chunk record is snappy-compressed chunk data followed by the CRC-32C of the compressed data.
//! - The index is a prefix map of (8-byte hash prefix, u32 ordinal) tuples sorted by prefix, then the u32 length of
//!   each record in file order, then the 12-byte hash suffix of each record in file order.
//! - The footer is the u32 chunk count, the u64 total uncompressed data size, and an 8-byte magic number.
//!
//! Every integer is big-endian.

use std::fs::File;
use std::path::Path;

use crate::chunk::{CHECKSUM_LEN, Chunk};
use crate::error::{Result, corrupt};
use crate::file::{be_u32, be_u64, read_at};
use crate::hash::Hash;

/// MAGIC ends every table file: the first 8 bytes of the SHA-256 of "https://github.com/attic-labs/nbs".
const MAGIC: &[u8; 8] = b"\xff\xb5\xd8\xc2\x24\x63\xee\x50";
/// ARCHIVE_SIGNATURE ends an archive, which callers must open as one.
const ARCHIVE_SIGNATURE: &[u8; 7] = b"DOLTARC";
/// FOOTER_LEN is the length of the footer.
const FOOTER_LEN: usize = 4 + 8 + 8;
/// PREFIX_TUPLE_LEN is the length of a prefix map entry.
const PREFIX_TUPLE_LEN: usize = Hash::PREFIX_LEN + 4;

/// index_len returns the length of the index of a table with the chunk count.
fn index_len(count: u64) -> u64 {
    count * (PREFIX_TUPLE_LEN as u64 + 4 + Hash::SUFFIX_LEN as u64)
}

/// TableReader reads the chunks of a table file, keeping its index in memory.
pub struct TableReader {
    file: File,
    /// The hash prefix of each prefix map entry, in prefix order.
    prefixes: Vec<u64>,
    /// The ordinal of each prefix map entry, in prefix order.
    ordinals: Vec<u32>,
    /// The offset of each record, in ordinal order, with the file's data length at the end.
    offsets: Vec<u64>,
    /// The hash suffix of each record, in ordinal order.
    suffixes: Vec<u8>,
}

impl TableReader {
    /// open reads the index of a table file.
    pub fn open(path: &Path) -> Result<TableReader> {
        let file = File::open(path)?;
        let size = file.metadata()?.len();
        if size < FOOTER_LEN as u64 {
            return Err(corrupt(format!("{} is too short for a table file", path.display())));
        }
        let footer = read_at(&file, size - FOOTER_LEN as u64, FOOTER_LEN)?;
        if &footer[12..] != MAGIC {
            if &footer[FOOTER_LEN - ARCHIVE_SIGNATURE.len()..] == ARCHIVE_SIGNATURE {
                return Err(corrupt("unsupported table file format"));
            }
            return Err(corrupt("invalid or corrupt table file"));
        }
        let count = be_u32(&footer, 0) as u64;
        let index_len = index_len(count);
        if index_len + FOOTER_LEN as u64 > size {
            return Err(corrupt(format!("{} is too short for its index", path.display())));
        }
        let index = read_at(&file, size - FOOTER_LEN as u64 - index_len, index_len as usize)?;
        let count = count as usize;
        let lengths_at = count * PREFIX_TUPLE_LEN;
        let suffixes_at = lengths_at + count * 4;
        let mut prefixes = Vec::with_capacity(count);
        let mut ordinals = Vec::with_capacity(count);
        for i in 0..count {
            prefixes.push(be_u64(&index, i * PREFIX_TUPLE_LEN));
            ordinals.push(be_u32(&index, i * PREFIX_TUPLE_LEN + Hash::PREFIX_LEN));
        }
        let mut offsets = Vec::with_capacity(count + 1);
        let mut offset = 0u64;
        offsets.push(0);
        for i in 0..count {
            let length = be_u32(&index, lengths_at + i * 4) as u64;
            if length <= CHECKSUM_LEN as u64 {
                return Err(corrupt(format!("{}: chunk record {i} is too short", path.display())));
            }
            offset += length;
            offsets.push(offset);
        }
        let reader = TableReader { file, prefixes, ordinals, offsets, suffixes: index[suffixes_at..].to_vec() };
        reader.validate(path, size - FOOTER_LEN as u64 - index_len)?;
        Ok(reader)
    }

    /// validate checks that the prefixes are sorted, the ordinals are in range, and the records fit before the index.
    fn validate(&self, path: &Path, data_len: u64) -> Result<()> {
        if self.prefixes.windows(2).any(|pair| pair[0] > pair[1]) {
            return Err(corrupt(format!("{}: table file index prefixes are not sorted", path.display())));
        }
        if self.ordinals.iter().any(|&ordinal| ordinal as usize >= self.ordinals.len()) {
            return Err(corrupt(format!("{}: table file index ordinal out of range", path.display())));
        }
        if *self.offsets.last().unwrap() > data_len {
            return Err(corrupt(format!("{}: table file records overrun the index", path.display())));
        }
        Ok(())
    }

    /// count returns the number of chunks.
    pub fn count(&self) -> usize {
        self.ordinals.len()
    }

    /// suffix returns the hash suffix of the record with the ordinal.
    fn suffix(&self, ordinal: usize) -> &[u8] {
        &self.suffixes[ordinal * Hash::SUFFIX_LEN..(ordinal + 1) * Hash::SUFFIX_LEN]
    }

    /// find returns the ordinal of the chunk's record.
    pub fn find(&self, hash: &Hash) -> Option<usize> {
        let prefix = hash.prefix();
        let start = self.prefixes.partition_point(|&p| p < prefix);
        (start..self.prefixes.len())
            .take_while(|&i| self.prefixes[i] == prefix)
            .map(|i| self.ordinals[i] as usize)
            .find(|&ordinal| self.suffix(ordinal) == hash.suffix())
    }

    /// has reports whether the table holds the chunk.
    pub fn has(&self, hash: &Hash) -> bool {
        self.find(hash).is_some()
    }

    /// get returns the chunk when the table holds it.
    pub fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        match self.find(hash) {
            Some(ordinal) => self.read(*hash, ordinal).map(Some),
            None => Ok(None),
        }
    }

    /// read reads and decompresses the record with the ordinal.
    fn read(&self, hash: Hash, ordinal: usize) -> Result<Chunk> {
        let start = self.offsets[ordinal];
        let record = read_at(&self.file, start, (self.offsets[ordinal + 1] - start) as usize)?;
        Chunk::from_record(hash, &record)
    }

    /// hashes returns the address of each chunk in file order.
    fn hashes(&self) -> Vec<Hash> {
        let mut hashes = vec![Hash::default(); self.count()];
        for (&prefix, &ordinal) in self.prefixes.iter().zip(&self.ordinals) {
            hashes[ordinal as usize] = Hash::from_parts(prefix, self.suffix(ordinal as usize));
        }
        hashes
    }

    /// for_each calls the function with every chunk in file order.
    pub fn for_each(&self, f: &mut dyn FnMut(Chunk) -> Result<()>) -> Result<()> {
        for (ordinal, hash) in self.hashes().into_iter().enumerate() {
            f(self.read(hash, ordinal)?)?;
        }
        Ok(())
    }

    /// for_each_record calls the function with the address and compressed record of every chunk in file order.
    pub fn for_each_record(&self, f: &mut dyn FnMut(Hash, &[u8]) -> Result<()>) -> Result<()> {
        for (ordinal, hash) in self.hashes().into_iter().enumerate() {
            let start = self.offsets[ordinal];
            f(hash, &read_at(&self.file, start, (self.offsets[ordinal + 1] - start) as usize)?)?;
        }
        Ok(())
    }
}
