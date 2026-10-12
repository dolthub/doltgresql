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
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

use common::*;
use driver::model::Server;
use driver::runner::Env;

/// is_table_file_name reports whether the name is 32 characters of Noms' base32 hash alphabet.
fn is_table_file_name(name: &str) -> bool {
    name.len() == 32 && name.bytes().all(|c| c.is_ascii_digit() || (b'a'..=b'v').contains(&c))
}

/// count_table_files counts the table files under the directory, including archives.
fn count_table_files(dir: &Path) -> usize {
    let mut count = 0;
    for entry in std::fs::read_dir(dir).unwrap_or_else(|e| panic!("{}: {e}", dir.display())) {
        let entry = entry.unwrap();
        if entry.file_type().unwrap().is_dir() {
            count += count_table_files(&entry.path());
        }
        let name = entry.file_name().to_string_lossy().into_owned();
        if is_table_file_name(name.strip_suffix(".darc").unwrap_or(&name)) {
            count += 1;
        }
    }
    count
}

/// commit_and_gc makes an empty commit and runs GC over a new connection, since a GC invalidates the old ones.
fn commit_and_gc(env: &mut Env) {
    let mut conn = db(env);
    exec(&mut conn, "SELECT DOLT_COMMIT('-A', '--allow-empty', '-m', 'creating a new commit')", &[]);
    exec(&mut conn, "SELECT DOLT_GC()", &[]);
}

/// upstream_len reads the number after `upstream_len=` like fmt.Sscanf's %d, or -1 when there is none.
fn upstream_len(line: &str, index: usize) -> i64 {
    let rest = &line[index + "upstream_len=".len()..];
    let end = rest
        .char_indices()
        .find(|&(i, c)| !(c.is_ascii_digit() || (i == 0 && (c == '+' || c == '-'))))
        .map_or(rest.len(), |(i, _)| i);
    rest[..end].parse().unwrap_or(-1)
}

#[test]
fn test_full_gc_no_oldgen_conjoin() {
    let dbname = "full_gc_no_oldgen_conjoin_test";
    let mut env = env();
    let repo = make_repo(&mut env, dbname);
    let settings = Server { dynamic_port: "server".into(), ..Server::default() };
    let conjoin_started = Arc::new(AtomicBool::new(false));
    let conjoin_finished = Arc::new(AtomicBool::new(false));
    let upstream_len_on_conjoin = Arc::new(Mutex::new(0i64));
    let (started, finished, len) = (conjoin_started.clone(), conjoin_finished.clone(), upstream_len_on_conjoin.clone());
    let visitor: driver::process::Visitor = Arc::new(move |out: &str| {
        let mut len = len.lock().unwrap();
        if *len <= 0 && out.contains("beginning conjoin of database") {
            if let Some(i) = out.find("upstream_len=") {
                *len = upstream_len(out, i);
            }
            started.store(true, Ordering::SeqCst);
        }
        if out.contains("conjoin completed successfully") {
            finished.store(true, Ordering::SeqCst);
        }
    });
    start(&mut env, SERVER, &repo, settings.clone(), Some(visitor));
    let oldgen_dir = repo.dir.join(".dolt/noms/oldgen");

    loop {
        commit_and_gc(&mut env);
        if conjoin_started.load(Ordering::SeqCst) {
            break;
        }
    }
    let target = *upstream_len_on_conjoin.lock().unwrap();
    assert!(target > 0);
    let deadline = Instant::now() + Duration::from_secs(5);
    while !conjoin_finished.load(Ordering::SeqCst) {
        assert!(Instant::now() < deadline, "conjoin did not finish");
        std::thread::sleep(Duration::from_millis(32));
    }

    while count_table_files(&oldgen_dir) as i64 != target {
        commit_and_gc(&mut env);
    }
    println!("now there are {}", count_table_files(&oldgen_dir));

    let mut conn = db(&mut env);
    count_table_files(&oldgen_dir);
    exec(&mut conn, "SELECT DOLT_GC('--full')", &[]);
    assert_eq!(count_table_files(&oldgen_dir), 1);
    conn.close();

    env.server(SERVER).graceful_stop().unwrap();
    let output = env.server(SERVER).output_text();
    assert_eq!(output.matches("beginning conjoin of database").count(), 1);
    assert_eq!(output.matches("conjoin dynamically disabled").count(), 1);
    assert_eq!(count_table_files(&oldgen_dir), 1);

    start(&mut env, "new", &repo, settings, None);
    db_on(&mut env, "new").close();
    finish(env);
}
