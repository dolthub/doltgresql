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
use super::costsize::{cost_opaque_scan, cost_seqscan, set_baserel_size_estimates};
use super::indxpath::create_index_paths;
use super::initsplan::JoinList;
use super::joinrels::join_search_one_level;
use super::nodes::{Path, PathKind, members};
use super::pathnode::{add_path, set_cheapest};

/// make_one_rel finds the paths of every base relation and then of the join of them all, returning that relation,
/// as Postgres' function of the same name does.
pub fn make_one_rel(root: &mut PlannerInfo<'_, '_>, joinlist: Vec<JoinList>) -> usize {
    let baserels: Vec<usize> = members(root.all_baserels).collect();
    root.total_table_pages = baserels.iter().map(|&rel| root.rels[rel].pages).sum();
    for &rel in &baserels {
        set_baserel_size_estimates(root, rel);
    }
    for &rel in &baserels {
        set_rel_pathlist(root, rel);
    }
    make_rel_from_joinlist(root, &joinlist)
}

/// set_rel_pathlist finds the paths of a base relation: reading every row, and its index paths, as Postgres'
/// set_plain_rel_pathlist does.
fn set_rel_pathlist(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    let parent = &root.rels[rel];
    let (startup_cost, total_cost) = match root.parse.rte(rel).table() {
        Some(_) => cost_seqscan(parent, root.enables),
        None => cost_opaque_scan(parent),
    };
    let path = Path {
        kind: PathKind::SeqScan,
        relids: parent.relids,
        param: 0,
        rows: parent.rows,
        width: parent.width,
        startup_cost,
        total_cost,
    };
    add_path(&mut root.rels[rel], Rc::new(path));
    create_index_paths(root, rel);
    set_cheapest(&mut root.rels[rel]);
}

/// make_rel_from_joinlist returns the relation that joins a joinlist's members, searching for its cheapest paths
/// when it has several, as Postgres' function of the same name does.
fn make_rel_from_joinlist(root: &mut PlannerInfo<'_, '_>, joinlist: &[JoinList]) -> usize {
    let initial_rels: Vec<usize> = joinlist
        .iter()
        .map(|item| match item {
            JoinList::Rel(varno) => *varno,
            JoinList::List(list) => make_rel_from_joinlist(root, list),
        })
        .collect();
    if let [rel] = initial_rels.as_slice() {
        return *rel;
    }
    root.initial_rels = initial_rels.clone();
    standard_join_search(root, initial_rels)
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
