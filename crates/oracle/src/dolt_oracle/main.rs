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

//! Checks Doltgres against Dolt's own behavior, as the sessions that Doltlite's oracle tests run against Dolt record
//! it:
//!
//! `dolt-oracle generate <record.jsonl> <doltgres> <corpus dir>` turns the sessions that record.py recorded into a
//! corpus of Postgres statements, one file per suite, whose expectations are what Doltgres returns where that matches
//! what Dolt returned, and what Dolt returned, marked as a divergence, where it does not.
//!
//! `dolt-oracle run <corpus dir> <doltgres> [suite...]` replays the corpus and reports each statement whose result
//! differs from its expectation, exiting non-zero when one that is not a divergence does.

mod corpus;
mod mysql;

use std::process::ExitCode;

/// main runs a subcommand.
fn main() -> ExitCode {
    let args: Vec<String> = std::env::args().skip(1).collect();
    let result = match args.first().map(String::as_str) {
        Some("generate") if args.len() == 4 => corpus::generate(&args[1], &args[2], &args[3]),
        Some("run") if args.len() >= 3 => corpus::run(&args[1], &args[2], &args[3..]),
        _ => Err("usage: dolt-oracle generate <record.jsonl> <doltgres> <corpus dir> | run <corpus dir> <doltgres> [suite...]".into()),
    };
    match result {
        Ok(true) => ExitCode::SUCCESS,
        Ok(false) => ExitCode::from(1),
        Err(err) => {
            eprintln!("{err}");
            ExitCode::from(2)
        }
    }
}
