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
use super::nodes::{
    FromExpr, JoinExpr, JoinTreeNode, JoinType, Path, PlannerGlobal, Query, RangeTblEntry, RteKind, TargetEntry,
};
use crate::expr::{CmpOp, Expr, SubPlan};
use crate::plan::Plan;
use crate::query::Ctx;
use crate::types::Value;

/// sublink_convertible reports whether pull_up_sublinks turns an `EXISTS` or `IN` subquery of the WHERE clause into
/// a join.
pub(crate) fn sublink_convertible(e: &Expr) -> bool {
    match e {
        Expr::Exists(plan) => exists_parts(plan).is_some(),
        Expr::AnySubquery(test, plan, false) => any_convertible(test, plan),
        _ => false,
    }
}

/// values_convertible reports whether pull_up_sublinks turns an `IN` test of a VALUES list of the WHERE clause into a
/// comparison with an array, as convert_values_to_any does: the test is one comparison, and the list has two or more
/// rows of one constant value.
pub(crate) fn values_convertible(e: &Expr) -> bool {
    let Expr::AnySubquery(test, plan, false) = e else { return false };
    let Plan::Values(rows) = &**plan else { return false };
    matches!(**test, Expr::Compare(..))
        && rows.len() >= 2
        && rows.iter().all(|row| row.len() == 1 && crate::indexscan::is_constant(&row[0]) && !volatile(&row[0]))
}

/// convert_values_to_any returns the comparison with an array of the values of a VALUES list that an `IN` test of
/// the list is, as Postgres' convert_VALUES_to_ANY does, when values_convertible allows it and the values evaluate.
fn convert_values_to_any(ctx: &mut Ctx<'_>, e: &Expr) -> Option<Expr> {
    if !values_convertible(e) {
        return None;
    }
    let Expr::AnySubquery(test, plan, _) = e else { return None };
    let Plan::Values(rows) = &**plan else { return None };
    let values = rows.iter().map(|row| row[0].eval(ctx, &[]).ok()).collect::<Option<Vec<Value>>>()?;
    let element = values.iter().find_map(super::nodefuncs::value_type)?;
    let array = Value::Array(Box::new(crate::array::Array::one_dimensional(element, values)));
    Some(Expr::AnyArray(test.clone(), Box::new(Expr::Const(array)), false))
}

/// make_subplan returns the SubPlan of a subquery expression over a row of `width` columns, whose subquery is the
/// binder's plan, planned already or not, as Postgres' function of the same name does: the subquery's reads of the
/// enclosing row become reads of the SubPlan's arguments, and an unplanned subquery is planned for the share of its
/// rows that the expression reads, once the planner plans the SELECT around it when that SELECT is being bound.
pub(crate) fn make_subplan(ctx: &mut Ctx<'_>, mut link: Expr, planned: bool, width: usize) -> Expr {
    let plan = link.subquery_mut().expect("a subquery expression");
    let args = replace_correlation_vars(plan, width);
    if planned {
        let plan = std::mem::replace(plan, Plan::OneRow);
        return Expr::SubPlan(Box::new(build_subplan(ctx, link, plan, None, args)));
    }
    let subplan = SubPlan { link, args, init_plan: false, planned: false, startup_cost: 0.0, per_call_cost: 0.0 };
    match ctx.defer_subplans {
        true => Expr::SubPlan(Box::new(subplan)),
        false => plan_subplan(ctx, subplan),
    }
}

/// plan_subplan plans the subquery of a SubPlan that its binding left unplanned, as the rest of Postgres'
/// make_subplan does, after the SubPlan's test. A correlated EXISTS whose conditions convert_EXISTS_to_ANY turns into
/// an IN test becomes an AlternativeSubPlan of both, which the plan around it picks from.
fn plan_subplan(ctx: &mut Ctx<'_>, subplan: SubPlan) -> Expr {
    let SubPlan { mut link, args, .. } = subplan;
    link = link.map_children(&mut |c| preprocess_subplans(ctx, c));
    let plan = std::mem::replace(link.subquery_mut().expect("a subquery expression"), Plan::OneRow);
    let tuple_fraction = match &link {
        Expr::Exists(_) => 1.0,
        Expr::AnySubquery(..) => 0.5,
        _ => 0.0,
    };
    let orig_subquery = matches!(link, Expr::Exists(_)).then(|| plan.clone());
    let mut glob = PlannerGlobal::default();
    let mut subquery = super::query::unbind(&mut glob, ctx, plan);
    let simple_exists = orig_subquery.is_some() && simplify_exists_query(&mut subquery);
    let (plan, path) = super::plan_subselect(ctx, &mut glob, subquery, tuple_fraction);
    let result = build_subplan(ctx, link, plan, Some(path), args.clone());
    let Some(orig_subquery) = orig_subquery.filter(|_| simple_exists && !result.init_plan) else {
        return Expr::SubPlan(Box::new(result));
    };
    let mut glob = PlannerGlobal::default();
    let mut subquery = super::query::unbind(&mut glob, ctx, orig_subquery);
    simplify_exists_query(&mut subquery);
    let Some((subquery, testexpr)) = convert_exists_to_any(&glob, subquery, &args) else {
        return Expr::SubPlan(Box::new(result));
    };
    let (plan, path) = super::plan_subselect(ctx, &mut glob, subquery, 0.0);
    if !subpath_is_hashable(&path) {
        return Expr::SubPlan(Box::new(result));
    }
    let link = Expr::AnySubquery(Box::new(testexpr), Box::new(Plan::OneRow), false);
    let hashplan = build_subplan(ctx, link, plan, Some(path), Vec::new());
    Expr::AlternativeSubPlan(vec![result, hashplan])
}

/// preprocess_subplans plans each SubPlan of an expression that its binding left unplanned, as Postgres'
/// preprocess_expression turns a query's sublinks into SubPlans.
pub fn preprocess_subplans(ctx: &mut Ctx<'_>, e: Expr) -> Expr {
    match e {
        Expr::SubPlan(subplan) if !subplan.planned => plan_subplan(ctx, *subplan),
        other => other.map_children(&mut |c| preprocess_subplans(ctx, c)),
    }
}

/// build_subplan makes the SubPlan of a subquery expression from its subquery's plan, the path the planner chose for
/// it, and its arguments, as Postgres' function of the same name does: a subquery that reads no column of the
/// enclosing row is an initplan unless it is an ANY or ALL test, an uncorrelated IN test hashes the subquery's rows
/// when they fit in memory, and the rows of a subquery that reads no enclosing row are kept for every call.
fn build_subplan(ctx: &Ctx<'_>, mut link: Expr, plan: Plan, mut path: Option<Rc<Path>>, args: Vec<Expr>) -> SubPlan {
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
    let mut subplan = SubPlan { link, args, init_plan, planned: true, startup_cost: 0.0, per_call_cost: 0.0 };
    if let Some(path) = &path {
        cost_subplan(&mut subplan, path, use_hash_table, materializes);
    }
    subplan
}

/// convert_exists_to_any returns the subquery and test of the IN test that an EXISTS subquery is, when the conditions
/// of its WHERE clause that read the enclosing row are equalities of an expression of the enclosing row and one of
/// the subquery's own rows, as Postgres' convert_EXISTS_to_ANY finds them: the subquery returns the expressions of its
/// own rows, without those conditions, and the test compares them to the expressions of the enclosing row, which read
/// the SubPlan's arguments `args` as the enclosing row's columns.
fn convert_exists_to_any(glob: &PlannerGlobal, mut subselect: Query, args: &[Expr]) -> Option<(Query, Expr)> {
    let where_clause = std::mem::take(&mut subselect.jointree.quals);
    if query_reads_enclosing(&subselect)
        || where_clause.iter().any(|c| super::clauses::contain_volatile_functions(glob, c))
    {
        return None;
    }
    let (mut leftargs, mut rightargs, mut new_where) = (Vec::new(), Vec::new(), Vec::new());
    for clause in where_clause {
        match clause {
            Expr::Compare(op, left, right) if hash_ok_operator(op) && reads_enclosing(&left) => {
                leftargs.push(*left);
                rightargs.push(*right);
            }
            Expr::Compare(op, left, right) if hash_ok_operator(op) && reads_enclosing(&right) => {
                leftargs.push(*right);
                rightargs.push(*left);
            }
            other => new_where.push(other),
        }
    }
    if leftargs.is_empty() || new_where.iter().chain(&rightargs).any(reads_enclosing) {
        return None;
    }
    let mut reads_own = false;
    leftargs.iter().for_each(|e| e.visit(&mut |x| reads_own |= matches!(x, Expr::Column(_))));
    if reads_own || leftargs.iter().any(super::clauses::contain_subplans) {
        return None;
    }
    subselect.jointree.quals = new_where;
    let n = rightargs.len();
    subselect.target_list =
        rightargs.into_iter().map(|expr| TargetEntry { expr, resjunk: false, ressortgroupref: 0 }).collect();
    let testlist = leftargs.into_iter().enumerate().map(|(i, left)| {
        let value = match n {
            1 => Expr::SubqueryValue,
            _ => Expr::Field(Box::new(Expr::SubqueryValue), i),
        };
        Expr::Compare(CmpOp::Eq, Box::new(increment_var_sublevels_up(left, args)), Box::new(value))
    });
    let testexpr = testlist.reduce(|a, b| Expr::And(Box::new(a), Box::new(b))).expect("a hash clause");
    Some((subselect, testexpr))
}

/// increment_var_sublevels_up rewrites an expression of a subquery that reads the enclosing row through the
/// SubPlan's arguments into one of the enclosing query, as Postgres' IncrementVarSublevelsUp moves it up a level.
fn increment_var_sublevels_up(e: Expr, args: &[Expr]) -> Expr {
    match e {
        Expr::Outer(1, n) => args[n].clone(),
        Expr::Outer(d, i) => Expr::Outer(d - 1, i),
        other => other.map_children(&mut |c| increment_var_sublevels_up(c, args)),
    }
}

/// reads_enclosing reports whether an expression of a subquery reads its enclosing row, as Postgres'
/// contain_vars_of_level does for level 1.
fn reads_enclosing(e: &Expr) -> bool {
    let mut reads = false;
    e.visit(&mut |x| match x {
        Expr::Outer(1, _) => reads = true,
        x => {
            for plan in x.subqueries() {
                let mut read = BTreeSet::new();
                reads |= !crate::indexscan::outer_reads(plan, 2, &mut read) || !read.is_empty();
            }
        }
    });
    reads
}

/// query_reads_enclosing reports whether anything of a query reads its enclosing row: its range table's inputs,
/// join conditions, and upper clauses.
fn query_reads_enclosing(query: &Query) -> bool {
    let mut reads = query.upper_exprs().iter().any(reads_enclosing);
    query.jointree.fromlist.iter().for_each(|node| jointree_quals(node, &mut |q| reads |= reads_enclosing(q)));
    reads
        || query.rtable.iter().any(|rte| match &rte.kind {
            RteKind::Subquery(_, plan) | RteKind::Plan(plan) => {
                let mut read = BTreeSet::new();
                !crate::indexscan::outer_reads(plan, 1, &mut read) || !read.is_empty()
            }
            _ => false,
        })
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
    for plan in e.subqueries_mut() {
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
    if query.has_aggs
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
    query_exprs(&mut root.parse, &mut |x| {
        if let Expr::SubPlan(subplan) = x
            && subplan.init_plan
        {
            initplan_cost += subplan.startup_cost + subplan.per_call_cost;
        }
    });
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

/// ss_process_cte returns the join tree entry of a reference to a WITH query, as Postgres' SS_process_ctes decides
/// for the query: the query itself in place of the reference, as inline_cte puts it, when it is NOT MATERIALIZED, or
/// is referenced once and not MATERIALIZED, and is not recursive and runs no volatile function; and otherwise a scan
/// of its rows, which the planner plans once for every reference to read.
pub fn ss_process_cte(
    glob: &mut PlannerGlobal,
    ctx: &mut Ctx<'_>,
    def: std::sync::Arc<crate::plan::CteDef>,
    rtable: &mut Vec<RangeTblEntry>,
    output: &mut Vec<Expr>,
) -> JoinTreeNode {
    let refcount = def.refcount.load(std::sync::atomic::Ordering::Relaxed);
    let mut volatile_plan = false;
    def.plan.clone().map_exprs(0, &mut |e, _| {
        volatile_plan |= volatile(&e);
        e
    });
    if def.materialized.map_or(refcount == 1, |always| !always) && !def.recursive && !volatile_plan {
        return match def.planned {
            true => super::push_relation(glob, rtable, output, RteKind::Plan(def.plan.clone()), def.coltypes.clone()),
            false => super::build_jointree(glob, ctx, def.plan.clone(), rtable, output),
        };
    }
    def.shared.get_or_init(|| match def.planned {
        true => (def.plan.clone(), crate::joins::estimate(ctx, &def.plan)),
        false => {
            let mut glob = PlannerGlobal::default();
            let subquery = super::query::unbind(&mut glob, ctx, def.plan.clone());
            let (plan, path) = super::plan_subselect(ctx, &mut glob, subquery, 0.0);
            (plan, path.rows)
        }
    });
    let coltypes = def.coltypes.clone();
    super::push_relation(glob, rtable, output, RteKind::Plan(Plan::CteScan(def)), coltypes)
}

/// query_exprs calls a function with every expression of a query's level and each of their descendants, outside
/// subquery plans.
pub fn query_exprs(query: &mut Query, f: &mut dyn FnMut(&Expr)) {
    query.upper_exprs_mut().into_iter().for_each(|e| e.visit(f));
    query.jointree.fromlist.iter().for_each(|node| jointree_quals(node, &mut |q| q.visit(f)));
    query.jointree.quals.iter().for_each(|q| q.visit(f));
}

/// preprocess_query_subplans plans the SubPlans of a query's expressions and of its range table's inputs that their
/// binding left unplanned, as Postgres' subquery_planner preprocesses each expression of a query.
pub fn preprocess_query_subplans(ctx: &mut Ctx<'_>, parse: &mut Query) {
    let mut process = |e: &mut Expr| {
        let old = std::mem::replace(e, Expr::SubqueryValue);
        *e = preprocess_subplans(ctx, old);
    };
    parse.upper_exprs_mut().into_iter().for_each(&mut process);
    parse.jointree.fromlist.iter_mut().for_each(|node| jointree_quals_mut(node, &mut process));
    parse.jointree.quals.iter_mut().for_each(&mut process);
    for rte in &mut parse.rtable {
        if let RteKind::Plan(plan) = &mut rte.kind {
            plan.map_exprs(0, &mut |e, _| preprocess_subplans(ctx, e));
        }
    }
}

/// jointree_quals_mut is jointree_quals for changing the quals.
fn jointree_quals_mut(node: &mut JoinTreeNode, f: &mut dyn FnMut(&mut Expr)) {
    match node {
        JoinTreeNode::Rel(_) => {}
        JoinTreeNode::From(from) => {
            from.fromlist.iter_mut().for_each(|n| jointree_quals_mut(n, f));
            from.quals.iter_mut().for_each(&mut *f);
        }
        JoinTreeNode::Join(join) => {
            jointree_quals_mut(&mut join.larg, f);
            jointree_quals_mut(&mut join.rarg, f);
            join.quals.iter_mut().for_each(&mut *f);
        }
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
/// the join tree with the subquery's rows, or into a comparison with an array of a VALUES list's values, returning
/// the new join tree and the other conditions, as Postgres' function of the same name does for the WHERE clause's
/// top-level conditions.
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
        if let Some(saop) = convert_values_to_any(ctx, &qual) {
            remaining.push(saop);
            continue;
        }
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

/// ss_make_initplan_from_plan returns the expression of an initplan that reads the value of the first column of the
/// first row of a plan that the planner made, at a cost, as Postgres' function of the same name makes it.
pub fn ss_make_initplan_from_plan(plan: Plan, startup_cost: f64) -> Expr {
    let uncorrelated = crate::joins::plan_lowest_level(&plan).is_some_and(|level| level >= 0);
    let link = Expr::Scalar(Box::new(crate::plan::share_subquery(plan, uncorrelated)));
    Expr::SubPlan(Box::new(SubPlan {
        link,
        args: Vec::new(),
        init_plan: true,
        planned: true,
        startup_cost,
        per_call_cost: 0.0,
    }))
}

/// increment_query_sublevels_up rewrites each read of an enclosing row in a query's expressions to read the row one
/// further out, for planning the query as a subquery of itself, as Postgres' IncrementVarSublevelsUp does.
pub fn increment_query_sublevels_up(query: &mut Query) {
    let mut process = |e: &mut Expr| {
        let old = std::mem::replace(e, Expr::SubqueryValue);
        *e = increment_sublevels_up(old, 0);
    };
    query.upper_exprs_mut().into_iter().for_each(&mut process);
    query.jointree.fromlist.iter_mut().for_each(|node| jointree_quals_mut(node, &mut process));
    query.jointree.quals.iter_mut().for_each(&mut process);
    for rte in &mut query.rtable {
        if let RteKind::Plan(plan) = &mut rte.kind {
            plan.map_exprs(0, &mut |e, depth| increment_sublevels_up(e, depth));
        }
    }
}

/// increment_sublevels_up rewrites each read of an enclosing row in an expression, `nesting` subqueries deep within
/// its query, to read the row one further out.
pub fn increment_sublevels_up(e: Expr, nesting: usize) -> Expr {
    let mut e = match e {
        Expr::Outer(d, i) if d > nesting => return Expr::Outer(d + 1, i),
        other => other.map_children(&mut |c| increment_sublevels_up(c, nesting)),
    };
    for plan in e.subqueries_mut() {
        plan.map_exprs(0, &mut |x, depth| increment_sublevels_up(x, nesting + 1 + depth));
    }
    e
}
