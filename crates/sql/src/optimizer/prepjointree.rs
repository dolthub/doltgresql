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

//! Postgres' optimizer/prep/prepjointree.c: preparing the join tree by pulling subqueries up into it, reducing outer
//! joins that conditions above make inner or anti joins, and removing RESULT relations that join nothing.

use std::collections::HashSet;

use super::clauses::{
    contain_nonstrict_functions, contain_subplans, contain_volatile_functions, find_forced_null_var,
    find_forced_null_vars, find_nonnullable_rels, find_nonnullable_vars,
};
use super::nodes::{
    AppendRelInfo, FromExpr, JoinTreeNode, JoinType, PlaceHolderVar, PlannerGlobal, Query, RangeTblEntry, Relids,
    RteKind, SetOpTree, TargetEntry, Var, VarNode,
};
use super::var::{
    add_nulling_relids, change_var_nodes_fn, contain_var_clause, mutate_query, offset_var_nodes,
    remove_nulling_relids_query, replace_columns,
};
use crate::expr::Expr;
use crate::plan::Plan;

/// pull_up_subqueries pulls each simple subquery, single-row VALUES list, and constant function of the join tree up
/// into the query, as Postgres' function of the same name does.
pub fn pull_up_subqueries(glob: &mut PlannerGlobal, parse: &mut Query) {
    for i in 0..parse.jointree.fromlist.len() {
        pull_up_subqueries_recurse(glob, parse, &mut vec![i], None);
    }
}

/// node_at returns the join tree node at a path of indexes into FROM lists and joins' two sides.
fn node_at<'q>(parse: &'q mut Query, path: &[usize]) -> &'q mut JoinTreeNode {
    let mut node = &mut parse.jointree.fromlist[path[0]];
    for &i in &path[1..] {
        node = match node {
            JoinTreeNode::From(from) => &mut from.fromlist[i],
            JoinTreeNode::Join(join) => match i {
                0 => &mut join.larg,
                _ => &mut join.rarg,
            },
            JoinTreeNode::Rel(_) => unreachable!("a relation has no children"),
        };
    }
    node
}

/// pull_up_subqueries_recurse is pull_up_subqueries for the join tree node at a path, given the relations of the
/// lowest outer join above it, replacing the node in place so that the whole join tree's Vars are rewritten as each
/// pull-up needs.
fn pull_up_subqueries_recurse(
    glob: &mut PlannerGlobal,
    parse: &mut Query,
    path: &mut Vec<usize>,
    lowest_outer_join: Option<Relids>,
) {
    let node = node_at(parse, path);
    let (children, lowest_outer_join) = match node {
        JoinTreeNode::Rel(varno) => {
            let varno = *varno;
            if let Some(node) = pull_up_relation(glob, parse, varno, lowest_outer_join.as_ref(), None) {
                *node_at(parse, path) = node;
            }
            return;
        }
        JoinTreeNode::From(from) => (from.fromlist.len(), lowest_outer_join),
        JoinTreeNode::Join(join) if join.jointype != JoinType::Inner => {
            (2, Some(get_relids_in_jointree(node, true, true)))
        }
        JoinTreeNode::Join(_) => (2, lowest_outer_join),
    };
    for i in 0..children {
        path.push(i);
        pull_up_subqueries_recurse(glob, parse, path, lowest_outer_join.clone());
        path.pop();
    }
}

/// pull_up_relation is pull_up_subqueries_recurse for a reference to a range table entry, given the relations of the
/// lowest outer join above it and the member of an append relation that it is, returning the join tree node that
/// replaces it.
fn pull_up_relation(
    glob: &mut PlannerGlobal,
    parse: &mut Query,
    varno: usize,
    lowest_outer_join: Option<&Relids>,
    containing_appendrel: Option<usize>,
) -> Option<JoinTreeNode> {
    let rte = parse.rte(varno);
    match &rte.kind {
        RteKind::Subquery(subquery, _)
            if is_simple_subquery(glob, subquery, rte, lowest_outer_join)
                && containing_appendrel.is_none_or(|_| is_safe_append_member(subquery)) =>
        {
            let lowest_outer_join = lowest_outer_join.cloned();
            Some(pull_up_simple_subquery(glob, parse, varno, lowest_outer_join.as_ref(), containing_appendrel))
        }
        RteKind::Subquery(subquery, _) if is_simple_union_all(subquery) => {
            Some(pull_up_simple_union_all(glob, parse, varno))
        }
        RteKind::Plan(Plan::Values(_)) if lowest_outer_join.is_none() && is_simple_values(glob, parse, varno) => {
            Some(pull_up_simple_values(glob, parse, varno))
        }
        RteKind::Plan(Plan::Function { .. }) => Some(pull_up_constant_function(glob, parse, varno)),
        _ => None,
    }
}

/// pull_up_simple_subquery pulls a simple subquery up into the query, after pulling up its own subqueries, returning
/// the join tree node that replaces its reference, as Postgres' function of the same name does. A member of an
/// append relation changes only that member's expressions.
fn pull_up_simple_subquery(
    glob: &mut PlannerGlobal,
    parse: &mut Query,
    varno: usize,
    lowest_outer_join: Option<&Relids>,
    containing_appendrel: Option<usize>,
) -> JoinTreeNode {
    let rte = parse.rte(varno).clone();
    let RteKind::Subquery(subquery, _) = &rte.kind else { unreachable!("the entry is a subquery") };
    let mut subquery = (**subquery).clone();
    pull_up_subqueries(glob, &mut subquery);
    if !is_simple_subquery(glob, &subquery, &rte, lowest_outer_join)
        || containing_appendrel.is_some_and(|_| !is_safe_append_member(&subquery))
    {
        return JoinTreeNode::Rel(varno);
    }
    let rtoffset = parse.rtable.len();
    offset_var_nodes(glob, &mut subquery, rtoffset);
    if rte.lateral {
        read_lateral_columns(glob, &mut subquery, &rte.lateral_columns);
    }
    let targetlist: Vec<Expr> = subquery.target_list.iter().map(|tle| tle.expr.clone()).collect();
    let wrap_option = if parse.grouping_sets.is_some() { ReplaceWrap::All } else { ReplaceWrap::None };
    let lateral = rte.lateral.then(|| LateralPullup {
        relids: get_relids_in_jointree(&JoinTreeNode::From(Box::new(subquery.jointree.clone())), true, true),
        nullingrels: get_nullingrels(parse),
    });
    let mut rvcontext = PullupReplaceVars {
        rv_cache: vec![None; targetlist.len()],
        targetlist,
        varno,
        wrap_option,
        lateral,
        done: HashSet::new(),
    };
    perform_pullup_replace_vars(glob, parse, &mut rvcontext, containing_appendrel);
    parse.rtable.extend(std::mem::take(&mut subquery.rtable));
    if glob.last_ph_id() != 0 || !parse.append_rel_list.is_empty() {
        let subrelids = get_relids_in_jointree(&JoinTreeNode::From(Box::new(subquery.jointree.clone())), true, false);
        if glob.last_ph_id() != 0 {
            substitute_phv_relids(glob, parse, varno, &subrelids);
        }
        fix_append_rel_relids(parse, varno, &subrelids);
    }
    parse.append_rel_list.extend(std::mem::take(&mut subquery.append_rel_list));
    parse.rtable[varno - 1].kind = RteKind::Result;
    match (subquery.jointree.quals.is_empty(), subquery.jointree.fromlist.len()) {
        (true, 1) => subquery.jointree.fromlist.pop().expect("one member"),
        _ => JoinTreeNode::From(Box::new(subquery.jointree)),
    }
}

/// is_simple_subquery reports whether a subquery can be pulled up into its parent, given the relations of the lowest
/// outer join above it, as Postgres' function of the same name decides: it has no set operations, aggregates, windows,
/// set-returning functions, grouping, HAVING, ORDER BY, DISTINCT, or LIMIT, its output runs no volatile function, and
/// a lateral subquery reads no relations of the query around it that an outer join would make unsafe to read.
fn is_simple_subquery(
    glob: &PlannerGlobal,
    subquery: &Query,
    rte: &RangeTblEntry,
    lowest_outer_join: Option<&Relids>,
) -> bool {
    if rte.lateral {
        let none = Relids::new();
        let safe_upper_varnos = lowest_outer_join.unwrap_or(&none);
        let jointree = JoinTreeNode::From(Box::new(subquery.jointree.clone()));
        if jointree_contains_lateral_outer_refs(
            glob,
            &jointree,
            lowest_outer_join.is_some(),
            safe_upper_varnos,
            &rte.lateral_columns,
        ) {
            return false;
        }
        if let Some(safe_upper_varnos) = lowest_outer_join {
            let tlist: Vec<Expr> = subquery.target_list.iter().map(|tle| tle.expr.clone()).collect();
            if !lateral_varnos(glob, &tlist, &rte.lateral_columns).is_subset(safe_upper_varnos) {
                return false;
            }
        }
    }
    !(subquery.set_operations.is_some()
        || subquery.has_aggs
        || !subquery.window_funcs.is_empty()
        || !subquery.target_srfs.is_empty()
        || !subquery.group_clause.is_empty()
        || subquery.grouping_sets.is_some()
        || subquery.having_qual.is_some()
        || !subquery.sort_clause.is_empty()
        || !subquery.distinct_clause.is_empty()
        || subquery.limit_offset.is_some()
        || subquery.limit_count.is_some())
        && !subquery.target_list.iter().any(|tle| contain_volatile_functions(glob, &tle.expr))
}

/// jointree_contains_lateral_outer_refs reports whether the quals of a lateral subquery's join tree read relations of
/// the query around it that it must not read: any at all in or below an outer join, and otherwise, when restricted,
/// those outside the given relations, as Postgres' function of the same name decides.
fn jointree_contains_lateral_outer_refs(
    glob: &PlannerGlobal,
    jtnode: &JoinTreeNode,
    restricted: bool,
    safe_upper_varnos: &Relids,
    lateral_columns: &[Expr],
) -> bool {
    match jtnode {
        JoinTreeNode::Rel(_) => false,
        JoinTreeNode::From(f) => {
            f.fromlist
                .iter()
                .any(|n| jointree_contains_lateral_outer_refs(glob, n, restricted, safe_upper_varnos, lateral_columns))
                || (restricted && !lateral_varnos(glob, &f.quals, lateral_columns).is_subset(safe_upper_varnos))
        }
        JoinTreeNode::Join(j) => {
            let none = Relids::new();
            let (restricted, safe_upper_varnos) = match j.jointype {
                JoinType::Inner => (restricted, safe_upper_varnos),
                _ => (true, &none),
            };
            jointree_contains_lateral_outer_refs(glob, &j.larg, restricted, safe_upper_varnos, lateral_columns)
                || jointree_contains_lateral_outer_refs(glob, &j.rarg, restricted, safe_upper_varnos, lateral_columns)
                || (restricted && !lateral_varnos(glob, &j.quals, lateral_columns).is_subset(safe_upper_varnos))
        }
    }
}

/// lateral_varnos returns the relations of the query around a lateral subquery that expressions of the subquery read
/// through the columns of its enclosing row, as Postgres' pull_varnos_of_level does for level 1.
fn lateral_varnos(glob: &PlannerGlobal, exprs: &[Expr], lateral_columns: &[Expr]) -> Relids {
    let mut varnos = Relids::new();
    for e in exprs {
        e.visit(&mut |x| {
            if let Expr::Outer(1, c) = x {
                varnos.add_members(&expr_varnos(glob, &lateral_columns[*c]));
            }
        });
    }
    varnos
}

/// expr_varnos returns the relations that an expression's Vars and PlaceHolderVars read, with the outer joins that can
/// make them NULL, as Postgres' pull_varnos does before PlaceHolderInfos exist.
fn expr_varnos(glob: &PlannerGlobal, e: &Expr) -> Relids {
    let mut varnos = Relids::new();
    super::var::visit_columns(e, &mut |id| match glob.node(id) {
        VarNode::Var(var) => {
            varnos.add_member(var.varno);
            varnos.add_members(&var.varnullingrels);
        }
        VarNode::PlaceHolderVar(phv) => {
            varnos.add_members(&glob.placeholder(phv.phid).phrels);
            varnos.add_members(&phv.phnullingrels);
        }
    });
    varnos
}

/// read_lateral_columns rewrites a pulled-up lateral subquery's reads of its enclosing row into the expressions of the
/// query around it that its columns hold, marking its own relations lateral, as Postgres' pull_up_simple_subquery does
/// with IncrementVarSublevelsUp and its LATERAL markers.
fn read_lateral_columns(glob: &mut PlannerGlobal, subquery: &mut Query, lateral_columns: &[Expr]) {
    let mut done = HashSet::new();
    let mut f = |e: Expr| {
        let e = super::read_left_vars(e, lateral_columns);
        read_lateral_placeholders(glob, &e, lateral_columns, &mut done);
        e
    };
    for e in subquery.upper_exprs_mut() {
        *e = f(std::mem::replace(e, Expr::Const(crate::types::Value::Null)));
    }
    for e in subquery.append_rel_list.iter_mut().flat_map(|appinfo| appinfo.translated_vars.iter_mut()) {
        *e = f(std::mem::replace(e, Expr::Const(crate::types::Value::Null)));
    }
    let mut node = JoinTreeNode::From(Box::new(std::mem::take(&mut subquery.jointree)));
    super::var::mutate_jointree(&mut node, &mut f);
    let JoinTreeNode::From(jointree) = node else { unreachable!("the join tree stays a FROM list") };
    subquery.jointree = *jointree;
    for rte in &mut subquery.rtable {
        match &mut rte.kind {
            RteKind::Plan(plan) => {
                plan.map_exprs(0, &mut |e, depth| match depth {
                    0 => super::read_left_vars(e, lateral_columns),
                    _ => super::subselect::decrement_sublevels_up(e, depth),
                });
                rte.lateral = true;
            }
            RteKind::Subquery(inner, _) if rte.lateral => {
                let mut read = Vec::new();
                shift_query_reads(inner, 2, rte.lateral_columns.len(), &mut read);
                let own = std::mem::take(&mut rte.lateral_columns).into_iter();
                let own = own.map(|e| super::read_left_vars(e, lateral_columns));
                rte.lateral_columns = own.chain(read.iter().map(|&c| lateral_columns[c].clone())).collect();
            }
            RteKind::Subquery(..) => {
                rte.lateral = true;
                rte.lateral_columns = lateral_columns.to_vec();
            }
            _ => {}
        }
    }
}

/// shift_query_reads rewrites a lateral query that a pulled-up lateral subquery held, whose expressions read the
/// subquery's enclosing row `level` rows out, to read that row's columns as lateral reads of its own, as
/// shift_enclosing_reads rewrites each expression. Its lateral subqueries read the row one level further out.
fn shift_query_reads(query: &mut Query, level: usize, offset: usize, read: &mut Vec<usize>) {
    let take = |e: &mut Expr| std::mem::replace(e, Expr::Const(crate::types::Value::Null));
    for e in query.upper_exprs_mut() {
        *e = shift_enclosing_reads(take(e), level, offset, read);
    }
    for e in query.append_rel_list.iter_mut().flat_map(|appinfo| appinfo.translated_vars.iter_mut()) {
        *e = shift_enclosing_reads(take(e), level, offset, read);
    }
    let mut node = JoinTreeNode::From(Box::new(std::mem::take(&mut query.jointree)));
    super::var::mutate_jointree(&mut node, &mut |e| shift_enclosing_reads(e, level, offset, read));
    let JoinTreeNode::From(jointree) = node else { unreachable!("the join tree stays a FROM list") };
    query.jointree = *jointree;
    for rte in &mut query.rtable {
        for e in &mut rte.lateral_columns {
            *e = shift_enclosing_reads(take(e), level, offset, read);
        }
        match &mut rte.kind {
            RteKind::Plan(plan) => {
                plan.map_exprs(0, &mut |e, depth| shift_enclosing_reads(e, level + depth, offset, read));
            }
            RteKind::Subquery(inner, _) => shift_query_reads(inner, level + usize::from(rte.lateral), offset, read),
            _ => {}
        }
    }
}

/// shift_enclosing_reads rewrites an expression that reads a pulled-up lateral subquery's enclosing row `level` rows
/// out to read it one level closer, through the lateral columns of the query that holds the expression, numbering
/// each column it reads by its position in `read` after the query's `offset` lateral columns, and the rows further
/// out one level closer, as Postgres' IncrementVarSublevelsUp moves Vars up a level.
fn shift_enclosing_reads(e: Expr, level: usize, offset: usize, read: &mut Vec<usize>) -> Expr {
    let mut e = match e {
        Expr::Outer(d, c) if d == level => {
            let k = match read.iter().position(|&r| r == c) {
                Some(k) => k,
                None => {
                    read.push(c);
                    read.len() - 1
                }
            };
            return Expr::Outer(level - 1, offset + k);
        }
        Expr::Outer(d, c) if d > level => return Expr::Outer(d - 1, c),
        other => other.map_children(&mut |x| shift_enclosing_reads(x, level, offset, read)),
    };
    for plan in e.subqueries_mut() {
        plan.map_exprs(0, &mut |x, depth| shift_enclosing_reads(x, level + 1 + depth, offset, read));
    }
    e
}

/// read_lateral_placeholders rewrites the shared expressions of the PlaceHolderVars that an expression reads, and of
/// those they read in turn, as read_lateral_columns rewrites the expression itself.
fn read_lateral_placeholders(glob: &mut PlannerGlobal, e: &Expr, lateral_columns: &[Expr], done: &mut HashSet<usize>) {
    let mut found = Vec::new();
    e.visit(&mut |x| {
        if let Expr::Column(id) = x
            && let VarNode::PlaceHolderVar(phv) = glob.node(*id)
            && done.insert(phv.phid)
        {
            found.push(phv.phid);
        }
    });
    for phid in found {
        let phexpr = super::read_left_vars(glob.placeholder(phid).phexpr.clone(), lateral_columns);
        read_lateral_placeholders(glob, &phexpr, lateral_columns, done);
        glob.placeholders[phid - 1].phexpr = phexpr;
    }
}

/// get_nullingrels returns the outer joins that can make each relation of a query's join tree NULL, by its range
/// table index, as Postgres' function of the same name does.
fn get_nullingrels(parse: &Query) -> Vec<Relids> {
    let mut nullingrels = vec![Relids::new(); parse.rtable.len() + 1];
    let jointree = JoinTreeNode::From(Box::new(parse.jointree.clone()));
    get_nullingrels_recurse(&jointree, &Relids::new(), &mut nullingrels);
    nullingrels
}

/// get_nullingrels_recurse is get_nullingrels for a join tree node under outer joins, as Postgres' function of the
/// same name does.
fn get_nullingrels_recurse(jtnode: &JoinTreeNode, upper_nullingrels: &Relids, info: &mut [Relids]) {
    match jtnode {
        JoinTreeNode::Rel(varno) => info[*varno] = upper_nullingrels.clone(),
        JoinTreeNode::From(f) => f.fromlist.iter().for_each(|n| get_nullingrels_recurse(n, upper_nullingrels, info)),
        JoinTreeNode::Join(j) => {
            let local_nullingrels = upper_nullingrels.clone().with_member(j.rtindex);
            let (left, right) = match j.jointype {
                JoinType::Inner => (upper_nullingrels, upper_nullingrels),
                JoinType::Full => (&local_nullingrels, &local_nullingrels),
                JoinType::Right => (&local_nullingrels, upper_nullingrels),
                _ => (upper_nullingrels, &local_nullingrels),
            };
            get_nullingrels_recurse(&j.larg, left, info);
            get_nullingrels_recurse(&j.rarg, right, info);
        }
    }
}

/// pull_up_simple_union_all makes a simple UNION ALL subquery an append relation, adding its leaf subqueries to the
/// range table as its members and pulling them up where they can be, as Postgres' function of the same name does.
fn pull_up_simple_union_all(glob: &mut PlannerGlobal, parse: &mut Query, varno: usize) -> JoinTreeNode {
    let RteKind::Subquery(subquery, _) = &parse.rte(varno).kind else { unreachable!("the entry is a subquery") };
    let subquery = (**subquery).clone();
    let rtoffset = parse.rtable.len();
    parse.rtable.extend(subquery.rtable.iter().cloned());
    if parse.rte(varno).lateral {
        let lateral_columns = parse.rte(varno).lateral_columns.clone();
        for child in &mut parse.rtable[rtoffset..] {
            child.lateral = true;
            child.lateral_columns = lateral_columns.clone();
        }
    }
    let topop = subquery.set_operations.clone().expect("a set operation");
    pull_up_union_leaf_queries(glob, parse, &SetOpTree::Op(topop), varno, &subquery.target_list, rtoffset);
    parse.rtable[varno - 1].inh = true;
    JoinTreeNode::Rel(varno)
}

/// pull_up_union_leaf_queries adds a member of an append relation for each leaf of a tree of UNION ALL operations,
/// whose range table entries were copied into the query after an offset, and pulls each up where it can be, as
/// Postgres' function of the same name does. The set operation's target list gives the columns.
fn pull_up_union_leaf_queries(
    glob: &mut PlannerGlobal,
    parse: &mut Query,
    set_op: &SetOpTree,
    parent_rtindex: usize,
    setop_tlist: &[TargetEntry],
    child_rtoffset: usize,
) {
    match set_op {
        SetOpTree::Rel(rtindex) => {
            let child_rtindex = child_rtoffset + rtindex;
            let translated_vars = make_setop_translation_list(glob, setop_tlist, child_rtindex);
            parse.append_rel_list.push(AppendRelInfo {
                parent_relid: parent_rtindex,
                child_relid: child_rtindex,
                translated_vars,
            });
            let appinfo = parse.append_rel_list.len() - 1;
            pull_up_relation(glob, parse, child_rtindex, None, Some(appinfo));
        }
        SetOpTree::Op(op) => {
            pull_up_union_leaf_queries(glob, parse, &op.larg, parent_rtindex, setop_tlist, child_rtoffset);
            pull_up_union_leaf_queries(glob, parse, &op.rarg, parent_rtindex, setop_tlist, child_rtoffset);
        }
    }
}

/// make_setop_translation_list returns a Var of a member of an append relation for each visible column of a set
/// operation, as Postgres' function of the same name does.
fn make_setop_translation_list(glob: &mut PlannerGlobal, setop_tlist: &[TargetEntry], newvarno: usize) -> Vec<Expr> {
    let width = setop_tlist.iter().filter(|tle| !tle.resjunk).count();
    (0..width).map(|attno| glob.var(newvarno, attno, Relids::new())).collect()
}

/// is_simple_union_all reports whether a subquery is a tree of UNION ALL operations whose leaves all return the
/// operation's column types, without ORDER BY, LIMIT, or recursion, as Postgres' function of the same name decides.
fn is_simple_union_all(subquery: &Query) -> bool {
    let Some(topop) = &subquery.set_operations else { return false };
    if !subquery.sort_clause.is_empty()
        || subquery.limit_offset.is_some()
        || subquery.limit_count.is_some()
        || subquery.recursion.is_some()
    {
        return false;
    }
    is_simple_union_all_recurse(&SetOpTree::Op(topop.clone()), subquery, &topop.col_types)
}

/// is_simple_union_all_recurse reports whether a tree of set operations is all UNION ALL operations whose leaves
/// return the given column types, as Postgres' function of the same name decides.
fn is_simple_union_all_recurse(set_op: &SetOpTree, set_op_query: &Query, col_types: &[Option<u32>]) -> bool {
    match set_op {
        SetOpTree::Rel(rtindex) => set_op_query.rte(*rtindex).coltypes == col_types,
        SetOpTree::Op(op) => {
            op.op == crate::plan::SetOp::Union
                && op.all
                && is_simple_union_all_recurse(&op.larg, set_op_query, col_types)
                && is_simple_union_all_recurse(&op.rarg, set_op_query, col_types)
        }
    }
}

/// is_safe_append_member reports whether a leaf of a UNION ALL can be pulled up into its append relation, as
/// Postgres' function of the same name decides: its join tree is one relation without conditions.
fn is_safe_append_member(subquery: &Query) -> bool {
    let mut jtnode = &subquery.jointree;
    loop {
        if !jtnode.quals.is_empty() {
            return false;
        }
        match jtnode.fromlist.as_slice() {
            [JoinTreeNode::Rel(_)] => return true,
            [JoinTreeNode::From(from)] => jtnode = from,
            _ => return false,
        }
    }
}

/// flatten_simple_union_all makes a query whose set operations are all UNION ALL an append relation of its leaves,
/// as Postgres' function of the same name does. Its columns, which are Vars of no relation here, become Vars of the
/// leftmost leaf's entry, which stands for the append relation while a copy of it becomes the leftmost member.
pub fn flatten_simple_union_all(glob: &mut PlannerGlobal, parse: &mut Query) {
    let mut topop = parse.set_operations.clone().expect("a set operation");
    if parse.recursion.is_some() || !is_simple_union_all_recurse(&SetOpTree::Op(topop.clone()), parse, &topop.col_types)
    {
        return;
    }
    let mut leftmostjtnode = &mut topop.larg;
    while let SetOpTree::Op(op) = leftmostjtnode {
        leftmostjtnode = &mut op.larg;
    }
    let SetOpTree::Rel(leftmost_rti) = *leftmostjtnode else { unreachable!("the leftmost leaf is a relation") };
    parse.rtable.push(parse.rte(leftmost_rti).clone());
    *leftmostjtnode = SetOpTree::Rel(parse.rtable.len());
    parse.rtable[leftmost_rti - 1].inh = true;
    parse.jointree.fromlist = vec![JoinTreeNode::Rel(leftmost_rti)];
    parse.set_operations = None;
    mutate_query(parse, &mut change_var_nodes_fn(glob, 0, leftmost_rti));
    let setop_tlist = parse.target_list.clone();
    pull_up_union_leaf_queries(glob, parse, &SetOpTree::Op(topop), leftmost_rti, &setop_tlist, 0);
}

/// pull_up_simple_values replaces the query's only relation, a VALUES list of one row, by its expressions, as
/// Postgres' function of the same name does.
fn pull_up_simple_values(glob: &mut PlannerGlobal, parse: &mut Query, varno: usize) -> JoinTreeNode {
    let RteKind::Plan(Plan::Values(rows)) = &parse.rte(varno).kind else { unreachable!("the entry is VALUES") };
    let targetlist = rows[0].clone();
    let mut rvcontext = PullupReplaceVars {
        rv_cache: vec![None; targetlist.len()],
        targetlist,
        varno,
        wrap_option: ReplaceWrap::None,
        lateral: None,
        done: HashSet::new(),
    };
    perform_pullup_replace_vars(glob, parse, &mut rvcontext, None);
    parse.rtable[varno - 1].kind = RteKind::Result;
    JoinTreeNode::Rel(varno)
}

/// is_simple_values reports whether a VALUES list can be pulled up, as Postgres' function of the same name decides:
/// it has one row of expressions that return no sets, run no volatile functions, and read no enclosing row, and it is
/// the query's only relation.
fn is_simple_values(glob: &PlannerGlobal, parse: &Query, varno: usize) -> bool {
    let RteKind::Plan(Plan::Values(rows)) = &parse.rte(varno).kind else { return false };
    let reads_outer = |e: &Expr| {
        let mut outer = false;
        e.visit(&mut |x| outer |= matches!(x, Expr::Outer(..) | Expr::SetRef(_)));
        outer
    };
    rows.len() == 1
        && !rows[0].iter().any(|e| contain_volatile_functions(glob, e) || contain_subplans(e) || reads_outer(e))
        && parse.rtable.len() == 1
}

/// pull_up_constant_function replaces a relation of a function call that is a constant of one column by the
/// constant, as Postgres' function of the same name does.
fn pull_up_constant_function(glob: &mut PlannerGlobal, parse: &mut Query, varno: usize) -> JoinTreeNode {
    let RteKind::Plan(Plan::Function { call: call @ Expr::Const(_), ordinality: false, width: 1, defined: None }) =
        &parse.rte(varno).kind
    else {
        return JoinTreeNode::Rel(varno);
    };
    let targetlist = vec![call.clone()];
    let wrap_option = if parse.grouping_sets.is_some() { ReplaceWrap::All } else { ReplaceWrap::None };
    let mut rvcontext =
        PullupReplaceVars { rv_cache: vec![None], targetlist, varno, wrap_option, lateral: None, done: HashSet::new() };
    perform_pullup_replace_vars(glob, parse, &mut rvcontext, None);
    parse.rtable[varno - 1].kind = RteKind::Result;
    JoinTreeNode::Rel(varno)
}

/// ReplaceWrap is when pullup_replace_vars wraps an expression in a PlaceHolderVar, as Postgres' ReplaceWrapOption
/// is: only where an outer join can make it NULL, also where it reads no Var, as a FULL JOIN's condition needs, or
/// always, as grouping sets need.
#[derive(Clone, Copy, PartialEq)]
enum ReplaceWrap {
    None,
    VarFree,
    All,
}

/// PullupReplaceVars is the state of replacing the Vars of a pulled-up relation by its expressions, as Postgres'
/// pullup_replace_vars_context is, with the PlaceHolderVars whose expressions were already rewritten.
struct PullupReplaceVars {
    targetlist: Vec<Expr>,
    varno: usize,
    wrap_option: ReplaceWrap,
    rv_cache: Vec<Option<Expr>>,
    lateral: Option<LateralPullup>,
    done: HashSet<usize>,
}

/// LateralPullup is what replacing the Vars of a pulled-up lateral subquery needs, as Postgres' relids and nullinfo of
/// pullup_replace_vars_context hold it: the relations of the subquery's join tree, and the outer joins that can make
/// each relation of the query NULL, by its range table index.
struct LateralPullup {
    relids: Relids,
    nullingrels: Vec<Relids>,
}

/// perform_pullup_replace_vars replaces the Vars of a pulled-up relation in the query's target list, join tree,
/// lateral relations, and append relation members, or only in the member of an append relation that it is, as
/// Postgres' function of the same name does.
fn perform_pullup_replace_vars(
    glob: &mut PlannerGlobal,
    parse: &mut Query,
    rvcontext: &mut PullupReplaceVars,
    containing_appendrel: Option<usize>,
) {
    if let Some(appinfo) = containing_appendrel {
        let save_wrap_option = rvcontext.wrap_option;
        rvcontext.wrap_option = ReplaceWrap::None;
        let translated_vars = std::mem::take(&mut parse.append_rel_list[appinfo].translated_vars);
        parse.append_rel_list[appinfo].translated_vars =
            translated_vars.into_iter().map(|e| pullup_replace_vars(glob, e, rvcontext)).collect();
        rvcontext.wrap_option = save_wrap_option;
        return;
    }
    for e in parse.upper_exprs_mut() {
        *e = pullup_replace_vars(glob, e.clone(), rvcontext);
    }
    for e in parse.append_rel_list.iter_mut().flat_map(|appinfo| appinfo.translated_vars.iter_mut()) {
        *e = pullup_replace_vars(glob, e.clone(), rvcontext);
    }
    super::var::mutate_lateral_rtes(parse, &mut |e| pullup_replace_vars(glob, e, rvcontext));
    let mut node = JoinTreeNode::From(Box::new(std::mem::replace(
        &mut parse.jointree,
        FromExpr { fromlist: Vec::new(), quals: Vec::new() },
    )));
    replace_vars_in_jointree(glob, &mut node, rvcontext);
    let JoinTreeNode::From(jointree) = node else { unreachable!("the join tree stays a FROM list") };
    parse.jointree = *jointree;
}

/// replace_vars_in_jointree replaces the Vars of a pulled-up relation in a join tree's quals, as Postgres' function
/// of the same name does.
fn replace_vars_in_jointree(glob: &mut PlannerGlobal, node: &mut JoinTreeNode, context: &mut PullupReplaceVars) {
    match node {
        JoinTreeNode::Rel(_) => {}
        JoinTreeNode::From(from) => {
            from.fromlist.iter_mut().for_each(|n| replace_vars_in_jointree(glob, n, context));
            from.quals =
                std::mem::take(&mut from.quals).into_iter().map(|q| pullup_replace_vars(glob, q, context)).collect();
        }
        JoinTreeNode::Join(join) => {
            replace_vars_in_jointree(glob, &mut join.larg, context);
            replace_vars_in_jointree(glob, &mut join.rarg, context);
            let save_wrap_option = context.wrap_option;
            if join.jointype == JoinType::Full {
                context.wrap_option = ReplaceWrap::VarFree;
            }
            join.quals =
                std::mem::take(&mut join.quals).into_iter().map(|q| pullup_replace_vars(glob, q, context)).collect();
            context.wrap_option = save_wrap_option;
        }
    }
}

/// pullup_replace_vars replaces the Vars of a pulled-up relation in an expression, and in the expressions of its
/// PlaceHolderVars, as Postgres' function of the same name does with replace_rte_variables.
fn pullup_replace_vars(glob: &mut PlannerGlobal, e: Expr, context: &mut PullupReplaceVars) -> Expr {
    replace_columns(e, &mut |id| match glob.node(id).clone() {
        VarNode::Var(var) if var.varno == context.varno => pullup_replace_vars_callback(glob, var, context),
        VarNode::PlaceHolderVar(phv) => {
            if context.done.insert(phv.phid) {
                let phexpr = glob.placeholder(phv.phid).phexpr.clone();
                glob.placeholders[phv.phid - 1].phexpr = pullup_replace_vars(glob, phexpr, context);
            }
            Expr::Column(id)
        }
        VarNode::Var(_) => Expr::Column(id),
    })
}

/// pullup_replace_vars_callback returns the expression that replaces a Var of a pulled-up relation: its expression,
/// wrapped in a PlaceHolderVar where an outer join can make the Var NULL but not the expression, with the Var's
/// nulling relations added, as Postgres' function of the same name does.
fn pullup_replace_vars_callback(glob: &mut PlannerGlobal, var: Var, rcon: &mut PullupReplaceVars) -> Expr {
    let varattno = var.varattno;
    let need_phv = !var.varnullingrels.is_empty() || rcon.wrap_option != ReplaceWrap::None;
    let newnode = match &rcon.rv_cache[varattno] {
        Some(cached) if need_phv => cached.clone(),
        _ => {
            let mut newnode = rcon.targetlist[varattno].clone();
            if need_phv {
                let nulled_apart = |lateral: &LateralPullup, lvarno: usize| {
                    !lateral.nullingrels[rcon.varno].is_subset(&lateral.nullingrels[lvarno])
                };
                let wrap = match (&newnode, &rcon.lateral) {
                    _ if rcon.wrap_option == ReplaceWrap::All => true,
                    (Expr::Column(id), Some(lateral)) => match glob.node(*id) {
                        VarNode::Var(var) => !lateral.relids.is_member(var.varno) && nulled_apart(lateral, var.varno),
                        VarNode::PlaceHolderVar(phv) => {
                            let phrels = &glob.placeholder(phv.phid).phrels;
                            !phrels.is_subset(&lateral.relids) && phrels.members().any(|l| nulled_apart(lateral, l))
                        }
                    },
                    (Expr::Column(_), None) => false,
                    (other, lateral) => {
                        let contain_nullable_vars = match lateral {
                            None => contain_var_clause(other),
                            Some(lateral) => {
                                let all_varnos = expr_varnos(glob, other);
                                all_varnos.overlap(&lateral.relids)
                                    || all_varnos
                                        .members()
                                        .any(|v| lateral.nullingrels[rcon.varno].is_subset(&lateral.nullingrels[v]))
                            }
                        };
                        !(contain_nullable_vars && !contain_nonstrict_functions(glob, other))
                    }
                };
                if wrap {
                    newnode = glob.make_placeholder_expr(newnode, Relids::singleton(rcon.varno));
                    rcon.rv_cache[varattno] = Some(newnode.clone());
                }
            }
            newnode
        }
    };
    if var.varnullingrels.is_empty() {
        return newnode;
    }
    match newnode {
        Expr::Column(id) => match glob.node(id).clone() {
            VarNode::Var(newvar) => {
                let varnullingrels = newvar.varnullingrels.union(&var.varnullingrels);
                glob.intern(VarNode::Var(Var { varnullingrels, ..newvar }))
            }
            VarNode::PlaceHolderVar(newphv) => {
                let phnullingrels = newphv.phnullingrels.union(&var.varnullingrels);
                glob.intern(VarNode::PlaceHolderVar(PlaceHolderVar { phnullingrels, ..newphv }))
            }
        },
        other => add_nulling_relids(glob, other, None, &var.varnullingrels),
    }
}

/// ReduceState is what the first pass of reduce_outer_joins learns about a join tree node: the relations under it,
/// whether an outer join is under it, and the same of its children, as Postgres' reduce_outer_joins_pass1_state is.
struct ReduceState {
    relids: Relids,
    contains_outer: bool,
    sub_states: Vec<ReduceState>,
}

/// ReduceResult is what the second pass of reduce_outer_joins reduced, as Postgres' reduce_outer_joins_pass2_state
/// is: the outer joins that became inner joins, each FULL JOIN that became a left join with the relations of the side
/// it still makes NULL, and the left joins that became anti joins.
#[derive(Default)]
struct ReduceResult {
    inner_reduced: Relids,
    partial_reduced: Vec<(usize, Relids)>,
    anti_reduced: Relids,
}

/// reduce_outer_joins turns outer joins into inner joins where a stricter condition above rejects the rows that
/// they pad with NULLs, and left joins into anti joins where a condition above keeps only those rows, removing the
/// reduced joins from the nulling relations of the query's Vars, as Postgres' reduce_outer_joins does.
pub fn reduce_outer_joins(glob: &mut PlannerGlobal, parse: &mut Query) {
    let state1 = reduce_outer_joins_pass1(&JoinTreeNode::From(Box::new(parse.jointree.clone())));
    if !state1.contains_outer {
        return;
    }
    let mut state2 = ReduceResult::default();
    let mut jointree = parse.jointree.clone();
    reduce_outer_joins_pass2_from(glob, parse, &mut jointree, &state1, &mut state2, &Relids::new(), &[]);
    parse.jointree = jointree;
    if !state2.inner_reduced.is_empty() {
        remove_nulling_relids_query(glob, parse, &state2.inner_reduced, &Relids::new());
    }
    for (full_join_rti, unreduced_side) in &state2.partial_reduced {
        remove_nulling_relids_query(glob, parse, &Relids::singleton(*full_join_rti), unreduced_side);
    }
    if !state2.anti_reduced.is_empty() {
        let antijoins = state2.anti_reduced.clone();
        let mut node = JoinTreeNode::From(Box::new(parse.jointree.clone()));
        remove_redundant_nullability_quals(glob, &mut node, &antijoins);
        let JoinTreeNode::From(jointree) = node else { unreachable!("the join tree stays a FROM list") };
        parse.jointree = *jointree;
    }
}

/// remove_redundant_nullability_quals removes the IS NULL tests of Vars that anti joins already make NULL from a
/// join tree's quals, as Postgres' function of the same name does.
fn remove_redundant_nullability_quals(glob: &PlannerGlobal, node: &mut JoinTreeNode, antijoins: &Relids) {
    let strip = |quals: &mut Vec<Expr>| {
        quals.retain(|clause| {
            !find_forced_null_var(glob, clause).is_some_and(|var| match glob.node(var) {
                VarNode::Var(var) => var.varnullingrels.overlap(antijoins),
                VarNode::PlaceHolderVar(_) => false,
            })
        })
    };
    match node {
        JoinTreeNode::Rel(_) => {}
        JoinTreeNode::From(from) => {
            from.fromlist.iter_mut().for_each(|n| remove_redundant_nullability_quals(glob, n, antijoins));
            strip(&mut from.quals);
        }
        JoinTreeNode::Join(join) => {
            remove_redundant_nullability_quals(glob, &mut join.larg, antijoins);
            remove_redundant_nullability_quals(glob, &mut join.rarg, antijoins);
            strip(&mut join.quals);
        }
    }
}

/// reduce_outer_joins_pass1 finds the relations and outer joins under a join tree node.
fn reduce_outer_joins_pass1(node: &JoinTreeNode) -> ReduceState {
    match node {
        JoinTreeNode::Rel(varno) => {
            ReduceState { relids: Relids::singleton(*varno), contains_outer: false, sub_states: Vec::new() }
        }
        JoinTreeNode::From(f) => {
            let sub_states: Vec<ReduceState> = f.fromlist.iter().map(reduce_outer_joins_pass1).collect();
            ReduceState {
                relids: sub_states.iter().fold(Relids::new(), |relids, s| relids.union(&s.relids)),
                contains_outer: sub_states.iter().any(|s| s.contains_outer),
                sub_states,
            }
        }
        JoinTreeNode::Join(j) => {
            let sub_states = vec![reduce_outer_joins_pass1(&j.larg), reduce_outer_joins_pass1(&j.rarg)];
            ReduceState {
                relids: sub_states[0].relids.union(&sub_states[1].relids),
                contains_outer: j.jointype.is_outer() || sub_states.iter().any(|s| s.contains_outer),
                sub_states,
            }
        }
    }
}

/// union_vars returns the Vars of two lists, as Postgres' mbms_add_members does.
fn union_vars(mut a: Vec<(usize, usize)>, b: &[(usize, usize)]) -> Vec<(usize, usize)> {
    for v in b {
        if !a.contains(v) {
            a.push(*v);
        }
    }
    a
}

/// reduce_outer_joins_pass2_from is reduce_outer_joins_pass2 for a FROM list, given the relations that conditions
/// above make non-nullable and the Vars they force to be NULL.
fn reduce_outer_joins_pass2_from(
    glob: &PlannerGlobal,
    parse: &mut Query,
    f: &mut FromExpr,
    state1: &ReduceState,
    state2: &mut ReduceResult,
    nonnullable_rels: &Relids,
    forced_null_vars: &[(usize, usize)],
) {
    let pass_nonnullable_rels =
        f.quals.iter().fold(nonnullable_rels.clone(), |relids, q| relids.union(&find_nonnullable_rels(glob, q)));
    let pass_forced_null_vars =
        f.quals.iter().fold(forced_null_vars.to_vec(), |vars, q| union_vars(vars, &find_forced_null_vars(glob, q)));
    for (node, sub_state) in f.fromlist.iter_mut().zip(&state1.sub_states) {
        if sub_state.contains_outer {
            reduce_outer_joins_pass2(
                glob,
                parse,
                node,
                sub_state,
                state2,
                &pass_nonnullable_rels,
                &pass_forced_null_vars,
            );
        }
    }
}

/// reduce_outer_joins_pass2 reduces the outer joins under a join tree node, as Postgres' function of the same name
/// does.
fn reduce_outer_joins_pass2(
    glob: &PlannerGlobal,
    parse: &mut Query,
    node: &mut JoinTreeNode,
    state1: &ReduceState,
    state2: &mut ReduceResult,
    nonnullable_rels: &Relids,
    forced_null_vars: &[(usize, usize)],
) {
    let j = match node {
        JoinTreeNode::From(f) => {
            return reduce_outer_joins_pass2_from(glob, parse, f, state1, state2, nonnullable_rels, forced_null_vars);
        }
        JoinTreeNode::Rel(_) => unreachable!("the second pass reaches only nodes with outer joins"),
        JoinTreeNode::Join(j) => j,
    };
    let rtindex = j.rtindex;
    let (mut left_state, mut right_state) = (&state1.sub_states[0], &state1.sub_states[1]);
    let mut jointype = j.jointype;
    match jointype {
        JoinType::Left if nonnullable_rels.overlap(&right_state.relids) => jointype = JoinType::Inner,
        JoinType::Right if nonnullable_rels.overlap(&left_state.relids) => jointype = JoinType::Inner,
        JoinType::Full => {
            match (nonnullable_rels.overlap(&left_state.relids), nonnullable_rels.overlap(&right_state.relids)) {
                (true, true) => jointype = JoinType::Inner,
                (true, false) => {
                    jointype = JoinType::Left;
                    state2.partial_reduced.push((rtindex, right_state.relids.clone()));
                }
                (false, true) => {
                    jointype = JoinType::Right;
                    state2.partial_reduced.push((rtindex, left_state.relids.clone()));
                }
                (false, false) => {}
            }
        }
        _ => {}
    }
    if jointype == JoinType::Right {
        std::mem::swap(&mut j.larg, &mut j.rarg);
        std::mem::swap(&mut left_state, &mut right_state);
        jointype = JoinType::Left;
    }
    if jointype == JoinType::Left {
        let nonnullable_vars =
            j.quals.iter().fold(Vec::new(), |vars, q| union_vars(vars, &find_nonnullable_vars(glob, q)));
        let overlap = nonnullable_vars.iter().filter(|v| forced_null_vars.contains(v)).map(|v| v.0);
        if overlap.collect::<Relids>().overlap(&right_state.relids) {
            jointype = JoinType::Anti;
        }
    }
    if rtindex != 0 && jointype != j.jointype {
        parse.rtable[rtindex - 1].kind = RteKind::Join(jointype);
        match jointype {
            JoinType::Inner => state2.inner_reduced.add_member(rtindex),
            JoinType::Anti => state2.anti_reduced.add_member(rtindex),
            _ => {}
        }
    }
    j.jointype = jointype;
    if !left_state.contains_outer && !right_state.contains_outer {
        return;
    }
    let (local_nonnullable_rels, local_forced_null_vars) = match jointype {
        JoinType::Full => (Relids::new(), Vec::new()),
        _ => {
            let mut rels =
                j.quals.iter().fold(Relids::new(), |relids, q| relids.union(&find_nonnullable_rels(glob, q)));
            let mut vars = j.quals.iter().fold(Vec::new(), |vars, q| union_vars(vars, &find_forced_null_vars(glob, q)));
            if matches!(jointype, JoinType::Inner | JoinType::Semi) {
                rels.add_members(nonnullable_rels);
                vars = union_vars(vars, forced_null_vars);
            }
            (rels, vars)
        }
    };
    if left_state.contains_outer {
        let (rels, vars) = match jointype {
            JoinType::Inner | JoinType::Semi => (local_nonnullable_rels.clone(), local_forced_null_vars.clone()),
            JoinType::Full => (Relids::new(), Vec::new()),
            _ => (nonnullable_rels.clone(), forced_null_vars.to_vec()),
        };
        reduce_outer_joins_pass2(glob, parse, &mut j.larg, left_state, state2, &rels, &vars);
    }
    if right_state.contains_outer {
        let (rels, vars) = match jointype {
            JoinType::Full => (Relids::new(), Vec::new()),
            _ => (local_nonnullable_rels, local_forced_null_vars),
        };
        reduce_outer_joins_pass2(glob, parse, &mut j.rarg, right_state, state2, &rels, &vars);
    }
}

/// remove_useless_result_rtes removes the RESULT relations of the join tree that join nothing, which pulled-up
/// subqueries and VALUES lists leave, along with the left joins to them, as Postgres' function of the same name does.
pub fn remove_useless_result_rtes(glob: &mut PlannerGlobal, parse: &mut Query) {
    let baserels = match glob.last_ph_id() {
        0 => Relids::new(),
        _ => get_relids_in_jointree(&JoinTreeNode::From(Box::new(parse.jointree.clone())), false, false),
    };
    let mut dropped_outer_joins = Relids::new();
    let mut node = JoinTreeNode::From(Box::new(parse.jointree.clone()));
    remove_useless_results_recurse(glob, parse, &mut node, &baserels, None, &mut dropped_outer_joins, true);
    let JoinTreeNode::From(jointree) = node else { unreachable!("the join tree stays a FROM list") };
    parse.jointree = *jointree;
    if !dropped_outer_joins.is_empty() {
        remove_nulling_relids_query(glob, parse, &dropped_outer_joins, &Relids::new());
    }
}

/// remove_useless_results_recurse is remove_useless_result_rtes for a join tree node, given the quals of the node
/// above that its quals may move into, as Postgres' function of the same name does.
fn remove_useless_results_recurse(
    glob: &mut PlannerGlobal,
    parse: &mut Query,
    node: &mut JoinTreeNode,
    baserels: &Relids,
    parent_quals: Option<&mut Vec<Expr>>,
    dropped_outer_joins: &mut Relids,
    top: bool,
) {
    match node {
        JoinTreeNode::Rel(_) => {}
        JoinTreeNode::From(f) => {
            let mut result_relids = Relids::new();
            let mut i = 0;
            while i < f.fromlist.len() {
                let mut child = std::mem::replace(&mut f.fromlist[i], JoinTreeNode::Rel(0));
                remove_useless_results_recurse(
                    glob,
                    parse,
                    &mut child,
                    baserels,
                    Some(&mut f.quals),
                    dropped_outer_joins,
                    false,
                );
                f.fromlist[i] = child;
                let varno = get_result_relid(parse, &f.fromlist[i]);
                if f.fromlist.len() > 1
                    && varno != 0
                    && !find_dependent_phvs_in_jointree(glob, &JoinTreeNode::From(f.clone()), varno, baserels)
                {
                    f.fromlist.remove(i);
                    result_relids.add_member(varno);
                } else {
                    i += 1;
                }
            }
            if !result_relids.is_empty() {
                let subrelids = get_relids_in_jointree(&JoinTreeNode::From(f.clone()), true, false);
                for varno in result_relids.members() {
                    remove_result_refs(glob, parse, varno, &subrelids);
                }
            }
            if f.fromlist.len() == 1 && !top && (f.quals.is_empty() || parent_quals.is_some()) {
                if let Some(parent_quals) = parent_quals {
                    let quals = std::mem::take(&mut f.quals);
                    parent_quals.splice(0..0, quals);
                }
                *node = f.fromlist.pop().expect("one member");
            }
        }
        JoinTreeNode::Join(j) => {
            let jointype = j.jointype;
            let mut larg = std::mem::replace(&mut j.larg, JoinTreeNode::Rel(0));
            let mut rarg = std::mem::replace(&mut j.rarg, JoinTreeNode::Rel(0));
            let mut parent_quals = parent_quals;
            match jointype {
                JoinType::Inner => {
                    remove_useless_results_recurse(
                        glob,
                        parse,
                        &mut larg,
                        baserels,
                        Some(&mut j.quals),
                        dropped_outer_joins,
                        false,
                    );
                }
                JoinType::Left => {
                    remove_useless_results_recurse(
                        glob,
                        parse,
                        &mut larg,
                        baserels,
                        parent_quals.as_deref_mut(),
                        dropped_outer_joins,
                        false,
                    );
                }
                _ => remove_useless_results_recurse(glob, parse, &mut larg, baserels, None, dropped_outer_joins, false),
            }
            match jointype {
                JoinType::Inner | JoinType::Left => remove_useless_results_recurse(
                    glob,
                    parse,
                    &mut rarg,
                    baserels,
                    Some(&mut j.quals),
                    dropped_outer_joins,
                    false,
                ),
                _ => remove_useless_results_recurse(glob, parse, &mut rarg, baserels, None, dropped_outer_joins, false),
            }
            j.larg = larg;
            j.rarg = rarg;
            let replace_with =
                |keep: JoinTreeNode, quals: Vec<Expr>, parent_quals: Option<&mut Vec<Expr>>| match parent_quals {
                    None if !quals.is_empty() => JoinTreeNode::From(Box::new(FromExpr { fromlist: vec![keep], quals })),
                    Some(parent_quals) => {
                        parent_quals.splice(0..0, quals);
                        keep
                    }
                    None => keep,
                };
            match jointype {
                JoinType::Inner => {
                    let varno = get_result_relid(parse, &j.larg);
                    if varno != 0 && !find_dependent_phvs_in_jointree(glob, &j.rarg, varno, baserels) {
                        let subrelids = get_relids_in_jointree(&j.rarg, true, false);
                        remove_result_refs(glob, parse, varno, &subrelids);
                        let rarg = std::mem::replace(&mut j.rarg, JoinTreeNode::Rel(0));
                        *node = replace_with(rarg, std::mem::take(&mut j.quals), parent_quals);
                        return;
                    }
                    let varno = get_result_relid(parse, &j.rarg);
                    if varno != 0 {
                        let subrelids = get_relids_in_jointree(&j.larg, true, false);
                        remove_result_refs(glob, parse, varno, &subrelids);
                        let larg = std::mem::replace(&mut j.larg, JoinTreeNode::Rel(0));
                        *node = replace_with(larg, std::mem::take(&mut j.quals), parent_quals);
                    }
                }
                JoinType::Left => {
                    let varno = get_result_relid(parse, &j.rarg);
                    if varno != 0 && (j.quals.is_empty() || !find_dependent_phvs(glob, parse, varno, baserels)) {
                        let subrelids = get_relids_in_jointree(&j.larg, true, false);
                        remove_result_refs(glob, parse, varno, &subrelids);
                        dropped_outer_joins.add_member(j.rtindex);
                        *node = std::mem::replace(&mut j.larg, JoinTreeNode::Rel(0));
                    }
                }
                JoinType::Semi => {
                    let varno = get_result_relid(parse, &j.rarg);
                    if varno != 0 {
                        let subrelids = get_relids_in_jointree(&j.larg, true, false);
                        remove_result_refs(glob, parse, varno, &subrelids);
                        let larg = std::mem::replace(&mut j.larg, JoinTreeNode::Rel(0));
                        *node = replace_with(larg, std::mem::take(&mut j.quals), parent_quals);
                    }
                }
                _ => {}
            }
        }
    }
}

/// get_result_relid returns the range table index of a join tree node that is a RESULT relation, or 0, as Postgres'
/// function of the same name does.
fn get_result_relid(parse: &Query, node: &JoinTreeNode) -> usize {
    match node {
        JoinTreeNode::Rel(varno) if matches!(parse.rte(*varno).kind, RteKind::Result) => *varno,
        _ => 0,
    }
}

/// remove_result_refs replaces a removed RESULT relation in the relations of the query's PlaceHolderVars by the
/// relations of the join tree node that took its place, as Postgres' function of the same name does.
fn remove_result_refs(glob: &mut PlannerGlobal, parse: &Query, varno: usize, subrelids: &Relids) {
    if glob.last_ph_id() != 0 {
        substitute_phv_relids(glob, parse, varno, subrelids);
    }
}

/// query_phids returns the IDs of the PlaceHolderVars that a query's expressions and its lateral relations' reads of
/// it hold, with those that their expressions hold.
fn query_phids(glob: &PlannerGlobal, parse: &Query) -> Vec<usize> {
    let mut exprs: Vec<Expr> = parse.upper_exprs();
    exprs.extend(parse.append_rel_list.iter().flat_map(|appinfo| appinfo.translated_vars.iter().cloned()));
    collect_quals(&JoinTreeNode::From(Box::new(parse.jointree.clone())), &mut exprs);
    for rte in parse.rtable.iter().filter(|rte| rte.lateral) {
        exprs.extend(rte.lateral_columns.iter().cloned());
        if let RteKind::Plan(plan) = &rte.kind {
            plan.clone().map_exprs(0, &mut |e, depth| {
                if depth == 0 {
                    exprs.push(e.clone());
                }
                e
            });
        }
    }
    exprs_phids(glob, &exprs)
}

/// collect_quals adds the quals of a join tree node to a list.
fn collect_quals(node: &JoinTreeNode, out: &mut Vec<Expr>) {
    match node {
        JoinTreeNode::Rel(_) => {}
        JoinTreeNode::From(f) => {
            f.fromlist.iter().for_each(|n| collect_quals(n, out));
            out.extend(f.quals.iter().cloned());
        }
        JoinTreeNode::Join(j) => {
            collect_quals(&j.larg, out);
            collect_quals(&j.rarg, out);
            out.extend(j.quals.iter().cloned());
        }
    }
}

/// exprs_phids returns the IDs of the PlaceHolderVars that expressions hold, with those that their expressions hold.
fn exprs_phids(glob: &PlannerGlobal, exprs: &[Expr]) -> Vec<usize> {
    let mut phids: Vec<usize> = Vec::new();
    let mut pending: Vec<Expr> = exprs.to_vec();
    while let Some(e) = pending.pop() {
        super::var::visit_columns(&e, &mut |id| {
            if let VarNode::PlaceHolderVar(phv) = glob.node(id)
                && !phids.contains(&phv.phid)
            {
                phids.push(phv.phid);
                pending.push(glob.placeholder(phv.phid).phexpr.clone());
            }
        });
    }
    phids
}

/// find_dependent_phvs reports whether a PlaceHolderVar of the query is evaluated at a relation alone among the base
/// relations, as Postgres' function of the same name does.
fn find_dependent_phvs(glob: &PlannerGlobal, parse: &Query, varno: usize, baserels: &Relids) -> bool {
    glob.last_ph_id() != 0
        && query_phids(glob, parse)
            .into_iter()
            .any(|phid| glob.placeholder(phid).phrels.intersect(baserels) == Relids::singleton(varno))
}

/// find_dependent_phvs_in_jointree is find_dependent_phvs for the quals of a join tree node.
fn find_dependent_phvs_in_jointree(glob: &PlannerGlobal, node: &JoinTreeNode, varno: usize, baserels: &Relids) -> bool {
    if glob.last_ph_id() == 0 {
        return false;
    }
    let mut quals = Vec::new();
    collect_quals(node, &mut quals);
    exprs_phids(glob, &quals)
        .into_iter()
        .any(|phid| glob.placeholder(phid).phrels.intersect(baserels) == Relids::singleton(varno))
}

/// substitute_phv_relids replaces a relation in the relations of the query's PlaceHolderVars by others, as Postgres'
/// function of the same name does.
fn substitute_phv_relids(glob: &mut PlannerGlobal, parse: &Query, varno: usize, subrelids: &Relids) {
    for phid in query_phids(glob, parse) {
        let placeholder = &mut glob.placeholders[phid - 1];
        if placeholder.phrels.is_member(varno) {
            placeholder.phrels = placeholder.phrels.union(subrelids).without_member(varno);
        }
    }
}

/// fix_append_rel_relids replaces a pulled-up member of an append relation by the one relation of its join tree, as
/// Postgres' function of the same name does, where substitute_phv_relids has already fixed the PlaceHolderVars of the
/// members' expressions.
fn fix_append_rel_relids(parse: &mut Query, varno: usize, subrelids: &Relids) {
    for appinfo in parse.append_rel_list.iter_mut().filter(|appinfo| appinfo.child_relid == varno) {
        appinfo.child_relid = subrelids.singleton_member().expect("a member pulls up into one relation");
    }
}

/// get_relids_in_jointree returns the base relations of a join tree node, with the outer or inner joins under it
/// when asked, as Postgres' function of the same name does.
pub fn get_relids_in_jointree(node: &JoinTreeNode, include_outer_joins: bool, include_inner_joins: bool) -> Relids {
    match node {
        JoinTreeNode::Rel(varno) => Relids::singleton(*varno),
        JoinTreeNode::From(f) => f.fromlist.iter().fold(Relids::new(), |relids, n| {
            relids.union(&get_relids_in_jointree(n, include_outer_joins, include_inner_joins))
        }),
        JoinTreeNode::Join(j) => {
            let mut result = get_relids_in_jointree(&j.larg, include_outer_joins, include_inner_joins)
                .union(&get_relids_in_jointree(&j.rarg, include_outer_joins, include_inner_joins));
            let include = match j.jointype {
                JoinType::Inner => include_inner_joins,
                _ => include_outer_joins,
            };
            if j.rtindex != 0 && include {
                result.add_member(j.rtindex);
            }
            result
        }
    }
}

/// has_outer_joins reports whether the query's join tree holds an outer join, as Postgres' hasOuterJoins records.
pub fn has_outer_joins(parse: &Query) -> bool {
    parse.rtable.iter().any(|rte| matches!(rte.kind, RteKind::Join(jointype) if jointype.is_outer()))
}
