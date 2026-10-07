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

//! The rows that a fresh Postgres 15 database has in its system catalogs, which describe the built-in objects.

use std::collections::HashMap;
use std::sync::{Mutex, OnceLock};

use crate::pgcatalog::CatalogTable;
use crate::types::{Reg, Value};

/// BUILTIN_ROWS holds every system catalog relation's rows in a fresh Postgres 15 database, compressed with zstd: a
/// `== schema.name` line before each relation's rows, which are in COPY's text format, with reg values written as
/// `<oid> <name>`.
const BUILTIN_ROWS: &[u8] = include_bytes!("builtin_rows.zst");

/// sections returns the text of each relation's rows by qualified name, decompressing them on first use.
fn sections() -> &'static HashMap<String, String> {
    static SECTIONS: OnceLock<HashMap<String, String>> = OnceLock::new();
    SECTIONS.get_or_init(|| {
        let bytes = zstd::decode_all(BUILTIN_ROWS).expect("the built-in catalog rows decompress");
        let text = String::from_utf8(bytes).expect("the built-in catalog rows are UTF-8");
        let mut sections = HashMap::new();
        let mut current: Option<(String, String)> = None;
        for line in text.lines() {
            if let Some(name) = line.strip_prefix("== ") {
                sections.extend(current.take());
                current = Some((name.to_string(), String::new()));
            } else if let Some((_, rows)) = current.as_mut() {
                rows.push_str(line);
                rows.push('\n');
            }
        }
        sections.extend(current);
        sections
    })
}

/// unescape reads a field of COPY's text format.
fn unescape(field: &str) -> String {
    let mut out = String::with_capacity(field.len());
    let mut chars = field.chars().peekable();
    while let Some(c) = chars.next() {
        if c != '\\' {
            out.push(c);
            continue;
        }
        match chars.next() {
            Some('t') => out.push('\t'),
            Some('n') => out.push('\n'),
            Some('r') => out.push('\r'),
            Some('b') => out.push('\u{8}'),
            Some('f') => out.push('\u{c}'),
            Some('v') => out.push('\u{b}'),
            Some(d @ '0'..='7') => {
                let mut code = d.to_digit(8).unwrap_or(0);
                for _ in 0..2 {
                    match chars.peek().and_then(|c| c.to_digit(8)) {
                        Some(digit) => {
                            code = code * 8 + digit;
                            chars.next();
                        }
                        None => break,
                    }
                }
                out.push(char::from_u32(code).unwrap_or('\u{fffd}'));
            }
            Some(other) => out.push(other),
            None => out.push('\\'),
        }
    }
    out
}

/// value reads a field as a value of the column's type, keeping the text for types without their own values.
fn value(field: &str, type_oid: u32) -> Value {
    if field == "\\N" {
        return Value::Null;
    }
    let text = unescape(field);
    if crate::cast::is_reg_type(type_oid) {
        let (oid, name) = text.split_once(' ').unwrap_or((text.as_str(), ""));
        return Value::Reg(Box::new(Reg { type_oid, oid: oid.parse().unwrap_or(0), name: name.to_string() }));
    }
    match crate::cast::input(&text, type_oid) {
        Ok(value) => value,
        Err(_) => Value::Text(text),
    }
}

/// Parsed holds each relation's rows once they are read, by schema and name.
type Parsed = Mutex<HashMap<(&'static str, &'static str), Vec<Vec<Value>>>>;

/// rows returns the built-in rows of a system catalog relation.
pub fn rows(table: &CatalogTable) -> Vec<Vec<Value>> {
    static PARSED: OnceLock<Parsed> = OnceLock::new();
    let parsed = PARSED.get_or_init(|| Mutex::new(HashMap::new()));
    let mut parsed = parsed.lock().unwrap_or_else(|e| e.into_inner());
    parsed
        .entry((table.schema, table.name))
        .or_insert_with(|| {
            let Some(section) = sections().get(&format!("{}.{}", table.schema, table.name)) else { return Vec::new() };
            section
                .lines()
                .map(|line| line.split('\t').zip(&table.columns).map(|(f, c)| value(f, c.type_oid)).collect())
                .collect()
        })
        .clone()
}

/// Implemented is an operator that a function implements: its name and its left and right operand types, with a left
/// type of 0 for a prefix operator.
pub type Implemented = (String, u32, u32);

/// operator_implementations returns the built-in operators by the name of the function that implements each, as
/// pg_operator's oprcode names it.
pub fn operator_implementations() -> &'static HashMap<String, Vec<Implemented>> {
    static IMPLEMENTATIONS: OnceLock<HashMap<String, Vec<Implemented>>> = OnceLock::new();
    IMPLEMENTATIONS.get_or_init(|| {
        let mut implementations: HashMap<String, Vec<Implemented>> = HashMap::new();
        let Some(table) = super::lookup("pg_catalog", "pg_operator") else { return implementations };
        let column = |name: &str| table.columns.iter().position(|c| c.name == name);
        let (Some(name), Some(left), Some(right), Some(code)) =
            (column("oprname"), column("oprleft"), column("oprright"), column("oprcode"))
        else {
            return implementations;
        };
        for row in rows(table) {
            let oid = |value: &Value| match value {
                Value::Oid(oid) => *oid,
                _ => 0,
            };
            if let (Value::Text(operator), Value::Reg(function)) = (&row[name], &row[code]) {
                let operator = (operator.clone(), oid(&row[left]), oid(&row[right]));
                implementations.entry(function.name.clone()).or_default().push(operator);
            }
        }
        implementations
    })
}
