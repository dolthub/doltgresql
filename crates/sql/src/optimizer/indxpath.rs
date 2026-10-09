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

//! Postgres' optimizer/path/indxpath.c: the index paths of a base relation. Doltgres' index scans choose an index
//! and its ranges for a relation's restrictions, and Doltgres' lookup joins find the index that a join's equalities
//! look rows up in, so each kind of path comes from them and takes Postgres' costs.

use std::rc::Rc;

use prolly::NodeStore;

use super::PlannerInfo;
use super::clauses::pull_varnos;
use super::costsize::{
    IndexOptInfo, QualCost, clamp_row_est, cost_index, cost_qual_eval, estimate_rel_pages, get_typavgwidth,
};
use super::nodes::{Path, PathKind, Relids, RestrictInfo, is_subset, members, singleton, var, var_parts};
use super::pathnode::add_path;
use crate::catalog::table::TableDef;
use crate::expr::{CmpOp, Expr};
use crate::plan::JoinMethod;

/// create_index_paths adds the paths of a base relation's index scans: the one that Doltgres chooses for its
/// restrictions, and a lookup for each set of other relations whose join equalities find its rows through an index,
/// as Postgres' function of the same name adds plain and parameterized index paths.
pub fn create_index_paths(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    if let Some(table) = root.parse.rte(rel).table().cloned() {
        create_restriction_index_path(root, rel, &table);
    }
    let mut outer_sets: Vec<Relids> = Vec::new();
    for rinfo in &root.rels[rel].joininfo {
        let Some((_, outer)) = join_equality(rinfo, rel) else { continue };
        let relids = pull_varnos(outer);
        if !outer_sets.contains(&relids) {
            outer_sets.push(relids);
        }
    }
    let all = outer_sets.iter().fold(0, |relids, r| relids | r);
    if outer_sets.len() > 1 {
        outer_sets.push(all);
    }
    for outer_relids in outer_sets {
        create_lookup_path(root, rel, outer_relids);
    }
}

/// join_equality returns the sides of a join clause that equates an expression of a relation with one of others,
/// as an index lookup of that relation's rows may search by.
fn join_equality(rinfo: &RestrictInfo, rel: usize) -> Option<(&Expr, &Expr)> {
    let Expr::Compare(CmpOp::Eq, l, r) = &rinfo.clause else { return None };
    match (rinfo.can_join, rinfo.left_relids == singleton(rel), rinfo.right_relids == singleton(rel)) {
        (true, true, false) => Some((l, r)),
        (true, false, true) => Some((r, l)),
        _ => None,
    }
}

/// create_restriction_index_path adds the path of the index scan that Doltgres chooses for a table's restrictions.
fn create_restriction_index_path(root: &mut PlannerInfo<'_, '_>, rel: usize, table: &TableDef) {
    let restrictinfo = root.rels[rel].baserestrictinfo.clone();
    let predicate = restrictinfo
        .iter()
        .filter(|r| !r.pseudoconstant)
        .map(|r| to_attnos(&r.clause, rel))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let Some(predicate) = predicate else { return };
    let Some((scan, exact)) = crate::indexscan::choose_with_cover(root.ctx, table, &predicate) else { return };
    let index_tuples = scan.estimate(root.ctx).unwrap_or(root.rels[rel].rows);
    let columns = scan.index_columns();
    let on_index = |r: &&Rc<RestrictInfo>| {
        super::nodes::members(pull_varnos(&r.clause)).count() == 1
            && attnos(&r.clause).iter().all(|a| columns.contains(a))
    };
    let index_quals = restrictinfo.iter().filter(on_index).count();
    let qpquals: Vec<Rc<RestrictInfo>> = match exact {
        true => Vec::new(),
        false => restrictinfo.iter().filter(|r| !on_index(r)).cloned().collect(),
    };
    let index = index_info(root, rel, table, scan.index, scan.covering(), index_quals);
    let (startup_cost, total_cost) =
        cost_index(root, &root.rels[rel], &index, index_tuples, cost_qual_eval(&qpquals), 1.0);
    let parent = &mut root.rels[rel];
    let path = Path {
        kind: PathKind::IndexScan(Box::new(scan), exact),
        relids: parent.relids,
        param: 0,
        rows: parent.rows,
        width: parent.width,
        startup_cost,
        total_cost,
    };
    add_path(parent, Rc::new(path));
}

/// create_lookup_path adds the path of a lookup of a relation's rows by its join equalities with a set of other
/// relations, when an index of its table or catalog lets it look them up, as Postgres adds the parameterized index
/// path of the clauses that those relations' rows supply.
fn create_lookup_path(root: &mut PlannerInfo<'_, '_>, rel: usize, outer_relids: Relids) {
    let mut outer_vars: Vec<usize> = Vec::new();
    let mut equalities = Vec::new();
    for rinfo in &root.rels[rel].joininfo {
        let Some((inner, outer)) = join_equality(rinfo, rel) else { continue };
        if !is_subset(pull_varnos(outer), outer_relids) {
            continue;
        }
        outer.visit(&mut |e| {
            if let Expr::Column(v) = e
                && !outer_vars.contains(v)
            {
                outer_vars.push(*v);
            }
        });
        equalities.push((inner.clone(), outer.clone()));
    }
    let left_width = outer_vars.len();
    let position = |e: &Expr| -> Expr { positional(e, &outer_vars, rel, left_width) };
    let condition = equalities
        .iter()
        .map(|(inner, outer)| Expr::Compare(CmpOp::Eq, Box::new(position(outer)), Box::new(position(inner))))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let Some(condition) = condition else { return };
    let plan = root.parse.rte(rel).plan.clone();
    let Some(found) = crate::joins::lookup(root.ctx, &plan, &condition, left_width) else { return };
    let method = match found.method {
        JoinMethod::Lookup { scan, keys } => {
            JoinMethod::Lookup { scan, keys: keys.iter().map(|k| from_positional(k, &outer_vars)).collect() }
        }
        JoinMethod::CatalogLookup { index, keys } => {
            JoinMethod::CatalogLookup { index, keys: keys.iter().map(|k| from_positional(k, &outer_vars)).collect() }
        }
        other => other,
    };
    let loop_count = members(outer_relids).map(|r| root.rels[r].rows).fold(f64::INFINITY, f64::min);
    let loop_count = if loop_count.is_finite() { loop_count } else { 1.0 };
    let index = match (&method, root.parse.rte(rel).table().cloned()) {
        (JoinMethod::Lookup { scan, .. }, Some(table)) => index_info(root, rel, &table, scan.index, false, 1),
        _ => IndexOptInfo {
            pages: 1.0,
            tuples: root.rels[rel].tuples,
            tree_height: 0.0,
            indexonly: true,
            correlation: 1.0,
            nquals: 1,
        },
    };
    let parent = &root.rels[rel];
    let selectivity = if parent.tuples > 0.0 { parent.rows / parent.tuples } else { 1.0 };
    let rows = clamp_row_est(found.matches * selectivity);
    let qpqual_cost: QualCost = cost_qual_eval(&parent.baserestrictinfo);
    let (startup_cost, total_cost) = cost_index(root, parent, &index, found.matches, qpqual_cost, loop_count);
    let path = Path {
        kind: PathKind::Lookup(method),
        relids: parent.relids,
        param: outer_relids,
        rows,
        width: parent.width,
        startup_cost,
        total_cost,
    };
    add_path(&mut root.rels[rel], Rc::new(path));
}

/// index_info returns what the planner knows of an index of a base relation's table: Dolt's primary index holds
/// the table's rows, and a secondary index holds its columns with the primary key's, which a scan reads alone when
/// it covers every column the scan needs.
fn index_info(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    table: &TableDef,
    index: Option<usize>,
    covering: bool,
    nquals: usize,
) -> IndexOptInfo {
    let parent = &root.rels[rel];
    let Some(i) = index else {
        let height = prolly::Node::decode(table.table.primary_index.clone()).map_or(0, |r| r.level());
        return IndexOptInfo {
            pages: parent.pages,
            tuples: parent.tuples,
            tree_height: f64::from(height),
            indexonly: true,
            correlation: 1.0,
            nquals,
        };
    };
    let width = |columns: &[usize]| -> f64 {
        columns.iter().filter_map(|&c| table.index_column(c)).map(|c| get_typavgwidth(c.ty)).sum()
    };
    let index_width = width(&table.indexes[i].columns) + width(&table.key_columns);
    let (tuples, pages) = (parent.tuples, estimate_rel_pages(parent.tuples, index_width));
    let height = root.ctx.db.read(&table.indexes[i].root).map_or(0, |r| r.level());
    let columns = &table.indexes[i].columns;
    let stats = crate::colstats::table_stats(root.ctx, table);
    let first = columns.first().and_then(|&c| stats.as_ref()?.columns.get(c)).map_or(0.0, |c| c.correlation);
    let correlation = if columns.len() > 1 { first * 0.75 } else { first };
    IndexOptInfo { pages, tuples, tree_height: f64::from(height), indexonly: covering, correlation, nquals }
}

/// to_attnos rewrites a restriction of a base relation over its Vars into one over the columns of its rows.
pub fn to_attnos(e: &Expr, rel: usize) -> Expr {
    match e {
        Expr::Column(c) if var_parts(*c).0 == rel => Expr::Column(var_parts(*c).1),
        other => other.clone().map_children(&mut |c| to_attnos(&c, rel)),
    }
}

/// attnos returns the columns of its relation that a clause reads.
fn attnos(e: &Expr) -> Vec<usize> {
    let mut out = Vec::new();
    e.visit(&mut |x| {
        if let Expr::Column(c) = x {
            out.push(var_parts(*c).1);
        }
    });
    out
}

/// positional rewrites an expression over Vars into one over a row of the outer Vars followed by the relation's
/// columns.
fn positional(e: &Expr, outer_vars: &[usize], rel: usize, left_width: usize) -> Expr {
    match e {
        Expr::Column(c) => match var_parts(*c) {
            (r, attno) if r == rel => Expr::Column(left_width + attno),
            _ => Expr::Column(outer_vars.iter().position(|v| v == c).expect("every outer Var has a position")),
        },
        other => other.clone().map_children(&mut |c| positional(&c, outer_vars, rel, left_width)),
    }
}

/// from_positional rewrites an expression over a row of the outer Vars into one over the Vars.
fn from_positional(e: &Expr, outer_vars: &[usize]) -> Expr {
    match e {
        Expr::Column(c) => Expr::Column(outer_vars[*c]),
        other => other.clone().map_children(&mut |c| from_positional(&c, outer_vars)),
    }
}

/// base_vars returns the Vars of a base relation's columns in order.
pub fn base_vars(rel: usize, width: usize) -> Vec<usize> {
    (0..width).map(|attno| var(rel, attno)).collect()
}
