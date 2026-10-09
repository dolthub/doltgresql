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

//! Column statistics as Postgres' ANALYZE gathers them into pg_statistic (analyze.c's compute_scalar_stats), from an
//! evenly spread sample of a table's rows, and the selectivity that Postgres' clausesel.c and selfuncs.c estimate from
//! them.

use std::cell::RefCell;
use std::cmp::Ordering;
use std::collections::HashMap;
use std::sync::Arc;

use crate::catalog::table::TableDef;
use crate::expr::{CmpOp, Expr, compare_values};
use crate::query::Ctx;
use crate::types::Value;

/// SAMPLE_ROWS is how many rows of a table the statistics read.
const SAMPLE_ROWS: usize = 3000;

/// STATISTICS_TARGET is the most common values and histogram bounds that a column keeps, as Postgres'
/// default_statistics_target is.
const STATISTICS_TARGET: usize = 100;

/// DEFAULT_EQ_SEL, DEFAULT_INEQ_SEL, and DEFAULT_SEL are the shares of rows that Postgres assumes an equality, an
/// inequality, and any other condition keeps without statistics.
const DEFAULT_EQ_SEL: f64 = 0.005;
const DEFAULT_INEQ_SEL: f64 = 1.0 / 3.0;
const DEFAULT_SEL: f64 = 0.5;

/// ColumnStats are a column's statistics: the share of NULLs, about how many distinct values it holds, its most common
/// values with their shares of all rows, the bounds that split its other values into equally full buckets, and how
/// closely the order of its values follows the order of the table's rows.
#[derive(Debug, Default)]
pub struct ColumnStats {
    pub null_frac: f64,
    pub distinct: f64,
    pub common: Vec<(Value, f64)>,
    pub histogram: Vec<Value>,
    pub correlation: f64,
}

/// TableStats are the statistics of a table's columns, with the row count that they were gathered at.
#[derive(Debug, Default)]
pub struct TableStats {
    pub rows: u64,
    pub columns: Vec<ColumnStats>,
}

/// TableKey names a table of a branch of a database by its schema and name.
type TableKey = (String, String, String, String);

thread_local! {
    /// TABLES are the statistics this thread gathered, by database, branch, schema, and table.
    static TABLES: RefCell<HashMap<TableKey, Arc<TableStats>>> = RefCell::new(HashMap::new());
}

/// table_stats returns the statistics of a table, gathering them again once its row count moved by more than a tenth
/// since they were gathered, as autovacuum's analyze threshold does, or its columns changed.
pub fn table_stats(ctx: &mut Ctx<'_>, table: &TableDef) -> Option<Arc<TableStats>> {
    let rows = prolly::Node::decode(table.table.primary_index.clone()).ok()?.tree_count();
    let key = (ctx.session.database.clone(), ctx.session.branch.clone(), table.schema.clone(), table.name.clone());
    if let Some(stats) = TABLES.with(|t| t.borrow().get(&key).cloned())
        && stats.columns.len() == table.columns.len()
        && (stats.rows as f64 - rows as f64).abs() <= stats.rows as f64 / 10.0
    {
        return Some(stats);
    }
    let stats = Arc::new(gather(ctx, table, rows).ok()?);
    TABLES.with(|t| t.borrow_mut().insert(key, stats.clone()));
    Some(stats)
}

/// gather reads a sample of a table's rows and computes each column's statistics from it.
fn gather(ctx: &mut Ctx<'_>, table: &TableDef, rows: u64) -> crate::error::Result<TableStats> {
    let root = Arc::new(prolly::Node::decode(table.table.primary_index.clone())?);
    let sampled = prolly::sample(ctx.db, root, SAMPLE_ROWS)?;
    let mut columns: Vec<Vec<Value>> = vec![Vec::with_capacity(sampled.len()); table.columns.len()];
    let mut row = Vec::new();
    for (key, value) in &sampled {
        let copies = table.decode_columns_into(ctx.db, key, value, None, &mut row)?;
        for (column, value) in columns.iter_mut().zip(&row) {
            for _ in 0..copies.max(1) {
                column.push(value.clone());
            }
        }
    }
    let columns = columns.into_iter().map(|values| column_stats(values, rows)).collect();
    Ok(TableStats { rows, columns })
}

/// column_stats computes a column's statistics from a sample of its values, in the order of the table's rows, in a
/// table of this many rows, as compute_scalar_stats does: the distinct count by the Haas-Stokes estimator, the values
/// common enough to stand out as most common values, a histogram of the rest, and the correlation between the
/// values' order and the rows' order.
fn column_stats(values: Vec<Value>, rows: u64) -> ColumnStats {
    let sampled = values.len() as f64;
    let mut values: Vec<(usize, Value)> = values.into_iter().filter(|v| !v.is_null()).enumerate().collect();
    if sampled == 0.0 {
        return ColumnStats::default();
    }
    let null_frac = 1.0 - values.len() as f64 / sampled;
    values.sort_by(|(a_tupno, a), (b_tupno, b)| compare_values(a, b).then(a_tupno.cmp(b_tupno)));
    let values_cnt = values.len() as f64;
    let corr_xysum: f64 = values.iter().enumerate().map(|(i, (tupno, _))| i as f64 * *tupno as f64).sum();
    let corr_xsum = (values_cnt - 1.0) * values_cnt / 2.0;
    let corr_x2sum = (values_cnt - 1.0) * values_cnt * (2.0 * values_cnt - 1.0) / 6.0;
    let correlation = match values_cnt > 1.0 {
        true => (values_cnt * corr_xysum - corr_xsum * corr_xsum) / (values_cnt * corr_x2sum - corr_xsum * corr_xsum),
        false => 0.0,
    };
    let values: Vec<Value> = values.into_iter().map(|(_, v)| v).collect();
    let mut runs: Vec<(usize, usize)> = Vec::new();
    for (i, value) in values.iter().enumerate() {
        match runs.last_mut() {
            Some((start, count)) if compare_values(&values[*start], value) == Ordering::Equal => *count += 1,
            _ => runs.push((i, 1)),
        }
    }
    let (n, d) = (values.len() as f64, runs.len() as f64);
    let total = rows as f64 * (1.0 - null_frac);
    let singles = runs.iter().filter(|(_, count)| *count == 1).count() as f64;
    let distinct = if n == 0.0 {
        0.0
    } else if singles == d || n >= total {
        if singles == d { total.max(d) } else { d }
    } else {
        (n * d / (n - singles + singles * n / total)).clamp(d, total)
    };
    let mut candidates: Vec<(usize, usize)> = runs.iter().copied().filter(|(_, count)| *count > 1).collect();
    candidates.sort_by(|a, b| b.1.cmp(&a.1).then(a.0.cmp(&b.0)));
    let common_count = if d <= STATISTICS_TARGET as f64 && distinct <= d && singles == 0.0 {
        candidates.len()
    } else {
        let average = n / distinct.max(1.0);
        let least = (1.25 * average).min(n / STATISTICS_TARGET as f64).max(2.0);
        candidates.iter().take(STATISTICS_TARGET).take_while(|(_, count)| *count as f64 >= least).count()
    };
    let common_runs: Vec<(usize, usize)> = candidates[..common_count].to_vec();
    let common = common_runs.iter().map(|&(start, count)| (values[start].clone(), count as f64 / sampled)).collect();
    let mut rest: Vec<Value> = Vec::new();
    let mut rest_distinct = 0;
    for &(start, count) in &runs {
        if !common_runs.iter().any(|&(s, _)| s == start) {
            rest.extend(values[start..start + count].iter().cloned());
            rest_distinct += 1;
        }
    }
    let bounds = rest_distinct.min(STATISTICS_TARGET + 1);
    let histogram = match bounds >= 2 {
        true => (0..bounds).map(|i| rest[i * (rest.len() - 1) / (bounds - 1)].clone()).collect(),
        false => Vec::new(),
    };
    ColumnStats { null_frac, distinct, common, histogram, correlation }
}

/// selectivity returns about what share of a table's rows a condition over its columns keeps, as Postgres'
/// clause_selectivity estimates it from the table's statistics, with Postgres' defaults where they say nothing.
pub fn selectivity(stats: &TableStats, condition: &Expr) -> f64 {
    let column = |e: &Expr| match e {
        Expr::Column(c) => stats.columns.get(*c),
        Expr::Cast(inner, ..) => match **inner {
            Expr::Column(c) => stats.columns.get(c),
            _ => None,
        },
        _ => None,
    };
    let selectivity = match condition {
        Expr::And(a, b) => selectivity(stats, a) * selectivity(stats, b),
        Expr::Or(a, b) => {
            let (a, b) = (selectivity(stats, a), selectivity(stats, b));
            a + b - a * b
        }
        Expr::Not(inner) => 1.0 - selectivity(stats, inner),
        Expr::IsNull(e, negated) => match (column(e), negated) {
            (Some(c), false) => c.null_frac,
            (Some(c), true) => 1.0 - c.null_frac,
            (None, false) => DEFAULT_EQ_SEL,
            (None, true) => 1.0 - DEFAULT_EQ_SEL,
        },
        Expr::Compare(op, l, r) => match (column(l), &**r, column(r), &**l) {
            (Some(c), Expr::Const(v), _, _) => compare_selectivity(c, *op, v),
            (_, _, Some(c), Expr::Const(v)) => compare_selectivity(c, flip(*op), v),
            (Some(a), _, Some(b), _) if *op == CmpOp::Eq => 1.0 / a.distinct.max(b.distinct).max(1.0),
            _ => match op {
                CmpOp::Eq => DEFAULT_EQ_SEL,
                CmpOp::Ne => 1.0 - DEFAULT_EQ_SEL,
                _ => DEFAULT_INEQ_SEL,
            },
        },
        Expr::Column(_) => column(condition).map_or(DEFAULT_SEL, |c| equal_selectivity(c, &Value::Bool(true))),
        _ => DEFAULT_SEL,
    };
    selectivity.clamp(0.0, 1.0)
}

/// flip returns the comparison with its operands swapped.
fn flip(op: CmpOp) -> CmpOp {
    match op {
        CmpOp::Lt => CmpOp::Gt,
        CmpOp::Le => CmpOp::Ge,
        CmpOp::Gt => CmpOp::Lt,
        CmpOp::Ge => CmpOp::Le,
        other => other,
    }
}

/// compare_selectivity returns the share of rows whose column compares with a constant as asked.
fn compare_selectivity(column: &ColumnStats, op: CmpOp, value: &Value) -> f64 {
    if value.is_null() {
        return 0.0;
    }
    match op {
        CmpOp::Eq => equal_selectivity(column, value),
        CmpOp::Ne => 1.0 - equal_selectivity(column, value) - column.null_frac,
        _ => range_selectivity(column, op, value),
    }
}

/// equal_selectivity returns the share of rows whose column equals a value, as eqsel judges it: the value's share
/// when it is a most common value, and otherwise an even share of what the most common values leave over the other
/// distinct values.
fn equal_selectivity(column: &ColumnStats, value: &Value) -> f64 {
    if let Some((_, share)) = column.common.iter().find(|(v, _)| compare_values(v, value) == Ordering::Equal) {
        return *share;
    }
    let common: f64 = column.common.iter().map(|(_, share)| share).sum();
    let others = column.distinct - column.common.len() as f64;
    match others >= 1.0 {
        true => ((1.0 - common - column.null_frac) / others).max(0.0),
        false => 0.0,
    }
}

/// range_selectivity returns the share of rows whose column compares with a value as an inequality asks, as
/// scalarineqsel judges it: the shares of the most common values that pass, plus the part of the histogram below or
/// above the value, interpolated within its bucket, of the share that the most common values leave.
fn range_selectivity(column: &ColumnStats, op: CmpOp, value: &Value) -> f64 {
    let passes = |v: &Value| op.test(compare_values(v, value));
    let common_share: f64 = column.common.iter().map(|(_, share)| share).sum();
    let common_passing: f64 = column.common.iter().filter(|(v, _)| passes(v)).map(|(_, share)| share).sum();
    let rest = 1.0 - common_share - column.null_frac;
    match below_fraction(&column.histogram, value) {
        Some(below) => {
            let fraction = if matches!(op, CmpOp::Lt | CmpOp::Le) { below } else { 1.0 - below };
            common_passing + fraction * rest
        }
        None if common_share > 0.0 => common_passing / common_share * (1.0 - column.null_frac),
        None => DEFAULT_INEQ_SEL,
    }
}

/// below_fraction returns about what share of a histogram's values lie below a value, interpolating within the bucket
/// that holds it when its bounds are numbers, as ineq_histogram_selectivity does, or None without a histogram.
fn below_fraction(histogram: &[Value], value: &Value) -> Option<f64> {
    if histogram.len() < 2 {
        return None;
    }
    let buckets = (histogram.len() - 1) as f64;
    if compare_values(value, &histogram[0]) != Ordering::Greater {
        return Some(0.0);
    }
    if compare_values(value, &histogram[histogram.len() - 1]) != Ordering::Less {
        return Some(1.0);
    }
    let upper = histogram.partition_point(|bound| compare_values(bound, value) == Ordering::Less);
    let (low, high) = (&histogram[upper - 1], &histogram[upper]);
    let within = match (number(low), number(high), number(value)) {
        (Some(l), Some(h), Some(v)) if h > l => ((v - l) / (h - l)).clamp(0.0, 1.0),
        _ => 0.5,
    };
    Some(((upper - 1) as f64 + within) / buckets)
}

/// number returns a value as a number that interpolation can place between two others, for the numeric and
/// date and time types.
fn number(value: &Value) -> Option<f64> {
    Some(match value {
        Value::Int2(v) => *v as f64,
        Value::Int4(v) | Value::Date(v) => *v as f64,
        Value::Int8(v) | Value::Time(v) | Value::Timestamp(v) | Value::TimestampTz(v) => *v as f64,
        Value::Float4(v) => *v as f64,
        Value::Float8(v) => *v,
        Value::Numeric(n) => n.to_f64(),
        _ => return None,
    })
}

/// column_distinct returns about how many distinct values a column of a table holds, from its statistics.
pub fn column_distinct(ctx: &mut Ctx<'_>, table: &TableDef, column: usize) -> Option<f64> {
    table_stats(ctx, table)?.columns.get(column).map(|c| c.distinct)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn unique_columns_count_every_row_as_distinct() {
        let stats = column_stats((0..1000).map(Value::Int4).collect(), 10_000);
        assert_eq!(stats.distinct, 10_000.0);
        assert!(stats.common.is_empty());
        assert_eq!(stats.histogram.len(), STATISTICS_TARGET + 1);
        let column = ColumnStats { distinct: stats.distinct, ..stats };
        assert!((range_selectivity(&column, CmpOp::Lt, &Value::Int4(250)) - 0.25).abs() < 0.01);
        assert!((equal_selectivity(&column, &Value::Int4(7)) - 0.0001).abs() < 1e-9);
    }

    #[test]
    fn skewed_columns_keep_their_common_values() {
        let mut values: Vec<Value> = (0..900).map(|_| Value::Text("new".into())).collect();
        values.extend((0..100).map(|i| Value::Text(format!("v{}", i % 50))));
        values.extend((0..100).map(|_| Value::Null));
        let stats = column_stats(values, 1100);
        assert!((stats.null_frac - 100.0 / 1100.0).abs() < 1e-9);
        assert_eq!(stats.distinct, 51.0);
        assert!((equal_selectivity(&stats, &Value::Text("new".into())) - 900.0 / 1100.0).abs() < 1e-9);
        assert!((equal_selectivity(&stats, &Value::Text("v1".into())) - 2.0 / 1100.0).abs() < 1e-9);
    }
}
