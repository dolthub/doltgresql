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
//! Doltgres' plan, whose expressions read columns of their input rows by position rather than Vars.

use std::rc::Rc;

use super::PlannerInfo;
use super::costsize::cost_qual_eval_node;
use super::indxpath::{base_vars, to_attnos};
use super::joinpath::clause_sides_match_join;
use super::nodes::{JoinType, Path, PathKind, Relids, RestrictInfo, is_subset};
use super::restrictinfo::is_pushed_down;
use crate::expr::Expr;
use crate::plan::{JoinKind, JoinMethod, Plan};

/// create_plan makes the plan of a path of the relation that joins every base relation, whose columns are in the
/// order of the query's output Vars.
pub fn create_plan(root: &mut PlannerInfo<'_, '_>, path: &Path) -> Plan {
    let (plan, layout) = create_plan_recurse(root, path);
    if layout == root.parse.output {
        return plan;
    }
    let exprs = root.parse.output.iter().map(|v| Expr::Column(position(&layout, *v))).collect();
    Plan::Project { input: Box::new(plan), exprs }
}

/// create_plan_recurse makes the plan of a path, returning it with the Var of each column of its rows.
fn create_plan_recurse(root: &mut PlannerInfo<'_, '_>, path: &Path) -> (Plan, Vec<usize>) {
    match &path.kind {
        PathKind::SeqScan | PathKind::Lookup(_) => create_scan_plan(root, path.relids.trailing_zeros() as usize),
        PathKind::IndexScan(scan, exact) => {
            let rel = path.relids.trailing_zeros() as usize;
            let layout = base_vars(rel, scan.table.columns.len());
            let plan = Plan::IndexScan(scan.clone());
            match *exact {
                true => (plan, layout),
                false => {
                    (filtered(plan, &order_qual_clauses(root.rels[rel].baserestrictinfo.clone()), &layout), layout)
                }
            }
        }
        PathKind::Material(subpath) => create_plan_recurse(root, subpath),
        PathKind::NestLoop(join) | PathKind::HashJoin(join) => {
            let (outer_plan, outer_layout) = create_plan_recurse(root, &join.outer);
            let (inner_plan, inner_layout) = create_plan_recurse(root, &join.inner);
            let layout = [outer_layout.as_slice(), inner_layout.as_slice()].concat();
            let (joinquals, otherquals): (Vec<Rc<RestrictInfo>>, Vec<Rc<RestrictInfo>>) = match join.jointype.is_outer()
            {
                true => join.joinrestrictinfo.iter().cloned().partition(|r| !is_pushed_down(r, path.relids)),
                false => (join.joinrestrictinfo.clone(), Vec::new()),
            };
            let joinquals = match &path.kind {
                PathKind::HashJoin(_) => {
                    let (hashclauses, rest): (Vec<Rc<RestrictInfo>>, Vec<Rc<RestrictInfo>>) =
                        joinquals.into_iter().partition(|r| {
                            r.hashjoinable && clause_sides_match_join(r, join.outer.relids, join.inner.relids)
                        });
                    [get_switched_clauses(&hashclauses, join.outer.relids), order_qual_clauses(rest)].concat()
                }
                _ => order_qual_clauses(joinquals),
            };
            let otherquals = order_qual_clauses(otherquals);
            let method = match (&path.kind, &join.inner.kind) {
                (PathKind::HashJoin(_), _) => JoinMethod::Hash,
                (_, PathKind::Lookup(JoinMethod::Lookup { scan, keys })) => JoinMethod::Lookup {
                    scan: scan.clone(),
                    keys: keys.iter().map(|k| positional(k, &outer_layout)).collect(),
                },
                (_, PathKind::Lookup(JoinMethod::CatalogLookup { index, keys })) => JoinMethod::CatalogLookup {
                    index,
                    keys: keys.iter().map(|k| positional(k, &outer_layout)).collect(),
                },
                _ => JoinMethod::NestedLoop,
            };
            let kind = match join.jointype {
                JoinType::Inner => JoinKind::Inner,
                JoinType::Left => JoinKind::Left,
                JoinType::Right => JoinKind::Right,
                JoinType::Full => JoinKind::Full,
                JoinType::Semi => JoinKind::Semi,
                JoinType::Anti => JoinKind::Anti,
            };
            let plan = Plan::Join {
                left: Box::new(outer_plan),
                right: Box::new(inner_plan),
                kind,
                condition: and(&joinquals, &layout),
                lateral: false,
                method,
            };
            (filtered(plan, &otherquals, &layout), layout)
        }
    }
}

/// create_scan_plan makes the plan that reads a base relation's rows and tests its restrictions: a scan of its table
/// under a filter, or its own plan with the restrictions pushed into it, where an index of a system catalog may
/// answer them.
fn create_scan_plan(root: &mut PlannerInfo<'_, '_>, rel: usize) -> (Plan, Vec<usize>) {
    let rte = root.parse.rte(rel);
    let plan = rte.plan.clone();
    let layout = base_vars(rel, plan.width());
    let restrictinfo = order_qual_clauses(root.rels[rel].baserestrictinfo.clone());
    if rte.table().is_some() {
        return (filtered(plan, &restrictinfo, &layout), layout);
    }
    let plan =
        match restrictinfo.iter().map(|r| to_attnos(&r.clause, rel)).reduce(|a, b| Expr::And(Box::new(a), Box::new(b)))
        {
            Some(predicate) => crate::plan::Planner { ctx: root.ctx, outer: Vec::new() }
                .use_indexes(crate::plan::push_down(plan, predicate)),
            None => plan,
        };
    (plan, layout)
}

/// order_qual_clauses sorts clauses by the cost of evaluating them, cheapest first and otherwise in their order, as
/// Postgres' function of the same name does.
fn order_qual_clauses(mut clauses: Vec<Rc<RestrictInfo>>) -> Vec<Rc<RestrictInfo>> {
    clauses
        .sort_by(|a, b| cost_qual_eval_node(&a.clause).per_tuple.total_cmp(&cost_qual_eval_node(&b.clause).per_tuple));
    clauses
}

/// get_switched_clauses returns hash clauses with each one's outer side first, as Postgres' function of the same name
/// does.
fn get_switched_clauses(clauses: &[Rc<RestrictInfo>], outer_relids: Relids) -> Vec<Rc<RestrictInfo>> {
    clauses
        .iter()
        .map(|r| match &r.clause {
            Expr::Compare(op, left, right) if r.can_join && is_subset(r.right_relids, outer_relids) => {
                let clause = Expr::Compare(crate::indexscan::swap(*op), right.clone(), left.clone());
                Rc::new(RestrictInfo {
                    clause,
                    left_relids: r.right_relids,
                    right_relids: r.left_relids,
                    ..(**r).clone()
                })
            }
            _ => r.clone(),
        })
        .collect()
}

/// filtered returns a plan under a filter of clauses over its rows, or the plan itself without clauses.
fn filtered(plan: Plan, clauses: &[Rc<RestrictInfo>], layout: &[usize]) -> Plan {
    match and(clauses, layout) {
        Some(predicate) => Plan::Filter { input: Box::new(plan), predicate },
        None => plan,
    }
}

/// and returns the conjunction of clauses over rows of the Vars of a layout, or None without clauses.
fn and(clauses: &[Rc<RestrictInfo>], layout: &[usize]) -> Option<Expr> {
    clauses.iter().map(|r| positional(&r.clause, layout)).reduce(|a, b| Expr::And(Box::new(a), Box::new(b)))
}

/// positional rewrites an expression over Vars into one over rows of the Vars of a layout, as Postgres' setrefs.c
/// rewrites a plan's Vars to refer to its inputs' columns.
fn positional(e: &Expr, layout: &[usize]) -> Expr {
    match e {
        Expr::Column(v) => Expr::Column(position(layout, *v)),
        other => other.clone().map_children(&mut |c| positional(&c, layout)),
    }
}

/// position returns the column of a layout's rows that holds a Var.
fn position(layout: &[usize], var: usize) -> usize {
    layout.iter().position(|v| *v == var).expect("every Var has a column")
}
