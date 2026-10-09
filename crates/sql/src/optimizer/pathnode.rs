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

//! Postgres' optimizer/util/pathnode.c: keeping the paths of a relation that may be part of the cheapest plan, and
//! making paths of each kind.

use std::cmp::Ordering;
use std::rc::Rc;

use super::nodes::{JoinPath, Path, RelOptInfo, Relids, is_subset};

/// STD_FUZZ_FACTOR is how much cheaper one path must be than another to count as cheaper, as Postgres' constant of
/// the same name is.
const STD_FUZZ_FACTOR: f64 = 1.01;

/// compare_path_costs_fuzzily compares two paths' costs, where costs within a factor count as equal, and startup
/// costs only break ties, as they do in Postgres when no LIMIT asks for the first rows.
fn compare_path_costs_fuzzily(path1: &Path, path2: &Path, fuzz_factor: f64) -> Ordering {
    if path1.total_cost > path2.total_cost * fuzz_factor {
        return Ordering::Greater;
    }
    if path2.total_cost > path1.total_cost * fuzz_factor {
        return Ordering::Less;
    }
    if path1.startup_cost > path2.startup_cost * fuzz_factor {
        return Ordering::Greater;
    }
    if path2.startup_cost > path1.startup_cost * fuzz_factor {
        return Ordering::Less;
    }
    Ordering::Equal
}

/// SubsetCompare is how two sets of relations compare, as Postgres' BMS_Comparison is.
#[derive(PartialEq)]
enum SubsetCompare {
    Equal,
    Subset1,
    Subset2,
    Different,
}

/// bms_subset_compare compares two sets of relations by inclusion.
fn bms_subset_compare(a: Relids, b: Relids) -> SubsetCompare {
    match (is_subset(a, b), is_subset(b, a)) {
        (true, true) => SubsetCompare::Equal,
        (true, false) => SubsetCompare::Subset1,
        (false, true) => SubsetCompare::Subset2,
        (false, false) => SubsetCompare::Different,
    }
}

/// add_path adds a path to a relation's paths unless one of them is as cheap, needs no more outer relations, and
/// produces no more rows, removing those that the new path beats the same way, as Postgres' add_path does for
/// paths without sort orders. The paths stay in order of total cost.
pub fn add_path(parent: &mut RelOptInfo, new_path: Rc<Path>) {
    let mut accept_new = true;
    let mut insert_at = 0;
    let mut i = 0;
    while i < parent.pathlist.len() {
        let old_path = &parent.pathlist[i];
        let mut remove_old = false;
        let outercmp = bms_subset_compare(new_path.param, old_path.param);
        match compare_path_costs_fuzzily(&new_path, old_path, STD_FUZZ_FACTOR) {
            Ordering::Equal => match outercmp {
                SubsetCompare::Equal => {
                    if new_path.rows < old_path.rows {
                        remove_old = true;
                    } else if new_path.rows > old_path.rows {
                        accept_new = false;
                    } else if compare_path_costs_fuzzily(&new_path, old_path, 1.0000000001) == Ordering::Less {
                        remove_old = true;
                    } else {
                        accept_new = false;
                    }
                }
                SubsetCompare::Subset1 if new_path.rows <= old_path.rows => remove_old = true,
                SubsetCompare::Subset2 if new_path.rows >= old_path.rows => accept_new = false,
                _ => {}
            },
            Ordering::Less => {
                if matches!(outercmp, SubsetCompare::Equal | SubsetCompare::Subset1) && new_path.rows <= old_path.rows {
                    remove_old = true;
                }
            }
            Ordering::Greater => {
                if matches!(outercmp, SubsetCompare::Equal | SubsetCompare::Subset2) && new_path.rows >= old_path.rows {
                    accept_new = false;
                }
            }
        }
        if remove_old {
            parent.pathlist.remove(i);
        } else {
            if new_path.total_cost >= parent.pathlist[i].total_cost {
                insert_at = i + 1;
            }
            i += 1;
        }
        if !accept_new {
            return;
        }
    }
    parent.pathlist.insert(insert_at, new_path);
}

/// set_cheapest finds a relation's cheapest paths: the cheapest total and startup costs among those that need no
/// outer relation, and the cheapest path of each set of outer relations, as Postgres' set_cheapest does.
pub fn set_cheapest(parent: &mut RelOptInfo) {
    let unparameterized = parent.pathlist.iter().filter(|p| p.param == 0);
    let cheapest_total = unparameterized.clone().min_by(|a, b| a.total_cost.total_cmp(&b.total_cost)).cloned();
    parent.cheapest_startup_path = unparameterized.min_by(|a, b| a.startup_cost.total_cmp(&b.startup_cost)).cloned();
    let mut parameterized: Vec<Rc<Path>> = cheapest_total.iter().cloned().collect();
    for path in parent.pathlist.iter().filter(|p| p.param != 0) {
        match parameterized.iter_mut().find(|p| p.param == path.param) {
            Some(best) if path.total_cost < best.total_cost => *best = path.clone(),
            Some(_) => {}
            None => parameterized.push(path.clone()),
        }
    }
    parent.cheapest_total_path = cheapest_total;
    parent.cheapest_parameterized_paths = parameterized;
}

/// create_join_path makes a path of a join of two paths, as Postgres' create_nestloop_path and create_hashjoin_path
/// do, with the join relation's row estimate and the costs given.
pub fn create_join_path(
    joinrel: &RelOptInfo,
    kind: fn(JoinPath) -> super::nodes::PathKind,
    join: JoinPath,
    (startup_cost, total_cost): (f64, f64),
) -> Rc<Path> {
    let param = (join.outer.param | join.inner.param) & !joinrel.relids;
    Rc::new(Path {
        kind: kind(join),
        relids: joinrel.relids,
        param,
        rows: joinrel.rows,
        width: joinrel.width,
        startup_cost,
        total_cost,
    })
}
