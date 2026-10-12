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

use std::path::Path;

/// SEPARATOR divides a query from its expected results.
pub const SEPARATOR: &str = "----";
const DEFAULT_HASH_THRESHOLD: i64 = 8;

/// RecordType is the kind of a record.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum RecordType {
    Statement,
    Query,
    Halt,
}

/// Condition runs or skips a record on an engine.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Condition {
    pub only: bool,
    pub engine: String,
}

/// Record is a statement, a query with expected results, or a halt.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Record {
    pub record_type: RecordType,
    pub expect_error: bool,
    pub conditions: Vec<Condition>,
    pub schema: String,
    pub sort_mode: String,
    pub query: String,
    pub line_num: usize,
    pub result: Vec<String>,
    pub label: String,
    pub hash_threshold: i64,
}

/// fields splits on whitespace like Go's strings.Fields.
fn fields(s: &str) -> Vec<&str> {
    s.split(char::is_whitespace).filter(|f| !f.is_empty()).collect()
}

/// lines splits text like Go's bufio.Scanner, dropping a carriage return before each newline.
fn lines(text: &str) -> Vec<&str> {
    let mut lines: Vec<&str> = text.split('\n').map(|l| l.strip_suffix('\r').unwrap_or(l)).collect();
    if text.ends_with('\n') || text.is_empty() {
        lines.pop();
    }
    lines
}

/// parse_test_file parses a sqllogictest file.
pub fn parse_test_file(path: &Path) -> Result<Vec<Record>, String> {
    let bytes = std::fs::read(path).map_err(|e| format!("{}: {e}", path.display()))?;
    parse(&String::from_utf8_lossy(&bytes))
}

/// parse parses the text of a sqllogictest file, where a hash threshold carries over to later records.
pub fn parse(text: &str) -> Result<Vec<Record>, String> {
    let lines = lines(text);
    let mut position = 0;
    let mut records: Vec<Record> = Vec::new();
    while let Some(mut record) = parse_record(&lines, &mut position)? {
        if record.hash_threshold == -1 {
            record.hash_threshold = records.last().map_or(DEFAULT_HASH_THRESHOLD, |r| r.hash_threshold);
        }
        records.push(record);
    }
    Ok(records)
}

/// parse_record parses the next record, returning None at the end of the file.
fn parse_record(lines: &[&str], position: &mut usize) -> Result<Option<Record>, String> {
    #[derive(PartialEq)]
    enum State {
        Start,
        Statement,
        Query,
        Results,
    }
    let mut record = Record {
        record_type: RecordType::Statement,
        expect_error: false,
        conditions: Vec::new(),
        schema: String::new(),
        sort_mode: String::new(),
        query: String::new(),
        line_num: 0,
        result: Vec::new(),
        label: String::new(),
        hash_threshold: -1,
    };
    let mut state = State::Start;
    let mut query = String::new();
    while *position < lines.len() {
        let line = lines[*position];
        *position += 1;
        let line_num = *position;
        let blank = line.trim().is_empty();
        let without_comment = line.split('#').next().unwrap_or_default();
        if line.starts_with('#') {
            continue;
        }
        let fields = fields(without_comment);
        match state {
            State::Start => {
                if blank {
                    continue;
                }
                let field =
                    |i: usize| fields.get(i).copied().ok_or_else(|| format!("missing field on line {line_num}"));
                match field(0)? {
                    "halt" => {
                        record.record_type = RecordType::Halt;
                        record.line_num = line_num;
                        return Ok(Some(record));
                    }
                    kind @ ("skipif" | "onlyif") => {
                        record.conditions.push(Condition { only: kind == "onlyif", engine: field(1)?.to_string() })
                    }
                    "hash-threshold" => record.hash_threshold = field(1)?.parse().unwrap_or(0),
                    "statement" => {
                        record.record_type = RecordType::Statement;
                        record.expect_error = match field(1)? {
                            "ok" => false,
                            "error" => true,
                            other => return Err(format!("unexpected token {other}")),
                        };
                        state = State::Statement;
                    }
                    "query" => {
                        record.record_type = RecordType::Query;
                        record.schema = field(1)?.to_string();
                        record.sort_mode = fields.get(2).copied().unwrap_or_default().to_string();
                        record.label = fields.get(3).copied().unwrap_or_default().to_string();
                        state = State::Query;
                    }
                    other => return Err(format!("Unhandled statement {other} on line {line_num}")),
                }
            }
            State::Statement => {
                if blank {
                    record.query = query;
                    return Ok(Some(record));
                }
                if record.line_num == 0 {
                    record.line_num = line_num;
                }
                query.push_str(without_comment);
            }
            State::Query => {
                if record.line_num == 0 {
                    record.line_num = line_num;
                }
                if fields.len() == 1 && fields[0] == SEPARATOR {
                    record.query = query.clone();
                    state = State::Results;
                } else if blank {
                    record.query = query;
                    return Ok(Some(record));
                }
                query.push_str(without_comment);
            }
            State::Results => {
                if blank {
                    return Ok(Some(record));
                }
                record.result.push(without_comment.to_string());
            }
        }
    }
    if record.line_num == 0 {
        return Ok(None);
    }
    if state == State::Statement {
        record.query = query;
    }
    Ok(Some(record))
}

/// HashMatch is the first match of `(\d+) values hashing to ([0-9a-f]+)` in a line.
struct HashMatch<'a> {
    before: &'a str,
    count: &'a str,
    hash: &'a str,
    after: &'a str,
}

/// hash_match finds the leftmost match of `(\d+) values hashing to ([0-9a-f]+)`.
fn hash_match(line: &str) -> Option<HashMatch<'_>> {
    const MIDDLE: &str = " values hashing to ";
    let b = line.as_bytes();
    for start in 0..b.len() {
        let digits = b[start..].iter().take_while(|c| c.is_ascii_digit()).count();
        if digits == 0 || (start > 0 && b[start - 1].is_ascii_digit()) {
            continue;
        }
        let rest = &line[start + digits..];
        let Some(after_middle) = rest.strip_prefix(MIDDLE) else {
            continue;
        };
        let hash_length = after_middle.bytes().take_while(|c| matches!(c, b'0'..=b'9' | b'a'..=b'f')).count();
        if hash_length == 0 {
            continue;
        }
        return Some(HashMatch {
            before: &line[..start],
            count: &line[start..start + digits],
            hash: &after_middle[..hash_length],
            after: &after_middle[hash_length..],
        });
    }
    None
}

impl Record {
    /// is_hash_result reports whether the expected result is a hash rather than the values.
    pub fn is_hash_result(&self) -> bool {
        self.result.len() == 1 && hash_match(&self.result[0]).is_some()
    }

    /// hash returns the expected hash, along with any text around it as Go's replacement keeps it.
    pub fn hash(&self) -> String {
        let m = hash_match(&self.result[0]).unwrap();
        format!("{}{}{}", m.before, m.hash, m.after)
    }

    /// num_results returns how many values the query is expected to return.
    pub fn num_results(&self) -> usize {
        if !self.is_hash_result() {
            return self.result.len();
        }
        let m = hash_match(&self.result[0]).unwrap();
        format!("{}{}{}", m.before, m.count, m.after).parse().unwrap_or(0)
    }

    /// should_execute_for_engine reports whether the record runs on the engine.
    pub fn should_execute_for_engine(&self, engine: &str) -> bool {
        if self.conditions.len() == 1 && self.conditions[0].only {
            return self.conditions[0].engine == engine;
        }
        !self.conditions.iter().any(|c| !c.only && c.engine == engine)
    }

    /// sort_results sorts the results by the record's sort mode, failing like the Go runner's panic otherwise.
    pub fn sort_results(&self, mut results: Vec<String>) -> Result<Vec<String>, String> {
        match self.sort_mode.as_str() {
            "nosort" => Ok(results),
            "rowsort" => {
                let columns = self.schema.len();
                if columns == 0 {
                    return Err("runtime error: integer divide by zero".to_string());
                }
                let full = results.len() / columns * columns;
                let remainder = results.split_off(full);
                let mut rows: Vec<Vec<String>> = results.chunks(columns).map(<[String]>::to_vec).collect();
                rows.sort();
                let mut sorted = rows.concat();
                sorted.extend(remainder);
                Ok(sorted)
            }
            "valuesort" => {
                results.sort();
                Ok(results)
            }
            other => Err(format!("Uncrecognized sort mode {other}")),
        }
    }
}
