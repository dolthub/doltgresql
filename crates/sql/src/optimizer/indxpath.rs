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

//! Postgres' optimizer/path/indxpath.c: the index paths of a base relation, from the clauses that each column of each
//! of its table's indexes can search by. Doltgres' index scans read the ranges that those clauses give, and its
//! lookup joins look rows up by the equalities of an index's leading columns, which a parameterized path's join
//! clauses must give. A system catalog's rows are an opaque plan, whose lookups Doltgres' lookup joins find.

use std::rc::Rc;

use prolly::NodeStore;

use super::PlannerInfo;
use super::costsize::{
    IndexCost, QualCost, clamp_row_est, cost_bitmap_heap_scan, cost_bitmap_tree_node, cost_catalog_lookup, cost_index,
    cost_qual_eval,
};
use super::equivclass::generate_implied_equalities_for_column;
use super::nodes::{
    EcId, EmId, IndexClause, IndexOptInfo, IndexPath, JoinType, Path, PathKind, RelOptKind, Relids, RinfoId, RteKind,
    VarNode,
};
use super::pathkeys::{build_index_pathkeys, truncate_useless_pathkeys};
use super::pathnode::{add_path, create_bitmap_and_path, create_bitmap_heap_path, create_bitmap_or_path};
use super::predtest;
use super::restrictinfo::{
    RestrictInfoArgs, binary_op_args, join_clause_is_movable_to, make_plain_restrictinfo, make_restrictinfo,
    restriction_is_or_clause,
};
use super::var::pull_varnos;
use crate::catalog::table::TableDef;
use crate::expr::{CmpOp, Expr};
use crate::plan::JoinMethod;

/// IndexClauseSet is the index clauses of each key column of an index, as Postgres' IndexClauseSet is.
type IndexClauseSet = Vec<Vec<IndexClause>>;

/// get_relation_indexes returns the indexes of a base relation's table, as Postgres' get_relation_info lists them:
/// Dolt's primary index, unless the table is keyless, and then each secondary index that is not a vector index, with
/// each key column's btree operator family, direction, and NULL placement, and a secondary index's primary key
/// columns after its own, which order its entries too.
pub fn get_relation_indexes(root: &mut PlannerInfo<'_, '_>, rel: usize, table: &TableDef) -> Vec<Rc<IndexOptInfo>> {
    let rules = root.ctx.index_rules(table).ok();
    let glob = &mut *root.glob;
    let mut to_vars = |e: &Expr| super::var::replace_columns(e.clone(), &mut |c| glob.var(rel, c, Relids::new()));
    let mut hidden = Vec::new();
    let mut predicates: Vec<Vec<Expr>> = Vec::new();
    if let Some(rules) = &rules {
        hidden = rules.hidden().iter().map(&mut to_vars).collect();
        for predicate in rules.predicates() {
            let conjuncts = predicate.iter().flat_map(crate::indexscan::conjuncts);
            predicates.push(conjuncts.map(&mut to_vars).collect());
        }
    }
    let width = |columns: &[usize]| -> f64 {
        columns
            .iter()
            .filter_map(|&c| table.index_column(c))
            .map(|c| super::costsize::get_typavgwidth(Some(c.ty.oid), c.ty.modifier))
            .sum()
    };
    let (rel_pages, rel_tuples) = (root.rels[rel].pages, root.rels[rel].tuples);
    let mut indexlist = Vec::new();
    let mut add = |root: &mut PlannerInfo<'_, '_>, index: Option<usize>, columns: &[usize], unique: bool| {
        let pk: Vec<usize> = match index {
            Some(_) => table.key_columns.iter().copied().filter(|c| !columns.contains(c)).collect(),
            None => Vec::new(),
        };
        let all: Vec<usize> = columns.iter().chain(&pk).copied().collect();
        let (reverse_sort, nulls_first): (Vec<bool>, Vec<bool>) = match index {
            Some(i) => {
                let def = &table.indexes[i];
                let own = (0..columns.len()).map(|c| (def.descending[c], !def.nulls_last[c]));
                own.chain(pk.iter().map(|_| (false, false))).unzip()
            }
            None => all.iter().map(|_| (false, false)).unzip(),
        };
        let indexkeys: Vec<Option<usize>> = all.iter().map(|&c| (c < table.columns.len()).then_some(c)).collect();
        let indexprs = all
            .iter()
            .filter_map(|&c| c.checked_sub(crate::catalog::table::HIDDEN_BASE))
            .map(|k| hidden[k].clone())
            .collect();
        let opfamily = all
            .iter()
            .map(|&c| table.index_column(c).and_then(|col| super::nodefuncs::btree_opfamily(col.ty.oid)))
            .collect();
        let (pages, tree_height) = match index {
            Some(i) => {
                let height = root.ctx.db.read(&table.indexes[i].root).map_or(0, |r| r.level());
                (super::costsize::estimate_rel_pages(rel_tuples, width(columns) + width(&table.key_columns)), height)
            }
            None => (rel_pages, prolly::Node::decode(table.table.primary_index.clone()).map_or(0, |r| r.level())),
        };
        let json = all.iter().any(|&c| {
            table.index_column(c).is_some_and(|col| matches!(col.ty.oid, crate::oid::JSON | crate::oid::JSONB))
        });
        indexlist.push(Rc::new(IndexOptInfo {
            index,
            pages,
            tuples: rel_tuples,
            tree_height: f64::from(tree_height),
            nkeycolumns: columns.len(),
            indexkeys,
            indexprs,
            opfamily,
            reverse_sort,
            nulls_first,
            sortable: !json && !crate::indexscan::hash_ordered(table, index),
            unique,
            indpred: index.and_then(|i| predicates.get(i).cloned()).unwrap_or_default(),
            pred_ok: false,
            indrestrictinfo: Vec::new(),
        }));
    };
    if !table.keyless() {
        add(root, None, &table.key_columns, true);
    }
    for (i, def) in table.indexes.iter().enumerate() {
        if def.vector.is_none() {
            add(root, Some(i), &def.columns, def.unique);
        }
    }
    indexlist
}

/// ScanTypeControl is which kinds of scan build_index_paths makes paths for, as Postgres' ScanTypeControl is.
#[derive(Clone, Copy, PartialEq, Eq)]
enum ScanTypeControl {
    BitmapScan,
    AnyScan,
}

/// create_index_paths adds the paths of a base relation's index scans: for each index, a scan by the restrictions
/// that its columns can search by, and scans parameterized by each set of other relations whose join clauses it can
/// search by, and then the bitmap heap scans of the most promising combination of those scans and of the bitmap scans
/// of OR clauses, unparameterized and for each parameterization, as Postgres' function of the same name does. A
/// system catalog's lookups come from Doltgres' lookup joins.
pub fn create_index_paths(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    if !matches!(root.parse.rte(rel).kind, RteKind::Relation(..)) {
        create_catalog_lookup_paths(root, rel);
        return;
    }
    let (mut bitindexpaths, mut bitjoinpaths, mut joinorclauses) = (Vec::new(), Vec::new(), Vec::new());
    for index in 0..root.rels[rel].indexlist.len() {
        let info = &root.rels[rel].indexlist[index];
        if !info.indpred.is_empty() && !info.pred_ok {
            continue;
        }
        let rclauseset = match_restriction_clauses_to_index(root, rel, index);
        get_index_paths(root, rel, index, &rclauseset, &mut bitindexpaths);
        let jclauseset = match_join_clauses_to_index(root, rel, index, &mut joinorclauses);
        let eclauseset = match_eclass_clauses_to_index(root, rel, index);
        let nonempty = |set: &IndexClauseSet| set.iter().any(|c| !c.is_empty());
        if nonempty(&jclauseset) || nonempty(&eclauseset) {
            consider_index_join_clauses(root, rel, index, &rclauseset, &jclauseset, &eclauseset, &mut bitjoinpaths);
        }
    }
    let baserestrictinfo = root.rels[rel].baserestrictinfo.clone();
    bitindexpaths.extend(generate_bitmap_or_paths(root, rel, &baserestrictinfo, &[]));
    bitjoinpaths.extend(generate_bitmap_or_paths(root, rel, &joinorclauses, &baserestrictinfo));
    if !bitindexpaths.is_empty() {
        let bitmapqual = choose_bitmap_and(root, rel, bitindexpaths.clone());
        let lateral_relids = root.rels[rel].lateral_relids.clone();
        let bpath = create_bitmap_heap_path(root, rel, bitmapqual, &lateral_relids, 1.0);
        add_path(&mut root.rels[rel], bpath);
    }
    let mut all_path_outers: Vec<Relids> = Vec::new();
    for path in &bitjoinpaths {
        if !all_path_outers.contains(&path.param) {
            all_path_outers.push(path.param.clone());
        }
    }
    for max_outers in all_path_outers {
        let mut this_path_set: Vec<Rc<Path>> =
            bitjoinpaths.iter().filter(|p| p.param.is_subset(&max_outers)).cloned().collect();
        this_path_set.extend(bitindexpaths.iter().cloned());
        let bitmapqual = choose_bitmap_and(root, rel, this_path_set);
        let required_outer = bitmapqual.param.clone();
        let loop_count = get_loop_count(root, rel, &required_outer);
        let bpath = create_bitmap_heap_path(root, rel, bitmapqual, &required_outer, loop_count);
        add_path(&mut root.rels[rel], bpath);
    }
}

/// consider_index_join_clauses builds the parameterized paths of an index for each set of outer relations that its
/// join clauses read, as Postgres' function of the same name does.
fn consider_index_join_clauses(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    index: usize,
    rclauseset: &IndexClauseSet,
    jclauseset: &IndexClauseSet,
    eclauseset: &IndexClauseSet,
    bitindexpaths: &mut Vec<Rc<Path>>,
) {
    let mut considered_clauses = 0;
    let mut considered_relids: Vec<Relids> = Vec::new();
    for indexcol in 0..root.rels[rel].indexlist[index].nkeycolumns {
        for set in [jclauseset, eclauseset] {
            considered_clauses += set[indexcol].len();
            consider_index_join_outer_rels(
                root,
                rel,
                index,
                [rclauseset, jclauseset, eclauseset],
                bitindexpaths,
                &set[indexcol],
                considered_clauses,
                &mut considered_relids,
            );
        }
    }
}

/// consider_index_join_outer_rels builds the parameterized paths of an index for the outer relations of each of a
/// column's join clauses, alone and with each set considered before, as Postgres' function of the same name does.
#[allow(clippy::too_many_arguments)]
fn consider_index_join_outer_rels(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    index: usize,
    clausesets: [&IndexClauseSet; 3],
    bitindexpaths: &mut Vec<Rc<Path>>,
    indexjoinclauses: &[IndexClause],
    considered_clauses: usize,
    considered_relids: &mut Vec<Relids>,
) {
    for iclause in indexjoinclauses {
        let clause_relids = root.rinfos[iclause.rinfo].clause_relids.clone();
        let parent_ec = root.rinfos[iclause.rinfo].parent_ec;
        if considered_relids.contains(&clause_relids) {
            continue;
        }
        let num_considered_relids = considered_relids.len();
        for pos in 0..num_considered_relids {
            let oldrelids = considered_relids[pos].clone();
            if clause_relids.subset_compare(&oldrelids) != super::nodes::SubsetCompare::Different {
                continue;
            }
            if parent_ec.is_some_and(|ec| eclass_already_used(root, ec, &oldrelids, indexjoinclauses)) {
                continue;
            }
            if considered_relids.len() >= 10 * considered_clauses {
                break;
            }
            let relids = clause_relids.union(&oldrelids);
            get_join_index_paths(root, rel, index, clausesets, bitindexpaths, &relids, considered_relids);
        }
        get_join_index_paths(root, rel, index, clausesets, bitindexpaths, &clause_relids, considered_relids);
    }
}

/// get_join_index_paths builds the paths of an index parameterized by a set of outer relations, from the join
/// clauses that those relations supply and the restrictions, as Postgres' function of the same name does.
fn get_join_index_paths(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    index: usize,
    [rclauseset, jclauseset, eclauseset]: [&IndexClauseSet; 3],
    bitindexpaths: &mut Vec<Rc<Path>>,
    relids: &Relids,
    considered_relids: &mut Vec<Relids>,
) {
    if considered_relids.contains(relids) {
        return;
    }
    let mut clauseset: IndexClauseSet = vec![Vec::new(); jclauseset.len()];
    for indexcol in 0..jclauseset.len() {
        for iclause in &jclauseset[indexcol] {
            if root.rinfos[iclause.rinfo].clause_relids.is_subset(relids) {
                clauseset[indexcol].push(iclause.clone());
            }
        }
        if let Some(iclause) =
            eclauseset[indexcol].iter().find(|iclause| root.rinfos[iclause.rinfo].clause_relids.is_subset(relids))
        {
            clauseset[indexcol].push(iclause.clone());
        }
        clauseset[indexcol].extend(rclauseset[indexcol].iter().cloned());
    }
    get_index_paths(root, rel, index, &clauseset, bitindexpaths);
    considered_relids.push(relids.clone());
}

/// eclass_already_used reports whether a join clause from an equivalence class was already used for a subset of a
/// set of outer relations, as Postgres' function of the same name does.
fn eclass_already_used(
    root: &PlannerInfo<'_, '_>,
    parent_ec: EcId,
    oldrelids: &Relids,
    indexjoinclauses: &[IndexClause],
) -> bool {
    indexjoinclauses.iter().any(|iclause| {
        let rinfo = &root.rinfos[iclause.rinfo];
        rinfo.parent_ec == Some(parent_ec) && rinfo.clause_relids.is_subset(oldrelids)
    })
}

/// get_index_paths adds the index paths that build_index_paths makes of an index and its clauses, and collects those
/// that a bitmap scan may use, the ones that do not read the whole index for its order, as Postgres' function of the
/// same name does.
fn get_index_paths(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    index: usize,
    clauses: &IndexClauseSet,
    bitindexpaths: &mut Vec<Rc<Path>>,
) {
    let useful_predicate = root.rels[rel].indexlist[index].pred_ok;
    for path in build_index_paths(root, rel, index, clauses, useful_predicate, ScanTypeControl::AnyScan) {
        let PathKind::IndexScan(ipath) = &path.kind else { unreachable!("build_index_paths makes index scans") };
        add_path(&mut root.rels[rel], path.clone());
        if path.pathkeys.is_empty() || ipath.indexselectivity < 1.0 {
            bitindexpaths.push(path);
        }
    }
}

/// build_index_paths makes the paths of a scan of an index by its clauses: a forward scan whose order may be useful,
/// and a backward one when its order is, or only a scan for a bitmap, which reads in no order, as Postgres' function
/// of the same name does.
fn build_index_paths(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    index: usize,
    clauses: &IndexClauseSet,
    useful_predicate: bool,
    scantype: ScanTypeControl,
) -> Vec<Rc<Path>> {
    let mut index_clauses: Vec<IndexClause> = Vec::new();
    let mut outer_relids = root.rels[rel].lateral_relids.clone();
    for indexcol_clauses in clauses {
        for iclause in indexcol_clauses {
            index_clauses.push(iclause.clone());
            outer_relids.add_members(&root.rinfos[iclause.rinfo].clause_relids);
        }
    }
    outer_relids.del_member(rel);
    let loop_count = get_loop_count(root, rel, &outer_relids);
    let pathkeys_possibly_useful = scantype != ScanTypeControl::BitmapScan
        && outer_relids.is_empty()
        && super::pathkeys::has_useful_pathkeys(root, rel);
    let index_is_ordered = root.rels[rel].indexlist[index].sortable;
    let index_only_scan = scantype != ScanTypeControl::BitmapScan && check_index_only(root, rel, index);
    let mut result = Vec::new();
    let useful_pathkeys = match index_is_ordered && pathkeys_possibly_useful {
        true => {
            let index_pathkeys = build_index_pathkeys(root, rel, index, false);
            truncate_useless_pathkeys(root, rel, &index_pathkeys)
        }
        false => Vec::new(),
    };
    if !index_clauses.is_empty() || !useful_pathkeys.is_empty() || useful_predicate || index_only_scan {
        let path = IndexPath {
            index,
            indexclauses: index_clauses.clone(),
            backward: false,
            indexonly: index_only_scan,
            indextotalcost: 0.0,
            indexselectivity: 1.0,
        };
        result.push(create_index_path(root, rel, path, useful_pathkeys, &outer_relids, loop_count));
    }
    if index_is_ordered && pathkeys_possibly_useful {
        let index_pathkeys = build_index_pathkeys(root, rel, index, true);
        let useful_pathkeys = truncate_useless_pathkeys(root, rel, &index_pathkeys);
        if !useful_pathkeys.is_empty() {
            let path = IndexPath {
                index,
                indexclauses: index_clauses,
                backward: true,
                indexonly: index_only_scan,
                indextotalcost: 0.0,
                indexselectivity: 1.0,
            };
            result.push(create_index_path(root, rel, path, useful_pathkeys, &outer_relids, loop_count));
        }
    }
    result
}

/// create_index_path makes the path of an index scan, with its parameterization's rows and its costs, as Postgres'
/// function of the same name in pathnode.c does. When the query only counts the rows of this one table and the scan
/// tests nothing after its index clauses, Doltgres' executor counts them from the index's entry counts, so the scan
/// costs only its descent.
fn create_index_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    mut path: IndexPath,
    pathkeys: Vec<super::nodes::PkId>,
    required_outer: &Relids,
    loop_count: f64,
) -> Rc<Path> {
    let ppi = super::relnode::get_baserel_parampathinfo(root, rel, required_outer);
    let (rows, ppi_clauses) = match &ppi {
        Some(ppi) => (ppi.ppi_rows, ppi.ppi_clauses.clone()),
        None => (root.rels[rel].rows, Vec::new()),
    };
    let ((disabled_nodes, startup_cost, mut total_cost), selectivity, indextotalcost) =
        cost_index(root, rel, &path, &ppi_clauses, rows, loop_count);
    path.indexselectivity = selectivity;
    path.indextotalcost = indextotalcost;
    let table = root.parse.rte(rel).table();
    let counted = root.counting.as_ref().is_some_and(|calls| {
        !calls.is_empty()
            && calls.iter().all(|call| call.counts_rows() || table.is_some_and(|t| call.counts_set_column(t)))
    });
    let index = &root.rels[rel].indexlist[path.index];
    let qpquals = super::costsize::extract_nonindex_conditions(root, &index.indrestrictinfo, &path.indexclauses);
    if counted && qpquals.is_empty() && required_outer.is_empty() && root.all_baserels == Relids::singleton(rel) {
        total_cost = startup_cost;
    }
    let parent = &root.rels[rel];
    Rc::new(Path {
        kind: PathKind::IndexScan(Box::new(path)),
        parent: rel,
        relids: parent.relids.clone(),
        param: required_outer.clone(),
        pathkeys,
        rows,
        width: parent.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
        pathtarget: None,
    })
}

/// check_index_only reports whether an index holds every column of its relation that the query reads, from the
/// relation's target and the restrictions the scan tests, as Postgres' function of the same name decides an
/// index-only scan: Dolt's primary index holds the rows, and a secondary index holds its columns and the primary
/// key's.
fn check_index_only(root: &PlannerInfo<'_, '_>, rel: usize, index: usize) -> bool {
    if !root.enables.indexonlyscan {
        return false;
    }
    let info = &root.rels[rel].indexlist[index];
    if info.index.is_none() {
        return true;
    }
    let mut attrs_used: Vec<usize> = Vec::new();
    let quals = info.indrestrictinfo.iter().map(|&r| &root.rinfos[r].clause);
    for e in root.rels[rel].reltarget.exprs.iter().chain(quals) {
        let vars = super::var::pull_var_clause(root.glob, e, false);
        for id in vars {
            if let VarNode::Var(var) = root.glob.node(id)
                && var.varno == rel
            {
                attrs_used.push(var.varattno);
            }
        }
    }
    attrs_used.iter().all(|a| info.indexkeys.contains(&Some(*a)))
}

/// get_loop_count returns how many times a nested loop runs a parameterized path of a relation, the fewest rows among
/// its outer relations, counting a semi join's outer rows once for each distinct join value, as Postgres' function of
/// the same name estimates it.
fn get_loop_count(root: &PlannerInfo<'_, '_>, cur_relid: usize, outer_relids: &Relids) -> f64 {
    let mut result = 0.0;
    for outer_relid in outer_relids.members() {
        let Some(outer_rel) = root.rels.get(outer_relid).filter(|r| r.reloptkind == RelOptKind::BaseRel) else {
            continue;
        };
        if super::joinrels::is_dummy_rel(root, outer_relid) {
            continue;
        }
        let rowcount = adjust_rowcount_for_semijoins(root, cur_relid, outer_relid, outer_rel.rows);
        if result == 0.0 || result > rowcount {
            result = rowcount;
        }
    }
    if result > 0.0 { result } else { 1.0 }
}

/// adjust_rowcount_for_semijoins returns the rows of an outer relation that a parameterized path is run for, at most
/// the distinct join values of a semi join whose inner side the relation is in and whose outer side is the path's
/// relation's, as Postgres' function of the same name does.
fn adjust_rowcount_for_semijoins(
    root: &PlannerInfo<'_, '_>,
    cur_relid: usize,
    outer_relid: usize,
    rowcount: f64,
) -> f64 {
    let mut rowcount = rowcount;
    for &sj in &root.join_info_list {
        let sjinfo = &root.sjinfos[sj];
        if sjinfo.jointype == JoinType::Semi
            && sjinfo.syn_lefthand.is_member(cur_relid)
            && sjinfo.syn_righthand.is_member(outer_relid)
        {
            let nraw = approximate_joinrel_size(root, &sjinfo.syn_righthand);
            let nunique = super::selfuncs::estimate_num_groups(root, &sjinfo.semi_rhs_exprs, nraw, None, None);
            rowcount = rowcount.min(nunique);
        }
    }
    rowcount
}

/// approximate_joinrel_size returns the product of the rows of a set of base relations, an upper bound on the rows of
/// their join, as Postgres' function of the same name does.
fn approximate_joinrel_size(root: &PlannerInfo<'_, '_>, relids: &Relids) -> f64 {
    relids
        .members()
        .filter(|&relid| root.rels.get(relid).is_some_and(|r| r.relid == relid))
        .filter(|&relid| !super::joinrels::is_dummy_rel(root, relid))
        .map(|relid| root.rels[relid].rows)
        .product()
}

/// build_paths_for_OR returns the bitmap scans of each index that can search by some of a list of clauses, also using
/// other clauses, where a partial index's predicate must follow from them, as Postgres' function of the same name
/// does.
#[allow(non_snake_case)]
fn build_paths_for_OR(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    clauses: &[RinfoId],
    other_clauses: &[RinfoId],
) -> Vec<Rc<Path>> {
    let mut result = Vec::new();
    for index in 0..root.rels[rel].indexlist.len() {
        let info = &root.rels[rel].indexlist[index];
        let mut useful_predicate = false;
        if !info.indpred.is_empty() && !info.pred_ok {
            let all_clauses: Vec<RinfoId> = clauses.iter().chain(other_clauses).copied().collect();
            if !predicate_implied_by(root, &info.indpred, &all_clauses) {
                continue;
            }
            useful_predicate = !predicate_implied_by(root, &info.indpred, other_clauses);
        }
        let mut clauseset = vec![Vec::new(); info.nkeycolumns];
        match_clauses_to_index(root, rel, clauses, index, &mut clauseset);
        if clauseset.iter().all(Vec::is_empty) && !useful_predicate {
            continue;
        }
        match_clauses_to_index(root, rel, other_clauses, index, &mut clauseset);
        result.extend(build_index_paths(root, rel, index, &clauseset, useful_predicate, ScanTypeControl::BitmapScan));
    }
    result
}

/// predicate_implied_by reports whether restrictions imply an index's predicate, as Postgres' function of the same
/// name does with a list of RestrictInfos.
fn predicate_implied_by(root: &PlannerInfo<'_, '_>, predicate: &[Expr], clauses: &[RinfoId]) -> bool {
    let clauses: Vec<Expr> = clauses.iter().map(|&r| root.rinfos[r].clause.clone()).collect();
    predtest::predicate_implied_by(root, predicate, &clauses, false)
}

/// OrArgIndexMatch is the index column that an argument of an OR clause compares, by the index's position, the
/// column's, the comparison's operator, and the other side's type, with the argument's position and its group's, as
/// Postgres' structure of the same name holds them.
#[derive(Clone, Copy)]
struct OrArgIndexMatch {
    indexnum: Option<usize>,
    colnum: usize,
    opno: Option<(CmpOp, Option<u32>)>,
    argindex: usize,
    groupindex: usize,
}

/// OrArgKey is what the arguments of an OR clause that one group gathers share: the index, the column, and the
/// comparison's operator and other side's type.
type OrArgKey = (Option<usize>, usize, Option<(u8, Option<u32>)>);

impl OrArgIndexMatch {
    /// key returns what arguments of one group share.
    fn key(&self) -> OrArgKey {
        (self.indexnum, self.colnum, self.opno.map(|(op, ty)| (op as u8, ty)))
    }
}

/// group_similar_or_args returns the arguments of an OR clause with those that compare the same index column by the
/// same operator gathered into one OR clause of them, which an index can search by as an array, or None when no
/// argument compares an index column, as Postgres' function of the same name does.
fn group_similar_or_args(root: &mut PlannerInfo<'_, '_>, rel: usize, rinfo: RinfoId) -> Option<Vec<Vec<RinfoId>>> {
    let orargs = root.rinfos[rinfo].orclause.clone().expect("an OR clause");
    let n = orargs.len();
    let mut matches: Vec<OrArgIndexMatch> = Vec::with_capacity(n);
    let mut matched = false;
    for (i, arg) in orargs.iter().enumerate() {
        let mut m = OrArgIndexMatch { indexnum: None, colnum: 0, opno: None, argindex: i, groupindex: i };
        if let [argrinfo] = arg.as_slice() {
            let r = &root.rinfos[*argrinfo];
            if let Expr::Compare(op, leftop, rightop) = &r.clause {
                let side = match (r.left_relids.is_member(rel), r.right_relids.is_member(rel)) {
                    (false, true) if !super::clauses::contain_volatile_functions(root.glob, leftop) => {
                        Some((crate::indexscan::swap(*op), &**rightop, &**leftop))
                    }
                    (true, false) if !super::clauses::contain_volatile_functions(root.glob, rightop) => {
                        Some((*op, &**leftop, &**rightop))
                    }
                    _ => None,
                };
                if let Some((op, non_const, other)) = side {
                    'indexes: for (indexnum, index) in root.rels[rel].indexlist.iter().enumerate() {
                        for colnum in 0..index.nkeycolumns {
                            if match_index_to_operand(root, non_const, colnum, index, rel) {
                                let ty = super::nodefuncs::expr_type(root, other);
                                m = OrArgIndexMatch { indexnum: Some(indexnum), colnum, opno: Some((op, ty)), ..m };
                                matched = true;
                                break 'indexes;
                            }
                        }
                    }
                }
            }
        }
        matches.push(m);
    }
    if !matched {
        return None;
    }
    matches.sort_by(|a, b| a.key().cmp(&b.key()).then(a.argindex.cmp(&b.argindex)));
    for i in 1..n {
        if matches[i].indexnum.is_some() && matches[i].key() == matches[i - 1].key() {
            matches[i].groupindex = matches[i - 1].groupindex;
        }
    }
    matches.sort_by_key(|m| (m.groupindex, m.argindex));
    let mut result = Vec::new();
    let mut group_start = 0;
    for i in 1..=n {
        if i < n && matches[i].indexnum.is_some() && matches[i].key() == matches[group_start].key() {
            continue;
        }
        if i - group_start == 1 {
            result.push(orargs[matches[group_start].argindex].clone());
        } else {
            let rargs: Vec<Vec<RinfoId>> = (group_start..i).map(|j| orargs[matches[j].argindex].clone()).collect();
            let clause = rargs
                .iter()
                .map(|arg| root.rinfos[arg[0]].clause.clone())
                .reduce(|a, b| Expr::Or(Box::new(a), Box::new(b)))
                .expect("a group of at least two arguments");
            let r = root.rinfos[rinfo].clone();
            let args = RestrictInfoArgs {
                is_pushed_down: r.is_pushed_down,
                has_clone: r.has_clone,
                is_clone: r.is_clone,
                pseudoconstant: r.pseudoconstant,
                security_level: r.security_level,
                required_relids: Some(r.required_relids),
                incompatible_relids: r.incompatible_relids,
                outer_relids: r.outer_relids,
            };
            result.push(vec![make_plain_restrictinfo(root, clause, Some(rargs), args)]);
        }
        group_start = i;
    }
    Some(result)
}

/// make_bitmap_paths_for_or_group returns the bitmap scans of a group of an OR clause's arguments that compare one
/// index column: one scan of the whole group, or one for each argument, whichever costs less, as Postgres' function of
/// the same name does.
fn make_bitmap_paths_for_or_group(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    ri: RinfoId,
    other_clauses: &[RinfoId],
) -> Vec<Rc<Path>> {
    let mut jointlist = Vec::new();
    let mut jointcost = 0.0;
    let indlist = build_paths_for_OR(root, rel, &[ri], other_clauses);
    if !indlist.is_empty() {
        let bitmapqual = choose_bitmap_and(root, rel, indlist);
        jointcost = bitmapqual.total_cost;
        jointlist.push(bitmapqual);
    }
    if !jointlist.is_empty() && other_clauses.is_empty() {
        return jointlist;
    }
    let mut splitlist = Vec::new();
    let mut splitcost = 0.0;
    for arg in root.rinfos[ri].orclause.clone().expect("an OR clause") {
        let indlist = build_paths_for_OR(root, rel, &arg, other_clauses);
        if indlist.is_empty() {
            splitlist.clear();
            break;
        }
        let bitmapqual = choose_bitmap_and(root, rel, indlist);
        splitcost += bitmapqual.total_cost;
        splitlist.push(bitmapqual);
    }
    match (jointlist.is_empty(), splitlist.is_empty()) {
        (_, true) => jointlist,
        (true, false) => splitlist,
        (false, false) if jointcost < splitcost => jointlist,
        (false, false) => splitlist,
    }
}

/// generate_bitmap_or_paths returns a BitmapOr for each OR clause of a list whose every argument some index can search
/// by, also using other clauses, as Postgres' function of the same name does.
fn generate_bitmap_or_paths(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    clauses: &[RinfoId],
    other_clauses: &[RinfoId],
) -> Vec<Rc<Path>> {
    let mut result = Vec::new();
    let all_clauses: Vec<RinfoId> = clauses.iter().chain(other_clauses).copied().collect();
    for &rinfo in clauses {
        if !restriction_is_or_clause(&root.rinfos[rinfo]) {
            continue;
        }
        let mut pathlist: Vec<Rc<Path>> = Vec::new();
        let grouped = group_similar_or_args(root, rel, rinfo);
        let inner_other_clauses: Vec<RinfoId> = match grouped {
            Some(_) => all_clauses.iter().copied().filter(|&r| r != rinfo).collect(),
            None => Vec::new(),
        };
        let orargs = grouped.unwrap_or_else(|| root.rinfos[rinfo].orclause.clone().expect("an OR clause"));
        for orarg in orargs {
            let indlist = match orarg.as_slice() {
                [ri] if restriction_is_or_clause(&root.rinfos[*ri]) => {
                    let indlist = make_bitmap_paths_for_or_group(root, rel, *ri, &inner_other_clauses);
                    if indlist.is_empty() {
                        pathlist.clear();
                        break;
                    }
                    pathlist.extend(indlist);
                    continue;
                }
                [ri] => build_paths_for_OR(root, rel, &[*ri], &all_clauses),
                andargs => {
                    let mut indlist = build_paths_for_OR(root, rel, andargs, &all_clauses);
                    indlist.extend(generate_bitmap_or_paths(root, rel, andargs, &all_clauses));
                    indlist
                }
            };
            if indlist.is_empty() {
                pathlist.clear();
                break;
            }
            pathlist.push(choose_bitmap_and(root, rel, indlist));
        }
        if !pathlist.is_empty() {
            result.push(create_bitmap_or_path(root, rel, pathlist));
        }
    }
    result
}

/// PathClauseUsage is the clauses and index predicates that a bitmap scan uses, by their positions among those of all
/// the scans compared, as Postgres' structure of the same name holds them.
struct PathClauseUsage {
    path: Rc<Path>,
    quals: Vec<Expr>,
    preds: Vec<Expr>,
    clauseids: Relids,
    unclassifiable: bool,
}

/// choose_bitmap_and returns the cheapest combination of bitmap scans to AND together, among those that each use
/// clauses that the others do not, trying each scan first in order of cost, as Postgres' function of the same name
/// does.
fn choose_bitmap_and(root: &mut PlannerInfo<'_, '_>, rel: usize, paths: Vec<Rc<Path>>) -> Rc<Path> {
    if paths.len() == 1 {
        return paths.into_iter().next().expect("a path");
    }
    let mut clauselist: Vec<Expr> = Vec::new();
    let mut pathinfoarray: Vec<PathClauseUsage> = Vec::new();
    for ipath in paths {
        let pathinfo = classify_index_clause_usage(root, ipath, &mut clauselist);
        if pathinfo.unclassifiable {
            pathinfoarray.push(pathinfo);
            continue;
        }
        let same = pathinfoarray.iter().position(|p| !p.unclassifiable && p.clauseids == pathinfo.clauseids);
        match same {
            Some(i) => {
                let (ncost, _) = cost_bitmap_tree_node(&pathinfo.path);
                let (ocost, _) = cost_bitmap_tree_node(&pathinfoarray[i].path);
                if ncost < ocost {
                    pathinfoarray[i] = pathinfo;
                }
            }
            None => pathinfoarray.push(pathinfo),
        }
    }
    if pathinfoarray.len() == 1 {
        return pathinfoarray.pop().expect("a path").path;
    }
    pathinfoarray.sort_by(path_usage_comparator);
    let mut bestpaths: Vec<Rc<Path>> = Vec::new();
    let mut bestcost = 0.0;
    for i in 0..pathinfoarray.len() {
        let first = &pathinfoarray[i];
        let mut paths = vec![first.path.clone()];
        let mut costsofar = bitmap_scan_cost_est(root, rel, &first.path);
        let mut qualsofar: Vec<Expr> = first.quals.iter().chain(&first.preds).cloned().collect();
        let mut clauseidsofar = first.clauseids.clone();
        for pathinfo in &pathinfoarray[i + 1..] {
            if pathinfo.clauseids.overlap(&clauseidsofar) {
                continue;
            }
            if pathinfo
                .preds
                .iter()
                .any(|np| predtest::predicate_implied_by(root, std::slice::from_ref(np), &qualsofar, false))
            {
                continue;
            }
            paths.push(pathinfo.path.clone());
            let newcost = bitmap_and_cost_est(root, rel, paths.clone());
            if newcost < costsofar {
                costsofar = newcost;
                qualsofar.extend(pathinfo.quals.iter().chain(&pathinfo.preds).cloned());
                clauseidsofar.add_members(&pathinfo.clauseids);
            } else {
                paths.pop();
            }
        }
        if i == 0 || costsofar < bestcost {
            bestpaths = paths;
            bestcost = costsofar;
        }
    }
    match bestpaths.len() {
        1 => bestpaths.pop().expect("a path"),
        _ => create_bitmap_and_path(root, rel, bestpaths),
    }
}

/// path_usage_comparator orders bitmap scans by their cost, then by the share of the rows they find, as Postgres'
/// function of the same name does.
fn path_usage_comparator(a: &PathClauseUsage, b: &PathClauseUsage) -> std::cmp::Ordering {
    let (acost, aselec) = cost_bitmap_tree_node(&a.path);
    let (bcost, bselec) = cost_bitmap_tree_node(&b.path);
    acost.total_cmp(&bcost).then(aselec.total_cmp(&bselec))
}

/// bitmap_scan_cost_est returns the total cost of a bitmap heap scan of a tree of bitmap scans, as Postgres' function
/// of the same name estimates it.
fn bitmap_scan_cost_est(root: &mut PlannerInfo<'_, '_>, rel: usize, ipath: &Path) -> f64 {
    let ppi = super::relnode::get_baserel_parampathinfo(root, rel, &ipath.param);
    let loop_count = get_loop_count(root, rel, &ipath.param);
    let ((_, _, total_cost), _) = cost_bitmap_heap_scan(root, rel, ppi.as_ref(), ipath, loop_count);
    total_cost
}

/// bitmap_and_cost_est returns the total cost of a bitmap heap scan of a BitmapAnd of bitmap scans, as Postgres'
/// function of the same name estimates it.
fn bitmap_and_cost_est(root: &mut PlannerInfo<'_, '_>, rel: usize, paths: Vec<Rc<Path>>) -> f64 {
    let apath = create_bitmap_and_path(root, rel, paths);
    bitmap_scan_cost_est(root, rel, &apath)
}

/// classify_index_clause_usage returns the clauses and index predicates that a tree of bitmap scans uses, by their
/// positions in a list of all of them, which it adds the new ones to, as Postgres' function of the same name does.
fn classify_index_clause_usage(
    root: &PlannerInfo<'_, '_>,
    path: Rc<Path>,
    clauselist: &mut Vec<Expr>,
) -> PathClauseUsage {
    let (mut quals, mut preds) = (Vec::new(), Vec::new());
    find_indexpath_quals(root, &path, &mut quals, &mut preds);
    if quals.len() + preds.len() > 100 {
        return PathClauseUsage { path, quals, preds, clauseids: Relids::new(), unclassifiable: true };
    }
    let mut clauseids = Relids::new();
    for node in quals.iter().chain(&preds) {
        clauseids.add_member(find_list_position(node, clauselist));
    }
    PathClauseUsage { path, quals, preds, clauseids, unclassifiable: false }
}

/// find_indexpath_quals adds the clauses that the index scans of a tree of bitmap scans search by, and the predicates
/// of their indexes, as Postgres' function of the same name does.
pub fn find_indexpath_quals(
    root: &PlannerInfo<'_, '_>,
    bitmapqual: &Path,
    quals: &mut Vec<Expr>,
    preds: &mut Vec<Expr>,
) {
    match &bitmapqual.kind {
        PathKind::BitmapAnd(bpath) | PathKind::BitmapOr(bpath) => {
            for subpath in &bpath.bitmapquals {
                find_indexpath_quals(root, subpath, quals, preds);
            }
        }
        PathKind::IndexScan(ipath) => {
            quals.extend(ipath.indexclauses.iter().map(|iclause| root.rinfos[iclause.rinfo].clause.clone()));
            preds.extend(root.rels[bitmapqual.parent].indexlist[ipath.index].indpred.iter().cloned());
        }
        _ => unreachable!("a bitmap tree holds index scans, BitmapAnds, and BitmapOrs"),
    }
}

/// find_list_position returns the position of an expression in a list, adding it at the end when it is not there, as
/// Postgres' function of the same name does.
fn find_list_position(node: &Expr, nodelist: &mut Vec<Expr>) -> usize {
    match nodelist.iter().position(|old| old == node) {
        Some(i) => i,
        None => {
            nodelist.push(node.clone());
            nodelist.len() - 1
        }
    }
}

/// match_restriction_clauses_to_index returns the restrictions that each column of an index can search by, as
/// Postgres' function of the same name does.
fn match_restriction_clauses_to_index(root: &mut PlannerInfo<'_, '_>, rel: usize, index: usize) -> IndexClauseSet {
    let mut clauseset = vec![Vec::new(); root.rels[rel].indexlist[index].nkeycolumns];
    let clauses = root.rels[rel].indexlist[index].indrestrictinfo.clone();
    match_clauses_to_index(root, rel, &clauses, index, &mut clauseset);
    clauseset
}

/// match_join_clauses_to_index returns the join clauses movable to a relation that each column of an index can search
/// by, collecting the OR clauses among them for bitmap scans, as Postgres' function of the same name does.
fn match_join_clauses_to_index(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    index: usize,
    joinorclauses: &mut Vec<RinfoId>,
) -> IndexClauseSet {
    let mut clauseset = vec![Vec::new(); root.rels[rel].indexlist[index].nkeycolumns];
    let clauses: Vec<RinfoId> = root.rels[rel]
        .joininfo
        .iter()
        .copied()
        .filter(|&r| join_clause_is_movable_to(&root.rinfos[r], &root.rels[rel]))
        .collect();
    for &rinfo in &clauses {
        if restriction_is_or_clause(&root.rinfos[rinfo]) && !joinorclauses.contains(&rinfo) {
            joinorclauses.push(rinfo);
        }
    }
    match_clauses_to_index(root, rel, &clauses, index, &mut clauseset);
    clauseset
}

/// match_eclass_clauses_to_index returns the join clauses that the equivalence classes imply between each key column
/// of an index and the members of other relations, as Postgres' function of the same name finds them.
fn match_eclass_clauses_to_index(root: &mut PlannerInfo<'_, '_>, rel: usize, index: usize) -> IndexClauseSet {
    let nkeycolumns = root.rels[rel].indexlist[index].nkeycolumns;
    let mut clauseset = vec![Vec::new(); nkeycolumns];
    if !root.rels[rel].has_eclass_joins {
        return clauseset;
    }
    let prohibited = root.rels[rel].lateral_referencers.clone();
    let info = root.rels[rel].indexlist[index].clone();
    for indexcol in 0..nkeycolumns {
        let callback = |root: &PlannerInfo<'_, '_>, ec: EcId, em: EmId| {
            ec_member_matches_indexcol(root, rel, &info, indexcol, ec, em)
        };
        let clauses = generate_implied_equalities_for_column(root, rel, &callback, &prohibited);
        match_clauses_to_index(root, rel, &clauses, index, &mut clauseset);
    }
    clauseset
}

/// match_clauses_to_index adds each clause of a list that a column of an index can search by to the column's index
/// clauses, as Postgres' function of the same name does.
fn match_clauses_to_index(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    clauses: &[RinfoId],
    index: usize,
    clauseset: &mut IndexClauseSet,
) {
    for &rinfo in clauses {
        match_clause_to_index(root, rel, rinfo, index, clauseset);
    }
}

/// match_clause_to_index adds a clause to the index clauses of the first column of an index that can search by it,
/// unless it is pseudoconstant or already there, as Postgres' function of the same name does.
fn match_clause_to_index(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    rinfo: RinfoId,
    index: usize,
    clauseset: &mut IndexClauseSet,
) {
    if root.rinfos[rinfo].pseudoconstant {
        return;
    }
    for (indexcol, indexclauses) in clauseset.iter_mut().enumerate() {
        if indexclauses.iter().any(|iclause| iclause.rinfo == rinfo) {
            return;
        }
        if let Some(iclause) = match_clause_to_indexcol(root, rel, rinfo, indexcol, index) {
            indexclauses.push(iclause);
            return;
        }
    }
}

/// match_clause_to_indexcol returns the index clause that a clause makes for a column of an index, when the column
/// can search by it, as Postgres' function of the same name does: a boolean column itself, a comparison of the column
/// with something that reads no Var of its relation, a LIKE of the column with a fixed prefix, an IN list or ANY
/// array, an OR of equalities of the column, or a NULL test of the column.
fn match_clause_to_indexcol(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    rinfo: RinfoId,
    indexcol: usize,
    index: usize,
) -> Option<IndexClause> {
    let info = root.rels[rel].indexlist[index].clone();
    if info.opfamily[indexcol] == super::nodefuncs::btree_opfamily(super::nodefuncs::BOOLOID)
        && let Some(iclause) = match_boolean_index_clause(root, rel, rinfo, indexcol, &info)
    {
        return Some(iclause);
    }
    let plain = IndexClause { rinfo, indexquals: vec![rinfo], lossy: false, indexcol };
    let clause = &root.rinfos[rinfo].clause;
    match clause {
        Expr::Compare(..) => match_opclause_to_indexcol(root, rel, rinfo, indexcol, &info),
        Expr::Func(..) => match_funcclause_to_indexcol(root, rel, rinfo, indexcol, &info),
        Expr::AnyArray(..) => match_saopclause_to_indexcol(root, rel, rinfo, indexcol, &info),
        Expr::RowCompare(..) => match_rowcompare_to_indexcol(root, rel, rinfo, indexcol, &info),
        _ if root.rinfos[rinfo].orclause.is_some() => match_orclause_to_indexcol(root, rel, rinfo, indexcol, &info),
        Expr::IsNull(arg, _) if !matches!(**arg, Expr::Row(..)) => {
            match_index_to_operand(root, arg, indexcol, &info, rel).then_some(plain)
        }
        Expr::Not(inner) => match &**inner {
            Expr::IsNull(arg, false) => match_index_to_operand(root, arg, indexcol, &info, rel).then_some(plain),
            _ => None,
        },
        _ => None,
    }
}

/// indexcol_is_bool_constant_for_query reports whether a restriction fixes a boolean index column to a constant, so
/// that the column does not change the order of the index's rows that the scan reads, as Postgres' function of the
/// same name does.
pub fn indexcol_is_bool_constant_for_query(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    index: usize,
    indexcol: usize,
) -> bool {
    let info = root.rels[rel].indexlist[index].clone();
    if info.opfamily[indexcol] != super::nodefuncs::btree_opfamily(super::nodefuncs::BOOLOID) {
        return false;
    }
    for rinfo in root.rels[rel].baserestrictinfo.clone() {
        if root.rinfos[rinfo].pseudoconstant {
            continue;
        }
        if match_boolean_index_clause(root, rel, rinfo, indexcol, &info).is_some() {
            return true;
        }
    }
    false
}

/// make_simple_restrictinfo returns the RestrictInfo of a clause that an index clause derives, as Postgres'
/// make_simple_restrictinfo does.
fn make_simple_restrictinfo(root: &mut PlannerInfo<'_, '_>, clause: Expr) -> RinfoId {
    make_restrictinfo(root, clause, RestrictInfoArgs { is_pushed_down: true, ..RestrictInfoArgs::default() })
}

/// match_boolean_index_clause returns the equality with true or false that a boolean column itself, its NOT, or its
/// IS TRUE or IS FALSE test makes an index clause of, as Postgres' function of the same name does.
fn match_boolean_index_clause(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    rinfo: RinfoId,
    indexcol: usize,
    index: &IndexOptInfo,
) -> Option<IndexClause> {
    let clause = root.rinfos[rinfo].clause.clone();
    let eq = |arg: &Expr, value: bool| {
        Expr::Compare(CmpOp::Eq, Box::new(arg.clone()), Box::new(Expr::Const(crate::types::Value::Bool(value))))
    };
    let op = if match_index_to_operand(root, &clause, indexcol, index, rel) {
        eq(&clause, true)
    } else {
        match &clause {
            Expr::Not(arg) if match_index_to_operand(root, arg, indexcol, index, rel) => eq(arg, false),
            Expr::BoolTest(arg, Some(value), false) if match_index_to_operand(root, arg, indexcol, index, rel) => {
                eq(arg, *value)
            }
            _ => return None,
        }
    };
    let indexqual = make_simple_restrictinfo(root, op);
    Some(IndexClause { rinfo, indexquals: vec![indexqual], lossy: false, indexcol })
}

/// in_opfamily reports whether a comparison is a btree operator of an index column's operator family: an ordering or
/// equality whose other side's type is in the family.
fn in_opfamily(root: &PlannerInfo<'_, '_>, op: CmpOp, other: &Expr, opfamily: Option<u32>) -> bool {
    let other_family = super::nodefuncs::expr_type(root, other).and_then(super::nodefuncs::btree_opfamily);
    matches!(op, CmpOp::Lt | CmpOp::Le | CmpOp::Eq | CmpOp::Ge | CmpOp::Gt)
        && opfamily.is_some()
        && other_family == opfamily
}

/// match_opclause_to_indexcol returns the index clause that a comparison of an index column with an expression that
/// reads no Var of the column's relation and runs no volatile function makes, with its sides swapped when the column
/// is on the right, as Postgres' function of the same name does.
fn match_opclause_to_indexcol(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    rinfo: RinfoId,
    indexcol: usize,
    index: &IndexOptInfo,
) -> Option<IndexClause> {
    let r = &root.rinfos[rinfo];
    let Expr::Compare(op, leftop, rightop) = &r.clause else { return None };
    if matches!(**leftop, Expr::Row(..)) {
        return None;
    }
    let opfamily = index.opfamily[indexcol];
    if match_index_to_operand(root, leftop, indexcol, index, rel)
        && !r.right_relids.is_member(rel)
        && !super::clauses::contain_volatile_functions(root.glob, rightop)
        && in_opfamily(root, *op, rightop, opfamily)
    {
        return Some(IndexClause { rinfo, indexquals: vec![rinfo], lossy: false, indexcol });
    }
    let commuted = match_index_to_operand(root, rightop, indexcol, index, rel)
        && !r.left_relids.is_member(rel)
        && !super::clauses::contain_volatile_functions(root.glob, leftop)
        && in_opfamily(root, crate::indexscan::swap(*op), leftop, opfamily);
    if commuted {
        let commrinfo = super::restrictinfo::commute_restrictinfo(root, rinfo);
        return Some(IndexClause { rinfo, indexquals: vec![commrinfo], lossy: false, indexcol });
    }
    None
}

/// match_funcclause_to_indexcol returns the index clauses that a LIKE of an index column with a pattern of a fixed
/// prefix makes, its prefix's bounds, as Postgres' like_support function makes them for get_index_clause_from_support.
fn match_funcclause_to_indexcol(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    rinfo: RinfoId,
    indexcol: usize,
    index: &IndexOptInfo,
) -> Option<IndexClause> {
    let Expr::Func(f, args) = root.rinfos[rinfo].clause.clone() else { return None };
    let [column, Expr::Const(crate::types::Value::Text(pattern))] = args.as_slice() else { return None };
    if crate::functions::function(f).name != "textlike"
        || !match_index_to_operand(root, column, indexcol, index, rel)
        || index.opfamily[indexcol] != super::nodefuncs::btree_opfamily(crate::oid::TEXT)
    {
        return None;
    }
    let (lower, upper) = crate::indexscan::like_prefix_bounds(pattern)?;
    let bound =
        |op, text| Expr::Compare(op, Box::new(column.clone()), Box::new(Expr::Const(crate::types::Value::Text(text))));
    let mut indexquals = vec![make_simple_restrictinfo(root, bound(CmpOp::Ge, lower))];
    if let Some(upper) = upper {
        indexquals.push(make_simple_restrictinfo(root, bound(CmpOp::Lt, upper)));
    }
    Some(IndexClause { rinfo, indexquals, lossy: true, indexcol })
}

/// match_saopclause_to_indexcol returns the index clause that an IN list or ANY array of an index column makes, as
/// Postgres' function of the same name does.
fn match_saopclause_to_indexcol(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    rinfo: RinfoId,
    indexcol: usize,
    index: &IndexOptInfo,
) -> Option<IndexClause> {
    let Expr::AnyArray(comparison, array, false) = &root.rinfos[rinfo].clause else { return None };
    let Expr::Compare(CmpOp::Eq, leftop, value) = &**comparison else { return None };
    if !matches!(**value, Expr::SubqueryValue)
        || !match_index_to_operand(root, leftop, indexcol, index, rel)
        || pull_varnos(root, array).is_member(rel)
        || super::clauses::contain_volatile_functions(root.glob, array)
    {
        return None;
    }
    Some(IndexClause { rinfo, indexquals: vec![rinfo], lossy: false, indexcol })
}

/// match_rowcompare_to_indexcol returns the index clause that a row comparison makes when the first field of one side
/// is an index column and the first field of the other reads no Var of its relation, as Postgres' function of the
/// same name does.
fn match_rowcompare_to_indexcol(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    rinfo: RinfoId,
    indexcol: usize,
    index: &IndexOptInfo,
) -> Option<IndexClause> {
    let Expr::RowCompare(op, largs, rargs) = &root.rinfos[rinfo].clause else { return None };
    let (op, leftop, rightop) = (*op, &largs[0], &rargs[0]);
    let usable = |root: &PlannerInfo<'_, '_>, other: &Expr| {
        !pull_varnos(root, other).is_member(rel) && !super::clauses::contain_volatile_functions(root.glob, other)
    };
    let (expr_op, other, var_on_left) =
        if match_index_to_operand(root, leftop, indexcol, index, rel) && usable(root, rightop) {
            (op, rightop, true)
        } else if match_index_to_operand(root, rightop, indexcol, index, rel) && usable(root, leftop) {
            (crate::indexscan::swap(op), leftop, false)
        } else {
            return None;
        };
    if !in_opfamily(root, expr_op, other, index.opfamily[indexcol]) {
        return None;
    }
    expand_indexqual_rowcompare(root, rel, rinfo, indexcol, index, expr_op, var_on_left)
}

/// expand_indexqual_rowcompare returns the index clause of a row comparison whose first fields an index column
/// matches: the comparison itself when its index columns are on the left and every field pair matches one, and
/// otherwise the comparison of the leading pairs that match, with the index columns on the left, where an ordering
/// that is strict becomes one that is not when pairs after them are left out, as Postgres' function of the same name
/// does.
fn expand_indexqual_rowcompare(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    rinfo: RinfoId,
    indexcol: usize,
    index: &IndexOptInfo,
    expr_op: CmpOp,
    var_on_left: bool,
) -> Option<IndexClause> {
    let Expr::RowCompare(_, largs, rargs) = root.rinfos[rinfo].clause.clone() else { return None };
    let (var_args, non_var_args) = if var_on_left { (largs, rargs) } else { (rargs, largs) };
    let mut matching_cols = 1;
    while matching_cols < var_args.len() {
        let (varop, constop) = (&var_args[matching_cols], &non_var_args[matching_cols]);
        if pull_varnos(root, constop).is_member(rel) || super::clauses::contain_volatile_functions(root.glob, constop) {
            break;
        }
        let matched = (0..index.nkeycolumns).any(|i| {
            match_index_to_operand(root, varop, i, index, rel) && in_opfamily(root, expr_op, constop, index.opfamily[i])
        });
        if !matched {
            break;
        }
        matching_cols += 1;
    }
    let lossy = matching_cols != var_args.len();
    if var_on_left && !lossy {
        return Some(IndexClause { rinfo, indexquals: vec![rinfo], lossy, indexcol });
    }
    let new_op = match (lossy, expr_op) {
        (true, CmpOp::Lt) => CmpOp::Le,
        (true, CmpOp::Gt) => CmpOp::Ge,
        (_, op) => op,
    };
    let clause = match matching_cols {
        1 => Expr::Compare(new_op, Box::new(var_args[0].clone()), Box::new(non_var_args[0].clone())),
        _ => Expr::RowCompare(new_op, var_args[..matching_cols].to_vec(), non_var_args[..matching_cols].to_vec()),
    };
    let indexqual = make_simple_restrictinfo(root, clause);
    Some(IndexClause { rinfo, indexquals: vec![indexqual], lossy, indexcol })
}

/// match_orclause_to_indexcol returns the index clause that an OR of equalities of an index column with expressions
/// that read no Var of its relation makes, as Postgres' function of the same name turns it into an ANY array.
fn match_orclause_to_indexcol(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    rinfo: RinfoId,
    indexcol: usize,
    index: &IndexOptInfo,
) -> Option<IndexClause> {
    let arms = root.rinfos[rinfo].orclause.clone()?;
    for arm in &arms {
        let [sub] = arm.as_slice() else { return None };
        let Expr::Compare(CmpOp::Eq, l, r) = &root.rinfos[*sub].clause else { return None };
        let (column, other) = match match_index_to_operand(root, l, indexcol, index, rel) {
            true => (l, r),
            false => (r, l),
        };
        if !match_index_to_operand(root, column, indexcol, index, rel)
            || pull_varnos(root, other).is_member(rel)
            || super::clauses::contain_volatile_functions(root.glob, other)
            || !in_opfamily(root, CmpOp::Eq, other, index.opfamily[indexcol])
        {
            return None;
        }
    }
    Some(IndexClause { rinfo, indexquals: vec![rinfo], lossy: false, indexcol })
}

/// ec_member_matches_indexcol reports whether an equivalence class's member is an index's key column, in the
/// column's operator family, as Postgres' function of the same name does.
fn ec_member_matches_indexcol(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    index: &IndexOptInfo,
    indexcol: usize,
    ec: EcId,
    em: EmId,
) -> bool {
    index.opfamily[indexcol].is_some_and(|f| root.eq_classes[ec].ec_opfamilies.contains(&f))
        && match_index_to_operand(root, &root.eq_members[em].em_expr, indexcol, index, rel)
}

/// check_index_predicates records which partial indexes the query's restrictions, join clauses movable to the
/// relation, and equalities that classes imply with other relations prove the predicates of, and the restrictions
/// that each must still test, those its predicate does not imply, as Postgres' function of the same name does.
pub fn check_index_predicates(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    let restrictinfo = root.rels[rel].baserestrictinfo.clone();
    for index in root.rels[rel].indexlist.iter_mut().map(Rc::make_mut) {
        index.indrestrictinfo = restrictinfo.clone();
    }
    if root.rels[rel].indexlist.iter().all(|index| index.indpred.is_empty()) {
        return;
    }
    let mut clauselist: Vec<RinfoId> = restrictinfo.clone();
    for &r in &root.rels[rel].joininfo {
        if join_clause_is_movable_to(&root.rinfos[r], &root.rels[rel]) {
            clauselist.push(r);
        }
    }
    let otherrels = root.all_query_rels.difference(&root.rels[rel].relids);
    if !otherrels.is_empty() {
        let joinrelids = root.rels[rel].relids.union(&otherrels);
        clauselist.extend(super::equivclass::generate_join_implied_equalities(root, &joinrelids, &otherrels, rel, 0));
    }
    let clauses: Vec<Expr> = clauselist.iter().map(|&r| root.rinfos[r].clause.clone()).collect();
    for index in 0..root.rels[rel].indexlist.len() {
        let info = root.rels[rel].indexlist[index].clone();
        if info.indpred.is_empty() {
            continue;
        }
        let pred_ok = info.pred_ok || predtest::predicate_implied_by(root, &info.indpred, &clauses, false);
        let indrestrictinfo: Vec<RinfoId> = restrictinfo
            .iter()
            .copied()
            .filter(|&r| {
                let clause = &root.rinfos[r].clause;
                super::clauses::contain_mutable_functions(root.glob, clause)
                    || !predtest::predicate_implied_by(root, std::slice::from_ref(clause), &info.indpred, false)
            })
            .collect();
        let info = Rc::make_mut(&mut root.rels[rel].indexlist[index]);
        info.pred_ok = pred_ok;
        info.indrestrictinfo = indrestrictinfo;
    }
}

/// lookup_keys returns the expressions that a parameterized scan of an index looks its rows up by, one for each of
/// the index's leading columns that an equality index clause gives, when the first one does and Doltgres' lookups can
/// search the index by them: ascending columns of types that a lookup's keys compare.
pub fn lookup_keys(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    index: usize,
    indexclauses: &[IndexClause],
) -> Option<Vec<Expr>> {
    let info = &root.rels[rel].indexlist[index];
    let table = root.parse.rte(rel).table()?;
    let mut keys = Vec::new();
    for indexcol in 0..info.nkeycolumns {
        let column = info.indexkeys[indexcol].and_then(|c| table.index_column(c));
        if info.reverse_sort[indexcol]
            || column.is_none_or(|c| crate::storage::is_adaptive(c.encoding) || !crate::exec::lookup_type(c.ty.oid))
        {
            break;
        }
        let key = indexclauses.iter().filter(|iclause| iclause.indexcol == indexcol).find_map(|iclause| {
            let [qual] = iclause.indexquals.as_slice() else { return None };
            match &root.rinfos[*qual].clause {
                Expr::Compare(CmpOp::Eq, _, other) => Some((**other).clone()),
                _ => None,
            }
        });
        match key {
            Some(key) => keys.push(key),
            None => break,
        }
    }
    (!keys.is_empty()).then_some(keys)
}

/// relation_has_unique_index_ext reports whether a unique index of a base relation has each of its key columns
/// equated with something by one of a list of clauses, whose outer sides are given, or by one of the relation's
/// restrictions, returning the restrictions that it used when asked, as Postgres' function of the same name does.
/// A key column may instead be one of a list of expressions that the caller knows are equated with something by
/// their type's equality.
pub fn relation_has_unique_index_ext(
    root: &PlannerInfo<'_, '_>,
    rel: usize,
    restrictlist: &[RinfoId],
    exprlist: &[Expr],
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
    if restrictlist.is_empty() && exprlist.is_empty() {
        return false;
    }
    for ind in &root.rels[rel].indexlist {
        if !ind.unique || !ind.indpred.is_empty() {
            continue;
        }
        let mut exprs = Vec::new();
        let all_matched = (0..ind.nkeycolumns).all(|c| {
            restrictlist.iter().any(|&rinfo| {
                let r = &root.rinfos[rinfo];
                if !ind.opfamily[c].is_some_and(|f| r.mergeopfamilies.contains(&f)) {
                    return false;
                }
                let Some((left, right)) = binary_op_args(&r.clause) else { return false };
                let rexpr = if r.outer_is_left.get() { right } else { left };
                if !match_index_to_operand(root, rexpr, c, ind, rel) {
                    return false;
                }
                if r.clause_relids.num_members() == 1 {
                    exprs.push(rinfo);
                }
                true
            }) || exprlist.iter().any(|expr| {
                match_index_to_operand(root, expr, c, ind, rel)
                    && super::nodefuncs::expr_type(root, expr)
                        .and_then(super::nodefuncs::btree_opfamily)
                        .is_some_and(|f| Some(f) == ind.opfamily[c])
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

/// match_index_to_operand reports whether an expression is an index's column: a Var of the index's relation that no
/// outer join makes NULL, or the expression of an expression column, looking through a PlaceHolderVar that nothing
/// makes NULL and a cast that converts nothing, as Postgres' function of the same name does, or that stays in a btree
/// operator family, where Postgres compares the types with a cross-type operator instead.
pub fn match_index_to_operand(
    root: &PlannerInfo<'_, '_>,
    operand: &Expr,
    indexcol: usize,
    index: &IndexOptInfo,
    rel: usize,
) -> bool {
    let operand = &super::placeholder::strip_noop_phvs(root.glob, operand.clone());
    let operand = match operand {
        Expr::Cast(arg, ty, _)
            if ty.modifier < 0
                && super::nodefuncs::expr_type(root, arg).is_some_and(|from| {
                    let family = super::nodefuncs::btree_opfamily(from);
                    crate::pgcatalog::binary_coercible(from, ty.oid)
                        || (family.is_some() && family == super::nodefuncs::btree_opfamily(ty.oid))
                }) =>
        {
            arg
        }
        other => other,
    };
    match index.indexkeys[indexcol] {
        Some(attno) => matches!(operand, Expr::Column(id) if matches!(root.glob.node(*id),
            VarNode::Var(var) if var.varno == rel && var.varattno == attno && var.varnullingrels.is_empty())),
        None => {
            let position = index.indexkeys[..indexcol].iter().filter(|k| k.is_none()).count();
            index.indexprs.get(position).is_some_and(|indexkey| indexkey == operand)
        }
    }
}

/// index_column_exprs returns the expression of each of an index's columns, as Postgres' indextlist holds them.
pub fn index_column_exprs(root: &mut PlannerInfo<'_, '_>, rel: usize, index: usize) -> Vec<Expr> {
    let info = root.rels[rel].indexlist[index].clone();
    let mut exprs = info.indexprs.iter();
    info.indexkeys
        .iter()
        .map(|key| match key {
            Some(attno) => root.glob.var(rel, *attno, Relids::new()),
            None => exprs.next().cloned().expect("an expression column has its expression"),
        })
        .collect()
}

/// create_catalog_lookup_paths adds the path of a lookup of a system catalog's rows for each set of other relations
/// whose join equalities find them through one of its indexes, which Doltgres' lookup joins find.
fn create_catalog_lookup_paths(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    let mut outer_sets: Vec<Relids> = Vec::new();
    for &rinfo in &root.rels[rel].joininfo {
        let Some((_, outer)) = join_equality(root, rinfo, rel) else { continue };
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
        create_catalog_lookup_path(root, rel, outer_relids);
    }
}

/// join_equality returns the sides of a join clause that equates an expression of a relation with one of others.
fn join_equality<'r>(root: &'r PlannerInfo<'_, '_>, rinfo: RinfoId, rel: usize) -> Option<(&'r Expr, &'r Expr)> {
    let rinfo = &root.rinfos[rinfo];
    let Expr::Compare(CmpOp::Eq, l, r) = &rinfo.clause else { return None };
    let singleton = Relids::singleton(rel);
    match (rinfo.can_join, rinfo.left_relids == singleton, rinfo.right_relids == singleton) {
        (true, true, false) => Some((l, r)),
        (true, false, true) => Some((r, l)),
        _ => None,
    }
}

/// create_catalog_lookup_path adds the path of a lookup of a system catalog's rows by its join equalities with a set
/// of other relations, when an index of the catalog lets it look them up. A relation that evaluates PlaceHolderVars
/// has none, as Doltgres' lookup joins return the looked-up rows as they are stored.
fn create_catalog_lookup_path(root: &mut PlannerInfo<'_, '_>, rel: usize, outer_relids: Relids) {
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
    for &rinfo in &root.rels[rel].joininfo {
        let Some((inner, outer)) = join_equality(root, rinfo, rel) else { continue };
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
    let RteKind::Plan(plan) = &root.parse.rte(rel).kind else { return };
    let plan = plan.clone();
    let Some(found) = crate::joins::lookup(root.ctx, &plan, &condition, left_width) else { return };
    let method = match found.method {
        JoinMethod::CatalogLookup { index, keys } => JoinMethod::CatalogLookup {
            index,
            keys: keys.iter().map(|k| from_positional(k.clone(), &outer_vars)).collect(),
        },
        _ => return,
    };
    let loop_count = get_loop_count(root, rel, &outer_relids);
    let index = IndexCost {
        pages: 1.0,
        tuples: root.rels[rel].tuples,
        tree_height: 0.0,
        indexonly: true,
        correlation: 1.0,
        nquals: 1,
    };
    let parent = &root.rels[rel];
    let selectivity = if parent.tuples > 0.0 { parent.rows / parent.tuples } else { 1.0 };
    let rows = clamp_row_est(found.matches * selectivity);
    let qpqual_cost: QualCost = cost_qual_eval(root, &parent.baserestrictinfo);
    let (disabled_nodes, startup_cost, total_cost) =
        cost_catalog_lookup(root, rel, &index, found.matches, qpqual_cost, loop_count);
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
        pathtarget: None,
    };
    add_path(&mut root.rels[rel], Rc::new(path));
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
