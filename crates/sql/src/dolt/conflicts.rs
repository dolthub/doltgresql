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

//! Dolt's tables of a merge's conflicts and constraint violations: which tables have them, each table's conflicting
//! and violating rows, deleting those rows to resolve them, and dolt_conflicts_resolve.

use std::collections::HashMap;

use base64::Engine;
use doltdb::root::Root;
use pg_query::NodeEnum;
use pg_query::protobuf::{DeleteStmt, UpdateStmt};
use store::Hash;

use crate::Outcome;
use crate::catalog::ColumnType;
use crate::catalog::table::{ColumnDef, TableDef};
use crate::dolt::args::error;
use crate::dolt::artifacts::{self, Artifact};
use crate::dolt::history;
use crate::error::Result;
use crate::expr::typ;
use crate::oid::{INT8, JSON, NUMERIC, TEXT, VARCHAR};
use crate::query::Ctx;
use crate::txn::read;
use crate::types::Value;

/// Kind is which table of a merge's leftovers an `ArtifactTable` is.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    Conflicts,
    Violations,
}

/// ArtifactTable is the table of one user table's conflicts or constraint violations: its kind, the user table's
/// schema and name, and the columns of the base, our, and their versions of its rows.
#[derive(Clone, Debug, PartialEq)]
pub struct ArtifactTable {
    pub kind: Kind,
    pub schema: String,
    pub name: String,
    pub keyless: bool,
    pub base: Vec<ColumnDef>,
    pub ours: Vec<ColumnDef>,
    pub theirs: Vec<ColumnDef>,
    /// The branches whose merge the DOLT_PREVIEW_MERGE_CONFLICTS function previews, instead of the working root.
    pub preview: Option<(String, String)>,
}

/// ordered returns a table's columns with its key columns first, as Dolt's conflict tables order them.
fn ordered(table: &TableDef) -> Vec<ColumnDef> {
    table.key_columns.iter().chain(&table.value_columns).map(|&i| table.columns[i].clone()).collect()
}

/// commit_table returns a table as a commit has it.
fn commit_table(ctx: &mut Ctx<'_>, commit: Hash, schema: &str, name: &str) -> Result<Option<TableDef>> {
    let root = Root::decode(&read(ctx.db, &history::load(ctx.db, commit)?.root)?)?;
    match root.table(ctx.db, schema, name)? {
        Some(address) => Ok(Some(TableDef::load(ctx.db, schema, name, address)?)),
        None => Ok(None),
    }
}

/// lookup returns the table of conflicts or constraint violations that a schema and name refer to when the name
/// starts with `dolt_conflicts_` or `dolt_constraint_violations_` and names a table that exists, searching the
/// session's schemas for an unqualified name.
pub fn lookup(ctx: &mut Ctx<'_>, schema: &str, name: &str) -> Result<Option<ArtifactTable>> {
    let (kind, table) = match (name.strip_prefix("dolt_conflicts_"), name.strip_prefix("dolt_constraint_violations_")) {
        (Some(table), _) => (Kind::Conflicts, table),
        (_, Some(table)) => (Kind::Violations, table),
        _ => return Ok(None),
    };
    let schemas = if schema.is_empty() { ctx.session.search_path() } else { vec![schema.to_string()] };
    for schema in schemas {
        let Some(ours) = ctx.txn.table(ctx.db, &schema, table)? else { continue };
        let mut base = ordered(&ours);
        let mut theirs = base.clone();
        if let Some(merge) = ctx.txn.merge.clone() {
            if let Some(found) = commit_table(ctx, merge.from_commit, &schema, table)? {
                theirs = ordered(&found);
            }
            let head = merge.pre_merge_head_commit.unwrap_or(ctx.txn.head);
            if let Some(commit) = history::merge_base(ctx.db, head, merge.from_commit)?
                && let Some(found) = commit_table(ctx, commit, &schema, table)?
            {
                base = ordered(&found);
            }
        }
        return Ok(Some(ArtifactTable {
            kind,
            schema,
            name: table.to_string(),
            keyless: ours.keyless(),
            ours: ordered(&ours),
            base,
            theirs,
            preview: None,
        }));
    }
    Ok(None)
}

/// violation_type returns the name of a kind of constraint violation.
fn violation_type(kind: u8) -> &'static str {
    match kind {
        artifacts::FOREIGN_KEY => "foreign key",
        artifacts::UNIQUE => "unique index",
        artifacts::CHECK => "check constraint",
        _ => "not null",
    }
}

/// conflict_id returns the identifier of a conflict, as Dolt's GetConflictId computes it.
fn conflict_id(key: &[u8], rootish: Hash) -> String {
    let mut bytes = key.to_vec();
    bytes.extend_from_slice(&rootish.0);
    let hash = xxhash_rust::xxh3::xxh3_128(&bytes).to_be_bytes();
    base64::engine::general_purpose::STANDARD_NO_PAD.encode(hash)
}

/// Version is a version of a table that conflict rows read: the table and its rows by key.
struct Version {
    table: Option<TableDef>,
}

impl Version {
    /// row returns a row's values in a column order and its cardinality, or None when the version lacks the row.
    fn row(&self, ctx: &mut Ctx<'_>, key: &[u8], columns: &[ColumnDef]) -> Result<Option<(Vec<Value>, u64)>> {
        let Some(table) = &self.table else { return Ok(None) };
        let root = std::sync::Arc::new(prolly::Node::decode(table.table.primary_index.clone())?);
        let Some(value) = prolly::get(ctx.db, root, key, &|a, b| table.compare_keys(a, b))? else { return Ok(None) };
        let (row, count) = table.decode_row(ctx.db, key, &value)?;
        let values = columns
            .iter()
            .map(|c| table.columns.iter().position(|t| t.tag == c.tag).map_or(Value::Null, |i| row[i].clone()))
            .collect();
        Ok(Some((values, count)))
    }
}

/// diff_type returns how a version of a conflicting row differs from the base.
fn diff_type(base: bool, other: bool) -> &'static str {
    match (base, other) {
        (false, _) => "added",
        (_, false) => "removed",
        _ => "modified",
    }
}

impl ArtifactTable {
    /// columns returns the table's column names and types.
    pub fn columns(&self) -> Vec<(String, ColumnType)> {
        let prefixed = |prefix: &str, columns: &[ColumnDef]| {
            columns.iter().map(|c| (format!("{prefix}{}", c.name), c.ty)).collect::<Vec<_>>()
        };
        let mut out = vec![("from_root_ish".to_string(), typ(TEXT))];
        match self.kind {
            Kind::Conflicts => {
                out.extend(prefixed("base_", &self.base));
                out.extend(prefixed("our_", &self.ours));
                out.push(("our_diff_type".into(), typ(TEXT)));
                out.extend(prefixed("their_", &self.theirs));
                out.push(("their_diff_type".into(), typ(TEXT)));
                out.push(("dolt_conflict_id".into(), typ(TEXT)));
                if self.keyless {
                    for side in ["base", "our", "their"] {
                        out.push((format!("{side}_cardinality"), typ(INT8)));
                    }
                }
            }
            Kind::Violations => {
                out.push(("violation_type".into(), typ(VARCHAR)));
                out.extend(prefixed("", &self.ours));
                out.push(("violation_info".into(), typ(JSON)));
            }
        }
        out
    }

    /// artifacts returns the artifacts of the kind that the table shows, with the user table.
    fn artifacts(&self, ctx: &mut Ctx<'_>) -> Result<(TableDef, Vec<Artifact>)> {
        let root = match &self.preview {
            Some((left, right)) => preview(ctx, left, right)?.0.root,
            None => ctx.txn.root.clone(),
        };
        let address = root.table(ctx.db, &self.schema, &self.name)?.ok_or_else(|| error("table not found"))?;
        let table = TableDef::load(ctx.db, &self.schema, &self.name, address)?;
        let all = artifacts::read(ctx.db, &table)?;
        let wanted = |a: &Artifact| (a.kind == artifacts::CONFLICT) == (self.kind == Kind::Conflicts);
        Ok((table, all.into_iter().filter(wanted).collect()))
    }

    /// rows returns the table's rows, each with the artifact it shows.
    fn rows_with_artifacts(&self, ctx: &mut Ctx<'_>) -> Result<Vec<(Vec<Value>, Artifact)>> {
        let (table, found) = self.artifacts(ctx)?;
        let mut out = Vec::with_capacity(found.len());
        let mut versions: HashMap<Hash, Version> = HashMap::new();
        for artifact in found {
            let mut row = vec![Value::Text(artifact.rootish.to_string())];
            if self.kind == Kind::Violations {
                let (info, value) = artifacts::violation_parts(&artifact.meta).unwrap_or_default();
                row.push(Value::Text(violation_type(artifact.kind).into()));
                let (values, _) = table.decode_row(ctx.db, &artifact.key, &value)?;
                row.extend(self.ours.iter().map(|c| {
                    table.columns.iter().position(|t| t.tag == c.tag).map_or(Value::Null, |i| values[i].clone())
                }));
                row.push(Value::Json(String::from_utf8_lossy(&info).into_owned()));
                out.push((row, artifact));
                continue;
            }
            let base_commit = artifacts::conflict_base(&artifact.meta).unwrap_or(Hash([0; Hash::LEN]));
            for commit in [base_commit, artifact.rootish] {
                if let std::collections::hash_map::Entry::Vacant(entry) = versions.entry(commit) {
                    let found = commit_table(ctx, commit, &self.schema, &self.name).unwrap_or(None);
                    entry.insert(Version { table: found });
                }
            }
            let ours = Version { table: Some(table.clone()) };
            let base = versions[&base_commit].row(ctx, &artifact.key, &self.base)?;
            let mine = ours.row(ctx, &artifact.key, &self.ours)?;
            let theirs = versions[&artifact.rootish].row(ctx, &artifact.key, &self.theirs)?;
            let values = |side: &Option<(Vec<Value>, u64)>, n: usize| {
                side.as_ref().map_or_else(|| vec![Value::Null; n], |(values, _)| values.clone())
            };
            row.extend(values(&base, self.base.len()));
            row.extend(values(&mine, self.ours.len()));
            row.push(Value::Text(diff_type(base.is_some(), mine.is_some()).into()));
            row.extend(values(&theirs, self.theirs.len()));
            row.push(Value::Text(diff_type(base.is_some(), theirs.is_some()).into()));
            row.push(Value::Text(conflict_id(&artifact.key, artifact.rootish)));
            if self.keyless {
                for side in [&base, &mine, &theirs] {
                    row.push(side.as_ref().map_or(Value::Null, |(_, count)| Value::Int8(*count as i64)));
                }
            }
            out.push((row, artifact));
        }
        Ok(out)
    }

    /// rows returns the table's rows.
    pub fn rows(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        Ok(self.rows_with_artifacts(ctx)?.into_iter().map(|(row, _)| row).collect())
    }
}

/// summary_rows returns the tables of the working root that have conflicts, or constraint violations, with how many,
/// and for conflicts the root objects with conflicting fields.
pub fn summary_rows(ctx: &mut Ctx<'_>, conflicts: bool) -> Result<Vec<Vec<Value>>> {
    let mut out = Vec::new();
    for ((schema, name), _) in crate::dolt::procedures::table_map(ctx.db, &ctx.txn.root.clone())? {
        let Some(table) = ctx.txn.table(ctx.db, &schema, &name)? else { continue };
        if table.table.artifacts.iter().all(|&b| b == 0) {
            continue;
        }
        let count =
            artifacts::read(ctx.db, &table)?.iter().filter(|a| (a.kind == artifacts::CONFLICT) == conflicts).count();
        if count > 0 {
            out.push(vec![Value::Text(name), Value::Numeric(crate::numeric::Numeric::from_i64(count as i64))]);
        }
    }
    if conflicts {
        out.extend(crate::dolt::objmerge::summary_rows(ctx)?);
    }
    Ok(out)
}

/// summary_columns returns the columns of the tables of tables with conflicts or constraint violations.
pub fn summary_columns(conflicts: bool) -> Vec<(&'static str, u32)> {
    let count = if conflicts { "num_conflicts" } else { "num_violations" };
    vec![("table", TEXT), (count, NUMERIC)]
}

/// remove removes artifacts from a table of the working root.
fn remove(ctx: &mut Ctx<'_>, table: &TableDef, doomed: &[Artifact]) -> Result<()> {
    let kept: Vec<Artifact> = artifacts::read(ctx.db, table)?.into_iter().filter(|a| !doomed.contains(a)).collect();
    let mut stored = table.table.clone();
    stored.artifacts = artifacts::write(ctx.db, table, kept)?;
    let address = stored.write(ctx.db)?;
    ctx.txn.root.put_table(ctx.db, &table.schema, &table.name, Some(address))?;
    Ok(())
}

impl Ctx<'_> {
    /// delete_artifacts runs a DELETE from a table of conflicts or constraint violations, which resolves the rows it
    /// deletes, or returns None when the DELETE is of another table.
    pub fn delete_artifacts(&mut self, delete: &DeleteStmt) -> Result<Option<Outcome>> {
        let Some(relation) = delete.relation.as_ref() else { return Ok(None) };
        if self.resolve_table(relation).is_ok() {
            return Ok(None);
        }
        let Some(table) = lookup(self, &relation.schemaname, &relation.relname)? else {
            let Some(table) =
                crate::dolt::objmerge::ObjectConflictTable::lookup(self, &relation.schemaname, &relation.relname)?
            else {
                return Ok(None);
            };
            let selected = self.select_rows(relation, vec![star()], delete.where_clause.clone())?;
            let count = table.delete(self, &selected)?;
            return Ok(Some(Outcome::command(format!("DELETE {count}"))));
        };
        let select = pg_query::protobuf::SelectStmt {
            target_list: vec![star()],
            from_clause: vec![pg_query::Node { node: Some(NodeEnum::RangeVar(relation.clone())) }],
            where_clause: delete.where_clause.clone(),
            ..Default::default()
        };
        let query = crate::plan::Planner { ctx: self, outer: Vec::new() }.plan_query(&select)?;
        let selected = query.plan.run(self)?;
        let all = table.rows_with_artifacts(self)?;
        let doomed: Vec<Artifact> =
            all.into_iter().filter(|(row, _)| selected.contains(row)).map(|(_, artifact)| artifact).collect();
        let user = self.txn.table(self.db, &table.schema, &table.name)?.ok_or_else(|| error("table not found"))?;
        remove(self, &user, &doomed)?;
        Ok(Some(Outcome::command(format!("DELETE {}", doomed.len()))))
    }
}

impl Ctx<'_> {
    /// select_rows runs a SELECT of targets from a relation with a WHERE clause.
    fn select_rows(
        &mut self,
        relation: &pg_query::protobuf::RangeVar,
        target_list: Vec<pg_query::Node>,
        where_clause: Option<Box<pg_query::Node>>,
    ) -> Result<Vec<Vec<Value>>> {
        let select = pg_query::protobuf::SelectStmt {
            target_list,
            from_clause: vec![pg_query::Node { node: Some(NodeEnum::RangeVar(relation.clone())) }],
            where_clause,
            ..Default::default()
        };
        let query = crate::plan::Planner { ctx: self, outer: Vec::new() }.plan_query(&select)?;
        query.plan.run(self)
    }

    /// is_conflicts_table reports whether a DML statement's target is a table of conflicts or constraint violations
    /// rather than a user table.
    pub fn is_conflicts_table(&mut self, relation: Option<&pg_query::protobuf::RangeVar>) -> Result<bool> {
        let Some(relation) = relation else { return Ok(false) };
        if self.resolve_table(relation).is_ok() {
            return Ok(false);
        }
        Ok(lookup(self, &relation.schemaname, &relation.relname)?.is_some()
            || crate::dolt::objmerge::ObjectConflictTable::lookup(self, &relation.schemaname, &relation.relname)?
                .is_some())
    }

    /// update_object_conflicts runs an UPDATE of a root object's conflicts table, which sets our values of the
    /// conflicting fields its rows select, leaving alone the rows it would not change, as Go's engine does.
    pub fn update_object_conflicts(&mut self, update: &UpdateStmt) -> Result<Option<Outcome>> {
        let Some(relation) = update.relation.as_ref() else { return Ok(None) };
        if self.resolve_table(relation).is_ok() {
            return Ok(None);
        }
        let Some(table) =
            crate::dolt::objmerge::ObjectConflictTable::lookup(self, &relation.schemaname, &relation.relname)?
        else {
            return Ok(None);
        };
        let our_value = update.target_list.iter().find_map(|target| match target.node.as_ref() {
            Some(NodeEnum::ResTarget(t)) if t.name == "our_value" => t.val.as_deref().cloned(),
            _ => None,
        });
        let target = |val: pg_query::Node| {
            let target = pg_query::protobuf::ResTarget { val: Some(Box::new(val)), location: -1, ..Default::default() };
            pg_query::Node { node: Some(NodeEnum::ResTarget(Box::new(target))) }
        };
        let rows = self.select_rows(
            relation,
            vec![
                target(column("dolt_conflict_id")),
                target(column("our_value")),
                target(our_value.unwrap_or_else(|| column("our_value"))),
            ],
            update.where_clause.clone(),
        )?;
        let changes: Vec<(String, Option<String>)> = rows
            .iter()
            .filter(|row| row[1] != row[2])
            .map(|row| (row[0].output().unwrap_or_default(), row[2].output()))
            .collect();
        table.update(self, &changes)?;
        Ok(Some(Outcome::command(format!("UPDATE {}", rows.len()))))
    }
}

/// column returns a reference to a column by name.
fn column(name: &str) -> pg_query::Node {
    let field = pg_query::Node { node: Some(NodeEnum::String(pg_query::protobuf::String { sval: name.to_string() })) };
    let column = pg_query::protobuf::ColumnRef { fields: vec![field], location: -1 };
    pg_query::Node { node: Some(NodeEnum::ColumnRef(column)) }
}

/// star returns the `*` target of a select list.
fn star() -> pg_query::Node {
    let star = pg_query::Node { node: Some(NodeEnum::AStar(pg_query::protobuf::AStar {})) };
    let column = pg_query::protobuf::ColumnRef { fields: vec![star], location: -1 };
    let target = pg_query::protobuf::ResTarget {
        val: Some(Box::new(pg_query::Node { node: Some(NodeEnum::ColumnRef(column)) })),
        location: -1,
        ..Default::default()
    };
    pg_query::Node { node: Some(NodeEnum::ResTarget(Box::new(target))) }
}

/// RESOLVE parses dolt_conflicts_resolve's arguments.
const RESOLVE: crate::dolt::args::Parser = crate::dolt::args::Parser {
    command: "conflicts resolve",
    options: &[("ours", "", crate::dolt::args::Kind::Flag), ("theirs", "", crate::dolt::args::Kind::Flag)],
    max_args: None,
};

/// dolt_conflicts_resolve resolves the conflicts of the named tables, or of every table for `.`, by keeping our rows
/// or taking theirs, as Dolt's procedure does.
pub fn dolt_conflicts_resolve(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let args: Vec<String> = args.iter().map(|a| a.output().unwrap_or_default()).collect();
    let parsed = RESOLVE.parse(&args)?;
    let (ours, theirs) = (parsed.has("ours"), parsed.has("theirs"));
    if ours == theirs {
        return Err(error("--ours or --theirs must be supplied"));
    }
    if parsed.args.is_empty() {
        return Err(error("specify at least one table to resolve conflicts for, or '.' for all tables"));
    }
    let mut names = Vec::new();
    let tables = crate::dolt::procedures::table_map(ctx.db, &ctx.txn.root.clone())?;
    for arg in &parsed.args {
        if arg == "." {
            names.extend(tables.keys().cloned());
            continue;
        }
        let found = ctx.session.search_path().into_iter().find_map(|s| {
            let key = (s, arg.clone());
            tables.contains_key(&key).then_some(key)
        });
        names.push(found.ok_or_else(|| error(format!("table not found: {arg}")))?);
    }
    for (schema, name) in names {
        let Some(table) = lookup(ctx, &schema, &format!("dolt_conflicts_{name}"))? else { continue };
        let rows = table.rows_with_artifacts(ctx)?;
        if rows.is_empty() {
            continue;
        }
        let user = ctx.txn.table(ctx.db, &schema, &name)?.ok_or_else(|| error("table not found"))?;
        let doomed: Vec<Artifact> = rows.iter().map(|(_, a)| a.clone()).collect();
        if theirs {
            let mut edits = Vec::new();
            for artifact in &doomed {
                let their_table = commit_table(ctx, artifact.rootish, &schema, &name)?;
                let value = match &their_table {
                    Some(t) => {
                        let root = std::sync::Arc::new(prolly::Node::decode(t.table.primary_index.clone())?);
                        prolly::get(ctx.db, root, &artifact.key, &|a, b| t.compare_keys(a, b))?
                    }
                    None => None,
                };
                edits.push((artifact.key.clone(), value));
            }
            crate::dolt::merge::apply_to_working(ctx, &user, edits)?;
        }
        let user = ctx.txn.table(ctx.db, &schema, &name)?.ok_or_else(|| error("table not found"))?;
        remove(ctx, &user, &doomed)?;
    }
    Ok(Value::Int8(0))
}

/// commit_check refuses to commit a working root with conflicts or constraint violations, as Dolt's transactions do
/// unless the session allows them, listing each table's violations.
pub fn commit_check(
    db: &mut doltdb::database::Database,
    root: &Root,
    schema_conflicts: bool,
    allow_conflicts: bool,
    force: bool,
    autocommit: bool,
) -> Result<()> {
    let mut conflicts = schema_conflicts;
    let mut violations: Vec<(String, Vec<Artifact>)> = Vec::new();
    for ((schema, name), address) in crate::dolt::procedures::table_map(db, root)? {
        let stored = doltdb::table::Table::decode(&read(db, &address)?)?;
        if stored.artifacts.iter().all(|&b| b == 0) {
            continue;
        }
        let table = TableDef::load(db, &schema, &name, address)?;
        let found = artifacts::read(db, &table)?;
        conflicts |= found.iter().any(|a| a.kind == artifacts::CONFLICT);
        let cvs: Vec<Artifact> = found.into_iter().filter(|a| a.kind != artifacts::CONFLICT).collect();
        if !cvs.is_empty() {
            violations.push((name, cvs));
        }
    }
    if conflicts && !(allow_conflicts || force) {
        return Err(error(if autocommit {
            "Merge conflict detected, @autocommit transaction rolled back. @autocommit must be disabled so that merge \
             conflicts can be resolved using the dolt_conflicts and dolt_schema_conflicts tables before manually \
             committing the transaction. Alternatively, to commit transactions with merge conflicts, set \
             @@dolt_allow_commit_conflicts = 1"
        } else {
            "Merge conflict detected, transaction rolled back. Merge conflicts must be resolved using the \
             dolt_conflicts and dolt_schema_conflicts tables before committing a transaction. To commit transactions \
             with merge conflicts, set @@dolt_allow_commit_conflicts = 1"
        }));
    }
    if violations.is_empty() || force {
        return Ok(());
    }
    let mut message = String::from(
        "Committing this transaction resulted in a working set with constraint violations, transaction rolled back. \
         This constraint violation may be the result of a previous merge or the result of transaction sequencing. \
         Constraint violations from a merge can be resolved using the dolt_constraint_violations table before \
         committing the transaction. To allow transactions to be committed with constraint violations from a merge \
         or transaction sequencing set @@dolt_force_transaction_commit=1.\nConstraint violations: ",
    );
    for (i, (_, cvs)) in violations.iter().enumerate() {
        let mut counts: std::collections::BTreeMap<String, usize> = Default::default();
        for cv in cvs {
            *counts.entry(describe(cv)).or_default() += 1;
        }
        if i > 0 {
            message.push_str(", ");
        }
        for (description, count) in counts {
            message.push_str(&description);
            if count > 1 {
                message.push_str(&format!(" ({count} row(s))"));
            }
        }
    }
    Err(error(message))
}

/// describe writes a constraint violation as Dolt's transaction error lists it.
fn describe(cv: &Artifact) -> String {
    let (info, _) = artifacts::violation_parts(&cv.meta).unwrap_or_default();
    let info: serde_json::Value = serde_json::from_slice(&info).unwrap_or_default();
    let text = |name: &str| info.get(name).and_then(|v| v.as_str()).unwrap_or_default().to_string();
    let list = |name: &str| {
        let items: Vec<String> = info
            .get(name)
            .and_then(|v| v.as_array())
            .map(|a| a.iter().filter_map(|i| i.as_str().map(str::to_string)).collect())
            .unwrap_or_default();
        format!("[{}]", items.join(" "))
    };
    match cv.kind {
        artifacts::FOREIGN_KEY => format!(
            "\nType: Foreign Key Constraint Violation\n\tForeignKey: {},\n\tTable: {},\n\tReferencedTable: {},\n\tIndex: \
             {},\n\tReferencedIndex: {}",
            text("ForeignKey"),
            text("Table"),
            text("ReferencedTable"),
            text("Index"),
            text("ReferencedIndex")
        ),
        artifacts::UNIQUE => format!(
            "\nType: Unique Key Constraint Violation,\n\tName: {},\n\tColumns: {}",
            text("Name"),
            list("Columns")
        ),
        artifacts::NOT_NULL => format!("\nType: Null Constraint Violation,\n\tColumns: {}", list("Columns")),
        _ => format!(
            "\nType: Check Constraint Violation,\n\tName: {},\n\tExpression: {}",
            text("Name"),
            text("Expression")
        ),
    }
}

/// unmerged_tables returns the names of the working root's tables with conflicts or constraint violations, and of
/// those whose schemas conflict in the merge, in name order.
pub fn unmerged_tables(ctx: &mut Ctx<'_>) -> Result<Vec<String>> {
    let mut names: Vec<String> = ctx
        .txn
        .merge
        .as_ref()
        .map(|m| m.unmergable_tables.iter().map(|t| String::from_utf8_lossy(t).into_owned()).collect())
        .unwrap_or_default();
    for ((_, name), address) in crate::dolt::procedures::table_map(ctx.db, &ctx.txn.root.clone())? {
        let stored = doltdb::table::Table::decode(&read(ctx.db, &address)?)?;
        if stored.artifacts.iter().any(|&b| b != 0) {
            names.push(name);
        }
    }
    names.sort();
    names.dedup();
    Ok(names)
}

/// preview merges one branch into another without changing either, returning the outcome and the commits it
/// merged.
fn preview(
    ctx: &mut Ctx<'_>,
    left: &str,
    right: &str,
) -> Result<(crate::dolt::merge::Outcome, crate::dolt::merge::Commits)> {
    let ours = history::resolve(ctx.db, ctx.txn.head, left)?;
    let theirs = history::resolve(ctx.db, ctx.txn.head, right)?;
    let base = history::merge_base(ctx.db, ours, theirs)?.ok_or_else(|| error("no common ancestor"))?;
    let root = |ctx: &mut Ctx<'_>, commit: Hash| -> Result<Root> {
        Ok(Root::decode(&read(ctx.db, &history::load(ctx.db, commit)?.root)?)?)
    };
    let (our_root, their_root, base_root) = (root(ctx, ours)?, root(ctx, theirs)?, root(ctx, base)?);
    let commits = crate::dolt::merge::Commits { ours, theirs, base };
    if base == theirs || base == ours {
        let outcome = crate::dolt::merge::Outcome { root: our_root, artifacts: false, schema_conflicts: Vec::new() };
        return Ok((outcome, commits));
    }
    let outcome = crate::dolt::merge::merge_roots(ctx, &our_root, &their_root, &base_root, commits)?;
    Ok((outcome, commits))
}

/// preview_function returns the table that a DOLT_PREVIEW_MERGE_CONFLICTS call's arguments describe: the conflicts
/// that merging the second branch into the first would leave in a table.
pub fn preview_function(ctx: &mut Ctx<'_>, args: &[String]) -> Result<ArtifactTable> {
    let [left, right, table] = args else {
        return Err(error(format!(
            "function 'dolt_preview_merge_conflicts' expected 3 arguments, {} received",
            args.len()
        )));
    };
    if table.is_empty() {
        return Err(error("table name cannot be empty"));
    }
    let (_, commits) = preview(ctx, left, right)?;
    let (schema, name) = match table.split_once('.') {
        Some((schema, name)) => (schema.to_string(), name.to_string()),
        None => (ctx.creation_schema()?, table.clone()),
    };
    let columns = |ctx: &mut Ctx<'_>, commit: Hash| -> Result<Option<Vec<ColumnDef>>> {
        Ok(commit_table(ctx, commit, &schema, &name)?.map(|t| ordered(&t)))
    };
    let ours = columns(ctx, commits.ours)?.ok_or_else(|| {
        crate::error::PgError::new(crate::error::code::UNDEFINED_TABLE, format!("relation \"{table}\" does not exist"))
    })?;
    let keyless = commit_table(ctx, commits.ours, &schema, &name)?.is_some_and(|t| t.keyless());
    Ok(ArtifactTable {
        kind: Kind::Conflicts,
        base: columns(ctx, commits.base)?.unwrap_or_else(|| ours.clone()),
        theirs: columns(ctx, commits.theirs)?.unwrap_or_else(|| ours.clone()),
        ours,
        keyless,
        schema,
        name,
        preview: Some((left.clone(), right.clone())),
    })
}

/// dolt_preview_merge_conflicts_summary returns each table that merging the second branch into the first would leave
/// with conflicts, with how many rows conflict, or how many schemas conflict, as rows of records.
pub fn dolt_preview_merge_conflicts_summary(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let args: Vec<String> = args.iter().map(|a| a.output().unwrap_or_default()).collect();
    let [left, right] = args.as_slice() else {
        return Err(error(format!(
            "function 'dolt_preview_merge_conflicts_summary' expected 2 arguments, {} received",
            args.len()
        )));
    };
    if left.is_empty() {
        return Err(error("left branch name cannot be empty"));
    }
    if right.is_empty() {
        return Err(error("right branch name cannot be empty"));
    }
    let (outcome, _) = preview(ctx, left, right)?;
    let mut rows = Vec::new();
    for (schema, name) in &outcome.schema_conflicts {
        let full = crate::dolt::diff::full_name(&(schema.clone(), name.clone()));
        rows.push(Value::Record(vec![Value::Text(full), Value::Null, Value::Int8(1)]));
    }
    for ((schema, name), address) in crate::dolt::procedures::table_map(ctx.db, &outcome.root)? {
        let table = TableDef::load(ctx.db, &schema, &name, address)?;
        let count = artifacts::read(ctx.db, &table)?.iter().filter(|a| a.kind == artifacts::CONFLICT).count();
        if count > 0 {
            let full = crate::dolt::diff::full_name(&(schema, name));
            rows.push(Value::Record(vec![Value::Text(full), Value::Int8(count as i64), Value::Int8(0)]));
        }
    }
    Ok(Value::Set(rows))
}
