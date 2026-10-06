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

//! Dolt's diffs: the system tables and table functions that show a table's rows at each commit, its row changes
//! between revisions, and its pending changes, and those that show which tables and root objects changed.

use std::cmp::Ordering;
use std::collections::{BTreeMap, HashMap};

use doltdb::database::Database;
use doltdb::root::Root;
use store::Hash;

use crate::catalog::ColumnType;
use crate::catalog::table::{ColumnDef, TableDef};
use crate::dolt::args::error;
use crate::dolt::history;
use crate::engine::quote_identifier;
use crate::error::{PgError, Result, code};
use crate::expr::{CmpOp, Expr, typ};
use crate::oid::{BOOL, INT8, TEXT, TIMESTAMP};
use crate::query::Ctx;
use crate::txn::read;
use crate::types::Value;

/// Kind is which of the system tables over a user table a `UserTable` is.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    History,
    Diff,
    CommitDiff,
    Workspace,
    /// The DOLT_DIFF table function, which compares the table between two revisions.
    DiffFunction,
}

/// PREFIXES are the name prefixes of the system tables over a user table.
const PREFIXES: &[(&str, Kind)] = &[
    ("dolt_history_", Kind::History),
    ("dolt_diff_", Kind::Diff),
    ("dolt_commit_diff_", Kind::CommitDiff),
    ("dolt_workspace_", Kind::Workspace),
];

/// UserTable is a system table over a user table: its kind, the user table's schema and name, its schema and name on
/// the older side of a change, its columns on the newer and older sides, and the commits a commit diff table compares.
#[derive(Clone, Debug, PartialEq)]
pub struct UserTable {
    pub kind: Kind,
    pub schema: String,
    pub name: String,
    pub from_table: Name,
    pub to: Vec<ColumnDef>,
    pub from: Vec<ColumnDef>,
    /// The `to_commit` and `from_commit` that a query's conditions give a commit diff table.
    pub commits: Option<(Expr, Expr)>,
    /// Whether rows end with their row number, as WITH ORDINALITY asks of the DOLT_DIFF function.
    pub ordinality: bool,
}

/// Change is a row's change between two versions of a table: the row before and the row after, one of which is
/// missing when the row was added or removed.
type Change = (Option<Vec<Value>>, Option<Vec<Value>>);

/// Side is one version of a table that a diff compares: its name, its commit time, and the table with its address.
#[derive(Clone)]
struct Side {
    name: String,
    date: Value,
    table: Option<(Hash, TableDef)>,
}

/// load returns a table of a root, if the root has it, matching its name without regard to case when no table has
/// the exact name, as Dolt's GetTableInsensitive does.
fn load(db: &mut Database, root: &Root, schema: &str, name: &str) -> Result<Option<(Hash, TableDef)>> {
    if let Some(address) = root.table(db, schema, name)? {
        return Ok(Some((address, TableDef::load(db, schema, name, address)?)));
    }
    let tables = crate::dolt::procedures::table_map(db, root)?;
    match tables.iter().find(|((s, n), _)| s == schema && n.eq_ignore_ascii_case(name)) {
        Some(((_, name), &address)) => Ok(Some((address, TableDef::load(db, schema, name, address)?))),
        None => Ok(None),
    }
}

/// commit_root returns the root value of a commit.
fn commit_root(db: &mut Database, commit: &history::CommitInfo) -> Result<Root> {
    Ok(Root::decode(&read(db, &commit.root)?)?)
}

/// lookup returns the system table that a schema and name refer to when the name starts with the prefix of a system
/// table over a user table that exists, searching the session's schemas for an unqualified name, or an empty
/// workspace table when only the schema exists, as Dolt's is.
pub fn lookup(ctx: &mut Ctx<'_>, schema: &str, name: &str) -> Result<Option<UserTable>> {
    let Some((kind, table)) = PREFIXES.iter().find_map(|(prefix, kind)| name.strip_prefix(prefix).map(|t| (*kind, t)))
    else {
        return Ok(None);
    };
    let schemas = if schema.is_empty() { ctx.session.search_path() } else { vec![schema.to_string()] };
    let head = Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
    let (working, staged) = (ctx.txn.root.clone(), ctx.txn.staged.clone());
    for schema in &schemas {
        let newest = match load(ctx.db, &working, schema, table)? {
            Some(found) => Some(found),
            None if kind == Kind::Workspace => match load(ctx.db, &staged, schema, table)? {
                Some(found) => Some(found),
                None => load(ctx.db, &head, schema, table)?,
            },
            None => None,
        };
        let Some((_, to)) = newest else { continue };
        let from = match kind {
            Kind::Workspace => {
                load(ctx.db, &head, schema, table)?.map_or_else(|| to.columns.clone(), |(_, t)| t.columns)
            }
            _ => to.columns.clone(),
        };
        return Ok(Some(UserTable {
            kind,
            schema: schema.clone(),
            name: to.name.clone(),
            from_table: (schema.clone(), to.name.clone()),
            to: to.columns,
            from,
            commits: None,
            ordinality: false,
        }));
    }
    let exists = |s: &str| ctx.txn.root.schemas.iter().any(|existing| existing == s.as_bytes());
    let empty_schema = match schema {
        "" => ctx.creation_schema().ok(),
        _ if exists(schema) => Some(schema.to_string()),
        _ => None,
    };
    Ok(match empty_schema {
        Some(schema) if kind == Kind::Workspace => Some(UserTable {
            kind,
            from_table: (schema.clone(), table.to_string()),
            schema,
            name: table.to_string(),
            to: Vec::new(),
            from: Vec::new(),
            commits: None,
            ordinality: false,
        }),
        _ => None,
    })
}

/// blame_view returns the definition of the blame view that a schema and name refer to when the name starts with
/// `dolt_blame_` and names a table that exists, which shows the commit that last changed each row, as Dolt's does.
pub fn blame_view(ctx: &mut Ctx<'_>, schema: &str, name: &str) -> Result<Option<String>> {
    let Some(table) = name.strip_prefix("dolt_blame_") else { return Ok(None) };
    let schemas = if schema.is_empty() { ctx.session.search_path() } else { vec![schema.to_string()] };
    for schema in schemas {
        let Some(def) = ctx.txn.table(ctx.db, &schema, table)? else { continue };
        if def.keyless() {
            return Err(error("unable to generate blame view for table without primary key"));
        }
        let keys: Vec<&str> = def.key_columns.iter().map(|&i| def.columns[i].name.as_str()).collect();
        let quoted = |prefix: &str, key: &str| quote_identifier(&format!("{prefix}{key}"));
        let list = |f: &dyn Fn(&str) -> String| keys.iter().map(|k| f(k)).collect::<Vec<_>>().join(", ");
        return Ok(Some(format!(
            "CREATE VIEW blame AS WITH sorted_diffs_by_pk AS (SELECT {}, to_commit, to_commit_date, diff_type, \
             ROW_NUMBER() OVER (PARTITION BY {} ORDER BY coalesce(to_commit_date, from_commit_date) DESC) row_num \
             FROM {}.{}) SELECT {}, sd.to_commit AS commit, sd.to_commit_date AS commit_date, dl.committer, dl.email, \
             dl.message FROM sorted_diffs_by_pk AS sd LEFT JOIN dolt.log AS dl ON dl.commit_hash = sd.to_commit \
             WHERE sd.row_num = 1 AND sd.diff_type <> 'removed' AND sd.to_commit <> 'WORKING' ORDER BY {}",
            list(&|k| quoted("to_", k)),
            list(&|k| format!("coalesce({}, {})", quoted("to_", k), quoted("from_", k))),
            quote_identifier(&schema),
            quote_identifier(&format!("dolt_diff_{table}")),
            list(&|k| format!("sd.{} AS {}", quoted("to_", k), quote_identifier(k))),
            list(&|k| format!("sd.{}", quoted("to_", k))),
        )));
    }
    Ok(None)
}

/// prefixed returns the names and types of columns with a prefix added to each name.
fn prefixed(columns: &[ColumnDef], prefix: &str) -> Vec<(String, ColumnType)> {
    columns.iter().map(|c| (format!("{prefix}{}", c.name), c.ty)).collect()
}

/// named returns a column's name and the type of an OID.
fn named(name: &str, oid: u32) -> (String, ColumnType) {
    (name.to_string(), typ(oid))
}

/// timestamp converts Unix milliseconds to a timestamp, or NULL without a time.
fn timestamp(millis: Option<i64>) -> Value {
    millis.map_or(Value::Null, |m| {
        Value::Timestamp(m * 1000 + crate::datetime::UNIX_EPOCH_DAYS * crate::datetime::USECS_PER_DAY)
    })
}

/// project arranges a row of a table as the target columns, matching columns by tag, with NULL for a column the
/// table lacks and for a value that cannot take the target column's type.
fn project(table: &TableDef, row: &[Value], target: &[ColumnDef]) -> Vec<Value> {
    target
        .iter()
        .map(|column| match table.columns.iter().position(|c| c.tag == column.tag) {
            Some(i) if table.columns[i].ty == column.ty => row[i].clone(),
            Some(i) => crate::cast::cast_value(row[i].clone(), column.ty, false).unwrap_or(Value::Null),
            None => Value::Null,
        })
        .collect()
}

/// entries returns the keys and values of a table's primary index in key order.
fn entries(db: &mut Database, table: &TableDef) -> Result<Vec<(Vec<u8>, Vec<u8>)>> {
    let node = prolly::Node::decode(table.table.primary_index.clone())?;
    let mut items = Vec::new();
    prolly::walk_leaves(db, &node, &mut |key, value| {
        items.push((key.to_vec(), value.to_vec()));
        Ok(())
    })?;
    Ok(items)
}

/// diffable reports whether rows of two versions of a table can be matched by their primary keys, which needs the
/// same key columns.
fn diffable(from: &TableDef, to: &TableDef) -> bool {
    let key = |t: &TableDef| t.key_columns.iter().map(|&i| (t.columns[i].tag, t.columns[i].ty)).collect::<Vec<_>>();
    key(from) == key(to)
}

/// changes returns the row changes between two versions of a table in key order, with a removed or added row for
/// each copy of a keyless row whose count changed, or None when the versions have different primary keys.
fn changes(db: &mut Database, from: Option<&TableDef>, to: Option<&TableDef>) -> Result<Option<Vec<Change>>> {
    if let (Some(f), Some(t)) = (from, to)
        && !diffable(f, t)
    {
        return Ok(None);
    }
    let from_entries = match from {
        Some(table) => entries(db, table)?,
        None => Vec::new(),
    };
    let to_entries = match to {
        Some(table) => entries(db, table)?,
        None => Vec::new(),
    };
    let Some(order) = to.or(from) else { return Ok(Some(Vec::new())) };
    let mut out = Vec::new();
    let (mut i, mut j) = (0, 0);
    while i < from_entries.len() || j < to_entries.len() {
        let ordering = match (from_entries.get(i), to_entries.get(j)) {
            (Some(f), Some(t)) => order.compare_keys(&f.0, &t.0),
            (Some(_), None) => Ordering::Less,
            _ => Ordering::Greater,
        };
        let old = (ordering != Ordering::Greater).then(|| &from_entries[i]);
        let new = (ordering != Ordering::Less).then(|| &to_entries[j]);
        i += old.is_some() as usize;
        j += new.is_some() as usize;
        if let (Some(old), Some(new)) = (old, new)
            && old.1 == new.1
        {
            continue;
        }
        let old = match (old, from) {
            (Some((key, value)), Some(table)) => Some(table.decode_row(db, key, value)?),
            _ => None,
        };
        let new = match (new, to) {
            (Some((key, value)), Some(table)) => Some(table.decode_row(db, key, value)?),
            _ => None,
        };
        if order.keyless() {
            let count = |side: &Option<(Vec<Value>, u64)>| side.as_ref().map_or(0, |(_, n)| *n);
            let (before, after) = (count(&old), count(&new));
            for _ in after..before {
                out.push((old.as_ref().map(|(row, _)| row.clone()), None));
            }
            for _ in before..after {
                out.push((None, new.as_ref().map(|(row, _)| row.clone())));
            }
            continue;
        }
        out.push((old.map(|(row, _)| row), new.map(|(row, _)| row)));
    }
    Ok(Some(out))
}

/// diff_type returns the name of a change's kind.
fn diff_type(change: &Change) -> &'static str {
    match change {
        (None, _) => "added",
        (_, None) => "removed",
        _ => "modified",
    }
}

impl UserTable {
    /// columns returns the table's column names and types.
    pub fn columns(&self) -> Vec<(String, ColumnType)> {
        let mut columns = Vec::new();
        match self.kind {
            Kind::History => {
                columns.extend(self.to.iter().map(|c| (c.name.clone(), c.ty)));
                columns.extend([named("commit_hash", TEXT), named("committer", TEXT), named("commit_date", TIMESTAMP)]);
            }
            Kind::Diff | Kind::CommitDiff | Kind::DiffFunction => {
                columns.extend(prefixed(&self.to, "to_"));
                columns.extend([named("to_commit", TEXT), named("to_commit_date", TIMESTAMP)]);
                columns.extend(prefixed(&self.from, "from_"));
                columns.extend([named("from_commit", TEXT), named("from_commit_date", TIMESTAMP)]);
                columns.push(named("diff_type", TEXT));
            }
            Kind::Workspace => {
                columns.extend([named("id", INT8), named("staged", BOOL)]);
                if !self.to.is_empty() {
                    columns.push(named("diff_type", TEXT));
                    columns.extend(prefixed(&self.to, "to_"));
                    columns.extend(prefixed(&self.from, "from_"));
                }
            }
        }
        if self.ordinality {
            columns.push(named("ordinality", INT8));
        }
        columns
    }

    /// take_commits takes the `to_commit` and `from_commit` that a commit diff table compares from the conditions
    /// of a query over it, which must compare each column to a value computed once.
    pub fn take_commits(&mut self, conditions: &[Expr]) {
        if self.kind != Kind::CommitDiff {
            return;
        }
        let to_commit = self.to.len();
        let from_commit = to_commit + 2 + self.from.len();
        let mut found = (None, None);
        for condition in conditions {
            let Expr::Compare(CmpOp::Eq, left, right) = condition else { continue };
            let (column, value) = match (left.as_ref(), right.as_ref()) {
                (Expr::Column(i), value) | (value, Expr::Column(i)) => (*i, value),
                _ => continue,
            };
            let mut constant = true;
            value.visit(&mut |e| {
                if matches!(
                    e,
                    Expr::Column(_) | Expr::Outer(..) | Expr::Exists(_) | Expr::Scalar(_) | Expr::AnySubquery(..)
                ) {
                    constant = false;
                }
            });
            if !constant {
                continue;
            }
            if column == to_commit {
                found.0 = Some(value.clone());
            } else if column == from_commit {
                found.1 = Some(value.clone());
            }
        }
        self.commits = match found {
            (Some(to), Some(from)) => Some((to, from)),
            _ => None,
        };
    }

    /// rows returns the table's rows.
    pub fn rows(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        let mut rows = match self.kind {
            Kind::History => self.history_rows(ctx)?,
            Kind::Diff => self.diff_rows(ctx)?,
            Kind::CommitDiff | Kind::DiffFunction => self.commit_diff_rows(ctx)?,
            Kind::Workspace => self.workspace_rows(ctx)?,
        };
        if self.ordinality {
            for (i, row) in rows.iter_mut().enumerate() {
                row.push(Value::Int8(i as i64 + 1));
            }
        }
        Ok(rows)
    }

    /// history_rows returns the table's rows at each commit reachable from the head, with the commit.
    fn history_rows(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        let mut scanned: HashMap<Hash, Vec<Vec<Value>>> = HashMap::new();
        let mut out = Vec::new();
        for commit in history::log(ctx.db, &[ctx.txn.head])? {
            let root = commit_root(ctx.db, &commit)?;
            let Some((address, table)) = load(ctx.db, &root, &self.schema, &self.name)? else { continue };
            let rows = match scanned.entry(address) {
                std::collections::hash_map::Entry::Occupied(rows) => rows.into_mut(),
                std::collections::hash_map::Entry::Vacant(entry) => {
                    let rows = crate::query::scan(ctx.db, &table)?;
                    entry.insert(rows.iter().map(|row| project(&table, row, &self.to)).collect())
                }
            };
            for row in rows.iter() {
                let mut row = row.clone();
                row.extend([
                    Value::Text(commit.hash.to_string()),
                    Value::Text(commit.committer_name.clone()),
                    timestamp(Some(commit.committer_millis as i64)),
                ]);
                out.push(row);
            }
        }
        Ok(out)
    }

    /// diff_row returns a diff table row of a change between two versions of the table.
    fn diff_row(&self, change: &Change, from: &Side, to: &Side) -> Vec<Value> {
        let side = |row: &Option<Vec<Value>>, side: &Side, columns: &[ColumnDef]| match (row, &side.table) {
            (Some(row), Some((_, table))) => project(table, row, columns),
            _ => vec![Value::Null; columns.len()],
        };
        let mut row = side(&change.1, to, &self.to);
        row.extend([Value::Text(to.name.clone()), to.date.clone()]);
        row.extend(side(&change.0, from, &self.from));
        row.extend([Value::Text(from.name.clone()), from.date.clone()]);
        row.push(Value::Text(diff_type(change).into()));
        row
    }

    /// diff_rows returns the changes to the table in each commit reachable from the head, and in the working root,
    /// newest first, until a commit changes the table's primary key, as Dolt's diff tables return them.
    fn diff_rows(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        let working = Side {
            name: "WORKING".into(),
            date: Value::Null,
            table: load(ctx.db, &ctx.txn.root.clone(), &self.schema, &self.name)?,
        };
        let mut newer: HashMap<Hash, Side> = HashMap::from([(ctx.txn.head, working)]);
        let mut out = Vec::new();
        for commit in history::log(ctx.db, &[ctx.txn.head])? {
            let root = commit_root(ctx.db, &commit)?;
            let side = Side {
                name: commit.hash.to_string(),
                date: timestamp(Some(commit.committer_millis as i64)),
                table: load(ctx.db, &root, &self.schema, &self.name)?,
            };
            let Some(to) = newer.get(&commit.hash).cloned() else { continue };
            for parent in &commit.parents {
                newer.insert(*parent, side.clone());
            }
            let address = |s: &Side| s.table.as_ref().map(|(a, _)| *a);
            if address(&side) == address(&to) {
                continue;
            }
            let table = |s: &Side| s.table.as_ref().map(|(_, t)| t.clone());
            let Some(changes) = changes(ctx.db, table(&side).as_ref(), table(&to).as_ref())? else { break };
            out.extend(changes.iter().map(|change| self.diff_row(change, &side, &to)));
        }
        Ok(out)
    }

    /// commit_side returns the version of the table that a commit diff table's commit names: the working or staged
    /// root, which the DOLT_DIFF function dates to the transaction's start, or a commit that a spec names.
    fn commit_side(&self, ctx: &mut Ctx<'_>, spec: &Expr, table: &Name) -> Result<Side> {
        let name = spec.eval(ctx, &[])?.output().unwrap_or_default();
        let now = || match self.kind {
            Kind::DiffFunction => Value::Timestamp(ctx.txn.started),
            _ => Value::Null,
        };
        let (root, date) = if name.eq_ignore_ascii_case("WORKING") {
            (ctx.txn.root.clone(), now())
        } else if name.eq_ignore_ascii_case("STAGED") {
            (ctx.txn.staged.clone(), now())
        } else {
            let hash = history::resolve(ctx.db, ctx.txn.head, &name)?;
            let commit = history::load(ctx.db, hash)?;
            (commit_root(ctx.db, &commit)?, timestamp(Some(commit.committer_millis as i64)))
        };
        let table = load(ctx.db, &root, &table.0, &table.1)?;
        Ok(Side { name, date, table })
    }

    /// commit_diff_rows returns the changes to the table between the two commits that the query's conditions name.
    fn commit_diff_rows(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        let Some((to, from)) = &self.commits else {
            return Err(error(format!(
                "error querying table dolt_commit_diff_{}: dolt_commit_diff_* tables must be filtered to a single \
                 'to_commit'",
                self.name
            )));
        };
        let to = self.commit_side(ctx, to, &(self.schema.clone(), self.name.clone()))?;
        let from = self.commit_side(ctx, from, &self.from_table)?;
        let table = |s: &Side| s.table.as_ref().map(|(_, t)| t.clone());
        let changes = changes(ctx.db, table(&from).as_ref(), table(&to).as_ref())?.unwrap_or_default();
        Ok(changes.iter().map(|change| self.diff_row(change, &from, &to)).collect())
    }

    /// workspace_rows returns the table's staged changes and then its changes that are not staged, numbered in order.
    fn workspace_rows(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        if self.to.is_empty() {
            return Ok(Vec::new());
        }
        let head = Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
        let roots = [head, ctx.txn.staged.clone(), ctx.txn.root.clone()];
        let mut tables = Vec::new();
        for root in &roots {
            tables.push(load(ctx.db, root, &self.schema, &self.name)?.map(|(_, t)| t));
        }
        let mut out = Vec::new();
        for (staged, from, to) in [(true, &tables[0], &tables[1]), (false, &tables[1], &tables[2])] {
            for change in changes(ctx.db, from.as_ref(), to.as_ref())?.unwrap_or_default() {
                let side =
                    |row: &Option<Vec<Value>>, table: &Option<TableDef>, columns: &[ColumnDef]| match (row, table) {
                        (Some(row), Some(table)) => project(table, row, columns),
                        _ => vec![Value::Null; columns.len()],
                    };
                let mut row =
                    vec![Value::Int8(out.len() as i64), Value::Bool(staged), Value::Text(diff_type(&change).into())];
                row.extend(side(&change.1, to, &self.to));
                row.extend(side(&change.0, from, &self.from));
                out.push(row);
            }
        }
        Ok(out)
    }
}

/// Name is a table or root object by schema and name.
type Name = (String, String);

/// Delta is a table or root object that differs between two roots: its name and address in each root that has it.
#[derive(Clone, Debug)]
pub struct Delta {
    pub from: Option<(Name, Hash)>,
    pub to: Option<(Name, Hash)>,
    /// Whether it is a root object, such as a sequence or a function, rather than a table.
    pub object: bool,
}

/// Summary is how a table or root object changed: its name, its names before and after, the kind of change, and
/// whether its rows and its schema changed.
#[derive(Clone, Debug)]
pub struct Summary {
    pub table_name: String,
    pub from_name: String,
    pub to_name: String,
    pub diff_type: &'static str,
    pub data_change: bool,
    pub schema_change: bool,
}

/// full_name returns a table's name qualified by its schema when it has one.
pub fn full_name(name: &Name) -> String {
    match name.0.as_str() {
        "" => name.1.clone(),
        schema => format!("{schema}.{}", name.1),
    }
}

/// COLLECTIONS are the positions of the root object collections that diffs show, which leave out conflicts.
const COLLECTIONS: &[usize] = &[0, 1, 2, 3, 4, 6, 7, 8, 9];

/// type_name returns a type's name as a root object's name shows it, qualified by its schema unless that is
/// `pg_catalog` or the object's own schema.
fn type_name(type_id: &str, schema: &str) -> String {
    let segments = crate::catalog::id::segments(type_id.as_bytes());
    match segments.as_slice() {
        [type_schema, name, ..] if type_schema == "pg_catalog" || type_schema == schema => name.clone(),
        [type_schema, name, ..] => format!("{type_schema}.{name}"),
        _ => String::new(),
    }
}

/// object_name returns the schema and name that a root object's ID shows in diffs, as Go's root objects name
/// themselves: routines with their parameter types, triggers after their tables, and casts and operators by their
/// types.
fn object_name(collection: usize, id: &[u8]) -> Name {
    let segments = crate::catalog::id::segments(id);
    let segment = |i: usize| segments.get(i).cloned().unwrap_or_default();
    match collection {
        2 | 6 | 9 => {
            let types: Vec<String> = segments.iter().skip(2).map(|t| type_name(t, &segment(0))).collect();
            (segment(0), format!("{}({})", segment(1), types.join(",")))
        }
        3 => (segment(0), format!("{}.{}", segment(1), segment(2))),
        4 => (String::new(), segment(0)),
        7 => {
            let parts = |t: &str| {
                let s = crate::catalog::id::segments(t.as_bytes());
                (s.first().cloned().unwrap_or_default(), s.get(1).cloned().unwrap_or_default())
            };
            let ((source_schema, source), (target_schema, target)) = (parts(&segment(0)), parts(&segment(1)));
            (String::new(), format!("({source_schema})|({source})|({target_schema})|({target})"))
        }
        8 => {
            let parts = |t: &str| {
                let s = crate::catalog::id::segments(t.as_bytes());
                format!("({})|({})", s.first().cloned().unwrap_or_default(), s.get(1).cloned().unwrap_or_default())
            };
            (segment(0), format!("({})|{}|{}", segment(1), parts(&segment(2)), parts(&segment(3))))
        }
        _ => (segment(0), segment(1)),
    }
}

/// object_map returns a root's root objects by the schema and name that diffs show, which the status and diffs show
/// alongside tables.
pub fn object_map(db: &mut Database, root: &Root) -> Result<BTreeMap<Name, Hash>> {
    let mut objects = BTreeMap::new();
    for &collection in COLLECTIONS {
        for (key, address) in root.objects(db, collection)? {
            objects.insert(object_name(collection, &key), address);
        }
    }
    Ok(objects)
}

/// table_columns returns the tags and names of a table's columns.
fn table_columns(db: &mut Database, name: &Name, address: Hash) -> Result<Vec<(u64, String)>> {
    let table = TableDef::load(db, &name.0, &name.1, address)?;
    Ok(table.columns.into_iter().map(|c| (c.tag, c.name)).collect())
}

/// deltas returns the tables and root objects that differ between two roots, matching a table that was renamed by a
/// column that kept its tag and name, ordered by their older names and then their newer names, as Dolt's
/// GetTableDeltas returns them.
pub fn deltas(db: &mut Database, from: &Root, to: &Root) -> Result<Vec<Delta>> {
    let from_tables = crate::dolt::procedures::table_map(db, from)?;
    let to_tables = crate::dolt::procedures::table_map(db, to)?;
    let (from_keys, to_keys) = (crate::foreign::load(db, from)?, crate::foreign::load(db, to)?);
    let mut out = Vec::new();
    for (name, &address) in &from_tables {
        if let Some(&new) = to_tables.get(name)
            && (address != new || foreign_keys_of(&from_keys, name) != foreign_keys_of(&to_keys, name))
        {
            out.push(Delta { from: Some((name.clone(), address)), to: Some((name.clone(), new)), object: false });
        }
    }
    let mut dropped: Vec<(Name, Hash)> =
        from_tables.iter().filter(|(n, _)| !to_tables.contains_key(*n)).map(|(n, a)| (n.clone(), *a)).collect();
    let mut added: Vec<(Name, Hash)> =
        to_tables.iter().filter(|(n, _)| !from_tables.contains_key(*n)).map(|(n, a)| (n.clone(), *a)).collect();
    let mut i = 0;
    while i < dropped.len() {
        let old = table_columns(db, &dropped[i].0, dropped[i].1)?;
        let mut renamed = None;
        for (j, (name, address)) in added.iter().enumerate() {
            if table_columns(db, name, *address)?.iter().any(|column| old.contains(column)) {
                renamed = Some(j);
                break;
            }
        }
        match renamed {
            Some(j) => {
                let new = added.remove(j);
                out.push(Delta { from: Some(dropped.remove(i)), to: Some(new), object: false });
            }
            None => i += 1,
        }
    }
    out.extend(dropped.into_iter().map(|d| Delta { from: Some(d), to: None, object: false }));
    out.extend(added.into_iter().map(|a| Delta { from: None, to: Some(a), object: false }));
    let (from_objects, to_objects) = (object_map(db, from)?, object_map(db, to)?);
    for (name, &address) in &from_objects {
        match to_objects.get(name) {
            Some(&new) if new == address => {}
            new => out.push(Delta {
                from: Some((name.clone(), address)),
                to: new.map(|&n| (name.clone(), n)),
                object: true,
            }),
        }
    }
    for (name, &address) in &to_objects {
        if !from_objects.contains_key(name) {
            out.push(Delta { from: None, to: Some((name.clone(), address)), object: true });
        }
    }
    let key = |side: &Option<(Name, Hash)>| side.as_ref().map(|(n, _)| n.clone()).unwrap_or_default();
    out.sort_by(|a, b| key(&a.from).cmp(&key(&b.from)).then_with(|| key(&a.to).cmp(&key(&b.to))));
    Ok(out)
}

/// foreign_keys_of returns the foreign keys that a table declares or that refer to it.
fn foreign_keys_of<'a>(
    keys: &'a [crate::foreign::ForeignKeyDef],
    name: &Name,
) -> Vec<&'a crate::foreign::ForeignKeyDef> {
    keys.iter()
        .filter(|k| {
            (k.child_schema == name.0 && k.child_table == name.1)
                || (k.parent_schema == name.0 && k.parent_table == name.1)
        })
        .collect()
}

/// stored_table returns a table as stored at an address.
fn stored_table(db: &mut Database, address: Hash) -> Result<doltdb::table::Table> {
    Ok(doltdb::table::Table::decode(&read(db, &address)?)?)
}

/// has_rows reports whether a stored table has any rows.
fn has_rows(table: &doltdb::table::Table) -> Result<bool> {
    Ok(prolly::Node::decode(table.primary_index.clone())?.count() > 0)
}

impl Delta {
    /// summary returns how the table or root object changed, as Dolt's GetSummary describes it.
    pub fn summary(&self, db: &mut Database, from_root: &Root, to_root: &Root) -> Result<Summary> {
        let name = |side: &Option<(Name, Hash)>| side.as_ref().map(|(n, _)| full_name(n)).unwrap_or_default();
        let (from_name, to_name) = (name(&self.from), name(&self.to));
        let summary = |table_name: &str, diff_type, data_change, schema_change| Summary {
            table_name: table_name.to_string(),
            from_name: from_name.clone(),
            to_name: to_name.clone(),
            diff_type,
            data_change,
            schema_change,
        };
        let (from, to) = match (&self.from, &self.to) {
            (None, Some((_, address))) => {
                let data = self.object || has_rows(&stored_table(db, *address)?)?;
                return Ok(summary(&to_name, "added", data, true));
            }
            (Some((_, address)), None) => {
                let data = self.object || has_rows(&stored_table(db, *address)?)?;
                return Ok(summary(&from_name, "dropped", data, true));
            }
            (Some(from), Some(to)) => (from, to),
            (None, None) => return Ok(summary("", "modified", false, false)),
        };
        if self.object {
            return Ok(summary(&from_name, "modified", from.1 != to.1, false));
        }
        let (old, new) = (stored_table(db, from.1)?, stored_table(db, to.1)?);
        let data_change = old.primary_index != new.primary_index;
        if from.0 != to.0 {
            return Ok(summary(&to_name, "renamed", data_change, true));
        }
        let (from_keys, to_keys) = (crate::foreign::load(db, from_root)?, crate::foreign::load(db, to_root)?);
        let schema_change = foreign_keys_of(&from_keys, &from.0) != foreign_keys_of(&to_keys, &to.0)
            || old.auto_increment != new.auto_increment
            || old.schema != new.schema;
        Ok(summary(&from_name, "modified", data_change, schema_change))
    }
}

/// ChangeSet is a set of changes that the unscoped diff tables show: the staged changes, the changes that are not
/// staged, or a commit's changes from its first parent, with the roots before and after.
struct ChangeSet {
    commit: Option<history::CommitInfo>,
    name: &'static str,
    from: Root,
    to: Root,
}

impl ChangeSet {
    /// all returns the staged changes, the changes that are not staged, and the changes of each commit reachable from
    /// the head that has a parent, newest first.
    fn all(ctx: &mut Ctx<'_>) -> Result<Vec<ChangeSet>> {
        let head = Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
        let (staged, working) = (ctx.txn.staged.clone(), ctx.txn.root.clone());
        let mut sets = vec![
            ChangeSet { commit: None, name: "STAGED", from: head, to: staged.clone() },
            ChangeSet { commit: None, name: "WORKING", from: staged, to: working },
        ];
        for commit in history::log(ctx.db, &[ctx.txn.head])? {
            let Some(parent) = commit.parents.first() else { continue };
            let to = commit_root(ctx.db, &commit)?;
            let from = commit_root(ctx.db, &history::load(ctx.db, *parent)?)?;
            sets.push(ChangeSet { commit: Some(commit), name: "", from, to });
        }
        Ok(sets)
    }

    /// row returns an unscoped diff table row of the change set: its commit and the table's name, the values that
    /// come before the commit's committer details, those details, the values that come after them, and the commit's
    /// author details, with NULL details for the changes of the working set.
    fn row(&self, table_name: String, before: Vec<Value>, after: Vec<Value>) -> Vec<Value> {
        let name = self.commit.as_ref().map_or_else(|| self.name.to_string(), |c| c.hash.to_string());
        let mut row = vec![Value::Text(name), Value::Text(table_name)];
        row.extend(before);
        row.extend(match &self.commit {
            Some(c) => vec![
                Value::Text(c.committer_name.clone()),
                Value::Text(c.committer_email.clone()),
                timestamp(Some(c.committer_millis as i64)),
                Value::Text(c.description.clone()),
            ],
            None => vec![Value::Null; 4],
        });
        row.extend(after);
        row.extend(match &self.commit {
            Some(c) => {
                vec![Value::Text(c.name.clone()), Value::Text(c.email.clone()), timestamp(Some(c.author_millis))]
            }
            None => vec![Value::Null; 3],
        });
        row
    }
}

/// unscoped_rows returns the tables that each commit reachable from the head changed from its first parent, after
/// the staged changes and then the changes that are not staged, as Dolt's unscoped diff table returns them.
pub fn unscoped_rows(ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
    let mut out = Vec::new();
    for set in ChangeSet::all(ctx)? {
        for delta in deltas(ctx.db, &set.from, &set.to)? {
            let summary = delta.summary(ctx.db, &set.from, &set.to)?;
            let changes = vec![Value::Bool(summary.data_change), Value::Bool(summary.schema_change)];
            out.push(set.row(summary.table_name, Vec::new(), changes));
        }
    }
    Ok(out)
}

/// column_changes returns the columns of a changed table with how each changed: added and dropped columns, then
/// the columns that both versions have whose values differ in any changed row, as Dolt's column diff table finds
/// them.
fn column_changes(db: &mut Database, delta: &Delta) -> Result<Vec<(String, &'static str)>> {
    let table = |db: &mut Database, side: &Option<(Name, Hash)>| -> Result<Option<TableDef>> {
        side.as_ref().map(|(name, address)| TableDef::load(db, &name.0, &name.1, *address)).transpose()
    };
    let (from, to) = (table(db, &delta.from)?, table(db, &delta.to)?);
    let (from, to) = match (from, to) {
        (None, Some(to)) => return Ok(to.columns.into_iter().map(|c| (c.name, "added")).collect()),
        (Some(from), None) => return Ok(from.columns.into_iter().map(|c| (c.name, "removed")).collect()),
        (Some(from), Some(to)) => (from, to),
        (None, None) => return Ok(Vec::new()),
    };
    let has = |t: &TableDef, tag: u64| t.columns.iter().any(|c| c.tag == tag);
    let mut out: Vec<(String, &'static str)> =
        to.columns.iter().filter(|c| !has(&from, c.tag)).map(|c| (c.name.clone(), "added")).collect();
    out.extend(from.columns.iter().filter(|c| !has(&to, c.tag)).map(|c| (c.name.clone(), "removed")));
    let shared: Vec<&ColumnDef> = to.columns.iter().filter(|c| has(&from, c.tag)).collect();
    let mut modified = vec![false; shared.len()];
    for (old, new) in changes(db, Some(&from), Some(&to))?.unwrap_or_default() {
        let old = old.map_or_else(|| vec![Value::Null; to.columns.len()], |row| project(&from, &row, &to.columns));
        let new = new.unwrap_or_else(|| vec![Value::Null; to.columns.len()]);
        for (i, column) in shared.iter().enumerate() {
            let position = to.columns.iter().position(|c| c.tag == column.tag).unwrap_or_default();
            modified[i] |= old[position] != new[position];
        }
    }
    out.extend(shared.iter().zip(modified).filter(|(_, m)| *m).map(|(c, _)| (c.name.clone(), "modified")));
    Ok(out)
}

/// column_rows returns the columns of the tables that each commit reachable from the head changed from its first
/// parent, after those of the staged changes and of the changes that are not staged, as Dolt's column diff table
/// returns them.
pub fn column_rows(ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
    let mut out = Vec::new();
    for set in ChangeSet::all(ctx)? {
        for delta in deltas(ctx.db, &set.from, &set.to)? {
            if delta.object {
                continue;
            }
            let name = delta.to.as_ref().or(delta.from.as_ref()).map(|(n, _)| full_name(n)).unwrap_or_default();
            for (column, diff_type) in column_changes(ctx.db, &delta)? {
                out.push(set.row(name.clone(), vec![Value::Text(column)], vec![Value::Text(diff_type.into())]));
            }
        }
    }
    Ok(out)
}

/// expected describes how many arguments a diff table function takes.
fn expected(counts: &std::ops::RangeInclusive<usize>) -> String {
    match counts.end() - counts.start() {
        0 => counts.start().to_string(),
        1 => format!("{} or {}", counts.start(), counts.end()),
        _ => format!("{} to {}", counts.start(), counts.end()),
    }
}

/// diff_refs returns the revisions that a diff table function's arguments compare, with the remaining arguments,
/// from two revisions or from one argument of the form `from..to`, or `from...to` to compare with their merge base,
/// each form taking the given numbers of arguments.
fn diff_refs(
    ctx: &mut Ctx<'_>,
    args: &[String],
    function: &str,
    plain: std::ops::RangeInclusive<usize>,
    dotted: std::ops::RangeInclusive<usize>,
) -> Result<(String, String, Vec<String>)> {
    let count_error = |counts: &std::ops::RangeInclusive<usize>| {
        error(format!("function '{function}' expected {} arguments, {} received", expected(counts), args.len()))
    };
    match args.first().map(|a| a.split_once("...").or_else(|| a.split_once(".."))) {
        Some(Some((from, to))) => {
            if !dotted.contains(&args.len()) {
                return Err(count_error(&dotted));
            }
            let mut from = from.to_string();
            if args[0].contains("...") {
                let left = history::resolve(ctx.db, ctx.txn.head, &from)?;
                let right = history::resolve(ctx.db, ctx.txn.head, to)?;
                let base = history::merge_base(ctx.db, left, right)?.ok_or_else(|| error("no common ancestor"))?;
                from = base.to_string();
            }
            Ok((from, to.to_string(), args[1..].to_vec()))
        }
        _ if !plain.contains(&args.len()) => Err(count_error(&plain)),
        _ => Ok((args[0].clone(), args[1].clone(), args[2..].to_vec())),
    }
}

/// ref_root returns the root that a revision names: the working or staged root, or a commit's.
fn ref_root(ctx: &mut Ctx<'_>, name: &str) -> Result<Root> {
    match name {
        "WORKING" => Ok(ctx.txn.root.clone()),
        "STAGED" => Ok(ctx.txn.staged.clone()),
        _ => {
            let hash = history::resolve(ctx.db, ctx.txn.head, name)?;
            commit_root(ctx.db, &history::load(ctx.db, hash)?)
        }
    }
}

/// matches reports whether a delta is of the table that a diff table function's argument names, comparing only the
/// table's name without its schema, as Dolt's findMatchingDelta does.
fn matches(side: &Option<(Name, Hash)>, table: &str) -> bool {
    side.as_ref().is_some_and(|(name, _)| name.1.eq_ignore_ascii_case(table))
}

/// dolt_diff_summary returns how each table and root object changed between two revisions, or only the table the
/// third argument names, as rows of records.
pub fn dolt_diff_summary(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let args: Vec<String> = args.iter().map(|a| a.output().unwrap_or_default()).collect();
    let (from, to, rest) = diff_refs(ctx, &args, "dolt_diff_summary", 2..=3, 1..=2)?;
    let (from, to) = (ref_root(ctx, &from)?, ref_root(ctx, &to)?);
    let mut all = deltas(ctx.db, &from, &to)?;
    let key = |side: &Option<(Name, Hash)>| side.as_ref().map(|(n, _)| n.clone()).unwrap_or_default();
    all.sort_by_key(|d| key(&d.to));
    if let Some(table) = rest.first() {
        let found = all.iter().find(|d| matches(&d.to, table)).or_else(|| all.iter().find(|d| matches(&d.from, table)));
        all = found.into_iter().cloned().collect();
    }
    let mut rows = Vec::new();
    for delta in all {
        let summary = delta.summary(ctx.db, &from, &to)?;
        rows.push(Value::Record(vec![
            Value::Text(summary.from_name),
            Value::Text(summary.to_name),
            Value::Text(summary.diff_type.into()),
            Value::Bool(summary.data_change),
            Value::Bool(summary.schema_change),
        ]));
    }
    Ok(Value::Set(rows))
}

/// find_table returns the name of a table that a root has, searching the session's schemas for an unqualified name.
fn find_table(ctx: &mut Ctx<'_>, root: &Root, table: &str) -> Result<Option<Name>> {
    let (schemas, name) = match table.split_once('.') {
        Some((schema, name)) => (vec![schema.to_string()], name),
        None => (ctx.session.search_path(), table),
    };
    for schema in schemas {
        if root.table(ctx.db, &schema, name)?.is_some() {
            return Ok(Some((schema, name.to_string())));
        }
    }
    Ok(None)
}

/// diff_function returns the table that a DOLT_DIFF call's arguments describe: the changes to a table between two
/// revisions, with the table's columns at each, following a table that was renamed between them.
pub fn diff_function(ctx: &mut Ctx<'_>, args: &[String]) -> Result<UserTable> {
    let (from_ref, to_ref, rest) = diff_refs(ctx, args, "dolt_diff", 3..=3, 2..=2)?;
    let table = &rest[0];
    let (from_root, to_root) = (ref_root(ctx, &from_ref)?, ref_root(ctx, &to_ref)?);
    let all = deltas(ctx.db, &from_root, &to_root)?;
    let found = all.iter().find(|d| matches(&d.to, table)).or_else(|| all.iter().find(|d| matches(&d.from, table)));
    let names = |side: &Option<(Name, Hash)>| side.as_ref().map(|(n, _)| n.clone());
    let (from_name, to_name) = match found.filter(|d| !d.object) {
        Some(delta) => (names(&delta.from), names(&delta.to)),
        None => (find_table(ctx, &from_root, table)?, find_table(ctx, &to_root, table)?),
    };
    let columns = |ctx: &mut Ctx<'_>, root: &Root, name: &Option<Name>| -> Result<Option<Vec<ColumnDef>>> {
        Ok(match name {
            Some(name) => load(ctx.db, root, &name.0, &name.1)?.map(|(_, t)| t.columns),
            None => None,
        })
    };
    let from_columns = columns(ctx, &from_root, &from_name)?;
    let to_columns = columns(ctx, &to_root, &to_name)?;
    let (Some(newest), Some(oldest)) = (to_name.clone().or(from_name.clone()), from_name.or(to_name)) else {
        return Err(PgError::new(code::UNDEFINED_TABLE, format!("relation \"{table}\" does not exist")));
    };
    let to = to_columns.clone().or(from_columns.clone()).unwrap_or_default();
    Ok(UserTable {
        kind: Kind::DiffFunction,
        schema: newest.0,
        name: newest.1,
        from_table: oldest,
        from: from_columns.unwrap_or_else(|| to.clone()),
        to,
        commits: Some((Expr::Const(Value::Text(to_ref)), Expr::Const(Value::Text(from_ref)))),
        ordinality: false,
    })
}

/// Stat counts how many rows and cells of a table changed, as Dolt's DiffStatProgress does.
#[derive(Clone, Copy, Debug, Default)]
struct Stat {
    adds: u64,
    removes: u64,
    changes: u64,
    cell_changes: u64,
    old_rows: u64,
    new_rows: u64,
    old_cells: u64,
    new_cells: u64,
}

/// changed_cells counts the cells of a modified row that changed: the older row's values that the newer row lacks,
/// stores differently, or holds a different value in, plus the newer row's extra values, as Dolt's
/// prollyCountCellDiff counts them.
fn changed_cells(from: &TableDef, to: &TableDef, old: &[Value], new: &[Value]) -> u64 {
    let mut changed = 0u64;
    for &i in &from.value_columns {
        let column = &from.columns[i];
        match to.value_columns.iter().find(|&&j| to.columns[j].name == column.name) {
            Some(&j) if to.columns[j].encoding == column.encoding && old[i] == new[j] => {}
            _ => changed += 1,
        }
    }
    changed.wrapping_add((to.value_columns.len() as u64).wrapping_sub(from.value_columns.len() as u64))
}

impl Delta {
    /// stat counts the rows and cells that changed, with whether the table has no primary key, or returns None when
    /// the table's primary key changed.
    fn stat(&self, db: &mut Database) -> Result<Option<(Stat, bool)>> {
        if self.object {
            let stat = match (&self.from, &self.to) {
                (Some(_), Some(_)) => Stat { changes: 1, ..Stat::default() },
                (None, _) => Stat { adds: 1, ..Stat::default() },
                (_, None) => Stat { removes: 1, ..Stat::default() },
            };
            return Ok(Some((stat, false)));
        }
        let table = |db: &mut Database, side: &Option<(Name, Hash)>| -> Result<Option<TableDef>> {
            side.as_ref().map(|(name, address)| TableDef::load(db, &name.0, &name.1, *address)).transpose()
        };
        let (from, to) = (table(db, &self.from)?, table(db, &self.to)?);
        let Some(found) = changes(db, from.as_ref(), to.as_ref())? else { return Ok(None) };
        let keyless = to.as_ref().or(from.as_ref()).is_some_and(TableDef::keyless);
        let mut stat = Stat::default();
        if !keyless {
            let count = |db: &mut Database, t: &Option<TableDef>| -> Result<(u64, u64)> {
                Ok(match t {
                    Some(t) => {
                        let rows = entries(db, t)?.len() as u64;
                        (rows, rows * t.columns.len() as u64)
                    }
                    None => (0, 0),
                })
            };
            (stat.old_rows, stat.old_cells) = count(db, &from)?;
            (stat.new_rows, stat.new_cells) = count(db, &to)?;
        }
        for change in &found {
            match change {
                (None, _) => stat.adds += 1,
                (_, None) => stat.removes += 1,
                (Some(old), Some(new)) => {
                    stat.changes += 1;
                    if let (Some(f), Some(t)) = (&from, &to) {
                        stat.cell_changes = stat.cell_changes.wrapping_add(changed_cells(f, t, old, new));
                    }
                }
            }
        }
        Ok(Some((stat, keyless)))
    }
}

/// stat_row returns a dolt_diff_stat row of a table's counts, which leaves out all but the row counts of a table
/// without a primary key, as Dolt's getRowFromDiffStat does.
fn stat_row(name: String, stat: Stat, new_columns: u64, keyless: bool) -> Value {
    let int = |n: u64| Value::Int8(n as i64);
    if keyless {
        let mut row = vec![Value::Text(name), Value::Null, int(stat.adds), int(stat.removes)];
        row.extend(std::iter::repeat_n(Value::Null, 8));
        return Value::Record(row);
    }
    let inserts = stat.adds as f64 * new_columns as f64;
    let deletes = stat.removes as f64 * new_columns as f64;
    let difference = stat.new_cells as f64 - stat.old_cells as f64;
    let (cells_added, cells_deleted) = if difference > 0.0 {
        (difference + deletes, deletes)
    } else if difference < 0.0 {
        (inserts, difference.abs() + inserts)
    } else if inserts != deletes {
        (inserts.max(deletes), inserts.max(deletes))
    } else {
        (inserts, deletes)
    };
    Value::Record(vec![
        Value::Text(name),
        int(stat.old_rows.wrapping_sub(stat.changes).wrapping_sub(stat.removes)),
        int(stat.adds),
        int(stat.removes),
        int(stat.changes),
        int(cells_added as u64),
        int(cells_deleted as u64),
        int(stat.cell_changes),
        int(stat.old_rows),
        int(stat.new_rows),
        int(stat.old_cells),
        int(stat.new_cells),
    ])
}

/// named_in reports whether a root has a table or root object that a diff table function's argument names.
fn named_in(db: &mut Database, root: &Root, table: &str) -> Result<bool> {
    let mut names = crate::dolt::procedures::table_map(db, root)?;
    names.extend(object_map(db, root)?);
    Ok(names.into_iter().any(|(name, address)| matches(&Some((name, address)), table)))
}

/// dolt_diff_stat returns how many rows and cells of each table changed between two revisions, or only of the
/// table the third argument names, as rows of records.
pub fn dolt_diff_stat(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let args: Vec<String> = args.iter().map(|a| a.output().unwrap_or_default()).collect();
    let (from, to, rest) = diff_refs(ctx, &args, "dolt_diff_stat", 2..=3, 1..=2)?;
    let (from, to) = (ref_root(ctx, &from)?, ref_root(ctx, &to)?);
    let all = deltas(ctx.db, &from, &to)?;
    let name = |side: &Option<(Name, Hash)>| side.as_ref().map(|(n, _)| n.clone());
    let mut rows = Vec::new();
    let selected: Vec<(Delta, String)> = match rest.first() {
        Some(table) => {
            if !named_in(ctx.db, &from, table)? && !named_in(ctx.db, &to, table)? {
                return Err(PgError::new(code::UNDEFINED_TABLE, format!("relation \"{table}\" does not exist")));
            }
            let found =
                all.iter().find(|d| matches(&d.to, table)).or_else(|| all.iter().find(|d| matches(&d.from, table)));
            found
                .map(|d| {
                    let schema = name(&d.from).or(name(&d.to)).map(|n| n.0).unwrap_or_default();
                    (d.clone(), full_name(&(schema, table.clone())))
                })
                .into_iter()
                .collect()
        }
        None => all
            .iter()
            .map(|d| (d.clone(), name(&d.to).or(name(&d.from)).map(|n| full_name(&n)).unwrap_or_default()))
            .collect(),
    };
    for (delta, table_name) in selected {
        let new_columns = match (&delta.to, delta.object) {
            (Some((name, address)), false) => TableDef::load(ctx.db, &name.0, &name.1, *address)?.columns.len() as u64,
            _ => 0,
        };
        match delta.stat(ctx.db)? {
            Some((stat, keyless)) => {
                if stat.adds + stat.removes + stat.changes == 0 && stat.old_cells == stat.new_cells {
                    continue;
                }
                rows.push(stat_row(table_name, stat, new_columns, keyless));
            }
            None if rest.is_empty() => rows.push(stat_row(table_name, Stat::default(), 0, false)),
            None => {
                return Err(error(format!(
                    "failed to compute diff stat for table {table_name}: primary key set changed"
                )));
            }
        }
    }
    Ok(Value::Set(rows))
}
