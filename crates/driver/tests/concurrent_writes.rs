// Copyright 2024 Dolthub, Inc.
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

use std::sync::atomic::{AtomicU32, Ordering};
use std::sync::{Arc, Barrier};

use common::*;
use driver::client::Db;
use harness::pgx::Arg;

#[test]
fn test_concurrent_writes() {
    const NUM_WRITERS: i64 = 32;
    let mut env = env();
    setup_test_server(&mut env, "concurrent_writes_test");
    let mut db = db(&mut env);
    exec(&mut db, "CREATE TABLE data (id VARCHAR(64) PRIMARY KEY, worker INT, data TEXT, created_at TIMESTAMP)", &[]);
    exec(&mut db, "SELECT DOLT_COMMIT('-Am', 'init with table')", &[]);
    db.close();
    let next_int = AtomicU32::new(0);
    let mut writers: Vec<Db> = (0..NUM_WRITERS).map(|_| common::db(&mut env)).collect();
    let start = Arc::new(Barrier::new(NUM_WRITERS as usize));
    let results: Vec<Result<(), String>> = std::thread::scope(|scope| {
        let handles: Vec<_> = writers
            .iter_mut()
            .enumerate()
            .map(|(i, conn)| {
                let (start, next_int) = (start.clone(), &next_int);
                scope.spawn(move || -> Result<(), String> {
                    start.wait();
                    for j in 0..16 {
                        let key = format!("main-{i}-{j}");
                        let args = [Arg::Str(key.clone()), Arg::Int(i as i64), Arg::Str(key.clone()), now()];
                        conn.exec_args("INSERT INTO data VALUES ($1,$2,$3,$4)", &args)?;
                        next_int.fetch_add(1, Ordering::SeqCst);
                        match conn.exec(&format!("SELECT DOLT_COMMIT('-Am', 'insert {key}')"), &[]) {
                            Err(e) if !e.contains("nothing to commit") => return Err(e),
                            _ => {}
                        }
                    }
                    Ok(())
                })
            })
            .collect();
        handles.into_iter().map(|h| h.join().unwrap()).collect()
    });
    for result in results {
        result.unwrap();
    }
    let next_int = next_int.load(Ordering::SeqCst);
    assert_eq!(next_int, 512);
    println!("wrote {next_int}");
    let mut conn = common::db(&mut env);
    assert_eq!(query_int(&mut conn, "SELECT COUNT(*) FROM data", &[]), next_int as i64);
    println!("ended with {} commits", query_int(&mut conn, "SELECT COUNT(*) FROM dolt_log", &[]));
    drop(writers);
    finish(env);
}
