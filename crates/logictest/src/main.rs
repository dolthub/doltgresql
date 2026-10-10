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

use std::collections::VecDeque;
use std::sync::Mutex;
use std::time::Duration;

use harness::server::{Server, Target};
use logictest::results;
use logictest::runner::{Harness, collect_test_files, run_test_file};

const USAGE: &str = "usage:
  logictest [-r=csv|json] [--target=<target> | --port=<port>] [--timeout=<seconds>] [--workers=<n>] run <path>...
  logictest [-r=csv|json] parse <version> <log file>

The target is doltgres:<binary> or postgres:<bin dir>:<template data dir>, and is started for the run. Without a
target, the run uses the server already listening on the port, 5432 by default.";

/// Flags holds the options that precede the command.
struct Flags {
    format: String,
    target: Option<String>,
    port: u16,
    timeout: Option<Duration>,
    workers: usize,
}

/// parse_flags reads `-name=value`, `--name=value`, and `-name value` options up to the command.
fn parse_flags(args: &mut VecDeque<String>) -> Result<Flags, String> {
    let mut flags = Flags { format: "json".to_string(), target: None, port: 5432, timeout: None, workers: 1 };
    while let Some(arg) = args.front().filter(|a| a.starts_with('-')).cloned() {
        args.pop_front();
        let arg = arg.trim_start_matches('-');
        let (name, value) = match arg.split_once('=') {
            Some((name, value)) => (name.to_string(), value.to_string()),
            None => (arg.to_string(), args.pop_front().ok_or_else(|| format!("missing value for -{arg}"))?),
        };
        let number = |value: &str| value.parse::<u64>().map_err(|e| format!("-{name}: {e}"));
        match name.as_str() {
            "r" => flags.format = value,
            "target" => flags.target = Some(value),
            "port" => flags.port = number(&value)? as u16,
            "timeout" => flags.timeout = Some(Duration::from_secs(number(&value)?)).filter(|d| !d.is_zero()),
            "workers" => flags.workers = number(&value)?.max(1) as usize,
            other => return Err(format!("unknown flag -{other}")),
        }
    }
    Ok(flags)
}

fn main() {
    let mut args: VecDeque<String> = std::env::args().skip(1).collect();
    let result = parse_flags(&mut args).and_then(|flags| match args.pop_front().as_deref() {
        Some("run") if !args.is_empty() => run(&flags, &Vec::from(args)),
        Some("parse") if args.len() == 2 => parse(&flags, &args[0], &args[1]),
        _ => Err(USAGE.to_string()),
    });
    if let Err(err) = result {
        eprintln!("{err}");
        std::process::exit(1);
    }
}

/// run runs the test files, sharing the files among the workers, each with its own database.
fn run(flags: &Flags, paths: &[String]) -> Result<(), String> {
    let server = flags.target.as_deref().map(|t| Target::parse(t).and_then(|t| Server::start(&t, ""))).transpose()?;
    let port = server.as_ref().map_or(flags.port, |s| s.port);
    let files = Mutex::new(collect_test_files(paths)?.into_iter().collect::<VecDeque<_>>());
    let output = Mutex::new(());
    std::thread::scope(|scope| {
        let workers: Vec<_> = (1..=flags.workers)
            .map(|n| {
                let (files, output) = (&files, &output);
                let database =
                    if flags.workers == 1 { "sqllogictest".to_string() } else { format!("sqllogictest_{n}") };
                scope.spawn(move || -> Result<(), String> {
                    let mut harness = Harness::new(port, &database, flags.timeout);
                    while let Some(file) = files.lock().unwrap().pop_front() {
                        run_test_file(&mut harness, &file, output)?;
                    }
                    Ok(())
                })
            })
            .collect();
        workers.into_iter().try_for_each(|w| w.join().map_err(|_| "worker panicked".to_string())?)
    })
}

/// parse prints a result log as JSON or CSV rows for importing into Dolt.
fn parse(flags: &Flags, version: &str, log: &str) -> Result<(), String> {
    let entries = results::parse_log(&std::fs::read_to_string(log).map_err(|e| format!("{log}: {e}"))?)?;
    match flags.format.as_str() {
        "csv" => print!("{}", results::to_csv(&entries, version)),
        _ => print!("{}", results::to_json(&entries, version)),
    }
    Ok(())
}
