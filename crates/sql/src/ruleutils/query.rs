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

//! Postgres' ruleutils for queries: shaping a parsed SELECT with the FROM items that its column references resolve
//! against, and printing it as Postgres' get_query_def prints a view's query.

use pg_query::protobuf::{
    CoercionForm, CteMaterialize, FuncCall, JoinType, RangeVar, SelectStmt, SetOperation, SortBy, SortByDir,
    SortByNulls, SubLink, SubLinkType, WindowDef,
};
use pg_query::{Node, NodeEnum};

use super::{Analyzer, INDENT_JOIN, INDENT_STD, INDENT_VAR, Printer, SubKind, TExpr, trim_spaces, typ};
use crate::catalog::ColumnType;
use crate::engine::quote_identifier;
use crate::error::{PgError, Result};
use crate::expr::node_name;
use crate::oid;

/// The bits of a window frame's options, as Postgres' parse tree holds them.
const FRAME_NONDEFAULT: i32 = 0x1;
const FRAME_RANGE: i32 = 0x2;
const FRAME_ROWS: i32 = 0x4;
const FRAME_BETWEEN: i32 = 0x10;
const FRAME_START_UNBOUNDED_PRECEDING: i32 = 0x20;
const FRAME_END_UNBOUNDED_FOLLOWING: i32 = 0x100;
const FRAME_START_CURRENT_ROW: i32 = 0x200;
const FRAME_END_CURRENT_ROW: i32 = 0x400;
const FRAME_START_OFFSET_PRECEDING: i32 = 0x800;
const FRAME_END_OFFSET_PRECEDING: i32 = 0x1000;
const FRAME_EXCLUDE_CURRENT_ROW: i32 = 0x8000;
const FRAME_EXCLUDE_GROUP: i32 = 0x10000;
const FRAME_EXCLUDE_TIES: i32 = 0x20000;

/// Rte is a FROM item that column references resolve against: the name that qualifies its columns, and its columns.
#[derive(Clone, Debug)]
pub(super) struct Rte {
    name: String,
    columns: Vec<(String, ColumnType)>,
}

/// TQuery is an analyzed query, shaped as Postgres' get_query_def prints it.
#[derive(Clone, Debug)]
pub(super) struct TQuery {
    ctes: Vec<Cte>,
    recursive: bool,
    body: Body,
    /// The ORDER BY items, each with the keywords that follow it.
    order: Vec<(TExpr, String)>,
    offset: Option<TExpr>,
    /// The LIMIT count, which is None for LIMIT ALL.
    limit: Option<Option<TExpr>>,
}

/// Cte is a common table expression of a WITH clause.
#[derive(Clone, Debug)]
struct Cte {
    name: String,
    columns: Vec<String>,
    /// The MATERIALIZED keywords, with a space after them, or nothing.
    materialized: &'static str,
    query: TQuery,
}

/// Body is what a query computes before its ORDER BY and LIMIT.
#[derive(Clone, Debug)]
enum Body {
    Select(Box<Select>),
    Values(Vec<Vec<TExpr>>),
    /// A set operation's keyword, whether it keeps duplicates, and its inputs.
    SetOp(&'static str, bool, Box<TQuery>, Box<TQuery>),
}

/// Select is the SELECT of a query.
#[derive(Clone, Debug)]
struct Select {
    /// DISTINCT, with the expressions of DISTINCT ON when it has them.
    distinct: Option<Vec<TExpr>>,
    /// The result columns, each with its name.
    targets: Vec<(TExpr, String)>,
    from: Vec<From>,
    filter: Option<TExpr>,
    group: Vec<TExpr>,
    having: Option<TExpr>,
    /// The named windows of the WINDOW clause.
    windows: Vec<(String, Window)>,
}

/// Alias is the name a FROM item takes, with the names of its columns when they were given.
#[derive(Clone, Debug)]
struct Alias {
    name: String,
    columns: Vec<String>,
}

/// From is an item of a FROM clause.
#[derive(Clone, Debug)]
enum From {
    /// A table, view, or common table expression, as its name prints, with the alias it was given.
    Relation(String, Option<Alias>),
    /// A subquery, which is LATERAL when the flag is set.
    Subquery(bool, Box<TQuery>, Alias),
    /// A function call, which is LATERAL when the first flag is set and numbers its rows when the second is.
    Function(bool, TExpr, bool, Alias),
    Join(Box<Join>),
}

impl From {
    /// entries returns how many range table entries Postgres makes for a FROM item: one for each relation,
    /// subquery, or function, and one for each join besides its inputs.
    fn entries(&self) -> usize {
        match self {
            From::Join(join) => 1 + join.left.entries() + join.right.entries(),
            _ => 1,
        }
    }
}

/// Join is a join of two FROM items.
#[derive(Clone, Debug)]
struct Join {
    /// The join's keywords, with spaces around them.
    kind: &'static str,
    left: From,
    right: From,
    on: Option<TExpr>,
    using: Vec<String>,
    alias: Option<Alias>,
}

/// Agg is an aggregate or window function call with its modifiers.
#[derive(Clone, Debug)]
pub(super) struct Agg {
    name: String,
    args: Vec<TExpr>,
    star: bool,
    distinct: bool,
    order: Vec<(TExpr, String)>,
    filter: Option<TExpr>,
    over: Option<Window>,
    pub(super) ret: u32,
}

/// Window is the window of a window function call: a named window, or a window specification.
#[derive(Clone, Debug)]
enum Window {
    Named(String),
    Spec(Box<WindowSpec>),
}

/// WindowSpec is a window specification.
#[derive(Clone, Debug)]
struct WindowSpec {
    refname: String,
    partition: Vec<TExpr>,
    order: Vec<(TExpr, String)>,
    frame: i32,
    start: Option<TExpr>,
    end: Option<TExpr>,
}

impl TQuery {
    /// is_plain_setop reports whether the query is a set operation without a WITH, ORDER BY, OFFSET, or LIMIT, which
    /// prints inside an enclosing set operation without parentheses of its own.
    fn is_plain_setop(&self) -> bool {
        matches!(self.body, Body::SetOp(..)) && self.is_plain()
    }

    /// is_plain reports whether the query has no WITH, ORDER BY, OFFSET, or LIMIT.
    fn is_plain(&self) -> bool {
        self.ctes.is_empty() && self.order.is_empty() && self.offset.is_none() && self.limit.is_none()
    }

    /// columns returns the query's result columns, named by the names given and otherwise by its targets.
    fn columns(&self, names: &[String]) -> Vec<(String, ColumnType)> {
        let columns: Vec<(String, ColumnType)> = match &self.body {
            Body::Select(select) => select.targets.iter().map(|(e, n)| (n.clone(), e.ty())).collect(),
            Body::Values(rows) => rows
                .first()
                .map(|row| row.iter().enumerate().map(|(i, e)| (format!("column{}", i + 1), e.ty())).collect())
                .unwrap_or_default(),
            Body::SetOp(_, _, left, _) => left.columns(&[]),
        };
        columns.into_iter().enumerate().map(|(i, (n, t))| (names.get(i).cloned().unwrap_or(n), t)).collect()
    }
}

/// figure_colname returns the name Postgres' FigureColname gives a result column without an alias.
fn figure_colname(node: &Node) -> String {
    figure(node).map_or_else(|| "?column?".to_string(), |(name, _)| name)
}

/// figure returns the name FigureColnameInternal gives an expression, with how strongly it holds.
fn figure(node: &Node) -> Option<(String, u8)> {
    match node.node.as_ref()? {
        NodeEnum::ColumnRef(column) => match column.fields.last()?.node.as_ref()? {
            NodeEnum::String(s) => Some((s.sval.clone(), 2)),
            _ => None,
        },
        NodeEnum::AIndirection(indirection) => {
            match indirection.indirection.iter().rev().find_map(|n| match n.node.as_ref() {
                Some(NodeEnum::String(s)) => Some(s.sval.clone()),
                _ => None,
            }) {
                Some(name) => Some((name, 2)),
                None => figure(indirection.arg.as_deref()?),
            }
        }
        NodeEnum::FuncCall(call) => Some((call.funcname.iter().filter_map(node_name).next_back()?.to_string(), 2)),
        NodeEnum::AExpr(e) if e.kind == pg_query::protobuf::AExprKind::AexprNullif as i32 => Some(("nullif".into(), 2)),
        NodeEnum::TypeCast(cast) => {
            let inner = cast.arg.as_deref().and_then(figure);
            match inner {
                Some((name, 2)) => Some((name, 2)),
                _ => {
                    let name = cast.type_name.as_ref()?.names.iter().filter_map(node_name).next_back()?;
                    Some((name.to_string(), 1))
                }
            }
        }
        NodeEnum::CollateClause(collate) => figure(collate.arg.as_deref()?),
        NodeEnum::SubLink(link) => match SubLinkType::try_from(link.sub_link_type) {
            Ok(SubLinkType::ExistsSublink) => Some(("exists".into(), 2)),
            Ok(SubLinkType::ArraySublink) => Some(("array".into(), 2)),
            Ok(SubLinkType::ExprSublink) => {
                let Some(NodeEnum::SelectStmt(select)) = link.subselect.as_deref()?.node.as_ref() else { return None };
                let Some(NodeEnum::ResTarget(target)) = select.target_list.first()?.node.as_ref() else { return None };
                if !target.name.is_empty() {
                    return Some((target.name.clone(), 2));
                }
                Some((figure_colname(target.val.as_deref()?), 2))
            }
            _ => None,
        },
        NodeEnum::CaseExpr(case) => match case.defresult.as_deref().and_then(figure) {
            Some((name, 2)) => Some((name, 2)),
            _ => Some(("case".into(), 1)),
        },
        NodeEnum::AArrayExpr(_) => Some(("array".into(), 2)),
        NodeEnum::RowExpr(_) => Some(("row".into(), 2)),
        NodeEnum::CoalesceExpr(_) => Some(("coalesce".into(), 2)),
        NodeEnum::MinMaxExpr(m) => {
            let greatest = m.op == pg_query::protobuf::MinMaxOp::IsGreatest as i32;
            Some((if greatest { "greatest" } else { "least" }.into(), 2))
        }
        NodeEnum::SqlvalueFunction(f) => {
            use pg_query::protobuf::SqlValueFunctionOp as Op;
            let name = match Op::try_from(f.op).ok()? {
                Op::SvfopCurrentDate => "current_date",
                Op::SvfopCurrentTime | Op::SvfopCurrentTimeN => "current_time",
                Op::SvfopCurrentTimestamp | Op::SvfopCurrentTimestampN => "current_timestamp",
                Op::SvfopLocaltime | Op::SvfopLocaltimeN => "localtime",
                Op::SvfopLocaltimestamp | Op::SvfopLocaltimestampN => "localtimestamp",
                Op::SvfopCurrentRole => "current_role",
                Op::SvfopCurrentUser => "current_user",
                Op::SvfopUser => "user",
                Op::SvfopSessionUser => "session_user",
                Op::SvfopCurrentCatalog => "current_catalog",
                Op::SvfopCurrentSchema => "current_schema",
                _ => return None,
            };
            Some((name.into(), 2))
        }
        _ => None,
    }
}

/// sort_suffix returns the keywords that print after an ORDER BY item, which leave out the default direction and
/// NULLS placement.
fn sort_suffix(sort: &SortBy) -> String {
    let descending = sort.sortby_dir == SortByDir::SortbyDesc as i32;
    let nulls_first = match SortByNulls::try_from(sort.sortby_nulls) {
        Ok(SortByNulls::SortbyNullsFirst) => true,
        Ok(SortByNulls::SortbyNullsLast) => false,
        _ => descending,
    };
    match (descending, nulls_first) {
        (false, false) => "",
        (false, true) => " NULLS FIRST",
        (true, true) => " DESC",
        (true, false) => " DESC NULLS LAST",
    }
    .to_string()
}

/// alias_of returns an alias's name and column names.
fn alias_of(alias: &pg_query::protobuf::Alias) -> Alias {
    Alias {
        name: alias.aliasname.clone(),
        columns: alias.colnames.iter().filter_map(node_name).map(str::to_string).collect(),
    }
}

/// renamed returns columns renamed by the column names of an alias, in order.
fn renamed(columns: Vec<(String, ColumnType)>, names: &[String]) -> Vec<(String, ColumnType)> {
    columns.into_iter().enumerate().map(|(i, (n, t))| (names.get(i).cloned().unwrap_or(n), t)).collect()
}

/// window_function_type returns the result type of a built-in window function that is not an aggregate.
fn window_function_type(name: &str, args: &[TExpr]) -> u32 {
    match name {
        "row_number" | "rank" | "dense_rank" | "ntile" => oid::INT8,
        "percent_rank" | "cume_dist" => oid::FLOAT8,
        _ => args.first().map_or(oid::UNKNOWN, |a| a.ty().oid),
    }
}

impl Analyzer<'_, '_> {
    /// deparse_query prints a view's query as pg_get_viewdef does, naming its result columns `names`, prettily when
    /// asked, wrapping its target and FROM lists after `wrap` columns.
    pub fn deparse_query(&mut self, select: &SelectStmt, names: &[String], pretty: bool, wrap: i32) -> Result<String> {
        let query = self.query(select)?;
        let names = query.columns(names).into_iter().map(|(n, _)| n).collect::<Vec<_>>();
        let mut out = String::new();
        let mut printer = Printer { pretty, indents: true, level: 0, wrap, varprefix: true, nested: false };
        printer.query(&query, &mut out, Some(&names), true);
        Ok(out)
    }

    /// resolve_column returns the column of a FROM item in scope that a column reference names, or a whole row of one
    /// for a reference ending in `*`.
    pub(super) fn resolve_column(&self, names: &[&str], star: bool) -> Option<TExpr> {
        for scope in self.scopes.iter().rev() {
            match (names, star) {
                ([.., relation], true) => {
                    if let Some(rte) = scope.iter().find(|r| r.name == *relation) {
                        return Some(TExpr::Var(Some(rte.name.clone()), "*".into(), typ(oid::RECORD)));
                    }
                }
                ([name], false) => {
                    let found = scope.iter().find_map(|r| r.columns.iter().find(|(n, _)| n == name).map(|c| (r, c)));
                    if let Some((rte, (n, t))) = found {
                        return Some(TExpr::Var(Some(rte.name.clone()), n.clone(), *t));
                    }
                }
                ([.., relation, name], false) => {
                    let rte = scope.iter().find(|r| r.name == *relation);
                    if let Some((rte, (n, t))) =
                        rte.and_then(|r| r.columns.iter().find(|(n, _)| n == name).map(|c| (r, c)))
                    {
                        return Some(TExpr::Var(Some(rte.name.clone()), n.clone(), *t));
                    }
                }
                _ => {}
            }
        }
        None
    }

    /// call shapes a function call: a function with a SQL-standard spelling, an aggregate or window function with
    /// modifiers, or a plain call.
    pub(super) fn call(&mut self, call: &FuncCall) -> Result<TExpr> {
        let name = call.funcname.iter().filter_map(node_name).next_back().unwrap_or_default().to_string();
        let mut args = Vec::with_capacity(call.args.len());
        for arg in &call.args {
            args.push(self.analyze(arg)?);
        }
        if matches!(name.as_str(), "nextval" | "currval" | "setval") && !args.is_empty() {
            let first = args.remove(0);
            args.insert(0, self.coerce(first, typ(oid::REGCLASS)));
        }
        let types: Vec<u32> = args.iter().map(|a| a.ty().oid).collect();
        let modified = call.agg_star
            || call.agg_distinct
            || !call.agg_order.is_empty()
            || call.agg_filter.is_some()
            || call.over.is_some();
        if modified || crate::functions::aggregate::exists(None, &name) {
            let (args, ret) = match crate::functions::aggregate::resolve(&name, &types, -1) {
                Ok((_, params, ret)) => {
                    let mut coerced = Vec::with_capacity(args.len());
                    for (arg, &param) in args.into_iter().zip(&params) {
                        let param = match param {
                            oid::VARCHAR => oid::TEXT,
                            crate::functions::ANY | crate::functions::ANYELEMENT => arg.ty().oid,
                            param => param,
                        };
                        coerced.push(self.coerce(arg, typ(param)));
                    }
                    let ret = if ret == oid::VARCHAR { oid::TEXT } else { ret };
                    (coerced, ret)
                }
                Err(_) if call.agg_star => (args, oid::INT8),
                Err(_) => {
                    let ret = window_function_type(&name, &args);
                    (args, ret)
                }
            };
            if !modified {
                return Ok(TExpr::Func(name, args, ret));
            }
            let order = self.sort_items(&call.agg_order, &[])?;
            let filter = match call.agg_filter.as_deref() {
                Some(filter) => {
                    let analyzed = self.analyze(filter)?;
                    Some(self.coerce(analyzed, typ(oid::BOOL)))
                }
                None => None,
            };
            let over = match call.over.as_deref() {
                Some(window) => Some(self.window(window)?),
                None => None,
            };
            let (star, distinct) = (call.agg_star, call.agg_distinct);
            return Ok(TExpr::Agg(Box::new(Agg { name, args, star, distinct, order, filter, over, ret })));
        }
        match crate::functions::resolve(&name, &types, -1) {
            Ok(resolved) => {
                let mut coerced = Vec::with_capacity(args.len());
                for (arg, &target) in args.into_iter().zip(&resolved.arg_types) {
                    let target = if target == crate::functions::ANY { arg.ty().oid } else { target };
                    coerced.push(self.coerce(arg, typ(target)));
                }
                if call.funcformat == CoercionForm::CoerceSqlSyntax as i32 {
                    return Ok(TExpr::Syntax(name, coerced, resolved.ret));
                }
                Ok(TExpr::Func(name, coerced, resolved.ret))
            }
            Err(_) if call.funcformat == CoercionForm::CoerceSqlSyntax as i32 => {
                Ok(TExpr::Syntax(name, args, oid::UNKNOWN))
            }
            Err(_) => Ok(TExpr::Func(name, args, oid::UNKNOWN)),
        }
    }

    /// window shapes the window of a window function call or of a WINDOW clause.
    fn window(&mut self, window: &WindowDef) -> Result<Window> {
        if !window.name.is_empty() && window.partition_clause.is_empty() && window.order_clause.is_empty() {
            return Ok(Window::Named(window.name.clone()));
        }
        let partition = window.partition_clause.iter().map(|p| self.analyze(p)).collect::<Result<Vec<_>>>()?;
        let order = self.sort_items(&window.order_clause, &[])?;
        let start = match window.start_offset.as_deref() {
            Some(offset) => Some(self.analyze(offset)?),
            None => None,
        };
        let end = match window.end_offset.as_deref() {
            Some(offset) => Some(self.analyze(offset)?),
            None => None,
        };
        let refname = window.refname.clone();
        Ok(Window::Spec(Box::new(WindowSpec { refname, partition, order, frame: window.frame_options, start, end })))
    }

    /// sublink shapes a subquery in an expression.
    pub(super) fn sublink(&mut self, link: &SubLink) -> Result<TExpr> {
        let Some(NodeEnum::SelectStmt(select)) = link.subselect.as_deref().and_then(|n| n.node.as_ref()) else {
            return Err(PgError::unsupported("this subquery"));
        };
        let query = Box::new(self.query(select)?);
        let first = query.columns(&[]).first().map_or(oid::UNKNOWN, |(_, t)| t.oid);
        let op = link.oper_name.iter().filter_map(node_name).next_back().unwrap_or("=").to_string();
        Ok(match SubLinkType::try_from(link.sub_link_type) {
            Ok(SubLinkType::ExistsSublink) => TExpr::Sub(SubKind::Exists, None, op, query, oid::BOOL),
            Ok(kind @ (SubLinkType::AnySublink | SubLinkType::AllSublink)) => {
                let test = match link.testexpr.as_deref() {
                    Some(test) => {
                        let analyzed = self.analyze(test)?;
                        let (left, _) = Self::operand_types(&op, analyzed.ty().oid, first);
                        Some(Box::new(self.coerce(analyzed, typ(left))))
                    }
                    None => None,
                };
                let kind = if kind == SubLinkType::AnySublink { SubKind::Any } else { SubKind::All };
                TExpr::Sub(kind, test, op, query, oid::BOOL)
            }
            Ok(SubLinkType::ArraySublink) => TExpr::Sub(SubKind::Array, None, op, query, crate::expr::array_of(first)),
            _ => TExpr::Sub(SubKind::Value, None, op, query, first),
        })
    }

    /// sort_items shapes ORDER BY items, which may name result columns by position or name.
    fn sort_items(&mut self, items: &[Node], targets: &[(TExpr, String)]) -> Result<Vec<(TExpr, String)>> {
        let mut out = Vec::with_capacity(items.len());
        for item in items {
            let Some(NodeEnum::SortBy(sort)) = item.node.as_ref() else { continue };
            let Some(node) = sort.node.as_deref() else { continue };
            out.push((self.target_reference(node, targets, true)?, sort_suffix(sort)));
        }
        Ok(out)
    }

    /// target_reference shapes an ORDER BY, GROUP BY, or DISTINCT ON item, which names a result column by its
    /// position, or by its name when `by_name` is set and no column of the FROM items has that name.
    fn target_reference(&mut self, node: &Node, targets: &[(TExpr, String)], by_name: bool) -> Result<TExpr> {
        match node.node.as_ref() {
            Some(NodeEnum::AConst(c)) => {
                if let Some(pg_query::protobuf::a_const::Val::Ival(i)) = &c.val
                    && let Some((expr, _)) = usize::try_from(i.ival).ok().and_then(|i| targets.get(i.wrapping_sub(1)))
                {
                    return Ok(expr.clone());
                }
            }
            Some(NodeEnum::ColumnRef(column)) if by_name && column.fields.len() == 1 => {
                let name = column.fields.iter().filter_map(node_name).next().unwrap_or_default();
                if let Some((expr, _)) = targets.iter().find(|(_, n)| n == name) {
                    return Ok(expr.clone());
                }
            }
            _ => {}
        }
        self.analyze(node)
    }

    /// query shapes a SELECT, VALUES, or set operation, with its WITH clause.
    pub(super) fn query(&mut self, select: &SelectStmt) -> Result<TQuery> {
        let cte_mark = self.ctes.len();
        let recursive = select.with_clause.as_ref().is_some_and(|w| w.recursive);
        let mut ctes = Vec::new();
        for node in select.with_clause.iter().flat_map(|w| &w.ctes) {
            let Some(NodeEnum::CommonTableExpr(cte)) = node.node.as_ref() else { continue };
            let Some(NodeEnum::SelectStmt(inner)) = cte.ctequery.as_deref().and_then(|n| n.node.as_ref()) else {
                continue;
            };
            let names: Vec<String> = cte.aliascolnames.iter().filter_map(node_name).map(str::to_string).collect();
            if recursive && let Some(first) = inner.larg.as_deref() {
                let columns = self.query(first)?.columns(&names);
                self.ctes.push((cte.ctename.clone(), columns));
            }
            let query = self.query(inner)?;
            if recursive && inner.larg.is_some() {
                self.ctes.pop();
            }
            self.ctes.push((cte.ctename.clone(), query.columns(&names)));
            let materialized = match CteMaterialize::try_from(cte.ctematerialized) {
                Ok(CteMaterialize::Always) => "MATERIALIZED ",
                Ok(CteMaterialize::Never) => "NOT MATERIALIZED ",
                _ => "",
            };
            ctes.push(Cte { name: cte.ctename.clone(), columns: names, materialized, query });
        }
        let op = SetOperation::try_from(select.op).unwrap_or(SetOperation::SetopNone);
        let mut query = match op {
            SetOperation::SetopUnion | SetOperation::SetopIntersect | SetOperation::SetopExcept => {
                let (Some(left), Some(right)) = (select.larg.as_deref(), select.rarg.as_deref()) else {
                    return Err(PgError::unsupported("this set operation"));
                };
                let keyword = match op {
                    SetOperation::SetopUnion => "UNION",
                    SetOperation::SetopIntersect => "INTERSECT",
                    _ => "EXCEPT",
                };
                let (left, right) = (self.query(left)?, self.query(right)?);
                let names: Vec<(TExpr, String)> =
                    left.columns(&[]).into_iter().map(|(n, t)| (TExpr::Var(None, n.clone(), t), n)).collect();
                let order = self.sort_items(&select.sort_clause, &names)?;
                let body = Body::SetOp(keyword, select.all, Box::new(left), Box::new(right));
                TQuery { ctes: Vec::new(), recursive: false, body, order, offset: None, limit: None }
            }
            _ if !select.values_lists.is_empty() => {
                let mut rows = Vec::new();
                for row in &select.values_lists {
                    let Some(NodeEnum::List(list)) = row.node.as_ref() else { continue };
                    rows.push(list.items.iter().map(|i| self.analyze(i)).collect::<Result<Vec<_>>>()?);
                }
                let width = rows.first().map_or(0, Vec::len);
                for column in 0..width {
                    let items: Vec<TExpr> = rows.iter().filter_map(|r| r.get(column).cloned()).collect();
                    let common = Self::common_type(&items);
                    for row in &mut rows {
                        if let Some(item) = row.get_mut(column) {
                            let taken = std::mem::replace(item, TExpr::Raw(String::new()));
                            *item = self.coerce(taken, typ(common));
                        }
                    }
                }
                let order = self.sort_items(&select.sort_clause, &[])?;
                TQuery {
                    ctes: Vec::new(),
                    recursive: false,
                    body: Body::Values(rows),
                    order,
                    offset: None,
                    limit: None,
                }
            }
            _ => self.select(select)?,
        };
        query.offset = match select.limit_offset.as_deref() {
            Some(offset) => {
                let analyzed = self.analyze(offset)?;
                Some(self.coerce(analyzed, typ(oid::INT8)))
            }
            None => None,
        };
        query.limit = match select.limit_count.as_deref() {
            Some(Node { node: Some(NodeEnum::AConst(c)) }) if c.isnull => Some(None),
            Some(count) => {
                let analyzed = self.analyze(count)?;
                Some(Some(self.coerce(analyzed, typ(oid::INT8))))
            }
            None => None,
        };
        query.ctes = ctes;
        query.recursive = recursive;
        self.ctes.truncate(cte_mark);
        Ok(query)
    }

    /// select shapes a SELECT with its FROM items in scope, with its ORDER BY.
    fn select(&mut self, select: &SelectStmt) -> Result<TQuery> {
        let mark = self.scopes.len();
        self.scopes.push(Vec::new());
        let result = self.select_in_scope(select);
        self.scopes.truncate(mark);
        result
    }

    /// select_in_scope shapes a SELECT whose FROM items go into the innermost scope.
    fn select_in_scope(&mut self, select: &SelectStmt) -> Result<TQuery> {
        let mut from = Vec::with_capacity(select.from_clause.len());
        for item in &select.from_clause {
            from.push(self.source(item)?);
        }
        let mut targets = Vec::new();
        for node in &select.target_list {
            let Some(NodeEnum::ResTarget(target)) = node.node.as_ref() else { continue };
            let Some(val) = target.val.as_deref() else { continue };
            if let Some(NodeEnum::ColumnRef(column)) = val.node.as_ref()
                && matches!(column.fields.last().and_then(|f| f.node.as_ref()), Some(NodeEnum::AStar(_)))
            {
                let relation = column.fields.iter().filter_map(node_name).next_back();
                let scope = self.scopes.last().cloned().unwrap_or_default();
                for rte in scope.iter().filter(|r| relation.is_none_or(|n| r.name == n)) {
                    for (name, ty) in &rte.columns {
                        targets.push((TExpr::Var(Some(rte.name.clone()), name.clone(), *ty), name.clone()));
                    }
                }
                continue;
            }
            let name = if target.name.is_empty() { figure_colname(val) } else { target.name.clone() };
            targets.push((self.analyze(val)?, name));
        }
        let filter = match select.where_clause.as_deref() {
            Some(filter) => {
                let analyzed = self.analyze(filter)?;
                Some(self.coerce(analyzed, typ(oid::BOOL)))
            }
            None => None,
        };
        let mut group = Vec::with_capacity(select.group_clause.len());
        for item in &select.group_clause {
            group.push(self.target_reference(item, &targets, false)?);
        }
        let having = match select.having_clause.as_deref() {
            Some(having) => {
                let analyzed = self.analyze(having)?;
                Some(self.coerce(analyzed, typ(oid::BOOL)))
            }
            None => None,
        };
        let distinct = match select.distinct_clause.as_slice() {
            [] => None,
            [Node { node: None }] => Some(Vec::new()),
            items => Some(items.iter().map(|i| self.target_reference(i, &targets, true)).collect::<Result<Vec<_>>>()?),
        };
        let mut windows = Vec::new();
        for node in &select.window_clause {
            let Some(NodeEnum::WindowDef(window)) = node.node.as_ref() else { continue };
            let spec = self.window(&WindowDef { name: String::new(), ..(**window).clone() })?;
            windows.push((window.name.clone(), spec));
        }
        let order = self.sort_items(&select.sort_clause, &targets)?;
        let body = Body::Select(Box::new(Select { distinct, targets, from, filter, group, having, windows }));
        Ok(TQuery { ctes: Vec::new(), recursive: false, body, order, offset: None, limit: None })
    }

    /// add_rte puts a FROM item's columns in the innermost scope.
    fn add_rte(&mut self, name: String, columns: Vec<(String, ColumnType)>) {
        if let Some(scope) = self.scopes.last_mut() {
            scope.push(Rte { name, columns });
        }
    }

    /// relation_columns returns the columns of a table or view, as `SELECT *` returns them.
    fn relation_columns(&mut self, relation: &RangeVar) -> Result<Vec<(String, ColumnType)>> {
        let name = match relation.schemaname.as_str() {
            "" => quote_identifier(&relation.relname),
            schema => format!("{}.{}", quote_identifier(schema), quote_identifier(&relation.relname)),
        };
        let parsed = pg_query::parse(&format!("SELECT * FROM {name}"), 0).map_err(PgError::internal)?;
        let statement = parsed.protobuf.stmts.into_iter().next().and_then(|s| s.stmt).and_then(|s| s.node);
        let Some(NodeEnum::SelectStmt(select)) = statement else { return Ok(Vec::new()) };
        let query = crate::plan::Planner { ctx: self.ctx, outer: Vec::new() }.plan_query(&select)?;
        Ok(query.columns.iter().zip(&query.types).map(|(c, &t)| (c.name.clone(), t)).collect())
    }

    /// source shapes an item of a FROM clause, putting its columns in scope.
    fn source(&mut self, node: &Node) -> Result<From> {
        Ok(match node.node.as_ref() {
            Some(NodeEnum::RangeVar(relation)) => {
                let alias = relation.alias.as_ref().map(alias_of);
                let cte =
                    self.ctes.iter().rev().find(|(n, _)| relation.schemaname.is_empty() && *n == relation.relname);
                let columns = match cte {
                    Some((_, columns)) => columns.clone(),
                    None => self.relation_columns(relation)?,
                };
                let visible =
                    relation.schemaname.is_empty() || self.ctx.effective_search_path().contains(&relation.schemaname);
                let text = if visible {
                    quote_identifier(&relation.relname)
                } else {
                    format!("{}.{}", quote_identifier(&relation.schemaname), quote_identifier(&relation.relname))
                };
                let name = alias.as_ref().map_or(relation.relname.clone(), |a| a.name.clone());
                let names = alias.as_ref().map(|a| a.columns.clone()).unwrap_or_default();
                self.add_rte(name, renamed(columns, &names));
                From::Relation(text, alias)
            }
            Some(NodeEnum::RangeSubselect(sub)) => {
                let Some(NodeEnum::SelectStmt(select)) = sub.subquery.as_deref().and_then(|n| n.node.as_ref()) else {
                    return Err(PgError::unsupported("this subquery in FROM"));
                };
                let query = self.query(select)?;
                let alias = sub
                    .alias
                    .as_ref()
                    .map(alias_of)
                    .unwrap_or(Alias { name: "unnamed_subquery".into(), columns: Vec::new() });
                let columns = query.columns(&alias.columns);
                self.add_rte(alias.name.clone(), columns);
                From::Subquery(sub.lateral, Box::new(query), alias)
            }
            Some(NodeEnum::RangeFunction(function)) => {
                let call = function.functions.first().and_then(|f| match f.node.as_ref() {
                    Some(NodeEnum::List(list)) => list.items.first().cloned(),
                    _ => None,
                });
                let Some(call) = call else { return Err(PgError::unsupported("this function in FROM")) };
                let analyzed = self.analyze(&call)?;
                let function_name = match call.node.as_ref() {
                    Some(NodeEnum::FuncCall(f)) => {
                        f.funcname.iter().filter_map(node_name).next_back().unwrap_or_default()
                    }
                    _ => "",
                }
                .to_string();
                let given = function.alias.as_ref().map(alias_of);
                let name = given.as_ref().map_or(function_name.clone(), |a| a.name.clone());
                let mut columns = given.map(|a| a.columns).unwrap_or_default();
                if columns.is_empty() {
                    columns.push(name.clone());
                }
                let mut typed: Vec<(String, ColumnType)> = columns.iter().map(|c| (c.clone(), analyzed.ty())).collect();
                if function.ordinality
                    && typed.len() > 1
                    && let Some(last) = typed.last_mut()
                {
                    last.1 = typ(oid::INT8);
                }
                self.add_rte(name.clone(), typed);
                From::Function(function.lateral, analyzed, function.ordinality, Alias { name, columns })
            }
            Some(NodeEnum::JoinExpr(join)) => {
                let (Some(left), Some(right)) = (join.larg.as_deref(), join.rarg.as_deref()) else {
                    return Err(PgError::unsupported("this join"));
                };
                let (left, right) = (self.source(left)?, self.source(right)?);
                let using: Vec<String> = join.using_clause.iter().filter_map(node_name).map(str::to_string).collect();
                let on = match join.quals.as_deref() {
                    Some(quals) => {
                        let analyzed = self.analyze(quals)?;
                        Some(self.coerce(analyzed, typ(oid::BOOL)))
                    }
                    None => None,
                };
                let kind = match JoinType::try_from(join.jointype) {
                    Ok(JoinType::JoinLeft) => " LEFT JOIN ",
                    Ok(JoinType::JoinFull) => " FULL JOIN ",
                    Ok(JoinType::JoinRight) => " RIGHT JOIN ",
                    _ if on.is_none() && using.is_empty() && !join.is_natural => " CROSS JOIN ",
                    _ => " JOIN ",
                };
                let alias = join.alias.as_ref().map(alias_of);
                From::Join(Box::new(Join { kind, left, right, on, using, alias }))
            }
            _ => return Err(PgError::unsupported("this FROM item")),
        })
    }
}

impl Printer {
    /// query prints a query as Postgres' get_query_def does, naming its result columns `names` when they are given,
    /// and showing the name of every result column when `visible` is set.
    pub(super) fn query(&mut self, query: &TQuery, out: &mut String, names: Option<&[String]>, visible: bool) {
        let saved = (self.level, self.varprefix, self.nested);
        let entries = match &query.body {
            Body::Select(select) => select.from.iter().map(From::entries).sum(),
            _ => 0,
        };
        self.varprefix = self.nested || entries != 1;
        self.nested = true;
        self.with_clause(query, out);
        match &query.body {
            Body::Select(select) => self.select(select, out, names, visible),
            Body::Values(rows) => {
                if self.indents {
                    self.level += INDENT_STD;
                    out.push(' ');
                }
                self.values(rows, out);
            }
            Body::SetOp(..) => self.setop(query, out, names, visible, true),
        }
        if !query.order.is_empty() {
            self.context_keyword(out, " ORDER BY ", -INDENT_STD, INDENT_STD, 1);
            let order = self.order_list(&query.order);
            out.push_str(&order);
        }
        if let Some(offset) = &query.offset {
            self.context_keyword(out, " OFFSET ", -INDENT_STD, INDENT_STD, 0);
            let text = self.print(offset, None, false);
            out.push_str(&text);
        }
        if let Some(limit) = &query.limit {
            self.context_keyword(out, " LIMIT ", -INDENT_STD, INDENT_STD, 0);
            let text = limit.as_ref().map_or_else(|| "ALL".to_string(), |l| self.print(l, None, false));
            out.push_str(&text);
        }
        (self.level, self.varprefix, self.nested) = saved;
    }

    /// with_clause prints a query's WITH clause, as Postgres' get_with_clause does.
    fn with_clause(&mut self, query: &TQuery, out: &mut String) {
        if query.ctes.is_empty() {
            return;
        }
        if self.indents {
            self.level += INDENT_STD;
            out.push(' ');
        }
        let mut separator = if query.recursive { "WITH RECURSIVE " } else { "WITH " };
        for cte in &query.ctes {
            out.push_str(separator);
            out.push_str(&quote_identifier(&cte.name));
            if !cte.columns.is_empty() {
                let columns: Vec<String> = cte.columns.iter().map(|c| quote_identifier(c)).collect();
                out.push_str(&format!("({})", columns.join(", ")));
            }
            out.push_str(" AS ");
            out.push_str(cte.materialized);
            out.push('(');
            if self.indents {
                self.context_keyword(out, "", 0, 0, 0);
            }
            self.query(&cte.query, out, None, true);
            if self.indents {
                self.context_keyword(out, "", 0, 0, 0);
            }
            out.push(')');
            separator = ", ";
        }
        if self.indents {
            self.level -= INDENT_STD;
            self.context_keyword(out, "", 0, 0, 0);
        } else {
            out.push(' ');
        }
    }

    /// select prints a SELECT, as Postgres' get_basic_select_query does.
    fn select(&mut self, select: &Select, out: &mut String, names: Option<&[String]>, visible: bool) {
        if self.indents {
            self.level += INDENT_STD;
            out.push(' ');
        }
        out.push_str("SELECT");
        match &select.distinct {
            Some(on) if on.is_empty() => out.push_str(" DISTINCT"),
            Some(on) => {
                let items: Vec<String> = on.iter().map(|e| self.sort_group(e)).collect();
                out.push_str(&format!(" DISTINCT ON ({})", items.join(", ")));
            }
            None => {}
        }
        self.target_list(&select.targets, out, names, visible);
        self.sources(&select.from, out);
        if let Some(filter) = &select.filter {
            self.context_keyword(out, " WHERE ", -INDENT_STD, INDENT_STD, 1);
            let text = self.print(filter, None, false);
            out.push_str(&text);
        }
        if !select.group.is_empty() {
            self.context_keyword(out, " GROUP BY ", -INDENT_STD, INDENT_STD, 1);
            let items: Vec<String> = select.group.iter().map(|e| self.sort_group(e)).collect();
            out.push_str(&items.join(", "));
        }
        if let Some(having) = &select.having {
            self.context_keyword(out, " HAVING ", -INDENT_STD, INDENT_STD, 0);
            let text = self.print(having, None, false);
            out.push_str(&text);
        }
        for (i, (name, window)) in select.windows.iter().enumerate() {
            if i == 0 {
                self.context_keyword(out, " WINDOW ", -INDENT_STD, INDENT_STD, 1);
            } else {
                out.push_str(", ");
            }
            let spec = self.window(window);
            out.push_str(&format!("{} AS {spec}", quote_identifier(name)));
        }
    }

    /// target_list prints the result columns of a SELECT, as Postgres' get_target_list does.
    fn target_list(&mut self, targets: &[(TExpr, String)], out: &mut String, names: Option<&[String]>, visible: bool) {
        let mut separator = " ";
        let mut last_multiline = false;
        for (i, (expr, name)) in targets.iter().enumerate() {
            out.push_str(separator);
            separator = ", ";
            let mut text = self.print(expr, None, true);
            let attname = match expr {
                TExpr::Var(_, column, _) => Some(column.as_str()),
                _ if visible => None,
                _ => Some("?column?"),
            };
            let colname = names.and_then(|n| n.get(i)).unwrap_or(name);
            if attname != Some(colname.as_str()) {
                text.push_str(&format!(" AS {}", quote_identifier(colname)));
            }
            if self.indents && self.wrap >= 0 {
                let leading = text.starts_with('\n');
                if leading {
                    trim_spaces(out);
                } else {
                    let line = out.rsplit('\n').next().map_or(0, str::len);
                    if i > 0 && (line + text.len() > self.wrap as usize || last_multiline) {
                        self.context_keyword(out, "", -INDENT_STD, INDENT_STD, INDENT_VAR);
                    }
                }
                last_multiline = text[usize::from(leading)..].contains('\n');
            }
            out.push_str(&text);
        }
    }

    /// sources prints the FROM clause of a SELECT, as Postgres' get_from_clause does.
    fn sources(&mut self, from: &[From], out: &mut String) {
        for (i, item) in from.iter().enumerate() {
            if i == 0 {
                self.context_keyword(out, " FROM ", -INDENT_STD, INDENT_STD, 2);
                let text = self.source(item);
                out.push_str(&text);
                continue;
            }
            out.push_str(", ");
            let text = self.source(item);
            if self.indents && self.wrap >= 0 {
                if text.starts_with('\n') {
                    trim_spaces(out);
                } else {
                    let line = out.rsplit('\n').next().map_or(0, str::len);
                    if line + text.len() > self.wrap as usize {
                        self.context_keyword(out, "", -INDENT_STD, INDENT_STD, INDENT_VAR);
                    }
                }
            }
            out.push_str(&text);
        }
    }

    /// alias prints the alias of a FROM item, with its column names when it has them.
    fn alias(alias: &Alias) -> String {
        let mut out = format!(" {}", quote_identifier(&alias.name));
        if !alias.columns.is_empty() {
            let columns: Vec<String> = alias.columns.iter().map(|c| quote_identifier(c)).collect();
            out.push_str(&format!("({})", columns.join(", ")));
        }
        out
    }

    /// source prints an item of a FROM clause, as Postgres' get_from_clause_item does.
    fn source(&mut self, item: &From) -> String {
        match item {
            From::Relation(text, alias) => format!("{text}{}", alias.as_ref().map(Self::alias).unwrap_or_default()),
            From::Subquery(lateral, query, alias) => {
                let mut out = if *lateral { "LATERAL (".to_string() } else { "(".to_string() };
                self.query(query, &mut out, None, true);
                out.push(')');
                out.push_str(&Self::alias(alias));
                out
            }
            From::Function(lateral, call, ordinality, alias) => {
                let mut out = if *lateral { "LATERAL ".to_string() } else { String::new() };
                out.push_str(&self.print(call, None, true));
                if *ordinality {
                    out.push_str(" WITH ORDINALITY");
                }
                out.push_str(&Self::alias(alias));
                out
            }
            From::Join(join) => {
                let mut out = String::new();
                let parenthesized = !self.pretty || join.alias.is_some();
                if parenthesized {
                    out.push('(');
                }
                let left = self.source(&join.left);
                out.push_str(&left);
                let kind = if join.kind == " JOIN " && join.on.is_none() && join.using.is_empty() {
                    " CROSS JOIN "
                } else {
                    join.kind
                };
                self.context_keyword(&mut out, kind, -INDENT_STD, INDENT_STD, INDENT_JOIN);
                let right_paren = self.pretty && matches!(&join.right, From::Join(right) if right.alias.is_none());
                if right_paren {
                    out.push('(');
                }
                let right = self.source(&join.right);
                out.push_str(&right);
                if right_paren {
                    out.push(')');
                }
                if !join.using.is_empty() {
                    let columns: Vec<String> = join.using.iter().map(|c| quote_identifier(c)).collect();
                    out.push_str(&format!(" USING ({})", columns.join(", ")));
                } else if let Some(on) = &join.on {
                    out.push_str(" ON ");
                    let text = self.print(on, None, false);
                    out.push_str(&if self.pretty { text } else { format!("({text})") });
                } else if !matches!(join.kind, " JOIN " | " CROSS JOIN ") {
                    out.push_str(" ON TRUE");
                }
                if parenthesized {
                    out.push(')');
                }
                if let Some(alias) = &join.alias {
                    out.push_str(&Self::alias(alias));
                }
                out
            }
        }
    }

    /// values prints the rows of a VALUES list, as Postgres' get_values_def does.
    fn values(&mut self, rows: &[Vec<TExpr>], out: &mut String) {
        out.push_str("VALUES ");
        for (i, row) in rows.iter().enumerate() {
            if i > 0 {
                out.push_str(", ");
            }
            let items: Vec<String> = row.iter().map(|e| self.print(e, None, false)).collect();
            out.push_str(&format!("({})", items.join(",")));
        }
    }

    /// setop prints a set operation, or a query under one, as Postgres' get_setop_query does.
    fn setop(&mut self, query: &TQuery, out: &mut String, names: Option<&[String]>, visible: bool, top: bool) {
        let Body::SetOp(keyword, all, left, right) = &query.body else {
            let parenthesized = !query.is_plain();
            if parenthesized {
                out.push('(');
            }
            self.query(query, out, names, visible);
            if parenthesized {
                out.push(')');
            }
            return;
        };
        if !top && !query.is_plain() {
            out.push('(');
            self.query(query, out, names, visible);
            out.push(')');
            return;
        }
        let left_paren = match &left.body {
            Body::SetOp(left_keyword, left_all, ..) if left.is_plain() => left_keyword != keyword || left_all != all,
            _ => false,
        };
        let indent = if left_paren { INDENT_STD } else { 0 };
        if left_paren {
            out.push('(');
            self.context_keyword(out, "", indent, 0, 0);
        }
        self.setop(left, out, names, visible, false);
        if left_paren {
            self.context_keyword(out, ") ", -indent, 0, 0);
        } else if self.indents {
            self.context_keyword(out, "", -indent, 0, 0);
        } else {
            out.push(' ');
        }
        out.push_str(keyword);
        out.push(' ');
        if *all {
            out.push_str("ALL ");
        }
        let right_paren = right.is_plain_setop();
        let indent = if right_paren { INDENT_STD } else { 0 };
        if right_paren {
            out.push('(');
        }
        self.context_keyword(out, "", indent, 0, 0);
        self.setop(right, out, names, false, false);
        if self.indents {
            self.level -= indent;
        }
        if right_paren {
            self.context_keyword(out, ")", 0, 0, 0);
        }
    }

    /// order_list prints ORDER BY items with their keywords.
    fn order_list(&mut self, items: &[(TExpr, String)]) -> String {
        let printed: Vec<String> = items.iter().map(|(e, suffix)| format!("{}{suffix}", self.sort_group(e))).collect();
        printed.join(", ")
    }

    /// sort_group prints an ORDER BY, GROUP BY, or DISTINCT ON item, as Postgres' get_rule_sortgroupclause does: a
    /// constant always with its type, so that it does not read as a column position, and a function call in
    /// parentheses.
    fn sort_group(&mut self, e: &TExpr) -> String {
        match e {
            TExpr::Const(text, ty) => {
                let body = Self::constant(text.as_deref(), *ty, false);
                let type_name = crate::cast::format_type(ty.oid, Some(ty.modifier)).unwrap_or_else(|| "unknown".into());
                format!("{body}::{type_name}")
            }
            TExpr::Var(..) => self.print(e, None, true),
            _ => {
                let text = self.print(e, None, true);
                if self.pretty || matches!(e, TExpr::Func(..) | TExpr::Agg(_)) { format!("({text})") } else { text }
            }
        }
    }

    /// sublink prints a subquery in an expression, as Postgres' get_sublink_expr does.
    pub(super) fn sublink(
        &mut self,
        kind: SubKind,
        test: Option<&TExpr>,
        op: &str,
        query: &TQuery,
        e: &TExpr,
    ) -> String {
        let mut out = String::from(if kind == SubKind::Array { "ARRAY(" } else { "(" });
        if let Some(test) = test {
            out.push_str(&self.print(test, Some(e), true));
        }
        let parenthesized = match kind {
            SubKind::Exists => {
                out.push_str("EXISTS ");
                true
            }
            SubKind::Any => {
                out.push_str(&if op == "=" { " IN ".to_string() } else { format!(" {op} ANY ") });
                true
            }
            SubKind::All => {
                out.push_str(&format!(" {op} ALL "));
                true
            }
            SubKind::Value | SubKind::Array => false,
        };
        if parenthesized {
            out.push('(');
        }
        self.query(query, &mut out, None, false);
        out.push_str(if parenthesized { "))" } else { ")" });
        out
    }

    /// aggregate prints an aggregate or window function call, as Postgres' get_agg_expr and get_windowfunc_expr do.
    pub(super) fn aggregate(&mut self, agg: &Agg) -> String {
        let mut out = format!("{}(", agg.name);
        if agg.distinct {
            out.push_str("DISTINCT ");
        }
        if agg.star {
            out.push('*');
        } else {
            let args: Vec<String> = agg.args.iter().map(|a| self.print(a, None, true)).collect();
            out.push_str(&args.join(", "));
        }
        if !agg.order.is_empty() {
            out.push_str(" ORDER BY ");
            let order = self.order_list(&agg.order);
            out.push_str(&order);
        }
        if let Some(filter) = &agg.filter {
            out.push_str(") FILTER (WHERE ");
            let text = self.print(filter, None, false);
            out.push_str(&text);
        }
        out.push(')');
        if let Some(window) = &agg.over {
            out.push_str(" OVER ");
            let spec = self.window(window);
            out.push_str(&spec);
        }
        out
    }

    /// window prints a window name or specification, as Postgres' get_rule_windowspec does.
    fn window(&mut self, window: &Window) -> String {
        let spec = match window {
            Window::Named(name) => return quote_identifier(name),
            Window::Spec(spec) => spec,
        };
        let mut parts = Vec::new();
        if !spec.refname.is_empty() {
            parts.push(quote_identifier(&spec.refname));
        }
        if !spec.partition.is_empty() && spec.refname.is_empty() {
            let items: Vec<String> = spec.partition.iter().map(|e| self.sort_group(e)).collect();
            parts.push(format!("PARTITION BY {}", items.join(", ")));
        }
        if !spec.order.is_empty() {
            parts.push(format!("ORDER BY {}", self.order_list(&spec.order)));
        }
        if spec.frame & FRAME_NONDEFAULT != 0 {
            let mut frame = String::from(if spec.frame & FRAME_RANGE != 0 {
                "RANGE "
            } else if spec.frame & FRAME_ROWS != 0 {
                "ROWS "
            } else {
                "GROUPS "
            });
            if spec.frame & FRAME_BETWEEN != 0 {
                frame.push_str("BETWEEN ");
            }
            if spec.frame & FRAME_START_UNBOUNDED_PRECEDING != 0 {
                frame.push_str("UNBOUNDED PRECEDING ");
            } else if spec.frame & FRAME_START_CURRENT_ROW != 0 {
                frame.push_str("CURRENT ROW ");
            } else if let Some(start) = &spec.start {
                frame.push_str(&self.print(start, None, false));
                let preceding = spec.frame & FRAME_START_OFFSET_PRECEDING != 0;
                frame.push_str(if preceding { " PRECEDING " } else { " FOLLOWING " });
            }
            if spec.frame & FRAME_BETWEEN != 0 {
                frame.push_str("AND ");
                if spec.frame & FRAME_END_UNBOUNDED_FOLLOWING != 0 {
                    frame.push_str("UNBOUNDED FOLLOWING ");
                } else if spec.frame & FRAME_END_CURRENT_ROW != 0 {
                    frame.push_str("CURRENT ROW ");
                } else if let Some(end) = &spec.end {
                    frame.push_str(&self.print(end, None, false));
                    let preceding = spec.frame & FRAME_END_OFFSET_PRECEDING != 0;
                    frame.push_str(if preceding { " PRECEDING " } else { " FOLLOWING " });
                }
            }
            if spec.frame & FRAME_EXCLUDE_CURRENT_ROW != 0 {
                frame.push_str("EXCLUDE CURRENT ROW ");
            } else if spec.frame & FRAME_EXCLUDE_GROUP != 0 {
                frame.push_str("EXCLUDE GROUP ");
            } else if spec.frame & FRAME_EXCLUDE_TIES != 0 {
                frame.push_str("EXCLUDE TIES ");
            }
            frame.pop();
            parts.push(frame);
        }
        format!("({})", parts.join(" "))
    }

    /// syntax prints a call of a function that the SQL standard spells with keywords, as Postgres' get_func_sql_syntax
    /// does.
    pub(super) fn syntax(&mut self, name: &str, args: &[TExpr], e: &TExpr) -> String {
        let printed: Vec<String> = args.iter().map(|a| self.print(a, Some(e), false)).collect();
        match (name, printed.as_slice()) {
            ("extract" | "date_part", [_, source]) => {
                let field = match &args[0] {
                    TExpr::Const(Some(text), _) => text.clone(),
                    other => self.print(other, Some(e), false),
                };
                format!("EXTRACT({field} FROM {source})")
            }
            ("substring", [text, from]) => format!("SUBSTRING({text} FROM {from})"),
            ("substring", [text, from, count]) if args[1].ty().oid != oid::TEXT => {
                format!("SUBSTRING({text} FROM {from} FOR {count})")
            }
            ("substring", [text, pattern, escape]) => format!("SUBSTRING({text} SIMILAR {pattern} ESCAPE {escape})"),
            ("btrim" | "ltrim" | "rtrim", [text, rest @ ..]) => {
                let side = match name {
                    "btrim" => "BOTH",
                    "ltrim" => "LEADING",
                    _ => "TRAILING",
                };
                let characters = rest.first().map(|c| format!(" {c}")).unwrap_or_default();
                format!("TRIM({side}{characters} FROM {text})")
            }
            ("position", [text, search]) => format!("POSITION(({search}) IN ({text}))"),
            ("overlay", [text, placing, from, rest @ ..]) => {
                let count = rest.first().map(|c| format!(" FOR {c}")).unwrap_or_default();
                format!("OVERLAY({text} PLACING {placing} FROM {from}{count})")
            }
            ("timezone", [_, _]) => {
                let value = self.print_paren(&args[1], e, false, true);
                let zone = self.print_paren(&args[0], e, false, false);
                format!("({value} AT TIME ZONE {zone})")
            }
            _ => format!("{name}({})", printed.join(", ")),
        }
    }
}
