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

//! One client connection: the startup handshake, authentication, and the message loop.

use std::collections::HashMap;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::sync::Arc;

use pgproto::{BackendMessage, ErrorFields, FieldDescription, FrameReader, FrontendMessage, PasswordKind};
use sql::{Column, Outcome, PgError, Prepared, Session, Value, code};

use crate::Server;
use crate::scram::Exchange;

/// PROTOCOL_VERSION is protocol 3.0.
const PROTOCOL_VERSION: u32 = 196608;

/// ConnError ends a connection: a failed read or write, or a client that stopped following the protocol.
#[derive(Debug)]
pub enum ConnError {
    Io(std::io::Error),
    Protocol(String),
}

impl std::fmt::Display for ConnError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ConnError::Io(err) => write!(f, "{err}"),
            ConnError::Protocol(message) => write!(f, "protocol violation: {message}"),
        }
    }
}

impl From<std::io::Error> for ConnError {
    fn from(err: std::io::Error) -> ConnError {
        ConnError::Io(err)
    }
}

/// error_fields converts an engine error to the fields of an ErrorResponse.
fn error_fields(err: &PgError) -> ErrorFields {
    ErrorFields {
        severity: err.severity.to_string(),
        severity_unlocalized: err.severity.to_string(),
        code: err.code.to_string(),
        message: err.message.clone(),
        detail: err.detail.clone().unwrap_or_default(),
        hint: err.hint.clone().unwrap_or_default(),
        position: err.position.map_or(0, |p| p as i32),
        ..ErrorFields::default()
    }
}

/// Conn is a client connection.
pub struct Conn {
    stream: TcpStream,
    frames: FrameReader,
    out: Vec<u8>,
    server: Arc<Server>,
    process_id: u32,
}

impl Conn {
    pub fn new(stream: TcpStream, server: Arc<Server>, process_id: u32) -> Conn {
        Conn { stream, frames: FrameReader::new(), out: Vec::new(), server, process_id }
    }

    /// queue adds a message to the output buffer.
    fn queue(&mut self, message: BackendMessage) {
        message.encode(&mut self.out);
    }

    /// flush sends the output buffer.
    fn flush(&mut self) -> Result<(), ConnError> {
        self.stream.write_all(&self.out)?;
        self.out.clear();
        Ok(())
    }

    /// fill reads more bytes from the client, failing at the end of the stream.
    fn fill(&mut self) -> Result<(), ConnError> {
        let mut buffer = [0; 8192];
        let n = self.stream.read(&mut buffer)?;
        if n == 0 {
            return Err(ConnError::Io(std::io::ErrorKind::UnexpectedEof.into()));
        }
        self.frames.extend(&buffer[..n]);
        Ok(())
    }

    /// read_startup reads an untyped startup-phase message.
    fn read_startup(&mut self) -> Result<FrontendMessage, ConnError> {
        loop {
            match self.frames.next_untyped_frame() {
                Ok(Some(body)) => {
                    return FrontendMessage::decode_startup(&body).map_err(|err| ConnError::Protocol(err.to_string()));
                }
                Ok(None) => self.fill()?,
                Err(err) => return Err(ConnError::Protocol(err.to_string())),
            }
        }
    }

    /// read reads a typed message, decoding a password message as the kind.
    fn read(&mut self, password_kind: PasswordKind) -> Result<FrontendMessage, ConnError> {
        loop {
            match self.frames.next_frame() {
                Ok(Some(frame)) => {
                    return FrontendMessage::decode(frame.tag, &frame.body, password_kind)
                        .map_err(|err| ConnError::Protocol(err.to_string()));
                }
                Ok(None) => self.fill()?,
                Err(err) => return Err(ConnError::Protocol(err.to_string())),
            }
        }
    }

    /// fatal sends a FATAL error, which ends the connection.
    fn fatal(&mut self, err: PgError) -> Result<(), ConnError> {
        self.queue(BackendMessage::ErrorResponse(error_fields(&PgError { severity: "FATAL", ..err })));
        self.flush()
    }

    /// run serves the connection until the client leaves.
    pub fn run(mut self) -> Result<(), ConnError> {
        let parameters = loop {
            match self.read_startup()? {
                FrontendMessage::SSLRequest | FrontendMessage::GSSEncRequest => {
                    self.stream.write_all(b"N")?;
                }
                FrontendMessage::CancelRequest { .. } => return Ok(()),
                FrontendMessage::StartupMessage { protocol_version, parameters } => {
                    if protocol_version >> 16 != PROTOCOL_VERSION >> 16 {
                        return self.fatal(PgError::new(
                            "0A000",
                            format!(
                                "unsupported frontend protocol {}.{}: server supports 3.0 to 3.0",
                                protocol_version >> 16,
                                protocol_version & 0xffff
                            ),
                        ));
                    }
                    break parameters;
                }
                other => return Err(ConnError::Protocol(format!("unexpected startup message {other:?}"))),
            }
        };
        let parameter = |name: &str| parameters.iter().find(|(n, _)| n == name).map(|(_, v)| v.clone());
        let user = parameter("user").unwrap_or_default();
        if !self.authenticate(&user)? {
            return Ok(());
        }
        let database = parameter("database").filter(|d| !d.is_empty()).unwrap_or_else(|| user.clone());
        let host = self.stream.peer_addr().map(|addr| addr.ip().to_string()).unwrap_or_default();
        let mut startup: Vec<(String, String)> = parameters
            .iter()
            .filter(|(name, _)| !matches!(name.as_str(), "user" | "database" | "options" | "replication"))
            .cloned()
            .collect();
        if !startup.iter().any(|(name, _)| name.eq_ignore_ascii_case("DateStyle")) {
            startup.push(("DateStyle".into(), "ISO, MDY".into()));
        }
        let mut session = match self.server.engine.session(&user, &host, &database, &startup) {
            Ok(session) => session,
            Err(err) => return self.fatal(err),
        };
        self.queue(BackendMessage::AuthenticationOk);
        self.queue_parameters(&mut session);
        self.queue(BackendMessage::BackendKeyData { process_id: self.process_id, secret_key: vec![0; 4] });
        self.queue(BackendMessage::ReadyForQuery { tx_status: b'I' });
        self.flush()?;
        self.serve(&mut session)
    }

    /// authenticate runs a SCRAM-SHA-256 exchange, reporting whether the client proved it knows the password.
    fn authenticate(&mut self, user: &str) -> Result<bool, ConnError> {
        let failed = PgError::fatal("28P01", format!("password authentication failed for user \"{user}\""));
        let Some(verifier) = self.server.verifier(user) else {
            self.fatal(failed)?;
            return Ok(false);
        };
        self.queue(BackendMessage::AuthenticationSASL { mechanisms: vec!["SCRAM-SHA-256".into()] });
        self.flush()?;
        let FrontendMessage::SASLInitialResponse { auth_mechanism, data } =
            self.read(PasswordKind::SASLInitialResponse)?
        else {
            return Err(ConnError::Protocol("expected a SASLInitialResponse".into()));
        };
        if auth_mechanism != "SCRAM-SHA-256" {
            self.fatal(PgError::fatal("08P01", "invalid SASL authentication mechanism"))?;
            return Ok(false);
        }
        let Some((exchange, server_first)) = Exchange::start(&data.unwrap_or_default(), &verifier) else {
            self.fatal(PgError::fatal("08P01", "malformed SCRAM message"))?;
            return Ok(false);
        };
        self.queue(BackendMessage::AuthenticationSASLContinue { data: server_first.into_bytes() });
        self.flush()?;
        let FrontendMessage::SASLResponse { data } = self.read(PasswordKind::SASLResponse)? else {
            return Err(ConnError::Protocol("expected a SASLResponse".into()));
        };
        match exchange.finish(&data, &verifier) {
            Some(server_final) => {
                self.queue(BackendMessage::AuthenticationSASLFinal { data: server_final.into_bytes() });
                Ok(true)
            }
            None => {
                self.fatal(failed)?;
                Ok(false)
            }
        }
    }

    /// serve runs the message loop.
    fn serve(&mut self, session: &mut Session) -> Result<(), ConnError> {
        let mut extended = Extended::default();
        loop {
            let message = self.read(PasswordKind::Password)?;
            if extended.failed && !matches!(message, FrontendMessage::Sync | FrontendMessage::Terminate) {
                continue;
            }
            match message {
                FrontendMessage::Query { query } => {
                    extended.statements.remove("");
                    extended.portals.remove("");
                    let (outcomes, error) = session.execute(&query);
                    self.queue_notices(session);
                    for outcome in outcomes {
                        self.queue_outcome(outcome, None);
                    }
                    if let Some(err) = error {
                        self.queue(BackendMessage::ErrorResponse(error_fields(&err)));
                    }
                    self.queue_parameters(session);
                    self.queue(BackendMessage::ReadyForQuery { tx_status: session.tx_status() });
                    self.flush()?;
                }
                FrontendMessage::Sync => {
                    extended.failed = false;
                    self.queue_parameters(session);
                    self.queue(BackendMessage::ReadyForQuery { tx_status: session.tx_status() });
                    self.flush()?;
                }
                FrontendMessage::Flush => self.flush()?,
                FrontendMessage::Terminate => return Ok(()),
                message => {
                    let result = self.extended_message(session, &mut extended, message);
                    self.queue_notices(session);
                    if let Err(err) = result {
                        self.queue(BackendMessage::ErrorResponse(error_fields(&err)));
                        extended.failed = true;
                    }
                }
            }
        }
    }

    /// extended_message handles a message of the extended query protocol other than Sync and Flush.
    fn extended_message(
        &mut self,
        session: &mut Session,
        extended: &mut Extended,
        message: FrontendMessage,
    ) -> Result<(), PgError> {
        match message {
            FrontendMessage::Parse { name, query, parameter_oids } => {
                if !name.is_empty() && extended.statements.contains_key(&name) {
                    return Err(PgError::new(
                        code::DUPLICATE_PREPARED_STATEMENT,
                        format!("prepared statement \"{name}\" already exists"),
                    ));
                }
                let prepared = session.prepare(&query, &parameter_oids)?;
                extended.statements.insert(name, Arc::new(prepared));
                self.queue(BackendMessage::ParseComplete);
            }
            FrontendMessage::Bind {
                destination_portal,
                prepared_statement,
                parameter_format_codes,
                parameters,
                result_format_codes,
            } => {
                let prepared = extended.statement(&prepared_statement)?;
                if parameters.len() != prepared.parameter_types.len() {
                    return Err(PgError::new(
                        code::PROTOCOL_VIOLATION,
                        format!(
                            "bind message supplies {} parameters, but prepared statement \"{prepared_statement}\" \
                             requires {}",
                            parameters.len(),
                            prepared.parameter_types.len()
                        ),
                    ));
                }
                let values = parameters
                    .iter()
                    .enumerate()
                    .map(|(i, value)| {
                        Value::decode(prepared.parameter_types[i], format(&parameter_format_codes, i), value.as_deref())
                    })
                    .collect::<sql::Result<Vec<Value>>>()?;
                let portal = Portal { prepared, parameters: values, result_formats: result_format_codes };
                extended.portals.insert(destination_portal, portal);
                self.queue(BackendMessage::BindComplete);
            }
            FrontendMessage::Describe { object_type: b'S', name } => {
                let prepared = extended.statement(&name)?;
                self.queue(BackendMessage::ParameterDescription { parameter_oids: prepared.parameter_types.clone() });
                self.queue_description(prepared.columns.as_deref(), &[]);
            }
            FrontendMessage::Describe { name, .. } => {
                let portal = extended.portal(&name)?;
                let (prepared, formats) = (portal.prepared.clone(), portal.result_formats.clone());
                self.queue_description(prepared.columns.as_deref(), &formats);
            }
            FrontendMessage::Execute { portal, .. } => {
                let portal = extended.portal(&portal)?;
                let (prepared, formats) = (portal.prepared.clone(), portal.result_formats.clone());
                let outcome = session.execute_prepared(&prepared, &portal.parameters.clone())?;
                self.queue_outcome(outcome, Some(&formats));
            }
            FrontendMessage::Close { object_type, name } => {
                if object_type == b'S' {
                    extended.statements.remove(&name);
                } else {
                    extended.portals.remove(&name);
                }
                self.queue(BackendMessage::CloseComplete);
            }
            other => return Err(PgError::unsupported(format!("the {} message", message_name(&other)))),
        }
        Ok(())
    }

    /// queue_parameters queues a ParameterStatus for each reported parameter that changed.
    fn queue_parameters(&mut self, session: &mut Session) {
        for (name, value) in session.parameter_changes() {
            self.queue(BackendMessage::ParameterStatus { name, value });
        }
    }

    /// queue_notices queues the notices the session raised.
    fn queue_notices(&mut self, session: &mut Session) {
        for notice in session.take_notices() {
            self.queue(BackendMessage::NoticeResponse(error_fields(&notice)));
        }
    }

    /// queue_description queues the RowDescription of the columns in the formats, or NoData without columns.
    fn queue_description(&mut self, columns: Option<&[Column]>, formats: &[i16]) {
        match columns {
            Some(columns) => self.queue(BackendMessage::RowDescription { fields: fields(columns, formats) }),
            None => self.queue(BackendMessage::NoData),
        }
    }

    /// queue_outcome queues the messages of one statement's outcome, with rows in the formats. Without formats, as
    /// for a simple query, it describes the rows first and sends them as text.
    fn queue_outcome(&mut self, outcome: Outcome, formats: Option<&[i16]>) {
        match outcome {
            Outcome::Rows { columns, rows, tag } => {
                let formats = match formats {
                    Some(formats) => formats,
                    None => {
                        self.queue(BackendMessage::RowDescription { fields: fields(&columns, &[]) });
                        &[]
                    }
                };
                for row in rows {
                    let values = row
                        .iter()
                        .enumerate()
                        .map(|(i, value)| value.encode(columns[i].type_oid, format(formats, i)))
                        .collect();
                    self.queue(BackendMessage::DataRow { values });
                }
                self.queue(BackendMessage::CommandComplete { command_tag: tag });
            }
            Outcome::Command { tag } => self.queue(BackendMessage::CommandComplete { command_tag: tag }),
            Outcome::Empty => self.queue(BackendMessage::EmptyQueryResponse),
        }
    }
}

/// Extended is the state of the extended query protocol: the prepared statements, the portals, and whether an error
/// is discarding messages until the next Sync.
#[derive(Default)]
struct Extended {
    statements: HashMap<String, Arc<Prepared>>,
    portals: HashMap<String, Portal>,
    failed: bool,
}

impl Extended {
    /// statement returns a prepared statement by name.
    fn statement(&self, name: &str) -> Result<Arc<Prepared>, PgError> {
        self.statements.get(name).cloned().ok_or_else(|| {
            PgError::new(code::INVALID_SQL_STATEMENT_NAME, format!("prepared statement \"{name}\" does not exist"))
        })
    }

    /// portal returns a portal by name.
    fn portal(&self, name: &str) -> Result<&Portal, PgError> {
        self.portals
            .get(name)
            .ok_or_else(|| PgError::new(code::INVALID_CURSOR_NAME, format!("portal \"{name}\" does not exist")))
    }
}

/// Portal is a prepared statement bound to parameter values and result formats.
struct Portal {
    prepared: Arc<Prepared>,
    parameters: Vec<Value>,
    result_formats: Vec<i16>,
}

/// format returns the format code of the value at the index: the only code when there is one, and text when there
/// are none.
fn format(codes: &[i16], index: usize) -> i16 {
    match codes {
        [] => 0,
        [code] => *code,
        codes => codes.get(index).copied().unwrap_or(0),
    }
}

/// fields describes the columns, with the result formats.
fn fields(columns: &[Column], formats: &[i16]) -> Vec<FieldDescription> {
    columns
        .iter()
        .enumerate()
        .map(|(i, c)| FieldDescription {
            name: c.name.clone(),
            data_type_oid: c.type_oid,
            data_type_size: c.type_size,
            type_modifier: c.type_modifier,
            format: format(formats, i),
            ..FieldDescription::default()
        })
        .collect()
}

/// message_name names a frontend message for errors.
fn message_name(message: &FrontendMessage) -> &'static str {
    match message {
        FrontendMessage::Parse { .. } => "Parse",
        FrontendMessage::Bind { .. } => "Bind",
        FrontendMessage::Describe { .. } => "Describe",
        FrontendMessage::Execute { .. } => "Execute",
        FrontendMessage::Sync => "Sync",
        FrontendMessage::Close { .. } => "Close",
        FrontendMessage::Flush => "Flush",
        _ => "client",
    }
}
