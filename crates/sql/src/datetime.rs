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
    /// INFINITY is the interval that follows every other, which has every field at its largest value.
    pub const INFINITY: Interval = Interval { months: i32::MAX, days: i32::MAX, micros: i64::MAX };

    /// NEG_INFINITY is the interval that precedes every other, which has every field at its smallest value.
    pub const NEG_INFINITY: Interval = Interval { months: i32::MIN, days: i32::MIN, micros: i64::MIN };

    /// is_finite reports whether the interval is neither infinity.
    pub fn is_finite(&self) -> bool {
        *self != Interval::INFINITY && *self != Interval::NEG_INFINITY
    }

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

/// Zone is a time zone: a fixed offset east of UTC with its name, a zone of the IANA database, or an abbreviation
/// whose offset is the one it had in a zone at the time.
#[derive(Clone, Debug, PartialEq)]
pub enum Zone {
    Fixed { offset: i32, name: String },
    Tz(chrono_tz::Tz),
    Dynamic { name: String, tz: chrono_tz::Tz },
}

/// DYNAMIC_ABBREVIATIONS are the abbreviations of Postgres' default set whose offsets come from a zone, as its
/// timezonesets/Default file defines them.
const DYNAMIC_ABBREVIATIONS: &[(&str, &str)] = &[
    ("ART", "America/Argentina/Buenos_Aires"),
    ("ARST", "America/Argentina/Buenos_Aires"),
    ("CLT", "America/Santiago"),
    ("GYT", "America/Guyana"),
    ("PYT", "America/Asuncion"),
    ("VET", "America/Caracas"),
    ("DAVT", "Antarctica/Davis"),
    ("MAWT", "Antarctica/Mawson"),
    ("AMST", "Asia/Yerevan"),
    ("ANAST", "Asia/Anadyr"),
    ("ANAT", "Asia/Anadyr"),
    ("AZST", "Asia/Baku"),
    ("AZT", "Asia/Baku"),
    ("GEST", "Asia/Tbilisi"),
    ("GET", "Asia/Tbilisi"),
    ("IRKST", "Asia/Irkutsk"),
    ("IRKT", "Asia/Irkutsk"),
    ("KGT", "Asia/Bishkek"),
    ("KRAST", "Asia/Krasnoyarsk"),
    ("KRAT", "Asia/Krasnoyarsk"),
    ("LKT", "Asia/Colombo"),
    ("MAGST", "Asia/Magadan"),
    ("MAGT", "Asia/Magadan"),
    ("NOVST", "Asia/Novosibirsk"),
    ("NOVT", "Asia/Novosibirsk"),
    ("OMSST", "Asia/Omsk"),
    ("OMST", "Asia/Omsk"),
    ("PETST", "Asia/Kamchatka"),
    ("PETT", "Asia/Kamchatka"),
    ("SGT", "Asia/Singapore"),
    ("TMT", "Asia/Ashgabat"),
    ("ULAT", "Asia/Ulaanbaatar"),
    ("VLAST", "Asia/Vladivostok"),
    ("VLAT", "Asia/Vladivostok"),
    ("YAKST", "Asia/Yakutsk"),
    ("YAKT", "Asia/Yakutsk"),
    ("YEKT", "Asia/Yekaterinburg"),
    ("FKST", "Atlantic/Stanley"),
    ("FKT", "Atlantic/Stanley"),
    ("LHDT", "Australia/Lord_Howe"),
    ("MSK", "Europe/Moscow"),
    ("VOLT", "Europe/Volgograd"),
    ("IOT", "Indian/Chagos"),
    ("CKT", "Pacific/Rarotonga"),
    ("EASST", "Pacific/Easter"),
    ("EAST", "Pacific/Easter"),
    ("KOST", "Pacific/Kosrae"),
    ("LINT", "Pacific/Kiritimati"),
    ("NUT", "Pacific/Niue"),
    ("TKT", "Pacific/Fakaofo"),
];

/// SEARCH_STEP and SEARCH_STEPS bound the search for when a zone used an abbreviation: weekly steps over about two
/// centuries.
const SEARCH_STEP: i64 = 7 * USECS_PER_DAY;
const SEARCH_STEPS: i64 = 200 * 53;

/// zone_offset returns a zone's offset east of UTC and its abbreviation at a UTC timestamp.
fn zone_offset(tz: chrono_tz::Tz, utc: i64) -> (i32, String) {
    let offset = tz.offset_from_utc_datetime(&naive_of(utc));
    let seconds = offset.fix().local_minus_utc();
    let abbreviation =
        chrono_tz::OffsetName::abbreviation(&offset).map_or_else(|| numeric_name(seconds), str::to_string);
    (seconds, abbreviation)
}

/// abbreviation_offset returns the offset that an abbreviation had in a zone at a UTC timestamp, from the latest
/// period of the zone that used it at or before then, or else the earliest one after, or the zone's offset when it
/// never used it, as Postgres' pg_interpret_timezone_abbrev and its callers do.
fn abbreviation_offset(tz: chrono_tz::Tz, name: &str, utc: i64) -> i32 {
    let (offset, abbreviation) = zone_offset(tz, utc);
    if abbreviation.eq_ignore_ascii_case(name) {
        return offset;
    }
    for direction in [-1, 1] {
        for step in 1..=SEARCH_STEPS {
            let (found, abbreviation) = zone_offset(tz, utc + direction * step * SEARCH_STEP);
            if abbreviation.eq_ignore_ascii_case(name) {
                return found;
            }
        }
    }
    offset
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

/// abbreviation_prefix returns the length of the longest time zone abbreviation that text starts with, ignoring case,
/// with its offset east of UTC, as Postgres' DecodeTimezoneAbbrevPrefix finds it.
pub fn abbreviation_prefix(text: &str) -> Option<(usize, i32)> {
    let lower = text.to_ascii_lowercase();
    let longest = lower.len().min(10);
    (1..=longest)
        .rev()
        .filter(|&n| lower.is_char_boundary(n))
        .find_map(|n| abbreviations().get(&lower[..n]).map(|&(offset, _)| (n, offset)))
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
        if let Some((name, tz)) = DYNAMIC_ABBREVIATIONS.iter().find(|(n, _)| n.eq_ignore_ascii_case(trimmed))
            && let Ok(tz) = tz.parse::<chrono_tz::Tz>()
        {
            return Some(Zone::Dynamic { name: name.to_string(), tz });
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
            Zone::Tz(tz) => zone_offset(*tz, utc),
            Zone::Dynamic { name, tz } => (abbreviation_offset(*tz, name, utc), name.clone()),
        }
    }

    /// offset_for_local returns the zone's offset east of UTC for a local timestamp, preferring the earlier offset in
    /// a spring-forward gap and the later one in a fall-back overlap, as Postgres does.
    pub fn offset_for_local(&self, local: i64) -> i32 {
        match self {
            Zone::Fixed { offset, .. } => *offset,
            Zone::Dynamic { name, tz } => {
                let utc = local - Zone::Tz(*tz).offset_for_local(local) as i64 * USECS_PER_SEC;
                abbreviation_offset(*tz, name, utc)
            }
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

/// write_digits appends a non-negative number with at least `width` digits, zero-padded.
fn write_digits(out: &mut Vec<u8>, value: i64, width: usize) {
    let mut buffer = itoa::Buffer::new();
    let digits = buffer.format(value);
    out.extend(std::iter::repeat_n(b'0', width.saturating_sub(digits.len())));
    out.extend_from_slice(digits.as_bytes());
}

/// write_iso appends a date of the common era, and with a time its time of day and zone offset, in the ISO DateStyle,
/// reporting false without writing when the value needs another style or is infinite or before the common era.
pub fn write_iso(out: &mut Vec<u8>, days: i64, time: Option<(i64, Option<i32>)>) -> bool {
    let (year, month, day) = j2date(days + POSTGRES_EPOCH_JDATE);
    if year <= 0 || !with_format(|f| f.style == Style::Iso) {
        return false;
    }
    write_digits(out, year, 4);
    out.push(b'-');
    write_digits(out, month, 2);
    out.push(b'-');
    write_digits(out, day, 2);
    let Some((time, offset)) = time else { return true };
    out.push(b' ');
    write_digits(out, time / USECS_PER_HOUR, 2);
    out.push(b':');
    write_digits(out, time / USECS_PER_MINUTE % 60, 2);
    out.push(b':');
    write_digits(out, time / USECS_PER_SEC % 60, 2);
    let mut micros = time % USECS_PER_SEC;
    if micros != 0 {
        let mut width = 6;
        while micros % 10 == 0 {
            micros /= 10;
            width -= 1;
        }
        out.push(b'.');
        write_digits(out, micros, width);
    }
    if let Some(offset) = offset {
        out.extend_from_slice(offset_text(offset).as_bytes());
    }
    true
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
    if !iv.is_finite() {
        return if *iv == Interval::INFINITY { "infinity" } else { "-infinity" }.to_string();
    }
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
            let mut out = "@".to_string();
            let (mut is_zero, mut is_before) = (true, false);
            for (value, unit) in [(year, "year"), (mon, "mon"), (mday, "day"), (hour, "hour"), (min, "min")] {
                if value == 0 {
                    continue;
                }
                let value = if is_zero {
                    is_before = value < 0;
                    value.abs()
                } else if is_before {
                    -value
                } else {
                    value
                };
                out.push_str(&format!(" {value} {unit}{}", if value == 1 { "" } else { "s" }));
                is_zero = false;
            }
            if sec != 0 || fsec != 0 {
                out.push(' ');
                if sec < 0 || (sec == 0 && fsec < 0) {
                    if is_zero {
                        is_before = true;
                    } else if !is_before {
                        out.push('-');
                    }
                } else if is_before {
                    out.push('-');
                }
                let plural = if sec.abs() != 1 || fsec != 0 { "s" } else { "" };
                out.push_str(&format!("{} sec{plural}", signed_seconds(sec.abs(), fsec.abs(), false)));
                is_zero = false;
            }
            if is_zero {
                out.push_str(" 0");
            }
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
        hint: hint.then(|| "Perhaps you need a different \"DateStyle\" setting.".to_string()),
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
                || concatenated(&current)
                || (current.matches('-').count() == 2 && current.starts_with(|d: char| d.is_ascii_digit())))
        {
            // A zone offset right after a time, as in 12:00:00-08.
            out.push(std::mem::take(&mut current));
            current.push(c);
        } else if c == 'T'
            && !current.is_empty()
            && current.chars().last().is_some_and(|l| l.is_ascii_digit())
            && chars.get(i + 1).is_some_and(|n| n.is_ascii_digit())
            && (current.contains('-') || concatenated(&current))
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

/// concatenated reports whether a word is a run of six or more digits, perhaps after a Julian day's J or a time's T
/// and before a fraction, which Postgres reads as concatenated date or time fields or a Julian day.
fn concatenated(word: &str) -> bool {
    let digits = word.strip_prefix(['j', 'J', 't', 'T']).unwrap_or(word);
    let whole = digits.split_once('.').map_or(digits, |(whole, _)| whole);
    whole.len() >= 6 && digits.bytes().all(|b| b.is_ascii_digit() || b == b'.')
}

/// parse_datetime reads the fields of a date, time, or timestamp in Postgres' input formats.
fn parse_datetime(text: &str, kind: Kind, order: Order) -> Result<Parsed> {
    let mut p = Parsed::default();
    let words = tokens(text.trim());
    if words.is_empty() {
        return Err(invalid(kind, text));
    }
    for (index, word) in words.iter().enumerate() {
        let word = match word.strip_prefix(['t', 'T']) {
            Some(rest) if rest.starts_with(|c: char| c.is_ascii_digit()) => rest,
            _ => word.as_str(),
        };
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
        if let Some((year, day)) = word.split_once('.')
            && day.len() == 3
            && (year.len() + day.len() + 1 == word.len())
            && year.bytes().chain(day.bytes()).all(|b| b.is_ascii_digit())
            && p.year.is_none()
            && p.month.is_none()
            && p.day.is_none()
        {
            let (year, day): (i64, i64) = (year.parse().map_err(|_| invalid(kind, text))?, day.parse().unwrap_or(0));
            if !(1..=366).contains(&day) {
                return Err(out_of_range(text, false));
            }
            let (_, month, day) = j2date(date2j(year, 1, 1) + day - 1);
            (p.year, p.month, p.day) = (Some(year), Some(month), Some(day));
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
            if p.year.is_none() && p.month.is_none() && word.len() >= 8 {
                let year = word.len() - 4;
                p.year = word[..year].parse().ok();
                p.month = word[year..year + 2].parse().ok();
                p.day = word[year + 2..].parse().ok();
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

/// IntervalError is why interval input failed, as Postgres' DTERR_BAD_FORMAT and DTERR_FIELD_OVERFLOW say.
#[derive(Debug, PartialEq)]
enum IntervalError {
    BadFormat,
    Overflow,
}

/// IntervalParts is the fields interval input adds up, as Postgres' pg_itm_in holds them, each checked for overflow.
#[derive(Default)]
struct IntervalParts {
    micros: i64,
    days: i32,
    months: i32,
    years: i32,
}

/// Step is the result of adding to interval parts.
type Step = std::result::Result<(), IntervalError>;

impl IntervalParts {
    /// fract_micros adds a fraction of a unit of some microseconds, rounding as AdjustFractMicroseconds does.
    fn fract_micros(&mut self, frac: f64, scale: i64) -> Step {
        if frac == 0.0 {
            return Ok(());
        }
        let frac = frac * scale as f64;
        let mut usec = frac as i64;
        let rest = frac - usec as f64;
        if rest > 0.5 {
            usec += 1;
        } else if rest < -0.5 {
            usec -= 1;
        }
        self.micros.checked_add(usec).map(|m| self.micros = m).ok_or(IntervalError::Overflow)
    }

    /// fract_days adds a fraction of a unit of some days, as AdjustFractDays does.
    fn fract_days(&mut self, frac: f64, scale: i32) -> Step {
        if frac == 0.0 {
            return Ok(());
        }
        let frac = frac * scale as f64;
        let extra = frac as i32;
        self.days = self.days.checked_add(extra).ok_or(IntervalError::Overflow)?;
        self.fract_micros(frac - extra as f64, USECS_PER_DAY)
    }

    /// fract_years adds a fraction of a unit of some years as whole months, as AdjustFractYears does.
    fn fract_years(&mut self, frac: f64, scale: i32) -> Step {
        let extra = (frac * scale as f64 * 12.0).round_ties_even() as i32;
        self.months = self.months.checked_add(extra).ok_or(IntervalError::Overflow)?;
        Ok(())
    }

    /// add_micros adds a number and its fraction of a unit of some microseconds.
    fn add_micros(&mut self, val: i64, fval: f64, scale: i64) -> Step {
        self.micros = val.checked_mul(scale).and_then(|p| self.micros.checked_add(p)).ok_or(IntervalError::Overflow)?;
        self.fract_micros(fval, scale)
    }

    /// add_days adds a number of a unit of some days.
    fn add_days(&mut self, val: i64, scale: i32) -> Step {
        let days = i32::try_from(val).ok().and_then(|v| v.checked_mul(scale)).and_then(|d| self.days.checked_add(d));
        days.map(|d| self.days = d).ok_or(IntervalError::Overflow)
    }

    /// add_months adds a number of months.
    fn add_months(&mut self, val: i64) -> Step {
        let months = i32::try_from(val).ok().and_then(|v| self.months.checked_add(v));
        months.map(|m| self.months = m).ok_or(IntervalError::Overflow)
    }

    /// add_years adds a number of a unit of some years.
    fn add_years(&mut self, val: i64, scale: i32) -> Step {
        let years = i32::try_from(val).ok().and_then(|v| v.checked_mul(scale)).and_then(|y| self.years.checked_add(y));
        years.map(|y| self.years = y).ok_or(IntervalError::Overflow)
    }
}

/// FieldKind is the kind of a field of date and time input, as Postgres' ParseDateTime assigns them.
#[derive(Clone, Copy, PartialEq)]
enum FieldKind {
    Number,
    Date,
    Time,
    Text,
    Special,
    Zone,
}

/// DATE_WORDS are the words of Postgres' datetktbl, which a word followed by a digit or `+` must not be to start a
/// date field.
const DATE_WORDS: &[&str] = &[
    "-infinity",
    "ad",
    "allballs",
    "am",
    "apr",
    "april",
    "at",
    "aug",
    "august",
    "bc",
    "d",
    "dec",
    "december",
    "dow",
    "doy",
    "dst",
    "epoch",
    "feb",
    "february",
    "fri",
    "friday",
    "h",
    "infinity",
    "isodow",
    "isoyear",
    "j",
    "jan",
    "january",
    "jd",
    "jul",
    "julian",
    "july",
    "jun",
    "june",
    "m",
    "mar",
    "march",
    "may",
    "mm",
    "mon",
    "monday",
    "nov",
    "november",
    "now",
    "oct",
    "october",
    "on",
    "pm",
    "s",
    "sat",
    "saturday",
    "sep",
    "sept",
    "september",
    "sun",
    "sunday",
    "t",
    "thu",
    "thur",
    "thurs",
    "thursday",
    "today",
    "tomorrow",
    "tue",
    "tues",
    "tuesday",
    "wed",
    "wednesday",
    "weds",
    "y",
    "yesterday",
];

/// is_c_space reports whether a byte is a space to C's isspace.
fn is_c_space(b: u8) -> bool {
    matches!(b, b' ' | b'\t' | b'\n' | 0x0b | 0x0c | b'\r')
}

/// token_is reports whether a word matches a token of Postgres' date tables, which compare at most ten characters.
fn token_is(word: &str, token: &str) -> bool {
    &word.as_bytes()[..word.len().min(10)] == token.as_bytes()
}

/// WorkBuffer counts the bytes of Postgres' 256-byte work buffer that ParseDateTime fills with fields.
struct WorkBuffer {
    used: usize,
}

impl WorkBuffer {
    /// push appends a byte to a field, failing when the work buffer would run out.
    fn push(&mut self, field: &mut String, byte: u8) -> Step {
        if self.used + 1 >= 256 {
            return Err(IntervalError::BadFormat);
        }
        self.used += 1;
        field.push(byte.to_ascii_lowercase() as char);
        Ok(())
    }
}

/// date_fields splits date and time input into lowercased fields with their kinds, as Postgres' ParseDateTime does.
fn date_fields(text: &str) -> std::result::Result<Vec<(FieldKind, String)>, IntervalError> {
    let b = text.as_bytes();
    let at = |i: usize| b.get(i).copied().unwrap_or(0);
    let mut work = WorkBuffer { used: 0 };
    let mut fields = Vec::new();
    let mut i = 0;
    while i < b.len() {
        if is_c_space(b[i]) {
            i += 1;
            continue;
        }
        if fields.len() >= 25 {
            return Err(IntervalError::BadFormat);
        }
        let mut f = String::new();
        let kind;
        if b[i].is_ascii_digit() {
            while at(i).is_ascii_digit() {
                work.push(&mut f, b[i])?;
                i += 1;
            }
            if at(i) == b':' {
                kind = FieldKind::Time;
                while at(i).is_ascii_digit() || at(i) == b':' || at(i) == b'.' {
                    work.push(&mut f, b[i])?;
                    i += 1;
                }
            } else if matches!(at(i), b'-' | b'/' | b'.') {
                let delim = b[i];
                work.push(&mut f, delim)?;
                i += 1;
                if at(i).is_ascii_digit() {
                    let mut k = if delim == b'.' { FieldKind::Number } else { FieldKind::Date };
                    while at(i).is_ascii_digit() {
                        work.push(&mut f, b[i])?;
                        i += 1;
                    }
                    if at(i) == delim {
                        k = FieldKind::Date;
                        while at(i).is_ascii_digit() || at(i) == delim {
                            work.push(&mut f, b[i])?;
                            i += 1;
                        }
                    }
                    kind = k;
                } else {
                    kind = FieldKind::Date;
                    while at(i).is_ascii_alphanumeric() || at(i) == delim {
                        work.push(&mut f, b[i])?;
                        i += 1;
                    }
                }
            } else {
                kind = FieldKind::Number;
            }
        } else if b[i] == b'.' {
            kind = FieldKind::Number;
            work.push(&mut f, b[i])?;
            i += 1;
            while at(i).is_ascii_digit() {
                work.push(&mut f, b[i])?;
                i += 1;
            }
        } else if b[i].is_ascii_alphabetic() {
            while at(i).is_ascii_alphabetic() {
                work.push(&mut f, b[i])?;
                i += 1;
            }
            let is_date = matches!(at(i), b'-' | b'/' | b'.')
                || ((at(i) == b'+' || at(i).is_ascii_digit()) && !DATE_WORDS.iter().any(|w| token_is(&f, w)));
            if is_date {
                kind = FieldKind::Date;
                loop {
                    work.push(&mut f, b[i])?;
                    i += 1;
                    if !(matches!(at(i), b'+' | b'-' | b'/' | b'_' | b'.' | b':') || at(i).is_ascii_alphanumeric()) {
                        break;
                    }
                }
            } else {
                kind = FieldKind::Text;
            }
        } else if b[i] == b'+' || b[i] == b'-' {
            work.push(&mut f, b[i])?;
            i += 1;
            while is_c_space(at(i)) {
                i += 1;
            }
            if at(i).is_ascii_digit() {
                kind = FieldKind::Zone;
                while at(i).is_ascii_digit() || matches!(at(i), b':' | b'.' | b'-') {
                    work.push(&mut f, b[i])?;
                    i += 1;
                }
            } else if at(i).is_ascii_alphabetic() {
                kind = FieldKind::Special;
                while at(i).is_ascii_alphabetic() {
                    work.push(&mut f, b[i])?;
                    i += 1;
                }
            } else {
                return Err(IntervalError::BadFormat);
            }
        } else if b[i].is_ascii_punctuation() {
            i += 1;
            continue;
        } else {
            return Err(IntervalError::BadFormat);
        }
        work.used += 1;
        fields.push((kind, f));
    }
    Ok(fields)
}

/// strtol reads a signed integer starting at a position as C's strtol does, returning it, where it ends, and whether
/// it overflowed; without digits it ends where it starts.
fn strtol(s: &[u8], from: usize) -> (i64, usize, bool) {
    let mut i = from;
    while s.get(i).is_some_and(|&b| is_c_space(b)) {
        i += 1;
    }
    let negative = s.get(i) == Some(&b'-');
    if matches!(s.get(i), Some(b'+' | b'-')) {
        i += 1;
    }
    let digits = i;
    let (mut value, mut overflow) = (0i64, false);
    while let Some(&b) = s.get(i).filter(|b| b.is_ascii_digit()) {
        let digit = (b - b'0') as i64;
        match value.checked_mul(10).and_then(|v| if negative { v.checked_sub(digit) } else { v.checked_add(digit) }) {
            Some(v) => value = v,
            None => overflow = true,
        }
        i += 1;
    }
    if i == digits {
        return (0, from, false);
    }
    if overflow {
        value = if negative { i64::MIN } else { i64::MAX };
    }
    (value, i, overflow)
}

/// strtoint reads an integer as `strtol` does, where values outside 32 bits overflow, as Postgres' strtoint does.
fn strtoint(s: &[u8], from: usize) -> (i64, usize, bool) {
    let (value, end, overflow) = strtol(s, from);
    (value, end, overflow || i32::try_from(value).is_err())
}

/// parse_fraction reads a fraction that starts with its decimal point and runs to the end, as ParseFraction does.
fn parse_fraction(s: &str) -> std::result::Result<f64, IntervalError> {
    if s.len() == 1 {
        return Ok(0.0);
    }
    if crate::basetypes::geometric::float_prefix(s) != s.len() {
        return Err(IntervalError::BadFormat);
    }
    s.parse().map_err(|_| IntervalError::BadFormat)
}

/// TIME_FIELDS is the mask of Postgres' hour, minute, second, millisecond, and microsecond field bits.
const TIME_FIELDS: i32 = (1 << 10) | (1 << 11) | (1 << 12) | (1 << 13) | (1 << 14);

/// decode_time reads a time field of interval input as microseconds, as Postgres' DecodeTimeForInterval does, taking
/// two numbers as minutes and seconds for a MINUTE TO SECOND range or when the seconds have a fraction.
fn decode_time(s: &str, range: i32) -> std::result::Result<i64, IntervalError> {
    let b = s.as_bytes();
    let (mut hour, end, overflow) = strtol(b, 0);
    if overflow {
        return Err(IntervalError::Overflow);
    }
    if b.get(end) != Some(&b':') {
        return Err(IntervalError::BadFormat);
    }
    let (mut minute, mut end, overflow) = strtoint(b, end + 1);
    if overflow {
        return Err(IntervalError::Overflow);
    }
    let (mut second, mut fsec) = (0, 0);
    match b.get(end) {
        None if range == (1 << 11) | (1 << 12) => {
            second = minute;
            minute = i32::try_from(hour).map_err(|_| IntervalError::Overflow)? as i64;
            hour = 0;
        }
        None => {}
        Some(b'.') => {
            fsec = (parse_fraction(&s[end..])? * 1e6).round_ties_even() as i64;
            second = minute;
            minute = i32::try_from(hour).map_err(|_| IntervalError::Overflow)? as i64;
            hour = 0;
        }
        Some(b':') => {
            let (seconds, seconds_end, overflow) = strtoint(b, end + 1);
            if overflow {
                return Err(IntervalError::Overflow);
            }
            (second, end) = (seconds, seconds_end);
            match b.get(end) {
                Some(b'.') => fsec = (parse_fraction(&s[end..])? * 1e6).round_ties_even() as i64,
                None => {}
                _ => return Err(IntervalError::BadFormat),
            }
        }
        _ => return Err(IntervalError::BadFormat),
    }
    if hour < 0 || !(0..60).contains(&minute) || !(0..=60).contains(&second) || !(0..=USECS_PER_SEC).contains(&fsec) {
        return Err(IntervalError::Overflow);
    }
    hour.checked_mul(USECS_PER_HOUR)
        .and_then(|h| h.checked_add(fsec))
        .and_then(|m| m.checked_add(minute * USECS_PER_MINUTE))
        .and_then(|m| m.checked_add(second * USECS_PER_SEC))
        .ok_or(IntervalError::Overflow)
}

/// IntervalUnit is a unit that interval input counts in, where Other stands for units and words Postgres' interval
/// input refuses numbers of.
#[derive(Clone, Copy, PartialEq)]
enum IntervalUnit {
    Microsecond,
    Millisecond,
    Second,
    Minute,
    Hour,
    Day,
    Week,
    Month,
    Year,
    Decade,
    Century,
    Millennium,
    Other,
}

/// interval_word returns the unit a word of interval input names, as Postgres' deltatktbl lists them, with whether
/// the word is `ago`.
fn interval_word(word: &str) -> Option<(IntervalUnit, bool)> {
    use IntervalUnit::*;
    const WORDS: [(&str, IntervalUnit); 60] = [
        ("c", Century),
        ("cent", Century),
        ("centuries", Century),
        ("century", Century),
        ("d", Day),
        ("day", Day),
        ("days", Day),
        ("dec", Decade),
        ("decade", Decade),
        ("decades", Decade),
        ("decs", Decade),
        ("h", Hour),
        ("hour", Hour),
        ("hours", Hour),
        ("hr", Hour),
        ("hrs", Hour),
        ("m", Minute),
        ("microsecon", Microsecond),
        ("mil", Millennium),
        ("millennia", Millennium),
        ("millennium", Millennium),
        ("millisecon", Millisecond),
        ("mils", Millennium),
        ("min", Minute),
        ("mins", Minute),
        ("minute", Minute),
        ("minutes", Minute),
        ("mon", Month),
        ("mons", Month),
        ("month", Month),
        ("months", Month),
        ("ms", Millisecond),
        ("msec", Millisecond),
        ("msecond", Millisecond),
        ("mseconds", Millisecond),
        ("msecs", Millisecond),
        ("qtr", Other),
        ("quarter", Other),
        ("s", Second),
        ("sec", Second),
        ("second", Second),
        ("seconds", Second),
        ("secs", Second),
        ("timezone", Other),
        ("timezone_h", Other),
        ("timezone_m", Other),
        ("us", Microsecond),
        ("usec", Microsecond),
        ("usecond", Microsecond),
        ("useconds", Microsecond),
        ("usecs", Microsecond),
        ("w", Week),
        ("week", Week),
        ("weeks", Week),
        ("y", Year),
        ("year", Year),
        ("years", Year),
        ("yr", Year),
        ("yrs", Year),
        ("ago", Other),
    ];
    WORDS.iter().find(|(token, _)| token_is(word, token)).map(|&(token, unit)| (unit, token == "ago"))
}

/// decode_interval reads the fields of interval input from right to left, as Postgres' DecodeInterval does, so a
/// number without a unit counts the last field of the range, or days before a time.
fn decode_interval(
    fields: &[(FieldKind, String)],
    range: i32,
    sql_standard: bool,
) -> std::result::Result<IntervalParts, IntervalError> {
    use IntervalUnit::*;
    let mut parts = IntervalParts::default();
    let force_negative = sql_standard
        && fields.first().is_some_and(|(_, f)| f.starts_with('-'))
        && !fields[1..].iter().any(|(_, f)| f.starts_with(['-', '+']));
    let (month, year, day, hour, minute, second) = (1 << 1, 1 << 2, 1 << 3, 1 << 10, 1 << 11, 1 << 12);
    let (mut fmask, mut unit, mut is_before) = (0, None, false);
    for (kind, field) in fields.iter().rev() {
        let time = match kind {
            FieldKind::Time => Some(decode_time(field, range)?),
            FieldKind::Zone if field.contains(':') => {
                decode_time(&field[1..], range).ok().map(|m| if field.starts_with('-') { -m } else { m })
            }
            _ => None,
        };
        let tmask = match kind {
            _ if let Some(micros) = time => {
                parts.micros = if force_negative && micros > 0 { -micros } else { micros };
                unit = Some(Day);
                TIME_FIELDS
            }
            FieldKind::Text | FieldKind::Special => {
                let (named, ago) = interval_word(field).ok_or(IntervalError::BadFormat)?;
                is_before |= ago;
                unit = Some(named);
                0
            }
            _ => {
                let current = *unit.get_or_insert(match range {
                    r if r == year => Year,
                    r if r == month || r == year | month => Month,
                    r if r == day => Day,
                    r if r == hour || r == day | hour => Hour,
                    r if r == minute || r == hour | minute || r == day | hour | minute => Minute,
                    _ => Second,
                });
                let b = field.as_bytes();
                let (mut val, end, overflow) = strtol(b, 0);
                if overflow {
                    return Err(IntervalError::Overflow);
                }
                let mut fval = 0.0;
                let current = match b.get(end) {
                    Some(b'-') => {
                        let (months, end, overflow) = strtoint(b, end + 1);
                        if overflow || !(0..12).contains(&months) {
                            return Err(IntervalError::Overflow);
                        }
                        if end != b.len() {
                            return Err(IntervalError::BadFormat);
                        }
                        let months = if b[0] == b'-' { -months } else { months };
                        val = val.checked_mul(12).and_then(|v| v.checked_add(months)).ok_or(IntervalError::Overflow)?;
                        unit = Some(Month);
                        Month
                    }
                    Some(b'.') => {
                        fval = parse_fraction(&field[end..])?;
                        if b[0] == b'-' {
                            fval = -fval;
                        }
                        current
                    }
                    None => current,
                    _ => return Err(IntervalError::BadFormat),
                };
                if force_negative {
                    if val > 0 {
                        val = -val;
                    }
                    if fval > 0.0 {
                        fval = -fval;
                    }
                }
                match current {
                    Microsecond => parts.add_micros(val, fval, 1).map(|_| 1 << 14)?,
                    Millisecond => parts.add_micros(val, fval, 1000).map(|_| 1 << 13)?,
                    Second => {
                        parts.add_micros(val, fval, USECS_PER_SEC)?;
                        if fval == 0.0 { second } else { second | (1 << 13) | (1 << 14) }
                    }
                    Minute => parts.add_micros(val, fval, USECS_PER_MINUTE).map(|_| minute)?,
                    Hour => {
                        parts.add_micros(val, fval, USECS_PER_HOUR)?;
                        unit = Some(Day);
                        hour
                    }
                    Day => parts.add_days(val, 1).and_then(|_| parts.fract_micros(fval, USECS_PER_DAY)).map(|_| day)?,
                    Week => parts.add_days(val, 7).and_then(|_| parts.fract_days(fval, 7)).map(|_| 1 << 24)?,
                    Month => parts.add_months(val).and_then(|_| parts.fract_days(fval, 30)).map(|_| month)?,
                    Year => parts.add_years(val, 1).and_then(|_| parts.fract_years(fval, 1)).map(|_| year)?,
                    Decade => parts.add_years(val, 10).and_then(|_| parts.fract_years(fval, 10)).map(|_| 1 << 25)?,
                    Century => parts.add_years(val, 100).and_then(|_| parts.fract_years(fval, 100)).map(|_| 1 << 26)?,
                    Millennium => {
                        parts.add_years(val, 1000).and_then(|_| parts.fract_years(fval, 1000)).map(|_| 1 << 27)?
                    }
                    Other => return Err(IntervalError::BadFormat),
                }
            }
        };
        if tmask & fmask != 0 {
            return Err(IntervalError::BadFormat);
        }
        fmask |= tmask;
    }
    if fmask == 0 {
        return Err(IntervalError::BadFormat);
    }
    if is_before {
        if parts.micros == i64::MIN || parts.days == i32::MIN || parts.months == i32::MIN || parts.years == i32::MIN {
            return Err(IntervalError::Overflow);
        }
        parts = IntervalParts { micros: -parts.micros, days: -parts.days, months: -parts.months, years: -parts.years };
    }
    Ok(parts)
}

/// iso_number reads a number of ISO 8601 interval input as Postgres' ParseISO8601Number does, returning its whole
/// part, its fraction, and where it ends.
fn iso_number(text: &str, from: usize) -> std::result::Result<(i64, f64, usize), IntervalError> {
    if !matches!(text.as_bytes().get(from), Some(b'0'..=b'9' | b'-' | b'.')) {
        return Err(IntervalError::BadFormat);
    }
    let length = crate::basetypes::geometric::float_prefix(&text[from..]);
    let number = &text[from..from + length];
    let val: f64 = number.parse().map_err(|_| IntervalError::BadFormat)?;
    if val.is_infinite() && !number.to_ascii_lowercase().contains("inf") {
        return Err(IntervalError::BadFormat);
    }
    if val.is_nan() || !(-1.0e15..=1.0e15).contains(&val) {
        return Err(IntervalError::Overflow);
    }
    let whole = val.trunc() as i64;
    Ok((whole, val - whole as f64, from + length))
}

/// decode_iso_interval reads ISO 8601 interval input, with designators or in the alternative format, as Postgres'
/// DecodeISO8601Interval does.
fn decode_iso_interval(text: &str) -> std::result::Result<IntervalParts, IntervalError> {
    let s = text.as_bytes();
    let at = |i: usize| s.get(i).copied().unwrap_or(0);
    let width =
        |start: usize| s[start + usize::from(s[start] == b'-')..].iter().take_while(|b| b.is_ascii_digit()).count();
    let mut p = IntervalParts::default();
    if s.len() < 2 || s[0] != b'P' {
        return Err(IntervalError::BadFormat);
    }
    let (mut i, mut datepart, mut havefield) = (1, true, false);
    while i < s.len() {
        if s[i] == b'T' {
            (datepart, havefield) = (false, false);
            i += 1;
            continue;
        }
        let start = i;
        let (val, fval, end) = iso_number(text, i)?;
        let unit = at(end);
        i = end + 1;
        if datepart {
            match unit {
                b'Y' => p.add_years(val, 1).and_then(|_| p.fract_years(fval, 1))?,
                b'M' => p.add_months(val).and_then(|_| p.fract_days(fval, 30))?,
                b'W' => p.add_days(val, 7).and_then(|_| p.fract_days(fval, 7))?,
                b'D' => p.add_days(val, 1).and_then(|_| p.fract_micros(fval, USECS_PER_DAY))?,
                b'T' | 0 | b'-' => {
                    if unit != b'-' && width(start) == 8 && !havefield {
                        p.add_years(val / 10000, 1)?;
                        p.add_months((val / 100) % 100)?;
                        p.add_days(val % 100, 1)?;
                        p.fract_micros(fval, USECS_PER_DAY)?;
                        if unit == 0 {
                            return Ok(p);
                        }
                        (datepart, havefield) = (false, false);
                        continue;
                    }
                    if havefield {
                        return Err(IntervalError::BadFormat);
                    }
                    p.add_years(val, 1).and_then(|_| p.fract_years(fval, 1))?;
                    match unit {
                        0 => return Ok(p),
                        b'T' => {
                            (datepart, havefield) = (false, false);
                            continue;
                        }
                        _ => {}
                    }
                    let (val, fval, end) = iso_number(text, i)?;
                    i = end;
                    p.add_months(val).and_then(|_| p.fract_days(fval, 30))?;
                    match at(i) {
                        0 => return Ok(p),
                        b'T' => {
                            (datepart, havefield) = (false, false);
                            continue;
                        }
                        b'-' => i += 1,
                        _ => return Err(IntervalError::BadFormat),
                    }
                    let (val, fval, end) = iso_number(text, i)?;
                    i = end;
                    p.add_days(val, 1).and_then(|_| p.fract_micros(fval, USECS_PER_DAY))?;
                    match at(i) {
                        0 => return Ok(p),
                        b'T' => {
                            (datepart, havefield) = (false, false);
                            continue;
                        }
                        _ => return Err(IntervalError::BadFormat),
                    }
                }
                _ => return Err(IntervalError::BadFormat),
            }
        } else {
            match unit {
                b'H' => p.add_micros(val, fval, USECS_PER_HOUR)?,
                b'M' => p.add_micros(val, fval, USECS_PER_MINUTE)?,
                b'S' => p.add_micros(val, fval, USECS_PER_SEC)?,
                0 | b':' => {
                    if unit == 0 && width(start) == 6 && !havefield {
                        p.add_micros(val / 10000, 0.0, USECS_PER_HOUR)?;
                        p.add_micros((val / 100) % 100, 0.0, USECS_PER_MINUTE)?;
                        p.add_micros(val % 100, 0.0, USECS_PER_SEC)?;
                        p.fract_micros(fval, 1)?;
                        return Ok(p);
                    }
                    if havefield {
                        return Err(IntervalError::BadFormat);
                    }
                    p.add_micros(val, fval, USECS_PER_HOUR)?;
                    if unit == 0 {
                        return Ok(p);
                    }
                    let (val, fval, end) = iso_number(text, i)?;
                    i = end;
                    p.add_micros(val, fval, USECS_PER_MINUTE)?;
                    match at(i) {
                        0 => return Ok(p),
                        b':' => i += 1,
                        _ => return Err(IntervalError::BadFormat),
                    }
                    let (val, fval, end) = iso_number(text, i)?;
                    p.add_micros(val, fval, USECS_PER_SEC)?;
                    return if end == s.len() { Ok(p) } else { Err(IntervalError::BadFormat) };
                }
                _ => return Err(IntervalError::BadFormat),
            }
        }
        havefield = true;
    }
    Ok(p)
}

/// INTERVAL_FULL_RANGE and INTERVAL_FULL_PRECISION are the parts of an interval modifier that leave the interval as is.
pub const INTERVAL_FULL_RANGE: i32 = 0x7fff;
pub const INTERVAL_FULL_PRECISION: i32 = 0xffff;

/// INTERVAL_RANGES lists the field ranges an interval modifier can hold, as masks of Postgres' field bits, with the
/// words format_type writes for each.
pub const INTERVAL_RANGES: [(i32, &str); 14] = [
    (1 << 2, " year"),
    (1 << 1, " month"),
    (1 << 3, " day"),
    (1 << 10, " hour"),
    (1 << 11, " minute"),
    (1 << 12, " second"),
    ((1 << 2) | (1 << 1), " year to month"),
    ((1 << 3) | (1 << 10), " day to hour"),
    ((1 << 3) | (1 << 10) | (1 << 11), " day to minute"),
    ((1 << 3) | (1 << 10) | (1 << 11) | (1 << 12), " day to second"),
    ((1 << 10) | (1 << 11), " hour to minute"),
    ((1 << 10) | (1 << 11) | (1 << 12), " hour to second"),
    ((1 << 11) | (1 << 12), " minute to second"),
    (INTERVAL_FULL_RANGE, ""),
];

/// adjust_interval truncates an interval to the fields of a modifier's range, as Postgres' AdjustIntervalForTypmod
/// does, leaving the fractional seconds to the modifier's precision.
pub fn adjust_interval(iv: Interval, modifier: i32) -> Interval {
    if modifier < 0 {
        return iv;
    }
    let truncate = |micros: i64, unit: i64| micros / unit * unit;
    let (year, month, day, hour, minute) = (1 << 2, 1 << 1, 1 << 3, 1 << 10, 1 << 11);
    match (modifier >> 16) & 0x7fff {
        r if r == year => Interval { months: iv.months / 12 * 12, days: 0, micros: 0 },
        r if r == month || r == year | month => Interval { days: 0, micros: 0, ..iv },
        r if r == day => Interval { micros: 0, ..iv },
        r if r == hour || r == day | hour => Interval { micros: truncate(iv.micros, USECS_PER_HOUR), ..iv },
        r if r == minute || r == day | hour | minute || r == hour | minute => {
            Interval { micros: truncate(iv.micros, USECS_PER_MINUTE), ..iv }
        }
        _ => iv,
    }
}

/// parse_interval reads an interval in Postgres' formats: unit words, `@` verbose form with `ago`, times, SQL
/// standard year-month and day-time, and ISO 8601.
pub fn parse_interval(text: &str) -> Result<Interval> {
    parse_interval_with_modifier(text, -1)
}

/// parse_interval_with_modifier reads an interval as `parse_interval` does, where a number without a unit counts the
/// last field of the modifier's range, or seconds without one.
pub fn parse_interval_with_modifier(text: &str, modifier: i32) -> Result<Interval> {
    match text.trim().to_lowercase().as_str() {
        "infinity" | "+infinity" => return Ok(Interval::INFINITY),
        "-infinity" => return Ok(Interval::NEG_INFINITY),
        _ => {}
    }
    let range = if modifier >= 0 { modifier >> 16 & 0x7fff } else { INTERVAL_FULL_RANGE };
    let sql_standard = with_format(|f| f.interval_style == IntervalStyle::SqlStandard);
    let mut parts = date_fields(text).and_then(|fields| decode_interval(&fields, range, sql_standard));
    if matches!(parts, Err(IntervalError::BadFormat)) {
        parts = decode_iso_interval(text);
    }
    let parts = parts.map_err(|e| match e {
        IntervalError::BadFormat => {
            PgError::new(code::INVALID_DATETIME_FORMAT, format!("invalid input syntax for type interval: \"{text}\""))
        }
        IntervalError::Overflow => {
            PgError::new(code::INTERVAL_FIELD_OVERFLOW, format!("interval field value out of range: \"{text}\""))
        }
    })?;
    let months = i32::try_from(parts.years as i64 * 12 + parts.months as i64)
        .map_err(|_| PgError::new(code::DATETIME_FIELD_OVERFLOW, "interval out of range"))?;
    Ok(Interval { months, days: parts.days, micros: parts.micros })
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
