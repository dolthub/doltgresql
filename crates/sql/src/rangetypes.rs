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

//! Range and multirange types: their values, text and binary formats, and the operations on them, after Postgres'
//! rangetypes.c and multirangetypes.c.

use std::cmp::Ordering;

use crate::error::{PgError, Result, code};
use crate::expr::compare_values;
use crate::oid;
use crate::types::Value;

/// The polymorphic range pseudo-types.
pub const ANYRANGE: u32 = 3831;
pub const ANYMULTIRANGE: u32 = 4537;

/// RangeType is a range type with its subtype and its multirange type, and whether its values are discrete, which
/// makes ranges canonical with an inclusive lower bound and an exclusive upper bound.
#[derive(Clone, Copy, Debug, PartialEq)]
pub struct RangeType {
    pub range: u32,
    pub subtype: u32,
    pub multirange: u32,
    pub discrete: bool,
}

/// BUILTIN are the built-in range types.
const BUILTIN: [RangeType; 6] = [
    RangeType { range: 3904, subtype: oid::INT4, multirange: 4451, discrete: true },
    RangeType { range: 3926, subtype: oid::INT8, multirange: 4536, discrete: true },
    RangeType { range: 3906, subtype: oid::NUMERIC, multirange: 4532, discrete: false },
    RangeType { range: 3908, subtype: oid::TIMESTAMP, multirange: 4533, discrete: false },
    RangeType { range: 3910, subtype: oid::TIMESTAMPTZ, multirange: 4534, discrete: false },
    RangeType { range: 3912, subtype: oid::DATE, multirange: 4535, discrete: true },
];

/// range_type returns the range type with an OID, built-in or user-defined.
pub fn range_type(type_oid: u32) -> Option<RangeType> {
    if let Some(builtin) = BUILTIN.iter().find(|t| t.range == type_oid) {
        return Some(*builtin);
    }
    match crate::usertypes::get(type_oid)?.kind {
        crate::usertypes::Kind::Range(subtype) => Some(RangeType {
            range: type_oid,
            subtype: subtype.oid,
            multirange: crate::usertypes::multirange_of(type_oid),
            discrete: false,
        }),
        _ => None,
    }
}

/// multirange_type returns the range type of a multirange type's OID.
pub fn multirange_type(type_oid: u32) -> Option<RangeType> {
    if let Some(builtin) = BUILTIN.iter().find(|t| t.multirange == type_oid) {
        return Some(*builtin);
    }
    match crate::usertypes::get(type_oid)?.kind {
        crate::usertypes::Kind::Multirange(range) => Some(RangeType { multirange: type_oid, ..range_type(range)? }),
        _ => None,
    }
}

/// is_range reports whether a type is a range type.
pub fn is_range(type_oid: u32) -> bool {
    range_type(type_oid).is_some()
}

/// is_multirange reports whether a type is a multirange type.
pub fn is_multirange(type_oid: u32) -> bool {
    multirange_type(type_oid).is_some()
}

/// Bound is one end of a range: its value, which is None for an infinite bound, and whether it includes the value.
#[derive(Clone, Debug, PartialEq)]
pub struct Bound {
    pub value: Option<Value>,
    pub inclusive: bool,
}

/// Range is a value of a range type, which is either empty or spans its bounds.
#[derive(Clone, Debug, PartialEq)]
pub struct Range {
    pub type_oid: u32,
    pub empty: bool,
    pub lower: Bound,
    pub upper: Bound,
}

/// Multirange is a value of a multirange type: its ranges, sorted, never empty, and neither overlapping nor adjacent.
#[derive(Clone, Debug, PartialEq)]
pub struct Multirange {
    pub type_oid: u32,
    pub ranges: Vec<Range>,
}

/// infinite returns an infinite bound.
fn infinite() -> Bound {
    Bound { value: None, inclusive: false }
}

/// empty returns the empty range of a range type.
pub fn empty(type_oid: u32) -> Range {
    Range { type_oid, empty: true, lower: infinite(), upper: infinite() }
}

/// compare_bounds orders two bounds, each a lower bound or not, as Postgres' range_cmp_bounds does.
fn compare_bounds(left: &Bound, left_lower: bool, right: &Bound, right_lower: bool) -> Ordering {
    let order_of = |lower: bool| if lower { Ordering::Less } else { Ordering::Greater };
    match (&left.value, &right.value) {
        (None, None) if left_lower == right_lower => Ordering::Equal,
        (None, _) => order_of(left_lower),
        (_, None) => order_of(right_lower).reverse(),
        (Some(l), Some(r)) => match compare_values(l, r) {
            Ordering::Equal => match (left.inclusive, right.inclusive) {
                (false, false) if left_lower == right_lower => Ordering::Equal,
                (false, false) => order_of(left_lower).reverse(),
                (false, true) => order_of(left_lower).reverse(),
                (true, false) => order_of(right_lower),
                (true, true) => Ordering::Equal,
            },
            other => other,
        },
    }
}

/// compare_bound_values orders two bounds by their values alone, as Postgres' range_cmp_bound_values does.
fn compare_bound_values(left: &Bound, left_lower: bool, right: &Bound, right_lower: bool) -> Ordering {
    let order_of = |lower: bool| if lower { Ordering::Less } else { Ordering::Greater };
    match (&left.value, &right.value) {
        (None, None) if left_lower == right_lower => Ordering::Equal,
        (None, _) => order_of(left_lower),
        (_, None) => order_of(right_lower).reverse(),
        (Some(l), Some(r)) => compare_values(l, r),
    }
}

/// step returns the value after a discrete value, or None for an infinite date, which has no next value.
fn step(value: &Value) -> Result<Option<Value>> {
    let overflow = crate::cast::int_out_of_range;
    Ok(Some(match value {
        Value::Int4(i) => Value::Int4(i.checked_add(1).ok_or_else(|| overflow(oid::INT4))?),
        Value::Int8(i) => Value::Int8(i.checked_add(1).ok_or_else(|| overflow(oid::INT8))?),
        Value::Date(d) if *d == crate::datetime::DATE_NOBEGIN || *d == crate::datetime::DATE_NOEND => return Ok(None),
        Value::Date(d) => Value::Date(
            d.checked_add(1).ok_or_else(|| PgError::new(code::DATETIME_FIELD_OVERFLOW, "date out of range"))?,
        ),
        _ => return Ok(None),
    }))
}

/// canonical moves a discrete range's bounds to an inclusive lower bound and an exclusive upper bound.
fn canonical(lower: &mut Bound, upper: &mut Bound) -> Result<()> {
    if let Some(value) = &lower.value
        && !lower.inclusive
        && let Some(next) = step(value)?
    {
        *lower = Bound { value: Some(next), inclusive: true };
    }
    if let Some(value) = &upper.value
        && upper.inclusive
        && let Some(next) = step(value)?
    {
        *upper = Bound { value: Some(next), inclusive: false };
    }
    Ok(())
}

/// make returns the range between two bounds, failing when the lower bound is above the upper one, as Postgres'
/// make_range does.
pub fn make(type_oid: u32, mut lower: Bound, mut upper: Bound) -> Result<Range> {
    if lower.value.is_none() {
        lower.inclusive = false;
    }
    if upper.value.is_none() {
        upper.inclusive = false;
    }
    match compare_bound_values(&lower, true, &upper, false) {
        Ordering::Greater => {
            return Err(PgError::new(
                code::DATA_EXCEPTION,
                "range lower bound must be less than or equal to range upper bound",
            ));
        }
        Ordering::Equal if !(lower.inclusive && upper.inclusive) => return Ok(empty(type_oid)),
        _ => {}
    }
    if range_type(type_oid).is_some_and(|t| t.discrete) {
        canonical(&mut lower, &mut upper)?;
        if compare_bound_values(&lower, true, &upper, false).is_eq() && !(lower.inclusive && upper.inclusive) {
            return Ok(empty(type_oid));
        }
    }
    Ok(Range { type_oid, empty: false, lower, upper })
}

/// malformed returns Postgres' error for range or multirange text that fails to parse.
fn malformed(kind: &str, text: &str, detail: &str) -> PgError {
    PgError {
        detail: Some(detail.to_string()),
        ..PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("malformed {kind} literal: \"{text}\""))
    }
}

/// parse_bound reads a bound's text from a range literal at a byte position, returning None for an infinite bound and
/// the position after it.
fn parse_bound(text: &str, mut at: usize) -> Result<(Option<String>, usize)> {
    let bytes = text.as_bytes();
    if matches!(bytes.get(at), Some(b',' | b')' | b']')) {
        return Ok((None, at));
    }
    let mut out = Vec::new();
    let mut quoted = false;
    loop {
        let Some(&ch) = bytes.get(at) else { return Err(malformed("range", text, "Unexpected end of input.")) };
        if !quoted && matches!(ch, b',' | b')' | b']') {
            break;
        }
        at += 1;
        match ch {
            b'\\' => {
                let Some(&next) = bytes.get(at) else {
                    return Err(malformed("range", text, "Unexpected end of input."));
                };
                out.push(next);
                at += 1;
            }
            b'"' if !quoted => quoted = true,
            b'"' if bytes.get(at) == Some(&b'"') => {
                out.push(b'"');
                at += 1;
            }
            b'"' => quoted = false,
            _ => out.push(ch),
        }
    }
    Ok((Some(String::from_utf8_lossy(&out).into_owned()), at))
}

/// skip_spaces returns the position after the whitespace at a byte position.
fn skip_spaces(bytes: &[u8], mut at: usize) -> usize {
    while bytes.get(at).is_some_and(|b| b.is_ascii_whitespace()) {
        at += 1;
    }
    at
}

/// parse reads a value of a range type from its text format.
pub fn parse(text: &str, type_oid: u32) -> Result<Range> {
    let subtype = range_type(type_oid).map_or(oid::TEXT, |t| t.subtype);
    let bytes = text.as_bytes();
    let mut at = skip_spaces(bytes, 0);
    if bytes.len() >= at + 5 && bytes[at..at + 5].eq_ignore_ascii_case(b"empty") {
        at = skip_spaces(bytes, at + 5);
        if at < bytes.len() {
            return Err(malformed("range", text, "Junk after \"empty\" key word."));
        }
        return Ok(empty(type_oid));
    }
    let lower_inclusive = match bytes.get(at) {
        Some(b'[') => true,
        Some(b'(') => false,
        _ => return Err(malformed("range", text, "Missing left parenthesis or bracket.")),
    };
    let (lower, next) = parse_bound(text, at + 1)?;
    if bytes.get(next) != Some(&b',') {
        return Err(malformed("range", text, "Missing comma after lower bound."));
    }
    let (upper, next) = parse_bound(text, next + 1)?;
    let upper_inclusive = match bytes.get(next) {
        Some(b']') => true,
        Some(b')') => false,
        _ => return Err(malformed("range", text, "Too many commas.")),
    };
    if skip_spaces(bytes, next + 1) < bytes.len() {
        return Err(malformed("range", text, "Junk after right parenthesis or bracket."));
    }
    let value = |bound: Option<String>| bound.map(|b| crate::cast::input(&b, subtype)).transpose();
    make(
        type_oid,
        Bound { value: value(lower)?, inclusive: lower_inclusive },
        Bound { value: value(upper)?, inclusive: upper_inclusive },
    )
}

/// quote returns a bound's text as a range literal holds it, quoting it when it is empty or holds a character that
/// the literal's syntax uses.
fn quote(text: &str) -> String {
    let special = |c: char| matches!(c, '"' | '\\' | '(' | ')' | '[' | ']' | ',') || c.is_ascii_whitespace();
    if !text.is_empty() && !text.chars().any(special) {
        return text.to_string();
    }
    let mut out = String::with_capacity(text.len() + 2);
    out.push('"');
    for c in text.chars() {
        if c == '"' || c == '\\' {
            out.push(c);
        }
        out.push(c);
    }
    out.push('"');
    out
}

/// format returns a range's text format.
pub fn format(range: &Range) -> String {
    if range.empty {
        return "empty".to_string();
    }
    let text = |bound: &Bound| bound.value.as_ref().and_then(Value::output).map(|t| quote(&t)).unwrap_or_default();
    format!(
        "{}{},{}{}",
        if range.lower.inclusive { '[' } else { '(' },
        text(&range.lower),
        text(&range.upper),
        if range.upper.inclusive { ']' } else { ')' }
    )
}

/// parse_multirange reads a value of a multirange type from its text format.
pub fn parse_multirange(text: &str, type_oid: u32) -> Result<Multirange> {
    let range_oid = multirange_type(type_oid).map_or(0, |t| t.range);
    let bytes = text.as_bytes();
    let mut at = skip_spaces(bytes, 0);
    if bytes.get(at) != Some(&b'{') {
        return Err(malformed("multirange", text, "Missing left brace."));
    }
    at += 1;
    #[derive(PartialEq)]
    enum State {
        Before,
        In,
        Escaped,
        Quoted,
        QuotedEscaped,
        After,
    }
    let mut state = State::Before;
    let (mut start, mut seen) = (0, 0);
    let mut ranges = Vec::new();
    loop {
        let Some(&ch) = bytes.get(at) else { return Err(malformed("multirange", text, "Unexpected end of input.")) };
        if ch.is_ascii_whitespace() {
            at += 1;
            continue;
        }
        match state {
            State::Before if matches!(ch, b'[' | b'(') => {
                start = at;
                state = State::In;
            }
            State::Before if ch == b'}' && seen == 0 => break,
            State::Before if bytes.len() >= at + 5 && bytes[at..at + 5].eq_ignore_ascii_case(b"empty") => {
                seen += 1;
                at += 4;
                state = State::After;
            }
            State::Before => return Err(malformed("multirange", text, "Expected range start.")),
            State::In if matches!(ch, b']' | b')') => {
                let range = parse(&text[start..=at], range_oid)?;
                if !range.empty {
                    ranges.push(range);
                }
                seen += 1;
                state = State::After;
            }
            State::In if ch == b'"' => state = State::Quoted,
            State::In if ch == b'\\' => state = State::Escaped,
            State::In => {}
            State::Escaped => state = State::In,
            State::Quoted if ch == b'"' && bytes.get(at + 1) == Some(&b'"') => at += 1,
            State::Quoted if ch == b'"' => state = State::In,
            State::Quoted if ch == b'\\' => state = State::QuotedEscaped,
            State::Quoted => {}
            State::QuotedEscaped => state = State::Quoted,
            State::After if ch == b',' => state = State::Before,
            State::After if ch == b'}' => break,
            State::After => return Err(malformed("multirange", text, "Expected comma or end of multirange.")),
        }
        at += 1;
    }
    if skip_spaces(bytes, at + 1) < bytes.len() {
        return Err(malformed("multirange", text, "Junk after closing right brace."));
    }
    normalize(type_oid, ranges)
}

/// format_multirange returns a multirange's text format.
pub fn format_multirange(multirange: &Multirange) -> String {
    let ranges: Vec<String> = multirange.ranges.iter().map(format).collect();
    format!("{{{}}}", ranges.join(","))
}

/// compare orders two ranges, with empty ranges first, as Postgres' range_cmp does.
pub fn compare(left: &Range, right: &Range) -> Ordering {
    match (left.empty, right.empty) {
        (true, true) => Ordering::Equal,
        (true, false) => Ordering::Less,
        (false, true) => Ordering::Greater,
        (false, false) => compare_bounds(&left.lower, true, &right.lower, true)
            .then_with(|| compare_bounds(&left.upper, false, &right.upper, false)),
    }
}

/// compare_multiranges orders two multiranges range by range, with a shorter one first when one is a prefix of the
/// other, as Postgres' multirange_cmp does.
pub fn compare_multiranges(left: &Multirange, right: &Multirange) -> Ordering {
    for (l, r) in left.ranges.iter().zip(&right.ranges) {
        let ordering = compare(l, r);
        if ordering != Ordering::Equal {
            return ordering;
        }
    }
    left.ranges.len().cmp(&right.ranges.len())
}

/// contains_element reports whether a range contains a value.
pub fn contains_element(range: &Range, value: &Value) -> bool {
    if range.empty {
        return false;
    }
    let element = Bound { value: Some(value.clone()), inclusive: true };
    compare_bounds(&range.lower, true, &element, true) != Ordering::Greater
        && compare_bounds(&range.upper, false, &element, false) != Ordering::Less
}

/// contains reports whether a range contains another range.
pub fn contains(outer: &Range, inner: &Range) -> bool {
    if inner.empty {
        return true;
    }
    !outer.empty
        && compare_bounds(&outer.lower, true, &inner.lower, true) != Ordering::Greater
        && compare_bounds(&outer.upper, false, &inner.upper, false) != Ordering::Less
}

/// overlaps reports whether two ranges share a value.
pub fn overlaps(left: &Range, right: &Range) -> bool {
    if left.empty || right.empty {
        return false;
    }
    let within = |a: &Range, b: &Range| {
        compare_bounds(&a.lower, true, &b.lower, true) != Ordering::Less
            && compare_bounds(&a.lower, true, &b.upper, false) != Ordering::Greater
    };
    within(left, right) || within(right, left)
}

/// before reports whether a range ends before another starts.
pub fn before(left: &Range, right: &Range) -> bool {
    !left.empty && !right.empty && compare_bounds(&left.upper, false, &right.lower, true) == Ordering::Less
}

/// after reports whether a range starts after another ends.
pub fn after(left: &Range, right: &Range) -> bool {
    !left.empty && !right.empty && compare_bounds(&left.lower, true, &right.upper, false) == Ordering::Greater
}

/// over_left reports whether a range ends no later than another, as the &< operator does.
pub fn over_left(left: &Range, right: &Range) -> bool {
    !left.empty && !right.empty && compare_bounds(&left.upper, false, &right.upper, false) != Ordering::Greater
}

/// over_right reports whether a range starts no earlier than another, as the &> operator does.
pub fn over_right(left: &Range, right: &Range) -> bool {
    !left.empty && !right.empty && compare_bounds(&left.lower, true, &right.lower, true) != Ordering::Less
}

/// bounds_adjacent reports whether an upper bound meets a lower bound with no value between them.
pub fn bounds_adjacent(type_oid: u32, upper: &Bound, lower: &Bound) -> bool {
    match compare_bound_values(upper, false, lower, true) {
        Ordering::Less if range_type(type_oid).is_some_and(|t| t.discrete) => {
            let between = make(
                type_oid,
                Bound { value: upper.value.clone(), inclusive: !upper.inclusive },
                Bound { value: lower.value.clone(), inclusive: !lower.inclusive },
            );
            between.is_ok_and(|r| r.empty)
        }
        Ordering::Equal => upper.inclusive != lower.inclusive,
        _ => false,
    }
}

/// adjacent reports whether two ranges meet with no value between them.
pub fn adjacent(left: &Range, right: &Range) -> bool {
    !left.empty
        && !right.empty
        && (bounds_adjacent(left.type_oid, &left.upper, &right.lower)
            || bounds_adjacent(left.type_oid, &right.upper, &left.lower))
}

/// merge returns the smallest range that covers two ranges, failing when `contiguous` is set and they leave a gap.
pub fn merge(left: &Range, right: &Range, contiguous: bool) -> Result<Range> {
    if left.empty {
        return Ok(right.clone());
    }
    if right.empty {
        return Ok(left.clone());
    }
    if contiguous && !overlaps(left, right) && !adjacent(left, right) {
        return Err(PgError::new(code::DATA_EXCEPTION, "result of range union would not be contiguous"));
    }
    let lower = match compare_bounds(&left.lower, true, &right.lower, true) {
        Ordering::Greater => right.lower.clone(),
        _ => left.lower.clone(),
    };
    let upper = match compare_bounds(&left.upper, false, &right.upper, false) {
        Ordering::Less => right.upper.clone(),
        _ => left.upper.clone(),
    };
    make(left.type_oid, lower, upper)
}

/// intersect returns the values two ranges share.
pub fn intersect(left: &Range, right: &Range) -> Result<Range> {
    if !overlaps(left, right) {
        return Ok(empty(left.type_oid));
    }
    let lower = match compare_bounds(&left.lower, true, &right.lower, true) {
        Ordering::Less => right.lower.clone(),
        _ => left.lower.clone(),
    };
    let upper = match compare_bounds(&left.upper, false, &right.upper, false) {
        Ordering::Greater => right.upper.clone(),
        _ => left.upper.clone(),
    };
    make(left.type_oid, lower, upper)
}

/// split returns the parts of a range that another range leaves, which number up to two.
fn split(left: &Range, right: &Range) -> Result<Vec<Range>> {
    if left.empty || right.empty || !overlaps(left, right) {
        return Ok(vec![left.clone()]);
    }
    let mut parts = Vec::new();
    if compare_bounds(&left.lower, true, &right.lower, true) == Ordering::Less {
        let upper = Bound { value: right.lower.value.clone(), inclusive: !right.lower.inclusive };
        parts.push(make(left.type_oid, left.lower.clone(), upper)?);
    }
    if compare_bounds(&left.upper, false, &right.upper, false) == Ordering::Greater {
        let lower = Bound { value: right.upper.value.clone(), inclusive: !right.upper.inclusive };
        parts.push(make(left.type_oid, lower, left.upper.clone())?);
    }
    parts.retain(|p| !p.empty);
    Ok(parts)
}

/// minus returns the values of a range that another range leaves, failing when they would be two ranges.
pub fn minus(left: &Range, right: &Range) -> Result<Range> {
    let mut parts = split(left, right)?;
    match parts.len() {
        0 => Ok(empty(left.type_oid)),
        1 => Ok(parts.remove(0)),
        _ => Err(PgError::new(code::DATA_EXCEPTION, "result of range difference would not be contiguous")),
    }
}

/// normalize returns the multirange of ranges: sorted, without empty ranges, and with ranges that overlap or meet
/// merged.
pub fn normalize(type_oid: u32, mut ranges: Vec<Range>) -> Result<Multirange> {
    ranges.retain(|r| !r.empty);
    ranges.sort_by(compare);
    let mut merged: Vec<Range> = Vec::with_capacity(ranges.len());
    for range in ranges {
        match merged.last_mut() {
            Some(last) if overlaps(last, &range) || adjacent(last, &range) => *last = merge(last, &range, false)?,
            _ => merged.push(range),
        }
    }
    Ok(Multirange { type_oid, ranges: merged })
}

/// span returns the smallest range that covers a multirange's ranges, which is empty for an empty multirange.
pub fn span(multirange: &Multirange) -> Range {
    let range_oid = multirange_type(multirange.type_oid).map_or(0, |t| t.range);
    match (multirange.ranges.first(), multirange.ranges.last()) {
        (Some(first), Some(last)) => {
            Range { type_oid: range_oid, empty: false, lower: first.lower.clone(), upper: last.upper.clone() }
        }
        _ => empty(range_oid),
    }
}

/// multirange_minus returns the values of a multirange that another multirange leaves.
pub fn multirange_minus(left: &Multirange, right: &Multirange) -> Result<Multirange> {
    let mut parts = Vec::new();
    for range in &left.ranges {
        let mut pieces = vec![range.clone()];
        for other in &right.ranges {
            let mut next = Vec::new();
            for piece in &pieces {
                next.extend(split(piece, other)?);
            }
            pieces = next;
        }
        parts.extend(pieces);
    }
    normalize(left.type_oid, parts)
}

/// multirange_intersect returns the values two multiranges share.
pub fn multirange_intersect(left: &Multirange, right: &Multirange) -> Result<Multirange> {
    let mut parts = Vec::new();
    for l in &left.ranges {
        for r in &right.ranges {
            parts.push(intersect(l, r)?);
        }
    }
    normalize(left.type_oid, parts)
}

/// FLAGS are the bits of a range's binary format, as Postgres numbers them.
const EMPTY: u8 = 0x01;
const LOWER_INCLUSIVE: u8 = 0x02;
const UPPER_INCLUSIVE: u8 = 0x04;
const LOWER_INFINITE: u8 = 0x08;
const UPPER_INFINITE: u8 = 0x10;

/// flags returns a range's flags byte.
fn flags(range: &Range) -> u8 {
    if range.empty {
        return EMPTY;
    }
    let mut flags = 0;
    flags |= if range.lower.inclusive { LOWER_INCLUSIVE } else { 0 };
    flags |= if range.upper.inclusive { UPPER_INCLUSIVE } else { 0 };
    flags |= if range.lower.value.is_none() { LOWER_INFINITE } else { 0 };
    flags |= if range.upper.value.is_none() { UPPER_INFINITE } else { 0 };
    flags
}

/// send returns a range's binary format: its flags, then each finite bound's length and binary format.
pub fn send(range: &Range, bound_send: &dyn Fn(&Value) -> Result<Vec<u8>>) -> Result<Vec<u8>> {
    let mut out = vec![flags(range)];
    for value in [&range.lower.value, &range.upper.value].into_iter().flatten().filter(|_| !range.empty) {
        let bytes = bound_send(value)?;
        out.extend_from_slice(&(bytes.len() as i32).to_be_bytes());
        out.extend_from_slice(&bytes);
    }
    Ok(out)
}

/// send_multirange returns a multirange's binary format: its range count, then each range's length and binary format.
pub fn send_multirange(multirange: &Multirange, bound_send: &dyn Fn(&Value) -> Result<Vec<u8>>) -> Result<Vec<u8>> {
    let mut out = (multirange.ranges.len() as i32).to_be_bytes().to_vec();
    for range in &multirange.ranges {
        let bytes = send(range, bound_send)?;
        out.extend_from_slice(&(bytes.len() as i32).to_be_bytes());
        out.extend_from_slice(&bytes);
    }
    Ok(out)
}

/// receive reads a range from its binary format, reading each bound with `bound_receive`.
pub fn receive(type_oid: u32, bytes: &[u8], bound_receive: &dyn Fn(&[u8]) -> Result<Value>) -> Result<Range> {
    let invalid = || PgError::new(code::INVALID_BINARY_REPRESENTATION, "incorrect binary data format");
    let (&flags, mut rest) = bytes.split_first().ok_or_else(invalid)?;
    if flags & EMPTY != 0 {
        return Ok(empty(type_oid));
    }
    let mut bound = |unbounded: bool, inclusive: bool| -> Result<Bound> {
        if unbounded {
            return Ok(infinite());
        }
        let length = rest.get(..4).ok_or_else(invalid)?;
        let length = i32::from_be_bytes(length.try_into().map_err(|_| invalid())?) as usize;
        let data = rest.get(4..4 + length).ok_or_else(invalid)?;
        rest = &rest[4 + length..];
        Ok(Bound { value: Some(bound_receive(data)?), inclusive })
    };
    let lower = bound(flags & LOWER_INFINITE != 0, flags & LOWER_INCLUSIVE != 0)?;
    let upper = bound(flags & UPPER_INFINITE != 0, flags & UPPER_INCLUSIVE != 0)?;
    make(type_oid, lower, upper)
}

/// receive_multirange reads a multirange from its binary format, reading each bound with `bound_receive`.
pub fn receive_multirange(
    type_oid: u32,
    bytes: &[u8],
    bound_receive: &dyn Fn(&[u8]) -> Result<Value>,
) -> Result<Multirange> {
    let invalid = || PgError::new(code::INVALID_BINARY_REPRESENTATION, "incorrect binary data format");
    let range_oid = multirange_type(type_oid).map_or(0, |t| t.range);
    let read_length = |at: usize| -> Result<usize> {
        let bytes: [u8; 4] = bytes.get(at..at + 4).ok_or_else(invalid)?.try_into().map_err(|_| invalid())?;
        Ok(i32::from_be_bytes(bytes) as usize)
    };
    let count = read_length(0)?;
    let (mut ranges, mut at) = (Vec::with_capacity(count), 4);
    for _ in 0..count {
        let length = read_length(at)?;
        let data = bytes.get(at + 4..at + 4 + length).ok_or_else(invalid)?;
        ranges.push(receive(range_oid, data, bound_receive)?);
        at += 4 + length;
    }
    normalize(type_oid, ranges)
}
