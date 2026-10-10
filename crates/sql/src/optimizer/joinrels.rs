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

//! Postgres' optimizer/path/joinrels.c: finding the join relations of each level of the join search, and which
//! joins of two relations are legal.

use std::rc::Rc;

use super::PlannerInfo;
use super::joininfo::have_relevant_joinclause;
use super::joinpath::add_paths_to_joinrel;
use super::nodes::{JoinType, Path, PathKind, Relids, RinfoId, SjId, SpecialJoinInfo};
use super::pathnode::{add_path, set_cheapest};
use super::relnode::{build_join_rel, min_join_parameterization};
use super::restrictinfo::rinfo_is_pushed_down;
use crate::expr::Expr;
use crate::types::Value;

/// join_search_one_level builds the join relations of a level of the join search from those of the lower levels, as
/// Postgres' function of the same name does: each relation of the level below joins each base relation that a join
/// clause or join order restriction connects it to, or every one when none does, and relations of two lower levels
/// join when something connects them.
pub fn join_search_one_level(root: &mut PlannerInfo<'_, '_>, level: usize) {
    root.join_rel_level.push(Vec::new());
    root.join_cur_level = level;
    let previous = root.join_rel_level[level - 1].clone();
    let initial = root.join_rel_level[1].clone();
    for (i, &old_rel) in previous.iter().enumerate() {
        let rel = &root.rels[old_rel];
        if !rel.joininfo.is_empty() || rel.has_eclass_joins || has_join_restriction(root, old_rel) {
            let first_rel = if level == 2 { i + 1 } else { 0 };
            make_rels_by_clause_joins(root, old_rel, &initial[first_rel..]);
        } else {
            make_rels_by_clauseless_joins(root, old_rel, &initial);
        }
    }
    for k in 2.. {
        let other_level = level - k;
        if k > other_level {
            break;
        }
        let olds = root.join_rel_level[k].clone();
        for (i, &old_rel) in olds.iter().enumerate() {
            let rel = &root.rels[old_rel];
            if rel.joininfo.is_empty() && !rel.has_eclass_joins && !has_join_restriction(root, old_rel) {
                continue;
            }
            let first_rel = if k == other_level { i + 1 } else { 0 };
            let new_rels = root.join_rel_level[other_level][first_rel..].to_vec();
            for new_rel in new_rels {
                if !root.rels[old_rel].relids.overlap(&root.rels[new_rel].relids)
                    && (have_relevant_joinclause(root, old_rel, new_rel)
                        || have_join_order_restriction(root, old_rel, new_rel))
                {
                    make_join_rel(root, old_rel, new_rel);
                }
            }
        }
    }
    if root.join_rel_level[level].is_empty() {
        for old_rel in previous {
            make_rels_by_clauseless_joins(root, old_rel, &initial);
        }
        assert!(
            !root.join_rel_level[level].is_empty() || !root.join_info_list.is_empty() || root.has_lateral_rtes,
            "failed to build any {level}-way joins"
        );
    }
}

/// make_rels_by_clause_joins joins a relation to each of the others that a join clause or join order restriction
/// connects it to, as Postgres' function of the same name does.
fn make_rels_by_clause_joins(root: &mut PlannerInfo<'_, '_>, old_rel: usize, other_rels: &[usize]) {
    for &other_rel in other_rels {
        if !root.rels[old_rel].relids.overlap(&root.rels[other_rel].relids)
            && (have_relevant_joinclause(root, old_rel, other_rel)
                || have_join_order_restriction(root, old_rel, other_rel))
        {
            make_join_rel(root, old_rel, other_rel);
        }
    }
}

/// make_rels_by_clauseless_joins joins a relation to each of the others that it does not overlap, as Postgres'
/// function of the same name does.
fn make_rels_by_clauseless_joins(root: &mut PlannerInfo<'_, '_>, old_rel: usize, other_rels: &[usize]) {
    for &other_rel in other_rels {
        if !root.rels[other_rel].relids.overlap(&root.rels[old_rel].relids) {
            make_join_rel(root, old_rel, other_rel);
        }
    }
}

/// join_is_legal reports whether two relations may be joined without breaking an outer, semi, or anti join or a
/// lateral reference, with the special join that their join forms and whether its sides are the other way around,
/// as Postgres' function of the same name does. A semi join's inner side is never made unique so as to join it as an
/// inner join, which unique paths will allow.
fn join_is_legal(
    root: &PlannerInfo<'_, '_>,
    rel1: usize,
    rel2: usize,
    joinrelids: &Relids,
) -> Option<(Option<SjId>, bool)> {
    let (relids1, relids2) = (&root.rels[rel1].relids, &root.rels[rel2].relids);
    let mut match_sjinfo: Option<SjId> = None;
    let mut reversed = false;
    let mut must_be_leftjoin = false;
    for &sj in &root.join_info_list {
        let sjinfo = &root.sjinfos[sj];
        if !sjinfo.min_righthand.overlap(joinrelids)
            || joinrelids.is_subset(&sjinfo.min_righthand)
            || (sjinfo.min_lefthand.is_subset(relids1) && sjinfo.min_righthand.is_subset(relids1))
            || (sjinfo.min_lefthand.is_subset(relids2) && sjinfo.min_righthand.is_subset(relids2))
        {
            continue;
        }
        if sjinfo.jointype == JoinType::Semi
            && ((sjinfo.syn_righthand.is_subset(relids1) && sjinfo.syn_righthand != *relids1)
                || (sjinfo.syn_righthand.is_subset(relids2) && sjinfo.syn_righthand != *relids2))
        {
            continue;
        }
        if sjinfo.min_lefthand.is_subset(relids1) && sjinfo.min_righthand.is_subset(relids2) {
            if match_sjinfo.is_some() {
                return None;
            }
            (match_sjinfo, reversed) = (Some(sj), false);
        } else if sjinfo.min_lefthand.is_subset(relids2) && sjinfo.min_righthand.is_subset(relids1) {
            if match_sjinfo.is_some() {
                return None;
            }
            (match_sjinfo, reversed) = (Some(sj), true);
        } else {
            if relids1.overlap(&sjinfo.min_righthand) && relids2.overlap(&sjinfo.min_righthand) {
                continue;
            }
            if sjinfo.jointype != JoinType::Left || joinrelids.overlap(&sjinfo.min_lefthand) {
                return None;
            }
            must_be_leftjoin = true;
        }
    }
    if must_be_leftjoin
        && !match_sjinfo.is_some_and(|s| root.sjinfos[s].jointype == JoinType::Left && root.sjinfos[s].lhs_strict)
    {
        return None;
    }
    if root.has_lateral_rtes {
        let (r1, r2) = (&root.rels[rel1], &root.rels[rel2]);
        let lateral_fwd = r1.relids.overlap(&r2.lateral_relids);
        let lateral_rev = r2.relids.overlap(&r1.lateral_relids);
        if lateral_fwd && lateral_rev {
            return None;
        }
        let full = match_sjinfo.is_some_and(|s| root.sjinfos[s].jointype == JoinType::Full);
        if lateral_fwd
            && ((match_sjinfo.is_some() && (reversed || full)) || !r1.relids.overlap(&r2.direct_lateral_relids))
        {
            return None;
        }
        if lateral_rev
            && ((match_sjinfo.is_some() && (!reversed || full)) || !r2.relids.overlap(&r1.direct_lateral_relids))
        {
            return None;
        }
        let join_lateral_rels = min_join_parameterization(root, joinrelids, rel1, rel2);
        if !join_lateral_rels.is_empty() {
            let mut join_plus_rhs = joinrelids.clone();
            loop {
                let mut more = false;
                for &sj in &root.join_info_list {
                    let sjinfo = &root.sjinfos[sj];
                    if sjinfo.jointype != JoinType::Full
                        && sjinfo.min_lefthand.overlap(&join_plus_rhs)
                        && !sjinfo.min_righthand.is_subset(&join_plus_rhs)
                    {
                        join_plus_rhs.add_members(&sjinfo.min_righthand);
                        more = true;
                    }
                }
                if !more {
                    break;
                }
            }
            if join_plus_rhs.overlap(&join_lateral_rels) {
                return None;
            }
            if join_lateral_rels.overlap(&root.outer_join_rels) {
                for &sj in &root.join_info_list {
                    let sjinfo = &root.sjinfos[sj];
                    if join_lateral_rels.is_member(sjinfo.ojrelid)
                        && (join_plus_rhs.overlap(&sjinfo.min_lefthand) || join_plus_rhs.overlap(&sjinfo.min_righthand))
                    {
                        return None;
                    }
                }
            }
        }
    }
    Some((match_sjinfo, reversed))
}

/// init_dummy_sjinfo returns the SpecialJoinInfo of an inner join of two sets of relations, as Postgres' function of
/// the same name does.
pub fn init_dummy_sjinfo(left_relids: &Relids, right_relids: &Relids) -> SpecialJoinInfo {
    SpecialJoinInfo {
        min_lefthand: left_relids.clone(),
        min_righthand: right_relids.clone(),
        syn_lefthand: left_relids.clone(),
        syn_righthand: right_relids.clone(),
        jointype: JoinType::Inner,
        ojrelid: 0,
        commute_above_l: Relids::new(),
        commute_above_r: Relids::new(),
        commute_below_l: Relids::new(),
        commute_below_r: Relids::new(),
        lhs_strict: false,
        semi_can_btree: false,
        semi_can_hash: false,
        semi_rhs_exprs: Vec::new(),
    }
}

/// make_join_rel joins two relations when that is legal, building their join relation and adding the paths of the
/// join, and returns the join relation, as Postgres' function of the same name does.
pub fn make_join_rel(root: &mut PlannerInfo<'_, '_>, rel1: usize, rel2: usize) -> Option<usize> {
    let joinrelids = root.rels[rel1].relids.union(&root.rels[rel2].relids);
    let (sjinfo, reversed) = join_is_legal(root, rel1, rel2, &joinrelids)?;
    let mut pushed_down_joins = Vec::new();
    let joinrelids = add_outer_joins_to_relids(root, joinrelids, sjinfo, Some(&mut pushed_down_joins));
    let (rel1, rel2) = if reversed { (rel2, rel1) } else { (rel1, rel2) };
    let sjinfo = match sjinfo {
        Some(sj) => root.sjinfos[sj].clone(),
        None => init_dummy_sjinfo(&root.rels[rel1].relids, &root.rels[rel2].relids),
    };
    let pushed_down: Vec<SpecialJoinInfo> = pushed_down_joins.iter().map(|&s| root.sjinfos[s].clone()).collect();
    let (joinrel, restrictlist) = build_join_rel(root, joinrelids, rel1, rel2, &sjinfo, &pushed_down);
    if is_dummy_rel(root, joinrel) {
        return Some(joinrel);
    }
    populate_joinrel_with_paths(root, rel1, rel2, joinrel, &sjinfo, &restrictlist);
    Some(joinrel)
}

/// add_outer_joins_to_relids adds the outer joins that a join of a set of relations completes to the set: the join's
/// own, and those of the left joins above it that commute with it and can be completed now, which are pushed down,
/// as Postgres' function of the same name does.
pub fn add_outer_joins_to_relids(
    root: &PlannerInfo<'_, '_>,
    mut input_relids: Relids,
    sjinfo: Option<SjId>,
    mut pushed_down_joins: Option<&mut Vec<SjId>>,
) -> Relids {
    let Some(sj) = sjinfo else { return input_relids };
    let sjinfo = &root.sjinfos[sj];
    if sjinfo.ojrelid == 0 {
        return input_relids;
    }
    if sjinfo.jointype != JoinType::Left {
        return input_relids.with_member(sjinfo.ojrelid);
    }
    if !sjinfo.commute_below_l.is_subset(&input_relids) {
        return input_relids;
    }
    input_relids.add_member(sjinfo.ojrelid);
    if !sjinfo.commute_above_l.is_empty() {
        let mut commute_above_rels = sjinfo.commute_above_l.clone();
        for &other in &root.join_info_list {
            let othersj = &root.sjinfos[other];
            if other == sj || othersj.ojrelid == 0 || othersj.jointype != JoinType::Left {
                continue;
            }
            if !commute_above_rels.is_member(othersj.ojrelid) {
                continue;
            }
            if !input_relids.is_member(othersj.ojrelid)
                && othersj.min_lefthand.is_subset(&input_relids)
                && othersj.min_righthand.is_subset(&input_relids)
                && othersj.commute_below_l.is_subset(&input_relids)
            {
                input_relids.add_member(othersj.ojrelid);
                if let Some(pushed) = pushed_down_joins.as_deref_mut() {
                    pushed.push(other);
                }
                commute_above_rels.add_members(&othersj.commute_above_l);
            }
        }
    }
    input_relids
}

/// populate_joinrel_with_paths adds the paths of a join of two relations to their join relation, with each relation
/// as the outer side where the join type allows it, or marks the join relation empty when a side is empty or a
/// clause is constantly false, as Postgres' function of the same name does.
fn populate_joinrel_with_paths(
    root: &mut PlannerInfo<'_, '_>,
    rel1: usize,
    rel2: usize,
    joinrel: usize,
    sjinfo: &SpecialJoinInfo,
    restrictlist: &[RinfoId],
) {
    let constant_false = |root: &PlannerInfo<'_, '_>, only_pushed_down| {
        restriction_is_constant_false(root, restrictlist, joinrel, only_pushed_down)
    };
    match sjinfo.jointype {
        JoinType::Inner => {
            if is_dummy_rel(root, rel1) || is_dummy_rel(root, rel2) || constant_false(root, false) {
                mark_dummy_rel(root, joinrel);
                return;
            }
            add_paths_to_joinrel(root, joinrel, rel1, rel2, JoinType::Inner, sjinfo, restrictlist);
            add_paths_to_joinrel(root, joinrel, rel2, rel1, JoinType::Inner, sjinfo, restrictlist);
        }
        JoinType::Left | JoinType::Anti => {
            if is_dummy_rel(root, rel1) || constant_false(root, true) {
                mark_dummy_rel(root, joinrel);
                return;
            }
            if constant_false(root, false) && root.rels[rel2].relids.is_subset(&sjinfo.syn_righthand) {
                mark_dummy_rel(root, rel2);
            }
            add_paths_to_joinrel(root, joinrel, rel1, rel2, sjinfo.jointype, sjinfo, restrictlist);
            let swapped = if sjinfo.jointype == JoinType::Left { JoinType::Right } else { JoinType::RightAnti };
            add_paths_to_joinrel(root, joinrel, rel2, rel1, swapped, sjinfo, restrictlist);
        }
        JoinType::Full => {
            if (is_dummy_rel(root, rel1) && is_dummy_rel(root, rel2)) || constant_false(root, true) {
                mark_dummy_rel(root, joinrel);
                return;
            }
            add_paths_to_joinrel(root, joinrel, rel1, rel2, JoinType::Full, sjinfo, restrictlist);
            add_paths_to_joinrel(root, joinrel, rel2, rel1, JoinType::Full, sjinfo, restrictlist);
        }
        JoinType::Semi => {
            if sjinfo.min_lefthand.is_subset(&root.rels[rel1].relids)
                && sjinfo.min_righthand.is_subset(&root.rels[rel2].relids)
            {
                if is_dummy_rel(root, rel1) || is_dummy_rel(root, rel2) || constant_false(root, false) {
                    mark_dummy_rel(root, joinrel);
                    return;
                }
                add_paths_to_joinrel(root, joinrel, rel1, rel2, JoinType::Semi, sjinfo, restrictlist);
                add_paths_to_joinrel(root, joinrel, rel2, rel1, JoinType::RightSemi, sjinfo, restrictlist);
            }
            if sjinfo.syn_righthand == root.rels[rel2].relids
                && let Some(cheapest) = root.rels[rel2].cheapest_total_path.clone()
                && super::pathnode::create_unique_path(root, rel2, cheapest, sjinfo).is_some()
            {
                if is_dummy_rel(root, rel1) || is_dummy_rel(root, rel2) || constant_false(root, false) {
                    mark_dummy_rel(root, joinrel);
                    return;
                }
                add_paths_to_joinrel(root, joinrel, rel1, rel2, JoinType::UniqueInner, sjinfo, restrictlist);
                add_paths_to_joinrel(root, joinrel, rel2, rel1, JoinType::UniqueOuter, sjinfo, restrictlist);
            }
        }
        JoinType::Right => unreachable!("reduce_outer_joins turns right joins into left joins"),
        JoinType::RightSemi | JoinType::RightAnti | JoinType::UniqueOuter | JoinType::UniqueInner => {
            unreachable!("only join paths are of the join types that swap or unique-ify a semi or anti join")
        }
    }
}

/// have_join_order_restriction reports whether a lateral reference, a PlaceHolderVar, or an outer, semi, or anti
/// join requires two relations to be joined before either joins anything else, as Postgres' function of the same
/// name does.
fn have_join_order_restriction(root: &PlannerInfo<'_, '_>, rel1: usize, rel2: usize) -> bool {
    let (r1, r2) = (&root.rels[rel1], &root.rels[rel2]);
    if r1.relids.overlap(&r2.direct_lateral_relids) || r2.relids.overlap(&r1.direct_lateral_relids) {
        return true;
    }
    if root.placeholder_list.iter().any(|ph| r1.relids.is_subset(&ph.ph_eval_at) && r2.relids.is_subset(&ph.ph_eval_at))
    {
        return true;
    }
    let restricted =
        root.join_info_list.iter().map(|&s| &root.sjinfos[s]).filter(|s| s.jointype != JoinType::Full).any(|s| {
            (s.min_lefthand.is_subset(&r1.relids) && s.min_righthand.is_subset(&r2.relids))
                || (s.min_lefthand.is_subset(&r2.relids) && s.min_righthand.is_subset(&r1.relids))
                || (s.min_righthand.overlap(&r1.relids) && s.min_righthand.overlap(&r2.relids))
                || (s.min_lefthand.overlap(&r1.relids) && s.min_lefthand.overlap(&r2.relids))
        });
    restricted && !has_legal_joinclause(root, rel1) && !has_legal_joinclause(root, rel2)
}

/// has_join_restriction reports whether a lateral reference, a PlaceHolderVar, or an outer, semi, or anti join
/// restricts what a relation may join, as Postgres' function of the same name does.
fn has_join_restriction(root: &PlannerInfo<'_, '_>, rel: usize) -> bool {
    let r = &root.rels[rel];
    if !r.lateral_relids.is_empty() || !r.lateral_referencers.is_empty() {
        return true;
    }
    if root.placeholder_list.iter().any(|ph| r.relids.is_subset(&ph.ph_eval_at) && r.relids != ph.ph_eval_at) {
        return true;
    }
    root.join_info_list.iter().map(|&s| &root.sjinfos[s]).filter(|s| s.jointype != JoinType::Full).any(|s| {
        !(s.min_lefthand.is_subset(&r.relids) && s.min_righthand.is_subset(&r.relids))
            && (s.min_lefthand.overlap(&r.relids) || s.min_righthand.overlap(&r.relids))
    })
}

/// has_legal_joinclause reports whether a relation has a join clause with one of the relations the join search
/// started from that it may legally join, as Postgres' function of the same name does.
fn has_legal_joinclause(root: &PlannerInfo<'_, '_>, rel: usize) -> bool {
    root.initial_rels.iter().any(|&rel2| {
        !root.rels[rel].relids.overlap(&root.rels[rel2].relids)
            && have_relevant_joinclause(root, rel, rel2)
            && join_is_legal(root, rel, rel2, &root.rels[rel].relids.union(&root.rels[rel2].relids)).is_some()
    })
}

/// is_dummy_rel reports whether a relation is known to be empty, as Postgres' function of the same name does.
pub fn is_dummy_rel(root: &PlannerInfo<'_, '_>, rel: usize) -> bool {
    root.rels[rel]
        .pathlist
        .first()
        .is_some_and(|p| matches!(&p.kind, PathKind::Append(subpaths) if subpaths.is_empty()))
}

/// mark_dummy_rel marks a relation as empty, with the one path of no rows, as Postgres' function of the same name
/// does.
pub fn mark_dummy_rel(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    if is_dummy_rel(root, rel) {
        return;
    }
    let r = &mut root.rels[rel];
    r.rows = 0.0;
    r.pathlist.clear();
    let path = Path {
        kind: PathKind::Append(Vec::new()),
        parent: rel,
        relids: r.relids.clone(),
        param: r.lateral_relids.clone(),
        pathkeys: Vec::new(),
        rows: 0.0,
        width: r.reltarget.width,
        disabled_nodes: 0,
        startup_cost: 0.0,
        total_cost: 0.0,
        pathtarget: None,
    };
    add_path(r, Rc::new(path));
    set_cheapest(r);
}

/// restriction_is_constant_false reports whether a join clause, or only one evaluated above the join when asked, is
/// the constant false or NULL, as Postgres' function of the same name does.
fn restriction_is_constant_false(
    root: &PlannerInfo<'_, '_>,
    restrictlist: &[RinfoId],
    joinrel: usize,
    only_pushed_down: bool,
) -> bool {
    restrictlist.iter().map(|&r| &root.rinfos[r]).any(|rinfo| {
        (!only_pushed_down || rinfo_is_pushed_down(rinfo, &root.rels[joinrel].relids))
            && matches!(rinfo.clause, Expr::Const(Value::Null | Value::Bool(false)))
    })
}
