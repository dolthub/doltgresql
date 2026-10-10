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

//! Compares the bytes that the Go suite's pgx client sent with the bytes the Rust harness sent for the same scripts,
//! proving that the ported tests exercise the server with the same messages.

use std::collections::BTreeMap;

use pgproto::{FrameReader, FrontendMessage, PasswordKind};

/// parse_connection splits the bytes one connection sent into messages, normalizing what legitimately varies:
/// startup parameters are sorted, SCRAM nonces and proofs are blanked, consecutive CopyData messages are joined,
/// and a final Terminate is dropped.
pub fn parse_connection(bytes: &[u8]) -> Result<Vec<String>, String> {
    let mut reader = FrameReader::new();
    reader.extend(bytes);
    let mut messages = Vec::new();
    let mut started = false;
    let mut sasl = false;
    while !reader.buffered().is_empty() {
        if !started {
            let body = reader.next_untyped_frame().map_err(|e| e.to_string())?.ok_or("truncated startup")?;
            let message = FrontendMessage::decode_startup(&body).map_err(|e| e.to_string())?;
            started = matches!(message, FrontendMessage::StartupMessage { .. });
            messages.push(match message {
                FrontendMessage::StartupMessage { protocol_version, mut parameters } => {
                    parameters.sort();
                    format!("StartupMessage {protocol_version} {parameters:?}")
                }
                other => format!("{other:?}"),
            });
            continue;
        }
        let frame = reader.next_frame().map_err(|e| e.to_string())?.ok_or("truncated message")?;
        let kind = if frame.tag != b'p' {
            PasswordKind::Password
        } else if frame.body.starts_with(b"SCRAM-SHA-256\0") {
            sasl = true;
            PasswordKind::SASLInitialResponse
        } else if sasl {
            PasswordKind::SASLResponse
        } else {
            PasswordKind::Password
        };
        let message = FrontendMessage::decode(frame.tag, &frame.body, kind).map_err(|e| e.to_string())?;
        let text = match message {
            FrontendMessage::SASLInitialResponse { auth_mechanism, data } => {
                let data = String::from_utf8_lossy(&data.unwrap_or_default()).into_owned();
                let header = data.split(",r=").next().unwrap_or_default().to_string();
                format!("SASLInitialResponse {auth_mechanism} {header},r=*")
            }
            FrontendMessage::SASLResponse { .. } => "SASLResponse *".to_string(),
            FrontendMessage::CopyData { data } => {
                if let Some(last) = messages.last_mut()
                    && let Some(previous) = last.strip_prefix("CopyData ")
                {
                    let mut joined = hex_decode(previous);
                    joined.extend_from_slice(&data);
                    *last = format!("CopyData {}", hex_encode(&joined));
                    continue;
                }
                format!("CopyData {}", hex_encode(&data))
            }
            other => format!("{other:?}"),
        };
        messages.push(text);
    }
    if messages.last().is_some_and(|m| m == "Terminate") {
        messages.pop();
    }
    Ok(messages)
}

/// read_raw reads a recording file into the typed messages each connection sent after it started up and
/// authenticated.
pub fn read_raw(path: &std::path::Path) -> Result<Vec<Vec<FrontendMessage>>, String> {
    let text = std::fs::read_to_string(path).map_err(|e| e.to_string())?;
    let mut connections = Vec::new();
    for line in text.lines().filter(|line| !line.is_empty()) {
        let mut reader = FrameReader::new();
        reader.extend(&hex_decode(line));
        let body = reader.next_untyped_frame().map_err(|e| e.to_string())?.ok_or("truncated startup")?;
        let mut messages = Vec::new();
        if matches!(FrontendMessage::decode_startup(&body), Ok(FrontendMessage::StartupMessage { .. })) {
            while let Some(frame) = reader.next_frame().map_err(|e| e.to_string())? {
                if frame.tag == b'p' {
                    continue;
                }
                messages.push(
                    FrontendMessage::decode(frame.tag, &frame.body, PasswordKind::Password)
                        .map_err(|e| e.to_string())?,
                );
            }
        }
        connections.push(messages);
    }
    Ok(connections)
}

/// hex_encode renders bytes as hexadecimal.
fn hex_encode(bytes: &[u8]) -> String {
    bytes.iter().map(|b| format!("{b:02x}")).collect()
}

/// hex_decode reads hexadecimal text.
fn hex_decode(text: &str) -> Vec<u8> {
    (0..text.len()).step_by(2).filter_map(|i| u8::from_str_radix(text.get(i..i + 2)?, 16).ok()).collect()
}

/// read_recording reads a recording file into the messages of each connection.
pub fn read_recording(path: &std::path::Path) -> Result<Vec<Vec<String>>, String> {
    let text = std::fs::read_to_string(path).map_err(|e| e.to_string())?;
    text.lines().filter(|line| !line.is_empty()).map(|line| parse_connection(&hex_decode(line))).collect()
}

/// compare compares every Go recording against the Rust recording of the same name, given a map from Go recording
/// names to Rust ones, and returns a report.
pub fn compare(go_dir: &str, rust_dir: &str, names: &BTreeMap<String, String>) -> String {
    let mut report = String::new();
    let (mut same, mut different, mut missing) = (0, 0, 0);
    for (go_name, rust_name) in names {
        let go_path = std::path::Path::new(go_dir).join(format!("{go_name}.hex"));
        let rust_path = std::path::Path::new(rust_dir).join(format!("{rust_name}.hex"));
        let (go, rust) = match (read_recording(&go_path), read_recording(&rust_path)) {
            (Ok(go), Ok(rust)) => (go, rust),
            (go, rust) => {
                missing += 1;
                report.push_str(&format!("MISSING {go_name}: go {:?} rust {:?}\n", go.err(), rust.err()));
                continue;
            }
        };
        if go == rust {
            same += 1;
            continue;
        }
        different += 1;
        report.push_str(&format!("DIFFERENT {go_name}\n"));
        for (index, (g, r)) in go.iter().zip(rust.iter()).enumerate() {
            if g != r {
                let position = g.iter().zip(r.iter()).position(|(a, b)| a != b).unwrap_or(g.len().min(r.len()));
                report.push_str(&format!(
                    "  connection {index}, message {position}:\n    go:   {:?}\n    rust: {:?}\n",
                    g.get(position),
                    r.get(position)
                ));
                break;
            }
        }
        if go.len() != rust.len() {
            report.push_str(&format!("  connection count: go {} rust {}\n", go.len(), rust.len()));
        }
    }
    format!("same {same}, different {different}, missing {missing}\n\n{report}")
}
