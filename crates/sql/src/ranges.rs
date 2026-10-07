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

//! Index ranges: the sets of index keys that an index scan reads, as go-mysql-server's MySQLRange types describe them,
//! with each range a run of column ranges and each column range a pair of cuts between values.

use std::cmp::Ordering;

use crate::catalog::ColumnType;
use crate::types::Value;

/// Cut is a point between index values: just below or above a value, around NULL, which sorts first, or after
/// every value.
#[derive(Clone, Debug, PartialEq)]
pub enum Cut {
    BelowNull,
    AboveNull,
    Below(Value),
    Above(Value),
    AboveAll,
}

impl Cut {
    /// rank orders the kinds of cuts, where cuts at values share a rank.
    fn rank(&self) -> u8 {
        match self {
            Cut::BelowNull => 0,
            Cut::AboveNull => 1,
            Cut::Below(_) | Cut::Above(_) => 2,
            Cut::AboveAll => 3,
        }
    }

    /// compare orders two cuts, as go-mysql-server's MySQLRangeCut.Compare does.
    pub fn compare(&self, other: &Cut) -> Ordering {
        match (self, other) {
            (Cut::Below(a) | Cut::Above(a), Cut::Below(b) | Cut::Above(b)) => crate::expr::compare_values(a, b)
                .then_with(|| match (self, other) {
                    (Cut::Below(_), Cut::Above(_)) => Ordering::Less,
                    (Cut::Above(_), Cut::Below(_)) => Ordering::Greater,
                    _ => Ordering::Equal,
                }),
            _ => self.rank().cmp(&other.rank()),
        }
    }
}

/// ColumnRange is the values of one index column between two cuts.
#[derive(Clone, Debug, PartialEq)]
pub struct ColumnRange {
    pub lower: Cut,
    pub upper: Cut,
}

/// Range is a run of column ranges, one for each index column it constrains, in index order.
pub type Range = Vec<ColumnRange>;

/// min_cut and max_cut return the lesser and greater of two cuts.
fn min_cut(a: &Cut, b: &Cut) -> Cut {
    if a.compare(b) == Ordering::Greater { b.clone() } else { a.clone() }
}

/// max_cut returns the greater of two cuts.
fn max_cut(a: &Cut, b: &Cut) -> Cut {
    if a.compare(b) == Ordering::Less { b.clone() } else { a.clone() }
}

/// go_time prints a UTC timestamp as Go's `time.Time` prints itself, such as `2024-01-03 00:00:00 +0000 UTC`.
fn go_time(ts: i64) -> String {
    let f = crate::datetime::fields_of_timestamp(ts);
    let fraction =
        if f.micros == 0 { String::new() } else { format!(".{:06}", f.micros).trim_end_matches('0').to_string() };
    format!(
        "{:04}-{:02}-{:02} {:02}:{:02}:{:02}{fraction} +0000 UTC",
        f.year, f.month, f.day, f.hour, f.minute, f.second
    )
}

/// key_text prints a range's key as go-mysql-server prints Doltgres' value, which is a Go `time.Time` for dates and
/// timestamps.
fn key_text(value: &Value) -> String {
    match value {
        Value::Timestamp(ts) | Value::TimestampTz(ts) => go_time(*ts),
        Value::Date(days) => go_time(*days as i64 * 86_400_000_000),
        other => other.output().unwrap_or_default(),
    }
}

impl ColumnRange {
    /// all returns the range of every value, NULL included.
    pub fn all() -> ColumnRange {
        ColumnRange { lower: Cut::BelowNull, upper: Cut::AboveAll }
    }

    /// empty returns the range of no values.
    pub fn empty() -> ColumnRange {
        ColumnRange { lower: Cut::AboveAll, upper: Cut::AboveAll }
    }

    /// null returns the range of only NULL.
    pub fn null() -> ColumnRange {
        ColumnRange { lower: Cut::BelowNull, upper: Cut::AboveNull }
    }

    /// not_null returns the range of every value but NULL.
    pub fn not_null() -> ColumnRange {
        ColumnRange { lower: Cut::AboveNull, upper: Cut::AboveAll }
    }

    /// closed returns the range of one value.
    pub fn closed(value: Value) -> ColumnRange {
        ColumnRange { lower: Cut::Below(value.clone()), upper: Cut::Above(value) }
    }

    /// is_empty reports whether the range holds no values.
    pub fn is_empty(&self) -> bool {
        self.lower.compare(&self.upper) != Ordering::Less
    }

    /// is_all reports whether the range holds every value, NULL included.
    pub fn is_all(&self) -> bool {
        self.lower == Cut::BelowNull && self.upper == Cut::AboveAll
    }

    /// contains reports whether a value lies in the range.
    pub fn contains(&self, value: &Value) -> bool {
        let (below, above) = match value {
            Value::Null => (Cut::BelowNull, Cut::AboveNull),
            v => (Cut::Below(v.clone()), Cut::Above(v.clone())),
        };
        self.lower.compare(&above) == Ordering::Less && below.compare(&self.upper) == Ordering::Less
    }

    /// try_intersect returns the values both ranges hold, or None when they share none.
    pub fn try_intersect(&self, other: &ColumnRange) -> Option<ColumnRange> {
        let lower = max_cut(&self.lower, &other.lower);
        let upper = min_cut(&self.upper, &other.upper);
        (lower.compare(&upper) == Ordering::Less).then_some(ColumnRange { lower, upper })
    }

    /// is_connected reports whether two ranges overlap or touch.
    fn is_connected(&self, other: &ColumnRange) -> bool {
        self.lower.compare(&other.upper) != Ordering::Greater && other.lower.compare(&self.upper) != Ordering::Greater
    }

    /// try_union returns the one range that holds the values of two ranges, or None when they leave a gap.
    pub fn try_union(&self, other: &ColumnRange) -> Option<ColumnRange> {
        if other.is_empty() {
            return Some(self.clone());
        }
        if self.is_empty() {
            return Some(other.clone());
        }
        if !self.is_connected(other) {
            return None;
        }
        Some(ColumnRange { lower: min_cut(&self.lower, &other.lower), upper: max_cut(&self.upper, &other.upper) })
    }

    /// overlap returns the values both ranges hold, or None when they share no values.
    fn overlap(&self, other: &ColumnRange) -> Option<ColumnRange> {
        if self.lower.compare(&other.upper) != Ordering::Less || other.lower.compare(&self.upper) != Ordering::Less {
            return None;
        }
        Some(ColumnRange { lower: max_cut(&self.lower, &other.lower), upper: min_cut(&self.upper, &other.upper) })
    }

    /// subtract returns the parts of the range outside another, as go-mysql-server's Subtract does.
    fn subtract(&self, other: &ColumnRange) -> Vec<ColumnRange> {
        if self.overlap(other).is_none() {
            return vec![self.clone()];
        }
        let lower = self.lower.compare(&other.lower);
        let upper = self.upper.compare(&other.upper);
        let before = || ColumnRange { lower: self.lower.clone(), upper: other.lower.clone() };
        let after = || ColumnRange { lower: other.upper.clone(), upper: self.upper.clone() };
        match (lower, upper) {
            (Ordering::Less, Ordering::Greater) => vec![before(), after()],
            (Ordering::Less, _) => vec![before()],
            (_, Ordering::Greater) => vec![after()],
            _ => Vec::new(),
        }
    }

    /// is_subset reports whether every value of the range lies in another.
    fn is_subset(&self, other: &ColumnRange) -> bool {
        self.lower.compare(&other.lower) != Ordering::Less && self.upper.compare(&other.upper) != Ordering::Greater
    }

    /// text prints the range as go-mysql-server's DebugString does, such as `[2, 5)` or `(NULL, ∞)`.
    pub fn text(&self) -> String {
        let key = key_text;
        let lower = match &self.lower {
            Cut::Above(v) => format!("({}", key(v)),
            Cut::Below(v) => format!("[{}", key(v)),
            Cut::AboveAll => "(∞".to_string(),
            Cut::AboveNull => "(NULL".to_string(),
            Cut::BelowNull => "[NULL".to_string(),
        };
        let upper = match &self.upper {
            Cut::Above(v) => format!("{}]", key(v)),
            Cut::Below(v) => format!("{})", key(v)),
            Cut::AboveAll => "∞)".to_string(),
            Cut::AboveNull => "NULL]".to_string(),
            Cut::BelowNull => "NULL)".to_string(),
        };
        format!("{lower}, {upper}")
    }
}

/// range_text prints a range as go-mysql-server does, such as `{[2, 2], (NULL, ∞)}`.
pub fn range_text(range: &Range) -> String {
    format!("{{{}}}", range.iter().map(ColumnRange::text).collect::<Vec<_>>().join(", "))
}

/// ranges_text prints ranges as go-mysql-server does, such as `[{[2, 2]}, {[4, 4]}]`.
pub fn ranges_text(ranges: &[Range]) -> String {
    format!("[{}]", ranges.iter().map(range_text).collect::<Vec<_>>().join(", "))
}

/// is_empty_range reports whether a range holds no keys, which any empty column range makes it.
pub fn is_empty_range(range: &Range) -> bool {
    range.is_empty() || range.iter().any(ColumnRange::is_empty)
}

/// range_contains reports whether a key's leading values lie in a range.
pub fn range_contains(range: &Range, key: &[Value]) -> bool {
    range.iter().zip(key).all(|(r, v)| r.contains(v))
}

/// compare_ranges orders ranges by each column's lower cut, then its upper cut.
fn compare_ranges(a: &Range, b: &Range) -> Ordering {
    for (x, y) in a.iter().zip(b) {
        let order = x.lower.compare(&y.lower).then_with(|| x.upper.compare(&y.upper));
        if order != Ordering::Equal {
            return order;
        }
    }
    Ordering::Equal
}

/// intersect_range returns the keys two ranges share, or an empty range.
fn intersect_range(a: &Range, b: &Range) -> Range {
    let mut out = Vec::with_capacity(a.len());
    for (x, y) in a.iter().zip(b) {
        match x.try_intersect(y) {
            Some(r) => out.push(r),
            None => return a.iter().map(|_| ColumnRange::empty()).collect(),
        }
    }
    out
}

/// is_superset reports whether a range holds every key of another.
fn is_superset(a: &Range, b: &Range) -> bool {
    a.len() == b.len() && b.iter().zip(a).all(|(x, y)| x.is_subset(y))
}

/// try_merge returns the one range that holds the keys of two ranges, when one holds the other or they differ in a
/// single column whose ranges join, as go-mysql-server's TryMerge does.
fn try_merge(a: &Range, b: &Range) -> Option<Range> {
    if is_superset(a, b) {
        return Some(a.clone());
    }
    if is_superset(b, a) {
        return Some(b.clone());
    }
    let mut differing = None;
    for (i, (x, y)) in a.iter().zip(b).enumerate() {
        if x != y {
            if differing.is_some() {
                return None;
            }
            differing = Some(i);
        }
    }
    let i = differing?;
    let merged = a[i].try_union(&b[i])?;
    let mut out = a.clone();
    out[i] = merged;
    Some(out)
}

/// overlaps reports whether two ranges share keys.
fn overlaps(a: &Range, b: &Range) -> bool {
    a.len() == b.len() && a.iter().zip(b).all(|(x, y)| x.overlap(y).is_some())
}

/// connects reports whether two ranges overlap or touch in every column, as go-mysql-server's range tree finds them.
fn connects(a: &Range, b: &Range) -> bool {
    a.len() == b.len() && a.iter().zip(b).all(|(x, y)| x.is_connected(y))
}

/// remove_overlap splits two overlapping ranges into disjoint ones, or merges them, as go-mysql-server's RemoveOverlap
/// does, reporting whether they changed.
fn remove_overlap(a: &Range, b: &Range) -> (Vec<Range>, bool) {
    if let Some(merged) = try_merge(a, b) {
        return (vec![merged], true);
    }
    if !overlaps(a, b) {
        return (vec![a.clone(), b.clone()], false);
    }
    let mut out = Vec::new();
    for i in 0..a.len() {
        if a[i] == b[i] {
            continue;
        }
        let Some(shared) = a[i].overlap(&b[i]) else { continue };
        let replace = |r: &Range, c: ColumnRange| {
            let mut r = r.clone();
            r[i] = c;
            r
        };
        out.extend(a[i].subtract(&shared).into_iter().map(|c| replace(a, c)));
        out.extend(b[i].subtract(&shared).into_iter().map(|c| replace(b, c)));
        let (rest, _) = remove_overlap(&replace(a, shared.clone()), &replace(b, shared));
        out.extend(rest);
        break;
    }
    (out, true)
}

/// remove_overlapping returns disjoint ranges that hold the keys of the ranges given, in key order, as
/// go-mysql-server's RemoveOverlappingRanges does.
pub fn remove_overlapping(ranges: Vec<Range>) -> Vec<Range> {
    let mut pending = ranges;
    let mut kept: Vec<Range> = Vec::new();
    let mut i = 0;
    while i < pending.len() {
        let range = pending[i].clone();
        i += 1;
        kept.sort_by(compare_ranges);
        let mut found = None;
        for (k, existing) in kept.iter().enumerate() {
            if !connects(existing, &range) {
                continue;
            }
            let (split, changed) = remove_overlap(existing, &range);
            if changed {
                found = Some((k, split));
                break;
            }
        }
        match found {
            Some((k, split)) => {
                kept.remove(k);
                pending.extend(split);
            }
            None => kept.push(range),
        }
    }
    kept.retain(|r| !is_empty_range(r));
    kept.sort_by(compare_ranges);
    kept
}

/// intersect returns the keys that both collections of ranges hold, as go-mysql-server's MySQLRangeCollection
/// Intersect does.
pub fn intersect(a: &[Range], b: &[Range]) -> Vec<Range> {
    let mut out = Vec::new();
    for x in a {
        for y in b {
            let r = intersect_range(x, y);
            if !r.is_empty() {
                out.push(r);
            }
        }
    }
    remove_overlapping(out)
}

/// simplify joins a column's connected ranges, in order, as go-mysql-server's SimplifyRangeColumn does.
fn simplify(ranges: &[ColumnRange]) -> Vec<ColumnRange> {
    let mut sorted = ranges.to_vec();
    sorted.sort_by(|a, b| a.lower.compare(&b.lower).then_with(|| a.upper.compare(&b.upper)));
    let mut out = Vec::new();
    let mut current = ColumnRange::empty();
    for r in sorted {
        match current.try_union(&r) {
            Some(merged) => current = merged,
            None => {
                if !current.is_empty() {
                    out.push(current);
                }
                current = r;
            }
        }
    }
    if !current.is_empty() {
        out.push(current);
    }
    out
}

/// IndexBuilder gathers the ranges of each index column that a conjunction of comparisons allows, as
/// go-mysql-server's MySQLIndexBuilder does.
pub struct IndexBuilder {
    /// The index's columns, each with its type and the ranges it allows.
    columns: Vec<(String, ColumnType, Vec<ColumnRange>)>,
    invalid: bool,
}

impl IndexBuilder {
    /// new returns a builder that allows every key of an index with the columns.
    pub fn new(columns: &[(String, ColumnType)]) -> IndexBuilder {
        IndexBuilder {
            columns: columns.iter().map(|(n, t)| (n.clone(), *t, vec![ColumnRange::all()])).collect(),
            invalid: false,
        }
    }

    /// key converts a value to a column's type, when it converts without changing.
    fn key(ty: ColumnType, value: &Value) -> Option<Value> {
        let converted = crate::cast::cast_value(value.clone(), ty, false).ok()?;
        let original = crate::functions::value_type(value);
        let back = crate::cast::cast_value(converted.clone(), crate::expr::typ(original), false).ok()?;
        (crate::expr::compare_values(&back, value) == Ordering::Equal).then_some(converted)
    }

    /// update narrows a column's ranges to those that also lie in any of the ranges given.
    fn update(&mut self, column: &str, allowed: Vec<ColumnRange>) {
        if self.invalid {
            return;
        }
        let Some((_, _, current)) = self.columns.iter_mut().find(|(n, _, _)| n == column) else {
            self.invalid = true;
            return;
        };
        let mut narrowed = Vec::new();
        for c in current.iter() {
            for a in &allowed {
                if let Some(r) = c.try_intersect(a)
                    && !r.is_empty()
                {
                    narrowed.push(r);
                }
            }
        }
        if narrowed.is_empty() {
            self.invalid = true;
            return;
        }
        *current = narrowed;
    }

    /// column_type returns a column's type.
    fn column_type(&self, column: &str) -> Option<ColumnType> {
        self.columns.iter().find(|(n, _, _)| n == column).map(|(_, t, _)| *t)
    }

    /// compare narrows a column to the values that a comparison with a value allows, which reports false when the
    /// value has no exact form in the column's type.
    pub fn compare(&mut self, column: &str, op: crate::expr::CmpOp, value: &Value) -> bool {
        use crate::expr::CmpOp;
        let Some(ty) = self.column_type(column) else { return false };
        if value.is_null() {
            self.update(column, vec![ColumnRange::empty()]);
            return true;
        }
        let Some(key) = Self::key(ty, value) else { return false };
        let ranges = match op {
            CmpOp::Eq => vec![ColumnRange::closed(key)],
            CmpOp::Ne => vec![
                ColumnRange { lower: Cut::Above(key.clone()), upper: Cut::AboveAll },
                ColumnRange { lower: Cut::AboveNull, upper: Cut::Below(key) },
            ],
            CmpOp::Gt => vec![ColumnRange { lower: Cut::Above(key), upper: Cut::AboveAll }],
            CmpOp::Ge => vec![ColumnRange { lower: Cut::Below(key), upper: Cut::AboveAll }],
            CmpOp::Lt => vec![ColumnRange { lower: Cut::AboveNull, upper: Cut::Below(key) }],
            CmpOp::Le => vec![ColumnRange { lower: Cut::AboveNull, upper: Cut::Above(key) }],
        };
        self.update(column, ranges);
        if op == CmpOp::Ne
            && !self.invalid
            && let Some((_, _, current)) = self.columns.iter_mut().find(|(n, _, _)| n == column)
        {
            *current = simplify(current);
            if current.is_empty() {
                self.invalid = true;
            }
        }
        true
    }

    /// is_null narrows a column to NULL, or away from it when negated.
    pub fn is_null(&mut self, column: &str, negated: bool) {
        self.update(column, vec![if negated { ColumnRange::not_null() } else { ColumnRange::null() }]);
    }

    /// ranges returns the ranges of the index's keys that the comparisons allow: every combination of the columns'
    /// ranges, as go-mysql-server's MySQLIndexBuilder.Ranges does, or one empty range when none remains.
    pub fn ranges(&self) -> Vec<Range> {
        let empty = || vec![self.columns.iter().map(|_| ColumnRange::empty()).collect::<Range>()];
        if self.invalid {
            return empty();
        }
        let mut out: Vec<Range> = vec![Vec::new()];
        for (_, _, ranges) in &self.columns {
            let mut next = Vec::new();
            for prefix in &out {
                for r in ranges {
                    let mut range = prefix.clone();
                    range.push(r.clone());
                    next.push(range);
                }
            }
            out = next;
        }
        out.retain(|r| !is_empty_range(r));
        if out.is_empty() { empty() } else { out }
    }
}

/// is_all_range reports whether ranges hold every key, which makes an index scan pointless.
pub fn is_all_range(ranges: &[Range]) -> bool {
    matches!(ranges, [range] if range.first().is_some_and(ColumnRange::is_all))
}
