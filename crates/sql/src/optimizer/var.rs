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

//! Postgres' optimizer/util/var.c and the parts of rewrite/rewriteManip.c that the planner calls: finding the Vars
//! and relations of expressions, and rewriting their Vars' range table indexes and nulling relations.

use std::collections::HashSet;

use super::PlannerInfo;
use super::nodes::{JoinTreeNode, PlaceHolderVar, PlannerGlobal, Query, Relids, Var, VarNode};
use crate::expr::Expr;

/// replace_columns rewrites each Var and PlaceHolderVar of an expression, outside subquery plans.
pub fn replace_columns(e: Expr, f: &mut dyn FnMut(usize) -> Expr) -> Expr {
    match e {
        Expr::Column(id) => f(id),
        other => other.map_children(&mut |c| replace_columns(c, f)),
    }
}

/// visit_columns calls a function with the ID of each Var and PlaceHolderVar of an expression, outside subquery
/// plans and without descending into PlaceHolderVars.
pub fn visit_columns(e: &Expr, f: &mut dyn FnMut(usize)) {
    e.visit(&mut |x| {
        if let Expr::Column(id) = x {
            f(*id);
        }
    });
}

/// pull_varnos returns the relations that an expression reads, with the outer joins that can make its Vars NULL, as
/// Postgres' function of the same name does.
pub fn pull_varnos(root: &PlannerInfo<'_, '_>, e: &Expr) -> Relids {
    let mut varnos = Relids::new();
    visit_columns(e, &mut |id| match root.glob.node(id) {
        VarNode::Var(var) => {
            varnos.add_member(var.varno);
            varnos.add_members(&var.varnullingrels);
        }
        VarNode::PlaceHolderVar(phv) => {
            let phrels = &root.glob.placeholder(phv.phid).phrels;
            match root.placeholder_array.get(&phv.phid) {
                None => varnos.add_members(phrels),
                Some(&i) => varnos.add_members(&root.placeholder_list[i].ph_eval_at),
            }
            varnos.add_members(&phv.phnullingrels);
        }
    });
    varnos
}

/// pull_varnos_list is pull_varnos over a list of expressions.
pub fn pull_varnos_list(root: &PlannerInfo<'_, '_>, exprs: &[Expr]) -> Relids {
    exprs.iter().fold(Relids::new(), |relids, e| relids.union(&pull_varnos(root, e)))
}

/// pull_var_clause returns the IDs of the Vars of an expression, with its PlaceHolderVars or with the Vars of their
/// expressions, as Postgres' function of the same name does with PVC_INCLUDE_PLACEHOLDERS or
/// PVC_RECURSE_PLACEHOLDERS.
pub fn pull_var_clause(glob: &PlannerGlobal, e: &Expr, include_placeholders: bool) -> Vec<usize> {
    let mut vars = Vec::new();
    visit_columns(e, &mut |id| match glob.node(id) {
        VarNode::PlaceHolderVar(phv) if !include_placeholders => {
            vars.extend(pull_var_clause(glob, &glob.placeholder(phv.phid).phexpr, false));
        }
        _ => vars.push(id),
    });
    vars
}

/// contain_var_clause reports whether an expression reads a Var or PlaceHolderVar of the query, as Postgres'
/// function of the same name does.
pub fn contain_var_clause(e: &Expr) -> bool {
    let mut found = false;
    visit_columns(e, &mut |_| found = true);
    found
}

/// mutate_query rewrites the expressions of a query's target list and join tree.
pub fn mutate_query(query: &mut Query, f: &mut dyn FnMut(Expr) -> Expr) {
    for e in query.target_list.iter_mut().flatten() {
        *e = f(std::mem::replace(e, Expr::Const(crate::types::Value::Null)));
    }
    let mut node = JoinTreeNode::From(Box::new(std::mem::replace(
        &mut query.jointree,
        super::nodes::FromExpr { fromlist: Vec::new(), quals: Vec::new() },
    )));
    mutate_jointree(&mut node, f);
    let JoinTreeNode::From(jointree) = node else { unreachable!("the join tree stays a FROM list") };
    query.jointree = *jointree;
}

/// mutate_jointree rewrites the quals of a join tree.
pub fn mutate_jointree(node: &mut JoinTreeNode, f: &mut dyn FnMut(Expr) -> Expr) {
    let quals = match node {
        JoinTreeNode::Rel(_) => return,
        JoinTreeNode::From(from) => {
            from.fromlist.iter_mut().for_each(|n| mutate_jointree(n, f));
            &mut from.quals
        }
        JoinTreeNode::Join(join) => {
            mutate_jointree(&mut join.larg, f);
            mutate_jointree(&mut join.rarg, f);
            &mut join.quals
        }
    };
    *quals = std::mem::take(quals).into_iter().map(&mut *f).collect();
}

/// VarMutator rewrites the Vars and PlaceHolderVars of expressions, rewriting each PlaceHolderVar's shared
/// expression and relations once.
struct VarMutator<'g, F: FnMut(&mut PlannerGlobal, VarNode) -> VarNode, P: FnMut(&mut Relids)> {
    glob: &'g mut PlannerGlobal,
    node: F,
    phrels: P,
    done: HashSet<usize>,
}

impl<F: FnMut(&mut PlannerGlobal, VarNode) -> VarNode, P: FnMut(&mut Relids)> VarMutator<'_, F, P> {
    /// expr rewrites an expression.
    fn expr(&mut self, e: Expr) -> Expr {
        replace_columns(e, &mut |id| {
            let node = self.glob.node(id).clone();
            if let VarNode::PlaceHolderVar(phv) = &node
                && self.done.insert(phv.phid)
            {
                let mut placeholder = self.glob.placeholder(phv.phid).clone();
                placeholder.phexpr = self.expr(placeholder.phexpr);
                (self.phrels)(&mut placeholder.phrels);
                self.glob.placeholders[phv.phid - 1] = placeholder;
            }
            let node = (self.node)(self.glob, node);
            self.glob.intern(node)
        })
    }
}

/// mutate_vars rewrites the Vars and PlaceHolderVars of expressions with one function, and the relations of each
/// PlaceHolderVar once with another, returning a function that rewrites an expression.
fn mutate_vars<'g>(
    glob: &'g mut PlannerGlobal,
    node: impl FnMut(&mut PlannerGlobal, VarNode) -> VarNode + 'g,
    phrels: impl FnMut(&mut Relids) + 'g,
) -> impl FnMut(Expr) -> Expr + 'g {
    let mut mutator = VarMutator { glob, node, phrels, done: HashSet::new() };
    move |e| mutator.expr(e)
}

/// offset_relid_set returns a set with each member moved up by an offset.
fn offset_relid_set(relids: &Relids, offset: usize) -> Relids {
    relids.members().map(|r| r + offset).collect()
}

/// offset_var_nodes moves the range table indexes of a query's Vars, PlaceHolderVars, and join tree up by an offset,
/// as Postgres' OffsetVarNodes does when a subquery's range table joins its parent's.
pub fn offset_var_nodes(glob: &mut PlannerGlobal, query: &mut Query, offset: usize) {
    let mut f = mutate_vars(
        glob,
        move |_, node| match node {
            VarNode::Var(var) => VarNode::Var(Var {
                varno: var.varno + offset,
                varnullingrels: offset_relid_set(&var.varnullingrels, offset),
                ..var
            }),
            VarNode::PlaceHolderVar(phv) => VarNode::PlaceHolderVar(PlaceHolderVar {
                phnullingrels: offset_relid_set(&phv.phnullingrels, offset),
                ..phv
            }),
        },
        move |phrels| *phrels = offset_relid_set(phrels, offset),
    );
    mutate_query(query, &mut f);
    offset_jointree(&mut JoinTreeNodeRef::From(&mut query.jointree), offset);
}

/// JoinTreeNodeRef is a join tree node or a query's top FROM list, which the join tree rewriters walk alike.
enum JoinTreeNodeRef<'a> {
    Node(&'a mut JoinTreeNode),
    From(&'a mut super::nodes::FromExpr),
}

/// offset_jointree moves the range table indexes of a join tree up by an offset.
fn offset_jointree(node: &mut JoinTreeNodeRef<'_>, offset: usize) {
    let fromlist = match node {
        JoinTreeNodeRef::From(from) => &mut from.fromlist,
        JoinTreeNodeRef::Node(node) => match &mut **node {
            JoinTreeNode::Rel(varno) => {
                *varno += offset;
                return;
            }
            JoinTreeNode::From(from) => &mut from.fromlist,
            JoinTreeNode::Join(join) => {
                if join.rtindex != 0 {
                    join.rtindex += offset;
                }
                offset_jointree(&mut JoinTreeNodeRef::Node(&mut join.larg), offset);
                offset_jointree(&mut JoinTreeNodeRef::Node(&mut join.rarg), offset);
                return;
            }
        },
    };
    for n in fromlist.iter_mut() {
        offset_jointree(&mut JoinTreeNodeRef::Node(n), offset);
    }
}

/// adjust_relid_set returns a set with one member replaced by another, or removed when the other is 0, as
/// Postgres' function of the same name does.
pub fn adjust_relid_set(relids: &Relids, oldrelid: usize, newrelid: usize) -> Relids {
    if !relids.is_member(oldrelid) {
        return relids.clone();
    }
    let relids = relids.clone().without_member(oldrelid);
    if newrelid == 0 { relids } else { relids.with_member(newrelid) }
}

/// change_var_nodes rewrites the Vars and PlaceHolderVars of an expression from one range table index to another,
/// as Postgres' ChangeVarNodes does.
pub fn change_var_nodes(glob: &mut PlannerGlobal, e: Expr, rt_index: usize, new_index: usize) -> Expr {
    let mut f = change_var_nodes_fn(glob, rt_index, new_index);
    f(e)
}

/// change_var_nodes_fn returns a function that rewrites expressions from one range table index to another, as
/// Postgres' ChangeVarNodes does.
pub fn change_var_nodes_fn(
    glob: &mut PlannerGlobal,
    rt_index: usize,
    new_index: usize,
) -> impl FnMut(Expr) -> Expr + '_ {
    mutate_vars(
        glob,
        move |_, node| match node {
            VarNode::Var(var) => VarNode::Var(Var {
                varno: if var.varno == rt_index { new_index } else { var.varno },
                varnullingrels: adjust_relid_set(&var.varnullingrels, rt_index, new_index),
                ..var
            }),
            VarNode::PlaceHolderVar(phv) => VarNode::PlaceHolderVar(PlaceHolderVar {
                phnullingrels: adjust_relid_set(&phv.phnullingrels, rt_index, new_index),
                ..phv
            }),
        },
        move |phrels| *phrels = adjust_relid_set(phrels, rt_index, new_index),
    )
}

/// add_nulling_relids adds outer joins to the nulling relations of an expression's Vars of target relations, or of
/// every Var without targets, and of its PlaceHolderVars that overlap them, as Postgres' function of the same name
/// does.
pub fn add_nulling_relids(glob: &mut PlannerGlobal, e: Expr, target_relids: Option<&Relids>, added: &Relids) -> Expr {
    replace_columns(e, &mut |id| match glob.node(id).clone() {
        VarNode::Var(var) if target_relids.is_none_or(|t| t.is_member(var.varno)) => {
            let varnullingrels = var.varnullingrels.union(added);
            glob.intern(VarNode::Var(Var { varnullingrels, ..var }))
        }
        VarNode::PlaceHolderVar(phv) if target_relids.is_none_or(|t| t.overlap(&glob.placeholder(phv.phid).phrels)) => {
            let phnullingrels = phv.phnullingrels.union(added);
            glob.intern(VarNode::PlaceHolderVar(PlaceHolderVar { phnullingrels, ..phv }))
        }
        _ => Expr::Column(id),
    })
}

/// remove_nulling_relids_fn returns a function that removes outer joins from the nulling relations of an
/// expression's Vars, except those of excepted relations, and of its PlaceHolderVars, as Postgres'
/// remove_nulling_relids does.
pub fn remove_nulling_relids_fn<'g>(
    glob: &'g mut PlannerGlobal,
    removable: &Relids,
    except: &Relids,
) -> impl FnMut(Expr) -> Expr + 'g {
    let (removable, except) = (removable.clone(), except.clone());
    let mut done: HashSet<usize> = HashSet::new();
    move |e| remove_nulling(glob, e, &removable, &except, &mut done)
}

/// remove_nulling is remove_nulling_relids over one expression, rewriting each PlaceHolderVar's expression and
/// relations once.
fn remove_nulling(
    glob: &mut PlannerGlobal,
    e: Expr,
    removable: &Relids,
    except: &Relids,
    done: &mut HashSet<usize>,
) -> Expr {
    replace_columns(e, &mut |id| match glob.node(id).clone() {
        VarNode::Var(var) if !except.is_member(var.varno) && var.varnullingrels.overlap(removable) => {
            let varnullingrels = var.varnullingrels.difference(removable);
            glob.intern(VarNode::Var(Var { varnullingrels, ..var }))
        }
        VarNode::PlaceHolderVar(phv) if !glob.placeholder(phv.phid).phrels.overlap(except) => {
            if done.insert(phv.phid) {
                let mut placeholder = glob.placeholder(phv.phid).clone();
                placeholder.phexpr = remove_nulling(glob, placeholder.phexpr, removable, except, done);
                placeholder.phrels = placeholder.phrels.difference(removable);
                glob.placeholders[phv.phid - 1] = placeholder;
            }
            let phnullingrels = phv.phnullingrels.difference(removable);
            glob.intern(VarNode::PlaceHolderVar(PlaceHolderVar { phnullingrels, ..phv }))
        }
        _ => Expr::Column(id),
    })
}

/// remove_nulling_relids_query is remove_nulling_relids over a query.
pub fn remove_nulling_relids_query(glob: &mut PlannerGlobal, query: &mut Query, removable: &Relids, except: &Relids) {
    let mut f = remove_nulling_relids_fn(glob, removable, except);
    mutate_query(query, &mut f);
}
