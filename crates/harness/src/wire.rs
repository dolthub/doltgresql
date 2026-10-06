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

//! Wire tests: raw protocol conversations whose every backend message is compared against what Postgres sends.
//! A conversation alternates between sending frontend messages and receiving an exact list of backend messages, and
//! may check committed state through a separate connection.

use std::io::{Read, Write};
use std::net::TcpStream;
use std::time::Duration;

use pgproto::{BackendMessage, ErrorFields, FieldDescription, FrameReader, FrontendMessage, PROTOCOL_VERSION_3};

use crate::script::{Cell, Session};
use crate::server::Target;

/// How long a receive waits for a message before deciding that the server has nothing more to send.
const QUIET_TIMEOUT: Duration = Duration::from_millis(500);
/// The error a receive returns when the server closes the connection.
const CONNECTION_CLOSED: &str = "connection closed";
/// How long a receive waits for a message that a test expects.
const RECEIVE_TIMEOUT: Duration = Duration::from_secs(30);

/// WireTest is a protocol conversation against its own fresh server.
#[derive(Clone, Copy, Debug)]
pub struct WireTest {
    /// The name of the test.
    pub name: &'static str,
    /// Statements that run before the conversation over a normal connection, which must succeed.
    pub set_up_script: &'static [&'static str],
    /// The conversation.
    pub steps: &'static [Step],
    /// When any test in a run sets this, only those tests run. It must never be committed.
    pub focus: bool,
    /// Skips the test for the given reason.
    pub skip: Option<&'static str>,
    /// The StartupMessage parameters of the conversation's connection.
    pub startup: &'static [(&'static str, &'static str)],
}

/// PG_REGRESS_STARTUP are the startup parameters of the Go suite's raw wire connection, which mimics pg_regress.
pub const PG_REGRESS_STARTUP: &[(&str, &str)] = &[
    ("timezone", "PST8PDT"),
    ("user", "postgres"),
    ("database", "postgres"),
    ("options", " -c intervalstyle=postgres_verbose"),
    ("application_name", "pg_regress"),
    ("client_encoding", "WIN1252"),
    ("datestyle", "Postgres, MDY"),
];

/// PGX_STARTUP are the startup parameters of the Go suite's pgx connections.
pub const PGX_STARTUP: &[(&str, &str)] = &[("DateStyle", "ISO, MDY"), ("database", "postgres"), ("user", "postgres")];

/// W is a WireTest with every field at its default, for use with struct update syntax.
pub const W: WireTest =
    WireTest { name: "", set_up_script: &[], steps: &[], focus: false, skip: None, startup: PG_REGRESS_STARTUP };

/// Step is one step of a conversation.
#[derive(Clone, Copy, Debug)]
pub enum Step {
    /// Sends these messages.
    Send(&'static [Send]),
    /// Receives exactly these messages, in order. Table OIDs and user-defined type OIDs in row descriptions are
    /// arbitrary, so they are compared as zero, and Postgres' source locations in errors and notices are ignored.
    Receive(&'static [Receive]),
    /// Runs a query on a separate normal connection and expects these rows in any order, which shows what the
    /// conversation has committed.
    OtherQuery {
        /// The query.
        query: &'static str,
        /// The expected rows.
        rows: &'static [&'static [Cell]],
    },
}

/// Datum is a value in a message: NULL, text, or bytes that are not valid UTF-8.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Datum {
    /// The SQL NULL.
    Null,
    /// Text.
    Text(&'static str),
    /// Bytes that are not valid UTF-8.
    Bytes(&'static [u8]),
}

impl Datum {
    /// to_bytes converts the value to its bytes, where None is NULL.
    fn to_bytes(self) -> Option<Vec<u8>> {
        match self {
            Datum::Null => None,
            Datum::Text(text) => Some(text.as_bytes().to_vec()),
            Datum::Bytes(bytes) => Some(bytes.to_vec()),
        }
    }
}

/// Send is a frontend message.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Send {
    /// A simple Query.
    Query(&'static str),
    /// Parse.
    Parse {
        /// The statement name.
        name: &'static str,
        /// The query.
        query: &'static str,
        /// The parameter types.
        parameter_oids: &'static [u32],
    },
    /// Bind.
    Bind {
        /// The portal name.
        portal: &'static str,
        /// The statement name.
        statement: &'static str,
        /// The parameter formats.
        parameter_formats: &'static [i16],
        /// The parameter values.
        parameters: &'static [Datum],
        /// The result formats.
        result_formats: &'static [i16],
    },
    /// Describe a statement ('S') or portal ('P').
    Describe(u8, &'static str),
    /// Execute a portal with a row limit, where zero means no limit.
    Execute(&'static str, u32),
    /// Close a statement ('S') or portal ('P').
    Close(u8, &'static str),
    /// Sync.
    Sync,
    /// Flush.
    Flush,
    /// CopyData.
    CopyData(&'static [u8]),
    /// CopyDone.
    CopyDone,
    /// CopyFail with a reason.
    CopyFail(&'static str),
}

impl Send {
    /// to_message converts to a protocol message.
    pub fn to_message(self) -> FrontendMessage {
        match self {
            Send::Query(query) => FrontendMessage::Query { query: query.to_string() },
            Send::Parse { name, query, parameter_oids } => FrontendMessage::Parse {
                name: name.to_string(),
                query: query.to_string(),
                parameter_oids: parameter_oids.to_vec(),
            },
            Send::Bind { portal, statement, parameter_formats, parameters, result_formats } => FrontendMessage::Bind {
                destination_portal: portal.to_string(),
                prepared_statement: statement.to_string(),
                parameter_format_codes: parameter_formats.to_vec(),
                parameters: parameters.iter().map(|d| d.to_bytes()).collect(),
                result_format_codes: result_formats.to_vec(),
            },
            Send::Describe(object_type, name) => FrontendMessage::Describe { object_type, name: name.to_string() },
            Send::Execute(portal, max_rows) => FrontendMessage::Execute { portal: portal.to_string(), max_rows },
            Send::Close(object_type, name) => FrontendMessage::Close { object_type, name: name.to_string() },
            Send::Sync => FrontendMessage::Sync,
            Send::Flush => FrontendMessage::Flush,
            Send::CopyData(data) => FrontendMessage::CopyData { data: data.to_vec() },
            Send::CopyDone => FrontendMessage::CopyDone,
            Send::CopyFail(message) => FrontendMessage::CopyFail { message: message.to_string() },
        }
    }
}

/// Field is a column of a RowDescription.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Field {
    /// The column name.
    pub name: &'static str,
    /// The attribute number of the source column.
    pub attnum: u16,
    /// The type OID, which is zero for user-defined types.
    pub type_oid: u32,
    /// The type size.
    pub size: i16,
    /// The type modifier.
    pub typmod: i32,
    /// The format.
    pub format: i16,
}

/// Fields is an ErrorResponse or NoticeResponse, without Postgres' source locations. Empty strings and zero
/// positions are absent fields.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Fields {
    /// The severity.
    pub severity: &'static str,
    /// The unlocalized severity.
    pub severity_unlocalized: &'static str,
    /// The SQLSTATE code.
    pub code: &'static str,
    /// The message.
    pub message: &'static str,
    /// The detail.
    pub detail: &'static str,
    /// The hint.
    pub hint: &'static str,
    /// The position.
    pub position: i32,
    /// The internal position.
    pub internal_position: i32,
    /// The internal query.
    pub internal_query: &'static str,
    /// The context.
    pub where_: &'static str,
    /// The schema name.
    pub schema: &'static str,
    /// The table name.
    pub table: &'static str,
    /// The column name.
    pub column: &'static str,
    /// The data type name.
    pub data_type: &'static str,
    /// The constraint name.
    pub constraint: &'static str,
}

/// F is a Fields with every field empty, for use with struct update syntax.
pub const F: Fields = Fields {
    severity: "",
    severity_unlocalized: "",
    code: "",
    message: "",
    detail: "",
    hint: "",
    position: 0,
    internal_position: 0,
    internal_query: "",
    where_: "",
    schema: "",
    table: "",
    column: "",
    data_type: "",
    constraint: "",
};

/// Receive is a backend message.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Receive {
    /// ParseComplete.
    ParseComplete,
    /// BindComplete.
    BindComplete,
    /// CloseComplete.
    CloseComplete,
    /// NoData.
    NoData,
    /// EmptyQueryResponse.
    EmptyQueryResponse,
    /// PortalSuspended.
    PortalSuspended,
    /// CopyDone.
    CopyDone,
    /// ParameterDescription.
    ParameterDescription(&'static [u32]),
    /// RowDescription.
    RowDescription(&'static [Field]),
    /// DataRow.
    DataRow(&'static [Datum]),
    /// CommandComplete.
    CommandComplete(&'static str),
    /// ReadyForQuery with its transaction status.
    ReadyForQuery(u8),
    /// ErrorResponse.
    Error(Fields),
    /// NoticeResponse.
    Notice(Fields),
    /// ParameterStatus.
    ParameterStatus(&'static str, &'static str),
    /// CopyInResponse.
    CopyInResponse(u8, &'static [u16]),
    /// CopyOutResponse.
    CopyOutResponse(u8, &'static [u16]),
    /// CopyData.
    CopyData(&'static [u8]),
}

/// normalize returns a message with its arbitrary parts zeroed, for comparison.
pub fn normalize(message: &BackendMessage) -> BackendMessage {
    let mut message = message.clone();
    match &mut message {
        BackendMessage::RowDescription { fields } => {
            for field in fields {
                field.table_oid = 0;
                if field.data_type_oid > 16383 {
                    field.data_type_oid = 0;
                }
            }
        }
        BackendMessage::ErrorResponse(fields) | BackendMessage::NoticeResponse(fields) => {
            fields.file.clear();
            fields.line = 0;
            fields.routine.clear();
        }
        _ => {}
    }
    message
}

impl Receive {
    /// to_message converts to the normalized protocol message.
    pub fn to_message(self) -> BackendMessage {
        let fields = |f: Fields| ErrorFields {
            severity: f.severity.into(),
            severity_unlocalized: f.severity_unlocalized.into(),
            code: f.code.into(),
            message: f.message.into(),
            detail: f.detail.into(),
            hint: f.hint.into(),
            position: f.position,
            internal_position: f.internal_position,
            internal_query: f.internal_query.into(),
            where_: f.where_.into(),
            schema_name: f.schema.into(),
            table_name: f.table.into(),
            column_name: f.column.into(),
            data_type_name: f.data_type.into(),
            constraint_name: f.constraint.into(),
            ..ErrorFields::default()
        };
        match self {
            Receive::ParseComplete => BackendMessage::ParseComplete,
            Receive::BindComplete => BackendMessage::BindComplete,
            Receive::CloseComplete => BackendMessage::CloseComplete,
            Receive::NoData => BackendMessage::NoData,
            Receive::EmptyQueryResponse => BackendMessage::EmptyQueryResponse,
            Receive::PortalSuspended => BackendMessage::PortalSuspended,
            Receive::CopyDone => BackendMessage::CopyDone,
            Receive::ParameterDescription(oids) => {
                BackendMessage::ParameterDescription { parameter_oids: oids.to_vec() }
            }
            Receive::RowDescription(fields) => BackendMessage::RowDescription {
                fields: fields
                    .iter()
                    .map(|f| FieldDescription {
                        name: f.name.into(),
                        table_oid: 0,
                        table_attribute_number: f.attnum,
                        data_type_oid: f.type_oid,
                        data_type_size: f.size,
                        type_modifier: f.typmod,
                        format: f.format,
                    })
                    .collect(),
            },
            Receive::DataRow(values) => {
                BackendMessage::DataRow { values: values.iter().map(|d| d.to_bytes()).collect() }
            }
            Receive::CommandComplete(tag) => BackendMessage::CommandComplete { command_tag: tag.into() },
            Receive::ReadyForQuery(status) => BackendMessage::ReadyForQuery { tx_status: status },
            Receive::Error(f) => BackendMessage::ErrorResponse(fields(f)),
            Receive::Notice(f) => BackendMessage::NoticeResponse(fields(f)),
            Receive::ParameterStatus(name, value) => {
                BackendMessage::ParameterStatus { name: name.into(), value: value.into() }
            }
            Receive::CopyInResponse(format, columns) => {
                BackendMessage::CopyInResponse { overall_format: format, column_format_codes: columns.to_vec() }
            }
            Receive::CopyOutResponse(format, columns) => {
                BackendMessage::CopyOutResponse { overall_format: format, column_format_codes: columns.to_vec() }
            }
            Receive::CopyData(data) => BackendMessage::CopyData { data: data.to_vec() },
        }
    }
}

/// RawConnection is a protocol connection for wire tests.
pub struct RawConnection {
    socket: TcpStream,
    reader: FrameReader,
}

impl RawConnection {
    /// connect connects with the startup parameters, authenticating as postgres, and discards the startup messages.
    pub fn connect(port: u16, parameters: &[(&str, &str)]) -> Result<RawConnection, String> {
        let socket = TcpStream::connect(("127.0.0.1", port)).map_err(|err| err.to_string())?;
        socket.set_nodelay(true).map_err(|err| err.to_string())?;
        let mut conn = RawConnection { socket, reader: FrameReader::new() };
        conn.send(&[FrontendMessage::StartupMessage {
            protocol_version: PROTOCOL_VERSION_3 as u32,
            parameters: parameters.iter().map(|(k, v)| (k.to_string(), v.to_string())).collect(),
        }])?;
        loop {
            match conn.receive(RECEIVE_TIMEOUT)?.ok_or("timed out during startup")? {
                BackendMessage::AuthenticationOk | BackendMessage::BackendKeyData { .. } => {}
                BackendMessage::ParameterStatus { .. } => {}
                BackendMessage::ReadyForQuery { .. } => return Ok(conn),
                BackendMessage::AuthenticationSASL { mechanisms } => {
                    if !mechanisms.iter().any(|m| m == "SCRAM-SHA-256") {
                        return Err(format!("no suitable SASL mechanism in {mechanisms:?}"));
                    }
                    conn.scram("postgres", "password")?;
                }
                BackendMessage::ErrorResponse(fields) => return Err(fields.message),
                other => return Err(format!("unexpected startup message {other:?}")),
            }
        }
    }

    /// scram authenticates with SCRAM-SHA-256, sending the user name in the client-first-message as the Go suite's
    /// SCRAM library does.
    fn scram(&mut self, user: &str, password: &str) -> Result<(), String> {
        let mut client = crate::pgx::ScramClient::with_user(password, user);
        self.send(&[FrontendMessage::SASLInitialResponse {
            auth_mechanism: "SCRAM-SHA-256".into(),
            data: Some(client.client_first_message()),
        }])?;
        let server_first = match self.receive(RECEIVE_TIMEOUT)? {
            Some(BackendMessage::AuthenticationSASLContinue { data }) => data,
            other => return Err(format!("expected AuthenticationSASLContinue, got {other:?}")),
        };
        let client_final = client.client_final_message(&server_first)?;
        self.send(&[FrontendMessage::SASLResponse { data: client_final }])?;
        match self.receive(RECEIVE_TIMEOUT)? {
            Some(BackendMessage::AuthenticationSASLFinal { data }) => client.verify_server_final_message(&data),
            other => Err(format!("expected AuthenticationSASLFinal, got {other:?}")),
        }
    }

    /// send writes messages in one batch.
    pub fn send(&mut self, messages: &[FrontendMessage]) -> Result<(), String> {
        let mut buffer = Vec::new();
        for message in messages {
            message.encode(&mut buffer);
        }
        self.socket.write_all(&buffer).map_err(|err| err.to_string())
    }

    /// receive returns the next message, or None when none arrives within the timeout.
    pub fn receive(&mut self, timeout: Duration) -> Result<Option<BackendMessage>, String> {
        loop {
            if let Some(frame) = self.reader.next_frame().map_err(|err| err.to_string())? {
                return BackendMessage::decode(frame.tag, &frame.body).map(Some).map_err(|err| err.to_string());
            }
            self.socket.set_read_timeout(Some(timeout)).map_err(|err| err.to_string())?;
            let mut buffer = [0u8; 16384];
            match self.socket.read(&mut buffer) {
                Ok(0) => return Err(CONNECTION_CLOSED.to_string()),
                Ok(count) => self.reader.extend(&buffer[..count]),
                Err(err) if matches!(err.kind(), std::io::ErrorKind::WouldBlock | std::io::ErrorKind::TimedOut) => {
                    return Ok(None);
                }
                Err(err) => return Err(err.to_string()),
            }
        }
    }

    /// receive_until reads messages until the predicate accepts one, which is included, or until the server goes
    /// quiet when no predicate is given.
    pub fn receive_until(
        &mut self,
        mut stop: Option<&mut dyn FnMut(&BackendMessage) -> bool>,
    ) -> Result<Vec<BackendMessage>, String> {
        let mut messages = Vec::new();
        loop {
            let timeout = if stop.is_some() { RECEIVE_TIMEOUT } else { QUIET_TIMEOUT };
            let received = match self.receive(timeout) {
                Err(err) if err == CONNECTION_CLOSED => return Ok(messages),
                received => received?,
            };
            match received {
                Some(message) => {
                    let done = stop.as_mut().is_some_and(|stop| stop(&message));
                    messages.push(message);
                    if done {
                        return Ok(messages);
                    }
                }
                None if stop.is_none() => return Ok(messages),
                None => return Err(format!("timed out after receiving {messages:?}")),
            }
        }
    }
}

/// expected_ready_count returns how many ReadyForQuery messages a batch of messages draws: one per Query and Sync.
/// Within COPY FROM STDIN the server ignores Sync and Flush, so a Sync only counts after the copy ends.
pub fn expected_ready_count(messages: &[FrontendMessage], in_copy: &mut bool) -> usize {
    let mut count = 0;
    for message in messages {
        match message {
            FrontendMessage::CopyDone | FrontendMessage::CopyFail { .. } => *in_copy = false,
            FrontendMessage::Query { .. } => count += 1,
            FrontendMessage::Sync if !*in_copy => count += 1,
            _ => {}
        }
    }
    count
}

/// WireCapture is everything a conversation received.
#[derive(Clone, Debug, Default)]
pub struct WireCapture {
    /// The error that stopped the conversation before it finished.
    pub error: Option<String>,
    /// The messages received at each Receive step, by step index.
    pub received: Vec<(usize, Vec<BackendMessage>)>,
    /// The rows of each OtherQuery step, by step index.
    pub other_rows: Vec<(usize, Vec<Vec<Option<String>>>)>,
}

/// run_steps runs a conversation. When capturing, Receive steps record what arrives instead of checking it: they
/// read through the ReadyForQuery messages the preceding sends draw, through a CopyInResponse when the next step
/// sends copy data, and otherwise until the server goes quiet.
fn run_steps(
    session: &mut Session,
    conn: &mut RawConnection,
    test: &WireTest,
    capture: bool,
) -> (WireCapture, Vec<String>) {
    let mut result = WireCapture::default();
    let mut failures = Vec::new();
    let mut pending_ready = 0;
    let mut in_copy = false;
    for (index, step) in test.steps.iter().enumerate() {
        match step {
            Step::Send(messages) => {
                let messages: Vec<FrontendMessage> = messages.iter().map(|m| m.to_message()).collect();
                pending_ready += expected_ready_count(&messages, &mut in_copy);
                if let Err(err) = conn.send(&messages) {
                    result.error = Some(format!("step {index}: {err}"));
                    return (result, failures);
                }
            }
            Step::Receive(expected) => {
                let received = if capture {
                    let next_is_copy = matches!(
                        test.steps.get(index + 1),
                        Some(Step::Send(next)) if next.iter().any(|m| matches!(m, Send::CopyData(_) | Send::CopyDone | Send::CopyFail(_)))
                    );
                    let mut seen_ready = 0;
                    let target = pending_ready;
                    let mut stop = |message: &BackendMessage| match message {
                        BackendMessage::ReadyForQuery { .. } => {
                            seen_ready += 1;
                            seen_ready >= target
                        }
                        BackendMessage::CopyInResponse { .. } => next_is_copy,
                        _ => false,
                    };
                    let outcome = if target > 0 || next_is_copy {
                        conn.receive_until(Some(&mut stop))
                    } else {
                        conn.receive_until(None)
                    };
                    match outcome {
                        Ok(messages) => {
                            pending_ready -= messages
                                .iter()
                                .filter(|m| matches!(m, BackendMessage::ReadyForQuery { .. }))
                                .count()
                                .min(pending_ready);
                            messages
                        }
                        Err(err) => {
                            result.error = Some(format!("step {index}: {err}"));
                            return (result, failures);
                        }
                    }
                } else {
                    let mut messages = Vec::new();
                    for _ in 0..expected.len() {
                        match conn.receive(RECEIVE_TIMEOUT) {
                            Ok(Some(message)) => messages.push(message),
                            Ok(None) => break,
                            Err(err) => {
                                failures.push(format!("step {index}: {err}"));
                                break;
                            }
                        }
                    }
                    pending_ready = 0;
                    messages
                };
                if !capture {
                    let expected: Vec<BackendMessage> = expected.iter().map(|m| m.to_message()).collect();
                    let actual: Vec<BackendMessage> = received.iter().map(normalize).collect();
                    if expected != actual {
                        failures.push(format!(
                            "step {index}: messages differ\n  expected: {expected:#?}\n  actual: {actual:#?}"
                        ));
                    }
                }
                if received.iter().any(|m| matches!(m, BackendMessage::CopyInResponse { .. })) {
                    in_copy = true;
                }
                result.received.push((index, received));
            }
            Step::OtherQuery { query, rows } => {
                let observation = session.run(&crate::script::ScriptTestAssertion {
                    query,
                    expected: crate::script::Expected::Rows { columns: &[], rows: &[], tag: "" },
                    ..crate::script::A
                });
                if !capture {
                    let problems = crate::script::check_other_rows(rows, &observation);
                    failures.extend(problems.into_iter().map(|p| format!("step {index}: {query}: {p}")));
                }
                result.other_rows.push((index, observation.rows));
            }
        }
    }
    if !capture && let Ok(Some(extra)) = conn.receive(QUIET_TIMEOUT) {
        failures.push(format!("the server sent an unexpected message after the conversation: {extra:?}"));
    }
    (result, failures)
}

/// start starts a server, runs the setup, and opens the raw connection.
fn start(target: &Target, test: &WireTest) -> Result<(Session, RawConnection), String> {
    let mut session = Session::start(target, "", "")?;
    for query in test.set_up_script {
        session.set_up(query)?;
    }
    let conn = RawConnection::connect(session.server.port, test.startup)?;
    Ok((session, conn))
}

/// capture_wire_test runs a conversation against a fresh server and records what it received.
pub fn capture_wire_test(target: &Target, test: &WireTest) -> WireCapture {
    match start(target, test) {
        Ok((mut session, mut conn)) => run_steps(&mut session, &mut conn, test, true).0,
        Err(err) => WireCapture { error: Some(err), ..WireCapture::default() },
    }
}

/// run_wire_tests runs conversations against the target in DOLTGRES_TEST_TARGET, panicking with every failure.
pub fn run_wire_tests(tests: &[WireTest]) {
    let target = Target::from_env().unwrap_or_else(|err| panic!("{err}"));
    let focus: Vec<&WireTest> = tests.iter().filter(|t| t.focus).collect();
    if !focus.is_empty() && std::env::var_os("GITHUB_ACTION").is_some() {
        panic!("the wire test {:?} has focus set, which CI does not allow", focus[0].name);
    }
    let selected: Vec<&WireTest> = if focus.is_empty() { tests.iter().collect() } else { focus };
    let mut failures = Vec::new();
    for test in selected {
        if test.skip.is_some() {
            continue;
        }
        match start(&target, test) {
            Ok((mut session, mut conn)) => {
                let (result, problems) = run_steps(&mut session, &mut conn, test, false);
                if let Some(err) = result.error {
                    crate::script::record_failure(test.name, "wire", "", &err);
                    failures.push(format!("{}: {err}", test.name));
                }
                for problem in problems {
                    crate::script::record_failure(test.name, "wire", "", &problem);
                    failures.push(format!("{}: {problem}", test.name));
                }
            }
            Err(err) => {
                crate::script::record_failure(test.name, "wire", "", &err);
                failures.push(format!("{}: {err}", test.name));
            }
        }
    }
    if !failures.is_empty() {
        panic!("{} failures:\n\n{}", failures.len(), failures.join("\n\n"));
    }
}
