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

//! Running plans: each node hands its rows one at a time to the node above it, so that rows flow through filters,
//! projections, and joins without being collected at every node, and a LIMIT stops reading its input early.

use std::borrow::Cow;
use std::cmp::Ordering;
use std::hash::{BuildHasher, Hash, Hasher};
use std::sync::Arc;

use doltdb::database::Database;

use crate::catalog::table::TableDef;
use crate::error::{Result, code};
use crate::expr::{CmpOp, Expr, compare_values};
use crate::functions::aggregate::{Accumulator, AggCall};
use crate::plan::{HashKey, JoinKind, JoinMethod, Plan, SetOp, SortKey, SubqueryRows};
use crate::query::Ctx;
use crate::types::Value;

/// Row is a row of values.
pub type Row = Vec<Value>;

/// Rows is a running plan node, which produces its rows one at a time.
pub trait Rows {
    /// next returns the node's next row, or None once it has no more.
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>>;

    /// next_into puts the node's next row in a buffer, reporting whether there was one. The buffer is empty or holds a
    /// row that this node put there, maybe with values taken out and left NULL, which lets a node reuse it and fill in
    /// only the values that change.
    fn next_into(&mut self, ctx: &mut Ctx<'_>, row: &mut Row) -> Result<bool> {
        match self.next(ctx)? {
            Some(next) => {
                *row = next;
                Ok(true)
            }
            None => Ok(false),
        }
    }

    /// skip passes over up to a number of rows, as an OFFSET discards them, returning how many it passed.
    fn skip(&mut self, ctx: &mut Ctx<'_>, count: usize) -> Result<usize> {
        let mut skipped = 0;
        while skipped < count && self.next(ctx)?.is_some() {
            skipped += 1;
        }
        Ok(skipped)
    }
}

/// Collected hands out rows that a node computed all at once.
struct Collected(std::vec::IntoIter<Row>);

impl Rows for Collected {
    fn next(&mut self, _: &mut Ctx<'_>) -> Result<Option<Row>> {
        Ok(self.0.next())
    }
}

/// collected returns a running node over rows computed all at once.
pub(crate) fn collected<'p>(rows: Vec<Row>) -> Box<dyn Rows + 'p> {
    Box::new(Collected(rows.into_iter()))
}

/// drain collects every remaining row of a running node.
pub(crate) fn drain(rows: &mut dyn Rows, ctx: &mut Ctx<'_>) -> Result<Vec<Row>> {
    let mut out = Vec::new();
    while let Some(row) = rows.next(ctx)? {
        out.push(row);
    }
    Ok(out)
}

/// TableWalk reads a table's rows in primary key order, decoding only the columns asked for.
pub(crate) struct TableWalk<'t> {
    table: &'t TableDef,
    items: prolly::Items,
    /// Which columns to decode, or None for every column.
    needed: Option<Vec<bool>>,
    /// A row of a keyless table to hand out again, with how many more times.
    repeat: Option<(Row, u64)>,
}

impl<'t> TableWalk<'t> {
    /// new starts a walk over the table's rows, decoding the needed columns or every column.
    pub(crate) fn new(db: &mut Database, table: &'t TableDef, needed: Option<&[usize]>) -> Result<TableWalk<'t>> {
        let root = Arc::new(prolly::Node::decode(table.table.primary_index.clone())?);
        let items = prolly::Items::first(db, root)?;
        let needed = needed.map(|columns| {
            let mut mask = vec![false; table.columns.len()];
            for &c in columns {
                if let Some(m) = mask.get_mut(c) {
                    *m = true;
                }
            }
            mask
        });
        Ok(TableWalk { table, items, needed, repeat: None })
    }

    /// next returns the table's next row.
    pub(crate) fn next(&mut self, db: &mut Database) -> Result<Option<Row>> {
        let mut row = Vec::new();
        Ok(self.next_into(db, &mut row)?.then_some(row))
    }

    /// next_into puts the table's next row in a buffer, as `Rows::next_into` describes.
    pub(crate) fn next_into(&mut self, db: &mut Database, row: &mut Row) -> Result<bool> {
        if let Some((repeated, remaining)) = self.repeat.as_mut() {
            row.clone_from(repeated);
            *remaining -= 1;
            if *remaining == 0 {
                self.repeat = None;
            }
            return Ok(true);
        }
        loop {
            let Some((key, value)) = self.items.current()? else { return Ok(false) };
            let cardinality = self.table.decode_columns_into(db, key, value, self.needed.as_deref(), row)?;
            self.items.advance(db)?;
            if cardinality > 1 {
                self.repeat = Some((row.clone(), cardinality - 1));
            }
            if cardinality > 0 {
                return Ok(true);
            }
        }
    }
}

impl Rows for TableWalk<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        TableWalk::next(self, ctx.db)
    }

    fn next_into(&mut self, ctx: &mut Ctx<'_>, row: &mut Row) -> Result<bool> {
        TableWalk::next_into(self, ctx.db, row)
    }

    fn skip(&mut self, ctx: &mut Ctx<'_>, count: usize) -> Result<usize> {
        if self.table.keyless() || self.repeat.is_some() {
            let mut skipped = 0;
            while skipped < count && TableWalk::next(self, ctx.db)?.is_some() {
                skipped += 1;
            }
            return Ok(skipped);
        }
        let mut skipped = 0;
        while skipped < count && self.items.current()?.is_some() {
            self.items.advance(ctx.db)?;
            skipped += 1;
        }
        Ok(skipped)
    }
}

/// FilterRows keeps the rows of its input that a predicate holds for.
struct FilterRows<'p> {
    input: Box<dyn Rows + 'p>,
    predicate: Cow<'p, Expr>,
}

impl Rows for FilterRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        let mut row = Vec::new();
        Ok(self.next_into(ctx, &mut row)?.then_some(row))
    }

    fn next_into(&mut self, ctx: &mut Ctx<'_>, row: &mut Row) -> Result<bool> {
        while self.input.next_into(ctx, row)? {
            if self.predicate.is_true(ctx, row)? {
                return Ok(true);
            }
        }
        Ok(false)
    }
}

/// ProjectRows computes expressions over each row of its input.
struct ProjectRows<'p> {
    input: Box<dyn Rows + 'p>,
    exprs: &'p [Expr],
    /// Whether each expression is a column that nothing else reads, whose value moves out of the input row.
    moves: Vec<bool>,
    /// The buffer that the input's rows arrive in.
    input_row: Row,
}

impl<'p> ProjectRows<'p> {
    /// new projects the input's rows.
    fn new(input: Box<dyn Rows + 'p>, exprs: &'p [Expr]) -> ProjectRows<'p> {
        let mut reads: Vec<usize> = Vec::new();
        for expr in exprs {
            expr.visit(&mut |e| {
                if let Expr::Column(i) = e {
                    if reads.len() <= *i {
                        reads.resize(i + 1, 0);
                    }
                    reads[*i] += 1;
                }
            });
        }
        let moves = exprs.iter().map(|e| matches!(e, Expr::Column(i) if reads[*i] == 1)).collect();
        ProjectRows { input, exprs, moves, input_row: Vec::new() }
    }
}

impl Rows for ProjectRows<'_> {
    fn skip(&mut self, ctx: &mut Ctx<'_>, count: usize) -> Result<usize> {
        match self.exprs.iter().all(|e| matches!(e, Expr::Column(_) | Expr::Const(_))) {
            true => self.input.skip(ctx, count),
            false => {
                let mut skipped = 0;
                while skipped < count && self.next(ctx)?.is_some() {
                    skipped += 1;
                }
                Ok(skipped)
            }
        }
    }

    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        let mut out = Vec::with_capacity(self.exprs.len());
        Ok(self.next_into(ctx, &mut out)?.then_some(out))
    }

    fn next_into(&mut self, ctx: &mut Ctx<'_>, out: &mut Row) -> Result<bool> {
        if !self.input.next_into(ctx, &mut self.input_row)? {
            return Ok(false);
        }
        let row = &mut self.input_row;
        out.resize(self.exprs.len(), Value::Null);
        for ((expr, &moves), value) in self.exprs.iter().zip(&self.moves).zip(out.iter_mut()) {
            if !moves {
                *value = match expr {
                    Expr::Column(i) => row[*i].clone(),
                    expr => expr.eval(ctx, row)?,
                };
            }
        }
        for ((expr, &moves), value) in self.exprs.iter().zip(&self.moves).zip(out.iter_mut()) {
            if let (true, Expr::Column(i)) = (moves, expr) {
                std::mem::swap(&mut row[*i], value);
            }
        }
        Ok(true)
    }
}

/// LimitRows skips the first rows of its input and then hands out at most a number of rows.
struct LimitRows<'p> {
    input: Box<dyn Rows + 'p>,
    skip: usize,
    remaining: Option<usize>,
}

impl Rows for LimitRows<'_> {
    fn next_into(&mut self, ctx: &mut Ctx<'_>, row: &mut Row) -> Result<bool> {
        if self.remaining == Some(0) {
            return Ok(false);
        }
        if self.skip > 0 {
            let skipped = self.input.skip(ctx, self.skip)?;
            if skipped < self.skip {
                self.skip = 0;
                self.remaining = Some(0);
                return Ok(false);
            }
            self.skip = 0;
        }
        let found = self.input.next_into(ctx, row)?;
        if let Some(remaining) = self.remaining.as_mut() {
            *remaining = if found { *remaining - 1 } else { 0 };
        }
        Ok(found)
    }

    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        if self.remaining == Some(0) {
            return Ok(None);
        }
        if self.skip > 0 {
            let skipped = self.input.skip(ctx, self.skip)?;
            if skipped < self.skip {
                self.skip = 0;
                self.remaining = Some(0);
                return Ok(None);
            }
            self.skip = 0;
        }
        let row = self.input.next(ctx)?;
        if let Some(remaining) = self.remaining.as_mut() {
            *remaining = if row.is_some() { *remaining - 1 } else { 0 };
        }
        Ok(row)
    }
}

/// ChainRows hands out the rows of each input in turn.
struct ChainRows<'p> {
    inputs: Vec<&'p Plan>,
    current: Option<Box<dyn Rows + 'p>>,
}

impl Rows for ChainRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        loop {
            if let Some(current) = self.current.as_mut()
                && let Some(row) = current.next(ctx)?
            {
                return Ok(Some(row));
            }
            if self.inputs.is_empty() {
                return Ok(None);
            }
            self.current = Some(self.inputs.remove(0).open(ctx)?);
        }
    }
}

/// SharedRows hands out the rows of a plan that ran once for the whole statement.
struct SharedRows {
    rows: Arc<SubqueryRows>,
    next: usize,
}

impl Rows for SharedRows {
    fn next(&mut self, _: &mut Ctx<'_>) -> Result<Option<Row>> {
        let row = self.rows.rows.get(self.next).cloned();
        self.next += 1;
        Ok(row)
    }
}

/// OnceFilterRows hands out the rows of a `Once` plan that a filter keeps for the enclosing row, checking only the
/// rows whose values match it in the filter's equality conditions when it has some.
struct OnceFilterRows {
    shared: Arc<SubqueryRows>,
    /// The bucket of rows that match the enclosing row, or None to check every row.
    bucket: Option<usize>,
    position: usize,
}

impl OnceFilterRows {
    /// open finds the rows of the `Once` input that may match the enclosing row.
    fn open(ctx: &mut Ctx<'_>, input: &Plan, predicate: &Expr) -> Result<OnceFilterRows> {
        let key = input as *const Plan as usize;
        let shared = match ctx.once.as_ref().and_then(|once| once.get(&key)).cloned() {
            Some(shared) => shared,
            None => {
                // Index the rows as the scan reads them, rather than walking them again.
                let Plan::Once(scanned) = input else {
                    return Err(crate::error::PgError::internal("a filter without its rows"));
                };
                let (inner, mut index) = crate::plan::index_parts(ctx, predicate);
                let mut source = scanned.open(ctx)?;
                let mut rows = Vec::new();
                while let Some(row) = source.next(ctx)? {
                    crate::plan::index_row(ctx, &inner, &mut index, &row, rows.len());
                    rows.push(row);
                }
                drop(source);
                let shared = Arc::new(SubqueryRows::indexed(rows, index));
                if let Some(once) = ctx.once.as_mut() {
                    once.insert(key, shared.clone());
                }
                shared
            }
        };
        let index = match shared.index.get() {
            Some(index) => index,
            None => {
                let index = crate::plan::row_index(ctx, &shared.rows, predicate);
                let _ = shared.index.set(index);
                shared.index.get().expect("an index")
            }
        };
        let mut bucket = None;
        if let Some((table, _)) = &index.table {
            let empty = OnceFilterRows { shared: shared.clone(), bucket: None, position: usize::MAX };
            let mut key: smallvec::SmallVec<[Option<HashKey>; 2]> = smallvec::SmallVec::new();
            for e in &index.outer {
                match e.eval(ctx, &[])? {
                    Value::Null => return Ok(empty),
                    value => key.push(HashKey::of(value)),
                }
            }
            if let Some(key) = key.into_iter().collect::<Option<crate::plan::JoinKey>>() {
                match table.get(&key) {
                    Some(&b) => bucket = Some(b),
                    None => return Ok(empty),
                }
            }
        }
        Ok(OnceFilterRows { shared, bucket, position: 0 })
    }
}

impl Rows for OnceFilterRows {
    fn next_into(&mut self, ctx: &mut Ctx<'_>, row: &mut Row) -> Result<bool> {
        match self.matching(ctx)? {
            Some(j) => {
                row.clone_from(&self.shared.rows[j]);
                Ok(true)
            }
            None => Ok(false),
        }
    }

    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        Ok(self.matching(ctx)?.map(|j| self.shared.rows[j].clone()))
    }
}

impl OnceFilterRows {
    /// matching returns the position of the next row that the filter keeps.
    fn matching(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<usize>> {
        let index = self.shared.index.get().expect("an index");
        loop {
            let (j, predicate) = match self.bucket {
                Some(b) => match index.table.as_ref().expect("a table").1[b].get(self.position) {
                    Some(&j) => (j, &index.residual),
                    None => return Ok(None),
                },
                None if self.position < self.shared.rows.len() => (self.position, &index.predicate),
                None => return Ok(None),
            };
            self.position += 1;
            if predicate.is_true(ctx, &self.shared.rows[j])? {
                return Ok(Some(j));
            }
        }
    }
}

/// DistinctRows hands out the first row of each run of rows with equal keys, or each row unlike every earlier one.
struct DistinctRows<'p> {
    input: Box<dyn Rows + 'p>,
    keys: Option<&'p [Expr]>,
    previous: Option<Row>,
    seen: Groups,
}

impl Rows for DistinctRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        while let Some(row) = self.input.next(ctx)? {
            match self.keys {
                Some(keys) => {
                    let key = keys.iter().map(|k| k.eval(ctx, &row)).collect::<Result<Vec<_>>>()?;
                    let first = self.previous.as_ref().is_none_or(|p| !crate::plan::rows_equal(p, &key));
                    self.previous = Some(key);
                    if first {
                        return Ok(Some(row));
                    }
                }
                None => {
                    if self.seen.insert(&row).1 {
                        return Ok(Some(row));
                    }
                }
            }
        }
        Ok(None)
    }
}

/// ProjectSetRows hands out, for each input row, the rows of its set-returning calls side by side after its values.
struct ProjectSetRows<'p> {
    input: Box<dyn Rows + 'p>,
    functions: &'p [Expr],
    dropped: &'p [usize],
    pending: std::vec::IntoIter<Row>,
}

impl Rows for ProjectSetRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        loop {
            if let Some(row) = self.pending.next() {
                return Ok(Some(row));
            }
            let Some(mut row) = self.input.next(ctx)? else { return Ok(None) };
            let mut columns = Vec::with_capacity(self.functions.len());
            for function in self.functions {
                columns.push(crate::plan::set_rows(ctx, function, &row)?);
            }
            for &i in self.dropped {
                row[i] = Value::Null;
            }
            let count = columns.iter().map(Vec::len).max().unwrap_or(0);
            let mut out = Vec::with_capacity(count);
            for i in 0..count {
                let mut new_row = row.clone();
                new_row.extend(columns.iter().map(|c| c.get(i).cloned().unwrap_or(Value::Null)));
                out.push(new_row);
            }
            self.pending = out.into_iter();
        }
    }
}

/// RecursiveRows hands out a recursive WITH query's rows: its non-recursive term's, then each round of its recursive
/// term's over the previous round's rows, until a round adds none.
struct RecursiveRows<'p> {
    work_table: usize,
    step: &'p Plan,
    all: bool,
    /// The rows of every round so far, which a query without ALL drops repeats of.
    seen: Groups,
    /// The rows of the round being handed out, which the next round reads.
    round: Vec<Row>,
    next: usize,
}

impl RecursiveRows<'_> {
    /// keep drops the rows of a round that repeat a row of the round or of an earlier one, without ALL.
    fn keep(&mut self, rows: Vec<Row>) -> Vec<Row> {
        match self.all {
            true => rows,
            false => rows.into_iter().filter(|r| self.seen.insert(r).1).collect(),
        }
    }
}

impl Rows for RecursiveRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        loop {
            if let Some(row) = self.round.get(self.next) {
                self.next += 1;
                return Ok(Some(row.clone()));
            }
            if self.round.is_empty() {
                return Ok(None);
            }
            let previous = ctx.work_tables.insert(self.work_table, std::mem::take(&mut self.round));
            let rows = self.step.run(ctx);
            match previous {
                Some(previous) => ctx.work_tables.insert(self.work_table, previous),
                None => ctx.work_tables.remove(&self.work_table),
            };
            self.round = self.keep(rows?);
            self.next = 0;
        }
    }
}

/// JoinRows pairs each row of its left input with the matching rows of its right input, finding candidates through a
/// hash table of the right rows' keys when the condition has equalities between the sides. When those equalities give
/// the right table's whole primary key, it looks each left row's match up instead, until the lookups would cost more
/// than reading the right input once.
struct JoinRows<'p> {
    left: Box<dyn Rows + 'p>,
    right_plan: &'p Plan,
    lookup: Option<Lookup<'p>>,
    /// The right row that the left row being joined found by its primary key.
    found: Option<Row>,
    right: Vec<Row>,
    kind: JoinKind,
    condition: Option<&'p Expr>,
    left_width: usize,
    right_width: usize,
    hash: Option<JoinHash>,
    right_matched: Vec<bool>,
    /// The left row being joined, the right rows it may match, the position among them, and whether one matched.
    current: Option<(Row, Candidates, usize, bool)>,
    left_done: bool,
    /// The next right row to check for a match, once the left rows are done.
    unmatched: usize,
}

/// fixed_value returns the constant that a filter's equality sets a column to, if it has one.
fn fixed_value(filter: &Expr, column: usize) -> Option<&Expr> {
    crate::indexscan::conjuncts(filter).into_iter().find_map(|c| match c {
        Expr::Compare(CmpOp::Eq, a, b) if **a == Expr::Column(column) && crate::indexscan::is_constant(b) => Some(&**b),
        Expr::Compare(CmpOp::Eq, a, b) if **b == Expr::Column(column) && crate::indexscan::is_constant(a) => Some(&**a),
        _ => None,
    })
}

/// Candidates are the right rows that a left row may match: all of them, one bucket of the hash table, or the row its
/// primary key lookup found.
#[derive(Clone, Copy)]
enum Candidates {
    All,
    Bucket(usize),
    Found,
    None,
}

/// Lookup finds the right rows of a join by the right table's primary key, whose every column an equality of the join
/// condition sets to a value of the left row, or the right input's filter sets to a constant.
struct Lookup<'p> {
    table: &'p TableDef,
    /// Which right columns to decode, or None for every column.
    needed: Option<Vec<bool>>,
    /// The right input's own filter.
    filter: Option<&'p Expr>,
    /// The left expressions that give the primary key's columns, in key order.
    keys: Vec<Expr>,
    root: Arc<prolly::Node>,
    walk: Option<prolly::Items>,
    /// How many more lookups to make before reading the right input instead.
    budget: usize,
}

impl<'p> Lookup<'p> {
    /// new returns the lookup that a join's right input allows: a scan of a keyed table, maybe filtered, or a filtered
    /// index scan of one, whose primary key the condition's equalities and the filter's constants give in full with
    /// columns of types the lookup can encode.
    fn new(right: &'p Plan, condition: Option<&Expr>, left_width: usize) -> Result<Option<Lookup<'p>>> {
        Lookup::of_side(right, condition, left_width, true)
    }

    /// finds_whole_keys reports whether a planned lookup in an index is one by the right table's whole primary key that
    /// `new` makes, which reuses its walk of the index from one lookup to the next.
    fn finds_whole_keys(
        scan: &crate::indexscan::IndexScan,
        keys: &[Expr],
        right: &Plan,
        condition: Option<&Expr>,
        left_width: usize,
    ) -> Result<bool> {
        Ok(scan.index.is_none()
            && keys.len() == scan.table.key_columns.len()
            && Lookup::new(right, condition, left_width)?.is_some())
    }

    /// of_side is `new` for the right input, or for the left input when `right` is false, which a join then finds by
    /// the values of each right row.
    fn of_side(
        input: &'p Plan,
        condition: Option<&Expr>,
        left_width: usize,
        right: bool,
    ) -> Result<Option<Lookup<'p>>> {
        let (scan, filter) = match input {
            Plan::Filter { input, predicate } => (&**input, Some(predicate)),
            other => (other, None),
        };
        let (table, needed) = match (scan, filter) {
            (Plan::Scan(table, needed), _) => (table, needed),
            (Plan::IndexScan(index), Some(_)) => (&index.table, &index.needed),
            _ => return Ok(None),
        };
        let Some(condition) = condition else { return Ok(None) };
        if table.keyless() {
            return Ok(None);
        }
        let (left_keys, right_keys) = crate::plan::join_keys(condition, left_width);
        let (left_keys, right_keys) = if right { (left_keys, right_keys) } else { (right_keys, left_keys) };
        let mut keys = Vec::with_capacity(table.key_columns.len());
        for &c in &table.key_columns {
            let column = &table.columns[c];
            if !lookup_type(column.ty.oid) || crate::storage::is_adaptive(column.encoding) {
                return Ok(None);
            }
            match right_keys.iter().position(|k| *k == Expr::Column(c)) {
                Some(i) => keys.push(left_keys[i].clone()),
                None => match filter.and_then(|f| fixed_value(f, c)) {
                    Some(value) => keys.push(value.clone()),
                    None => return Ok(None),
                },
            }
        }
        let root = Arc::new(prolly::Node::decode(table.table.primary_index.clone())?);
        let budget = (root.tree_count() / 8).max(32) as usize;
        let needed = needed.as_ref().map(|columns| {
            let mut mask = vec![false; table.columns.len()];
            for &c in columns {
                if let Some(m) = mask.get_mut(c) {
                    *m = true;
                }
            }
            mask
        });
        Ok(Some(Lookup { table, needed, filter, keys, root, walk: None, budget }))
    }

    /// find returns the right row that a left row's key values find, or None when there is none or its filter
    /// rejects it, or Err(None) when a value is not one the lookup can encode.
    fn find(
        &mut self,
        ctx: &mut Ctx<'_>,
        left: &[Value],
    ) -> std::result::Result<Option<Row>, Option<crate::error::PgError>> {
        let mut fields: smallvec::SmallVec<[Option<Vec<u8>>; 4]> = smallvec::SmallVec::new();
        for (expr, &c) in self.keys.iter().zip(&self.table.key_columns) {
            let column = &self.table.columns[c];
            let value = match expr.eval(ctx, left).map_err(Some)? {
                Value::Null => return Ok(None),
                value => match lookup_value(value, column.ty.oid) {
                    Some(Some(value)) => value,
                    Some(None) => return Ok(None),
                    None => return Err(None),
                },
            };
            fields.push(crate::storage::encode_field(&value, column.encoding, column.ty).map_err(|_| None)?);
        }
        let key =
            prolly::val::build_tuple(&fields.iter().map(Option::as_deref).collect::<smallvec::SmallVec<[_; 4]>>());
        let table = self.table;
        let compare = |a: &[u8], b: &[u8]| table.compare_keys(a, b);
        let walk = match &mut self.walk {
            Some(walk) => {
                walk.seek(ctx.db, &key, &compare).map_err(|e| Some(e.into()))?;
                walk
            }
            None => self
                .walk
                .insert(prolly::Items::at_key(ctx.db, self.root.clone(), &key, &compare).map_err(|e| Some(e.into()))?),
        };
        let mut row = Vec::new();
        match walk.current().map_err(|e| Some(e.into()))? {
            Some((k, value)) if compare(k, &key) == std::cmp::Ordering::Equal => {
                table.decode_columns_into(ctx.db, k, value, self.needed.as_deref(), &mut row).map_err(Some)?;
            }
            _ => return Ok(None),
        }
        match self.filter {
            Some(filter) if !filter.is_true(ctx, &row).map_err(Some)? => Ok(None),
            _ => Ok(Some(row)),
        }
    }
}

/// lookup_type reports whether a join can look rows up by a primary key column of the type, whose values encode the
/// same whenever they are equal.
pub(crate) fn lookup_type(type_oid: u32) -> bool {
    use crate::oid;
    matches!(
        type_oid,
        oid::INT2
            | oid::INT4
            | oid::INT8
            | oid::BOOL
            | oid::DATE
            | oid::TIMESTAMP
            | oid::TIMESTAMPTZ
            | oid::UUID
            | oid::TEXT
            | oid::VARCHAR
    )
}

/// lookup_value converts a left value to the type of a primary key column, returning Some(None) for an integer out of
/// the column's range, which no row has, and None for a value the lookup cannot use.
pub(crate) fn lookup_value(value: Value, type_oid: u32) -> Option<Option<Value>> {
    use crate::oid;
    let integer = match value {
        Value::Int2(i) => Some(i as i64),
        Value::Int4(i) => Some(i as i64),
        Value::Int8(i) => Some(i),
        _ => None,
    };
    Some(match (type_oid, integer, value) {
        (oid::INT2, Some(i), _) => i16::try_from(i).ok().map(Value::Int2),
        (oid::INT4, Some(i), _) => i32::try_from(i).ok().map(Value::Int4),
        (oid::INT8, Some(i), _) => Some(Value::Int8(i)),
        (oid::BOOL, _, v @ Value::Bool(_)) => Some(v),
        (oid::DATE, _, v @ Value::Date(_)) => Some(v),
        (oid::TIMESTAMP, _, v @ Value::Timestamp(_)) => Some(v),
        (oid::TIMESTAMPTZ, _, v @ Value::TimestampTz(_)) => Some(v),
        (oid::UUID, _, v @ Value::Uuid(_)) => Some(v),
        (oid::TEXT | oid::VARCHAR, _, v @ Value::Text(_)) => Some(v),
        _ => return None,
    })
}

/// JoinHash finds right rows by the values of the left row's side of the join's equalities.
struct JoinHash {
    left_keys: Vec<Expr>,
    /// The kind of key each right key holds, which a left key must share for the table to answer it.
    kinds: Vec<Option<std::mem::Discriminant<HashKey>>>,
    /// The hashed rows grouped by bucket in their own order, where bucket `b` is `rows[starts[b]..starts[b + 1]]`.
    rows: Vec<usize>,
    starts: Vec<usize>,
    table: crate::plan::KeyMap<usize>,
    /// Whether the condition is only the hashed equalities, so that every row of a bucket matches.
    exact: bool,
}

impl<'p> JoinRows<'p> {
    /// open starts a join of the inputs, reading the right input's rows unless `lookups` lets it look them up.
    fn open(
        ctx: &mut Ctx<'_>,
        left: &'p Plan,
        right: &'p Plan,
        kind: JoinKind,
        condition: Option<&'p Expr>,
        lookups: bool,
    ) -> Result<JoinRows<'p>> {
        let (left_width, right_width) = (left.width(), right.width());
        let left_rows = left.open(ctx)?;
        let lookup = match kind {
            JoinKind::Inner | JoinKind::Left | JoinKind::Anti | JoinKind::Semi if lookups => {
                Lookup::new(right, condition, left_width)?
            }
            _ => None,
        };
        let right_rows = if lookup.is_some() { Vec::new() } else { right.run(ctx)? };
        let hash = condition.and_then(|c| JoinHash::build(ctx, c, left_width, &right_rows, true));
        Ok(JoinRows {
            left: left_rows,
            right_plan: right,
            lookup,
            found: None,
            right_matched: vec![false; right_rows.len()],
            right: right_rows,
            kind,
            condition,
            left_width,
            right_width,
            hash,
            current: None,
            left_done: false,
            unmatched: 0,
        })
    }

    /// read_right stops looking rows up and reads the right input whole, hashing it, returning the candidates of the
    /// left row being joined.
    fn read_right(&mut self, ctx: &mut Ctx<'_>, left: &[Value]) -> Result<Candidates> {
        self.lookup = None;
        self.right = self.right_plan.run(ctx)?;
        self.right_matched = vec![false; self.right.len()];
        self.hash = self.condition.and_then(|c| JoinHash::build(ctx, c, self.left_width, &self.right, true));
        Ok(self.candidates(ctx, left))
    }

    /// candidates returns the right rows that a left row may match.
    fn candidates(&self, ctx: &mut Ctx<'_>, row: &[Value]) -> Candidates {
        match &self.hash {
            Some(hash) => hash.candidates(ctx, row),
            None => Candidates::All,
        }
    }
}

impl JoinHash {
    /// candidates returns the hashed rows that a row of the other input may match, by its side of the equalities.
    fn candidates(&self, ctx: &mut Ctx<'_>, row: &[Value]) -> Candidates {
        let mut key = crate::plan::JoinKey::with_capacity(self.left_keys.len());
        for (expr, kind) in self.left_keys.iter().zip(&self.kinds) {
            match expr.eval(ctx, row) {
                Ok(Value::Null) => return Candidates::None,
                Ok(value) => match HashKey::of(value) {
                    Some(k) if kind.is_none_or(|kind| kind == std::mem::discriminant(&k)) => key.push(k),
                    _ => return Candidates::All,
                },
                Err(_) => return Candidates::All,
            }
        }
        match self.table.get(&key) {
            Some(&bucket) => Candidates::Bucket(bucket),
            None => Candidates::None,
        }
    }

    /// bucket returns the hashed rows of a bucket.
    fn bucket(&self, bucket: usize) -> &[usize] {
        &self.rows[self.starts[bucket]..self.starts[bucket + 1]]
    }

    /// build hashes the rows of one side by their side of the equality conditions between the two sides, the right
    /// side's rows when `right` is set and the left side's otherwise, or returns None when the condition has none or a
    /// key is not one it can hash.
    fn build(ctx: &mut Ctx<'_>, condition: &Expr, width: usize, rows: &[Row], right: bool) -> Option<JoinHash> {
        let (left_keys, right_keys) = crate::plan::join_keys(condition, width);
        let exact = crate::indexscan::conjuncts(condition).len() == left_keys.len();
        let (probe, build) = if right { (left_keys, right_keys) } else { (right_keys, left_keys) };
        JoinHash::of_keys(ctx, probe, build, rows, exact)
    }

    /// of_keys hashes rows by the build side's expressions of the join's equalities, which the probe side's
    /// expressions over the other input's rows look up.
    fn of_keys(
        ctx: &mut Ctx<'_>,
        left_keys: Vec<Expr>,
        right_keys: Vec<Expr>,
        right_rows: &[Row],
        exact: bool,
    ) -> Option<JoinHash> {
        if left_keys.is_empty() {
            return None;
        }
        let mut kinds = vec![None; right_keys.len()];
        let mut of_row = vec![usize::MAX; right_rows.len()];
        let mut sizes: Vec<usize> = Vec::new();
        let mut table: crate::plan::KeyMap<usize> = Default::default();
        'rows: for (j, row) in right_rows.iter().enumerate() {
            let mut key = crate::plan::JoinKey::with_capacity(right_keys.len());
            for (e, kind) in right_keys.iter().zip(kinds.iter_mut()) {
                match e.eval(ctx, row).ok()? {
                    Value::Null => continue 'rows,
                    value => {
                        let k = HashKey::of(value)?;
                        let d = std::mem::discriminant(&k);
                        if kind.is_some_and(|kind| kind != d) {
                            return None;
                        }
                        *kind = Some(d);
                        key.push(k);
                    }
                }
            }
            let bucket = *table.entry(key).or_insert_with(|| {
                sizes.push(0);
                sizes.len() - 1
            });
            sizes[bucket] += 1;
            of_row[j] = bucket;
        }
        let mut starts = Vec::with_capacity(sizes.len() + 1);
        let mut total = 0;
        for size in &sizes {
            starts.push(total);
            total += size;
        }
        starts.push(total);
        let (mut next, mut rows) = (starts.clone(), vec![0; total]);
        for (j, &bucket) in of_row.iter().enumerate().filter(|(_, b)| **b != usize::MAX) {
            rows[next[bucket]] = j;
            next[bucket] += 1;
        }
        Some(JoinHash { left_keys, kinds, rows, starts, table, exact })
    }
}

impl Rows for JoinRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        loop {
            if let Some((left, candidates, position, matched)) = self.current.as_mut() {
                if let Candidates::Found = *candidates {
                    *candidates = Candidates::None;
                    if let Some(found) = self.found.take() {
                        let mut row = Vec::with_capacity(left.len() + found.len());
                        row.extend_from_slice(left);
                        row.extend(found);
                        if self.condition.map_or(Ok(true), |c| c.is_true(ctx, &row))? {
                            *matched = true;
                            if !self.kind.tests_matches() {
                                return Ok(Some(row));
                            }
                        }
                    }
                }
                let (bucket, exact): (&[usize], bool) = match (*candidates, &self.hash) {
                    (Candidates::Bucket(b), Some(hash)) => (hash.bucket(b), hash.exact),
                    _ => (&[], false),
                };
                while !*matched || !self.kind.tests_matches() {
                    let j = match *candidates {
                        Candidates::All if *position < self.right.len() => *position,
                        Candidates::Bucket(_) if *position < bucket.len() => bucket[*position],
                        _ => break,
                    };
                    *position += 1;
                    if exact && self.kind.tests_matches() {
                        *matched = true;
                        break;
                    }
                    let mut row = Vec::with_capacity(left.len() + self.right[j].len());
                    row.extend_from_slice(left);
                    row.extend_from_slice(&self.right[j]);
                    if exact || self.condition.map_or(Ok(true), |c| c.is_true(ctx, &row))? {
                        *matched = true;
                        self.right_matched[j] = true;
                        if !self.kind.tests_matches() {
                            return Ok(Some(row));
                        }
                    }
                }
                let (mut left, _, _, matched) = self.current.take().expect("a current left row");
                if (!matched && matches!(self.kind, JoinKind::Left | JoinKind::Full | JoinKind::Anti))
                    || (matched && self.kind == JoinKind::Semi)
                {
                    left.extend(std::iter::repeat_n(Value::Null, self.right_width));
                    return Ok(Some(left));
                }
            }
            if self.left_done {
                if !matches!(self.kind, JoinKind::Right | JoinKind::Full) {
                    return Ok(None);
                }
                while self.unmatched < self.right.len() {
                    let j = self.unmatched;
                    self.unmatched += 1;
                    if !self.right_matched[j] {
                        let mut row = vec![Value::Null; self.left_width];
                        row.extend_from_slice(&self.right[j]);
                        return Ok(Some(row));
                    }
                }
                return Ok(None);
            }
            match self.left.next(ctx)? {
                Some(left) => {
                    let candidates = match self.lookup.as_mut() {
                        Some(lookup) if lookup.budget > 0 => {
                            lookup.budget -= 1;
                            match lookup.find(ctx, &left) {
                                Ok(found) => {
                                    self.found = found;
                                    Candidates::Found
                                }
                                Err(Some(err)) => return Err(err),
                                Err(None) => self.read_right(ctx, &left)?,
                            }
                        }
                        Some(_) => self.read_right(ctx, &left)?,
                        None => self.candidates(ctx, &left),
                    };
                    self.current = Some((left, candidates, 0, false));
                }
                None => self.left_done = true,
            }
        }
    }
}

/// AntiRows runs an anti join the other way around when the left input's table is no larger than the right input's:
/// it hashes the left rows, reads the right input marking the left rows each right row matches, and then hands out
/// the unmarked left rows padded with NULLs.
struct AntiRows<'p> {
    left: Vec<Row>,
    matched: Vec<bool>,
    /// The right input, until it has been read.
    right: Option<Box<dyn Rows + 'p>>,
    hash: JoinHash,
    condition: Option<&'p Expr>,
    right_width: usize,
    next: usize,
}

impl<'p> AntiRows<'p> {
    /// open returns the anti join of the inputs when both scan tables, the left table holds no more rows than the
    /// right one, and the condition has equalities to hash the left rows by.
    fn open(
        ctx: &mut Ctx<'_>,
        left: &'p Plan,
        right: &'p Plan,
        condition: Option<&'p Expr>,
    ) -> Result<Option<AntiRows<'p>>> {
        let (Some(condition), Some(left_rows), Some(right_rows)) = (condition, table_rows(left)?, table_rows(right)?)
        else {
            return Ok(None);
        };
        if left_rows > right_rows || crate::plan::join_keys(condition, left.width()).0.is_empty() {
            return Ok(None);
        }
        let rows = left.run(ctx)?;
        let Some(hash) = JoinHash::build(ctx, condition, left.width(), &rows, false) else { return Ok(None) };
        Ok(Some(AntiRows {
            matched: vec![false; rows.len()],
            left: rows,
            right: Some(right.open(ctx)?),
            hash,
            condition: Some(condition),
            right_width: right.width(),
            next: 0,
        }))
    }
}

impl Rows for AntiRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        if let Some(mut right) = self.right.take() {
            let (mut r, mut row) = (Vec::new(), Vec::new());
            while right.next_into(ctx, &mut r)? {
                let candidates = self.hash.candidates(ctx, &r);
                let count = match candidates {
                    Candidates::Bucket(b) => self.hash.bucket(b).len(),
                    Candidates::All => self.left.len(),
                    Candidates::Found | Candidates::None => 0,
                };
                for k in 0..count {
                    let i = match candidates {
                        Candidates::Bucket(b) => self.hash.bucket(b)[k],
                        _ => k,
                    };
                    if self.matched[i] {
                        continue;
                    }
                    if self.hash.exact && matches!(candidates, Candidates::Bucket(_)) {
                        self.matched[i] = true;
                        continue;
                    }
                    row.clear();
                    row.extend_from_slice(&self.left[i]);
                    row.extend_from_slice(&r);
                    if self.condition.map_or(Ok(true), |c| c.is_true(ctx, &row))? {
                        self.matched[i] = true;
                    }
                }
            }
        }
        while self.next < self.left.len() {
            let i = self.next;
            self.next += 1;
            if !self.matched[i] {
                let mut row = std::mem::take(&mut self.left[i]);
                row.extend(std::iter::repeat_n(Value::Null, self.right_width));
                return Ok(Some(row));
            }
        }
        Ok(None)
    }
}

/// table_rows returns how many rows the table under a plan holds, when the plan scans one.
fn table_rows(plan: &Plan) -> Result<Option<u64>> {
    let table = match plan {
        Plan::Scan(table, _) => table,
        Plan::IndexScan(scan) => &scan.table,
        Plan::Filter { input, .. } => return table_rows(input),
        _ => return Ok(None),
    };
    Ok(Some(prolly::Node::decode(table.table.primary_index.clone())?.tree_count()))
}

/// ProbeRows runs an inner join the other way around: it reads its right input in order and finds each right row's
/// matches among the left input's rows by the left table's primary key, until the lookups would cost more than
/// reading the left input once, and then through a hash table of the left rows.
struct ProbeRows<'p> {
    right: Box<dyn Rows + 'p>,
    left_plan: &'p Plan,
    lookup: Option<Lookup<'p>>,
    condition: Option<&'p Expr>,
    left_width: usize,
    left: Vec<Row>,
    hash: Option<JoinHash>,
    /// The right row being joined, the left rows it may match, and the position among them.
    current: Option<(Row, Candidates, usize)>,
    found: Option<Row>,
}

impl<'p> ProbeRows<'p> {
    /// open returns the probing join of the inputs when the condition gives the whole primary key of the left
    /// input's table, and the right input's table offers no such lookup.
    fn open(
        ctx: &mut Ctx<'_>,
        left: &'p Plan,
        right: &'p Plan,
        condition: Option<&'p Expr>,
    ) -> Result<Option<ProbeRows<'p>>> {
        let left_width = left.width();
        if Lookup::new(right, condition, left_width)?.is_some() {
            return Ok(None);
        }
        let Some(lookup) = Lookup::of_side(left, condition, left_width, false)? else { return Ok(None) };
        Ok(Some(ProbeRows {
            right: right.open(ctx)?,
            left_plan: left,
            lookup: Some(lookup),
            condition,
            left_width,
            left: Vec::new(),
            hash: None,
            current: None,
            found: None,
        }))
    }

    /// read_left stops looking rows up and reads the left input whole, hashing it, returning the candidates of the
    /// right row being joined.
    fn read_left(&mut self, ctx: &mut Ctx<'_>, right: &[Value]) -> Result<Candidates> {
        self.lookup = None;
        self.left = self.left_plan.run(ctx)?;
        if let Some(condition) = self.condition {
            self.hash = JoinHash::build(ctx, condition, self.left_width, &self.left, false);
        }
        Ok(self.candidates(ctx, right))
    }

    /// candidates returns the left rows that a right row may match.
    fn candidates(&self, ctx: &mut Ctx<'_>, row: &[Value]) -> Candidates {
        match &self.hash {
            Some(hash) => hash.candidates(ctx, row),
            None => Candidates::All,
        }
    }
}

impl Rows for ProbeRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        loop {
            if let Some((right, candidates, position)) = self.current.as_mut() {
                if let Candidates::Found = *candidates {
                    *candidates = Candidates::None;
                    if let Some(found) = self.found.take() {
                        let mut row = found;
                        row.extend_from_slice(right);
                        if self.condition.map_or(Ok(true), |c| c.is_true(ctx, &row))? {
                            return Ok(Some(row));
                        }
                    }
                }
                let (bucket, exact): (&[usize], bool) = match (*candidates, &self.hash) {
                    (Candidates::Bucket(b), Some(hash)) => (hash.bucket(b), hash.exact),
                    _ => (&[], false),
                };
                loop {
                    let i = match *candidates {
                        Candidates::All if *position < self.left.len() => *position,
                        Candidates::Bucket(_) if *position < bucket.len() => bucket[*position],
                        _ => break,
                    };
                    *position += 1;
                    let mut row = Vec::with_capacity(self.left[i].len() + right.len());
                    row.extend_from_slice(&self.left[i]);
                    row.extend_from_slice(right);
                    if exact || self.condition.map_or(Ok(true), |c| c.is_true(ctx, &row))? {
                        return Ok(Some(row));
                    }
                }
                self.current = None;
            }
            let Some(right) = self.right.next(ctx)? else { return Ok(None) };
            let candidates = match self.lookup.as_mut() {
                Some(lookup) if lookup.budget > 0 => {
                    lookup.budget -= 1;
                    match lookup.find(ctx, &right) {
                        Ok(found) => {
                            self.found = found;
                            Candidates::Found
                        }
                        Err(Some(err)) => return Err(err),
                        Err(None) => self.read_left(ctx, &right)?,
                    }
                }
                Some(_) => self.read_left(ctx, &right)?,
                None => self.candidates(ctx, &right),
            };
            self.current = Some((right, candidates, 0));
        }
    }
}

/// LateralRows runs its right input again for each row of its left input, which the right input sees as its
/// enclosing row.
struct LateralRows<'p> {
    left: Box<dyn Rows + 'p>,
    right: &'p Plan,
    kind: JoinKind,
    condition: Option<&'p Expr>,
    right_width: usize,
    pending: std::vec::IntoIter<Row>,
}

impl Rows for LateralRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        loop {
            if let Some(row) = self.pending.next() {
                return Ok(Some(row));
            }
            let Some(l) = self.left.next(ctx)? else { return Ok(None) };
            ctx.outer.push(l.clone());
            let right_rows = self.right.run(ctx);
            ctx.outer.pop();
            let mut out = Vec::new();
            for r in right_rows? {
                let mut row = l.clone();
                row.extend(r);
                if self.condition.map_or(Ok(true), |c| c.is_true(ctx, &row))? {
                    out.push(row);
                }
            }
            if out.is_empty() && self.kind == JoinKind::Left {
                let mut row = l;
                row.extend(std::iter::repeat_n(Value::Null, self.right_width));
                out.push(row);
            }
            self.pending = out.into_iter();
        }
    }
}

/// LookupRows joins each left row with the right rows that its key values find in an index of the right input's
/// table, keeping those that the right input's filter, the ranges of its index scan, and the join condition keep.
struct LookupRows<'p> {
    left: Box<dyn Rows + 'p>,
    right_plan: &'p Plan,
    keys: &'p [Expr],
    found: crate::indexscan::IndexRows<'p>,
    /// The table columns of the index's keys.
    columns: Vec<usize>,
    table: &'p TableDef,
    filter: Option<&'p Expr>,
    /// The right input's index scan, when it is one with no filter above it, whose ranges the rows must lie in.
    ranges: Option<&'p crate::indexscan::IndexScan>,
    kind: JoinKind,
    condition: Option<&'p Expr>,
    right_width: usize,
    /// The right input's rows, read once a left row's key values are ones the index cannot hold.
    all: Option<Vec<Row>>,
    pending: std::vec::IntoIter<Row>,
}

impl<'p> LookupRows<'p> {
    /// open starts a join that looks up each left row's matches through a scan of the right input's table's index.
    fn open(
        ctx: &mut Ctx<'_>,
        left: &'p Plan,
        right: &'p Plan,
        kind: JoinKind,
        condition: Option<&'p Expr>,
        scan: &'p crate::indexscan::IndexScan,
        keys: &'p [Expr],
    ) -> Result<LookupRows<'p>> {
        let (filter, ranges) = match right {
            Plan::Filter { predicate, .. } => (Some(predicate), None),
            Plan::IndexScan(scan) => (None, Some(&**scan)),
            _ => (None, None),
        };
        Ok(LookupRows {
            left: left.open(ctx)?,
            right_plan: right,
            keys,
            found: scan.open_lookup(ctx)?,
            columns: scan.index_columns(),
            table: &scan.table,
            filter,
            ranges,
            kind,
            condition,
            right_width: right.width(),
            all: None,
            pending: Vec::new().into_iter(),
        })
    }

    /// matches returns the right rows that a left row finds.
    fn matches(&mut self, ctx: &mut Ctx<'_>, left: &[Value]) -> Result<Vec<Row>> {
        let mut range = Vec::with_capacity(self.columns.len());
        for (expr, &c) in self.keys.iter().zip(&self.columns) {
            let ty = self.table.index_column(c).map_or(0, |c| c.ty.oid);
            match expr.eval(ctx, left)? {
                Value::Null => return Ok(Vec::new()),
                value => match lookup_value(value, ty) {
                    Some(Some(value)) => range.push(crate::ranges::ColumnRange::closed(value)),
                    Some(None) => return Ok(Vec::new()),
                    None => {
                        if self.all.is_none() {
                            self.all = Some(self.right_plan.run(ctx)?);
                        }
                        return Ok(self.all.clone().unwrap_or_default());
                    }
                },
            }
        }
        range.resize(self.columns.len(), crate::ranges::ColumnRange::all());
        self.found.restart(vec![range]);
        let mut rows = Vec::new();
        while let Some(row) = self.found.next(ctx)? {
            if let Some(scan) = self.ranges {
                let key: Vec<Value> = scan.index_columns().iter().map(|&c| row[c].clone()).collect();
                if !scan.ranges.iter().any(|r| crate::ranges::range_contains(r, &key)) {
                    continue;
                }
            }
            if self.filter.map_or(Ok(true), |f| f.is_true(ctx, &row))? {
                rows.push(row);
            }
        }
        Ok(rows)
    }
}

impl Rows for LookupRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        loop {
            if let Some(row) = self.pending.next() {
                return Ok(Some(row));
            }
            let Some(l) = self.left.next(ctx)? else { return Ok(None) };
            let mut out = Vec::new();
            for r in self.matches(ctx, &l)? {
                let mut row = l.clone();
                row.extend(r);
                if self.condition.map_or(Ok(true), |c| c.is_true(ctx, &row))? {
                    out.push(row);
                }
            }
            match self.kind {
                JoinKind::Anti if out.is_empty() => {}
                JoinKind::Anti => continue,
                JoinKind::Semi if out.is_empty() => continue,
                JoinKind::Semi => {}
                JoinKind::Left if out.is_empty() => {}
                _ => {
                    self.pending = out.into_iter();
                    continue;
                }
            }
            let mut row = l;
            row.extend(std::iter::repeat_n(Value::Null, self.right_width));
            return Ok(Some(row));
        }
    }
}

/// HASHER builds the hashes of grouping keys, the same for every run.
const HASHER: foldhash::fast::FixedState = foldhash::fast::FixedState::with_seed(0);

/// Groups numbers distinct rows of values in the order they first appear, where NULLs are equal to each other and
/// values that compare equal are the same, as GROUP BY, DISTINCT, and set operations group rows.
#[derive(Clone, Default)]
pub struct Groups {
    table: hashbrown::HashTable<usize>,
    hashes: Vec<u64>,
    pub keys: Vec<Row>,
}

impl Groups {
    /// new returns an empty set of groups.
    pub fn new() -> Groups {
        Groups::default()
    }

    /// find returns the number of a row's group, if it has one.
    pub(crate) fn find(&self, row: &[Value]) -> Option<usize> {
        let hash = hash_row(row);
        self.table.find(hash, |&i| same_row(&self.keys[i], row)).copied()
    }

    /// insert returns the number of a row's group, adding a group for it when it has none, with whether it added one.
    pub fn insert(&mut self, row: &[Value]) -> (usize, bool) {
        let hash = hash_row(row);
        if let Some(&i) = self.table.find(hash, |&i| same_row(&self.keys[i], row)) {
            return (i, false);
        }
        let i = self.keys.len();
        let hashes = &self.hashes;
        self.table.insert_unique(hash, i, |&j| hashes[j]);
        self.hashes.push(hash);
        self.keys.push(row.to_vec());
        (i, true)
    }
}

/// hash_row hashes a row of values so that rows that `same_row` finds the same hash alike.
fn hash_row(row: &[Value]) -> u64 {
    let mut hasher = HASHER.build_hasher();
    for value in row {
        hash_value(value, &mut hasher);
    }
    hasher.finish()
}

/// class returns the kind of a value that grouping compares, where values of different kinds are never the same.
fn class(value: &Value) -> u8 {
    match value {
        Value::Null => 0,
        Value::Int2(_) | Value::Int4(_) | Value::Int8(_) | Value::Oid(_) | Value::Reg(_) => 1,
        Value::Float4(_) | Value::Float8(_) => 2,
        Value::Numeric(_) => 3,
        Value::Text(_) => 4,
        Value::Bool(_) => 5,
        Value::Date(_) => 6,
        Value::Time(_) => 7,
        Value::TimeTz(..) => 8,
        Value::Timestamp(_) => 9,
        Value::TimestampTz(_) => 10,
        Value::Interval(_) => 11,
        Value::Array(_) => 12,
        Value::Record(_) | Value::Composite(_) => 13,
        Value::Json(_) => 14,
        Value::Jsonb(_) => 15,
        Value::Xml(_) => 16,
        Value::Set(_) => 17,
        Value::Enum(_) => 18,
        Value::Bytea(_) => 19,
        Value::Uuid(_) => 20,
        Value::Bit(_) => 21,
        Value::Base(_) => 22,
        Value::Range(_) => 23,
        Value::Multirange(_) => 24,
    }
}

/// hash_value feeds a value to a hasher so that values that `same_value` finds the same hash alike.
fn hash_value(value: &Value, hasher: &mut impl Hasher) {
    hasher.write_u8(class(value));
    match value {
        Value::Null | Value::Set(_) => {}
        Value::Int2(i) => hasher.write_i64(*i as i64),
        Value::Int4(i) => hasher.write_i64(*i as i64),
        Value::Int8(i) => hasher.write_i64(*i),
        Value::Oid(o) => hasher.write_i64(*o as i64),
        Value::Reg(reg) => hasher.write_i64(reg.oid as i64),
        Value::Float4(f) => hash_float(*f as f64, hasher),
        Value::Float8(f) => hash_float(*f, hasher),
        Value::Numeric(n) => match n.trimmed() {
            crate::numeric::Numeric::Finite { negative, coefficient, scale } => {
                hasher.write_u8(u8::from(negative && coefficient != Default::default()));
                coefficient.hash(hasher);
                hasher.write_u32(scale);
            }
            other => hasher.write_u8(match other {
                crate::numeric::Numeric::NaN => 1,
                crate::numeric::Numeric::Infinity => 2,
                _ => 3,
            }),
        },
        Value::Text(s) | Value::Json(s) | Value::Xml(s) | Value::Bit(s) => s.hash(hasher),
        Value::Bool(b) => b.hash(hasher),
        Value::Date(d) => hasher.write_i32(*d),
        Value::Time(t) | Value::Timestamp(t) | Value::TimestampTz(t) => hasher.write_i64(*t),
        Value::TimeTz(t, z) => {
            hasher.write_i64(*t);
            hasher.write_i32(*z);
        }
        Value::Interval(i) => hasher.write_i128(i.cmp_key()),
        Value::Array(a) => {
            a.dims.hash(hasher);
            for v in &a.values {
                hash_value(v, hasher);
            }
        }
        Value::Record(fields) => fields.iter().for_each(|v| hash_value(v, hasher)),
        Value::Composite(c) => c.fields.iter().for_each(|v| hash_value(v, hasher)),
        Value::Jsonb(json) => crate::plan::trimmed_json(json).to_text().hash(hasher),
        Value::Enum(e) => e.label.hash(hasher),
        Value::Bytea(b) => b.hash(hasher),
        Value::Uuid(u) => u.hash(hasher),
        Value::Base(b) => hasher.write_u32(b.type_oid),
        Value::Range(r) => hash_range(r, hasher),
        Value::Multirange(m) => m.ranges.iter().for_each(|r| hash_range(r, hasher)),
    }
}

/// hash_range feeds a range to a hasher by its bounds, which ranges that compare equal share.
fn hash_range(range: &crate::rangetypes::Range, hasher: &mut impl Hasher) {
    hasher.write_u8(u8::from(range.empty));
    for bound in [&range.lower, &range.upper] {
        hasher.write_u8(u8::from(bound.inclusive));
        match &bound.value {
            Some(value) => hash_value(value, hasher),
            None => hasher.write_u8(0xff),
        }
    }
}

/// hash_float feeds a float to a hasher with zero's sign dropped and every NaN alike, as floats compare.
fn hash_float(f: f64, hasher: &mut impl Hasher) {
    let f = if f == 0.0 {
        0.0
    } else if f.is_nan() {
        f64::NAN
    } else {
        f
    };
    hasher.write_u64(f.to_bits());
}

/// same_row reports whether two rows group together.
fn same_row(a: &[Value], b: &[Value]) -> bool {
    a.len() == b.len() && a.iter().zip(b).all(|(l, r)| same_value(l, r))
}

/// same_value reports whether two values group together: both NULL, or of one kind and equal.
fn same_value(a: &Value, b: &Value) -> bool {
    match (a, b) {
        (Value::Null, Value::Null) => true,
        (Value::Text(l), Value::Text(r)) => l == r,
        (Value::Json(l), Value::Json(r)) | (Value::Xml(l), Value::Xml(r)) => l == r,
        _ if class(a) != class(b) || class(a) == 0 => false,
        _ => compare_values(a, b) == Ordering::Equal,
    }
}

/// aggregate groups a node's rows by the group keys and computes the aggregates of each group, as `Plan::Aggregate`
/// describes.
fn aggregate(
    ctx: &mut Ctx<'_>,
    input: &Plan,
    groups: &[Expr],
    aggregates: &[AggCall],
    sets: Option<&[Vec<usize>]>,
) -> Result<Vec<Row>> {
    let all: Vec<usize> = (0..groups.len()).collect();
    let masked = sets.is_some();
    let sets = sets.unwrap_or(std::slice::from_ref(&all));
    // Several grouping sets read the input once each.
    let stored = match sets.len() {
        1 => None,
        _ => Some(input.run(ctx)?),
    };
    let mut out = Vec::new();
    for set in sets {
        let mut rows = match &stored {
            Some(stored) => collected(stored.clone()),
            None => input.open(ctx)?,
        };
        let mut index = Groups::new();
        let mut states: Vec<Vec<Accumulator>> = Vec::new();
        let mut key = vec![Value::Null; groups.len()];
        let mut row = Vec::new();
        while rows.next_into(ctx, &mut row)? {
            for &i in set {
                key[i] = groups[i].eval(ctx, &row)?;
            }
            let (slot, added) = match set.is_empty() && !states.is_empty() {
                true => (0, false),
                false => index.insert(&key),
            };
            if added {
                states.push(aggregates.iter().map(Accumulator::new).collect());
            }
            for (agg, state) in aggregates.iter().zip(&mut states[slot]) {
                state.add(ctx, agg, &row)?;
            }
        }
        // Without group keys, an aggregate over no rows still returns one row.
        if set.is_empty() && states.is_empty() {
            index.insert(&key);
            states.push(aggregates.iter().map(Accumulator::new).collect());
        }
        let outside = (0..groups.len()).filter(|i| !set.contains(i)).fold(0i64, |m, i| m | 1 << i);
        for (key, state) in index.keys.into_iter().zip(states) {
            let mut row = key;
            for (agg, s) in aggregates.iter().zip(state) {
                row.push(s.finish(ctx, agg)?);
            }
            if masked {
                row.push(Value::Int8(outside));
            }
            out.push(row);
        }
    }
    Ok(out)
}

/// sort sorts a node's rows by the keys, keeping the order of rows with equal keys, and returns the first `limit`
/// of them when a limit is given.
fn sort(ctx: &mut Ctx<'_>, input: &Plan, keys: &[SortKey], limit: Option<usize>) -> Result<Vec<Row>> {
    let mut rows = input.open(ctx)?;
    let mut keyed: Vec<(Vec<Value>, Row)> = Vec::new();
    let order = |a: &(Vec<Value>, Row), b: &(Vec<Value>, Row)| crate::plan::compare_sorted(keys, &a.0, &b.0);
    while let Some(row) = rows.next(ctx)? {
        let values = keys.iter().map(|k| k.expr.eval(ctx, &row)).collect::<Result<Vec<_>>>()?;
        keyed.push((values, row));
        // A bounded sort keeps only the rows that can still be among the first, as Postgres' top-N heapsort does.
        if let Some(limit) = limit
            && keyed.len() >= limit.max(1) * 2 + 64
        {
            keyed.sort_by(order);
            keyed.truncate(limit);
        }
    }
    keyed.sort_by(order);
    if let Some(limit) = limit {
        keyed.truncate(limit);
    }
    Ok(keyed.into_iter().map(|(_, row)| row).collect())
}

impl Plan {
    /// open starts running the plan, returning the node that hands out its rows.
    pub fn open<'p>(&'p self, ctx: &mut Ctx<'_>) -> Result<Box<dyn Rows + 'p>> {
        Ok(match self {
            Plan::Scan(table, needed) => Box::new(TableWalk::new(ctx.db, table, needed.as_deref())?),
            Plan::IndexScan(scan) => scan.open(ctx)?,
            Plan::Filter { input, predicate } if matches!(**input, Plan::Once(_)) => {
                Box::new(OnceFilterRows::open(ctx, input, predicate)?)
            }
            Plan::Filter { input, predicate } => {
                let predicate = match predicate.foldable() {
                    true => Cow::Owned(predicate.clone().fold(ctx)),
                    false => Cow::Borrowed(predicate),
                };
                Box::new(FilterRows { input: input.open(ctx)?, predicate })
            }
            Plan::Project { input, exprs } => Box::new(ProjectRows::new(input.open(ctx)?, exprs)),
            Plan::Limit { input, limit, offset } => {
                let offset =
                    crate::plan::limit_value(offset, ctx, "OFFSET", code::INVALID_ROW_COUNT_IN_RESULT_OFFSET_CLAUSE)?;
                let limit = crate::plan::limit_value(limit, ctx, "LIMIT", code::INVALID_ROW_COUNT_IN_LIMIT_CLAUSE)?;
                let skip = offset.unwrap_or(0) as usize;
                let remaining = limit.map(|limit| limit as usize);
                let input = match (&**input, remaining) {
                    (Plan::Sort { input, keys }, Some(remaining)) => {
                        collected(sort(ctx, input, keys, Some(skip.saturating_add(remaining)))?)
                    }
                    _ if remaining == Some(0) => collected(Vec::new()),
                    _ => input.open(ctx)?,
                };
                Box::new(LimitRows { input, skip, remaining })
            }
            Plan::Join { left, right, kind, condition, lateral: true, .. } => Box::new(LateralRows {
                left: left.open(ctx)?,
                right,
                kind: *kind,
                condition: condition.as_ref(),
                right_width: right.width(),
                pending: Vec::new().into_iter(),
            }),
            Plan::Join { left, right, kind, condition, method: JoinMethod::Lookup { scan, keys }, .. }
                if !Lookup::finds_whole_keys(scan, keys, right, condition.as_ref(), left.width())? =>
            {
                Box::new(LookupRows::open(ctx, left, right, *kind, condition.as_ref(), scan, keys)?)
            }
            Plan::Join { left, right, kind, condition, method, .. } => {
                let condition = condition.as_ref();
                let unplanned = matches!(method, JoinMethod::Unplanned | JoinMethod::Ordered);
                let other: Option<Box<dyn Rows + 'p>> = match kind {
                    JoinKind::Inner if unplanned => {
                        ProbeRows::open(ctx, left, right, condition)?.map(|p| Box::new(p) as _)
                    }
                    JoinKind::Anti if unplanned || *method == JoinMethod::Hash => {
                        AntiRows::open(ctx, left, right, condition)?.map(|a| Box::new(a) as _)
                    }
                    _ => None,
                };
                let lookups = unplanned || matches!(method, JoinMethod::Lookup { .. });
                match other {
                    Some(rows) => rows,
                    None => Box::new(JoinRows::open(ctx, left, right, *kind, condition, lookups)?),
                }
            }
            Plan::Aggregate { input, groups, aggregates, sets } => match (&**input, sets) {
                // A table's row count is in its primary index's root, as go-mysql-server reads it for COUNT(*).
                (Plan::Scan(table, _), None)
                    if groups.is_empty()
                        && !table.keyless()
                        && aggregates.iter().all(|a| a.counts_rows() || a.counts_set_column(table)) =>
                {
                    let count = prolly::Node::decode(table.table.primary_index.clone())?.tree_count() as i64;
                    collected(vec![vec![Value::Int8(count); aggregates.len()]])
                }
                // An exact index scan's row count is the distance between its ranges' ends in the index.
                (Plan::IndexScan(scan), None)
                    if groups.is_empty()
                        && aggregates.iter().all(|a| a.counts_rows() || a.counts_set_column(&scan.table)) =>
                {
                    match scan.count(ctx)? {
                        Some(count) => collected(vec![vec![Value::Int8(count as i64); aggregates.len()]]),
                        None => collected(aggregate(ctx, input, groups, aggregates, None)?),
                    }
                }
                _ => collected(aggregate(ctx, input, groups, aggregates, sets.as_deref())?),
            },
            Plan::Sort { input, keys } => collected(sort(ctx, input, keys, None)?),
            Plan::Distinct { input, keys } => Box::new(DistinctRows {
                input: input.open(ctx)?,
                keys: keys.as_deref(),
                previous: None,
                seen: Groups::new(),
            }),
            Plan::SetOp { op: SetOp::Union, all: true, left, right } => {
                Box::new(ChainRows { inputs: vec![left, right], current: None })
            }
            Plan::SetOp { op, all, left, right } => {
                let left_rows = left.run(ctx)?;
                let right_rows = right.run(ctx)?;
                collected(crate::plan::set_operation(*op, *all, left_rows, right_rows))
            }
            Plan::Recursive { work_table, anchor, step, all } => {
                let mut recursive = RecursiveRows {
                    work_table: *work_table,
                    step,
                    all: *all,
                    seen: Groups::new(),
                    round: Vec::new(),
                    next: 0,
                };
                let rows = anchor.run(ctx)?;
                recursive.round = recursive.keep(rows);
                Box::new(recursive)
            }
            Plan::ProjectSet { input, functions, dropped } => Box::new(ProjectSetRows {
                input: input.open(ctx)?,
                functions,
                dropped,
                pending: Vec::new().into_iter(),
            }),
            Plan::Once(_) => Box::new(SharedRows { rows: self.shared_rows(ctx)?, next: 0 }),
            _ => collected(self.run_leaf(ctx)?),
        })
    }
}
