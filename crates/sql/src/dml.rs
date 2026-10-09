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

//! Statements that change rows: INSERT, UPDATE, and DELETE.

use std::borrow::Cow;
use std::cmp::Ordering;
use std::sync::Arc;

use doltdb::database::Database;
use pg_query::NodeEnum;
use pg_query::protobuf::{DeleteStmt, InsertStmt, UpdateStmt};
use prolly::{NodeStore, Tuple, get};

use crate::auth::Object;
use crate::cast::cast_value;
use crate::catalog::ColumnType;
use crate::catalog::table::{HIDDEN_BASE, IndexDef, TableDef};
use crate::error::{ErrorObjects, PgError, Result, code};
use crate::expr::{Binder, Expr, Scope, ScopeColumn, arg_location, assign, coerce, position, typ};
use crate::foreign::Change;
use crate::plan::{JoinKind, Plan, Planner, push_down};
use crate::query::{Ctx, scan};
use crate::triggers::{AFTER, BEFORE, Event, Triggers};
use crate::txn::Txn;
use crate::types::Value;
use crate::{Outcome, oid};

/// InsertSource is where an INSERT's rows come from.
#[derive(Clone, Debug)]
enum InsertSource {
    /// Rows of expressions, already converted to the target columns' types.
    Values(Vec<Vec<Expr>>),
    Select(Box<Plan>),
}

/// RowRules are what a written row must go through: its columns' defaults and its table's check constraints.
#[derive(Clone, Debug)]
pub struct RowRules {
    /// Each column's default, over no row.
    defaults: Vec<Option<Expr>>,
    /// Each check constraint's name and condition, over the row.
    checks: Vec<(String, Expr)>,
    /// Each generated column and its expression, over the row.
    generated: Vec<(usize, Expr)>,
}

/// IndexRules are the bound expressions of a table's indexes: its hidden expression columns, with the text that
/// error details name them by, and each index's predicate.
#[derive(Clone, Debug, Default)]
pub(crate) struct IndexRules {
    /// Each hidden expression column's expression, over the row.
    hidden: Vec<Expr>,
    /// The text that error details name each hidden expression column by.
    names: Vec<String>,
    /// Each index's predicate over the row, or None for an index of every row.
    predicates: Vec<Option<Expr>>,
}

impl IndexRules {
    /// column_name returns the name that error details give a column of an index.
    pub(crate) fn column_name<'t>(&'t self, table: &'t TableDef, column: usize) -> &'t str {
        match column.checked_sub(HIDDEN_BASE) {
            Some(k) => &self.names[k],
            None => &table.columns[column].name,
        }
    }

    /// hidden returns the expressions of the hidden expression columns.
    pub(crate) fn hidden(&self) -> &[Expr] {
        &self.hidden
    }

    /// predicates returns each index's predicate.
    pub(crate) fn predicates(&self) -> &[Option<Expr>] {
        &self.predicates
    }

    /// indexed returns a row extended with its hidden expression columns' values, and whether each index holds it.
    pub(crate) fn indexed<'r>(&self, ctx: &mut Ctx<'_>, row: &'r [Value]) -> Result<(Cow<'r, [Value]>, Vec<bool>)> {
        let mut held = Vec::with_capacity(self.predicates.len());
        for predicate in &self.predicates {
            held.push(match predicate {
                Some(predicate) => predicate.is_true(ctx, row)?,
                None => true,
            });
        }
        if self.hidden.is_empty() {
            return Ok((Cow::Borrowed(row), held));
        }
        let mut extended = row.to_vec();
        for expr in &self.hidden {
            extended.push(expr.eval(ctx, row)?);
        }
        Ok((Cow::Owned(extended), held))
    }
}

/// Returning is a RETURNING list: its expressions over the written rows, and its result columns.
#[derive(Clone, Debug)]
pub struct Returning {
    exprs: Vec<Expr>,
    pub columns: Vec<crate::Column>,
}

/// ConflictTarget is what an ON CONFLICT clause checks for conflicts: every unique constraint, the primary key, or
/// one unique index by its position.
#[derive(Clone, Debug, PartialEq)]
enum ConflictTarget {
    Any,
    Primary,
    Index(usize),
}

/// ConflictAction is what an ON CONFLICT clause does with a conflicting row.
#[derive(Clone, Debug)]
enum ConflictAction {
    Nothing,
    /// The new value of each assigned column, and the condition for updating, over the existing row followed by the
    /// proposed row.
    Update {
        assignments: Vec<(usize, Expr)>,
        filter: Option<Expr>,
    },
}

/// OnConflict is an INSERT's ON CONFLICT clause.
#[derive(Clone, Debug)]
struct OnConflict {
    target: ConflictTarget,
    action: ConflictAction,
}

/// InsertPlan is a planned INSERT.
#[derive(Clone, Debug)]
pub struct InsertPlan {
    table: TableDef,
    rules: RowRules,
    /// The table column that each source value goes to.
    targets: Vec<usize>,
    source: InsertSource,
    on_conflict: Option<OnConflict>,
    pub returning: Option<Returning>,
}

/// UpdatePlan is a planned UPDATE.
#[derive(Clone, Debug)]
pub struct UpdatePlan {
    table: TableDef,
    rules: RowRules,
    /// The filter, over the old row followed by a row of the FROM list.
    filter: Option<Expr>,
    /// The new value of each assigned column, over the old row followed by a row of the FROM list.
    assignments: Vec<(usize, Expr)>,
    from: Option<Box<Plan>>,
    /// The RETURNING list, over the new row followed by the row of the FROM list.
    pub returning: Option<Returning>,
}

/// DeletePlan is a planned DELETE.
#[derive(Clone, Debug)]
pub struct DeletePlan {
    table: TableDef,
    /// The filter, over the row followed by a row of the USING list.
    filter: Option<Expr>,
    using: Option<Box<Plan>>,
    pub returning: Option<Returning>,
}

/// table_scope returns the scope of a table's columns under its name or alias.
pub(crate) fn table_scope(table: &TableDef, alias: Option<&str>) -> Scope {
    let name = alias.unwrap_or(&table.name);
    let table_oid = crate::pgcatalog::snapshot::table_oid(&table.schema, &table.name);
    Scope {
        columns: table
            .columns
            .iter()
            .enumerate()
            .map(|(i, c)| ScopeColumn {
                table: name.to_string(),
                name: c.name.clone(),
                ty: c.ty,
                hidden: false,
                origin: (table_oid, i as u16 + 1),
            })
            .collect(),
    }
}

/// row_text renders a row as Postgres does in constraint error details.
fn row_text(values: &[Value]) -> String {
    values.iter().map(|v| v.output().unwrap_or_else(|| "null".into())).collect::<Vec<_>>().join(", ")
}

/// check_row fails as Postgres does when a row has NULL in a NOT NULL column or fails a check constraint.
fn check_row(ctx: &mut Ctx<'_>, table: &TableDef, rules: &RowRules, row: &mut [Value]) -> Result<()> {
    for (i, expr) in &rules.generated {
        row[*i] = expr.eval(ctx, row)?;
    }
    for (column, value) in table.columns.iter().zip(row.iter()) {
        ctx.check_domain(value, column.ty)?;
        if !column.nullable && value.is_null() {
            return Err(PgError {
                detail: Some(format!("Failing row contains ({}).", row_text(row))),
                objects: Some(Box::new(ErrorObjects {
                    schema: Some(table.schema.clone()),
                    table: Some(table.name.clone()),
                    column: Some(column.name.clone()),
                    ..ErrorObjects::default()
                })),
                ..PgError::new(
                    code::NOT_NULL_VIOLATION,
                    format!(
                        "null value in column \"{}\" of relation \"{}\" violates not-null constraint",
                        column.name, table.name
                    ),
                )
            });
        }
    }
    for (name, condition) in &rules.checks {
        if condition.eval(ctx, row)? == Value::Bool(false) {
            return Err(PgError {
                detail: Some(format!("Failing row contains ({}).", row_text(row))),
                objects: Some(Box::new(ErrorObjects {
                    schema: Some(table.schema.clone()),
                    table: Some(table.name.clone()),
                    constraint: Some(name.clone()),
                    ..ErrorObjects::default()
                })),
                ..PgError::new(
                    code::CHECK_VIOLATION,
                    format!("new row for relation \"{}\" violates check constraint \"{name}\"", table.name),
                )
            });
        }
    }
    Ok(())
}

/// parse_expression parses the SQL text of a stored expression.
pub fn parse_expression(text: &str) -> Result<pg_query::Node> {
    parse_expressions(text)?.into_iter().next().ok_or_else(|| PgError::internal("an empty expression"))
}

/// parse_expressions parses the SQL text of a comma-separated list of stored expressions.
pub fn parse_expressions(text: &str) -> Result<Vec<pg_query::Node>> {
    let result = pg_query::parse(&format!("SELECT {text}")).map_err(PgError::internal)?;
    let statement = result.protobuf.stmts.into_iter().next().and_then(|s| s.stmt).and_then(|s| s.node);
    let Some(NodeEnum::SelectStmt(select)) = statement else {
        return Err(PgError::internal(format!("a stored expression that is not one: {text}")));
    };
    let mut out = Vec::with_capacity(select.target_list.len());
    for target in select.target_list {
        match target.node {
            Some(NodeEnum::ResTarget(target)) => {
                out.push(target.val.map(|v| *v).ok_or_else(|| PgError::internal("an empty expression"))?)
            }
            _ => return Err(PgError::internal(format!("a stored expression that is not one: {text}"))),
        }
    }
    Ok(out)
}

/// duplicate_key returns Postgres' error for a row whose primary key already exists.
fn duplicate_key(table: &TableDef, row: &[Value]) -> PgError {
    let names: Vec<&str> = table.key_columns.iter().map(|&i| table.columns[i].name.as_str()).collect();
    let values: Vec<Value> = table.key_columns.iter().map(|&i| row[i].clone()).collect();
    PgError {
        detail: Some(format!("Key ({})=({}) already exists.", names.join(", "), row_text(&values))),
        objects: Some(Box::new(ErrorObjects {
            schema: Some(table.schema.clone()),
            table: Some(table.name.clone()),
            constraint: Some(table.primary_name()),
            ..ErrorObjects::default()
        })),
        ..PgError::new(
            code::UNIQUE_VIOLATION,
            format!("duplicate key value violates unique constraint \"{}\"", table.primary_name()),
        )
    }
}

/// unique_violation returns Postgres' error for a row, extended with its hidden expression columns, that duplicates
/// another's values in a unique index.
pub(crate) fn unique_violation(table: &TableDef, rules: &IndexRules, index: &IndexDef, row: &[Value]) -> PgError {
    let names: Vec<&str> = index.columns.iter().map(|&i| rules.column_name(table, i)).collect();
    let values: Vec<Value> = index.columns.iter().map(|&i| row[table.row_position(i)].clone()).collect();
    PgError {
        detail: Some(format!("Key ({})=({}) already exists.", names.join(", "), row_text(&values))),
        objects: Some(Box::new(ErrorObjects {
            schema: Some(table.schema.clone()),
            table: Some(table.name.clone()),
            constraint: Some(index.name.clone()),
            ..ErrorObjects::default()
        })),
        ..PgError::new(
            code::UNIQUE_VIOLATION,
            format!("duplicate key value violates unique constraint \"{}\"", index.name),
        )
    }
}

/// KeyEdits are changes to an index by key, where None deletes the key.
type KeyEdits = Vec<(Vec<u8>, Option<Vec<u8>>)>;

/// PrefixPositions are the positions of an index's edits by the bytes of their indexed values.
type PrefixPositions = std::collections::HashMap<Vec<Option<Vec<u8>>>, Vec<usize>>;

/// Kept is what an update keeps of a row whose primary key it keeps: whether each secondary index keeps the row's key,
/// and the row's encoded primary key and new value, which is None when the update leaves the stored row as it was.
struct Kept {
    indexes: Vec<bool>,
    key: Vec<u8>,
    value: Option<Vec<u8>>,
}

/// Edits collects changes to a table's primary index and secondary indexes by key.
struct Edits<'a> {
    table: &'a TableDef,
    rules: IndexRules,
    edits: Vec<(Vec<u8>, Option<Vec<u8>>)>,
    /// The position of each key's last edit by the key's bytes, for a table whose equal keys have equal bytes.
    latest: Option<std::collections::HashMap<Vec<u8>, usize>>,
    /// Each secondary index's changes, in the order of the table's indexes.
    index_edits: Vec<KeyEdits>,
    /// The positions of each secondary index's changes by the bytes of their indexed values, for an index whose equal
    /// keys have equal bytes.
    index_positions: Vec<Option<PrefixPositions>>,
    /// Whether each secondary index's unique constraint is deferred, whose checks wait for the commit.
    deferred: Vec<bool>,
    /// The unique checks that deferred constraints owe, by index position and indexed values.
    deferred_checks: Vec<(usize, Vec<Value>)>,
}

impl<'a> Edits<'a> {
    /// new starts collecting changes to the table.
    fn new(ctx: &mut Ctx<'_>, table: &'a TableDef) -> Result<Edits<'a>> {
        Ok(Edits {
            table,
            rules: ctx.index_rules(table)?,
            edits: Vec::new(),
            latest: table.key_columns.iter().all(|&c| canonical(table, c)).then(std::collections::HashMap::new),
            index_edits: vec![Vec::new(); table.indexes.len()],
            index_positions: table
                .indexes
                .iter()
                .map(|index| {
                    table
                        .index_key_columns(index)
                        .iter()
                        .all(|&c| canonical(table, c))
                        .then(std::collections::HashMap::new)
                })
                .collect(),
            deferred: vec![false; table.indexes.len()],
            deferred_checks: Vec::new(),
        })
    }

    /// deferring starts collecting changes to the table, leaving the checks of the unique constraints that are deferred
    /// now to the commit.
    fn deferring(ctx: &mut Ctx<'_>, table: &'a TableDef) -> Result<Edits<'a>> {
        let mut edits = Edits::new(ctx, table)?;
        edits.deferred = table
            .indexes
            .iter()
            .map(|i| i.unique && ctx.is_deferred(&table.schema, &i.name, i.deferrable, i.initially_deferred))
            .collect();
        Ok(edits)
    }

    /// owe records the unique checks that the deferred constraints of the collected changes owe.
    fn owe(&mut self, ctx: &mut Ctx<'_>) {
        for (i, values) in std::mem::take(&mut self.deferred_checks) {
            let (schema, table) = (self.table.schema.clone(), self.table.name.clone());
            ctx.defer(crate::deferred::Pending::Unique(schema, table, self.table.indexes[i].name.clone(), values));
        }
    }

    /// index_taken reports whether another row already has a key's indexed values in a unique index.
    fn index_taken(&self, db: &mut Database, i: usize, key: &[u8]) -> Result<bool> {
        Ok(self.index_match(db, i, key)?.is_some())
    }

    /// index_match returns the key of another row that has a key's indexed values in a unique index.
    fn index_match(&self, db: &mut Database, i: usize, key: &[u8]) -> Result<Option<Vec<u8>>> {
        let index = &self.table.indexes[i];
        let width = index.columns.len();
        let same = |k: &[u8]| self.table.compare_index_prefix(index, width, k, key) == Ordering::Equal;
        let edits = &self.index_edits[i];
        let candidates: Vec<usize> = match &self.index_positions[i] {
            Some(positions) => positions.get(&key_prefix(key, width)).cloned().unwrap_or_default(),
            None => (0..edits.len()).filter(|&j| same(&edits[j].0)).collect(),
        };
        let latest = |k: &[u8]| {
            candidates
                .iter()
                .rev()
                .map(|&j| &edits[j])
                .find(|(p, _)| self.table.compare_index_keys(index, p, k) == Ordering::Equal)
                .map(|(_, v)| v.is_some())
        };
        if let Some(&j) = candidates.iter().find(|&&j| latest(&edits[j].0) == Some(true)) {
            return Ok(Some(edits[j].0.clone()));
        }
        let mut found = None;
        let root = db.read(&index.root)?;
        prolly::scan_from(db, root, key, &|a, b| self.table.compare_index_prefix(index, width, a, b), &mut |k, _| {
            if !same(k) {
                return Ok(false);
            }
            if latest(k) != Some(false) {
                found = Some(k.to_vec());
            }
            Ok(found.is_none())
        })?;
        Ok(found)
    }

    /// primary_of returns the primary key tuple that an index key ends with.
    fn primary_of(&self, index: &IndexDef, key: &[u8]) -> Result<Vec<u8>> {
        let columns = self.table.index_key_columns(index);
        let tuple = Tuple(key);
        let fields: Vec<Option<&[u8]>> = if self.table.keyless() {
            vec![tuple.field(columns.len() - 1)?]
        } else {
            self.table
                .key_columns
                .iter()
                .map(|c| {
                    let position = columns.iter().position(|k| k == c).unwrap_or(0);
                    tuple.field(position)
                })
                .collect::<std::result::Result<_, _>>()?
        };
        Ok(prolly::val::build_tuple(&fields))
    }

    /// conflicting_row returns an existing row that a row would duplicate in the target's unique constraints.
    fn conflicting_row(&self, ctx: &mut Ctx<'_>, row: &[Value], target: &ConflictTarget) -> Result<Option<Vec<Value>>> {
        let (row, held) = self.rules.indexed(ctx, row)?;
        let db = &mut *ctx.db;
        if !self.table.keyless() && matches!(target, ConflictTarget::Any | ConflictTarget::Primary) {
            let (key, _) = self.table.encode_row(db, &row)?;
            if let Some(value) = self.current(db, &key)? {
                return Ok(Some(self.table.decode_row(db, &key, &value)?.0));
            }
        }
        for (i, index) in self.table.indexes.iter().enumerate() {
            let wanted = match target {
                ConflictTarget::Any => index.unique,
                ConflictTarget::Index(t) => *t == i,
                ConflictTarget::Primary => false,
            };
            if !wanted || !held[i] || index.columns.iter().any(|&c| row[self.table.row_position(c)].is_null()) {
                continue;
            }
            let (primary, _) = self.table.encode_row(db, &row)?;
            let key = self.table.index_key(db, index, &row, &primary)?;
            if let Some(found) = self.index_match(db, i, &key)? {
                let primary = self.primary_of(index, &found)?;
                if let Some(value) = self.current(db, &primary)? {
                    return Ok(Some(self.table.decode_row(db, &primary, &value)?.0));
                }
            }
        }
        Ok(None)
    }

    /// index_row adds a row's keys to the secondary indexes that hold it, or removes them, checking unique indexes as
    /// it adds.
    fn index_row(&mut self, ctx: &mut Ctx<'_>, row: &[Value], primary: &[u8], add: bool) -> Result<()> {
        self.index_row_with_kept(ctx, row, primary, add, &[])
    }

    /// index_row_with_kept is `index_row` leaving alone the indexes marked kept, whose keys an update leaves as they
    /// were.
    fn index_row_with_kept(
        &mut self,
        ctx: &mut Ctx<'_>,
        row: &[Value],
        primary: &[u8],
        add: bool,
        kept: &[bool],
    ) -> Result<()> {
        let (row, held) = self.rules.indexed(ctx, row)?;
        let db = &mut *ctx.db;
        for i in (0..self.table.indexes.len()).filter(|&i| held[i] && !kept.get(i).copied().unwrap_or(false)) {
            let index = &self.table.indexes[i];
            let key = self.table.index_key(db, index, &row, primary)?;
            let value = |c: usize| &row[self.table.row_position(c)];
            let checked = add && index.unique && index.columns.iter().all(|&c| !value(c).is_null());
            if checked && self.deferred[i] {
                self.deferred_checks.push((i, index.columns.iter().map(|&c| value(c).clone()).collect()));
            } else if checked && self.index_taken(db, i, &key)? {
                return Err(unique_violation(self.table, &self.rules, index, &row));
            }
            let value = add.then(|| prolly::val::build_tuple(&[]));
            if let Some(positions) = self.index_positions[i].as_mut() {
                positions.entry(key_prefix(&key, index.columns.len())).or_default().push(self.index_edits[i].len());
            }
            self.index_edits[i].push((key, value));
        }
        Ok(())
    }

    /// push records an edit of a key, where a value of None deletes it.
    fn push(&mut self, key: Vec<u8>, value: Option<Vec<u8>>) {
        if let Some(latest) = self.latest.as_mut() {
            latest.insert(key.clone(), self.edits.len());
        }
        self.edits.push((key, value));
    }

    /// pending returns the value an earlier edit gave a key: None when no edit touched it, and Some(None) when one
    /// deleted it.
    fn pending(&self, key: &[u8]) -> Option<Option<Vec<u8>>> {
        if let Some(latest) = &self.latest {
            return latest.get(key).map(|&i| self.edits[i].1.clone());
        }
        self.edits
            .iter()
            .rev()
            .find(|(k, _)| self.table.compare_keys(k, key) == Ordering::Equal)
            .map(|(_, v)| v.clone())
    }

    /// current returns the value of a key with the edits so far applied.
    fn current(&self, db: &mut Database, key: &[u8]) -> Result<Option<Vec<u8>>> {
        if let Some(value) = self.pending(key) {
            return Ok(value);
        }
        let root = Arc::new(prolly::Node::decode(self.table.table.primary_index.clone())?);
        Ok(get(db, root, key, &|a, b| self.table.compare_keys(a, b))?)
    }

    /// insert adds a row, failing on a duplicate primary key, and adding to the cardinality of a keyless row unless a
    /// unique index holds it.
    fn insert(&mut self, ctx: &mut Ctx<'_>, row: &[Value]) -> Result<()> {
        let (key, mut value) = self.table.encode_row(ctx.db, row)?;
        if let Some(existing) = self.current(ctx.db, &key)? {
            if !self.table.keyless() {
                return Err(duplicate_key(self.table, row));
            }
            let (indexed, held) = self.rules.indexed(ctx, row)?;
            for (i, index) in self.table.indexes.iter().enumerate().filter(|(i, index)| held[*i] && index.unique) {
                let values: Vec<Value> =
                    index.columns.iter().map(|&c| indexed[self.table.row_position(c)].clone()).collect();
                if values.iter().any(Value::is_null) {
                    continue;
                }
                if !self.deferred[i] {
                    return Err(unique_violation(self.table, &self.rules, index, &indexed));
                }
                self.deferred_checks.push((i, values));
            }
            value = with_cardinality(&existing, cardinality(&existing) + 1);
        } else {
            self.index_row(ctx, row, &key, true)?;
        }
        self.push(key, Some(value));
        Ok(())
    }

    /// delete removes one copy of a row.
    fn delete(&mut self, ctx: &mut Ctx<'_>, row: &[Value]) -> Result<()> {
        let (key, _) = self.table.encode_row(ctx.db, row)?;
        if self.table.keyless()
            && let Some(existing) = self.current(ctx.db, &key)?
            && cardinality(&existing) > 1
        {
            let value = with_cardinality(&existing, cardinality(&existing) - 1);
            self.push(key, Some(value));
            return Ok(());
        }
        self.index_row(ctx, row, &key, false)?;
        self.push(key, None);
        Ok(())
    }

    /// kept returns what an update of a row keeps when it keeps the row's primary key: whether each secondary index
    /// keeps the row's key, with the row's encoded key and new value, or None when the update changes the primary key
    /// or the table is keyless.
    fn kept(&self, ctx: &mut Ctx<'_>, old: &[Value], new: &[Value]) -> Result<Option<Kept>> {
        if self.table.keyless() {
            return Ok(None);
        }
        let (key, value) = self.table.encode_row(ctx.db, new)?;
        let (old_key, old_value) = self.table.encode_row(ctx.db, old)?;
        if old_key != key {
            return Ok(None);
        }
        if old_value == value {
            return Ok(Some(Kept { indexes: vec![true; self.table.indexes.len()], key, value: None }));
        }
        let (old, old_held) = self.rules.indexed(ctx, old)?;
        let (new, new_held) = self.rules.indexed(ctx, new)?;
        let indexes = (self.table.indexes.iter().enumerate())
            .map(|(i, index)| {
                old_held[i] == new_held[i]
                    && index.columns.iter().all(|&c| {
                        let position = self.table.row_position(c);
                        old[position] == new[position]
                    })
            })
            .collect();
        Ok(Some(Kept { indexes, key, value: Some(value) }))
    }

    /// retire removes the keys of a row that an update changes from the secondary indexes, where the update keeps the
    /// row's primary key.
    fn retire(&mut self, ctx: &mut Ctx<'_>, row: &[Value], kept: &Kept) -> Result<()> {
        if kept.value.is_none() {
            return Ok(());
        }
        self.index_row_with_kept(ctx, row, &kept.key, false, &kept.indexes)
    }

    /// replace writes the new values of a row whose primary key an update keeps, adding its changed keys to the
    /// secondary indexes.
    fn replace(&mut self, ctx: &mut Ctx<'_>, row: &[Value], kept: Kept) -> Result<()> {
        let Some(value) = kept.value else { return Ok(()) };
        self.index_row_with_kept(ctx, row, &kept.key, true, &kept.indexes)?;
        self.push(kept.key, Some(value));
        Ok(())
    }

    /// apply writes the edits to the table as the working root now holds it and the table to the transaction's working
    /// root.
    fn apply(self, db: &mut Database, txn: &mut Txn) -> Result<()> {
        if self.edits.is_empty() {
            return Ok(());
        }
        let table = self.table;
        let index_edits = self.index_edits;
        let mut edits = self.edits;
        // A stable sort keeps the edits of one key in order, and the last of them wins.
        edits.sort_by(|a, b| table.compare_keys(&a.0, &b.0));
        let mut merged: Vec<(Vec<u8>, Option<Vec<u8>>)> = Vec::with_capacity(edits.len());
        for edit in edits {
            match merged.last_mut() {
                Some(last) if table.compare_keys(&last.0, &edit.0) == Ordering::Equal => *last = edit,
                _ => merged.push(edit),
            }
        }
        let current = txn.table(db, &table.schema, &table.name)?;
        let current = current.as_ref().unwrap_or(table);
        let mut stored = current.table.clone();
        stored.edit_rows(
            db,
            merged,
            &|a, b| table.compare_keys(a, b),
            (&table.key_encodings(), &table.value_encodings()),
        )?;
        let mut rebuilt: Option<TableDef> = None;
        for (index, mut edits) in current.indexes.iter().zip(index_edits) {
            if edits.is_empty() {
                continue;
            }
            if let Some(distance) = index.vector {
                let rebuilt = rebuilt.get_or_insert_with(|| TableDef { table: stored.clone(), ..current.clone() });
                let root = rebuilt.write_vector_index(db, index, distance)?;
                stored.put_index(db, &index.name, Some(root))?;
                continue;
            }
            let compare = |a: &[u8], b: &[u8]| table.compare_index_keys(index, a, b);
            edits.sort_by(|a, b| compare(&a.0, &b.0));
            let mut merged: Vec<(Vec<u8>, Option<Vec<u8>>)> = Vec::with_capacity(edits.len());
            for edit in edits {
                match merged.last_mut() {
                    Some(last) if compare(&last.0, &edit.0) == Ordering::Equal => *last = edit,
                    _ => merged.push(edit),
                }
            }
            stored.edit_index(db, &index.name, index.root, merged, &compare, &table.index_encodings(index))?;
        }
        let address = stored.write(db)?;
        txn.root.put_table(db, &table.schema, &table.name, Some(address))?;
        Ok(())
    }
}

/// canonical reports whether equal values of a table column, or of a keyless table's row hash, always have equal bytes
/// in a key.
fn canonical(table: &TableDef, column: usize) -> bool {
    column == crate::catalog::table::KEYLESS_HASH
        || table.columns.get(column).is_some_and(|c| {
            matches!(
                c.ty.oid,
                oid::INT2
                    | oid::INT4
                    | oid::INT8
                    | oid::OID
                    | oid::TEXT
                    | oid::VARCHAR
                    | oid::NAME
                    | oid::BOOL
                    | oid::UUID
                    | oid::DATE
                    | oid::TIMESTAMP
                    | oid::TIMESTAMPTZ
                    | oid::BYTEA
            )
        })
}

/// key_prefix returns the bytes of the first fields of a key.
fn key_prefix(key: &[u8], width: usize) -> Vec<Option<Vec<u8>>> {
    (0..width).map(|i| Tuple(key).field(i).ok().flatten().map(<[u8]>::to_vec)).collect()
}

/// cardinality returns the cardinality of a keyless row's value.
fn cardinality(value: &[u8]) -> u64 {
    Tuple(value).field(0).ok().flatten().and_then(|f| f.try_into().ok()).map_or(1, u64::from_le_bytes)
}

/// with_cardinality returns a keyless row's value with another cardinality.
fn with_cardinality(value: &[u8], cardinality: u64) -> Vec<u8> {
    let mut value = value.to_vec();
    value[..8].copy_from_slice(&cardinality.to_le_bytes());
    value
}

impl Ctx<'_> {
    /// row_rules binds a table's column defaults and check constraints.
    pub fn row_rules(&mut self, table: &TableDef) -> Result<RowRules> {
        let mut defaults = Vec::with_capacity(table.columns.len());
        let mut generated = Vec::new();
        for (i, column) in table.columns.iter().enumerate() {
            if column.generated {
                let node = parse_expression(&column.default)?;
                let bound = Binder::new(self, table_scope(table, None)).bind(&node)?;
                generated.push((i, assign(bound, column.ty, &column.name, -1)?.0));
                defaults.push(None);
                continue;
            }
            let domain_default = match crate::usertypes::get(column.ty.oid).map(|t| t.kind.clone()) {
                Some(crate::usertypes::Kind::Domain(domain)) => domain.default,
                _ => None,
            };
            let Some(default) = (!column.default.is_empty()).then(|| column.default.clone()).or(domain_default) else {
                defaults.push(None);
                continue;
            };
            let node = parse_expression(&default)?;
            let bound = Binder::new(self, Scope::default()).bind(&node)?;
            let expr = assign(bound, ColumnType { modifier: -1, ..column.ty }, &column.name, -1)?.0;
            defaults.push(Some(match column.ty.modifier {
                -1 => expr,
                _ => Expr::Cast(Box::new(expr), column.ty, false),
            }));
        }
        let mut checks = Vec::with_capacity(table.checks.len());
        for check in &table.checks {
            let node = parse_expression(&check.expression)?;
            let bound = Binder::new(self, table_scope(table, None)).bind(&node)?;
            checks.push((check.name.clone(), coerce(bound, typ(oid::BOOL), false, -1)?.0));
        }
        Ok(RowRules { defaults, checks, generated })
    }

    /// index_rules binds a table's hidden expression columns and index predicates.
    pub(crate) fn index_rules(&mut self, table: &TableDef) -> Result<IndexRules> {
        let mut rules = IndexRules::default();
        for column in &table.hidden {
            let bound = Binder::new(self, table_scope(table, None)).bind(&parse_expression(&column.default)?)?;
            rules.hidden.push(assign(bound, column.ty, &column.name, -1)?.0);
            let columns = table.columns.iter().map(|c| (c.name.clone(), c.ty)).collect();
            rules.names.push(crate::ruleutils::Analyzer::new(self, columns).index_column(&column.default, true)?);
        }
        for index in &table.indexes {
            rules.predicates.push(match index.predicate.is_empty() {
                true => None,
                false => {
                    let node = parse_expression(&index.predicate)?;
                    let bound = Binder::new(self, table_scope(table, None)).bind(&node)?;
                    Some(coerce(bound, typ(oid::BOOL), false, -1)?.0)
                }
            });
        }
        Ok(rules)
    }

    /// with_queries runs a function with the WITH queries of a statement's WITH clause in scope, as plan_query brings
    /// them into scope for a SELECT.
    fn with_queries<T>(
        &mut self,
        with: Option<&pg_query::protobuf::WithClause>,
        f: impl FnOnce(&mut Ctx<'_>) -> Result<T>,
    ) -> Result<T> {
        let Some(with) = with else { return f(self) };
        let depth = self.ctes.len();
        let result = Planner { ctx: self, outer: Vec::new() }.plan_with(with).and_then(|_| f(self));
        self.ctes.truncate(depth);
        result
    }

    /// plan_insert plans an INSERT, with the WITH queries it defines in scope.
    pub fn plan_insert(&mut self, insert: &InsertStmt) -> Result<InsertPlan> {
        self.with_queries(insert.with_clause.as_ref(), |ctx| ctx.plan_insert_statement(insert))
    }

    /// plan_insert_statement plans an INSERT whose WITH queries are in scope.
    fn plan_insert_statement(&mut self, insert: &InsertStmt) -> Result<InsertPlan> {
        let relation = insert.relation.as_ref().ok_or_else(|| PgError::internal("INSERT without a table"))?;
        if let Some(catalog) = self.system_catalog(&relation.schemaname, &relation.relname) {
            return Err(self.catalog_insert_error(catalog, insert)?);
        }
        let table = self.resolve_target(relation, "a")?;
        let object = Object::Table(table.schema.clone(), table.name.clone());
        self.require(&object, "a", relation.location)?;
        if insert.on_conflict_clause.as_ref().is_some_and(|c| !c.target_list.is_empty()) {
            self.require(&object, "w", relation.location)?;
        }
        if !insert.returning_list.is_empty() || insert.on_conflict_clause.is_some() {
            self.require(&object, "r", relation.location)?;
        }
        let alias = relation.alias.as_ref().map(|a| a.aliasname.clone());
        let targets: Vec<usize> = if insert.cols.is_empty() {
            (0..table.columns.len()).collect()
        } else {
            insert
                .cols
                .iter()
                .map(|col| {
                    let Some(NodeEnum::ResTarget(target)) = col.node.as_ref() else {
                        return Err(PgError::internal("an INSERT column that is not a target"));
                    };
                    table.columns.iter().position(|c| c.name == target.name).ok_or_else(|| PgError {
                        position: position(target.location),
                        ..PgError::new(
                            code::UNDEFINED_COLUMN,
                            format!("column \"{}\" of relation \"{}\" does not exist", target.name, table.name),
                        )
                    })
                })
                .collect::<Result<_>>()?
        };
        let overriding = pg_query::protobuf::OverridingKind::try_from(insert.r#override)
            .unwrap_or(pg_query::protobuf::OverridingKind::OverridingNotSet);
        let empty_row = pg_query::Node { node: Some(NodeEnum::List(pg_query::protobuf::List { items: Vec::new() })) };
        let default_values = pg_query::protobuf::SelectStmt { values_lists: vec![empty_row], ..Default::default() };
        let select = match insert.select_stmt.as_deref().and_then(|n| n.node.as_ref()) {
            Some(NodeEnum::SelectStmt(select)) => select.as_ref(),
            None => &default_values,
            _ => return Err(PgError::unsupported("this INSERT")),
        };
        let too_many = |location: i32| PgError {
            position: position(location),
            ..PgError::new(code::SYNTAX_ERROR, "INSERT has more expressions than target columns")
        };
        let set_row = match select.values_lists.as_slice() {
            [list] => match list.node.as_ref() {
                Some(NodeEnum::List(list)) if list.items.iter().any(calls_set_function) => Some(list.items.clone()),
                _ => None,
            },
            _ => None,
        };
        let source = if select.values_lists.is_empty()
            || !select.sort_clause.is_empty()
            || select.limit_count.is_some()
            || set_row.is_some()
        {
            let projected;
            let select = match set_row {
                Some(items) => {
                    let target_list = items
                        .into_iter()
                        .map(|item| pg_query::Node {
                            node: Some(NodeEnum::ResTarget(Box::new(pg_query::protobuf::ResTarget {
                                val: Some(Box::new(item)),
                                ..Default::default()
                            }))),
                        })
                        .collect();
                    projected = pg_query::protobuf::SelectStmt { target_list, ..Default::default() };
                    &projected
                }
                None => select,
            };
            for (target, &column) in select.target_list.iter().zip(&targets) {
                let Some(NodeEnum::ResTarget(target)) = target.node.as_ref() else { continue };
                let Some(NodeEnum::ParamRef(param)) = target.val.as_deref().and_then(|v| v.node.as_ref()) else {
                    continue;
                };
                let index = param.number as usize - 1;
                if self.parameters.len() <= index {
                    self.parameters.resize(index + 1, 0);
                }
                if self.parameters[index] == 0 {
                    self.parameters[index] = table.columns[column].ty.oid;
                }
            }
            let query = Planner { ctx: self, outer: Vec::new() }.plan_query(select)?;
            if query.columns.len() > targets.len() {
                return Err(too_many(-1));
            }
            for (ty, &target) in query.types.iter().zip(&targets) {
                let column = &table.columns[target];
                if column.generated {
                    return Err(generated_error(&column.name, "cannot insert a non-DEFAULT value into column"));
                }
                if column.identity != 0 && overriding == pg_query::protobuf::OverridingKind::OverridingUserValue {
                    return Err(PgError::unsupported("OVERRIDING USER VALUE with a query"));
                }
                if column.identity == b'a' && overriding != pg_query::protobuf::OverridingKind::OverridingSystemValue {
                    return Err(identity_error(&column.name, "cannot insert a non-DEFAULT value into column"));
                }
                assign((Expr::Column(0), *ty), column.ty, &column.name, -1)?;
            }
            InsertSource::Select(Box::new(query.plan))
        } else {
            let mut binder = Binder::new(self, Scope::default());
            binder.clause = "VALUES";
            let mut rows = Vec::new();
            for list in &select.values_lists {
                let Some(NodeEnum::List(list)) = list.node.as_ref() else { continue };
                if list.items.len() > targets.len() {
                    return Err(too_many(arg_location(&list.items[targets.len()])));
                }
                let mut row = Vec::new();
                for (item, &target) in list.items.iter().zip(&targets) {
                    let column = &table.columns[target];
                    if matches!(item.node.as_ref(), Some(NodeEnum::SetToDefault(_))) {
                        row.push(Expr::Default(target));
                        continue;
                    }
                    if column.generated {
                        return Err(generated_error(&column.name, "cannot insert a non-DEFAULT value into column"));
                    }
                    if column.identity != 0 && overriding == pg_query::protobuf::OverridingKind::OverridingUserValue {
                        row.push(Expr::Default(target));
                        continue;
                    }
                    if column.identity == b'a'
                        && overriding != pg_query::protobuf::OverridingKind::OverridingSystemValue
                    {
                        return Err(identity_error(&column.name, "cannot insert a non-DEFAULT value into column"));
                    }
                    let bound = binder.bind(item)?;
                    if let Expr::Param(i) = bound.0
                        && binder.ctx.parameters[i] == 0
                    {
                        binder.ctx.parameters[i] = column.ty.oid;
                    }
                    let bound = binder.reg_literal(bound, column.ty, arg_location(item))?;
                    row.push(assign(bound, column.ty, &column.name, arg_location(item))?.0);
                }
                rows.push(row);
            }
            InsertSource::Values(rows)
        };
        let rules = self.row_rules(&table)?;
        let on_conflict = match insert.on_conflict_clause.as_deref() {
            Some(clause) => Some(self.plan_on_conflict(&table, alias.as_deref(), clause)?),
            None => None,
        };
        let scope = table_scope(&table, alias.as_deref());
        let returning = self.plan_returning(scope, &insert.returning_list)?;
        Ok(InsertPlan { table, rules, targets, source, on_conflict, returning })
    }

    /// system_catalog returns the pg_catalog table that a schema and name denote, where an unqualified name means
    /// pg_catalog first unless the search path names pg_catalog itself.
    pub(crate) fn system_catalog(&self, schema: &str, name: &str) -> Option<&'static crate::pgcatalog::CatalogTable> {
        let implicit = schema.is_empty() && !self.session.search_path().iter().any(|s| s == "pg_catalog");
        (schema == "pg_catalog" || implicit)
            .then(|| crate::pgcatalog::lookup("pg_catalog", name))
            .flatten()
            .filter(|c| c.kind == "r")
    }

    /// catalog_insert_error returns Postgres' error for inserting into a system catalog when the first row leaves a
    /// column NULL that the catalog requires, and otherwise reports that Doltgres cannot change system catalogs.
    fn catalog_insert_error(
        &mut self,
        catalog: &crate::pgcatalog::CatalogTable,
        insert: &InsertStmt,
    ) -> Result<PgError> {
        let mut row = vec![Value::Null; catalog.columns.len()];
        let targets: Vec<usize> = match insert.cols.is_empty() {
            true => (0..catalog.columns.len()).collect(),
            false => insert
                .cols
                .iter()
                .filter_map(|c| match c.node.as_ref() {
                    Some(NodeEnum::ResTarget(t)) => catalog.columns.iter().position(|c| c.name == t.name),
                    _ => None,
                })
                .collect(),
        };
        let first = match insert.select_stmt.as_deref().and_then(|n| n.node.as_ref()) {
            Some(NodeEnum::SelectStmt(select)) => select.values_lists.first().cloned(),
            _ => None,
        };
        if let Some(NodeEnum::List(list)) = first.and_then(|n| n.node) {
            for (item, &target) in list.items.iter().zip(&targets) {
                row[target] = match Binder::new(self, Scope::default()).bind(item)?.0 {
                    Expr::Const(value) => value,
                    _ => Value::Text(String::new()),
                };
            }
        }
        let Some(column) = catalog.columns.iter().zip(&row).position(|(c, v)| c.not_null && v.is_null()) else {
            return Ok(PgError::unsupported("inserting into system catalogs"));
        };
        let shown: Vec<String> = row.iter().map(|v| v.output().unwrap_or_else(|| "null".into())).collect();
        Ok(PgError {
            detail: Some(format!("Failing row contains ({}).", shown.join(", "))),
            objects: Some(Box::new(crate::error::ErrorObjects {
                schema: Some(catalog.schema.into()),
                table: Some(catalog.name.into()),
                column: Some(catalog.columns[column].name.into()),
                ..Default::default()
            })),
            ..PgError::new(
                code::NOT_NULL_VIOLATION,
                format!(
                    "null value in column \"{}\" of relation \"{}\" violates not-null constraint",
                    catalog.columns[column].name, catalog.name
                ),
            )
        })
    }

    /// plan_copy plans inserting rows of values into columns of a table, as COPY FROM does.
    pub fn plan_copy(&mut self, table: TableDef, targets: Vec<usize>, rows: Vec<Vec<Value>>) -> Result<InsertPlan> {
        let rules = self.row_rules(&table)?;
        let rows = rows.into_iter().map(|row| row.into_iter().map(Expr::Const).collect()).collect();
        Ok(InsertPlan { table, rules, targets, source: InsertSource::Values(rows), on_conflict: None, returning: None })
    }

    /// plan_returning binds a RETURNING list over a scope.
    fn plan_returning(&mut self, scope: Scope, list: &[pg_query::Node]) -> Result<Option<Returning>> {
        if list.is_empty() {
            return Ok(None);
        }
        let mut binder = Binder::new(self, scope.clone());
        binder.clause = "RETURNING";
        let mut exprs = Vec::new();
        let mut columns = Vec::new();
        for target in list {
            let Some(NodeEnum::ResTarget(target)) = target.node.as_ref() else { continue };
            let value = target.val.as_deref().ok_or_else(|| PgError::internal("a target without a value"))?;
            if let Some(NodeEnum::ColumnRef(c)) = value.node.as_ref()
                && matches!(c.fields.last().and_then(|f| f.node.as_ref()), Some(NodeEnum::AStar(_)))
            {
                let table = if c.fields.len() > 1 { c.fields.first().and_then(crate::expr::node_name) } else { None };
                for (i, column) in scope.columns.iter().enumerate() {
                    if !column.hidden && table.is_none_or(|t| column.table == t) {
                        exprs.push(Expr::Column(i));
                        columns.push(crate::query::column(column.name.clone(), column.ty));
                    }
                }
                continue;
            }
            let (expr, ty) = binder.bind(value)?;
            let name = if target.name.is_empty() { crate::expr::figure_name(value) } else { target.name.clone() };
            exprs.push(expr);
            columns.push(crate::query::column(name, ty));
        }
        Ok(Some(Returning { exprs, columns }))
    }

    /// plan_on_conflict plans an ON CONFLICT clause, whose target must name a unique constraint.
    fn plan_on_conflict(
        &mut self,
        table: &TableDef,
        alias: Option<&str>,
        clause: &pg_query::protobuf::OnConflictClause,
    ) -> Result<OnConflict> {
        use pg_query::protobuf::OnConflictAction;
        let no_match = || {
            PgError::new(
                code::INVALID_COLUMN_REFERENCE,
                "there is no unique or exclusion constraint matching the ON CONFLICT specification",
            )
        };
        let target = match clause.infer.as_deref() {
            None => ConflictTarget::Any,
            Some(infer) if !infer.conname.is_empty() => {
                if !table.keyless() && infer.conname == table.primary_name() {
                    ConflictTarget::Primary
                } else if let Some(i) =
                    table.indexes.iter().position(|ix| ix.unique && !ix.plain && ix.name == infer.conname)
                {
                    ConflictTarget::Index(i)
                } else {
                    return Err(PgError::new(
                        code::UNDEFINED_OBJECT,
                        format!("constraint \"{}\" for table \"{}\" does not exist", infer.conname, table.name),
                    ));
                }
            }
            Some(infer) => {
                let mut columns = Vec::new();
                for elem in &infer.index_elems {
                    let Some(NodeEnum::IndexElem(elem)) = elem.node.as_ref() else { continue };
                    let i = table.columns.iter().position(|c| c.name == elem.name).ok_or_else(|| PgError {
                        position: position(infer.location),
                        ..PgError::new(code::UNDEFINED_COLUMN, format!("column \"{}\" does not exist", elem.name))
                    })?;
                    columns.push(i);
                }
                let same = |a: &[usize]| {
                    let (mut x, mut y) = (a.to_vec(), columns.clone());
                    x.sort_unstable();
                    y.sort_unstable();
                    x.dedup();
                    y.dedup();
                    x == y
                };
                if !table.keyless() && same(&table.key_columns) {
                    ConflictTarget::Primary
                } else if let Some(i) =
                    table.indexes.iter().position(|ix| ix.unique && ix.predicate.is_empty() && same(&ix.columns))
                {
                    ConflictTarget::Index(i)
                } else {
                    return Err(no_match());
                }
            }
        };
        let action = if OnConflictAction::try_from(clause.action) == Ok(OnConflictAction::OnconflictUpdate) {
            if target == ConflictTarget::Any {
                return Err(PgError {
                    position: position(clause.location),
                    hint: Some("For example, ON CONFLICT (column_name).".into()),
                    ..PgError::new(
                        code::SYNTAX_ERROR,
                        "ON CONFLICT DO UPDATE requires inference specification or constraint name",
                    )
                });
            }
            let mut scope = table_scope(table, alias);
            scope.columns.extend(table_scope(table, Some("excluded")).columns);
            let mut binder = Binder::new(self, scope);
            let filter = match clause.where_clause.as_deref() {
                Some(node) => Some(coerce(binder.bind(node)?, typ(oid::BOOL), false, -1)?.0),
                None => None,
            };
            let assignments = bind_assignments(&mut binder, table, &clause.target_list)?;
            ConflictAction::Update { assignments, filter }
        } else {
            ConflictAction::Nothing
        };
        Ok(OnConflict { target, action })
    }

    /// plan_update plans an UPDATE, with its FROM list joined to the table's rows.
    pub fn plan_update(&mut self, update: &UpdateStmt) -> Result<UpdatePlan> {
        self.with_queries(update.with_clause.as_ref(), |ctx| ctx.plan_update_statement(update))
    }

    /// plan_update_statement plans an UPDATE whose WITH queries are in scope.
    fn plan_update_statement(&mut self, update: &UpdateStmt) -> Result<UpdatePlan> {
        let relation = update.relation.as_ref().ok_or_else(|| PgError::internal("UPDATE without a table"))?;
        let table = self.resolve_target(relation, "w")?;
        let object = Object::Table(table.schema.clone(), table.name.clone());
        self.require(&object, "w", relation.location)?;
        if update.where_clause.is_some() || !update.returning_list.is_empty() {
            self.require(&object, "r", relation.location)?;
        }
        let alias = relation.alias.as_ref().map(|a| a.aliasname.clone());
        let mut scope = table_scope(&table, alias.as_deref());
        let from = if update.from_clause.is_empty() {
            None
        } else {
            let (plan, from_scope) = Planner { ctx: self, outer: Vec::new() }.plan_from(&update.from_clause)?;
            scope.columns.extend(from_scope.columns);
            Some(Box::new(plan))
        };
        let mut binder = Binder::new(self, scope.clone());
        binder.clause = "WHERE";
        let filter = match update.where_clause.as_deref() {
            Some(node) => Some(crate::expr::condition(binder.bind(node)?, "WHERE", arg_location(node))?),
            None => None,
        };
        let assignments = bind_assignments(&mut binder, &table, &update.target_list)?;
        let rules = self.row_rules(&table)?;
        let returning = self.plan_returning(scope, &update.returning_list)?;
        Ok(UpdatePlan { table, rules, filter, assignments, from, returning })
    }

    /// plan_delete plans a DELETE, with its USING list joined to the table's rows.
    pub fn plan_delete(&mut self, delete: &DeleteStmt) -> Result<DeletePlan> {
        self.with_queries(delete.with_clause.as_ref(), |ctx| ctx.plan_delete_statement(delete))
    }

    /// plan_delete_statement plans a DELETE whose WITH queries are in scope.
    fn plan_delete_statement(&mut self, delete: &DeleteStmt) -> Result<DeletePlan> {
        let relation = delete.relation.as_ref().ok_or_else(|| PgError::internal("DELETE without a table"))?;
        let table = self.resolve_target(relation, "d")?;
        let object = Object::Table(table.schema.clone(), table.name.clone());
        self.require(&object, "d", relation.location)?;
        if delete.where_clause.is_some() || !delete.returning_list.is_empty() {
            self.require(&object, "r", relation.location)?;
        }
        let mut scope = table_scope(&table, relation.alias.as_ref().map(|a| a.aliasname.as_str()));
        let using = if delete.using_clause.is_empty() {
            None
        } else {
            let (plan, using_scope) = Planner { ctx: self, outer: Vec::new() }.plan_from(&delete.using_clause)?;
            scope.columns.extend(using_scope.columns);
            Some(Box::new(plan))
        };
        let mut binder = Binder::new(self, scope.clone());
        binder.clause = "WHERE";
        let filter = match delete.where_clause.as_deref() {
            Some(node) => Some(crate::expr::condition(binder.bind(node)?, "WHERE", arg_location(node))?),
            None => None,
        };
        let returning = self.plan_returning(scope, &delete.returning_list)?;
        Ok(DeletePlan { table, filter, using, returning })
    }
}

/// bind_assignments binds the SET list of an UPDATE or ON CONFLICT DO UPDATE.
fn bind_assignments(
    binder: &mut Binder<'_, '_>,
    table: &TableDef,
    list: &[pg_query::Node],
) -> Result<Vec<(usize, Expr)>> {
    let mut assignments = Vec::new();
    for target in list {
        let Some(NodeEnum::ResTarget(target)) = target.node.as_ref() else { continue };
        let i = table.columns.iter().position(|c| c.name == target.name).ok_or_else(|| PgError {
            position: position(target.location),
            ..PgError::new(
                code::UNDEFINED_COLUMN,
                format!("column \"{}\" of relation \"{}\" does not exist", target.name, table.name),
            )
        })?;
        let value = target.val.as_deref().ok_or_else(|| PgError::internal("an assignment without a value"))?;
        let column = &table.columns[i];
        if !target.indirection.is_empty() {
            let subscripted = matches!(target.indirection[0].node, Some(NodeEnum::AIndices(_)));
            if subscripted && !crate::array::is_array_type(column.ty.oid) {
                return Err(PgError {
                    position: position(target.location),
                    ..PgError::new(
                        code::DATATYPE_MISMATCH,
                        format!(
                            "cannot subscript type {} because it does not support subscripting",
                            crate::cast::type_display(column.ty.oid)
                        ),
                    )
                });
            }
            if !crate::array::is_array_type(column.ty.oid) {
                return Err(PgError::unsupported("this assignment"));
            }
            let (subscripts, slice) = binder.subscripts(&target.indirection)?;
            let element = crate::expr::element_type(column.ty.oid);
            let ty = if slice { column.ty } else { ColumnType { oid: element, modifier: column.ty.modifier } };
            let bound = binder.bind(value)?;
            let value = assign(bound, ty, &column.name, arg_location(value))?.0;
            let base = match assignments.iter().position(|(c, _)| *c == i) {
                Some(at) => assignments.remove(at).1,
                None => Expr::Column(i),
            };
            let expr = Expr::SubscriptAssign(Box::new(base), element, subscripts, slice, Box::new(value));
            assignments.push((i, expr));
            continue;
        }
        if assignments.iter().any(|(c, _)| *c == i) {
            return Err(PgError::new(
                code::SYNTAX_ERROR,
                format!("multiple assignments to same column \"{}\"", target.name),
            ));
        }
        let expr = if matches!(value.node.as_ref(), Some(NodeEnum::SetToDefault(_))) {
            Expr::Default(i)
        } else if column.generated {
            return Err(PgError {
                detail: Some(format!("Column \"{}\" is a generated column.", column.name)),
                ..PgError::new(
                    code::GENERATED_ALWAYS,
                    format!("column \"{}\" can only be updated to DEFAULT", column.name),
                )
            });
        } else if column.identity == b'a' {
            return Err(PgError {
                detail: Some(format!("Column \"{}\" is an identity column defined as GENERATED ALWAYS.", column.name)),
                ..PgError::new(
                    code::GENERATED_ALWAYS,
                    format!("column \"{}\" can only be updated to DEFAULT", column.name),
                )
            });
        } else {
            let bound = binder.bind(value)?;
            if let Expr::Param(p) = bound.0
                && binder.ctx.parameters[p] == 0
            {
                binder.ctx.parameters[p] = column.ty.oid;
            }
            let bound = binder.reg_literal(bound, column.ty, arg_location(value))?;
            assign(bound, column.ty, &column.name, arg_location(value))?.0
        };
        assignments.push((i, expr));
    }
    Ok(assignments)
}

/// returned evaluates a RETURNING list over the written rows.
fn returned(ctx: &mut Ctx<'_>, returning: &Returning, rows: &[Vec<Value>]) -> Result<Vec<Vec<Value>>> {
    rows.iter().map(|row| returning.exprs.iter().map(|e| e.eval(ctx, row)).collect()).collect()
}

/// outcome returns a statement's result: its RETURNING rows, or its command tag alone.
fn outcome(ctx: &mut Ctx<'_>, returning: &Option<Returning>, rows: &[Vec<Value>], tag: String) -> Result<Outcome> {
    match returning {
        Some(returning) => {
            let rows = returned(ctx, returning, rows)?;
            Ok(Outcome::Rows { columns: returning.columns.clone(), rows, tag })
        }
        None => Ok(Outcome::command(tag)),
    }
}

impl InsertPlan {
    /// run inserts the rows, resolving conflicts as the ON CONFLICT clause says.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Outcome> {
        let triggers = ctx.table_triggers(&self.table)?;
        triggers.statement(ctx, Event::Insert, BEFORE, &[])?;
        let outcome = self.insert(ctx, &triggers)?;
        triggers.statement(ctx, Event::Insert, AFTER, &[])?;
        Ok(outcome)
    }

    /// insert inserts the rows, running the row triggers around each.
    fn insert(&self, ctx: &mut Ctx<'_>, triggers: &Triggers) -> Result<Outcome> {
        let mut sources: Vec<Vec<Value>> = Vec::new();
        match &self.source {
            InsertSource::Values(rows) => {
                for row in rows {
                    let mut values = Vec::with_capacity(row.len());
                    for expr in row {
                        values.push(match expr {
                            Expr::Default(i) => default_value(ctx, &self.rules, *i)?,
                            expr => expr.eval(ctx, &[])?,
                        });
                    }
                    sources.push(values);
                }
            }
            InsertSource::Select(plan) => sources = plan.run(ctx)?,
        }
        let mut rows = Vec::with_capacity(sources.len());
        for source in &sources {
            let mut row = Vec::with_capacity(self.table.columns.len());
            for i in 0..self.table.columns.len() {
                row.push(match self.targets.iter().position(|&t| t == i) {
                    Some(j) if j < source.len() => cast_value(source[j].clone(), self.table.columns[i].ty, false)?,
                    _ => default_value(ctx, &self.rules, i)?,
                });
            }
            rows.push(row);
        }
        let Some(on_conflict) = &self.on_conflict else {
            let rows = if triggers.fires(Event::Insert, BEFORE, true) {
                let mut written = Vec::with_capacity(rows.len());
                for row in rows {
                    if let Some(row) = triggers.before_row(ctx, Event::Insert, None, Some(row), &[])? {
                        written.extend(insert_checked_rows(ctx, &self.table, &self.rules, vec![row])?);
                    }
                }
                written
            } else {
                insert_checked_rows(ctx, &self.table, &self.rules, rows)?
            };
            let changes: Vec<Change> = rows.iter().map(|r| (None, Some(r.clone()))).collect();
            ctx.enforce_foreign_keys(&self.table, &changes)?;
            for row in &rows {
                triggers.after_row(ctx, Event::Insert, None, Some(row), &[])?;
            }
            let tag = format!("INSERT 0 {}", rows.len());
            return outcome(ctx, &self.returning, &rows, tag);
        };
        let table = &self.table;
        let mut edits = Edits::deferring(ctx, table)?;
        let mut written = Vec::new();
        let mut changes: Vec<Change> = Vec::new();
        let mut touched: Vec<Vec<u8>> = Vec::new();
        for row in rows {
            let Some(mut row) = triggers.before_row(ctx, Event::Insert, None, Some(row), &[])? else { continue };
            check_row(ctx, table, &self.rules, &mut row)?;
            let Some(existing) = edits.conflicting_row(ctx, &row, &on_conflict.target)? else {
                edits.insert(ctx, &row)?;
                touched.push(table.encode_row(ctx.db, &row)?.0);
                changes.push((None, Some(row.clone())));
                written.push(row);
                continue;
            };
            let ConflictAction::Update { assignments, filter } = &on_conflict.action else { continue };
            let existing_key = table.encode_row(ctx.db, &existing)?.0;
            if touched.iter().any(|k| table.compare_keys(k, &existing_key) == Ordering::Equal) {
                return Err(PgError {
                    hint: Some(
                        "Ensure that no rows proposed for insertion within the same command have duplicate constrained values."
                            .into(),
                    ),
                    ..PgError::new(
                        code::CARDINALITY_VIOLATION,
                        "ON CONFLICT DO UPDATE command cannot affect row a second time",
                    )
                });
            }
            let mut combined = existing.clone();
            combined.extend(row.iter().cloned());
            if let Some(filter) = filter
                && !filter.is_true(ctx, &combined)?
            {
                continue;
            }
            let mut new_row = existing.clone();
            for (i, expr) in assignments {
                new_row[*i] = match expr {
                    Expr::Default(c) => default_value(ctx, &self.rules, *c)?,
                    expr => expr.eval(ctx, &combined)?,
                };
            }
            let updated: Vec<usize> = assignments.iter().map(|(i, _)| *i).collect();
            let Some(mut new_row) =
                triggers.before_row(ctx, Event::Update, Some(&existing), Some(new_row), &updated)?
            else {
                continue;
            };
            check_row(ctx, table, &self.rules, &mut new_row)?;
            edits.delete(ctx, &existing)?;
            edits.insert(ctx, &new_row)?;
            touched.push(table.encode_row(ctx.db, &new_row)?.0);
            changes.push((Some(existing), Some(new_row.clone())));
            written.push(new_row);
        }
        edits.owe(ctx);
        edits.apply(ctx.db, ctx.txn)?;
        ctx.enforce_foreign_keys(table, &changes)?;
        if let ConflictAction::Update { assignments, .. } = &on_conflict.action {
            let updated: Vec<usize> = assignments.iter().map(|(i, _)| *i).collect();
            for (old, new) in &changes {
                let event = if old.is_some() { Event::Update } else { Event::Insert };
                triggers.after_row(ctx, event, old.as_deref(), new.as_deref(), &updated)?;
            }
        } else {
            for (_, new) in &changes {
                triggers.after_row(ctx, Event::Insert, None, new.as_deref(), &[])?;
            }
        }
        let tag = format!("INSERT 0 {}", written.len());
        outcome(ctx, &self.returning, &written, tag)
    }
}

/// identity_error returns Postgres' error for a value given to an identity column that is GENERATED ALWAYS.
fn identity_error(column: &str, message: &str) -> PgError {
    PgError {
        detail: Some(format!("Column \"{column}\" is an identity column defined as GENERATED ALWAYS.")),
        hint: Some("Use OVERRIDING SYSTEM VALUE to override.".into()),
        ..PgError::new(code::GENERATED_ALWAYS, format!("{message} \"{column}\""))
    }
}

/// generated_error returns Postgres' error for a value given to a generated column.
fn generated_error(column: &str, message: &str) -> PgError {
    PgError {
        detail: Some(format!("Column \"{column}\" is a generated column.")),
        ..PgError::new(code::GENERATED_ALWAYS, format!("{message} \"{column}\""))
    }
}

/// default_value evaluates a column's default, which is NULL without one.
pub(crate) fn default_value(ctx: &mut Ctx<'_>, rules: &RowRules, column: usize) -> Result<Value> {
    match &rules.defaults[column] {
        Some(expr) => expr.eval(ctx, &[]),
        None => Ok(Value::Null),
    }
}

/// insert_checked_rows checks and inserts rows in the table's column order.
fn insert_checked_rows(
    ctx: &mut Ctx<'_>,
    table: &TableDef,
    rules: &RowRules,
    mut rows: Vec<Vec<Value>>,
) -> Result<Vec<Vec<Value>>> {
    for row in &mut rows {
        check_row(ctx, table, rules, row)?;
    }
    let mut edits = Edits::deferring(ctx, table)?;
    for row in &rows {
        edits.insert(ctx, row)?;
    }
    edits.owe(ctx);
    edits.apply(ctx.db, ctx.txn)?;
    Ok(rows)
}

/// write_rows inserts rows already in a table's column order and types, without checking them, as rebuilding a table
/// does.
pub fn write_rows(ctx: &mut Ctx<'_>, table: &TableDef, rows: &[Vec<Value>]) -> Result<()> {
    let mut edits = Edits::new(ctx, table)?;
    for row in rows {
        edits.insert(ctx, row)?;
    }
    edits.apply(ctx.db, ctx.txn)
}

/// delete_rows deletes rows read from a table.
pub fn delete_rows(ctx: &mut Ctx<'_>, table: &TableDef, rows: &[Vec<Value>]) -> Result<()> {
    let mut edits = Edits::new(ctx, table)?;
    for row in rows {
        edits.delete(ctx, row)?;
    }
    edits.apply(ctx.db, ctx.txn)
}

/// apply_changes checks and writes the changes that a foreign key's action makes to a referencing table.
pub(crate) fn apply_changes(ctx: &mut Ctx<'_>, table: &TableDef, changes: &[Change]) -> Result<()> {
    let rules = ctx.row_rules(table)?;
    let mut changes = changes.to_vec();
    for (_, new) in &mut changes {
        if let Some(row) = new {
            check_row(ctx, table, &rules, row)?;
        }
    }
    let mut edits = Edits::deferring(ctx, table)?;
    for (old, _) in &changes {
        if let Some(row) = old {
            edits.delete(ctx, row)?;
        }
    }
    for (_, new) in &changes {
        if let Some(row) = new {
            edits.insert(ctx, row)?;
        }
    }
    edits.owe(ctx);
    edits.apply(ctx.db, ctx.txn)
}

/// insert_rows converts rows to a table's column types and inserts them, for CREATE TABLE AS.
pub fn insert_rows(ctx: &mut Ctx<'_>, table: &TableDef, rows: Vec<Vec<Value>>) -> Result<()> {
    let rules = ctx.row_rules(table)?;
    let rows = rows
        .into_iter()
        .map(|row| {
            row.into_iter().zip(&table.columns).map(|(v, c)| cast_value(v, c.ty, false)).collect::<Result<Vec<_>>>()
        })
        .collect::<Result<Vec<_>>>()?;
    insert_checked_rows(ctx, table, &rules, rows).map(|_| ())
}

/// matches pairs each of a table's rows with one row of a FROM list that meets a filter, or with no row
/// without a FROM list, as UPDATE ... FROM and DELETE ... USING join them.
fn matches(
    ctx: &mut Ctx<'_>,
    table: &TableDef,
    from: &Option<Box<Plan>>,
    filter: &Option<Expr>,
) -> Result<Vec<(Vec<Value>, Vec<Value>)>> {
    if ctx.once.is_none() {
        ctx.once = Some(std::collections::HashMap::new());
        let result = matches(ctx, table, from, filter);
        ctx.once = None;
        return result;
    }
    if let Some(from) = from
        && !table.keyless()
    {
        let join = Plan::Join {
            left: Box::new(Plan::Scan(Box::new(table.clone()), None)),
            right: from.clone(),
            kind: JoinKind::Inner,
            condition: None,
            lateral: false,
            method: crate::plan::JoinMethod::Unplanned,
        };
        let plan = match filter {
            Some(filter) => Planner { ctx, outer: Vec::new() }.use_indexes(push_down(join, filter.clone())),
            None => join,
        };
        let (width, mut seen, mut out) = (table.columns.len(), crate::exec::Groups::new(), Vec::new());
        for mut row in plan.run(ctx)? {
            let from_row = row.split_off(width);
            let key: Vec<Value> = table.key_columns.iter().map(|&i| row[i].clone()).collect();
            if seen.insert(&key).1 {
                out.push((row, from_row));
            }
        }
        return Ok(out);
    }
    let from_rows = match from {
        Some(plan) => Some(plan.run(ctx)?),
        None => None,
    };
    let index_scan = match (&from_rows, filter) {
        (None, Some(filter)) => crate::indexscan::choose(ctx, table, filter),
        _ => None,
    };
    let rows = match index_scan {
        Some(index_scan) => index_scan.run(ctx)?,
        None => scan(ctx.db, table)?,
    };
    let mut out = Vec::new();
    for row in rows {
        match &from_rows {
            None => {
                if filter.as_ref().map_or(Ok(true), |f| f.is_true(ctx, &row))? {
                    out.push((row, Vec::new()));
                }
            }
            Some(from_rows) => {
                for from_row in from_rows {
                    let mut combined = row.clone();
                    combined.extend(from_row.iter().cloned());
                    if filter.as_ref().map_or(Ok(true), |f| f.is_true(ctx, &combined))? {
                        out.push((row.clone(), from_row.clone()));
                        break;
                    }
                }
            }
        }
    }
    Ok(out)
}

impl UpdatePlan {
    /// run updates the matching rows.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Outcome> {
        let triggers = ctx.table_triggers(&self.table)?;
        let updated: Vec<usize> = self.assignments.iter().map(|(i, _)| *i).collect();
        triggers.statement(ctx, Event::Update, BEFORE, &updated)?;
        let outcome = self.update(ctx, &triggers, &updated)?;
        triggers.statement(ctx, Event::Update, AFTER, &updated)?;
        Ok(outcome)
    }

    /// update updates the matching rows, running the row triggers around each.
    fn update(&self, ctx: &mut Ctx<'_>, triggers: &Triggers, updated: &[usize]) -> Result<Outcome> {
        let mut changes = Vec::new();
        for (row, from_row) in matches(ctx, &self.table, &self.from, &self.filter)? {
            let mut combined = row.clone();
            combined.extend(from_row.iter().cloned());
            let mut new_row = row.clone();
            for (i, expr) in &self.assignments {
                new_row[*i] = match expr {
                    Expr::Default(c) => default_value(ctx, &self.rules, *c)?,
                    expr => expr.eval(ctx, &combined)?,
                };
            }
            let Some(mut new_row) = triggers.before_row(ctx, Event::Update, Some(&row), Some(new_row), updated)? else {
                continue;
            };
            check_row(ctx, &self.table, &self.rules, &mut new_row)?;
            changes.push((row, new_row, from_row));
        }
        let mut edits = Edits::deferring(ctx, &self.table)?;
        let mut kept = Vec::with_capacity(changes.len());
        for (row, new_row, _) in &changes {
            let same = edits.kept(ctx, row, new_row)?;
            match &same {
                Some(same) => edits.retire(ctx, row, same)?,
                None => edits.delete(ctx, row)?,
            }
            kept.push(same);
        }
        for ((_, new_row, _), same) in changes.iter().zip(kept) {
            match same {
                Some(same) => edits.replace(ctx, new_row, same)?,
                None => edits.insert(ctx, new_row)?,
            }
        }
        edits.owe(ctx);
        edits.apply(ctx.db, ctx.txn)?;
        let edited: Vec<Change> =
            changes.iter().map(|(row, new_row, _)| (Some(row.clone()), Some(new_row.clone()))).collect();
        ctx.enforce_foreign_keys(&self.table, &edited)?;
        for (row, new_row, _) in &changes {
            triggers.after_row(ctx, Event::Update, Some(row), Some(new_row), updated)?;
        }
        let written: Vec<Vec<Value>> = changes
            .into_iter()
            .map(|(_, mut new_row, from_row)| {
                new_row.extend(from_row);
                new_row
            })
            .collect();
        let tag = format!("UPDATE {}", written.len());
        outcome(ctx, &self.returning, &written, tag)
    }
}

impl DeletePlan {
    /// run deletes the matching rows.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Outcome> {
        let triggers = ctx.table_triggers(&self.table)?;
        triggers.statement(ctx, Event::Delete, BEFORE, &[])?;
        let outcome = self.delete(ctx, &triggers)?;
        triggers.statement(ctx, Event::Delete, AFTER, &[])?;
        Ok(outcome)
    }

    /// delete deletes the matching rows, running the row triggers around each.
    fn delete(&self, ctx: &mut Ctx<'_>, triggers: &Triggers) -> Result<Outcome> {
        let mut doomed = Vec::new();
        for (row, using_row) in matches(ctx, &self.table, &self.using, &self.filter)? {
            if let Some(row) = triggers.before_row(ctx, Event::Delete, Some(&row), None, &[])? {
                doomed.push((row, using_row));
            }
        }
        let mut edits = Edits::new(ctx, &self.table)?;
        for (row, _) in &doomed {
            edits.delete(ctx, row)?;
        }
        edits.apply(ctx.db, ctx.txn)?;
        let changes: Vec<Change> = doomed.iter().map(|(row, _)| (Some(row.clone()), None)).collect();
        ctx.enforce_foreign_keys(&self.table, &changes)?;
        for (row, _) in &doomed {
            triggers.after_row(ctx, Event::Delete, Some(row), None, &[])?;
        }
        let deleted: Vec<Vec<Value>> = doomed
            .into_iter()
            .map(|(mut row, using_row)| {
                row.extend(using_row);
                row
            })
            .collect();
        let tag = format!("DELETE {}", deleted.len());
        outcome(ctx, &self.returning, &deleted, tag)
    }
}

/// calls_set_function reports whether an expression calls a set-returning function.
fn calls_set_function(node: &pg_query::Node) -> bool {
    let wrapped = NodeEnum::ResTarget(Box::new(pg_query::protobuf::ResTarget {
        val: Some(Box::new(node.clone())),
        ..Default::default()
    }));
    wrapped.nodes().into_iter().any(|(node, ..)| match node {
        pg_query::NodeRef::FuncCall(call) => call
            .funcname
            .iter()
            .filter_map(crate::expr::node_name)
            .next_back()
            .is_some_and(crate::functions::returns_set),
        _ => false,
    })
}

/// MergeAction is what a WHEN clause of MERGE does to a row.
enum MergeAction {
    Update(Vec<(usize, Expr)>),
    Delete,
    Insert(Vec<Option<Expr>>),
    Nothing,
}

/// MergeChange is a row change that MERGE makes: its event, the old and new rows, and the columns an update sets.
type MergeChange = (Event, Option<Vec<Value>>, Option<Vec<Value>>, Vec<usize>);

/// MergeClause is a WHEN clause of MERGE: whether it applies to a target row that a source row matches, its extra
/// condition, and its action, which see the target row and then the source row for a matched row, and only the
/// source row otherwise.
struct MergeClause {
    matched: bool,
    condition: Option<Expr>,
    action: MergeAction,
}

impl Ctx<'_> {
    /// merge runs MERGE: each source row that the join condition matches with target rows takes the first WHEN
    /// MATCHED clause whose condition holds for each of them, and each other source row the first WHEN NOT MATCHED
    /// clause whose condition holds, as Postgres' ExecMerge does.
    pub fn merge(&mut self, stmt: &pg_query::protobuf::MergeStmt) -> Result<Outcome> {
        let relation = stmt.relation.as_ref().ok_or_else(|| PgError::internal("MERGE without a table"))?;
        let table = match self.resolve_table(relation) {
            Ok(table) => table,
            Err(err) => match self.find_view(&relation.schemaname, &relation.relname)? {
                Some(_) => {
                    return Err(PgError {
                        detail: Some("This operation is not supported for views.".into()),
                        ..PgError::new(
                            code::FEATURE_NOT_SUPPORTED,
                            format!("cannot execute MERGE on relation \"{}\"", relation.relname),
                        )
                    });
                }
                None => return Err(err),
            },
        };
        let object = Object::Table(table.schema.clone(), table.name.clone());
        self.require(&object, "r", relation.location)?;
        let alias = relation.alias.as_ref().map(|a| a.aliasname.clone());
        let mut scope = table_scope(&table, alias.as_deref());
        let source = stmt.source_relation.as_deref().ok_or_else(|| PgError::internal("MERGE without a source"))?;
        let (source_plan, source_scope) =
            Planner { ctx: self, outer: Vec::new() }.plan_from(std::slice::from_ref(source))?;
        scope.columns.extend(source_scope.columns.iter().cloned());
        let mut binder = Binder::new(self, scope.clone());
        binder.clause = "JOIN/ON";
        let join = stmt.join_condition.as_deref().ok_or_else(|| PgError::internal("MERGE without a condition"))?;
        let join = crate::expr::condition(binder.bind(join)?, "JOIN/ON", arg_location(join))?;
        let target_name = alias.clone().unwrap_or_else(|| table.name.clone());
        let mut clauses = Vec::with_capacity(stmt.merge_when_clauses.len());
        let mut events = Vec::new();
        for clause in &stmt.merge_when_clauses {
            let bound = self.merge_clause(clause, &table, &scope, &source_scope, &mut events);
            clauses.push(bound.map_err(|err| hidden_target(err, &table, &target_name))?);
        }
        for (_, privilege) in &events {
            self.require(&object, privilege, relation.location)?;
        }
        let rules = self.row_rules(&table)?;
        let triggers = self.table_triggers(&table)?;
        let has = |event: Event| events.iter().any(|(e, _)| *e == event);
        for event in [Event::Insert, Event::Update, Event::Delete] {
            if has(event) {
                triggers.statement(self, event, BEFORE, &[])?;
            }
        }
        let targets = scan(self.db, &table)?;
        let sources = source_plan.run(self)?;
        let mut actions: Vec<MergeChange> = Vec::new();
        let mut touched: Vec<Vec<u8>> = Vec::new();
        for source in &sources {
            let mut matched_any = false;
            for target in &targets {
                let mut combined = target.clone();
                combined.extend(source.iter().cloned());
                if !join.is_true(self, &combined)? {
                    continue;
                }
                matched_any = true;
                let Some(clause) = first_clause(self, &clauses, true, &combined)? else { continue };
                if matches!(clause.action, MergeAction::Nothing) {
                    continue;
                }
                let key = table.encode_row(self.db, target)?.0;
                if touched.iter().any(|k| table.compare_keys(k, &key) == Ordering::Equal) {
                    return Err(PgError {
                        hint: Some("Ensure that not more than one source row matches any one target row.".into()),
                        ..PgError::new(code::CARDINALITY_VIOLATION, "MERGE command cannot affect row a second time")
                    });
                }
                touched.push(key);
                match &clause.action {
                    MergeAction::Update(assignments) => {
                        let mut new_row = target.clone();
                        for (i, expr) in assignments {
                            new_row[*i] = match expr {
                                Expr::Default(c) => default_value(self, &rules, *c)?,
                                expr => expr.eval(self, &combined)?,
                            };
                        }
                        let updated: Vec<usize> = assignments.iter().map(|(i, _)| *i).collect();
                        let before = triggers.before_row(self, Event::Update, Some(target), Some(new_row), &updated)?;
                        if let Some(mut new_row) = before {
                            check_row(self, &table, &rules, &mut new_row)?;
                            actions.push((Event::Update, Some(target.clone()), Some(new_row), updated));
                        }
                    }
                    MergeAction::Delete => {
                        if let Some(row) = triggers.before_row(self, Event::Delete, Some(target), None, &[])? {
                            actions.push((Event::Delete, Some(row), None, Vec::new()));
                        }
                    }
                    _ => {}
                }
            }
            if matched_any {
                continue;
            }
            let combined = source;
            let Some(clause) = first_clause(self, &clauses, false, combined)? else { continue };
            let MergeAction::Insert(values) = &clause.action else { continue };
            let mut row = Vec::with_capacity(table.columns.len());
            for (i, value) in values.iter().enumerate() {
                row.push(match value {
                    Some(Expr::Default(c)) => default_value(self, &rules, *c)?,
                    None => default_value(self, &rules, i)?,
                    Some(expr) => expr.eval(self, combined)?,
                });
            }
            if let Some(mut row) = triggers.before_row(self, Event::Insert, None, Some(row), &[])? {
                check_row(self, &table, &rules, &mut row)?;
                actions.push((Event::Insert, None, Some(row), Vec::new()));
            }
        }
        let mut edits = Edits::deferring(self, &table)?;
        for (_, old, new, _) in &actions {
            if let Some(old) = old {
                edits.delete(self, old)?;
            }
            if let Some(new) = new {
                edits.insert(self, new)?;
            }
        }
        edits.owe(self);
        edits.apply(self.db, self.txn)?;
        let changes: Vec<Change> = actions.iter().map(|(_, old, new, _)| (old.clone(), new.clone())).collect();
        self.enforce_foreign_keys(&table, &changes)?;
        for (event, old, new, updated) in &actions {
            triggers.after_row(self, *event, old.as_deref(), new.as_deref(), updated)?;
        }
        for event in [Event::Delete, Event::Update, Event::Insert] {
            if has(event) {
                triggers.statement(self, event, AFTER, &[])?;
            }
        }
        Ok(Outcome::command(format!("MERGE {}", actions.len())))
    }

    /// merge_clause binds a WHEN clause of MERGE, noting the change it makes in `events` with the privilege it needs.
    fn merge_clause(
        &mut self,
        clause: &pg_query::Node,
        table: &TableDef,
        scope: &Scope,
        source_scope: &Scope,
        events: &mut Vec<(Event, &'static str)>,
    ) -> Result<MergeClause> {
        use pg_query::protobuf::{CmdType, MergeMatchKind};
        let Some(NodeEnum::MergeWhenClause(clause)) = clause.node.as_ref() else {
            return Err(PgError::internal("a MERGE clause that is not one"));
        };
        let matched = clause.match_kind == MergeMatchKind::MergeWhenMatched as i32;
        let mut binder = Binder::new(self, if matched { scope.clone() } else { source_scope.clone() });
        binder.clause = "WHEN";
        let condition = match clause.condition.as_deref() {
            Some(node) => Some(crate::expr::condition(binder.bind(node)?, "WHEN", arg_location(node))?),
            None => None,
        };
        let action = match CmdType::try_from(clause.command_type) {
            Ok(CmdType::CmdUpdate) => {
                binder.clause = "UPDATE";
                events.push((Event::Update, "w"));
                MergeAction::Update(bind_assignments(&mut binder, table, &clause.target_list)?)
            }
            Ok(CmdType::CmdDelete) => {
                events.push((Event::Delete, "d"));
                MergeAction::Delete
            }
            Ok(CmdType::CmdInsert) => {
                binder.clause = "VALUES";
                events.push((Event::Insert, "a"));
                let targets: Vec<usize> = match clause.target_list.is_empty() {
                    true => (0..table.columns.len()).collect(),
                    false => clause
                        .target_list
                        .iter()
                        .filter_map(|t| match t.node.as_ref() {
                            Some(NodeEnum::ResTarget(t)) => Some(t),
                            _ => None,
                        })
                        .map(|t| {
                            table.columns.iter().position(|c| c.name == t.name).ok_or_else(|| PgError {
                                position: position(t.location),
                                ..PgError::new(
                                    code::UNDEFINED_COLUMN,
                                    format!("column \"{}\" of relation \"{}\" does not exist", t.name, table.name),
                                )
                            })
                        })
                        .collect::<Result<_>>()?,
                };
                if clause.values.len() > targets.len() {
                    return Err(PgError {
                        position: position(arg_location(&clause.values[targets.len()])),
                        ..PgError::new(code::SYNTAX_ERROR, "INSERT has more expressions than target columns")
                    });
                }
                let mut values: Vec<Option<Expr>> = vec![None; table.columns.len()];
                for (item, &target) in clause.values.iter().zip(&targets) {
                    let column = &table.columns[target];
                    values[target] = Some(match item.node.as_ref() {
                        Some(NodeEnum::SetToDefault(_)) => Expr::Default(target),
                        _ => {
                            let bound = binder.bind(item)?;
                            let bound = binder.reg_literal(bound, column.ty, arg_location(item))?;
                            assign(bound, column.ty, &column.name, arg_location(item))?.0
                        }
                    });
                }
                MergeAction::Insert(values)
            }
            _ => MergeAction::Nothing,
        };
        Ok(MergeClause { matched, condition, action })
    }
}

/// hidden_target explains an error of a WHEN NOT MATCHED clause of MERGE that refers to the target table, which such a
/// clause cannot see, as Postgres does.
fn hidden_target(err: PgError, table: &TableDef, name: &str) -> PgError {
    if err.code == code::UNDEFINED_TABLE && err.message == format!("missing FROM-clause entry for table \"{name}\"") {
        return PgError {
            message: format!("invalid reference to FROM-clause entry for table \"{name}\""),
            hint: Some(format!(
                "There is an entry for table \"{name}\", but it cannot be referenced from this part of the query."
            )),
            ..err
        };
    }
    let column = err.message.strip_prefix("column \"").and_then(|m| m.strip_suffix("\" does not exist"));
    match column {
        Some(column) if err.code == code::UNDEFINED_COLUMN && table.columns.iter().any(|c| c.name == column) => {
            PgError {
                hint: Some(format!(
                    "There is a column named \"{column}\" in table \"{name}\", but it cannot be referenced from this part \
                 of the query."
                )),
                ..err
            }
        }
        _ => err,
    }
}

/// first_clause returns the first WHEN clause of MERGE, for matched rows or for the others, whose condition holds.
fn first_clause<'c>(
    ctx: &mut Ctx<'_>,
    clauses: &'c [MergeClause],
    matched: bool,
    row: &[Value],
) -> Result<Option<&'c MergeClause>> {
    for clause in clauses.iter().filter(|c| c.matched == matched) {
        match &clause.condition {
            Some(condition) if !condition.is_true(ctx, row)? => continue,
            _ => return Ok(Some(clause)),
        }
    }
    Ok(None)
}

impl Ctx<'_> {
    /// update_catalog_statistics runs an UPDATE of pg_class that only sets planner statistics as one that changes
    /// nothing, reporting the rows it matches.
    pub fn update_catalog_statistics(&mut self, update: &UpdateStmt) -> Result<Option<Outcome>> {
        let Some(relation) = update.relation.as_ref().filter(|_| sets_catalog_statistics(update)) else {
            return Ok(None);
        };
        let one = pg_query::protobuf::AConst {
            val: Some(pg_query::protobuf::a_const::Val::Ival(pg_query::protobuf::Integer { ival: 1 })),
            ..Default::default()
        };
        let target = pg_query::protobuf::ResTarget {
            val: Some(Box::new(pg_query::Node { node: Some(NodeEnum::AConst(one)) })),
            ..Default::default()
        };
        let select = pg_query::protobuf::SelectStmt {
            target_list: vec![pg_query::Node { node: Some(NodeEnum::ResTarget(Box::new(target))) }],
            from_clause: vec![pg_query::Node { node: Some(NodeEnum::RangeVar(relation.clone())) }],
            where_clause: update.where_clause.clone(),
            ..Default::default()
        };
        let matched = match self.run(&NodeEnum::SelectStmt(Box::new(select)))? {
            Outcome::Rows { rows, .. } => rows.len(),
            _ => 0,
        };
        Ok(Some(Outcome::command(format!("UPDATE {matched}"))))
    }
}

/// sets_catalog_statistics reports whether an UPDATE only sets pg_class's planner statistics, which Doltgres does not
/// keep.
pub fn sets_catalog_statistics(update: &UpdateStmt) -> bool {
    let Some(relation) = &update.relation else { return false };
    relation.relname == "pg_class"
        && matches!(relation.schemaname.as_str(), "" | "pg_catalog")
        && update.target_list.iter().all(|target| {
            matches!(target.node.as_ref(), Some(NodeEnum::ResTarget(r)) if matches!(r.name.as_str(), "reltuples" | "relpages" | "relallvisible"))
        })
}
