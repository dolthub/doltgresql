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

//! Postgres' optimizer/util/placeholder.c: the planner's knowledge of PlaceHolderVars, which keep an expression
//! that a pulled-up subquery computed below an outer join from being evaluated above it.

use super::PlannerInfo;
use super::costsize::{cost_qual_eval_node, get_typavgwidth};
use super::initsplan::add_vars_to_targetlist;
use super::nodes::{JoinTreeNode, PlaceHolderInfo, Relids, VarNode};
use super::var::{pull_var_clause, pull_varnos, visit_columns};
use crate::expr::Expr;

/// find_placeholder_info returns the index in `placeholder_list` of a PlaceHolderVar's PlaceHolderInfo, building it
/// when it is new, as Postgres' function of the same name does.
pub fn find_placeholder_info(root: &mut PlannerInfo<'_, '_>, phid: usize) -> usize {
    if let Some(&i) = root.placeholder_array.get(&phid) {
        return i;
    }
    assert!(!root.placeholders_frozen, "too late to create a new PlaceHolderInfo");
    let placeholder = root.glob.placeholder(phid).clone();
    let rels_used = pull_varnos(root, &placeholder.phexpr);
    let ph_lateral = rels_used.difference(&placeholder.phrels);
    let mut ph_eval_at = rels_used.intersect(&placeholder.phrels);
    if ph_eval_at.is_empty() {
        ph_eval_at = placeholder.phrels.clone();
    }
    let ph_width = get_typavgwidth(super::nodefuncs::expr_type(root, &placeholder.phexpr), -1);
    root.placeholder_list.push(PlaceHolderInfo { phid, ph_eval_at, ph_lateral, ph_needed: Relids::new(), ph_width });
    let i = root.placeholder_list.len() - 1;
    root.placeholder_array.insert(phid, i);
    find_placeholders_in_expr(root, &placeholder.phexpr);
    i
}

/// find_placeholders_in_jointree builds the PlaceHolderInfo of each PlaceHolderVar in the join tree's quals, as
/// Postgres' function of the same name does.
pub fn find_placeholders_in_jointree(root: &mut PlannerInfo<'_, '_>) {
    if root.glob.last_ph_id() != 0 {
        let jointree = JoinTreeNode::From(Box::new(root.parse.jointree.clone()));
        find_placeholders_recurse(root, &jointree);
    }
}

/// find_placeholders_recurse is find_placeholders_in_jointree for a join tree node.
fn find_placeholders_recurse(root: &mut PlannerInfo<'_, '_>, node: &JoinTreeNode) {
    let quals = match node {
        JoinTreeNode::Rel(_) => return,
        JoinTreeNode::From(f) => {
            f.fromlist.iter().for_each(|n| find_placeholders_recurse(root, n));
            &f.quals
        }
        JoinTreeNode::Join(j) => {
            find_placeholders_recurse(root, &j.larg);
            find_placeholders_recurse(root, &j.rarg);
            &j.quals
        }
    };
    for qual in quals {
        find_placeholders_in_expr(root, qual);
    }
}

/// find_placeholders_in_expr builds the PlaceHolderInfo of each PlaceHolderVar of an expression.
fn find_placeholders_in_expr(root: &mut PlannerInfo<'_, '_>, e: &Expr) {
    for id in pull_var_clause(root.glob, e, true) {
        if let VarNode::PlaceHolderVar(phv) = root.glob.node(id) {
            let phid = phv.phid;
            find_placeholder_info(root, phid);
        }
    }
}

/// fix_placeholder_input_needed_levels marks the Vars that each PlaceHolderVar's expression reads as needed where
/// it is evaluated, as Postgres' function of the same name does.
pub fn fix_placeholder_input_needed_levels(root: &mut PlannerInfo<'_, '_>) {
    for i in 0..root.placeholder_list.len() {
        let phinfo = &root.placeholder_list[i];
        let (phexpr, eval_at) = (root.glob.placeholder(phinfo.phid).phexpr.clone(), phinfo.ph_eval_at.clone());
        let vars = pull_var_clause(root.glob, &phexpr, true);
        add_vars_to_targetlist(root, &vars, &eval_at);
    }
}

/// add_placeholders_to_base_rels adds each PlaceHolderVar that a base relation evaluates for joins above it to the
/// relation's target, as Postgres' function of the same name does.
pub fn add_placeholders_to_base_rels(root: &mut PlannerInfo<'_, '_>) {
    for i in 0..root.placeholder_list.len() {
        let phinfo = &root.placeholder_list[i];
        if let Some(varno) = phinfo.ph_eval_at.singleton_member()
            && phinfo.ph_needed.nonempty_difference(&phinfo.ph_eval_at)
        {
            let phv = root.glob.intern(VarNode::PlaceHolderVar(super::nodes::PlaceHolderVar {
                phid: phinfo.phid,
                phnullingrels: Relids::new(),
            }));
            root.rels[varno].reltarget.exprs.push(phv);
        }
    }
}

/// add_placeholders_to_joinrel adds each PlaceHolderVar that a join relation evaluates for joins above it to the
/// relation's target, with its cost and width, as Postgres' function of the same name does.
pub fn add_placeholders_to_joinrel(root: &mut PlannerInfo<'_, '_>, joinrel: usize, outer_rel: usize, inner_rel: usize) {
    let relids = root.rels[joinrel].relids.clone();
    let mut tuple_width = root.rels[joinrel].reltarget.width;
    for i in 0..root.placeholder_list.len() {
        let phinfo = root.placeholder_list[i].clone();
        if !phinfo.ph_eval_at.is_subset(&relids) {
            continue;
        }
        if phinfo.ph_needed.nonempty_difference(&relids)
            && !phinfo.ph_eval_at.is_subset(&root.rels[outer_rel].relids)
            && !phinfo.ph_eval_at.is_subset(&root.rels[inner_rel].relids)
        {
            let phv = root.glob.intern(VarNode::PlaceHolderVar(super::nodes::PlaceHolderVar {
                phid: phinfo.phid,
                phnullingrels: Relids::new(),
            }));
            let cost = cost_qual_eval_node(&root.glob.placeholder(phinfo.phid).phexpr);
            let target = &mut root.rels[joinrel].reltarget;
            target.exprs.push(phv);
            target.cost.startup += cost.startup;
            target.cost.per_tuple += cost.per_tuple;
            tuple_width += phinfo.ph_width;
        }
        root.rels[joinrel].direct_lateral_relids.add_members(&phinfo.ph_lateral);
    }
    root.rels[joinrel].reltarget.width = tuple_width;
}

/// contain_placeholder_references_to reports whether a clause holds a PlaceHolderVar evaluated over a relation, as
/// Postgres' function of the same name does.
pub fn contain_placeholder_references_to(root: &PlannerInfo<'_, '_>, clause: &[Expr], relid: usize) -> bool {
    if root.glob.last_ph_id() == 0 {
        return false;
    }
    let mut found = false;
    for e in clause {
        visit_columns(e, &mut |id| {
            if let VarNode::PlaceHolderVar(phv) = root.glob.node(id) {
                found |= root.glob.placeholder(phv.phid).phrels.is_member(relid);
            }
        });
    }
    found
}
