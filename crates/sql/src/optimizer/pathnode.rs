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

use super::PlannerInfo;
use super::costsize::{
    Costs, JoinCostWorkspace, JoinPathExtraData, cost_bitmap_and_node, cost_bitmap_heap_scan, cost_bitmap_or_node,
    final_cost_hashjoin, final_cost_mergejoin, final_cost_nestloop,
};
use super::nodes::{
    AggStrategy, BitmapPath, HashPath, JoinPath, JoinType, MemoizePath, MergePath, Path, PathKind, PkId, RelOptInfo,
    Relids, RinfoId, RteKind, SortGroupClause, SpecialJoinInfo, SubsetCompare, TargetEntry, UniquePath,
    UniquePathMethod,
};
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
    parent_rel.cheapest_unique_path = None;
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

/// create_join_path makes a path of a join of two paths with the join relation's row estimate and the given costs and
/// order of its rows, as Postgres' create_nestloop_path, create_mergejoin_path, and create_hashjoin_path fill one in.
pub fn create_join_path(
    joinrel: usize,
    rel: &RelOptInfo,
    kind: PathKind,
    (disabled_nodes, startup_cost, total_cost): Costs,
    pathkeys: Vec<PkId>,
) -> Rc<Path> {
    let join = kind.join().expect("a join path");
    let param = join.outer.param.union(&join.inner.param).difference(&rel.relids);
    Rc::new(Path {
        kind,
        parent: joinrel,
        relids: rel.relids.clone(),
        param,
        pathkeys,
        rows: rel.rows,
        width: rel.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
        pathtarget: None,
    })
}

/// create_nestloop_path makes the path of a nested loop of an inner path over an outer one, which leaves out the
/// clauses that a parameterized inner path tests already, as Postgres' function of the same name does.
#[allow(clippy::too_many_arguments)]
pub fn create_nestloop_path(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    jointype: JoinType,
    workspace: &JoinCostWorkspace,
    extra: &JoinPathExtraData,
    outer_path: Rc<Path>,
    inner_path: Rc<Path>,
    mut restrict_clauses: Vec<RinfoId>,
    pathkeys: Vec<PkId>,
) -> Rc<Path> {
    if inner_path.param.overlap(&root.rels[outer_path.parent].relids) {
        let enforced_serials = super::relnode::get_param_path_clause_serials(root, &inner_path);
        restrict_clauses.retain(|&r| !enforced_serials.is_member(root.rinfos[r].rinfo_serial));
    }
    let jpath = JoinPath { jointype, outer: outer_path, inner: inner_path, joinrestrictinfo: restrict_clauses };
    let cost = final_cost_nestloop(root, joinrel, &jpath, workspace, extra);
    create_join_path(joinrel, &root.rels[joinrel], PathKind::NestLoop(jpath), cost, pathkeys)
}

/// create_mergejoin_path makes the path of a merge join of two paths by merge clauses, sorting each side first by
/// its sort keys when those are given, as Postgres' function of the same name does.
#[allow(clippy::too_many_arguments)]
pub fn create_mergejoin_path(
    root: &PlannerInfo<'_, '_>,
    joinrel: usize,
    jointype: JoinType,
    workspace: &JoinCostWorkspace,
    extra: &JoinPathExtraData,
    outer_path: Rc<Path>,
    inner_path: Rc<Path>,
    restrict_clauses: Vec<RinfoId>,
    pathkeys: Vec<PkId>,
    mergeclauses: Vec<RinfoId>,
    outersortkeys: Vec<PkId>,
    innersortkeys: Vec<PkId>,
) -> Rc<Path> {
    let jpath = JoinPath { jointype, outer: outer_path, inner: inner_path, joinrestrictinfo: restrict_clauses };
    let mut pathnode = MergePath {
        jpath,
        path_mergeclauses: mergeclauses,
        outersortkeys,
        innersortkeys,
        skip_mark_restore: false,
        materialize_inner: false,
    };
    let cost = final_cost_mergejoin(root, &mut pathnode, workspace, extra);
    create_join_path(joinrel, &root.rels[joinrel], PathKind::MergeJoin(Box::new(pathnode)), cost, pathkeys)
}

/// create_hashjoin_path makes the path of a hash join of two paths by hash clauses, whose rows are in no order, as
/// Postgres' function of the same name does.
#[allow(clippy::too_many_arguments)]
pub fn create_hashjoin_path(
    root: &PlannerInfo<'_, '_>,
    joinrel: usize,
    jointype: JoinType,
    workspace: &JoinCostWorkspace,
    extra: &JoinPathExtraData,
    outer_path: Rc<Path>,
    inner_path: Rc<Path>,
    restrict_clauses: Vec<RinfoId>,
    hashclauses: Vec<RinfoId>,
) -> Rc<Path> {
    let jpath = JoinPath { jointype, outer: outer_path, inner: inner_path, joinrestrictinfo: restrict_clauses };
    let mut pathnode = HashPath { jpath, path_hashclauses: hashclauses, num_batches: 0.0 };
    let cost = final_cost_hashjoin(root, &mut pathnode, workspace, extra);
    create_join_path(joinrel, &root.rels[joinrel], PathKind::HashJoin(Box::new(pathnode)), cost, Vec::new())
}

/// create_memoize_path makes a path that caches a parameterized path's rows by the values of its parameter
/// expressions, for a number of runs, as Postgres' function of the same name does.
pub fn create_memoize_path(
    subpath: &Rc<Path>,
    param_exprs: Vec<crate::expr::Expr>,
    binary_mode: bool,
    calls: f64,
) -> Rc<Path> {
    let mpath = MemoizePath {
        subpath: subpath.clone(),
        param_exprs,
        binary_mode,
        calls: super::costsize::clamp_row_est(calls),
    };
    Rc::new(Path {
        kind: PathKind::Memoize(Box::new(mpath)),
        startup_cost: subpath.startup_cost + super::costsize::CPU_TUPLE_COST,
        total_cost: subpath.total_cost + super::costsize::CPU_TUPLE_COST,
        ..(**subpath).clone()
    })
}

/// create_unique_path returns the path of a semi join's inner relation with the inner values of the join's
/// equalities made unique, from the relation's cheapest path, building it once, as Postgres' function of the same
/// name does: rows that a unique index or the subquery proves unique already, or rows hashed or sorted, whichever
/// costs less, and None when the join's equalities can neither sort nor hash.
pub fn create_unique_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    sjinfo: &SpecialJoinInfo,
) -> Option<Rc<Path>> {
    if let Some(path) = &root.rels[rel].cheapest_unique_path {
        return Some(path.clone());
    }
    let mut semi_can_hash = sjinfo.semi_can_hash;
    if !(sjinfo.semi_can_btree || semi_can_hash) {
        return None;
    }
    let mut uniq_exprs = Vec::new();
    let (mut newtlist, mut sort_list) = (Vec::new(), Vec::new());
    for uniqexpr in &sjinfo.semi_rhs_exprs {
        if super::nodefuncs::expr_type(root, uniqexpr).and_then(super::nodefuncs::btree_opfamily).is_some() {
            let ressortgroupref = newtlist.len() + 1;
            newtlist.push(TargetEntry { expr: uniqexpr.clone(), resjunk: false, ressortgroupref });
            sort_list.push(SortGroupClause {
                tle_sort_group_ref: ressortgroupref,
                descending: false,
                nulls_first: false,
                hashable: false,
            });
            let sort_pathkeys =
                super::planner::make_pathkeys_for_sortclauses(root, &sort_list, &newtlist).unwrap_or_default();
            if sort_pathkeys.len() != sort_list.len() {
                sort_list.pop();
                newtlist.pop();
                continue;
            }
        }
        uniq_exprs.push(uniqexpr.clone());
    }
    if uniq_exprs.is_empty() {
        return None;
    }
    let relid = root.rels[rel].relid;
    let unique_already = match relid {
        0 => false,
        _ => match &root.parse.rte(relid).kind {
            RteKind::Relation(..) => {
                sjinfo.semi_can_btree
                    && super::indxpath::relation_has_unique_index_ext(root, rel, &[], &uniq_exprs, None)
            }
            RteKind::Subquery(subquery, _) => {
                super::analyzejoins::query_supports_distinctness(subquery)
                    && translate_sub_tlist(root, &uniq_exprs, relid)
                        .is_some_and(|colnos| super::analyzejoins::query_is_distinct_for(subquery, &colnos))
            }
            _ => false,
        },
    };
    let rel_rows = root.rels[rel].rows;
    let path = |umethod, rows, (disabled_nodes, startup_cost, total_cost), pathkeys| {
        let upath = UniquePath { subpath: subpath.clone(), umethod, uniq_exprs: uniq_exprs.clone() };
        Rc::new(Path {
            kind: PathKind::UniquePath(Box::new(upath)),
            pathkeys,
            rows,
            disabled_nodes,
            startup_cost,
            total_cost,
            pathtarget: None,
            ..(*subpath).clone()
        })
    };
    if unique_already {
        let costs = (subpath.disabled_nodes, subpath.startup_cost, subpath.total_cost);
        let pathnode = path(UniquePathMethod::Noop, rel_rows, costs, subpath.pathkeys.clone());
        root.rels[rel].cheapest_unique_path = Some(pathnode.clone());
        return Some(pathnode);
    }
    let rows = super::selfuncs::estimate_num_groups(root, &uniq_exprs, rel_rows, None, None);
    let num_cols = uniq_exprs.len();
    let sort_path = sjinfo.semi_can_btree.then(|| {
        let (disabled_nodes, startup_cost, total_cost) = super::costsize::cost_sort(root, &subpath, -1.0);
        (disabled_nodes, startup_cost, total_cost + super::costsize::CPU_OPERATOR_COST * rel_rows * num_cols as f64)
    });
    let mut agg_path = None;
    if semi_can_hash {
        let hashentrysize = subpath.width + 64.0;
        if hashentrysize * rows > super::costsize::HASH_MEM {
            semi_can_hash = false;
        } else {
            let costs = (subpath.disabled_nodes, subpath.startup_cost, subpath.total_cost);
            let aggcosts = super::prepagg::AggClauseCosts::default();
            let (_, costs) = super::costsize::cost_agg(
                root,
                AggStrategy::Hashed,
                &aggcosts,
                num_cols,
                rows,
                &[],
                costs,
                rel_rows,
                subpath.width,
            );
            agg_path = Some(costs);
        }
    }
    let (umethod, costs) = match (sort_path, agg_path.filter(|_| semi_can_hash)) {
        (Some(sort), Some(agg)) if agg.0 < sort.0 || (agg.0 == sort.0 && agg.2 < sort.2) => {
            (UniquePathMethod::Hash, agg)
        }
        (Some(sort), _) => (UniquePathMethod::Sort, sort),
        (None, Some(agg)) => (UniquePathMethod::Hash, agg),
        (None, None) => return None,
    };
    let pathnode = path(umethod, rows, costs, Vec::new());
    root.rels[rel].cheapest_unique_path = Some(pathnode.clone());
    Some(pathnode)
}

/// translate_sub_tlist returns the subquery output columns that a list of a subquery relation's Vars reads, or None
/// when one is not such a Var, as Postgres' function of the same name does.
fn translate_sub_tlist(root: &PlannerInfo<'_, '_>, tlist: &[crate::expr::Expr], relid: usize) -> Option<Vec<usize>> {
    tlist
        .iter()
        .map(|e| match e {
            crate::expr::Expr::Column(id) => match root.glob.node(*id) {
                super::nodes::VarNode::Var(var) if var.varno == relid => Some(var.varattno),
                _ => None,
            },
            _ => None,
        })
        .collect()
}

/// add_path_precheck reports whether a path of these costs and order could be added to a relation's paths, before
/// it is built, as Postgres' function of the same name does: false when a path that costs no more in total, starts
/// as cheaply where that matters, and is as well ordered needs the same outer relations.
pub fn add_path_precheck(
    parent_rel: &RelOptInfo,
    disabled_nodes: usize,
    startup_cost: f64,
    total_cost: f64,
    pathkeys: &[PkId],
    required_outer: &Relids,
) -> bool {
    let new_path_pathkeys: &[PkId] = if required_outer.is_empty() { pathkeys } else { &[] };
    let consider_startup = match required_outer.is_empty() {
        true => parent_rel.consider_startup,
        false => parent_rel.consider_param_startup,
    };
    for old_path in &parent_rel.pathlist {
        if old_path.disabled_nodes != disabled_nodes {
            if disabled_nodes < old_path.disabled_nodes {
                break;
            }
        } else if total_cost <= old_path.total_cost * STD_FUZZ_FACTOR {
            break;
        }
        if startup_cost > old_path.startup_cost * STD_FUZZ_FACTOR || !consider_startup {
            let old_path_pathkeys: &[PkId] = if old_path.param.is_empty() { &old_path.pathkeys } else { &[] };
            let keyscmp = compare_pathkeys(new_path_pathkeys, old_path_pathkeys);
            if matches!(keyscmp, PathKeysComparison::Equal | PathKeysComparison::Better2)
                && *required_outer == old_path.param
            {
                return false;
            }
        }
    }
    true
}

/// calc_nestloop_required_outer returns the relations that a nested loop of two paths needs rows of from outside it,
/// which the outer path supplies to the inner one, as Postgres' function of the same name does.
pub fn calc_nestloop_required_outer(
    outerrelids: &Relids,
    outer_paramrels: &Relids,
    inner_paramrels: &Relids,
) -> Relids {
    if inner_paramrels.is_empty() {
        return outer_paramrels.clone();
    }
    outer_paramrels.union(inner_paramrels).difference(outerrelids)
}

/// calc_non_nestloop_required_outer returns the relations that a merge or hash join of two paths needs rows of from
/// outside it, as Postgres' function of the same name does.
pub fn calc_non_nestloop_required_outer(outer_path: &Path, inner_path: &Path) -> Relids {
    outer_path.param.union(&inner_path.param)
}

/// create_bitmap_heap_path makes the path of a scan of a base relation's rows whose keys a tree of index scans finds,
/// parameterized by the given outer relations, as Postgres' function of the same name does.
pub fn create_bitmap_heap_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    bitmapqual: Rc<Path>,
    required_outer: &Relids,
    loop_count: f64,
) -> Rc<Path> {
    let ppi = super::relnode::get_baserel_parampathinfo(root, rel, required_outer);
    let ((disabled_nodes, startup_cost, total_cost), rows) =
        cost_bitmap_heap_scan(root, rel, ppi.as_ref(), &bitmapqual, loop_count);
    let parent = &root.rels[rel];
    Rc::new(Path {
        kind: PathKind::BitmapHeapScan(bitmapqual),
        parent: rel,
        relids: parent.relids.clone(),
        param: required_outer.clone(),
        pathkeys: Vec::new(),
        rows,
        width: parent.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
        pathtarget: None,
    })
}

/// create_bitmap_and_path makes the path of a BitmapAnd of index scans, as Postgres' function of the same name does.
pub fn create_bitmap_and_path(root: &PlannerInfo<'_, '_>, rel: usize, bitmapquals: Vec<Rc<Path>>) -> Rc<Path> {
    let (total_cost, bitmapselectivity) = cost_bitmap_and_node(&bitmapquals);
    create_bitmap_tree_path(root, rel, PathKind::BitmapAnd, BitmapPath { bitmapquals, bitmapselectivity }, total_cost)
}

/// create_bitmap_or_path makes the path of a BitmapOr of index scans, as Postgres' function of the same name does.
pub fn create_bitmap_or_path(root: &PlannerInfo<'_, '_>, rel: usize, bitmapquals: Vec<Rc<Path>>) -> Rc<Path> {
    let (total_cost, bitmapselectivity) = cost_bitmap_or_node(&bitmapquals);
    create_bitmap_tree_path(root, rel, PathKind::BitmapOr, BitmapPath { bitmapquals, bitmapselectivity }, total_cost)
}

/// create_bitmap_tree_path makes the path of a BitmapAnd or BitmapOr, parameterized by every outer relation that its
/// index scans are, which returns no rows of its own.
fn create_bitmap_tree_path(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    kind: fn(BitmapPath) -> PathKind,
    bitmap: BitmapPath,
    total_cost: f64,
) -> Rc<Path> {
    let param = bitmap.bitmapquals.iter().fold(Relids::new(), |outer, p| outer.union(&p.param));
    let parent = &root.rels[rel];
    Rc::new(Path {
        kind: kind(bitmap),
        parent: rel,
        relids: parent.relids.clone(),
        param,
        pathkeys: Vec::new(),
        rows: 0.0,
        width: parent.reltarget.width,
        disabled_nodes: 0,
        startup_cost: total_cost,
        total_cost,
        pathtarget: None,
    })
}

/// upper_path returns a path of an upper relation over another path, with its target, rows, and costs, and its
/// subpath's order unless given.
#[allow(clippy::too_many_arguments)]
fn upper_path(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    kind: PathKind,
    target: Option<Rc<super::nodes::PathTarget>>,
    pathkeys: Vec<PkId>,
    rows: f64,
    (disabled_nodes, startup_cost, total_cost): Costs,
) -> Rc<Path> {
    let parent = &root.rels[rel];
    let width = target.as_ref().map_or(parent.reltarget.width, |t| t.width);
    Rc::new(Path {
        kind,
        parent: rel,
        relids: parent.relids.clone(),
        param: Relids::new(),
        pathkeys,
        rows,
        width,
        disabled_nodes,
        startup_cost,
        total_cost,
        pathtarget: target,
    })
}

/// path_target returns the target of a path's rows.
pub fn path_target(root: &PlannerInfo<'_, '_>, path: &Path) -> Rc<super::nodes::PathTarget> {
    match &path.pathtarget {
        Some(target) => target.clone(),
        None => Rc::new(root.rels[path.parent].reltarget.clone()),
    }
}

/// create_sort_path makes the path of a sort of another path's rows by pathkeys, of which a LIMIT may read only some,
/// as Postgres' function of the same name does.
pub fn create_sort_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    pathkeys: Vec<PkId>,
    limit_tuples: f64,
) -> Rc<Path> {
    let costs = super::costsize::cost_sort(root, &subpath, limit_tuples);
    let target = Some(path_target(root, &subpath));
    let rows = subpath.rows;
    upper_path(root, rel, PathKind::Sort(subpath), target, pathkeys, rows, costs)
}

/// create_incremental_sort_path makes the path of a sort of another path's rows by pathkeys whose leading keys it
/// is sorted by already, as Postgres' function of the same name does.
pub fn create_incremental_sort_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    pathkeys: Vec<PkId>,
    presorted_keys: usize,
    limit_tuples: f64,
) -> Rc<Path> {
    let costs = super::costsize::cost_incremental_sort(root, &pathkeys, presorted_keys, &subpath, limit_tuples);
    let target = Some(path_target(root, &subpath));
    let rows = subpath.rows;
    upper_path(root, rel, PathKind::IncrementalSort(subpath), target, pathkeys, rows, costs)
}

/// is_projection_capable_path reports whether a path's plan computes any target, as Postgres' function of the same
/// name in createplan.c decides: every plan but a sort, a unique pass, a limit, or a set operation.
pub fn is_projection_capable_path(path: &Path) -> bool {
    !matches!(
        path.kind,
        PathKind::Sort(_)
            | PathKind::IncrementalSort(_)
            | PathKind::Unique(..)
            | PathKind::Limit(_)
            | PathKind::Material(_)
            | PathKind::Append(_)
            | PathKind::ProjectSet(_)
    )
}

/// create_projection_path makes the path that computes a target over another path's rows, which costs only the
/// target's own expressions when the subpath can compute it itself, as Postgres' function of the same name does.
pub fn create_projection_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    target: Rc<super::nodes::PathTarget>,
) -> Rc<Path> {
    let subpath = match &subpath.kind {
        PathKind::Projection(inner) if inner.parent == rel => inner.clone(),
        _ => subpath,
    };
    let oldtarget = path_target(root, &subpath);
    let (startup_cost, total_cost) = match is_projection_capable_path(&subpath) || oldtarget.exprs == target.exprs {
        true => (
            subpath.startup_cost + (target.cost.startup - oldtarget.cost.startup),
            subpath.total_cost
                + (target.cost.startup - oldtarget.cost.startup)
                + (target.cost.per_tuple - oldtarget.cost.per_tuple) * subpath.rows,
        ),
        false => (
            subpath.startup_cost + target.cost.startup,
            subpath.total_cost
                + target.cost.startup
                + (super::costsize::CPU_TUPLE_COST + target.cost.per_tuple) * subpath.rows,
        ),
    };
    let (rows, disabled_nodes, pathkeys) = (subpath.rows, subpath.disabled_nodes, subpath.pathkeys.clone());
    let costs = (disabled_nodes, startup_cost, total_cost);
    upper_path(root, rel, PathKind::Projection(subpath), Some(target), pathkeys, rows, costs)
}

/// apply_projection_to_path makes a path compute a target, changing its target when its plan can compute any, as
/// Postgres' function of the same name does.
pub fn apply_projection_to_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    path: Rc<Path>,
    target: Rc<super::nodes::PathTarget>,
) -> Rc<Path> {
    if !is_projection_capable_path(&path) {
        return create_projection_path(root, rel, path, target);
    }
    let oldcost = path_target(root, &path).cost;
    let mut path = (*path).clone();
    path.startup_cost += target.cost.startup - oldcost.startup;
    path.total_cost += target.cost.startup - oldcost.startup + (target.cost.per_tuple - oldcost.per_tuple) * path.rows;
    path.width = target.width;
    path.pathtarget = Some(target);
    Rc::new(path)
}

/// create_set_projection_path makes the path that computes a target with set-returning functions over another
/// path's rows, as Postgres' function of the same name does.
pub fn create_set_projection_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    target: Rc<super::nodes::PathTarget>,
) -> Rc<Path> {
    let tlist_rows =
        target.exprs.iter().map(|e| super::clauses::expression_returns_set_rows(root, e)).fold(1.0, f64::max);
    let rows = subpath.rows * tlist_rows;
    let startup_cost = subpath.startup_cost + target.cost.startup;
    let total_cost = subpath.total_cost
        + target.cost.startup
        + (super::costsize::CPU_TUPLE_COST + target.cost.per_tuple) * subpath.rows
        + (rows - subpath.rows) * super::costsize::CPU_TUPLE_COST / 2.0;
    let (disabled_nodes, pathkeys) = (subpath.disabled_nodes, subpath.pathkeys.clone());
    upper_path(
        root,
        rel,
        PathKind::ProjectSet(subpath),
        Some(target),
        pathkeys,
        rows,
        (disabled_nodes, startup_cost, total_cost),
    )
}

/// create_agg_path makes the path of an aggregation of another path's rows, as Postgres' function of the same name
/// does: a sorted aggregation keeps the order of its group keys.
#[allow(clippy::too_many_arguments)]
pub fn create_agg_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    target: Rc<super::nodes::PathTarget>,
    aggstrategy: super::nodes::AggStrategy,
    group_clause: Vec<super::nodes::SortGroupClause>,
    qual: Vec<crate::expr::Expr>,
    aggcosts: &super::prepagg::AggClauseCosts,
    num_groups: f64,
) -> Rc<Path> {
    let pathkeys = match aggstrategy {
        super::nodes::AggStrategy::Sorted => subpath.pathkeys.iter().take(root.num_groupby_pathkeys).copied().collect(),
        _ => Vec::new(),
    };
    let input_width = path_target(root, &subpath).width;
    let (rows, (disabled_nodes, mut startup_cost, mut total_cost)) = super::costsize::cost_agg(
        root,
        aggstrategy,
        aggcosts,
        group_clause.len(),
        num_groups,
        &qual,
        (subpath.disabled_nodes, subpath.startup_cost, subpath.total_cost),
        subpath.rows,
        input_width,
    );
    startup_cost += target.cost.startup;
    total_cost += target.cost.startup + target.cost.per_tuple * rows;
    let kind = PathKind::Agg(Box::new(super::nodes::AggPath { subpath, group_clause, qual }));
    upper_path(root, rel, kind, Some(target), pathkeys, rows, (disabled_nodes, startup_cost, total_cost))
}

/// create_group_path makes the path of a GROUP BY without aggregates over another path's sorted rows, as Postgres'
/// function of the same name does.
pub fn create_group_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    group_clause: Vec<super::nodes::SortGroupClause>,
    qual: Vec<crate::expr::Expr>,
    num_groups: f64,
) -> Rc<Path> {
    let target = Rc::new(root.rels[rel].reltarget.clone());
    let (rows, (disabled_nodes, mut startup_cost, mut total_cost)) = super::costsize::cost_group(
        root,
        group_clause.len(),
        num_groups,
        &qual,
        (subpath.disabled_nodes, subpath.startup_cost, subpath.total_cost),
        subpath.rows,
    );
    startup_cost += target.cost.startup;
    total_cost += target.cost.startup + target.cost.per_tuple * rows;
    let pathkeys = subpath.pathkeys.clone();
    let kind = PathKind::Group(Box::new(super::nodes::GroupPath { subpath, group_clause, qual }));
    upper_path(root, rel, kind, Some(target), pathkeys, rows, (disabled_nodes, startup_cost, total_cost))
}

/// create_upper_unique_path makes the path of the first row of each run of another sorted path's rows with equal
/// first keys, as Postgres' function of the same name does.
pub fn create_upper_unique_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    num_cols: usize,
    num_groups: f64,
) -> Rc<Path> {
    let target = Some(path_target(root, &subpath));
    let costs = (
        subpath.disabled_nodes,
        subpath.startup_cost,
        subpath.total_cost + super::costsize::CPU_OPERATOR_COST * subpath.rows * num_cols as f64,
    );
    let pathkeys = subpath.pathkeys.clone();
    upper_path(root, rel, PathKind::Unique(subpath, num_cols), target, pathkeys, num_groups, costs)
}

/// create_windowagg_path makes the path of one window's calls over another path's rows, as Postgres' function of the
/// same name does.
pub fn create_windowagg_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    target: Rc<super::nodes::PathTarget>,
    calls: Vec<usize>,
) -> Rc<Path> {
    let (disabled_nodes, mut startup_cost, mut total_cost) = super::costsize::cost_windowagg(
        root,
        &calls,
        (subpath.disabled_nodes, subpath.startup_cost, subpath.total_cost),
        subpath.rows,
    );
    let rows = subpath.rows;
    startup_cost += target.cost.startup;
    total_cost += target.cost.startup + target.cost.per_tuple * rows;
    let pathkeys = subpath.pathkeys.clone();
    let kind = PathKind::WindowAgg(Box::new(super::nodes::WindowAggPath { subpath, calls }));
    upper_path(root, rel, kind, Some(target), pathkeys, rows, (disabled_nodes, startup_cost, total_cost))
}

/// create_limit_path makes the path of the query's LIMIT and OFFSET over another path, with their estimated row
/// counts, as Postgres' function of the same name does.
pub fn create_limit_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    offset_est: i64,
    count_est: i64,
) -> Rc<Path> {
    let (limit_offset, limit_count) = (root.parse.limit_offset.clone(), root.parse.limit_count.clone());
    limit_path(root, rel, subpath, limit_offset, limit_count, offset_est, count_est)
}

/// create_limit_path_one makes the path of a LIMIT 1 over another path, as Postgres' create_final_distinct_paths
/// makes one for a DISTINCT whose keys are all constant.
pub fn create_limit_path_one(root: &mut PlannerInfo<'_, '_>, rel: usize, subpath: Rc<Path>) -> Rc<Path> {
    let one = Some(crate::expr::Expr::Const(crate::types::Value::Int8(1)));
    limit_path(root, rel, subpath, None, one, 0, 1)
}

/// limit_path makes the path of a LIMIT and OFFSET over another path.
fn limit_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    limit_offset: Option<crate::expr::Expr>,
    limit_count: Option<crate::expr::Expr>,
    offset_est: i64,
    count_est: i64,
) -> Rc<Path> {
    let (mut rows, mut startup_cost, mut total_cost) = (subpath.rows, subpath.startup_cost, subpath.total_cost);
    adjust_limit_rows_costs(&mut rows, &mut startup_cost, &mut total_cost, offset_est, count_est);
    let target = Some(path_target(root, &subpath));
    let (disabled_nodes, pathkeys) = (subpath.disabled_nodes, subpath.pathkeys.clone());
    let kind = PathKind::Limit(Box::new(super::nodes::LimitPath { subpath, limit_offset, limit_count }));
    upper_path(root, rel, kind, target, pathkeys, rows, (disabled_nodes, startup_cost, total_cost))
}

/// adjust_limit_rows_costs adjusts a path's rows and costs for the rows that a LIMIT and OFFSET read, where -1 is an
/// estimate that is not a constant, as Postgres' function of the same name does.
pub fn adjust_limit_rows_costs(
    rows: &mut f64,
    startup_cost: &mut f64,
    total_cost: &mut f64,
    offset_est: i64,
    count_est: i64,
) {
    let (input_rows, input_startup_cost, input_total_cost) = (*rows, *startup_cost, *total_cost);
    if offset_est != 0 {
        let offset_rows = match offset_est > 0 {
            true => offset_est as f64,
            false => super::costsize::clamp_row_est(input_rows * 0.10),
        }
        .min(*rows);
        if input_rows > 0.0 {
            *startup_cost += (input_total_cost - input_startup_cost) * offset_rows / input_rows;
        }
        *rows = (*rows - offset_rows).max(1.0);
    }
    if count_est != 0 {
        let count_rows = match count_est > 0 {
            true => count_est as f64,
            false => super::costsize::clamp_row_est(input_rows * 0.10),
        }
        .min(*rows);
        if input_rows > 0.0 {
            *total_cost = *startup_cost + (input_total_cost - input_startup_cost) * count_rows / input_rows;
        }
        *rows = count_rows.max(1.0);
    }
}

/// create_group_result_path makes the path of the one row of a grouping without input rows or group keys under its
/// HAVING conditions, as Postgres' function of the same name does.
pub fn create_group_result_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    target: Rc<super::nodes::PathTarget>,
    havingqual: Vec<crate::expr::Expr>,
) -> Rc<Path> {
    let mut startup_cost = target.cost.startup;
    let mut total_cost = target.cost.startup + super::costsize::CPU_TUPLE_COST + target.cost.per_tuple;
    for qual in &havingqual {
        let cost = super::costsize::cost_qual_eval_node(qual);
        startup_cost += cost.startup + cost.per_tuple;
        total_cost += cost.startup + cost.per_tuple;
    }
    upper_path(root, rel, PathKind::Result(havingqual), Some(target), Vec::new(), 1.0, (0, startup_cost, total_cost))
}

/// create_append_path makes the path of the rows of several paths in turn, as Postgres' function of the same name
/// does without parallel workers.
pub fn create_append_path(root: &mut PlannerInfo<'_, '_>, rel: usize, subpaths: Vec<Rc<Path>>) -> Rc<Path> {
    let rows = subpaths.iter().map(|p| p.rows).sum();
    let startup_cost = subpaths.first().map_or(0.0, |p| p.startup_cost);
    let total_cost = subpaths.iter().map(|p| p.total_cost).sum::<f64>()
        + super::costsize::CPU_TUPLE_COST * super::costsize::APPEND_CPU_COST_MULTIPLIER * rows;
    let disabled_nodes = subpaths.iter().map(|p| p.disabled_nodes).sum();
    let target = Some(Rc::new(root.rels[rel].reltarget.clone()));
    upper_path(
        root,
        rel,
        PathKind::Append(subpaths),
        target,
        Vec::new(),
        rows,
        (disabled_nodes, startup_cost, total_cost),
    )
}

/// create_subqueryscan_path returns the path of a scan of a subquery relation's rows from one of its subquery's final
/// paths, by its position among the relation's `subplans`, as Postgres' function of the same name does.
pub fn create_subqueryscan_path(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    subplan: usize,
    trivial_pathtarget: bool,
    pathkeys: Vec<PkId>,
) -> Rc<Path> {
    let parent = &root.rels[rel];
    let subpath = &parent.subplans[subplan].path;
    let (rows, (disabled_nodes, startup_cost, total_cost)) =
        super::costsize::cost_subqueryscan(root, rel, subpath, trivial_pathtarget);
    Rc::new(Path {
        kind: PathKind::SubqueryScan(subplan),
        parent: rel,
        relids: parent.relids.clone(),
        param: parent.lateral_relids.clone(),
        pathkeys,
        rows,
        width: parent.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
        pathtarget: None,
    })
}

/// create_setop_path returns the path of an INTERSECT or EXCEPT of two paths, hashed or of sorted inputs, with its
/// estimated groups and rows, as Postgres' function of the same name does.
pub fn create_setop_path(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    (leftpath, rightpath): (Rc<Path>, Rc<Path>),
    (op, all): (crate::plan::SetOp, bool),
    hashed: bool,
    group_list: Vec<super::nodes::SortGroupClause>,
    (num_groups, output_rows): (f64, f64),
) -> Rc<Path> {
    use super::costsize::{CPU_OPERATOR_COST, HASH_MEM, maxalign};
    let mut disabled_nodes = leftpath.disabled_nodes + rightpath.disabled_nodes;
    let compare_cost = CPU_OPERATOR_COST * (leftpath.rows + rightpath.rows) * group_list.len() as f64;
    let (startup_cost, mut total_cost) = match hashed {
        false => {
            (leftpath.startup_cost + rightpath.startup_cost, leftpath.total_cost + rightpath.total_cost + compare_cost)
        }
        true => {
            let startup = leftpath.total_cost + rightpath.total_cost + compare_cost;
            disabled_nodes += usize::from(!root.enables.hashagg);
            let hashentrysize = maxalign(leftpath.width) + maxalign(MINIMAL_TUPLE_HEADER);
            disabled_nodes += usize::from(hashentrysize * num_groups > HASH_MEM);
            (startup, startup)
        }
    };
    total_cost += CPU_OPERATOR_COST * output_rows;
    let parent = &root.rels[rel];
    let pathkeys = match hashed {
        false => leftpath.pathkeys.clone(),
        true => Vec::new(),
    };
    let kind = PathKind::SetOp(Box::new(super::nodes::SetOpPath { leftpath, rightpath, op, all }));
    Rc::new(Path {
        kind,
        parent: rel,
        relids: parent.relids.clone(),
        param: Relids::new(),
        pathkeys,
        rows: output_rows,
        width: parent.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
        pathtarget: Some(Rc::new(parent.reltarget.clone())),
    })
}

/// MINIMAL_TUPLE_HEADER is the size of the header of a row that a hash table keeps, as Postgres'
/// SizeofMinimalTupleHeader is.
const MINIMAL_TUPLE_HEADER: f64 = 15.0;

/// create_recursiveunion_path returns the path of a recursive WITH query's union of its non-recursive term's path
/// and its recursive term's path, as Postgres' function of the same name does.
pub fn create_recursiveunion_path(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    leftpath: Rc<Path>,
    rightpath: Rc<Path>,
    distinct: bool,
    wt_param_id: usize,
) -> Rc<Path> {
    let parent = &root.rels[rel];
    let mut target = parent.reltarget.clone();
    let (rows, costs) = super::costsize::cost_recursive_union(&leftpath, &rightpath);
    target.width = leftpath.width.max(rightpath.width);
    Rc::new(Path {
        parent: rel,
        relids: parent.relids.clone(),
        param: Relids::new(),
        pathkeys: Vec::new(),
        rows,
        width: target.width,
        disabled_nodes: costs.0,
        startup_cost: costs.1,
        total_cost: costs.2,
        pathtarget: Some(Rc::new(target)),
        kind: PathKind::RecursiveUnion(Box::new(super::nodes::RecursiveUnionPath {
            leftpath,
            rightpath,
            distinct,
            wt_param_id,
        })),
    })
}

/// create_merge_append_path returns the path of the rows of several paths merged in the order of pathkeys, sorting
/// those that are not in it, as Postgres' function of the same name does.
pub fn create_merge_append_path(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    subpaths: Vec<Rc<Path>>,
    pathkeys: Vec<PkId>,
) -> Rc<Path> {
    let parent = &root.rels[rel];
    let limit_tuples = match parent.relids == root.all_query_rels {
        true => root.limit_tuples,
        false => -1.0,
    };
    let (mut rows, mut input_disabled_nodes, mut input_startup_cost, mut input_total_cost) = (0.0, 0, 0.0, 0.0);
    for subpath in &subpaths {
        rows += subpath.rows;
        let (disabled_nodes, startup_cost, total_cost) =
            match super::pathkeys::pathkeys_contained_in(&pathkeys, &subpath.pathkeys) {
                true => (subpath.disabled_nodes, subpath.startup_cost, subpath.total_cost),
                false => super::costsize::cost_sort(root, subpath, limit_tuples),
            };
        input_disabled_nodes += disabled_nodes;
        input_startup_cost += startup_cost;
        input_total_cost += total_cost;
    }
    let costs = match subpaths.len() {
        1 => (input_disabled_nodes, input_startup_cost, input_total_cost),
        n => super::costsize::cost_merge_append(n, (input_disabled_nodes, input_startup_cost, input_total_cost), rows),
    };
    let target = Some(Rc::new(parent.reltarget.clone()));
    upper_path(root, rel, PathKind::MergeAppend(subpaths), target, pathkeys, rows, costs)
}

/// create_minmaxagg_path returns the path of the one row of a query's MIN and MAX aggregates that initplans read from
/// indexes, under its HAVING conditions, as Postgres' function of the same name does.
pub fn create_minmaxagg_path(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    target: Rc<super::nodes::PathTarget>,
    mmaggregates: Vec<super::nodes::MinMaxAggInfo>,
    quals: Vec<crate::expr::Expr>,
) -> Rc<Path> {
    let initplan_cost: f64 = mmaggregates.iter().map(|m| m.pathcost).sum();
    let disabled_nodes = mmaggregates.iter().map(|m| m.disabled_nodes).sum();
    let mut startup_cost = initplan_cost + target.cost.startup;
    let mut total_cost = initplan_cost + target.cost.startup + target.cost.per_tuple + super::costsize::CPU_TUPLE_COST;
    if !quals.is_empty() {
        let qual_cost = quals.iter().fold(super::costsize::QualCost::default(), |cost, q| {
            let one = super::costsize::cost_qual_eval_node(q);
            super::costsize::QualCost { startup: cost.startup + one.startup, per_tuple: cost.per_tuple + one.per_tuple }
        });
        startup_cost += qual_cost.startup;
        total_cost += qual_cost.startup + qual_cost.per_tuple;
    }
    let kind = PathKind::MinMaxAgg(Box::new(super::nodes::MinMaxAggPath { mmaggregates, quals }));
    upper_path(root, rel, kind, Some(target), Vec::new(), 1.0, (disabled_nodes, startup_cost, total_cost))
}

/// create_groupingsets_path returns the path of an aggregation by grouping sets over a path's rows, in rollups that
/// are sorted or hashed, under the HAVING conditions, as Postgres' function of the same name does: the first rollup
/// costs an aggregation of the path's rows, each sorted rollup after the first sorted one a sort and an aggregation of
/// them, and each other one an aggregation of them.
pub fn create_groupingsets_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    subpath: Rc<Path>,
    having_qual: Vec<crate::expr::Expr>,
    aggstrategy: super::nodes::AggStrategy,
    rollups: Vec<super::nodes::RollupData>,
    agg_costs: &super::prepagg::AggClauseCosts,
) -> Rc<Path> {
    use super::nodes::AggStrategy;
    let target = Rc::new(root.rels[rel].reltarget.clone());
    let mut aggstrategy = aggstrategy;
    if aggstrategy == AggStrategy::Sorted && rollups.len() == 1 && rollups[0].group_clause.is_empty() {
        aggstrategy = AggStrategy::Plain;
    }
    if aggstrategy == AggStrategy::Mixed && rollups.len() == 1 {
        aggstrategy = AggStrategy::Hashed;
    }
    let pathkeys = match aggstrategy == AggStrategy::Sorted && rollups.len() == 1 {
        true => root.group_pathkeys.clone(),
        false => Vec::new(),
    };
    let input_width = path_target(root, &subpath).width;
    let (mut rows, (mut disabled_nodes, mut startup_cost, mut total_cost)) = (0.0, (0, 0.0, 0.0));
    let mut is_first_sort = true;
    for (i, rollup) in rollups.iter().enumerate() {
        let num_group_cols = rollup.gsets.first().map_or(0, Vec::len);
        if i == 0 {
            let input = (subpath.disabled_nodes, subpath.startup_cost, subpath.total_cost);
            (rows, (disabled_nodes, startup_cost, total_cost)) = super::costsize::cost_agg(
                root,
                aggstrategy,
                agg_costs,
                num_group_cols,
                rollup.num_groups,
                &having_qual,
                input,
                subpath.rows,
                input_width,
            );
            if !rollup.is_hashed {
                is_first_sort = false;
            }
            continue;
        }
        let (agg_rows, (agg_disabled, _, agg_total)) = if rollup.is_hashed || is_first_sort {
            let strategy = if rollup.is_hashed { AggStrategy::Hashed } else { AggStrategy::Sorted };
            if !rollup.is_hashed {
                is_first_sort = false;
            }
            super::costsize::cost_agg(
                root,
                strategy,
                agg_costs,
                num_group_cols,
                rollup.num_groups,
                &having_qual,
                (0, 0.0, 0.0),
                subpath.rows,
                input_width,
            )
        } else {
            let (sort_startup, sort_run) = super::costsize::cost_tuplesort(subpath.rows, input_width, 0.0, -1.0);
            let sort_disabled = usize::from(!root.enables.sort);
            super::costsize::cost_agg(
                root,
                AggStrategy::Sorted,
                agg_costs,
                num_group_cols,
                rollup.num_groups,
                &having_qual,
                (sort_disabled, sort_startup, sort_startup + sort_run),
                subpath.rows,
                input_width,
            )
        };
        disabled_nodes += agg_disabled;
        total_cost += agg_total;
        rows += agg_rows;
    }
    startup_cost += target.cost.startup;
    total_cost += target.cost.startup + target.cost.per_tuple * rows;
    let kind = PathKind::GroupingSets(Box::new(super::nodes::GroupingSetsPath { subpath, rollups, qual: having_qual }));
    upper_path(root, rel, kind, Some(target), pathkeys, rows, (disabled_nodes, startup_cost, total_cost))
}
