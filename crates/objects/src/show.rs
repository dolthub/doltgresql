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

//! A canonical text form of root objects that matches how the Go graph oracle prints the Go structs: fields in
//! declaration order, strings as quoted hex, sorted maps, and interface values prefixed by their Go type.

use std::collections::BTreeMap;

/// quote renders a string as hex in quotes.
pub fn quote(bytes: &[u8]) -> String {
    let hex: String = bytes.iter().map(|b| format!("{b:02x}")).collect();
    format!("\"{hex}\"")
}

/// strings renders a list of strings.
pub fn strings(values: &[Vec<u8>]) -> String {
    format!("[{}]", values.iter().map(|v| quote(v)).collect::<Vec<_>>().join(" "))
}

/// string_map renders a map of strings, sorted by the rendered entries as Go's oracle sorts them.
pub fn string_map(map: &BTreeMap<Vec<u8>, Vec<u8>>) -> String {
    let mut entries: Vec<String> = map.iter().map(|(k, v)| format!("{}:{}", quote(k), quote(v))).collect();
    entries.sort();
    format!("map[{}]", entries.join(" "))
}

/// Fields renders a struct's fields in order.
#[derive(Default)]
pub struct Fields(Vec<String>);

impl Fields {
    pub fn new() -> Fields {
        Fields::default()
    }

    /// field adds a rendered field.
    pub fn field(mut self, name: &str, value: impl std::fmt::Display) -> Fields {
        self.0.push(format!("{name}:{value}"));
        self
    }

    /// string adds a string field.
    pub fn string(self, name: &str, value: &[u8]) -> Fields {
        self.field(name, quote(value))
    }

    pub fn finish(self) -> String {
        format!("{{{}}}", self.0.join(" "))
    }
}

/// list renders a list of rendered values.
pub fn list(values: impl IntoIterator<Item = String>) -> String {
    format!("[{}]", values.into_iter().collect::<Vec<_>>().join(" "))
}
