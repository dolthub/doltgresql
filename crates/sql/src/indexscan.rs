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

//! Index scans: choosing the index that answers a table's filters as go-mysql-server's costedIndexScans does without
//! statistics, building the ranges of its keys that the filters allow, and reading the rows of those ranges.

use std::collections::BTreeSet;
use std::sync::Arc;

use doltdb::database::Database;
use prolly::NodeStore;

use crate::catalog::table::{HIDDEN_BASE, TableDef};
use crate::error::Result;
use crate::expr::{CmpOp, Expr};
use crate::plan::Plan;
use crate::query::Ctx;
use crate::ranges::{Cut, IndexBuilder, Range};
use crate::types::Value;

/// IndexScan reads the rows of a table whose keys in an index lie in ranges, in the index's order or its reverse.
#[derive(Clone, Debug, PartialEq)]
pub struct IndexScan {
    pub table: Arc<TableDef>,
    /// The secondary index read, or None for the primary key.
    pub index: Option<usize>,
    pub ranges: Vec<Range>,
    pub reverse: bool,
    /// The vector search that a vector index answers in place of ranges.
    pub nearest: Option<Nearest>,
    /// The table columns that the plan above the scan reads, or None for every column, which lets a scan of a
    /// secondary index that holds them all skip the primary index.
    pub needed: Option<Vec<usize>>,
    /// Some when the scan reads so many of the table's rows that a full scan beats looking them up in the primary
    /// index, which the plan does in its place unless the index holds every column the plan reads. It holds whether
    /// the scan's ranges alone answer the filter above it, which goes when the scan stays.
    pub lookup_heavy: Option<bool>,
    /// The index conditions of a scan that reads its enclosing row, which build its ranges each time it runs, as
    /// Postgres' index scans compute their runtime keys.
    pub parameterized: Option<Expr>,
}

/// Item is a key of an index and its value.
type Item = (Vec<u8>, Vec<u8>);

/// ESTIMATED_RANGES is how many of a scan's ranges `IndexScan::estimate` counts the entries of.
const ESTIMATED_RANGES: usize = 16;

/// ESTIMATED_NEAREST is how many rows `IndexScan::estimate` expects a vector search to find.
const ESTIMATED_NEAREST: f64 = 10.0;

/// Bounds is where the keys of a range lie in an index.
struct Bounds {
    /// The key that the range's keys start at.
    start: Vec<u8>,
    /// The number of fields of the start key: the single values, then the lower bound of the next column if it has one.
    start_width: usize,
    /// Whether the range leaves out the keys equal to its start key, as a lower bound above a value does.
    exclusive: bool,
    /// The number of leading columns that the range holds at a single value, which every key of the range shares.
    points: usize,
    /// The key that the range's keys end at, with whether it is inclusive, when the column after the single values has
    /// an upper bound.
    end: Option<(Vec<u8>, bool)>,
    /// Whether every key between the bounds lies in the range.
    exact: bool,
}

/// Nearest is a search of a vector index for the keys closest to a query vector, as many as a LIMIT and OFFSET keep.
#[derive(Clone, Debug, PartialEq)]
pub struct Nearest {
    /// The distance that orders the rows, which EXPLAIN shows.
    pub order: Expr,
    pub query: Expr,
    pub limit: Option<Expr>,
    pub offset: Option<Expr>,
}

/// Op is what a filter leaf tests of a column.
#[derive(Clone, Copy, Debug, PartialEq)]
enum Op {
    Compare(CmpOp),
    IsNull,
    IsNotNull,
}

/// Leaf is a filter that tests one column against a constant, as go-mysql-server's iScanLeaf.
#[derive(Clone, Debug)]
struct Leaf {
    id: usize,
    column: String,
    op: Op,
    value: Value,
}

/// And is a conjunction of filters, as go-mysql-server's iScanAnd, with its leaves grouped by column in the order
/// they appear.
#[derive(Clone, Debug, Default)]
struct And {
    id: usize,
    leaves: Vec<(String, Vec<Leaf>)>,
    ors: Vec<Or>,
    count: usize,
}

/// Or is a disjunction of filters, as go-mysql-server's iScanOr.
#[derive(Clone, Debug)]
struct Or {
    id: usize,
    children: Vec<Filter>,
}

/// Filter is a node of the filter tree that index costing builds.
#[derive(Clone, Debug)]
enum Filter {
    And(And),
    Or(Or),
    Leaf(Leaf),
}

impl Filter {
    /// id returns the node's ID.
    fn id(&self) -> usize {
        match self {
            Filter::And(a) => a.id,
            Filter::Or(o) => o.id,
            Filter::Leaf(l) => l.id,
        }
    }
}

impl And {
    /// add_leaf adds a leaf under its column.
    fn add_leaf(&mut self, leaf: Leaf) {
        match self.leaves.iter_mut().find(|(c, _)| *c == leaf.column) {
            Some((_, leaves)) => leaves.push(leaf),
            None => self.leaves.push((leaf.column.clone(), vec![leaf])),
        }
        self.count += 1;
    }

    /// sorted_leaves returns the leaves in ID order.
    fn sorted_leaves(&self) -> Vec<&Leaf> {
        let mut leaves: Vec<&Leaf> = self.leaves.iter().flat_map(|(_, l)| l).collect();
        leaves.sort_by_key(|l| l.id);
        leaves
    }
}

/// Fds is the part of go-mysql-server's FuncDepSet that index costing reads: the constant columns and the leading
/// key.
#[derive(Clone, Debug, Default)]
struct Fds {
    constants: BTreeSet<usize>,
    /// The leading key's columns, and whether it is strict.
    key: Option<(BTreeSet<usize>, bool)>,
}

impl Fds {
    /// lookup returns the dependencies of a unique index lookup, as go-mysql-server's NewLookupFDs builds them, where
    /// the constants are the positions of the index columns that equalities fix, counted from 1.
    fn lookup(index_columns: &BTreeSet<usize>, not_null: &BTreeSet<usize>, constants: BTreeSet<usize>) -> Fds {
        let cols: BTreeSet<usize> = index_columns.difference(&constants).copied().collect();
        let strict = index_columns.is_subset(not_null);
        Fds { constants, key: Some((cols, strict)) }
    }

    /// max_one_row reports whether the dependencies allow at most one row.
    fn max_one_row(&self) -> bool {
        matches!(&self.key, Some((cols, true)) if cols.is_empty())
    }

    /// strict_key returns the leading key when it is strict.
    fn strict_key(&self) -> Option<&BTreeSet<usize>> {
        self.key.as_ref().filter(|(_, strict)| *strict).map(|(c, _)| c)
    }

    /// has_lax_key reports whether the leading key is lax.
    fn has_lax_key(&self) -> bool {
        matches!(&self.key, Some((_, false)))
    }
}

/// Candidate is an index that a scan may read, as costing sees it.
struct Candidate {
    /// The index, or None for the primary key.
    index: Option<usize>,
    name: String,
    columns: Vec<String>,
    unique: bool,
    /// The table column IDs of the index's columns, counted from 1.
    column_ids: BTreeSet<usize>,
}

/// Collector gathers the effect of a conjunction's leaves on an index, as go-mysql-server's conjCollector does.
#[derive(Default)]
struct Collector {
    constant: BTreeSet<usize>,
    inequality: BTreeSet<usize>,
    applied: BTreeSet<usize>,
    missing_prefix: usize,
}

impl Collector {
    /// add applies a leaf on one of the index's columns.
    fn add(&mut self, leaf: &Leaf, ordinal: usize) {
        self.applied.insert(leaf.id);
        if leaf.op != Op::Compare(CmpOp::Eq) {
            self.inequality.insert(ordinal);
            return;
        }
        if !self.constant.insert(ordinal + 1) || ordinal != self.missing_prefix {
            return;
        }
        let mut last = ordinal;
        while self.constant.contains(&(last + 1)) {
            last += 1;
        }
        self.missing_prefix = last;
    }
}

/// Cost is what costing finds of an index: the filters it applies, its dependencies, the length of its key prefix that
/// equalities fix, and whether an inequality bounds the next column.
#[derive(Default)]
struct Cost {
    filters: BTreeSet<usize>,
    fds: Fds,
    prefix: usize,
    has_range: bool,
}

/// Best is the index that costing has chosen so far, or the table scan.
struct Best<'c> {
    candidate: Option<&'c Candidate>,
    cost: Cost,
}

/// Coster chooses an index for a table's filters, as go-mysql-server's indexCoster does.
struct Coster<'t> {
    table: &'t TableDef,
    /// The lower-case name and expression of each hidden expression column, which filters match before columns.
    hidden: Vec<(String, Expr)>,
    next: usize,
    /// The IDs of the leaves that are equalities and of those that test for NULL.
    equalities: BTreeSet<usize>,
    null_tests: BTreeSet<usize>,
    /// Whether the filter tree holds every part of the predicate.
    complete: bool,
}

/// swap returns the comparison with its operands exchanged.
pub(crate) fn swap(op: CmpOp) -> CmpOp {
    match op {
        CmpOp::Lt => CmpOp::Gt,
        CmpOp::Le => CmpOp::Ge,
        CmpOp::Gt => CmpOp::Lt,
        CmpOp::Ge => CmpOp::Le,
        other => other,
    }
}

/// is_constant reports whether an expression reads no column or subquery and calls no volatile function, so that
/// planning can evaluate it.
pub(crate) fn is_constant(e: &Expr) -> bool {
    let mut constant = true;
    e.visit(&mut |e| {
        if matches!(
            e,
            Expr::Column(_)
                | Expr::Outer(..)
                | Expr::Exists(_)
                | Expr::Scalar(_)
                | Expr::ArraySubquery(..)
                | Expr::AnySubquery(..)
                | Expr::SubPlan(_)
                | Expr::AlternativeSubPlan(_)
                | Expr::SubqueryValue
                | Expr::InputColumn(_)
                | Expr::AggRef(_)
                | Expr::WindowRef(_)
                | Expr::SetRef(_)
                | Expr::Default(_)
                | Expr::Routine(..)
        ) {
            constant = false;
        }
        if let Expr::Func(index, _) = e
            && crate::pgcatalog::is_volatile(crate::functions::function(*index).name)
        {
            constant = false;
        }
    });
    constant
}

impl Coster<'_> {
    /// column returns the lower-case name of the hidden expression column whose expression an expression is, or of the
    /// table column that it reads, under any casts.
    fn column(&self, e: &Expr) -> Option<String> {
        if let Some((name, _)) = self.hidden.iter().find(|(_, h)| h == e) {
            return Some(name.clone());
        }
        match e {
            Expr::Column(i) => self.table.columns.get(*i).map(|c| c.name.to_lowercase()),
            Expr::Cast(inner, ..) => self.column(inner),
            _ => None,
        }
    }

    /// leaf builds a leaf from a comparison of a column with a constant, or a NULL test of a column.
    fn leaf(&self, ctx: &mut Ctx<'_>, id: usize, e: &Expr) -> Option<Leaf> {
        let (column, op, value) = match e {
            Expr::Compare(op, l, r) => {
                let (column, op, constant) = match (self.column(l), self.column(r)) {
                    (Some(c), _) if is_constant(r) => (c, *op, r),
                    (_, Some(c)) if is_constant(l) => (c, swap(*op), l),
                    _ => return None,
                };
                (column, Op::Compare(op), constant.eval(ctx, &[]).ok()?)
            }
            Expr::IsNull(inner, negated) => {
                (self.column(inner)?, if *negated { Op::IsNotNull } else { Op::IsNull }, Value::Null)
            }
            Expr::Not(inner) => match inner.as_ref() {
                Expr::IsNull(inner, false) => (self.column(inner)?, Op::IsNotNull, Value::Null),
                _ => return None,
            },
            _ => return None,
        };
        Some(Leaf { id, column, op, value })
    }

    /// covers reports whether the ranges that a filter tree builds with the included filters hold exactly the rows
    /// that its predicate keeps: every node applies to the scan and every leaf compares the column with a value of
    /// its own kind.
    fn covers(&self, root: &Filter, include: &BTreeSet<usize>) -> bool {
        match root {
            Filter::Leaf(leaf) => include.contains(&leaf.id) && self.precise_leaf(leaf),
            Filter::Or(or) => include.contains(&or.id) && or.children.iter().all(|c| self.precise(c)),
            Filter::And(and) => {
                and.ors.iter().all(|o| include.contains(&o.id) && o.children.iter().all(|c| self.precise(c)))
                    && and.leaves.iter().flat_map(|(_, l)| l).all(|l| include.contains(&l.id) && self.precise_leaf(l))
            }
        }
    }

    /// precise reports whether every leaf under a filter node compares its column with a value of its own kind.
    fn precise(&self, filter: &Filter) -> bool {
        match filter {
            Filter::Leaf(leaf) => self.precise_leaf(leaf),
            Filter::Or(or) => or.children.iter().all(|c| self.precise(c)),
            Filter::And(and) => {
                and.ors.iter().all(|o| o.children.iter().all(|c| self.precise(c)))
                    && and.leaves.iter().flat_map(|(_, l)| l).all(|l| self.precise_leaf(l))
            }
        }
    }

    /// precise_leaf reports whether a leaf tests NULL or compares its column with a value of the column's own kind,
    /// so that the range it builds holds exactly the rows it keeps.
    fn precise_leaf(&self, leaf: &Leaf) -> bool {
        use crate::oid;
        let op = match leaf.op {
            Op::IsNull | Op::IsNotNull => return true,
            Op::Compare(op) => op,
        };
        if !matches!(op, CmpOp::Eq | CmpOp::Lt | CmpOp::Le | CmpOp::Gt | CmpOp::Ge) {
            return false;
        }
        let column = match self.hidden.iter().position(|(name, _)| *name == leaf.column) {
            Some(k) => self.table.hidden.get(k),
            None => self.table.columns.iter().find(|c| c.name.to_lowercase() == leaf.column),
        };
        let Some(column) = column else { return false };
        matches!(
            (column.ty.oid, &leaf.value),
            (oid::INT2 | oid::INT4 | oid::INT8, Value::Int2(_) | Value::Int4(_) | Value::Int8(_))
                | (oid::FLOAT4 | oid::FLOAT8, Value::Float4(_) | Value::Float8(_))
                | (oid::NUMERIC, Value::Numeric(_))
                | (oid::TEXT | oid::VARCHAR | oid::NAME, Value::Text(_))
                | (oid::BOOL, Value::Bool(_))
                | (oid::DATE, Value::Date(_))
                | (oid::TIMESTAMP, Value::Timestamp(_))
                | (oid::TIMESTAMPTZ, Value::TimestampTz(_))
                | (oid::UUID, Value::Uuid(_))
                | (oid::OID, Value::Oid(_))
        )
    }

    /// note records whether a leaf is an equality or a NULL test, which costing prefers.
    fn note(&mut self, leaf: &Leaf) {
        match leaf.op {
            Op::Compare(CmpOp::Eq) => {
                self.equalities.insert(leaf.id);
            }
            Op::IsNull => {
                self.null_tests.insert(leaf.id);
            }
            _ => {}
        }
    }

    /// build_root builds the filter tree of a predicate, as go-mysql-server's buildRoot does, or None when no part of
    /// it tests columns.
    fn build_root(&mut self, ctx: &mut Ctx<'_>, e: &Expr) -> Option<Filter> {
        let id = self.next;
        self.next += 1;
        match e {
            Expr::And(..) => {
                let mut and = And { id, ..And::default() };
                self.complete = self.build_and(ctx, e, &mut and);
                Some(Filter::And(and))
            }
            Expr::Or(..) => {
                let mut or = Or { id, children: Vec::new() };
                self.build_or(ctx, e, &mut or).then_some(Filter::Or(or))
            }
            _ => {
                let leaf = self.leaf(ctx, id, e)?;
                self.note(&leaf);
                Some(Filter::Leaf(leaf))
            }
        }
    }

    /// build_and adds the children of a conjunction to a conjunction node, as go-mysql-server's buildAnd does,
    /// reporting whether every child tests columns.
    fn build_and(&mut self, ctx: &mut Ctx<'_>, e: &Expr, and: &mut And) -> bool {
        let Expr::And(left, right) = e else { return false };
        let mut valid = true;
        for child in [left.as_ref(), right.as_ref()] {
            let id = self.next;
            self.next += 1;
            match child {
                Expr::And(..) => valid &= self.build_and(ctx, child, and),
                Expr::Or(..) => {
                    let mut or = Or { id, children: Vec::new() };
                    if self.build_or(ctx, child, &mut or) {
                        and.ors.push(or);
                        and.count += 1;
                    } else {
                        valid = false;
                    }
                }
                _ => match self.leaf(ctx, id, child) {
                    Some(leaf) => {
                        self.note(&leaf);
                        and.add_leaf(leaf);
                    }
                    None => valid = false,
                },
            }
        }
        valid
    }

    /// build_or adds the children of a disjunction to a disjunction node, as go-mysql-server's buildOr does, reporting
    /// whether every child tests columns.
    fn build_or(&mut self, ctx: &mut Ctx<'_>, e: &Expr, or: &mut Or) -> bool {
        let Expr::Or(left, right) = e else { return false };
        for child in [left.as_ref(), right.as_ref()] {
            match child {
                Expr::And(..) => {
                    let mut and = And { id: self.next, ..And::default() };
                    self.next += 1;
                    if !self.build_and(ctx, child, &mut and) {
                        return false;
                    }
                    or.children.push(Filter::And(and));
                }
                Expr::Or(..) => {
                    self.next += 1;
                    if !self.build_or(ctx, child, or) {
                        return false;
                    }
                }
                _ => {
                    let Some(leaf) = self.leaf(ctx, self.next, child) else { return false };
                    self.next += 1;
                    self.note(&leaf);
                    or.children.push(Filter::Leaf(leaf));
                }
            }
        }
        true
    }

    /// cost_and costs a conjunction against an index, as go-mysql-server's costIndexScanAnd does.
    fn cost_and(&self, and: &And, candidate: &Candidate) -> Cost {
        let mut filters: BTreeSet<usize> =
            and.ors.iter().filter(|o| self.cost_or(o, candidate)).map(|o| o.id).collect();
        let mut collector = Collector::default();
        for (ordinal, column) in candidate.columns.iter().enumerate() {
            for leaf in and.leaves.iter().filter(|(c, _)| c == column).flat_map(|(_, l)| l) {
                collector.add(leaf, ordinal);
            }
        }
        let fds = self.fds(candidate, &collector);
        let has_range = collector.inequality.contains(&collector.missing_prefix);
        filters.extend(collector.applied);
        Cost { filters, fds, prefix: collector.missing_prefix, has_range }
    }

    /// cost_or reports whether an index can answer every child of a disjunction, as go-mysql-server's
    /// costIndexScanOr does.
    fn cost_or(&self, or: &Or, candidate: &Candidate) -> bool {
        or.children.iter().all(|child| match child {
            Filter::And(and) => self.cost_and(and, candidate).filters.len() == and.count,
            Filter::Leaf(leaf) => candidate.columns.contains(&leaf.column),
            Filter::Or(_) => false,
        })
    }

    /// fds returns the dependencies of a unique index's lookup, or none for another index.
    fn fds(&self, candidate: &Candidate, collector: &Collector) -> Fds {
        if !candidate.unique {
            return Fds::default();
        }
        let not_null: BTreeSet<usize> =
            self.table.columns.iter().enumerate().filter(|(_, c)| !c.nullable).map(|(i, _)| i + 1).collect();
        Fds::lookup(&candidate.column_ids, &not_null, collector.constant.clone())
    }

    /// cost costs a filter tree against an index, as go-mysql-server's indexCoster.cost does.
    fn cost(&self, root: &Filter, candidate: &Candidate) -> Cost {
        match root {
            Filter::And(and) => self.cost_and(and, candidate),
            Filter::Or(or) => {
                let filters = if self.cost_or(or, candidate) { BTreeSet::from([or.id]) } else { BTreeSet::new() };
                Cost { filters, ..Cost::default() }
            }
            Filter::Leaf(leaf) => match candidate.columns.iter().position(|c| *c == leaf.column) {
                Some(ordinal) => {
                    let mut collector = Collector::default();
                    collector.add(leaf, ordinal);
                    let fds = self.fds(candidate, &collector);
                    Cost { filters: BTreeSet::from([leaf.id]), fds, prefix: collector.missing_prefix, has_range: false }
                }
                None => Cost::default(),
            },
        }
    }

    /// better reports whether an index's cost beats the best so far, as go-mysql-server's updateBest decides without
    /// statistics, where every index estimates the same number of rows.
    fn better(&self, best: &Best<'_>, candidate: &Candidate, cost: &Cost) -> bool {
        if cost.filters.is_empty() {
            return false;
        }
        let Some(current) = best.candidate else { return true };
        let old = &best.cost;
        if old.fds.max_one_row() {
            return false;
        }
        if old.prefix == 0 || cost.prefix == 0 && old.prefix != cost.prefix {
            return cost.prefix > old.prefix;
        }
        if cost.fds.max_one_row() {
            return true;
        }
        let same_prefix = |n: usize| {
            n <= current.columns.len() && n <= candidate.columns.len() && current.columns[..n] == candidate.columns[..n]
        };
        if cost.prefix > old.prefix && same_prefix(old.prefix) {
            return true;
        }
        if cost.prefix == old.prefix && same_prefix(old.prefix) && cost.has_range && !old.has_range {
            return true;
        }
        if old.prefix > cost.prefix && same_prefix(cost.prefix) {
            return false;
        }
        if old.prefix == cost.prefix && same_prefix(cost.prefix) && !cost.has_range && old.has_range {
            return false;
        }
        match (old.fds.strict_key(), cost.fds.strict_key()) {
            (None, Some(_)) => return true,
            (Some(_), None) => return false,
            (Some(b), Some(c)) if c.len() < b.len() => return true,
            _ => {}
        }
        let (old_lax, new_lax) = (old.fds.has_lax_key(), cost.fds.has_lax_key());
        match cost.fds.constants.len().cmp(&old.fds.constants.len()) {
            std::cmp::Ordering::Greater => return !(old_lax && !new_lax),
            std::cmp::Ordering::Less => return new_lax && !old_lax,
            std::cmp::Ordering::Equal => {}
        }
        match cost.filters.len().cmp(&old.filters.len()) {
            std::cmp::Ordering::Greater => return true,
            std::cmp::Ordering::Less => return false,
            std::cmp::Ordering::Equal => {}
        }
        let unused = |c: &Candidate, f: &BTreeSet<usize>| c.columns.len() as i64 - f.len() as i64;
        if unused(candidate, &cost.filters) < unused(current, &old.filters) {
            return true;
        }
        let count = |set: &BTreeSet<usize>, f: &BTreeSet<usize>| f.intersection(set).count();
        if count(&self.equalities, &cost.filters) > count(&self.equalities, &old.filters)
            || count(&self.null_tests, &cost.filters) > count(&self.null_tests, &old.filters)
        {
            return true;
        }
        if candidate.index.is_none() {
            return true;
        }
        if current.index.is_none() {
            return false;
        }
        candidate.name < current.name
    }

    /// build_ranges builds the ranges of an index's keys that a filter tree allows, applying the filters in
    /// `include` and anything under them, as go-mysql-server's indexScanRangeBuilder does.
    fn build_ranges(
        &self,
        root: &Filter,
        columns: &[(String, crate::catalog::ColumnType)],
        include: &BTreeSet<usize>,
    ) -> Vec<Range> {
        let in_scan = include.contains(&root.id());
        let ranges = match root {
            Filter::And(and) => self.ranges_and(and, columns, include, in_scan),
            Filter::Or(or) => self.ranges_or(or, columns, include, in_scan).unwrap_or_default(),
            Filter::Leaf(leaf) => self.ranges_leaf(leaf, columns, include, in_scan),
        };
        crate::ranges::remove_overlapping(ranges)
    }

    /// ranges_and builds the ranges a conjunction allows.
    fn ranges_and(
        &self,
        and: &And,
        columns: &[(String, crate::catalog::ColumnType)],
        include: &BTreeSet<usize>,
        in_scan: bool,
    ) -> Vec<Range> {
        let in_scan = in_scan || include.contains(&and.id);
        let mut result: Option<Vec<Range>> = None;
        for or in &and.ors {
            let Some(ranges) = self.ranges_or(or, columns, include, in_scan) else { continue };
            result = Some(match result {
                None => ranges,
                Some(previous) => crate::ranges::intersect(&previous, &ranges),
            });
        }
        let mut builder = IndexBuilder::new(columns);
        for leaf in and.sorted_leaves() {
            self.apply(&mut builder, leaf, include, in_scan);
        }
        match result {
            None => builder.ranges(),
            Some(previous) => crate::ranges::intersect(&previous, &builder.ranges()),
        }
    }

    /// ranges_or builds the ranges a disjunction allows, or None when the scan leaves it to the filter.
    fn ranges_or(
        &self,
        or: &Or,
        columns: &[(String, crate::catalog::ColumnType)],
        include: &BTreeSet<usize>,
        in_scan: bool,
    ) -> Option<Vec<Range>> {
        if !in_scan && !include.contains(&or.id) {
            return None;
        }
        let mut out = Vec::new();
        for child in &or.children {
            match child {
                Filter::And(and) => out.extend(self.ranges_and(and, columns, include, true)),
                Filter::Leaf(leaf) => out.extend(self.ranges_leaf(leaf, columns, include, true)),
                Filter::Or(_) => {}
            }
        }
        (!out.is_empty()).then_some(out)
    }

    /// ranges_leaf builds the ranges one leaf allows.
    fn ranges_leaf(
        &self,
        leaf: &Leaf,
        columns: &[(String, crate::catalog::ColumnType)],
        include: &BTreeSet<usize>,
        in_scan: bool,
    ) -> Vec<Range> {
        let mut builder = IndexBuilder::new(columns);
        self.apply(&mut builder, leaf, include, in_scan);
        builder.ranges()
    }

    /// apply narrows a builder's ranges by a leaf that the scan applies.
    fn apply(&self, builder: &mut IndexBuilder, leaf: &Leaf, include: &BTreeSet<usize>, in_scan: bool) {
        if !in_scan && !include.contains(&leaf.id) {
            return;
        }
        match leaf.op {
            Op::Compare(op) => {
                builder.compare(&leaf.column, op, &leaf.value);
            }
            Op::IsNull => builder.is_null(&leaf.column, false),
            Op::IsNotNull => builder.is_null(&leaf.column, true),
        }
    }
}

/// candidates returns the indexes a scan of a table may read: the primary key, then the secondary indexes, leaving out
/// vector indexes and the partial indexes whose predicate is not one of the filters, as go-mysql-server's
/// canUsePartialIndex does.
fn candidates(table: &TableDef, predicates: &[Option<Expr>], filters: &[&Expr]) -> Vec<Candidate> {
    let id_of = |c: usize| c + 1;
    let names = |columns: &[usize]| columns.iter().map(|&c| index_column_name(table, c)).collect();
    let mut out = Vec::new();
    if !table.keyless() {
        out.push(Candidate {
            index: None,
            name: "primary".into(),
            columns: names(&table.key_columns),
            unique: true,
            column_ids: table.key_columns.iter().map(|&c| id_of(c)).collect(),
        });
    }
    for (i, index) in table.indexes.iter().enumerate() {
        let usable = predicates[i].as_ref().is_none_or(|p| filters.contains(&p));
        if index.vector.is_some() || !usable {
            continue;
        }
        out.push(Candidate {
            index: Some(i),
            name: index.name.to_lowercase(),
            columns: names(&index.columns),
            unique: index.unique,
            column_ids: index.columns.iter().map(|&c| id_of(c)).collect(),
        });
    }
    out
}

/// choose returns the index scan that answers a table's filter, as go-mysql-server's getCostedIndexScan chooses it, or
/// None when a full scan serves as well.
pub fn choose(ctx: &mut Ctx<'_>, table: &TableDef, predicate: &Expr) -> Option<IndexScan> {
    choose_with_cover(ctx, table, predicate).map(|(scan, _)| scan)
}

/// choose_with_cover returns the index scan that `choose` returns, with whether its ranges hold exactly the rows that
/// the filter keeps, so that the filter need not run again, as go-mysql-server drops filters that leave nothing over.
pub fn choose_with_cover(ctx: &mut Ctx<'_>, table: &TableDef, predicate: &Expr) -> Option<(IndexScan, bool)> {
    let rules = ctx.index_rules(table).ok()?;
    let hidden = (0..table.hidden.len()).map(|k| index_column_name(table, HIDDEN_BASE + k));
    let hidden = hidden.zip(rules.hidden().iter().cloned()).collect();
    let mut coster =
        Coster { table, hidden, next: 1, equalities: BTreeSet::new(), null_tests: BTreeSet::new(), complete: true };
    let root = coster.build_root(ctx, &with_like_bounds(table, &predicate.clone().expand_row_compares()))?;
    let candidates = candidates(table, rules.predicates(), &conjuncts(predicate));
    let mut best = Best { candidate: None, cost: Cost::default() };
    for candidate in &candidates {
        let cost = coster.cost(&root, candidate);
        if coster.better(&best, candidate, &cost) {
            best = Best { candidate: Some(candidate), cost };
        }
    }
    let chosen = best.candidate?;
    let columns: Vec<(String, crate::catalog::ColumnType)> = match chosen.index {
        Some(i) => {
            table.indexes[i].columns.iter().map(|&c| (index_column_name(table, c), index_type(table, c))).collect()
        }
        None => {
            table.key_columns.iter().map(|&c| (table.columns[c].name.to_lowercase(), table.columns[c].ty)).collect()
        }
    };
    let ranges = coster.build_ranges(&root, &columns, &best.cost.filters);
    if crate::ranges::is_all_range(&ranges) {
        return None;
    }
    let covered = coster.complete && coster.covers(&root, &best.cost.filters);
    let scan = IndexScan {
        table: Arc::new(table.clone()),
        index: chosen.index,
        ranges,
        reverse: false,
        nearest: None,
        needed: None,
        lookup_heavy: None,
        parameterized: None,
    };
    Some((scan, covered))
}

/// scan_of_index returns the scan of one index of a table, in its order or the reverse, whose ranges the conjuncts of
/// a predicate narrow, with whether those ranges hold exactly the rows that the predicate keeps, as the planner's
/// index paths read an index that they chose, having proved a partial index's predicate. Without a predicate the
/// scan reads every entry.
pub(crate) fn scan_of_index(
    ctx: &mut Ctx<'_>,
    table: &Arc<TableDef>,
    index: Option<usize>,
    predicate: Option<&Expr>,
    reverse: bool,
) -> Option<(IndexScan, bool)> {
    let columns: Vec<(String, crate::catalog::ColumnType)> = match index {
        Some(i) => {
            table.indexes[i].columns.iter().map(|&c| (index_column_name(table, c), index_type(table, c))).collect()
        }
        None => {
            table.key_columns.iter().map(|&c| (table.columns[c].name.to_lowercase(), table.columns[c].ty)).collect()
        }
    };
    let scan = |ranges| IndexScan {
        table: table.clone(),
        index,
        ranges,
        reverse,
        nearest: None,
        needed: None,
        lookup_heavy: None,
        parameterized: None,
    };
    let Some(predicate) = predicate else { return Some((scan(IndexBuilder::new(&columns).ranges()), true)) };
    let rules = ctx.index_rules(table).ok()?;
    let hidden = (0..table.hidden.len()).map(|k| index_column_name(table, HIDDEN_BASE + k));
    let hidden = hidden.zip(rules.hidden().iter().cloned()).collect();
    let mut coster =
        Coster { table, hidden, next: 1, equalities: BTreeSet::new(), null_tests: BTreeSet::new(), complete: true };
    let root = coster.build_root(ctx, &with_like_bounds(table, &predicate.clone().expand_row_compares()))?;
    let unrestricted = vec![None; table.indexes.len()];
    let candidate = candidates(table, &unrestricted, &[]).into_iter().find(|c| c.index == index)?;
    let cost = coster.cost(&root, &candidate);
    let ranges = coster.build_ranges(&root, &columns, &cost.filters);
    let covered = coster.complete && coster.covers(&root, &cost.filters);
    Some((scan(ranges), covered))
}

/// with_like_bounds returns a predicate with each LIKE that its ANDs reach bounded by the fixed prefix of its pattern,
/// as Doltgres' AddLikePrefixRanges does, so that an index on the column can serve it.
fn with_like_bounds(table: &TableDef, e: &Expr) -> Expr {
    let and = |l: Expr, r: Expr| Expr::And(Box::new(l), Box::new(r));
    match e {
        Expr::And(l, r) => and(with_like_bounds(table, l), with_like_bounds(table, r)),
        Expr::Func(f, args) if crate::functions::function(*f).name == "textlike" => {
            let [column, Expr::Const(Value::Text(pattern))] = args.as_slice() else { return e.clone() };
            let read = match column {
                Expr::Cast(inner, ..) => inner.as_ref(),
                other => other,
            };
            let Expr::Column(c) = read else { return e.clone() };
            if !matches!(table.columns.get(*c).map(|c| c.ty.oid), Some(crate::oid::TEXT | crate::oid::VARCHAR)) {
                return e.clone();
            }
            let Some((lower, upper)) = like_prefix_bounds(pattern) else { return e.clone() };
            let bound = |op: CmpOp, text: String| {
                Expr::Compare(op, Box::new(column.clone()), Box::new(Expr::Const(Value::Text(text))))
            };
            let mut bounds = bound(CmpOp::Ge, lower);
            if let Some(upper) = upper {
                bounds = and(bounds, bound(CmpOp::Lt, upper));
            }
            and(bounds, e.clone())
        }
        other => other.clone(),
    }
}

/// like_prefix_bounds returns the bounds of the strings that a LIKE pattern's fixed prefix starts, at or above the
/// prefix and below the prefix with its last character's successor, as Postgres' like_support finds them, or None
/// when the pattern starts with a wildcard or an escape.
pub(crate) fn like_prefix_bounds(pattern: &str) -> Option<(String, Option<String>)> {
    let prefix = match pattern.find(['%', '_', '\\']) {
        Some(i) if pattern[i..].starts_with('\\') => return None,
        Some(i) => &pattern[..i],
        None => pattern,
    };
    if prefix.is_empty() {
        return None;
    }
    let last = prefix.chars().next_back().unwrap_or_default();
    let next = match last {
        '\u{D7FF}' => Some('\u{E000}'),
        last => char::from_u32(last as u32 + 1),
    };
    let upper = next.map(|next| format!("{}{next}", &prefix[..prefix.len() - last.len_utf8()]));
    Some((prefix.to_string(), upper))
}

/// conjuncts returns the expressions that a predicate ANDs together.
pub(crate) fn conjuncts(predicate: &Expr) -> Vec<&Expr> {
    match predicate {
        Expr::And(left, right) => {
            let mut out = conjuncts(left);
            out.extend(conjuncts(right));
            out
        }
        other => vec![other],
    }
}

/// index_column_name returns the lower-case name of a column at a position an index gives.
fn index_column_name(table: &TableDef, column: usize) -> String {
    table.index_column(column).map_or_else(String::new, |c| c.name.to_lowercase())
}

/// index_type returns the type of a column at a position an index gives.
fn index_type(table: &TableDef, column: usize) -> crate::catalog::ColumnType {
    table.index_column(column).map_or(crate::catalog::ColumnType { oid: 0, modifier: -1 }, |c| c.ty)
}

impl IndexScan {
    /// index_columns returns the table columns of the index's keys, in key order.
    pub fn index_columns(&self) -> Vec<usize> {
        match self.index {
            Some(i) => self.table.indexes[i].columns.clone(),
            None => self.table.key_columns.clone(),
        }
    }

    /// index_name returns the name of the index read, which is the primary key's constraint name for the primary key.
    pub fn index_name(&self) -> String {
        match self.index {
            Some(i) => self.table.indexes[i].name.clone(),
            None => self.table.primary_name(),
        }
    }

    /// covering reports whether the scan reads a secondary index that holds every column the plan above it reads.
    pub(crate) fn covering(&self) -> bool {
        self.needed.as_ref().is_some_and(|needed| self.covers(needed))
    }

    /// covers reports whether a scan of a secondary index holds the given columns, with the primary key that its
    /// entries end with.
    pub(crate) fn covers(&self, needed: &[usize]) -> bool {
        let Some(i) = self.index else { return false };
        let index = &self.table.indexes[i];
        !self.table.keyless()
            && self.nearest.is_none()
            && needed.iter().all(|c| index.columns.contains(c) || self.table.key_columns.contains(c))
    }

    /// bounds returns where the keys of a range lie in the index. Only columns of plain encodings in ascending order
    /// bound the keys, since other columns do not store values in the order that ranges compare them.
    fn bounds(&self, range: &Range) -> Bounds {
        let columns = self.index_columns();
        let mut fields: Vec<Option<Vec<u8>>> = Vec::new();
        let mut end = None;
        let mut lower = false;
        let mut exclusive = false;
        let mut exact = true;
        let mut used = 0;
        for (i, column_range) in range.iter().enumerate() {
            let Some(column) = columns.get(i).and_then(|&c| self.table.index_column(c)) else { break };
            let (descending, nulls_last) = match self.index {
                Some(x) => {
                    let index = &self.table.indexes[x];
                    (
                        index.descending.get(i).copied().unwrap_or(false),
                        index.nulls_last.get(i).copied().unwrap_or(false),
                    )
                }
                None => (false, false),
            };
            if descending || crate::storage::is_adaptive(column.encoding) {
                break;
            }
            let encode = |v: &Value| crate::storage::encode_field(v, column.encoding, column.ty).ok();
            match (&column_range.lower, &column_range.upper) {
                (Cut::BelowNull, Cut::AboveNull) => fields.push(None),
                (Cut::Below(low), Cut::Above(high))
                    if crate::expr::compare_values(low, high) == std::cmp::Ordering::Equal =>
                {
                    match encode(low) {
                        Some(field) => fields.push(field),
                        None => break,
                    }
                }
                (low, high) => {
                    let start = match low {
                        Cut::Below(v) | Cut::Above(v) => match encode(v) {
                            Some(field) => Some(field),
                            None => break,
                        },
                        _ => None,
                    };
                    // NULLs sort before every value unless the index puts them last, which keeps them out of the run of
                    // keys between the bounds, and a column that is NOT NULL has none.
                    let nulls_inside = matches!(low, Cut::BelowNull) && column.nullable;
                    let nulls_last = nulls_last && column.nullable;
                    let high = match high {
                        _ if nulls_inside && nulls_last => None,
                        Cut::Below(v) => encode(v).map(|high| (high, false)),
                        Cut::Above(v) => encode(v).map(|high| (high, true)),
                        _ => None,
                    };
                    exclusive = matches!(low, Cut::Above(_));
                    exact = match low {
                        Cut::Below(_) | Cut::Above(_) => true,
                        Cut::BelowNull => !nulls_last,
                        Cut::AboveNull => nulls_last || !column.nullable,
                        _ => false,
                    } && (high.is_some() || (matches!(column_range.upper, Cut::AboveAll) && !nulls_last));
                    end = high.map(|(high, inclusive)| {
                        let mut key: Vec<Option<&[u8]>> = fields.iter().map(Option::as_deref).collect();
                        key.push(high.as_deref());
                        (prolly::val::build_tuple(&key), inclusive)
                    });
                    if let Some(start) = start {
                        fields.push(start);
                        lower = true;
                    }
                    used = 1;
                    break;
                }
            }
        }
        let points = fields.len() - usize::from(lower);
        let unconstrained =
            |r: &crate::ranges::ColumnRange| matches!((&r.lower, &r.upper), (Cut::BelowNull, Cut::AboveAll));
        exact &= range.iter().skip(points + used).all(unconstrained);
        let key: Vec<Option<&[u8]>> = fields.iter().map(Option::as_deref).collect();
        Bounds { start: prolly::val::build_tuple(&key), start_width: fields.len(), exclusive, points, end, exact }
    }

    /// separated reports whether every key of one range sorts before every key of the next, as ranges of ascending
    /// columns that differ in their first column that is not the same single value in both do, unless the index puts
    /// the NULLs that one of them holds after every value.
    fn separated(&self, a: &Range, b: &Range) -> bool {
        let columns = self.index_columns();
        for (i, (x, y)) in a.iter().zip(b).enumerate() {
            let Some(column) = columns.get(i).and_then(|&c| self.table.index_column(c)) else { return false };
            let (descending, nulls_last) = match self.index {
                Some(index) => {
                    let index = &self.table.indexes[index];
                    (
                        index.descending.get(i).copied().unwrap_or(false),
                        index.nulls_last.get(i).copied().unwrap_or(false),
                    )
                }
                None => (false, false),
            };
            if descending || crate::storage::is_adaptive(column.encoding) {
                return false;
            }
            let point = |r: &crate::ranges::ColumnRange| match (&r.lower, &r.upper) {
                (Cut::Below(l), Cut::Above(h)) => crate::expr::compare_values(l, h) == std::cmp::Ordering::Equal,
                (Cut::BelowNull, Cut::AboveNull) => true,
                _ => false,
            };
            if point(x) && point(y) && x.lower.compare(&y.lower) == std::cmp::Ordering::Equal {
                continue;
            }
            let has_null = |r: &crate::ranges::ColumnRange| matches!(r.lower, Cut::BelowNull);
            if nulls_last && (has_null(x) || has_null(y)) {
                return false;
            }
            return x.upper.compare(&y.lower) != std::cmp::Ordering::Greater;
        }
        false
    }

    /// compare_prefix orders the first fields of two keys of the index read, as the index stores them.
    fn compare_prefix(&self, fields: usize, left: &[u8], right: &[u8]) -> std::cmp::Ordering {
        if let Some(i) = self.index {
            return self.table.compare_index_prefix(&self.table.indexes[i], fields, left, right);
        }
        let (left, right) = (prolly::Tuple(left), prolly::Tuple(right));
        for (i, &c) in self.table.key_columns.iter().enumerate().take(fields) {
            let column = &self.table.columns[c];
            let (l, r) = (left.field(i).ok().flatten(), right.field(i).ok().flatten());
            let order = crate::storage::compare_key_field(column.encoding, column.ty, l, r);
            if order != std::cmp::Ordering::Equal {
                return order;
            }
        }
        std::cmp::Ordering::Equal
    }

    /// closest returns the keys of the scan's vector index closest to its query vector, closest first, with their
    /// values.
    fn closest(&self, ctx: &mut Ctx<'_>, root: &prolly::Node, nearest: &Nearest) -> Result<Vec<(Vec<u8>, Vec<u8>)>> {
        let offset = crate::plan::limit_value(
            &nearest.offset,
            ctx,
            "OFFSET",
            crate::error::code::INVALID_ROW_COUNT_IN_RESULT_OFFSET_CLAUSE,
        )?;
        let limit = crate::plan::limit_value(
            &nearest.limit,
            ctx,
            "LIMIT",
            crate::error::code::INVALID_ROW_COUNT_IN_LIMIT_CLAUSE,
        )?;
        let limit = (limit.unwrap_or(0) + offset.unwrap_or(0)) as usize;
        let query = vector_of(&nearest.query.eval(ctx, &[])?)?;
        let index = &self.table.indexes[self.index.unwrap_or_default()];
        let distance = index.vector.unwrap_or(prolly::Distance::L2Squared);
        let column = &self.table.columns[index.columns[0]];
        let db = &*ctx.db;
        prolly::closest(db, root, limit, &mut |key| {
            let value = crate::storage::decode_field(db, prolly::Tuple(key).field(0)?, column.encoding, column.ty)?;
            let vector = vector_of(&value)?;
            if vector.len() != query.len() {
                return Err(crate::error::PgError::new(
                    crate::error::code::DATA_EXCEPTION,
                    format!("different vector dimensions {} and {}", vector.len(), query.len()),
                ));
            }
            Ok(distance.eval(&vector, &query))
        })
    }

    /// open starts reading the rows whose keys lie in the scan's ranges, a range at a time and seeking each range's
    /// first key, or reads them all at once when the ranges overlap or interleave or a vector search finds them.
    pub fn open<'p>(&'p self, ctx: &mut Ctx<'_>) -> Result<Box<dyn crate::exec::Rows + 'p>> {
        if self.parameterized.is_some() {
            return Ok(crate::exec::collected(self.run(ctx)?));
        }
        let streamed = self.nearest.is_none() && self.ranges.windows(2).all(|pair| self.separated(&pair[0], &pair[1]));
        if !streamed {
            return Ok(crate::exec::collected(self.run(ctx)?));
        }
        let root = match self.index {
            Some(i) => ctx.db.read(&self.table.indexes[i].root)?,
            None => Arc::new(prolly::Node::decode(self.table.table.primary_index.clone())?),
        };
        let mut bounds: Vec<(Bounds, usize)> =
            self.ranges.iter().enumerate().map(|(i, r)| (self.bounds(r), i)).collect();
        if self.reverse {
            bounds.reverse();
        }
        let exact = bounds.iter().all(|(b, _)| b.exact);
        Ok(Box::new(IndexRows {
            reader: self.reader()?,
            root,
            ranges: std::borrow::Cow::Borrowed(&self.ranges),
            bounds,
            exact,
            current: 0,
            items: None,
            edge: None,
            repeat: None,
        }))
    }

    /// open_lookup starts a reader of the scan's index that reads nothing until `IndexRows::restart` gives it ranges,
    /// which a join that looks up each left row's matches reuses for every lookup.
    pub(crate) fn open_lookup<'p>(&'p self, ctx: &mut Ctx<'_>) -> Result<IndexRows<'p>> {
        let root = match self.index {
            Some(i) => ctx.db.read(&self.table.indexes[i].root)?,
            None => Arc::new(prolly::Node::decode(self.table.table.primary_index.clone())?),
        };
        Ok(IndexRows {
            reader: self.reader()?,
            root,
            ranges: std::borrow::Cow::Owned(Vec::new()),
            bounds: Vec::new(),
            exact: true,
            current: 0,
            items: None,
            edge: None,
            repeat: None,
        })
    }

    /// estimate returns about how many rows the scan reads: the index entries between the ends of its first ranges,
    /// scaled up to the rest of them, which counts every entry of a range whose keys need checking.
    pub(crate) fn estimate(&self, ctx: &mut Ctx<'_>) -> Result<f64> {
        self.estimate_with_samples(ctx, ESTIMATED_RANGES)
    }

    /// estimate_with_samples estimates as `estimate` does from at most a number of the scan's first ranges.
    pub(crate) fn estimate_with_samples(&self, ctx: &mut Ctx<'_>, samples: usize) -> Result<f64> {
        if self.nearest.is_some() {
            return Ok(ESTIMATED_NEAREST);
        }
        let root = match self.index {
            Some(i) => ctx.db.read(&self.table.indexes[i].root)?,
            None => Arc::new(prolly::Node::decode(self.table.table.primary_index.clone())?),
        };
        let sampled = self.ranges.len().min(samples);
        let mut total = 0;
        for range in &self.ranges[..sampled] {
            let bounds = self.bounds(range);
            let start = bounds.walk(self, ctx.db, &root, false)?.ordinal()?;
            total += bounds.past(self, ctx.db, &root)?.saturating_sub(start);
        }
        Ok(total as f64 * self.ranges.len() as f64 / sampled.max(1) as f64)
    }

    /// count returns how many rows the scan reads, from the positions of its ranges' ends in the index, or None when
    /// that needs the rows themselves: a vector search, a keyless table, ranges that overlap, or a range whose keys
    /// need checking.
    pub fn count(&self, ctx: &mut Ctx<'_>) -> Result<Option<u64>> {
        if self.nearest.is_some()
            || self.table.keyless()
            || !self.ranges.windows(2).all(|pair| self.separated(&pair[0], &pair[1]))
        {
            return Ok(None);
        }
        let bounds: Vec<Bounds> = self.ranges.iter().map(|r| self.bounds(r)).collect();
        if !bounds.iter().all(|b| b.exact) {
            return Ok(None);
        }
        let root = match self.index {
            Some(i) => ctx.db.read(&self.table.indexes[i].root)?,
            None => Arc::new(prolly::Node::decode(self.table.table.primary_index.clone())?),
        };
        let mut total = 0;
        for bounds in &bounds {
            let start = bounds.walk(self, ctx.db, &root, false)?.ordinal()?;
            total += bounds.past(self, ctx.db, &root)?.saturating_sub(start);
        }
        Ok(Some(total))
    }

    /// reader returns what decodes the scan's index entries into rows.
    fn reader(&self) -> Result<Reader<'_>> {
        let covering = self.covering();
        let primary = match (self.index, covering) {
            (Some(_), false) => Some(Arc::new(prolly::Node::decode(self.table.table.primary_index.clone())?)),
            _ => None,
        };
        let needed = self.needed.as_ref().map(|columns| {
            let mut mask = vec![false; self.table.columns.len()];
            for &c in columns {
                if let Some(m) = mask.get_mut(c) {
                    *m = true;
                }
            }
            mask
        });
        Ok(Reader { scan: self, columns: self.index_columns(), covering, primary, lookup: None, needed })
    }

    /// run reads the rows whose keys lie in the scan's ranges, in index order or its reverse, or the rows a vector
    /// search finds, closest first.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        if let Some(cond) = &self.parameterized {
            let (built, cond, covered) = self.bound(ctx, cond)?;
            let mut rows = built.run(ctx)?;
            if !covered {
                let mut kept = Vec::with_capacity(rows.len());
                for row in rows {
                    if cond.is_true(ctx, &row)? {
                        kept.push(row);
                    }
                }
                rows = kept;
            }
            return Ok(rows);
        }
        let root = match self.index {
            Some(i) => ctx.db.read(&self.table.indexes[i].root)?,
            None => Arc::new(prolly::Node::decode(self.table.table.primary_index.clone())?),
        };
        let mut items: Vec<Item> = Vec::new();
        let mut exact = false;
        match &self.nearest {
            Some(nearest) => items = self.closest(ctx, &root, nearest)?,
            None => {
                exact = true;
                for range in &self.ranges {
                    let bounds = self.bounds(range);
                    exact &= bounds.exact;
                    let mut walk = bounds.walk(self, ctx.db, &root, false)?;
                    while let Some((key, value)) = walk.current()? {
                        if !bounds.holds(self, key, false) {
                            break;
                        }
                        items.push((key.to_vec(), value.to_vec()));
                        walk.advance(ctx.db)?;
                    }
                }
                if self.ranges.len() > 1 {
                    let width = match self.index {
                        Some(i) => self.table.index_key_columns(&self.table.indexes[i]).len(),
                        None => self.table.key_columns.len(),
                    };
                    let ordered =
                        |a: &Item, b: &Item| self.compare_prefix(width, &a.0, &b.0) == std::cmp::Ordering::Less;
                    if !items.windows(2).all(|pair| ordered(&pair[0], &pair[1])) {
                        items.sort_by(|a, b| self.compare_prefix(width, &a.0, &b.0));
                        items.dedup_by(|a, b| a.0 == b.0);
                    }
                }
            }
        }
        if self.reverse {
            items.reverse();
        }
        let mut reader = self.reader()?;
        let mut rows = Vec::new();
        for (key, value) in items {
            let check = (!exact && self.nearest.is_none()).then_some(self.ranges.as_slice());
            let Some((row, cardinality)) = reader.row(ctx.db, &key, &value, check)? else {
                continue;
            };
            for _ in 0..cardinality {
                rows.push(row.clone());
            }
        }
        Ok(rows)
    }

    /// bound returns the scan that a parameterized scan makes for its enclosing row, with its index conditions over that
    /// row's values and whether the scan's ranges hold exactly the rows that they keep.
    fn bound(&self, ctx: &mut Ctx<'_>, cond: &Expr) -> Result<(IndexScan, Expr, bool)> {
        let cond = bind_outer(ctx, cond.clone())?;
        let (built, covered) = match scan_of_index(ctx, &self.table, self.index, Some(&cond), self.reverse) {
            Some(found) => found,
            None => (IndexScan { parameterized: None, ..self.clone() }, false),
        };
        Ok((IndexScan { needed: self.needed.clone(), ..built }, cond, covered))
    }

    /// index_values decodes the index columns of an entry of the scan's index.
    fn index_values(&self, db: &mut Database, columns: &[usize], key: &[u8]) -> Result<Vec<Value>> {
        let tuple = prolly::Tuple(key);
        let mut values = Vec::with_capacity(columns.len());
        for (field, &c) in columns.iter().enumerate() {
            let column = self.table.index_column(c).expect("an index column");
            values.push(crate::storage::decode_field(db, tuple.field(field)?, column.encoding, column.ty)?);
        }
        Ok(values)
    }

    /// primary_keys returns the primary keys of the rows whose index entries lie in the scan's ranges, in the table's
    /// key order and without duplicates, as a bitmap index scan marks them.
    fn primary_keys(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<u8>>> {
        let root = match self.index {
            Some(i) => ctx.db.read(&self.table.indexes[i].root)?,
            None => Arc::new(prolly::Node::decode(self.table.table.primary_index.clone())?),
        };
        let columns = self.index_columns();
        let mut keys = Vec::new();
        for range in &self.ranges {
            let bounds = self.bounds(range);
            let mut walk = bounds.walk(self, ctx.db, &root, false)?;
            while let Some((key, _)) = walk.current()? {
                if !bounds.holds(self, key, false) {
                    break;
                }
                let key = key.to_vec();
                walk.advance(ctx.db)?;
                if !bounds.exact && !crate::ranges::range_contains(range, &self.index_values(ctx.db, &columns, &key)?) {
                    continue;
                }
                keys.push(match self.index {
                    Some(i) => primary_key_of(&self.table, &self.table.indexes[i], &key)?,
                    None => key,
                });
            }
        }
        keys.sort_by(|a, b| self.table.compare_keys(a, b));
        keys.dedup();
        Ok(keys)
    }
}

/// primary_key_of returns the primary key that an entry of a secondary index holds after its own columns, or the
/// content hash of a keyless table's row.
fn primary_key_of(table: &TableDef, index: &crate::catalog::table::IndexDef, key: &[u8]) -> Result<Vec<u8>> {
    let tuple = prolly::Tuple(key);
    let mut fields = Vec::with_capacity(table.key_columns.len() + 1);
    let mut extra = index.columns.len();
    if table.keyless() {
        fields.push(tuple.field(extra)?);
    }
    for c in &table.key_columns {
        let position = match index.columns.iter().position(|ic| ic == c) {
            Some(p) => p,
            None => {
                extra += 1;
                extra - 1
            }
        };
        fields.push(tuple.field(position)?);
    }
    Ok(prolly::val::build_tuple(&fields))
}

/// BitmapHeapScan reads the rows of a table whose primary keys a tree of index scans finds, in primary key order, as
/// Postgres' bitmap heap scan reads the table pages that its bitmap marks.
#[derive(Clone, Debug, PartialEq)]
pub struct BitmapHeapScan {
    pub table: Arc<TableDef>,
    pub bitmap: Bitmap,
    /// The clauses that the bitmap's index scans answer and no filter above tests, which EXPLAIN shows, and which the
    /// scan tests on each row when it is lossy, as the ranges of some index scan keep keys that its conditions do not.
    pub recheck: Option<Expr>,
    pub lossy: bool,
    /// The table columns that the plan above the scan reads, or None for every column.
    pub needed: Option<Vec<usize>>,
}

/// Bitmap is a tree of index scans whose primary keys a bitmap heap scan reads, as Postgres' bitmap index scans,
/// BitmapAnd, and BitmapOr make bitmaps.
#[derive(Clone, Debug, PartialEq)]
pub enum Bitmap {
    /// The keys that a scan of one index finds.
    Index(Box<IndexScan>),
    And(Vec<Bitmap>),
    Or(Vec<Bitmap>),
}

impl Bitmap {
    /// conditions returns the index conditions of the tree's parameterized scans.
    pub(crate) fn conditions(&self) -> Vec<&Expr> {
        match self {
            Bitmap::Index(scan) => scan.parameterized.iter().collect(),
            Bitmap::And(children) | Bitmap::Or(children) => children.iter().flat_map(Bitmap::conditions).collect(),
        }
    }

    /// conditions_mut returns the index conditions of the tree's parameterized scans to change.
    pub(crate) fn conditions_mut(&mut self) -> Vec<&mut Expr> {
        match self {
            Bitmap::Index(scan) => scan.parameterized.iter_mut().collect(),
            Bitmap::And(children) | Bitmap::Or(children) => {
                children.iter_mut().flat_map(Bitmap::conditions_mut).collect()
            }
        }
    }

    /// keys returns the primary keys that the tree finds, in the table's key order, setting `lossy` when the ranges of
    /// a parameterized scan keep keys that its conditions do not.
    fn keys(&self, ctx: &mut Ctx<'_>, table: &TableDef, lossy: &mut bool) -> Result<Vec<Vec<u8>>> {
        let order = |a: &Vec<u8>, b: &Vec<u8>| table.compare_keys(a, b);
        match self {
            Bitmap::Index(scan) => match &scan.parameterized {
                Some(cond) => {
                    let (built, _, covered) = scan.bound(ctx, cond)?;
                    *lossy |= !covered;
                    built.primary_keys(ctx)
                }
                None => scan.primary_keys(ctx),
            },
            Bitmap::And(children) => {
                let mut keys = children[0].keys(ctx, table, lossy)?;
                for child in &children[1..] {
                    let other = child.keys(ctx, table, lossy)?;
                    let mut j = 0;
                    keys.retain(|key| {
                        while j < other.len() && order(&other[j], key) == std::cmp::Ordering::Less {
                            j += 1;
                        }
                        j < other.len() && order(&other[j], key) == std::cmp::Ordering::Equal
                    });
                }
                Ok(keys)
            }
            Bitmap::Or(children) => {
                let mut keys = Vec::new();
                for child in children {
                    keys.extend(child.keys(ctx, table, lossy)?);
                }
                keys.sort_by(order);
                keys.dedup();
                Ok(keys)
            }
        }
    }
}

/// bind_outer replaces the enclosing row's columns that an expression reads with their values.
fn bind_outer(ctx: &mut Ctx<'_>, e: Expr) -> Result<Expr> {
    match e {
        Expr::Outer(..) => Ok(Expr::Const(e.eval(ctx, &[])?)),
        other => {
            let mut failed = None;
            let bound = other.map_children(&mut |c| {
                bind_outer(ctx, c).unwrap_or_else(|err| {
                    failed = Some(err);
                    Expr::Const(Value::Null)
                })
            });
            failed.map_or(Ok(bound), Err)
        }
    }
}

impl BitmapHeapScan {
    /// run reads the rows whose primary keys the scan's bitmap finds, testing the recheck conditions on each when the
    /// scan is lossy.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        let table = &*self.table;
        let mut lossy = self.lossy;
        let keys = self.bitmap.keys(ctx, table, &mut lossy)?;
        let needed = self.needed.as_ref().filter(|_| !lossy).map(|columns| {
            let mut mask = vec![false; table.columns.len()];
            for &c in columns {
                if let Some(m) = mask.get_mut(c) {
                    *m = true;
                }
            }
            mask
        });
        let primary = Arc::new(prolly::Node::decode(table.table.primary_index.clone())?);
        let compare = |a: &[u8], b: &[u8]| table.compare_keys(a, b);
        let mut walk: Option<prolly::Items> = None;
        let mut rows = Vec::new();
        for key in &keys {
            let items = match &mut walk {
                Some(items) => {
                    items.seek(ctx.db, key, &compare)?;
                    items
                }
                None => walk.insert(prolly::Items::at_key(ctx.db, primary.clone(), key, &compare)?),
            };
            let Some((found, stored)) = items.current()? else { break };
            if compare(found, key) != std::cmp::Ordering::Equal {
                continue;
            }
            let mut row = Vec::new();
            let cardinality = table.decode_columns_into(ctx.db, key, stored, needed.as_deref(), &mut row)?;
            if lossy
                && let Some(recheck) = &self.recheck
                && !recheck.is_true(ctx, &row)?
            {
                continue;
            }
            for _ in 1..cardinality {
                rows.push(row.clone());
            }
            if cardinality > 0 {
                rows.push(row);
            }
        }
        Ok(rows)
    }
}

/// Reader decodes the entries of the index a scan reads into the table's rows.
struct Reader<'s> {
    scan: &'s IndexScan,
    /// The table columns of the index's keys, in key order.
    columns: Vec<usize>,
    covering: bool,
    /// The primary index, which a scan of a secondary index that does not cover the plan looks rows up in, and the
    /// walk of it that the lookups move.
    primary: Option<Arc<prolly::Node>>,
    lookup: Option<prolly::Items>,
    /// Which table columns the plan reads, or None for every column.
    needed: Option<Vec<bool>>,
}

impl Reader<'_> {
    /// row returns the row of an index entry with its cardinality, or None when the entry lies outside all of the
    /// ranges to check it against, if any, or its row is missing from the primary index.
    fn row(
        &mut self,
        db: &mut Database,
        key: &[u8],
        value: &[u8],
        check: Option<&[Range]>,
    ) -> Result<Option<(Vec<Value>, u64)>> {
        let mut row = Vec::new();
        Ok(self.row_into(db, key, value, check, &mut row)?.map(|cardinality| (row, cardinality)))
    }

    /// row_into is `row` into a buffer that holds nothing or a row it put there before, whose columns that the plan
    /// does not read are still NULL, returning the cardinality.
    fn row_into(
        &mut self,
        db: &mut Database,
        key: &[u8],
        value: &[u8],
        check: Option<&[Range]>,
        row: &mut Vec<Value>,
    ) -> Result<Option<u64>> {
        let (scan, table) = (self.scan, &*self.scan.table);
        let tuple = prolly::Tuple(key);
        let wanted = |c: usize| self.needed.as_ref().is_none_or(|n| n.get(c).copied().unwrap_or(false));
        let mut values: Vec<Value> = Vec::new();
        if let Some(ranges) = check {
            values = scan.index_values(db, &self.columns, key)?;
            if !ranges.iter().any(|r| crate::ranges::range_contains(r, &values)) {
                return Ok(None);
            }
        }
        let Some(i) = scan.index else {
            return table.decode_columns_into(db, key, value, self.needed.as_deref(), row).map(Some);
        };
        let index = &table.indexes[i];
        if self.covering {
            if row.len() != table.columns.len() {
                *row = Value::nulls(table.columns.len());
            }
            for (field, &c) in self.columns.iter().enumerate() {
                if c < row.len() && wanted(c) {
                    row[c] = match values.get_mut(field) {
                        Some(value) => std::mem::replace(value, Value::Null),
                        None => {
                            let column = &table.columns[c];
                            crate::storage::decode_field(db, tuple.field(field)?, column.encoding, column.ty)?
                        }
                    };
                }
            }
            let width = index.columns.len();
            let mut position = width;
            for &c in &table.key_columns {
                if index.columns.contains(&c) {
                    continue;
                }
                if wanted(c) {
                    let column = &table.columns[c];
                    row[c] = crate::storage::decode_field(db, tuple.field(position)?, column.encoding, column.ty)?;
                }
                position += 1;
            }
            return Ok(Some(1));
        }
        let primary_key = primary_key_of(table, index, key)?;
        let compare = |a: &[u8], b: &[u8]| table.compare_keys(a, b);
        let lookup = match &mut self.lookup {
            Some(lookup) => {
                lookup.seek(db, &primary_key, &compare)?;
                lookup
            }
            None => {
                let primary = self.primary.clone().expect("a primary index");
                self.lookup.insert(prolly::Items::at_key(db, primary, &primary_key, &compare)?)
            }
        };
        match lookup.current()? {
            Some((key, stored)) if compare(key, &primary_key) == std::cmp::Ordering::Equal => {
                table.decode_columns_into(db, &primary_key, stored, self.needed.as_deref(), row).map(Some)
            }
            _ => Ok(None),
        }
    }
}

impl Bounds {
    /// walk returns a walk of the index at the range's first key, or at its last key in reverse.
    fn walk(
        &self,
        scan: &IndexScan,
        db: &mut Database,
        root: &Arc<prolly::Node>,
        reverse: bool,
    ) -> Result<prolly::Items> {
        if !reverse {
            let compare = |a: &[u8], b: &[u8]| match scan.compare_prefix(self.start_width, a, b) {
                std::cmp::Ordering::Equal if self.exclusive => std::cmp::Ordering::Greater,
                order => order,
            };
            return Ok(prolly::Items::at_key(db, root.clone(), &self.start, &compare)?);
        }
        // Seek the first key past the range, then step back to its last key.
        let (target, width) = match &self.end {
            Some((end, inclusive)) => (end, (self.points + 1, *inclusive)),
            None if self.points > 0 => (&self.start, (self.points, true)),
            None => return Ok(prolly::Items::last(db, root.clone())?),
        };
        let (width, past_equal) = width;
        let compare = |a: &[u8], b: &[u8]| match scan.compare_prefix(width, a, b) {
            std::cmp::Ordering::Equal if past_equal => std::cmp::Ordering::Greater,
            order => order,
        };
        let mut items = prolly::Items::at_key(db, root.clone(), target, &compare)?;
        if items.current()?.is_none() {
            return Ok(prolly::Items::last(db, root.clone())?);
        }
        items.retreat(db)?;
        Ok(items)
    }

    /// past returns the ordinal of the first key after the range: past its end, past the keys that share its single
    /// values, or the end of the index.
    fn past(&self, scan: &IndexScan, db: &mut Database, root: &Arc<prolly::Node>) -> Result<u64> {
        let (target, width, past_equal) = match &self.end {
            Some((end, inclusive)) => (end, self.points + 1, *inclusive),
            None if self.points > 0 => (&self.start, self.points, true),
            None => return Ok(root.tree_count()),
        };
        let compare = |a: &[u8], b: &[u8]| match scan.compare_prefix(width, a, b) {
            std::cmp::Ordering::Equal if past_equal => std::cmp::Ordering::Greater,
            order => order,
        };
        Ok(prolly::Items::at_key(db, root.clone(), target, &compare)?.ordinal()?)
    }

    /// holds reports whether a key that a walk reached still lies within the range's bounds, given the direction.
    fn holds(&self, scan: &IndexScan, key: &[u8], reverse: bool) -> bool {
        use std::cmp::Ordering::{Equal, Greater, Less};
        if self.points > 0 && scan.compare_prefix(self.points, key, &self.start) != Equal {
            return false;
        }
        if reverse {
            let lower = || match scan.compare_prefix(self.start_width, key, &self.start) {
                Equal => !self.exclusive,
                order => order == Greater,
            };
            return self.start_width == self.points || lower();
        }
        match &self.end {
            Some((end, inclusive)) => match scan.compare_prefix(self.points + 1, key, end) {
                Greater => false,
                Equal => *inclusive,
                Less => true,
            },
            None => true,
        }
    }
}

/// IndexRows hands out the rows of an index scan whose ranges are in key order, walking one range at a time.
pub(crate) struct IndexRows<'p> {
    reader: Reader<'p>,
    root: Arc<prolly::Node>,
    /// The ranges read, which are the scan's own unless a lookup gave others.
    ranges: std::borrow::Cow<'p, [Range]>,
    /// The bounds of each range with the range's position among the ranges, in the order the scan reads them.
    bounds: Vec<(Bounds, usize)>,
    /// Whether every range is exact.
    exact: bool,
    current: usize,
    items: Option<prolly::Items>,
    /// The position the walk reaches next if it stays in its leaf, and the position in that leaf where the range's
    /// items end: the first one past the range walking forward, or the first one in it walking backward.
    edge: Option<(usize, usize)>,
    /// A row of a keyless table to hand out again, with how many more times.
    repeat: Option<(Vec<Value>, u64)>,
}

/// partition returns the first position from `start` up to `end` where a test that holds for a run of positions and
/// then fails stops holding.
fn partition(start: usize, end: usize, holds: impl Fn(usize) -> store::Result<bool>) -> Result<usize> {
    let (mut low, mut high) = (start, end);
    while low < high {
        let middle = low + (high - low) / 2;
        if holds(middle)? {
            low = middle + 1;
        } else {
            high = middle;
        }
    }
    Ok(low)
}

impl IndexRows<'_> {
    /// restart makes the reader read the rows of other ranges from their start.
    pub(crate) fn restart(&mut self, ranges: Vec<Range>) {
        let scan = self.reader.scan;
        self.bounds = ranges.iter().enumerate().map(|(i, r)| (scan.bounds(r), i)).collect();
        self.exact = self.bounds.iter().all(|(b, _)| b.exact);
        self.ranges = std::borrow::Cow::Owned(ranges);
        self.current = 0;
        self.items = None;
        self.edge = None;
        self.repeat = None;
    }
}

impl crate::exec::Rows for IndexRows<'_> {
    fn skip(&mut self, ctx: &mut Ctx<'_>, count: usize) -> Result<usize> {
        let (scan, reverse) = (self.reader.scan, self.reader.scan.reverse);
        // Only an exact scan of the primary index or of a covering index has a row for every entry it walks.
        if !self.exact
            || self.repeat.is_some()
            || scan.table.keyless()
            || (scan.index.is_some() && !self.reader.covering)
        {
            let mut skipped = 0;
            while skipped < count && self.next(ctx)?.is_some() {
                skipped += 1;
            }
            return Ok(skipped);
        }
        let mut skipped = 0;
        while skipped < count {
            let Some((bounds, _)) = self.bounds.get(self.current) else { break };
            if self.items.is_none() {
                self.items = Some(bounds.walk(scan, ctx.db, &self.root, reverse)?);
            }
            let items = self.items.as_mut().expect("a walk");
            match items.current()? {
                Some((key, _)) if bounds.holds(scan, key, reverse) => {}
                _ => {
                    self.items = None;
                    self.current += 1;
                    continue;
                }
            }
            match reverse {
                true => items.retreat(ctx.db)?,
                false => items.advance(ctx.db)?,
            }
            skipped += 1;
        }
        Ok(skipped)
    }

    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Vec<Value>>> {
        let mut row = Vec::new();
        Ok(self.next_into(ctx, &mut row)?.then_some(row))
    }

    fn next_into(&mut self, ctx: &mut Ctx<'_>, out: &mut Vec<Value>) -> Result<bool> {
        if let Some((repeated, remaining)) = self.repeat.as_mut() {
            out.clone_from(repeated);
            *remaining -= 1;
            if *remaining == 0 {
                self.repeat = None;
            }
            return Ok(true);
        }
        let (scan, reverse) = (self.reader.scan, self.reader.scan.reverse);
        loop {
            let Some((bounds, range)) = self.bounds.get(self.current) else { return Ok(false) };
            if self.items.is_none() {
                self.items = Some(bounds.walk(scan, ctx.db, &self.root, reverse)?);
            }
            let items = self.items.as_mut().expect("a walk");
            let Some((leaf, at)) = items.leaf() else {
                self.items = None;
                self.current += 1;
                continue;
            };
            // Each leaf's items hold the bounds up to one position in the direction of the walk, which a binary
            // search finds once for the leaf. The walk is still in the leaf when it moved by one position.
            let edge = match self.edge {
                Some((expected, edge)) if expected == at => edge,
                _ => {
                    let holds = |i: usize| leaf.key(i).map(|key| bounds.holds(scan, key, reverse));
                    match reverse {
                        false => partition(at, leaf.count(), holds)?,
                        true => partition(0, at + 1, |i| holds(i).map(|h| !h))?,
                    }
                }
            };
            self.edge = match reverse {
                false => Some((at + 1, edge)),
                true => at.checked_sub(1).map(|previous| (previous, edge)),
            };
            if (!reverse && at >= edge) || (reverse && at < edge) {
                self.items = None;
                self.edge = None;
                self.current += 1;
                continue;
            }
            let (key, value) = (leaf.key(at)?, leaf.value(at)?);
            let check = (!bounds.exact).then(|| std::slice::from_ref(&self.ranges[*range]));
            let cardinality = self.reader.row_into(ctx.db, key, value, check, out)?;
            match reverse {
                true => items.retreat(ctx.db)?,
                false => items.advance(ctx.db)?,
            }
            let Some(cardinality) = cardinality.filter(|&c| c > 0) else { continue };
            if cardinality > 1 {
                self.repeat = Some((out.clone(), cardinality - 1));
            }
            return Ok(true);
        }
    }
}

/// Order is how a scan orders a column: descending or not, and with NULLs first or not.
type Order = (bool, bool);

/// index_orders returns how an index orders its columns, the primary key ascending.
fn index_orders(table: &TableDef, index: Option<usize>) -> Vec<Order> {
    match index {
        Some(i) => {
            let index = &table.indexes[i];
            (0..index.columns.len()).map(|c| (index.descending[c], !index.nulls_last[c])).collect()
        }
        None => table.key_columns.iter().map(|_| (false, false)).collect(),
    }
}

/// hash_ordered reports whether an index orders its keys by content hashes rather than values, as Dolt's
/// HasContentHashedField does: a unique index with a column stored out of band or adaptively.
pub(crate) fn hash_ordered(table: &TableDef, index: Option<usize>) -> bool {
    let (unique, columns) = match index {
        Some(i) => (table.indexes[i].unique, &table.indexes[i].columns),
        None => (true, &table.key_columns),
    };
    unique
        && columns.iter().any(|&c| {
            let encoding = table.index_column(c).map_or(0, |c| c.encoding);
            use prolly::val::encoding::*;
            matches!(encoding, BYTES_ADDR | COMMIT_ADDR | STRING_ADDR | JSON_ADDR | GEOM_ADDR | EXTENDED_ADDR)
                || crate::storage::is_adaptive(encoding)
        })
}

/// provides reports whether reading an index in a direction gives the sort keys' order for the index's columns at the
/// positions given, returning the direction it must read in, where NULL placement only matters for a nullable column
/// and json columns, whose keys sort by their text, give no order.
fn provides(
    table: &TableDef,
    columns: &[usize],
    orders: &[Order],
    keys: &[(usize, usize, &crate::plan::SortKey)],
) -> Option<bool> {
    let mut reverse = None;
    for &(position, _, key) in keys {
        if matches!(table.columns[columns[position]].ty.oid, crate::oid::JSON | crate::oid::JSONB) {
            return None;
        }
        let (descending, nulls_first) = orders[position];
        let flip = key.descending != descending;
        if reverse.is_some_and(|r| r != flip) {
            return None;
        }
        reverse = Some(flip);
        let provided = nulls_first != flip;
        if provided != key.nulls_first && table.columns[columns[position]].nullable {
            return None;
        }
    }
    reverse
}

/// matching pairs each sort key with the index column it orders, as go-mysql-server's
/// sortExprsMatchIdxColExprsWithConstantColumns does: the keys must follow the index's columns in order, skipping
/// columns that equalities fix, which keys may also name.
fn matching(key_columns: &[usize], columns: &[usize], constant: &BTreeSet<usize>) -> Option<Vec<(usize, usize)>> {
    if key_columns.len() > columns.len() {
        return None;
    }
    let mut pairs = Vec::new();
    let mut k = 0;
    let mut out_of_order = Vec::new();
    for (position, column) in columns.iter().enumerate() {
        if k >= key_columns.len() {
            break;
        }
        if key_columns[k] == *column {
            if !constant.contains(&position) {
                pairs.push((position, k));
            }
            k += 1;
        } else if !constant.contains(&position) {
            out_of_order.push(key_columns[k]);
        }
    }
    out_of_order.extend(&key_columns[k..]);
    for column in out_of_order {
        let position = columns.iter().position(|c| *c == column)?;
        if !constant.contains(&position) {
            return None;
        }
    }
    Some(pairs)
}

/// constant_columns returns the positions of the index columns that a scan's one range fixes to a value.
fn constant_columns(ranges: &[Range]) -> BTreeSet<usize> {
    let [range] = ranges else { return BTreeSet::new() };
    range
        .iter()
        .enumerate()
        .filter(|(_, r)| {
            matches!((&r.lower, &r.upper), (crate::ranges::Cut::Below(a), crate::ranges::Cut::Above(b))
            if crate::expr::compare_values(a, b) == std::cmp::Ordering::Equal)
        })
        .map(|(i, _)| i)
        .collect()
}

/// ordered returns a plan whose rows come in the sort keys' order by reading its table through an index, as
/// go-mysql-server's replaceIdxSort does, or None when no index gives that order.
pub(crate) fn ordered(plan: &Plan, keys: &[crate::plan::SortKey]) -> Option<Plan> {
    match plan {
        Plan::Project { input, exprs } => {
            let mut mapped = Vec::with_capacity(keys.len());
            for key in keys {
                let Expr::Column(i) = key.expr else { return None };
                mapped.push(crate::plan::SortKey { expr: exprs.get(i)?.clone(), ..key.clone() });
            }
            let input = ordered(input, &mapped)?;
            Some(Plan::Project { input: Box::new(input), exprs: exprs.clone() })
        }
        Plan::Filter { input, predicate } => {
            Some(Plan::Filter { input: Box::new(ordered(input, keys)?), predicate: predicate.clone() })
        }
        Plan::Distinct { input, keys: None } => {
            Some(Plan::Distinct { input: Box::new(ordered(input, keys)?), keys: None })
        }
        Plan::Join {
            left,
            right,
            kind: crate::plan::JoinKind::Inner,
            condition,
            lateral: false,
            method: crate::plan::JoinMethod::Unplanned | crate::plan::JoinMethod::Ordered,
        } => {
            let width = left.width();
            if keys.iter().any(|k| !matches!(k.expr, Expr::Column(i) if i < width)) {
                return None;
            }
            if keys.iter().any(|k| k.descending != keys[0].descending) {
                return None;
            }
            Some(Plan::Join {
                left: Box::new(ordered(left, keys)?),
                right: right.clone(),
                kind: crate::plan::JoinKind::Inner,
                condition: condition.clone(),
                lateral: false,
                method: crate::plan::JoinMethod::Ordered,
            })
        }
        Plan::Scan(table, _) => {
            let key_columns: Vec<usize> = keys
                .iter()
                .map(|k| if let Expr::Column(c) = k.expr { Some(c) } else { None })
                .collect::<Option<_>>()?;
            let mut indexes: Vec<Option<usize>> = Vec::new();
            if !table.keyless() {
                indexes.push(None);
            }
            let whole = |i: usize| table.indexes[i].vector.is_none() && table.indexes[i].predicate.is_empty();
            indexes.extend((0..table.indexes.len()).filter(|&i| whole(i)).map(Some));
            for index in indexes.into_iter().filter(|&i| !hash_ordered(table, i)) {
                let columns = match index {
                    Some(i) => table.indexes[i].columns.clone(),
                    None => table.key_columns.clone(),
                };
                if key_columns.len() > columns.len() || columns[..key_columns.len()] != key_columns[..] {
                    continue;
                }
                let paired: Vec<(usize, usize, &crate::plan::SortKey)> =
                    keys.iter().enumerate().map(|(k, key)| (k, k, key)).collect();
                let Some(reverse) = provides(table, &columns, &index_orders(table, index), &paired) else { continue };
                let names: Vec<(String, crate::catalog::ColumnType)> =
                    columns.iter().map(|&c| (table.columns[c].name.to_lowercase(), table.columns[c].ty)).collect();
                let ranges = IndexBuilder::new(&names).ranges();
                return Some(Plan::IndexScan(Box::new(IndexScan {
                    table: Arc::new((**table).clone()),
                    index,
                    ranges,
                    reverse,
                    nearest: None,
                    needed: None,
                    lookup_heavy: None,
                    parameterized: None,
                })));
            }
            None
        }
        Plan::IndexScan(scan) => {
            let key_columns: Vec<usize> = keys
                .iter()
                .map(|k| if let Expr::Column(c) = k.expr { Some(c) } else { None })
                .collect::<Option<_>>()?;
            if hash_ordered(&scan.table, scan.index) {
                return None;
            }
            let columns = scan.index_columns();
            let pairs = matching(&key_columns, &columns, &constant_columns(&scan.ranges))?;
            for column in 0..keys.len().min(columns.len()) {
                for (i, a) in scan.ranges.iter().enumerate() {
                    if scan.ranges[i + 1..].iter().any(|b| a[column].try_intersect(&b[column]).is_some()) {
                        return None;
                    }
                }
            }
            let paired: Vec<(usize, usize, &crate::plan::SortKey)> =
                pairs.iter().map(|&(position, k)| (position, k, &keys[k])).collect();
            let reverse = if paired.is_empty() {
                false
            } else {
                provides(&scan.table, &columns, &index_orders(&scan.table, scan.index), &paired)?
            };
            Some(Plan::IndexScan(Box::new(IndexScan { reverse, lookup_heavy: None, ..(**scan).clone() })))
        }
        _ => None,
    }
}

/// order_by_index removes a sort whose order an index of its table already gives, reading the table through that
/// index, as go-mysql-server's replaceIdxSort does.
pub fn order_by_index(plan: Plan) -> Plan {
    match plan {
        Plan::Sort { input, keys } => match ordered(&input, &keys) {
            Some(input) => input,
            None => Plan::Sort { input, keys },
        },
        other => other,
    }
}

/// vector_of returns the elements of a vector value of a type that a vector index holds.
fn vector_of(value: &Value) -> Result<Vec<f32>> {
    match value {
        Value::Base(base) => match crate::types::base_type(base.type_oid).and_then(|t| t.vector) {
            Some(vector) => Ok(vector(&base.data)),
            None => Err(crate::error::PgError::internal("a vector search with a value that is not a vector")),
        },
        _ => Err(crate::error::PgError::internal("a vector search with a value that is not a vector")),
    }
}

/// routine_distance returns the distance that a vector index orders its keys by for a pgvector distance routine.
fn routine_distance(routine: &crate::routines::Routine) -> Option<prolly::Distance> {
    Some(match routine.name.as_str() {
        "l2_distance" => prolly::Distance::L2Squared,
        "cosine_distance" => prolly::Distance::Cosine,
        "vector_negative_inner_product" | "halfvec_negative_inner_product" => prolly::Distance::InnerProduct,
        "l1_distance" => prolly::Distance::L1,
        _ => return None,
    })
}

/// is_query_vector reports whether an expression can be the query vector of a vector search: it reads no row and
/// holds no NULL constant, as Doltgres' isRowIndependentQueryVector requires.
fn is_query_vector(e: &Expr) -> bool {
    let mut null = false;
    e.visit(&mut |e| null |= matches!(e, Expr::Const(Value::Null)));
    is_constant(e) && !null
}

/// nearest_scan returns the input of a sort with its table read through a vector index, when the sort orders the
/// table's rows by their distance from a query vector and a LIMIT keeps the first of them.
fn nearest_scan(sort: &Plan, limit: &Option<Expr>, offset: &Option<Expr>) -> Option<Plan> {
    let Plan::Sort { input, keys } = sort else { return None };
    let ([key], Plan::Project { input: scanned, exprs }) = (keys.as_slice(), input.as_ref()) else { return None };
    let (Plan::Scan(table, _), Expr::Column(i), false) = (scanned.as_ref(), &key.expr, key.descending) else {
        return None;
    };
    let Expr::Operator(_, routine, l, r) = exprs.get(*i)? else { return None };
    let distance = routine_distance(routine)?;
    let (column, query) = match (l.as_ref(), r.as_ref()) {
        (Expr::Column(c), query) if is_query_vector(query) => (*c, query),
        (query, Expr::Column(c)) if is_query_vector(query) => (*c, query),
        _ => return None,
    };
    let index = table.indexes.iter().position(|index| index.columns == [column] && index.vector == Some(distance))?;
    let nearest =
        Nearest { order: exprs[*i].clone(), query: query.clone(), limit: limit.clone(), offset: offset.clone() };
    let scan = IndexScan {
        table: Arc::new((**table).clone()),
        index: Some(index),
        ranges: Vec::new(),
        reverse: false,
        nearest: Some(nearest),
        needed: None,
        lookup_heavy: None,
        parameterized: None,
    };
    Some(Plan::Project { input: Box::new(Plan::IndexScan(Box::new(scan))), exprs: exprs.clone() })
}

/// prune tells the scans under a plan which columns the nodes above them read, so that a table scan decodes only
/// those columns and a scan of a secondary index that holds them skips the primary index.
pub fn prune(plan: &mut Plan) {
    prune_to(plan, None);
}

/// prune_to tells the scans under a plan which columns they need, given the columns of the plan's rows that the
/// nodes above it read, or None for every column.
fn prune_to(plan: &mut Plan, needed: Option<BTreeSet<usize>>) {
    let union = |a: Option<BTreeSet<usize>>, b: Option<BTreeSet<usize>>| Some(a?.union(&b?).copied().collect());
    match plan {
        Plan::IndexScan(scan) => {
            let tested = columns_read(scan.parameterized.iter());
            scan.needed = union(needed, tested).map(|n: BTreeSet<usize>| n.into_iter().collect());
        }
        Plan::BitmapHeapScan(scan) => scan.needed = needed.map(|n| n.into_iter().collect()),
        Plan::Scan(_, columns) => *columns = needed.map(|n| n.into_iter().collect()),
        Plan::Join { left, right, condition, lateral, method, .. } => {
            let width = left.width();
            let needed: Option<BTreeSet<usize>> = union(needed, columns_read(condition.iter()));
            let side = |right: bool| -> Option<BTreeSet<usize>> {
                needed.as_ref().map(|n: &BTreeSet<usize>| {
                    n.iter().filter(|&&c| (c >= width) == right).map(|&c| if right { c - width } else { c }).collect()
                })
            };
            let (left_needed, right_needed) = (side(false), side(true));
            if let crate::plan::JoinMethod::Lookup { scan, .. } = method {
                let (read, input) = match &**right {
                    Plan::Project { input, exprs } => match &right_needed {
                        Some(needed) => (columns_read(needed.iter().filter_map(|&i| exprs.get(i))), &**input),
                        None => (columns_read(exprs.iter()), &**input),
                    },
                    other => (right_needed.clone(), other),
                };
                let checked = match input {
                    Plan::Filter { predicate, .. } => columns_read([predicate]),
                    Plan::IndexScan(ranged) => Some(ranged.index_columns().into_iter().collect()),
                    _ => Some(BTreeSet::new()),
                };
                scan.needed = union(read, checked).map(|n| n.into_iter().collect());
            }
            prune_to(left, if *lateral { None } else { left_needed });
            prune_to(right, right_needed);
        }
        Plan::Distinct { input, keys: None } => prune_to(input, None),
        Plan::Window { input, calls } => {
            let width = input.width();
            let mut exprs: Vec<&Expr> = Vec::new();
            for call in calls.iter() {
                exprs.extend(&call.args);
                exprs.extend(&call.filter);
                exprs.extend(&call.partition);
                exprs.extend(call.order.iter().map(|k| &k.expr));
                if let Some(range) = &call.range {
                    exprs.push(&range.key);
                    exprs.extend(range.start.iter().chain(&range.end).map(|(e, _)| e));
                }
            }
            let below = needed.map(|n| n.into_iter().filter(|&c| c < width).collect());
            prune_to(input, union(below, columns_read(exprs)));
        }
        Plan::ProjectSet { input, functions, dropped } => {
            let width = input.width();
            let below: Option<BTreeSet<usize>> = needed.map(|n| n.into_iter().filter(|&c| c < width).collect());
            if let Some(below) = &below {
                *dropped = (0..width).filter(|c| !below.contains(c)).collect();
            }
            prune_to(input, union(below, columns_read(functions.iter())));
        }
        Plan::Once(input) => prune_to(input, needed),
        Plan::Filter { input, predicate } => {
            prune_to(input, union(needed, columns_read([&*predicate])));
            if let Plan::IndexScan(scan) = &mut **input
                && let Some(exact) = scan.lookup_heavy.take()
            {
                if !scan.covering() {
                    **input = Plan::Scan(Box::new((*scan.table).clone()), scan.needed.clone());
                } else if exact {
                    *plan = std::mem::replace(&mut **input, Plan::Values(Vec::new()));
                }
            }
        }
        Plan::Project { input, exprs } => {
            let read = match &needed {
                Some(needed) => columns_read(needed.iter().filter_map(|&i| exprs.get(i))),
                None => columns_read(exprs.iter()),
            };
            prune_to(input, read);
        }
        Plan::Sort { input, keys } => prune_to(input, union(needed, columns_read(keys.iter().map(|k| &k.expr)))),
        Plan::Limit { input, .. } => prune_to(input, needed),
        Plan::Distinct { input, keys: Some(keys) } => prune_to(input, union(needed, columns_read(keys.iter()))),
        Plan::Aggregate { input, groups, aggregates, .. } => {
            let mut exprs: Vec<&Expr> = groups.iter().collect();
            for call in aggregates.iter() {
                exprs.extend(&call.args);
                exprs.extend(&call.filter);
                exprs.extend(call.order.iter().map(|(e, _, _)| e));
            }
            prune_to(input, columns_read(exprs));
        }
        _ => {}
    }
}

/// columns_read returns the columns of the input row that expressions read, counting those that their subqueries read
/// as their enclosing row, or None when one reads the row in a way that the set cannot tell.
pub(crate) fn columns_read<'e>(exprs: impl IntoIterator<Item = &'e Expr>) -> Option<BTreeSet<usize>> {
    let mut columns = BTreeSet::new();
    let mut known = true;
    for expr in exprs {
        expr.visit(&mut |e| match e {
            Expr::Column(i) => {
                columns.insert(*i);
            }
            Expr::Exists(plan) | Expr::Scalar(plan) | Expr::ArraySubquery(plan, _) | Expr::AnySubquery(_, plan, _) => {
                known &= outer_reads(plan, 1, &mut columns);
            }
            Expr::SubPlan(_) | Expr::AlternativeSubPlan(_) => {
                for plan in e.subqueries() {
                    known &= outer_reads(plan, 1, &mut BTreeSet::new());
                }
            }
            Expr::InputColumn(_) | Expr::AggRef(_) | Expr::Default(_) => known = false,
            _ => {}
        });
    }
    known.then_some(columns)
}

/// outer_reads adds the columns of the enclosing row `depth` rows out that a plan's expressions read, reporting
/// whether it could see every expression of the plan.
pub(crate) fn outer_reads(plan: &Plan, depth: usize, out: &mut BTreeSet<usize>) -> bool {
    let mut known = true;
    let mut read = |exprs: &mut dyn Iterator<Item = &Expr>| {
        for expr in exprs {
            expr.visit(&mut |e| match e {
                Expr::Outer(d, i) if *d == depth => {
                    out.insert(*i);
                }
                e => {
                    for plan in e.subqueries() {
                        known &= outer_reads(plan, depth + 1, out);
                    }
                }
            });
        }
    };
    let inputs: Vec<(&Plan, usize)> = match plan {
        Plan::OneRow | Plan::Scan(..) | Plan::Catalog(_) | Plan::CatalogIndexScan(_) | Plan::WorkTable(..) => {
            Vec::new()
        }
        Plan::System(_) | Plan::QueryDiff(..) | Plan::XmlTable(_) | Plan::JsonTable(_) => return false,
        Plan::IndexScan(scan) => {
            if let Some(n) = &scan.nearest {
                read(&mut [&n.order, &n.query].into_iter().chain(&n.limit).chain(&n.offset));
            }
            read(&mut scan.parameterized.iter());
            Vec::new()
        }
        Plan::BitmapHeapScan(scan) => {
            read(&mut scan.recheck.iter().chain(scan.bitmap.conditions()));
            Vec::new()
        }
        Plan::Values(rows) => {
            read(&mut rows.iter().flatten());
            Vec::new()
        }
        Plan::Function { call, .. } => {
            read(&mut std::iter::once(call));
            Vec::new()
        }
        Plan::RowsFrom { calls, .. } => {
            read(&mut calls.iter());
            Vec::new()
        }
        Plan::Filter { input, predicate } => {
            read(&mut std::iter::once(predicate));
            vec![(input, depth)]
        }
        Plan::Project { input, exprs } => {
            read(&mut exprs.iter());
            vec![(input, depth)]
        }
        Plan::Join { left, right, condition, lateral, .. } => {
            read(&mut condition.iter());
            vec![(left, depth), (right, depth + usize::from(*lateral))]
        }
        Plan::Aggregate { input, groups, aggregates, .. } => {
            read(&mut groups.iter());
            for call in aggregates {
                read(&mut call.args.iter().chain(&call.filter).chain(call.order.iter().map(|(e, _, _)| e)));
            }
            vec![(input, depth)]
        }
        Plan::Sort { input, keys } => {
            read(&mut keys.iter().map(|k| &k.expr));
            vec![(input, depth)]
        }
        Plan::Distinct { input, keys } => {
            read(&mut keys.iter().flatten());
            vec![(input, depth)]
        }
        Plan::Limit { input, limit, offset } => {
            read(&mut limit.iter().chain(offset));
            vec![(input, depth)]
        }
        Plan::SetOp { left, right, .. } => vec![(left, depth), (right, depth)],
        Plan::Recursive { anchor, step, .. } => vec![(anchor, depth), (step, depth)],
        Plan::ProjectSet { input, functions, .. } => {
            read(&mut functions.iter());
            vec![(input, depth)]
        }
        Plan::Window { input, calls } => {
            for call in calls {
                read(
                    &mut call
                        .args
                        .iter()
                        .chain(&call.filter)
                        .chain(&call.partition)
                        .chain(call.order.iter().map(|k| &k.expr)),
                );
                if let Some(range) = &call.range {
                    read(&mut std::iter::once(&range.key).chain(range.start.iter().chain(&range.end).map(|(e, _)| e)));
                }
            }
            vec![(input, depth)]
        }
        Plan::Once(input) => vec![(input, depth)],
    };
    for (input, depth) in inputs {
        known &= outer_reads(input, depth, out);
    }
    known
}

/// nearest replaces a sort under a LIMIT with a search of a vector index when the sort orders a table's rows by their
/// distance from a query vector, as go-mysql-server's replaceIdxOrderByDistance does.
pub fn nearest(plan: Plan) -> Plan {
    match plan {
        Plan::Limit { input, limit: Some(limit), offset } => {
            let limit = Some(limit);
            match nearest_scan(&input, &limit, &offset) {
                Some(scan) => Plan::Limit { input: Box::new(scan), limit, offset },
                None => Plan::Limit { input, limit, offset },
            }
        }
        other => other,
    }
}
