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
use std::collections::{BTreeSet, HashSet};

use pg_query::protobuf::{
    CoercionForm, GroupingSetKind, JoinExpr, JoinType, RangeFunction, RangeSubselect, SelectStmt, SetOperation,
    SortByDir, SortByNulls,
};
use pg_query::{Node, NodeEnum};

use crate::catalog::ColumnType;
use crate::catalog::table::TableDef;
use crate::error::{PgError, Result, code};
use crate::expr::{
    Binder, CmpOp, Expr, Scope, ScopeColumn, coerce, common_type, compare_values, figure_name, node_name, position, typ,
};
use crate::functions::aggregate::AggCall;
use crate::query::{Ctx, column};
use crate::types::Value;
use crate::{Column, oid};

/// JoinKind is how a join keeps rows that find no match.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum JoinKind {
    Inner,
    Left,
    Right,
    Full,
    /// Only the left rows that find no match, padded with NULLs as a left join pads them.
    Anti,
    /// Only the left rows that find a match, once each, padded with NULLs as an anti join pads them.
    Semi,
}

impl JoinKind {
    /// tests_matches reports whether the join keeps only left rows, by whether they find a match, as anti and semi
    /// joins do.
    pub fn tests_matches(self) -> bool {
        matches!(self, JoinKind::Anti | JoinKind::Semi)
    }
}

/// JoinMethod is how a join finds the right rows that match each left row.
#[derive(Clone, Debug, PartialEq)]
pub enum JoinMethod {
    /// Not planned yet: the join looks rows up by a primary key or hashes an input as it runs.
    Unplanned,
    /// Not planned yet, and the left input must stay on the left, since its order replaced a sort.
    Ordered,
    /// Each left row looks its matches up in an index of the right input's table, which a scan of no ranges reads,
    /// by the left expressions that give the index's first columns.
    Lookup { scan: Box<crate::indexscan::IndexScan>, keys: Vec<Expr> },
    /// Each left row looks its matches up in an index of the system catalog relation that the right input reads, by
    /// the left expressions that give the index's first columns.
    CatalogLookup { index: &'static crate::pgcatalog::indexes::CatalogIndex, keys: Vec<Expr> },
    /// The right rows are hashed by their side of the condition's equalities.
    Hash,
    /// Each left row is compared with every right row.
    NestedLoop,
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
    /// The rows of a table, with only the given columns read when they are known and the others left NULL.
    Scan(Box<TableDef>, Option<Vec<usize>>),
    /// The rows of a table whose keys in an index lie in ranges.
    IndexScan(Box<crate::indexscan::IndexScan>),
    /// The rows of one of Dolt's system tables.
    System(crate::dolt::tables::SystemTable),
    /// The rows of a system catalog relation.
    Catalog(&'static crate::pgcatalog::CatalogTable),
    /// The rows of a system catalog relation whose keys in one of its indexes lie in ranges.
    CatalogIndexScan(Box<crate::pgcatalog::indexes::CatalogIndexScan>),
    /// Rows of expressions, evaluated without an input row.
    Values(Vec<Vec<Expr>>),
    /// The rows a set-returning function returns for its arguments, with a row number when asked, spreading
    /// records over the given number of columns.
    Function {
        call: Expr,
        ordinality: bool,
        width: usize,
        defined: Option<Vec<(String, ColumnType)>>,
    },
    /// The rows that differ between two queries' results, with a row number when asked.
    QueryDiff(Box<crate::dolt::querydiff::QueryDiff>, bool),
    Filter {
        input: Box<Plan>,
        predicate: Expr,
    },
    Project {
        input: Box<Plan>,
        exprs: Vec<Expr>,
    },
    /// The rows of both inputs side by side that meet the condition, where a lateral right input runs again for
    /// each left row, which it sees as its enclosing row.
    Join {
        left: Box<Plan>,
        right: Box<Plan>,
        kind: JoinKind,
        condition: Option<Expr>,
        lateral: bool,
        method: JoinMethod,
    },
    /// The rows of an XMLTABLE.
    XmlTable(Box<crate::xml::table::XmlTable>),
    /// The rows of a JSON_TABLE.
    JsonTable(Box<crate::jsontable::JsonTable>),
    /// The rows of several set-returning calls side by side, padded with NULLs, with a row number when asked, as
    /// ROWS FROM and unnest of several arrays return them.
    RowsFrom {
        calls: Vec<Expr>,
        ordinality: bool,
    },
    /// One row per group of the input, with the group keys and then the aggregate results. With grouping sets, the
    /// input is grouped by each set of keys in turn, the keys outside the set are NULL, and a last column holds a mask
    /// with a bit set for each key outside the set.
    Aggregate {
        input: Box<Plan>,
        groups: Vec<Expr>,
        aggregates: Vec<AggCall>,
        sets: Option<Vec<Vec<usize>>>,
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
    /// The rows of the previous round of a recursive WITH query, by the query's ID, and their width.
    WorkTable(usize, usize),
    /// For each input row, the rows of its set-returning calls side by side, padded with NULLs, after its values,
    /// leaving NULL the input columns that nothing above reads.
    ProjectSet {
        input: Box<Plan>,
        functions: Vec<Expr>,
        dropped: Vec<usize>,
    },
    /// The input's rows, each followed by the value of every window call for it.
    Window {
        input: Box<Plan>,
        calls: Vec<crate::window::WindowCall>,
    },
    /// A plan that reads nothing of any enclosing row, which runs once for each run of the whole plan, as Postgres'
    /// initplans do.
    Once(Box<Plan>),
}

/// SubqueryRows holds the rows of a plan, with the hash keys of their one column once an IN test asks for them, and the
/// rows by the keys of a filter's equality conditions once the filter above a `Once` plan asks for them.
#[derive(Debug)]
pub struct SubqueryRows {
    pub rows: Vec<Vec<Value>>,
    keys: std::sync::OnceLock<Option<(KeySet, bool)>>,
    pub(crate) index: std::sync::OnceLock<RowIndex>,
}

/// RowIndex is how the filter above a `Once` plan finds its rows: the sides of its equality conditions that read the
/// enclosing rows, with the rows by the values of the sides that read them, or no table when a value has no hash key,
/// and the filter with its constant parts computed, whole and without the equalities that the table answers.
#[derive(Debug)]
pub(crate) struct RowIndex {
    pub(crate) outer: Vec<Expr>,
    pub(crate) predicate: Expr,
    pub(crate) residual: Expr,
    pub(crate) table: Option<(KeyMap<usize>, Vec<Vec<usize>>)>,
}

impl SubqueryRows {
    /// indexed returns the rows of a plan with the index that the filter above it finds them by.
    pub(crate) fn indexed(rows: Vec<Vec<Value>>, index: RowIndex) -> SubqueryRows {
        SubqueryRows { rows, keys: std::sync::OnceLock::new(), index: std::sync::OnceLock::from(index) }
    }

    /// keys returns the set of the rows' hash keys and whether a row is NULL, or None when a value has no hash key.
    pub fn keys(&self) -> Option<&(KeySet, bool)> {
        self.keys
            .get_or_init(|| {
                let (mut keys, mut null) = (KeySet::default(), false);
                for row in &self.rows {
                    match row.first() {
                        Some(Value::Null) | None => null = true,
                        Some(value) => {
                            keys.insert(HashKey::of(value.clone())?);
                        }
                    }
                }
                Some((keys, null))
            })
            .as_ref()
    }
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
pub fn is_aggregate(schema: Option<&str>, name: &str) -> bool {
    crate::functions::aggregate::exists(schema, name)
}

/// has_aggregate reports whether an expression calls an aggregate outside any subquery.
fn has_aggregate(node: &Node) -> bool {
    match node.node.as_ref() {
        Some(NodeEnum::FuncCall(call)) => {
            let (schema, name) = crate::routines::function_names(&call.funcname);
            let schema = (!schema.is_empty()).then_some(schema.as_str());
            (call.over.is_none() && (is_aggregate(schema, &name) || call.agg_star))
                || call.args.iter().any(has_aggregate)
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
        Some(NodeEnum::GroupingFunc(_)) => true,
        _ => false,
    }
}

/// out_columns returns the columns of a call of a built-in function or a routine that returns a row with named
/// columns, as its OUT parameters name them.
pub(crate) fn out_columns(call: &Expr) -> Option<Vec<(String, ColumnType)>> {
    match call {
        Expr::Func(index, _)
            if crate::functions::function(*index).name == "unnest"
                && crate::functions::function(*index).args == [3614] =>
        {
            Some(crate::functions::textsearch::UNNEST_COLUMNS.iter().map(|(n, t)| (n.to_string(), typ(*t))).collect())
        }
        Expr::Func(index, _) => {
            let name = crate::functions::function(*index).name;
            let lists = crate::dolt::procedures::OUT_COLUMNS
                .iter()
                .chain(crate::functions::JSON_OUT_COLUMNS)
                .chain(crate::functions::CATALOG_OUT_COLUMNS)
                .chain(crate::functions::textsearch::OUT_COLUMNS);
            let (_, columns) = lists.into_iter().find(|(n, _)| *n == name)?;
            Some(columns.iter().map(|(n, t)| (n.to_string(), typ(*t))).collect())
        }
        Expr::Routine(routine, _) => (!routine.columns.is_empty()).then(|| routine.columns.clone()),
        _ => None,
    }
}

/// grouping_sets expands an item of GROUP BY into the lists of expressions that it groups by, as Postgres'
/// expand_grouping_sets does: an expression or an implicit row is one list, ROLLUP drops its items from the end one
/// at a time, CUBE takes every subset of its items, and GROUPING SETS joins the lists of its elements.
fn grouping_sets(node: &Node) -> Vec<Vec<&Node>> {
    let Some(NodeEnum::GroupingSet(set)) = node.node.as_ref() else { return vec![row_items_of(node)] };
    let items: Vec<Vec<&Node>> = set.content.iter().map(row_items_of).collect();
    match GroupingSetKind::try_from(set.kind) {
        Ok(GroupingSetKind::GroupingSetEmpty) => vec![Vec::new()],
        Ok(GroupingSetKind::GroupingSetRollup) => (0..=items.len()).rev().map(|n| items[..n].concat()).collect(),
        Ok(GroupingSetKind::GroupingSetCube) => (0..1usize << items.len())
            .map(|mask| {
                items.iter().enumerate().filter(|(i, _)| mask & (1 << i) != 0).flat_map(|(_, n)| n.clone()).collect()
            })
            .collect(),
        Ok(GroupingSetKind::GroupingSetSets) => set.content.iter().flat_map(grouping_sets).collect(),
        _ => vec![items.concat()],
    }
}

/// row_items_of returns the expressions of an implicit row, which grouping treats as a list, or else the expression.
fn row_items_of(node: &Node) -> Vec<&Node> {
    match node.node.as_ref() {
        Some(NodeEnum::RowExpr(row)) if row.row_format == CoercionForm::CoerceImplicitCast as i32 => {
            row.args.iter().collect()
        }
        _ => vec![node],
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

/// place_grouping checks that the arguments of each GROUPING in a grouped expression are group keys, and points each
/// GROUPING at the column of the grouping set mask.
fn place_grouping(expr: Expr, keys: usize, mask: Option<usize>) -> Result<Expr> {
    let mut invalid = None;
    expr.visit(&mut |e| {
        if let Expr::Grouping(args, locations, _) = e
            && let Some(i) = args.iter().position(|a| !matches!(a, Expr::Column(i) if *i < keys))
        {
            invalid = invalid.or(Some(locations[i]));
        }
    });
    if let Some(location) = invalid {
        return Err(PgError {
            position: position(location),
            ..PgError::new(
                code::GROUPING_ERROR,
                "arguments to GROUPING must be grouping expressions of the associated query level",
            )
        });
    }
    Ok(set_mask(expr, mask))
}

/// set_mask points each GROUPING in an expression at the column of the grouping set mask.
fn set_mask(expr: Expr, mask: Option<usize>) -> Expr {
    match expr {
        Expr::Grouping(args, locations, _) => Expr::Grouping(args, locations, mask),
        other => other.map_children(&mut |child| set_mask(child, mask)),
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
        if let Some(into) = &select.into_clause {
            return Err(PgError {
                position: position(into.rel.as_ref().map_or(-1, |r| r.location)),
                ..PgError::new(code::SYNTAX_ERROR, "SELECT ... INTO is not allowed here")
            });
        }
        let Some(with) = select.with_clause.as_ref() else { return self.plan_query_body(select) };
        let depth = self.ctx.ctes.len();
        let result = self.plan_with(with).and_then(|_| self.plan_query_body(select));
        self.ctx.ctes.truncate(depth);
        result
    }

    /// plan_with plans the WITH queries of a WITH clause and brings them into scope.
    pub(crate) fn plan_with(&mut self, with: &pg_query::protobuf::WithClause) -> Result<()> {
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
            let mut aliases: Vec<String> = cte.aliascolnames.iter().filter_map(node_name).map(str::to_string).collect();
            let recursive = with.recursive && references(query, &cte.ctename);
            if recursive {
                check_recursion(cte, query)?;
            }
            if with.recursive && mutually_recursive(with, cte) {
                return Err(PgError {
                    position: position(cte.location),
                    ..PgError::new(
                        code::FEATURE_NOT_SUPPORTED,
                        "mutual recursion between WITH items is not implemented",
                    )
                });
            }
            if cte.search_clause.is_some() || cte.cycle_clause.is_some() {
                self.check_search_and_cycle(cte, query, recursive)?;
            }
            let planned = if recursive {
                if cte.search_clause.is_some() || cte.cycle_clause.is_some() {
                    let rewritten = rewrite_search_and_cycle(cte, query, &mut aliases)?;
                    self.plan_recursive(cte, &rewritten, &aliases)?
                } else {
                    self.plan_recursive(cte, query, &aliases)?
                }
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

    /// check_search_and_cycle fails as Postgres' transformWithClause and analyzeCTE do for a SEARCH or CYCLE clause
    /// that its WITH query cannot have.
    fn check_search_and_cycle(
        &mut self,
        cte: &pg_query::protobuf::CommonTableExpr,
        query: &SelectStmt,
        recursive: bool,
    ) -> Result<()> {
        if !recursive {
            return Err(PgError {
                position: position(cte.location),
                ..PgError::new(code::SYNTAX_ERROR, "WITH query is not recursive")
            });
        }
        let leaf = |side: Option<&SelectStmt>| {
            side.is_some_and(|s| {
                SetOperation::try_from(s.op).unwrap_or(SetOperation::SetopNone) == SetOperation::SetopNone
            })
        };
        if !leaf(query.larg.as_deref()) {
            return Err(PgError::new(
                code::FEATURE_NOT_SUPPORTED,
                "with a SEARCH or CYCLE clause, the left side of the UNION must be a SELECT",
            ));
        }
        if !leaf(query.rarg.as_deref()) {
            return Err(PgError::new(
                code::SYNTAX_ERROR,
                "with a SEARCH or CYCLE clause, the right side of the UNION must be a SELECT",
            ));
        }
        let Some(cycle) = cte.cycle_clause.as_ref() else { return Ok(()) };
        let names: Vec<String> = match cte.aliascolnames.is_empty() {
            true => self
                .plan_query(query.larg.as_deref().expect("a UNION has a left side"))?
                .columns
                .iter()
                .map(|c| c.name.clone())
                .collect(),
            false => cte.aliascolnames.iter().filter_map(node_name).map(str::to_string).collect(),
        };
        let at = |message: String, code: &'static str| PgError {
            position: position(cycle.location),
            ..PgError::new(code, message)
        };
        let mut seen = Vec::new();
        for column in cycle.cycle_col_list.iter().filter_map(node_name) {
            if seen.contains(&column) {
                return Err(at(format!("cycle column \"{column}\" specified more than once"), code::DUPLICATE_COLUMN));
            }
            if !names.iter().any(|n| n == column) {
                return Err(at(format!("cycle column \"{column}\" not in WITH query column list"), code::SYNTAX_ERROR));
            }
            seen.push(column);
        }
        if cycle.cycle_mark_column == cycle.cycle_path_column {
            return Err(at(
                "cycle mark column name and cycle path column name are the same".into(),
                code::SYNTAX_ERROR,
            ));
        }
        for added in [&cycle.cycle_mark_column, &cycle.cycle_path_column] {
            if !names.contains(added) {
                continue;
            }
            let right =
                NodeEnum::SelectStmt(Box::new(query.rarg.as_deref().expect("a UNION has a right side").clone()));
            let location = right.nodes().into_iter().find_map(|(n, ..)| match n {
                pg_query::NodeRef::ColumnRef(c) if c.fields.iter().filter_map(node_name).next_back() == Some(added) => {
                    Some(c.location)
                }
                _ => None,
            });
            return Err(PgError {
                position: position(location.unwrap_or(cycle.location)),
                ..PgError::new(code::AMBIGUOUS_COLUMN, format!("column reference \"{added}\" is ambiguous"))
            });
        }
        if let (Some(value), Some(default)) = (cycle.cycle_mark_value.as_deref(), cycle.cycle_mark_default.as_deref()) {
            let mut binder = self.binder(Scope::default());
            let types = [
                (binder.bind(value)?.1, crate::expr::arg_location(value)),
                (binder.bind(default)?.1, crate::expr::arg_location(default)),
            ];
            crate::expr::common_type(&types, "CYCLE")?;
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
        for (ty, column) in anchor.types.iter_mut().zip(&mut anchor.columns) {
            if ty.oid == oid::UNKNOWN {
                *ty = typ(oid::TEXT);
                column.type_oid = oid::TEXT;
            }
        }
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
            if a != s {
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
                            crate::cast::format_type(a.oid, Some(a.modifier)).unwrap_or_default(),
                            crate::cast::format_type(s.oid, Some(-1)).unwrap_or_default()
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
        let mut count = |node: Option<&Node>, clause: &str| -> Result<Option<Expr>> {
            let Some(node) = node else { return Ok(None) };
            let bound = binder.bind(node)?;
            if let Expr::Param(i) = bound.0
                && binder.ctx.parameters[i] == 0
            {
                binder.ctx.parameters[i] = oid::INT8;
            }
            let location = crate::expr::arg_location(node);
            if !crate::expr::assignable(bound.1.oid, oid::INT8) {
                return Err(PgError {
                    position: position(location),
                    ..PgError::new(
                        code::DATATYPE_MISMATCH,
                        format!(
                            "argument of {clause} must be type bigint, not type {}",
                            crate::cast::type_display(bound.1.oid)
                        ),
                    )
                });
            }
            Ok(Some(crate::expr::assign(bound, typ(oid::INT8), "", location)?.0))
        };
        let limit = count(select.limit_count.as_deref(), "LIMIT")?;
        let offset = count(select.limit_offset.as_deref(), "OFFSET")?;
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
                .map(|(c, &ty)| ScopeColumn {
                    table: String::new(),
                    name: c.name.clone(),
                    ty,
                    hidden: false,
                    origin: (0, 0),
                })
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
                .map(|(c, &ty)| ScopeColumn {
                    table: String::new(),
                    name: c.name.clone(),
                    ty,
                    hidden: false,
                    origin: (0, 0),
                })
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
            result = Some(match result {
                None => self.plan_from_item(item)?,
                Some((left, mut left_scope)) => {
                    let lateral = is_lateral(item);
                    let (plan, scope) = self.plan_lateral_item(item, &left_scope, lateral)?;
                    check_duplicate_aliases(&left_scope, &scope)?;
                    left_scope.columns.extend(scope.columns);
                    let join = Plan::Join {
                        left: Box::new(left),
                        right: Box::new(plan),
                        kind: JoinKind::Inner,
                        condition: None,
                        lateral,
                        method: JoinMethod::Unplanned,
                    };
                    (join, left_scope)
                }
            });
        }
        Ok(result.unwrap_or((Plan::OneRow, Scope::default())))
    }

    /// use_indexes replaces each table scan under a filter with an index scan when an index answers the filter, and
    /// drops the filter when the scan's ranges hold exactly its rows, as go-mysql-server's costedIndexScans does.
    pub(crate) fn use_indexes(&mut self, plan: Plan) -> Plan {
        match plan {
            Plan::Filter { input, predicate } => match *input {
                Plan::Scan(table, needed) => match crate::indexscan::choose_with_cover(self.ctx, &table, &predicate) {
                    Some((mut scan, exact)) if scan.index.is_some() && self.lookups_cost_more(&scan) => {
                        scan.lookup_heavy = Some(exact);
                        Plan::Filter { input: Box::new(Plan::IndexScan(Box::new(scan))), predicate }
                    }
                    Some((scan, true)) => Plan::IndexScan(Box::new(scan)),
                    Some((scan, false)) => Plan::Filter { input: Box::new(Plan::IndexScan(Box::new(scan))), predicate },
                    None => Plan::Filter { input: Box::new(Plan::Scan(table, needed)), predicate },
                },
                Plan::Catalog(table) => match crate::pgcatalog::indexes::choose(self.ctx, table, &predicate) {
                    Some((scan, true)) => Plan::CatalogIndexScan(Box::new(scan)),
                    Some((scan, false)) => {
                        Plan::Filter { input: Box::new(Plan::CatalogIndexScan(Box::new(scan))), predicate }
                    }
                    None => Plan::Filter { input: Box::new(Plan::Catalog(table)), predicate },
                },
                other => Plan::Filter { input: Box::new(self.use_indexes(other)), predicate },
            },
            Plan::Join { left, right, kind, condition, lateral, method } => Plan::Join {
                left: Box::new(self.use_indexes(*left)),
                right: Box::new(self.use_indexes(*right)),
                kind,
                condition,
                lateral,
                method,
            },
            other => other,
        }
    }

    /// lookups_cost_more reports whether looking up the rows that a scan of a secondary index reads in the primary
    /// index would cost more than reading the whole table in order, judging from the scan's first range alone so
    /// that the check stays cheap for the many small scans. A table whose rows fit in one node is never worth it.
    fn lookups_cost_more(&mut self, scan: &crate::indexscan::IndexScan) -> bool {
        let Ok(root) = prolly::Node::decode(scan.table.table.primary_index.clone()) else { return false };
        !root.is_leaf()
            && scan
                .estimate_with_samples(self.ctx, 1)
                .is_ok_and(|estimate| estimate * crate::joins::SEEK >= root.tree_count() as f64)
    }

    /// plan_lateral_item plans a FROM item after others, which sees their columns when it is lateral and otherwise
    /// fails as Postgres does when it refers to one of them.
    fn plan_lateral_item(&mut self, item: &Node, left: &Scope, lateral: bool) -> Result<(Plan, Scope)> {
        if !lateral {
            return self.plan_from_item(item).map_err(|err| {
                let table = err.message.strip_prefix("missing FROM-clause entry for table \"").and_then(|t| t.strip_suffix('"'));
                match table {
                    Some(table) if err.code == code::UNDEFINED_TABLE && left.columns.iter().any(|c| c.table == table) => PgError {
                        message: format!("invalid reference to FROM-clause entry for table \"{table}\""),
                        hint: Some(format!(
                            "There is an entry for table \"{table}\", but it cannot be referenced from this part of the query."
                        )),
                        ..err
                    },
                    _ => err,
                }
            });
        }
        self.outer.push(left.clone());
        let planned = self.plan_from_item(item);
        self.outer.pop();
        planned
    }

    /// plan_json_table plans a JSON_TABLE, whose columns are named after its alias.
    fn plan_json_table(&mut self, table: &pg_query::protobuf::JsonTable) -> Result<(Plan, Scope)> {
        let mut binder = self.binder(Scope::default());
        let planned = crate::jsontable::plan(&mut binder, table)?;
        let alias = table.alias.as_ref();
        let name = alias.map_or("json_table".to_string(), |a| a.aliasname.clone());
        let renames: Vec<&str> = alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
        let scope = Scope {
            columns: planned
                .columns
                .iter()
                .enumerate()
                .map(|(i, c)| ScopeColumn {
                    table: name.clone(),
                    name: renames.get(i).map_or(c.name.clone(), |r| r.to_string()),
                    ty: c.ty,
                    hidden: false,
                    origin: (0, 0),
                })
                .collect(),
        };
        Ok((Plan::JsonTable(Box::new(planned)), scope))
    }

    /// plan_xml_table plans an XMLTABLE.
    fn plan_xml_table(&mut self, function: &pg_query::protobuf::RangeTableFunc) -> Result<(Plan, Scope)> {
        use crate::xml::table::{XmlColumn, XmlTable};
        let mut binder = self.binder(Scope::default());
        let document_node =
            function.docexpr.as_deref().ok_or_else(|| PgError::internal("XMLTABLE without a document"))?;
        let document = binder.xml_arg(document_node, oid::XML, "XMLTABLE")?;
        let row_node = function.rowexpr.as_deref().ok_or_else(|| PgError::internal("XMLTABLE without a row path"))?;
        let row = binder.xml_arg(row_node, oid::TEXT, "XMLTABLE")?;
        let mut namespaces = Vec::new();
        for namespace in &function.namespaces {
            let Some(NodeEnum::ResTarget(target)) = namespace.node.as_ref() else { continue };
            if target.name.is_empty() {
                return Err(PgError::new(code::FEATURE_NOT_SUPPORTED, "DEFAULT namespace is not supported"));
            }
            let value = target.val.as_deref().ok_or_else(|| PgError::internal("a namespace without a URI"))?;
            namespaces.push((target.name.clone(), binder.xml_arg(value, oid::TEXT, "XMLTABLE")?));
        }
        let mut columns: Vec<XmlColumn> = Vec::new();
        let mut ordinality = false;
        for node in &function.columns {
            let Some(NodeEnum::RangeTableFuncCol(column)) = node.node.as_ref() else { continue };
            let located = |message: String| PgError {
                position: position(column.location),
                ..PgError::new(code::SYNTAX_ERROR, message)
            };
            if columns.iter().any(|c| c.name == column.colname) {
                return Err(located(format!("column name \"{}\" is not unique", column.colname)));
            }
            if column.for_ordinality {
                if ordinality {
                    return Err(located("only one FOR ORDINALITY column is allowed".into()));
                }
                ordinality = true;
                columns.push(XmlColumn {
                    name: column.colname.clone(),
                    ty: typ(oid::INT4),
                    path: None,
                    default: None,
                    not_null: false,
                });
                continue;
            }
            let type_name = column.type_name.as_ref().ok_or_else(|| PgError::internal("a column without a type"))?;
            binder.ctx.prepare_type(type_name)?;
            let ty = crate::expr::resolve_type_name(type_name)?;
            let path = match column.colexpr.as_deref() {
                Some(path) => binder.xml_arg(path, oid::TEXT, "XMLTABLE")?,
                None => Expr::Const(Value::Text(column.colname.clone())),
            };
            let default = match column.coldefexpr.as_deref() {
                Some(default) => Some(binder.typed_arg(default, ty, "XMLTABLE")?),
                None => None,
            };
            columns.push(XmlColumn {
                name: column.colname.clone(),
                ty,
                path: Some(path),
                default,
                not_null: column.is_not_null,
            });
        }
        let alias = function.alias.as_ref();
        let table = alias.map_or("xmltable".to_string(), |a| a.aliasname.clone());
        let renames: Vec<&str> = alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
        let scope = Scope {
            columns: columns
                .iter()
                .enumerate()
                .map(|(i, c)| ScopeColumn {
                    table: table.clone(),
                    name: renames.get(i).map_or(c.name.clone(), |r| r.to_string()),
                    ty: c.ty,
                    hidden: false,
                    origin: (0, 0),
                })
                .collect(),
        };
        Ok((Plan::XmlTable(Box::new(XmlTable { document, row, namespaces, columns })), scope))
    }

    /// plan_sequence plans a sequence read as a relation: one row of its last value, the values it logged ahead, which
    /// are none here, and whether nextval has handed the last value out.
    fn plan_sequence(
        &mut self,
        sequence: objects::Sequence,
        relation: &pg_query::protobuf::RangeVar,
    ) -> Result<(Plan, Scope)> {
        let (schema, name) = crate::sequences::schema_and_name(&sequence);
        self.ctx.require(&crate::auth::Object::Sequence(schema, name.clone()), "r", relation.location)?;
        let sequence = self.ctx.latest(sequence)?;
        let table = relation.alias.as_ref().map_or(name, |a| a.aliasname.clone());
        let columns = [("last_value", oid::INT8), ("log_cnt", oid::INT8), ("is_called", oid::BOOL)]
            .into_iter()
            .map(|(name, ty)| ScopeColumn {
                table: table.clone(),
                name: name.into(),
                ty: typ(ty),
                hidden: false,
                origin: (0, 0),
            })
            .collect();
        let last_value = match sequence.has_been_called && !sequence.is_at_end {
            true => sequence.current - sequence.increment,
            false => sequence.current,
        };
        let row = vec![
            Expr::Const(Value::Int8(last_value)),
            Expr::Const(Value::Int8(0)),
            Expr::Const(Value::Bool(sequence.has_been_called)),
        ];
        Ok((Plan::Values(vec![row]), Scope { columns }))
    }

    /// plan_from_item plans one FROM item, failing as Postgres does when its alias names more columns than it has.
    fn plan_from_item(&mut self, item: &Node) -> Result<(Plan, Scope)> {
        let (plan, scope) = self.plan_item(item)?;
        let alias = match item.node.as_ref() {
            Some(NodeEnum::RangeVar(relation)) => relation.alias.as_ref(),
            Some(NodeEnum::RangeFunction(function)) => function.alias.as_ref(),
            _ => None,
        };
        let available = scope.columns.iter().filter(|c| !c.hidden).count();
        if let Some(alias) = alias.filter(|a| a.colnames.len() > available) {
            return Err(PgError::new(
                code::INVALID_COLUMN_REFERENCE,
                format!(
                    "table \"{}\" has {available} columns available but {} columns specified",
                    alias.aliasname,
                    alias.colnames.len()
                ),
            ));
        }
        Ok((plan, scope))
    }

    /// plan_item plans one FROM item.
    fn plan_item(&mut self, item: &Node) -> Result<(Plan, Scope)> {
        match item.node.as_ref() {
            Some(NodeEnum::RangeVar(relation)) => {
                if relation.schemaname.is_empty()
                    && let Some(cte) = self.ctx.ctes.iter().rev().find(|c| c.name == relation.relname).cloned()
                {
                    return Ok(self.plan_cte(cte, relation));
                }
                if let Some(catalog) = self.ctx.catalog_relation(&relation.schemaname, &relation.relname)? {
                    self.ctx.require_catalog(catalog.name)?;
                    return Ok(self.plan_catalog(catalog, relation));
                }
                let as_of =
                    self.ctx.session.as_of.iter().find(|(l, _)| *l == relation.location).map(|(_, r)| r.clone());
                let qualified;
                let relation = match self.ctx.view_before_table(relation)? {
                    Some(schema) if as_of.is_none() => {
                        qualified = pg_query::protobuf::RangeVar { schemaname: schema, ..relation.clone() };
                        &qualified
                    }
                    _ => relation,
                };
                let resolved = match (&as_of, self.ctx.catalog_root(relation)?) {
                    (Some(revision), _) => self.ctx.resolve_table_as_of(relation, revision),
                    (None, Some(root)) => self.ctx.resolve_table_in(relation, &root),
                    (None, None) => self.ctx.resolve_table(relation),
                };
                let table = match resolved {
                    Ok(table) => table,
                    Err(err) => {
                        if let Some(sequence) = self.ctx.find_sequence(relation)? {
                            return self.plan_sequence(sequence, relation);
                        }
                        if let Some(view) =
                            crate::dolt::diff::blame_view(self.ctx, &relation.schemaname, &relation.relname)?
                        {
                            return self.plan_view(&view, relation);
                        }
                        if let Some((schema, fragment)) = self.ctx.find_view(&relation.schemaname, &relation.relname)? {
                            self.ctx.require_view(&schema, &relation.relname, "r", relation.location)?;
                            let key = (schema.clone(), relation.relname.clone());
                            if self.ctx.expanding.contains(&key) {
                                return Err(PgError::new(
                                    code::INVALID_OBJECT_DEFINITION,
                                    format!(
                                        "infinite recursion detected in rules for relation \"{}\"",
                                        relation.relname
                                    ),
                                ));
                            }
                            self.ctx.expanding.push(key);
                            let object = crate::auth::Object::Table(schema.clone(), relation.relname.clone());
                            let owner = self.ctx.owner_name(&object)?;
                            let role = std::mem::replace(&mut self.ctx.session.role, owner);
                            let schema = self.ctx.session.view_schema.replace(schema);
                            let planned = self.plan_view(&fragment, relation);
                            self.ctx.session.view_schema = schema;
                            self.ctx.session.role = role;
                            self.ctx.expanding.pop();
                            return planned;
                        }
                        let schema = match relation.schemaname.as_str() {
                            "" if self.ctx.session.search_path().iter().any(|s| s == "dolt") => "dolt",
                            schema => schema,
                        };
                        let schema_exists = relation.schemaname.is_empty()
                            || self.ctx.txn.root.schemas.iter().any(|s| s == relation.schemaname.as_bytes());
                        if let Some(system) = crate::dolt::tables::lookup(schema, &relation.relname)
                            .or_else(|| crate::dolt::tables::lookup(&relation.schemaname, &relation.relname))
                            .filter(|s| !s.per_schema() || schema_exists)
                        {
                            return Ok(self.plan_system(system, relation));
                        }
                        if let Some(table) =
                            crate::dolt::conflicts::lookup(self.ctx, &relation.schemaname, &relation.relname)?
                        {
                            let system = crate::dolt::tables::SystemTable::Artifacts(Box::new(table));
                            return Ok(self.plan_system(system, relation));
                        }
                        if let Some(table) = crate::dolt::objmerge::ObjectConflictTable::lookup(
                            self.ctx,
                            &relation.schemaname,
                            &relation.relname,
                        )? {
                            let system = crate::dolt::tables::SystemTable::ObjectConflicts(Box::new(table));
                            return Ok(self.plan_system(system, relation));
                        }
                        return match crate::dolt::diff::lookup(self.ctx, &relation.schemaname, &relation.relname)? {
                            Some(mut table) => {
                                if let Some(revision) = &as_of {
                                    let revision = self.ctx.constant_text(revision)?;
                                    let head = self.ctx.txn.head;
                                    table.head = Some(crate::dolt::history::resolve(self.ctx.db, head, &revision)?);
                                }
                                Ok(self.plan_system(crate::dolt::tables::SystemTable::User(Box::new(table)), relation))
                            }
                            None => Err(err),
                        };
                    }
                };
                if table.schema == "dolt"
                    && let Some(system) = crate::dolt::tables::lookup("dolt", &table.name)
                {
                    return Ok(self.plan_system(system, relation));
                }
                if !table.name.starts_with("dolt_") {
                    let object = crate::auth::Object::Table(table.schema.clone(), table.name.clone());
                    self.ctx.require(&object, "r", relation.location)?;
                }
                let alias = relation.alias.as_ref();
                let name = alias.map_or(table.name.clone(), |a| a.aliasname.clone());
                let renames: Vec<&str> =
                    alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
                let table_oid = crate::pgcatalog::snapshot::table_oid(&table.schema, &table.name);
                crate::usertypes::register_row_type(&table);
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
                            origin: (table_oid, i as u16 + 1),
                        })
                        .collect(),
                };
                let table = TableDef { alias: alias.map(|a| a.aliasname.clone()), ..table };
                Ok((Plan::Scan(Box::new(table), None), scope))
            }
            Some(NodeEnum::JoinExpr(join)) => self.plan_join(join),
            Some(NodeEnum::RangeSubselect(subselect)) => self.plan_subselect(subselect),
            Some(NodeEnum::RangeFunction(function)) => self.plan_range_function(function),
            Some(NodeEnum::RangeTableFunc(function)) => self.plan_xml_table(function),
            Some(NodeEnum::JsonTable(table)) => self.plan_json_table(table),
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
                origin: c.origin,
            })
            .collect();
        let plan = cte.plan.unwrap_or(Plan::WorkTable(cte.work_table, cte.columns.len()));
        (plan, Scope { columns })
    }

    /// plan_view plans a view's query in place of the view, naming its columns after the view's column names.
    fn plan_view(&mut self, fragment: &str, relation: &pg_query::protobuf::RangeVar) -> Result<(Plan, Scope)> {
        let (select, aliases) = crate::views::view_query(fragment)?;
        let as_of = std::mem::take(&mut self.ctx.session.as_of);
        let query = Planner { ctx: self.ctx, outer: Vec::new() }.plan_query(&select);
        self.ctx.session.as_of = as_of;
        let query = query?;
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
                origin: c.origin,
            })
            .collect();
        Ok((query.plan, Scope { columns }))
    }

    /// plan_catalog plans a scan of a system catalog relation, whose table columns come from the catalog's OID and their
    /// attribute numbers, as Postgres describes them.
    fn plan_catalog(
        &mut self,
        catalog: &'static crate::pgcatalog::CatalogTable,
        relation: &pg_query::protobuf::RangeVar,
    ) -> (Plan, Scope) {
        let alias = relation.alias.as_ref();
        let name = alias.map_or(relation.relname.clone(), |a| a.aliasname.clone());
        let renames: Vec<&str> = alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
        let columns = catalog
            .columns
            .iter()
            .enumerate()
            .map(|(i, column)| ScopeColumn {
                table: name.clone(),
                name: renames.get(i).map_or(column.name.to_string(), |r| r.to_string()),
                ty: typ(column.type_oid),
                hidden: false,
                origin: match crate::pgcatalog::builtin_view_definition(catalog.oid) {
                    Some(_) => (0, 0),
                    None => (catalog.oid, i as u16 + 1),
                },
            })
            .collect();
        (Plan::Catalog(catalog), Scope { columns })
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
                name: renames.get(i).map_or(column, |r| r.to_string()),
                ty,
                hidden: false,
                origin: (0, 0),
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
            let (what, example) = match select.values_lists.is_empty() {
                true => ("subquery", "SELECT"),
                false => ("VALUES", "VALUES"),
            };
            return Err(PgError {
                hint: Some(format!("For example, FROM ({example} ...) [AS] foo.")),
                position: opening_paren(&self.ctx.session.source, first_location(select)).and_then(position),
                ..PgError::new(code::SYNTAX_ERROR, format!("{what} in FROM must have an alias"))
            });
        };
        let query = Planner { ctx: self.ctx, outer: self.outer.clone() }.plan_query(select)?;
        let renames: Vec<&str> = alias.colnames.iter().filter_map(node_name).collect();
        if renames.len() > query.columns.len() {
            return Err(PgError::new(
                code::INVALID_COLUMN_REFERENCE,
                format!(
                    "table \"{}\" has {} columns available but {} columns specified",
                    alias.aliasname,
                    query.columns.len(),
                    renames.len()
                ),
            ));
        }
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
                    origin: c.origin,
                })
                .collect(),
        };
        Ok((query.plan, scope))
    }

    /// plan_range_function plans a set-returning function in FROM, where pg_show_all_settings() reads pg_settings, the
    /// view Postgres defines over it.
    fn plan_range_function(&mut self, function: &RangeFunction) -> Result<(Plan, Scope)> {
        if let [item] = function.functions.as_slice()
            && let Some(NodeEnum::List(list)) = item.node.as_ref()
            && let Some(node) = list.items.first()
            && matches!(node.node, Some(NodeEnum::SqlvalueFunction(_)))
        {
            let (expr, ty) = self.binder(Scope::default()).bind(node)?;
            let alias = function.alias.as_ref();
            let figured = crate::expr::figure_name(node);
            let table = alias.map_or(figured.clone(), |a| a.aliasname.clone());
            let name = alias
                .and_then(|a| a.colnames.first().and_then(node_name).map(str::to_string).or(Some(a.aliasname.clone())))
                .unwrap_or(figured);
            let column = ScopeColumn { table, name, ty, hidden: false, origin: (0, 0) };
            return Ok((Plan::Values(vec![vec![expr]]), Scope { columns: vec![column] }));
        }
        let calls = rows_from_calls(function)?;
        if calls.len() > 1 {
            return self.plan_rows_from(function, &calls);
        }
        let call = &calls[0];
        let name = call.funcname.iter().filter_map(node_name).next_back().unwrap_or_default().to_string();
        if name == "dolt_preview_merge_conflicts" {
            for arg in &call.args {
                if matches!(self.binder(Scope::default()).bind(arg)?, (Expr::Const(Value::Null), _)) {
                    return Err(crate::dolt::conflicts::null_argument(&name));
                }
            }
            let args = call.args.iter().map(|arg| self.ctx.constant_text(arg)).collect::<Result<Vec<_>>>()?;
            let table = crate::dolt::conflicts::preview_function(self.ctx, &args)?;
            let relation =
                pg_query::protobuf::RangeVar { relname: name, alias: function.alias.clone(), ..Default::default() };
            return Ok(self.plan_system(crate::dolt::tables::SystemTable::Artifacts(Box::new(table)), &relation));
        }
        if name == "pg_show_all_settings"
            && call.args.is_empty()
            && let Some(catalog) = self.ctx.catalog_relation("pg_catalog", "pg_settings")?
        {
            let relation =
                pg_query::protobuf::RangeVar { relname: name, alias: function.alias.clone(), ..Default::default() };
            return Ok(self.plan_catalog(catalog, &relation));
        }
        if name == "dolt_query_diff" {
            let args = call.args.iter().map(|arg| self.ctx.constant_text(arg)).collect::<Result<Vec<_>>>()?;
            let (diff, diff_columns) = self.ctx.plan_query_diff(&args)?;
            let alias = function.alias.as_ref();
            let table = alias.map_or(name.clone(), |a| a.aliasname.clone());
            let renames: Vec<&str> =
                alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
            let mut columns: Vec<ScopeColumn> = diff_columns
                .into_iter()
                .enumerate()
                .map(|(i, (n, ty))| ScopeColumn {
                    table: table.clone(),
                    name: renames.get(i).map_or(n, |r| r.to_string()),
                    ty,
                    hidden: false,
                    origin: (0, 0),
                })
                .collect();
            if function.ordinality {
                let name = renames.get(columns.len()).map_or("ordinality".to_string(), |r| r.to_string());
                columns.push(ScopeColumn { table, name, ty: typ(oid::INT8), hidden: false, origin: (0, 0) });
            }
            return Ok((Plan::QueryDiff(Box::new(diff), function.ordinality), Scope { columns }));
        }
        if name == "dolt_diff" {
            let args = call.args.iter().map(|arg| self.ctx.constant_text(arg)).collect::<Result<Vec<_>>>()?;
            let mut table = crate::dolt::diff::diff_function(self.ctx, &args)?;
            table.ordinality = function.ordinality;
            let relation =
                pg_query::protobuf::RangeVar { relname: name, alias: function.alias.clone(), ..Default::default() };
            return Ok(self.plan_system(crate::dolt::tables::SystemTable::User(Box::new(table)), &relation));
        }
        let mut binder = self.binder(Scope::default());
        binder.set_functions = Some(Vec::new());
        let (expr, ty) = binder.bind(&Node { node: Some(NodeEnum::FuncCall(Box::new(call.clone()))) })?;
        let expr = match expr {
            Expr::SetRef(k) => binder.set_functions.take().unwrap_or_default().swap_remove(k),
            other => other,
        };
        if !matches!(expr, Expr::Func(..) | Expr::Routine(..)) {
            return Err(PgError::unsupported("this function in FROM"));
        }
        let alias = function.alias.as_ref();
        let table = alias.map_or(name.clone(), |a| a.aliasname.clone());
        let renames: Vec<&str> = alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
        // A function returning one column, or one OUT parameter, names the column after the alias when there is one.
        let column_name = renames
            .first()
            .map(|r| r.to_string())
            .unwrap_or_else(|| alias.map_or(name.clone(), |a| a.aliasname.clone()));
        let out_columns = out_columns(&expr);
        let dolt_procedure = crate::dolt::procedures::OUT_COLUMNS.iter().any(|(n, _)| *n == name);
        let composite = match (&expr, crate::usertypes::get(ty.oid).map(|t| t.kind.clone())) {
            (Expr::Func(..), Some(crate::usertypes::Kind::Composite(attributes))) => Some(attributes),
            _ => None,
        };
        let mut columns = match out_columns {
            None if let Some(attributes) = composite => attributes
                .into_iter()
                .enumerate()
                .map(|(i, (n, t))| ScopeColumn {
                    table: table.clone(),
                    name: renames.get(i).map_or(n, |r| r.to_string()),
                    ty: t,
                    hidden: false,
                    origin: (0, 0),
                })
                .collect(),
            Some(out) if out.len() == 1 && alias.is_some() && !dolt_procedure => {
                vec![ScopeColumn {
                    table: table.clone(),
                    name: column_name,
                    ty: out[0].1,
                    hidden: false,
                    origin: (0, 0),
                }]
            }
            Some(out) => out
                .iter()
                .enumerate()
                .map(|(i, (n, t))| ScopeColumn {
                    table: table.clone(),
                    name: renames.get(i).map_or(n.to_string(), |r| r.to_string()),
                    ty: *t,
                    hidden: false,
                    origin: (0, 0),
                })
                .collect(),
            _ if ty.oid == oid::RECORD && function.coldeflist.is_empty() => {
                return Err(PgError {
                    position: crate::expr::position(call.location),
                    ..PgError::new(
                        code::SYNTAX_ERROR,
                        "a column definition list is required for functions returning \"record\"",
                    )
                });
            }
            _ => vec![ScopeColumn { table: table.clone(), name: column_name, ty, hidden: false, origin: (0, 0) }],
        };
        if !function.coldeflist.is_empty() {
            if ty.oid != oid::RECORD {
                let location = match function.coldeflist[0].node.as_ref() {
                    Some(NodeEnum::ColumnDef(definition)) => definition.location,
                    _ => -1,
                };
                return Err(PgError {
                    position: crate::expr::position(location),
                    ..PgError::new(
                        code::SYNTAX_ERROR,
                        "a column definition list is only allowed for functions returning \"record\"",
                    )
                });
            }
            columns = Vec::with_capacity(function.coldeflist.len());
            for node in &function.coldeflist {
                let Some(NodeEnum::ColumnDef(definition)) = node.node.as_ref() else { continue };
                let type_name =
                    definition.type_name.as_ref().ok_or_else(|| PgError::internal("a column without a type"))?;
                let ty = crate::expr::resolve_type_name(type_name)?;
                columns.push(ScopeColumn {
                    table: table.clone(),
                    name: definition.colname.clone(),
                    ty,
                    hidden: false,
                    origin: (0, 0),
                });
            }
        }
        let width = columns.len();
        if function.ordinality {
            let name = renames.get(1).map_or("ordinality".to_string(), |r| r.to_string());
            columns.push(ScopeColumn { table, name, ty: typ(oid::INT8), hidden: false, origin: (0, 0) });
        }
        let defined =
            (!function.coldeflist.is_empty()).then(|| columns.iter().map(|c| (c.name.clone(), c.ty)).collect());
        Ok((Plan::Function { call: expr, ordinality: function.ordinality, width, defined }, Scope { columns }))
    }

    /// plan_rows_from plans several set-returning calls in FROM, each giving one column.
    fn plan_rows_from(
        &mut self,
        function: &RangeFunction,
        calls: &[pg_query::protobuf::FuncCall],
    ) -> Result<(Plan, Scope)> {
        let alias = function.alias.as_ref();
        let renames: Vec<&str> = alias.map(|a| a.colnames.iter().filter_map(node_name).collect()).unwrap_or_default();
        let table = alias.map_or_else(|| "rows".to_string(), |a| a.aliasname.clone());
        let mut exprs = Vec::with_capacity(calls.len());
        let mut columns = Vec::with_capacity(calls.len() + 1);
        for (i, call) in calls.iter().enumerate() {
            let name = call.funcname.iter().filter_map(node_name).next_back().unwrap_or_default().to_string();
            let mut binder = self.binder(Scope::default());
            binder.set_functions = Some(Vec::new());
            let (expr, ty) = binder.bind(&Node { node: Some(NodeEnum::FuncCall(Box::new(call.clone()))) })?;
            let expr = match expr {
                Expr::SetRef(k) => binder.set_functions.take().unwrap_or_default().swap_remove(k),
                other => other,
            };
            exprs.push(expr);
            let name = renames.get(i).map_or(name, |r| r.to_string());
            columns.push(ScopeColumn { table: table.clone(), name, ty, hidden: false, origin: (0, 0) });
        }
        if function.ordinality {
            let name = renames.get(calls.len()).map_or("ordinality".to_string(), |r| r.to_string());
            columns.push(ScopeColumn { table, name, ty: typ(oid::INT8), hidden: false, origin: (0, 0) });
        }
        Ok((Plan::RowsFrom { calls: exprs, ordinality: function.ordinality }, Scope { columns }))
    }

    /// plan_join plans a join with its condition, merging the columns that USING or NATURAL name.
    fn plan_join(&mut self, join: &JoinExpr) -> Result<(Plan, Scope)> {
        let left = join.larg.as_deref().ok_or_else(|| PgError::internal("a join without a left side"))?;
        let right = join.rarg.as_deref().ok_or_else(|| PgError::internal("a join without a right side"))?;
        let (left_plan, left_scope) = self.plan_from_item(left)?;
        let lateral = is_lateral(right);
        let (right_plan, right_scope) = self.plan_lateral_item(right, &left_scope, lateral)?;
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
            let plan = Plan::Join {
                left: Box::new(left_plan),
                right: Box::new(right_plan),
                kind,
                condition,
                lateral,
                method: JoinMethod::Unplanned,
            };
            // The merged columns come first, and the joined columns they replace stay reachable only by table name.
            let mut exprs: Vec<Expr> = merged.iter().map(|m| m.1.clone()).collect();
            let mut columns: Vec<ScopeColumn> = merged
                .iter()
                .map(|m| ScopeColumn {
                    table: String::new(),
                    name: m.0.clone(),
                    ty: m.2,
                    hidden: false,
                    origin: (0, 0),
                })
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
                    Some(crate::expr::condition(binder.bind(quals)?, "JOIN/ON", crate::expr::arg_location(quals))?)
                }
                None => None,
            }
        };
        let plan = Plan::Join {
            left: Box::new(left_plan),
            right: Box::new(right_plan),
            kind,
            condition,
            lateral,
            method: JoinMethod::Unplanned,
        };
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
            let predicate = crate::expr::condition(binder.bind(node)?, "WHERE", crate::expr::arg_location(node))?;
            let mut kept = Vec::new();
            let mut existences = Vec::new();
            for c in crate::indexscan::conjuncts(&predicate) {
                match matches!(c, Expr::Exists(_))
                    || matches!(c, Expr::Not(inner) if matches!(**inner, Expr::Exists(_)))
                    || crate::joins::any_input(c).is_some()
                {
                    true => existences.push(c.clone()),
                    false => kept.push(c.clone()),
                }
            }
            if crate::optimizer::enabled() && crate::optimizer::plannable(&plan) {
                plan = crate::optimizer::query_planner(self.ctx, plan, kept);
            } else {
                if let Some(kept) = kept.into_iter().reduce(|a, b| Expr::And(Box::new(a), Box::new(b))) {
                    plan = push_down(plan, kept);
                }
                plan = self.use_indexes(plan);
            }
            for existence in existences {
                plan = crate::joins::filter_existence(plan, existence);
            }
        } else if crate::optimizer::enabled() && crate::optimizer::plannable(&plan) {
            plan = crate::optimizer::query_planner(self.ctx, plan, Vec::new());
        }
        let hints = crate::joins::hints(&self.ctx.session.source);
        if !hints.is_empty() {
            plan = crate::joins::apply_hints(self.ctx, plan, &scope, &hints);
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
                if table.is_none() && select.from_clause.is_empty() {
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
        let origins: Vec<(u32, u16)> = targets
            .iter()
            .map(|t| match t.0 {
                Expr::Column(i) => scope.columns.get(i).map_or((0, 0), |c| c.origin),
                _ => (0, 0),
            })
            .collect();
        // ORDER BY keys may name output columns, refer to them by position, or be expressions of the input.
        let mut sorts: Vec<(Expr, bool, bool)> = Vec::new();
        for sort in &select.sort_clause {
            let Some(NodeEnum::SortBy(sort)) = sort.node.as_ref() else { continue };
            let node = sort.node.as_deref().ok_or_else(|| PgError::internal("ORDER BY without a key"))?;
            let (descending, nulls_first) = sort_order(sort);
            let (expr, ty) = if let Some((n, location)) = ordinal(node) {
                output_ordinal_named(&names, n, location)?;
                (targets[n - 1].0.clone(), targets[n - 1].1)
            } else if let Some(NodeEnum::ColumnRef(c)) = node.node.as_ref()
                && c.fields.len() == 1
                && let Some(name) = node_name(&c.fields[0])
                && names.iter().any(|n| n == name)
            {
                let mut matching = names.iter().enumerate().filter(|(_, n)| *n == name).map(|(i, _)| &targets[i]);
                let target = matching.next().expect("a target has the name");
                if matching.any(|other| other.0 != target.0) {
                    return Err(PgError {
                        position: position(c.location),
                        ..PgError::new(code::AMBIGUOUS_COLUMN, format!("ORDER BY \"{name}\" is ambiguous"))
                    });
                }
                (target.0.clone(), target.1)
            } else {
                binder.bind(node)?
            };
            if matches!(ty.oid, oid::XID | oid::CID | oid::XML) {
                return Err(PgError {
                    position: position(crate::expr::arg_location(node)),
                    hint: Some("Use an explicit ordering operator or modify the query.".into()),
                    ..PgError::new(
                        code::UNDEFINED_FUNCTION,
                        format!(
                            "could not identify an ordering operator for type {}",
                            crate::cast::type_display(ty.oid)
                        ),
                    )
                });
            }
            sorts.push((expr, descending, nulls_first));
        }
        let windows = binder.windows.take();
        let set_functions = binder.set_functions.take().unwrap_or_default();
        let mut having = None;
        if let Some(node) = select.having_clause.as_deref() {
            binder.clause = "HAVING";
            having = Some(crate::expr::condition(binder.bind(node)?, "HAVING", crate::expr::arg_location(node))?);
        }
        let aggregates = binder.aggregates.take();
        let columns_bound = std::mem::take(&mut binder.columns);
        drop(binder);
        let mut windows = windows;
        let mut set_functions = set_functions;
        if let Some(aggregates) = aggregates {
            // Group keys may refer to output columns by position or name, or be expressions of the input.
            let mut node_sets: Vec<Vec<&Node>> = vec![Vec::new()];
            for item in &select.group_clause {
                let expanded = grouping_sets(item);
                node_sets = node_sets
                    .iter()
                    .flat_map(|prefix| expanded.iter().map(|set| [&prefix[..], set].concat()))
                    .collect();
            }
            let mut groups: Vec<Expr> = Vec::new();
            let mut sets: Vec<Vec<usize>> = Vec::with_capacity(node_sets.len());
            let mut binder = self.binder(scope.clone());
            binder.clause = "GROUP BY";
            for node_set in &node_sets {
                let mut set: Vec<usize> = Vec::with_capacity(node_set.len());
                for node in node_set {
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
                    let expr = mark_input(expr);
                    let i = groups.iter().position(|g| *g == expr).unwrap_or_else(|| {
                        groups.push(expr);
                        groups.len() - 1
                    });
                    if !set.contains(&i) {
                        set.push(i);
                    }
                }
                sets.push(set);
            }
            drop(binder);
            if select.group_distinct {
                let mut seen = HashSet::new();
                sets.retain(|set| {
                    let mut members = set.clone();
                    members.sort_unstable();
                    seen.insert(members)
                });
            }
            let uses_sets = select.group_clause.iter().any(|n| matches!(n.node, Some(NodeEnum::GroupingSet(_))));
            let sets = uses_sets.then_some(sets);
            let mask = sets.as_ref().map(|_| groups.len() + aggregates.len());
            let finish = |expr: Expr| -> Result<Expr> {
                let expr = place_grouping(replace_groups(mark_input(expr), &groups), groups.len(), mask)?;
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
            plan = Plan::Aggregate { input: Box::new(plan), groups: groups_plan.collect(), aggregates, sets };
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
            let mut levels: Vec<usize> = Vec::with_capacity(set_functions.len());
            for function in &set_functions {
                let mut level = 0;
                function.visit(&mut |e| {
                    if let Expr::SetRef(j) = e {
                        level = level.max(levels.get(*j).map_or(0, |l| l + 1));
                    }
                });
                levels.push(level);
            }
            let mut columns = vec![0; set_functions.len()];
            for level in 0..=levels.iter().copied().max().unwrap_or(0) {
                let input_width = plan.width();
                let mut functions = Vec::new();
                for (k, function) in set_functions.iter().enumerate().filter(|(k, _)| levels[*k] == level) {
                    columns[k] = input_width + functions.len();
                    functions.push(replace_set_functions(function.clone(), &columns));
                }
                plan = Plan::ProjectSet { input: Box::new(plan), functions, dropped: Vec::new() };
            }
            let place = |expr: Expr| replace_set_functions(expr, &columns);
            targets = targets.into_iter().map(|(e, t, n, l)| (place(e), t, n, l)).collect();
            sorts = sorts.into_iter().map(|(e, d, n)| (place(e), d, n)).collect();
        }
        let width = targets.len();
        let types: Vec<ColumnType> = targets.iter().map(|t| t.1).collect();
        let columns: Vec<Column> =
            targets.iter().zip(&origins).map(|(t, &origin)| Column { origin, ..column(t.2.clone(), t.1) }).collect();
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
        let mut on_locations = Vec::new();
        if !distinct_on.is_empty() {
            let mut binder = self.binder(scope.clone());
            let mut on = Vec::new();
            for node in distinct_on {
                on_locations.push(crate::expr::arg_location(node));
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
            let mismatch = |location: i32| PgError {
                position: position(location),
                ..PgError::new(
                    code::INVALID_COLUMN_REFERENCE,
                    "SELECT DISTINCT ON expressions must match initial ORDER BY expressions",
                )
            };
            let location_of = |expr: &Expr| on.iter().position(|o| o == expr).map_or(-1, |i| on_locations[i]);
            let mut sort_keys: Vec<SortKey> = Vec::new();
            let mut skipped = false;
            for key in &keys {
                if !on.contains(&key.expr) {
                    skipped = true;
                } else if skipped {
                    return Err(mismatch(location_of(&key.expr)));
                } else if !sort_keys.iter().any(|k| k.expr == key.expr) {
                    sort_keys.push(key.clone());
                }
            }
            let prefix = sort_keys.len();
            for (key, &location) in on.iter().zip(&on_locations) {
                if sort_keys.iter().any(|k| k.expr == *key) {
                    continue;
                }
                if skipped {
                    return Err(mismatch(location));
                }
                sort_keys.push(SortKey { expr: key.clone(), descending: false, nulls_first: false });
            }
            sort_keys.extend(keys.iter().skip(prefix).cloned());
            plan = Plan::Sort { input: Box::new(plan), keys: sort_keys };
            plan = Plan::Distinct { input: Box::new(plan), keys: distinct_keys };
        } else {
            if distinct {
                plan = Plan::Distinct { input: Box::new(plan), keys: None };
            }
            if !keys.is_empty() {
                plan = crate::indexscan::order_by_index(Plan::Sort { input: Box::new(plan), keys });
            }
        }
        plan = crate::indexscan::nearest(self.limit(plan, select)?);
        if !matches!(&plan, Plan::Project { exprs, .. } if exprs.len() == width) {
            let visible = (0..width).map(Expr::Column).collect();
            plan = Plan::Project { input: Box::new(plan), exprs: visible };
        }
        plan = crate::joins::plan_joins(self.ctx, plan);
        crate::indexscan::prune(&mut plan);
        Ok(Query { plan, columns, types })
    }
}

/// first_location returns the location of a SELECT's first value or target, or -1 without one.
fn first_location(select: &SelectStmt) -> i32 {
    let first = match select.values_lists.first().and_then(|l| l.node.as_ref()) {
        Some(NodeEnum::List(list)) => list.items.first().map(crate::expr::arg_location),
        _ => select.target_list.first().and_then(|t| match t.node.as_ref() {
            Some(NodeEnum::ResTarget(target)) => Some(target.location),
            _ => None,
        }),
    };
    first.unwrap_or(-1)
}

/// opening_paren returns the location of the parenthesis that opens a subquery whose first value or target is at a
/// location, the leftmost one before it with only keywords, spaces, and parentheses between.
fn opening_paren(source: &str, location: i32) -> Option<i32> {
    let end = usize::try_from(location).ok().filter(|&l| l <= source.len())?;
    let mut found = None;
    for (i, c) in source[..end].char_indices().rev() {
        match c {
            '(' => found = Some(i as i32),
            c if c.is_alphabetic() || c.is_whitespace() => {}
            _ => break,
        }
    }
    found
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
pub(crate) fn compare_sorted(keys: &[SortKey], a: &[Value], b: &[Value]) -> Ordering {
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

/// replace_set_functions replaces set-returning call references with the columns that ProjectSet nodes append for
/// them.
fn replace_set_functions(expr: Expr, columns: &[usize]) -> Expr {
    match expr {
        Expr::SetRef(k) => Expr::Column(columns[k]),
        other => other.map_children(&mut |child| replace_set_functions(child, columns)),
    }
}

/// replace_windows replaces window call references with the columns that a window node appends after its input's.
fn replace_windows(expr: Expr, input_width: usize) -> Expr {
    match expr {
        Expr::WindowRef(k) => Expr::Column(input_width + k),
        other => other.map_children(&mut |child| replace_windows(child, input_width)),
    }
}

/// target_expression returns the expression of a simple SELECT's output column, by position.
fn target_expression(select: &SelectStmt, i: usize) -> Result<Node> {
    let unsupported = || PgError::unsupported("CYCLE and SEARCH clauses over this query");
    if SetOperation::try_from(select.op).unwrap_or(SetOperation::SetopNone) != SetOperation::SetopNone {
        return Err(unsupported());
    }
    match select.target_list.get(i).and_then(|t| t.node.as_ref()) {
        Some(NodeEnum::ResTarget(t)) => {
            let value = t.val.as_deref().ok_or_else(unsupported)?;
            if matches!(value.node.as_ref(), Some(NodeEnum::ColumnRef(c)) if c.fields.iter().any(|f| matches!(f.node, Some(NodeEnum::AStar(_)))))
            {
                return Err(unsupported());
            }
            Ok(value.clone())
        }
        _ => Err(unsupported()),
    }
}

/// worktable_name returns the name that a recursive term uses for its WITH query, its alias when it has one.
fn worktable_name(select: &SelectStmt, cte: &str) -> String {
    for (node, ..) in NodeEnum::SelectStmt(Box::new(select.clone())).nodes() {
        if let pg_query::NodeRef::RangeVar(r) = node
            && r.schemaname.is_empty()
            && r.relname == cte
        {
            return r.alias.as_ref().map_or(r.relname.clone(), |a| a.aliasname.clone());
        }
    }
    cte.to_string()
}

/// add_target appends an output column to a SELECT from the SQL text of its expression.
fn add_target(select: &mut SelectStmt, name: &str, text: &str) -> Result<()> {
    let value = crate::dml::parse_expression(text)?;
    select.target_list.push(Node {
        node: Some(NodeEnum::ResTarget(Box::new(pg_query::protobuf::ResTarget {
            name: name.to_string(),
            val: Some(Box::new(value)),
            location: -1,
            ..Default::default()
        }))),
    });
    Ok(())
}

/// clause_columns returns the SQL text of the expressions of a SEARCH or CYCLE clause's columns in both terms.
fn clause_columns(
    names: &[String],
    left: &SelectStmt,
    right: &SelectStmt,
    list: &[Node],
    what: &str,
) -> Result<(String, String)> {
    use crate::ddl::expression_text;
    let mut left_parts = Vec::new();
    let mut right_parts = Vec::new();
    for column in list.iter().filter_map(node_name) {
        let i = names.iter().position(|n| n == column).ok_or_else(|| {
            PgError::new(code::UNDEFINED_COLUMN, format!("{what} column \"{column}\" not in WITH query column list"))
        })?;
        left_parts.push(expression_text(&target_expression(left, i)?)?);
        right_parts.push(expression_text(&target_expression(right, i)?)?);
    }
    Ok((left_parts.join(", "), right_parts.join(", ")))
}

/// rewrite_search_and_cycle adds the columns of a recursive query's SEARCH and CYCLE clauses to both of its terms,
/// as Postgres' rewriteSearchAndCycle does, so that the recursive term stops at cycles.
fn rewrite_search_and_cycle(
    cte: &pg_query::protobuf::CommonTableExpr,
    query: &SelectStmt,
    aliases: &mut Vec<String>,
) -> Result<SelectStmt> {
    use crate::ddl::expression_text;
    use crate::engine::quote_identifier as q;
    let (Some(left), Some(right)) = (query.larg.as_deref(), query.rarg.as_deref()) else {
        return Err(PgError::internal("a recursive query without both terms"));
    };
    let (mut left, mut right) = (left.clone(), right.clone());
    if let Some(Some(NodeEnum::List(first))) = left.values_lists.first().map(|v| v.node.as_ref()) {
        let columns: Vec<String> = (1..=first.items.len()).map(|i| format!("column{i}")).collect();
        let shown: Vec<String> =
            columns.iter().enumerate().map(|(i, c)| format!("{c} AS {}", q(aliases.get(i).unwrap_or(c)))).collect();
        let values = pg_query::NodeRef::SelectStmt(&left).deparse().map_err(PgError::internal)?;
        let text = format!("SELECT {} FROM ({values}) AS anchor({})", shown.join(", "), columns.join(", "));
        let parsed = pg_query::parse(&text).map_err(PgError::internal)?;
        let Some(NodeEnum::SelectStmt(select)) =
            parsed.protobuf.stmts.into_iter().next().and_then(|s| s.stmt).and_then(|s| s.node)
        else {
            return Err(PgError::internal("an anchor of VALUES"));
        };
        left = *select;
    }
    let names: Vec<String> = if aliases.is_empty() {
        left.target_list
            .iter()
            .map(|t| match t.node.as_ref() {
                Some(NodeEnum::ResTarget(t)) if !t.name.is_empty() => t.name.clone(),
                Some(NodeEnum::ResTarget(t)) => t.val.as_deref().map(figure_name).unwrap_or_default(),
                _ => String::new(),
            })
            .collect()
    } else {
        aliases.clone()
    };
    let w = q(&worktable_name(&right, &cte.ctename));
    let mut added = Vec::new();
    if let Some(search) = cte.search_clause.as_ref() {
        if search.search_breadth_first {
            return Err(PgError::unsupported("SEARCH BREADTH FIRST"));
        }
        let (l, r) = clause_columns(&names, &left, &right, &search.search_col_list, "search")?;
        let seq = &search.search_seq_column;
        add_target(&mut left, seq, &format!("ARRAY[ROW({l})]"))?;
        add_target(&mut right, seq, &format!("{w}.{} || ROW({r})", q(seq)))?;
        added.push(seq.clone());
    }
    if let Some(cycle) = cte.cycle_clause.as_ref() {
        let (l, r) = clause_columns(&names, &left, &right, &cycle.cycle_col_list, "cycle")?;
        let mark_value = match cycle.cycle_mark_value.as_deref() {
            Some(node) => expression_text(node)?,
            None => "true".into(),
        };
        let mark_default = match cycle.cycle_mark_default.as_deref() {
            Some(node) => expression_text(node)?,
            None => "false".into(),
        };
        let (mark, path) = (&cycle.cycle_mark_column, &cycle.cycle_path_column);
        add_target(&mut left, mark, &mark_default)?;
        add_target(&mut left, path, &format!("ARRAY[ROW({l})]"))?;
        add_target(
            &mut right,
            mark,
            &format!("CASE WHEN ROW({r}) = ANY({w}.{}) THEN {mark_value} ELSE {mark_default} END", q(path)),
        )?;
        add_target(&mut right, path, &format!("{w}.{} || ROW({r})", q(path)))?;
        let condition = format!("{w}.{} <> {mark_value}", q(mark));
        let condition = match right.where_clause.as_deref() {
            Some(existing) => format!("({}) AND ({condition})", expression_text(existing)?),
            None => condition,
        };
        right.where_clause = Some(Box::new(crate::dml::parse_expression(&condition)?));
        added.push(mark.clone());
        added.push(path.clone());
    }
    if !aliases.is_empty() {
        aliases.extend(added);
    }
    Ok(SelectStmt { larg: Some(Box::new(left)), rarg: Some(Box::new(right)), ..query.clone() })
}

/// references reports whether a query refers to a relation by an unqualified name.
fn references(select: &SelectStmt, name: &str) -> bool {
    NodeEnum::SelectStmt(Box::new(select.clone()))
        .nodes()
        .into_iter()
        .any(|(n, ..)| matches!(n, pg_query::NodeRef::RangeVar(r) if r.schemaname.is_empty() && r.relname == name))
}

/// check_recursion fails as Postgres' checkWellFormedRecursion does for a recursive WITH query it cannot run: one that is
/// not a UNION, one whose reference to itself sits somewhere other than once in its recursive term outside subqueries,
/// outer joins, INTERSECT, and EXCEPT, and one with ORDER BY, LIMIT, OFFSET, or FOR UPDATE atop the UNION.
fn check_recursion(cte: &pg_query::protobuf::CommonTableExpr, query: &SelectStmt) -> Result<()> {
    let name = &cte.ctename;
    let (Some(left), Some(right), Ok(SetOperation::SetopUnion)) =
        (query.larg.as_deref(), query.rarg.as_deref(), SetOperation::try_from(query.op))
    else {
        return Err(PgError {
            position: position(cte.location),
            ..PgError::new(
                code::INVALID_RECURSION,
                format!(
                    "recursive query \"{name}\" does not have the form non-recursive-term UNION [ALL] recursive-term"
                ),
            )
        });
    };
    let mut check = RecursionCheck { name, context: Recursion::NonRecursiveTerm, references: 0 };
    check.select(left)?;
    check.context = Recursion::Ok;
    check.select(right)?;
    let unsupported = |what: &str, at: i32| PgError {
        position: position(at),
        ..PgError::new(code::FEATURE_NOT_SUPPORTED, format!("{what} in a recursive query is not implemented"))
    };
    if let Some(sort) = query.sort_clause.first() {
        let at = match sort.node.as_ref() {
            Some(NodeEnum::SortBy(s)) => s.node.as_deref().map_or(-1, crate::expr::arg_location),
            _ => -1,
        };
        return Err(unsupported("ORDER BY", at));
    }
    if let Some(offset) = query.limit_offset.as_deref() {
        return Err(unsupported("OFFSET", crate::expr::arg_location(offset)));
    }
    if let Some(limit) = query.limit_count.as_deref() {
        return Err(unsupported("LIMIT", crate::expr::arg_location(limit)));
    }
    if !query.locking_clause.is_empty() {
        return Err(unsupported("FOR UPDATE/SHARE", -1));
    }
    Ok(())
}

/// mutually_recursive reports whether a query of a WITH clause reaches itself through another of the clause's queries.
fn mutually_recursive(with: &pg_query::protobuf::WithClause, start: &pg_query::protobuf::CommonTableExpr) -> bool {
    let ctes: Vec<(&str, &SelectStmt)> = with
        .ctes
        .iter()
        .filter_map(|c| match c.node.as_ref() {
            Some(NodeEnum::CommonTableExpr(c)) => match c.ctequery.as_deref().and_then(|q| q.node.as_ref()) {
                Some(NodeEnum::SelectStmt(s)) => Some((c.ctename.as_str(), &**s)),
                _ => None,
            },
            _ => None,
        })
        .collect();
    let mut reached: Vec<&str> = Vec::new();
    let mut pending = vec![start.ctename.as_str()];
    while let Some(name) = pending.pop() {
        let Some((_, query)) = ctes.iter().find(|(n, _)| *n == name) else { continue };
        for (other, _) in ctes.iter().filter(|(n, _)| *n != name && references(query, n)) {
            if *other == start.ctename {
                return true;
            }
            if !reached.contains(other) {
                reached.push(other);
                pending.push(other);
            }
        }
    }
    false
}

/// Recursion is where a recursive WITH query's reference to itself sits, as Postgres' RecursionContext names it.
#[derive(Clone, Copy, PartialEq)]
enum Recursion {
    Ok,
    NonRecursiveTerm,
    Sublink,
    OuterJoin,
    Intersect,
    Except,
}

/// RecursionCheck walks the terms of a recursive WITH query, counting its references to itself.
struct RecursionCheck<'n> {
    name: &'n str,
    context: Recursion,
    references: usize,
}

impl RecursionCheck<'_> {
    /// reference checks a reference to the query at a location.
    fn reference(&mut self, location: i32) -> Result<()> {
        let what = match self.context {
            Recursion::NonRecursiveTerm => "must not appear within its non-recursive term",
            Recursion::Sublink => "must not appear within a subquery",
            Recursion::OuterJoin => "must not appear within an outer join",
            Recursion::Intersect => "must not appear within INTERSECT",
            Recursion::Except => "must not appear within EXCEPT",
            Recursion::Ok if self.references > 0 => "must not appear more than once",
            Recursion::Ok => {
                self.references += 1;
                return Ok(());
            }
        };
        Err(PgError {
            position: position(location),
            ..PgError::new(code::INVALID_RECURSION, format!("recursive reference to query \"{}\" {what}", self.name))
        })
    }

    /// select walks a SELECT or set operation, skipping one whose own WITH clause hides the query's name.
    fn select(&mut self, select: &SelectStmt) -> Result<()> {
        let ctes = select.with_clause.iter().flat_map(|w| &w.ctes).filter_map(|c| match c.node.as_ref() {
            Some(NodeEnum::CommonTableExpr(c)) => Some(c),
            _ => None,
        });
        if ctes.clone().any(|c| c.ctename == self.name) {
            return Ok(());
        }
        for cte in ctes {
            if let Some(query) = cte.ctequery.as_deref() {
                self.expression(query)?;
            }
        }
        let saved = self.context;
        match SetOperation::try_from(select.op).unwrap_or(SetOperation::SetopNone) {
            SetOperation::SetopNone | SetOperation::Undefined => {
                let before = self.references;
                for item in &select.from_clause {
                    self.from(item)?;
                }
                let having = select.having_clause.as_deref();
                let aggregate = select.target_list.iter().chain(having).find(|n| has_aggregate(n));
                if self.references > before
                    && let Some(node) = aggregate
                {
                    return Err(PgError {
                        position: match node.node.as_ref() {
                            Some(NodeEnum::ResTarget(t)) => t.val.as_deref().and_then(aggregate_location),
                            _ => aggregate_location(node),
                        }
                        .and_then(position),
                        ..PgError::new(
                            code::INVALID_RECURSION,
                            "aggregate functions are not allowed in a recursive query's recursive term",
                        )
                    });
                }
                let lists = [&select.target_list, &select.group_clause, &select.values_lists, &select.window_clause];
                for node in lists.into_iter().flatten() {
                    self.expression(node)?;
                }
                for node in [&select.where_clause, &select.having_clause].into_iter().flatten() {
                    self.expression(node)?;
                }
            }
            op => {
                let (left, right) = match op {
                    SetOperation::SetopIntersect if select.all => (Recursion::Intersect, Recursion::Intersect),
                    SetOperation::SetopExcept if select.all => (Recursion::Except, Recursion::Except),
                    SetOperation::SetopExcept => (saved, Recursion::Except),
                    _ => (saved, saved),
                };
                self.context = left;
                if let Some(larg) = select.larg.as_deref() {
                    self.select(larg)?;
                }
                self.context = right;
                if let Some(rarg) = select.rarg.as_deref() {
                    self.select(rarg)?;
                }
                self.context = saved;
            }
        }
        let limits = [&select.limit_offset, &select.limit_count].into_iter().flatten().map(|n| &**n);
        for node in select.sort_clause.iter().chain(limits) {
            self.expression(node)?;
        }
        Ok(())
    }

    /// from walks a FROM item, where the nullable sides of outer joins are outer join contexts.
    fn from(&mut self, node: &Node) -> Result<()> {
        match node.node.as_ref() {
            Some(NodeEnum::RangeVar(r)) if r.schemaname.is_empty() && r.relname == self.name => {
                self.reference(r.location)
            }
            Some(NodeEnum::JoinExpr(join)) => {
                let saved = self.context;
                let (left, right) = match pg_query::protobuf::JoinType::try_from(join.jointype) {
                    Ok(pg_query::protobuf::JoinType::JoinLeft) => (saved, Recursion::OuterJoin),
                    Ok(pg_query::protobuf::JoinType::JoinRight) => (Recursion::OuterJoin, saved),
                    Ok(pg_query::protobuf::JoinType::JoinFull) => (Recursion::OuterJoin, Recursion::OuterJoin),
                    _ => (saved, saved),
                };
                self.context = left;
                let left_result = join.larg.as_deref().map_or(Ok(()), |l| self.from(l));
                self.context = right;
                let right_result = join.rarg.as_deref().map_or(Ok(()), |r| self.from(r));
                self.context = saved;
                left_result?;
                right_result?;
                join.quals.as_deref().map_or(Ok(()), |q| self.expression(q))
            }
            Some(NodeEnum::RangeSubselect(subselect)) => {
                match subselect.subquery.as_deref().and_then(|q| q.node.as_ref()) {
                    Some(NodeEnum::SelectStmt(select)) => self.select(select),
                    _ => Ok(()),
                }
            }
            _ => self.expression(node),
        }
    }

    /// expression walks an expression, where a reference can only sit inside a subquery.
    fn expression(&mut self, node: &Node) -> Result<()> {
        let Some(node) = node.node.as_ref() else { return Ok(()) };
        let found = node.nodes().into_iter().find_map(|(n, ..)| match n {
            pg_query::NodeRef::RangeVar(r) if r.schemaname.is_empty() && r.relname == self.name => Some(r.location),
            _ => None,
        });
        let Some(location) = found else { return Ok(()) };
        let saved = std::mem::replace(&mut self.context, Recursion::Sublink);
        let result = self.reference(location);
        self.context = saved;
        result
    }
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

/// trimmed_json returns a jsonb value with its numbers' trailing fractional zeroes removed.
pub(crate) fn trimmed_json(json: &crate::json::Json) -> crate::json::Json {
    use crate::json::Json;
    match json {
        Json::Number(n) => Json::Number(n.trimmed()),
        Json::Array(items) => Json::Array(items.iter().map(trimmed_json).collect()),
        Json::Object(fields) => Json::Object(fields.iter().map(|(k, v)| (k.clone(), trimmed_json(v))).collect()),
        other => other.clone(),
    }
}

/// limit_value evaluates a LIMIT or OFFSET, failing as Postgres does when it is negative.
pub(crate) fn limit_value(
    expr: &Option<Expr>,
    ctx: &mut Ctx<'_>,
    what: &str,
    error_code: &'static str,
) -> Result<Option<i64>> {
    let Some(expr) = expr else { return Ok(None) };
    match expr.eval(ctx, &[])? {
        Value::Int8(n) if n < 0 => Err(PgError::new(error_code, format!("{what} must not be negative"))),
        Value::Int8(n) => Ok(Some(n)),
        _ => Ok(None),
    }
}

impl Plan {
    /// map_exprs replaces each expression that the plan's nodes hold with what a function makes of it, given how many
    /// rows out from the plan's own rows the expression sits, as a lateral join's right input sees its left row one
    /// row out. It reports false when the plan holds a node whose expressions it cannot reach, which it leaves alone.
    pub(crate) fn map_exprs(&mut self, depth: usize, f: &mut dyn FnMut(Expr, usize) -> Expr) -> bool {
        let mut map = |e: &mut Expr, depth: usize| {
            let old = std::mem::replace(e, Expr::Const(Value::Null));
            *e = f(old, depth);
        };
        let mut inputs: Vec<(&mut Plan, usize)> = Vec::new();
        match self {
            Plan::OneRow | Plan::Scan(..) | Plan::Catalog(_) | Plan::CatalogIndexScan(_) | Plan::WorkTable(..) => {}
            Plan::System(_) | Plan::QueryDiff(..) | Plan::XmlTable(_) | Plan::JsonTable(_) => return false,
            Plan::IndexScan(scan) => {
                if let Some(n) = &mut scan.nearest {
                    for e in [&mut n.order, &mut n.query].into_iter().chain(&mut n.limit).chain(&mut n.offset) {
                        map(e, depth);
                    }
                }
            }
            Plan::Values(rows) => rows.iter_mut().flatten().for_each(|e| map(e, depth)),
            Plan::Function { call, .. } => map(call, depth),
            Plan::RowsFrom { calls, .. } => calls.iter_mut().for_each(|e| map(e, depth)),
            Plan::Filter { input, predicate } => {
                map(predicate, depth);
                inputs.push((input, depth));
            }
            Plan::Project { input, exprs } => {
                exprs.iter_mut().for_each(|e| map(e, depth));
                inputs.push((input, depth));
            }
            Plan::Join { left, right, condition, lateral, method, .. } => {
                condition.iter_mut().for_each(|e| map(e, depth));
                if let JoinMethod::Lookup { keys, .. } | JoinMethod::CatalogLookup { keys, .. } = method {
                    keys.iter_mut().for_each(|e| map(e, depth));
                }
                let right_depth = depth + usize::from(*lateral);
                inputs.push((left, depth));
                inputs.push((right, right_depth));
            }
            Plan::Aggregate { input, groups, aggregates, .. } => {
                groups.iter_mut().for_each(|e| map(e, depth));
                for call in aggregates {
                    let order = call.order.iter_mut().map(|(e, _, _)| e);
                    call.args.iter_mut().chain(&mut call.filter).chain(order).for_each(|e| map(e, depth));
                }
                inputs.push((input, depth));
            }
            Plan::Sort { input, keys } => {
                keys.iter_mut().for_each(|k| map(&mut k.expr, depth));
                inputs.push((input, depth));
            }
            Plan::Distinct { input, keys } => {
                keys.iter_mut().flatten().for_each(|e| map(e, depth));
                inputs.push((input, depth));
            }
            Plan::Limit { input, limit, offset } => {
                limit.iter_mut().chain(offset).for_each(|e| map(e, depth));
                inputs.push((input, depth));
            }
            Plan::SetOp { left, right, .. } => {
                inputs.push((left, depth));
                inputs.push((right, depth));
            }
            Plan::Recursive { anchor, step, .. } => {
                inputs.push((anchor, depth));
                inputs.push((step, depth));
            }
            Plan::ProjectSet { input, functions, .. } => {
                functions.iter_mut().for_each(|e| map(e, depth));
                inputs.push((input, depth));
            }
            Plan::Window { input, calls } => {
                for call in calls {
                    let order = call.order.iter_mut().map(|k| &mut k.expr);
                    let args = call.args.iter_mut().chain(&mut call.filter).chain(&mut call.partition).chain(order);
                    args.for_each(|e| map(e, depth));
                    for bound in [&mut call.start, &mut call.end] {
                        if let crate::window::Bound::Preceding(e) | crate::window::Bound::Following(e) = bound {
                            map(e, depth);
                        }
                    }
                    if let Some(range) = &mut call.range {
                        map(&mut range.key, depth);
                        range.start.iter_mut().chain(&mut range.end).for_each(|(e, _)| map(e, depth));
                    }
                }
                inputs.push((input, depth));
            }
            Plan::Once(input) => inputs.push((input, depth)),
        }
        inputs.into_iter().fold(true, |known, (input, depth)| input.map_exprs(depth, f) && known)
    }

    /// width returns the number of columns of the plan's rows.
    pub(crate) fn width(&self) -> usize {
        match self {
            Plan::OneRow => 0,
            Plan::Scan(table, _) => table.columns.len(),
            Plan::IndexScan(scan) => scan.table.columns.len(),
            Plan::Recursive { anchor, .. } => anchor.width(),
            Plan::WorkTable(_, width) => *width,
            Plan::Window { input, calls } => input.width() + calls.len(),
            Plan::ProjectSet { input, functions, .. } => input.width() + functions.len(),
            Plan::System(system) => system.columns().len(),
            Plan::Catalog(table) => table.columns.len(),
            Plan::CatalogIndexScan(scan) => scan.table.columns.len(),
            Plan::Values(rows) => rows.first().map_or(0, Vec::len),
            Plan::Function { ordinality, width, .. } => width + *ordinality as usize,
            Plan::Filter { input, .. }
            | Plan::Sort { input, .. }
            | Plan::Distinct { input, .. }
            | Plan::Limit { input, .. } => input.width(),
            Plan::Project { exprs, .. } => exprs.len(),
            Plan::Join { left, right, .. } => left.width() + right.width(),
            Plan::XmlTable(table) => table.columns.len(),
            Plan::JsonTable(table) => table.columns.len(),
            Plan::RowsFrom { calls, ordinality } => calls.len() + *ordinality as usize,
            Plan::QueryDiff(diff, ordinality) => diff.from_width + diff.to_width + 1 + *ordinality as usize,
            Plan::Aggregate { groups, aggregates, sets, .. } => {
                groups.len() + aggregates.len() + usize::from(sets.is_some())
            }
            Plan::SetOp { left, .. } => left.width(),
            Plan::Once(input) => input.width(),
        }
    }

    /// subquery_rows runs a subquery for an enclosing row.
    pub fn subquery_rows(&self, ctx: &mut Ctx<'_>, row: &[Value]) -> Result<std::sync::Arc<SubqueryRows>> {
        ctx.outer.push(row.to_vec());
        let rows = self.shared_rows(ctx);
        ctx.outer.pop();
        rows
    }

    /// subquery_exists reports whether a subquery returns a row for an enclosing row, stopping at the first row.
    pub fn subquery_exists(&self, ctx: &mut Ctx<'_>, row: &[Value]) -> Result<bool> {
        if let Plan::Once(_) = self {
            return Ok(!self.subquery_rows(ctx, row)?.rows.is_empty());
        }
        // EXISTS ignores the subquery's select list, as Postgres' planner drops it.
        let mut plan = self;
        while let Plan::Project { input, .. } = plan {
            plan = input;
        }
        ctx.outer.push(row.to_vec());
        let found = plan.open(ctx).and_then(|mut rows| rows.next(ctx));
        ctx.outer.pop();
        Ok(found?.is_some())
    }

    /// shared_rows runs the plan, reusing the rows of a `Once` plan that the running plan has already evaluated.
    pub(crate) fn shared_rows(&self, ctx: &mut Ctx<'_>) -> Result<std::sync::Arc<SubqueryRows>> {
        let key = self as *const Plan as usize;
        if let Plan::Once(_) = self
            && let Some(rows) = ctx.once.as_ref().and_then(|once| once.get(&key))
        {
            return Ok(rows.clone());
        }
        let rows = match self {
            Plan::Once(input) => input.run(ctx)?,
            plan => plan.run(ctx)?,
        };
        let rows = std::sync::Arc::new(SubqueryRows {
            rows,
            keys: std::sync::OnceLock::new(),
            index: std::sync::OnceLock::new(),
        });
        if let Plan::Once(_) = self
            && let Some(once) = ctx.once.as_mut()
        {
            once.insert(key, rows.clone());
        }
        Ok(rows)
    }

    /// run runs the plan and returns its rows.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        if ctx.once.is_some() {
            return self.open(ctx).and_then(|mut rows| crate::exec::drain(&mut *rows, ctx));
        }
        ctx.once = Some(std::collections::HashMap::new());
        let rows = self.open(ctx).and_then(|mut rows| crate::exec::drain(&mut *rows, ctx));
        ctx.once = None;
        rows
    }

    /// run_leaf computes the rows of a plan node that reads no other plan node, or of a window node.
    pub(crate) fn run_leaf(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        Ok(match self {
            Plan::OneRow => vec![Vec::new()],
            Plan::WorkTable(id, _) => ctx.work_tables.get(id).cloned().unwrap_or_default(),
            Plan::System(system) => ctx.without_temp(|ctx| system.rows(ctx))?,
            Plan::Catalog(table) => ctx.catalog_rows(table)?,
            Plan::CatalogIndexScan(scan) => scan.run(ctx)?,
            Plan::Values(rows) => {
                let mut out = Vec::with_capacity(rows.len());
                for row in rows {
                    out.push(row.iter().map(|e| e.eval(ctx, &[])).collect::<Result<Vec<_>>>()?);
                }
                out
            }
            Plan::QueryDiff(diff, ordinality) => {
                let mut rows = diff.run(ctx)?;
                if *ordinality {
                    for (i, row) in rows.iter_mut().enumerate() {
                        row.push(Value::Int8(i as i64 + 1));
                    }
                }
                rows
            }
            Plan::Function { call, ordinality, width, defined } => {
                let expected = std::mem::replace(&mut ctx.session.expected_columns, defined.clone());
                let rows = set_rows(ctx, call, &[]);
                ctx.session.expected_columns = expected;
                let rows = rows?;
                rows.into_iter()
                    .enumerate()
                    .map(|(i, v)| {
                        let mut row = match v {
                            Value::Record(fields) if *width > 1 || fields.len() == *width => fields,
                            Value::Composite(c) if *width > 1 || c.fields.len() == *width => c.fields,
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
            Plan::XmlTable(table) => crate::xml::table::rows(ctx, table)?,
            Plan::JsonTable(table) => crate::jsontable::rows(ctx, table)?,
            Plan::RowsFrom { calls, ordinality } => {
                let columns = calls.iter().map(|c| set_rows(ctx, c, &[])).collect::<Result<Vec<_>>>()?;
                let count = columns.iter().map(Vec::len).max().unwrap_or(0);
                (0..count)
                    .map(|i| {
                        let mut row: Vec<Value> =
                            columns.iter().map(|c| c.get(i).cloned().unwrap_or(Value::Null)).collect();
                        if *ordinality {
                            row.push(Value::Int8(i as i64 + 1));
                        }
                        row
                    })
                    .collect()
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
            _ => unreachable!("open runs the plan nodes that read other nodes"),
        })
    }
}

/// set_operation combines two inputs' rows as UNION, INTERSECT, or EXCEPT do, keeping duplicates with ALL.
pub(crate) fn set_operation(op: SetOp, all: bool, left: Vec<Vec<Value>>, right: Vec<Vec<Value>>) -> Vec<Vec<Value>> {
    let mut right_groups = crate::exec::Groups::new();
    let mut right_counts: Vec<usize> = Vec::new();
    for row in &right {
        let (group, added) = right_groups.insert(row);
        if added {
            right_counts.push(0);
        }
        right_counts[group] += 1;
    }
    let mut out = Vec::new();
    let mut seen = crate::exec::Groups::new();
    match op {
        SetOp::Union => {
            for row in left.into_iter().chain(right) {
                if all || seen.insert(&row).1 {
                    out.push(row);
                }
            }
        }
        SetOp::Intersect => {
            for row in left {
                if let Some(group) = right_groups.find(&row)
                    && right_counts[group] > 0
                {
                    if all {
                        right_counts[group] -= 1;
                        out.push(row);
                    } else if seen.insert(&row).1 {
                        out.push(row);
                    }
                }
            }
        }
        SetOp::Except => {
            for row in left {
                match right_groups.find(&row) {
                    Some(group) if right_counts[group] > 0 => {
                        if all {
                            right_counts[group] -= 1;
                        }
                    }
                    _ => {
                        if all || seen.insert(&row).1 {
                            out.push(row);
                        }
                    }
                }
            }
        }
    }
    out
}

/// set_rows calls a function in FROM or a select list for its rows: each value a set-returning function returns, or
/// the one value of any other function.
pub(crate) fn set_rows(ctx: &mut Ctx<'_>, call: &Expr, row: &[Value]) -> Result<Vec<Value>> {
    match call {
        Expr::Func(index, args) => {
            let values = args.iter().map(|a| a.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
            crate::functions::call_set(ctx, *index, &values)
        }
        other => match other.eval(ctx, row)? {
            Value::Set(values) => Ok(values),
            value => Ok(vec![value]),
        },
    }
}

/// is_lateral reports whether a FROM item can see the items before it: a LATERAL subquery, a function, an XMLTABLE,
/// or a JSON_TABLE.
fn is_lateral(item: &Node) -> bool {
    match item.node.as_ref() {
        Some(NodeEnum::RangeSubselect(subselect)) => subselect.lateral,
        Some(NodeEnum::RangeFunction(_) | NodeEnum::RangeTableFunc(_) | NodeEnum::JsonTable(_)) => true,
        _ => false,
    }
}

/// conjuncts splits a predicate into the conditions that AND joins.
fn conjuncts(predicate: Expr, out: &mut Vec<Expr>) {
    match predicate {
        Expr::And(l, r) => {
            conjuncts(*l, out);
            conjuncts(*r, out);
        }
        other => out.push(other),
    }
}

/// push_down filters a plan's rows by a predicate, applying each condition of an inner join to the input it alone
/// reads and joining by the rest, so that the join never pairs rows they reject and a lateral side never runs for
/// them, as Postgres plans it, and giving a commit diff table the commits its conditions name.
pub(crate) fn push_down(plan: Plan, predicate: Expr) -> Plan {
    if let Plan::System(crate::dolt::tables::SystemTable::User(mut table)) = plan {
        let mut all = Vec::new();
        conjuncts(predicate.clone(), &mut all);
        table.take_commits(&all);
        return Plan::Filter {
            input: Box::new(Plan::System(crate::dolt::tables::SystemTable::User(table))),
            predicate,
        };
    }
    let plan = match plan {
        Plan::Join { left, right, kind, condition, lateral: false, method }
            if matches!(kind, JoinKind::Left | JoinKind::Right | JoinKind::Full) =>
        {
            let width = left.width();
            let mut strict = BTreeSet::new();
            strict_columns(&predicate, &mut strict);
            let (on_left, on_right) = (strict.iter().any(|&c| c < width), strict.iter().any(|&c| c >= width));
            let kind = match kind {
                JoinKind::Left if on_right => JoinKind::Inner,
                JoinKind::Right if on_left => JoinKind::Inner,
                JoinKind::Full if on_left && on_right => JoinKind::Inner,
                JoinKind::Full if on_left => JoinKind::Left,
                JoinKind::Full if on_right => JoinKind::Right,
                kind => kind,
            };
            Plan::Join { left, right, kind, condition, lateral: false, method }
        }
        plan => plan,
    };
    let plan = match plan {
        Plan::Join { left, right, kind: JoinKind::Left, condition, lateral: false, method } => {
            let width = left.width();
            let (_, keys) = condition.as_ref().map_or_else(Default::default, |c| join_keys(c, width));
            let unmatched = |c: &&Expr| match c {
                Expr::IsNull(column, false) => match **column {
                    Expr::Column(i) if i >= width => {
                        never_null(&right, i - width) || keys.contains(&Expr::Column(i - width))
                    }
                    _ => false,
                },
                _ => false,
            };
            let kind = match crate::indexscan::conjuncts(&predicate).iter().any(unmatched) {
                true => JoinKind::Anti,
                false => JoinKind::Left,
            };
            Plan::Join { left, right, kind, condition, lateral: false, method }
        }
        plan => plan,
    };
    let plan = match plan {
        Plan::Join { left, right, kind, condition, lateral: false, method }
            if matches!(kind, JoinKind::Left | JoinKind::Right | JoinKind::Anti | JoinKind::Semi) =>
        {
            return push_beside_outer_join(*left, *right, kind, condition, method, predicate);
        }
        plan => plan,
    };
    let Plan::Join { left, right, kind: JoinKind::Inner, condition, lateral, method } = plan else {
        return Plan::Filter { input: Box::new(plan), predicate };
    };
    let width = left.width();
    let mut all = Vec::new();
    conjuncts(predicate, &mut all);
    let implied = implied_equalities(all.iter().chain(condition.iter().flat_map(crate::indexscan::conjuncts)));
    all.extend(implied);
    let (mut to_left, mut to_right, mut to_join) = (Vec::new(), Vec::new(), Vec::new());
    for c in all {
        let (mut reads_left, mut reads_right, mut subquery) = (false, false, false);
        c.visit(&mut |e| match e {
            Expr::Column(i) if *i >= width => reads_right = true,
            Expr::Column(_) => reads_left = true,
            Expr::Exists(_) | Expr::Scalar(_) | Expr::ArraySubquery(..) | Expr::AnySubquery(..) => subquery = true,
            _ => {}
        });
        match (reads_left, reads_right, subquery) {
            (_, false, false) => to_left.push(c),
            (false, true, false) if !lateral => to_right.push(shift_columns(c, width)),
            _ => to_join.push(c),
        }
    }
    let and = |conditions: Vec<Expr>| conditions.into_iter().reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let left = match and(to_left) {
        Some(condition) => push_down(*left, condition),
        None => *left,
    };
    let right = match and(to_right) {
        Some(condition) => push_down(*right, condition),
        None => *right,
    };
    if lateral {
        let join = Plan::Join {
            left: Box::new(left),
            right: Box::new(right),
            kind: JoinKind::Inner,
            condition,
            lateral,
            method,
        };
        return match and(to_join) {
            Some(predicate) => Plan::Filter { input: Box::new(join), predicate },
            None => join,
        };
    }
    let condition = and(condition.into_iter().chain(to_join).collect());
    Plan::Join { left: Box::new(left), right: Box::new(right), kind: JoinKind::Inner, condition, lateral, method }
}

/// push_beside_outer_join pushes the conditions of a filter above an outer, semi, or anti join that read only the
/// input whose rows the join keeps whole, the left one or a right join's right one, down into that input, as Postgres
/// distributes quals that only that side's relations reference, keeping the others above the join.
fn push_beside_outer_join(
    left: Plan,
    right: Plan,
    kind: JoinKind,
    condition: Option<Expr>,
    method: JoinMethod,
    predicate: Expr,
) -> Plan {
    let width = left.width();
    let (mut to_left, mut to_right, mut kept) = (Vec::new(), Vec::new(), Vec::new());
    for c in crate::indexscan::conjuncts(&predicate) {
        let (mut reads_left, mut reads_right) = (false, false);
        c.visit(&mut |e| match e {
            Expr::Column(i) if *i >= width => reads_right = true,
            Expr::Column(_) => reads_left = true,
            _ => {}
        });
        match (reads_left, reads_right, has_subquery(c), kind) {
            (_, false, false, JoinKind::Left | JoinKind::Anti | JoinKind::Semi) => to_left.push(c.clone()),
            (false, true, false, JoinKind::Right) => to_right.push(shift_columns(c.clone(), width)),
            _ => kept.push(c.clone()),
        }
    }
    let and = |conditions: Vec<Expr>| conditions.into_iter().reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let left = match and(to_left) {
        Some(condition) => push_down(left, condition),
        None => left,
    };
    let right = match and(to_right) {
        Some(condition) => push_down(right, condition),
        None => right,
    };
    let join = Plan::Join { left: Box::new(left), right: Box::new(right), kind, condition, lateral: false, method };
    match and(kept) {
        Some(predicate) => Plan::Filter { input: Box::new(join), predicate },
        None => join,
    }
}

/// implied_equalities returns the equalities of columns with constants that a set of conditions imply without
/// stating, as Postgres' equivalence classes derive them: a column that equals another equals the constants that the
/// other equals. Only equalities that compare columns and constants directly, with no cast between differing types,
/// take part, since only those compare their values the same way throughout.
fn implied_equalities<'e>(conditions: impl Iterator<Item = &'e Expr>) -> Vec<Expr> {
    let mut classes: Vec<BTreeSet<usize>> = Vec::new();
    let mut constants: Vec<(usize, Expr)> = Vec::new();
    for c in conditions {
        let Expr::Compare(CmpOp::Eq, l, r) = c else { continue };
        match (&**l, &**r) {
            (Expr::Column(a), Expr::Column(b)) => {
                let (a, b) = (*a, *b);
                let found: Vec<usize> =
                    (0..classes.len()).filter(|&i| classes[i].contains(&a) || classes[i].contains(&b)).collect();
                let mut merged: BTreeSet<usize> = [a, b].into();
                for &i in found.iter().rev() {
                    merged.extend(classes.remove(i));
                }
                classes.push(merged);
            }
            (Expr::Column(a), constant @ Expr::Const(v)) | (constant @ Expr::Const(v), Expr::Column(a))
                if !v.is_null() =>
            {
                constants.push((*a, constant.clone()));
            }
            _ => {}
        }
    }
    let mut implied = Vec::new();
    for class in &classes {
        for (column, constant) in constants.iter().filter(|(c, _)| class.contains(c)) {
            for &other in class.iter().filter(|&o| o != column) {
                if !constants.iter().any(|(c, k)| *c == other && k == constant) {
                    implied.push(Expr::Compare(CmpOp::Eq, Box::new(Expr::Column(other)), Box::new(constant.clone())));
                }
            }
        }
    }
    implied.dedup();
    implied
}

/// strict_columns adds the columns that a condition can only be true with when they are not NULL, as Postgres'
/// find_nonnullable_vars finds them for reduce_outer_joins: the columns that comparisons and strict functions read
/// through casts and arithmetic, those of every condition an AND joins, and those of every arm of an OR.
fn strict_columns(condition: &Expr, out: &mut BTreeSet<usize>) {
    match condition {
        Expr::And(a, b) => {
            strict_columns(a, out);
            strict_columns(b, out);
        }
        Expr::Or(a, b) => {
            let (mut left, mut right) = (BTreeSet::new(), BTreeSet::new());
            strict_columns(a, &mut left);
            strict_columns(b, &mut right);
            out.extend(left.intersection(&right));
        }
        Expr::Compare(_, l, r) => {
            strict_reads(l, out);
            strict_reads(r, out);
        }
        Expr::IsNull(e, true) | Expr::BoolTest(e, Some(_), false) => strict_reads(e, out),
        Expr::Func(f, args) if crate::functions::function(*f).strict => args.iter().for_each(|a| strict_reads(a, out)),
        _ => {}
    }
}

/// strict_reads adds the columns whose NULL makes an expression NULL: a column itself, and those that casts,
/// arithmetic, and negation read.
fn strict_reads(e: &Expr, out: &mut BTreeSet<usize>) {
    match e {
        Expr::Column(i) => {
            out.insert(*i);
        }
        Expr::Cast(inner, ..) | Expr::Neg(inner, _) => strict_reads(inner, out),
        Expr::Arith(_, l, r, _) => {
            strict_reads(l, out);
            strict_reads(r, out);
        }
        _ => {}
    }
}

/// never_null reports whether a plan's rows never have NULL in the column, as a table's NOT NULL columns don't.
fn never_null(plan: &Plan, column: usize) -> bool {
    match plan {
        Plan::Scan(table, _) => table.columns.get(column).is_some_and(|c| !c.nullable),
        Plan::IndexScan(scan) => scan.table.columns.get(column).is_some_and(|c| !c.nullable),
        Plan::Filter { input, .. } => never_null(input, column),
        _ => false,
    }
}

/// join_keys returns the two sides of the equality conditions between a join's inputs, each side reading only its
/// own input's row, as a hash join finds matches by them.
pub(crate) fn join_keys(condition: &Expr, width: usize) -> (Vec<Expr>, Vec<Expr>) {
    let side = |e: &Expr| {
        let (mut left, mut right, mut other) = (false, false, false);
        e.visit(&mut |e| match e {
            Expr::Column(i) if *i >= width => right = true,
            Expr::Column(_) => left = true,
            Expr::Exists(_) | Expr::Scalar(_) | Expr::ArraySubquery(..) | Expr::AnySubquery(..) => other = true,
            _ => {}
        });
        match (left, right, other) {
            (true, false, false) => Some(true),
            (false, true, false) => Some(false),
            _ => None,
        }
    };
    let (mut left_keys, mut right_keys) = (Vec::new(), Vec::new());
    for c in crate::indexscan::conjuncts(condition) {
        if let Expr::Compare(CmpOp::Eq, a, b) = c {
            match (side(a), side(b)) {
                (Some(true), Some(false)) => {
                    left_keys.push((**a).clone());
                    right_keys.push(shift_columns((**b).clone(), width));
                }
                (Some(false), Some(true)) => {
                    left_keys.push((**b).clone());
                    right_keys.push(shift_columns((**a).clone(), width));
                }
                _ => {}
            }
        }
    }
    (left_keys, right_keys)
}

/// row_index indexes the rows of a `Once` plan by the sides of a filter's equality conditions that read them, as
/// Postgres' hashed subplans do.
pub(crate) fn row_index(ctx: &mut Ctx<'_>, rows: &[Vec<Value>], predicate: &Expr) -> RowIndex {
    let (inner, mut index) = index_parts(ctx, predicate);
    for (j, row) in rows.iter().enumerate() {
        index_row(ctx, &inner, &mut index, row, j);
    }
    index
}

/// index_parts splits a filter over a `Once` plan's rows into the sides of its equality conditions that read the rows,
/// which it returns, and a row index without rows yet, which has no table when there are no such conditions.
pub(crate) fn index_parts(ctx: &mut Ctx<'_>, predicate: &Expr) -> (Vec<Expr>, RowIndex) {
    let reads = |e: &Expr| {
        let (mut column, mut other) = (false, false);
        e.visit(&mut |e| match e {
            Expr::Column(_) => column = true,
            Expr::Outer(..) | Expr::Exists(_) | Expr::Scalar(_) | Expr::ArraySubquery(..) | Expr::AnySubquery(..) => {
                other = true
            }
            _ => {}
        });
        (column, other)
    };
    let (mut inner, mut outer, mut rest) = (Vec::new(), Vec::new(), Vec::new());
    for c in crate::indexscan::conjuncts(predicate) {
        if let Expr::Compare(CmpOp::Eq, a, b) = c {
            match (reads(a), reads(b)) {
                ((true, false), (false, _)) if !has_subquery(b) => {
                    inner.push((**a).clone());
                    outer.push((**b).clone());
                    continue;
                }
                ((false, _), (true, false)) if !has_subquery(a) => {
                    inner.push((**b).clone());
                    outer.push((**a).clone());
                    continue;
                }
                _ => {}
            }
        }
        rest.push(c.clone());
    }
    let residual = rest.into_iter().reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let residual = residual.unwrap_or(Expr::Const(Value::Bool(true))).fold(ctx);
    let predicate = predicate.clone().fold(ctx);
    let table = (!inner.is_empty()).then(Default::default);
    (inner, RowIndex { outer, predicate, residual, table })
}

/// index_row adds the row at a position to an index by the values of the sides of the equality conditions that read
/// it, leaving out a row with a NULL among them, which no condition holds for, and dropping the table when a value
/// has no hash key.
pub(crate) fn index_row(ctx: &mut Ctx<'_>, inner: &[Expr], index: &mut RowIndex, row: &[Value], j: usize) {
    let Some((table, buckets)) = index.table.as_mut() else { return };
    let mut key = JoinKey::with_capacity(inner.len());
    for e in inner {
        match e.eval(ctx, row) {
            Ok(Value::Null) => return,
            Ok(value) => match HashKey::of(value) {
                Some(k) => key.push(k),
                None => {
                    index.table = None;
                    return;
                }
            },
            Err(_) => {
                index.table = None;
                return;
            }
        }
    }
    let bucket = *table.entry(key).or_insert_with(|| {
        buckets.push(Vec::new());
        buckets.len() - 1
    });
    buckets[bucket].push(j);
}

/// share_scans wraps the table scan under the filter of a subquery that reads its enclosing rows in a `Once` plan, so
/// that each run of the subquery filters the same rows.
pub(crate) fn share_scans(plan: Plan) -> Plan {
    match plan {
        Plan::Filter { input, predicate }
            if matches!(*input, Plan::Scan(..) | Plan::IndexScan(_) | Plan::Catalog(_) | Plan::CatalogIndexScan(_)) =>
        {
            Plan::Filter { input: Box::new(Plan::Once(input)), predicate }
        }
        Plan::Project { input, exprs } => Plan::Project { input: Box::new(share_scans(*input)), exprs },
        Plan::Limit { input, limit, offset } => Plan::Limit { input: Box::new(share_scans(*input)), limit, offset },
        Plan::Sort { input, keys } => Plan::Sort { input: Box::new(share_scans(*input)), keys },
        Plan::Distinct { input, keys } => Plan::Distinct { input: Box::new(share_scans(*input)), keys },
        Plan::Aggregate { input, groups, aggregates, sets } => {
            Plan::Aggregate { input: Box::new(share_scans(*input)), groups, aggregates, sets }
        }
        other => other,
    }
}

/// has_subquery reports whether an expression holds a subquery.
pub(crate) fn has_subquery(e: &Expr) -> bool {
    let mut found = false;
    e.visit(&mut |e| {
        if matches!(e, Expr::Exists(_) | Expr::Scalar(_) | Expr::ArraySubquery(..) | Expr::AnySubquery(..)) {
            found = true;
        }
    });
    found
}

/// KeyMap maps join and filter keys to values, hashing them quickly.
pub(crate) type KeyMap<V> = std::collections::HashMap<JoinKey, V, foldhash::fast::FixedState>;

/// JoinKey is the hash keys of a row's join or filter values, held inline for the usual one or two.
pub(crate) type JoinKey = smallvec::SmallVec<[HashKey; 2]>;

/// KeySet is a set of IN keys, hashed quickly.
pub type KeySet = HashSet<HashKey, foldhash::fast::FixedState>;

/// HashKey is a join or IN key value whose equality matches the `=` comparison of the values it stands for.
#[derive(Debug, PartialEq, Eq, Hash)]
pub enum HashKey {
    Int(i64),
    Bytes(Vec<u8>),
    Bool(bool),
}

impl HashKey {
    /// of returns the key of a value, or None for a value whose equal values can differ in their representation.
    pub fn of(value: Value) -> Option<HashKey> {
        Some(match value {
            Value::Int2(i) => HashKey::Int(i as i64),
            Value::Int4(i) => HashKey::Int(i as i64),
            Value::Int8(i) => HashKey::Int(i),
            Value::Oid(o) => HashKey::Int(o as i64),
            Value::Date(d) => HashKey::Int(d as i64),
            Value::Text(text) => HashKey::Bytes(text.into_bytes()),
            Value::Bytea(bytes) => HashKey::Bytes(bytes),
            Value::Uuid(uuid) => HashKey::Bytes(uuid.to_vec()),
            Value::Bool(b) => HashKey::Bool(b),
            _ => return None,
        })
    }
}

/// shift_columns moves an expression's column references back by `width`, so that it reads the right input of a join
/// on its own.
fn shift_columns(expr: Expr, width: usize) -> Expr {
    match expr {
        Expr::Column(i) => Expr::Column(i - width),
        other => other.map_children(&mut |e| shift_columns(e, width)),
    }
}

/// rows_from_calls returns the function calls of a FROM function item: each call of ROWS FROM, one unnest call for
/// each array that unnest of several arrays expands, or the one call.
fn rows_from_calls(function: &RangeFunction) -> Result<Vec<pg_query::protobuf::FuncCall>> {
    let mut calls = Vec::new();
    for item in &function.functions {
        let Some(NodeEnum::List(list)) = item.node.as_ref() else { return Err(PgError::unsupported("this function")) };
        let Some(NodeEnum::FuncCall(call)) = list.items.first().and_then(|n| n.node.as_ref()) else {
            return Err(PgError::unsupported("this function in FROM"));
        };
        calls.push((**call).clone());
    }
    if let [call] = calls.as_slice()
        && call.args.len() > 1
        && call.funcname.iter().filter_map(node_name).next_back() == Some("unnest")
    {
        let funcname = ["pg_catalog", "unnest"]
            .map(|s| Node { node: Some(NodeEnum::String(pg_query::protobuf::String { sval: s.into() })) })
            .to_vec();
        return Ok(call
            .args
            .iter()
            .map(|arg| pg_query::protobuf::FuncCall {
                args: vec![arg.clone()],
                funcname: funcname.clone(),
                ..call.clone()
            })
            .collect());
    }
    Ok(calls)
}
