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

//! Postgres' optimizer/plan/subselect.c, with pull_up_sublinks from prepjointree.c: turning `EXISTS`, `NOT
//! EXISTS`, and `IN` subqueries of the WHERE clause into semi and anti joins, and every other subquery expression
//! into a SubPlan. A subquery's columns are its own row's, and the enclosing query's row is one level out, as
//! `Expr::Outer` reads it. A SubPlan's subquery reads the values of its arguments as that row, as Postgres' Params.

use std::collections::BTreeSet;
use std::rc::Rc;

use super::clauses::is_volatile_node;
use super::costsize::{Enables, HASH_MEM, TUPLE_HEADER, cost_material, cost_subplan, maxalign};
use super::nodes::{FromExpr, JoinExpr, JoinTreeNode, JoinType, Path, PlannerGlobal, Query, RangeTblEntry};
use crate::expr::{CmpOp, Expr, SubPlan};
use crate::plan::Plan;
use crate::query::Ctx;

/// sublink_convertible reports whether pull_up_sublinks turns an `EXISTS` or `IN` subquery of the WHERE clause into
/// a join.
pub(crate) fn sublink_convertible(e: &Expr) -> bool {
    match e {
        Expr::Exists(plan) => exists_parts(plan).is_some(),
        Expr::AnySubquery(test, plan, false) => any_convertible(test, plan),
        _ => false,
    }
}

/// make_subplan returns the SubPlan of a subquery expression over a row of `width` columns, whose subquery is the
/// binder's plan, planned already or not, as Postgres' function of the same name does: the subquery's reads of the
/// enclosing row become reads of the SubPlan's arguments, and an unplanned subquery is planned for the share of its
/// rows that the expression reads.
pub(crate) fn make_subplan(ctx: &mut Ctx<'_>, mut link: Expr, planned: bool, width: usize) -> Expr {
    let plan = link.subquery_mut().expect("a subquery expression");
    let args = replace_correlation_vars(plan, width);
    let plan = std::mem::replace(plan, Plan::OneRow);
    let tuple_fraction = match &link {
        Expr::Exists(_) => 1.0,
        Expr::AnySubquery(..) => 0.5,
        _ => 0.0,
    };
    let (plan, path) = match planned {
        true => (plan, None),
        false => {
            let (plan, path) = super::plan_subselect(ctx, plan, matches!(link, Expr::Exists(_)), tuple_fraction);
            (plan, Some(path))
        }
    };
    build_subplan(ctx, link, plan, path, args)
}

/// build_subplan makes the SubPlan of a subquery expression from its subquery's plan, the path the planner chose for
/// it, and its arguments, as Postgres' function of the same name does: a subquery that reads no column of the
/// enclosing row is an initplan unless it is an ANY or ALL test, an uncorrelated IN test hashes the subquery's rows
/// when they fit in memory, and the rows of a subquery that reads no enclosing row are kept for every call.
fn build_subplan(ctx: &Ctx<'_>, mut link: Expr, plan: Plan, mut path: Option<Rc<Path>>, args: Vec<Expr>) -> Expr {
    let init_plan = args.is_empty() && !matches!(link, Expr::AnySubquery(..));
    let use_hash_table = match (&link, &path) {
        (Expr::AnySubquery(test, _, false), Some(path)) => {
            args.is_empty() && subpath_is_hashable(path) && testexpr_is_hashable(test)
        }
        _ => false,
    };
    let enables = Enables::read(&ctx.session.settings);
    let mut materializes = exec_materializes_output(&plan);
    if args.is_empty() && !init_plan && !use_hash_table && enables.material && !materializes {
        if let Some(input) = &mut path {
            let (disabled_nodes, startup_cost, total_cost) = cost_material(&enables, input);
            *input = Rc::new(Path { disabled_nodes, startup_cost, total_cost, ..(**input).clone() });
        }
        materializes = true;
    }
    let uncorrelated = args.is_empty() && crate::joins::plan_lowest_level(&plan).is_some_and(|level| level >= 0);
    *link.subquery_mut().expect("a subquery expression") = crate::plan::share_subquery(plan, uncorrelated);
    let mut subplan = SubPlan { link, args, init_plan, startup_cost: 0.0, per_call_cost: 0.0 };
    if let Some(path) = &path {
        cost_subplan(&mut subplan, path, use_hash_table, materializes);
    }
    Expr::SubPlan(Box::new(subplan))
}

/// exec_materializes_output reports whether a plan's top node keeps its rows for reading again, as Postgres'
/// ExecMaterializesOutput does.
fn exec_materializes_output(plan: &Plan) -> bool {
    matches!(
        plan,
        Plan::Sort { .. }
            | Plan::Function { .. }
            | Plan::RowsFrom { .. }
            | Plan::WorkTable(..)
            | Plan::XmlTable(_)
            | Plan::JsonTable(_)
            | Plan::Once(_)
    )
}

/// subpath_is_hashable reports whether the rows of an IN test's subquery fit in a hash table's memory, as Postgres'
/// function of the same name does.
fn subpath_is_hashable(path: &Path) -> bool {
    path.rows * (maxalign(path.width) + TUPLE_HEADER) <= HASH_MEM
}

/// testexpr_is_hashable reports whether an IN test is an equality, or an AND of equalities, that hashing can answer,
/// as Postgres' function of the same name does.
fn testexpr_is_hashable(testexpr: &Expr) -> bool {
    match testexpr {
        Expr::Compare(..) => test_opexpr_is_hashable(testexpr),
        Expr::And(..) => crate::indexscan::conjuncts(testexpr)
            .into_iter()
            .all(|c| matches!(c, Expr::Compare(..)) && test_opexpr_is_hashable(c)),
        _ => false,
    }
}

/// test_opexpr_is_hashable reports whether a comparison of an IN test is a hashable operator over the enclosing row
/// on its left and the subquery's value on its right, as Postgres' function of the same name does.
fn test_opexpr_is_hashable(testexpr: &Expr) -> bool {
    let Expr::Compare(op, left, right) = testexpr else { return false };
    let (mut left_reads_value, mut right_reads_row) = (false, false);
    left.visit(&mut |e| left_reads_value |= matches!(e, Expr::SubqueryValue));
    right.visit(&mut |e| right_reads_row |= matches!(e, Expr::Column(_)));
    hash_ok_operator(*op) && !left_reads_value && !right_reads_row
}

/// hash_ok_operator reports whether a comparison can hash its values, as Postgres' function of the same name does:
/// equality, which every Doltgres type can hash.
fn hash_ok_operator(op: CmpOp) -> bool {
    op == CmpOp::Eq
}

/// replace_correlation_vars rewrites a subquery's reads of the enclosing row, a row of `width` columns, into reads of
/// the parameters it returns, each a column of the enclosing row, as Postgres' function of the same name does with
/// replace_outer_var. A subquery holding a plan whose expressions it cannot see reads every column.
fn replace_correlation_vars(plan: &mut Plan, width: usize) -> Vec<Expr> {
    let mut read = BTreeSet::new();
    if !crate::indexscan::outer_reads(plan, 1, &mut read) {
        return (0..width).map(Expr::Column).collect();
    }
    let params: Vec<usize> = read.into_iter().collect();
    plan.map_exprs(0, &mut |e, depth| assign_params(e, depth, &params));
    params.into_iter().map(Expr::Column).collect()
}

/// assign_params rewrites an expression `nesting` subqueries deep within a subquery to read each column of the
/// enclosing row as the parameter at that column's position in `params`.
fn assign_params(e: Expr, nesting: usize, params: &[usize]) -> Expr {
    let mut e = match e {
        Expr::Outer(d, c) if d == nesting + 1 => {
            return Expr::Outer(d, params.binary_search(&c).expect("a parameter"));
        }
        other => other.map_children(&mut |x| assign_params(x, nesting, params)),
    };
    if let Some(plan) = e.subquery_mut() {
        plan.map_exprs(0, &mut |x, depth| assign_params(x, nesting + 1 + depth, params));
    }
    e
}

/// process_sublinks turns each subquery expression of an expression over a row of `width` columns into a SubPlan,
/// as Postgres' SS_process_sublinks does.
pub fn process_sublinks(ctx: &mut Ctx<'_>, e: Expr, width: usize) -> Expr {
    match e {
        link @ (Expr::Exists(_) | Expr::Scalar(_) | Expr::ArraySubquery(..) | Expr::AnySubquery(..)) => {
            let link = link.map_children(&mut |c| process_sublinks(ctx, c, width));
            make_subplan(ctx, link, false, width)
        }
        other => other.map_children(&mut |c| process_sublinks(ctx, c, width)),
    }
}

/// has_sublink reports whether an expression holds a subquery expression that is not a SubPlan yet.
pub fn has_sublink(e: &Expr) -> bool {
    let mut found = false;
    e.visit(&mut |x| {
        found |= matches!(x, Expr::Exists(_) | Expr::Scalar(_) | Expr::ArraySubquery(..) | Expr::AnySubquery(..))
    });
    found
}

/// simplify_exists_query drops what an EXISTS subquery's result does not depend on, its target list, grouping,
/// DISTINCT, ORDER BY, and a positive or NULL constant LIMIT, unless it has aggregates, grouping sets, windows,
/// set-returning functions, HAVING, or OFFSET, returning whether it did, as Postgres' simplify_EXISTS_query does.
pub fn simplify_exists_query(query: &mut Query) -> bool {
    if query.has_aggs()
        || query.grouping_sets.is_some()
        || !query.window_funcs.is_empty()
        || !query.target_srfs.is_empty()
        || query.having_qual.is_some()
        || query.limit_offset.is_some()
    {
        return false;
    }
    if let Some(limit) = &query.limit_count {
        if !positive_or_null(limit) {
            return false;
        }
        query.limit_count = None;
    }
    query.target_list.clear();
    query.group_clause.clear();
    query.distinct_clause.clear();
    query.sort_clause.clear();
    query.has_distinct_on = false;
    true
}

/// ss_charge_for_initplans adds the costs of a query's initplans, which run once for the query, to every path of its
/// final relation, as Postgres' SS_charge_for_initplans does.
pub fn ss_charge_for_initplans(root: &mut super::PlannerInfo<'_, '_>, final_rel: usize) {
    let mut initplan_cost = 0.0;
    let mut add = |e: &Expr| {
        e.visit(&mut |x| {
            if let Expr::SubPlan(subplan) = x
                && subplan.init_plan
            {
                initplan_cost += subplan.startup_cost + subplan.per_call_cost;
            }
        })
    };
    root.parse.upper_exprs_mut().into_iter().for_each(|e| add(e));
    root.parse.jointree.fromlist.iter().for_each(|node| jointree_quals(node, &mut add));
    root.parse.jointree.quals.iter().for_each(&mut add);
    if initplan_cost == 0.0 {
        return;
    }
    for path in &mut root.rels[final_rel].pathlist {
        let mut charged = (**path).clone();
        charged.startup_cost += initplan_cost;
        charged.total_cost += initplan_cost;
        *path = Rc::new(charged);
    }
}

/// jointree_quals calls a function with each qual of a join tree.
fn jointree_quals(node: &JoinTreeNode, f: &mut dyn FnMut(&Expr)) {
    match node {
        JoinTreeNode::Rel(_) => {}
        JoinTreeNode::From(from) => {
            let FromExpr { fromlist, quals } = &**from;
            fromlist.iter().for_each(|n| jointree_quals(n, f));
            quals.iter().for_each(&mut *f);
        }
        JoinTreeNode::Join(join) => {
            jointree_quals(&join.larg, f);
            jointree_quals(&join.rarg, f);
            join.quals.iter().for_each(&mut *f);
        }
    }
}

/// pull_up_sublinks turns each of the WHERE clause's conditions that pull_up_sublinks can into a semi or anti join of
/// the join tree with the subquery's rows, returning the new join tree and the other conditions, as Postgres'
/// function of the same name does for the WHERE clause's top-level conditions.
pub fn pull_up_sublinks(
    glob: &mut PlannerGlobal,
    ctx: &mut Ctx<'_>,
    mut jointree: JoinTreeNode,
    quals: Vec<Expr>,
    rtable: &mut Vec<RangeTblEntry>,
    output: &[Expr],
) -> (JoinTreeNode, Vec<Expr>) {
    let mut remaining = Vec::new();
    for qual in quals {
        let converted = match &qual {
            Expr::Exists(plan) => convert_exists_sublink_to_join(glob, ctx, plan, false, rtable, output),
            Expr::Not(inner) => match &**inner {
                Expr::Exists(plan) => convert_exists_sublink_to_join(glob, ctx, plan, true, rtable, output),
                _ => None,
            },
            Expr::AnySubquery(test, plan, false) => convert_any_sublink_to_join(glob, ctx, test, plan, rtable, output),
            _ => None,
        };
        match converted {
            Some((jointype, rarg, quals)) => {
                jointree = JoinTreeNode::Join(Box::new(JoinExpr { jointype, larg: jointree, rarg, quals, rtindex: 0 }))
            }
            None => remaining.push(qual),
        }
    }
    (jointree, remaining)
}

/// any_convertible reports whether convert_ANY_sublink_to_join turns an `IN` test of a subquery into a semi join:
/// the subquery of one column reads nothing of the enclosing rows, and the test reads the enclosing row and runs no
/// volatile function or subquery.
fn any_convertible(test: &Expr, plan: &Plan) -> bool {
    let plan = match plan {
        Plan::Once(inner) => inner,
        other => other,
    };
    let mut reads_row = false;
    test.visit(&mut |e| reads_row |= matches!(e, Expr::Column(_)));
    plan.width() == 1
        && reads_row
        && !crate::plan::has_subquery(test)
        && !volatile(test)
        && crate::joins::plan_lowest_level(plan).is_some_and(|level| level >= 0)
}

/// convert_any_sublink_to_join returns the semi join that an `IN` test of a subquery becomes, with the subquery as
/// its inner side, as Postgres' convert_ANY_sublink_to_join does.
fn convert_any_sublink_to_join(
    glob: &mut PlannerGlobal,
    ctx: &mut Ctx<'_>,
    test: &Expr,
    plan: &Plan,
    rtable: &mut Vec<RangeTblEntry>,
    output: &[Expr],
) -> Option<(JoinType, JoinTreeNode, Vec<Expr>)> {
    if !any_convertible(test, plan) {
        return None;
    }
    let plan = match plan {
        Plan::Once(inner) => inner,
        other => other,
    };
    let mut subquery_vars = Vec::new();
    let rarg = super::build_jointree(glob, ctx, plan.clone(), rtable, &mut subquery_vars);
    let quals = vec![convert_testexpr(test.clone(), &subquery_vars[0], output)];
    Some((JoinType::Semi, rarg, quals))
}

/// convert_testexpr rewrites an `IN` test over the enclosing row and the subquery's value into one over their
/// expressions.
fn convert_testexpr(e: Expr, value: &Expr, output: &[Expr]) -> Expr {
    match e {
        Expr::SubqueryValue => value.clone(),
        Expr::Column(c) => output[c].clone(),
        other => other.map_children(&mut |c| convert_testexpr(c, value, output)),
    }
}

/// exists_parts returns the rows below an `EXISTS` subquery's WHERE clause and that clause's conditions, when
/// convert_EXISTS_sublink_to_join can turn it into a join: the rows read nothing of the enclosing rows, and the
/// conditions read the enclosing row and run no volatile function or subquery, as Postgres' simplify_EXISTS_query
/// and convert_EXISTS_sublink_to_join require.
fn exists_parts(plan: &Plan) -> Option<(&Plan, Vec<&Expr>)> {
    let mut input = plan;
    loop {
        input = match input {
            Plan::Aggregate { aggregates, .. } if !aggregates.is_empty() => return None,
            Plan::Project { input, .. }
            | Plan::Once(input)
            | Plan::Sort { input, .. }
            | Plan::Distinct { input, .. }
            | Plan::Aggregate { input, sets: None, .. } => input,
            Plan::Limit { input, limit, offset: None } if limit.as_ref().is_none_or(positive_or_null) => input,
            _ => break,
        };
    }
    let mut where_clause = Vec::new();
    while let Plan::Filter { input: inner, predicate } = input {
        where_clause.extend(crate::indexscan::conjuncts(predicate));
        input = inner;
    }
    let mut reads_enclosing = false;
    for c in &where_clause {
        if crate::plan::has_subquery(c) || volatile(c) {
            return None;
        }
        c.visit(&mut |e| reads_enclosing |= matches!(e, Expr::Outer(1, _)));
    }
    let own = crate::joins::plan_lowest_level(input).is_some_and(|level| level >= 0);
    (reads_enclosing && own).then_some((input, where_clause))
}

/// positive_or_null reports whether a LIMIT is a positive constant or NULL, which an `EXISTS` subquery ignores.
fn positive_or_null(limit: &Expr) -> bool {
    matches!(limit, Expr::Const(crate::types::Value::Null))
        || matches!(limit, Expr::Const(crate::types::Value::Int8(n)) if *n > 0)
}

/// convert_exists_sublink_to_join returns the semi join, or anti join under NOT, that an `EXISTS` subquery becomes,
/// with the rows below its WHERE clause as its inner side and that clause as its condition, as Postgres'
/// convert_EXISTS_sublink_to_join does.
fn convert_exists_sublink_to_join(
    glob: &mut PlannerGlobal,
    ctx: &mut Ctx<'_>,
    plan: &Plan,
    under_not: bool,
    rtable: &mut Vec<RangeTblEntry>,
    output: &[Expr],
) -> Option<(JoinType, JoinTreeNode, Vec<Expr>)> {
    let (input, where_clause) = exists_parts(plan)?;
    let mut subquery_vars = Vec::new();
    let rarg = super::build_jointree(glob, ctx, input.clone(), rtable, &mut subquery_vars);
    let quals = where_clause.into_iter().map(|c| pull_up_level(c.clone(), &subquery_vars, output)).collect();
    Some((if under_not { JoinType::Anti } else { JoinType::Semi }, rarg, quals))
}

/// pull_up_level rewrites a condition of a subquery, over its own row and the enclosing ones, into one of the
/// enclosing query: its own columns become their expressions, the enclosing row's become that row's, and rows
/// further out come one level closer, as Postgres' IncrementVarSublevelsUp moves them.
fn pull_up_level(e: Expr, subquery_vars: &[Expr], output: &[Expr]) -> Expr {
    match e {
        Expr::Column(c) => subquery_vars[c].clone(),
        Expr::Outer(1, c) => output[c].clone(),
        Expr::Outer(depth, c) => Expr::Outer(depth - 1, c),
        other => other.map_children(&mut |c| pull_up_level(c, subquery_vars, output)),
    }
}

/// volatile reports whether an expression of a subquery, before its Vars are built, calls a volatile function.
fn volatile(e: &Expr) -> bool {
    let mut found = false;
    e.visit(&mut |x| found |= is_volatile_node(x));
    found
}
