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

//! Postgres' optimizer/plan/analyzejoins.c: removing left joins that cannot change the query's result, turning semi
//! joins whose inner side is unique into inner joins, removing joins of a table to itself by a unique key, and
//! proving that a join's inner relation matches each outer row at most once. Each removal changes the query, after
//! which query_planner starts over.

use super::PlannerInfo;
use super::clauses::contain_volatile_functions;
use super::equivclass::generate_join_implied_equalities;
use super::indxpath::relation_has_unique_index_ext;
use super::initsplan::JoinList;
use super::joinpath::clause_sides_match_join;
use super::nodes::{FromExpr, JoinTreeNode, JoinType, RelOptKind, Relids, RinfoId, RteKind, SpecialJoinInfo};
use super::prepjointree::get_relids_in_jointree;
use super::restrictinfo::{binary_op_args, rinfo_is_pushed_down};
use super::var::{change_var_nodes, change_var_nodes_fn, mutate_query, pull_varnos};
use crate::expr::Expr;

/// remove_useless_outer_joins removes each left join to one relation whose columns nothing above reads and whose
/// clauses find at most one inner row for each outer row, reporting whether it removed any, as Postgres' function of
/// the same name does.
pub fn remove_useless_outer_joins(root: &mut PlannerInfo<'_, '_>) -> bool {
    let mut removed_relids = Relids::new();
    for sj in root.join_info_list.clone() {
        let sjinfo = root.sjinfos[sj].clone();
        if !join_is_removable(root, &sjinfo) {
            continue;
        }
        let innerrelid = sjinfo.syn_righthand.singleton_member().expect("one inner relation");
        let mut jointree = JoinTreeNode::From(Box::new(root.parse.jointree.clone()));
        let nremoved = remove_join_from_jointree(&mut jointree, sjinfo.ojrelid);
        assert_eq!(nremoved, 1, "failed to find join {} in jointree", sjinfo.ojrelid);
        let JoinTreeNode::From(jointree) = jointree else { unreachable!("the join tree stays a FROM list") };
        root.parse.jointree = *jointree;
        removed_relids.add_member(innerrelid);
        removed_relids.add_member(sjinfo.ojrelid);
    }
    if removed_relids.is_empty() {
        return false;
    }
    remove_rels_from_query_tree(root, &removed_relids);
    true
}

/// join_is_removable reports whether a left join to one relation can be removed: nothing above it reads the relation,
/// and its clauses find at most one of the relation's rows for each outer row, as Postgres' function of the same name
/// decides.
fn join_is_removable(root: &mut PlannerInfo<'_, '_>, sjinfo: &SpecialJoinInfo) -> bool {
    if sjinfo.jointype != JoinType::Left {
        return false;
    }
    let Some(innerrelid) = sjinfo.syn_righthand.singleton_member() else { return false };
    if !rel_supports_distinctness(root, innerrelid) {
        return false;
    }
    let inputrelids = sjinfo.min_lefthand.union(&sjinfo.min_righthand);
    let joinrelids = inputrelids.clone().with_member(sjinfo.ojrelid);
    let innerrel = &root.rels[innerrelid];
    if innerrel.attr_needed.iter().any(|needed| !needed.is_subset(&inputrelids)) {
        return false;
    }
    for phinfo in &root.placeholder_list {
        if phinfo.ph_lateral.overlap(&innerrel.relids) {
            return false;
        }
        if !phinfo.ph_eval_at.overlap(&innerrel.relids) || phinfo.ph_needed.is_subset(&inputrelids) {
            continue;
        }
        if !phinfo.ph_eval_at.is_member(sjinfo.ojrelid)
            || !sjinfo.min_lefthand.overlap(&phinfo.ph_eval_at)
            || pull_varnos(root, &root.glob.placeholder(phinfo.phid).phexpr).overlap(&innerrel.relids)
        {
            return false;
        }
    }
    let mut clause_list = Vec::new();
    for &rinfo in &innerrel.joininfo {
        let r = &root.rinfos[rinfo];
        if r.is_clone || rinfo_is_pushed_down(r, &joinrelids) || !r.can_join || r.mergeopfamilies.is_empty() {
            continue;
        }
        if clause_sides_match_join(r, &sjinfo.min_lefthand, &innerrel.relids) {
            clause_list.push(rinfo);
        }
    }
    rel_is_distinct_for(root, innerrelid, &clause_list, None)
}

/// remove_join_from_jointree replaces the join of a range table index by its left side, returning how many joins it
/// replaced, as Postgres' function of the same name does.
fn remove_join_from_jointree(node: &mut JoinTreeNode, ojrelid: usize) -> usize {
    match node {
        JoinTreeNode::Rel(_) => 0,
        JoinTreeNode::From(f) => f.fromlist.iter_mut().map(|n| remove_join_from_jointree(n, ojrelid)).sum(),
        JoinTreeNode::Join(j) if j.rtindex == ojrelid => {
            *node = std::mem::replace(&mut j.larg, JoinTreeNode::Rel(0));
            1
        }
        JoinTreeNode::Join(j) => {
            remove_join_from_jointree(&mut j.larg, ojrelid) + remove_join_from_jointree(&mut j.rarg, ojrelid)
        }
    }
}

/// remove_rels_from_query_tree removes removed relations from the nulling relations of the query's Vars and the
/// relations of its PlaceHolderVars, as Postgres' function of the same name does.
fn remove_rels_from_query_tree(root: &mut PlannerInfo<'_, '_>, removed_relids: &Relids) {
    for relid in removed_relids.members() {
        let mut f = change_var_nodes_fn(root.glob, relid, 0);
        mutate_query(&mut root.parse, &mut f);
    }
}

/// reduce_unique_semijoins turns each semi join whose one inner relation has at most one row for each outer row
/// into an inner join, reporting whether it changed any, as Postgres' function of the same name does.
pub fn reduce_unique_semijoins(root: &mut PlannerInfo<'_, '_>) -> bool {
    let mut changed = false;
    for sj in root.join_info_list.clone() {
        let sjinfo = root.sjinfos[sj].clone();
        if sjinfo.jointype != JoinType::Semi {
            continue;
        }
        let Some(innerrelid) = sjinfo.syn_righthand.singleton_member() else { continue };
        if !rel_supports_distinctness(root, innerrelid) {
            continue;
        }
        let joinrelids = sjinfo.min_lefthand.union(&sjinfo.min_righthand);
        let mut restrictlist = generate_join_implied_equalities(root, &joinrelids, &sjinfo.min_lefthand, innerrelid, 0);
        restrictlist.extend(root.rels[innerrelid].joininfo.clone());
        if !innerrel_is_unique(root, &joinrelids, &sjinfo.min_lefthand, innerrelid, JoinType::Semi, &restrictlist, true)
        {
            continue;
        }
        let mut jointree = JoinTreeNode::From(Box::new(root.parse.jointree.clone()));
        assert!(
            reduce_semijoin_in_jointree(&mut jointree, &sjinfo.syn_righthand),
            "failed to find semijoin in jointree"
        );
        let JoinTreeNode::From(jointree) = jointree else { unreachable!("the join tree stays a FROM list") };
        root.parse.jointree = *jointree;
        changed = true;
    }
    changed
}

/// reduce_semijoin_in_jointree turns the semi join of an inner side into an inner join, reporting whether it found
/// it, as Postgres' function of the same name does.
fn reduce_semijoin_in_jointree(node: &mut JoinTreeNode, syn_righthand: &Relids) -> bool {
    match node {
        JoinTreeNode::Rel(_) => false,
        JoinTreeNode::From(f) => f.fromlist.iter_mut().any(|n| reduce_semijoin_in_jointree(n, syn_righthand)),
        JoinTreeNode::Join(j) => {
            if j.jointype == JoinType::Semi && get_relids_in_jointree(&j.rarg, true, false) == *syn_righthand {
                j.jointype = JoinType::Inner;
                return true;
            }
            reduce_semijoin_in_jointree(&mut j.larg, syn_righthand)
                || reduce_semijoin_in_jointree(&mut j.rarg, syn_righthand)
        }
    }
}

/// rel_supports_distinctness reports whether a base relation could be proven to have at most one row for some values
/// of its columns: a table with a unique index that is immediate and not partial, as Postgres' function of the same
/// name checks. A subquery that the planner did not pull up is already planned, so its distinctness is unknown.
fn rel_supports_distinctness(root: &PlannerInfo<'_, '_>, relid: usize) -> bool {
    let rel = &root.rels[relid];
    rel.reloptkind == RelOptKind::BaseRel && rel.indexlist.iter().any(|ind| ind.unique && !ind.has_predicate)
}

/// rel_is_distinct_for reports whether a base relation has at most one row for each set of values of the inner sides
/// of a list of mergejoinable join clauses, as Postgres' function of the same name proves it.
fn rel_is_distinct_for(
    root: &mut PlannerInfo<'_, '_>,
    relid: usize,
    clause_list: &[RinfoId],
    extra_clauses: Option<&mut Vec<RinfoId>>,
) -> bool {
    root.rels[relid].reloptkind == RelOptKind::BaseRel
        && matches!(root.parse.rte(relid).kind, RteKind::Relation(..))
        && relation_has_unique_index_ext(root, relid, clause_list, extra_clauses)
}

/// innerrel_is_unique reports whether each outer row matches at most one row of the inner relation by a join's
/// clauses, caching the answer for the outer relations, as Postgres' function of the same name does.
pub fn innerrel_is_unique(
    root: &mut PlannerInfo<'_, '_>,
    joinrelids: &Relids,
    outerrelids: &Relids,
    innerrel: usize,
    jointype: JoinType,
    restrictlist: &[RinfoId],
    force_cache: bool,
) -> bool {
    innerrel_is_unique_ext(root, joinrelids, outerrelids, innerrel, jointype, restrictlist, force_cache, None)
}

/// innerrel_is_unique_ext is innerrel_is_unique that also returns the inner relation's restrictions that the proof
/// for a self join used, as Postgres' function of the same name does.
#[allow(clippy::too_many_arguments)]
fn innerrel_is_unique_ext(
    root: &mut PlannerInfo<'_, '_>,
    joinrelids: &Relids,
    outerrelids: &Relids,
    innerrel: usize,
    jointype: JoinType,
    restrictlist: &[RinfoId],
    force_cache: bool,
    extra_clauses: Option<&mut Vec<RinfoId>>,
) -> bool {
    if restrictlist.is_empty() || !rel_supports_distinctness(root, innerrel) {
        return false;
    }
    let self_join = extra_clauses.is_some();
    for unique in &root.rels[innerrel].unique_for_rels {
        if (!self_join && unique.outerrelids.is_subset(outerrelids))
            || (self_join && unique.outerrelids == *outerrelids && unique.self_join)
        {
            if let Some(extra) = extra_clauses {
                *extra = unique.extra_clauses.clone();
            }
            return true;
        }
    }
    if root.rels[innerrel].non_unique_for_rels.iter().any(|r| outerrelids.is_subset(r)) {
        return false;
    }
    let mut outer_exprs = Vec::new();
    let unique = is_innerrel_unique_for(
        root,
        joinrelids,
        outerrelids,
        innerrel,
        jointype,
        restrictlist,
        self_join.then_some(&mut outer_exprs),
    );
    if unique {
        root.rels[innerrel].unique_for_rels.push(super::nodes::UniqueRelInfo {
            outerrelids: outerrelids.clone(),
            self_join,
            extra_clauses: outer_exprs.clone(),
        });
        if let Some(extra) = extra_clauses {
            *extra = outer_exprs;
        }
    } else if force_cache {
        root.rels[innerrel].non_unique_for_rels.push(outerrelids.clone());
    }
    unique
}

/// is_innerrel_unique_for reports whether the mergejoinable join clauses between the outer relations and the inner
/// one prove the inner one unique for each outer row, as Postgres' function of the same name does.
fn is_innerrel_unique_for(
    root: &mut PlannerInfo<'_, '_>,
    joinrelids: &Relids,
    outerrelids: &Relids,
    innerrel: usize,
    jointype: JoinType,
    restrictlist: &[RinfoId],
    extra_clauses: Option<&mut Vec<RinfoId>>,
) -> bool {
    let inner_relids = root.rels[innerrel].relids.clone();
    let clause_list: Vec<RinfoId> = restrictlist
        .iter()
        .copied()
        .filter(|&r| {
            let r = &root.rinfos[r];
            !(jointype.is_outer() && rinfo_is_pushed_down(r, joinrelids))
                && r.can_join
                && !r.mergeopfamilies.is_empty()
                && clause_sides_match_join(r, outerrelids, &inner_relids)
        })
        .collect();
    let relid = root.rels[innerrel].relid;
    rel_is_distinct_for(root, relid, &clause_list, extra_clauses)
}

/// remove_self_join_rel removes one of two relations of the same table that a unique key joins, making the query
/// read the other wherever it read the removed one, as Postgres' function of the same name does.
fn remove_self_join_rel(root: &mut PlannerInfo<'_, '_>, to_keep: usize, to_remove: usize) {
    let mut jointree = JoinTreeNode::From(Box::new(root.parse.jointree.clone()));
    let (mut orphan_quals, mut nremoved) = (Vec::new(), 0);
    let removed = remove_rel_from_jointree(&mut jointree, to_remove, &mut orphan_quals, &mut nremoved);
    assert!(!removed && nremoved == 1, "failed to find relation {to_remove} in jointree");
    let JoinTreeNode::From(jointree) = jointree else { unreachable!("the join tree stays a FROM list") };
    root.parse.jointree = *jointree;
    {
        let mut f = change_var_nodes_fn(root.glob, to_remove, to_keep);
        mutate_query(&mut root.parse, &mut f);
    }
    let mut jointree = JoinTreeNode::From(Box::new(root.parse.jointree.clone()));
    let (mut hoist_quals, mut found_relid) = (Vec::new(), false);
    fixup_selfjoin_jointree(root, &mut jointree, to_keep, &mut hoist_quals, &mut found_relid);
    let JoinTreeNode::From(jointree) = jointree else { unreachable!("the join tree stays a FROM list") };
    root.parse.jointree = *jointree;
}

/// remove_rel_from_jointree removes a relation's reference from a join tree, moving the quals of a FROM list or
/// inner join that it leaves empty up, and reports whether the node itself is gone, as Postgres' function of the same
/// name does.
fn remove_rel_from_jointree(
    node: &mut JoinTreeNode,
    relid: usize,
    orphan_quals: &mut Vec<Expr>,
    nremoved: &mut usize,
) -> bool {
    match node {
        JoinTreeNode::Rel(varno) => {
            if *varno == relid {
                *nremoved += 1;
                return true;
            }
            false
        }
        JoinTreeNode::From(f) => {
            let mut sub_orphans = Vec::new();
            f.fromlist.retain_mut(|n| !remove_rel_from_jointree(n, relid, &mut sub_orphans, nremoved));
            sub_orphans.append(&mut f.quals);
            f.quals = sub_orphans;
            if f.fromlist.is_empty() {
                orphan_quals.splice(0..0, std::mem::take(&mut f.quals));
                return true;
            }
            false
        }
        JoinTreeNode::Join(j) => {
            let mut sub_orphans = Vec::new();
            let larg_gone = remove_rel_from_jointree(&mut j.larg, relid, &mut sub_orphans, nremoved);
            let rarg_gone = remove_rel_from_jointree(&mut j.rarg, relid, &mut sub_orphans, nremoved);
            if larg_gone || rarg_gone {
                let surviving =
                    std::mem::replace(if larg_gone { &mut j.rarg } else { &mut j.larg }, JoinTreeNode::Rel(0));
                sub_orphans.append(&mut j.quals);
                *node = JoinTreeNode::From(Box::new(FromExpr { fromlist: vec![surviving], quals: sub_orphans }));
            }
            false
        }
    }
}

/// fixup_selfjoin_jointree moves the quals that read the kept relation down to the lowest node above it, and turns
/// each of its equalities of an expression with itself into an IS NOT NULL test, as Postgres' function of the same
/// name does.
fn fixup_selfjoin_jointree(
    root: &PlannerInfo<'_, '_>,
    node: &mut JoinTreeNode,
    relid: usize,
    hoist_quals: &mut Vec<Expr>,
    found_relid: &mut bool,
) {
    let (children, quals): (Vec<&mut JoinTreeNode>, &mut Vec<Expr>) = match node {
        JoinTreeNode::Rel(varno) => {
            if *varno == relid {
                *found_relid = true;
            }
            return;
        }
        JoinTreeNode::From(f) => (f.fromlist.iter_mut().collect(), &mut f.quals),
        JoinTreeNode::Join(j) => (vec![&mut j.larg, &mut j.rarg], &mut j.quals),
    };
    let (mut sub_hoist_quals, mut sub_found_relid) = (Vec::new(), false);
    for child in children {
        fixup_selfjoin_jointree(root, child, relid, &mut sub_hoist_quals, &mut sub_found_relid);
    }
    if sub_found_relid {
        sub_hoist_quals.append(quals);
        *quals = sub_hoist_quals;
        *found_relid = true;
    } else {
        let (hoistable, keepable): (Vec<Expr>, Vec<Expr>) =
            std::mem::take(quals).into_iter().partition(|q| pull_varnos(root, q).is_member(relid));
        *quals = keepable;
        sub_hoist_quals.extend(hoistable);
        hoist_quals.splice(0..0, sub_hoist_quals);
    }
    let mut result: Vec<Expr> = Vec::new();
    for qual in std::mem::take(quals) {
        if pull_varnos(root, &qual).is_member(relid) {
            let qual = replace_selfjoin_qual(root, qual);
            if result.contains(&qual) {
                continue;
            }
            result.push(qual);
        } else {
            result.push(qual);
        }
    }
    *quals = result;
}

/// replace_selfjoin_qual turns a strict mergejoinable equality of an expression with itself into an IS NOT NULL test,
/// as Postgres' function of the same name does.
fn replace_selfjoin_qual(root: &PlannerInfo<'_, '_>, qual: Expr) -> Expr {
    let Expr::Compare(crate::expr::CmpOp::Eq, left, right) = &qual else { return qual };
    let types = (super::nodefuncs::expr_type(root, left), super::nodefuncs::expr_type(root, right));
    if left != right
        || super::nodefuncs::get_mergejoin_opfamilies(types.0, types.1).is_empty()
        || contain_volatile_functions(root.glob, left)
    {
        return qual;
    }
    Expr::IsNull(left.clone(), true)
}

/// split_selfjoin_quals splits a self join's clauses into the mergejoinable equalities of an expression of one
/// relation with the same expression of the other, and the rest, as Postgres' function of the same name does.
fn split_selfjoin_quals(root: &mut PlannerInfo<'_, '_>, joinquals: &[RinfoId]) -> (Vec<RinfoId>, Vec<RinfoId>) {
    let (mut sjoinquals, mut ojoinquals) = (Vec::new(), Vec::new());
    for &rinfo in joinquals {
        let r = root.rinfos[rinfo].clone();
        let (Some(left_relid), Some(right_relid)) =
            (r.left_relids.singleton_member(), r.right_relids.singleton_member())
        else {
            ojoinquals.push(rinfo);
            continue;
        };
        if r.mergeopfamilies.is_empty() || r.clause_relids.num_members() != 2 {
            ojoinquals.push(rinfo);
            continue;
        }
        let Some((leftexpr, rightexpr)) = binary_op_args(&r.clause) else {
            ojoinquals.push(rinfo);
            continue;
        };
        let rightexpr = change_var_nodes(root.glob, rightexpr.clone(), right_relid, left_relid);
        if *leftexpr == rightexpr {
            sjoinquals.push(rinfo);
        } else {
            ojoinquals.push(rinfo);
        }
    }
    (sjoinquals, ojoinquals)
}

/// match_unique_clauses reports whether the kept relation has each restriction that proved the removed one unique,
/// as Postgres' function of the same name does.
fn match_unique_clauses(root: &mut PlannerInfo<'_, '_>, outer: usize, uclauses: &[RinfoId], relid: usize) -> bool {
    let outer_relid = root.rels[outer].relid;
    for &rinfo in uclauses {
        let r = root.rinfos[rinfo].clone();
        let clause = change_var_nodes(root.glob, r.clause.clone(), relid, outer_relid);
        let Some((left, right)) = binary_op_args(&clause) else { return false };
        let (iclause, c1) = if r.left_relids.is_empty() { (right, left) } else { (left, right) };
        let matched = root.rels[outer].baserestrictinfo.iter().any(|&o| {
            let o = &root.rinfos[o];
            if o.mergeopfamilies.is_empty() {
                return false;
            }
            let Some((oleft, oright)) = binary_op_args(&o.clause) else { return false };
            let (oclause, c2) = if o.left_relids.is_empty() { (oright, oleft) } else { (oleft, oright) };
            iclause == oclause && c1 == c2
        });
        if !matched {
            return false;
        }
    }
    true
}

/// remove_self_joins_one_group removes the self joins among relations of one table, reporting whether it removed
/// any, as Postgres' function of the same name does.
fn remove_self_joins_one_group(root: &mut PlannerInfo<'_, '_>, relids: &Relids) -> bool {
    let mut removed = false;
    let members: Vec<usize> = relids.members().collect();
    for (ri, &r) in members.iter().enumerate() {
        for &k in &members[ri + 1..] {
            let jinfo_check = root.join_info_list.iter().all(|&s| {
                let info = &root.sjinfos[s];
                info.syn_lefthand.is_member(k) == info.syn_lefthand.is_member(r)
                    && info.syn_righthand.is_member(k) == info.syn_righthand.is_member(r)
            });
            if !jinfo_check {
                continue;
            }
            let joinrelids = Relids::singleton(r).with_member(k);
            let rrel_relids = root.rels[r].relids.clone();
            let restrictlist = generate_join_implied_equalities(root, &joinrelids, &rrel_relids, k, 0);
            if restrictlist.is_empty() {
                continue;
            }
            let (mut selfjoinquals, otherjoinquals) = split_selfjoin_quals(root, &restrictlist);
            selfjoinquals.extend(root.rels[k].baserestrictinfo.clone());
            let mut uclauses = Vec::new();
            if !innerrel_is_unique_ext(
                root,
                &joinrelids,
                &rrel_relids,
                k,
                JoinType::Inner,
                &selfjoinquals,
                otherjoinquals.is_empty(),
                Some(&mut uclauses),
            ) {
                continue;
            }
            if !match_unique_clauses(root, r, &uclauses, k) {
                continue;
            }
            remove_self_join_rel(root, k, r);
            removed = true;
            break;
        }
    }
    removed
}

/// remove_self_joins_recurse removes the self joins among the relations of a joinlist and its sublists, grouping
/// the relations by table, reporting whether it removed any, as Postgres' function of the same name does.
fn remove_self_joins_recurse(root: &mut PlannerInfo<'_, '_>, joinlist: &[JoinList]) -> bool {
    let mut removed = false;
    let mut candidates: Vec<(String, String, usize)> = Vec::new();
    for item in joinlist {
        match item {
            JoinList::Rel(varno) => {
                if let RteKind::Relation(_, table) = &root.parse.rte(*varno).kind {
                    candidates.push((table.schema.clone(), table.name.clone(), *varno));
                }
            }
            JoinList::List(list) => removed |= remove_self_joins_recurse(root, list),
        }
    }
    candidates.sort();
    for group in candidates.chunk_by(|a, b| a.0 == b.0 && a.1 == b.1) {
        if group.len() >= 2 {
            let relids: Relids = group.iter().map(|c| c.2).collect();
            removed |= remove_self_joins_one_group(root, &relids);
        }
    }
    removed
}

/// remove_useless_self_joins removes the joins of a table to itself by a unique key, reporting whether it removed
/// any, as Postgres' function of the same name does with enable_self_join_elimination on.
pub fn remove_useless_self_joins(root: &mut PlannerInfo<'_, '_>, joinlist: &[JoinList]) -> bool {
    if !root.enables.self_join_elimination || joinlist.is_empty() || matches!(joinlist, [JoinList::Rel(_)]) {
        return false;
    }
    remove_self_joins_recurse(root, joinlist)
}
