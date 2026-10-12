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

//! Postgres' optimizer/util/relnode.c, with the parts of plancat.c that describe a table: building the base
//! relations and the join relations.

use super::PlannerInfo;
use super::costsize::{clamp_row_est, set_joinrel_size_estimates};
use super::equivclass::{generate_join_implied_equalities, has_relevant_eclass_joinclause};
use super::nodes::{Path, PathKind, RelOptInfo, RelOptKind, Relids, RinfoId, RteKind, SpecialJoinInfo, VarNode};
use super::placeholder::{add_placeholders_to_joinrel, find_placeholder_info};
use crate::plan::Plan;

/// setup_simple_rel_arrays makes a slot for the base relation of each range table entry, as Postgres' function of the
/// same name does, after a slot at index 0 so that each lies at its range table index.
pub fn setup_simple_rel_arrays(root: &mut PlannerInfo<'_, '_>) {
    root.rels = vec![RelOptInfo::default(); root.parse.rtable.len() + 1];
}

/// build_simple_rel builds the base relation of a range table entry, or the relation of a member of an append
/// relation with its parent's restrictions, as Postgres' function of the same name does, sized as get_relation_info
/// and estimate_rel_size size a table, with the statistics of a table that was analyzed, where autovacuum analyzes a
/// table that holds enough rows. Any other entry's rows are the estimate of `joins::estimate`.
pub fn build_simple_rel(root: &mut PlannerInfo<'_, '_>, relid: usize, parent: Option<usize>) {
    let rte = root.parse.rte(relid).clone();
    let width = rte.coltypes.len();
    let mut rel = RelOptInfo {
        reloptkind: match parent {
            Some(_) => RelOptKind::OtherMemberRel,
            None => RelOptKind::BaseRel,
        },
        relids: Relids::singleton(relid),
        consider_startup: root.tuple_fraction > 0.0,
        relid,
        baserestrict_min_security: usize::MAX,
        attr_needed: vec![Relids::new(); width],
        attr_widths: vec![0.0; width],
        ..RelOptInfo::default()
    };
    if let Some(parent) = parent {
        let parent = &root.rels[parent];
        rel.top_parent_relids = match parent.top_parent_relids.is_empty() {
            true => parent.relids.clone(),
            false => parent.top_parent_relids.clone(),
        };
        rel.nulling_relids = parent.nulling_relids.clone();
        rel.direct_lateral_relids = parent.direct_lateral_relids.clone();
        rel.lateral_relids = parent.lateral_relids.clone();
        rel.lateral_referencers = parent.lateral_referencers.clone();
    }
    match &rte.kind {
        RteKind::Plan(Plan::WorkTable(..)) if let Some(rows) = root.glob.non_recursive_rows => {
            rel.tuples = clamp_row_est(RECURSIVE_WORKTABLE_FACTOR * rows);
        }
        RteKind::Plan(plan) => rel.tuples = crate::joins::estimate(root.ctx.get_mut(), plan),
        RteKind::Result => rel.tuples = 1.0,
        RteKind::Relation(..) | RteKind::Subquery(..) => {}
        RteKind::Join(_) => unreachable!("only base relations are built"),
    }
    root.rels[relid] = rel;
    if let RteKind::Relation(_, table) = &rte.kind {
        super::plancat::get_relation_info(root, relid, table);
    }
    if let Some(parent) = parent {
        let appinfo = super::appendinfo::member_appinfo(root, relid);
        if !super::inherit::apply_child_basequals(root, parent, relid, &appinfo) {
            super::joinrels::mark_dummy_rel(root, relid);
        }
    }
}

/// find_childrel_parents returns the append relations that a member of one belongs to, up to the topmost, as
/// Postgres' function of the same name does.
pub fn find_childrel_parents(root: &PlannerInfo<'_, '_>, rel: usize) -> Relids {
    let mut result = Relids::new();
    let mut relid = rel;
    while root.rels[relid].reloptkind == RelOptKind::OtherMemberRel {
        relid = super::appendinfo::member_appinfo(root, relid).parent_relid;
        result.add_member(relid);
    }
    result
}

/// RECURSIVE_WORKTABLE_FACTOR is how many times a recursive WITH query's working table is the size of its
/// non-recursive term, as Postgres' recursive_worktable_factor setting defaults to and set_cte_size_estimates reads.
const RECURSIVE_WORKTABLE_FACTOR: f64 = 10.0;

/// find_join_rel returns the index of the join relation of a set of relations, when it was built, as Postgres'
/// function of the same name does.
pub fn find_join_rel(root: &PlannerInfo<'_, '_>, relids: &Relids) -> Option<usize> {
    root.join_rel_hash.get(relids).copied()
}

/// build_join_rel returns the join relation of two relations, building it with its target, clauses, and size when
/// it is new, with the clauses that the join evaluates, as Postgres' function of the same name does.
pub fn build_join_rel(
    root: &mut PlannerInfo<'_, '_>,
    joinrelids: Relids,
    outer_rel: usize,
    inner_rel: usize,
    sjinfo: &SpecialJoinInfo,
    pushed_down_joins: &[SpecialJoinInfo],
) -> (usize, Vec<RinfoId>) {
    if let Some(joinrel) = find_join_rel(root, &joinrelids) {
        return (joinrel, build_joinrel_restrictlist(root, joinrel, outer_rel, inner_rel, sjinfo));
    }
    let lateral_relids = min_join_parameterization(root, &joinrelids, outer_rel, inner_rel);
    let joinrel = RelOptInfo {
        reloptkind: RelOptKind::JoinRel,
        relids: joinrelids.clone(),
        consider_startup: root.tuple_fraction > 0.0,
        direct_lateral_relids: root.rels[outer_rel]
            .direct_lateral_relids
            .union(&root.rels[inner_rel].direct_lateral_relids),
        lateral_relids,
        baserestrict_min_security: usize::MAX,
        ..RelOptInfo::default()
    };
    root.rels.push(joinrel);
    let joinrel = root.rels.len() - 1;
    build_joinrel_tlist(
        root,
        joinrel,
        outer_rel,
        sjinfo,
        pushed_down_joins,
        sjinfo.jointype == super::nodes::JoinType::Full,
    );
    build_joinrel_tlist(
        root,
        joinrel,
        inner_rel,
        sjinfo,
        pushed_down_joins,
        sjinfo.jointype != super::nodes::JoinType::Inner,
    );
    add_placeholders_to_joinrel(root, joinrel, outer_rel, inner_rel);
    root.rels[joinrel].direct_lateral_relids.del_members(&joinrelids);
    let restrictlist = build_joinrel_restrictlist(root, joinrel, outer_rel, inner_rel, sjinfo);
    build_joinrel_joinlist(root, joinrel, outer_rel, inner_rel);
    root.rels[joinrel].has_eclass_joins = has_relevant_eclass_joinclause(root, joinrel);
    set_joinrel_size_estimates(root, joinrel, outer_rel, inner_rel, sjinfo, &restrictlist);
    root.join_rel_hash.insert(joinrelids, joinrel);
    if let Some(level) = root.join_rel_level.get_mut(root.join_cur_level) {
        level.push(joinrel);
    }
    (joinrel, restrictlist)
}

/// min_join_parameterization returns the relations that a join of two relations must read laterally, as Postgres'
/// function of the same name does.
pub fn min_join_parameterization(
    root: &PlannerInfo<'_, '_>,
    joinrelids: &Relids,
    outer_rel: usize,
    inner_rel: usize,
) -> Relids {
    root.rels[outer_rel].lateral_relids.union(&root.rels[inner_rel].lateral_relids).difference(joinrelids)
}

/// build_joinrel_tlist adds the Vars and PlaceHolderVars of an input relation's target that joins above or the query's
/// output read to a join relation's target, adding the outer joins that the join completes to their nulling relations
/// on the side that the join can make NULL, as Postgres' function of the same name does.
fn build_joinrel_tlist(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    input_rel: usize,
    sjinfo: &SpecialJoinInfo,
    pushed_down_joins: &[SpecialJoinInfo],
    can_null: bool,
) {
    let relids = root.rels[joinrel].relids.clone();
    let mut tuple_width = root.rels[joinrel].reltarget.width;
    for e in root.rels[input_rel].reltarget.exprs.clone() {
        let crate::expr::Expr::Column(id) = e else { unreachable!("a relation's target holds Vars") };
        let mut node = root.glob.node(id).clone();
        let (nullingrels, rels) = match &mut node {
            VarNode::PlaceHolderVar(phv) => {
                let i = find_placeholder_info(root, phv.phid);
                let phinfo = &root.placeholder_list[i];
                if !phinfo.ph_needed.nonempty_difference(&relids) {
                    continue;
                }
                tuple_width += phinfo.ph_width;
                (&mut phv.phnullingrels, root.glob.placeholder(phv.phid).phrels.clone())
            }
            VarNode::Var(var) => {
                let baserel = &root.rels[var.varno];
                if !baserel.attr_needed[var.varattno].nonempty_difference(&relids) {
                    continue;
                }
                tuple_width += baserel.attr_widths[var.varattno];
                (&mut var.varnullingrels, Relids::singleton(var.varno))
            }
        };
        if can_null {
            if sjinfo.ojrelid != 0
                && relids.is_member(sjinfo.ojrelid)
                && (rels.is_subset(&sjinfo.syn_righthand)
                    || (sjinfo.jointype == super::nodes::JoinType::Full && rels.is_subset(&sjinfo.syn_lefthand)))
            {
                nullingrels.add_member(sjinfo.ojrelid);
            }
            for othersj in pushed_down_joins {
                if rels.is_subset(&othersj.syn_righthand) {
                    nullingrels.add_member(othersj.ojrelid);
                }
            }
            nullingrels.add_members(&sjinfo.commute_above_r.intersect(&relids));
        }
        let e = root.glob.intern(node);
        root.rels[joinrel].reltarget.exprs.push(e);
    }
    root.rels[joinrel].reltarget.width = super::costsize::clamp_width_est(tuple_width);
}

/// build_joinrel_restrictlist returns the join clauses of the two relations that a join of them evaluates, with the
/// equalities that equivalence classes imply between them, as Postgres' function of the same name does.
fn build_joinrel_restrictlist(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outer_rel: usize,
    inner_rel: usize,
    sjinfo: &SpecialJoinInfo,
) -> Vec<RinfoId> {
    let both_input_relids = root.rels[outer_rel].relids.union(&root.rels[inner_rel].relids);
    let mut result = subbuild_joinrel_restrictlist(root, joinrel, outer_rel, &both_input_relids, Vec::new());
    result = subbuild_joinrel_restrictlist(root, joinrel, inner_rel, &both_input_relids, result);
    let (joinrelids, outer_relids) = (root.rels[joinrel].relids.clone(), root.rels[outer_rel].relids.clone());
    result.extend(generate_join_implied_equalities(root, &joinrelids, &outer_relids, inner_rel, sjinfo.ojrelid));
    result
}

/// build_joinrel_joinlist gives a join relation the join clauses of its inputs that it cannot evaluate, as Postgres'
/// function of the same name does.
fn build_joinrel_joinlist(root: &mut PlannerInfo<'_, '_>, joinrel: usize, outer_rel: usize, inner_rel: usize) {
    let joinrelids = root.rels[joinrel].relids.clone();
    let mut result: Vec<RinfoId> = Vec::new();
    for rel in [outer_rel, inner_rel] {
        for &rinfo in &root.rels[rel].joininfo {
            if !root.rinfos[rinfo].required_relids.is_subset(&joinrelids) && !result.contains(&rinfo) {
                result.push(rinfo);
            }
        }
    }
    root.rels[joinrel].joininfo = result;
}

/// subbuild_joinrel_restrictlist adds the join clauses of an input relation that a join evaluates to a list, where a
/// clone of an outer join clause is evaluated only at a join of exactly its relations that it is compatible with, as
/// Postgres' function of the same name does.
fn subbuild_joinrel_restrictlist(
    root: &PlannerInfo<'_, '_>,
    joinrel: usize,
    input_rel: usize,
    both_input_relids: &Relids,
    mut new_restrictlist: Vec<RinfoId>,
) -> Vec<RinfoId> {
    let joinrelids = &root.rels[joinrel].relids;
    for &rinfo in &root.rels[input_rel].joininfo {
        let r = &root.rinfos[rinfo];
        if !r.required_relids.is_subset(joinrelids) {
            continue;
        }
        if (r.has_clone || r.is_clone)
            && (!r.required_relids.is_subset(both_input_relids) || r.incompatible_relids.overlap(both_input_relids))
        {
            continue;
        }
        if !new_restrictlist.contains(&rinfo) {
            new_restrictlist.push(rinfo);
        }
    }
    new_restrictlist
}

/// get_param_path_clause_serials returns the serial numbers of the join clauses that a parameterized path tests, as
/// Postgres' function of the same name does: those of a base relation's parameterization, and for a join, those of its
/// inputs and its own clauses.
pub fn get_param_path_clause_serials(root: &mut PlannerInfo<'_, '_>, path: &Path) -> Relids {
    if path.param.is_empty() {
        return Relids::new();
    }
    if let Some(join) = path.kind.join() {
        let mut pserials = get_param_path_clause_serials(root, &join.outer);
        pserials.add_members(&get_param_path_clause_serials(root, &join.inner));
        for &r in &join.joinrestrictinfo {
            pserials.add_member(root.rinfos[r].rinfo_serial);
        }
        return pserials;
    }
    match &path.kind {
        PathKind::Append(subpaths) => {
            let mut serials = subpaths.iter().map(|subpath| get_param_path_clause_serials(root, subpath));
            let first = serials.next().unwrap_or_default();
            serials.fold(first, |pserials, subserials| pserials.intersect(&subserials))
        }
        _ => {
            let ppi = get_baserel_parampathinfo(root, path.parent, &path.param);
            ppi.map_or_else(Relids::new, |ppi| ppi.ppi_clauses.iter().map(|&r| root.rinfos[r].rinfo_serial).collect())
        }
    }
}

/// get_baserel_parampathinfo returns what a parameterization by outer relations gives a base relation's paths, building
/// it when it is new: the join clauses movable into the relation and the equalities that classes imply with the outer
/// relations, and the rows that remain, as Postgres' function of the same name does.
pub fn get_baserel_parampathinfo(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    required_outer: &Relids,
) -> Option<super::nodes::ParamPathInfo> {
    if required_outer.is_empty() {
        return None;
    }
    if let Some(ppi) = find_param_path_info(&root.rels[rel], required_outer) {
        return Some(ppi.clone());
    }
    let baserelids = root.rels[rel].relids.clone();
    let joinrelids = baserelids.union(required_outer);
    let mut pclauses: Vec<RinfoId> = root.rels[rel]
        .joininfo
        .iter()
        .copied()
        .filter(|&r| super::restrictinfo::join_clause_is_movable_into(&root.rinfos[r], &baserelids, &joinrelids))
        .collect();
    pclauses.extend(generate_join_implied_equalities(root, &joinrelids, required_outer, rel, 0));
    let rows = super::costsize::get_parameterized_baserel_size(root, rel, &pclauses);
    let ppi =
        super::nodes::ParamPathInfo { ppi_req_outer: required_outer.clone(), ppi_rows: rows, ppi_clauses: pclauses };
    root.rels[rel].ppilist.push(ppi.clone());
    Some(ppi)
}

/// get_joinrel_parampathinfo returns what a parameterization gives a join of two paths, building it when it is new,
/// after adding to the join's clauses those join clauses that the parameterization moves into the join but into
/// neither input path, including the equalities of classes whose clause the inner path took, as Postgres' function of
/// the same name does. An unparameterized join has none.
pub fn get_joinrel_parampathinfo(
    root: &mut PlannerInfo<'_, '_>,
    joinrel: usize,
    outer_path: &Path,
    inner_path: &Path,
    sjinfo: &SpecialJoinInfo,
    required_outer: &Relids,
    restrict_clauses: &mut Vec<RinfoId>,
) -> Option<super::nodes::ParamPathInfo> {
    use super::restrictinfo::join_clause_is_movable_into;
    if required_outer.is_empty() {
        return None;
    }
    let joinrelids = root.rels[joinrel].relids.clone();
    let join_and_req = joinrelids.union(required_outer);
    let outer_relids = root.rels[outer_path.parent].relids.clone();
    let inner_relids = root.rels[inner_path.parent].relids.clone();
    let outer_and_req = match outer_path.param.is_empty() {
        true => Relids::new(),
        false => outer_relids.union(&outer_path.param),
    };
    let inner_and_req = match inner_path.param.is_empty() {
        true => Relids::new(),
        false => inner_relids.union(&inner_path.param),
    };
    let mut pclauses: Vec<RinfoId> = root.rels[joinrel]
        .joininfo
        .iter()
        .copied()
        .filter(|&r| {
            let rinfo = &root.rinfos[r];
            join_clause_is_movable_into(rinfo, &joinrelids, &join_and_req)
                && !join_clause_is_movable_into(rinfo, &outer_relids, &outer_and_req)
                && !join_clause_is_movable_into(rinfo, &inner_relids, &inner_and_req)
        })
        .collect();
    let eclauses = generate_join_implied_equalities(root, &join_and_req, required_outer, joinrel, 0);
    let mut dropped_ecs = Vec::new();
    for r in eclauses {
        let rinfo = &root.rinfos[r];
        if join_clause_is_movable_into(rinfo, &outer_relids, &outer_and_req) {
            continue;
        }
        if join_clause_is_movable_into(rinfo, &inner_relids, &inner_and_req) {
            dropped_ecs.extend(rinfo.left_ec);
            continue;
        }
        pclauses.push(r);
    }
    if !dropped_ecs.is_empty() {
        let real_outer_and_req = outer_relids.union(required_outer);
        let eclauses = super::equivclass::generate_join_implied_equalities_for_ecs(
            root,
            &dropped_ecs,
            &real_outer_and_req,
            required_outer,
            outer_path.parent,
        );
        for r in eclauses {
            if !join_clause_is_movable_into(&root.rinfos[r], &outer_relids, &outer_and_req) {
                pclauses.push(r);
            }
        }
    }
    pclauses.append(restrict_clauses);
    *restrict_clauses = pclauses;
    if let Some(ppi) = find_param_path_info(&root.rels[joinrel], required_outer) {
        return Some(ppi.clone());
    }
    let rows = super::costsize::get_parameterized_joinrel_size(
        root,
        joinrel,
        outer_path,
        inner_path,
        sjinfo,
        restrict_clauses,
    );
    let ppi =
        super::nodes::ParamPathInfo { ppi_req_outer: required_outer.clone(), ppi_rows: rows, ppi_clauses: Vec::new() };
    root.rels[joinrel].ppilist.push(ppi.clone());
    Some(ppi)
}

/// get_appendrel_parampathinfo returns the parameterization of an append relation's paths by a set of outer
/// relations, which moves no join clauses since the relation's members test them, as Postgres' function of the same
/// name does.
pub fn get_appendrel_parampathinfo(
    root: &mut PlannerInfo<'_, '_>,
    appendrel: usize,
    required_outer: &Relids,
) -> Option<super::nodes::ParamPathInfo> {
    if required_outer.is_empty() {
        return None;
    }
    if let Some(ppi) = find_param_path_info(&root.rels[appendrel], required_outer) {
        return Some(ppi.clone());
    }
    let ppi =
        super::nodes::ParamPathInfo { ppi_req_outer: required_outer.clone(), ppi_rows: 0.0, ppi_clauses: Vec::new() };
    root.rels[appendrel].ppilist.push(ppi.clone());
    Some(ppi)
}

/// find_param_path_info returns what a parameterization gives a relation's paths when it was built, as Postgres'
/// function of the same name does.
pub fn find_param_path_info<'r>(
    rel: &'r RelOptInfo,
    required_outer: &Relids,
) -> Option<&'r super::nodes::ParamPathInfo> {
    rel.ppilist.iter().find(|ppi| ppi.ppi_req_outer == *required_outer)
}

/// fetch_upper_rel returns the relation of a step of the query's upper processing over a set of relations, which
/// only a set operation's steps tell apart, building it the first time, as Postgres' function of the same name does.
pub fn fetch_upper_rel(
    root: &mut PlannerInfo<'_, '_>,
    kind: super::nodes::UpperRelationKind,
    relids: &Relids,
) -> usize {
    if let Some(rel) = root.rels.iter().position(|r| r.reloptkind == RelOptKind::UpperRel(kind) && r.relids == *relids)
    {
        return rel;
    }
    root.rels.push(RelOptInfo {
        reloptkind: RelOptKind::UpperRel(kind),
        relids: relids.clone(),
        ..RelOptInfo::default()
    });
    root.rels.len() - 1
}
