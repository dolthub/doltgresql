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
    /// The working set's address when the transaction began, or as it last wrote it.
    pub(crate) working_set: Hash,
    /// The branch's head commit and its root value.
    pub head: Hash,
    pub head_root: Hash,
    pub root: Root,
    /// The working root as it was when the transaction began.
    original: Vec<u8>,
    /// The working root as it was when the transaction began, decoded.
    original_root: Root,
    pub staged: Root,
    original_staged: Vec<u8>,
    original_staged_root: Root,
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
    /// The schema of the session's temporary tables while a statement runs with them in the root, which writes of
    /// the root leave out.
    pub temp_schema: Option<String>,
    /// The sequences that nextval advanced in the running statement, which the statement writes to the root once it
    /// ends rather than at every call.
    pub pending_sequences: HashMap<Vec<u8>, objects::Sequence>,
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

/// read_working_set reads the working set at the address.
pub fn read_working_set(db: &Database, address: &Hash) -> Result<WorkingSetFields> {
    let data = read(db, address)?;
    let ws = WorkingSet::new(Message(&data))?;
    let meta = match (ws.name(), ws.email(), ws.description()) {
        (Ok(name), Ok(email), Ok(description)) => Some(Meta {
            name: name.to_vec(),
            email: email.to_vec(),
            description: description.to_vec(),
            timestamp_millis: ws.timestamp_millis()?,
            user_timestamp_millis: 0,
        }),
        _ => None,
    };
    Ok(WorkingSetFields {
        working_root: ws.working_root()?,
        staged_root: ws.staged_root()?,
        merge_state: ws.merge_state()?.map(|t| merge_state_fields(&MergeState(t))).transpose()?,
        rebase_state: ws.rebase_state()?.map(|t| rebase_state_fields(&RebaseState(t))).transpose()?,
        meta,
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
            original_root: root.clone(),
            root,
            original,
            original_staged_root: staged.clone(),
            staged,
            original_staged,
            original_merge: merge.clone(),
            merge,
            original_rebase: rebase.clone(),
            rebase,
            started: crate::datetime::clock(),
            detached,
            temp_schema: None,
            pending_sequences: HashMap::new(),
        })
    }

    /// changed reports whether the transaction changed the working root.
    pub fn changed(&self) -> bool {
        self.root.encode() != self.original
    }

    /// store_pending writes the sequences that the running statement advanced to the working root.
    pub fn store_pending(&mut self, db: &mut Database) -> Result<()> {
        for (_, sequence) in std::mem::take(&mut self.pending_sequences) {
            crate::sequences::store(db, &mut self.root, &sequence)?;
        }
        Ok(())
    }

    /// gc_roots returns what garbage collection must keep for the transaction: the addresses of its working set and
    /// head as they were when it began, and its working and staged roots when it has changed them.
    pub fn gc_roots(&self) -> (Vec<Hash>, Vec<Vec<u8>>) {
        let addresses = vec![self.working_set, self.head];
        let mut roots = Vec::new();
        if self.root != self.original_root {
            roots.push(self.root.encode());
        }
        if self.staged != self.original_staged_root {
            roots.push(self.staged.encode());
        }
        (addresses, roots)
    }

    /// working_set_fields writes the working and staged roots and returns the working set that holds them.
    fn working_set_fields(&self, db: &mut Database, user: &str, host: &str) -> Result<WorkingSetFields> {
        let persisted = self.persisted_root(db)?.encode();
        let working_root = db.write_value(persisted)?;
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

    /// persisted_root returns the working root without the session's temporary objects, as it is written.
    fn persisted_root(&self, db: &mut Database) -> Result<Root> {
        let mut root = self.root.clone();
        if let Some(schema) = &self.temp_schema {
            strip_schema(db, &mut root, schema)?;
        }
        Ok(root)
    }

    /// changed_persisted reports whether the transaction changed what it writes back to the working set.
    pub(crate) fn changed_persisted(&self, db: &mut Database) -> Result<bool> {
        Ok(self.persisted_root(db)?.encode() != self.original
            || self.staged.encode() != self.original_staged
            || self.merge != self.original_merge
            || self.rebase != self.original_rebase)
    }

    /// inject_temp puts the session's temporary objects in their schema of the root.
    pub fn inject_temp(&mut self, db: &mut Database, schema: &str, objects: &TempObjects) -> Result<()> {
        self.temp_schema = Some(schema.to_string());
        if !self.root.schemas.iter().any(|s| s == schema.as_bytes()) {
            self.root.schemas.push(schema.as_bytes().to_vec());
            self.root.schemas.sort();
        }
        for (name, address) in &objects.tables {
            self.root.put_table(db, schema, name, Some(*address))?;
        }
        for (collection, id, address) in &objects.objects {
            self.root.put_object(db, *collection, id, Some(*address))?;
        }
        if !objects.foreign_keys.is_empty() {
            let mut keys = crate::foreign::load(db, &self.root)?;
            keys.extend(objects.foreign_keys.iter().cloned());
            crate::foreign::store(db, &mut self.root, &keys)?;
        }
        Ok(())
    }

    /// take_temp removes the session's temporary objects and their schema from the root, returning them, or None when
    /// the root holds no temporary schema.
    pub fn take_temp(&mut self, db: &mut Database) -> Result<Option<TempObjects>> {
        let Some(schema) = self.temp_schema.take() else { return Ok(None) };
        if !self.root.schemas.iter().any(|s| s == schema.as_bytes()) {
            return Ok(None);
        }
        strip_schema(db, &mut self.root, &schema).map(Some)
    }

    /// flush writes the working and staged roots to the working set now, when they changed, and continues the
    /// transaction from there.
    pub fn flush(&mut self, db: &mut Database, user: &str, host: &str) -> Result<()> {
        self.store_pending(db)?;
        if !self.changed_persisted(db)? {
            return Ok(());
        }
        let fields = self.working_set_fields(db, user, host)?;
        self.working_set = match db.update_working_set(&working_set_ref(&self.branch), &fields, self.working_set) {
            Ok(address) => address,
            Err(database::Error::OptimisticLockFailed) => return Err(serialization_failure()),
            Err(err) => return Err(err.into()),
        };
        let persisted = self.persisted_root(db)?;
        self.original = persisted.encode();
        self.original_root = persisted;
        self.original_staged = self.staged.encode();
        self.original_staged_root = self.staged.clone();
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
        self.store_pending(db)?;
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
        let persisted = self.persisted_root(db)?;
        self.original = persisted.encode();
        self.original_root = persisted;
        self.original_staged = self.staged.encode();
        self.original_staged_root = self.staged.clone();
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

/// retry_transaction_error returns Dolt's error for a transaction that conflicts with one another client committed,
/// after a detail that may be empty.
pub fn retry_transaction_error(detail: &str) -> PgError {
    let detail = if detail.is_empty() { String::new() } else { format!("{detail}: ") };
    PgError::new(
        code::SERIALIZATION_FAILURE,
        format!(
            "serialization failure: {detail}this transaction conflicts with a committed transaction from another \
             client, try restarting transaction"
        ),
    )
}

/// working_roots returns the working and staged roots of the working set at the address.
pub fn working_roots(db: &Database, address: Hash) -> Result<(Root, Root)> {
    let data = read(db, &address)?;
    let ws = WorkingSet::new(Message(&data))?;
    let working = Root::decode(&read(db, &ws.working_root()?)?)?;
    let staged = match ws.staged_root()? {
        Some(staged) => Root::decode(&read(db, &staged)?)?,
        None => working.clone(),
    };
    Ok((working, staged))
}

/// serialization_failure returns the error for a transaction that lost a race with a concurrent one.
fn serialization_failure() -> PgError {
    PgError::new(code::SERIALIZATION_FAILURE, "could not serialize access due to concurrent update")
}

/// TempObjects are the tables, root objects, and foreign keys of a session's temporary schema.
#[derive(Clone, Default)]
pub struct TempObjects {
    pub tables: Vec<(String, Hash)>,
    /// The root objects, each with its collection, ID, and address.
    pub objects: Vec<(usize, Vec<u8>, Hash)>,
    pub foreign_keys: Vec<crate::foreign::ForeignKeyDef>,
}

/// strip_schema removes a schema and everything in it from a root and returns what it removed, leaving a root object
/// collection it empties unset.
fn strip_schema(db: &mut Database, root: &mut Root, schema: &str) -> Result<TempObjects> {
    let (foreign_keys, kept): (Vec<_>, Vec<_>) =
        crate::foreign::load(db, root)?.into_iter().partition(|key| key.child_schema == schema);
    let mut removed = TempObjects { foreign_keys, ..TempObjects::default() };
    if !removed.foreign_keys.is_empty() {
        crate::foreign::store(db, root, &kept)?;
    }
    for collection in 0..serial::write::ROOT_OBJECT_COLLECTIONS {
        for (id, address) in root.objects(db, collection)? {
            if crate::catalog::id::segments(&id).first().is_some_and(|s| s == schema) {
                root.put_object(db, collection, &id, None)?;
                removed.objects.push((collection, id, address));
            }
        }
        if removed.objects.iter().any(|(c, ..)| *c == collection) && root.objects(db, collection)?.is_empty() {
            root.root_objects[collection] = None;
        }
    }
    let prefix = doltdb::root::table_key(schema, "");
    for (key, address) in root.tables(db)? {
        if let Some(name) = key.strip_prefix(prefix.as_slice()) {
            let name = String::from_utf8_lossy(name).into_owned();
            root.put_table(db, schema, &name, None)?;
            removed.tables.push((name, address));
        }
    }
    root.schemas.retain(|s| s != schema.as_bytes());
    Ok(removed)
}
