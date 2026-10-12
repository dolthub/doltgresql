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

use base64::Engine;
use base64::engine::general_purpose::STANDARD;
use regression::gostd::Location;
use regression::pgx::{Value, read_cell, row_to_string};
use serde_json::Value as Json;

#[test]
fn cells_read_like_the_go_replay() {
    let fixture = std::fs::read_to_string(concat!(env!("CARGO_MANIFEST_DIR"), "/tests/data/pgx_cells.jsonl")).unwrap();
    let mut failures = Vec::new();
    for line in fixture.lines() {
        let cell: Json = serde_json::from_str(line).unwrap();
        let oid = cell["oid"].as_u64().unwrap() as u32;
        let value = cell["v"].as_str().map(|v| STANDARD.decode(v).unwrap());
        let actual = read_cell(oid, value.as_deref());
        let key = row_to_string(std::slice::from_ref(&actual));
        let location = match &actual {
            Value::Time(time) => match time.location {
                Location::Utc => Some("UTC".to_string()),
                Location::Fixed(offset) => Some(format!("Fixed({offset})")),
            },
            _ => None,
        };
        let negative_zero = matches!(actual, Value::Float64(v) if v == 0.0 && v.is_sign_negative());
        if key != cell["key"].as_str().unwrap()
            || actual.go_type() != cell["type"].as_str().unwrap()
            || location.as_deref() != cell["loc"].as_str()
            || negative_zero != cell["negzero"].as_bool().unwrap_or(false)
        {
            failures.push(format!("{line}\n  got {key:?} {} {location:?} {negative_zero}", actual.go_type()));
        }
    }
    assert!(failures.is_empty(), "{} cells differ:\n{}", failures.len(), failures[..failures.len().min(20)].join("\n"));
}
