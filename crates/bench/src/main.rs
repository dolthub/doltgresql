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

#![forbid(unsafe_code)]

//! Times every query of `queries.sql` against each server named on the command line and prints the median latency
//! of each, with how the `rust` server compares to the `go` server when both are named.
//!
//! Usage: `bench [--time <ms>] [--filter <substring>] [--save <file>] [--load <file>]... <label>=<target>...`, where a
//! target is `doltgres:<binary>` or `postgres:<bin dir>:<template data dir>`, as `DOLTGRES_TEST_TARGET` takes. Saved
//! results can be loaded again instead of running a server, so that baselines need to run only once.

use std::collections::BTreeMap;
use std::time::{Duration, Instant};

use harness::pgx::{Conn, ConnConfig};
use harness::server::{Server, Target};
use md5::{Digest, Md5};

/// Query is one timed query of the corpus.
struct Query {
    name: String,
    sql: String,
    extended: bool,
    write: bool,
}

/// Timing is what one server did with one query: its median and fastest latency in microseconds and a digest of its
/// rows, or the error it failed with.
#[derive(Clone)]
struct Timing {
    median: f64,
    fastest: f64,
    digest: String,
    error: Option<String>,
}

/// Results are each label's timings by query name.
type Results = BTreeMap<String, BTreeMap<String, Timing>>;

fn main() {
    if let Err(err) = run() {
        eprintln!("bench: {err}");
        std::process::exit(1);
    }
}

/// run parses the command line, times the queries, and prints the comparison.
fn run() -> Result<(), String> {
    let mut budget = Duration::from_millis(400);
    let (mut filter, mut save, mut labels) = (None, None, Vec::new());
    let mut hold = None;
    let mut show = false;
    let mut results = Results::new();
    let mut args = std::env::args().skip(1);
    while let Some(arg) = args.next() {
        let mut value = || args.next().ok_or_else(|| format!("{arg} needs a value"));
        match arg.as_str() {
            "--time" => budget = Duration::from_millis(value()?.parse().map_err(|_| "invalid --time")?),
            "--filter" => filter = Some(value()?),
            "--show" => show = true,
            "--hold" => hold = Some(Duration::from_secs(value()?.parse().map_err(|_| "invalid --hold")?)),
            "--save" => save = Some(value()?),
            "--load" => {
                for (label, timings) in load(&value()?)? {
                    labels.retain(|l: &(String, Option<Target>)| l.0 != label);
                    labels.push((label.clone(), None));
                    results.insert(label, timings);
                }
            }
            _ => {
                let (label, target) = arg.split_once('=').ok_or_else(|| format!("invalid argument {arg}"))?;
                labels.push((label.to_string(), Some(Target::parse(target)?)));
            }
        }
    }
    let corpus = include_str!("../queries.sql");
    let (setup, queries) = parse(corpus);
    let queries: Vec<Query> =
        queries.into_iter().filter(|q| filter.as_ref().is_none_or(|f| q.name.contains(f.as_str()))).collect();
    for (label, target) in &labels {
        let Some(target) = target else { continue };
        if let Some(hold) = hold {
            return repeat(target, &setup, &queries, hold);
        }
        if show {
            print_rows(label, target, &setup, &queries)?;
            continue;
        }
        eprintln!("{label}: starting");
        results.insert(label.clone(), time_target(target, &setup, &queries, budget)?);
    }
    if let Some(path) = save {
        store(&path, &results)?;
    }
    print(&labels.iter().map(|l| l.0.clone()).collect::<Vec<_>>(), &queries, &results);
    Ok(())
}

/// parse splits the corpus into its setup statements and its queries.
fn parse(corpus: &str) -> (Vec<String>, Vec<Query>) {
    let (mut setup, mut queries) = (Vec::new(), Vec::<Query>::new());
    let mut in_setup = false;
    for line in corpus.lines() {
        let trimmed = line.trim();
        if trimmed == "-- setup" {
            in_setup = true;
        } else if let Some(name) = trimmed.strip_prefix("-- name: ") {
            in_setup = false;
            queries.push(Query { name: name.to_string(), sql: String::new(), extended: false, write: false });
        } else if trimmed == "-- extended" {
            if let Some(query) = queries.last_mut() {
                query.extended = true;
            }
        } else if trimmed == "-- write" {
            if let Some(query) = queries.last_mut() {
                query.write = true;
            }
        } else if trimmed.is_empty() || trimmed.starts_with("--") {
            continue;
        } else if in_setup {
            setup.push(trimmed.to_string());
        } else if let Some(query) = queries.last_mut() {
            if !query.sql.is_empty() {
                query.sql.push(' ');
            }
            query.sql.push_str(trimmed);
        }
    }
    (setup, queries)
}

/// time_target starts a fresh server, builds the data, and times every query on it.
fn time_target(
    target: &Target,
    setup: &[String],
    queries: &[Query],
    budget: Duration,
) -> Result<BTreeMap<String, Timing>, String> {
    let server = Server::start(target, "")?;
    let url = format!("postgres://postgres:password@127.0.0.1:{}/postgres", server.port);
    let mut conn = Conn::connect(ConnConfig::parse(&url).map_err(|e| e.to_string())?).map_err(|e| e.to_string())?;
    let started = Instant::now();
    for statement in setup {
        if let Some(err) = conn.simple_query(statement).map_err(|e| e.to_string())?.error {
            return Err(format!("setup failed: {statement}: {err}"));
        }
    }
    let mut timings = BTreeMap::new();
    let micros = started.elapsed().as_secs_f64() * 1e6;
    timings.insert("setup".to_string(), Timing { median: micros, fastest: micros, digest: String::new(), error: None });
    for query in queries {
        timings.insert(query.name.clone(), time_query(&mut conn, query, budget));
    }
    Ok(timings)
}

/// repeat builds the data on a fresh server, says so on stderr, and then runs the queries over and over for the given
/// time, for a profiler to watch the server.
fn repeat(target: &Target, setup: &[String], queries: &[Query], hold: Duration) -> Result<(), String> {
    let server = Server::start(target, "")?;
    let url = format!("postgres://postgres:password@127.0.0.1:{}/postgres", server.port);
    let mut conn = Conn::connect(ConnConfig::parse(&url).map_err(|e| e.to_string())?).map_err(|e| e.to_string())?;
    for statement in setup {
        if let Some(err) = conn.simple_query(statement).map_err(|e| e.to_string())?.error {
            return Err(format!("setup failed: {statement}: {err}"));
        }
    }
    eprintln!("ready");
    let (started, mut runs) = (Instant::now(), 0);
    while started.elapsed() < hold {
        for query in queries {
            match query.extended {
                true => drop(conn.query(&query.sql, &[]).map_err(|e| e.to_string())?),
                false => drop(conn.simple_query(&query.sql).map_err(|e| e.to_string())?),
            }
            runs += 1;
        }
    }
    eprintln!("{runs} runs, {:.0} us each", started.elapsed().as_secs_f64() * 1e6 / runs as f64);
    Ok(())
}

/// print_rows builds the data on a fresh server and prints the rows of each query in text, sorted.
fn print_rows(label: &str, target: &Target, setup: &[String], queries: &[Query]) -> Result<(), String> {
    let server = Server::start(target, "")?;
    let url = format!("postgres://postgres:password@127.0.0.1:{}/postgres", server.port);
    let mut conn = Conn::connect(ConnConfig::parse(&url).map_err(|e| e.to_string())?).map_err(|e| e.to_string())?;
    for statement in setup {
        if let Some(err) = conn.simple_query(statement).map_err(|e| e.to_string())?.error {
            return Err(format!("setup failed: {statement}: {err}"));
        }
    }
    for query in queries {
        println!("== {label} {}", query.name);
        let result = match query.extended {
            true => conn.query(&query.sql, &[]),
            false => conn.simple_query_rows(&query.sql),
        };
        match result {
            Ok(result) => {
                let mut rows: Vec<String> = result
                    .rows
                    .iter()
                    .map(|row| {
                        row.iter()
                            .map(|v| v.as_ref().map_or("NULL".into(), |v| String::from_utf8_lossy(v).into_owned()))
                            .collect::<Vec<_>>()
                            .join("|")
                    })
                    .collect();
                rows.sort();
                for row in rows {
                    println!("{row}");
                }
            }
            Err(err) => println!("error: {err}"),
        }
    }
    Ok(())
}

/// time_query runs a query once for its rows, then again until the budget is spent, and returns its timing.
fn time_query(conn: &mut Conn, query: &Query, budget: Duration) -> Timing {
    let failed = |error: String| Timing { median: 0.0, fastest: 0.0, digest: String::new(), error: Some(error) };
    let digest = match first_run(conn, query) {
        Ok(digest) => digest,
        Err(err) => return failed(err),
    };
    let mut samples = Vec::new();
    let started = Instant::now();
    while samples.len() < 5 || (started.elapsed() < budget && samples.len() < 2000) {
        let run = Instant::now();
        let outcome = match query.extended {
            true => conn.query(&query.sql, &[]).map(|r| r.error),
            false => conn.simple_query(&query.sql).map(|r| r.error),
        };
        match outcome {
            Ok(None) => samples.push(run.elapsed().as_secs_f64() * 1e6),
            Ok(Some(err)) => return failed(err.to_string()),
            Err(err) => return failed(err.to_string()),
        }
    }
    samples.sort_by(f64::total_cmp);
    Timing { median: samples[samples.len() / 2], fastest: samples[0], digest, error: None }
}

/// first_run runs a query and returns a digest of its rows in text, sorted so that servers that return rows of equal
/// order keys in different orders still agree.
fn first_run(conn: &mut Conn, query: &Query) -> Result<String, String> {
    let result = match query.extended {
        true => conn.query(&query.sql, &[]),
        false => conn.simple_query_rows(&query.sql),
    }
    .map_err(|e| e.to_string())?;
    if let Some(err) = result.error {
        return Err(err.to_string());
    }
    if query.write {
        return Ok(String::new());
    }
    let mut rows: Vec<String> = result
        .rows
        .iter()
        .map(|row| {
            row.iter()
                .map(|v| v.as_ref().map_or("NULL".to_string(), |v| String::from_utf8_lossy(v).into_owned()))
                .collect::<Vec<_>>()
                .join("|")
        })
        .collect();
    rows.sort();
    let digest = Md5::digest(rows.join("\n").as_bytes());
    Ok(format!("{}:{:x}", rows.len(), digest))
}

/// store writes results as tab-separated lines of label, query, median, fastest, digest, and error.
fn store(path: &str, results: &Results) -> Result<(), String> {
    let mut out = String::new();
    for (label, timings) in results {
        for (name, t) in timings {
            let error = t.error.as_deref().unwrap_or("").replace(['\t', '\n'], " ");
            out.push_str(&format!("{label}\t{name}\t{}\t{}\t{}\t{error}\n", t.median, t.fastest, t.digest));
        }
    }
    std::fs::write(path, out).map_err(|e| format!("cannot write {path}: {e}"))
}

/// load reads results that `store` wrote.
fn load(path: &str) -> Result<Results, String> {
    let text = std::fs::read_to_string(path).map_err(|e| format!("cannot read {path}: {e}"))?;
    let mut results = Results::new();
    for line in text.lines() {
        let fields: Vec<&str> = line.split('\t').collect();
        let [label, name, median, fastest, digest, error] = fields[..] else {
            return Err(format!("invalid line in {path}: {line}"));
        };
        let timing = Timing {
            median: median.parse().unwrap_or(0.0),
            fastest: fastest.parse().unwrap_or(0.0),
            digest: digest.to_string(),
            error: (!error.is_empty()).then(|| error.to_string()),
        };
        results.entry(label.to_string()).or_default().insert(name.to_string(), timing);
    }
    Ok(results)
}

/// print prints each query's median latency per label, the `rust`/`go` ratio, and a summary of the ratios.
fn print(labels: &[String], queries: &[Query], results: &Results) {
    let mut header = format!("{:<26}", "query");
    for label in labels {
        header.push_str(&format!("{label:>12}"));
    }
    let compare = labels.iter().any(|l| l == "rust") && labels.iter().any(|l| l == "go");
    if compare {
        header.push_str("  rust/go");
    }
    println!("{header}");
    let (mut log_sum, mut count, mut slower) = (0.0, 0, Vec::new());
    let names = std::iter::once("setup").chain(queries.iter().map(|q| q.name.as_str()));
    for name in names {
        let mut line = format!("{name:<26}");
        let mut digests = Vec::new();
        for label in labels {
            match results.get(label).and_then(|t| t.get(name)) {
                Some(Timing { error: Some(_), .. }) => line.push_str(&format!("{:>12}", "error")),
                Some(t) => {
                    line.push_str(&format!("{:>12.0}", t.median));
                    if !t.digest.is_empty() {
                        digests.push(t.digest.clone());
                    }
                }
                None => line.push_str(&format!("{:>12}", "-")),
            }
        }
        let timing = |label: &str| results.get(label).and_then(|t| t.get(name)).filter(|t| t.error.is_none());
        if compare && let (Some(rust), Some(go)) = (timing("rust"), timing("go")) {
            let ratio = rust.median / go.median;
            line.push_str(&format!("  {ratio:7.2}"));
            if name != "setup" {
                log_sum += ratio.ln();
                count += 1;
                if ratio > 1.0 {
                    slower.push(name.to_string());
                }
            }
        }
        digests.dedup();
        if digests.len() > 1 {
            line.push_str("  (results differ)");
        }
        println!("{line}");
        for label in labels {
            if let Some(Timing { error: Some(err), .. }) = results.get(label).and_then(|t| t.get(name)) {
                println!("    {label}: {}", err.chars().take(150).collect::<String>());
            }
        }
    }
    if count > 0 {
        println!(
            "geometric mean rust/go: {:.2}; slower than go: {} of {count}",
            (log_sum / count as f64).exp(),
            slower.len()
        );
        if !slower.is_empty() {
            println!("slower: {}", slower.join(", "));
        }
    }
}
