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

//! Garbage collection in phases, with writes between its copy and its finish.

use std::path::Path;

use doltdb::create::{CreateTimes, create_database};
use doltdb::database::{Database, GcConfig, GcMode};
use store::Hash;

/// map writes an address map of one entry and returns its address.
fn map(db: &mut Database, name: &str, address: Hash) -> Hash {
    let bytes = db.address_map(&[(name.to_string(), address)]).unwrap();
    db.write_value(bytes).unwrap()
}

/// persist writes out the chunks put so far without moving the store root.
fn persist(db: &mut Database) {
    let root = db.root();
    assert!(db.commit_root(root, root).unwrap());
}

#[test]
fn writes_during_a_collection_survive_it() {
    for (mode, archive) in [(GcMode::Default, true), (GcMode::Full, false), (GcMode::Shallow, false)] {
        let dir = Path::new(env!("CARGO_TARGET_TMPDIR")).join(format!("gc_phases/{mode:?}/postgres"));
        let _ = std::fs::remove_dir_all(&dir);
        std::fs::create_dir_all(&dir).unwrap();
        let times = CreateTimes {
            init_author_millis: 1_791_276_329_745,
            init_committer_millis: 1_791_276_329_756,
            environment_seconds: 1_791_276_329,
            session_seconds: 1_791_276_329,
            commit_millis: 1_791_276_329_815,
        };
        create_database(&dir, "main", "postgres", "localhost", &times).unwrap();
        let noms = dir.join(".dolt/noms");
        let mut db = Database::open(&noms).unwrap();
        let empty_map = db.address_map(&[]).unwrap();
        let empty = db.write_value(empty_map).unwrap();
        let before = map(&mut db, "before", empty);
        let orphan = map(&mut db, "orphan", empty);
        persist(&mut db);

        let mut run = db.gc_begin(GcConfig { mode, archive, incremental_file_size: 0 }).unwrap();
        assert!(
            db.gc_begin(GcConfig { mode, archive, incremental_file_size: 0 }).is_err(),
            "{mode:?}: two collections ran"
        );
        run.copy().unwrap();
        let kept = map(&mut db, "kept", empty);
        let published = map(&mut db, "published", empty);
        let garbage = map(&mut db, "garbage", empty);
        persist(&mut db);
        let pending = map(&mut db, "pending", empty);
        let parent = map(&mut db, "parent", orphan);
        let root_value = db.address_map(&[("root".to_string(), published)]).unwrap();
        db.gc_finish(run, vec![kept], &[root_value]).unwrap();

        for (name, address) in
            [("kept", kept), ("published", published), ("pending", pending), ("empty", empty), ("orphan", orphan)]
        {
            assert!(db.read_value(&address).unwrap().is_some(), "{mode:?}: lost the {name} chunk");
        }
        for (name, address) in [("before", before), ("garbage", garbage)] {
            assert!(db.read_value(&address).unwrap().is_none(), "{mode:?}: kept the unreachable {name} chunk");
        }
        let specs = store::Manifest::read(&noms).unwrap().unwrap().specs;
        let files = specs.iter().filter(|s| s.name.to_string() != store::JOURNAL_FILE).count();
        assert_eq!(files, 1, "{mode:?}: the chunks written during the copy went to another file");
        assert!(
            db.gc_begin(GcConfig { mode, archive, incremental_file_size: 0 }).is_ok(),
            "{mode:?}: stayed collecting"
        );
        persist(&mut db);
        drop(db);
        let db = Database::open(&noms).unwrap();
        for (name, address) in [("kept", kept), ("published", published), ("pending", pending), ("parent", parent)] {
            assert!(db.read_value(&address).unwrap().is_some(), "{mode:?}: lost the {name} chunk after reopening");
        }
    }
}

#[test]
fn chunks_put_before_a_collection_moves_their_children_still_commit() {
    let dir = Path::new(env!("CARGO_TARGET_TMPDIR")).join("gc_phases/moved_children/postgres");
    let _ = std::fs::remove_dir_all(&dir);
    std::fs::create_dir_all(&dir).unwrap();
    let times = CreateTimes {
        init_author_millis: 1_791_276_329_745,
        init_committer_millis: 1_791_276_329_756,
        environment_seconds: 1_791_276_329,
        session_seconds: 1_791_276_329,
        commit_millis: 1_791_276_329_815,
    };
    create_database(&dir, "main", "postgres", "localhost", &times).unwrap();
    let mut db = Database::open(&dir.join(".dolt/noms")).unwrap();
    let head = db.head("refs/heads/main").unwrap().expect("a main branch");
    let mut run = db.gc_begin(GcConfig { mode: GcMode::Default, archive: false, incremental_file_size: 0 }).unwrap();
    run.copy().unwrap();
    let pending = map(&mut db, "pending", head);
    db.gc_finish(run, Vec::new(), &[]).unwrap();
    let root = db.address_map(&[("pending".to_string(), pending)]).unwrap();
    let root = db.write_value(root).unwrap();
    persist(&mut db);
    assert!(db.read_value(&root).unwrap().is_some(), "lost the chunk put before the collection finished");
    assert!(db.read_value(&head).unwrap().is_some(), "lost the commit that the old generation now holds");
}
