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

use crate::format::{Reader, Writer};

/// Tracker holds the results of replaying one file.
#[derive(Clone, Debug, Default, PartialEq)]
pub struct Tracker {
    pub file: String,
    pub success: u32,
    pub partial_success: u32,
    pub failed: u32,
    pub success_items: Vec<Item>,
    pub fail_partial_items: Vec<Item>,
}

/// Item is one statement's result.
#[derive(Clone, Debug, Default, PartialEq)]
pub struct Item {
    pub query: String,
    pub partial_success: Vec<String>,
    pub unexpected_error: String,
    pub expected_error: String,
}

impl Tracker {
    /// new returns an empty tracker for the file.
    pub fn new(file: &str) -> Tracker {
        Tracker { file: file.to_string(), ..Tracker::default() }
    }

    /// fail counts a failure.
    pub fn fail(&mut self, query: &str, unexpected_error: &str, expected_error: &str) {
        self.failed += 1;
        self.fail_partial_items.push(Item {
            query: query.to_string(),
            partial_success: Vec::new(),
            unexpected_error: unexpected_error.to_string(),
            expected_error: expected_error.to_string(),
        });
    }

    /// succeed counts a success.
    pub fn succeed(&mut self, query: &str) {
        self.success += 1;
        self.success_items.push(Item { query: query.to_string(), ..Item::default() });
    }
}

/// serialize writes trackers in the version 2 format, sorted by file.
pub fn serialize(trackers: &[Tracker]) -> Vec<u8> {
    let mut sorted: Vec<&Tracker> = trackers.iter().collect();
    sorted.sort_by(|a, b| a.file.cmp(&b.file));
    let mut w = Writer::default();
    w.u32(2);
    w.u32(sorted.len() as u32);
    for tracker in sorted {
        w.string(&tracker.file);
        w.u32(tracker.success);
        w.u32(tracker.partial_success);
        w.u32(tracker.failed);
        w.u32(tracker.success_items.len() as u32);
        for item in &tracker.success_items {
            w.string(&item.query);
        }
        w.u32(tracker.fail_partial_items.len() as u32);
        for item in &tracker.fail_partial_items {
            w.string(&item.query);
            w.strings(&item.partial_success);
            w.string(&item.unexpected_error);
            w.string(&item.expected_error);
        }
    }
    w.finish()
}

/// deserialize reads trackers in the version 2 format, sorted by file.
pub fn deserialize(data: &[u8]) -> Result<Vec<Tracker>, String> {
    let mut r = Reader::new(data);
    let version = r.u32()?;
    if version != 2 {
        return Err(format!("version {version} is not supported"));
    }
    let mut trackers = Vec::new();
    for _ in 0..r.u32()? {
        let mut tracker = Tracker { file: r.string()?, success: r.u32()?, ..Tracker::default() };
        tracker.partial_success = r.u32()?;
        tracker.failed = r.u32()?;
        for _ in 0..r.u32()? {
            tracker.success_items.push(Item { query: r.string()?, ..Item::default() });
        }
        for _ in 0..r.u32()? {
            tracker.fail_partial_items.push(Item {
                query: r.string()?,
                partial_success: r.strings()?,
                unexpected_error: r.string()?,
                expected_error: r.string()?,
            });
        }
        trackers.push(tracker);
    }
    if !r.is_empty() {
        return Err("additional data remaining after all trackers have been read".to_string());
    }
    trackers.sort_by(|a, b| a.file.cmp(&b.file));
    Ok(trackers)
}
