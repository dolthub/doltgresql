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

//! Creating a database writes what the Go server wrote when it created the `empty` fixture, given the same clock.

use std::path::Path;

use doltdb::create::{CreateTimes, create_database};
use store::{JOURNAL_FILE, JournalRecord, read_records};

#[test]
fn creating_a_database_writes_the_journal_go_wrote() {
    let fixture = Path::new(env!("CARGO_MANIFEST_DIR")).join("../store/tests/fixtures/empty/postgres/.dolt");
    let dir = Path::new(env!("CARGO_TARGET_TMPDIR")).join("create/postgres");
    let _ = std::fs::remove_dir_all(&dir);
    std::fs::create_dir_all(&dir).unwrap();
    // The clock readings in the fixture's commits and working sets.
    let times = CreateTimes {
        init_author_millis: 1_791_273_086_698,
        init_committer_millis: 1_791_273_086_708,
        environment_seconds: 1_791_273_086,
        session_seconds: 1_791_273_086,
        commit_millis: 1_791_273_086_754,
    };
    create_database(&dir, "main", "postgres", "localhost", &times).unwrap();

    let expected = std::fs::read(fixture.join("noms").join(JOURNAL_FILE)).unwrap();
    let actual = std::fs::read(dir.join(".dolt/noms").join(JOURNAL_FILE)).unwrap();
    let (expected, actual) = (read_records(&expected).unwrap(), read_records(&actual).unwrap());
    // Root hash records carry the time they were written, which only the hashes they hold are compared without.
    let describe = |record: &JournalRecord<'_>, raw: &[u8]| match record {
        JournalRecord::Root { hash, .. } => format!("root {hash}"),
        JournalRecord::Chunk { hash, .. } => format!("chunk {hash} {}", raw.len()),
    };
    let expected: Vec<String> = expected.iter().map(|(r, raw)| describe(r, raw)).collect();
    let actual: Vec<String> = actual.iter().map(|(r, raw)| describe(r, raw)).collect();
    assert_eq!(actual, expected);
    for file in ["config.json", "repo_state.json"] {
        assert_eq!(
            std::fs::read_to_string(dir.join(".dolt").join(file)).unwrap(),
            std::fs::read_to_string(fixture.join(file)).unwrap(),
            "{file}"
        );
    }
    let manifest = |dir: &Path| store::Manifest::read(&dir.join("noms")).unwrap().unwrap();
    let (expected, actual) = (manifest(&fixture), manifest(&dir.join(".dolt")));
    assert_eq!(actual.specs, expected.specs, "the manifest's files");
}
