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

use crate::decode::decode_value;
use crate::decode::reader::Reader;

/// is_array_space reports whether a byte is whitespace to Postgres' scanner, which forces quoting.
fn is_array_space(byte: u8) -> bool {
    matches!(byte, b' ' | b'\t' | b'\n' | b'\r' | 0x0b | 0x0c)
}

/// decode_nested decodes an element of an array, record, or range. A user-defined type has no known format, so it is
/// read as a composite when its bytes form exactly one, and as text otherwise, which covers composites and enums.
fn decode_nested(oid: u32, bytes: &[u8]) -> Result<String, String> {
    if oid >= 16384 {
        let mut r = Reader::new(bytes);
        if let Ok(text) = record_text(&mut r)
            && r.is_empty()
        {
            return Ok(text);
        }
        return String::from_utf8(bytes.to_vec())
            .map_err(|_| format!("cannot decode a value of user-defined type {oid}: {bytes:?}"));
    }
    let mut r = Reader::new(bytes);
    let text = decode_value(oid, &mut r)?;
    r.expect_end(oid)?;
    Ok(text)
}

/// array_text renders a binary array the way array_out does.
pub(crate) fn array_text(oid: u32, expected_element_oid: u32, r: &mut Reader<'_>) -> Result<String, String> {
    let dimension_count = r.i32()?;
    let has_nulls = r.i32()?;
    let element_oid = r.u32()?;
    if element_oid != expected_element_oid {
        return Err(format!("array type {oid} has element type {element_oid} instead of {expected_element_oid}"));
    }
    if !(0..=6).contains(&dimension_count) {
        return Err(format!("invalid array dimension count {dimension_count}"));
    }
    if has_nulls != 0 && has_nulls != 1 {
        return Err(format!("invalid array null flag {has_nulls}"));
    }
    if dimension_count == 0 {
        return Ok("{}".to_string());
    }
    let mut lengths = Vec::new();
    let mut lower_bounds = Vec::new();
    for _ in 0..dimension_count {
        let length = r.i32()?;
        if length < 0 {
            return Err(format!("invalid array length {length}"));
        }
        lengths.push(length as usize);
        lower_bounds.push(r.i32()?);
    }
    let delimiter = if element_oid == 603 { ';' } else { ',' };
    let total: usize = lengths.iter().product();
    let mut elements = Vec::with_capacity(total);
    let mut saw_null = false;
    for _ in 0..total {
        match r.length_prefixed()? {
            None => {
                saw_null = true;
                elements.push("NULL".to_string());
            }
            Some(bytes) => elements.push(quote_array_element(&decode_nested(element_oid, bytes)?, delimiter)),
        }
    }
    if saw_null && has_nulls == 0 {
        return Err("array contains NULL but its null flag is unset".to_string());
    }
    let mut text = String::new();
    if lower_bounds.iter().any(|bound| *bound != 1) {
        for (length, lower) in lengths.iter().zip(&lower_bounds) {
            text.push_str(&format!("[{}:{}]", lower, *lower as i64 + *length as i64 - 1));
        }
        text.push('=');
    }
    let mut index = 0;
    nest_array(&lengths, &elements, &mut index, delimiter, &mut text);
    Ok(text)
}

/// nest_array writes the elements of each dimension within braces.
fn nest_array(lengths: &[usize], elements: &[String], index: &mut usize, delimiter: char, text: &mut String) {
    text.push('{');
    for i in 0..lengths[0] {
        if i > 0 {
            text.push(delimiter);
        }
        if lengths.len() == 1 {
            text.push_str(&elements[*index]);
            *index += 1;
        } else {
            nest_array(&lengths[1..], elements, index, delimiter, text);
        }
    }
    text.push('}');
}

/// quote_array_element quotes an array element when array_out would.
fn quote_array_element(text: &str, delimiter: char) -> String {
    let needs_quotes = text.is_empty()
        || text.eq_ignore_ascii_case("NULL")
        || text
            .bytes()
            .any(|b| b == b'"' || b == b'\\' || b == b'{' || b == b'}' || b == delimiter as u8 || is_array_space(b));
    if !needs_quotes {
        return text.to_string();
    }
    let mut quoted = String::with_capacity(text.len() + 2);
    quoted.push('"');
    for c in text.chars() {
        if c == '"' || c == '\\' {
            quoted.push('\\');
        }
        quoted.push(c);
    }
    quoted.push('"');
    quoted
}

/// record_text renders a binary record the way record_out does.
pub(crate) fn record_text(r: &mut Reader<'_>) -> Result<String, String> {
    let count = r.i32()?;
    if count < 0 {
        return Err(format!("invalid record field count {count}"));
    }
    let mut text = String::from("(");
    for i in 0..count {
        if i > 0 {
            text.push(',');
        }
        let field_oid = r.u32()?;
        if let Some(bytes) = r.length_prefixed()? {
            let value = decode_nested(field_oid, bytes)?;
            let needs_quotes = value.is_empty()
                || value.bytes().any(|b| matches!(b, b'"' | b'\\' | b'(' | b')' | b',') || is_array_space(b));
            if needs_quotes {
                text.push('"');
                for c in value.chars() {
                    if c == '"' || c == '\\' {
                        text.push(c);
                    }
                    text.push(c);
                }
                text.push('"');
            } else {
                text.push_str(&value);
            }
        }
    }
    text.push(')');
    Ok(text)
}

/// tsvector_text renders a binary tsvector the way tsvectorout does.
pub(crate) fn tsvector_text(r: &mut Reader<'_>) -> Result<String, String> {
    let count = r.i32()?;
    if count < 0 {
        return Err(format!("invalid lexeme count {count}"));
    }
    let mut lexemes = Vec::with_capacity(count as usize);
    for _ in 0..count {
        let mut bytes = Vec::new();
        loop {
            let byte = r.u8()?;
            if byte == 0 {
                break;
            }
            bytes.push(byte);
        }
        let lexeme = String::from_utf8(bytes).map_err(|err| format!("invalid lexeme: {err}"))?;
        let mut text = String::from("'");
        for c in lexeme.chars() {
            if c == '\'' || c == '\\' {
                text.push(c);
            }
            text.push(c);
        }
        text.push('\'');
        let position_count = r.u16()?;
        for p in 0..position_count {
            let position = r.u16()?;
            text.push(if p == 0 { ':' } else { ',' });
            text.push_str(&(position & 0x3fff).to_string());
            match position >> 14 {
                3 => text.push('A'),
                2 => text.push('B'),
                1 => text.push('C'),
                _ => {}
            }
        }
        lexemes.push(text);
    }
    Ok(lexemes.join(" "))
}

/// range_element_oid returns the element type of a built-in range type.
fn range_element_oid(oid: u32) -> Result<u32, String> {
    Ok(match oid {
        3904 => 23,
        3906 => 1700,
        3908 => 1114,
        3910 => 1184,
        3912 => 1082,
        3926 => 20,
        _ => return Err(format!("type {oid} is not a built-in range")),
    })
}

/// range_text renders a binary range the way range_out does.
pub(crate) fn range_text(oid: u32, r: &mut Reader<'_>) -> Result<String, String> {
    let element_oid = range_element_oid(oid)?;
    let flags = r.u8()?;
    if flags & 0x01 != 0 {
        return Ok("empty".to_string());
    }
    let mut text = String::new();
    text.push(if flags & 0x02 != 0 { '[' } else { '(' });
    if flags & 0x08 == 0 {
        let bytes = r.length_prefixed()?.ok_or("range lower bound is NULL")?;
        text.push_str(&quote_range_bound(&decode_nested(element_oid, bytes)?));
    }
    text.push(',');
    if flags & 0x10 == 0 {
        let bytes = r.length_prefixed()?.ok_or("range upper bound is NULL")?;
        text.push_str(&quote_range_bound(&decode_nested(element_oid, bytes)?));
    }
    text.push(if flags & 0x04 != 0 { ']' } else { ')' });
    Ok(text)
}

/// quote_range_bound quotes a range bound when range_out would.
fn quote_range_bound(text: &str) -> String {
    let needs_quotes = text.is_empty()
        || text.bytes().any(|b| matches!(b, b'"' | b'\\' | b'(' | b')' | b'[' | b']' | b',') || is_array_space(b));
    if !needs_quotes {
        return text.to_string();
    }
    let mut quoted = String::from("\"");
    for c in text.chars() {
        if c == '"' || c == '\\' {
            quoted.push(c);
        }
        quoted.push(c);
    }
    quoted.push('"');
    quoted
}

/// multirange_text renders a binary multirange the way multirange_out does.
pub(crate) fn multirange_text(oid: u32, r: &mut Reader<'_>) -> Result<String, String> {
    let range_oid = match oid {
        4451 => 3904,
        4532 => 3906,
        4533 => 3908,
        4534 => 3910,
        4535 => 3912,
        4536 => 3926,
        _ => return Err(format!("type {oid} is not a built-in multirange")),
    };
    let count = r.i32()?;
    if count < 0 {
        return Err(format!("invalid range count {count}"));
    }
    let mut ranges = Vec::with_capacity(count as usize);
    for _ in 0..count {
        let bytes = r.length_prefixed()?.ok_or("multirange contains a NULL range")?;
        let mut nested = Reader::new(bytes);
        ranges.push(range_text(range_oid, &mut nested)?);
        nested.expect_end(range_oid)?;
    }
    Ok(format!("{{{}}}", ranges.join(",")))
}
