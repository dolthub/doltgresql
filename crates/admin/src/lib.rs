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

//! The offline administration tool for Doltgres databases. It finds and repairs the corruption that releases before
//! 0.56.3 wrote, where nodes holding out-of-band adaptive values (TEXT, JSON, and so on) failed to record their chunk
//! addresses, so that push, clone, and garbage collection could lose the values, and that every earlier release wrote
//! for adaptive primary key columns.
//!
//! `report` scans the head of every branch of every database in `-dir` and writes an HTML report. `repair` does the
//! same, then rewrites every corrupt node in every commit that a branch or tag reaches, moves the refs to the repaired
//! history, and adds a scan after the repair to the report. Commits that need no repair keep their hashes. Repair must
//! run while no server is using the databases.

pub mod repair;
pub mod report;
pub mod rewrite;

use std::path::{Path, PathBuf};

use doltdb::database::Database;
use doltdb::root::Root;
use sql::integrity::{Scanner, tables_for_root};
use sql::txn::read;

use crate::repair::Repairer;
use crate::report::{BranchReport, DatabaseReport, TableReport};

/// USAGE is the error for missing or unknown modes.
const USAGE: &str = "usage: admin <report|repair> [-dir <path>] [-out <report.html>] [-verbose]";

/// Options are the flags of a run.
struct Options {
    dir: String,
    out: String,
    verbose: bool,
}

/// Exit is how a run ends early: with an error to print, or with the status that Go's flag package exits with.
pub enum Exit {
    Error(String),
    Status(i32),
}

impl From<String> for Exit {
    fn from(message: String) -> Exit {
        Exit::Error(message)
    }
}

impl From<sql::PgError> for Exit {
    fn from(err: sql::PgError) -> Exit {
        Exit::Error(err.message)
    }
}

/// run runs the tool with its arguments, leaving out the program name.
pub fn run(args: &[String]) -> Result<(), Exit> {
    let mode = match args.first().map(String::as_str) {
        Some(mode @ ("report" | "repair")) => mode,
        _ => return Err(Exit::Error(USAGE.to_string())),
    };
    let options = parse_flags(mode, &args[1..])?;
    let mut reports = Vec::new();
    for (name, path) in database_dirs(Path::new(&options.dir))? {
        let mut db = Database::open(&path.join(".dolt").join("noms"))
            .map_err(|e| format!("failed to load database {name}: {e}"))?;
        let report = process_database(&mut db, &name, mode, options.verbose)
            .map_err(|e| format!("database {name}: {}", e.message))?;
        db.close().map_err(|e| format!("database {name}: {e}"))?;
        reports.push(report);
    }
    let generated = generated_time();
    std::fs::write(&options.out, report::html_report(&generated, mode, &reports))
        .map_err(|e| format!("open {}: {e}", options.out))?;
    print!("{}", summary(&options.out, &reports));
    Ok(())
}

/// parse_flags reads the flags as Go's flag package does, printing the usage and returning the status to exit with
/// when they are wrong or ask for help.
fn parse_flags(mode: &str, args: &[String]) -> Result<Options, Exit> {
    let mut options = Options { dir: ".".into(), out: "adaptive-corruption-report.html".into(), verbose: false };
    let mut i = 0;
    while i < args.len() {
        let arg = &args[i];
        if arg == "--" {
            break;
        }
        let Some(flag) = arg.strip_prefix("--").or_else(|| arg.strip_prefix('-')).filter(|f| !f.is_empty()) else {
            break;
        };
        let (name, value) = match flag.split_once('=') {
            Some((name, value)) => (name, Some(value.to_string())),
            None => (flag, None),
        };
        match name {
            "dir" | "out" => {
                let value = match value {
                    Some(value) => value,
                    None => {
                        i += 1;
                        match args.get(i) {
                            Some(value) => value.clone(),
                            None => return Err(flag_error(mode, &format!("flag needs an argument: -{name}"))),
                        }
                    }
                };
                if name == "dir" { options.dir = value } else { options.out = value }
            }
            "verbose" => {
                options.verbose = match value.as_deref() {
                    None | Some("1" | "t" | "T" | "true" | "TRUE" | "True") => true,
                    Some("0" | "f" | "F" | "false" | "FALSE" | "False") => false,
                    Some(other) => {
                        let message = format!("invalid boolean value \"{other}\" for -verbose: parse error");
                        return Err(flag_error(mode, &message));
                    }
                }
            }
            "h" | "help" => {
                eprint!("{}", flag_usage(mode));
                return Err(Exit::Status(0));
            }
            _ => return Err(flag_error(mode, &format!("flag provided but not defined: -{name}"))),
        }
        i += 1;
    }
    Ok(options)
}

/// flag_error prints a flag error and the usage, returning Go's exit status for them.
fn flag_error(mode: &str, message: &str) -> Exit {
    eprintln!("{message}");
    eprint!("{}", flag_usage(mode));
    Exit::Status(2)
}

/// flag_usage returns the usage that Go's flag package prints for the mode's flags.
fn flag_usage(mode: &str) -> String {
    format!(
        "Usage of {mode}:\n  -dir string\n    \tdatabase directory, or a data directory containing databases (default \
         \".\")\n  -out string\n    \tpath of the HTML report to write (default \
         \"adaptive-corruption-report.html\")\n  -verbose\n    \tlog progress to stderr\n"
    )
}

/// database_dirs returns the databases in a directory, which is a database itself when it holds a Dolt database, and
/// otherwise holds databases in its subdirectories, each named after its directory.
fn database_dirs(dir: &Path) -> Result<Vec<(String, PathBuf)>, String> {
    let name = |path: &Path| {
        let absolute = std::path::absolute(path).unwrap_or_else(|_| path.to_path_buf());
        absolute.file_name().map(|n| n.to_string_lossy().replace(['-', ' '], "_")).unwrap_or_default()
    };
    let mut databases = Vec::new();
    let holds_database = |p: &Path| p.join(".dolt").join("noms").is_dir();
    if holds_database(dir) {
        databases.push((name(dir), dir.to_path_buf()));
    }
    let mut entries: Vec<PathBuf> = std::fs::read_dir(dir)
        .map_err(|e| format!("open {}: {e}", dir.display()))?
        .filter_map(|e| e.ok().map(|e| e.path()))
        .filter(|p| holds_database(p))
        .collect();
    entries.sort();
    databases.extend(entries.into_iter().map(|p| (name(&p), p)));
    if databases.is_empty() {
        return Err(format!("no databases found in {}", dir.display()));
    }
    Ok(databases)
}

/// process_database scans one database and, in repair mode, repairs it and scans it again.
pub fn process_database(db: &mut Database, name: &str, mode: &str, verbose: bool) -> sql::Result<DatabaseReport> {
    let mut scanner = Scanner::new(db);
    if verbose {
        log(&format!("scanning database {name}"));
    }
    let branches = scan_branch_heads(&mut scanner, verbose)?;
    let mut report = DatabaseReport { name: name.to_string(), branches, repair: None, post_repair: Vec::new() };
    if mode == "repair" {
        if verbose {
            log(&format!("repairing database {name}"));
        }
        let mut repairer = Repairer::new(scanner, verbose);
        report.repair = Some(repairer.repair_database()?);
        if verbose {
            log(&format!("verifying database {name} post-repair"));
        }
        scanner = repairer.rewriter.scanner;
        report.post_repair = scan_branch_heads(&mut scanner, verbose)?;
    }
    if verbose {
        log(&format!("database {name}: {} subtree scans satisfied from cache", scanner.cache_hits));
    }
    Ok(report)
}

/// scan_branch_heads scans every table at the head of every branch.
pub fn scan_branch_heads(scanner: &mut Scanner<'_>, verbose: bool) -> sql::Result<Vec<BranchReport>> {
    let mut reports = Vec::new();
    for (dataset, head) in scanner.db.datasets()? {
        let Some(branch) = dataset.strip_prefix("refs/heads/") else { continue };
        let commit = sql::dolt::history::load(scanner.db, head)?;
        let root = Root::decode(&read(scanner.db, &commit.root)?)?;
        let mut report = BranchReport { branch: branch.to_string(), tables: Vec::new() };
        for table in tables_for_root(scanner.db, &root)? {
            let mut row = TableReport {
                schema: table.schema.clone(),
                table: table.name.clone(),
                adaptive_value_columns: table.adaptive_value_columns.clone(),
                adaptive_key_columns: table.adaptive_key_columns.clone(),
                values_impacted: table.values_impacted(),
                keys_impacted: table.keys_impacted(),
                stats: None,
                error: String::new(),
            };
            if row.values_impacted || row.keys_impacted {
                match scanner.scan_table(&table) {
                    Ok(stats) => {
                        if verbose {
                            log(&format!(
                                "branch {branch} table {}: {}/{} rows corrupt, {}/{} adaptive values corrupt",
                                table.shown(),
                                stats.corrupt_rows,
                                stats.rows,
                                stats.corrupt_values,
                                stats.adaptive_values
                            ));
                        }
                        row.stats = Some(stats);
                    }
                    Err(err) => row.error = err.message,
                }
            }
            report.tables.push(row);
        }
        reports.push(report);
    }
    Ok(reports)
}

/// summary returns the short text summary of the results.
fn summary(out: &str, reports: &[DatabaseReport]) -> String {
    let mut text = String::new();
    for report in reports {
        let (mut rows, mut corrupt_rows, mut values, mut corrupt_values, mut missing) = (0, 0, 0, 0, 0);
        for stats in report.branches.iter().flat_map(|b| &b.tables).filter_map(|t| t.stats.as_ref()) {
            rows += stats.rows;
            corrupt_rows += stats.corrupt_rows;
            values += stats.adaptive_values + stats.key_adaptive_values;
            corrupt_values += stats.corrupt();
            missing += stats.missing_chunks;
        }
        let name = &report.name;
        text.push_str(&format!(
            "database {name}: {corrupt_rows} corrupt rows (of {rows} scanned), {corrupt_values} corrupt adaptive \
             values (of {values} scanned) across all branch heads\n"
        ));
        if missing > 0 {
            text.push_str(&format!(
                "database {name}: WARNING: {missing} out-of-band chunks are already missing from the chunk store; the \
                 values referencing them are unrecoverable\n"
            ));
        }
        if let Some(r) = &report.repair {
            text.push_str(&format!(
                "database {name}: repaired {} leaf chunks, rewrote {} of {} commits, updated {} branches, {} tags, {} \
                 working sets\n",
                r.leaf_chunks_rewritten,
                r.commits_rewritten,
                r.commits_examined,
                r.branches_updated,
                r.tags_updated,
                r.working_sets_fixed
            ));
        }
    }
    text.push_str(&format!("full report written to {out}\n"));
    text
}

/// local_now returns the current time in the machine's time zone.
fn local_now() -> chrono::DateTime<chrono_tz::Tz> {
    let zone = sql::settings::local_timezone().parse().unwrap_or(chrono_tz::UTC);
    chrono::Utc::now().with_timezone(&zone)
}

/// generated_time returns the current time as Go's time.RFC1123 formats it.
fn generated_time() -> String {
    local_now().format("%a, %d %b %Y %H:%M:%S %Z").to_string()
}

/// log prints a progress message to stderr with the time, as Go's log package does.
pub fn log(message: &str) {
    eprintln!("{} {message}", local_now().format("%Y/%m/%d %H:%M:%S"));
}
