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

//! Postgres' optimizer/path/joinpath.c: the nested loop, merge join, and hash join paths of a join of two relations.
//! A path that needs another relation's current row is only the inner side of a nested loop whose outer side supplies
//! it, as Doltgres' executor runs lookups.

use std::rc::Rc;

use super::PlannerInfo;
use super::costsize::{
    JoinPathExtraData, SemiAntiJoinFactors, compute_semi_anti_join_factors, cost_hashjoin, cost_material,
    cost_nestloop, has_indexed_join_quals, initial_cost_mergejoin,
};
use super::nodes::{JoinPath, JoinType, Path, PathKind, PkId, Relids, RestrictInfo, RinfoId, SpecialJoinInfo, VarNode};
use super::pathkeys::{
    build_join_pathkeys, find_mergeclauses_for_outer_pathkeys, get_cheapest_path_for_pathkeys,
    make_inner_pathkeys_for_merge, pathkeys_contained_in, pathkeys_count_contained_in, select_outer_pathkeys_for_merge,
    trim_mergeclauses_for_inner_pathkeys, update_mergeclause_eclasses,
};
use super::pathnode::{
    CostSelector, add_path, add_path_precheck, calc_nestloop_required_outer, calc_non_nestloop_required_outer,
    compare_path_costs, create_join_path, create_memoize_path, create_mergejoin_path, create_unique_path,
};
use super::restrictinfo::rinfo_is_pushed_down;
use crate::expr::Expr;

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
        JoinType::UniqueInner => sjinfo.min_lefthand.is_subset(&outerrelids),
        JoinType::UniqueOuter => super::analyzejoins::innerrel_is_unique(
            root,
            &joinrelids,
            &outerrelids,
            innerrel,
            JoinType::Inner,
            restrictlist,
            false,
        ),
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
    let mut mergejoin_allowed = true;
    let mut mergeclause_list = Vec::new();
    if root.enables.mergejoin || jointype == JoinType::Full {
        (mergeclause_list, mergejoin_allowed) =
            select_mergejoin_clauses(root, joinrel, outerrel, innerrel, restrictlist, jointype);
    }
    let semifactors = match matches!(jointype, JoinType::Semi | JoinType::Anti) || inner_unique {
        true => compute_semi_anti_join_factors(root, joinrel, outerrel, innerrel, jointype, sjinfo, restrictlist),
        false => SemiAntiJoinFactors::default(),
    };
    let extra = JoinPathExtraData {
        restrictlist: restrictlist.to_vec(),
        mergeclause_list,
        inner_unique,
        sjinfo: sjinfo.clone(),
        semifactors,
    };
    if mergejoin_allowed {
        sort_inner_and_outer(root, joinrel, outerrel, innerrel, jointype, &extra);
        match_unsorted_outer(root, joinrel, outerrel, innerrel, jointype, &extra);
    }
    if root.enables.hashjoin || jointype == JoinType::Full {
        hash_inner_and_outer(root, joinrel, outerrel, innerrel, jointype, &extra);
    }
}

/// path_param_by_rel reports whether a path needs rows of a relation's, as Postgres' PATH_PARAM_BY_REL does.
fn path_param_by_rel(root: &PlannerInfo<'_, '_>, path: &Path, rel: usize) -> bool {
    path.param.overlap(&root.rels[rel].relids)
}

/// paraminfo_get_equal_hashops returns the outer expressions that a parameterized path reads, by which a cache of its
/// rows is keyed, and whether the cache must compare them by their bytes, as Postgres' function of the same name does,
/// or None when a parameterizing clause is not a comparison of an outer and an inner expression or a lateral
/// expression is volatile. Every Doltgres type can hash its values and has one equality, which compares them.
fn paraminfo_get_equal_hashops(
    root: &PlannerInfo<'_, '_>,
    param_info: Option<&super::nodes::ParamPathInfo>,
    outerrel: usize,
    innerrel: usize,
    ph_lateral_vars: Vec<Expr>,
) -> Option<(Vec<Expr>, bool)> {
    let mut param_exprs: Vec<Expr> = Vec::new();
    let mut binary_mode = false;
    for &r in param_info.map_or(&[][..], |p| &p.ppi_clauses) {
        let rinfo = &root.rinfos[r];
        let Expr::Compare(_, left, right) = &rinfo.clause else { return None };
        if !clause_sides_match_join(rinfo, &root.rels[outerrel].relids, &root.rels[innerrel].relids) {
            return None;
        }
        let expr = if rinfo.outer_is_left.get() { left } else { right };
        if !param_exprs.contains(expr) {
            param_exprs.push((**expr).clone());
        }
        if !rinfo.hashjoinable {
            binary_mode = true;
        }
    }
    //TODO: add the inner relation's lateral_vars, once find_lateral_references is ported.
    for expr in ph_lateral_vars {
        if super::clauses::contain_volatile_functions(root.glob, &expr) {
            return None;
        }
        if !param_exprs.contains(&expr) {
            param_exprs.push(expr);
        }
        binary_mode = true;
    }
    Some((param_exprs, binary_mode))
}

/// extract_lateral_vars_from_phvs returns the lateral references of the PlaceHolderVars that a base relation computes,
/// which a cache of its rows must also be keyed by, as Postgres' extract_lateral_vars_from_PHVs does.
fn extract_lateral_vars_from_phvs(root: &PlannerInfo<'_, '_>, innerrelids: &Relids) -> Vec<Expr> {
    let mut ph_lateral_vars = Vec::new();
    if !root.has_lateral_rtes || innerrelids.num_members() > 1 {
        return ph_lateral_vars;
    }
    for phinfo in &root.placeholder_list {
        if phinfo.ph_lateral.is_empty() || phinfo.ph_eval_at != *innerrelids {
            continue;
        }
        let phexpr = &root.glob.placeholder(phinfo.phid).phexpr;
        if !super::var::pull_varnos(root, phexpr).overlap(innerrelids) {
            ph_lateral_vars.push(phexpr.clone());
            continue;
        }
        for id in super::var::pull_var_clause(root.glob, phexpr, true) {
            let lateral = match root.glob.node(id) {
                VarNode::Var(var) => phinfo.ph_lateral.is_member(var.varno),
                VarNode::PlaceHolderVar(phv) => root
                    .placeholder_list
                    .iter()
                    .find(|p| p.phid == phv.phid)
                    .is_some_and(|p| p.ph_eval_at.is_subset(&phinfo.ph_lateral)),
            };
            if lateral {
                ph_lateral_vars.push(Expr::Column(id));
            }
        }
    }
    ph_lateral_vars
}

/// get_memoize_path returns a path that caches a parameterized inner path's rows by the outer values it reads, for a
/// nested loop with an outer path, when the cache could save reading it again, as Postgres' function of the same name
/// does.
fn get_memoize_path(
    root: &mut PlannerInfo<'_, '_>,
    innerrel: usize,
    outerrel: usize,
    inner_path: &Rc<Path>,
    outer_path: &Path,
    jointype: JoinType,
    extra: &JoinPathExtraData,
) -> Option<Rc<Path>> {
    if !root.enables.memoize || root.rels[outer_path.parent].rows < 2.0 {
        return None;
    }
    let ph_lateral_vars = extract_lateral_vars_from_phvs(root, &root.rels[innerrel].relids);
    let param_info = super::relnode::get_baserel_parampathinfo(root, inner_path.parent, &inner_path.param);
    if param_info.as_ref().is_none_or(|p| p.ppi_clauses.is_empty()) && ph_lateral_vars.is_empty() {
        return None;
    }
    if !extra.inner_unique && matches!(jointype, JoinType::Semi | JoinType::Anti) {
        return None;
    }
    if extra.inner_unique {
        let param_info = param_info.as_ref()?;
        let ppi_serials: Relids = param_info.ppi_clauses.iter().map(|&r| root.rinfos[r].rinfo_serial).collect();
        if extra.restrictlist.iter().any(|&r| !ppi_serials.is_member(root.rinfos[r].rinfo_serial)) {
            return None;
        }
    }
    let volatile = |e: &Expr| super::clauses::contain_volatile_functions(root.glob, e);
    let rel = &root.rels[innerrel];
    if rel.reltarget.exprs.iter().any(volatile)
        || rel.baserestrictinfo.iter().any(|&r| volatile(&root.rinfos[r].clause))
        || param_info.iter().flat_map(|p| &p.ppi_clauses).any(|&r| volatile(&root.rinfos[r].clause))
    {
        return None;
    }
    let (param_exprs, binary_mode) =
        paraminfo_get_equal_hashops(root, param_info.as_ref(), outerrel, innerrel, ph_lateral_vars)?;
    Some(create_memoize_path(inner_path, param_exprs, binary_mode, outer_path.rows))
}

/// try_nestloop_path adds a nested loop of two paths to the join relation, whose rows are in the order of the given
/// pathkeys, as Postgres' function of the same name does.
fn try_nestloop_path(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outer: Rc<Path>,
    inner: Rc<Path>,
    pathkeys: Vec<PkId>,
    jointype: JoinType,
    extra: &JoinPathExtraData,
) {
    let ojrelid = extra.sjinfo.ojrelid;
    if ojrelid != 0 && (inner.param.is_member(ojrelid) || outer.param.is_member(ojrelid)) {
        return;
    }
    let outerrelids = &root.rels[outer.parent].relids;
    if !calc_nestloop_required_outer(outerrelids, &outer.param, &inner.param).is_empty() {
        return;
    }
    let joinrelids = root.rels[joinrel].relids.clone();
    let mut restrict_clauses = extra.restrictlist.clone();
    if inner.param.overlap(&outer.relids) {
        let enforced_serials = super::relnode::get_param_path_clause_serials(root, &inner);
        restrict_clauses.retain(|&r| !enforced_serials.is_member(root.rinfos[r].rinfo_serial));
    }
    let has_indexed_join_quals = has_indexed_join_quals(root, &joinrelids, &inner, &restrict_clauses);
    let cost = cost_nestloop(root, jointype, &outer, &inner, extra, &restrict_clauses, has_indexed_join_quals);
    let join = JoinPath { jointype, outer, inner, joinrestrictinfo: restrict_clauses };
    let path = create_join_path(joinrel, &root.rels[joinrel], PathKind::NestLoop(join), cost, pathkeys);
    add_path(&mut root.rels[joinrel], path);
}

/// try_mergejoin_path adds a merge join of two paths by merge clauses to the join relation, sorting a side first by
/// its sort keys unless its path is in that order already, when the path could be cheap enough to keep, as Postgres'
/// function of the same name does.
#[allow(clippy::too_many_arguments)]
fn try_mergejoin_path(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outer_path: Rc<Path>,
    inner_path: Rc<Path>,
    pathkeys: Vec<PkId>,
    mergeclauses: Vec<RinfoId>,
    mut outersortkeys: Vec<PkId>,
    mut innersortkeys: Vec<PkId>,
    jointype: JoinType,
    extra: &JoinPathExtraData,
) {
    let ojrelid = extra.sjinfo.ojrelid;
    if ojrelid != 0 && (inner_path.param.is_member(ojrelid) || outer_path.param.is_member(ojrelid)) {
        return;
    }
    let required_outer = calc_non_nestloop_required_outer(&outer_path, &inner_path);
    if !required_outer.is_empty() {
        return;
    }
    let mut outer_presorted_keys = 0;
    if !outersortkeys.is_empty() {
        let (contained, n_common) = pathkeys_count_contained_in(&outersortkeys, &outer_path.pathkeys);
        outer_presorted_keys = n_common;
        if contained {
            outersortkeys.clear();
        }
    }
    if !innersortkeys.is_empty() && pathkeys_contained_in(&innersortkeys, &inner_path.pathkeys) {
        innersortkeys.clear();
    }
    let workspace = initial_cost_mergejoin(
        root,
        jointype,
        &mergeclauses,
        &outer_path,
        &inner_path,
        &outersortkeys,
        &innersortkeys,
        outer_presorted_keys,
    );
    if !add_path_precheck(
        &root.rels[joinrel],
        workspace.disabled_nodes,
        workspace.startup_cost,
        workspace.total_cost,
        &pathkeys,
        &required_outer,
    ) {
        return;
    }
    let path = create_mergejoin_path(
        root,
        joinrel,
        jointype,
        &workspace,
        extra,
        outer_path,
        inner_path,
        extra.restrictlist.clone(),
        pathkeys,
        mergeclauses,
        outersortkeys,
        innersortkeys,
    );
    add_path(&mut root.rels[joinrel], path);
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
    let ojrelid = extra.sjinfo.ojrelid;
    if ojrelid != 0 && (inner.param.is_member(ojrelid) || outer.param.is_member(ojrelid)) {
        return;
    }
    if !calc_non_nestloop_required_outer(&outer, &inner).is_empty() {
        return;
    }
    let cost = cost_hashjoin(root, jointype, hashclauses, &outer, &inner, extra);
    let join = JoinPath { jointype, outer, inner, joinrestrictinfo: extra.restrictlist.clone() };
    let path = create_join_path(joinrel, &root.rels[joinrel], PathKind::HashJoin(join), cost, Vec::new());
    add_path(&mut root.rels[joinrel], path);
}

/// sort_inner_and_outer adds the merge joins of the relations' cheapest paths sorted by every mergejoinable clause,
/// once for each clause's class leading the sort, as Postgres' function of the same name does.
fn sort_inner_and_outer(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outerrel: usize,
    innerrel: usize,
    mut jointype: JoinType,
    extra: &JoinPathExtraData,
) {
    if extra.mergeclause_list.is_empty() {
        return;
    }
    let (Some(mut outer_path), Some(mut inner_path)) =
        (root.rels[outerrel].cheapest_total_path.clone(), root.rels[innerrel].cheapest_total_path.clone())
    else {
        return;
    };
    if path_param_by_rel(root, &outer_path, innerrel) || path_param_by_rel(root, &inner_path, outerrel) {
        return;
    }
    match jointype {
        JoinType::UniqueOuter => {
            outer_path = create_unique_path(root, outerrel, outer_path, &extra.sjinfo).expect("a unique outer path");
            jointype = JoinType::Inner;
        }
        JoinType::UniqueInner => {
            inner_path = create_unique_path(root, innerrel, inner_path, &extra.sjinfo).expect("a unique inner path");
            jointype = JoinType::Inner;
        }
        _ => {}
    }
    let all_pathkeys = select_outer_pathkeys_for_merge(root, &extra.mergeclause_list, joinrel);
    for (i, &front_pathkey) in all_pathkeys.iter().enumerate() {
        let outerkeys = match i {
            0 => all_pathkeys.clone(),
            _ => {
                let mut outerkeys = all_pathkeys.clone();
                outerkeys.remove(i);
                outerkeys.insert(0, front_pathkey);
                outerkeys
            }
        };
        let cur_mergeclauses = find_mergeclauses_for_outer_pathkeys(root, &outerkeys, &extra.mergeclause_list);
        let innerkeys = make_inner_pathkeys_for_merge(root, &cur_mergeclauses, &outerkeys);
        let merge_pathkeys = build_join_pathkeys(root, joinrel, jointype, &outerkeys);
        try_mergejoin_path(
            root,
            joinrel,
            outer_path.clone(),
            inner_path.clone(),
            merge_pathkeys,
            cur_mergeclauses,
            outerkeys,
            innerkeys,
            jointype,
            extra,
        );
    }
}

/// generate_mergejoin_paths adds the merge joins of an outer path, by the merge clauses its order allows, with the
/// inner relation's cheapest path sorted to match and with each of its paths already in an order that some of those
/// clauses can use, as Postgres' function of the same name does.
#[allow(clippy::too_many_arguments)]
fn generate_mergejoin_paths(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    innerrel: usize,
    outerpath: &Rc<Path>,
    jointype: JoinType,
    extra: &JoinPathExtraData,
    useallclauses: bool,
    inner_cheapest_total: &Rc<Path>,
    merge_pathkeys: &[PkId],
) {
    let save_jointype = jointype;
    let jointype = match jointype {
        JoinType::UniqueOuter | JoinType::UniqueInner => JoinType::Inner,
        other => other,
    };
    let mergeclauses = find_mergeclauses_for_outer_pathkeys(root, &outerpath.pathkeys, &extra.mergeclause_list);
    if mergeclauses.is_empty() && jointype != JoinType::Full {
        return;
    }
    if useallclauses && mergeclauses.len() != extra.mergeclause_list.len() {
        return;
    }
    let innersortkeys = make_inner_pathkeys_for_merge(root, &mergeclauses, &outerpath.pathkeys);
    try_mergejoin_path(
        root,
        joinrel,
        outerpath.clone(),
        inner_cheapest_total.clone(),
        merge_pathkeys.to_vec(),
        mergeclauses.clone(),
        Vec::new(),
        innersortkeys.clone(),
        jointype,
        extra,
    );
    if save_jointype == JoinType::UniqueInner {
        return;
    }
    let (mut cheapest_startup_inner, mut cheapest_total_inner) =
        match pathkeys_contained_in(&innersortkeys, &inner_cheapest_total.pathkeys) {
            true => (Some(inner_cheapest_total.clone()), Some(inner_cheapest_total.clone())),
            false => (None, None),
        };
    let num_sortkeys = innersortkeys.len();
    let mut trialsortkeys = innersortkeys;
    for sortkeycnt in (1..=num_sortkeys).rev() {
        trialsortkeys.truncate(sortkeycnt);
        let mut newclauses = Vec::new();
        let innerpath = get_cheapest_path_for_pathkeys(
            &root.rels[innerrel].pathlist,
            &trialsortkeys,
            &Relids::new(),
            CostSelector::Total,
        );
        if let Some(innerpath) = innerpath
            && cheapest_total_inner
                .as_ref()
                .is_none_or(|cheapest| compare_path_costs(&innerpath, cheapest, CostSelector::Total).is_lt())
        {
            newclauses = match sortkeycnt < num_sortkeys {
                true => trim_mergeclauses_for_inner_pathkeys(root, &mergeclauses, &trialsortkeys),
                false => mergeclauses.clone(),
            };
            try_mergejoin_path(
                root,
                joinrel,
                outerpath.clone(),
                innerpath.clone(),
                merge_pathkeys.to_vec(),
                newclauses.clone(),
                Vec::new(),
                Vec::new(),
                jointype,
                extra,
            );
            cheapest_total_inner = Some(innerpath);
        }
        let innerpath = get_cheapest_path_for_pathkeys(
            &root.rels[innerrel].pathlist,
            &trialsortkeys,
            &Relids::new(),
            CostSelector::Startup,
        );
        if let Some(innerpath) = innerpath
            && cheapest_startup_inner
                .as_ref()
                .is_none_or(|cheapest| compare_path_costs(&innerpath, cheapest, CostSelector::Startup).is_lt())
        {
            if cheapest_total_inner.as_ref().is_none_or(|cheapest| !Rc::ptr_eq(&innerpath, cheapest)) {
                if newclauses.is_empty() {
                    newclauses = match sortkeycnt < num_sortkeys {
                        true => trim_mergeclauses_for_inner_pathkeys(root, &mergeclauses, &trialsortkeys),
                        false => mergeclauses.clone(),
                    };
                }
                try_mergejoin_path(
                    root,
                    joinrel,
                    outerpath.clone(),
                    innerpath.clone(),
                    merge_pathkeys.to_vec(),
                    newclauses,
                    Vec::new(),
                    Vec::new(),
                    jointype,
                    extra,
                );
            }
            cheapest_startup_inner = Some(innerpath);
        }
        if useallclauses {
            break;
        }
    }
}

/// match_unsorted_outer adds the nested loops of each of the outer relation's paths over the inner relation's
/// cheapest paths, kept in memory or looking rows up, and the merge joins that its order allows, as Postgres'
/// function of the same name does.
fn match_unsorted_outer(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outerrel: usize,
    innerrel: usize,
    jointype: JoinType,
    extra: &JoinPathExtraData,
) {
    let save_jointype = jointype;
    let (nestjoin_ok, useallclauses, jointype) = match jointype {
        JoinType::RightSemi => return,
        JoinType::Inner | JoinType::Left | JoinType::Semi | JoinType::Anti => (true, false, jointype),
        JoinType::Right | JoinType::RightAnti | JoinType::Full => (false, true, jointype),
        JoinType::UniqueOuter | JoinType::UniqueInner => (true, false, JoinType::Inner),
    };
    let mut inner_cheapest_total = root.rels[innerrel].cheapest_total_path.clone();
    if inner_cheapest_total.as_ref().is_some_and(|inner| path_param_by_rel(root, inner, outerrel)) {
        inner_cheapest_total = None;
    }
    let mut matpath = None;
    if save_jointype == JoinType::UniqueInner {
        let Some(inner) = inner_cheapest_total else { return };
        inner_cheapest_total =
            Some(create_unique_path(root, innerrel, inner, &extra.sjinfo).expect("a unique inner path"));
    } else if nestjoin_ok
        && root.enables.material
        && let Some(inner) = &inner_cheapest_total
        && !exec_materializes_output(inner)
    {
        matpath = Some(create_material_path(root, inner));
    }
    let outer_paths = root.rels[outerrel].pathlist.clone();
    for mut outerpath in outer_paths {
        if path_param_by_rel(root, &outerpath, innerrel) {
            continue;
        }
        if save_jointype == JoinType::UniqueOuter {
            if root.rels[outerrel].cheapest_total_path.as_ref().is_none_or(|total| !Rc::ptr_eq(&outerpath, total)) {
                continue;
            }
            outerpath = create_unique_path(root, outerrel, outerpath, &extra.sjinfo).expect("a unique outer path");
        }
        let merge_pathkeys = build_join_pathkeys(root, joinrel, jointype, &outerpath.pathkeys);
        if save_jointype == JoinType::UniqueInner {
            let inner = inner_cheapest_total.clone().expect("a unique inner path");
            try_nestloop_path(root, joinrel, outerpath.clone(), inner, merge_pathkeys.clone(), jointype, extra);
        } else if nestjoin_ok {
            let inner_paths = root.rels[innerrel].cheapest_parameterized_paths.clone();
            for innerpath in inner_paths {
                try_nestloop_path(
                    root,
                    joinrel,
                    outerpath.clone(),
                    innerpath.clone(),
                    merge_pathkeys.clone(),
                    jointype,
                    extra,
                );
                if let Some(mpath) = get_memoize_path(root, innerrel, outerrel, &innerpath, &outerpath, jointype, extra)
                {
                    try_nestloop_path(root, joinrel, outerpath.clone(), mpath, merge_pathkeys.clone(), jointype, extra);
                }
            }
            if let Some(matpath) = &matpath {
                try_nestloop_path(
                    root,
                    joinrel,
                    outerpath.clone(),
                    matpath.clone(),
                    merge_pathkeys.clone(),
                    jointype,
                    extra,
                );
            }
        }
        if save_jointype == JoinType::UniqueOuter {
            continue;
        }
        let Some(inner_cheapest_total) = &inner_cheapest_total else { continue };
        generate_mergejoin_paths(
            root,
            joinrel,
            innerrel,
            &outerpath,
            save_jointype,
            extra,
            useallclauses,
            inner_cheapest_total,
            &merge_pathkeys,
        );
    }
}

/// exec_materializes_output reports whether a path's plan keeps its rows already, so that a Material over it gains
/// nothing, as Postgres' ExecMaterializesOutput does.
fn exec_materializes_output(path: &Path) -> bool {
    matches!(path.kind, PathKind::Material(_) | PathKind::Sort(_))
}

/// create_material_path makes a path that keeps another path's rows in memory, as Postgres' function of the same
/// name does.
fn create_material_path(root: &PlannerInfo<'_, '_>, subpath: &Rc<Path>) -> Rc<Path> {
    let (disabled_nodes, startup_cost, total_cost) = cost_material(&root.enables, subpath);
    Rc::new(Path {
        kind: PathKind::Material(subpath.clone()),
        disabled_nodes,
        startup_cost,
        total_cost,
        ..(**subpath).clone()
    })
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
    let cheapest_startup_outer = root.rels[outerrel].cheapest_startup_path.clone();
    let (Some(cheapest_total_outer), Some(cheapest_total_inner)) =
        (root.rels[outerrel].cheapest_total_path.clone(), root.rels[innerrel].cheapest_total_path.clone())
    else {
        return;
    };
    if path_param_by_rel(root, &cheapest_total_outer, innerrel)
        || path_param_by_rel(root, &cheapest_total_inner, outerrel)
    {
        return;
    }
    match jointype {
        JoinType::UniqueOuter => {
            let outer = create_unique_path(root, outerrel, cheapest_total_outer, &extra.sjinfo);
            let outer = outer.expect("a unique outer path");
            try_hashjoin_path(root, joinrel, outer, cheapest_total_inner, &hashclauses, JoinType::Inner, extra);
            return;
        }
        JoinType::UniqueInner => {
            let inner = create_unique_path(root, innerrel, cheapest_total_inner, &extra.sjinfo);
            let inner = inner.expect("a unique inner path");
            try_hashjoin_path(
                root,
                joinrel,
                cheapest_total_outer.clone(),
                inner.clone(),
                &hashclauses,
                JoinType::Inner,
                extra,
            );
            if let Some(startup) = cheapest_startup_outer
                && !Rc::ptr_eq(&startup, &cheapest_total_outer)
            {
                try_hashjoin_path(root, joinrel, startup, inner, &hashclauses, JoinType::Inner, extra);
            }
            return;
        }
        _ => {}
    }
    if let Some(cheapest_startup_outer) = &cheapest_startup_outer {
        try_hashjoin_path(
            root,
            joinrel,
            cheapest_startup_outer.clone(),
            cheapest_total_inner.clone(),
            &hashclauses,
            jointype,
            extra,
        );
    }
    let outer_paths = root.rels[outerrel].cheapest_parameterized_paths.clone();
    let inner_paths = root.rels[innerrel].cheapest_parameterized_paths.clone();
    for outerpath in &outer_paths {
        if path_param_by_rel(root, outerpath, innerrel) {
            continue;
        }
        for innerpath in &inner_paths {
            if path_param_by_rel(root, innerpath, outerrel) {
                continue;
            }
            if cheapest_startup_outer.as_ref().is_some_and(|startup| Rc::ptr_eq(outerpath, startup))
                && Rc::ptr_eq(innerpath, &cheapest_total_inner)
            {
                continue;
            }
            try_hashjoin_path(root, joinrel, outerpath.clone(), innerpath.clone(), &hashclauses, jointype, extra);
        }
    }
}

/// select_mergejoin_clauses returns the join clauses that a merge join of the relations could use, and whether a
/// merge join is allowed at all, which a right or full join is not when one of its clauses cannot be merged, as
/// Postgres' function of the same name does.
fn select_mergejoin_clauses(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outerrel: usize,
    innerrel: usize,
    restrictlist: &[RinfoId],
    jointype: JoinType,
) -> (Vec<RinfoId>, bool) {
    let mut result_list = Vec::new();
    let mut have_nonmergeable_joinclause = false;
    if jointype == JoinType::RightSemi {
        return (result_list, false);
    }
    for &rinfo in restrictlist {
        let r = &root.rinfos[rinfo];
        if jointype.is_outer() && rinfo_is_pushed_down(r, &root.rels[joinrel].relids) {
            continue;
        }
        if !r.can_join || r.mergeopfamilies.is_empty() {
            if !matches!(r.clause, Expr::Const(_)) {
                have_nonmergeable_joinclause = true;
            }
            continue;
        }
        if !clause_sides_match_join(r, &root.rels[outerrel].relids, &root.rels[innerrel].relids) {
            have_nonmergeable_joinclause = true;
            continue;
        }
        update_mergeclause_eclasses(root, rinfo);
        let r = &root.rinfos[rinfo];
        let must_be_redundant = |ec: Option<usize>| ec.is_some_and(|ec| root.eq_classes[ec].ec_has_const);
        if must_be_redundant(r.left_ec) || must_be_redundant(r.right_ec) {
            have_nonmergeable_joinclause = true;
            continue;
        }
        result_list.push(rinfo);
    }
    let mergejoin_allowed = match jointype {
        JoinType::Right | JoinType::RightAnti | JoinType::Full => !have_nonmergeable_joinclause,
        _ => true,
    };
    (result_list, mergejoin_allowed)
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
