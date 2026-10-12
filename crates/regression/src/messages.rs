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

use std::path::Path;

use pgproto::{BackendMessage, FrontendMessage, PasswordKind};

use crate::format::Reader;

/// Message is one recorded message, where COPY messages have no side since the recording does not say who sent them.
#[derive(Clone, Debug, PartialEq)]
#[allow(clippy::large_enum_variant)]
pub enum Message {
    Frontend(FrontendMessage),
    Backend(BackendMessage),
    CopyData(Vec<u8>),
    CopyDone,
}

/// Direction says how a message type's body is decoded.
#[derive(Clone, Copy)]
enum Direction {
    Backend(u8),
    Frontend(u8, PasswordKind),
    Startup,
    CopyData,
    CopyDone,
}

/// TYPES lists every message type in the order of the recorded type numbers, which must never change.
const TYPES: [(&str, Direction); 49] = [
    ("AuthenticationCleartextPassword", Direction::Backend(b'R')),
    ("AuthenticationGSS", Direction::Backend(b'R')),
    ("AuthenticationGSSContinue", Direction::Backend(b'R')),
    ("AuthenticationMD5Password", Direction::Backend(b'R')),
    ("AuthenticationOk", Direction::Backend(b'R')),
    ("AuthenticationSASL", Direction::Backend(b'R')),
    ("AuthenticationSASLContinue", Direction::Backend(b'R')),
    ("AuthenticationSASLFinal", Direction::Backend(b'R')),
    ("BackendKeyData", Direction::Backend(b'K')),
    ("Bind", Direction::Frontend(b'B', PasswordKind::Password)),
    ("BindComplete", Direction::Backend(b'2')),
    ("CancelRequest", Direction::Startup),
    ("Close", Direction::Frontend(b'C', PasswordKind::Password)),
    ("CloseComplete", Direction::Backend(b'3')),
    ("CommandComplete", Direction::Backend(b'C')),
    ("CopyBothResponse", Direction::Backend(b'W')),
    ("CopyData", Direction::CopyData),
    ("CopyDone", Direction::CopyDone),
    ("CopyFail", Direction::Frontend(b'f', PasswordKind::Password)),
    ("CopyInResponse", Direction::Backend(b'G')),
    ("CopyOutResponse", Direction::Backend(b'H')),
    ("DataRow", Direction::Backend(b'D')),
    ("Describe", Direction::Frontend(b'D', PasswordKind::Password)),
    ("EmptyQueryResponse", Direction::Backend(b'I')),
    ("ErrorResponse", Direction::Backend(b'E')),
    ("Execute", Direction::Frontend(b'E', PasswordKind::Password)),
    ("Flush", Direction::Frontend(b'H', PasswordKind::Password)),
    ("FunctionCall", Direction::Frontend(b'F', PasswordKind::Password)),
    ("FunctionCallResponse", Direction::Backend(b'V')),
    ("GSSEncRequest", Direction::Startup),
    ("GSSResponse", Direction::Frontend(b'p', PasswordKind::GSSResponse)),
    ("NoData", Direction::Backend(b'n')),
    ("NoticeResponse", Direction::Backend(b'N')),
    ("NotificationResponse", Direction::Backend(b'A')),
    ("ParameterDescription", Direction::Backend(b't')),
    ("ParameterStatus", Direction::Backend(b'S')),
    ("Parse", Direction::Frontend(b'P', PasswordKind::Password)),
    ("ParseComplete", Direction::Backend(b'1')),
    ("PasswordMessage", Direction::Frontend(b'p', PasswordKind::Password)),
    ("PortalSuspended", Direction::Backend(b's')),
    ("Query", Direction::Frontend(b'Q', PasswordKind::Password)),
    ("ReadyForQuery", Direction::Backend(b'Z')),
    ("RowDescription", Direction::Backend(b'T')),
    ("SASLInitialResponse", Direction::Frontend(b'p', PasswordKind::SASLInitialResponse)),
    ("SASLResponse", Direction::Frontend(b'p', PasswordKind::SASLResponse)),
    ("SSLRequest", Direction::Startup),
    ("StartupMessage", Direction::Startup),
    ("Sync", Direction::Frontend(b'S', PasswordKind::Password)),
    ("Terminate", Direction::Frontend(b'X', PasswordKind::Password)),
];

/// decode decodes one recorded message from its type number and body.
pub fn decode(message_type: u16, body: &[u8]) -> Result<Message, String> {
    let (name, direction) =
        TYPES.get(message_type as usize).ok_or_else(|| format!("unknown message type {message_type}"))?;
    let decoded = match *direction {
        Direction::Backend(tag) => BackendMessage::decode(tag, body).map(Message::Backend),
        Direction::Frontend(tag, kind) => FrontendMessage::decode(tag, body, kind).map(Message::Frontend),
        Direction::Startup => FrontendMessage::decode_startup(body).map(Message::Frontend),
        Direction::CopyData => Ok(Message::CopyData(body.to_vec())),
        Direction::CopyDone => Ok(Message::CopyDone),
    };
    decoded.map_err(|e| format!("{name}: {e}"))
}

/// read_messages reads a `.results` file, pointing each COPY ... FROM query at the data directory.
pub fn read_messages(path: &Path, data_dir: &Path) -> Result<Vec<Message>, String> {
    let data = std::fs::read(path).map_err(|e| format!("{}: {e}", path.display()))?;
    let mut reader = Reader::new(&data);
    let count = reader.u32()?;
    let mut messages = Vec::with_capacity(count as usize);
    for _ in 0..count {
        let message_type = reader.u16()?;
        let mut message = decode(message_type, reader.bytes()?)?;
        if let Message::Frontend(FrontendMessage::Query { query }) = &mut message {
            *query = rewrite_copy_to_local(query, data_dir)?;
        }
        messages.push(message);
    }
    if !reader.is_empty() {
        return Err(format!("{}: file has additional data", path.display()));
    }
    Ok(messages)
}

/// copy_from_path finds the file path in a `COPY <table> FROM '<path>'` query, matching
/// `COPY\s+\S+\s+FROM\s+'([a-zA-Z0-9_\/\\%#@!~+=:.-]+)'`.
fn copy_from_path(query: &str) -> Option<&str> {
    let bytes = query.as_bytes();
    let is_space = |b: u8| matches!(b, b' ' | b'\t' | b'\n' | b'\x0c' | b'\r');
    let is_path = |b: u8| b.is_ascii_alphanumeric() || b"_/\\%#@!~+=:.-".contains(&b);
    let mut start = 0;
    while let Some(found) = query[start..].find("COPY") {
        let at = start + found;
        start = at + 1;
        let mut i = at + 4;
        let spaces = |i: &mut usize| {
            let from = *i;
            while *i < bytes.len() && is_space(bytes[*i]) {
                *i += 1;
            }
            *i > from
        };
        if !spaces(&mut i) {
            continue;
        }
        let table_start = i;
        while i < bytes.len() && !is_space(bytes[i]) {
            i += 1;
        }
        if i == table_start {
            continue;
        }
        let mut j = i;
        if !spaces(&mut j) || !query[j..].starts_with("FROM") {
            continue;
        }
        j += 4;
        if !spaces(&mut j) || bytes.get(j) != Some(&b'\'') {
            continue;
        }
        let path_start = j + 1;
        let mut end = path_start;
        while end < bytes.len() && is_path(bytes[end]) {
            end += 1;
        }
        if end > path_start && bytes.get(end) == Some(&b'\'') {
            return Some(&query[path_start..end]);
        }
    }
    None
}

/// rewrite_copy_to_local points a COPY ... FROM query at the data file of the same name in the data directory.
fn rewrite_copy_to_local(query: &str, data_dir: &Path) -> Result<String, String> {
    let Some(path) = copy_from_path(query) else {
        return Ok(query.to_string());
    };
    let name = path.rsplit('/').next().unwrap_or(path);
    let local = data_dir.join(name);
    if !local.exists() {
        return Err(format!("file does not exist: '{}'", local.display()));
    }
    Ok(query.replace(path, &local.to_string_lossy()))
}
