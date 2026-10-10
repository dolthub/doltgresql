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
use super::costsize::{estimate_rel_size, get_typavgwidth, set_joinrel_size_estimates};
use super::equivclass::{generate_join_implied_equalities, has_relevant_eclass_joinclause};
use super::nodes::{Path, PathKind, RelOptInfo, RelOptKind, Relids, RinfoId, RteKind, SpecialJoinInfo, VarNode};
use super::placeholder::{add_placeholders_to_joinrel, find_placeholder_info};

/// AUTOVACUUM_ANALYZE_THRESHOLD is how many rows a table must hold for Postgres' autovacuum to have analyzed it, at
/// its default autovacuum_analyze_threshold.
const AUTOVACUUM_ANALYZE_THRESHOLD: f64 = 50.0;

/// setup_simple_rel_arrays makes a slot for the base relation of each range table entry, as Postgres' function of the
/// same name does, after a slot at index 0 so that each lies at its range table index.
pub fn setup_simple_rel_arrays(root: &mut PlannerInfo<'_, '_>) {
    root.rels = vec![RelOptInfo::default(); root.parse.rtable.len() + 1];
}

/// build_simple_rel builds the base relation of a range table entry, as Postgres' function of the same name does,
/// sized as get_relation_info and estimate_rel_size size a table, with the statistics of a table that was analyzed,
/// where autovacuum analyzes a table that holds enough rows. Any other entry's rows are the older planner's estimate.
pub fn build_simple_rel(root: &mut PlannerInfo<'_, '_>, relid: usize) {
    let rte = root.parse.rte(relid).clone();
    let width = rte.coltypes.len();
    let mut rel = RelOptInfo {
        reloptkind: RelOptKind::BaseRel,
        relids: Relids::singleton(relid),
        consider_startup: root.tuple_fraction > 0.0,
        relid,
        baserestrict_min_security: usize::MAX,
        attr_needed: vec![Relids::new(); width],
        attr_widths: vec![0.0; width],
        ..RelOptInfo::default()
    };
    match &rte.kind {
        RteKind::Relation(_, table) => {
            let rows = prolly::Node::decode(table.table.primary_index.clone()).map_or(0.0, |r| r.tree_count() as f64);
            let data_width = table.columns.iter().map(|c| get_typavgwidth(Some(c.ty.oid), c.ty.modifier)).sum();
            let session = &root.ctx.session;
            let vacuumed = session.engine.vacuumed(&session.database, &table.schema, &table.name);
            let autovacuumed = rows > AUTOVACUUM_ANALYZE_THRESHOLD;
            (rel.pages, rel.tuples) = estimate_rel_size(rows, data_width, autovacuumed || vacuumed.is_some());
            let analyzed = autovacuumed || vacuumed == Some(true);
            rel.stats = analyzed.then(|| crate::colstats::table_stats(root.ctx, table)).flatten();
            rel.notnullattnums = (0..table.columns.len()).filter(|&c| !table.columns[c].nullable).collect();
        }
        RteKind::Plan(plan) => rel.tuples = crate::joins::estimate(root.ctx, plan),
        RteKind::Result => rel.tuples = 1.0,
        RteKind::Subquery(..) | RteKind::Join(_) => unreachable!("only base relations are built"),
    }
    root.rels[relid] = rel;
    if let RteKind::Relation(_, table) = &rte.kind {
        root.rels[relid].indexlist = super::indxpath::get_relation_indexes(root, relid, table);
    }
}

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
    root.rels[joinrel].reltarget.width = tuple_width;
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
    match &path.kind {
        PathKind::NestLoop(join) | PathKind::HashJoin(join) => {
            let mut pserials = get_param_path_clause_serials(root, &join.outer);
            pserials.add_members(&get_param_path_clause_serials(root, &join.inner));
            for &r in &join.joinrestrictinfo {
                pserials.add_member(root.rinfos[r].rinfo_serial);
            }
            pserials
        }
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
    if let Some(ppi) = root.rels[rel].ppilist.iter().find(|p| p.ppi_req_outer == *required_outer) {
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

/// fetch_upper_rel returns the relation of a step of the query's upper processing, building it the first time, as
/// Postgres' function of the same name does.
pub fn fetch_upper_rel(root: &mut PlannerInfo<'_, '_>, kind: super::nodes::UpperRelationKind) -> usize {
    if let Some(rel) = root.rels.iter().position(|r| r.reloptkind == RelOptKind::UpperRel(kind)) {
        return rel;
    }
    root.rels.push(RelOptInfo { reloptkind: RelOptKind::UpperRel(kind), ..RelOptInfo::default() });
    root.rels.len() - 1
}
