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
use crate::expr::{Expr, compare_values};
use crate::functions::aggregate::{Accumulator, AggCall};
use crate::plan::{HashKey, JoinKind, Plan, SetOp, SortKey, SubqueryRows};
use crate::query::Ctx;
use crate::types::Value;

/// Row is a row of values.
pub type Row = Vec<Value>;

/// Rows is a running plan node, which produces its rows one at a time.
pub trait Rows {
    /// next returns the node's next row, or None once it has no more.
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>>;

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
        if let Some((row, remaining)) = self.repeat.as_mut() {
            if *remaining > 1 {
                *remaining -= 1;
                return Ok(Some(row.clone()));
            }
            return Ok(self.repeat.take().map(|(row, _)| row));
        }
        let Some((key, value)) = self.items.current()? else { return Ok(None) };
        let (row, cardinality) = self.table.decode_columns(db, key, value, self.needed.as_deref())?;
        self.items.advance(db)?;
        if cardinality > 1 {
            self.repeat = Some((row.clone(), cardinality - 1));
        } else if cardinality == 0 {
            return self.next(db);
        }
        Ok(Some(row))
    }
}

impl Rows for TableWalk<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        TableWalk::next(self, ctx.db)
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
        while let Some(row) = self.input.next(ctx)? {
            if self.predicate.is_true(ctx, &row)? {
                return Ok(Some(row));
            }
        }
        Ok(None)
    }
}

/// ProjectRows computes expressions over each row of its input.
struct ProjectRows<'p> {
    input: Box<dyn Rows + 'p>,
    exprs: &'p [Expr],
    /// Whether each expression is a column that nothing else reads, whose value moves out of the input row.
    moves: Vec<bool>,
}

impl<'p> ProjectRows<'p> {
    /// new projects the input's rows.
    fn new(input: Box<dyn Rows + 'p>, exprs: &'p [Expr]) -> ProjectRows<'p> {
        let mut reads: std::collections::HashMap<usize, usize> = std::collections::HashMap::new();
        for expr in exprs {
            expr.visit(&mut |e| {
                if let Expr::Column(i) = e {
                    *reads.entry(*i).or_default() += 1;
                }
            });
        }
        let moves = exprs.iter().map(|e| matches!(e, Expr::Column(i) if reads[i] == 1)).collect();
        ProjectRows { input, exprs, moves }
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
        let Some(mut row) = self.input.next(ctx)? else { return Ok(None) };
        let mut out = Vec::with_capacity(self.exprs.len());
        for (expr, &moves) in self.exprs.iter().zip(&self.moves) {
            out.push(match expr {
                _ if moves => Value::Null,
                Expr::Column(i) => row[*i].clone(),
                expr => expr.eval(ctx, &row)?,
            });
        }
        for ((expr, &moves), value) in self.exprs.iter().zip(&self.moves).zip(out.iter_mut()) {
            if let (true, Expr::Column(i)) = (moves, expr) {
                *value = std::mem::replace(&mut row[*i], Value::Null);
            }
        }
        Ok(Some(out))
    }
}

/// LimitRows skips the first rows of its input and then hands out at most a number of rows.
struct LimitRows<'p> {
    input: Box<dyn Rows + 'p>,
    skip: usize,
    remaining: Option<usize>,
}

impl Rows for LimitRows<'_> {
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
        let shared = input.shared_rows(ctx)?;
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
            let mut key = Vec::with_capacity(index.outer.len());
            for e in &index.outer {
                match e.eval(ctx, &[])? {
                    Value::Null => return Ok(empty),
                    value => key.push(HashKey::of(value)),
                }
            }
            if let Some(key) = key.into_iter().collect::<Option<Vec<_>>>() {
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
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
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
            let row = &self.shared.rows[j];
            if predicate.is_true(ctx, row)? {
                return Ok(Some(row.clone()));
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
    pending: std::vec::IntoIter<Row>,
}

impl Rows for ProjectSetRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        loop {
            if let Some(row) = self.pending.next() {
                return Ok(Some(row));
            }
            let Some(row) = self.input.next(ctx)? else { return Ok(None) };
            let mut columns = Vec::with_capacity(self.functions.len());
            for function in self.functions {
                columns.push(crate::plan::set_rows(ctx, function, &row)?);
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

/// JoinRows pairs each row of its left input with the matching rows of its right input, which it reads first, finding
/// candidates through a hash table of the right rows' keys when the condition has equalities between the sides.
struct JoinRows<'p> {
    left: Box<dyn Rows + 'p>,
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

/// Candidates are the right rows that a left row may match: all of them, or one bucket of the hash table.
#[derive(Clone, Copy)]
enum Candidates {
    All,
    Bucket(usize),
    None,
}

/// JoinHash finds right rows by the values of the left row's side of the join's equalities.
struct JoinHash {
    left_keys: Vec<Expr>,
    /// The kind of key each right key holds, which a left key must share for the table to answer it.
    kinds: Vec<Option<std::mem::Discriminant<HashKey>>>,
    buckets: Vec<Vec<usize>>,
    table: crate::plan::KeyMap<usize>,
}

impl<'p> JoinRows<'p> {
    /// open starts a join of the inputs, reading the right input's rows.
    fn open(
        ctx: &mut Ctx<'_>,
        left: &'p Plan,
        right: &'p Plan,
        kind: JoinKind,
        condition: Option<&'p Expr>,
    ) -> Result<JoinRows<'p>> {
        let (left_width, right_width) = (left.width(), right.width());
        let left_rows = left.open(ctx)?;
        let right_rows = right.run(ctx)?;
        let hash = condition.and_then(|c| JoinHash::build(ctx, c, left_width, &right_rows));
        Ok(JoinRows {
            left: left_rows,
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

    /// candidates returns the right rows that a left row may match.
    fn candidates(&self, ctx: &mut Ctx<'_>, row: &[Value]) -> Candidates {
        let Some(hash) = &self.hash else { return Candidates::All };
        let mut key = Vec::with_capacity(hash.left_keys.len());
        for (expr, kind) in hash.left_keys.iter().zip(&hash.kinds) {
            match expr.eval(ctx, row) {
                Ok(Value::Null) => return Candidates::None,
                Ok(value) => match HashKey::of(value) {
                    Some(k) if kind.is_none_or(|kind| kind == std::mem::discriminant(&k)) => key.push(k),
                    _ => return Candidates::All,
                },
                Err(_) => return Candidates::All,
            }
        }
        match hash.table.get(&key) {
            Some(&bucket) => Candidates::Bucket(bucket),
            None => Candidates::None,
        }
    }
}

impl JoinHash {
    /// build hashes the right rows by their side of the equality conditions between the two sides, or returns None
    /// when the condition has none or a right key is not one it can hash.
    fn build(ctx: &mut Ctx<'_>, condition: &Expr, width: usize, right_rows: &[Row]) -> Option<JoinHash> {
        let (left_keys, right_keys) = crate::plan::join_keys(condition, width);
        if left_keys.is_empty() {
            return None;
        }
        let mut kinds = vec![None; right_keys.len()];
        let mut buckets: Vec<Vec<usize>> = Vec::new();
        let mut table: crate::plan::KeyMap<usize> = Default::default();
        'rows: for (j, row) in right_rows.iter().enumerate() {
            let mut key = Vec::with_capacity(right_keys.len());
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
                buckets.push(Vec::new());
                buckets.len() - 1
            });
            buckets[bucket].push(j);
        }
        Some(JoinHash { left_keys, kinds, buckets, table })
    }
}

impl Rows for JoinRows<'_> {
    fn next(&mut self, ctx: &mut Ctx<'_>) -> Result<Option<Row>> {
        loop {
            if let Some((left, candidates, position, matched)) = self.current.as_mut() {
                let bucket: &[usize] = match (*candidates, &self.hash) {
                    (Candidates::Bucket(b), Some(hash)) => &hash.buckets[b],
                    _ => &[],
                };
                loop {
                    let j = match *candidates {
                        Candidates::All if *position < self.right.len() => *position,
                        Candidates::Bucket(_) if *position < bucket.len() => bucket[*position],
                        _ => break,
                    };
                    *position += 1;
                    let mut row = Vec::with_capacity(left.len() + self.right[j].len());
                    row.extend_from_slice(left);
                    row.extend_from_slice(&self.right[j]);
                    if self.condition.map_or(Ok(true), |c| c.is_true(ctx, &row))? {
                        *matched = true;
                        self.right_matched[j] = true;
                        return Ok(Some(row));
                    }
                }
                let (mut left, _, _, matched) = self.current.take().expect("a current left row");
                if !matched && matches!(self.kind, JoinKind::Left | JoinKind::Full) {
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
                    let candidates = self.candidates(ctx, &left);
                    self.current = Some((left, candidates, 0, false));
                }
                None => self.left_done = true,
            }
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
        while let Some(row) = rows.next(ctx)? {
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
            Plan::Join { left, right, kind, condition, lateral: true } => Box::new(LateralRows {
                left: left.open(ctx)?,
                right,
                kind: *kind,
                condition: condition.as_ref(),
                right_width: right.width(),
                pending: Vec::new().into_iter(),
            }),
            Plan::Join { left, right, kind, condition, .. } => {
                Box::new(JoinRows::open(ctx, left, right, *kind, condition.as_ref())?)
            }
            Plan::Aggregate { input, groups, aggregates, sets } => match (&**input, sets) {
                // A table's row count is in its primary index's root, as go-mysql-server reads it for COUNT(*).
                (Plan::Scan(table, _), None)
                    if groups.is_empty() && !table.keyless() && aggregates.iter().all(AggCall::counts_rows) =>
                {
                    let count = prolly::Node::decode(table.table.primary_index.clone())?.tree_count() as i64;
                    collected(vec![vec![Value::Int8(count); aggregates.len()]])
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
            Plan::ProjectSet { input, functions } => {
                Box::new(ProjectSetRows { input: input.open(ctx)?, functions, pending: Vec::new().into_iter() })
            }
            Plan::Once(_) => Box::new(SharedRows { rows: self.shared_rows(ctx)?, next: 0 }),
            _ => collected(self.run_leaf(ctx)?),
        })
    }
}
