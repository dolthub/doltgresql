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
use prolly::{Tuple, get};

use crate::cast::cast_value;
use crate::catalog::table::TableDef;
use crate::error::{ErrorObjects, PgError, Result, code};
use crate::expr::{Binder, Expr, Scope, ScopeColumn, arg_location, assign, coerce, position, typ};
use crate::plan::{Plan, Planner};
use crate::query::{Ctx, scan};
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
}

/// InsertPlan is a planned INSERT.
#[derive(Clone, Debug)]
pub struct InsertPlan {
    table: TableDef,
    rules: RowRules,
    /// The table column that each source value goes to.
    targets: Vec<usize>,
    source: InsertSource,
}

/// UpdatePlan is a planned UPDATE.
#[derive(Clone, Debug)]
pub struct UpdatePlan {
    table: TableDef,
    rules: RowRules,
    filter: Option<Expr>,
    /// The new value of each assigned column, over the old row.
    assignments: Vec<(usize, Expr)>,
}

/// DeletePlan is a planned DELETE.
#[derive(Clone, Debug)]
pub struct DeletePlan {
    table: TableDef,
    filter: Option<Expr>,
}

/// table_scope returns the scope of a table's columns under its name or alias.
fn table_scope(table: &TableDef, alias: Option<&str>) -> Scope {
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
fn check_row(ctx: &mut Ctx<'_>, table: &TableDef, rules: &RowRules, row: &[Value]) -> Result<()> {
    for (column, value) in table.columns.iter().zip(row) {
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

/// Edits collects changes to a table's primary index by key.
struct Edits<'a> {
    table: &'a TableDef,
    edits: Vec<(Vec<u8>, Option<Vec<u8>>)>,
}

impl<'a> Edits<'a> {
    /// new starts collecting changes to the table.
    fn new(table: &'a TableDef) -> Edits<'a> {
        Edits { table, edits: Vec::new() }
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
        self.edits.push((key, None));
        Ok(())
    }

    /// apply writes the edits to the table and the table to the transaction's working root.
    fn apply(self, db: &mut Database, txn: &mut Txn) -> Result<()> {
        if self.edits.is_empty() {
            return Ok(());
        }
        let table = self.table;
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
        let mut stored = table.table.clone();
        stored.edit_rows(db, merged, &|a, b| table.compare_keys(a, b))?;
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
        for column in &table.columns {
            if column.default.is_empty() {
                defaults.push(None);
                continue;
            }
            let node = parse_expression(&column.default)?;
            let bound = Binder::new(self, Scope::default()).bind(&node)?;
            defaults.push(Some(assign(bound, column.ty, &column.name, -1)?.0));
        }
        let mut checks = Vec::with_capacity(table.checks.len());
        for check in &table.checks {
            let node = parse_expression(&check.expression)?;
            let bound = Binder::new(self, table_scope(table, None)).bind(&node)?;
            checks.push((check.name.clone(), coerce(bound, typ(oid::BOOL), false, -1)?.0));
        }
        Ok(RowRules { defaults, checks })
    }

    /// plan_insert plans an INSERT.
    pub fn plan_insert(&mut self, insert: &InsertStmt) -> Result<InsertPlan> {
        if insert.on_conflict_clause.is_some() || !insert.returning_list.is_empty() || insert.with_clause.is_some() {
            return Err(PgError::unsupported("this INSERT"));
        }
        let relation = insert.relation.as_ref().ok_or_else(|| PgError::internal("INSERT without a table"))?;
        let table = self.resolve_table(relation)?;
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
        let Some(NodeEnum::SelectStmt(select)) = insert.select_stmt.as_deref().and_then(|n| n.node.as_ref()) else {
            return Err(PgError::unsupported("INSERT without VALUES or SELECT"));
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
        Ok(InsertPlan { table, rules, targets, source })
    }

    /// plan_update plans an UPDATE.
    pub fn plan_update(&mut self, update: &UpdateStmt) -> Result<UpdatePlan> {
        if !update.from_clause.is_empty() || !update.returning_list.is_empty() || update.with_clause.is_some() {
            return Err(PgError::unsupported("this UPDATE"));
        }
        let relation = update.relation.as_ref().ok_or_else(|| PgError::internal("UPDATE without a table"))?;
        let table = self.resolve_table(relation)?;
        let scope = table_scope(&table, relation.alias.as_ref().map(|a| a.aliasname.as_str()));
        let mut binder = Binder::new(self, scope);
        let filter = match update.where_clause.as_deref() {
            Some(node) => Some(coerce(binder.bind(node)?, typ(oid::BOOL), false, -1)?.0),
            None => None,
        };
        let mut assignments = Vec::new();
        for target in &update.target_list {
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
            let expr = if matches!(value.node.as_ref(), Some(NodeEnum::SetToDefault(_))) {
                Expr::Default(i)
            } else {
                let bound = binder.bind(value)?;
                assign(bound, column.ty, &column.name, arg_location(value))?.0
            };
            if !target.indirection.is_empty() {
                return Err(PgError::unsupported("this assignment"));
            }
            assignments.push((i, expr));
        }
        let rules = self.row_rules(&table)?;
        Ok(UpdatePlan { table, rules, filter, assignments })
    }

    /// plan_delete plans a DELETE.
    pub fn plan_delete(&mut self, delete: &DeleteStmt) -> Result<DeletePlan> {
        if !delete.using_clause.is_empty() || !delete.returning_list.is_empty() || delete.with_clause.is_some() {
            return Err(PgError::unsupported("this DELETE"));
        }
        let relation = delete.relation.as_ref().ok_or_else(|| PgError::internal("DELETE without a table"))?;
        let table = self.resolve_table(relation)?;
        let scope = table_scope(&table, relation.alias.as_ref().map(|a| a.aliasname.as_str()));
        let mut binder = Binder::new(self, scope);
        let filter = match delete.where_clause.as_deref() {
            Some(node) => Some(coerce(binder.bind(node)?, typ(oid::BOOL), false, -1)?.0),
            None => None,
        };
        Ok(DeletePlan { table, filter })
    }
}

impl InsertPlan {
    /// run inserts the rows.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Outcome> {
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
        insert_checked_rows(ctx, &self.table, &self.rules, rows)?;
        Ok(Outcome::command(format!("INSERT 0 {}", sources.len())))
    }
}

/// default_value evaluates a column's default, which is NULL without one.
fn default_value(ctx: &mut Ctx<'_>, rules: &RowRules, column: usize) -> Result<Value> {
    match &rules.defaults[column] {
        Some(expr) => expr.eval(ctx, &[]),
        None => Ok(Value::Null),
    }
}

/// insert_checked_rows checks and inserts rows in the table's column order.
fn insert_checked_rows(ctx: &mut Ctx<'_>, table: &TableDef, rules: &RowRules, rows: Vec<Vec<Value>>) -> Result<()> {
    for row in &rows {
        check_row(ctx, table, rules, row)?;
    }
    let (db, txn) = (&mut *ctx.db, &mut *ctx.txn);
    let mut edits = Edits::new(table);
    for row in &rows {
        edits.insert(db, row)?;
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
    insert_checked_rows(ctx, table, &rules, rows)
}

impl UpdatePlan {
    /// run updates the matching rows.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Outcome> {
        let mut changes = Vec::new();
        for row in scan(ctx.db, &self.table)? {
            if let Some(filter) = &self.filter
                && !filter.is_true(ctx, &row)?
            {
                continue;
            }
            let mut new_row = row.clone();
            for (i, expr) in &self.assignments {
                new_row[*i] = match expr {
                    Expr::Default(c) => default_value(ctx, &self.rules, *c)?,
                    expr => expr.eval(ctx, &row)?,
                };
            }
            check_row(ctx, &self.table, &self.rules, &new_row)?;
            changes.push((row, new_row));
        }
        let (db, txn) = (&mut *ctx.db, &mut *ctx.txn);
        let mut edits = Edits::new(&self.table);
        let mut new_rows = Vec::new();
        for (row, new_row) in changes {
            edits.delete(db, &row)?;
            new_rows.push(new_row);
        }
        let count = new_rows.len();
        for row in &new_rows {
            edits.insert(db, row)?;
        }
        edits.apply(db, txn)?;
        Ok(Outcome::command(format!("UPDATE {count}")))
    }
}

impl DeletePlan {
    /// run deletes the matching rows.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Outcome> {
        let mut doomed = Vec::new();
        for row in scan(ctx.db, &self.table)? {
            if let Some(filter) = &self.filter
                && !filter.is_true(ctx, &row)?
            {
                continue;
            }
            doomed.push(row);
        }
        let (db, txn) = (&mut *ctx.db, &mut *ctx.txn);
        let mut edits = Edits::new(&self.table);
        let count = doomed.len();
        for row in &doomed {
            edits.delete(db, row)?;
        }
        edits.apply(db, txn)?;
        Ok(Outcome::command(format!("DELETE {count}")))
    }
}
