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

use std::collections::{HashSet, VecDeque};

use sha2::{Digest, Sha256};

use crate::pgx::StatementDescription;
use crate::pgx::auth::hex;

/// statement_name returns the name pgx gives a cached statement.
pub(crate) fn statement_name(sql: &str) -> String {
    let digest = Sha256::digest(sql.as_bytes());
    format!("stmtcache_{}", hex(&digest[..24]))
}

/// LruCache mirrors pgx's statement cache: a least-recently-used cache whose evicted and invalidated statements are
/// remembered until they are deallocated.
pub(crate) struct LruCache {
    capacity: usize,
    entries: VecDeque<StatementDescription>,
    invalidated: Vec<StatementDescription>,
    invalidated_sql: HashSet<String>,
}

impl LruCache {
    /// new returns an empty cache with the given capacity.
    pub(crate) fn new(capacity: usize) -> LruCache {
        LruCache { capacity, entries: VecDeque::new(), invalidated: Vec::new(), invalidated_sql: HashSet::new() }
    }

    /// get returns the cached statement for the SQL, marking it most recently used.
    pub(crate) fn get(&mut self, sql: &str) -> Option<StatementDescription> {
        let index = self.entries.iter().position(|sd| sd.sql == sql)?;
        let sd = self.entries.remove(index)?;
        self.entries.push_front(sd.clone());
        Some(sd)
    }

    /// put caches a statement unless it is already cached or awaiting deallocation, evicting the least recently used
    /// statement when the cache is full.
    pub(crate) fn put(&mut self, sd: StatementDescription) {
        if self.entries.iter().any(|entry| entry.sql == sd.sql) || self.invalidated_sql.contains(&sd.sql) {
            return;
        }
        if self.entries.len() == self.capacity
            && let Some(oldest) = self.entries.pop_back()
        {
            self.invalidated_sql.insert(oldest.sql.clone());
            self.invalidated.push(oldest);
        }
        self.entries.push_front(sd);
    }

    /// invalidate removes the statement for the SQL, remembering it for deallocation.
    pub(crate) fn invalidate(&mut self, sql: &str) {
        if let Some(index) = self.entries.iter().position(|sd| sd.sql == sql) {
            let sd = self.entries.remove(index).unwrap();
            self.invalidated_sql.insert(sd.sql.clone());
            self.invalidated.push(sd);
        }
    }

    /// invalidated returns the statements awaiting deallocation, in the order they were removed.
    pub(crate) fn invalidated(&self) -> &[StatementDescription] {
        &self.invalidated
    }

    /// remove_invalidated forgets the statements awaiting deallocation.
    pub(crate) fn remove_invalidated(&mut self) {
        self.invalidated.clear();
        self.invalidated_sql.clear();
    }
}
