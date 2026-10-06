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

//! Converts the Go suite's wire tests and message flow tests into wire conversations, and renders captured
//! backend messages as Rust.

use base64::Engine;
use base64::engine::general_purpose::STANDARD;
use harness::wire::{Datum, Send, Step, W, WireTest};
use pgproto::BackendMessage;
use serde_json::{Value, json};

use crate::rust;

/// leak turns a string into a static one.
fn leak(text: &str) -> &'static str {
    Box::leak(text.to_string().into_boxed_str())
}

/// leak_slice turns a vector into a static slice.
fn leak_slice<T>(items: Vec<T>) -> &'static [T] {
    Box::leak(items.into_boxed_slice())
}

/// text returns a string field of a dumped message.
fn text(value: &Value, name: &str) -> &'static str {
    leak(value.get(name).and_then(Value::as_str).unwrap_or_default())
}

/// datum converts a dumped pgproto3 value, which is null, {"text": ...}, or {"binary": hex}.
fn datum(value: &Value) -> Result<Datum, String> {
    if value.is_null() {
        return Ok(Datum::Null);
    }
    if let Some(text) = value.get("text").and_then(Value::as_str) {
        return Ok(Datum::Text(leak(text)));
    }
    if let Some(hex) = value.get("binary").and_then(Value::as_str) {
        let bytes = (0..hex.len())
            .step_by(2)
            .map(|i| u8::from_str_radix(&hex[i..i + 2], 16).map_err(|err| err.to_string()))
            .collect::<Result<Vec<u8>, _>>()?;
        return Ok(match String::from_utf8(bytes.clone()) {
            Ok(text) => Datum::Text(leak(&text)),
            Err(_) => Datum::Bytes(leak_slice(bytes)),
        });
    }
    Err(format!("unknown datum {value}"))
}

/// object_type reads a Describe or Close object type, which defaults to a statement.
fn object_type(value: &Value) -> u8 {
    match value.get("ObjectType") {
        Some(Value::String(s)) => s.as_bytes().first().copied().unwrap_or(b'S'),
        Some(Value::Number(n)) if n.as_u64() != Some(0) => n.as_u64().unwrap() as u8,
        _ => b'S',
    }
}

/// i16s reads a list of integers.
fn i16s(value: &Value, name: &str) -> &'static [i16] {
    leak_slice(
        value
            .get(name)
            .and_then(Value::as_array)
            .map(|a| a.iter().map(|v| v.as_i64().unwrap() as i16).collect())
            .unwrap_or_default(),
    )
}

/// send converts a dumped pgproto3 frontend message.
pub fn send(message: &Value) -> Result<Send, String> {
    let kind = message["$type"].as_str().unwrap_or_default();
    let value = &message["$value"]["$json"];
    Ok(match kind {
        "*pgproto3.Query" => Send::Query(text(value, "String")),
        "*pgproto3.Parse" => Send::Parse {
            name: text(value, "Name"),
            query: text(value, "Query"),
            parameter_oids: leak_slice(
                value["ParameterOIDs"]
                    .as_array()
                    .map(|a| a.iter().map(|v| v.as_u64().unwrap() as u32).collect())
                    .unwrap_or_default(),
            ),
        },
        "*pgproto3.Bind" => Send::Bind {
            portal: text(value, "DestinationPortal"),
            statement: text(value, "PreparedStatement"),
            parameter_formats: i16s(value, "ParameterFormatCodes"),
            parameters: leak_slice(
                value["Parameters"]
                    .as_array()
                    .map(|a| a.iter().map(datum).collect::<Result<Vec<_>, _>>())
                    .transpose()?
                    .unwrap_or_default(),
            ),
            result_formats: i16s(value, "ResultFormatCodes"),
        },
        "*pgproto3.Describe" => Send::Describe(object_type(value), text(value, "Name")),
        "*pgproto3.Close" => Send::Close(object_type(value), text(value, "Name")),
        "*pgproto3.Execute" => Send::Execute(text(value, "Portal"), value["MaxRows"].as_u64().unwrap_or(0) as u32),
        "*pgproto3.Sync" => Send::Sync,
        "*pgproto3.Flush" => Send::Flush,
        "*pgproto3.CopyDone" => Send::CopyDone,
        "*pgproto3.CopyFail" => Send::CopyFail(text(value, "Message")),
        "*pgproto3.CopyData" => {
            let data = value["Data"].as_str().unwrap_or_default();
            Send::CopyData(leak_slice(hex_bytes(data)?))
        }
        _ => return Err(format!("unsupported frontend message {kind}")),
    })
}

/// hex_bytes decodes hexadecimal text.
fn hex_bytes(hex: &str) -> Result<Vec<u8>, String> {
    (0..hex.len()).step_by(2).map(|i| u8::from_str_radix(&hex[i..i + 2], 16).map_err(|err| err.to_string())).collect()
}

/// from_wire_script converts a Go WireScriptTest: each assertion sends its batch and receives what it draws.
pub fn from_wire_script(test: &Value) -> Result<WireTest, String> {
    let mut steps = Vec::new();
    for assertion in test["Assertions"].as_array().cloned().unwrap_or_default() {
        let sends = assertion["Send"].as_array().cloned().unwrap_or_default();
        steps.push(Step::Send(leak_slice(sends.iter().map(send).collect::<Result<Vec<_>, _>>()?)));
        steps.push(Step::Receive(&[]));
    }
    Ok(WireTest { name: text(test, "Name"), set_up_script: set_up(test), steps: leak_slice(steps), ..W })
}

/// set_up returns a test's setup statements.
fn set_up(test: &Value) -> &'static [&'static str] {
    leak_slice(
        test["SetUpScript"]
            .as_array()
            .map(|a| a.iter().map(|s| leak(s.as_str().unwrap())).collect())
            .unwrap_or_default(),
    )
}

/// from_message_flow converts a Go MessageFlowTest. Extended-protocol steps are batched until a Sync or Flush,
/// which receives everything the batch draws, and a simple query receives through its ReadyForQuery, stopping at
/// each CopyInResponse to send that COPY's input.
pub fn from_message_flow(test: &Value) -> Result<WireTest, String> {
    let mut steps = Vec::new();
    let mut batch: Vec<Send> = Vec::new();
    let flush = |batch: &mut Vec<Send>, steps: &mut Vec<Step>| {
        if !batch.is_empty() {
            steps.push(Step::Send(leak_slice(std::mem::take(batch))));
        }
    };
    for step in test["Steps"].as_array().cloned().unwrap_or_default() {
        let kind = step["$type"].as_str().unwrap_or_default();
        let value = &step["$value"];
        match kind {
            "_go.SimpleQuery" => {
                flush(&mut batch, &mut steps);
                steps.push(Step::Send(leak_slice(vec![Send::Query(text(value, "Query"))])));
                for input in value["CopyInputs"].as_array().cloned().unwrap_or_default() {
                    steps.push(Step::Receive(&[]));
                    let mut sends = Vec::new();
                    for message in input["BeforeData"].as_array().cloned().unwrap_or_default() {
                        sends.push(send(&message)?);
                    }
                    for chunk in input["Chunks"].as_array().cloned().unwrap_or_default() {
                        let bytes =
                            STANDARD.decode(chunk["$bytes"].as_str().unwrap_or_default()).map_err(|e| e.to_string())?;
                        sends.push(Send::CopyData(leak_slice(bytes)));
                    }
                    match input.get("FailMessage").and_then(Value::as_str) {
                        Some(message) if !message.is_empty() => sends.push(Send::CopyFail(leak(message))),
                        _ => sends.push(Send::CopyDone),
                    }
                    steps.push(Step::Send(leak_slice(sends)));
                }
                steps.push(Step::Receive(&[]));
            }
            "_go.Parse" => {
                batch.push(Send::Parse { name: text(value, "Name"), query: text(value, "Query"), parameter_oids: &[] })
            }
            "_go.Bind" => batch.push(Send::Bind {
                portal: text(value, "Portal"),
                statement: text(value, "PreparedStatement"),
                parameter_formats: &[],
                parameters: leak_slice(
                    value["Parameters"]
                        .as_array()
                        .map(|a| a.iter().map(|p| Datum::Text(leak(p.as_str().unwrap()))).collect())
                        .unwrap_or_default(),
                ),
                result_formats: &[],
            }),
            "_go.Describe" => batch.push(Send::Describe(object_type(value), text(value, "Name"))),
            "_go.Close" => batch.push(Send::Close(object_type(value), text(value, "Name"))),
            "_go.Execute" => batch.push(Send::Execute(text(value, "Portal"), 0)),
            "_go.Sync" => {
                batch.push(Send::Sync);
                flush(&mut batch, &mut steps);
                steps.push(Step::Receive(&[]));
            }
            "_go.Flush" => {
                batch.push(Send::Flush);
                flush(&mut batch, &mut steps);
                steps.push(Step::Receive(&[]));
            }
            "_go.QueryOnOtherConnection" => {
                flush(&mut batch, &mut steps);
                steps.push(Step::OtherQuery { query: text(value, "Query"), rows: &[] });
            }
            _ => return Err(format!("unsupported flow step {kind}")),
        }
    }
    flush(&mut batch, &mut steps);
    Ok(WireTest { name: text(test, "Name"), set_up_script: set_up(test), steps: leak_slice(steps), ..W })
}

/// capture_json captures a conversation and converts it into JSON, recording each received message as the Rust
/// literal that expects it.
pub fn capture_json(target: &harness::server::Target, test: &WireTest) -> Value {
    let capture = harness::wire::capture_wire_test(target, test);
    json!({
        "error": capture.error,
        "received": capture.received.iter().map(|(index, messages)| json!({
            "step": index,
            "messages": messages.iter().map(receive_code).collect::<Vec<_>>(),
        })).collect::<Vec<_>>(),
        "other_rows": capture.other_rows.iter().map(|(index, rows)| json!({"step": index, "rows": rows})).collect::<Vec<_>>(),
    })
}

/// send_code renders a frontend message as a Send literal.
pub fn send_code(message: &Send) -> String {
    let datum = |d: &Datum| match d {
        Datum::Null => "Datum::Null".to_string(),
        Datum::Text(text) => format!("Datum::Text({})", rust::string(text)),
        Datum::Bytes(bytes) => format!("Datum::Bytes(&{bytes:?})"),
    };
    match message {
        Send::Query(query) => format!("Send::Query({})", rust::string(query)),
        Send::Parse { name, query, parameter_oids } => format!(
            "Send::Parse {{ name: {}, query: {}, parameter_oids: &[{}] }}",
            rust::string(name),
            rust::string(query),
            parameter_oids.iter().map(|o| rust::type_oid(*o)).collect::<Vec<_>>().join(", ")
        ),
        Send::Bind { portal, statement, parameter_formats, parameters, result_formats } => format!(
            "Send::Bind {{ portal: {}, statement: {}, parameter_formats: &{parameter_formats:?}, parameters: &[{}], \
             result_formats: &{result_formats:?} }}",
            rust::string(portal),
            rust::string(statement),
            parameters.iter().map(datum).collect::<Vec<_>>().join(", ")
        ),
        Send::Describe(object_type, name) => {
            format!("Send::Describe(b'{}', {})", *object_type as char, rust::string(name))
        }
        Send::Execute(portal, max_rows) => format!("Send::Execute({}, {max_rows})", rust::string(portal)),
        Send::Close(object_type, name) => format!("Send::Close(b'{}', {})", *object_type as char, rust::string(name)),
        Send::Sync => "Send::Sync".to_string(),
        Send::Flush => "Send::Flush".to_string(),
        Send::CopyData(data) => format!("Send::CopyData({})", rust::bytes(data)),
        Send::CopyDone => "Send::CopyDone".to_string(),
        Send::CopyFail(message) => format!("Send::CopyFail({})", rust::string(message)),
    }
}

/// receive_code renders a received message, normalized, as a Receive literal.
pub fn receive_code(message: &BackendMessage) -> String {
    let message = harness::wire::normalize(message);
    let datum = |value: &Option<Vec<u8>>| match value {
        None => "Datum::Null".to_string(),
        Some(bytes) => match std::str::from_utf8(bytes) {
            Ok(text) => format!("Datum::Text({})", rust::string(text)),
            Err(_) => format!("Datum::Bytes(&{bytes:?})"),
        },
    };
    let fields = |f: &pgproto::ErrorFields| {
        let mut parts = Vec::new();
        let mut put = |name: &str, value: &str| {
            if !value.is_empty() {
                parts.push(format!("{name}: {}", rust::string(value)));
            }
        };
        put("severity", &f.severity);
        put("severity_unlocalized", &f.severity_unlocalized);
        put("code", &f.code);
        put("message", &f.message);
        put("detail", &f.detail);
        put("hint", &f.hint);
        put("internal_query", &f.internal_query);
        put("where_", &f.where_);
        put("schema", &f.schema_name);
        put("table", &f.table_name);
        put("column", &f.column_name);
        put("data_type", &f.data_type_name);
        put("constraint", &f.constraint_name);
        if f.position != 0 {
            parts.push(format!("position: {}", f.position));
        }
        if f.internal_position != 0 {
            parts.push(format!("internal_position: {}", f.internal_position));
        }
        format!("Fields {{ {}, ..F }}", parts.join(", "))
    };
    match &message {
        BackendMessage::ParseComplete => "Receive::ParseComplete".into(),
        BackendMessage::BindComplete => "Receive::BindComplete".into(),
        BackendMessage::CloseComplete => "Receive::CloseComplete".into(),
        BackendMessage::NoData => "Receive::NoData".into(),
        BackendMessage::EmptyQueryResponse => "Receive::EmptyQueryResponse".into(),
        BackendMessage::PortalSuspended => "Receive::PortalSuspended".into(),
        BackendMessage::CopyDone => "Receive::CopyDone".into(),
        BackendMessage::ParameterDescription { parameter_oids } => format!(
            "Receive::ParameterDescription(&[{}])",
            parameter_oids.iter().map(|o| rust::type_oid(*o)).collect::<Vec<_>>().join(", ")
        ),
        BackendMessage::RowDescription { fields } => format!(
            "Receive::RowDescription(&[{}])",
            fields
                .iter()
                .map(|f| format!(
                    "Field {{ name: {}, attnum: {}, type_oid: {}, size: {}, typmod: {}, format: {} }}",
                    rust::string(&f.name),
                    f.table_attribute_number,
                    if f.data_type_oid == 0 { "0".to_string() } else { rust::type_oid(f.data_type_oid) },
                    f.data_type_size,
                    f.type_modifier,
                    f.format
                ))
                .collect::<Vec<_>>()
                .join(", ")
        ),
        BackendMessage::DataRow { values } => {
            format!("Receive::DataRow(&[{}])", values.iter().map(datum).collect::<Vec<_>>().join(", "))
        }
        BackendMessage::CommandComplete { command_tag } => {
            format!("Receive::CommandComplete({})", rust::string(command_tag))
        }
        BackendMessage::ReadyForQuery { tx_status } => format!("Receive::ReadyForQuery(b'{}')", *tx_status as char),
        BackendMessage::ErrorResponse(f) => format!("Receive::Error({})", fields(f)),
        BackendMessage::NoticeResponse(f) => format!("Receive::Notice({})", fields(f)),
        BackendMessage::ParameterStatus { name, value } => {
            format!("Receive::ParameterStatus({}, {})", rust::string(name), rust::string(value))
        }
        BackendMessage::CopyInResponse { overall_format, column_format_codes } => {
            format!("Receive::CopyInResponse({overall_format}, &{column_format_codes:?})")
        }
        BackendMessage::CopyOutResponse { overall_format, column_format_codes } => {
            format!("Receive::CopyOutResponse({overall_format}, &{column_format_codes:?})")
        }
        BackendMessage::CopyData { data } => format!("Receive::CopyData(&{data:?})"),
        other => format!("/* unsupported message {other:?} */"),
    }
}
