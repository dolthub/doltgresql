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
use std::path::Path;

use base64::Engine;
use base64::engine::general_purpose::STANDARD;
use pgproto::BackendMessage;
use regression::files::FILES;
use regression::gostd::Location;
use regression::messages::{Message, read_messages};
use regression::pgx::{Value, read_cell, row_to_string};
use serde_json::{Value as Json, json};

/// collect writes every distinct cell of the recordings and of the server cell file as JSON lines.
pub fn collect(root: &str, server_cells: &str, out: &str) {
    let root = Path::new(root);
    let mut seen = HashSet::<(u32, Option<Vec<u8>>)>::new();
    let mut cells = Vec::new();
    for file in FILES {
        let messages =
            read_messages(&root.join("results").join(format!("{file}.results")), &root.join("data")).unwrap();
        let mut oids: Vec<u32> = Vec::new();
        for message in messages {
            match message {
                Message::Backend(BackendMessage::RowDescription { fields }) => {
                    oids = fields.iter().map(|f| f.data_type_oid).collect();
                }
                Message::Backend(BackendMessage::DataRow { values }) if values.len() == oids.len() => {
                    for (oid, value) in oids.iter().zip(values) {
                        if seen.insert((*oid, value.clone())) {
                            cells.push((*oid, value));
                        }
                    }
                }
                _ => {}
            }
        }
    }
    let recorded = cells.len();
    for line in std::fs::read_to_string(server_cells).unwrap_or_default().lines() {
        let (oid, value) = line.split_once(' ').unwrap();
        let value = (value != "-").then(|| {
            (0..value.len()).step_by(2).map(|i| u8::from_str_radix(&value[i..i + 2], 16).unwrap()).collect::<Vec<u8>>()
        });
        let oid: u32 = oid.parse().unwrap();
        if seen.insert((oid, value.clone())) {
            cells.push((oid, value));
        }
    }
    let lines: String = cells
        .iter()
        .map(|(oid, value)| format!("{}\n", json!({"oid": oid, "v": value.as_ref().map(|v| STANDARD.encode(v))})))
        .collect();
    std::fs::write(out, lines).unwrap();
    eprintln!("{recorded} recorded cells and {} server cells", cells.len() - recorded);
}

/// check compares the Rust decoding of each cell against the Go oracle's output, printing mismatches by type.
pub fn check(cells: &str, oracle: &str) {
    let cells = std::fs::read_to_string(cells).unwrap();
    let oracle = std::fs::read_to_string(oracle).unwrap();
    let mut mismatches = BTreeMap::<u32, Vec<String>>::new();
    let mut checked = 0;
    let mut local = 0;
    for (cell, expected) in cells.lines().zip(oracle.lines()) {
        let cell: Json = serde_json::from_str(cell).unwrap();
        let expected: Json = serde_json::from_str(expected).unwrap();
        let oid = cell["oid"].as_u64().unwrap() as u32;
        let value = cell["v"].as_str().map(|v| STANDARD.decode(v).unwrap());
        checked += 1;
        if let Some(panic) = expected["panic"].as_str() {
            mismatches.entry(oid).or_default().push(format!("Go panics: {panic} on {:?}", lossy(&value)));
            continue;
        }
        let actual = read_cell(oid, value.as_deref());
        let key = row_to_string(std::slice::from_ref(&actual));
        let mut problems = Vec::new();
        if key != expected["key"].as_str().unwrap() {
            problems.push(format!("key {key:?} vs Go {:?}", expected["key"].as_str().unwrap()));
        }
        if actual.go_type() != expected["type"].as_str().unwrap() {
            problems.push(format!("type {} vs Go {}", actual.go_type(), expected["type"].as_str().unwrap()));
        }
        if let Value::Time(time) = &actual {
            let ours = match time.location {
                Location::Utc => "UTC".to_string(),
                Location::Fixed(offset) => format!("Fixed({offset})"),
            };
            match expected["loc"].as_str().unwrap_or_default() {
                "Local" => local += 1,
                theirs if theirs != ours => problems.push(format!("location {ours} vs Go {theirs}")),
                _ => {}
            }
        }
        if let Value::Float64(v) = actual
            && (v == 0.0 && v.is_sign_negative()) != expected["negzero"].as_bool().unwrap_or(false)
        {
            problems.push("negative zero differs".to_string());
        }
        if !problems.is_empty() {
            mismatches.entry(oid).or_default().push(format!("{:?}: {}", lossy(&value), problems.join("; ")));
        }
    }
    println!("checked {checked} cells, {local} times in the local zone");
    for (oid, problems) in &mismatches {
        println!("\noid {oid} ({}): {} mismatches", harness::oid::name(*oid).unwrap_or("?"), problems.len());
        for problem in problems.iter().take(8) {
            println!("  {}", problem.chars().take(400).collect::<String>());
        }
    }
}

fn lossy(value: &Option<Vec<u8>>) -> Option<String> {
    value.as_ref().map(|v| String::from_utf8_lossy(v).into_owned())
}
