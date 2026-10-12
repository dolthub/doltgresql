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

//! Stores the results of a run, and reports a comparison of two runs in the regression comment's format.

use std::collections::HashMap;
use std::fmt::Write;

use serde_json::{Value, json};

/// How many regressions or progressions a comparison lists.
const LISTED: usize = 40;
/// How many lines of a unit's output a comparison shows.
const SHOWN_LINES: usize = 20;

/// FileResult is the result of one test.
#[derive(Clone, Debug, PartialEq)]
pub struct FileResult {
    pub name: String,
    pub units: Vec<UnitResult>,
}

/// UnitResult is the result of one unit of a test: whether its output matched, and when it did not, the expected
/// output and the output received, which is None when psql stopped before the unit.
#[derive(Clone, Debug, PartialEq)]
pub struct UnitResult {
    pub query: String,
    pub passed: bool,
    pub expected: Option<String>,
    pub received: Option<String>,
}

/// to_json returns the results as JSON.
pub fn to_json(files: &[FileResult]) -> String {
    let files: Vec<Value> = files
        .iter()
        .map(|file| {
            let units: Vec<Value> = file
                .units
                .iter()
                .map(|u| json!({"query": u.query, "passed": u.passed, "expected": u.expected, "received": u.received}))
                .collect();
            json!({"name": file.name, "units": units})
        })
        .collect();
    json!({ "files": files }).to_string()
}

/// from_json reads results written by to_json.
pub fn from_json(text: &str) -> Result<Vec<FileResult>, String> {
    let value: Value = serde_json::from_str(text).map_err(|e| e.to_string())?;
    let string = |v: &Value| v.as_str().map(str::to_string);
    let files = value["files"].as_array().ok_or("the results have no files")?;
    let file = |file: &Value| -> Option<FileResult> {
        let units = file["units"].as_array()?.iter().map(|u| {
            Some(UnitResult {
                query: string(&u["query"])?,
                passed: u["passed"].as_bool()?,
                expected: string(&u["expected"]),
                received: string(&u["received"]),
            })
        });
        Some(FileResult { name: string(&file["name"])?, units: units.collect::<Option<_>>()? })
    };
    files.iter().map(|f| file(f).ok_or_else(|| "malformed results".to_string())).collect()
}

/// summary lists each test's passing and total units, the tests with the most failures first.
pub fn summary(files: &[FileResult]) -> String {
    let (passed, total) = totals(files);
    let mut out = format!("TOTAL: {total}  SUCCESS: {passed} ({:.2}%)\n\n", percent(passed, total));
    let mut sorted: Vec<&FileResult> = files.iter().collect();
    sorted.sort_by_key(|f| std::cmp::Reverse(f.units.iter().filter(|u| !u.passed).count()));
    for file in sorted {
        let passed = file.units.iter().filter(|u| u.passed).count();
        let _ = writeln!(
            out,
            "{:<32} {passed:>5} / {:<5} ({:.1}%)",
            file.name,
            file.units.len(),
            percent(passed, file.units.len())
        );
    }
    out
}

/// compare_markdown reports the totals of a main run and a pull request run, and the units that only one of them
/// passes, matched by test, query, and occurrence.
pub fn compare_markdown(main: &[FileResult], pr: &[FileResult]) -> String {
    let (main_passed, main_total) = totals(main);
    let (pr_passed, pr_total) = totals(pr);
    let mut out = String::new();
    out.push_str("|   | Main | PR |\n| --- | --- | --- |\n");
    let _ = writeln!(out, "| Total | {main_total} | {pr_total} |");
    let _ = writeln!(out, "| Successful | {main_passed} | {pr_passed} |");
    let _ = writeln!(out, "| Failures | {} | {} |", main_total - main_passed, pr_total - pr_passed);
    out.push_str("\n|   | Main | PR |\n| --- | --- | --- |\n");
    let _ = writeln!(
        out,
        "| Successful | {:.4}% | {:.4}% |",
        percent(main_passed, main_total),
        percent(pr_passed, pr_total)
    );
    let _ = writeln!(
        out,
        "| Failures | {:.4}% | {:.4}% |",
        percent(main_total - main_passed, main_total),
        percent(pr_total - pr_passed, pr_total)
    );
    let main_passes = passes(main);
    let mut regressions = Vec::new();
    let mut progressions = Vec::new();
    for file in pr {
        let mut seen: HashMap<&str, usize> = HashMap::new();
        for unit in &file.units {
            let occurrence = seen.entry(unit.query.as_str()).or_default();
            *occurrence += 1;
            let Some(&passed_on_main) = main_passes.get(&(file.name.as_str(), unit.query.as_str(), *occurrence)) else {
                continue;
            };
            match (passed_on_main, unit.passed) {
                (true, false) => regressions.push((file.name.as_str(), unit)),
                (false, true) => progressions.push((file.name.as_str(), unit)),
                _ => {}
            }
        }
    }
    let _ = writeln!(out, "\n## ${{\\color{{red}}Regressions ({})}}$", regressions.len());
    let mut last = "";
    for (file, unit) in regressions.iter().take(LISTED) {
        if *file != last {
            let _ = writeln!(out, "### {file}");
            last = file;
        }
        let received = unit.received.as_deref().unwrap_or("(psql stopped before this statement)");
        let _ = writeln!(
            out,
            "```\nQUERY:\n{}\nEXPECTED:\n{}\nRECEIVED:\n{}\n```",
            unit.query,
            shown(unit.expected.as_deref().unwrap_or_default()),
            shown(received)
        );
    }
    let _ = writeln!(out, "\n## ${{\\color{{lightgreen}}Progressions ({})}}$", progressions.len());
    last = "";
    for (file, unit) in progressions.iter().take(LISTED) {
        if *file != last {
            let _ = writeln!(out, "### {file}");
            last = file;
        }
        let _ = writeln!(out, "```\nQUERY:\n{}\n```", unit.query);
    }
    out
}

/// passes maps each unit of a run, by test, query, and occurrence, to whether it passed.
fn passes(files: &[FileResult]) -> HashMap<(&str, &str, usize), bool> {
    let mut passes = HashMap::new();
    for file in files {
        let mut seen: HashMap<&str, usize> = HashMap::new();
        for unit in &file.units {
            let occurrence = seen.entry(unit.query.as_str()).or_default();
            *occurrence += 1;
            passes.insert((file.name.as_str(), unit.query.as_str(), *occurrence), unit.passed);
        }
    }
    passes
}

/// totals returns a run's passing units and all of its units.
fn totals(files: &[FileResult]) -> (usize, usize) {
    let units = files.iter().flat_map(|f| &f.units);
    (units.clone().filter(|u| u.passed).count(), units.count())
}

/// percent returns a share as a percentage, or 0 of nothing.
fn percent(n: usize, of: usize) -> f64 {
    if of == 0 { 0.0 } else { n as f64 / of as f64 * 100.0 }
}

/// shown returns the first lines of an output, noting how many more it has.
fn shown(output: &str) -> String {
    let lines: Vec<&str> = output.lines().collect();
    match lines.len() > SHOWN_LINES {
        true => format!("{}\n... ({} more lines)", lines[..SHOWN_LINES].join("\n"), lines.len() - SHOWN_LINES),
        false => output.to_string(),
    }
}
