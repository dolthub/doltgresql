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

//! Postgres' optimizer/util/tlist.c: target lists, the sort and group clauses that refer to their entries, and path
//! targets, including splitting a target at its set-returning functions. A set-returning function is the query's
//! `Expr::SetRef` of it, whose call is its entry of the query's `target_srfs`.

use super::PlannerInfo;
use super::nodes::{PathTarget, Query, SortGroupClause, TargetEntry};
use crate::expr::Expr;

/// get_sortgroupref_tle returns the target entry that a sort or group clause number refers to, as Postgres'
/// function of the same name does.
pub fn get_sortgroupref_tle(sortref: usize, tlist: &[TargetEntry]) -> &TargetEntry {
    tlist.iter().find(|tle| tle.ressortgroupref == sortref).expect("ORDER/GROUP BY expression not found in targetlist")
}

/// get_sortgroupclause_expr returns the expression that a sort or group clause reads, as Postgres' function of the
/// same name does.
pub fn get_sortgroupclause_expr(clause: &SortGroupClause, tlist: &[TargetEntry]) -> Expr {
    get_sortgroupref_tle(clause.tle_sort_group_ref, tlist).expr.clone()
}

/// get_sortgrouplist_exprs returns the expressions that sort or group clauses read, as Postgres' function of the same
/// name does.
pub fn get_sortgrouplist_exprs(clauses: &[SortGroupClause], tlist: &[TargetEntry]) -> Vec<Expr> {
    clauses.iter().map(|c| get_sortgroupclause_expr(c, tlist)).collect()
}

/// grouping_is_sortable reports whether every group clause can sort, as Postgres' function of the same name does,
/// which every Doltgres type can.
pub fn grouping_is_sortable(_clauses: &[SortGroupClause]) -> bool {
    true
}

/// grouping_is_hashable reports whether every group clause can hash, as Postgres' function of the same name does.
pub fn grouping_is_hashable(clauses: &[SortGroupClause]) -> bool {
    clauses.iter().all(|c| c.hashable)
}

/// make_pathtarget_from_tlist returns the path target of a target list, its expressions and their clause numbers, as
/// Postgres' function of the same name does.
pub fn make_pathtarget_from_tlist(tlist: &[TargetEntry]) -> PathTarget {
    PathTarget {
        exprs: tlist.iter().map(|tle| tle.expr.clone()).collect(),
        sortgrouprefs: tlist.iter().map(|tle| tle.ressortgroupref).collect(),
        ..PathTarget::default()
    }
}

/// create_pathtarget returns the path target of a target list with its cost and width, as Postgres' macro of the
/// same name does.
pub fn create_pathtarget(root: &PlannerInfo<'_, '_>, tlist: &[TargetEntry]) -> PathTarget {
    let mut target = make_pathtarget_from_tlist(tlist);
    super::costsize::set_pathtarget_cost_width(root, &mut target);
    target
}

/// add_column_to_pathtarget appends an expression to a path target with its clause number, as Postgres' function of
/// the same name does.
pub fn add_column_to_pathtarget(target: &mut PathTarget, expr: Expr, sortgroupref: usize) {
    target.exprs.push(expr);
    target.sortgrouprefs.resize(target.exprs.len() - 1, 0);
    target.sortgrouprefs.push(sortgroupref);
}

/// add_new_column_to_pathtarget appends an expression to a path target unless it holds it already, as Postgres'
/// function of the same name does.
pub fn add_new_column_to_pathtarget(target: &mut PathTarget, expr: Expr) {
    if !target.exprs.contains(&expr) {
        add_column_to_pathtarget(target, expr, 0);
    }
}

/// add_new_columns_to_pathtarget is add_new_column_to_pathtarget for several expressions.
pub fn add_new_columns_to_pathtarget(target: &mut PathTarget, exprs: Vec<Expr>) {
    for expr in exprs {
        add_new_column_to_pathtarget(target, expr);
    }
}

/// get_pathtarget_sortgroupref returns the clause number of a path target's column, or 0.
pub fn get_pathtarget_sortgroupref(target: &PathTarget, i: usize) -> usize {
    target.sortgrouprefs.get(i).copied().unwrap_or(0)
}

/// SplitItem is an expression that splitting a target at its set-returning functions found, with the clause number
/// of the target column it came from, as Postgres' split_pathtarget_item is.
#[derive(Clone)]
struct SplitItem {
    expr: Expr,
    sortgroupref: usize,
}

/// SplitContext is what splitting a target at its set-returning functions gathers, as Postgres'
/// split_pathtarget_context is: the input target's expressions, and for each level of nested functions, the
/// functions and the Vars and functions of lower levels that they read.
struct SplitContext<'q> {
    query: &'q Query,
    input_target_exprs: Vec<Expr>,
    level_srfs: Vec<Vec<SplitItem>>,
    level_input_vars: Vec<Vec<SplitItem>>,
    level_input_srfs: Vec<Vec<SplitItem>>,
    current_input_vars: Vec<SplitItem>,
    current_input_srfs: Vec<SplitItem>,
    current_depth: usize,
    current_sgref: usize,
}

/// split_pathtarget_at_srfs splits a target whose expressions call set-returning functions into the targets of the
/// projections that compute it level by level over an input target, as Postgres' function of the same name does,
/// returning them bottom first with whether each holds set-returning functions.
pub fn split_pathtarget_at_srfs(
    root: &PlannerInfo<'_, '_>,
    target: &PathTarget,
    input_target: Option<&PathTarget>,
) -> (Vec<PathTarget>, Vec<bool>) {
    let mut context = SplitContext {
        query: &root.parse,
        input_target_exprs: input_target.map(|t| t.exprs.clone()).unwrap_or_default(),
        level_srfs: vec![Vec::new()],
        level_input_vars: vec![Vec::new()],
        level_input_srfs: vec![Vec::new()],
        current_input_vars: Vec::new(),
        current_input_srfs: Vec::new(),
        current_depth: 0,
        current_sgref: 0,
    };
    let mut max_depth = 0;
    let mut need_extra_projection = false;
    for (i, node) in target.exprs.iter().enumerate() {
        context.current_sgref = get_pathtarget_sortgroupref(target, i);
        context.current_depth = 0;
        split_pathtarget_walker(node, &mut context);
        if context.current_depth == 0 {
            continue;
        }
        if max_depth < context.current_depth {
            max_depth = context.current_depth;
            need_extra_projection = false;
        }
        if max_depth == context.current_depth && !matches!(node, Expr::SetRef(_)) {
            need_extra_projection = true;
        }
    }
    if max_depth == 0 {
        return (vec![target.clone()], vec![false]);
    }
    let input_vars = std::mem::take(&mut context.current_input_vars);
    let input_srfs = std::mem::take(&mut context.current_input_srfs);
    if need_extra_projection {
        context.level_srfs.push(Vec::new());
        context.level_input_vars.push(input_vars);
        context.level_input_srfs.push(input_srfs);
    } else {
        context.level_input_vars[max_depth].extend(input_vars);
        context.level_input_srfs[max_depth].extend(input_srfs);
    }
    let (mut targets, mut contain_srfs) = (Vec::new(), Vec::new());
    let mut prev_level_tlist: Vec<Expr> = Vec::new();
    let levels = context.level_srfs.len();
    for level in 0..levels {
        let ntarget = match level + 1 == levels {
            true => target.clone(),
            false => {
                let mut ntarget = PathTarget::default();
                add_items(&mut ntarget, &context.level_srfs[level]);
                for input_vars in &context.level_input_vars[level + 1..] {
                    add_items(&mut ntarget, input_vars);
                }
                for input_srfs in &context.level_input_srfs[level + 1..] {
                    let items: Vec<SplitItem> =
                        input_srfs.iter().filter(|item| prev_level_tlist.contains(&item.expr)).cloned().collect();
                    add_items(&mut ntarget, &items);
                }
                super::costsize::set_pathtarget_cost_width(root, &mut ntarget);
                ntarget
            }
        };
        contain_srfs.push(!context.level_srfs[level].is_empty());
        prev_level_tlist = ntarget.exprs.clone();
        targets.push(ntarget);
    }
    (targets, contain_srfs)
}

/// add_items adds split items to a target, as Postgres' add_sp_items_to_pathtarget does: an expression the target
/// holds takes the item's clause number.
fn add_items(target: &mut PathTarget, items: &[SplitItem]) {
    for item in items {
        let found = target.exprs.iter().enumerate().position(|(i, e)| {
            let sgref = get_pathtarget_sortgroupref(target, i);
            (item.sortgroupref == sgref || item.sortgroupref == 0 || sgref == 0) && *e == item.expr
        });
        match found {
            Some(i) => {
                if item.sortgroupref != 0 {
                    target.sortgrouprefs.resize(target.exprs.len(), 0);
                    target.sortgrouprefs[i] = item.sortgroupref;
                }
            }
            None => add_column_to_pathtarget(target, item.expr.clone(), item.sortgroupref),
        }
    }
}

/// split_pathtarget_walker finds the set-returning functions of an expression and the expressions they read that an
/// input target holds or that are Vars, aggregates, or window calls, as Postgres' function of the same name does.
fn split_pathtarget_walker(node: &Expr, context: &mut SplitContext<'_>) {
    if context.input_target_exprs.contains(node)
        || matches!(node, Expr::Column(_) | Expr::AggRef(_) | Expr::WindowRef(_) | Expr::Grouping(..))
    {
        context.current_input_vars.push(SplitItem { expr: node.clone(), sortgroupref: context.current_sgref });
        return;
    }
    if let Expr::SetRef(k) = node {
        let item = SplitItem { expr: node.clone(), sortgroupref: context.current_sgref };
        let save_input_vars = std::mem::take(&mut context.current_input_vars);
        let save_input_srfs = std::mem::take(&mut context.current_input_srfs);
        let save_current_depth = context.current_depth;
        context.current_depth = 0;
        context.current_sgref = 0;
        let call = context.query.target_srfs[*k].clone();
        call.visit_children(&mut |child| split_pathtarget_walker(child, context));
        let srf_depth = context.current_depth + 1;
        if srf_depth >= context.level_srfs.len() {
            context.level_srfs.push(Vec::new());
            context.level_input_vars.push(Vec::new());
            context.level_input_srfs.push(Vec::new());
        }
        context.level_srfs[srf_depth].push(item.clone());
        let vars = std::mem::take(&mut context.current_input_vars);
        let srfs = std::mem::take(&mut context.current_input_srfs);
        context.level_input_vars[srf_depth].extend(vars);
        context.level_input_srfs[srf_depth].extend(srfs);
        context.current_input_vars = save_input_vars;
        context.current_input_srfs = save_input_srfs;
        context.current_input_srfs.push(item);
        context.current_depth = save_current_depth.max(srf_depth);
        return;
    }
    context.current_sgref = 0;
    node.visit_children(&mut |child| split_pathtarget_walker(child, context));
}
