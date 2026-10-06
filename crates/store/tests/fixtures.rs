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

//! Databases written by the Go server, by older Doltgres releases, and by Dolt releases that wrote older archive
//! formats, each with `oracle.txt` holding what the current Dolt chunk store read from it.

use std::path::{Path, PathBuf};

use store::{
    Chunk, GenerationalStore, Hash, JOURNAL_FILE, JOURNAL_INDEX_FILE, JournalRecord, JournalWriter, MANIFEST_FILE,
    Manifest, TableReader, TableWriter, lock_hash, read_records,
};

/// fixture returns the directory of a fixture's database, the one directory in it that holds a `.dolt` directory.
fn fixture(name: &str) -> PathBuf {
    let dir = Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures").join(name);
    std::fs::read_dir(&dir)
        .unwrap()
        .map(|entry| entry.unwrap().path())
        .find(|path| path.join(".dolt").is_dir())
        .unwrap_or_else(|| panic!("no database in {}", dir.display()))
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
fn reads_a_merge_with_conflicts() {
    check_fixture("conflict");
}

#[test]
fn reads_root_objects_tags_and_many_types() {
    check_fixture("rich");
}

#[test]
fn reads_many_tables_and_branches() {
    check_fixture("wide");
}

#[test]
fn reads_stashes_and_rebase_state() {
    check_fixture("states");
}

#[test]
fn reads_table_files_from_doltgres_0_50() {
    check_fixture("doltgres-v0.50.0-gc");
}

#[test]
fn reads_a_journal_from_doltgres_0_56() {
    check_fixture("doltgres-v0.56.0");
}

#[test]
fn reads_archives_from_doltgres_0_57() {
    check_fixture("doltgres-v0.57.0-gc");
}

#[test]
fn reads_a_journal_from_doltgres_1_0() {
    check_fixture("doltgres-v1.0.0");
}

#[test]
fn reads_version_1_archives() {
    check_fixture("dolt-v1.45.0-archive-v1");
}

#[test]
fn reads_version_2_archives() {
    check_fixture("dolt-v1.51.0-archive-v2");
}

#[test]
fn hash_strings_round_trip() {
    let text = "b45a39lbakbo1lvppskat4cd0nd5bp9q";
    assert_eq!(Hash::parse(text).unwrap().to_string(), text);
    assert_eq!(Hash::parse("b45a39lbakbo1lvppskat4cd0nd5bp9w"), None);
}

#[test]
fn reads_constraints_and_indexes() {
    check_fixture("schemas");
}

#[test]
fn reads_blobs_of_every_size() {
    check_fixture("blobs");
}

#[test]
fn snappy_compresses_every_chunk_as_go_did() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures");
    let (mut checked, mut failures) = (0, Vec::new());
    for entry in std::fs::read_dir(&fixtures).unwrap() {
        let path = entry.unwrap().path();
        if !path.is_dir() {
            continue;
        }
        let noms = fixture(path.file_name().unwrap().to_str().unwrap()).join(".dolt/noms");
        let store = GenerationalStore::open(&noms).unwrap();
        for generation in [&store.new_gen, &store.old_gen] {
            generation
                .for_each_record(&mut |hash, record| {
                    checked += 1;
                    if Chunk::from_record(hash, record)?.to_record() != record {
                        failures.push(format!("{}: {hash}", noms.display()));
                    }
                    Ok(())
                })
                .unwrap();
        }
    }
    assert!(checked > 1000, "only {checked} records were checked");
    assert!(failures.is_empty(), "{} of {checked} differ:\n{}", failures.len(), failures.join("\n"));
}

#[test]
fn table_writer_rewrites_every_table_file_go_wrote() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures");
    let (mut checked, mut failures) = (0, Vec::new());
    for entry in std::fs::read_dir(&fixtures).unwrap() {
        let path = entry.unwrap().path();
        if !path.is_dir() {
            continue;
        }
        let noms = fixture(path.file_name().unwrap().to_str().unwrap()).join(".dolt/noms");
        for dir in [noms.clone(), noms.join("oldgen")] {
            let Ok(entries) = std::fs::read_dir(&dir) else { continue };
            for file in entries {
                let file = file.unwrap().path();
                let name = file.file_name().unwrap().to_str().unwrap().to_string();
                if name.len() != 32 || name.chars().all(|c| c == 'v') || Hash::parse(&name).is_none() {
                    continue;
                }
                let bytes = std::fs::read(&file).unwrap();
                if bytes.ends_with(b"DOLTARC") {
                    continue;
                }
                let table = TableReader::open(&file).unwrap();
                let mut writer = TableWriter::new();
                table
                    .for_each_record(&mut |hash, record| {
                        writer.add_record(hash, record, Chunk::from_record(hash, record)?.data.len() as u64);
                        Ok(())
                    })
                    .unwrap();
                let (written_name, written) = writer.finish();
                checked += 1;
                if written_name.to_string() != name || written != bytes {
                    failures.push(format!("{}: written as {written_name}", file.display()));
                }
            }
        }
    }
    assert!(checked >= 4, "only {checked} table files were checked");
    assert!(failures.is_empty(), "{} of {checked} differ:\n{}", failures.len(), failures.join("\n"));
}

#[test]
fn manifests_format_to_the_text_go_wrote() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures");
    let (mut checked, mut failures) = (0, Vec::new());
    for entry in std::fs::read_dir(&fixtures).unwrap() {
        let path = entry.unwrap().path();
        if !path.is_dir() {
            continue;
        }
        let noms = fixture(path.file_name().unwrap().to_str().unwrap()).join(".dolt/noms");
        for dir in [noms.clone(), noms.join("oldgen")] {
            let Ok(text) = std::fs::read_to_string(dir.join(MANIFEST_FILE)) else { continue };
            let manifest = Manifest::parse(text.as_bytes()).unwrap();
            checked += 1;
            if manifest.format() != text {
                failures.push(format!("{}: formatted as {}", dir.display(), manifest.format()));
            }
            if manifest.lock != lock_hash(&manifest.root, &manifest.specs, &[], b"") {
                failures.push(format!(
                    "{}: lock {} is not the hash of the root and files",
                    dir.display(),
                    manifest.lock
                ));
            }
        }
    }
    assert!(checked > 10, "only {checked} manifests were checked");
    assert!(failures.is_empty(), "{} of {checked} differ:\n{}", failures.len(), failures.join("\n"));
}

#[test]
fn journal_records_encode_to_the_bytes_go_wrote() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures");
    let (mut checked, mut roots, mut failures) = (0, 0, Vec::new());
    for entry in std::fs::read_dir(&fixtures).unwrap() {
        let path = entry.unwrap().path();
        if !path.is_dir() {
            continue;
        }
        let journal = fixture(path.file_name().unwrap().to_str().unwrap()).join(".dolt/noms").join(JOURNAL_FILE);
        let Ok(bytes) = std::fs::read(&journal) else { continue };
        for (record, raw) in read_records(&bytes).unwrap() {
            checked += 1;
            roots += matches!(record, JournalRecord::Root { .. }) as usize;
            if record.encode() != raw {
                failures.push(format!("{}: {record:?}", journal.display()));
            }
        }
    }
    assert!(checked > 1000 && roots > 10, "only {checked} records and {roots} roots were checked");
    assert!(failures.is_empty(), "{} of {checked} differ:\n{}", failures.len(), failures.join("\n"));
}

#[test]
fn journal_writer_writes_the_journals_go_wrote() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures");
    let scratch = Path::new(env!("CARGO_TARGET_TMPDIR")).join("journal_writer");
    let (mut checked, mut failures) = (0, Vec::new());
    for entry in std::fs::read_dir(&fixtures).unwrap() {
        let path = entry.unwrap().path();
        if !path.is_dir() {
            continue;
        }
        let name = path.file_name().unwrap().to_str().unwrap().to_string();
        let journal = fixture(&name).join(".dolt/noms").join(JOURNAL_FILE);
        let Ok(bytes) = std::fs::read(&journal) else { continue };
        let dir = scratch.join(&name);
        let _ = std::fs::remove_dir_all(&dir);
        std::fs::create_dir_all(&dir).unwrap();
        let records = read_records(&bytes).unwrap();
        let (mut writer, root) = JournalWriter::open(&dir).unwrap();
        assert!(root.is_empty());
        let mut last_root = Hash::default();
        for (record, _) in &records {
            match *record {
                JournalRecord::Chunk { hash, record } => writer.write_chunk(hash, record).unwrap(),
                JournalRecord::Root { hash, timestamp } => {
                    writer.commit_root_at(hash, timestamp).unwrap();
                    last_root = hash;
                }
            }
        }
        writer.close().unwrap();
        checked += 1;
        let written = std::fs::read(dir.join(JOURNAL_FILE)).unwrap();
        let valid_len: usize = records.iter().map(|(_, raw)| raw.len()).sum();
        if written != bytes[..valid_len] {
            failures.push(format!("{name}: the journal differs"));
            continue;
        }
        // Reopening reads the root and every chunk back, through the index file and the journal.
        let (writer, root) = JournalWriter::open(&dir).unwrap();
        assert_eq!(root, last_root, "{name}");
        for (record, _) in &records {
            if let JournalRecord::Chunk { hash, record } = *record {
                assert_eq!(writer.get(&hash).unwrap(), Some(Chunk::from_record(hash, record).unwrap()), "{name}");
            }
        }
        writer.close().unwrap();
    }
    assert!(checked > 5, "only {checked} journals were checked");
    assert!(failures.is_empty(), "{} of {checked} differ:\n{}", failures.len(), failures.join("\n"));
}

/// write_indexed_journal writes a store whose journal holds enough chunks over several commits to end index batches,
/// with a manifest naming the journal, and returns its chunks and last root.
fn write_indexed_journal(dir: &Path) -> (Vec<Chunk>, Hash) {
    let _ = std::fs::remove_dir_all(dir);
    std::fs::create_dir_all(dir.join("oldgen")).unwrap();
    let (mut writer, _) = JournalWriter::open(dir).unwrap();
    let mut chunks = Vec::new();
    let mut root = Hash::default();
    for commit in 0..5u32 {
        for i in 0..9000u32 {
            let chunk = Chunk::new(format!("chunk {commit} {i}").into_bytes());
            writer.write_chunk(chunk.hash, &chunk.to_record()).unwrap();
            chunks.push(chunk);
        }
        root = chunks.last().unwrap().hash;
        writer.commit_root_at(root, 1_700_000_000 + commit as u64).unwrap();
    }
    writer.close().unwrap();
    let specs = vec![store::TableSpec { name: Hash::parse(JOURNAL_FILE).unwrap(), chunk_count: chunks.len() as u32 }];
    let manifest = Manifest {
        version: "5".to_string(),
        format: "__DOLT__".to_string(),
        lock: lock_hash(&root, &specs, &[], b""),
        root,
        gc_gen: Hash::default(),
        specs,
    };
    std::fs::write(dir.join(MANIFEST_FILE), manifest.format()).unwrap();
    (chunks, root)
}

#[test]
fn journal_writer_indexes_large_journals() {
    let dir = Path::new(env!("CARGO_TARGET_TMPDIR")).join("indexed_journal");
    let (chunks, root) = write_indexed_journal(&dir);
    let index = std::fs::read(dir.join(JOURNAL_INDEX_FILE)).unwrap();
    // Each lookup is 29 bytes and each meta record 41, and the second and fourth commits end batches.
    assert_eq!(index.len(), chunks.len() * 29 + 2 * 41);
    let (writer, reopened_root) = JournalWriter::open(&dir).unwrap();
    assert_eq!(reopened_root, root);
    for chunk in &chunks {
        assert_eq!(writer.get(&chunk.hash).unwrap().as_ref(), Some(chunk));
    }
    writer.close().unwrap();
    assert_eq!(std::fs::read(dir.join(JOURNAL_INDEX_FILE)).unwrap(), index, "reopening changed the index");
}

#[test]
#[ignore = "needs the Go store oracle that testing/go/regression/out/build_store_fixtures.sh builds"]
fn go_reads_journals_and_indexes_rust_wrote() {
    let oracle = Path::new(env!("CARGO_MANIFEST_DIR")).join("../../target/storeoracle");
    let dir = Path::new(env!("CARGO_TARGET_TMPDIR")).join("indexed_journal_for_go");
    write_indexed_journal(&dir);
    let index = std::fs::read(dir.join(JOURNAL_INDEX_FILE)).unwrap();
    let expected = store::dump(&dir).unwrap();
    let output = std::process::Command::new(&oracle).arg(&dir).output().unwrap();
    assert!(output.status.success(), "{}", String::from_utf8_lossy(&output.stderr));
    // Go reports the chunks it found through the index by the first 16 bytes of their addresses, so the chunks are
    // compared by generation, size, and the hash of their data.
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
    assert!(chunks(&actual) == chunks(&expected), "Go reads other chunks than Rust");
    // Go truncates an index it rejects, so its index starts with Rust's when it accepted it.
    let after = std::fs::read(dir.join(JOURNAL_INDEX_FILE)).unwrap();
    assert!(after.starts_with(&index), "Go rewrote the index ({} bytes, then {})", index.len(), after.len());
}
