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

//! Values and their text and binary wire formats.

use crate::array::{self, Array};
use crate::datetime::{self, Interval};
use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;
use crate::oid;

/// TEXT_FORMAT and BINARY_FORMAT are the wire format codes.
pub const TEXT_FORMAT: i16 = 0;
pub const BINARY_FORMAT: i16 = 1;

/// Value is a SQL value.
#[derive(Clone, Debug)]
pub enum Value {
    Null,
    Bool(bool),
    Int2(i16),
    Int4(i32),
    Int8(i64),
    Float4(f32),
    Float8(f64),
    Numeric(Numeric),
    /// Days from 2000-01-01.
    Date(i32),
    /// Microseconds from midnight.
    Time(i64),
    /// Microseconds from midnight and the zone in seconds west of UTC.
    TimeTz(i64, i32),
    /// Microseconds from 2000-01-01 00:00:00.
    Timestamp(i64),
    /// Microseconds from 2000-01-01 00:00:00 UTC.
    TimestampTz(i64),
    Interval(Interval),
    Array(Box<Array>),
    /// A row value, whose fields print as Postgres prints records.
    Record(Vec<Value>),
    /// A json value: its text exactly as written.
    Json(String),
    Jsonb(Box<crate::json::Json>),
    /// An xml value: its text as written, which prints without a declaration that only repeats the defaults.
    Xml(String),
    /// A string of the text types, and the value of an untyped literal.
    Text(String),
    /// The rows of a set-returning function, which never reach a client.
    Set(Vec<Value>),
    /// An object identifier, as the oid, xid, and cid types hold.
    Oid(u32),
    /// A value of one of the reg types, such as regclass.
    Reg(Box<Reg>),
    /// A label of an enum type.
    Enum(Box<EnumValue>),
    /// A value of a composite type, whose fields print as a record's do.
    Composite(Box<CompositeValue>),
    /// A bytea value.
    Bytea(Vec<u8>),
    /// A uuid.
    Uuid([u8; 16]),
    /// A value of the bit and bit varying types, as its binary digits.
    Bit(String),
    /// A value of a base type that an extension provides.
    Base(Box<BaseValue>),
    /// A value of a range type.
    Range(Box<crate::rangetypes::Range>),
    /// A value of a multirange type.
    Multirange(Box<crate::rangetypes::Multirange>),
}

impl PartialEq for Value {
    /// eq compares values as Postgres' equal() compares constants, so NaN equals NaN as it does in Postgres.
    fn eq(&self, other: &Value) -> bool {
        match (self, other) {
            (Value::Null, Value::Null) => true,
            (Value::Bool(a), Value::Bool(b)) => a == b,
            (Value::Int2(a), Value::Int2(b)) => a == b,
            (Value::Int4(a), Value::Int4(b)) => a == b,
            (Value::Int8(a), Value::Int8(b)) => a == b,
            (Value::Float4(a), Value::Float4(b)) => a == b || (a.is_nan() && b.is_nan()),
            (Value::Float8(a), Value::Float8(b)) => a == b || (a.is_nan() && b.is_nan()),
            (Value::Numeric(a), Value::Numeric(b)) => a == b,
            (Value::Date(a), Value::Date(b)) => a == b,
            (Value::Time(a), Value::Time(b)) => a == b,
            (Value::TimeTz(a0, a1), Value::TimeTz(b0, b1)) => a0 == b0 && a1 == b1,
            (Value::Timestamp(a), Value::Timestamp(b)) => a == b,
            (Value::TimestampTz(a), Value::TimestampTz(b)) => a == b,
            (Value::Interval(a), Value::Interval(b)) => a == b,
            (Value::Array(a), Value::Array(b)) => a == b,
            (Value::Record(a), Value::Record(b)) => a == b,
            (Value::Json(a), Value::Json(b)) => a == b,
            (Value::Jsonb(a), Value::Jsonb(b)) => a == b,
            (Value::Xml(a), Value::Xml(b)) => a == b,
            (Value::Text(a), Value::Text(b)) => a == b,
            (Value::Set(a), Value::Set(b)) => a == b,
            (Value::Oid(a), Value::Oid(b)) => a == b,
            (Value::Reg(a), Value::Reg(b)) => a == b,
            (Value::Enum(a), Value::Enum(b)) => a == b,
            (Value::Composite(a), Value::Composite(b)) => a == b,
            (Value::Bytea(a), Value::Bytea(b)) => a == b,
            (Value::Uuid(a), Value::Uuid(b)) => a == b,
            (Value::Bit(a), Value::Bit(b)) => a == b,
            (Value::Base(a), Value::Base(b)) => a == b,
            (Value::Range(a), Value::Range(b)) => a == b,
            (Value::Multirange(a), Value::Multirange(b)) => a == b,
            _ => false,
        }
    }
}

/// BaseValue is a value of a base type that an extension provides, as the bytes that Doltgres stores for it.
#[derive(Clone, Debug, PartialEq)]
pub struct BaseValue {
    pub type_oid: u32,
    pub data: Vec<u8>,
}

/// EnumValue is a label of an enum type, which orders labels by their position in the type.
#[derive(Clone, Debug, PartialEq)]
pub struct EnumValue {
    pub type_oid: u32,
    pub label: String,
}

/// CompositeValue is a value of a composite type, whose attributes name its fields.
#[derive(Clone, Debug, PartialEq)]
pub struct CompositeValue {
    pub type_oid: u32,
    pub fields: Vec<Value>,
}

/// Reg is a value of a reg type: the type, the object's OID, and the name it prints as, which is the OID for a
/// missing object.
#[derive(Clone, Debug, PartialEq)]
pub struct Reg {
    pub type_oid: u32,
    pub oid: u32,
    pub name: String,
}

thread_local! {
    /// EXTRA_FLOAT_DIGITS is the extra_float_digits setting of the session running on this thread.
    static EXTRA_FLOAT_DIGITS: std::cell::Cell<i32> = const { std::cell::Cell::new(1) };
}

/// set_extra_float_digits sets the extra_float_digits that floats print with on this thread.
pub fn set_extra_float_digits(digits: i32) {
    EXTRA_FLOAT_DIGITS.with(|cell| cell.set(digits));
}

/// write_float appends a float's text as Postgres prints it. With a positive extra_float_digits that is its shortest
/// digits that read back as the float, which Postgres also finds with Ryu, with an exponent when the decimal exponent
/// is below -4 or at least `max_exponent`. Otherwise it is C's `%g` with `max_exponent` plus extra_float_digits
/// significant digits.
fn write_float(out: &mut Vec<u8>, value: f64, shortest: &str, max_exponent: i32) {
    if value.is_nan() {
        return out.extend_from_slice(b"NaN");
    }
    if value.is_infinite() {
        return out.extend_from_slice(if value < 0.0 { b"-Infinity" } else { b"Infinity" });
    }
    let extra = EXTRA_FLOAT_DIGITS.with(std::cell::Cell::get);
    if extra <= 0 {
        return write_general_float(out, value, (max_exponent + extra).max(1) as usize);
    }
    // Split ryu's text, which is either plain like 0.0001 or 123.0, or exponential like 1e-7 or 1.5e20, into its
    // digits and the decimal exponent of the first digit.
    let (sign, text) = match shortest.strip_prefix('-') {
        Some(text) => (true, text),
        None => (false, shortest),
    };
    let (mantissa, mut exponent) = match text.split_once('e') {
        Some((m, e)) => (m, e.parse::<i32>().unwrap_or(0)),
        None => (text, 0),
    };
    let (whole, fraction) = mantissa.split_once('.').unwrap_or((mantissa, ""));
    let mut digits: smallvec::SmallVec<[u8; 32]> = whole.bytes().chain(fraction.bytes()).collect();
    exponent += whole.len() as i32 - 1;
    let leading = digits.iter().take_while(|&&d| d == b'0').count().min(digits.len() - 1);
    digits.drain(..leading);
    exponent -= leading as i32;
    while digits.len() > 1 && digits.last() == Some(&b'0') {
        digits.pop();
    }
    if digits.as_slice() == b"0" {
        exponent = 0;
    }
    if sign {
        out.push(b'-');
    }
    if exponent < -4 || exponent >= max_exponent {
        out.push(digits[0]);
        if digits.len() > 1 {
            out.push(b'.');
            out.extend_from_slice(&digits[1..]);
        }
        out.extend_from_slice(if exponent < 0 { b"e-" } else { b"e+" });
        let abs = exponent.unsigned_abs();
        if abs < 10 {
            out.push(b'0');
        }
        out.extend_from_slice(itoa::Buffer::new().format(abs).as_bytes());
    } else if exponent < 0 {
        out.extend_from_slice(b"0.");
        out.extend(std::iter::repeat_n(b'0', (-exponent - 1) as usize));
        out.extend_from_slice(&digits);
    } else {
        let point = exponent as usize + 1;
        if digits.len() <= point {
            out.extend_from_slice(&digits);
            out.extend(std::iter::repeat_n(b'0', point - digits.len()));
        } else {
            out.extend_from_slice(&digits[..point]);
            out.push(b'.');
            out.extend_from_slice(&digits[point..]);
        }
    }
}

/// write_general_float appends a float as C's `%.*g` writes it with a precision of significant digits: in exponential
/// form when its decimal exponent is below -4 or at least the precision, and without trailing zeros.
fn write_general_float(out: &mut Vec<u8>, value: f64, precision: usize) {
    let scientific = format!("{value:.*e}", precision - 1);
    let (mantissa, exponent) = scientific.split_once('e').unwrap_or((&scientific, "0"));
    let exponent: i32 = exponent.parse().unwrap_or(0);
    let trim = |text: &str| -> String {
        match text.contains('.') {
            true => text.trim_end_matches('0').trim_end_matches('.').to_string(),
            false => text.to_string(),
        }
    };
    if exponent < -4 || exponent >= precision as i32 {
        out.extend_from_slice(trim(mantissa).as_bytes());
        let sign = if exponent < 0 { '-' } else { '+' };
        out.extend_from_slice(format!("e{sign}{:02}", exponent.unsigned_abs()).as_bytes());
    } else {
        let fixed = format!("{value:.*}", (precision as i32 - 1 - exponent).max(0) as usize);
        out.extend_from_slice(trim(&fixed).as_bytes());
    }
}

impl Value {
    /// nulls returns a row of NULLs, built faster than cloning NULL into each place.
    pub fn nulls(width: usize) -> Vec<Value> {
        std::iter::repeat_with(|| Value::Null).take(width).collect()
    }

    /// to_i64 returns the value of an integer.
    pub fn to_i64(&self) -> Option<i64> {
        match self {
            Value::Int2(n) => Some(i64::from(*n)),
            Value::Int4(n) => Some(i64::from(*n)),
            Value::Int8(n) => Some(*n),
            _ => None,
        }
    }

    /// is_null reports whether the value is NULL.
    pub fn is_null(&self) -> bool {
        matches!(self, Value::Null)
    }

    /// encode returns the value in the format for a column of the type, or None for NULL.
    pub fn encode(&self, type_oid: u32, format: i16) -> Option<Vec<u8>> {
        if format == BINARY_FORMAT {
            return self.send(type_oid);
        }
        self.output().map(String::into_bytes)
    }

    /// write_text appends the value's text format, as `output` returns it, reporting false for NULL, which appends
    /// nothing.
    pub fn write_text(&self, out: &mut Vec<u8>) -> bool {
        use std::io::Write;
        let mut integer = itoa::Buffer::new();
        let _ = match self {
            Value::Null | Value::Set(_) => return false,
            Value::Bool(b) => out.write_all(if *b { b"t" } else { b"f" }),
            Value::Int2(i) => out.write_all(integer.format(*i).as_bytes()),
            Value::Int4(i) => out.write_all(integer.format(*i).as_bytes()),
            Value::Int8(i) => out.write_all(integer.format(*i).as_bytes()),
            Value::Oid(o) => out.write_all(integer.format(*o).as_bytes()),
            Value::Float4(f) => {
                write_float(out, *f as f64, ryu::Buffer::new().format(*f), 6);
                Ok(())
            }
            Value::Float8(f) => {
                write_float(out, *f, ryu::Buffer::new().format(*f), 15);
                Ok(())
            }
            Value::Date(d) if *d != datetime::DATE_NOBEGIN && *d != datetime::DATE_NOEND => {
                match datetime::write_iso(out, *d as i64, None) {
                    true => Ok(()),
                    false => out.write_all(self.output().unwrap_or_default().as_bytes()),
                }
            }
            Value::Timestamp(ts) if *ts != datetime::TIMESTAMP_NOBEGIN && *ts != datetime::TIMESTAMP_NOEND => {
                let (days, time) = (ts.div_euclid(datetime::USECS_PER_DAY), ts.rem_euclid(datetime::USECS_PER_DAY));
                match datetime::write_iso(out, days, Some((time, None))) {
                    true => Ok(()),
                    false => out.write_all(self.output().unwrap_or_default().as_bytes()),
                }
            }
            Value::Numeric(n) => {
                n.write_text(out);
                Ok(())
            }
            Value::Text(s) | Value::Json(s) | Value::Bit(s) => out.write_all(s.as_bytes()),
            Value::Reg(reg) => out.write_all(reg.name.as_bytes()),
            Value::Enum(e) => out.write_all(e.label.as_bytes()),
            other => match other.output() {
                Some(text) => out.write_all(text.as_bytes()),
                None => return false,
            },
        };
        true
    }

    /// output returns the value's text format, or None for NULL.
    pub fn output(&self) -> Option<String> {
        Some(match self {
            Value::Null => return None,
            Value::Bool(b) => if *b { "t" } else { "f" }.to_string(),
            Value::Int2(i) => i.to_string(),
            Value::Int4(i) => i.to_string(),
            Value::Int8(i) => i.to_string(),
            Value::Float4(_) | Value::Float8(_) => {
                let mut out = Vec::new();
                self.write_text(&mut out);
                String::from_utf8(out).unwrap_or_default()
            }
            Value::Numeric(n) => n.to_string(),
            Value::Date(d) => datetime::with_format(|f| datetime::format_date(*d, f)),
            Value::Time(t) => datetime::format_time(*t),
            Value::TimeTz(t, z) => datetime::format_timetz(*t, *z),
            Value::Timestamp(ts) => datetime::with_format(|f| datetime::format_timestamp(*ts, None, f)),
            Value::TimestampTz(ts) => datetime::with_format(|f| {
                let (offset, name) = f.zone.offset_at(*ts);
                datetime::format_timestamp(*ts, Some((offset, &name)), f)
            }),
            Value::Interval(iv) => datetime::with_format(|f| datetime::format_interval(iv, f.interval_style)),
            Value::Array(a)
                if array::is_vector_type(a.element) && !a.values.iter().any(|v| matches!(v, Value::Array(_))) =>
            {
                a.values.iter().map(|v| v.output().unwrap_or_default()).collect::<Vec<_>>().join(" ")
            }
            Value::Array(a) => array::format(a, &|v| v.output().unwrap_or_default()),
            Value::Record(fields) => format_record(fields),
            Value::Json(text) => text.clone(),
            Value::Jsonb(json) => json.to_text(),
            Value::Xml(text) => crate::xml::output(text),
            Value::Text(s) => s.clone(),
            Value::Set(_) => return None,
            Value::Oid(o) => o.to_string(),
            Value::Reg(reg) => reg.name.clone(),
            Value::Enum(e) => e.label.clone(),
            Value::Composite(c) => format_record(&c.fields),
            Value::Bytea(bytes) => crate::binary::format_bytea(bytes),
            Value::Uuid(uuid) => crate::binary::format_uuid(uuid),
            Value::Bit(bits) => bits.clone(),
            Value::Base(base) => (base_type(base.type_oid)?.output)(&base.data),
            Value::Range(range) => crate::rangetypes::format(range),
            Value::Multirange(multirange) => crate::rangetypes::format_multirange(multirange),
        })
    }

    /// send returns the value's binary format, or None for NULL.
    fn send(&self, type_oid: u32) -> Option<Vec<u8>> {
        Some(match self {
            Value::Null => return None,
            Value::Bool(b) => vec![*b as u8],
            Value::Int2(i) => i.to_be_bytes().to_vec(),
            Value::Int4(i) => i.to_be_bytes().to_vec(),
            Value::Int8(i) => i.to_be_bytes().to_vec(),
            Value::Float4(f) => f.to_be_bytes().to_vec(),
            Value::Float8(f) => f.to_be_bytes().to_vec(),
            Value::Numeric(n) => n.send(),
            Value::Date(d) => d.to_be_bytes().to_vec(),
            Value::Time(t) => t.to_be_bytes().to_vec(),
            Value::TimeTz(t, z) => [t.to_be_bytes().as_slice(), &z.to_be_bytes()].concat(),
            Value::Timestamp(ts) | Value::TimestampTz(ts) => ts.to_be_bytes().to_vec(),
            Value::Interval(iv) => {
                [iv.micros.to_be_bytes().as_slice(), &iv.days.to_be_bytes(), &iv.months.to_be_bytes()].concat()
            }
            Value::Array(a) => {
                let element = a.element_type();
                array::send(a, &|v| v.send(element))
            }
            Value::Record(fields) => {
                let mut out = (fields.len() as i32).to_be_bytes().to_vec();
                for field in fields {
                    let field_type = match field {
                        Value::Null => oid::UNKNOWN,
                        other => crate::functions::value_type(other),
                    };
                    out.extend_from_slice(&field_type.to_be_bytes());
                    match field.send(field_type) {
                        Some(bytes) => {
                            out.extend_from_slice(&(bytes.len() as i32).to_be_bytes());
                            out.extend_from_slice(&bytes);
                        }
                        None => out.extend_from_slice(&(-1i32).to_be_bytes()),
                    }
                }
                out
            }
            Value::Json(_) => return self.output().map(String::into_bytes),
            Value::Xml(text) => crate::xml::send(text).into_bytes(),
            Value::Jsonb(json) => [&[1u8][..], json.to_text().as_bytes()].concat(),
            Value::Text(_) if type_oid == oid::UNKNOWN => return self.output().map(String::into_bytes),
            Value::Text(s) if type_oid == oid::JSONPATH => [&[1u8][..], s.as_bytes()].concat(),
            Value::Text(s) if type_oid == oid::CHAR => match s.strip_prefix('\\') {
                Some(octal) if octal.len() == 3 => vec![u8::from_str_radix(octal, 8).unwrap_or(0)],
                _ => s.bytes().take(1).collect(),
            },
            Value::Text(s) => s.clone().into_bytes(),
            Value::Set(_) => return None,
            Value::Oid(o) => o.to_be_bytes().to_vec(),
            Value::Reg(reg) => reg.oid.to_be_bytes().to_vec(),
            Value::Enum(e) => e.label.clone().into_bytes(),
            Value::Composite(c) => {
                let types: Vec<u32> = match crate::usertypes::get(c.type_oid).map(|t| t.kind.clone()) {
                    Some(crate::usertypes::Kind::Composite(attributes)) => {
                        attributes.iter().map(|(_, t)| t.oid).collect()
                    }
                    _ => Vec::new(),
                };
                let mut out = (c.fields.len() as i32).to_be_bytes().to_vec();
                for (i, field) in c.fields.iter().enumerate() {
                    let field_type = types.get(i).copied().unwrap_or_else(|| crate::functions::value_type(field));
                    out.extend_from_slice(&field_type.to_be_bytes());
                    match field.send(field_type) {
                        Some(bytes) => {
                            out.extend_from_slice(&(bytes.len() as i32).to_be_bytes());
                            out.extend_from_slice(&bytes);
                        }
                        None => out.extend_from_slice(&(-1i32).to_be_bytes()),
                    }
                }
                out
            }
            Value::Bytea(bytes) => bytes.clone(),
            Value::Uuid(uuid) => uuid.to_vec(),
            Value::Bit(bits) => {
                [(bits.len() as i32).to_be_bytes().as_slice(), &crate::binary::pack_bits(bits)].concat()
            }
            Value::Base(base) => (base_type(base.type_oid)?.send)(&base.data),
            Value::Range(range) => {
                let subtype = crate::rangetypes::range_type(range.type_oid).map_or(oid::TEXT, |t| t.subtype);
                crate::rangetypes::send(range, &|v| Ok(v.send(subtype).unwrap_or_default())).ok()?
            }
            Value::Multirange(multirange) => {
                let subtype = crate::rangetypes::multirange_type(multirange.type_oid).map_or(oid::TEXT, |t| t.subtype);
                crate::rangetypes::send_multirange(multirange, &|v| Ok(v.send(subtype).unwrap_or_default())).ok()?
            }
        })
    }

    /// decode returns a parameter value sent in the format for the type, where a zero type OID means unspecified, and
    /// the text of a reg type stays text for the session to look up.
    pub fn decode(type_oid: u32, format: i16, bytes: Option<&[u8]>) -> Result<Value> {
        let Some(bytes) = bytes else { return Ok(Value::Null) };
        if format == BINARY_FORMAT && array::is_array_type(type_oid) {
            let element = crate::catalog::builtin_type(type_oid).map_or(0, |t| t.elem);
            let parsed = array::receive(bytes, &|oid, data| Value::decode(oid, BINARY_FORMAT, Some(data)))?;
            return Ok(Value::Array(Box::new(Array { element, ..parsed })));
        }
        let invalid =
            || PgError::new(code::INVALID_BINARY_REPRESENTATION, "incorrect binary data format in bind parameter");
        if format == BINARY_FORMAT {
            return match type_oid {
                oid::BOOL => Ok(Value::Bool(*bytes.first().ok_or_else(invalid)? != 0)),
                oid::CHAR => Ok(Value::Text(match bytes.first() {
                    None => String::new(),
                    Some(&b) if b.is_ascii() => (b as char).to_string(),
                    Some(&b) => format!("\\{b:03o}"),
                })),
                oid::JSONB => match bytes.split_first() {
                    Some((1, text)) => {
                        crate::cast::input(std::str::from_utf8(text).map_err(|_| invalid())?, oid::JSONB)
                    }
                    _ => Err(PgError::new(code::INVALID_BINARY_REPRESENTATION, "unsupported jsonb version number")),
                },
                oid::INT2 => Ok(Value::Int2(i16::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::INT4 => Ok(Value::Int4(i32::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::INT8 => Ok(Value::Int8(i64::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::OID | oid::XID => Ok(Value::Oid(u32::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::FLOAT4 => Ok(Value::Float4(f32::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::FLOAT8 => Ok(Value::Float8(f64::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::NUMERIC => Numeric::receive(bytes).map(Value::Numeric).ok_or_else(invalid),
                oid::DATE => Ok(Value::Date(i32::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::TIME => Ok(Value::Time(i64::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::TIMESTAMP => Ok(Value::Timestamp(i64::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::TIMESTAMPTZ => {
                    Ok(Value::TimestampTz(i64::from_be_bytes(bytes.try_into().map_err(|_| invalid())?)))
                }
                oid::TIMETZ if bytes.len() == 12 => Ok(Value::TimeTz(
                    i64::from_be_bytes(bytes[..8].try_into().map_err(|_| invalid())?),
                    i32::from_be_bytes(bytes[8..].try_into().map_err(|_| invalid())?),
                )),
                oid::INTERVAL if bytes.len() == 16 => Ok(Value::Interval(Interval {
                    micros: i64::from_be_bytes(bytes[..8].try_into().map_err(|_| invalid())?),
                    days: i32::from_be_bytes(bytes[8..12].try_into().map_err(|_| invalid())?),
                    months: i32::from_be_bytes(bytes[12..].try_into().map_err(|_| invalid())?),
                })),
                oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME | oid::UNKNOWN | 0 => {
                    Ok(Value::Text(String::from_utf8(bytes.to_vec()).map_err(|_| invalid())?))
                }
                oid::JSON => Ok(Value::Json(String::from_utf8(bytes.to_vec()).map_err(|_| invalid())?)),
                oid::XML => crate::cast::input(std::str::from_utf8(bytes).map_err(|_| invalid())?, oid::XML),
                oid::JSONPATH if bytes.first() == Some(&1) => {
                    crate::cast::input(std::str::from_utf8(&bytes[1..]).map_err(|_| invalid())?, oid::JSONPATH)
                }
                oid::BYTEA => Ok(Value::Bytea(bytes.to_vec())),
                oid::UUID => Ok(Value::Uuid(bytes.try_into().map_err(|_| invalid())?)),
                oid::BIT | oid::VARBIT if bytes.len() >= 4 => {
                    let length = i32::from_be_bytes(bytes[..4].try_into().map_err(|_| invalid())?);
                    let length = usize::try_from(length).map_err(|_| invalid())?;
                    if bytes.len() - 4 != length.div_ceil(8) {
                        return Err(invalid());
                    }
                    Ok(Value::Bit(crate::binary::unpack_bits(&bytes[4..], length)))
                }
                oid::CID => Ok(Value::Oid(u32::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                t if crate::cast::is_reg_type(t) => {
                    Ok(Value::Text(u32::from_be_bytes(bytes.try_into().map_err(|_| invalid())?).to_string()))
                }
                _ => match (base_type(type_oid), crate::usertypes::get(type_oid).map(|t| t.kind.clone())) {
                    (Some(base), _) => {
                        Ok(Value::Base(Box::new(BaseValue { type_oid, data: (base.receive)(bytes, -1)? })))
                    }
                    (None, Some(crate::usertypes::Kind::Enum(_))) => Value::decode(type_oid, 0, Some(bytes)),
                    (None, Some(crate::usertypes::Kind::Domain(domain))) => {
                        Value::decode(domain.base.oid, BINARY_FORMAT, Some(bytes))
                    }
                    (None, Some(crate::usertypes::Kind::Composite(attributes))) => {
                        receive_record(type_oid, &attributes, bytes)
                    }
                    _ => Err(PgError::unsupported(format!("binary parameters of type {type_oid}"))),
                },
            };
        }
        let text = std::str::from_utf8(bytes).map_err(|_| {
            PgError::new(code::CHARACTER_NOT_IN_REPERTOIRE, "invalid byte sequence for encoding \"UTF8\"")
        })?;
        match type_oid {
            oid::TEXT | oid::UNKNOWN | 0 => Ok(Value::Text(text.to_string())),
            _ if crate::cast::is_reg_type(type_oid) => Ok(Value::Text(text.to_string())),
            _ => crate::cast::input(text, type_oid),
        }
    }
}

/// base_type returns the definition of a base type that a built-in type Go lacks or an extension provides.
pub fn base_type(type_oid: u32) -> Option<&'static crate::extensions::BaseType> {
    if let Some(definition) = crate::basetypes::get(type_oid) {
        return Some(definition);
    }
    match crate::usertypes::get(type_oid)?.kind {
        crate::usertypes::Kind::Base(definition) => Some(definition),
        _ => None,
    }
}

/// receive_record reads a composite value from Postgres' binary record format, the field count, then each field's type
/// OID, length, and data, as record_recv does.
fn receive_record(type_oid: u32, attributes: &[(String, crate::catalog::ColumnType)], bytes: &[u8]) -> Result<Value> {
    let short = || PgError::new(code::PROTOCOL_VIOLATION, "insufficient data left in message");
    let mut at = 0;
    let mut take = |n: usize| -> Result<&[u8]> {
        let slice = bytes.get(at..at + n).ok_or_else(short)?;
        at += n;
        Ok(slice)
    };
    let count = i32::from_be_bytes(take(4)?.try_into().map_err(|_| short())?);
    if count != attributes.len() as i32 {
        return Err(PgError::new(
            code::DATATYPE_MISMATCH,
            format!("wrong number of columns: {count}, expected {}", attributes.len()),
        ));
    }
    let mut fields = Vec::with_capacity(attributes.len());
    for (_, ty) in attributes {
        let field_type = u32::from_be_bytes(take(4)?.try_into().map_err(|_| short())?);
        if field_type != ty.oid {
            return Err(PgError::new(
                code::DATATYPE_MISMATCH,
                format!("wrong data type: {field_type}, expected {}", ty.oid),
            ));
        }
        let length = i32::from_be_bytes(take(4)?.try_into().map_err(|_| short())?);
        let data = match length {
            -1 => None,
            n if n < -1 => return Err(short()),
            n => Some(take(n as usize)?),
        };
        let value = Value::decode(field_type, BINARY_FORMAT, data)?;
        fields.push(crate::cast::cast_value(value, *ty, false)?);
    }
    Ok(Value::Composite(Box::new(CompositeValue { type_oid, fields })))
}

/// format_record prints a record's fields as Postgres' record_out does, quoting fields that need it.
fn format_record(fields: &[Value]) -> String {
    let mut out = String::from("(");
    for (i, field) in fields.iter().enumerate() {
        if i > 0 {
            out.push(',');
        }
        let Some(text) = field.output() else { continue };
        let quote = text.is_empty()
            || text.chars().any(|c| matches!(c, '"' | '\\' | '(' | ')' | ',') || c.is_ascii_whitespace());
        if quote {
            out.push('"');
            for c in text.chars() {
                if c == '"' || c == '\\' {
                    out.push(c);
                }
                out.push(c);
            }
            out.push('"');
        } else {
            out.push_str(&text);
        }
    }
    out.push(')');
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn floats_print_as_postgres_does() {
        let float8 = |f: f64| Value::Float8(f).output().unwrap();
        assert_eq!(float8(1.0), "1");
        assert_eq!(float8(1.5), "1.5");
        assert_eq!(float8(0.0001), "0.0001");
        assert_eq!(float8(0.00001), "1e-05");
        assert_eq!(float8(123456789012345.0), "123456789012345");
        assert_eq!(float8(1e15), "1e+15");
        assert_eq!(float8(-2.5e-7), "-2.5e-07");
        assert_eq!(float8(0.1 + 0.2), "0.30000000000000004");
        assert_eq!(float8(f64::INFINITY), "Infinity");
        assert_eq!(float8(-0.0), "-0");
        assert_eq!(Value::Float4(1.25).output().unwrap(), "1.25");
        assert_eq!(Value::Float4(1234567.0).output().unwrap(), "1.234567e+06");
    }

    #[test]
    fn written_text_matches_output() {
        let mut values = vec![
            Value::Null,
            Value::Bool(true),
            Value::Bool(false),
            Value::Int2(-32768),
            Value::Int4(0),
            Value::Int8(i64::MIN),
            Value::Oid(4294967295),
            Value::Text("héllo".into()),
            Value::Date(0),
            Value::Date(-800_000),
            Value::Date(crate::datetime::DATE_NOEND),
            Value::Timestamp(0),
            Value::Timestamp(1_234_567_890_123_456),
            Value::Timestamp(-63_000_000_000_000_000),
            Value::Timestamp(86_399_999_999),
            Value::Timestamp(crate::datetime::TIMESTAMP_NOBEGIN),
            Value::Numeric(crate::numeric::Numeric::parse("0.000123").unwrap()),
            Value::Numeric(crate::numeric::Numeric::parse("-12.50").unwrap()),
            Value::Numeric(crate::numeric::Numeric::parse("0").unwrap()),
            Value::Numeric(crate::numeric::Numeric::parse("0.00").unwrap()),
            Value::Numeric(crate::numeric::Numeric::parse("123456789012345678901234567890123456789.5").unwrap()),
            Value::Numeric(crate::numeric::Numeric::parse("NaN").unwrap()),
            Value::Numeric(crate::numeric::Numeric::parse("-Infinity").unwrap()),
        ];
        let mut seed: u64 = 7;
        for _ in 0..5000 {
            seed ^= seed << 13;
            seed ^= seed >> 7;
            seed ^= seed << 17;
            values.push(Value::Float8(f64::from_bits(seed)));
            values.push(Value::Float4(f32::from_bits(seed as u32)));
            values.push(Value::Float8((seed % 100_000) as f64 / 7.0));
            values.push(Value::Timestamp((seed >> 8) as i64 - (1 << 55)));
        }
        for f in [0.0, -0.0, 1.0, 1e15, 1e-5, 123456789012345.0, f64::NAN, f64::INFINITY, f64::NEG_INFINITY, 0.1 + 0.2]
        {
            values.push(Value::Float8(f));
            values.push(Value::Float4(f as f32));
        }
        for value in values {
            let mut out = Vec::new();
            let written = value.write_text(&mut out);
            assert_eq!(written.then_some(out), value.output().map(String::into_bytes), "{value:?}");
        }
    }
}
