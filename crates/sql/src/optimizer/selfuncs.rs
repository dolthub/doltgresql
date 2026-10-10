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

//! The parts of Postgres' utils/adt/selfuncs.c that the planner calls for joins: the statistics of a join clause's
//! Vars, the selectivity of equality joins, and the share of a hash table's rows in one bucket. Restriction clauses
//! use the selectivity of `colstats`, which ports selfuncs.c's restriction estimates.

use std::sync::Arc;

use super::PlannerInfo;
use super::costsize::clamp_row_est;
use super::nodes::{IndexClause, IndexOptInfo, IndexPath, JoinType, Relids, RinfoId, SpecialJoinInfo, VarNode};
use crate::colstats::{ColumnStats, TableStats};
use crate::expr::Expr;
use crate::types::Value;

/// DEFAULT_NUM_DISTINCT is how many distinct values Postgres assumes a column holds without statistics.
const DEFAULT_NUM_DISTINCT: f64 = 200.0;

/// VariableStatData is what the planner knows of an expression's values, as Postgres' VariableStatData holds it:
/// the base relation whose Var it is, that Var's statistics, and whether a unique index makes its values distinct.
pub struct VariableStatData {
    pub rel: Option<usize>,
    stats: Option<(Arc<TableStats>, usize)>,
    isunique: bool,
    isbool: bool,
}

impl VariableStatData {
    /// column returns the statistics of the expression's Var.
    pub fn column(&self) -> Option<&ColumnStats> {
        self.stats.as_ref().and_then(|(stats, attno)| stats.columns.get(*attno))
    }
}

/// examine_variable finds what the planner knows of an expression's values, as Postgres' examine_variable does: the
/// statistics of a Var of a table, maybe under a cast, and nothing of any other expression.
pub fn examine_variable(root: &PlannerInfo<'_, '_>, e: &Expr) -> VariableStatData {
    let inner = match e {
        Expr::Cast(inner, ..) => inner,
        other => other,
    };
    let Expr::Column(id) = *inner else {
        return VariableStatData { rel: None, stats: None, isunique: false, isbool: false };
    };
    let VarNode::Var(var) = root.glob.node(id) else {
        return VariableStatData { rel: None, stats: None, isunique: false, isbool: false };
    };
    let (varno, attno) = (var.varno, var.varattno);
    let Some(table) = root.parse.rte(varno).table() else {
        return VariableStatData { rel: Some(varno), stats: None, isunique: false, isbool: false };
    };
    let isunique = (table.key_columns == [attno] && !table.keyless())
        || table.indexes.iter().any(|i| i.unique && i.predicate.is_empty() && i.columns == [attno]);
    let isbool = table.columns.get(attno).is_some_and(|c| c.ty.oid == crate::oid::BOOL);
    let stats = root.rels[varno].stats.clone().map(|stats| (stats, attno));
    VariableStatData { rel: Some(varno), stats, isunique, isbool }
}

/// get_variable_numdistinct returns about how many distinct values an expression takes, and whether that is only
/// the default, as Postgres' function of the same name does.
pub fn get_variable_numdistinct(root: &PlannerInfo<'_, '_>, vardata: &VariableStatData) -> (f64, bool) {
    let column = vardata.column();
    let stanullfrac = column.map_or(0.0, |c| c.null_frac);
    let mut stadistinct = match column {
        Some(c) => c.distinct,
        None if vardata.isbool => 2.0,
        None => 0.0,
    };
    if vardata.isunique {
        stadistinct = -(1.0 - stanullfrac);
    }
    if stadistinct > 0.0 {
        return (clamp_row_est(stadistinct), false);
    }
    let Some(rel) = vardata.rel else { return (DEFAULT_NUM_DISTINCT, true) };
    let ntuples = root.rels[rel].tuples;
    if ntuples <= 0.0 {
        return (DEFAULT_NUM_DISTINCT, true);
    }
    if stadistinct < 0.0 {
        return (clamp_row_est(-stadistinct * ntuples), false);
    }
    if ntuples < DEFAULT_NUM_DISTINCT {
        return (clamp_row_est(ntuples), true);
    }
    (DEFAULT_NUM_DISTINCT, true)
}

/// mcvs returns the most common values of a column, which the equality of two columns may compare when they are
/// values of the same kind.
fn mcvs(vardata: &VariableStatData) -> Option<&[(Value, f64)]> {
    vardata.column().map(|c| c.common.as_slice()).filter(|c| !c.is_empty())
}

/// values_equal reports whether two most common values are equal, comparing only values of the same kind.
fn values_equal(a: &Value, b: &Value) -> bool {
    std::mem::discriminant(a) == std::mem::discriminant(b) && crate::expr::compare_values(a, b).is_eq()
}

/// rowcomparesel returns the selectivity of a row comparison from its first pair of fields alone, compared as an
/// ordinary comparison, as Postgres' function of the same name estimates it.
pub fn rowcomparesel(
    root: &PlannerInfo<'_, '_>,
    op: crate::expr::CmpOp,
    (left, right): (&Expr, &Expr),
    varrelid: usize,
    jointype: super::nodes::JoinType,
    sjinfo: Option<&SpecialJoinInfo>,
) -> f64 {
    let opclause = Expr::Compare(op, Box::new(left.clone()), Box::new(right.clone()));
    super::clausesel::clause_selectivity(root, &opclause, None, varrelid, jointype, sjinfo)
}

/// eqjoinsel returns the selectivity of an equality join clause between two expressions, as Postgres' eqjoinsel
/// estimates it for the join it is part of.
pub fn eqjoinsel(root: &PlannerInfo<'_, '_>, left: &Expr, right: &Expr, sjinfo: &SpecialJoinInfo) -> f64 {
    let (vardata1, vardata2) = (examine_variable(root, left), examine_variable(root, right));
    let within = |v: &VariableStatData, relids: &Relids| v.rel.is_some_and(|r| relids.is_member(r));
    let join_is_reversed = within(&vardata1, &sjinfo.syn_righthand) || within(&vardata2, &sjinfo.syn_lefthand);
    let (nd1, isdefault1) = get_variable_numdistinct(root, &vardata1);
    let (nd2, isdefault2) = get_variable_numdistinct(root, &vardata2);
    let selec_inner = eqjoinsel_inner(&vardata1, &vardata2, nd1, nd2);
    let selec = match sjinfo.jointype {
        JoinType::Semi | JoinType::Anti => {
            let inner_rows = root.find_rel(&sjinfo.min_righthand).map_or(nd2, |r| root.rels[r].rows);
            let semi = match join_is_reversed {
                false => eqjoinsel_semi(root, &vardata1, &vardata2, nd1, nd2, isdefault1, isdefault2, inner_rows),
                true => eqjoinsel_semi(root, &vardata2, &vardata1, nd2, nd1, isdefault2, isdefault1, inner_rows),
            };
            semi.min(inner_rows * selec_inner)
        }
        _ => selec_inner,
    };
    selec.clamp(0.0, 1.0)
}

/// eqjoinsel_inner is eqjoinsel for an inner join: from the matches between the two sides' most common values when
/// both have them, and otherwise from their distinct counts.
fn eqjoinsel_inner(vardata1: &VariableStatData, vardata2: &VariableStatData, nd1: f64, nd2: f64) -> f64 {
    let (nullfrac1, nullfrac2) =
        (vardata1.column().map_or(0.0, |c| c.null_frac), vardata2.column().map_or(0.0, |c| c.null_frac));
    let (Some(mcv1), Some(mcv2)) = (mcvs(vardata1), mcvs(vardata2)) else {
        return (1.0 - nullfrac1) * (1.0 - nullfrac2) / nd1.max(nd2);
    };
    let (mut hasmatch1, mut hasmatch2) = (vec![false; mcv1.len()], vec![false; mcv2.len()]);
    let (mut matchprodfreq, mut nmatches) = (0.0, 0.0);
    for (i, (value1, freq1)) in mcv1.iter().enumerate() {
        for (j, (value2, freq2)) in mcv2.iter().enumerate() {
            if !hasmatch2[j] && values_equal(value1, value2) {
                (hasmatch1[i], hasmatch2[j]) = (true, true);
                matchprodfreq += freq1 * freq2;
                nmatches += 1.0;
                break;
            }
        }
    }
    let matchprodfreq = matchprodfreq.clamp(0.0, 1.0);
    let split = |mcv: &[(Value, f64)], hasmatch: &[bool]| {
        let matched: f64 = mcv.iter().zip(hasmatch).filter(|(_, m)| **m).map(|((_, f), _)| f).sum();
        let unmatched: f64 = mcv.iter().zip(hasmatch).filter(|(_, m)| !**m).map(|((_, f), _)| f).sum();
        (matched.clamp(0.0, 1.0), unmatched.clamp(0.0, 1.0))
    };
    let (matchfreq1, unmatchfreq1) = split(mcv1, &hasmatch1);
    let (matchfreq2, unmatchfreq2) = split(mcv2, &hasmatch2);
    let otherfreq1 = (1.0 - nullfrac1 - matchfreq1 - unmatchfreq1).clamp(0.0, 1.0);
    let otherfreq2 = (1.0 - nullfrac2 - matchfreq2 - unmatchfreq2).clamp(0.0, 1.0);
    let (nvalues1, nvalues2) = (mcv1.len() as f64, mcv2.len() as f64);
    let mut totalsel1 = matchprodfreq;
    if nd2 > nvalues2 {
        totalsel1 += unmatchfreq1 * otherfreq2 / (nd2 - nvalues2);
    }
    if nd2 > nmatches {
        totalsel1 += otherfreq1 * (otherfreq2 + unmatchfreq2) / (nd2 - nmatches);
    }
    let mut totalsel2 = matchprodfreq;
    if nd1 > nvalues1 {
        totalsel2 += unmatchfreq2 * otherfreq1 / (nd1 - nvalues1);
    }
    if nd1 > nmatches {
        totalsel2 += otherfreq2 * (otherfreq1 + unmatchfreq1) / (nd1 - nmatches);
    }
    totalsel1.min(totalsel2)
}

/// eqjoinsel_semi is eqjoinsel for a semi or anti join: the share of outer rows that find a match.
#[allow(clippy::too_many_arguments)]
fn eqjoinsel_semi(
    root: &PlannerInfo<'_, '_>,
    vardata1: &VariableStatData,
    vardata2: &VariableStatData,
    mut nd1: f64,
    mut nd2: f64,
    isdefault1: bool,
    mut isdefault2: bool,
    inner_rows: f64,
) -> f64 {
    if let Some(rel) = vardata2.rel
        && nd2 >= root.rels[rel].rows
    {
        nd2 = root.rels[rel].rows;
        isdefault2 = false;
    }
    if nd2 >= inner_rows {
        nd2 = inner_rows;
        isdefault2 = false;
    }
    let nullfrac1 = vardata1.column().map_or(0.0, |c| c.null_frac);
    let (Some(mcv1), Some(mcv2)) = (mcvs(vardata1), mcvs(vardata2)) else {
        return match !isdefault1 && !isdefault2 {
            true if nd1 <= nd2 || nd2 < 0.0 => 1.0 - nullfrac1,
            true => (nd2 / nd1) * (1.0 - nullfrac1),
            false => 0.5 * (1.0 - nullfrac1),
        };
    };
    let clamped_nvalues2 = (mcv2.len() as f64).min(nd2) as usize;
    let mut hasmatch2 = vec![false; clamped_nvalues2];
    let (mut matchfreq1, mut nmatches) = (0.0, 0.0);
    for (value1, freq1) in mcv1 {
        for (j, (value2, _)) in mcv2[..clamped_nvalues2].iter().enumerate() {
            if !hasmatch2[j] && values_equal(value1, value2) {
                hasmatch2[j] = true;
                matchfreq1 += freq1;
                nmatches += 1.0;
                break;
            }
        }
    }
    let matchfreq1 = f64::clamp(matchfreq1, 0.0, 1.0);
    let uncertainfrac = match !isdefault1 && !isdefault2 {
        true => {
            nd1 -= nmatches;
            nd2 -= nmatches;
            if nd1 <= nd2 || nd2 < 0.0 { 1.0 } else { nd2 / nd1 }
        }
        false => 0.5,
    };
    let uncertain = (1.0 - matchfreq1 - nullfrac1).clamp(0.0, 1.0);
    matchfreq1 + uncertainfrac * uncertain
}

/// estimate_hash_bucket_stats returns the frequency of the most common value of a hash key and the share of the
/// hashed rows that one bucket of a hash table of a number of buckets holds, as Postgres' function of the same name
/// estimates them.
pub fn estimate_hash_bucket_stats(root: &PlannerInfo<'_, '_>, hashkey: &Expr, nbuckets: f64) -> (f64, f64) {
    let vardata = examine_variable(root, hashkey);
    let mcv_freq = vardata.column().and_then(|c| c.common.first()).map_or(0.0, |(_, f)| *f);
    let (mut ndistinct, isdefault) = get_variable_numdistinct(root, &vardata);
    if isdefault {
        return (mcv_freq, 0.1);
    }
    let stanullfrac = vardata.column().map_or(0.0, |c| c.null_frac);
    let avgfreq = (1.0 - stanullfrac) / ndistinct;
    if let Some(rel) = vardata.rel
        && root.rels[rel].tuples > 0.0
    {
        ndistinct = clamp_row_est(ndistinct * root.rels[rel].rows / root.rels[rel].tuples);
    }
    let mut estfract = if ndistinct > nbuckets { 1.0 / nbuckets } else { 1.0 / ndistinct };
    if avgfreq > 0.0 && mcv_freq > avgfreq {
        estfract *= mcv_freq / avgfreq;
    }
    (mcv_freq, estfract.clamp(1.0e-6, 1.0))
}

/// DEFAULT_PAGE_CPU_MULTIPLIER is how many operators' cost Postgres charges for processing one index page in a
/// btree descent.
const DEFAULT_PAGE_CPU_MULTIPLIER: f64 = 50.0;

/// IndexCostEstimate is what an index access method's cost estimate returns, as Postgres' amcostestimate does: the
/// startup and total costs of reading the index, the share of its entries it reads, and how closely its order follows
/// the table's.
pub struct IndexCostEstimate {
    pub startup_cost: f64,
    pub total_cost: f64,
    pub selectivity: f64,
    pub correlation: f64,
}

/// get_quals_from_indexclauses returns the clauses that an index path's index clauses search by, as Postgres'
/// function of the same name does.
fn get_quals_from_indexclauses(indexclauses: &[IndexClause]) -> Vec<RinfoId> {
    indexclauses.iter().flat_map(|iclause| iclause.indexquals.iter().copied()).collect()
}

/// index_other_operands_eval_cost returns the cost of evaluating the sides of index clauses that are not the index
/// column, once per scan, as Postgres' function of the same name does.
fn index_other_operands_eval_cost(root: &PlannerInfo<'_, '_>, indexquals: &[RinfoId]) -> f64 {
    let mut qual_arg_cost = 0.0;
    for &q in indexquals {
        let other_operand = match &root.rinfos[q].clause {
            Expr::Compare(_, _, other) => Some(&**other),
            Expr::AnyArray(_, array, _) => Some(&**array),
            _ => None,
        };
        if let Some(other) = other_operand {
            let cost = super::costsize::cost_qual_eval_node(other);
            qual_arg_cost += cost.startup + cost.per_tuple;
        }
    }
    qual_arg_cost
}

/// add_predicate_to_index_quals returns an index's clauses with its predicate's conjuncts that they do not imply, as
/// Postgres' function of the same name does, where a clause implies a conjunct equal to it.
fn add_predicate_to_index_quals<'q>(
    root: &'q PlannerInfo<'_, '_>,
    index: &'q IndexOptInfo,
    indexquals: &[RinfoId],
) -> Vec<(&'q Expr, Option<&'q super::nodes::RestrictInfo>)> {
    let mut quals: Vec<(&Expr, Option<&super::nodes::RestrictInfo>)> = index
        .indpred
        .iter()
        .filter(|pred| !indexquals.iter().any(|&q| root.rinfos[q].clause == **pred))
        .map(|pred| (pred, None))
        .collect();
    quals.extend(indexquals.iter().map(|&q| (&root.rinfos[q].clause, Some(&root.rinfos[q]))));
    quals
}

/// estimate_array_length returns how many elements an array expression has, or 10 when that is unknown, as
/// Postgres' function of the same name estimates it.
pub fn estimate_array_length(arrayexpr: &Expr) -> f64 {
    match arrayexpr {
        Expr::Const(Value::Array(array)) => array.values.len() as f64,
        Expr::Array(_, items, false) => items.len() as f64,
        _ => 10.0,
    }
}

/// genericcostestimate returns the costs of reading an index that every access method shares, given the entries
/// and scans that btcostestimate found, as Postgres' function of the same name does.
fn genericcostestimate(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    path: &IndexPath,
    loop_count: f64,
    num_index_tuples: f64,
    num_sa_scans: f64,
) -> IndexCostEstimate {
    let index = &root.rels[rel].indexlist[path.index];
    let index_quals = get_quals_from_indexclauses(&path.indexclauses);
    let selectivity_quals = add_predicate_to_index_quals(root, index, &index_quals);
    let index_selectivity = super::clausesel::list_selectivity(root, &selectivity_quals, rel, JoinType::Inner, None);
    let tuples = root.rels[rel].tuples;
    let mut num_index_tuples = match num_index_tuples <= 0.0 {
        true => (index_selectivity * tuples / num_sa_scans).round(),
        false => num_index_tuples,
    };
    num_index_tuples = num_index_tuples.min(index.tuples).max(1.0);
    let num_index_pages = match index.pages > 1.0 && index.tuples > 1.0 {
        true => (num_index_tuples * index.pages / index.tuples).ceil(),
        false => 1.0,
    };
    let num_scans = num_sa_scans * loop_count;
    let mut index_total_cost = match num_scans > 1.0 {
        true => {
            let pages_fetched =
                super::costsize::index_pages_fetched(root, num_index_pages * num_scans, index.pages, index.pages);
            pages_fetched * super::costsize::RANDOM_PAGE_COST / loop_count
        }
        false => num_index_pages * super::costsize::RANDOM_PAGE_COST,
    };
    let qual_arg_cost = index_other_operands_eval_cost(root, &index_quals);
    let qual_op_cost = super::costsize::CPU_OPERATOR_COST * index_quals.len() as f64;
    index_total_cost += qual_arg_cost;
    index_total_cost += num_index_tuples * num_sa_scans * (super::costsize::CPU_INDEX_TUPLE_COST + qual_op_cost);
    IndexCostEstimate {
        startup_cost: qual_arg_cost,
        total_cost: index_total_cost,
        selectivity: index_selectivity,
        correlation: 0.0,
    }
}

/// examine_indexcol_variable finds what the planner knows of the values of an index's column, as Postgres' function
/// of the same name does.
fn examine_indexcol_variable(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    index: &IndexOptInfo,
    indexcol: usize,
) -> VariableStatData {
    match index.indexkeys[indexcol] {
        Some(attno) => {
            let table = root.parse.rte(rel).table();
            let isunique = table.is_some_and(|table| {
                (table.key_columns == [attno] && !table.keyless())
                    || table.indexes.iter().any(|i| i.unique && i.predicate.is_empty() && i.columns == [attno])
            });
            let isbool = table.and_then(|t| t.columns.get(attno)).is_some_and(|c| c.ty.oid == crate::oid::BOOL);
            let stats = root.rels[rel].stats.clone().map(|stats| (stats, attno));
            VariableStatData { rel: Some(rel), stats, isunique, isbool }
        }
        None => VariableStatData { rel: Some(rel), stats: None, isunique: false, isbool: false },
    }
}

/// btcost_correlation returns how closely an index's order follows its table's, from the correlation of its first
/// column, as Postgres' function of the same name does.
fn btcost_correlation(index: &IndexOptInfo, vardata: &VariableStatData) -> f64 {
    let Some(column) = vardata.column() else { return 0.0 };
    let var_correlation = if index.reverse_sort[0] { -column.correlation } else { column.correlation };
    if index.nkeycolumns > 1 { var_correlation * 0.75 } else { var_correlation }
}

/// btcostestimate returns the costs of reading a btree index for an index path, counting the entries that its
/// bounding clauses read and the scans that its arrays and skipped columns start, as Postgres' function of the same
/// name does.
pub fn btcostestimate(root: &PlannerInfo<'_, '_>, rel: usize, path: &IndexPath, loop_count: f64) -> IndexCostEstimate {
    let index = &root.rels[rel].indexlist[path.index];
    let mut index_bound_quals: Vec<RinfoId> = Vec::new();
    let mut index_skip_quals: Vec<RinfoId> = Vec::new();
    let mut indexcol = 0;
    let (mut eq_qual_here, mut found_row_compare, mut found_array, mut found_is_null_op) = (false, false, false, false);
    let mut have_correlation = false;
    let mut num_sa_scans: f64 = 1.0;
    let mut correlation = 0.0;
    'clauses: for iclause in &path.indexclauses {
        if indexcol < iclause.indexcol {
            let num_sa_scans_prev_cols = num_sa_scans;
            if found_row_compare {
                break;
            }
            if eq_qual_here {
                indexcol += 1;
                index_skip_quals.clear();
            }
            eq_qual_here = false;
            while indexcol < iclause.indexcol {
                found_array = true;
                let vardata = examine_indexcol_variable(root, rel, index, indexcol);
                let (mut ndistinct, isdefault) = get_variable_numdistinct(root, &vardata);
                if indexcol == 0 {
                    if vardata.column().is_some() {
                        correlation = btcost_correlation(index, &vardata);
                    }
                    have_correlation = true;
                }
                if isdefault {
                    num_sa_scans = num_sa_scans_prev_cols;
                    break;
                }
                if !index_skip_quals.is_empty() {
                    let partial_skip_quals = add_predicate_to_index_quals(root, index, &index_skip_quals);
                    let ndistinctfrac =
                        super::clausesel::list_selectivity(root, &partial_skip_quals, rel, JoinType::Inner, None);
                    if ndistinctfrac < super::clausesel::DEFAULT_RANGE_INEQ_SEL {
                        num_sa_scans = num_sa_scans_prev_cols;
                        break;
                    }
                    ndistinct = (ndistinct * ndistinctfrac).round().max(1.0);
                }
                if index_skip_quals.is_empty() {
                    ndistinct += 1.0;
                }
                num_sa_scans *= ndistinct;
                if index.pages < num_sa_scans {
                    num_sa_scans = num_sa_scans_prev_cols;
                    break;
                }
                indexcol += 1;
                index_skip_quals.clear();
            }
            if indexcol != iclause.indexcol {
                break 'clauses;
            }
        }
        for &q in &iclause.indexquals {
            match &root.rinfos[q].clause {
                Expr::RowCompare(..) => found_row_compare = true,
                Expr::Compare(crate::expr::CmpOp::Eq, ..) => eq_qual_here = true,
                Expr::AnyArray(_, array, _) => {
                    let alength = estimate_array_length(array);
                    found_array = true;
                    if alength > 1.0 {
                        num_sa_scans *= alength;
                    }
                    eq_qual_here = true;
                }
                Expr::IsNull(_, false) => {
                    found_is_null_op = true;
                    eq_qual_here = true;
                }
                _ => {}
            }
            index_bound_quals.push(q);
            if !eq_qual_here && !found_row_compare && indexcol + 1 < index.nkeycolumns {
                index_skip_quals.push(q);
            }
        }
    }
    let num_index_tuples =
        if index.unique && indexcol + 1 == index.nkeycolumns && eq_qual_here && !found_array && !found_is_null_op {
            1.0
        } else {
            let selectivity_quals = add_predicate_to_index_quals(root, index, &index_bound_quals);
            let btree_selectivity =
                super::clausesel::list_selectivity(root, &selectivity_quals, rel, JoinType::Inner, None);
            let num_index_tuples = btree_selectivity * root.rels[rel].tuples;
            num_sa_scans = num_sa_scans.min((index.pages * 0.3333333).ceil()).max(1.0);
            (num_index_tuples / num_sa_scans).round()
        };
    let mut costs = genericcostestimate(root, rel, path, loop_count, num_index_tuples, num_sa_scans);
    if index.tuples > 1.0 {
        let descent_cost = index.tuples.log2().ceil() * super::costsize::CPU_OPERATOR_COST;
        costs.startup_cost += descent_cost;
        costs.total_cost += num_sa_scans * descent_cost;
    }
    let descent_cost = (index.tree_height + 1.0) * DEFAULT_PAGE_CPU_MULTIPLIER * super::costsize::CPU_OPERATOR_COST;
    costs.startup_cost += descent_cost;
    costs.total_cost += num_sa_scans * descent_cost;
    costs.correlation = match have_correlation {
        true => correlation,
        false => {
            let vardata = examine_indexcol_variable(root, rel, index, 0);
            btcost_correlation(index, &vardata)
        }
    };
    costs
}

/// GroupVarInfo is a Var of a grouping with its relation and estimated distinct values, as Postgres' GroupVarInfo
/// holds it.
struct GroupVarInfo {
    var: Expr,
    rel: Option<usize>,
    ndistinct: f64,
}

/// estimate_num_groups estimates the groups that grouping rows by expressions makes, or of a grouping set's
/// positions among them, as Postgres' function of the same name does: a boolean has two values, an expression with
/// statistics or a unique Var has its distinct values, and any other expression the distinct values of its Vars,
/// multiplied over the relations, each relation's product clamped to its rows.
pub fn estimate_num_groups(
    root: &PlannerInfo<'_, '_>,
    group_exprs: &[Expr],
    input_rows: f64,
    pgset: Option<&[usize]>,
) -> f64 {
    let input_rows = clamp_row_est(input_rows);
    if group_exprs.is_empty() || pgset.is_some_and(|s| s.is_empty()) {
        return 1.0;
    }
    let mut varinfos: Vec<GroupVarInfo> = Vec::new();
    let mut srf_multiplier: f64 = 1.0;
    let mut numdistinct = 1.0;
    for (i, groupexpr) in group_exprs.iter().enumerate() {
        if pgset.is_some_and(|s| !s.contains(&i)) {
            continue;
        }
        srf_multiplier = srf_multiplier.max(super::clauses::expression_returns_set_rows(root, groupexpr));
        if super::nodefuncs::expr_type(root, groupexpr) == Some(super::nodefuncs::BOOLOID) {
            numdistinct *= 2.0;
            continue;
        }
        let vardata = examine_variable(root, groupexpr);
        if vardata.stats.is_some() || vardata.isunique {
            add_unique_group_var(root, &mut varinfos, groupexpr.clone(), &vardata);
            continue;
        }
        let varshere = super::var::pull_var_clause(root.glob, groupexpr, false);
        if varshere.is_empty() {
            if super::clauses::contain_volatile_functions(root.glob, groupexpr) {
                return input_rows;
            }
            continue;
        }
        for var in varshere {
            let var = Expr::Column(var);
            let vardata = examine_variable(root, &var);
            add_unique_group_var(root, &mut varinfos, var, &vardata);
        }
    }
    if varinfos.is_empty() {
        return (numdistinct * srf_multiplier).ceil().min(input_rows).max(1.0);
    }
    while !varinfos.is_empty() {
        let rel = varinfos[0].rel;
        let (relvarinfos, newvarinfos): (Vec<GroupVarInfo>, Vec<GroupVarInfo>) =
            varinfos.into_iter().partition(|v| v.rel == rel);
        let mut reldistinct = 1.0;
        let mut relmaxndistinct = reldistinct;
        let relvarcount = relvarinfos.len();
        for varinfo in &relvarinfos {
            reldistinct *= varinfo.ndistinct;
            relmaxndistinct = f64::max(relmaxndistinct, varinfo.ndistinct);
        }
        if let Some(rel) = rel.map(|r| &root.rels[r])
            && rel.tuples > 0.0
        {
            let mut clamp = rel.tuples;
            if relvarcount > 1 {
                clamp *= 0.1;
                if clamp < relmaxndistinct {
                    clamp = relmaxndistinct.min(rel.tuples);
                }
            }
            reldistinct = reldistinct.min(clamp);
            if reldistinct > 0.0 && rel.rows < rel.tuples {
                reldistinct *= 1.0 - ((rel.tuples - rel.rows) / rel.tuples).powf(rel.tuples / reldistinct);
            }
            numdistinct *= clamp_row_est(reldistinct);
        }
        varinfos = newvarinfos;
    }
    (numdistinct * srf_multiplier).ceil().min(input_rows).max(1.0)
}

/// add_unique_group_var adds a grouping's Var to its Var infos unless it is there already, or a Var of another
/// relation that it is known equal to has fewer distinct values, as Postgres' function of the same name does.
fn add_unique_group_var(
    root: &PlannerInfo<'_, '_>,
    varinfos: &mut Vec<GroupVarInfo>,
    var: Expr,
    vardata: &VariableStatData,
) {
    let (ndistinct, _) = get_variable_numdistinct(root, vardata);
    let base = |e: &Expr| match e {
        Expr::Column(id) => match root.glob.node(*id) {
            VarNode::Var(v) => Some((v.varno, v.varattno)),
            VarNode::PlaceHolderVar(_) => None,
        },
        _ => None,
    };
    let mut i = 0;
    while i < varinfos.len() {
        let same = varinfos[i].var == var || (base(&var).is_some() && base(&var) == base(&varinfos[i].var));
        if same {
            return;
        }
        if vardata.rel != varinfos[i].rel && super::equivclass::exprs_known_equal(root, &var, &varinfos[i].var) {
            if varinfos[i].ndistinct <= ndistinct {
                return;
            }
            varinfos.remove(i);
            continue;
        }
        i += 1;
    }
    varinfos.push(GroupVarInfo { var, rel: vardata.rel, ndistinct });
}
