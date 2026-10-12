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

//! Stores whose table files and manifest are blobs in a blobstore, such as a bucket of a cloud object store, as
//! Dolt's NBS on a blobstore keeps them.

use std::collections::{HashMap, HashSet};
use std::path::PathBuf;
use std::sync::Arc;

use crate::error::{Error, corrupt};
use crate::file::ReadAt;
use crate::journal_store::MANIFEST_VERSION;
use crate::store::{ARCHIVE_SUFFIX, Source};
use crate::{ArchiveReader, Chunk, ChunkReader, ChunkStore, Hash, Manifest, Result, TableReader, TableSpec, lock_hash};

/// MANIFEST_KEY is the blob that holds a store's manifest.
pub const MANIFEST_KEY: &str = "manifest";

/// RECORDS_SUFFIX names the blob of a table file's chunk records, which its blob concatenates with its tail.
const RECORDS_SUFFIX: &str = ".records";
/// TAIL_SUFFIX names the blob of a table file's index and footer.
const TAIL_SUFFIX: &str = ".tail";

/// MEM_TABLE_SIZE is how many bytes of chunks a store collects before it writes them to a table file.
const MEM_TABLE_SIZE: u64 = 256 << 20;

/// BlobRange is a range of a blob's bytes: from an offset, or from the end for a negative offset, for a length, or to
/// the end for a length of 0.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub struct BlobRange {
    pub offset: i64,
    pub length: i64,
}

impl BlobRange {
    /// ALL is the whole blob.
    pub const ALL: BlobRange = BlobRange { offset: 0, length: 0 };

    /// is_all reports whether the range is the whole blob.
    pub fn is_all(&self) -> bool {
        self.offset == 0 && self.length == 0
    }

    /// positive returns the range from its start offset for a blob of the size.
    pub fn positive(&self, size: i64) -> BlobRange {
        let offset = if self.offset < 0 { size + self.offset } else { self.offset };
        let length = if self.length == 0 || offset + self.length > size { size - offset } else { self.length };
        BlobRange { offset, length }
    }

    /// http_header returns the range as an HTTP Range header, or None for the whole blob.
    pub fn http_header(&self) -> Option<String> {
        if self.is_all() {
            return None;
        }
        if self.length == 0 || self.offset < 0 {
            return Some(format!("bytes={}", self.offset));
        }
        Some(format!("bytes={}-{}", self.offset, self.offset + self.length - 1))
    }
}

/// Blob is the bytes of a range of a blob, with the size of the whole blob, or 0 when the store cannot tell, and the
/// blob's version.
pub struct Blob {
    pub data: Vec<u8>,
    pub size: u64,
    pub version: String,
}

/// Blobstore keeps blobs by key, as Dolt's blobstore.Blobstore does.
pub trait Blobstore: Send + Sync {
    /// path returns the blobstore's location.
    fn path(&self) -> String;

    /// exists reports whether the blobstore holds a blob.
    fn exists(&self, key: &str) -> Result<bool>;

    /// get returns a range of a blob, failing with NotFound when the blobstore lacks it.
    fn get(&self, key: &str, range: BlobRange) -> Result<Blob>;

    /// put stores a blob, returning its version.
    fn put(&self, key: &str, data: &[u8]) -> Result<String>;

    /// check_and_put_manifest stores the manifest when its version is still the expected one, which is empty when no
    /// manifest is expected, failing with VersionMismatch otherwise.
    fn check_and_put_manifest(&self, expected: &str, data: &[u8]) -> Result<String>;

    /// concatenate stores a blob of the blobs with the keys, in order, returning its version.
    fn concatenate(&self, key: &str, sources: &[String]) -> Result<String>;
}

/// not_found returns the error for a blob that a blobstore lacks.
pub fn not_found(key: &str) -> Error {
    Error::NotFound(key.to_string())
}

/// BlobReader reads a blob in ranges, knowing its size.
struct BlobReader {
    blobs: Arc<dyn Blobstore>,
    key: String,
    size: u64,
}

impl BlobReader {
    /// open learns a blob's size from its end, failing with NotFound when the blobstore lacks it.
    fn open(blobs: Arc<dyn Blobstore>, key: String) -> Result<BlobReader> {
        let mut size = blobs.get(&key, BlobRange { offset: -1, length: 0 })?.size;
        if size == 0 {
            size = blobs.get(&key, BlobRange::ALL)?.data.len() as u64;
        }
        Ok(BlobReader { blobs, key, size })
    }
}

impl ReadAt for BlobReader {
    fn read_at(&self, offset: u64, len: usize) -> Result<Vec<u8>> {
        if len == 0 {
            return Ok(Vec::new());
        }
        let blob = self.blobs.get(&self.key, BlobRange { offset: offset as i64, length: len as i64 })?;
        if blob.data.len() != len {
            return Err(corrupt(format!("{}: read {} bytes at {offset}, expected {len}", self.key, blob.data.len())));
        }
        Ok(blob.data)
    }

    fn size(&self) -> Result<u64> {
        Ok(self.size)
    }
}

/// open_source opens a table file or archive of a blobstore, trying a table file first as Dolt does.
fn open_source(blobs: &Arc<dyn Blobstore>, name: &Hash) -> Result<Source> {
    match BlobReader::open(blobs.clone(), name.to_string()) {
        Ok(reader) => {
            Ok(Source::Table(TableReader::open_reader(Box::new(reader), name.to_string(), &name.to_string())?))
        }
        Err(Error::NotFound(_)) => {
            let key = format!("{name}{ARCHIVE_SUFFIX}");
            let reader = BlobReader::open(blobs.clone(), key.clone())?;
            Ok(Source::Archive(ArchiveReader::open_reader(Box::new(reader), key.clone(), &key)?))
        }
        Err(err) => Err(err),
    }
}

/// read_manifest returns a blobstore's manifest with its version, or None when it has none.
fn read_manifest(blobs: &dyn Blobstore) -> Result<Option<(Manifest, String)>> {
    match blobs.get(MANIFEST_KEY, BlobRange::ALL) {
        Ok(blob) => Ok(Some((Manifest::parse(&blob.data)?, blob.version))),
        Err(Error::NotFound(_)) => Ok(None),
        Err(err) => Err(err),
    }
}

/// BlobChunkStore is a chunk store whose table files and manifest are blobs, as Dolt's NBS on a blobstore is. Chunks
/// put collect in memory until a commit writes them as one table file, and a commit replaces the manifest only when
/// no other writer replaced it first.
pub struct BlobChunkStore {
    blobs: Arc<dyn Blobstore>,
    /// Whether table files are written as their records and tail and then concatenated, as Dolt does for blobstores
    /// that can concatenate, rather than written whole.
    concatenates: bool,
    upstream: Manifest,
    /// The version of the manifest blob that upstream was read from, empty when there was none.
    version: String,
    sources: Vec<Source>,
    /// Table files written since the last commit.
    pending: Vec<TableSpec>,
    chunks: HashMap<Hash, Vec<u8>>,
    order: Vec<Hash>,
    refs: Vec<Hash>,
    size: u64,
}

impl BlobChunkStore {
    /// open opens the store that a blobstore holds, which is empty without a manifest.
    pub fn open(blobs: Arc<dyn Blobstore>, format: &str, concatenates: bool) -> Result<BlobChunkStore> {
        let (upstream, version) = read_manifest(blobs.as_ref())?.unwrap_or_else(|| {
            let empty = Manifest {
                version: MANIFEST_VERSION.to_string(),
                format: format.to_string(),
                lock: Hash::default(),
                root: Hash::default(),
                gc_gen: Hash::default(),
                specs: Vec::new(),
            };
            (empty, String::new())
        });
        let sources = upstream.specs.iter().map(|spec| open_source(&blobs, &spec.name)).collect::<Result<_>>()?;
        Ok(BlobChunkStore {
            blobs,
            concatenates,
            upstream,
            version,
            sources,
            pending: Vec::new(),
            chunks: HashMap::new(),
            order: Vec::new(),
            refs: Vec::new(),
            size: 0,
        })
    }

    /// has_persisted reports whether a table file holds the chunk.
    fn has_persisted(&self, hash: &Hash) -> bool {
        self.sources.iter().any(|source| source.has(hash))
    }

    /// persist writes the chunks put since the last write that the store lacks to a new table file, as its records
    /// and tail and then their concatenation, as Dolt's blobstore persister does.
    fn persist(&mut self) -> Result<()> {
        if self.order.is_empty() {
            return Ok(());
        }
        let refs = std::mem::take(&mut self.refs);
        let mut absent: Vec<Hash> =
            refs.into_iter().filter(|h| !self.chunks.contains_key(h) && !self.has_persisted(h)).collect();
        if !absent.is_empty() {
            absent.sort_by_key(|hash| hash.0);
            absent.dedup();
            self.clear_memtable();
            return Err(Error::DanglingRef(absent));
        }
        let mut writer = crate::table::TableWriter::new();
        let mut written = HashSet::new();
        for hash in std::mem::take(&mut self.order) {
            if !self.has_persisted(&hash) && written.insert(hash) {
                writer.add_chunk(&Chunk { hash, data: self.chunks[&hash].clone() });
            }
        }
        self.clear_memtable();
        if writer.count() == 0 {
            return Ok(());
        }
        let chunk_count = writer.count() as u32;
        let (name, bytes) = writer.finish();
        let key = name.to_string();
        if self.concatenates {
            let split = bytes.len() - crate::table::tail_len(chunk_count as u64) as usize;
            let (records, tail) = (format!("{key}{RECORDS_SUFFIX}"), format!("{key}{TAIL_SUFFIX}"));
            self.blobs.put(&records, &bytes[..split])?;
            self.blobs.put(&tail, &bytes[split..])?;
            self.blobs.concatenate(&key, &[records, tail])?;
        } else {
            self.blobs.put(&key, &bytes)?;
        }
        self.sources.push(open_source(&self.blobs, &name)?);
        self.pending.push(TableSpec { name, chunk_count });
        Ok(())
    }

    /// clear_memtable drops the chunks put since the last write.
    fn clear_memtable(&mut self) {
        self.chunks.clear();
        self.order.clear();
        self.refs.clear();
        self.size = 0;
    }

    /// refresh reads the manifest that another writer replaced, opening the files it adds.
    fn refresh(&mut self) -> Result<()> {
        let Some((manifest, version)) = read_manifest(self.blobs.as_ref())? else { return Ok(()) };
        for spec in &manifest.specs {
            if !self.upstream.specs.iter().chain(&self.pending).any(|s| s.name == spec.name) {
                self.sources.push(open_source(&self.blobs, &spec.name)?);
            }
        }
        self.upstream = manifest;
        self.version = version;
        Ok(())
    }

    /// path returns the location of the store's blobstore.
    pub fn path(&self) -> PathBuf {
        PathBuf::from(self.blobs.path())
    }
}

impl ChunkReader for BlobChunkStore {
    fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        if let Some(data) = self.chunks.get(hash) {
            return Ok(Some(Chunk { hash: *hash, data: data.clone() }));
        }
        for source in &self.sources {
            if let Some(chunk) = source.get(hash)? {
                return Ok(Some(chunk));
            }
        }
        Ok(None)
    }
}

impl ChunkStore for BlobChunkStore {
    fn has(&self, hash: &Hash) -> bool {
        self.chunks.contains_key(hash) || self.has_persisted(hash)
    }

    fn put(&mut self, chunk: Chunk, refs: Vec<Hash>) -> Result<()> {
        if self.chunks.contains_key(&chunk.hash) {
            return Ok(());
        }
        if self.size + chunk.data.len() as u64 > MEM_TABLE_SIZE {
            self.persist()?;
        }
        self.size += chunk.data.len() as u64;
        self.order.push(chunk.hash);
        self.chunks.insert(chunk.hash, chunk.data);
        self.refs.extend(refs);
        Ok(())
    }

    fn commit(&mut self, current: Hash, last: Hash) -> Result<bool> {
        if self.upstream.root != last {
            return Ok(false);
        }
        if self.order.is_empty() && self.pending.is_empty() && current == last {
            return Ok(true);
        }
        self.persist()?;
        if !current.is_empty() && !self.has(&current) {
            return Err(Error::DanglingRef(vec![current]));
        }
        loop {
            let mut specs = self.upstream.specs.clone();
            specs.extend(
                self.pending.iter().filter(|p| !specs.iter().any(|s| s.name == p.name)).copied().collect::<Vec<_>>(),
            );
            specs.sort_by_key(|spec| spec.name.0);
            let next = Manifest {
                version: MANIFEST_VERSION.to_string(),
                format: self.upstream.format.clone(),
                lock: lock_hash(&current, &specs, &[], b""),
                root: current,
                gc_gen: self.upstream.gc_gen,
                specs,
            };
            match self.blobs.check_and_put_manifest(&self.version, next.format().as_bytes()) {
                Ok(version) => {
                    self.upstream = next;
                    self.version = version;
                    self.pending.clear();
                    return Ok(true);
                }
                Err(Error::VersionMismatch { .. }) => {
                    self.refresh()?;
                    if self.upstream.root != last {
                        return Ok(false);
                    }
                }
                Err(err) => return Err(err),
            }
        }
    }

    fn root(&self) -> Hash {
        self.upstream.root
    }
}

/// LocalBlobstore keeps blobs as files of a directory, each named by its key with a `.bs` extension, whose version is
/// its modification time, as Dolt's LocalBlobstore does.
pub struct LocalBlobstore {
    dir: PathBuf,
}

impl LocalBlobstore {
    /// new returns the blobstore of a directory.
    pub fn new(dir: PathBuf) -> LocalBlobstore {
        LocalBlobstore { dir }
    }

    /// file returns the path of a blob's file.
    fn file(&self, key: &str) -> PathBuf {
        self.dir.join(format!("{key}.bs"))
    }

    /// version returns a blob file's version.
    fn version(&self, key: &str) -> Result<String> {
        let modified = std::fs::metadata(self.file(key))?.modified()?;
        let since = modified.duration_since(std::time::UNIX_EPOCH).unwrap_or_default();
        Ok(since.as_nanos().to_string())
    }
}

impl Blobstore for LocalBlobstore {
    fn path(&self) -> String {
        self.dir.display().to_string()
    }

    fn exists(&self, key: &str) -> Result<bool> {
        Ok(self.file(key).exists())
    }

    fn get(&self, key: &str, range: BlobRange) -> Result<Blob> {
        let file = match std::fs::File::open(self.file(key)) {
            Ok(file) => file,
            Err(err) if err.kind() == std::io::ErrorKind::NotFound => return Err(not_found(key)),
            Err(err) => return Err(err.into()),
        };
        let size = file.metadata()?.len();
        let range = range.positive(size as i64);
        let data = crate::file::read_at(&file, range.offset as u64, range.length.max(0) as usize)?;
        Ok(Blob { data, size, version: self.version(key)? })
    }

    fn put(&self, key: &str, data: &[u8]) -> Result<String> {
        std::fs::create_dir_all(&self.dir)?;
        let temp = self.dir.join(format!(".{key}.{}.tmp", std::process::id()));
        std::fs::write(&temp, data)?;
        std::thread::sleep(std::time::Duration::from_millis(10));
        std::fs::rename(&temp, self.file(key))?;
        self.version(key)
    }

    fn check_and_put_manifest(&self, expected: &str, data: &[u8]) -> Result<String> {
        std::fs::create_dir_all(&self.dir)?;
        let lock = std::fs::File::create(self.dir.join(format!("{MANIFEST_KEY}.bs.lock")))?;
        lock.lock()?;
        let actual = match self.exists(MANIFEST_KEY)? {
            true => self.version(MANIFEST_KEY)?,
            false => String::new(),
        };
        if actual != expected {
            return Err(Error::VersionMismatch { key: MANIFEST_KEY.into(), expected: expected.into(), actual });
        }
        self.put(MANIFEST_KEY, data)
    }

    fn concatenate(&self, key: &str, sources: &[String]) -> Result<String> {
        let mut data = Vec::new();
        for source in sources {
            data.extend(std::fs::read(self.file(source))?);
        }
        self.put(key, &data)
    }
}
