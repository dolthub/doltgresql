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

use std::sync::Arc;
use std::sync::atomic::{AtomicBool, AtomicI32, Ordering};

use common::*;
use driver::client::Db;
use driver::model::Server;
use driver::runner::Env;
use rand::Rng;

/// AutoGcTest is a server with auto GC configured, and what its log has shown.
struct AutoGcTest {
    env: Env,
    db: Db,
    gc_count: Arc<AtomicI32>,
    saw_dangling_ref: Arc<AtomicBool>,
}

/// setup starts a server with auto GC configured and creates the vals table, which only covers Dolt's base
/// single-server variant since Doltgres has no cluster replication or remotes API.
fn setup(enable: bool, archive: bool) -> AutoGcTest {
    let mut env = env();
    let gc_count = Arc::new(AtomicI32::new(0));
    let saw_dangling_ref = Arc::new(AtomicBool::new(false));
    let repo = make_repo(&mut env, "auto_gc_test");
    let archive_fragment = if archive { "\n    archive_level: 1" } else { "" };
    let behavior = format!(
        "\nbehavior:\n  auto_gc_behavior:\n    enable: {enable}{archive_fragment}\nlistener:\n  port: {{{{get_port \"server_port\"}}}}\n"
    );
    write_file(&mut env, &repo.store.dir, "server.yaml", &behavior);
    let (count, dangling) = (gc_count.clone(), saw_dangling_ref.clone());
    let visitor: driver::process::Visitor = Arc::new(move |line: &str| {
        if line.contains("Successfully completed auto GC") {
            count.fetch_add(1, Ordering::SeqCst);
        }
        if line.contains("dangling references requested during GC") {
            dangling.store(true, Ordering::SeqCst);
        }
    });
    let server = Server {
        name: "primary".into(),
        args: vec!["--config".into(), "server.yaml".into()],
        dynamic_port: "server_port".into(),
        envs: vec!["DOLT_GC_SCHEDULER=NONE".into()],
        ..Server::default()
    };
    start(&mut env, SERVER, &repo, server, Some(visitor));
    let mut db = db(&mut env);
    exec(
        &mut db,
        "\ncreate table vals (\n    id bigint primary key,\n    v1 bigint,\n    v2 bigint,\n    v3 bigint,\n    v4 bigint\n)\n",
        &[],
    );
    let indexes: [&[&str]; 16] = [
        &["v1"],
        &["v2"],
        &["v3"],
        &["v4"],
        &["v1", "v2"],
        &["v1", "v3"],
        &["v1", "v4"],
        &["v2", "v3"],
        &["v2", "v4"],
        &["v2", "v1"],
        &["v3", "v1"],
        &["v3", "v2"],
        &["v3", "v4"],
        &["v4", "v1"],
        &["v4", "v2"],
        &["v4", "v3"],
    ];
    for cols in indexes {
        exec(&mut db, &format!("create index vals_{}_idx on vals ({})", cols.join("_"), cols.join(",")), &[]);
    }
    exec(&mut db, "select dolt_commit('-Am', 'create vals table')", &[]);
    AutoGcTest { env, db, gc_count, saw_dangling_ref }
}

/// insert_statement returns an insert of 1024 rows of random values, starting from the statement's first id.
fn insert_statement(i: usize) -> String {
    let mut rng = rand::thread_rng();
    let vals: Vec<String> = (i * 1024..(i + 1) * 1024)
        .map(|j| {
            let vs: Vec<String> = (0..4).map(|_| rng.gen_range(0..=i64::MAX).to_string()).collect();
            format!("({j},{})", vs.join(","))
        })
        .collect();
    format!("insert into vals values {}", vals.join(","))
}

/// run_statement inserts a batch over the pooled connection, committing after every commit_every batches.
fn run_statement(test: &mut AutoGcTest, i: usize, commit_every: usize) {
    exec(&mut test.db, &insert_statement(i), &[]);
    if (i + 1).is_multiple_of(commit_every) {
        exec(&mut test.db, &format!("select dolt_commit('-am', 'insert from {}')", i * 1024), &[]);
    }
}

/// run_until_gc inserts until auto GC has completed target_gc_count times, failing if it does not within 1024
/// statements or if the server logs dangling references.
fn run_until_gc(archive: bool, target_gc_count: i32, commit_every: usize) {
    const MAX_STATEMENTS: usize = 1024;
    let mut test = setup(true, archive);
    for i in 0..MAX_STATEMENTS {
        assert!(!test.saw_dangling_ref.load(Ordering::SeqCst), "saw dangling references message during auto GC");
        if test.gc_count.load(Ordering::SeqCst) >= target_gc_count {
            println!("reached {target_gc_count} auto GCs after {i} statements");
            break;
        }
        run_statement(&mut test, i, commit_every);
    }
    assert!(!test.saw_dangling_ref.load(Ordering::SeqCst), "saw dangling references message during auto GC");
    let count = test.gc_count.load(Ordering::SeqCst);
    assert!(count >= target_gc_count, "did not reach {target_gc_count} auto GCs within {MAX_STATEMENTS} statements");
    println!("auto GC count: {count}");
    finish(test.env);
}

#[test]
fn test_auto_gc_enable_archive() {
    run_until_gc(true, 3, 16);
}

#[test]
fn test_auto_gc_enable_no_archive() {
    run_until_gc(false, 3, 16);
}

#[test]
fn test_auto_gc_disabled() {
    let mut test = setup(false, false);
    for i in 0..64 {
        run_statement(&mut test, i, 16);
    }
    assert_eq!(test.gc_count.load(Ordering::SeqCst), 0, "auto GC should not run when disabled");
    finish(test.env);
}
