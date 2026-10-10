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

use super::PlannerInfo;
use super::clauses::contain_volatile_functions;
use super::nodes::{JoinType, RestrictInfo, RinfoId, SpecialJoinInfo, VarNode};
use super::selfuncs;
use super::var::pull_varnos;
use crate::expr::{CmpOp, Expr};
use crate::types::Value;

/// DEFAULT_INEQ_SEL is the share of rows that Postgres assumes an inequality keeps without statistics.
pub const DEFAULT_INEQ_SEL: f64 = 1.0 / 3.0;

/// DEFAULT_RANGE_INEQ_SEL is the share of rows that Postgres assumes a pair of bounds on one value keeps without
/// statistics.
pub const DEFAULT_RANGE_INEQ_SEL: f64 = 0.005;

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
    root: &PlannerInfo<'_, '_>,
    clauses: &[RinfoId],
    varrelid: usize,
    jointype: JoinType,
    sjinfo: Option<&SpecialJoinInfo>,
) -> f64 {
    let clauses: Vec<(&Expr, Option<&RestrictInfo>)> =
        clauses.iter().map(|&r| (&root.rinfos[r].clause, Some(&root.rinfos[r]))).collect();
    list_selectivity(root, &clauses, varrelid, jointype, sjinfo)
}

/// list_selectivity is clauselist_selectivity for clauses that may lack a RestrictInfo, such as an AND's arguments.
/// A pair of bounds on one expression keeps the rows between them rather than the product of their shares.
pub fn list_selectivity(
    root: &PlannerInfo<'_, '_>,
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
            Some(rinfo) => rinfo.clause_relids.difference(&root.outer_join_rels).num_members() == 1,
            None => pull_varnos(root, clause).difference(&root.outer_join_rels).num_members() == 1,
        };
        let varonleft = match (is_pseudo_constant(root, r), is_pseudo_constant(root, l)) {
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
fn is_pseudo_constant(root: &PlannerInfo<'_, '_>, e: &Expr) -> bool {
    pull_varnos(root, e).is_empty() && !contain_volatile_functions(root.glob, e)
}

/// clause_selectivity returns the share of rows that one clause keeps, as Postgres' clause_selectivity estimates it:
/// as a join clause when it reads several relations and a join is given, and otherwise as a restriction of one.
pub fn clause_selectivity(
    root: &PlannerInfo<'_, '_>,
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
    let cache =
        rinfo.filter(|r| varrelid == 0 || r.clause_relids == super::nodes::Relids::singleton(varrelid)).map(|r| {
            match jointype {
                JoinType::Inner => &r.norm_selec,
                _ => &r.outer_selec,
            }
        });
    if let Some(cached) = cache
        && cached.get() >= 0.0
    {
        return cached.get();
    }
    let s1 = clause_selectivity_uncached(root, clause, rinfo, varrelid, jointype, sjinfo);
    if let Some(cached) = cache {
        cached.set(s1);
    }
    s1
}

/// clause_selectivity_uncached is clause_selectivity for a clause whose selectivity is not cached.
fn clause_selectivity_uncached(
    root: &PlannerInfo<'_, '_>,
    clause: &Expr,
    rinfo: Option<&RestrictInfo>,
    varrelid: usize,
    jointype: JoinType,
    sjinfo: Option<&SpecialJoinInfo>,
) -> f64 {
    let is_join = || treat_as_join_clause(root, clause, rinfo, varrelid, sjinfo);
    match clause {
        Expr::Column(id) => match root.glob.node(*id) {
            VarNode::Var(var) if varrelid == 0 || varrelid == var.varno => selfuncs::boolvarsel(root, clause, varrelid),
            VarNode::Var(_) => 0.5,
            VarNode::PlaceHolderVar(phv) => {
                let phexpr = root.glob.placeholder(phv.phid).phexpr.clone();
                clause_selectivity(root, &phexpr, None, varrelid, jointype, sjinfo)
            }
        },
        Expr::Const(Value::Bool(b)) => f64::from(u8::from(*b)),
        Expr::Const(Value::Null) => 0.0,
        Expr::Param(_) | Expr::Outer(..) => 0.5,
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
        Expr::Compare(op, l, r) => match (is_join(), sjinfo) {
            (true, Some(sjinfo)) => selfuncs::join_selectivity(root, *op, l, r, jointype, sjinfo),
            _ => selfuncs::restriction_selectivity(root, *op, l, r, varrelid),
        },
        Expr::DistinctFrom(l, r, negated) => {
            let s1 = match (is_join(), sjinfo) {
                (true, Some(sjinfo)) => selfuncs::join_selectivity(root, CmpOp::Eq, l, r, jointype, sjinfo),
                _ => selfuncs::restriction_selectivity(root, CmpOp::Eq, l, r, varrelid),
            };
            match negated {
                true => s1,
                false => 1.0 - s1,
            }
        }
        Expr::Func(..) | Expr::Routine(..) => selfuncs::function_selectivity(),
        Expr::AnyArray(comparison, array, all) => {
            selfuncs::scalararraysel(root, comparison, array, !all, is_join(), varrelid, jointype, sjinfo)
        }
        Expr::RowCompare(op, l, r) => {
            super::selfuncs::rowcomparesel(root, *op, (&l[0], &r[0]), varrelid, jointype, sjinfo)
        }
        Expr::IsNull(arg, negated) => selfuncs::nulltestsel(root, *negated, arg, varrelid),
        Expr::BoolTest(arg, value, negated) => {
            selfuncs::booltestsel(root, *value, *negated, arg, varrelid, jointype, sjinfo)
        }
        Expr::Cast(inner, ty, _) if super::nodefuncs::expr_type(root, inner) == Some(ty.oid) => {
            clause_selectivity(root, inner, None, varrelid, jointype, sjinfo)
        }
        Expr::Cast(..) => selfuncs::function_selectivity(),
        other => selfuncs::boolvarsel(root, other, varrelid),
    }
}

/// treat_as_join_clause reports whether to estimate a clause as a join clause, as Postgres' function of the same
/// name decides: when no relation is being restricted, a join is given, and the clause reads more than one base
/// relation.
fn treat_as_join_clause(
    root: &PlannerInfo<'_, '_>,
    clause: &Expr,
    rinfo: Option<&RestrictInfo>,
    varrelid: usize,
    sjinfo: Option<&SpecialJoinInfo>,
) -> bool {
    if varrelid != 0 || sjinfo.is_none() {
        return false;
    }
    let relids = rinfo.map_or_else(|| pull_varnos(root, clause), |r| r.clause_relids.clone());
    relids.difference(&root.outer_join_rels).num_members() > 1
}
