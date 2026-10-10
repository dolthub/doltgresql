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

//! Queries: planning a SELECT into scans, filters, sorts, and projections, and running the plan.

use doltdb::database::Database;
use pg_query::Node;
use pg_query::protobuf::RangeVar;

use crate::catalog::builtin_type;
use crate::catalog::table::TableDef;
use crate::engine::SessionState;
use crate::error::{PgError, Result, code};
use crate::expr::{Binder, Scope, position};
use crate::txn::Txn;
use crate::types::Value;
use crate::{Column, oid};

/// Ctx is what planning and running a statement reads and changes: the database, the transaction's working root, the
/// session, and the statement's parameters.
pub struct Ctx<'a> {
    pub db: &'a mut Database,
    pub txn: &'a mut Txn,
    /// The session's open transactions on the database's other branches, which statements that change another
    /// branch's tables join.
    pub branches: &'a mut Vec<Txn>,
    pub session: &'a mut SessionState,
    /// The types of the statement's parameters, where 0 is a type not yet known.
    pub parameters: &'a mut Vec<u32>,
    /// The values of the statement's parameters, empty while only planning.
    pub params: &'a [Value],
    /// The rows of the enclosing queries of a running subquery, innermost last.
    pub outer: Vec<Vec<Value>>,
    /// The subquery value that an ANY or ALL comparison is testing.
    pub subquery_value: Value,
    /// The WITH queries in scope while planning, innermost last.
    pub ctes: Vec<crate::plan::Cte>,
    /// The rows of each recursive WITH query's working table while it runs, by its ID.
    pub work_tables: std::collections::HashMap<usize, Vec<Vec<Value>>>,
    /// The name of the SQL function whose body is running and the names of its parameters, which the body can refer
    /// to its parameters by.
    pub named_params: Option<(String, Vec<String>)>,
    /// The rows of the `Once` plans that the running plan has evaluated, by their address, or None outside a plan run.
    pub once: Option<std::collections::HashMap<usize, std::sync::Arc<crate::plan::SubqueryRows>>>,
    /// The outermost enclosing scope that the expressions bound so far refer to, as an index into the binder's
    /// scopes, or `usize::MAX` for none.
    pub outer_reach: usize,
    /// The scopes, as indexes into the binder's scopes, of the aggregate calls whose arguments are being bound.
    pub aggregate_levels: Vec<usize>,
    /// What the statement has read of the catalogs, for the root value it read them from.
    pub catalog: Option<crate::pgcatalog::snapshot::CatalogCache>,
    /// The PL/pgSQL variables that the statement's parameters stand for, by position, which a column of the same name
    /// makes ambiguous.
    pub variables: Vec<String>,
    /// The views being expanded while planning, outermost first, by schema and name.
    pub expanding: Vec<(String, String)>,
    /// The locations of the subqueries of the WHERE clause being bound that the ported optimizer may turn into joins,
    /// which their binding leaves unplanned when it can.
    pub deferred_sublinks: Vec<i32>,
    /// Whether the subqueries of the SELECT being bound become SubPlans that the ported optimizer plans when it plans
    /// the query around them, rather than as they are bound.
    pub defer_subplans: bool,
    /// The error that the ported optimizer found while planning, which it raises once planning ends.
    pub planner_error: Option<PgError>,
}

/// column returns the description of a result column of the type.
pub fn column(name: String, ty: crate::catalog::ColumnType) -> Column {
    let ty = crate::usertypes::base_type(ty);
    let type_size = match builtin_type(ty.oid) {
        Some(t) => t.definition.typ_length,
        None => crate::usertypes::get(ty.oid).map_or(-1, |t| t.definition.typ_length),
    };
    let type_oid = if ty.oid == oid::UNKNOWN { oid::TEXT } else { ty.oid };
    Column {
        name,
        type_oid,
        type_size: if ty.oid == oid::UNKNOWN { -1 } else { type_size },
        type_modifier: ty.modifier,
        origin: (0, 0),
    }
}

impl Ctx<'_> {
    /// resolve_table loads the table that a range variable names.
    pub fn resolve_table(&mut self, relation: &RangeVar) -> Result<TableDef> {
        if let Some(table) = self.nonlocal_table(relation)? {
            return Ok(table);
        }
        let schemas: Vec<String> = if relation.schemaname.is_empty() {
            self.session.search_path()
        } else {
            vec![self.session.named_schema(&relation.schemaname)]
        };
        for schema in &schemas {
            if let Some(table) = self.txn.table(self.db, schema, &relation.relname)? {
                self.session.temp_used |= *schema == self.session.temp_schema();
                return Ok(table);
            }
        }
        if relation.schemaname.is_empty() && relation.relname.starts_with("dolt_") {
            let name = if relation.relname == "dolt_rebase" { "rebase" } else { &relation.relname };
            if let Some(table) = self.txn.table(self.db, "dolt", name)? {
                return Ok(table);
            }
        }
        Err(undefined_table(relation))
    }

    /// resolve_table_as_of loads the table that a range variable names as it was at a revision: a commit spec,
    /// `WORKING`, `STAGED`, or a time, which names the newest commit of the branch made by then, as Dolt's
    /// resolveAsOf does.
    pub fn resolve_table_as_of(&mut self, relation: &RangeVar, revision: &Node) -> Result<TableDef> {
        let revision = self.constant_text(revision)?;
        let Some(root) = self.revision_root(&revision)? else {
            return Err(undefined_table(relation));
        };
        self.resolve_table_in(relation, &root)
    }

    /// resolve_table_in loads the table that a range variable names from a root value.
    pub fn resolve_table_in(&mut self, relation: &RangeVar, root: &doltdb::root::Root) -> Result<TableDef> {
        let schemas: Vec<String> = if relation.schemaname.is_empty() {
            self.session.search_path()
        } else {
            vec![self.session.named_schema(&relation.schemaname)]
        };
        for schema in &schemas {
            if let Some(address) = root.table(self.db, schema, &relation.relname)? {
                return TableDef::load(self.db, schema, &relation.relname, address);
            }
        }
        Err(undefined_table(relation))
    }

    /// catalog_root returns the working root of the branch that a range variable's database names, either as
    /// `database/branch` or as a database whose checked-out branch the session is not on, or None for the session's
    /// own branch.
    pub fn catalog_root(&mut self, relation: &RangeVar) -> Result<Option<doltdb::root::Root>> {
        let (database, branch) = match relation.catalogname.split_once('/') {
            Some((database, branch)) => (database, branch.to_string()),
            None if relation.catalogname.is_empty() || !self.session.display.contains('/') => return Ok(None),
            None => (relation.catalogname.as_str(), self.session.checked_out_branch(&relation.catalogname)),
        };
        if database != self.session.database || branch == self.txn.branch {
            return Ok(None);
        }
        match self.branch_root(&branch)? {
            Some(root) => Ok(Some(root)),
            None => Err(PgError::new(
                code::INVALID_CATALOG_NAME,
                format!("database \"{}\" does not exist", relation.catalogname),
            )),
        }
    }

    /// branch_root returns the working root of a branch of the session's database as the session's transaction on it
    /// holds it, or as its working set last stored it, or None for a missing branch.
    pub fn branch_root(&mut self, branch: &str) -> Result<Option<doltdb::root::Root>> {
        if let Some(txn) = self.branches.iter().find(|t| t.database == self.txn.database && t.branch == branch) {
            return Ok(Some(txn.root.clone()));
        }
        let address = match self.db.head(&doltdb::create::working_set_ref(branch))? {
            Some(address) => {
                let data = crate::txn::read(self.db, &address)?;
                serial::WorkingSet::new(serial::Message(&data))?.working_root()?
            }
            None => match self.db.head(&doltdb::create::branch_ref(branch))? {
                Some(head) => crate::dolt::history::load(self.db, head)?.root,
                None => return Ok(None),
            },
        };
        Ok(Some(doltdb::root::Root::decode(&crate::txn::read(self.db, &address)?)?))
    }

    /// on_branch runs a function with the session's transaction on another branch of the database in place of its own,
    /// beginning that transaction when the session has none, so that the function's changes go to that branch.
    pub fn on_branch<T>(&mut self, branch: &str, f: impl FnOnce(&mut Ctx<'_>) -> Result<T>) -> Result<T> {
        let index = match self.branches.iter().position(|t| t.database == self.txn.database && t.branch == branch) {
            Some(index) => index,
            None => {
                let (handle, sequences) = (self.txn.handle.clone(), self.txn.sequences.clone());
                let mut txn = Txn::begin_locked(self.db, handle, sequences, &self.txn.database, branch)?;
                txn.started = self.txn.started;
                self.branches.push(txn);
                self.branches.len() - 1
            }
        };
        std::mem::swap(self.txn, &mut self.branches[index]);
        let result = f(self);
        std::mem::swap(self.txn, &mut self.branches[index]);
        result
    }

    /// revision_root returns the root value at a revision, or None for a time before the branch's first commit.
    pub(crate) fn revision_root(&mut self, revision: &str) -> Result<Option<doltdb::root::Root>> {
        let address = match revision.to_ascii_uppercase().as_str() {
            "WORKING" => return Ok(Some(self.txn.root.clone())),
            "STAGED" => return Ok(Some(self.txn.staged.clone())),
            _ => match crate::dolt::history::resolve(self.db, self.txn.head, revision) {
                Ok(commit) => crate::dolt::history::load(self.db, commit)?.root,
                Err(err) => {
                    let millis = crate::dolt::procedures::parse_date(revision).map_err(|_| err)?;
                    let log = crate::dolt::history::log(self.db, &[self.txn.head])?;
                    match log.into_iter().find(|c| c.committer_millis as i64 <= millis) {
                        Some(commit) => commit.root,
                        None => return Ok(None),
                    }
                }
            },
        };
        Ok(Some(doltdb::root::Root::decode(&crate::txn::read(self.db, &address)?)?))
    }

    /// shown_relation names a relation as Postgres' messages do: qualified by its schema unless that schema is on the
    /// search path.
    pub fn shown_relation(&self, schema: &str, name: &str) -> String {
        match self.session.search_path().iter().any(|s| s == schema) {
            true => name.to_string(),
            false => format!("{schema}.{name}"),
        }
    }

    /// creation_schema returns the schema that an unqualified new object goes in: the first schema of the search
    /// path that exists.
    pub fn creation_schema(&self) -> Result<String> {
        self.session
            .search_path()
            .into_iter()
            .find(|s| *s != self.session.temp_schema() && self.txn.root.schemas.iter().any(|e| e == s.as_bytes()))
            .ok_or_else(|| PgError::new(code::INVALID_SCHEMA_NAME, "no schema has been selected to create in"))
    }

    /// without_temp runs a function with the session's temporary tables out of the root, putting them back after.
    pub fn without_temp<T>(&mut self, f: impl FnOnce(&mut Ctx<'_>) -> Result<T>) -> Result<T> {
        if self.txn.temp_schema.is_none() {
            return f(self);
        }
        let objects = self.txn.take_temp(self.db)?.unwrap_or_default();
        let result = f(self);
        self.txn.inject_temp(self.db, &self.session.temp_schema(), &objects)?;
        result
    }

    /// constant_text evaluates an expression without columns and returns its text.
    pub fn constant_text(&mut self, node: &Node) -> Result<String> {
        let (expr, _) = Binder::new(self, Scope::default()).bind(node)?;
        Ok(expr.eval(self, &[])?.output().unwrap_or_default())
    }
}

/// undefined_table returns the error for a range variable that names no table.
fn undefined_table(relation: &RangeVar) -> PgError {
    let name = if relation.schemaname.is_empty() {
        relation.relname.clone()
    } else {
        format!("{}.{}", relation.schemaname, relation.relname)
    };
    PgError {
        position: position(relation.location),
        ..PgError::new(code::UNDEFINED_TABLE, format!("relation \"{name}\" does not exist"))
    }
}

/// scan returns every row of a table in key order, repeating each keyless row by its cardinality.
pub fn scan(db: &mut Database, table: &TableDef) -> Result<Vec<Vec<Value>>> {
    let mut walk = crate::exec::TableWalk::new(db, table, None)?;
    let mut rows = Vec::new();
    while let Some(row) = walk.next(db)? {
        rows.push(row);
    }
    Ok(rows)
}
