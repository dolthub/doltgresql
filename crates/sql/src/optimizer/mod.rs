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

//! A port of Postgres 15's optimizer, file by file: each module ports the Postgres source file it names, keeping
//! its functions' names and logic, over Doltgres' bound expressions and plans. Planning starts from a SELECT's FROM
//! clause, as the binder planned it without choosing join orders or methods, and its WHERE clause, and produces the
//! plan that Postgres' query_planner would find cheapest.

mod allpaths;
mod analyzejoins;
mod clauses;
mod clausesel;
mod costsize;
mod createplan;
mod indxpath;
mod initsplan;
mod joinpath;
mod joinrels;
pub mod nodes;
mod pathkeys;
mod pathnode;
mod planner;
mod prepjointree;
mod relnode;
mod restrictinfo;
mod selfuncs;
mod subselect;

use std::collections::{BTreeSet, HashMap};

pub(crate) use subselect::{plan_sublink, sublink_convertible};

use nodes::{
    FromExpr, JoinExpr, JoinTreeNode, JoinType, Query, RangeTblEntry, RelOptInfo, Relids, SpecialJoinInfo, var,
};

use crate::expr::Expr;
use crate::functions::aggregate::AggCall;
use crate::indexscan::columns_read;
use crate::plan::{JoinKind, JoinMethod, Plan, SortKey};
use crate::query::Ctx;

/// MAX_RELATIONS is how many range table entries a query may have for Relids to hold them, as range table indexes
/// start at 1.
const MAX_RELATIONS: usize = 63;

/// PlannerInfo is the state of planning one query, as Postgres' PlannerInfo holds it.
pub struct PlannerInfo<'r, 'a> {
    pub ctx: &'r mut Ctx<'a>,
    pub parse: Query,
    /// The relations, by index: the base relations at their range table indexes, then the join relations.
    pub rels: Vec<RelOptInfo>,
    /// The join relations' indexes by their relations.
    pub join_rel_hash: HashMap<Relids, usize>,
    /// The relations of each level of the join search running now, from level 1.
    pub join_rel_level: Vec<Vec<usize>>,
    /// The relations that the join search running now joins.
    pub initial_rels: Vec<usize>,
    pub all_baserels: Relids,
    /// The outer, semi, and anti joins of the query.
    pub join_info_list: Vec<SpecialJoinInfo>,
    /// The pages of every table that the query reads, which Postgres' index_pages_fetched shares the cache among.
    pub total_table_pages: f64,
    pub enables: costsize::Enables,
    /// The rows that the query reads of the join of every relation, or a fraction of them below one, or zero for all.
    pub tuple_fraction: f64,
    /// The order of the rows that the query's ORDER BY asks for, over Vars, as Postgres' query_pathkeys.
    pub query_pathkeys: Vec<SortKey>,
    /// The aggregate calls of an aggregate without groups over the join's rows, which Doltgres' executor answers from
    /// an index's entry counts when they only count rows.
    pub counting: Option<Vec<AggCall>>,
}

/// LIMIT_FRACTION is the share of a query's rows that Postgres assumes a LIMIT that is not a constant reads.
const LIMIT_FRACTION: f64 = 0.10;

/// enabled reports whether the DOLTGRES_PG_PLANNER environment variable asks for this planner, which stands beside
/// the older one until it plans every query that the older one does.
pub(crate) fn enabled() -> bool {
    static ENABLED: std::sync::OnceLock<bool> = std::sync::OnceLock::new();
    *ENABLED.get_or_init(|| std::env::var_os("DOLTGRES_PG_PLANNER").is_some())
}

/// plannable reports whether query_planner can plan a FROM clause's plan: anything but the empty FROM clause's one
/// row, of at most MAX_RELATIONS inputs.
fn plannable(from: &Plan) -> bool {
    /// inputs counts the plans that a FROM clause's plan may turn into range table entries.
    fn inputs(plan: &Plan) -> usize {
        match plan {
            Plan::Join { left, right, .. } => inputs(left) + inputs(right),
            Plan::Project { input, .. } | Plan::Filter { input, .. } => inputs(input),
            _ => 1,
        }
    }
    !matches!(from, Plan::OneRow) && inputs(from) <= MAX_RELATIONS
}

/// Upper is what the nodes above a query's FROM and WHERE clauses ask of them: the columns of their rows that the
/// nodes read, or None for every column, the rows they read, as Postgres' tuple_fraction counts them (zero for all),
/// the most rows that a LIMIT reads (zero without one), the order of ORDER BY keys over their rows that a sort above
/// asks for, and the aggregate calls of an aggregate without groups directly above them.
#[derive(Clone, Default)]
struct Upper {
    needed: Option<BTreeSet<usize>>,
    tuple_fraction: f64,
    limit_tuples: f64,
    order: Option<Vec<SortKey>>,
    counting: Option<Vec<AggCall>>,
}

/// planner plans the FROM and WHERE clauses of a query whose upper parts are already built over them, given what the
/// upper parts read, as Postgres' grouping_planner calls query_planner, removing a sort that the cheapest plan's
/// order makes unnecessary.
pub(crate) fn planner(ctx: &mut Ctx<'_>, plan: Plan) -> Plan {
    plan_upper(ctx, plan, Upper::default()).0
}

/// plan_upper is planner for a plan under nodes that ask what `upper` holds, also reporting whether the plan's rows
/// come in the order that `upper` asks for.
fn plan_upper(ctx: &mut Ctx<'_>, plan: Plan, upper: Upper) -> (Plan, bool) {
    let union = |a: Option<BTreeSet<usize>>, b: Option<BTreeSet<usize>>| Some(a?.union(&b?).copied().collect());
    let below =
        |needed: Option<BTreeSet<usize>>, width: usize| needed.map(|n| n.into_iter().filter(|&c| c < width).collect());
    let all = |needed| Upper { needed, ..Upper::default() };
    let unordered = |plan: Plan| (plan, false);
    let Upper { needed, tuple_fraction, limit_tuples, order, counting } = upper;
    match plan {
        Plan::Project { input, exprs } => {
            let read = match &needed {
                Some(needed) => columns_read(needed.iter().filter_map(|&i| exprs.get(i))),
                None => columns_read(exprs.iter()),
            };
            let order = order.and_then(|keys| {
                let project = |k: &SortKey| match k.expr {
                    Expr::Column(i) => Some(SortKey { expr: exprs.get(i)?.clone(), ..k.clone() }),
                    _ => None,
                };
                keys.iter().map(project).collect::<Option<Vec<SortKey>>>()
            });
            let upper = Upper { needed: read, tuple_fraction, limit_tuples, order, counting: None };
            let (input, ordered) = plan_upper(ctx, *input, upper);
            (Plan::Project { input: Box::new(input), exprs }, ordered)
        }
        Plan::Sort { input, keys } => {
            let read = union(needed, columns_read(keys.iter().map(|k| &k.expr)));
            let order = Some(keys.clone());
            let upper = Upper { needed: read, tuple_fraction, limit_tuples, order, counting: None };
            match plan_upper(ctx, *input, upper) {
                (input, true) => unordered(input),
                (input, false) => unordered(Plan::Sort { input: Box::new(input), keys }),
            }
        }
        Plan::Limit { input, limit, offset } => {
            let count = |e: &Option<Expr>| match e {
                Some(Expr::Const(crate::types::Value::Int8(n))) => Some(*n as f64),
                Some(_) => None,
                None => Some(0.0),
            };
            let (tuple_fraction, limit_tuples) = match (count(&limit), count(&offset)) {
                (Some(limit), Some(offset)) if limit > 0.0 => (limit + offset, limit + offset),
                _ => (LIMIT_FRACTION, 0.0),
            };
            let upper = Upper { needed, tuple_fraction, limit_tuples, order: None, counting: None };
            unordered(Plan::Limit { input: Box::new(plan_upper(ctx, *input, upper).0), limit, offset })
        }
        Plan::Distinct { input, keys } => {
            let read = match &keys {
                Some(keys) => union(needed, columns_read(keys.iter())),
                None => None,
            };
            unordered(Plan::Distinct { input: Box::new(plan_upper(ctx, *input, all(read)).0), keys })
        }
        Plan::Aggregate { input, groups, aggregates, sets } => {
            let mut exprs: Vec<&Expr> = groups.iter().collect();
            for call in &aggregates {
                exprs.extend(&call.args);
                exprs.extend(&call.filter);
                exprs.extend(call.order.iter().map(|(e, _, _)| e));
            }
            let read = columns_read(exprs);
            let counting = (groups.is_empty() && sets.is_none()).then(|| aggregates.clone());
            let upper = Upper { needed: read, counting, ..Upper::default() };
            let input = Box::new(plan_upper(ctx, *input, upper).0);
            unordered(Plan::Aggregate { input, groups, aggregates, sets })
        }
        Plan::Window { input, calls } => {
            let width = input.width();
            let mut exprs: Vec<&Expr> = Vec::new();
            for call in &calls {
                exprs.extend(&call.args);
                exprs.extend(&call.filter);
                exprs.extend(&call.partition);
                exprs.extend(call.order.iter().map(|k| &k.expr));
                if let Some(range) = &call.range {
                    exprs.push(&range.key);
                    exprs.extend(range.start.iter().chain(&range.end).map(|(e, _)| e));
                }
            }
            let read = union(below(needed, width), columns_read(exprs));
            unordered(Plan::Window { input: Box::new(plan_upper(ctx, *input, all(read)).0), calls })
        }
        Plan::ProjectSet { input, functions, dropped } => {
            let read = union(below(needed, input.width()), columns_read(functions.iter()));
            unordered(Plan::ProjectSet { input: Box::new(plan_upper(ctx, *input, all(read)).0), functions, dropped })
        }
        Plan::Filter { input, predicate } if matches!(*input, Plan::Aggregate { .. } | Plan::Window { .. }) => {
            let read = union(needed, columns_read([&predicate]));
            unordered(Plan::Filter { input: Box::new(plan_upper(ctx, *input, all(read)).0), predicate })
        }
        Plan::Filter { input, predicate } if plannable(&input) => {
            let quals = crate::indexscan::conjuncts(&predicate).into_iter().cloned().collect();
            query_planner(ctx, *input, quals, Upper { needed, tuple_fraction, limit_tuples, order, counting })
        }
        scan @ Plan::Scan(..) if order.is_some() => {
            query_planner(ctx, scan, Vec::new(), Upper { needed, tuple_fraction, limit_tuples, order, counting })
        }
        join @ Plan::Join { .. } if plannable(&join) => {
            query_planner(ctx, join, Vec::new(), Upper { needed, tuple_fraction, limit_tuples, order, counting })
        }
        other => unordered(other),
    }
}

/// plan_subquery plans a subquery that its query left unplanned for pulling up, when it stays a relation of its
/// own, as Postgres' subquery_planner plans a subquery that it could not pull up.
fn plan_subquery(ctx: &mut Ctx<'_>, plan: Plan) -> Plan {
    match plan {
        join @ Plan::Join { .. } if !decomposable(&join) => match join {
            Plan::Join { left, right, kind, condition, lateral, method } => Plan::Join {
                left: Box::new(plan_subquery(ctx, *left)),
                right: Box::new(plan_subquery(ctx, *right)),
                kind,
                condition,
                lateral,
                method,
            },
            other => other,
        },
        other => planner(ctx, other),
    }
}

/// decomposable reports whether a join's inputs can join in any order the planner finds: it is not lateral, not
/// already planned, and has no subquery in its condition.
fn decomposable(join: &Plan) -> bool {
    matches!(join, Plan::Join { condition, lateral: false, method: JoinMethod::Unplanned, .. }
        if !condition.as_ref().is_some_and(crate::plan::has_subquery))
}

/// query_planner plans a FROM clause's plan under the conjuncts of the WHERE clause, as Postgres' query_planner
/// plans a query's join tree: it reduces outer joins, distributes the clauses to the relations they restrict, finds
/// each relation's cheapest paths, searches for the cheapest join order and methods, and makes a plan of the
/// cheapest path, whose columns are in the FROM clause's order, given the columns that the query reads above it, or
/// None for every column, also reporting whether the plan's rows come in the order that `upper` asks for.
/// Conditions with subqueries filter the plan's rows afterwards.
fn query_planner(ctx: &mut Ctx<'_>, from: Plan, quals: Vec<Expr>, upper: Upper) -> (Plan, bool) {
    let Upper { needed, tuple_fraction, limit_tuples, order, counting } = upper;
    let (mut rtable, mut output) = (Vec::new(), Vec::new());
    let node = build_jointree(ctx, from, &mut rtable, &mut output, false);
    let (node, quals) = subselect::pull_up_sublinks(ctx, node, quals, &mut rtable, &output);
    let (kept, later): (Vec<Expr>, Vec<Expr>) = quals.into_iter().partition(|q| !crate::plan::has_subquery(q));
    let mut jointree =
        FromExpr { fromlist: vec![node], quals: kept.into_iter().map(|q| to_vars(q, &output)).collect() };
    let usable = |keys: &Vec<SortKey>| keys.iter().all(|k| !crate::plan::has_subquery(&k.expr));
    let query_pathkeys: Vec<SortKey> = order.filter(usable).map_or_else(Vec::new, |keys| {
        keys.into_iter().map(|k| SortKey { expr: to_vars(k.expr, &output), ..k }).collect()
    });
    let needed = needed.and_then(|n| Some(n.union(&columns_read(&later)?).copied().collect::<BTreeSet<usize>>()));
    let output =
        output.into_iter().enumerate().map(|(i, e)| needed.as_ref().is_none_or(|n| n.contains(&i)).then_some(e));
    let output = output.collect();
    prepjointree::reduce_outer_joins(&mut jointree);
    let parse = Query { rtable, jointree, output };
    let enables = costsize::Enables::read(&ctx.session.settings);
    let mut root = PlannerInfo {
        ctx,
        all_baserels: (1..=parse.rtable.len()).fold(0, |relids, varno| relids | nodes::singleton(varno)),
        parse,
        rels: Vec::new(),
        join_rel_hash: HashMap::new(),
        join_rel_level: Vec::new(),
        initial_rels: Vec::new(),
        join_info_list: Vec::new(),
        total_table_pages: 0.0,
        enables,
        tuple_fraction,
        query_pathkeys,
        counting,
    };
    relnode::add_base_rels_to_query(&mut root);
    let jointree = root.parse.jointree.clone();
    let joinlist = initsplan::deconstruct_jointree(&mut root, &jointree);
    let joinlist = analyzejoins::remove_useless_joins(&mut root, joinlist);
    let final_rel = allpaths::make_one_rel(&mut root, joinlist);
    let path = planner::create_ordered_paths(&mut root, final_rel, limit_tuples);
    let ordered = !root.query_pathkeys.is_empty() && !matches!(path.kind, nodes::PathKind::Sort(_));
    let plan = createplan::create_plan(&mut root, &path);
    let plan = match later.into_iter().reduce(|a, b| Expr::And(Box::new(a), Box::new(b))) {
        Some(predicate) => Plan::Filter { input: Box::new(plan), predicate },
        None => plan,
    };
    (plan, ordered)
}

/// build_jointree adds the inputs of a FROM clause's joins to a range table and returns its join tree, adding the
/// expression of each column of the plan's rows to `output`, given whether an outer join can make the plan's rows
/// NULL. A right join becomes a left join of its inputs swapped. A projection or filter over any input is pulled up
/// into the join tree, as Postgres' pull_up_subqueries pulls up a simple subquery, unless one of its expressions
/// would need a PlaceHolderVar, which is not ported yet. Any other input, or any join that is lateral, already
/// planned, or has a subquery in its condition, becomes a relation of its own.
pub(super) fn build_jointree(
    ctx: &mut Ctx<'_>,
    plan: Plan,
    rtable: &mut Vec<RangeTblEntry>,
    output: &mut Vec<Expr>,
    nullable: bool,
) -> JoinTreeNode {
    match plan {
        Plan::Join { left, right, kind, condition, .. } if decomposable(&plan) => {
            let start = output.len();
            let (left_nullable, right_nullable) = match kind {
                JoinKind::Inner => (nullable, nullable),
                JoinKind::Left | JoinKind::Semi | JoinKind::Anti => (nullable, true),
                JoinKind::Right => (true, nullable),
                JoinKind::Full => (true, true),
            };
            let larg = build_jointree(ctx, *left, rtable, output, left_nullable);
            let rarg = build_jointree(ctx, *right, rtable, output, right_nullable);
            let conjuncts = condition.as_ref().map(crate::indexscan::conjuncts).unwrap_or_default();
            let quals = conjuncts.into_iter().map(|c| to_vars(c.clone(), &output[start..])).collect();
            let (jointype, larg, rarg) = match kind {
                JoinKind::Inner => (JoinType::Inner, larg, rarg),
                JoinKind::Left => (JoinType::Left, larg, rarg),
                JoinKind::Right => (JoinType::Left, rarg, larg),
                JoinKind::Full => (JoinType::Full, larg, rarg),
                JoinKind::Semi => (JoinType::Semi, larg, rarg),
                JoinKind::Anti => (JoinType::Anti, larg, rarg),
            };
            JoinTreeNode::Join(Box::new(JoinExpr { jointype, larg, rarg, quals }))
        }
        Plan::Project { input, exprs } if exprs.iter().all(|e| pullable(e, nullable)) => {
            let mut columns = Vec::new();
            let node = build_jointree(ctx, *input, rtable, &mut columns, nullable);
            output.extend(exprs.into_iter().map(|e| to_vars(e, &columns)));
            node
        }
        Plan::Filter { input, predicate } if !crate::plan::has_subquery(&predicate) => {
            let start = output.len();
            let node = build_jointree(ctx, *input, rtable, output, nullable);
            let conjuncts = crate::indexscan::conjuncts(&predicate);
            let quals = conjuncts.into_iter().map(|c| to_vars(c.clone(), &output[start..])).collect();
            JoinTreeNode::From(Box::new(FromExpr { fromlist: vec![node], quals }))
        }
        other => {
            let varno = rtable.len() + 1;
            output.extend((0..other.width()).map(|attno| Expr::Column(var(varno, attno))));
            rtable.push(RangeTblEntry { plan: plan_subquery(ctx, other) });
            JoinTreeNode::Rel(varno)
        }
    }
}

/// pullable reports whether a projection's expression can replace references to its column in the join tree above
/// it: one that runs no volatile function or subquery and reads only the projection's input, and, where an outer join
/// can make the input's rows NULL, one that is NULL then too, as a Var is and a strict expression of Vars is, which
/// Postgres would otherwise wrap in a PlaceHolderVar.
fn pullable(e: &Expr, nullable: bool) -> bool {
    let mut reads_only_columns = true;
    e.visit(&mut |x| {
        reads_only_columns &= !matches!(
            x,
            Expr::InputColumn(_)
                | Expr::AggRef(_)
                | Expr::WindowRef(_)
                | Expr::SetRef(_)
                | Expr::Grouping(..)
                | Expr::SubqueryValue
                | Expr::Default(_)
        )
    });
    if !reads_only_columns || crate::plan::has_subquery(e) || clauses::contain_volatile_functions(e) {
        return false;
    }
    !nullable || matches!(e, Expr::Column(_)) || (clauses::contain_vars(e) && !clauses::contain_nonstrict_functions(e))
}

/// to_vars rewrites an expression over a row of columns into one over their expressions.
fn to_vars(e: Expr, columns: &[Expr]) -> Expr {
    match e {
        Expr::Column(c) => columns[c].clone(),
        other => other.map_children(&mut |c| to_vars(c, columns)),
    }
}
