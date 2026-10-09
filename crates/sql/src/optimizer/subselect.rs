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

//! The sublink conversions of Postgres' optimizer/plan/subselect.c, with pull_up_sublinks from prepjointree.c:
//! turning `EXISTS`, `NOT EXISTS`, and `IN` subqueries of the WHERE clause into semi and anti joins. A subquery's
//! columns are its own row's, and the enclosing query's row is one level out, as `Expr::Outer` reads it.

use super::clauses::is_volatile_node;
use super::nodes::{JoinExpr, JoinTreeNode, JoinType, PlannerGlobal, RangeTblEntry};
use crate::expr::Expr;
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

/// plan_sublink plans the subquery of an `EXISTS` or `IN` subquery of the WHERE clause that its binding left
/// unplanned but that pull_up_sublinks cannot turn into a join, as it runs for each enclosing row.
pub(crate) fn plan_sublink(ctx: &mut Ctx<'_>, e: Expr, uncorrelated: bool) -> Expr {
    let mut plan = |p: Box<Plan>| Box::new(crate::plan::share_subquery(super::planner(ctx, *p), uncorrelated));
    match e {
        Expr::Exists(p) => Expr::Exists(plan(p)),
        Expr::AnySubquery(test, p, all) => Expr::AnySubquery(test, plan(p), all),
        other => other,
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
