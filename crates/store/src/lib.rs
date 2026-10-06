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

mod archive;
mod chunk;
mod error;
mod file;
mod hash;
mod journal;
mod manifest;
mod store;
mod table;

pub use chunk::Chunk;
pub use error::{Error, Result};
pub use hash::Hash;
pub use journal::{JOURNAL_FILE, JournalRecord, read_records};
pub use manifest::{Manifest, TableSpec};
pub use store::{BlockStore, GenerationalStore};
pub use table::{TableReader, TableWriter};

/// ChunkReader reads chunks by address.
pub trait ChunkReader {
    /// get returns the chunk when the reader holds it.
    fn get(&self, hash: &Hash) -> Result<Option<Chunk>>;

    /// require returns the chunk, failing when the reader lacks it.
    fn require(&self, hash: &Hash) -> Result<Chunk> {
        self.get(hash)?.ok_or_else(|| Error::Corrupt(format!("chunk {hash} is missing")))
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
