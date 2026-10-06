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

//! Casts between types, and reading values from their text format.

use crate::catalog::{ColumnType, builtin_type};
use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;
use crate::oid;
use crate::types::Value;

/// type_display returns the name Postgres uses for a type in error messages.
pub fn type_display(type_oid: u32) -> std::borrow::Cow<'static, str> {
    if crate::array::is_array_type(type_oid) {
        let element = builtin_type(type_oid).map_or(0, |t| t.elem);
        return format!("{}[]", type_display(element)).into();
    }
    match type_oid {
        oid::BOOL => "boolean",
        oid::INT2 => "smallint",
        oid::INT4 => "integer",
        oid::INT8 => "bigint",
        oid::FLOAT4 => "real",
        oid::FLOAT8 => "double precision",
        oid::NUMERIC => "numeric",
        oid::DATE => "date",
        oid::TIME => "time without time zone",
        oid::TIMETZ => "time with time zone",
        oid::TIMESTAMP => "timestamp without time zone",
        oid::TIMESTAMPTZ => "timestamp with time zone",
        oid::INTERVAL => "interval",
        oid::VARCHAR => "character varying",
        oid::BPCHAR => "character",
        oid::CHAR => "\"char\"",
        _ => builtin_type(type_oid).map_or("unknown", |t| t.name),
    }
    .into()
}

/// jsonb_scalar converts a jsonb number or boolean to a numeric or boolean type, as Postgres' jsonb casts do.
fn jsonb_scalar(json: &crate::json::Json, to: ColumnType) -> Result<Value> {
    use crate::json::Json;
    match (json, to.oid) {
        (Json::Bool(b), oid::BOOL) => Ok(Value::Bool(*b)),
        (Json::Number(n), target) if target != oid::BOOL => cast_value(Value::Numeric(n.clone()), to, true),
        _ => {
            let kind = if matches!(json, Json::Number(_)) { "numeric" } else { json.type_name() };
            Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                format!("cannot cast jsonb {kind} to type {}", type_display(to.oid)),
            ))
        }
    }
}

/// invalid_syntax returns Postgres' error for text that is not a value of the type.
pub fn invalid_syntax(type_oid: u32, text: &str) -> PgError {
    PgError::new(
        code::INVALID_TEXT_REPRESENTATION,
        format!("invalid input syntax for type {}: \"{text}\"", type_display(type_oid)),
    )
}

/// out_of_range returns Postgres' error for text whose value is outside the type's range.
pub fn out_of_range(type_oid: u32, text: &str) -> PgError {
    PgError::new(
        code::NUMERIC_VALUE_OUT_OF_RANGE,
        format!("value \"{text}\" is out of range for type {}", type_display(type_oid)),
    )
}

/// parse_integer reads an integer as Postgres' integer input functions do: optional whitespace, a sign, and digits
/// (with underscores between digit groups, or a 0x, 0o, or 0b prefix).
fn parse_integer(text: &str, type_oid: u32, min: i128, max: i128) -> Result<i128> {
    let trimmed = text.trim_matches(|c: char| c.is_ascii_whitespace());
    let (negative, digits) = match trimmed.as_bytes().first() {
        Some(b'-') => (true, &trimmed[1..]),
        Some(b'+') => (false, &trimmed[1..]),
        _ => (false, trimmed),
    };
    let (radix, digits) = match digits.get(..2).map(str::to_ascii_lowercase).as_deref() {
        Some("0x") => (16, &digits[2..]),
        Some("0o") => (8, &digits[2..]),
        Some("0b") => (2, &digits[2..]),
        _ => (10, digits),
    };
    let valid = !digits.is_empty()
        && !digits.starts_with('_')
        && !digits.ends_with('_')
        && !digits.contains("__")
        && digits.chars().all(|c| c == '_' || c.is_digit(radix));
    if !valid {
        return Err(invalid_syntax(type_oid, text));
    }
    let mut value: i128 = 0;
    for c in digits.chars().filter(|&c| c != '_') {
        value = value * radix as i128 + c.to_digit(radix).unwrap() as i128;
        if value > max + 1 {
            return Err(out_of_range(type_oid, text));
        }
    }
    let value = if negative { -value } else { value };
    if value < min || value > max {
        return Err(out_of_range(type_oid, text));
    }
    Ok(value)
}

/// parse_float reads a float as Postgres' float input functions do.
fn parse_float(text: &str, type_oid: u32) -> Result<f64> {
    let trimmed = text.trim_matches(|c: char| c.is_ascii_whitespace());
    let lower = trimmed.to_ascii_lowercase();
    let special = match lower.as_str() {
        "nan" | "+nan" | "-nan" => Some(f64::NAN),
        "infinity" | "+infinity" | "inf" | "+inf" => Some(f64::INFINITY),
        "-infinity" | "-inf" => Some(f64::NEG_INFINITY),
        _ => None,
    };
    if let Some(value) = special {
        return Ok(value);
    }
    let valid = lower.chars().all(|c| c.is_ascii_digit() || matches!(c, '.' | 'e' | '+' | '-'))
        && lower.chars().any(|c| c.is_ascii_digit());
    let value: f64 = match lower.parse() {
        Ok(value) if valid => value,
        _ => return Err(invalid_syntax(type_oid, text)),
    };
    let float4 = type_oid == oid::FLOAT4;
    let mantissa_nonzero = lower.split('e').next().unwrap_or_default().chars().any(|c| ('1'..='9').contains(&c));
    let overflow = value.is_infinite() || (float4 && (value as f32).is_infinite());
    let underflow = mantissa_nonzero && (value == 0.0 || (float4 && (value as f32) == 0.0));
    if overflow || underflow {
        return Err(PgError::new(
            code::NUMERIC_VALUE_OUT_OF_RANGE,
            format!("\"{trimmed}\" is out of range for type {}", type_display(type_oid)),
        ));
    }
    Ok(value)
}

/// parse_bool reads a boolean as Postgres' boolin does: any unique prefix of true, false, yes, no, on, or off, or 1
/// or 0, ignoring case and surrounding whitespace.
fn parse_bool(text: &str) -> Result<bool> {
    let word = text.trim_matches(|c: char| c.is_ascii_whitespace()).to_ascii_lowercase();
    let matches = |full: &str, min: usize| word.len() >= min && full.starts_with(&word);
    if matches("true", 1) || matches("yes", 1) || matches("on", 2) || word == "1" {
        Ok(true)
    } else if matches("false", 1) || matches("no", 1) || matches("off", 2) || word == "0" {
        Ok(false)
    } else {
        Err(invalid_syntax(oid::BOOL, text))
    }
}

/// is_reg_type reports whether a type is one of the reg types, whose values name catalog objects.
pub fn is_reg_type(type_oid: u32) -> bool {
    matches!(
        type_oid,
        oid::REGPROC
            | oid::REGPROCEDURE
            | oid::REGOPER
            | oid::REGOPERATOR
            | oid::REGCLASS
            | oid::REGTYPE
            | oid::REGNAMESPACE
            | oid::REGROLE
    )
}

/// parse_oid reads an OID as Postgres' oidin does, where negative numbers wrap around.
fn parse_oid(text: &str, type_oid: u32) -> Result<u32> {
    let trimmed = text.trim_matches(|c: char| c.is_ascii_whitespace());
    let digits = trimmed.strip_prefix(['-', '+']).unwrap_or(trimmed);
    if digits.is_empty() || !digits.bytes().all(|b| b.is_ascii_digit()) {
        return Err(invalid_syntax(type_oid, text));
    }
    let value: i128 = trimmed.parse().map_err(|_| out_of_range(type_oid, text))?;
    if !(i32::MIN as i128..=u32::MAX as i128).contains(&value) {
        return Err(out_of_range(type_oid, text));
    }
    Ok(value as i64 as u32)
}

/// char_value returns the value of the "char" type for text: its first byte, written as an octal escape when it is not
/// ASCII.
fn char_value(text: &str) -> Value {
    Value::Text(match text.as_bytes().first() {
        None => String::new(),
        Some(&b) if b.is_ascii() => (b as char).to_string(),
        Some(&b) => format!("\\{b:03o}"),
    })
}

/// is_char_value reports whether text is already a value of the "char" type: one ASCII character or an octal escape.
fn is_char_value(text: &str) -> bool {
    text.len() <= 1
        || (text.len() == 4 && text.starts_with('\\') && text[1..].bytes().all(|b| (b'0'..=b'7').contains(&b)))
}

/// input reads a value of the type from its text format.
pub fn input(text: &str, type_oid: u32) -> Result<Value> {
    if crate::array::is_array_type(type_oid) {
        let element = builtin_type(type_oid).map_or(oid::TEXT, |t| t.elem);
        let parsed = crate::array::parse(text, element, &|item| input(item, element))?;
        return Ok(Value::Array(Box::new(parsed)));
    }
    Ok(match type_oid {
        oid::JSON => {
            crate::json::parse(text, false)?;
            Value::Json(text.to_string())
        }
        oid::JSONB => Value::Jsonb(Box::new(crate::json::parse(text, true)?)),
        oid::BOOL => Value::Bool(parse_bool(text)?),
        oid::INT2 => Value::Int2(parse_integer(text, type_oid, i16::MIN as i128, i16::MAX as i128)? as i16),
        oid::INT4 => Value::Int4(parse_integer(text, type_oid, i32::MIN as i128, i32::MAX as i128)? as i32),
        oid::INT8 => Value::Int8(parse_integer(text, type_oid, i64::MIN as i128, i64::MAX as i128)? as i64),
        oid::FLOAT4 => Value::Float4(parse_float(text, type_oid)? as f32),
        oid::FLOAT8 => Value::Float8(parse_float(text, type_oid)?),
        oid::NUMERIC => Value::Numeric(Numeric::parse(text)?),
        oid::DATE => crate::datetime::with_format(|f| crate::datetime::parse_date(text, f, crate::datetime::now()))
            .map(Value::Date)?,
        oid::TIME => crate::datetime::with_format(|f| crate::datetime::parse_time(text, f)).map(Value::Time)?,
        oid::TIMETZ => {
            let (time, zone) =
                crate::datetime::with_format(|f| crate::datetime::parse_timetz(text, f, crate::datetime::now()))?;
            Value::TimeTz(time, zone)
        }
        oid::TIMESTAMP => {
            crate::datetime::with_format(|f| crate::datetime::parse_timestamp(text, false, f, crate::datetime::now()))
                .map(Value::Timestamp)?
        }
        oid::TIMESTAMPTZ => {
            crate::datetime::with_format(|f| crate::datetime::parse_timestamp(text, true, f, crate::datetime::now()))
                .map(Value::TimestampTz)?
        }
        oid::INTERVAL => Value::Interval(crate::datetime::parse_interval(text)?),
        oid::OID | oid::XID | oid::CID => Value::Oid(parse_oid(text, type_oid)?),
        oid::CHAR => char_value(text),
        oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME | oid::UNKNOWN => Value::Text(text.to_string()),
        _ => return Err(PgError::unsupported(format!("reading values of type {}", type_display(type_oid)))),
    })
}

/// int_out_of_range returns Postgres' error for an integer cast whose value does not fit the type.
fn int_out_of_range(type_oid: u32) -> PgError {
    let name = match type_oid {
        oid::INT2 => "smallint",
        oid::INT4 => "integer",
        _ => "bigint",
    };
    PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, format!("{name} out of range"))
}

/// to_integer converts a value to an integer type, rounding floats half to even as Postgres' rint does.
fn to_integer(value: Value, type_oid: u32) -> Result<Value> {
    let wide: i64 = match value {
        Value::Int2(i) => i as i64,
        Value::Int4(i) => i as i64,
        Value::Int8(i) => i,
        Value::Float4(f) => float_to_i64(f as f64, type_oid)?,
        Value::Float8(f) => float_to_i64(f, type_oid)?,
        Value::Numeric(n) => match &n {
            Numeric::NaN => {
                return Err(PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    format!("cannot convert NaN to {}", type_display(type_oid)),
                ));
            }
            Numeric::Infinity | Numeric::NegativeInfinity => {
                return Err(PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    format!("cannot convert infinity to {}", type_display(type_oid)),
                ));
            }
            _ => n.to_i64().ok_or_else(|| int_out_of_range(type_oid))?,
        },
        Value::Bool(b) if type_oid == oid::INT4 => b as i64,
        Value::Oid(o) if type_oid == oid::INT4 => o as i32 as i64,
        Value::Reg(reg) if type_oid == oid::INT4 => reg.oid as i32 as i64,
        Value::Oid(o) => o as i64,
        Value::Reg(reg) => reg.oid as i64,
        Value::Text(text) => return input(&text, type_oid),
        other => return Err(cannot_cast(&other, type_oid)),
    };
    match type_oid {
        oid::INT2 => i16::try_from(wide).map(Value::Int2).map_err(|_| int_out_of_range(type_oid)),
        oid::INT4 => i32::try_from(wide).map(Value::Int4).map_err(|_| int_out_of_range(type_oid)),
        _ => Ok(Value::Int8(wide)),
    }
}

/// float_to_i64 rounds a float half to even and converts it to a 64-bit integer.
fn float_to_i64(value: f64, type_oid: u32) -> Result<i64> {
    let rounded = value.round_ties_even();
    if !rounded.is_finite() || rounded < -9.223372036854776e18 || rounded >= 9.223372036854776e18 {
        return Err(int_out_of_range(type_oid));
    }
    Ok(rounded as i64)
}

/// cannot_cast returns the error for a value that has no cast to the type.
fn cannot_cast(value: &Value, type_oid: u32) -> PgError {
    PgError::new(
        code::CANNOT_COERCE,
        format!(
            "cannot cast {} to {}",
            format!("{value:?}").split('(').next().unwrap_or_default(),
            type_display(type_oid)
        ),
    )
}

/// apply_length applies a varchar or char length limit, truncating an explicit cast and rejecting an assignment that
/// would drop characters other than spaces, and pads a char value to its length.
fn apply_length(text: String, to: ColumnType, explicit: bool) -> Result<String> {
    if to.modifier < 4 || !matches!(to.oid, oid::VARCHAR | oid::BPCHAR) {
        return Ok(text);
    }
    let length = (to.modifier - 4) as usize;
    let count = text.chars().count();
    if count > length {
        let cut: usize = text.char_indices().nth(length).map_or(text.len(), |(i, _)| i);
        if !explicit && !text[cut..].chars().all(|c| c == ' ') {
            let name = if to.oid == oid::VARCHAR { "character varying" } else { "character" };
            return Err(PgError::new(
                code::STRING_DATA_RIGHT_TRUNCATION,
                format!("value too long for type {name}({length})"),
            ));
        }
        return Ok(text[..cut].to_string());
    }
    if to.oid == oid::BPCHAR {
        return Ok(format!("{text}{}", " ".repeat(length - count)));
    }
    Ok(text)
}

/// round_micros rounds microseconds to a precision of fractional digits, half away from zero.
fn round_micros(micros: i64, precision: i32) -> i64 {
    if !(0..6).contains(&precision) {
        return micros;
    }
    let unit = 10i64.pow(6 - precision as u32);
    let half = unit / 2;
    if micros >= 0 { (micros + half) / unit * unit } else { -((-micros + half) / unit * unit) }
}

/// cast_datetime converts a value to a date, time, timestamp, or interval type, in the session's time zone where a
/// zone matters.
fn cast_datetime(value: Value, to: ColumnType) -> Result<Value> {
    use crate::datetime::{self as dt, USECS_PER_DAY, USECS_PER_SEC};
    let zone_offset = |utc: i64| dt::with_format(|f| f.zone.offset_at(utc).0) as i64 * USECS_PER_SEC;
    let local_offset = |local: i64| dt::with_format(|f| f.zone.offset_for_local(local)) as i64 * USECS_PER_SEC;
    let finite = |ts: i64| ts != dt::TIMESTAMP_NOBEGIN && ts != dt::TIMESTAMP_NOEND;
    let date_to_timestamp = |d: i32| match d {
        dt::DATE_NOBEGIN => dt::TIMESTAMP_NOBEGIN,
        dt::DATE_NOEND => dt::TIMESTAMP_NOEND,
        d => d as i64 * USECS_PER_DAY,
    };
    let timestamp_to_date = |ts: i64| match ts {
        dt::TIMESTAMP_NOBEGIN => dt::DATE_NOBEGIN,
        dt::TIMESTAMP_NOEND => dt::DATE_NOEND,
        ts => ts.div_euclid(USECS_PER_DAY) as i32,
    };
    let result = match (value, to.oid) {
        (Value::Text(text), _) => return input(&text, to.oid).and_then(|v| cast_datetime(v, to)),
        (v @ Value::Date(_), oid::DATE) => v,
        (Value::Timestamp(ts), oid::DATE) => Value::Date(timestamp_to_date(ts)),
        (Value::TimestampTz(ts), oid::DATE) => {
            Value::Date(timestamp_to_date(if finite(ts) { ts + zone_offset(ts) } else { ts }))
        }
        (Value::Date(d), oid::TIMESTAMP) => Value::Timestamp(date_to_timestamp(d)),
        (Value::Timestamp(ts), oid::TIMESTAMP) => {
            Value::Timestamp(if finite(ts) { round_micros(ts, to.modifier) } else { ts })
        }
        (Value::TimestampTz(ts), oid::TIMESTAMP) => {
            Value::Timestamp(if finite(ts) { ts + zone_offset(ts) } else { ts })
        }
        (Value::Date(d), oid::TIMESTAMPTZ) => {
            let local = date_to_timestamp(d);
            Value::TimestampTz(if finite(local) { local - local_offset(local) } else { local })
        }
        (Value::Timestamp(ts), oid::TIMESTAMPTZ) => {
            Value::TimestampTz(if finite(ts) { ts - local_offset(ts) } else { ts })
        }
        (Value::TimestampTz(ts), oid::TIMESTAMPTZ) => {
            Value::TimestampTz(if finite(ts) { round_micros(ts, to.modifier) } else { ts })
        }
        (Value::Time(t), oid::TIME) => Value::Time(round_micros(t, to.modifier)),
        (Value::TimeTz(t, _), oid::TIME) => Value::Time(t),
        (Value::Timestamp(ts), oid::TIME) => Value::Time(ts.rem_euclid(USECS_PER_DAY)),
        (Value::TimestampTz(ts), oid::TIME) => Value::Time((ts + zone_offset(ts)).rem_euclid(USECS_PER_DAY)),
        (Value::Interval(iv), oid::TIME) => Value::Time(iv.micros.rem_euclid(USECS_PER_DAY)),
        (Value::Time(t), oid::TIMETZ) => Value::TimeTz(t, -(zone_offset(dt::now().timestamp) / USECS_PER_SEC) as i32),
        (Value::TimeTz(t, z), oid::TIMETZ) => Value::TimeTz(round_micros(t, to.modifier), z),
        (Value::TimestampTz(ts), oid::TIMETZ) => {
            let offset = zone_offset(ts);
            Value::TimeTz((ts + offset).rem_euclid(USECS_PER_DAY), -(offset / USECS_PER_SEC) as i32)
        }
        (Value::Interval(iv), oid::INTERVAL) => {
            Value::Interval(dt::Interval { micros: round_micros(iv.micros, to.modifier & 0xffff), ..iv })
        }
        (Value::Time(t), oid::INTERVAL) => Value::Interval(dt::Interval { months: 0, days: 0, micros: t }),
        (other, _) => return Err(cannot_cast(&other, to.oid)),
    };
    Ok(result)
}

/// cast_value converts a value to the type, as an explicit cast or as an implicit or assignment conversion.
pub fn cast_value(value: Value, to: ColumnType, explicit: bool) -> Result<Value> {
    if value.is_null() {
        return Ok(Value::Null);
    }
    if crate::array::is_array_type(to.oid) {
        let element = builtin_type(to.oid).map_or(oid::TEXT, |t| t.elem);
        let element_type = ColumnType { oid: element, modifier: to.modifier };
        return match value {
            Value::Array(array) => {
                let values = array
                    .values
                    .into_iter()
                    .map(|v| cast_value(v, element_type, explicit))
                    .collect::<Result<Vec<_>>>()?;
                Ok(Value::Array(Box::new(crate::array::Array { element, dims: array.dims, values })))
            }
            Value::Text(text) => {
                let parsed = crate::array::parse(&text, element, &|item| {
                    input(item, element).and_then(|v| cast_value(v, element_type, explicit))
                })?;
                Ok(Value::Array(Box::new(parsed)))
            }
            other => Err(cannot_cast(&other, to.oid)),
        };
    }
    if let Value::Jsonb(json) = &value
        && matches!(to.oid, oid::INT2 | oid::INT4 | oid::INT8 | oid::FLOAT4 | oid::FLOAT8 | oid::NUMERIC | oid::BOOL)
    {
        return jsonb_scalar(json, to);
    }
    Ok(match to.oid {
        oid::JSON => match value {
            Value::Json(text) => Value::Json(text),
            Value::Jsonb(json) => Value::Json(json.to_text()),
            Value::Text(text) => input(&text, to.oid)?,
            other => return Err(cannot_cast(&other, to.oid)),
        },
        oid::JSONB => match value {
            Value::Jsonb(json) => Value::Jsonb(json),
            Value::Json(text) | Value::Text(text) => input(&text, to.oid)?,
            other => return Err(cannot_cast(&other, to.oid)),
        },
        oid::INT2 | oid::INT4 | oid::INT8 => to_integer(value, to.oid)?,
        oid::FLOAT4 | oid::FLOAT8 => {
            let f = match value {
                Value::Int2(i) => i as f64,
                Value::Int4(i) => i as f64,
                Value::Int8(i) => i as f64,
                Value::Float4(f) => f as f64,
                Value::Float8(f) => f,
                Value::Numeric(n) => n.to_f64(),
                Value::Text(text) => return input(&text, to.oid),
                other => return Err(cannot_cast(&other, to.oid)),
            };
            if to.oid == oid::FLOAT4 {
                if f.is_finite() && (f as f32).is_infinite() {
                    return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value out of range: overflow"));
                }
                Value::Float4(f as f32)
            } else {
                Value::Float8(f)
            }
        }
        oid::NUMERIC => {
            let n = match value {
                Value::Int2(i) => Numeric::from_i64(i as i64),
                Value::Int4(i) => Numeric::from_i64(i as i64),
                Value::Int8(i) => Numeric::from_i64(i),
                Value::Float4(f) => {
                    Numeric::from_f64(Value::Float4(f).output().unwrap_or_default().parse().unwrap_or(f64::NAN))
                }
                Value::Float8(f) => Numeric::from_f64(f),
                Value::Numeric(n) => n,
                Value::Text(text) => Numeric::parse(&text)?,
                other => return Err(cannot_cast(&other, to.oid)),
            };
            Value::Numeric(n.apply_typmod(to.modifier)?)
        }
        oid::DATE | oid::TIME | oid::TIMETZ | oid::TIMESTAMP | oid::TIMESTAMPTZ | oid::INTERVAL => {
            cast_datetime(value, to)?
        }
        oid::BOOL => match value {
            Value::Bool(b) => Value::Bool(b),
            Value::Int4(i) => Value::Bool(i != 0),
            Value::Text(text) => input(&text, to.oid)?,
            other => return Err(cannot_cast(&other, to.oid)),
        },
        oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME | oid::UNKNOWN => {
            let text = match value {
                Value::Bool(b) => if b { "true" } else { "false" }.to_string(),
                Value::Text(text) if to.oid == oid::BPCHAR => text,
                Value::Text(text) => text,
                other => other.output().unwrap_or_default(),
            };
            Value::Text(apply_length(text, to, explicit)?)
        }
        oid::CHAR => match value {
            Value::Text(text) if is_char_value(&text) => Value::Text(text),
            other => char_value(&other.output().unwrap_or_default()),
        },
        oid::OID | oid::XID | oid::CID => match value {
            Value::Oid(o) => Value::Oid(o),
            Value::Reg(reg) => Value::Oid(reg.oid),
            Value::Int2(i) => Value::Oid(i as u32),
            Value::Int4(i) => Value::Oid(i as u32),
            Value::Int8(i) => Value::Oid(
                u32::try_from(i)
                    .or_else(|_| i32::try_from(i).map(|i| i as u32))
                    .map_err(|_| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "OID out of range"))?,
            ),
            Value::Text(text) => input(&text, to.oid)?,
            other => return Err(cannot_cast(&other, to.oid)),
        },
        _ => return Err(PgError::unsupported(format!("casts to {}", type_display(to.oid)))),
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn integers_read_as_postgres_reads_them() {
        assert_eq!(input(" 42 ", oid::INT4).unwrap(), Value::Int4(42));
        assert_eq!(input("-0x1F", oid::INT8).unwrap(), Value::Int8(-31));
        assert_eq!(input("1_000", oid::INT2).unwrap(), Value::Int2(1000));
        assert_eq!(input("-32768", oid::INT2).unwrap(), Value::Int2(i16::MIN));
        let err = input("32768", oid::INT2).unwrap_err();
        assert_eq!((err.code, err.message.as_str()), ("22003", "value \"32768\" is out of range for type smallint"));
        let err = input("1.5", oid::INT4).unwrap_err();
        assert_eq!((err.code, err.message.as_str()), ("22P02", "invalid input syntax for type integer: \"1.5\""));
    }

    #[test]
    fn booleans_and_floats_read_as_postgres_reads_them() {
        assert_eq!(input("tr", oid::BOOL).unwrap(), Value::Bool(true));
        assert_eq!(input("OFF", oid::BOOL).unwrap(), Value::Bool(false));
        assert!(input("o", oid::BOOL).is_err());
        assert_eq!(input("-Infinity", oid::FLOAT8).unwrap(), Value::Float8(f64::NEG_INFINITY));
        assert_eq!(input("1e3", oid::FLOAT4).unwrap(), Value::Float4(1000.0));
        assert_eq!(
            input("1e400", oid::FLOAT8).unwrap_err().message,
            "\"1e400\" is out of range for type double precision"
        );
    }
}
