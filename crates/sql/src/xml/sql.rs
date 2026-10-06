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

//! The SQL/XML expressions: XMLELEMENT, XMLFOREST, XMLCONCAT, XMLPARSE, XMLPI, XMLROOT, XMLSERIALIZE, and IS
//! DOCUMENT.

use crate::error::{PgError, Result, code};
use crate::types::Value;

/// XmlOp is an SQL/XML expression, whose arguments the expression holds.
#[derive(Clone, Debug, PartialEq)]
pub enum XmlOp {
    /// An element of a name, with attributes of names, whose values come first among the arguments and its content
    /// after them.
    Element {
        name: String,
        attributes: Vec<String>,
    },
    /// Elements of names, one for each argument that is not NULL.
    Forest(Vec<String>),
    Concat,
    /// Parsing text as a document, or as content.
    Parse {
        document: bool,
    },
    /// A processing instruction of a target, with an optional text argument.
    Pi(String),
    /// Setting an xml value's version, its first argument, and its standalone value, where None leaves the value's
    /// own.
    Root(Option<Option<bool>>),
    /// The text of an xml value, which must be a document when the flag is set.
    Serialize {
        document: bool,
    },
    IsDocument,
}

/// escape_name writes an SQL identifier as an XML name, as Postgres' map_sql_identifier_to_xml_name does.
pub fn escape_name(identifier: &str) -> String {
    let mut out = String::new();
    let chars: Vec<char> = identifier.chars().collect();
    for (i, &c) in chars.iter().enumerate() {
        if c == ':' && i == 0 {
            out.push_str("_x003A_");
        } else if c == '_' && chars.get(i + 1) == Some(&'x') {
            out.push_str("_x005F_");
        } else {
            let valid = if i == 0 {
                c.is_alphabetic() || c == '_' || c == ':'
            } else {
                c.is_alphanumeric() || matches!(c, '_' | ':' | '-' | '.' | '\u{b7}')
            };
            if valid {
                out.push(c);
            } else {
                out.push_str(&format!("_x{:04X}_", c as u32));
            }
        }
    }
    out
}

/// base64 encodes bytes in base64, as xmlbinary base64 writes them.
fn base64(bytes: &[u8]) -> String {
    const ALPHABET: &[u8; 64] = b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    let mut out = String::new();
    for chunk in bytes.chunks(3) {
        let n = chunk.iter().enumerate().fold(0u32, |n, (i, &b)| n | (b as u32) << (16 - 8 * i));
        for i in 0..4 {
            if i <= chunk.len() {
                out.push(ALPHABET[(n >> (18 - 6 * i) & 63) as usize] as char);
            } else {
                out.push('=');
            }
        }
    }
    out
}

/// value_text writes a value as XML text, as Postgres' map_sql_value_to_xml_value does, escaped for an attribute
/// or for content.
pub fn value_text(value: &Value, attribute: bool) -> String {
    use crate::datetime as dt;
    let escaped = |text: &str| if attribute { super::parse::escape_attribute(text) } else { super::escape(text) };
    match value {
        Value::Xml(text) => {
            let text = super::output(text);
            if attribute { super::parse::escape_attribute(&text) } else { text }
        }
        Value::Array(array) => {
            array.values.iter().map(|v| format!("<element>{}</element>", value_text(v, attribute))).collect()
        }
        Value::Bool(b) => if *b { "true" } else { "false" }.to_string(),
        Value::Bytea(bytes) => {
            if super::hex_binary() {
                bytes.iter().map(|b| format!("{b:02X}")).collect()
            } else {
                base64(bytes)
            }
        }
        Value::Date(days) => {
            let f = dt::fields_of_timestamp(*days as i64 * dt::USECS_PER_DAY);
            format!("{:04}-{:02}-{:02}", f.year, f.month, f.day)
        }
        Value::Timestamp(ts) | Value::TimestampTz(ts) => {
            let (local, offset) = match value {
                Value::TimestampTz(_) => {
                    let offset = dt::with_format(|f| f.zone.offset_at(*ts).0);
                    (ts + offset as i64 * dt::USECS_PER_SEC, Some(offset))
                }
                _ => (*ts, None),
            };
            let f = dt::fields_of_timestamp(local);
            let mut text =
                format!("{:04}-{:02}-{:02}T{:02}:{:02}:{:02}", f.year, f.month, f.day, f.hour, f.minute, f.second);
            if f.micros != 0 {
                text.push_str(format!(".{:06}", f.micros).trim_end_matches('0'));
            }
            if let Some(offset) = offset {
                let sign = if offset < 0 { '-' } else { '+' };
                text.push_str(&format!("{sign}{:02}:{:02}", offset.abs() / 3600, offset.abs() / 60 % 60));
            }
            text
        }
        other => escaped(&other.output().unwrap_or_default()),
    }
}

/// eval evaluates an SQL/XML expression over its argument values.
pub fn eval(op: &XmlOp, args: Vec<Value>) -> Result<Value> {
    let text = |value: &Value| match value {
        Value::Text(s) | Value::Xml(s) => s.clone(),
        other => other.output().unwrap_or_default(),
    };
    Ok(match op {
        XmlOp::Element { name, attributes } => {
            let mut out = format!("<{name}");
            for (attribute, value) in attributes.iter().zip(&args) {
                if !value.is_null() {
                    out.push_str(&format!(" {attribute}=\"{}\"", value_text(value, true)));
                }
            }
            let content: String =
                args[attributes.len()..].iter().filter(|v| !v.is_null()).map(|v| value_text(v, false)).collect();
            if content.is_empty() {
                out.push_str("/>");
            } else {
                out.push_str(&format!(">{content}</{name}>"));
            }
            Value::Xml(out)
        }
        XmlOp::Forest(names) => {
            let elements: Vec<String> = names
                .iter()
                .zip(&args)
                .filter(|(_, v)| !v.is_null())
                .map(|(name, v)| format!("<{name}>{}</{name}>", value_text(v, false)))
                .collect();
            if elements.is_empty() { Value::Null } else { Value::Xml(elements.concat()) }
        }
        XmlOp::Concat => {
            let values: Vec<String> = args.iter().filter(|v| !v.is_null()).map(text).collect();
            if values.is_empty() {
                Value::Null
            } else {
                Value::Xml(super::concat(&values.iter().map(String::as_str).collect::<Vec<_>>()))
            }
        }
        XmlOp::Parse { document } => {
            if args[0].is_null() {
                return Ok(Value::Null);
            }
            let input = text(&args[0]);
            let warnings = super::check(&input, *document)?;
            super::warn(warnings);
            Value::Xml(input)
        }
        XmlOp::Pi(target) => match args.first() {
            Some(Value::Null) => Value::Null,
            Some(value) => {
                let data = text(value);
                if data.contains("?>") {
                    return Err(PgError {
                        detail: Some("XML processing instruction cannot contain \"?>\".".into()),
                        ..PgError::new(code::INVALID_XML_PROCESSING_INSTRUCTION, "invalid XML processing instruction")
                    });
                }
                Value::Xml(format!("<?{target} {}?>", data.trim_start_matches(' ')))
            }
            None => Value::Xml(format!("<?{target}?>")),
        },
        XmlOp::Root(standalone) => {
            if args[0].is_null() {
                return Ok(Value::Null);
            }
            let version = (!args[1].is_null()).then(|| text(&args[1]));
            Value::Xml(super::root(&text(&args[0]), version.as_deref(), *standalone))
        }
        XmlOp::Serialize { document } => {
            if args[0].is_null() {
                return Ok(Value::Null);
            }
            let input = text(&args[0]);
            if *document && !super::is_document(&input) {
                return Err(PgError::new(code::NOT_AN_XML_DOCUMENT, "not an XML document"));
            }
            Value::Text(input)
        }
        XmlOp::IsDocument => match &args[0] {
            Value::Null => Value::Null,
            value => Value::Bool(super::is_document(&text(value))),
        },
    })
}

/// wrong_type returns Postgres' error for an argument of an SQL/XML expression that is not of the type it requires.
pub fn wrong_type(construct: &str, wanted: &str, actual: &str, position: Option<u32>) -> PgError {
    PgError {
        position,
        ..PgError::new(
            code::DATATYPE_MISMATCH,
            format!("argument of {construct} must be type {wanted}, not type {actual}"),
        )
    }
}
