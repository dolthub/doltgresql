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

use std::collections::{HashMap, HashSet};
use std::sync::{Arc, Mutex};
use std::time::{SystemTime, UNIX_EPOCH};

use doltdb::create::{branch_ref, working_set_ref};
use doltdb::database::{self, CommitMeta, Database, PendingCommit};
use doltdb::root::Root;
use serial::write::{MergeStateFields, Meta, RebaseStateFields, WorkingSetFields, write_working_set};
use serial::{Commit, MergeState, Message, RebaseState, TableSchema, WorkingSet};
use store::Hash;

use crate::catalog::table::TableDef;
use crate::error::{PgError, Result, code};

/// DbHandle is an open database that sessions share, one statement at a time.
pub type DbHandle = Arc<Mutex<Database>>;

/// SequenceTracker holds the latest state of each of a database's sequences across every branch and transaction,
/// since sequence values are never handed out twice.
pub type SequenceTracker = Arc<Mutex<HashMap<Vec<u8>, objects::Sequence>>>;

/// Txn is an open transaction on a branch of a database.
#[derive(Clone)]
pub struct Txn {
    pub database: String,
    pub branch: String,
    pub handle: DbHandle,
    pub sequences: SequenceTracker,
    /// The working set's address when the transaction began.
    working_set: Hash,
    /// The branch's head commit and its root value.
    pub head: Hash,
    pub head_root: Hash,
    pub root: Root,
    /// The working root as it was when the transaction began.
    original: Vec<u8>,
    pub staged: Root,
    original_staged: Vec<u8>,
    /// The merge in progress, if any.
    pub merge: Option<MergeStateFields>,
    original_merge: Option<MergeStateFields>,
    /// The rebase in progress, if any.
    pub rebase: Option<RebaseStateFields>,
    original_rebase: Option<RebaseStateFields>,
    /// When the transaction began, as a UTC timestamp.
    pub started: i64,
    /// Whether the transaction reads a revision that is not a branch, such as a tag, which it cannot change.
    pub detached: bool,
}

/// read returns the message at the address, failing when the database lacks it.
pub fn read(db: &Database, address: &Hash) -> Result<Vec<u8>> {
    db.read_value(address)?.ok_or_else(|| PgError::internal(format!("missing chunk {address}")))
}

/// rebase_state_fields reads a working set's rebase in progress.
fn rebase_state_fields(state: &RebaseState<'_>) -> Result<RebaseStateFields> {
    let address = |bytes: &[u8]| serial::hash(bytes).map_err(PgError::from);
    Ok(RebaseStateFields {
        pre_working_root: address(state.pre_working_root()?)?,
        onto_commit: address(state.onto_commit()?)?,
        branch: state.branch()?.to_vec(),
        commit_becomes_empty_handling: state.commit_becomes_empty_handling()?,
        empty_commit_handling: state.empty_commit_handling()?,
        last_attempted_step: state.last_attempted_step()?,
        rebasing_started: state.rebasing_started()?,
        skip_verification: state.skip_verification()?,
    })
}

/// merge_state_fields reads a working set's merge in progress.
fn merge_state_fields(state: &MergeState<'_>) -> Result<MergeStateFields> {
    let address = |bytes: &[u8]| serial::hash(bytes).map_err(PgError::from);
    let head = state.pre_merge_head_commit()?;
    Ok(MergeStateFields {
        pre_working_root: address(state.pre_working_root()?)?,
        from_commit: address(state.from_commit()?)?,
        from_commit_spec: state.from_commit_spec()?.to_vec(),
        unmergable_tables: state.unmergable_tables()?.into_iter().map(<[u8]>::to_vec).collect(),
        is_cherry_pick: state.is_cherry_pick()?,
        is_revert: state.is_revert()?,
        pre_merge_head_commit: if head.is_empty() { None } else { Some(address(head)?) },
        pending_commit_hashes: state.pending_commit_hashes()?.into_iter().map(<[u8]>::to_vec).collect(),
    })
}

impl Txn {
    /// begin starts a transaction on the branch, reading its working set.
    pub fn begin(handle: DbHandle, sequences: SequenceTracker, database: &str, branch: &str) -> Result<Txn> {
        let db = handle.clone();
        let mut db = db.lock().map_err(|_| PgError::internal("a database lock was poisoned"))?;
        Txn::begin_locked(&mut db, handle, sequences, database, branch)
    }

    /// begin_locked starts a transaction on the branch of a database whose lock the caller holds, or a detached one on
    /// the commit of another revision, as Dolt's revision databases are.
    pub fn begin_locked(
        db: &mut Database,
        handle: DbHandle,
        sequences: SequenceTracker,
        database: &str,
        branch: &str,
    ) -> Result<Txn> {
        let not_found =
            || PgError::new(code::INVALID_CATALOG_NAME, format!("database \"{database}/{branch}\" does not exist"));
        let (head, detached) = match db.head(&branch_ref(branch))? {
            Some(head) => (head, false),
            None => {
                let spec = crate::dolt::history::tag_spelling(db, branch)?;
                (crate::dolt::history::resolve(db, Hash::default(), &spec).map_err(|_| not_found())?, true)
            }
        };
        let commit = read(db, &head)?;
        let head_root = Commit::new(Message(&commit))?.root()?;
        let working_set = if detached { None } else { db.head(&working_set_ref(branch))? };
        let (working_set, working, staged, merge, rebase) = match working_set {
            Some(address) => {
                let data = read(db, &address)?;
                let ws = WorkingSet::new(Message(&data))?;
                let working = ws.working_root()?;
                let merge = ws.merge_state()?.map(|t| merge_state_fields(&MergeState(t))).transpose()?;
                let rebase = ws.rebase_state()?.map(|t| rebase_state_fields(&RebaseState(t))).transpose()?;
                (address, working, ws.staged_root()?.unwrap_or(working), merge, rebase)
            }
            None => (Hash::default(), head_root, head_root, None, None),
        };
        let original = read(db, &working)?;
        let original_staged = read(db, &staged)?;
        let root = Root::decode(&original)?;
        let staged = Root::decode(&original_staged)?;
        Ok(Txn {
            database: database.to_string(),
            branch: branch.to_string(),
            handle,
            sequences,
            working_set,
            head,
            head_root,
            root,
            original,
            staged,
            original_staged,
            original_merge: merge.clone(),
            merge,
            original_rebase: rebase.clone(),
            rebase,
            started: crate::datetime::clock(),
            detached,
        })
    }

    /// changed reports whether the transaction changed the working root.
    pub fn changed(&self) -> bool {
        self.root.encode() != self.original
    }

    /// working_set_fields writes the working and staged roots and returns the working set that holds them.
    fn working_set_fields(&self, db: &mut Database, user: &str, host: &str) -> Result<WorkingSetFields> {
        let working_root = db.write_value(self.root.encode())?;
        let staged_root = db.write_value(self.staged.encode())?;
        let seconds = SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_secs());
        Ok(WorkingSetFields {
            working_root,
            staged_root: Some(staged_root),
            merge_state: self.merge.clone(),
            rebase_state: self.rebase.clone(),
            meta: Some(Meta {
                name: user.as_bytes().to_vec(),
                email: format!("{user}@{host}").into_bytes(),
                description: b"sql transaction".to_vec(),
                timestamp_millis: seconds,
                user_timestamp_millis: 0,
            }),
        })
    }

    /// commit writes the working and staged roots back to the working set when the transaction changed them, as the
    /// user connected from the host.
    pub fn commit(self, db: &mut Database, user: &str, host: &str) -> Result<()> {
        if self.detached
            || self.root.encode() == self.original
                && self.staged.encode() == self.original_staged
                && self.merge == self.original_merge
                && self.rebase == self.original_rebase
        {
            return Ok(());
        }
        let fields = self.working_set_fields(db, user, host)?;
        match db.update_working_set(&working_set_ref(&self.branch), &fields, self.working_set) {
            Ok(_) => Ok(()),
            Err(database::Error::OptimisticLockFailed) => Err(serialization_failure()),
            Err(err) => Err(err.into()),
        }
    }

    /// flush writes the working and staged roots to the working set now, when they changed, and continues the
    /// transaction from there.
    pub fn flush(&mut self, db: &mut Database, user: &str, host: &str) -> Result<()> {
        if self.root.encode() == self.original
            && self.staged.encode() == self.original_staged
            && self.merge == self.original_merge
            && self.rebase == self.original_rebase
        {
            return Ok(());
        }
        let fields = self.working_set_fields(db, user, host)?;
        self.working_set = match db.update_working_set(&working_set_ref(&self.branch), &fields, self.working_set) {
            Ok(address) => address,
            Err(database::Error::OptimisticLockFailed) => return Err(serialization_failure()),
            Err(err) => return Err(err.into()),
        };
        self.original = self.root.encode();
        self.original_staged = self.staged.encode();
        self.original_merge = self.merge.clone();
        self.original_rebase = self.rebase.clone();
        Ok(())
    }

    /// dolt_commit commits the staged root on the branch's head with any extra parents, writes the working set
    /// alongside it, and continues the transaction from the new commit.
    pub fn dolt_commit(
        &mut self,
        db: &mut Database,
        user: &str,
        host: &str,
        parents: Vec<Hash>,
        meta: CommitMeta,
    ) -> Result<Hash> {
        self.merge = None;
        let fields = self.working_set_fields(db, user, host)?;
        let pending = PendingCommit { root_value: self.staged.encode(), parents, meta };
        let commit = match db.commit_with_working_set(
            &branch_ref(&self.branch),
            &working_set_ref(&self.branch),
            &fields,
            self.working_set,
            pending,
        ) {
            Ok(commit) => commit,
            Err(database::Error::OptimisticLockFailed) => return Err(serialization_failure()),
            Err(err) => return Err(err.into()),
        };
        self.working_set = Hash::of(&write_working_set(&fields));
        self.head = commit.hash;
        self.head_root = fields.staged_root.unwrap_or_default();
        self.original = self.root.encode();
        self.original_staged = self.staged.encode();
        self.original_merge = None;
        Ok(commit.hash)
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
        let roots = vec![self.root.clone(), Root::decode(&read(db, &self.head_root)?)?];
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

/// serialization_failure returns the error for a transaction that lost a race with a concurrent one.
fn serialization_failure() -> PgError {
    PgError::new(code::SERIALIZATION_FAILURE, "could not serialize access due to concurrent update")
}
