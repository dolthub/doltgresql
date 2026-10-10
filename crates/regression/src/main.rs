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

use std::path::Path;

use harness::server::{Server, Target};
use regression::files::{FILES, QUERIES_TO_SKIP};
use regression::messages::read_messages;
use regression::replay::{Options, replay};
use regression::{report, tracker};

const USAGE: &str = "usage:
  regression replay <regression dir> <target> <out.trackers> [file...]
  regression dump <in.trackers> <out.txt>
  regression compare <main.trackers> <pr.trackers>
  regression parity <a.trackers> <b.trackers>

The target is doltgres:<binary> or postgres:<bin dir>:<template data dir>.";

/// main dispatches the subcommand.
fn main() {
    let args: Vec<String> = std::env::args().collect();
    let result = match args.get(1).map(String::as_str) {
        Some("replay") if args.len() >= 5 => run_replay(Path::new(&args[2]), &args[3], Path::new(&args[4]), &args[5..]),
        Some("dump") if args.len() == 4 => {
            read_trackers(&args[2]).and_then(|t| std::fs::write(&args[3], report::dump(&t)).map_err(|e| e.to_string()))
        }
        Some("compare") if args.len() == 4 => read_trackers(&args[2]).and_then(|from| {
            println!("{}", report::compare_markdown(&from, &read_trackers(&args[3])?));
            Ok(())
        }),
        Some("parity") if args.len() == 4 => read_trackers(&args[2]).and_then(|a| {
            print!("{}", report::parity(&a, &read_trackers(&args[3])?));
            Ok(())
        }),
        _ => Err(USAGE.to_string()),
    };
    if let Err(err) = result {
        eprintln!("{err}");
        std::process::exit(1);
    }
}

/// read_trackers reads a `.trackers` file.
fn read_trackers(path: &str) -> Result<Vec<tracker::Tracker>, String> {
    tracker::deserialize(&std::fs::read(path).map_err(|e| format!("{path}: {e}"))?)
}

/// run_replay replays every file, or only the named ones, in order against one server and writes the trackers.
fn run_replay(root: &Path, target: &str, out: &Path, only: &[String]) -> Result<(), String> {
    let root = &std::fs::canonicalize(root).map_err(|e| format!("{}: {e}", root.display()))?;
    let target = Target::parse(target)?;
    let server = Server::start(&target, "")?;
    let mut trackers = Vec::new();
    for file in FILES.iter().filter(|f| only.is_empty() || only.iter().any(|o| o == *f)) {
        let messages = read_messages(&root.join("results").join(format!("{file}.results")), &root.join("data"))?;
        let tracker = replay(Options {
            file,
            port: server.port,
            messages,
            print_queries: std::env::var_os("REGRESSION_PRINT").is_some(),
            fail_psql: true,
            fail_queries: &QUERIES_TO_SKIP,
            password: "password",
        })
        .map_err(|e| format!("{file}: {e}\nserver log: {}", server.log_tail()))?;
        trackers.push(tracker);
    }
    std::fs::write(out, tracker::serialize(&trackers)).map_err(|e| e.to_string())?;
    println!("Finished, wrote {}", out.display());
    Ok(())
}
