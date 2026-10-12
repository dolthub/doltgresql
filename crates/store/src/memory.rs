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

use std::collections::HashMap;

use crate::{Chunk, ChunkReader, ChunkStore, Hash, Result};

/// MemoryStore is a chunk store that keeps its chunks in memory and never writes them anywhere, for a database that
/// lasts only as long as the server, such as Dolt's dolt_cluster database.
#[derive(Default)]
pub struct MemoryStore {
    chunks: HashMap<Hash, Chunk>,
    root: Hash,
}

impl ChunkReader for MemoryStore {
    fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        Ok(self.chunks.get(hash).cloned())
    }
}

impl ChunkStore for MemoryStore {
    fn has(&self, hash: &Hash) -> bool {
        self.chunks.contains_key(hash)
    }

    fn put(&mut self, chunk: Chunk, _refs: Vec<Hash>) -> Result<()> {
        self.chunks.insert(chunk.hash, chunk);
        Ok(())
    }

    fn commit(&mut self, current: Hash, last: Hash) -> Result<bool> {
        if self.root != last {
            return Ok(false);
        }
        self.root = current;
        Ok(true)
    }

    fn root(&self) -> Hash {
        self.root
    }
}
