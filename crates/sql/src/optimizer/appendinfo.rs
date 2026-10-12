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

//! Postgres' optimizer/util/appendinfo.c: translating expressions over an append relation's columns into ones over a
//! member's.

use super::PlannerInfo;
use super::nodes::{AppendRelInfo, PlannerGlobal, RelOptKind, Relids, RinfoId, Var, VarNode};
use super::var::replace_columns;
use crate::expr::Expr;

/// add_row_identity_var adds a column that identifies the rows of the result relation to the target list that the
/// plan computes, as Postgres' function of the same name does for a result relation that is not an inheritance
/// parent, which is the only kind Doltgres has.
pub fn add_row_identity_var(root: &mut PlannerInfo<'_, '_>, orig_var: Expr) {
    root.processed_tlist.push(super::nodes::TargetEntry { expr: orig_var, resjunk: true, ressortgroupref: 0 });
}

/// add_row_identity_columns adds the columns that identify the rows of an UPDATE's, DELETE's, or MERGE's result
/// relation, as Postgres' function of the same name does: every column of the row, as Postgres identifies a foreign
/// table's rows by its whole row, since Doltgres has no ctid and its executor reads each old value.
pub fn add_row_identity_columns(root: &mut PlannerInfo<'_, '_>, rtindex: usize) {
    for attno in 0..root.parse.rte(rtindex).coltypes.len() {
        let var = root.glob.var(rtindex, attno, Relids::new());
        add_row_identity_var(root, var);
    }
}

/// adjust_appendrel_attrs rewrites an expression over an append relation's columns into one over a member's, as
/// Postgres' function of the same name does. A PlaceHolderVar evaluated at the append relation becomes its
/// rewritten expression, which the member's scan computes, since every copy of a PlaceHolderVar shares its expression.
pub fn adjust_appendrel_attrs(glob: &mut PlannerGlobal, e: Expr, appinfo: &AppendRelInfo) -> Expr {
    replace_columns(e, &mut |id| match glob.node(id).clone() {
        VarNode::Var(var) if var.varno == appinfo.parent_relid => {
            let newnode = appinfo.translated_vars[var.varattno].clone();
            if var.varnullingrels.is_empty() {
                return newnode;
            }
            let Expr::Column(newid) = newnode else { unreachable!("failed to apply nullingrels to a non-Var") };
            match glob.node(newid).clone() {
                VarNode::Var(newvar) => {
                    let varnullingrels = newvar.varnullingrels.union(&var.varnullingrels);
                    glob.intern(VarNode::Var(Var { varnullingrels, ..newvar }))
                }
                VarNode::PlaceHolderVar(_) => unreachable!("failed to apply nullingrels to a non-Var"),
            }
        }
        VarNode::PlaceHolderVar(phv) if glob.placeholder(phv.phid).phrels.is_member(appinfo.parent_relid) => {
            let phexpr = glob.placeholder(phv.phid).phexpr.clone();
            adjust_appendrel_attrs(glob, phexpr, appinfo)
        }
        _ => Expr::Column(id),
    })
}

/// adjust_appendrel_restrictinfo is adjust_appendrel_attrs for a RestrictInfo, returning a copy over a member's
/// columns and relations, with its cached estimates reset, as Postgres' adjust_appendrel_attrs_mutator copies one.
pub fn adjust_appendrel_restrictinfo(
    root: &mut PlannerInfo<'_, '_>,
    rinfo: RinfoId,
    appinfo: &AppendRelInfo,
) -> RinfoId {
    let mut newinfo = root.rinfos[rinfo].clone();
    newinfo.clause = adjust_appendrel_attrs(root.glob, newinfo.clause, appinfo);
    newinfo.orclause = newinfo.orclause.map(|arms| {
        arms.into_iter()
            .map(|arm| arm.into_iter().map(|r| adjust_appendrel_restrictinfo(root, r, appinfo)).collect())
            .collect()
    });
    newinfo.clause_relids = adjust_child_relids(&newinfo.clause_relids, appinfo);
    newinfo.required_relids = adjust_child_relids(&newinfo.required_relids, appinfo);
    newinfo.outer_relids = adjust_child_relids(&newinfo.outer_relids, appinfo);
    newinfo.left_relids = adjust_child_relids(&newinfo.left_relids, appinfo);
    newinfo.right_relids = adjust_child_relids(&newinfo.right_relids, appinfo);
    newinfo.norm_selec = std::cell::Cell::new(-1.0);
    newinfo.outer_selec = std::cell::Cell::new(-1.0);
    newinfo.left_em = None;
    newinfo.right_em = None;
    newinfo.scansel_cache = Default::default();
    root.rinfos.push(newinfo);
    root.rinfos.len() - 1
}

/// adjust_child_relids replaces an append relation by its member in a set of relations, as Postgres' function of the
/// same name does.
pub fn adjust_child_relids(relids: &Relids, appinfo: &AppendRelInfo) -> Relids {
    super::var::adjust_relid_set(relids, appinfo.parent_relid, appinfo.child_relid)
}

/// adjust_appendrel_attrs_multilevel rewrites an expression over the columns of a member's topmost append relation
/// into one over the member's, through each append relation between them, as Postgres' function of the same name
/// does.
pub fn adjust_appendrel_attrs_multilevel(root: &mut PlannerInfo<'_, '_>, e: Expr, child_rel: usize) -> Expr {
    let appinfo = member_appinfo(root, child_rel);
    let e = match root.rels[appinfo.parent_relid].reloptkind {
        RelOptKind::OtherMemberRel => adjust_appendrel_attrs_multilevel(root, e, appinfo.parent_relid),
        _ => e,
    };
    adjust_appendrel_attrs(root.glob, e, &appinfo)
}

/// adjust_appendrel_restrictinfo_multilevel is adjust_appendrel_attrs_multilevel for a RestrictInfo.
pub fn adjust_appendrel_restrictinfo_multilevel(
    root: &mut PlannerInfo<'_, '_>,
    rinfo: RinfoId,
    child_rel: usize,
) -> RinfoId {
    let appinfo = member_appinfo(root, child_rel);
    let rinfo = match root.rels[appinfo.parent_relid].reloptkind {
        RelOptKind::OtherMemberRel => adjust_appendrel_restrictinfo_multilevel(root, rinfo, appinfo.parent_relid),
        _ => rinfo,
    };
    adjust_appendrel_restrictinfo(root, rinfo, &appinfo)
}

/// member_appinfo returns the AppendRelInfo of a member of an append relation, as Postgres' append_rel_array holds it.
pub fn member_appinfo(root: &PlannerInfo<'_, '_>, child_rel: usize) -> AppendRelInfo {
    root.parse.append_rel_list.iter().find(|a| a.child_relid == child_rel).expect("a member").clone()
}
