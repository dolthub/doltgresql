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

//! The money and pg_lsn types, read and written as Postgres' cash.c, in the C locale, and pg_lsn.c do, and stored as
//! their binary formats.

use crate::error::{PgError, Result, code};
use crate::extensions::BaseType;

/// MONEY is the money type, an amount in cents.
pub const MONEY: BaseType = BaseType {
    name: "money",
    input: |text, _| cash_in(text).map(|cents| cents.to_be_bytes().to_vec()),
    output: |bytes| cash_out(i64::from_be_bytes(bytes.try_into().unwrap_or_default())),
    receive: |bytes, _| eight(bytes),
    send: <[u8]>::to_vec,
    typmod_in: |_| Ok(-1),
    typmod: |_, _| Ok(()),
    compare: |l, r| {
        i64::from_be_bytes(l.try_into().unwrap_or_default()).cmp(&i64::from_be_bytes(r.try_into().unwrap_or_default()))
    },
    vector: None,
};

/// PG_LSN is the pg_lsn type, a write-ahead log position.
pub const PG_LSN: BaseType = BaseType {
    name: "pg_lsn",
    input: |text, _| lsn_in(text).map(|lsn| lsn.to_be_bytes().to_vec()),
    output: |bytes| {
        let lsn = u64::from_be_bytes(bytes.try_into().unwrap_or_default());
        format!("{:X}/{:X}", lsn >> 32, lsn & 0xffff_ffff)
    },
    receive: |bytes, _| eight(bytes),
    send: <[u8]>::to_vec,
    typmod_in: |_| Ok(-1),
    typmod: |_, _| Ok(()),
    compare: <[u8]>::cmp,
    vector: None,
};

/// eight reads the binary format of an eight-byte value.
fn eight(bytes: &[u8]) -> Result<Vec<u8>> {
    match bytes.len() {
        8 => Ok(bytes.to_vec()),
        n if n < 8 => Err(PgError::new(code::PROTOCOL_VIOLATION, "insufficient data left in message")),
        _ => Err(PgError::new(code::INVALID_BINARY_REPRESENTATION, "incorrect binary data format")),
    }
}

/// cash_in reads an amount of money with an optional sign, parentheses, currency symbol, and thousands separators,
/// rounding past two decimal places, as cash_in does in the C locale.
fn cash_in(text: &str) -> Result<i64> {
    let out_of_range =
        || PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, format!("value \"{text}\" is out of range for type money"));
    let bytes = text.as_bytes();
    let mut i = 0;
    let skip_space = |i: &mut usize| {
        while bytes.get(*i).is_some_and(u8::is_ascii_whitespace) {
            *i += 1;
        }
    };
    let eat = |i: &mut usize, byte: u8| {
        let found = bytes.get(*i) == Some(&byte);
        if found {
            *i += 1;
        }
        found
    };
    skip_space(&mut i);
    eat(&mut i, b'$');
    skip_space(&mut i);
    let mut negative = false;
    if eat(&mut i, b'-') || eat(&mut i, b'(') {
        negative = true;
    } else {
        eat(&mut i, b'+');
    }
    skip_space(&mut i);
    eat(&mut i, b'$');
    skip_space(&mut i);
    let (mut value, mut seen_dot, mut decimals) = (0i64, false, 0);
    while let Some(&ch) = bytes.get(i) {
        if ch.is_ascii_digit() && (!seen_dot || decimals < 2) {
            value = value.checked_mul(10).and_then(|v| v.checked_sub(i64::from(ch - b'0'))).ok_or_else(out_of_range)?;
            if seen_dot {
                decimals += 1;
            }
        } else if ch == b'.' && !seen_dot {
            seen_dot = true;
        } else if ch != b',' {
            break;
        }
        i += 1;
    }
    if bytes.get(i).is_some_and(|&ch| (b'5'..=b'9').contains(&ch)) {
        value = value.checked_sub(1).ok_or_else(out_of_range)?;
    }
    for _ in decimals..2 {
        value = value.checked_mul(10).ok_or_else(out_of_range)?;
    }
    while bytes.get(i).is_some_and(u8::is_ascii_digit) {
        i += 1;
    }
    while let Some(&ch) = bytes.get(i) {
        match ch {
            b')' => i += 1,
            b'-' => {
                negative = true;
                i += 1;
            }
            b'+' | b'$' => i += 1,
            _ if ch.is_ascii_whitespace() => i += 1,
            _ => {
                return Err(PgError::new(
                    code::INVALID_TEXT_REPRESENTATION,
                    format!("invalid input syntax for type money: \"{text}\""),
                ));
            }
        }
    }
    match negative {
        true => Ok(value),
        false => value.checked_neg().ok_or_else(out_of_range),
    }
}

/// cash_out writes an amount of money with a currency symbol, thousands separators, and two decimal places, as
/// cash_out does in the C locale.
fn cash_out(cents: i64) -> String {
    let mut digits = cents.unsigned_abs().to_string();
    if digits.len() < 3 {
        digits = format!("{digits:0>3}");
    }
    let (whole, fraction) = digits.split_at(digits.len() - 2);
    let mut grouped = String::new();
    for (i, ch) in whole.chars().enumerate() {
        if i > 0 && (whole.len() - i) % 3 == 0 {
            grouped.push(',');
        }
        grouped.push(ch);
    }
    let sign = if cents < 0 { "-" } else { "" };
    format!("{sign}${grouped}.{fraction}")
}

/// lsn_in reads a write-ahead log position as two hex numbers of at most eight digits around a slash.
fn lsn_in(text: &str) -> Result<u64> {
    let invalid =
        || PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("invalid input syntax for type pg_lsn: \"{text}\""));
    let (high, low) = text.split_once('/').ok_or_else(invalid)?;
    let part = |s: &str| {
        let valid = (1..=8).contains(&s.len()) && s.bytes().all(|b| b.is_ascii_hexdigit());
        valid.then(|| u64::from_str_radix(s, 16).ok()).flatten().ok_or_else(invalid)
    };
    Ok(part(high)? << 32 | part(low)?)
}
