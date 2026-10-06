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

use std::cmp::Ordering;
use std::sync::Arc;

use doltdb::database::Database;
use pg_query::NodeEnum;
use pg_query::protobuf::{DeleteStmt, InsertStmt, UpdateStmt};
use prolly::{NodeStore, Tuple, get};

use crate::auth::Object;
use crate::cast::cast_value;
use crate::catalog::ColumnType;
use crate::catalog::table::{IndexDef, TableDef};
use crate::error::{ErrorObjects, PgError, Result, code};
use crate::expr::{Binder, Expr, Scope, ScopeColumn, arg_location, assign, coerce, position, typ};
use crate::foreign::Change;
use crate::plan::{Plan, Planner};
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
    Scope {
        columns: table
            .columns
            .iter()
            .map(|c| ScopeColumn { table: name.to_string(), name: c.name.clone(), ty: c.ty, hidden: false })
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
    let result = pg_query::parse(&format!("SELECT {text}")).map_err(PgError::internal)?;
    let statement = result.protobuf.stmts.into_iter().next().and_then(|s| s.stmt).and_then(|s| s.node);
    let Some(NodeEnum::SelectStmt(select)) = statement else {
        return Err(PgError::internal(format!("a stored expression that is not one: {text}")));
    };
    match select.target_list.into_iter().next().and_then(|t| t.node) {
        Some(NodeEnum::ResTarget(target)) => {
            target.val.map(|v| *v).ok_or_else(|| PgError::internal("an empty expression"))
        }
        _ => Err(PgError::internal(format!("a stored expression that is not one: {text}"))),
    }
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
            constraint: Some(format!("{}_pkey", table.name)),
            ..ErrorObjects::default()
        })),
        ..PgError::new(
            code::UNIQUE_VIOLATION,
            format!("duplicate key value violates unique constraint \"{}_pkey\"", table.name),
        )
    }
}

/// unique_violation returns Postgres' error for a row that duplicates another's values in a unique index.
fn unique_violation(table: &TableDef, index: &IndexDef, row: &[Value]) -> PgError {
    let names: Vec<&str> = index.columns.iter().map(|&i| table.columns[i].name.as_str()).collect();
    let values: Vec<Value> = index.columns.iter().map(|&i| row[i].clone()).collect();
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

/// Edits collects changes to a table's primary index and secondary indexes by key.
struct Edits<'a> {
    table: &'a TableDef,
    edits: Vec<(Vec<u8>, Option<Vec<u8>>)>,
    /// Each secondary index's changes, in the order of the table's indexes.
    index_edits: Vec<KeyEdits>,
}

impl<'a> Edits<'a> {
    /// new starts collecting changes to the table.
    fn new(table: &'a TableDef) -> Edits<'a> {
        Edits { table, edits: Vec::new(), index_edits: vec![Vec::new(); table.indexes.len()] }
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
        let mut pending: Vec<(&[u8], bool)> = Vec::new();
        for (k, v) in &self.index_edits[i] {
            match pending.iter_mut().find(|(p, _)| self.table.compare_index_keys(index, p, k) == Ordering::Equal) {
                Some(entry) => entry.1 = v.is_some(),
                None => pending.push((k, v.is_some())),
            }
        }
        if let Some((k, _)) = pending.iter().find(|(k, present)| *present && same(k)) {
            return Ok(Some(k.to_vec()));
        }
        let mut found = None;
        let root = db.read(&index.root)?;
        prolly::scan_from(db, root, key, &|a, b| self.table.compare_index_prefix(index, width, a, b), &mut |k, _| {
            if !same(k) {
                return Ok(false);
            }
            let deleted = pending
                .iter()
                .any(|(p, present)| !present && self.table.compare_index_keys(index, p, k) == Ordering::Equal);
            if !deleted {
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
    fn conflicting_row(&self, db: &mut Database, row: &[Value], target: &ConflictTarget) -> Result<Option<Vec<Value>>> {
        if !self.table.keyless() && matches!(target, ConflictTarget::Any | ConflictTarget::Primary) {
            let (key, _) = self.table.encode_row(db, row)?;
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
            if !wanted || index.columns.iter().any(|&c| row[c].is_null()) {
                continue;
            }
            let (primary, _) = self.table.encode_row(db, row)?;
            let key = self.table.index_key(db, index, row, &primary)?;
            if let Some(found) = self.index_match(db, i, &key)? {
                let primary = self.primary_of(index, &found)?;
                if let Some(value) = self.current(db, &primary)? {
                    return Ok(Some(self.table.decode_row(db, &primary, &value)?.0));
                }
            }
        }
        Ok(None)
    }

    /// index_row adds a row's keys to the secondary indexes, or removes them, checking unique indexes as it adds.
    fn index_row(&mut self, db: &mut Database, row: &[Value], primary: &[u8], add: bool) -> Result<()> {
        for i in 0..self.table.indexes.len() {
            let index = &self.table.indexes[i];
            let key = self.table.index_key(db, index, row, primary)?;
            if add
                && index.unique
                && index.columns.iter().all(|&c| !row[c].is_null())
                && self.index_taken(db, i, &key)?
            {
                return Err(unique_violation(self.table, index, row));
            }
            let value = add.then(|| prolly::val::build_tuple(&[]));
            self.index_edits[i].push((key, value));
        }
        Ok(())
    }

    /// pending returns the value an earlier edit gave a key: None when no edit touched it, and Some(None) when one
    /// deleted it.
    fn pending(&self, key: &[u8]) -> Option<Option<Vec<u8>>> {
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

    /// insert adds a row, failing on a duplicate primary key, and adding to the cardinality of a keyless row.
    fn insert(&mut self, db: &mut Database, row: &[Value]) -> Result<()> {
        let (key, mut value) = self.table.encode_row(db, row)?;
        if let Some(existing) = self.current(db, &key)? {
            if !self.table.keyless() {
                return Err(duplicate_key(self.table, row));
            }
            value = with_cardinality(&existing, cardinality(&existing) + 1);
        } else {
            self.index_row(db, row, &key, true)?;
        }
        self.edits.push((key, Some(value)));
        Ok(())
    }

    /// delete removes one copy of a row.
    fn delete(&mut self, db: &mut Database, row: &[Value]) -> Result<()> {
        let (key, _) = self.table.encode_row(db, row)?;
        if self.table.keyless()
            && let Some(existing) = self.current(db, &key)?
            && cardinality(&existing) > 1
        {
            let value = with_cardinality(&existing, cardinality(&existing) - 1);
            self.edits.push((key, Some(value)));
            return Ok(());
        }
        self.index_row(db, row, &key, false)?;
        self.edits.push((key, None));
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
        for (index, mut edits) in current.indexes.iter().zip(index_edits) {
            if edits.is_empty() {
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
            if column.default.is_empty() {
                defaults.push(None);
                continue;
            }
            let node = parse_expression(&column.default)?;
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

    /// plan_insert plans an INSERT.
    pub fn plan_insert(&mut self, insert: &InsertStmt) -> Result<InsertPlan> {
        if insert.with_clause.is_some() {
            return Err(PgError::unsupported("WITH in INSERT"));
        }
        let relation = insert.relation.as_ref().ok_or_else(|| PgError::internal("INSERT without a table"))?;
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
        let source = if select.values_lists.is_empty() || !select.sort_clause.is_empty() || select.limit_count.is_some()
        {
            let query = Planner { ctx: self, outer: Vec::new() }.plan_query(select)?;
            if query.columns.len() > targets.len() {
                return Err(too_many(-1));
            }
            for (ty, &target) in query.types.iter().zip(&targets) {
                let column = &table.columns[target];
                if column.generated {
                    return Err(generated_error(&column.name, "cannot insert a non-DEFAULT value into column"));
                }
                assign((Expr::Column(0), *ty), column.ty, &column.name, -1)?;
            }
            InsertSource::Select(Box::new(query.plan))
        } else {
            let mut binder = Binder::new(self, Scope::default());
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
                    let bound = binder.bind(item)?;
                    if let Expr::Param(i) = bound.0
                        && binder.ctx.parameters[i] == 0
                    {
                        binder.ctx.parameters[i] = column.ty.oid;
                    }
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
                if !table.keyless() && infer.conname == format!("{}_pkey", table.name) {
                    ConflictTarget::Primary
                } else if let Some(i) = table.indexes.iter().position(|ix| ix.unique && ix.name == infer.conname) {
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
                } else if let Some(i) = table.indexes.iter().position(|ix| ix.unique && same(&ix.columns)) {
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
        if update.with_clause.is_some() {
            return Err(PgError::unsupported("WITH in UPDATE"));
        }
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
            Some(node) => Some(coerce(binder.bind(node)?, typ(oid::BOOL), false, -1)?.0),
            None => None,
        };
        let assignments = bind_assignments(&mut binder, &table, &update.target_list)?;
        let rules = self.row_rules(&table)?;
        let returning = self.plan_returning(scope, &update.returning_list)?;
        Ok(UpdatePlan { table, rules, filter, assignments, from, returning })
    }

    /// plan_delete plans a DELETE, with its USING list joined to the table's rows.
    pub fn plan_delete(&mut self, delete: &DeleteStmt) -> Result<DeletePlan> {
        if delete.with_clause.is_some() {
            return Err(PgError::unsupported("WITH in DELETE"));
        }
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
            Some(node) => Some(coerce(binder.bind(node)?, typ(oid::BOOL), false, -1)?.0),
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
        if !target.indirection.is_empty() {
            return Err(PgError::unsupported("this assignment"));
        }
        if assignments.iter().any(|(c, _)| *c == i) {
            return Err(PgError::new(
                code::SYNTAX_ERROR,
                format!("multiple assignments to same column \"{}\"", target.name),
            ));
        }
        let value = target.val.as_deref().ok_or_else(|| PgError::internal("an assignment without a value"))?;
        let column = &table.columns[i];
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
        } else {
            let bound = binder.bind(value)?;
            if let Expr::Param(p) = bound.0
                && binder.ctx.parameters[p] == 0
            {
                binder.ctx.parameters[p] = column.ty.oid;
            }
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
        let mut edits = Edits::new(table);
        let mut written = Vec::new();
        let mut changes: Vec<Change> = Vec::new();
        let mut touched: Vec<Vec<u8>> = Vec::new();
        for row in rows {
            let Some(mut row) = triggers.before_row(ctx, Event::Insert, None, Some(row), &[])? else { continue };
            check_row(ctx, table, &self.rules, &mut row)?;
            let Some(existing) = edits.conflicting_row(ctx.db, &row, &on_conflict.target)? else {
                edits.insert(ctx.db, &row)?;
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
            edits.delete(ctx.db, &existing)?;
            edits.insert(ctx.db, &new_row)?;
            touched.push(table.encode_row(ctx.db, &new_row)?.0);
            changes.push((Some(existing), Some(new_row.clone())));
            written.push(new_row);
        }
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
    let (db, txn) = (&mut *ctx.db, &mut *ctx.txn);
    let mut edits = Edits::new(table);
    for row in &rows {
        edits.insert(db, row)?;
    }
    edits.apply(db, txn)?;
    Ok(rows)
}

/// write_rows inserts rows already in a table's column order and types, without checking them, as rebuilding a table
/// does.
pub fn write_rows(ctx: &mut Ctx<'_>, table: &TableDef, rows: &[Vec<Value>]) -> Result<()> {
    let (db, txn) = (&mut *ctx.db, &mut *ctx.txn);
    let mut edits = Edits::new(table);
    for row in rows {
        edits.insert(db, row)?;
    }
    edits.apply(db, txn)
}

/// delete_rows deletes rows read from a table.
pub fn delete_rows(ctx: &mut Ctx<'_>, table: &TableDef, rows: &[Vec<Value>]) -> Result<()> {
    let (db, txn) = (&mut *ctx.db, &mut *ctx.txn);
    let mut edits = Edits::new(table);
    for row in rows {
        edits.delete(db, row)?;
    }
    edits.apply(db, txn)
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
    let (db, txn) = (&mut *ctx.db, &mut *ctx.txn);
    let mut edits = Edits::new(table);
    for (old, _) in &changes {
        if let Some(row) = old {
            edits.delete(db, row)?;
        }
    }
    for (_, new) in &changes {
        if let Some(row) = new {
            edits.insert(db, row)?;
        }
    }
    edits.apply(db, txn)
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

/// matches pairs each of a table's rows with the first row of a FROM list that meets a filter, or with no row
/// without a FROM list, as UPDATE ... FROM and DELETE ... USING join them.
fn matches(
    ctx: &mut Ctx<'_>,
    table: &TableDef,
    from: &Option<Box<Plan>>,
    filter: &Option<Expr>,
) -> Result<Vec<(Vec<Value>, Vec<Value>)>> {
    let from_rows = match from {
        Some(plan) => Some(plan.run(ctx)?),
        None => None,
    };
    let mut out = Vec::new();
    for row in scan(ctx.db, table)? {
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
        let mut edits = Edits::new(&self.table);
        for (row, _, _) in &changes {
            edits.delete(ctx.db, row)?;
        }
        for (_, new_row, _) in &changes {
            edits.insert(ctx.db, new_row)?;
        }
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
        let mut edits = Edits::new(&self.table);
        for (row, _) in &doomed {
            edits.delete(ctx.db, row)?;
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
