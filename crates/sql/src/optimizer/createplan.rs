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

use super::PlannerInfo;
use super::costsize::cost_qual_eval_node;
use super::indxpath::{lookup_keys, to_attnos};
use super::joinpath::clause_sides_match_join;
use super::nodes::{IndexPath, JoinType, Path, PathKind, RinfoId, RteKind, VarNode};
use super::restrictinfo::rinfo_is_pushed_down;
use crate::expr::Expr;
use crate::plan::{JoinKind, JoinMethod, Plan};
use crate::types::Value;

/// Slot is what a column of a plan's rows holds: a relation's attribute, or a PlaceHolderVar by its ID.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Slot {
    Var(usize, usize),
    PlaceHolder(usize),
}

/// create_plan makes the plan of a path of the relation that joins every base relation, whose columns are the
/// query's target list, where those that nothing reads are NULL or whatever the path's rows hold there.
pub fn create_plan(root: &mut PlannerInfo<'_, '_>, path: &Path) -> Plan {
    let (plan, layout) = create_plan_recurse(root, path);
    let exprs: Vec<Option<Expr>> =
        root.parse.target_list.iter().map(|e| e.as_ref().map(|e| positional(root, e.clone(), &layout))).collect();
    if exprs.len() == layout.len()
        && exprs.iter().enumerate().all(|(i, e)| e.as_ref().is_none_or(|e| *e == Expr::Column(i)))
    {
        return plan;
    }
    let exprs = exprs.into_iter().map(|e| e.unwrap_or(Expr::Const(Value::Null))).collect();
    Plan::Project { input: Box::new(plan), exprs }
}

/// create_plan_recurse makes the plan of a path, returning it with what each column of its rows holds.
fn create_plan_recurse(root: &mut PlannerInfo<'_, '_>, path: &Path) -> (Plan, Vec<Slot>) {
    let (plan, layout) = match &path.kind {
        PathKind::SeqScan | PathKind::Lookup(_) => create_scan_plan(root, path.parent),
        PathKind::Result(quals) => {
            let quals = quals.iter().map(|q| positional(root, q.clone(), &[]));
            let predicate = quals.reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
            match predicate {
                Some(predicate) => (Plan::Filter { input: Box::new(Plan::OneRow), predicate }, Vec::new()),
                None => (Plan::OneRow, Vec::new()),
            }
        }
        PathKind::IndexScan(_) if !path.param.is_empty() => create_scan_plan(root, path.parent),
        PathKind::IndexScan(best_path) => create_indexscan_plan(root, path.parent, best_path),
        PathKind::BitmapHeapScan(_) => create_bitmap_scan_plan(root, path, &[]),
        PathKind::BitmapAnd(_) | PathKind::BitmapOr(_) => unreachable!("a bitmap tree is planned by its heap scan"),
        PathKind::Append(_) => {
            let layout: Vec<Slot> = root.rels[path.parent].reltarget.exprs.iter().map(|e| slot(root, e)).collect();
            let nulls =
                Plan::Project { input: Box::new(Plan::OneRow), exprs: vec![Expr::Const(Value::Null); layout.len()] };
            (Plan::Filter { input: Box::new(nulls), predicate: Expr::Const(Value::Bool(false)) }, layout)
        }
        PathKind::Material(subpath) | PathKind::Sort(subpath) => return create_plan_recurse(root, subpath),
        PathKind::NestLoop(join) | PathKind::HashJoin(join) => {
            let (outer_plan, outer_layout) = create_plan_recurse(root, &join.outer);
            let lateral = matches!(join.inner.kind, PathKind::BitmapHeapScan(_)) && !join.inner.param.is_empty();
            let (inner_plan, inner_layout) = match lateral {
                true => {
                    let (plan, layout) = create_bitmap_scan_plan(root, &join.inner, &outer_layout);
                    let (mut plan, layout) = add_placeholders(root, join.inner.parent, plan, layout);
                    plan.map_exprs(0, &mut |e, depth| read_lateral_row(e, depth));
                    (plan, layout)
                }
                false => create_plan_recurse(root, &join.inner),
            };
            let layout = [outer_layout.as_slice(), inner_layout.as_slice()].concat();
            let (joinquals, otherquals): (Vec<RinfoId>, Vec<RinfoId>) = match join.jointype.is_outer() {
                true => join
                    .joinrestrictinfo
                    .iter()
                    .copied()
                    .partition(|&r| !rinfo_is_pushed_down(&root.rinfos[r], &path.relids)),
                false => (join.joinrestrictinfo.clone(), Vec::new()),
            };
            let joinquals: Vec<Expr> = match &path.kind {
                PathKind::HashJoin(_) => {
                    let (hashclauses, rest): (Vec<RinfoId>, Vec<RinfoId>) = joinquals.into_iter().partition(|&r| {
                        let r = &root.rinfos[r];
                        r.hashjoinable && clause_sides_match_join(r, &join.outer.relids, &join.inner.relids)
                    });
                    let mut quals = get_switched_clauses(root, &hashclauses, &join.outer.relids);
                    quals.extend(order_qual_clauses(root, rest).into_iter().map(|r| root.rinfos[r].clause.clone()));
                    quals
                }
                _ => order_qual_clauses(root, joinquals).into_iter().map(|r| root.rinfos[r].clause.clone()).collect(),
            };
            let otherquals = order_qual_clauses(root, otherquals);
            let method = match (&path.kind, &join.inner.kind) {
                (PathKind::HashJoin(_), _) => JoinMethod::Hash,
                (_, PathKind::IndexScan(best_path)) if !join.inner.param.is_empty() => {
                    let rel = join.inner.parent;
                    let keys = lookup_keys(root, rel, best_path.index, &best_path.indexclauses)
                        .expect("a parameterized index path has lookup keys");
                    let table = root.parse.rte(rel).table().expect("an index path scans a table").clone();
                    let scan = crate::indexscan::IndexScan {
                        table: Box::new(table),
                        index: root.rels[rel].indexlist[best_path.index].index,
                        ranges: Vec::new(),
                        reverse: false,
                        nearest: None,
                        needed: None,
                        lookup_heavy: None,
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
            (filtered(root, plan, &otherquals, &layout), layout)
        }
    };
    add_placeholders(root, path.parent, plan, layout)
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
/// under a filter, the one row of a RESULT relation, or its own plan with the restrictions pushed into it, where an
/// index of a system catalog may answer them.
fn create_scan_plan(root: &mut PlannerInfo<'_, '_>, rel: usize) -> (Plan, Vec<Slot>) {
    let restrictinfo = order_qual_clauses(root, root.rels[rel].baserestrictinfo.clone());
    match root.parse.rte(rel).kind.clone() {
        RteKind::Relation(plan, _) => {
            let layout = base_slots(rel, plan.width());
            (filtered(root, plan, &restrictinfo, &layout), layout)
        }
        RteKind::Result => (filtered(root, Plan::OneRow, &restrictinfo, &[]), Vec::new()),
        RteKind::Plan(plan) => {
            let layout = base_slots(rel, plan.width());
            let predicate = restrictinfo
                .iter()
                .map(|&r| to_attnos(root, root.rinfos[r].clause.clone(), rel))
                .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
            let plan = match predicate {
                Some(predicate) => crate::plan::Planner { ctx: root.ctx, outer: Vec::new() }
                    .use_indexes(crate::plan::push_down(plan, predicate)),
                None => plan,
            };
            (plan, layout)
        }
        RteKind::Subquery(..) | RteKind::Join(_) => unreachable!("only base relations are scanned"),
    }
}

/// create_indexscan_plan makes the plan of a scan of an index, reading the ranges that its index clauses give and
/// testing the restrictions that those ranges do not answer exactly, as Postgres' function of the same name does. A
/// scan of every entry of the primary index in its order is the table's sequential scan.
fn create_indexscan_plan(root: &mut PlannerInfo<'_, '_>, rel: usize, best_path: &IndexPath) -> (Plan, Vec<Slot>) {
    let table = root.parse.rte(rel).table().expect("an index path scans a table").clone();
    let info = &root.rels[rel].indexlist[best_path.index];
    let (index, scan_clauses) = (info.index, info.indrestrictinfo.clone());
    if index.is_none() && best_path.indexclauses.is_empty() && !best_path.backward {
        return create_scan_plan(root, rel);
    }
    let indexquals = best_path.indexclauses.iter().flat_map(|iclause| &iclause.indexquals);
    let predicate = indexquals
        .map(|&r| to_attnos(root, root.rinfos[r].clause.clone(), rel))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let scan_of = |root: &mut PlannerInfo<'_, '_>, predicate: Option<&Expr>| {
        crate::indexscan::scan_of_index(root.ctx, &table, index, predicate, best_path.backward)
    };
    let (scan, exact) = match scan_of(root, predicate.as_ref()) {
        Some(found) => found,
        None => (scan_of(root, None).expect("a scan of every entry is always possible").0, false),
    };
    let qpqual: Vec<RinfoId> = scan_clauses
        .into_iter()
        .filter(|&r| !exact || !best_path.indexclauses.iter().any(|iclause| iclause.rinfo == r && !iclause.lossy))
        .collect();
    let qpqual = order_qual_clauses(root, qpqual);
    let layout = base_slots(rel, table.columns.len());
    (filtered(root, Plan::IndexScan(Box::new(scan)), &qpqual, &layout), layout)
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
    let table = root.parse.rte(rel).table().expect("a bitmap scan reads a table").clone();
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
                && !indexquals.iter().any(|q| same_clause(q, &rinfo.clause))
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
    let scan =
        crate::indexscan::BitmapHeapScan { table: Box::new(table), bitmap, recheck, lossy: !exact, needed: None };
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
    table: &crate::catalog::table::TableDef,
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
            let indexquals: Vec<Expr> = ipath
                .indexclauses
                .iter()
                .flat_map(|iclause| &iclause.indexquals)
                .map(|&r| root.rinfos[r].clause.clone())
                .collect();
            let index_ecs: Vec<super::nodes::EcId> =
                ipath.indexclauses.iter().filter_map(|iclause| root.rinfos[iclause.rinfo].parent_ec).collect();
            let every = |root: &mut PlannerInfo<'_, '_>| {
                crate::indexscan::scan_of_index(root.ctx, table, index, None, false).expect("a scan of every entry").0
            };
            if !bitmapqual.param.is_empty() {
                let layout = base_slots(rel, table.columns.len());
                let cond = and(indexquals.iter().map(|e| param_positional(root, e.clone(), &layout, outer)).collect());
                return (Bitmap::Index(Box::new(every(root)), cond), quals, indexquals, index_ecs);
            }
            let predicate = and(indexquals.iter().map(|e| to_attnos(root, e.clone(), rel)).collect());
            let scan = match crate::indexscan::scan_of_index(root.ctx, table, index, predicate.as_ref(), false) {
                Some((scan, covered)) => {
                    *exact &= covered;
                    scan
                }
                None => {
                    *exact = false;
                    every(root)
                }
            };
            (Bitmap::Index(Box::new(scan), None), quals, indexquals, index_ecs)
        }
        _ => unreachable!("a bitmap tree holds index scans, BitmapAnds, and BitmapOrs"),
    }
}

/// same_clause reports whether two clauses are the same, as ANDs and ORs of the same arguments in any order are.
fn same_clause(a: &Expr, b: &Expr) -> bool {
    use super::restrictinfo::{and_args, or_args};
    let same_args =
        |x: Vec<&Expr>, y: Vec<&Expr>| x.len() == y.len() && x.iter().all(|e| y.iter().any(|f| same_clause(e, f)));
    match (a, b) {
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
    if let Expr::Exists(p) | Expr::Scalar(p) | Expr::ArraySubquery(p, _) | Expr::AnySubquery(_, p, _) = &mut e {
        p.map_exprs(0, &mut |x, d| read_lateral_row(x, depth + 1 + d));
    }
    e
}

/// order_qual_clauses sorts clauses by the cost of evaluating them, cheapest first and otherwise in their order, as
/// Postgres' function of the same name does.
fn order_qual_clauses(root: &PlannerInfo<'_, '_>, mut clauses: Vec<RinfoId>) -> Vec<RinfoId> {
    clauses.sort_by(|&a, &b| {
        let cost = |r: RinfoId| cost_qual_eval_node(&root.rinfos[r].clause).per_tuple;
        cost(a).total_cmp(&cost(b))
    });
    clauses
}

/// get_switched_clauses returns the clauses of hash clauses with each one's outer side first, as Postgres' function
/// of the same name does.
fn get_switched_clauses(
    root: &PlannerInfo<'_, '_>,
    clauses: &[RinfoId],
    outer_relids: &super::nodes::Relids,
) -> Vec<Expr> {
    clauses
        .iter()
        .map(|&r| {
            let r = &root.rinfos[r];
            match &r.clause {
                Expr::Compare(op, left, right) if r.can_join && r.right_relids.is_subset(outer_relids) => {
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
/// Postgres' setrefs.c rewrites a plan's Vars to refer to its inputs' columns. A PlaceHolderVar that the rows do not
/// hold is computed from its expression.
fn positional(root: &PlannerInfo<'_, '_>, e: Expr, layout: &[Slot]) -> Expr {
    match e {
        Expr::Column(id) => {
            let target = slot(root, &Expr::Column(id));
            match layout.iter().position(|s| *s == target) {
                Some(i) => Expr::Column(i),
                None => match root.glob.node(id) {
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
