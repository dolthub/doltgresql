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

/// parse_int matches strconv.ParseInt(s, 10, bits).
pub fn parse_int(s: &str, bits: u32) -> Option<i64> {
    let value: i64 = s.parse().ok()?;
    let limit = 1i128 << (bits - 1);
    ((value as i128) >= -limit && (value as i128) < limit).then_some(value)
}

/// parse_uint matches strconv.ParseUint(s, 10, bits), which accepts no sign.
pub fn parse_uint(s: &str, bits: u32) -> Option<u64> {
    if s.starts_with('+') {
        return None;
    }
    let value: u64 = s.parse().ok()?;
    (bits == 64 || value < (1u64 << bits)).then_some(value)
}

/// special matches strconv's parsing of infinities and NaN, where only infinities take a sign.
fn special(s: &str) -> Option<f64> {
    let (negative, rest) = match s.as_bytes().first() {
        Some(b'+') => (false, &s[1..]),
        Some(b'-') => (true, &s[1..]),
        _ => (false, s),
    };
    let lower = rest.to_ascii_lowercase();
    if lower == "inf" || lower == "infinity" {
        return Some(if negative { f64::NEG_INFINITY } else { f64::INFINITY });
    }
    (rest.len() == s.len() && lower == "nan").then_some(f64::NAN)
}

/// is_decimal_float reports whether s is the decimal syntax that strconv.ParseFloat accepts.
fn is_decimal_float(s: &str) -> bool {
    let b = s.as_bytes();
    let mut i = usize::from(matches!(b.first(), Some(b'+' | b'-')));
    let mut digits = 0;
    while i < b.len() && b[i].is_ascii_digit() {
        i += 1;
        digits += 1;
    }
    if i < b.len() && b[i] == b'.' {
        i += 1;
        while i < b.len() && b[i].is_ascii_digit() {
            i += 1;
            digits += 1;
        }
    }
    if digits == 0 {
        return false;
    }
    if i < b.len() && (b[i] == b'e' || b[i] == b'E') {
        i += 1;
        if i < b.len() && (b[i] == b'+' || b[i] == b'-') {
            i += 1;
        }
        let start = i;
        while i < b.len() && b[i].is_ascii_digit() {
            i += 1;
        }
        if i == start {
            return false;
        }
    }
    i == b.len()
}

/// parse_float64 matches strconv.ParseFloat(s, 64), rejecting values that overflow.
pub fn parse_float64(s: &str) -> Option<f64> {
    if let Some(value) = special(s) {
        return Some(value);
    }
    if !is_decimal_float(s) {
        return None;
    }
    let value: f64 = s.parse().ok()?;
    value.is_finite().then_some(value)
}

/// parse_float32 matches strconv.ParseFloat(s, 32), rejecting values that overflow.
pub fn parse_float32(s: &str) -> Option<f32> {
    if let Some(value) = special(s) {
        return Some(value as f32);
    }
    if !is_decimal_float(s) {
        return None;
    }
    let value: f32 = s.parse().ok()?;
    value.is_finite().then_some(value)
}

/// special_text formats infinities and NaN the way fmt does, with a sign on infinities.
fn special_text(value: f64) -> Option<String> {
    if value.is_nan() {
        Some("NaN".to_string())
    } else if value.is_infinite() {
        Some(if value > 0.0 { "+Inf" } else { "-Inf" }.to_string())
    } else {
        None
    }
}

/// format_f6_64 matches fmt's `%f` for a float64.
pub fn format_f6_64(value: f64) -> String {
    format_f64(value, 6)
}

/// format_f64 matches fmt's `%.Nf` for a float64.
pub fn format_f64(value: f64, precision: usize) -> String {
    special_text(value).unwrap_or_else(|| format!("{value:.precision$}"))
}

/// format_f6_32 matches fmt's `%f` for a float32.
pub fn format_f6_32(value: f32) -> String {
    special_text(value as f64).unwrap_or_else(|| format!("{value:.6}"))
}

/// format_shortest_fixed matches strconv.FormatFloat(value, 'f', -1, 64).
pub fn format_shortest_fixed(value: f64) -> String {
    special_text(value).unwrap_or_else(|| format!("{value}"))
}

/// shortest_digits returns the shortest round-trip digits of a `{:e}` formatted value and its decimal point position.
fn shortest_digits(scientific: &str) -> (String, i32) {
    let (mantissa, exponent) = scientific.split_once('e').unwrap();
    let digits: String = mantissa.chars().filter(|c| c.is_ascii_digit()).collect();
    (digits, exponent.parse::<i32>().unwrap() + 1)
}

/// format_g matches strconv.FormatFloat(value, 'g', -1, bits), which is also fmt's `%v` for floats.
fn format_g(value: f64, scientific: String) -> String {
    if let Some(text) = special_text(value) {
        return text;
    }
    if value == 0.0 {
        return if value.is_sign_negative() { "-0" } else { "0" }.to_string();
    }
    let negative = value < 0.0;
    let (digits, point) = shortest_digits(scientific.trim_start_matches('-'));
    let exponent = point - 1;
    let mut out = String::new();
    if negative {
        out.push('-');
    }
    if !(-4..6).contains(&exponent) {
        out.push_str(&digits[..1]);
        if digits.len() > 1 {
            out.push('.');
            out.push_str(&digits[1..]);
        }
        out.push('e');
        out.push(if exponent < 0 { '-' } else { '+' });
        out.push_str(&format!("{:02}", exponent.abs()));
        return out;
    }
    if point <= 0 {
        out.push_str("0.");
        out.push_str(&"0".repeat((-point) as usize));
        out.push_str(&digits);
    } else if point as usize >= digits.len() {
        out.push_str(&digits);
        out.push_str(&"0".repeat(point as usize - digits.len()));
    } else {
        out.push_str(&digits[..point as usize]);
        out.push('.');
        out.push_str(&digits[point as usize..]);
    }
    out
}

/// format_v64 matches fmt's `%v` for a float64.
pub fn format_v64(value: f64) -> String {
    format_g(value, format!("{value:e}"))
}

/// format_v32 matches fmt's `%v` for a float32.
pub fn format_v32(value: f32) -> String {
    format_g(value as f64, format!("{value:e}"))
}

/// Location is the location of a time.Time that pgx's text parsing produces.
#[derive(Clone, Copy, Debug, PartialEq)]
pub enum Location {
    Utc,
    Fixed(i32),
}

/// GoTime is a time.Time without a monotonic reading.
#[derive(Clone, Copy, Debug)]
pub struct GoTime {
    pub unix: i64,
    pub nanos: u32,
    pub location: Location,
}

impl GoTime {
    /// go_eq matches Go's `==`, where fixed zones are only shared for whole-hour offsets from -12 to +14 hours.
    pub fn go_eq(&self, other: &GoTime) -> bool {
        if self.unix != other.unix || self.nanos != other.nanos {
            return false;
        }
        match (self.location, other.location) {
            (Location::Utc, Location::Utc) => true,
            (Location::Fixed(a), Location::Fixed(b)) => a == b && a % 3600 == 0 && (-12..=14).contains(&(a / 3600)),
            _ => false,
        }
    }

    fn offset(&self) -> i32 {
        match self.location {
            Location::Utc => 0,
            Location::Fixed(offset) => offset,
        }
    }

    /// string matches time.Time.String().
    pub fn string(&self) -> String {
        let offset = self.offset();
        let local = self.unix + offset as i64;
        let days = local.div_euclid(86400);
        let seconds = local.rem_euclid(86400);
        let (year, month, day) = civil_from_days(days);
        let mut out = String::new();
        if year < 0 {
            out.push('-');
        }
        out.push_str(&format!(
            "{:04}-{:02}-{:02} {:02}:{:02}:{:02}",
            year.unsigned_abs(),
            month,
            day,
            seconds / 3600,
            seconds / 60 % 60,
            seconds % 60
        ));
        if self.nanos != 0 {
            let fraction = format!("{:09}", self.nanos);
            out.push('.');
            out.push_str(fraction.trim_end_matches('0'));
        }
        let zone = |out: &mut String| {
            let minutes = offset / 60;
            out.push(if minutes < 0 { '-' } else { '+' });
            out.push_str(&format!("{:02}{:02}", minutes.abs() / 60, minutes.abs() % 60));
        };
        out.push(' ');
        zone(&mut out);
        out.push(' ');
        match self.location {
            Location::Utc => out.push_str("UTC"),
            Location::Fixed(_) => zone(&mut out),
        }
        out
    }
}

/// days_from_civil returns the days since 1970-01-01 of a proleptic Gregorian date.
pub fn days_from_civil(year: i64, month: i64, day: i64) -> i64 {
    let year = if month <= 2 { year - 1 } else { year };
    let era = year.div_euclid(400);
    let year_of_era = year - era * 400;
    let month_index = (month + 9) % 12;
    let day_of_year = (153 * month_index + 2) / 5 + day - 1;
    let day_of_era = year_of_era * 365 + year_of_era / 4 - year_of_era / 100 + day_of_year;
    era * 146097 + day_of_era - 719468
}

/// civil_from_days returns the proleptic Gregorian date of a day count since 1970-01-01.
pub fn civil_from_days(days: i64) -> (i64, i64, i64) {
    let z = days + 719468;
    let era = z.div_euclid(146097);
    let day_of_era = z - era * 146097;
    let year_of_era = (day_of_era - day_of_era / 1460 + day_of_era / 36524 - day_of_era / 146096) / 365;
    let day_of_year = day_of_era - (365 * year_of_era + year_of_era / 4 - year_of_era / 100);
    let month_index = (5 * day_of_year + 2) / 153;
    let day = day_of_year - (153 * month_index + 2) / 5 + 1;
    let month = if month_index < 10 { month_index + 3 } else { month_index - 9 };
    let year = year_of_era + era * 400 + i64::from(month <= 2);
    (year, month, day)
}

/// days_in returns the number of days in a month.
fn days_in(month: i64, year: i64) -> i64 {
    match month {
        2 if year % 4 == 0 && (year % 100 != 0 || year % 400 == 0) => 29,
        2 => 28,
        4 | 6 | 9 | 11 => 30,
        _ => 31,
    }
}

/// go_date matches time.Date for a month within 1 to 12.
#[allow(clippy::too_many_arguments)]
pub fn go_date(
    year: i64,
    month: i64,
    day: i64,
    hour: i64,
    min: i64,
    sec: i64,
    nanos: i64,
    location: Location,
) -> GoTime {
    let offset = match location {
        Location::Utc => 0,
        Location::Fixed(offset) => offset as i64,
    };
    let seconds = sec + nanos.div_euclid(1_000_000_000);
    let nanos = nanos.rem_euclid(1_000_000_000);
    let local = days_from_civil(year, month, 1) * 86400 + (day - 1) * 86400 + hour * 3600 + min * 60 + seconds;
    GoTime { unix: local - offset, nanos: nanos as u32, location }
}

/// TimeZone is the zone suffix of a time layout.
#[derive(Clone, Copy, PartialEq)]
pub enum TimeZone {
    None,
    Hour,
    Minute,
    Second,
}

/// get_num matches the time package's getnum, reading one or two digits, or exactly two when fixed.
fn get_num(s: &[u8], fixed: bool) -> Option<(i64, &[u8])> {
    let digit = |i: usize| s.get(i).is_some_and(u8::is_ascii_digit);
    if !digit(0) {
        return None;
    }
    if !digit(1) {
        if fixed {
            return None;
        }
        return Some(((s[0] - b'0') as i64, &s[1..]));
    }
    Some((((s[0] - b'0') * 10 + (s[1] - b'0')) as i64, &s[2..]))
}

/// go_atoi matches the time package's atoi, which takes an optional sign.
fn go_atoi(s: &[u8]) -> Option<i64> {
    let (negative, digits) = match s.first() {
        Some(b'-') => (true, &s[1..]),
        Some(b'+') => (false, &s[1..]),
        _ => (false, s),
    };
    if digits.is_empty() || !digits.iter().all(u8::is_ascii_digit) {
        return None;
    }
    let mut value: i64 = 0;
    for d in digits {
        value = value.checked_mul(10)?.checked_add((d - b'0') as i64)?;
        if value >= 1 << 62 {
            return None;
        }
    }
    Some(if negative { -value } else { value })
}

/// parse_time matches time.Parse with the layout "2006-01-02 15:04:05.999999999" and the given zone suffix.
pub fn parse_time(value: &str, zone: TimeZone) -> Option<GoTime> {
    let mut v = value.as_bytes();
    let literal = |v: &mut &[u8], text: &[u8]| -> Option<()> { v.starts_with(text).then(|| *v = &v[text.len()..]) };
    if v.len() < 4 {
        return None;
    }
    let year = go_atoi(&v[..4])?;
    v = &v[4..];
    literal(&mut v, b"-")?;
    let (month, rest) = get_num(v, true)?;
    v = rest;
    if !(1..=12).contains(&month) {
        return None;
    }
    literal(&mut v, b"-")?;
    let (day, rest) = get_num(v, true)?;
    v = rest;
    literal(&mut v, b" ")?;
    let (hour, rest) = get_num(v, false)?;
    v = rest;
    if !(0..24).contains(&hour) {
        return None;
    }
    literal(&mut v, b":")?;
    let (min, rest) = get_num(v, true)?;
    v = rest;
    if !(0..60).contains(&min) {
        return None;
    }
    literal(&mut v, b":")?;
    let (sec, rest) = get_num(v, true)?;
    v = rest;
    if !(0..60).contains(&sec) {
        return None;
    }
    let mut nanos = 0;
    if v.len() >= 2 && (v[0] == b'.' || v[0] == b',') && v[1].is_ascii_digit() {
        let mut i = 0;
        while i + 1 < v.len() && v[i + 1].is_ascii_digit() {
            i += 1;
        }
        let nbytes = (1 + i).min(10);
        nanos = go_atoi(&v[1..nbytes])?;
        for _ in 0..10 - nbytes {
            nanos *= 10;
        }
        v = &v[1 + i..];
    }
    let mut location = None;
    if zone != TimeZone::None {
        if v.first() == Some(&b'Z') {
            v = &v[1..];
            location = Some(Location::Utc);
        } else {
            let (sign, hours, minutes, seconds) = match zone {
                TimeZone::Hour if v.len() >= 3 => {
                    let parts = (v[0], &v[1..3], &b"00"[..], &b"00"[..]);
                    v = &v[3..];
                    parts
                }
                TimeZone::Minute if v.len() >= 6 && v[3] == b':' => {
                    let parts = (v[0], &v[1..3], &v[4..6], &b"00"[..]);
                    v = &v[6..];
                    parts
                }
                TimeZone::Second if v.len() >= 9 && v[3] == b':' && v[6] == b':' => {
                    let parts = (v[0], &v[1..3], &v[4..6], &v[7..9]);
                    v = &v[9..];
                    parts
                }
                _ => return None,
            };
            let (h, _) = get_num(hours, true)?;
            let (m, _) = get_num(minutes, true)?;
            let (s, _) = get_num(seconds, true)?;
            if h > 24 || m > 60 || s > 60 {
                return None;
            }
            let offset = ((h * 60 + m) * 60 + s) as i32;
            location = Some(Location::Fixed(match sign {
                b'+' => offset,
                b'-' => -offset,
                _ => return None,
            }));
        }
    }
    if !v.is_empty() || day < 1 || day > days_in(month, year) {
        return None;
    }
    Some(go_date(year, month, day, hour, min, sec, nanos, location.unwrap_or(Location::Utc)))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn floats_format_like_fmt() {
        assert_eq!(format_v64(1000000.0), "1e+06");
        assert_eq!(format_v64(123456.0), "123456");
        assert_eq!(format_v64(0.0001), "0.0001");
        assert_eq!(format_v64(0.00001), "1e-05");
        assert_eq!(format_v64(1.5), "1.5");
        assert_eq!(format_v64(-2.25e-7), "-2.25e-07");
        assert_eq!(format_f6_64(f64::INFINITY), "+Inf");
        assert_eq!(format_f6_64(2.5), "2.500000");
        assert_eq!(format_shortest_fixed(1e21), "1000000000000000000000");
    }

    #[test]
    fn times_parse_like_go() {
        let t = parse_time("2000-01-01 00:00:00-08", TimeZone::Hour).unwrap();
        assert_eq!(t.string(), "2000-01-01 00:00:00 -0800 -0800");
        let t = parse_time("1999-12-31 23:59:59.5", TimeZone::None).unwrap();
        assert_eq!(t.string(), "1999-12-31 23:59:59.5 +0000 UTC");
        assert!(parse_time("2000-02-30 00:00:00", TimeZone::None).is_none());
    }
}
