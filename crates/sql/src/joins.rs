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

//! Join planning: which input of each join to read first and how to find each row's matches, chosen by rules from
//! row estimates that the tables' index trees give, along with the `lookup_join` hints that go-mysql-server takes.

use crate::catalog::table::{HIDDEN_BASE, TableDef};
use crate::expr::{CmpOp, Expr, Scope};
use crate::indexscan::IndexScan;
use crate::plan::{JoinKind, JoinMethod, Plan};
use crate::query::Ctx;

/// SEEK is the cost of finding a key in an index, where reading one row in order costs one.
pub(crate) const SEEK: f64 = 4.0;

/// HASH_ROW is the cost of adding a row to a hash join's table.
const HASH_ROW: f64 = 1.5;

/// HASH_SETUP is the cost of building a hash join's table at all.
const HASH_SETUP: f64 = 20.0;

/// COMPARE is the cost of testing one pair of rows in a nested loop.
const COMPARE: f64 = 1.0;

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

/// JOIN_COLLAPSE_LIMIT is the most inputs of a tree of inner joins whose order the planner searches, as Postgres'
/// join_collapse_limit is.
const JOIN_COLLAPSE_LIMIT: usize = 8;

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

/// plan_joins chooses how each join under a plan runs, from its inputs up: which input it reads first and whether it
/// looks rows up in an index, hashes an input, or compares every pair.
pub(crate) fn plan_joins(ctx: &mut Ctx<'_>, plan: Plan) -> Plan {
    plan_limited(ctx, plan, None)
}

/// plan_limited is `plan_joins` for a plan of which a LIMIT above reads at most this many rows, when it is known.
fn plan_limited(ctx: &mut Ctx<'_>, plan: Plan, limit: Option<f64>) -> Plan {
    if limit.is_none()
        && let Some(reordered) = reorder(ctx, &plan)
    {
        return reordered;
    }
    match plan {
        Plan::Join { left, right, kind, condition, lateral, method } => {
            let ordered = match method {
                JoinMethod::Unplanned => false,
                JoinMethod::Ordered => true,
                method => {
                    return Plan::Join {
                        left: whole(ctx, *left),
                        right: whole(ctx, *right),
                        kind,
                        condition,
                        lateral,
                        method,
                    };
                }
            };
            let streamed = (kind == JoinKind::Inner).then_some(limit).flatten();
            let left = Box::new(plan_limited(ctx, *left, streamed));
            let right = whole(ctx, *right);
            match lateral || !matches!(kind, JoinKind::Inner | JoinKind::Left | JoinKind::Anti | JoinKind::Semi) {
                true => Plan::Join { left, right, kind, condition, lateral, method: JoinMethod::Unplanned },
                false => choose(ctx, *left, *right, kind, condition, ordered, streamed),
            }
        }
        Plan::Limit { input, limit: Some(Expr::Const(crate::types::Value::Int8(n))), offset: None } => Plan::Limit {
            input: Box::new(plan_limited(ctx, *input, Some(n as f64))),
            limit: Some(Expr::Const(crate::types::Value::Int8(n))),
            offset: None,
        },
        Plan::Filter { input, predicate } => {
            Plan::Filter { input: Box::new(plan_limited(ctx, *input, limit)), predicate }
        }
        Plan::Project { input, exprs } => Plan::Project { input: Box::new(plan_limited(ctx, *input, limit)), exprs },
        Plan::Sort { input, keys } => Plan::Sort { input: whole(ctx, *input), keys },
        Plan::Distinct { input, keys } => Plan::Distinct { input: whole(ctx, *input), keys },
        Plan::Limit { input, limit, offset } => Plan::Limit { input: whole(ctx, *input), limit, offset },
        Plan::Aggregate { input, groups, aggregates, sets } => {
            Plan::Aggregate { input: whole(ctx, *input), groups, aggregates, sets }
        }
        Plan::Window { input, calls } => Plan::Window { input: whole(ctx, *input), calls },
        Plan::ProjectSet { input, functions, dropped } => {
            Plan::ProjectSet { input: whole(ctx, *input), functions, dropped }
        }
        Plan::Once(input) => Plan::Once(whole(ctx, *input)),
        other => other,
    }
}

/// reorder plans a tree of three or more inner joins in the order that costs least, as Postgres' join search does by
/// dynamic programming over the sets of its inputs: the cheapest join of each set comes from the cheapest joins of two
/// smaller sets that a condition connects, or of any two when none does. A projection puts the columns back in the
/// tree's order. It returns None when it cannot reorder the tree.
fn reorder(ctx: &mut Ctx<'_>, plan: &Plan) -> Option<Plan> {
    let mut inputs = Vec::new();
    let mut conditions = Vec::new();
    if !flatten(plan, 0, &mut inputs, &mut conditions)
        || !(3..=JOIN_COLLAPSE_LIMIT).contains(&inputs.len())
        || conditions.iter().any(crate::plan::has_subquery)
    {
        return None;
    }
    let starts: Vec<usize> = inputs.iter().map(|(_, start)| *start).collect();
    let widths: Vec<usize> = inputs.iter().map(|(input, _)| input.width()).collect();
    let owner = |column: usize| (0..starts.len()).rev().find(|&i| starts[i] <= column).unwrap_or(0);
    let masks: Vec<u32> = conditions
        .iter()
        .map(|c| {
            let mut mask = 0;
            c.visit(&mut |e| {
                if let Expr::Column(i) = e {
                    mask |= 1 << owner(*i);
                }
            });
            mask
        })
        .collect();
    let full = (1u32 << inputs.len()) - 1;
    let and = |conditions: Vec<Expr>| conditions.into_iter().reduce(|x, y| Expr::And(Box::new(x), Box::new(y)));
    let mut own: Vec<Vec<Expr>> = vec![Vec::new(); inputs.len()];
    let mut constant = Vec::new();
    let (mut joining, mut joining_masks) = (Vec::new(), Vec::new());
    for (c, m) in conditions.into_iter().zip(masks) {
        match m.count_ones() {
            0 => constant.push(c),
            1 => {
                let input = m.trailing_zeros() as usize;
                let local: Vec<usize> =
                    (0..starts[input] + widths[input]).map(|i| i.saturating_sub(starts[input])).collect();
                own[input].push(renumber(c, &local));
            }
            _ => {
                joining.push(c);
                joining_masks.push(m);
            }
        }
    }
    let (conditions, masks) = (joining, joining_masks);
    let mut best: std::collections::HashMap<u32, (f64, Plan, Vec<usize>)> = std::collections::HashMap::new();
    for (i, ((input, _), own)) in inputs.into_iter().zip(own).enumerate() {
        let input = match and(own) {
            Some(predicate) => crate::plan::push_down(input, predicate),
            None => input,
        };
        let planned = *whole(ctx, input);
        best.insert(1 << i, (cost(ctx, &planned), planned, vec![i]));
    }
    for size in 2..=widths.len() as u32 {
        for set in (1..=full).filter(|s: &u32| s.count_ones() == size) {
            let splits: Vec<(u32, u32)> = (1..set)
                .filter(|&part| part & set == part && part < set ^ part)
                .map(|part| (part, set ^ part))
                .filter(|(a, b)| best.contains_key(a) && best.contains_key(b))
                .collect();
            let connected = |a: u32, b: u32| masks.iter().any(|&m| m & a != 0 && m & b != 0 && m & !(a | b) == 0);
            let any_connected = splits.iter().any(|&(a, b)| connected(a, b));
            for (a, b) in splits {
                if any_connected && !connected(a, b) {
                    continue;
                }
                let layout: Vec<usize> = best[&a].2.iter().chain(&best[&b].2).copied().collect();
                let position = positions(&layout, &starts, &widths);
                let joined = and(conditions
                    .iter()
                    .zip(&masks)
                    .filter(|&(_, &m)| m & set == m && m & a != m && m & b != m)
                    .map(|(c, _)| renumber(c.clone(), &position))
                    .collect());
                let (left, right) = (best[&a].1.clone(), best[&b].1.clone());
                let planned = choose(ctx, left, right, JoinKind::Inner, joined, false, None);
                let spent = cost(ctx, &planned);
                if best.get(&set).is_none_or(|(c, ..)| spent < *c) {
                    best.insert(set, (spent, planned, layout));
                }
            }
        }
    }
    let (_, planned, layout) = best.remove(&full)?;
    let position = positions(&layout, &starts, &widths);
    let exprs = (0..position.len()).map(|column| Expr::Column(position[column])).collect();
    let planned = match and(constant) {
        Some(predicate) => Plan::Filter { input: Box::new(planned), predicate },
        None => planned,
    };
    Some(Plan::Project { input: Box::new(planned), exprs })
}

/// flatten gathers the inputs of a tree of inner joins that `reorder` can reorder, with the column of the tree's row
/// that each one's columns start at, and the conjuncts of the joins' conditions over the tree's row. It reports
/// whether the plan is such a join at all.
fn flatten(plan: &Plan, start: usize, inputs: &mut Vec<(Plan, usize)>, conditions: &mut Vec<Expr>) -> bool {
    match plan {
        Plan::Join { left, right, kind: JoinKind::Inner, condition, lateral: false, method: JoinMethod::Unplanned } => {
            let width = left.width();
            if !flatten(left, start, inputs, conditions) {
                inputs.push(((**left).clone(), start));
            }
            if !flatten(right, start + width, inputs, conditions) {
                inputs.push(((**right).clone(), start + width));
            }
            for c in condition.iter().flat_map(crate::indexscan::conjuncts) {
                conditions.push(renumber(c.clone(), &(0..start + plan.width()).map(|i| i + start).collect::<Vec<_>>()));
            }
            true
        }
        _ => false,
    }
}

/// positions returns where each column of a tree's row lies in the row of its inputs joined in a layout's order.
fn positions(layout: &[usize], starts: &[usize], widths: &[usize]) -> Vec<usize> {
    let mut position = vec![0; starts.iter().zip(widths).map(|(s, w)| s + w).max().unwrap_or(0)];
    let mut next = 0;
    for &input in layout {
        for column in 0..widths[input] {
            position[starts[input] + column] = next;
            next += 1;
        }
    }
    position
}

/// renumber rewrites the columns that an expression reads by a mapping from old column to new.
fn renumber(e: Expr, position: &[usize]) -> Expr {
    match e {
        Expr::Column(i) => Expr::Column(position.get(i).copied().unwrap_or(i)),
        other => other.map_children(&mut |c| renumber(c, position)),
    }
}

/// cost returns about how much work a planned plan takes, in the units of the join costs: one for each row read in
/// order.
fn cost(ctx: &mut Ctx<'_>, plan: &Plan) -> f64 {
    match plan {
        Plan::Join { left, right, method, .. } => {
            let (l, rows) = (estimate(ctx, left), estimate(ctx, plan));
            let rest = match method {
                JoinMethod::Lookup { .. } | JoinMethod::CatalogLookup { .. } => l * SEEK + rows,
                JoinMethod::Hash => {
                    let r = estimate(ctx, right);
                    cost(ctx, right) + l + r * HASH_ROW + HASH_SETUP
                }
                _ => {
                    let r = estimate(ctx, right);
                    cost(ctx, right) + l * r * COMPARE
                }
            };
            cost(ctx, left) + rest
        }
        Plan::Project { input, .. } | Plan::Filter { input, .. } => cost(ctx, input),
        other => estimate(ctx, other),
    }
}

/// whole is `plan_joins` for an input that the plan above it reads every row of.
fn whole(ctx: &mut Ctx<'_>, input: Plan) -> Box<Plan> {
    Box::new(plan_limited(ctx, input, None))
}

/// choose plans a join of two inputs: a lookup in the right input's table when that costs least, else a hash join
/// when the condition has equalities, else a nested loop. An inner join may read its right input first, which puts
/// the smaller input on the side it hashes or loops over, unless its left input's order replaced a sort. A LIMIT
/// above an inner join reads only `limit` of its rows, so the input that the join reads first, and looks up or probes
/// with, stops early, while a hashed input is read whole. On equal estimates, a table with a primary key is read
/// first.
fn choose(
    ctx: &mut Ctx<'_>,
    left: Plan,
    right: Plan,
    kind: JoinKind,
    condition: Option<Expr>,
    ordered: bool,
    limit: Option<f64>,
) -> Plan {
    let swappable = kind == JoinKind::Inner && !ordered && !condition.as_ref().is_some_and(crate::plan::has_subquery);
    let (l_all, r_all) = (estimate(ctx, &left), estimate(ctx, &right));
    let first = |rows: f64, matches: f64| rows.min(limit.map_or(f64::INFINITY, |n| n / matches.max(f64::MIN_POSITIVE)));
    let (l, r) = (first(l_all, 1.0), first(r_all, 1.0));
    let Some(condition) = condition else {
        let swap = swappable && reads_first(r_all, &right, l_all, &left);
        return join(left, right, kind, None, JoinMethod::NestedLoop, swap);
    };
    let width = left.width();
    let mut best: Option<(f64, bool, JoinMethod)> = None;
    let mut consider = |cost: f64, swap: bool, method: JoinMethod| {
        if best.as_ref().is_none_or(|(b, ..)| cost < *b) {
            best = Some((cost, swap, method));
        }
    };
    if !crate::plan::has_subquery(&condition) {
        if let Some(found) = lookup(ctx, &right, &condition, width) {
            consider(first(l_all, found.matches) * (SEEK + found.matches), false, found.method);
        }
        if swappable {
            let flipped = flip(&condition, width, right.width());
            if let Some(found) = lookup(ctx, &left, &flipped, right.width()) {
                consider(first(r_all, found.matches) * (SEEK + found.matches), true, found.method);
            }
        }
    }
    let keyed = !crate::plan::join_keys(&condition, width).0.is_empty();
    let swap = swappable && reads_first(r_all, &right, l_all, &left);
    let (probed, built) = if swap { (r, l_all) } else { (l, r_all) };
    if keyed {
        consider(probed + built + built * HASH_ROW + HASH_SETUP, swap, JoinMethod::Hash);
    }
    consider(probed * built * COMPARE + probed + built, swap, JoinMethod::NestedLoop);
    let (_, swap, method) = best.expect("a nested loop is always possible");
    join(left, right, kind, Some(condition), method, swap)
}

/// reads_first reports whether a join should read the input with the first estimate before the other one: when it
/// is larger, or as large and a table with a primary key while the other is not.
fn reads_first(rows: f64, plan: &Plan, other_rows: f64, other: &Plan) -> bool {
    let keyed = |plan: &Plan| target(plan).is_some_and(|t| !t.keyless());
    rows > other_rows || (rows == other_rows && keyed(plan) && !keyed(other))
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

/// filter_existence filters a plan's rows by a WHERE condition `EXISTS (subquery)` or `NOT EXISTS (subquery)` as a
/// semi or anti join, when the subquery filters one relation by conditions that read only its own row and the plan's
/// row, with at least one of them reading the plan's row, and by the condition itself otherwise.
pub(crate) fn filter_existence(plan: Plan, condition: Expr) -> Plan {
    match decorrelate(&condition, plan.width()) {
        Some((right, kind, joined)) => Plan::Join {
            left: Box::new(plan),
            right: Box::new(right),
            kind,
            condition: Some(joined),
            lateral: false,
            method: JoinMethod::Unplanned,
        },
        None => Plan::Filter { input: Box::new(plan), predicate: condition },
    }
}

/// decorrelate returns the right input, kind, and condition of the join that `filter_existence` makes of a condition
/// over rows this wide, when it can make one, as Postgres' convert_EXISTS_sublink_to_join does: the subquery's rows
/// below its WHERE filters, which must read nothing of the rows outside the subquery, become the right input, and the
/// filters' conditions that read the enclosing row, even from inside a nested subquery, become the join's condition.
fn decorrelate(condition: &Expr, width: usize) -> Option<(Plan, JoinKind, Expr)> {
    if let Some(right) = any_input(condition) {
        let Expr::AnySubquery(comparison, ..) = condition else { return None };
        return Some((right.clone(), JoinKind::Semi, with_subquery_value((**comparison).clone(), width)));
    }
    let (subquery, kind) = match condition {
        Expr::Exists(subquery) => (subquery, JoinKind::Semi),
        Expr::Not(inner) => match &**inner {
            Expr::Exists(subquery) => (subquery, JoinKind::Anti),
            _ => return None,
        },
        _ => return None,
    };
    let mut input = &**subquery;
    while let Plan::Project { input: inner, .. } | Plan::Once(inner) = input {
        input = inner;
    }
    let mut predicates = Vec::new();
    while let Plan::Filter { input: inner, predicate } = input {
        predicates.push(predicate);
        input = match &**inner {
            Plan::Once(inner) => inner,
            other => other,
        };
    }
    if predicates.is_empty() || plan_lowest_level(input)? < 0 {
        return None;
    }
    let (mut own, mut correlated) = (Vec::new(), Vec::new());
    for c in predicates.into_iter().flat_map(crate::indexscan::conjuncts) {
        match lowest_level(c, 0)? {
            ..-1 => return None,
            -1 => correlated.push(rebase(c.clone(), width, 0)),
            _ => own.push(c.clone()),
        }
    }
    let and = |conditions: Vec<Expr>| conditions.into_iter().reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let joined = and(correlated)?;
    let right = match and(own) {
        Some(predicate) => Plan::Filter { input: Box::new(input.clone()), predicate },
        None => input.clone(),
    };
    Some((right, kind, joined))
}

/// any_input returns the rows of an `= ANY` (IN) subquery that a semi join can read in its place, as Postgres'
/// convert_ANY_sublink_to_join does: the one column of a subquery that reads nothing outside it, compared by a test
/// without subqueries of its own.
pub(crate) fn any_input(condition: &Expr) -> Option<&Plan> {
    let Expr::AnySubquery(comparison, subquery, false) = condition else { return None };
    let input = match &**subquery {
        Plan::Once(inner) => &**inner,
        other => other,
    };
    (input.width() == 1 && !crate::plan::has_subquery(comparison) && plan_lowest_level(input)? >= 0).then_some(input)
}

/// with_subquery_value rewrites the comparison of an `= ANY` test to read the subquery's value from the column of a
/// join's row that follows the enclosing row's columns, which are this wide.
fn with_subquery_value(e: Expr, width: usize) -> Expr {
    match e {
        Expr::SubqueryValue => Expr::Column(width),
        other => other.map_children(&mut |c| with_subquery_value(c, width)),
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
        Expr::Exists(p) | Expr::Scalar(p) | Expr::ArraySubquery(p, _) | Expr::AnySubquery(_, p, _) => {
            let mut plan = (**p).clone();
            let reachable = plan.map_exprs(0, &mut |e, depth| {
                match lowest_level(&e, nesting + 1 + depth as i64) {
                    Some(level) => lowest = lowest.min(level),
                    None => known = false,
                }
                e
            });
            known &= reachable;
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

/// rebase rewrites an expression of a subquery, `nesting` subqueries deep within it, to read the row of the join that
/// its pulled-up rows make with the enclosing rows, which come first in the join's row and are this wide.
fn rebase(e: Expr, width: usize, nesting: usize) -> Expr {
    let level = |d: usize| nesting as i64 - d as i64;
    let e = match e {
        Expr::Column(i) if nesting == 0 => return Expr::Column(i + width),
        Expr::Outer(d, i) if level(d) == 0 => return Expr::Outer(d, i + width),
        Expr::Outer(d, i) if level(d) == -1 && nesting == 0 => return Expr::Column(i),
        Expr::Outer(d, i) if level(d) == -1 => return Expr::Outer(nesting, i),
        Expr::Outer(d, i) if level(d) < -1 => return Expr::Outer(d - 1, i),
        other => other,
    };
    let mut e = e.map_children(&mut |c| rebase(c, width, nesting));
    if let Expr::Exists(p) | Expr::Scalar(p) | Expr::ArraySubquery(p, _) | Expr::AnySubquery(_, p, _) = &mut e {
        p.map_exprs(0, &mut |x, depth| rebase(x, width, nesting + 1 + depth));
    }
    e
}
