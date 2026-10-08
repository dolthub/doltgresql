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
#[derive(Clone, Debug, PartialEq)]
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

/// format_float formats a float as Postgres does with the default extra_float_digits: the shortest digits that read
/// back exactly, in exponential notation when the exponent is below -4 or at least `max_exponent`.
fn format_float(shortest_exponential: String, max_exponent: i32) -> String {
    match shortest_exponential.as_str() {
        "NaN" => return "NaN".into(),
        "inf" => return "Infinity".into(),
        "-inf" => return "-Infinity".into(),
        _ => {}
    }
    let (mantissa, exponent) = shortest_exponential.split_once('e').unwrap_or((&shortest_exponential, "0"));
    let exponent: i32 = exponent.parse().unwrap_or(0);
    let (sign, mantissa) = mantissa.strip_prefix('-').map_or(("", mantissa), |m| ("-", m));
    let digits: String = mantissa.chars().filter(char::is_ascii_digit).collect();
    if exponent < -4 || exponent >= max_exponent {
        let rest = if digits.len() > 1 { format!(".{}", &digits[1..]) } else { String::new() };
        let exponent_sign = if exponent < 0 { '-' } else { '+' };
        return format!("{sign}{}{rest}e{exponent_sign}{:02}", &digits[..1], exponent.abs());
    }
    if exponent < 0 {
        return format!("{sign}0.{}{digits}", "0".repeat((-exponent - 1) as usize));
    }
    let point = exponent as usize + 1;
    if digits.len() <= point {
        format!("{sign}{digits}{}", "0".repeat(point - digits.len()))
    } else {
        format!("{sign}{}.{}", &digits[..point], &digits[point..])
    }
}

impl Value {
    /// nulls returns a row of NULLs, built faster than cloning NULL into each place.
    pub fn nulls(width: usize) -> Vec<Value> {
        std::iter::repeat_with(|| Value::Null).take(width).collect()
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

    /// output returns the value's text format, or None for NULL.
    pub fn output(&self) -> Option<String> {
        Some(match self {
            Value::Null => return None,
            Value::Bool(b) => if *b { "t" } else { "f" }.to_string(),
            Value::Int2(i) => i.to_string(),
            Value::Int4(i) => i.to_string(),
            Value::Int8(i) => i.to_string(),
            Value::Float4(f) => format_float(format!("{f:e}"), 6),
            Value::Float8(f) => format_float(format!("{f:e}"), 15),
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
}
