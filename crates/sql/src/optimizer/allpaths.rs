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

//! Postgres' optimizer/path/allpaths.c: finding the paths of the base relations, and of the relation that joins
//! them all.

use std::rc::Rc;

use super::PlannerInfo;
use super::costsize::{Costs, cost_opaque_scan, cost_resultscan, cost_seqscan, set_baserel_size_estimates};
use super::indxpath::{check_index_predicates, create_index_paths};
use super::initsplan::JoinList;
use super::joinrels::{is_dummy_rel, join_search_one_level};
use super::nodes::{JoinType, Path, PathKind, RelOptKind, RteKind, VarNode};
use super::pathnode::{add_path, set_cheapest};
use crate::expr::Expr;
use crate::plan::Plan;

/// make_one_rel finds the paths of every base relation and then of the join of them all, returning that relation,
/// as Postgres' function of the same name does.
pub fn make_one_rel(root: &mut PlannerInfo<'_, '_>, joinlist: &[JoinList]) -> usize {
    set_base_rel_consider_startup(root);
    set_base_rel_sizes(root);
    root.total_table_pages = (1..=root.parse.rtable.len())
        .filter(|&rti| root.rels[rti].reloptkind == RelOptKind::BaseRel && !is_dummy_rel(root, rti))
        .map(|rti| root.rels[rti].pages)
        .sum();
    set_base_rel_pathlists(root);
    make_rel_from_joinlist(root, joinlist).expect("the join tree has a relation")
}

/// set_base_rel_consider_startup marks the one inner relation of each semi or anti join as worth paths that start
/// cheaply when parameterized, as Postgres' function of the same name does.
fn set_base_rel_consider_startup(root: &mut PlannerInfo<'_, '_>) {
    for sj in root.join_info_list.clone() {
        let sjinfo = &root.sjinfos[sj];
        if matches!(sjinfo.jointype, JoinType::Semi | JoinType::Anti)
            && let Some(varno) = sjinfo.syn_righthand.singleton_member()
        {
            root.rels[varno].consider_param_startup = true;
        }
    }
}

/// set_base_rel_sizes estimates the size of each base relation, as Postgres' function of the same name does.
fn set_base_rel_sizes(root: &mut PlannerInfo<'_, '_>) {
    for rti in 1..=root.parse.rtable.len() {
        if root.rels[rti].reloptkind == RelOptKind::BaseRel {
            set_rel_size(root, rti);
        }
    }
}

/// set_rel_size estimates the size of a base relation, or marks it empty when its restrictions refute it, as
/// Postgres' function of the same name does.
fn set_rel_size(root: &mut PlannerInfo<'_, '_>, rti: usize) {
    if super::plancat::relation_excluded_by_constraints(root, rti) {
        set_dummy_rel_pathlist(root, rti);
        return;
    }
    if matches!(root.parse.rte(rti).kind, RteKind::Subquery(..)) {
        set_subquery_pathlist(root, rti);
        return;
    }
    if matches!(root.parse.rte(rti).kind, RteKind::Relation(..)) {
        check_index_predicates(root, rti);
    }
    set_baserel_size_estimates(root, rti);
}

/// set_dummy_rel_pathlist marks a base relation as returning no rows, as Postgres' function of the same name does.
fn set_dummy_rel_pathlist(root: &mut PlannerInfo<'_, '_>, rti: usize) {
    root.rels[rti].reltarget.width = 0.0;
    super::joinrels::mark_dummy_rel(root, rti);
}

/// set_subquery_pathlist plans a subquery relation after pushing its restrictions that it can test itself into it,
/// and adds a SubqueryScan path over each of its final paths, as Postgres' function of the same name does.
fn set_subquery_pathlist(root: &mut PlannerInfo<'_, '_>, rti: usize) {
    let RteKind::Subquery(subquery, _) = &root.parse.rte(rti).kind else { unreachable!("a subquery relation") };
    let mut subquery = (**subquery).clone();
    let mut safety_info =
        PushdownSafetyInfo { unsafe_flags: vec![0; subquery.target_list.len()], unsafe_volatile: false };
    if !root.rels[rti].baserestrictinfo.is_empty()
        && subquery_is_pushdown_safe(root, &subquery, &subquery, &mut safety_info)
    {
        let mut upperrestrictlist = Vec::new();
        for rinfo in root.rels[rti].baserestrictinfo.clone() {
            if root.rinfos[rinfo].pseudoconstant {
                upperrestrictlist.push(rinfo);
                continue;
            }
            match qual_is_pushdown_safe(root, rti, rinfo, &safety_info) {
                PushdownSafeType::Safe => {
                    let clause = root.rinfos[rinfo].clause.clone();
                    subquery_push_qual(root, &mut subquery, rti, clause);
                }
                PushdownSafeType::WindowClauseRunCond | PushdownSafeType::Unsafe => upperrestrictlist.push(rinfo),
            }
        }
        root.rels[rti].baserestrictinfo = upperrestrictlist;
    }
    remove_unused_subquery_outputs(root, &mut subquery, rti);
    let parse = &root.parse;
    let tuple_fraction = match parse.has_aggs
        || !parse.group_clause.is_empty()
        || parse.grouping_sets.is_some()
        || root.has_having_qual
        || !parse.distinct_clause.is_empty()
        || !parse.sort_clause.is_empty()
        || root.all_baserels.num_members() > 1
    {
        true => 0.0,
        false => root.tuple_fraction,
    };
    if !plan_subquery_rel(root, rti, subquery, tuple_fraction) {
        set_dummy_rel_pathlist(root, rti);
        return;
    }
    let rel = &root.rels[rti];
    let trivial_pathtarget = rel.reltarget.exprs.len() == rel.attr_widths.len()
        && rel.reltarget.exprs.iter().enumerate().all(|(i, e)| match e {
            Expr::Column(id) => {
                matches!(root.glob.node(*id), VarNode::Var(var) if var.varno == rti && var.varattno == i)
            }
            _ => false,
        });
    for i in 0..root.rels[rti].subplans.len() {
        let order = root.rels[rti].subplans[i].order.clone();
        let pathkeys = super::pathkeys::convert_subquery_pathkeys(root, rti, &order);
        let path = super::pathnode::create_subqueryscan_path(root, rti, i, trivial_pathtarget, pathkeys);
        add_path(&mut root.rels[rti], path);
    }
}

/// plan_subquery_rel plans the subquery of a subquery relation for a share of its rows, keeping the plans of its
/// final paths in the relation's `subplans`, its estimated number of distinct rows, and its size estimates, as Postgres'
/// set_subquery_pathlist and recurse_set_operations plan a subquery and set_subquery_size_estimates sizes its
/// relation, returning false when the subquery returns no rows.
pub fn plan_subquery_rel(
    root: &mut PlannerInfo<'_, '_>,
    rti: usize,
    subquery: super::nodes::Query,
    tuple_fraction: f64,
) -> bool {
    let (tuples, attr_widths, subplans, num_groups, dummy) = {
        let mut subroot = super::subquery_planner(root.ctx, root.glob, subquery, tuple_fraction);
        let sub_final_rel = super::relnode::fetch_upper_rel(
            &mut subroot,
            super::nodes::UpperRelationKind::Final,
            &super::nodes::Relids::new(),
        );
        let dummy = is_dummy_rel(&subroot, sub_final_rel);
        let cheapest = subroot.rels[sub_final_rel].cheapest_total_path.clone().expect("a final path");
        let attr_widths = super::costsize::subquery_attr_widths(&subroot);
        let parse = &subroot.parse;
        let num_groups = match !parse.group_clause.is_empty()
            || parse.grouping_sets.is_some()
            || !parse.distinct_clause.is_empty()
            || subroot.has_having_qual
            || parse.has_aggs
        {
            true => cheapest.rows,
            false => {
                let exprs: Vec<Expr> =
                    parse.target_list.iter().filter(|tle| !tle.resjunk).map(|tle| tle.expr.clone()).collect();
                super::selfuncs::estimate_num_groups(&subroot, &exprs, cheapest.rows, None, None)
            }
        };
        let mut subplans = Vec::new();
        for path in subroot.rels[sub_final_rel].pathlist.clone() {
            let order = super::pathkeys::subquery_output_order(&subroot, &path.pathkeys);
            let plan = super::createplan::create_plan(&mut subroot, &path);
            subplans.push(super::nodes::SubqueryPlan { path, plan, order });
        }
        (cheapest.rows, attr_widths, subplans, num_groups, dummy)
    };
    if dummy {
        return false;
    }
    root.rels[rti].subplans = subplans;
    root.rels[rti].subquery_groups = num_groups;
    super::costsize::set_subquery_size_estimates(root, rti, tuples, attr_widths);
    true
}

/// PushdownSafetyInfo is what subquery_is_pushdown_safe finds of a subquery, as Postgres' pushdown_safety_info holds
/// it: the reasons each output column is unsafe to read in a pushed-down qual, and whether a qual that runs a
/// volatile function is unsafe.
struct PushdownSafetyInfo {
    unsafe_flags: Vec<u8>,
    unsafe_volatile: bool,
}

/// The reasons that an output column of a subquery is unsafe to read in a pushed-down qual, as Postgres' UNSAFE_
/// flags are.
const UNSAFE_HAS_VOLATILE_FUNC: u8 = 1 << 0;
const UNSAFE_HAS_SET_FUNC: u8 = 1 << 1;
const UNSAFE_NOTIN_DISTINCTON_CLAUSE: u8 = 1 << 2;
const UNSAFE_NOTIN_PARTITIONBY_CLAUSE: u8 = 1 << 3;
const UNSAFE_TYPE_MISMATCH: u8 = 1 << 4;

/// PushdownSafeType is whether a qual can be pushed down into a subquery, as Postgres' pushdown_safe_type is.
enum PushdownSafeType {
    Unsafe,
    Safe,
    WindowClauseRunCond,
}

/// subquery_is_pushdown_safe reports whether a subquery, or a leaf of the set operations of the subquery `topquery`,
/// can take quals pushed down into it, noting the columns that they must not read, as Postgres' function of the same
/// name does.
fn subquery_is_pushdown_safe(
    root: &PlannerInfo<'_, '_>,
    subquery: &super::nodes::Query,
    topquery: &super::nodes::Query,
    safety_info: &mut PushdownSafetyInfo,
) -> bool {
    if subquery.limit_offset.is_some() || subquery.limit_count.is_some() {
        return false;
    }
    if !subquery.group_clause.is_empty() && subquery.grouping_sets.is_some() {
        return false;
    }
    if !subquery.distinct_clause.is_empty() || !subquery.window_funcs.is_empty() || !subquery.target_srfs.is_empty() {
        safety_info.unsafe_volatile = true;
    }
    if subquery.set_operations.is_none() {
        check_output_expressions(root, subquery, safety_info);
    }
    if std::ptr::eq(subquery, topquery) {
        if let Some(setops) = &subquery.set_operations {
            let tree = super::nodes::SetOpTree::Op(setops.clone());
            return recurse_pushdown_safe(root, &tree, topquery, safety_info);
        }
    } else {
        if subquery.set_operations.is_some() {
            return false;
        }
        let topop = topquery.set_operations.as_ref().expect("the leaf of a set operation");
        compare_tlist_datatypes(root, subquery, &topop.col_types, safety_info);
    }
    true
}

/// recurse_pushdown_safe reports whether each leaf of a tree of set operations can take pushed-down quals, as
/// Postgres' function of the same name does: none can under an EXCEPT.
fn recurse_pushdown_safe(
    root: &PlannerInfo<'_, '_>,
    set_op: &super::nodes::SetOpTree,
    topquery: &super::nodes::Query,
    safety_info: &mut PushdownSafetyInfo,
) -> bool {
    match set_op {
        super::nodes::SetOpTree::Rel(rtindex) => {
            let RteKind::Subquery(subquery, _) = &topquery.rte(*rtindex).kind else { unreachable!("a leaf subquery") };
            subquery_is_pushdown_safe(root, subquery, topquery, safety_info)
        }
        super::nodes::SetOpTree::Op(op) => {
            op.op != crate::plan::SetOp::Except
                && recurse_pushdown_safe(root, &op.larg, topquery, safety_info)
                && recurse_pushdown_safe(root, &op.rarg, topquery, safety_info)
        }
    }
}

/// compare_tlist_datatypes notes the output columns of a leaf of a set operation whose types differ from the set
/// operation's, which a pushed-down qual must not read, as Postgres' function of the same name does.
fn compare_tlist_datatypes(
    root: &PlannerInfo<'_, '_>,
    subquery: &super::nodes::Query,
    col_types: &[Option<u32>],
    safety_info: &mut PushdownSafetyInfo,
) {
    let visible = subquery.target_list.iter().enumerate().filter(|(_, tle)| !tle.resjunk);
    for ((i, tle), col_type) in visible.zip(col_types) {
        if super::nodefuncs::query_expr_type(root.glob, subquery, &tle.expr) != *col_type {
            safety_info.unsafe_flags[i] |= UNSAFE_TYPE_MISMATCH;
        }
    }
}

/// check_output_expressions notes the output columns of a subquery that a pushed-down qual must not read, as
/// Postgres' function of the same name does: those that return sets or run volatile functions, those outside a
/// DISTINCT ON list, and those outside the partitioning of any window.
fn check_output_expressions(
    root: &PlannerInfo<'_, '_>,
    subquery: &super::nodes::Query,
    safety_info: &mut PushdownSafetyInfo,
) {
    for (i, tle) in subquery.target_list.iter().enumerate() {
        if tle.resjunk {
            continue;
        }
        let flags = &mut safety_info.unsafe_flags[i];
        if !subquery.target_srfs.is_empty()
            && *flags & UNSAFE_HAS_SET_FUNC == 0
            && super::clauses::expression_returns_set(&tle.expr)
        {
            *flags |= UNSAFE_HAS_SET_FUNC;
            continue;
        }
        if *flags & UNSAFE_HAS_VOLATILE_FUNC == 0 && super::clauses::contain_volatile_functions(root.glob, &tle.expr) {
            *flags |= UNSAFE_HAS_VOLATILE_FUNC;
            continue;
        }
        let in_distinct_on = tle.ressortgroupref != 0
            && subquery.distinct_clause.iter().any(|c| c.tle_sort_group_ref == tle.ressortgroupref);
        if subquery.has_distinct_on && *flags & UNSAFE_NOTIN_DISTINCTON_CLAUSE == 0 && !in_distinct_on {
            *flags |= UNSAFE_NOTIN_DISTINCTON_CLAUSE;
            continue;
        }
        let in_all_partitions = subquery.window_funcs.iter().all(|call| call.partition.contains(&tle.expr));
        if !subquery.window_funcs.is_empty() && *flags & UNSAFE_NOTIN_PARTITIONBY_CLAUSE == 0 && !in_all_partitions {
            *flags |= UNSAFE_NOTIN_PARTITIONBY_CLAUSE;
            continue;
        }
    }
}

/// qual_is_pushdown_safe reports whether a restriction of a subquery relation can be pushed down into the subquery,
/// as Postgres' function of the same name does: it has no subquery, runs no volatile function where the subquery
/// forbids it, and reads only the relation's columns that are safe to read. Its equality never conflicts with the
/// subquery's grouping, since each Doltgres type has one equality.
fn qual_is_pushdown_safe(
    root: &PlannerInfo<'_, '_>,
    rti: usize,
    rinfo: super::nodes::RinfoId,
    safety_info: &PushdownSafetyInfo,
) -> PushdownSafeType {
    let qual = &root.rinfos[rinfo].clause;
    if super::clauses::contain_subplans(qual) {
        return PushdownSafeType::Unsafe;
    }
    if safety_info.unsafe_volatile && super::clauses::contain_volatile_functions(root.glob, qual) {
        return PushdownSafeType::Unsafe;
    }
    let mut safe = PushdownSafeType::Safe;
    for id in super::var::pull_var_clause(root.glob, qual, true) {
        let VarNode::Var(var) = root.glob.node(id) else { return PushdownSafeType::Unsafe };
        if var.varno != rti {
            return PushdownSafeType::Unsafe;
        }
        let flags = safety_info.unsafe_flags[var.varattno];
        if flags != 0 {
            let unsafe_flags =
                UNSAFE_HAS_VOLATILE_FUNC | UNSAFE_HAS_SET_FUNC | UNSAFE_NOTIN_DISTINCTON_CLAUSE | UNSAFE_TYPE_MISMATCH;
            if flags & unsafe_flags != 0 {
                return PushdownSafeType::Unsafe;
            }
            safe = PushdownSafeType::WindowClauseRunCond;
        }
    }
    safe
}

/// subquery_push_qual pushes a restriction of a subquery relation down into the subquery, reading its output
/// expressions in place of the relation's columns, into its HAVING when it groups and otherwise into its WHERE, as
/// Postgres' function of the same name does.
fn subquery_push_qual(root: &PlannerInfo<'_, '_>, subquery: &mut super::nodes::Query, rti: usize, qual: Expr) {
    if let Some(setops) = subquery.set_operations.clone() {
        recurse_push_qual(root, &super::nodes::SetOpTree::Op(setops), subquery, rti, &qual);
        return;
    }
    let qual = replace_vars_from_target_list(root, qual, rti, subquery);
    if subquery.has_aggs
        || !subquery.group_clause.is_empty()
        || subquery.grouping_sets.is_some()
        || subquery.having_qual.is_some()
    {
        subquery.having_qual = Some(match subquery.having_qual.take() {
            Some(having) => Expr::And(Box::new(having), Box::new(qual)),
            None => qual,
        });
    } else {
        subquery.jointree.quals.push(qual);
    }
}

/// recurse_push_qual pushes a restriction of a subquery relation down into each leaf of the subquery's set
/// operations, as Postgres' function of the same name does.
fn recurse_push_qual(
    root: &PlannerInfo<'_, '_>,
    set_op: &super::nodes::SetOpTree,
    topquery: &mut super::nodes::Query,
    rti: usize,
    qual: &Expr,
) {
    match set_op {
        super::nodes::SetOpTree::Rel(rtindex) => {
            let RteKind::Subquery(subquery, _) = &mut topquery.rtable[rtindex - 1].kind else {
                unreachable!("a leaf subquery")
            };
            subquery_push_qual(root, subquery, rti, qual.clone());
        }
        super::nodes::SetOpTree::Op(op) => {
            recurse_push_qual(root, &op.larg, topquery, rti, qual);
            recurse_push_qual(root, &op.rarg, topquery, rti, qual);
        }
    }
}

/// replace_vars_from_target_list rewrites each Var of a relation in an expression into the expression of the
/// subquery's output column that it reads, as Postgres' ReplaceVarsFromTargetList does.
fn replace_vars_from_target_list(
    root: &PlannerInfo<'_, '_>,
    e: Expr,
    rti: usize,
    subquery: &super::nodes::Query,
) -> Expr {
    match e {
        Expr::Column(id) => match root.glob.node(id) {
            VarNode::Var(var) if var.varno == rti => subquery.target_list[var.varattno].expr.clone(),
            _ => Expr::Column(id),
        },
        other => other.map_children(&mut |c| replace_vars_from_target_list(root, c, rti, subquery)),
    }
}

/// remove_unused_subquery_outputs replaces each output column of a subquery that the query around it does not read
/// with a NULL, unless its grouping, ordering, or DISTINCT needs it or it returns sets or runs a volatile function, as
/// Postgres' function of the same name does.
fn remove_unused_subquery_outputs(root: &PlannerInfo<'_, '_>, subquery: &mut super::nodes::Query, rti: usize) {
    if subquery.set_operations.is_some() || (!subquery.distinct_clause.is_empty() && !subquery.has_distinct_on) {
        return;
    }
    let rel = &root.rels[rti];
    let mut attrs_used = std::collections::BTreeSet::new();
    let exprs = rel.reltarget.exprs.iter().chain(rel.baserestrictinfo.iter().map(|&r| &root.rinfos[r].clause));
    for e in exprs {
        for id in super::var::pull_var_clause(root.glob, e, false) {
            if let VarNode::Var(var) = root.glob.node(id)
                && var.varno == rti
            {
                attrs_used.insert(var.varattno);
            }
        }
    }
    for (i, tle) in subquery.target_list.iter_mut().enumerate() {
        if tle.ressortgroupref != 0 || tle.resjunk || attrs_used.contains(&i) {
            continue;
        }
        if !subquery.target_srfs.is_empty() && super::clauses::expression_returns_set(&tle.expr) {
            continue;
        }
        if super::clauses::contain_volatile_functions(root.glob, &tle.expr) {
            continue;
        }
        tle.expr = Expr::Const(crate::types::Value::Null);
    }
}

/// set_base_rel_pathlists finds the paths of each base relation, as Postgres' function of the same name does.
fn set_base_rel_pathlists(root: &mut PlannerInfo<'_, '_>) {
    for rti in 1..=root.parse.rtable.len() {
        if root.rels[rti].reloptkind == RelOptKind::BaseRel {
            set_rel_pathlist(root, rti);
        }
    }
}

/// set_rel_pathlist finds the paths of a base relation by its range table entry's kind, as Postgres' function of the
/// same name does: a table's sequential and index paths, the one row of a RESULT relation, or the scan of any other
/// input's plan.
fn set_rel_pathlist(root: &mut PlannerInfo<'_, '_>, rel: usize) {
    if !is_dummy_rel(root, rel) {
        match &root.parse.rte(rel).kind {
            RteKind::Relation(..) => {
                let costs = cost_seqscan(root, rel);
                add_scan_path(root, rel, PathKind::SeqScan, costs);
                create_index_paths(root, rel);
            }
            RteKind::Result => {
                let costs = cost_resultscan(root, rel);
                add_scan_path(root, rel, PathKind::SeqScan, costs);
            }
            RteKind::Plan(plan) => {
                let catalog = matches!(plan, Plan::Catalog(_));
                let costs = cost_opaque_scan(root, rel, catalog);
                add_scan_path(root, rel, PathKind::SeqScan, costs);
                create_index_paths(root, rel);
            }
            RteKind::Subquery(..) => {}
            RteKind::Join(_) => unreachable!("only base relations have paths"),
        }
    }
    set_cheapest(&mut root.rels[rel]);
}

/// add_scan_path adds a path of a base relation that reads its rows, as Postgres' create_seqscan_path,
/// create_resultscan_path, and create_functionscan_path make.
fn add_scan_path(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    kind: PathKind,
    (disabled_nodes, startup_cost, total_cost): Costs,
) {
    let parent = &root.rels[rel];
    let path = Path {
        kind,
        parent: rel,
        relids: parent.relids.clone(),
        param: parent.lateral_relids.clone(),
        pathkeys: Vec::new(),
        rows: parent.rows,
        width: parent.reltarget.width,
        disabled_nodes,
        startup_cost,
        total_cost,
        pathtarget: None,
    };
    add_path(&mut root.rels[rel], Rc::new(path));
}

/// make_rel_from_joinlist returns the relation that joins a joinlist's members, searching for its cheapest paths
/// when it has several, as Postgres' function of the same name does.
fn make_rel_from_joinlist(root: &mut PlannerInfo<'_, '_>, joinlist: &[JoinList]) -> Option<usize> {
    let mut initial_rels = Vec::new();
    for item in joinlist {
        initial_rels.push(match item {
            JoinList::Rel(varno) => *varno,
            JoinList::List(list) => make_rel_from_joinlist(root, list)?,
        });
    }
    match initial_rels.as_slice() {
        [] => None,
        [rel] => Some(*rel),
        _ => {
            root.initial_rels = initial_rels.clone();
            Some(standard_join_search(root, initial_rels))
        }
    }
}

/// standard_join_search finds the join relations of each number of relations in turn, from pairs to all of them,
/// keeping each one's cheapest paths, as Postgres' function of the same name does.
fn standard_join_search(root: &mut PlannerInfo<'_, '_>, initial_rels: Vec<usize>) -> usize {
    let levels_needed = initial_rels.len();
    root.join_rel_level = vec![Vec::new(), initial_rels];
    for level in 2..=levels_needed {
        join_search_one_level(root, level);
        for rel in root.join_rel_level[level].clone() {
            set_cheapest(&mut root.rels[rel]);
        }
    }
    let rel = *root.join_rel_level[levels_needed].first().expect("the join search joins every relation");
    root.join_rel_level.clear();
    rel
}
