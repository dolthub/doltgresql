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

use std::sync::Arc;

use doltdb::database::Database;
use pg_query::Node;
use pg_query::protobuf::RangeVar;
use prolly::walk_leaves;

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
        let schemas: Vec<String> =
            if relation.schemaname.is_empty() { self.session.search_path() } else { vec![relation.schemaname.clone()] };
        for schema in &schemas {
            if let Some(table) = self.txn.table(self.db, schema, &relation.relname)? {
                return Ok(table);
            }
        }
        let name = if relation.schemaname.is_empty() {
            relation.relname.clone()
        } else {
            format!("{}.{}", relation.schemaname, relation.relname)
        };
        Err(PgError {
            position: position(relation.location),
            ..PgError::new(code::UNDEFINED_TABLE, format!("relation \"{name}\" does not exist"))
        })
    }

    /// creation_schema returns the schema that an unqualified new object goes in: the first schema of the search
    /// path that exists.
    pub fn creation_schema(&self) -> Result<String> {
        self.session
            .search_path()
            .into_iter()
            .find(|s| self.txn.root.schemas.iter().any(|existing| existing == s.as_bytes()))
            .ok_or_else(|| PgError::new(code::INVALID_SCHEMA_NAME, "no schema has been selected to create in"))
    }

    /// constant_text evaluates an expression without columns and returns its text.
    pub fn constant_text(&mut self, node: &Node) -> Result<String> {
        let (expr, _) = Binder::new(self, Scope::default()).bind(node)?;
        Ok(expr.eval(self, &[])?.output().unwrap_or_default())
    }
}

/// scan returns every row of a table in key order, repeating each keyless row by its cardinality.
pub fn scan(db: &mut Database, table: &TableDef) -> Result<Vec<Vec<Value>>> {
    let node = Arc::new(prolly::Node::decode(table.table.primary_index.clone())?);
    let mut items = Vec::new();
    walk_leaves(db, &node, &mut |key, value| {
        items.push((key.to_vec(), value.to_vec()));
        Ok(())
    })?;
    let mut rows = Vec::new();
    let mut failure = None;
    for (key, value) in items {
        match table.decode_row(db, &key, &value) {
            Ok((row, cardinality)) => {
                for _ in 0..cardinality {
                    rows.push(row.clone());
                }
            }
            Err(err) => failure = Some(err),
        }
    }
    match failure {
        Some(err) => Err(err),
        None => Ok(rows),
    }
}
