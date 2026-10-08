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

//! A writable chunk store whose new chunks go to the chunk journal, as Dolt's journaling NomsBlockStore.

use std::collections::{HashMap, HashSet};
use std::fs::File;
use std::path::{Path, PathBuf};
use std::sync::Arc;

use crate::chunk::Chunk;
use crate::error::{Error, Result, corrupt};
use crate::hash::{BuildAddrHasher, Hash};
use crate::journal::JOURNAL_FILE;
use crate::journal_writer::JournalWriter;
use crate::manifest::{MANIFEST_FILE, Manifest, TableSpec, lock_hash};
use crate::store::Source;

/// LOCK_FILE is the file a process locks while it has a store open.
const LOCK_FILE: &str = "LOCK";

/// MANIFEST_VERSION is the manifest version Dolt writes.
pub(crate) const MANIFEST_VERSION: &str = "5";

/// MEM_TABLE_SIZE is the size of chunk data a store holds in memory before writing it to the journal.
const MEM_TABLE_SIZE: u64 = 128 << 20;

/// MemTable holds the chunks put since the store last wrote to the journal, in the order they were put.
#[derive(Default)]
struct MemTable {
    chunks: HashMap<Hash, Vec<u8>, BuildAddrHasher>,
    order: Vec<Hash>,
    size: u64,
    /// The addresses the chunks refer to, which must be in the store before the chunks are written.
    refs: Vec<Hash>,
}

/// Sources are the table files and archives that a store's manifest names, shared with the reads planned over them.
type Sources = Arc<Vec<Arc<Source>>>;

/// Snapshot is the chunks that a store's files held when it was taken.
pub struct Snapshot {
    sources: Sources,
    journal: Option<crate::JournalView>,
}

impl crate::ChunkReader for Snapshot {
    fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        if let Some(journal) = &self.journal
            && let Some(chunk) = journal.get(hash)?
        {
            return Ok(Some(chunk));
        }
        for source in self.sources.iter() {
            if let Some(chunk) = source.get(hash)? {
                return Ok(Some(chunk));
            }
        }
        Ok(None)
    }

    fn get_stored(&self, hash: &Hash) -> Result<Option<(Chunk, Option<crate::Stored>)>> {
        if let Some(journal) = &self.journal
            && let Some(chunk) = journal.get(hash)?
        {
            return Ok(Some((chunk, None)));
        }
        for source in self.sources.iter() {
            if let Some(found) = source.get_stored(hash)? {
                return Ok(Some(found));
            }
        }
        Ok(None)
    }
}

/// JournalStore is a writable chunk store whose new chunks go to the chunk journal, or to new table files when it
/// was opened without one.
pub struct JournalStore {
    dir: PathBuf,
    /// The lock that keeps other processes from opening the store while it is open, which only journaling stores
    /// hold.
    _lock: Option<File>,
    /// Whether new chunks go to the journal rather than to table files.
    journaled: bool,
    /// The table files written since the last commit, which the next manifest names.
    pending: Vec<TableSpec>,
    /// The manifest as of the last commit, whose root comes from the journal.
    upstream: Manifest,
    /// The table files and archives the manifest names.
    sources: Sources,
    journal: Option<JournalWriter>,
    memtable: MemTable,
}

impl JournalStore {
    /// open opens the store in a noms directory for writing, as Dolt's NewLocalJournalingStore does: the journal's
    /// last root hash replaces the manifest's, and a store without a manifest starts empty.
    pub fn open(dir: &Path, format: &str) -> Result<JournalStore> {
        if !dir.is_dir() {
            return Err(corrupt(format!("path is not a directory: {}", dir.display())));
        }
        let lock = File::options().read(true).write(true).create(true).truncate(false).open(dir.join(LOCK_FILE))?;
        lock.try_lock()
            .map_err(|_| corrupt(format!("the database at {} is locked by another process", dir.display())))?;
        JournalStore::open_with(dir, format, Some(lock))
    }

    /// open_tables opens the store in a directory for writing without a journal, so that each commit writes its new
    /// chunks to a table file, as Dolt's NewLocalStore does for file remotes and backups.
    pub fn open_tables(dir: &Path, format: &str) -> Result<JournalStore> {
        if !dir.is_dir() {
            return Err(corrupt(format!("path is not a directory: {}", dir.display())));
        }
        File::options().read(true).write(true).create(true).truncate(false).open(dir.join(LOCK_FILE))?;
        JournalStore::open_with(dir, format, None)
    }

    /// open_with opens the store, reading the journal when the lock shows it is a journaling store.
    fn open_with(dir: &Path, format: &str, lock: Option<File>) -> Result<JournalStore> {
        let journaled = lock.is_some();
        let manifest = Manifest::read(dir)?;
        let mut upstream = manifest.clone().unwrap_or_else(|| Manifest {
            version: MANIFEST_VERSION.to_string(),
            format: format.to_string(),
            lock: Hash::default(),
            root: Hash::default(),
            gc_gen: Hash::default(),
            specs: Vec::new(),
        });
        let mut journal = None;
        if journaled && dir.join(JOURNAL_FILE).exists() {
            let (mut writer, root) = JournalWriter::open(dir)?;
            if root.is_empty() {
                if let Some(manifest) = &manifest {
                    writer.commit_root(manifest.root)?;
                }
            } else if manifest.is_some() {
                // The manifest takes the journal's root, as Dolt's trueUpBackingManifest does.
                upstream.root = root;
                upstream.lock = lock_hash(&root, &upstream.specs, &[], b"");
                write_manifest(dir, &upstream)?;
            }
            journal = Some(writer);
        }
        let journal_name = Hash::parse(JOURNAL_FILE).unwrap();
        let sources = upstream
            .specs
            .iter()
            .filter(|spec| spec.name != journal_name)
            .map(|spec| Source::open_file(dir, &spec.name).map(Arc::new))
            .collect::<Result<Vec<_>>>()?;
        let sources = Arc::new(sources);
        Ok(JournalStore {
            dir: dir.to_path_buf(),
            _lock: lock,
            journaled,
            pending: Vec::new(),
            upstream,
            sources,
            journal,
            memtable: MemTable::default(),
        })
    }

    /// root returns the root hash of the last commit.
    pub fn root(&self) -> Hash {
        self.upstream.root
    }

    /// manifest returns the manifest as of the last commit.
    pub fn manifest(&self) -> &Manifest {
        &self.upstream
    }

    /// get returns the chunk when the store holds it, including chunks put but not yet committed.
    pub fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        if let Some(data) = self.memtable.chunks.get(hash) {
            return Ok(Some(Chunk { hash: *hash, data: data.clone() }));
        }
        self.get_persisted(hash)
    }

    /// plan returns how to read a chunk once the store is let go: the chunk itself when it is in memory, where it is
    /// in the journal, or the files to look in.
    pub fn plan(&self, hash: &Hash) -> Result<crate::Plan> {
        if let Some(data) = self.memtable.chunks.get(hash) {
            return Ok(crate::Plan::ready(Some(Chunk { hash: *hash, data: data.clone() })));
        }
        if let Some(plan) = self.journal.as_ref().map(|j| j.plan(hash)).transpose()?.flatten() {
            return Ok(plan);
        }
        Ok(crate::Plan(crate::PlanKind::Files(self.sources.clone(), *hash)))
    }

    /// get_persisted returns the chunk when the journal or a file holds it.
    fn get_persisted(&self, hash: &Hash) -> Result<Option<Chunk>> {
        if let Some(chunk) = self.journal.as_ref().map(|j| j.get(hash)).transpose()?.flatten() {
            return Ok(Some(chunk));
        }
        for source in self.sources.iter() {
            if let Some(chunk) = source.get(hash)? {
                return Ok(Some(chunk));
            }
        }
        Ok(None)
    }

    /// locate returns where a committed chunk is in the store's files, writing out the journal's buffered records first
    /// so that its file holds them.
    pub fn locate(&mut self, hash: &Hash) -> Result<Option<crate::Location>> {
        if let Some(journal) = self.journal.as_mut() {
            journal.flush()?;
            if let Some(location) = journal.locate(hash) {
                return Ok(Some(location));
            }
        }
        Ok(self.sources.iter().find_map(|source| source.locate(hash)))
    }

    /// has reports whether the store holds the chunk, including chunks put but not yet committed.
    pub fn has(&self, hash: &Hash) -> bool {
        self.memtable.chunks.contains_key(hash) || self.has_persisted(hash)
    }

    /// has_persisted reports whether the journal or a file holds the chunk.
    fn has_persisted(&self, hash: &Hash) -> bool {
        self.journal.as_ref().is_some_and(|j| j.has(hash)) || self.sources.iter().any(|s| s.has(hash))
    }

    /// put adds a chunk with the addresses it refers to, which becomes durable when a later commit succeeds. Writing
    /// it fails when it refers to a chunk the store lacks.
    pub fn put(&mut self, chunk: Chunk, refs: impl IntoIterator<Item = Hash>) -> Result<()> {
        assert!(!chunk.data.is_empty(), "NBS blocks cannot be zero length");
        if self.memtable.chunks.contains_key(&chunk.hash) {
            return Ok(());
        }
        if self.memtable.size + chunk.data.len() as u64 > MEM_TABLE_SIZE {
            self.persist()?;
        }
        self.memtable.size += chunk.data.len() as u64;
        self.memtable.order.push(chunk.hash);
        self.memtable.chunks.insert(chunk.hash, chunk.data);
        self.memtable.refs.extend(refs);
        Ok(())
    }

    /// check_refs fails when an address is in neither the memtable nor the store, as Dolt's refCheck does, dropping
    /// the memtable as Dolt does on a dangling reference.
    fn check_refs(&mut self, refs: &[Hash]) -> Result<()> {
        let mut absent: Vec<Hash> = refs
            .iter()
            .filter(|hash| !self.memtable.chunks.contains_key(hash) && !self.has_persisted(hash))
            .copied()
            .collect();
        if absent.is_empty() {
            return Ok(());
        }
        absent.sort_by_key(|hash| hash.0);
        absent.dedup();
        self.memtable = MemTable::default();
        Err(Error::DanglingRef(absent))
    }

    /// persist writes the memtable's chunks that the store lacks to the journal, in the order they were put, as
    /// Dolt's ChunkJournal.Persist does, or to a new table file in a store without a journal.
    fn persist(&mut self) -> Result<()> {
        if self.memtable.order.is_empty() {
            return Ok(());
        }
        let refs = std::mem::take(&mut self.memtable.refs);
        self.check_refs(&refs)?;
        let memtable = std::mem::take(&mut self.memtable);
        if !self.journaled {
            let mut writer = crate::table::TableWriter::new();
            let mut written = HashSet::<Hash, BuildAddrHasher>::default();
            for hash in memtable.order.iter().filter(|hash| !self.has_persisted(hash)) {
                if written.insert(*hash) {
                    writer.add_chunk(&Chunk { hash: *hash, data: memtable.chunks[hash].clone() });
                }
            }
            if writer.count() == 0 {
                return Ok(());
            }
            let chunk_count = writer.count() as u32;
            let (name, bytes) = writer.finish();
            let path = self.dir.join(name.to_string());
            std::fs::write(&path, bytes)?;
            File::open(&path)?.sync_all()?;
            Arc::make_mut(&mut self.sources).push(Arc::new(Source::open_file(&self.dir, &name)?));
            self.pending.push(TableSpec { name, chunk_count });
            return Ok(());
        }
        if self.journal.is_none() {
            let (mut writer, _) = JournalWriter::open(&self.dir)?;
            if !self.upstream.lock.is_empty() {
                writer.commit_root(self.upstream.root)?;
            }
            self.journal = Some(writer);
        }
        let novel: Vec<Hash> = memtable.order.iter().filter(|hash| !self.has_persisted(hash)).copied().collect();
        let journal = self.journal.as_mut().unwrap();
        let mut written = HashSet::<Hash, BuildAddrHasher>::default();
        for hash in novel {
            if written.insert(hash) {
                let chunk = Chunk { hash, data: memtable.chunks[&hash].clone() };
                journal.write_chunk(hash, &chunk.to_record())?;
            }
        }
        Ok(())
    }

    /// specs returns the files holding the store's chunks, sorted by name, as Dolt's tableSet.toSpecs does.
    /// table_files returns the files the manifest names, with the journal as it is now, as Dolt's Sources lists them.
    pub fn table_files(&self) -> Vec<TableSpec> {
        self.specs()
    }

    /// add_table_files adds table files or archives already written to the store's directory, which the next commit's
    /// manifest names, failing for a file that is missing.
    pub fn add_table_files(&mut self, specs: &[TableSpec]) -> Result<()> {
        for spec in specs {
            if self.specs().iter().any(|existing| existing.name == spec.name) {
                continue;
            }
            Arc::make_mut(&mut self.sources).push(Arc::new(Source::open_file(&self.dir, &spec.name)?));
            self.pending.push(*spec);
        }
        Ok(())
    }

    fn specs(&self) -> Vec<TableSpec> {
        let journal_name = Hash::parse(JOURNAL_FILE).unwrap();
        let mut specs: Vec<TableSpec> =
            self.upstream.specs.iter().filter(|spec| spec.name != journal_name).copied().collect();
        specs.extend(self.pending.iter().copied());
        if let Some(journal) = self.journal.as_ref().filter(|j| j.count() > 0) {
            specs.push(TableSpec { name: journal_name, chunk_count: journal.count() as u32 });
        }
        specs.sort_by_key(|spec| spec.name.0);
        specs
    }

    /// commit makes the root current when `last` is still the root, writing the chunks put since the last commit,
    /// and reports whether it did, as Dolt's NomsBlockStore.Commit does. The manifest is rewritten only when the set
    /// of files changes, since the journal records each root.
    pub fn commit(&mut self, current: Hash, last: Hash) -> Result<bool> {
        if self.upstream.root != last {
            return Ok(false);
        }
        if self.memtable.order.is_empty() && current == last {
            return Ok(true);
        }
        self.persist()?;
        if !current.is_empty() {
            self.check_refs(&[current])?;
        }
        let specs = self.specs();
        let next = Manifest {
            version: MANIFEST_VERSION.to_string(),
            format: self.upstream.format.clone(),
            lock: lock_hash(&current, &specs, &[], b""),
            root: current,
            gc_gen: self.upstream.gc_gen,
            specs,
        };
        let same_files = next.specs.len() == self.upstream.specs.len()
            && next.specs.iter().all(|spec| self.upstream.specs.iter().any(|s| s.name == spec.name));
        match self.journal.as_mut() {
            Some(journal) => {
                if !same_files {
                    write_manifest(&self.dir, &next)?;
                }
                journal.commit_root(current)?;
            }
            None => write_manifest(&self.dir, &next)?,
        }
        self.upstream = next;
        self.pending.clear();
        Ok(true)
    }

    /// dir returns the store's noms directory.
    pub fn dir(&self) -> &Path {
        &self.dir
    }

    /// rewrite replaces the store's files with the table files given, as garbage collection does: chunks not yet
    /// written stay pending, the journal is closed, the manifest names only those files at the current root, and the
    /// files it no longer names are deleted. The next write starts a new journal.
    pub fn rewrite(&mut self, specs: Vec<TableSpec>) -> Result<()> {
        if let Some(journal) = self.journal.take() {
            journal.close()?;
        }
        let manifest = crate::gc::replace_files(&self.dir, self.upstream.root, &self.upstream.format, specs)?;
        let sources = manifest.specs.iter().map(|spec| Source::open_file(&self.dir, &spec.name).map(Arc::new));
        self.sources = Arc::new(sources.collect::<Result<_>>()?);
        self.pending.clear();
        self.upstream = manifest;
        Ok(())
    }

    /// unwritten returns the addresses of the chunks put since the last commit, which garbage collection keeps along
    /// with what they refer to.
    pub fn unwritten(&self) -> Vec<Hash> {
        self.memtable.order.clone()
    }

    /// snapshot returns a read-only view of the chunks that the store's files hold now, leaving out the chunks put
    /// since the last commit, which garbage collection reads while the store goes on writing.
    pub fn snapshot(&mut self) -> Result<Snapshot> {
        let journal = match self.journal.as_mut() {
            Some(journal) => Some(journal.view()?),
            None => None,
        };
        Ok(Snapshot { sources: self.sources.clone(), journal })
    }

    /// sync writes out the journal's buffered records and index, leaving the store open.
    pub fn sync(&mut self) -> Result<()> {
        match self.journal.as_mut() {
            Some(journal) => journal.sync(),
            None => Ok(()),
        }
    }

    /// close writes out the journal's buffered records and index.
    pub fn close(self) -> Result<()> {
        match self.journal {
            Some(journal) => journal.close(),
            None => Ok(()),
        }
    }
}

/// write_manifest replaces the manifest through a synced temporary file, as Dolt's file manifest does.
pub(crate) fn write_manifest(dir: &Path, manifest: &Manifest) -> Result<()> {
    let temp = dir.join(format!("nbs_manifest_{}", Hash::of(manifest.format().as_bytes())));
    std::fs::write(&temp, manifest.format())?;
    File::open(&temp)?.sync_all()?;
    std::fs::rename(&temp, dir.join(MANIFEST_FILE))?;
    #[cfg(unix)]
    File::open(dir)?.sync_all()?;
    Ok(())
}
