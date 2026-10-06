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

use std::cmp::Ordering;
use std::sync::Arc;

use doltdb::database::Database;
use pg_query::protobuf::{RangeVar, SelectStmt, SortByDir, SortByNulls};
use pg_query::{Node, NodeEnum};
use prolly::walk_leaves;

use crate::catalog::builtin_type;
use crate::catalog::table::TableDef;
use crate::error::{PgError, Result, code};
use crate::expr::{Binder, Expr, Scope, ScopeColumn, coerce, compare_values, figure_name, node_name, position, typ};
use crate::txn::Txn;
use crate::types::Value;
use crate::{Column, oid};

/// Ctx is what planning a statement reads: the database and the transaction's working root.
pub struct Ctx<'a> {
    pub db: &'a mut Database,
    pub txn: &'a mut Txn,
    /// The types of the statement's parameters, where 0 is a type not yet known.
    pub parameters: &'a mut Vec<u32>,
    /// The notices the statement raised.
    pub notices: &'a mut Vec<PgError>,
}

/// SortKey is an ORDER BY key over the input row.
#[derive(Clone, Debug)]
struct SortKey {
    expr: Expr,
    descending: bool,
    nulls_first: bool,
}

/// SelectPlan is a planned SELECT.
#[derive(Clone, Debug)]
pub struct SelectPlan {
    pub columns: Vec<Column>,
    from: Option<TableDef>,
    filter: Option<Expr>,
    targets: Vec<Expr>,
    order: Vec<SortKey>,
    limit: Option<Expr>,
    offset: Option<Expr>,
}

/// column returns the description of a result column of the type.
pub fn column(name: String, ty: crate::catalog::ColumnType) -> Column {
    let type_size = builtin_type(ty.oid).map_or(-1, |t| t.definition.typ_length);
    let type_oid = if ty.oid == oid::UNKNOWN { oid::TEXT } else { ty.oid };
    Column {
        name,
        type_oid,
        type_size: if ty.oid == oid::UNKNOWN { -1 } else { type_size },
        type_modifier: ty.modifier,
    }
}

/// SEARCH_PATH is the schemas that unqualified names resolve in.
const SEARCH_PATH: [&str; 1] = ["public"];

impl Ctx<'_> {
    /// resolve_table loads the table that a range variable names.
    pub fn resolve_table(&mut self, relation: &RangeVar) -> Result<TableDef> {
        let schemas: Vec<&str> =
            if relation.schemaname.is_empty() { SEARCH_PATH.to_vec() } else { vec![relation.schemaname.as_str()] };
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

    /// plan_select plans a SELECT.
    pub fn plan_select(&mut self, select: &SelectStmt) -> Result<SelectPlan> {
        if !select.values_lists.is_empty() || select.larg.is_some() {
            return Err(PgError::unsupported("this kind of SELECT"));
        }
        let mut scope = Scope::default();
        let from = match select.from_clause.as_slice() {
            [] => None,
            [node] => match node.node.as_ref() {
                Some(NodeEnum::RangeVar(relation)) => {
                    let table = self.resolve_table(relation)?;
                    let alias = relation.alias.as_ref().map_or(table.name.clone(), |a| a.aliasname.clone());
                    for c in &table.columns {
                        scope.columns.push(ScopeColumn { table: alias.clone(), name: c.name.clone(), ty: c.ty });
                    }
                    Some(table)
                }
                _ => return Err(PgError::unsupported("this FROM item")),
            },
            _ => return Err(PgError::unsupported("joins")),
        };
        let mut binder = Binder { scope: &scope, parameters: self.parameters };
        let filter = match select.where_clause.as_deref() {
            Some(node) => Some(coerce(binder.bind(node)?, typ(oid::BOOL), false, -1)?.0),
            None => None,
        };
        let mut columns = Vec::new();
        let mut targets = Vec::new();
        for target in &select.target_list {
            let Some(NodeEnum::ResTarget(target)) = target.node.as_ref() else { continue };
            let value = target.val.as_deref().ok_or_else(|| PgError::internal("a target without a value"))?;
            if let Some(NodeEnum::ColumnRef(c)) = value.node.as_ref()
                && matches!(c.fields.last().and_then(|f| f.node.as_ref()), Some(NodeEnum::AStar(_)))
            {
                let table = c.fields.first().and_then(node_name);
                let before = targets.len();
                for (i, sc) in scope.columns.iter().enumerate() {
                    if table.is_none_or(|t| sc.table == t) {
                        targets.push(Expr::Column(i));
                        columns.push(column(sc.name.clone(), sc.ty));
                    }
                }
                if targets.len() == before && table.is_some() {
                    return Err(PgError {
                        position: position(c.location),
                        ..PgError::new(
                            code::UNDEFINED_TABLE,
                            format!("missing FROM-clause entry for table \"{}\"", table.unwrap_or_default()),
                        )
                    });
                }
                continue;
            }
            let (expr, ty) = binder.bind(value)?;
            let name = if target.name.is_empty() { figure_name(value) } else { target.name.clone() };
            targets.push(expr);
            columns.push(column(name, ty));
        }
        let mut order = Vec::new();
        for sort in &select.sort_clause {
            let Some(NodeEnum::SortBy(sort)) = sort.node.as_ref() else { continue };
            let node = sort.node.as_deref().ok_or_else(|| PgError::internal("ORDER BY without a key"))?;
            let expr = match node.node.as_ref() {
                Some(NodeEnum::AConst(c)) if matches!(c.val, Some(pg_query::protobuf::a_const::Val::Ival(_))) => {
                    let Some(pg_query::protobuf::a_const::Val::Ival(n)) = &c.val else { unreachable!() };
                    targets.get((n.ival as usize).wrapping_sub(1)).cloned().ok_or_else(|| PgError {
                        position: position(c.location),
                        ..PgError::new(
                            code::INVALID_COLUMN_REFERENCE,
                            format!("ORDER BY position {} is not in select list", n.ival),
                        )
                    })?
                }
                Some(NodeEnum::ColumnRef(c))
                    if c.fields.len() == 1
                        && columns.iter().filter(|col| Some(col.name.as_str()) == node_name(&c.fields[0])).count()
                            == 1
                        && !scope.columns.iter().any(|sc| Some(sc.name.as_str()) == node_name(&c.fields[0])) =>
                {
                    let i = columns.iter().position(|col| Some(col.name.as_str()) == node_name(&c.fields[0])).unwrap();
                    targets[i].clone()
                }
                _ => binder.bind(node)?.0,
            };
            let descending = SortByDir::try_from(sort.sortby_dir) == Ok(SortByDir::SortbyDesc);
            let nulls_first = match SortByNulls::try_from(sort.sortby_nulls) {
                Ok(SortByNulls::SortbyNullsFirst) => true,
                Ok(SortByNulls::SortbyNullsLast) => false,
                _ => descending,
            };
            order.push(SortKey { expr, descending, nulls_first });
        }
        let mut count = |node: Option<&Node>| -> Result<Option<Expr>> {
            match node {
                Some(node) => Ok(Some(coerce(binder.bind(node)?, typ(oid::INT8), false, -1)?.0)),
                None => Ok(None),
            }
        };
        let limit = count(select.limit_count.as_deref())?;
        let offset = count(select.limit_offset.as_deref())?;
        Ok(SelectPlan { columns, from, filter, targets, order, limit, offset })
    }
}

/// scan returns every row of a table in key order, repeating each keyless row by its cardinality.
pub fn scan(db: &mut Database, table: &TableDef) -> Result<Vec<Vec<Value>>> {
    let node = Arc::new(prolly::Node::decode(table.table.primary_index.clone())?);
    let mut rows = Vec::new();
    let mut failure = None;
    walk_leaves(db, &node, &mut |key, value| {
        match table.decode_row(key, value) {
            Ok((row, cardinality)) => {
                for _ in 0..cardinality {
                    rows.push(row.clone());
                }
            }
            Err(err) => failure = Some(err),
        }
        Ok(())
    })?;
    match failure {
        Some(err) => Err(err),
        None => Ok(rows),
    }
}

/// limit_value evaluates a LIMIT or OFFSET, failing as Postgres does when it is negative.
fn limit_value(expr: &Option<Expr>, params: &[Value], what: &str, error_code: &'static str) -> Result<Option<i64>> {
    let Some(expr) = expr else { return Ok(None) };
    match expr.eval(&[], params)? {
        Value::Int8(n) if n < 0 => Err(PgError::new(error_code, format!("{what} must not be negative"))),
        Value::Int8(n) => Ok(Some(n)),
        _ => Ok(None),
    }
}

impl SelectPlan {
    /// run runs the plan and returns its rows.
    pub fn run(&self, db: &mut Database, params: &[Value]) -> Result<Vec<Vec<Value>>> {
        let input = match &self.from {
            Some(table) => scan(db, table)?,
            None => vec![Vec::new()],
        };
        let mut rows: Vec<(Vec<Value>, Vec<Value>)> = Vec::new();
        for row in input {
            if let Some(filter) = &self.filter
                && !filter.is_true(&row, params)?
            {
                continue;
            }
            let keys = self.order.iter().map(|k| k.expr.eval(&row, params)).collect::<Result<Vec<_>>>()?;
            let output = self.targets.iter().map(|t| t.eval(&row, params)).collect::<Result<Vec<_>>>()?;
            rows.push((keys, output));
        }
        if !self.order.is_empty() {
            rows.sort_by(|a, b| {
                for (i, key) in self.order.iter().enumerate() {
                    let ordering = match (&a.0[i], &b.0[i]) {
                        (Value::Null, Value::Null) => Ordering::Equal,
                        (Value::Null, _) => {
                            if key.nulls_first {
                                Ordering::Less
                            } else {
                                Ordering::Greater
                            }
                        }
                        (_, Value::Null) => {
                            if key.nulls_first {
                                Ordering::Greater
                            } else {
                                Ordering::Less
                            }
                        }
                        (l, r) => {
                            let ordering = compare_values(l, r);
                            if key.descending { ordering.reverse() } else { ordering }
                        }
                    };
                    if ordering != Ordering::Equal {
                        return ordering;
                    }
                }
                Ordering::Equal
            });
        }
        let offset = limit_value(&self.offset, params, "OFFSET", code::INVALID_ROW_COUNT_IN_RESULT_OFFSET_CLAUSE)?;
        let limit = limit_value(&self.limit, params, "LIMIT", code::INVALID_ROW_COUNT_IN_LIMIT_CLAUSE)?;
        let rows = rows.into_iter().map(|(_, output)| output).skip(offset.unwrap_or(0) as usize);
        Ok(match limit {
            Some(limit) => rows.take(limit as usize).collect(),
            None => rows.collect(),
        })
    }
}
