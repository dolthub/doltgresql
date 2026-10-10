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

use harness::server::Target;
use regress::report;
use regress::run::{Options, run};

const USAGE: &str = "usage:
  regress run <suite dir> <target> <psql> <out dir> [test...]
  regress summary <results.json>
  regress compare <main results.json> <pr results.json>

The suite directory is Postgres' src/test/regress. The target is doltgres:<binary> or
postgres:<bin dir>:<template data dir>. A run writes each test's output and results.json to the out directory.";

/// main dispatches the subcommand.
fn main() {
    let args: Vec<String> = std::env::args().collect();
    let result = match args.get(1).map(String::as_str) {
        Some("run") if args.len() >= 6 => run_suite(&args[2], &args[3], &args[4], &args[5], &args[6..]),
        Some("summary") if args.len() == 3 => read(&args[2]).map(|files| print!("{}", report::summary(&files))),
        Some("compare") if args.len() == 4 => read(&args[2]).and_then(|main| {
            print!("{}", report::compare_markdown(&main, &read(&args[3])?));
            Ok(())
        }),
        _ => Err(USAGE.to_string()),
    };
    if let Err(err) = result {
        eprintln!("{err}");
        std::process::exit(1);
    }
}

/// read reads a results file.
fn read(path: &str) -> Result<Vec<report::FileResult>, String> {
    report::from_json(&std::fs::read_to_string(path).map_err(|e| format!("{path}: {e}"))?)
}

/// run_suite runs the suite's tests, or only the named ones, and writes the results.
fn run_suite(suite: &str, target: &str, psql: &str, out: &str, only: &[String]) -> Result<(), String> {
    let absolute = |path: &str| std::fs::canonicalize(path).map_err(|e| format!("{path}: {e}"));
    std::fs::create_dir_all(out).map_err(|e| format!("{out}: {e}"))?;
    let (suite, psql, out) = (absolute(suite)?, absolute(psql)?, absolute(out)?);
    let target = Target::parse(target)?;
    let files = run(&Options { suite: &suite, target: &target, psql: &psql, out: &out, only })?;
    let path = out.join("results.json");
    std::fs::write(&path, report::to_json(&files)).map_err(|e| e.to_string())?;
    print!("{}", report::summary(&files));
    println!("Finished, wrote {}", path.display());
    Ok(())
}
