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

//! Unbinding: turning the plan that the binder built for a SELECT back into the Query that Postgres' parse analysis
//! would have made of it. The binder stacks the query's upper clauses over its FROM and WHERE clauses in a fixed
//! order: an aggregation and its HAVING filter, the windows, the set-returning functions, the projection of the
//! select list and the hidden sort keys after it, DISTINCT and ORDER BY, LIMIT and OFFSET, and a projection of the
//! visible columns. Each layer's columns become expressions of the Query: Vars of the join tree, `Expr::AggRef`,
//! `Expr::WindowRef`, and `Expr::SetRef` of its aggregate calls, window calls, and set-returning functions.

use super::nodes::{FromExpr, JoinTreeNode, PlannerGlobal, Query, SortGroupClause, TargetEntry};
use crate::expr::Expr;
use crate::plan::{Plan, SortKey};
use crate::query::Ctx;

/// Aggregation is the group keys, aggregate calls, and grouping sets of a SELECT's aggregation.
type Aggregation = (Vec<Expr>, Vec<crate::functions::aggregate::AggCall>, Option<Vec<Vec<usize>>>);

/// Upper is the upper clauses of a SELECT's plan, peeled off the layers above its FROM and WHERE clauses, with what
/// each column of the topmost layer that it read holds, as an expression over the layer below it.
#[derive(Default)]
struct Upper {
    limit: Option<(Option<Expr>, Option<Expr>)>,
    sort: Vec<SortKey>,
    distinct: Option<Option<Vec<Expr>>>,
    visible: Option<usize>,
    target: Vec<Expr>,
    srfs: Vec<Vec<Expr>>,
    windows: Vec<Vec<crate::window::WindowCall>>,
    having: Option<Expr>,
    aggregation: Option<Aggregation>,
}

/// unbind returns the Query of a SELECT's plan, with the range table and join tree of its FROM and WHERE clauses
/// and its upper clauses.
pub fn unbind(glob: &mut PlannerGlobal, ctx: &mut Ctx<'_>, plan: Plan) -> Query {
    let (from, upper) = peel(plan);
    let (mut rtable, mut columns) = (Vec::new(), Vec::new());
    let jointree = match super::build_jointree(glob, ctx, from, &mut rtable, &mut columns) {
        JoinTreeNode::From(from) => *from,
        other => FromExpr { fromlist: vec![other], quals: Vec::new() },
    };
    let mut query = Query { rtable, jointree, ..Query::default() };
    let mut meaning = columns;
    let Some(upper) = upper else {
        query.target_list =
            meaning.into_iter().map(|expr| TargetEntry { expr, resjunk: false, ressortgroupref: 0 }).collect();
        return query;
    };
    let mut group_exprs = Vec::new();
    if let Some((groups, aggregates, sets)) = upper.aggregation {
        group_exprs = groups.into_iter().map(|g| replace(&g, &meaning)).collect::<Vec<Expr>>();
        query.aggregates = aggregates.into_iter().map(|call| map_call(call, &meaning)).collect();
        meaning = group_exprs.clone();
        meaning.extend((0..query.aggregates.len()).map(Expr::AggRef));
        if sets.is_some() {
            meaning.push(Expr::Const(crate::types::Value::Null));
        }
        query.grouping_sets = sets;
        query.having_qual = upper.having.map(|h| replace(&h, &meaning));
    }
    for calls in upper.windows {
        let first = query.window_funcs.len();
        let calls: Vec<crate::window::WindowCall> = calls.into_iter().map(|c| map_window(c, &meaning)).collect();
        meaning.extend((first..first + calls.len()).map(Expr::WindowRef));
        query.window_funcs.extend(calls);
    }
    for functions in upper.srfs {
        let first = query.target_srfs.len();
        let functions: Vec<Expr> = functions.into_iter().map(|f| replace(&f, &meaning)).collect();
        meaning.extend((first..first + functions.len()).map(Expr::SetRef));
        query.target_srfs.extend(functions);
    }
    let width = upper.visible.unwrap_or(upper.target.len());
    query.target_list = upper
        .target
        .iter()
        .enumerate()
        .map(|(i, e)| TargetEntry { expr: replace(e, &meaning), resjunk: i >= width, ressortgroupref: 0 })
        .collect();
    let mut next_ref = 0;
    let mut sortgroupref = |query: &mut Query, position: usize| {
        let tle = &mut query.target_list[position];
        if tle.ressortgroupref == 0 {
            next_ref += 1;
            tle.ressortgroupref = next_ref;
        }
        tle.ressortgroupref
    };
    let clause = |tle_sort_group_ref: usize, descending: bool, nulls_first: bool| SortGroupClause {
        tle_sort_group_ref,
        descending,
        nulls_first,
        hashable: true,
    };
    for key in &upper.sort {
        let Expr::Column(position) = key.expr else { continue };
        let tle_sort_group_ref = sortgroupref(&mut query, position);
        query.sort_clause.push(clause(tle_sort_group_ref, key.descending, key.nulls_first));
    }
    for group in group_exprs {
        let position = match query.target_list.iter().position(|tle| tle.expr == group) {
            Some(position) => position,
            None => {
                query.target_list.push(TargetEntry { expr: group, resjunk: true, ressortgroupref: 0 });
                query.target_list.len() - 1
            }
        };
        let tle_sort_group_ref = sortgroupref(&mut query, position);
        let sorted = query.sort_clause.iter().find(|c| c.tle_sort_group_ref == tle_sort_group_ref);
        let (descending, nulls_first) = sorted.map_or((false, false), |c| (c.descending, c.nulls_first));
        query.group_clause.push(clause(tle_sort_group_ref, descending, nulls_first));
    }
    match upper.distinct {
        Some(Some(on)) => {
            query.has_distinct_on = true;
            for key in on {
                let Expr::Column(position) = key else { continue };
                let tle_sort_group_ref = sortgroupref(&mut query, position);
                let sorted = query.sort_clause.iter().find(|c| c.tle_sort_group_ref == tle_sort_group_ref);
                let (descending, nulls_first) = sorted.map_or((false, false), |c| (c.descending, c.nulls_first));
                query.distinct_clause.push(clause(tle_sort_group_ref, descending, nulls_first));
            }
        }
        Some(None) => {
            for position in 0..width {
                let tle_sort_group_ref = sortgroupref(&mut query, position);
                let sorted = query.sort_clause.iter().find(|c| c.tle_sort_group_ref == tle_sort_group_ref);
                let (descending, nulls_first) = sorted.map_or((false, false), |c| (c.descending, c.nulls_first));
                query.distinct_clause.push(clause(tle_sort_group_ref, descending, nulls_first));
            }
        }
        None => {}
    }
    if let Some((limit, offset)) = upper.limit {
        query.limit_count = limit;
        query.limit_offset = offset;
    }
    query
}

/// peel returns a SELECT's FROM and WHERE plan with the upper clauses above it, or None when the plan has no layers
/// above a projection of its select list.
fn peel(plan: Plan) -> (Plan, Option<Upper>) {
    let mut upper = Upper::default();
    let mut plan = plan;
    if let Plan::Project { input, exprs } = &plan
        && exprs.iter().enumerate().all(|(i, e)| *e == Expr::Column(i))
        && matches!(**input, Plan::Limit { .. } | Plan::Sort { .. } | Plan::Distinct { .. } | Plan::Project { .. })
        && input.width() > exprs.len()
    {
        upper.visible = Some(exprs.len());
        let Plan::Project { input, .. } = plan else { unreachable!("a projection") };
        plan = *input;
    }
    if let Plan::Limit { input, limit, offset } = plan {
        upper.limit = Some((limit, offset));
        plan = *input;
    }
    plan = match plan {
        Plan::Sort { input, keys } if matches!(*input, Plan::Distinct { .. } | Plan::Project { .. }) => {
            upper.sort = keys;
            *input
        }
        other => other,
    };
    plan = match plan {
        Plan::Distinct { input, keys } => {
            upper.distinct = Some(keys);
            match *input {
                Plan::Sort { input, keys } if upper.sort.is_empty() => {
                    upper.sort = keys;
                    *input
                }
                other => other,
            }
        }
        other => other,
    };
    let Plan::Project { input, exprs } = plan else {
        if upper.limit.is_some() || !upper.sort.is_empty() || upper.distinct.is_some() {
            let width = plan.width();
            upper.target = (0..width).map(Expr::Column).collect();
            return peel_from(plan, upper);
        }
        return (plan, None);
    };
    upper.target = exprs;
    peel_from(*input, upper)
}

/// peel_from peels the set-returning functions, windows, HAVING, and aggregation off the plan under a SELECT's
/// projection.
fn peel_from(plan: Plan, mut upper: Upper) -> (Plan, Option<Upper>) {
    let mut plan = plan;
    loop {
        plan = match plan {
            Plan::ProjectSet { input, functions, .. } => {
                upper.srfs.insert(0, functions);
                *input
            }
            Plan::Window { input, calls } => {
                upper.windows.insert(0, calls);
                match *input {
                    Plan::Sort { input, .. } => *input,
                    other => other,
                }
            }
            other => break plan = other,
        };
    }
    if let Plan::Filter { input, predicate } = plan {
        match *input {
            Plan::Aggregate { .. } => {
                upper.having = Some(predicate);
                plan = *input;
            }
            other => plan = Plan::Filter { input: Box::new(other), predicate },
        }
    }
    if let Plan::Aggregate { input, groups, aggregates, sets } = plan {
        upper.aggregation = Some((groups, aggregates, sets));
        plan = *input;
    }
    (plan, Some(upper))
}

/// replace rewrites an expression over a layer's columns into one over what they hold.
fn replace(e: &Expr, meaning: &[Expr]) -> Expr {
    match e {
        Expr::Column(c) => meaning[*c].clone(),
        Expr::Grouping(args, locations, _) => Expr::Grouping(args.clone(), locations.clone(), None),
        other => other.clone().map_children(&mut |c| replace(&c, meaning)),
    }
}

/// map_call rewrites an aggregate call's expressions over its input's columns into ones over what they hold.
fn map_call(mut call: crate::functions::aggregate::AggCall, meaning: &[Expr]) -> crate::functions::aggregate::AggCall {
    call.args = call.args.iter().map(|a| replace(a, meaning)).collect();
    call.filter = call.filter.as_ref().map(|f| replace(f, meaning));
    call.order = call.order.iter().map(|(e, d, n)| (replace(e, meaning), *d, *n)).collect();
    call
}

/// map_window rewrites a window call's expressions over its input's columns into ones over what they hold.
fn map_window(mut call: crate::window::WindowCall, meaning: &[Expr]) -> crate::window::WindowCall {
    call.args = call.args.iter().map(|a| replace(a, meaning)).collect();
    call.filter = call.filter.as_ref().map(|f| replace(f, meaning));
    call.partition = call.partition.iter().map(|p| replace(p, meaning)).collect();
    for key in &mut call.order {
        key.expr = replace(&key.expr, meaning);
    }
    if let Some(range) = &mut call.range {
        range.key = replace(&range.key, meaning);
        for (e, _) in range.start.iter_mut().chain(range.end.iter_mut()) {
            *e = replace(e, meaning);
        }
    }
    call
}
