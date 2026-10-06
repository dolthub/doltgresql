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

//! Creating a database as Doltgres' CREATE DATABASE does: Dolt's repository initialization, then the default schemas
//! committed by the session's user, written value by value in Go's order.

use std::path::Path;

use serial::write::{Meta, ROOT_OBJECT_COLLECTIONS, RootValueFields, WorkingSetFields, write_root_value};
use store::Hash;

use crate::database::{CommitMeta, Database, PendingCommit, Result};

/// FEATURE_VERSION is Doltgres' root value feature version, which is Dolt's.
pub const FEATURE_VERSION: i64 = 7;

/// DEFAULT_COLLATION is the collation of a new root value, utf8mb4_0900_bin.
pub const DEFAULT_COLLATION: u16 = 309;

/// DEFAULT_SCHEMAS are the schemas every Doltgres database starts with.
pub const DEFAULT_SCHEMAS: [&str; 3] = ["public", "pg_catalog", "dolt"];

/// SYSTEM_NAME is the author of a repository's initial commit.
pub const SYSTEM_NAME: &str = "Dolt System Account";

/// SYSTEM_EMAIL is the email of a repository's initial commit.
pub const SYSTEM_EMAIL: &str = "doltuser@dolthub.com";

/// CREATION_REF is the dataset of a repository's initial commit.
const CREATION_REF: &str = "refs/internal/create";

/// CreateTimes are the clock readings that creating a database records.
#[derive(Clone, Copy, Debug)]
pub struct CreateTimes {
    /// When initialization started, the initial commit's author time in Unix milliseconds.
    pub init_author_millis: i64,
    /// When the initial commit was built, its committer time in Unix milliseconds.
    pub init_committer_millis: u64,
    /// The Unix seconds of the environment's working set updates.
    pub environment_seconds: u64,
    /// The Unix seconds of the session's working set.
    pub session_seconds: u64,
    /// When the CREATE DATABASE commit was built, in Unix milliseconds.
    pub commit_millis: u64,
}

/// empty_root_value returns a Doltgres root value without tables or root objects, with the schemas sorted by name.
pub fn empty_root_value(schemas: &[&str]) -> Vec<u8> {
    let mut schemas: Vec<&str> = schemas.to_vec();
    schemas.sort();
    let tables = prolly::serialize_address_map(&[], &[], &[], 0);
    write_root_value(&RootValueFields {
        feature_version: FEATURE_VERSION,
        collation: DEFAULT_COLLATION,
        tables: &tables,
        schemas: schemas.iter().map(|s| s.as_bytes()).collect(),
        foreign_keys: &[0; Hash::LEN],
        root_objects: [None; ROOT_OBJECT_COLLECTIONS],
        added_root_object: None,
    })
}

/// branch_ref names a branch's dataset.
pub fn branch_ref(branch: &str) -> String {
    format!("refs/heads/{branch}")
}

/// working_set_ref names a branch's working set dataset.
pub fn working_set_ref(branch: &str) -> String {
    format!("workingSets/heads/{branch}")
}

/// environment_meta returns the working set author of Dolt's environment updates, which has no name or email.
pub fn environment_meta(seconds: u64) -> Meta {
    Meta {
        name: Vec::new(),
        email: Vec::new(),
        description: b"updated from dolt environment".to_vec(),
        timestamp_millis: seconds,
        user_timestamp_millis: 0,
    }
}

/// update_roots sets the working set's working root to the root, keeping the earlier staged root, and then its staged
/// root too, as Dolt's environment does with UpdateWorkingRoot and UpdateStagedRoot, returning the working set's
/// address.
fn update_roots(
    db: &mut Database,
    branch: &str,
    root: &[u8],
    earlier_staged: &[u8],
    previous: Hash,
    seconds: u64,
) -> Result<Hash> {
    let mut previous = previous;
    for staged_root in [earlier_staged, root] {
        let working = db.write_value(root.to_vec())?;
        let staged = db.write_value(staged_root.to_vec())?;
        let fields = WorkingSetFields {
            working_root: working,
            staged_root: Some(staged),
            merge_state: None,
            rebase_state: None,
            meta: Some(environment_meta(seconds)),
        };
        previous = db.update_working_set(&working_set_ref(branch), &fields, previous)?;
    }
    Ok(previous)
}

/// create_files creates the directories and files of an empty database in the directory, whose checked out branch
/// is the branch, as Dolt's environment does before it writes any data.
pub fn create_files(dir: &Path, branch: &str) -> Result<()> {
    let dolt = dir.join(".dolt");
    let noms = dolt.join("noms");
    std::fs::create_dir_all(noms.join("oldgen")).map_err(store::Error::from)?;
    std::fs::create_dir_all(dolt.join("temptf")).map_err(store::Error::from)?;
    std::fs::write(dolt.join("config.json"), "{}").map_err(store::Error::from)?;
    std::fs::write(
        dolt.join("repo_state.json"),
        format!("{{\n  \"head\": \"refs/heads/{branch}\",\n  \"remotes\": {{}},\n  \"backups\": {{}},\n  \"branches\": {{}}\n}}"),
    )
    .map_err(store::Error::from)?;
    std::fs::File::create(noms.join("oldgen/LOCK")).map_err(store::Error::from)?;
    Ok(())
}

/// create_database creates a database in the directory, whose initial commit is by Dolt's system account and whose
/// default schemas are committed by the user connected from the host, as Go's CREATE DATABASE does.
pub fn create_database(dir: &Path, branch: &str, user: &str, host: &str, times: &CreateTimes) -> Result<()> {
    create_files(dir, branch)?;
    let mut db = Database::open(&dir.join(".dolt/noms"))?;
    // Dolt's WriteEmptyRepo: the empty root, its commit on the creation ref, and the branch.
    let empty = empty_root_value(&[]);
    db.write_value(empty.clone())?;
    let init = CommitMeta {
        name: SYSTEM_NAME.into(),
        email: SYSTEM_EMAIL.into(),
        description: "Initialize data repository".into(),
        author_millis: times.init_author_millis,
        committer_millis: times.init_committer_millis,
        signature: String::new(),
        committer_name: None,
        committer_email: None,
    };
    let first = db.build_commit(None, empty.clone(), Vec::new(), &init)?;
    db.write_commit(CREATION_REF, None, &first)?;
    db.set_head(&branch_ref(branch), first.hash)?;
    // InitializeRepoState, then the default schemas through the environment's working and staged roots.
    let ws = update_roots(&mut db, branch, &empty, &empty, Hash::default(), times.environment_seconds)?;
    let schemas = empty_root_value(&DEFAULT_SCHEMAS);
    let ws = update_roots(&mut db, branch, &schemas, &empty, ws, times.environment_seconds)?;
    // The session's DoltCommit of the new schemas.
    let session = WorkingSetFields {
        working_root: Hash::of(&schemas),
        staged_root: Some(Hash::of(&schemas)),
        merge_state: None,
        rebase_state: None,
        meta: Some(Meta {
            name: user.as_bytes().to_vec(),
            email: format!("{user}@{host}").into_bytes(),
            description: b"sql transaction".to_vec(),
            timestamp_millis: times.session_seconds,
            user_timestamp_millis: 0,
        }),
    };
    let commit = CommitMeta {
        name: user.into(),
        email: format!("{user}@{host}"),
        description: "CREATE DATABASE".into(),
        author_millis: times.commit_millis as i64,
        committer_millis: times.commit_millis,
        signature: String::new(),
        committer_name: None,
        committer_email: None,
    };
    let pending = PendingCommit { root_value: schemas, parents: vec![first.hash], meta: commit };
    db.commit_with_working_set(&branch_ref(branch), &working_set_ref(branch), &session, ws, pending)?;
    db.close()
}
