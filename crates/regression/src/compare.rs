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

use std::collections::{BTreeMap, HashMap};

use pgproto::FieldDescription;

use crate::pgx::{Value, read_cell, row_to_string};

/// MIN_USER_OID is the first OID that Postgres assigns to user objects.
pub const MIN_USER_OID: u32 = 16384;

/// Row is the values of a DataRow, where None is NULL.
pub type Row = Vec<Option<Vec<u8>>>;

/// OidMap maps the OIDs of the recorded session to the OIDs that the server assigned to the same objects.
#[derive(Default, Clone, Debug)]
pub struct OidMap {
    pub oids: HashMap<u32, u32>,
}

impl OidMap {
    /// get returns the server OID for a recorded OID.
    pub fn get(&self, recorded: u32) -> Option<u32> {
        self.oids.get(&recorded).copied()
    }

    /// put_all records the given pairs, replacing earlier ones.
    pub fn put_all(&mut self, replacements: HashMap<u32, u32>) {
        self.oids.extend(replacements);
    }

    /// rewrite_query replaces every standalone run of 5 to 10 digits that is a recorded OID with its server OID.
    pub fn rewrite_query(&self, query: &str) -> String {
        if self.oids.is_empty() {
            return query.to_string();
        }
        let b = query.as_bytes();
        let is_word = |c: u8| c.is_ascii_alphanumeric() || c == b'_';
        let mut out = String::with_capacity(query.len() + 16);
        let mut last = 0;
        let mut i = 0;
        while i < b.len() {
            if !b[i].is_ascii_digit() {
                i += 1;
                continue;
            }
            let start = i;
            while i < b.len() && b[i].is_ascii_digit() {
                i += 1;
            }
            let length = i - start;
            if !(5..=10).contains(&length) || (start > 0 && is_word(b[start - 1])) || (i < b.len() && is_word(b[i])) {
                continue;
            }
            let Ok(parsed) = query[start..i].parse::<u32>() else {
                continue;
            };
            let Some(mapped) = self.get(parsed) else {
                continue;
            };
            out.push_str(&query[last..start]);
            out.push_str(&mapped.to_string());
            last = i;
        }
        out.push_str(&query[last..]);
        out
    }
}

/// read_rows matches the replay's ReadRows.
pub fn read_rows(fields: &[FieldDescription], rows: &[Row]) -> Vec<Vec<Value>> {
    rows.iter()
        .map(|row| {
            if row.len() != fields.len() {
                return Vec::new();
            }
            fields.iter().zip(row).map(|(field, cell)| read_cell(field.data_type_oid, cell.as_deref())).collect()
        })
        .collect()
}

/// oid_columns returns, for each recorded column, whether it has the `oid` type.
fn oid_columns(fields: &[FieldDescription]) -> Vec<bool> {
    fields.iter().map(|field| field.data_type_oid == 26).collect()
}

/// cells_match matches the replay's cellsMatch, tentatively accepting user OID pairs that are not candidates yet.
fn cells_match(a: &Value, b: &Value, is_oid_column: bool, candidates: &mut HashMap<u32, u32>) -> bool {
    if a.go_eq(b) {
        return true;
    }
    if !is_oid_column {
        return false;
    }
    let (Some(a_oid), Some(b_oid)) = (a.as_oid(), b.as_oid()) else {
        return false;
    };
    if a_oid < MIN_USER_OID || b_oid == 0 {
        return false;
    }
    if let Some(mapped) = candidates.get(&a_oid) {
        return *mapped == b_oid;
    }
    candidates.insert(a_oid, b_oid);
    true
}

/// rows_to_error_string matches the replay's rowsToErrorString.
fn rows_to_error_string(postgres: &[Vec<Value>], doltgres: &[Vec<Value>]) -> String {
    let mut out = String::from("    Postgres:\n");
    for row in postgres {
        out.push_str(&format!("        {{{}}}\n", row_to_string(row)));
    }
    out.push_str("    Doltgres:\n");
    for row in doltgres {
        out.push_str(&format!("        {{{}}}\n", row_to_string(row)));
    }
    out.pop();
    out
}

/// compare_rows_ordered matches the replay's CompareRowsOrdered.
pub fn compare_rows_ordered(
    oid_map: &mut OidMap,
    a_fields: &[FieldDescription],
    b_fields: &[FieldDescription],
    a_rows: &[Row],
    b_rows: &[Row],
) -> Result<(), String> {
    if a_rows.len() != b_rows.len() {
        return Err(format!("expected a row count of {} but received {}", a_rows.len(), b_rows.len()));
    }
    let a_read = read_rows(a_fields, a_rows);
    let b_read = read_rows(b_fields, b_rows);
    let oid_columns = oid_columns(a_fields);
    let mut candidates = HashMap::new();
    for (a_row, b_row) in a_read.iter().zip(&b_read) {
        if a_row.len() != b_row.len() {
            return Err(format!("expected a row column count of {} but received {}", a_row.len(), b_row.len()));
        }
        for (column, (a, b)) in a_row.iter().zip(b_row).enumerate() {
            if !cells_match(a, b, oid_columns[column], &mut candidates) {
                return Err(if a_read.len() + b_read.len() < 8 {
                    format!("row sets differ:\n{}", rows_to_error_string(&a_read, &b_read))
                } else {
                    format!(
                        "rows differ\n    Postgres:\n        {{{}}}\n    Doltgres:\n        {{{}}}",
                        row_to_string(a_row),
                        row_to_string(b_row)
                    )
                });
            }
        }
    }
    oid_map.put_all(candidates);
    Ok(())
}

/// compare_rows_unordered matches the replay's CompareRowsUnordered, which compares the rows as multisets.
pub fn compare_rows_unordered(
    oid_map: &mut OidMap,
    a_fields: &[FieldDescription],
    b_fields: &[FieldDescription],
    a_rows: &[Row],
    b_rows: &[Row],
) -> Result<(), String> {
    if a_rows.len() != b_rows.len() {
        return Err(format!("expected a row count of {} but received {}", a_rows.len(), b_rows.len()));
    }
    let mut a_read = read_rows(a_fields, a_rows);
    let b_read = read_rows(b_fields, b_rows);
    let oid_columns = oid_columns(a_fields);
    let has_oid_columns = oid_columns.contains(&true);
    if has_oid_columns {
        for row in &mut a_read {
            for (column, cell) in row.iter_mut().enumerate() {
                if !oid_columns[column] {
                    continue;
                }
                if let Some(oid) = cell.as_oid().filter(|oid| *oid >= MIN_USER_OID)
                    && let Some(mapped) = oid_map.get(oid)
                {
                    *cell = Value::Uint32(mapped);
                }
            }
        }
    }
    let result = compare_multiset(&a_read, &b_read);
    if result.is_err()
        && has_oid_columns
        && a_read.len() <= 2000
        && let Some(candidates) = match_with_oid_replacement(&oid_columns, &a_read, &b_read)
    {
        oid_map.put_all(candidates);
        return Ok(());
    }
    result
}

/// match_with_oid_replacement greedily matches each recorded row to a server row, returning the learned OID pairs.
fn match_with_oid_replacement(
    oid_columns: &[bool],
    a_read: &[Vec<Value>],
    b_read: &[Vec<Value>],
) -> Option<HashMap<u32, u32>> {
    let mut candidates = HashMap::new();
    let mut used = vec![false; b_read.len()];
    for a_row in a_read {
        let mut matched = false;
        for (b_index, b_row) in b_read.iter().enumerate() {
            if used[b_index] || a_row.len() != b_row.len() {
                continue;
            }
            let mut trial = candidates.clone();
            if a_row.iter().zip(b_row).enumerate().all(|(c, (a, b))| cells_match(a, b, oid_columns[c], &mut trial)) {
                candidates = trial;
                used[b_index] = true;
                matched = true;
                break;
            }
        }
        if !matched {
            return None;
        }
    }
    Some(candidates)
}

/// row_kvs_to_error_string matches the replay's rowKVsToErrorString.
fn row_kvs_to_error_string(postgres: &BTreeMap<String, usize>, doltgres: &BTreeMap<String, usize>) -> String {
    let mut out = String::from("    Postgres:\n");
    for (key, count) in postgres {
        for _ in 0..*count {
            out.push_str(&format!("        {{{key}}}\n"));
        }
    }
    out.push_str("    Doltgres:\n");
    for (key, count) in doltgres {
        for _ in 0..*count {
            out.push_str(&format!("        {{{key}}}\n"));
        }
    }
    out.pop();
    out
}

/// compare_multiset matches the replay's compareRowsMultiset, which uses each row's printed form as its identity.
fn compare_multiset(a_read: &[Vec<Value>], b_read: &[Vec<Value>]) -> Result<(), String> {
    let mut a_counts = BTreeMap::<String, usize>::new();
    let mut b_counts = BTreeMap::<String, usize>::new();
    for (a_row, b_row) in a_read.iter().zip(b_read) {
        if a_row.len() != b_row.len() {
            return Err(format!("expected a row column count of {} but received {}", a_row.len(), b_row.len()));
        }
        *a_counts.entry(row_to_string(a_row)).or_default() += 1;
        *b_counts.entry(row_to_string(b_row)).or_default() += 1;
    }
    let total = a_counts.values().sum::<usize>() + b_counts.values().sum::<usize>();
    if a_counts.len() != b_counts.len() {
        return Err(if total < 8 {
            format!("row sets differ:\n{}", row_kvs_to_error_string(&a_counts, &b_counts))
        } else {
            "row sets differ (too large to display)".to_string()
        });
    }
    for ((a_key, a_count), (b_key, b_count)) in a_counts.iter().zip(&b_counts) {
        if a_key != b_key || a_count != b_count {
            return Err(if total < 8 {
                format!("row sets differ:\n{}", row_kvs_to_error_string(&a_counts, &b_counts))
            } else if a_key != b_key {
                format!("could not find the following row in the result set:\n        {{{a_key}}}")
            } else {
                format!(
                    "for the following row, expected to find {a_count} duplicates but found {b_count}:\n {{{a_key}}}"
                )
            });
        }
    }
    Ok(())
}

/// compare_copy_data_ordered matches the replay's CompareCopyDataOrdered, where each message is one row.
pub fn compare_copy_data_ordered(expected: &[Vec<u8>], actual: &[Vec<u8>]) -> Result<(), String> {
    if expected.len() != actual.len() {
        return Err(format!("expected a COPY data row count of {} but received {}", expected.len(), actual.len()));
    }
    for (e, a) in expected.iter().zip(actual) {
        if e != a {
            return Err(format!(
                "COPY data differs from the expected data:\n    Postgres: {{{}}}\n    Doltgres: {{{}}}",
                String::from_utf8_lossy(e),
                String::from_utf8_lossy(a)
            ));
        }
    }
    Ok(())
}

/// compare_copy_data_unordered matches the replay's CompareCopyDataUnordered.
pub fn compare_copy_data_unordered(expected: &[Vec<u8>], actual: &[Vec<u8>]) -> Result<(), String> {
    if expected.len() != actual.len() {
        return Err(format!("expected a COPY data row count of {} but received {}", expected.len(), actual.len()));
    }
    let mut counts = HashMap::<&[u8], i64>::new();
    for e in expected {
        *counts.entry(e).or_default() += 1;
    }
    for a in actual {
        let count = counts.entry(a).or_default();
        *count -= 1;
        if *count < 0 {
            return Err(format!("COPY data contains an unexpected row: {{{}}}", String::from_utf8_lossy(a)));
        }
    }
    Ok(())
}
