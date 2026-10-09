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
use super::clausesel::clauselist_selectivity;
use super::costsize::{
    IndexCost, QualCost, clamp_row_est, cost_index, cost_qual_eval, estimate_rel_pages, get_typavgwidth,
};
use super::nodes::{
    IndexOptInfo, JoinType, Path, PathKind, RelOptKind, Relids, RestrictInfo, RinfoId, RteKind, VarNode,
};
use super::pathnode::add_path;
use super::var::pull_varnos;
use crate::catalog::table::TableDef;
use crate::expr::{CmpOp, Expr};
use crate::plan::{JoinMethod, Plan, SortKey};

/// create_index_paths adds the paths of a base relation's index scans: the one that Doltgres chooses for its
/// restrictions, one that reads its rows in the order of the query's ORDER BY, and a lookup for each set of other
/// relations whose join equalities find its rows through an index, as Postgres' function of the same name adds plain,
/// ordered, and parameterized index paths.
pub fn create_index_paths(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    if let RteKind::Relation(_, table) = &root.parse.rte(rel).kind {
        let table = table.clone();
        let restricted = create_restriction_index_path(root, rel, &table);
        create_ordered_index_path(root, rel, &table, restricted);
    }
    let mut joinclauses = root.rels[rel].joininfo.clone();
    for clause in match_eclass_clauses_to_index(root, rel) {
        if !joinclauses.contains(&clause) {
            joinclauses.push(clause);
        }
    }
    root.rels[rel].lookup_clauses = joinclauses.clone();
    let mut outer_sets: Vec<Relids> = Vec::new();
    for &rinfo in &joinclauses {
        let Some((_, outer)) = join_equality(&root.rinfos[rinfo], rel) else { continue };
        let relids = pull_varnos(root, outer);
        if !outer_sets.contains(&relids) {
            outer_sets.push(relids);
        }
    }
    let all = outer_sets.iter().fold(Relids::new(), |relids, r| relids.union(r));
    if outer_sets.len() > 1 {
        outer_sets.push(all);
    }
    for outer_relids in outer_sets {
        create_lookup_path(root, rel, outer_relids);
    }
}

/// match_eclass_clauses_to_index returns the join clauses that the equivalence classes imply between each key column
/// of each index of a base relation and the members of other relations, as Postgres' function of the same name finds
/// them for each index.
fn match_eclass_clauses_to_index(root: &mut PlannerInfo<'_, '_>, rel: usize) -> Vec<RinfoId> {
    let mut clauses = Vec::new();
    if !root.rels[rel].has_eclass_joins {
        return clauses;
    }
    let prohibited = root.rels[rel].lateral_referencers.clone();
    for index in root.rels[rel].indexlist.clone() {
        for indexcol in 0..index.indexkeys.len() {
            let callback = |root: &PlannerInfo<'_, '_>, ec: super::nodes::EcId, em: super::nodes::EmId| {
                ec_member_matches_indexcol(root, rel, &index, indexcol, ec, em)
            };
            clauses.extend(super::equivclass::generate_implied_equalities_for_column(
                root,
                rel,
                &callback,
                &prohibited,
            ));
        }
    }
    clauses
}

/// ec_member_matches_indexcol reports whether an equivalence class's member is an index's key column, in the
/// column's operator family, as Postgres' function of the same name does.
fn ec_member_matches_indexcol(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    index: &IndexOptInfo,
    indexcol: usize,
    ec: super::nodes::EcId,
    em: super::nodes::EmId,
) -> bool {
    index.opfamily[indexcol].is_some_and(|f| root.eq_classes[ec].ec_opfamilies.contains(&f))
        && match_index_to_operand(root, &root.eq_members[em].em_expr, indexcol, index, rel)
}

/// join_equality returns the sides of a join clause that equates an expression of a relation with one of others,
/// as an index lookup of that relation's rows may search by.
fn join_equality(rinfo: &RestrictInfo, rel: usize) -> Option<(&Expr, &Expr)> {
    let Expr::Compare(CmpOp::Eq, l, r) = &rinfo.clause else { return None };
    let singleton = Relids::singleton(rel);
    match (rinfo.can_join, rinfo.left_relids == singleton, rinfo.right_relids == singleton) {
        (true, true, false) => Some((l, r)),
        (true, false, true) => Some((r, l)),
        _ => None,
    }
}

/// create_restriction_index_path adds the path of the index scan that Doltgres chooses for a table's restrictions,
/// returning its index. When the query only counts the rows of that one table that the scan's ranges hold exactly,
/// Doltgres' executor counts them from the index's entry counts, so the scan costs only its descent.
fn create_restriction_index_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    table: &TableDef,
) -> Option<Option<usize>> {
    let restrictinfo = root.rels[rel].baserestrictinfo.clone();
    let predicate = restrictinfo
        .iter()
        .filter(|&&r| !root.rinfos[r].pseudoconstant)
        .map(|&r| to_attnos(root, root.rinfos[r].clause.clone(), rel))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let predicate = predicate?;
    let (scan, exact) = crate::indexscan::choose_with_cover(root.ctx, table, &predicate)?;
    let chosen = scan.index;
    let ordered =
        local_pathkeys(root, rel).and_then(|keys| ordered_scan(Plan::IndexScan(Box::new(scan.clone())), &keys));
    let (scan, pathkeys) = match ordered {
        Some(ordered) => (ordered, root.query_pathkeys.clone()),
        None => (scan, Vec::new()),
    };
    let columns = scan.index_columns();
    let on_index = |r: &RinfoId| {
        let clause = &root.rinfos[*r].clause;
        pull_varnos(root, clause).num_members() == 1 && attnos(root, clause).iter().all(|a| columns.contains(a))
    };
    let index_quals: Vec<RinfoId> = restrictinfo.iter().copied().filter(on_index).collect();
    let selectivity = clauselist_selectivity(root, &index_quals, rel, JoinType::Inner, None);
    let index_tuples = clamp_row_est(selectivity * root.rels[rel].tuples);
    let nquals = index_quals.len();
    let qpquals: Vec<RinfoId> = match exact {
        true => Vec::new(),
        false => restrictinfo.iter().copied().filter(|r| !on_index(r)).collect(),
    };
    let index = index_info(root, rel, table, scan.index, check_index_only(root, rel, &scan), nquals);
    let (disabled_nodes, startup_cost, mut total_cost) =
        cost_index(root, rel, &index, index_tuples, cost_qual_eval(root, &qpquals), 1.0);
    let counted = root.counting.as_ref().is_some_and(|calls| {
        !calls.is_empty() && calls.iter().all(|call| call.counts_rows() || call.counts_set_column(table))
    });
    if exact && counted && root.all_baserels == Relids::singleton(rel) {
        total_cost = startup_cost;
    }
    let parent = &mut root.rels[rel];
    let path = Path {
        kind: PathKind::IndexScan(Box::new(scan), exact),
        parent: rel,
        relids: parent.relids.clone(),
        param: Relids::new(),
        pathkeys,
        rows: parent.rows,
        width: parent.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
    };
    add_path(parent, Rc::new(path));
    Some(chosen)
}

/// create_ordered_index_path adds the path of a scan of every entry of an index of a table that reads its rows in
/// the order of the query's ORDER BY, testing the table's restrictions on each, as Postgres adds an index path for its
/// useful pathkeys alone, unless it is the index that the restrictions' scan reads, whose path has its order already.
fn create_ordered_index_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    table: &TableDef,
    restricted: Option<Option<usize>>,
) {
    let Some(keys) = local_pathkeys(root, rel) else { return };
    let Some(scan) = ordered_scan(Plan::Scan(Box::new(table.clone()), None), &keys) else { return };
    if restricted == Some(scan.index) {
        return;
    }
    let index = index_info(root, rel, table, scan.index, check_index_only(root, rel, &scan), 0);
    let qpqual_cost = cost_qual_eval(root, &root.rels[rel].baserestrictinfo);
    let tuples = root.rels[rel].tuples;
    let (disabled_nodes, startup_cost, total_cost) = cost_index(root, rel, &index, tuples, qpqual_cost, 1.0);
    let parent = &root.rels[rel];
    let path = Path {
        kind: PathKind::IndexScan(Box::new(scan), false),
        parent: rel,
        relids: parent.relids.clone(),
        param: Relids::new(),
        pathkeys: root.query_pathkeys.clone(),
        rows: parent.rows,
        width: parent.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
    };
    add_path(&mut root.rels[rel], Rc::new(path));
}

/// local_pathkeys returns the query's ORDER BY keys over the columns of a base relation, when they read only it.
fn local_pathkeys(root: &PlannerInfo<'_, '_>, rel: usize) -> Option<Vec<SortKey>> {
    let singleton = Relids::singleton(rel);
    if root.query_pathkeys.is_empty() || root.query_sortkeys.iter().any(|k| pull_varnos(root, &k.expr) != singleton) {
        return None;
    }
    Some(
        root.query_sortkeys
            .iter()
            .map(|k| SortKey { expr: to_attnos(root, k.expr.clone(), rel), ..k.clone() })
            .collect(),
    )
}

/// ordered_scan returns the index scan that reads a scan's rows in the order of keys over its columns, as Doltgres'
/// index scans can.
fn ordered_scan(plan: Plan, keys: &[SortKey]) -> Option<crate::indexscan::IndexScan> {
    match crate::indexscan::ordered(&plan, keys)? {
        Plan::IndexScan(scan) => Some(*scan),
        _ => None,
    }
}

/// create_lookup_path adds the path of a lookup of a relation's rows by its join equalities with a set of other
/// relations, when an index of its table or catalog lets it look them up, as Postgres adds the parameterized index
/// path of the clauses that those relations' rows supply. A relation that evaluates PlaceHolderVars has none, as
/// Doltgres' lookup joins return the looked-up rows as they are stored.
fn create_lookup_path(root: &mut PlannerInfo<'_, '_>, rel: usize, outer_relids: Relids) {
    let placeholders = root.rels[rel]
        .reltarget
        .exprs
        .iter()
        .any(|e| matches!(e, Expr::Column(id) if matches!(root.glob.node(*id), VarNode::PlaceHolderVar(_))));
    if placeholders {
        return;
    }
    let mut outer_vars: Vec<usize> = Vec::new();
    let mut equalities = Vec::new();
    for &rinfo in &root.rels[rel].lookup_clauses {
        let Some((inner, outer)) = join_equality(&root.rinfos[rinfo], rel) else { continue };
        if !pull_varnos(root, outer).is_subset(&outer_relids) {
            continue;
        }
        super::var::visit_columns(outer, &mut |v| {
            if !outer_vars.contains(&v) {
                outer_vars.push(v);
            }
        });
        equalities.push((inner.clone(), outer.clone()));
    }
    let left_width = outer_vars.len();
    let position = |e: &Expr| -> Expr { positional(root, e.clone(), &outer_vars, rel, left_width) };
    let condition = equalities
        .iter()
        .map(|(inner, outer)| Expr::Compare(CmpOp::Eq, Box::new(position(outer)), Box::new(position(inner))))
        .reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
    let Some(condition) = condition else { return };
    let plan = match &root.parse.rte(rel).kind {
        RteKind::Relation(plan, _) | RteKind::Plan(plan) => plan.clone(),
        _ => return,
    };
    let Some(found) = crate::joins::lookup(root.ctx, &plan, &condition, left_width) else { return };
    let method = match found.method {
        JoinMethod::Lookup { scan, keys } => {
            JoinMethod::Lookup { scan, keys: keys.iter().map(|k| from_positional(k.clone(), &outer_vars)).collect() }
        }
        JoinMethod::CatalogLookup { index, keys } => JoinMethod::CatalogLookup {
            index,
            keys: keys.iter().map(|k| from_positional(k.clone(), &outer_vars)).collect(),
        },
        other => other,
    };
    let loop_count = outer_relids
        .members()
        .filter(|&r| root.rels[r].reloptkind == RelOptKind::BaseRel)
        .map(|r| root.rels[r].rows)
        .fold(f64::INFINITY, f64::min);
    let loop_count = if loop_count.is_finite() { loop_count } else { 1.0 };
    let index = match (&method, &root.parse.rte(rel).kind) {
        (JoinMethod::Lookup { scan, .. }, RteKind::Relation(_, table)) => {
            let table = table.clone();
            index_info(root, rel, &table, scan.index, false, 1)
        }
        _ => IndexCost {
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
    let qpqual_cost: QualCost = cost_qual_eval(root, &parent.baserestrictinfo);
    let (disabled_nodes, startup_cost, total_cost) =
        cost_index(root, rel, &index, found.matches, qpqual_cost, loop_count);
    let parent = &root.rels[rel];
    let path = Path {
        kind: PathKind::Lookup(method),
        parent: rel,
        relids: parent.relids.clone(),
        param: outer_relids,
        pathkeys: Vec::new(),
        rows,
        width: parent.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
    };
    add_path(&mut root.rels[rel], Rc::new(path));
}

/// check_index_only reports whether an index scan of a base relation reads every column that the query needs of
/// the relation, from what the joins and output above it read and its restrictions, as Postgres' function of the same
/// name decides an index-only scan.
fn check_index_only(root: &PlannerInfo<'_, '_>, rel: usize, scan: &crate::indexscan::IndexScan) -> bool {
    let parent = &root.rels[rel];
    let mut needed: Vec<usize> = (0..parent.attr_needed.len()).filter(|&a| !parent.attr_needed[a].is_empty()).collect();
    for &rinfo in &parent.baserestrictinfo {
        needed.extend(attnos(root, &root.rinfos[rinfo].clause));
    }
    scan.covers(&needed)
}

/// index_info returns what the planner knows of an index of a base relation's table for its costs: Dolt's primary
/// index holds the table's rows, and a secondary index holds its columns with the primary key's, which a scan reads
/// alone when it covers every column the scan needs.
fn index_info(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    table: &TableDef,
    index: Option<usize>,
    covering: bool,
    nquals: usize,
) -> IndexCost {
    let parent = &root.rels[rel];
    let Some(i) = index else {
        let height = prolly::Node::decode(table.table.primary_index.clone()).map_or(0, |r| r.level());
        return IndexCost {
            pages: parent.pages,
            tuples: parent.tuples,
            tree_height: f64::from(height),
            indexonly: true,
            correlation: 1.0,
            nquals,
        };
    };
    let width = |columns: &[usize]| -> f64 {
        columns
            .iter()
            .filter_map(|&c| table.index_column(c))
            .map(|c| get_typavgwidth(Some(c.ty.oid), c.ty.modifier))
            .sum()
    };
    let index_width = width(&table.indexes[i].columns) + width(&table.key_columns);
    let (tuples, pages) = (parent.tuples, estimate_rel_pages(parent.tuples, index_width));
    let height = root.ctx.db.read(&table.indexes[i].root).map_or(0, |r| r.level());
    let columns = &table.indexes[i].columns;
    let stats = root.rels[rel].stats.as_ref();
    let first = columns.first().and_then(|&c| stats?.columns.get(c)).map_or(0.0, |c| c.correlation);
    let correlation = if columns.len() > 1 { first * 0.75 } else { first };
    IndexCost { pages, tuples, tree_height: f64::from(height), indexonly: covering, correlation, nquals }
}

/// get_relation_indexes returns the indexes of a table, as Postgres' get_relation_info lists them: Dolt's primary
/// index, unless the table is keyless, and then each secondary index, with each key column's btree operator family.
pub fn get_relation_indexes(table: &TableDef) -> Vec<IndexOptInfo> {
    let opfamily = |c: usize| table.columns.get(c).and_then(|col| super::nodefuncs::btree_opfamily(col.ty.oid));
    let key = |c: usize| (c < table.columns.len()).then_some(c);
    let mut indexlist = Vec::new();
    if !table.keyless() {
        indexlist.push(IndexOptInfo {
            indexkeys: table.key_columns.iter().map(|&c| key(c)).collect(),
            opfamily: table.key_columns.iter().map(|&c| opfamily(c)).collect(),
            unique: true,
            has_predicate: false,
        });
    }
    for index in &table.indexes {
        indexlist.push(IndexOptInfo {
            indexkeys: index.columns.iter().map(|&c| key(c)).collect(),
            opfamily: index.columns.iter().map(|&c| opfamily(c)).collect(),
            unique: index.unique,
            has_predicate: !index.predicate.is_empty(),
        });
    }
    indexlist
}

/// relation_has_unique_index_ext reports whether a unique index of a base relation has each of its key columns
/// equated with something by one of a list of clauses, whose outer sides are given, or by one of the relation's
/// restrictions, returning the restrictions that it used when asked, as Postgres' function of the same name does.
pub fn relation_has_unique_index_ext(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    restrictlist: &[RinfoId],
    extra_clauses: Option<&mut Vec<RinfoId>>,
) -> bool {
    if root.rels[rel].indexlist.is_empty() {
        return false;
    }
    let mut restrictlist = restrictlist.to_vec();
    for &rinfo in &root.rels[rel].baserestrictinfo {
        let r = &root.rinfos[rinfo];
        if r.mergeopfamilies.is_empty() {
            continue;
        }
        if r.left_relids.is_empty() {
            r.outer_is_left.set(true);
        } else if r.right_relids.is_empty() {
            r.outer_is_left.set(false);
        } else {
            continue;
        }
        restrictlist.push(rinfo);
    }
    if restrictlist.is_empty() {
        return false;
    }
    for ind in &root.rels[rel].indexlist {
        if !ind.unique || ind.has_predicate {
            continue;
        }
        let mut exprs = Vec::new();
        let all_matched = (0..ind.indexkeys.len()).all(|c| {
            restrictlist.iter().any(|&rinfo| {
                let r = &root.rinfos[rinfo];
                if !ind.opfamily[c].is_some_and(|f| r.mergeopfamilies.contains(&f)) {
                    return false;
                }
                let Some((left, right)) = super::restrictinfo::binary_op_args(&r.clause) else { return false };
                let rexpr = if r.outer_is_left.get() { right } else { left };
                if !match_index_to_operand(root, rexpr, c, ind, rel) {
                    return false;
                }
                if r.clause_relids.num_members() == 1 {
                    exprs.push(rinfo);
                }
                true
            })
        });
        if all_matched {
            if let Some(extra) = extra_clauses {
                *extra = exprs;
            }
            return true;
        }
    }
    false
}

/// match_index_to_operand reports whether an expression is an index's key column, as a Var of the index's relation
/// that no outer join makes NULL, as Postgres' function of the same name does for column keys.
pub fn match_index_to_operand(
    root: &PlannerInfo<'_, '_>,
    operand: &Expr,
    indexcol: usize,
    index: &IndexOptInfo,
    rel: usize,
) -> bool {
    let Expr::Column(id) = operand else { return false };
    let VarNode::Var(var) = root.glob.node(*id) else { return false };
    index.indexkeys[indexcol].is_some_and(|k| var.varno == rel && var.varattno == k && var.varnullingrels.is_empty())
}

/// to_attnos rewrites a restriction of a base relation over its Vars into one over the columns of its rows, computing
/// its PlaceHolderVars from their expressions, as the relation where they are evaluated does.
pub fn to_attnos(root: &PlannerInfo<'_, '_>, e: Expr, rel: usize) -> Expr {
    match e {
        Expr::Column(id) => match root.glob.node(id) {
            VarNode::Var(var) if var.varno == rel => Expr::Column(var.varattno),
            VarNode::PlaceHolderVar(phv) => to_attnos(root, root.glob.placeholder(phv.phid).phexpr.clone(), rel),
            VarNode::Var(_) => Expr::Column(id),
        },
        other => other.map_children(&mut |c| to_attnos(root, c, rel)),
    }
}

/// attnos returns the columns of its relation that a clause reads.
fn attnos(root: &PlannerInfo<'_, '_>, e: &Expr) -> Vec<usize> {
    let mut out = Vec::new();
    super::var::visit_columns(e, &mut |id| {
        if let VarNode::Var(var) = root.glob.node(id) {
            out.push(var.varattno);
        }
    });
    out
}

/// positional rewrites an expression over Vars into one over a row of the outer Vars followed by the relation's
/// columns.
fn positional(root: &PlannerInfo<'_, '_>, e: Expr, outer_vars: &[usize], rel: usize, left_width: usize) -> Expr {
    match e {
        Expr::Column(id) => match (root.glob.node(id), outer_vars.iter().position(|&v| v == id)) {
            (VarNode::Var(var), _) if var.varno == rel => Expr::Column(left_width + var.varattno),
            (_, Some(i)) => Expr::Column(i),
            (VarNode::PlaceHolderVar(phv), None) => {
                positional(root, root.glob.placeholder(phv.phid).phexpr.clone(), outer_vars, rel, left_width)
            }
            (VarNode::Var(_), None) => unreachable!("every outer Var has a position"),
        },
        other => other.map_children(&mut |c| positional(root, c, outer_vars, rel, left_width)),
    }
}

/// from_positional rewrites an expression over a row of the outer Vars into one over the Vars.
fn from_positional(e: Expr, outer_vars: &[usize]) -> Expr {
    match e {
        Expr::Column(c) => Expr::Column(outer_vars[c]),
        other => other.map_children(&mut |c| from_positional(c, outer_vars)),
    }
}
