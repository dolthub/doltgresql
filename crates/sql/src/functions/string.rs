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

//! String functions.

use super::{ANY, Function, text};
use crate::error::{PgError, Result, code};
use crate::oid::{BOOL, CHAR, INT4, TEXT};
use crate::query::Ctx;
use crate::types::Value;

/// f declares a strict, fixed-arity string function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the string functions.
pub const FUNCTIONS: &[Function] = &[
    f("length", &[TEXT], INT4, length),
    f("char_length", &[TEXT], INT4, length),
    f("character_length", &[TEXT], INT4, length),
    f("octet_length", &[TEXT], INT4, octet_length),
    f("bit_length", &[TEXT], INT4, bit_length),
    f("lower", &[TEXT], TEXT, lower),
    f("upper", &[TEXT], TEXT, upper),
    f("initcap", &[TEXT], TEXT, initcap),
    f("repeat", &[TEXT, INT4], TEXT, repeat),
    f("left", &[TEXT, INT4], TEXT, left),
    f("right", &[TEXT, INT4], TEXT, right),
    f("substr", &[TEXT, INT4], TEXT, substr_from),
    f("substr", &[TEXT, INT4, INT4], TEXT, substr),
    f("substring", &[TEXT, INT4], TEXT, substr_from),
    f("substring", &[TEXT, INT4, INT4], TEXT, substr),
    f("strpos", &[TEXT, TEXT], INT4, strpos),
    f("position", &[TEXT, TEXT], INT4, position),
    f("replace", &[TEXT, TEXT, TEXT], TEXT, replace),
    f("reverse", &[TEXT], TEXT, reverse),
    f("btrim", &[TEXT], TEXT, btrim_spaces),
    f("btrim", &[TEXT, TEXT], TEXT, btrim),
    f("ltrim", &[TEXT], TEXT, ltrim_spaces),
    f("ltrim", &[TEXT, TEXT], TEXT, ltrim),
    f("rtrim", &[TEXT], TEXT, rtrim_spaces),
    f("rtrim", &[TEXT, TEXT], TEXT, rtrim),
    f("lpad", &[TEXT, INT4], TEXT, lpad_spaces),
    f("lpad", &[TEXT, INT4, TEXT], TEXT, lpad),
    f("rpad", &[TEXT, INT4], TEXT, rpad_spaces),
    f("rpad", &[TEXT, INT4, TEXT], TEXT, rpad),
    f("split_part", &[TEXT, TEXT, INT4], TEXT, split_part),
    f("starts_with", &[TEXT, TEXT], BOOL, starts_with),
    f("ascii", &[TEXT], INT4, ascii),
    f("chr", &[INT4], TEXT, chr),
    f("int4", &[CHAR], INT4, char_to_int4),
    f("char", &[INT4], CHAR, int4_to_char),
    f("md5", &[TEXT], TEXT, md5),
    f("quote_ident", &[TEXT], TEXT, quote_ident),
    f("quote_literal", &[TEXT], TEXT, quote_literal),
    Function { name: "concat", args: &[ANY], ret: TEXT, strict: false, variadic: true, implementation: concat },
    Function {
        name: "concat_ws",
        args: &[TEXT, ANY],
        ret: TEXT,
        strict: false,
        variadic: true,
        implementation: concat_ws,
    },
];

/// length returns the number of characters.
fn length(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(text(&args[0]).chars().count() as i32))
}

/// octet_length returns the number of bytes.
fn octet_length(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(text(&args[0]).len() as i32))
}

/// bit_length returns the number of bits.
fn bit_length(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(text(&args[0]).len() as i32 * 8))
}

/// lower lowercases the string.
fn lower(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(text(&args[0]).to_lowercase()))
}

/// upper uppercases the string.
fn upper(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(text(&args[0]).to_uppercase()))
}

/// initcap uppercases the first letter of each word and lowercases the rest.
fn initcap(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let mut out = String::new();
    let mut start = true;
    for c in text(&args[0]).chars() {
        if c.is_alphanumeric() {
            if start {
                out.extend(c.to_uppercase());
            } else {
                out.extend(c.to_lowercase());
            }
            start = false;
        } else {
            out.push(c);
            start = true;
        }
    }
    Ok(Value::Text(out))
}

/// int returns an int4 argument.
fn int(value: &Value) -> i32 {
    match value {
        Value::Int4(i) => *i,
        _ => 0,
    }
}

/// repeat repeats the string.
fn repeat(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let count = int(&args[1]).max(0) as usize;
    let s = text(&args[0]);
    if s.len().saturating_mul(count) > (1 << 30) - 1 {
        return Err(PgError::new(code::PROGRAM_LIMIT_EXCEEDED, "requested length too large"));
    }
    Ok(Value::Text(s.repeat(count)))
}

/// left returns the first n characters, or all but the last -n.
fn left(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let chars: Vec<char> = text(&args[0]).chars().collect();
    let n = int(&args[1]);
    let take = if n >= 0 { (n as usize).min(chars.len()) } else { chars.len().saturating_sub((-n) as usize) };
    Ok(Value::Text(chars[..take].iter().collect()))
}

/// right returns the last n characters, or all but the first -n.
fn right(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let chars: Vec<char> = text(&args[0]).chars().collect();
    let n = int(&args[1]);
    let skip = if n >= 0 { chars.len().saturating_sub(n as usize) } else { ((-n) as usize).min(chars.len()) };
    Ok(Value::Text(chars[skip..].iter().collect()))
}

/// substring_chars returns the characters from the 1-based start for the count, as Postgres' substring does for
/// starts before the string and unbounded counts.
fn substring_chars(s: &str, start: i64, count: Option<i64>) -> Result<String> {
    if let Some(count) = count
        && count < 0
    {
        return Err(PgError::new(code::SUBSTRING_ERROR, "negative substring length not allowed"));
    }
    let end = count.map(|c| start.saturating_add(c));
    let from = start.max(1) as usize - 1;
    Ok(match end {
        Some(end) if end <= 1 => String::new(),
        Some(end) => s.chars().skip(from).take((end as usize - 1).saturating_sub(from)).collect(),
        None => s.chars().skip(from).collect(),
    })
}

/// substr_from returns the characters from the 1-based start.
fn substr_from(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(substring_chars(text(&args[0]), int(&args[1]) as i64, None)?))
}

/// substr returns the characters from the 1-based start for the count.
fn substr(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(substring_chars(text(&args[0]), int(&args[1]) as i64, Some(int(&args[2]) as i64))?))
}

/// char_position returns the 1-based character position of a substring, or 0.
fn char_position(haystack: &str, needle: &str) -> i32 {
    haystack.find(needle).map_or(0, |i| haystack[..i].chars().count() as i32 + 1)
}

/// strpos returns where the second string starts in the first.
fn strpos(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(char_position(text(&args[0]), text(&args[1]))))
}

/// position returns where the second string starts in the first, as POSITION(a IN b) calls it.
fn position(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(char_position(text(&args[0]), text(&args[1]))))
}

/// replace replaces every occurrence of a substring.
fn replace(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (s, from, to) = (text(&args[0]), text(&args[1]), text(&args[2]));
    Ok(Value::Text(if from.is_empty() { s.to_string() } else { s.replace(from, to) }))
}

/// reverse reverses the characters.
fn reverse(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(text(&args[0]).chars().rev().collect()))
}

/// trim removes characters of the set from the chosen ends.
fn trim(s: &str, set: &str, start: bool, end: bool) -> String {
    let in_set = |c: char| set.contains(c);
    let s = if start { s.trim_start_matches(in_set) } else { s };
    let s = if end { s.trim_end_matches(in_set) } else { s };
    s.to_string()
}

/// btrim_spaces removes spaces from both ends.
fn btrim_spaces(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(trim(text(&args[0]), " ", true, true)))
}

/// btrim removes the characters from both ends.
fn btrim(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(trim(text(&args[0]), text(&args[1]), true, true)))
}

/// ltrim_spaces removes spaces from the start.
fn ltrim_spaces(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(trim(text(&args[0]), " ", true, false)))
}

/// ltrim removes the characters from the start.
fn ltrim(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(trim(text(&args[0]), text(&args[1]), true, false)))
}

/// rtrim_spaces removes spaces from the end.
fn rtrim_spaces(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(trim(text(&args[0]), " ", false, true)))
}

/// rtrim removes the characters from the end.
fn rtrim(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(trim(text(&args[0]), text(&args[1]), false, true)))
}

/// pad pads or truncates the string to the length, filling from the left or right.
fn pad(s: &str, length: i32, fill: &str, left: bool) -> String {
    let length = length.max(0) as usize;
    let chars: Vec<char> = s.chars().collect();
    if chars.len() >= length || fill.is_empty() {
        return chars[..length.min(chars.len())].iter().collect();
    }
    let filler: String = fill.chars().cycle().take(length - chars.len()).collect();
    if left { format!("{filler}{s}") } else { format!("{s}{filler}") }
}

/// lpad_spaces pads on the left with spaces.
fn lpad_spaces(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(pad(text(&args[0]), int(&args[1]), " ", true)))
}

/// lpad pads on the left with the fill.
fn lpad(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(pad(text(&args[0]), int(&args[1]), text(&args[2]), true)))
}

/// rpad_spaces pads on the right with spaces.
fn rpad_spaces(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(pad(text(&args[0]), int(&args[1]), " ", false)))
}

/// rpad pads on the right with the fill.
fn rpad(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(pad(text(&args[0]), int(&args[1]), text(&args[2]), false)))
}

/// split_part returns the nth field split by the delimiter, counting from the end for a negative n.
fn split_part(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (s, delimiter, n) = (text(&args[0]), text(&args[1]), int(&args[2]));
    if n == 0 {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "field position must not be zero"));
    }
    let fields: Vec<&str> = if delimiter.is_empty() { vec![s] } else { s.split(delimiter).collect() };
    let index = if n > 0 { n as usize - 1 } else { fields.len().wrapping_sub((-n) as usize) };
    Ok(Value::Text(fields.get(index).copied().unwrap_or("").to_string()))
}

/// starts_with reports whether the string starts with the prefix.
fn starts_with(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(text(&args[0]).starts_with(text(&args[1]))))
}

/// ascii returns the code point of the first character.
fn ascii(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(text(&args[0]).chars().next().map_or(0, |c| c as i32)))
}

/// char_to_int4 returns the signed byte of a "char" value.
fn char_to_int4(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let text = text(&args[0]);
    let byte = match text.strip_prefix('\\') {
        Some(octal) if octal.len() == 3 => u8::from_str_radix(octal, 8).unwrap_or(0),
        _ => text.bytes().next().unwrap_or(0),
    };
    Ok(Value::Int4(byte as i8 as i32))
}

/// int4_to_char returns the "char" value of a signed byte.
fn int4_to_char(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let value = int(&args[0]);
    let byte =
        i8::try_from(value).map_err(|_| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "\"char\" out of range"))? as u8;
    Ok(Value::Text(match byte {
        0 => String::new(),
        b if b.is_ascii() => (b as char).to_string(),
        b => format!("\\{b:03o}"),
    }))
}

/// chr returns the character of the code point.
fn chr(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let point = int(&args[0]);
    if point == 0 {
        return Err(PgError::new(code::PROGRAM_LIMIT_EXCEEDED, "null character not permitted"));
    }
    char::from_u32(point as u32)
        .map(|c| Value::Text(c.to_string()))
        .ok_or_else(|| PgError::new(code::PROGRAM_LIMIT_EXCEEDED, "requested character too large for encoding"))
}

/// md5 returns the MD5 hash in hex.
fn md5(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    use md5::{Digest, Md5};
    let hash = Md5::digest(text(&args[0]).as_bytes());
    Ok(Value::Text(hash.iter().map(|b| format!("{b:02x}")).collect()))
}

/// quote_ident quotes an identifier when it needs quotes.
fn quote_ident(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(crate::engine::quote_identifier(text(&args[0]))))
}

/// quote_literal quotes a string literal, with the E prefix when it has backslashes.
fn quote_literal(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let s = text(&args[0]);
    let quoted = s.replace('\'', "''");
    Ok(Value::Text(if s.contains('\\') {
        format!("E'{}'", quoted.replace('\\', "\\\\"))
    } else {
        format!("'{quoted}'")
    }))
}

/// concat joins the text of every non-NULL argument.
fn concat(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(args.iter().filter_map(Value::output).collect()))
}

/// concat_ws joins the text of every non-NULL argument after the first with the first.
fn concat_ws(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Some(separator) = args[0].output() else { return Ok(Value::Null) };
    let parts: Vec<String> = args[1..].iter().filter_map(Value::output).collect();
    Ok(Value::Text(parts.join(&separator)))
}
