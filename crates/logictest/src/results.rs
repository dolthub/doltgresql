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

/// Entry is one line of a result log.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Entry {
    pub test_file: String,
    pub line_num: usize,
    pub query: String,
    pub duration_ms: u64,
    pub result: &'static str,
    pub error_message: String,
}

/// is_rfc3339 reports whether text looks like a time.RFC3339Nano timestamp.
fn is_rfc3339(text: &str) -> bool {
    let b = text.as_bytes();
    b.len() >= 20 && b[4] == b'-' && b[7] == b'-' && b[10] == b'T' && b[13] == b':' && b[16] == b':'
}

/// parse_line parses a result log line, returning None for lines that are not results.
pub fn parse_line(line: &str) -> Result<Option<Entry>, String> {
    let Some(first_space) = line.find(' ') else {
        return Ok(None);
    };
    if !is_rfc3339(&line[..first_space]) {
        return Ok(None);
    }
    let Some(second_space) = line[first_space + 1..].find(' ').map(|i| i + first_space) else {
        return Ok(None);
    };
    let Ok(duration_ms) = line[first_space + 1..second_space + 1].parse::<u64>() else {
        return Ok(None);
    };
    let result = if line.ends_with("ok") {
        "ok"
    } else if line.contains("not ok:") {
        "not ok"
    } else if line.ends_with("timeout") {
        "timeout"
    } else if line.ends_with("skipped") {
        "skipped"
    } else if line.ends_with("did not run") {
        "did not run"
    } else {
        return Err(format!("Couldn't determine result of log line {line}"));
    };
    let rest = &line[second_space + 1..];
    let colon = rest.find(':').ok_or_else(|| format!("Malformed line {line}"))? + second_space + 1;
    let test_file = line[second_space + 2..colon].to_string();
    let colon2 = line[colon + 1..].find(':').ok_or_else(|| format!("Malformed line {line}"))? + colon + 1;
    let line_num = line[colon + 1..colon2].parse().map_err(|_| format!("Failed to parse line number in {line}"))?;
    let marker = match result {
        "not ok" => "not ok: ",
        "ok" => "ok",
        other => other,
    };
    let end_of_query = line[colon2 + 1..].find(marker).map_or(colon2 + 1, |i| i + colon2 + 1);
    let query = line.get(colon2 + 2..end_of_query.saturating_sub(1)).unwrap_or_default().to_string();
    let error_message =
        if result == "not ok" { line[end_of_query + marker.len()..].to_string() } else { String::new() };
    Ok(Some(Entry { test_file, line_num, query, duration_ms, result, error_message }))
}

/// parse_log parses every result line of a log.
pub fn parse_log(text: &str) -> Result<Vec<Entry>, String> {
    text.lines().filter_map(|line| parse_line(line).transpose()).collect()
}

/// json_string encodes a string for JSON without escaping HTML characters, like the Go runner's encoder.
fn json_string(value: &str) -> String {
    let mut out = String::from("\"");
    for c in value.chars() {
        match c {
            '"' => out.push_str("\\\""),
            '\\' => out.push_str("\\\\"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            '\u{8}' => out.push_str("\\b"),
            '\u{c}' => out.push_str("\\f"),
            '\u{2028}' => out.push_str("\\u2028"),
            '\u{2029}' => out.push_str("\\u2029"),
            c if (c as u32) < 0x20 => out.push_str(&format!("\\u{:04x}", c as u32)),
            c => out.push(c),
        }
    }
    out.push('"');
    out
}

/// to_json matches the Go runner's JSON output of parsed results for importing into Dolt.
pub fn to_json(entries: &[Entry], version: &str) -> String {
    let rows: Vec<String> = entries
        .iter()
        .map(|e| {
            let mut row = format!(
                "{{\"version\":{},\"test_file\":{},\"line_num\":{},\"query_string\":{},\"duration\":{},\"result\":{}",
                json_string(version),
                json_string(&e.test_file),
                e.line_num,
                json_string(&e.query),
                e.duration_ms,
                json_string(e.result)
            );
            if !e.error_message.is_empty() {
                row.push_str(&format!(",\"error_message\":{}", json_string(&e.error_message)));
            }
            row.push('}');
            row
        })
        .collect();
    format!("{{\"rows\":[{}]}}\n", rows.join(","))
}

/// csv_field quotes a field the way Go's encoding/csv does.
fn csv_field(field: &str) -> String {
    let needs_quotes = field == "\\."
        || field.contains([',', '"', '\r', '\n'])
        || field.chars().next().is_some_and(char::is_whitespace);
    if needs_quotes { format!("\"{}\"", field.replace('"', "\"\"")) } else { field.to_string() }
}

/// to_csv matches the Go runner's CSV output of parsed results.
pub fn to_csv(entries: &[Entry], version: &str) -> String {
    let mut out = String::from("version,test_file,line_num,query_string,duration,result,error_message\n");
    for e in entries {
        let fields = [
            version.to_string(),
            e.test_file.clone(),
            e.line_num.to_string(),
            e.query.clone(),
            e.duration_ms.to_string(),
            e.result.to_string(),
            e.error_message.clone(),
        ];
        out.push_str(&fields.iter().map(|f| csv_field(f)).collect::<Vec<_>>().join(","));
        out.push('\n');
    }
    out
}
