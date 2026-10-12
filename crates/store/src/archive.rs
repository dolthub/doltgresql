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
use crate::file::{ReadAt, be_u32, be_u64};
use crate::hash::Hash;

/// SIGNATURE ends every archive.
const SIGNATURE: &[u8; 7] = b"DOLTARC";
/// FOOTER_LEN is the length of a version 3 footer, and version 1 and 2 footers are 4 bytes shorter.
const FOOTER_LEN: usize = 8 + 4 + 4 + 4 + 192 + 1 + SIGNATURE.len();
/// MAX_VERSION is the newest format version.
const MAX_VERSION: u8 = 3;
/// SNAPPY_VERSION is the first version that stores snappy chunks.
const SNAPPY_VERSION: u8 = 2;

/// Dictionary is a dictionary span of an archive as it is stored, and the dictionary prepared for decompressing.
type Dictionary = (Arc<Vec<u8>>, Arc<zstd::dict::DecoderDictionary<'static>>);

/// Stored is a chunk as an archive stores it: its data compressed with a dictionary, and that dictionary's span as the
/// archive stores it, which another archive can take over without compressing the chunk again.
pub struct Stored {
    dictionary: Arc<Vec<u8>>,
    data: Vec<u8>,
}

/// ArchiveReader reads the chunks of an archive, keeping its index in memory.
pub struct ArchiveReader {
    file: Box<dyn ReadAt>,
    /// The file's name.
    name: String,
    version: u8,
    /// The end offset of each span, in span id order starting from id 1.
    span_ends: Vec<u64>,
    /// The hash prefix of each chunk, sorted.
    prefixes: Vec<u64>,
    /// The dictionary and data span ids of each chunk.
    refs: Vec<(u32, u32)>,
    /// The hash suffix of each chunk.
    suffixes: Vec<u8>,
    /// The dictionaries read so far, by span id: each span as stored, and the dictionary prepared for decompressing.
    dictionaries: Mutex<HashMap<u32, Dictionary>>,
}

impl ArchiveReader {
    /// open reads the footer and index of an archive.
    pub fn open(path: &Path) -> Result<ArchiveReader> {
        let name = path.file_name().map(|n| n.to_string_lossy().into_owned()).unwrap_or_default();
        ArchiveReader::open_reader(Box::new(File::open(path)?), name, &path.display().to_string())
    }

    /// open_reader reads the index of a archive that a reader reads, with its name and the name that errors show.
    pub fn open_reader(file: Box<dyn ReadAt>, name: String, shown: &str) -> Result<ArchiveReader> {
        let size = file.size()?;
        if size < FOOTER_LEN as u64 {
            return Err(corrupt(format!("{shown} is too short for an archive")));
        }
        let footer = file.read_at(size - FOOTER_LEN as u64, FOOTER_LEN)?;
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
            return Err(corrupt(format!("{}: corrupt archive index", shown)));
        }
        let Some(index_at) = size.checked_sub(footer_len as u64 + metadata_len + index_len) else {
            return Err(corrupt(format!("{}: corrupt archive index", shown)));
        };
        let index = file.read_at(index_at, index_len as usize)?;
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
            return Err(corrupt(format!("{}: corrupt archive index", shown)));
        }
        Ok(ArchiveReader {
            file,
            name,
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

    /// get_stored returns the chunk when the archive holds it, with its stored form when a dictionary compressed it.
    pub fn get_stored(&self, hash: &Hash) -> Result<Option<(Chunk, Option<Stored>)>> {
        match self.find(hash) {
            Some(id) => self.read_stored(*hash, id).map(Some),
            None => Ok(None),
        }
    }

    /// span_range returns the offset and length of the span with the id.
    fn span_range(&self, id: u32) -> (u64, u32) {
        let index = id as usize - 1;
        let start = if index == 0 { 0 } else { self.span_ends[index - 1] };
        (start, (self.span_ends[index] - start) as u32)
    }

    /// locate returns where the chunk's data span is, and its dictionary's span, when the archive holds it.
    pub fn locate(&self, hash: &Hash) -> Option<crate::Location> {
        let (dictionary, data) = self.refs[self.find(hash)?];
        let (offset, length) = self.span_range(data);
        let dictionary = (dictionary != 0).then(|| self.span_range(dictionary));
        Some(crate::Location { file: self.name.clone(), offset, length, dictionary })
    }

    /// span reads the span with the id.
    fn span(&self, id: u32) -> Result<Vec<u8>> {
        let index = id as usize - 1;
        let start = if index == 0 { 0 } else { self.span_ends[index - 1] };
        self.file.read_at(start, (self.span_ends[index] - start) as usize)
    }

    /// dictionary returns the dictionary in the span with the id, prepared for decompressing.
    fn dictionary(&self, id: u32) -> Result<Dictionary> {
        if let Some(dictionary) = self.dictionaries.lock().unwrap().get(&id) {
            return Ok(dictionary.clone());
        }
        let span = self.span(id)?;
        let bytes = zstd::stream::decode_all(span.as_slice())
            .map_err(|err| corrupt(format!("cannot decompress archive dictionary: {err}")))?;
        let dictionary = (Arc::new(span), Arc::new(zstd::dict::DecoderDictionary::copy(&bytes)));
        self.dictionaries.lock().unwrap().insert(id, dictionary.clone());
        Ok(dictionary)
    }

    /// read reads and decompresses the chunk with the id.
    fn read(&self, hash: Hash, id: usize) -> Result<Chunk> {
        self.read_stored(hash, id).map(|(chunk, _)| chunk)
    }

    /// read_stored reads and decompresses the chunk with the id, returning its stored form too when a dictionary
    /// compressed it.
    fn read_stored(&self, hash: Hash, id: usize) -> Result<(Chunk, Option<Stored>)> {
        let (dictionary, data) = self.refs[id];
        let compressed = self.span(data)?;
        if dictionary == 0 {
            if self.version < SNAPPY_VERSION {
                return Err(corrupt("runtime error: unable to get archived chunk. dictionary is nil"));
            }
            return Ok((Chunk::from_record(hash, &compressed)?, None));
        }
        let (span, dictionary) = self.dictionary(dictionary)?;
        let mut data = Vec::new();
        zstd::stream::read::Decoder::with_prepared_dictionary(compressed.as_slice(), &dictionary)
            .and_then(|mut decoder| std::io::Read::read_to_end(&mut decoder, &mut data))
            .map_err(|err| corrupt(format!("cannot decompress archived chunk {hash}: {err}")))?;
        Ok((Chunk { hash, data }, Some(Stored { dictionary: span, data: compressed })))
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

/// MAX_SAMPLES is how many chunks an archive writer collects before it trains a dictionary from them, as Dolt's
/// maxSamples.
const MAX_SAMPLES: usize = 1000;
/// DICTIONARY_SIZE is the size of the dictionary an archive writer trains, as Dolt's defaultDictionarySize.
const DICTIONARY_SIZE: usize = 1 << 12;
/// MIN_SAMPLES_LEN is how many sample bytes dictionary training needs, as ZDICT_DICTSIZE_MIN.
const MIN_SAMPLES_LEN: usize = 256;
/// LEVEL is the zstd compression level, gozstd's default.
const LEVEL: i32 = 3;
/// DOLT_VERSION is the version of Dolt's storage library that the Go server embeds, which archives record.
const DOLT_VERSION: &str = "2.4.1";

/// ArchiveWriter builds an archive in memory as Dolt's ArchiveStreamWriter does: it holds the first chunks back until
/// it has enough to train a dictionary, compresses every chunk with that dictionary from then on, and stores the
/// chunks of an archive too small to train one as snappy records.
#[derive(Default)]
pub struct ArchiveWriter {
    buf: Vec<u8>,
    span_ends: Vec<u64>,
    /// Each chunk's address with its dictionary and data span ids.
    chunks: Vec<(Hash, u32, u32)>,
    /// The chunks held back until a dictionary is trained.
    queue: Vec<Chunk>,
    /// The span id of the trained dictionary, with a compressor that uses it.
    dictionary: Option<(u32, zstd::bulk::Compressor<'static>)>,
    /// How many bytes the writer moved to a spill file, which come before `buf`.
    spilled: u64,
    /// The span ids of the dictionaries that chunks taken over from other archives use, by the address of the
    /// dictionaries' shared bytes, which the writer holds on to so that no other dictionary takes that address.
    taken: HashMap<usize, (u32, Arc<Vec<u8>>)>,
}

impl ArchiveWriter {
    pub fn new() -> ArchiveWriter {
        ArchiveWriter::default()
    }

    /// count returns the number of chunks added.
    pub fn count(&self) -> usize {
        self.chunks.len() + self.queue.len()
    }

    /// buffered returns how many bytes the writer holds in memory.
    pub(crate) fn buffered(&self) -> usize {
        self.buf.len()
    }

    /// spill moves the spans written so far to a spill file.
    pub(crate) fn spill(&mut self, spill: &mut crate::file::Spill) -> Result<()> {
        self.spilled += self.buf.len() as u64;
        spill.write(&mut self.buf)
    }

    /// span appends a byte span and returns its id.
    fn span(&mut self, bytes: &[u8]) -> u32 {
        self.buf.extend_from_slice(bytes);
        self.span_ends.push(self.spilled + self.buf.len() as u64);
        self.span_ends.len() as u32
    }

    /// add_chunk adds a chunk, training the dictionary once enough chunks wait for it.
    pub fn add_chunk(&mut self, chunk: Chunk) -> Result<()> {
        if self.dictionary.is_none() {
            self.queue.push(chunk);
            if self.queue.len() < MAX_SAMPLES {
                return Ok(());
            }
            let dictionary = train_dictionary(&self.queue);
            let compressed = zstd::bulk::compress(&dictionary, LEVEL)?;
            let id = self.span(&compressed);
            self.dictionary = Some((id, zstd::bulk::Compressor::with_dictionary(LEVEL, &dictionary)?));
            for chunk in std::mem::take(&mut self.queue) {
                self.compress(chunk)?;
            }
            return Ok(());
        }
        self.compress(chunk)
    }

    /// add_stored adds a chunk that another archive stores, as that archive stores it, writing its dictionary's span
    /// the first time a chunk uses it.
    pub fn add_stored(&mut self, hash: Hash, stored: Stored) {
        let key = Arc::as_ptr(&stored.dictionary) as usize;
        let dictionary = match self.taken.get(&key) {
            Some((id, _)) => *id,
            None => {
                let id = self.span(&stored.dictionary);
                self.taken.insert(key, (id, stored.dictionary));
                id
            }
        };
        let data = self.span(&stored.data);
        self.chunks.push((hash, dictionary, data));
    }

    /// compress adds a chunk compressed with the trained dictionary.
    fn compress(&mut self, chunk: Chunk) -> Result<()> {
        let Some((id, compressor)) = &mut self.dictionary else { return Ok(()) };
        let id = *id;
        let data = compressor.compress(&chunk.data)?;
        let span = self.span(&data);
        self.chunks.push((chunk.hash, id, span));
        Ok(())
    }

    /// finish writes the chunks still held back as snappy records, then the index, the metadata naming the Dolt
    /// version, and the footer, returning the archive's name and bytes.
    pub fn finish(mut self) -> (Hash, Vec<u8>) {
        for chunk in std::mem::take(&mut self.queue) {
            let span = self.span(&chunk.to_record());
            self.chunks.push((chunk.hash, 0, span));
        }
        let index_at = self.buf.len();
        for end in std::mem::take(&mut self.span_ends).iter() {
            self.buf.extend_from_slice(&end.to_be_bytes());
        }
        let span_count = (self.buf.len() - index_at) / 8;
        self.chunks.sort_by_key(|(hash, ..)| hash.0);
        for (hash, ..) in &self.chunks {
            self.buf.extend_from_slice(&hash.prefix().to_be_bytes());
        }
        for (_, dictionary, data) in &self.chunks {
            self.buf.extend_from_slice(&dictionary.to_be_bytes());
            self.buf.extend_from_slice(&data.to_be_bytes());
        }
        for (hash, ..) in &self.chunks {
            self.buf.extend_from_slice(hash.suffix());
        }
        let index_len = (self.buf.len() - index_at) as u64;
        let metadata = format!("{{\"dolt_version\":\"{DOLT_VERSION}\"}}");
        self.buf.extend_from_slice(metadata.as_bytes());
        self.buf.extend_from_slice(&index_len.to_be_bytes());
        self.buf.extend_from_slice(&(span_count as u32).to_be_bytes());
        self.buf.extend_from_slice(&(self.chunks.len() as u32).to_be_bytes());
        self.buf.extend_from_slice(&(metadata.len() as u32).to_be_bytes());
        self.buf.extend_from_slice(&[0; 192]);
        self.buf.push(MAX_VERSION);
        self.buf.extend_from_slice(SIGNATURE);
        (Hash::of(&self.buf), self.buf)
    }
}

/// train_dictionary trains a zstd dictionary on the chunks, padding small samples as gozstd's BuildDict does, and
/// returns an empty dictionary when training fails.
fn train_dictionary(chunks: &[Chunk]) -> Vec<u8> {
    let mut samples: Vec<Vec<u8>> = chunks.iter().map(|c| c.data.clone()).collect();
    let mut total: usize = samples.iter().map(Vec::len).sum();
    while total < MIN_SAMPLES_LEN {
        let fake = format!("this is a fake sample {total}").into_bytes();
        total += fake.len();
        samples.push(fake);
    }
    zstd::dict::from_samples(&samples, DICTIONARY_SIZE).unwrap_or_default()
}
