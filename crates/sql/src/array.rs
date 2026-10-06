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

//! Arrays: their text and binary formats, and how Doltgres stores them.

use std::cmp::Ordering;

use crate::catalog::builtin_type;
use crate::error::{PgError, Result, code};
use crate::expr::{ArrayOp, compare_values};
use crate::types::Value;

/// MAX_DIMENSIONS is the most dimensions an array may have.
const MAX_DIMENSIONS: usize = 6;

/// Array is an array value: its element type, its dimensions as lengths and lower bounds, and its elements in row
/// order. An empty array has no dimensions.
#[derive(Clone, Debug, PartialEq)]
pub struct Array {
    pub element: u32,
    pub dims: Vec<(i32, i32)>,
    pub values: Vec<Value>,
}

impl Array {
    /// one_dimensional returns a one-dimensional array of the elements, which is empty without them.
    pub fn one_dimensional(element: u32, values: Vec<Value>) -> Array {
        let dims = if values.is_empty() { Vec::new() } else { vec![(values.len() as i32, 1)] };
        Array { element, dims, values }
    }

    /// array_type returns the OID of the array type of the elements.
    pub fn array_type(&self) -> u32 {
        builtin_type(self.element).map_or(0, |t| t.array)
    }
}

/// malformed returns Postgres' error for array text it cannot read.
fn malformed(text: &str, detail: &str) -> PgError {
    PgError {
        detail: Some(detail.to_string()),
        ..PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("malformed array literal: \"{text}\""))
    }
}

/// is_array_space reports whether a byte is whitespace as Postgres' array_isspace sees it.
fn is_array_space(c: u8) -> bool {
    matches!(c, b' ' | b'\t' | b'\n' | b'\r' | b'\x0b' | b'\x0c')
}

/// State is a state of Postgres' array literal parser.
#[derive(Clone, Copy, PartialEq, Eq)]
enum State {
    NoLevel,
    LevelStarted,
    ElemStarted,
    QuotedElemStarted,
    QuotedElemCompleted,
    ElemDelimited,
    LevelCompleted,
    LevelDelimited,
}

/// count_dimensions checks the structure of array text and returns the length of each dimension, empty for an empty
/// array, as Postgres 15's ArrayCount does.
fn count_dimensions(original: &str, text: &[u8], delimiter: u8) -> Result<Vec<i32>> {
    let mut nest_level = 0usize;
    let mut ndim = 1usize;
    let mut temp = [0i32; MAX_DIMENSIONS];
    let mut nelems = [1i32; MAX_DIMENSIONS];
    let mut nelems_last = [0i32; MAX_DIMENSIONS];
    let mut in_quotes = false;
    let mut empty = true;
    let mut state = State::NoLevel;
    let mut i = 0;
    let unexpected = |c: char| malformed(original, &format!("Unexpected \"{c}\" character."));
    loop {
        let mut item_done = false;
        let mut end_of_array = false;
        while !item_done {
            if matches!(state, State::ElemStarted | State::QuotedElemStarted) {
                empty = false;
            }
            let Some(&c) = text.get(i) else { return Err(malformed(original, "Unexpected end of input.")) };
            match c {
                b'\\' => {
                    if !matches!(
                        state,
                        State::LevelStarted | State::ElemStarted | State::QuotedElemStarted | State::ElemDelimited
                    ) {
                        return Err(unexpected('\\'));
                    }
                    if state != State::QuotedElemStarted {
                        state = State::ElemStarted;
                    }
                    if i + 1 < text.len() {
                        i += 1;
                    } else {
                        return Err(malformed(original, "Unexpected end of input."));
                    }
                }
                b'"' => {
                    if !matches!(state, State::LevelStarted | State::QuotedElemStarted | State::ElemDelimited) {
                        return Err(malformed(original, "Unexpected array element."));
                    }
                    in_quotes = !in_quotes;
                    state = if in_quotes { State::QuotedElemStarted } else { State::QuotedElemCompleted };
                }
                b'{' if !in_quotes => {
                    if !matches!(state, State::NoLevel | State::LevelStarted | State::LevelDelimited) {
                        return Err(unexpected('{'));
                    }
                    state = State::LevelStarted;
                    if nest_level >= MAX_DIMENSIONS {
                        return Err(PgError::new(
                            code::PROGRAM_LIMIT_EXCEEDED,
                            format!(
                                "number of array dimensions ({}) exceeds the maximum allowed ({MAX_DIMENSIONS})",
                                nest_level + 1
                            ),
                        ));
                    }
                    temp[nest_level] = 0;
                    nest_level += 1;
                    ndim = ndim.max(nest_level);
                }
                b'}' if !in_quotes => {
                    let allowed =
                        matches!(state, State::ElemStarted | State::QuotedElemCompleted | State::LevelCompleted)
                            || (nest_level == 1 && state == State::LevelStarted);
                    if !allowed {
                        return Err(unexpected('}'));
                    }
                    state = State::LevelCompleted;
                    if nest_level == 0 {
                        return Err(malformed(original, "Unmatched \"}\" character."));
                    }
                    nest_level -= 1;
                    if nelems_last[nest_level] != 0 && nelems[nest_level] != nelems_last[nest_level] {
                        return Err(malformed(
                            original,
                            "Multidimensional arrays must have sub-arrays with matching dimensions.",
                        ));
                    }
                    nelems_last[nest_level] = nelems[nest_level];
                    nelems[nest_level] = 1;
                    if nest_level == 0 {
                        end_of_array = true;
                        item_done = true;
                    } else {
                        temp[nest_level - 1] += 1;
                    }
                }
                c if !in_quotes && c == delimiter => {
                    if !matches!(state, State::ElemStarted | State::QuotedElemCompleted | State::LevelCompleted) {
                        return Err(unexpected(delimiter as char));
                    }
                    state = if state == State::LevelCompleted { State::LevelDelimited } else { State::ElemDelimited };
                    item_done = true;
                    nelems[nest_level.saturating_sub(1)] += 1;
                }
                c if !in_quotes && !is_array_space(c) => {
                    if !matches!(state, State::LevelStarted | State::ElemStarted | State::ElemDelimited) {
                        return Err(malformed(original, "Unexpected array element."));
                    }
                    state = State::ElemStarted;
                }
                _ => {}
            }
            if !item_done {
                i += 1;
            }
        }
        temp[ndim - 1] += 1;
        i += 1;
        if end_of_array {
            break;
        }
    }
    if text[i..].iter().any(|&c| !is_array_space(c)) {
        return Err(malformed(original, "Junk after closing right brace."));
    }
    if empty {
        return Ok(Vec::new());
    }
    Ok(temp[..ndim].to_vec())
}

/// read_elements extracts the element texts of validated array text, with None for NULL, as Postgres' ReadArrayStr
/// does.
fn read_elements(text: &[u8], delimiter: u8, count: usize) -> Vec<Option<String>> {
    let mut out = Vec::with_capacity(count);
    let mut i = 0;
    let mut nest_level = 0;
    let mut end_of_array = false;
    while !end_of_array {
        let mut item = Vec::new();
        let mut item_end = 0;
        let mut in_quotes = false;
        let mut leading_space = true;
        let mut has_quoting = false;
        let mut item_done = false;
        let mut started = false;
        while !item_done {
            let Some(&c) = text.get(i) else { break };
            match c {
                b'\\' => {
                    i += 1;
                    if let Some(&escaped) = text.get(i) {
                        item.push(escaped);
                    }
                    item_end = item.len();
                    leading_space = false;
                    has_quoting = true;
                    started = true;
                    i += 1;
                }
                b'"' => {
                    in_quotes = !in_quotes;
                    if in_quotes {
                        leading_space = false;
                    } else {
                        item_end = item.len();
                    }
                    has_quoting = true;
                    started = true;
                    i += 1;
                }
                b'{' if !in_quotes => {
                    nest_level += 1;
                    i += 1;
                }
                b'}' if !in_quotes => {
                    nest_level -= 1;
                    if nest_level == 0 {
                        end_of_array = true;
                        item_done = true;
                    }
                    i += 1;
                }
                c if !in_quotes && c == delimiter => {
                    item_done = true;
                    i += 1;
                }
                c if !in_quotes && is_array_space(c) => {
                    if !leading_space {
                        item.push(c);
                    }
                    i += 1;
                }
                c => {
                    item.push(c);
                    if !in_quotes {
                        leading_space = false;
                    }
                    item_end = item.len();
                    started = true;
                    i += 1;
                }
            }
        }
        if !started {
            continue;
        }
        item.truncate(item_end);
        let text = String::from_utf8_lossy(&item).into_owned();
        out.push(if !has_quoting && text.eq_ignore_ascii_case("NULL") { None } else { Some(text) });
    }
    out
}

/// parse reads array text in Postgres' format: optional dimension bounds, then nested braces of elements, which each
/// read as the element type.
pub fn parse(text: &str, element: u32, read: &dyn Fn(&str) -> Result<Value>) -> Result<Array> {
    let delimiter = b',';
    let bytes = text.as_bytes();
    let mut i = 0;
    while i < bytes.len() && is_array_space(bytes[i]) {
        i += 1;
    }
    let mut lower_bounds: Vec<(i32, i32)> = Vec::new();
    while bytes.get(i) == Some(&b'[') {
        let end = text[i..]
            .find(']')
            .map(|e| e + i)
            .ok_or_else(|| malformed(text, "Missing \"]\" after array dimensions."))?;
        let spec = &text[i + 1..end];
        let (lower, upper) = match spec.split_once(':') {
            Some((l, u)) => (l.trim().parse::<i32>(), u.trim().parse::<i32>()),
            None => (Ok(1), spec.trim().parse::<i32>()),
        };
        let (Ok(lower), Ok(upper)) = (lower, upper) else {
            return Err(malformed(text, "\"[\" must introduce explicitly-specified array dimensions."));
        };
        if upper < lower {
            return Err(PgError::new(code::ARRAY_SUBSCRIPT_ERROR, "upper bound cannot be less than lower bound"));
        }
        lower_bounds.push((upper - lower + 1, lower));
        i = end + 1;
    }
    if !lower_bounds.is_empty() {
        while i < bytes.len() && is_array_space(bytes[i]) {
            i += 1;
        }
        if bytes.get(i) != Some(&b'=') {
            return Err(malformed(text, "Missing \"=\" after array dimensions."));
        }
        i += 1;
        while i < bytes.len() && is_array_space(bytes[i]) {
            i += 1;
        }
    }
    if bytes.get(i) != Some(&b'{') {
        return Err(malformed(text, "Array value must start with \"{\" or dimension information."));
    }
    let body = &bytes[i..];
    let lengths = count_dimensions(text, body, delimiter)?;
    if lengths.is_empty() {
        return Ok(Array { element, dims: Vec::new(), values: Vec::new() });
    }
    let dims: Vec<(i32, i32)> = if lower_bounds.is_empty() {
        lengths.iter().map(|&n| (n, 1)).collect()
    } else {
        if lower_bounds.len() != lengths.len() || lower_bounds.iter().zip(&lengths).any(|((n, _), l)| n != l) {
            return Err(malformed(text, "Specified array dimensions do not match array contents."));
        }
        lower_bounds
    };
    let count: usize = dims.iter().map(|(n, _)| *n as usize).product();
    let elements = read_elements(body, delimiter, count);
    let values = elements
        .into_iter()
        .map(|e| match e {
            None => Ok(Value::Null),
            Some(text) => read(&text),
        })
        .collect::<Result<Vec<_>>>()?;
    Ok(Array { element, dims, values })
}

/// quote_element returns an element's text as array output writes it, quoted and escaped when it must be.
fn quote_element(text: &str) -> String {
    let needs = text.is_empty()
        || text.eq_ignore_ascii_case("NULL")
        || text.bytes().any(|c| matches!(c, b'"' | b'\\' | b'{' | b'}' | b',') || is_array_space(c));
    if !needs {
        return text.to_string();
    }
    let mut out = String::with_capacity(text.len() + 2);
    out.push('"');
    for c in text.chars() {
        if c == '"' || c == '\\' {
            out.push('\\');
        }
        out.push(c);
    }
    out.push('"');
    out
}

/// format prints an array as Postgres' array_out does, with each element printed by the function.
pub fn format(array: &Array, print: &dyn Fn(&Value) -> String) -> String {
    if array.dims.is_empty() {
        return "{}".into();
    }
    let mut out = String::new();
    if array.dims.iter().any(|(_, lower)| *lower != 1) {
        for (n, lower) in &array.dims {
            out.push_str(&format!("[{}:{}]", lower, lower + n - 1));
        }
        out.push('=');
    }
    let mut index = 0;
    format_level(array, 0, &mut index, print, &mut out);
    out
}

/// format_level prints one level of an array's nesting.
fn format_level(array: &Array, level: usize, index: &mut usize, print: &dyn Fn(&Value) -> String, out: &mut String) {
    out.push('{');
    let n = array.dims[level].0 as usize;
    for i in 0..n {
        if i > 0 {
            out.push(',');
        }
        if level + 1 < array.dims.len() {
            format_level(array, level + 1, index, print, out);
        } else {
            match &array.values[*index] {
                Value::Null => out.push_str("NULL"),
                value => out.push_str(&quote_element(&print(value))),
            }
            *index += 1;
        }
    }
    out.push('}');
}

/// send writes an array in Postgres' binary format: dimensions, a NULL flag, the element type, bounds, and each
/// element's length and binary form.
pub fn send(array: &Array, element_send: &dyn Fn(&Value) -> Option<Vec<u8>>) -> Vec<u8> {
    let mut out = Vec::new();
    out.extend_from_slice(&(array.dims.len() as i32).to_be_bytes());
    out.extend_from_slice(&(array.values.iter().any(Value::is_null) as i32).to_be_bytes());
    out.extend_from_slice(&array.element.to_be_bytes());
    for (n, lower) in &array.dims {
        out.extend_from_slice(&n.to_be_bytes());
        out.extend_from_slice(&lower.to_be_bytes());
    }
    for value in &array.values {
        match element_send(value) {
            Some(bytes) => {
                out.extend_from_slice(&(bytes.len() as i32).to_be_bytes());
                out.extend_from_slice(&bytes);
            }
            None => out.extend_from_slice(&(-1i32).to_be_bytes()),
        }
    }
    out
}

/// receive reads Postgres' binary array format.
pub fn receive(bytes: &[u8], element_receive: &dyn Fn(u32, &[u8]) -> Result<Value>) -> Result<Array> {
    let invalid = || PgError::new(code::INVALID_BINARY_REPRESENTATION, "insufficient data left in message");
    let int = |at: usize| -> Result<i32> {
        Ok(i32::from_be_bytes(bytes.get(at..at + 4).ok_or_else(invalid)?.try_into().map_err(|_| invalid())?))
    };
    let ndim = int(0)? as usize;
    if ndim > MAX_DIMENSIONS {
        return Err(PgError::new(code::INVALID_BINARY_REPRESENTATION, format!("invalid number of dimensions: {ndim}")));
    }
    let element = int(8)? as u32;
    let mut dims = Vec::with_capacity(ndim);
    let mut at = 12;
    for _ in 0..ndim {
        dims.push((int(at)?, int(at + 4)?));
        at += 8;
    }
    let count: usize = dims.iter().map(|(n, _)| *n as usize).product();
    let mut values = Vec::with_capacity(count);
    for _ in 0..if ndim == 0 { 0 } else { count } {
        let length = int(at)?;
        at += 4;
        if length < 0 {
            values.push(Value::Null);
            continue;
        }
        let data = bytes.get(at..at + length as usize).ok_or_else(invalid)?;
        values.push(element_receive(element, data)?);
        at += length as usize;
    }
    Ok(Array { element, dims, values })
}

/// serialize writes an array as Doltgres stores it: the element count and the offset of each element, both
/// little-endian, then each element with a NULL flag, then the lengths of a multidimensional array's dimensions.
pub fn serialize(array: &Array, element_serialize: &dyn Fn(&Value) -> Result<Vec<u8>>) -> Result<Vec<u8>> {
    let count = array.values.len();
    let mut data = Vec::new();
    let mut offsets = Vec::with_capacity(count + 1);
    let mut offset = (4 + (count + 1) * 4) as u32;
    for value in &array.values {
        offsets.push(offset);
        if value.is_null() {
            data.push(1);
            offset += 1;
        } else {
            let bytes = element_serialize(value)?;
            data.push(0);
            offset += 1 + bytes.len() as u32;
            data.extend_from_slice(&bytes);
        }
    }
    offsets.push(offset);
    let mut out = (count as u32).to_le_bytes().to_vec();
    for o in offsets {
        out.extend_from_slice(&o.to_le_bytes());
    }
    out.extend_from_slice(&data);
    if array.dims.len() > 1 {
        for (n, _) in &array.dims {
            out.extend_from_slice(&(*n as u32).to_le_bytes());
        }
    }
    Ok(out)
}

/// deserialize reads an array as Doltgres stores it.
pub fn deserialize(bytes: &[u8], element: u32, element_deserialize: &dyn Fn(&[u8]) -> Result<Value>) -> Result<Array> {
    let corrupt = || PgError::internal("a stored array is corrupt");
    let word = |at: usize| -> Result<u32> {
        Ok(u32::from_le_bytes(bytes.get(at..at + 4).ok_or_else(corrupt)?.try_into().map_err(|_| corrupt())?))
    };
    let count = word(0)? as usize;
    let mut values = Vec::with_capacity(count);
    for i in 0..count {
        let (start, end) = (word((i + 1) * 4)? as usize, word((i + 2) * 4)? as usize);
        let item = bytes.get(start..end).ok_or_else(corrupt)?;
        if item.first() == Some(&1) {
            values.push(Value::Null);
        } else {
            values.push(element_deserialize(&item[1..])?);
        }
    }
    let dims_start = word((count + 1) * 4)? as usize;
    let mut dims: Vec<(i32, i32)> = bytes[dims_start..]
        .chunks(4)
        .filter(|c| c.len() == 4)
        .map(|c| (u32::from_le_bytes(c.try_into().unwrap()) as i32, 1))
        .collect();
    if dims.is_empty() && count > 0 {
        dims.push((count as i32, 1));
    }
    Ok(Array { element, dims, values })
}

/// nest builds a multidimensional array from arrays of matching dimensions, as an ARRAY constructor of arrays does,
/// treating NULL items as empty arrays.
pub fn nest(element: u32, items: Vec<Value>) -> Result<Array> {
    let mismatch = || {
        PgError::new(
            code::ARRAY_SUBSCRIPT_ERROR,
            "multidimensional arrays must have array expressions with matching dimensions",
        )
    };
    let mut inner_dims: Option<Vec<(i32, i32)>> = None;
    let mut count = 0;
    let mut values = Vec::new();
    for item in items {
        let dims = match &item {
            Value::Array(a) => a.dims.clone(),
            _ => Vec::new(),
        };
        match &inner_dims {
            Some(d) if *d != dims => return Err(mismatch()),
            Some(_) => {}
            None => inner_dims = Some(dims),
        }
        if let Value::Array(a) = item {
            values.extend(a.values);
        }
        count += 1;
    }
    let inner = inner_dims.unwrap_or_default();
    if inner.is_empty() {
        return Ok(Array { element, dims: Vec::new(), values: Vec::new() });
    }
    if inner.len() >= MAX_DIMENSIONS {
        return Err(PgError::new(
            code::PROGRAM_LIMIT_EXCEEDED,
            format!("number of array dimensions ({}) exceeds the maximum allowed ({MAX_DIMENSIONS})", inner.len() + 1),
        ));
    }
    let mut dims = vec![(count, 1)];
    dims.extend(inner);
    Ok(Array { element, dims, values })
}

/// element returns the element at the subscripts, or None when they are outside the array.
pub fn element<'v>(array: &'v Array, indexes: &[i32]) -> Option<&'v Value> {
    if indexes.len() != array.dims.len() {
        return None;
    }
    let mut offset = 0usize;
    for (&index, &(length, lower)) in indexes.iter().zip(&array.dims) {
        let position = index.checked_sub(lower)?;
        if position < 0 || position >= length {
            return None;
        }
        offset = offset * length as usize + position as usize;
    }
    array.values.get(offset)
}

/// slice returns the part of an array within the bounds, where a missing bound is the array's own, as a slice
/// subscript does, and the result's dimensions start at 1.
pub fn slice(array: &Array, bounds: &[(Option<i32>, Option<i32>)]) -> Array {
    let empty = Array { element: array.element, dims: Vec::new(), values: Vec::new() };
    if array.dims.is_empty() || bounds.len() > array.dims.len() {
        return empty;
    }
    let mut ranges = Vec::with_capacity(array.dims.len());
    for (i, &(length, lower)) in array.dims.iter().enumerate() {
        let (from, to) = bounds.get(i).copied().unwrap_or((None, None));
        let upper = lower + length - 1;
        let from = from.map_or(lower, |f| f.max(lower));
        let to = to.map_or(upper, |t| t.min(upper));
        if from > to {
            return empty;
        }
        ranges.push(((from - lower) as usize, (to - lower) as usize));
    }
    let mut values = Vec::new();
    let mut index = vec![0usize; ranges.len()];
    collect_slice(array, &ranges, 0, &mut index, &mut values);
    let dims = ranges.iter().map(|(f, t)| ((t - f + 1) as i32, 1)).collect();
    Array { element: array.element, dims, values }
}

/// collect_slice appends the elements of a slice in row order, one dimension at a time.
fn collect_slice(array: &Array, ranges: &[(usize, usize)], level: usize, index: &mut Vec<usize>, out: &mut Vec<Value>) {
    for i in ranges[level].0..=ranges[level].1 {
        index[level] = i;
        if level + 1 < ranges.len() {
            collect_slice(array, ranges, level + 1, index, out);
        } else {
            let offset = index.iter().zip(&array.dims).fold(0, |acc, (&i, &(n, _))| acc * n as usize + i);
            out.push(array.values[offset].clone());
        }
    }
}

/// compare orders arrays as Postgres does: by their elements, with NULLs last, then by their element counts,
/// dimension counts, dimension lengths, and lower bounds.
pub fn compare(left: &Array, right: &Array) -> Ordering {
    for (l, r) in left.values.iter().zip(&right.values) {
        let ordering = match (l.is_null(), r.is_null()) {
            (true, true) => Ordering::Equal,
            (true, false) => Ordering::Greater,
            (false, true) => Ordering::Less,
            (false, false) => compare_values(l, r),
        };
        if ordering != Ordering::Equal {
            return ordering;
        }
    }
    left.values
        .len()
        .cmp(&right.values.len())
        .then(left.dims.len().cmp(&right.dims.len()))
        .then_with(|| left.dims.iter().cmp(right.dims.iter()))
}

/// operate applies an array operator to two values.
pub fn operate(op: ArrayOp, left: Value, right: Value) -> Result<Value> {
    Ok(match op {
        ArrayOp::Concat => match (left, right) {
            (Value::Null, other) | (other, Value::Null) => other,
            (Value::Array(l), Value::Array(r)) => Value::Array(Box::new(concat(*l, *r)?)),
            _ => Value::Null,
        },
        ArrayOp::Append => match left {
            Value::Array(array) => Value::Array(Box::new(append(*array, right, false)?)),
            _ => Value::Null,
        },
        ArrayOp::Prepend => match right {
            Value::Array(array) => Value::Array(Box::new(append(*array, left, true)?)),
            _ => Value::Null,
        },
        ArrayOp::Contains | ArrayOp::ContainedBy | ArrayOp::Overlaps => {
            let (Value::Array(l), Value::Array(r)) = (left, right) else { return Ok(Value::Null) };
            let (haystack, needles) = if op == ArrayOp::ContainedBy { (r, l) } else { (l, r) };
            let found = |needle: &Value| {
                !needle.is_null()
                    && haystack.values.iter().any(|v| !v.is_null() && compare_values(v, needle) == Ordering::Equal)
            };
            Value::Bool(if op == ArrayOp::Overlaps {
                needles.values.iter().any(found)
            } else {
                needles.values.iter().all(found)
            })
        }
    })
}

/// concat joins two arrays as array_cat does: along their first dimension when they have the same number of
/// dimensions, or adding the smaller one as a new element of the larger one.
pub fn concat(left: Array, right: Array) -> Result<Array> {
    if left.dims.is_empty() {
        return Ok(Array { element: left.element, ..right });
    }
    if right.dims.is_empty() {
        return Ok(left);
    }
    let incompatible = |detail: String| PgError {
        detail: Some(detail),
        ..PgError::new(code::ARRAY_SUBSCRIPT_ERROR, "cannot concatenate incompatible arrays")
    };
    let (l, r) = (left.dims.len(), right.dims.len());
    let dims = if l == r {
        if left.dims[1..] != right.dims[1..] {
            return Err(incompatible(
                "Arrays with differing element dimensions are not compatible for concatenation.".into(),
            ));
        }
        let mut dims = left.dims.clone();
        dims[0].0 += right.dims[0].0;
        dims
    } else if l + 1 == r {
        if left.dims[..] != right.dims[1..] {
            return Err(incompatible("Arrays with differing dimensions are not compatible for concatenation.".into()));
        }
        let mut dims = right.dims.clone();
        dims[0].0 += 1;
        dims
    } else if l == r + 1 {
        if left.dims[1..] != right.dims[..] {
            return Err(incompatible("Arrays with differing dimensions are not compatible for concatenation.".into()));
        }
        let mut dims = left.dims.clone();
        dims[0].0 += 1;
        dims
    } else {
        return Err(incompatible(format!("Arrays of {l} and {r} dimensions are not compatible for concatenation.")));
    };
    let mut values = left.values;
    values.extend(right.values);
    Ok(Array { element: left.element, dims, values })
}

/// append adds an element to the end, or the start, of an array that is empty or one-dimensional.
pub fn append(mut array: Array, value: Value, prepend: bool) -> Result<Array> {
    match array.dims.len() {
        0 => return Ok(Array::one_dimensional(array.element, vec![value])),
        1 => {}
        _ => {
            return Err(PgError::new(code::DATA_EXCEPTION, "argument must be empty or one-dimensional array"));
        }
    }
    let (length, lower) = array.dims[0];
    if prepend {
        lower.checked_sub(1).ok_or_else(|| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "integer out of range"))?;
        array.values.insert(0, value);
    } else {
        lower
            .checked_add(length)
            .ok_or_else(|| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "integer out of range"))?;
        array.values.push(value);
    }
    array.dims[0].0 += 1;
    Ok(array)
}

/// is_array_type reports whether a type is an array type.
pub fn is_array_type(type_oid: u32) -> bool {
    match builtin_type(type_oid) {
        Some(t) => t.elem != 0 && t.definition.typ_category == b"A",
        None => crate::usertypes::get(type_oid).is_some_and(|t| t.is_array()),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::oid;

    /// text_array parses text into a text array.
    fn text_array(text: &str) -> Result<Array> {
        parse(text, oid::TEXT, &|s| Ok(Value::Text(s.to_string())))
    }

    #[test]
    fn literals_parse_and_print_as_postgres_does() {
        let a = text_array(r#"{"this", "is", null, "NULL", quoted , {} }"#);
        assert!(a.is_err());
        let a = text_array(r#"{"this", "is", null, "NULL", " sp "}"#).unwrap();
        assert_eq!(format(&a, &|v| v.output().unwrap_or_default()), r#"{this,is,NULL,"NULL"," sp "}"#);
        let a = text_array("{{1,2},{3,4}}").unwrap();
        assert_eq!(a.dims, vec![(2, 1), (2, 1)]);
        let a = text_array("[0:1]={a,b}").unwrap();
        assert_eq!(format(&a, &|v| v.output().unwrap_or_default()), "[0:1]={a,b}");
        assert_eq!(text_array("{}").unwrap().dims, vec![]);
    }

    #[test]
    fn malformed_literals_report_postgres_details() {
        for (text, detail) in [
            ("{{1,2},{3}}", "Multidimensional arrays must have sub-arrays with matching dimensions."),
            ("{{}}", "Unexpected \"}\" character."),
            ("{a,}", "Unexpected \"}\" character."),
            ("{a,b,c\"}", "Unexpected array element."),
            ("{a,b,c", "Unexpected end of input."),
            ("{a,b,\"c}", "Unexpected end of input."),
            ("{a\",b,c}", "Unexpected array element."),
            ("{1,{2}}", "Unexpected \"{\" character."),
            ("{\"abc\"\"\",\"def\"}", "Unexpected array element."),
            ("a,b,c}", "Array value must start with \"{\" or dimension information."),
            ("{a} b", "Junk after closing right brace."),
        ] {
            let err = text_array(text).unwrap_err();
            assert_eq!(err.detail.as_deref(), Some(detail), "{text}");
        }
    }

    #[test]
    fn values_nest_slice_and_concatenate_as_postgres_does() {
        let ints = |values: &[i32]| Array::one_dimensional(oid::INT4, values.iter().map(|&i| Value::Int4(i)).collect());
        let print = |a: &Array| format(a, &|v| v.output().unwrap_or_default());
        let square =
            nest(oid::INT4, vec![Value::Array(Box::new(ints(&[1, 2]))), Value::Array(Box::new(ints(&[3, 4])))]);
        let square = square.unwrap();
        assert_eq!(print(&square), "{{1,2},{3,4}}");
        assert!(
            nest(oid::INT4, vec![Value::Array(Box::new(ints(&[1]))), Value::Array(Box::new(ints(&[1, 2])))]).is_err()
        );
        assert_eq!(element(&square, &[2, 1]), Some(&Value::Int4(3)));
        assert_eq!(element(&square, &[3, 1]), None);
        assert_eq!(print(&slice(&square, &[(Some(1), Some(2)), (Some(2), None)])), "{{2},{4}}");
        assert_eq!(print(&slice(&ints(&[1, 2, 3]), &[(Some(5), None)])), "{}");
        assert_eq!(print(&concat(square.clone(), ints(&[5, 6])).unwrap()), "{{1,2},{3,4},{5,6}}");
        assert!(concat(square, ints(&[5])).is_err());
        assert_eq!(compare(&ints(&[1, 2]), &ints(&[1, 2, 0])), Ordering::Less);
        assert_eq!(compare(&ints(&[1, 3]), &ints(&[1, 2, 0])), Ordering::Greater);
    }
}
