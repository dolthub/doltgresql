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

//! Postgres' optimizer/prep/prepjointree.c: preparing the join tree by pulling subqueries up into it, reducing outer
//! joins that conditions above make inner or anti joins, and removing RESULT relations that join nothing.

use std::collections::HashSet;

use super::clauses::{
    contain_nonstrict_functions, contain_subplans, contain_volatile_functions, find_forced_null_var,
    find_forced_null_vars, find_nonnullable_rels, find_nonnullable_vars,
};
use super::nodes::{
    FromExpr, JoinTreeNode, JoinType, PlaceHolderVar, PlannerGlobal, Query, Relids, RteKind, Var, VarNode,
};
use super::var::{
    add_nulling_relids, contain_var_clause, offset_var_nodes, remove_nulling_relids_query, replace_columns,
};
use crate::expr::Expr;
use crate::plan::Plan;

/// pull_up_subqueries pulls each simple subquery, single-row VALUES list, and constant function of the join tree up
/// into the query, as Postgres' function of the same name does.
pub fn pull_up_subqueries(glob: &mut PlannerGlobal, parse: &mut Query) {
    for i in 0..parse.jointree.fromlist.len() {
        pull_up_subqueries_recurse(glob, parse, &mut vec![i], false);
    }
}

/// node_at returns the join tree node at a path of indexes into FROM lists and joins' two sides.
fn node_at<'q>(parse: &'q mut Query, path: &[usize]) -> &'q mut JoinTreeNode {
    let mut node = &mut parse.jointree.fromlist[path[0]];
    for &i in &path[1..] {
        node = match node {
            JoinTreeNode::From(from) => &mut from.fromlist[i],
            JoinTreeNode::Join(join) => match i {
                0 => &mut join.larg,
                _ => &mut join.rarg,
            },
            JoinTreeNode::Rel(_) => unreachable!("a relation has no children"),
        };
    }
    node
}

/// pull_up_subqueries_recurse is pull_up_subqueries for the join tree node at a path, given whether it is under an
/// outer join, replacing the node in place so that the whole join tree's Vars are rewritten as each pull-up needs.
fn pull_up_subqueries_recurse(
    glob: &mut PlannerGlobal,
    parse: &mut Query,
    path: &mut Vec<usize>,
    lowest_outer_join: bool,
) {
    let (children, lowest_outer_join) = match node_at(parse, path) {
        JoinTreeNode::Rel(varno) => {
            let varno = *varno;
            let pulled = match &parse.rte(varno).kind {
                RteKind::Subquery(subquery, _) if is_simple_subquery(glob, subquery) => {
                    Some(pull_up_simple_subquery(glob, parse, varno))
                }
                RteKind::Plan(Plan::Values(_)) if !lowest_outer_join && is_simple_values(glob, parse, varno) => {
                    Some(pull_up_simple_values(glob, parse, varno))
                }
                RteKind::Plan(Plan::Function { .. }) => Some(pull_up_constant_function(glob, parse, varno)),
                _ => None,
            };
            if let Some(node) = pulled {
                *node_at(parse, path) = node;
            }
            return;
        }
        JoinTreeNode::From(from) => (from.fromlist.len(), lowest_outer_join),
        JoinTreeNode::Join(join) => (2, lowest_outer_join || join.jointype != JoinType::Inner),
    };
    for i in 0..children {
        path.push(i);
        pull_up_subqueries_recurse(glob, parse, path, lowest_outer_join);
        path.pop();
    }
}

/// pull_up_simple_subquery pulls a simple subquery up into the query, after pulling up its own subqueries, returning
/// the join tree node that replaces its reference, as Postgres' function of the same name does.
fn pull_up_simple_subquery(glob: &mut PlannerGlobal, parse: &mut Query, varno: usize) -> JoinTreeNode {
    let RteKind::Subquery(subquery, _) = &parse.rte(varno).kind else { unreachable!("the entry is a subquery") };
    let mut subquery = (**subquery).clone();
    pull_up_subqueries(glob, &mut subquery);
    if !is_simple_subquery(glob, &subquery) {
        return JoinTreeNode::Rel(varno);
    }
    let rtoffset = parse.rtable.len();
    offset_var_nodes(glob, &mut subquery, rtoffset);
    let targetlist: Vec<Expr> = subquery.target_list.iter().map(|tle| tle.expr.clone()).collect();
    let mut rvcontext = PullupReplaceVars {
        rv_cache: vec![None; targetlist.len()],
        targetlist,
        varno,
        wrap_option: ReplaceWrap::None,
        done: HashSet::new(),
    };
    perform_pullup_replace_vars(glob, parse, &mut rvcontext);
    parse.rtable.extend(std::mem::take(&mut subquery.rtable));
    if glob.last_ph_id() != 0 {
        let subrelids = get_relids_in_jointree(&JoinTreeNode::From(Box::new(subquery.jointree.clone())), true, false);
        substitute_phv_relids(glob, parse, varno, &subrelids);
    }
    parse.rtable[varno - 1].kind = RteKind::Result;
    match (subquery.jointree.quals.is_empty(), subquery.jointree.fromlist.len()) {
        (true, 1) => subquery.jointree.fromlist.pop().expect("one member"),
        _ => JoinTreeNode::From(Box::new(subquery.jointree)),
    }
}

/// is_simple_subquery reports whether a subquery can be pulled up into its parent, as Postgres' function of the same
/// name decides: it has no aggregates, windows, set-returning functions, grouping, HAVING, ORDER BY, DISTINCT, or
/// LIMIT, and its output runs no volatile function.
fn is_simple_subquery(glob: &PlannerGlobal, subquery: &Query) -> bool {
    !(subquery.has_aggs
        || !subquery.window_funcs.is_empty()
        || !subquery.target_srfs.is_empty()
        || !subquery.group_clause.is_empty()
        || subquery.grouping_sets.is_some()
        || subquery.having_qual.is_some()
        || !subquery.sort_clause.is_empty()
        || !subquery.distinct_clause.is_empty()
        || subquery.limit_offset.is_some()
        || subquery.limit_count.is_some())
        && !subquery.target_list.iter().any(|tle| contain_volatile_functions(glob, &tle.expr))
}

/// pull_up_simple_values replaces the query's only relation, a VALUES list of one row, by its expressions, as
/// Postgres' function of the same name does.
fn pull_up_simple_values(glob: &mut PlannerGlobal, parse: &mut Query, varno: usize) -> JoinTreeNode {
    let RteKind::Plan(Plan::Values(rows)) = &parse.rte(varno).kind else { unreachable!("the entry is VALUES") };
    let targetlist = rows[0].clone();
    let mut rvcontext = PullupReplaceVars {
        rv_cache: vec![None; targetlist.len()],
        targetlist,
        varno,
        wrap_option: ReplaceWrap::None,
        done: HashSet::new(),
    };
    perform_pullup_replace_vars(glob, parse, &mut rvcontext);
    parse.rtable[varno - 1].kind = RteKind::Result;
    JoinTreeNode::Rel(varno)
}

/// is_simple_values reports whether a VALUES list can be pulled up, as Postgres' function of the same name decides:
/// it has one row of expressions that return no sets, run no volatile functions, and read no enclosing row, and it is
/// the query's only relation.
fn is_simple_values(glob: &PlannerGlobal, parse: &Query, varno: usize) -> bool {
    let RteKind::Plan(Plan::Values(rows)) = &parse.rte(varno).kind else { return false };
    let reads_outer = |e: &Expr| {
        let mut outer = false;
        e.visit(&mut |x| outer |= matches!(x, Expr::Outer(..) | Expr::SetRef(_)));
        outer
    };
    rows.len() == 1
        && !rows[0].iter().any(|e| contain_volatile_functions(glob, e) || contain_subplans(e) || reads_outer(e))
        && parse.rtable.len() == 1
}

/// pull_up_constant_function replaces a relation of a function call that is a constant of one column by the
/// constant, as Postgres' function of the same name does.
fn pull_up_constant_function(glob: &mut PlannerGlobal, parse: &mut Query, varno: usize) -> JoinTreeNode {
    let RteKind::Plan(Plan::Function { call: call @ Expr::Const(_), ordinality: false, width: 1, defined: None }) =
        &parse.rte(varno).kind
    else {
        return JoinTreeNode::Rel(varno);
    };
    let targetlist = vec![call.clone()];
    let mut rvcontext = PullupReplaceVars {
        rv_cache: vec![None],
        targetlist,
        varno,
        wrap_option: ReplaceWrap::None,
        done: HashSet::new(),
    };
    perform_pullup_replace_vars(glob, parse, &mut rvcontext);
    parse.rtable[varno - 1].kind = RteKind::Result;
    JoinTreeNode::Rel(varno)
}

/// ReplaceWrap is when pullup_replace_vars wraps an expression in a PlaceHolderVar, as Postgres' ReplaceWrapOption
/// is: only where an outer join can make it NULL, or also where it reads no Var, as a FULL JOIN's condition needs.
#[derive(Clone, Copy, PartialEq)]
enum ReplaceWrap {
    None,
    VarFree,
}

/// PullupReplaceVars is the state of replacing the Vars of a pulled-up relation by its expressions, as Postgres'
/// pullup_replace_vars_context is, with the PlaceHolderVars whose expressions were already rewritten.
struct PullupReplaceVars {
    targetlist: Vec<Expr>,
    varno: usize,
    wrap_option: ReplaceWrap,
    rv_cache: Vec<Option<Expr>>,
    done: HashSet<usize>,
}

/// perform_pullup_replace_vars replaces the Vars of a pulled-up relation in the query's target list and join tree,
/// as Postgres' function of the same name does.
fn perform_pullup_replace_vars(glob: &mut PlannerGlobal, parse: &mut Query, rvcontext: &mut PullupReplaceVars) {
    for e in parse.upper_exprs_mut() {
        *e = pullup_replace_vars(glob, e.clone(), rvcontext);
    }
    let mut node = JoinTreeNode::From(Box::new(std::mem::replace(
        &mut parse.jointree,
        FromExpr { fromlist: Vec::new(), quals: Vec::new() },
    )));
    replace_vars_in_jointree(glob, &mut node, rvcontext);
    let JoinTreeNode::From(jointree) = node else { unreachable!("the join tree stays a FROM list") };
    parse.jointree = *jointree;
}

/// replace_vars_in_jointree replaces the Vars of a pulled-up relation in a join tree's quals, as Postgres' function
/// of the same name does.
fn replace_vars_in_jointree(glob: &mut PlannerGlobal, node: &mut JoinTreeNode, context: &mut PullupReplaceVars) {
    match node {
        JoinTreeNode::Rel(_) => {}
        JoinTreeNode::From(from) => {
            from.fromlist.iter_mut().for_each(|n| replace_vars_in_jointree(glob, n, context));
            from.quals =
                std::mem::take(&mut from.quals).into_iter().map(|q| pullup_replace_vars(glob, q, context)).collect();
        }
        JoinTreeNode::Join(join) => {
            replace_vars_in_jointree(glob, &mut join.larg, context);
            replace_vars_in_jointree(glob, &mut join.rarg, context);
            let save_wrap_option = context.wrap_option;
            if join.jointype == JoinType::Full {
                context.wrap_option = ReplaceWrap::VarFree;
            }
            join.quals =
                std::mem::take(&mut join.quals).into_iter().map(|q| pullup_replace_vars(glob, q, context)).collect();
            context.wrap_option = save_wrap_option;
        }
    }
}

/// pullup_replace_vars replaces the Vars of a pulled-up relation in an expression, and in the expressions of its
/// PlaceHolderVars, as Postgres' function of the same name does with replace_rte_variables.
fn pullup_replace_vars(glob: &mut PlannerGlobal, e: Expr, context: &mut PullupReplaceVars) -> Expr {
    replace_columns(e, &mut |id| match glob.node(id).clone() {
        VarNode::Var(var) if var.varno == context.varno => pullup_replace_vars_callback(glob, var, context),
        VarNode::PlaceHolderVar(phv) => {
            if context.done.insert(phv.phid) {
                let phexpr = glob.placeholder(phv.phid).phexpr.clone();
                glob.placeholders[phv.phid - 1].phexpr = pullup_replace_vars(glob, phexpr, context);
            }
            Expr::Column(id)
        }
        VarNode::Var(_) => Expr::Column(id),
    })
}

/// pullup_replace_vars_callback returns the expression that replaces a Var of a pulled-up relation: its expression,
/// wrapped in a PlaceHolderVar where an outer join can make the Var NULL but not the expression, with the Var's
/// nulling relations added, as Postgres' function of the same name does.
fn pullup_replace_vars_callback(glob: &mut PlannerGlobal, var: Var, rcon: &mut PullupReplaceVars) -> Expr {
    let varattno = var.varattno;
    let need_phv = !var.varnullingrels.is_empty() || rcon.wrap_option != ReplaceWrap::None;
    let newnode = match &rcon.rv_cache[varattno] {
        Some(cached) if need_phv => cached.clone(),
        _ => {
            let mut newnode = rcon.targetlist[varattno].clone();
            if need_phv {
                let wrap = match &newnode {
                    Expr::Column(_) => false,
                    other => !(contain_var_clause(other) && !contain_nonstrict_functions(glob, other)),
                };
                if wrap {
                    newnode = glob.make_placeholder_expr(newnode, Relids::singleton(rcon.varno));
                    rcon.rv_cache[varattno] = Some(newnode.clone());
                }
            }
            newnode
        }
    };
    if var.varnullingrels.is_empty() {
        return newnode;
    }
    match newnode {
        Expr::Column(id) => match glob.node(id).clone() {
            VarNode::Var(newvar) => {
                let varnullingrels = newvar.varnullingrels.union(&var.varnullingrels);
                glob.intern(VarNode::Var(Var { varnullingrels, ..newvar }))
            }
            VarNode::PlaceHolderVar(newphv) => {
                let phnullingrels = newphv.phnullingrels.union(&var.varnullingrels);
                glob.intern(VarNode::PlaceHolderVar(PlaceHolderVar { phnullingrels, ..newphv }))
            }
        },
        other => add_nulling_relids(glob, other, None, &var.varnullingrels),
    }
}

/// ReduceState is what the first pass of reduce_outer_joins learns about a join tree node: the relations under it,
/// whether an outer join is under it, and the same of its children, as Postgres' reduce_outer_joins_pass1_state is.
struct ReduceState {
    relids: Relids,
    contains_outer: bool,
    sub_states: Vec<ReduceState>,
}

/// ReduceResult is what the second pass of reduce_outer_joins reduced, as Postgres' reduce_outer_joins_pass2_state
/// is: the outer joins that became inner joins, each FULL JOIN that became a left join with the relations of the side
/// it still makes NULL, and the left joins that became anti joins.
#[derive(Default)]
struct ReduceResult {
    inner_reduced: Relids,
    partial_reduced: Vec<(usize, Relids)>,
    anti_reduced: Relids,
}

/// reduce_outer_joins turns outer joins into inner joins where a stricter condition above rejects the rows that
/// they pad with NULLs, and left joins into anti joins where a condition above keeps only those rows, removing the
/// reduced joins from the nulling relations of the query's Vars, as Postgres' reduce_outer_joins does.
pub fn reduce_outer_joins(glob: &mut PlannerGlobal, parse: &mut Query) {
    let state1 = reduce_outer_joins_pass1(&JoinTreeNode::From(Box::new(parse.jointree.clone())));
    if !state1.contains_outer {
        return;
    }
    let mut state2 = ReduceResult::default();
    let mut jointree = parse.jointree.clone();
    reduce_outer_joins_pass2_from(glob, parse, &mut jointree, &state1, &mut state2, &Relids::new(), &[]);
    parse.jointree = jointree;
    if !state2.inner_reduced.is_empty() {
        remove_nulling_relids_query(glob, parse, &state2.inner_reduced, &Relids::new());
    }
    for (full_join_rti, unreduced_side) in &state2.partial_reduced {
        remove_nulling_relids_query(glob, parse, &Relids::singleton(*full_join_rti), unreduced_side);
    }
    if !state2.anti_reduced.is_empty() {
        let antijoins = state2.anti_reduced.clone();
        let mut node = JoinTreeNode::From(Box::new(parse.jointree.clone()));
        remove_redundant_nullability_quals(glob, &mut node, &antijoins);
        let JoinTreeNode::From(jointree) = node else { unreachable!("the join tree stays a FROM list") };
        parse.jointree = *jointree;
    }
}

/// remove_redundant_nullability_quals removes the IS NULL tests of Vars that anti joins already make NULL from a
/// join tree's quals, as Postgres' function of the same name does.
fn remove_redundant_nullability_quals(glob: &PlannerGlobal, node: &mut JoinTreeNode, antijoins: &Relids) {
    let strip = |quals: &mut Vec<Expr>| {
        quals.retain(|clause| {
            !find_forced_null_var(glob, clause).is_some_and(|var| match glob.node(var) {
                VarNode::Var(var) => var.varnullingrels.overlap(antijoins),
                VarNode::PlaceHolderVar(_) => false,
            })
        })
    };
    match node {
        JoinTreeNode::Rel(_) => {}
        JoinTreeNode::From(from) => {
            from.fromlist.iter_mut().for_each(|n| remove_redundant_nullability_quals(glob, n, antijoins));
            strip(&mut from.quals);
        }
        JoinTreeNode::Join(join) => {
            remove_redundant_nullability_quals(glob, &mut join.larg, antijoins);
            remove_redundant_nullability_quals(glob, &mut join.rarg, antijoins);
            strip(&mut join.quals);
        }
    }
}

/// reduce_outer_joins_pass1 finds the relations and outer joins under a join tree node.
fn reduce_outer_joins_pass1(node: &JoinTreeNode) -> ReduceState {
    match node {
        JoinTreeNode::Rel(varno) => {
            ReduceState { relids: Relids::singleton(*varno), contains_outer: false, sub_states: Vec::new() }
        }
        JoinTreeNode::From(f) => {
            let sub_states: Vec<ReduceState> = f.fromlist.iter().map(reduce_outer_joins_pass1).collect();
            ReduceState {
                relids: sub_states.iter().fold(Relids::new(), |relids, s| relids.union(&s.relids)),
                contains_outer: sub_states.iter().any(|s| s.contains_outer),
                sub_states,
            }
        }
        JoinTreeNode::Join(j) => {
            let sub_states = vec![reduce_outer_joins_pass1(&j.larg), reduce_outer_joins_pass1(&j.rarg)];
            ReduceState {
                relids: sub_states[0].relids.union(&sub_states[1].relids),
                contains_outer: j.jointype.is_outer() || sub_states.iter().any(|s| s.contains_outer),
                sub_states,
            }
        }
    }
}

/// union_vars returns the Vars of two lists, as Postgres' mbms_add_members does.
fn union_vars(mut a: Vec<(usize, usize)>, b: &[(usize, usize)]) -> Vec<(usize, usize)> {
    for v in b {
        if !a.contains(v) {
            a.push(*v);
        }
    }
    a
}

/// reduce_outer_joins_pass2_from is reduce_outer_joins_pass2 for a FROM list, given the relations that conditions
/// above make non-nullable and the Vars they force to be NULL.
fn reduce_outer_joins_pass2_from(
    glob: &PlannerGlobal,
    parse: &mut Query,
    f: &mut FromExpr,
    state1: &ReduceState,
    state2: &mut ReduceResult,
    nonnullable_rels: &Relids,
    forced_null_vars: &[(usize, usize)],
) {
    let pass_nonnullable_rels =
        f.quals.iter().fold(nonnullable_rels.clone(), |relids, q| relids.union(&find_nonnullable_rels(glob, q)));
    let pass_forced_null_vars =
        f.quals.iter().fold(forced_null_vars.to_vec(), |vars, q| union_vars(vars, &find_forced_null_vars(glob, q)));
    for (node, sub_state) in f.fromlist.iter_mut().zip(&state1.sub_states) {
        if sub_state.contains_outer {
            reduce_outer_joins_pass2(
                glob,
                parse,
                node,
                sub_state,
                state2,
                &pass_nonnullable_rels,
                &pass_forced_null_vars,
            );
        }
    }
}

/// reduce_outer_joins_pass2 reduces the outer joins under a join tree node, as Postgres' function of the same name
/// does.
fn reduce_outer_joins_pass2(
    glob: &PlannerGlobal,
    parse: &mut Query,
    node: &mut JoinTreeNode,
    state1: &ReduceState,
    state2: &mut ReduceResult,
    nonnullable_rels: &Relids,
    forced_null_vars: &[(usize, usize)],
) {
    let j = match node {
        JoinTreeNode::From(f) => {
            return reduce_outer_joins_pass2_from(glob, parse, f, state1, state2, nonnullable_rels, forced_null_vars);
        }
        JoinTreeNode::Rel(_) => unreachable!("the second pass reaches only nodes with outer joins"),
        JoinTreeNode::Join(j) => j,
    };
    let rtindex = j.rtindex;
    let (mut left_state, mut right_state) = (&state1.sub_states[0], &state1.sub_states[1]);
    let mut jointype = j.jointype;
    match jointype {
        JoinType::Left if nonnullable_rels.overlap(&right_state.relids) => jointype = JoinType::Inner,
        JoinType::Right if nonnullable_rels.overlap(&left_state.relids) => jointype = JoinType::Inner,
        JoinType::Full => {
            match (nonnullable_rels.overlap(&left_state.relids), nonnullable_rels.overlap(&right_state.relids)) {
                (true, true) => jointype = JoinType::Inner,
                (true, false) => {
                    jointype = JoinType::Left;
                    state2.partial_reduced.push((rtindex, right_state.relids.clone()));
                }
                (false, true) => {
                    jointype = JoinType::Right;
                    state2.partial_reduced.push((rtindex, left_state.relids.clone()));
                }
                (false, false) => {}
            }
        }
        _ => {}
    }
    if jointype == JoinType::Right {
        std::mem::swap(&mut j.larg, &mut j.rarg);
        std::mem::swap(&mut left_state, &mut right_state);
        jointype = JoinType::Left;
    }
    if jointype == JoinType::Left {
        let nonnullable_vars =
            j.quals.iter().fold(Vec::new(), |vars, q| union_vars(vars, &find_nonnullable_vars(glob, q)));
        let overlap = nonnullable_vars.iter().filter(|v| forced_null_vars.contains(v)).map(|v| v.0);
        if overlap.collect::<Relids>().overlap(&right_state.relids) {
            jointype = JoinType::Anti;
        }
    }
    if rtindex != 0 && jointype != j.jointype {
        parse.rtable[rtindex - 1].kind = RteKind::Join(jointype);
        match jointype {
            JoinType::Inner => state2.inner_reduced.add_member(rtindex),
            JoinType::Anti => state2.anti_reduced.add_member(rtindex),
            _ => {}
        }
    }
    j.jointype = jointype;
    if !left_state.contains_outer && !right_state.contains_outer {
        return;
    }
    let (local_nonnullable_rels, local_forced_null_vars) = match jointype {
        JoinType::Full => (Relids::new(), Vec::new()),
        _ => {
            let mut rels =
                j.quals.iter().fold(Relids::new(), |relids, q| relids.union(&find_nonnullable_rels(glob, q)));
            let mut vars = j.quals.iter().fold(Vec::new(), |vars, q| union_vars(vars, &find_forced_null_vars(glob, q)));
            if matches!(jointype, JoinType::Inner | JoinType::Semi) {
                rels.add_members(nonnullable_rels);
                vars = union_vars(vars, forced_null_vars);
            }
            (rels, vars)
        }
    };
    if left_state.contains_outer {
        let (rels, vars) = match jointype {
            JoinType::Inner | JoinType::Semi => (local_nonnullable_rels.clone(), local_forced_null_vars.clone()),
            JoinType::Full => (Relids::new(), Vec::new()),
            _ => (nonnullable_rels.clone(), forced_null_vars.to_vec()),
        };
        reduce_outer_joins_pass2(glob, parse, &mut j.larg, left_state, state2, &rels, &vars);
    }
    if right_state.contains_outer {
        let (rels, vars) = match jointype {
            JoinType::Full => (Relids::new(), Vec::new()),
            _ => (local_nonnullable_rels, local_forced_null_vars),
        };
        reduce_outer_joins_pass2(glob, parse, &mut j.rarg, right_state, state2, &rels, &vars);
    }
}

/// remove_useless_result_rtes removes the RESULT relations of the join tree that join nothing, which pulled-up
/// subqueries and VALUES lists leave, along with the left joins to them, as Postgres' function of the same name does.
pub fn remove_useless_result_rtes(glob: &mut PlannerGlobal, parse: &mut Query) {
    let baserels = match glob.last_ph_id() {
        0 => Relids::new(),
        _ => get_relids_in_jointree(&JoinTreeNode::From(Box::new(parse.jointree.clone())), false, false),
    };
    let mut dropped_outer_joins = Relids::new();
    let mut node = JoinTreeNode::From(Box::new(parse.jointree.clone()));
    remove_useless_results_recurse(glob, parse, &mut node, &baserels, None, &mut dropped_outer_joins, true);
    let JoinTreeNode::From(jointree) = node else { unreachable!("the join tree stays a FROM list") };
    parse.jointree = *jointree;
    if !dropped_outer_joins.is_empty() {
        remove_nulling_relids_query(glob, parse, &dropped_outer_joins, &Relids::new());
    }
}

/// remove_useless_results_recurse is remove_useless_result_rtes for a join tree node, given the quals of the node
/// above that its quals may move into, as Postgres' function of the same name does.
fn remove_useless_results_recurse(
    glob: &mut PlannerGlobal,
    parse: &mut Query,
    node: &mut JoinTreeNode,
    baserels: &Relids,
    parent_quals: Option<&mut Vec<Expr>>,
    dropped_outer_joins: &mut Relids,
    top: bool,
) {
    match node {
        JoinTreeNode::Rel(_) => {}
        JoinTreeNode::From(f) => {
            let mut result_relids = Relids::new();
            let mut i = 0;
            while i < f.fromlist.len() {
                let mut child = std::mem::replace(&mut f.fromlist[i], JoinTreeNode::Rel(0));
                remove_useless_results_recurse(
                    glob,
                    parse,
                    &mut child,
                    baserels,
                    Some(&mut f.quals),
                    dropped_outer_joins,
                    false,
                );
                f.fromlist[i] = child;
                let varno = get_result_relid(parse, &f.fromlist[i]);
                if f.fromlist.len() > 1
                    && varno != 0
                    && !find_dependent_phvs_in_jointree(glob, &JoinTreeNode::From(f.clone()), varno, baserels)
                {
                    f.fromlist.remove(i);
                    result_relids.add_member(varno);
                } else {
                    i += 1;
                }
            }
            if !result_relids.is_empty() {
                let subrelids = get_relids_in_jointree(&JoinTreeNode::From(f.clone()), true, false);
                for varno in result_relids.members() {
                    remove_result_refs(glob, parse, varno, &subrelids);
                }
            }
            if f.fromlist.len() == 1 && !top && (f.quals.is_empty() || parent_quals.is_some()) {
                if let Some(parent_quals) = parent_quals {
                    let quals = std::mem::take(&mut f.quals);
                    parent_quals.splice(0..0, quals);
                }
                *node = f.fromlist.pop().expect("one member");
            }
        }
        JoinTreeNode::Join(j) => {
            let jointype = j.jointype;
            let mut larg = std::mem::replace(&mut j.larg, JoinTreeNode::Rel(0));
            let mut rarg = std::mem::replace(&mut j.rarg, JoinTreeNode::Rel(0));
            let mut parent_quals = parent_quals;
            match jointype {
                JoinType::Inner => {
                    remove_useless_results_recurse(
                        glob,
                        parse,
                        &mut larg,
                        baserels,
                        Some(&mut j.quals),
                        dropped_outer_joins,
                        false,
                    );
                }
                JoinType::Left => {
                    remove_useless_results_recurse(
                        glob,
                        parse,
                        &mut larg,
                        baserels,
                        parent_quals.as_deref_mut(),
                        dropped_outer_joins,
                        false,
                    );
                }
                _ => remove_useless_results_recurse(glob, parse, &mut larg, baserels, None, dropped_outer_joins, false),
            }
            match jointype {
                JoinType::Inner | JoinType::Left => remove_useless_results_recurse(
                    glob,
                    parse,
                    &mut rarg,
                    baserels,
                    Some(&mut j.quals),
                    dropped_outer_joins,
                    false,
                ),
                _ => remove_useless_results_recurse(glob, parse, &mut rarg, baserels, None, dropped_outer_joins, false),
            }
            j.larg = larg;
            j.rarg = rarg;
            let replace_with =
                |keep: JoinTreeNode, quals: Vec<Expr>, parent_quals: Option<&mut Vec<Expr>>| match parent_quals {
                    None if !quals.is_empty() => JoinTreeNode::From(Box::new(FromExpr { fromlist: vec![keep], quals })),
                    Some(parent_quals) => {
                        parent_quals.splice(0..0, quals);
                        keep
                    }
                    None => keep,
                };
            match jointype {
                JoinType::Inner => {
                    let varno = get_result_relid(parse, &j.larg);
                    if varno != 0 && !find_dependent_phvs_in_jointree(glob, &j.rarg, varno, baserels) {
                        let subrelids = get_relids_in_jointree(&j.rarg, true, false);
                        remove_result_refs(glob, parse, varno, &subrelids);
                        let rarg = std::mem::replace(&mut j.rarg, JoinTreeNode::Rel(0));
                        *node = replace_with(rarg, std::mem::take(&mut j.quals), parent_quals);
                        return;
                    }
                    let varno = get_result_relid(parse, &j.rarg);
                    if varno != 0 {
                        let subrelids = get_relids_in_jointree(&j.larg, true, false);
                        remove_result_refs(glob, parse, varno, &subrelids);
                        let larg = std::mem::replace(&mut j.larg, JoinTreeNode::Rel(0));
                        *node = replace_with(larg, std::mem::take(&mut j.quals), parent_quals);
                    }
                }
                JoinType::Left => {
                    let varno = get_result_relid(parse, &j.rarg);
                    if varno != 0 && (j.quals.is_empty() || !find_dependent_phvs(glob, parse, varno, baserels)) {
                        let subrelids = get_relids_in_jointree(&j.larg, true, false);
                        remove_result_refs(glob, parse, varno, &subrelids);
                        dropped_outer_joins.add_member(j.rtindex);
                        *node = std::mem::replace(&mut j.larg, JoinTreeNode::Rel(0));
                    }
                }
                JoinType::Semi => {
                    let varno = get_result_relid(parse, &j.rarg);
                    if varno != 0 {
                        let subrelids = get_relids_in_jointree(&j.larg, true, false);
                        remove_result_refs(glob, parse, varno, &subrelids);
                        let larg = std::mem::replace(&mut j.larg, JoinTreeNode::Rel(0));
                        *node = replace_with(larg, std::mem::take(&mut j.quals), parent_quals);
                    }
                }
                _ => {}
            }
        }
    }
}

/// get_result_relid returns the range table index of a join tree node that is a RESULT relation, or 0, as Postgres'
/// function of the same name does.
fn get_result_relid(parse: &Query, node: &JoinTreeNode) -> usize {
    match node {
        JoinTreeNode::Rel(varno) if matches!(parse.rte(*varno).kind, RteKind::Result) => *varno,
        _ => 0,
    }
}

/// remove_result_refs replaces a removed RESULT relation in the relations of the query's PlaceHolderVars by the
/// relations of the join tree node that took its place, as Postgres' function of the same name does.
fn remove_result_refs(glob: &mut PlannerGlobal, parse: &Query, varno: usize, subrelids: &Relids) {
    if glob.last_ph_id() != 0 {
        substitute_phv_relids(glob, parse, varno, subrelids);
    }
}

/// query_phids returns the IDs of the PlaceHolderVars that a query's expressions hold, with those that their
/// expressions hold.
fn query_phids(glob: &PlannerGlobal, parse: &Query) -> Vec<usize> {
    let mut exprs: Vec<Expr> = parse.upper_exprs();
    collect_quals(&JoinTreeNode::From(Box::new(parse.jointree.clone())), &mut exprs);
    exprs_phids(glob, &exprs)
}

/// collect_quals adds the quals of a join tree node to a list.
fn collect_quals(node: &JoinTreeNode, out: &mut Vec<Expr>) {
    match node {
        JoinTreeNode::Rel(_) => {}
        JoinTreeNode::From(f) => {
            f.fromlist.iter().for_each(|n| collect_quals(n, out));
            out.extend(f.quals.iter().cloned());
        }
        JoinTreeNode::Join(j) => {
            collect_quals(&j.larg, out);
            collect_quals(&j.rarg, out);
            out.extend(j.quals.iter().cloned());
        }
    }
}

/// exprs_phids returns the IDs of the PlaceHolderVars that expressions hold, with those that their expressions hold.
fn exprs_phids(glob: &PlannerGlobal, exprs: &[Expr]) -> Vec<usize> {
    let mut phids: Vec<usize> = Vec::new();
    let mut pending: Vec<Expr> = exprs.to_vec();
    while let Some(e) = pending.pop() {
        super::var::visit_columns(&e, &mut |id| {
            if let VarNode::PlaceHolderVar(phv) = glob.node(id)
                && !phids.contains(&phv.phid)
            {
                phids.push(phv.phid);
                pending.push(glob.placeholder(phv.phid).phexpr.clone());
            }
        });
    }
    phids
}

/// find_dependent_phvs reports whether a PlaceHolderVar of the query is evaluated at a relation alone among the base
/// relations, as Postgres' function of the same name does.
fn find_dependent_phvs(glob: &PlannerGlobal, parse: &Query, varno: usize, baserels: &Relids) -> bool {
    glob.last_ph_id() != 0
        && query_phids(glob, parse)
            .into_iter()
            .any(|phid| glob.placeholder(phid).phrels.intersect(baserels) == Relids::singleton(varno))
}

/// find_dependent_phvs_in_jointree is find_dependent_phvs for the quals of a join tree node.
fn find_dependent_phvs_in_jointree(glob: &PlannerGlobal, node: &JoinTreeNode, varno: usize, baserels: &Relids) -> bool {
    if glob.last_ph_id() == 0 {
        return false;
    }
    let mut quals = Vec::new();
    collect_quals(node, &mut quals);
    exprs_phids(glob, &quals)
        .into_iter()
        .any(|phid| glob.placeholder(phid).phrels.intersect(baserels) == Relids::singleton(varno))
}

/// substitute_phv_relids replaces a relation in the relations of the query's PlaceHolderVars by others, as Postgres'
/// function of the same name does.
fn substitute_phv_relids(glob: &mut PlannerGlobal, parse: &Query, varno: usize, subrelids: &Relids) {
    for phid in query_phids(glob, parse) {
        let placeholder = &mut glob.placeholders[phid - 1];
        if placeholder.phrels.is_member(varno) {
            placeholder.phrels = placeholder.phrels.union(subrelids).without_member(varno);
        }
    }
}

/// get_relids_in_jointree returns the base relations of a join tree node, with the outer or inner joins under it
/// when asked, as Postgres' function of the same name does.
pub fn get_relids_in_jointree(node: &JoinTreeNode, include_outer_joins: bool, include_inner_joins: bool) -> Relids {
    match node {
        JoinTreeNode::Rel(varno) => Relids::singleton(*varno),
        JoinTreeNode::From(f) => f.fromlist.iter().fold(Relids::new(), |relids, n| {
            relids.union(&get_relids_in_jointree(n, include_outer_joins, include_inner_joins))
        }),
        JoinTreeNode::Join(j) => {
            let mut result = get_relids_in_jointree(&j.larg, include_outer_joins, include_inner_joins)
                .union(&get_relids_in_jointree(&j.rarg, include_outer_joins, include_inner_joins));
            let include = match j.jointype {
                JoinType::Inner => include_inner_joins,
                _ => include_outer_joins,
            };
            if j.rtindex != 0 && include {
                result.add_member(j.rtindex);
            }
            result
        }
    }
}

/// has_outer_joins reports whether the query's join tree holds an outer join, as Postgres' hasOuterJoins records.
pub fn has_outer_joins(parse: &Query) -> bool {
    parse.rtable.iter().any(|rte| matches!(rte.kind, RteKind::Join(jointype) if jointype.is_outer()))
}
