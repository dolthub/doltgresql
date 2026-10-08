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

//! An open database that sessions and servers share, with the locks that order their work on it.

use std::ops::{Deref, DerefMut};
use std::sync::{Mutex, MutexGuard, RwLock, RwLockReadGuard, RwLockWriteGuard};

use crate::database::Database;

/// Handle is an open database that sessions share. Statements hold its gate shared while they run, so that garbage
/// collection, which holds it alone, sees none running, and statements that change its datasets also hold its
/// writer lock, one at a time.
pub struct Handle {
    database: Database,
    gate: RwLock<()>,
    writer: Mutex<()>,
}

impl Handle {
    /// new returns a handle on an open database.
    pub fn new(database: Database) -> Handle {
        Handle { database, gate: RwLock::new(()), writer: Mutex::new(()) }
    }

    /// read returns the database for work that only changes chunks and the session's own transaction, alongside
    /// any other work but garbage collection.
    pub fn read(&self) -> Guard<'_> {
        let shared = self.gate.read().unwrap_or_else(|poisoned| poisoned.into_inner());
        Guard { database: self.database.clone(), _shared: Some(shared), _exclusive: None, _writer: None }
    }

    /// write returns the database for work that may change its datasets, alongside other readers but after other
    /// writers.
    pub fn write(&self) -> Guard<'_> {
        let shared = self.gate.read().unwrap_or_else(|poisoned| poisoned.into_inner());
        let writer = self.writer.lock().unwrap_or_else(|poisoned| poisoned.into_inner());
        Guard { database: self.database.clone(), _shared: Some(shared), _exclusive: None, _writer: Some(writer) }
    }

    /// exclusive returns the database once no other work is running on it, holding off new work until it is dropped.
    pub fn exclusive(&self) -> Guard<'_> {
        let exclusive = self.gate.write().unwrap_or_else(|poisoned| poisoned.into_inner());
        Guard { database: self.database.clone(), _shared: None, _exclusive: Some(exclusive), _writer: None }
    }
}

/// Guard is a database that a handle lent out with its locks, which it holds until it is dropped.
pub struct Guard<'a> {
    database: Database,
    _shared: Option<RwLockReadGuard<'a, ()>>,
    _exclusive: Option<RwLockWriteGuard<'a, ()>>,
    _writer: Option<MutexGuard<'a, ()>>,
}

impl Deref for Guard<'_> {
    type Target = Database;

    fn deref(&self) -> &Database {
        &self.database
    }
}

impl DerefMut for Guard<'_> {
    fn deref_mut(&mut self) -> &mut Database {
        &mut self.database
    }
}
