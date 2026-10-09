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

//! Postgres' optimizer/plan/initsplan.c: breaking the join tree into the relations to join, the clauses that restrict
//! them, and the outer joins that constrain their join order.

use std::rc::Rc;

use super::PlannerInfo;
use super::clauses::{contain_volatile_functions, find_nonnullable_rels, pull_varnos};
use super::nodes::{FromExpr, JoinTreeNode, JoinType, Relids, RestrictInfo, SpecialJoinInfo, is_subset, overlap};
use super::restrictinfo::make_restrictinfo;
use crate::expr::Expr;

/// JoinList is a list of relations to join in any order, where a member is a range table index or a list that must
/// be joined on its own first, as Postgres' joinlists are.
#[derive(Clone, Debug)]
pub enum JoinList {
    Rel(usize),
    List(Vec<JoinList>),
}

/// JOIN_COLLAPSE_LIMIT and FROM_COLLAPSE_LIMIT are the most members that a joinlist collapses into one search
/// problem, as Postgres' settings of the same names default to.
const JOIN_COLLAPSE_LIMIT: usize = 8;
const FROM_COLLAPSE_LIMIT: usize = 8;

/// deconstruct_jointree distributes the query's clauses to the relations they restrict and records its outer joins,
/// returning the joinlist of relations to join, as Postgres' deconstruct_jointree does.
pub fn deconstruct_jointree(root: &mut PlannerInfo<'_, '_>, jointree: &FromExpr) -> Vec<JoinList> {
    let (mut qualscope, mut inner_join_rels) = (0, 0);
    deconstruct_from(root, jointree, false, &mut qualscope, &mut inner_join_rels)
}

/// deconstruct_from is deconstruct_recurse for a FROM list.
fn deconstruct_from(
    root: &mut PlannerInfo<'_, '_>,
    f: &FromExpr,
    below_outer_join: bool,
    qualscope: &mut Relids,
    inner_join_rels: &mut Relids,
) -> Vec<JoinList> {
    *qualscope = 0;
    *inner_join_rels = 0;
    let mut joinlist = Vec::new();
    let mut remaining = f.fromlist.len();
    for node in &f.fromlist {
        let mut sub_qualscope = 0;
        let sub_joinlist = deconstruct_recurse(root, node, below_outer_join, &mut sub_qualscope, inner_join_rels);
        *qualscope |= sub_qualscope;
        remaining -= 1;
        if sub_joinlist.len() <= 1 || joinlist.len() + sub_joinlist.len() + remaining <= FROM_COLLAPSE_LIMIT {
            joinlist.extend(sub_joinlist);
        } else {
            joinlist.push(JoinList::List(sub_joinlist));
        }
    }
    if f.fromlist.len() > 1 {
        *inner_join_rels = *qualscope;
    }
    for qual in &f.quals {
        distribute_qual_to_rels(root, qual, below_outer_join, *qualscope, 0, 0);
    }
    joinlist
}

/// deconstruct_recurse is one level of deconstruct_jointree: it distributes a join tree node's clauses, records an
/// outer join as a SpecialJoinInfo, and returns the node's joinlist, given whether the node is on the nullable side
/// of a higher outer join, and sets the relations the node includes and those under its inner joins.
fn deconstruct_recurse(
    root: &mut PlannerInfo<'_, '_>,
    node: &JoinTreeNode,
    below_outer_join: bool,
    qualscope: &mut Relids,
    inner_join_rels: &mut Relids,
) -> Vec<JoinList> {
    let j = match node {
        JoinTreeNode::Rel(varno) => {
            *qualscope = super::nodes::singleton(*varno);
            *inner_join_rels = 0;
            return vec![JoinList::Rel(*varno)];
        }
        JoinTreeNode::From(f) => return deconstruct_from(root, f, below_outer_join, qualscope, inner_join_rels),
        JoinTreeNode::Join(j) => j,
    };
    let (mut leftids, mut rightids, mut left_inners, mut right_inners) = (0, 0, 0, 0);
    let (left_below, right_below) = match j.jointype {
        JoinType::Inner | JoinType::Semi => (below_outer_join, below_outer_join),
        JoinType::Left | JoinType::Anti => (below_outer_join, true),
        JoinType::Full => (true, true),
        JoinType::Right => unreachable!("reduce_outer_joins turns right joins into left joins"),
    };
    let leftjoinlist = deconstruct_recurse(root, &j.larg, left_below, &mut leftids, &mut left_inners);
    let rightjoinlist = deconstruct_recurse(root, &j.rarg, right_below, &mut rightids, &mut right_inners);
    *qualscope = leftids | rightids;
    *inner_join_rels = match j.jointype {
        JoinType::Inner => *qualscope,
        _ => left_inners | right_inners,
    };
    let nonnullable_rels = match j.jointype {
        JoinType::Left | JoinType::Anti => leftids,
        JoinType::Full => *qualscope,
        _ => 0,
    };
    let (sjinfo, ojscope) = match j.jointype {
        JoinType::Inner => (None, 0),
        jointype => {
            let sjinfo = make_outerjoininfo(root, leftids, rightids, *inner_join_rels, jointype, &j.quals);
            let ojscope = if jointype == JoinType::Semi { 0 } else { sjinfo.min_lefthand | sjinfo.min_righthand };
            (Some(sjinfo), ojscope)
        }
    };
    for qual in &j.quals {
        distribute_qual_to_rels(root, qual, below_outer_join, *qualscope, ojscope, nonnullable_rels);
    }
    if let Some(sjinfo) = sjinfo {
        root.join_info_list.push(sjinfo);
    }
    if j.jointype == JoinType::Full {
        vec![JoinList::List(vec![JoinList::List(leftjoinlist), JoinList::List(rightjoinlist)])]
    } else if leftjoinlist.len() + rightjoinlist.len() <= JOIN_COLLAPSE_LIMIT {
        leftjoinlist.into_iter().chain(rightjoinlist).collect()
    } else {
        let part = |list: Vec<JoinList>| match list.len() {
            1 => list.into_iter().next().expect("one member"),
            _ => JoinList::List(list),
        };
        vec![part(leftjoinlist), part(rightjoinlist)]
    }
}

/// make_outerjoininfo builds the SpecialJoinInfo of an outer, semi, or anti join, with the least sets of relations
/// that its sides must hold before it can be formed, as Postgres' make_outerjoininfo computes them.
fn make_outerjoininfo(
    root: &PlannerInfo<'_, '_>,
    left_rels: Relids,
    right_rels: Relids,
    inner_join_rels: Relids,
    jointype: JoinType,
    clause: &[Expr],
) -> SpecialJoinInfo {
    let mut sjinfo = SpecialJoinInfo {
        min_lefthand: left_rels,
        min_righthand: right_rels,
        syn_lefthand: left_rels,
        syn_righthand: right_rels,
        jointype,
        lhs_strict: false,
        delay_upper_joins: false,
    };
    if jointype == JoinType::Full {
        return sjinfo;
    }
    let clause_relids = clause.iter().fold(0, |relids, c| relids | pull_varnos(c));
    let strict_relids = clause.iter().fold(0, |relids, c| relids | find_nonnullable_rels(c));
    sjinfo.lhs_strict = overlap(strict_relids, left_rels);
    let mut min_lefthand = clause_relids & left_rels;
    let mut min_righthand = (clause_relids | inner_join_rels) & right_rels;
    for other in &root.join_info_list {
        if other.jointype == JoinType::Full {
            if overlap(left_rels, other.syn_lefthand) || overlap(left_rels, other.syn_righthand) {
                min_lefthand |= other.syn_lefthand | other.syn_righthand;
            }
            if overlap(right_rels, other.syn_lefthand) || overlap(right_rels, other.syn_righthand) {
                min_righthand |= other.syn_lefthand | other.syn_righthand;
            }
            continue;
        }
        if overlap(left_rels, other.syn_righthand)
            && overlap(clause_relids, other.syn_righthand)
            && (matches!(jointype, JoinType::Semi | JoinType::Anti) || !overlap(strict_relids, other.min_righthand))
        {
            min_lefthand |= other.syn_lefthand | other.syn_righthand;
        }
        if overlap(right_rels, other.syn_righthand)
            && (overlap(clause_relids, other.syn_righthand)
                || !overlap(clause_relids, other.min_lefthand)
                || matches!(jointype, JoinType::Semi | JoinType::Anti)
                || matches!(other.jointype, JoinType::Semi | JoinType::Anti)
                || !other.lhs_strict
                || other.delay_upper_joins)
        {
            min_righthand |= other.syn_lefthand | other.syn_righthand;
        }
    }
    sjinfo.min_lefthand = if min_lefthand == 0 { left_rels } else { min_lefthand };
    sjinfo.min_righthand = if min_righthand == 0 { right_rels } else { min_righthand };
    sjinfo
}

/// distribute_qual_to_rels makes a RestrictInfo of a clause and attaches it to the relations it restricts, or to the
/// join that must evaluate it, given its syntactic scope, the scope of the outer join whose ON clause it is, and that
/// join's non-nullable side, as Postgres' distribute_qual_to_rels does. The clause's equivalences are not yet fed to
/// equivalence classes, which equivclass.c will do.
fn distribute_qual_to_rels(
    root: &mut PlannerInfo<'_, '_>,
    clause: &Expr,
    below_outer_join: bool,
    qualscope: Relids,
    ojscope: Relids,
    outerjoin_nonnullable: Relids,
) {
    let mut relids = pull_varnos(clause);
    let mut pseudoconstant = false;
    if relids == 0 {
        if ojscope != 0 {
            relids = ojscope;
        } else {
            relids = qualscope;
            if !contain_volatile_functions(clause) {
                pseudoconstant = true;
                if !below_outer_join {
                    relids = root.all_baserels;
                }
            }
        }
    }
    let is_pushed_down = !overlap(relids, outerjoin_nonnullable);
    if is_pushed_down {
        let outerjoin_delayed = check_outerjoin_delay(root, &mut relids);
        if outerjoin_delayed && check_redundant_nullability_qual(root, clause) {
            return;
        }
    } else {
        relids = ojscope;
    }
    let restrictinfo = make_restrictinfo(clause.clone(), is_pushed_down, pseudoconstant, relids);
    distribute_restrictinfo_to_rels(root, Rc::new(restrictinfo));
}

/// check_redundant_nullability_qual reports whether a clause only tests that a Var is NULL where an anti join already
/// makes it so, as Postgres' function of the same name does.
fn check_redundant_nullability_qual(root: &PlannerInfo<'_, '_>, clause: &Expr) -> bool {
    let (Expr::IsNull(inner, false) | Expr::BoolTest(inner, None, false)) = clause else { return false };
    let Expr::Column(var) = **inner else { return false };
    let forced_null_rel = super::nodes::var_parts(var).0;
    root.join_info_list.iter().any(|sjinfo| {
        sjinfo.jointype == JoinType::Anti && overlap(super::nodes::singleton(forced_null_rel), sjinfo.syn_righthand)
    })
}

/// check_outerjoin_delay widens the relations a pushed-down clause needs by the outer joins that it must wait for,
/// returning whether any delays it, as Postgres' check_outerjoin_delay does.
fn check_outerjoin_delay(root: &mut PlannerInfo<'_, '_>, relids: &mut Relids) -> bool {
    if root.join_info_list.is_empty() {
        return false;
    }
    let (mut widened, mut delayed) = (*relids, false);
    loop {
        let mut found_some = false;
        for sjinfo in root.join_info_list.iter_mut() {
            if overlap(widened, sjinfo.min_righthand)
                || (sjinfo.jointype == JoinType::Full && overlap(widened, sjinfo.min_lefthand))
            {
                if !is_subset(sjinfo.min_lefthand, widened) || !is_subset(sjinfo.min_righthand, widened) {
                    widened |= sjinfo.min_lefthand | sjinfo.min_righthand;
                    delayed = true;
                    found_some = true;
                }
                if sjinfo.jointype != JoinType::Full && overlap(widened, sjinfo.min_lefthand) {
                    sjinfo.delay_upper_joins = true;
                }
            }
        }
        if !found_some {
            break;
        }
    }
    *relids = widened;
    delayed
}

/// distribute_restrictinfo_to_rels attaches a clause to the base relation it restricts, or as a join clause to each
/// relation it reads, as Postgres' distribute_restrictinfo_to_rels does.
fn distribute_restrictinfo_to_rels(root: &mut PlannerInfo<'_, '_>, restrictinfo: Rc<RestrictInfo>) {
    let relids = restrictinfo.required_relids;
    if relids.count_ones() == 1 {
        root.rels[relids.trailing_zeros() as usize].baserestrictinfo.push(restrictinfo);
    } else {
        for varno in super::nodes::members(relids) {
            root.rels[varno].joininfo.push(restrictinfo.clone());
        }
    }
}
