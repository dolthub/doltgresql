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

//! Functions and operators of the bytea, bit, bit varying, and uuid types.

use std::ops::Range;

use super::Function;
use crate::binary::{decode_escape, decode_hex, encode_hex};
use crate::encodings::Encoding;
use crate::error::{PgError, Result, code};
use crate::oid::{BIT, BOOL, BYTEA, FLOAT4, FLOAT8, INT2, INT4, INT8, NAME, TEXT, UUID, VARBIT};
use crate::query::Ctx;
use crate::types::Value;

/// f declares a strict function of these types.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the functions and operators of the bytea, bit, bit varying, and uuid types.
pub const FUNCTIONS: &[Function] = &[
    f("||", &[BYTEA, BYTEA], BYTEA, bytea_concat),
    f("float4send", &[FLOAT4], BYTEA, send),
    f("float8send", &[FLOAT8], BYTEA, send),
    f("int2send", &[INT2], BYTEA, send),
    f("int4send", &[INT4], BYTEA, send),
    f("int8send", &[INT8], BYTEA, send),
    f("boolsend", &[BOOL], BYTEA, send),
    f("textsend", &[TEXT], BYTEA, send),
    f("byteasend", &[BYTEA], BYTEA, send),
    f("length", &[BYTEA], INT4, bytea_length),
    f("octet_length", &[BYTEA], INT4, bytea_length),
    f("bit_length", &[BYTEA], INT4, bytea_bit_length),
    f("get_byte", &[BYTEA, INT4], INT4, get_byte),
    f("set_byte", &[BYTEA, INT4, INT4], BYTEA, set_byte),
    f("get_bit", &[BYTEA, INT8], INT4, bytea_get_bit),
    f("set_bit", &[BYTEA, INT8, INT4], BYTEA, bytea_set_bit),
    f("substring", &[BYTEA, INT4], BYTEA, bytea_substring),
    f("substring", &[BYTEA, INT4, INT4], BYTEA, bytea_substring),
    f("substr", &[BYTEA, INT4], BYTEA, bytea_substring),
    f("substr", &[BYTEA, INT4, INT4], BYTEA, bytea_substring),
    f("position", &[BYTEA, BYTEA], INT4, bytea_position),
    f("btrim", &[BYTEA, BYTEA], BYTEA, bytea_btrim),
    f("byteacmp", &[BYTEA, BYTEA], INT4, byteacmp),
    f("byteaeq", &[BYTEA, BYTEA], BOOL, byteaeq),
    f("byteane", &[BYTEA, BYTEA], BOOL, byteane),
    f("bytealt", &[BYTEA, BYTEA], BOOL, bytealt),
    f("byteale", &[BYTEA, BYTEA], BOOL, byteale),
    f("byteagt", &[BYTEA, BYTEA], BOOL, byteagt),
    f("byteage", &[BYTEA, BYTEA], BOOL, byteage),
    f("~~", &[BYTEA, BYTEA], BOOL, bytea_like),
    f("!~~", &[BYTEA, BYTEA], BOOL, bytea_not_like),
    f("encode", &[BYTEA, TEXT], TEXT, encode),
    f("decode", &[TEXT, TEXT], BYTEA, decode),
    f("md5", &[BYTEA], TEXT, md5),
    f("sha224", &[BYTEA], BYTEA, sha224),
    f("sha256", &[BYTEA], BYTEA, sha256),
    f("sha384", &[BYTEA], BYTEA, sha384),
    f("sha512", &[BYTEA], BYTEA, sha512),
    f("convert_from", &[BYTEA, NAME], TEXT, convert_from),
    f("convert_to", &[TEXT, NAME], BYTEA, convert_to),
    f("convert", &[BYTEA, NAME, NAME], BYTEA, convert),
    f("pg_char_to_encoding", &[NAME], INT4, pg_char_to_encoding),
    f("pg_encoding_to_char", &[INT4], NAME, pg_encoding_to_char),
    f("||", &[VARBIT, VARBIT], VARBIT, bit_concat),
    f("&", &[BIT, BIT], BIT, bit_and),
    f("|", &[BIT, BIT], BIT, bit_or),
    f("#", &[BIT, BIT], BIT, bit_xor),
    f("~", &[BIT], BIT, bit_not),
    f("<<", &[BIT, INT4], BIT, bit_shift_left),
    f(">>", &[BIT, INT4], BIT, bit_shift_right),
    f("length", &[BIT], INT4, bit_length),
    f("bit_length", &[BIT], INT4, bit_length),
    f("octet_length", &[BIT], INT4, bit_octet_length),
    f("get_bit", &[BIT, INT4], INT4, bit_get_bit),
    f("set_bit", &[BIT, INT4, INT4], BIT, bit_set_bit),
    f("substring", &[BIT, INT4], BIT, bit_substring),
    f("substring", &[BIT, INT4, INT4], BIT, bit_substring),
    f("position", &[BIT, BIT], INT4, bit_position),
    f("gen_random_uuid", &[], UUID, gen_random_uuid),
];

/// bytes returns a bytea argument.
fn bytes(value: &Value) -> &[u8] {
    match value {
        Value::Bytea(bytes) => bytes,
        _ => &[],
    }
}

/// bits returns a bit string argument.
fn bits(value: &Value) -> &str {
    match value {
        Value::Bit(bits) | Value::Text(bits) => bits,
        _ => "",
    }
}

/// text returns a text argument.
fn text(value: &Value) -> &str {
    match value {
        Value::Text(text) => text,
        _ => "",
    }
}

/// int returns an integer argument widened.
fn int(value: &Value) -> i64 {
    match value {
        Value::Int2(i) => *i as i64,
        Value::Int4(i) => *i as i64,
        Value::Int8(i) => *i,
        _ => 0,
    }
}

/// substring_range returns the range of a 1-based substring of `length` items, as Postgres' substring does.
fn substring_range(length: usize, start: i64, count: Option<i64>) -> Result<Range<usize>> {
    if count.is_some_and(|c| c < 0) {
        return Err(PgError::new(code::SUBSTRING_ERROR, "negative substring length not allowed"));
    }
    let from = (start.max(1) - 1).min(length as i64) as usize;
    let end = match count {
        Some(count) => (start.saturating_add(count).max(1) - 1).clamp(0, length as i64) as usize,
        None => length,
    };
    Ok(from..end.max(from))
}

/// index_error returns Postgres' error for a byte or bit index outside a bytea value.
fn index_error(index: i64, last: i64) -> PgError {
    PgError::new(code::ARRAY_SUBSCRIPT_ERROR, format!("index {index} out of valid range, 0..{last}"))
}

/// bytea_concat concatenates two bytea values.
fn bytea_concat(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bytea([bytes(&args[0]), bytes(&args[1])].concat()))
}

/// bytea_length returns the number of bytes.
fn bytea_length(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(bytes(&args[0]).len() as i32))
}

/// bytea_bit_length returns the number of bits.
fn bytea_bit_length(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(bytes(&args[0]).len() as i32 * 8))
}

/// get_byte returns the byte at a 0-based index.
fn get_byte(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (data, index) = (bytes(&args[0]), int(&args[1]));
    let byte = usize::try_from(index)
        .ok()
        .and_then(|i| data.get(i))
        .ok_or_else(|| index_error(index, data.len() as i64 - 1))?;
    Ok(Value::Int4(*byte as i32))
}

/// set_byte replaces the byte at a 0-based index.
fn set_byte(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (mut data, index) = (bytes(&args[0]).to_vec(), int(&args[1]));
    let last = data.len() as i64 - 1;
    let byte = usize::try_from(index).ok().and_then(|i| data.get_mut(i)).ok_or_else(|| index_error(index, last))?;
    *byte = int(&args[2]) as u8;
    Ok(Value::Bytea(data))
}

/// bytea_get_bit returns the bit at a 0-based index, counting from the low bit of each byte.
fn bytea_get_bit(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (data, index) = (bytes(&args[0]), int(&args[1]));
    let byte = usize::try_from(index / 8)
        .ok()
        .filter(|_| index >= 0)
        .and_then(|i| data.get(i))
        .ok_or_else(|| index_error(index, data.len() as i64 * 8 - 1))?;
    Ok(Value::Int4((byte >> (index % 8)) as i32 & 1))
}

/// bytea_set_bit replaces the bit at a 0-based index, counting from the low bit of each byte.
fn bytea_set_bit(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (mut data, index, bit) = (bytes(&args[0]).to_vec(), int(&args[1]), int(&args[2]));
    let last = data.len() as i64 * 8 - 1;
    let byte = usize::try_from(index / 8)
        .ok()
        .filter(|_| index >= 0)
        .and_then(|i| data.get_mut(i))
        .ok_or_else(|| index_error(index, last))?;
    if bit != 0 && bit != 1 {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "new bit must be 0 or 1"));
    }
    *byte = (*byte & !(1 << (index % 8))) | ((bit as u8) << (index % 8));
    Ok(Value::Bytea(data))
}

/// bytea_substring returns the bytes from a 1-based start, for a count when one is given.
fn bytea_substring(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let data = bytes(&args[0]);
    let range = substring_range(data.len(), int(&args[1]), args.get(2).map(int))?;
    Ok(Value::Bytea(data[range].to_vec()))
}

/// bytea_position returns the 1-based position where the second value starts in the first, or 0.
fn bytea_position(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (haystack, needle) = (bytes(&args[0]), bytes(&args[1]));
    let found = if needle.is_empty() { Some(0) } else { haystack.windows(needle.len()).position(|w| w == needle) };
    Ok(Value::Int4(found.map_or(0, |i| i as i32 + 1)))
}

/// bytea_btrim removes the bytes of the set from both ends.
fn bytea_btrim(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (data, set) = (bytes(&args[0]), bytes(&args[1]));
    let start = data.iter().position(|b| !set.contains(b)).unwrap_or(data.len());
    let end = data.iter().rposition(|b| !set.contains(b)).map_or(start, |i| i + 1);
    Ok(Value::Bytea(data[start..end.max(start)].to_vec()))
}

/// byteacmp compares two byte strings, returning -1, 0, or 1.
fn byteacmp(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(bytes(&args[0]).cmp(bytes(&args[1])) as i32))
}

/// byteaeq reports whether two byte strings are equal.
fn byteaeq(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(bytes(&args[0]) == bytes(&args[1])))
}

/// byteane reports whether two byte strings differ.
fn byteane(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(bytes(&args[0]) != bytes(&args[1])))
}

/// bytealt reports whether a byte string sorts before another.
fn bytealt(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(bytes(&args[0]) < bytes(&args[1])))
}

/// byteale reports whether a byte string sorts before another or equals it.
fn byteale(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(bytes(&args[0]) <= bytes(&args[1])))
}

/// byteagt reports whether a byte string sorts after another.
fn byteagt(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(bytes(&args[0]) > bytes(&args[1])))
}

/// byteage reports whether a byte string sorts after another or equals it.
fn byteage(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(bytes(&args[0]) >= bytes(&args[1])))
}

/// bytea_like implements LIKE, matching byte by byte.
fn bytea_like(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(super::pattern::like_matches(bytes(&args[0]), bytes(&args[1]), false)?))
}

/// bytea_not_like implements NOT LIKE, matching byte by byte.
fn bytea_not_like(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(bytea_like(ctx, args)? == Value::Bool(false)))
}

/// unrecognized_encoding returns Postgres' error for an encoding that encode and decode do not know.
fn unrecognized_encoding(name: &str) -> PgError {
    PgError::new(code::INVALID_PARAMETER_VALUE, format!("unrecognized encoding: \"{name}\""))
}

/// encode writes bytes as text in the hex, base64, or escape format.
fn encode(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    use base64::Engine;
    let data = bytes(&args[0]);
    Ok(Value::Text(match text(&args[1]).to_ascii_lowercase().as_str() {
        "hex" => encode_hex(data),
        "base64" => {
            let encoded = base64::engine::general_purpose::STANDARD.encode(data);
            let lines: Vec<&str> =
                encoded.as_bytes().chunks(76).map(|c| std::str::from_utf8(c).unwrap_or("")).collect();
            lines.join("\n")
        }
        "escape" => {
            let mut out = String::with_capacity(data.len());
            for &b in data {
                match b {
                    0 | 0x80..=0xff => out.push_str(&format!("\\{b:03o}")),
                    b'\\' => out.push_str("\\\\"),
                    _ => out.push(b as char),
                }
            }
            out
        }
        _ => return Err(unrecognized_encoding(text(&args[1]))),
    }))
}

/// decode reads bytes from text in the hex, base64, or escape format.
fn decode(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    use base64::Engine;
    let data = text(&args[0]);
    Ok(Value::Bytea(match text(&args[1]).to_ascii_lowercase().as_str() {
        "hex" => decode_hex(data)?,
        "base64" => {
            let compact: String = data.chars().filter(|c| !c.is_ascii_whitespace()).collect();
            if let Some(c) = compact.chars().find(|c| !(c.is_ascii_alphanumeric() || matches!(c, '+' | '/' | '='))) {
                return Err(PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    format!("invalid symbol \"{c}\" found while decoding base64 sequence"),
                ));
            }
            base64::engine::general_purpose::STANDARD
                .decode(&compact)
                .map_err(|_| PgError::new(code::INVALID_PARAMETER_VALUE, "invalid base64 end sequence"))?
        }
        "escape" => decode_escape(data)
            .ok_or_else(|| PgError::new(code::INVALID_TEXT_REPRESENTATION, "invalid input syntax for type bytea"))?,
        _ => return Err(unrecognized_encoding(text(&args[1]))),
    }))
}

/// md5 returns the hexadecimal MD5 hash.
fn md5(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    use md5::{Digest, Md5};
    Ok(Value::Text(encode_hex(&Md5::digest(bytes(&args[0])))))
}

/// sha224 returns the SHA-224 hash.
fn sha224(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    use sha2::Digest;
    Ok(Value::Bytea(sha2::Sha224::digest(bytes(&args[0])).to_vec()))
}

/// sha256 returns the SHA-256 hash.
fn sha256(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    use sha2::Digest;
    Ok(Value::Bytea(sha2::Sha256::digest(bytes(&args[0])).to_vec()))
}

/// sha384 returns the SHA-384 hash.
fn sha384(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    use sha2::Digest;
    Ok(Value::Bytea(sha2::Sha384::digest(bytes(&args[0])).to_vec()))
}

/// sha512 returns the SHA-512 hash.
fn sha512(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    use sha2::Digest;
    Ok(Value::Bytea(sha2::Sha512::digest(bytes(&args[0])).to_vec()))
}

/// encoding returns the encoding that an argument names, as the source or the destination of a conversion.
fn encoding(value: &Value, role: &str) -> Result<Encoding> {
    let name = text(value);
    Encoding::lookup(name)
        .ok_or_else(|| PgError::new(code::INVALID_PARAMETER_VALUE, format!("invalid {role} encoding name \"{name}\"")))
}

/// convert_from reads text from bytes in an encoding.
fn convert_from(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(encoding(&args[1], "source")?.decode(bytes(&args[0]))?))
}

/// convert_to writes text as bytes in an encoding.
fn convert_to(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bytea(encoding(&args[1], "destination")?.encode(text(&args[0]))?))
}

/// convert converts bytes from one encoding to another.
fn convert(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (source, destination) = (encoding(&args[1], "source")?, encoding(&args[2], "destination")?);
    Ok(Value::Bytea(destination.encode(&source.decode(bytes(&args[0]))?)?))
}

/// pg_char_to_encoding returns Postgres' number of the encoding a name names, or -1 for none.
fn pg_char_to_encoding(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(Encoding::lookup(text(&args[0])).map_or(-1, Encoding::number)))
}

/// pg_encoding_to_char returns the name of the encoding of Postgres' number, or an empty name for none.
fn pg_encoding_to_char(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let number = if let Value::Int4(n) = args[0] { n } else { -1 };
    Ok(Value::Text(Encoding::from_number(number).map_or("", Encoding::name).to_string()))
}

/// bit_concat concatenates two bit strings.
fn bit_concat(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bit(format!("{}{}", bits(&args[0]), bits(&args[1]))))
}

/// bitwise combines two bit strings of the same length bit by bit.
fn bitwise(args: &[Value], name: &str, op: fn(bool, bool) -> bool) -> Result<Value> {
    let (left, right) = (bits(&args[0]), bits(&args[1]));
    if left.len() != right.len() {
        return Err(PgError::new(
            code::STRING_DATA_LENGTH_MISMATCH,
            format!("cannot {name} bit strings of different sizes"),
        ));
    }
    let combined = left.bytes().zip(right.bytes()).map(|(l, r)| if op(l == b'1', r == b'1') { '1' } else { '0' });
    Ok(Value::Bit(combined.collect()))
}

/// bit_and returns the bitwise AND.
fn bit_and(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    bitwise(args, "AND", |l, r| l && r)
}

/// bit_or returns the bitwise OR.
fn bit_or(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    bitwise(args, "OR", |l, r| l || r)
}

/// bit_xor returns the bitwise exclusive OR.
fn bit_xor(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    bitwise(args, "XOR", |l, r| l != r)
}

/// bit_not returns the bitwise complement.
fn bit_not(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bit(bits(&args[0]).chars().map(|c| if c == '1' { '0' } else { '1' }).collect()))
}

/// shift_bits shifts a bit string left by a count, or right for a negative count, keeping its length.
fn shift_bits(value: &str, count: i64) -> String {
    let length = value.len();
    let amount = count.unsigned_abs().min(length as u64) as usize;
    if count >= 0 {
        format!("{}{}", &value[amount..], "0".repeat(amount))
    } else {
        format!("{}{}", "0".repeat(amount), &value[..length - amount])
    }
}

/// bit_shift_left shifts a bit string left, filling with zeros.
fn bit_shift_left(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bit(shift_bits(bits(&args[0]), int(&args[1]))))
}

/// bit_shift_right shifts a bit string right, filling with zeros.
fn bit_shift_right(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bit(shift_bits(bits(&args[0]), -int(&args[1]))))
}

/// bit_length returns the number of bits.
fn bit_length(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(bits(&args[0]).len() as i32))
}

/// bit_octet_length returns the number of bytes the bits fill.
fn bit_octet_length(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(bits(&args[0]).len().div_ceil(8) as i32))
}

/// bit_index returns a 0-based index into a bit string, failing outside it.
fn bit_index(value: &str, index: i64) -> Result<usize> {
    usize::try_from(index).ok().filter(|&i| i < value.len()).ok_or_else(|| {
        PgError::new(
            code::ARRAY_SUBSCRIPT_ERROR,
            format!("bit index {index} out of valid range (0..{})", value.len() as i64 - 1),
        )
    })
}

/// bit_get_bit returns the bit at a 0-based index.
fn bit_get_bit(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let value = bits(&args[0]);
    let index = bit_index(value, int(&args[1]))?;
    Ok(Value::Int4((value.as_bytes()[index] == b'1') as i32))
}

/// bit_set_bit replaces the bit at a 0-based index.
fn bit_set_bit(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let value = bits(&args[0]);
    let index = bit_index(value, int(&args[1]))?;
    let digit = match int(&args[2]) {
        0 => "0",
        1 => "1",
        _ => return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "new bit must be 0 or 1")),
    };
    Ok(Value::Bit(format!("{}{digit}{}", &value[..index], &value[index + 1..])))
}

/// bit_substring returns the bits from a 1-based start, for a count when one is given.
fn bit_substring(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let value = bits(&args[0]);
    let range = substring_range(value.len(), int(&args[1]), args.get(2).map(int))?;
    Ok(Value::Bit(value[range].to_string()))
}

/// bit_position returns the 1-based position where the second bit string starts in the first, or 0.
fn bit_position(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (haystack, needle) = (bits(&args[0]), bits(&args[1]));
    Ok(Value::Int4(if needle.is_empty() { 0 } else { haystack.find(needle).map_or(0, |i| i as i32 + 1) }))
}

/// gen_random_uuid returns a random version 4 uuid.
fn gen_random_uuid(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    let mut uuid: [u8; 16] = rand::random();
    uuid[6] = (uuid[6] & 0x0f) | 0x40;
    uuid[8] = (uuid[8] & 0x3f) | 0x80;
    Ok(Value::Uuid(uuid))
}

/// send returns a value's binary format, as the type's send function does.
fn send(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bytea(match &args[0] {
        Value::Float4(f) => f.to_be_bytes().to_vec(),
        Value::Float8(f) => f.to_be_bytes().to_vec(),
        Value::Int2(i) => i.to_be_bytes().to_vec(),
        Value::Int4(i) => i.to_be_bytes().to_vec(),
        Value::Int8(i) => i.to_be_bytes().to_vec(),
        Value::Bool(b) => vec![*b as u8],
        Value::Text(text) => text.clone().into_bytes(),
        Value::Bytea(bytes) => bytes.clone(),
        other => return Err(PgError::internal(format!("no binary format for {other:?}"))),
    }))
}
