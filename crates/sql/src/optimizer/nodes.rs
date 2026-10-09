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

//! The planner's data structures, from Postgres' nodes/pathnodes.h, nodes/bitmapset.c, and the parts of
//! nodes/parsenodes.h and primnodes.h that the planner reads: the query's range table and join tree, its Vars and
//! PlaceHolderVars, and the planner's relations, paths, restriction clauses, and equivalence classes. Expressions are
//! Doltgres' bound expressions, where `Expr::Column` holds the ID of a Var or PlaceHolderVar in the statement's
//! `PlannerGlobal`.

use std::cell::Cell;
use std::collections::HashMap;
use std::rc::Rc;

use smallvec::SmallVec;

use crate::catalog::table::TableDef;
use crate::expr::Expr;
use crate::plan::Plan;

/// Relids is a set of range table indexes, as Postgres' Relids bitmapset is.
#[derive(Clone, Default, PartialEq, Eq, Hash, Debug)]
pub struct Relids(SmallVec<[u64; 1]>);

/// SubsetCompare is how two sets compare by inclusion, as Postgres' BMS_Comparison is.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum SubsetCompare {
    Equal,
    Subset1,
    Subset2,
    Different,
}

impl Relids {
    /// new returns the empty set.
    pub fn new() -> Relids {
        Relids(SmallVec::new())
    }

    /// singleton returns the set of one member, as bms_make_singleton does.
    pub fn singleton(x: usize) -> Relids {
        let mut set = Relids::new();
        set.add_member(x);
        set
    }

    /// trim drops the empty words at the end, so that equal sets compare and hash equally.
    fn trim(&mut self) {
        while self.0.last() == Some(&0) {
            self.0.pop();
        }
    }

    /// is_empty reports whether the set has no members.
    pub fn is_empty(&self) -> bool {
        self.0.is_empty()
    }

    /// is_member reports whether a value is a member.
    pub fn is_member(&self, x: usize) -> bool {
        self.0.get(x / 64).is_some_and(|w| w & (1 << (x % 64)) != 0)
    }

    /// add_member adds a value.
    pub fn add_member(&mut self, x: usize) {
        if self.0.len() <= x / 64 {
            self.0.resize(x / 64 + 1, 0);
        }
        self.0[x / 64] |= 1 << (x % 64);
    }

    /// with_member returns the set with a value added.
    pub fn with_member(mut self, x: usize) -> Relids {
        self.add_member(x);
        self
    }

    /// del_member removes a value.
    pub fn del_member(&mut self, x: usize) {
        if let Some(w) = self.0.get_mut(x / 64) {
            *w &= !(1 << (x % 64));
            self.trim();
        }
    }

    /// without_member returns the set with a value removed.
    pub fn without_member(mut self, x: usize) -> Relids {
        self.del_member(x);
        self
    }

    /// add_members adds every member of another set, as bms_add_members does.
    pub fn add_members(&mut self, other: &Relids) {
        if self.0.len() < other.0.len() {
            self.0.resize(other.0.len(), 0);
        }
        for (w, o) in self.0.iter_mut().zip(&other.0) {
            *w |= o;
        }
    }

    /// del_members removes every member of another set, as bms_del_members does.
    pub fn del_members(&mut self, other: &Relids) {
        for (w, o) in self.0.iter_mut().zip(&other.0) {
            *w &= !o;
        }
        self.trim();
    }

    /// int_members keeps only the members of another set, as bms_int_members does.
    pub fn int_members(&mut self, other: &Relids) {
        self.0.truncate(other.0.len());
        for (w, o) in self.0.iter_mut().zip(&other.0) {
            *w &= o;
        }
        self.trim();
    }

    /// union returns the members of either set.
    pub fn union(&self, other: &Relids) -> Relids {
        let mut set = self.clone();
        set.add_members(other);
        set
    }

    /// intersect returns the members of both sets.
    pub fn intersect(&self, other: &Relids) -> Relids {
        let mut set = self.clone();
        set.int_members(other);
        set
    }

    /// difference returns the members of this set that the other lacks.
    pub fn difference(&self, other: &Relids) -> Relids {
        let mut set = self.clone();
        set.del_members(other);
        set
    }

    /// is_subset reports whether every member of this set is a member of the other.
    pub fn is_subset(&self, other: &Relids) -> bool {
        self.0.iter().enumerate().all(|(i, w)| w & !other.0.get(i).copied().unwrap_or(0) == 0)
    }

    /// overlap reports whether the sets share a member.
    pub fn overlap(&self, other: &Relids) -> bool {
        self.0.iter().zip(&other.0).any(|(w, o)| w & o != 0)
    }

    /// nonempty_difference reports whether this set has a member that the other lacks.
    pub fn nonempty_difference(&self, other: &Relids) -> bool {
        !self.is_subset(other)
    }

    /// num_members counts the members.
    pub fn num_members(&self) -> usize {
        self.0.iter().map(|w| w.count_ones() as usize).sum()
    }

    /// members returns the members in increasing order.
    pub fn members(&self) -> impl Iterator<Item = usize> + '_ {
        self.0
            .iter()
            .enumerate()
            .flat_map(|(i, &w)| (0..64).filter(move |b| w & (1 << b) != 0).map(move |b| i * 64 + b))
    }

    /// singleton_member returns the only member of a set of one, as bms_get_singleton_member does.
    pub fn singleton_member(&self) -> Option<usize> {
        let mut members = self.members();
        let first = members.next()?;
        members.next().is_none().then_some(first)
    }

    /// subset_compare compares two sets by inclusion, as bms_subset_compare does.
    pub fn subset_compare(&self, other: &Relids) -> SubsetCompare {
        match (self.is_subset(other), other.is_subset(self)) {
            (true, true) => SubsetCompare::Equal,
            (true, false) => SubsetCompare::Subset1,
            (false, true) => SubsetCompare::Subset2,
            (false, false) => SubsetCompare::Different,
        }
    }
}

impl FromIterator<usize> for Relids {
    fn from_iter<I: IntoIterator<Item = usize>>(iter: I) -> Relids {
        let mut set = Relids::new();
        for x in iter {
            set.add_member(x);
        }
        set
    }
}

/// Var is a column of a range table entry, with the outer joins that can make it NULL, as Postgres' Var with its
/// varnullingrels is. Attributes are numbered from 0.
#[derive(Clone, Debug, PartialEq, Eq, Hash)]
pub struct Var {
    pub varno: usize,
    pub varattno: usize,
    pub varnullingrels: Relids,
}

/// PlaceHolderVar is an expression evaluated below outer joins that can make it NULL, as Postgres' PlaceHolderVar
/// is. Its expression and relations belong to its ID, in `PlannerGlobal::placeholders`, which every copy shares.
#[derive(Clone, Debug, PartialEq, Eq, Hash)]
pub struct PlaceHolderVar {
    pub phid: usize,
    pub phnullingrels: Relids,
}

/// PlaceHolder is the expression of a PlaceHolderVar and the relations whose rows it is evaluated over.
#[derive(Clone, Debug)]
pub struct PlaceHolder {
    pub phexpr: Expr,
    pub phrels: Relids,
}

/// VarNode is what an `Expr::Column` of the planner stands for.
#[derive(Clone, Debug, PartialEq, Eq, Hash)]
pub enum VarNode {
    Var(Var),
    PlaceHolderVar(PlaceHolderVar),
}

/// PlannerGlobal holds the state of planning a statement that its queries share, as Postgres' PlannerGlobal does:
/// here, the Vars and PlaceHolderVars, which expressions refer to by ID so that equal ones have equal IDs, and the
/// expressions of the PlaceHolderVars by their IDs, from 1.
#[derive(Default)]
pub struct PlannerGlobal {
    nodes: Vec<VarNode>,
    ids: HashMap<VarNode, usize>,
    pub placeholders: Vec<PlaceHolder>,
}

impl PlannerGlobal {
    /// intern returns the expression that refers to a Var or PlaceHolderVar.
    pub fn intern(&mut self, node: VarNode) -> Expr {
        if let Some(&id) = self.ids.get(&node) {
            return Expr::Column(id);
        }
        let id = self.nodes.len();
        self.nodes.push(node.clone());
        self.ids.insert(node, id);
        Expr::Column(id)
    }

    /// var returns the expression of a Var.
    pub fn var(&mut self, varno: usize, varattno: usize, varnullingrels: Relids) -> Expr {
        self.intern(VarNode::Var(Var { varno, varattno, varnullingrels }))
    }

    /// node returns the Var or PlaceHolderVar of an ID.
    pub fn node(&self, id: usize) -> &VarNode {
        &self.nodes[id]
    }

    /// placeholder returns the expression and relations of a PlaceHolderVar's ID.
    pub fn placeholder(&self, phid: usize) -> &PlaceHolder {
        &self.placeholders[phid - 1]
    }

    /// last_ph_id returns the highest PlaceHolderVar ID, or 0 before any, as Postgres' lastPHId is.
    pub fn last_ph_id(&self) -> usize {
        self.placeholders.len()
    }

    /// make_placeholder_expr returns a new PlaceHolderVar of an expression evaluated over relations, as Postgres'
    /// function of the same name does.
    pub fn make_placeholder_expr(&mut self, phexpr: Expr, phrels: Relids) -> Expr {
        self.placeholders.push(PlaceHolder { phexpr, phrels });
        let phid = self.placeholders.len();
        self.intern(VarNode::PlaceHolderVar(PlaceHolderVar { phid, phnullingrels: Relids::new() }))
    }
}

/// JoinType is a join's kind, as Postgres' JoinType is, where a right join is a left join with its inputs swapped,
/// which the join tree holds only until reduce_outer_joins swaps them, and which a join path uses to hash the rows
/// that it keeps.
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

/// RteKind is what a range table entry reads, as Postgres' RTEKind is.
#[derive(Clone, Debug)]
pub enum RteKind {
    /// A table, with the plan that scans it.
    Relation(Plan, Rc<TableDef>),
    /// A subquery, with the plan that its binding built, which runs when the subquery is not pulled up.
    Subquery(Box<Query>, Plan),
    /// A JOIN, whose index stands for an outer join in the sets of relations that can make a Var NULL.
    Join(JoinType),
    /// One row without columns, as an empty FROM clause reads.
    Result,
    /// Any other input, already planned, which the planner treats as an opaque base relation, as Postgres treats a
    /// function or VALUES list.
    Plan(Plan),
}

/// RangeTblEntry is a relation that the query reads, as Postgres' RangeTblEntry is, with the OID of the type of each
/// of its columns, or None where the planner does not know it.
#[derive(Clone, Debug)]
pub struct RangeTblEntry {
    pub kind: RteKind,
    pub coltypes: Vec<Option<u32>>,
}

impl RangeTblEntry {
    /// table returns the table that a relation entry reads, or None for any other entry.
    pub fn table(&self) -> Option<&TableDef> {
        match &self.kind {
            RteKind::Relation(_, table) => Some(table),
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

/// JoinExpr is a JOIN of two join tree nodes with its ON conditions and its range table index, which is 0 for a semi
/// or anti join that a subquery became, as Postgres' JoinExpr is.
#[derive(Clone, Debug)]
pub struct JoinExpr {
    pub jointype: JoinType,
    pub larg: JoinTreeNode,
    pub rarg: JoinTreeNode,
    pub quals: Vec<Expr>,
    pub rtindex: usize,
}

/// FromExpr is a FROM list, which joins its members by its WHERE conditions.
#[derive(Clone, Debug)]
pub struct FromExpr {
    pub fromlist: Vec<JoinTreeNode>,
    pub quals: Vec<Expr>,
}

/// Query is the part of a query that the planner plans: its range table, indexed from 1, its join tree, and the
/// expression of each column of its rows, or None for a column that nothing reads, as Postgres' Query holds them.
#[derive(Clone, Debug)]
pub struct Query {
    pub rtable: Vec<RangeTblEntry>,
    pub jointree: FromExpr,
    pub target_list: Vec<Option<Expr>>,
}

impl Query {
    /// rte returns the range table entry at an index, as Postgres' rt_fetch does.
    pub fn rte(&self, varno: usize) -> &RangeTblEntry {
        &self.rtable[varno - 1]
    }
}

/// RinfoId is a RestrictInfo's index in `PlannerInfo::rinfos`.
pub type RinfoId = usize;

/// RestrictInfo is a WHERE or JOIN/ON clause with the relations it reads and the facts that decide where it is
/// evaluated, as Postgres' RestrictInfo holds them.
#[derive(Clone, Debug)]
pub struct RestrictInfo {
    pub clause: Expr,
    /// Whether the clause can be applied at a level other than its syntactic one: true for WHERE and inner join
    /// clauses and degenerate outer join clauses, false for an outer join's own ON clauses.
    pub is_pushed_down: bool,
    pub can_join: bool,
    /// Whether the clause reads no relation of the query and runs no volatile function, so it gates a plan once.
    pub pseudoconstant: bool,
    /// Whether the clause has clones for different join orders, or is such a clone.
    pub has_clone: bool,
    pub is_clone: bool,
    pub security_level: usize,
    pub clause_relids: Relids,
    /// The relations that must be joined before the clause can be evaluated.
    pub required_relids: Relids,
    /// The relations of a join where a clone of the clause must not be evaluated.
    pub incompatible_relids: Relids,
    /// The relations of the left side of an outer join whose ON clause this is.
    pub outer_relids: Relids,
    /// The relations of each side of a binary operator clause.
    pub left_relids: Relids,
    pub right_relids: Relids,
    /// The RestrictInfos of the conjuncts of each argument of an OR clause.
    pub orclause: Option<Vec<Vec<RinfoId>>>,
    pub rinfo_serial: usize,
    /// The equivalence class that the clause was derived from.
    pub parent_ec: Option<EcId>,
    /// The share of rows that the clause keeps, once estimated for an inner join or a restriction, and for an outer
    /// join, or -1 before then, as Postgres caches them.
    pub norm_selec: Cell<f64>,
    pub outer_selec: Cell<f64>,
    /// The btree operator families whose equality the clause is, when a merge join can use it.
    pub mergeopfamilies: Vec<u32>,
    /// The equivalence classes and members of the sides of a merge join clause.
    pub left_ec: Option<EcId>,
    pub right_ec: Option<EcId>,
    pub left_em: Option<EmId>,
    pub right_em: Option<EmId>,
    /// Whether the outer side of a join that the clause was last matched to is its left side.
    pub outer_is_left: Cell<bool>,
    /// Whether the clause is an equality that a hash join can use.
    pub hashjoinable: bool,
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
    /// The range table index of an outer join, or 0 for an inner, semi, or anti join.
    pub ojrelid: usize,
    /// The outer joins that this one commutes with, above and below it, on its left and right sides.
    pub commute_above_l: Relids,
    pub commute_above_r: Relids,
    pub commute_below_l: Relids,
    pub commute_below_r: Relids,
    /// Whether the join clause is strict for some relation of the left side.
    pub lhs_strict: bool,
    /// For a semi join, whether its right side can be made unique by sorting or hashing, by the equality operators'
    /// right-hand expressions.
    pub semi_can_btree: bool,
    pub semi_can_hash: bool,
    pub semi_rhs_exprs: Vec<Expr>,
}

/// OuterJoinClauseInfo is a mergejoinable ON clause of an outer join, with the join, as Postgres keeps them for
/// reconsider_outer_join_clauses.
#[derive(Clone, Debug)]
pub struct OuterJoinClauseInfo {
    pub rinfo: RinfoId,
    pub sjinfo: SjId,
}

/// SjId is a SpecialJoinInfo's index in `PlannerInfo::sjinfos`.
pub type SjId = usize;

/// JoinDomain is a set of relations that inner joins join, where equalities hold among them, as Postgres'
/// JoinDomain is.
#[derive(Clone, Debug)]
pub struct JoinDomain {
    pub jd_relids: Relids,
}

/// EcId is an equivalence class's index in `PlannerInfo::eq_classes`.
pub type EcId = usize;

/// EmId is an equivalence member's index in `PlannerInfo::eq_members`.
pub type EmId = usize;

/// EquivalenceClass is a set of expressions known to be equal, as Postgres' EquivalenceClass is.
#[derive(Clone, Debug)]
pub struct EquivalenceClass {
    pub ec_opfamilies: Vec<u32>,
    pub ec_members: Vec<EmId>,
    /// The clauses that the class came from, and those generated from it.
    pub ec_sources: Vec<RinfoId>,
    pub ec_derives: Vec<RinfoId>,
    /// The derived clauses by their members and parent class, as Postgres' ec_derives_hash keys them.
    pub ec_derives_hash: HashMap<(Option<EmId>, EmId, Option<EcId>), RinfoId>,
    pub ec_relids: Relids,
    pub ec_has_const: bool,
    pub ec_has_volatile: bool,
    pub ec_broken: bool,
    pub ec_sortref: usize,
    pub ec_min_security: usize,
    pub ec_max_security: usize,
    /// The class that this one was merged into.
    pub ec_merged: Option<EcId>,
}

/// EquivalenceMember is an expression of an equivalence class, as Postgres' EquivalenceMember is.
#[derive(Clone, Debug)]
pub struct EquivalenceMember {
    pub em_expr: Expr,
    pub em_relids: Relids,
    pub em_is_const: bool,
    pub em_datatype: u32,
    pub em_jdomain: usize,
}

/// PkId is a canonical pathkey's index in `PlannerInfo::canon_pathkeys`.
pub type PkId = usize;

/// PathKey is a key of the order of a path's rows, as Postgres' PathKey is.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PathKey {
    pub pk_eclass: EcId,
    pub pk_opfamily: u32,
    /// Whether the key is descending.
    pub pk_descending: bool,
    pub pk_nulls_first: bool,
}

/// PlaceHolderInfo is the planner's knowledge of a PlaceHolderVar, as Postgres' PlaceHolderInfo is.
#[derive(Clone, Debug)]
pub struct PlaceHolderInfo {
    pub phid: usize,
    /// The relations where the expression is evaluated, those it reads laterally, and those that read it.
    pub ph_eval_at: Relids,
    pub ph_lateral: Relids,
    pub ph_needed: Relids,
    pub ph_width: f64,
}

/// PathTarget is the expressions of a path's rows with their estimated cost and width, as Postgres' PathTarget is.
#[derive(Clone, Debug, Default)]
pub struct PathTarget {
    pub exprs: Vec<Expr>,
    pub cost: super::costsize::QualCost,
    pub width: f64,
}

/// PathKind is how a path produces its relation's rows.
#[derive(Clone, Debug)]
pub enum PathKind {
    /// A sequential scan of a base relation, or the opaque plan of a non-table entry.
    SeqScan,
    /// The one row of a relation without columns, under its quals, as Postgres' GroupResultPath is.
    Result(Vec<Expr>),
    /// The rows of each of the paths in turn, where no paths make an empty relation, as Postgres' AppendPath is.
    Append(Vec<Rc<Path>>),
    /// A scan of an index of a base relation, which a nested loop runs again for each outer row when the path is
    /// parameterized.
    IndexScan(Box<IndexPath>),
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
    pub joinrestrictinfo: Vec<RinfoId>,
}

/// Path is a way to produce a relation's rows, with its estimated row count and costs.
#[derive(Clone, Debug)]
pub struct Path {
    pub kind: PathKind,
    /// The relation whose rows the path produces.
    pub parent: usize,
    pub relids: Relids,
    /// The relations whose current row the path reads, which a nested loop must supply as its outer side.
    pub param: Relids,
    /// The order of the path's rows.
    pub pathkeys: Vec<PkId>,
    pub rows: f64,
    /// The estimated average width of a row, in bytes.
    pub width: f64,
    /// The number of plan nodes that a disabled planner method makes, as Postgres 18 counts them before costs.
    pub disabled_nodes: usize,
    pub startup_cost: f64,
    pub total_cost: f64,
}

/// IndexOptInfo is the planner's knowledge of an index of a table, as Postgres' IndexOptInfo holds it: Dolt's
/// primary index, as None, or a secondary index by position, its size, its columns, as table attributes or None for
/// an expression, where the first `nkeycolumns` are the ones index clauses search and a secondary index then orders
/// its entries by the primary key's columns, the expression of each expression column, each column's btree operator
/// family, direction, and NULL placement, whether its order gives its entries' order (a unique index of content
/// hashes does not), whether it is unique, its predicate's conjuncts and whether the query's clauses imply them, and
/// the restrictions that a scan of it must test.
#[derive(Clone, Debug)]
pub struct IndexOptInfo {
    pub index: Option<usize>,
    pub pages: f64,
    pub tuples: f64,
    pub tree_height: f64,
    pub indexkeys: Vec<Option<usize>>,
    pub nkeycolumns: usize,
    pub indexprs: Vec<Expr>,
    pub opfamily: Vec<Option<u32>>,
    pub reverse_sort: Vec<bool>,
    pub nulls_first: Vec<bool>,
    pub sortable: bool,
    pub unique: bool,
    pub indpred: Vec<Expr>,
    pub pred_ok: bool,
    pub indrestrictinfo: Vec<RinfoId>,
}

/// IndexClause is a clause that an index column can search by, as Postgres' IndexClause is: its RestrictInfo, the
/// clauses the index searches by in its place, with its sides swapped or derived from it, whether those clauses
/// keep more rows than it does, and the column.
#[derive(Clone, Debug)]
pub struct IndexClause {
    pub rinfo: RinfoId,
    pub indexquals: Vec<RinfoId>,
    pub lossy: bool,
    pub indexcol: usize,
}

/// IndexPath is the index scan of a path, as Postgres' IndexPath holds it: the index by its position in the
/// relation's index list, the clauses it searches by, whether it reads backward, whether it reads nothing but the
/// index, and the share of the index's entries that it reads.
#[derive(Clone, Debug)]
pub struct IndexPath {
    pub index: usize,
    pub indexclauses: Vec<IndexClause>,
    pub backward: bool,
    pub indexonly: bool,
    pub indexselectivity: f64,
}

/// UniqueRelInfo records outer relations whose join clauses prove a relation's rows unique, as Postgres'
/// UniqueRelInfo does, with whether the proof was for a self join and the restrictions that it used then.
#[derive(Clone, Debug)]
pub struct UniqueRelInfo {
    pub outerrelids: Relids,
    pub self_join: bool,
    pub extra_clauses: Vec<RinfoId>,
}

/// ParamPathInfo is what a parameterization by outer relations gives a base relation's paths, as Postgres'
/// ParamPathInfo holds it: the outer relations, the rows that the relation's paths return under it, and the join
/// clauses that its paths can test.
#[derive(Clone, Debug)]
pub struct ParamPathInfo {
    pub ppi_req_outer: Relids,
    pub ppi_rows: f64,
    pub ppi_clauses: Vec<RinfoId>,
}

/// RelOptKind is a relation's kind, as Postgres' RelOptKind is, where an unused slot of the simple relations is a
/// range table entry that is not a base relation.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub enum RelOptKind {
    #[default]
    Unused,
    BaseRel,
    JoinRel,
}

/// RelOptInfo is a base or join relation: its rows and width estimates, its paths, and the clauses that apply to
/// it, as Postgres' RelOptInfo holds them.
#[derive(Clone, Debug, Default)]
pub struct RelOptInfo {
    pub reloptkind: RelOptKind,
    pub relids: Relids,
    /// The estimated number of rows after the relation's restrictions.
    pub rows: f64,
    /// Whether paths that start more cheaply are worth keeping, as they are when a LIMIT asks for the first rows.
    pub consider_startup: bool,
    pub consider_param_startup: bool,
    /// The expressions of the relation's rows that joins above and the query's output read.
    pub reltarget: PathTarget,
    pub pathlist: Vec<Rc<Path>>,
    pub cheapest_total_path: Option<Rc<Path>>,
    pub cheapest_startup_path: Option<Rc<Path>>,
    /// The cheapest unparameterized path and the cheapest path of each parameterization.
    pub cheapest_parameterized_paths: Vec<Rc<Path>>,
    /// The relations that the relation's paths must read laterally.
    pub direct_lateral_relids: Relids,
    pub lateral_relids: Relids,
    /// For a base relation: its range table index, and its size in pages and rows before restrictions.
    pub relid: usize,
    pub pages: f64,
    pub tuples: f64,
    /// The statistics of a base relation's table.
    pub stats: Option<std::sync::Arc<crate::colstats::TableStats>>,
    /// For a base relation, the relations that read each of its columns, where relation 0 is the query's output,
    /// as Postgres' attr_needed holds them, and each column's estimated width.
    pub attr_needed: Vec<Relids>,
    pub attr_widths: Vec<f64>,
    /// The columns of a base relation that are never NULL.
    pub notnullattnums: Vec<usize>,
    /// The outer joins that can make a base relation's columns NULL.
    pub nulling_relids: Relids,
    /// The indexes of a base relation's table.
    pub indexlist: Vec<IndexOptInfo>,
    /// The parameterizations of a base relation's paths that were built.
    pub ppilist: Vec<ParamPathInfo>,
    /// The outer relations that make a base relation's rows unique, and those that were found not to.
    pub unique_for_rels: Vec<UniqueRelInfo>,
    pub non_unique_for_rels: Vec<Relids>,
    /// The relations that read a base relation laterally.
    pub lateral_referencers: Relids,
    /// The equivalence classes that mention the relation.
    pub eclass_indexes: Relids,
    /// The restriction clauses of a base relation.
    pub baserestrictinfo: Vec<RinfoId>,
    pub baserestrict_min_security: usize,
    /// The join clauses that read this relation and others.
    pub joininfo: Vec<RinfoId>,
    /// Whether an equivalence class may give the relation join clauses.
    pub has_eclass_joins: bool,
}
