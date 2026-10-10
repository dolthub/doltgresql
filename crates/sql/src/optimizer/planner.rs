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

//! Postgres' optimizer/plan/planner.c: planning a query, from preparing its join tree to the paths of its upper
//! processing: grouping and aggregation, windows, DISTINCT, ORDER BY, and LIMIT.

use std::collections::BTreeSet;
use std::rc::Rc;

use super::PlannerInfo;
use super::nodes::{
    AggStrategy, GroupingSetData, Path, PathTarget, PkId, Query, RelOptInfo, Relids, RollupData, SortGroupClause,
    TargetEntry, UpperRelationKind,
};
use super::pathkeys::{
    PathKeysComparison, compare_pathkeys, make_pathkeys_for_sortclauses_extended, pathkeys_contained_in,
    pathkeys_count_contained_in,
};
use super::pathnode::{add_path, compare_fractional_path_costs, set_cheapest};
use super::tlist;
use crate::expr::Expr;
use crate::types::Value;

/// DEFAULT_LIMIT_FRACTION is the share of a query's rows that a LIMIT or OFFSET that is not a constant reads, as
/// Postgres assumes.
const DEFAULT_LIMIT_FRACTION: f64 = 0.10;

/// QpExtra is what standard_qp_callback needs besides the query, as Postgres' standard_qp_extra holds it: the windows
/// to evaluate, in order, as the positions of their calls in the query's window calls, the group clauses of the
/// first rollup of the query's grouping sets, and the set operation over the query that wants its rows sorted.
struct QpExtra<'s> {
    active_windows: Vec<Vec<usize>>,
    gset_group_clause: Option<Vec<SortGroupClause>>,
    setop: Option<&'s super::nodes::SetOperationStmt>,
}

/// GroupingSetsData is what preprocess_grouping_sets finds of a query's grouping sets, as Postgres'
/// grouping_sets_data holds it: the rollups that sorted aggregations compute, the sets that only hashing could, by
/// their keys' positions among the group clauses, and their estimated groups, whether any rollup can be hashed, and
/// the clause numbers of the group keys that cannot be.
struct GroupingSetsData {
    rollups: Vec<RollupData>,
    hash_sets_idx: Vec<Vec<usize>>,
    d_num_hash_groups: f64,
    any_hashable: bool,
    unhashable_refs: BTreeSet<usize>,
    unsortable_sets: Vec<GroupingSetData>,
}

/// grouping_planner plans a query's join tree, or its set operations, and then its upper processing, adding the paths
/// of its final relation, as Postgres' function of the same name does. A LIMIT's row count becomes the share of rows
/// that the query reads.
pub fn grouping_planner(
    root: &mut PlannerInfo<'_, '_>,
    tuple_fraction: f64,
    setops: Option<&super::nodes::SetOperationStmt>,
) {
    let mut tuple_fraction = tuple_fraction;
    let (mut offset_est, mut count_est) = (0i64, 0i64);
    let mut limit_tuples = -1.0;
    if root.parse.limit_count.is_some() || root.parse.limit_offset.is_some() {
        tuple_fraction = preprocess_limit(root, tuple_fraction, &mut offset_est, &mut count_est);
        if count_est > 0 && offset_est >= 0 {
            limit_tuples = count_est as f64 + offset_est as f64;
        }
    }
    root.tuple_fraction = tuple_fraction;
    let mut gset_data = None;
    if root.parse.grouping_sets.is_some() {
        gset_data = Some(preprocess_grouping_sets(root));
    } else if !root.parse.group_clause.is_empty() {
        root.processed_group_clause = preprocess_groupclause(root, None);
    }
    root.processed_tlist = root.parse.target_list.clone();
    let active_windows = match root.parse.window_funcs.is_empty() {
        true => Vec::new(),
        false => select_active_windows(root),
    };
    let parse = &root.parse;
    root.limit_tuples = match !parse.group_clause.is_empty()
        || parse.grouping_sets.is_some()
        || !parse.distinct_clause.is_empty()
        || parse.has_aggs
        || !parse.window_funcs.is_empty()
        || !parse.target_srfs.is_empty()
        || root.has_having_qual
    {
        true => -1.0,
        false => limit_tuples,
    };
    if root.parse.set_operations.is_some() {
        if !root.parse.sort_clause.is_empty() {
            root.tuple_fraction = 0.0;
        }
        let mut current_rel = super::prepunion::plan_set_operations(root);
        root.processed_tlist =
            postprocess_setop_tlist(std::mem::take(&mut root.processed_tlist), &root.parse.target_list);
        let cheapest = root.rels[current_rel].cheapest_total_path.clone().expect("a path of a set operation");
        let final_target = super::pathnode::path_target(root, &cheapest);
        let (sort_clause, tlist) = (root.parse.sort_clause.clone(), root.processed_tlist.clone());
        root.sort_pathkeys = make_pathkeys_for_sortclauses(root, &sort_clause, &tlist).unwrap_or_default();
        if !root.parse.sort_clause.is_empty() {
            current_rel = create_ordered_paths(root, current_rel, &final_target, limit_tuples);
        }
        add_final_paths(root, current_rel, offset_est, count_est);
        return;
    }
    if root.parse.has_aggs {
        super::planagg::preprocess_minmax_aggregates(root);
    }
    let gset_group_clause = gset_data
        .as_ref()
        .map(|gd: &GroupingSetsData| gd.rollups.first().map(|r| r.group_clause.clone()).unwrap_or_default());
    let qp_extra = QpExtra { active_windows: active_windows.clone(), gset_group_clause, setop: setops };
    let mut current_rel = super::query_planner(root, &mut |root| standard_qp_callback(root, &qp_extra));
    let final_target = Rc::new(tlist::create_pathtarget(root, &root.processed_tlist));
    let mut have_postponed_srfs = false;
    let sort_input_target = match root.parse.sort_clause.is_empty() {
        true => final_target.clone(),
        false => Rc::new(make_sort_input_target(root, &final_target, &mut have_postponed_srfs)),
    };
    let grouping_target = match active_windows.is_empty() {
        true => sort_input_target.clone(),
        false => Rc::new(make_window_input_target(root, &final_target, &active_windows)),
    };
    let parse = &root.parse;
    let have_grouping =
        !parse.group_clause.is_empty() || parse.grouping_sets.is_some() || parse.has_aggs || root.has_having_qual;
    let scanjoin_target = match have_grouping {
        true => Rc::new(make_group_input_target(root, &final_target)),
        false => grouping_target.clone(),
    };
    let srfs = !root.parse.target_srfs.is_empty();
    let (mut final_targets, mut sort_input_targets, mut grouping_targets, scanjoin_targets) = match srfs {
        true => {
            let split = |root: &PlannerInfo<'_, '_>, t: &PathTarget, input: Option<&PathTarget>| {
                let (targets, contain) = tlist::split_pathtarget_at_srfs(root, t, input);
                (targets.into_iter().map(Rc::new).collect::<Vec<_>>(), contain)
            };
            (
                split(root, &final_target, Some(&sort_input_target)),
                split(root, &sort_input_target, Some(&grouping_target)),
                split(root, &grouping_target, Some(&scanjoin_target)),
                split(root, &scanjoin_target, None),
            )
        }
        false => (
            (Vec::new(), Vec::new()),
            (Vec::new(), Vec::new()),
            (Vec::new(), Vec::new()),
            (vec![scanjoin_target.clone()], vec![false]),
        ),
    };
    let final_target = final_targets.0.first().cloned().unwrap_or(final_target);
    let sort_input_target = sort_input_targets.0.first().cloned().unwrap_or(sort_input_target);
    let grouping_target = grouping_targets.0.first().cloned().unwrap_or(grouping_target);
    let scanjoin_target_same_exprs =
        scanjoin_targets.0.len() == 1 && scanjoin_targets.0[0].exprs == root.rels[current_rel].reltarget.exprs;
    apply_scanjoin_target_to_paths(
        root,
        current_rel,
        &scanjoin_targets.0,
        &scanjoin_targets.1,
        scanjoin_target_same_exprs,
    );
    if have_grouping {
        current_rel = create_grouping_paths(root, current_rel, grouping_target.clone(), gset_data.as_mut());
        if srfs {
            adjust_paths_for_srfs(root, current_rel, &mut grouping_targets);
        }
    }
    if !active_windows.is_empty() {
        current_rel = create_window_paths(root, current_rel, &grouping_target, &sort_input_target, &active_windows);
        if srfs {
            adjust_paths_for_srfs(root, current_rel, &mut sort_input_targets);
        }
    }
    if !root.parse.distinct_clause.is_empty() {
        current_rel = create_distinct_paths(root, current_rel, &sort_input_target);
    }
    if !root.parse.sort_clause.is_empty() {
        let limit = if have_postponed_srfs { -1.0 } else { limit_tuples };
        current_rel = create_ordered_paths(root, current_rel, &final_target, limit);
        if srfs {
            adjust_paths_for_srfs(root, current_rel, &mut final_targets);
        }
    }
    add_final_paths(root, current_rel, offset_est, count_est);
}

/// add_final_paths adds each path of the relation of the query's upper processing to its final relation, under the
/// query's LIMIT and OFFSET, as grouping_planner does last.
fn add_final_paths(root: &mut PlannerInfo<'_, '_>, current_rel: usize, offset_est: i64, count_est: i64) {
    let final_rel = super::relnode::fetch_upper_rel(root, UpperRelationKind::Final, &Relids::new());
    let limit = limit_needed(&root.parse);
    for path in root.rels[current_rel].pathlist.clone() {
        let path = match limit {
            true => super::pathnode::create_limit_path(root, final_rel, path, offset_est, count_est),
            false => path,
        };
        add_path(&mut root.rels[final_rel], path);
    }
}

/// postprocess_setop_tlist gives the target list of a set operation's result the sort and group clause numbers of
/// the query's own target list, as Postgres' function of the same name does.
fn postprocess_setop_tlist(mut new_tlist: Vec<TargetEntry>, orig_tlist: &[TargetEntry]) -> Vec<TargetEntry> {
    for (new_tle, orig_tle) in new_tlist.iter_mut().filter(|tle| !tle.resjunk).zip(orig_tlist) {
        new_tle.ressortgroupref = orig_tle.ressortgroupref;
    }
    new_tlist
}

/// preprocess_limit returns the share of a query's rows that its LIMIT and OFFSET read, and sets their estimated
/// values, -1 for one that is not a constant, as Postgres' function of the same name does.
fn preprocess_limit(
    root: &mut PlannerInfo<'_, '_>,
    tuple_fraction: f64,
    offset_est: &mut i64,
    count_est: &mut i64,
) -> f64 {
    let estimate = |root: &mut PlannerInfo<'_, '_>, e: &Expr| match e.clone().fold(root.ctx.get_mut()) {
        Expr::Const(Value::Null) => Some(None),
        Expr::Const(v) => Some(v.to_i64()),
        _ => None,
    };
    *count_est = match root.parse.limit_count.clone() {
        Some(count) => match estimate(root, &count) {
            Some(None) => 0,
            Some(Some(n)) => n.max(1),
            None => -1,
        },
        None => 0,
    };
    *offset_est = match root.parse.limit_offset.clone() {
        Some(offset) => match estimate(root, &offset) {
            Some(None) => 0,
            Some(Some(n)) => n.max(0),
            None => -1,
        },
        None => 0,
    };
    let mut tuple_fraction = tuple_fraction;
    if *count_est != 0 {
        let limit_fraction = match *count_est < 0 || *offset_est < 0 {
            true => DEFAULT_LIMIT_FRACTION,
            false => *count_est as f64 + *offset_est as f64,
        };
        if tuple_fraction >= 1.0 {
            if limit_fraction >= 1.0 {
                tuple_fraction = tuple_fraction.min(limit_fraction);
            }
        } else if tuple_fraction > 0.0 {
            tuple_fraction = match limit_fraction >= 1.0 {
                true => limit_fraction,
                false => tuple_fraction.min(limit_fraction),
            };
        } else {
            tuple_fraction = limit_fraction;
        }
    } else if *offset_est != 0 && tuple_fraction > 0.0 {
        let limit_fraction = if *offset_est < 0 { DEFAULT_LIMIT_FRACTION } else { *offset_est as f64 };
        if tuple_fraction >= 1.0 {
            tuple_fraction = match limit_fraction >= 1.0 {
                true => tuple_fraction + limit_fraction,
                false => limit_fraction,
            };
        } else if limit_fraction < 1.0 {
            tuple_fraction += limit_fraction;
            if tuple_fraction >= 1.0 {
                tuple_fraction = 0.0;
            }
        }
    }
    tuple_fraction
}

/// limit_needed reports whether a query's LIMIT and OFFSET can drop rows, as Postgres' function of the same name
/// does.
pub fn limit_needed(parse: &Query) -> bool {
    let non_null = |e: &Expr| !matches!(e, Expr::Const(Value::Null));
    if parse.limit_count.as_ref().is_some_and(non_null) {
        return true;
    }
    match &parse.limit_offset {
        Some(Expr::Const(v)) => !v.is_null() && v.to_i64() != Some(0),
        Some(_) => true,
        None => false,
    }
}

/// preprocess_groupclause returns a query's group clauses in the order of its ORDER BY as far as it lists them, then
/// the rest, or in the order of a grouping set's clause numbers, as Postgres' function of the same name does.
fn preprocess_groupclause(root: &PlannerInfo<'_, '_>, force: Option<&[usize]>) -> Vec<SortGroupClause> {
    let parse = &root.parse;
    if let Some(force) = force {
        return force
            .iter()
            .map(|&r| *parse.group_clause.iter().find(|gc| gc.tle_sort_group_ref == r).expect("a group clause"))
            .collect();
    }
    if parse.sort_clause.is_empty() {
        return parse.group_clause.clone();
    }
    let mut new_groupclause: Vec<SortGroupClause> = Vec::new();
    for sc in &parse.sort_clause {
        match parse.group_clause.iter().find(|gc| *gc == sc) {
            Some(gc) => new_groupclause.push(*gc),
            None => break,
        }
    }
    if new_groupclause.is_empty() {
        return parse.group_clause.clone();
    }
    for gc in &parse.group_clause {
        if !new_groupclause.contains(gc) {
            new_groupclause.push(*gc);
        }
    }
    new_groupclause
}

/// preprocess_grouping_sets chains the query's grouping sets into rollups that sorted aggregations compute, each set
/// of a rollup a prefix of the next, as Postgres' function of the same name does. Every Doltgres type can be sorted.
fn preprocess_grouping_sets(root: &mut PlannerInfo<'_, '_>) -> GroupingSetsData {
    let parse = &root.parse;
    root.processed_group_clause = parse.group_clause.clone();
    let unhashable_refs: BTreeSet<usize> =
        parse.group_clause.iter().filter(|gc| !gc.hashable).map(|gc| gc.tle_sort_group_ref).collect();
    let mut grouping_sets: Vec<Vec<usize>> = parse
        .grouping_sets
        .iter()
        .flatten()
        .map(|set| set.iter().map(|&p| parse.group_clause[p].tle_sort_group_ref).collect())
        .collect();
    grouping_sets.sort_by_key(Vec::len);
    let sets = extract_rollup_sets(grouping_sets);
    let mut gd = GroupingSetsData {
        rollups: Vec::new(),
        hash_sets_idx: Vec::new(),
        d_num_hash_groups: 0.0,
        any_hashable: false,
        unhashable_refs,
        unsortable_sets: Vec::new(),
    };
    let only_one = sets.len() == 1;
    for current_sets in sets {
        let sort_clause = match only_one {
            true => root.parse.sort_clause.clone(),
            false => Vec::new(),
        };
        let current_sets = reorder_grouping_sets(current_sets, &sort_clause);
        let gs = &current_sets[0];
        let group_clause = match gs.set.is_empty() {
            true => Vec::new(),
            false => preprocess_groupclause(root, Some(&gs.set)),
        };
        let hashable = !gs.set.is_empty() && !gs.set.iter().any(|r| gd.unhashable_refs.contains(r));
        gd.any_hashable |= hashable;
        let gsets = remap_to_groupclause_idx(&group_clause, &current_sets);
        gd.rollups.push(RollupData {
            group_clause,
            gsets,
            gsets_data: current_sets,
            num_groups: 0.0,
            hashable,
            is_hashed: false,
        });
    }
    gd
}

/// remap_to_groupclause_idx returns each grouping set as the positions of its keys among group clauses, as Postgres'
/// function of the same name does.
fn remap_to_groupclause_idx(group_clause: &[SortGroupClause], gsets: &[GroupingSetData]) -> Vec<Vec<usize>> {
    let position = |r: &usize| group_clause.iter().position(|gc| gc.tle_sort_group_ref == *r).expect("a group key");
    gsets.iter().map(|gs| gs.set.iter().map(position).collect()).collect()
}

/// extract_rollup_sets splits grouping sets, sorted by size, into the fewest chains of sets that each contain the one
/// before, keeping the empty sets with the first chain, as Postgres' function of the same name does with a maximum
/// matching of the graph of which sets contain which.
fn extract_rollup_sets(grouping_sets: Vec<Vec<usize>>) -> Vec<Vec<Vec<usize>>> {
    let num_empty = grouping_sets.iter().take_while(|s| s.is_empty()).count();
    if num_empty == grouping_sets.len() {
        return vec![grouping_sets];
    }
    let mut orig_sets: Vec<Vec<Vec<usize>>> = vec![Vec::new()];
    let mut set_masks: Vec<BTreeSet<usize>> = vec![BTreeSet::new()];
    let mut adjacency: Vec<Vec<usize>> = vec![Vec::new()];
    let (mut j_size, mut j) = (0, 0);
    for candidate in grouping_sets.into_iter().skip(num_empty) {
        let candidate_set: BTreeSet<usize> = candidate.iter().copied().collect();
        let i = orig_sets.len();
        let mut dup_of = None;
        if j_size == candidate.len() {
            dup_of = (j..i).find(|&k| set_masks[k] == candidate_set);
        } else if j_size < candidate.len() {
            j_size = candidate.len();
            j = i;
        }
        match dup_of {
            Some(k) => orig_sets[k].push(candidate),
            None => {
                orig_sets.push(vec![candidate]);
                adjacency.push((1..j).rev().filter(|&k| set_masks[k].is_subset(&candidate_set)).collect());
                set_masks.push(candidate_set);
            }
        }
    }
    let num_sets = orig_sets.len() - 1;
    let state = super::bipartite_match::bipartite_match(num_sets, num_sets, &adjacency);
    let mut chains = vec![0; num_sets + 1];
    let mut num_chains = 0;
    for i in 1..=num_sets {
        let (u, v) = (state.pair_vu[i], state.pair_uv[i]);
        chains[i] = if u > 0 && u < i {
            chains[u]
        } else if v > 0 && v < i {
            chains[v]
        } else {
            num_chains += 1;
            num_chains
        };
    }
    let mut results: Vec<Vec<Vec<usize>>> = vec![Vec::new(); num_chains + 1];
    for (i, sets) in orig_sets.into_iter().enumerate().skip(1) {
        results[chains[i]].extend(sets);
    }
    for _ in 0..num_empty {
        results[1].insert(0, Vec::new());
    }
    results.into_iter().skip(1).collect()
}

/// reorder_grouping_sets orders the keys of each set of a chain of grouping sets so that each set is a prefix of the
/// next, following the ORDER BY's keys as far as it can, and returns the sets largest first, as Postgres' function of
/// the same name does.
fn reorder_grouping_sets(grouping_sets: Vec<Vec<usize>>, sortclause: &[SortGroupClause]) -> Vec<GroupingSetData> {
    let mut sortclause = sortclause;
    let mut previous: Vec<usize> = Vec::new();
    let mut result = Vec::new();
    for candidate in grouping_sets {
        let mut new_elems: Vec<usize> = candidate.into_iter().filter(|r| !previous.contains(r)).collect();
        while sortclause.len() > previous.len() && !new_elems.is_empty() {
            let r = sortclause[previous.len()].tle_sort_group_ref;
            match new_elems.iter().position(|&e| e == r) {
                Some(p) => {
                    previous.push(r);
                    new_elems.remove(p);
                }
                None => {
                    sortclause = &[];
                    break;
                }
            }
        }
        previous.extend(new_elems);
        result.insert(0, GroupingSetData { set: previous.clone(), num_groups: 0.0 });
    }
    result
}

/// consider_groupingsets_paths adds the paths of an aggregation by grouping sets over a path, as Postgres' function
/// of the same name does: over an unsorted path, hashing every set that it can, with the first rollup sorted when
/// the path is in its order; over a sorted path, hashing the rollups that fit in memory beside sorting the others,
/// and sorting every rollup.
#[allow(clippy::too_many_arguments)]
fn consider_groupingsets_paths(
    root: &mut PlannerInfo<'_, '_>,
    grouped_rel: usize,
    path: Rc<Path>,
    is_sorted: bool,
    can_hash: bool,
    gd: &GroupingSetsData,
    agg_costs: &super::prepagg::AggClauseCosts,
    d_num_groups: f64,
) {
    let hash_mem_limit = super::costsize::HASH_MEM;
    let having: Vec<Expr> = root.parse.having_qual.iter().flat_map(crate::indexscan::conjuncts).cloned().collect();
    let hashed_rollup = |root: &PlannerInfo<'_, '_>, gs: &GroupingSetData| {
        let group_clause = preprocess_groupclause(root, Some(&gs.set));
        let gsets = remap_to_groupclause_idx(&group_clause, std::slice::from_ref(gs));
        RollupData {
            group_clause,
            gsets,
            gsets_data: vec![gs.clone()],
            num_groups: gs.num_groups,
            hashable: true,
            is_hashed: true,
        }
    };
    if !is_sorted {
        let mut new_rollups = Vec::new();
        let mut unhashed_rollup = None;
        let mut l_start = 0;
        let mut strat = AggStrategy::Hashed;
        let mut exclude_groups = 0.0;
        if let Some(first) = gd.rollups.first()
            && super::pathkeys::pathkeys_contained_in(&root.group_pathkeys, &path.pathkeys)
        {
            unhashed_rollup = Some(first.clone());
            exclude_groups = first.num_groups;
            l_start = 1;
        }
        let hashsize =
            super::selfuncs::estimate_hashagg_tablesize(root, &path, agg_costs, d_num_groups - exclude_groups);
        if hashsize > hash_mem_limit && !gd.rollups.is_empty() {
            return;
        }
        let mut sets_data = gd.unsortable_sets.clone();
        for rollup in &gd.rollups[l_start..] {
            if !rollup.hashable {
                return;
            }
            sets_data.extend(rollup.gsets_data.iter().cloned());
        }
        let (mut empty_sets_data, mut empty_sets) = (Vec::new(), Vec::new());
        for gs in sets_data {
            match gs.set.is_empty() {
                true => {
                    empty_sets_data.push(gs);
                    empty_sets.push(Vec::new());
                }
                false => new_rollups.push(hashed_rollup(root, &gs)),
            }
        }
        if new_rollups.is_empty() {
            return;
        }
        if let Some(unhashed_rollup) = unhashed_rollup {
            new_rollups.push(unhashed_rollup);
            strat = AggStrategy::Mixed;
        } else if !empty_sets.is_empty() {
            new_rollups.push(RollupData {
                group_clause: Vec::new(),
                num_groups: empty_sets.len() as f64,
                gsets: empty_sets,
                gsets_data: empty_sets_data,
                hashable: false,
                is_hashed: false,
            });
            strat = AggStrategy::Mixed;
        }
        let path =
            super::pathnode::create_groupingsets_path(root, grouped_rel, path, having, strat, new_rollups, agg_costs);
        add_path(&mut root.rels[grouped_rel], path);
        return;
    }
    if gd.rollups.is_empty() {
        return;
    }
    if can_hash && gd.any_hashable {
        let mut rollups: Vec<RollupData> = Vec::new();
        let mut hash_sets = gd.unsortable_sets.clone();
        let availspace =
            hash_mem_limit - super::selfuncs::estimate_hashagg_tablesize(root, &path, agg_costs, gd.d_num_hash_groups);
        if availspace > 0.0 && gd.rollups.len() > 1 {
            let num_rollups = gd.rollups.len();
            let scale = (availspace / (20.0 * num_rollups as f64)).max(1.0);
            let k_capacity = (availspace / scale).floor() as usize;
            let mut k_weights = Vec::new();
            for rollup in gd.rollups[1..].iter().filter(|r| r.hashable) {
                let sz = super::selfuncs::estimate_hashagg_tablesize(root, &path, agg_costs, rollup.num_groups);
                k_weights.push((sz / scale).floor().min(k_capacity as f64 + 1.0) as usize);
            }
            if !k_weights.is_empty() {
                let hash_items = super::knapsack::discrete_knapsack(k_capacity, &k_weights, None);
                if !hash_items.is_empty() {
                    rollups.push(gd.rollups[0].clone());
                    let mut i = 0;
                    for rollup in &gd.rollups[1..] {
                        if rollup.hashable {
                            match hash_items.contains(&i) {
                                true => hash_sets.extend(rollup.gsets_data.iter().cloned()),
                                false => rollups.push(rollup.clone()),
                            }
                            i += 1;
                        } else {
                            rollups.push(rollup.clone());
                        }
                    }
                }
            }
        }
        if rollups.is_empty() && !hash_sets.is_empty() {
            rollups = gd.rollups.clone();
        }
        for gs in &hash_sets {
            rollups.insert(0, hashed_rollup(root, gs));
        }
        if !rollups.is_empty() {
            let path = super::pathnode::create_groupingsets_path(
                root,
                grouped_rel,
                path.clone(),
                having.clone(),
                AggStrategy::Mixed,
                rollups,
                agg_costs,
            );
            add_path(&mut root.rels[grouped_rel], path);
        }
    }
    if gd.unsortable_sets.is_empty() {
        let rollups = gd.rollups.clone();
        let path = super::pathnode::create_groupingsets_path(
            root,
            grouped_rel,
            path,
            having,
            AggStrategy::Sorted,
            rollups,
            agg_costs,
        );
        add_path(&mut root.rels[grouped_rel], path);
    }
}

/// standard_qp_callback sets the orders that the query's upper processing asks for, and the order that the join of
/// its relations should give, as Postgres' function of the same name does.
fn standard_qp_callback(root: &mut PlannerInfo<'_, '_>, qp_extra: &QpExtra<'_>) {
    let tlist = root.processed_tlist.clone();
    if let Some(group_clause) = &qp_extra.gset_group_clause {
        match tlist::grouping_is_sortable(group_clause)
            .then(|| make_pathkeys_for_sortclauses(root, group_clause, &tlist))
            .flatten()
        {
            Some(pathkeys) => {
                root.num_groupby_pathkeys = pathkeys.len();
                root.group_pathkeys = pathkeys;
            }
            None => {
                root.group_pathkeys = Vec::new();
                root.num_groupby_pathkeys = 0;
            }
        }
    } else if !root.parse.group_clause.is_empty() || root.num_ordered_aggs > 0 {
        let mut clauses = std::mem::take(&mut root.processed_group_clause);
        let pathkeys = make_pathkeys_for_sortclauses_extended(root, &mut clauses, &tlist, true, true);
        root.processed_group_clause = clauses;
        match pathkeys {
            None => {
                root.group_pathkeys = Vec::new();
                root.num_groupby_pathkeys = 0;
            }
            Some(pathkeys) => {
                root.num_groupby_pathkeys = pathkeys.len();
                root.group_pathkeys = pathkeys;
                if root.num_ordered_aggs > 0 {
                    adjust_group_pathkeys_for_groupagg(root);
                }
            }
        }
    } else {
        root.group_pathkeys = Vec::new();
        root.num_groupby_pathkeys = 0;
    }
    root.window_pathkeys = match qp_extra.active_windows.first() {
        Some(window) => make_pathkeys_for_window(root, window[0]),
        None => Vec::new(),
    };
    if root.parse.distinct_clause.is_empty() {
        root.distinct_pathkeys = Vec::new();
    } else {
        let mut clauses = root.parse.distinct_clause.clone();
        root.distinct_pathkeys =
            make_pathkeys_for_sortclauses_extended(root, &mut clauses, &tlist, true, false).unwrap_or_default();
        root.processed_distinct_clause = clauses;
    }
    let sort_clause = root.parse.sort_clause.clone();
    root.sort_pathkeys = make_pathkeys_for_sortclauses(root, &sort_clause, &tlist).unwrap_or_default();
    root.setop_pathkeys = match qp_extra.setop {
        Some(setop) => {
            let mut group_clauses = generate_setop_child_grouplist(root, setop);
            let tlist = root.processed_tlist.clone();
            make_pathkeys_for_sortclauses_extended(root, &mut group_clauses, &tlist, false, false).unwrap_or_default()
        }
        None => Vec::new(),
    };
    root.query_pathkeys = if !root.group_pathkeys.is_empty() {
        root.group_pathkeys.clone()
    } else if !root.window_pathkeys.is_empty() {
        root.window_pathkeys.clone()
    } else if root.distinct_pathkeys.len() > root.sort_pathkeys.len() {
        root.distinct_pathkeys.clone()
    } else if !root.sort_pathkeys.is_empty() {
        root.sort_pathkeys.clone()
    } else {
        root.setop_pathkeys.clone()
    };
}

/// adjust_group_pathkeys_for_groupagg extends the grouping's order with the order of the most aggregate calls with
/// ORDER BY or DISTINCT that one order serves, so that a sorted aggregation gives them their input in order, as
/// Postgres' function of the same name does. A call with a FILTER over anything but columns and constants keeps
/// sorting its own input.
fn adjust_group_pathkeys_for_groupagg(root: &mut PlannerInfo<'_, '_>) {
    if !root.enables.presorted_aggregate {
        return;
    }
    let grouppathkeys = root.group_pathkeys.clone();
    let mut unprocessed_aggs: Vec<usize> = Vec::new();
    for (i, call) in root.parse.aggregates.iter().enumerate() {
        if !call.distinct && call.order.is_empty() {
            continue;
        }
        if call.filter.is_some() {
            let allow_presort = call.args.iter().all(|arg| match arg {
                Expr::Column(id) => matches!(root.glob.node(*id), super::nodes::VarNode::Var(_)),
                Expr::Const(_) => true,
                _ => false,
            });
            if !allow_presort {
                continue;
            }
        }
        unprocessed_aggs.push(i);
    }
    let (mut bestpathkeys, mut bestaggs): (Vec<PkId>, Vec<usize>) = (Vec::new(), Vec::new());
    while unprocessed_aggs.len() > bestaggs.len() {
        let mut aggindexes = Vec::new();
        let mut currpathkeys: Vec<PkId> = Vec::new();
        for i in unprocessed_aggs.clone() {
            let call = &root.parse.aggregates[i];
            let sortlist: Vec<crate::plan::SortKey> = match call.distinct {
                true => call
                    .args
                    .iter()
                    .map(|arg| crate::plan::SortKey { expr: arg.clone(), descending: false, nulls_first: false })
                    .collect(),
                false => call
                    .order
                    .iter()
                    .map(|(expr, descending, nulls_first)| crate::plan::SortKey {
                        expr: expr.clone(),
                        descending: *descending,
                        nulls_first: *nulls_first,
                    })
                    .collect(),
            };
            let pathkeys = super::pathkeys::make_pathkeys_for_sortclauses(root, &sortlist);
            let Some(pathkeys) = pathkeys.filter(|pathkeys| !has_volatile_pathkey(root, pathkeys)) else {
                unprocessed_aggs.retain(|&a| a != i);
                continue;
            };
            if currpathkeys.is_empty() {
                currpathkeys = match grouppathkeys.is_empty() {
                    true => pathkeys,
                    false => super::pathkeys::append_pathkeys(root, grouppathkeys.clone(), &pathkeys),
                };
                aggindexes.push(i);
            } else {
                let pathkeys = match grouppathkeys.is_empty() {
                    true => pathkeys,
                    false => super::pathkeys::append_pathkeys(root, grouppathkeys.clone(), &pathkeys),
                };
                match compare_pathkeys(&currpathkeys, &pathkeys) {
                    PathKeysComparison::Better2 => {
                        currpathkeys = pathkeys;
                        aggindexes.push(i);
                    }
                    PathKeysComparison::Better1 | PathKeysComparison::Equal => aggindexes.push(i),
                    PathKeysComparison::Different => {}
                }
            }
        }
        unprocessed_aggs.retain(|a| !aggindexes.contains(a));
        if aggindexes.len() > bestaggs.len() {
            bestaggs = aggindexes;
            bestpathkeys = currpathkeys;
        }
    }
    if !bestpathkeys.is_empty() {
        root.group_pathkeys = bestpathkeys;
    }
}

/// has_volatile_pathkey reports whether an order has a key over a volatile expression, as Postgres' function of the
/// same name does.
fn has_volatile_pathkey(root: &PlannerInfo<'_, '_>, keys: &[PkId]) -> bool {
    keys.iter().any(|&pk| root.eq_classes[root.canon_pathkeys[pk].pk_eclass].ec_has_volatile)
}

/// generate_setop_child_grouplist returns the group clauses that a set operation over the query sorts its target
/// list's columns by, numbering the columns that have no number yet, or none when a column's type differs from the
/// set operation's, as Postgres' function of the same name does.
fn generate_setop_child_grouplist(
    root: &mut PlannerInfo<'_, '_>,
    op: &super::nodes::SetOperationStmt,
) -> Vec<SortGroupClause> {
    let mut grouplist = Vec::new();
    let mut col_types = op.col_types.iter();
    for i in 0..root.processed_tlist.len() {
        if root.processed_tlist[i].resjunk {
            continue;
        }
        let coltype = col_types.next().expect("a column type for each output column");
        if *coltype != super::nodefuncs::expr_type(root, &root.processed_tlist[i].expr) {
            return Vec::new();
        }
        if root.processed_tlist[i].ressortgroupref == 0 {
            let max_ref = root.processed_tlist.iter().map(|tle| tle.ressortgroupref).max().unwrap_or(0);
            root.processed_tlist[i].ressortgroupref = max_ref + 1;
        }
        let tle_sort_group_ref = root.processed_tlist[i].ressortgroupref;
        grouplist.push(SortGroupClause { tle_sort_group_ref, descending: false, nulls_first: false, hashable: true });
    }
    grouplist
}

/// make_pathkeys_for_sortclauses returns the pathkeys of sort or group clauses over a target list, as Postgres'
/// function of the same name does.
pub fn make_pathkeys_for_sortclauses(
    root: &mut PlannerInfo<'_, '_>,
    clauses: &[SortGroupClause],
    tlist: &[super::nodes::TargetEntry],
) -> Option<Vec<PkId>> {
    make_pathkeys_for_sortclauses_extended(root, &mut clauses.to_vec(), tlist, false, false)
}

/// create_grouping_paths returns the relation of the query's grouped rows, with its paths over the paths of the
/// join of its relations, as Postgres' function of the same name does.
fn create_grouping_paths(
    root: &mut PlannerInfo<'_, '_>,
    input_rel: usize,
    target: Rc<PathTarget>,
    gd: Option<&mut GroupingSetsData>,
) -> usize {
    let agg_costs = super::prepagg::get_agg_clause_costs(root);
    let grouped_rel = make_grouping_rel(root, input_rel, target);
    if !root.minmax_aggs.is_empty() {
        let target = Rc::new(tlist::create_pathtarget(root, &root.processed_tlist));
        let quals = root.parse.having_qual.iter().flat_map(crate::indexscan::conjuncts).cloned().collect();
        let mmaggregates = root.minmax_aggs.clone();
        let path = super::pathnode::create_minmaxagg_path(root, grouped_rel, target, mmaggregates, quals);
        add_path(&mut root.rels[grouped_rel], path);
    }
    if is_degenerate_grouping(root) {
        create_degenerate_grouping_paths(root, grouped_rel);
    } else {
        let can_sort = gd.as_ref().is_some_and(|gd| !gd.rollups.is_empty())
            || tlist::grouping_is_sortable(&root.processed_group_clause);
        let can_hash = !root.parse.group_clause.is_empty()
            && root.num_ordered_aggs == 0
            && match &gd {
                Some(gd) => gd.any_hashable,
                None => tlist::grouping_is_hashable(&root.processed_group_clause),
            };
        create_ordinary_grouping_paths(root, input_rel, grouped_rel, &agg_costs, gd, can_sort, can_hash);
    }
    set_cheapest(&mut root.rels[grouped_rel]);
    grouped_rel
}

/// make_grouping_rel returns the relation of the query's grouped rows with its target, as Postgres' function of the
/// same name does.
fn make_grouping_rel(root: &mut PlannerInfo<'_, '_>, input_rel: usize, target: Rc<PathTarget>) -> usize {
    let grouped_rel = super::relnode::fetch_upper_rel(root, UpperRelationKind::GroupAgg, &Relids::new());
    root.rels[grouped_rel].reltarget = (*target).clone();
    root.rels[grouped_rel].consider_startup = root.rels[input_rel].consider_startup;
    grouped_rel
}

/// is_degenerate_grouping reports whether a query groups without group keys or aggregates, so that it returns one
/// row or none for each grouping set, as Postgres' function of the same name does.
fn is_degenerate_grouping(root: &PlannerInfo<'_, '_>) -> bool {
    (root.has_having_qual || root.parse.grouping_sets.is_some())
        && !root.parse.has_aggs
        && root.parse.group_clause.is_empty()
}

/// create_degenerate_grouping_paths adds the path of the one row of a degenerate grouping under its HAVING, as
/// Postgres' function of the same name does.
fn create_degenerate_grouping_paths(root: &mut PlannerInfo<'_, '_>, grouped_rel: usize) {
    let having: Vec<Expr> = root.parse.having_qual.iter().flat_map(crate::indexscan::conjuncts).cloned().collect();
    let nrows = root.parse.grouping_sets.as_ref().map_or(1, |sets| sets.len());
    let target = Rc::new(root.rels[grouped_rel].reltarget.clone());
    let one = |root: &mut PlannerInfo<'_, '_>| {
        super::pathnode::create_group_result_path(root, grouped_rel, target.clone(), having.clone())
    };
    let path = match nrows > 1 {
        true => {
            let paths = (0..nrows).map(|_| one(root)).collect();
            super::pathnode::create_append_path(root, grouped_rel, paths)
        }
        false => one(root),
    };
    add_path(&mut root.rels[grouped_rel], path);
}

/// create_ordinary_grouping_paths adds the paths of the query's grouping over the join of its relations, as
/// Postgres' function of the same name does without partial or partitionwise aggregation.
fn create_ordinary_grouping_paths(
    root: &mut PlannerInfo<'_, '_>,
    input_rel: usize,
    grouped_rel: usize,
    agg_costs: &super::prepagg::AggClauseCosts,
    mut gd: Option<&mut GroupingSetsData>,
    can_sort: bool,
    can_hash: bool,
) {
    let cheapest_path = root.rels[input_rel].cheapest_total_path.clone().expect("every relation has a path");
    let target_list = root.parse.target_list.clone();
    let d_num_groups = get_number_of_groups(root, cheapest_path.rows, gd.as_deref_mut(), &target_list);
    add_paths_to_grouping_rel(root, input_rel, grouped_rel, agg_costs, gd.as_deref(), d_num_groups, can_sort, can_hash);
}

/// add_paths_to_grouping_rel adds sorted aggregations over each of the join's paths, sorted as the group keys need,
/// and a hashed aggregation over its cheapest path, as Postgres' function of the same name does.
#[allow(clippy::too_many_arguments)]
fn add_paths_to_grouping_rel(
    root: &mut PlannerInfo<'_, '_>,
    input_rel: usize,
    grouped_rel: usize,
    agg_costs: &super::prepagg::AggClauseCosts,
    gd: Option<&GroupingSetsData>,
    d_num_groups: f64,
    can_sort: bool,
    can_hash: bool,
) {
    let cheapest_path = root.rels[input_rel].cheapest_total_path.clone().expect("every relation has a path");
    let having: Vec<Expr> = root.parse.having_qual.iter().flat_map(crate::indexscan::conjuncts).cloned().collect();
    let target = Rc::new(root.rels[grouped_rel].reltarget.clone());
    if can_sort {
        for path in root.rels[input_rel].pathlist.clone() {
            for (pathkeys, clauses) in get_useful_group_keys_orderings(root, &path) {
                let Some(path) = make_ordered_path(root, grouped_rel, path.clone(), &cheapest_path, &pathkeys, -1.0)
                else {
                    continue;
                };
                if let Some(gd) = gd {
                    consider_groupingsets_paths(root, grouped_rel, path, true, can_hash, gd, agg_costs, d_num_groups);
                    continue;
                }
                let new_path = if root.parse.has_aggs {
                    let strategy = match root.parse.group_clause.is_empty() {
                        true => AggStrategy::Plain,
                        false => AggStrategy::Sorted,
                    };
                    super::pathnode::create_agg_path(
                        root,
                        grouped_rel,
                        path,
                        target.clone(),
                        strategy,
                        clauses,
                        having.clone(),
                        agg_costs,
                        d_num_groups,
                    )
                } else if !root.parse.group_clause.is_empty() {
                    super::pathnode::create_group_path(root, grouped_rel, path, clauses, having.clone(), d_num_groups)
                } else {
                    continue;
                };
                add_path(&mut root.rels[grouped_rel], new_path);
            }
        }
    }
    if can_hash && let Some(gd) = gd {
        consider_groupingsets_paths(root, grouped_rel, cheapest_path, false, true, gd, agg_costs, d_num_groups);
    } else if can_hash {
        let clauses = root.processed_group_clause.clone();
        let path = super::pathnode::create_agg_path(
            root,
            grouped_rel,
            cheapest_path,
            target,
            AggStrategy::Hashed,
            clauses,
            having,
            agg_costs,
            d_num_groups,
        );
        add_path(&mut root.rels[grouped_rel], path);
    }
}

/// get_useful_group_keys_orderings returns the orders of the group keys that a sorted aggregation over a path could
/// use, with the group clauses in that order: the query's own order, as Postgres' function of the same name begins.
fn get_useful_group_keys_orderings(root: &PlannerInfo<'_, '_>, _path: &Path) -> Vec<(Vec<PkId>, Vec<SortGroupClause>)> {
    vec![(root.group_pathkeys.clone(), root.processed_group_clause.clone())]
}

/// make_ordered_path returns a path in an order: the path itself when it is in that order, the cheapest path sorted,
/// or a path sorted incrementally over the order's leading keys that it gives, as Postgres' function of the same name
/// does, or None for any other path.
fn make_ordered_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    path: Rc<Path>,
    cheapest_path: &Rc<Path>,
    pathkeys: &[PkId],
    limit_tuples: f64,
) -> Option<Rc<Path>> {
    let (is_sorted, presorted_keys) = pathkeys_count_contained_in(pathkeys, &path.pathkeys);
    if is_sorted {
        return Some(path);
    }
    if !Rc::ptr_eq(&path, cheapest_path) && (presorted_keys == 0 || !root.enables.incremental_sort) {
        return None;
    }
    Some(match presorted_keys == 0 || !root.enables.incremental_sort {
        true => super::pathnode::create_sort_path(root, rel, path, pathkeys.to_vec(), limit_tuples),
        false => super::pathnode::create_incremental_sort_path(
            root,
            rel,
            path,
            pathkeys.to_vec(),
            presorted_keys,
            limit_tuples,
        ),
    })
}

/// get_number_of_groups estimates the groups of a query's grouping of a number of rows, setting the estimate of each
/// grouping set and rollup, as Postgres' function of the same name does.
fn get_number_of_groups(
    root: &mut PlannerInfo<'_, '_>,
    path_rows: f64,
    gd: Option<&mut GroupingSetsData>,
    target_list: &[super::nodes::TargetEntry],
) -> f64 {
    if !root.parse.group_clause.is_empty() {
        let Some(gd) = gd else {
            let group_exprs = tlist::get_sortgrouplist_exprs(&root.processed_group_clause, target_list);
            return super::selfuncs::estimate_num_groups(root, &group_exprs, path_rows, None, None);
        };
        let mut d_num_groups = 0.0;
        for rollup in &mut gd.rollups {
            let group_exprs = tlist::get_sortgrouplist_exprs(&rollup.group_clause, target_list);
            rollup.num_groups = 0.0;
            for (gset, gs) in rollup.gsets.iter().zip(&mut rollup.gsets_data) {
                let num_groups = super::selfuncs::estimate_num_groups(root, &group_exprs, path_rows, Some(gset), None);
                gs.num_groups = num_groups;
                rollup.num_groups += num_groups;
            }
            d_num_groups += rollup.num_groups;
        }
        if !gd.hash_sets_idx.is_empty() {
            gd.d_num_hash_groups = 0.0;
            let group_exprs = tlist::get_sortgrouplist_exprs(&root.parse.group_clause, target_list);
            for (gset, gs) in gd.hash_sets_idx.iter().zip(&mut gd.unsortable_sets) {
                let num_groups = super::selfuncs::estimate_num_groups(root, &group_exprs, path_rows, Some(gset), None);
                gs.num_groups = num_groups;
                gd.d_num_hash_groups += num_groups;
            }
            d_num_groups += gd.d_num_hash_groups;
        }
        return d_num_groups;
    }
    match &root.parse.grouping_sets {
        Some(sets) => sets.len() as f64,
        None => 1.0,
    }
}

/// select_active_windows returns the query's windows, each as the positions of the window calls that share its
/// partition and order, in the order that sorts their rows least, as Postgres' function of the same name does: a
/// window whose sort keys lead another's comes after it.
fn select_active_windows(root: &PlannerInfo<'_, '_>) -> Vec<Vec<usize>> {
    let mut actives: Vec<(Vec<crate::plan::SortKey>, Vec<usize>)> = Vec::new();
    for (k, call) in root.parse.window_funcs.iter().enumerate() {
        let unique_order = window_unique_order(call);
        match actives.iter_mut().find(|(order, _)| *order == unique_order) {
            Some((_, members)) => members.push(k),
            None => actives.push((unique_order, vec![k])),
        }
    }
    let tlist = &root.parse.target_list;
    let mut refs: Vec<Expr> =
        root.parse.sort_clause.iter().map(|c| tlist::get_sortgroupclause_expr(c, tlist)).collect();
    for (unique_order, _) in &actives {
        for key in unique_order {
            if !refs.contains(&key.expr) {
                refs.push(key.expr.clone());
            }
        }
    }
    let rank = |key: &crate::plan::SortKey| (refs.iter().position(|e| *e == key.expr), key.descending, key.nulls_first);
    actives.sort_by(|(a, _), (b, _)| {
        a.iter().zip(b).map(|(x, y)| rank(y).cmp(&rank(x))).find(|o| o.is_ne()).unwrap_or(b.len().cmp(&a.len()))
    });
    actives.into_iter().map(|(_, members)| members).collect()
}

/// window_unique_order returns a window call's partition keys and then its order keys that are not among them, as
/// Postgres' select_active_windows lists a window's sort keys.
fn window_unique_order(call: &crate::window::WindowCall) -> Vec<crate::plan::SortKey> {
    let mut unique_order: Vec<crate::plan::SortKey> = call
        .partition
        .iter()
        .map(|e| crate::plan::SortKey { expr: e.clone(), descending: false, nulls_first: false })
        .collect();
    for key in &call.order {
        if !unique_order.contains(key) {
            unique_order.push(key.clone());
        }
    }
    unique_order
}

/// make_pathkeys_for_window returns the order of a window's rows, its partition and then its order, as Postgres'
/// function of the same name does.
fn make_pathkeys_for_window(root: &mut PlannerInfo<'_, '_>, call: usize) -> Vec<PkId> {
    let keys = window_unique_order(&root.parse.window_funcs[call]);
    super::pathkeys::make_pathkeys_for_sortclauses(root, &keys).unwrap_or_default()
}

/// create_window_paths returns the relation of the query's rows with its window calls, with the path of each window
/// in turn over the grouped relation's paths, as Postgres' function of the same name does.
fn create_window_paths(
    root: &mut PlannerInfo<'_, '_>,
    input_rel: usize,
    input_target: &Rc<PathTarget>,
    output_target: &Rc<PathTarget>,
    active_windows: &[Vec<usize>],
) -> usize {
    let window_rel = super::relnode::fetch_upper_rel(root, UpperRelationKind::Window, &Relids::new());
    root.rels[window_rel].reltarget = (**output_target).clone();
    for path in root.rels[input_rel].pathlist.clone() {
        let cheapest = root.rels[input_rel].cheapest_total_path.clone().expect("every relation has a path");
        if Rc::ptr_eq(&path, &cheapest) || pathkeys_count_contained_in(&root.window_pathkeys, &path.pathkeys).1 > 0 {
            create_one_window_path(root, window_rel, path, input_target, output_target, active_windows);
        }
    }
    set_cheapest(&mut root.rels[window_rel]);
    window_rel
}

/// create_one_window_path adds the path of the query's windows over one path: for each window, a sort of the rows by
/// its keys unless they are in that order, then its window calls, as Postgres' function of the same name does.
fn create_one_window_path(
    root: &mut PlannerInfo<'_, '_>,
    window_rel: usize,
    path: Rc<Path>,
    input_target: &Rc<PathTarget>,
    output_target: &Rc<PathTarget>,
    active_windows: &[Vec<usize>],
) {
    let mut path = path;
    let mut window_target = input_target.clone();
    for (i, calls) in active_windows.iter().enumerate() {
        let window_pathkeys = make_pathkeys_for_window(root, calls[0]);
        let (is_sorted, presorted_keys) = pathkeys_count_contained_in(&window_pathkeys, &path.pathkeys);
        if !is_sorted {
            path = match presorted_keys == 0 || !root.enables.incremental_sort {
                true => super::pathnode::create_sort_path(root, window_rel, path, window_pathkeys, -1.0),
                false => super::pathnode::create_incremental_sort_path(
                    root,
                    window_rel,
                    path,
                    window_pathkeys,
                    presorted_keys,
                    -1.0,
                ),
            };
        }
        let topwindow = i + 1 == active_windows.len();
        window_target = match topwindow {
            true => output_target.clone(),
            false => {
                let mut target = (*window_target).clone();
                for &call in calls {
                    tlist::add_column_to_pathtarget(&mut target, Expr::WindowRef(call), 0);
                }
                super::costsize::set_pathtarget_cost_width(root, &mut target);
                Rc::new(target)
            }
        };
        path = super::pathnode::create_windowagg_path(root, window_rel, path, window_target.clone(), calls.clone());
    }
    add_path(&mut root.rels[window_rel], path);
}

/// create_distinct_paths returns the relation of the query's distinct rows, as Postgres' function of the same name
/// does.
fn create_distinct_paths(root: &mut PlannerInfo<'_, '_>, input_rel: usize, target: &Rc<PathTarget>) -> usize {
    let distinct_rel = super::relnode::fetch_upper_rel(root, UpperRelationKind::Distinct, &Relids::new());
    root.rels[distinct_rel].reltarget = (**target).clone();
    create_final_distinct_paths(root, input_rel, distinct_rel);
    set_cheapest(&mut root.rels[distinct_rel]);
    distinct_rel
}

/// create_final_distinct_paths adds the paths of DISTINCT over the input relation's paths: a unique pass over each
/// path sorted by the distinct keys, and a hashed aggregation over the cheapest path, as Postgres' function of the
/// same name does.
fn create_final_distinct_paths(root: &mut PlannerInfo<'_, '_>, input_rel: usize, distinct_rel: usize) {
    let cheapest_input_path = root.rels[input_rel].cheapest_total_path.clone().expect("every relation has a path");
    let parse = &root.parse;
    let num_distinct_rows =
        match !parse.group_clause.is_empty() || parse.grouping_sets.is_some() || parse.has_aggs || root.has_having_qual
        {
            true => cheapest_input_path.rows,
            false => {
                let distinct_exprs =
                    tlist::get_sortgrouplist_exprs(&root.processed_distinct_clause, &parse.target_list);
                super::selfuncs::estimate_num_groups(root, &distinct_exprs, cheapest_input_path.rows, None, None)
            }
        };
    if tlist::grouping_is_sortable(&root.processed_distinct_clause) {
        let limittuples = if root.distinct_pathkeys.is_empty() { 1.0 } else { -1.0 };
        let needed_pathkeys =
            match root.parse.has_distinct_on && root.distinct_pathkeys.len() < root.sort_pathkeys.len() {
                true => root.sort_pathkeys.clone(),
                false => root.distinct_pathkeys.clone(),
            };
        for input_path in root.rels[input_rel].pathlist.clone() {
            for useful_pathkeys in get_useful_pathkeys_for_distinct(root, &needed_pathkeys, &input_path.pathkeys) {
                let Some(sorted_path) = make_ordered_path(
                    root,
                    distinct_rel,
                    input_path.clone(),
                    &cheapest_input_path,
                    &useful_pathkeys,
                    limittuples,
                ) else {
                    continue;
                };
                let path = match root.distinct_pathkeys.is_empty() {
                    true => super::pathnode::create_limit_path_one(root, distinct_rel, sorted_path),
                    false => {
                        let num_keys = root.distinct_pathkeys.len();
                        super::pathnode::create_upper_unique_path(
                            root,
                            distinct_rel,
                            sorted_path,
                            num_keys,
                            num_distinct_rows,
                        )
                    }
                };
                add_path(&mut root.rels[distinct_rel], path);
            }
        }
    }
    let allow_hash =
        root.rels[distinct_rel].pathlist.is_empty() || !(root.parse.has_distinct_on || !root.enables.hashagg);
    if allow_hash && tlist::grouping_is_hashable(&root.processed_distinct_clause) {
        let target = Rc::new(root.rels[distinct_rel].reltarget.clone());
        let clauses = root.processed_distinct_clause.clone();
        let costs = super::prepagg::AggClauseCosts::default();
        let path = super::pathnode::create_agg_path(
            root,
            distinct_rel,
            cheapest_input_path,
            target,
            AggStrategy::Hashed,
            clauses,
            Vec::new(),
            &costs,
            num_distinct_rows,
        );
        add_path(&mut root.rels[distinct_rel], path);
    }
}

/// get_useful_pathkeys_for_distinct returns the orders that a DISTINCT's unique pass could use: the needed order,
/// and the query's ORDER BY when the distinct keys lead it, as Postgres' function of the same name does.
fn get_useful_pathkeys_for_distinct(
    root: &PlannerInfo<'_, '_>,
    needed_pathkeys: &[PkId],
    _path_pathkeys: &[PkId],
) -> Vec<Vec<PkId>> {
    let mut useful_pathkeys_list = vec![needed_pathkeys.to_vec()];
    if !root.parse.has_distinct_on
        && root.sort_pathkeys.len() >= needed_pathkeys.len()
        && pathkeys_contained_in(needed_pathkeys, &root.sort_pathkeys)
        && root.sort_pathkeys != needed_pathkeys
    {
        useful_pathkeys_list.push(root.sort_pathkeys.clone());
    }
    useful_pathkeys_list
}

/// create_ordered_paths returns the relation of the query's rows in its ORDER BY order, with each input path in that
/// order, the cheapest path sorted, and paths sorted incrementally over the leading keys they give, as Postgres'
/// function of the same name does.
fn create_ordered_paths(
    root: &mut PlannerInfo<'_, '_>,
    input_rel: usize,
    target: &Rc<PathTarget>,
    limit_tuples: f64,
) -> usize {
    let ordered_rel = super::relnode::fetch_upper_rel(root, UpperRelationKind::Ordered, &Relids::new());
    root.rels[ordered_rel].reltarget = (**target).clone();
    root.rels[ordered_rel].consider_startup = root.tuple_fraction > 0.0;
    let cheapest_input_path = root.rels[input_rel].cheapest_total_path.clone().expect("every relation has a path");
    let sort_pathkeys = root.sort_pathkeys.clone();
    for input_path in root.rels[input_rel].pathlist.clone() {
        let (is_sorted, presorted_keys) = pathkeys_count_contained_in(&sort_pathkeys, &input_path.pathkeys);
        let sorted_path = if is_sorted {
            input_path.clone()
        } else if !Rc::ptr_eq(&input_path, &cheapest_input_path)
            && (presorted_keys == 0 || !root.enables.incremental_sort)
        {
            continue;
        } else if presorted_keys == 0 || !root.enables.incremental_sort {
            super::pathnode::create_sort_path(
                root,
                ordered_rel,
                input_path.clone(),
                sort_pathkeys.clone(),
                limit_tuples,
            )
        } else {
            super::pathnode::create_incremental_sort_path(
                root,
                ordered_rel,
                input_path.clone(),
                sort_pathkeys.clone(),
                presorted_keys,
                limit_tuples,
            )
        };
        let sorted_path = match path_exprs(root, &sorted_path) == target.exprs {
            true => sorted_path,
            false => super::pathnode::apply_projection_to_path(root, ordered_rel, sorted_path, target.clone()),
        };
        add_path(&mut root.rels[ordered_rel], sorted_path);
    }
    set_cheapest(&mut root.rels[ordered_rel]);
    ordered_rel
}

/// path_exprs returns the expressions of a path's rows.
pub fn path_exprs(root: &PlannerInfo<'_, '_>, path: &Path) -> Vec<Expr> {
    match &path.pathtarget {
        Some(target) => target.exprs.clone(),
        None => root.rels[path.parent].reltarget.exprs.clone(),
    }
}

/// make_group_input_target returns the target of the rows that the query's grouping reads: its group keys and the
/// Vars and PlaceHolderVars of the rest of its target and HAVING outside aggregates, as Postgres' function of the
/// same name does.
fn make_group_input_target(root: &PlannerInfo<'_, '_>, final_target: &PathTarget) -> PathTarget {
    let mut input_target = PathTarget::default();
    let mut non_group_cols: Vec<Expr> = Vec::new();
    for (i, expr) in final_target.exprs.iter().enumerate() {
        let sgref = tlist::get_pathtarget_sortgroupref(final_target, i);
        if sgref != 0
            && !root.processed_group_clause.is_empty()
            && root.processed_group_clause.iter().any(|c| c.tle_sort_group_ref == sgref)
        {
            tlist::add_column_to_pathtarget(&mut input_target, expr.clone(), sgref);
        } else {
            non_group_cols.push(expr.clone());
        }
    }
    non_group_cols.extend(root.parse.having_qual.iter().cloned());
    let mut non_group_vars: Vec<Expr> = Vec::new();
    for expr in &non_group_cols {
        collect_grouping_input_vars(root, expr, &mut non_group_vars);
    }
    tlist::add_new_columns_to_pathtarget(&mut input_target, non_group_vars);
    super::costsize::set_pathtarget_cost_width(root, &mut input_target);
    input_target
}

/// collect_grouping_input_vars adds the Vars and PlaceHolderVars that an expression of a grouped query reads, with
/// the arguments of its aggregate calls, as Postgres' pull_var_clause with PVC_RECURSE_AGGREGATES finds them.
fn collect_grouping_input_vars(root: &PlannerInfo<'_, '_>, e: &Expr, out: &mut Vec<Expr>) {
    match e {
        Expr::AggRef(k) => {
            let call = &root.parse.aggregates[*k];
            for arg in call.args.iter().chain(&call.filter).chain(call.order.iter().map(|(e, ..)| e)) {
                collect_grouping_input_vars(root, arg, out);
            }
        }
        Expr::Grouping(..) => {}
        Expr::Column(id) => {
            if !out.contains(e) {
                let _ = id;
                out.push(e.clone());
            }
        }
        other => other.visit_children(&mut |c| collect_grouping_input_vars(root, c, out)),
    }
}

/// make_window_input_target returns the target of the rows that the query's first window reads: the partition and
/// order keys of every window and the Vars, aggregates, and grouped expressions of the rest of its target outside
/// window calls, as Postgres' function of the same name does.
fn make_window_input_target(
    root: &PlannerInfo<'_, '_>,
    final_target: &PathTarget,
    active_windows: &[Vec<usize>],
) -> PathTarget {
    let mut input_target = PathTarget::default();
    let mut flattenable_cols: Vec<Expr> = Vec::new();
    let mut window_keys: Vec<Expr> = Vec::new();
    for window in active_windows {
        for key in window_unique_order(&root.parse.window_funcs[window[0]]) {
            if !window_keys.contains(&key.expr) {
                window_keys.push(key.expr);
            }
        }
    }
    for (i, expr) in final_target.exprs.iter().enumerate() {
        let sgref = tlist::get_pathtarget_sortgroupref(final_target, i);
        if sgref != 0 && window_keys.contains(expr) {
            tlist::add_column_to_pathtarget(&mut input_target, expr.clone(), sgref);
        } else {
            flattenable_cols.push(expr.clone());
        }
    }
    for key in window_keys {
        tlist::add_new_column_to_pathtarget(&mut input_target, key);
    }
    let mut flattenable_vars: Vec<Expr> = Vec::new();
    for expr in &flattenable_cols {
        collect_window_input_vars(root, expr, &mut flattenable_vars);
    }
    tlist::add_new_columns_to_pathtarget(&mut input_target, flattenable_vars);
    super::costsize::set_pathtarget_cost_width(root, &mut input_target);
    input_target
}

/// collect_window_input_vars adds the Vars, PlaceHolderVars, aggregate calls, and GROUPING calls that an expression
/// reads, with the arguments of its window calls, as Postgres' pull_var_clause with PVC_INCLUDE_AGGREGATES and
/// PVC_RECURSE_WINDOWFUNCS finds them.
fn collect_window_input_vars(root: &PlannerInfo<'_, '_>, e: &Expr, out: &mut Vec<Expr>) {
    match e {
        Expr::WindowRef(k) => {
            let call = &root.parse.window_funcs[*k];
            for arg in call.args.iter().chain(&call.filter) {
                collect_window_input_vars(root, arg, out);
            }
        }
        Expr::Column(_) | Expr::AggRef(_) | Expr::Grouping(..) => {
            if !out.contains(e) {
                out.push(e.clone());
            }
        }
        other => other.visit_children(&mut |c| collect_window_input_vars(root, c, out)),
    }
}

/// make_sort_input_target returns the target of the rows that the query's sort reads: the final target, where
/// expressions that need not be computed before the sort, as set-returning functions and costly expressions that a
/// LIMIT may skip, become the Vars they read, as Postgres' function of the same name does.
fn make_sort_input_target(
    root: &PlannerInfo<'_, '_>,
    final_target: &PathTarget,
    have_postponed_srfs: &mut bool,
) -> PathTarget {
    let mut postpone = vec![false; final_target.exprs.len()];
    let mut have_srf = false;
    let mut have_volatile = false;
    let mut have_expensive = false;
    let mut have_srf_sortcols = false;
    for (i, expr) in final_target.exprs.iter().enumerate() {
        let sgref = tlist::get_pathtarget_sortgroupref(final_target, i);
        let contains_srf = super::clauses::expression_returns_set(expr);
        if sgref != 0 {
            if contains_srf {
                have_srf_sortcols = true;
            }
            continue;
        }
        if contains_srf {
            postpone[i] = true;
            have_srf = true;
        } else if super::clauses::contain_volatile_functions(root.glob, expr) {
            postpone[i] = true;
            have_volatile = true;
        } else if root.tuple_fraction > 0.0
            && super::costsize::cost_qual_eval_node(expr).per_tuple > 10.0 * super::costsize::CPU_OPERATOR_COST
        {
            postpone[i] = true;
            have_expensive = true;
        }
    }
    if have_srf_sortcols {
        have_srf = false;
    }
    if !(have_srf || have_volatile || have_expensive) {
        return final_target.clone();
    }
    *have_postponed_srfs = have_srf;
    let mut input_target = PathTarget::default();
    let mut postponable_cols: Vec<Expr> = Vec::new();
    for (i, expr) in final_target.exprs.iter().enumerate() {
        match postpone[i] {
            true => postponable_cols.push(expr.clone()),
            false => tlist::add_column_to_pathtarget(
                &mut input_target,
                expr.clone(),
                tlist::get_pathtarget_sortgroupref(final_target, i),
            ),
        }
    }
    let mut postponable_vars: Vec<Expr> = Vec::new();
    for expr in &postponable_cols {
        collect_window_input_vars(root, expr, &mut postponable_vars);
    }
    tlist::add_new_columns_to_pathtarget(&mut input_target, postponable_vars);
    super::costsize::set_pathtarget_cost_width(root, &mut input_target);
    input_target
}

/// apply_scanjoin_target_to_paths makes the paths of the join of the query's relations compute the scan/join target,
/// adding projections where a path cannot compute it itself, and the set-returning functions of its levels, as
/// Postgres' function of the same name does.
fn apply_scanjoin_target_to_paths(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    scanjoin_targets: &[Rc<PathTarget>],
    scanjoin_targets_contain_srfs: &[bool],
    tlist_same_exprs: bool,
) {
    let scanjoin_target = scanjoin_targets[0].clone();
    let paths = std::mem::take(&mut root.rels[rel].pathlist);
    for path in paths {
        let path = match tlist_same_exprs {
            true => path,
            false => super::pathnode::create_projection_path(root, rel, path, scanjoin_target.clone()),
        };
        root.rels[rel].pathlist.push(path);
    }
    if root.parse.target_srfs.is_empty() || scanjoin_targets.len() <= 1 {
        root.rels[rel].reltarget = (*scanjoin_target).clone();
    }
    if !root.parse.target_srfs.is_empty() {
        let mut targets = (scanjoin_targets.to_vec(), scanjoin_targets_contain_srfs.to_vec());
        adjust_paths_for_srfs(root, rel, &mut targets);
    }
    set_cheapest(&mut root.rels[rel]);
}

/// adjust_paths_for_srfs puts the projections of a target's levels of set-returning functions above each of a
/// relation's paths, as Postgres' function of the same name does.
fn adjust_paths_for_srfs(root: &mut PlannerInfo<'_, '_>, rel: usize, targets: &mut (Vec<Rc<PathTarget>>, Vec<bool>)) {
    let (targets, contain_srfs) = targets;
    if targets.len() <= 1 {
        return;
    }
    let paths = std::mem::take(&mut root.rels[rel].pathlist);
    for path in paths {
        let mut newpath = path;
        for (target, &srfs) in targets.iter().zip(contain_srfs.iter()).skip(1) {
            newpath = match srfs {
                true => super::pathnode::create_set_projection_path(root, rel, newpath, target.clone()),
                false => super::pathnode::apply_projection_to_path(root, rel, newpath, target.clone()),
            };
        }
        root.rels[rel].pathlist.push(newpath);
    }
    set_cheapest(&mut root.rels[rel]);
}

/// get_cheapest_fractional_path returns the path of a relation that reads the rows a query reads most cheaply,
/// given as a count, or as a fraction below one, or zero for all of them, as Postgres' function of the same name
/// does.
pub fn get_cheapest_fractional_path(rel: &RelOptInfo, tuple_fraction: f64) -> Rc<Path> {
    let mut best_path = rel.cheapest_total_path.clone().expect("every relation has a path");
    if tuple_fraction <= 0.0 {
        return best_path;
    }
    let fraction = match tuple_fraction >= 1.0 && best_path.rows > 0.0 {
        true => tuple_fraction / best_path.rows,
        false => tuple_fraction,
    };
    for path in rel.pathlist.iter().filter(|p| p.param.is_empty()) {
        if !Rc::ptr_eq(path, &best_path) && compare_fractional_path_costs(&best_path, path, fraction).is_gt() {
            best_path = path.clone();
        }
    }
    best_path
}
