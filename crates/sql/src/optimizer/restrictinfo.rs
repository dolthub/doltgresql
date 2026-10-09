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

//! Postgres' optimizer/util/restrictinfo.c: building the RestrictInfo of a clause, and the tests of where a clause
//! may be evaluated.

use std::cell::Cell;

use super::PlannerInfo;
use super::nodes::{Relids, RestrictInfo, RinfoId};
use super::var::pull_varnos;
use crate::expr::Expr;

/// RestrictInfoArgs is what make_restrictinfo takes besides the clause, as Postgres' function of the same name
/// takes it.
#[derive(Clone, Default)]
pub struct RestrictInfoArgs {
    pub is_pushed_down: bool,
    pub has_clone: bool,
    pub is_clone: bool,
    pub pseudoconstant: bool,
    pub security_level: usize,
    /// The relations the clause needs, or None for those it reads.
    pub required_relids: Option<Relids>,
    pub incompatible_relids: Relids,
    pub outer_relids: Relids,
}

/// binary_op_args returns the arguments of an operator of two arguments, as Postgres' is_opclause and get_leftop and
/// get_rightop find them.
pub fn binary_op_args(e: &Expr) -> Option<(&Expr, &Expr)> {
    match e {
        Expr::Compare(_, l, r)
        | Expr::Operator(_, _, l, r)
        | Expr::Arith(_, l, r, _)
        | Expr::Concat(l, r)
        | Expr::DateTime(_, l, r)
        | Expr::ArrayOp(_, l, r) => Some((l, r)),
        _ => None,
    }
}

/// or_args returns the arguments of an OR, flattened.
pub fn or_args(e: &Expr) -> Vec<&Expr> {
    match e {
        Expr::Or(a, b) => [or_args(a), or_args(b)].concat(),
        other => vec![other],
    }
}

/// and_args returns the arguments of an AND, flattened.
pub fn and_args(e: &Expr) -> Vec<&Expr> {
    match e {
        Expr::And(a, b) => [and_args(a), and_args(b)].concat(),
        other => vec![other],
    }
}

/// make_restrictinfo builds the RestrictInfo of a clause and returns its ID, giving an OR clause the RestrictInfos of
/// its arguments, as Postgres' function of the same name does.
pub fn make_restrictinfo(root: &mut PlannerInfo<'_, '_>, clause: Expr, args: RestrictInfoArgs) -> RinfoId {
    if matches!(clause, Expr::Or(..)) {
        return make_sub_restrictinfos(root, clause, args);
    }
    make_plain_restrictinfo(root, clause, None, args)
}

/// make_plain_restrictinfo builds the RestrictInfo of a clause that is not an OR, or of an OR with its arguments'
/// RestrictInfos, as Postgres' function of the same name does.
pub fn make_plain_restrictinfo(
    root: &mut PlannerInfo<'_, '_>,
    clause: Expr,
    orclause: Option<Vec<Vec<RinfoId>>>,
    args: RestrictInfoArgs,
) -> RinfoId {
    let (left_relids, right_relids, clause_relids, can_join) = match binary_op_args(&clause) {
        Some((l, r)) => {
            let (left, right) = (pull_varnos(root, l), pull_varnos(root, r));
            let clause_relids = left.union(&right);
            let can_join = !left.is_empty() && !right.is_empty() && !left.overlap(&right);
            (left, right, clause_relids, can_join)
        }
        None => (Relids::new(), Relids::new(), pull_varnos(root, &clause), false),
    };
    let required_relids = args.required_relids.unwrap_or_else(|| clause_relids.clone());
    root.last_rinfo_serial += 1;
    let rinfo = RestrictInfo {
        clause,
        is_pushed_down: args.is_pushed_down,
        can_join,
        pseudoconstant: args.pseudoconstant,
        has_clone: args.has_clone,
        is_clone: args.is_clone,
        security_level: args.security_level,
        clause_relids,
        required_relids,
        incompatible_relids: args.incompatible_relids,
        outer_relids: args.outer_relids,
        left_relids,
        right_relids,
        orclause,
        rinfo_serial: root.last_rinfo_serial,
        parent_ec: None,
        norm_selec: Cell::new(-1.0),
        outer_selec: Cell::new(-1.0),
        mergeopfamilies: Vec::new(),
        left_ec: None,
        right_ec: None,
        left_em: None,
        right_em: None,
        outer_is_left: Cell::new(false),
        hashjoinable: false,
    };
    root.rinfos.push(rinfo);
    root.rinfos.len() - 1
}

/// make_sub_restrictinfos builds the RestrictInfo of an OR clause, with RestrictInfos of the conjuncts of each of
/// its arguments, as Postgres' function of the same name does.
fn make_sub_restrictinfos(root: &mut PlannerInfo<'_, '_>, clause: Expr, args: RestrictInfoArgs) -> RinfoId {
    let sub_args = RestrictInfoArgs { required_relids: None, ..args.clone() };
    let orlist = or_args(&clause)
        .into_iter()
        .map(|arm| {
            and_args(arm)
                .into_iter()
                .map(|c| match c {
                    Expr::Or(..) => make_sub_restrictinfos(root, c.clone(), sub_args.clone()),
                    other => make_plain_restrictinfo(root, other.clone(), None, sub_args.clone()),
                })
                .collect()
        })
        .collect();
    make_plain_restrictinfo(root, clause, Some(orlist), args)
}

/// commute_restrictinfo returns the RestrictInfo of a comparison with its sides swapped, as Postgres' function of the
/// same name does.
pub fn commute_restrictinfo(root: &mut PlannerInfo<'_, '_>, rinfo: RinfoId) -> RinfoId {
    let mut result = root.rinfos[rinfo].clone();
    let Expr::Compare(op, left, right) = result.clause else { unreachable!("only a comparison is commuted") };
    result.clause = Expr::Compare(crate::indexscan::swap(op), right, left);
    std::mem::swap(&mut result.left_relids, &mut result.right_relids);
    std::mem::swap(&mut result.left_ec, &mut result.right_ec);
    std::mem::swap(&mut result.left_em, &mut result.right_em);
    result.outer_selec = Cell::new(-1.0);
    root.rinfos.push(result);
    root.rinfos.len() - 1
}

/// restriction_is_or_clause reports whether a RestrictInfo is of an OR clause, as Postgres' function of the same name
/// does.
pub fn restriction_is_or_clause(rinfo: &RestrictInfo) -> bool {
    rinfo.orclause.is_some()
}

/// rinfo_is_pushed_down reports whether a clause is evaluated as a filter on a join's result rather than as the
/// join's own condition, as Postgres' RINFO_IS_PUSHED_DOWN does.
pub fn rinfo_is_pushed_down(rinfo: &RestrictInfo, joinrelids: &Relids) -> bool {
    rinfo.is_pushed_down || !rinfo.required_relids.is_subset(joinrelids)
}

/// join_clause_is_movable_into reports whether a join clause can be evaluated at a relation whose paths are
/// parameterized by others, as Postgres' function of the same name does.
pub fn join_clause_is_movable_into(rinfo: &RestrictInfo, currentrelids: &Relids, current_and_outer: &Relids) -> bool {
    rinfo.clause_relids.is_subset(current_and_outer)
        && currentrelids.overlap(&rinfo.clause_relids)
        && !currentrelids.overlap(&rinfo.outer_relids)
}

/// join_clause_is_movable_to reports whether a join clause can be evaluated at a base relation's scan, when the
/// relation's paths are parameterized by the other relations it reads, as Postgres' function of the same name does.
pub fn join_clause_is_movable_to(rinfo: &RestrictInfo, baserel: &super::nodes::RelOptInfo) -> bool {
    rinfo.clause_relids.is_member(baserel.relid)
        && !rinfo.outer_relids.is_member(baserel.relid)
        && !rinfo.clause_relids.overlap(&baserel.nulling_relids)
        && !baserel.lateral_referencers.overlap(&rinfo.clause_relids)
        && !rinfo.is_clone
}
