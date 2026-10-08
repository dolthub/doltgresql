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

#![forbid(unsafe_code)]

//! Dolt's chunk storage, which holds content-addressed chunks in table files, archives, and the chunk journal of a
//! database's noms directory, with a manifest naming the files and the root.

use std::sync::Arc;

mod archive;
mod blob;
mod chunk;
mod error;
mod file;
mod gc;
mod hash;
mod journal;
mod journal_store;
mod journal_writer;
mod manifest;
mod memory;
mod store;
mod table;

pub use archive::{ArchiveReader, ArchiveWriter, Stored};
pub use blob::{Blob, BlobChunkStore, BlobRange, Blobstore, LocalBlobstore, MANIFEST_KEY, not_found};
pub use chunk::Chunk;
pub use error::{Error, Result};
pub use file::{ReadAt, remove_spills};
pub use gc::{GcWriter, add_to_manifest, replace_files, write_files, write_table};
pub use hash::Hash;
pub use journal::{JOURNAL_FILE, JournalRecord, read_records};
pub use journal_store::{JournalStore, Snapshot};
pub use journal_writer::{JOURNAL_INDEX_FILE, JournalView, JournalWriter, PendingSync, defer_syncs, take_sync};
pub use manifest::{MANIFEST_FILE, Manifest, TableSpec, lock_hash};
pub use memory::MemoryStore;
pub use store::{BlockStore, GenerationalStore};
pub use table::{TableReader, TableWriter};

/// Location is where a chunk's compressed bytes are in a store's files: the file's name, the offset and length of the
/// chunk's record or data span, and for an archive chunk compressed with a dictionary, the offset and length of the
/// dictionary's span.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Location {
    pub file: String,
    pub offset: u64,
    pub length: u32,
    pub dictionary: Option<(u64, u32)>,
}

/// ChunkReader reads chunks by address.
pub trait ChunkReader {
    /// get returns the chunk when the reader holds it.
    fn get(&self, hash: &Hash) -> Result<Option<Chunk>>;

    /// get_stored returns the chunk when the reader holds it, with its stored form when an archive's dictionary
    /// compressed it.
    fn get_stored(&self, hash: &Hash) -> Result<Option<(Chunk, Option<Stored>)>> {
        Ok(self.get(hash)?.map(|chunk| (chunk, None)))
    }

    /// require returns the chunk, failing when the reader lacks it.
    fn require(&self, hash: &Hash) -> Result<Chunk> {
        self.get(hash)?.ok_or_else(|| Error::Corrupt(format!("chunk {hash} is missing")))
    }

    /// get_many returns each chunk the reader holds, in the order asked for, which a remote reader fetches together.
    fn get_many(&self, hashes: &[Hash]) -> Result<Vec<Option<Chunk>>> {
        hashes.iter().map(|hash| self.get(hash)).collect()
    }
}

/// ChunkStore is a writable chunk store whose root a compare-and-set commit moves: a local journaling store, or a
/// remote reached over the network.
pub trait ChunkStore: ChunkReader + Send {
    /// has reports whether the store holds the chunk.
    fn has(&self, hash: &Hash) -> bool;

    /// has_many reports whether the store holds each chunk, in the order asked for, which a remote store answers
    /// together.
    fn has_many(&self, hashes: &[Hash]) -> Vec<bool> {
        hashes.iter().map(|hash| self.has(hash)).collect()
    }

    /// put adds a chunk with the addresses it refers to, which becomes durable when a later commit succeeds.
    fn put(&mut self, chunk: Chunk, refs: Vec<Hash>) -> Result<()>;

    /// commit moves the root from the last root to the current one, reporting false when another writer moved it
    /// first.
    fn commit(&mut self, current: Hash, last: Hash) -> Result<bool>;

    /// root returns the store's root.
    fn root(&self) -> Hash;

    /// journal returns the store as a local journaling store, which only local stores are.
    fn journal(&mut self) -> Option<&mut JournalStore> {
        None
    }

    /// plan returns how to read a chunk once the store is let go, so that the reading itself doesn't hold the store.
    fn plan(&self, hash: &Hash) -> Result<Plan> {
        Ok(Plan::ready(self.get(hash)?))
    }
}

/// Plan is how to read a chunk after letting go of the store that planned it: the chunk itself, when the store had
/// it at hand, or where to read it.
pub struct Plan(PlanKind);

/// PlanKind is what a plan holds.
enum PlanKind {
    Ready(Option<Chunk>),
    Journal(journal_writer::JournalChunk),
    Files(Arc<Vec<Arc<store::Source>>>, Hash),
}

impl Plan {
    /// ready returns a plan that already holds the chunk, or knows the store lacks it.
    fn ready(chunk: Option<Chunk>) -> Plan {
        Plan(PlanKind::Ready(chunk))
    }

    /// read reads the chunk the plan is for, if the store held it.
    pub fn read(self) -> Result<Option<Chunk>> {
        match self.0 {
            PlanKind::Ready(chunk) => Ok(chunk),
            PlanKind::Journal(unread) => unread.read().map(Some),
            PlanKind::Files(sources, hash) => {
                for source in sources.iter() {
                    if let Some(chunk) = source.get(&hash)? {
                        return Ok(Some(chunk));
                    }
                }
                Ok(None)
            }
        }
    }
}

impl ChunkReader for JournalStore {
    fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        JournalStore::get(self, hash)
    }
}

impl ChunkStore for JournalStore {
    fn has(&self, hash: &Hash) -> bool {
        JournalStore::has(self, hash)
    }

    fn put(&mut self, chunk: Chunk, refs: Vec<Hash>) -> Result<()> {
        JournalStore::put(self, chunk, refs)
    }

    fn commit(&mut self, current: Hash, last: Hash) -> Result<bool> {
        JournalStore::commit(self, current, last)
    }

    fn root(&self) -> Hash {
        JournalStore::root(self)
    }

    fn journal(&mut self) -> Option<&mut JournalStore> {
        Some(self)
    }

    fn plan(&self, hash: &Hash) -> Result<Plan> {
        JournalStore::plan(self, hash)
    }
}

impl ChunkReader for BlockStore {
    fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        BlockStore::get(self, hash)
    }
}

impl ChunkReader for GenerationalStore {
    fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        GenerationalStore::get(self, hash)
    }
}

/// dump renders the root and every chunk of a database's noms directory, one chunk per line as its generation,
/// address, length, and the address of its data, sorted by address and then generation.
pub fn dump(dir: &std::path::Path) -> Result<String> {
    let store = GenerationalStore::open(dir)?;
    let mut lines = Vec::new();
    for (generation, block_store) in [("new", &store.new_gen), ("old", &store.old_gen)] {
        block_store.for_each(&mut |chunk| {
            lines.push((chunk.hash.to_string(), generation, chunk.data.len(), Hash::of(&chunk.data).to_string()));
            Ok(())
        })?;
    }
    lines.sort();
    let mut text = format!("root {}\n", store.root());
    for (address, generation, len, data) in lines {
        text.push_str(&format!("{generation} {address} {len} {data}\n"));
    }
    Ok(text)
}
