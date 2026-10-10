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

//! Postgres' optimizer/plan/initsplan.c: building the base relations and their target lists, and breaking the join
//! tree into the relations to join, the clauses that restrict them, the equivalence classes of their equalities, and
//! the outer joins that constrain their join order, with the clones of outer join clauses that let outer joins
//! commute.

use std::collections::{BTreeMap, BTreeSet};

use super::PlannerInfo;
use super::clauses::{contain_volatile_functions, find_nonnullable_rels};
use super::equivclass::process_equivalence;
use super::nodefuncs::{btree_opfamily, expr_type};
use super::nodes::{
    JoinDomain, JoinTreeNode, JoinType, OuterJoinClauseInfo, RelOptKind, Relids, RestrictInfo, RinfoId, RteKind, SjId,
    SpecialJoinInfo, VarNode,
};
use super::pathkeys::initialize_mergeclause_eclasses;
use super::placeholder::{contain_placeholder_references_to, find_placeholder_info};
use super::relnode::build_simple_rel;
use super::restrictinfo::{RestrictInfoArgs, binary_op_args, make_restrictinfo, restriction_is_or_clause};
use super::tlist::get_sortgroupclause_expr;
use super::var::{add_nulling_relids, pull_var_clause, pull_varnos, pull_varnos_list, remove_nulling_relids_fn};
use crate::expr::{CmpOp, Expr};

/// JoinList is a list of relations to join in any order, where a member is a range table index or a list that must
/// be joined on its own first, as Postgres' joinlists are.
#[derive(Clone, Debug)]
pub enum JoinList {
    Rel(usize),
    List(Vec<JoinList>),
}

/// JOIN_COLLAPSE_LIMIT and FROM_COLLAPSE_LIMIT are the most members that a joinlist collapses into one search
/// problem, as Postgres' settings of the same names default to.
const JOIN_COLLAPSE_LIMIT: usize = 8;
const FROM_COLLAPSE_LIMIT: usize = 8;

/// add_base_rels_to_query builds the base relation of each range table entry that the join tree references, as
/// Postgres' function of the same name does.
pub fn add_base_rels_to_query(root: &mut PlannerInfo<'_, '_>, node: &JoinTreeNode) {
    match node {
        JoinTreeNode::Rel(varno) => build_simple_rel(root, *varno),
        JoinTreeNode::From(f) => f.fromlist.iter().for_each(|n| add_base_rels_to_query(root, n)),
        JoinTreeNode::Join(j) => {
            add_base_rels_to_query(root, &j.larg);
            add_base_rels_to_query(root, &j.rarg);
        }
    }
}

/// remove_useless_groupby_columns drops the GROUP BY columns of a table that its other GROUP BY columns determine,
/// since those hold every column of a unique index whose columns are NOT NULL, as Postgres' function of the same name
/// does.
pub fn remove_useless_groupby_columns(root: &mut PlannerInfo<'_, '_>) {
    if root.processed_group_clause.len() < 2 || root.parse.grouping_sets.is_some() {
        return;
    }
    let mut groupbycols: BTreeMap<usize, Vec<(usize, Option<u32>)>> = BTreeMap::new();
    let mut tryremove = false;
    for sgc in &root.processed_group_clause {
        let expr = get_sortgroupclause_expr(sgc, &root.parse.target_list);
        let Expr::Column(id) = expr else { continue };
        let VarNode::Var(var) = root.glob.node(id) else { continue };
        let eq_opfamily = expr_type(root, &expr).and_then(btree_opfamily);
        let cols = groupbycols.entry(var.varno).or_default();
        tryremove |= cols.iter().any(|&(attno, _)| attno != var.varattno);
        cols.push((var.varattno, eq_opfamily));
    }
    if !tryremove {
        return;
    }
    let mut surplusvars: BTreeMap<usize, BTreeSet<usize>> = BTreeMap::new();
    for (&relid, cols) in &groupbycols {
        if !matches!(root.parse.rte(relid).kind, RteKind::Relation(..)) {
            continue;
        }
        let relattnos: BTreeSet<usize> = cols.iter().map(|&(attno, _)| attno).collect();
        if relattnos.len() < 2 {
            continue;
        }
        let rel = &root.rels[relid];
        let mut best_keycolumns: Option<(usize, BTreeSet<usize>)> = None;
        for index in &rel.indexlist {
            if !index.unique || !index.indpred.is_empty() || !index.indexprs.is_empty() {
                continue;
            }
            let mut ind_attnos = BTreeSet::new();
            let index_check_ok = (0..index.nkeycolumns).all(|i| {
                let Some(attno) = index.indexkeys[i] else { return false };
                let matched = rel.notnullattnums.contains(&attno)
                    && cols.iter().any(|&(a, family)| a == attno && family.is_some() && family == index.opfamily[i]);
                ind_attnos.insert(attno);
                matched
            });
            if !index_check_ok || ind_attnos.len() >= relattnos.len() || !ind_attnos.is_subset(&relattnos) {
                continue;
            }
            if best_keycolumns.as_ref().is_none_or(|(n, _)| index.nkeycolumns < *n) {
                best_keycolumns = Some((index.nkeycolumns, ind_attnos));
            }
        }
        if let Some((_, keycolumns)) = best_keycolumns {
            surplusvars.insert(relid, relattnos.difference(&keycolumns).copied().collect());
        }
    }
    if surplusvars.is_empty() {
        return;
    }
    let target_list = &root.parse.target_list;
    let glob = &root.glob;
    root.processed_group_clause.retain(|sgc| match get_sortgroupclause_expr(sgc, target_list) {
        Expr::Column(id) => match glob.node(id) {
            VarNode::Var(var) => !surplusvars.get(&var.varno).is_some_and(|vars| vars.contains(&var.varattno)),
            VarNode::PlaceHolderVar(_) => true,
        },
        _ => true,
    });
}

/// find_lateral_references records the Vars of other relations that each lateral base relation reads and marks them
/// as needed by it, as Postgres' function of the same name does.
pub fn find_lateral_references(root: &mut PlannerInfo<'_, '_>) {
    for rti in 1..=root.parse.rtable.len() {
        if root.rels.get(rti).is_none_or(|rel| rel.reloptkind != RelOptKind::BaseRel) {
            continue;
        }
        extract_lateral_references(root, rti);
    }
}

/// extract_lateral_references records the Vars of other relations that a lateral base relation reads, in its own
/// expressions, and marks them as needed by it, as Postgres' function of the same name does.
fn extract_lateral_references(root: &mut PlannerInfo<'_, '_>, rtindex: usize) {
    let rte = root.parse.rte(rtindex);
    let RteKind::Plan(plan) = &rte.kind else { return };
    if !rte.lateral {
        return;
    }
    let mut vars = Vec::new();
    let mut plan = plan.clone();
    plan.map_exprs(0, &mut |e, depth| {
        if depth == 0 {
            for id in pull_var_clause(root.glob, &e, true) {
                if !vars.contains(&id) {
                    vars.push(id);
                }
            }
        }
        e
    });
    if vars.is_empty() {
        return;
    }
    root.has_lateral_rtes = true;
    add_vars_to_targetlist(root, &vars, &Relids::singleton(rtindex));
    root.rels[rtindex].lateral_vars = vars.into_iter().map(Expr::Column).collect();
}

/// create_lateral_join_info sets the relations that each base relation reads laterally, directly or through the
/// relations it reads, and those that read it, as Postgres' function of the same name does.
pub fn create_lateral_join_info(root: &mut PlannerInfo<'_, '_>) {
    if !root.has_lateral_rtes {
        return;
    }
    let mut found_laterals = false;
    let base_rels: Vec<usize> = (1..=root.parse.rtable.len())
        .filter(|&rti| root.rels.get(rti).is_some_and(|rel| rel.reloptkind == RelOptKind::BaseRel && rel.relid == rti))
        .collect();
    for &rti in &base_rels {
        let mut lateral_relids = Relids::new();
        for e in root.rels[rti].lateral_vars.clone() {
            let Expr::Column(id) = e else { continue };
            found_laterals = true;
            match root.glob.node(id).clone() {
                VarNode::Var(var) => lateral_relids.add_member(var.varno),
                VarNode::PlaceHolderVar(phv) => {
                    let i = find_placeholder_info(root, phv.phid);
                    lateral_relids.add_members(&root.placeholder_list[i].ph_eval_at);
                }
            }
        }
        root.rels[rti].direct_lateral_relids = lateral_relids.clone();
        root.rels[rti].lateral_relids = lateral_relids;
    }
    for phinfo in root.placeholder_list.clone() {
        if phinfo.ph_lateral.is_empty() {
            continue;
        }
        found_laterals = true;
        let lateral_refs = phinfo.ph_lateral.intersect(&root.all_baserels);
        match phinfo.ph_eval_at.singleton_member() {
            Some(varno) => {
                root.rels[varno].direct_lateral_relids.add_members(&lateral_refs);
                root.rels[varno].lateral_relids.add_members(&lateral_refs);
            }
            None => {
                for varno in phinfo.ph_eval_at.members().filter(|varno| base_rels.contains(varno)) {
                    root.rels[varno].lateral_relids.add_members(&lateral_refs);
                }
            }
        }
    }
    if !found_laterals {
        root.has_lateral_rtes = false;
        return;
    }
    for &rti in &base_rels {
        let outer_lateral_relids = root.rels[rti].lateral_relids.clone();
        if outer_lateral_relids.is_empty() {
            continue;
        }
        for &rti2 in &base_rels {
            if root.rels[rti2].lateral_relids.is_member(rti) {
                root.rels[rti2].lateral_relids.add_members(&outer_lateral_relids);
            }
        }
    }
    for &rti in &base_rels {
        for rti2 in root.rels[rti].lateral_relids.clone().members() {
            if base_rels.contains(&rti2) {
                root.rels[rti2].lateral_referencers.add_member(rti);
            }
        }
    }
}

/// build_base_rel_tlists marks the Vars of the query's output as needed by it, as Postgres' function of the same
/// name does.
pub fn build_base_rel_tlists(root: &mut PlannerInfo<'_, '_>, final_tlist: &[Expr]) {
    let vars: Vec<usize> = final_tlist.iter().flat_map(|e| pull_var_clause(root.glob, e, true)).collect();
    if !vars.is_empty() {
        add_vars_to_targetlist(root, &vars, &Relids::singleton(0));
    }
}

/// add_vars_to_targetlist adds Vars and PlaceHolderVars to the targets of the base relations that compute them and
/// marks them as needed by a set of relations, as Postgres' function of the same name does.
pub fn add_vars_to_targetlist(root: &mut PlannerInfo<'_, '_>, vars: &[usize], where_needed: &Relids) {
    for &id in vars {
        match root.glob.node(id).clone() {
            VarNode::Var(var) => {
                let rel = &root.rels[var.varno];
                if where_needed.is_subset(&rel.relids) {
                    continue;
                }
                if rel.attr_needed[var.varattno].is_empty() {
                    let plain = root.glob.var(var.varno, var.varattno, Relids::new());
                    root.rels[var.varno].reltarget.exprs.push(plain);
                }
                root.rels[var.varno].attr_needed[var.varattno].add_members(where_needed);
            }
            VarNode::PlaceHolderVar(phv) => {
                let i = find_placeholder_info(root, phv.phid);
                root.placeholder_list[i].ph_needed.add_members(where_needed);
            }
        }
    }
}

/// ItemKind is the join tree node of a JoinTreeItem.
#[derive(Clone)]
enum ItemKind {
    Rel,
    From(Vec<Expr>),
    Join(JoinType, Vec<Expr>, usize),
}

/// JoinTreeItem is what deconstruct_jointree learns of a join tree node, as Postgres' JoinTreeItem is.
#[derive(Clone)]
struct JoinTreeItem {
    kind: ItemKind,
    jdomain: usize,
    qualscope: Relids,
    inner_join_rels: Relids,
    left_rels: Relids,
    right_rels: Relids,
    nonnullable_rels: Relids,
    sjinfo: Option<SjId>,
    oj_joinclauses: Vec<Expr>,
}

/// deconstruct_jointree distributes the query's clauses to the relations they restrict and records its outer joins,
/// returning the joinlist of relations to join, as Postgres' deconstruct_jointree does: one pass builds the join
/// domains and relation sets of each node, a second distributes each node's clauses, and a third distributes the
/// clauses of left joins that can commute with others, with clones for each order.
pub fn deconstruct_jointree(root: &mut PlannerInfo<'_, '_>) -> Vec<JoinList> {
    root.placeholders_frozen = true;
    root.join_domains[0].jd_relids = Relids::new();
    root.all_baserels = Relids::new();
    root.outer_join_rels = Relids::new();
    let jointree = JoinTreeNode::From(Box::new(root.parse.jointree.clone()));
    let mut item_list = Vec::new();
    let result = deconstruct_recurse(root, &jointree, 0, &mut item_list);
    root.all_query_rels = root.all_baserels.union(&root.outer_join_rels);
    for i in 0..item_list.len() {
        deconstruct_distribute(root, &mut item_list, i);
    }
    if !root.join_info_list.is_empty() {
        for i in 0..item_list.len() {
            if !item_list[i].oj_joinclauses.is_empty() {
                deconstruct_distribute_oj_quals(root, &item_list, i);
            }
        }
    }
    match result {
        JoinList::List(list) => list,
        rel => vec![rel],
    }
}

/// deconstruct_recurse is the first pass of deconstruct_jointree over a join tree node in a join domain: it records
/// the node's relation sets in a JoinTreeItem appended to the list, and returns its joinlist.
fn deconstruct_recurse(
    root: &mut PlannerInfo<'_, '_>,
    node: &JoinTreeNode,
    parent_domain: usize,
    item_list: &mut Vec<JoinTreeItem>,
) -> JoinList {
    let mut jtitem = JoinTreeItem {
        kind: ItemKind::Rel,
        jdomain: parent_domain,
        qualscope: Relids::new(),
        inner_join_rels: Relids::new(),
        left_rels: Relids::new(),
        right_rels: Relids::new(),
        nonnullable_rels: Relids::new(),
        sjinfo: None,
        oj_joinclauses: Vec::new(),
    };
    let joinlist = match node {
        JoinTreeNode::Rel(varno) => {
            root.all_baserels.add_member(*varno);
            root.join_domains[parent_domain].jd_relids.add_member(*varno);
            jtitem.qualscope = Relids::singleton(*varno);
            JoinList::Rel(*varno)
        }
        JoinTreeNode::From(f) => {
            jtitem.kind = ItemKind::From(f.quals.clone());
            let mut joinlist = Vec::new();
            let mut remaining = f.fromlist.len();
            for n in &f.fromlist {
                let sub_joinlist = deconstruct_recurse(root, n, parent_domain, item_list);
                let sub_item = item_list.last().expect("the child's item");
                jtitem.qualscope.add_members(&sub_item.qualscope);
                jtitem.inner_join_rels = sub_item.inner_join_rels.clone();
                let sub_joinlist = match sub_joinlist {
                    JoinList::List(list) => list,
                    rel => vec![rel],
                };
                remaining -= 1;
                if sub_joinlist.len() <= 1 || joinlist.len() + sub_joinlist.len() + remaining <= FROM_COLLAPSE_LIMIT {
                    joinlist.extend(sub_joinlist);
                } else {
                    joinlist.push(JoinList::List(sub_joinlist));
                }
            }
            if f.fromlist.len() > 1 {
                jtitem.inner_join_rels = jtitem.qualscope.clone();
            }
            JoinList::List(joinlist)
        }
        JoinTreeNode::Join(j) => {
            jtitem.kind = ItemKind::Join(j.jointype, j.quals.clone(), j.rtindex);
            let (leftjoinlist, rightjoinlist);
            match j.jointype {
                JoinType::Inner | JoinType::Semi => {
                    leftjoinlist = deconstruct_recurse(root, &j.larg, parent_domain, item_list);
                    let left_item = item_list.last().expect("the child's item").clone();
                    rightjoinlist = deconstruct_recurse(root, &j.rarg, parent_domain, item_list);
                    let right_item = item_list.last().expect("the child's item").clone();
                    jtitem.qualscope = left_item.qualscope.union(&right_item.qualscope);
                    jtitem.inner_join_rels = match j.jointype {
                        JoinType::Inner => jtitem.qualscope.clone(),
                        _ => left_item.inner_join_rels.union(&right_item.inner_join_rels),
                    };
                    jtitem.left_rels = left_item.qualscope;
                    jtitem.right_rels = right_item.qualscope;
                }
                JoinType::Left | JoinType::Anti => {
                    root.join_domains.push(JoinDomain { jd_relids: Relids::new() });
                    let child_domain = root.join_domains.len() - 1;
                    jtitem.jdomain = child_domain;
                    leftjoinlist = deconstruct_recurse(root, &j.larg, parent_domain, item_list);
                    let left_item = item_list.last().expect("the child's item").clone();
                    rightjoinlist = deconstruct_recurse(root, &j.rarg, child_domain, item_list);
                    let right_item = item_list.last().expect("the child's item").clone();
                    let child_relids = root.join_domains[child_domain].jd_relids.clone();
                    root.join_domains[parent_domain].jd_relids.add_members(&child_relids);
                    jtitem.qualscope = left_item.qualscope.union(&right_item.qualscope);
                    if j.rtindex != 0 {
                        root.join_domains[parent_domain].jd_relids.add_member(j.rtindex);
                        jtitem.qualscope.add_member(j.rtindex);
                        root.outer_join_rels.add_member(j.rtindex);
                        mark_rels_nulled_by_join(root, j.rtindex, &right_item.qualscope);
                    }
                    jtitem.inner_join_rels = left_item.inner_join_rels.union(&right_item.inner_join_rels);
                    jtitem.nonnullable_rels = left_item.qualscope.clone();
                    jtitem.left_rels = left_item.qualscope;
                    jtitem.right_rels = right_item.qualscope;
                }
                JoinType::Full => {
                    root.join_domains.push(JoinDomain { jd_relids: Relids::new() });
                    let fj_domain = root.join_domains.len() - 1;
                    jtitem.jdomain = fj_domain;
                    root.join_domains.push(JoinDomain { jd_relids: Relids::new() });
                    let left_domain = root.join_domains.len() - 1;
                    leftjoinlist = deconstruct_recurse(root, &j.larg, left_domain, item_list);
                    let left_item = item_list.last().expect("the child's item").clone();
                    root.join_domains[fj_domain].jd_relids = root.join_domains[left_domain].jd_relids.clone();
                    root.join_domains.push(JoinDomain { jd_relids: Relids::new() });
                    let right_domain = root.join_domains.len() - 1;
                    rightjoinlist = deconstruct_recurse(root, &j.rarg, right_domain, item_list);
                    let right_item = item_list.last().expect("the child's item").clone();
                    let right_relids = root.join_domains[right_domain].jd_relids.clone();
                    root.join_domains[fj_domain].jd_relids.add_members(&right_relids);
                    let fj_relids = root.join_domains[fj_domain].jd_relids.clone();
                    root.join_domains[parent_domain].jd_relids.add_members(&fj_relids);
                    jtitem.qualscope = left_item.qualscope.union(&right_item.qualscope);
                    root.join_domains[parent_domain].jd_relids.add_member(j.rtindex);
                    jtitem.qualscope.add_member(j.rtindex);
                    root.outer_join_rels.add_member(j.rtindex);
                    mark_rels_nulled_by_join(root, j.rtindex, &left_item.qualscope);
                    mark_rels_nulled_by_join(root, j.rtindex, &right_item.qualscope);
                    jtitem.inner_join_rels = left_item.inner_join_rels.union(&right_item.inner_join_rels);
                    jtitem.left_rels = left_item.qualscope;
                    jtitem.right_rels = right_item.qualscope;
                    jtitem.nonnullable_rels = jtitem.qualscope.clone();
                }
                JoinType::Right => unreachable!("reduce_outer_joins turns right joins into left joins"),
                JoinType::RightSemi | JoinType::RightAnti | JoinType::UniqueOuter | JoinType::UniqueInner => {
                    unreachable!("only join paths are of the join types that swap or unique-ify a semi or anti join")
                }
            }
            let as_list = |list: JoinList| match list {
                JoinList::List(list) => list,
                rel => vec![rel],
            };
            let (leftjoinlist, rightjoinlist) = (as_list(leftjoinlist), as_list(rightjoinlist));
            if j.jointype == JoinType::Full {
                JoinList::List(vec![JoinList::List(vec![JoinList::List(leftjoinlist), JoinList::List(rightjoinlist)])])
            } else if leftjoinlist.len() + rightjoinlist.len() <= JOIN_COLLAPSE_LIMIT {
                JoinList::List(leftjoinlist.into_iter().chain(rightjoinlist).collect())
            } else {
                let part = |list: Vec<JoinList>| match list.len() {
                    1 => list.into_iter().next().expect("one member"),
                    _ => JoinList::List(list),
                };
                JoinList::List(vec![part(leftjoinlist), part(rightjoinlist)])
            }
        }
    };
    item_list.push(jtitem);
    joinlist
}

/// deconstruct_distribute is the second pass of deconstruct_jointree over one JoinTreeItem: it distributes the
/// node's clauses, building the SpecialJoinInfo of an outer, semi, or anti join, and postpones the clauses of a left
/// join that is strict on its left side to the third pass.
fn deconstruct_distribute(root: &mut PlannerInfo<'_, '_>, item_list: &mut [JoinTreeItem], i: usize) {
    let jtitem = item_list[i].clone();
    match &jtitem.kind {
        ItemKind::Rel => {}
        ItemKind::From(quals) => {
            let args = DistributeArgs::new(jtitem.qualscope.clone());
            distribute_quals_to_rels(root, quals, &jtitem, None, &args, None);
        }
        ItemKind::Join(jointype, quals, rtindex) => {
            let (sjinfo, mut ojscope) = match jointype {
                JoinType::Inner => (None, None),
                _ => {
                    let sjinfo = make_outerjoininfo(
                        root,
                        &jtitem.left_rels,
                        &jtitem.right_rels,
                        &jtitem.inner_join_rels,
                        *jointype,
                        *rtindex,
                        quals,
                    );
                    root.sjinfos.push(sjinfo);
                    let id = root.sjinfos.len() - 1;
                    item_list[i].sjinfo = Some(id);
                    let ojscope = match jointype {
                        JoinType::Semi => None,
                        _ => Some(root.sjinfos[id].min_lefthand.union(&root.sjinfos[id].min_righthand)),
                    };
                    (Some(id), ojscope)
                }
            };
            let mut postponed = Vec::new();
            let postpone = *jointype == JoinType::Left && root.sjinfos[sjinfo.expect("a left join")].lhs_strict;
            if postpone {
                let s = &root.sjinfos[sjinfo.expect("a left join")];
                let scope = ojscope.as_mut().expect("a left join's scope");
                scope.add_members(&s.commute_below_l);
                scope.add_members(&s.commute_below_r);
            }
            let args = DistributeArgs {
                ojscope,
                outerjoin_nonnullable: jtitem.nonnullable_rels.clone(),
                ..DistributeArgs::new(jtitem.qualscope.clone())
            };
            distribute_quals_to_rels(root, quals, &jtitem, sjinfo, &args, postpone.then_some(&mut postponed));
            item_list[i].oj_joinclauses = postponed;
            if let Some(sjinfo) = sjinfo {
                root.join_info_list.push(sjinfo);
            }
        }
    }
}

/// mark_rels_nulled_by_join records an outer join among those that can make each of a set of base relations NULL,
/// as Postgres' function of the same name does.
fn mark_rels_nulled_by_join(root: &mut PlannerInfo<'_, '_>, ojrelid: usize, lower_rels: &Relids) {
    for relid in lower_rels.members() {
        let rel = &mut root.rels[relid];
        if rel.reloptkind == RelOptKind::BaseRel {
            rel.nulling_relids.add_member(ojrelid);
        }
    }
}

/// make_outerjoininfo builds the SpecialJoinInfo of an outer, semi, or anti join, with the least sets of relations
/// that its sides must hold before it can be formed and the outer joins below that it commutes with, as Postgres'
/// make_outerjoininfo computes them.
fn make_outerjoininfo(
    root: &mut PlannerInfo<'_, '_>,
    left_rels: &Relids,
    right_rels: &Relids,
    inner_join_rels: &Relids,
    jointype: JoinType,
    ojrelid: usize,
    clause: &[Expr],
) -> SpecialJoinInfo {
    let mut sjinfo = SpecialJoinInfo {
        min_lefthand: Relids::new(),
        min_righthand: Relids::new(),
        syn_lefthand: left_rels.clone(),
        syn_righthand: right_rels.clone(),
        jointype,
        ojrelid,
        commute_above_l: Relids::new(),
        commute_above_r: Relids::new(),
        commute_below_l: Relids::new(),
        commute_below_r: Relids::new(),
        lhs_strict: false,
        semi_can_btree: false,
        semi_can_hash: false,
        semi_rhs_exprs: Vec::new(),
    };
    compute_semijoin_info(root, &mut sjinfo, clause);
    if jointype == JoinType::Full {
        sjinfo.min_lefthand = left_rels.clone();
        sjinfo.min_righthand = right_rels.clone();
        return sjinfo;
    }
    let clause_relids = pull_varnos_list(root, clause);
    let strict_relids = clause.iter().fold(Relids::new(), |r, c| r.union(&find_nonnullable_rels(root.glob, c)));
    sjinfo.lhs_strict = strict_relids.overlap(left_rels);
    let mut min_lefthand = clause_relids.intersect(left_rels);
    let mut min_righthand = clause_relids.union(inner_join_rels).intersect(right_rels);
    let (mut commute_below_l, mut commute_below_r) = (Relids::new(), Relids::new());
    for &other in &root.join_info_list {
        let otherinfo = &root.sjinfos[other];
        if otherinfo.jointype == JoinType::Full {
            if left_rels.overlap(&otherinfo.syn_lefthand) || left_rels.overlap(&otherinfo.syn_righthand) {
                min_lefthand.add_members(&otherinfo.syn_lefthand);
                min_lefthand.add_members(&otherinfo.syn_righthand);
                min_lefthand.add_member(otherinfo.ojrelid);
            }
            if right_rels.overlap(&otherinfo.syn_lefthand) || right_rels.overlap(&otherinfo.syn_righthand) {
                min_righthand.add_members(&otherinfo.syn_lefthand);
                min_righthand.add_members(&otherinfo.syn_righthand);
                min_righthand.add_member(otherinfo.ojrelid);
            }
            continue;
        }
        let have_unsafe_phvs =
            otherinfo.ojrelid != 0 && contain_placeholder_references_to(root, clause, otherinfo.ojrelid);
        let semi_or_anti = matches!(jointype, JoinType::Semi | JoinType::Anti);
        if left_rels.overlap(&otherinfo.syn_righthand) {
            if clause_relids.overlap(&otherinfo.syn_righthand)
                && (have_unsafe_phvs || semi_or_anti || !strict_relids.overlap(&otherinfo.min_righthand))
            {
                min_lefthand.add_members(&otherinfo.syn_lefthand);
                min_lefthand.add_members(&otherinfo.syn_righthand);
                if otherinfo.ojrelid != 0 {
                    min_lefthand.add_member(otherinfo.ojrelid);
                }
            } else if jointype == JoinType::Left
                && otherinfo.jointype == JoinType::Left
                && strict_relids.overlap(&otherinfo.min_righthand)
                && !clause_relids.overlap(&otherinfo.syn_lefthand)
            {
                min_lefthand.del_member(otherinfo.ojrelid);
                commute_below_l.add_member(otherinfo.ojrelid);
            }
        }
        if right_rels.overlap(&otherinfo.syn_righthand) {
            if clause_relids.overlap(&otherinfo.syn_righthand)
                || !clause_relids.overlap(&otherinfo.min_lefthand)
                || have_unsafe_phvs
                || semi_or_anti
                || matches!(otherinfo.jointype, JoinType::Semi | JoinType::Anti)
                || !otherinfo.lhs_strict
            {
                min_righthand.add_members(&otherinfo.syn_lefthand);
                min_righthand.add_members(&otherinfo.syn_righthand);
                if otherinfo.ojrelid != 0 {
                    min_righthand.add_member(otherinfo.ojrelid);
                }
            } else if jointype == JoinType::Left && otherinfo.jointype == JoinType::Left && otherinfo.lhs_strict {
                min_righthand.del_member(otherinfo.ojrelid);
                commute_below_r.add_member(otherinfo.ojrelid);
            }
        }
    }
    for phinfo in &root.placeholder_list {
        if root.glob.placeholder(phinfo.phid).phrels.is_subset(right_rels) {
            min_righthand.add_members(&phinfo.ph_eval_at);
        }
    }
    if min_lefthand.is_empty() {
        min_lefthand = left_rels.clone();
    }
    if min_righthand.is_empty() {
        min_righthand = right_rels.clone();
    }
    commute_below_l.del_members(&min_lefthand);
    commute_below_r.del_members(&min_righthand);
    sjinfo.min_lefthand = min_lefthand;
    sjinfo.min_righthand = min_righthand;
    if !commute_below_l.is_empty() || !commute_below_r.is_empty() {
        for &other in &root.join_info_list.clone() {
            let otherinfo = &mut root.sjinfos[other];
            if commute_below_l.is_member(otherinfo.ojrelid) {
                otherinfo.commute_above_l.add_member(ojrelid);
            } else if commute_below_r.is_member(otherinfo.ojrelid) {
                otherinfo.commute_above_r.add_member(ojrelid);
            }
        }
        sjinfo.commute_below_l = commute_below_l;
        sjinfo.commute_below_r = commute_below_r;
    }
    sjinfo
}

/// compute_semijoin_info records whether a semi join's right side can be made unique by its equality operators'
/// right-hand expressions, by sorting or hashing, as Postgres' function of the same name does.
fn compute_semijoin_info(root: &PlannerInfo<'_, '_>, sjinfo: &mut SpecialJoinInfo, clause: &[Expr]) {
    if sjinfo.jointype != JoinType::Semi {
        return;
    }
    let mut semi_rhs_exprs = Vec::new();
    let (mut all_btree, mut all_hash) = (true, root.enables.hashagg);
    for op in clause {
        let all_varnos = pull_varnos(root, op);
        let local = !all_varnos.overlap(&sjinfo.syn_righthand) || all_varnos.is_subset(&sjinfo.syn_righthand);
        let Some((left_expr, right_expr)) = binary_op_args(op) else {
            if local && !contain_volatile_functions(root.glob, op) {
                continue;
            }
            return;
        };
        if local {
            if contain_volatile_functions(root.glob, op) {
                return;
            }
            continue;
        }
        let (left_varnos, right_varnos) = (pull_varnos(root, left_expr), pull_varnos(root, right_expr));
        let right_expr = if !right_varnos.is_empty()
            && right_varnos.is_subset(&sjinfo.syn_righthand)
            && !left_varnos.overlap(&sjinfo.syn_righthand)
        {
            right_expr
        } else if !left_varnos.is_empty()
            && left_varnos.is_subset(&sjinfo.syn_righthand)
            && !right_varnos.overlap(&sjinfo.syn_righthand)
        {
            left_expr
        } else {
            return;
        };
        let equality = matches!(op, Expr::Compare(CmpOp::Eq, ..));
        let types = (super::nodefuncs::expr_type(root, left_expr), super::nodefuncs::expr_type(root, right_expr));
        all_btree &= equality && !super::nodefuncs::get_mergejoin_opfamilies(types.0, types.1).is_empty();
        all_hash &= equality;
        if !(all_btree || all_hash) {
            return;
        }
        semi_rhs_exprs.push(right_expr.clone());
    }
    if semi_rhs_exprs.is_empty() || semi_rhs_exprs.iter().any(|e| contain_volatile_functions(root.glob, e)) {
        return;
    }
    sjinfo.semi_can_btree = all_btree;
    sjinfo.semi_can_hash = all_hash;
    sjinfo.semi_rhs_exprs = semi_rhs_exprs;
}

/// deconstruct_distribute_oj_quals distributes the postponed clauses of a left join that is strict on its left
/// side, with a clone of them for each order in which it can commute with the outer joins above and below it, each
/// with the nulling relations that the order gives its Vars, as Postgres' function of the same name does.
fn deconstruct_distribute_oj_quals(root: &mut PlannerInfo<'_, '_>, jtitems: &[JoinTreeItem], i: usize) {
    let jtitem = &jtitems[i];
    let sjinfo = root.sjinfos[jtitem.sjinfo.expect("an outer join")].clone();
    let qualscope = sjinfo.syn_lefthand.union(&sjinfo.syn_righthand).with_member(sjinfo.ojrelid);
    let ojscope = sjinfo.min_lefthand.union(&sjinfo.min_righthand);
    let nonnullable_rels = sjinfo.syn_lefthand.clone();
    if sjinfo.commute_above_r.is_empty() && sjinfo.commute_below_l.is_empty() {
        let args = DistributeArgs {
            ojscope: Some(ojscope),
            outerjoin_nonnullable: nonnullable_rels,
            ..DistributeArgs::new(qualscope)
        };
        distribute_quals_to_rels(root, &jtitem.oj_joinclauses, jtitem, jtitem.sjinfo, &args, None);
        return;
    }
    let joins_above = sjinfo.commute_above_r.clone();
    let joins_below = sjinfo.commute_below_l.clone();
    let mut quals = jtitem.oj_joinclauses.clone();
    if !joins_below.is_empty() {
        let mut strip = remove_nulling_relids_fn(root.glob, &joins_below, &Relids::new());
        quals = quals.into_iter().map(&mut strip).collect();
    }
    let mut incompatible_joins = joins_below.union(&joins_above).with_member(sjinfo.ojrelid);
    let save_last_rinfo_serial = root.last_rinfo_serial;
    let mut joins_so_far = Relids::new();
    for otherjtitem in jtitems {
        let Some(other) = otherjtitem.sjinfo else { continue };
        let othersj = root.sjinfos[other].clone();
        let (below_sjinfo, above_sjinfo) = if joins_below.is_member(othersj.ojrelid) {
            (true, false)
        } else if Some(other) == jtitem.sjinfo {
            (false, false)
        } else if joins_above.is_member(othersj.ojrelid) {
            (false, true)
        } else {
            continue;
        };
        root.last_rinfo_serial = save_last_rinfo_serial;
        if above_sjinfo {
            let added = Relids::singleton(othersj.ojrelid);
            quals = quals
                .into_iter()
                .map(|q| add_nulling_relids(root.glob, q, Some(&sjinfo.syn_lefthand), &added))
                .collect();
            incompatible_joins.del_member(othersj.ojrelid);
        }
        let mut this_qualscope = qualscope.union(&joins_so_far);
        let mut this_ojscope = ojscope.union(&joins_so_far);
        if above_sjinfo {
            this_qualscope.add_member(othersj.ojrelid);
            this_ojscope.add_member(othersj.ojrelid);
            this_ojscope.del_member(sjinfo.ojrelid);
        }
        let allow_equivalence = joins_so_far.is_empty();
        let args = DistributeArgs {
            ojscope: Some(this_ojscope),
            outerjoin_nonnullable: nonnullable_rels.clone(),
            incompatible_relids: incompatible_joins.clone(),
            allow_equivalence,
            has_clone: allow_equivalence,
            is_clone: !allow_equivalence,
            ..DistributeArgs::new(this_qualscope)
        };
        distribute_quals_to_rels(root, &quals, otherjtitem, jtitem.sjinfo, &args, None);
        if below_sjinfo {
            let added = Relids::singleton(othersj.ojrelid);
            quals = quals
                .into_iter()
                .map(|q| add_nulling_relids(root.glob, q, Some(&othersj.syn_righthand), &added))
                .collect();
            incompatible_joins.del_member(othersj.ojrelid);
        }
        joins_so_far.add_member(othersj.ojrelid);
    }
}

/// DistributeArgs is what distribute_qual_to_rels takes besides the clause, its join tree item, and its outer join:
/// the relations of its syntactic level, the relations of its outer join, that join's non-nullable side, the joins
/// where a clone must not be evaluated, whether it may feed an equivalence class, and whether it has or is a clone.
#[derive(Clone)]
struct DistributeArgs {
    qualscope: Relids,
    ojscope: Option<Relids>,
    outerjoin_nonnullable: Relids,
    incompatible_relids: Relids,
    allow_equivalence: bool,
    has_clone: bool,
    is_clone: bool,
}

impl DistributeArgs {
    /// new returns the arguments of a WHERE or inner join clause of a syntactic level.
    fn new(qualscope: Relids) -> DistributeArgs {
        DistributeArgs {
            qualscope,
            ojscope: None,
            outerjoin_nonnullable: Relids::new(),
            incompatible_relids: Relids::new(),
            allow_equivalence: true,
            has_clone: false,
            is_clone: false,
        }
    }
}

/// distribute_quals_to_rels is distribute_qual_to_rels over a list of clauses.
fn distribute_quals_to_rels(
    root: &mut PlannerInfo<'_, '_>,
    clauses: &[Expr],
    jtitem: &JoinTreeItem,
    sjinfo: Option<SjId>,
    args: &DistributeArgs,
    mut postponed_oj_qual_list: Option<&mut Vec<Expr>>,
) {
    for clause in clauses {
        distribute_qual_to_rels(root, clause, jtitem, sjinfo, args, postponed_oj_qual_list.as_deref_mut());
    }
}

/// distribute_qual_to_rels makes a RestrictInfo of a clause and adds it to an equivalence class or attaches it to the
/// relations it restricts, or to the join that must evaluate it, or postpones a left join's clause to the third pass
/// of deconstruct_jointree, as Postgres' function of the same name does.
fn distribute_qual_to_rels(
    root: &mut PlannerInfo<'_, '_>,
    clause: &Expr,
    jtitem: &JoinTreeItem,
    sjinfo: Option<SjId>,
    args: &DistributeArgs,
    postponed_oj_qual_list: Option<&mut Vec<Expr>>,
) {
    let mut relids = pull_varnos(root, clause);
    assert!(relids.is_subset(&args.qualscope), "lateral references are not ported yet");
    let mut pseudoconstant = false;
    if relids.is_empty() {
        if let Some(ojscope) = &args.ojscope {
            relids = ojscope.clone();
        } else if contain_volatile_functions(root.glob, clause) {
            relids = args.qualscope.clone();
        } else {
            relids = match jtitem.jdomain {
                0 => root.join_domains[0].jd_relids.clone(),
                _ => args.qualscope.clone(),
            };
            pseudoconstant = true;
            root.has_pseudo_constant_quals = true;
        }
    }
    let (is_pushed_down, maybe_equivalence, maybe_outer_join);
    if relids.overlap(&args.outerjoin_nonnullable) {
        if let Some(postponed) = postponed_oj_qual_list {
            postponed.push(clause.clone());
            return;
        }
        (is_pushed_down, maybe_equivalence, maybe_outer_join) = (false, false, true);
        relids = args.ojscope.clone().expect("an outer join's scope");
    } else {
        (is_pushed_down, maybe_equivalence, maybe_outer_join) = (true, args.allow_equivalence, false);
    }
    let mut restrictinfo = make_restrictinfo(
        root,
        clause.clone(),
        RestrictInfoArgs {
            is_pushed_down,
            has_clone: args.has_clone,
            is_clone: args.is_clone,
            pseudoconstant,
            security_level: 0,
            required_relids: Some(relids.clone()),
            incompatible_relids: args.incompatible_relids.clone(),
            outer_relids: args.outerjoin_nonnullable.clone(),
        },
    );
    if relids.num_members() > 1 {
        let vars = pull_var_clause(root.glob, clause, true);
        let where_needed = match args.is_clone {
            true => relids.intersect(&root.all_baserels),
            false => relids.clone(),
        };
        add_vars_to_targetlist(root, &vars, &where_needed);
    }
    check_mergejoinable(root, restrictinfo);
    if !root.rinfos[restrictinfo].mergeopfamilies.is_empty() {
        if maybe_equivalence {
            if process_equivalence(root, &mut restrictinfo, jtitem.jdomain) {
                return;
            }
            if !root.rinfos[restrictinfo].mergeopfamilies.is_empty() {
                initialize_mergeclause_eclasses(root, restrictinfo);
            }
        } else if maybe_outer_join && root.rinfos[restrictinfo].can_join {
            initialize_mergeclause_eclasses(root, restrictinfo);
            let sjinfo = sjinfo.expect("an outer join's clause has its join");
            let r = &root.rinfos[restrictinfo];
            let ojcinfo = OuterJoinClauseInfo { rinfo: restrictinfo, sjinfo };
            if r.left_relids.is_subset(&args.outerjoin_nonnullable)
                && !r.right_relids.overlap(&args.outerjoin_nonnullable)
            {
                root.left_join_clauses.push(ojcinfo);
                return;
            }
            if r.right_relids.is_subset(&args.outerjoin_nonnullable)
                && !r.left_relids.overlap(&args.outerjoin_nonnullable)
            {
                root.right_join_clauses.push(ojcinfo);
                return;
            }
            if root.sjinfos[sjinfo].jointype == JoinType::Full {
                root.full_join_clauses.push(ojcinfo);
                return;
            }
        } else {
            initialize_mergeclause_eclasses(root, restrictinfo);
        }
    }
    distribute_restrictinfo_to_rels(root, restrictinfo);
}

/// add_base_clause_to_rel adds a restriction to a base relation, dropping it when the relation's NOT NULL columns
/// make it always true and making it the constant false when they make it always false, as Postgres' function of the
/// same name does.
fn add_base_clause_to_rel(root: &mut PlannerInfo<'_, '_>, relid: usize, mut restrictinfo: RinfoId) {
    if restriction_is_always_true(root, &root.rinfos[restrictinfo]) {
        return;
    }
    if restriction_is_always_false(root, &root.rinfos[restrictinfo]) {
        let r = root.rinfos[restrictinfo].clone();
        let (save_rinfo_serial, save_last_rinfo_serial) = (r.rinfo_serial, root.last_rinfo_serial);
        restrictinfo = make_restrictinfo(
            root,
            Expr::Const(crate::types::Value::Bool(false)),
            RestrictInfoArgs {
                is_pushed_down: r.is_pushed_down,
                has_clone: r.has_clone,
                is_clone: r.is_clone,
                pseudoconstant: r.pseudoconstant,
                security_level: 0,
                required_relids: Some(r.required_relids.clone()),
                incompatible_relids: r.incompatible_relids.clone(),
                outer_relids: r.outer_relids.clone(),
            },
        );
        root.rinfos[restrictinfo].rinfo_serial = save_rinfo_serial;
        root.last_rinfo_serial = save_last_rinfo_serial;
    }
    let security_level = root.rinfos[restrictinfo].security_level;
    let rel = &mut root.rels[relid];
    rel.baserestrictinfo.push(restrictinfo);
    rel.baserestrict_min_security = rel.baserestrict_min_security.min(security_level);
}

/// expr_is_nonnullable reports whether an expression is a Var of a NOT NULL column that no outer join makes NULL, as
/// Postgres' function of the same name does.
fn expr_is_nonnullable(root: &PlannerInfo<'_, '_>, e: &Expr) -> bool {
    let Expr::Column(id) = e else { return false };
    let VarNode::Var(var) = root.glob.node(*id) else { return false };
    var.varnullingrels.is_empty() && root.rels[var.varno].notnullattnums.contains(&var.varattno)
}

/// restriction_is_always_true reports whether a restriction is an IS NOT NULL test of a NOT NULL column, or an OR of
/// one, as Postgres' function of the same name does.
pub fn restriction_is_always_true(root: &PlannerInfo<'_, '_>, rinfo: &RestrictInfo) -> bool {
    if rinfo.has_clone || rinfo.is_clone {
        return false;
    }
    if let Expr::IsNull(arg, negated) = &rinfo.clause {
        return *negated && !matches!(**arg, Expr::Row(..)) && expr_is_nonnullable(root, arg);
    }
    if let Some(orclause) = &rinfo.orclause {
        return orclause.iter().any(|arm| match arm.as_slice() {
            [orarg] => restriction_is_always_true(root, &root.rinfos[*orarg]),
            _ => false,
        });
    }
    false
}

/// restriction_is_always_false reports whether a restriction is an IS NULL test of a NOT NULL column, or an OR of
/// such tests, as Postgres' function of the same name does.
pub fn restriction_is_always_false(root: &PlannerInfo<'_, '_>, rinfo: &RestrictInfo) -> bool {
    if rinfo.has_clone || rinfo.is_clone {
        return false;
    }
    if let Expr::IsNull(arg, negated) = &rinfo.clause {
        return !*negated && !matches!(**arg, Expr::Row(..)) && expr_is_nonnullable(root, arg);
    }
    if restriction_is_or_clause(rinfo) {
        return rinfo.orclause.iter().flatten().all(|arm| match arm.as_slice() {
            [orarg] => restriction_is_always_false(root, &root.rinfos[*orarg]),
            _ => false,
        });
    }
    false
}

/// distribute_restrictinfo_to_rels attaches a clause to the base relation it restricts, or as a join clause to each
/// relation it reads, as Postgres' function of the same name does.
pub fn distribute_restrictinfo_to_rels(root: &mut PlannerInfo<'_, '_>, restrictinfo: RinfoId) {
    let relids = root.rinfos[restrictinfo].required_relids.clone();
    assert!(!relids.is_empty(), "cannot cope with variable-free clause");
    match relids.singleton_member() {
        Some(relid) => add_base_clause_to_rel(root, relid, restrictinfo),
        None => {
            check_hashjoinable(root, restrictinfo);
            super::joininfo::add_join_clause_to_rels(root, restrictinfo, &relids);
        }
    }
}

/// process_implied_equality distributes an equality that an equivalence class implies, returning its RestrictInfo,
/// or None when it is between constants that are equal, as Postgres' function of the same name does.
pub fn process_implied_equality(
    root: &mut PlannerInfo<'_, '_>,
    item1: Expr,
    item2: Expr,
    qualscope: &Relids,
    security_level: usize,
    both_const: bool,
) -> Option<RinfoId> {
    if both_const
        && let (Expr::Const(a), Expr::Const(b)) = (&item1, &item2)
        && *a != crate::types::Value::Null
        && a == b
    {
        return None;
    }
    let clause = Expr::Compare(CmpOp::Eq, Box::new(item1), Box::new(item2));
    let mut relids = pull_varnos(root, &clause);
    let mut pseudoconstant = false;
    if relids.is_empty() {
        relids = get_join_domain_min_rels(root, qualscope);
        pseudoconstant = true;
        root.has_pseudo_constant_quals = true;
    }
    let restrictinfo = make_restrictinfo(
        root,
        clause.clone(),
        RestrictInfoArgs {
            is_pushed_down: true,
            pseudoconstant,
            security_level,
            required_relids: Some(relids.clone()),
            ..RestrictInfoArgs::default()
        },
    );
    if relids.num_members() > 1 {
        let vars = pull_var_clause(root.glob, &clause, true);
        add_vars_to_targetlist(root, &vars, &relids);
    }
    check_mergejoinable(root, restrictinfo);
    distribute_restrictinfo_to_rels(root, restrictinfo);
    Some(restrictinfo)
}

/// build_implied_join_equality returns the RestrictInfo of an equality that an equivalence class implies between two
/// relations, without distributing it, as Postgres' function of the same name does.
pub fn build_implied_join_equality(
    root: &mut PlannerInfo<'_, '_>,
    clause: Expr,
    qualscope: Relids,
    security_level: usize,
) -> RinfoId {
    let restrictinfo = make_restrictinfo(
        root,
        clause,
        RestrictInfoArgs {
            is_pushed_down: true,
            security_level,
            required_relids: Some(qualscope),
            ..RestrictInfoArgs::default()
        },
    );
    check_mergejoinable(root, restrictinfo);
    check_hashjoinable(root, restrictinfo);
    restrictinfo
}

/// get_join_domain_min_rels returns the relations of a join domain without those on the nullable sides of its left
/// joins, where a pseudoconstant clause of the domain is evaluated, as Postgres' function of the same name does.
fn get_join_domain_min_rels(root: &PlannerInfo<'_, '_>, domain_relids: &Relids) -> Relids {
    let mut result = domain_relids.clone();
    if result == root.all_query_rels {
        return result;
    }
    for &sj in &root.join_info_list {
        let sjinfo = &root.sjinfos[sj];
        if sjinfo.jointype == JoinType::Left && result.is_member(sjinfo.ojrelid) {
            result.del_member(sjinfo.ojrelid);
            result.del_members(&sjinfo.syn_righthand);
        }
    }
    result
}

/// check_mergejoinable records the btree operator families of an equality whose sides' types have one, which merge
/// joins and equivalence classes use, as Postgres' function of the same name does.
fn check_mergejoinable(root: &mut PlannerInfo<'_, '_>, rinfo: RinfoId) {
    let r = &root.rinfos[rinfo];
    if r.pseudoconstant {
        return;
    }
    let Expr::Compare(CmpOp::Eq, left, right) = &r.clause else { return };
    let types = (super::nodefuncs::expr_type(root, left), super::nodefuncs::expr_type(root, right));
    if !contain_volatile_functions(root.glob, &r.clause) {
        root.rinfos[rinfo].mergeopfamilies = super::nodefuncs::get_mergejoin_opfamilies(types.0, types.1);
    }
}

/// check_hashjoinable records whether a clause is an equality that a hash join can use, as Postgres' function of the
/// same name does: Doltgres' hash joins hash any values whose equality they test.
fn check_hashjoinable(root: &mut PlannerInfo<'_, '_>, rinfo: RinfoId) {
    let r = &root.rinfos[rinfo];
    if r.pseudoconstant || !matches!(r.clause, Expr::Compare(CmpOp::Eq, ..)) {
        return;
    }
    root.rinfos[rinfo].hashjoinable =
        !contain_volatile_functions(root.glob, &r.clause) && !super::clauses::contain_subplans(&r.clause);
}

/// match_foreign_keys_to_quals keeps the query's foreign keys whose every column pair an equivalence class or a join
/// clause equates, recording which, as Postgres' function of the same name does.
pub fn match_foreign_keys_to_quals(root: &mut PlannerInfo<'_, '_>) {
    let mut newlist = Vec::new();
    for mut fkinfo in std::mem::take(&mut root.fkey_list) {
        let rel_kind = |relid: usize| root.rels.get(relid).map(|r| (r.relid == relid, r.reloptkind));
        if rel_kind(fkinfo.con_relid) != Some((true, RelOptKind::BaseRel))
            || rel_kind(fkinfo.ref_relid) != Some((true, RelOptKind::BaseRel))
        {
            continue;
        }
        for colno in 0..fkinfo.conkey.len() {
            if let Some(ec) = super::equivclass::match_eclasses_to_foreign_key_col(root, &mut fkinfo, colno) {
                fkinfo.nmatched_ec += 1;
                if root.eq_classes[ec].ec_has_const {
                    fkinfo.nconst_ec += 1;
                }
                continue;
            }
            let (con_attno, ref_attno) = (fkinfo.conkey[colno], fkinfo.confkey[colno]);
            let is_var = |e: &Expr, varno: usize, attno: usize| match e {
                Expr::Column(id) => {
                    matches!(root.glob.node(*id), VarNode::Var(v) if v.varno == varno && v.varattno == attno)
                }
                _ => false,
            };
            for &rinfo in &root.rels[fkinfo.con_relid].joininfo {
                let Expr::Compare(CmpOp::Eq, left, right) = &root.rinfos[rinfo].clause else { continue };
                if (is_var(left, fkinfo.ref_relid, ref_attno) && is_var(right, fkinfo.con_relid, con_attno))
                    || (is_var(right, fkinfo.ref_relid, ref_attno) && is_var(left, fkinfo.con_relid, con_attno))
                {
                    fkinfo.rinfos[colno].push(rinfo);
                    fkinfo.nmatched_ri += 1;
                }
            }
            if !fkinfo.rinfos[colno].is_empty() {
                fkinfo.nmatched_rcols += 1;
            }
        }
        if fkinfo.nmatched_ec + fkinfo.nmatched_rcols == fkinfo.conkey.len() {
            newlist.push(fkinfo);
        }
    }
    root.fkey_list = newlist;
}
