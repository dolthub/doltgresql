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

#![allow(dead_code)]

use std::path::Path;

use driver::client::{Db, Value};
use driver::model::Server;
use driver::process::{Repo, Visitor};
use driver::runner::Env;
use harness::pgx::{Arg, Time};

/// SERVER is the key of the server that setup_test_server starts.
pub const SERVER: &str = "server";

/// env returns an environment for a test.
pub fn env() -> Env {
    Env::new(&Path::new(env!("CARGO_MANIFEST_DIR")).join("../../integration-tests/go-sql-server-driver"))
}

/// setup_test_server starts a server for a new database of the given name.
pub fn setup_test_server(env: &mut Env, repo_name: &str) -> Repo {
    start_server(env, repo_name, Server { dynamic_port: "server_port".into(), ..Server::default() }, None)
}

/// start_server creates the database in a new store and starts a server for it.
pub fn start_server(env: &mut Env, repo_name: &str, server: Server, visitor: Option<Visitor>) -> Repo {
    let repo = make_repo(env, repo_name);
    start(env, SERVER, &repo, server, visitor);
    repo
}

/// make_repo creates the database in a new store.
pub fn make_repo(env: &mut Env, repo_name: &str) -> Repo {
    let user = env.user().unwrap();
    let store = user.make_repo_store().unwrap();
    store.make_repo(repo_name).unwrap()
}

/// start starts a server under the key for the database's store, serving the database.
pub fn start(env: &mut Env, key: &str, repo: &Repo, server: Server, visitor: Option<Visitor>) {
    assert!(env.start_server(key, &repo.store, &server, visitor).unwrap());
    env.server(key).db_name = repo.name.clone();
}

/// write_file writes the file into the directory after applying the port templates.
pub fn write_file(env: &mut Env, dir: &Path, name: &str, contents: &str) {
    let contents = env.resources.apply_template(contents).unwrap();
    std::fs::write(dir.join(name), contents).unwrap();
}

/// db connects to the test's server as postgres.
pub fn db(env: &mut Env) -> Db {
    db_on(env, SERVER)
}

/// db_on connects to the server under the key as postgres.
pub fn db_on(env: &mut Env, key: &str) -> Db {
    env.server(key).db("postgres", "password", "", &[]).unwrap()
}

/// finish stops the servers and fails the test when stopping them or checking their logs failed.
pub fn finish(mut env: Env) {
    let problems = env.finish();
    assert!(problems.is_empty(), "{}", problems.join("\n"));
}

/// query_row returns the single row of a query.
pub fn query_row(db: &mut Db, sql: &str, args: &[Arg]) -> Vec<Value> {
    let (_, mut rows) = db.query_values(sql, args).unwrap_or_else(|e| panic!("{sql}: {e}"));
    assert_eq!(rows.len(), 1, "{sql}: expected one row");
    rows.remove(0)
}

/// query_int returns the single integer of a query.
pub fn query_int(db: &mut Db, sql: &str, args: &[Arg]) -> i64 {
    match query_row(db, sql, args).as_slice() {
        [Value::Int(v)] => *v,
        other => panic!("{sql}: expected an integer, got {other:?}"),
    }
}

/// exec runs a statement, failing the test on error.
pub fn exec(db: &mut Db, sql: &str, args: &[Arg]) {
    db.exec_args(sql, args).unwrap_or_else(|e| panic!("{sql}: {e}"));
}

/// now returns the current time in the local zone, like Go's time.Now.
pub fn now() -> Arg {
    let now = jiff::Zoned::now();
    Arg::Time(Time {
        year: now.year() as i32,
        month: now.month() as u32,
        day: now.day() as u32,
        hour: now.hour() as u32,
        minute: now.minute() as u32,
        second: now.second() as u32,
        nanosecond: now.subsec_nanosecond() as u32,
        offset_seconds: now.offset().seconds(),
    })
}

/// make_test_text returns a deterministic ASCII string of exactly size bytes, differing by seed.
pub fn make_test_text(seed: i64, size: usize) -> String {
    let chunk = format!("[row{seed:07}-filler]");
    chunk.repeat(size.div_ceil(chunk.len()))[..size].to_string()
}

/// make_test_blob_data returns deterministic bytes of exactly size bytes, differing by seed.
pub fn make_test_blob_data(seed: i64, size: usize) -> Vec<u8> {
    (0..size as i64).map(|i| ((seed * 37 + i * 17 + seed * i * 3) & 0xFF) as u8).collect()
}

/// make_test_json_string returns a JSON document of at least target_size bytes, built like the Go test's
/// largeJSONDoc with encoding/json.
pub fn make_test_json_string(seed: i64, target_size: usize) -> String {
    let description = format!("desc-seed{seed:07}-").repeat(10);
    let tags: Vec<String> = (0..15).map(|i| format!("\"tag-{seed}-{i}\"")).collect();
    let mut items = Vec::new();
    for i in 0.. {
        let payload = format!("pl-{seed}-{i}-").repeat(12);
        items.push(format!("{{\"index\":{i},\"name\":\"item-{seed}-{i:05}\",\"payload\":\"{payload}\"}}"));
        let doc = format!(
            "{{\"id\":{seed},\"description\":\"{description}\",\"tags\":[{}],\"items\":[{}]}}",
            tags.join(","),
            items.join(",")
        );
        if doc.len() >= target_size {
            return doc;
        }
    }
    unreachable!()
}

/// json_id returns the top-level id of a JSON document.
pub fn json_id(doc: &str) -> i64 {
    let value: serde_json::Value = serde_json::from_str(doc).unwrap();
    value["id"].as_i64().unwrap()
}
