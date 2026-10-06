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

//! Dolt archives (`.darc` files). An archive is a sequence of byte spans followed by an index, JSON metadata, and a
//! footer. Each chunk refers to a data span and, when it is zstd-compressed with a dictionary, a dictionary span whose
//! contents are themselves zstd-compressed. Version 2 adds snappy chunks, stored like table file records with no
//! dictionary, and version 3 widens the index length in the footer from 32 to 64 bits.
//!
//! The index is the u64 end offset of each span, then the u64 hash prefix of each chunk in sorted order, then the u32
//! dictionary and data span ids of each chunk, then the 12-byte hash suffix of each chunk. Span ids count from 1, and 0
//! means no span. The footer is the index length, the u32 span count, the u32 chunk count, the u32 metadata length,
//! 192 unused checksum bytes, the u8 format version, and the signature "DOLTARC". Every integer is big-endian.

use std::collections::HashMap;
use std::fs::File;
use std::path::Path;
use std::sync::{Arc, Mutex};

use crate::chunk::Chunk;
use crate::error::{Result, corrupt};
use crate::file::{be_u32, be_u64, read_at};
use crate::hash::Hash;

/// SIGNATURE ends every archive.
const SIGNATURE: &[u8; 7] = b"DOLTARC";
/// FOOTER_LEN is the length of a version 3 footer, and version 1 and 2 footers are 4 bytes shorter.
const FOOTER_LEN: usize = 8 + 4 + 4 + 4 + 192 + 1 + SIGNATURE.len();
/// MAX_VERSION is the newest format version.
const MAX_VERSION: u8 = 3;
/// SNAPPY_VERSION is the first version that stores snappy chunks.
const SNAPPY_VERSION: u8 = 2;

/// ArchiveReader reads the chunks of an archive, keeping its index in memory.
pub struct ArchiveReader {
    file: File,
    version: u8,
    /// The end offset of each span, in span id order starting from id 1.
    span_ends: Vec<u64>,
    /// The hash prefix of each chunk, sorted.
    prefixes: Vec<u64>,
    /// The dictionary and data span ids of each chunk.
    refs: Vec<(u32, u32)>,
    /// The hash suffix of each chunk.
    suffixes: Vec<u8>,
    /// The decompressed dictionaries read so far, by span id.
    dictionaries: Mutex<HashMap<u32, Arc<Vec<u8>>>>,
}

impl ArchiveReader {
    /// open reads the footer and index of an archive.
    pub fn open(path: &Path) -> Result<ArchiveReader> {
        let file = File::open(path)?;
        let size = file.metadata()?.len();
        if size < FOOTER_LEN as u64 {
            return Err(corrupt(format!("{} is too short for an archive", path.display())));
        }
        let footer = read_at(&file, size - FOOTER_LEN as u64, FOOTER_LEN)?;
        if &footer[FOOTER_LEN - SIGNATURE.len()..] != SIGNATURE {
            return Err(corrupt("invalid file signature"));
        }
        let version = footer[FOOTER_LEN - SIGNATURE.len() - 1];
        if version > MAX_VERSION {
            return Err(corrupt("invalid format version"));
        }
        let (index_len, footer_len) =
            if version < 3 { (be_u32(&footer, 4) as u64, FOOTER_LEN - 4) } else { (be_u64(&footer, 0), FOOTER_LEN) };
        let span_count = be_u32(&footer, 8) as usize;
        let chunk_count = be_u32(&footer, 12) as usize;
        let metadata_len = be_u32(&footer, 16) as u64;
        let expected = (span_count * 8 + chunk_count * (8 + 8 + Hash::SUFFIX_LEN)) as u64;
        if index_len != expected {
            return Err(corrupt(format!("{}: corrupt archive index", path.display())));
        }
        let Some(index_at) = size.checked_sub(footer_len as u64 + metadata_len + index_len) else {
            return Err(corrupt(format!("{}: corrupt archive index", path.display())));
        };
        let index = read_at(&file, index_at, index_len as usize)?;
        let span_ends: Vec<u64> = (0..span_count).map(|i| be_u64(&index, i * 8)).collect();
        let prefixes_at = span_count * 8;
        let prefixes: Vec<u64> = (0..chunk_count).map(|i| be_u64(&index, prefixes_at + i * 8)).collect();
        let refs_at = prefixes_at + chunk_count * 8;
        let refs: Vec<(u32, u32)> =
            (0..chunk_count).map(|i| (be_u32(&index, refs_at + i * 8), be_u32(&index, refs_at + i * 8 + 4))).collect();
        let suffixes = index[refs_at + chunk_count * 8..].to_vec();
        if prefixes.windows(2).any(|pair| pair[0] > pair[1])
            || span_ends.windows(2).any(|pair| pair[0] > pair[1])
            || span_ends.last().is_some_and(|&end| end > index_at)
            || refs
                .iter()
                .any(|&(dictionary, data)| data == 0 || data as usize > span_count || dictionary as usize > span_count)
        {
            return Err(corrupt(format!("{}: corrupt archive index", path.display())));
        }
        Ok(ArchiveReader {
            file,
            version,
            span_ends,
            prefixes,
            refs,
            suffixes,
            dictionaries: Mutex::new(HashMap::new()),
        })
    }

    /// count returns the number of chunks.
    pub fn count(&self) -> usize {
        self.prefixes.len()
    }

    /// suffix returns the hash suffix of the chunk with the id.
    fn suffix(&self, id: usize) -> &[u8] {
        &self.suffixes[id * Hash::SUFFIX_LEN..(id + 1) * Hash::SUFFIX_LEN]
    }

    /// find returns the id of the chunk.
    pub fn find(&self, hash: &Hash) -> Option<usize> {
        let prefix = hash.prefix();
        let start = self.prefixes.partition_point(|&p| p < prefix);
        (start..self.prefixes.len())
            .take_while(|&id| self.prefixes[id] == prefix)
            .find(|&id| self.suffix(id) == hash.suffix())
    }

    /// has reports whether the archive holds the chunk.
    pub fn has(&self, hash: &Hash) -> bool {
        self.find(hash).is_some()
    }

    /// get returns the chunk when the archive holds it.
    pub fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        match self.find(hash) {
            Some(id) => self.read(*hash, id).map(Some),
            None => Ok(None),
        }
    }

    /// span reads the span with the id.
    fn span(&self, id: u32) -> Result<Vec<u8>> {
        let index = id as usize - 1;
        let start = if index == 0 { 0 } else { self.span_ends[index - 1] };
        read_at(&self.file, start, (self.span_ends[index] - start) as usize)
    }

    /// dictionary returns the decompressed dictionary in the span with the id.
    fn dictionary(&self, id: u32) -> Result<Arc<Vec<u8>>> {
        if let Some(dictionary) = self.dictionaries.lock().unwrap().get(&id) {
            return Ok(dictionary.clone());
        }
        let dictionary = Arc::new(
            zstd::stream::decode_all(self.span(id)?.as_slice())
                .map_err(|err| corrupt(format!("cannot decompress archive dictionary: {err}")))?,
        );
        self.dictionaries.lock().unwrap().insert(id, dictionary.clone());
        Ok(dictionary)
    }

    /// read reads and decompresses the chunk with the id.
    fn read(&self, hash: Hash, id: usize) -> Result<Chunk> {
        let (dictionary, data) = self.refs[id];
        let compressed = self.span(data)?;
        if dictionary == 0 {
            if self.version < SNAPPY_VERSION {
                return Err(corrupt("runtime error: unable to get archived chunk. dictionary is nil"));
            }
            return Chunk::from_record(hash, &compressed);
        }
        let dictionary = self.dictionary(dictionary)?;
        let mut data = Vec::new();
        zstd::stream::read::Decoder::with_dictionary(compressed.as_slice(), &dictionary)
            .and_then(|mut decoder| std::io::Read::read_to_end(&mut decoder, &mut data))
            .map_err(|err| corrupt(format!("cannot decompress archived chunk {hash}: {err}")))?;
        Ok(Chunk { hash, data })
    }

    /// hashes returns the address of every chunk, in index order.
    pub fn hashes(&self) -> Vec<Hash> {
        (0..self.count()).map(|id| Hash::from_parts(self.prefixes[id], self.suffix(id))).collect()
    }

    /// for_each calls the function with every chunk in index order.
    pub fn for_each(&self, f: &mut dyn FnMut(Chunk) -> Result<()>) -> Result<()> {
        for (id, hash) in self.hashes().into_iter().enumerate() {
            f(self.read(hash, id)?)?;
        }
        Ok(())
    }
}
