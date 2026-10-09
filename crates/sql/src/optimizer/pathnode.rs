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
//! making paths of each kind. Doltgres has no parallel workers, so no path is parallel safe.

use std::cmp::Ordering;
use std::rc::Rc;

use super::costsize::Costs;
use super::nodes::{JoinPath, Path, PathKind, PkId, RelOptInfo, SubsetCompare};
use super::pathkeys::{PathKeysComparison, compare_pathkeys};

/// STD_FUZZ_FACTOR is how much cheaper one path must be than another to count as cheaper, as Postgres' constant of
/// the same name is.
const STD_FUZZ_FACTOR: f64 = 1.01;

/// CostSelector is which cost compare_path_costs compares first, as Postgres' CostSelector is.
#[derive(Clone, Copy, PartialEq, Eq)]
pub enum CostSelector {
    Startup,
    Total,
}

/// compare_path_costs compares two paths by their disabled plan nodes and then by one cost, breaking ties by the
/// other, as Postgres' function of the same name does.
pub fn compare_path_costs(path1: &Path, path2: &Path, criterion: CostSelector) -> Ordering {
    if path1.disabled_nodes != path2.disabled_nodes {
        return path1.disabled_nodes.cmp(&path2.disabled_nodes);
    }
    let (startup, total) =
        (path1.startup_cost.total_cmp(&path2.startup_cost), path1.total_cost.total_cmp(&path2.total_cost));
    match criterion {
        CostSelector::Startup => startup.then(total),
        CostSelector::Total => total.then(startup),
    }
}

/// compare_fractional_path_costs compares the costs of two paths reading a fraction of their rows, as Postgres'
/// function of the same name does.
pub fn compare_fractional_path_costs(path1: &Path, path2: &Path, fraction: f64) -> Ordering {
    if path1.disabled_nodes != path2.disabled_nodes {
        return path1.disabled_nodes.cmp(&path2.disabled_nodes);
    }
    if fraction <= 0.0 || fraction >= 1.0 {
        return compare_path_costs(path1, path2, CostSelector::Total);
    }
    let cost1 = path1.startup_cost + fraction * (path1.total_cost - path1.startup_cost);
    let cost2 = path2.startup_cost + fraction * (path2.total_cost - path2.startup_cost);
    cost1.total_cmp(&cost2)
}

/// PathCostComparison is how two paths' costs compare, as Postgres' PathCostComparison is.
#[derive(Clone, Copy, PartialEq, Eq)]
enum PathCostComparison {
    Equal,
    Better1,
    Better2,
    Different,
}

/// consider_path_startup_cost reports whether a path's relation keeps paths that start more cheaply, by whether
/// the path is parameterized, as Postgres' CONSIDER_PATH_STARTUP_COST does.
fn consider_path_startup_cost(rel: &RelOptInfo, path: &Path) -> bool {
    match path.param.is_empty() {
        true => rel.consider_startup,
        false => rel.consider_param_startup,
    }
}

/// compare_path_costs_fuzzily compares two paths' disabled plan nodes and then their costs, where costs within a
/// factor count as equal, and a path cheaper only to start is neither better nor worse unless its relation ignores
/// startup costs, as Postgres' function of the same name does.
fn compare_path_costs_fuzzily(rel: &RelOptInfo, path1: &Path, path2: &Path, fuzz_factor: f64) -> PathCostComparison {
    if path1.disabled_nodes != path2.disabled_nodes {
        return match path1.disabled_nodes < path2.disabled_nodes {
            true => PathCostComparison::Better1,
            false => PathCostComparison::Better2,
        };
    }
    if path1.total_cost > path2.total_cost * fuzz_factor {
        if consider_path_startup_cost(rel, path1) && path2.startup_cost > path1.startup_cost * fuzz_factor {
            return PathCostComparison::Different;
        }
        return PathCostComparison::Better2;
    }
    if path2.total_cost > path1.total_cost * fuzz_factor {
        if consider_path_startup_cost(rel, path2) && path1.startup_cost > path2.startup_cost * fuzz_factor {
            return PathCostComparison::Different;
        }
        return PathCostComparison::Better1;
    }
    if path1.startup_cost > path2.startup_cost * fuzz_factor {
        return PathCostComparison::Better2;
    }
    if path2.startup_cost > path1.startup_cost * fuzz_factor {
        return PathCostComparison::Better1;
    }
    PathCostComparison::Equal
}

/// set_cheapest finds a relation's cheapest paths: the cheapest total and startup costs among those that need no
/// outer relation, and the paths of each parameterization, as Postgres' set_cheapest does.
pub fn set_cheapest(parent_rel: &mut RelOptInfo) {
    assert!(!parent_rel.pathlist.is_empty(), "could not devise a query plan for the given query");
    let mut cheapest_startup_path: Option<Rc<Path>> = None;
    let mut cheapest_total_path: Option<Rc<Path>> = None;
    let mut best_param_path: Option<Rc<Path>> = None;
    let mut parameterized_paths = Vec::new();
    for path in &parent_rel.pathlist {
        if !path.param.is_empty() {
            parameterized_paths.push(path.clone());
            if cheapest_total_path.is_some() {
                continue;
            }
            match &best_param_path {
                None => best_param_path = Some(path.clone()),
                Some(best) => match path.param.subset_compare(&best.param) {
                    SubsetCompare::Equal if compare_path_costs(path, best, CostSelector::Total).is_lt() => {
                        best_param_path = Some(path.clone())
                    }
                    SubsetCompare::Subset1 => best_param_path = Some(path.clone()),
                    _ => {}
                },
            }
            continue;
        }
        let (Some(startup), Some(total)) = (&cheapest_startup_path, &cheapest_total_path) else {
            (cheapest_startup_path, cheapest_total_path) = (Some(path.clone()), Some(path.clone()));
            continue;
        };
        let better = |cheapest: &Path, criterion| match compare_path_costs(cheapest, path, criterion) {
            Ordering::Greater => true,
            Ordering::Equal => compare_pathkeys(&cheapest.pathkeys, &path.pathkeys) == PathKeysComparison::Better2,
            Ordering::Less => false,
        };
        if better(startup, CostSelector::Startup) {
            cheapest_startup_path = Some(path.clone());
        }
        if better(total, CostSelector::Total) {
            cheapest_total_path = Some(path.clone());
        }
    }
    if let Some(total) = &cheapest_total_path {
        parameterized_paths.insert(0, total.clone());
    }
    parent_rel.cheapest_startup_path = cheapest_startup_path;
    parent_rel.cheapest_total_path = cheapest_total_path.or(best_param_path);
    parent_rel.cheapest_parameterized_paths = parameterized_paths;
}

/// add_path adds a path to a relation's paths unless one of them is as cheap, as well ordered, needs no more outer
/// relations, and produces no more rows, removing those that the new path beats the same way, as Postgres' add_path
/// does. The paths stay in order of disabled plan nodes and then total cost.
pub fn add_path(parent_rel: &mut RelOptInfo, new_path: Rc<Path>) {
    let mut accept_new = true;
    let mut insert_at = 0;
    let no_keys: &[PkId] = &[];
    let new_path_pathkeys = if new_path.param.is_empty() { new_path.pathkeys.as_slice() } else { no_keys };
    let mut i = 0;
    while i < parent_rel.pathlist.len() {
        let old_path = parent_rel.pathlist[i].clone();
        let mut remove_old = false;
        let costcmp = compare_path_costs_fuzzily(parent_rel, &new_path, &old_path, STD_FUZZ_FACTOR);
        if costcmp != PathCostComparison::Different {
            let old_path_pathkeys = if old_path.param.is_empty() { old_path.pathkeys.as_slice() } else { no_keys };
            let keyscmp = compare_pathkeys(new_path_pathkeys, old_path_pathkeys);
            if keyscmp != PathKeysComparison::Different {
                let outercmp = || new_path.param.subset_compare(&old_path.param);
                let new_dominates = |outercmp: SubsetCompare| {
                    matches!(outercmp, SubsetCompare::Equal | SubsetCompare::Subset1) && new_path.rows <= old_path.rows
                };
                let old_dominates = |outercmp: SubsetCompare| {
                    matches!(outercmp, SubsetCompare::Equal | SubsetCompare::Subset2) && new_path.rows >= old_path.rows
                };
                match costcmp {
                    PathCostComparison::Equal => match keyscmp {
                        PathKeysComparison::Better1 => remove_old = new_dominates(outercmp()),
                        PathKeysComparison::Better2 => accept_new = !old_dominates(outercmp()),
                        _ => match outercmp() {
                            SubsetCompare::Equal => {
                                if new_path.rows < old_path.rows {
                                    remove_old = true;
                                } else if new_path.rows > old_path.rows {
                                    accept_new = false;
                                } else if compare_path_costs_fuzzily(parent_rel, &new_path, &old_path, 1.0000000001)
                                    == PathCostComparison::Better1
                                {
                                    remove_old = true;
                                } else {
                                    accept_new = false;
                                }
                            }
                            SubsetCompare::Subset1 if new_path.rows <= old_path.rows => remove_old = true,
                            SubsetCompare::Subset2 if new_path.rows >= old_path.rows => accept_new = false,
                            _ => {}
                        },
                    },
                    PathCostComparison::Better1 if keyscmp != PathKeysComparison::Better2 => {
                        remove_old = new_dominates(outercmp())
                    }
                    PathCostComparison::Better2 if keyscmp != PathKeysComparison::Better1 => {
                        accept_new = !old_dominates(outercmp())
                    }
                    _ => {}
                }
            }
        }
        if remove_old {
            parent_rel.pathlist.remove(i);
        } else {
            if new_path.disabled_nodes > old_path.disabled_nodes
                || (new_path.disabled_nodes == old_path.disabled_nodes && new_path.total_cost >= old_path.total_cost)
            {
                insert_at = i + 1;
            }
            i += 1;
        }
        if !accept_new {
            return;
        }
    }
    parent_rel.pathlist.insert(insert_at, new_path);
}

/// create_join_path makes a path of a join of two paths, as Postgres' create_nestloop_path and create_hashjoin_path
/// do, with the join relation's row estimate, the costs, and the order of its rows given.
pub fn create_join_path(
    joinrel: usize,
    rel: &RelOptInfo,
    kind: fn(JoinPath) -> PathKind,
    join: JoinPath,
    (disabled_nodes, startup_cost, total_cost): Costs,
    pathkeys: Vec<PkId>,
) -> Rc<Path> {
    let param = join.outer.param.union(&join.inner.param).difference(&rel.relids);
    Rc::new(Path {
        kind: kind(join),
        parent: joinrel,
        relids: rel.relids.clone(),
        param,
        pathkeys,
        rows: rel.rows,
        width: rel.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
    })
}
