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
use std::sync::Arc;

use prolly::{AddressMapSerializer, CommitClosureSerializer, Node, NodeStore, apply_mutations};
use serial::write::{CommitFields, WorkingSetFields, write_commit, write_store_root, write_working_set};
use serial::{Commit, Message, StoreRoot};
use store::{BlockStore, Chunk, ChunkReader, Hash, JournalStore};

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

/// Database is a chunk store whose store root names its datasets.
pub struct Database {
    store: JournalStore,
    old_gen: Option<BlockStore>,
    nodes: HashMap<Hash, Arc<Node>>,
}

impl ChunkReader for Database {
    fn get(&self, hash: &Hash) -> store::Result<Option<Chunk>> {
        if let Some(chunk) = self.store.get(hash)? {
            return Ok(Some(chunk));
        }
        match &self.old_gen {
            Some(old_gen) => old_gen.get(hash),
            None => Ok(None),
        }
    }
}

impl NodeStore for Database {
    fn read(&mut self, hash: &Hash) -> store::Result<Arc<Node>> {
        if let Some(node) = self.nodes.get(hash) {
            return Ok(node.clone());
        }
        let node = Arc::new(Node::load(self, hash)?);
        self.nodes.insert(*hash, node.clone());
        Ok(node)
    }

    fn write(&mut self, hash: Hash, bytes: Vec<u8>) -> store::Result<Arc<Node>> {
        let node = Arc::new(Node::decode(bytes.clone())?);
        self.put(Chunk { hash, data: bytes })?;
        self.nodes.insert(hash, node.clone());
        Ok(node)
    }
}

impl Database {
    /// open opens the database in a noms directory for writing, with its old generation for reading.
    pub fn open(noms: &Path) -> Result<Database> {
        let store = JournalStore::open(noms, "__DOLT__")?;
        let old_gen_dir = noms.join("oldgen");
        let old_gen =
            if old_gen_dir.join(store::MANIFEST_FILE).exists() { Some(BlockStore::open(&old_gen_dir)?) } else { None };
        Ok(Database { store, old_gen, nodes: HashMap::new() })
    }

    /// open_remote opens a file remote or backup in a directory for writing, as Dolt's FileFactory does: each commit
    /// writes its new chunks to a table file, and a missing old generation directory is made.
    pub fn open_remote(dir: &Path) -> Result<Database> {
        let store = JournalStore::open_tables(dir, "__DOLT__")?;
        let old_gen_dir = dir.join("oldgen");
        std::fs::create_dir_all(&old_gen_dir).map_err(store::Error::from)?;
        let old_gen =
            if old_gen_dir.join(store::MANIFEST_FILE).exists() { Some(BlockStore::open(&old_gen_dir)?) } else { None };
        Ok(Database { store, old_gen, nodes: HashMap::new() })
    }

    /// has reports whether the database holds the chunk.
    pub fn has(&self, hash: &Hash) -> bool {
        self.store.has(hash) || self.old_gen.as_ref().is_some_and(|old_gen| old_gen.has(hash))
    }

    /// pull copies the chunks reachable from an address that the database lacks from another database, children
    /// before the chunks that refer to them, taking a chunk the database holds to have everything it refers to, as
    /// Dolt's puller does.
    pub fn pull(&mut self, from: &Database, address: Hash) -> Result<()> {
        let mut seen = std::collections::HashSet::new();
        let mut stack = vec![(address, false)];
        let mut loaded: HashMap<Hash, Chunk> = HashMap::new();
        while let Some((hash, expanded)) = stack.pop() {
            if expanded {
                if let Some(chunk) = loaded.remove(&hash) {
                    self.put(chunk)?;
                }
                continue;
            }
            if hash.is_empty() || !seen.insert(hash) || self.has(&hash) {
                continue;
            }
            let chunk = from.require(&hash)?;
            stack.push((hash, true));
            serial::walk::walk_addrs(Message(&chunk.data), &mut |child| {
                stack.push((child, false));
                Ok(())
            })?;
            loaded.insert(hash, chunk);
        }
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
        while !self.store.commit(root, self.root())? {}
        Ok(())
    }

    /// set_heads points datasets at addresses already in the database, and deletes those without one, in one update
    /// of the store root.
    pub fn set_heads(&mut self, heads: &[(String, Option<Hash>)]) -> Result<()> {
        self.update(|_, _| Ok(heads.to_vec()))
    }

    /// root returns the address of the store root.
    pub fn root(&self) -> Hash {
        self.store.root()
    }

    /// put adds a chunk with the addresses its message refers to.
    fn put(&mut self, chunk: Chunk) -> store::Result<()> {
        let mut refs = Vec::new();
        serial::walk::walk_addrs(Message(&chunk.data), &mut |address| {
            refs.push(address);
            Ok(())
        })?;
        self.store.put(chunk, refs)
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
            if self.store.commit(store_root, root)? {
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

    /// close writes out the store's buffered journal records.
    pub fn close(self) -> Result<()> {
        Ok(self.store.close()?)
    }
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
