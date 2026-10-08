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

//! The chunk journal's writer, which appends records to the journal and keeps its index file as Dolt's journalWriter
//! does.
//!
//! The index file lists where each chunk's record is, so that opening a journal need not read all of it. It is a
//! sequence of lookups, each a 0 byte, the first 16 bytes of the chunk's address, and the u64 offset and u32 length
//! of its compressed record in the journal. Batches of lookups end with a meta record: a 1 byte, the journal offsets
//! where the batch starts and ends, the CRC-32C of the batch's 16-byte addresses, and the root hash whose record is
//! at the batch's end. A batch is written once more than 16,384 chunks are unindexed.

use std::collections::HashMap;
use std::fs::{File, OpenOptions};
use std::path::Path;
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::{Arc, Mutex};
use std::time::{SystemTime, UNIX_EPOCH};

use crate::chunk::{CASTAGNOLI, Chunk};
use crate::error::{Result, corrupt};
use crate::file::{be_u32, be_u64, read_at, write_at};
use crate::hash::Hash;
use crate::journal::{
    JOURNAL_FILE, JournalRecord, MAX_RECORD_LEN, PAYLOAD_OFFSET, ROOT_RECORD_LEN, data_loss_error, parse_record,
    possible_data_loss, record_len_at,
};

/// JOURNAL_INDEX_FILE is the name of the journal's index file.
pub const JOURNAL_INDEX_FILE: &str = "journal.idx";

/// MAX_NOVEL is the number of unindexed chunks above which a commit writes an index batch.
const MAX_NOVEL: usize = 16384;

/// MAYBE_SYNC_THRESHOLD is the number of unsynced bytes above which writing a chunk commits the current root again.
const MAYBE_SYNC_THRESHOLD: u64 = 64 * 1024 * 1024;

/// JournalView reads the chunks a journal held when the view was taken, through its own handle on the file, while the
/// writer goes on appending.
pub struct JournalView {
    file: File,
    novel: HashMap<Hash, Range>,
    cached: HashMap<[u8; 16], Range>,
}

impl JournalView {
    /// get returns the chunk when the journal held it.
    pub fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        let Some(range) = self.novel.get(hash).or_else(|| self.cached.get(&addr16(hash))) else { return Ok(None) };
        Chunk::from_record(*hash, &read_at(&self.file, range.offset, range.len as usize)?).map(Some)
    }
}

thread_local! {
    /// DEFERRED is whether this thread's commits leave syncing the journal to it, and the sync they left it.
    static DEFERRED: std::cell::RefCell<(bool, Option<PendingSync>)> = const { std::cell::RefCell::new((false, None)) };
}

/// defer_syncs sets whether this thread's later commits leave syncing the journal to it, which takes each sync with
/// `take_sync` and waits on it after letting other writers in.
pub fn defer_syncs(defer: bool) {
    DEFERRED.with(|deferred| deferred.borrow_mut().0 = defer);
}

/// take_sync returns the sync that this thread's commits since the last call left to it, if any.
pub fn take_sync() -> Option<PendingSync> {
    DEFERRED.with(|deferred| deferred.borrow_mut().1.take())
}

/// PAD_LEN is how far past its last record the journal is filled with zeros, so that a commit's sync never changes the
/// file's size and fdatasync can skip its metadata, as Dolt pads it on Linux.
const PAD_LEN: u64 = 4 << 20;

/// INDEX_BUFFER_LEN is the size of the index file's write buffer.
const INDEX_BUFFER_LEN: usize = 16384;

/// Index record kinds and lengths after the kind byte.
const INDEX_LOOKUP: u8 = 0;
const INDEX_META: u8 = 1;
const LOOKUP_LEN: usize = 16 + 8 + 4;
const META_LEN: usize = 8 + 8 + 4 + Hash::LEN;

/// Range is where a chunk's compressed record is in the journal.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
struct Range {
    offset: u64,
    len: u32,
}

/// Durable is how much of a journal file is written and how much of it is synced, shared with the commits that sync it
/// after their writer moved on.
struct Durable {
    file: File,
    written: AtomicU64,
    synced: Mutex<u64>,
}

/// PendingSync is a commit's root record that is written to its journal but not yet synced.
pub struct PendingSync {
    durable: Arc<Durable>,
    end: u64,
}

impl PendingSync {
    /// wait syncs the journal unless a sync that began after the record was written already did, so that commits made
    /// at the same time share one sync.
    pub fn wait(self) -> Result<()> {
        let mut synced = self.durable.synced.lock().unwrap_or_else(|poisoned| poisoned.into_inner());
        if *synced < self.end {
            let written = self.durable.written.load(Ordering::Acquire);
            self.durable.file.sync_data()?;
            *synced = written;
        }
        Ok(())
    }
}

/// addr16 returns the first 16 bytes of an address, which the index file keys chunks by.
fn addr16(hash: &Hash) -> [u8; 16] {
    hash.0[..16].try_into().unwrap()
}

/// unix_now returns the current Unix time in seconds, which root hash records carry.
fn unix_now() -> u64 {
    SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_secs())
}

/// JournalWriter appends chunk and root hash records to a journal, buffering them until a root is committed.
pub struct JournalWriter {
    journal: File,
    index: File,
    /// Index bytes not yet written to the index file.
    index_buf: Vec<u8>,
    /// The length of the index file.
    index_len: u64,
    /// The length of the journal file, which buffered records follow.
    off: u64,
    buf: Vec<u8>,
    /// The chunks written since the last index batch.
    novel: HashMap<Hash, Range>,
    /// The chunks in earlier index batches, by the first 16 bytes of their addresses.
    cached: HashMap<[u8; 16], Range>,
    /// The journal offset that the index file covers.
    indexed: u64,
    batch_crc: crc::Digest<'static, u32>,
    unsynced: u64,
    current_root: Hash,
    uncompressed: u64,
    /// How much of the journal is written and synced, which deferred syncs share.
    durable: Arc<Durable>,
    /// How far the journal file is filled with zeros past its records.
    padded: u64,
}

impl JournalWriter {
    /// open opens the journal in a noms directory for writing, creating it and its index file when missing, and
    /// returns the writer with the last root hash written to the journal, which is empty when there is none.
    pub fn open(dir: &Path) -> Result<(JournalWriter, Hash)> {
        let options = || {
            let mut options = OpenOptions::new();
            options.read(true).write(true).create(true).truncate(false);
            options
        };
        let journal = options().open(dir.join(JOURNAL_FILE))?;
        let durable = Durable { file: journal.try_clone()?, written: AtomicU64::new(0), synced: Mutex::new(0) };
        let mut writer = JournalWriter {
            journal,
            index: options().open(dir.join(JOURNAL_INDEX_FILE))?,
            index_buf: Vec::new(),
            index_len: 0,
            off: 0,
            buf: Vec::with_capacity(MAX_RECORD_LEN as usize),
            novel: HashMap::new(),
            cached: HashMap::new(),
            indexed: 0,
            batch_crc: CASTAGNOLI.digest(),
            unsynced: 0,
            current_root: Hash::default(),
            uncompressed: 0,
            durable: Arc::new(durable),
            padded: 0,
        };
        let root = writer.bootstrap(&dir.join(JOURNAL_FILE))?;
        Ok((writer, root))
    }

    /// bootstrap loads the index file, then reads the journal from where the index ends, truncating it after its
    /// last valid record, as Dolt's bootstrapJournal does.
    fn bootstrap(&mut self, path: &Path) -> Result<Hash> {
        let index = std::fs::read(path.with_file_name(JOURNAL_INDEX_FILE))?;
        match self.read_index(&index) {
            Ok(safe) => self.index_len = safe,
            Err(_) => {
                // A corrupt index is rebuilt from the journal.
                self.index_len = 0;
                self.indexed = 0;
                self.cached.clear();
                self.batch_crc = CASTAGNOLI.digest();
            }
        }
        self.index.set_len(self.index_len)?;

        let bytes = std::fs::read(path)?;
        let mut at = self.indexed as usize;
        let (mut last, mut last_offset) = (Hash::default(), 0);
        let mut recovered = false;
        while at + 4 <= bytes.len() {
            let Some(len) = record_len_at(&bytes, at) else {
                recovered = true;
                break;
            };
            let record = parse_record(&bytes[at..at + len])?;
            match record {
                (JournalRecord::Chunk { hash, record }, payload_at) => {
                    let range = Range { offset: (at + payload_at) as u64, len: record.len() as u32 };
                    self.novel.insert(hash, range);
                    let compressed = &record[..record.len().saturating_sub(4)];
                    self.uncompressed += snap::raw::decompress_len(compressed).unwrap_or(0) as u64;
                    self.write_lookup(&hash, range)?;
                }
                (JournalRecord::Root { hash, .. }, _) => {
                    last = hash;
                    last_offset = at as u64;
                }
            }
            at += len;
        }
        if at < bytes.len() {
            recovered = true;
        }
        if recovered && possible_data_loss(&bytes[at..]) {
            return Err(data_loss_error(path, at));
        }
        self.journal.set_len(at as u64)?;
        self.journal.sync_all()?;
        self.off = at as u64;
        if self.novel.len() > MAX_NOVEL {
            self.flush_index_record(last, last_offset)?;
        }
        self.current_root = last;
        Ok(last)
    }

    /// read_index loads the complete batches of the index file, checking each against the journal, and returns the
    /// length of the index file that they cover.
    fn read_index(&mut self, bytes: &[u8]) -> Result<u64> {
        let (mut at, mut safe, mut previous) = (0, 0, 0);
        let mut batch = Vec::new();
        let mut digest = CASTAGNOLI.digest();
        while at < bytes.len() {
            let kind = bytes[at];
            at += 1;
            match kind {
                INDEX_LOOKUP => {
                    if at + LOOKUP_LEN > bytes.len() {
                        break;
                    }
                    let prefix: [u8; 16] = bytes[at..at + 16].try_into().unwrap();
                    digest.update(&prefix);
                    batch.push((prefix, Range { offset: be_u64(bytes, at + 16), len: be_u32(bytes, at + 24) }));
                    at += LOOKUP_LEN;
                }
                INDEX_META => {
                    if at + META_LEN > bytes.len() {
                        break;
                    }
                    let (start, end) = (be_u64(bytes, at), be_u64(bytes, at + 8));
                    let checksum = be_u32(bytes, at + 16);
                    let root = Hash(bytes[at + 20..at + 20 + Hash::LEN].try_into().unwrap());
                    let actual = std::mem::replace(&mut digest, CASTAGNOLI.digest()).finalize();
                    if checksum != actual {
                        return Err(corrupt(format!("invalid index checksum ({actual} != {checksum})")));
                    }
                    if start != previous {
                        return Err(corrupt(format!(
                            "index records do not cover contiguous region ({start} != {previous})"
                        )));
                    }
                    previous = end;
                    if self.root_at(end)? != root {
                        return Err(corrupt(format!("invalid index record hash at {end}")));
                    }
                    self.cached.extend(batch.drain(..));
                    self.indexed = end;
                    at += META_LEN;
                    safe = at as u64;
                }
                _ => return Err(corrupt("journal index is malformed")),
            }
        }
        Ok(safe)
    }

    /// root_at returns the address of the root hash record at the journal offset.
    fn root_at(&self, offset: u64) -> Result<Hash> {
        let bytes = read_at(&self.journal, offset, ROOT_RECORD_LEN)?;
        match record_len_at(&bytes, 0).map(|len| parse_record(&bytes[..len])) {
            Some(Ok((JournalRecord::Root { hash, .. }, _))) => Ok(hash),
            _ => Err(corrupt(format!("expected a root hash record at {offset}"))),
        }
    }

    /// write_lookup adds a chunk's lookup to the index.
    fn write_lookup(&mut self, hash: &Hash, range: Range) -> Result<()> {
        let prefix = addr16(hash);
        self.index_buf.push(INDEX_LOOKUP);
        self.index_buf.extend_from_slice(&prefix);
        self.index_buf.extend_from_slice(&range.offset.to_be_bytes());
        self.index_buf.extend_from_slice(&range.len.to_be_bytes());
        self.batch_crc.update(&prefix);
        if self.index_buf.len() >= INDEX_BUFFER_LEN {
            self.flush_index()?;
        }
        Ok(())
    }

    /// flush_index writes the buffered index bytes to the index file.
    fn flush_index(&mut self) -> Result<()> {
        write_at(&self.index, self.index_len, &self.index_buf)?;
        self.index_len += self.index_buf.len() as u64;
        self.index_buf.clear();
        Ok(())
    }

    /// flush_index_record ends the index batch at the root hash record at the journal offset.
    fn flush_index_record(&mut self, root: Hash, end: u64) -> Result<()> {
        let checksum = std::mem::replace(&mut self.batch_crc, CASTAGNOLI.digest()).finalize();
        self.index_buf.push(INDEX_META);
        self.index_buf.extend_from_slice(&self.indexed.to_be_bytes());
        self.index_buf.extend_from_slice(&end.to_be_bytes());
        self.index_buf.extend_from_slice(&checksum.to_be_bytes());
        self.index_buf.extend_from_slice(&root.0);
        self.flush_index()?;
        for (hash, range) in self.novel.drain() {
            self.cached.insert(addr16(&hash), range);
        }
        self.indexed = end;
        Ok(())
    }

    /// offset returns the journal offset where the next record goes.
    fn offset(&self) -> u64 {
        self.off + self.buf.len() as u64
    }

    /// reserve makes room in the buffer for a record of the length, writing the buffer out when it is full.
    fn reserve(&mut self, len: usize) -> Result<()> {
        if len > MAX_RECORD_LEN as usize {
            return Err(corrupt(format!("requested bytes ({len}) exceeds capacity ({MAX_RECORD_LEN})")));
        }
        if self.buf.len() + len > MAX_RECORD_LEN as usize {
            self.flush()?;
        }
        Ok(())
    }

    /// flush writes the buffered records to the journal file.
    pub fn flush(&mut self) -> Result<()> {
        write_at(&self.journal, self.off, &self.buf)?;
        self.off += self.buf.len() as u64;
        if !self.buf.is_empty() && self.off > self.padded {
            write_at(&self.journal, self.off, &vec![0; PAD_LEN as usize])?;
            self.padded = self.off + PAD_LEN;
        }
        self.buf.clear();
        self.durable.written.store(self.off, Ordering::Release);
        Ok(())
    }

    /// write_chunk appends a chunk's record, given its compressed chunk record.
    pub fn write_chunk(&mut self, hash: Hash, record: &[u8]) -> Result<()> {
        let encoded = JournalRecord::Chunk { hash, record }.encode();
        self.reserve(encoded.len())?;
        let range = Range { offset: self.offset() + PAYLOAD_OFFSET as u64, len: record.len() as u32 };
        self.buf.extend_from_slice(&encoded);
        self.unsynced += encoded.len() as u64;
        self.novel.insert(hash, range);
        self.write_lookup(&hash, range)?;
        if self.unsynced > MAYBE_SYNC_THRESHOLD && !self.current_root.is_empty() {
            self.commit_root(self.current_root)?;
        }
        Ok(())
    }

    /// commit_root appends a root hash record stamped with the current time and syncs the journal, unless syncs are
    /// deferred.
    pub fn commit_root(&mut self, root: Hash) -> Result<()> {
        self.commit_root_at(root, unix_now())
    }

    /// commit_root_at appends a root hash record stamped with the Unix time in seconds and syncs the journal unless
    /// syncs are deferred, ending an index batch when enough chunks are unindexed.
    pub fn commit_root_at(&mut self, root: Hash, timestamp: u64) -> Result<()> {
        let encoded = JournalRecord::Root { hash: root, timestamp }.encode();
        self.reserve(encoded.len())?;
        let start = self.offset();
        self.current_root = root;
        self.buf.extend_from_slice(&encoded);
        self.flush()?;
        let deferred = DEFERRED.with(|deferred| {
            let mut deferred = deferred.borrow_mut();
            if deferred.0 {
                deferred.1 = Some(PendingSync { durable: self.durable.clone(), end: self.off });
            }
            deferred.0
        });
        if !deferred {
            self.journal.sync_data()?;
        }
        self.unsynced = 0;
        if self.novel.len() > MAX_NOVEL {
            self.flush_index_record(root, start)?;
        }
        Ok(())
    }

    /// view writes out the buffered records and returns a view of the chunks the journal holds now.
    pub fn view(&mut self) -> Result<JournalView> {
        self.flush()?;
        Ok(JournalView { file: self.journal.try_clone()?, novel: self.novel.clone(), cached: self.cached.clone() })
    }

    /// root returns the last root hash committed.
    pub fn root(&self) -> Hash {
        self.current_root
    }

    /// range returns where the chunk's compressed record is, when the journal holds it.
    fn range(&self, hash: &Hash) -> Option<Range> {
        self.novel.get(hash).or_else(|| self.cached.get(&addr16(hash))).copied()
    }

    /// has reports whether the journal holds the chunk.
    pub fn has(&self, hash: &Hash) -> bool {
        self.range(hash).is_some()
    }

    /// locate returns where the chunk's record is in the journal file, when the journal holds it.
    pub fn locate(&self, hash: &Hash) -> Option<crate::Location> {
        let range = self.range(hash)?;
        Some(crate::Location {
            file: crate::journal::JOURNAL_FILE.to_string(),
            offset: range.offset,
            length: range.len,
            dictionary: None,
        })
    }

    /// get returns the chunk when the journal holds it, reading buffered records from memory.
    pub fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        let Some(range) = self.range(hash) else { return Ok(None) };
        let record = if range.offset >= self.off {
            let start = (range.offset - self.off) as usize;
            self.buf[start..start + range.len as usize].to_vec()
        } else {
            read_at(&self.journal, range.offset, range.len as usize)?
        };
        Chunk::from_record(*hash, &record).map(Some)
    }

    /// count returns the number of chunks the journal holds, as its spec in the manifest counts them.
    pub fn count(&self) -> usize {
        self.novel.len() + self.cached.len()
    }

    /// uncompressed_size returns the total uncompressed size of the chunks read or written since opening.
    pub fn uncompressed_size(&self) -> u64 {
        self.uncompressed
    }

    /// close writes out the buffered records and index, and syncs the journal.
    pub fn close(mut self) -> Result<()> {
        self.sync()?;
        if self.padded > self.off {
            self.journal.set_len(self.off)?;
            self.journal.sync_all()?;
        }
        Ok(())
    }

    /// sync writes out the buffered records and index, and syncs the journal, which stays open.
    pub fn sync(&mut self) -> Result<()> {
        self.flush()?;
        self.flush_index()?;
        self.journal.sync_all()?;
        Ok(())
    }
}
