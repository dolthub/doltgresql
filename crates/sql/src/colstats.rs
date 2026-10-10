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

/// STATISTICS_TARGET is the most common values and histogram bounds that a column keeps, as Postgres'
/// default_statistics_target is.
const STATISTICS_TARGET: usize = 100;

/// SAMPLE_ROWS is how many rows of a table the statistics read, as Postgres' std_typanalyze asks of ANALYZE.
const SAMPLE_ROWS: usize = 300 * STATISTICS_TARGET;

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
    /// The distinct values as Postgres' stadistinct holds them: a count, or the negated share of the rows when the
    /// count grows with the table, or 0 when unknown.
    pub stadistinct: f64,
    /// The average width of the column's values that are not NULL, as Postgres' stawidth is.
    pub width: f64,
    pub common: Vec<(Value, f64)>,
    pub histogram: Vec<Value>,
    pub correlation: f64,
    /// The most common values in order, each with the shares of the values up to and including it, which find a
    /// value's share and the share below or above a value by binary search.
    common_by_value: Vec<(Value, f64)>,
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
    let columns = columns
        .into_iter()
        .zip(&table.columns)
        .map(|(values, column)| column_stats(values, rows, column.ty.oid))
        .collect();
    Ok(TableStats { rows, columns })
}

/// WIDTH_THRESHOLD is the width beyond which a value takes no part in the most common values and histogram, as
/// Postgres' constant of the same name is.
const WIDTH_THRESHOLD: usize = 1024;

/// column_stats computes a column's statistics from a sample of its values, in the order of the table's rows, in a
/// table of this many rows, as Postgres' std_typanalyze chooses how: compute_scalar_stats for a type with an
/// ordering, compute_distinct_stats for one with only an equality, and compute_trivial_stats otherwise.
fn column_stats(values: Vec<Value>, rows: u64, oid: u32) -> ColumnStats {
    let typlen = crate::catalog::builtin_type(oid).map_or(-1, |t| t.definition.typ_length);
    let mut stats = match oid {
        crate::oid::JSON | crate::oid::XML => compute_trivial_stats(&values, typlen),
        _ if crate::optimizer::btree_opfamily(oid).is_none() => compute_distinct_stats(&values, rows, typlen),
        _ => compute_scalar_stats(values, rows, typlen),
    };
    stats.distinct = match stats.stadistinct < 0.0 {
        true => -stats.stadistinct * rows as f64,
        false => stats.stadistinct,
    };
    let mut sorted: Vec<&(Value, f64)> = stats.common.iter().collect();
    sorted.sort_by(|a, b| compare_values(&a.0, &b.0));
    let mut cumulative = 0.0;
    stats.common_by_value = sorted
        .into_iter()
        .map(|(value, share)| {
            cumulative += share;
            (value.clone(), cumulative)
        })
        .collect();
    stats
}

/// value_width returns the bytes that Postgres stores a value of a type of a length in, a variable-length value's
/// header included, and whether it is too wide for the most common values and histogram.
fn value_width(value: &Value, typlen: i16) -> (f64, bool) {
    if typlen > 0 {
        return (f64::from(typlen), false);
    }
    let len = match value {
        Value::Text(s) | Value::Json(s) | Value::Bit(s) => s.len(),
        Value::Bytea(b) => b.len(),
        other => other.output().map_or(0, |s| s.len()),
    };
    let header = if typlen == -1 && len + 1 <= 127 {
        1
    } else if typlen == -1 {
        4
    } else {
        1
    };
    ((len + header) as f64, len > WIDTH_THRESHOLD)
}

/// distinct_estimate returns Postgres' stadistinct of a column from a sample of `n` values that are not NULL, of which
/// `d` are distinct and `f1` appear once, in a table of `totalrows` rows with a share of NULLs, by the Haas-Stokes
/// estimator, rounded, as a negated share of the rows when it is more than a tenth of them.
fn distinct_estimate(n: f64, d: f64, f1: f64, null_frac: f64, totalrows: f64) -> f64 {
    let big_n = totalrows * (1.0 - null_frac);
    let estimate = if big_n > 0.0 { (n * d) / ((n - f1) + f1 * n / big_n) } else { 0.0 };
    let estimate = (estimate.max(d).min(big_n) + 0.5).floor();
    match estimate > 0.1 * totalrows {
        true => -(estimate / totalrows),
        false => estimate,
    }
}

/// compute_scalar_stats computes the statistics of a column whose type has an ordering, as Postgres' function of the
/// same name does: the share of NULLs, the average width, the distinct values, the values common enough to stand out
/// as most common values, a histogram of the rest, and the correlation between the values' order and the rows' order.
fn compute_scalar_stats(sample: Vec<Value>, rows: u64, typlen: i16) -> ColumnStats {
    let samplerows = sample.len() as f64;
    let totalrows = rows as f64;
    let (mut null_cnt, mut nonnull_cnt, mut toowide_cnt, mut total_width) = (0, 0, 0, 0.0);
    let mut values: Vec<(usize, Value)> = Vec::new();
    for value in sample {
        if value.is_null() {
            null_cnt += 1;
            continue;
        }
        nonnull_cnt += 1;
        let (width, toowide) = value_width(&value, typlen);
        total_width += width;
        if toowide {
            toowide_cnt += 1;
            continue;
        }
        values.push((values.len(), value));
    }
    let null_frac = if samplerows > 0.0 { null_cnt as f64 / samplerows } else { 0.0 };
    let width = match (nonnull_cnt, typlen > 0) {
        (_, true) => f64::from(typlen),
        (0, false) => 0.0,
        (_, false) => total_width / nonnull_cnt as f64,
    };
    if values.is_empty() {
        let stadistinct = match nonnull_cnt > 0 {
            true => -(1.0 - null_frac),
            false => 0.0,
        };
        return ColumnStats { null_frac, stadistinct, width, ..Default::default() };
    }
    values.sort_by(|(a_tupno, a), (b_tupno, b)| compare_values(a, b).then(a_tupno.cmp(b_tupno)));
    let values_cnt = values.len() as f64;
    let corr_xysum: f64 = values.iter().enumerate().map(|(i, (tupno, _))| i as f64 * *tupno as f64).sum();
    let values: Vec<Value> = values.into_iter().map(|(_, v)| v).collect();
    let mut runs: Vec<(usize, usize)> = Vec::new();
    for (i, value) in values.iter().enumerate() {
        match runs.last_mut() {
            Some((start, count)) if compare_values(&values[*start], value) == Ordering::Equal => *count += 1,
            _ => runs.push((i, 1)),
        }
    }
    let ndistinct = runs.len();
    let nmultiple = runs.iter().filter(|(_, count)| *count > 1).count();
    let stadistinct = if nmultiple == 0 {
        -(1.0 - null_frac)
    } else if toowide_cnt == 0 && nmultiple == ndistinct {
        let d = ndistinct as f64;
        if d > 0.1 * totalrows { -(d / totalrows) } else { d }
    } else {
        let f1 = (ndistinct - nmultiple + toowide_cnt) as f64;
        distinct_estimate(samplerows - null_cnt as f64, f1 + nmultiple as f64, f1, null_frac, totalrows)
    };
    let mut track: Vec<(usize, usize)> = runs.iter().copied().filter(|(_, count)| *count > 1).collect();
    track.sort_by(|a, b| b.1.cmp(&a.1).then(a.0.cmp(&b.0)));
    track.truncate(STATISTICS_TARGET);
    let num_mcv = if track.len() == ndistinct && toowide_cnt == 0 && stadistinct > 0.0 {
        track.len()
    } else {
        let counts: Vec<usize> = track.iter().map(|(_, count)| *count).collect();
        let ndistinct_table = if stadistinct < 0.0 { -stadistinct * totalrows } else { stadistinct };
        match counts.is_empty() {
            true => 0,
            false => analyze_mcv_list(&counts, ndistinct_table, null_frac, samplerows, totalrows),
        }
    };
    let common_runs: Vec<(usize, usize)> = track[..num_mcv].to_vec();
    let common: Vec<(Value, f64)> =
        common_runs.iter().map(|&(start, count)| (values[start].clone(), count as f64 / samplerows)).collect();
    let mut num_hist = ndistinct - num_mcv;
    if num_hist > STATISTICS_TARGET {
        num_hist = STATISTICS_TARGET + 1;
    }
    let histogram = match num_hist >= 2 {
        true => {
            let mut rest: Vec<&Value> = Vec::new();
            for &(start, count) in &runs {
                if !common_runs.iter().any(|&(s, _)| s == start) {
                    rest.extend(&values[start..start + count]);
                }
            }
            (0..num_hist).map(|i| rest[i * (rest.len() - 1) / (num_hist - 1)].clone()).collect()
        }
        false => Vec::new(),
    };
    let correlation = match values_cnt > 1.0 {
        true => {
            let corr_xsum = (values_cnt - 1.0) * values_cnt / 2.0;
            let corr_x2sum = (values_cnt - 1.0) * values_cnt * (2.0 * values_cnt - 1.0) / 6.0;
            (values_cnt * corr_xysum - corr_xsum * corr_xsum) / (values_cnt * corr_x2sum - corr_xsum * corr_xsum)
        }
        false => 0.0,
    };
    ColumnStats { null_frac, stadistinct, width, common, histogram, correlation, ..Default::default() }
}

/// compute_distinct_stats computes the statistics of a column whose type has an equality but no ordering, as Postgres'
/// function of the same name does: the share of NULLs, the average width, the distinct values, and the most common
/// values, which it tracks among twice as many candidates as it keeps.
fn compute_distinct_stats(sample: &[Value], rows: u64, typlen: i16) -> ColumnStats {
    let samplerows = sample.len() as f64;
    let totalrows = rows as f64;
    let track_max = (2 * STATISTICS_TARGET).max(10);
    let mut track: Vec<(Value, usize)> = Vec::with_capacity(track_max);
    let (mut null_cnt, mut nonnull_cnt, mut toowide_cnt, mut total_width) = (0, 0, 0, 0.0);
    for value in sample {
        if value.is_null() {
            null_cnt += 1;
            continue;
        }
        nonnull_cnt += 1;
        let (width, toowide) = value_width(value, typlen);
        total_width += width;
        if toowide {
            toowide_cnt += 1;
            continue;
        }
        let mut firstcount1 = track.len();
        let mut found = None;
        for (j, (tracked, count)) in track.iter().enumerate() {
            if compare_values(value, tracked) == Ordering::Equal {
                found = Some(j);
                break;
            }
            if j < firstcount1 && *count == 1 {
                firstcount1 = j;
            }
        }
        match found {
            Some(mut j) => {
                track[j].1 += 1;
                while j > 0 && track[j].1 > track[j - 1].1 {
                    track.swap(j, j - 1);
                    j -= 1;
                }
            }
            None => {
                if track.len() < track_max {
                    track.push((Value::Null, 0));
                }
                if firstcount1 < track.len() {
                    for j in (firstcount1 + 1..track.len()).rev() {
                        track[j] = track[j - 1].clone();
                    }
                    track[firstcount1] = (value.clone(), 1);
                }
            }
        }
    }
    if nonnull_cnt == 0 {
        let null_frac = if null_cnt > 0 { 1.0 } else { 0.0 };
        let width = if typlen > 0 { f64::from(typlen) } else { 0.0 };
        return ColumnStats { null_frac, width, ..Default::default() };
    }
    let null_frac = null_cnt as f64 / samplerows;
    let width = if typlen > 0 { f64::from(typlen) } else { total_width / nonnull_cnt as f64 };
    let nmultiple = track.iter().take_while(|(_, count)| *count > 1).count();
    let summultiple: usize = track[..nmultiple].iter().map(|(_, count)| count).sum();
    let stadistinct = if nmultiple == 0 {
        -(1.0 - null_frac)
    } else if track.len() < track_max && toowide_cnt == 0 && nmultiple == track.len() {
        let d = track.len() as f64;
        if d > 0.1 * totalrows { -(d / totalrows) } else { d }
    } else {
        let f1 = (nonnull_cnt - summultiple) as f64;
        distinct_estimate(samplerows - null_cnt as f64, f1 + nmultiple as f64, f1, null_frac, totalrows)
    };
    let mut num_mcv = STATISTICS_TARGET;
    if track.len() < track_max && toowide_cnt == 0 && stadistinct > 0.0 && track.len() <= num_mcv {
        num_mcv = track.len();
    } else {
        num_mcv = num_mcv.min(track.len());
        if num_mcv > 0 {
            let counts: Vec<usize> = track[..num_mcv].iter().map(|(_, count)| *count).collect();
            let ndistinct_table = if stadistinct < 0.0 { -stadistinct * totalrows } else { stadistinct };
            num_mcv = analyze_mcv_list(&counts, ndistinct_table, null_frac, samplerows, totalrows);
        }
    }
    let common = track[..num_mcv].iter().map(|(value, count)| (value.clone(), *count as f64 / samplerows)).collect();
    ColumnStats { null_frac, stadistinct, width, common, ..Default::default() }
}

/// compute_trivial_stats computes the share of NULLs and the average width of a column whose type has no equality, as
/// Postgres' function of the same name does.
fn compute_trivial_stats(sample: &[Value], typlen: i16) -> ColumnStats {
    let samplerows = sample.len() as f64;
    let nonnull: Vec<&Value> = sample.iter().filter(|v| !v.is_null()).collect();
    let null_frac = match (nonnull.is_empty(), samplerows > 0.0) {
        (true, true) => 1.0,
        (_, true) => 1.0 - nonnull.len() as f64 / samplerows,
        (_, false) => 0.0,
    };
    let width = match (typlen > 0, nonnull.is_empty()) {
        (true, _) => f64::from(typlen),
        (false, true) => 0.0,
        (false, false) => nonnull.iter().map(|v| value_width(v, typlen).0).sum::<f64>() / nonnull.len() as f64,
    };
    ColumnStats { null_frac, width, ..Default::default() }
}

/// analyze_mcv_list returns how many of the most common values of a sample, by their counts in descending order, are
/// significantly more common than the values left out would suggest, as Postgres' function of the same name decides:
/// a value is kept when its count lies more than two standard errors of the hypergeometric distribution above what
/// an even share of the rest would give it.
fn analyze_mcv_list(
    mcv_counts: &[usize],
    ndistinct_table: f64,
    stanullfrac: f64,
    samplerows: f64,
    totalrows: f64,
) -> usize {
    let mut num_mcv = mcv_counts.len();
    if samplerows == totalrows || totalrows <= 1.0 {
        return num_mcv;
    }
    let mut sumcount: f64 = mcv_counts[..num_mcv.saturating_sub(1)].iter().map(|&c| c as f64).sum();
    while num_mcv > 0 {
        let mut selec = (1.0 - sumcount / samplerows - stanullfrac).clamp(0.0, 1.0);
        let otherdistinct = ndistinct_table - (num_mcv - 1) as f64;
        if otherdistinct > 1.0 {
            selec /= otherdistinct;
        }
        let (big_n, n) = (totalrows, samplerows);
        let k = big_n * mcv_counts[num_mcv - 1] as f64 / n;
        let variance = n * k * (big_n - k) * (big_n - n) / (big_n * big_n * (big_n - 1.0));
        if mcv_counts[num_mcv - 1] as f64 > selec * samplerows + 2.0 * variance.sqrt() + 0.5 {
            break;
        }
        num_mcv -= 1;
        if num_mcv == 0 {
            break;
        }
        sumcount -= mcv_counts[num_mcv - 1] as f64;
    }
    num_mcv
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
        Expr::RowCompare(..) => selectivity(stats, &condition.clone().expand_row_compares()),
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
    let by_value = &column.common_by_value;
    if let Ok(i) = by_value.binary_search_by(|(v, _)| compare_values(v, value)) {
        return by_value[i].1 - if i == 0 { 0.0 } else { by_value[i - 1].1 };
    }
    let common = by_value.last().map_or(0.0, |(_, cumulative)| *cumulative);
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
    let by_value = &column.common_by_value;
    let common_share = by_value.last().map_or(0.0, |(_, cumulative)| *cumulative);
    let below = |inclusive: bool| {
        let end = by_value.partition_point(|(v, _)| match compare_values(v, value) {
            Ordering::Less => true,
            Ordering::Equal => inclusive,
            Ordering::Greater => false,
        });
        if end == 0 { 0.0 } else { by_value[end - 1].1 }
    };
    let common_passing = match op {
        CmpOp::Lt => below(false),
        CmpOp::Le => below(true),
        CmpOp::Gt => common_share - below(true),
        _ => common_share - below(false),
    };
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
        let stats = column_stats((0..1000).map(Value::Int4).collect(), 10_000, crate::oid::INT4);
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
        let stats = column_stats(values, 1100, crate::oid::TEXT);
        assert!((stats.null_frac - 100.0 / 1100.0).abs() < 1e-9);
        assert_eq!(stats.distinct, 51.0);
        assert!((equal_selectivity(&stats, &Value::Text("new".into())) - 900.0 / 1100.0).abs() < 1e-9);
        assert!((equal_selectivity(&stats, &Value::Text("v1".into())) - 2.0 / 1100.0).abs() < 1e-9);
    }
}
