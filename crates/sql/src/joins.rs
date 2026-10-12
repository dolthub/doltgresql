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

//! Join helpers for the planner: row estimates of plans that the tables' index trees give, the index lookups that a
//! join can make into a table or system catalog, and the `lookup_join` hints that go-mysql-server takes.

use crate::catalog::table::{HIDDEN_BASE, TableDef};
use crate::expr::{CmpOp, Expr, Scope};
use crate::indexscan::IndexScan;
use crate::plan::{JoinKind, JoinMethod, Plan};
use crate::query::Ctx;

/// SEEK is the cost of finding a key in an index, where reading one row in order costs one.
pub(crate) const SEEK: f64 = 4.0;

/// UNKNOWN_ROWS is the estimate for an input of unknown size, as Postgres assumes for a set-returning function.
const UNKNOWN_ROWS: f64 = 1000.0;

/// SMALL_CATALOG_ROWS is the estimate for a system catalog that holds a few rows, such as pg_namespace.
const SMALL_CATALOG_ROWS: f64 = 10.0;

/// EQUALITY_SELECTIVITY is the share of rows that an equality keeps, or that one key value finds in an index without
/// statistics, as Postgres' DEFAULT_EQ_SEL gives.
const EQUALITY_SELECTIVITY: f64 = 0.005;

/// RANGE_SELECTIVITY is the share of rows that an inequality keeps, as Postgres' DEFAULT_INEQ_SEL gives.
const RANGE_SELECTIVITY: f64 = 1.0 / 3.0;

/// OTHER_SELECTIVITY is the share of rows that any other condition keeps.
const OTHER_SELECTIVITY: f64 = 0.5;

/// DEFAULT_DISTINCT is how many distinct values a join key takes without statistics, as Postgres'
/// DEFAULT_NUM_DISTINCT assumes.
const DEFAULT_DISTINCT: f64 = 200.0;

/// estimate returns about how many rows a plan produces.
pub(crate) fn estimate(ctx: &mut Ctx<'_>, plan: &Plan) -> f64 {
    match plan {
        Plan::OneRow => 1.0,
        Plan::Scan(table, _) => table_rows(table),
        Plan::IndexScan(scan) => scan.estimate(ctx).unwrap_or(UNKNOWN_ROWS),
        Plan::CatalogIndexScan(scan) => catalog_scan_rows(scan, estimate(ctx, &Plan::Catalog(scan.table))),
        Plan::Values(rows) => rows.len() as f64,
        Plan::CteScan(def) => match def.shared.get() {
            Some((_, rows)) => *rows,
            None => estimate(ctx, &def.plan),
        },
        Plan::Catalog(table) => match table.name {
            "pg_namespace" | "pg_database" | "pg_tablespace" | "pg_am" | "pg_authid" | "pg_roles" => SMALL_CATALOG_ROWS,
            _ => UNKNOWN_ROWS,
        },
        Plan::Filter { input, predicate } => {
            let rows = estimate(ctx, input);
            match &**input {
                Plan::Scan(table, _) if let Some(stats) = crate::colstats::table_stats(ctx, table) => {
                    rows * crate::colstats::selectivity(&stats, predicate)
                }
                _ => rows * selectivity(predicate),
            }
        }
        Plan::Project { input, .. }
        | Plan::Sort { input, .. }
        | Plan::Window { input, .. }
        | Plan::ProjectSet { input, .. }
        | Plan::Once(input) => estimate(ctx, input),
        Plan::Limit { input, limit: Some(Expr::Const(crate::types::Value::Int8(n))), .. } => {
            estimate(ctx, input).min(*n as f64)
        }
        Plan::Limit { input, .. } => estimate(ctx, input),
        Plan::Aggregate { groups, .. } if groups.is_empty() => 1.0,
        Plan::Aggregate { input, .. } | Plan::Distinct { input, .. } => (estimate(ctx, input) / 10.0).max(1.0),
        Plan::Join { left, right, kind, condition, .. } => {
            let (l, r) = (estimate(ctx, left), estimate(ctx, right));
            let (left_keys, right_keys) =
                condition.as_ref().map_or_else(Default::default, |c| crate::plan::join_keys(c, left.width()));
            let joined = match left_keys.is_empty() {
                false => {
                    let distinct = key_distinct(ctx, left, &left_keys, l).max(key_distinct(ctx, right, &right_keys, r));
                    l * r / distinct
                }
                true => l * r * condition.as_ref().map_or(1.0, selectivity),
            };
            match kind {
                JoinKind::Inner => joined,
                JoinKind::Left => joined.max(l),
                JoinKind::Right => joined.max(r),
                JoinKind::Full => joined.max(l).max(r),
                JoinKind::Anti | JoinKind::Semi => l * OTHER_SELECTIVITY,
                JoinKind::RightAnti | JoinKind::RightSemi => r * OTHER_SELECTIVITY,
            }
        }
        _ => UNKNOWN_ROWS,
    }
}

/// key_distinct returns about how many distinct values a join's keys take in an input's rows, as Postgres'
/// eqjoinsel judges them: every row of the input's table when the keys cover one of its unique indexes, the table's
/// rows over what one key finds through an index that starts with a key, which ANALYZE's counts refine, and otherwise
/// the default, never more than the input's rows.
fn key_distinct(ctx: &mut Ctx<'_>, plan: &Plan, keys: &[Expr], rows: f64) -> f64 {
    let column = |k: &Expr| match k {
        Expr::Column(c) => Some(*c),
        Expr::Cast(inner, ..) => match **inner {
            Expr::Column(c) => Some(c),
            _ => None,
        },
        _ => None,
    };
    let columns: Option<Vec<usize>> = keys.iter().map(column).collect();
    let distinct = match (target(plan), columns) {
        (Some(table), Some(columns)) => {
            let total = table_rows(table);
            let covers = |index: &[usize]| index.iter().all(|c| columns.contains(c));
            let unique = (!table.keyless() && covers(&table.key_columns))
                || table.indexes.iter().any(|i| i.unique && i.predicate.is_empty() && covers(&i.columns));
            let leading = |index: &[usize]| index.first().is_some_and(|c| columns.contains(c));
            let index = match leading(&table.key_columns) && !table.keyless() {
                true => Some(None),
                false => table.indexes.iter().position(|i| leading(&i.columns)).map(Some),
            };
            match (unique, index) {
                (true, _) => total,
                (false, Some(index)) => total / per_key(ctx, table, index, total),
                (false, None) => match columns.as_slice() {
                    [column] => crate::colstats::column_distinct(ctx, table, *column).unwrap_or(DEFAULT_DISTINCT),
                    _ => DEFAULT_DISTINCT,
                },
            }
        }
        _ => DEFAULT_DISTINCT,
    };
    distinct.min(rows).max(1.0)
}

/// catalog_scan_rows returns about how many rows a scan of a system catalog relation's index reads, given the
/// relation's estimate: one for each range of single values of a unique index's every column, and otherwise the
/// share of rows that each column's single value or bounds keep, as an equality or inequality keeps them.
fn catalog_scan_rows(scan: &crate::pgcatalog::indexes::CatalogIndexScan, rows: f64) -> f64 {
    use crate::ranges::Cut;
    let point = |c: &crate::ranges::ColumnRange| matches!((&c.lower, &c.upper), (Cut::Below(l), Cut::Above(h)) if crate::expr::compare_values(l, h).is_eq());
    let found: f64 = scan
        .ranges
        .iter()
        .map(|range| match scan.index.unique && range.iter().all(point) {
            true => 1.0,
            false => {
                range
                    .iter()
                    .map(|c| match (point(c), c.is_all()) {
                        (true, _) => EQUALITY_SELECTIVITY,
                        (false, true) => 1.0,
                        (false, false) => RANGE_SELECTIVITY,
                    })
                    .product::<f64>()
                    * rows
            }
        })
        .sum();
    found.min(rows)
}

/// table_rows returns how many rows a table holds, from its primary index's root.
fn table_rows(table: &TableDef) -> f64 {
    prolly::Node::decode(table.table.primary_index.clone()).map_or(UNKNOWN_ROWS, |root| root.tree_count() as f64)
}

/// selectivity returns about what share of rows a condition keeps.
fn selectivity(condition: &Expr) -> f64 {
    crate::indexscan::conjuncts(condition)
        .into_iter()
        .map(|c| match c {
            Expr::Compare(CmpOp::Eq, ..) | Expr::IsNull(_, false) => EQUALITY_SELECTIVITY,
            Expr::Compare(CmpOp::Lt | CmpOp::Le | CmpOp::Gt | CmpOp::Ge, ..) => RANGE_SELECTIVITY,
            _ => OTHER_SELECTIVITY,
        })
        .product()
}

/// target returns the table that a join's right input reads, when the join can look rows up in it: a scan of the
/// table, or an index scan of it whose columns are all table columns, either one maybe under a filter.
fn target(right: &Plan) -> Option<&TableDef> {
    let input = match right {
        Plan::Filter { input, .. } => input,
        other => other,
    };
    match input {
        Plan::Scan(table, _) => Some(table),
        Plan::IndexScan(scan) if scan.nearest.is_none() && scan.index_columns().iter().all(|&c| c < HIDDEN_BASE) => {
            Some(&scan.table)
        }
        _ => None,
    }
}

/// Lookup is how a join can look a left row's matches up in an index of the right input's table or catalog.
pub(crate) struct Lookup {
    pub(crate) method: JoinMethod,
    /// About how many right rows each lookup finds.
    pub(crate) matches: f64,
}

/// lookup returns the index of the right input's table or catalog that a join's equalities let it look rows up in
/// most cheaply, with the left expressions that give its first columns, when the right input allows lookups at all.
pub(crate) fn lookup(ctx: &mut Ctx<'_>, right: &Plan, condition: &Expr, left_width: usize) -> Option<Lookup> {
    let (left_keys, right_keys) = crate::plan::join_keys(condition, left_width);
    if let Some(catalog) = catalog_target(right) {
        return catalog_lookup(catalog, &left_keys, &right_keys);
    }
    let table = target(right)?;
    let mut indexes: Vec<Option<usize>> = Vec::new();
    if !table.keyless() {
        indexes.push(None);
    }
    indexes.extend(
        (0..table.indexes.len())
            .filter(|&i| table.indexes[i].vector.is_none() && table.indexes[i].predicate.is_empty())
            .map(Some),
    );
    let rows = table_rows(table);
    let mut best: Option<(f64, Vec<Expr>, Option<usize>)> = None;
    for index in indexes {
        let (columns, unique, descending) = match index {
            Some(i) => (&table.indexes[i].columns, table.indexes[i].unique, &table.indexes[i].descending),
            None => (&table.key_columns, true, &Vec::new()),
        };
        let mut keys = Vec::new();
        for (position, &c) in columns.iter().enumerate() {
            let Some(column) = table.index_column(c) else { break };
            if descending.get(position).copied().unwrap_or(false)
                || crate::storage::is_adaptive(column.encoding)
                || !crate::exec::lookup_type(column.ty.oid)
            {
                break;
            }
            match right_keys.iter().position(|k| *k == Expr::Column(c)) {
                Some(k) => keys.push(left_keys[k].clone()),
                None => break,
            }
        }
        if keys.is_empty() {
            continue;
        }
        let matches = match unique && keys.len() == columns.len() {
            true => 1.0,
            false => per_key(ctx, table, index, rows),
        };
        if best.as_ref().is_none_or(|(b, k, _)| matches < *b || (matches == *b && keys.len() > k.len())) {
            best = Some((matches, keys, index));
        }
    }
    let (matches, keys, index) = best?;
    let scan = IndexScan {
        table: std::sync::Arc::new(table.clone()),
        index,
        ranges: Vec::new(),
        reverse: false,
        nearest: None,
        needed: None,
        lookup_heavy: None,
        parameterized: None,
    };
    Some(Lookup { method: JoinMethod::Lookup { scan: Box::new(scan), keys }, matches })
}

/// catalog_target returns the system catalog relation that a join's right input reads, maybe through one of its
/// indexes or under a filter.
fn catalog_target(right: &Plan) -> Option<&'static crate::pgcatalog::CatalogTable> {
    let input = match right {
        Plan::Filter { input, .. } => input,
        other => other,
    };
    match input {
        Plan::Catalog(table) => Some(table),
        Plan::CatalogIndexScan(scan) => Some(scan.table),
        _ => None,
    }
}

/// catalog_lookup returns the index of a system catalog relation whose first columns the right sides of a join's
/// equalities give the most of, with the left sides that give them.
fn catalog_lookup(
    catalog: &'static crate::pgcatalog::CatalogTable,
    left_keys: &[Expr],
    right_keys: &[Expr],
) -> Option<Lookup> {
    let mut best: Option<(&'static crate::pgcatalog::indexes::CatalogIndex, Vec<Expr>)> = None;
    for index in crate::pgcatalog::indexes::indexes(catalog) {
        let mut keys = Vec::new();
        for c in crate::pgcatalog::indexes::key_columns(catalog, index) {
            match right_keys.iter().position(|k| *k == Expr::Column(c)) {
                Some(k) => keys.push(left_keys[k].clone()),
                None => break,
            }
        }
        let full = |index: &crate::pgcatalog::indexes::CatalogIndex, keys: &[Expr]| {
            index.unique && keys.len() == index.columns.len()
        };
        let better = best.as_ref().is_none_or(|(b, k)| (full(index, &keys) && !full(b, k)) || keys.len() > k.len());
        if !keys.is_empty() && better {
            best = Some((index, keys));
        }
    }
    let (index, keys) = best?;
    let matches =
        if index.unique && keys.len() == index.columns.len() { 1.0 } else { UNKNOWN_ROWS * EQUALITY_SELECTIVITY };
    Some(Lookup { method: JoinMethod::CatalogLookup { index, keys }, matches })
}

/// per_key returns about how many rows of a table one value of an index's key finds: the rows over the distinct keys
/// that ANALYZE counted, or the share of rows an equality keeps without statistics.
fn per_key(ctx: &mut Ctx<'_>, table: &TableDef, index: Option<usize>, rows: f64) -> f64 {
    let name = match index {
        Some(i) => table.indexes[i].name.clone(),
        None => table.primary_name(),
    };
    let (database, branch) = (ctx.session.database.clone(), ctx.session.branch.clone());
    let distinct: u64 = ctx
        .session
        .engine
        .statistics(&database, &branch)
        .iter()
        .find(|s| s.schema == table.schema && s.table == table.name && s.index == name)
        .map_or(0, |s| s.buckets.iter().map(|b| b.distinct).sum());
    match distinct {
        0 => (rows * EQUALITY_SELECTIVITY).max(1.0),
        distinct => (rows / distinct as f64).max(1.0),
    }
}

/// join returns the join of two inputs by a method, reading the right input first when `swap` is set, under a
/// projection that puts the columns back in the order of the left input's then the right input's.
fn join(left: Plan, right: Plan, kind: JoinKind, condition: Option<Expr>, method: JoinMethod, swap: bool) -> Plan {
    if !swap {
        return Plan::Join { left: Box::new(left), right: Box::new(right), kind, condition, lateral: false, method };
    }
    let (left_width, right_width) = (left.width(), right.width());
    let condition = condition.map(|c| flip(&c, left_width, right_width));
    let exprs = (right_width..right_width + left_width).chain(0..right_width).map(Expr::Column).collect();
    let joined = Plan::Join { left: Box::new(right), right: Box::new(left), kind, condition, lateral: false, method };
    Plan::Project { input: Box::new(joined), exprs }
}

/// flip rewrites a condition over a left row followed by a right row of the widths to read the right row first.
fn flip(condition: &Expr, left_width: usize, right_width: usize) -> Expr {
    match condition {
        Expr::Column(i) if *i < left_width => Expr::Column(i + right_width),
        Expr::Column(i) => Expr::Column(i - left_width),
        other => other.clone().map_children(&mut |e| flip(&e, left_width, right_width)),
    }
}

/// hints returns the `lookup_join(outer, inner)` hints of a statement's `/*+ ... */` comment, as go-mysql-server reads
/// them, with each pair's names in lower case.
pub(crate) fn hints(source: &str) -> Vec<(String, String)> {
    let Some(start) = source.find("/*+") else { return Vec::new() };
    let Some(end) = source[start..].find("*/") else { return Vec::new() };
    let mut text = &source[start + 3..start + end];
    let mut found = Vec::new();
    while let Some(at) = text.to_lowercase().find("lookup_join(") {
        let rest = &text[at + "lookup_join(".len()..];
        let Some(close) = rest.find(')') else { break };
        let names: Vec<String> = rest[..close].split(',').map(|n| n.trim().to_lowercase()).collect();
        if let [outer, inner] = names.as_slice() {
            found.push((outer.clone(), inner.clone()));
        }
        text = &rest[close..];
    }
    found
}

/// apply_hints makes each join whose inputs a `lookup_join(outer, inner)` hint names read the outer input first and
/// look its matches up in the inner input's table, when the condition lets it, given the scope of the plan's columns,
/// which names the table of each.
pub(crate) fn apply_hints(ctx: &mut Ctx<'_>, plan: Plan, scope: &Scope, hints: &[(String, String)]) -> Plan {
    hinted(ctx, plan, &scope.columns.iter().map(|c| c.table.to_lowercase()).collect::<Vec<_>>(), hints)
}

/// hinted is `apply_hints` with the table of each of the plan's columns.
fn hinted(ctx: &mut Ctx<'_>, plan: Plan, tables: &[String], hints: &[(String, String)]) -> Plan {
    match plan {
        Plan::Filter { input, predicate } => {
            Plan::Filter { input: Box::new(hinted(ctx, *input, tables, hints)), predicate }
        }
        Plan::Join { left, right, kind: JoinKind::Inner, condition: Some(condition), lateral: false, method } => {
            let width = left.width().min(tables.len());
            let (left_tables, right_tables) = tables.split_at(width);
            let left = hinted(ctx, *left, left_tables, hints);
            let right = hinted(ctx, *right, right_tables, hints);
            let names = |tables: &[String], name: &str| tables.first().is_some_and(|t| t == name);
            for (outer, inner) in hints {
                let swap = match (names(left_tables, outer), names(right_tables, inner)) {
                    (true, true) => false,
                    _ if names(right_tables, outer) && names(left_tables, inner) => true,
                    _ => continue,
                };
                let (from, into) = if swap { (&right, &left) } else { (&left, &right) };
                let keys = if swap { flip(&condition, left.width(), right.width()) } else { condition.clone() };
                if let Some(found) = lookup(ctx, into, &keys, from.width()) {
                    return join(left, right, JoinKind::Inner, Some(condition), found.method, swap);
                }
            }
            Plan::Join {
                left: Box::new(left),
                right: Box::new(right),
                kind: JoinKind::Inner,
                condition: Some(condition),
                lateral: false,
                method,
            }
        }
        other => other,
    }
}

/// lowest_level returns the outermost row that an expression of a subquery reads, counting the subquery's own row as
/// level 0 and its enclosing row as -1, where the expression sits `nesting` subqueries deep within the subquery and
/// reads its own row as that level. It is i64::MAX for an expression that reads no row, and None when the expression
/// holds a plan whose expressions it cannot see.
fn lowest_level(e: &Expr, nesting: i64) -> Option<i64> {
    let mut lowest = i64::MAX;
    let mut known = true;
    e.visit(&mut |x| match x {
        Expr::Column(_) => lowest = lowest.min(nesting),
        Expr::Outer(d, _) => lowest = lowest.min(nesting - *d as i64),
        x if !x.subqueries().is_empty() => {
            for plan in x.subqueries() {
                let mut plan = plan.clone();
                let reachable = plan.map_exprs(0, &mut |e, depth| {
                    match lowest_level(&e, nesting + 1 + depth as i64) {
                        Some(level) => lowest = lowest.min(level),
                        None => known = false,
                    }
                    e
                });
                known &= reachable;
            }
        }
        _ => {}
    });
    known.then_some(lowest)
}

/// plan_lowest_level returns the outermost row that a subquery's plan reads, as `lowest_level` counts rows.
pub(crate) fn plan_lowest_level(plan: &Plan) -> Option<i64> {
    let (mut lowest, mut known) = (i64::MAX, true);
    let reachable = plan.clone().map_exprs(0, &mut |e, depth| {
        match lowest_level(&e, depth as i64) {
            Some(level) => lowest = lowest.min(level),
            None => known = false,
        }
        e
    });
    (known && reachable).then_some(lowest)
}
