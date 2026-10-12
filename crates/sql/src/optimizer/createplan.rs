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

//! Postgres' optimizer/plan/createplan.c and the parts of setrefs.c that it needs: turning the cheapest path into
//! Doltgres' plan, whose expressions read columns of their input rows by position rather than Vars. A Var reads its
//! relation's column wherever an outer join has made it NULL, as the executor pads the rows that outer joins add, and
//! a PlaceHolderVar is computed where it is evaluated and read from there above.

use std::rc::Rc;

use super::PlannerInfo;
use super::costsize::cost_qual_eval_node;
use super::indxpath::{lookup_keys, to_attnos};
use super::nodes::{IndexPath, JoinType, Path, PathKind, RinfoId, RteKind, VarNode};
use super::restrictinfo::{extract_actual_clauses, extract_actual_join_clauses, get_actual_clauses};
use crate::expr::Expr;
use crate::plan::{JoinKind, JoinMethod, Plan};
use crate::types::Value;

/// Slot is what a column of a plan's rows holds: a relation's attribute, a PlaceHolderVar by its ID, or an
/// expression of the query's upper processing that a plan below computed, as an aggregate call or a group key.
#[derive(Clone, Debug, PartialEq)]
enum Slot {
    Var(usize, usize),
    PlaceHolder(usize),
    Expr(Expr),
}

/// create_plan makes the plan of the query's final path, whose columns are the visible columns of the query's target
/// list, or for an UPDATE, DELETE, or MERGE, the columns of the rows it changes.
pub fn create_plan(root: &mut PlannerInfo<'_, '_>, path: &Path) -> Plan {
    let (plan, layout) = create_plan_recurse(root, path);
    if let PathKind::ModifyTable(_) = path.kind {
        return plan;
    }
    let exprs: Vec<Expr> = root
        .parse
        .target_list
        .iter()
        .enumerate()
        .filter(|(_, tle)| !tle.resjunk)
        .map(|(i, tle)| positional_at(root, tle.expr.clone(), &layout, i))
        .collect();
    if exprs.len() == layout.len() && exprs.iter().enumerate().all(|(i, e)| *e == Expr::Column(i)) {
        return plan;
    }
    fix_alternative_subplans(root, Plan::Project { input: Box::new(plan), exprs }, path.rows)
}

/// target_slots returns the slots of the expressions of a target.
fn target_slots(root: &PlannerInfo<'_, '_>, exprs: &[Expr]) -> Vec<Slot> {
    exprs.iter().map(|e| expr_slot(root, e)).collect()
}

/// expr_slot returns the slot of an expression: a Var's or PlaceHolderVar's own slot, or the expression.
fn expr_slot(root: &PlannerInfo<'_, '_>, e: &Expr) -> Slot {
    match e {
        Expr::Column(_) => slot(root, e),
        other => Slot::Expr(other.clone()),
    }
}

/// project returns a plan that computes a path's target over a plan's rows of a layout, unless the rows are that
/// target already.
fn project(root: &PlannerInfo<'_, '_>, plan: Plan, layout: Vec<Slot>, path: &Path) -> (Plan, Vec<Slot>) {
    let exprs = super::planner::path_exprs(root, path);
    let slots = target_slots(root, &exprs);
    if slots == layout {
        return (plan, layout);
    }
    let positional_exprs = exprs.iter().enumerate().map(|(i, e)| positional_at(root, e.clone(), &layout, i)).collect();
    (Plan::Project { input: Box::new(plan), exprs: positional_exprs }, slots)
}

/// positional_at is positional for the expression of a target's column at a position, which reads the rows' column at
/// that position when the rows hold the expression there, so that equal calls of a volatile function each keep their
/// own value.
fn positional_at(root: &PlannerInfo<'_, '_>, e: Expr, layout: &[Slot], position: usize) -> Expr {
    match layout.get(position) == Some(&expr_slot(root, &e)) {
        true => Expr::Column(position),
        false => positional(root, e, layout),
    }
}

/// sort_keys returns the sort keys of pathkeys over a plan's rows of a layout, as Postgres' prepare_sort_from_pathkeys
/// finds them: the first column of the rows that holds a member of each key's class, as find_ec_member_matching_expr
/// matches it, or else the first member that the columns compute, as find_computable_ec_member finds it. Child members
/// count for the child relations whose columns the rows hold.
fn sort_keys(
    root: &PlannerInfo<'_, '_>,
    pathkeys: &[super::nodes::PkId],
    layout: &[Slot],
) -> Vec<crate::plan::SortKey> {
    let relids: Vec<usize> = layout
        .iter()
        .filter_map(|s| match s {
            Slot::Var(varno, _) => Some(*varno),
            _ => None,
        })
        .collect();
    pathkeys
        .iter()
        .filter_map(|&pk| {
            let pathkey = &root.canon_pathkeys[pk];
            let class = &root.eq_classes[pathkey.pk_eclass];
            let children = class.ec_childmembers.iter().enumerate().filter(|(relid, _)| relids.contains(relid));
            let members: Vec<&Expr> = class
                .ec_members
                .iter()
                .chain(children.flat_map(|(_, members)| members))
                .map(|&em| &root.eq_members[em])
                .filter(|em| !em.em_is_const)
                .map(|em| &em.em_expr)
                .collect();
            let column = layout.iter().position(|s| members.iter().any(|e| expr_slot(root, e) == *s));
            let expr = match column {
                Some(column) => Expr::Column(column),
                None => positional(root, (*members.iter().find(|e| computable(root, e, layout))?).clone(), layout),
            };
            Some(crate::plan::SortKey { expr, descending: pathkey.pk_descending, nulls_first: pathkey.pk_nulls_first })
        })
        .collect()
}

/// computable reports whether an expression can be computed from rows of a layout: the rows hold it, or its Vars
/// and PlaceHolderVars, without reading an aggregate, window, or set-returning call that they do not hold.
fn computable(root: &PlannerInfo<'_, '_>, e: &Expr, layout: &[Slot]) -> bool {
    if layout.contains(&Slot::Expr(e.clone())) {
        return true;
    }
    match e {
        Expr::Column(_) => layout.contains(&slot(root, e)),
        Expr::AggRef(_) | Expr::WindowRef(_) | Expr::SetRef(_) => false,
        other => {
            let mut ok = true;
            other.visit_children(&mut |c| ok &= computable(root, c, layout));
            ok
        }
    }
}

/// create_upper_plan makes the plan of a path of the query's upper processing, as Postgres' createplan functions of
/// each kind do, returning it with what each column of its rows holds.
fn create_upper_plan(root: &mut PlannerInfo<'_, '_>, path: &Path) -> (Plan, Vec<Slot>) {
    match &path.kind {
        PathKind::Projection(subpath) => {
            let (plan, layout) = create_plan_recurse(root, subpath);
            project(root, plan, layout, path)
        }
        PathKind::Sort(subpath) | PathKind::IncrementalSort(subpath) => {
            let (plan, layout) = create_plan_recurse(root, subpath);
            let keys = sort_keys(root, &path.pathkeys, &layout);
            (Plan::Sort { input: Box::new(plan), keys }, layout)
        }
        PathKind::Unique(subpath, num_keys) => {
            let (plan, layout) = create_plan_recurse(root, subpath);
            let keys = sort_keys(root, &subpath.pathkeys[..*num_keys], &layout);
            let keys = Some(keys.into_iter().map(|k| k.expr).collect());
            (Plan::Distinct { input: Box::new(plan), keys }, layout)
        }
        PathKind::Limit(lpath) => {
            let (plan, layout) = create_plan_recurse(root, &lpath.subpath);
            let plan = Plan::Limit {
                input: Box::new(plan),
                limit: lpath.limit_count.clone(),
                offset: lpath.limit_offset.clone(),
            };
            (plan, layout)
        }
        PathKind::Agg(apath) => {
            let group_exprs = tlist_exprs(root, &apath.group_clause);
            create_agg_plan(root, path, &apath.subpath, group_exprs, &apath.qual, true, None)
        }
        PathKind::Group(gpath) => {
            let group_exprs = tlist_exprs(root, &gpath.group_clause);
            create_agg_plan(root, path, &gpath.subpath, group_exprs, &gpath.qual, false, None)
        }
        PathKind::GroupingSets(gspath) => create_groupingsets_plan(root, path, gspath),
        PathKind::WindowAgg(wpath) => {
            let (plan, mut layout) = create_plan_recurse(root, &wpath.subpath);
            let calls: Vec<crate::window::WindowCall> =
                wpath.calls.iter().map(|&k| window_call(root, root.parse.window_funcs[k].clone(), &layout)).collect();
            layout.extend(wpath.calls.iter().map(|&k| Slot::Expr(Expr::WindowRef(k))));
            let plan = Plan::Window { input: Box::new(plan), calls };
            project(root, plan, layout, path)
        }
        PathKind::ProjectSet(subpath) => {
            let (plan, mut layout) = create_plan_recurse(root, subpath);
            let mut functions = Vec::new();
            for e in super::planner::path_exprs(root, path) {
                if let Expr::SetRef(k) = e
                    && !layout.contains(&Slot::Expr(Expr::SetRef(k)))
                {
                    functions.push(positional(root, root.parse.target_srfs[k].clone(), &layout));
                    layout.push(Slot::Expr(Expr::SetRef(k)));
                }
            }
            let plan = Plan::ProjectSet { input: Box::new(plan), functions, dropped: Vec::new() };
            project(root, plan, layout, path)
        }
        PathKind::Append(subpaths) if root.rels[path.parent].reloptkind != super::nodes::RelOptKind::BaseRel => {
            let mut plans = subpaths.iter().map(|p| create_projected_plan(root, p)).collect::<Vec<Plan>>().into_iter();
            let first = plans.next().expect("an Append of upper paths has paths");
            let plan = plans.fold(first, |left, right| Plan::SetOp {
                op: crate::plan::SetOp::Union,
                all: true,
                left: Box::new(left),
                right: Box::new(right),
            });
            (plan, target_slots(root, &super::planner::path_exprs(root, path)))
        }
        PathKind::MergeAppend(subpaths) => create_merge_append_plan(root, path, subpaths),
        PathKind::MinMaxAgg(minmax) => create_minmaxagg_plan(root, path, minmax),
        PathKind::SetOp(setop) => {
            let left = create_projected_plan(root, &setop.leftpath);
            let right = create_projected_plan(root, &setop.rightpath);
            let plan = Plan::SetOp { op: setop.op, all: setop.all, left: Box::new(left), right: Box::new(right) };
            (plan, target_slots(root, &super::planner::path_exprs(root, path)))
        }
        PathKind::RecursiveUnion(runion) => {
            let anchor = create_projected_plan(root, &runion.leftpath);
            let step = create_projected_plan(root, &runion.rightpath);
            let plan = Plan::Recursive {
                work_table: runion.wt_param_id,
                anchor: Box::new(anchor),
                step: Box::new(step),
                all: !runion.distinct,
            };
            (plan, target_slots(root, &super::planner::path_exprs(root, path)))
        }
        PathKind::Result(quals) => {
            let quals = quals.iter().map(|q| positional(root, q.clone(), &[]));
            let predicate = quals.reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
            let plan = match predicate {
                Some(condition) => Plan::OneTimeFilter { input: Box::new(Plan::OneRow), condition },
                None => Plan::OneRow,
            };
            project(root, plan, Vec::new(), path)
        }
        _ => unreachable!("an upper path"),
    }
}

/// create_append_plan makes the plan of an append relation's rows, each member's plan computing the relation's target
/// in turn, as Postgres' function of the same name does, where a plan of one member is that member's plan. On the
/// inner side of a nested loop, a parameterized member's plan tests its parameterizing join clauses, reading the
/// columns of the outer rows of the given layout as `Expr::Outer(0, _)`.
fn create_append_plan(
    root: &mut PlannerInfo<'_, '_>,
    path: &Path,
    subpaths: &[Rc<Path>],
    outer: &[Slot],
) -> (Plan, Vec<Slot>) {
    let mut plans = Vec::with_capacity(subpaths.len());
    for subpath in subpaths {
        let (plan, layout) = match &subpath.kind {
            _ if subpath.param.is_empty() => create_plan_recurse(root, subpath),
            PathKind::BitmapHeapScan(_) => create_bitmap_scan_plan(root, subpath, outer),
            PathKind::IndexScan(_) => create_param_indexscan_plan(root, subpath, outer),
            PathKind::SeqScan => create_lateral_scan_plan(root, subpath, outer),
            PathKind::SubqueryScan(subplan) => {
                let (plan, layout) = create_subqueryscan_plan(root, subpath.parent, *subplan);
                (param_filtered(root, subpath, plan, &layout, outer), layout)
            }
            _ => create_plan_recurse(root, subpath),
        };
        plans.push(project(root, plan, layout, subpath).0);
    }
    let plan = plans
        .into_iter()
        .reduce(|left, right| Plan::SetOp {
            op: crate::plan::SetOp::Union,
            all: true,
            left: Box::new(left),
            right: Box::new(right),
        })
        .expect("an Append of members has paths");
    (plan, target_slots(root, &super::planner::path_exprs(root, path)))
}

/// create_modifytable_plan makes the plan of the rows that an UPDATE, DELETE, or MERGE changes, its subpath's plan
/// computing the subpath's target, whose last columns identify the result relation's rows, as Postgres' function of
/// the same name plans the subpath that its ModifyTable reads. The statement's executor in dml.rs is the ModifyTable.
fn create_modifytable_plan(root: &mut PlannerInfo<'_, '_>, subpath: &Path) -> (Plan, Vec<Slot>) {
    let (plan, layout) = create_plan_recurse(root, subpath);
    project(root, plan, layout, subpath)
}

/// create_projected_plan makes the plan of a path whose rows hold the columns of its target in order, as the inputs of
/// a set operation's Append, MergeAppend, SetOp, or RecursiveUnion must.
fn create_projected_plan(root: &mut PlannerInfo<'_, '_>, path: &Path) -> Plan {
    let (plan, layout) = create_plan_recurse(root, path);
    project(root, plan, layout, path).0
}

/// create_minmaxagg_plan makes the plan of the one row of a query's MIN and MAX aggregates, each the value of the
/// first row of its plan under a LIMIT of 1, read by an initplan, under the query's HAVING conditions, as Postgres'
/// function of the same name does.
fn create_minmaxagg_plan(
    root: &mut PlannerInfo<'_, '_>,
    path: &Path,
    minmax: &super::nodes::MinMaxAggPath,
) -> (Plan, Vec<Slot>) {
    let mut initplans = vec![None; root.parse.aggregates.len()];
    for info in &minmax.mmaggregates {
        let limit = Some(Expr::Const(Value::Int8(1)));
        let plan = Plan::Limit { input: Box::new(info.plan.clone()), limit, offset: None };
        initplans[info.agg] = Some(super::subselect::ss_make_initplan_from_plan(plan, info.pathcost));
    }
    let replace = |e: Expr| replace_minmax_aggs(e, &initplans);
    let exprs: Vec<Expr> = super::planner::path_exprs(root, path).into_iter().map(replace).collect();
    let layout = target_slots(root, &super::planner::path_exprs(root, path));
    let mut plan = Plan::OneRow;
    if let Some(predicate) =
        minmax.quals.iter().cloned().map(replace).reduce(|a, b| Expr::And(Box::new(a), Box::new(b)))
    {
        plan = Plan::Filter { input: Box::new(plan), predicate };
    }
    (Plan::Project { input: Box::new(plan), exprs }, layout)
}

/// replace_minmax_aggs rewrites each aggregate reference of an expression into the initplan that reads its value, as
/// Postgres' setrefs.c replaces it with find_minmax_agg_replacement_param.
fn replace_minmax_aggs(e: Expr, initplans: &[Option<Expr>]) -> Expr {
    match e {
        Expr::AggRef(k) => initplans[k].clone().expect("an initplan of each aggregate"),
        other => other.map_children(&mut |c| replace_minmax_aggs(c, initplans)),
    }
}

/// create_merge_append_plan makes the plan of a MergeAppend, sorting each input whose rows are not in the path's
/// order, as Postgres' function of the same name does, where a plan of one input is that input's plan, as Postgres'
/// set_mergeappend_references removes the MergeAppend.
fn create_merge_append_plan(root: &mut PlannerInfo<'_, '_>, path: &Path, subpaths: &[Rc<Path>]) -> (Plan, Vec<Slot>) {
    let layout = target_slots(root, &super::planner::path_exprs(root, path));
    let mut inputs = Vec::with_capacity(subpaths.len());
    for subpath in subpaths {
        let (plan, sub_layout) = create_plan_recurse(root, subpath);
        let plan = match super::pathkeys::pathkeys_contained_in(&path.pathkeys, &subpath.pathkeys) {
            true => plan,
            false => {
                let keys = sort_keys(root, &path.pathkeys, &sub_layout);
                Plan::Sort { input: Box::new(plan), keys }
            }
        };
        inputs.push(project(root, plan, sub_layout, subpath).0);
    }
    if inputs.len() == 1 {
        return (inputs.pop().expect("one input"), layout);
    }
    let keys = sort_keys(root, &path.pathkeys, &layout);
    (Plan::MergeAppend { inputs, keys }, layout)
}

/// tlist_exprs returns the expressions of the target entries that group clauses read.
fn tlist_exprs(root: &PlannerInfo<'_, '_>, clauses: &[super::nodes::SortGroupClause]) -> Vec<Expr> {
    super::tlist::get_sortgrouplist_exprs(clauses, &root.parse.target_list)
}

/// create_agg_plan makes the plan of an aggregation or GROUP BY over a path's rows: their group keys and the query's
/// aggregate calls over them that the rows do not hold already, under the HAVING conditions, computing the path's target, as Postgres' create_agg_plan
/// and create_group_plan do. Doltgres' aggregation keeps its groups in the order it first meets them, which is the
/// order of sorted rows. A GROUP BY whose keys are all redundant groups by a constant, so that it still returns no row
/// for no input, as Postgres' sorted aggregation of no keys does.
fn create_agg_plan(
    root: &mut PlannerInfo<'_, '_>,
    path: &Path,
    subpath: &Path,
    group_exprs: Vec<Expr>,
    qual: &[Expr],
    with_aggregates: bool,
    sets: Option<Vec<Vec<usize>>>,
) -> (Plan, Vec<Slot>) {
    let (plan, layout) = create_plan_recurse(root, subpath);
    let group_exprs = match group_exprs.is_empty() && !root.parse.group_clause.is_empty() && sets.is_none() {
        true => vec![Expr::Const(Value::Bool(true))],
        false => group_exprs,
    };
    let groups = group_exprs.iter().map(|g| positional(root, g.clone(), &layout)).collect();
    let computed: Vec<usize> = match with_aggregates {
        true => (0..root.parse.aggregates.len()).filter(|&k| !layout.contains(&Slot::Expr(Expr::AggRef(k)))).collect(),
        false => Vec::new(),
    };
    let mut aggregates: Vec<crate::functions::aggregate::AggCall> =
        computed.iter().map(|&k| agg_call(root, root.parse.aggregates[k].clone(), &layout)).collect();
    let mut agg_layout: Vec<Slot> = group_exprs.iter().map(|g| expr_slot(root, g)).collect();
    agg_layout.extend(computed.iter().map(|&k| Slot::Expr(Expr::AggRef(k))));
    for var in dependent_vars(root, path, qual, &agg_layout) {
        aggregates.push(min_call(root, &var, &layout));
        agg_layout.push(slot(root, &var));
    }
    let grouping_sets = sets.is_some();
    let mut plan = Plan::Aggregate { input: Box::new(plan), groups, aggregates, sets };
    if grouping_sets {
        agg_layout.push(Slot::Expr(Expr::Const(Value::Text("grouping mask".into()))));
    }
    let predicate = qual
        .iter()
        .map(|q| grouping_mask(positional(root, q.clone(), &agg_layout), &agg_layout))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    if let Some(predicate) = predicate {
        plan = Plan::Filter { input: Box::new(plan), predicate };
    }
    let (plan, slots) = project(root, plan, agg_layout.clone(), path);
    let plan = match plan {
        Plan::Project { input, exprs } => {
            Plan::Project { input, exprs: exprs.into_iter().map(|e| grouping_mask(e, &agg_layout)).collect() }
        }
        other => other,
    };
    (plan, slots)
}

/// create_groupingsets_plan makes the plan of an aggregation by grouping sets: one aggregation of every rollup's sets
/// over the query's group keys, which Doltgres' executor computes in one pass whether Postgres would sort or hash each
/// rollup, as Postgres' function of the same name makes a chain of them.
fn create_groupingsets_plan(
    root: &mut PlannerInfo<'_, '_>,
    path: &Path,
    gspath: &super::nodes::GroupingSetsPath,
) -> (Plan, Vec<Slot>) {
    let group_clause = root.processed_group_clause.clone();
    let position = |r: &usize| group_clause.iter().position(|gc| gc.tle_sort_group_ref == *r).expect("a group key");
    let sets = gspath
        .rollups
        .iter()
        .flat_map(|rollup| rollup.gsets_data.iter().map(|gs| gs.set.iter().map(position).collect()))
        .collect();
    let group_exprs = tlist_exprs(root, &group_clause);
    let with_aggregates = root.parse.has_aggs;
    create_agg_plan(root, path, &gspath.subpath, group_exprs, &gspath.qual, with_aggregates, Some(sets))
}

/// dependent_vars returns the Vars that an aggregation's target and conditions read outside its group keys and
/// aggregates, which a redundant group key that make_pathkeys_for_sortclauses_extended removed leaves: each is one value
/// in each group, which Postgres' Agg reads from the group's first row and Doltgres' as the group's MIN.
fn dependent_vars(root: &PlannerInfo<'_, '_>, path: &Path, qual: &[Expr], agg_layout: &[Slot]) -> Vec<Expr> {
    let mut vars: Vec<Expr> = Vec::new();
    for e in super::planner::path_exprs(root, path).iter().chain(qual) {
        if computable(root, e, agg_layout) {
            continue;
        }
        for id in super::var::pull_var_clause(root.glob, e, true) {
            let var = Expr::Column(id);
            if !agg_layout.contains(&slot(root, &var)) && !vars.contains(&var) {
                vars.push(var);
            }
        }
    }
    vars
}

/// min_call returns the MIN call over a layout's rows that gives each group's value of a Var that the group keys
/// determine, which Postgres' Agg reads from the group's first row.
fn min_call(root: &PlannerInfo<'_, '_>, var: &Expr, layout: &[Slot]) -> crate::functions::aggregate::AggCall {
    let ret = super::nodefuncs::expr_type(root, var).unwrap_or(crate::oid::UNKNOWN);
    let index = crate::functions::aggregate::AGGREGATES
        .iter()
        .position(|aggregate| aggregate.name == "min")
        .expect("the min aggregate");
    crate::functions::aggregate::AggCall {
        index,
        args: vec![positional(root, var.clone(), layout)],
        distinct: false,
        filter: None,
        order: Vec::new(),
        ret,
        user: None,
    }
}

/// grouping_mask points the GROUPING calls of an expression over an aggregation's rows at the column of the mask that
/// grouping sets add.
fn grouping_mask(e: Expr, layout: &[Slot]) -> Expr {
    let mask = layout.iter().position(|s| *s == Slot::Expr(Expr::Const(Value::Text("grouping mask".into()))));
    match e {
        Expr::Grouping(args, locations, _) => Expr::Grouping(args, locations, mask),
        other => other.map_children(&mut |c| grouping_mask(c, layout)),
    }
}

/// agg_call rewrites an aggregate call's expressions over a layout's rows.
fn agg_call(
    root: &PlannerInfo<'_, '_>,
    mut call: crate::functions::aggregate::AggCall,
    layout: &[Slot],
) -> crate::functions::aggregate::AggCall {
    call.args = call.args.into_iter().map(|a| positional(root, a, layout)).collect();
    call.filter = call.filter.map(|f| positional(root, f, layout));
    call.order = call.order.into_iter().map(|(e, d, n)| (positional(root, e, layout), d, n)).collect();
    call
}

/// window_call rewrites a window call's expressions over a layout's rows.
fn window_call(
    root: &PlannerInfo<'_, '_>,
    mut call: crate::window::WindowCall,
    layout: &[Slot],
) -> crate::window::WindowCall {
    call.args = call.args.into_iter().map(|a| positional(root, a, layout)).collect();
    call.filter = call.filter.map(|f| positional(root, f, layout));
    call.partition = call.partition.into_iter().map(|p| positional(root, p, layout)).collect();
    for key in &mut call.order {
        key.expr = positional(root, key.expr.clone(), layout);
    }
    for condition in &mut call.run_condition {
        condition.arg = positional(root, condition.arg.clone(), layout);
    }
    if let Some(range) = &mut call.range {
        range.key = positional(root, range.key.clone(), layout);
        for (e, _) in range.start.iter_mut().chain(range.end.iter_mut()) {
            *e = positional(root, e.clone(), layout);
        }
    }
    call
}

/// create_plan_recurse makes the plan of a path, returning it with what each column of its rows holds.
fn create_plan_recurse(root: &mut PlannerInfo<'_, '_>, path: &Path) -> (Plan, Vec<Slot>) {
    let (plan, layout) = match &path.kind {
        PathKind::SeqScan | PathKind::Lookup(_) if !path.param.is_empty() => create_lateral_scan_plan(root, path, &[]),
        PathKind::SeqScan | PathKind::Lookup(_) => create_scan_plan(root, path.parent, &[]),
        PathKind::SubqueryScan(subplan) => {
            let (plan, layout) = create_subqueryscan_plan(root, path.parent, *subplan);
            (param_filtered(root, path, plan, &layout, &[]), layout)
        }
        PathKind::Result(_) if path.pathtarget.is_some() => {
            let (plan, layout) = create_upper_plan(root, path);
            return (fix_alternative_subplans(root, plan, path.rows), layout);
        }
        PathKind::Result(quals) => {
            let quals = quals.iter().map(|q| positional(root, q.clone(), &[]));
            let predicate = quals.reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
            match predicate {
                Some(condition) => (Plan::OneTimeFilter { input: Box::new(Plan::OneRow), condition }, Vec::new()),
                None => (Plan::OneRow, Vec::new()),
            }
        }
        PathKind::IndexScan(_) if !path.param.is_empty() => create_param_indexscan_plan(root, path, &[]),
        PathKind::IndexScan(best_path) => create_indexscan_plan(root, path, best_path),
        PathKind::BitmapHeapScan(_) => create_bitmap_scan_plan(root, path, &[]),
        PathKind::BitmapAnd(_) | PathKind::BitmapOr(_) => unreachable!("a bitmap tree is planned by its heap scan"),
        PathKind::Projection(_)
        | PathKind::IncrementalSort(_)
        | PathKind::ProjectSet(_)
        | PathKind::Agg(_)
        | PathKind::GroupingSets(_)
        | PathKind::Group(_)
        | PathKind::Unique(..)
        | PathKind::WindowAgg(_)
        | PathKind::Limit(_) => return create_upper_plan(root, path),
        PathKind::Sort(_) => return create_upper_plan(root, path),
        PathKind::MergeAppend(_) | PathKind::SetOp(_) | PathKind::RecursiveUnion(_) | PathKind::MinMaxAgg(_) => {
            return create_upper_plan(root, path);
        }
        PathKind::Append(subpaths)
            if !subpaths.is_empty() && root.rels[path.parent].reloptkind != super::nodes::RelOptKind::BaseRel =>
        {
            return create_upper_plan(root, path);
        }
        PathKind::Append(subpaths) if !subpaths.is_empty() => create_append_plan(root, path, subpaths, &[]),
        PathKind::Append(_) => {
            let layout = target_slots(root, &super::planner::path_exprs(root, path));
            let nulls =
                Plan::Project { input: Box::new(Plan::OneRow), exprs: vec![Expr::Const(Value::Null); layout.len()] };
            (Plan::OneTimeFilter { input: Box::new(nulls), condition: Expr::Const(Value::Bool(false)) }, layout)
        }
        PathKind::Material(subpath) => return create_plan_recurse(root, subpath),
        PathKind::ModifyTable(subpath) => return create_modifytable_plan(root, subpath),
        PathKind::Memoize(mpath) => return create_plan_recurse(root, &mpath.subpath),
        PathKind::UniquePath(upath) => return create_unique_plan(root, path, upath),
        PathKind::NestLoop(_) | PathKind::HashJoin(_) | PathKind::MergeJoin(_) => {
            let join = path.kind.join().expect("a join path");
            let inner = match &join.inner.kind {
                PathKind::Memoize(mpath) => &mpath.subpath,
                _ => &join.inner,
            };
            let (mut outer_plan, mut outer_layout) = create_plan_recurse(root, &join.outer);
            let is_nestloop = matches!(path.kind, PathKind::NestLoop(_));
            let outerrelids = join.outer.relids.clone();
            let save_outer_rels = root.cur_outer_rels.clone();
            if is_nestloop {
                root.cur_outer_rels = save_outer_rels.union(&outerrelids);
            }
            let lookup = is_nestloop
                && !inner.param.is_empty()
                && match &inner.kind {
                    PathKind::IndexScan(ipath) => {
                        lookup_keys(root, inner.parent, ipath.index, &ipath.indexclauses).is_some()
                    }
                    PathKind::Lookup(_) => true,
                    _ => false,
                };
            let mut lateral = !inner.param.is_empty()
                && match &inner.kind {
                    PathKind::BitmapHeapScan(_) => true,
                    PathKind::SeqScan => root.parse.rte(inner.parent).lateral,
                    PathKind::IndexScan(ipath) => {
                        lookup_keys(root, inner.parent, ipath.index, &ipath.indexclauses).is_none()
                    }
                    PathKind::Append(subpaths) => !subpaths.is_empty(),
                    _ => false,
                };
            let (mut inner_plan, inner_layout) = match lateral {
                true => {
                    let (plan, layout) = match &inner.kind {
                        PathKind::BitmapHeapScan(_) => create_bitmap_scan_plan(root, inner, &outer_layout),
                        PathKind::SeqScan => create_lateral_scan_plan(root, inner, &outer_layout),
                        PathKind::Append(subpaths) => create_append_plan(root, inner, subpaths, &outer_layout),
                        _ => create_param_indexscan_plan(root, inner, &outer_layout),
                    };
                    let (mut plan, layout) = add_placeholders(root, inner.parent, plan, layout);
                    plan.map_exprs(0, &mut |e, depth| read_lateral_row(e, depth));
                    (plan, layout)
                }
                false if lookup => {
                    let (plan, layout) = create_scan_plan(root, inner.parent, &[]);
                    let (plan, layout) = add_placeholders(root, inner.parent, plan, layout);
                    (fix_alternative_subplans(root, plan, inner.rows), layout)
                }
                false => create_plan_recurse(root, inner),
            };
            root.cur_outer_rels = save_outer_rels;
            let nest_params = match is_nestloop {
                true => super::paramassign::identify_current_nestloop_params(root, &outerrelids, &path.param),
                false => Vec::new(),
            };
            for nlp in &nest_params {
                let Slot::PlaceHolder(phid) = slot(root, &nlp.paramval) else { continue };
                if outer_layout.contains(&Slot::PlaceHolder(phid)) {
                    continue;
                }
                let mut exprs: Vec<Expr> = (0..outer_layout.len()).map(Expr::Column).collect();
                exprs.push(positional(root, root.glob.placeholder(phid).phexpr.clone(), &outer_layout));
                outer_plan = Plan::Project { input: Box::new(outer_plan), exprs };
                outer_layout.push(Slot::PlaceHolder(phid));
            }
            if !nest_params.is_empty() {
                if !lateral {
                    inner_plan.map_exprs(0, &mut |e, depth| read_lateral_row(e, depth));
                    lateral = true;
                }
                let columns: Vec<(usize, usize)> = nest_params
                    .iter()
                    .map(|nlp| {
                        let slot = slot(root, &nlp.paramval);
                        let column = outer_layout.iter().position(|s| *s == slot).expect("the outer side supplies it");
                        (nlp.paramno, column)
                    })
                    .collect();
                inner_plan.map_exprs(0, &mut |e, depth| supply_nestloop_params(e, depth, &columns));
            }
            if let PathKind::Memoize(mpath) = &join.inner.kind {
                let keys = mpath.param_exprs.iter().map(|e| param_positional(root, e.clone(), &[], &outer_layout));
                let keys = keys.map(|e| read_lateral_row(e, 0)).collect();
                inner_plan = Plan::Memoize { input: Box::new(inner_plan), keys, binary: mpath.binary_mode };
            }
            if let PathKind::MergeJoin(mpath) = &path.kind {
                for (plan, sortkeys, layout) in [
                    (&mut outer_plan, &mpath.outersortkeys, &outer_layout),
                    (&mut inner_plan, &mpath.innersortkeys, &inner_layout),
                ] {
                    if !sortkeys.is_empty() {
                        let keys = sort_keys(root, sortkeys, layout);
                        *plan = Plan::Sort { input: Box::new(std::mem::replace(plan, Plan::OneRow)), keys };
                    }
                }
            }
            let layout = [outer_layout.as_slice(), inner_layout.as_slice()].concat();
            let (mut joinquals, otherquals) = match join.jointype.is_outer() {
                true => extract_actual_join_clauses(root, &join.joinrestrictinfo, &path.relids),
                false => (extract_actual_clauses(root, &join.joinrestrictinfo, false), Vec::new()),
            };
            if lookup
                && !lateral
                && let Some(ppi) = super::relnode::get_baserel_parampathinfo(root, inner.parent, &inner.param)
            {
                joinquals.extend(ppi.ppi_clauses.into_iter().filter(|r| !join.joinrestrictinfo.contains(r)));
            }
            let joinquals: Vec<Expr> = match &path.kind {
                PathKind::HashJoin(hpath) => {
                    let rest = joinquals.into_iter().filter(|r| !hpath.path_hashclauses.contains(r)).collect();
                    let mut quals = get_switched_clauses(root, &hpath.path_hashclauses, &join.outer.relids);
                    quals.extend(get_actual_clauses(root, &order_qual_clauses(root, rest)));
                    quals
                }
                PathKind::MergeJoin(mpath) => {
                    let rest = joinquals.into_iter().filter(|r| !mpath.path_mergeclauses.contains(r)).collect();
                    let mut quals = get_switched_clauses(root, &mpath.path_mergeclauses, &join.outer.relids);
                    quals.extend(get_actual_clauses(root, &order_qual_clauses(root, rest)));
                    quals
                }
                _ => get_actual_clauses(root, &order_qual_clauses(root, joinquals)),
            };
            let otherquals = order_qual_clauses(root, otherquals);
            let method = match (&path.kind, &inner.kind) {
                (PathKind::HashJoin(_), _) => JoinMethod::Hash,
                (PathKind::MergeJoin(mpath), _) => {
                    JoinMethod::Merge { clauses: mpath.path_mergeclauses.len(), materialized: mpath.materialize_inner }
                }
                (_, PathKind::IndexScan(best_path)) if !inner.param.is_empty() && !lateral => {
                    let rel = inner.parent;
                    let keys = lookup_keys(root, rel, best_path.index, &best_path.indexclauses)
                        .expect("a parameterized index path has lookup keys");
                    let table =
                        std::sync::Arc::new(root.parse.rte(rel).table().expect("an index path scans a table").clone());
                    let scan = crate::indexscan::IndexScan {
                        table,
                        index: root.rels[rel].indexlist[best_path.index].index,
                        ranges: Vec::new(),
                        reverse: false,
                        nearest: None,
                        needed: None,
                        lookup_heavy: None,
                        parameterized: None,
                    };
                    let keys = keys.into_iter().map(|k| positional(root, k, &outer_layout)).collect();
                    JoinMethod::Lookup { scan: Box::new(scan), keys }
                }
                (_, PathKind::Lookup(JoinMethod::CatalogLookup { index, keys })) => JoinMethod::CatalogLookup {
                    index,
                    keys: keys.iter().map(|k| positional(root, k.clone(), &outer_layout)).collect(),
                },
                (_, PathKind::Material(_)) => JoinMethod::MaterializedLoop,
                _ => JoinMethod::NestedLoop,
            };
            let kind = match join.jointype {
                JoinType::Inner => JoinKind::Inner,
                JoinType::Left => JoinKind::Left,
                JoinType::Right => JoinKind::Right,
                JoinType::Full => JoinKind::Full,
                JoinType::Semi => JoinKind::Semi,
                JoinType::Anti => JoinKind::Anti,
                JoinType::RightSemi => JoinKind::RightSemi,
                JoinType::RightAnti => JoinKind::RightAnti,
                JoinType::UniqueOuter | JoinType::UniqueInner => {
                    unreachable!("a semi join made unique is planned as an inner join")
                }
            };
            let condition = joinquals
                .into_iter()
                .map(|c| positional(root, c, &layout))
                .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
            let plan = Plan::Join {
                left: Box::new(outer_plan),
                right: Box::new(inner_plan),
                kind,
                condition,
                lateral,
                method,
            };
            let mut layout = layout;
            let dropped = match kind {
                JoinKind::Semi | JoinKind::Anti => outer_layout.len()..layout.len(),
                JoinKind::RightSemi | JoinKind::RightAnti => 0..outer_layout.len(),
                _ => 0..0,
            };
            layout[dropped].fill(Slot::Expr(Expr::Const(Value::Null)));
            (filtered(root, plan, &otherquals, &layout), layout)
        }
    };
    let gating_clauses = match &path.kind {
        PathKind::Result(_) | PathKind::Append(_) => Vec::new(),
        kind => match kind.join() {
            Some(join) => get_gating_quals(root, join.joinrestrictinfo.clone()),
            None => get_gating_quals(root, root.rels[path.parent].baserestrictinfo.clone()),
        },
    };
    let plan = match gating_clauses.is_empty() {
        true => plan,
        false => create_gating_plan(root, plan, &gating_clauses),
    };
    let (plan, layout) = add_placeholders(root, path.parent, plan, layout);
    (fix_alternative_subplans(root, plan, path.rows), layout)
}

/// create_unique_plan makes the plan of a semi join's inner relation made unique by the inner values of the join's
/// equalities, as Postgres' function of the same name does: the rows themselves when they are unique already, a hashed
/// aggregation by those values, or a Unique over a sort by them. A hashed aggregation that would drop a column the
/// relation's target needs is a hashed DISTINCT ON those values instead.
fn create_unique_plan(
    root: &mut PlannerInfo<'_, '_>,
    path: &Path,
    upath: &super::nodes::UniquePath,
) -> (Plan, Vec<Slot>) {
    let (plan, layout) = create_plan_recurse(root, &upath.subpath);
    let groups: Vec<Expr> = upath.uniq_exprs.iter().map(|e| positional(root, e.clone(), &layout)).collect();
    match upath.umethod {
        super::nodes::UniquePathMethod::Noop => (plan, layout),
        super::nodes::UniquePathMethod::Hash => {
            let mut group_layout: Vec<Slot> = upath.uniq_exprs.iter().map(|e| expr_slot(root, e)).collect();
            let mut aggregates = Vec::new();
            for var in root.rels[path.parent].reltarget.exprs.clone() {
                if group_layout.contains(&slot(root, &var)) {
                    continue;
                }
                aggregates.push(min_call(root, &var, &layout));
                group_layout.push(slot(root, &var));
            }
            (Plan::Aggregate { input: Box::new(plan), groups, aggregates, sets: None }, group_layout)
        }
        super::nodes::UniquePathMethod::Sort => {
            let keys = groups
                .iter()
                .map(|e| crate::plan::SortKey { expr: e.clone(), descending: false, nulls_first: false })
                .collect();
            let sorted = Plan::Sort { input: Box::new(plan), keys };
            (Plan::Distinct { input: Box::new(sorted), keys: Some(groups) }, layout)
        }
    }
}

/// get_gating_quals returns the pseudoconstant clauses of a node's quals in the order to test them, as Postgres'
/// function of the same name does.
fn get_gating_quals(root: &PlannerInfo<'_, '_>, quals: Vec<RinfoId>) -> Vec<RinfoId> {
    if !root.has_pseudo_constant_quals {
        return Vec::new();
    }
    extract_actual_clauses(root, &order_qual_clauses(root, quals), true)
}

/// create_gating_plan returns a plan under a Result that tests pseudoconstant clauses once, as Postgres' function of
/// the same name does.
fn create_gating_plan(root: &PlannerInfo<'_, '_>, plan: Plan, gating_quals: &[RinfoId]) -> Plan {
    let condition = gating_quals
        .iter()
        .map(|&r| positional(root, root.rinfos[r].clause.clone(), &[]))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)))
        .expect("a gating plan has clauses");
    Plan::OneTimeFilter { input: Box::new(plan), condition }
}

/// fix_alternative_subplans replaces each AlternativeSubPlan of a path's plan by the SubPlan that costs least for the
/// rows the plan returns, as Postgres' set_plan_refs does, which runs a target list once for each row and a condition
/// twice.
fn fix_alternative_subplans(root: &PlannerInfo<'_, '_>, mut plan: Plan, rows: f64) -> Plan {
    if !root.has_alternative_subplans {
        return plan;
    }
    match &mut plan {
        Plan::Project { input, exprs } => {
            for e in exprs.iter_mut() {
                *e = fix_alternative_subplan(std::mem::replace(e, Expr::SubqueryValue), rows);
            }
            input.map_exprs(0, &mut |e, _| fix_alternative_subplan(e, rows * 2.0));
        }
        other => {
            other.map_exprs(0, &mut |e, _| fix_alternative_subplan(e, rows * 2.0));
        }
    }
    plan
}

/// fix_alternative_subplan replaces each AlternativeSubPlan of an expression by the SubPlan that costs least for an
/// estimated number of runs, the later one of equal costs, as Postgres' function of the same name does.
fn fix_alternative_subplan(e: Expr, num_exec: f64) -> Expr {
    match e {
        Expr::AlternativeSubPlan(subplans) => {
            let cost = |s: &crate::expr::SubPlan| s.startup_cost + num_exec * s.per_call_cost;
            let best = subplans.into_iter().reduce(|best, cur| if cost(&cur) <= cost(&best) { cur } else { best });
            Expr::SubPlan(Box::new(best.expect("an alternative")))
        }
        other => other.map_children(&mut |c| fix_alternative_subplan(c, num_exec)),
    }
}

/// add_placeholders computes the PlaceHolderVars of a relation's target that its plan's rows do not hold yet, after
/// its columns, as Postgres evaluates them in the target list of the plan where they are evaluated.
fn add_placeholders(root: &PlannerInfo<'_, '_>, rel: usize, plan: Plan, mut layout: Vec<Slot>) -> (Plan, Vec<Slot>) {
    let width = layout.len();
    let mut exprs: Vec<Expr> = (0..width).map(Expr::Column).collect();
    for e in &root.rels[rel].reltarget.exprs {
        let Expr::Column(id) = e else { continue };
        let VarNode::PlaceHolderVar(phv) = root.glob.node(*id) else { continue };
        if layout.contains(&Slot::PlaceHolder(phv.phid)) {
            continue;
        }
        exprs.push(positional(root, root.glob.placeholder(phv.phid).phexpr.clone(), &layout));
        layout.push(Slot::PlaceHolder(phv.phid));
    }
    match exprs.len() == width {
        true => (plan, layout),
        false => (Plan::Project { input: Box::new(plan), exprs }, layout),
    }
}

/// create_scan_plan makes the plan that reads a base relation's rows and tests its restrictions: a scan of its table
/// under a filter, the one row of a RESULT relation, its set-returning calls, VALUES list, table function, or WITH
/// query under a filter, as Postgres' create_functionscan_plan and its kin make them, or its own plan with the
/// restrictions pushed into it, where an index of a system catalog may answer them.
fn create_scan_plan(root: &mut PlannerInfo<'_, '_>, rel: usize, outer: &[Slot]) -> (Plan, Vec<Slot>) {
    let restrictinfo = order_qual_clauses(root, extract_actual_clauses(root, &root.rels[rel].baserestrictinfo, false));
    let lateral = root.parse.rte(rel).lateral;
    match root.parse.rte(rel).kind.clone() {
        RteKind::Relation(plan, _) => {
            let layout = base_slots(rel, plan.width());
            (filtered(root, plan, &restrictinfo, &layout), layout)
        }
        RteKind::Result => (filtered(root, Plan::OneRow, &restrictinfo, &[]), Vec::new()),
        RteKind::Plan(mut plan) => {
            if lateral {
                plan.map_exprs(0, &mut |e, depth| match depth {
                    0 => param_positional(root, e, &[], outer),
                    _ => e,
                });
            }
            let layout = base_slots(rel, plan.width());
            if matches!(
                plan,
                Plan::Function { .. }
                    | Plan::RowsFrom { .. }
                    | Plan::Values(_)
                    | Plan::XmlTable(_)
                    | Plan::JsonTable(_)
                    | Plan::CteScan(_)
                    | Plan::WorkTable(..)
            ) {
                return (filtered(root, plan, &restrictinfo, &layout), layout);
            }
            let predicate = restrictinfo
                .iter()
                .map(|&r| to_attnos(root, root.rinfos[r].clause.clone(), rel))
                .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
            let plan = match predicate {
                Some(predicate) => crate::plan::Planner { ctx: root.ctx.get_mut(), outer: Vec::new() }
                    .use_indexes(crate::plan::push_down(plan, predicate)),
                None => plan,
            };
            (plan, layout)
        }
        RteKind::Subquery(..) | RteKind::Join(_) => unreachable!("only base relations are scanned"),
    }
}

/// create_lateral_scan_plan makes the plan of a lateral function or VALUES list on the inner side of a nested loop,
/// testing the join clauses that its parameterization moved into it, as Postgres' create_scan_plan adds a
/// parameterized path's clauses. It reads the columns of the outer rows of the given layout as `Expr::Outer(0, _)`.
fn create_lateral_scan_plan(root: &mut PlannerInfo<'_, '_>, path: &Path, outer: &[Slot]) -> (Plan, Vec<Slot>) {
    let (plan, layout) = create_scan_plan(root, path.parent, outer);
    (param_filtered(root, path, plan, &layout, outer), layout)
}

/// param_filtered returns a plan of a parameterized base relation path's rows under the join clauses that its
/// parameterization moved into it, as Postgres' create_scan_plan adds them, reading the columns of the outer rows of
/// the given layout as `Expr::Outer(0, _)` and those of the nested loops further out as nestloop parameters.
fn param_filtered(root: &mut PlannerInfo<'_, '_>, path: &Path, plan: Plan, layout: &[Slot], outer: &[Slot]) -> Plan {
    let ppi_clauses = match super::relnode::get_baserel_parampathinfo(root, path.parent, &path.param) {
        Some(ppi) => order_qual_clauses(root, extract_actual_clauses(root, &ppi.ppi_clauses, false)),
        None => Vec::new(),
    };
    let predicate = ppi_clauses
        .iter()
        .map(|&r| param_positional(root, root.rinfos[r].clause.clone(), layout, outer))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    match predicate {
        Some(predicate) => Plan::Filter { input: Box::new(plan), predicate },
        None => plan,
    }
}

/// create_subqueryscan_plan makes the plan that reads a subquery relation's rows from the plan of one of its
/// subquery's final paths and tests the restrictions that the subquery did not take, as Postgres' function of the
/// same name does.
fn create_subqueryscan_plan(root: &mut PlannerInfo<'_, '_>, rel: usize, subplan: usize) -> (Plan, Vec<Slot>) {
    let scan_clauses = order_qual_clauses(root, extract_actual_clauses(root, &root.rels[rel].baserestrictinfo, false));
    let mut plan = root.rels[rel].subplans[subplan].plan.clone();
    let lateral_columns = root.parse.rte(rel).lateral_columns.clone();
    if root.parse.rte(rel).lateral {
        plan.map_exprs(0, &mut |e, depth| read_lateral_columns(root, e, depth, &lateral_columns));
    }
    let layout = base_slots(rel, plan.width());
    (filtered(root, plan, &scan_clauses, &layout), layout)
}

/// read_lateral_columns rewrites an expression of the plan of a lateral subquery, at a depth of enclosing rows within
/// it, to read the expressions of the query around it that the columns of its enclosing row hold, as nestloop
/// parameters, and the rows further out one row closer, as Postgres' process_subquery_nestloop_params turns the
/// subquery's outer parameters into nestloop parameters.
fn read_lateral_columns(root: &PlannerInfo<'_, '_>, e: Expr, depth: usize, lateral_columns: &[Expr]) -> Expr {
    let mut e = match e {
        Expr::Outer(d, c) if d == depth + 1 => return positional(root, lateral_columns[c].clone(), &[]),
        Expr::Outer(d, c) if d > depth + 1 => return Expr::Outer(d - 1, c),
        other => other.map_children(&mut |x| read_lateral_columns(root, x, depth, lateral_columns)),
    };
    for p in e.subqueries_mut() {
        p.map_exprs(0, &mut |x, d| read_lateral_columns(root, x, depth + 1 + d, lateral_columns));
    }
    e
}

/// create_indexscan_plan makes the plan of a scan of an index, reading the ranges that its index clauses give and
/// testing the restrictions that those ranges do not answer exactly, as Postgres' function of the same name does. A
/// scan of every entry of the primary index in its order is the table's sequential scan, index clauses that read values
/// known only at run time build their ranges then, and a vector index's scan in the order of a distance is its search
/// for the query's LIMIT and OFFSET closest rows.
fn create_indexscan_plan(root: &mut PlannerInfo<'_, '_>, path: &Path, best_path: &IndexPath) -> (Plan, Vec<Slot>) {
    let rel = path.parent;
    let table = std::sync::Arc::new(root.parse.rte(rel).table().expect("an index path scans a table").clone());
    let info = &root.rels[rel].indexlist[best_path.index];
    let (index, scan_clauses) = (info.index, info.indrestrictinfo.clone());
    if let [Expr::Operator(_, _, leftop, rightop)] = best_path.indexorderbys.as_slice() {
        let order = to_attnos(root, best_path.indexorderbys[0].clone(), rel);
        let query = match crate::indexscan::is_query_vector(leftop) {
            true => (**leftop).clone(),
            false => (**rightop).clone(),
        };
        let nearest = crate::indexscan::Nearest {
            order,
            query,
            limit: root.parse.limit_count.clone(),
            offset: root.parse.limit_offset.clone(),
        };
        let scan = crate::indexscan::IndexScan {
            table: table.clone(),
            index,
            ranges: Vec::new(),
            reverse: false,
            nearest: Some(nearest),
            needed: None,
            lookup_heavy: None,
            parameterized: None,
        };
        return (Plan::IndexScan(Box::new(scan)), base_slots(rel, table.columns.len()));
    }
    if index.is_none() && best_path.indexclauses.is_empty() && !best_path.backward {
        return create_scan_plan(root, rel, &[]);
    }
    let indexquals = best_path.indexclauses.iter().flat_map(|iclause| &iclause.indexquals);
    let predicate = indexquals
        .map(|&r| to_attnos(root, root.rinfos[r].clause.clone(), rel))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let scan_of = |root: &mut PlannerInfo<'_, '_>, predicate: Option<&Expr>| {
        crate::indexscan::scan_of_index(root.ctx.get_mut(), &table, index, predicate, best_path.backward)
    };
    let (scan, exact) = match scan_of(root, predicate.as_ref()) {
        Some(found) => found,
        None if predicate.as_ref().is_some_and(reads_run_time_values) => {
            return create_param_indexscan_plan(root, path, &[]);
        }
        None => (scan_of(root, None).expect("a scan of every entry is always possible").0, false),
    };
    let qpqual: Vec<RinfoId> = extract_actual_clauses(root, &scan_clauses, false)
        .into_iter()
        .filter(|&r| !exact || !best_path.indexclauses.iter().any(|iclause| iclause.rinfo == r && !iclause.lossy))
        .collect();
    let qpqual = order_qual_clauses(root, qpqual);
    let layout = base_slots(rel, table.columns.len());
    (filtered(root, Plan::IndexScan(Box::new(scan)), &qpqual, &layout), layout)
}

/// reads_run_time_values reports whether an expression reads a value known only when its plan runs: an enclosing row's
/// column or a statement parameter, as Postgres' Params are.
fn reads_run_time_values(e: &Expr) -> bool {
    let mut reads = false;
    e.visit(&mut |x| reads |= matches!(x, Expr::Outer(..) | Expr::Param(_)));
    reads
}

/// create_param_indexscan_plan makes the plan of a scan of an index parameterized by outer relations whose join
/// clauses Doltgres' lookups cannot search by, the inner side of a lateral nested loop, as create_indexscan_plan does
/// for such a path: an index scan that builds its ranges from its index conditions over each outer row, under the
/// restrictions and parameterizing join clauses that those conditions do not give. It reads the columns of the outer
/// rows of the given layout as `Expr::Outer(0, _)`, which `read_lateral_row` turns into reads of the enclosing row.
fn create_param_indexscan_plan(root: &mut PlannerInfo<'_, '_>, path: &Path, outer: &[Slot]) -> (Plan, Vec<Slot>) {
    let PathKind::IndexScan(best_path) = &path.kind else { unreachable!("an index scan path") };
    let rel = path.parent;
    let table = std::sync::Arc::new(root.parse.rte(rel).table().expect("an index path scans a table").clone());
    let index = root.rels[rel].indexlist[best_path.index].index;
    let layout = base_slots(rel, table.columns.len());
    let indexquals: Vec<RinfoId> =
        best_path.indexclauses.iter().flat_map(|iclause| iclause.indexquals.iter().copied()).collect();
    let indexquals = get_actual_clauses(root, &indexquals);
    let to_row = |root: &PlannerInfo<'_, '_>, e: Expr| param_positional(root, e, &layout, outer);
    let cond = indexquals.iter().map(|e| to_row(root, e.clone())).reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let every = crate::indexscan::scan_of_index(root.ctx.get_mut(), &table, index, None, best_path.backward)
        .expect("a scan of every entry")
        .0;
    let scan = crate::indexscan::IndexScan { parameterized: cond, ..every };
    let mut scan_clauses = root.rels[rel].baserestrictinfo.clone();
    if let Some(ppi) = super::relnode::get_baserel_parampathinfo(root, rel, &path.param) {
        scan_clauses.extend(ppi.ppi_clauses);
    }
    let qpqual: Vec<RinfoId> = scan_clauses
        .into_iter()
        .filter(|&r| {
            let rinfo = &root.rinfos[r];
            !rinfo.pseudoconstant
                && !super::equivclass::is_redundant_with_indexclauses(root, r, &best_path.indexclauses)
                && !implied_by_indexquals(root, &rinfo.clause, &indexquals)
        })
        .collect();
    let qpqual = order_qual_clauses(root, qpqual);
    let predicate = qpqual
        .iter()
        .map(|&r| to_row(root, root.rinfos[r].clause.clone()))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let plan = Plan::IndexScan(Box::new(scan));
    match predicate {
        Some(predicate) => (Plan::Filter { input: Box::new(plan), predicate }, layout),
        None => (plan, layout),
    }
}

/// create_bitmap_scan_plan makes the plan of a scan of a base relation's rows whose keys a tree of index scans finds,
/// as Postgres' function of the same name does: a bitmap heap scan under the restrictions and parameterizing join
/// clauses that the tree's index conditions do not give, which rechecks on each row the clauses that the tree answers
/// and the filter does not test when an index scan's ranges keep keys that its conditions do not. A parameterized
/// scan reads the columns of the outer rows of the given layout as `Expr::Outer(0, _)`, which `read_lateral_row` turns
/// into reads of the enclosing row.
fn create_bitmap_scan_plan(root: &mut PlannerInfo<'_, '_>, path: &Path, outer: &[Slot]) -> (Plan, Vec<Slot>) {
    let PathKind::BitmapHeapScan(bitmapqual) = &path.kind else { unreachable!("a bitmap heap scan path") };
    let rel = path.parent;
    let table = std::sync::Arc::new(root.parse.rte(rel).table().expect("a bitmap scan reads a table").clone());
    let mut exact = true;
    let subplan = create_bitmap_subplan(root, rel, &table, bitmapqual, outer, &mut exact);
    let (bitmap, bitmapqualorig, indexquals, index_ecs) = subplan;
    let mut scan_clauses = root.rels[rel].baserestrictinfo.clone();
    if let Some(ppi) = super::relnode::get_baserel_parampathinfo(root, rel, &path.param) {
        scan_clauses.extend(ppi.ppi_clauses);
    }
    let qpqual: Vec<RinfoId> = scan_clauses
        .into_iter()
        .filter(|&r| {
            let rinfo = &root.rinfos[r];
            !rinfo.pseudoconstant
                && !implied_by_indexquals(root, &rinfo.clause, &indexquals)
                && !rinfo.parent_ec.is_some_and(|ec| index_ecs.contains(&ec))
        })
        .collect();
    let qpqual = order_qual_clauses(root, qpqual);
    let layout = base_slots(rel, table.columns.len());
    let to_row = |root: &PlannerInfo<'_, '_>, e: Expr| param_positional(root, e, &layout, outer);
    let recheck = bitmapqualorig
        .into_iter()
        .filter(|q| !qpqual.iter().any(|&r| same_clause(q, &root.rinfos[r].clause)))
        .map(|e| to_row(root, e))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let scan = crate::indexscan::BitmapHeapScan { table, bitmap, recheck, lossy: !exact, needed: None };
    let predicate = qpqual
        .iter()
        .map(|&r| to_row(root, root.rinfos[r].clause.clone()))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let plan = Plan::BitmapHeapScan(Box::new(scan));
    match predicate {
        Some(predicate) => (Plan::Filter { input: Box::new(plan), predicate }, layout),
        None => (plan, layout),
    }
}

/// BitmapSubplan is the tree of index scans of a bitmap path, with the clauses it answers, the index conditions that it
/// searches by, and the equivalence classes that those conditions come from.
type BitmapSubplan = (crate::indexscan::Bitmap, Vec<Expr>, Vec<Expr>, Vec<super::nodes::EcId>);

/// create_bitmap_subplan makes the tree of index scans of a bitmap path, with the clauses it answers, the index
/// conditions that it searches by, and their equivalence classes, clearing `exact` when an index scan's ranges keep
/// keys that its conditions do not, as Postgres' function of the same name does. An index scan whose conditions read
/// outer rows keeps them, over the outer rows of the given layout, to build its ranges each time it runs.
fn create_bitmap_subplan(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    table: &std::sync::Arc<crate::catalog::table::TableDef>,
    bitmapqual: &Path,
    outer: &[Slot],
    exact: &mut bool,
) -> BitmapSubplan {
    use crate::indexscan::Bitmap;
    let and = |quals: Vec<Expr>| quals.into_iter().reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    match &bitmapqual.kind {
        PathKind::BitmapAnd(bpath) => {
            let (mut children, mut quals, mut indexquals, mut index_ecs) =
                (Vec::new(), Vec::new(), Vec::new(), Vec::new());
            for subpath in &bpath.bitmapquals {
                let (child, subqual, subindexqual, subindex_ecs) =
                    create_bitmap_subplan(root, rel, table, subpath, outer, exact);
                children.push(child);
                for q in subqual {
                    if !quals.contains(&q) {
                        quals.push(q);
                    }
                }
                for q in subindexqual {
                    if !indexquals.contains(&q) {
                        indexquals.push(q);
                    }
                }
                index_ecs.extend(subindex_ecs);
            }
            (Bitmap::And(children), quals, indexquals, index_ecs)
        }
        PathKind::BitmapOr(bpath) => {
            let (mut children, mut subquals, mut subindexquals) = (Vec::new(), Vec::new(), Vec::new());
            for subpath in &bpath.bitmapquals {
                let (child, subqual, subindexqual, _) = create_bitmap_subplan(root, rel, table, subpath, outer, exact);
                children.push(child);
                subquals.push(and(subqual));
                subindexquals.push(and(subindexqual));
            }
            let or = |quals: Vec<Option<Expr>>| -> Vec<Expr> {
                let quals: Option<Vec<Expr>> = quals.into_iter().collect();
                quals
                    .and_then(|quals| quals.into_iter().reduce(|a, b| Expr::Or(Box::new(a), Box::new(b))))
                    .into_iter()
                    .collect()
            };
            (Bitmap::Or(children), or(subquals), or(subindexquals), Vec::new())
        }
        PathKind::IndexScan(ipath) => {
            let info = &root.rels[rel].indexlist[ipath.index];
            let index = info.index;
            let mut quals: Vec<Expr> =
                ipath.indexclauses.iter().map(|iclause| root.rinfos[iclause.rinfo].clause.clone()).collect();
            for pred in info.indpred.clone() {
                if !quals.contains(&pred) {
                    quals.push(pred);
                }
            }
            let indexquals: Vec<RinfoId> =
                ipath.indexclauses.iter().flat_map(|iclause| iclause.indexquals.iter().copied()).collect();
            let indexquals = get_actual_clauses(root, &indexquals);
            let index_ecs: Vec<super::nodes::EcId> =
                ipath.indexclauses.iter().filter_map(|iclause| root.rinfos[iclause.rinfo].parent_ec).collect();
            let every = |root: &mut PlannerInfo<'_, '_>| {
                crate::indexscan::scan_of_index(root.ctx.get_mut(), table, index, None, false)
                    .expect("a scan of every entry")
                    .0
            };
            if !bitmapqual.param.is_empty() {
                let layout = base_slots(rel, table.columns.len());
                let cond = and(indexquals.iter().map(|e| param_positional(root, e.clone(), &layout, outer)).collect());
                let scan = crate::indexscan::IndexScan { parameterized: cond, ..every(root) };
                return (Bitmap::Index(Box::new(scan)), quals, indexquals, index_ecs);
            }
            let predicate = and(indexquals.iter().map(|e| to_attnos(root, e.clone(), rel)).collect());
            let scan =
                match crate::indexscan::scan_of_index(root.ctx.get_mut(), table, index, predicate.as_ref(), false) {
                    Some((scan, covered)) => {
                        *exact &= covered;
                        scan
                    }
                    None => {
                        *exact = false;
                        every(root)
                    }
                };
            (Bitmap::Index(Box::new(scan)), quals, indexquals, index_ecs)
        }
        _ => unreachable!("a bitmap tree holds index scans, BitmapAnds, and BitmapOrs"),
    }
}

/// implied_by_indexquals reports whether index conditions answer a clause: it is one of them, or it calls only
/// immutable functions and they imply it, as Postgres' create_indexscan_plan and create_bitmap_scan_plan test.
fn implied_by_indexquals(root: &PlannerInfo<'_, '_>, clause: &Expr, indexquals: &[Expr]) -> bool {
    indexquals.iter().any(|q| same_clause(q, clause))
        || !super::clauses::contain_mutable_functions(root.glob, clause)
            && super::predtest::predicate_implied_by(root, std::slice::from_ref(clause), indexquals, false)
}

/// same_clause reports whether two clauses are the same, as ANDs and ORs of the same arguments in any order are.
fn same_clause(a: &Expr, b: &Expr) -> bool {
    use super::restrictinfo::{and_args, or_args};
    let same_args =
        |x: Vec<&Expr>, y: Vec<&Expr>| x.len() == y.len() && x.iter().all(|e| y.iter().any(|f| same_clause(e, f)));
    match (a, b) {
        _ if a == b => true,
        (Expr::Or(..), Expr::Or(..)) => same_args(or_args(a), or_args(b)),
        (Expr::And(..), Expr::And(..)) => same_args(and_args(a), and_args(b)),
        _ => a == b,
    }
}

/// param_positional rewrites an expression over Vars and PlaceHolderVars into one over rows of a layout's slots, as
/// `positional` does, where the slots of the outer rows of another layout are read as `Expr::Outer(0, _)`.
fn param_positional(root: &PlannerInfo<'_, '_>, e: Expr, layout: &[Slot], outer: &[Slot]) -> Expr {
    match e {
        Expr::Column(id) => {
            let target = slot(root, &Expr::Column(id));
            match outer.iter().position(|s| *s == target) {
                Some(i) if !layout.contains(&target) => Expr::Outer(0, i),
                _ => positional(root, Expr::Column(id), layout),
            }
        }
        other => other.map_children(&mut |c| param_positional(root, c, layout, outer)),
    }
}

/// read_lateral_row rewrites an expression of the inner plan of a lateral join, at a depth of subqueries within it,
/// for the outer row that the join pushes as the enclosing row: a read of the outer row as `Expr::Outer(0, _)` reads
/// that enclosing row, and a read of an enclosing row reads it one row further out.
fn read_lateral_row(e: Expr, depth: usize) -> Expr {
    let mut e = match e {
        Expr::Outer(0, i) if depth == 0 => return Expr::Outer(1, i),
        Expr::Outer(d, i) if d > depth => return Expr::Outer(d + 1, i),
        other => other.map_children(&mut |c| read_lateral_row(c, depth)),
    };
    for p in e.subqueries_mut() {
        p.map_exprs(0, &mut |x, d| read_lateral_row(x, depth + 1 + d));
    }
    e
}

/// supply_nestloop_params replaces the nestloop parameters that a lateral nested loop supplies, by their numbers and
/// the columns of its outer rows that hold them, in an expression of its inner plan at a depth of enclosing rows
/// within it, by reads of the outer row that the nested loop pushes as an enclosing row, as Postgres' nestloop's
/// nestParams set them.
fn supply_nestloop_params(e: Expr, depth: usize, columns: &[(usize, usize)]) -> Expr {
    let mut e = match e {
        Expr::NestLoopParam(paramno) => match columns.iter().find(|(p, _)| *p == paramno) {
            Some(&(_, column)) => Expr::Outer(depth + 1, column),
            None => e,
        },
        other => other.map_children(&mut |c| supply_nestloop_params(c, depth, columns)),
    };
    for p in e.subqueries_mut() {
        p.map_exprs(0, &mut |x, d| supply_nestloop_params(x, depth + 1 + d, columns));
    }
    e
}

/// order_qual_clauses sorts clauses by security level and then by the cost of evaluating them, cheapest first and
/// otherwise in their order, as Postgres' function of the same name does. Clauses are never leakproof, since nothing
/// sets a security level above 0 to need it.
fn order_qual_clauses(root: &PlannerInfo<'_, '_>, clauses: Vec<RinfoId>) -> Vec<RinfoId> {
    let mut items: Vec<(RinfoId, f64, usize)> = clauses
        .into_iter()
        .map(|r| (r, cost_qual_eval_node(&root.rinfos[r].clause).per_tuple, root.rinfos[r].security_level))
        .collect();
    items.sort_by(|a, b| a.2.cmp(&b.2).then(a.1.total_cmp(&b.1)));
    items.into_iter().map(|(r, ..)| r).collect()
}

/// get_switched_clauses returns the clauses of hash or merge clauses with each one's outer side first, and records
/// which side that was, as Postgres' function of the same name does.
fn get_switched_clauses(
    root: &PlannerInfo<'_, '_>,
    clauses: &[RinfoId],
    outer_relids: &super::nodes::Relids,
) -> Vec<Expr> {
    clauses
        .iter()
        .map(|&r| {
            let r = &root.rinfos[r];
            let switched = r.right_relids.is_subset(outer_relids);
            r.outer_is_left.set(!switched);
            match &r.clause {
                Expr::Compare(op, left, right) if switched => {
                    Expr::Compare(crate::indexscan::swap(*op), right.clone(), left.clone())
                }
                other => other.clone(),
            }
        })
        .collect()
}

/// filtered returns a plan under a filter of clauses over its rows, or the plan itself without clauses.
fn filtered(root: &PlannerInfo<'_, '_>, plan: Plan, clauses: &[RinfoId], layout: &[Slot]) -> Plan {
    let predicate = clauses
        .iter()
        .map(|&r| positional(root, root.rinfos[r].clause.clone(), layout))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    match predicate {
        Some(predicate) => Plan::Filter { input: Box::new(plan), predicate },
        None => plan,
    }
}

/// base_slots returns the slots of a base relation's columns in order.
fn base_slots(rel: usize, width: usize) -> Vec<Slot> {
    (0..width).map(|attno| Slot::Var(rel, attno)).collect()
}

/// slot returns the slot of a Var or PlaceHolderVar.
fn slot(root: &PlannerInfo<'_, '_>, e: &Expr) -> Slot {
    let Expr::Column(id) = e else { unreachable!("a relation's target holds Vars") };
    match root.glob.node(*id) {
        VarNode::Var(var) => Slot::Var(var.varno, var.varattno),
        VarNode::PlaceHolderVar(phv) => Slot::PlaceHolder(phv.phid),
    }
}

/// positional rewrites an expression over Vars and PlaceHolderVars into one over rows of a layout's slots, as
/// Postgres' setrefs.c rewrites a plan's Vars to refer to its inputs' columns. A Var or PlaceHolderVar that the rows
/// do not hold and the outer side of a nested loop above supplies becomes a nestloop parameter, as Postgres'
/// replace_nestloop_params does, and any other PlaceHolderVar that the rows do not hold is computed from its
/// expression.
fn positional(root: &PlannerInfo<'_, '_>, e: Expr, layout: &[Slot]) -> Expr {
    if !matches!(e, Expr::Column(_))
        && let Some(i) = layout.iter().position(|s| matches!(s, Slot::Expr(x) if *x == e))
    {
        return Expr::Column(i);
    }
    match e {
        Expr::Column(id) => {
            let target = slot(root, &Expr::Column(id));
            match layout.iter().position(|s| *s == target) {
                Some(i) => Expr::Column(i),
                None => match root.glob.node(id) {
                    VarNode::Var(var) if root.cur_outer_rels.is_member(var.varno) => {
                        super::paramassign::replace_nestloop_param(root, Expr::Column(id))
                    }
                    VarNode::PlaceHolderVar(phv)
                        if !root.cur_outer_rels.is_empty()
                            && root.placeholder_list[root.placeholder_array[&phv.phid]]
                                .ph_eval_at
                                .is_subset(&root.cur_outer_rels) =>
                    {
                        super::paramassign::replace_nestloop_param(root, Expr::Column(id))
                    }
                    VarNode::PlaceHolderVar(phv) => {
                        positional(root, root.glob.placeholder(phv.phid).phexpr.clone(), layout)
                    }
                    VarNode::Var(_) => unreachable!("every Var has a column"),
                },
            }
        }
        other => other.map_children(&mut |c| positional(root, c, layout)),
    }
}
