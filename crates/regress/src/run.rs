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

//! Runs the tests of a suite's schedule one after another against a server through psql, as pg_regress runs them
//! against an installed server, and compares each unit's output with the expected output.

use std::fs::File;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Stdio};
use std::time::{Duration, Instant};

use harness::server::{Server, Target};

use crate::report::{FileResult, UnitResult};
use crate::{align, script};

/// The database that the tests run in, as pg_regress names it.
const DATABASE: &str = "regression";
/// How long one test may run before it is stopped.
const TEST_TIMEOUT: Duration = Duration::from_secs(300);
/// How long the server may take to answer a query between tests.
const PING_TIMEOUT: Duration = Duration::from_secs(30);

/// Options configures a run.
pub struct Options<'a> {
    /// The suite's directory, holding parallel_schedule and the sql, expected, and data directories.
    pub suite: &'a Path,
    pub target: &'a Target,
    /// The psql binary, which should be the suite's version.
    pub psql: &'a Path,
    /// The directory that receives each test's output and the files that tests write.
    pub out: &'a Path,
    /// The tests to run, or every test of the schedule when empty.
    pub only: &'a [String],
}

/// run runs the tests and returns their results. A test that stops the server, or runs too long, is followed by a
/// restart of the server, which keeps its data.
pub fn run(options: &Options<'_>) -> Result<Vec<FileResult>, String> {
    let tests = schedule(&options.suite.join("parallel_schedule"))?;
    std::fs::create_dir_all(options.out.join("results")).map_err(|e| e.to_string())?;
    let mut server = Server::start(options.target, "")?;
    create_database(options, &server);
    let mut files = Vec::new();
    for test in tests.iter().filter(|t| options.only.is_empty() || options.only.contains(t)) {
        let started = Instant::now();
        let (output, finished) = run_test(options, &server, test)?;
        let file = compare(options.suite, test, &output)?;
        let passed = file.units.iter().filter(|u| u.passed).count();
        println!("{test}: {passed}/{} in {:.1}s", file.units.len(), started.elapsed().as_secs_f64());
        if !finished || !server.is_running() || !psql_succeeds(options, &server, "SELECT 1", PING_TIMEOUT) {
            println!("{test}: restarting the server\n{}", server.log_tail());
            server.restart().map_err(|e| format!("{test}: {e}"))?;
        }
        files.push(file);
    }
    Ok(files)
}

/// schedule returns the tests of a schedule file in order.
fn schedule(path: &Path) -> Result<Vec<String>, String> {
    let text = std::fs::read_to_string(path).map_err(|e| format!("{}: {e}", path.display()))?;
    let tests = text.lines().filter_map(|line| line.strip_prefix("test:")).flat_map(str::split_whitespace);
    Ok(tests.map(str::to_string).collect())
}

/// create_database creates the database that the tests run in, with the settings that pg_regress gives it. The
/// failures of a server that lacks a setting are printed and otherwise ignored.
fn create_database(options: &Options<'_>, server: &Server) {
    let database = format!("\"{DATABASE}\"");
    let created =
        [format!("DROP DATABASE IF EXISTS {database}"), format!("CREATE DATABASE {database} TEMPLATE=template0")]
            .iter()
            .all(|statement| psql_succeeds(options, server, statement, PING_TIMEOUT));
    if !created {
        psql_succeeds(options, server, &format!("CREATE DATABASE {database}"), PING_TIMEOUT);
    }
    for setting in [
        "lc_messages TO 'C'",
        "lc_monetary TO 'C'",
        "lc_numeric TO 'C'",
        "lc_time TO 'C'",
        "bytea_output TO 'hex'",
        "timezone_abbreviations TO 'Default'",
    ] {
        psql_succeeds(options, server, &format!("ALTER DATABASE {database} SET {setting}"), PING_TIMEOUT);
    }
}

/// psql returns a psql command that connects to the server with the environment that pg_regress gives its tests.
fn psql(options: &Options<'_>, server: &Server, test: &str) -> Command {
    let mut command = Command::new(options.psql);
    command
        .env("PGHOST", "127.0.0.1")
        .env("PGPORT", server.port.to_string())
        .env("PGUSER", "postgres")
        .env("PGPASSWORD", "password")
        .env("PGAPPNAME", format!("pg_regress/{test}"))
        .env("PG_ABS_SRCDIR", options.suite)
        .env("PG_ABS_BUILDDIR", options.out)
        .env("PGTZ", "America/Los_Angeles")
        .env("PGDATESTYLE", "Postgres, MDY")
        .env("PGOPTIONS", "-c intervalstyle=postgres_verbose")
        .env_remove("PGDATABASE");
    if std::env::var_os("PG_LIBDIR").is_none() {
        command.env("PG_LIBDIR", options.suite).env("PG_DLSUFFIX", ".so");
    }
    command
}

/// psql_succeeds runs one statement in the tests' database, or in the postgres database while that does not exist
/// yet, and reports whether it succeeded in time, printing its output when it failed.
fn psql_succeeds(options: &Options<'_>, server: &Server, statement: &str, timeout: Duration) -> bool {
    let database = match statement.contains(DATABASE) {
        true => "postgres",
        false => DATABASE,
    };
    let path = options.out.join("psql.log");
    let Ok(log) = File::create(&path) else { return false };
    let Ok(stderr) = log.try_clone() else { return false };
    let child = psql(options, server, "setup")
        .args(["-X", "-q", "-v", "ON_ERROR_STOP=1", "-d", database, "-c", statement])
        .stdin(Stdio::null())
        .stdout(log)
        .stderr(stderr)
        .spawn();
    let succeeded = child.is_ok_and(|mut child| wait(&mut child, timeout).is_some_and(|code| code == Some(0)));
    if !succeeded {
        println!("{statement}: {}", std::fs::read_to_string(&path).unwrap_or_default().trim());
    }
    succeeded
}

/// run_test runs a test's script through psql as pg_regress does, and returns its output and whether it finished in
/// time.
fn run_test(options: &Options<'_>, server: &Server, test: &str) -> Result<(String, bool), String> {
    let script = options.suite.join("sql").join(format!("{test}.sql"));
    let path = options.out.join("results").join(format!("{test}.out"));
    let output = File::create(&path).map_err(|e| format!("{}: {e}", path.display()))?;
    let mut child = psql(options, server, test)
        .args(["-X", "-a", "-q", "-d", DATABASE, "-v", "HIDE_TABLEAM=on", "-v", "HIDE_TOAST_COMPRESSION=on"])
        .stdin(File::open(&script).map_err(|e| format!("{}: {e}", script.display()))?)
        .stdout(output.try_clone().map_err(|e| e.to_string())?)
        .stderr(output)
        .spawn()
        .map_err(|e| format!("cannot start {}: {e}", options.psql.display()))?;
    let finished = wait(&mut child, TEST_TIMEOUT).is_some();
    let output = std::fs::read(&path).map_err(|e| format!("{}: {e}", path.display()))?;
    Ok((String::from_utf8_lossy(&output).into_owned(), finished))
}

/// wait waits for a process to exit and returns its exit code, or stops it and returns None when it runs too long.
fn wait(child: &mut Child, timeout: Duration) -> Option<Option<i32>> {
    let deadline = Instant::now() + timeout;
    loop {
        if let Ok(Some(status)) = child.try_wait() {
            return Some(status.code());
        }
        if Instant::now() > deadline {
            let _ = child.kill();
            let _ = child.wait();
            return None;
        }
        std::thread::sleep(Duration::from_millis(20));
    }
}

/// compare compares a test's output with each of its expected outputs, keeping the one that the most units match,
/// as pg_regress accepts any of a test's alternative outputs. A unit matches when its output is the expected output,
/// or when both lack one, as the units after a `\quit` do.
fn compare(suite: &Path, test: &str, output: &str) -> Result<FileResult, String> {
    let read = |path: &Path| {
        let bytes = std::fs::read(path).map_err(|e| format!("{}: {e}", path.display()))?;
        Ok::<_, String>(String::from_utf8_lossy(&bytes).into_owned())
    };
    let units = script::units(&read(&suite.join("sql").join(format!("{test}.sql")))?);
    let received = align::split(output, &units);
    let mut best: Option<(usize, Vec<Option<String>>)> = None;
    for path in expected_files(suite, test) {
        let expected = align::split(&read(&path)?, &units);
        let matched = received.iter().zip(&expected).filter(|(r, e)| r == e).count();
        if best.as_ref().is_none_or(|(most, _)| matched > *most) {
            best = Some((matched, expected));
        }
    }
    let Some((_, expected)) = best else { return Err(format!("{test}: no expected output")) };
    let units = units.iter().zip(received).zip(expected).map(|((unit, received), expected)| {
        let passed = received == expected;
        UnitResult {
            query: unit.query(),
            passed,
            expected: if passed { None } else { expected },
            received: if passed { None } else { received },
        }
    });
    Ok(FileResult { name: test.to_string(), units: units.collect() })
}

/// expected_files returns the paths of a test's expected output and its alternatives that exist.
fn expected_files(suite: &Path, test: &str) -> Vec<PathBuf> {
    let expected = suite.join("expected");
    let alternatives = (1..10).map(|n| expected.join(format!("{test}_{n}.out")));
    std::iter::once(expected.join(format!("{test}.out"))).chain(alternatives).filter(|p| p.is_file()).collect()
}
