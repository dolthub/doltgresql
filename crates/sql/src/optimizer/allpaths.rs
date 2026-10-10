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

//! Postgres' optimizer/path/allpaths.c: finding the paths of the base relations, and of the relation that joins
//! them all.

use std::rc::Rc;

use super::PlannerInfo;
use super::costsize::{Costs, cost_opaque_scan, cost_resultscan, cost_seqscan, set_baserel_size_estimates};
use super::indxpath::{check_index_predicates, create_index_paths};
use super::initsplan::JoinList;
use super::joinrels::{is_dummy_rel, join_search_one_level};
use super::nodes::{JoinType, Path, PathKind, RelOptKind, RteKind};
use super::pathnode::{add_path, set_cheapest};
use crate::plan::Plan;

/// make_one_rel finds the paths of every base relation and then of the join of them all, returning that relation,
/// as Postgres' function of the same name does.
pub fn make_one_rel(root: &mut PlannerInfo<'_, '_>, joinlist: &[JoinList]) -> usize {
    set_base_rel_consider_startup(root);
    set_base_rel_sizes(root);
    root.total_table_pages = (1..=root.parse.rtable.len())
        .filter(|&rti| root.rels[rti].reloptkind == RelOptKind::BaseRel && !is_dummy_rel(root, rti))
        .map(|rti| root.rels[rti].pages)
        .sum();
    set_base_rel_pathlists(root);
    make_rel_from_joinlist(root, joinlist).expect("the join tree has a relation")
}

/// set_base_rel_consider_startup marks the one inner relation of each semi or anti join as worth paths that start
/// cheaply when parameterized, as Postgres' function of the same name does.
fn set_base_rel_consider_startup(root: &mut PlannerInfo<'_, '_>) {
    for sj in root.join_info_list.clone() {
        let sjinfo = &root.sjinfos[sj];
        if matches!(sjinfo.jointype, JoinType::Semi | JoinType::Anti)
            && let Some(varno) = sjinfo.syn_righthand.singleton_member()
        {
            root.rels[varno].consider_param_startup = true;
        }
    }
}

/// set_base_rel_sizes estimates the size of each base relation, as Postgres' function of the same name does.
fn set_base_rel_sizes(root: &mut PlannerInfo<'_, '_>) {
    for rti in 1..=root.parse.rtable.len() {
        if root.rels[rti].reloptkind == RelOptKind::BaseRel {
            set_rel_size(root, rti);
        }
    }
}

/// set_rel_size estimates the size of a base relation, or marks it empty when its restrictions refute it, as
/// Postgres' function of the same name does.
fn set_rel_size(root: &mut PlannerInfo<'_, '_>, rti: usize) {
    if super::plancat::relation_excluded_by_constraints(root, rti) {
        set_dummy_rel_pathlist(root, rti);
        return;
    }
    if matches!(root.parse.rte(rti).kind, RteKind::Relation(..)) {
        check_index_predicates(root, rti);
    }
    set_baserel_size_estimates(root, rti);
}

/// set_dummy_rel_pathlist marks a base relation as returning no rows, as Postgres' function of the same name does.
fn set_dummy_rel_pathlist(root: &mut PlannerInfo<'_, '_>, rti: usize) {
    root.rels[rti].reltarget.width = 0.0;
    super::joinrels::mark_dummy_rel(root, rti);
}

/// set_base_rel_pathlists finds the paths of each base relation, as Postgres' function of the same name does.
fn set_base_rel_pathlists(root: &mut PlannerInfo<'_, '_>) {
    for rti in 1..=root.parse.rtable.len() {
        if root.rels[rti].reloptkind == RelOptKind::BaseRel {
            set_rel_pathlist(root, rti);
        }
    }
}

/// set_rel_pathlist finds the paths of a base relation by its range table entry's kind, as Postgres' function of the
/// same name does: a table's sequential and index paths, the one row of a RESULT relation, or the scan of any other
/// input's plan.
fn set_rel_pathlist(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    if !is_dummy_rel(root, rel) {
        match &root.parse.rte(rel).kind {
            RteKind::Relation(..) => {
                let costs = cost_seqscan(root, rel);
                add_scan_path(root, rel, PathKind::SeqScan, costs);
                create_index_paths(root, rel);
            }
            RteKind::Result => {
                let costs = cost_resultscan(root, rel);
                add_scan_path(root, rel, PathKind::SeqScan, costs);
            }
            RteKind::Plan(plan) => {
                let catalog = matches!(plan, Plan::Catalog(_));
                let costs = cost_opaque_scan(root, rel, catalog);
                add_scan_path(root, rel, PathKind::SeqScan, costs);
                create_index_paths(root, rel);
            }
            RteKind::Subquery(..) | RteKind::Join(_) => unreachable!("only base relations have paths"),
        }
    }
    set_cheapest(&mut root.rels[rel]);
}

/// add_scan_path adds a path of a base relation that reads its rows, as Postgres' create_seqscan_path,
/// create_resultscan_path, and create_functionscan_path make.
fn add_scan_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    kind: PathKind,
    (disabled_nodes, startup_cost, total_cost): Costs,
) {
    let parent = &root.rels[rel];
    let path = Path {
        kind,
        parent: rel,
        relids: parent.relids.clone(),
        param: parent.lateral_relids.clone(),
        pathkeys: Vec::new(),
        rows: parent.rows,
        width: parent.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
        pathtarget: None,
    };
    add_path(&mut root.rels[rel], Rc::new(path));
}

/// make_rel_from_joinlist returns the relation that joins a joinlist's members, searching for its cheapest paths
/// when it has several, as Postgres' function of the same name does.
fn make_rel_from_joinlist(root: &mut PlannerInfo<'_, '_>, joinlist: &[JoinList]) -> Option<usize> {
    let mut initial_rels = Vec::new();
    for item in joinlist {
        initial_rels.push(match item {
            JoinList::Rel(varno) => *varno,
            JoinList::List(list) => make_rel_from_joinlist(root, list)?,
        });
    }
    match initial_rels.as_slice() {
        [] => None,
        [rel] => Some(*rel),
        _ => {
            root.initial_rels = initial_rels.clone();
            Some(standard_join_search(root, initial_rels))
        }
    }
}

/// standard_join_search finds the join relations of each number of relations in turn, from pairs to all of them,
/// keeping each one's cheapest paths, as Postgres' function of the same name does.
fn standard_join_search(root: &mut PlannerInfo<'_, '_>, initial_rels: Vec<usize>) -> usize {
    let levels_needed = initial_rels.len();
    root.join_rel_level = vec![Vec::new(), initial_rels];
    for level in 2..=levels_needed {
        join_search_one_level(root, level);
        for rel in root.join_rel_level[level].clone() {
            set_cheapest(&mut root.rels[rel]);
        }
    }
    let rel = *root.join_rel_level[levels_needed].first().expect("the join search joins every relation");
    root.join_rel_level.clear();
    rel
}
