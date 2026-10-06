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

use crate::catalog::ColumnType;
use crate::error::{PgError, Result};
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
        (Value::Text(s), encoding::STRING_ADAPTIVE) => inline(s.as_bytes()),
        (Value::Bool(b), encoding::EXTENDED) if ty.oid == crate::oid::BOOL => vec![*b as u8],
        (Value::Date(d), encoding::EXTENDED) => {
            let ts = match *d {
                crate::datetime::DATE_NOBEGIN => crate::datetime::TIMESTAMP_NOBEGIN,
                crate::datetime::DATE_NOEND => crate::datetime::TIMESTAMP_NOEND,
                d => d as i64 * crate::datetime::USECS_PER_DAY,
            };
            let (seconds, nanos) = crate::datetime::timestamp_to_go(ts);
            crate::datetime::go_time::marshal(seconds, nanos)
        }
        (Value::Timestamp(ts) | Value::TimestampTz(ts), encoding::EXTENDED) => {
            let (seconds, nanos) = crate::datetime::timestamp_to_go(*ts);
            crate::datetime::go_time::marshal(seconds, nanos)
        }
        (Value::Time(t), encoding::EXTENDED) => offset_i64(*t).to_vec(),
        (Value::TimeTz(t, z), encoding::EXTENDED) => [offset_i64(*t).as_slice(), &offset_i32(*z)].concat(),
        (Value::Interval(iv), encoding::EXTENDED) => {
            let sort_nanos = (iv.months as i64 * 30 * crate::datetime::USECS_PER_DAY
                + iv.days as i64 * crate::datetime::USECS_PER_DAY
                + iv.micros)
                .saturating_mul(1000);
            [offset_i64(sort_nanos).as_slice(), &offset_i32(iv.months), &offset_i32(iv.days)].concat()
        }
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
        encoding::STRING_ADAPTIVE => Value::Text(String::from_utf8(field.to_vec()).map_err(|_| corrupt())?),
        encoding::EXTENDED if ty.oid == crate::oid::BOOL => Value::Bool(field.first().is_some_and(|&b| b != 0)),
        encoding::EXTENDED if matches!(ty.oid, crate::oid::DATE | crate::oid::TIMESTAMP | crate::oid::TIMESTAMPTZ) => {
            let (seconds, nanos) = crate::datetime::go_time::unmarshal(field).ok_or_else(corrupt)?;
            let ts = crate::datetime::timestamp_from_go(seconds, nanos);
            match ty.oid {
                crate::oid::DATE => Value::Date(match ts {
                    crate::datetime::TIMESTAMP_NOBEGIN => crate::datetime::DATE_NOBEGIN,
                    crate::datetime::TIMESTAMP_NOEND => crate::datetime::DATE_NOEND,
                    ts => ts.div_euclid(crate::datetime::USECS_PER_DAY) as i32,
                }),
                crate::oid::TIMESTAMP => Value::Timestamp(go_local(field, ts)),
                _ => Value::TimestampTz(ts),
            }
        }
        encoding::EXTENDED if ty.oid == crate::oid::TIME => Value::Time(read_offset_i64(field).ok_or_else(corrupt)?),
        encoding::EXTENDED if ty.oid == crate::oid::TIMETZ && field.len() == 12 => Value::TimeTz(
            read_offset_i64(&field[..8]).ok_or_else(corrupt)?,
            read_offset_i32(&field[8..]).ok_or_else(corrupt)?,
        ),
        encoding::EXTENDED if ty.oid == crate::oid::INTERVAL && field.len() == 16 => {
            let sort_nanos = read_offset_i64(&field[..8]).ok_or_else(corrupt)?;
            let months = read_offset_i32(&field[8..12]).ok_or_else(corrupt)?;
            let days = read_offset_i32(&field[12..]).ok_or_else(corrupt)?;
            let nanos = sort_nanos - (months as i64 * 30 + days as i64) * crate::datetime::USECS_PER_DAY * 1000;
            Value::Interval(crate::datetime::Interval { months, days, micros: nanos / 1000 })
        }
        _ => return Err(PgError::unsupported(format!("reading fields of encoding {field_encoding}"))),
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

/// compare_key_field orders two key fields of an encoding, comparing numerics by value and inline adaptive values by
/// their bytes.
pub fn compare_key_field(field_encoding: u8, left: Option<&[u8]>, right: Option<&[u8]>) -> Ordering {
    match (field_encoding, left, right) {
        (encoding::DECIMAL, Some(l), Some(r)) => match (Numeric::decode(l), Numeric::decode(r)) {
            (Some(l), Some(r)) => l.cmp_numeric(&r),
            _ => l.cmp(r),
        },
        (e, Some(l), Some(r)) if is_adaptive(e) => {
            let strip = |b: &[u8]| if b.first() == Some(&0) { b[1..].to_vec() } else { b.to_vec() };
            strip(l).cmp(&strip(r))
        }
        (e, l, r) => compare_field(e, l, r),
    }
}
