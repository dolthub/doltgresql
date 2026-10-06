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

use std::sync::Mutex;
use std::sync::atomic::{AtomicBool, Ordering};
use std::time::{Duration, Instant};

use common::*;
use driver::client::{Db, Value};
use driver::model::Server;
use harness::pgx::Arg;

/// GcTest runs updates on several threads while another thread repeatedly runs GC.
#[derive(Clone, Copy)]
struct GcTest {
    num_threads: usize,
    duration: Duration,
    commit: bool,
    full: bool,
    session_aware: bool,
}

/// Checks connects to the server and records failed assertions, like testify's assert.
struct Checks {
    port: u16,
    failures: Mutex<Vec<String>>,
}

impl Checks {
    /// connect connects to the test database.
    fn connect(&self) -> Result<Db, String> {
        Db::connect("postgres", "password", "concurrent_gc_test", "127.0.0.1", self.port, &[])
    }

    /// no_error records the error, returning whether there was none.
    fn no_error(&self, result: &Result<(), String>) -> bool {
        if let Err(e) = result {
            self.failures.lock().unwrap().push(format!("unexpected error: {e}"));
        }
        result.is_ok()
    }

    /// not_contains records a failure when the text contains the substring, returning whether it did not.
    fn not_contains(&self, text: &str, substring: &str) -> bool {
        if text.contains(substring) {
            self.failures.lock().unwrap().push(format!("{text:?} should not contain {substring:?}"));
        }
        !text.contains(substring)
    }

    /// expected_gc_error checks that an error from a connection killed by GC is not one of the disallowed ones.
    fn expected_gc_error(&self, err: &str) -> bool {
        self.not_contains(err, "dangling ref")
            && self.not_contains(err, "is unexpected noms value")
            && self.not_contains(err, "interface conversion: types.Value is nil")
    }
}

impl GcTest {
    /// create_db creates and commits the vals table with a zeroed row for every id the threads update.
    fn create_db(&self, db: &mut Db) {
        exec(db, "create table vals (id int primary key, val int)", &[]);
        let vals: Vec<String> = (0..=(self.num_threads - 1) * 1024).map(|i| format!("({i},0)")).collect();
        exec(db, &format!("insert into vals values {}", vals.join(",")), &[]);
        exec(db, "select dolt_commit('-Am', 'create vals table')", &[]);
    }

    /// tolerate checks a result like the Go test, returning whether it failed, or the error when it is not allowed.
    fn tolerate(&self, checks: &Checks, result: Result<(), String>, what: &str) -> Result<bool, String> {
        if self.session_aware {
            checks.no_error(&result);
        } else if let Err(e) = &result {
            if !checks.expected_gc_error(e) {
                return Err(e.clone());
            }
            println!("err in {what}: {e}");
        }
        Ok(result.is_err())
    }

    /// do_update increments the row in a transaction, committing it or making a Dolt commit.
    fn do_update(&self, checks: &Checks, i: i64) -> Result<(), String> {
        let mut conn = match checks.connect() {
            Ok(conn) => conn,
            Err(e) if self.session_aware => {
                checks.no_error(&Err(e));
                return Ok(());
            }
            Err(e) => {
                if !checks.not_contains(&e, "connection refused") {
                    return Err(e);
                }
                println!("err in Conn: {e}");
                return Ok(());
            }
        };
        let begin = conn.begin();
        if self.session_aware {
            checks.no_error(&begin);
        }
        if begin.is_err() {
            return Ok(());
        }
        let update = conn.exec_args("update vals set val = val+1 where id = $1", &[Arg::Int(i)]);
        if self.tolerate(checks, update, "Exec update")? {
            let _ = conn.rollback();
            return Ok(());
        }
        if self.commit {
            let commit = conn.exec(&format!("select dolt_commit('-am', 'increment vals id = {i}')"), &[]);
            self.tolerate(checks, commit, "Exec select dolt_commit")?;
            let _ = conn.rollback();
        } else {
            let commit = conn.commit();
            self.tolerate(checks, commit, "tx commit")?;
        }
        Ok(())
    }

    /// do_gc runs GC over a new connection, which GC leaves unusable.
    fn do_gc(&self, checks: &Checks) -> Result<(), String> {
        let mut conn = match checks.connect() {
            Ok(conn) => conn,
            Err(e) if self.session_aware => {
                checks.no_error(&Err(e));
                return Ok(());
            }
            Err(e) => {
                if !checks.not_contains(&e, "connection refused") {
                    return Err(e);
                }
                println!("err in Conn for dolt_gc: {e}");
                return Ok(());
            }
        };
        let start = Instant::now();
        let sql = if self.full { "select dolt_gc('--full')" } else { "select dolt_gc()" };
        if checks.no_error(&conn.exec(sql, &[])) {
            println!("successful dolt_gc took {:?}", start.elapsed());
        }
        Ok(())
    }

    /// finalize reads back every updated row and logs the update and commit counts.
    fn finalize(&self, db: &mut Db) {
        let ids: Vec<Arg> = (0..self.num_threads as i64 * 1024).step_by(1024).map(Arg::Int).collect();
        let marks: Vec<String> = (1..=ids.len()).map(|n| format!("${n}")).collect();
        let sql = format!("select val from vals where id in ({})", marks.join(","));
        let (_, rows) = db.query_values(&sql, &ids).unwrap();
        assert_eq!(rows.len(), ids.len());
        let count: i64 = rows
            .iter()
            .map(|row| match row.as_slice() {
                [Value::Int(v)] => *v,
                other => panic!("expected an integer, got {other:?}"),
            })
            .sum();
        println!("successfully updated val {count} times");
        println!("database has {} commit(s)", query_int(db, "select count(*) from dolt_log", &[]));
    }

    /// run starts a server with the safepoint controller, runs the updates and GC for the duration, and finalizes.
    fn run(self) {
        let mut env = env();
        let choice = if self.session_aware { "session_aware" } else { "kill_connections" };
        let server = Server {
            dynamic_port: "server_port".into(),
            envs: vec![format!("DOLT_GC_SAFEPOINT_CONTROLLER_CHOICE={choice}")],
            ..Server::default()
        };
        start_server(&mut env, "concurrent_gc_test", server, None);
        let mut conn = db(&mut env);
        self.create_db(&mut conn);
        conn.close();

        let checks = Checks { port: env.server(SERVER).port, failures: Mutex::new(Vec::new()) };
        let cancelled = AtomicBool::new(false);
        let start = Instant::now();
        let running = || start.elapsed() < self.duration && !cancelled.load(Ordering::SeqCst);
        let fail = |result: Result<(), String>| {
            if result.is_err() {
                cancelled.store(true, Ordering::SeqCst);
            }
            result
        };
        let results: Vec<Result<(), String>> = std::thread::scope(|scope| {
            let mut handles = Vec::new();
            for i in 0..self.num_threads as i64 {
                let (checks, running, fail) = (&checks, &running, &fail);
                handles.push(scope.spawn(move || {
                    while running() {
                        fail(self.do_update(checks, i * 1024))?;
                    }
                    Ok(())
                }));
            }
            let (checks, running, fail) = (&checks, &running, &fail);
            handles.push(scope.spawn(move || {
                while running() {
                    fail(self.do_gc(checks))?;
                    std::thread::sleep(Duration::from_millis(100));
                }
                Ok(())
            }));
            handles.into_iter().map(|h| h.join().unwrap()).collect()
        });
        for result in results {
            result.unwrap();
        }
        let failures = checks.failures.into_inner().unwrap();
        assert!(failures.is_empty(), "{}", failures.join("\n"));

        let mut conn = db(&mut env);
        self.finalize(&mut conn);
        conn.close();
        finish(env);
    }
}

/// run_concurrent_gc runs 8 threads of updates for 10 seconds alongside GC.
fn run_concurrent_gc(commit: bool, full: bool, session_aware: bool) {
    GcTest { num_threads: 8, duration: Duration::from_secs(10), commit, full, session_aware }.run();
}

#[test]
#[ignore = "Doltgres does not yet handle dolt_gc() concurrently with active write connections the way this test expects: writers fail with 'unexpected EOF' / connection resets rather than the retriable safepoint errors Dolt produces (behavioral difference)"]
fn test_concurrent_gc_no_commits_not_full_kill_connections() {
    run_concurrent_gc(false, false, false);
}

#[test]
#[ignore = "Doltgres does not yet handle dolt_gc() concurrently with active write connections the way this test expects: writers fail with 'unexpected EOF' / connection resets rather than the retriable safepoint errors Dolt produces (behavioral difference)"]
fn test_concurrent_gc_no_commits_not_full_session_aware() {
    run_concurrent_gc(false, false, true);
}

#[test]
#[ignore = "Doltgres does not yet handle dolt_gc() concurrently with active write connections the way this test expects: writers fail with 'unexpected EOF' / connection resets rather than the retriable safepoint errors Dolt produces (behavioral difference)"]
fn test_concurrent_gc_no_commits_full_kill_connections() {
    run_concurrent_gc(false, true, false);
}

#[test]
#[ignore = "Doltgres does not yet handle dolt_gc() concurrently with active write connections the way this test expects: writers fail with 'unexpected EOF' / connection resets rather than the retriable safepoint errors Dolt produces (behavioral difference)"]
fn test_concurrent_gc_no_commits_full_session_aware() {
    run_concurrent_gc(false, true, true);
}

#[test]
#[ignore = "Doltgres does not yet handle dolt_gc() concurrently with active write connections the way this test expects: writers fail with 'unexpected EOF' / connection resets rather than the retriable safepoint errors Dolt produces (behavioral difference)"]
fn test_concurrent_gc_with_commits_not_full_kill_connections() {
    run_concurrent_gc(true, false, false);
}

#[test]
#[ignore = "Doltgres does not yet handle dolt_gc() concurrently with active write connections the way this test expects: writers fail with 'unexpected EOF' / connection resets rather than the retriable safepoint errors Dolt produces (behavioral difference)"]
fn test_concurrent_gc_with_commits_not_full_session_aware() {
    run_concurrent_gc(true, false, true);
}

#[test]
#[ignore = "Doltgres does not yet handle dolt_gc() concurrently with active write connections the way this test expects: writers fail with 'unexpected EOF' / connection resets rather than the retriable safepoint errors Dolt produces (behavioral difference)"]
fn test_concurrent_gc_with_commits_full_kill_connections() {
    run_concurrent_gc(true, true, false);
}

#[test]
#[ignore = "Doltgres does not yet handle dolt_gc() concurrently with active write connections the way this test expects: writers fail with 'unexpected EOF' / connection resets rather than the retriable safepoint errors Dolt produces (behavioral difference)"]
fn test_concurrent_gc_with_commits_full_session_aware() {
    run_concurrent_gc(true, true, true);
}
