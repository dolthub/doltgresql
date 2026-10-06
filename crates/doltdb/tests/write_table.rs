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

//! Creating a table and inserting rows gives the object graph that the Go server gave the `onetable` fixture, given
//! the same clock.

use std::path::Path;

use doltdb::create::{CreateTimes, create_database, working_set_ref};
use doltdb::database::Database;
use doltdb::root::Root;
use doltdb::table::Table;
use prolly::val::{build_tuple, compare_tuples, encoding};
use serial::write::{ColumnFields, DEFAULT_TARGET_ROW_SIZE, Meta, SchemaFields, WorkingSetFields, write_schema};
use serial::{Message, WorkingSet};
use store::{GenerationalStore, Hash};

/// INT8 is Doltgres' serialized int8 type.
const INT8: &str = concat!(
    "001223020a0470675f636174616c6f67696e74388008010162014e0001012c0000001323020a0570675f636174616c6f675f",
    "696e74382a0f030a061570675f636174616c6f67696e7438696e23020a0770675f636174616c6f6763737472696e67280f03",
    "0a071270675f636174616c6f67696e74386f757423020a0470675f636174616c6f67696e74382d0f030a081670675f636174",
    "616c6f67696e74387265637623020a0870675f636174616c6f67696e7465726e616c290f030a081270675f636174616c6f67",
    "696e743873656e6423020a0470675f636174616c6f67696e74380000000164017000007fffffff8000000000000000007fff",
    "ffff3d0f040a09121270675f636174616c6f676274696e7438636d7023020a0470675f636174616c6f67696e743823020a04",
    "70675f636174616c6f67696e7438000006626967696e74",
);

/// column returns an int8 column of the test table.
fn column<'a>(name: &'a [u8], sql_type: &'a [u8], tag: u64, primary_key: bool) -> ColumnFields<'a> {
    ColumnFields {
        name,
        sql_type,
        default_value: b"",
        comment: b"",
        on_update: b"",
        tag,
        encoding: encoding::INT64,
        primary_key,
        auto_increment: false,
        nullable: !primary_key,
        generated: false,
        is_virtual: false,
        adaptive_encoding: false,
        hidden: false,
        hidden_system: false,
    }
}

/// commit writes the root value and makes it the working root, as a session's transaction commit does.
fn commit(db: &mut Database, root: &Root, staged: Hash, previous: Hash, seconds: u64) -> Hash {
    let working_root = db.write_value(root.encode()).unwrap();
    let meta = Meta {
        name: b"postgres".to_vec(),
        email: b"postgres@127.0.0.1".to_vec(),
        description: b"sql transaction".to_vec(),
        timestamp_millis: seconds,
        user_timestamp_millis: 0,
    };
    let fields = WorkingSetFields {
        working_root,
        staged_root: Some(staged),
        merge_state: None,
        rebase_state: None,
        meta: Some(meta),
    };
    db.update_working_set(&working_set_ref("main"), &fields, previous).unwrap()
}

#[test]
fn creating_a_table_and_inserting_gives_the_graph_go_gave() {
    let fixture = Path::new(env!("CARGO_MANIFEST_DIR")).join("../store/tests/fixtures/onetable/postgres");
    let dir = Path::new(env!("CARGO_TARGET_TMPDIR")).join("write_table/postgres");
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
    let mut working_set = db.head(&working_set_ref("main")).unwrap().unwrap();
    let data = db.read_value(&working_set).unwrap().unwrap();
    let ws = WorkingSet::new(Message(&data)).unwrap();
    let staged = ws.staged_root().unwrap().unwrap();
    let mut root = Root::decode(&db.read_value(&ws.working_root().unwrap()).unwrap().unwrap()).unwrap();

    // CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);
    let sql_type = format!("extended_{INT8}").into_bytes();
    let schema = write_schema(&SchemaFields {
        columns: vec![column(b"pk", &sql_type, 14384, true), column(b"v1", &sql_type, 15991, false)],
        keyless: false,
        key_columns: vec![0],
        value_columns: vec![1],
        indexes: Vec::new(),
        checks: Vec::new(),
        collation: 309,
        comment: b"",
        target_row_size: DEFAULT_TARGET_ROW_SIZE,
    });
    let (address, mut table) = Table::create(&mut db, schema).unwrap();
    root.put_table(&mut db, "public", "test", Some(address)).unwrap();
    working_set = commit(&mut db, &root, staged, working_set, 1_791_276_329);

    // INSERT INTO test VALUES (1, 1), (2, 2);
    let int = |i: i64| build_tuple(&[Some(&i.to_le_bytes())]);
    let edits = vec![(int(1), Some(int(1))), (int(2), Some(int(2)))];
    table.edit_rows(&mut db, edits, &|a, b| compare_tuples(&[encoding::INT64], a, b)).unwrap();
    let address = table.write(&mut db).unwrap();
    root.put_table(&mut db, "public", "test", Some(address)).unwrap();
    commit(&mut db, &root, staged, working_set, 1_791_276_330);
    db.close().unwrap();

    let expected = std::fs::read_to_string(fixture.join("graph.txt")).unwrap();
    let store = GenerationalStore::open(&dir.join(".dolt/noms")).unwrap();
    let actual = doltdb::dump_graph(&store, store.root()).unwrap();
    assert!(expected == actual, "differs from Go:\n{actual}");
}
