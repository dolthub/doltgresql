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

//! Postgres' optimizer/util/relnode.c, with the parts of plancat.c that size a table: building the base relations
//! and the join relations.

use std::rc::Rc;

use super::PlannerInfo;
use super::costsize::{calc_joinrel_size_estimate, estimate_rel_pages, get_typavgwidth};
use super::nodes::{RelOptInfo, Relids, RestrictInfo, SpecialJoinInfo, is_subset, singleton};

/// OPAQUE_COLUMN_WIDTH is the width that Postgres assumes for a value of a variable-length type, which is every
/// column of a relation whose types the planner does not look up.
const OPAQUE_COLUMN_WIDTH: f64 = 32.0;

impl PlannerInfo<'_, '_> {
    /// find_rel returns the index of the base or join relation of a set of relations, when it was built.
    pub fn find_rel(&self, relids: Relids) -> Option<usize> {
        match relids.count_ones() {
            1 => Some(relids.trailing_zeros() as usize),
            _ => self.join_rel_hash.get(&relids).copied(),
        }
    }
}

/// add_base_rels_to_query builds the base relation of each range table entry, as Postgres' function of the same name
/// does, after a placeholder at index 0 so that each lies at its range table index, and marks the columns that the
/// query's output reads as needed there, as Postgres' build_base_rel_tlists does.
pub fn add_base_rels_to_query(root: &mut PlannerInfo<'_, '_>) {
    root.rels.push(RelOptInfo::default());
    for varno in 1..=root.parse.rtable.len() {
        let rel = build_simple_rel(root, varno);
        root.rels.push(rel);
    }
    for e in root.parse.output.clone().into_iter().flatten() {
        add_vars_to_targetlist(root, &e, singleton(0));
    }
}

/// add_vars_to_targetlist marks the columns that an expression reads as needed by a set of relations, as Postgres'
/// function of the same name does.
pub fn add_vars_to_targetlist(root: &mut PlannerInfo<'_, '_>, e: &crate::expr::Expr, where_needed: Relids) {
    e.visit(&mut |x| {
        if let crate::expr::Expr::Column(c) = x {
            let (varno, attno) = super::nodes::var_parts(*c);
            root.rels[varno].attr_needed[attno] |= where_needed;
        }
    });
}

/// build_simple_rel builds the base relation of a range table entry, sized as Postgres' get_relation_info and
/// estimate_rel_size size a table: its rows from its primary index, and its pages from their width. Any other entry's
/// rows are the older planner's estimate.
fn build_simple_rel(root: &mut PlannerInfo<'_, '_>, varno: usize) -> RelOptInfo {
    let rte = root.parse.rte(varno);
    let (tuples, width, pages, stats) = match rte.table() {
        Some(table) => {
            let tuples = prolly::Node::decode(table.table.primary_index.clone()).map_or(0.0, |r| r.tree_count() as f64);
            let width = table.columns.iter().map(|c| get_typavgwidth(c.ty)).sum();
            (tuples, width, estimate_rel_pages(tuples, width), crate::colstats::table_stats(root.ctx, table))
        }
        None => {
            let width = rte.plan.width() as f64 * OPAQUE_COLUMN_WIDTH;
            (crate::joins::estimate(root.ctx, &rte.plan), width, 0.0, None)
        }
    };
    let attr_needed = vec![0; rte.plan.width()];
    let consider_startup = root.tuple_fraction > 0.0;
    RelOptInfo {
        relids: singleton(varno),
        relid: varno,
        tuples,
        width,
        pages,
        stats,
        attr_needed,
        consider_startup,
        ..RelOptInfo::default()
    }
}

/// build_join_rel returns the join relation of two relations, building it with its clauses and size when it is new,
/// with the clauses that the join evaluates, as Postgres' build_join_rel does.
pub fn build_join_rel(
    root: &mut PlannerInfo<'_, '_>,
    joinrelids: Relids,
    outer_rel: usize,
    inner_rel: usize,
    sjinfo: &SpecialJoinInfo,
) -> (usize, Vec<Rc<RestrictInfo>>) {
    if let Some(joinrel) = root.join_rel_hash.get(&joinrelids).copied() {
        return (joinrel, build_joinrel_restrictlist(root, joinrelids, outer_rel, inner_rel));
    }
    let mut joininfo = Vec::new();
    for rel in [outer_rel, inner_rel] {
        for rinfo in &root.rels[rel].joininfo {
            if !is_subset(rinfo.required_relids, joinrelids) && !joininfo.iter().any(|r| Rc::ptr_eq(r, rinfo)) {
                joininfo.push(rinfo.clone());
            }
        }
    }
    let restrictlist = build_joinrel_restrictlist(root, joinrelids, outer_rel, inner_rel);
    let (outer_rows, inner_rows) = (root.rels[outer_rel].rows, root.rels[inner_rel].rows);
    let rows = calc_joinrel_size_estimate(root, joinrelids, outer_rows, inner_rows, sjinfo, &restrictlist);
    let joinrel = RelOptInfo {
        relids: joinrelids,
        rows,
        width: root.rels[outer_rel].width + root.rels[inner_rel].width,
        joininfo,
        consider_startup: root.tuple_fraction > 0.0,
        ..RelOptInfo::default()
    };
    let index = root.rels.len();
    root.rels.push(joinrel);
    root.join_rel_hash.insert(joinrelids, index);
    if let Some(level) = root.join_rel_level.last_mut() {
        level.push(index);
    }
    (index, restrictlist)
}

/// build_joinrel_restrictlist returns the join clauses of two relations that a join of them can evaluate, as
/// Postgres' function of the same name does.
fn build_joinrel_restrictlist(
    root: &PlannerInfo<'_, '_>,
    joinrelids: Relids,
    outer_rel: usize,
    inner_rel: usize,
) -> Vec<Rc<RestrictInfo>> {
    let mut restrictlist: Vec<Rc<RestrictInfo>> = Vec::new();
    for rel in [outer_rel, inner_rel] {
        for rinfo in &root.rels[rel].joininfo {
            if is_subset(rinfo.required_relids, joinrelids) && !restrictlist.iter().any(|r| Rc::ptr_eq(r, rinfo)) {
                restrictlist.push(rinfo.clone());
            }
        }
    }
    restrictlist
}
