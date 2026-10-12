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

//! Postgres' optimizer/util/paramassign.c: the parameters by which the inner side of a nested loop reads the Vars and
//! PlaceHolderVars of its outer side. Doltgres' executor pushes the outer row of a lateral join for its inner side to
//! read as an enclosing row, so the nested loop that supplies a parameter replaces it by that row's column.

use super::PlannerInfo;
use super::nodes::{NestLoopParam, Relids, VarNode};
use crate::expr::Expr;

/// replace_nestloop_param returns the nestloop parameter of a Var of an outer relation or a PlaceHolderVar evaluated
/// over outer relations, reusing the one that the nested loops above have yet to supply or numbering a new one, as
/// Postgres' replace_nestloop_param_var and replace_nestloop_param_placeholdervar do.
pub fn replace_nestloop_param(root: &PlannerInfo<'_, '_>, paramval: Expr) -> Expr {
    let mut params = root.cur_outer_params.borrow_mut();
    if let Some(nlp) = params.iter().find(|nlp| nlp.paramval == paramval) {
        return Expr::NestLoopParam(nlp.paramno);
    }
    let paramno = generate_new_exec_param(root);
    params.push(NestLoopParam { paramno, paramval });
    Expr::NestLoopParam(paramno)
}

/// identify_current_nestloop_params removes and returns the nestloop parameters that a nested loop's outer relations
/// supply: its Vars, and the PlaceHolderVars evaluated over them and the nested loop's own outer relations, as
/// Postgres' function of the same name does.
pub fn identify_current_nestloop_params(
    root: &PlannerInfo<'_, '_>,
    leftrelids: &Relids,
    outerrelids: &Relids,
) -> Vec<NestLoopParam> {
    let allleftrelids = leftrelids.union(outerrelids);
    let supplied = |nlp: &NestLoopParam| {
        let Expr::Column(id) = nlp.paramval else { return false };
        match root.glob.node(id) {
            VarNode::Var(var) => leftrelids.is_member(var.varno),
            VarNode::PlaceHolderVar(phv) => {
                let eval_at = &root.placeholder_list[root.placeholder_array[&phv.phid]].ph_eval_at;
                eval_at.is_subset(&allleftrelids) && eval_at.overlap(leftrelids)
            }
        }
    };
    let (result, rest) = root.cur_outer_params.take().into_iter().partition(supplied);
    root.cur_outer_params.replace(rest);
    result
}

/// generate_new_exec_param returns the number of a new nestloop parameter, as Postgres' function of the same name
/// numbers a PARAM_EXEC parameter.
pub fn generate_new_exec_param(root: &PlannerInfo<'_, '_>) -> usize {
    let paramno = root.num_nestloop_params.get();
    root.num_nestloop_params.set(paramno + 1);
    paramno
}

/// replace_outer_agg rewrites the arguments of a SubPlan whose subquery reads aggregate calls or GROUPING operations
/// of the query around it, which the binder gives it as columns past that query's row of `width` columns, into those
/// expressions of the query, as Postgres' replace_outer_agg and replace_outer_grouping pass them to the subquery as
/// parameters.
pub(crate) fn replace_outer_agg(e: Expr, width: usize, outer_exprs: &[Expr]) -> Expr {
    let replace = |subplan: &mut crate::expr::SubPlan| {
        for arg in &mut subplan.args {
            if let Expr::Column(c) = *arg
                && c >= width
            {
                *arg = outer_exprs[c - width].clone();
            }
        }
    };
    match e {
        Expr::SubPlan(mut subplan) => {
            replace(&mut subplan);
            Expr::SubPlan(subplan)
        }
        Expr::AlternativeSubPlan(mut subplans) => {
            subplans.iter_mut().for_each(replace);
            Expr::AlternativeSubPlan(subplans)
        }
        other => other,
    }
}
