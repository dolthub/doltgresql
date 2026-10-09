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

//! Three-way merges of root values, as Dolt merges a commit into a branch: each table's rows merge by primary key,
//! a row that both sides changed differently becomes a conflict, and a row that breaks a unique index or a foreign
//! key becomes a constraint violation.

use std::cmp::Ordering;
use std::collections::{BTreeMap, HashMap};

use doltdb::database::Database;
use doltdb::root::Root;
use prolly::Tuple;
use store::Hash;

use crate::catalog::table::{HIDDEN_BASE, IndexDef, TableDef};
use crate::dolt::args::error;
use crate::dolt::artifacts::{self, Artifact};
use crate::error::Result;
use crate::json::Json;
use crate::query::Ctx;
use crate::txn::read;
use crate::types::Value;
use serial::write::DEFAULT_TARGET_ROW_SIZE;

/// Name is a table by schema and name.
type Name = (String, String);

/// Entry is a row as stored: its key and value tuples.
type Entry = (Vec<u8>, Vec<u8>);

/// Edit is a change to a stored row: its key and its new value, or None to delete it.
type Edit = (Vec<u8>, Option<Vec<u8>>);

/// Commits are the commits a merge combines: ours, the one whose changes it brings in, and the merge base.
#[derive(Clone, Copy, Debug)]
pub struct Commits {
    pub ours: Hash,
    pub theirs: Hash,
    pub base: Hash,
}

/// Outcome is the result of a merge: the merged root, whether it left conflicts or constraint violations in any
/// table, and the tables whose schemas conflict.
pub struct Outcome {
    pub root: Root,
    pub artifacts: bool,
    pub schema_conflicts: Vec<Name>,
}

/// Change is a row's change on one side of a merge: its key, its value in the merge base, and its value on that side,
/// where a missing value is a missing row.
#[derive(Clone, Debug)]
struct Change {
    key: Vec<u8>,
    from: Option<Vec<u8>>,
    to: Option<Vec<u8>>,
}

/// load returns a table of a root, if the root has it.
fn load(db: &mut Database, root: &Root, name: &Name) -> Result<Option<(Hash, TableDef)>> {
    match root.table(db, &name.0, &name.1)? {
        Some(address) => Ok(Some((address, TableDef::load(db, &name.0, &name.1, address)?))),
        None => Ok(None),
    }
}

/// entries returns the stored rows of a table in key order.
fn entries(db: &mut Database, table: &TableDef) -> Result<Vec<Entry>> {
    let node = prolly::Node::decode(table.table.primary_index.clone())?;
    let mut items = Vec::new();
    prolly::walk_leaves(db, &node, &mut |key, value| {
        items.push((key.to_vec(), value.to_vec()));
        Ok(())
    })?;
    Ok(items)
}

/// changes returns the rows that differ between two versions of a table's rows, in key order.
fn changes(table: &TableDef, from: &[Entry], to: &[Entry]) -> Vec<Change> {
    let mut out = Vec::new();
    let (mut i, mut j) = (0, 0);
    while i < from.len() || j < to.len() {
        let ordering = match (from.get(i), to.get(j)) {
            (Some(f), Some(t)) => table.compare_keys(&f.0, &t.0),
            (Some(_), None) => Ordering::Less,
            _ => Ordering::Greater,
        };
        match ordering {
            Ordering::Less => {
                out.push(Change { key: from[i].0.clone(), from: Some(from[i].1.clone()), to: None });
                i += 1;
            }
            Ordering::Greater => {
                out.push(Change { key: to[j].0.clone(), from: None, to: Some(to[j].1.clone()) });
                j += 1;
            }
            Ordering::Equal => {
                if from[i].1 != to[j].1 {
                    out.push(Change { key: to[j].0.clone(), from: Some(from[i].1.clone()), to: Some(to[j].1.clone()) });
                }
                i += 1;
                j += 1;
            }
        }
    }
    out
}

/// changes_between returns the rows that differ between two versions of a table's rows, in key order, reading only
/// the nodes of the subtrees where the versions differ.
fn changes_between(db: &mut Database, table: &TableDef, from: &TableDef, to: &TableDef) -> Result<Vec<Change>> {
    let from = prolly::Node::decode(from.table.primary_index.clone())?;
    let to = prolly::Node::decode(to.table.primary_index.clone())?;
    let mut out = Vec::new();
    if from.bytes() != to.bytes() {
        diff_runs(db, table, vec![from], vec![to], &mut out)?;
    }
    Ok(out)
}

/// diff_runs adds the changes between two key-ordered runs of nodes that cover the same keys to a list, expanding
/// the higher run until both are of one level and then skipping the subtrees that both runs share.
fn diff_runs(
    db: &mut Database,
    table: &TableDef,
    mut from: Vec<prolly::Node>,
    mut to: Vec<prolly::Node>,
    out: &mut Vec<Change>,
) -> Result<()> {
    let level = |run: &[prolly::Node]| run.iter().map(prolly::Node::level).max().unwrap_or(0);
    while level(&from) > level(&to) {
        from = children(db, &from)?;
    }
    while level(&to) > level(&from) {
        to = children(db, &to)?;
    }
    if level(&from) == 0 {
        out.extend(changes(table, &node_rows(&from)?, &node_rows(&to)?));
        return Ok(());
    }
    let child_addresses = |run: &[prolly::Node]| -> Result<Vec<Hash>> {
        let mut addresses = Vec::new();
        for node in run {
            for i in 0..node.count() {
                addresses.push(node.child(i)?);
            }
        }
        Ok(addresses)
    };
    let (a, b) = (child_addresses(&from)?, child_addresses(&to)?);
    let positions: HashMap<Hash, usize> = b.iter().enumerate().map(|(i, h)| (*h, i)).collect();
    let (mut i, mut j) = (0, 0);
    while i < a.len() || j < b.len() {
        if i < a.len() && j < b.len() && a[i] == b[j] {
            i += 1;
            j += 1;
            continue;
        }
        let (mut k, mut l) = (i, b.len());
        while k < a.len() {
            if let Some(&p) = positions.get(&a[k])
                && p >= j
            {
                l = p;
                break;
            }
            k += 1;
        }
        let load = |db: &mut Database, addresses: &[Hash]| -> Result<Vec<prolly::Node>> {
            Ok(addresses.iter().map(|h| prolly::Node::load(db, h)).collect::<store::Result<_>>()?)
        };
        let (stretch_from, stretch_to) = (load(db, &a[i..k])?, load(db, &b[j..l])?);
        diff_runs(db, table, stretch_from, stretch_to, out)?;
        (i, j) = (k, l);
    }
    Ok(())
}

/// children returns the child nodes of a run of internal nodes, in order.
fn children(db: &mut Database, run: &[prolly::Node]) -> Result<Vec<prolly::Node>> {
    let mut out = Vec::new();
    for node in run {
        for i in 0..node.count() {
            out.push(prolly::Node::load(db, &node.child(i)?)?);
        }
    }
    Ok(out)
}

/// node_rows returns the rows of a run of leaf nodes, in order.
fn node_rows(run: &[prolly::Node]) -> Result<Vec<Entry>> {
    let mut rows = Vec::new();
    for node in run {
        for i in 0..node.count() {
            rows.push((node.key(i)?.to_vec(), node.value(i)?.to_vec()));
        }
    }
    Ok(rows)
}

/// tree_value returns the value that a table's primary index holds for a key, if it holds the key.
fn tree_value(db: &mut Database, table: &TableDef, key: &[u8]) -> Result<Option<Vec<u8>>> {
    let mut node = prolly::Node::decode(table.table.primary_index.clone())?;
    loop {
        let (mut low, mut high) = (0, node.count());
        while low < high {
            let middle = (low + high) / 2;
            match table.compare_keys(node.key(middle)?, key) {
                Ordering::Less => low = middle + 1,
                _ => high = middle,
            }
        }
        if low == node.count() {
            return Ok(None);
        }
        if node.is_leaf() {
            let found = table.compare_keys(node.key(low)?, key) == Ordering::Equal;
            return Ok(found.then(|| node.value(low).map(<[u8]>::to_vec)).transpose()?);
        }
        node = prolly::Node::load(db, &node.child(low)?)?;
    }
}

/// value_fields returns the fields of a row's value that hold its columns, which follow the cardinality of a keyless
/// row.
fn value_fields(table: &TableDef, value: &[u8]) -> Vec<Option<Vec<u8>>> {
    let tuple = Tuple(value);
    let offset = table.keyless() as usize;
    (0..table.value_columns.len()).map(|i| tuple.field(i + offset).ok().flatten().map(<[u8]>::to_vec)).collect()
}

/// try_merge merges two sides' changes to a row column by column, as Dolt's valueMerger does: a column that only one
/// side changed takes that side's value, a JSON document that both changed merges key by key when given the database
/// to read it from, and any other column that both changed differently is a conflict, which returns None. A missing
/// value is a deleted row, which merges only when the other side left every column as the base had it.
fn try_merge(
    table: &TableDef,
    mut json: Option<&mut Database>,
    base: Option<&[u8]>,
    left: Option<&[u8]>,
    right: Option<&[u8]>,
) -> Option<Option<Vec<u8>>> {
    if table.keyless() {
        return None;
    }
    let fields = |v: Option<&[u8]>| v.map(|v| value_fields(table, v));
    let (base, left, right) = (fields(base), fields(left), fields(right));
    match (&base, &left, &right) {
        (Some(base), None, Some(kept)) | (Some(base), Some(kept), None) => {
            return (base == kept).then_some(None);
        }
        (_, None, _) | (_, _, None) => return None,
        _ => {}
    }
    let (left, right) = (left.unwrap_or_default(), right.unwrap_or_default());
    let mut merged: Vec<Option<Vec<u8>>> = Vec::with_capacity(left.len());
    for i in 0..left.len() {
        let (l, r) = (&left[i], &right[i]);
        let b = base.as_ref().map(|b| &b[i]);
        let value = if l == r {
            l.clone()
        } else {
            match b {
                Some(b) if l == b => r.clone(),
                Some(b) if r == b => l.clone(),
                Some(b) => merge_json_field(json.as_deref_mut()?, &table.columns[table.value_columns[i]], b, l, r)?,
                None => return None,
            }
        };
        merged.push(value);
    }
    if let Some(db) = json {
        crate::storage::place_adaptive(db, &mut merged, &table.value_encodings(), DEFAULT_TARGET_ROW_SIZE as usize)
            .ok()?;
    }
    Some(Some(prolly::val::build_tuple(&merged.iter().map(Option::as_deref).collect::<Vec<_>>())))
}

/// merge_json_field merges a json or jsonb column's documents that both sides changed, returning the merged field, or
/// None when the column holds something else or the changes conflict.
fn merge_json_field(
    db: &mut Database,
    column: &crate::catalog::table::ColumnDef,
    base: &Option<Vec<u8>>,
    left: &Option<Vec<u8>>,
    right: &Option<Vec<u8>>,
) -> Option<Option<Vec<u8>>> {
    if !matches!(column.ty.oid, crate::oid::JSON | crate::oid::JSONB) {
        return None;
    }
    let document = |field: &Option<Vec<u8>>| -> Option<Json> {
        match crate::storage::decode_field(db, field.as_deref(), column.encoding, column.ty).ok()? {
            Value::Json(text) => crate::json::parse(&text, false).ok(),
            Value::Jsonb(json) => Some(*json),
            _ => None,
        }
    };
    let merged = merge_json(&document(base)?, &document(left)?, &document(right)?)?;
    let value = match column.ty.oid {
        crate::oid::JSONB => Value::Jsonb(Box::new(merged)),
        _ => Value::Json(merged.compact()),
    };
    crate::storage::encode_field(&value, column.encoding, column.ty).ok()
}

/// merge_json three-way merges JSON documents as Dolt's ThreeWayJsonDiffer does: objects merge key by key, and any
/// other value that both sides changed differently is a conflict, which returns None.
fn merge_json(base: &Json, left: &Json, right: &Json) -> Option<Json> {
    if left == right || right == base {
        return Some(left.clone());
    }
    if left == base {
        return Some(right.clone());
    }
    let (Json::Object(b), Json::Object(l), Json::Object(r)) = (base, left, right) else { return None };
    let get = |items: &[(String, Json)], key: &str| items.iter().find(|(k, _)| k == key).map(|(_, v)| v.clone());
    let keys = union([l, r, b].map(|items| items.iter().map(|(k, _)| k.clone()).collect()));
    let mut merged = Vec::with_capacity(keys.len());
    for key in keys {
        let (bv, lv, rv) = (get(b, &key), get(l, &key), get(r, &key));
        let value = match pick(bv.as_ref(), lv.as_ref(), rv.as_ref()) {
            Some(value) => value,
            None => Some(merge_json(&bv?, &lv?, &rv?)?),
        };
        if let Some(value) = value {
            merged.push((key, value));
        }
    }
    Some(Json::Object(merged))
}

/// Merged is the outcome of merging a table's rows: the edits to make to our rows, and the conflicts and constraint
/// violations to record.
#[derive(Default)]
struct Merged {
    edits: Vec<Edit>,
    artifacts: Vec<Artifact>,
    new_artifacts: bool,
}

/// UniqueIndex tracks the rows of a unique index during a merge: each row's indexed values by its key.
struct UniqueIndex<'a> {
    index: &'a IndexDef,
    info: Vec<u8>,
    rows: HashMap<Vec<Option<Vec<u8>>>, Vec<Vec<u8>>>,
}

/// json_string writes a string as Go's JSON encoder does, escaping HTML characters.
fn json_string(text: &str) -> String {
    let mut out = String::from("\"");
    for c in text.chars() {
        match c {
            '"' => out.push_str("\\\""),
            '\\' => out.push_str("\\\\"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            '<' => out.push_str("\\u003c"),
            '>' => out.push_str("\\u003e"),
            '&' => out.push_str("\\u0026"),
            '\u{2028}' => out.push_str("\\u2028"),
            '\u{2029}' => out.push_str("\\u2029"),
            c if (c as u32) < 0x20 => out.push_str(&format!("\\u{:04x}", c as u32)),
            c => out.push(c),
        }
    }
    out.push('"');
    out
}

/// json_strings writes a list of strings as Go's JSON encoder does.
pub fn json_strings(items: &[String]) -> String {
    format!("[{}]", items.iter().map(|s| json_string(s)).collect::<Vec<_>>().join(","))
}

/// RowMerger merges one table's rows, recording conflicts and unique index violations as it goes.
struct RowMerger<'a> {
    table: &'a TableDef,
    commits: Commits,
    merged: Merged,
    uniques: Vec<UniqueIndex<'a>>,
    /// The current value of each row whose value the merge changed or validated, by key.
    current: BTreeMap<Vec<u8>, Option<Vec<u8>>>,
    ours: &'a [Entry],
    db: &'a mut Database,
    /// Whether JSON documents that both sides changed merge key by key, which dolt_dont_merge_json turns off.
    merge_json: bool,
}

impl<'a> RowMerger<'a> {
    /// value_of returns a row's current value: the merge's value for it, or our value.
    fn value_of(&self, key: &[u8]) -> Option<Vec<u8>> {
        if let Some(value) = self.current.get(key) {
            return value.clone();
        }
        self.ours.binary_search_by(|(k, _)| self.table.compare_keys(k, key)).ok().map(|i| self.ours[i].1.clone())
    }

    /// violation records that a row violates a unique index, replacing an earlier record of the same violation.
    fn violation(&mut self, key: &[u8], value: &[u8], info: &[u8]) {
        self.record(key, value, info, artifacts::UNIQUE, self.commits.theirs);
    }

    /// record records a constraint violation of a row, replacing an earlier record of the same violation, as Dolt's
    /// ReplaceConstraintViolation does.
    fn record(&mut self, key: &[u8], value: &[u8], info: &[u8], kind: u8, rootish: Hash) {
        let meta = artifacts::violation_meta(info, value);
        self.merged.artifacts.retain(|a| !(a.key == key && a.kind == kind && a.meta == meta));
        self.merged.artifacts.push(Artifact {
            key: key.to_vec(),
            rootish,
            kind,
            info_hash: artifacts::info_hash(info),
            meta,
        });
        self.merged.new_artifacts = true;
    }

    /// null_violation records the NOT NULL columns that a row leaves NULL, as Dolt's nullValidator does, and reports
    /// whether it found any.
    fn null_violation(&mut self, key: &[u8], value: &[u8], rootish: Hash) -> bool {
        let fields = value_fields(self.table, value);
        let columns: Vec<String> = self
            .table
            .value_columns
            .iter()
            .zip(&fields)
            .filter(|(c, field)| !self.table.columns[**c].nullable && field.is_none())
            .map(|(c, _)| self.table.columns[*c].name.clone())
            .collect();
        if columns.is_empty() {
            return false;
        }
        let info = format!("{{\"Columns\":{}}}", json_strings(&columns));
        self.record(key, value, info.as_bytes(), artifacts::NOT_NULL, rootish);
        true
    }

    /// remove_unique removes a row from the unique indexes.
    fn remove_unique(&mut self, key: &[u8], value: &[u8]) {
        for i in 0..self.uniques.len() {
            let Some(values) = indexed(self.table, self.uniques[i].index, key, value) else { continue };
            if let Some(keys) = self.uniques[i].rows.get_mut(&values) {
                keys.retain(|k| k != key);
            }
        }
    }

    /// check_unique records the unique index violations of a row that the merge leaves with a value, along with the
    /// rows it collides with, and adds it to the unique indexes, as Dolt's uniqValidator does.
    fn check_unique(&mut self, key: &[u8], value: &[u8], previous: Option<&[u8]>) {
        if let Some(previous) = previous {
            self.remove_unique(key, previous);
        }
        for i in 0..self.uniques.len() {
            let Some(values) = indexed(self.table, self.uniques[i].index, key, value) else { continue };
            let others: Vec<Vec<u8>> = self.uniques[i]
                .rows
                .get(&values)
                .map(|keys| keys.iter().filter(|k| *k != key).cloned().collect())
                .unwrap_or_default();
            let info = self.uniques[i].info.clone();
            let mut collided = false;
            for other in others {
                if let Some(other_value) = self.value_of(&other) {
                    collided = true;
                    self.violation(&other, &other_value, &info);
                }
            }
            if collided {
                self.violation(key, value, &info);
            }
            self.uniques[i].rows.entry(values).or_default().push(key.to_vec());
        }
    }

    /// clear_unique removes the unique index violations of a row that the merge deletes, and those of the rows its
    /// last value collided with.
    fn clear_unique(&mut self, key: &[u8], previous: &[u8]) {
        self.remove_unique(key, previous);
        let before = self.merged.artifacts.len();
        self.merged.artifacts.retain(|a| !(a.key == key && a.kind == artifacts::UNIQUE));
        if self.merged.artifacts.len() == before {
            return;
        }
        for i in 0..self.uniques.len() {
            let Some(values) = indexed(self.table, self.uniques[i].index, key, previous) else { continue };
            let others = self.uniques[i].rows.get(&values).cloned().unwrap_or_default();
            self.merged.artifacts.retain(|a| !(others.contains(&a.key) && a.kind == artifacts::UNIQUE));
        }
    }

    /// conflict records that both sides changed a row differently.
    fn conflict(&mut self, key: &[u8]) {
        self.merged.artifacts.push(Artifact {
            key: key.to_vec(),
            rootish: self.commits.theirs,
            kind: artifacts::CONFLICT,
            info_hash: vec![0; Hash::LEN],
            meta: artifacts::conflict_meta(self.commits.base),
        });
        self.merged.new_artifacts = true;
    }

    /// set records the merge's value for a row.
    fn set(&mut self, key: &[u8], value: Option<Vec<u8>>) {
        self.current.insert(key.to_vec(), value.clone());
        self.merged.edits.push((key.to_vec(), value));
    }

    /// merge applies the three-way changes of our side and their side, in key order, as Dolt's three-way differ
    /// and its consumers do.
    fn merge(&mut self, left: &[Change], right: &[Change]) {
        let keyless = self.table.keyless();
        let (mut i, mut j) = (0, 0);
        while i < left.len() || j < right.len() {
            let ordering = match (left.get(i), right.get(j)) {
                (Some(l), Some(r)) => self.table.compare_keys(&l.key, &r.key),
                (Some(_), None) => Ordering::Less,
                _ => Ordering::Greater,
            };
            match ordering {
                Ordering::Less => {
                    let l = left[i].clone();
                    if let Some(to) = &l.to {
                        self.check_unique(&l.key, to, None);
                        if self.null_violation(&l.key, to, self.commits.ours) {
                            self.set(&l.key, None);
                        }
                    }
                    i += 1;
                }
                Ordering::Greater => {
                    let r = right[j].clone();
                    match &r.to {
                        Some(to) => {
                            let previous = self.value_of(&r.key);
                            self.check_unique(&r.key, to, previous.as_deref());
                            if self.null_violation(&r.key, to, self.commits.theirs) {
                                j += 1;
                                continue;
                            }
                        }
                        None => {
                            if let Some(from) = &r.from {
                                self.clear_unique(&r.key, from);
                            }
                        }
                    }
                    self.set(&r.key, r.to.clone());
                    j += 1;
                }
                Ordering::Equal => {
                    let (l, r) = (left[i].clone(), right[j].clone());
                    i += 1;
                    j += 1;
                    let convergent = l.to == r.to;
                    if convergent {
                        if keyless {
                            self.conflict(&l.key);
                        }
                        continue;
                    }
                    let json = self.merge_json.then_some(&mut *self.db);
                    match try_merge(self.table, json, l.from.as_deref(), l.to.as_deref(), r.to.as_deref()) {
                        Some(None) => self.set(&l.key, None),
                        Some(Some(merged)) => {
                            self.check_unique(&l.key, &merged, l.to.as_deref());
                            let violated = self.null_violation(&l.key, &merged, self.commits.ours);
                            self.set(&l.key, (!violated).then_some(merged));
                        }
                        None => self.conflict(&l.key),
                    }
                }
            }
        }
    }
}

/// apply makes row edits to a table, keeping its secondary indexes in step, and returns the stored table.
fn apply(ctx: &mut Ctx<'_>, table: &TableDef, ours: &[Entry], mut edits: Vec<Edit>) -> Result<doltdb::table::Table> {
    let mut stored = table.table.clone();
    if edits.is_empty() {
        return Ok(stored);
    }
    let rules = ctx.index_rules(table)?;
    edits.sort_by(|a, b| table.compare_keys(&a.0, &b.0));
    let mut deduped: Vec<Edit> = Vec::with_capacity(edits.len());
    for edit in edits {
        match deduped.last_mut() {
            Some(last) if table.compare_keys(&last.0, &edit.0) == Ordering::Equal => *last = edit,
            _ => deduped.push(edit),
        }
    }
    let mut index_edits: Vec<Vec<Edit>> = vec![Vec::new(); table.indexes.len()];
    for (key, value) in &deduped {
        let old = ours.binary_search_by(|(k, _)| table.compare_keys(k, key)).ok().map(|i| &ours[i].1);
        for (stored, added) in [(old, false), (value.as_ref(), true)] {
            let Some(stored) = stored else { continue };
            let (row, _) = table.decode_row(ctx.db, key, stored)?;
            let (row, held) = rules.indexed(ctx, &row)?;
            for (i, index) in table.indexes.iter().enumerate().filter(|(i, _)| held[*i]) {
                let value = added.then(|| prolly::val::build_tuple(&[]));
                index_edits[i].push((table.index_key(ctx.db, index, &row, key)?, value));
            }
        }
    }
    let db = &mut *ctx.db;
    stored.edit_rows(
        db,
        deduped,
        &|a, b| table.compare_keys(a, b),
        (&table.key_encodings(), &table.value_encodings()),
    )?;
    for (index, mut edits) in table.indexes.iter().zip(index_edits) {
        if edits.is_empty() {
            continue;
        }
        if let Some(distance) = index.vector {
            let rebuilt = TableDef { table: stored.clone(), ..table.clone() };
            let root = rebuilt.write_vector_index(db, index, distance)?;
            stored.put_index(db, &index.name, Some(root))?;
            continue;
        }
        let compare = |a: &[u8], b: &[u8]| table.compare_index_keys(index, a, b);
        edits.sort_by(|a, b| compare(&a.0, &b.0));
        let mut merged: Vec<Edit> = Vec::with_capacity(edits.len());
        for edit in edits {
            match merged.last_mut() {
                Some(last) if compare(&last.0, &edit.0) == Ordering::Equal => *last = edit,
                _ => merged.push(edit),
            }
        }
        stored.edit_index(db, &index.name, index.root, merged, &compare, &table.index_encodings(index))?;
    }
    Ok(stored)
}

/// unique_indexes returns the unique indexes of a table, each with the information that its violations record and
/// the rows it holds.
fn unique_indexes<'a>(table: &'a TableDef, rows: &[Entry]) -> Vec<UniqueIndex<'a>> {
    let mut out = Vec::new();
    let plain = |i: &IndexDef| i.predicate.is_empty() && i.columns.iter().all(|&c| c < HIDDEN_BASE);
    for index in table.indexes.iter().filter(|i| i.unique && plain(i)) {
        let columns: Vec<String> = index.columns.iter().map(|&c| table.columns[c].name.clone()).collect();
        let info = format!("{{\"Columns\":{},\"Name\":{}}}", json_strings(&columns), json_string(&index.name));
        out.push(UniqueIndex { index, info: info.into_bytes(), rows: HashMap::new() });
    }
    for unique in &mut out {
        for (key, value) in rows {
            if let Some(values) = indexed(table, unique.index, key, value) {
                unique.rows.entry(values).or_default().push(key.clone());
            }
        }
    }
    out
}

/// indexed returns the values of a row that a unique index holds, or None when one is NULL, since NULLs never collide.
fn indexed(table: &TableDef, index: &IndexDef, key: &[u8], value: &[u8]) -> Option<Vec<Option<Vec<u8>>>> {
    let (key, values) = (Tuple(key), value_fields(table, value));
    let mut out = Vec::with_capacity(index.columns.len());
    for &c in &index.columns {
        let field = match table.key_columns.iter().position(|&k| k == c) {
            Some(i) => key.field(i).ok().flatten().map(<[u8]>::to_vec),
            None => table.value_columns.iter().position(|&v| v == c).and_then(|i| values[i].clone()),
        };
        out.push(Some(field?));
    }
    Some(out)
}

/// convert rewrites rows of a table as rows of another version of it, matching columns by tag and giving a column
/// the other version lacks its default, or NULL.
fn convert(ctx: &mut Ctx<'_>, from: &TableDef, to: &TableDef, rows: &[Entry]) -> Result<Vec<Entry>> {
    let mut defaults = Vec::with_capacity(to.columns.len());
    for column in &to.columns {
        let expr =
            match from.columns.iter().any(|c| c.tag == column.tag) || column.default.is_empty() || column.generated {
                true => None,
                false => {
                    let node = crate::dml::parse_expression(&column.default)?;
                    let bound = crate::expr::Binder::new(ctx, crate::expr::Scope::default()).bind(&node)?;
                    Some(crate::expr::assign(bound, column.ty, &column.name, -1)?.0)
                }
            };
        defaults.push(expr);
    }
    let mut out = Vec::with_capacity(rows.len());
    for (key, value) in rows {
        let (row, count) = from.decode_row(ctx.db, key, value)?;
        let mut converted = Vec::with_capacity(to.columns.len());
        for (i, column) in to.columns.iter().enumerate() {
            converted.push(match from.columns.iter().position(|c| c.tag == column.tag) {
                Some(p) if from.columns[p].ty == column.ty => row[p].clone(),
                Some(p) => crate::cast::cast_value(row[p].clone(), column.ty, false).unwrap_or(Value::Null),
                None => match &defaults[i] {
                    Some(expr) => expr.eval(ctx, &[])?,
                    None => Value::Null,
                },
            });
        }
        let (key, mut value) = to.encode_row(ctx.db, &converted)?;
        if to.keyless() && count != 1 {
            value[..8].copy_from_slice(&count.to_le_bytes());
        }
        out.push((key, value));
    }
    out.sort_by(|a, b| to.compare_keys(&a.0, &b.0));
    Ok(out)
}

/// TableOutcome is what merging a table does to our root: keep our table, put a merged table, remove the table, or
/// keep ours with a schema conflict.
enum TableOutcome {
    Keep,
    Put(Box<doltdb::table::Table>, bool),
    Remove,
    SchemaConflict,
}

/// merge_rows merges the rows of a table whose three versions the merge has as rows of one schema.
fn merge_rows(
    ctx: &mut Ctx<'_>,
    table: &TableDef,
    ours: &[Entry],
    theirs: &[Entry],
    base: &[Entry],
    commits: Commits,
    brought: bool,
) -> Result<TableOutcome> {
    let (left, right) = (changes(table, base, ours), changes(table, base, theirs));
    merge_changes(ctx, table, ours, &left, &right, commits, brought)
}

/// merge_changes merges both sides' changes to the rows of a table, given our rows, or at least our rows of every key
/// that either side changed when the table has no unique indexes.
fn merge_changes(
    ctx: &mut Ctx<'_>,
    table: &TableDef,
    ours: &[Entry],
    left: &[Change],
    right: &[Change],
    commits: Commits,
    brought: bool,
) -> Result<TableOutcome> {
    let merge_json = !ctx.session.setting_on("dolt_dont_merge_json");
    let mut merger = RowMerger {
        table,
        commits,
        merged: Merged { artifacts: artifacts::read(ctx.db, table)?, new_artifacts: brought, ..Merged::default() },
        uniques: unique_indexes(table, ours),
        current: BTreeMap::new(),
        ours,
        db: ctx.db,
        merge_json,
    };
    merger.merge(left, right);
    let Merged { edits, artifacts: found, new_artifacts } = merger.merged;
    let mut stored = apply(ctx, table, ours, edits)?;
    stored.artifacts = artifacts::write(ctx.db, table, found)?;
    Ok(TableOutcome::Put(Box::new(stored), new_artifacts))
}

/// merge_table merges a table's three versions, taking the shortcuts Dolt's MaybeShortCircuit takes when a side left
/// the table as the base had it.
fn merge_table(
    ctx: &mut Ctx<'_>,
    name: &Name,
    ours: Option<(Hash, TableDef)>,
    theirs: Option<(Hash, TableDef)>,
    base: Option<(Hash, TableDef)>,
    commits: Commits,
) -> Result<TableOutcome> {
    let address = |t: &Option<(Hash, TableDef)>| t.as_ref().map(|(a, _)| *a);
    let (o, t, a) = (address(&ours), address(&theirs), address(&base));
    let keyless = ours.as_ref().is_some_and(|(_, d)| d.keyless());
    if o.is_some() && o == t && (o == a || !keyless) {
        return Ok(TableOutcome::Keep);
    }
    let schema = |t: &Option<(Hash, TableDef)>| t.as_ref().map(|(_, d)| d.table.schema);
    match (&ours, &theirs, &base) {
        (Some(_), Some(_), None) if schema(&ours) != schema(&theirs) => {
            return Err(error(format!("table with same name '{}' added in 2 commits can't be merged", name.1)));
        }
        (Some(_), None, None) => return Ok(TableOutcome::Keep),
        (None, Some((_, def)), None) => return Ok(TableOutcome::Put(Box::new(def.table.clone()), false)),
        (None, None, Some(_)) => return Ok(TableOutcome::Remove),
        (Some(_), None, Some(_)) | (None, Some(_), Some(_)) => {
            let child = o.or(t);
            if child != a {
                return Ok(TableOutcome::SchemaConflict);
            }
            return Ok(TableOutcome::Remove);
        }
        _ => {}
    }
    if t == a {
        return Ok(TableOutcome::Keep);
    }
    if o == a {
        let (_, def) = theirs.as_ref().ok_or_else(|| error("missing table"))?;
        return Ok(TableOutcome::Put(Box::new(def.table.clone()), false));
    }
    let (Some((_, mut ours)), Some((_, theirs))) = (ours, theirs) else { return Ok(TableOutcome::Keep) };
    let brought = carry_artifacts(ctx.db, &mut ours, &theirs, base.as_ref().map(|(_, b)| b))?;
    let base = match base {
        Some((_, base)) => base,
        None => {
            let mut empty = ours.clone();
            empty.table.primary_index = doltdb::table::empty_rows();
            empty
        }
    };
    if ours.table.schema == theirs.table.schema
        && base.table.schema == ours.table.schema
        && unique_indexes(&ours, &[]).is_empty()
    {
        let left = changes_between(ctx.db, &ours, &base, &ours)?;
        let right = changes_between(ctx.db, &ours, &base, &theirs)?;
        let mut keys: Vec<&[u8]> = left.iter().chain(&right).map(|c| c.key.as_slice()).collect();
        keys.sort_by(|a, b| ours.compare_keys(a, b));
        keys.dedup_by(|a, b| ours.compare_keys(a, b) == Ordering::Equal);
        let mut our_rows = Vec::with_capacity(keys.len());
        for key in keys {
            if let Some(value) = tree_value(ctx.db, &ours, key)? {
                our_rows.push((key.to_vec(), value));
            }
        }
        return merge_changes(ctx, &ours, &our_rows, &left, &right, commits, brought);
    }
    let (our_rows, their_rows, base_rows) =
        (entries(ctx.db, &ours)?, entries(ctx.db, &theirs)?, entries(ctx.db, &base)?);
    if ours.table.schema == theirs.table.schema {
        let base_rows = match base.table.schema == ours.table.schema {
            true => base_rows,
            false => convert(ctx, &base, &ours, &base_rows)?,
        };
        return merge_rows(ctx, &ours, &our_rows, &their_rows, &base_rows, commits, brought);
    }
    if ours.table.schema == base.table.schema {
        let our_rows = convert(ctx, &ours, &theirs, &our_rows)?;
        let base_rows = convert(ctx, &base, &theirs, &base_rows)?;
        let mut target = theirs.clone();
        target.table.artifacts = ours.table.artifacts.clone();
        return merge_rows(ctx, &target, &their_rows, &our_rows, &base_rows, commits, brought);
    }
    if theirs.table.schema == base.table.schema {
        let their_rows = convert(ctx, &theirs, &ours, &their_rows)?;
        let base_rows = convert(ctx, &base, &ours, &base_rows)?;
        return merge_rows(ctx, &ours, &our_rows, &their_rows, &base_rows, commits, brought);
    }
    let Some(mut merged) = merge_schemas(&base, &ours, &theirs) else { return Ok(TableOutcome::SchemaConflict) };
    let message = merged.schema_message()?;
    let our_converted = convert(ctx, &ours, &merged, &our_rows)?;
    if same_layout(&merged, &ours) {
        merged.table.schema = ctx.db.write_value(message)?;
    } else {
        let (_, mut stored) = doltdb::table::Table::create(ctx.db, message)?;
        for index in &mut merged.indexes {
            index.root = index.empty_root(ctx.db)?;
            stored.put_index(ctx.db, &index.name, Some(index.root))?;
        }
        stored.artifacts = ours.table.artifacts.clone();
        stored.auto_increment = ours.table.auto_increment;
        merged.table = stored;
        let inserts = our_converted.iter().map(|(key, value)| (key.clone(), Some(value.clone()))).collect();
        merged.table = apply(ctx, &merged, &[], inserts)?;
    }
    let their_rows = convert(ctx, &theirs, &merged, &their_rows)?;
    let base_rows = convert(ctx, &base, &merged, &base_rows)?;
    merge_rows(ctx, &merged, &our_converted, &their_rows, &base_rows, commits, brought)
}

/// pick returns the version of a part of a schema that a three-way merge keeps: the one side that changed it, or
/// either when both made the same change, or None when both changed it differently.
fn pick<T: PartialEq + Clone>(base: Option<&T>, ours: Option<&T>, theirs: Option<&T>) -> Option<Option<T>> {
    match (ours == theirs, ours == base, theirs == base) {
        (true, ..) | (_, _, true) => Some(ours.cloned()),
        (_, true, _) => Some(theirs.cloned()),
        _ => None,
    }
}

/// IndexShape is an index as a merge compares it: its definition with its columns by tag, without its rows.
type IndexShape = (String, Vec<u64>, bool, Vec<bool>, Vec<bool>, Vec<String>, String);

/// index_shape returns an index's definition with its columns by tag, or None for an index of an expression.
fn index_shape(table: &TableDef, index: &IndexDef) -> Option<IndexShape> {
    let tags = index.columns.iter().map(|&c| table.columns.get(c).map(|c| c.tag)).collect::<Option<Vec<_>>>()?;
    let parts = (index.descending.clone(), index.nulls_last.clone(), index.op_classes.clone());
    Some((index.name.clone(), tags, index.unique, parts.0, parts.1, parts.2, index.predicate.clone()))
}

/// merge_schemas merges the schemas of a table that both sides changed, column by column and index by index as Dolt's
/// SchemaMerge does: each part that one side changed takes that side's version, and parts that both changed must
/// match. It returns None for a conflict, or for changes it cannot merge, such as to the primary key.
fn merge_schemas(base: &TableDef, ours: &TableDef, theirs: &TableDef) -> Option<TableDef> {
    let key_tags = |t: &TableDef| t.key_columns.iter().map(|&k| t.columns[k].tag).collect::<Vec<_>>();
    if key_tags(base) != key_tags(ours) || key_tags(ours) != key_tags(theirs) || ours.primary != theirs.primary {
        return None;
    }
    if !(base.hidden.is_empty() && ours.hidden.is_empty() && theirs.hidden.is_empty()) {
        return None;
    }
    let column = |t: &TableDef, tag: u64| t.columns.iter().find(|c| c.tag == tag).cloned();
    let tags = union([ours, theirs, base].map(|t| t.columns.iter().map(|c| c.tag).collect()));
    let mut columns = Vec::new();
    for tag in tags {
        let (b, o, t) = (column(base, tag), column(ours, tag), column(theirs, tag));
        if let Some(column) = pick(b.as_ref(), o.as_ref(), t.as_ref())? {
            columns.push(column);
        }
    }
    let mut names: Vec<&str> = columns.iter().map(|c| c.name.as_str()).collect();
    names.sort_unstable();
    if names.windows(2).any(|w| w[0] == w[1]) {
        return None;
    }
    let mut merged = ours.clone();
    let position = |tag: u64| columns.iter().position(|c| c.tag == tag);
    merged.key_columns = key_tags(ours).into_iter().map(position).collect::<Option<_>>()?;
    merged.value_columns = (0..columns.len()).filter(|i| !merged.key_columns.contains(i)).collect();
    let check = |t: &TableDef, name: &str| t.checks.iter().find(|c| c.name == name).cloned();
    let check_names = union([ours, theirs, base].map(|t| t.checks.iter().map(|c| c.name.clone()).collect()));
    merged.checks = Vec::new();
    for name in check_names {
        let (b, o, t) = (check(base, &name), check(ours, &name), check(theirs, &name));
        merged.checks.extend(pick(b.as_ref(), o.as_ref(), t.as_ref())?);
    }
    let shapes = |t: &TableDef| t.indexes.iter().map(|i| index_shape(t, i)).collect::<Option<Vec<_>>>();
    let (base_shapes, our_shapes, their_shapes) = (shapes(base)?, shapes(ours)?, shapes(theirs)?);
    let index_names = union([ours, theirs, base].map(|t| t.indexes.iter().map(|i| i.name.clone()).collect()));
    merged.indexes = Vec::new();
    for name in index_names {
        let find = |shapes: &[IndexShape]| shapes.iter().find(|s| s.0 == name).cloned();
        let (b, o, t) = (find(&base_shapes), find(&our_shapes), find(&their_shapes));
        let Some(shape) = pick(b.as_ref(), o.as_ref(), t.as_ref())? else { continue };
        let source = if o.as_ref() == Some(&shape) { ours } else { theirs };
        let mut index = source.indexes.iter().find(|i| i.name == name)?.clone();
        index.columns = shape.1.iter().map(|&tag| position(tag)).collect::<Option<_>>()?;
        merged.indexes.push(index);
    }
    merged.columns = columns;
    merged.comment = pick(Some(&base.comment), Some(&ours.comment), Some(&theirs.comment))??;
    Some(merged)
}

/// union returns the items of lists in order, each once.
fn union<T: PartialEq>(lists: [Vec<T>; 3]) -> Vec<T> {
    let mut out = Vec::new();
    for item in lists.into_iter().flatten() {
        if !out.contains(&item) {
            out.push(item);
        }
    }
    out
}

/// same_layout reports whether a merged table stores its rows and indexes as our table does, so that it can keep our
/// table's storage under its new schema.
fn same_layout(merged: &TableDef, ours: &TableDef) -> bool {
    let layout = |t: &TableDef| {
        let columns: Vec<_> = t.columns.iter().map(|c| (c.tag, c.encoding, c.ty)).collect();
        let indexes: Vec<_> = t.indexes.iter().map(|i| (i.name.clone(), i.columns.clone(), i.unique)).collect();
        (columns, t.key_columns.clone(), t.value_columns.clone(), indexes)
    };
    layout(merged) == layout(ours)
}

/// carry_artifacts adds to our table the conflicts and constraint violations that their side added since the merge
/// base, and drops the ones their side resolved, as Dolt's mergeTableArtifacts does, reporting whether it added any.
fn carry_artifacts(db: &mut Database, ours: &mut TableDef, theirs: &TableDef, base: Option<&TableDef>) -> Result<bool> {
    let (mut mine, their) = (artifacts::read(db, ours)?, artifacts::read(db, theirs)?);
    let ancestor = match base {
        Some(base) => artifacts::read(db, base)?,
        None => Vec::new(),
    };
    let added: Vec<Artifact> = their.iter().filter(|a| !ancestor.contains(a) && !mine.contains(a)).cloned().collect();
    let count = mine.len();
    mine.retain(|a| !ancestor.contains(a) || their.contains(a));
    if added.is_empty() && mine.len() == count {
        return Ok(false);
    }
    let brought = !added.is_empty();
    mine.extend(added);
    ours.table.artifacts = artifacts::write(db, ours, mine)?;
    Ok(brought)
}

/// merge_roots merges their root into ours, given the merge base's root, as Dolt's MergeRoots does, without the
/// foreign key checks, which `check_foreign_keys` makes afterwards.
pub fn merge_roots(ctx: &mut Ctx<'_>, ours: &Root, theirs: &Root, base: &Root, commits: Commits) -> Result<Outcome> {
    let mut merged = ours.clone();
    let mut names: Vec<Name> = crate::dolt::procedures::table_map(ctx.db, ours)?.into_keys().collect();
    names.extend(crate::dolt::procedures::table_map(ctx.db, theirs)?.into_keys());
    names.sort();
    names.dedup();
    let mut outcome = Outcome { root: merged.clone(), artifacts: false, schema_conflicts: Vec::new() };
    for name in names {
        let o = load(ctx.db, ours, &name)?;
        let t = load(ctx.db, theirs, &name)?;
        let a = load(ctx.db, base, &name)?;
        match merge_table(ctx, &name, o, t, a, commits)? {
            TableOutcome::Keep => {}
            TableOutcome::Put(table, artifacts) => {
                if !merged.schemas.iter().any(|s| s == name.0.as_bytes()) {
                    merged.schemas.push(name.0.as_bytes().to_vec());
                    merged.schemas.sort();
                }
                let address = table.write(ctx.db)?;
                merged.put_table(ctx.db, &name.0, &name.1, Some(address))?;
                outcome.artifacts |= artifacts;
            }
            TableOutcome::Remove => merged.put_table(ctx.db, &name.0, &name.1, None)?,
            TableOutcome::SchemaConflict => {
                outcome.artifacts = true;
                outcome.schema_conflicts.push(name);
            }
        }
    }
    for schema in &theirs.schemas {
        if !base.schemas.contains(schema) && !merged.schemas.contains(schema) {
            merged.schemas.push(schema.clone());
        }
    }
    merged.schemas.sort();
    if merged.foreign_keys == base.foreign_keys {
        merged.foreign_keys = theirs.foreign_keys.clone();
    }
    if crate::dolt::objmerge::merge_collections(ctx, &mut merged, (ours, theirs, base), commits.theirs)? > 0 {
        outcome.artifacts = true;
    }
    outcome.root = merged;
    Ok(outcome)
}

/// has_artifacts reports whether any table of a root has conflicts or constraint violations.
pub fn has_artifacts(db: &mut Database, root: &Root) -> Result<bool> {
    for (_, address) in crate::dolt::procedures::table_map(db, root)? {
        let table = doltdb::table::Table::decode(&read(db, &address)?)?;
        if table.artifacts.iter().any(|&b| b != 0) {
            return Ok(true);
        }
    }
    Ok(false)
}

/// reduced_rule returns the name Dolt records for a foreign key action, which treats every action that refuses a
/// change as RESTRICT.
fn reduced_rule(rule: crate::foreign::Rule) -> &'static str {
    match rule {
        crate::foreign::Rule::Cascade => "CASCADE",
        crate::foreign::Rule::SetNull => "SET NULL",
        crate::foreign::Rule::SetDefault => "SET DEFAULT",
        _ => "RESTRICT",
    }
}

/// foreign_key_info returns the information that a foreign key's violations record, as Dolt's foreignKeyCVJson
/// writes it.
fn foreign_key_info(fk: &crate::foreign::ForeignKeyDef) -> Vec<u8> {
    format!(
        "{{\"Columns\":{},\"ForeignKey\":{},\"Index\":{},\"OnDelete\":{},\"OnUpdate\":{},\"ReferencedColumns\":{},\
         \"ReferencedIndex\":{},\"ReferencedTable\":{},\"Table\":{}}}",
        json_strings(&fk.child_columns),
        json_string(&fk.name),
        json_string(&fk.child_index),
        json_string(reduced_rule(fk.on_delete)),
        json_string(reduced_rule(fk.on_update)),
        json_strings(&fk.parent_columns),
        json_string(&fk.parent_index),
        json_string(&fk.parent_table),
        json_string(&fk.child_table),
    )
    .into_bytes()
}

/// referenced returns the values of a row's columns that a foreign key compares, written as the referenced columns'
/// type shows them, or None when one is NULL, since a NULL never violates a foreign key.
fn referenced(
    db: &mut Database,
    table: &TableDef,
    key: &[u8],
    value: &[u8],
    columns: &[String],
    types: &[crate::catalog::ColumnType],
) -> Result<Option<Vec<String>>> {
    let (row, _) = table.decode_row(db, key, value)?;
    let mut out = Vec::with_capacity(columns.len());
    for (name, ty) in columns.iter().zip(types) {
        let Some(position) = table.columns.iter().position(|c| &c.name == name) else { return Ok(None) };
        let value = row[position].clone();
        if value.is_null() {
            return Ok(None);
        }
        let value = crate::cast::cast_value(value, *ty, false).unwrap_or(Value::Null);
        out.push(value.output().unwrap_or_default());
    }
    Ok(Some(out))
}

/// check_foreign_keys records the rows of a merged root that break a foreign key because the merge added or changed
/// them, or removed or changed the rows they refer to, as Dolt's AddForeignKeyViolations does, and reports whether
/// it recorded any.
pub fn check_foreign_keys(ctx: &mut Ctx<'_>, merged: &mut Root, base: &Root, theirs: Hash) -> Result<bool> {
    let mut found = false;
    for fk in crate::foreign::load(ctx.db, merged)? {
        let child_name = (fk.child_schema.clone(), fk.child_table.clone());
        let parent_name = (fk.parent_schema.clone(), fk.parent_table.clone());
        let (Some((_, child)), Some((_, parent))) =
            (load(ctx.db, merged, &child_name)?, load(ctx.db, merged, &parent_name)?)
        else {
            continue;
        };
        let types: Vec<crate::catalog::ColumnType> = fk
            .parent_columns
            .iter()
            .filter_map(|n| parent.columns.iter().find(|c| &c.name == n).map(|c| c.ty))
            .collect();
        if types.len() != fk.parent_columns.len() {
            continue;
        }
        let parent_rows = entries(ctx.db, &parent)?;
        let mut parents = std::collections::HashSet::new();
        for (key, value) in &parent_rows {
            if let Some(values) = referenced(ctx.db, &parent, key, value, &fk.parent_columns, &types)? {
                parents.insert(values);
            }
        }
        let child_rows = entries(ctx.db, &child)?;
        let pre_child = load(ctx.db, base, &child_name)?;
        let pre_child_rows = match &pre_child {
            Some((_, table)) if table.table.schema == child.table.schema => entries(ctx.db, table)?,
            _ => Vec::new(),
        };
        let mut violating: Vec<Entry> = Vec::new();
        for change in changes(&child, &pre_child_rows, &child_rows) {
            let Some(to) = change.to else { continue };
            if let Some(values) = referenced(ctx.db, &child, &change.key, &to, &fk.child_columns, &types)?
                && !parents.contains(&values)
            {
                violating.push((change.key, to));
            }
        }
        let pre_parent = load(ctx.db, base, &parent_name)?;
        if let Some((_, pre_parent)) = pre_parent.filter(|(_, t)| t.table.schema == parent.table.schema) {
            let pre_parent_rows = entries(ctx.db, &pre_parent)?;
            let mut gone = std::collections::HashSet::new();
            for change in changes(&parent, &pre_parent_rows, &parent_rows) {
                let Some(from) = change.from else { continue };
                if let Some(values) = referenced(ctx.db, &parent, &change.key, &from, &fk.parent_columns, &types)?
                    && !parents.contains(&values)
                {
                    gone.insert(values);
                }
            }
            if !gone.is_empty() {
                for (key, value) in &child_rows {
                    if let Some(values) = referenced(ctx.db, &child, key, value, &fk.child_columns, &types)?
                        && gone.contains(&values)
                    {
                        violating.push((key.clone(), value.clone()));
                    }
                }
            }
        }
        if violating.is_empty() {
            continue;
        }
        found = true;
        let info = foreign_key_info(&fk);
        let mut recorded = artifacts::read(ctx.db, &child)?;
        for (key, value) in violating {
            let meta = artifacts::violation_meta(&info, &value);
            recorded.retain(|a| !(a.key == key && a.kind == artifacts::FOREIGN_KEY && a.meta == meta));
            recorded.push(Artifact {
                key,
                rootish: theirs,
                kind: artifacts::FOREIGN_KEY,
                info_hash: artifacts::info_hash(&info),
                meta,
            });
        }
        let mut stored = child.table.clone();
        stored.artifacts = artifacts::write(ctx.db, &child, recorded)?;
        let address = stored.write(ctx.db)?;
        merged.put_table(ctx.db, &child_name.0, &child_name.1, Some(address))?;
    }
    Ok(found)
}

/// apply_to_working makes row edits to a table of the working root, keeping its secondary indexes in step.
pub fn apply_to_working(ctx: &mut Ctx<'_>, table: &TableDef, edits: Vec<Edit>) -> Result<()> {
    let ours = entries(ctx.db, table)?;
    let stored = apply(ctx, table, &ours, edits)?;
    let address = stored.write(ctx.db)?;
    ctx.txn.root.put_table(ctx.db, &table.schema, &table.name, Some(address))?;
    Ok(())
}

impl Ctx<'_> {
    /// merge_concurrent merges the changes that other transactions committed to the branch's working set since this
    /// transaction began into its working and staged roots, as Dolt's transaction commit does, failing with Dolt's
    /// retry error when the merge leaves conflicts.
    pub fn merge_concurrent(&mut self) -> Result<()> {
        let current = self.db.head(&doltdb::create::working_set_ref(&self.txn.branch))?.unwrap_or_default();
        if current == self.txn.working_set || current.is_empty() || self.txn.working_set.is_empty() {
            return Ok(());
        }
        let (start_working, start_staged) = crate::txn::working_roots(self.db, self.txn.working_set)?;
        let (working, staged) = crate::txn::working_roots(self.db, current)?;
        let commits = Commits { ours: self.txn.head, theirs: self.txn.head, base: self.txn.head };
        if working.encode() != self.txn.root.encode() && working.encode() != start_working.encode() {
            let ours = self.txn.root.clone();
            let outcome = merge_roots(self, &working, &ours, &start_working, commits)?;
            if outcome.artifacts {
                return Err(crate::txn::retry_transaction_error(""));
            }
            self.txn.root = outcome.root;
        }
        if staged.encode() != self.txn.staged.encode() && staged.encode() != start_staged.encode() {
            let ours = self.txn.staged.clone();
            let outcome = merge_roots(self, &staged, &ours, &start_staged, commits)?;
            if outcome.artifacts {
                return Err(crate::txn::retry_transaction_error(""));
            }
            self.txn.staged = outcome.root;
        }
        self.txn.working_set = current;
        Ok(())
    }
}
