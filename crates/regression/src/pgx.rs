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

use crate::gostd::{self, GoTime, TimeZone, format_f6_32, format_f6_64, format_shortest_fixed, format_v32, format_v64};
use crate::json;

/// Any is a value as pgx's DecodeValue returns it, before the replay's normalization.
#[derive(Clone, Debug)]
pub enum Any {
    Nil,
    Bool(bool),
    Int16(i16),
    Int32(i32),
    Int64(i64),
    Uint32(u32),
    Uint64(u64),
    Float32(f32),
    Float64(f64),
    String(String),
    Bytes(Vec<u8>),
    Uuid([u8; 16]),
    Time(GoTime),
    Infinity(i8),
    Numeric(Numeric),
    Interval { micros: i64, days: i32, months: i32 },
    TimeOfDay(i64),
    Bits { bytes: Vec<u8>, len: i32 },
    Prefix(String),
    HardwareAddr(String),
    Point(f64, f64),
    Lseg([f64; 4]),
    Box([f64; 4]),
    Line([f64; 3]),
    Circle([f64; 3]),
    Path { points: Vec<(f64, f64)>, closed: bool },
    Polygon(Vec<(f64, f64)>),
    Tid(u32, u16),
    TsVector(Vec<Lexeme>),
    Range { lower: Box<Any>, upper: Box<Any>, lower_type: u8, upper_type: u8 },
    Multirange,
    Map(Vec<(String, Any)>),
    Slice(Vec<Any>),
}

/// Lexeme is a tsvector word with its positions and weight letters.
pub type Lexeme = (String, Vec<(u16, u8)>);

/// Numeric is a pgtype.Numeric, where a None int is a nil *big.Int.
#[derive(Clone, Debug)]
pub struct Numeric {
    pub int: Option<String>,
    pub exp: i32,
    pub nan: bool,
    pub infinity: i8,
}

/// Value is a cell after the replay's normalization in ReadRows.
#[derive(Clone, Debug)]
pub enum Value {
    Nil,
    Bool(bool),
    Int16(i16),
    Int32(i32),
    Int64(i64),
    Uint32(u32),
    Uint64(u64),
    Float32(f32),
    Float64(f64),
    String(String),
    Time(GoTime),
    Bytes(Vec<u8>),
}

impl Value {
    /// go_eq matches Go's `==` on the interface values, comparing byte slices by content where Go panics.
    pub fn go_eq(&self, other: &Value) -> bool {
        match (self, other) {
            (Value::Nil, Value::Nil) => true,
            (Value::Bool(a), Value::Bool(b)) => a == b,
            (Value::Int16(a), Value::Int16(b)) => a == b,
            (Value::Int32(a), Value::Int32(b)) => a == b,
            (Value::Int64(a), Value::Int64(b)) => a == b,
            (Value::Uint32(a), Value::Uint32(b)) => a == b,
            (Value::Uint64(a), Value::Uint64(b)) => a == b,
            (Value::Float32(a), Value::Float32(b)) => a == b,
            (Value::Float64(a), Value::Float64(b)) => a == b,
            (Value::String(a), Value::String(b)) => a == b,
            (Value::Time(a), Value::Time(b)) => a.go_eq(b),
            (Value::Bytes(a), Value::Bytes(b)) => a == b,
            _ => false,
        }
    }

    /// go_type returns the Go type name of the value, as `%T` prints it.
    pub fn go_type(&self) -> &'static str {
        match self {
            Value::Nil => "<nil>",
            Value::Bool(_) => "bool",
            Value::Int16(_) => "int16",
            Value::Int32(_) => "int32",
            Value::Int64(_) => "int64",
            Value::Uint32(_) => "uint32",
            Value::Uint64(_) => "uint64",
            Value::Float32(_) => "float32",
            Value::Float64(_) => "float64",
            Value::String(_) => "string",
            Value::Time(_) => "time.Time",
            Value::Bytes(_) => "[]uint8",
        }
    }

    /// as_oid matches the replay's cellToOID.
    pub fn as_oid(&self) -> Option<u32> {
        match self {
            Value::Uint32(v) => Some(*v),
            Value::Int64(v) => u32::try_from(*v).ok(),
            Value::Uint64(v) => u32::try_from(*v).ok(),
            Value::Int32(v) => u32::try_from(*v).ok(),
            Value::String(v) => gostd::parse_uint(v, 32).map(|v| v as u32),
            _ => None,
        }
    }
}

/// element_oid returns the element type of an array type that pgx's default type map has a codec for.
fn element_oid(oid: u32) -> Option<u32> {
    Some(match oid {
        1034 => 1033,
        1561 => 1560,
        1000 => 16,
        1020 => 603,
        1014 => 1042,
        1001 => 17,
        1002 => 18,
        1012 => 29,
        651 => 650,
        719 => 718,
        1182 => 1082,
        3913 => 3912,
        1021 => 700,
        1022 => 701,
        1041 => 869,
        1005 => 21,
        1007 => 23,
        3905 => 3904,
        1016 => 20,
        3927 => 3926,
        1187 => 1186,
        199 => 114,
        3807 => 3802,
        4073 => 4072,
        629 => 628,
        1018 => 601,
        1040 => 829,
        1003 => 19,
        1231 => 1700,
        3907 => 3906,
        1028 => 26,
        1019 => 602,
        1017 => 600,
        1027 => 604,
        2287 => 2249,
        1009 => 25,
        1010 => 27,
        3643 => 3614,
        1183 => 1083,
        1115 => 1114,
        1185 => 1184,
        3909 => 3908,
        3911 => 3910,
        2951 => 2950,
        1563 => 1562,
        1015 => 1043,
        1011 => 28,
        271 => 5069,
        143 => 142,
        _ => return None,
    })
}

/// range_element_oid returns the element type of a registered range type.
fn range_element_oid(oid: u32) -> Option<u32> {
    Some(match oid {
        3912 => 1082,
        3904 => 23,
        3926 => 20,
        3906 => 1700,
        3908 => 1114,
        3910 => 1184,
        _ => return None,
    })
}

/// is_multirange reports whether the OID is a registered multirange type.
fn is_multirange(oid: u32) -> bool {
    matches!(oid, 4535 | 4451 | 4536 | 4532 | 4533 | 4534)
}

/// decode matches DecodeValue of pgx's default codec for the OID on a text cell, returning None for an error.
pub fn decode(oid: u32, src: Option<&[u8]>) -> Option<Any> {
    let Some(src) = src else {
        return Some(Any::Nil);
    };
    let text = String::from_utf8_lossy(src);
    let s = text.as_ref();
    Some(match oid {
        16 => Any::Bool(parse_bool(s)?),
        17 => Any::Bytes(decode_hex_bytea(s)?),
        18 => {
            if src.len() > 1 {
                return None;
            }
            Any::Int32(src.first().map_or(0, |b| *b as i32))
        }
        20 => Any::Int64(gostd::parse_int(s, 64)?),
        21 => Any::Int16(gostd::parse_int(s, 16)? as i16),
        23 => Any::Int32(gostd::parse_int(s, 32)? as i32),
        26 | 28 | 29 => Any::Uint32(gostd::parse_uint(s, 32)? as u32),
        5069 => Any::Uint64(gostd::parse_uint(s, 64)?),
        700 => Any::Float32(gostd::parse_float32(s)?),
        701 => Any::Float64(gostd::parse_float64(s)?),
        1700 => Any::Numeric(parse_numeric(s)?),
        25 | 19 | 1042 | 1043 | 705 | 1033 | 4072 | 2249 => Any::String(s.to_string()),
        1082 => parse_date(src)?,
        1114 => parse_timestamp(s, false)?,
        1184 => parse_timestamp(s, true)?,
        1083 => Any::TimeOfDay(parse_time_of_day(s)?),
        1186 => parse_interval(s)?,
        650 | 869 => Any::Prefix(parse_prefix(s)?),
        774 | 829 => Any::HardwareAddr(parse_mac(s)?),
        2950 => Any::Uuid(parse_uuid(src)?),
        1560 | 1562 => {
            let mut bytes = vec![0u8; src.len().div_ceil(8)];
            for (i, b) in src.iter().enumerate() {
                if *b == b'1' {
                    bytes[i / 8] |= 128 >> (i % 8);
                }
            }
            Any::Bits { bytes, len: src.len() as i32 }
        }
        600 => parse_point(src)?,
        601 => parse_lseg(s)?,
        603 => parse_box(s)?,
        628 => parse_line(s)?,
        718 => parse_circle(s)?,
        602 => parse_path(s)?,
        604 => parse_polygon(s)?,
        27 => parse_tid(s)?,
        3614 => Any::TsVector(parse_tsvector(s)?),
        114 | 3802 => json::unmarshal(src)?,
        142 => Any::Bytes(src.to_vec()),
        _ if range_element_oid(oid).is_some() => parse_range(oid, s)?,
        _ if is_multirange(oid) => Any::Multirange,
        _ => match element_oid(oid) {
            Some(element) => decode_array(element, s)?,
            None => Any::String(s.to_string()),
        },
    })
}

/// parse_bool matches pgx's text bool parsing, which accepts any prefix of true, yes, false, no, or off.
fn parse_bool(s: &str) -> Option<bool> {
    let s = s.trim().to_lowercase();
    if s.is_empty() {
        return None;
    }
    if "true".starts_with(&s) || "yes".starts_with(&s) || s == "on" || s == "1" {
        Some(true)
    } else if "false".starts_with(&s) || "no".starts_with(&s) || "off".starts_with(&s) || s == "0" {
        Some(false)
    } else {
        None
    }
}

/// decode_hex matches encoding/hex decoding.
fn decode_hex(s: &[u8]) -> Option<Vec<u8>> {
    if !s.len().is_multiple_of(2) {
        return None;
    }
    let nibble = |c: u8| (c as char).to_digit(16).map(|d| d as u8);
    s.chunks(2).map(|pair| Some(nibble(pair[0])? << 4 | nibble(pair[1])?)).collect()
}

/// decode_hex_bytea matches pgx's bytea text decoding, which only accepts the hex format.
fn decode_hex_bytea(s: &str) -> Option<Vec<u8>> {
    decode_hex(s.strip_prefix("\\x")?.as_bytes())
}

/// parse_numeric matches pgx's text numeric parsing.
fn parse_numeric(s: &str) -> Option<Numeric> {
    match s {
        "NaN" => return Some(Numeric { int: None, exp: 0, nan: true, infinity: 0 }),
        "Infinity" => return Some(Numeric { int: None, exp: 0, nan: false, infinity: 1 }),
        "-Infinity" => return Some(Numeric { int: None, exp: 0, nan: false, infinity: -1 }),
        _ => {}
    }
    let mut digits = s.to_string();
    let mut exp: i32 = 0;
    match s.find('.') {
        None => {
            while digits.len() > 1 && digits.ends_with('0') && digits.as_bytes()[digits.len() - 2] != b'-' {
                digits.pop();
                exp += 1;
            }
        }
        Some(index) => {
            exp = -((s.len() - index - 1) as i32);
            digits.remove(index);
        }
    }
    Some(Numeric { int: Some(big_int_string(&digits)?), exp, nan: false, infinity: 0 })
}

/// big_int_string matches big.Int.SetString(s, 10) followed by String().
fn big_int_string(s: &str) -> Option<String> {
    let (negative, digits) = match s.as_bytes().first() {
        Some(b'-') => (true, &s[1..]),
        Some(b'+') => (false, &s[1..]),
        _ => (false, s),
    };
    if digits.is_empty() || !digits.bytes().all(|b| b.is_ascii_digit()) {
        return None;
    }
    let trimmed = digits.trim_start_matches('0');
    if trimmed.is_empty() {
        return Some("0".to_string());
    }
    Some(if negative { format!("-{trimmed}") } else { trimmed.to_string() })
}

impl Numeric {
    /// text matches Numeric.Value().
    fn text(&self) -> String {
        if self.nan {
            return "NaN".to_string();
        }
        match self.infinity {
            1 => return "Infinity".to_string(),
            -1 => return "-Infinity".to_string(),
            _ => {}
        }
        let Some(int) = &self.int else {
            return "0".to_string();
        };
        let (sign, int) = match int.strip_prefix('-') {
            Some(rest) => ("-", rest),
            None => ("", int.as_str()),
        };
        let exp = self.exp;
        let body = if exp > 0 {
            format!("{int}{}", "0".repeat(exp as usize))
        } else if exp < 0 {
            let places = (-exp) as usize;
            if int.len() <= places {
                format!("0.{}{int}", "0".repeat(places - int.len()))
            } else {
                format!("{}.{}", &int[..int.len() - places], &int[int.len() - places..])
            }
        } else {
            int.to_string()
        };
        format!("{sign}{body}")
    }
}

/// parse_digits matches pgx's date parsing helper, rejecting anything but ASCII digits.
fn parse_digits(b: &[u8]) -> Option<i64> {
    if b.is_empty() {
        return None;
    }
    let mut n: i64 = 0;
    for c in b {
        if !c.is_ascii_digit() {
            return None;
        }
        n = n.wrapping_mul(10).wrapping_add((c - b'0') as i64);
    }
    Some(n)
}

/// parse_date matches pgx's text date parsing, which only accepts the ISO format.
fn parse_date(src: &[u8]) -> Option<Any> {
    match src {
        b"infinity" => return Some(Any::Infinity(1)),
        b"-infinity" => return Some(Any::Infinity(-1)),
        _ => {}
    }
    if src.len() < 10 {
        return None;
    }
    let (date, bc) = match src.strip_suffix(b" BC") {
        Some(date) if src.len() >= 13 => (date, true),
        _ => (src, false),
    };
    let mut year_end = None;
    for (i, c) in date.iter().enumerate().skip(4) {
        if *c == b'-' {
            year_end = Some(i);
            break;
        }
        if !c.is_ascii_digit() {
            return None;
        }
    }
    let year_end = year_end?;
    if year_end + 6 > date.len() || date[year_end + 3] != b'-' {
        return None;
    }
    let two = |b: &[u8]| {
        (b.len() == 2 && b.iter().all(u8::is_ascii_digit)).then(|| ((b[0] - b'0') * 10 + (b[1] - b'0')) as i64)
    };
    let mut year = parse_digits(&date[..year_end])?;
    let month = two(&date[year_end + 1..year_end + 3])?;
    let day = two(&date[year_end + 4..year_end + 6])?;
    if year_end + 6 != date.len() {
        return None;
    }
    if bc {
        year = -year + 1;
    }
    Some(Any::Time(go_date_any_month(year, month, day, 0, 0, 0, 0, gostd::Location::Utc)))
}

/// go_date_any_month matches time.Date, normalizing a month outside of 1 to 12 into the year.
#[allow(clippy::too_many_arguments)]
fn go_date_any_month(
    year: i64,
    month: i64,
    day: i64,
    hour: i64,
    min: i64,
    sec: i64,
    nanos: i64,
    location: gostd::Location,
) -> GoTime {
    let month0 = month - 1;
    let year = year + month0.div_euclid(12);
    let month = month0.rem_euclid(12) + 1;
    gostd::go_date(year, month, day, hour, min, sec, nanos, location)
}

/// parse_timestamp matches pgx's text timestamp and timestamptz parsing.
fn parse_timestamp(s: &str, with_zone: bool) -> Option<Any> {
    match s {
        "infinity" => return Some(Any::Infinity(1)),
        "-infinity" => return Some(Any::Infinity(-1)),
        _ => {}
    }
    let (s, bc) = match s.strip_suffix(" BC") {
        Some(rest) => (rest, true),
        None => (s, false),
    };
    let zone = if !with_zone {
        TimeZone::None
    } else {
        let b = s.as_bytes();
        if b.len() >= 9 && matches!(b[b.len() - 9], b'-' | b'+') {
            TimeZone::Second
        } else if b.len() >= 6 && matches!(b[b.len() - 6], b'-' | b'+') {
            TimeZone::Minute
        } else {
            TimeZone::Hour
        }
    };
    let mut time = gostd::parse_time(s, zone)?;
    if bc {
        let offset = match time.location {
            gostd::Location::Utc => 0,
            gostd::Location::Fixed(offset) => offset as i64,
        };
        let local = time.unix + offset;
        let (year, month, day) = gostd::civil_from_days(local.div_euclid(86400));
        let seconds = local.rem_euclid(86400);
        time = gostd::go_date(
            -year + 1,
            month,
            day,
            seconds / 3600,
            seconds / 60 % 60,
            seconds % 60,
            time.nanos as i64,
            time.location,
        );
    }
    Some(Any::Time(time))
}

/// parse_time_of_day matches pgx's text time parsing.
fn parse_time_of_day(s: &str) -> Option<i64> {
    let b = s.as_bytes();
    if b.len() < 8 || b[2] != b':' || b[5] != b':' {
        return None;
    }
    let int = |part: &[u8]| gostd::parse_int(std::str::from_utf8(part).ok()?, 64);
    let mut usec = int(&b[0..2])? * 3_600_000_000;
    usec += int(&b[3..5])? * 60_000_000;
    usec += int(&b[6..8])? * 1_000_000;
    if b.len() > 9 {
        if b[8] != b'.' || b.len() > 15 {
            return None;
        }
        let fraction = &b[9..];
        let mut n = int(fraction)?;
        for _ in fraction.len()..6 {
            n *= 10;
        }
        usec += n;
    }
    Some(usec)
}

/// parse_interval matches pgx's text interval parsing, which only reads the postgres interval style.
fn parse_interval(s: &str) -> Option<Any> {
    let parts: Vec<&str> = s.split(' ').collect();
    let mut months: i32 = 0;
    let mut days: i32 = 0;
    let mut micros: i64 = 0;
    let mut i = 0;
    while i + 1 < parts.len() {
        let scalar = gostd::parse_int(parts[i], 64)?;
        match parts[i + 1] {
            "year" | "years" => months = months.wrapping_add(scalar.wrapping_mul(12) as i32),
            "mon" | "mons" => months = months.wrapping_add(scalar as i32),
            "day" | "days" => days = scalar as i32,
            _ => return None,
        }
        i += 2;
    }
    if parts.len() % 2 == 1 {
        let time_parts: Vec<&str> = parts[parts.len() - 1].splitn(3, ':').collect();
        if time_parts.len() != 3 {
            return None;
        }
        let mut hours_text = time_parts[0];
        let negative = hours_text.as_bytes().first()? == &b'-';
        if negative {
            hours_text = &hours_text[1..];
        }
        let hours = gostd::parse_int(hours_text, 64)?;
        let minutes = gostd::parse_int(time_parts[1], 64)?;
        let (sec, fraction) = match time_parts[2].split_once('.') {
            Some((sec, fraction)) => (sec, Some(fraction)),
            None => (time_parts[2], None),
        };
        let seconds = gostd::parse_int(sec, 64)?;
        let mut usec = 0i64;
        if let Some(fraction) = fraction {
            usec = gostd::parse_int(fraction, 64)?;
            for _ in fraction.len()..6 {
                usec = usec.wrapping_mul(10);
            }
        }
        micros = hours
            .wrapping_mul(3_600_000_000)
            .wrapping_add(minutes.wrapping_mul(60_000_000))
            .wrapping_add(seconds.wrapping_mul(1_000_000))
            .wrapping_add(usec);
        if negative {
            micros = micros.wrapping_neg();
        }
    }
    Some(Any::Interval { micros, days, months })
}

/// interval_text matches pgx's interval text encoding.
fn interval_text(micros: i64, days: i32, months: i32) -> String {
    let mut out = String::new();
    if months != 0 {
        out.push_str(&format!("{months} mon "));
    }
    if days != 0 {
        out.push_str(&format!("{days} day "));
    }
    let mut abs = micros;
    if abs < 0 {
        abs = abs.wrapping_neg();
        out.push('-');
    }
    out.push_str(&format!(
        "{:02}:{:02}:{:02}",
        abs / 3_600_000_000,
        abs % 3_600_000_000 / 60_000_000,
        abs % 60_000_000 / 1_000_000
    ));
    if abs % 1_000_000 != 0 {
        out.push_str(&format!(".{:06}", abs % 1_000_000));
    }
    out
}

/// parse_prefix matches pgx's text inet and cidr parsing into a netip.Prefix, returning its String() form.
fn parse_prefix(s: &str) -> Option<String> {
    use std::net::IpAddr;
    let (addr, bits) = match s.split_once('/') {
        None => (s, None),
        Some((addr, bits)) => (addr, Some(bits)),
    };
    let (addr_text, zone) = match addr.split_once('%') {
        Some((addr, zone)) if bits.is_none() && !zone.is_empty() => (addr, Some(zone)),
        Some(_) => return None,
        None => (addr, None),
    };
    let ip: IpAddr = addr_text.parse().ok()?;
    if zone.is_some() && ip.is_ipv4() {
        return None;
    }
    let bit_len = if ip.is_ipv4() { 32 } else { 128 };
    let bits = match bits {
        None => bit_len,
        Some(bits) => {
            if bits.is_empty() || !bits.bytes().all(|b| b.is_ascii_digit()) || (bits.len() > 1 && bits.starts_with('0'))
            {
                return None;
            }
            let bits: u32 = bits.parse().ok()?;
            if bits > bit_len {
                return None;
            }
            bits
        }
    };
    let addr_string = match (ip, zone) {
        (IpAddr::V6(v6), _) if v6.to_ipv4_mapped().is_some() => {
            let v4 = v6.to_ipv4_mapped().unwrap();
            format!("::ffff:{v4}")
        }
        (ip, Some(zone)) => format!("{ip}%{zone}"),
        (ip, None) => ip.to_string(),
    };
    Some(format!("{addr_string}/{bits}"))
}

/// parse_mac matches net.ParseMAC followed by HardwareAddr.String().
fn parse_mac(s: &str) -> Option<String> {
    let b = s.as_bytes();
    if b.len() < 14 {
        return None;
    }
    let hex = |pair: &[u8]| decode_hex(pair).map(|v| v[0]);
    let mut out = Vec::new();
    if b[2] == b':' || b[2] == b'-' {
        if !(b.len() + 1).is_multiple_of(3) {
            return None;
        }
        let n = (b.len() + 1) / 3;
        if n != 6 && n != 8 && n != 20 {
            return None;
        }
        for i in 0..n {
            let x = i * 3;
            if i + 1 < n && b[x + 2] != b[2] {
                return None;
            }
            out.push(hex(&b[x..x + 2])?);
        }
    } else if b[4] == b'.' {
        if !(b.len() + 1).is_multiple_of(5) {
            return None;
        }
        let n = 2 * (b.len() + 1) / 5;
        if n != 6 && n != 8 && n != 20 {
            return None;
        }
        for i in (0..n).step_by(2) {
            let x = i / 2 * 5;
            if i + 2 < n && b[x + 4] != b'.' {
                return None;
            }
            out.push(hex(&b[x..x + 2])?);
            out.push(hex(&b[x + 2..x + 4])?);
        }
    } else {
        return None;
    }
    Some(out.iter().map(|b| format!("{b:02x}")).collect::<Vec<_>>().join(":"))
}

/// parse_uuid matches pgx's text uuid parsing, which does not check the dash positions.
fn parse_uuid(src: &[u8]) -> Option<[u8; 16]> {
    let hex: Vec<u8> = match src.len() {
        36 => [&src[0..8], &src[9..13], &src[14..18], &src[19..23], &src[24..]].concat(),
        32 => src.to_vec(),
        _ => return None,
    };
    decode_hex(&hex)?.try_into().ok()
}

/// float matches strconv.ParseFloat(s, 64) on a piece of a geometric value.
fn float(s: &str) -> Option<f64> {
    gostd::parse_float64(s)
}

/// parse_point matches pgx's text point parsing.
fn parse_point(src: &[u8]) -> Option<Any> {
    if src == b"null" {
        return Some(Any::Nil);
    }
    if src.len() < 5 {
        return None;
    }
    let src = if src[0] == b'"' && src[src.len() - 1] == b'"' { &src[1..src.len() - 1] } else { src };
    if src.len() < 2 {
        return None;
    }
    let inner = String::from_utf8_lossy(&src[1..src.len() - 1]).into_owned();
    let (x, y) = inner.split_once(',')?;
    Some(Any::Point(float(x)?, float(y)?))
}

/// split_at_byte returns the text before and after the first `byte`, where Go would panic when it is missing.
fn split_at_byte(s: &str, byte: char) -> Option<(&str, &str)> {
    let index = s.find(byte)?;
    Some((&s[..index], &s[index + 1..]))
}

/// parse_lseg matches pgx's text lseg parsing.
fn parse_lseg(s: &str) -> Option<Any> {
    if s.len() < 11 {
        return None;
    }
    let str = s.get(2..)?;
    let (x1, str) = split_at_byte(str, ',')?;
    let (y1, str) = split_at_byte(str, ')')?;
    let str = str.get(1..)?;
    let (x2, str) = split_at_byte(str, ',')?;
    let y2 = str.get(..str.len().checked_sub(2)?)?;
    Some(Any::Lseg([float(x1)?, float(y1)?, float(x2)?, float(y2)?]))
}

/// parse_box matches pgx's text box parsing.
fn parse_box(s: &str) -> Option<Any> {
    if s.len() < 11 {
        return None;
    }
    let str = s.get(1..)?;
    let (x1, str) = split_at_byte(str, ',')?;
    let (y1, str) = split_at_byte(str, ')')?;
    let str = str.get(2..)?;
    let (x2, str) = split_at_byte(str, ',')?;
    let y2 = str.get(..str.len().checked_sub(1)?)?;
    Some(Any::Box([float(x1)?, float(y1)?, float(x2)?, float(y2)?]))
}

/// parse_line matches pgx's text line parsing.
fn parse_line(s: &str) -> Option<Any> {
    if s.len() < 7 {
        return None;
    }
    let parts: Vec<&str> = s.get(1..s.len() - 1)?.splitn(3, ',').collect();
    if parts.len() < 3 {
        return None;
    }
    Some(Any::Line([float(parts[0])?, float(parts[1])?, float(parts[2])?]))
}

/// parse_circle matches pgx's text circle parsing.
fn parse_circle(s: &str) -> Option<Any> {
    if s.len() < 9 {
        return None;
    }
    let str = s.get(2..)?;
    let (x, str) = split_at_byte(str, ',')?;
    let (y, str) = split_at_byte(str, ')')?;
    let r = str.get(1..str.len().checked_sub(1)?)?;
    Some(Any::Circle([float(x)?, float(y)?, float(r)?]))
}

/// parse_points matches the point loop shared by pgx's text path and polygon parsing.
fn parse_points(s: &str) -> Option<Vec<(f64, f64)>> {
    let mut str = s.get(2..)?;
    let mut points = Vec::new();
    loop {
        let (x, rest) = split_at_byte(str, ',')?;
        let end = rest.find(')')?;
        points.push((float(x)?, float(&rest[..end])?));
        if end + 3 < rest.len() {
            str = rest.get(end + 3..)?;
        } else {
            break;
        }
    }
    Some(points)
}

/// parse_path matches pgx's text path parsing.
fn parse_path(s: &str) -> Option<Any> {
    if s.len() < 7 {
        return None;
    }
    Some(Any::Path { points: parse_points(s)?, closed: s.starts_with('(') })
}

/// parse_polygon matches pgx's text polygon parsing.
fn parse_polygon(s: &str) -> Option<Any> {
    if s.len() < 7 {
        return None;
    }
    Some(Any::Polygon(parse_points(s)?))
}

/// parse_tid matches pgx's text tid parsing.
fn parse_tid(s: &str) -> Option<Any> {
    if s.len() < 5 {
        return None;
    }
    let (block, offset) = s.get(1..s.len() - 1)?.split_once(',')?;
    Some(Any::Tid(gostd::parse_uint(block, 32)? as u32, gostd::parse_uint(offset, 16)? as u16))
}

/// parse_tsvector matches pgx's text tsvector parsing.
fn parse_tsvector(s: &str) -> Option<Vec<Lexeme>> {
    let b = s.trim().as_bytes();
    let mut pos = 0;
    let mut lexemes = Vec::new();
    while pos < b.len() {
        while pos < b.len() && b[pos] == b' ' {
            pos += 1;
        }
        if pos >= b.len() {
            break;
        }
        if b[pos] != b'\'' {
            return None;
        }
        pos += 1;
        let mut word = Vec::new();
        loop {
            let ch = *b.get(pos)?;
            pos += 1;
            match ch {
                b'\'' if b.get(pos) == Some(&b'\'') => {
                    pos += 1;
                    word.push(b'\'');
                }
                b'\'' => break,
                b'\\' => {
                    word.push(*b.get(pos)?);
                    pos += 1;
                }
                _ => word.push(ch),
            }
        }
        let mut positions = Vec::new();
        if b.get(pos) == Some(&b':') {
            pos += 1;
            loop {
                let start = pos;
                while pos < b.len() && b[pos].is_ascii_digit() {
                    pos += 1;
                }
                if pos == start {
                    return None;
                }
                let number = gostd::parse_uint(std::str::from_utf8(&b[start..pos]).ok()?, 16)? as u16;
                let mut weight = b'D';
                if let Some(c) = b.get(pos) {
                    let known = match c {
                        b'A' | b'a' => Some(b'A'),
                        b'B' | b'b' => Some(b'B'),
                        b'C' | b'c' => Some(b'C'),
                        b'D' | b'd' => Some(b'D'),
                        _ => None,
                    };
                    if let Some(known) = known {
                        weight = known;
                        pos += 1;
                    }
                }
                positions.push((number, weight));
                if b.get(pos) != Some(&b',') {
                    break;
                }
                pos += 1;
            }
        }
        lexemes.push((String::from_utf8_lossy(&word).into_owned(), positions));
    }
    Some(lexemes)
}

/// tsvector_text matches pgx's tsvector text encoding.
fn tsvector_text(lexemes: &[Lexeme]) -> String {
    let mut out = String::new();
    for (i, (word, positions)) in lexemes.iter().enumerate() {
        if i > 0 {
            out.push(' ');
        }
        out.push('\'');
        out.push_str(&word.replace('\\', "\\\\").replace('\'', "\\'"));
        out.push('\'');
        let mut separator = ':';
        for (position, weight) in positions {
            out.push(separator);
            out.push_str(&position_string(*position, *weight));
            separator = ',';
        }
    }
    out
}

/// position_string matches TSVectorPosition.String().
fn position_string(position: u16, weight: u8) -> String {
    if weight != 0 && weight != b'D' { format!("{position}{}", weight as char) } else { position.to_string() }
}

/// parse_range matches pgx's text range parsing for a registered range type.
fn parse_range(oid: u32, s: &str) -> Option<Any> {
    let element = range_element_oid(oid).unwrap();
    let range = parse_untyped_range(s)?;
    let bound = |text: &Option<String>| -> Option<Any> {
        match text {
            None => Some(Any::Nil),
            Some(text) => decode(element, Some(text.as_bytes())),
        }
    };
    Some(Any::Range {
        lower: Box::new(bound(&range.lower)?),
        upper: Box::new(bound(&range.upper)?),
        lower_type: range.lower_type,
        upper_type: range.upper_type,
    })
}

/// UntypedRange is a range's bounds as text, where None is a missing bound.
struct UntypedRange {
    lower: Option<String>,
    upper: Option<String>,
    lower_type: u8,
    upper_type: u8,
}

/// parse_untyped_range matches pgx's parseUntypedTextRange.
fn parse_untyped_range(src: &str) -> Option<UntypedRange> {
    if src == "empty" {
        return Some(UntypedRange { lower: None, upper: None, lower_type: b'E', upper_type: b'E' });
    }
    let chars: Vec<char> = src.chars().collect();
    let mut i = 0;
    let skip_space = |i: &mut usize| {
        while *i < chars.len() && chars[*i].is_whitespace() {
            *i += 1;
        }
    };
    skip_space(&mut i);
    let lower_type = match chars.get(i)? {
        '(' => b'e',
        '[' => b'i',
        _ => return None,
    };
    i += 1;
    let (lower, lower_type) =
        if chars.get(i)? == &',' { (None, b'U') } else { (Some(range_value(&chars, &mut i)?), lower_type) };
    if chars.get(i)? != &',' {
        return None;
    }
    i += 1;
    let next = *chars.get(i)?;
    let (upper, upper_type) = if next == ')' || next == ']' {
        i += 1;
        (None, b'U')
    } else {
        let value = range_value(&chars, &mut i)?;
        let upper_type = match chars.get(i)? {
            ')' => b'e',
            ']' => b'i',
            _ => return None,
        };
        i += 1;
        (Some(value), upper_type)
    };
    skip_space(&mut i);
    (i == chars.len()).then_some(UntypedRange { lower, upper, lower_type, upper_type })
}

/// range_value matches pgx's rangeParseValue, reading a quoted or bare bound.
fn range_value(chars: &[char], i: &mut usize) -> Option<String> {
    let mut out = String::new();
    if chars.get(*i)? == &'"' {
        *i += 1;
        loop {
            let c = *chars.get(*i)?;
            *i += 1;
            match c {
                '\\' => {
                    out.push(*chars.get(*i)?);
                    *i += 1;
                }
                '"' => {
                    if chars.get(*i)? == &'"' {
                        out.push('"');
                        *i += 1;
                    } else {
                        return Some(out);
                    }
                }
                _ => out.push(c),
            }
        }
    }
    loop {
        let c = *chars.get(*i)?;
        match c {
            '\\' => {
                *i += 1;
                out.push(*chars.get(*i)?);
                *i += 1;
            }
            ',' | '[' | ']' | '(' | ')' => return Some(out),
            _ => {
                out.push(c);
                *i += 1;
            }
        }
    }
}

/// decode_array matches pgx's text array scanning into a []any, which flattens every dimension.
fn decode_array(element: u32, s: &str) -> Option<Any> {
    let array = parse_untyped_array(s)?;
    let count = array.cardinality;
    if array.elements.len() > count {
        return None;
    }
    let mut values = vec![Any::Nil; count];
    for (i, (text, quoted)) in array.elements.iter().enumerate() {
        let src = (text != "NULL" || *quoted).then_some(text.as_bytes());
        values[i] = decode(element, src)?;
    }
    Some(Any::Slice(values))
}

/// UntypedArray is a text array's elements and how many values its dimensions hold.
struct UntypedArray {
    elements: Vec<(String, bool)>,
    cardinality: usize,
}

/// parse_untyped_array matches pgx's parseUntypedTextArray.
fn parse_untyped_array(src: &str) -> Option<UntypedArray> {
    let chars: Vec<char> = src.chars().collect();
    let mut i = 0;
    let skip_space = |i: &mut usize| {
        while *i < chars.len() && chars[*i].is_whitespace() {
            *i += 1;
        }
    };
    skip_space(&mut i);
    let mut explicit: Vec<i64> = Vec::new();
    if chars.get(i)? == &'[' {
        loop {
            let c = *chars.get(i)?;
            i += 1;
            if c == '=' {
                break;
            } else if c != '[' {
                return None;
            }
            let lower = array_integer(&chars, &mut i)?;
            if chars.get(i)? != &':' {
                return None;
            }
            i += 1;
            let upper = array_integer(&chars, &mut i)?;
            if chars.get(i)? != &']' {
                return None;
            }
            i += 1;
            explicit.push((upper as i64) - (lower as i64) + 1);
        }
    }
    if chars.get(i)? != &'{' {
        return None;
    }
    i += 1;
    let mut implicit: Vec<i64> = vec![0];
    loop {
        if chars.get(i)? == &'{' {
            *implicit.last_mut().unwrap() = 1;
            implicit.push(0);
            i += 1;
        } else {
            break;
        }
    }
    let mut current = implicit.len() as i64 - 1;
    let mut counter = current;
    let mut elements = Vec::new();
    loop {
        let c = *chars.get(i)?;
        i += 1;
        match c {
            '{' => {
                if current == counter {
                    implicit[current as usize] += 1;
                }
                current += 1;
            }
            ',' => {}
            '}' => {
                current -= 1;
                if current < counter {
                    counter = current;
                }
            }
            _ => {
                i -= 1;
                let value = array_value(&chars, &mut i)?;
                if current == counter {
                    *implicit.get_mut(usize::try_from(current).ok()?)? += 1;
                }
                elements.push(value);
            }
        }
        if current < 0 {
            break;
        }
    }
    skip_space(&mut i);
    if i < chars.len() {
        return None;
    }
    let dimensions = if elements.is_empty() {
        Vec::new()
    } else if !explicit.is_empty() {
        explicit
    } else {
        implicit
    };
    let cardinality = match dimensions.split_first() {
        None => 0,
        Some((first, rest)) => {
            let count = rest.iter().fold(*first as i32, |acc, d| acc.wrapping_mul(*d as i32));
            count.max(0) as usize
        }
    };
    Some(UntypedArray { elements, cardinality })
}

/// array_integer matches pgx's arrayParseInteger.
fn array_integer(chars: &[char], i: &mut usize) -> Option<i32> {
    let mut s = String::new();
    loop {
        let c = *chars.get(*i)?;
        if c.is_ascii_digit() || c == '-' {
            s.push(c);
            *i += 1;
        } else {
            return gostd::parse_int(&s, 32).map(|v| v as i32);
        }
    }
}

/// array_value matches pgx's arrayParseValue, returning the element and whether it was quoted.
fn array_value(chars: &[char], i: &mut usize) -> Option<(String, bool)> {
    let mut out = String::new();
    if chars.get(*i)? == &'"' {
        *i += 1;
        loop {
            let mut c = *chars.get(*i)?;
            *i += 1;
            match c {
                '\\' => {
                    c = *chars.get(*i)?;
                    *i += 1;
                }
                '"' => {
                    chars.get(*i)?;
                    return Some((out, true));
                }
                _ => {}
            }
            out.push(c);
        }
    }
    loop {
        let c = *chars.get(*i)?;
        if c == ',' || c == '}' {
            return Some((out, false));
        }
        out.push(c);
        *i += 1;
    }
}

/// normalize matches the replay's ReadRows conversion of a decoded value.
pub fn normalize(value: Any, raw: &[u8]) -> Value {
    let raw_string = || Value::String(String::from_utf8_lossy(raw).into_owned());
    match value {
        Any::Nil => Value::Nil,
        Any::Bool(v) => Value::Bool(v),
        Any::Int16(v) => Value::Int16(v),
        Any::Int32(v) => Value::Int32(v),
        Any::Int64(v) => Value::Int64(v),
        Any::Uint32(v) => Value::Uint32(v),
        Any::Uint64(v) => Value::Uint64(v),
        Any::Float32(v) => Value::Float32(v),
        Any::Float64(v) => Value::Float64(v),
        Any::String(v) => Value::String(v),
        Any::Bytes(v) => Value::String(v.iter().map(|b| format!("{b:02x}")).collect()),
        Any::Uuid(v) => Value::String(v.iter().map(|b| format!("{b:02x}")).collect()),
        Any::Time(v) => Value::Time(v),
        Any::Infinity(v) => Value::Float64(match v {
            1 => f64::INFINITY,
            -1 => f64::NEG_INFINITY,
            _ => 0.0,
        }),
        Any::Range { .. } | Any::Multirange => raw_string(),
        Any::Map(entries) => Value::Bytes(json::marshal_map(&entries).into_bytes()),
        Any::Slice(values) => Value::String(format!("[{}]", any_row_to_string(&values))),
        other => Value::String(valuer_text(&other)),
    }
}

/// valuer_text returns what a pgtype value's driver.Valuer or fmt.Stringer method returns.
fn valuer_text(value: &Any) -> String {
    let fixed = |v: f64| format_shortest_fixed(v);
    match value {
        Any::Numeric(n) => n.text(),
        Any::Interval { micros, days, months } => interval_text(*micros, *days, *months),
        Any::TimeOfDay(usec) => {
            let hours = usec / 3_600_000_000;
            let rest = usec - hours * 3_600_000_000;
            let minutes = rest / 60_000_000;
            let rest = rest - minutes * 60_000_000;
            let seconds = rest / 1_000_000;
            let rest = rest - seconds * 1_000_000;
            format!("{hours:02}:{minutes:02}:{seconds:02}.{rest:06}")
        }
        Any::Bits { bytes, len } => {
            (0..*len).map(|i| if bytes[(i / 8) as usize] & (128 >> (i % 8)) > 0 { '1' } else { '0' }).collect()
        }
        Any::Prefix(text) | Any::HardwareAddr(text) => text.clone(),
        Any::Point(x, y) => format!("({},{})", fixed(*x), fixed(*y)),
        Any::Lseg(p) => format!("[({},{}),({},{})]", fixed(p[0]), fixed(p[1]), fixed(p[2]), fixed(p[3])),
        Any::Box(p) => format!("({},{}),({},{})", fixed(p[0]), fixed(p[1]), fixed(p[2]), fixed(p[3])),
        Any::Line(p) => format!("{{{},{},{}}}", fixed(p[0]), fixed(p[1]), fixed(p[2])),
        Any::Circle(p) => format!("<({},{}),{}>", fixed(p[0]), fixed(p[1]), fixed(p[2])),
        Any::Path { points, closed } => {
            let inner: Vec<String> = points.iter().map(|(x, y)| format!("({},{})", fixed(*x), fixed(*y))).collect();
            if *closed { format!("({})", inner.join(",")) } else { format!("[{}]", inner.join(",")) }
        }
        Any::Polygon(points) => {
            let inner: Vec<String> = points.iter().map(|(x, y)| format!("({},{})", fixed(*x), fixed(*y))).collect();
            format!("({})", inner.join(","))
        }
        Any::Tid(block, offset) => format!("({block},{offset})"),
        Any::TsVector(lexemes) => tsvector_text(lexemes),
        _ => unreachable!("not a driver.Valuer or fmt.Stringer: {value:?}"),
    }
}

/// read_cell matches the replay's ReadRows for one cell: decode, fall back to the raw text on error, and normalize.
pub fn read_cell(oid: u32, src: Option<&[u8]>) -> Value {
    match decode(oid, src) {
        Some(value) => normalize(value, src.unwrap_or_default()),
        None => Value::String(String::from_utf8_lossy(src.unwrap_or_default()).into_owned()),
    }
}

/// quote matches the replay's `"%s"` quoting, which only escapes double quotes.
fn quote(s: &str) -> String {
    format!("\"{}\"", s.replace('"', "\\\""))
}

/// row_to_string matches the replay's RowToString.
pub fn row_to_string(row: &[Value]) -> String {
    row.iter()
        .map(|value| match value {
            Value::Nil => "\u{FFFD}".to_string(),
            Value::Bool(v) => v.to_string(),
            Value::Int16(v) => v.to_string(),
            Value::Int32(v) => v.to_string(),
            Value::Int64(v) => v.to_string(),
            Value::Uint32(v) => v.to_string(),
            Value::Uint64(v) => v.to_string(),
            Value::Float32(v) => format_f6_32(*v),
            Value::Float64(v) => format_f6_64(*v),
            Value::String(v) => quote(v),
            Value::Time(v) => v.string(),
            Value::Bytes(v) => format!("[{}]", v.iter().map(u8::to_string).collect::<Vec<_>>().join(" ")),
        })
        .collect::<Vec<_>>()
        .join(", ")
}

/// any_row_to_string matches RowToString on the unnormalized elements of a []any.
fn any_row_to_string(values: &[Any]) -> String {
    values
        .iter()
        .map(|value| match value {
            Any::Nil => "\u{FFFD}".to_string(),
            Any::Float32(v) => format_f6_32(*v),
            Any::Float64(v) => format_f6_64(*v),
            Any::String(v) => quote(v),
            other => go_v(other),
        })
        .collect::<Vec<_>>()
        .join(", ")
}

/// go_v matches fmt's `%v` of a value.
fn go_v(value: &Any) -> String {
    let point = |x: f64, y: f64| format!("{{{} {}}}", format_v64(x), format_v64(y));
    let bytes = |b: &[u8]| format!("[{}]", b.iter().map(u8::to_string).collect::<Vec<_>>().join(" "));
    match value {
        Any::Nil => "<nil>".to_string(),
        Any::Bool(v) => v.to_string(),
        Any::Int16(v) => v.to_string(),
        Any::Int32(v) => v.to_string(),
        Any::Int64(v) => v.to_string(),
        Any::Uint32(v) => v.to_string(),
        Any::Uint64(v) => v.to_string(),
        Any::Float32(v) => format_v32(*v),
        Any::Float64(v) => format_v64(*v),
        Any::String(v) => v.clone(),
        Any::Bytes(v) => bytes(v),
        Any::Uuid(v) => bytes(v),
        Any::Time(v) => v.string(),
        Any::Infinity(v) => match v {
            0 => "finite",
            1 => "infinity",
            -1 => "-infinity",
            _ => "invalid",
        }
        .to_string(),
        Any::Numeric(n) => format!(
            "{{{} {} {} {} true}}",
            n.int.as_deref().unwrap_or("<nil>"),
            n.exp,
            n.nan,
            go_v(&Any::Infinity(n.infinity))
        ),
        Any::Interval { micros, days, months } => format!("{{{micros} {days} {months} true}}"),
        Any::TimeOfDay(usec) => format!("{{{usec} true}}"),
        Any::Bits { bytes: b, len } => format!("{{{} {len} true}}", bytes(b)),
        Any::Prefix(text) | Any::HardwareAddr(text) => text.clone(),
        Any::Point(x, y) => format!("{{{} true}}", point(*x, *y)),
        Any::Lseg(p) | Any::Box(p) => format!("{{[{} {}] true}}", point(p[0], p[1]), point(p[2], p[3])),
        Any::Line(p) => format!("{{{} {} {} true}}", format_v64(p[0]), format_v64(p[1]), format_v64(p[2])),
        Any::Circle(p) => format!("{{{} {} true}}", point(p[0], p[1]), format_v64(p[2])),
        Any::Path { points, closed } => {
            let inner: Vec<String> = points.iter().map(|(x, y)| point(*x, *y)).collect();
            format!("{{[{}] {closed} true}}", inner.join(" "))
        }
        Any::Polygon(points) => {
            let inner: Vec<String> = points.iter().map(|(x, y)| point(*x, *y)).collect();
            format!("{{[{}] true}}", inner.join(" "))
        }
        Any::Tid(block, offset) => format!("{{{block} {offset} true}}"),
        Any::TsVector(lexemes) => {
            let inner: Vec<String> = lexemes
                .iter()
                .map(|(word, positions)| {
                    let positions: Vec<String> = positions.iter().map(|(p, w)| position_string(*p, *w)).collect();
                    format!("{{{word} [{}]}}", positions.join(" "))
                })
                .collect();
            format!("{{[{}] true}}", inner.join(" "))
        }
        Any::Range { lower, upper, lower_type, upper_type } => {
            format!("{{{} {} {} {} true}}", go_v(lower), go_v(upper), *lower_type as char, *upper_type as char)
        }
        Any::Multirange => unreachable!("multiranges are never array elements"),
        Any::Map(entries) => {
            let inner: Vec<String> = entries.iter().map(|(k, v)| format!("{k}:{}", go_v(v))).collect();
            format!("map[{}]", inner.join(" "))
        }
        Any::Slice(values) => format!("[{}]", values.iter().map(go_v).collect::<Vec<_>>().join(" ")),
    }
}
