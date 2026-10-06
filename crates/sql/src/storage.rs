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

//! Values as tuple fields, in the encodings Doltgres stores each type with.

use std::cmp::Ordering;

use doltdb::database::Database;
use prolly::val::{compare_field, encoding};
use store::Hash;

use num_traits::Zero;

use crate::catalog::ColumnType;
use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;
use crate::types::Value;

/// encode_field returns a value of the type as a tuple field of the encoding, or None for NULL.
pub fn encode_field(value: &Value, field_encoding: u8, ty: ColumnType) -> Result<Option<Vec<u8>>> {
    Ok(Some(match (value, field_encoding) {
        (Value::Null, _) => return Ok(None),
        (Value::Int2(i), encoding::INT16) => i.to_le_bytes().to_vec(),
        (Value::Int4(i), encoding::INT32) => i.to_le_bytes().to_vec(),
        (Value::Int8(i), encoding::INT64) => i.to_le_bytes().to_vec(),
        (Value::Float4(f), encoding::FLOAT32) => f.to_le_bytes().to_vec(),
        (Value::Float8(f), encoding::FLOAT64) => f.to_le_bytes().to_vec(),
        (Value::Numeric(n), encoding::DECIMAL) => n.encode(),
        (Value::Text(s), encoding::STRING) => {
            let mut field = s.clone().into_bytes();
            field.push(0);
            field
        }
        (Value::Text(s) | Value::Json(s), encoding::STRING_ADAPTIVE | encoding::JSON_ADAPTIVE) => inline(s.as_bytes()),
        (Value::Jsonb(json), encoding::JSON_ADAPTIVE) => inline(json.compact().as_bytes()),
        (Value::Bytea(bytes), encoding::BYTES_ADAPTIVE) => inline(bytes),
        (value, encoding::EXTENDED) => serialize_value(value, ty)?,
        (value, encoding::EXTENDED_ADAPTIVE) => inline(&serialize_value(value, ty)?),
        (value, field_encoding) => {
            return Err(PgError::unsupported(format!("storing {value:?} with encoding {field_encoding}")));
        }
    }))
}

/// decode_field reads a value of the type from a tuple field of the encoding, where None is NULL.
pub fn decode_field(db: &Database, field: Option<&[u8]>, field_encoding: u8, ty: ColumnType) -> Result<Value> {
    let Some(field) = field else { return Ok(Value::Null) };
    let resolved;
    let field = if is_adaptive(field_encoding) {
        resolved = adaptive_bytes(db, field)?;
        resolved.as_slice()
    } else {
        field
    };
    let corrupt = || PgError::internal(format!("a field of encoding {field_encoding} has {} bytes", field.len()));
    Ok(match field_encoding {
        encoding::INT16 => Value::Int2(i16::from_le_bytes(field.try_into().map_err(|_| corrupt())?)),
        encoding::INT32 => Value::Int4(i32::from_le_bytes(field.try_into().map_err(|_| corrupt())?)),
        encoding::INT64 => Value::Int8(i64::from_le_bytes(field.try_into().map_err(|_| corrupt())?)),
        encoding::FLOAT32 => Value::Float4(f32::from_le_bytes(field.try_into().map_err(|_| corrupt())?)),
        encoding::FLOAT64 => Value::Float8(f64::from_le_bytes(field.try_into().map_err(|_| corrupt())?)),
        encoding::DECIMAL => Value::Numeric(Numeric::decode(field).ok_or_else(corrupt)?),
        encoding::STRING => {
            let bytes = field.strip_suffix(&[0]).ok_or_else(corrupt)?;
            Value::Text(String::from_utf8(bytes.to_vec()).map_err(|_| corrupt())?)
        }
        encoding::JSON_ADAPTIVE if ty.oid == crate::oid::JSONB => {
            let text = std::str::from_utf8(field).map_err(|_| corrupt())?;
            Value::Jsonb(Box::new(crate::json::parse(text, true)?))
        }
        encoding::JSON_ADAPTIVE if ty.oid == crate::oid::JSON => {
            Value::Json(String::from_utf8(field.to_vec()).map_err(|_| corrupt())?)
        }
        encoding::STRING_ADAPTIVE | encoding::JSON_ADAPTIVE => {
            Value::Text(String::from_utf8(field.to_vec()).map_err(|_| corrupt())?)
        }
        encoding::BYTES_ADAPTIVE => Value::Bytea(field.to_vec()),
        encoding::EXTENDED | encoding::EXTENDED_ADAPTIVE => deserialize_value(field, ty)?,
        _ => return Err(PgError::unsupported(format!("reading fields of encoding {field_encoding}"))),
    })
}

/// serialize_value writes a value of a type as Doltgres' type serialization does, which extended columns and array
/// elements store.
pub fn serialize_value(value: &Value, ty: ColumnType) -> Result<Vec<u8>> {
    use crate::datetime::{self as dt, USECS_PER_DAY};
    Ok(match value {
        Value::Bool(b) => vec![*b as u8],
        Value::Int2(i) => ((*i as u16) ^ (1 << 15)).to_be_bytes().to_vec(),
        Value::Int4(i) => offset_i32(*i).to_vec(),
        Value::Int8(i) => offset_i64(*i).to_vec(),
        Value::Float4(f) => {
            let bits = f.to_bits();
            (if *f >= 0.0 { bits ^ (1 << 31) } else { !bits }).to_be_bytes().to_vec()
        }
        Value::Float8(f) => {
            let bits = f.to_bits();
            (if *f >= 0.0 { bits ^ (1 << 63) } else { !bits }).to_be_bytes().to_vec()
        }
        Value::Numeric(n) => numeric_gob(n)?,
        Value::Text(s) | Value::Bit(s) => {
            let mut out = Vec::with_capacity(s.len() + 2);
            write_uvarint(&mut out, s.len() as u64);
            out.extend_from_slice(s.as_bytes());
            out
        }
        Value::Bytea(bytes) => {
            let mut out = Vec::with_capacity(bytes.len() + 2);
            write_uvarint(&mut out, bytes.len() as u64);
            out.extend_from_slice(bytes);
            out
        }
        Value::Uuid(uuid) => uuid.to_vec(),
        Value::Base(base) => base.data.clone(),
        Value::Date(d) => {
            let ts = match *d {
                dt::DATE_NOBEGIN => dt::TIMESTAMP_NOBEGIN,
                dt::DATE_NOEND => dt::TIMESTAMP_NOEND,
                d => d as i64 * USECS_PER_DAY,
            };
            let (seconds, nanos) = dt::timestamp_to_go(ts);
            dt::go_time::marshal(seconds, nanos)
        }
        Value::Timestamp(ts) | Value::TimestampTz(ts) => {
            let (seconds, nanos) = dt::timestamp_to_go(*ts);
            dt::go_time::marshal(seconds, nanos)
        }
        Value::Time(t) => offset_i64(*t).to_vec(),
        Value::TimeTz(t, z) => [offset_i64(*t).as_slice(), &offset_i32(*z)].concat(),
        Value::Interval(iv) => {
            let sort_nanos = (iv.months as i64 * 30 * USECS_PER_DAY + iv.days as i64 * USECS_PER_DAY + iv.micros)
                .saturating_mul(1000);
            [offset_i64(sort_nanos).as_slice(), &offset_i32(iv.months), &offset_i32(iv.days)].concat()
        }
        Value::Array(a) => {
            let element = ColumnType { oid: a.element, modifier: ty.modifier };
            crate::array::serialize(a, &|v| serialize_value(v, element))?
        }
        Value::Json(text) => text.as_bytes().to_vec(),
        Value::Jsonb(json) => {
            let mut out = Vec::new();
            serialize_json(json, &mut out)?;
            out
        }
        Value::Enum(e) => serialize_value(&Value::Text(e.label.clone()), ty)?,
        Value::Composite(c) => {
            let attributes = match crate::usertypes::get(c.type_oid).map(|t| t.kind.clone()) {
                Some(crate::usertypes::Kind::Composite(attributes)) => attributes,
                _ => return Err(PgError::internal(format!("an unknown composite type {}", c.type_oid))),
            };
            let mut out = vec![0];
            write_uvarint(&mut out, c.fields.len() as u64);
            for (field, (_, field_type)) in c.fields.iter().zip(&attributes) {
                let type_id = crate::usertypes::type_id(field_type.oid);
                write_uvarint(&mut out, type_id.len() as u64);
                out.extend_from_slice(&type_id);
                let bytes = if field.is_null() { Vec::new() } else { serialize_value(field, *field_type)? };
                write_uvarint(&mut out, bytes.len() as u64);
                out.extend_from_slice(&bytes);
            }
            out
        }
        other => return Err(PgError::unsupported(format!("storing {other:?}"))),
    })
}

/// deserialize_value reads a value of a type from Doltgres' type serialization.
pub fn deserialize_value(field: &[u8], ty: ColumnType) -> Result<Value> {
    use crate::datetime::{self as dt, USECS_PER_DAY};
    use crate::oid;
    let corrupt = || PgError::internal(format!("a stored value of type {} has {} bytes", ty.oid, field.len()));
    if crate::array::is_array_type(ty.oid) {
        let element = crate::expr::element_type(ty.oid);
        let element_type = ColumnType { oid: element, modifier: ty.modifier };
        let array = crate::array::deserialize(field, element, &|bytes| deserialize_value(bytes, element_type))?;
        return Ok(Value::Array(Box::new(array)));
    }
    if crate::catalog::builtin_type(ty.oid).is_none()
        && let Some(user_type) = crate::usertypes::get(ty.oid)
    {
        return deserialize_user_value(field, &user_type);
    }
    Ok(match ty.oid {
        oid::JSON => Value::Json(String::from_utf8(field.to_vec()).map_err(|_| corrupt())?),
        oid::JSONB => {
            let mut position = 0;
            Value::Jsonb(Box::new(deserialize_json(field, &mut position).ok_or_else(corrupt)?))
        }
        oid::BOOL => Value::Bool(field.first().is_some_and(|&b| b != 0)),
        oid::INT2 => Value::Int2((u16::from_be_bytes(field.try_into().map_err(|_| corrupt())?) ^ (1 << 15)) as i16),
        oid::INT4 => Value::Int4(read_offset_i32(field).ok_or_else(corrupt)?),
        oid::INT8 => Value::Int8(read_offset_i64(field).ok_or_else(corrupt)?),
        oid::FLOAT4 => {
            let bits = u32::from_be_bytes(field.try_into().map_err(|_| corrupt())?);
            Value::Float4(f32::from_bits(if bits & (1 << 31) != 0 { bits ^ (1 << 31) } else { !bits }))
        }
        oid::FLOAT8 => {
            let bits = u64::from_be_bytes(field.try_into().map_err(|_| corrupt())?);
            Value::Float8(f64::from_bits(if bits & (1 << 63) != 0 { bits ^ (1 << 63) } else { !bits }))
        }
        oid::NUMERIC => Value::Numeric(numeric_from_gob(field).ok_or_else(corrupt)?),
        oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME | oid::BIT | oid::VARBIT | oid::BYTEA => {
            let mut i = 0;
            let length = read_uvarint(field, &mut i).ok_or_else(corrupt)?;
            let bytes = field.get(i..i + length as usize).ok_or_else(corrupt)?;
            match ty.oid {
                oid::BYTEA => Value::Bytea(bytes.to_vec()),
                oid::BIT | oid::VARBIT => Value::Bit(String::from_utf8(bytes.to_vec()).map_err(|_| corrupt())?),
                _ => Value::Text(String::from_utf8(bytes.to_vec()).map_err(|_| corrupt())?),
            }
        }
        oid::UUID => Value::Uuid(field.try_into().map_err(|_| corrupt())?),
        oid::DATE | oid::TIMESTAMP | oid::TIMESTAMPTZ => {
            let (seconds, nanos) = dt::go_time::unmarshal(field).ok_or_else(corrupt)?;
            let ts = dt::timestamp_from_go(seconds, nanos);
            match ty.oid {
                oid::DATE => Value::Date(match ts {
                    dt::TIMESTAMP_NOBEGIN => dt::DATE_NOBEGIN,
                    dt::TIMESTAMP_NOEND => dt::DATE_NOEND,
                    ts => ts.div_euclid(USECS_PER_DAY) as i32,
                }),
                oid::TIMESTAMP => Value::Timestamp(go_local(field, ts)),
                _ => Value::TimestampTz(ts),
            }
        }
        oid::TIME => Value::Time(read_offset_i64(field).ok_or_else(corrupt)?),
        oid::TIMETZ if field.len() == 12 => Value::TimeTz(
            read_offset_i64(&field[..8]).ok_or_else(corrupt)?,
            read_offset_i32(&field[8..]).ok_or_else(corrupt)?,
        ),
        oid::INTERVAL if field.len() == 16 => {
            let sort_nanos = read_offset_i64(&field[..8]).ok_or_else(corrupt)?;
            let months = read_offset_i32(&field[8..12]).ok_or_else(corrupt)?;
            let days = read_offset_i32(&field[12..]).ok_or_else(corrupt)?;
            let nanos = sort_nanos - (months as i64 * 30 + days as i64) * USECS_PER_DAY * 1000;
            Value::Interval(dt::Interval { months, days, micros: nanos / 1000 })
        }
        other => return Err(PgError::unsupported(format!("reading stored values of type {other}"))),
    })
}

/// deserialize_user_value reads a value of a user-defined type from Doltgres' type serialization.
fn deserialize_user_value(field: &[u8], user_type: &crate::usertypes::UserType) -> Result<Value> {
    use crate::usertypes::Kind;
    let corrupt = || PgError::internal(format!("a stored value of type {} is corrupt", user_type.name));
    match &user_type.kind {
        Kind::Enum(_) => match deserialize_value(field, ColumnType { oid: crate::oid::TEXT, modifier: -1 })? {
            Value::Text(label) => Ok(Value::Enum(Box::new(crate::types::EnumValue { type_oid: user_type.oid, label }))),
            _ => Err(corrupt()),
        },
        Kind::Domain(domain) => deserialize_value(field, domain.base),
        Kind::Composite(_) => {
            if field.first() != Some(&0) {
                return Err(corrupt());
            }
            let mut position = 1;
            let count = read_uvarint(field, &mut position).ok_or_else(corrupt)?;
            let mut fields = Vec::with_capacity(count as usize);
            for _ in 0..count {
                let length = read_uvarint(field, &mut position).ok_or_else(corrupt)? as usize;
                let type_id = field.get(position..position + length).ok_or_else(corrupt)?;
                position += length;
                let length = read_uvarint(field, &mut position).ok_or_else(corrupt)? as usize;
                let bytes = field.get(position..position + length).ok_or_else(corrupt)?;
                position += length;
                let field_type = ColumnType { oid: crate::usertypes::type_oid(type_id), modifier: -1 };
                fields.push(if bytes.is_empty() { Value::Null } else { deserialize_value(bytes, field_type)? });
            }
            Ok(Value::Composite(Box::new(crate::types::CompositeValue { type_oid: user_type.oid, fields })))
        }
        Kind::Array(element) => {
            let element_type = ColumnType { oid: *element, modifier: -1 };
            let array = crate::array::deserialize(field, *element, &|bytes| deserialize_value(bytes, element_type))?;
            Ok(Value::Array(Box::new(array)))
        }
        Kind::Base(_) => {
            Ok(Value::Base(Box::new(crate::types::BaseValue { type_oid: user_type.oid, data: field.to_vec() })))
        }
    }
}

/// write_uvarint writes an unsigned varint.
fn write_uvarint(out: &mut Vec<u8>, mut value: u64) {
    while value >= 0x80 {
        out.push(value as u8 | 0x80);
        value >>= 7;
    }
    out.push(value as u8);
}

/// read_uvarint reads an unsigned varint.
fn read_uvarint(data: &[u8], position: &mut usize) -> Option<u64> {
    let (mut value, mut shift) = (0u64, 0);
    loop {
        let byte = *data.get(*position)?;
        *position += 1;
        value |= ((byte & 0x7f) as u64) << shift;
        if byte < 0x80 {
            return Some(value);
        }
        shift += 7;
    }
}

/// serialize_json writes a jsonb value as Doltgres' JsonValueSerialize does, with strings keeping the escapes of
/// backslashes, newlines, tabs, and carriage returns as Go's JSON documents do.
fn serialize_json(json: &crate::json::Json, out: &mut Vec<u8>) -> Result<()> {
    use crate::json::Json;
    let string = |out: &mut Vec<u8>, s: &str| {
        write_uvarint(out, s.len() as u64);
        out.extend_from_slice(s.as_bytes());
    };
    match json {
        Json::Object(items) => {
            out.push(0);
            write_uvarint(out, items.len() as u64);
            for (key, value) in items {
                string(out, key);
                serialize_json(value, out)?;
            }
        }
        Json::Array(values) => {
            out.push(1);
            write_uvarint(out, values.len() as u64);
            for value in values {
                serialize_json(value, out)?;
            }
        }
        Json::String(s) => {
            out.push(2);
            let escaped = s.replace('\\', "\\\\").replace('\n', "\\n").replace('\t', "\\t").replace('\r', "\\r");
            string(out, &escaped);
        }
        Json::Number(n) => {
            out.push(3);
            let bytes = numeric_gob(n)?;
            write_uvarint(out, bytes.len() as u64);
            out.extend_from_slice(&bytes);
        }
        Json::Bool(b) => {
            out.push(4);
            out.push(*b as u8);
        }
        Json::Null => out.push(5),
    }
    Ok(())
}

/// unescape_go reverses the escapes that Go's JSON documents keep in strings, reading left to right.
fn unescape_go(text: &str) -> String {
    let mut out = String::with_capacity(text.len());
    let mut chars = text.chars();
    while let Some(c) = chars.next() {
        if c != '\\' {
            out.push(c);
            continue;
        }
        match chars.next() {
            Some('\\') => out.push('\\'),
            Some('n') => out.push('\n'),
            Some('t') => out.push('\t'),
            Some('r') => out.push('\r'),
            Some(other) => {
                out.push('\\');
                out.push(other);
            }
            None => out.push('\\'),
        }
    }
    out
}

/// deserialize_json reads what serialize_json writes.
fn deserialize_json(data: &[u8], position: &mut usize) -> Option<crate::json::Json> {
    use crate::json::Json;
    let string = |position: &mut usize| -> Option<String> {
        let len = read_uvarint(data, position)? as usize;
        let s = String::from_utf8(data.get(*position..*position + len)?.to_vec()).ok()?;
        *position += len;
        Some(s)
    };
    let kind = *data.get(*position)?;
    *position += 1;
    Some(match kind {
        0 => {
            let count = read_uvarint(data, position)?;
            let mut items = Vec::new();
            for _ in 0..count {
                let key = string(position)?;
                items.push((key, deserialize_json(data, position)?));
            }
            crate::json::normalize(Json::Object(items))
        }
        1 => {
            let count = read_uvarint(data, position)?;
            let mut values = Vec::new();
            for _ in 0..count {
                values.push(deserialize_json(data, position)?);
            }
            Json::Array(values)
        }
        2 => Json::String(unescape_go(&string(position)?)),
        3 => {
            let len = read_uvarint(data, position)? as usize;
            let n = numeric_from_gob(data.get(*position..*position + len)?)?;
            *position += len;
            Json::Number(n)
        }
        4 => {
            let b = *data.get(*position)? != 0;
            *position += 1;
            Json::Bool(b)
        }
        5 => Json::Null,
        _ => return None,
    })
}

/// numeric_gob writes a numeric as shopspring's Decimal.MarshalBinary does: the exponent as a big-endian 32-bit
/// integer, then the coefficient as Go's big.Int.GobEncode writes it.
fn numeric_gob(n: &Numeric) -> Result<Vec<u8>> {
    let Numeric::Finite { negative, coefficient, scale } = n else {
        return Err(PgError::new(code::FEATURE_NOT_SUPPORTED, "cannot store NaN or infinite numeric array elements"));
    };
    let mut out = (-(*scale as i32)).to_be_bytes().to_vec();
    out.push(2 | (*negative as u8));
    if !coefficient.is_zero() {
        out.extend_from_slice(&coefficient.to_bytes_be());
    }
    Ok(out)
}

/// numeric_from_gob reads what numeric_gob writes.
fn numeric_from_gob(bytes: &[u8]) -> Option<Numeric> {
    let exponent = i32::from_be_bytes(bytes.get(..4)?.try_into().ok()?);
    let flags = *bytes.get(4)?;
    let coefficient = num_bigint::BigUint::from_bytes_be(&bytes[5..]);
    let negative = flags & 1 == 1 && !coefficient.is_zero();
    Some(if exponent >= 0 {
        Numeric::Finite {
            negative,
            coefficient: coefficient * num_bigint::BigUint::from(10u32).pow(exponent as u32),
            scale: 0,
        }
    } else {
        Numeric::Finite { negative, coefficient, scale: (-exponent) as u32 }
    })
}

/// offset_i64 writes an integer as Doltgres' writer does: offset by 2^63 and big-endian, so bytes sort as numbers.
fn offset_i64(value: i64) -> [u8; 8] {
    ((value as u64) ^ (1 << 63)).to_be_bytes()
}

/// offset_i32 writes an integer offset by 2^31 and big-endian.
fn offset_i32(value: i32) -> [u8; 4] {
    ((value as u32) ^ (1 << 31)).to_be_bytes()
}

/// read_offset_i64 reads an integer that offset_i64 wrote.
fn read_offset_i64(bytes: &[u8]) -> Option<i64> {
    Some((u64::from_be_bytes(bytes.try_into().ok()?) ^ (1 << 63)) as i64)
}

/// read_offset_i32 reads an integer that offset_i32 wrote.
fn read_offset_i32(bytes: &[u8]) -> Option<i32> {
    Some((u32::from_be_bytes(bytes.try_into().ok()?) ^ (1 << 31)) as i32)
}

/// go_local returns a timestamp without a zone from a Go time that may carry a zone offset, by keeping its local
/// wall clock.
fn go_local(field: &[u8], utc: i64) -> i64 {
    if utc == crate::datetime::TIMESTAMP_NOBEGIN || utc == crate::datetime::TIMESTAMP_NOEND || field.len() < 15 {
        return utc;
    }
    let minutes = i16::from_be_bytes([field[13], field[14]]);
    if minutes == -1 { utc } else { utc + minutes as i64 * 60 * crate::datetime::USECS_PER_SEC }
}

/// is_adaptive reports whether an encoding stores values inline or out of band by size.
pub fn is_adaptive(field_encoding: u8) -> bool {
    matches!(
        field_encoding,
        encoding::STRING_ADAPTIVE
            | encoding::BYTES_ADAPTIVE
            | encoding::EXTENDED_ADAPTIVE
            | encoding::GEOM_ADAPTIVE
            | encoding::JSON_ADAPTIVE
    )
}

/// inline returns bytes as an inline adaptive value: a zero byte, then the bytes.
fn inline(bytes: &[u8]) -> Vec<u8> {
    let mut field = Vec::with_capacity(bytes.len() + 1);
    field.push(0);
    field.extend_from_slice(bytes);
    field
}

/// adaptive_bytes returns the bytes of an adaptive value, reading an out-of-band value from its blob.
pub fn adaptive_bytes(db: &Database, field: &[u8]) -> Result<Vec<u8>> {
    if field.first() == Some(&0) {
        return Ok(field[1..].to_vec());
    }
    let address = field.len().checked_sub(Hash::LEN).ok_or_else(|| PgError::internal("a short adaptive value"))?;
    Ok(prolly::read_blob(db, &serial::hash(&field[address..])?)?)
}

/// varint encodes an integer as the SQLite4 variable-length integers that Dolt's adaptive values use.
fn varint(x: u64) -> Vec<u8> {
    if x <= 240 {
        return vec![x as u8];
    }
    if x <= 2287 {
        let y = x - 240;
        return vec![(y / 256 + 241) as u8, (y % 256) as u8];
    }
    if x <= 67823 {
        let y = x - 2288;
        return vec![249, (y / 256) as u8, (y % 256) as u8];
    }
    let bytes = x.to_be_bytes();
    let length = 8 - (x.leading_zeros() / 8) as usize;
    let mut out = vec![247 + length as u8];
    out.extend_from_slice(&bytes[8 - length..]);
    out
}

/// out_of_band writes an adaptive value's bytes as a blob and returns its out-of-band form: the length, then the
/// blob's address.
fn out_of_band(db: &mut Database, bytes: &[u8]) -> Result<Vec<u8>> {
    let mut sink = |_: Hash, node: &[u8]| -> store::Result<()> {
        db.write_value(node.to_vec()).map_err(|err| store::Error::Corrupt(err.to_string()))?;
        Ok(())
    };
    let (address, _) = prolly::write_blob(bytes, &mut sink)?.ok_or_else(|| PgError::internal("an empty blob"))?;
    let mut field = varint(bytes.len() as u64);
    field.extend_from_slice(&address.0);
    Ok(field)
}

/// place_adaptive moves adaptive values out of band, largest first, until the tuple fits the target size, as Dolt's
/// TupleBuilder does. A value too large to ever fit goes out of band on its own.
pub fn place_adaptive(
    db: &mut Database,
    fields: &mut [Option<Vec<u8>>],
    encodings: &[u8],
    target: usize,
) -> Result<()> {
    for (field, &field_encoding) in fields.iter_mut().zip(encodings) {
        if let Some(bytes) = field
            && is_adaptive(field_encoding)
            && bytes.len() > target
        {
            *bytes = out_of_band(db, &bytes[1..])?;
        }
    }
    let mut total: usize = fields.iter().map(|f| f.as_ref().map_or(0, Vec::len)).sum();
    if total <= target {
        return Ok(());
    }
    let mut candidates: Vec<(usize, usize)> = fields
        .iter()
        .zip(encodings)
        .enumerate()
        .filter_map(|(i, (field, &field_encoding))| {
            let bytes = field.as_ref()?;
            if !is_adaptive(field_encoding) || bytes.first() != Some(&0) {
                return None;
            }
            let savings = bytes.len().saturating_sub(varint(bytes.len() as u64 - 1).len() + Hash::LEN);
            (savings > 0).then_some((i, savings))
        })
        .collect();
    candidates.sort_by_key(|c| std::cmp::Reverse(c.1));
    for (i, savings) in candidates {
        let bytes = fields[i].take().unwrap_or_default();
        fields[i] = Some(out_of_band(db, &bytes[1..])?);
        total -= savings;
        if total <= target {
            break;
        }
    }
    Ok(())
}

/// compare_key_field orders two key fields of an encoding and type, comparing numerics and Doltgres' extended values
/// by value and inline adaptive values by their bytes.
pub fn compare_key_field(field_encoding: u8, ty: ColumnType, left: Option<&[u8]>, right: Option<&[u8]>) -> Ordering {
    match (field_encoding, left, right) {
        (encoding::DECIMAL, Some(l), Some(r)) => match (Numeric::decode(l), Numeric::decode(r)) {
            (Some(l), Some(r)) => l.cmp_numeric(&r),
            _ => l.cmp(r),
        },
        (encoding::EXTENDED | encoding::EXTENDED_ADAPTIVE, Some(l), Some(r)) => {
            let value = |b: &[u8]| {
                let b =
                    if field_encoding == encoding::EXTENDED_ADAPTIVE && b.first() == Some(&0) { &b[1..] } else { b };
                deserialize_value(b, ty).ok()
            };
            match (value(l), value(r)) {
                (Some(l), Some(r)) => crate::expr::compare_values(&l, &r),
                _ => l.cmp(r),
            }
        }
        (e, Some(l), Some(r)) if is_adaptive(e) => {
            let strip = |b: &[u8]| if b.first() == Some(&0) { b[1..].to_vec() } else { b.to_vec() };
            strip(l).cmp(&strip(r))
        }
        (e, l, r) => compare_field(e, l, r),
    }
}
