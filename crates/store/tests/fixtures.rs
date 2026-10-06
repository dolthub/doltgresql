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

//! Databases written by the Go server, each with `oracle.txt` holding what Dolt's chunk store read from it.

use std::path::{Path, PathBuf};

use store::Hash;

/// fixture returns the directory of a fixture database.
fn fixture(name: &str) -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures").join(name).join("postgres")
}

/// check_fixture checks that the store reads the root and chunks that Dolt read.
fn check_fixture(name: &str) {
    let dir = fixture(name);
    let expected = std::fs::read_to_string(dir.join("oracle.txt")).unwrap();
    let actual = store::dump(&dir.join(".dolt/noms")).unwrap();
    assert!(expected == actual, "{name} differs from Dolt:\n{actual}");
}

#[test]
fn reads_a_journal() {
    check_fixture("journal");
}

#[test]
fn reads_table_files() {
    check_fixture("gc");
}

#[test]
fn reads_archives() {
    check_fixture("archive");
}

#[test]
fn reads_large_chunks_from_archives() {
    check_fixture("large");
}

#[test]
fn hash_strings_round_trip() {
    let text = "b45a39lbakbo1lvppskat4cd0nd5bp9q";
    assert_eq!(Hash::parse(text).unwrap().to_string(), text);
    assert_eq!(Hash::parse("b45a39lbakbo1lvppskat4cd0nd5bp9w"), None);
}
