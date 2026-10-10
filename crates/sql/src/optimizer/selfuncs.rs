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

//! Postgres' utils/adt/selfuncs.c: the selectivity of restriction and join clauses from the statistics of their
//! Vars, the number of distinct groups of expressions, the share of a hash table's rows in one bucket, and the cost of
//! an index scan. Doltgres' statistics hold what pg_statistic does: each column's NULL share, distinct count, most
//! common values with their shares, and histogram bounds, in the order of the column's type, where the C collation
//! orders text.

use std::cmp::Ordering;
use std::sync::Arc;

use super::PlannerInfo;
use super::clausesel::DEFAULT_INEQ_SEL;
use super::costsize::clamp_row_est;
use super::nodes::{IndexClause, IndexOptInfo, IndexPath, JoinType, Relids, RinfoId, SpecialJoinInfo, VarNode};
use crate::colstats::{ColumnStats, TableStats};
use crate::expr::{CmpOp, Expr};
use crate::types::Value;

/// DEFAULT_NUM_DISTINCT is how many distinct values Postgres assumes a column holds without statistics.
const DEFAULT_NUM_DISTINCT: f64 = 200.0;

/// The selectivities that Postgres assumes without statistics: of an equality, IS NULL and IS NOT NULL, and of a
/// boolean function.
const DEFAULT_EQ_SEL: f64 = 0.005;
const DEFAULT_UNK_SEL: f64 = 0.005;
const DEFAULT_NOT_UNK_SEL: f64 = 1.0 - DEFAULT_UNK_SEL;
const DEFAULT_FUNCTION_SEL: f64 = 0.3333333;

/// VariableStatData is what the planner knows of an expression's values, as Postgres' VariableStatData holds it: the
/// expression without its PlaceHolderVars, the base or join relation whose Vars it reads, the statistics of its Var,
/// and whether a unique index makes its values distinct.
pub struct VariableStatData {
    pub var: Expr,
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

/// examine_variable finds what the planner knows of an expression's values, as Postgres' function of the same name
/// does: the statistics of a Var of a table, under casts, and the relation of an expression of one relation's Vars,
/// or of a join's when `varrelid` is zero, where only the Vars of the relation that `varrelid` names are variables
/// when it is not. Doltgres gathers no statistics of index expressions or extended statistics.
pub fn examine_variable(root: &PlannerInfo<'_, '_>, node: &Expr, varrelid: usize) -> VariableStatData {
    let mut basenode = strip_all_phvs_deep(root, node);
    while let Expr::Cast(inner, ..) = basenode {
        basenode = *inner;
    }
    let mut vardata = VariableStatData { var: node.clone(), rel: None, stats: None, isunique: false, isbool: false };
    if let Expr::Column(id) = basenode
        && let VarNode::Var(var) = root.glob.node(id)
        && (varrelid == 0 || varrelid == var.varno)
        && var.varno != 0
    {
        vardata.var = basenode;
        vardata.rel = Some(var.varno);
        examine_simple_variable(root, var.varno, var.varattno, &mut vardata);
        return vardata;
    }
    let varnos = super::var::pull_varnos(root, &basenode);
    let basevarnos = varnos.difference(&root.outer_join_rels);
    if !basevarnos.is_empty() {
        match basevarnos.singleton_member() {
            Some(relid) if varrelid == 0 || varrelid == relid => {
                vardata.rel = Some(relid);
                vardata.var = basenode;
            }
            Some(_) => {}
            None if varrelid == 0 => {
                vardata.rel = root.find_rel(&varnos);
                vardata.var = basenode;
            }
            None if varnos.is_member(varrelid) => {
                vardata.rel = Some(varrelid);
                vardata.var = basenode;
            }
            None => {}
        }
    }
    vardata
}

/// examine_simple_variable finds the statistics of a Var of a base relation, and whether a unique index makes its
/// values distinct, as Postgres' function of the same name does for a table. Doltgres keeps no statistics of a
/// subquery's output.
fn examine_simple_variable(root: &PlannerInfo<'_, '_>, varno: usize, attno: usize, vardata: &mut VariableStatData) {
    let Some(table) = root.parse.rte(varno).table() else { return };
    vardata.isunique = (table.key_columns == [attno] && !table.keyless())
        || table.indexes.iter().any(|i| i.unique && i.predicate.is_empty() && i.columns == [attno]);
    vardata.isbool = table.columns.get(attno).is_some_and(|c| c.ty.oid == crate::oid::BOOL);
    vardata.stats = root.rels[varno].stats.clone().map(|stats| (stats, attno));
}

/// strip_all_phvs_deep returns an expression with each PlaceHolderVar replaced by its expression, as Postgres'
/// function of the same name does.
fn strip_all_phvs_deep(root: &PlannerInfo<'_, '_>, e: &Expr) -> Expr {
    match e {
        Expr::Column(id) => match root.glob.node(*id) {
            VarNode::PlaceHolderVar(phv) => strip_all_phvs_deep(root, &root.glob.placeholder(phv.phid).phexpr),
            VarNode::Var(_) => e.clone(),
        },
        other => other.clone().map_children(&mut |c| strip_all_phvs_deep(root, &c)),
    }
}

/// estimate_expression_value returns an expression with the parts it can evaluate now folded into constants, as
/// Postgres' function of the same name does, which the binder did already for the expressions that read no Var.
pub fn estimate_expression_value(e: &Expr) -> Expr {
    e.clone()
}

/// get_restriction_variable returns the statistics of the variable side of a binary clause and its other side,
/// with whether the variable is on the left, when one side reads the relation's Vars and the other reads none, as
/// Postgres' function of the same name does.
pub fn get_restriction_variable(
    root: &PlannerInfo<'_, '_>,
    left: &Expr,
    right: &Expr,
    varrelid: usize,
) -> Option<(VariableStatData, Expr, bool)> {
    let vardata = examine_variable(root, left, varrelid);
    let rdata = examine_variable(root, right, varrelid);
    match (vardata.rel, rdata.rel) {
        (Some(_), None) => Some((vardata, estimate_expression_value(&rdata.var), true)),
        (None, Some(_)) => Some((rdata, estimate_expression_value(&vardata.var), false)),
        _ => None,
    }
}

/// get_join_variables returns the statistics of the two sides of a join clause, with whether they are reversed
/// from the join's own sides, as Postgres' function of the same name does.
pub fn get_join_variables(
    root: &PlannerInfo<'_, '_>,
    left: &Expr,
    right: &Expr,
    sjinfo: &SpecialJoinInfo,
) -> (VariableStatData, VariableStatData, bool) {
    let (vardata1, vardata2) = (examine_variable(root, left, 0), examine_variable(root, right, 0));
    let within = |v: &VariableStatData, relids: &Relids| v.rel.is_some_and(|r| root.rels[r].relids.is_subset(relids));
    let reversed = within(&vardata1, &sjinfo.syn_righthand) && within(&vardata2, &sjinfo.syn_lefthand);
    (vardata1, vardata2, reversed)
}

/// restriction_selectivity returns the selectivity of a comparison of two expressions as a restriction of a
/// relation, by the comparison's restriction estimator, as Postgres' function of the same name calls its operator's
/// oprrest: eqsel, neqsel, or scalarltsel and its kin.
pub fn restriction_selectivity(
    root: &PlannerInfo<'_, '_>,
    op: CmpOp,
    left: &Expr,
    right: &Expr,
    varrelid: usize,
) -> f64 {
    match op {
        CmpOp::Eq => eqsel_internal(root, left, right, varrelid, false),
        CmpOp::Ne => eqsel_internal(root, left, right, varrelid, true),
        CmpOp::Lt => scalarineqsel_wrapper(root, left, right, varrelid, false, false),
        CmpOp::Le => scalarineqsel_wrapper(root, left, right, varrelid, false, true),
        CmpOp::Gt => scalarineqsel_wrapper(root, left, right, varrelid, true, false),
        CmpOp::Ge => scalarineqsel_wrapper(root, left, right, varrelid, true, true),
    }
}

/// join_selectivity returns the selectivity of a comparison of two expressions as a join clause, by the
/// comparison's join estimator, as Postgres' function of the same name calls its operator's oprjoin: eqjoinsel,
/// neqjoinsel, or scalarltjoinsel and its kin.
pub fn join_selectivity(
    root: &PlannerInfo<'_, '_>,
    op: CmpOp,
    left: &Expr,
    right: &Expr,
    jointype: JoinType,
    sjinfo: &SpecialJoinInfo,
) -> f64 {
    match op {
        CmpOp::Eq => eqjoinsel(root, left, right, sjinfo),
        CmpOp::Ne => neqjoinsel(root, left, right, jointype, sjinfo),
        _ => DEFAULT_INEQ_SEL,
    }
}

/// function_selectivity returns the selectivity of a boolean function call, as Postgres' function of the same name
/// does for a function without a support function.
pub fn function_selectivity() -> f64 {
    DEFAULT_FUNCTION_SEL
}

/// eqsel_internal returns the selectivity of an equality, or of an inequality when `negate` asks, as Postgres'
/// function of the same name does.
fn eqsel_internal(root: &PlannerInfo<'_, '_>, left: &Expr, right: &Expr, varrelid: usize, negate: bool) -> f64 {
    let Some((vardata, other, varonleft)) = get_restriction_variable(root, left, right, varrelid) else {
        return if negate { 1.0 - DEFAULT_EQ_SEL } else { DEFAULT_EQ_SEL };
    };
    match &other {
        Expr::Const(value) => var_eq_const(root, &vardata, value, varonleft, negate),
        _ => var_eq_non_const(root, &vardata, negate),
    }
}

/// var_eq_const returns the selectivity of an equality of a variable and a constant, or of their inequality, as
/// Postgres' function of the same name does: the constant's share when it is a most common value, and otherwise an
/// even share of what the most common values leave over the other distinct values.
pub fn var_eq_const(
    root: &PlannerInfo<'_, '_>,
    vardata: &VariableStatData,
    constval: &Value,
    _varonleft: bool,
    negate: bool,
) -> f64 {
    if constval.is_null() {
        return 0.0;
    }
    let nullfrac = vardata.column().map_or(0.0, |c| c.null_frac);
    let tuples = vardata.rel.map_or(0.0, |r| root.rels[r].tuples);
    let mut selec = if vardata.isunique && tuples >= 1.0 {
        1.0 / tuples
    } else if let Some(column) = vardata.column() {
        match column.common.iter().find(|(v, _)| compare_values(v, constval) == Some(Ordering::Equal)) {
            Some((_, freq)) => *freq,
            None => {
                let sumcommon: f64 = column.common.iter().map(|(_, f)| f).sum();
                let mut selec = (1.0 - sumcommon - nullfrac).clamp(0.0, 1.0);
                let otherdistinct = get_variable_numdistinct(root, vardata).0 - column.common.len() as f64;
                if otherdistinct > 1.0 {
                    selec /= otherdistinct;
                }
                if let Some((_, least)) = column.common.last()
                    && selec > *least
                {
                    selec = *least;
                }
                selec
            }
        }
    } else {
        1.0 / get_variable_numdistinct(root, vardata).0
    };
    if negate {
        selec = 1.0 - selec - nullfrac;
    }
    selec.clamp(0.0, 1.0)
}

/// var_eq_non_const returns the selectivity of an equality of a variable and an expression whose value is unknown,
/// or of their inequality, as Postgres' function of the same name does.
fn var_eq_non_const(root: &PlannerInfo<'_, '_>, vardata: &VariableStatData, negate: bool) -> f64 {
    let nullfrac = vardata.column().map_or(0.0, |c| c.null_frac);
    let tuples = vardata.rel.map_or(0.0, |r| root.rels[r].tuples);
    let mut selec = if vardata.isunique && tuples >= 1.0 {
        1.0 / tuples
    } else if let Some(column) = vardata.column() {
        let mut selec = 1.0 - nullfrac;
        let ndistinct = get_variable_numdistinct(root, vardata).0;
        if ndistinct > 1.0 {
            selec /= ndistinct;
        }
        if let Some((_, most)) = column.common.first()
            && selec > *most
        {
            selec = *most;
        }
        selec
    } else {
        1.0 / get_variable_numdistinct(root, vardata).0
    };
    if negate {
        selec = 1.0 - selec - nullfrac;
    }
    selec.clamp(0.0, 1.0)
}

/// scalarineqsel_wrapper returns the selectivity of an inequality, putting the variable on the left, as Postgres'
/// function of the same name does.
fn scalarineqsel_wrapper(
    root: &PlannerInfo<'_, '_>,
    left: &Expr,
    right: &Expr,
    varrelid: usize,
    isgt: bool,
    iseq: bool,
) -> f64 {
    let Some((vardata, other, varonleft)) = get_restriction_variable(root, left, right, varrelid) else {
        return DEFAULT_INEQ_SEL;
    };
    let Expr::Const(constval) = &other else { return DEFAULT_INEQ_SEL };
    if constval.is_null() {
        return 0.0;
    }
    let isgt = if varonleft { isgt } else { !isgt };
    scalarineqsel(root, isgt, iseq, &vardata, constval)
}

/// scalarineqsel returns the share of rows whose variable compares with a constant as `var < const`, `var <= const`,
/// `var > const`, or `var >= const` asks, as Postgres' function of the same name does: the shares of the most common
/// values that pass, plus the histogram's share of what they leave.
pub fn scalarineqsel(
    root: &PlannerInfo<'_, '_>,
    isgt: bool,
    iseq: bool,
    vardata: &VariableStatData,
    constval: &Value,
) -> f64 {
    let Some(column) = vardata.column() else { return DEFAULT_INEQ_SEL };
    let op = match (isgt, iseq) {
        (false, false) => CmpOp::Lt,
        (false, true) => CmpOp::Le,
        (true, false) => CmpOp::Gt,
        (true, true) => CmpOp::Ge,
    };
    let (mcv_selec, sumcommon) = mcv_selectivity(vardata, op, constval, true);
    let hist_selec = ineq_histogram_selectivity(root, vardata, isgt, iseq, constval);
    let mut selec = 1.0 - column.null_frac - sumcommon;
    match hist_selec >= 0.0 {
        true => selec *= hist_selec,
        false => selec *= 0.5,
    }
    selec += mcv_selec;
    selec.clamp(0.0, 1.0)
}

/// mcv_selectivity returns the total share of the most common values that a comparison with a constant keeps, and
/// the total share of all of them, as Postgres' function of the same name does.
fn mcv_selectivity(vardata: &VariableStatData, op: CmpOp, constval: &Value, varonleft: bool) -> (f64, f64) {
    let (mut mcv_selec, mut sumcommon) = (0.0, 0.0);
    for (value, freq) in vardata.column().map_or(&[][..], |c| c.common.as_slice()) {
        let passes = match varonleft {
            true => apply_comparison(op, value, constval),
            false => apply_comparison(op, constval, value),
        };
        if passes == Some(true) {
            mcv_selec += freq;
        }
        sumcommon += freq;
    }
    (mcv_selec, sumcommon)
}

/// ineq_histogram_selectivity returns the share of the histogram's population that an inequality with a constant
/// keeps, interpolating within the bin that holds the constant, or -1 without a histogram, as Postgres' function of
/// the same name does. Doltgres does not probe an index for a column's actual extremes, which Postgres'
/// get_actual_variable_range reads.
fn ineq_histogram_selectivity(
    root: &PlannerInfo<'_, '_>,
    vardata: &VariableStatData,
    isgt: bool,
    iseq: bool,
    constval: &Value,
) -> f64 {
    let Some(column) = vardata.column() else { return -1.0 };
    let values = &column.histogram;
    if values.len() <= 1 {
        return -1.0;
    }
    let op = match iseq {
        true => CmpOp::Le,
        false => CmpOp::Lt,
    };
    let (mut lobound, mut hibound) = (0, values.len());
    while lobound < hibound {
        let probe = (lobound + hibound) / 2;
        let mut ltcmp = apply_comparison(op, &values[probe], constval).unwrap_or(false);
        if isgt {
            ltcmp = !ltcmp;
        }
        match ltcmp {
            true => lobound = probe + 1,
            false => hibound = probe,
        }
    }
    let histfrac = if lobound == 0 {
        0.0
    } else if lobound >= values.len() {
        1.0
    } else {
        let i = lobound;
        let mut eq_selec = 0.0;
        if i == 1 || isgt == iseq {
            let otherdistinct = get_variable_numdistinct(root, vardata).0 - column.common.len() as f64;
            if otherdistinct > 1.0 {
                eq_selec = 1.0 / otherdistinct;
            }
        }
        let binfrac = match convert_to_scalar(constval, &values[i - 1], &values[i]) {
            Some((val, low, high)) => {
                if high <= low {
                    0.5
                } else if val <= low {
                    0.0
                } else if val >= high {
                    1.0
                } else {
                    let binfrac = (val - low) / (high - low);
                    if binfrac.is_nan() || !(0.0..=1.0).contains(&binfrac) { 0.5 } else { binfrac }
                }
            }
            None => 0.5,
        };
        let mut histfrac = ((i - 1) as f64 + binfrac) / (values.len() - 1) as f64;
        if i == 1 {
            histfrac += eq_selec * (1.0 - binfrac);
        }
        if isgt == iseq {
            histfrac -= eq_selec;
        }
        histfrac
    };
    let hist_selec = if isgt { 1.0 - histfrac } else { histfrac };
    let cutoff = 0.01 / (values.len() - 1) as f64;
    hist_selec.clamp(cutoff, 1.0 - cutoff)
}

/// compare_values compares two values of the statistics or of a clause, when they are of comparable kinds.
fn compare_values(a: &Value, b: &Value) -> Option<Ordering> {
    match a.is_null() || b.is_null() {
        true => None,
        false => Some(crate::expr::compare_values(a, b)),
    }
}

/// apply_comparison returns what a comparison of two values gives, as Postgres calls a comparison operator's
/// function on the values of statistics.
fn apply_comparison(op: CmpOp, a: &Value, b: &Value) -> Option<bool> {
    let ordering = compare_values(a, b)?;
    Some(match op {
        CmpOp::Eq => ordering.is_eq(),
        CmpOp::Ne => ordering.is_ne(),
        CmpOp::Lt => ordering.is_lt(),
        CmpOp::Le => ordering.is_le(),
        CmpOp::Gt => ordering.is_gt(),
        CmpOp::Ge => ordering.is_ge(),
    })
}

/// boolvarsel returns the selectivity of a boolean expression, as Postgres' function of the same name does: as the
/// equality `expression = true` when its Var has statistics, and otherwise one half.
pub fn boolvarsel(root: &PlannerInfo<'_, '_>, arg: &Expr, varrelid: usize) -> f64 {
    let vardata = examine_variable(root, arg, varrelid);
    match vardata.column() {
        Some(_) => var_eq_const(root, &vardata, &Value::Bool(true), true, false),
        None => 0.5,
    }
}

/// booltestsel returns the selectivity of an IS TRUE, IS FALSE, or IS UNKNOWN test, or their negations, as Postgres'
/// function of the same name does.
pub fn booltestsel(
    root: &PlannerInfo<'_, '_>,
    value: Option<bool>,
    negated: bool,
    arg: &Expr,
    varrelid: usize,
    jointype: JoinType,
    sjinfo: Option<&SpecialJoinInfo>,
) -> f64 {
    let vardata = examine_variable(root, arg, varrelid);
    let selec = match vardata.column() {
        Some(column) => {
            let freq_null = column.null_frac;
            match column.common.first() {
                Some((first, freq)) => {
                    let freq_true = match first {
                        Value::Bool(true) => *freq,
                        _ => 1.0 - freq - freq_null,
                    };
                    let freq_false = 1.0 - freq_true - freq_null;
                    match (value, negated) {
                        (None, false) => freq_null,
                        (None, true) => 1.0 - freq_null,
                        (Some(true), false) => freq_true,
                        (Some(true), true) => 1.0 - freq_true,
                        (Some(false), false) => freq_false,
                        (Some(false), true) => 1.0 - freq_false,
                    }
                }
                None => match (value, negated) {
                    (None, false) => freq_null,
                    (None, true) => 1.0 - freq_null,
                    (Some(_), false) => (1.0 - freq_null) / 2.0,
                    (Some(_), true) => (freq_null + 1.0) / 2.0,
                },
            }
        }
        None => match (value, negated) {
            (None, false) => DEFAULT_UNK_SEL,
            (None, true) => DEFAULT_NOT_UNK_SEL,
            (Some(true), false) | (Some(false), true) => {
                super::clausesel::clause_selectivity(root, arg, None, varrelid, jointype, sjinfo)
            }
            (Some(false), false) | (Some(true), true) => {
                1.0 - super::clausesel::clause_selectivity(root, arg, None, varrelid, jointype, sjinfo)
            }
        },
    };
    selec.clamp(0.0, 1.0)
}

/// nulltestsel returns the selectivity of an IS NULL or IS NOT NULL test, as Postgres' function of the same name
/// does.
pub fn nulltestsel(root: &PlannerInfo<'_, '_>, negated: bool, arg: &Expr, varrelid: usize) -> f64 {
    let vardata = examine_variable(root, arg, varrelid);
    let selec = match (vardata.column(), negated) {
        (Some(column), false) => column.null_frac,
        (Some(column), true) => 1.0 - column.null_frac,
        (None, false) => DEFAULT_UNK_SEL,
        (None, true) => DEFAULT_NOT_UNK_SEL,
    };
    selec.clamp(0.0, 1.0)
}

/// scalararraysel returns the selectivity of a comparison with any, or every, element of an array, as Postgres'
/// function of the same name does: combining the comparison's selectivity with each element of a constant array or
/// array expression, as disjoint events for an equality with any element, and otherwise assuming ten elements. Doltgres
/// gathers no statistics of array columns, which scalararraysel_containment reads.
#[allow(clippy::too_many_arguments)]
pub fn scalararraysel(
    root: &PlannerInfo<'_, '_>,
    comparison: &Expr,
    array: &Expr,
    use_or: bool,
    is_join_clause: bool,
    varrelid: usize,
    jointype: JoinType,
    sjinfo: Option<&SpecialJoinInfo>,
) -> f64 {
    let Expr::Compare(op, left, right) = comparison else { return 0.5 };
    let (is_equality, is_inequality) = (*op == CmpOp::Eq, *op == CmpOp::Ne);
    let element_sel = |element: &Expr| {
        let substitute = |e: &Expr| match e {
            Expr::SubqueryValue => element.clone(),
            other => estimate_expression_value(other),
        };
        let (l, r) = (substitute(left), substitute(right));
        match (is_join_clause, sjinfo) {
            (true, Some(sjinfo)) => join_selectivity(root, *op, &l, &r, jointype, sjinfo),
            _ => restriction_selectivity(root, *op, &l, &r, varrelid),
        }
    };
    let elements: Option<Vec<Expr>> = match estimate_expression_value(array) {
        Expr::Const(Value::Null) => return 0.0,
        Expr::Const(Value::Array(array)) => Some(array.values.iter().cloned().map(Expr::Const).collect()),
        Expr::Array(_, items, _) => Some(items),
        _ => None,
    };
    let mut s1 = if use_or { 0.0 } else { 1.0 };
    match elements {
        Some(elements) => {
            let mut s1disjoint = s1;
            for element in &elements {
                let s2 = element_sel(element);
                if use_or {
                    s1 = s1 + s2 - s1 * s2;
                    if is_equality {
                        s1disjoint += s2;
                    }
                } else {
                    s1 *= s2;
                    if is_inequality {
                        s1disjoint += s2 - 1.0;
                    }
                }
            }
            if (if use_or { is_equality } else { is_inequality }) && (0.0..=1.0).contains(&s1disjoint) {
                s1 = s1disjoint;
            }
        }
        None => {
            let s2 = element_sel(&Expr::Param(usize::MAX));
            for _ in 0..10 {
                s1 = if use_or { s1 + s2 - s1 * s2 } else { s1 * s2 };
            }
        }
    }
    s1.clamp(0.0, 1.0)
}

/// neqjoinsel returns the selectivity of an inequality join clause, as Postgres' function of the same name does: the
/// share of rows that are not NULL for a semi or anti join, and otherwise the complement of the equality's.
fn neqjoinsel(
    root: &PlannerInfo<'_, '_>,
    left: &Expr,
    right: &Expr,
    jointype: JoinType,
    sjinfo: &SpecialJoinInfo,
) -> f64 {
    match jointype {
        JoinType::Semi | JoinType::Anti => {
            let (leftvar, rightvar, reversed) = get_join_variables(root, left, right, sjinfo);
            let vardata = if reversed { &rightvar } else { &leftvar };
            1.0 - vardata.column().map_or(0.0, |c| c.null_frac)
        }
        _ => 1.0 - eqjoinsel(root, left, right, sjinfo),
    }
}

/// mergejoinscansel returns the shares of each side's rows, in a merge join's order, that come before the first row
/// that can match and up to the last one, as the left start, left end, right start, and right end, from the range of
/// the other side's values, as Postgres' function of the same name does.
pub fn mergejoinscansel(root: &PlannerInfo<'_, '_>, clause: &Expr, descending: bool, nulls_first: bool) -> [f64; 4] {
    let (mut leftstart, mut leftend, mut rightstart, mut rightend) = (0.0, 1.0, 0.0, 1.0);
    let Expr::Compare(CmpOp::Eq, left, right) = clause else { return [leftstart, leftend, rightstart, rightend] };
    let (leftvar, rightvar) = (examine_variable(root, left, 0), examine_variable(root, right, 0));
    let (Some((leftmin, leftmax)), Some((rightmin, rightmax))) =
        (get_variable_range(&leftvar), get_variable_range(&rightvar))
    else {
        return [leftstart, leftend, rightstart, rightend];
    };
    let isgt = descending;
    let ((leftmin, leftmax), (rightmin, rightmax)) = match isgt {
        false => ((leftmin, leftmax), (rightmin, rightmax)),
        true => ((leftmax, leftmin), (rightmax, rightmin)),
    };
    let selec = scalarineqsel(root, isgt, true, &leftvar, &rightmax);
    if selec != DEFAULT_INEQ_SEL {
        leftend = selec;
    }
    let selec = scalarineqsel(root, isgt, true, &rightvar, &leftmax);
    if selec != DEFAULT_INEQ_SEL {
        rightend = selec;
    }
    if leftend > rightend {
        leftend = 1.0;
    } else if leftend < rightend {
        rightend = 1.0;
    } else {
        (leftend, rightend) = (1.0, 1.0);
    }
    let selec = scalarineqsel(root, isgt, false, &leftvar, &rightmin);
    if selec != DEFAULT_INEQ_SEL {
        leftstart = selec;
    }
    let selec = scalarineqsel(root, isgt, false, &rightvar, &leftmin);
    if selec != DEFAULT_INEQ_SEL {
        rightstart = selec;
    }
    if leftstart < rightstart {
        leftstart = 0.0;
    } else if leftstart > rightstart {
        rightstart = 0.0;
    } else {
        (leftstart, rightstart) = (0.0, 0.0);
    }
    if nulls_first {
        if let Some(column) = leftvar.column() {
            leftstart = (leftstart + column.null_frac).clamp(0.0, 1.0);
            leftend = (leftend + column.null_frac).clamp(0.0, 1.0);
        }
        if let Some(column) = rightvar.column() {
            rightstart = (rightstart + column.null_frac).clamp(0.0, 1.0);
            rightend = (rightend + column.null_frac).clamp(0.0, 1.0);
        }
    }
    if leftstart >= leftend {
        (leftstart, leftend) = (0.0, 1.0);
    }
    if rightstart >= rightend {
        (rightstart, rightend) = (0.0, 1.0);
    }
    [leftstart, leftend, rightstart, rightend]
}

/// convert_to_scalar returns a value and the bounds of its histogram bin as numbers on one scale, which
/// interpolation can place the value between, as Postgres' function of the same name does for the numeric, string,
/// bytea, and date and time types, or None for any other.
fn convert_to_scalar(value: &Value, lobound: &Value, hibound: &Value) -> Option<(f64, f64, f64)> {
    match (value, lobound, hibound) {
        (Value::Text(v), Value::Text(lo), Value::Text(hi)) => {
            Some(convert_string_to_scalar(v.as_bytes(), lo.as_bytes(), hi.as_bytes()))
        }
        (Value::Bytea(v), Value::Bytea(lo), Value::Bytea(hi)) => Some(convert_bytea_to_scalar(v, lo, hi)),
        _ => {
            let scalar = |v: &Value| convert_numeric_to_scalar(v).or_else(|| convert_timevalue_to_scalar(v));
            Some((scalar(value)?, scalar(lobound)?, scalar(hibound)?))
        }
    }
}

/// convert_numeric_to_scalar returns a value of a numeric type, a boolean, or an object identifier as a number, as
/// Postgres' function of the same name does.
fn convert_numeric_to_scalar(value: &Value) -> Option<f64> {
    Some(match value {
        Value::Bool(b) => f64::from(u8::from(*b)),
        Value::Int2(v) => f64::from(*v),
        Value::Int4(v) => f64::from(*v),
        Value::Int8(v) => *v as f64,
        Value::Float4(v) => f64::from(*v),
        Value::Float8(v) => *v,
        Value::Numeric(n) => n.to_f64(),
        Value::Oid(o) => f64::from(*o),
        _ => return None,
    })
}

/// convert_timevalue_to_scalar returns a value of a date or time type as microseconds, as Postgres' function of the
/// same name does.
fn convert_timevalue_to_scalar(value: &Value) -> Option<f64> {
    const USECS_PER_DAY: f64 = 86_400_000_000.0;
    Some(match value {
        Value::Timestamp(t) | Value::TimestampTz(t) | Value::Time(t) => *t as f64,
        Value::Date(d) => f64::from(*d) * USECS_PER_DAY,
        Value::Interval(iv) => {
            iv.micros as f64
                + f64::from(iv.days) * USECS_PER_DAY
                + f64::from(iv.months) * ((365.25 / 12.0) * USECS_PER_DAY)
        }
        Value::TimeTz(t, zone) => *t as f64 + f64::from(*zone) * 1_000_000.0,
        _ => return None,
    })
}

/// convert_string_to_scalar returns three strings as numbers on one scale, over the range of the characters they
/// use after their common prefix, as Postgres' function of the same name does.
fn convert_string_to_scalar(value: &[u8], lobound: &[u8], hibound: &[u8]) -> (f64, f64, f64) {
    let (mut rangelo, mut rangehi) = match hibound.first() {
        Some(&c) => (i32::from(c), i32::from(c)),
        None => (0, 0),
    };
    for &c in lobound.iter().chain(hibound) {
        rangelo = rangelo.min(i32::from(c));
        rangehi = rangehi.max(i32::from(c));
    }
    for (lo, hi) in [(b'A', b'Z'), (b'a', b'z'), (b'0', b'9')] {
        if rangelo <= i32::from(hi) && rangehi >= i32::from(lo) {
            rangelo = rangelo.min(i32::from(lo));
            rangehi = rangehi.max(i32::from(hi));
        }
    }
    if rangehi - rangelo < 9 {
        rangelo = i32::from(b' ');
        rangehi = 127;
    }
    let prefix = lobound.iter().zip(hibound).zip(value).take_while(|((l, h), v)| l == h && l == v).count();
    let convert = |s: &[u8]| convert_one_string_to_scalar(&s[prefix.min(s.len())..], rangelo, rangehi);
    (convert(value), convert(lobound), convert(hibound))
}

/// convert_one_string_to_scalar returns a string as a fraction in base of its character range, from its first
/// twelve characters, as Postgres' function of the same name does.
fn convert_one_string_to_scalar(value: &[u8], rangelo: i32, rangehi: i32) -> f64 {
    let base = f64::from(rangehi - rangelo + 1);
    let (mut num, mut denom) = (0.0, base);
    for &c in value.iter().take(12) {
        let ch = i32::from(c).clamp(rangelo - 1, rangehi + 1);
        num += f64::from(ch - rangelo) / denom;
        denom *= base;
    }
    num
}

/// convert_bytea_to_scalar returns three byte strings as numbers on one scale after their common prefix, as Postgres'
/// function of the same name does.
fn convert_bytea_to_scalar(value: &[u8], lobound: &[u8], hibound: &[u8]) -> (f64, f64, f64) {
    let prefix = lobound.iter().zip(hibound).zip(value).take_while(|((l, h), v)| l == h && l == v).count();
    let convert = |s: &[u8]| convert_one_bytea_to_scalar(&s[prefix.min(s.len())..], 0, 255);
    (convert(value), convert(lobound), convert(hibound))
}

/// convert_one_bytea_to_scalar returns a byte string as a fraction in base 256, from its first ten bytes, as
/// Postgres' function of the same name does.
fn convert_one_bytea_to_scalar(value: &[u8], rangelo: i32, rangehi: i32) -> f64 {
    let base = f64::from(rangehi - rangelo + 1);
    let (mut num, mut denom) = (0.0, base);
    for &c in value.iter().take(10) {
        num += f64::from(i32::from(c) - rangelo) / denom;
        denom *= base;
    }
    num
}

/// get_variable_range returns the smallest and largest values of a variable that its statistics know, from its
/// histogram's ends and its most common values, as Postgres' function of the same name does.
pub fn get_variable_range(vardata: &VariableStatData) -> Option<(Value, Value)> {
    let column = vardata.column()?;
    let mut range = match (column.histogram.first(), column.histogram.last()) {
        (Some(min), Some(max)) => Some((min.clone(), max.clone())),
        _ => None,
    };
    let sumcommon: f64 = column.common.iter().map(|(_, f)| f).sum();
    if range.is_some() || sumcommon + column.null_frac > 0.99999 {
        for (value, _) in &column.common {
            range = Some(match range {
                None => (value.clone(), value.clone()),
                Some((min, max)) => {
                    let min = if compare_values(value, &min) == Some(Ordering::Less) { value.clone() } else { min };
                    let max = if compare_values(value, &max) == Some(Ordering::Greater) { value.clone() } else { max };
                    (min, max)
                }
            });
        }
    }
    range
}

/// get_variable_numdistinct returns about how many distinct values an expression takes, and whether that is only
/// the default, as Postgres' function of the same name does.
pub fn get_variable_numdistinct(root: &PlannerInfo<'_, '_>, vardata: &VariableStatData) -> (f64, bool) {
    let column = vardata.column();
    let stanullfrac = column.map_or(0.0, |c| c.null_frac);
    let mut stadistinct = match column {
        Some(c) => c.stadistinct,
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
    let (vardata1, vardata2) = (examine_variable(root, left, 0), examine_variable(root, right, 0));
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
    let vardata = examine_variable(root, hashkey, 0);
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
            VariableStatData { var: Expr::Column(attno), rel: Some(rel), stats, isunique, isbool }
        }
        None => VariableStatData {
            var: Expr::Const(Value::Null),
            rel: Some(rel),
            stats: None,
            isunique: false,
            isbool: false,
        },
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
    /// Whether the distinct values are only the default estimate.
    isdefault: bool,
}

/// EstimationInfo is what estimate_num_groups reports about its estimate, as Postgres' EstimationInfo does: whether
/// it used a default estimate for some Var, Postgres' SELFLAG_USED_DEFAULT.
#[derive(Clone, Copy, Debug, Default)]
pub struct EstimationInfo {
    pub used_default: bool,
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
    mut estinfo: Option<&mut EstimationInfo>,
) -> f64 {
    if let Some(estinfo) = estinfo.as_deref_mut() {
        *estinfo = EstimationInfo::default();
    }
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
        let vardata = examine_variable(root, groupexpr, 0);
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
            let vardata = examine_variable(root, &var, 0);
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
            if let Some(estinfo) = estinfo.as_deref_mut()
                && varinfo.isdefault
            {
                estinfo.used_default = true;
            }
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
    let (ndistinct, isdefault) = get_variable_numdistinct(root, vardata);
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
    varinfos.push(GroupVarInfo { var, rel: vardata.rel, ndistinct, isdefault });
}

/// estimate_hashagg_tablesize estimates the memory that a hashed aggregation of a path's rows into a number of groups
/// takes, as Postgres' function of the same name does.
pub fn estimate_hashagg_tablesize(
    root: &PlannerInfo<'_, '_>,
    path: &super::nodes::Path,
    agg_costs: &super::prepagg::AggClauseCosts,
    num_groups: f64,
) -> f64 {
    let width = super::pathnode::path_target(root, path).width;
    let hashentrysize =
        super::costsize::hash_agg_entry_size(root.parse.aggregates.len(), width, agg_costs.transition_space);
    hashentrysize * num_groups
}
