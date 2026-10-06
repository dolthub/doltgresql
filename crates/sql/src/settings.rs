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

//! Configuration parameters: Postgres 15's settings and Dolt's session variables, as a session sees and sets them.

use std::collections::HashMap;
use std::sync::OnceLock;

use crate::error::{PgError, Result, code};

/// SETTINGS is Postgres 15's pg_settings as a fresh cluster reports it, one tab-separated row per setting: name,
/// setting, unit, category, short description, extra description, context, type, source, minimum, maximum, enum
/// values, boot value, and reset value.
const SETTINGS: &str = include_str!("settings.tsv");

/// DOLT_VARIABLES are Dolt's session variables: name, type, and default.
const DOLT_VARIABLES: &[(&str, &str, &str)] = &[
    ("dolt_allow_commit_conflicts", "bool", "off"),
    ("dolt_force_transaction_commit", "bool", "off"),
    ("dolt_transaction_commit", "bool", "off"),
    ("dolt_transaction_commit_message", "string", ""),
    ("dolt_transactions_disabled", "bool", "off"),
    ("dolt_show_branch_databases", "bool", "off"),
    ("dolt_show_system_tables", "bool", "off"),
    ("dolt_dont_merge_json", "bool", "off"),
    ("dolt_optimize_json", "bool", "on"),
    ("dolt_override_schema", "string", ""),
    ("dolt_author_name", "string", ""),
    ("dolt_author_email", "string", ""),
    ("dolt_author_date", "string", ""),
    ("dolt_committer_name", "string", ""),
    ("dolt_committer_email", "string", ""),
    ("dolt_committer_date", "string", ""),
    ("dolt_allow_ci_creation", "bool", "off"),
    ("dolt_commit_verification_groups", "string", ""),
    ("dolt_stats_enabled", "bool", "on"),
    ("dolt_stats_paused", "bool", "on"),
    ("dolt_stats_memory_only", "bool", "off"),
    ("dolt_stats_job_interval", "integer", "30"),
    ("dolt_stats_gc_interval", "integer", "3600000"),
    ("dolt_stats_gc_enabled", "bool", "on"),
    ("dolt_stats_branches", "string", ""),
];

/// Setting is a configuration parameter's definition.
#[derive(Clone, Debug)]
pub struct Setting {
    pub name: String,
    /// The value in the base unit, as pg_settings shows it.
    pub default: String,
    pub unit: String,
    pub category: String,
    pub description: String,
    pub extra_description: String,
    pub context: String,
    pub kind: String,
    pub min: String,
    pub max: String,
    pub enum_values: Vec<String>,
}

/// Definitions indexes the settings by lowercase name.
struct Definitions {
    settings: Vec<Setting>,
    by_name: HashMap<String, usize>,
}

/// definitions returns every setting, reading the table on first use.
fn definitions() -> &'static Definitions {
    static DEFINITIONS: OnceLock<Definitions> = OnceLock::new();
    DEFINITIONS.get_or_init(|| {
        let mut settings: Vec<Setting> = SETTINGS
            .lines()
            .filter_map(|line| {
                let f: Vec<&str> = line.split('\t').collect();
                (f.len() >= 14).then(|| Setting {
                    name: f[0].to_string(),
                    default: f[1].to_string(),
                    unit: f[2].to_string(),
                    category: f[3].to_string(),
                    description: f[4].to_string(),
                    extra_description: f[5].to_string(),
                    context: f[6].to_string(),
                    kind: f[7].to_string(),
                    min: f[9].to_string(),
                    max: f[10].to_string(),
                    enum_values: if f[11].is_empty() {
                        Vec::new()
                    } else {
                        f[11].split(',').map(str::to_string).collect()
                    },
                })
            })
            .collect();
        for &(name, kind, default) in DOLT_VARIABLES {
            settings.push(Setting {
                name: name.to_string(),
                default: default.to_string(),
                unit: String::new(),
                category: "Dolt".into(),
                description: String::new(),
                extra_description: String::new(),
                context: "user".into(),
                kind: kind.to_string(),
                min: String::new(),
                max: String::new(),
                enum_values: Vec::new(),
            });
        }
        let by_name = settings.iter().enumerate().map(|(i, s)| (s.name.to_ascii_lowercase(), i)).collect();
        Definitions { settings, by_name }
    })
}

/// setting returns a setting's definition by name, ignoring case.
pub fn setting(name: &str) -> Option<&'static Setting> {
    let d = definitions();
    d.by_name.get(&name.to_ascii_lowercase()).map(|&i| &d.settings[i])
}

/// all_settings returns every setting in name order.
pub fn all_settings() -> &'static [Setting] {
    &definitions().settings
}

/// unrecognized returns Postgres' error for an unknown parameter.
pub fn unrecognized(name: &str) -> PgError {
    PgError::new(code::UNDEFINED_OBJECT, format!("unrecognized configuration parameter \"{name}\""))
}

/// invalid_value returns Postgres' error for a value a parameter rejects.
fn invalid_value(name: &str, value: &str) -> PgError {
    PgError::new(code::INVALID_PARAMETER_VALUE, format!("invalid value for parameter \"{name}\": \"{value}\""))
}

/// MEMORY_UNITS and TIME_UNITS are the units integer settings accept, with their size in the smallest unit.
const MEMORY_UNITS: &[(&str, i64)] = &[("TB", 1 << 40), ("GB", 1 << 30), ("MB", 1 << 20), ("kB", 1 << 10), ("B", 1)];
const TIME_UNITS: &[(&str, i64)] =
    &[("d", 86_400_000_000), ("h", 3_600_000_000), ("min", 60_000_000), ("s", 1_000_000), ("ms", 1_000), ("us", 1)];

/// unit_size returns the size of a setting's base unit in bytes or microseconds, and the table of units of its kind.
fn unit_size(unit: &str) -> Option<(i64, &'static [(&'static str, i64)])> {
    match unit {
        "B" => Some((1, MEMORY_UNITS)),
        "kB" => Some((1 << 10, MEMORY_UNITS)),
        "8kB" => Some((8 << 10, MEMORY_UNITS)),
        "MB" => Some((1 << 20, MEMORY_UNITS)),
        "us" => Some((1, TIME_UNITS)),
        "ms" => Some((1_000, TIME_UNITS)),
        "s" => Some((1_000_000, TIME_UNITS)),
        "min" => Some((60_000_000, TIME_UNITS)),
        _ => None,
    }
}

/// display returns how SHOW prints a setting's stored value, in the largest unit that divides it evenly.
pub fn display(definition: &Setting, value: &str) -> String {
    if definition.name == "tcp_user_timeout" {
        return "0".to_string();
    }
    if definition.kind == "real" {
        return display_real(definition, value);
    }
    if definition.kind != "integer" || definition.name.starts_with("tcp_") {
        return value.to_string();
    }
    let (Ok(number), Some((base, units))) = (value.parse::<i64>(), unit_size(&definition.unit)) else {
        return value.to_string();
    };
    if number <= 0 {
        return number.to_string();
    }
    let amount = number * base;
    let smallest = if std::ptr::eq(units, MEMORY_UNITS) { base.min(1 << 10) } else { base };
    for &(name, size) in units {
        if size >= smallest && amount % size == 0 {
            return format!("{}{name}", amount / size);
        }
    }
    format!("{number}{}", definition.unit)
}

/// display_real returns how SHOW prints a real setting's stored value, in the first unit, largest first, that
/// holds it as a whole number, or else in the smallest, as Postgres' convert_real_from_base_unit chooses.
fn display_real(definition: &Setting, value: &str) -> String {
    let (Ok(number), Some((base, units))) = (value.parse::<f64>(), unit_size(&definition.unit)) else {
        return value.to_string();
    };
    if number == 0.0 {
        return value.to_string();
    }
    let mut shown = (number, definition.unit.as_str());
    for &(name, size) in units {
        shown = (number * base as f64 / size as f64, name);
        if shown.0 > 0.0 && ((shown.0.round() / shown.0) - 1.0).abs() <= 1e-8 {
            break;
        }
    }
    format!("{}{}", crate::types::Value::Float8(shown.0).output().unwrap_or_default(), shown.1)
}

/// parse_bool reads a boolean setting value as Postgres' parse_bool does.
fn parse_bool(value: &str) -> Option<bool> {
    let word = value.trim().to_ascii_lowercase();
    let matches = |full: &str, min: usize| word.len() >= min && full.starts_with(&word);
    if matches("true", 1) || matches("yes", 1) || matches("on", 2) || word == "1" {
        Some(true)
    } else if matches("false", 1) || matches("no", 1) || matches("off", 2) || word == "0" {
        Some(false)
    } else {
        None
    }
}

/// normalize checks a value for a setting and returns it as the setting stores it.
pub fn normalize(definition: &Setting, value: &str) -> Result<String> {
    let name = &definition.name;
    match definition.kind.as_str() {
        "bool" => match parse_bool(value) {
            Some(b) => Ok(if b { "on" } else { "off" }.to_string()),
            None => Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                format!("parameter \"{name}\" requires a Boolean value"),
            )),
        },
        "integer" => {
            let trimmed = value.trim();
            let digits_end = trimmed
                .char_indices()
                .find(|&(i, c)| !(c.is_ascii_digit() || (i == 0 && (c == '-' || c == '+'))))
                .map_or(trimmed.len(), |(i, _)| i);
            let number: f64 = trimmed[..digits_end].parse().map_err(|_| invalid_value(name, value))?;
            let unit = trimmed[digits_end..].trim();
            let number = if unit.is_empty() {
                number
            } else {
                let (base, units) = unit_size(&definition.unit).ok_or_else(|| invalid_value(name, value))?;
                let size = units.iter().find(|(u, _)| *u == unit).map(|(_, s)| *s).ok_or_else(|| PgError {
                    hint: Some(if units == MEMORY_UNITS {
                        "Valid units for this parameter are \"B\", \"kB\", \"MB\", \"GB\", and \"TB\".".into()
                    } else {
                        "Valid units for this parameter are \"us\", \"ms\", \"s\", \"min\", \"h\", and \"d\".".into()
                    }),
                    ..invalid_value(name, value)
                })?;
                (number * size as f64 / base as f64).round()
            };
            let (min, max) = (definition.min.parse::<f64>().ok(), definition.max.parse::<f64>().ok());
            if min.is_some_and(|m| number < m) || max.is_some_and(|m| number > m) {
                return Err(PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    format!(
                        "{trimmed} is outside the valid range for parameter \"{name}\" ({} .. {})",
                        definition.min, definition.max
                    ),
                ));
            }
            if number != 0.0 && matches!(name.as_str(), "effective_io_concurrency" | "maintenance_io_concurrency") {
                return Err(PgError {
                    detail: Some(format!("{name} must be set to 0 on platforms that lack posix_fadvise().")),
                    ..PgError::new(
                        code::INVALID_PARAMETER_VALUE,
                        format!("invalid value for parameter \"{name}\": {}", number as i64),
                    )
                });
            }
            Ok((number as i64).to_string())
        }
        "real" => {
            let number: f64 = value.trim().parse().map_err(|_| invalid_value(name, value))?;
            let (min, max) = (definition.min.parse::<f64>().ok(), definition.max.parse::<f64>().ok());
            if min.is_some_and(|m| number < m) || max.is_some_and(|m| number > m) {
                return Err(PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    format!(
                        "{} is outside the valid range for parameter \"{name}\" ({} .. {})",
                        value.trim(),
                        definition.min,
                        definition.max
                    ),
                ));
            }
            Ok(crate::types::Value::Float8(number).output().unwrap_or_default())
        }
        "enum" => {
            let lower = value.trim().to_ascii_lowercase();
            definition.enum_values.iter().find(|v| v.to_ascii_lowercase() == lower).cloned().ok_or_else(|| PgError {
                hint: Some(format!("Available values: {}.", definition.enum_values.join(", "))),
                ..invalid_value(name, value)
            })
        }
        _ => match definition.name.as_str() {
            "TimeZone" | "log_timezone" => timezone(value).ok_or_else(|| invalid_value(name, value)),
            "DateStyle" => date_style(value).ok_or_else(|| PgError {
                detail: Some(format!("Unrecognized key word: \"{}\".", value.split(',').next().unwrap_or("").trim())),
                ..invalid_value(name, value)
            }),
            "timezone_abbreviations" => match value {
                "Default" => Ok(value.to_string()),
                "" => Err(PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    "could not read time zone file \"\": Is a directory",
                )),
                "Australia" | "India" => Err(PgError::unsupported("this time zone abbreviation file")),
                _ if value.chars().all(|c| c.is_ascii_alphabetic()) => Err(PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    format!("could not open time zone file \"{value}\": No such file or directory"),
                )),
                _ => Err(invalid_value(name, value)),
            },
            "client_encoding" => crate::encodings::Encoding::lookup(value)
                .map(|e| e.name().to_string())
                .ok_or_else(|| invalid_value(name, value)),
            _ => Ok(value.to_string()),
        },
    }
}

/// timezone returns the canonical name of a time zone: an IANA name in its database casing, or a numeric offset
/// written as Postgres writes it, `<-07>+07`.
pub fn timezone(value: &str) -> Option<String> {
    let trimmed = value.trim();
    if trimmed.eq_ignore_ascii_case("local") {
        return Some(local_timezone());
    }
    if let Ok(hours) = trimmed.parse::<f64>() {
        if hours.abs() > 168.0 {
            return None;
        }
        return Some(offset_zone_name((hours * 3600.0).round() as i64));
    }
    if let Some(zone) = chrono_tz::TZ_VARIANTS.iter().find(|z| z.name().eq_ignore_ascii_case(trimmed)) {
        return Some(zone.name().to_string());
    }
    posix_zone(trimmed).then(|| trimmed.to_ascii_uppercase())
}

/// posix_zone reports whether a time zone is a POSIX zone specification: a name of three or more letters, or one in
/// angle brackets, then an offset, then optionally a daylight saving name and rules. An offset alone must have a
/// colon.
fn posix_zone(text: &str) -> bool {
    let bytes = text.as_bytes();
    let name_end = if bytes.first() == Some(&b'<') {
        match text.find('>') {
            Some(end) => end + 1,
            None => return false,
        }
    } else {
        bytes.iter().take_while(|b| b.is_ascii_alphabetic()).count()
    };
    if name_end != 0 && name_end < 3 {
        return false;
    }
    let mut i = name_end;
    if matches!(bytes.get(i), Some(b'+' | b'-')) {
        i += 1;
    }
    let digits = bytes[i..].iter().take_while(|b| b.is_ascii_digit()).count();
    if digits == 0 || digits > 3 {
        return false;
    }
    i += digits;
    let mut colons = 0;
    while bytes.get(i) == Some(&b':') {
        let digits = bytes[i + 1..].iter().take_while(|b| b.is_ascii_digit()).count();
        if digits == 0 || digits > 2 || colons == 2 {
            return false;
        }
        colons += 1;
        i += 1 + digits;
    }
    if name_end == 0 && colons == 0 {
        return false;
    }
    let rest = &text[i..];
    rest.is_empty() || rest.as_bytes().iter().take_while(|b| b.is_ascii_alphabetic()).count() >= 3
}

/// offset_zone_name names a fixed offset east of UTC as Postgres does for a numeric time zone, which in POSIX style
/// gives the offset west of UTC.
pub fn offset_zone_name(seconds_east: i64) -> String {
    let format = |seconds: i64| {
        let sign = if seconds < 0 { '-' } else { '+' };
        let abs = seconds.abs();
        let (h, m, s) = (abs / 3600, abs / 60 % 60, abs % 60);
        match (m, s) {
            (0, 0) => format!("{sign}{h:02}"),
            (_, 0) => format!("{sign}{h:02}:{m:02}"),
            _ => format!("{sign}{h:02}:{m:02}:{s:02}"),
        }
    };
    format!("<{}>{}", format(seconds_east), format(-seconds_east))
}

/// local_timezone returns the IANA name of the machine's time zone, from TZ or the /etc/localtime link, or UTC.
pub fn local_timezone() -> String {
    if let Ok(zone) = std::env::var("TZ")
        && !zone.is_empty()
    {
        return zone.trim_start_matches(':').to_string();
    }
    std::fs::read_link("/etc/localtime")
        .ok()
        .and_then(|path| path.to_str().and_then(|p| p.split_once("zoneinfo/").map(|(_, zone)| zone.to_string())))
        .unwrap_or_else(|| "UTC".to_string())
}

/// date_style returns the canonical DateStyle for a value: an output style and a field order.
fn date_style(value: &str) -> Option<String> {
    let mut style: Option<&str> = None;
    let mut order: Option<&str> = None;
    for word in value.split(',').map(|w| w.trim().to_ascii_lowercase()) {
        match word.as_str() {
            "iso" => style = Some("ISO"),
            "sql" => style = Some("SQL"),
            "postgres" => style = Some("Postgres"),
            "german" => {
                style = Some("German");
                order = order.or(Some("DMY"));
            }
            "ymd" => order = Some("YMD"),
            "dmy" | "euro" | "european" => order = Some("DMY"),
            "mdy" | "us" | "noneuro" | "noneuropean" => order = Some("MDY"),
            "default" => {}
            _ => return None,
        }
    }
    Some(format!("{}, {}", style.unwrap_or("ISO"), order.unwrap_or("MDY")))
}

/// Settings is a session's parameter values, with the changes its transaction can undo.
#[derive(Clone, Debug, Default)]
pub struct Settings {
    values: HashMap<String, String>,
    /// The values to restore when the transaction rolls back, oldest first.
    transaction_undo: Vec<(String, Option<String>)>,
    /// The values to restore when the transaction ends, for SET LOCAL.
    local_undo: Vec<(String, Option<String>)>,
}

impl Settings {
    /// new returns the settings a connection starts with, given the startup parameters it sent, failing as
    /// Postgres does on an unknown parameter or a value its parameter rejects.
    pub fn new(startup: &[(String, String)]) -> Result<Settings> {
        let mut settings = Settings::default();
        settings.values.insert("timezone".into(), local_timezone());
        for (name, value) in startup {
            let definition = setting(name).ok_or_else(|| unrecognized(name))?;
            let value = normalize(definition, value)?;
            settings.values.insert(definition.name.to_ascii_lowercase(), value);
        }
        Ok(settings)
    }

    /// get returns a parameter's value as stored, or None for an unknown one.
    pub fn get(&self, name: &str) -> Option<String> {
        let key = name.to_ascii_lowercase();
        if let Some(value) = self.values.get(&key) {
            return Some(value.clone());
        }
        setting(name).map(|s| s.default.clone())
    }

    /// show returns a parameter's value as SHOW prints it.
    pub fn show(&self, name: &str) -> Result<String> {
        match name.to_ascii_lowercase().as_str() {
            "role" => return Ok(self.raw("role").unwrap_or_else(|| "none".into())),
            "session_authorization" => return Ok(self.raw("session_authorization").unwrap_or_default()),
            _ => {}
        }
        let value = self.get(name).ok_or_else(|| unrecognized(name))?;
        Ok(match setting(name) {
            Some(definition) => display(definition, &value),
            None => value,
        })
    }

    /// set sets a parameter, or resets it to its default without a value, locally to the transaction when asked.
    /// Inside a transaction, the change is undone when the transaction rolls back.
    pub fn set(&mut self, name: &str, value: Option<&str>, local: bool, in_transaction: bool) -> Result<()> {
        let key = name.to_ascii_lowercase();
        let value = match setting(name) {
            Some(definition) => {
                match definition.context.as_str() {
                    "internal" => {
                        return Err(PgError::new(
                            code::CANT_CHANGE_RUNTIME_PARAM,
                            format!("parameter \"{name}\" cannot be changed"),
                        ));
                    }
                    "postmaster" => {
                        return Err(PgError::new(
                            code::CANT_CHANGE_RUNTIME_PARAM,
                            format!("parameter \"{name}\" cannot be changed without restarting the server"),
                        ));
                    }
                    "sighup" => {
                        return Err(PgError::new(
                            code::CANT_CHANGE_RUNTIME_PARAM,
                            format!("parameter \"{name}\" cannot be changed now"),
                        ));
                    }
                    "backend" | "superuser-backend" => {
                        return Err(PgError::new(
                            code::CANT_CHANGE_RUNTIME_PARAM,
                            format!("parameter \"{name}\" cannot be set after connection start"),
                        ));
                    }
                    _ => {}
                }
                value.map(|v| normalize(definition, v)).transpose()?
            }
            None if name.contains('.') => value.map(str::to_string),
            None => return Err(unrecognized(name)),
        };
        let value = match value {
            None if name.contains('.') && setting(name).is_none() => Some(String::new()),
            value => value,
        };
        self.store(key, value, local, in_transaction);
        Ok(())
    }

    /// set_raw sets one of the parameters that SET ROLE and SET SESSION AUTHORIZATION change, which have no
    /// definitions, or removes it without a value.
    pub fn set_raw(&mut self, name: &str, value: Option<String>, local: bool, in_transaction: bool) {
        self.store(name.to_string(), value, local, in_transaction);
    }

    /// raw returns the value of a parameter as stored, without a default.
    pub fn raw(&self, name: &str) -> Option<String> {
        self.values.get(name).cloned()
    }

    /// store stores a value, remembering the previous one for the transaction to restore.
    fn store(&mut self, key: String, value: Option<String>, local: bool, in_transaction: bool) {
        let previous = self.values.get(&key).cloned();
        if local {
            if in_transaction {
                self.local_undo.push((key.clone(), previous));
            }
        } else if in_transaction {
            self.transaction_undo.push((key.clone(), previous));
        }
        if local && !in_transaction {
            return;
        }
        match value {
            Some(value) => self.values.insert(key, value),
            None => self.values.remove(&key),
        };
    }

    /// reset_all resets every parameter the session can change.
    pub fn reset_all(&mut self, in_transaction: bool) {
        let keys: Vec<String> = self.values.keys().cloned().collect();
        for key in keys {
            if in_transaction {
                self.transaction_undo.push((key.clone(), self.values.get(&key).cloned()));
            }
            if !matches!(key.as_str(), "timezone" | "role" | "session_authorization") {
                self.values.remove(&key);
            }
        }
        self.values.insert("timezone".into(), local_timezone());
    }

    /// end_transaction ends the transaction, undoing SET LOCAL, and undoing every change when it rolled back.
    pub fn end_transaction(&mut self, committed: bool) {
        let mut undo = std::mem::take(&mut self.local_undo);
        if committed {
            self.transaction_undo.clear();
        } else {
            undo.splice(0..0, std::mem::take(&mut self.transaction_undo));
        }
        for (key, previous) in undo.into_iter().rev() {
            match previous {
                Some(value) => self.values.insert(key, value),
                None => self.values.remove(&key),
            };
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn values_normalize_as_postgres_does() {
        let s = |name: &str| setting(name).unwrap();
        assert_eq!(normalize(s("enable_hashjoin"), "OFF").unwrap(), "off");
        assert_eq!(normalize(s("work_mem"), "64MB").unwrap(), "65536");
        assert_eq!(display(s("work_mem"), "65536"), "64MB");
        assert_eq!(display(s("shared_buffers"), "16384"), "128MB");
        assert_eq!(display(s("statement_timeout"), "0"), "0");
        assert_eq!(normalize(s("DateStyle"), "german").unwrap(), "German, DMY");
        assert_eq!(normalize(s("TimeZone"), "america/new_york").unwrap(), "America/New_York");
        assert_eq!(normalize(s("TimeZone"), "-7").unwrap(), "<-07>+07");
        assert_eq!(normalize(s("TimeZone"), "5.5").unwrap(), "<+05:30>-05:30");
        for (zone, shown) in [("+00:00", "+00:00"), ("abc+5", "ABC+5"), ("EST5EDT", "EST5EDT"), ("UTC+3", "UTC+3")] {
            assert_eq!(normalize(s("TimeZone"), zone).unwrap(), shown);
        }
        assert!(normalize(s("TimeZone"), "foo").is_err());
        assert!(normalize(s("TimeZone"), "Z").is_err());
        let err = normalize(s("enable_hashjoin"), "maybe").unwrap_err();
        assert_eq!(err.message, "parameter \"enable_hashjoin\" requires a Boolean value");
        let err = normalize(s("xmlbinary"), "nope").unwrap_err();
        assert_eq!(err.hint.unwrap(), "Available values: base64, hex.");
    }

    #[test]
    fn transactions_undo_their_changes() {
        let mut settings = Settings::default();
        settings.set("enable_seqscan", Some("off"), false, false).unwrap();
        settings.set("enable_seqscan", Some("on"), false, true).unwrap();
        settings.set("work_mem", Some("8MB"), true, true).unwrap();
        settings.end_transaction(false);
        assert_eq!(settings.show("enable_seqscan").unwrap(), "off");
        assert_eq!(settings.show("work_mem").unwrap(), "4MB");
        settings.set("my.var", Some("x"), false, false).unwrap();
        assert_eq!(settings.show("my.var").unwrap(), "x");
        assert_eq!(settings.show("nope").unwrap_err().code, "42704");
    }
}
