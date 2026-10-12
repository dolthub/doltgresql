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

use std::path::{Path, PathBuf};
use std::sync::Mutex;
use std::time::{Duration, Instant, SystemTime};

use harness::gostd;
use md5::{Digest, Md5};

use crate::client::{Client, Error};
use crate::parser::{Record, RecordType, parse_test_file};

/// ENGINE is the engine name that test files skip or select with conditions.
pub const ENGINE: &str = "postgresql";

/// DEFAULT_TIMEOUT is how long a record may run when no timeout is given.
const DEFAULT_TIMEOUT: Duration = Duration::from_secs(20 * 60);

/// Outcome is a record's result as the log records it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Outcome {
    Ok,
    NotOk(String),
    Skipped,
    Timeout,
    DidNotRun,
}

/// Harness connects to a server and resets one database before each test file.
pub struct Harness {
    pub port: u16,
    pub database: String,
    pub timeout: Option<Duration>,
    client: Option<Client>,
}

impl Harness {
    /// new returns a harness for the database on the server.
    pub fn new(port: u16, database: &str, timeout: Option<Duration>) -> Harness {
        Harness { port, database: database.to_string(), timeout, client: None }
    }

    /// init drops and recreates the database and connects to it.
    pub fn init(&mut self) -> Result<(), Error> {
        self.client = None;
        let mut admin = Client::connect(self.port, "")?;
        admin.exec(&format!("DROP DATABASE IF EXISTS {}", self.database))?;
        admin.exec(&format!("CREATE DATABASE {}", self.database))?;
        let mut client = Client::connect(self.port, &self.database)?;
        client.ping()?;
        self.client = Some(client);
        Ok(())
    }

    /// client returns the connection, reconnecting when the last one broke the way database/sql does.
    fn client(&mut self) -> Result<&mut Client, Error> {
        if self.client.as_ref().is_none_or(|c| c.closed) {
            self.client = Some(Client::connect(self.port, &self.database)?);
        }
        Ok(self.client.as_mut().unwrap())
    }
}

/// Column is how database/sql scans a result column, chosen by the type name like the Go harness does.
#[derive(Clone, Copy)]
enum Column {
    Bool,
    Int,
    Float,
    Text,
}

/// column returns how a column scans and its schema letter.
fn column(oid: u32) -> Result<(Column, char), String> {
    Ok(match oid {
        16 | 1560 => (Column::Bool, 'I'),
        25 | 1043 | 18 | 19 | 17 => (Column::Text, 'T'),
        700 | 701 | 1700 => (Column::Float, 'R'),
        21 | 23 | 20 => (Column::Int, 'I'),
        705 => (Column::Text, 'I'),
        _ => {
            let name = harness::oid::name(oid).map_or_else(|| oid.to_string(), str::to_string);
            return Err(format!("Unhandled type {name}"));
        }
    })
}

/// go_parse_bool matches strconv.ParseBool.
fn go_parse_bool(s: &str) -> Option<bool> {
    match s {
        "1" | "t" | "T" | "TRUE" | "true" | "True" => Some(true),
        "0" | "f" | "F" | "FALSE" | "false" | "False" => Some(false),
        _ => None,
    }
}

/// decode_hex_bytea matches pgx v4's bytea text decoding.
fn decode_hex_bytea(src: &[u8]) -> Option<Vec<u8>> {
    let hex = src.strip_prefix(b"\\x")?;
    if hex.len() % 2 != 0 {
        return None;
    }
    let nibble = |c: u8| (c as char).to_digit(16).map(|d| d as u8);
    hex.chunks(2).map(|pair| Some(nibble(pair[0])? << 4 | nibble(pair[1])?)).collect()
}

/// value matches how the Go harness scans and prints a cell.
fn value(column: Column, oid: u32, format: i16, src: Option<&[u8]>) -> Result<String, String> {
    let Some(src) = src else {
        return Ok("NULL".to_string());
    };
    let text = || String::from_utf8_lossy(src).into_owned();
    let invalid = |what: &str| format!("convert field failed: invalid length for {what}: {}", src.len());
    let binary = format == 1;
    Ok(match column {
        Column::Bool => {
            let value = match oid {
                16 if binary => {
                    if src.len() != 1 {
                        return Err(invalid("bool"));
                    }
                    src[0] == 1
                }
                16 => {
                    if src.len() != 1 {
                        return Err(invalid("bool"));
                    }
                    src[0] == b't'
                }
                _ => go_parse_bool(&text())
                    .ok_or_else(|| format!("sql/driver: couldn't convert {:?} into type bool", text()))?,
            };
            if value { "1" } else { "0" }.to_string()
        }
        Column::Int => {
            let value = if binary {
                match (oid, src.len()) {
                    (21, 2) => i16::from_be_bytes(src.try_into().unwrap()) as i64,
                    (23, 4) => i32::from_be_bytes(src.try_into().unwrap()) as i64,
                    (20, 8) => i64::from_be_bytes(src.try_into().unwrap()),
                    _ => return Err(invalid("int")),
                }
            } else {
                let bits = match oid {
                    21 => 16,
                    23 => 32,
                    _ => 64,
                };
                gostd::parse_int(&text(), bits).ok_or_else(|| format!("convert field failed: {}", text()))?
            };
            value.to_string()
        }
        Column::Float => {
            let value = match oid {
                700 if binary && src.len() == 4 => f32::from_be_bytes(src.try_into().unwrap()) as f64,
                701 if binary && src.len() == 8 => f64::from_be_bytes(src.try_into().unwrap()),
                700 | 701 if binary => return Err(invalid("float")),
                700 => gostd::parse_float32(&text()).ok_or_else(|| format!("convert field failed: {}", text()))? as f64,
                _ => gostd::parse_float64(&text()).ok_or_else(|| {
                    format!("converting driver.Value type string ({:?}) to a float64: invalid syntax", text())
                })?,
            };
            gostd::format_f64(value, 3)
        }
        Column::Text if oid == 17 && !binary => {
            String::from_utf8_lossy(&decode_hex_bytea(src).ok_or("convert field failed: invalid hex format")?)
                .into_owned()
        }
        Column::Text => text(),
    })
}

/// execute_query runs a query and returns its schema and its values in order.
fn execute_query(client: &mut Client, sql: &str) -> Result<(String, Vec<String>), String> {
    let result = client.query(sql).map_err(|e| e.to_string())?;
    let columns = result.fields.iter().map(|f| column(f.data_type_oid)).collect::<Result<Vec<_>, _>>()?;
    let schema: String = columns.iter().map(|(_, letter)| *letter).collect();
    let mut values = Vec::new();
    for row in &result.rows {
        for ((field, (kind, _)), cell) in result.fields.iter().zip(&columns).zip(row) {
            values.push(value(*kind, field.data_type_oid, field.format, cell.as_deref())?);
        }
    }
    Ok((schema, values))
}

/// normalize_results matches the Go runner's normalizeResults.
fn normalize_results(results: &[String], schema: &str) -> Result<Vec<String>, String> {
    if schema.is_empty() && !results.is_empty() {
        return Err("runtime error: integer divide by zero".to_string());
    }
    let schema = schema.as_bytes();
    Ok(results
        .iter()
        .enumerate()
        .map(|(i, result)| {
            let kind = schema[i % schema.len()];
            if kind == b'R' && !result.contains('.') && result.parse::<i64>().is_ok() {
                return format!("{result}.000");
            }
            if kind == b'I'
                && result.contains('.')
                && let Some(f) = gostd::parse_float64(result)
                && f.trunc() == f
            {
                return format!("{}", f as i64);
            }
            result.clone()
        })
        .collect())
}

/// hash_results matches the original sqllogictest hash: the MD5 of every value followed by a newline.
pub fn hash_results(results: &[String]) -> String {
    let mut hasher = Md5::new();
    for result in results {
        hasher.update(result.as_bytes());
        hasher.update(b"\n");
    }
    hasher.finalize().iter().map(|b| format!("{b:02x}")).collect()
}

/// verify compares a query's results against the record, matching the Go runner's checks and messages.
fn verify(record: &Record, schema: &str, results: Vec<String>) -> Result<Outcome, String> {
    let empty = record.num_results() == 0 && results.is_empty();
    if !empty && schema != record.schema {
        let compatible = schema.len() == record.schema.len()
            && record
                .schema
                .bytes()
                .zip(schema.bytes())
                .all(|(e, a)| e == a || matches!((e, a), (b'R', b'I') | (b'I', b'R')));
        if !compatible {
            return Ok(Outcome::NotOk(format!("Schemas differ. Expected {}, got {schema}", record.schema)));
        }
    }
    if results.len() != record.num_results() {
        return Ok(Outcome::NotOk(format!(
            "Incorrect number of results. Expected {}, got {}",
            record.num_results(),
            results.len()
        )));
    }
    let results = record.sort_results(normalize_results(&results, &record.schema)?)?;
    if record.is_hash_result() {
        let results = record.sort_results(results)?;
        let computed = hash_results(&results);
        if record.hash() != computed {
            return Ok(Outcome::NotOk(format!("Hash of results differ. Expected {}, got {computed}", record.hash())));
        }
        return Ok(Outcome::Ok);
    }
    for (i, expected) in record.result.iter().enumerate() {
        if *expected != results[i] {
            return Ok(Outcome::NotOk(format!(
                "Incorrect result at position {i}. Expected {expected}, got {}",
                results[i]
            )));
        }
    }
    Ok(Outcome::Ok)
}

/// execute runs a record, returning its outcome, or None when nothing is logged.
fn execute(harness: &mut Harness, record: &Record) -> Option<Outcome> {
    if !record.should_execute_for_engine(ENGINE) {
        return (record.record_type != RecordType::Halt).then_some(Outcome::Skipped);
    }
    let deadline = Instant::now() + harness.timeout.unwrap_or(DEFAULT_TIMEOUT);
    let client = match harness.client() {
        Ok(client) => client,
        Err(e) => return Some(Outcome::NotOk(format!("Unexpected error {e}"))),
    };
    client.set_deadline(Some(deadline));
    let outcome = match record.record_type {
        RecordType::Statement => match (client.exec(&record.query), record.expect_error) {
            (Err(Error::Timeout), _) => Outcome::Timeout,
            (Ok(()), true) => Outcome::NotOk("Expected error but didn't get one".to_string()),
            (Err(e), false) => Outcome::NotOk(format!("Unexpected error {e}")),
            _ => Outcome::Ok,
        },
        RecordType::Query => match execute_query(client, &record.query) {
            Err(_) if client.closed && Instant::now() >= deadline => Outcome::Timeout,
            Err(e) => Outcome::NotOk(format!("Unexpected error {e}")),
            Ok((schema, results)) => verify(record, &schema, results)
                .unwrap_or_else(|panic| Outcome::NotOk(format!("Caught panic: {panic}"))),
        },
        RecordType::Halt => unreachable!(),
    };
    client.set_deadline(None);
    Some(outcome)
}

/// test_file_path returns the last four elements of a test file's path, stopping at the `test` directory.
pub fn test_file_path(file: &Path) -> String {
    let mut elements: Vec<String> = Vec::new();
    for component in file.components().rev() {
        let name = component.as_os_str().to_string_lossy().into_owned();
        if elements.len() >= 4 || name == "test" || name == "/" {
            break;
        }
        elements.insert(0, name);
    }
    elements.join("/")
}

/// rfc3339_nano formats a time like Go's time.RFC3339Nano in UTC.
fn rfc3339_nano(time: SystemTime) -> String {
    let since = time.duration_since(SystemTime::UNIX_EPOCH).unwrap_or_default();
    let days = (since.as_secs() / 86400) as i64;
    let seconds = since.as_secs() % 86400;
    let (year, month, day) = gostd::civil_from_days(days);
    let mut out =
        format!("{year:04}-{month:02}-{day:02}T{:02}:{:02}:{:02}", seconds / 3600, seconds / 60 % 60, seconds % 60);
    if since.subsec_nanos() != 0 {
        out.push('.');
        out.push_str(format!("{:09}", since.subsec_nanos()).trim_end_matches('0'));
    }
    out.push('Z');
    out
}

/// log_line formats a result the way the Go runner logs it.
pub fn log_line(file: &Path, record: &Record, started: Instant, outcome: &Outcome) -> String {
    let prefix = format!(
        "{} {} {}:{}: {}",
        rfc3339_nano(SystemTime::now()),
        started.elapsed().as_millis(),
        test_file_path(file),
        record.line_num,
        record.query
    );
    match outcome {
        Outcome::Ok => format!("{prefix} ok"),
        Outcome::NotOk(message) => format!("{prefix} not ok: {message}").replace('\n', " "),
        Outcome::Skipped => format!("{prefix} skipped"),
        Outcome::Timeout => format!("{prefix} timeout"),
        Outcome::DidNotRun => format!("{prefix} did not run"),
    }
}

/// run_test_file runs a test file, printing each record's result.
pub fn run_test_file(harness: &mut Harness, file: &Path, output: &Mutex<()>) -> Result<(), String> {
    harness.init().map_err(|e| format!("{}: {e}", file.display()))?;
    let records = parse_test_file(file)?;
    let mut timed_out = false;
    for record in &records {
        let started = Instant::now();
        let outcome = if timed_out {
            Some(Outcome::DidNotRun)
        } else if record.record_type == RecordType::Halt && record.should_execute_for_engine(ENGINE) {
            break;
        } else {
            execute(harness, record)
        };
        if let Some(outcome) = outcome {
            timed_out |= outcome == Outcome::Timeout;
            let line = log_line(file, record, started, &outcome);
            let _guard = output.lock().unwrap();
            println!("{line}");
        }
    }
    Ok(())
}

/// collect_test_files returns the `.test` files under the paths, in the order the Go runner walks them.
pub fn collect_test_files(paths: &[String]) -> Result<Vec<PathBuf>, String> {
    fn walk(dir: &Path, files: &mut Vec<PathBuf>) -> Result<(), String> {
        let mut entries: Vec<PathBuf> = std::fs::read_dir(dir)
            .map_err(|e| format!("{}: {e}", dir.display()))?
            .map(|e| e.map(|e| e.path()).map_err(|e| e.to_string()))
            .collect::<Result<_, _>>()?;
        entries.sort();
        for entry in entries {
            if entry.is_dir() {
                walk(&entry, files)?;
            } else if entry.to_string_lossy().ends_with(".test") {
                files.push(entry);
            }
        }
        Ok(())
    }
    let mut files = Vec::new();
    for path in paths {
        let path = Path::new(path);
        if path.is_dir() {
            walk(path, &mut files)?;
        } else {
            files.push(std::fs::canonicalize(path).map_err(|e| format!("{}: {e}", path.display()))?);
        }
    }
    Ok(files)
}
