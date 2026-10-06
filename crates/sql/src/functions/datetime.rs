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

//! Date and time functions and operators.

use super::{Function, text};
use crate::datetime::{
    self as dt, DATE_NOBEGIN, DATE_NOEND, Fields, Interval, POSTGRES_EPOCH_JDATE, TIMESTAMP_NOBEGIN, TIMESTAMP_NOEND,
    USECS_PER_DAY, USECS_PER_HOUR, USECS_PER_MINUTE, USECS_PER_SEC, Zone,
};
use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;
use crate::oid::{DATE, FLOAT8, INT4, INTERVAL, NUMERIC, TEXT, TIME, TIMESTAMP, TIMESTAMPTZ, TIMETZ};
use crate::query::Ctx;
use crate::types::Value;

/// f declares a strict datetime function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the date and time functions.
pub const FUNCTIONS: &[Function] = &[
    f("now", &[], TIMESTAMPTZ, now),
    f("transaction_timestamp", &[], TIMESTAMPTZ, now),
    f("current_timestamp", &[], TIMESTAMPTZ, now),
    f("statement_timestamp", &[], TIMESTAMPTZ, clock_timestamp),
    f("clock_timestamp", &[], TIMESTAMPTZ, clock_timestamp),
    f("localtimestamp", &[], TIMESTAMP, localtimestamp),
    f("current_date", &[], DATE, current_date),
    f("current_time", &[], TIMETZ, current_time),
    f("localtime", &[], TIME, localtime),
    f("timeofday", &[], TEXT, timeofday),
    f("extract", &[TEXT, TIMESTAMP], NUMERIC, extract),
    f("extract", &[TEXT, TIMESTAMPTZ], NUMERIC, extract),
    f("extract", &[TEXT, DATE], NUMERIC, extract),
    f("extract", &[TEXT, TIME], NUMERIC, extract),
    f("extract", &[TEXT, TIMETZ], NUMERIC, extract),
    f("extract", &[TEXT, INTERVAL], NUMERIC, extract),
    f("date_part", &[TEXT, TIMESTAMP], FLOAT8, date_part),
    f("date_part", &[TEXT, TIMESTAMPTZ], FLOAT8, date_part),
    f("date_part", &[TEXT, DATE], FLOAT8, date_part),
    f("date_part", &[TEXT, TIME], FLOAT8, date_part),
    f("date_part", &[TEXT, TIMETZ], FLOAT8, date_part),
    f("date_part", &[TEXT, INTERVAL], FLOAT8, date_part),
    f("date_trunc", &[TEXT, TIMESTAMP], TIMESTAMP, date_trunc),
    f("date_trunc", &[TEXT, TIMESTAMPTZ], TIMESTAMPTZ, date_trunc),
    f("date_trunc", &[TEXT, INTERVAL], INTERVAL, date_trunc),
    f("age", &[TIMESTAMP, TIMESTAMP], INTERVAL, age),
    f("age", &[TIMESTAMPTZ, TIMESTAMPTZ], INTERVAL, age),
    f("age", &[TIMESTAMP], INTERVAL, age_now),
    f("age", &[TIMESTAMPTZ], INTERVAL, age_now),
    f("make_date", &[INT4, INT4, INT4], DATE, make_date),
    f("make_time", &[INT4, INT4, FLOAT8], TIME, make_time),
    f("make_timestamp", &[INT4, INT4, INT4, INT4, INT4, FLOAT8], TIMESTAMP, make_timestamp),
    f("make_timestamptz", &[INT4, INT4, INT4, INT4, INT4, FLOAT8], TIMESTAMPTZ, make_timestamptz),
    f("make_timestamptz", &[INT4, INT4, INT4, INT4, INT4, FLOAT8, TEXT], TIMESTAMPTZ, make_timestamptz),
    f("make_interval", &[INT4, INT4, INT4, INT4, INT4, INT4, FLOAT8], INTERVAL, make_interval),
    f("to_timestamp", &[FLOAT8], TIMESTAMPTZ, to_timestamp_epoch),
    f("justify_days", &[INTERVAL], INTERVAL, justify_days),
    f("justify_hours", &[INTERVAL], INTERVAL, justify_hours),
    f("justify_interval", &[INTERVAL], INTERVAL, justify_interval),
    f("isfinite", &[DATE], crate::oid::BOOL, isfinite),
    f("isfinite", &[TIMESTAMP], crate::oid::BOOL, isfinite),
    f("isfinite", &[TIMESTAMPTZ], crate::oid::BOOL, isfinite),
    f("isfinite", &[INTERVAL], crate::oid::BOOL, isfinite),
    f("timezone", &[TEXT, TIMESTAMPTZ], TIMESTAMP, timezone_of_timestamptz),
    f("timezone", &[TEXT, TIMESTAMP], TIMESTAMPTZ, timezone_of_timestamp),
    f("timezone", &[INTERVAL, TIMESTAMPTZ], TIMESTAMP, timezone_of_timestamptz),
    f("timezone", &[INTERVAL, TIMESTAMP], TIMESTAMPTZ, timezone_of_timestamp),
];

/// now returns the transaction's start.
fn now(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::TimestampTz(ctx.txn.started))
}

/// clock_timestamp returns the current time.
fn clock_timestamp(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::TimestampTz(dt::clock()))
}

/// zone_offset returns the session zone's offset east of UTC at a UTC timestamp, in microseconds.
fn zone_offset(utc: i64) -> i64 {
    dt::with_format(|f| f.zone.offset_at(utc).0) as i64 * USECS_PER_SEC
}

/// local_offset returns the session zone's offset east of UTC for a local timestamp, in microseconds.
fn local_offset(local: i64) -> i64 {
    dt::with_format(|f| f.zone.offset_for_local(local)) as i64 * USECS_PER_SEC
}

/// localtimestamp returns the transaction's start in the session's zone.
fn localtimestamp(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    let start = ctx.txn.started;
    Ok(Value::Timestamp(start + zone_offset(start)))
}

/// current_date returns the transaction's start date in the session's zone.
fn current_date(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    let start = ctx.txn.started;
    Ok(Value::Date((start + zone_offset(start)).div_euclid(USECS_PER_DAY) as i32))
}

/// current_time returns the transaction's start time with the session's zone.
fn current_time(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    let start = ctx.txn.started;
    let offset = zone_offset(start);
    Ok(Value::TimeTz((start + offset).rem_euclid(USECS_PER_DAY), -(offset / USECS_PER_SEC) as i32))
}

/// localtime returns the transaction's start time in the session's zone.
fn localtime(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    let start = ctx.txn.started;
    Ok(Value::Time((start + zone_offset(start)).rem_euclid(USECS_PER_DAY)))
}

/// timeofday returns the current time as text, in the Postgres style.
fn timeofday(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    let now = dt::clock();
    let (offset, name) = dt::with_format(|f| f.zone.offset_at(now));
    let local = dt::fields_of_timestamp(now + offset as i64 * USECS_PER_SEC);
    const MONTHS: [&str; 12] = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
    const DAYS: [&str; 7] = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
    let dow = DAYS[(dt::date2j(local.year, local.month, local.day) + 1).rem_euclid(7) as usize];
    Ok(Value::Text(format!(
        "{dow} {} {:02} {:02}:{:02}:{:02}.{:06} {} {name}",
        MONTHS[local.month as usize - 1],
        local.day,
        local.hour,
        local.minute,
        local.second,
        local.micros,
        local.year
    )))
}

/// unit_name normalizes a field or truncation unit name, as Postgres' DecodeUnits and DecodeSpecial read them.
fn unit_name(name: &str) -> Option<&'static str> {
    Some(match name.trim().to_ascii_lowercase().as_str() {
        "microsecond" | "microseconds" | "us" | "usec" | "usecs" | "useconds" => "microseconds",
        "millisecond" | "milliseconds" | "ms" | "msec" | "msecs" | "mseconds" => "milliseconds",
        "second" | "seconds" | "s" | "sec" | "secs" => "second",
        "minute" | "minutes" | "m" | "min" | "mins" => "minute",
        "hour" | "hours" | "h" | "hr" | "hrs" => "hour",
        "day" | "days" | "d" => "day",
        "week" | "weeks" | "w" => "week",
        "month" | "months" | "mon" | "mons" => "month",
        "quarter" | "qtr" => "quarter",
        "year" | "years" | "y" | "yr" | "yrs" => "year",
        "decade" | "decades" | "dec" | "decs" => "decade",
        "century" | "centuries" | "c" | "cent" => "century",
        "millennium" | "millennia" | "mil" | "mils" => "millennium",
        "epoch" => "epoch",
        "dow" => "dow",
        "isodow" => "isodow",
        "doy" => "doy",
        "isoyear" => "isoyear",
        "julian" | "j" => "julian",
        "timezone" | "tz" => "timezone",
        "timezone_h" | "timezone_hour" => "timezone_hour",
        "timezone_m" | "timezone_minute" => "timezone_minute",
        _ => return None,
    })
}

/// type_name names a datetime argument's type for errors.
fn type_name(value: &Value) -> &'static str {
    match value {
        Value::Date(_) => "date",
        Value::Time(_) => "time without time zone",
        Value::TimeTz(..) => "time with time zone",
        Value::Timestamp(_) => "timestamp without time zone",
        Value::TimestampTz(_) => "timestamp with time zone",
        _ => "interval",
    }
}

/// unsupported_unit returns Postgres' error for a unit a type doesn't support.
fn unsupported_unit(unit: &str, value: &Value) -> PgError {
    PgError::new(code::FEATURE_NOT_SUPPORTED, format!("unit \"{unit}\" not supported for type {}", type_name(value)))
}

/// unrecognized_unit returns Postgres' error for an unknown unit.
fn unrecognized_unit(unit: &str, value: &Value) -> PgError {
    PgError::new(code::INVALID_PARAMETER_VALUE, format!("unit \"{unit}\" not recognized for type {}", type_name(value)))
}

/// iso_week returns the ISO 8601 week number and week-numbering year of a date.
fn iso_week(year: i64, month: i64, day: i64) -> (i64, i64) {
    let jd = dt::date2j(year, month, day);
    // ISO weeks start on Monday, and week 1 holds January 4th.
    let week_start = |y: i64| {
        let jan4 = dt::date2j(y, 1, 4);
        jan4 - (jan4 % 7)
    };
    let mut iso_year = year;
    if jd < week_start(year) {
        iso_year -= 1;
    } else if jd >= week_start(year + 1) {
        iso_year += 1;
    }
    ((jd - week_start(iso_year)) / 7 + 1, iso_year)
}

/// numeric_with_micros returns a whole value plus microseconds as a numeric with six decimals.
fn numeric_with_micros(micros: i128) -> Numeric {
    Numeric::parse(&format!(
        "{}{}.{:06}",
        if micros < 0 { "-" } else { "" },
        micros.abs() / 1_000_000,
        micros.abs() % 1_000_000
    ))
    .unwrap_or(Numeric::NaN)
}

/// extract_field computes a field of a datetime value as a numeric, as Postgres' extract does.
fn extract_field(unit_text: &str, value: &Value) -> Result<Numeric> {
    let unit = unit_name(unit_text).ok_or_else(|| unrecognized_unit(unit_text, value))?;
    let int = |n: i64| Numeric::from_i64(n);
    let infinite = match value {
        Value::Date(d) => *d == DATE_NOBEGIN || *d == DATE_NOEND,
        Value::Timestamp(t) | Value::TimestampTz(t) => *t == TIMESTAMP_NOBEGIN || *t == TIMESTAMP_NOEND,
        _ => false,
    };
    if infinite {
        let positive = matches!(
            value,
            Value::Date(DATE_NOEND) | Value::Timestamp(TIMESTAMP_NOEND) | Value::TimestampTz(TIMESTAMP_NOEND)
        );
        return Ok(match unit {
            "epoch" | "julian" | "year" | "decade" | "century" | "millennium" | "isoyear" => {
                if positive {
                    Numeric::Infinity
                } else {
                    Numeric::NegativeInfinity
                }
            }
            _ => return Ok(Numeric::NaN),
        });
    }
    match value {
        Value::Interval(iv) => {
            let (years, months) = (iv.months as i64 / 12, iv.months as i64 % 12);
            return Ok(match unit {
                "microseconds" => int(iv.micros % USECS_PER_MINUTE),
                "milliseconds" => {
                    numeric_with_micros((iv.micros % USECS_PER_MINUTE) as i128 * 1000).div_exact(0).with_scale(3)
                }
                "second" => numeric_with_micros((iv.micros % USECS_PER_MINUTE) as i128),
                "minute" => int(iv.micros / USECS_PER_MINUTE % 60),
                "hour" => int(iv.micros / USECS_PER_HOUR),
                "day" => int(iv.days as i64),
                "month" => int(months),
                "quarter" => int(months / 3 + 1),
                "year" => int(years),
                "decade" => int(years / 10),
                "century" => int(years / 100),
                "millennium" => int(years / 1000),
                "epoch" => {
                    let micros = iv.micros as i128
                        + iv.days as i128 * USECS_PER_DAY as i128
                        + (iv.months as i128 / 12) * 31_557_600_000_000
                        + (iv.months as i128 % 12) * 30 * USECS_PER_DAY as i128;
                    numeric_with_micros(micros)
                }
                _ => return Err(unsupported_unit(unit_text, value)),
            });
        }
        Value::Time(_) | Value::TimeTz(..) => {
            let (time, zone) = match value {
                Value::TimeTz(t, z) => (*t, Some(*z)),
                Value::Time(t) => (*t, None),
                _ => unreachable!(),
            };
            return Ok(match unit {
                "microseconds" => int(time % USECS_PER_MINUTE),
                "milliseconds" => numeric_with_micros((time % USECS_PER_MINUTE) as i128 * 1000).with_scale(3),
                "second" => numeric_with_micros((time % USECS_PER_MINUTE) as i128),
                "minute" => int(time / USECS_PER_MINUTE % 60),
                "hour" => int(time / USECS_PER_HOUR),
                "epoch" => numeric_with_micros((time + zone.unwrap_or(0) as i64 * USECS_PER_SEC) as i128),
                "timezone" if zone.is_some() => int(-zone.unwrap_or(0) as i64),
                "timezone_hour" if zone.is_some() => int(-zone.unwrap_or(0) as i64 / 3600),
                "timezone_minute" if zone.is_some() => int(-zone.unwrap_or(0) as i64 / 60 % 60),
                _ => return Err(unsupported_unit(unit_text, value)),
            });
        }
        _ => {}
    }
    let (local, utc, offset) = match value {
        Value::Date(d) => (*d as i64 * USECS_PER_DAY, None, None),
        Value::Timestamp(t) => (*t, None, None),
        Value::TimestampTz(t) => {
            let offset = zone_offset(*t);
            (*t + offset, Some(*t), Some(offset / USECS_PER_SEC))
        }
        _ => return Err(unsupported_unit(unit_text, value)),
    };
    let date_only = matches!(value, Value::Date(_));
    let f = dt::fields_of_timestamp(local);
    let jd = dt::date2j(f.year, f.month, f.day);
    let display_year = if f.year > 0 { f.year } else { f.year - 1 };
    let time_unit = matches!(
        unit,
        "microseconds"
            | "milliseconds"
            | "second"
            | "minute"
            | "hour"
            | "timezone"
            | "timezone_hour"
            | "timezone_minute"
    );
    if date_only && time_unit {
        return Err(unsupported_unit(unit_text, value));
    }
    Ok(match unit {
        "microseconds" => int(f.second * USECS_PER_SEC + f.micros),
        "milliseconds" => numeric_with_micros((f.second * USECS_PER_SEC + f.micros) as i128 * 1000).with_scale(3),
        "second" => numeric_with_micros((f.second * USECS_PER_SEC + f.micros) as i128),
        "minute" => int(f.minute),
        "hour" => int(f.hour),
        "day" => int(f.day),
        "month" => int(f.month),
        "quarter" => int((f.month - 1) / 3 + 1),
        "week" => int(iso_week(f.year, f.month, f.day).0),
        "year" => int(display_year),
        "isoyear" => {
            let y = iso_week(f.year, f.month, f.day).1;
            int(if y > 0 { y } else { y - 1 })
        }
        "decade" => int(if f.year >= 0 { f.year / 10 } else { -((8 - (f.year - 1)) / 10) }),
        "century" => int(if f.year > 0 { (f.year + 99) / 100 } else { -((99 - (f.year - 1)) / 100) }),
        "millennium" => int(if f.year > 0 { (f.year + 999) / 1000 } else { -((999 - (f.year - 1)) / 1000) }),
        "dow" => int((jd + 1).rem_euclid(7)),
        "isodow" => int({
            let d = (jd + 1).rem_euclid(7);
            if d == 0 { 7 } else { d }
        }),
        "doy" => int(jd - dt::date2j(f.year, 1, 1) + 1),
        "julian" => {
            if date_only {
                int(jd)
            } else {
                let fraction =
                    (f.hour * USECS_PER_HOUR + f.minute * USECS_PER_MINUTE + f.second * USECS_PER_SEC + f.micros)
                        as i128;
                Numeric::from_i64(jd).add(
                    &Numeric::parse(&format!("{fraction}"))
                        .unwrap_or(Numeric::zero(0))
                        .div(&Numeric::from_i64(USECS_PER_DAY))?,
                )
            }
        }
        "epoch" => {
            let base = utc.unwrap_or(local) - dt::UNIX_EPOCH_DAYS * USECS_PER_DAY;
            if date_only { int(base / USECS_PER_SEC) } else { numeric_with_micros(base as i128) }
        }
        "timezone" => int(offset.ok_or_else(|| unsupported_unit(unit_text, value))?),
        "timezone_hour" => int(offset.ok_or_else(|| unsupported_unit(unit_text, value))? / 3600),
        "timezone_minute" => int(offset.ok_or_else(|| unsupported_unit(unit_text, value))? / 60 % 60),
        _ => return Err(unsupported_unit(unit_text, value)),
    })
}

/// extract returns a field of a datetime value as a numeric.
fn extract(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Numeric(extract_field(text(&args[0]), &args[1])?))
}

/// date_part returns a field of a datetime value as a float.
fn date_part(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(extract_field(text(&args[0]), &args[1])?.to_f64()))
}

/// truncate_fields truncates broken-down fields to a unit.
fn truncate_fields(f: &mut Fields, unit: &str) -> bool {
    match unit {
        "millennium" => {
            f.year = if f.year > 0 {
                ((f.year + 999) / 1000) * 1000 - 999
            } else {
                -((999 - (f.year - 1)) / 1000) * 1000 + 1
            };
            f.month = 1;
            f.day = 1;
        }
        "century" => {
            f.year = if f.year > 0 { ((f.year + 99) / 100) * 100 - 99 } else { -((99 - (f.year - 1)) / 100) * 100 + 1 };
            f.month = 1;
            f.day = 1;
        }
        "decade" => {
            f.year = if f.year > 0 { (f.year / 10) * 10 } else { -((8 - (f.year - 1)) / 10) * 10 };
            f.month = 1;
            f.day = 1;
        }
        "year" => {
            f.month = 1;
            f.day = 1;
        }
        "quarter" => {
            f.month = 3 * ((f.month - 1) / 3) + 1;
            f.day = 1;
        }
        "month" => f.day = 1,
        "week" => {
            let jd = dt::date2j(f.year, f.month, f.day);
            let monday = jd - jd.rem_euclid(7);
            (f.year, f.month, f.day) = dt::j2date(monday);
        }
        "day" => {}
        "hour" | "minute" | "second" | "milliseconds" | "microseconds" => {}
        _ => return false,
    }
    match unit {
        "millennium" | "century" | "decade" | "year" | "quarter" | "month" | "week" | "day" => {
            f.hour = 0;
            f.minute = 0;
            f.second = 0;
            f.micros = 0;
        }
        "hour" => {
            f.minute = 0;
            f.second = 0;
            f.micros = 0;
        }
        "minute" => {
            f.second = 0;
            f.micros = 0;
        }
        "second" => f.micros = 0,
        "milliseconds" => f.micros = f.micros / 1000 * 1000,
        _ => {}
    }
    true
}

/// date_trunc truncates a timestamp or interval to a unit, in the session's zone for a timestamptz.
fn date_trunc(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let unit_text = text(&args[0]);
    let value = &args[1];
    let unit = unit_name(unit_text).ok_or_else(|| unrecognized_unit(unit_text, value))?;
    match value {
        Value::Timestamp(ts) | Value::TimestampTz(ts) if *ts == TIMESTAMP_NOBEGIN || *ts == TIMESTAMP_NOEND => {
            Ok(value.clone())
        }
        Value::Timestamp(ts) => {
            let mut f = dt::fields_of_timestamp(*ts);
            if !truncate_fields(&mut f, unit) {
                return Err(unsupported_unit(unit_text, value));
            }
            Ok(Value::Timestamp(dt::timestamp_of_fields(&f).ok_or_else(timestamp_out_of_range)?))
        }
        Value::TimestampTz(ts) => {
            let mut f = dt::fields_of_timestamp(*ts + zone_offset(*ts));
            if !truncate_fields(&mut f, unit) {
                return Err(unsupported_unit(unit_text, value));
            }
            let local = dt::timestamp_of_fields(&f).ok_or_else(timestamp_out_of_range)?;
            let utc = if matches!(unit, "hour" | "minute" | "second" | "milliseconds" | "microseconds") {
                *ts - (*ts + zone_offset(*ts) - local)
            } else {
                local - local_offset(local)
            };
            Ok(Value::TimestampTz(utc))
        }
        Value::Interval(iv) => {
            let (mut years, mut months) = (iv.months / 12, iv.months % 12);
            let (mut days, mut micros) = (iv.days, iv.micros);
            let (hours, minutes, seconds, us) = (
                micros / USECS_PER_HOUR,
                micros / USECS_PER_MINUTE % 60,
                micros / USECS_PER_SEC % 60,
                micros % USECS_PER_SEC,
            );
            let time =
                |h: i64, m: i64, s: i64, u: i64| h * USECS_PER_HOUR + m * USECS_PER_MINUTE + s * USECS_PER_SEC + u;
            match unit {
                "millennium" => {
                    years = years / 1000 * 1000;
                    months = 0;
                    days = 0;
                    micros = 0;
                }
                "century" => {
                    years = years / 100 * 100;
                    months = 0;
                    days = 0;
                    micros = 0;
                }
                "decade" => {
                    years = years / 10 * 10;
                    months = 0;
                    days = 0;
                    micros = 0;
                }
                "year" => {
                    months = 0;
                    days = 0;
                    micros = 0;
                }
                "quarter" => {
                    months = 3 * (months / 3);
                    days = 0;
                    micros = 0;
                }
                "month" => {
                    days = 0;
                    micros = 0;
                }
                "day" => micros = 0,
                "hour" => micros = time(hours, 0, 0, 0),
                "minute" => micros = time(hours, minutes, 0, 0),
                "second" => micros = time(hours, minutes, seconds, 0),
                "milliseconds" => micros = time(hours, minutes, seconds, us / 1000 * 1000),
                "microseconds" => {}
                _ => return Err(unsupported_unit(unit_text, value)),
            }
            Ok(Value::Interval(Interval { months: years * 12 + months, days, micros }))
        }
        _ => Err(unsupported_unit(unit_text, value)),
    }
}

/// timestamp_out_of_range returns Postgres' error for a timestamp result out of range.
fn timestamp_out_of_range() -> PgError {
    PgError::new(code::DATETIME_FIELD_OVERFLOW, "timestamp out of range")
}

/// age_between returns the symbolic difference of two local timestamps in years, months, days, and time, as
/// Postgres' age does.
fn age_between(a: i64, b: i64) -> Interval {
    let (fa, fb) = (dt::fields_of_timestamp(a), dt::fields_of_timestamp(b));
    let mut micros = (fa.hour - fb.hour) * USECS_PER_HOUR
        + (fa.minute - fb.minute) * USECS_PER_MINUTE
        + (fa.second - fb.second) * USECS_PER_SEC
        + (fa.micros - fb.micros);
    let mut day = fa.day - fb.day;
    let mut month = fa.month - fb.month;
    let mut year = fa.year - fb.year;
    // Borrow from the larger units so that every unit shares the sign of the whole.
    let negative = a < b;
    let flip = |x: i64| if negative { -x } else { x };
    let (mut m, mut d, mut y, mut t) = (flip(month), flip(day), flip(year), flip(micros));
    while t < 0 {
        t += USECS_PER_DAY;
        d -= 1;
    }
    while d < 0 {
        // Borrow the length of the month before the later date.
        let (ref_year, ref_month) = if negative { (fa.year, fa.month) } else { (fb.year, fb.month) };
        d += dt::days_in_month(ref_year, ref_month);
        m -= 1;
    }
    while m < 0 {
        m += 12;
        y -= 1;
    }
    (month, day, year, micros) = (flip(m), flip(d), flip(y), flip(t));
    Interval { months: (year * 12 + month) as i32, days: day as i32, micros }
}

/// age returns the symbolic difference of two timestamps.
fn age(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (a, b) = match (&args[0], &args[1]) {
        (Value::Timestamp(a), Value::Timestamp(b)) => (*a, *b),
        (Value::TimestampTz(a), Value::TimestampTz(b)) => (*a + zone_offset(*a), *b + zone_offset(*b)),
        _ => return Ok(Value::Null),
    };
    Ok(Value::Interval(age_between(a, b)))
}

/// age_now returns the symbolic difference from a timestamp to midnight of the current date.
fn age_now(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let start = ctx.txn.started;
    let today = (start + zone_offset(start)).div_euclid(USECS_PER_DAY) * USECS_PER_DAY;
    let b = match &args[0] {
        Value::Timestamp(b) => *b,
        Value::TimestampTz(b) => *b + zone_offset(*b),
        _ => return Ok(Value::Null),
    };
    Ok(Value::Interval(age_between(today, b)))
}

/// int returns an int4 argument.
fn int(value: &Value) -> i64 {
    match value {
        Value::Int4(i) => *i as i64,
        _ => 0,
    }
}

/// float returns a float8 argument.
fn float(value: &Value) -> f64 {
    match value {
        Value::Float8(f) => *f,
        _ => 0.0,
    }
}

/// make_date_julian returns the Julian day of a date, failing as Postgres' make_date does.
fn make_date_julian(year: i64, month: i64, day: i64) -> Result<i64> {
    let invalid = || {
        PgError::new(
            code::DATETIME_FIELD_OVERFLOW,
            format!("date field value out of range: {year}-{month:02}-{day:02}"),
        )
    };
    if year == 0 || !(1..=12).contains(&month) {
        return Err(invalid());
    }
    let astronomical = if year < 0 { year + 1 } else { year };
    if day < 1 || day > dt::days_in_month(astronomical, month) {
        return Err(invalid());
    }
    Ok(dt::date2j(astronomical, month, day))
}

/// make_date builds a date.
fn make_date(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let jd = make_date_julian(int(&args[0]), int(&args[1]), int(&args[2]))?;
    Ok(Value::Date((jd - POSTGRES_EPOCH_JDATE) as i32))
}

/// time_micros builds a time of day from hours, minutes, and fractional seconds, failing out of range.
fn time_micros(hour: i64, minute: i64, seconds: f64) -> Result<i64> {
    let micros = (seconds * USECS_PER_SEC as f64).round() as i64;
    let total = hour * USECS_PER_HOUR + minute * USECS_PER_MINUTE + micros;
    if !(0..=23).contains(&hour) && !(hour == 24 && minute == 0 && micros == 0)
        || !(0..=59).contains(&minute)
        || !(0.0..=60.0).contains(&seconds)
        || total > USECS_PER_DAY
    {
        return Err(PgError::new(
            code::DATETIME_FIELD_OVERFLOW,
            format!(
                "time field value out of range: {hour}:{minute:02}:{}",
                crate::types::Value::Float8(seconds).output().unwrap_or_default()
            ),
        ));
    }
    Ok(total)
}

/// make_time builds a time of day.
fn make_time(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Time(time_micros(int(&args[0]), int(&args[1]), float(&args[2]))?))
}

/// make_local builds a local timestamp from its fields.
fn make_local(args: &[Value]) -> Result<i64> {
    let jd = make_date_julian(int(&args[0]), int(&args[1]), int(&args[2]))?;
    let time = time_micros(int(&args[3]), int(&args[4]), float(&args[5]))?;
    Ok((jd - POSTGRES_EPOCH_JDATE) * USECS_PER_DAY + time)
}

/// make_timestamp builds a timestamp.
fn make_timestamp(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Timestamp(make_local(args)?))
}

/// make_timestamptz builds a timestamptz in the session's zone or the named one.
fn make_timestamptz(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let local = make_local(args)?;
    let offset = match args.get(6) {
        Some(zone) => named_zone(text(zone))?.offset_for_local(local) as i64 * USECS_PER_SEC,
        None => local_offset(local),
    };
    Ok(Value::TimestampTz(local - offset))
}

/// named_zone returns a time zone by name, failing as Postgres does for an unknown one.
fn named_zone(name: &str) -> Result<Zone> {
    Zone::named(name)
        .ok_or_else(|| PgError::new(code::INVALID_PARAMETER_VALUE, format!("time zone \"{name}\" not recognized")))
}

/// make_interval builds an interval from years, months, weeks, days, hours, minutes, and seconds.
fn make_interval(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let micros = int(&args[4]) * USECS_PER_HOUR
        + int(&args[5]) * USECS_PER_MINUTE
        + (float(&args[6]) * USECS_PER_SEC as f64).round() as i64;
    Ok(Value::Interval(Interval {
        months: (int(&args[0]) * 12 + int(&args[1])) as i32,
        days: (int(&args[2]) * 7 + int(&args[3])) as i32,
        micros,
    }))
}

/// to_timestamp converts Unix seconds to a timestamptz.
fn to_timestamp_epoch(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let seconds = float(&args[0]);
    if seconds.is_nan() {
        return Err(PgError::new(code::DATETIME_FIELD_OVERFLOW, "timestamp cannot be NaN"));
    }
    if seconds.is_infinite() {
        return Ok(Value::TimestampTz(if seconds > 0.0 { TIMESTAMP_NOEND } else { TIMESTAMP_NOBEGIN }));
    }
    let micros = (seconds * USECS_PER_SEC as f64).round();
    if micros.abs() > 9.2e18 {
        return Err(PgError::new(
            code::DATETIME_FIELD_OVERFLOW,
            format!("timestamp out of range: \"{}\"", Value::Float8(seconds).output().unwrap_or_default()),
        ));
    }
    Ok(Value::TimestampTz(micros as i64 + dt::UNIX_EPOCH_DAYS * USECS_PER_DAY))
}

/// interval_of returns an interval argument.
fn interval_of(value: &Value) -> Interval {
    match value {
        Value::Interval(iv) => *iv,
        _ => Interval::default(),
    }
}

/// justify_days_of turns each 30 days into a month.
pub fn justify_days_of(iv: Interval) -> Interval {
    let mut result = iv;
    let whole = result.days / 30;
    result.days -= whole * 30;
    result.months += whole;
    if result.months > 0 && result.days < 0 {
        result.days += 30;
        result.months -= 1;
    } else if result.months < 0 && result.days > 0 {
        result.days -= 30;
        result.months += 1;
    }
    result
}

/// justify_hours_of turns each 24 hours into a day.
pub fn justify_hours_of(iv: Interval) -> Interval {
    let mut result = iv;
    let whole = result.micros / USECS_PER_DAY;
    result.micros -= whole * USECS_PER_DAY;
    result.days += whole as i32;
    if result.days > 0 && result.micros < 0 {
        result.micros += USECS_PER_DAY;
        result.days -= 1;
    } else if result.days < 0 && result.micros > 0 {
        result.micros -= USECS_PER_DAY;
        result.days += 1;
    }
    result
}

/// justify_days turns each 30 days into a month.
fn justify_days(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Interval(justify_days_of(interval_of(&args[0]))))
}

/// justify_hours turns each 24 hours into a day.
fn justify_hours(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Interval(justify_hours_of(interval_of(&args[0]))))
}

/// justify_interval justifies days and hours together.
fn justify_interval(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let iv = interval_of(&args[0]);
    let mut result = iv;
    let whole_days = result.micros / USECS_PER_DAY;
    result.micros -= whole_days * USECS_PER_DAY;
    result.days += whole_days as i32;
    let whole_months = result.days / 30;
    result.days -= whole_months * 30;
    result.months += whole_months;
    if result.months > 0 && (result.days < 0 || (result.days == 0 && result.micros < 0)) {
        result.days += 30;
        result.months -= 1;
    } else if result.months < 0 && (result.days > 0 || (result.days == 0 && result.micros > 0)) {
        result.days -= 30;
        result.months += 1;
    }
    if result.days > 0 && result.micros < 0 {
        result.micros += USECS_PER_DAY;
        result.days -= 1;
    } else if result.days < 0 && result.micros > 0 {
        result.micros -= USECS_PER_DAY;
        result.days += 1;
    }
    Ok(Value::Interval(result))
}

/// isfinite reports whether a value is neither infinity.
fn isfinite(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(match &args[0] {
        Value::Date(d) => *d != DATE_NOBEGIN && *d != DATE_NOEND,
        Value::Timestamp(t) | Value::TimestampTz(t) => *t != TIMESTAMP_NOBEGIN && *t != TIMESTAMP_NOEND,
        _ => true,
    }))
}

/// argument_zone returns the zone an AT TIME ZONE argument names: a zone name or an interval offset.
fn argument_zone(value: &Value) -> Result<Zone> {
    match value {
        Value::Interval(iv) => Ok(Zone::Fixed { offset: (iv.micros / USECS_PER_SEC) as i32, name: String::new() }),
        other => named_zone(text(other)),
    }
}

/// timezone_of_timestamptz returns a timestamptz's local time in a zone, for AT TIME ZONE.
fn timezone_of_timestamptz(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let zone = argument_zone(&args[0])?;
    let Value::TimestampTz(ts) = args[1] else { return Ok(Value::Null) };
    if ts == TIMESTAMP_NOBEGIN || ts == TIMESTAMP_NOEND {
        return Ok(Value::Timestamp(ts));
    }
    Ok(Value::Timestamp(ts + zone.offset_at(ts).0 as i64 * USECS_PER_SEC))
}

/// timezone_of_timestamp interprets a timestamp as local time in a zone, for AT TIME ZONE.
fn timezone_of_timestamp(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let zone = argument_zone(&args[0])?;
    let Value::Timestamp(ts) = args[1] else { return Ok(Value::Null) };
    if ts == TIMESTAMP_NOBEGIN || ts == TIMESTAMP_NOEND {
        return Ok(Value::TimestampTz(ts));
    }
    Ok(Value::TimestampTz(ts - zone.offset_for_local(ts) as i64 * USECS_PER_SEC))
}

/// add_months adds months and days to a local timestamp as Postgres does, clamping the day to the month's end.
pub fn add_months_days(local: i64, months: i32, days: i32) -> Result<i64> {
    let mut f = dt::fields_of_timestamp(local);
    if months != 0 {
        let total = f.year * 12 + (f.month - 1) + months as i64;
        f.year = total.div_euclid(12);
        f.month = total.rem_euclid(12) + 1;
        f.day = f.day.min(dt::days_in_month(f.year, f.month));
    }
    let mut ts = dt::timestamp_of_fields(&f).ok_or_else(timestamp_out_of_range)?;
    if days != 0 {
        ts = ts.checked_add(days as i64 * USECS_PER_DAY).ok_or_else(timestamp_out_of_range)?;
    }
    Ok(ts)
}

/// timestamp_plus_interval adds an interval to a timestamp, adding months and days in the session's local time for
/// a timestamptz.
pub fn timestamp_plus_interval(ts: i64, iv: Interval, with_zone: bool) -> Result<i64> {
    if ts == TIMESTAMP_NOBEGIN || ts == TIMESTAMP_NOEND {
        return Ok(ts);
    }
    let mut result = ts;
    if iv.months != 0 || iv.days != 0 {
        result = if with_zone {
            let local = add_months_days(ts + zone_offset(ts), iv.months, iv.days)?;
            local - local_offset(local)
        } else {
            add_months_days(ts, iv.months, iv.days)?
        };
    }
    let result = result.checked_add(iv.micros).ok_or_else(timestamp_out_of_range)?;
    if !(-211_813_488_000_000_000..9_223_371_331_200_000_000).contains(&result) {
        return Err(timestamp_out_of_range());
    }
    Ok(result)
}

/// negate_interval negates an interval, failing on overflow.
pub fn negate_interval(iv: Interval) -> Result<Interval> {
    let overflow = || PgError::new(code::DATETIME_FIELD_OVERFLOW, "interval out of range");
    Ok(Interval {
        months: iv.months.checked_neg().ok_or_else(overflow)?,
        days: iv.days.checked_neg().ok_or_else(overflow)?,
        micros: iv.micros.checked_neg().ok_or_else(overflow)?,
    })
}

/// ts_round rounds to microsecond precision as Postgres' TSROUND does.
fn ts_round(value: f64) -> f64 {
    (value * 1_000_000.0).round_ties_even() / 1_000_000.0
}

/// interval_multiply scales an interval, cascading fractional months into days and fractional days into time, as
/// Postgres' interval_mul and interval_div do.
pub fn interval_multiply(iv: Interval, factor: f64, divide: bool) -> Result<Interval> {
    let scale = |x: f64| if divide { x / factor } else { x * factor };
    if divide && factor == 0.0 {
        return Err(PgError::new(code::DIVISION_BY_ZERO, "division by zero"));
    }
    let months_exact = scale(iv.months as f64);
    let days_exact = scale(iv.days as f64);
    if !months_exact.is_finite() || months_exact.abs() > i32::MAX as f64 || days_exact.abs() > i32::MAX as f64 {
        return Err(PgError::new(code::DATETIME_FIELD_OVERFLOW, "interval out of range"));
    }
    let months = months_exact as i32;
    let mut days = days_exact as i32;
    let month_remainder_days = ts_round((months_exact - months as f64) * 30.0);
    let mut seconds_remainder =
        ts_round((days_exact - days as f64 + month_remainder_days - (month_remainder_days as i32) as f64) * 86_400.0);
    if seconds_remainder.abs() >= 86_400.0 {
        days += (seconds_remainder / 86_400.0) as i32;
        seconds_remainder -= ((seconds_remainder / 86_400.0) as i32) as f64 * 86_400.0;
    }
    days += month_remainder_days as i32;
    let micros = (scale(iv.micros as f64) + seconds_remainder * USECS_PER_SEC as f64).round_ties_even();
    if !micros.is_finite() || micros.abs() >= 9.2e18 {
        return Err(PgError::new(code::DATETIME_FIELD_OVERFLOW, "interval out of range"));
    }
    Ok(Interval { months, days, micros: micros as i64 })
}
