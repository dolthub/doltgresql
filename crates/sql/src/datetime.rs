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

//! Dates, times, timestamps, and intervals as Postgres represents, reads, and prints them.
//!
//! A date is a day count from 2000-01-01, a timestamp a microsecond count from 2000-01-01 00:00:00 (UTC for
//! timestamptz), a time a microsecond count from midnight, and an interval months, days, and microseconds. Infinities
//! are the extreme values.

use std::collections::HashMap;
use std::sync::OnceLock;

use chrono::{NaiveDate, NaiveDateTime, Offset, TimeZone};

use crate::error::{PgError, Result, code};

/// USECS_PER_* are microsecond counts.
pub const USECS_PER_SEC: i64 = 1_000_000;
pub const USECS_PER_MINUTE: i64 = 60 * USECS_PER_SEC;
pub const USECS_PER_HOUR: i64 = 60 * USECS_PER_MINUTE;
pub const USECS_PER_DAY: i64 = 24 * USECS_PER_HOUR;

/// POSTGRES_EPOCH_JDATE is the Julian day of 2000-01-01.
pub const POSTGRES_EPOCH_JDATE: i64 = 2_451_545;

/// UNIX_EPOCH_DAYS is 1970-01-01 as a day count from 2000-01-01.
pub const UNIX_EPOCH_DAYS: i64 = -10_957;

/// DATE_NOBEGIN and DATE_NOEND are -infinity and infinity dates.
pub const DATE_NOBEGIN: i32 = i32::MIN;
pub const DATE_NOEND: i32 = i32::MAX;

/// TIMESTAMP_NOBEGIN and TIMESTAMP_NOEND are -infinity and infinity timestamps.
pub const TIMESTAMP_NOBEGIN: i64 = i64::MIN;
pub const TIMESTAMP_NOEND: i64 = i64::MAX;

/// MIN_TIMESTAMP and END_TIMESTAMP bound the finite timestamps: 4714-11-24 BC and 294277-01-01.
const MIN_TIMESTAMP: i64 = -211_813_488_000_000_000;
const END_TIMESTAMP: i64 = 9_223_371_331_200_000_000;

/// MIN_JULIAN and END_DATE_JULIAN bound the finite dates.
const MIN_JULIAN: i64 = 0;
const END_DATE_JULIAN: i64 = 2_147_483_494;

/// Interval is a Postgres interval.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash)]
pub struct Interval {
    pub months: i32,
    pub days: i32,
    pub micros: i64,
}

impl Interval {
    /// cmp_key returns the value Postgres compares intervals by: their length with 30-day months and 24-hour days.
    pub fn cmp_key(&self) -> i128 {
        self.months as i128 * 30 * USECS_PER_DAY as i128
            + self.days as i128 * USECS_PER_DAY as i128
            + self.micros as i128
    }
}

/// date2j returns the Julian day of a date, whose year is astronomical (year 0 is 1 BC), as Postgres' date2j does.
pub fn date2j(year: i64, month: i64, day: i64) -> i64 {
    let (mut y, mut m) = (year, month);
    if m > 2 {
        m += 1;
        y += 4800;
    } else {
        m += 13;
        y += 4799;
    }
    let century = y.div_euclid(100);
    let mut julian = y * 365 - 32167;
    julian += y.div_euclid(4) - century + century.div_euclid(4);
    julian += 7834 * m / 256 + day;
    julian
}

/// j2date returns the astronomical year, month, and day of a Julian day, as Postgres' j2date does.
pub fn j2date(jd: i64) -> (i64, i64, i64) {
    let mut julian = jd as u64;
    julian += 32044;
    let mut quad = julian / 146097;
    let extra = (julian - quad * 146097) * 4 + 3;
    julian += 60 + quad * 3 + extra / 146097;
    quad = julian / 1461;
    julian -= quad * 1461;
    let mut y = (julian * 4 / 1461) as i64;
    julian = if y != 0 { (julian + 305) % 365 } else { (julian + 306) % 366 } + 123;
    y += quad as i64 * 4;
    let year = y - 4800;
    let quad = julian * 2141 / 65536;
    let day = julian as i64 - (7834 * quad / 256) as i64;
    let month = ((quad + 10) % 12 + 1) as i64;
    (year, month, day)
}

/// days_in_month returns the number of days of a month of an astronomical year.
pub fn days_in_month(year: i64, month: i64) -> i64 {
    match month {
        2 if year.rem_euclid(4) == 0 && (year.rem_euclid(100) != 0 || year.rem_euclid(400) == 0) => 29,
        2 => 28,
        4 | 6 | 9 | 11 => 30,
        _ => 31,
    }
}

/// Fields are the broken-down parts of a date and time, with an astronomical year.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub struct Fields {
    pub year: i64,
    pub month: i64,
    pub day: i64,
    pub hour: i64,
    pub minute: i64,
    pub second: i64,
    pub micros: i64,
}

/// fields_of_timestamp breaks a timestamp into its fields.
pub fn fields_of_timestamp(ts: i64) -> Fields {
    let days = ts.div_euclid(USECS_PER_DAY);
    let time = ts.rem_euclid(USECS_PER_DAY);
    let (year, month, day) = j2date(days + POSTGRES_EPOCH_JDATE);
    Fields {
        year,
        month,
        day,
        hour: time / USECS_PER_HOUR,
        minute: time / USECS_PER_MINUTE % 60,
        second: time / USECS_PER_SEC % 60,
        micros: time % USECS_PER_SEC,
    }
}

/// timestamp_of_fields builds a timestamp from fields, failing when it is out of range.
pub fn timestamp_of_fields(f: &Fields) -> Option<i64> {
    let days = date2j(f.year, f.month, f.day) - POSTGRES_EPOCH_JDATE;
    let time = f.hour * USECS_PER_HOUR + f.minute * USECS_PER_MINUTE + f.second * USECS_PER_SEC + f.micros;
    let ts = days.checked_mul(USECS_PER_DAY)?.checked_add(time)?;
    (MIN_TIMESTAMP..END_TIMESTAMP).contains(&ts).then_some(ts)
}

/// date_in_range reports whether a Julian day is a valid date.
fn date_in_range(julian: i64) -> bool {
    (MIN_JULIAN..=END_DATE_JULIAN).contains(&julian)
}

/// Style is the DateStyle output format.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Style {
    Iso,
    Sql,
    Postgres,
    German,
}

/// Order is the DateStyle order of ambiguous day, month, and year fields.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Order {
    Mdy,
    Dmy,
    Ymd,
}

/// IntervalStyle is the IntervalStyle output format.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum IntervalStyle {
    Postgres,
    PostgresVerbose,
    SqlStandard,
    Iso8601,
}

/// Zone is a time zone: a fixed offset east of UTC with its name, or a zone of the IANA database.
#[derive(Clone, Debug, PartialEq)]
pub enum Zone {
    Fixed { offset: i32, name: String },
    Tz(chrono_tz::Tz),
}

/// abbreviations returns Postgres' default time zone abbreviations with their offsets east of UTC.
fn abbreviations() -> &'static HashMap<String, (i32, bool)> {
    static ABBREVIATIONS: OnceLock<HashMap<String, (i32, bool)>> = OnceLock::new();
    ABBREVIATIONS.get_or_init(|| {
        include_str!("tzabbrevs.tsv")
            .lines()
            .filter_map(|line| {
                let mut f = line.split('\t');
                let (name, offset, dst) = (f.next()?, f.next()?, f.next()?);
                Some((name.to_ascii_lowercase(), (offset.parse().ok()?, dst == "t")))
            })
            .collect()
    })
}

/// parse_offset reads a numeric zone offset such as `+05`, `-0800`, or `+05:30:15`, returning seconds east of UTC.
fn parse_offset(text: &str) -> Option<i32> {
    let (sign, rest) = match text.as_bytes().first()? {
        b'+' => (1, &text[1..]),
        b'-' => (-1, &text[1..]),
        _ => return None,
    };
    let parts: Vec<&str> = rest.split(':').collect();
    let (h, m, s) = if parts.len() == 1 {
        let digits = parts[0];
        if !digits.bytes().all(|b| b.is_ascii_digit()) || digits.is_empty() || digits.len() > 6 {
            return None;
        }
        match digits.len() {
            1 | 2 => (digits.parse().ok()?, 0, 0),
            3 | 4 => (digits[..digits.len() - 2].parse().ok()?, digits[digits.len() - 2..].parse().ok()?, 0),
            _ => (
                digits[..digits.len() - 4].parse().ok()?,
                digits[digits.len() - 4..digits.len() - 2].parse().ok()?,
                digits[digits.len() - 2..].parse().ok()?,
            ),
        }
    } else {
        let num = |s: Option<&&str>| s.map_or(Some(0), |s| s.parse::<i32>().ok());
        (num(parts.first())?, num(parts.get(1))?, num(parts.get(2))?)
    };
    if h > 15 * 24 || m > 59 || s > 59 {
        return None;
    }
    Some(sign * (h * 3600 + m * 60 + s))
}

impl Zone {
    /// named returns the zone of a TimeZone setting or an AT TIME ZONE argument.
    pub fn named(name: &str) -> Option<Zone> {
        let trimmed = name.trim();
        if let Some(tz) = chrono_tz::TZ_VARIANTS.iter().find(|z| z.name().eq_ignore_ascii_case(trimmed)) {
            return Some(Zone::Tz(*tz));
        }
        if let Some(&(offset, _)) = abbreviations().get(&trimmed.to_ascii_lowercase()) {
            return Some(Zone::Fixed { offset, name: trimmed.to_ascii_uppercase() });
        }
        // <-07>+07 and POSIX names give their offset west of UTC after the name.
        let rest = if let Some(end) = trimmed.strip_prefix('<').and_then(|r| r.find('>')) {
            &trimmed[end + 2..]
        } else {
            trimmed.trim_start_matches(|c: char| c.is_ascii_alphabetic())
        };
        let name_part = &trimmed[..trimmed.len() - rest.len()];
        let offset_text = rest.split(|c: char| c.is_ascii_alphabetic() || c == ',').next().unwrap_or("");
        if offset_text.is_empty() {
            return None;
        }
        let signed =
            if offset_text.starts_with(['+', '-']) { offset_text.to_string() } else { format!("+{offset_text}") };
        let west = parse_offset(&signed)?;
        let display = name_part.trim_start_matches('<').trim_end_matches('>').to_string();
        Some(Zone::Fixed { offset: -west, name: display })
    }

    /// offset_at returns the zone's offset east of UTC, and its abbreviation, at a UTC timestamp.
    pub fn offset_at(&self, utc: i64) -> (i32, String) {
        match self {
            Zone::Fixed { offset, name } => (*offset, name.clone()),
            Zone::Tz(tz) => {
                let naive = naive_of(utc);
                let offset = tz.offset_from_utc_datetime(&naive);
                let seconds = offset.fix().local_minus_utc();
                let abbreviation =
                    chrono_tz::OffsetName::abbreviation(&offset).map_or_else(|| numeric_name(seconds), str::to_string);
                (seconds, abbreviation)
            }
        }
    }

    /// offset_for_local returns the zone's offset east of UTC for a local timestamp, preferring the earlier offset in
    /// a spring-forward gap and the later one in a fall-back overlap, as Postgres does.
    pub fn offset_for_local(&self, local: i64) -> i32 {
        match self {
            Zone::Fixed { offset, .. } => *offset,
            Zone::Tz(tz) => {
                let naive = naive_of(local);
                match tz.offset_from_local_datetime(&naive) {
                    chrono::LocalResult::Single(offset) => offset.fix().local_minus_utc(),
                    chrono::LocalResult::Ambiguous(_, later) => later.fix().local_minus_utc(),
                    chrono::LocalResult::None => {
                        let before = naive - chrono::Duration::hours(3);
                        tz.offset_from_local_datetime(&before).earliest().map_or(0, |o| o.fix().local_minus_utc())
                    }
                }
            }
        }
    }
}

/// numeric_name names an offset the way a zone without an abbreviation shows it, like `+05` or `-0330`.
fn numeric_name(seconds: i32) -> String {
    let sign = if seconds < 0 { '-' } else { '+' };
    let abs = seconds.abs();
    let (h, m) = (abs / 3600, abs / 60 % 60);
    if m == 0 { format!("{sign}{h:02}") } else { format!("{sign}{h:02}{m:02}") }
}

/// naive_of converts a timestamp to chrono's date-time, clamping at chrono's range.
fn naive_of(ts: i64) -> NaiveDateTime {
    let unix = ts.saturating_add(-UNIX_EPOCH_DAYS * USECS_PER_DAY);
    chrono::DateTime::from_timestamp_micros(unix).map(|d| d.naive_utc()).unwrap_or_else(|| {
        if ts > 0 { NaiveDate::MAX.and_hms_opt(0, 0, 0).unwrap() } else { NaiveDate::MIN.and_hms_opt(0, 0, 0).unwrap() }
    })
}

/// Format is how values print: the session's DateStyle, IntervalStyle, and time zone.
#[derive(Clone, Debug)]
pub struct Format {
    pub style: Style,
    pub order: Order,
    pub interval_style: IntervalStyle,
    pub zone: Zone,
}

impl Default for Format {
    fn default() -> Format {
        Format {
            style: Style::Iso,
            order: Order::Mdy,
            interval_style: IntervalStyle::Postgres,
            zone: Zone::Fixed { offset: 0, name: "UTC".into() },
        }
    }
}

impl Format {
    /// from_settings reads the formats from the DateStyle, IntervalStyle, and TimeZone settings.
    pub fn from_settings(date_style: &str, interval_style: &str, timezone: &str) -> Format {
        let mut format = Format::default();
        for word in date_style.split(',').map(|w| w.trim().to_ascii_lowercase()) {
            match word.as_str() {
                "iso" => format.style = Style::Iso,
                "sql" => format.style = Style::Sql,
                "postgres" => format.style = Style::Postgres,
                "german" => format.style = Style::German,
                "mdy" => format.order = Order::Mdy,
                "dmy" => format.order = Order::Dmy,
                "ymd" => format.order = Order::Ymd,
                _ => {}
            }
        }
        format.interval_style = match interval_style {
            "postgres_verbose" => IntervalStyle::PostgresVerbose,
            "sql_standard" => IntervalStyle::SqlStandard,
            "iso_8601" => IntervalStyle::Iso8601,
            _ => IntervalStyle::Postgres,
        };
        if let Some(zone) = Zone::named(timezone) {
            format.zone = zone;
        }
        format
    }
}

thread_local! {
    /// FORMAT is the running session's output formats, which each connection's thread installs for its statements.
    static FORMAT: std::cell::RefCell<Format> = std::cell::RefCell::new(Format::default());
    /// NOW is the running transaction's start, which `now` and `today` read as.
    static NOW: std::cell::Cell<i64> = const { std::cell::Cell::new(0) };
}

/// install_format makes the formats the ones values on this thread print with.
pub fn install_format(format: Format) {
    FORMAT.with(|f| *f.borrow_mut() = format);
}

/// install_now makes the timestamp the transaction start that values on this thread read `now` as.
pub fn install_now(timestamp: i64) {
    NOW.with(|n| n.set(timestamp));
}

/// now returns the installed transaction start.
pub fn now() -> Now {
    Now { timestamp: NOW.with(|n| n.get()) }
}

/// clock returns the current time as a UTC timestamp.
pub fn clock() -> i64 {
    let micros = std::time::SystemTime::now().duration_since(std::time::UNIX_EPOCH).map_or(0, |d| d.as_micros() as i64);
    micros + UNIX_EPOCH_DAYS * USECS_PER_DAY
}

/// with_format runs a function with the installed formats.
pub fn with_format<R>(f: impl FnOnce(&Format) -> R) -> R {
    FORMAT.with(|format| f(&format.borrow()))
}

/// MONTHS and DAYS are the English month and weekday abbreviations Postgres prints and reads.
const MONTHS: [&str; 12] = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
const DAYS: [&str; 7] = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

/// year_text returns a year as Postgres prints it, with the BC suffix separate.
fn year_text(year: i64) -> (String, bool) {
    if year > 0 { (format!("{year:04}"), false) } else { (format!("{:04}", -(year - 1)), true) }
}

/// seconds_text prints seconds with a fraction trimmed of trailing zeros.
fn seconds_text(seconds: i64, micros: i64) -> String {
    if micros == 0 {
        return format!("{seconds:02}");
    }
    let fraction = format!("{micros:06}");
    format!("{seconds:02}.{}", fraction.trim_end_matches('0'))
}

/// offset_text prints an offset east of UTC as Postgres' EncodeTimezone does: `+05`, `+05:30`, or `+05:30:15`.
pub fn offset_text(seconds: i32) -> String {
    let sign = if seconds < 0 { '-' } else { '+' };
    let abs = seconds.abs();
    let (h, m, s) = (abs / 3600, abs / 60 % 60, abs % 60);
    match (m, s) {
        (0, 0) => format!("{sign}{h:02}"),
        (_, 0) => format!("{sign}{h:02}:{m:02}"),
        _ => format!("{sign}{h:02}:{m:02}:{s:02}"),
    }
}

/// format_date prints a date in the DateStyle.
pub fn format_date(days: i32, format: &Format) -> String {
    match days {
        DATE_NOBEGIN => return "-infinity".into(),
        DATE_NOEND => return "infinity".into(),
        _ => {}
    }
    let (y, m, d) = j2date(days as i64 + POSTGRES_EPOCH_JDATE);
    let (year, bc) = year_text(y);
    let date = match format.style {
        Style::Iso => format!("{year}-{m:02}-{d:02}"),
        Style::Sql if format.order == Order::Dmy => format!("{d:02}/{m:02}/{year}"),
        Style::Sql => format!("{m:02}/{d:02}/{year}"),
        Style::German => format!("{d:02}.{m:02}.{year}"),
        Style::Postgres if format.order == Order::Dmy => format!("{d:02}-{m:02}-{year}"),
        Style::Postgres => format!("{m:02}-{d:02}-{year}"),
    };
    if bc { format!("{date} BC") } else { date }
}

/// format_timestamp prints a timestamp in the DateStyle, with its zone when it has one: the offset east of UTC and
/// the zone's abbreviation.
pub fn format_timestamp(ts: i64, zone: Option<(i32, &str)>, format: &Format) -> String {
    match ts {
        TIMESTAMP_NOBEGIN => return "-infinity".into(),
        TIMESTAMP_NOEND => return "infinity".into(),
        _ => {}
    }
    let local = ts + zone.map_or(0, |(offset, _)| offset as i64 * USECS_PER_SEC);
    let f = fields_of_timestamp(local);
    let (year, bc) = year_text(f.year);
    let time = format!("{:02}:{:02}:{}", f.hour, f.minute, seconds_text(f.second, f.micros));
    let mut out = match format.style {
        Style::Iso => {
            let mut s = format!("{year}-{:02}-{:02} {time}", f.month, f.day);
            if let Some((offset, _)) = zone {
                s.push_str(&offset_text(offset));
            }
            s
        }
        Style::Sql | Style::German => {
            let date = match (format.style, format.order) {
                (Style::German, _) => format!("{:02}.{:02}.{year}", f.day, f.month),
                (_, Order::Dmy) => format!("{:02}/{:02}/{year}", f.day, f.month),
                _ => format!("{:02}/{:02}/{year}", f.month, f.day),
            };
            let mut s = format!("{date} {time}");
            if let Some((_, name)) = zone {
                s.push(' ');
                s.push_str(name);
            }
            s
        }
        Style::Postgres => {
            let dow = DAYS[((f_julian(&f) + 1).rem_euclid(7)) as usize];
            let month = MONTHS[f.month as usize - 1];
            let mut s = if format.order == Order::Dmy {
                format!("{dow} {:02} {month} {time} {year}", f.day)
            } else {
                format!("{dow} {month} {:02} {time} {year}", f.day)
            };
            if let Some((_, name)) = zone {
                s.push(' ');
                s.push_str(name);
            }
            s
        }
    };
    if bc {
        out.push_str(" BC");
    }
    out
}

/// f_julian returns the Julian day of fields' date.
fn f_julian(f: &Fields) -> i64 {
    date2j(f.year, f.month, f.day)
}

/// format_time prints a time of day.
pub fn format_time(micros: i64) -> String {
    let (h, m, s, us) =
        (micros / USECS_PER_HOUR, micros / USECS_PER_MINUTE % 60, micros / USECS_PER_SEC % 60, micros % USECS_PER_SEC);
    format!("{h:02}:{m:02}:{}", seconds_text(s, us))
}

/// format_timetz prints a time of day with its zone, which Postgres stores as seconds west of UTC.
pub fn format_timetz(micros: i64, west: i32) -> String {
    format!("{}{}", format_time(micros), offset_text(-west))
}

/// interval_fields breaks an interval into years, months, days, hours, minutes, seconds, and microseconds.
fn interval_fields(iv: &Interval) -> (i64, i64, i64, i64, i64, i64, i64) {
    let years = iv.months as i64 / 12;
    let months = iv.months as i64 % 12;
    let time = iv.micros;
    let hours = time / USECS_PER_HOUR;
    let minutes = time / USECS_PER_MINUTE % 60;
    let seconds = time / USECS_PER_SEC % 60;
    let micros = time % USECS_PER_SEC;
    (years, months, iv.days as i64, hours, minutes, seconds, micros)
}

/// signed_seconds prints seconds and microseconds that share a sign, trimming the fraction.
fn signed_seconds(seconds: i64, micros: i64, pad: bool) -> String {
    let negative = seconds < 0 || micros < 0;
    let (s, us) = (seconds.abs(), micros.abs());
    let text = if us == 0 {
        if pad { format!("{s:02}") } else { s.to_string() }
    } else {
        let fraction = format!("{us:06}");
        let whole = if pad { format!("{s:02}") } else { s.to_string() };
        format!("{whole}.{}", fraction.trim_end_matches('0'))
    };
    if negative { format!("-{text}") } else { text }
}

/// format_interval prints an interval in the IntervalStyle, as Postgres' EncodeInterval does.
pub fn format_interval(iv: &Interval, style: IntervalStyle) -> String {
    let (year, mon, mday, hour, min, sec, fsec) = interval_fields(iv);
    match style {
        IntervalStyle::Postgres => {
            let mut out = String::new();
            let mut is_zero = true;
            let mut is_before = false;
            for (value, unit) in [(year, "year"), (mon, "mon"), (mday, "day")] {
                if value == 0 {
                    continue;
                }
                if !is_zero {
                    out.push(' ');
                }
                if is_before && value > 0 {
                    out.push('+');
                }
                out.push_str(&format!("{value} {unit}{}", if value != 1 { "s" } else { "" }));
                is_before = value < 0;
                is_zero = false;
            }
            if is_zero || hour != 0 || min != 0 || sec != 0 || fsec != 0 {
                let minus = hour < 0 || min < 0 || sec < 0 || fsec < 0;
                if !is_zero {
                    out.push(' ');
                }
                out.push_str(if minus {
                    "-"
                } else if is_before {
                    "+"
                } else {
                    ""
                });
                out.push_str(&format!(
                    "{:02}:{:02}:{}",
                    hour.abs(),
                    min.abs(),
                    signed_seconds(sec.abs(), fsec.abs(), true)
                ));
            }
            out
        }
        IntervalStyle::PostgresVerbose => {
            let mut parts = Vec::new();
            let mut is_before = false;
            for (value, unit) in [(year, "year"), (mon, "mon"), (mday, "day"), (hour, "hour"), (min, "min")] {
                if value != 0 {
                    is_before = value < 0;
                    parts.push(format!("{} {unit}{}", value.abs(), if value.abs() != 1 { "s" } else { "" }));
                }
            }
            if sec != 0 || fsec != 0 {
                if sec < 0 || fsec < 0 {
                    is_before = true;
                }
                let text = signed_seconds(sec.abs(), fsec.abs(), false);
                let plural = if sec.abs() != 1 || fsec != 0 { "s" } else { "" };
                parts.push(format!("{text} sec{plural}"));
            }
            if parts.is_empty() {
                return "@ 0".into();
            }
            let mut out = format!("@ {}", parts.join(" "));
            if is_before {
                out.push_str(" ago");
            }
            out
        }
        IntervalStyle::SqlStandard => {
            let has_negative = year < 0 || mon < 0 || mday < 0 || hour < 0 || min < 0 || sec < 0 || fsec < 0;
            let has_positive = year > 0 || mon > 0 || mday > 0 || hour > 0 || min > 0 || sec > 0 || fsec > 0;
            let has_year_month = year != 0 || mon != 0;
            let has_day_time = hour != 0 || min != 0 || sec != 0 || fsec != 0 || mday != 0;
            let has_day = mday != 0;
            let mixed = has_negative && has_positive;
            if !has_negative && !has_positive {
                return "0".into();
            }
            if !mixed && !(has_year_month && has_day_time) && !(has_year_month && has_day) {
                let sign = if has_negative { "-" } else { "" };
                if has_year_month {
                    return format!("{sign}{}-{}", year.abs(), mon.abs());
                }
                if has_day {
                    return format!(
                        "{sign}{} {}:{:02}:{}",
                        mday.abs(),
                        hour.abs(),
                        min.abs(),
                        signed_seconds(sec.abs(), fsec.abs(), true)
                    );
                }
                return format!(
                    "{sign}{}:{:02}:{}",
                    hour.abs(),
                    min.abs(),
                    signed_seconds(sec.abs(), fsec.abs(), true)
                );
            }
            let year_sign = if year < 0 || mon < 0 { "-" } else { "+" };
            let day_sign = if mday < 0 { "-" } else { "+" };
            let time_sign = if hour < 0 || min < 0 || sec < 0 || fsec < 0 { "-" } else { "+" };
            format!(
                "{year_sign}{}-{} {day_sign}{} {time_sign}{}:{:02}:{}",
                year.abs(),
                mon.abs(),
                mday.abs(),
                hour.abs(),
                min.abs(),
                signed_seconds(sec.abs(), fsec.abs(), true)
            )
        }
        IntervalStyle::Iso8601 => {
            if year == 0 && mon == 0 && mday == 0 && hour == 0 && min == 0 && sec == 0 && fsec == 0 {
                return "PT0S".into();
            }
            let mut out = "P".to_string();
            for (value, unit) in [(year, 'Y'), (mon, 'M'), (mday, 'D')] {
                if value != 0 {
                    out.push_str(&format!("{value}{unit}"));
                }
            }
            if hour != 0 || min != 0 || sec != 0 || fsec != 0 {
                out.push('T');
                for (value, unit) in [(hour, 'H'), (min, 'M')] {
                    if value != 0 {
                        out.push_str(&format!("{value}{unit}"));
                    }
                }
                if sec != 0 || fsec != 0 {
                    out.push_str(&format!("{}S", signed_seconds(sec, fsec, false)));
                }
            }
            out
        }
    }
}

/// Kind is the type a datetime value is read as, for its error messages.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    Date,
    Time,
    TimeTz,
    Timestamp,
    TimestampTz,
}

impl Kind {
    /// name returns the type's name in error messages.
    fn name(self) -> &'static str {
        match self {
            Kind::Date => "date",
            Kind::Time => "time",
            Kind::TimeTz => "time with time zone",
            Kind::Timestamp => "timestamp",
            Kind::TimestampTz => "timestamp with time zone",
        }
    }
}

/// invalid returns Postgres' error for text that is not a value of the kind.
fn invalid(kind: Kind, text: &str) -> PgError {
    PgError::new(code::INVALID_DATETIME_FORMAT, format!("invalid input syntax for type {}: \"{text}\"", kind.name()))
}

/// out_of_range returns Postgres' error for a field out of range, suggesting another DateStyle when asked.
fn out_of_range(text: &str, hint: bool) -> PgError {
    PgError {
        hint: hint.then(|| "Perhaps you need a different \"datestyle\" setting.".to_string()),
        ..PgError::new(code::DATETIME_FIELD_OVERFLOW, format!("date/time field value out of range: \"{text}\""))
    }
}

/// Parsed is what reading a datetime found: special values, the date and time fields, and a zone.
#[derive(Debug, Default)]
struct Parsed {
    special: Option<&'static str>,
    year: Option<i64>,
    month: Option<i64>,
    day: Option<i64>,
    hour: Option<i64>,
    minute: i64,
    second: i64,
    micros: i64,
    bc: bool,
    pm: Option<bool>,
    /// The zone's offset east of UTC, or a named zone whose offset depends on the date.
    offset: Option<i32>,
    zone: Option<Zone>,
    julian: Option<i64>,
    /// Whether a month name set the month.
    text_month: bool,
    /// Whether the year had two digits or fewer, which makes it a year near 2000.
    two_digit: bool,
}

impl Parsed {
    /// decode_number sets the next date field to a number, as Postgres' DecodeNumber chooses it from the fields
    /// already set, whether a month name set the month, and the DateStyle order, failing when the date is complete.
    fn decode_number(&mut self, n: i64, digits: usize, order: Order) -> bool {
        let year = |p: &mut Parsed| {
            p.year = Some(n);
            p.two_digit = digits <= 2;
        };
        match (self.year.is_some(), self.month.is_some(), self.day.is_some()) {
            (false, false, false) if digits >= 3 || order == Order::Ymd => year(self),
            (false, false, false) if order == Order::Dmy => self.day = Some(n),
            (false, false, false) | (true, false, false) | (false, false, true) => self.month = Some(n),
            (false, true, false) if self.text_month && (digits >= 3 || order == Order::Ymd) => year(self),
            (false, true, false) => self.day = Some(n),
            (true, true, false) if self.text_month && digits >= 3 && self.two_digit => {
                self.day = self.year;
                self.year = Some(n);
                self.two_digit = false;
            }
            (true, true, false) => self.day = Some(n),
            (false, true, true) => year(self),
            _ => return false,
        }
        true
    }
}

/// month_number returns the number of a month name or its three-letter abbreviation.
fn month_number(word: &str) -> Option<i64> {
    const NAMES: [&str; 12] = [
        "january",
        "february",
        "march",
        "april",
        "may",
        "june",
        "july",
        "august",
        "september",
        "october",
        "november",
        "december",
    ];
    let w = word.to_ascii_lowercase();
    NAMES.iter().position(|n| *n == w || n[..3] == w || (w == "sept" && *n == "september")).map(|i| i as i64 + 1)
}

/// is_weekday reports whether a word names a day of the week, in full or abbreviated.
fn is_weekday(word: &str) -> bool {
    const NAMES: [&str; 7] = ["sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"];
    let w = word.to_ascii_lowercase();
    NAMES.iter().any(|n| *n == w || n[..3] == w) || matches!(w.as_str(), "tues" | "thur" | "thurs")
}

/// parse_time_field reads `HH:MM`, `HH:MM:SS[.ffffff]`, or `MM:SS.ffffff` into hours, minutes, seconds, and
/// microseconds.
fn parse_time_field(text: &str) -> Option<(i64, i64, i64, i64)> {
    let parts: Vec<&str> = text.split(':').collect();
    let int =
        |s: &str| (!s.is_empty() && s.bytes().all(|b| b.is_ascii_digit())).then(|| s.parse::<i64>().ok()).flatten();
    let seconds = |s: &str| -> Option<(i64, i64)> {
        let (whole, fraction) = s.split_once('.').unwrap_or((s, ""));
        fraction.bytes().all(|b| b.is_ascii_digit()).then_some(())?;
        Some((int(whole)?, fraction_micros(fraction)))
    };
    match parts.as_slice() {
        [m, s] if s.contains('.') => {
            let (second, micros) = seconds(s)?;
            Some((0, int(m)?, second, micros))
        }
        [h, m] => Some((int(h)?, int(m)?, 0, 0)),
        [h, m, s] => {
            let (second, micros) = seconds(s)?;
            Some((int(h)?, int(m)?, second, micros))
        }
        _ => None,
    }
}

/// fraction_micros rounds the digits of a fraction to microseconds.
fn fraction_micros(fraction: &str) -> i64 {
    if fraction.is_empty() {
        return 0;
    }
    let digits: String = fraction.chars().take(7).collect();
    let padded = format!("{digits:0<7}");
    let n: i64 = padded.parse().unwrap_or(0);
    (n + 5) / 10
}

/// tokens splits datetime text into words, keeping signed numbers and zone offsets together.
fn tokens(text: &str) -> Vec<String> {
    let mut out = Vec::new();
    let mut current = String::new();
    let chars: Vec<char> = text.chars().collect();
    let mut i = 0;
    while i < chars.len() {
        let c = chars[i];
        if c.is_whitespace() || c == ',' {
            if !current.is_empty() {
                out.push(std::mem::take(&mut current));
            }
        } else if (c == '+' || c == '-')
            && !current.is_empty()
            && (current.contains(':')
                || c == '+'
                || (current.matches('-').count() == 2 && current.starts_with(|d: char| d.is_ascii_digit())))
        {
            // A zone offset right after a time, as in 12:00:00-08.
            out.push(std::mem::take(&mut current));
            current.push(c);
        } else if c == 'T'
            && !current.is_empty()
            && current.chars().last().is_some_and(|l| l.is_ascii_digit())
            && chars.get(i + 1).is_some_and(|n| n.is_ascii_digit())
            && current.contains('-')
        {
            out.push(std::mem::take(&mut current));
        } else if c == 'Z' && !current.is_empty() && current.contains(':') && i + 1 == chars.len() {
            out.push(std::mem::take(&mut current));
            current.push('Z');
        } else if (c.is_ascii_alphabetic()
            && !current.is_empty()
            && (current.bytes().all(|b| b.is_ascii_digit())
                || (current.contains(':') && current.starts_with(|d: char| d.is_ascii_digit()))))
            || (c.is_ascii_digit()
                && current.chars().all(|l| l.is_ascii_alphabetic())
                && (month_number(&current).is_some() || is_weekday(&current)))
        {
            out.push(std::mem::take(&mut current));
            current.push(c);
        } else {
            current.push(c);
        }
        i += 1;
    }
    if !current.is_empty() {
        out.push(current);
    }
    out
}

/// parse_datetime reads the fields of a date, time, or timestamp in Postgres' input formats.
fn parse_datetime(text: &str, kind: Kind, order: Order) -> Result<Parsed> {
    let mut p = Parsed::default();
    let words = tokens(text.trim());
    if words.is_empty() {
        return Err(invalid(kind, text));
    }
    for (index, word) in words.iter().enumerate() {
        let lower = word.to_ascii_lowercase();
        match lower.as_str() {
            "epoch" | "infinity" | "-infinity" | "+infinity" | "now" | "today" | "tomorrow" | "yesterday"
            | "allballs" => {
                p.special = Some(match lower.as_str() {
                    "epoch" => "epoch",
                    "infinity" | "+infinity" => "infinity",
                    "-infinity" => "-infinity",
                    "now" => "now",
                    "today" => "today",
                    "tomorrow" => "tomorrow",
                    "yesterday" => "yesterday",
                    _ => "allballs",
                });
                continue;
            }
            "bc" => {
                p.bc = true;
                continue;
            }
            "ad" => continue,
            "am" | "a.m." => {
                p.pm = Some(false);
                continue;
            }
            "pm" | "p.m." => {
                p.pm = Some(true);
                continue;
            }
            "z" | "zulu" | "utc" | "gmt" | "ut" => {
                p.offset = Some(0);
                continue;
            }
            "at" | "on" => continue,
            _ => {}
        }
        if let Some(rest) = lower.strip_prefix('j')
            && !rest.is_empty()
            && rest.bytes().all(|b| b.is_ascii_digit())
        {
            p.julian = rest.parse().ok();
            continue;
        }
        if word.contains(':') && word.chars().next().is_some_and(|c| c.is_ascii_digit()) {
            let (h, m, s, us) = parse_time_field(word).ok_or_else(|| invalid(kind, text))?;
            p.hour = Some(h);
            p.minute = m;
            p.second = s;
            p.micros = us;
            continue;
        }
        if (word.starts_with('+') || word.starts_with('-')) && index > 0 {
            if let Some(offset) = parse_offset(word) {
                p.offset = Some(offset);
                continue;
            }
            return Err(invalid(kind, text));
        }
        if let Some(m) = month_number(word) {
            if let (Some(numeric), false, None) = (p.month, p.text_month, p.day)
                && (1..=31).contains(&numeric)
            {
                p.day = Some(numeric);
            }
            p.text_month = true;
            p.month = Some(m);
            continue;
        }
        if is_weekday(word) {
            continue;
        }
        if let Some((whole, fraction)) = word.split_once('.')
            && whole.len() == 6
            && whole.bytes().chain(fraction.bytes()).all(|b| b.is_ascii_digit())
            && p.hour.is_none()
            && (matches!(kind, Kind::Time | Kind::TimeTz) || p.day.is_some())
        {
            let field = format!("{}:{}:{}.{fraction}", &whole[..2], &whole[2..4], &whole[4..]);
            let (h, m, s, us) = parse_time_field(&field).ok_or_else(|| invalid(kind, text))?;
            (p.hour, p.minute, p.second, p.micros) = (Some(h), m, s, us);
            continue;
        }
        if word.starts_with(|c: char| c.is_ascii_alphabetic())
            && let Some(zone) = Zone::named(word)
        {
            p.zone = Some(zone);
            continue;
        }
        if word.contains(['-', '/', '.']) && word.chars().next().is_some_and(|c| c.is_ascii_alphanumeric()) {
            let parts: Vec<&str> = word.split(['-', '/', '.']).collect();
            if parts.len() == 3 && parts.iter().all(|s| !s.is_empty()) {
                if let Some(m) = parts.iter().find_map(|s| month_number(s)) {
                    p.text_month = true;
                    p.month = Some(m);
                }
                for part in parts.iter().filter(|s| month_number(s).is_none()) {
                    let n: i64 = part.parse().map_err(|_| invalid(kind, text))?;
                    if !p.decode_number(n, part.len(), order) {
                        return Err(invalid(kind, text));
                    }
                }
                continue;
            }
            if parts.len() == 2 && order != Order::Ymd {
                // Month and day without a year is not a complete date.
                return Err(invalid(kind, text));
            }
        }
        if word.bytes().all(|b| b.is_ascii_digit()) {
            if p.year.is_none() && p.month.is_none() && word.len() == 8 {
                p.year = word[..4].parse().ok();
                p.month = word[4..6].parse().ok();
                p.day = word[6..].parse().ok();
                continue;
            }
            let time_only = matches!(kind, Kind::Time | Kind::TimeTz);
            if p.year.is_none() && p.month.is_none() && word.len() == 6 && words.len() == 1 && !time_only {
                p.year = Some(two_digit_year(word[..2].parse().unwrap_or(0)));
                p.month = word[2..4].parse().ok();
                p.day = word[4..].parse().ok();
                continue;
            }
            if p.hour.is_none() && ((p.year.is_some() && p.day.is_some()) || time_only) && matches!(word.len(), 4 | 6) {
                p.hour = word[..2].parse().ok();
                p.minute = word[2..4].parse().unwrap_or(0);
                p.second = word.get(4..6).and_then(|s| s.parse().ok()).unwrap_or(0);
                continue;
            }
            if !p.decode_number(word.parse().map_err(|_| invalid(kind, text))?, word.len(), order) {
                return Err(invalid(kind, text));
            }
            continue;
        }
        if word.contains('.') && word.bytes().all(|b| b.is_ascii_digit() || b == b'.') && p.hour.is_some() {
            return Err(invalid(kind, text));
        }
        if let Some(zone) = Zone::named(word) {
            p.zone = Some(zone);
            continue;
        }
        return Err(invalid(kind, text));
    }
    if p.two_digit && !p.bc {
        p.year = p.year.map(two_digit_year);
    }
    Ok(p)
}

/// two_digit_year expands a two-digit year to the nearest century year, as Postgres does: 70 to 99 are 19xx.
fn two_digit_year(year: i64) -> i64 {
    if year < 70 {
        year + 2000
    } else if year < 100 {
        year + 1900
    } else {
        year
    }
}

/// resolve_date returns the Julian day of parsed date fields, failing when they are missing or out of range.
fn resolve_date(p: &Parsed, kind: Kind, text: &str, order: Order) -> Result<i64> {
    if let Some(julian) = p.julian {
        return Ok(julian);
    }
    let (Some(mut year), Some(month), Some(day)) = (p.year, p.month, p.day) else { return Err(invalid(kind, text)) };
    if p.bc {
        if year <= 0 {
            return Err(out_of_range(text, false));
        }
        year = -(year - 1);
    }
    if !(1..=12).contains(&month) {
        return Err(out_of_range(text, order != Order::Ymd && (1..=12).contains(&day)));
    }
    if day < 1 || day > days_in_month(year, month) {
        return Err(out_of_range(text, false));
    }
    let julian = date2j(year, month, day);
    if !date_in_range(julian) {
        return Err(PgError::new(code::DATETIME_FIELD_OVERFLOW, format!("date out of range: \"{text}\"")));
    }
    Ok(julian)
}

/// resolve_time returns the microseconds of parsed time fields, applying AM and PM.
fn resolve_time(p: &Parsed, text: &str) -> Result<i64> {
    let mut hour = p.hour.unwrap_or(0);
    match p.pm {
        Some(true) if hour < 12 => hour += 12,
        Some(false) if hour == 12 => hour = 0,
        _ => {}
    }
    if p.pm.is_some() && !(0..=12).contains(&p.hour.unwrap_or(0)) {
        return Err(out_of_range(text, false));
    }
    let micros = hour * USECS_PER_HOUR + p.minute * USECS_PER_MINUTE + p.second * USECS_PER_SEC + p.micros;
    if hour < 0 || p.minute > 59 || p.second > 60 || micros > USECS_PER_DAY {
        return Err(out_of_range(text, false));
    }
    Ok(micros)
}

/// Now is the clock readings that special values like `now` and `today` resolve against.
#[derive(Clone, Copy, Debug)]
pub struct Now {
    /// The transaction's start as a UTC timestamp.
    pub timestamp: i64,
}

/// parse_date reads a date.
pub fn parse_date(text: &str, format: &Format, now: Now) -> Result<i32> {
    let p = parse_datetime(text, Kind::Date, format.order)?;
    match p.special {
        Some("infinity") => return Ok(DATE_NOEND),
        Some("-infinity") => return Ok(DATE_NOBEGIN),
        Some("epoch") => return Ok(UNIX_EPOCH_DAYS as i32),
        Some(special @ ("now" | "today" | "tomorrow" | "yesterday")) => {
            let local = now.timestamp + format.zone.offset_at(now.timestamp).0 as i64 * USECS_PER_SEC;
            let today = local.div_euclid(USECS_PER_DAY);
            let shift = match special {
                "tomorrow" => 1,
                "yesterday" => -1,
                _ => 0,
            };
            return Ok((today + shift) as i32);
        }
        Some(_) => return Err(invalid(Kind::Date, text)),
        None => {}
    }
    let julian = resolve_date(&p, Kind::Date, text, format.order)?;
    Ok((julian - POSTGRES_EPOCH_JDATE) as i32)
}

/// parse_timestamp reads a timestamp, or with a zone a timestamptz as a UTC timestamp, where text without a zone is
/// in the session's zone.
pub fn parse_timestamp(text: &str, with_zone: bool, format: &Format, now: Now) -> Result<i64> {
    let kind = if with_zone { Kind::TimestampTz } else { Kind::Timestamp };
    let p = parse_datetime(text, kind, format.order)?;
    let zone_local = |local: i64, p: &Parsed| -> i64 {
        if !with_zone {
            return local;
        }
        let offset = match (&p.offset, &p.zone) {
            (Some(offset), _) => *offset,
            (None, Some(zone)) => zone.offset_for_local(local),
            (None, None) => format.zone.offset_for_local(local),
        };
        local - offset as i64 * USECS_PER_SEC
    };
    match p.special {
        Some("infinity") => return Ok(TIMESTAMP_NOEND),
        Some("-infinity") => return Ok(TIMESTAMP_NOBEGIN),
        Some("epoch") => return Ok(UNIX_EPOCH_DAYS * USECS_PER_DAY),
        Some("now") => {
            if with_zone {
                return Ok(now.timestamp);
            }
            return Ok(now.timestamp + format.zone.offset_at(now.timestamp).0 as i64 * USECS_PER_SEC);
        }
        Some(special @ ("today" | "tomorrow" | "yesterday")) => {
            let local = now.timestamp + format.zone.offset_at(now.timestamp).0 as i64 * USECS_PER_SEC;
            let shift = match special {
                "tomorrow" => 1,
                "yesterday" => -1,
                _ => 0,
            };
            let day = local.div_euclid(USECS_PER_DAY) + shift;
            let time = if p.hour.is_some() { resolve_time(&p, text)? } else { 0 };
            return Ok(zone_local(day * USECS_PER_DAY + time, &p));
        }
        Some(_) => return Err(invalid(kind, text)),
        None => {}
    }
    let julian = resolve_date(&p, kind, text, format.order)?;
    let time = resolve_time(&p, text)?;
    let local = (julian - POSTGRES_EPOCH_JDATE) * USECS_PER_DAY + time;
    let ts = zone_local(local, &p);
    if !(MIN_TIMESTAMP..END_TIMESTAMP).contains(&ts) {
        return Err(PgError::new(code::DATETIME_FIELD_OVERFLOW, format!("timestamp out of range: \"{text}\"")));
    }
    Ok(ts)
}

/// parse_time reads a time of day, ignoring any date and zone.
pub fn parse_time(text: &str, format: &Format) -> Result<i64> {
    let p = parse_datetime(text, Kind::Time, format.order)?;
    match p.special {
        Some("allballs") => return Ok(0),
        Some(_) | None if p.hour.is_none() && p.special.is_none() => return Err(invalid(Kind::Time, text)),
        _ => {}
    }
    resolve_time(&p, text)
}

/// parse_timetz reads a time of day with a zone, returning the time and the zone in seconds west of UTC.
pub fn parse_timetz(text: &str, format: &Format, now: Now) -> Result<(i64, i32)> {
    let p = parse_datetime(text, Kind::TimeTz, format.order)?;
    if p.special == Some("now") {
        let offset = format.zone.offset_at(now.timestamp).0;
        return Ok(((now.timestamp + offset as i64 * USECS_PER_SEC).rem_euclid(USECS_PER_DAY), -offset));
    }
    if p.hour.is_none() && p.special != Some("allballs") {
        return Err(invalid(Kind::TimeTz, text));
    }
    let time = if p.special == Some("allballs") { 0 } else { resolve_time(&p, text)? };
    let offset = match (&p.offset, &p.zone) {
        (Some(offset), _) => *offset,
        (None, Some(zone)) => zone.offset_at(now.timestamp).0,
        (None, None) => format.zone.offset_at(now.timestamp).0,
    };
    Ok((time, -offset))
}

/// interval_unit returns the canonical name of an interval unit that Postgres reads.
fn interval_unit(word: &str) -> Option<&'static str> {
    Some(match word {
        "microsecond" | "microseconds" | "us" | "usec" | "usecs" | "useconds" => "us",
        "millisecond" | "milliseconds" | "ms" | "msec" | "msecs" | "mseconds" => "ms",
        "second" | "seconds" | "s" | "sec" | "secs" => "s",
        "minute" | "minutes" | "m" | "min" | "mins" => "min",
        "hour" | "hours" | "h" | "hr" | "hrs" => "h",
        "day" | "days" | "d" => "d",
        "week" | "weeks" | "w" => "w",
        "month" | "months" | "mon" | "mons" => "mon",
        "year" | "years" | "y" | "yr" | "yrs" => "y",
        "decade" | "decades" | "dec" | "decs" => "dec",
        "century" | "centuries" | "c" | "cent" => "cent",
        "millennium" | "millennia" | "mil" | "mils" => "mil",
        _ => return None,
    })
}

/// IntervalBuilder accumulates interval fields, cascading fractions of larger units into smaller ones.
#[derive(Default)]
struct IntervalBuilder {
    years: f64,
    months: f64,
    days: f64,
    micros: f64,
}

impl IntervalBuilder {
    /// add adds an amount of a unit.
    fn add(&mut self, amount: f64, unit: &str) {
        match unit {
            "us" => self.micros += amount,
            "ms" => self.micros += amount * 1000.0,
            "s" => self.micros += amount * USECS_PER_SEC as f64,
            "min" => self.micros += amount * USECS_PER_MINUTE as f64,
            "h" => self.micros += amount * USECS_PER_HOUR as f64,
            "d" => self.add_days(amount),
            "w" => self.add_days(amount * 7.0),
            "mon" => self.add_months(amount),
            "y" => self.add_years(amount),
            "dec" => self.add_years(amount * 10.0),
            "cent" => self.add_years(amount * 100.0),
            _ => self.add_years(amount * 1000.0),
        }
    }

    /// add_years adds years, cascading a fraction into months.
    fn add_years(&mut self, amount: f64) {
        let whole = amount.trunc();
        self.years += whole;
        self.add_months((amount - whole) * 12.0);
    }

    /// add_days adds days, cascading a fraction into microseconds.
    fn add_days(&mut self, amount: f64) {
        let whole = amount.trunc();
        self.days += whole;
        self.micros += (amount - whole) * USECS_PER_DAY as f64;
    }

    /// add_months adds months, cascading a fraction into days of 30.
    fn add_months(&mut self, amount: f64) {
        let whole = amount.trunc();
        self.months += whole;
        self.add_days((amount - whole) * 30.0);
    }

    /// build returns the interval of some text, rounding microseconds, failing as Postgres does when a field or the
    /// whole interval is out of range.
    fn build(&self, text: &str) -> Result<Interval> {
        let (years, months, days, micros) =
            (self.years.round(), self.months.round(), self.days.round(), self.micros.round());
        if [years, months, days].iter().any(|f| f.abs() > i32::MAX as f64) || micros.abs() >= i64::MAX as f64 {
            return Err(PgError::new(
                code::INTERVAL_FIELD_OVERFLOW,
                format!("interval field value out of range: \"{text}\""),
            ));
        }
        let total = years * 12.0 + months;
        if total.abs() > i32::MAX as f64 {
            return Err(PgError::new(code::DATETIME_FIELD_OVERFLOW, "interval out of range"));
        }
        Ok(Interval { months: total as i32, days: days as i32, micros: micros as i64 })
    }
}

/// parse_interval reads an interval in Postgres' formats: unit words, `@` verbose form with `ago`, times, SQL
/// standard year-month and day-time, and ISO 8601.
pub fn parse_interval(text: &str) -> Result<Interval> {
    let invalid =
        || PgError::new(code::INVALID_DATETIME_FORMAT, format!("invalid input syntax for type interval: \"{text}\""));
    let trimmed = text.trim();
    if let Some(iso) = trimmed.strip_prefix('P').or_else(|| trimmed.strip_prefix('p')) {
        return parse_iso_interval(iso).ok_or_else(invalid);
    }
    let mut b = IntervalBuilder::default();
    let mut words: Vec<String> = trimmed.split_whitespace().map(str::to_ascii_lowercase).collect();
    let mut ago = false;
    if words.first().map(String::as_str) == Some("@") {
        words.remove(0);
    } else if let Some(first) = words.first_mut()
        && first.starts_with('@')
    {
        first.remove(0);
    }
    if words.last().map(String::as_str) == Some("ago") {
        words.pop();
        ago = true;
    }
    if words.is_empty() {
        return Err(invalid());
    }
    let mut i = 0;
    let mut seen_any = false;
    while i < words.len() {
        let word = &words[i];
        if word.contains(':') {
            let (sign, rest) = match word.as_bytes()[0] {
                b'-' => (-1.0, &word[1..]),
                b'+' => (1.0, &word[1..]),
                _ => (1.0, &word[..]),
            };
            let parts: Vec<&str> = rest.split(':').collect();
            let num = |s: &str| s.parse::<f64>().map_err(|_| invalid());
            match parts.len() {
                2 => {
                    b.add(sign * num(parts[0])?, "h");
                    b.add(sign * num(parts[1])?, "min");
                }
                3 => {
                    b.add(sign * num(parts[0])?, "h");
                    b.add(sign * num(parts[1])?, "min");
                    b.add(sign * num(parts[2])?, "s");
                }
                _ => return Err(invalid()),
            }
            seen_any = true;
            i += 1;
            continue;
        }
        // SQL standard year-month, like 1-2.
        if let Some((y, m)) = word.split_once('-').filter(|(y, m)| {
            !y.is_empty()
                && !m.is_empty()
                && y.bytes().all(|c| c.is_ascii_digit())
                && m.bytes().all(|c| c.is_ascii_digit())
        }) {
            b.add(y.parse::<f64>().map_err(|_| invalid())?, "y");
            b.add(m.parse::<f64>().map_err(|_| invalid())?, "mon");
            seen_any = true;
            i += 1;
            continue;
        }
        let (number_text, unit_text) = match word.find(|c: char| c.is_ascii_alphabetic()) {
            Some(p) if p > 0 => (&word[..p], Some(&word[p..])),
            _ => (&word[..], None),
        };
        let amount: f64 = number_text.parse().map_err(|_| invalid())?;
        let unit = match unit_text {
            Some(u) => Some(interval_unit(u).ok_or_else(invalid)?),
            None => match words.get(i + 1).and_then(|w| interval_unit(w)) {
                Some(u) => {
                    i += 1;
                    Some(u)
                }
                None => None,
            },
        };
        match unit {
            Some(u) => b.add(amount, u),
            // A bare number is days when a time follows, and seconds otherwise.
            None if words.get(i + 1).is_some_and(|w| w.contains(':')) => b.add(amount, "d"),
            None => b.add(amount, "s"),
        }
        seen_any = true;
        i += 1;
    }
    if !seen_any {
        return Err(invalid());
    }
    let mut iv = b.build(text)?;
    if ago {
        iv = Interval { months: -iv.months, days: -iv.days, micros: -iv.micros };
    }
    Ok(iv)
}

/// parse_iso_interval reads an ISO 8601 interval after its `P`.
fn parse_iso_interval(text: &str) -> Option<Interval> {
    let mut b = IntervalBuilder::default();
    let mut in_time = false;
    let mut number = String::new();
    for c in text.chars() {
        match c {
            'T' | 't' => in_time = true,
            '0'..='9' | '.' | '-' | '+' => number.push(c),
            _ => {
                let amount: f64 = number.parse().ok()?;
                number.clear();
                let unit = match (c.to_ascii_uppercase(), in_time) {
                    ('Y', false) => "y",
                    ('M', false) => "mon",
                    ('W', false) => "w",
                    ('D', false) => "d",
                    ('H', true) => "h",
                    ('M', true) => "min",
                    ('S', true) => "s",
                    _ => return None,
                };
                b.add(amount, unit);
            }
        }
    }
    if !number.is_empty() {
        return None;
    }
    b.build("").ok()
}

/// Go stores dates and timestamps as Go times, with its own times standing for the infinities.
pub mod go_time {
    /// UNIX_SECONDS_FROM_YEAR_ONE is the seconds from 0001-01-01 to 1970-01-01, the offset of Go's binary times.
    pub const UNIX_SECONDS_FROM_YEAR_ONE: i64 = 62_135_596_800;

    /// INFINITY and NEG_INFINITY are the Unix seconds and nanoseconds of the times Go uses for the infinities.
    pub const INFINITY: (i64, i32) = (9_224_318_016_000 - 1, 999_999_000);
    pub const NEG_INFINITY: (i64, i32) = (-210_866_803_200, 0);

    /// marshal returns a UTC time as Go's Time.MarshalBinary writes it: version 1, seconds from year one, nanoseconds,
    /// and the UTC zone marker.
    pub fn marshal(unix_seconds: i64, nanos: i32) -> Vec<u8> {
        let mut out = vec![1];
        out.extend_from_slice(&(unix_seconds + UNIX_SECONDS_FROM_YEAR_ONE).to_be_bytes());
        out.extend_from_slice(&nanos.to_be_bytes());
        out.extend_from_slice(&(-1i16).to_be_bytes());
        out
    }

    /// unmarshal reads Go's Time.MarshalBinary, returning the Unix seconds and nanoseconds of the instant.
    pub fn unmarshal(bytes: &[u8]) -> Option<(i64, i32)> {
        let version = *bytes.first()?;
        if !(version == 1 && bytes.len() == 15 || version == 2 && bytes.len() == 16) {
            return None;
        }
        let seconds = i64::from_be_bytes(bytes[1..9].try_into().ok()?);
        let nanos = i32::from_be_bytes(bytes[9..13].try_into().ok()?);
        Some((seconds - UNIX_SECONDS_FROM_YEAR_ONE, nanos))
    }
}

/// timestamp_to_go returns a timestamp as the Unix seconds and nanoseconds of Go's time, mapping the infinities.
pub fn timestamp_to_go(ts: i64) -> (i64, i32) {
    match ts {
        TIMESTAMP_NOEND => go_time::INFINITY,
        TIMESTAMP_NOBEGIN => go_time::NEG_INFINITY,
        _ => {
            let unix = ts + (-UNIX_EPOCH_DAYS) * USECS_PER_DAY;
            (unix.div_euclid(USECS_PER_SEC), (unix.rem_euclid(USECS_PER_SEC) * 1000) as i32)
        }
    }
}

/// timestamp_from_go returns the timestamp of Go's time, mapping the infinities.
pub fn timestamp_from_go(seconds: i64, nanos: i32) -> i64 {
    if (seconds, nanos) >= go_time::INFINITY {
        return TIMESTAMP_NOEND;
    }
    if (seconds, nanos) <= go_time::NEG_INFINITY {
        return TIMESTAMP_NOBEGIN;
    }
    (seconds - (-UNIX_EPOCH_DAYS) * 86_400) * USECS_PER_SEC + nanos as i64 / 1000
}

#[cfg(test)]
mod tests {
    use super::*;

    /// now is a fixed clock for tests.
    fn now() -> Now {
        Now { timestamp: 0 }
    }

    #[test]
    fn calendar_math_matches_postgres() {
        assert_eq!(date2j(2000, 1, 1), POSTGRES_EPOCH_JDATE);
        for jd in [0, 1_721_426, 2_440_588, 2_451_545, 5_373_484] {
            let (y, m, d) = j2date(jd);
            assert_eq!(date2j(y, m, d), jd);
        }
        assert_eq!(j2date(2_440_588), (1970, 1, 1));
    }

    #[test]
    fn dates_and_timestamps_round_trip() {
        let f = Format::default();
        let date = |t: &str| format_date(parse_date(t, &f, now()).unwrap(), &f);
        assert_eq!(date("2020-02-29"), "2020-02-29");
        assert_eq!(date("January 8, 1999"), "1999-01-08");
        assert_eq!(date("1/8/1999"), "1999-01-08");
        assert_eq!(date("19990108"), "1999-01-08");
        assert_eq!(date("0001-01-01 BC"), "0001-01-01 BC");
        assert_eq!(date("infinity"), "infinity");
        assert!(parse_date("2020-02-30", &f, now()).is_err());
        let ts = |t: &str| format_timestamp(parse_timestamp(t, false, &f, now()).unwrap(), None, &f);
        assert_eq!(ts("2020-01-02 03:04:05.5"), "2020-01-02 03:04:05.5");
        assert_eq!(ts("2020-01-02T03:04:05"), "2020-01-02 03:04:05");
        assert_eq!(ts("2020-01-02 03:04 PM"), "2020-01-02 15:04:00");
        let tz = |t: &str| {
            let utc = parse_timestamp(t, true, &f, now()).unwrap();
            format_timestamp(utc, Some((0, "UTC")), &f)
        };
        assert_eq!(tz("2020-01-02 03:04:05-08"), "2020-01-02 11:04:05+00");
        assert_eq!(tz("2020-01-02 03:04:05+05:30"), "2020-01-01 21:34:05+00");
        let la = Format { zone: Zone::named("America/Los_Angeles").unwrap(), ..Format::default() };
        let utc = parse_timestamp("2020-07-01 12:00:00", true, &la, now()).unwrap();
        let (offset, name) = la.zone.offset_at(utc);
        assert_eq!(format_timestamp(utc, Some((offset, &name)), &la), "2020-07-01 12:00:00-07");
    }

    #[test]
    fn intervals_read_and_print_as_postgres_does() {
        let iv = |t: &str| format_interval(&parse_interval(t).unwrap(), IntervalStyle::Postgres);
        assert_eq!(iv("1 year 2 months 3 days 04:05:06.789"), "1 year 2 mons 3 days 04:05:06.789");
        assert_eq!(iv("1.5 days"), "1 day 12:00:00");
        assert_eq!(iv("-1 days +2 hours"), "-1 days +02:00:00");
        assert_eq!(iv("@ 1 hour ago"), "-01:00:00");
        assert_eq!(iv("P1Y2M3DT4H5M6S"), "1 year 2 mons 3 days 04:05:06");
        assert_eq!(iv("0"), "00:00:00");
        let i = parse_interval("1 year 2 mons 3 days 04:05:06").unwrap();
        assert_eq!(format_interval(&i, IntervalStyle::Iso8601), "P1Y2M3DT4H5M6S");
        assert_eq!(format_interval(&i, IntervalStyle::SqlStandard), "+1-2 +3 +4:05:06");
        assert_eq!(format_interval(&i, IntervalStyle::PostgresVerbose), "@ 1 year 2 mons 3 days 4 hours 5 mins 6 secs");
    }

    #[test]
    fn go_times_map_the_infinities() {
        assert_eq!(timestamp_from_go(go_time::INFINITY.0, go_time::INFINITY.1), TIMESTAMP_NOEND);
        let ts = 123_456_789_012;
        let (s, n) = timestamp_to_go(ts);
        assert_eq!(timestamp_from_go(s, n), ts);
        assert_eq!(go_time::unmarshal(&go_time::marshal(s, n)), Some((s, n)));
    }
}
