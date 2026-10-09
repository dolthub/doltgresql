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

//! The datetime templates of to_char, to_timestamp, and to_date, read and written as Postgres' formatting.c does.

use crate::datetime::{
    self as dt, Fields, Interval, POSTGRES_EPOCH_JDATE, USECS_PER_HOUR, USECS_PER_MINUTE, USECS_PER_SEC, Zone,
};
use crate::error::{PgError, Result, code};
use Case::{Capital, Lower, Upper};
use Mode::{Gregorian, IsoWeek};

/// MONTHS are the full names of the months.
const MONTHS: [&str; 12] = [
    "January",
    "February",
    "March",
    "April",
    "May",
    "June",
    "July",
    "August",
    "September",
    "October",
    "November",
    "December",
];

/// MONTH_ABBREVIATIONS are the abbreviated names of the months.
const MONTH_ABBREVIATIONS: [&str; 12] =
    ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

/// DAYS are the full names of the days of the week, from Sunday.
const DAYS: [&str; 7] = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

/// DAY_ABBREVIATIONS are the abbreviated names of the days of the week, from Sunday.
const DAY_ABBREVIATIONS: [&str; 7] = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

/// ROMAN_MONTHS are the months in Roman numerals from December back to January, so that longer numerals match first.
const ROMAN_MONTHS: [&str; 12] = ["xii", "xi", "x", "ix", "viii", "vii", "vi", "v", "iv", "iii", "ii", "i"];

/// ERAS are the era markers, where an odd position is BC.
const ERAS: [&str; 4] = ["ad", "bc", "AD", "BC"];

/// ERAS_DOTTED are the era markers with periods, where an odd position is BC.
const ERAS_DOTTED: [&str; 4] = ["a.d.", "b.c.", "A.D.", "B.C."];

/// MERIDIEMS are the meridiem markers, where an odd position is PM.
const MERIDIEMS: [&str; 4] = ["am", "pm", "AM", "PM"];

/// MERIDIEMS_DOTTED are the meridiem markers with periods, where an odd position is PM.
const MERIDIEMS_DOTTED: [&str; 4] = ["a.m.", "p.m.", "A.M.", "P.M."];

/// Id is what a template keyword stands for.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Id {
    EraDotted,
    Era,
    MeridiemDotted,
    Meridiem,
    Cc,
    DayName,
    DayAbbreviation,
    Ddd,
    Dd,
    D,
    Ff(u8),
    Fx,
    Hh24,
    Hh12,
    Iddd,
    IsoD,
    Iw,
    Iyyy,
    Iyy,
    Iy,
    I,
    J,
    Mi,
    Mm,
    MonthName,
    MonthAbbreviation,
    Ms,
    Of,
    Q,
    Rm,
    Ssss,
    Ss,
    Tzh,
    Tzm,
    Tz,
    Us,
    Ww,
    W,
    YComma,
    Yyyy,
    Yyy,
    Yy,
    Y,
}

/// Case is the letter case a keyword writes its words in.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Case {
    Upper,
    Capital,
    Lower,
}

/// Mode is the date convention a keyword belongs to.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Mode {
    None,
    Gregorian,
    IsoWeek,
}

/// Keyword is a template keyword: its text, meaning, letter case, whether it reads digits, and its date convention.
struct Keyword {
    name: &'static str,
    id: Id,
    case: Case,
    digit: bool,
    mode: Mode,
}

/// k declares a keyword.
const fn k(name: &'static str, id: Id, case: Case, digit: bool, mode: Mode) -> Keyword {
    Keyword { name, id, case, digit, mode }
}

/// KEYWORDS are the template keywords in Postgres' search order, where a longer keyword precedes its prefixes.
const KEYWORDS: &[Keyword] = &[
    k("A.D.", Id::EraDotted, Upper, false, Mode::None),
    k("A.M.", Id::MeridiemDotted, Upper, false, Mode::None),
    k("AD", Id::Era, Upper, false, Mode::None),
    k("AM", Id::Meridiem, Upper, false, Mode::None),
    k("B.C.", Id::EraDotted, Upper, false, Mode::None),
    k("BC", Id::Era, Upper, false, Mode::None),
    k("CC", Id::Cc, Upper, true, Mode::None),
    k("DAY", Id::DayName, Upper, false, Mode::None),
    k("DDD", Id::Ddd, Upper, true, Gregorian),
    k("DD", Id::Dd, Upper, true, Gregorian),
    k("DY", Id::DayAbbreviation, Upper, false, Mode::None),
    k("Day", Id::DayName, Capital, false, Mode::None),
    k("Dy", Id::DayAbbreviation, Capital, false, Mode::None),
    k("D", Id::D, Upper, true, Gregorian),
    k("FF1", Id::Ff(1), Upper, true, Mode::None),
    k("FF2", Id::Ff(2), Upper, true, Mode::None),
    k("FF3", Id::Ff(3), Upper, true, Mode::None),
    k("FF4", Id::Ff(4), Upper, true, Mode::None),
    k("FF5", Id::Ff(5), Upper, true, Mode::None),
    k("FF6", Id::Ff(6), Upper, true, Mode::None),
    k("FX", Id::Fx, Upper, false, Mode::None),
    k("HH24", Id::Hh24, Upper, true, Mode::None),
    k("HH12", Id::Hh12, Upper, true, Mode::None),
    k("HH", Id::Hh12, Upper, true, Mode::None),
    k("IDDD", Id::Iddd, Upper, true, IsoWeek),
    k("ID", Id::IsoD, Upper, true, IsoWeek),
    k("IW", Id::Iw, Upper, true, IsoWeek),
    k("IYYY", Id::Iyyy, Upper, true, IsoWeek),
    k("IYY", Id::Iyy, Upper, true, IsoWeek),
    k("IY", Id::Iy, Upper, true, IsoWeek),
    k("I", Id::I, Upper, true, IsoWeek),
    k("J", Id::J, Upper, true, Mode::None),
    k("MI", Id::Mi, Upper, true, Mode::None),
    k("MM", Id::Mm, Upper, true, Gregorian),
    k("MONTH", Id::MonthName, Upper, false, Gregorian),
    k("MON", Id::MonthAbbreviation, Upper, false, Gregorian),
    k("MS", Id::Ms, Upper, true, Mode::None),
    k("Month", Id::MonthName, Capital, false, Gregorian),
    k("Mon", Id::MonthAbbreviation, Capital, false, Gregorian),
    k("OF", Id::Of, Upper, false, Mode::None),
    k("P.M.", Id::MeridiemDotted, Upper, false, Mode::None),
    k("PM", Id::Meridiem, Upper, false, Mode::None),
    k("Q", Id::Q, Upper, true, Mode::None),
    k("RM", Id::Rm, Upper, false, Gregorian),
    k("SSSSS", Id::Ssss, Upper, true, Mode::None),
    k("SSSS", Id::Ssss, Upper, true, Mode::None),
    k("SS", Id::Ss, Upper, true, Mode::None),
    k("TZH", Id::Tzh, Upper, false, Mode::None),
    k("TZM", Id::Tzm, Upper, true, Mode::None),
    k("TZ", Id::Tz, Upper, false, Mode::None),
    k("US", Id::Us, Upper, true, Mode::None),
    k("WW", Id::Ww, Upper, true, Gregorian),
    k("W", Id::W, Upper, true, Gregorian),
    k("Y,YYY", Id::YComma, Upper, true, Gregorian),
    k("YYYY", Id::Yyyy, Upper, true, Gregorian),
    k("YYY", Id::Yyy, Upper, true, Gregorian),
    k("YY", Id::Yy, Upper, true, Gregorian),
    k("Y", Id::Y, Upper, true, Gregorian),
    k("a.d.", Id::EraDotted, Lower, false, Mode::None),
    k("a.m.", Id::MeridiemDotted, Lower, false, Mode::None),
    k("ad", Id::Era, Lower, false, Mode::None),
    k("am", Id::Meridiem, Lower, false, Mode::None),
    k("b.c.", Id::EraDotted, Lower, false, Mode::None),
    k("bc", Id::Era, Lower, false, Mode::None),
    k("cc", Id::Cc, Upper, true, Mode::None),
    k("day", Id::DayName, Lower, false, Mode::None),
    k("ddd", Id::Ddd, Upper, true, Gregorian),
    k("dd", Id::Dd, Upper, true, Gregorian),
    k("dy", Id::DayAbbreviation, Lower, false, Mode::None),
    k("d", Id::D, Upper, true, Gregorian),
    k("ff1", Id::Ff(1), Upper, true, Mode::None),
    k("ff2", Id::Ff(2), Upper, true, Mode::None),
    k("ff3", Id::Ff(3), Upper, true, Mode::None),
    k("ff4", Id::Ff(4), Upper, true, Mode::None),
    k("ff5", Id::Ff(5), Upper, true, Mode::None),
    k("ff6", Id::Ff(6), Upper, true, Mode::None),
    k("fx", Id::Fx, Upper, false, Mode::None),
    k("hh24", Id::Hh24, Upper, true, Mode::None),
    k("hh12", Id::Hh12, Upper, true, Mode::None),
    k("hh", Id::Hh12, Upper, true, Mode::None),
    k("iddd", Id::Iddd, Upper, true, IsoWeek),
    k("id", Id::IsoD, Upper, true, IsoWeek),
    k("iw", Id::Iw, Upper, true, IsoWeek),
    k("iyyy", Id::Iyyy, Upper, true, IsoWeek),
    k("iyy", Id::Iyy, Upper, true, IsoWeek),
    k("iy", Id::Iy, Upper, true, IsoWeek),
    k("i", Id::I, Upper, true, IsoWeek),
    k("j", Id::J, Upper, true, Mode::None),
    k("mi", Id::Mi, Upper, true, Mode::None),
    k("mm", Id::Mm, Upper, true, Gregorian),
    k("month", Id::MonthName, Lower, false, Gregorian),
    k("mon", Id::MonthAbbreviation, Lower, false, Gregorian),
    k("ms", Id::Ms, Upper, true, Mode::None),
    k("of", Id::Of, Upper, false, Mode::None),
    k("p.m.", Id::MeridiemDotted, Lower, false, Mode::None),
    k("pm", Id::Meridiem, Lower, false, Mode::None),
    k("q", Id::Q, Upper, true, Mode::None),
    k("rm", Id::Rm, Lower, false, Gregorian),
    k("sssss", Id::Ssss, Upper, true, Mode::None),
    k("ssss", Id::Ssss, Upper, true, Mode::None),
    k("ss", Id::Ss, Upper, true, Mode::None),
    k("tzh", Id::Tzh, Upper, false, Mode::None),
    k("tzm", Id::Tzm, Upper, true, Mode::None),
    k("tz", Id::Tz, Lower, false, Mode::None),
    k("us", Id::Us, Upper, true, Mode::None),
    k("ww", Id::Ww, Upper, true, Gregorian),
    k("w", Id::W, Upper, true, Gregorian),
    k("y,yyy", Id::YComma, Upper, true, Gregorian),
    k("yyyy", Id::Yyyy, Upper, true, Gregorian),
    k("yyy", Id::Yyy, Upper, true, Gregorian),
    k("yy", Id::Yy, Upper, true, Gregorian),
    k("y", Id::Y, Upper, true, Gregorian),
];

/// FM is the prefix that turns off padding.
const FM: u8 = 1;

/// TH_UPPER is the suffix that adds an upper-case ordinal suffix.
const TH_UPPER: u8 = 2;

/// TH_LOWER is the suffix that adds a lower-case ordinal suffix.
const TH_LOWER: u8 = 4;

/// SP is the spell-mode suffix, which Postgres reads but ignores.
const SP: u8 = 8;

/// TM is the prefix that translates names, which in Doltgres' only locale turns off padding.
const TM: u8 = 16;

/// Node is one part of a parsed template.
enum Node {
    /// A keyword with its suffixes.
    Action(&'static Keyword, u8),
    /// A literal character.
    Char(char),
    /// A punctuation character that separates fields.
    Separator(char),
    Space(char),
}

impl Node {
    /// text returns the literal text of a node that is not a keyword.
    fn text(&self) -> Option<char> {
        match self {
            Node::Action(..) => None,
            Node::Char(c) | Node::Separator(c) | Node::Space(c) => Some(*c),
        }
    }
}

/// is_separator reports whether a character is ASCII punctuation, which separates fields.
fn is_separator(c: char) -> bool {
    c.is_ascii_graphic() && !c.is_ascii_alphanumeric()
}

/// is_space reports whether a byte is white space, as C's isspace does.
fn is_space(b: u8) -> bool {
    matches!(b, b' ' | b'\t' | b'\n' | 0x0b | 0x0c | b'\r')
}

/// parse_format splits a template into keywords with their suffixes, quoted text, and other characters.
fn parse_format(template: &str) -> Vec<Node> {
    let mut nodes = Vec::new();
    let mut rest = template;
    while !rest.is_empty() {
        let mut suffix = 0;
        for (prefix, flag) in [("FM", FM), ("fm", FM), ("TM", TM), ("tm", TM)] {
            if let Some(after) = rest.strip_prefix(prefix) {
                suffix |= flag;
                rest = after;
                break;
            }
        }
        let first = rest.as_bytes().first().copied();
        let keyword = first.and_then(|first| {
            KEYWORDS.iter().filter(|k| k.name.as_bytes()[0] == first).find(|k| rest.starts_with(k.name))
        });
        if let Some(keyword) = keyword {
            rest = &rest[keyword.name.len()..];
            for (postfix, flag) in [("TH", TH_UPPER), ("th", TH_LOWER), ("SP", SP)] {
                if let Some(after) = rest.strip_prefix(postfix) {
                    suffix |= flag;
                    rest = after;
                    break;
                }
            }
            nodes.push(Node::Action(keyword, suffix));
            continue;
        }
        let mut chars = rest.chars();
        let Some(c) = chars.next() else { break };
        if c == '"' {
            while let Some(c) = chars.next() {
                if c == '"' {
                    break;
                }
                let quoted = chars.clone().next().filter(|_| c == '\\');
                if quoted.is_some() {
                    chars.next();
                }
                nodes.push(Node::Char(quoted.unwrap_or(c)));
            }
            rest = chars.as_str();
            continue;
        }
        let c = if c == '\\' && rest[1..].starts_with('"') {
            chars.next();
            '"'
        } else {
            c
        };
        nodes.push(if is_separator(c) {
            Node::Separator(c)
        } else if c.is_ascii() && is_space(c as u8) {
            Node::Space(c)
        } else {
            Node::Char(c)
        });
        rest = chars.as_str();
    }
    nodes
}

/// ordinal returns the English ordinal suffix of a number's text, as Postgres' get_th does.
fn ordinal(number: &str, upper: bool) -> Result<&'static str> {
    let bytes = number.as_bytes();
    let Some(&last) = bytes.last().filter(|b| b.is_ascii_digit()) else {
        return Err(PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("\"{number}\" is not a number")));
    };
    let last = if bytes.len() > 1 && bytes[bytes.len() - 2] == b'1' { b'0' } else { last };
    let index = match last {
        b'1' => 0,
        b'2' => 1,
        b'3' => 2,
        _ => 3,
    };
    Ok(if upper { ["ST", "ND", "RD", "TH"][index] } else { ["st", "nd", "rd", "th"][index] })
}

/// padded writes a number padded with zeros to a width, or unpadded in fill mode.
fn padded(value: i64, width: usize, suffix: u8) -> String {
    if suffix & FM != 0 { value.to_string() } else { format!("{value:0width$}") }
}

/// cased writes a word in a letter case.
fn cased(word: &str, case: Case) -> String {
    match case {
        Upper => word.to_ascii_uppercase(),
        Lower => word.to_ascii_lowercase(),
        Capital => word.to_string(),
    }
}

/// j2day returns the day of the week of a Julian day, from Sunday as 0.
fn j2day(julian: i64) -> i64 {
    (julian + 1).rem_euclid(7)
}

/// isoweek2j returns the Julian day of the Monday of an ISO week of an ISO year.
fn isoweek2j(year: i64, week: i64) -> i64 {
    let day4 = dt::date2j(year, 1, 4);
    let day0 = j2day(day4 - 1);
    (week - 1) * 7 + (day4 - day0)
}

/// iso_year_and_week returns the ISO year and week of a date, as Postgres' date2isoyear and date2isoweek do.
fn iso_year_and_week(year: i64, month: i64, day: i64) -> (i64, i64) {
    let dayn = dt::date2j(year, month, day);
    let mut year = year;
    let start = |y: i64| {
        let day4 = dt::date2j(y, 1, 4);
        day4 - j2day(day4 - 1)
    };
    let mut first = start(year);
    if dayn < first {
        year -= 1;
        first = start(year);
    }
    let mut week = (dayn - first) / 7 + 1;
    if week >= 52 {
        let next = start(year + 1);
        if dayn >= next {
            year += 1;
            week = (dayn - next) / 7 + 1;
        }
    }
    (year, week)
}

/// Tm is the broken-down value that to_char writes, which for an interval holds its parts.
#[derive(Default)]
struct Tm {
    sec: i64,
    min: i64,
    hour: i64,
    mday: i64,
    mon: i64,
    year: i64,
    wday: i64,
    yday: i64,
    gmtoff: i64,
    fsec: i64,
    zone: Option<String>,
}

/// adjust_year converts an astronomical year to the year number written with an era, keeping interval years.
fn adjust_year(year: i64, interval: bool) -> i64 {
    if interval || year > 0 { year } else { -(year - 1) }
}

/// invalid_for_interval returns Postgres' error for a template keyword that an interval cannot fill.
fn invalid_for_interval() -> PgError {
    PgError {
        hint: Some("Intervals are not tied to specific calendar dates.".into()),
        ..PgError::new(code::INVALID_DATETIME_FORMAT, "invalid format specification for an interval value")
    }
}

/// write_template writes a broken-down value with a parsed template, as Postgres' DCH_to_char does.
fn write_template(nodes: &[Node], tm: &Tm, interval: bool) -> Result<String> {
    let mut out = String::new();
    for node in nodes {
        let (keyword, suffix) = match node {
            Node::Action(keyword, suffix) => (*keyword, *suffix),
            other => {
                out.extend(other.text());
                continue;
            }
        };
        let fill = suffix & FM != 0;
        let translated = suffix & TM != 0;
        let case = keyword.case;
        let not_interval = || if interval { Err(invalid_for_interval()) } else { Ok(()) };
        let width = |value: i64, positive: usize| if value >= 0 { positive } else { positive + 1 };
        let word = |name: &str, pad: bool| {
            let text = cased(name, case);
            if pad && !fill && !translated { format!("{text:<9}") } else { text }
        };
        let iso = || iso_year_and_week(tm.year, tm.mon, tm.mday);
        let number = match keyword.id {
            Id::MeridiemDotted | Id::Meridiem => {
                let pm = tm.hour % 24 >= 12;
                let text = match (keyword.id, pm) {
                    (Id::MeridiemDotted, false) => "A.M.",
                    (Id::MeridiemDotted, true) => "P.M.",
                    (_, false) => "AM",
                    (_, true) => "PM",
                };
                out.push_str(&cased(text, case));
                None
            }
            Id::Hh12 => {
                let hour = if tm.hour % 12 == 0 { 12 } else { tm.hour % 12 };
                Some(padded(hour, width(tm.hour, 2), suffix))
            }
            Id::Hh24 => Some(padded(tm.hour, width(tm.hour, 2), suffix)),
            Id::Mi => Some(padded(tm.min, width(tm.min, 2), suffix)),
            Id::Ss => Some(padded(tm.sec, width(tm.sec, 2), suffix)),
            Id::Ff(digits) => {
                let divisor = 10i64.pow(6 - digits as u32);
                Some(format!("{:0width$}", tm.fsec / divisor, width = digits as usize))
            }
            Id::Ms => Some(format!("{:03}", tm.fsec / 1000)),
            Id::Us => Some(format!("{:06}", tm.fsec)),
            Id::Ssss => Some((tm.hour * 3600 + tm.min * 60 + tm.sec).to_string()),
            Id::Tz => {
                not_interval()?;
                if let Some(zone) = &tm.zone {
                    out.push_str(&if case == Lower { zone.to_ascii_lowercase() } else { zone.clone() });
                }
                None
            }
            Id::Tzh => {
                not_interval()?;
                let sign = if tm.gmtoff >= 0 { '+' } else { '-' };
                out.push_str(&format!("{sign}{:02}", tm.gmtoff.abs() / 3600));
                None
            }
            Id::Tzm => {
                not_interval()?;
                out.push_str(&format!("{:02}", tm.gmtoff.abs() % 3600 / 60));
                None
            }
            Id::Of => {
                not_interval()?;
                let sign = if tm.gmtoff >= 0 { '+' } else { '-' };
                out.push_str(&format!("{sign}{}", padded(tm.gmtoff.abs() / 3600, 2, suffix)));
                if tm.gmtoff.abs() % 3600 != 0 {
                    out.push_str(&format!(":{:02}", tm.gmtoff.abs() % 3600 / 60));
                }
                None
            }
            Id::EraDotted | Id::Era => {
                not_interval()?;
                let bc = tm.year <= 0;
                let text = match (keyword.id, bc) {
                    (Id::EraDotted, false) => "A.D.",
                    (Id::EraDotted, true) => "B.C.",
                    (_, false) => "AD",
                    (_, true) => "BC",
                };
                out.push_str(&cased(text, case));
                None
            }
            Id::MonthName | Id::MonthAbbreviation => {
                not_interval()?;
                if tm.mon != 0 {
                    let index = (tm.mon - 1) as usize;
                    out.push_str(&if keyword.id == Id::MonthName {
                        word(MONTHS[index], true)
                    } else {
                        word(MONTH_ABBREVIATIONS[index], false)
                    });
                }
                None
            }
            Id::Mm => Some(padded(tm.mon, width(tm.mon, 2), suffix)),
            Id::DayName | Id::DayAbbreviation => {
                not_interval()?;
                let index = tm.wday as usize;
                out.push_str(&if keyword.id == Id::DayName {
                    word(DAYS[index], true)
                } else {
                    word(DAY_ABBREVIATIONS[index], false)
                });
                None
            }
            Id::Ddd => Some(padded(tm.yday, 3, suffix)),
            Id::Iddd => {
                let (iso_year, _) = iso();
                let day = dt::date2j(tm.year, tm.mon, tm.mday) - isoweek2j(iso_year, 1) + 1;
                Some(padded(day, 3, suffix))
            }
            Id::Dd => Some(padded(tm.mday, 2, suffix)),
            Id::D => {
                not_interval()?;
                Some((tm.wday + 1).to_string())
            }
            Id::IsoD => {
                not_interval()?;
                Some((if tm.wday == 0 { 7 } else { tm.wday }).to_string())
            }
            Id::Ww => Some(padded((tm.yday - 1) / 7 + 1, 2, suffix)),
            Id::Iw => Some(padded(iso().1, 2, suffix)),
            Id::Q => (tm.mon != 0).then(|| ((tm.mon - 1) / 3 + 1).to_string()),
            Id::Cc => {
                let century = if interval {
                    tm.year / 100
                } else if tm.year > 0 {
                    (tm.year - 1) / 100 + 1
                } else {
                    tm.year / 100 - 1
                };
                Some(if (-99..=99).contains(&century) {
                    padded(century, width(century, 2), suffix)
                } else {
                    century.to_string()
                })
            }
            Id::YComma => {
                let year = adjust_year(tm.year, interval);
                let thousands = year / 1000;
                Some(format!("{thousands},{:03}", year - thousands * 1000))
            }
            Id::Yyyy | Id::Iyyy | Id::Yyy | Id::Iyy | Id::Yy | Id::Iy | Id::Y | Id::I => {
                let iso_based = matches!(keyword.id, Id::Iyyy | Id::Iyy | Id::Iy | Id::I);
                let year = adjust_year(if iso_based { iso().0 } else { tm.year }, interval);
                Some(match keyword.id {
                    Id::Yyyy | Id::Iyyy => padded(year, width(year, 4), suffix),
                    Id::Yyy | Id::Iyy => padded(year % 1000, width(year, 3), suffix),
                    Id::Yy | Id::Iy => padded(year % 100, width(year, 2), suffix),
                    _ => (year % 10).to_string(),
                })
            }
            Id::Rm => {
                if tm.mon != 0 || tm.year != 0 {
                    let index = if tm.mon == 0 {
                        if tm.year >= 0 { 0 } else { 11 }
                    } else if tm.mon < 0 {
                        (-(tm.mon + 1)) as usize
                    } else {
                        (12 - tm.mon) as usize
                    };
                    let numeral = cased(ROMAN_MONTHS[index], if case == Lower { Lower } else { Upper });
                    out.push_str(&if fill { numeral } else { format!("{numeral:<4}") });
                }
                None
            }
            Id::W => Some(((tm.mday - 1) / 7 + 1).to_string()),
            Id::J => Some(dt::date2j(tm.year, tm.mon, tm.mday).to_string()),
            Id::Fx => None,
        };
        if let Some(number) = number {
            out.push_str(&number);
            if suffix & (TH_UPPER | TH_LOWER) != 0 {
                out.push_str(ordinal(&number, suffix & TH_UPPER != 0)?);
            }
        }
    }
    Ok(out)
}

/// timestamp_to_char writes a timestamp with a template, where a timestamptz brings its local offset east of UTC and
/// its zone's abbreviation, returning None for an empty template or an infinite timestamp.
pub fn timestamp_to_char(local: i64, zone: Option<(i32, String)>, template: &str) -> Result<Option<String>> {
    if template.is_empty() || local == dt::TIMESTAMP_NOBEGIN || local == dt::TIMESTAMP_NOEND {
        return Ok(None);
    }
    let f = dt::fields_of_timestamp(local);
    let julian = dt::date2j(f.year, f.month, f.day);
    let tm = Tm {
        sec: f.second,
        min: f.minute,
        hour: f.hour,
        mday: f.day,
        mon: f.month,
        year: f.year,
        wday: j2day(julian),
        yday: julian - dt::date2j(f.year, 1, 1) + 1,
        gmtoff: zone.as_ref().map_or(0, |z| z.0 as i64),
        fsec: f.micros,
        zone: zone.map(|z| z.1),
    };
    write_template(&parse_format(template), &tm, false).map(Some)
}

/// interval_to_char writes an interval with a template, returning None for an empty template.
pub fn interval_to_char(interval: &Interval, template: &str) -> Result<Option<String>> {
    if template.is_empty() {
        return Ok(None);
    }
    let micros = interval.micros;
    let year = interval.months as i64 / 12;
    let mon = interval.months as i64 % 12;
    let mday = interval.days as i64;
    let tm = Tm {
        hour: micros / USECS_PER_HOUR,
        min: micros % USECS_PER_HOUR / USECS_PER_MINUTE,
        sec: micros % USECS_PER_MINUTE / USECS_PER_SEC,
        fsec: micros % USECS_PER_SEC,
        mday,
        mon,
        year,
        yday: (year * 12 + mon) * 30 + mday,
        ..Tm::default()
    };
    write_template(&parse_format(template), &tm, true).map(Some)
}

/// FromChar is what reading text with a template found, as Postgres' TmFromChar holds it.
#[derive(Default)]
struct FromChar {
    mode: Option<Mode>,
    hh: i64,
    pm: i64,
    mi: i64,
    ss: i64,
    ssss: i64,
    d: i64,
    dd: i64,
    ddd: i64,
    mm: i64,
    ms: i64,
    year: i64,
    bc: i64,
    ww: i64,
    w: i64,
    cc: i64,
    j: i64,
    us: i64,
    yysz: i64,
    twelve_hour: bool,
    tzsign: i64,
    tzh: i64,
    tzm: i64,
    ff: i64,
}

/// invalid_format returns an error about the template or the text it reads.
fn invalid_format(message: impl Into<String>) -> PgError {
    PgError::new(code::INVALID_DATETIME_FORMAT, message)
}

/// set_field stores a field's value, failing when a different value was stored before.
fn set_field(field: &mut i64, value: i64, keyword: &Keyword) -> Result<()> {
    if *field != 0 && *field != value {
        return Err(PgError {
            detail: Some("This value contradicts a previous setting for the same field type.".into()),
            ..invalid_format(format!("conflicting values for \"{}\" field in formatting string", keyword.name))
        });
    }
    *field = value;
    Ok(())
}

/// Reader reads text with a template's nodes.
struct Reader<'a> {
    text: &'a [u8],
    at: usize,
}

impl Reader<'_> {
    /// peek returns the byte being read, or zero at the end.
    fn peek(&self) -> u8 {
        self.text.get(self.at).copied().unwrap_or(0)
    }

    /// char_len returns the length of the character being read.
    fn char_len(&self) -> usize {
        match self.peek() {
            0 => 0,
            b if b < 0x80 => 1,
            b if b >= 0xf0 => 4,
            b if b >= 0xe0 => 3,
            _ => 2,
        }
        .min(self.text.len() - self.at)
    }

    /// skip_spaces skips white space, returning how much it skipped.
    fn skip_spaces(&mut self) -> usize {
        let start = self.at;
        while is_space(self.peek()) {
            self.at += 1;
        }
        self.at - start
    }

    /// rest returns the text left to read.
    fn rest(&self) -> String {
        String::from_utf8_lossy(&self.text[self.at..]).into_owned()
    }

    /// strtol reads an integer as C's strtol does from a position, returning it and the position after it, or None
    /// when no digits follow.
    fn strtol(text: &[u8], from: usize) -> Option<(i128, usize)> {
        let mut i = from;
        while text.get(i).is_some_and(|&b| is_space(b)) {
            i += 1;
        }
        let negative = match text.get(i) {
            Some(b'-') => {
                i += 1;
                true
            }
            Some(b'+') => {
                i += 1;
                false
            }
            _ => false,
        };
        let digits_start = i;
        let mut value: i128 = 0;
        while let Some(&b) = text.get(i).filter(|b| b.is_ascii_digit()) {
            value = (value * 10 + (b - b'0') as i128).min(i128::from(i64::MAX) + 1);
            i += 1;
        }
        (i > digits_start).then_some((if negative { -value } else { value }, i))
    }

    /// parse_int reads an integer field of a keyword, which is a fixed number of characters unless the keyword is in
    /// fill mode or a non-digit follows it, as Postgres' from_char_parse_int_len does, returning it with the number of
    /// characters read.
    fn parse_int(&mut self, len: usize, nodes: &[Node], index: usize) -> Result<(i64, usize)> {
        let Node::Action(keyword, suffix) = &nodes[index] else { return Ok((0, 0)) };
        let init = self.at;
        self.skip_spaces();
        let copy: String = String::from_utf8_lossy(&self.text[self.at..(self.at + len).min(self.text.len())]).into();
        let available = self.text.len() - self.at;
        let too_wide = || PgError {
            hint: Some("If your source string is not fixed-width, try using the \"FM\" modifier.".into()),
            ..invalid_format(format!("invalid value \"{copy}\" for \"{}\"", keyword.name))
        };
        let result = if suffix & FM != 0 || next_is_separator(nodes, index) {
            match Self::strtol(self.text, init) {
                Some((value, end)) => {
                    self.at = end;
                    value
                }
                None => {
                    self.at = init;
                    0
                }
            }
        } else {
            if available < len {
                return Err(PgError {
                    detail: Some(format!("Field requires {len} characters, but only {available} remain.")),
                    hint: Some("If your source string is not fixed-width, try using the \"FM\" modifier.".into()),
                    ..invalid_format(format!("source string too short for \"{}\" formatting field", keyword.name))
                });
            }
            let (value, used) = match Self::strtol(copy.as_bytes(), 0) {
                Some((value, end)) => (value, end),
                None => (0, 0),
            };
            if used > 0 && used < len {
                return Err(PgError {
                    detail: Some(format!("Field requires {len} characters, but only {used} could be parsed.")),
                    ..too_wide()
                });
            }
            self.at += used;
            value
        };
        if self.at == init {
            return Err(PgError {
                detail: Some("Value must be an integer.".into()),
                ..invalid_format(format!("invalid value \"{copy}\" for \"{}\"", keyword.name))
            });
        }
        if result < i128::from(i32::MIN) || result > i128::from(i32::MAX) {
            return Err(PgError {
                detail: Some("Value must be in the range -2147483648 to 2147483647.".into()),
                ..PgError::new(
                    code::DATETIME_FIELD_OVERFLOW,
                    format!("value for \"{}\" in source string is out of range", keyword.name),
                )
            });
        }
        Ok((result as i64, self.at - init))
    }

    /// search reads the first of some words that the text starts with, ignoring ASCII case, returning its position.
    fn search(&mut self, words: &[&str], keyword: &Keyword) -> Result<usize> {
        let rest = &self.text[self.at..];
        let found =
            words.iter().position(|w| rest.len() >= w.len() && rest[..w.len()].eq_ignore_ascii_case(w.as_bytes()));
        match found {
            Some(index) => {
                self.at += words[index].len();
                Ok(index)
            }
            None => {
                let rest = self.rest();
                let word = rest.split(|c: char| c.is_ascii() && is_space(c as u8)).next().unwrap_or_default();
                Err(PgError {
                    detail: Some("The given value did not match any of the allowed values for this field.".into()),
                    ..invalid_format(format!("invalid value \"{word}\" for \"{}\"", keyword.name))
                })
            }
        }
    }

    /// skip_ordinal skips the two characters of an ordinal suffix after a field that has one.
    fn skip_ordinal(&mut self, suffix: u8) {
        if suffix & (TH_UPPER | TH_LOWER) != 0 {
            for _ in 0..2 {
                self.at += self.char_len();
            }
        }
    }
}

/// next_is_separator reports whether a non-digit follows the keyword at an index, as Postgres' is_next_separator does.
fn next_is_separator(nodes: &[Node], index: usize) -> bool {
    if let Node::Action(_, suffix) = &nodes[index]
        && suffix & (TH_UPPER | TH_LOWER) != 0
    {
        return true;
    }
    match nodes.get(index + 1) {
        None => true,
        Some(Node::Action(keyword, _)) => !keyword.digit,
        Some(other) => !other.text().is_some_and(|c| c.is_ascii_digit()),
    }
}

/// read_template reads text with a parsed template, as Postgres' DCH_from_char does outside standard mode.
fn read_template(nodes: &[Node], text: &str) -> Result<FromChar> {
    let mut out = FromChar::default();
    let mut r = Reader { text: text.as_bytes(), at: 0 };
    let mut fx = false;
    let mut extra_skip: i64 = 0;
    let mut index = 0;
    while index < nodes.len() && r.peek() != 0 {
        let node = &nodes[index];
        let is_fx = matches!(node, Node::Action(k, _) if k.id == Id::Fx);
        if !fx && !is_fx && (matches!(node, Node::Action(..)) || index == 0) {
            extra_skip += r.skip_spaces() as i64;
        }
        let (keyword, suffix) = match node {
            Node::Space(_) | Node::Separator(_) => {
                if !fx {
                    extra_skip -= 1;
                    let b = r.peek();
                    if is_space(b) || (b.is_ascii() && is_separator(b as char)) {
                        r.at += 1;
                        extra_skip += 1;
                    }
                } else {
                    r.at += r.char_len();
                }
                index += 1;
                continue;
            }
            Node::Char(_) => {
                if !fx && extra_skip > 0 {
                    extra_skip -= 1;
                } else {
                    r.at += r.char_len();
                }
                index += 1;
                continue;
            }
            Node::Action(keyword, suffix) => (*keyword, *suffix),
        };
        if keyword.mode != Mode::None {
            match out.mode {
                None => out.mode = Some(keyword.mode),
                Some(mode) if mode != keyword.mode => {
                    return Err(PgError {
                        hint: Some(
                            "Do not mix Gregorian and ISO week date conventions in a formatting template.".into(),
                        ),
                        ..invalid_format("invalid combination of date conventions")
                    });
                }
                _ => {}
            }
        }
        let len = keyword.name.len();
        match keyword.id {
            Id::Fx => fx = true,
            Id::MeridiemDotted | Id::Meridiem => {
                let words: &[&str] = if keyword.id == Id::Meridiem { &MERIDIEMS } else { &MERIDIEMS_DOTTED };
                let value = r.search(words, keyword)? as i64;
                set_field(&mut out.pm, value % 2, keyword)?;
                out.twelve_hour = true;
            }
            Id::Hh12 | Id::Hh24 => {
                let (value, _) = r.parse_int(2, nodes, index)?;
                set_field(&mut out.hh, value, keyword)?;
                if keyword.id == Id::Hh12 {
                    out.twelve_hour = true;
                }
                r.skip_ordinal(suffix);
            }
            Id::Mi
            | Id::Ss
            | Id::Ssss
            | Id::Mm
            | Id::Ddd
            | Id::Dd
            | Id::D
            | Id::Ww
            | Id::Iw
            | Id::Cc
            | Id::W
            | Id::J => {
                let (value, _) = r.parse_int(len, nodes, index)?;
                let field = match keyword.id {
                    Id::Mi => &mut out.mi,
                    Id::Ss => &mut out.ss,
                    Id::Ssss => &mut out.ssss,
                    Id::Mm => &mut out.mm,
                    Id::Ddd => &mut out.ddd,
                    Id::Dd => &mut out.dd,
                    Id::D => &mut out.d,
                    Id::Ww | Id::Iw => &mut out.ww,
                    Id::Cc => &mut out.cc,
                    Id::W => &mut out.w,
                    _ => &mut out.j,
                };
                set_field(field, value, keyword)?;
                r.skip_ordinal(suffix);
            }
            Id::Ms => {
                let (value, used) = r.parse_int(3, nodes, index)?;
                set_field(&mut out.ms, value, keyword)?;
                out.ms *= match used {
                    1 => 100,
                    2 => 10,
                    _ => 1,
                };
                r.skip_ordinal(suffix);
            }
            Id::Ff(_) | Id::Us => {
                if let Id::Ff(digits) = keyword.id {
                    out.ff = digits as i64;
                }
                let width = if keyword.id == Id::Us { 6 } else { out.ff as usize };
                let (value, used) = r.parse_int(width, nodes, index)?;
                set_field(&mut out.us, value, keyword)?;
                out.us *= 10i64.pow(6u32.saturating_sub(used.min(6) as u32).min(5));
                r.skip_ordinal(suffix);
            }
            Id::Tz | Id::Of => {
                return Err(PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    format!("formatting field \"{}\" is only supported in to_char", keyword.name),
                ));
            }
            Id::Tzh => {
                match r.peek() {
                    b @ (b'+' | b'-' | b' ') => {
                        out.tzsign = if b == b'-' { -1 } else { 1 };
                        r.at += 1;
                    }
                    _ => {
                        out.tzsign = if extra_skip > 0 && r.at > 0 && r.text[r.at - 1] == b'-' { -1 } else { 1 };
                    }
                }
                let (value, _) = r.parse_int(2, nodes, index)?;
                set_field(&mut out.tzh, value, keyword)?;
            }
            Id::Tzm => {
                if out.tzsign == 0 {
                    out.tzsign = 1;
                }
                let (value, _) = r.parse_int(2, nodes, index)?;
                set_field(&mut out.tzm, value, keyword)?;
            }
            Id::EraDotted | Id::Era => {
                let words: &[&str] = if keyword.id == Id::Era { &ERAS } else { &ERAS_DOTTED };
                let value = r.search(words, keyword)? as i64;
                set_field(&mut out.bc, value % 2, keyword)?;
            }
            Id::MonthName | Id::MonthAbbreviation => {
                let words: &[&str] = if keyword.id == Id::MonthName { &MONTHS } else { &MONTH_ABBREVIATIONS };
                let value = r.search(words, keyword)? as i64;
                set_field(&mut out.mm, value + 1, keyword)?;
            }
            Id::DayName | Id::DayAbbreviation => {
                let words: &[&str] = if keyword.id == Id::DayName { &DAYS } else { &DAY_ABBREVIATIONS };
                let value = r.search(words, keyword)? as i64;
                set_field(&mut out.d, value, keyword)?;
                out.d += 1;
            }
            Id::Iddd => {
                let (value, _) = r.parse_int(3, nodes, index)?;
                set_field(&mut out.ddd, value, keyword)?;
                r.skip_ordinal(suffix);
            }
            Id::IsoD => {
                let (value, _) = r.parse_int(1, nodes, index)?;
                set_field(&mut out.d, value, keyword)?;
                out.d += 1;
                if out.d > 7 {
                    out.d = 1;
                }
                r.skip_ordinal(suffix);
            }
            Id::Q => {
                r.parse_int(len, nodes, index)?;
                r.skip_ordinal(suffix);
            }
            Id::YComma => {
                let rest = &r.text[r.at..];
                let parsed = Reader::strtol(rest, 0).and_then(|(millennia, end)| {
                    if rest.get(end) != Some(&b',') {
                        return None;
                    }
                    let digits = rest[end + 1..].iter().take(3).take_while(|b| b.is_ascii_digit()).count();
                    let years: i128 = std::str::from_utf8(&rest[end + 1..end + 1 + digits]).ok()?.parse().ok()?;
                    Some((millennia * 1000 + years, end + 1 + digits))
                });
                let Some((years, used)) = parsed else {
                    return Err(invalid_format("invalid input string for \"Y,YYY\""));
                };
                set_field(&mut out.year, years as i64, keyword)?;
                out.yysz = 4;
                r.at += used;
                r.skip_ordinal(suffix);
            }
            Id::Yyyy | Id::Iyyy | Id::Yyy | Id::Iyy | Id::Yy | Id::Iy | Id::Y | Id::I => {
                let (value, used) = r.parse_int(len, nodes, index)?;
                set_field(&mut out.year, value, keyword)?;
                let size = match keyword.id {
                    Id::Yyyy | Id::Iyyy => 4,
                    Id::Yyy | Id::Iyy => 3,
                    Id::Yy | Id::Iy => 2,
                    _ => 1,
                };
                if size < 4 && used < 4 {
                    out.year = adjust_partial_year(out.year);
                }
                out.yysz = size;
                r.skip_ordinal(suffix);
            }
            Id::Rm => {
                let value = r.search(&ROMAN_MONTHS, keyword)? as i64;
                set_field(&mut out.mm, 12 - value, keyword)?;
            }
        }
        if !fx {
            extra_skip = r.skip_spaces() as i64;
        }
        index += 1;
    }
    Ok(out)
}

/// adjust_partial_year completes a year of fewer than four digits toward 2020.
fn adjust_partial_year(year: i64) -> i64 {
    match year {
        ..70 => year + 2000,
        70..100 => year + 1900,
        100..520 => year + 2000,
        520..1000 => year + 1000,
        _ => year,
    }
}

/// Parsed is a date and time that text read with a template holds, with the offset east of UTC it named.
struct Parsed {
    fields: Fields,
    precision: i64,
    offset: Option<i64>,
}

/// field_overflow returns Postgres' error for text whose fields are out of range.
fn field_overflow(text: &str) -> PgError {
    PgError::new(code::DATETIME_FIELD_OVERFLOW, format!("date/time field value out of range: \"{text}\""))
}

/// is_leap reports whether an astronomical year is a leap year.
fn is_leap(year: i64) -> bool {
    year % 4 == 0 && (year % 100 != 0 || year % 400 == 0)
}

/// read_datetime reads text with a template into a date and time, as Postgres' do_to_timestamp does.
fn read_datetime(text: &str, template: &str) -> Result<Parsed> {
    let mut tmfc =
        if template.is_empty() { FromChar::default() } else { read_template(&parse_format(template), text)? };
    let mut tm = Fields { month: 1, day: 1, ..Fields::default() };
    let (mut year_set, mut month_set, mut day_set) = (false, false, false);
    if tmfc.ssss != 0 {
        tm.hour = tmfc.ssss / 3600;
        tm.minute = tmfc.ssss % 3600 / 60;
        tm.second = tmfc.ssss % 60;
    }
    if tmfc.ss != 0 {
        tm.second = tmfc.ss;
    }
    if tmfc.mi != 0 {
        tm.minute = tmfc.mi;
    }
    if tmfc.hh != 0 {
        tm.hour = tmfc.hh;
    }
    if tmfc.twelve_hour {
        if tm.hour < 1 || tm.hour > 12 {
            return Err(PgError {
                hint: Some("Use the 24-hour clock, or give an hour between 1 and 12.".into()),
                ..invalid_format(format!("hour \"{}\" is invalid for the 12-hour clock", tm.hour))
            });
        }
        if tmfc.pm != 0 && tm.hour < 12 {
            tm.hour += 12;
        } else if tmfc.pm == 0 && tm.hour == 12 {
            tm.hour = 0;
        }
    }
    if tmfc.year != 0 {
        if tmfc.cc != 0 && tmfc.yysz <= 2 {
            if tmfc.bc != 0 {
                tmfc.cc = -tmfc.cc;
            }
            tm.year = tmfc.year % 100;
            if tm.year != 0 {
                if tmfc.cc >= 0 {
                    tm.year += (tmfc.cc - 1) * 100;
                } else {
                    tm.year = (tmfc.cc + 1) * 100 - tm.year + 1;
                }
            } else {
                tm.year = tmfc.cc * 100 + if tmfc.cc >= 0 { 0 } else { 1 };
            }
        } else {
            tm.year = if tmfc.bc != 0 { -tmfc.year } else { tmfc.year };
            if tm.year < 0 {
                tm.year += 1;
            }
        }
        year_set = true;
    } else if tmfc.cc != 0 {
        if tmfc.bc != 0 {
            tmfc.cc = -tmfc.cc;
        }
        tm.year = if tmfc.cc >= 0 { (tmfc.cc - 1) * 100 + 1 } else { tmfc.cc * 100 + 1 };
        year_set = true;
    }
    if tmfc.j != 0 {
        (tm.year, tm.month, tm.day) = dt::j2date(tmfc.j);
        (year_set, month_set, day_set) = (true, true, true);
    }
    if tmfc.ww != 0 {
        if tmfc.mode == Some(Mode::IsoWeek) {
            let mut julian = isoweek2j(tm.year, tmfc.ww);
            if tmfc.d != 0 {
                julian += if tmfc.d > 1 { tmfc.d - 2 } else { 6 };
            }
            (tm.year, tm.month, tm.day) = dt::j2date(julian);
            (year_set, month_set, day_set) = (true, true, true);
        } else {
            tmfc.ddd = (tmfc.ww - 1) * 7 + 1;
        }
    }
    if tmfc.w != 0 {
        tmfc.dd = (tmfc.w - 1) * 7 + 1;
    }
    if tmfc.dd != 0 {
        tm.day = tmfc.dd;
        day_set = true;
    }
    if tmfc.mm != 0 {
        tm.month = tmfc.mm;
        month_set = true;
    }
    if tmfc.ddd != 0 && (tm.month <= 1 || tm.day <= 1) {
        if tm.year == 0 && tmfc.bc == 0 {
            return Err(invalid_format("cannot calculate day of year without year information"));
        }
        if tmfc.mode == Some(Mode::IsoWeek) {
            let start = isoweek2j(tm.year, 1) - 1;
            (tm.year, tm.month, tm.day) = dt::j2date(start + tmfc.ddd);
            (year_set, month_set, day_set) = (true, true, true);
        } else {
            let sums: [i64; 13] = if is_leap(tm.year) {
                [0, 31, 60, 91, 121, 152, 182, 213, 244, 274, 305, 335, 366]
            } else {
                [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334, 365]
            };
            let month = (1..=12).find(|&i| tmfc.ddd <= sums[i]).unwrap_or(13);
            if tm.month <= 1 {
                tm.month = month as i64;
            }
            if tm.day <= 1 {
                tm.day = tmfc.ddd - sums[month - 1];
            }
            (month_set, day_set) = (true, true);
        }
    }
    tm.micros = tmfc.ms * 1000 + tmfc.us;
    if (month_set && !(1..=12).contains(&tm.month))
        || (day_set && !(1..=31).contains(&tm.day))
        || (year_set && month_set && day_set && tm.day > dt::days_in_month(tm.year, tm.month))
    {
        return Err(field_overflow(text));
    }
    if !(0..24).contains(&tm.hour)
        || !(0..60).contains(&tm.minute)
        || !(0..60).contains(&tm.second)
        || !(0..USECS_PER_SEC).contains(&tm.micros)
    {
        return Err(field_overflow(text));
    }
    let offset = if tmfc.tzsign != 0 {
        if !(0..=15).contains(&tmfc.tzh) || !(0..60).contains(&tmfc.tzm) {
            return Err(PgError::new(
                code::INVALID_TIME_ZONE_DISPLACEMENT,
                format!("time zone displacement out of range: \"{text}\""),
            ));
        }
        Some(tmfc.tzsign * (tmfc.tzh * 3600 + tmfc.tzm * 60))
    } else {
        None
    };
    Ok(Parsed { fields: tm, precision: tmfc.ff, offset })
}

/// to_timestamp reads text with a template as a timestamptz, in a zone unless the text names an offset.
pub fn to_timestamp(text: &str, template: &str, zone: &Zone) -> Result<i64> {
    let parsed = read_datetime(text, template)?;
    let out_of_range = || PgError::new(code::DATETIME_FIELD_OVERFLOW, "timestamp out of range");
    let local = dt::timestamp_of_fields(&parsed.fields).ok_or_else(out_of_range)?;
    let offset = parsed.offset.unwrap_or_else(|| zone.offset_for_local(local) as i64);
    let mut result = local - offset * USECS_PER_SEC;
    if parsed.precision != 0 {
        let scale = 10i64.pow(6 - parsed.precision as u32);
        let half = scale / 2;
        result = if result >= 0 { (result + half) / scale * scale } else { -((-result + half) / scale * scale) };
    }
    Ok(result)
}

/// DatetimeParts is a date and time that text read with a template holds for SQL/JSON, with the offset east of UTC
/// it named and whether the template has date, time, and zone fields.
pub struct DatetimeParts {
    pub fields: Fields,
    pub offset: Option<i64>,
    pub dated: bool,
    pub timed: bool,
    pub zoned: bool,
}

/// parse_datetime reads text with a template as Postgres' parse_datetime does for SQL/JSON's `.datetime()` method.
pub fn parse_datetime(text: &str, template: &str) -> Result<DatetimeParts> {
    let parsed = read_datetime(text, template)?;
    let (mut dated, mut timed, mut zoned) = (false, false, false);
    for node in parse_format(template) {
        let Node::Action(keyword, _) = node else { continue };
        match keyword.id {
            Id::Tzh | Id::Tzm | Id::Tz | Id::Of => zoned = true,
            Id::Hh24
            | Id::Hh12
            | Id::Mi
            | Id::Ss
            | Id::Ms
            | Id::Us
            | Id::Ff(_)
            | Id::Ssss
            | Id::Meridiem
            | Id::MeridiemDotted => timed = true,
            Id::DayName | Id::DayAbbreviation | Id::D | Id::IsoD | Id::Fx => {}
            _ => dated = true,
        }
    }
    Ok(DatetimeParts { fields: parsed.fields, offset: parsed.offset, dated, timed, zoned })
}

/// to_date reads text with a template as a date.
pub fn to_date(text: &str, template: &str) -> Result<i32> {
    let parsed = read_datetime(text, template)?;
    let f = parsed.fields;
    let out_of_range = || PgError::new(code::DATETIME_FIELD_OVERFLOW, format!("date out of range: \"{text}\""));
    let valid_julian = (f.year > -4713 || (f.year == -4713 && f.month >= 11)) && f.year < 5_874_898
        || (f.year == 5_874_898 && f.month < 6);
    if !valid_julian {
        return Err(out_of_range());
    }
    let days = dt::date2j(f.year, f.month, f.day) - POSTGRES_EPOCH_JDATE;
    if !(-POSTGRES_EPOCH_JDATE..=2_147_483_494 - POSTGRES_EPOCH_JDATE).contains(&days) {
        return Err(out_of_range());
    }
    Ok(days as i32)
}
