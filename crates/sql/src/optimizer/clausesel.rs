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

//! Postgres' optimizer/path/clausesel.c: the share of rows that a list of clauses keeps.

use std::rc::Rc;

use super::PlannerInfo;
use super::clauses::{contain_volatile_functions, pull_varnos};
use super::nodes::{JoinType, RestrictInfo, SpecialJoinInfo, var_parts};
use crate::colstats::TableStats;
use crate::expr::{CmpOp, Expr};
use crate::types::Value;

/// DEFAULT_INEQ_SEL is the share of rows that Postgres assumes an inequality keeps without statistics.
const DEFAULT_INEQ_SEL: f64 = 1.0 / 3.0;

/// DEFAULT_RANGE_INEQ_SEL is the share of rows that Postgres assumes a pair of bounds on one value keeps without
/// statistics.
const DEFAULT_RANGE_INEQ_SEL: f64 = 0.005;

/// RangeQueryClause is the selectivities of the lower and upper bounds that a list's clauses put on one expression.
struct RangeQueryClause<'e> {
    var: &'e Expr,
    lobound: Option<f64>,
    hibound: Option<f64>,
}

/// clauselist_selectivity returns the share of rows that a list of restriction clauses keeps, as Postgres'
/// clauselist_selectivity estimates it, for the relation at a range table index or, when it is zero, for the join
/// that a SpecialJoinInfo describes.
pub fn clauselist_selectivity(
    root: &mut PlannerInfo<'_, '_>,
    clauses: &[Rc<RestrictInfo>],
    varrelid: usize,
    jointype: JoinType,
    sjinfo: Option<&SpecialJoinInfo>,
) -> f64 {
    let clauses: Vec<(&Expr, Option<&RestrictInfo>)> = clauses.iter().map(|r| (&r.clause, Some(&**r))).collect();
    list_selectivity(root, &clauses, varrelid, jointype, sjinfo)
}

/// list_selectivity is clauselist_selectivity for clauses that may lack a RestrictInfo, such as an AND's arguments.
/// A pair of bounds on one expression keeps the rows between them rather than the product of their shares.
fn list_selectivity(
    root: &mut PlannerInfo<'_, '_>,
    clauses: &[(&Expr, Option<&RestrictInfo>)],
    varrelid: usize,
    jointype: JoinType,
    sjinfo: Option<&SpecialJoinInfo>,
) -> f64 {
    if let [(clause, rinfo)] = clauses {
        return clause_selectivity(root, clause, *rinfo, varrelid, jointype, sjinfo);
    }
    let mut s1 = 1.0;
    let mut rqlist: Vec<RangeQueryClause> = Vec::new();
    for &(clause, rinfo) in clauses {
        let s2 = clause_selectivity(root, clause, rinfo, varrelid, jointype, sjinfo);
        if rinfo.is_some_and(|r| r.pseudoconstant) {
            s1 *= s2;
            continue;
        }
        let Expr::Compare(op, l, r) = clause else {
            s1 *= s2;
            continue;
        };
        let single = match rinfo {
            Some(rinfo) => rinfo.clause_relids.count_ones() == 1,
            None => pull_varnos(clause).count_ones() == 1,
        };
        let varonleft = match (is_pseudo_constant(r), is_pseudo_constant(l)) {
            (true, _) => true,
            (false, true) => false,
            (false, false) => {
                s1 *= s2;
                continue;
            }
        };
        let is_lt = match op {
            CmpOp::Lt | CmpOp::Le if single => true,
            CmpOp::Gt | CmpOp::Ge if single => false,
            _ => {
                s1 *= s2;
                continue;
            }
        };
        let (var, is_lobound) = if varonleft { (&**l, !is_lt) } else { (&**r, is_lt) };
        add_range_clause(&mut rqlist, var, is_lobound, s2);
    }
    for rq in rqlist {
        s1 *= match (rq.lobound, rq.hibound) {
            (Some(lo), Some(hi)) if lo == DEFAULT_INEQ_SEL || hi == DEFAULT_INEQ_SEL => DEFAULT_RANGE_INEQ_SEL,
            (Some(lo), Some(hi)) => {
                let null = Expr::IsNull(Box::new(rq.var.clone()), false);
                let s2 = hi + lo - 1.0 + clause_selectivity(root, &null, None, varrelid, jointype, sjinfo);
                match s2 {
                    s2 if s2 < -0.01 => DEFAULT_RANGE_INEQ_SEL,
                    s2 if s2 <= 0.0 => 1.0e-10,
                    s2 => s2,
                }
            }
            (Some(bound), None) | (None, Some(bound)) => bound,
            (None, None) => 1.0,
        };
    }
    s1
}

/// add_range_clause adds a lower or upper bound's selectivity to the bounds of its expression, keeping the more
/// restrictive of two bounds on the same side, as Postgres' addRangeClause does.
fn add_range_clause<'e>(rqlist: &mut Vec<RangeQueryClause<'e>>, var: &'e Expr, is_lobound: bool, s2: f64) {
    let rq = match rqlist.iter_mut().find(|rq| rq.var == var) {
        Some(rq) => rq,
        None => {
            rqlist.push(RangeQueryClause { var, lobound: None, hibound: None });
            rqlist.last_mut().expect("just pushed")
        }
    };
    let bound = if is_lobound { &mut rq.lobound } else { &mut rq.hibound };
    *bound = Some(bound.map_or(s2, |b| b.min(s2)));
}

/// is_pseudo_constant reports whether an expression reads no Var of the query and calls no volatile function, so it
/// is constant for one run of the plan, as Postgres' is_pseudo_constant_clause does.
fn is_pseudo_constant(e: &Expr) -> bool {
    pull_varnos(e) == 0 && !contain_volatile_functions(e)
}

/// clause_selectivity returns the share of rows that one clause keeps, as Postgres' clause_selectivity estimates it:
/// as a join clause when it reads several relations and a join is given, and otherwise as a restriction of one.
pub fn clause_selectivity(
    root: &mut PlannerInfo<'_, '_>,
    clause: &Expr,
    rinfo: Option<&RestrictInfo>,
    varrelid: usize,
    jointype: JoinType,
    sjinfo: Option<&SpecialJoinInfo>,
) -> f64 {
    if let Some(rinfo) = rinfo
        && rinfo.pseudoconstant
        && !matches!(clause, Expr::Const(_))
    {
        return 1.0;
    }
    let treat_as_join_clause = || {
        let relids = rinfo.map_or_else(|| pull_varnos(clause), |r| r.clause_relids);
        varrelid == 0 && sjinfo.is_some() && relids.count_ones() > 1
    };
    match clause {
        Expr::Const(Value::Bool(b)) => f64::from(u8::from(*b)),
        Expr::Const(Value::Null) => 0.0,
        Expr::Not(inner) => 1.0 - clause_selectivity(root, inner, None, varrelid, jointype, sjinfo),
        Expr::And(..) => {
            let args: Vec<(&Expr, Option<&RestrictInfo>)> =
                crate::indexscan::conjuncts(clause).into_iter().map(|c| (c, None)).collect();
            list_selectivity(root, &args, varrelid, jointype, sjinfo)
        }
        Expr::Or(a, b) => {
            let s1 = clause_selectivity(root, a, None, varrelid, jointype, sjinfo);
            let s2 = clause_selectivity(root, b, None, varrelid, jointype, sjinfo);
            s1 + s2 - s1 * s2
        }
        Expr::Compare(op, l, r) if treat_as_join_clause() => {
            let sjinfo = sjinfo.expect("join clauses have a join");
            match op {
                CmpOp::Eq => super::selfuncs::eqjoinsel(root, l, r, sjinfo),
                CmpOp::Ne => 1.0 - super::selfuncs::eqjoinsel(root, l, r, sjinfo),
                _ => DEFAULT_INEQ_SEL,
            }
        }
        other => restriction_selectivity(root, other, varrelid),
    }
}

/// restriction_selectivity returns the share of a relation's rows that a clause keeps, from the statistics of the
/// relation at the range table index, or of the one relation that the clause reads when it is zero, as Postgres'
/// restriction estimators find them, where the Vars of other relations are unknown values.
fn restriction_selectivity(root: &mut PlannerInfo<'_, '_>, clause: &Expr, varrelid: usize) -> f64 {
    let relids = pull_varnos(clause);
    let varno = match varrelid {
        0 if relids.count_ones() == 1 => relids.trailing_zeros() as usize,
        varrelid => varrelid,
    };
    let table = (varno != 0).then(|| root.parse.rte(varno).table().cloned()).flatten();
    let stats = table.and_then(|table| crate::colstats::table_stats(root.ctx, &table));
    let local = to_attnos(clause, varno);
    match stats {
        Some(stats) => crate::colstats::selectivity(&stats, &local),
        None => crate::colstats::selectivity(&TableStats::default(), &local),
    }
}

/// to_attnos rewrites a clause's Vars of the relation at a range table index into columns of its rows, and the
/// Vars of other relations into parameters, whose values are unknown.
fn to_attnos(e: &Expr, varno: usize) -> Expr {
    match e {
        Expr::Column(c) => match var_parts(*c) {
            (v, attno) if v == varno => Expr::Column(attno),
            _ => Expr::Param(usize::MAX),
        },
        other => other.clone().map_children(&mut |c| to_attnos(&c, varno)),
    }
}
