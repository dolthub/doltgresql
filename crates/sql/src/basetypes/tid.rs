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

//! The tid type, a tuple's block and offset, a port of Postgres' tid.c, stored as its binary format as Go stores it.

use crate::error::{PgError, Result, code};
use crate::extensions::BaseType;

/// TID is the tid type.
pub const TID: BaseType = BaseType {
    name: "tid",
    input: |text, _| tid_in(text),
    output: |bytes| match bytes {
        [b0, b1, b2, b3, o0, o1] => {
            format!("({},{})", u32::from_be_bytes([*b0, *b1, *b2, *b3]), u16::from_be_bytes([*o0, *o1]))
        }
        _ => String::new(),
    },
    receive: |bytes, _| match bytes.len() {
        6 => Ok(bytes.to_vec()),
        n if n < 6 => Err(PgError::new(code::PROTOCOL_VIOLATION, "insufficient data left in message")),
        _ => Err(PgError::new(code::INVALID_BINARY_REPRESENTATION, "incorrect binary data format")),
    },
    send: <[u8]>::to_vec,
    typmod_in: |_| Ok(-1),
    typmod: |_, _| Ok(()),
    compare: <[u8]>::cmp,
    vector: None,
};

/// strtoul reads an unsigned number as C's strtoul does, after spaces and an optional sign, with a negative number
/// wrapping around, returning the number, the text after it, and whether it overflowed.
fn strtoul(text: &str) -> (u64, &str, bool) {
    let trimmed = text.trim_start_matches([' ', '\t', '\n', '\r', '\x0b', '\x0c']);
    let (negative, rest) = match trimmed.as_bytes().first() {
        Some(b'-') => (true, &trimmed[1..]),
        Some(b'+') => (false, &trimmed[1..]),
        _ => (false, trimmed),
    };
    let digits = rest.bytes().take_while(u8::is_ascii_digit).count();
    if digits == 0 {
        return (0, text, false);
    }
    let mut value: u64 = 0;
    let mut overflow = false;
    for b in rest[..digits].bytes() {
        match value.checked_mul(10).and_then(|v| v.checked_add(u64::from(b - b'0'))) {
            Some(v) => value = v,
            None => overflow = true,
        }
    }
    let value = if overflow {
        u64::MAX
    } else if negative {
        value.wrapping_neg()
    } else {
        value
    };
    (value, &rest[digits..], overflow)
}

/// tid_in reads a tid as Postgres' tidin does: `(block,offset)`.
fn tid_in(text: &str) -> Result<Vec<u8>> {
    let invalid =
        || PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("invalid input syntax for type tid: \"{text}\""));
    let mut coordinates = Vec::with_capacity(2);
    for (i, c) in text.char_indices() {
        if coordinates.len() == 2 || c == ')' {
            break;
        }
        if c == ',' || (c == '(' && coordinates.is_empty()) {
            coordinates.push(&text[i + 1..]);
        }
    }
    let [first, second] = coordinates.as_slice() else { return Err(invalid()) };
    let (block, rest, overflow) = strtoul(first);
    if overflow || !rest.starts_with(',') {
        return Err(invalid());
    }
    let truncated = block as u32;
    if block != u64::from(truncated) && block != (truncated as i32 as i64) as u64 {
        return Err(invalid());
    }
    let (offset, rest, overflow) = strtoul(second);
    if overflow || !rest.starts_with(')') || offset > u64::from(u16::MAX) {
        return Err(invalid());
    }
    Ok([truncated.to_be_bytes().as_slice(), &(offset as u16).to_be_bytes()].concat())
}
