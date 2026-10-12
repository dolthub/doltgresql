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

//! Changes to Dolt's workspace tables: an UPDATE of `staged` stages or unstages the rows it selects, and a DELETE
//! discards the working changes of the rows it selects, as Dolt's WorkspaceTableUpdater and WorkspaceTableDeleter do.

use pg_query::NodeEnum;
use pg_query::protobuf::{DeleteStmt, RangeVar, UpdateStmt};

use crate::Outcome;
use crate::catalog::table::TableDef;
use crate::dolt::args::error;
use crate::dolt::conflicts::star;
use crate::dolt::diff::{Kind, UserTable, lookup};
use crate::error::{PgError, Result, code};
use crate::expr::typ;
use crate::oid::BOOL;
use crate::query::Ctx;
use crate::types::Value;

/// STAGED is the position of a workspace row's `staged` column.
const STAGED: usize = 1;

/// FIRST_USER_COLUMN is the position of a workspace row's first `to_` column, after `id`, `staged`, and `diff_type`.
const FIRST_USER_COLUMN: usize = 3;

/// all_null reports whether every value of a row is NULL, which marks the missing side of an added or removed row.
fn all_null(row: &[Value]) -> bool {
    row.iter().all(Value::is_null)
}

impl Ctx<'_> {
    /// workspace_table returns the workspace table that a DML statement's target names, unless a user table has that
    /// name.
    fn workspace_table(&mut self, relation: Option<&RangeVar>) -> Result<Option<UserTable>> {
        let Some(relation) = relation else { return Ok(None) };
        if self.resolve_table(relation).is_ok() {
            return Ok(None);
        }
        Ok(lookup(self, &relation.schemaname, &relation.relname)?.filter(|t| t.kind == Kind::Workspace))
    }

    /// is_workspace_table reports whether a DML statement's target is a workspace table rather than a user table.
    pub fn is_workspace_table(&mut self, relation: Option<&RangeVar>) -> Result<bool> {
        Ok(self.workspace_table(relation)?.is_some())
    }

    /// modifiable_workspace returns the user table of a workspace table, failing as Dolt does when the statement
    /// changes rows and the table's schema differs between HEAD and the working root.
    fn modifiable_workspace(
        &mut self,
        relation: &RangeVar,
        table: &UserTable,
        rows: usize,
    ) -> Result<Option<TableDef>> {
        if rows > 0 && table.to != table.from {
            return Err(error(format!("{} table is not modifiable due to schema change", relation.relname)));
        }
        self.txn.table(self.db, &table.schema, &table.name)
    }

    /// update_workspace runs an UPDATE of a workspace table, which may only set `staged`: rows set to staged have
    /// their working values written to the staged root, and rows set to unstaged have their HEAD values written there.
    /// It returns None when the UPDATE is of another table.
    pub fn update_workspace(&mut self, update: &UpdateStmt) -> Result<Option<Outcome>> {
        let Some(table) = self.workspace_table(update.relation.as_ref())? else { return Ok(None) };
        let Some(relation) = update.relation.as_ref() else { return Ok(None) };
        self.check_branch_access(crate::dolt::branch_control::WRITE)?;
        let mut columns = vec!["id".to_string(), "staged".to_string(), "diff_type".to_string()];
        columns.extend(table.to.iter().map(|c| format!("to_{}", c.name)));
        columns.extend(table.from.iter().map(|c| format!("from_{}", c.name)));
        let mut assigned = Vec::new();
        let mut targets = vec![star()];
        for target in &update.target_list {
            let Some(NodeEnum::ResTarget(target)) = target.node.as_ref() else { continue };
            let Some(position) = columns.iter().position(|c| *c == target.name) else {
                return Err(PgError::new(
                    code::UNDEFINED_COLUMN,
                    format!("column \"{}\" of relation \"{}\" does not exist", target.name, relation.relname),
                ));
            };
            assigned.push(position);
            targets.push(pg_query::Node { node: Some(NodeEnum::ResTarget(target.clone())) });
        }
        let select = pg_query::protobuf::SelectStmt {
            target_list: targets,
            from_clause: vec![pg_query::Node { node: Some(NodeEnum::RangeVar(relation.clone())) }],
            where_clause: update.where_clause.clone(),
            ..Default::default()
        };
        let query = crate::plan::Planner { ctx: self, outer: Vec::new() }.plan_query(&select)?;
        for (&position, ty) in assigned.iter().zip(&query.types[columns.len()..]) {
            if position == STAGED && !crate::expr::assignable(ty.oid, BOOL) {
                return Err(PgError {
                    hint: Some("You will need to rewrite or cast the expression.".into()),
                    ..PgError::new(
                        code::DATATYPE_MISMATCH,
                        format!(
                            "column \"staged\" is of type boolean but expression is of type {}",
                            crate::cast::type_display(ty.oid)
                        ),
                    )
                });
            }
        }
        let rows = query.plan.run(self)?;
        let Some(user) = self.modifiable_workspace(relation, &table, rows.len())? else {
            return Ok(Some(Outcome::command(format!("UPDATE {}", rows.len()))));
        };
        let width = user.columns.len();
        let (mut deletes, mut puts) = (Vec::new(), Vec::new());
        for row in &rows {
            let (old, values) = row.split_at(columns.len());
            let mut new = old.to_vec();
            for (&position, value) in assigned.iter().zip(values) {
                new[position] = match position {
                    STAGED => crate::cast::cast_value(value.clone(), typ(BOOL), false)?,
                    _ => value.clone(),
                };
            }
            if old.iter().zip(&new).enumerate().any(|(i, (a, b))| i != STAGED && a != b) {
                return Err(error("only update of column 'staged' is allowed"));
            }
            let (mut to, mut from) = (&old[FIRST_USER_COLUMN..][..width], &old[FIRST_USER_COLUMN + width..]);
            if new[STAGED] != Value::Bool(true) {
                std::mem::swap(&mut to, &mut from);
            }
            if !all_null(from) {
                deletes.push(from.to_vec());
            }
            if !all_null(to) {
                puts.push(to.to_vec());
            }
        }
        self.write_staged(&user, &deletes, &puts)?;
        Ok(Some(Outcome::command(format!("UPDATE {}", rows.len()))))
    }

    /// write_staged deletes rows from a table of the staged root and then inserts others, first copying the table's
    /// schema to the staged root without its rows when only the working root has it.
    fn write_staged(&mut self, user: &TableDef, deletes: &[Vec<Value>], puts: &[Vec<Value>]) -> Result<()> {
        std::mem::swap(&mut self.txn.root, &mut self.txn.staged);
        let written = self.write_rows(user, deletes, puts);
        std::mem::swap(&mut self.txn.root, &mut self.txn.staged);
        written
    }

    /// write_rows deletes rows from a table of the working root and then inserts others, creating the table without
    /// rows first when the working root lacks it.
    fn write_rows(&mut self, user: &TableDef, deletes: &[Vec<Value>], puts: &[Vec<Value>]) -> Result<()> {
        if self.txn.table(self.db, &user.schema, &user.name)?.is_none() {
            let mut stored = user.table.clone();
            stored.primary_index = doltdb::table::empty_rows();
            for index in &user.indexes {
                let empty = index.empty_root(self.db)?;
                stored.put_index(self.db, &index.name, Some(empty))?;
            }
            let address = stored.write(self.db)?;
            self.txn.root.put_table(self.db, &user.schema, &user.name, Some(address))?;
        }
        let current = |ctx: &mut Ctx<'_>| {
            ctx.txn.table(ctx.db, &user.schema, &user.name)?.ok_or_else(|| error("table not found"))
        };
        let table = current(self)?;
        crate::dml::delete_rows(self, &table, deletes)?;
        let table = current(self)?;
        crate::dml::write_rows(self, &table, puts)
    }

    /// delete_workspace runs a DELETE from a workspace table, which returns the working rows it selects to their staged
    /// values through the user table, running its foreign key actions, or returns None when the DELETE is of another
    /// table.
    pub fn delete_workspace(&mut self, delete: &DeleteStmt) -> Result<Option<Outcome>> {
        let Some(table) = self.workspace_table(delete.relation.as_ref())? else { return Ok(None) };
        let Some(relation) = delete.relation.as_ref() else { return Ok(None) };
        self.check_branch_access(crate::dolt::branch_control::WRITE)?;
        let rows = self.select_rows(relation, vec![star()], delete.where_clause.clone())?;
        let Some(user) = self.modifiable_workspace(relation, &table, rows.len())? else {
            return Ok(Some(Outcome::command(format!("DELETE {}", rows.len()))));
        };
        let width = user.columns.len();
        let mut changes = Vec::new();
        for row in &rows {
            if row[STAGED] == Value::Bool(true) {
                return Err(error("cannot delete staged rows from workspace"));
            }
            let (to, from) = (&row[FIRST_USER_COLUMN..][..width], &row[FIRST_USER_COLUMN + width..]);
            let side = |values: &[Value]| (!all_null(values)).then(|| values.to_vec());
            changes.push((side(to), side(from)));
        }
        crate::dml::apply_changes(self, &user, &changes)?;
        self.enforce_foreign_keys(&user, &changes)?;
        Ok(Some(Outcome::command(format!("DELETE {}", rows.len()))))
    }
}
