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

use std::collections::{HashMap, HashSet};
use std::path::Path;

use doltdb::database::Database;
use doltdb::root::Root;
use prolly::Node;
use serial::{Message, WorkingSet};
use store::Hash;

use crate::catalog::table::TableDef;
use crate::error::{PgError, Result};

/// SENTINEL is the file in a database's .dolt directory that records that the database passed the startup integrity
/// check, so that later starts skip it.
pub const SENTINEL: &str = ".integrity_check_passed";

/// CHECK_VERSION is what a sentinel records, which changes whenever the check learns to find a new kind of
/// corruption.
const CHECK_VERSION: &str = "1";

/// Stats counts the chunk addresses that a tree's nodes fail to record, and the rows that hold them.
#[derive(Clone, Copy, Default)]
struct Stats {
    missing: u64,
    rows: u64,
}

/// Scanner checks trees, remembering each node it has checked with the encodings it read the node with.
struct Scanner<'d> {
    db: &'d mut Database,
    checked: HashMap<(Hash, Vec<u8>, Vec<u8>), Stats>,
}

/// check_data_dir checks every database of a data directory that has not yet passed the check, as the Go server's
/// startup does, recording a pass in each healthy database's sentinel, and returns the error of the first corrupt
/// database.
pub fn check_data_dir(data_dir: &Path) -> Result<Option<String>> {
    let mut databases: Vec<_> = std::fs::read_dir(data_dir)
        .map_err(PgError::internal)?
        .filter_map(|e| e.ok())
        .filter(|e| e.path().join(".dolt").is_dir())
        .collect();
    databases.sort_by_key(|e| e.file_name());
    for entry in databases {
        let dolt = entry.path().join(".dolt");
        let passed = std::fs::read_to_string(dolt.join(SENTINEL)).is_ok_and(|text| text.trim() == CHECK_VERSION);
        if passed {
            continue;
        }
        let name = entry.file_name().to_string_lossy().into_owned();
        let mut db = Database::open(&dolt.join("noms"))?;
        if let Some(message) = check_database(&mut db, &name)? {
            return Ok(Some(message));
        }
        std::fs::write(dolt.join(SENTINEL), format!("{CHECK_VERSION}\n")).map_err(PgError::internal)?;
    }
    Ok(None)
}

/// check_database scans every table of every commit that a branch reaches, and of each branch's working set, for
/// values whose chunk addresses their nodes fail to record, as earlier releases wrote them, returning the error that
/// describes the first corrupt table found.
pub fn check_database(db: &mut Database, name: &str) -> Result<Option<String>> {
    let branches: Vec<(String, Hash)> = db
        .datasets()?
        .into_iter()
        .filter_map(|(r, head)| r.strip_prefix("refs/heads/").map(|branch| (branch.to_string(), head)))
        .collect();
    let mut scanner = Scanner { db, checked: HashMap::new() };
    let mut visited = HashSet::new();
    for (branch, head) in branches {
        let mut stack = vec![head];
        while let Some(commit) = stack.pop() {
            if !visited.insert(commit) {
                continue;
            }
            let Ok(info) = crate::dolt::history::load(scanner.db, commit) else { continue };
            if let Some((table, stats)) = scanner.check_root(info.root)? {
                let location = format!("commit {commit} (reachable from branch {branch})");
                return Ok(Some(corruption(name, &table, &location, stats)));
            }
            stack.extend(info.parents);
        }
        let Some(address) = scanner.db.head(&doltdb::create::working_set_ref(&branch))? else { continue };
        let data = crate::txn::read(scanner.db, &address)?;
        let working_set = WorkingSet::new(Message(&data))?;
        for root in [Some(working_set.working_root()?), working_set.staged_root()?].into_iter().flatten() {
            if let Some((table, stats)) = scanner.check_root(root)? {
                let location = format!("the working set of branch {branch}");
                return Ok(Some(corruption(name, &table, &location, stats)));
            }
        }
    }
    Ok(None)
}

/// corruption returns the Go server's error for a corrupt table, with how to repair it.
fn corruption(database: &str, table: &str, location: &str, stats: Stats) -> String {
    format!(
        "database \"{database}\" failed a startup integrity check: table {table} at {location} has {} values (in {} \
         rows) that were serialized incorrectly due to an error in a previous release of Doltgres. To avoid data loss \
         with this version, the server will now exit.

To repair this database:
  1. MAKE A BACKUP COPY of the database directory before doing anything else.
  2. Build the repair tool from the DoltgreSQL repository at https://github.com/dolthub/doltgresql/ with this command:
\t\t\t\tgo build -o doltgres-admin ./cmd/admin
  3. Run: doltgres-admin report -dir <data-dir> to see the full extent of the corruption, then
     doltgres-admin repair -dir <data-dir> to repair it.

To start the server without this check (unsafe until the database is repaired), set \
         behavior.skip_startup_integrity_check to true in config.yaml.",
        stats.missing, stats.rows
    )
}

impl Scanner<'_> {
    /// check_root scans the tables of a root value that hold adaptive values, returning the first corrupt one's name
    /// and counts.
    fn check_root(&mut self, address: Hash) -> Result<Option<(String, Stats)>> {
        let root = Root::decode(&crate::txn::read(self.db, &address)?)?;
        for (key, address) in root.tables(self.db)? {
            let text = String::from_utf8_lossy(&key).into_owned();
            let mut parts = text.splitn(3, '\0').skip(1);
            let (Some(schema), Some(name)) = (parts.next(), parts.next()) else { continue };
            let table = TableDef::load(self.db, schema, name, address)?;
            let (keys, values) = (table.key_encodings(), table.value_encodings());
            if !keys.iter().chain(&values).any(|e| doltdb::table::ADAPTIVE_ENCODINGS.contains(e)) {
                continue;
            }
            let rows = &table.table.primary_index;
            let stats = self.scan(Node::decode(rows.clone())?, Hash::of(rows), &keys, &values)?;
            if stats.missing > 0 {
                let shown = if schema.is_empty() { name.to_string() } else { format!("{schema}.{name}") };
                return Ok(Some((shown, stats)));
            }
        }
        Ok(None)
    }

    /// scan counts the addresses that a node and the nodes below it fail to record.
    fn scan(&mut self, node: Node, hash: Hash, keys: &[u8], values: &[u8]) -> Result<Stats> {
        let cache_key = (hash, keys.to_vec(), values.to_vec());
        if let Some(&stats) = self.checked.get(&cache_key) {
            return Ok(stats);
        }
        let (recorded_keys, recorded_values) = node.address_offsets()?;
        let mut rows = vec![false; node.count()];
        let mut stats = Stats { missing: unrecorded(&node, true, keys, &recorded_keys, &mut rows)?, rows: 0 };
        if node.is_leaf() {
            stats.missing += unrecorded(&node, false, values, &recorded_values, &mut rows)?;
            stats.rows = rows.iter().filter(|r| **r).count() as u64;
        } else {
            for i in 0..node.count() {
                let child = node.child(i)?;
                let below = self.scan(Node::decode(crate::txn::read(self.db, &child)?)?, child, keys, values)?;
                stats.missing += below.missing;
                stats.rows += below.rows;
            }
        }
        self.checked.insert(cache_key, stats);
        Ok(stats)
    }
}

/// unrecorded counts the chunk addresses in a node's keys or values that the node's recorded positions leave out,
/// marking the items that hold them.
fn unrecorded(node: &Node, keys: bool, encodings: &[u8], recorded: &[u16], rows: &mut [bool]) -> Result<u64> {
    let items = if keys { node.key_items() } else { node.value_items() };
    let mut remaining: HashMap<u16, usize> = HashMap::new();
    for &offset in recorded {
        *remaining.entry(offset).or_default() += 1;
    }
    let mut missing = 0;
    for (i, row) in rows.iter_mut().enumerate() {
        let item = if keys { node.key(i)? } else { node.value(i)? };
        let start = item.as_ptr() as usize - items.as_ptr() as usize;
        for offset in doltdb::table::address_offsets(item, encodings) {
            let at = offset as usize;
            if item.get(at..at + Hash::LEN).is_none_or(|address| address.iter().all(|&b| b == 0)) {
                continue;
            }
            match remaining.get_mut(&((start + at) as u16)) {
                Some(count) if *count > 0 => *count -= 1,
                _ => {
                    missing += 1;
                    *row = true;
                }
            }
        }
    }
    Ok(missing)
}
