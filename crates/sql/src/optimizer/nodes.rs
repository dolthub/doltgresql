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

//! The planner's data structures, from Postgres' nodes/pathnodes.h and the parts of nodes/parsenodes.h and
//! primnodes.h that the planner reads: the query's range table and join tree, and the planner's relations, paths,
//! and restriction clauses. Expressions are Doltgres' bound expressions, where a Var is a column numbered by `var`.

use std::rc::Rc;

use crate::catalog::table::TableDef;
use crate::expr::Expr;
use crate::plan::Plan;

/// Relids is a set of range table indexes, as Postgres' Relids bitmapset is, for queries of up to 63 relations.
pub type Relids = u64;

/// VAR_SHIFT is how far `var` shifts a range table index above an attribute number.
const VAR_SHIFT: usize = 16;

/// var returns the column number that stands for a Var of a range table entry's attribute, numbered from 0.
pub const fn var(varno: usize, attno: usize) -> usize {
    varno << VAR_SHIFT | attno
}

/// var_parts returns the range table index and attribute number of a Var's column number.
pub const fn var_parts(column: usize) -> (usize, usize) {
    (column >> VAR_SHIFT, column & ((1 << VAR_SHIFT) - 1))
}

/// singleton returns the set of one range table index.
pub const fn singleton(varno: usize) -> Relids {
    1 << varno
}

/// is_subset reports whether every member of one set is a member of the other.
pub const fn is_subset(a: Relids, b: Relids) -> bool {
    a & !b == 0
}

/// overlap reports whether two sets share a member.
pub const fn overlap(a: Relids, b: Relids) -> bool {
    a & b != 0
}

/// members returns the members of a set in order.
pub fn members(set: Relids) -> impl Iterator<Item = usize> {
    (0..64).filter(move |&i| set & (1 << i) != 0)
}

/// JoinType is a join's kind, where a right join keeps the rows of the second input, which the join tree holds only
/// until Postgres' reduce_outer_joins turns it into a left join, and which a join path uses to hash the rows that it
/// keeps.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum JoinType {
    Inner,
    Left,
    Right,
    Full,
    Semi,
    Anti,
}

impl JoinType {
    /// is_outer reports whether the join keeps rows that find no match, as Postgres' IS_OUTER_JOIN does.
    pub fn is_outer(self) -> bool {
        matches!(self, JoinType::Left | JoinType::Right | JoinType::Full | JoinType::Anti)
    }
}

/// RangeTblEntry is a relation that the query reads, as the plan of its rows: a scan of a table, or any other input,
/// already planned, which the planner treats as an opaque base relation, as Postgres treats a function or VALUES list.
#[derive(Clone, Debug)]
pub struct RangeTblEntry {
    pub plan: Plan,
}

impl RangeTblEntry {
    /// table returns the table that a relation entry scans, or None for an opaque entry.
    pub fn table(&self) -> Option<&TableDef> {
        match &self.plan {
            Plan::Scan(table, _) => Some(table),
            _ => None,
        }
    }
}

/// JoinTreeNode is a node of the query's join tree: a reference to a range table entry, a JOIN, or a FROM list.
#[derive(Clone, Debug)]
pub enum JoinTreeNode {
    Rel(usize),
    Join(Box<JoinExpr>),
    From(Box<FromExpr>),
}

/// JoinExpr is a JOIN of two join tree nodes with its ON conditions.
#[derive(Clone, Debug)]
pub struct JoinExpr {
    pub jointype: JoinType,
    pub larg: JoinTreeNode,
    pub rarg: JoinTreeNode,
    pub quals: Vec<Expr>,
}

/// FromExpr is a FROM list, which joins its members by its WHERE conditions.
#[derive(Clone, Debug)]
pub struct FromExpr {
    pub fromlist: Vec<JoinTreeNode>,
    pub quals: Vec<Expr>,
}

/// Query is the part of a query that the planner's query_planner plans: its range table, indexed from 1, its join
/// tree, and the expression of each column of the rows that the plan above the join tree reads, or None for a column
/// that it does not read.
#[derive(Clone, Debug)]
pub struct Query {
    pub rtable: Vec<RangeTblEntry>,
    pub jointree: FromExpr,
    pub output: Vec<Option<Expr>>,
}

impl Query {
    /// rte returns the range table entry at an index.
    pub fn rte(&self, varno: usize) -> &RangeTblEntry {
        &self.rtable[varno - 1]
    }
}

/// RestrictInfo is a WHERE or JOIN/ON clause with the relations it reads and the facts that decide where it is
/// evaluated, as Postgres' RestrictInfo holds them.
#[derive(Clone, Debug)]
pub struct RestrictInfo {
    pub clause: Expr,
    /// Whether the clause can be applied at a level other than its syntactic one: true for WHERE and inner join
    /// clauses and degenerate outer join clauses, false for an outer join's own ON clauses.
    pub is_pushed_down: bool,
    /// Whether the clause reads no relation and has no volatile functions, so it gates a plan once.
    pub pseudoconstant: bool,
    pub clause_relids: Relids,
    /// The relations that must be joined before the clause can be evaluated.
    pub required_relids: Relids,
    /// The relations of each side of a binary operator clause, for join clauses.
    pub left_relids: Relids,
    pub right_relids: Relids,
    /// Whether the clause is an operator over two sides that read disjoint, nonempty sets of relations.
    pub can_join: bool,
    /// Whether the clause is an equality that a hash join can use.
    pub hashjoinable: bool,
    /// The share of rows that the clause keeps, once estimated for an inner join or a restriction, and for an outer
    /// join, or -1 before then, as Postgres caches them.
    pub norm_selec: std::cell::Cell<f64>,
    pub outer_selec: std::cell::Cell<f64>,
}

/// SpecialJoinInfo describes an outer, semi, or anti join, which restricts the orders in which the planner can join
/// relations, as Postgres' SpecialJoinInfo does.
#[derive(Clone, Debug)]
pub struct SpecialJoinInfo {
    pub min_lefthand: Relids,
    pub min_righthand: Relids,
    pub syn_lefthand: Relids,
    pub syn_righthand: Relids,
    pub jointype: JoinType,
    /// Whether the join clause is strict for some relation of the left side.
    pub lhs_strict: bool,
    /// Whether the join cannot commute with an upper outer join's right side.
    pub delay_upper_joins: bool,
}

/// PathKind is how a path produces its relation's rows.
#[derive(Clone, Debug)]
pub enum PathKind {
    /// A sequential scan of a base relation, or the opaque plan of a non-table entry.
    SeqScan,
    /// A scan of an index of a base relation, through Doltgres' index scan, with whether its ranges answer every
    /// restriction it was chosen for.
    IndexScan(Box<crate::indexscan::IndexScan>, bool),
    /// A base relation's rows that one index lookup finds for each row of the relations it is parameterized by, the
    /// inner side of a nested loop, whose lookup keys read those relations' Vars.
    Lookup(crate::plan::JoinMethod),
    /// The rows of another path of the same relation, kept in memory so that a nested loop reads them again cheaply.
    Material(Rc<Path>),
    /// The rows of another path of the same relation sorted in the query's ORDER BY order, which the sort that the
    /// query already holds above the join's rows does.
    Sort(Rc<Path>),
    /// A nested loop of an outer path over an inner one.
    NestLoop(JoinPath),
    /// A hash join probing a hash table of the inner path with the outer path's rows.
    HashJoin(JoinPath),
}

/// JoinPath is the inputs and clauses of a join path.
#[derive(Clone, Debug)]
pub struct JoinPath {
    pub jointype: JoinType,
    pub outer: Rc<Path>,
    pub inner: Rc<Path>,
    /// The clauses the join evaluates.
    pub joinrestrictinfo: Vec<Rc<RestrictInfo>>,
}

/// Path is a way to produce a relation's rows, with its estimated row count and costs.
#[derive(Clone, Debug)]
pub struct Path {
    pub kind: PathKind,
    /// The relations whose rows the path produces.
    pub relids: Relids,
    /// The relations whose current row the path reads, which a nested loop must supply as its outer side.
    pub param: Relids,
    /// The order of the path's rows, as the keys of the query's ORDER BY over Vars, as Postgres' pathkeys, kept only
    /// when the path gives the query's whole order.
    pub pathkeys: Vec<crate::plan::SortKey>,
    pub rows: f64,
    /// The estimated average width of a row, in bytes.
    pub width: f64,
    pub startup_cost: f64,
    pub total_cost: f64,
}

/// RelOptInfo is a base or join relation: its rows and width estimates, its paths, and the clauses that apply to
/// it, as Postgres' RelOptInfo holds them.
#[derive(Clone, Debug, Default)]
pub struct RelOptInfo {
    pub relids: Relids,
    /// The estimated number of rows after the relation's restrictions.
    pub rows: f64,
    /// The estimated average width of a row, in bytes.
    pub width: f64,
    pub pathlist: Vec<Rc<Path>>,
    /// Whether paths that start more cheaply are worth keeping, as they are when a LIMIT asks for the first rows.
    pub consider_startup: bool,
    pub cheapest_total_path: Option<Rc<Path>>,
    pub cheapest_startup_path: Option<Rc<Path>>,
    /// The cheapest unparameterized path and the cheapest path of each parameterization.
    pub cheapest_parameterized_paths: Vec<Rc<Path>>,
    /// For a base relation: its range table index, its size in pages and rows before restrictions.
    pub relid: usize,
    pub pages: f64,
    pub tuples: f64,
    /// The statistics of a base relation's table.
    pub stats: Option<std::sync::Arc<crate::colstats::TableStats>>,
    /// For a base relation, the relations of the joins above it that read each of its columns, where relation 0 is
    /// the query's output, as Postgres' attr_needed holds them.
    pub attr_needed: Vec<Relids>,
    /// The restriction clauses of a base relation.
    pub baserestrictinfo: Vec<Rc<RestrictInfo>>,
    /// The join clauses that read this relation and others.
    pub joininfo: Vec<Rc<RestrictInfo>>,
}
