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

//! The store crate's fixture databases, each with `graph.txt` holding the object graph that Dolt and Doltgres read.

use std::path::{Path, PathBuf};

use store::GenerationalStore;

/// fixture returns the directory of a fixture database.
fn fixture(name: &str) -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("../store/tests/fixtures").join(name).join("postgres")
}

/// check_fixture checks that the object graph matches the one Dolt and Doltgres read.
fn check_fixture(name: &str) {
    let dir = fixture(name);
    let expected = std::fs::read_to_string(dir.join("graph.txt")).unwrap();
    let store = GenerationalStore::open(&dir.join(".dolt/noms")).unwrap();
    let actual = doltdb::dump_graph(&store, store.root()).unwrap();
    assert!(expected == actual, "{name} differs from Dolt:\n{actual}");
}

#[test]
fn reads_commits_and_working_sets() {
    check_fixture("journal");
}

#[test]
fn reads_the_graph_from_table_files() {
    check_fixture("gc");
}

#[test]
fn reads_the_graph_from_archives() {
    check_fixture("archive");
}

#[test]
fn reads_tables_with_large_values() {
    check_fixture("large");
}

#[test]
fn reads_merge_state() {
    check_fixture("conflict");
}

#[test]
fn reads_root_objects_and_tags() {
    check_fixture("rich");
}

#[test]
fn reads_multi_level_address_maps() {
    check_fixture("wide");
}
