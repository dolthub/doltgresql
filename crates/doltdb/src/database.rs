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

//! A database's datasets, the named heads in its store root (branches, tags, working sets), written as Dolt's datas
//! package writes them, value by value in the same order.

use std::collections::HashMap;
use std::fmt;
use std::path::Path;
use std::sync::{Arc, Mutex, RwLock};

use prolly::{AddressMapSerializer, CommitClosureSerializer, Node, NodeStore, apply_mutations};
use serial::write::{CommitFields, WorkingSetFields, write_commit, write_store_root, write_working_set};
use serial::{Commit, Message, StoreRoot};
use store::{BlockStore, BuildAddrHasher, Chunk, ChunkReader, ChunkStore, Hash, JournalStore};

/// Error is a failure of a database operation.
#[derive(Debug)]
pub enum Error {
    Store(store::Error),
    /// A dataset moved since the caller read it, which Dolt calls ErrMergeNeeded.
    MergeNeeded,
    /// A working set moved since the caller read it, which Dolt calls ErrOptimisticLockFailed.
    OptimisticLockFailed,
    /// A commit is already a dataset's head, which Dolt calls ErrAlreadyCommitted.
    AlreadyCommitted,
    /// An operation is not valid.
    Invalid(String),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::Store(err) => write!(f, "{err}"),
            Error::MergeNeeded => write!(f, "dataset head is not ancestor of commit"),
            Error::OptimisticLockFailed => write!(f, "optimistic lock failed on database Root update"),
            Error::AlreadyCommitted => write!(f, "dataset head is already the commit"),
            Error::Invalid(message) => write!(f, "{message}"),
        }
    }
}

impl std::error::Error for Error {}

impl From<store::Error> for Error {
    fn from(err: store::Error) -> Error {
        Error::Store(err)
    }
}

/// Result is a database result.
pub type Result<T> = std::result::Result<T, Error>;

/// CommitMeta is who made a commit, when, and why.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CommitMeta {
    pub name: String,
    pub email: String,
    pub description: String,
    /// The author's time in Unix milliseconds.
    pub author_millis: i64,
    /// The committer's time in Unix milliseconds.
    pub committer_millis: u64,
    pub signature: String,
    /// The committer's name and email, when they differ from the author's.
    pub committer_name: Option<String>,
    pub committer_email: Option<String>,
}

/// PendingCommit is a commit to build: its root value, its parents besides the dataset's head, and its author.
#[derive(Clone, Debug)]
pub struct PendingCommit {
    pub root_value: Vec<u8>,
    pub parents: Vec<Hash>,
    pub meta: CommitMeta,
}

/// NewCommit is a commit built but not yet written.
#[derive(Clone, Debug)]
pub struct NewCommit {
    pub hash: Hash,
    pub bytes: Vec<u8>,
    pub height: u64,
}

/// commit_closure_key returns a commit's key in a commit closure: its little-endian height and its address.
fn commit_closure_key(height: u64, address: &Hash) -> Vec<u8> {
    let mut key = height.to_le_bytes().to_vec();
    key.extend_from_slice(&address.0);
    key
}

/// compare_commit_closure_keys orders commit closure keys by height and then by address, as Dolt's
/// commitClosureKeyOrdering does.
fn compare_commit_closure_keys(a: &[u8], b: &[u8]) -> std::cmp::Ordering {
    let height = |k: &[u8]| u64::from_le_bytes(k[..8].try_into().unwrap());
    height(a).cmp(&height(b)).then_with(|| a[8..].cmp(&b[8..]))
}

/// empty_node returns an empty leaf of the serializer's kind, which Dolt builds without writing.
fn empty_node(bytes: Vec<u8>) -> Result<Arc<Node>> {
    Ok(Arc::new(Node::decode(bytes)?))
}

/// GcMode is which garbage collection dolt_gc runs.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum GcMode {
    Default,
    Full,
    Shallow,
}

/// LOGGER writes a line to the server's log, once the server sets it.
pub static LOGGER: std::sync::OnceLock<fn(&str)> = std::sync::OnceLock::new();

/// MAX_TABLES is how many files a generation holds before its next manifest update conjoins them, as Dolt's
/// defaultMaxTables.
const MAX_TABLES: usize = 256;

/// log writes a line to the server's log.
fn log(line: &str) {
    if let Some(logger) = LOGGER.get() {
        logger(line);
    }
}

/// choose_conjoinees chooses the files to conjoin as Dolt's inlineConjoiner does: the smallest ones by chunk count,
/// for as long as their sum is above the next file's count or too many files would remain.
fn choose_conjoinees(specs: &[store::TableSpec]) -> Vec<store::TableSpec> {
    let mut sorted = specs.to_vec();
    sorted.sort_by_key(|spec| spec.chunk_count);
    let mut i = 2;
    let mut sum = sorted[0].chunk_count + sorted[1].chunk_count;
    while i < sorted.len() {
        let next = sorted[i].chunk_count;
        if sum <= next && sorted.len() - i < MAX_TABLES {
            break;
        }
        sum += next;
        i += 1;
    }
    sorted.truncate(i);
    sorted
}

/// GcConfig is how a garbage collection writes, as Dolt's chunks.GCConfig: its mode, whether it writes archives
/// rather than table files, and the size of the incremental files of leaf chunks, or 0 for none.
#[derive(Clone, Copy, Debug)]
pub struct GcConfig {
    pub mode: GcMode,
    pub archive: bool,
    pub incremental_file_size: u64,
}

/// CACHE_SIZE is how many bytes of chunks, and of decoded nodes, a database keeps in memory, as Dolt's node store
/// cache holds.
const CACHE_SIZE: usize = 256 << 20;

/// Cache keeps the values most recently used up to a total size, in two generations: when the young generation
/// fills, it becomes the old one and the old one is dropped, and a value used from the old generation moves back.
struct Cache<V> {
    young: HashMap<Hash, (V, usize), BuildAddrHasher>,
    old: HashMap<Hash, (V, usize), BuildAddrHasher>,
    young_size: usize,
}

/// SHARDS is how many separately locked parts each of a database's caches has, so that sessions rarely wait on each
/// other for one.
const SHARDS: usize = 16;

/// Caches is a cache split into separately locked shards by address.
struct Caches<V> {
    shards: Vec<Mutex<Cache<V>>>,
}

impl<V: Clone> Caches<V> {
    /// new returns empty caches.
    fn new() -> Caches<V> {
        Caches { shards: (0..SHARDS).map(|_| Mutex::new(Cache::new())).collect() }
    }

    /// shard returns the shard that caches the address.
    fn shard(&self, hash: &Hash) -> std::sync::MutexGuard<'_, Cache<V>> {
        self.shards[hash.0[0] as usize % SHARDS].lock().unwrap_or_else(|poisoned| poisoned.into_inner())
    }

    /// get returns the value cached at the address.
    fn get(&self, hash: &Hash) -> Option<V> {
        self.shard(hash).get(hash)
    }

    /// insert caches a value of a size at the address.
    fn insert(&self, hash: Hash, value: V, size: usize) {
        self.shard(&hash).insert(hash, value, size);
    }

    /// clear drops every cached value.
    fn clear(&self) {
        for shard in &self.shards {
            shard.lock().unwrap_or_else(|poisoned| poisoned.into_inner()).clear();
        }
    }
}

impl<V: Clone> Cache<V> {
    /// new returns an empty cache.
    fn new() -> Cache<V> {
        Cache { young: HashMap::default(), old: HashMap::default(), young_size: 0 }
    }

    /// get returns the value cached at the address.
    fn get(&mut self, hash: &Hash) -> Option<V> {
        if let Some((value, _)) = self.young.get(hash) {
            return Some(value.clone());
        }
        let (value, size) = self.old.remove(hash)?;
        self.insert(*hash, value.clone(), size);
        Some(value)
    }

    /// insert caches a value of a size at the address.
    fn insert(&mut self, hash: Hash, value: V, size: usize) {
        self.young_size += size;
        self.young.insert(hash, (value, size));
        if self.young_size > CACHE_SIZE / SHARDS / 2 {
            self.old = std::mem::take(&mut self.young);
            self.young_size = 0;
        }
    }

    /// clear drops every cached value.
    fn clear(&mut self) {
        self.young.clear();
        self.old.clear();
        self.young_size = 0;
    }
}

/// Database is a chunk store whose store root names its datasets. Its clones are handles on the same database, which
/// sessions use at the same time.
#[derive(Clone)]
pub struct Database {
    shared: Arc<Shared>,
}

/// Shared is what the handles of one database share.
struct Shared {
    store: Mutex<Box<dyn ChunkStore>>,
    old_gen: RwLock<Option<Arc<BlockStore>>>,
    nodes: Caches<Arc<Node>>,
    /// The chunks most recently read, decompressed.
    chunks: Caches<Arc<Vec<u8>>>,
    /// Whether a garbage collection is running, which its `GcRun` clears when it ends.
    collecting: Arc<std::sync::atomic::AtomicBool>,
}

impl ChunkReader for Database {
    fn get(&self, hash: &Hash) -> store::Result<Option<Chunk>> {
        if let Some(data) = self.shared.chunks.get(hash) {
            return Ok(Some(Chunk { hash: *hash, data: data.to_vec() }));
        }
        let plan = self.store().plan(hash)?;
        let chunk = match plan.read()? {
            Some(chunk) => Some(chunk),
            None => match self.old_gen() {
                Some(old_gen) => old_gen.get(hash)?,
                None => None,
            },
        };
        if let Some(chunk) = &chunk {
            self.shared.chunks.insert(*hash, Arc::new(chunk.data.clone()), chunk.data.len());
        }
        Ok(chunk)
    }

    fn get_many(&self, hashes: &[Hash]) -> store::Result<Vec<Option<Chunk>>> {
        let mut chunks = self.store().get_many(hashes)?;
        let old_gen = self.old_gen();
        for (hash, chunk) in hashes.iter().zip(chunks.iter_mut()) {
            if chunk.is_none()
                && let Some(old_gen) = &old_gen
            {
                *chunk = old_gen.get(hash)?;
            }
        }
        Ok(chunks)
    }
}

impl NodeStore for Database {
    fn read(&mut self, hash: &Hash) -> store::Result<Arc<Node>> {
        if let Some(node) = self.shared.nodes.get(hash) {
            return Ok(node);
        }
        let node = Arc::new(Node::load(self, hash)?);
        self.shared.nodes.insert(*hash, node.clone(), node.bytes().len());
        Ok(node)
    }

    fn write(&mut self, hash: Hash, bytes: Vec<u8>) -> store::Result<Arc<Node>> {
        let node = Arc::new(Node::decode(bytes.clone())?);
        self.put(Chunk { hash, data: bytes })?;
        self.shared.nodes.insert(hash, node.clone(), node.bytes().len());
        Ok(node)
    }
}

impl Database {
    /// open opens the database in a noms directory for writing, with its old generation for reading.
    pub fn open(noms: &Path) -> Result<Database> {
        let store = Box::new(JournalStore::open(noms, "__DOLT__")?);
        let old_gen_dir = noms.join("oldgen");
        store::remove_spills(noms);
        store::remove_spills(&old_gen_dir);
        let old_gen =
            if old_gen_dir.join(store::MANIFEST_FILE).exists() { Some(BlockStore::open(&old_gen_dir)?) } else { None };
        Ok(Database::new(store, old_gen))
    }

    /// open_remote opens a file remote or backup in a directory for writing, as Dolt's FileFactory does: each commit
    /// writes its new chunks to a table file, and a missing old generation directory is made.
    pub fn open_remote(dir: &Path) -> Result<Database> {
        let store = Box::new(JournalStore::open_tables(dir, "__DOLT__")?);
        let old_gen_dir = dir.join("oldgen");
        std::fs::create_dir_all(&old_gen_dir).map_err(store::Error::from)?;
        let old_gen =
            if old_gen_dir.join(store::MANIFEST_FILE).exists() { Some(BlockStore::open(&old_gen_dir)?) } else { None };
        Ok(Database::new(store, old_gen))
    }

    /// with_store opens a database over another kind of chunk store, such as a remote, with no old generation.
    pub fn with_store(store: Box<dyn ChunkStore>) -> Database {
        Database::new(store, None)
    }

    /// new returns a database over a chunk store and an old generation.
    fn new(store: Box<dyn ChunkStore>, old_gen: Option<BlockStore>) -> Database {
        Database {
            shared: Arc::new(Shared {
                store: Mutex::new(store),
                old_gen: RwLock::new(old_gen.map(Arc::new)),
                nodes: Caches::new(),
                chunks: Caches::new(),
                collecting: Arc::default(),
            }),
        }
    }

    /// store locks the database's chunk store.
    fn store(&self) -> std::sync::MutexGuard<'_, Box<dyn ChunkStore>> {
        self.shared.store.lock().unwrap_or_else(|poisoned| poisoned.into_inner())
    }

    /// old_gen returns the database's old generation, if it has one.
    fn old_gen(&self) -> Option<Arc<BlockStore>> {
        self.shared.old_gen.read().unwrap_or_else(|poisoned| poisoned.into_inner()).clone()
    }

    /// set_old_gen replaces the database's old generation.
    fn set_old_gen(&self, old_gen: Option<BlockStore>) {
        *self.shared.old_gen.write().unwrap_or_else(|poisoned| poisoned.into_inner()) = old_gen.map(Arc::new);
    }

    /// with_journal runs a function on the database's local journaling store, failing for a database over another
    /// kind of store.
    fn with_journal<T>(&self, f: impl FnOnce(&mut JournalStore) -> Result<T>) -> Result<T> {
        let mut store = self.store();
        f(store.journal().ok_or_else(|| Error::Invalid("not a local database".into()))?)
    }

    /// noms_dir returns the directory of the database's local store.
    pub fn noms_dir(&mut self) -> Result<std::path::PathBuf> {
        self.with_journal(|journal| Ok(journal.dir().to_path_buf()))
    }

    /// locate returns where a committed chunk is in the files of the local store, by its path relative to the store's
    /// directory, as Dolt's GetChunkLocationsWithPaths finds it.
    pub fn locate(&mut self, hash: &Hash) -> Result<Option<store::Location>> {
        if let Some(location) = self.with_journal(|journal| Ok(journal.locate(hash)?))? {
            return Ok(Some(location));
        }
        Ok(self
            .old_gen()
            .as_ref()
            .and_then(|old_gen| old_gen.locate(hash))
            .map(|location| store::Location { file: format!("oldgen/{}", location.file), ..location }))
    }

    /// table_files returns the root and the files of both generations, by their paths relative to the store's
    /// directory, with their chunk counts, as Dolt's Sources lists them.
    pub fn table_files(&mut self) -> Result<(Hash, Vec<(String, u32)>)> {
        let root = self.root();
        let dir = self.noms_dir()?;
        let name = |dir: &Path, spec: &store::TableSpec| match dir.join(format!("{}.darc", spec.name)).exists() {
            true => format!("{}.darc", spec.name),
            false => spec.name.to_string(),
        };
        let mut files: Vec<(String, u32)> = self.with_journal(|journal| {
            Ok(journal.table_files().iter().map(|spec| (name(&dir, spec), spec.chunk_count)).collect())
        })?;
        let old_dir = dir.join("oldgen");
        for spec in store::Manifest::read(&old_dir)?.map(|m| m.specs).unwrap_or_default() {
            files.push((format!("oldgen/{}", name(&old_dir, &spec)), spec.chunk_count));
        }
        Ok((root, files))
    }

    /// add_table_files adds uploaded table files or archives in the local store's directory to the files the next
    /// commit names.
    pub fn add_table_files(&mut self, specs: &[store::TableSpec]) -> Result<()> {
        self.with_journal(|journal| Ok(journal.add_table_files(specs)?))
    }

    /// has reports whether the database holds the chunk.
    pub fn has(&self, hash: &Hash) -> bool {
        let held = self.store().has(hash);
        held || self.old_gen().is_some_and(|old_gen| old_gen.has(hash))
    }

    /// pull copies the chunks reachable from an address that the database lacks from another database, children
    /// before the chunks that refer to them, taking a chunk the database holds to have everything it refers to, as
    /// Dolt's puller does.
    pub fn pull(&mut self, from: &Database, address: Hash) -> Result<()> {
        let mut seen = std::collections::HashSet::<Hash, BuildAddrHasher>::default();
        let mut levels = Vec::new();
        let mut frontier = vec![address];
        while !frontier.is_empty() {
            frontier.retain(|hash| !hash.is_empty() && seen.insert(*hash));
            let present = self.has_many(&frontier);
            let wanted: Vec<Hash> = frontier.iter().zip(present).filter(|(_, held)| !held).map(|(h, _)| *h).collect();
            let mut next = Vec::new();
            let mut level = Vec::with_capacity(wanted.len());
            for (hash, chunk) in wanted.iter().zip(from.get_many(&wanted)?) {
                let chunk = chunk.ok_or_else(|| store::Error::Corrupt(format!("chunk {hash} is missing")))?;
                serial::walk::walk_addrs(Message(&chunk.data), &mut |child| {
                    next.push(child);
                    Ok(())
                })?;
                level.push(chunk);
            }
            levels.push(level);
            frontier = next;
        }
        for chunk in levels.into_iter().rev().flatten() {
            self.put(chunk)?;
        }
        Ok(())
    }

    /// has_many reports whether the database holds each chunk, in the order asked for.
    pub fn has_many(&self, hashes: &[Hash]) -> Vec<bool> {
        let held = self.store().has_many(hashes);
        let old_gen = self.old_gen();
        hashes
            .iter()
            .zip(held)
            .map(|(hash, held)| held || old_gen.as_ref().is_some_and(|old_gen| old_gen.has(hash)))
            .collect()
    }

    /// commit_root moves the store root from the last root to the current one, reporting false when it moved first,
    /// as a remote's Commit request asks.
    pub fn commit_root(&mut self, current: Hash, last: Hash) -> Result<bool> {
        Ok(self.store().commit(current, last)?)
    }

    /// gc keeps only the chunks reachable from the store root, as Dolt's garbage collection does, starting from the old
    /// generation's files as they are now, since an interrupted collection may have added some: a shallow collection
    /// rewrites the new generation's chunks to one table file, and the others move the chunks that commits reach from
    /// the new generation to the old generation, keeping the chunks that only working sets reach in a new generation
    /// table file, where a full collection also rewrites the old generation's chunks. It also keeps the chunks that
    /// the given addresses reach, with the working sets' chunks.
    pub fn gc(&mut self, config: GcConfig, keep: Vec<Hash>) -> Result<()> {
        let mut run = self.gc_begin(config)?;
        run.copy()?;
        self.gc_finish(run, keep, &[])
    }

    /// gc_begin starts a garbage collection as `gc` describes, noting the roots it keeps and taking a snapshot of the
    /// store's files that `GcRun::copy` reads while the database goes on.
    pub fn gc_begin(&mut self, config: GcConfig) -> Result<GcRun> {
        if self.shared.collecting.swap(true, std::sync::atomic::Ordering::AcqRel) {
            return Err(Error::Invalid("a garbage collection is already running".into()));
        }
        let collecting = Collecting(self.shared.collecting.clone());
        let dir = self.noms_dir()?;
        let old_dir = dir.join("oldgen");
        if old_dir.join(store::MANIFEST_FILE).exists() {
            self.set_old_gen(Some(BlockStore::open(&old_dir)?));
        }
        let committed: Vec<Hash> = self
            .datasets()?
            .into_iter()
            .filter(|(name, _)| !name.starts_with("workingSets/"))
            .map(|(_, h)| h)
            .collect();
        let old_gen = match old_dir.join(store::MANIFEST_FILE).exists() {
            true => Some(BlockStore::open(&old_dir)?),
            false => None,
        };
        Ok(GcRun {
            _collecting: collecting,
            config,
            root: self.root(),
            committed,
            new_gen: self.with_journal(|journal| Ok(journal.snapshot()?))?,
            old_gen,
            dir,
            seen: std::collections::HashSet::default(),
            old_specs: None,
            new_gen_writer: None,
        })
    }

    /// gc_finish ends a garbage collection that `GcRun::copy` ran: it copies the chunks that the store root, the
    /// given addresses, the addresses within the given root values, and the chunks not yet written reach and the copy
    /// did not see, which were written since it began, and then replaces the store's files with the collection's.
    pub fn gc_finish(&mut self, mut run: GcRun, keep: Vec<Hash>, roots: &[Vec<u8>]) -> Result<()> {
        let mut starts = vec![self.root()];
        starts.extend(keep);
        starts.extend(self.with_journal(|journal| Ok(journal.unwritten()))?);
        for root in roots {
            serial::walk::walk_addrs(Message(root), &mut |child| {
                starts.push(child);
                Ok(())
            })?;
        }
        let full = run.config.mode == GcMode::Full;
        let mut late = match run.new_gen_writer.take() {
            Some(writer) => writer,
            None => store::GcWriter::new(&run.dir, run.config.archive, 0)?,
        };
        let old_gen = run.old_gen.as_ref();
        walk(&*self, starts, &mut run.seen, &mut |chunk, stored, leaf| match full
            || !old_gen.is_some_and(|g| g.has(&chunk.hash))
        {
            true => late.add(chunk, stored, leaf, &mut |_| Ok(())).map_err(Error::from),
            false => Ok(()),
        })?;
        let new_specs = late.finish(&mut |_| Ok(()))?;
        if let Some(specs) = run.old_specs {
            let old_dir = run.dir.join("oldgen");
            self.set_old_gen(None);
            store::replace_files(&old_dir, self.root(), "__DOLT__", specs)?;
            self.set_old_gen(Some(BlockStore::open(&old_dir)?));
        }
        if let Some(old_gen) = self.old_gen() {
            self.with_journal(|journal| {
                journal.forget_refs(&|hash| old_gen.has(hash));
                Ok(())
            })?;
        }
        self.with_journal(|journal| Ok(journal.rewrite(new_specs)?))?;
        self.shared.nodes.clear();
        self.shared.chunks.clear();
        Ok(())
    }

    /// address_map builds an address map of names to addresses, writing any nodes below its root, and returns its
    /// root node's bytes, which messages such as stash lists embed.
    pub fn address_map(&mut self, entries: &[(String, Hash)]) -> Result<Vec<u8>> {
        let empty = empty_node(prolly::serialize_address_map(&[], &[], &[], 0))?;
        let mut edits: Vec<(Vec<u8>, Option<Vec<u8>>)> =
            entries.iter().map(|(name, address)| (name.clone().into_bytes(), Some(address.0.to_vec()))).collect();
        edits.sort();
        let (_, map) = apply_mutations(self, empty, AddressMapSerializer, edits, &|a: &[u8], b: &[u8]| a.cmp(b))?;
        Ok(map.bytes().to_vec())
    }

    /// address_map_entries returns the names and addresses of an address map, given its root node's bytes, in name
    /// order.
    pub fn address_map_entries(&mut self, bytes: &[u8]) -> Result<Vec<(String, Hash)>> {
        if bytes.is_empty() {
            return Ok(Vec::new());
        }
        let node = Node::decode(bytes.to_vec())?;
        let mut entries = Vec::new();
        prolly::walk_leaves(self, &node, &mut |key, value| {
            entries.push((String::from_utf8_lossy(key).into_owned(), serial::hash(value)?));
            Ok(())
        })?;
        Ok(entries)
    }

    /// replace_root makes a store root already in the database current, whatever the root was, as Dolt's CommitRoot
    /// does when it syncs one database to another.
    pub fn replace_root(&mut self, root: Hash) -> Result<()> {
        loop {
            let current = self.root();
            if self.store().commit(root, current)? {
                return Ok(());
            }
        }
    }

    /// set_heads points datasets at addresses already in the database, and deletes those without one, in one update
    /// of the store root.
    pub fn set_heads(&mut self, heads: &[(String, Option<Hash>)]) -> Result<()> {
        self.update(|_, _| Ok(heads.to_vec()))
    }

    /// root returns the address of the store root.
    pub fn root(&self) -> Hash {
        self.store().root()
    }

    /// put adds a chunk with the addresses its message refers to, leaving out the ones the old generation holds, which
    /// the store cannot see.
    fn put(&mut self, chunk: Chunk) -> store::Result<()> {
        let (mut refs, old_gen) = (Vec::new(), self.old_gen());
        serial::walk::walk_addrs(Message(&chunk.data), &mut |address| {
            if !old_gen.as_ref().is_some_and(|old_gen| old_gen.has(&address)) {
                refs.push(address);
            }
            Ok(())
        })?;
        self.store().put(chunk, refs)
    }

    /// write_value writes a message, as Dolt's ValueStore.WriteValue does, and returns its address.
    pub fn write_value(&mut self, data: Vec<u8>) -> Result<Hash> {
        let chunk = Chunk::new(data);
        let hash = chunk.hash;
        self.put(chunk)?;
        Ok(hash)
    }

    /// read_value returns the message at the address, if the database holds it.
    pub fn read_value(&self, hash: &Hash) -> Result<Option<Vec<u8>>> {
        Ok(self.get(hash)?.map(|chunk| chunk.data))
    }

    /// refs_node returns the root node of the datasets map in the store root at the address.
    fn refs_node(&mut self, root: Hash) -> Result<Arc<Node>> {
        if root.is_empty() {
            return empty_node(prolly::serialize_address_map(&[], &[], &[], 0));
        }
        let data = self.require(&root)?.data;
        let map = StoreRoot::new(Message(&data))?.address_map()?.unwrap_or_default().to_vec();
        if map.is_empty() {
            return empty_node(prolly::serialize_address_map(&[], &[], &[], 0));
        }
        Ok(Arc::new(Node::decode(map)?))
    }

    /// datasets returns every dataset's name and head in name order.
    pub fn datasets(&mut self) -> Result<Vec<(String, Hash)>> {
        let node = self.refs_node(self.root())?;
        let mut entries = Vec::new();
        prolly::walk_leaves(self, &node, &mut |key, value| {
            entries.push((String::from_utf8_lossy(key).into_owned(), serial::hash(value)?));
            Ok(())
        })?;
        Ok(entries)
    }

    /// head returns the dataset's head, if it has one.
    pub fn head(&mut self, id: &str) -> Result<Option<Hash>> {
        Ok(self.datasets()?.into_iter().find(|(name, _)| name == id).map(|(_, hash)| hash))
    }

    /// update edits the datasets map and commits a new store root, retrying when another writer moved the root, as
    /// Dolt's database.update does. The edit returns the datasets to set, or to delete when they have no address.
    fn update(
        &mut self,
        mut edit: impl FnMut(&mut Database, &[(String, Hash)]) -> Result<Vec<(String, Option<Hash>)>>,
    ) -> Result<()> {
        loop {
            let root = self.root();
            let current = self.datasets()?;
            let mut edits = edit(self, &current)?;
            edits.sort_by(|a, b| a.0.cmp(&b.0));
            let node = self.refs_node(root)?;
            let edits: Vec<(Vec<u8>, Option<Vec<u8>>)> =
                edits.into_iter().map(|(k, v)| (k.into_bytes(), v.map(|h| h.0.to_vec()))).collect();
            let (_, map) = apply_mutations(self, node, AddressMapSerializer, edits, &|a: &[u8], b: &[u8]| a.cmp(b))?;
            let store_root = self.write_value(write_store_root(map.bytes()))?;
            if self.store().commit(store_root, root)? {
                return Ok(());
            }
        }
    }

    /// commit_closure writes the closure of a new commit's parents, the union of their closures and the parents
    /// themselves, as Dolt's writeFbCommitParentClosure does, and returns its address, which is empty without
    /// parents.
    fn commit_closure(&mut self, parents: &[Hash]) -> Result<Hash> {
        if parents.is_empty() {
            return Ok(Hash::default());
        }
        let mut commits = Vec::new();
        for parent in parents {
            let data = self.require(parent)?.data;
            let commit = Commit::new(Message(&data))?;
            let closure = commit.parent_closure_bytes()?.map(serial::hash).transpose()?.unwrap_or_default();
            commits.push((commit.height()?, closure));
        }
        let closures: Vec<Arc<Node>> = commits
            .iter()
            .map(|(_, closure)| {
                if closure.is_empty() {
                    empty_node(prolly::serialize_commit_closure(&[], &[], &[], 0))
                } else {
                    Ok(self.read(closure)?)
                }
            })
            .collect::<Result<_>>()?;
        // New keys carry Dolt's one-byte empty closure value, which counts toward where nodes end.
        let mut edits: Vec<(Vec<u8>, Option<Vec<u8>>)> = Vec::new();
        let first = closure_keys(self, &closures[0])?;
        for closure in &closures[1..] {
            for key in closure_keys(self, closure)? {
                if first.binary_search_by(|k| compare_commit_closure_keys(k, &key)).is_err() {
                    edits.push((key, Some(vec![0])));
                }
            }
        }
        for (parent, (height, _)) in parents.iter().zip(&commits) {
            edits.push((commit_closure_key(*height, parent), Some(vec![0])));
        }
        edits.sort_by(|a, b| compare_commit_closure_keys(&a.0, &b.0));
        edits.dedup_by(|a, b| a.0 == b.0);
        let (hash, _) =
            apply_mutations(self, closures[0].clone(), CommitClosureSerializer, edits, &compare_commit_closure_keys)?;
        Ok(hash)
    }

    /// build_commit writes the root value and the parents' closure, and builds a commit of the root value on the
    /// dataset's head, as Dolt's BuildNewCommit does. Without explicit parents, the head is the parent.
    pub fn build_commit(
        &mut self,
        head: Option<Hash>,
        root_value: Vec<u8>,
        mut parents: Vec<Hash>,
        meta: &CommitMeta,
    ) -> Result<NewCommit> {
        if let Some(head) = head {
            if parents.is_empty() {
                parents.push(head);
            } else if !parents.contains(&head) {
                return Err(Error::MergeNeeded);
            }
        }
        let root = self.write_value(root_value)?;
        let mut max_height = 0;
        for parent in &parents {
            let data = self.require(parent)?.data;
            max_height = max_height.max(Commit::new(Message(&data))?.height()?);
        }
        let parent_closure = self.commit_closure(&parents)?;
        let bytes = write_commit(&CommitFields {
            root,
            height: max_height + 1,
            parents,
            parent_closure,
            name: meta.name.clone().into_bytes(),
            email: meta.email.clone().into_bytes(),
            description: meta.description.clone().into_bytes(),
            timestamp_millis: meta.committer_millis,
            user_timestamp_millis: meta.author_millis,
            signature: meta.signature.clone().into_bytes(),
            committer_name: meta.committer_name.clone().map(String::into_bytes),
            committer_email: meta.committer_email.clone().map(String::into_bytes),
        });
        Ok(NewCommit { hash: Hash::of(&bytes), bytes, height: max_height + 1 })
    }

    /// write_commit writes the commit and makes it the dataset's head, failing when the head moved from `head`, as
    /// Dolt's WriteCommit does.
    pub fn write_commit(&mut self, dataset: &str, head: Option<Hash>, commit: &NewCommit) -> Result<()> {
        self.write_value(commit.bytes.clone())?;
        let id = dataset.to_string();
        self.update(|_, current| {
            let curr = current.iter().find(|(name, _)| *name == id).map(|(_, h)| *h);
            if curr != head {
                return Err(Error::MergeNeeded);
            }
            if curr == Some(commit.hash) {
                return Err(Error::AlreadyCommitted);
            }
            Ok(vec![(id.clone(), Some(commit.hash))])
        })
    }

    /// set_head points the dataset at a commit or tag already in the database, as Dolt's SetHead does without a
    /// working set path.
    pub fn set_head(&mut self, dataset: &str, address: Hash) -> Result<()> {
        if self.read_value(&address)?.is_none() {
            return Err(Error::Invalid(
                "SetHead failed: attempt to set a dataset head to an address which is not in the store".into(),
            ));
        }
        let id = dataset.to_string();
        self.update(|_, _| Ok(vec![(id.clone(), Some(address))]))
    }

    /// delete_heads removes the datasets.
    pub fn delete_heads(&mut self, datasets: &[String]) -> Result<()> {
        self.update(|_, current| {
            Ok(datasets
                .iter()
                .filter(|d| current.iter().any(|(name, _)| name == *d))
                .map(|d| (d.clone(), None))
                .collect())
        })
    }

    /// update_working_set writes the working set and makes it the dataset's head, failing when the head moved from
    /// `previous`, as Dolt's UpdateWorkingSet does, and returns its address.
    pub fn update_working_set(
        &mut self,
        dataset: &str,
        working_set: &WorkingSetFields,
        previous: Hash,
    ) -> Result<Hash> {
        let address = self.write_value(write_working_set(working_set))?;
        let id = dataset.to_string();
        self.update(|_, current| {
            let curr = current.iter().find(|(name, _)| *name == id).map(|(_, h)| *h).unwrap_or_default();
            if curr != previous {
                return Err(Error::OptimisticLockFailed);
            }
            Ok(vec![(id.clone(), Some(address))])
        })?;
        Ok(address)
    }

    /// commit_with_working_set writes the working set and a commit of the root value, then moves the commit dataset
    /// and the working set dataset together, as Dolt's CommitWithWorkingSet does, and returns the commit.
    pub fn commit_with_working_set(
        &mut self,
        commit_dataset: &str,
        working_set_dataset: &str,
        working_set: &WorkingSetFields,
        previous_working_set: Hash,
        pending: PendingCommit,
    ) -> Result<NewCommit> {
        let PendingCommit { root_value, mut parents, meta } = pending;
        let working_set_address = self.write_value(write_working_set(working_set))?;
        let head = self.head(commit_dataset)?;
        if let Some(head) = head
            && !parents.is_empty()
            && !parents.contains(&head)
        {
            parents.insert(0, head);
        }
        let commit = self.build_commit(head, root_value, parents, &meta)?;
        self.write_value(commit.bytes.clone())?;
        let (commit_id, working_set_id) = (commit_dataset.to_string(), working_set_dataset.to_string());
        self.update(|_, current| {
            let find = |id: &str| current.iter().find(|(name, _)| name == id).map(|(_, h)| *h);
            if find(&working_set_id).unwrap_or_default() != previous_working_set {
                return Err(Error::OptimisticLockFailed);
            }
            if find(&commit_id) != head {
                return Err(Error::MergeNeeded);
            }
            Ok(vec![(commit_id.clone(), Some(commit.hash)), (working_set_id.clone(), Some(working_set_address))])
        })?;
        Ok(commit)
    }

    /// sync writes out the store's buffered journal records, leaving the database open.
    pub fn sync(&mut self) -> Result<()> {
        match self.store().journal() {
            Some(journal) => Ok(journal.sync()?),
            None => Ok(()),
        }
    }

    /// close writes out the store's buffered journal records.
    pub fn close(mut self) -> Result<()> {
        self.sync()
    }
}

/// GcRun is a garbage collection that `Database::gc_begin` started: the roots it keeps, a snapshot of the store's
/// files as they were then, and the files that `copy` wrote for `Database::gc_finish` to swap in.
pub struct GcRun {
    _collecting: Collecting,
    config: GcConfig,
    root: Hash,
    /// The heads of the datasets that are not working sets, whose chunks move to the old generation.
    committed: Vec<Hash>,
    new_gen: store::Snapshot,
    old_gen: Option<BlockStore>,
    dir: std::path::PathBuf,
    /// The addresses the collection has read.
    seen: std::collections::HashSet<Hash, BuildAddrHasher>,
    /// The old generation's files after the collection, or None when they stay as they are.
    old_specs: Option<Vec<store::TableSpec>>,
    /// The writer of the new generation's file, which the chunks written during the copy join before it finishes.
    new_gen_writer: Option<store::GcWriter>,
}

impl GcRun {
    /// copy writes the chunks that the collection's roots reach to new files, reading the snapshot rather than the
    /// database, so that the database can go on meanwhile.
    pub fn copy(&mut self) -> Result<()> {
        let mode = self.config.mode;
        let old_dir = self.dir.join("oldgen");
        let reader = GcReader { new_gen: &self.new_gen, old_gen: self.old_gen.as_ref() };
        let old_gen = self.old_gen.as_ref();
        let in_old_gen = |hash: &Hash| old_gen.is_some_and(|old_gen| old_gen.has(hash));
        let (archive, size, root) = (self.config.archive, self.config.incremental_file_size, self.root);
        if mode == GcMode::Shallow {
            let mut writer = store::GcWriter::new(&self.dir, false, 0)?;
            let starts = [self.committed.clone(), vec![root]].concat();
            walk(&reader, starts, &mut self.seen, &mut |chunk, stored, leaf| match in_old_gen(&chunk.hash) {
                true => Ok(()),
                false => writer.add(chunk, stored, leaf, &mut |_| Ok(())).map_err(Error::from),
            })?;
            self.new_gen_writer = Some(writer);
            return Ok(());
        }
        let mut specs = match (mode, store::Manifest::read(&old_dir)?) {
            (GcMode::Default, Some(manifest)) => manifest.specs,
            _ => Vec::new(),
        };
        let database = self.dir.parent().and_then(Path::parent).and_then(Path::file_name);
        let database = database.map(|n| n.to_string_lossy().into_owned()).unwrap_or_default();
        if specs.len() > MAX_TABLES {
            specs = conjoin(&old_dir, specs, &database)?;
        } else if mode == GcMode::Full && store::Manifest::read(&old_dir)?.is_some_and(|m| m.specs.len() > MAX_TABLES) {
            log(&format!(
                "level=info msg=\"conjoin dynamically disabled. not conjoining.\" database={database} generation=old \
                 pkg=store.noms"
            ));
        }
        let mut add = |spec: &store::TableSpec| match mode {
            GcMode::Default => store::add_to_manifest(&old_dir, root, "__DOLT__", spec),
            _ => Ok(()),
        };
        let full = mode == GcMode::Full;
        let mut moved = store::GcWriter::new(&old_dir, archive, size)?;
        walk(&reader, self.committed.clone(), &mut self.seen, &mut |chunk, stored, leaf| match full
            || !in_old_gen(&chunk.hash)
        {
            true => moved.add(chunk, stored, leaf, &mut add).map_err(Error::from),
            false => Ok(()),
        })?;
        let mut working = store::GcWriter::new(&self.dir, archive, size)?;
        walk(&reader, vec![root], &mut self.seen, &mut |chunk, stored, leaf| match in_old_gen(&chunk.hash) {
            true if full => moved.add(chunk, stored, leaf, &mut add).map_err(Error::from),
            true => Ok(()),
            false => working.add(chunk, stored, leaf, &mut |_| Ok(())).map_err(Error::from),
        })?;
        specs.extend(moved.finish(&mut add)?);
        self.old_specs = Some(specs);
        self.new_gen_writer = Some(working);
        Ok(())
    }
}

/// Collecting marks a database's garbage collection as running until it is dropped.
struct Collecting(Arc<std::sync::atomic::AtomicBool>);

impl Drop for Collecting {
    fn drop(&mut self) {
        self.0.store(false, std::sync::atomic::Ordering::Release);
    }
}

/// GcReader reads a garbage collection's snapshot of the new generation, then the old generation.
struct GcReader<'a> {
    new_gen: &'a store::Snapshot,
    old_gen: Option<&'a BlockStore>,
}

impl ChunkReader for GcReader<'_> {
    fn get(&self, hash: &Hash) -> store::Result<Option<Chunk>> {
        match self.new_gen.get(hash)? {
            Some(chunk) => Ok(Some(chunk)),
            None => match self.old_gen {
                Some(old_gen) => old_gen.get(hash),
                None => Ok(None),
            },
        }
    }

    fn get_stored(&self, hash: &Hash) -> store::Result<Option<(Chunk, Option<store::Stored>)>> {
        match self.new_gen.get_stored(hash)? {
            Some(found) => Ok(Some(found)),
            None => match self.old_gen {
                Some(old_gen) => old_gen.get_stored(hash),
                None => Ok(None),
            },
        }
    }
}

/// walk visits the chunks that the addresses reach and the seen set has not yet seen, adding them to it, each with
/// its stored form in an archive when it has one and whether it is a leaf, which refers to no other chunk.
fn walk(
    reader: &dyn ChunkReader,
    starts: Vec<Hash>,
    seen: &mut std::collections::HashSet<Hash, BuildAddrHasher>,
    visit: &mut dyn FnMut(Chunk, Option<store::Stored>, bool) -> Result<()>,
) -> Result<()> {
    let mut stack = starts;
    while let Some(hash) = stack.pop() {
        if hash.is_empty() || !seen.insert(hash) {
            continue;
        }
        let missing = || store::Error::Corrupt(format!("chunk {hash} is missing"));
        let (chunk, stored) = reader.get_stored(&hash)?.ok_or_else(missing)?;
        let mut leaf = true;
        serial::walk::walk_addrs(Message(&chunk.data), &mut |child| {
            leaf = false;
            stack.push(child);
            Ok(())
        })?;
        visit(chunk, stored, leaf)?;
    }
    Ok(())
}

/// conjoin writes the chunks of the files that Dolt's conjoiner chooses among an old generation's files to one file,
/// returning the generation's files with that one in their place, and logging as Dolt does.
fn conjoin(dir: &Path, specs: Vec<store::TableSpec>, database: &str) -> Result<Vec<store::TableSpec>> {
    log(&format!(
        "level=info msg=\"beginning conjoin of database\" database={database} generation=old pkg=store.noms \
         upstream_len={}",
        specs.len()
    ));
    let chosen = choose_conjoinees(&specs);
    let mut chunks = Vec::new();
    let mut archive = false;
    for spec in &chosen {
        let path = dir.join(format!("{}.darc", spec.name));
        let mut add = |chunk: Chunk| {
            chunks.push((chunk, false));
            Ok(())
        };
        if path.exists() {
            archive = true;
            store::ArchiveReader::open(&path)?.for_each(&mut add)?;
        } else {
            store::TableReader::open(&dir.join(spec.name.to_string()))?.for_each(&mut add)?;
        }
    }
    let mut conjoined = store::write_files(dir, chunks, archive, 0, &mut |_| Ok(()))?;
    conjoined.extend(specs.into_iter().filter(|spec| !chosen.iter().any(|c| c.name == spec.name)));
    log(&format!(
        "level=info msg=\"conjoin completed successfully\" database={database} generation=old \
         new_upstream_len={} pkg=store.noms",
        conjoined.len()
    ));
    Ok(conjoined)
}

/// closure_keys returns the keys of the commit closure at the node in order.
fn closure_keys(db: &mut Database, node: &Node) -> Result<Vec<Vec<u8>>> {
    let mut keys = Vec::new();
    prolly::walk_leaves(db, node, &mut |key, _| {
        keys.push(key.to_vec());
        Ok(())
    })?;
    Ok(keys)
}
