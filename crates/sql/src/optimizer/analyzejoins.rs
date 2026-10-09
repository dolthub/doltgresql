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

//! Postgres' optimizer/plan/analyzejoins.c: removing left joins that cannot change the query's result, and proving
//! that a join's inner relation matches each outer row at most once.

use std::rc::Rc;

use super::PlannerInfo;
use super::clauses::pull_varnos;
use super::initsplan::{JoinList, distribute_restrictinfo_to_rels};
use super::joinpath::clause_sides_match_join;
use super::nodes::{JoinType, Relids, RestrictInfo, SpecialJoinInfo, is_subset, members, singleton, var};
use super::restrictinfo::is_pushed_down;
use crate::expr::{CmpOp, Expr};

/// remove_useless_joins removes each left join whose inner relation's columns nothing above reads and whose
/// clauses find at most one inner row for each outer row, returning the joinlist without the removed relations, as
/// Postgres' function of the same name does.
pub fn remove_useless_joins(root: &mut PlannerInfo<'_, '_>, mut joinlist: Vec<JoinList>) -> Vec<JoinList> {
    while let Some(i) = (0..root.join_info_list.len()).find(|&i| join_is_removable(root, &root.join_info_list[i])) {
        let sjinfo = root.join_info_list.remove(i);
        let innerrelid = sjinfo.min_righthand.trailing_zeros() as usize;
        remove_rel_from_query(root, innerrelid, sjinfo.min_lefthand | sjinfo.min_righthand);
        joinlist = remove_rel_from_joinlist(joinlist, innerrelid);
    }
    joinlist
}

/// join_is_removable reports whether a left join to a single base relation can be removed, as Postgres' function of
/// the same name does.
fn join_is_removable(root: &PlannerInfo<'_, '_>, sjinfo: &SpecialJoinInfo) -> bool {
    if sjinfo.jointype != JoinType::Left || sjinfo.delay_upper_joins || sjinfo.min_righthand.count_ones() != 1 {
        return false;
    }
    let innerrelid = sjinfo.min_righthand.trailing_zeros() as usize;
    let innerrel = &root.rels[innerrelid];
    if !rel_supports_distinctness(root, innerrelid) {
        return false;
    }
    let joinrelids = sjinfo.min_lefthand | sjinfo.min_righthand;
    if innerrel.attr_needed.iter().any(|&needed| !is_subset(needed, joinrelids)) {
        return false;
    }
    let mut clause_list = Vec::new();
    for rinfo in &innerrel.joininfo {
        if is_pushed_down(rinfo, joinrelids) {
            if rinfo.clause_relids & singleton(innerrelid) != 0 {
                return false;
            }
            continue;
        }
        if rinfo.can_join && rinfo.hashjoinable && clause_sides_match_join(rinfo, sjinfo.min_lefthand, innerrel.relids)
        {
            clause_list.push(rinfo.clone());
        }
    }
    rel_is_distinct_for(root, innerrelid, &clause_list)
}

/// remove_rel_from_query removes a base relation from the planner's state, re-attaching the clauses of other joins
/// that waited for it, as Postgres' function of the same name does.
fn remove_rel_from_query(root: &mut PlannerInfo<'_, '_>, relid: usize, joinrelids: Relids) {
    let removed = !singleton(relid);
    root.all_baserels &= removed;
    for rel in root.rels.iter_mut() {
        for needed in rel.attr_needed.iter_mut() {
            *needed &= removed;
        }
    }
    for sjinfo in root.join_info_list.iter_mut() {
        sjinfo.min_lefthand &= removed;
        sjinfo.min_righthand &= removed;
        sjinfo.syn_lefthand &= removed;
        sjinfo.syn_righthand &= removed;
    }
    for rinfo in root.rels[relid].joininfo.clone() {
        for other in members(rinfo.required_relids) {
            root.rels[other].joininfo.retain(|r| !Rc::ptr_eq(r, &rinfo));
        }
        if is_pushed_down(&rinfo, joinrelids) {
            let required_relids = rinfo.required_relids & removed;
            distribute_restrictinfo_to_rels(root, Rc::new(RestrictInfo { required_relids, ..(*rinfo).clone() }));
        }
    }
}

/// remove_rel_from_joinlist returns a joinlist without a relation, as Postgres' function of the same name does.
fn remove_rel_from_joinlist(joinlist: Vec<JoinList>, relid: usize) -> Vec<JoinList> {
    joinlist
        .into_iter()
        .filter_map(|item| match item {
            JoinList::Rel(varno) if varno == relid => None,
            JoinList::List(list) => {
                let list = remove_rel_from_joinlist(list, relid);
                (!list.is_empty()).then_some(JoinList::List(list))
            }
            other => Some(other),
        })
        .collect()
}

/// rel_supports_distinctness reports whether a base relation could be proven to have at most one row for some
/// values of its columns: a table with a primary key or a unique index, as Postgres' function of the same name
/// checks.
fn rel_supports_distinctness(root: &PlannerInfo<'_, '_>, relid: usize) -> bool {
    root.parse
        .rte(relid)
        .table()
        .is_some_and(|table| !table.keyless() || table.indexes.iter().any(|i| i.unique && i.predicate.is_empty()))
}

/// rel_is_distinct_for reports whether a base relation has at most one row for each set of values of the inner sides
/// of a list of join equalities: one of its unique indexes has each of its columns fixed by such an equality or by an
/// equality with a value in the relation's own restrictions, as Postgres' rel_is_distinct_for proves with
/// relation_has_unique_index_for.
fn rel_is_distinct_for(root: &PlannerInfo<'_, '_>, relid: usize, clause_list: &[Rc<RestrictInfo>]) -> bool {
    let Some(table) = root.parse.rte(relid).table() else { return false };
    let rel = &root.rels[relid];
    let inner_side = |r: &RestrictInfo| match &r.clause {
        Expr::Compare(_, l, _) if is_subset(r.left_relids, rel.relids) => Some((**l).clone()),
        Expr::Compare(_, _, rhs) => Some((**rhs).clone()),
        _ => None,
    };
    let mut fixed: Vec<Expr> = clause_list.iter().filter_map(|r| inner_side(r)).collect();
    for r in &rel.baserestrictinfo {
        if let Expr::Compare(CmpOp::Eq, l, rhs) = &r.clause {
            match (pull_varnos(l) == 0, pull_varnos(rhs) == 0) {
                (false, true) => fixed.push((**l).clone()),
                (true, false) => fixed.push((**rhs).clone()),
                _ => {}
            }
        }
    }
    let covers = |columns: &[usize]| columns.iter().all(|&c| fixed.contains(&Expr::Column(var(relid, c))));
    (!table.keyless() && covers(&table.key_columns))
        || table.indexes.iter().any(|i| i.unique && i.predicate.is_empty() && covers(&i.columns))
}

/// innerrel_is_unique reports whether each outer row matches at most one row of the inner relation by a join's
/// equalities, as Postgres' function of the same name proves it for a base relation.
pub fn innerrel_is_unique(
    root: &PlannerInfo<'_, '_>,
    joinrelids: Relids,
    outerrel: usize,
    innerrel: usize,
    jointype: JoinType,
    restrictlist: &[Rc<RestrictInfo>],
) -> bool {
    let inner = &root.rels[innerrel];
    if inner.relid == 0 || !rel_supports_distinctness(root, inner.relid) {
        return false;
    }
    let outer_relids = root.rels[outerrel].relids;
    let clause_list: Vec<Rc<RestrictInfo>> = restrictlist
        .iter()
        .filter(|r| !(jointype.is_outer() && is_pushed_down(r, joinrelids)))
        .filter(|r| r.can_join && r.hashjoinable && clause_sides_match_join(r, outer_relids, inner.relids))
        .cloned()
        .collect();
    rel_is_distinct_for(root, inner.relid, &clause_list)
}
