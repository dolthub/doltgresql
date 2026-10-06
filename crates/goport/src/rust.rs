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

//! Renders Rust source for test definitions.

use std::fmt::Write as _;

use harness::script::{BindVar, Flow};
use serde_json::Value;

/// LICENSE_HEADER starts every generated file.
pub const LICENSE_HEADER: &str = "// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the \"License\");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an \"AS IS\" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
";

/// TESTDATA_PATHS holds the absolute testdata paths that captures saw, which generated strings replace with the
/// harness's testdata token.
pub static TESTDATA_PATHS: std::sync::OnceLock<Vec<String>> = std::sync::OnceLock::new();

/// string renders a string literal, using a raw string when escapes would hurt readability.
pub fn string(text: &str) -> String {
    let mut replaced = text.to_string();
    for path in TESTDATA_PATHS.get().map(Vec::as_slice).unwrap_or_default() {
        replaced = replaced.replace(path.as_str(), harness::script::TESTDATA_TOKEN);
    }
    literal(&replaced)
}

/// literal renders a string literal without any replacement.
fn literal(text: &str) -> String {
    let needs_escape = text.contains(['"', '\\']) || text.contains('\n');
    let printable = text.chars().all(|c| c == '\n' || c == '\t' || !c.is_control());
    if !needs_escape || !printable {
        return format!("{text:?}");
    }
    let mut hashes = 1;
    while text.contains(&format!("\"{}", "#".repeat(hashes))) {
        hashes += 1;
    }
    let fence = "#".repeat(hashes);
    format!("r{fence}\"{text}\"{fence}")
}

/// bytes renders a byte string literal when the bytes are printable ASCII, and a byte array otherwise.
pub fn bytes(data: &[u8]) -> String {
    if data.iter().all(|b| b.is_ascii() && (!b.is_ascii_control() || matches!(b, b'\n' | b'\t' | b'\r'))) {
        let text = std::str::from_utf8(data).unwrap();
        format!("b{text:?}")
    } else {
        format!("&{data:?}")
    }
}

/// snake_case converts a Go test name such as TestAlterTable into test_alter_table.
pub fn snake_case(name: &str) -> String {
    let mut out = String::new();
    let chars: Vec<char> = name.chars().collect();
    for (i, c) in chars.iter().enumerate() {
        if c.is_ascii_uppercase() {
            let previous_lower = i > 0 && (chars[i - 1].is_ascii_lowercase() || chars[i - 1].is_ascii_digit());
            let plural_acronym = i + 1 < chars.len()
                && chars[i + 1] == 's'
                && (i + 2 == chars.len() || chars[i + 2].is_ascii_uppercase());
            let next_lower = i + 1 < chars.len() && chars[i + 1].is_ascii_lowercase() && !plural_acronym;
            let previous_upper = i > 0 && chars[i - 1].is_ascii_uppercase();
            if i > 0 && (previous_lower || (previous_upper && next_lower)) {
                out.push('_');
            }
            out.push(c.to_ascii_lowercase());
        } else if c.is_ascii_alphanumeric() {
            out.push(*c);
        } else {
            out.push('_');
        }
    }
    out
}

/// type_oid renders a type OID as its named constant, or USER_DEFINED for types outside the built-in range.
pub fn type_oid(oid: u32) -> String {
    if oid >= 16384 {
        return "USER_DEFINED".to_string();
    }
    harness::oid::name(oid).map(str::to_string).unwrap_or_else(|| oid.to_string())
}

/// flow renders a flow.
pub fn flow(flow: Flow) -> &'static str {
    match flow {
        Flow::Auto => "Flow::Auto",
        Flow::Exec => "Flow::Exec",
        Flow::Query => "Flow::Query",
        Flow::Simple => "Flow::Simple",
    }
}

/// bind_var renders a parameter value.
pub fn bind_var(value: &BindVar) -> String {
    match value {
        BindVar::Null => "BindVar::Null".to_string(),
        BindVar::Int(v) => format!("BindVar::Int({v})"),
        BindVar::Int32(v) => format!("BindVar::Int32({v})"),
        BindVar::Int64(v) => format!("BindVar::Int64({v})"),
        BindVar::Float64(v) => {
            if v.is_nan() {
                "BindVar::Float64(f64::NAN)".to_string()
            } else if v.is_infinite() {
                if *v > 0.0 { "BindVar::Float64(f64::INFINITY)" } else { "BindVar::Float64(f64::NEG_INFINITY)" }
                    .to_string()
            } else {
                format!("BindVar::Float64({v:?})")
            }
        }
        BindVar::Bool(v) => format!("BindVar::Bool({v})"),
        BindVar::Str(v) => format!("BindVar::Str({})", string(v)),
        BindVar::Bytes(v) => format!("BindVar::Bytes(&{v:?})"),
        BindVar::Time(t) => format!("BindVar::Time({})", time(t)),
        BindVar::Date(t) => format!("BindVar::Date({})", time(t)),
        BindVar::Timestamp(t) => format!("BindVar::Timestamp({})", time(t)),
        BindVar::Numeric(v) => format!("BindVar::Numeric({})", string(v)),
        BindVar::Uuid(v) => format!("BindVar::Uuid({v:?})"),
        BindVar::StrArray(v) => {
            format!("BindVar::StrArray(&[{}])", v.iter().map(|s| string(s)).collect::<Vec<_>>().join(", "))
        }
        BindVar::Int32Array(v) => format!("BindVar::Int32Array(&{v:?})"),
    }
}

/// time renders a time value.
fn time(t: &harness::pgx::Time) -> String {
    format!(
        "Time {{ year: {}, month: {}, day: {}, hour: {}, minute: {}, second: {}, nanosecond: {}, offset_seconds: {} }}",
        t.year, t.month, t.day, t.hour, t.minute, t.second, t.nanosecond, t.offset_seconds
    )
}

/// diagnostic renders an error or notice from captured fields, starting from E for errors and N for notices.
pub fn diagnostic(fields: &Value) -> String {
    let severity = fields["severity"].as_str().unwrap_or_default();
    let base = match severity {
        "ERROR" => "E",
        "NOTICE" => "N",
        _ => "",
    };
    let mut parts = Vec::new();
    if base.is_empty() {
        parts.push(format!("severity: {}", string(severity)));
    }
    let mut put = |field: &str, key: &str| {
        if let Some(text) = fields.get(key).and_then(Value::as_str)
            && !text.is_empty()
        {
            parts.push(format!("{field}: {}", string(text)));
        }
    };
    put("code", "code");
    put("message", "message");
    put("detail", "detail");
    put("hint", "hint");
    if let Some(position) = fields.get("position").and_then(Value::as_i64) {
        parts.push(format!("position: {position}"));
    }
    let mut put = |field: &str, key: &str| {
        if let Some(text) = fields.get(key).and_then(Value::as_str)
            && !text.is_empty()
        {
            parts.push(format!("{field}: {}", string(text)));
        }
    };
    put("schema", "schema");
    put("table", "table");
    put("column", "column");
    put("data_type", "data_type");
    put("constraint", "constraint");
    let mut out = String::from("Diagnostic { ");
    for part in &parts {
        let _ = write!(out, "{part}, ");
    }
    if base.is_empty() {
        out.push_str("..E }");
    } else {
        let _ = write!(out, "..{base} }}");
    }
    out
}

/// cell renders a captured value.
pub fn cell(value: &Value) -> String {
    match value.as_str() {
        Some(text) => format!("T({})", string(text)),
        None => "Null".to_string(),
    }
}
