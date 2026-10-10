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

//! Postgres' optimizer/plan/planagg.c: answering a query of MIN and MAX aggregates over one table without grouping
//! by reading, for each aggregate, the first row of the table in the aggregate's order that an index gives, as an
//! initplan.

use super::PlannerInfo;
use super::nodes::{JoinTreeNode, MinMaxAggInfo, RteKind, SortGroupClause, TargetEntry};
use crate::expr::Expr;
use crate::types::Value;

/// preprocess_minmax_aggregates finds the plan that answers each of the query's MIN and MAX aggregates from an index,
/// when every aggregate of a query over one table without grouping is one that an index can answer, as Postgres'
/// function of the same name does, keeping them for create_grouping_paths to add the path that reads them, since the
/// query's relations are built after this.
pub fn preprocess_minmax_aggregates(root: &mut PlannerInfo<'_, '_>) {
    let parse = &root.parse;
    if !parse.has_aggs
        || !parse.group_clause.is_empty()
        || parse.grouping_sets.as_ref().is_some_and(|sets| sets.len() > 1)
        || !parse.window_funcs.is_empty()
    {
        return;
    }
    let mut fromlist = &parse.jointree.fromlist;
    let rtindex = loop {
        let [node] = fromlist.as_slice() else { return };
        match node {
            JoinTreeNode::From(from) => fromlist = &from.fromlist,
            JoinTreeNode::Rel(rtindex) => break *rtindex,
            JoinTreeNode::Join(_) => return,
        }
    };
    if !matches!(parse.rte(rtindex).kind, RteKind::Relation(..)) {
        return;
    }
    let Some(aggs_list) = can_minmax_aggs(root) else { return };
    let mut mmaggregates = Vec::with_capacity(aggs_list.len());
    for (agg, target, reverse) in aggs_list {
        let Some(info) = build_minmax_path(root, agg, &target, reverse, reverse)
            .or_else(|| build_minmax_path(root, agg, &target, reverse, !reverse))
        else {
            return;
        };
        mmaggregates.push(info);
    }
    root.minmax_aggs = mmaggregates;
}

/// can_minmax_aggs returns each of the query's aggregate calls with its argument and whether it reads the largest
/// value first, when every call is a MIN or MAX of one argument without ORDER BY or FILTER, over an expression that an
/// index may hold, as Postgres' function of the same name does.
fn can_minmax_aggs(root: &PlannerInfo<'_, '_>) -> Option<Vec<(usize, Expr, bool)>> {
    let mut aggs_list = Vec::with_capacity(root.parse.aggregates.len());
    for (agg, call) in root.parse.aggregates.iter().enumerate() {
        let [target] = call.args.as_slice() else { return None };
        if !call.order.is_empty() || call.filter.is_some() {
            return None;
        }
        let reverse = fetch_agg_sort_op(call)?;
        if super::clauses::contain_mutable_functions(root.glob, target)
            || super::nodefuncs::expr_type(root, target) == Some(crate::oid::RECORD)
        {
            return None;
        }
        aggs_list.push((agg, target.clone(), reverse));
    }
    Some(aggs_list)
}

/// build_minmax_path plans the query that reads the first row of the table, in the order of an aggregate's argument
/// with NULLs where `nulls_first` puts them, whose argument is not NULL, when an index gives that order, returning
/// the plan of that row's argument with its cost, as Postgres' function of the same name does.
fn build_minmax_path(
    root: &mut PlannerInfo<'_, '_>,
    agg: usize,
    target: &Expr,
    reverse_sort: bool,
    nulls_first: bool,
) -> Option<MinMaxAggInfo> {
    let mut parse = root.parse.clone();
    super::subselect::increment_query_sublevels_up(&mut parse);
    let target = super::subselect::increment_sublevels_up(target.clone(), 0);
    parse.target_list = vec![TargetEntry { expr: target.clone(), resjunk: false, ressortgroupref: 1 }];
    parse.having_qual = None;
    parse.distinct_clause.clear();
    parse.has_distinct_on = false;
    parse.has_aggs = false;
    parse.aggregates.clear();
    let ntest = Expr::IsNull(Box::new(target), true);
    if !parse.jointree.quals.contains(&ntest) {
        parse.jointree.quals.insert(0, ntest);
    }
    parse.sort_clause =
        vec![SortGroupClause { tle_sort_group_ref: 1, descending: reverse_sort, nulls_first, hashable: false }];
    parse.limit_offset = None;
    parse.limit_count = Some(Expr::Const(Value::Int8(1)));
    let mut subroot = super::new_planner_info(root.ctx, root.glob, parse, 1.0, false);
    subroot.processed_tlist = subroot.parse.target_list.clone();
    subroot.limit_tuples = 1.0;
    let final_rel = super::query_planner(&mut subroot, &mut minmax_qp_callback);
    super::subselect::ss_charge_for_initplans(&mut subroot, final_rel);
    let rows = subroot.rels[final_rel].rows;
    let path_fraction = if rows > 1.0 { 1.0 / rows } else { 1.0 };
    let sorted_path = super::pathkeys::get_cheapest_fractional_path_for_pathkeys(
        &subroot.rels[final_rel].pathlist,
        &subroot.query_pathkeys,
        path_fraction,
    )?;
    let target = std::rc::Rc::new(super::tlist::create_pathtarget(&subroot, &subroot.processed_tlist));
    let sorted_path = super::pathnode::apply_projection_to_path(&mut subroot, final_rel, sorted_path, target);
    let pathcost = sorted_path.startup_cost + path_fraction * (sorted_path.total_cost - sorted_path.startup_cost);
    let plan = super::createplan::create_plan(&mut subroot, &sorted_path);
    Some(MinMaxAggInfo { agg, plan, pathcost, disabled_nodes: sorted_path.disabled_nodes })
}

/// minmax_qp_callback sets the order that build_minmax_path's query asks for, its ORDER BY, as Postgres' function of
/// the same name does.
fn minmax_qp_callback(root: &mut PlannerInfo<'_, '_>) {
    root.group_pathkeys.clear();
    root.window_pathkeys.clear();
    root.distinct_pathkeys.clear();
    let (sort_clause, tlist) = (root.parse.sort_clause.clone(), root.parse.target_list.clone());
    root.sort_pathkeys = super::planner::make_pathkeys_for_sortclauses(root, &sort_clause, &tlist).unwrap_or_default();
    root.query_pathkeys = root.sort_pathkeys.clone();
}

/// fetch_agg_sort_op returns whether an aggregate call's value is the largest of its argument's values rather than
/// the smallest, or None for an aggregate that is neither, as Postgres' function of the same name finds its sort
/// operator. The built-in MIN, MAX, BOOL_AND, BOOL_OR, and EVERY are the ones with sort operators.
fn fetch_agg_sort_op(call: &crate::functions::aggregate::AggCall) -> Option<bool> {
    if call.user.is_some() {
        return None;
    }
    match crate::functions::aggregate::AGGREGATES[call.index].name {
        "min" | "bool_and" | "every" => Some(false),
        "max" | "bool_or" => Some(true),
        _ => None,
    }
}
