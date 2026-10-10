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

//! Postgres' optimizer/path/costsize.c: the costs of paths and the sizes of relations, in Postgres' units, with
//! its default cost settings. Dolt has no pages, so a table's pages are those its rows would fill in Postgres' heap,
//! and its primary index holds its rows in key order, so a scan of that index fetches nothing else.

use std::rc::Rc;

use super::PlannerInfo;
use super::clausesel::{self, clause_selectivity, clauselist_selectivity};
use super::nodes::{JoinType, Path, PathKind, Relids, RinfoId, SpecialJoinInfo};
use super::restrictinfo::{join_clause_is_movable_into, rinfo_is_pushed_down};
use crate::expr::Expr;
use crate::types::Value;

/// SEQ_PAGE_COST, RANDOM_PAGE_COST, CPU_TUPLE_COST, CPU_INDEX_TUPLE_COST, and CPU_OPERATOR_COST are Postgres' default
/// cost settings: of reading a page in order and out of order, and of processing a row, an index entry, and an
/// operator.
const SEQ_PAGE_COST: f64 = 1.0;
pub const RANDOM_PAGE_COST: f64 = 4.0;
pub const CPU_TUPLE_COST: f64 = 0.01;
pub const CPU_INDEX_TUPLE_COST: f64 = 0.005;
pub const CPU_OPERATOR_COST: f64 = 0.0025;

/// EFFECTIVE_CACHE_SIZE is the pages that Postgres assumes the cache holds, as its default effective_cache_size.
const EFFECTIVE_CACHE_SIZE: f64 = 524288.0;

/// DISABLE_COST is the cost that Postgres adds to a hash join whose most common value's rows would overflow the
/// hash table's memory.
const DISABLE_COST: f64 = 1.0e10;

/// BLCKSZ is the size of a Postgres page, and PAGE_HEADER and TUPLE_HEADER the sizes of a page's and a heap row's
/// headers, aligned, with ITEM_ID the size of a row's pointer on its page.
const BLCKSZ: f64 = 8192.0;
const PAGE_HEADER: f64 = 24.0;
pub const TUPLE_HEADER: f64 = 24.0;
const ITEM_ID: f64 = 4.0;

/// HASH_MEM is the memory that a hash join's table may take, as Postgres' default work_mem times its default
/// hash_mem_multiplier.
pub const HASH_MEM: f64 = 4.0 * 1024.0 * 1024.0 * 2.0;

/// ROUTINE_COST is the cost of a call of a user-defined routine in units of an operator's, as Postgres' default COST
/// of a function written in SQL or PL/pgSQL.
const ROUTINE_COST: f64 = 100.0;

/// Enables are the enable_* settings that the planner reads, which count a disabled plan node for a kind of path
/// when off, as Postgres 18's do.
#[derive(Clone, Copy, Debug)]
pub struct Enables {
    pub seqscan: bool,
    pub indexscan: bool,
    pub indexonlyscan: bool,
    pub bitmapscan: bool,
    pub nestloop: bool,
    pub hashjoin: bool,
    pub material: bool,
    pub sort: bool,
    pub incremental_sort: bool,
    pub hashagg: bool,
    pub self_join_elimination: bool,
}

impl Enables {
    /// read returns the enable_* settings of a session.
    pub fn read(settings: &crate::settings::Settings) -> Enables {
        let on = |name: &str| settings.get(name).is_none_or(|value| value != "off");
        Enables {
            seqscan: on("enable_seqscan"),
            indexscan: on("enable_indexscan"),
            indexonlyscan: on("enable_indexonlyscan"),
            bitmapscan: on("enable_bitmapscan"),
            nestloop: on("enable_nestloop"),
            hashjoin: on("enable_hashjoin"),
            material: on("enable_material"),
            sort: on("enable_sort"),
            incremental_sort: on("enable_incremental_sort"),
            hashagg: on("enable_hashagg"),
            self_join_elimination: on("enable_self_join_elimination"),
        }
    }
}

/// disabled counts the plan node of a kind of path that a setting turned off.
fn disabled(enabled: bool) -> usize {
    usize::from(!enabled)
}

/// Costs are the disabled plan nodes, startup cost, and total cost of a path.
pub type Costs = (usize, f64, f64);

/// clamp_row_est rounds a row estimate to a whole number of at least one, as Postgres' function of the same name does.
pub fn clamp_row_est(nrows: f64) -> f64 {
    if nrows.is_nan() || nrows > 1.0e100 {
        1.0e100
    } else if nrows <= 1.0 {
        1.0
    } else {
        nrows.round()
    }
}

/// QualCost is the cost of evaluating expressions: once when the plan starts, and for each row.
#[derive(Clone, Copy, Debug, Default)]
pub struct QualCost {
    pub startup: f64,
    pub per_tuple: f64,
}

/// cost_qual_eval returns the cost of evaluating a list of clauses, as Postgres' function of the same name does.
pub fn cost_qual_eval(root: &PlannerInfo<'_, '_>, quals: &[RinfoId]) -> QualCost {
    quals.iter().fold(QualCost::default(), |cost, &q| {
        let one = cost_qual_eval_node(&root.rinfos[q].clause);
        QualCost { startup: cost.startup + one.startup, per_tuple: cost.per_tuple + one.per_tuple }
    })
}

/// cost_qual_eval_node returns the cost of evaluating one expression, which charges each operator, function, and
/// cast it calls, and the subplans it runs, as Postgres' cost_qual_eval_walker does.
pub fn cost_qual_eval_node(e: &Expr) -> QualCost {
    let mut total = QualCost::default();
    cost_qual_eval_walker(e, &mut total);
    total
}

/// cost_qual_eval_walker adds the cost of evaluating an expression to a total, as Postgres' function of the same name
/// does: a subplan costs what cost_subplan found, and an initplan costs nothing here, since the query pays for it once.
fn cost_qual_eval_walker(e: &Expr, total: &mut QualCost) {
    let subplan = match e {
        Expr::SubPlan(subplan) => Some(&**subplan),
        Expr::AlternativeSubPlan(subplans) => subplans.first(),
        _ => None,
    };
    if let Some(subplan) = subplan {
        if !subplan.init_plan {
            total.startup += subplan.startup_cost;
            total.per_tuple += subplan.per_call_cost;
        }
        return;
    }
    total.per_tuple += match e {
        Expr::Routine(..) | Expr::Operator(..) => ROUTINE_COST * CPU_OPERATOR_COST,
        Expr::Func(..)
        | Expr::Cast(..)
        | Expr::Arith(..)
        | Expr::Neg(..)
        | Expr::Compare(..)
        | Expr::Concat(..)
        | Expr::DateTime(..)
        | Expr::ArrayOp(..)
        | Expr::DistinctFrom(..)
        | Expr::NullIf(..)
        | Expr::MinMax(..)
        | Expr::Exists(_)
        | Expr::Scalar(_)
        | Expr::ArraySubquery(..)
        | Expr::AnySubquery(..) => CPU_OPERATOR_COST,
        Expr::AnyArray(..) => CPU_OPERATOR_COST * 0.5 * 10.0,
        Expr::RowCompare(_, fields, _) => CPU_OPERATOR_COST * fields.len() as f64,
        _ => 0.0,
    };
    e.visit_children(&mut |c| cost_qual_eval_walker(c, total));
}

/// cost_subplan sets a subplan's startup and per-call costs from the path of its plan, whose rows it hashes or reads
/// again for each call, as Postgres' function of the same name does.
pub fn cost_subplan(subplan: &mut crate::expr::SubPlan, plan: &Path, use_hash_table: bool, materializes: bool) {
    let mut sp_cost = match &subplan.link {
        Expr::AnySubquery(test, ..) => cost_qual_eval_node(test),
        _ => QualCost::default(),
    };
    if use_hash_table {
        sp_cost.startup += plan.total_cost + CPU_OPERATOR_COST * plan.rows;
    } else {
        let plan_run_cost = plan.total_cost - plan.startup_cost;
        match &subplan.link {
            Expr::Exists(_) => sp_cost.per_tuple += plan_run_cost / clamp_row_est(plan.rows),
            Expr::AnySubquery(..) => {
                sp_cost.per_tuple += 0.50 * plan_run_cost;
                sp_cost.per_tuple += 0.50 * plan.rows * CPU_OPERATOR_COST;
            }
            _ => sp_cost.per_tuple += plan_run_cost,
        }
        if subplan.args.is_empty() && materializes {
            sp_cost.startup += plan.startup_cost;
        } else {
            sp_cost.per_tuple += plan.startup_cost;
        }
    }
    subplan.startup_cost = sp_cost.startup;
    subplan.per_call_cost = sp_cost.per_tuple;
}

/// get_typavgwidth returns the average width of the values of a type and modifier, as Postgres' function of the same
/// name estimates it: a fixed-length type's length, and otherwise a guess from its type modifier's maximum, where an
/// unknown type is variable-length.
pub fn get_typavgwidth(oid: Option<u32>, modifier: i32) -> f64 {
    let oid = oid.unwrap_or(0);
    let typlen = crate::catalog::builtin_type(oid).map_or(-1, |t| t.definition.typ_length);
    if typlen > 0 {
        return f64::from(typlen);
    }
    let maxwidth = match oid {
        crate::oid::VARCHAR | crate::oid::BPCHAR if modifier > 4 => (modifier - 4) * 4 + 4,
        _ => -1,
    };
    match maxwidth {
        w if w > 0 && oid == crate::oid::BPCHAR => f64::from(w),
        w if w > 0 && w <= 32 => f64::from(w),
        w if w > 0 && w < 1000 => f64::from(32 + (w - 32) / 2),
        w if w > 0 => f64::from(32 + (1000 - 32) / 2),
        _ => 32.0,
    }
}

/// estimate_rel_pages returns how many pages of Postgres' heap a relation's rows fill, as Postgres' estimate_rel_size
/// estimates the pages of a table from its row width.
pub fn estimate_rel_pages(tuples: f64, width: f64) -> f64 {
    let tuple_width = maxalign(width) + TUPLE_HEADER + ITEM_ID;
    let density = ((BLCKSZ - PAGE_HEADER) / tuple_width).floor().max(1.0);
    (tuples / density).ceil()
}

/// estimate_rel_size returns the pages and rows of a table of rows of a width, as Postgres' function of the same name
/// estimates them: its own once VACUUM or ANALYZE measured it, and otherwise at least ten pages, at the density that
/// its column types give, since a table that was never measured may not stay small.
pub fn estimate_rel_size(rows: f64, width: f64, measured: bool) -> (f64, f64) {
    let curpages = estimate_rel_pages(rows, width);
    if measured {
        return (curpages, rows);
    }
    let curpages = curpages.max(10.0);
    let density = ((BLCKSZ - PAGE_HEADER) / (width.floor() + TUPLE_HEADER + ITEM_ID)).floor();
    (curpages, (density * curpages).round())
}

/// maxalign rounds a width up to a multiple of eight bytes, as Postgres' MAXALIGN does.
pub fn maxalign(width: f64) -> f64 {
    (width / 8.0).ceil() * 8.0
}

/// relation_byte_size returns how many bytes a relation's rows take in memory, as Postgres' function of the same name
/// estimates it.
fn relation_byte_size(tuples: f64, width: f64) -> f64 {
    tuples * (maxalign(width) + TUPLE_HEADER)
}

/// page_size returns how many pages a relation's rows take in memory, as Postgres' function of the same name
/// estimates it.
fn page_size(tuples: f64, width: f64) -> f64 {
    (relation_byte_size(tuples, width) / BLCKSZ).ceil()
}

/// cost_seqscan returns the costs of reading every row of a base relation and testing its restrictions, as Postgres'
/// function of the same name does.
pub fn cost_seqscan(root: &PlannerInfo<'_, '_>, rel: usize) -> Costs {
    let rel = &root.rels[rel];
    let qpqual_cost = cost_qual_eval(root, &rel.baserestrictinfo);
    let startup_cost = qpqual_cost.startup;
    let disk_run_cost = SEQ_PAGE_COST * rel.pages;
    let cpu_run_cost = (CPU_TUPLE_COST + qpqual_cost.per_tuple) * rel.tuples;
    (disabled(root.enables.seqscan), startup_cost, startup_cost + cpu_run_cost + disk_run_cost)
}

/// cost_opaque_scan returns the costs of reading the rows of a relation that is not a table and testing its
/// restrictions, as Postgres' cost_functionscan does for a function's rows, where reading a system catalog in full is
/// a sequential scan, which enable_seqscan may turn off.
pub fn cost_opaque_scan(root: &PlannerInfo<'_, '_>, rel: usize, catalog: bool) -> Costs {
    let rel = &root.rels[rel];
    let qpqual_cost = cost_qual_eval(root, &rel.baserestrictinfo);
    let startup = qpqual_cost.startup + CPU_OPERATOR_COST;
    let total = startup + (CPU_TUPLE_COST + qpqual_cost.per_tuple) * rel.tuples;
    (disabled(root.enables.seqscan || !catalog), startup, total)
}

/// cost_resultscan returns the costs of the one row of a RESULT relation under its restrictions, as Postgres'
/// function of the same name does.
pub fn cost_resultscan(root: &PlannerInfo<'_, '_>, rel: usize) -> Costs {
    let qpqual_cost = cost_qual_eval(root, &root.rels[rel].baserestrictinfo);
    let startup = qpqual_cost.startup;
    (0, startup, startup + CPU_TUPLE_COST + qpqual_cost.per_tuple)
}

/// IndexCost is what the costs of a lookup in a system catalog's index know of the index: its size, whether a scan
/// of it reads nothing else, how closely its order follows the table's, and how many clauses it searches by.
pub struct IndexCost {
    pub pages: f64,
    pub tuples: f64,
    pub tree_height: f64,
    /// Whether the scan reads nothing but the index, which Postgres calls an index-only scan.
    pub indexonly: bool,
    /// The correlation between the index's order and the order of the table's rows, as btcostestimate takes it from
    /// the statistics of its first column.
    pub correlation: f64,
    pub nquals: usize,
}

/// cost_catalog_lookup returns the costs of a lookup in a system catalog's index that reads about
/// `num_index_tuples` of the index's entries for each of `loop_count` runs, and tests the clauses it does not search
/// by, by Postgres' cost_index and btcostestimate formulas.
pub fn cost_catalog_lookup(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    index: &IndexCost,
    num_index_tuples: f64,
    qpqual_cost: QualCost,
    loop_count: f64,
) -> Costs {
    let rel = &root.rels[rel];
    let num_index_tuples = num_index_tuples.min(index.tuples).max(1.0);
    let index_selectivity = (num_index_tuples / rel.tuples.max(1.0)).min(1.0);
    let num_index_pages = match index.pages > 1.0 && index.tuples > 1.0 {
        true => (num_index_tuples * index.pages / index.tuples).ceil(),
        false => 1.0,
    };
    let mut index_total_cost = match loop_count > 1.0 {
        true => {
            let fetched = index_pages_fetched(root, num_index_pages * loop_count, index.pages, index.pages);
            fetched * RANDOM_PAGE_COST / loop_count
        }
        false => num_index_pages * RANDOM_PAGE_COST,
    };
    index_total_cost += num_index_tuples * (CPU_INDEX_TUPLE_COST + CPU_OPERATOR_COST * index.nquals as f64);
    let mut index_startup_cost = 0.0;
    if index.tuples > 1.0 {
        let descent = index.tuples.log2().ceil() * CPU_OPERATOR_COST;
        index_startup_cost += descent;
        index_total_cost += descent;
    }
    let descent = (index.tree_height + 1.0) * 50.0 * CPU_OPERATOR_COST;
    index_startup_cost += descent;
    index_total_cost += descent;
    let tuples_fetched = clamp_row_est(index_selectivity * rel.tuples);
    let (max_io_cost, min_io_cost) = match (index.indexonly, loop_count > 1.0) {
        (true, _) => (0.0, 0.0),
        (false, true) => {
            let fetched = index_pages_fetched(root, tuples_fetched * loop_count, rel.pages, index.pages);
            let pages = (index_selectivity * rel.pages).ceil();
            let min_fetched = index_pages_fetched(root, pages * loop_count, rel.pages, index.pages);
            (fetched * RANDOM_PAGE_COST / loop_count, min_fetched * RANDOM_PAGE_COST / loop_count)
        }
        (false, false) => {
            let fetched = index_pages_fetched(root, tuples_fetched, rel.pages, index.pages);
            let pages = (index_selectivity * rel.pages).ceil();
            let min_io_cost = match pages > 0.0 {
                true => RANDOM_PAGE_COST + (pages - 1.0).max(0.0) * SEQ_PAGE_COST,
                false => 0.0,
            };
            (fetched * RANDOM_PAGE_COST, min_io_cost)
        }
    };
    let csquared = index.correlation * index.correlation;
    let run_cost = index_total_cost - index_startup_cost + max_io_cost + csquared * (min_io_cost - max_io_cost);
    let enabled = if index.indexonly { root.enables.indexonlyscan } else { root.enables.indexscan };
    let startup_cost = index_startup_cost + qpqual_cost.startup;
    let cpu_run_cost = (CPU_TUPLE_COST + qpqual_cost.per_tuple) * tuples_fetched;
    (disabled(enabled), startup_cost, startup_cost + run_cost + cpu_run_cost)
}

/// cost_index returns the costs of an index path, the share of the index's entries that it reads, and the cost of
/// reading the index itself, given the join clauses that its parameterization adds, the rows it returns, and how many
/// times a nested loop runs it, as Postgres' function of the same name does with btcostestimate's costs of the index
/// itself. Dolt has no visibility map, as though every page were all-visible, so an index-only scan reads no table
/// pages.
pub fn cost_index(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    path: &super::nodes::IndexPath,
    ppi_clauses: &[RinfoId],
    rows: f64,
    loop_count: f64,
) -> (Costs, f64, f64) {
    let baserel = &root.rels[rel];
    let index = &baserel.indexlist[path.index];
    let mut qpquals = extract_nonindex_conditions(root, &index.indrestrictinfo, &path.indexclauses);
    qpquals.extend(extract_nonindex_conditions(root, ppi_clauses, &path.indexclauses));
    let disabled_nodes = disabled(root.enables.indexscan);
    let estimate = super::selfuncs::btcostestimate(root, rel, path, loop_count);
    let (mut startup_cost, mut run_cost) = (estimate.startup_cost, estimate.total_cost - estimate.startup_cost);
    let tuples_fetched = clamp_row_est(estimate.selectivity * baserel.tuples);
    let allvisfrac = if path.indexonly { 1.0 } else { 0.0 };
    let (max_io_cost, min_io_cost);
    if loop_count > 1.0 {
        let pages_fetched = index_pages_fetched(root, tuples_fetched * loop_count, baserel.pages, index.pages);
        let pages_fetched = (pages_fetched * (1.0 - allvisfrac)).ceil();
        max_io_cost = pages_fetched * RANDOM_PAGE_COST / loop_count;
        let pages_fetched = (estimate.selectivity * baserel.pages).ceil();
        let pages_fetched = index_pages_fetched(root, pages_fetched * loop_count, baserel.pages, index.pages);
        let pages_fetched = (pages_fetched * (1.0 - allvisfrac)).ceil();
        min_io_cost = pages_fetched * RANDOM_PAGE_COST / loop_count;
    } else {
        let pages_fetched = index_pages_fetched(root, tuples_fetched, baserel.pages, index.pages);
        let pages_fetched = (pages_fetched * (1.0 - allvisfrac)).ceil();
        max_io_cost = pages_fetched * RANDOM_PAGE_COST;
        let pages_fetched = ((estimate.selectivity * baserel.pages).ceil() * (1.0 - allvisfrac)).ceil();
        min_io_cost = match pages_fetched > 0.0 {
            true => RANDOM_PAGE_COST + (pages_fetched - 1.0).max(0.0) * SEQ_PAGE_COST,
            false => 0.0,
        };
    }
    let csquared = estimate.correlation * estimate.correlation;
    run_cost += max_io_cost + csquared * (min_io_cost - max_io_cost);
    let qpqual_cost = cost_qual_eval(root, &qpquals);
    startup_cost += qpqual_cost.startup;
    let cpu_per_tuple = CPU_TUPLE_COST + qpqual_cost.per_tuple;
    let mut cpu_run_cost = cpu_per_tuple * tuples_fetched;
    startup_cost += baserel.reltarget.cost.startup;
    cpu_run_cost += baserel.reltarget.cost.per_tuple * rows;
    run_cost += cpu_run_cost;
    ((disabled_nodes, startup_cost, startup_cost + run_cost), estimate.selectivity, estimate.total_cost)
}

/// cost_bitmap_heap_scan returns the costs and rows of a scan of a base relation's rows whose keys a tree of index
/// scans finds, given its parameterization and how many times a nested loop runs it, as Postgres' function of the
/// same name does. Dolt's rows lie in its primary index's pages in key order, which the scan reads in that order.
pub fn cost_bitmap_heap_scan(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    ppi: Option<&super::nodes::ParamPathInfo>,
    bitmapqual: &Path,
    loop_count: f64,
) -> (Costs, f64) {
    let baserel = &root.rels[rel];
    let rows = ppi.map_or(baserel.rows, |ppi| ppi.ppi_rows);
    let (pages_fetched, index_total_cost, tuples_fetched) = compute_bitmap_pages(root, rel, bitmapqual, loop_count);
    let mut startup_cost = index_total_cost;
    let t = if baserel.pages > 1.0 { baserel.pages } else { 1.0 };
    let cost_per_page = match pages_fetched >= 2.0 {
        true => RANDOM_PAGE_COST - (RANDOM_PAGE_COST - SEQ_PAGE_COST) * (pages_fetched / t).sqrt(),
        false => RANDOM_PAGE_COST,
    };
    let mut run_cost = pages_fetched * cost_per_page;
    let qpqual_cost = get_restriction_qual_cost(root, rel, ppi);
    startup_cost += qpqual_cost.startup;
    run_cost += (CPU_TUPLE_COST + qpqual_cost.per_tuple) * tuples_fetched;
    startup_cost += baserel.reltarget.cost.startup;
    run_cost += baserel.reltarget.cost.per_tuple * rows;
    ((disabled(root.enables.bitmapscan), startup_cost, startup_cost + run_cost), rows)
}

/// get_restriction_qual_cost returns the cost of testing a base relation's restrictions and the join clauses that a
/// parameterization adds, as Postgres' function of the same name does.
fn get_restriction_qual_cost(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    ppi: Option<&super::nodes::ParamPathInfo>,
) -> QualCost {
    let mut cost = cost_qual_eval(root, &root.rels[rel].baserestrictinfo);
    if let Some(ppi) = ppi {
        let more = cost_qual_eval(root, &ppi.ppi_clauses);
        cost.startup += more.startup;
        cost.per_tuple += more.per_tuple;
    }
    cost
}

/// cost_bitmap_tree_node returns the cost of a node of a tree of index scans and the share of the table's rows whose
/// keys it finds, as Postgres' function of the same name does.
pub fn cost_bitmap_tree_node(path: &Path) -> (f64, f64) {
    match &path.kind {
        PathKind::IndexScan(ipath) => {
            (ipath.indextotalcost + 0.1 * CPU_OPERATOR_COST * path.rows, ipath.indexselectivity)
        }
        PathKind::BitmapAnd(bpath) | PathKind::BitmapOr(bpath) => (path.total_cost, bpath.bitmapselectivity),
        _ => unreachable!("a bitmap tree holds index scans, BitmapAnds, and BitmapOrs"),
    }
}

/// cost_bitmap_and_node returns the cost of a BitmapAnd of index scans and the share of the table's rows whose keys
/// it finds, as Postgres' function of the same name does.
pub fn cost_bitmap_and_node(bitmapquals: &[Rc<Path>]) -> (f64, f64) {
    let (mut total_cost, mut selec) = (0.0, 1.0);
    for (i, subpath) in bitmapquals.iter().enumerate() {
        let (sub_cost, subselec) = cost_bitmap_tree_node(subpath);
        selec *= subselec;
        total_cost += sub_cost;
        if i > 0 {
            total_cost += 100.0 * CPU_OPERATOR_COST;
        }
    }
    (total_cost, selec)
}

/// cost_bitmap_or_node returns the cost of a BitmapOr of index scans and the share of the table's rows whose keys it
/// finds, as Postgres' function of the same name does.
pub fn cost_bitmap_or_node(bitmapquals: &[Rc<Path>]) -> (f64, f64) {
    let (mut total_cost, mut selec) = (0.0, 0.0);
    for (i, subpath) in bitmapquals.iter().enumerate() {
        let (sub_cost, subselec) = cost_bitmap_tree_node(subpath);
        selec += subselec;
        total_cost += sub_cost;
        if i > 0 && !matches!(subpath.kind, PathKind::IndexScan(_)) {
            total_cost += 100.0 * CPU_OPERATOR_COST;
        }
    }
    (total_cost, f64::min(selec, 1.0))
}

/// tbm_calculate_entries returns how many pages a bitmap of the given size holds before it marks whole pages lossy,
/// as Postgres' function of the same name estimates it with its page table entries of 64 bytes.
fn tbm_calculate_entries(maxbytes: f64) -> f64 {
    (maxbytes / 64.0).floor().clamp(16.0, f64::from(i32::MAX - 1))
}

/// get_indexpath_pages returns the pages of the indexes that a tree of index scans reads, as Postgres' function of
/// the same name counts them.
fn get_indexpath_pages(root: &PlannerInfo<'_, '_>, bitmapqual: &Path) -> f64 {
    match &bitmapqual.kind {
        PathKind::IndexScan(ipath) => root.rels[bitmapqual.parent].indexlist[ipath.index].pages,
        PathKind::BitmapAnd(bpath) | PathKind::BitmapOr(bpath) => {
            bpath.bitmapquals.iter().map(|p| get_indexpath_pages(root, p)).sum()
        }
        _ => 0.0,
    }
}

/// compute_bitmap_pages returns how many of a base relation's pages a scan by a tree of index scans reads, the cost
/// of the tree, and how many rows the scan fetches, as Postgres' function of the same name estimates them.
pub fn compute_bitmap_pages(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    bitmapqual: &Path,
    loop_count: f64,
) -> (f64, f64, f64) {
    let baserel = &root.rels[rel];
    let (index_total_cost, index_selectivity) = cost_bitmap_tree_node(bitmapqual);
    let mut tuples_fetched = clamp_row_est(index_selectivity * baserel.tuples);
    let t = if baserel.pages > 1.0 { baserel.pages } else { 1.0 };
    let mut pages_fetched = (2.0 * t * tuples_fetched) / (2.0 * t + tuples_fetched);
    let heap_pages = pages_fetched.min(baserel.pages);
    let maxentries = tbm_calculate_entries(SORT_MEM);
    if loop_count > 1.0 {
        let index_pages = get_indexpath_pages(root, bitmapqual);
        pages_fetched = index_pages_fetched(root, tuples_fetched * loop_count, baserel.pages, index_pages);
        pages_fetched /= loop_count;
    }
    pages_fetched = if pages_fetched >= t { t } else { pages_fetched.ceil() };
    if maxentries < heap_pages {
        let lossy_pages = (heap_pages - maxentries / 2.0).max(0.0);
        let exact_pages = heap_pages - lossy_pages;
        if lossy_pages > 0.0 {
            tuples_fetched = clamp_row_est(
                index_selectivity * (exact_pages / heap_pages) * baserel.tuples
                    + (lossy_pages / heap_pages) * baserel.tuples,
            );
        }
    }
    (pages_fetched, index_total_cost, tuples_fetched)
}

/// extract_nonindex_conditions returns the clauses of a list that an index scan must test on each row because its
/// index clauses do not answer them, leaving out pseudoconstant ones, as Postgres' function of the same name does.
pub fn extract_nonindex_conditions(
    root: &PlannerInfo<'_, '_>,
    qual_clauses: &[RinfoId],
    indexclauses: &[super::nodes::IndexClause],
) -> Vec<RinfoId> {
    qual_clauses
        .iter()
        .copied()
        .filter(|&r| {
            !root.rinfos[r].pseudoconstant && !super::equivclass::is_redundant_with_indexclauses(root, r, indexclauses)
        })
        .collect()
}

/// index_pages_fetched returns how many pages a scan that fetches a number of rows from a table of a number of pages
/// reads, given the cache that the query's tables share, as Postgres' function of the same name estimates it by the
/// Mackert and Lohman formula.
pub fn index_pages_fetched(root: &PlannerInfo<'_, '_>, tuples_fetched: f64, pages: f64, index_pages: f64) -> f64 {
    let t = pages.max(1.0);
    let total_pages = (root.total_table_pages + index_pages).max(1.0);
    let b = match EFFECTIVE_CACHE_SIZE * t / total_pages {
        b if b <= 1.0 => 1.0,
        b => b.ceil(),
    };
    if t <= b {
        let fetched = (2.0 * t * tuples_fetched) / (2.0 * t + tuples_fetched);
        return if fetched >= t { t } else { fetched.ceil() };
    }
    let lim = (2.0 * t * b) / (2.0 * t - b);
    let fetched = match tuples_fetched <= lim {
        true => (2.0 * t * tuples_fetched) / (2.0 * t + tuples_fetched),
        false => b + (tuples_fetched - lim) * (t - b) / t,
    };
    fetched.ceil()
}

/// cost_material returns the costs of keeping a path's rows in memory as they are read, as Postgres' function of the
/// same name does.
pub fn cost_material(enables: &Enables, input: &Path) -> Costs {
    let mut run_cost = input.total_cost - input.startup_cost + 2.0 * CPU_OPERATOR_COST * input.rows;
    let nbytes = relation_byte_size(input.rows, input.width);
    if nbytes > HASH_MEM / 2.0 {
        run_cost += SEQ_PAGE_COST * (nbytes / BLCKSZ).ceil();
    }
    (input.disabled_nodes + disabled(enables.material), input.startup_cost, input.startup_cost + run_cost)
}

/// SORT_MEM is the memory that a sort may take, as Postgres' default work_mem.
const SORT_MEM: f64 = 4.0 * 1024.0 * 1024.0;

/// cost_sort returns the costs of sorting a path's rows, of which a LIMIT may read only some, as
/// Postgres' cost_sort and cost_tuplesort estimate them for an in-memory quicksort, a bounded heap sort, or an
/// external merge sort.
pub fn cost_sort(root: &PlannerInfo<'_, '_>, input: &Path, limit_tuples: f64) -> Costs {
    let tuples = input.rows.max(2.0);
    let comparison_cost = 2.0 * CPU_OPERATOR_COST;
    let input_bytes = relation_byte_size(tuples, input.width);
    let (output_tuples, output_bytes) = match limit_tuples > 0.0 && limit_tuples < tuples {
        true => (limit_tuples, relation_byte_size(limit_tuples, input.width)),
        false => (tuples, input_bytes),
    };
    let mut startup_cost = if output_bytes > SORT_MEM {
        let npages = (input_bytes / BLCKSZ).ceil();
        let nruns = input_bytes / SORT_MEM;
        let mergeorder = (SORT_MEM / (BLCKSZ * 2.0 + BLCKSZ * 32.0)).floor().clamp(6.0, 500.0);
        let log_runs = if nruns > mergeorder { (nruns.ln() / mergeorder.ln()).ceil() } else { 1.0 };
        let npageaccesses = 2.0 * npages * log_runs;
        comparison_cost * tuples * tuples.log2() + npageaccesses * (SEQ_PAGE_COST * 0.75 + RANDOM_PAGE_COST * 0.25)
    } else if tuples > 2.0 * output_tuples || input_bytes > SORT_MEM {
        comparison_cost * tuples * (2.0 * output_tuples).log2()
    } else {
        comparison_cost * tuples * tuples.log2()
    };
    startup_cost += input.total_cost;
    (input.disabled_nodes + disabled(root.enables.sort), startup_cost, startup_cost + CPU_OPERATOR_COST * tuples)
}

/// cost_rescan returns the startup and total costs of reading a path's rows again, as Postgres' function of the same
/// name does: kept rows cost little to read again, and a hash table or a function's rows need not be built again.
fn cost_rescan(path: &Path) -> (f64, f64) {
    match &path.kind {
        PathKind::Material(_) => (0.0, CPU_OPERATOR_COST * path.rows),
        PathKind::HashJoin(_) => (0.0, path.total_cost - path.startup_cost),
        _ => (path.startup_cost, path.total_cost),
    }
}

/// SemiAntiJoinFactors are the share of outer rows that find a match and the average matches of one, which a semi,
/// anti, or unique inner join stops reading after the first of, as Postgres' SemiAntiJoinFactors are.
#[derive(Clone, Copy, Debug, Default)]
pub struct SemiAntiJoinFactors {
    pub outer_match_frac: f64,
    pub match_count: f64,
}

/// JoinPathExtraData is what every path of a join of two relations shares, as Postgres' JoinPathExtraData holds it.
pub struct JoinPathExtraData {
    pub restrictlist: Vec<RinfoId>,
    pub inner_unique: bool,
    pub semifactors: SemiAntiJoinFactors,
}

/// cost_nestloop returns the costs of a nested loop of an inner path over an outer one, as
/// Postgres' initial_cost_nestloop and final_cost_nestloop do, given the clauses that the join tests and whether the
/// inner path looks its rows up by the join's clauses.
pub fn cost_nestloop(
    root: &PlannerInfo<'_, '_>,
    jointype: JoinType,
    outer: &Path,
    inner: &Path,
    extra: &JoinPathExtraData,
    joinrestrictinfo: &[RinfoId],
    has_indexed_join_quals: bool,
) -> Costs {
    let (inner_rescan_start_cost, inner_rescan_total_cost) = cost_rescan(inner);
    let disabled_nodes = disabled(root.enables.nestloop) + inner.disabled_nodes + outer.disabled_nodes;
    let mut startup_cost = outer.startup_cost + inner.startup_cost;
    let mut run_cost = outer.total_cost - outer.startup_cost;
    if outer.rows > 1.0 {
        run_cost += (outer.rows - 1.0) * inner_rescan_start_cost;
    }
    let inner_run_cost = inner.total_cost - inner.startup_cost;
    let inner_rescan_run_cost = inner_rescan_total_cost - inner_rescan_start_cost;
    let inner_rows = inner.rows.max(1.0);
    let ntuples = if matches!(jointype, JoinType::Semi | JoinType::Anti) || extra.inner_unique {
        let mut outer_matched_rows = (outer.rows * extra.semifactors.outer_match_frac).round();
        let mut outer_unmatched_rows = outer.rows - outer_matched_rows;
        let inner_scan_frac = 2.0 / (extra.semifactors.match_count + 1.0);
        let mut ntuples = outer_matched_rows * inner_rows * inner_scan_frac;
        if has_indexed_join_quals {
            run_cost += inner_run_cost * inner_scan_frac;
            if outer_matched_rows > 1.0 {
                run_cost += (outer_matched_rows - 1.0) * inner_rescan_run_cost * inner_scan_frac;
            }
            run_cost += outer_unmatched_rows * inner_rescan_run_cost / inner_rows;
        } else {
            ntuples += outer_unmatched_rows * inner_rows;
            run_cost += inner_run_cost;
            if outer_unmatched_rows >= 1.0 {
                outer_unmatched_rows -= 1.0;
            } else {
                outer_matched_rows -= 1.0;
            }
            if outer_matched_rows > 0.0 {
                run_cost += outer_matched_rows * inner_rescan_run_cost * inner_scan_frac;
            }
            if outer_unmatched_rows > 0.0 {
                run_cost += outer_unmatched_rows * inner_rescan_run_cost;
            }
        }
        ntuples
    } else {
        run_cost += inner_run_cost;
        if outer.rows > 1.0 {
            run_cost += (outer.rows - 1.0) * inner_rescan_run_cost;
        }
        outer.rows.max(1.0) * inner_rows
    };
    let restrict_qual_cost = cost_qual_eval(root, joinrestrictinfo);
    startup_cost += restrict_qual_cost.startup;
    run_cost += (CPU_TUPLE_COST + restrict_qual_cost.per_tuple) * ntuples;
    (disabled_nodes, startup_cost, startup_cost + run_cost)
}

/// has_indexed_join_quals reports whether a nested loop tests no clauses of its own and its inner index path, or bitmap
/// scan of one index, searches by each join clause that it is parameterized by, as Postgres' function of the same name
/// does.
pub fn has_indexed_join_quals(
    root: &mut PlannerInfo<'_, '_>,
    joinrelids: &Relids,
    inner: &Path,
    joinrestrictinfo: &[RinfoId],
) -> bool {
    if !joinrestrictinfo.is_empty() {
        return false;
    }
    let Some(param_info) = super::relnode::get_baserel_parampathinfo(root, inner.parent, &inner.param) else {
        return false;
    };
    let index_path = match &inner.kind {
        PathKind::IndexScan(index_path) => index_path,
        PathKind::BitmapHeapScan(bitmapqual) => match &bitmapqual.kind {
            PathKind::IndexScan(index_path) => index_path,
            _ => return false,
        },
        _ => return false,
    };
    let mut found_one = false;
    for rinfo in param_info.ppi_clauses {
        if join_clause_is_movable_into(&root.rinfos[rinfo], &inner.relids, joinrelids) {
            if !super::equivclass::is_redundant_with_indexclauses(root, rinfo, &index_path.indexclauses) {
                return false;
            }
            found_one = true;
        }
    }
    found_one
}

/// exec_choose_hash_table_size returns the buckets and batches of a hash table of a number of rows of a width, as
/// Postgres' ExecChooseHashTableSize chooses them without skew optimization.
fn exec_choose_hash_table_size(ntuples: f64, width: f64) -> (f64, f64) {
    let tupsize = 16.0 + 16.0 + maxalign(width);
    let inner_rel_bytes = ntuples * tupsize;
    let max_pointers = (HASH_MEM / 8.0).floor();
    let dbuckets = ntuples.ceil().min(max_pointers);
    let nbuckets = dbuckets.max(1.0).log2().ceil().exp2().max(1024.0);
    if inner_rel_bytes + 8.0 * nbuckets <= HASH_MEM {
        return (nbuckets, 1.0);
    }
    let bucket_size = tupsize + 8.0;
    let nbuckets = (HASH_MEM / bucket_size).log2().floor().exp2().min(max_pointers).max(1024.0);
    let dbatch = (inner_rel_bytes / (HASH_MEM - 8.0 * nbuckets)).ceil().min(max_pointers);
    (nbuckets, dbatch.max(1.0).log2().ceil().exp2())
}

/// cost_hashjoin returns the costs of a hash join that hashes the inner path's rows by the hash
/// clauses and probes them with the outer path's, as Postgres' initial_cost_hashjoin and final_cost_hashjoin do.
pub fn cost_hashjoin(
    root: &PlannerInfo<'_, '_>,
    jointype: JoinType,
    hashclauses: &[RinfoId],
    outer: &Path,
    inner: &Path,
    extra: &JoinPathExtraData,
) -> Costs {
    let num_hashclauses = hashclauses.len() as f64;
    let disabled_nodes = disabled(root.enables.hashjoin) + inner.disabled_nodes + outer.disabled_nodes;
    let mut startup_cost = outer.startup_cost + inner.total_cost;
    let mut run_cost = outer.total_cost - outer.startup_cost;
    startup_cost += (CPU_OPERATOR_COST * num_hashclauses + CPU_TUPLE_COST) * inner.rows;
    run_cost += CPU_OPERATOR_COST * num_hashclauses * outer.rows;
    let (numbuckets, numbatches) = exec_choose_hash_table_size(inner.rows, inner.width);
    if numbatches > 1.0 {
        let (outerpages, innerpages) = (page_size(outer.rows, outer.width), page_size(inner.rows, inner.width));
        startup_cost += SEQ_PAGE_COST * innerpages;
        run_cost += SEQ_PAGE_COST * (innerpages + 2.0 * outerpages);
    }
    let virtualbuckets = numbuckets * numbatches;
    let (mut innerbucketsize, mut innermcvfreq) = (1.0f64, 1.0f64);
    for &rinfo in hashclauses {
        let rinfo = &root.rinfos[rinfo];
        let Expr::Compare(_, l, r) = &rinfo.clause else { continue };
        let key = if rinfo.right_relids.is_subset(&inner.relids) { r } else { l };
        let (mcvfreq, bucketsize) = super::selfuncs::estimate_hash_bucket_stats(root, key, virtualbuckets);
        innerbucketsize = innerbucketsize.min(bucketsize);
        innermcvfreq = innermcvfreq.min(mcvfreq);
    }
    if relation_byte_size(clamp_row_est(inner.rows * innermcvfreq), inner.width) > HASH_MEM {
        startup_cost += DISABLE_COST;
    }
    let hash_qual_cost = cost_qual_eval(root, hashclauses);
    let mut qp_qual_cost = cost_qual_eval(root, &extra.restrictlist);
    qp_qual_cost.startup -= hash_qual_cost.startup;
    qp_qual_cost.per_tuple -= hash_qual_cost.per_tuple;
    startup_cost += hash_qual_cost.startup;
    let hashjointuples = if matches!(jointype, JoinType::Semi | JoinType::Anti) || extra.inner_unique {
        let outer_matched_rows = (outer.rows * extra.semifactors.outer_match_frac).round();
        let inner_scan_frac = 2.0 / (extra.semifactors.match_count + 1.0);
        run_cost += hash_qual_cost.per_tuple
            * outer_matched_rows
            * clamp_row_est(inner.rows * innerbucketsize * inner_scan_frac)
            * 0.5;
        run_cost += hash_qual_cost.per_tuple
            * (outer.rows - outer_matched_rows)
            * clamp_row_est(inner.rows / virtualbuckets)
            * 0.05;
        match jointype {
            JoinType::Anti => outer.rows - outer_matched_rows,
            _ => outer_matched_rows,
        }
    } else {
        run_cost += hash_qual_cost.per_tuple * outer.rows * clamp_row_est(inner.rows * innerbucketsize) * 0.5;
        approx_tuple_count(root, outer, inner, hashclauses)
    };
    startup_cost += qp_qual_cost.startup;
    run_cost += (CPU_TUPLE_COST + qp_qual_cost.per_tuple) * hashjointuples;
    (disabled_nodes, startup_cost, startup_cost + run_cost)
}

/// compute_semi_anti_join_factors returns the share of outer rows that find a match in a join and how many they find
/// on average, as Postgres' function of the same name estimates them.
pub fn compute_semi_anti_join_factors(
    root: &PlannerInfo<'_, '_>,
    joinrel: usize,
    outerrel: usize,
    innerrel: usize,
    jointype: JoinType,
    sjinfo: &SpecialJoinInfo,
    restrictlist: &[RinfoId],
) -> SemiAntiJoinFactors {
    let joinrelids = &root.rels[joinrel].relids;
    let joinquals: Vec<RinfoId> = match jointype.is_outer() {
        true => restrictlist.iter().copied().filter(|&r| !rinfo_is_pushed_down(&root.rinfos[r], joinrelids)).collect(),
        false => restrictlist.to_vec(),
    };
    let selec_type = if jointype == JoinType::Anti { JoinType::Anti } else { JoinType::Semi };
    let jselec = clauselist_selectivity(root, &joinquals, 0, selec_type, Some(sjinfo));
    let norm_sjinfo = super::joinrels::init_dummy_sjinfo(&root.rels[outerrel].relids, &root.rels[innerrel].relids);
    let nselec = clauselist_selectivity(root, &joinquals, 0, JoinType::Inner, Some(&norm_sjinfo));
    let inner_rows = root.rels[innerrel].rows;
    let match_count = if jselec > 0.0 { (nselec * inner_rows / jselec).max(1.0) } else { 1.0 };
    SemiAntiJoinFactors { outer_match_frac: jselec, match_count }
}

/// approx_tuple_count returns about how many pairs of the two paths' rows pass a list of clauses, as an inner join
/// would, as Postgres' function of the same name estimates it.
fn approx_tuple_count(root: &PlannerInfo<'_, '_>, outer: &Path, inner: &Path, quals: &[RinfoId]) -> f64 {
    let sjinfo = super::joinrels::init_dummy_sjinfo(&outer.relids, &inner.relids);
    let selec: f64 = quals
        .iter()
        .map(|&q| {
            let q = &root.rinfos[q];
            clause_selectivity(root, &q.clause, Some(q), 0, JoinType::Inner, Some(&sjinfo))
        })
        .product();
    clamp_row_est(selec * outer.rows * inner.rows)
}

/// set_baserel_size_estimates estimates the rows of a base relation that its restrictions keep and the width of its
/// rows, as Postgres' function of the same name does.
pub fn set_baserel_size_estimates(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    let selectivity = clauselist_selectivity(root, &root.rels[rel].baserestrictinfo, 0, JoinType::Inner, None);
    root.rels[rel].rows = clamp_row_est(root.rels[rel].tuples * selectivity);
    set_rel_width(root, rel);
}

/// set_rel_width estimates the width of a base relation's rows from the columns and PlaceHolderVars of its target,
/// as Postgres' function of the same name does.
pub fn set_rel_width(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    let mut tuple_width = 0.0;
    for e in root.rels[rel].reltarget.exprs.clone() {
        let Expr::Column(id) = e else { continue };
        tuple_width += match root.glob.node(id) {
            super::nodes::VarNode::Var(var) => {
                let (oid, modifier) = match root.parse.rte(var.varno).table() {
                    Some(table) => (Some(table.columns[var.varattno].ty.oid), table.columns[var.varattno].ty.modifier),
                    None => (root.parse.rte(var.varno).coltypes[var.varattno], -1),
                };
                let width = get_typavgwidth(oid, modifier);
                root.rels[rel].attr_widths[var.varattno] = width;
                width
            }
            super::nodes::VarNode::PlaceHolderVar(phv) => {
                let i = root.placeholder_array[&phv.phid];
                root.placeholder_list[i].ph_width
            }
        };
    }
    root.rels[rel].reltarget.width = tuple_width;
}

/// get_parameterized_baserel_size returns the rows of a base relation that its restrictions and a parameterization's
/// join clauses keep, as Postgres' function of the same name does.
pub fn get_parameterized_baserel_size(root: &PlannerInfo<'_, '_>, rel: usize, param_clauses: &[RinfoId]) -> f64 {
    let allclauses: Vec<RinfoId> = param_clauses.iter().chain(&root.rels[rel].baserestrictinfo).copied().collect();
    let nrows =
        clamp_row_est(root.rels[rel].tuples * clauselist_selectivity(root, &allclauses, rel, JoinType::Inner, None));
    nrows.min(root.rels[rel].rows)
}

/// set_joinrel_size_estimates estimates the rows of a join relation, as Postgres' function of the same name does.
pub fn set_joinrel_size_estimates(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outer_rel: usize,
    inner_rel: usize,
    sjinfo: &SpecialJoinInfo,
    restrictlist: &[RinfoId],
) {
    let (outer_rows, inner_rows) = (root.rels[outer_rel].rows, root.rels[inner_rel].rows);
    root.rels[joinrel].rows = calc_joinrel_size_estimate(root, joinrel, outer_rows, inner_rows, sjinfo, restrictlist);
}

/// calc_joinrel_size_estimate returns about how many rows a join of two relations produces, given the rows of each
/// and the join's clauses, as Postgres' function of the same name estimates it.
fn calc_joinrel_size_estimate(
    root: &PlannerInfo<'_, '_>,
    joinrel: usize,
    outer_rows: f64,
    inner_rows: f64,
    sjinfo: &SpecialJoinInfo,
    restrictlist: &[RinfoId],
) -> f64 {
    let joinrelids = &root.rels[joinrel].relids;
    let jointype = sjinfo.jointype;
    let (jselec, pselec) = match jointype.is_outer() {
        true => {
            let (pushedquals, joinquals): (Vec<RinfoId>, Vec<RinfoId>) =
                restrictlist.iter().copied().partition(|&r| rinfo_is_pushed_down(&root.rinfos[r], joinrelids));
            (
                clauselist_selectivity(root, &joinquals, 0, jointype, Some(sjinfo)),
                clauselist_selectivity(root, &pushedquals, 0, jointype, Some(sjinfo)),
            )
        }
        false => (clauselist_selectivity(root, restrictlist, 0, jointype, Some(sjinfo)), 0.0),
    };
    let nrows = match jointype {
        JoinType::Inner => outer_rows * inner_rows * jselec,
        JoinType::Left | JoinType::Right => (outer_rows * inner_rows * jselec).max(outer_rows) * pselec,
        JoinType::Full => (outer_rows * inner_rows * jselec).max(outer_rows).max(inner_rows) * pselec,
        JoinType::Semi => outer_rows * jselec,
        JoinType::Anti => outer_rows * (1.0 - jselec) * pselec,
    };
    clamp_row_est(nrows)
}

/// APPEND_CPU_COST_MULTIPLIER is the share of a tuple's processing cost that an Append charges for each row it passes
/// on, as Postgres' constant of the same name is.
pub const APPEND_CPU_COST_MULTIPLIER: f64 = 0.5;

/// HASH_MEM_LIMIT is the memory that a hashed aggregation may take, Postgres' default work_mem times its default
/// hash_mem_multiplier.
const HASH_MEM_LIMIT: f64 = 4.0 * 1024.0 * 1024.0 * 2.0;

/// cost_tuplesort returns the startup and run costs of sorting rows of a width, of which a LIMIT may read only
/// some, as Postgres' function of the same name estimates them.
pub fn cost_tuplesort(tuples: f64, width: f64, comparison_cost: f64, limit_tuples: f64) -> (f64, f64) {
    let tuples = tuples.max(2.0);
    let comparison_cost = comparison_cost + 2.0 * CPU_OPERATOR_COST;
    let input_bytes = relation_byte_size(tuples, width);
    let (output_tuples, output_bytes) = match limit_tuples > 0.0 && limit_tuples < tuples {
        true => (limit_tuples, relation_byte_size(limit_tuples, width)),
        false => (tuples, input_bytes),
    };
    let startup_cost = if output_bytes > SORT_MEM {
        let npages = (input_bytes / BLCKSZ).ceil();
        let nruns = input_bytes / SORT_MEM;
        let mergeorder = (SORT_MEM / (BLCKSZ * 2.0 + BLCKSZ * 32.0)).floor().clamp(6.0, 500.0);
        let log_runs = if nruns > mergeorder { (nruns.ln() / mergeorder.ln()).ceil() } else { 1.0 };
        let npageaccesses = 2.0 * npages * log_runs;
        comparison_cost * tuples * tuples.log2() + npageaccesses * (SEQ_PAGE_COST * 0.75 + RANDOM_PAGE_COST * 0.25)
    } else if tuples > 2.0 * output_tuples || input_bytes > SORT_MEM {
        comparison_cost * tuples * (2.0 * output_tuples).log2()
    } else {
        comparison_cost * tuples * tuples.log2()
    };
    (startup_cost, CPU_OPERATOR_COST * tuples)
}

/// cost_incremental_sort returns the costs of sorting a path's rows by pathkeys whose leading keys order them
/// already, sorting each run of rows with equal leading keys, as Postgres' function of the same name does.
pub fn cost_incremental_sort(
    root: &mut PlannerInfo<'_, '_>,
    pathkeys: &[super::nodes::PkId],
    presorted_keys: usize,
    input: &Path,
    limit_tuples: f64,
) -> Costs {
    let input_run_cost = input.total_cost - input.startup_cost;
    let input_tuples = input.rows.max(2.0);
    let mut input_groups = input_tuples.min(200.0);
    let mut presorted_exprs = Vec::new();
    let mut unknown_varno = false;
    for &pk in pathkeys.iter().take(presorted_keys) {
        let ec = root.canon_pathkeys[pk].pk_eclass;
        let member = root.eq_classes[ec].ec_members[0];
        let expr = root.eq_members[member].em_expr.clone();
        if super::var::pull_varnos(root, &expr).is_empty() {
            unknown_varno = true;
            break;
        }
        presorted_exprs.push(expr);
    }
    if !unknown_varno {
        input_groups = super::selfuncs::estimate_num_groups(root, &presorted_exprs, input_tuples, None);
    }
    let group_tuples = input_tuples / input_groups;
    let group_input_run_cost = input_run_cost / input_groups;
    let (group_startup_cost, group_run_cost) = cost_tuplesort(group_tuples, input.width, 0.0, limit_tuples);
    let startup_cost = group_startup_cost + input.startup_cost + group_input_run_cost;
    let mut run_cost = group_run_cost
        + (group_run_cost + group_startup_cost) * (input_groups - 1.0)
        + group_input_run_cost * (input_groups - 1.0);
    run_cost += (CPU_TUPLE_COST + 0.0) * input_tuples;
    run_cost += 2.0 * CPU_TUPLE_COST * input_groups;
    (input.disabled_nodes, startup_cost, startup_cost + run_cost)
}

/// hash_agg_entry_size returns the memory that one group of a hashed aggregation takes, as Postgres' function of the
/// same name estimates it for its transition states and grouped tuple.
fn hash_agg_entry_size(num_trans: usize, tuple_width: f64, transition_space: f64) -> f64 {
    const CHUNKHDRSZ: f64 = 8.0;
    let tuple_size = 16.0 + tuple_width;
    let pergroup_size = num_trans as f64 * 16.0;
    let tuple_chunk_size = CHUNKHDRSZ + tuple_size;
    let pergroup_chunk_size = if pergroup_size > 0.0 { CHUNKHDRSZ + pergroup_size } else { 0.0 };
    let transition_chunk_size = if transition_space > 0.0 { CHUNKHDRSZ + transition_space } else { 0.0 };
    24.0 + tuple_chunk_size + pergroup_chunk_size + transition_chunk_size
}

/// hash_agg_set_limits returns the memory limit, the most groups, and the partitions that a hashed aggregation of
/// groups of an entry size spills into, as Postgres' function of the same name chooses them.
fn hash_agg_set_limits(hashentrysize: f64, input_groups: f64) -> (f64, f64, f64) {
    if input_groups * hashentrysize <= HASH_MEM_LIMIT {
        return (HASH_MEM_LIMIT, (HASH_MEM_LIMIT / hashentrysize).floor(), 0.0);
    }
    let npartitions = hash_choose_num_partitions(input_groups, hashentrysize);
    let partition_mem = BLCKSZ + BLCKSZ * npartitions;
    let mem_limit = match HASH_MEM_LIMIT > 4.0 * partition_mem {
        true => HASH_MEM_LIMIT - partition_mem,
        false => HASH_MEM_LIMIT * 0.75,
    };
    let ngroups_limit = if mem_limit > hashentrysize { (mem_limit / hashentrysize).floor() } else { 1.0 };
    (mem_limit, ngroups_limit, npartitions)
}

/// hash_choose_num_partitions returns how many partitions a hashed aggregation that spills writes, a power of two,
/// as Postgres' function of the same name chooses it.
fn hash_choose_num_partitions(input_groups: f64, hashentrysize: f64) -> f64 {
    let partition_limit = (HASH_MEM_LIMIT * 0.25) / BLCKSZ;
    let mem_wanted = 1.50 * input_groups * hashentrysize;
    let dpartitions = (1.0 + mem_wanted / HASH_MEM_LIMIT).min(partition_limit).clamp(4.0, 1024.0);
    let partition_bits = (dpartitions.floor() as u32).next_power_of_two().trailing_zeros().min(32);
    f64::from(1u32 << partition_bits)
}

/// cost_agg returns the rows and costs of an aggregation of rows with a cost, as Postgres' function of the same name
/// estimates them: a plain aggregation returns one row after reading them all, a sorted one returns each group as
/// it ends, and a hashed one returns the groups after reading every row, spilling those beyond its memory.
#[allow(clippy::too_many_arguments)]
pub fn cost_agg(
    root: &PlannerInfo<'_, '_>,
    aggstrategy: super::nodes::AggStrategy,
    aggcosts: &super::prepagg::AggClauseCosts,
    num_group_cols: usize,
    num_groups: f64,
    quals: &[Expr],
    (mut disabled_nodes, input_startup_cost, input_total_cost): Costs,
    input_tuples: f64,
    input_width: f64,
) -> (f64, Costs) {
    use super::nodes::AggStrategy;
    let (mut startup_cost, mut total_cost, mut output_tuples);
    match aggstrategy {
        AggStrategy::Plain => {
            startup_cost = input_total_cost
                + aggcosts.trans_cost.startup
                + aggcosts.trans_cost.per_tuple * input_tuples
                + aggcosts.final_cost.startup
                + aggcosts.final_cost.per_tuple;
            total_cost = startup_cost + CPU_TUPLE_COST;
            output_tuples = 1.0;
        }
        AggStrategy::Sorted | AggStrategy::Mixed => {
            startup_cost = input_startup_cost;
            total_cost = input_total_cost;
            if aggstrategy == AggStrategy::Mixed && !root.enables.hashagg {
                disabled_nodes += 1;
            }
            total_cost += aggcosts.trans_cost.startup + aggcosts.trans_cost.per_tuple * input_tuples;
            total_cost += CPU_OPERATOR_COST * num_group_cols as f64 * input_tuples;
            total_cost += aggcosts.final_cost.startup + aggcosts.final_cost.per_tuple * num_groups;
            total_cost += CPU_TUPLE_COST * num_groups;
            output_tuples = num_groups;
        }
        AggStrategy::Hashed => {
            startup_cost = input_total_cost;
            if !root.enables.hashagg {
                disabled_nodes += 1;
            }
            startup_cost += aggcosts.trans_cost.startup + aggcosts.trans_cost.per_tuple * input_tuples;
            startup_cost += CPU_OPERATOR_COST * num_group_cols as f64 * input_tuples;
            startup_cost += aggcosts.final_cost.startup;
            total_cost = startup_cost + aggcosts.final_cost.per_tuple * num_groups + CPU_TUPLE_COST * num_groups;
            output_tuples = num_groups;
        }
    }
    if matches!(aggstrategy, AggStrategy::Hashed | AggStrategy::Mixed) {
        let hashentrysize = hash_agg_entry_size(root.parse.aggregates.len(), input_width, aggcosts.transition_space);
        let (mem_limit, ngroups_limit, num_partitions) = hash_agg_set_limits(hashentrysize, num_groups);
        let nbatches = ((num_groups * hashentrysize) / mem_limit).max(num_groups / ngroups_limit).ceil().max(1.0);
        let num_partitions = num_partitions.max(2.0);
        let depth = (nbatches.ln() / num_partitions.ln()).ceil();
        let pages = relation_byte_size(input_tuples, input_width) / BLCKSZ;
        let pages_written = pages * depth * 2.0;
        let pages_read = pages * depth * 2.0;
        startup_cost += pages_written * RANDOM_PAGE_COST;
        total_cost += pages_written * RANDOM_PAGE_COST + pages_read * SEQ_PAGE_COST;
        let spill_cost = depth * input_tuples * 2.0 * CPU_TUPLE_COST;
        startup_cost += spill_cost;
        total_cost += spill_cost;
    }
    if !quals.is_empty() {
        let qual_cost = quals.iter().fold(QualCost::default(), |c, q| {
            let one = cost_qual_eval_node(q);
            QualCost { startup: c.startup + one.startup, per_tuple: c.per_tuple + one.per_tuple }
        });
        startup_cost += qual_cost.startup;
        total_cost += qual_cost.startup + output_tuples * qual_cost.per_tuple;
        let clauses: Vec<(&Expr, Option<&super::nodes::RestrictInfo>)> = quals.iter().map(|q| (q, None)).collect();
        output_tuples =
            clamp_row_est(output_tuples * clausesel::list_selectivity(root, &clauses, 0, JoinType::Inner, None));
    }
    (output_tuples, (disabled_nodes, startup_cost, total_cost))
}

/// cost_group returns the rows and costs of a GROUP BY without aggregates over sorted rows, as Postgres' function of
/// the same name estimates them.
pub fn cost_group(
    root: &PlannerInfo<'_, '_>,
    num_group_cols: usize,
    num_groups: f64,
    quals: &[Expr],
    (disabled_nodes, input_startup_cost, input_total_cost): Costs,
    input_tuples: f64,
) -> (f64, Costs) {
    let mut output_tuples = num_groups;
    let mut startup_cost = input_startup_cost;
    let mut total_cost = input_total_cost + CPU_OPERATOR_COST * input_tuples * num_group_cols as f64;
    if !quals.is_empty() {
        let qual_cost = quals.iter().fold(QualCost::default(), |c, q| {
            let one = cost_qual_eval_node(q);
            QualCost { startup: c.startup + one.startup, per_tuple: c.per_tuple + one.per_tuple }
        });
        startup_cost += qual_cost.startup;
        total_cost += qual_cost.startup + output_tuples * qual_cost.per_tuple;
        let clauses: Vec<(&Expr, Option<&super::nodes::RestrictInfo>)> = quals.iter().map(|q| (q, None)).collect();
        output_tuples =
            clamp_row_est(output_tuples * clausesel::list_selectivity(root, &clauses, 0, JoinType::Inner, None));
    }
    (output_tuples, (disabled_nodes, startup_cost, total_cost))
}

/// cost_windowagg returns the costs of a window's calls over rows with a cost, as Postgres' function of the same name
/// estimates them, where a call costs its function and arguments for each row and the window compares its keys.
pub fn cost_windowagg(
    root: &mut PlannerInfo<'_, '_>,
    calls: &[usize],
    (disabled_nodes, input_startup_cost, input_total_cost): Costs,
    input_tuples: f64,
) -> Costs {
    let winclause = &root.parse.window_funcs[calls[0]];
    let num_part_cols = winclause.partition.len();
    let num_order_cols = winclause.order.len();
    let mut startup_cost = input_startup_cost;
    let mut total_cost = input_total_cost;
    for &k in calls {
        let call = &root.parse.window_funcs[k];
        let mut wfunccost = CPU_OPERATOR_COST;
        for arg in call.args.iter().chain(&call.filter) {
            let cost = cost_qual_eval_node(arg);
            startup_cost += cost.startup;
            wfunccost += cost.per_tuple;
        }
        total_cost += wfunccost * input_tuples;
    }
    total_cost += CPU_OPERATOR_COST * (num_part_cols + num_order_cols) as f64 * input_tuples;
    total_cost += CPU_TUPLE_COST * input_tuples;
    let startup_tuples = get_windowclause_startup_tuples(root, calls[0], input_tuples);
    if startup_tuples > 1.0 {
        startup_cost += (total_cost - startup_cost) / input_tuples * (startup_tuples - 1.0);
    }
    (disabled_nodes, startup_cost, total_cost)
}

/// get_windowclause_startup_tuples estimates how many rows a window reads before it returns its first, from its
/// partition's and peer group's estimated rows and its frame's end, as Postgres' function of the same name does.
fn get_windowclause_startup_tuples(root: &mut PlannerInfo<'_, '_>, call: usize, input_tuples: f64) -> f64 {
    use crate::window::frame;
    let wc = root.parse.window_funcs[call].clone();
    let partition_tuples = match wc.partition.is_empty() {
        true => input_tuples,
        false => input_tuples / super::selfuncs::estimate_num_groups(root, &wc.partition, input_tuples, None),
    };
    let peer_tuples = match wc.order.is_empty() {
        true => 1.0,
        false => {
            let orderexprs: Vec<Expr> = wc.order.iter().map(|k| k.expr.clone()).collect();
            partition_tuples / super::selfuncs::estimate_num_groups(root, &orderexprs, partition_tuples, None)
        }
    };
    let options = wc.options;
    let return_tuples = if options & frame::END_UNBOUNDED_FOLLOWING != 0 {
        partition_tuples
    } else if options & frame::END_OFFSET_PRECEDING != 0 {
        1.0
    } else if options & frame::END_OFFSET_FOLLOWING != 0 {
        let end_offset_value = match &wc.end {
            crate::window::Bound::Following(Expr::Const(v)) => match v {
                Value::Null => 1.0,
                Value::Int2(_) | Value::Int4(_) | Value::Int8(_) => v.to_i64().unwrap_or(1) as f64,
                _ => partition_tuples / peer_tuples * clausesel::DEFAULT_INEQ_SEL,
            },
            _ => partition_tuples / peer_tuples * clausesel::DEFAULT_INEQ_SEL,
        };
        match options & frame::ROWS != 0 {
            true => end_offset_value + 1.0,
            false => peer_tuples * (end_offset_value + 1.0),
        }
    } else if options & frame::ROWS != 0 {
        1.0
    } else if wc.order.is_empty() {
        partition_tuples
    } else {
        peer_tuples
    };
    let return_tuples = match !wc.partition.is_empty() || !wc.order.is_empty() {
        true => (return_tuples + 1.0).min(partition_tuples),
        false => return_tuples.min(partition_tuples),
    };
    clamp_row_est(return_tuples)
}

/// set_pathtarget_cost_width sets the cost of evaluating a target's expressions and the width of its rows, as
/// Postgres' function of the same name does.
pub fn set_pathtarget_cost_width(root: &PlannerInfo<'_, '_>, target: &mut super::nodes::PathTarget) {
    let mut cost = QualCost::default();
    let mut tuple_width = 0.0;
    for expr in &target.exprs {
        tuple_width += get_expr_width(root, expr);
        let is_var = matches!(expr, Expr::Column(id) if matches!(root.glob.node(*id), super::nodes::VarNode::Var(_)));
        if !is_var {
            let one = cost_qual_eval_node(expr);
            cost.startup += one.startup;
            cost.per_tuple += one.per_tuple;
        }
    }
    target.cost = cost;
    target.width = tuple_width;
}

/// get_expr_width returns the estimated width of an expression's values: a Var's column width, or its type's
/// average width, as Postgres' function of the same name does.
pub fn get_expr_width(root: &PlannerInfo<'_, '_>, expr: &Expr) -> f64 {
    if let Expr::Column(id) = expr
        && let super::nodes::VarNode::Var(var) = root.glob.node(*id)
        && let Some(width) = root.rels.get(var.varno).and_then(|rel| rel.attr_widths.get(var.varattno))
        && *width > 0.0
    {
        return *width;
    }
    get_typavgwidth(super::nodefuncs::expr_type(root, expr), -1)
}
