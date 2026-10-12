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

use std::fmt::Write;

use sql::integrity::Stats;

use crate::repair::RepairSummary;

/// DatabaseReport is what the tool found in one database, and what it repaired.
pub struct DatabaseReport {
    pub name: String,
    pub branches: Vec<BranchReport>,
    /// What the repair rewrote, when the tool repaired the database.
    pub repair: Option<RepairSummary>,
    /// The scan of every branch after the repair, when the tool repaired the database.
    pub post_repair: Vec<BranchReport>,
}

/// BranchReport is what the tool found in every table at the head of one branch.
pub struct BranchReport {
    pub branch: String,
    pub tables: Vec<TableReport>,
}

/// TableReport is what the tool found in one table.
pub struct TableReport {
    pub schema: String,
    pub table: String,
    pub adaptive_value_columns: Vec<String>,
    pub adaptive_key_columns: Vec<String>,
    pub values_impacted: bool,
    pub keys_impacted: bool,
    /// The scan's counts, which are None for a table that holds no adaptive values and so was not scanned.
    pub stats: Option<Stats>,
    /// The error that scanning the table returned, which is empty when the scan succeeded.
    pub error: String,
}

/// HEAD is the report's page up to the first database.
const HEAD: &str = r#"<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Adaptive value_address_offsets corruption report</title>
<style>
body { font-family: sans-serif; margin: 2em; }
h1 { font-size: 1.4em; }
h2 { font-size: 1.2em; margin-top: 1.5em; }
h3 { font-size: 1.0em; margin-top: 1em; }
table { border-collapse: collapse; margin: 0.5em 0 1.5em 0; }
th, td { border: 1px solid #ccc; padding: 4px 10px; text-align: right; font-size: 0.9em; }
th { background: #f0f0f0; }
td.name { text-align: left; font-family: monospace; }
tr.corrupt td { background: #fff0f0; }
tr.clean td { background: #f4fff4; }
tr.skipped td { color: #999; }
.summary { background: #f8f8f8; border: 1px solid #ddd; padding: 0.75em 1em; margin: 1em 0; }
.warn { color: #a00; }
</style>
</head>
<body>
<h1>Adaptive value_address_offsets corruption report</h1>
"#;

/// COLUMNS is the header row of each branch's table.
const COLUMNS: &str = "<tr>
<th>Schema</th><th>Table</th><th>Adaptive columns</th>
<th>Rows</th><th>Corrupt rows</th><th>% rows</th>
<th>Adaptive values</th><th>Out-of-band</th><th>Corrupt values</th><th>% values</th>
<th>Chunks</th><th>Corrupt chunks</th>
<th>Key out-of-band</th><th>Corrupt key values</th><th>Missing chunks</th>
</tr>
";

/// html_report renders the report of every database, laid out as the Go tool's template lays it out.
pub fn html_report(generated: &str, mode: &str, databases: &[DatabaseReport]) -> String {
    let mut out = String::from(HEAD);
    let _ = writeln!(out, "<p>Generated {} · mode: {}</p>", escape(generated), escape(mode));
    for database in databases {
        let _ = writeln!(out, "\n<h2>Database: {}</h2>", escape(&database.name));
        if let Some(r) = &database.repair {
            let _ = writeln!(
                out,
                "\n<div class=\"summary\">\n<b>Repair summary:</b>\ncommits examined: {},\ncommits rewritten: {},\n\
                 branches updated: {},\ntags updated: {},\nworking sets repaired: {},\nleaf chunks rewritten: {},\n\
                 internal chunks rewritten: {}\n</div>",
                r.commits_examined,
                r.commits_rewritten,
                r.branches_updated,
                r.tags_updated,
                r.working_sets_fixed,
                r.leaf_chunks_rewritten,
                r.internal_chunks_rewritten
            );
        }
        out.push('\n');
        branches(&mut out, &database.branches);
        out.push('\n');
        if !database.post_repair.is_empty() {
            out.push_str("\n<h3>Post-repair verification scan</h3>\n");
            branches(&mut out, &database.post_repair);
            out.push('\n');
        }
        out.push('\n');
    }
    out.push_str("\n</body>\n</html>\n\n\n");
    out
}

/// branches renders the tables of each branch.
fn branches(out: &mut String, branches: &[BranchReport]) {
    out.push('\n');
    for branch in branches {
        let _ = write!(out, "\n<h3>Branch: {}</h3>\n<table>\n{COLUMNS}", escape(&branch.branch));
        for table in &branch.tables {
            out.push_str("\n\n");
            row(out, table);
            out.push_str("\n\n");
        }
        out.push_str("\n</table>\n");
    }
    out.push('\n');
}

/// row renders one table's row.
fn row(out: &mut String, table: &TableReport) {
    let (schema, name) = (escape(&table.schema), escape(&table.table));
    if !table.error.is_empty() {
        let _ = write!(
            out,
            "<tr class=\"corrupt\"><td class=\"name\">{schema}</td><td class=\"name\">{name}</td><td colspan=\"13\" \
             class=\"warn\">error: {}</td></tr>",
            escape(&table.error)
        );
        return;
    }
    let Some(s) = &table.stats else {
        let _ = write!(
            out,
            "<tr class=\"skipped\"><td class=\"name\">{schema}</td><td class=\"name\">{name}</td><td \
             class=\"name\">(schema not impacted)</td><td colspan=\"12\"></td></tr>"
        );
        return;
    };
    let class = if s.corrupt() > 0 { "corrupt" } else { "clean" };
    let mut columns = escape(&table.adaptive_value_columns.join(", "));
    if !table.adaptive_key_columns.is_empty() {
        let _ = write!(columns, " [key: {}]", escape(&table.adaptive_key_columns.join(", ")));
    }
    let internal = match s.internal_key_corrupt_values {
        0 => String::new(),
        n => format!(" (+{n} internal)"),
    };
    let missing = match s.missing_chunks {
        0 => "0".to_string(),
        n => format!("<span class=\"warn\">{n}</span>"),
    };
    let _ = write!(
        out,
        "<tr class=\"{class}\">\n<td class=\"name\">{schema}</td>\n<td class=\"name\">{name}</td>\n<td \
         class=\"name\">{columns}</td>\n<td>{}</td>\n<td>{}</td>\n<td>{}</td>\n<td>{}</td>\n<td>{}</td>\n<td>{}</td>\n\
         <td>{}</td>\n<td>{}</td>\n<td>{}</td>\n<td>{}</td>\n<td>{}{internal}</td>\n<td>{missing}</td>\n</tr>",
        s.rows,
        s.corrupt_rows,
        percent(s.corrupt_rows, s.rows),
        s.adaptive_values,
        s.out_of_band_values,
        s.corrupt_values,
        percent(s.corrupt_values, s.adaptive_values),
        s.chunks,
        s.corrupt_chunks,
        s.key_out_of_band_values,
        s.key_corrupt_values
    );
}

/// percent returns a part of a whole as a percentage with two decimals.
fn percent(part: u64, whole: u64) -> String {
    if whole == 0 {
        return "0.00%".to_string();
    }
    format!("{:.2}%", part as f64 / whole as f64 * 100.0)
}

/// escape escapes text for HTML as Go's html/template does.
fn escape(text: &str) -> String {
    let mut out = String::with_capacity(text.len());
    for c in text.chars() {
        match c {
            '&' => out.push_str("&amp;"),
            '<' => out.push_str("&lt;"),
            '>' => out.push_str("&gt;"),
            '"' => out.push_str("&#34;"),
            '\'' => out.push_str("&#39;"),
            '+' => out.push_str("&#43;"),
            '\0' => out.push('\u{FFFD}'),
            c => out.push(c),
        }
    }
    out
}
