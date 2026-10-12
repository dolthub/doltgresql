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

//! Advisory locks: locks on application-defined keys that sessions take for themselves or for their transactions,
//! exclusively or shared, as Postgres' pg_advisory_lock family takes them.

use std::collections::HashMap;
use std::sync::{Condvar, Mutex};

/// Key is what an advisory lock locks: a database, and a bigint key or a pair of integer keys, which Postgres keeps
/// apart.
#[derive(Clone, Debug, PartialEq, Eq, Hash)]
pub struct Key {
    pub database: String,
    pub key: i64,
    /// Whether the key is a pair of integers rather than one bigint.
    pub pair: bool,
}

impl Key {
    /// new returns the key of an advisory lock call's arguments in a database: one bigint, or a pair of integers.
    pub fn new(database: &str, args: &[i64]) -> Key {
        let database = database.to_string();
        match args {
            [high, low, ..] => Key { database, key: (high << 32) | (low & 0xffff_ffff), pair: true },
            [single] => Key { database, key: *single, pair: false },
            [] => Key { database, key: 0, pair: false },
        }
    }
}

/// Holds counts how many times a session holds a lock in each mode and scope.
#[derive(Clone, Copy, Debug, Default)]
struct Holds {
    exclusive_session: u32,
    exclusive_transaction: u32,
    shared_session: u32,
    shared_transaction: u32,
}

impl Holds {
    /// exclusive reports whether the session holds the lock exclusively.
    fn exclusive(&self) -> bool {
        self.exclusive_session + self.exclusive_transaction > 0
    }

    /// any reports whether the session holds the lock at all.
    fn any(&self) -> bool {
        self.exclusive() || self.shared_session + self.shared_transaction > 0
    }

    /// count returns the count of a mode and scope.
    fn count(&mut self, exclusive: bool, transaction: bool) -> &mut u32 {
        match (exclusive, transaction) {
            (true, false) => &mut self.exclusive_session,
            (true, true) => &mut self.exclusive_transaction,
            (false, false) => &mut self.shared_session,
            (false, true) => &mut self.shared_transaction,
        }
    }
}

/// AdvisoryLocks are the advisory locks that the sessions of an engine hold.
#[derive(Debug, Default)]
pub struct AdvisoryLocks {
    held: Mutex<HashMap<Key, HashMap<u64, Holds>>>,
    released: Condvar,
}

impl AdvisoryLocks {
    /// blocked reports whether another session holds a lock in a mode that conflicts with the one a session wants.
    fn blocked(holders: Option<&HashMap<u64, Holds>>, session: u64, exclusive: bool) -> bool {
        holders.is_some_and(|holders| {
            holders.iter().any(|(&other, holds)| other != session && (holds.exclusive() || (exclusive && holds.any())))
        })
    }

    /// wait waits until no other session holds a lock in a mode that conflicts with the one a session wants.
    pub fn wait(&self, session: u64, key: &Key, exclusive: bool) {
        let mut held = self.held.lock().unwrap_or_else(|e| e.into_inner());
        while Self::blocked(held.get(key), session, exclusive) {
            held = self.released.wait(held).unwrap_or_else(|e| e.into_inner());
        }
    }

    /// acquire takes a lock for a session, exclusively or shared, for the session or for its transaction, and
    /// reports whether it took the lock, which it does not when another session holds it in a conflicting mode.
    pub fn acquire(&self, session: u64, key: Key, exclusive: bool, transaction: bool) -> bool {
        let mut held = self.held.lock().unwrap_or_else(|e| e.into_inner());
        if Self::blocked(held.get(&key), session, exclusive) {
            return false;
        }
        *held.entry(key).or_default().entry(session).or_default().count(exclusive, transaction) += 1;
        true
    }

    /// release releases one session-level hold of a lock, reporting whether the session held it.
    pub fn release(&self, session: u64, key: &Key, exclusive: bool) -> bool {
        let mut held = self.held.lock().unwrap_or_else(|e| e.into_inner());
        let Some(holds) = held.get_mut(key).and_then(|holders| holders.get_mut(&session)) else { return false };
        let count = holds.count(exclusive, false);
        if *count == 0 {
            return false;
        }
        *count -= 1;
        Self::tidy(&mut held);
        self.released.notify_all();
        true
    }

    /// release_all releases a session's holds, only those for its transaction when `transaction` is set, and
    /// otherwise only its session-level holds unless `everything` is set too.
    pub fn release_all(&self, session: u64, transaction: bool, everything: bool) {
        let mut held = self.held.lock().unwrap_or_else(|e| e.into_inner());
        for holders in held.values_mut() {
            if let Some(holds) = holders.get_mut(&session) {
                if transaction || everything {
                    holds.exclusive_transaction = 0;
                    holds.shared_transaction = 0;
                }
                if !transaction {
                    holds.exclusive_session = 0;
                    holds.shared_session = 0;
                }
            }
        }
        Self::tidy(&mut held);
        self.released.notify_all();
    }

    /// tidy forgets the locks that nobody holds.
    fn tidy(held: &mut HashMap<Key, HashMap<u64, Holds>>) {
        for holders in held.values_mut() {
            holders.retain(|_, holds| holds.any());
        }
        held.retain(|_, holders| !holders.is_empty());
    }
}
