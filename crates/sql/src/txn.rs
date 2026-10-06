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

//! Transactions: a session's view of a branch's working root, written back to the branch's working set on commit.

use std::collections::HashSet;
use std::sync::{Arc, Mutex};
use std::time::{SystemTime, UNIX_EPOCH};

use doltdb::create::{branch_ref, working_set_ref};
use doltdb::database::{self, Database};
use doltdb::root::Root;
use serial::write::{Meta, WorkingSetFields};
use serial::{Commit, Message, TableSchema, WorkingSet};
use store::Hash;

use crate::catalog::table::TableDef;
use crate::error::{PgError, Result, code};

/// DbHandle is an open database that sessions share, one statement at a time.
pub type DbHandle = Arc<Mutex<Database>>;

/// Txn is an open transaction on a branch of a database.
pub struct Txn {
    pub database: String,
    pub branch: String,
    pub handle: DbHandle,
    /// The working set's address when the transaction began.
    working_set: Hash,
    staged: Hash,
    /// The branch head's root value, which column tags must also avoid.
    head_root: Option<Hash>,
    pub root: Root,
    /// The working root as it was when the transaction began.
    original: Vec<u8>,
    /// When the transaction began, as a UTC timestamp.
    pub started: i64,
}

/// read returns the message at the address, failing when the database lacks it.
pub fn read(db: &Database, address: &Hash) -> Result<Vec<u8>> {
    db.read_value(address)?.ok_or_else(|| PgError::internal(format!("missing chunk {address}")))
}

impl Txn {
    /// begin starts a transaction on the branch, reading its working set.
    pub fn begin(handle: DbHandle, database: &str, branch: &str) -> Result<Txn> {
        let mut db = handle.lock().map_err(|_| PgError::internal("a database lock was poisoned"))?;
        let not_found = || PgError::new(code::INVALID_CATALOG_NAME, format!("database not found: {database}/{branch}"));
        let head = db.head(&branch_ref(branch))?.ok_or_else(not_found)?;
        let commit = read(&db, &head)?;
        let head_root = Commit::new(Message(&commit))?.root()?;
        let (working_set, working, staged) = match db.head(&working_set_ref(branch))? {
            Some(address) => {
                let data = read(&db, &address)?;
                let ws = WorkingSet::new(Message(&data))?;
                let working = ws.working_root()?;
                (address, working, ws.staged_root()?.unwrap_or(working))
            }
            None => (Hash::default(), head_root, head_root),
        };
        let original = read(&db, &working)?;
        let root = Root::decode(&original)?;
        drop(db);
        Ok(Txn {
            database: database.to_string(),
            branch: branch.to_string(),
            handle,
            working_set,
            staged,
            head_root: Some(head_root),
            root,
            original,
            started: crate::datetime::clock(),
        })
    }

    /// changed reports whether the transaction changed the working root.
    pub fn changed(&self) -> bool {
        self.root.encode() != self.original
    }

    /// commit writes the working root back to the working set when the transaction changed it, as the user connected
    /// from the host.
    pub fn commit(self, db: &mut Database, user: &str, host: &str) -> Result<()> {
        let encoded = self.root.encode();
        if encoded == self.original {
            return Ok(());
        }
        let working_root = db.write_value(encoded)?;
        let seconds = SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_secs());
        let fields = WorkingSetFields {
            working_root,
            staged_root: Some(self.staged),
            merge_state: None,
            rebase_state: None,
            meta: Some(Meta {
                name: user.as_bytes().to_vec(),
                email: format!("{user}@{host}").into_bytes(),
                description: b"sql transaction".to_vec(),
                timestamp_millis: seconds,
                user_timestamp_millis: 0,
            }),
        };
        match db.update_working_set(&working_set_ref(&self.branch), &fields, self.working_set) {
            Ok(_) => Ok(()),
            Err(database::Error::OptimisticLockFailed) => {
                Err(PgError::new(code::SERIALIZATION_FAILURE, "could not serialize access due to concurrent update"))
            }
            Err(err) => Err(err.into()),
        }
    }

    /// table loads a table of the working root, if it has the table.
    pub fn table(&self, db: &mut Database, schema: &str, name: &str) -> Result<Option<TableDef>> {
        match self.root.table(db, schema, name)? {
            Some(address) => Ok(Some(TableDef::load(db, schema, name, address)?)),
            None => Ok(None),
        }
    }

    /// all_tags returns the column tags of every table in the working root and the branch head, which new columns
    /// must avoid.
    pub fn all_tags(&self, db: &mut Database) -> Result<HashSet<u64>> {
        let mut tags = HashSet::new();
        let mut roots = vec![self.root.clone()];
        if let Some(head) = self.head_root {
            roots.push(Root::decode(&read(db, &head)?)?);
        }
        for root in roots {
            for (_, address) in root.tables(db)? {
                let table = doltdb::table::Table::decode(&read(db, &address)?)?;
                let schema = read(db, &table.schema)?;
                for column in TableSchema::new(Message(&schema))?.columns()? {
                    tags.insert(column.tag);
                }
            }
        }
        Ok(tags)
    }
}
