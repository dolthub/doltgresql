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

//! Postgres' optimizer/path/joinrels.c, with have_relevant_joinclause from joininfo.c: finding the join relations of
//! each level of the join search, and which joins of two relations are legal.

use super::PlannerInfo;
use super::costsize::inner_sjinfo;
use super::joinpath::add_paths_to_joinrel;
use super::nodes::{JoinType, Relids, SpecialJoinInfo, is_subset, overlap};
use super::relnode::build_join_rel;

/// join_search_one_level builds the join relations of a level of the join search from those of the lower levels, as
/// Postgres' function of the same name does: each relation of the level below joins each base relation that a join
/// clause or join order restriction connects it to, or every one when none does, and relations of two lower levels
/// join when something connects them.
pub fn join_search_one_level(root: &mut PlannerInfo<'_, '_>, level: usize) {
    root.join_rel_level.push(Vec::new());
    let previous = root.join_rel_level[level - 1].clone();
    let initial = root.join_rel_level[1].clone();
    for (i, &old_rel) in previous.iter().enumerate() {
        if !root.rels[old_rel].joininfo.is_empty() || has_join_restriction(root, old_rel) {
            let other_rels = if level == 2 { &previous[i + 1..] } else { &initial[..] };
            make_rels_by_clause_joins(root, old_rel, other_rels);
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
            if root.rels[old_rel].joininfo.is_empty() && !has_join_restriction(root, old_rel) {
                continue;
            }
            let others = match k == other_level {
                true => olds[i + 1..].to_vec(),
                false => root.join_rel_level[other_level].clone(),
            };
            for new_rel in others {
                if !overlap(root.rels[old_rel].relids, root.rels[new_rel].relids)
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
    }
}

/// make_rels_by_clause_joins joins a relation to each of the others that a join clause or join order restriction
/// connects it to.
fn make_rels_by_clause_joins(root: &mut PlannerInfo<'_, '_>, old_rel: usize, other_rels: &[usize]) {
    for &other_rel in other_rels {
        if !overlap(root.rels[old_rel].relids, root.rels[other_rel].relids)
            && (have_relevant_joinclause(root, old_rel, other_rel)
                || have_join_order_restriction(root, old_rel, other_rel))
        {
            make_join_rel(root, old_rel, other_rel);
        }
    }
}

/// make_rels_by_clauseless_joins joins a relation to each of the others that it does not overlap.
fn make_rels_by_clauseless_joins(root: &mut PlannerInfo<'_, '_>, old_rel: usize, other_rels: &[usize]) {
    for &other_rel in other_rels {
        if !overlap(root.rels[other_rel].relids, root.rels[old_rel].relids) {
            make_join_rel(root, old_rel, other_rel);
        }
    }
}

/// join_is_legal reports whether two relations may be joined without breaking an outer, semi, or anti join, with
/// the special join that their join forms and whether its sides are the other way around, as Postgres' function of
/// the same name does. A semi join's inner side is never made unique so as to join it as an inner join.
fn join_is_legal(
    root: &PlannerInfo<'_, '_>,
    rel1: Relids,
    rel2: Relids,
    joinrelids: Relids,
) -> Option<(Option<SpecialJoinInfo>, bool)> {
    let mut match_sjinfo: Option<&SpecialJoinInfo> = None;
    let mut reversed = false;
    let mut must_be_leftjoin = false;
    for sjinfo in &root.join_info_list {
        if !overlap(sjinfo.min_righthand, joinrelids)
            || is_subset(joinrelids, sjinfo.min_righthand)
            || (is_subset(sjinfo.min_lefthand, rel1) && is_subset(sjinfo.min_righthand, rel1))
            || (is_subset(sjinfo.min_lefthand, rel2) && is_subset(sjinfo.min_righthand, rel2))
        {
            continue;
        }
        if sjinfo.jointype == JoinType::Semi
            && ((is_subset(sjinfo.syn_righthand, rel1) && sjinfo.syn_righthand != rel1)
                || (is_subset(sjinfo.syn_righthand, rel2) && sjinfo.syn_righthand != rel2))
        {
            continue;
        }
        if is_subset(sjinfo.min_lefthand, rel1) && is_subset(sjinfo.min_righthand, rel2) {
            if match_sjinfo.is_some() {
                return None;
            }
            match_sjinfo = Some(sjinfo);
            reversed = false;
        } else if is_subset(sjinfo.min_lefthand, rel2) && is_subset(sjinfo.min_righthand, rel1) {
            if match_sjinfo.is_some() {
                return None;
            }
            match_sjinfo = Some(sjinfo);
            reversed = true;
        } else {
            if overlap(rel1, sjinfo.min_righthand) && overlap(rel2, sjinfo.min_righthand) {
                continue;
            }
            if sjinfo.jointype != JoinType::Left || overlap(joinrelids, sjinfo.min_lefthand) {
                return None;
            }
            must_be_leftjoin = true;
        }
    }
    if must_be_leftjoin && !match_sjinfo.is_some_and(|s| s.jointype == JoinType::Left && s.lhs_strict) {
        return None;
    }
    Some((match_sjinfo.cloned(), reversed))
}

/// make_join_rel joins two relations when that is legal, building their join relation and adding the paths of the
/// join, as Postgres' function of the same name does.
fn make_join_rel(root: &mut PlannerInfo<'_, '_>, rel1: usize, rel2: usize) {
    let (relids1, relids2) = (root.rels[rel1].relids, root.rels[rel2].relids);
    let joinrelids = relids1 | relids2;
    let Some((sjinfo, reversed)) = join_is_legal(root, relids1, relids2, joinrelids) else { return };
    let (rel1, rel2) = if reversed { (rel2, rel1) } else { (rel1, rel2) };
    let sjinfo = sjinfo.unwrap_or_else(|| inner_sjinfo(root.rels[rel1].relids, root.rels[rel2].relids));
    let (joinrel, restrictlist) = build_join_rel(root, joinrelids, rel1, rel2, &sjinfo);
    populate_joinrel_with_paths(root, rel1, rel2, joinrel, &sjinfo, &restrictlist);
}

/// populate_joinrel_with_paths adds the paths of a join of two relations to their join relation, with each relation
/// as the outer side where the join type allows it, as Postgres' function of the same name does.
fn populate_joinrel_with_paths(
    root: &mut PlannerInfo<'_, '_>,
    rel1: usize,
    rel2: usize,
    joinrel: usize,
    sjinfo: &SpecialJoinInfo,
    restrictlist: &[std::rc::Rc<super::nodes::RestrictInfo>],
) {
    match sjinfo.jointype {
        JoinType::Inner => {
            add_paths_to_joinrel(root, joinrel, rel1, rel2, JoinType::Inner, sjinfo, restrictlist);
            add_paths_to_joinrel(root, joinrel, rel2, rel1, JoinType::Inner, sjinfo, restrictlist);
        }
        JoinType::Left => {
            add_paths_to_joinrel(root, joinrel, rel1, rel2, JoinType::Left, sjinfo, restrictlist);
            add_paths_to_joinrel(root, joinrel, rel2, rel1, JoinType::Right, sjinfo, restrictlist);
        }
        JoinType::Full => {
            add_paths_to_joinrel(root, joinrel, rel1, rel2, JoinType::Full, sjinfo, restrictlist);
            add_paths_to_joinrel(root, joinrel, rel2, rel1, JoinType::Full, sjinfo, restrictlist);
        }
        JoinType::Semi => {
            if is_subset(sjinfo.min_lefthand, root.rels[rel1].relids)
                && is_subset(sjinfo.min_righthand, root.rels[rel2].relids)
            {
                add_paths_to_joinrel(root, joinrel, rel1, rel2, JoinType::Semi, sjinfo, restrictlist);
            }
        }
        JoinType::Anti => add_paths_to_joinrel(root, joinrel, rel1, rel2, JoinType::Anti, sjinfo, restrictlist),
        JoinType::Right => unreachable!("reduce_outer_joins turns right joins into left joins"),
    }
}

/// have_join_order_restriction reports whether an outer, semi, or anti join requires two relations to be joined
/// before either joins anything else, as Postgres' function of the same name does.
fn have_join_order_restriction(root: &PlannerInfo<'_, '_>, rel1: usize, rel2: usize) -> bool {
    let (relids1, relids2) = (root.rels[rel1].relids, root.rels[rel2].relids);
    let restricted = root.join_info_list.iter().filter(|s| s.jointype != JoinType::Full).any(|s| {
        (is_subset(s.min_lefthand, relids1) && is_subset(s.min_righthand, relids2))
            || (is_subset(s.min_lefthand, relids2) && is_subset(s.min_righthand, relids1))
            || (overlap(s.min_righthand, relids1) && overlap(s.min_righthand, relids2))
            || (overlap(s.min_lefthand, relids1) && overlap(s.min_lefthand, relids2))
    });
    restricted && !has_legal_joinclause(root, rel1) && !has_legal_joinclause(root, rel2)
}

/// has_join_restriction reports whether an outer, semi, or anti join restricts what a relation may join, as
/// Postgres' function of the same name does.
fn has_join_restriction(root: &PlannerInfo<'_, '_>, rel: usize) -> bool {
    let relids = root.rels[rel].relids;
    root.join_info_list.iter().filter(|s| s.jointype != JoinType::Full).any(|s| {
        !(is_subset(s.min_lefthand, relids) && is_subset(s.min_righthand, relids))
            && (overlap(s.min_lefthand, relids) || overlap(s.min_righthand, relids))
    })
}

/// has_legal_joinclause reports whether a relation has a join clause with one of the relations the join search
/// started from that it may legally join, as Postgres' function of the same name does.
fn has_legal_joinclause(root: &PlannerInfo<'_, '_>, rel: usize) -> bool {
    let relids = root.rels[rel].relids;
    root.initial_rels.iter().any(|&rel2| {
        let relids2 = root.rels[rel2].relids;
        !overlap(relids, relids2)
            && have_relevant_joinclause(root, rel, rel2)
            && join_is_legal(root, relids, relids2, relids | relids2).is_some()
    })
}

/// have_relevant_joinclause reports whether a join clause reads both relations, as Postgres' function of the same
/// name does.
fn have_relevant_joinclause(root: &PlannerInfo<'_, '_>, rel1: usize, rel2: usize) -> bool {
    let (rel1, rel2) = (&root.rels[rel1], &root.rels[rel2]);
    let (joininfo, other_relids) = match rel1.joininfo.len() <= rel2.joininfo.len() {
        true => (&rel1.joininfo, rel2.relids),
        false => (&rel2.joininfo, rel1.relids),
    };
    joininfo.iter().any(|r| overlap(other_relids, r.required_relids))
}
