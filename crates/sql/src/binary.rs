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

//! The text formats of the bytea, uuid, bit, and bit varying types.

use std::cell::Cell;

use crate::catalog::ColumnType;
use crate::error::{PgError, Result, code};
use crate::oid;

thread_local! {
    /// ESCAPE_OUTPUT is whether bytea values on this thread print in the escape format rather than hex.
    static ESCAPE_OUTPUT: Cell<bool> = const { Cell::new(false) };
}

/// install_output makes bytea values on this thread print in the format that `bytea_output` names.
pub fn install_output(bytea_output: &str) {
    ESCAPE_OUTPUT.with(|e| e.set(bytea_output.eq_ignore_ascii_case("escape")));
}

/// hex_digit returns the value of a hexadecimal digit.
fn hex_digit(c: u8) -> Option<u8> {
    (c as char).to_digit(16).map(|d| d as u8)
}

/// decode_hex reads hexadecimal digit pairs as Postgres' hex_decode does, skipping whitespace between pairs.
pub fn decode_hex(text: &str) -> Result<Vec<u8>> {
    let invalid = |c: char| PgError::new(code::INVALID_PARAMETER_VALUE, format!("invalid hexadecimal digit: \"{c}\""));
    let bytes = text.as_bytes();
    let mut out = Vec::with_capacity(bytes.len() / 2);
    let mut i = 0;
    while i < bytes.len() {
        if matches!(bytes[i], b' ' | b'\n' | b'\t' | b'\r') {
            i += 1;
            continue;
        }
        let high = hex_digit(bytes[i]).ok_or_else(|| invalid(text[i..].chars().next().unwrap_or(' ')))?;
        let Some(&next) = bytes.get(i + 1) else {
            return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "invalid hexadecimal data: odd number of digits"));
        };
        let low = hex_digit(next).ok_or_else(|| invalid(text[i + 1..].chars().next().unwrap_or(' ')))?;
        out.push(high << 4 | low);
        i += 2;
    }
    Ok(out)
}

/// parse_bytea reads a bytea value as Postgres' byteain does, in the hex format or the escape format.
pub fn parse_bytea(text: &str) -> Result<Vec<u8>> {
    if let Some(hex) = text.strip_prefix("\\x") {
        return decode_hex(hex);
    }
    decode_escape(text)
        .ok_or_else(|| PgError::new(code::INVALID_TEXT_REPRESENTATION, "invalid input syntax for type bytea"))
}

/// decode_escape reads the escape format, where a backslash starts an octal escape or escapes a backslash.
pub fn decode_escape(text: &str) -> Option<Vec<u8>> {
    let bytes = text.as_bytes();
    let mut out = Vec::with_capacity(bytes.len());
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] != b'\\' {
            out.push(bytes[i]);
            i += 1;
        } else if bytes.get(i + 1) == Some(&b'\\') {
            out.push(b'\\');
            i += 2;
        } else {
            let digits = bytes.get(i + 1..i + 4)?;
            let valid = (b'0'..=b'3').contains(&digits[0]) && digits[1..].iter().all(|d| (b'0'..=b'7').contains(d));
            if !valid {
                return None;
            }
            out.push((digits[0] - b'0') << 6 | (digits[1] - b'0') << 3 | (digits[2] - b'0'));
            i += 4;
        }
    }
    Some(out)
}

/// encode_hex writes bytes as lowercase hexadecimal digits.
pub fn encode_hex(bytes: &[u8]) -> String {
    const DIGITS: &[u8; 16] = b"0123456789abcdef";
    let mut out = String::with_capacity(bytes.len() * 2);
    for &b in bytes {
        out.push(DIGITS[(b >> 4) as usize] as char);
        out.push(DIGITS[(b & 0xf) as usize] as char);
    }
    out
}

/// encode_escape writes bytes in the escape format, with octal escapes for bytes that are not printable ASCII.
pub fn encode_escape(bytes: &[u8]) -> String {
    let mut out = String::with_capacity(bytes.len());
    for &b in bytes {
        match b {
            b'\\' => out.push_str("\\\\"),
            0x20..=0x7e => out.push(b as char),
            _ => out.push_str(&format!("\\{b:03o}")),
        }
    }
    out
}

/// format_bytea prints a bytea value in the session's `bytea_output` format.
pub fn format_bytea(bytes: &[u8]) -> String {
    if ESCAPE_OUTPUT.with(Cell::get) { encode_escape(bytes) } else { format!("\\x{}", encode_hex(bytes)) }
}

/// parse_uuid reads a uuid as Postgres' uuid_in does: 32 hexadecimal digits, optionally in braces, with an optional
/// hyphen after any group of four digits.
pub fn parse_uuid(text: &str) -> Result<[u8; 16]> {
    let invalid =
        || PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("invalid input syntax for type uuid: \"{text}\""));
    let mut rest = text.as_bytes();
    let braces = rest.first() == Some(&b'{');
    if braces {
        rest = &rest[1..];
    }
    let mut out = [0u8; 16];
    for (i, byte) in out.iter_mut().enumerate() {
        let (Some(high), Some(low)) =
            (rest.first().and_then(|&c| hex_digit(c)), rest.get(1).and_then(|&c| hex_digit(c)))
        else {
            return Err(invalid());
        };
        *byte = high << 4 | low;
        rest = &rest[2..];
        if rest.first() == Some(&b'-') && i % 2 == 1 && i < 15 {
            rest = &rest[1..];
        }
    }
    if braces {
        rest = rest.strip_prefix(b"}").ok_or_else(invalid)?;
    }
    if !rest.is_empty() {
        return Err(invalid());
    }
    Ok(out)
}

/// format_uuid prints a uuid in its canonical hyphenated form.
pub fn format_uuid(bytes: &[u8; 16]) -> String {
    let hex = encode_hex(bytes);
    format!("{}-{}-{}-{}-{}", &hex[..8], &hex[8..12], &hex[12..16], &hex[16..20], &hex[20..])
}

/// parse_bits reads a bit string as Postgres' bit_in does: binary digits, or hexadecimal digits after an X.
pub fn parse_bits(text: &str) -> Result<String> {
    let (hex, digits) = match text.as_bytes().first() {
        Some(b'x' | b'X') => (true, &text[1..]),
        Some(b'b' | b'B') => (false, &text[1..]),
        _ => (false, text),
    };
    let mut out = String::with_capacity(if hex { digits.len() * 4 } else { digits.len() });
    for c in digits.chars() {
        if hex {
            let value = c.to_digit(16).ok_or_else(|| {
                PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("\"{c}\" is not a valid hexadecimal digit"))
            })?;
            out.push_str(&format!("{value:04b}"));
        } else if c == '0' || c == '1' {
            out.push(c);
        } else {
            return Err(PgError::new(
                code::INVALID_TEXT_REPRESENTATION,
                format!("\"{c}\" is not a valid binary digit"),
            ));
        }
    }
    Ok(out)
}

/// fit_bits applies a bit or bit varying length to a bit string, where an explicit cast pads or truncates and any
/// other conversion requires the length to fit.
pub fn fit_bits(bits: String, to: ColumnType, explicit: bool) -> Result<String> {
    if to.modifier <= 0 {
        return Ok(bits);
    }
    let length = to.modifier as usize;
    if to.oid == oid::BIT && bits.len() != length {
        if !explicit {
            return Err(PgError::new(
                code::STRING_DATA_LENGTH_MISMATCH,
                format!("bit string length {} does not match type bit({length})", bits.len()),
            ));
        }
        let mut bits = bits;
        bits.truncate(length);
        return Ok(format!("{bits:0<length$}"));
    }
    if to.oid == oid::VARBIT && bits.len() > length {
        if !explicit {
            return Err(PgError::new(
                code::STRING_DATA_RIGHT_TRUNCATION,
                format!("bit string too long for type bit varying({length})"),
            ));
        }
        return Ok(bits[..length].to_string());
    }
    Ok(bits)
}

/// pack_bits packs a bit string into bytes, most significant bit first, padding the last byte with zeros.
pub fn pack_bits(bits: &str) -> Vec<u8> {
    bits.as_bytes()
        .chunks(8)
        .map(|chunk| chunk.iter().enumerate().fold(0u8, |byte, (i, &c)| byte | (((c == b'1') as u8) << (7 - i))))
        .collect()
}

/// unpack_bits reads `length` bits from packed bytes, most significant bit first.
pub fn unpack_bits(bytes: &[u8], length: usize) -> String {
    (0..length).map(|i| if bytes.get(i / 8).is_some_and(|b| b & (0x80 >> (i % 8)) != 0) { '1' } else { '0' }).collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn values_read_as_postgres_reads_them() {
        assert_eq!(parse_bytea("\\x 01 02ff").unwrap(), vec![1, 2, 255]);
        assert_eq!(parse_bytea("a\\001\\\\").unwrap(), b"a\x01\\".to_vec());
        assert_eq!(parse_bytea("\\x012").unwrap_err().message, "invalid hexadecimal data: odd number of digits");
        assert_eq!(parse_bytea("a\\b").unwrap_err().code, "22P02");
        let uuid = parse_uuid("{A0EEBC999C0B4EF8BB6D6BB9BD380A11}").unwrap();
        assert_eq!(format_uuid(&uuid), "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11");
        assert_eq!(parse_uuid("a0ee-bc99-9c0b-4ef8-bb6d-6bb9-bd38-0a11").unwrap(), uuid);
        assert!(parse_uuid("a0eebc99-9c0b4-ef8-bb6d-6bb9bd380a11").is_err());
        assert_eq!(parse_bits("X1F").unwrap(), "00011111");
        assert_eq!(fit_bits("10".into(), ColumnType { oid: oid::BIT, modifier: 3 }, true).unwrap(), "100");
        assert_eq!(unpack_bits(&pack_bits("1011001"), 7), "1011001");
    }
}
