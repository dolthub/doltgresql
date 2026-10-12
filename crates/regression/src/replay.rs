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

use std::collections::VecDeque;
use std::io::{Read, Write};
use std::net::{Shutdown, TcpStream};
use std::time::{Duration, Instant};

use harness::pgx::ScramClient;
use pgproto::{BackendMessage, ErrorFields, FieldDescription, FrameReader, FrontendMessage};

use crate::compare::{self, OidMap, Row};
use crate::messages::Message;
use crate::tracker::Tracker;

/// TIMEOUT is how long a single send or receive may take before the connection is abandoned.
const TIMEOUT: Duration = Duration::from_secs(15);

/// CELLS_ENV names a file that collects every cell the server returns, for checking the cell decoding.
pub const CELLS_ENV: &str = "DOLTGRES_REGRESSION_CELLS";

/// record_cells appends the server's cells to the file named by CELLS_ENV, when it is set.
fn record_cells(fields: &[FieldDescription], rows: &[Row]) {
    static FILE: std::sync::OnceLock<Option<std::sync::Mutex<std::fs::File>>> = std::sync::OnceLock::new();
    let file = FILE.get_or_init(|| {
        let path = std::env::var(CELLS_ENV).ok()?;
        let file = std::fs::OpenOptions::new().create(true).append(true).open(path).ok()?;
        Some(std::sync::Mutex::new(file))
    });
    let Some(file) = file else {
        return;
    };
    let mut text = String::new();
    for row in rows.iter().filter(|row| row.len() == fields.len()) {
        for (field, cell) in fields.iter().zip(row) {
            let value = cell.as_ref().map_or("-".to_string(), |v| v.iter().map(|b| format!("{b:02x}")).collect());
            text.push_str(&format!("{} {value}\n", field.data_type_oid));
        }
    }
    let _ = file.lock().unwrap().write_all(text.as_bytes());
}

/// PSQL_INDEX_QUERY starts the query that psql's `\d` runs to list indexes, which is failed without being sent.
const PSQL_INDEX_QUERY: &str = "SELECT c2.relname, i.indisprimary, i.indisunique, i.indisclustered, i.indisvalid, \
                                pg_catalog.pg_get_indexdef(i.indexrelid, 0, true),";

/// Options configures a replay.
pub struct Options<'a> {
    pub file: &'a str,
    pub port: u16,
    pub messages: Vec<Message>,
    pub print_queries: bool,
    pub fail_psql: bool,
    pub fail_queries: &'a [&'a str],
    pub password: &'a str,
}

/// MessageReader iterates over the recorded messages, with a queue of messages to return first.
struct MessageReader {
    messages: Vec<Message>,
    queue: VecDeque<Message>,
    index: usize,
}

impl MessageReader {
    fn is_empty(&self) -> bool {
        self.queue.is_empty() && self.index >= self.messages.len()
    }

    fn next(&mut self) -> Option<Message> {
        if let Some(message) = self.queue.pop_front() {
            return Some(message);
        }
        let message = self.messages.get(self.index).cloned();
        if message.is_some() {
            self.index += 1;
        }
        message
    }

    fn peek(&self) -> Option<&Message> {
        self.queue.front().or_else(|| self.messages.get(self.index))
    }

    /// sync_to_next_query moves past the next ReadyForQuery, or up to the next Terminate.
    fn sync_to_next_query(&mut self) {
        loop {
            match self.next() {
                Some(Message::Backend(BackendMessage::ReadyForQuery { .. })) | None => return,
                Some(Message::Frontend(FrontendMessage::Terminate)) => {
                    self.index = self.index.saturating_sub(1);
                    return;
                }
                Some(_) => {}
            }
        }
    }
}

/// filter_messages keeps the recorded messages that the replay acts on.
fn filter_messages(messages: Vec<Message>) -> Vec<Message> {
    messages
        .into_iter()
        .filter(|message| match message {
            Message::Frontend(m) => !matches!(
                m,
                FrontendMessage::Flush
                    | FrontendMessage::GSSEncRequest
                    | FrontendMessage::GSSResponse { .. }
                    | FrontendMessage::PasswordMessage { .. }
                    | FrontendMessage::SASLInitialResponse { .. }
                    | FrontendMessage::SASLResponse { .. }
                    | FrontendMessage::SSLRequest
            ),
            Message::Backend(m) => !matches!(
                m,
                BackendMessage::AuthenticationOk
                    | BackendMessage::AuthenticationCleartextPassword
                    | BackendMessage::AuthenticationGSS
                    | BackendMessage::AuthenticationGSSContinue { .. }
                    | BackendMessage::AuthenticationMD5Password { .. }
                    | BackendMessage::AuthenticationSASL { .. }
                    | BackendMessage::AuthenticationSASLContinue { .. }
                    | BackendMessage::AuthenticationSASLFinal { .. }
                    | BackendMessage::BackendKeyData { .. }
                    | BackendMessage::NoticeResponse(_)
                    | BackendMessage::NotificationResponse { .. }
                    | BackendMessage::ParameterDescription { .. }
                    | BackendMessage::ParameterStatus { .. }
                    | BackendMessage::PortalSuspended
                    | BackendMessage::NegotiateProtocolVersion { .. }
            ),
            Message::CopyData(_) | Message::CopyDone => true,
        })
        .collect()
}

/// Connection is the replay's connection to the server, which requeues the startup messages when it fails.
struct Connection {
    stream: TcpStream,
    frames: FrameReader,
    pending: Vec<u8>,
    startup: Option<Message>,
}

impl Connection {
    fn connect(port: u16) -> Result<Connection, String> {
        let stream = TcpStream::connect(("127.0.0.1", port)).map_err(|e| e.to_string())?;
        stream.set_nodelay(true).map_err(|e| e.to_string())?;
        stream.set_write_timeout(Some(TIMEOUT)).map_err(|e| e.to_string())?;
        Ok(Connection { stream, frames: FrameReader::new(), pending: Vec::new(), startup: None })
    }

    fn queue(&mut self, message: &FrontendMessage) {
        if matches!(message, FrontendMessage::StartupMessage { .. }) {
            self.startup = Some(Message::Frontend(message.clone()));
            let mut buffer = Vec::new();
            message.encode(&mut buffer);
            self.pending.extend_from_slice(&buffer);
        } else {
            message.encode(&mut self.pending);
        }
    }

    /// fail closes the connection after an error, optionally moving the reader to the next query.
    fn fail(&mut self, reader: &mut MessageReader, sync: bool, error: String) -> String {
        if sync {
            reader.sync_to_next_query();
        }
        if let Some(startup) = self.startup.clone() {
            reader.queue.push_back(startup);
        }
        reader.queue.push_back(Message::Backend(BackendMessage::ReadyForQuery { tx_status: b'I' }));
        let _ = self.stream.shutdown(Shutdown::Both);
        error
    }

    /// send writes the queued messages, moving the reader to the next query on error when sync is set.
    fn send(&mut self, reader: &mut MessageReader, sync: bool) -> Result<(), String> {
        let pending = std::mem::take(&mut self.pending);
        match self.stream.write_all(&pending).and_then(|_| self.stream.flush()) {
            Ok(()) => Ok(()),
            Err(e) if matches!(e.kind(), std::io::ErrorKind::WouldBlock | std::io::ErrorKind::TimedOut) => {
                Err(self.fail(reader, sync, "timeout during Send".to_string()))
            }
            Err(e) => Err(self.fail(reader, sync, e.to_string())),
        }
    }

    /// read_frame reads the next backend message within the deadline.
    fn read_frame(&mut self, deadline: Instant) -> Result<BackendMessage, String> {
        loop {
            if let Some(frame) = self.frames.next_frame().map_err(|e| e.to_string())? {
                return BackendMessage::decode(frame.tag, &frame.body).map_err(|e| e.to_string());
            }
            let remaining = deadline.saturating_duration_since(Instant::now());
            if remaining.is_zero() {
                return Err("timeout during Receive".to_string());
            }
            self.stream.set_read_timeout(Some(remaining)).map_err(|e| e.to_string())?;
            let mut buffer = [0u8; 65536];
            match self.stream.read(&mut buffer) {
                Ok(0) => return Err("unexpected EOF".to_string()),
                Ok(n) => self.frames.extend(&buffer[..n]),
                Err(e) if matches!(e.kind(), std::io::ErrorKind::WouldBlock | std::io::ErrorKind::TimedOut) => {
                    return Err("timeout during Receive".to_string());
                }
                Err(e) => return Err(e.to_string()),
            }
        }
    }

    /// receive reads the next backend message.
    fn receive(&mut self, reader: &mut MessageReader) -> Result<BackendMessage, String> {
        self.read_frame(Instant::now() + TIMEOUT).map_err(|e| self.fail(reader, false, e))
    }

    /// empty_receive_buffer discards anything already read past the last message, reporting whether there was any.
    fn empty_receive_buffer(&mut self) -> Result<(), String> {
        if self.frames.buffered().is_empty() {
            return Ok(());
        }
        while !self.frames.buffered().is_empty() {
            if self.read_frame(Instant::now() + TIMEOUT).is_err() {
                break;
            }
        }
        Err("Doltgres sent additional messages after ReadyForQuery".to_string())
    }

    /// authenticate answers a SASL request with SCRAM-SHA-256.
    fn authenticate(&mut self, reader: &mut MessageReader, user: &str, password: &str) -> Result<(), String> {
        let mut client = ScramClient::with_user(password, user);
        let initial = FrontendMessage::SASLInitialResponse {
            auth_mechanism: "SCRAM-SHA-256".to_string(),
            data: Some(client.client_first_message()),
        };
        self.queue(&initial);
        self.send(reader, false)?;
        let BackendMessage::AuthenticationSASLContinue { data } = self.receive(reader)? else {
            return Err("expected AuthenticationSASLContinue".to_string());
        };
        let response = FrontendMessage::SASLResponse { data: client.client_final_message(&data)? };
        self.queue(&response);
        self.send(reader, false)?;
        let BackendMessage::AuthenticationSASLFinal { data } = self.receive(reader)? else {
            return Err("expected AuthenticationSASLFinal".to_string());
        };
        client.verify_server_final_message(&data)
    }
}

/// RecordedCopy keeps each recorded COPY FROM STDIN input apart, along with the recorded COPY TO STDOUT output.
#[derive(Default)]
struct RecordedCopy {
    direction: Option<bool>,
    inputs: Vec<Vec<Vec<u8>>>,
    output: Vec<Vec<u8>>,
}

impl RecordedCopy {
    fn record(&mut self, message: &Message) {
        match message {
            Message::Backend(BackendMessage::CopyInResponse { .. }) => {
                self.direction = Some(true);
                self.inputs.push(Vec::new());
            }
            Message::Backend(BackendMessage::CopyOutResponse { .. }) => self.direction = Some(false),
            Message::CopyData(data) => match self.direction {
                Some(true) => self.inputs.last_mut().unwrap().push(data.clone()),
                Some(false) => self.output.push(data.clone()),
                None => {}
            },
            Message::CopyDone | Message::Backend(BackendMessage::CommandComplete { .. }) => self.direction = None,
            _ => {}
        }
    }
}

/// is_asynchronous reports whether a message may arrive at any point of a response and is ignored.
fn is_asynchronous(message: &BackendMessage) -> bool {
    matches!(
        message,
        BackendMessage::NoticeResponse(_)
            | BackendMessage::ParameterStatus { .. }
            | BackendMessage::NotificationResponse { .. }
    )
}

/// unknown is the error for a message that the replay does not expect, which stops the replay.
fn unknown(message: &impl std::fmt::Debug) -> String {
    let text = format!("{message:?}");
    format!("unable to determine what to do with {}", text.split([' ', '{', '(']).next().unwrap_or_default())
}

/// error_message returns the message of an error response, or an empty string.
fn error_message(fields: &Option<ErrorFields>) -> &str {
    fields.as_ref().map_or("", |f| f.message.as_str())
}

/// replay replays the messages onto the server, returning an error when the replay cannot continue.
pub fn replay(options: Options) -> Result<Tracker, String> {
    let mut tracker = Tracker::new(options.file);
    let mut reader = MessageReader { messages: filter_messages(options.messages), queue: VecDeque::new(), index: 0 };
    let mut oid_map = OidMap::default();
    let started = Instant::now();
    println!("-------------------- {} --------------------", tracker.file);
    'listener: while !reader.is_empty() {
        let mut connection = Connection::connect(options.port)?;
        let Some(Message::Frontend(startup @ FrontendMessage::StartupMessage { .. })) = reader.next() else {
            return Err(format!("{}: first message is not StartupMessage", options.file));
        };
        if !matches!(reader.next(), Some(Message::Backend(BackendMessage::ReadyForQuery { .. }))) {
            return Err("expected message after StartupMessage to be ReadyForQuery".to_string());
        }
        connection.queue(&startup);
        connection.send(&mut reader, false)?;
        loop {
            match connection.receive(&mut reader)? {
                BackendMessage::AuthenticationOk
                | BackendMessage::BackendKeyData { .. }
                | BackendMessage::ParameterStatus { .. } => {}
                BackendMessage::AuthenticationSASL { .. } => {
                    let FrontendMessage::StartupMessage { parameters, .. } = &startup else { unreachable!() };
                    let user = parameters.iter().find(|(k, _)| k == "user").map_or("", |(_, v)| v.as_str());
                    connection.authenticate(&mut reader, user, options.password)?;
                }
                BackendMessage::ErrorResponse(fields) => return Err(fields.message),
                BackendMessage::ReadyForQuery { .. } => break,
                other => return Err(format!("unknown StartupMessage response type: {}", unknown(&other))),
            }
        }
        while let Some(message) = reader.next() {
            match message {
                // TODO: messages are out of order in `copy2`, so a stray COPY message skips to the next query
                Message::CopyData(_) => reader.sync_to_next_query(),
                Message::Frontend(FrontendMessage::Describe { .. }) => {
                    let Message::Frontend(describe) = &message else { unreachable!() };
                    connection.queue(describe);
                    if matches!(reader.peek(), Some(Message::Frontend(FrontendMessage::Sync))) {
                        reader.next();
                        connection.queue(&FrontendMessage::Sync);
                    }
                    if let Err(e) = connection.send(&mut reader, true) {
                        tracker.fail("DESCRIBE", &e, "");
                        continue 'listener;
                    }
                    let mut expected_error = None;
                    let mut expected_fields = None;
                    loop {
                        match reader.next() {
                            Some(Message::Backend(BackendMessage::EmptyQueryResponse | BackendMessage::NoData)) => {}
                            Some(Message::Backend(BackendMessage::ErrorResponse(fields))) => {
                                expected_error = Some(fields)
                            }
                            Some(Message::Backend(BackendMessage::ReadyForQuery { .. })) => break,
                            Some(Message::Backend(BackendMessage::RowDescription { fields })) => {
                                expected_fields = Some(fields)
                            }
                            other => return Err(unknown(&other)),
                        }
                    }
                    let mut response_error = None;
                    let mut response_fields = None;
                    loop {
                        let response = match connection.receive(&mut reader) {
                            Ok(response) => response,
                            Err(e) => {
                                tracker.fail("DESCRIBE", &e, "");
                                continue 'listener;
                            }
                        };
                        match response {
                            response if is_asynchronous(&response) => {}
                            BackendMessage::EmptyQueryResponse
                            | BackendMessage::NoData
                            | BackendMessage::ParameterDescription { .. } => {}
                            BackendMessage::ErrorResponse(fields) => response_error = Some(fields),
                            BackendMessage::ReadyForQuery { .. } => break,
                            BackendMessage::RowDescription { fields } => response_fields = Some(fields),
                            other => return Err(unknown(&other)),
                        }
                    }
                    if let Err(e) = connection.empty_receive_buffer() {
                        tracker.fail("DESCRIBE", &e, "");
                        continue;
                    }
                    if expected_error.is_none() {
                        if let Some(response_error) = &response_error {
                            tracker.fail("DESCRIBE", &response_error.message, "");
                            continue;
                        }
                        let Some(expected_fields) = expected_fields else {
                            if response_fields.is_none() {
                                tracker.success += 1;
                            } else {
                                tracker.fail("DESCRIBE", "expected no row description but received a description", "");
                            }
                            continue;
                        };
                        let Some(response_fields) = response_fields else {
                            tracker.fail("DESCRIBE", "expected rows but received none", "");
                            continue;
                        };
                        if expected_fields.len() != response_fields.len() {
                            tracker.fail(
                                "DESCRIBE",
                                &format!(
                                    "expected column count {} but received {}",
                                    expected_fields.len(),
                                    response_fields.len()
                                ),
                                "",
                            );
                            continue;
                        }
                        tracker.success += 1;
                        if expected_fields.iter().zip(&response_fields).any(|(e, r)| e.name != r.name) {
                            tracker.partial_success += 1;
                        }
                    } else {
                        let expected_message = error_message(&expected_error).to_string();
                        let Some(response_error) = response_error else {
                            tracker.fail("DESCRIBE", "", &expected_message);
                            continue;
                        };
                        tracker.success += 1;
                        if expected_message != response_error.message {
                            tracker.partial_success += 1;
                            tracker.fail_partial_items.push(crate::tracker::Item {
                                query: "DESCRIBE".to_string(),
                                partial_success: Vec::new(),
                                unexpected_error: response_error.message,
                                expected_error: expected_message,
                            });
                        }
                    }
                }
                Message::Frontend(FrontendMessage::FunctionCall {
                    function,
                    argument_format_codes,
                    arguments,
                    result_format_code,
                }) => {
                    let label = format!("Function OID: {function}");
                    let call = FrontendMessage::FunctionCall {
                        function: oid_map.get(function).unwrap_or(function),
                        argument_format_codes,
                        arguments,
                        result_format_code,
                    };
                    connection.queue(&call);
                    if let Err(e) = connection.send(&mut reader, true) {
                        tracker.fail(&label, &e, "");
                        continue 'listener;
                    }
                    let mut expected_error = None;
                    let mut expected_data = None;
                    loop {
                        match reader.next() {
                            Some(Message::Backend(BackendMessage::ErrorResponse(fields))) => {
                                expected_error = Some(fields)
                            }
                            Some(Message::Backend(BackendMessage::FunctionCallResponse { result })) => {
                                expected_data = Some(result)
                            }
                            Some(Message::Backend(BackendMessage::ReadyForQuery { .. })) => break,
                            other => return Err(unknown(&other)),
                        }
                    }
                    let mut response_error = None;
                    let mut response_data = None;
                    loop {
                        let response = match connection.receive(&mut reader) {
                            Ok(response) => response,
                            Err(e) => {
                                tracker.fail(&label, &e, "");
                                continue 'listener;
                            }
                        };
                        match response {
                            response if is_asynchronous(&response) => {}
                            BackendMessage::EmptyQueryResponse => {}
                            BackendMessage::ErrorResponse(fields) => response_error = Some(fields),
                            BackendMessage::FunctionCallResponse { result } => response_data = Some(result),
                            BackendMessage::ReadyForQuery { .. } => break,
                            other => return Err(unknown(&other)),
                        }
                    }
                    if let Err(e) = connection.empty_receive_buffer() {
                        tracker.fail(&label, &e, "");
                        continue;
                    }
                    if expected_error.is_none() {
                        if let Some(response_error) = response_error {
                            tracker.fail(&label, &response_error.message, "");
                            continue;
                        }
                        match (expected_data, response_data) {
                            (Some(_), None) => {
                                tracker.fail(&label, "expected a result but received no result", "");
                                continue;
                            }
                            (Some(e), Some(r))
                                if e.as_deref().unwrap_or_default() != r.as_deref().unwrap_or_default() =>
                            {
                                tracker.fail(&label, "result is incorrect", "");
                                continue;
                            }
                            (None, Some(_)) => {
                                tracker.fail(&label, "expected no result but received a result", "");
                                continue;
                            }
                            _ => {}
                        }
                        tracker.succeed(&label);
                    } else {
                        let expected_message = error_message(&expected_error).to_string();
                        let Some(response_error) = response_error else {
                            tracker.fail(&label, "", &expected_message);
                            continue;
                        };
                        tracker.succeed(&label);
                        if expected_message != response_error.message {
                            tracker.partial_success += 1;
                            tracker.fail_partial_items.push(crate::tracker::Item {
                                query: label,
                                partial_success: Vec::new(),
                                unexpected_error: response_error.message,
                                expected_error: expected_message,
                            });
                        }
                    }
                }
                Message::Frontend(FrontendMessage::Parse { name, query, parameter_oids }) => {
                    let parse = FrontendMessage::Parse { name, query: oid_map.rewrite_query(&query), parameter_oids };
                    connection.queue(&parse);
                    if matches!(reader.peek(), Some(Message::Frontend(FrontendMessage::Sync))) {
                        reader.next();
                        connection.queue(&FrontendMessage::Sync);
                    }
                    if let Err(e) = connection.send(&mut reader, true) {
                        tracker.fail(&query, &e, "");
                        continue 'listener;
                    }
                    let mut expected_error = None;
                    loop {
                        match reader.next() {
                            Some(Message::Backend(
                                BackendMessage::EmptyQueryResponse
                                | BackendMessage::NoData
                                | BackendMessage::ParseComplete,
                            )) => {}
                            Some(Message::Backend(BackendMessage::ErrorResponse(fields))) => {
                                expected_error = Some(fields)
                            }
                            Some(Message::Backend(BackendMessage::ReadyForQuery { .. })) => break,
                            other => return Err(unknown(&other)),
                        }
                    }
                    let mut response_error = None;
                    loop {
                        let response = match connection.receive(&mut reader) {
                            Ok(response) => response,
                            Err(e) => {
                                tracker.fail(&query, &e, "");
                                continue 'listener;
                            }
                        };
                        match response {
                            response if is_asynchronous(&response) => {}
                            BackendMessage::EmptyQueryResponse
                            | BackendMessage::NoData
                            | BackendMessage::ParseComplete => {}
                            BackendMessage::ErrorResponse(fields) => response_error = Some(fields),
                            BackendMessage::ReadyForQuery { .. } => break,
                            other => return Err(unknown(&other)),
                        }
                    }
                    if let Err(e) = connection.empty_receive_buffer() {
                        tracker.fail(&query, &e, "");
                        continue;
                    }
                    verdict_without_rows(&mut tracker, &query, expected_error, response_error);
                }
                Message::Frontend(FrontendMessage::Query { query }) => {
                    if options.print_queries {
                        println!("QUERY: {query}");
                    }
                    if options.fail_psql && query.starts_with(PSQL_INDEX_QUERY) {
                        tracker.fail(&query, "set to automatically fail PSQL commands", "");
                        reader.sync_to_next_query();
                        continue;
                    }
                    if options.fail_queries.iter().any(|skip| query.contains(skip)) {
                        tracker.fail(
                            &query,
                            "set to automatically fail due to catastrophic error (OOM, stack limit, etc.)",
                            "",
                        );
                        reader.sync_to_next_query();
                        continue;
                    }
                    connection.queue(&FrontendMessage::Query { query: oid_map.rewrite_query(&query) });
                    if let Err(e) = connection.send(&mut reader, true) {
                        tracker.fail(&query, &e, "");
                        continue 'listener;
                    }
                    if replay_query(&mut connection, &mut reader, &mut tracker, &mut oid_map, &query)? {
                        continue 'listener;
                    }
                }
                Message::Frontend(FrontendMessage::Terminate) => {
                    connection.queue(&FrontendMessage::Terminate);
                    connection.send(&mut reader, false)?;
                    break;
                }
                other => return Err(unknown(&other)),
            }
        }
        let _ = connection.stream.shutdown(Shutdown::Both);
    }
    println!(
        "-------------------- {} done in {:.6}s --------------------",
        tracker.file,
        started.elapsed().as_secs_f64()
    );
    Ok(tracker)
}

/// verdict_without_rows records the result of a statement whose result has no rows.
fn verdict_without_rows(
    tracker: &mut Tracker,
    query: &str,
    expected_error: Option<ErrorFields>,
    response_error: Option<ErrorFields>,
) {
    match (expected_error, response_error) {
        (None, None) => tracker.succeed(query),
        (None, Some(response)) => tracker.fail(query, &response.message, ""),
        (Some(expected), None) => tracker.fail(query, "", &expected.message),
        (Some(expected), Some(response)) => {
            tracker.succeed(query);
            if expected.message != response.message {
                tracker.partial_success += 1;
                tracker.fail_partial_items.push(crate::tracker::Item {
                    query: query.to_string(),
                    partial_success: Vec::new(),
                    unexpected_error: response.message,
                    expected_error: expected.message,
                });
            }
        }
    }
}

/// replay_query records the verdict of a simple query, returning true when the replay must reconnect.
fn replay_query(
    connection: &mut Connection,
    reader: &mut MessageReader,
    tracker: &mut Tracker,
    oid_map: &mut OidMap,
    query: &str,
) -> Result<bool, String> {
    let mut expected_error = None;
    let mut expected_fields: Option<Vec<FieldDescription>> = None;
    let mut expected_rows: Vec<Row> = Vec::new();
    let mut recorded_copy = RecordedCopy::default();
    loop {
        let message = reader.next();
        match &message {
            Some(
                m @ (Message::Backend(
                    BackendMessage::CommandComplete { .. }
                    | BackendMessage::CopyInResponse { .. }
                    | BackendMessage::CopyOutResponse { .. },
                )
                | Message::CopyData(_)
                | Message::CopyDone),
            ) => recorded_copy.record(m),
            Some(Message::Backend(BackendMessage::DataRow { values })) => expected_rows.push(values.clone()),
            Some(Message::Backend(BackendMessage::EmptyQueryResponse)) => {}
            Some(Message::Backend(BackendMessage::ErrorResponse(fields))) => expected_error = Some(fields.clone()),
            Some(Message::Backend(BackendMessage::ReadyForQuery { .. })) => break,
            Some(Message::Backend(BackendMessage::RowDescription { fields })) => expected_fields = Some(fields.clone()),
            other => return Err(unknown(other)),
        }
    }
    let mut response_error: Option<ErrorFields> = None;
    let mut response_fields: Option<Vec<FieldDescription>> = None;
    let mut response_rows: Vec<Row> = Vec::new();
    let mut response_copy_data: Vec<Vec<u8>> = Vec::new();
    let mut received_copy_out = false;
    let mut next_copy_input = 0;
    let mut unexpected_copy_from = false;
    loop {
        let response = match connection.receive(reader) {
            Ok(response) => response,
            Err(e) => {
                tracker.fail(query, &e, "");
                return Ok(true);
            }
        };
        match response {
            response if is_asynchronous(&response) => {}
            BackendMessage::CommandComplete { .. } | BackendMessage::CopyDone | BackendMessage::EmptyQueryResponse => {}
            BackendMessage::CopyData { data } => response_copy_data.push(data),
            BackendMessage::CopyInResponse { .. } => {
                if next_copy_input >= recorded_copy.inputs.len() {
                    unexpected_copy_from = true;
                    connection.queue(&FrontendMessage::CopyFail {
                        message: "unexpected COPY FROM STDIN request".to_string(),
                    });
                    if connection.send(reader, false).is_err() {
                        return Ok(true);
                    }
                    continue;
                }
                for data in &recorded_copy.inputs[next_copy_input] {
                    connection.queue(&FrontendMessage::CopyData { data: data.clone() });
                }
                next_copy_input += 1;
                connection.queue(&FrontendMessage::CopyDone);
                if let Err(e) = connection.send(reader, false) {
                    tracker.fail(query, &e, "");
                    return Ok(true);
                }
            }
            BackendMessage::CopyOutResponse { .. } => received_copy_out = true,
            BackendMessage::DataRow { values } => response_rows.push(values),
            BackendMessage::ErrorResponse(fields) => response_error = Some(fields),
            BackendMessage::ReadyForQuery { .. } => break,
            BackendMessage::RowDescription { fields } => response_fields = Some(fields),
            other => return Err(unknown(&other)),
        }
    }
    if let Err(e) = connection.empty_receive_buffer() {
        tracker.fail(query, &e, "");
        return Ok(false);
    }
    if unexpected_copy_from {
        tracker.fail(query, "Doltgres requested more COPY FROM STDIN streams than PostgreSQL", "");
        return Ok(false);
    }
    if next_copy_input != recorded_copy.inputs.len() {
        let message =
            format!("expected {} COPY FROM STDIN streams but received {next_copy_input}", recorded_copy.inputs.len());
        tracker.fail(query, &message, "");
        return Ok(false);
    }
    let ordered = query.to_lowercase().contains("order by");
    if received_copy_out && expected_error.is_none() && response_error.is_none() {
        let result = if ordered {
            compare::compare_copy_data_ordered(&recorded_copy.output, &response_copy_data)
        } else {
            compare::compare_copy_data_unordered(&recorded_copy.output, &response_copy_data)
        };
        if let Err(e) = result {
            tracker.fail(query, &e, "");
            return Ok(false);
        }
    }
    if expected_error.is_some() || response_error.is_some() {
        verdict_without_rows(tracker, query, expected_error, response_error);
        return Ok(false);
    }
    let Some(expected_fields) = expected_fields else {
        if response_fields.is_none() {
            tracker.succeed(query);
        } else {
            tracker.fail(query, "expected no rows but received rows", "");
        }
        return Ok(false);
    };
    let Some(response_fields) = response_fields else {
        tracker.fail(query, "expected rows but received none", "");
        return Ok(false);
    };
    if expected_fields.len() != response_fields.len() {
        let message = format!("expected column count {} but received {}", expected_fields.len(), response_fields.len());
        tracker.fail(query, &message, "");
        return Ok(false);
    }
    record_cells(&response_fields, &response_rows);
    let renamed = expected_fields.iter().zip(&response_fields).any(|(e, r)| e.name != r.name);
    if expected_rows.len() != response_rows.len() {
        let message = format!("expected row count {} but received {}", expected_rows.len(), response_rows.len());
        tracker.fail(query, &message, "");
        return Ok(false);
    }
    let result = if ordered {
        compare::compare_rows_ordered(oid_map, &expected_fields, &response_fields, &expected_rows, &response_rows)
    } else {
        compare::compare_rows_unordered(oid_map, &expected_fields, &response_fields, &expected_rows, &response_rows)
    };
    if let Err(e) = result {
        tracker.fail(query, &e, "");
        return Ok(false);
    }
    tracker.succeed(query);
    if renamed {
        tracker.partial_success += 1;
    }
    Ok(false)
}
