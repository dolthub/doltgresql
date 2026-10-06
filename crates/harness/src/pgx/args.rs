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

use crate::pgx::formats::{BINARY, TEXT, format_code_for_oid};

/// Arg is a query argument, mirroring the Go value that a test would pass to pgx. Each variant is encoded the way
/// pgx encodes the corresponding Go type.
#[derive(Clone, Debug, PartialEq)]
pub enum Arg {
    /// An untyped Go nil.
    Null,
    /// A Go int.
    Int(i64),
    /// A Go int32.
    Int32(i32),
    /// A Go int64.
    Int64(i64),
    /// A Go float64.
    Float64(f64),
    /// A Go float32.
    Float32(f32),
    /// A Go uint64.
    Uint64(u64),
    /// A Go bool.
    Bool(bool),
    /// A Go string.
    Str(String),
    /// A Go []byte.
    Bytes(Vec<u8>),
    /// A Go time.Time.
    Time(Time),
    /// A pgtype.Date built from a time.
    Date(Time),
    /// A pgtype.Timestamp built from a time.
    Timestamp(Time),
    /// A pgtype.Numeric built by scanning the given text.
    Numeric(String),
    /// A pgtype.UUID.
    Uuid([u8; 16]),
    /// A Go []string.
    StrArray(Vec<String>),
    /// A Go []int32.
    Int32Array(Vec<i32>),
}

/// Time is a Go time.Time: a civil date and time with a UTC offset.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Time {
    /// The year, which may be zero or negative.
    pub year: i32,
    /// The month, from 1 to 12.
    pub month: u32,
    /// The day of the month, from 1.
    pub day: u32,
    /// The hour, from 0 to 23.
    pub hour: u32,
    /// The minute, from 0 to 59.
    pub minute: u32,
    /// The second, from 0 to 59.
    pub second: u32,
    /// The nanosecond within the second.
    pub nanosecond: u32,
    /// The offset from UTC in seconds, where zero is UTC.
    pub offset_seconds: i32,
}

impl Time {
    /// utc returns the time with the given civil date and time in UTC.
    pub fn utc(year: i32, month: u32, day: u32, hour: u32, minute: u32, second: u32, nanosecond: u32) -> Time {
        Time { year, month, day, hour, minute, second, nanosecond, offset_seconds: 0 }
    }

    /// days_since_2000 returns the number of days from 2000-01-01 to this civil date.
    fn days_since_2000(&self) -> i64 {
        days_from_civil(self.year as i64, self.month as i64, self.day as i64) - days_from_civil(2000, 1, 1)
    }

    /// civil_microseconds_since_2000 returns the microseconds from 2000-01-01 00:00:00 to this civil time, ignoring
    /// the offset.
    fn civil_microseconds_since_2000(&self) -> i64 {
        let seconds =
            self.days_since_2000() * 86_400 + self.hour as i64 * 3600 + self.minute as i64 * 60 + self.second as i64;
        seconds * 1_000_000 + (self.nanosecond / 1000) as i64
    }

    /// microseconds_of_day returns the microseconds since midnight, which is how pgx encodes a time of day.
    fn microseconds_of_day(&self) -> i64 {
        (self.hour as i64 * 3600 + self.minute as i64 * 60 + self.second as i64) * 1_000_000
            + (self.nanosecond / 1000) as i64
    }

    /// utc_microseconds_since_2000 returns the microseconds from 2000-01-01 00:00:00 UTC to this instant.
    fn utc_microseconds_since_2000(&self) -> i64 {
        self.civil_microseconds_since_2000() - self.offset_seconds as i64 * 1_000_000
    }
}

/// days_from_civil returns the days since 1970-01-01 of a proleptic Gregorian date.
fn days_from_civil(year: i64, month: i64, day: i64) -> i64 {
    let year = if month <= 2 { year - 1 } else { year };
    let era = year.div_euclid(400);
    let year_of_era = year - era * 400;
    let month_index = if month > 2 { month - 3 } else { month + 9 };
    let day_of_year = (153 * month_index + 2) / 5 + day - 1;
    let day_of_era = year_of_era * 365 + year_of_era / 4 - year_of_era / 100 + day_of_year;
    era * 146_097 + day_of_era - 719_468
}

/// EncodedArg is an argument encoded for a Bind message.
#[derive(Clone, Debug, PartialEq, Eq)]
pub(crate) struct EncodedArg {
    /// The format code.
    pub(crate) format: i16,
    /// The encoded value, where None is NULL.
    pub(crate) value: Option<Vec<u8>>,
}

/// encode_arg encodes an argument for a parameter of the given type, choosing the format like pgx's
/// ExtendedQueryBuilder: the preferred format first, then the other format when the preferred one cannot encode it.
pub(crate) fn encode_arg(oid: u32, arg: &Arg) -> Result<EncodedArg, String> {
    let preferred = match arg {
        Arg::Str(_) => TEXT,
        _ => format_code_for_oid(oid),
    };
    match encode_with_format(oid, preferred, arg) {
        Ok(value) => Ok(EncodedArg { format: preferred, value }),
        Err(preferred_err) => {
            let other = if preferred == TEXT { BINARY } else { TEXT };
            match encode_with_format(oid, other, arg) {
                Ok(value) => Ok(EncodedArg { format: other, value }),
                Err(_) => Err(preferred_err),
            }
        }
    }
}

/// encode_with_format encodes an argument in the given format, or returns an error when pgx has no plan for it.
fn encode_with_format(oid: u32, format: i16, arg: &Arg) -> Result<Option<Vec<u8>>, String> {
    let unsupported = || Err(format!("unable to encode {arg:?} into format {format} for OID {oid}"));
    match arg {
        Arg::Null => Ok(None),
        Arg::Str(value) => {
            if format == TEXT {
                Ok(Some(value.as_bytes().to_vec()))
            } else {
                unsupported()
            }
        }
        Arg::Int(value) | Arg::Int64(value) => {
            encode_integer(oid, format, *value).map(Some).ok_or(()).or_else(|_| unsupported())
        }
        Arg::Int32(value) => encode_integer(oid, format, *value as i64).map(Some).ok_or(()).or_else(|_| unsupported()),
        Arg::Float64(value) => match (oid, format) {
            (701, BINARY) => Ok(Some(value.to_bits().to_be_bytes().to_vec())),
            (700, BINARY) => Ok(Some((*value as f32).to_bits().to_be_bytes().to_vec())),
            (1700, BINARY) => Ok(Some(encode_numeric(&format_go_float(*value))?)),
            (20 | 21 | 23 | 26, _) => {
                if value.fract() != 0.0 || !value.is_finite() || *value < i64::MIN as f64 || *value >= i64::MAX as f64 {
                    return Err(format!("cannot convert {} to int64", format_go_float(*value)));
                }
                encode_integer(oid, format, *value as i64).map(Some).ok_or(()).or_else(|_| unsupported())
            }
            _ => unsupported(),
        },
        Arg::Float32(value) => match (oid, format) {
            (700, BINARY) => Ok(Some(value.to_bits().to_be_bytes().to_vec())),
            _ => encode_with_format(oid, format, &Arg::Float64(*value as f64)),
        },
        Arg::Uint64(value) => match (oid, format) {
            (1700, BINARY) => Ok(Some(encode_numeric(&value.to_string())?)),
            _ => match i64::try_from(*value) {
                Ok(value) => encode_with_format(oid, format, &Arg::Int64(value)),
                Err(_) => Err(format!("{value} is greater than maximum value for int64")),
            },
        },
        Arg::Bool(value) => match (oid, format) {
            (16, BINARY) => Ok(Some(vec![*value as u8])),
            _ => unsupported(),
        },
        Arg::Bytes(value) => match (oid, format) {
            (17, BINARY) => Ok(Some(value.clone())),
            _ => unsupported(),
        },
        Arg::Time(time) => match (oid, format) {
            (1184, BINARY) => Ok(Some(time.utc_microseconds_since_2000().to_be_bytes().to_vec())),
            (1114, BINARY) => Ok(Some(time.civil_microseconds_since_2000().to_be_bytes().to_vec())),
            (1082, BINARY) => Ok(Some((time.days_since_2000() as i32).to_be_bytes().to_vec())),
            (1083, BINARY) => Ok(Some(time.microseconds_of_day().to_be_bytes().to_vec())),
            _ => unsupported(),
        },
        Arg::Date(time) => match (oid, format) {
            (1082, BINARY) => Ok(Some((time.days_since_2000() as i32).to_be_bytes().to_vec())),
            _ => unsupported(),
        },
        Arg::Timestamp(time) => match (oid, format) {
            (1114, BINARY) => Ok(Some(time.civil_microseconds_since_2000().to_be_bytes().to_vec())),
            (1184, BINARY) => Ok(Some(time.civil_microseconds_since_2000().to_be_bytes().to_vec())),
            _ => unsupported(),
        },
        Arg::Numeric(text) => match (oid, format) {
            (1700, BINARY) => Ok(Some(encode_numeric(text)?)),
            _ => unsupported(),
        },
        Arg::Uuid(value) => match (oid, format) {
            (2950, BINARY) => Ok(Some(value.to_vec())),
            _ => unsupported(),
        },
        Arg::StrArray(values) => match (oid, format) {
            (1009, BINARY) => Ok(Some(encode_array(25, values.iter().map(|v| v.as_bytes().to_vec()).collect()))),
            (1015, BINARY) => Ok(Some(encode_array(1043, values.iter().map(|v| v.as_bytes().to_vec()).collect()))),
            _ => unsupported(),
        },
        Arg::Int32Array(values) => match (oid, format) {
            (1007, BINARY) => Ok(Some(encode_array(23, values.iter().map(|v| v.to_be_bytes().to_vec()).collect()))),
            (1016, BINARY) => {
                Ok(Some(encode_array(20, values.iter().map(|v| (*v as i64).to_be_bytes().to_vec()).collect())))
            }
            _ => unsupported(),
        },
    }
}

/// encode_integer encodes a Go integer for an integer-like parameter, or returns None when pgx has no plan or the
/// value is out of range.
fn encode_integer(oid: u32, format: i16, value: i64) -> Option<Vec<u8>> {
    match (oid, format) {
        (20, BINARY) => Some(value.to_be_bytes().to_vec()),
        (23, BINARY) => i32::try_from(value).ok().map(|v| v.to_be_bytes().to_vec()),
        (21, BINARY) => i16::try_from(value).ok().map(|v| v.to_be_bytes().to_vec()),
        (26, BINARY) => u32::try_from(value).ok().map(|v| v.to_be_bytes().to_vec()),
        (700, BINARY) => Some((value as f32).to_bits().to_be_bytes().to_vec()),
        (701, BINARY) => Some((value as f64).to_bits().to_be_bytes().to_vec()),
        (1700, BINARY) => encode_numeric(&value.to_string()).ok(),
        _ => None,
    }
}

/// encode_array encodes a one-dimensional array without NULLs in the binary format.
fn encode_array(element_oid: u32, elements: Vec<Vec<u8>>) -> Vec<u8> {
    let mut buffer = Vec::new();
    if elements.is_empty() {
        buffer.extend_from_slice(&0i32.to_be_bytes());
        buffer.extend_from_slice(&0i32.to_be_bytes());
        buffer.extend_from_slice(&element_oid.to_be_bytes());
        return buffer;
    }
    buffer.extend_from_slice(&1i32.to_be_bytes());
    buffer.extend_from_slice(&0i32.to_be_bytes());
    buffer.extend_from_slice(&element_oid.to_be_bytes());
    buffer.extend_from_slice(&(elements.len() as i32).to_be_bytes());
    buffer.extend_from_slice(&1i32.to_be_bytes());
    for element in elements {
        buffer.extend_from_slice(&(element.len() as i32).to_be_bytes());
        buffer.extend_from_slice(&element);
    }
    buffer
}

/// format_go_float formats a float the way Go's strconv.FormatFloat(value, 'f', -1, 64) does: the shortest digits
/// that round-trip, without an exponent, which is also how Rust displays a float.
fn format_go_float(value: f64) -> String {
    format!("{value}")
}

/// encode_numeric encodes numeric text in the binary format exactly as pgx encodes a scanned pgtype.Numeric,
/// including its quirks: trailing zeros of a whole number move into the exponent, and every fractional digit group
/// is kept.
pub(crate) fn encode_numeric(text: &str) -> Result<Vec<u8>, String> {
    let mut buffer = Vec::new();
    match text {
        "NaN" => {
            buffer.extend_from_slice(&0x0000_0000_C000_0000u64.to_be_bytes());
            return Ok(buffer);
        }
        "Infinity" => {
            buffer.extend_from_slice(&0x0000_0000_D000_0000u64.to_be_bytes());
            return Ok(buffer);
        }
        "-Infinity" => {
            buffer.extend_from_slice(&0x0000_0000_F000_0000u64.to_be_bytes());
            return Ok(buffer);
        }
        _ => {}
    }
    let (signed_digits, mut exp) = parse_numeric_string(text)?;
    let negative = signed_digits.starts_with('-');
    let digits = signed_digits.trim_start_matches(['-', '+']).trim_start_matches('0').to_string();
    let negative = negative && !digits.is_empty();
    let original_exp = exp;

    let shift = ((exp % 4) + 4) % 4;
    let mut digits = digits;
    if shift != 0 {
        exp -= shift;
        if !digits.is_empty() {
            digits.push_str(&"0".repeat(shift as usize));
        }
    }

    let (whole, fraction) = if exp < 0 {
        let fraction_len = (-exp) as usize;
        let padded = if digits.len() < fraction_len {
            format!("{}{}", "0".repeat(fraction_len - digits.len()), digits)
        } else {
            digits.clone()
        };
        let split = padded.len() - fraction_len;
        (padded[..split].trim_start_matches('0').to_string(), padded[split..].to_string())
    } else {
        (digits.clone(), String::new())
    };

    let whole_groups = group_digits(&whole);
    let fraction_groups: Vec<i16> =
        fraction.as_bytes().chunks(4).map(|chunk| std::str::from_utf8(chunk).unwrap().parse().unwrap()).collect();

    let weight = if !whole_groups.is_empty() {
        let mut weight = whole_groups.len() as i16 - 1;
        if exp > 0 {
            weight += (exp / 4) as i16;
        }
        weight
    } else {
        (exp / 4) as i16 - 1 + fraction_groups.len() as i16
    };
    let dscale = if original_exp < 0 { (-original_exp) as i16 } else { 0 };

    buffer.extend_from_slice(&((whole_groups.len() + fraction_groups.len()) as i16).to_be_bytes());
    buffer.extend_from_slice(&weight.to_be_bytes());
    buffer.extend_from_slice(&(if negative { 16384i16 } else { 0 }).to_be_bytes());
    buffer.extend_from_slice(&dscale.to_be_bytes());
    for group in whole_groups.iter().chain(fraction_groups.iter()) {
        buffer.extend_from_slice(&group.to_be_bytes());
    }
    Ok(buffer)
}

/// parse_numeric_string mirrors pgx's parseNumericString, returning the digits (with any sign) and the exponent.
fn parse_numeric_string(text: &str) -> Result<(String, i32), String> {
    let mut digits = text.to_string();
    let mut exp = 0i32;
    match digits.find('.') {
        None => {
            while digits.len() > 1 && digits.ends_with('0') && !digits[..digits.len() - 1].ends_with('-') {
                digits.pop();
                exp += 1;
            }
        }
        Some(index) => {
            exp = -((digits.len() - index - 1) as i32);
            digits.remove(index);
        }
    }
    let unsigned = digits.strip_prefix(['-', '+']).unwrap_or(&digits);
    if unsigned.is_empty() || !unsigned.bytes().all(|b| b.is_ascii_digit()) {
        return Err(format!("{digits} is not a number"));
    }
    Ok((digits, exp))
}

/// group_digits splits a whole number's digits into base-10000 groups, most significant first.
fn group_digits(digits: &str) -> Vec<i16> {
    let mut groups = Vec::new();
    let bytes = digits.as_bytes();
    let mut end = bytes.len();
    while end > 0 {
        let start = end.saturating_sub(4);
        groups.push(std::str::from_utf8(&bytes[start..end]).unwrap().parse().unwrap());
        end = start;
    }
    groups.reverse();
    groups
}
