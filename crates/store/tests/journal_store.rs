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

//! The writable journaling store, on new databases and on copies of databases the Go server wrote.

use std::path::{Path, PathBuf};

use store::{BlockStore, Chunk, Hash, JOURNAL_FILE, JournalStore, Manifest};

/// scratch returns an empty directory for a test.
fn scratch(name: &str) -> PathBuf {
    let dir = Path::new(env!("CARGO_TARGET_TMPDIR")).join("journal_store").join(name);
    let _ = std::fs::remove_dir_all(&dir);
    std::fs::create_dir_all(dir.join("oldgen")).unwrap();
    dir
}

/// copy_dir copies a directory tree.
fn copy_dir(from: &Path, to: &Path) {
    std::fs::create_dir_all(to).unwrap();
    for entry in std::fs::read_dir(from).unwrap() {
        let path = entry.unwrap().path();
        let target = to.join(path.file_name().unwrap());
        if path.is_dir() {
            copy_dir(&path, &target);
        } else {
            std::fs::copy(&path, &target).unwrap();
        }
    }
}

/// chunks returns distinct chunks for a commit.
fn chunks(commit: u32, count: u32) -> Vec<Chunk> {
    (0..count).map(|i| Chunk::new(format!("commit {commit} chunk {i}").into_bytes())).collect()
}

#[test]
fn new_stores_commit_and_reopen() {
    let dir = scratch("new");
    let mut store = JournalStore::open(&dir, "__DOLT__").unwrap();
    assert!(store.root().is_empty());
    let first = chunks(1, 10);
    for chunk in &first {
        store.put(chunk.clone(), []).unwrap();
    }
    assert_eq!(store.get(&first[3].hash).unwrap().as_ref(), Some(&first[3]));
    assert!(!store.commit(first[0].hash, first[1].hash).unwrap(), "a commit from another root succeeded");
    assert!(store.commit(first[0].hash, Hash::default()).unwrap());
    let second = chunks(2, 10);
    for chunk in &second {
        store.put(chunk.clone(), []).unwrap();
    }
    assert!(store.commit(second[0].hash, first[0].hash).unwrap());
    store.close().unwrap();

    let manifest = Manifest::read(&dir).unwrap().unwrap();
    assert_eq!(manifest.specs.len(), 1);
    assert_eq!(manifest.specs[0].name.to_string(), JOURNAL_FILE);
    let reader = BlockStore::open(&dir).unwrap();
    assert_eq!(reader.root(), second[0].hash);
    let store = JournalStore::open(&dir, "__DOLT__").unwrap();
    assert_eq!(store.root(), second[0].hash);
    for chunk in first.iter().chain(&second) {
        assert_eq!(reader.get(&chunk.hash).unwrap().as_ref(), Some(chunk));
        assert_eq!(store.get(&chunk.hash).unwrap().as_ref(), Some(chunk));
    }
    store.close().unwrap();
}

#[test]
fn stores_lock_out_other_writers() {
    let dir = scratch("locked");
    let store = JournalStore::open(&dir, "__DOLT__").unwrap();
    assert!(JournalStore::open(&dir, "__DOLT__").is_err());
    store.close().unwrap();
    JournalStore::open(&dir, "__DOLT__").unwrap().close().unwrap();
}

/// extend_fixture copies a Go-written fixture database and commits new chunks on top of its root.
fn extend_fixture(name: &str) -> (PathBuf, Vec<Chunk>, Hash) {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures").join(name);
    let database =
        std::fs::read_dir(&fixtures).unwrap().map(|e| e.unwrap().path()).find(|p| p.join(".dolt").is_dir()).unwrap();
    let dir = scratch(&format!("extended-{name}"));
    std::fs::remove_dir_all(&dir).unwrap();
    copy_dir(&database.join(".dolt/noms"), &dir);
    let mut store = JournalStore::open(&dir, "__DOLT__").unwrap();
    let before = BlockStore::open(&dir).unwrap().root();
    assert_eq!(store.root(), before);
    let added = chunks(3, 100);
    for chunk in &added {
        store.put(chunk.clone(), []).unwrap();
    }
    let root = added[0].hash;
    assert!(store.commit(root, before).unwrap());
    store.close().unwrap();
    (dir, added, root)
}

#[test]
fn stores_extend_databases_go_wrote() {
    for name in ["journal", "gc", "archive", "doltgres-v0.50.0-gc"] {
        let (dir, added, root) = extend_fixture(name);
        let reader = BlockStore::open(&dir).unwrap();
        assert_eq!(reader.root(), root, "{name}");
        for chunk in &added {
            assert_eq!(reader.get(&chunk.hash).unwrap().as_ref(), Some(chunk), "{name}");
        }
        // A commit that keeps the same files leaves the manifest's root to the next open, as in Go.
        let before = Manifest::read(&dir).unwrap().unwrap();
        assert_ne!(before.root, root, "{name}: the manifest's root");
        JournalStore::open(&dir, "__DOLT__").unwrap().close().unwrap();
        let after = Manifest::read(&dir).unwrap().unwrap();
        assert_eq!(after.root, root, "{name}: the manifest's root after reopening");
        assert_eq!(after.specs, before.specs, "{name}: the manifest's files after reopening");
    }
}

#[test]
fn commits_fail_on_dangling_references() {
    let dir = scratch("dangling");
    let mut store = JournalStore::open(&dir, "__DOLT__").unwrap();
    let missing = Chunk::new(b"never put".to_vec());
    let parent = Chunk::new(b"parent".to_vec());
    store.put(parent.clone(), [missing.hash]).unwrap();
    let err = store.commit(parent.hash, Hash::default()).unwrap_err();
    assert!(matches!(&err, store::Error::DanglingRef(hashes) if hashes == &[missing.hash]), "{err}");
    assert!(!store.has(&parent.hash), "the memtable survived a dangling reference");
    let child = Chunk::new(b"child".to_vec());
    store.put(child.clone(), []).unwrap();
    store.put(parent.clone(), [child.hash]).unwrap();
    assert!(store.commit(parent.hash, Hash::default()).unwrap());
    let err = store.commit(missing.hash, parent.hash).unwrap_err();
    assert!(matches!(&err, store::Error::DanglingRef(hashes) if hashes == &[missing.hash]), "{err}");
    store.close().unwrap();
}

#[test]
#[ignore = "needs the Go store oracle that testing/go/regression/out/build_store_fixtures.sh builds"]
fn go_reads_databases_rust_extended() {
    let oracle = Path::new(env!("CARGO_MANIFEST_DIR")).join("../../target/storeoracle");
    for name in ["journal", "gc", "archive", "doltgres-v0.50.0-gc"] {
        let (dir, _, _) = extend_fixture(name);
        let expected = store::dump(&dir).unwrap();
        let output = std::process::Command::new(&oracle).arg(&dir).output().unwrap();
        assert!(output.status.success(), "{name}: {}", String::from_utf8_lossy(&output.stderr));
        // Go reports the chunks it found through the journal's index by the first 16 bytes of their addresses.
        let chunks = |dump: &str| {
            let mut lines: Vec<String> = dump
                .lines()
                .map(|line| match line.split(' ').collect::<Vec<_>>()[..] {
                    [generation, _, size, data] => format!("{generation} {size} {data}"),
                    _ => line.to_string(),
                })
                .collect();
            lines.sort();
            lines
        };
        let actual = String::from_utf8(output.stdout).unwrap();
        assert!(chunks(&actual) == chunks(&expected), "{name}: Go reads other chunks than Rust");
    }
}

#[test]
fn deferred_syncs_are_taken_by_the_caller() {
    let dir = scratch("deferred");
    let mut store = JournalStore::open(&dir, "__DOLT__").unwrap();
    let first = chunks(1, 5);
    for chunk in &first {
        store.put(chunk.clone(), []).unwrap();
    }
    assert!(store.commit(first[0].hash, Hash::default()).unwrap());
    assert!(store::take_sync().is_none(), "a commit without deferral left a sync");
    store::defer_syncs(true);
    let second = chunks(2, 5);
    for chunk in &second {
        store.put(chunk.clone(), []).unwrap();
    }
    assert!(store.commit(second[0].hash, first[0].hash).unwrap());
    let third = chunks(3, 5);
    for chunk in &third {
        store.put(chunk.clone(), []).unwrap();
    }
    assert!(store.commit(third[0].hash, second[0].hash).unwrap());
    store::defer_syncs(false);
    let sync = store::take_sync().expect("deferred commits left no sync");
    assert!(store::take_sync().is_none(), "a sync was taken twice");
    sync.wait().unwrap();
    store.close().unwrap();

    let store = JournalStore::open(&dir, "__DOLT__").unwrap();
    assert_eq!(store.root(), third[0].hash);
    for chunk in first.iter().chain(&second).chain(&third) {
        assert_eq!(store.get(&chunk.hash).unwrap().as_ref(), Some(chunk));
    }
    store.close().unwrap();
}

#[test]
fn journals_reopen_past_padding_and_close_without_it() {
    let dir = scratch("padded");
    let mut store = JournalStore::open(&dir, "__DOLT__").unwrap();
    let first = chunks(1, 5);
    for chunk in &first {
        store.put(chunk.clone(), []).unwrap();
    }
    assert!(store.commit(first[0].hash, Hash::default()).unwrap());
    store.close().unwrap();
    let journal = dir.join(JOURNAL_FILE);
    let mut bytes = std::fs::read(&journal).unwrap();
    bytes.extend(std::iter::repeat_n(0, 1 << 16));
    std::fs::write(&journal, &bytes).unwrap();

    let mut store = JournalStore::open(&dir, "__DOLT__").unwrap();
    assert_eq!(store.root(), first[0].hash);
    let second = chunks(2, 5);
    for chunk in &second {
        store.put(chunk.clone(), []).unwrap();
    }
    assert!(store.commit(second[0].hash, first[0].hash).unwrap());
    store.close().unwrap();
    let bytes = std::fs::read(&journal).unwrap();
    let records = store::read_records(&bytes).unwrap();
    assert_eq!(records.iter().map(|(_, raw)| raw.len()).sum::<usize>(), bytes.len(), "padding was left behind");
    let store = JournalStore::open(&dir, "__DOLT__").unwrap();
    assert_eq!(store.root(), second[0].hash);
    for chunk in first.iter().chain(&second) {
        assert_eq!(store.get(&chunk.hash).unwrap().as_ref(), Some(chunk));
    }
    store.close().unwrap();
}

#[test]
fn stale_spill_files_are_removed() {
    let dir = scratch("spills");
    std::fs::write(dir.join(".spill-1-7.tmp"), b"partial archive").unwrap();
    std::fs::write(dir.join("keep.tmp"), b"not a spill").unwrap();
    store::remove_spills(&dir);
    assert!(!dir.join(".spill-1-7.tmp").exists(), "a stale spill file stayed");
    assert!(dir.join("keep.tmp").exists(), "a file that isn't a spill was removed");
}
