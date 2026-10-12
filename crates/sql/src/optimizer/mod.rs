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

//! A port of Postgres 18's optimizer, file by file: each module ports the Postgres source file it names, keeping
//! its functions' names and logic, over Doltgres' bound expressions and plans. Planning starts from a SELECT's FROM
//! clause, as the binder built it without choosing join orders or methods, and its WHERE clause, which it turns back
//! into Postgres' Query, and produces the plan that Postgres' query_planner would find cheapest.

mod allpaths;
mod analyzejoins;
mod appendinfo;
mod bipartite_match;
mod clauses;
mod clausesel;
mod costsize;
mod createplan;
mod equivclass;
mod geqo;
mod indxpath;
mod inherit;
mod initsplan;
mod joininfo;
mod joinpath;
mod joinrels;
mod knapsack;
mod nodefuncs;
pub mod nodes;
mod orclauses;
mod paramassign;
mod pathkeys;
mod pathnode;
mod placeholder;
mod planagg;
mod plancat;
mod planner;
mod predtest;
mod prepagg;
mod prepjointree;
mod prepqual;
mod preptlist;
mod prepunion;
mod query;
mod relnode;
mod restrictinfo;
mod selfuncs;
mod subselect;
mod tlist;
mod var;

use std::collections::HashMap;
use std::rc::Rc;

pub(crate) use subselect::{make_subplan, sublink_convertible, values_convertible};

use nodes::{
    EquivalenceClass, EquivalenceMember, FromExpr, JoinDomain, JoinExpr, JoinTreeNode, JoinType, OuterJoinClauseInfo,
    PathKey, PkId, PlaceHolderInfo, PlannerGlobal, Query, RangeTblEntry, RelOptInfo, Relids, RestrictInfo, RteKind,
    SjId, SpecialJoinInfo,
};

use crate::expr::Expr;
use crate::functions::aggregate::AggCall;
use crate::plan::{JoinKind, JoinMethod, Plan};
use crate::query::Ctx;
use crate::types::Value;

pub(crate) use nodefuncs::btree_opfamily;
pub(crate) use paramassign::replace_outer_agg;
pub(crate) use plancat::infer_arbiter_indexes;
pub(crate) use restrictinfo::or_args;

/// PlannerInfo is the state of planning one query, as Postgres' PlannerInfo holds it. Relations, RestrictInfos,
/// SpecialJoinInfos, equivalence classes and members, and canonical pathkeys live in vectors here and are referred
/// to by index.
pub struct PlannerInfo<'r, 'a> {
    pub ctx: std::cell::RefCell<&'r mut Ctx<'a>>,
    pub glob: &'r mut PlannerGlobal,
    pub parse: Query,
    /// The relations, by index: the simple relations at their range table indexes, then the join relations.
    pub rels: Vec<RelOptInfo>,
    /// The join relations' indexes by their relations.
    pub join_rel_hash: HashMap<Relids, usize>,
    /// The relations of each level of the join search running now, from level 1, and the level being built.
    pub join_rel_level: Vec<Vec<usize>>,
    pub join_cur_level: usize,
    /// The relations that the join search running now joins.
    pub initial_rels: Vec<usize>,
    pub all_baserels: Relids,
    pub outer_join_rels: Relids,
    pub all_query_rels: Relids,
    pub join_domains: Vec<JoinDomain>,
    pub eq_classes: Vec<EquivalenceClass>,
    pub eq_members: Vec<EquivalenceMember>,
    pub ec_merging_done: bool,
    pub canon_pathkeys: Vec<PathKey>,
    /// The mergejoinable clauses of outer joins, by which side the join can make NULL.
    pub left_join_clauses: Vec<OuterJoinClauseInfo>,
    pub right_join_clauses: Vec<OuterJoinClauseInfo>,
    pub full_join_clauses: Vec<OuterJoinClauseInfo>,
    pub sjinfos: Vec<SpecialJoinInfo>,
    /// The outer, semi, and anti joins of the query.
    pub join_info_list: Vec<SjId>,
    pub last_rinfo_serial: usize,
    pub rinfos: Vec<RestrictInfo>,
    pub placeholder_list: Vec<PlaceHolderInfo>,
    /// The index in `placeholder_list` of each PlaceHolderVar's PlaceHolderInfo, by its ID.
    pub placeholder_array: HashMap<usize, usize>,
    pub placeholders_frozen: bool,
    pub has_pseudo_constant_quals: bool,
    pub has_lateral_rtes: bool,
    /// The pages of every table that the query reads, which Postgres' index_pages_fetched shares the cache among.
    pub total_table_pages: f64,
    pub enables: costsize::Enables,
    /// The rows that the query reads of the join of every relation, or a fraction of them below one, or zero for all.
    pub tuple_fraction: f64,
    /// The order of the rows that the query's upper processing asks for, as Postgres' query_pathkeys.
    pub query_pathkeys: Vec<PkId>,
    /// The aggregate calls of an aggregate without groups over the join's rows, which Doltgres' executor answers from
    /// an index's entry counts when they only count rows.
    pub counting: Option<Vec<AggCall>>,
    /// The query's target list and group clauses as the upper planning uses them, as Postgres' processed_tlist and
    /// processed_groupClause.
    pub processed_tlist: Vec<nodes::TargetEntry>,
    pub processed_group_clause: Vec<nodes::SortGroupClause>,
    pub has_having_qual: bool,
    /// The most rows that the query's LIMIT reads of the join of every relation, or -1 for all.
    pub limit_tuples: f64,
    /// The orders that the query's grouping, windows, DISTINCT, and ORDER BY ask for, and how many pathkeys of the
    /// grouping's are its group keys'.
    pub group_pathkeys: Vec<PkId>,
    pub num_groupby_pathkeys: usize,
    pub window_pathkeys: Vec<PkId>,
    pub distinct_pathkeys: Vec<PkId>,
    pub sort_pathkeys: Vec<PkId>,
    /// The foreign keys between the query's base relations, as Postgres' fkey_list.
    pub fkey_list: Vec<nodes::ForeignKeyOptInfo>,
    /// The order that a set operation over the query wants its rows in, as Postgres' setop_pathkeys.
    pub setop_pathkeys: Vec<PkId>,
    /// The query's DISTINCT clauses without redundant ones, as Postgres' processed_distinctClause.
    pub processed_distinct_clause: Vec<nodes::SortGroupClause>,
    /// How many aggregate calls take ordered or DISTINCT input.
    pub num_ordered_aggs: usize,
    /// Whether the query's expressions hold AlternativeSubPlans, as Postgres' hasAlternativeSubPlans records.
    pub has_alternative_subplans: bool,
    /// The MIN and MAX aggregates that indexes answer, as Postgres' minmax_aggs holds them.
    pub minmax_aggs: Vec<nodes::MinMaxAggInfo>,
    /// While making plans, the relations of the outer sides of the nested loops above, as Postgres' curOuterRels.
    pub cur_outer_rels: Relids,
    /// While making plans, the nestloop parameters that plans below read and nested loops above have yet to supply,
    /// as Postgres' curOuterParams holds them, and how many parameters were numbered.
    pub cur_outer_params: std::cell::RefCell<Vec<nodes::NestLoopParam>>,
    pub num_nestloop_params: std::cell::Cell<usize>,
}

impl PlannerInfo<'_, '_> {
    /// find_rel returns the index of the base or join relation of a set of relations, when it was built.
    pub fn find_rel(&self, relids: &Relids) -> Option<usize> {
        match relids.singleton_member() {
            Some(relid) => Some(relid),
            None => self.join_rel_hash.get(relids).copied(),
        }
    }
}

/// planner plans the statement that the binder built a plan of with standard_planner, and raises the error that
/// planning found, as Postgres' function of the same name does: a SELECT, or the rows that an UPDATE, DELETE, or
/// MERGE changes, whose plan reads its result relation first.
pub(crate) fn planner(
    ctx: &mut Ctx<'_>,
    plan: Plan,
    command: nodes::CmdType,
    has_returning: bool,
) -> crate::error::Result<Plan> {
    let plan = standard_planner(ctx, plan, command, has_returning);
    match ctx.planner_error.take() {
        Some(e) => Err(e),
        None => Ok(plan),
    }
}

/// standard_planner plans the statement that the binder built a plan of, as Postgres' function of the same name does:
/// it turns the plan back into a Query and plans it.
fn standard_planner(ctx: &mut Ctx<'_>, plan: Plan, command: nodes::CmdType, has_returning: bool) -> Plan {
    let mut glob = PlannerGlobal::default();
    let mut parse = query::unbind(&mut glob, ctx, plan);
    if command != nodes::CmdType::Select {
        parse.command_type = command;
        let first_relation = parse.rtable.iter().position(|rte| matches!(rte.kind, RteKind::Relation(..)));
        parse.result_relation = first_relation.expect("a result relation") + 1;
        parse.has_returning = has_returning;
    }
    let mut root = subquery_planner(ctx, &mut glob, parse, 0.0, None);
    create_final_plan(&mut root, 0.0).0
}

/// create_final_plan makes the plan of the cheapest path of a planned query's final relation for the rows it reads,
/// whose columns are the query's visible columns, as Postgres' standard_planner and make_subplan choose and plan it,
/// and returns the path too.
fn create_final_plan(root: &mut PlannerInfo<'_, '_>, tuple_fraction: f64) -> (Plan, Rc<nodes::Path>) {
    let final_rel = relnode::fetch_upper_rel(root, nodes::UpperRelationKind::Final, &Relids::new());
    let best_path = planner::get_cheapest_fractional_path(&root.rels[final_rel], tuple_fraction);
    (createplan::create_plan(root, &best_path), best_path)
}

/// plan_subselect plans the subquery of a subquery expression for the share of its rows that the expression reads,
/// and returns its plan and path, as Postgres' make_subplan plans it.
fn plan_subselect(
    ctx: &mut Ctx<'_>,
    glob: &mut PlannerGlobal,
    parse: Query,
    tuple_fraction: f64,
) -> (Plan, Rc<nodes::Path>) {
    let mut root = subquery_planner(ctx, glob, parse, tuple_fraction, None);
    let (mut plan, path) = create_final_plan(&mut root, tuple_fraction);
    crate::indexscan::prune(&mut plan);
    (plan, path)
}

/// preprocess_query_expressions preprocesses each expression of a query, of its append relation members, and of its
/// range table's inputs, conditions with preprocess_qual_conditions, as Postgres' subquery_planner does.
fn preprocess_query_expressions(ctx: &mut Ctx<'_>, parse: &mut Query) {
    for e in parse.upper_exprs_mut() {
        let old = std::mem::replace(e, Expr::SubqueryValue);
        *e = preprocess_expression(ctx, old);
    }
    for e in parse.append_rel_list.iter_mut().flat_map(|appinfo| appinfo.translated_vars.iter_mut()) {
        let old = std::mem::replace(e, Expr::SubqueryValue);
        *e = preprocess_expression(ctx, old);
    }
    let mut quals = |quals: &mut Vec<Expr>| *quals = preprocess_qual_conditions(ctx, std::mem::take(quals));
    parse.jointree.fromlist.iter_mut().for_each(|node| subselect::jointree_quals_mut(node, &mut quals));
    quals(&mut parse.jointree.quals);
    if parse.having_qual.as_ref().is_some_and(|having| *having == Expr::Const(Value::Bool(true))) {
        parse.having_qual = None;
    }
    for rte in &mut parse.rtable {
        if let RteKind::Plan(plan) = &mut rte.kind {
            if matches!(plan, Plan::Function { .. } | Plan::RowsFrom { .. } | Plan::Values(_)) {
                plan.map_exprs(0, &mut |e, depth| match depth {
                    0 => clauses::eval_const_expressions(ctx, e),
                    _ => e,
                });
            }
            plan.map_exprs(0, &mut |e, _| subselect::preprocess_subplans(ctx, e));
        }
    }
}

/// preprocess_phv_expressions preprocesses the expressions of the PlaceHolderVars that pulling up this query's
/// subqueries made, from the given ID, as Postgres' preprocess_expression does inside the PlaceHolderVars of the
/// query's expressions and its preprocess_phv_expression and preprocess_subquery_phvs do for their copies in lateral
/// subqueries and sublinks, since every copy here shares its expression.
fn preprocess_phv_expressions(ctx: &mut Ctx<'_>, glob: &mut PlannerGlobal, first_phid: usize) {
    for phid in first_phid..=glob.last_ph_id() {
        let phexpr = std::mem::replace(&mut glob.placeholders[phid - 1].phexpr, Expr::SubqueryValue);
        glob.placeholders[phid - 1].phexpr = preprocess_expression(ctx, phexpr);
    }
}

/// preprocess_expression simplifies an expression of a query with eval_const_expressions and plans its SubPlans, as
/// Postgres' function of the same name does.
fn preprocess_expression(ctx: &mut Ctx<'_>, e: Expr) -> Expr {
    let e = clauses::eval_const_expressions(ctx, e);
    subselect::preprocess_subplans(ctx, e)
}

/// preprocess_qual_conditions preprocesses a list of conditions that must all hold as their AND, which canonicalize_qual
/// also simplifies, returning its arguments again, as Postgres' preprocess_expression and make_ands_implicit do for a
/// qual: none for TRUE.
fn preprocess_qual_conditions(ctx: &mut Ctx<'_>, quals: Vec<Expr>) -> Vec<Expr> {
    if quals.is_empty() {
        return quals;
    }
    let qual = clauses::eval_const_expressions(ctx, prepqual::make_andclause(quals));
    let qual = subselect::preprocess_subplans(ctx, prepqual::canonicalize_qual(qual));
    match qual {
        Expr::Const(Value::Bool(true)) => Vec::new(),
        qual => restrictinfo::and_args(&qual).into_iter().cloned().collect(),
    }
}

/// subquery_planner prepares a query's join tree, pulling up its sublinks and subqueries, reducing its outer joins,
/// and moving the HAVING conditions that read no aggregate to WHERE, then plans it and charges its final paths for
/// its initplans, returning its planner state, as Postgres' function of the same name does.
fn subquery_planner<'r, 'a>(
    ctx: &'r mut Ctx<'a>,
    glob: &'r mut PlannerGlobal,
    mut parse: Query,
    tuple_fraction: f64,
    setops: Option<&nodes::SetOperationStmt>,
) -> PlannerInfo<'r, 'a> {
    let first_phid = glob.last_ph_id() + 1;
    prepjointree::pull_up_subqueries(glob, &mut parse);
    if parse.set_operations.is_some() {
        prepjointree::flatten_simple_union_all(glob, &mut parse);
    }
    let has_having_qual = parse.having_qual.is_some();
    preprocess_query_expressions(ctx, &mut parse);
    preprocess_phv_expressions(ctx, glob, first_phid);
    preprocess_having(glob, &mut parse);
    if prepjointree::has_outer_joins(&parse) {
        prepjointree::reduce_outer_joins(glob, &mut parse);
    }
    prepjointree::remove_useless_result_rtes(glob, &mut parse);
    prepagg::preprocess_aggrefs(glob, &mut parse);
    let mut root = new_planner_info(ctx, glob, parse, tuple_fraction, has_having_qual);
    subselect::query_exprs(&mut root.parse, &mut |e| {
        root.has_alternative_subplans |= matches!(e, Expr::AlternativeSubPlan(_))
    });
    root.num_ordered_aggs = prepagg::count_ordered_aggs(&root);
    planner::grouping_planner(&mut root, tuple_fraction, setops);
    let final_rel = relnode::fetch_upper_rel(&mut root, nodes::UpperRelationKind::Final, &Relids::new());
    subselect::ss_charge_for_initplans(&mut root, final_rel);
    pathnode::set_cheapest(&mut root.rels[final_rel]);
    root
}

/// new_planner_info returns the planner state of a query before planning it, as Postgres' subquery_planner sets up
/// its PlannerInfo.
fn new_planner_info<'r, 'a>(
    ctx: &'r mut Ctx<'a>,
    glob: &'r mut PlannerGlobal,
    parse: Query,
    tuple_fraction: f64,
    has_having_qual: bool,
) -> PlannerInfo<'r, 'a> {
    let enables = costsize::Enables::read(&ctx.session.settings);
    let counting = (parse.group_clause.is_empty() && parse.grouping_sets.is_none() && parse.has_aggs)
        .then(|| parse.aggregates.clone());
    let has_lateral_rtes = parse.rtable.iter().any(|rte| rte.lateral);
    PlannerInfo {
        ctx: std::cell::RefCell::new(ctx),
        glob,
        parse,
        rels: Vec::new(),
        join_rel_hash: HashMap::new(),
        join_rel_level: Vec::new(),
        join_cur_level: 0,
        initial_rels: Vec::new(),
        all_baserels: Relids::new(),
        outer_join_rels: Relids::new(),
        all_query_rels: Relids::new(),
        join_domains: vec![JoinDomain { jd_relids: Relids::new() }],
        eq_classes: Vec::new(),
        eq_members: Vec::new(),
        ec_merging_done: false,
        canon_pathkeys: Vec::new(),
        left_join_clauses: Vec::new(),
        right_join_clauses: Vec::new(),
        full_join_clauses: Vec::new(),
        sjinfos: Vec::new(),
        join_info_list: Vec::new(),
        last_rinfo_serial: 0,
        rinfos: Vec::new(),
        placeholder_list: Vec::new(),
        placeholder_array: HashMap::new(),
        placeholders_frozen: false,
        has_pseudo_constant_quals: false,
        has_lateral_rtes,
        total_table_pages: 0.0,
        enables,
        tuple_fraction,
        query_pathkeys: Vec::new(),
        counting,
        processed_tlist: Vec::new(),
        processed_group_clause: Vec::new(),
        has_having_qual,
        limit_tuples: -1.0,
        group_pathkeys: Vec::new(),
        num_groupby_pathkeys: 0,
        window_pathkeys: Vec::new(),
        distinct_pathkeys: Vec::new(),
        sort_pathkeys: Vec::new(),
        fkey_list: Vec::new(),
        setop_pathkeys: Vec::new(),
        processed_distinct_clause: Vec::new(),
        num_ordered_aggs: 0,
        has_alternative_subplans: false,
        minmax_aggs: Vec::new(),
        cur_outer_rels: Relids::new(),
        cur_outer_params: Default::default(),
        num_nestloop_params: Default::default(),
    }
}

/// preprocess_having moves each HAVING condition that reads no aggregate, runs no volatile function, and has no
/// subquery to WHERE, keeping it in HAVING too when the query has no group keys, as Postgres' subquery_planner does.
fn preprocess_having(glob: &PlannerGlobal, parse: &mut Query) {
    let Some(having) = parse.having_qual.take() else { return };
    let mut new_having = Vec::new();
    for clause in crate::indexscan::conjuncts(&having).into_iter().cloned() {
        let mut has_agg = false;
        clause.visit(&mut |e| has_agg |= matches!(e, Expr::AggRef(_) | Expr::Grouping(..)));
        let grouped_column =
            !parse.group_clause.is_empty() && parse.grouping_sets.is_some() && var::contain_var_clause(&clause);
        if has_agg
            || clauses::contain_volatile_functions(glob, &clause)
            || crate::plan::has_subquery(&clause)
            || grouped_column
        {
            new_having.push(clause);
        } else if !parse.group_clause.is_empty()
            && parse.grouping_sets.as_ref().is_none_or(|sets| sets.first().is_some_and(|s| !s.is_empty()))
        {
            parse.jointree.quals.push(clause);
        } else {
            parse.jointree.quals.push(clause.clone());
            new_having.push(clause);
        }
    }
    parse.having_qual = new_having.into_iter().reduce(|a, b| Expr::And(Box::new(a), Box::new(b)));
}

/// plan_subquery plans a subquery that its query left unplanned for pulling up, when it stays a relation of its
/// own, as Postgres' subquery_planner plans a subquery that it could not pull up.
fn plan_subquery(ctx: &mut Ctx<'_>, plan: Plan) -> Plan {
    match plan {
        join @ Plan::Join { .. } if !decomposable(&join) => match join {
            Plan::Join { left, right, kind, condition, lateral, method } => {
                let width = left.width() + right.width();
                Plan::Join {
                    left: Box::new(plan_subquery(ctx, *left)),
                    right: Box::new(plan_subquery(ctx, *right)),
                    kind,
                    condition: condition.map(|c| subselect::process_sublinks(ctx, c, width)),
                    lateral,
                    method,
                }
            }
            other => other,
        },
        Plan::SetOp { op, all, left, right } => Plan::SetOp {
            op,
            all,
            left: Box::new(plan_subquery(ctx, *left)),
            right: Box::new(plan_subquery(ctx, *right)),
        },
        Plan::Recursive { work_table, anchor, step, all } => Plan::Recursive {
            work_table,
            anchor: Box::new(plan_subquery(ctx, *anchor)),
            step: Box::new(plan_subquery(ctx, *step)),
            all,
        },
        Plan::Once(input) => Plan::Once(Box::new(plan_subquery(ctx, *input))),
        Plan::OneTimeFilter { input, condition } => {
            Plan::OneTimeFilter { input: Box::new(plan_subquery(ctx, *input)), condition }
        }
        leaf @ (Plan::OneRow
        | Plan::Scan(..)
        | Plan::System(_)
        | Plan::Catalog(_)
        | Plan::CatalogIndexScan(_)
        | Plan::IndexScan(_)
        | Plan::BitmapHeapScan(_)
        | Plan::Values(_)
        | Plan::Function { .. }
        | Plan::RowsFrom { .. }
        | Plan::QueryDiff(..)
        | Plan::XmlTable(_)
        | Plan::JsonTable(_)
        | Plan::WorkTable(..)
        | Plan::Memoize { .. }
        | Plan::MergeAppend { .. }) => leaf,
        other => standard_planner(ctx, other, nodes::CmdType::Select, false),
    }
}

/// unlateral returns the right input of a lateral join without its lateral reach, reading the rows of the enclosing
/// queries one level closer, when it reads nothing of the join's left rows, as Postgres plans a LATERAL item that
/// refers to no item before it as an ordinary relation.
fn unlateral(right: &Plan) -> Option<Plan> {
    let mut read = std::collections::BTreeSet::new();
    if !crate::indexscan::outer_reads(right, 1, &mut read) || !read.is_empty() {
        return None;
    }
    let mut right = right.clone();
    let reached = right.map_exprs(0, &mut |e, depth| subselect::decrement_sublevels_up(e, depth));
    reached.then_some(right)
}

/// decomposable reports whether a join's inputs can join in any order the planner finds: it is not already planned,
/// and it is not lateral or its lateral input reads the left rows in its own expressions only, outside its
/// subqueries, and is a function or VALUES list in an inner join or a subquery in an inner or left join.
fn decomposable(join: &Plan) -> bool {
    match join {
        Plan::Join { lateral: false, method: JoinMethod::Unplanned, .. } => true,
        Plan::Join { right, kind, lateral: true, method: JoinMethod::Unplanned, .. } => {
            let relation = matches!(**right, Plan::Function { .. } | Plan::RowsFrom { .. } | Plan::Values(_));
            let joins = match relation {
                true => *kind == JoinKind::Inner,
                false => matches!(kind, JoinKind::Inner | JoinKind::Left),
            };
            joins && lateral_reads_shallow(right)
        }
        _ => false,
    }
}

/// lateral_reads_shallow reports whether the right input of a lateral join reads the left rows only in its own
/// expressions, outside its subqueries.
pub(super) fn lateral_reads_shallow(right: &Plan) -> bool {
    let mut shallow = true;
    let mut right = right.clone();
    let reached = right.map_exprs(0, &mut |e, depth| {
        if depth > 0 {
            let mut read = std::collections::BTreeSet::new();
            e.visit(&mut |x| {
                if let Expr::Outer(d, c) = x
                    && *d == depth + 1
                {
                    read.insert(*c);
                }
            });
            shallow &= read.is_empty();
        }
        for plan in e.subqueries() {
            let mut read = std::collections::BTreeSet::new();
            shallow &= crate::indexscan::outer_reads(plan, depth + 2, &mut read) && read.is_empty();
        }
        e
    });
    reached && shallow
}

/// lateral_relation adds a lateral function or VALUES list to a range table, reading the Vars of the left rows' columns
/// in place of its reads of them, as Postgres' parser builds a LATERAL item's Vars, and returns its reference.
fn lateral_relation(
    glob: &mut PlannerGlobal,
    mut right: Plan,
    left_columns: &[Expr],
    rtable: &mut Vec<RangeTblEntry>,
    output: &mut Vec<Expr>,
) -> JoinTreeNode {
    right.map_exprs(0, &mut |e, depth| match depth {
        0 => read_left_vars(e, left_columns),
        _ => subselect::decrement_sublevels_up(e, depth),
    });
    let coltypes = plan_coltypes(&right);
    let node = push_relation(glob, rtable, output, RteKind::Plan(right), coltypes);
    rtable.last_mut().expect("the lateral relation").lateral = true;
    node
}

/// lateral_subquery_relation adds a lateral subquery to a range table, with the expressions of the left rows' columns
/// that it reads, renumbered in the order of those columns, as Postgres' parser keeps a LATERAL subquery's references
/// to the query around it as Vars of level 1, and returns its reference.
pub(super) fn lateral_subquery_relation(
    glob: &mut PlannerGlobal,
    ctx: &mut Ctx<'_>,
    mut right: Plan,
    left_columns: &[Expr],
    rtable: &mut Vec<RangeTblEntry>,
    output: &mut Vec<Expr>,
) -> JoinTreeNode {
    let mut read = std::collections::BTreeSet::new();
    crate::indexscan::outer_reads(&right, 1, &mut read);
    let read: Vec<usize> = read.into_iter().collect();
    right.map_exprs(0, &mut |e, depth| renumber_left_reads(e, depth, &read));
    let node = subquery_relation(glob, ctx, right, rtable, output);
    let rte = rtable.last_mut().expect("the lateral subquery");
    rte.lateral = true;
    rte.lateral_columns = read.iter().map(|&c| left_columns[c].clone()).collect();
    node
}

/// renumber_left_reads rewrites the reads of the left row in an expression `nesting` rows out from a lateral
/// subquery's own rows, and in its subqueries, to read the position of their column among those that the subquery
/// reads.
fn renumber_left_reads(e: Expr, nesting: usize, read: &[usize]) -> Expr {
    let mut e = match e {
        Expr::Outer(d, c) if d == nesting + 1 => {
            return Expr::Outer(d, read.binary_search(&c).expect("a column that the subquery reads"));
        }
        other => other.map_children(&mut |x| renumber_left_reads(x, nesting, read)),
    };
    for plan in e.subqueries_mut() {
        plan.map_exprs(0, &mut |x, depth| renumber_left_reads(x, nesting + 1 + depth, read));
    }
    e
}

/// read_left_vars rewrites an expression of a lateral input to read the Vars of the left rows' columns in place of
/// its reads of the left row, and the rows of the enclosing queries one level closer. A SubPlan whose subquery reads
/// the left row passes those Vars as arguments of its own, as Postgres' SubPlans pass upper-level Vars as Params.
pub(super) fn read_left_vars(e: Expr, left_columns: &[Expr]) -> Expr {
    let mut e = match e {
        Expr::Outer(1, c) => return left_columns[c].clone(),
        Expr::Outer(d, c) if d > 1 => return Expr::Outer(d - 1, c),
        Expr::SubPlan(mut subplan) => {
            subplan.map_children(&mut |x| read_left_vars(x, left_columns));
            let (base, mut read) = (subplan.args.len(), Vec::new());
            for plan in subplan.link.subqueries_mut() {
                plan.map_exprs(0, &mut |x, depth| pass_left_reads(x, depth, base, &mut read));
            }
            subplan.init_plan &= read.is_empty();
            subplan.args.extend(read.iter().map(|&c| left_columns[c].clone()));
            return Expr::SubPlan(subplan);
        }
        other => other.map_children(&mut |x| read_left_vars(x, left_columns)),
    };
    for plan in e.subqueries_mut() {
        plan.map_exprs(0, &mut |x, depth| subselect::decrement_sublevels_up(x, depth + 1));
    }
    e
}

/// pass_left_reads rewrites an expression `nesting` rows out from a SubPlan's subquery rows to read the columns of
/// the left row as SubPlan arguments past its first `base`, numbered by their position in `read`, and the rows
/// further out one level closer.
fn pass_left_reads(e: Expr, nesting: usize, base: usize, read: &mut Vec<usize>) -> Expr {
    let mut e = match e {
        Expr::Outer(d, c) if d == nesting + 2 => {
            let k = match read.iter().position(|&r| r == c) {
                Some(k) => k,
                None => {
                    read.push(c);
                    read.len() - 1
                }
            };
            return Expr::Outer(nesting + 1, base + k);
        }
        Expr::Outer(d, c) if d > nesting + 2 => return Expr::Outer(d - 1, c),
        other => other.map_children(&mut |x| pass_left_reads(x, nesting, base, read)),
    };
    for plan in e.subqueries_mut() {
        plan.map_exprs(0, &mut |x, depth| pass_left_reads(x, nesting + 1 + depth, base, read));
    }
    e
}

/// query_planner builds the base relations, distributes the join tree's clauses, sets the orders that the upper
/// processing asks for by its callback, and finds the cheapest paths of the join of every relation, starting over
/// whenever it removes a join, and returns that relation, as Postgres' function of the same name does. A join tree of
/// one RESULT relation is that relation's one row.
fn query_planner(root: &mut PlannerInfo<'_, '_>, qp_callback: &mut dyn FnMut(&mut PlannerInfo<'_, '_>)) -> usize {
    loop {
        root.rels.clear();
        root.join_rel_hash.clear();
        root.join_rel_level.clear();
        root.join_cur_level = 0;
        root.all_baserels = Relids::new();
        root.outer_join_rels = Relids::new();
        root.all_query_rels = Relids::new();
        root.eq_classes.clear();
        root.eq_members.clear();
        root.ec_merging_done = false;
        root.canon_pathkeys.clear();
        root.left_join_clauses.clear();
        root.right_join_clauses.clear();
        root.full_join_clauses.clear();
        root.sjinfos.clear();
        root.join_info_list.clear();
        root.last_rinfo_serial = 0;
        root.rinfos.clear();
        root.placeholder_list.clear();
        root.placeholder_array.clear();
        root.placeholders_frozen = false;
        root.initial_rels.clear();
        root.has_pseudo_constant_quals = false;
        root.fkey_list.clear();
        root.join_domains.truncate(1);
        relnode::setup_simple_rel_arrays(root);
        if let [JoinTreeNode::Rel(varno)] = root.parse.jointree.fromlist.as_slice()
            && matches!(root.parse.rte(*varno).kind, RteKind::Result)
        {
            let varno = *varno;
            relnode::build_simple_rel(root, varno, None);
            let quals = root.parse.jointree.quals.clone();
            let rel = &root.rels[varno];
            let path = nodes::Path {
                kind: nodes::PathKind::Result(quals),
                parent: varno,
                relids: rel.relids.clone(),
                param: Relids::new(),
                pathkeys: Vec::new(),
                rows: 1.0,
                width: 0.0,
                disabled_nodes: 0,
                startup_cost: 0.0,
                total_cost: 0.01,
                pathtarget: None,
            };
            pathnode::add_path(&mut root.rels[varno], Rc::new(path));
            pathnode::set_cheapest(&mut root.rels[varno]);
            root.ec_merging_done = true;
            qp_callback(root);
            return varno;
        }
        let jointree = JoinTreeNode::From(Box::new(root.parse.jointree.clone()));
        initsplan::add_base_rels_to_query(root, &jointree);
        initsplan::remove_useless_groupby_columns(root);
        let mut final_tlist: Vec<Expr> = root.parse.upper_exprs();
        final_tlist.extend(root.processed_tlist.iter().skip(root.parse.target_list.len()).map(|tle| tle.expr.clone()));
        initsplan::build_base_rel_tlists(root, &final_tlist);
        placeholder::find_placeholders_in_jointree(root);
        initsplan::find_lateral_references(root);
        let joinlist = initsplan::deconstruct_jointree(root);
        equivclass::reconsider_outer_join_clauses(root);
        equivclass::generate_base_implied_equalities(root);
        qp_callback(root);
        placeholder::fix_placeholder_input_needed_levels(root);
        if analyzejoins::remove_useless_outer_joins(root)
            || analyzejoins::reduce_unique_semijoins(root)
            || analyzejoins::remove_useless_self_joins(root, &joinlist)
        {
            continue;
        }
        placeholder::add_placeholders_to_base_rels(root);
        initsplan::create_lateral_join_info(root);
        initsplan::match_foreign_keys_to_quals(root);
        orclauses::extract_restriction_or_clauses(root);
        initsplan::add_other_rels_to_query(root);
        return allpaths::make_one_rel(root, &joinlist);
    }
}

/// build_jointree turns a FROM clause's plan into a join tree, as Postgres' parser builds one, adding its inputs to a
/// range table and the expression of each column of the plan's rows to `output`: a decomposable join becomes a JOIN
/// with a range table index of its own, whose nullable sides' columns are Vars that the join can make NULL, a
/// SELECT's projection, LIMIT, DISTINCT, or ORDER BY, or a set operation, becomes a subquery, a filter becomes a FROM list with quals, the
/// one row of an empty FROM clause becomes a RESULT relation, a scan of a table becomes a relation, a WITH query's
/// reference becomes what ss_process_cte decides, a SQL set-returning function that inline_set_returning_function
/// inlines becomes a subquery of its body, and any other input becomes a relation of its own plan, planned already, as
/// is any join that is lateral or already planned. The subqueries of a filter's or a join's conditions
/// pull up as Postgres' pull_up_sublinks_jointree_recurse pulls them up.
pub(super) fn build_jointree(
    glob: &mut PlannerGlobal,
    ctx: &mut Ctx<'_>,
    plan: Plan,
    rtable: &mut Vec<RangeTblEntry>,
    output: &mut Vec<Expr>,
) -> JoinTreeNode {
    let plan = match plan {
        Plan::Join { left, right, kind, condition, lateral: true, method: JoinMethod::Unplanned } => {
            match unlateral(&right) {
                Some(right) => {
                    let method = JoinMethod::Unplanned;
                    Plan::Join { left, right: Box::new(right), kind, condition, lateral: false, method }
                }
                None => Plan::Join { left, right, kind, condition, lateral: true, method: JoinMethod::Unplanned },
            }
        }
        other => other,
    };
    match plan {
        Plan::Join { left, right, kind, condition, lateral, .. } if decomposable(&plan) && !kind.tests_matches() => {
            let jointype = match kind {
                JoinKind::Inner => JoinType::Inner,
                JoinKind::Left => JoinType::Left,
                JoinKind::Right => JoinType::Right,
                JoinKind::Full => JoinType::Full,
                JoinKind::Semi | JoinKind::Anti | JoinKind::RightSemi | JoinKind::RightAnti => {
                    unreachable!("semi and anti joins stay planned")
                }
            };
            rtable.push(RangeTblEntry {
                kind: RteKind::Join(jointype),
                coltypes: Vec::new(),
                lateral: false,
                lateral_columns: Vec::new(),
                inh: false,
            });
            let rtindex = rtable.len();
            let (mut left_columns, mut right_columns) = (Vec::new(), Vec::new());
            let mut larg = build_jointree(glob, ctx, *left, rtable, &mut left_columns);
            let mut rarg = match lateral {
                true if matches!(*right, Plan::Function { .. } | Plan::RowsFrom { .. } | Plan::Values(_)) => {
                    lateral_relation(glob, *right, &left_columns, rtable, &mut right_columns)
                }
                true => lateral_subquery_relation(glob, ctx, *right, &left_columns, rtable, &mut right_columns),
                false => build_jointree(glob, ctx, *right, rtable, &mut right_columns),
            };
            let both = [left_columns.as_slice(), right_columns.as_slice()].concat();
            let conjuncts = condition.as_ref().map(crate::indexscan::conjuncts).unwrap_or_default();
            let conjuncts: Vec<Expr> = conjuncts.into_iter().cloned().collect();
            let leftrelids = prepjointree::get_relids_in_jointree(&larg, true, true);
            let rightrelids = prepjointree::get_relids_in_jointree(&rarg, true, true);
            let mut pull_up = |jtlink: &mut JoinTreeNode, available_rels: &Relids, conjuncts| {
                let none = Relids::new();
                let quals = subselect::pull_up_sublinks_qual_recurse(
                    glob,
                    ctx,
                    conjuncts,
                    jtlink,
                    available_rels,
                    None,
                    &none,
                    rtable,
                    &both,
                );
                quals.into_iter().map(|c| to_vars(subselect::process_sublinks(ctx, c, both.len()), &both)).collect()
            };
            let join =
                |larg, rarg, quals| JoinTreeNode::Join(Box::new(JoinExpr { jointype, larg, rarg, quals, rtindex }));
            let node = match jointype {
                JoinType::Inner => {
                    let mut node = join(larg, rarg, Vec::new());
                    let quals = pull_up(&mut node, &leftrelids.union(&rightrelids), conjuncts);
                    set_sublink_base_quals(&mut node, quals);
                    node
                }
                JoinType::Left => {
                    let quals = pull_up(&mut rarg, &rightrelids, conjuncts);
                    join(larg, rarg, quals)
                }
                JoinType::Right => {
                    let quals = pull_up(&mut larg, &leftrelids, conjuncts);
                    join(larg, rarg, quals)
                }
                _ => join(larg, rarg, conjuncts.into_iter().map(|c| to_vars(c, &both)).collect()),
            };
            let nulled = Relids::singleton(rtindex);
            let (left_nulled, right_nulled) = match jointype {
                JoinType::Inner => (false, false),
                JoinType::Left => (false, true),
                JoinType::Right => (true, false),
                _ => (true, true),
            };
            for (columns, is_nulled) in [(left_columns, left_nulled), (right_columns, right_nulled)] {
                output.extend(columns.into_iter().map(|e| match is_nulled {
                    true => var::add_nulling_relids(glob, e, None, &nulled),
                    false => e,
                }));
            }
            node
        }
        Plan::Project { .. }
        | Plan::Limit { .. }
        | Plan::Distinct { .. }
        | Plan::SetOp { .. }
        | Plan::Recursive { .. } => subquery_relation(glob, ctx, plan, rtable, output),
        Plan::Sort { ref input, .. } if !matches!(**input, Plan::Window { .. }) => {
            subquery_relation(glob, ctx, plan, rtable, output)
        }
        Plan::Filter { input, predicate } => {
            let mut columns = Vec::new();
            let node = build_jointree(glob, ctx, *input, rtable, &mut columns);
            let frelids = prepjointree::get_relids_in_jointree(&node, true, true);
            let mut jtlink = JoinTreeNode::From(Box::new(FromExpr { fromlist: vec![node], quals: Vec::new() }));
            let conjuncts = crate::indexscan::conjuncts(&predicate).into_iter().cloned().collect();
            let none = Relids::new();
            let conjuncts = subselect::pull_up_sublinks_qual_recurse(
                glob,
                ctx,
                conjuncts,
                &mut jtlink,
                &frelids,
                None,
                &none,
                rtable,
                &columns,
            );
            let width = columns.len();
            let quals =
                conjuncts.into_iter().map(|c| to_vars(subselect::process_sublinks(ctx, c, width), &columns)).collect();
            set_sublink_base_quals(&mut jtlink, quals);
            output.extend(columns);
            jtlink
        }
        Plan::OneRow => push_relation(glob, rtable, output, RteKind::Result, Vec::new()),
        Plan::CteScan(def) => subselect::ss_process_cte(glob, ctx, def, rtable, output),
        Plan::Scan(table, None) => {
            let coltypes = table.columns.iter().map(|c| Some(c.ty.oid)).collect();
            let shared = Rc::new((*table).clone());
            push_relation(glob, rtable, output, RteKind::Relation(Plan::Scan(table, None), shared), coltypes)
        }
        other => {
            if let Plan::Function { call: Expr::Routine(routine, args), ordinality: false, defined: None, .. } = &other
                && let Some(body) = clauses::inline_set_returning_function(glob, ctx, routine, args)
            {
                return subquery_relation(glob, ctx, body, rtable, output);
            }
            let coltypes = plan_coltypes(&other);
            push_relation(glob, rtable, output, RteKind::Plan(plan_subquery(ctx, other)), coltypes)
        }
    }
}

/// subquery_relation adds a subquery of the plan of a SELECT to a range table, adding a Var of each of its visible
/// columns to `output`, and returns its reference.
pub(super) fn subquery_relation(
    glob: &mut PlannerGlobal,
    ctx: &mut Ctx<'_>,
    plan: Plan,
    rtable: &mut Vec<RangeTblEntry>,
    output: &mut Vec<Expr>,
) -> JoinTreeNode {
    let original = plan.clone();
    let subquery = query::unbind(glob, ctx, plan);
    let visible = subquery.target_list.iter().filter(|tle| !tle.resjunk);
    let coltypes = visible.map(|tle| nodefuncs::query_expr_type(glob, &subquery, &tle.expr)).collect();
    push_relation(glob, rtable, output, RteKind::Subquery(Box::new(subquery), original), coltypes)
}

/// push_relation adds a relation to a range table, adding a Var of each of its columns to `output`, and returns its
/// reference.
pub(super) fn push_relation(
    glob: &mut PlannerGlobal,
    rtable: &mut Vec<RangeTblEntry>,
    output: &mut Vec<Expr>,
    kind: RteKind,
    coltypes: Vec<Option<u32>>,
) -> JoinTreeNode {
    let varno = rtable.len() + 1;
    output.extend((0..coltypes.len()).map(|attno| glob.var(varno, attno, Relids::new())));
    rtable.push(RangeTblEntry { kind, coltypes, lateral: false, lateral_columns: Vec::new(), inh: false });
    JoinTreeNode::Rel(varno)
}

/// plan_coltypes returns the types of the columns of a plan's rows that the planner knows: a VALUES list's constants,
/// and casts, a catalog's columns, and the declared columns of a function.
fn plan_coltypes(plan: &Plan) -> Vec<Option<u32>> {
    let width = plan.width();
    match plan {
        Plan::Values(rows) if !rows.is_empty() => rows[0]
            .iter()
            .map(|e| match e {
                Expr::Const(value) => nodefuncs::value_type(value),
                Expr::Cast(_, ty, _) => Some(ty.oid),
                _ => None,
            })
            .collect(),
        Plan::Catalog(table) => table.columns.iter().map(|c| Some(c.type_oid)).collect(),
        Plan::Function { defined: Some(columns), ordinality, .. } => {
            let mut types: Vec<Option<u32>> = columns.iter().map(|(_, ty)| Some(ty.oid)).collect();
            if *ordinality {
                types.push(Some(20));
            }
            types
        }
        _ => vec![None; width],
    }
}

/// set_sublink_base_quals sets the quals of the join tree node under the joins that pulled-up subqueries made above
/// it, as Postgres' pull_up_sublinks_jointree_recurse leaves a node's remaining quals on it.
fn set_sublink_base_quals(node: &mut JoinTreeNode, quals: Vec<Expr>) {
    match node {
        JoinTreeNode::Join(j) if j.rtindex == 0 => set_sublink_base_quals(&mut j.larg, quals),
        JoinTreeNode::Join(j) => j.quals = quals,
        JoinTreeNode::From(f) => f.quals = quals,
        JoinTreeNode::Rel(_) => unreachable!("subqueries pull up above a FROM list or a join"),
    }
}

/// to_vars rewrites an expression over a row of columns into one over their expressions.
fn to_vars(e: Expr, columns: &[Expr]) -> Expr {
    match e {
        Expr::Column(c) => columns[c].clone(),
        other => other.map_children(&mut |c| to_vars(c, columns)),
    }
}
