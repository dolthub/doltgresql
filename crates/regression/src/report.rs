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

use std::collections::{BTreeMap, HashSet};
use std::fmt::Write;

use crate::tracker::Tracker;

/// dump matches the Go tool's TestDumpTrackers report.
pub fn dump(trackers: &[Tracker]) -> String {
    let mut out = String::new();
    let success: u32 = trackers.iter().map(|t| t.success).sum();
    let failed: u32 = trackers.iter().map(|t| t.failed).sum();
    let partial: u32 = trackers.iter().map(|t| t.partial_success).sum();
    let total = success + failed;
    let percent = |n: u32, of: u32| n as f64 / of as f64 * 100.0;
    let _ = writeln!(
        out,
        "TOTAL: {total}  SUCCESS: {success} ({:.2}%)  FAIL: {failed} ({:.2}%)  PARTIAL: {partial}\n",
        percent(success, total),
        percent(failed, total)
    );
    let mut stats: Vec<&Tracker> = trackers.iter().collect();
    stats.sort_by_key(|t| std::cmp::Reverse(t.failed));
    out.push_str("PER-FILE (sorted by failures):\n");
    for t in stats {
        let file_total = t.success + t.failed;
        if file_total == 0 {
            continue;
        }
        let _ = writeln!(
            out,
            "{:<40} total={:<5} success={:<5} fail={:<5} ({:.1}% pass)",
            t.file,
            file_total,
            t.success,
            t.failed,
            percent(t.success, file_total)
        );
    }
    out.push_str("\n==================== FAILURE DETAILS ====================\n");
    for t in trackers {
        if t.fail_partial_items.is_empty() {
            continue;
        }
        let _ =
            writeln!(out, "\n########## FILE: {} (fail={} partial={}) ##########", t.file, t.failed, t.partial_success);
        for item in &t.fail_partial_items {
            let _ = writeln!(out, "---\nQUERY: {}", item.query);
            if !item.expected_error.is_empty() {
                let _ = writeln!(out, "EXPECTED ERROR: {}", item.expected_error);
            }
            if !item.unexpected_error.is_empty() {
                let _ = writeln!(out, "RECEIVED ERROR: {}", item.unexpected_error);
            }
            for partial in &item.partial_success {
                let _ = writeln!(out, "PARTIAL: {partial}");
            }
        }
    }
    out
}

/// Totals holds a run's counts.
struct Totals {
    total: u32,
    success: u32,
    partial: u32,
    failed: u32,
}

fn totals(trackers: &[Tracker]) -> Totals {
    let success = trackers.iter().map(|t| t.success).sum();
    let failed = trackers.iter().map(|t| t.failed).sum();
    Totals { total: success + failed, success, partial: trackers.iter().map(|t| t.partial_success).sum(), failed }
}

/// compare_markdown matches the Go tool's comparison of a base run and a pull request run.
pub fn compare_markdown(from: &[Tracker], to: &[Tracker]) -> String {
    let f = totals(from);
    let t = totals(to);
    let mut out = String::new();
    out.push_str("|   | Main | PR |\n| --- | --- | --- |\n");
    let _ = writeln!(out, "| Total | {} | {} |", f.total, t.total);
    let _ = writeln!(out, "| Successful | {} | {} |", f.success, t.success);
    let _ = writeln!(out, "| Failures | {} | {} |", f.failed, t.failed);
    let _ = writeln!(out, "| Partial Successes[^1] | {} | {} |", f.partial, t.partial);
    out.push_str("\n|   | Main | PR |\n| --- | --- | --- |\n");
    let percent = |n: u32, of: u32| n as f64 / of as f64 * 100.0;
    let _ = writeln!(out, "| Successful | {:.4}% | {:.4}% |", percent(f.success, f.total), percent(t.success, t.total));
    let _ = writeln!(out, "| Failures | {:.4}% | {:.4}% |", percent(f.failed, f.total), percent(t.failed, t.total));
    let mut regressions = 0;
    let mut progressions = 0;
    if from.len() == to.len() {
        let mut found_any = false;
        for (a, b) in from.iter().zip(to) {
            if a.file != b.file {
                continue;
            }
            let mut found_file = false;
            let failed: HashSet<&str> = a.fail_partial_items.iter().map(|i| i.query.as_str()).collect();
            for item in &b.fail_partial_items {
                if failed.contains(item.query.as_str()) {
                    continue;
                }
                if regressions < 40 {
                    if !found_any {
                        found_any = true;
                        out.push_str("\n## ${\\color{red}Regressions__&&&&&&}$\n");
                    }
                    if !found_file {
                        found_file = true;
                        let _ = writeln!(out, "### {}", a.file);
                    }
                    let _ = writeln!(out, "```\nQUERY:          {}", item.query);
                    if !item.expected_error.is_empty() {
                        let _ = writeln!(out, "EXPECTED ERROR: {}", item.expected_error);
                    }
                    if !item.unexpected_error.is_empty() {
                        let _ = writeln!(out, "RECEIVED ERROR: {}", item.unexpected_error);
                    }
                    for partial in &item.partial_success {
                        let _ = writeln!(out, "PARTIAL:        {partial}");
                    }
                    out.push_str("```\n");
                }
                regressions += 1;
            }
        }
        let mut found_any = false;
        for (a, b) in from.iter().zip(to) {
            if a.file != b.file {
                continue;
            }
            let mut found_file = false;
            let succeeded: HashSet<&str> = a.success_items.iter().map(|i| i.query.as_str()).collect();
            for item in &b.success_items {
                if succeeded.contains(item.query.as_str()) {
                    continue;
                }
                if progressions < 40 {
                    if !found_any {
                        found_any = true;
                        out.push_str("\n## ${\\color{lightgreen}Progressions__&&&&&&}$\n");
                    }
                    if !found_file {
                        found_file = true;
                        let _ = writeln!(out, "### {}", a.file);
                    }
                    let _ = writeln!(out, "```\nQUERY: {}\n```", item.query);
                }
                progressions += 1;
            }
        }
    }
    out.push_str(
        "[^1]: These are tests that we're marking as `Successful`, however they do not match the expected output in \
         some way. This is due to small differences, such as different wording on the error messages, or the column \
         names being incorrect while the data itself is correct.",
    );
    out.replace("Regressions__&&&&&&", &format!("Regressions ({regressions})"))
        .replace("Progressions__&&&&&&", &format!("Progressions ({progressions})"))
}

/// parity lists, per file, the statements that succeed in one run but not the other.
pub fn parity(a: &[Tracker], b: &[Tracker]) -> String {
    fn counts(t: &Tracker) -> BTreeMap<&str, i64> {
        let mut counts = BTreeMap::new();
        for item in &t.success_items {
            *counts.entry(item.query.as_str()).or_default() += 1;
        }
        counts
    }
    let mut out = String::new();
    let mut only_a = 0;
    let mut only_b = 0;
    let b_files: BTreeMap<&str, &Tracker> = b.iter().map(|t| (t.file.as_str(), t)).collect();
    for ta in a {
        let Some(tb) = b_files.get(ta.file.as_str()) else {
            let _ = writeln!(out, "{}: missing from the second run", ta.file);
            continue;
        };
        if ta.success == tb.success && ta.failed == tb.failed && counts(ta) == counts(tb) {
            continue;
        }
        let _ = writeln!(
            out,
            "\n## {}: success {} vs {}, failed {} vs {}",
            ta.file, ta.success, tb.success, ta.failed, tb.failed
        );
        let mut difference = counts(ta);
        for (query, count) in counts(tb) {
            *difference.entry(query).or_default() -= count;
        }
        for (query, count) in difference {
            if count == 0 {
                continue;
            }
            let (side, n) = if count > 0 { ("first", count) } else { ("second", -count) };
            if count > 0 {
                only_a += n;
            } else {
                only_b += n;
            }
            let reason = |t: &Tracker| {
                t.fail_partial_items.iter().find(|i| i.query == query).map(|i| {
                    let text = if i.unexpected_error.is_empty() { &i.expected_error } else { &i.unexpected_error };
                    text.chars().take(300).collect::<String>()
                })
            };
            let other = if count > 0 { reason(tb) } else { reason(ta) };
            let _ = writeln!(
                out,
                "- only the {side} run succeeds{}: {}\n    other run: {}",
                if n > 1 { format!(" ({n} times)") } else { String::new() },
                query.chars().take(300).collect::<String>().replace('\n', " "),
                other.unwrap_or_default().replace('\n', " | ")
            );
        }
    }
    format!(
        "statements that only the first run passes: {only_a}\nstatements that only the second run passes: {only_b}\n{out}"
    )
}
