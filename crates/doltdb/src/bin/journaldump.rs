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

//! Prints a chunk journal's records in order, describing each chunk's message, to show the order a writer wrote them:
//!
//!     journaldump <journal file>

use serial::{Commit, DoltgresRootValue, Message, StoreRoot, WorkingSet};
use store::{Chunk, JournalRecord, read_records};

/// text renders bytes as text.
fn text(bytes: &[u8]) -> String {
    String::from_utf8_lossy(bytes).into_owned()
}

/// describe summarizes a chunk's message.
fn describe(data: &[u8]) -> String {
    let message = Message(data);
    let id = message.file_id();
    let detail = match id {
        serial::STORE_ROOT => {
            let map = StoreRoot::new(message).unwrap().address_map().unwrap().unwrap_or_default();
            let node = prolly::Node::decode(map.to_vec()).unwrap();
            (0..node.count())
                .map(|i| format!("{}={}", text(node.key(i).unwrap()), node.child(i).unwrap()))
                .collect::<Vec<_>>()
                .join(" ")
        }
        serial::COMMIT => {
            let c = Commit::new(message).unwrap();
            format!(
                "root={} parents={:?} height={} name={} email={} desc={:?} ts={} uts={}",
                c.root().unwrap(),
                c.parents().unwrap().iter().map(|p| p.to_string()).collect::<Vec<_>>(),
                c.height().unwrap(),
                text(c.name().unwrap()),
                text(c.email().unwrap()),
                text(c.description().unwrap()),
                c.timestamp_millis().unwrap(),
                c.user_timestamp_millis().unwrap()
            )
        }
        serial::WORKING_SET => {
            let w = WorkingSet::new(message).unwrap();
            format!(
                "working={} staged={:?} name={} email={} desc={:?} ts={}",
                w.working_root().unwrap(),
                w.staged_root().unwrap().map(|h| h.to_string()),
                text(w.name().unwrap()),
                text(w.email().unwrap()),
                text(w.description().unwrap()),
                w.timestamp_millis().unwrap()
            )
        }
        serial::DOLTGRES_ROOT_VALUE => {
            let r = DoltgresRootValue::new(message).unwrap();
            let schemas: Vec<String> = r.schemas().unwrap().iter().map(|s| text(s)).collect();
            format!("fv={} collation={} schemas={schemas:?}", r.feature_version().unwrap(), r.collation().unwrap())
        }
        _ => String::new(),
    };
    format!("{id} {detail}")
}

fn main() {
    let path = std::env::args().nth(1).expect("usage: journaldump <journal file>");
    let bytes = std::fs::read(path).unwrap();
    let mut offset = 0;
    for (record, raw) in read_records(&bytes).unwrap() {
        match record {
            JournalRecord::Root { hash, timestamp } => println!("{offset} root {hash} at {timestamp}"),
            JournalRecord::Chunk { hash, record } => {
                let chunk = Chunk::from_record(hash, record).unwrap();
                println!("{offset} chunk {hash} {}", describe(&chunk.data));
            }
        }
        offset += raw.len();
    }
}
