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

//! Postgres' optimizer/prep/prepunion.c: the paths of a query's tree of UNION, INTERSECT, and EXCEPT operations,
//! whose leaves are subqueries, and of a recursive WITH query's union. A set operation's columns are Vars of no
//! relation, numbered by position, as Postgres' generate_append_tlist makes them.

use std::rc::Rc;

use super::PlannerInfo;
use super::nodes::{
    AggStrategy, PkId, Relids, SetOpTree, SetOperationStmt, SortGroupClause, TargetEntry, UpperRelationKind,
};
use super::pathnode::{add_path, create_recursiveunion_path, create_setop_path, set_cheapest};
use crate::plan::SetOp;

/// plan_set_operations returns the relation of the paths of the query's tree of set operations, setting the query's
/// processed target list to its columns, as Postgres' function of the same name does.
pub fn plan_set_operations(root: &mut PlannerInfo<'_, '_>) -> usize {
    let topop = root.parse.set_operations.clone().expect("a set operation");
    root.ec_merging_done = true;
    super::relnode::setup_simple_rel_arrays(root);
    let (setop_rel, top_tlist) = match root.parse.recursion {
        Some(wt_param_id) => generate_recursion_path(root, &topop, wt_param_id),
        None => {
            let (rel, tlist, _) = recurse_set_operations(root, &SetOpTree::Op(topop.clone()), &topop.col_types);
            (rel, tlist)
        }
    };
    root.processed_tlist = top_tlist;
    setop_rel
}

/// recurse_set_operations returns the relation of a step of a tree of set operations, with its target list and
/// whether that list is its input's columns as they are, as Postgres' function of the same name does: a leaf is its
/// subquery's relation, planned but without paths yet, and an operation has the paths of its rows.
fn recurse_set_operations(
    root: &mut PlannerInfo<'_, '_>,
    set_op: &SetOpTree,
    col_types: &[Option<u32>],
) -> (usize, Vec<TargetEntry>, bool) {
    match set_op {
        SetOpTree::Rel(rti) => {
            let rti = *rti;
            super::relnode::build_simple_rel(root, rti);
            let super::nodes::RteKind::Subquery(subquery, _) = &root.parse.rte(rti).kind else {
                unreachable!("a leaf of a set operation is a subquery")
            };
            let subquery = (**subquery).clone();
            let tuple_fraction = root.tuple_fraction;
            if !super::allpaths::plan_subquery_rel(root, rti, subquery, tuple_fraction) {
                super::joinrels::mark_dummy_rel(root, rti);
            }
            let tlist = generate_setop_tlist(root, col_types, rti);
            root.rels[rti].reltarget = super::tlist::create_pathtarget(root, &tlist);
            (rti, tlist, true)
        }
        SetOpTree::Op(op) => {
            let (rel, tlist) = match op.op {
                SetOp::Union => generate_union_paths(root, op),
                _ => generate_nonunion_paths(root, op),
            };
            postprocess_setop_rel(root, rel);
            (rel, tlist, true)
        }
    }
}

/// generate_recursion_path returns the relation of a recursive WITH query's union of its non-recursive and
/// recursive terms, with its target list, as Postgres' function of the same name does.
fn generate_recursion_path(
    root: &mut PlannerInfo<'_, '_>,
    set_op: &SetOperationStmt,
    wt_param_id: usize,
) -> (usize, Vec<TargetEntry>) {
    let (lrel, lpath_tlist, lpath_trivial_tlist) = recurse_set_operations(root, &set_op.larg, &set_op.col_types);
    build_setop_child_paths(root, lrel, lpath_trivial_tlist, &lpath_tlist, &[]);
    let lpath = root.rels[lrel].cheapest_total_path.clone().expect("a path of the non-recursive term");
    root.glob.non_recursive_rows = Some(lpath.rows);
    let (rrel, rpath_tlist, rpath_trivial_tlist) = recurse_set_operations(root, &set_op.rarg, &set_op.col_types);
    build_setop_child_paths(root, rrel, rpath_trivial_tlist, &rpath_tlist, &[]);
    let rpath = root.rels[rrel].cheapest_total_path.clone().expect("a path of the recursive term");
    root.glob.non_recursive_rows = None;
    let tlist = generate_append_tlist(root, &set_op.col_types);
    let relids = root.rels[lrel].relids.union(&root.rels[rrel].relids);
    let result_rel = super::relnode::fetch_upper_rel(root, UpperRelationKind::SetOp, &relids);
    root.rels[result_rel].reltarget = super::tlist::create_pathtarget(root, &tlist);
    let path = create_recursiveunion_path(root, result_rel, lpath, rpath, !set_op.all, wt_param_id);
    add_path(&mut root.rels[result_rel], path);
    postprocess_setop_rel(root, result_rel);
    (result_rel, tlist)
}

/// build_setop_child_paths adds the paths of a set operation's leaf, a subquery relation, as Postgres' function of
/// the same name does: a scan of its subquery's cheapest path, and when the operation wants its rows in an order,
/// that scan or a scan of each of its other paths sorted in it, returning the estimated number of distinct rows.
fn build_setop_child_paths(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    trivial_tlist: bool,
    child_tlist: &[TargetEntry],
    interesting_pathkeys: &[PkId],
) -> f64 {
    if root.rels[rel].reloptkind != super::nodes::RelOptKind::BaseRel {
        return root.rels[rel].rows;
    }
    if super::joinrels::is_dummy_rel(root, rel) {
        set_cheapest(&mut root.rels[rel]);
        return 0.0;
    }
    if !interesting_pathkeys.is_empty() {
        super::equivclass::add_setop_child_rel_equivalences(root, rel, child_tlist, interesting_pathkeys);
    }
    let cheapest = (0..root.rels[rel].subplans.len())
        .min_by(|&a, &b| {
            let (pa, pb) = (&root.rels[rel].subplans[a].path, &root.rels[rel].subplans[b].path);
            super::pathnode::compare_path_costs(pa, pb, super::pathnode::CostSelector::Total)
        })
        .expect("a subquery path");
    for i in 0..root.rels[rel].subplans.len() {
        let order = root.rels[rel].subplans[i].order.clone();
        let pathkeys = super::pathkeys::convert_subquery_pathkeys(root, rel, &order);
        let scan = super::pathnode::create_subqueryscan_path(root, rel, i, trivial_tlist, pathkeys);
        if i == cheapest {
            add_path(&mut root.rels[rel], scan.clone());
        }
        if interesting_pathkeys.is_empty() {
            continue;
        }
        let (is_sorted, presorted_keys) =
            super::pathkeys::pathkeys_count_contained_in(interesting_pathkeys, &scan.pathkeys);
        if is_sorted {
            if i != cheapest {
                add_path(&mut root.rels[rel], scan);
            }
            continue;
        }
        if i != cheapest && (presorted_keys == 0 || !root.enables.incremental_sort) {
            continue;
        }
        let sorted = match presorted_keys == 0 || !root.enables.incremental_sort {
            true => super::pathnode::create_sort_path(root, rel, scan, interesting_pathkeys.to_vec(), -1.0),
            false => super::pathnode::create_incremental_sort_path(
                root,
                rel,
                scan,
                interesting_pathkeys.to_vec(),
                presorted_keys,
                -1.0,
            ),
        };
        add_path(&mut root.rels[rel], sorted);
    }
    set_cheapest(&mut root.rels[rel]);
    root.rels[rel].subquery_groups
}

/// generate_union_paths returns the relation of a UNION, with its target list, as Postgres' function of the same
/// name does: an Append of its children's cheapest paths, and for a UNION without ALL, that Append hashed or sorted
/// to drop duplicates, or a MergeAppend of its children's sorted paths without duplicates.
fn generate_union_paths(root: &mut PlannerInfo<'_, '_>, op: &SetOperationStmt) -> (usize, Vec<TargetEntry>) {
    let children = plan_union_children(root, op);
    let tlist = generate_append_tlist(root, &op.col_types);
    let mut group_list = Vec::new();
    let mut try_sorted = false;
    let mut union_pathkeys = Vec::new();
    if !op.all {
        group_list = generate_setop_grouplist(&tlist);
        if super::tlist::grouping_is_sortable(&group_list)
            && let Some(pathkeys) = super::planner::make_pathkeys_for_sortclauses(root, &group_list, &tlist)
        {
            try_sorted = true;
            union_pathkeys = pathkeys;
            root.query_pathkeys = union_pathkeys.clone();
        }
    }
    for (rel, child_tlist, trivial_tlist) in &children {
        build_setop_child_paths(root, *rel, *trivial_tlist, child_tlist, &union_pathkeys);
    }
    let (mut cheapest_pathlist, mut ordered_pathlist) = (Vec::new(), Vec::new());
    let mut relids = Relids::new();
    for (rel, _, _) in &children {
        cheapest_pathlist.push(root.rels[*rel].cheapest_total_path.clone().expect("a path of a union's child"));
        if try_sorted {
            match super::pathkeys::get_cheapest_path_for_pathkeys(&root.rels[*rel].pathlist, &union_pathkeys) {
                Some(path) => ordered_pathlist.push(path),
                None => try_sorted = false,
            }
        }
        relids = relids.union(&root.rels[*rel].relids);
    }
    let result_rel = super::relnode::fetch_upper_rel(root, UpperRelationKind::SetOp, &relids);
    root.rels[result_rel].reltarget = super::tlist::create_pathtarget(root, &tlist);
    root.rels[result_rel].consider_startup = root.tuple_fraction > 0.0;
    let apath = super::pathnode::create_append_path(root, result_rel, cheapest_pathlist);
    root.rels[result_rel].rows = apath.rows;
    if op.all {
        add_path(&mut root.rels[result_rel], apath);
        return (result_rel, tlist);
    }
    let d_num_groups = apath.rows;
    if super::tlist::grouping_is_hashable(&group_list) {
        let target = Rc::new(super::tlist::create_pathtarget(root, &tlist));
        let path = super::pathnode::create_agg_path(
            root,
            result_rel,
            apath.clone(),
            target,
            AggStrategy::Hashed,
            group_list.clone(),
            Vec::new(),
            &super::prepagg::AggClauseCosts::default(),
            d_num_groups,
        );
        add_path(&mut root.rels[result_rel], path);
    }
    if super::tlist::grouping_is_sortable(&group_list)
        && let Some(pathkeys) = super::planner::make_pathkeys_for_sortclauses(root, &group_list, &tlist)
    {
        let sorted = super::pathnode::create_sort_path(root, result_rel, apath, pathkeys, -1.0);
        let num_cols = sorted.pathkeys.len();
        let path = super::pathnode::create_upper_unique_path(root, result_rel, sorted, num_cols, d_num_groups);
        add_path(&mut root.rels[result_rel], path);
    }
    if try_sorted && !group_list.is_empty() {
        let path = super::pathnode::create_merge_append_path(root, result_rel, ordered_pathlist, union_pathkeys);
        let path = super::pathnode::create_upper_unique_path(root, result_rel, path, tlist.len(), d_num_groups);
        add_path(&mut root.rels[result_rel], path);
    }
    (result_rel, tlist)
}

/// generate_nonunion_paths returns the relation of an INTERSECT or EXCEPT, with its target list, as Postgres'
/// function of the same name does: a SetOp over its two inputs, hashed, or merged with both sorted, putting the
/// input of fewer groups first for an INTERSECT.
fn generate_nonunion_paths(root: &mut PlannerInfo<'_, '_>, op: &SetOperationStmt) -> (usize, Vec<TargetEntry>) {
    let save_fraction = root.tuple_fraction;
    root.tuple_fraction = 0.0;
    let (mut lrel, mut lpath_tlist, lpath_trivial_tlist) = recurse_set_operations(root, &op.larg, &op.col_types);
    let (mut rrel, mut rpath_tlist, rpath_trivial_tlist) = recurse_set_operations(root, &op.rarg, &op.col_types);
    let tlist = generate_setop_tlist(root, &op.col_types, 0);
    let group_list = generate_setop_grouplist(&tlist);
    let can_sort = super::tlist::grouping_is_sortable(&group_list);
    let can_hash = super::tlist::grouping_is_hashable(&group_list);
    let mut nonunion_pathkeys = Vec::new();
    if can_sort && let Some(pathkeys) = super::planner::make_pathkeys_for_sortclauses(root, &group_list, &tlist) {
        nonunion_pathkeys = pathkeys;
        root.query_pathkeys = nonunion_pathkeys.clone();
    }
    let mut d_left_groups = build_setop_child_paths(root, lrel, lpath_trivial_tlist, &lpath_tlist, &nonunion_pathkeys);
    let mut d_right_groups = build_setop_child_paths(root, rrel, rpath_trivial_tlist, &rpath_tlist, &nonunion_pathkeys);
    root.tuple_fraction = save_fraction;
    if op.op != SetOp::Except && d_left_groups > d_right_groups {
        std::mem::swap(&mut lrel, &mut rrel);
        std::mem::swap(&mut lpath_tlist, &mut rpath_tlist);
        std::mem::swap(&mut d_left_groups, &mut d_right_groups);
    }
    let lpath = root.rels[lrel].cheapest_total_path.clone().expect("a path of the left input");
    let rpath = root.rels[rrel].cheapest_total_path.clone().expect("a path of the right input");
    let relids = root.rels[lrel].relids.union(&root.rels[rrel].relids);
    let result_rel = super::relnode::fetch_upper_rel(root, UpperRelationKind::SetOp, &relids);
    root.rels[result_rel].reltarget = super::tlist::create_pathtarget(root, &tlist);
    let d_num_groups = d_left_groups;
    let d_num_output_rows = match (op.op, op.all) {
        (SetOp::Except, true) => lpath.rows,
        (_, true) => lpath.rows.min(rpath.rows),
        (_, false) => d_num_groups,
    };
    root.rels[result_rel].rows = d_num_output_rows;
    if can_hash {
        let inputs = (lpath.clone(), rpath.clone());
        let estimates = (d_num_groups, d_num_output_rows);
        let path = create_setop_path(root, result_rel, inputs, (op.op, op.all), true, group_list.clone(), estimates);
        add_path(&mut root.rels[result_rel], path);
    }
    if can_sort {
        let mut sorted = Vec::new();
        for (rel, path, child_tlist) in [(lrel, &lpath, &lpath_tlist), (rrel, &rpath, &rpath_tlist)] {
            let pathkeys =
                super::planner::make_pathkeys_for_sortclauses(root, &group_list, child_tlist).unwrap_or_default();
            sorted.push(match super::pathkeys::pathkeys_contained_in(&pathkeys, &path.pathkeys) {
                true => path.clone(),
                false => {
                    match super::pathkeys::get_cheapest_path_for_pathkeys(&root.rels[rel].pathlist, &nonunion_pathkeys)
                    {
                        Some(sorted) => sorted,
                        None => super::pathnode::create_sort_path(root, rel, path.clone(), pathkeys, -1.0),
                    }
                }
            });
        }
        let (slpath, srpath) = (sorted[0].clone(), sorted[1].clone());
        let estimates = (d_num_groups, d_num_output_rows);
        let path = create_setop_path(root, result_rel, (slpath, srpath), (op.op, op.all), false, group_list, estimates);
        add_path(&mut root.rels[result_rel], path);
    }
    (result_rel, tlist)
}

/// plan_union_children returns the relations of the inputs of a UNION, with their target lists, folding the inputs
/// of each UNION below it of the same kind into its own, as Postgres' function of the same name does.
fn plan_union_children(
    root: &mut PlannerInfo<'_, '_>,
    top_union: &SetOperationStmt,
) -> Vec<(usize, Vec<TargetEntry>, bool)> {
    let mut pending_rels = vec![top_union.larg.clone(), top_union.rarg.clone()];
    let mut result = Vec::new();
    while !pending_rels.is_empty() {
        let set_op = pending_rels.remove(0);
        if let SetOpTree::Op(op) = &set_op
            && op.op == top_union.op
            && (op.all == top_union.all || op.all)
            && op.col_types == top_union.col_types
        {
            pending_rels.insert(0, op.rarg.clone());
            pending_rels.insert(0, op.larg.clone());
            continue;
        }
        result.push(recurse_set_operations(root, &set_op, &top_union.col_types));
    }
    result
}

/// postprocess_setop_rel picks the cheapest paths of a set operation's relation, as Postgres' function of the same
/// name does.
fn postprocess_setop_rel(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    set_cheapest(&mut root.rels[rel]);
}

/// generate_setop_tlist returns the target list of a set operation's step over a relation's columns, or over the
/// set operation's own columns for relation 0, each its own sort and group key by position, as Postgres' function of
/// the same name does. The binder coerces each input to the set operation's column types already.
fn generate_setop_tlist(root: &mut PlannerInfo<'_, '_>, col_types: &[Option<u32>], varno: usize) -> Vec<TargetEntry> {
    (0..col_types.len())
        .map(|i| TargetEntry { expr: root.glob.var(varno, i, Relids::new()), resjunk: false, ressortgroupref: i + 1 })
        .collect()
}

/// generate_append_tlist returns the target list of an Append of a set operation's inputs: the set operation's own
/// columns, as Postgres' function of the same name does.
fn generate_append_tlist(root: &mut PlannerInfo<'_, '_>, col_types: &[Option<u32>]) -> Vec<TargetEntry> {
    generate_setop_tlist(root, col_types, 0)
}

/// generate_setop_grouplist returns the group clauses of a set operation, one for each column of its target list,
/// as Postgres' function of the same name does.
fn generate_setop_grouplist(targetlist: &[TargetEntry]) -> Vec<SortGroupClause> {
    targetlist
        .iter()
        .filter(|tle| !tle.resjunk)
        .map(|tle| SortGroupClause {
            tle_sort_group_ref: tle.ressortgroupref,
            descending: false,
            nulls_first: false,
            hashable: true,
        })
        .collect()
}
