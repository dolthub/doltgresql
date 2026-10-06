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

//! Reads the Go test definitions that the instrumented Go suite dumped, and converts them into runnable scripts.

use base64::Engine;
use base64::engine::general_purpose::STANDARD;
use harness::pgx::Time;
use harness::script::{A, BindVar, Expected, Flow, S, ScriptTest, ScriptTestAssertion};
use serde_json::Value;

/// Record is one call of a Go test runner.
pub struct Record {
    /// The line number of the record in the dump, from zero.
    pub index: usize,
    /// The Go test's name.
    pub test: String,
    /// The runner that was called, such as RunScripts.
    pub runner: String,
    /// The file and line of each caller, innermost first.
    pub callers: Vec<String>,
    /// Runner-specific arguments, such as the repetitions of RunScriptN.
    pub extra: Value,
    /// The test definitions passed to the runner.
    pub tests: Vec<Value>,
}

/// read_dump reads every record of a dump.
pub fn read_dump(path: &str) -> Vec<Record> {
    let text = std::fs::read_to_string(path).unwrap_or_else(|err| panic!("cannot read {path}: {err}"));
    text.lines()
        .enumerate()
        .map(|(index, line)| {
            let value: Value = serde_json::from_str(line).unwrap();
            Record {
                index,
                test: value["test"].as_str().unwrap().to_string(),
                runner: value["runner"].as_str().unwrap().to_string(),
                callers: value["callers"].as_array().unwrap().iter().map(|c| c.as_str().unwrap().to_string()).collect(),
                extra: value["extra"].clone(),
                tests: value["tests"].as_array().cloned().unwrap_or_default(),
            }
        })
        .collect()
}

/// leak turns a string into a static one, which is fine for a short-lived tool.
fn leak(text: &str) -> &'static str {
    Box::leak(text.to_string().into_boxed_str())
}

/// leak_slice turns a vector into a static slice.
fn leak_slice<T>(items: Vec<T>) -> &'static [T] {
    Box::leak(items.into_boxed_slice())
}

/// string_field returns a string field, or empty when it is absent.
fn string_field(value: &Value, name: &str) -> String {
    value.get(name).and_then(Value::as_str).unwrap_or_default().to_string()
}

/// bool_field returns a boolean field, or false when it is absent.
fn bool_field(value: &Value, name: &str) -> bool {
    value.get(name).and_then(Value::as_bool).unwrap_or(false)
}

/// strings_field returns a list of strings, or empty when it is absent.
fn strings_field(value: &Value, name: &str) -> Vec<String> {
    value
        .get(name)
        .and_then(Value::as_array)
        .map(|items| items.iter().map(|item| item.as_str().unwrap().to_string()).collect())
        .unwrap_or_default()
}

/// go_flow returns how the Go suite sends an assertion, which depends on what the assertion checks.
pub fn go_flow(assertion: &Value) -> Flow {
    if bool_field(assertion, "SkipResultsCheck")
        || !string_field(assertion, "ExpectedErr").is_empty()
        || !string_field(assertion, "ExpectedErrCode").is_empty()
        || !string_field(assertion, "ExpectedTag").is_empty()
    {
        Flow::Exec
    } else {
        Flow::Query
    }
}

/// transaction_client returns the client named by a "/* client x */" comment, as the Go suite parses it.
pub fn transaction_client(query: &str) -> Result<String, String> {
    let start = query.find("/*").ok_or_else(|| format!("no client comment found in query {query}"))?;
    let end = query.find("*/").ok_or_else(|| format!("no client comment found in query {query}"))?;
    if end < start {
        return Err(format!("no client comment found in query {query}"));
    }
    let comment = query[start + 2..end].trim();
    if !comment.to_lowercase().starts_with("client ") {
        return Err(format!("no client comment found in query {query}"));
    }
    Ok(comment["client ".len()..].trim().to_string())
}

/// to_assertion converts a Go assertion into a runnable one that is sent the same way. Its expectation is a
/// placeholder, since captures record outcomes instead of checking them, and it is never skipped, since a capture
/// runs everything.
pub fn to_assertion(assertion: &Value, transaction: bool) -> Result<ScriptTestAssertion, String> {
    let query = string_field(assertion, "Query");
    let bind_vars = match assertion.get("BindVars").and_then(Value::as_array) {
        Some(values) => values.iter().map(to_bind_var).collect::<Result<Vec<_>, _>>()?,
        None => Vec::new(),
    };
    let client = match assertion.get("Client").and_then(Value::as_str) {
        Some(client) => client.to_string(),
        None if transaction => transaction_client(&query)?,
        None => String::new(),
    };
    Ok(ScriptTestAssertion {
        query: leak(&query),
        bind_vars: leak_slice(bind_vars),
        expected: Expected::Ok,
        flow: go_flow(assertion),
        username: leak(&string_field(assertion, "Username")),
        password: leak(&string_field(assertion, "Password")),
        client: leak(&client),
        expected_blocking: bool_field(assertion, "ExpectedBlocking"),
        close_client: bool_field(assertion, "CloseClient"),
        copy_from_stdin_file: leak(&string_field(assertion, "CopyFromStdInFile")),
        copy_to_stdout_file: leak(&string_field(assertion, "CopyToStdOutFile")),
        copy_round_trip_stdin_query: leak(&string_field(assertion, "CopyRoundTripStdInQuery")),
        prepare: leak(&string_field(assertion, "Prepare")),
        ..A
    })
}

/// to_script converts a Go script into a runnable one. Every assertion runs, including those the Go test skips.
pub fn to_script(test: &Value, transaction: bool) -> Result<ScriptTest, String> {
    if test.get("ServerConfig").is_some() {
        return Err("ServerConfig has no YAML override".to_string());
    }
    let server_config = leak(test.get("ServerConfigYAML").and_then(Value::as_str).unwrap_or_default());
    let assertions = match test.get("Assertions").and_then(Value::as_array) {
        Some(assertions) => assertions.iter().map(|a| to_assertion(a, transaction)).collect::<Result<Vec<_>, _>>()?,
        None => Vec::new(),
    };
    Ok(ScriptTest {
        name: leak(&string_field(test, "Name")),
        database: leak(&string_field(test, "Database")),
        set_up_script: leak_slice(strings_field(test, "SetUpScript").iter().map(|s| leak(s)).collect()),
        assertions: leak_slice(assertions),
        server_config,
        ..S
    })
}

/// to_bind_var converts a dumped Go value into the parameter value pgx would encode.
pub fn to_bind_var(value: &Value) -> Result<BindVar, String> {
    let kind = value["$type"].as_str().unwrap_or_default();
    let inner = &value["$value"];
    let unsupported = || format!("unsupported bind variable: {value}");
    Ok(match kind {
        "nil" => BindVar::Null,
        "string" => BindVar::Str(leak(inner.as_str().ok_or_else(unsupported)?)),
        "int" => BindVar::Int(inner.as_i64().ok_or_else(unsupported)?),
        "int64" => BindVar::Int64(inner.as_i64().ok_or_else(unsupported)?),
        "int32" => BindVar::Int32(inner.as_i64().ok_or_else(unsupported)? as i32),
        "bool" => BindVar::Bool(inner.as_bool().ok_or_else(unsupported)?),
        "float64" => BindVar::Float64(go_float(&inner["$float"]).ok_or_else(unsupported)?),
        "[]uint8" => BindVar::Bytes(leak_slice(
            STANDARD.decode(inner["$bytes"].as_str().ok_or_else(unsupported)?).map_err(|err| err.to_string())?,
        )),
        "[]int32" => BindVar::Int32Array(leak_slice(
            inner.as_array().ok_or_else(unsupported)?.iter().map(|v| v.as_i64().unwrap() as i32).collect(),
        )),
        "[]string" => BindVar::StrArray(leak_slice(
            inner.as_array().ok_or_else(unsupported)?.iter().map(|v| leak(v.as_str().unwrap())).collect(),
        )),
        "time.Time" => BindVar::Time(parse_time(inner["$json"].as_str().ok_or_else(unsupported)?)?),
        "pgtype.Date" => BindVar::Date(parse_time(inner["$json"].as_str().ok_or_else(unsupported)?)?),
        "pgtype.Timestamp" => BindVar::Timestamp(parse_time(inner["$json"].as_str().ok_or_else(unsupported)?)?),
        "pgtype.UUID" => {
            let text = inner["$json"].as_str().ok_or_else(unsupported)?.replace('-', "");
            let mut bytes = [0u8; 16];
            for (i, byte) in bytes.iter_mut().enumerate() {
                *byte = u8::from_str_radix(&text[i * 2..i * 2 + 2], 16).map_err(|err| err.to_string())?;
            }
            BindVar::Uuid(bytes)
        }
        "pgtype.Numeric" => BindVar::Numeric(leak(&numeric_text(inner["$go"].as_str().ok_or_else(unsupported)?)?)),
        _ => return Err(unsupported()),
    })
}

/// go_float reads a dumped float, which is a number or a string for NaN and the infinities.
fn go_float(value: &Value) -> Option<f64> {
    match value {
        Value::Number(number) => number.as_f64(),
        Value::String(text) => match text.as_str() {
            "NaN" => Some(f64::NAN),
            "+Inf" => Some(f64::INFINITY),
            "-Inf" => Some(f64::NEG_INFINITY),
            _ => None,
        },
        _ => None,
    }
}

/// numeric_text rebuilds the text that a pgtype.Numeric was scanned from, out of its Go rendering, so that pgx's
/// parser recovers the same digits and exponent.
fn numeric_text(go: &str) -> Result<String, String> {
    let field = |name: &str| -> Option<&str> {
        let start = go.find(&format!("{name}:"))? + name.len() + 1;
        let rest = &go[start..];
        let end = rest.find([',', '}']).unwrap_or(rest.len());
        Some(rest[..end].trim())
    };
    if field("NaN") == Some("true") {
        return Ok("NaN".to_string());
    }
    match field("InfinityModifier") {
        Some("1") => return Ok("Infinity".to_string()),
        Some("-1") => return Ok("-Infinity".to_string()),
        _ => {}
    }
    let int = field("Int").ok_or_else(|| format!("cannot read numeric {go}"))?;
    let exp: i32 =
        field("Exp").ok_or_else(|| format!("cannot read numeric {go}"))?.parse().map_err(|_| go.to_string())?;
    let (negative, digits) = match int.strip_prefix('-') {
        Some(digits) => (true, digits.to_string()),
        None => (false, int.to_string()),
    };
    let mut text = if exp >= 0 {
        format!("{digits}{}", "0".repeat(exp as usize))
    } else {
        let scale = (-exp) as usize;
        let padded =
            if digits.len() <= scale { format!("{}{digits}", "0".repeat(scale - digits.len() + 1)) } else { digits };
        format!("{}.{}", &padded[..padded.len() - scale], &padded[padded.len() - scale..])
    };
    if negative {
        text.insert(0, '-');
    }
    Ok(text)
}

/// parse_time parses an ISO 8601 date or date-time with an optional offset into a time.
pub fn parse_time(text: &str) -> Result<Time, String> {
    let invalid = || format!("invalid time: {text}");
    let number = |s: &str| -> Result<u32, String> { s.parse().map_err(|_| invalid()) };
    let (date, rest) = match text.split_once('T') {
        Some((date, rest)) => (date, rest),
        None => (text, ""),
    };
    let mut date_parts = date.split('-');
    let year: i32 = date_parts.next().ok_or_else(invalid)?.parse().map_err(|_| invalid())?;
    let month = number(date_parts.next().ok_or_else(invalid)?)?;
    let day = number(date_parts.next().ok_or_else(invalid)?)?;
    let mut time = Time::utc(year, month, day, 0, 0, 0, 0);
    if rest.is_empty() {
        return Ok(time);
    }
    let (clock, offset) = if let Some(clock) = rest.strip_suffix('Z') {
        (clock, 0)
    } else if let Some(index) = rest.rfind(['+', '-']) {
        let (clock, offset) = rest.split_at(index);
        let sign = if offset.starts_with('-') { -1 } else { 1 };
        let (hours, minutes) = offset[1..].split_once(':').ok_or_else(invalid)?;
        (clock, sign * (number(hours)? as i32 * 3600 + number(minutes)? as i32 * 60))
    } else {
        (rest, 0)
    };
    let (whole, fraction) = clock.split_once('.').unwrap_or((clock, ""));
    let mut clock_parts = whole.split(':');
    time.hour = number(clock_parts.next().ok_or_else(invalid)?)?;
    time.minute = number(clock_parts.next().ok_or_else(invalid)?)?;
    time.second = number(clock_parts.next().ok_or_else(invalid)?)?;
    if !fraction.is_empty() {
        time.nanosecond = number(&format!("{fraction:0<9}")[..9])?;
    }
    time.offset_seconds = offset;
    Ok(time)
}
