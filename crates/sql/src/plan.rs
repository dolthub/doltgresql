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

//! Planning queries into trees of scans, joins, filters, aggregates, sorts, and projections, and running them.

use std::cmp::Ordering;
use std::collections::HashSet;

use pg_query::protobuf::{
    JoinExpr, JoinType, RangeFunction, RangeSubselect, SelectStmt, SetOperation, SortByDir, SortByNulls,
};
use pg_query::{Node, NodeEnum};

use crate::catalog::ColumnType;
use crate::catalog::table::TableDef;
use crate::error::{PgError, Result, code};
use crate::expr::{
    Binder, Expr, Scope, ScopeColumn, coerce, common_type, compare_values, figure_name, node_name, position, typ,
};
use crate::functions::aggregate::{Accumulator, AggCall};
use crate::query::{Ctx, column, scan};
use crate::types::Value;
use crate::{Column, oid};

/// JoinKind is how a join keeps rows that find no match.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum JoinKind {
    Inner,
    Left,
    Right,
    Full,
}

/// SortKey is an ORDER BY key over a plan's rows.
#[derive(Clone, Debug, PartialEq)]
pub struct SortKey {
    pub expr: Expr,
    pub descending: bool,
    pub nulls_first: bool,
}

/// SetOp is a set operation.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum SetOp {
    Union,
    Intersect,
    Except,
}

/// Plan is a node of a query plan, which produces rows.
#[derive(Clone, Debug, PartialEq)]
pub enum Plan {
    /// One row without columns, the input of a SELECT without FROM.
    OneRow,
    Scan(Box<TableDef>),
    /// The rows of one of Dolt's system tables.
    System(crate::dolt::tables::SystemTable),
    /// Rows of expressions, evaluated without an input row.
    Values(Vec<Vec<Expr>>),
    /// The rows a set-returning function returns for its arguments, with a row number when asked, spreading
    /// records over the given number of columns.
    Function {
        index: usize,
        args: Vec<Expr>,
        ordinality: bool,
        width: usize,
    },
    Filter {
        input: Box<Plan>,
        predicate: Expr,
    },
    Project {
        input: Box<Plan>,
        exprs: Vec<Expr>,
    },
    /// The rows of both inputs side by side that meet the condition.
    Join {
        left: Box<Plan>,
        right: Box<Plan>,
        kind: JoinKind,
        condition: Option<Expr>,
    },
    /// One row per group of the input, with the group keys and then the aggregate results.
    Aggregate {
        input: Box<Plan>,
        groups: Vec<Expr>,
        aggregates: Vec<AggCall>,
    },
    Sort {
        input: Box<Plan>,
        keys: Vec<SortKey>,
    },
    /// The first row of each run of rows with equal keys, or of equal rows without keys.
    Distinct {
        input: Box<Plan>,
        keys: Option<Vec<Expr>>,
    },
    Limit {
        input: Box<Plan>,
        limit: Option<Expr>,
        offset: Option<Expr>,
    },
    SetOp {
        op: SetOp,
        all: bool,
        left: Box<Plan>,
        right: Box<Plan>,
    },
    /// A recursive WITH query: its non-recursive term's rows, then its recursive term's rows over the previous
    /// round's rows until a round adds none, keeping duplicates when the flag is set.
    Recursive {
        work_table: usize,
        anchor: Box<Plan>,
        step: Box<Plan>,
        all: bool,
    },
    /// The rows of the previous round of a recursive WITH query.
    WorkTable(usize),
    /// For each input row, the rows of its set-returning calls side by side, padded with NULLs, after its values.
    ProjectSet {
        input: Box<Plan>,
        functions: Vec<Expr>,
    },
    /// The input's rows, each followed by the value of every window call for it.
    Window {
        input: Box<Plan>,
        calls: Vec<crate::window::WindowCall>,
    },
}

/// Cte is a WITH query in scope: its name, its columns, and its plan, or the ID of its working table while its
/// recursive term is being planned.
#[derive(Clone, Debug, PartialEq)]
pub struct Cte {
    pub name: String,
    pub columns: Vec<Column>,
    pub types: Vec<ColumnType>,
    pub plan: Option<Plan>,
    pub work_table: usize,
}

/// Query is a planned query: its plan and the columns of its rows.
#[derive(Clone, Debug, PartialEq)]
pub struct Query {
    pub plan: Plan,
    pub columns: Vec<Column>,
    /// The types of the columns, with their modifiers.
    pub types: Vec<ColumnType>,
}

/// is_aggregate reports whether a function is an aggregate, which a query must group to call.
pub fn is_aggregate(name: &str) -> bool {
    crate::functions::aggregate::exists(name)
}

/// has_aggregate reports whether an expression calls an aggregate outside any subquery.
fn has_aggregate(node: &Node) -> bool {
    match node.node.as_ref() {
        Some(NodeEnum::FuncCall(call)) => {
            let name = call.funcname.iter().filter_map(node_name).next_back().unwrap_or_default();
            (call.over.is_none() && (is_aggregate(name) || call.agg_star)) || call.args.iter().any(has_aggregate)
        }
        Some(NodeEnum::SubLink(_)) | None => false,
        Some(NodeEnum::AExpr(e)) => {
            e.lexpr.as_deref().is_some_and(has_aggregate) || e.rexpr.as_deref().is_some_and(has_aggregate)
        }
        Some(NodeEnum::BoolExpr(e)) => e.args.iter().any(has_aggregate),
        Some(NodeEnum::TypeCast(c)) => c.arg.as_deref().is_some_and(has_aggregate),
        Some(NodeEnum::NullTest(t)) => t.arg.as_deref().is_some_and(has_aggregate),
        Some(NodeEnum::ResTarget(t)) => t.val.as_deref().is_some_and(has_aggregate),
        Some(NodeEnum::SortBy(s)) => s.node.as_deref().is_some_and(has_aggregate),
        Some(NodeEnum::CaseExpr(c)) => {
            c.arg.as_deref().is_some_and(has_aggregate)
                || c.args.iter().any(has_aggregate)
                || c.defresult.as_deref().is_some_and(has_aggregate)
        }
        Some(NodeEnum::CaseWhen(w)) => {
            w.expr.as_deref().is_some_and(has_aggregate) || w.result.as_deref().is_some_and(has_aggregate)
        }
        Some(NodeEnum::CoalesceExpr(c)) => c.args.iter().any(has_aggregate),
        Some(NodeEnum::MinMaxExpr(m)) => m.args.iter().any(has_aggregate),
        Some(NodeEnum::List(l)) => l.items.iter().any(has_aggregate),
        _ => false,
    }
}

/// sort_order reads an ORDER BY item's direction and NULLS placement.
fn sort_order(sort: &pg_query::protobuf::SortBy) -> (bool, bool) {
    let descending = SortByDir::try_from(sort.sortby_dir) == Ok(SortByDir::SortbyDesc);
    let nulls_first = match SortByNulls::try_from(sort.sortby_nulls) {
        Ok(SortByNulls::SortbyNullsFirst) => true,
        Ok(SortByNulls::SortbyNullsLast) => false,
        _ => descending,
    };
    (descending, nulls_first)
}

/// replace_groups replaces the subexpressions of a grouped query's expression that equal a group key with a reference
/// to that key in the aggregate's output, and aggregate references with theirs after the keys.
fn replace_groups(expr: Expr, groups: &[Expr]) -> Expr {
    if let Some(i) = groups.iter().position(|g| *g == expr) {
        return Expr::Column(i);
    }
    match expr {
        Expr::AggRef(k) => Expr::Column(groups.len() + k),
        other => other.map_children(&mut |child| replace_groups(child, groups)),
    }
}

/// ungrouped_column returns a column of the input row that a grouped expression still refers to.
fn ungrouped_column(expr: &Expr) -> Option<usize> {
    let mut found = None;
    expr.visit(&mut |e| {
        if let Expr::InputColumn(i) = e {
            found = found.or(Some(*i));
        }
    });
    found
}

/// mark_input turns column references into input column references, so that grouping can tell which remain.
fn mark_input(expr: Expr) -> Expr {
    match expr {
        Expr::Column(i) => Expr::InputColumn(i),
        other => other.map_children(&mut mark_input),
    }
}

/// Planner plans the queries of a statement, with the scopes of the enclosing queries for correlated subqueries.
pub struct Planner<'b, 'a> {
    pub ctx: &'b mut Ctx<'a>,
    /// The scopes of the enclosing queries, innermost last.
    pub outer: Vec<Scope>,
}

impl<'b, 'a> Planner<'b, 'a> {
    /// binder returns a binder over the scope inside the planner's enclosing scopes.
    fn binder(&mut self, scope: Scope) -> Binder<'_, 'a> {
        let mut scopes = self.outer.clone();
        scopes.push(scope);
        Binder::with_scopes(self.ctx, scopes)
    }

    /// plan_query plans a SELECT, VALUES, or set operation, with the WITH queries it defines in scope.
    pub fn plan_query(&mut self, select: &SelectStmt) -> Result<Query> {
        let Some(with) = select.with_clause.as_ref() else { return self.plan_query_body(select) };
        let depth = self.ctx.ctes.len();
        let result = self.plan_with(with).and_then(|_| self.plan_query_body(select));
        self.ctx.ctes.truncate(depth);
        result
    }

    /// plan_with plans the WITH queries of a WITH clause and brings them into scope.
    fn plan_with(&mut self, with: &pg_query::protobuf::WithClause) -> Result<()> {
        let depth = self.ctx.ctes.len();
        for cte in &with.ctes {
            let Some(NodeEnum::CommonTableExpr(cte)) = cte.node.as_ref() else { continue };
            if self.ctx.ctes[depth..].iter().any(|c| c.name == cte.ctename) {
                return Err(PgError {
                    position: position(cte.location),
                    ..PgError::new(
                        code::DUPLICATE_ALIAS,
                        format!("WITH query name \"{}\" specified more than once", cte.ctename),
                    )
                });
            }
            let Some(NodeEnum::SelectStmt(query)) = cte.ctequery.as_deref().and_then(|q| q.node.as_ref()) else {
                return Err(PgError::unsupported("data-modifying statements in WITH"));
            };
            let aliases: Vec<String> = cte.aliascolnames.iter().filter_map(node_name).map(str::to_string).collect();
            let op = SetOperation::try_from(query.op).unwrap_or(SetOperation::SetopNone);
            let recursive = with.recursive && op == SetOperation::SetopUnion && references(query, &cte.ctename);
            let planned = if recursive {
                self.plan_recursive(cte, query, &aliases)?
            } else {
                let mut planned = self.plan_query(query)?;
                rename_columns(&cte.ctename, &mut planned.columns, &aliases, cte.location)?;
                planned
            };
            let work_table = self.ctx.work_tables.len() + self.ctx.ctes.len();
            self.ctx.ctes.push(Cte {
                name: cte.ctename.clone(),
                columns: planned.columns,
                types: planned.types,
                plan: Some(planned.plan),
                work_table,
            });
        }
        Ok(())
    }

    /// plan_recursive plans a recursive WITH query: its non-recursive term, then its recursive term over a working
    /// table of the non-recursive term's columns.
    fn plan_recursive(
        &mut self,
        cte: &pg_query::protobuf::CommonTableExpr,
        query: &SelectStmt,
        aliases: &[String],
    ) -> Result<Query> {
        let (Some(left), Some(right)) = (query.larg.as_deref(), query.rarg.as_deref()) else {
            return Err(PgError::internal("a recursive query without both terms"));
        };
        let mut anchor = self.plan_query(left)?;
        rename_columns(&cte.ctename, &mut anchor.columns, aliases, cte.location)?;
        let work_table = 1_000_000 + self.ctx.ctes.len();
        self.ctx.ctes.push(Cte {
            name: cte.ctename.clone(),
            columns: anchor.columns.clone(),
            types: anchor.types.clone(),
            plan: None,
            work_table,
        });
        let step = self.plan_query(right);
        self.ctx.ctes.pop();
        let step = step?;
        if step.columns.len() != anchor.columns.len() {
            return Err(PgError {
                position: position(cte.location),
                ..PgError::new(code::SYNTAX_ERROR, "each UNION query must have the same number of columns")
            });
        }
        for (i, (a, s)) in anchor.types.iter().zip(&step.types).enumerate() {
            if a.oid != s.oid {
                let location = left
                    .target_list
                    .get(i)
                    .and_then(|t| match t.node.as_ref() {
                        Some(NodeEnum::ResTarget(t)) => t.val.as_deref().map(crate::expr::arg_location),
                        _ => None,
                    })
                    .unwrap_or(cte.location);
                return Err(PgError {
                    position: position(location),
                    hint: Some("Cast the output of the non-recursive term to the correct type.".into()),
                    ..PgError::new(
                        code::DATATYPE_MISMATCH,
                        format!(
                            "recursive query \"{}\" column {} has type {} in non-recursive term but type {} overall",
                            cte.ctename,
                            i + 1,
                            crate::cast::type_display(a.oid),
                            crate::cast::type_display(s.oid)
                        ),
                    )
                });
            }
        }
        let plan =
            Plan::Recursive { work_table, anchor: Box::new(anchor.plan), step: Box::new(step.plan), all: query.all };
        Ok(Query { plan, columns: anchor.columns, types: anchor.types })
    }

    /// plan_query_body plans a SELECT, VALUES, or set operation without its WITH clause.
    fn plan_query_body(&mut self, select: &SelectStmt) -> Result<Query> {
        let op = SetOperation::try_from(select.op).unwrap_or(SetOperation::SetopNone);
        let (mut query, scope) = match op {
            SetOperation::SetopNone | SetOperation::Undefined if !select.values_lists.is_empty() => {
                self.plan_values(select)?
            }
            SetOperation::SetopNone | SetOperation::Undefined => return self.plan_select(select),
            _ => self.plan_set_operation(select, op)?,
        };
        // ORDER BY and LIMIT of VALUES and set operations apply to the result's columns.
        if !select.sort_clause.is_empty() {
            let mut keys = Vec::new();
            for sort in &select.sort_clause {
                let Some(NodeEnum::SortBy(sort)) = sort.node.as_ref() else { continue };
                let node = sort.node.as_deref().ok_or_else(|| PgError::internal("ORDER BY without a key"))?;
                let expr = match ordinal(node) {
                    Some((n, location)) => {
                        output_ordinal(&query.columns, n, location)?;
                        Expr::Column(n - 1)
                    }
                    None => self.binder(scope.clone()).bind(node)?.0,
                };
                let (descending, nulls_first) = sort_order(sort);
                keys.push(SortKey { expr, descending, nulls_first });
            }
            query.plan = Plan::Sort { input: Box::new(query.plan), keys };
        }
        query.plan = self.limit(query.plan, select)?;
        Ok(query)
    }

    /// limit wraps a plan in its LIMIT and OFFSET.
    fn limit(&mut self, plan: Plan, select: &SelectStmt) -> Result<Plan> {
        if select.limit_count.is_none() && select.limit_offset.is_none() {
            return Ok(plan);
        }
        let mut binder = self.binder(Scope::default());
        let mut count = |node: Option<&Node>| -> Result<Option<Expr>> {
            match node {
                Some(node) => Ok(Some(coerce(binder.bind(node)?, typ(oid::INT8), false, -1)?.0)),
                None => Ok(None),
            }
        };
        let limit = count(select.limit_count.as_deref())?;
        let offset = count(select.limit_offset.as_deref())?;
        Ok(Plan::Limit { input: Box::new(plan), limit, offset })
    }

    /// plan_values plans a VALUES list, whose columns take the common type of their values.
    fn plan_values(&mut self, select: &SelectStmt) -> Result<(Query, Scope)> {
        let mut rows: Vec<Vec<(Expr, ColumnType, i32)>> = Vec::new();
        let mut binder = self.binder(Scope::default());
        for list in &select.values_lists {
            let Some(NodeEnum::List(list)) = list.node.as_ref() else { continue };
            let mut row = Vec::new();
            for item in &list.items {
                let (expr, ty) = binder.bind(item)?;
                row.push((expr, ty, crate::expr::arg_location(item)));
            }
            if let Some(first) = rows.first()
                && first.len() != row.len()
            {
                return Err(PgError {
                    position: position(row.last().map_or(-1, |r| r.2)),
                    ..PgError::new(code::SYNTAX_ERROR, "VALUES lists must all be the same length")
                });
            }
            rows.push(row);
        }
        let width = rows.first().map_or(0, Vec::len);
        let mut types = Vec::with_capacity(width);
        for i in 0..width {
            let column_types: Vec<(ColumnType, i32)> = rows.iter().map(|r| (r[i].1, r[i].2)).collect();
            types.push(common_type(&column_types, "VALUES")?);
        }
        let mut exprs = Vec::with_capacity(rows.len());
        for row in rows {
            let mut out = Vec::with_capacity(width);
            for (i, (expr, ty, location)) in row.into_iter().enumerate() {
                out.push(coerce((expr, ty), types[i], false, location)?.0);
            }
            exprs.push(out);
        }
        let columns: Vec<Column> =
            types.iter().enumerate().map(|(i, &t)| column(format!("column{}", i + 1), t)).collect();
        let scope = Scope {
            columns: columns
                .iter()
                .zip(&types)
                .map(|(c, &ty)| ScopeColumn { table: String::new(), name: c.name.clone(), ty, hidden: false })
                .collect(),
        };
        Ok((Query { plan: Plan::Values(exprs), columns, types }, scope))
    }

    /// plan_set_operation plans UNION, INTERSECT, or EXCEPT, whose columns take the names of the left query and the
    /// common types of both.
    fn plan_set_operation(&mut self, select: &SelectStmt, op: SetOperation) -> Result<(Query, Scope)> {
        let left = select.larg.as_deref().ok_or_else(|| PgError::internal("a set operation without a left side"))?;
        let right = select.rarg.as_deref().ok_or_else(|| PgError::internal("a set operation without a right side"))?;
        let mut left = self.plan_query(left)?;
        let mut right = self.plan_query(right)?;
        let name = match op {
            SetOperation::SetopUnion => "UNION",
            SetOperation::SetopIntersect => "INTERSECT",
            _ => "EXCEPT",
        };
        if left.columns.len() != right.columns.len() {
            return Err(PgError::new(
                code::SYNTAX_ERROR,
                format!("each {name} query must have the same number of columns"),
            ));
        }
        let mut types = Vec::new();
        for (l, r) in left.types.iter().zip(&right.types) {
            types.push(common_type(&[(*l, -1), (*r, -1)], name)?);
        }
        for query in [&mut left, &mut right] {
            if query.types != types {
                let exprs = (0..types.len())
                    .map(|i| coerce((Expr::Column(i), query.types[i]), types[i], false, -1).map(|b| b.0))
                    .collect::<Result<Vec<_>>>()?;
                query.plan = Plan::Project { input: Box::new(std::mem::replace(&mut query.plan, Plan::OneRow)), exprs };
            }
        }
        let columns: Vec<Column> = left.columns.iter().zip(&types).map(|(c, &t)| column(c.name.clone(), t)).collect();
        let scope = Scope {
            columns: columns
                .iter()
                .zip(&types)
                .map(|(c, &ty)| ScopeColumn { table: String::new(), name: c.name.clone(), ty, hidden: false })
                .collect(),
        };
        let op = match op {
            SetOperation::SetopUnion => SetOp::Union,
            SetOperation::SetopIntersect => SetOp::Intersect,
            _ => SetOp::Except,
        };
        let plan = Plan::SetOp { op, all: select.all, left: Box::new(left.plan), right: Box::new(right.plan) };
        Ok((Query { plan, columns, types }, scope))
    }

    /// plan_from plans the FROM list, joining its items, and returns the plan with its scope.
    pub fn plan_from(&mut self, from: &[Node]) -> Result<(Plan, Scope)> {
        let mut result: Option<(Plan, Scope)> = None;
        for item in from {
            let (plan, scope) = self.plan_from_item(item)?;
            result = Some(match result {
                None => (plan, scope),
                Some((left, mut left_scope)) => {
                    check_duplicate_aliases(&left_scope, &scope)?;
                    left_scope.columns.extend(scope.columns);
                    let join = Plan::Join {
                        left: Box::new(left),
                        right: Box::new(plan),
                        kind: JoinKind::Inner,
                        condition: None,
                    };
                    (join, left_scope)
                }
            });
        }
        Ok(result.unwrap_or((Plan::OneRow, Scope::default())))
    }

    /// plan_from_item plans one FROM item.
    fn plan_from_item(&mut self, item: &Node) -> Result<(Plan, Scope)> {
        match item.node.as_ref() {
            Some(NodeEnum::RangeVar(relation)) => {
                if relation.schemaname.is_empty()
                    && let Some(cte) = self.ctx.ctes.iter().rev().find(|c| c.name == relation.relname).cloned()
                {
                    return Ok(self.plan_cte(cte, relation));
                }
                let table = match self.ctx.resolve_table(relation) {
                    Ok(table) => table,
                    Err(err) => {
                        if let Some((_, fragment)) = self.ctx.find_view(&relation.schemaname, &relation.relname)? {
                            return self.plan_view(&fragment, relation);
                        }
                        match crate::dolt::tables::lookup(&relation.schemaname, &relation.relname) {
                            Some(system) => return Ok(self.plan_system(system, relation)),
                            None => return Err(err),
                        }
                    }
                };
                let alias = relation.alias.as_ref();
                let name = alias.map_or(table.name.clone(), |a| a.aliasname.clone());
                let renames: Vec<&str> =
                    alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
                let scope = Scope {
                    columns: table
                        .columns
                        .iter()
                        .enumerate()
                        .map(|(i, c)| ScopeColumn {
                            table: name.clone(),
                            name: renames.get(i).map_or(c.name.clone(), |r| r.to_string()),
                            ty: c.ty,
                            hidden: false,
                        })
                        .collect(),
                };
                Ok((Plan::Scan(Box::new(table)), scope))
            }
            Some(NodeEnum::JoinExpr(join)) => self.plan_join(join),
            Some(NodeEnum::RangeSubselect(subselect)) => self.plan_subselect(subselect),
            Some(NodeEnum::RangeFunction(function)) => self.plan_range_function(function),
            _ => Err(PgError::unsupported("this FROM item")),
        }
    }

    /// plan_cte plans a reference to a WITH query, or to the working table of the recursive query being planned.
    fn plan_cte(&mut self, cte: Cte, relation: &pg_query::protobuf::RangeVar) -> (Plan, Scope) {
        let alias = relation.alias.as_ref();
        let table = alias.map_or(cte.name.clone(), |a| a.aliasname.clone());
        let renames: Vec<&str> = alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
        let columns = cte
            .columns
            .iter()
            .zip(&cte.types)
            .enumerate()
            .map(|(i, (c, &ty))| ScopeColumn {
                table: table.clone(),
                name: renames.get(i).map_or(c.name.clone(), |r| r.to_string()),
                ty,
                hidden: false,
            })
            .collect();
        let plan = cte.plan.unwrap_or(Plan::WorkTable(cte.work_table));
        (plan, Scope { columns })
    }

    /// plan_view plans a view's query in place of the view, naming its columns after the view's column names.
    fn plan_view(&mut self, fragment: &str, relation: &pg_query::protobuf::RangeVar) -> Result<(Plan, Scope)> {
        let (select, aliases) = crate::views::view_query(fragment)?;
        let query = Planner { ctx: self.ctx, outer: Vec::new() }.plan_query(&select)?;
        let alias = relation.alias.as_ref();
        let table = alias.map_or(relation.relname.clone(), |a| a.aliasname.clone());
        let renames: Vec<&str> = alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
        let columns = query
            .columns
            .iter()
            .zip(&query.types)
            .enumerate()
            .map(|(i, (c, &ty))| ScopeColumn {
                table: table.clone(),
                name: renames
                    .get(i)
                    .map(|r| r.to_string())
                    .or_else(|| aliases.get(i).cloned())
                    .unwrap_or_else(|| c.name.clone()),
                ty,
                hidden: false,
            })
            .collect();
        Ok((query.plan, Scope { columns }))
    }

    /// plan_system plans a scan of one of Dolt's system tables.
    fn plan_system(
        &mut self,
        system: crate::dolt::tables::SystemTable,
        relation: &pg_query::protobuf::RangeVar,
    ) -> (Plan, Scope) {
        let alias = relation.alias.as_ref();
        let name = alias.map_or(relation.relname.clone(), |a| a.aliasname.clone());
        let renames: Vec<&str> = alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
        let columns = system
            .columns()
            .into_iter()
            .enumerate()
            .map(|(i, (column, ty))| ScopeColumn {
                table: name.clone(),
                name: renames.get(i).map_or(column.to_string(), |r| r.to_string()),
                ty: typ(ty),
                hidden: false,
            })
            .collect();
        (Plan::System(system), Scope { columns })
    }

    /// plan_subselect plans a subquery in FROM.
    fn plan_subselect(&mut self, subselect: &RangeSubselect) -> Result<(Plan, Scope)> {
        let Some(NodeEnum::SelectStmt(select)) = subselect.subquery.as_deref().and_then(|n| n.node.as_ref()) else {
            return Err(PgError::unsupported("this subquery"));
        };
        let Some(alias) = subselect.alias.as_ref() else {
            return Err(PgError {
                hint: Some("For example, FROM (SELECT ...) [AS] foo.".into()),
                ..PgError::new(code::SYNTAX_ERROR, "subquery in FROM must have an alias")
            });
        };
        let query = Planner { ctx: self.ctx, outer: self.outer.clone() }.plan_query(select)?;
        let renames: Vec<&str> = alias.colnames.iter().filter_map(node_name).collect();
        let scope = Scope {
            columns: query
                .columns
                .iter()
                .zip(&query.types)
                .enumerate()
                .map(|(i, (c, &ty))| ScopeColumn {
                    table: alias.aliasname.clone(),
                    name: renames.get(i).map_or(c.name.clone(), |r| r.to_string()),
                    ty,
                    hidden: false,
                })
                .collect(),
        };
        Ok((query.plan, scope))
    }

    /// plan_range_function plans a set-returning function in FROM.
    fn plan_range_function(&mut self, function: &RangeFunction) -> Result<(Plan, Scope)> {
        let [item] = function.functions.as_slice() else { return Err(PgError::unsupported("ROWS FROM")) };
        let Some(NodeEnum::List(list)) = item.node.as_ref() else { return Err(PgError::unsupported("this function")) };
        let Some(NodeEnum::FuncCall(call)) = list.items.first().and_then(|n| n.node.as_ref()) else {
            return Err(PgError::unsupported("this function in FROM"));
        };
        let name = call.funcname.iter().filter_map(node_name).next_back().unwrap_or_default().to_string();
        let mut binder = self.binder(Scope::default());
        binder.set_functions = Some(Vec::new());
        let (expr, ty) = binder.bind(&Node { node: Some(NodeEnum::FuncCall(call.clone())) })?;
        let expr = match expr {
            Expr::SetRef(k) => binder.set_functions.take().unwrap_or_default().swap_remove(k),
            other => other,
        };
        let Expr::Func(index, args) = expr else { return Err(PgError::unsupported("this function in FROM")) };
        let alias = function.alias.as_ref();
        let table = alias.map_or(name.clone(), |a| a.aliasname.clone());
        let renames: Vec<&str> = alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
        // A function returning one column names the column after the alias when there is one.
        let column_name = renames
            .first()
            .map(|r| r.to_string())
            .unwrap_or_else(|| alias.map_or(name.clone(), |a| a.aliasname.clone()));
        let out_columns = crate::dolt::procedures::OUT_COLUMNS
            .iter()
            .chain(crate::functions::JSON_OUT_COLUMNS)
            .find(|(n, _)| *n == name)
            .map(|(_, c)| *c);
        let mut columns = match out_columns {
            Some(out) => out
                .iter()
                .enumerate()
                .map(|(i, (n, t))| ScopeColumn {
                    table: table.clone(),
                    name: renames.get(i).map_or(n.to_string(), |r| r.to_string()),
                    ty: typ(*t),
                    hidden: false,
                })
                .collect(),
            _ => vec![ScopeColumn { table: table.clone(), name: column_name, ty, hidden: false }],
        };
        let width = columns.len();
        if function.ordinality {
            let name = renames.get(1).map_or("ordinality".to_string(), |r| r.to_string());
            columns.push(ScopeColumn { table, name, ty: typ(oid::INT8), hidden: false });
        }
        Ok((Plan::Function { index, args, ordinality: function.ordinality, width }, Scope { columns }))
    }

    /// plan_join plans a join with its condition, merging the columns that USING or NATURAL name.
    fn plan_join(&mut self, join: &JoinExpr) -> Result<(Plan, Scope)> {
        let left = join.larg.as_deref().ok_or_else(|| PgError::internal("a join without a left side"))?;
        let right = join.rarg.as_deref().ok_or_else(|| PgError::internal("a join without a right side"))?;
        let (left_plan, left_scope) = self.plan_from_item(left)?;
        let (right_plan, right_scope) = self.plan_from_item(right)?;
        check_duplicate_aliases(&left_scope, &right_scope)?;
        let kind = match JoinType::try_from(join.jointype) {
            Ok(JoinType::JoinLeft) => JoinKind::Left,
            Ok(JoinType::JoinRight) => JoinKind::Right,
            Ok(JoinType::JoinFull) => JoinKind::Full,
            _ => JoinKind::Inner,
        };
        let mut scope = Scope { columns: left_scope.columns.iter().chain(&right_scope.columns).cloned().collect() };
        let width = left_scope.columns.len();
        let using: Vec<String> = if join.is_natural {
            left_scope
                .columns
                .iter()
                .filter(|l| !l.hidden && right_scope.columns.iter().any(|r| !r.hidden && r.name == l.name))
                .map(|l| l.name.clone())
                .collect()
        } else {
            join.using_clause.iter().filter_map(node_name).map(str::to_string).collect()
        };
        let condition = if !using.is_empty() || join.is_natural {
            let mut condition: Option<Expr> = None;
            let mut merged = Vec::new();
            for name in &using {
                let find = |s: &Scope, side: &str| {
                    let mut found = s.columns.iter().enumerate().filter(|(_, c)| !c.hidden && c.name == *name);
                    match (found.next(), found.next()) {
                        (Some((i, c)), None) => Ok((i, c.ty)),
                        (Some(_), Some(_)) => Err(PgError::new(
                            code::AMBIGUOUS_COLUMN,
                            format!("common column name \"{name}\" appears more than once in {side} table"),
                        )),
                        _ => Err(PgError::new(
                            code::UNDEFINED_COLUMN,
                            format!("column \"{name}\" specified in USING clause does not exist in {side} table"),
                        )),
                    }
                };
                let (l, lt) = find(&left_scope, "left")?;
                let (r, rt) = find(&right_scope, "right")?;
                let mut binder = self.binder(scope.clone());
                let (test, _) = binder.compare("=", (Expr::Column(l), lt), (Expr::Column(width + r), rt), -1)?;
                condition = Some(match condition {
                    Some(c) => Expr::And(Box::new(c), Box::new(test)),
                    None => test,
                });
                let ty = common_type(&[(lt, -1), (rt, -1)], "JOIN/USING")?;
                let value = match kind {
                    JoinKind::Right => Expr::Column(width + r),
                    JoinKind::Full => Expr::Coalesce(vec![
                        coerce((Expr::Column(l), lt), ty, false, -1)?.0,
                        coerce((Expr::Column(width + r), rt), ty, false, -1)?.0,
                    ]),
                    _ => Expr::Column(l),
                };
                merged.push((name.clone(), value, ty, l, width + r));
            }
            let plan = Plan::Join { left: Box::new(left_plan), right: Box::new(right_plan), kind, condition };
            // The merged columns come first, and the joined columns they replace stay reachable only by table name.
            let mut exprs: Vec<Expr> = merged.iter().map(|m| m.1.clone()).collect();
            let mut columns: Vec<ScopeColumn> = merged
                .iter()
                .map(|m| ScopeColumn { table: String::new(), name: m.0.clone(), ty: m.2, hidden: false })
                .collect();
            let replaced: HashSet<usize> = merged.iter().flat_map(|m| [m.3, m.4]).collect();
            for (i, c) in scope.columns.iter().enumerate() {
                exprs.push(Expr::Column(i));
                let mut c = c.clone();
                if replaced.contains(&i) {
                    c.hidden = true;
                }
                columns.push(c);
            }
            scope = Scope { columns };
            let project = Plan::Project { input: Box::new(plan), exprs };
            return Ok((project, scope));
        } else {
            match join.quals.as_deref() {
                Some(quals) => {
                    let mut binder = self.binder(scope.clone());
                    Some(coerce(binder.bind(quals)?, typ(oid::BOOL), false, -1)?.0)
                }
                None => None,
            }
        };
        let plan = Plan::Join { left: Box::new(left_plan), right: Box::new(right_plan), kind, condition };
        Ok((plan, scope))
    }

    /// plan_select plans a simple SELECT.
    fn plan_select(&mut self, select: &SelectStmt) -> Result<Query> {
        let (mut plan, scope) = self.plan_from(&select.from_clause)?;
        if let Some(node) = select.where_clause.as_deref() {
            if has_aggregate(node) {
                return Err(PgError {
                    position: aggregate_location(node).and_then(position),
                    ..PgError::new(code::GROUPING_ERROR, "aggregate functions are not allowed in WHERE")
                });
            }
            let mut binder = self.binder(scope.clone());
            binder.clause = "WHERE";
            let predicate = coerce(binder.bind(node)?, typ(oid::BOOL), false, -1)?.0;
            plan = Plan::Filter { input: Box::new(plan), predicate };
        }
        let windowed = select.target_list.iter().any(crate::window::has_window)
            || select.sort_clause.iter().any(crate::window::has_window);
        let grouped = !select.group_clause.is_empty()
            || select.having_clause.is_some()
            || select.target_list.iter().any(has_aggregate)
            || select.sort_clause.iter().any(has_aggregate);
        // Bind the targets, expanding stars.
        let mut binder = self.binder(scope.clone());
        if grouped {
            binder.aggregates = Some(Vec::new());
        }
        if windowed {
            binder.windows = Some(Vec::new());
        }
        binder.set_functions = Some(Vec::new());
        binder.named_windows = crate::window::window_names(&select.window_clause);
        let mut targets: Vec<(Expr, ColumnType, String, i32)> = Vec::new();
        for target in &select.target_list {
            let Some(NodeEnum::ResTarget(target)) = target.node.as_ref() else { continue };
            let value = target.val.as_deref().ok_or_else(|| PgError::internal("a target without a value"))?;
            if let Some(NodeEnum::ColumnRef(c)) = value.node.as_ref()
                && matches!(c.fields.last().and_then(|f| f.node.as_ref()), Some(NodeEnum::AStar(_)))
            {
                let table = if c.fields.len() > 1 { c.fields.first().and_then(node_name) } else { None };
                let before = targets.len();
                for (i, sc) in scope.columns.iter().enumerate() {
                    if (table.is_none() && !sc.hidden) || table.is_some_and(|t| sc.table == t && !sc.table.is_empty()) {
                        targets.push((Expr::Column(i), sc.ty, sc.name.clone(), c.location));
                    }
                }
                if table.is_none() && scope.columns.is_empty() {
                    return Err(PgError {
                        position: position(c.location),
                        ..PgError::new(code::SYNTAX_ERROR, "SELECT * with no tables specified is not valid")
                    });
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
            targets.push((expr, ty, name, target.location));
        }
        let names: Vec<String> = targets.iter().map(|t| t.2.clone()).collect();
        // ORDER BY keys may name output columns, refer to them by position, or be expressions of the input.
        let mut sorts: Vec<(Expr, bool, bool)> = Vec::new();
        for sort in &select.sort_clause {
            let Some(NodeEnum::SortBy(sort)) = sort.node.as_ref() else { continue };
            let node = sort.node.as_deref().ok_or_else(|| PgError::internal("ORDER BY without a key"))?;
            let (descending, nulls_first) = sort_order(sort);
            let expr = if let Some((n, location)) = ordinal(node) {
                output_ordinal_named(&names, n, location)?;
                targets[n - 1].0.clone()
            } else if let Some(NodeEnum::ColumnRef(c)) = node.node.as_ref()
                && c.fields.len() == 1
                && let Some(name) = node_name(&c.fields[0])
                && names.iter().filter(|n| *n == name).count() == 1
            {
                targets[names.iter().position(|n| n == name).unwrap()].0.clone()
            } else {
                binder.bind(node)?.0
            };
            sorts.push((expr, descending, nulls_first));
        }
        let windows = binder.windows.take();
        let set_functions = binder.set_functions.take().unwrap_or_default();
        let mut having = None;
        if let Some(node) = select.having_clause.as_deref() {
            binder.clause = "HAVING";
            having = Some(coerce(binder.bind(node)?, typ(oid::BOOL), false, -1)?.0);
        }
        let aggregates = binder.aggregates.take();
        let columns_bound = std::mem::take(&mut binder.columns);
        drop(binder);
        let mut windows = windows;
        let mut set_functions = set_functions;
        if let Some(aggregates) = aggregates {
            // Group keys may refer to output columns by position or name, or be expressions of the input.
            let mut groups: Vec<Expr> = Vec::new();
            let mut binder = self.binder(scope.clone());
            binder.clause = "GROUP BY";
            for node in &select.group_clause {
                let expr = if let Some((n, location)) = ordinal(node) {
                    output_ordinal_named(&names, n, location)?;
                    targets[n - 1].0.clone()
                } else if let Some(NodeEnum::ColumnRef(c)) = node.node.as_ref()
                    && c.fields.len() == 1
                    && let Some(name) = node_name(&c.fields[0])
                    && !scope.columns.iter().any(|sc| sc.name == name)
                    && let Some(i) = names.iter().position(|n| n == name)
                {
                    targets[i].0.clone()
                } else {
                    binder.bind(node)?.0
                };
                groups.push(mark_input(expr));
            }
            drop(binder);
            let finish = |expr: Expr| -> Result<Expr> {
                let expr = replace_groups(mark_input(expr), &groups);
                if let Some(i) = ungrouped_column(&expr) {
                    let column = &scope.columns[i];
                    let location = columns_bound.iter().find(|(c, _)| *c == i).map_or(-1, |(_, l)| *l);
                    let name = if column.table.is_empty() {
                        column.name.clone()
                    } else {
                        format!("{}.{}", column.table, column.name)
                    };
                    return Err(PgError {
                        position: position(location),
                        ..PgError::new(
                            code::GROUPING_ERROR,
                            format!(
                                "column \"{name}\" must appear in the GROUP BY clause or be used in an aggregate function"
                            ),
                        )
                    });
                }
                Ok(expr)
            };
            let mut new_targets = Vec::with_capacity(targets.len());
            for (expr, ty, name, location) in targets {
                new_targets.push((finish(expr)?, ty, name, location));
            }
            targets = new_targets;
            sorts = sorts.into_iter().map(|(e, d, n)| finish(e).map(|e| (e, d, n))).collect::<Result<_>>()?;
            having = having.map(finish).transpose()?;
            set_functions = set_functions.into_iter().map(finish).collect::<Result<_>>()?;
            if let Some(windows) = windows.as_mut() {
                for call in windows.iter_mut() {
                    call.args = call.args.drain(..).map(finish).collect::<Result<_>>()?;
                    call.partition = call.partition.drain(..).map(finish).collect::<Result<_>>()?;
                    call.filter = call.filter.take().map(finish).transpose()?;
                    for key in &mut call.order {
                        key.expr = finish(key.expr.clone())?;
                    }
                }
            }
            let groups_plan = groups.into_iter().map(|g| match g {
                Expr::InputColumn(i) => Expr::Column(i),
                other => unmark_input(other),
            });
            plan = Plan::Aggregate { input: Box::new(plan), groups: groups_plan.collect(), aggregates };
            if let Some(predicate) = having {
                plan = Plan::Filter { input: Box::new(plan), predicate };
            }
        }
        if let Some(calls) = windows.filter(|calls| !calls.is_empty()) {
            let input_width = plan.width();
            let place = |expr: Expr| replace_windows(expr, input_width);
            targets = targets.into_iter().map(|(e, t, n, l)| (place(e), t, n, l)).collect();
            sorts = sorts.into_iter().map(|(e, d, n)| (place(e), d, n)).collect();
            plan = Plan::Window { input: Box::new(plan), calls };
        }
        if !set_functions.is_empty() {
            let input_width = plan.width();
            let place = |expr: Expr| replace_set_functions(expr, input_width);
            targets = targets.into_iter().map(|(e, t, n, l)| (place(e), t, n, l)).collect();
            sorts = sorts.into_iter().map(|(e, d, n)| (place(e), d, n)).collect();
            plan = Plan::ProjectSet { input: Box::new(plan), functions: set_functions };
        }
        let width = targets.len();
        let types: Vec<ColumnType> = targets.iter().map(|t| t.1).collect();
        let columns: Vec<Column> = targets.iter().map(|t| column(t.2.clone(), t.1)).collect();
        let mut exprs: Vec<Expr> = targets.into_iter().map(|t| t.0).collect();
        let distinct = !select.distinct_clause.is_empty();
        let distinct_on: Vec<&Node> = select.distinct_clause.iter().filter(|n| n.node.is_some()).collect();
        // Sort keys that aren't output columns travel as hidden columns after the output.
        let mut keys = Vec::new();
        for (expr, descending, nulls_first) in sorts {
            let index = match exprs.iter().position(|e| *e == expr) {
                Some(i) => i,
                None => {
                    if distinct && distinct_on.is_empty() {
                        return Err(PgError::new(
                            code::INVALID_COLUMN_REFERENCE,
                            "for SELECT DISTINCT, ORDER BY expressions must appear in select list",
                        ));
                    }
                    exprs.push(expr);
                    exprs.len() - 1
                }
            };
            keys.push(SortKey { expr: Expr::Column(index), descending, nulls_first });
        }
        let mut distinct_keys = None;
        if !distinct_on.is_empty() {
            let mut binder = self.binder(scope.clone());
            let mut on = Vec::new();
            for node in distinct_on {
                let expr = if let Some((n, location)) = ordinal(node) {
                    output_ordinal_named(&names, n, location)?;
                    Expr::Column(n - 1)
                } else {
                    let expr = binder.bind(node)?.0;
                    match exprs.iter().position(|e| *e == expr) {
                        Some(i) => Expr::Column(i),
                        None => {
                            exprs.push(expr);
                            Expr::Column(exprs.len() - 1)
                        }
                    }
                };
                on.push(expr);
            }
            distinct_keys = Some(on);
        }
        plan = Plan::Project { input: Box::new(plan), exprs };
        if let Some(on) = &distinct_keys {
            // DISTINCT ON keeps the first row of each group in sort order, so the keys sort first.
            let mut sort_keys: Vec<SortKey> = Vec::new();
            for (i, key) in on.iter().enumerate() {
                match keys.get(i) {
                    Some(k) if k.expr == *key => sort_keys.push(k.clone()),
                    Some(_) => {
                        return Err(PgError::new(
                            code::INVALID_COLUMN_REFERENCE,
                            "SELECT DISTINCT ON expressions must match initial ORDER BY expressions",
                        ));
                    }
                    None => sort_keys.push(SortKey { expr: key.clone(), descending: false, nulls_first: false }),
                }
            }
            sort_keys.extend(keys.iter().skip(on.len()).cloned());
            plan = Plan::Sort { input: Box::new(plan), keys: sort_keys };
            plan = Plan::Distinct { input: Box::new(plan), keys: distinct_keys };
        } else {
            if distinct {
                plan = Plan::Distinct { input: Box::new(plan), keys: None };
            }
            if !keys.is_empty() {
                plan = Plan::Sort { input: Box::new(plan), keys };
            }
        }
        plan = self.limit(plan, select)?;
        if matches!(&plan, Plan::Project { exprs, .. } if exprs.len() == width) {
            return Ok(Query { plan, columns, types });
        }
        let visible = (0..width).map(Expr::Column).collect();
        Ok(Query { plan: Plan::Project { input: Box::new(plan), exprs: visible }, columns, types })
    }
}

/// unmark_input turns input column references back into column references.
fn unmark_input(expr: Expr) -> Expr {
    match expr {
        Expr::InputColumn(i) => Expr::Column(i),
        other => other.map_children(&mut unmark_input),
    }
}

/// aggregate_location returns the location of the first aggregate call in an expression.
fn aggregate_location(node: &Node) -> Option<i32> {
    match node.node.as_ref() {
        Some(NodeEnum::FuncCall(call)) if has_aggregate(node) => Some(call.location),
        Some(NodeEnum::AExpr(e)) => {
            e.lexpr.as_deref().and_then(aggregate_location).or_else(|| e.rexpr.as_deref().and_then(aggregate_location))
        }
        Some(NodeEnum::BoolExpr(e)) => e.args.iter().find_map(aggregate_location),
        _ => None,
    }
}

/// check_duplicate_aliases fails as Postgres does when two FROM items have the same name.
fn check_duplicate_aliases(left: &Scope, right: &Scope) -> Result<()> {
    let names: HashSet<&str> = left.columns.iter().filter(|c| !c.table.is_empty()).map(|c| c.table.as_str()).collect();
    if let Some(c) = right.columns.iter().find(|c| !c.table.is_empty() && names.contains(c.table.as_str())) {
        return Err(PgError::new(
            code::DUPLICATE_ALIAS,
            format!("table name \"{}\" specified more than once", c.table),
        ));
    }
    Ok(())
}

/// ordinal returns the number of an integer constant used as an output column position, with its location.
fn ordinal(node: &Node) -> Option<(usize, i32)> {
    match node.node.as_ref() {
        Some(NodeEnum::AConst(c)) => match &c.val {
            Some(pg_query::protobuf::a_const::Val::Ival(i)) => Some((i.ival.max(0) as usize, c.location)),
            _ => None,
        },
        _ => None,
    }
}

/// output_ordinal checks an output column position.
fn output_ordinal(columns: &[Column], n: usize, location: i32) -> Result<()> {
    if n == 0 || n > columns.len() {
        return Err(PgError {
            position: position(location),
            ..PgError::new(code::INVALID_COLUMN_REFERENCE, format!("ORDER BY position {n} is not in select list"))
        });
    }
    Ok(())
}

/// output_ordinal_named checks an output column position against the output names.
fn output_ordinal_named(names: &[String], n: usize, location: i32) -> Result<()> {
    if n == 0 || n > names.len() {
        return Err(PgError {
            position: position(location),
            ..PgError::new(code::INVALID_COLUMN_REFERENCE, format!("ORDER BY position {n} is not in select list"))
        });
    }
    Ok(())
}

/// compare_sorted orders two rows by sort keys already evaluated into them.
fn compare_sorted(keys: &[SortKey], a: &[Value], b: &[Value]) -> Ordering {
    for (i, key) in keys.iter().enumerate() {
        let ordering = match (&a[i], &b[i]) {
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
}

/// replace_set_functions replaces set-returning call references with the columns a ProjectSet node appends.
fn replace_set_functions(expr: Expr, input_width: usize) -> Expr {
    match expr {
        Expr::SetRef(k) => Expr::Column(input_width + k),
        other => other.map_children(&mut |child| replace_set_functions(child, input_width)),
    }
}

/// replace_windows replaces window call references with the columns that a window node appends after its input's.
fn replace_windows(expr: Expr, input_width: usize) -> Expr {
    match expr {
        Expr::WindowRef(k) => Expr::Column(input_width + k),
        other => other.map_children(&mut |child| replace_windows(child, input_width)),
    }
}

/// dedupe drops rows that equal an earlier row or a row of the existing rows.
fn dedupe(rows: Vec<Vec<Value>>, existing: &[Vec<Value>]) -> Vec<Vec<Value>> {
    let mut seen: std::collections::HashSet<String> = existing.iter().map(|r| row_key(r)).collect();
    rows.into_iter().filter(|r| seen.insert(row_key(r))).collect()
}

/// references reports whether a query refers to a relation by an unqualified name.
fn references(select: &SelectStmt, name: &str) -> bool {
    NodeEnum::SelectStmt(Box::new(select.clone()))
        .nodes()
        .into_iter()
        .any(|(n, ..)| matches!(n, pg_query::NodeRef::RangeVar(r) if r.schemaname.is_empty() && r.relname == name))
}

/// rename_columns names a WITH query's columns after its column list, failing when the list is too long.
fn rename_columns(name: &str, columns: &mut [Column], aliases: &[String], location: i32) -> Result<()> {
    if aliases.len() > columns.len() {
        return Err(PgError {
            position: position(location),
            ..PgError::new(
                code::INVALID_COLUMN_REFERENCE,
                format!(
                    "WITH query \"{name}\" has {} columns available but {} columns specified",
                    columns.len(),
                    aliases.len()
                ),
            )
        });
    }
    for (column, alias) in columns.iter_mut().zip(aliases) {
        column.name = alias.clone();
    }
    Ok(())
}

/// rows_equal reports whether two rows are the same for DISTINCT and set operations, where NULLs are equal.
pub fn rows_equal(a: &[Value], b: &[Value]) -> bool {
    a.len() == b.len()
        && a.iter().zip(b).all(|(l, r)| match (l, r) {
            (Value::Null, Value::Null) => true,
            (Value::Null, _) | (_, Value::Null) => false,
            (l, r) => compare_values(l, r) == Ordering::Equal,
        })
}

/// row_key returns a hashable key of a row for DISTINCT and set operations.
fn row_key(row: &[Value]) -> String {
    row.iter()
        .map(|v| v.output().map_or("\u{0}N".to_string(), |s| format!("{}\u{1}{s}", type_tag(v))))
        .collect::<Vec<_>>()
        .join("\u{2}")
}

/// type_tag distinguishes values whose text is the same but which differ.
fn type_tag(value: &Value) -> &'static str {
    match value {
        Value::Float4(_) | Value::Float8(_) => "f",
        Value::Numeric(_) => "n",
        _ => "",
    }
}

/// limit_value evaluates a LIMIT or OFFSET, failing as Postgres does when it is negative.
fn limit_value(expr: &Option<Expr>, ctx: &mut Ctx<'_>, what: &str, error_code: &'static str) -> Result<Option<i64>> {
    let Some(expr) = expr else { return Ok(None) };
    match expr.eval(ctx, &[])? {
        Value::Int8(n) if n < 0 => Err(PgError::new(error_code, format!("{what} must not be negative"))),
        Value::Int8(n) => Ok(Some(n)),
        _ => Ok(None),
    }
}

impl Plan {
    /// width returns the number of columns of the plan's rows.
    fn width(&self) -> usize {
        match self {
            Plan::OneRow => 0,
            Plan::Scan(table) => table.columns.len(),
            Plan::Recursive { anchor, .. } => anchor.width(),
            Plan::WorkTable(_) => 0,
            Plan::Window { input, calls } => input.width() + calls.len(),
            Plan::ProjectSet { input, functions } => input.width() + functions.len(),
            Plan::System(system) => system.columns().len(),
            Plan::Values(rows) => rows.first().map_or(0, Vec::len),
            Plan::Function { ordinality, width, .. } => width + *ordinality as usize,
            Plan::Filter { input, .. }
            | Plan::Sort { input, .. }
            | Plan::Distinct { input, .. }
            | Plan::Limit { input, .. } => input.width(),
            Plan::Project { exprs, .. } => exprs.len(),
            Plan::Join { left, right, .. } => left.width() + right.width(),
            Plan::Aggregate { groups, aggregates, .. } => groups.len() + aggregates.len(),
            Plan::SetOp { left, .. } => left.width(),
        }
    }

    /// run runs the plan and returns its rows.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        Ok(match self {
            Plan::OneRow => vec![Vec::new()],
            Plan::Scan(table) => scan(ctx.db, table)?,
            Plan::WorkTable(id) => ctx.work_tables.get(id).cloned().unwrap_or_default(),
            Plan::ProjectSet { input, functions } => {
                let mut out = Vec::new();
                for row in input.run(ctx)? {
                    let mut columns = Vec::with_capacity(functions.len());
                    for function in functions {
                        let Expr::Func(index, args) = function else {
                            return Err(PgError::internal("a set-returning call that is not a function"));
                        };
                        let values = args.iter().map(|a| a.eval(ctx, &row)).collect::<Result<Vec<_>>>()?;
                        columns.push(crate::functions::call_set(ctx, *index, &values)?);
                    }
                    let count = columns.iter().map(Vec::len).max().unwrap_or(0);
                    for i in 0..count {
                        let mut new_row = row.clone();
                        new_row.extend(columns.iter().map(|c| c.get(i).cloned().unwrap_or(Value::Null)));
                        out.push(new_row);
                    }
                }
                out
            }
            Plan::Window { input, calls } => {
                let mut rows = input.run(ctx)?;
                for call in calls {
                    let values = call.compute(ctx, &rows)?;
                    for (row, value) in rows.iter_mut().zip(values) {
                        row.push(value);
                    }
                }
                rows
            }
            Plan::Recursive { work_table, anchor, step, all } => {
                let mut result = anchor.run(ctx)?;
                if !*all {
                    result = dedupe(result, &[]);
                }
                let mut working = result.clone();
                while !working.is_empty() {
                    let previous = ctx.work_tables.insert(*work_table, working);
                    let rows = step.run(ctx);
                    match previous {
                        Some(previous) => ctx.work_tables.insert(*work_table, previous),
                        None => ctx.work_tables.remove(work_table),
                    };
                    let mut rows = rows?;
                    if !*all {
                        rows = dedupe(rows, &result);
                    }
                    result.extend(rows.iter().cloned());
                    working = rows;
                }
                result
            }
            Plan::System(system) => system.rows(ctx)?,
            Plan::Values(rows) => {
                let mut out = Vec::with_capacity(rows.len());
                for row in rows {
                    out.push(row.iter().map(|e| e.eval(ctx, &[])).collect::<Result<Vec<_>>>()?);
                }
                out
            }
            Plan::Function { index, args, ordinality, width } => {
                let values = args.iter().map(|a| a.eval(ctx, &[])).collect::<Result<Vec<_>>>()?;
                let rows = crate::functions::call_set(ctx, *index, &values)?;
                rows.into_iter()
                    .enumerate()
                    .map(|(i, v)| {
                        let mut row = match v {
                            Value::Record(fields) if *width > 1 => fields,
                            v => vec![v],
                        };
                        row.resize(*width, Value::Null);
                        if *ordinality {
                            row.push(Value::Int8(i as i64 + 1));
                        }
                        row
                    })
                    .collect()
            }
            Plan::Filter { input, predicate } => {
                let mut out = Vec::new();
                for row in input.run(ctx)? {
                    if predicate.is_true(ctx, &row)? {
                        out.push(row);
                    }
                }
                out
            }
            Plan::Project { input, exprs } => {
                let rows = input.run(ctx)?;
                let mut out = Vec::with_capacity(rows.len());
                for row in rows {
                    out.push(exprs.iter().map(|e| e.eval(ctx, &row)).collect::<Result<Vec<_>>>()?);
                }
                out
            }
            Plan::Join { left, right, kind, condition } => {
                let (left_width, right_width) = (left.width(), right.width());
                let left_rows = left.run(ctx)?;
                let right_rows = right.run(ctx)?;
                let mut out = Vec::new();
                let mut right_matched = vec![false; right_rows.len()];
                for l in &left_rows {
                    let mut matched = false;
                    for (j, r) in right_rows.iter().enumerate() {
                        let mut row = l.clone();
                        row.extend(r.iter().cloned());
                        if condition.as_ref().map_or(Ok(true), |c| c.is_true(ctx, &row))? {
                            matched = true;
                            right_matched[j] = true;
                            out.push(row);
                        }
                    }
                    if !matched && matches!(kind, JoinKind::Left | JoinKind::Full) {
                        let mut row = l.clone();
                        row.extend(std::iter::repeat_n(Value::Null, right_width));
                        out.push(row);
                    }
                }
                if matches!(kind, JoinKind::Right | JoinKind::Full) {
                    for (j, r) in right_rows.iter().enumerate() {
                        if !right_matched[j] {
                            let mut row = vec![Value::Null; left_width];
                            row.extend(r.iter().cloned());
                            out.push(row);
                        }
                    }
                }
                out
            }
            Plan::Aggregate { input, groups, aggregates } => {
                let rows = input.run(ctx)?;
                let mut keys: Vec<Vec<Value>> = Vec::new();
                let mut index: std::collections::HashMap<String, usize> = std::collections::HashMap::new();
                let mut states: Vec<Vec<Accumulator>> = Vec::new();
                for row in &rows {
                    let key = groups.iter().map(|g| g.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
                    let k = row_key(&key);
                    let slot = match index.get(&k) {
                        Some(&slot) => slot,
                        None => {
                            index.insert(k, keys.len());
                            keys.push(key);
                            states.push(aggregates.iter().map(Accumulator::new).collect());
                            keys.len() - 1
                        }
                    };
                    for (agg, state) in aggregates.iter().zip(&mut states[slot]) {
                        state.add(ctx, agg, row)?;
                    }
                }
                // Without groups, an aggregate over no rows still returns one row.
                if groups.is_empty() && keys.is_empty() {
                    keys.push(Vec::new());
                    states.push(aggregates.iter().map(Accumulator::new).collect());
                }
                let mut out = Vec::with_capacity(keys.len());
                for (key, state) in keys.into_iter().zip(states) {
                    let mut row = key;
                    for (agg, s) in aggregates.iter().zip(state) {
                        row.push(s.finish(ctx, agg)?);
                    }
                    out.push(row);
                }
                out
            }
            Plan::Sort { input, keys } => {
                let rows = input.run(ctx)?;
                let mut keyed = Vec::with_capacity(rows.len());
                for row in rows {
                    let values = keys.iter().map(|k| k.expr.eval(ctx, &row)).collect::<Result<Vec<_>>>()?;
                    keyed.push((values, row));
                }
                keyed.sort_by(|a, b| compare_sorted(keys, &a.0, &b.0));
                keyed.into_iter().map(|(_, row)| row).collect()
            }
            Plan::Distinct { input, keys } => {
                let rows = input.run(ctx)?;
                let mut out: Vec<Vec<Value>> = Vec::new();
                match keys {
                    Some(keys) => {
                        let mut previous: Option<Vec<Value>> = None;
                        for row in rows {
                            let key = keys.iter().map(|k| k.eval(ctx, &row)).collect::<Result<Vec<_>>>()?;
                            if previous.as_ref().is_none_or(|p| !rows_equal(p, &key)) {
                                out.push(row);
                            }
                            previous = Some(key);
                        }
                    }
                    None => {
                        let mut seen = HashSet::new();
                        for row in rows {
                            if seen.insert(row_key(&row)) {
                                out.push(row);
                            }
                        }
                    }
                }
                out
            }
            Plan::Limit { input, limit, offset } => {
                let offset = limit_value(offset, ctx, "OFFSET", code::INVALID_ROW_COUNT_IN_RESULT_OFFSET_CLAUSE)?;
                let limit = limit_value(limit, ctx, "LIMIT", code::INVALID_ROW_COUNT_IN_LIMIT_CLAUSE)?;
                let rows = input.run(ctx)?.into_iter().skip(offset.unwrap_or(0) as usize);
                match limit {
                    Some(limit) => rows.take(limit as usize).collect(),
                    None => rows.collect(),
                }
            }
            Plan::SetOp { op, all, left, right } => {
                let left_rows = left.run(ctx)?;
                let right_rows = right.run(ctx)?;
                set_operation(*op, *all, left_rows, right_rows)
            }
        })
    }
}

/// set_operation combines two inputs' rows as UNION, INTERSECT, or EXCEPT do, keeping duplicates with ALL.
fn set_operation(op: SetOp, all: bool, left: Vec<Vec<Value>>, right: Vec<Vec<Value>>) -> Vec<Vec<Value>> {
    use std::collections::HashMap;
    let mut right_counts: HashMap<String, usize> = HashMap::new();
    for row in &right {
        *right_counts.entry(row_key(row)).or_default() += 1;
    }
    let mut out = Vec::new();
    let mut seen = HashSet::new();
    match op {
        SetOp::Union => {
            for row in left.into_iter().chain(right) {
                if all || seen.insert(row_key(&row)) {
                    out.push(row);
                }
            }
        }
        SetOp::Intersect => {
            for row in left {
                let key = row_key(&row);
                let count = right_counts.get_mut(&key);
                match count {
                    Some(n) if *n > 0 => {
                        if all {
                            *n -= 1;
                            out.push(row);
                        } else if seen.insert(key) {
                            out.push(row);
                        }
                    }
                    _ => {}
                }
            }
        }
        SetOp::Except => {
            for row in left {
                let key = row_key(&row);
                match right_counts.get_mut(&key) {
                    Some(n) if *n > 0 => {
                        if all {
                            *n -= 1;
                        }
                    }
                    _ => {
                        if all || seen.insert(key) {
                            out.push(row);
                        }
                    }
                }
            }
        }
    }
    out
}
