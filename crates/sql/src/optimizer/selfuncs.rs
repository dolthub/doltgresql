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
use super::nodes::{JoinType, SpecialJoinInfo, is_subset, singleton, var_parts};
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
pub fn examine_variable(root: &mut PlannerInfo<'_, '_>, e: &Expr) -> VariableStatData {
    let inner = match e {
        Expr::Cast(inner, ..) => inner,
        other => other,
    };
    let Expr::Column(var) = *inner else {
        return VariableStatData { rel: None, stats: None, isunique: false, isbool: false };
    };
    let (varno, attno) = var_parts(var);
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

/// eqjoinsel returns the selectivity of an equality join clause between two expressions, as Postgres' eqjoinsel
/// estimates it for the join it is part of.
pub fn eqjoinsel(root: &mut PlannerInfo<'_, '_>, left: &Expr, right: &Expr, sjinfo: &SpecialJoinInfo) -> f64 {
    let (vardata1, vardata2) = (examine_variable(root, left), examine_variable(root, right));
    let within = |v: &VariableStatData, relids| v.rel.is_some_and(|r| is_subset(singleton(r), relids));
    let join_is_reversed = within(&vardata1, sjinfo.syn_righthand) || within(&vardata2, sjinfo.syn_lefthand);
    let (nd1, isdefault1) = get_variable_numdistinct(root, &vardata1);
    let (nd2, isdefault2) = get_variable_numdistinct(root, &vardata2);
    let selec_inner = eqjoinsel_inner(&vardata1, &vardata2, nd1, nd2);
    let selec = match sjinfo.jointype {
        JoinType::Semi | JoinType::Anti => {
            let inner_rows = root.find_rel(sjinfo.min_righthand).map_or(nd2, |r| root.rels[r].rows);
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
pub fn estimate_hash_bucket_stats(root: &mut PlannerInfo<'_, '_>, hashkey: &Expr, nbuckets: f64) -> (f64, f64) {
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
