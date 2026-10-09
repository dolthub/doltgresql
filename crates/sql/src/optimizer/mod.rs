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
mod clauses;
mod clausesel;
mod costsize;
mod createplan;
mod indxpath;
mod initsplan;
mod joinpath;
mod joinrels;
pub mod nodes;
mod pathnode;
mod prepjointree;
mod relnode;
mod restrictinfo;
mod selfuncs;

use std::collections::HashMap;

use nodes::{
    FromExpr, JoinExpr, JoinTreeNode, JoinType, Query, RangeTblEntry, RelOptInfo, Relids, SpecialJoinInfo, var,
};

use crate::expr::Expr;
use crate::plan::{JoinKind, JoinMethod, Plan};
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
}

/// enabled reports whether the DOLTGRES_PG_PLANNER environment variable asks for this planner, which stands beside
/// the older one until it plans every query that the older one does.
pub(crate) fn enabled() -> bool {
    static ENABLED: std::sync::OnceLock<bool> = std::sync::OnceLock::new();
    *ENABLED.get_or_init(|| std::env::var_os("DOLTGRES_PG_PLANNER").is_some())
}

/// plannable reports whether query_planner can plan a FROM clause's plan: joins that are not lateral, whose
/// conditions have no subqueries, of at most MAX_RELATIONS inputs.
pub(crate) fn plannable(from: &Plan) -> bool {
    /// inputs counts the inputs of plannable joins, or returns None when a join is not plannable.
    fn inputs(plan: &Plan) -> Option<usize> {
        match plan {
            Plan::Join { left, right, condition, lateral, method, .. } => {
                if *lateral
                    || *method != JoinMethod::Unplanned
                    || condition.as_ref().is_some_and(crate::plan::has_subquery)
                {
                    return None;
                }
                Some(inputs(left)? + inputs(right)?)
            }
            _ => Some(1),
        }
    }
    !matches!(from, Plan::OneRow) && inputs(from).is_some_and(|n| n <= MAX_RELATIONS)
}

/// query_planner plans a FROM clause's plan under the conjuncts of the WHERE clause, as Postgres' query_planner
/// plans a query's join tree: it reduces outer joins, distributes the clauses to the relations they restrict, finds
/// each relation's cheapest paths, searches for the cheapest join order and methods, and makes a plan of the
/// cheapest path, whose columns are in the FROM clause's order. Conditions with subqueries filter the plan's rows
/// afterwards.
pub(crate) fn query_planner(ctx: &mut Ctx<'_>, from: Plan, quals: Vec<Expr>) -> Plan {
    let (mut rtable, mut output) = (Vec::new(), Vec::new());
    let node = build_jointree(from, &mut rtable, &mut output);
    let (kept, later): (Vec<Expr>, Vec<Expr>) = quals.into_iter().partition(|q| !crate::plan::has_subquery(q));
    let mut jointree =
        FromExpr { fromlist: vec![node], quals: kept.into_iter().map(|q| to_vars(q, &output)).collect() };
    prepjointree::reduce_outer_joins(&mut jointree);
    let parse = Query { rtable, jointree, output };
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
    };
    relnode::add_base_rels_to_query(&mut root);
    let jointree = root.parse.jointree.clone();
    let joinlist = initsplan::deconstruct_jointree(&mut root, &jointree);
    let final_rel = allpaths::make_one_rel(&mut root, joinlist);
    let path = root.rels[final_rel].cheapest_total_path.clone().expect("every relation has a path");
    let plan = createplan::create_plan(&mut root, &path);
    match later.into_iter().reduce(|a, b| Expr::And(Box::new(a), Box::new(b))) {
        Some(predicate) => Plan::Filter { input: Box::new(plan), predicate },
        None => plan,
    }
}

/// build_jointree adds the inputs of a FROM clause's joins to a range table and returns its join tree, adding the
/// Var of each column of the plan's rows to `output`. A right join becomes a left join of its inputs swapped.
fn build_jointree(plan: Plan, rtable: &mut Vec<RangeTblEntry>, output: &mut Vec<usize>) -> JoinTreeNode {
    match plan {
        Plan::Join { left, right, kind, condition, .. } => {
            let start = output.len();
            let larg = build_jointree(*left, rtable, output);
            let rarg = build_jointree(*right, rtable, output);
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
        other => {
            let varno = rtable.len() + 1;
            output.extend((0..other.width()).map(|attno| var(varno, attno)));
            rtable.push(RangeTblEntry { plan: other });
            JoinTreeNode::Rel(varno)
        }
    }
}

/// to_vars rewrites an expression over a row of columns into one over their Vars.
fn to_vars(e: Expr, vars: &[usize]) -> Expr {
    match e {
        Expr::Column(c) => Expr::Column(vars[c]),
        other => other.map_children(&mut |c| to_vars(c, vars)),
    }
}
