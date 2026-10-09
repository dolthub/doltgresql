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

//! The parts of Postgres' optimizer/plan/planner.c that choose the final path.

use std::rc::Rc;

use super::PlannerInfo;
use super::costsize::cost_sort;
use super::nodes::{Path, PathKind, RelOptInfo};
use super::pathkeys::pathkeys_contained_in;
use super::pathnode::{add_path, compare_fractional_path_costs, set_cheapest};

/// create_ordered_paths returns the path that reads the final relation's rows in the order of the query's ORDER BY
/// most cheaply, for the rows that the query reads, as Postgres' function of the same name and
/// get_cheapest_fractional_path choose it: a path already in that order, or the cheapest path sorted, of whose rows
/// a LIMIT may read only some. Without an ORDER BY, it is the cheapest path.
pub fn create_ordered_paths(root: &mut PlannerInfo<'_, '_>, final_rel: usize, limit_tuples: f64) -> Rc<Path> {
    let rel = &root.rels[final_rel];
    if root.query_pathkeys.is_empty() {
        return get_cheapest_fractional_path(rel, root.tuple_fraction);
    }
    let cheapest = rel.cheapest_total_path.clone().expect("every relation has a path");
    let mut ordered_rel = RelOptInfo { consider_startup: root.tuple_fraction > 0.0, ..RelOptInfo::default() };
    for path in rel.pathlist.iter().filter(|p| p.param.is_empty()) {
        if pathkeys_contained_in(&root.query_pathkeys, &path.pathkeys) {
            add_path(&mut ordered_rel, path.clone());
        } else if Rc::ptr_eq(path, &cheapest) {
            let (disabled_nodes, startup_cost, total_cost) = cost_sort(root, path, limit_tuples);
            let sorted = Path {
                kind: PathKind::Sort(path.clone()),
                pathkeys: root.query_pathkeys.clone(),
                disabled_nodes,
                startup_cost,
                total_cost,
                ..(**path).clone()
            };
            add_path(&mut ordered_rel, Rc::new(sorted));
        }
    }
    set_cheapest(&mut ordered_rel);
    get_cheapest_fractional_path(&ordered_rel, root.tuple_fraction)
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
