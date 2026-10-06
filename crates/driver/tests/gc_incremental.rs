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

mod common;

use std::path::Path;

use common::*;
use driver::client::Db;
use driver::model::Server;

/// NO_ARCHIVE and SIMPLE_ARCHIVE are the archive levels that dolt_gc's --archive-level option accepts.
const NO_ARCHIVE: i64 = 0;
const SIMPLE_ARCHIVE: i64 = 1;

/// SMALL_FILE_SIZE puts every leaf chunk in its own table file.
const SMALL_FILE_SIZE: i64 = 1;

/// MEDIUM_FILE_SIZE puts several chunks in each table file, leaving one in progress at the end of the mark and
/// sweep.
const MEDIUM_FILE_SIZE: i64 = 5_000;

/// LARGE_FILE_SIZE fits every leaf chunk in a single table file.
const LARGE_FILE_SIZE: i64 = 10_000_000;

/// RepoSize is the size and file count of a database's generations.
#[derive(Debug, Default)]
struct RepoSize {
    journal: u64,
    new_gen: u64,
    new_gen_c: usize,
    old_gen: u64,
    old_gen_c: usize,
}

/// generation_files returns the size and name of each file in the directory, skipping the manifest and lock.
fn generation_files(dir: &Path) -> Vec<(String, u64)> {
    let mut files = Vec::new();
    for entry in std::fs::read_dir(dir).unwrap() {
        let entry = entry.unwrap();
        let name = entry.file_name().to_string_lossy().into_owned();
        let metadata = entry.metadata().unwrap();
        if name == "manifest" || name == "LOCK" || metadata.is_dir() {
            continue;
        }
        files.push((name, metadata.len()));
    }
    files.sort();
    files
}

/// get_repo_size measures the database's new and old generations.
fn get_repo_size(dir: &Path) -> RepoSize {
    let mut size = RepoSize::default();
    for (name, len) in generation_files(&dir.join(".dolt/noms")) {
        size.new_gen += len;
        size.new_gen_c += 1;
        if name == "vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv" {
            size.journal += len;
        }
    }
    for (_, len) in generation_files(&dir.join(".dolt/noms/oldgen")) {
        size.old_gen += len;
        size.old_gen_c += 1;
    }
    size
}

/// populate_db creates a committed table, whose chunks GC moves to the old generation, and an uncommitted one, whose
/// chunks stay in the new generation.
fn populate_db(conn: &mut Db) {
    exec(conn, "create table vals (id bigint primary key, val bigint)", &[]);
    let vals: Vec<String> = (0..=1024).map(|i| format!("({i},0)")).collect();
    exec(conn, &format!("insert into vals values {}", vals.join(",")), &[]);
    exec(conn, "select dolt_commit('-Am', 'create vals table')", &[]);
    exec(conn, "create table vals2 (id bigint primary key, val bigint)", &[]);
    exec(conn, &format!("insert into vals2 values {}", vals.join(",")), &[]);
}

/// run_gc_incremental_test runs an incremental GC and checks that the database is still queryable, standing in for
/// Dolt's fsck, and that both generations hold several files.
fn run_gc_incremental_test(archive_level: i64, file_size: i64, full: bool) {
    let mut env = env();
    let repo_name = format!("incremental_gc_test_archivelevel_{archive_level}_filesize_{file_size}_full_{full}");
    let repo = start_server(&mut env, &repo_name, Server { dynamic_port: "server".into(), ..Server::default() }, None);
    let mut conn = db(&mut env);
    populate_db(&mut conn);
    let full_arg = if full { ",'--full'" } else { "" };
    let gc_sql = format!(
        "select dolt_gc('--archive-level','{archive_level}','--incremental-file-size','{file_size}'{full_arg})"
    );
    exec(&mut conn, &gc_sql, &[]);
    conn.close();

    let mut conn = db(&mut env);
    assert!(query_int(&mut conn, "select count(*) from vals", &[]) > 0);
    conn.close();

    let size = get_repo_size(&repo.dir);
    assert!(size.new_gen_c > 1, "{size:?}");
    assert!(size.old_gen_c > 1, "{size:?}");
    finish(env);
}

/// require_file_added_to_manifest checks whether the generation's first table file is listed in its manifest.
fn require_file_added_to_manifest(path: &Path, expected: bool) {
    let manifest = match std::fs::read(path.join("manifest")) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            assert!(!expected);
            return;
        }
        manifest => manifest.unwrap(),
    };
    let files = generation_files(path);
    let (table_file, _) = files.first().expect("no table file");
    let manifest = String::from_utf8_lossy(&manifest[..manifest.len().min(200)]).into_owned();
    assert_eq!(manifest.contains(table_file.as_str()), expected, "{table_file} in {manifest}");
}

#[test]
fn test_gc_incremental_archive_level_0_file_size_1_full_true() {
    run_gc_incremental_test(NO_ARCHIVE, SMALL_FILE_SIZE, true);
}

#[test]
fn test_gc_incremental_archive_level_0_file_size_1_full_false() {
    run_gc_incremental_test(NO_ARCHIVE, SMALL_FILE_SIZE, false);
}

#[test]
fn test_gc_incremental_archive_level_0_file_size_5000_full_true() {
    run_gc_incremental_test(NO_ARCHIVE, MEDIUM_FILE_SIZE, true);
}

#[test]
fn test_gc_incremental_archive_level_0_file_size_5000_full_false() {
    run_gc_incremental_test(NO_ARCHIVE, MEDIUM_FILE_SIZE, false);
}

#[test]
fn test_gc_incremental_archive_level_0_file_size_10000000_full_true() {
    run_gc_incremental_test(NO_ARCHIVE, LARGE_FILE_SIZE, true);
}

#[test]
fn test_gc_incremental_archive_level_0_file_size_10000000_full_false() {
    run_gc_incremental_test(NO_ARCHIVE, LARGE_FILE_SIZE, false);
}

#[test]
fn test_gc_incremental_archive_level_1_file_size_1_full_true() {
    run_gc_incremental_test(SIMPLE_ARCHIVE, SMALL_FILE_SIZE, true);
}

#[test]
fn test_gc_incremental_archive_level_1_file_size_1_full_false() {
    run_gc_incremental_test(SIMPLE_ARCHIVE, SMALL_FILE_SIZE, false);
}

#[test]
fn test_gc_incremental_archive_level_1_file_size_5000_full_true() {
    run_gc_incremental_test(SIMPLE_ARCHIVE, MEDIUM_FILE_SIZE, true);
}

#[test]
fn test_gc_incremental_archive_level_1_file_size_5000_full_false() {
    run_gc_incremental_test(SIMPLE_ARCHIVE, MEDIUM_FILE_SIZE, false);
}

#[test]
fn test_gc_incremental_archive_level_1_file_size_10000000_full_true() {
    run_gc_incremental_test(SIMPLE_ARCHIVE, LARGE_FILE_SIZE, true);
}

#[test]
fn test_gc_incremental_archive_level_1_file_size_10000000_full_false() {
    run_gc_incremental_test(SIMPLE_ARCHIVE, LARGE_FILE_SIZE, false);
}

#[test]
fn test_resumable_gc() {
    let mut env = env();
    let server = Server {
        dynamic_port: "server".into(),
        envs: vec!["DOLT_TEST_ABORT_GC_AFTER_INCREMENTAL_FILE_WRITE=true".into()],
        ..Server::default()
    };
    let repo = start_server(&mut env, "resumable_gc_test", server, None);
    let oldgen = repo.dir.join(".dolt/noms/oldgen");
    let mut conn = db(&mut env);
    populate_db(&mut conn);

    let gc = "select dolt_gc('--archive-level','0','--incremental-file-size','1','--full')";
    assert!(conn.exec(gc, &[]).is_err(), "GC aborting after writing incremental table file");
    require_file_added_to_manifest(&oldgen, false);

    let gc = "select dolt_gc('--archive-level','0','--incremental-file-size','1')";
    assert!(conn.exec(gc, &[]).is_err(), "GC aborting after writing incremental table file");
    require_file_added_to_manifest(&oldgen, true);

    let _ = conn.exec(gc, &[]);

    exec(&mut conn, "insert into vals2 values (1025, 0)", &[]);
    assert!(conn.exec(gc, &[]).is_err(), "GC aborting after writing incremental table file");
    require_file_added_to_manifest(&oldgen, false);
    conn.close();
    finish(env);
}
