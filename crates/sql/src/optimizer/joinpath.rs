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

//! Postgres' optimizer/path/joinpath.c: the nested loop and hash join paths of a join of two relations. A path that
//! needs another relation's current row is only the inner side of a nested loop whose outer side supplies it, as
//! Doltgres' executor runs lookups.

use std::rc::Rc;

use super::PlannerInfo;
use super::costsize::{
    JoinPathExtraData, SemiAntiJoinFactors, compute_semi_anti_join_factors, cost_hashjoin, cost_material,
    cost_nestloop, has_indexed_join_quals,
};
use super::nodes::{JoinPath, JoinType, Path, PathKind, Relids, RestrictInfo, RinfoId, SpecialJoinInfo};
use super::pathkeys::build_join_pathkeys;
use super::pathnode::{add_path, create_join_path};
use super::restrictinfo::rinfo_is_pushed_down;

/// add_paths_to_joinrel adds the paths of a join of an outer relation to an inner one to their join relation, as
/// Postgres' function of the same name does.
pub fn add_paths_to_joinrel(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outerrel: usize,
    innerrel: usize,
    jointype: JoinType,
    sjinfo: &SpecialJoinInfo,
    restrictlist: &[RinfoId],
) {
    let joinrelids = root.rels[joinrel].relids.clone();
    let outerrelids = root.rels[outerrel].relids.clone();
    let inner_unique = match jointype {
        JoinType::Semi | JoinType::Anti => false,
        _ => super::analyzejoins::innerrel_is_unique(
            root,
            &joinrelids,
            &outerrelids,
            innerrel,
            jointype,
            restrictlist,
            false,
        ),
    };
    let semifactors = match matches!(jointype, JoinType::Semi | JoinType::Anti) || inner_unique {
        true => compute_semi_anti_join_factors(root, joinrel, outerrel, innerrel, jointype, sjinfo, restrictlist),
        false => SemiAntiJoinFactors::default(),
    };
    let extra = JoinPathExtraData { restrictlist: restrictlist.to_vec(), inner_unique, semifactors };
    match_unsorted_outer(root, joinrel, outerrel, innerrel, jointype, &extra);
    hash_inner_and_outer(root, joinrel, outerrel, innerrel, jointype, &extra);
    if jointype == JoinType::Full && root.rels[joinrel].pathlist.is_empty() {
        let outer = root.rels[outerrel].cheapest_total_path.clone().expect("every relation has a path");
        let inner = root.rels[innerrel].cheapest_total_path.clone().expect("every relation has a path");
        try_nestloop_path(root, joinrel, outer, inner, jointype, &extra);
    }
}

/// match_unsorted_outer adds the nested loops of each of the outer relation's paths over the inner relation's
/// cheapest paths, kept in memory or looking rows up, as Postgres' function of the same name does for joins that a
/// nested loop can run.
fn match_unsorted_outer(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outerrel: usize,
    innerrel: usize,
    jointype: JoinType,
    extra: &JoinPathExtraData,
) {
    if !matches!(jointype, JoinType::Inner | JoinType::Left | JoinType::Semi | JoinType::Anti) {
        return;
    }
    let matpath = root.rels[innerrel].cheapest_total_path.clone().and_then(|inner| match inner.kind {
        PathKind::Material(_) => None,
        _ => root.enables.material.then(|| create_material_path(root, &inner)),
    });
    let outer_relids = root.rels[outerrel].relids.clone();
    let outer_paths: Vec<Rc<Path>> =
        root.rels[outerrel].pathlist.iter().filter(|p| p.param.is_empty()).cloned().collect();
    let inner_paths: Vec<Rc<Path>> = root.rels[innerrel]
        .cheapest_parameterized_paths
        .iter()
        .filter(|p| p.param.is_subset(&outer_relids))
        .cloned()
        .collect();
    for outerpath in outer_paths {
        for innerpath in &inner_paths {
            try_nestloop_path(root, joinrel, outerpath.clone(), innerpath.clone(), jointype, extra);
        }
        if let Some(matpath) = &matpath {
            try_nestloop_path(root, joinrel, outerpath.clone(), matpath.clone(), jointype, extra);
        }
    }
}

/// create_material_path makes a path that keeps another path's rows in memory, as Postgres' function of the same
/// name does.
fn create_material_path(root: &PlannerInfo<'_, '_>, subpath: &Rc<Path>) -> Rc<Path> {
    let (disabled_nodes, startup_cost, total_cost) = cost_material(root, subpath);
    Rc::new(Path {
        kind: PathKind::Material(subpath.clone()),
        disabled_nodes,
        startup_cost,
        total_cost,
        ..(**subpath).clone()
    })
}

/// try_nestloop_path adds a nested loop of two paths to the join relation, whose rows keep the outer path's order,
/// as Postgres' function of the same name does with build_join_pathkeys.
fn try_nestloop_path(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outer: Rc<Path>,
    inner: Rc<Path>,
    jointype: JoinType,
    extra: &JoinPathExtraData,
) {
    let joinrelids = root.rels[joinrel].relids.clone();
    let mut restrict_clauses = extra.restrictlist.clone();
    if inner.param.overlap(&outer.relids) {
        let enforced_serials = super::relnode::get_param_path_clause_serials(root, &inner);
        restrict_clauses.retain(|&r| !enforced_serials.is_member(root.rinfos[r].rinfo_serial));
    }
    let has_indexed_join_quals = has_indexed_join_quals(root, &joinrelids, &inner, &restrict_clauses);
    let cost = cost_nestloop(root, jointype, &outer, &inner, extra, &restrict_clauses, has_indexed_join_quals);
    let pathkeys = build_join_pathkeys(root, joinrel, jointype, &outer.pathkeys);
    let join = JoinPath { jointype, outer, inner, joinrestrictinfo: restrict_clauses };
    let path = create_join_path(joinrel, &root.rels[joinrel], PathKind::NestLoop, join, cost, pathkeys);
    add_path(&mut root.rels[joinrel], path);
}

/// hash_inner_and_outer adds the hash joins of the relations' cheapest paths by the join's hashable equalities
/// between them, as Postgres' function of the same name does.
fn hash_inner_and_outer(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outerrel: usize,
    innerrel: usize,
    jointype: JoinType,
    extra: &JoinPathExtraData,
) {
    let joinrelids = &root.rels[joinrel].relids;
    let (outer_relids, inner_relids) = (&root.rels[outerrel].relids, &root.rels[innerrel].relids);
    let hashclauses: Vec<RinfoId> = extra
        .restrictlist
        .iter()
        .copied()
        .filter(|&r| {
            let r = &root.rinfos[r];
            !(jointype.is_outer() && rinfo_is_pushed_down(r, joinrelids))
                && r.can_join
                && r.hashjoinable
                && clause_sides_match_join(r, outer_relids, inner_relids)
        })
        .collect();
    if hashclauses.is_empty() {
        return;
    }
    let (Some(cheapest_startup_outer), Some(cheapest_total_outer), Some(cheapest_total_inner)) = (
        root.rels[outerrel].cheapest_startup_path.clone(),
        root.rels[outerrel].cheapest_total_path.clone(),
        root.rels[innerrel].cheapest_total_path.clone(),
    ) else {
        return;
    };
    try_hashjoin_path(
        root,
        joinrel,
        cheapest_startup_outer.clone(),
        cheapest_total_inner.clone(),
        &hashclauses,
        jointype,
        extra,
    );
    if !Rc::ptr_eq(&cheapest_startup_outer, &cheapest_total_outer) {
        try_hashjoin_path(root, joinrel, cheapest_total_outer, cheapest_total_inner, &hashclauses, jointype, extra);
    }
}

/// try_hashjoin_path adds a hash join of two paths to the join relation, as Postgres' function of the same name
/// does.
fn try_hashjoin_path(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outer: Rc<Path>,
    inner: Rc<Path>,
    hashclauses: &[RinfoId],
    jointype: JoinType,
    extra: &JoinPathExtraData,
) {
    let cost = cost_hashjoin(root, jointype, hashclauses, &outer, &inner, extra);
    let join = JoinPath { jointype, outer, inner, joinrestrictinfo: extra.restrictlist.clone() };
    let path = create_join_path(joinrel, &root.rels[joinrel], PathKind::HashJoin, join, cost, Vec::new());
    add_path(&mut root.rels[joinrel], path);
}

/// clause_sides_match_join reports whether a join clause compares an expression of the outer relations with one of
/// the inner relations, recording which side is the outer one, as Postgres' function of the same name does.
pub fn clause_sides_match_join(rinfo: &RestrictInfo, outer_relids: &Relids, inner_relids: &Relids) -> bool {
    if rinfo.left_relids.is_subset(outer_relids) && rinfo.right_relids.is_subset(inner_relids) {
        rinfo.outer_is_left.set(true);
        return true;
    }
    if rinfo.left_relids.is_subset(inner_relids) && rinfo.right_relids.is_subset(outer_relids) {
        rinfo.outer_is_left.set(false);
        return true;
    }
    false
}
