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

/// The number of days from 1970-01-01 to 2000-01-01, the epoch of binary dates and timestamps.
const POSTGRES_EPOCH_DAYS: i64 = 10957;
/// The microseconds in a day.
const MICROSECONDS_PER_DAY: i64 = 86_400_000_000;

/// civil_from_days returns the proleptic Gregorian (year, month, day) of a day count from 1970-01-01.
fn civil_from_days(days: i64) -> (i64, u32, u32) {
    let z = days + 719_468;
    let era = z.div_euclid(146_097);
    let day_of_era = z - era * 146_097;
    let year_of_era = (day_of_era - day_of_era / 1460 + day_of_era / 36524 - day_of_era / 146_096) / 365;
    let day_of_year = day_of_era - (365 * year_of_era + year_of_era / 4 - year_of_era / 100);
    let month_index = (5 * day_of_year + 2) / 153;
    let day = (day_of_year - (153 * month_index + 2) / 5 + 1) as u32;
    let month = if month_index < 10 { month_index + 3 } else { month_index - 9 } as u32;
    let year = year_of_era + era * 400 + if month <= 2 { 1 } else { 0 };
    (year, month, day)
}

/// date_part renders a civil date in ISO style, with years before 1 shown with BC.
fn date_part(days_since_2000: i64) -> (String, bool) {
    let (year, month, day) = civil_from_days(days_since_2000 + POSTGRES_EPOCH_DAYS);
    if year <= 0 {
        (format!("{:04}-{month:02}-{day:02}", 1 - year), true)
    } else {
        (format!("{year:04}-{month:02}-{day:02}"), false)
    }
}

/// time_part renders a time of day, dropping trailing zeros of the fraction.
fn time_part(microseconds: i64) -> String {
    let seconds = microseconds / 1_000_000;
    let fraction = microseconds % 1_000_000;
    let mut text = format!("{:02}:{:02}:{:02}", seconds / 3600, (seconds / 60) % 60, seconds % 60);
    if fraction != 0 {
        let digits = format!("{fraction:06}");
        text.push('.');
        text.push_str(digits.trim_end_matches('0'));
    }
    text
}

/// date_text renders a binary date.
pub(crate) fn date_text(days: i32) -> String {
    match days {
        i32::MAX => "infinity".to_string(),
        i32::MIN => "-infinity".to_string(),
        _ => {
            let (date, bc) = date_part(days as i64);
            if bc { format!("{date} BC") } else { date }
        }
    }
}

/// time_text renders a binary time.
pub(crate) fn time_text(microseconds: i64) -> String {
    time_part(microseconds)
}

/// timestamp_text renders a binary timestamp, or a timestamptz in UTC when with_zone is set.
pub(crate) fn timestamp_text(microseconds: i64, with_zone: bool) -> String {
    match microseconds {
        i64::MAX => return "infinity".to_string(),
        i64::MIN => return "-infinity".to_string(),
        _ => {}
    }
    let days = microseconds.div_euclid(MICROSECONDS_PER_DAY);
    let time = microseconds.rem_euclid(MICROSECONDS_PER_DAY);
    let (date, bc) = date_part(days);
    let mut text = format!("{date} {}", time_part(time));
    if with_zone {
        text.push_str("+00");
    }
    if bc {
        text.push_str(" BC");
    }
    text
}

/// interval_text renders a binary interval the way interval_out does with IntervalStyle postgres.
pub(crate) fn interval_text(months: i32, days: i32, microseconds: i64) -> String {
    let years = months / 12;
    let months = months % 12;
    let mut parts: Vec<String> = Vec::new();
    let mut is_before = false;
    let mut is_zero = true;
    let mut append = |value: i64, unit: &str, parts: &mut Vec<String>| {
        if value == 0 {
            return;
        }
        let sign = if is_before && value > 0 { "+" } else { "" };
        let plural = if value == 1 { "" } else { "s" };
        parts.push(format!("{sign}{value} {unit}{plural}"));
        is_before = value < 0;
        is_zero = false;
    };
    append(years as i64, "year", &mut parts);
    let month_unit = "mon";
    append(months as i64, month_unit, &mut parts);
    append(days as i64, "day", &mut parts);
    let mut text = parts.join(" ");
    if is_zero || microseconds != 0 {
        let negative = microseconds < 0;
        let absolute = microseconds.unsigned_abs();
        let hours = absolute / 3_600_000_000;
        let minutes = (absolute / 60_000_000) % 60;
        let seconds = (absolute / 1_000_000) % 60;
        let fraction = absolute % 1_000_000;
        let sign = if negative {
            "-"
        } else if is_before {
            "+"
        } else {
            ""
        };
        let mut time = format!("{sign}{hours:02}:{minutes:02}:{seconds:02}");
        if fraction != 0 {
            let digits = format!("{fraction:06}");
            time.push('.');
            time.push_str(digits.trim_end_matches('0'));
        }
        if !text.is_empty() {
            text.push(' ');
        }
        text.push_str(&time);
    }
    text
}
