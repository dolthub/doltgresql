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

use std::io::{Read, Write};
use std::net::TcpStream;
use std::sync::Arc;

use pgproto::{BackendMessage, ErrorFields, FieldDescription, FrameReader, FrontendMessage, PasswordKind};
use sql::{Outcome, PgError, Session};

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

/// local_timezone returns the name of the machine's time zone, from the /etc/localtime link, or UTC.
fn local_timezone() -> String {
    if let Ok(zone) = std::env::var("TZ")
        && !zone.is_empty()
    {
        return zone.trim_start_matches(':').to_string();
    }
    std::fs::read_link("/etc/localtime")
        .ok()
        .and_then(|path| path.to_str().and_then(|p| p.split_once("zoneinfo/").map(|(_, zone)| zone.to_string())))
        .unwrap_or_else(|| "UTC".to_string())
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
        let mut session = match self.server.engine.session(&user, &host, &database) {
            Ok(session) => session,
            Err(err) => return self.fatal(err),
        };
        self.queue(BackendMessage::AuthenticationOk);
        let date_style = parameter("DateStyle").unwrap_or_else(|| "ISO, MDY".into());
        let timezone = parameter("TimeZone").or_else(|| parameter("timezone")).unwrap_or_else(local_timezone);
        for (name, value) in [
            ("application_name", parameter("application_name").unwrap_or_default()),
            ("client_encoding", "UTF8".to_string()),
            ("DateStyle", date_style),
            ("default_transaction_read_only", "off".into()),
            ("in_hot_standby", "off".into()),
            ("integer_datetimes", "on".into()),
            ("IntervalStyle", "postgres".into()),
            ("is_superuser", "on".into()),
            ("server_encoding", "UTF8".into()),
            ("server_version", "15.17".into()),
            ("session_authorization", user.clone()),
            ("standard_conforming_strings", "on".into()),
            ("TimeZone", timezone),
        ] {
            self.queue(BackendMessage::ParameterStatus { name: name.into(), value });
        }
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
        loop {
            match self.read(PasswordKind::Password)? {
                FrontendMessage::Query { query } => {
                    let (outcomes, error) = session.execute(&query);
                    for outcome in outcomes {
                        self.queue_outcome(outcome);
                    }
                    if let Some(err) = error {
                        self.queue(BackendMessage::ErrorResponse(error_fields(&err)));
                    }
                    self.queue(BackendMessage::ReadyForQuery { tx_status: b'I' });
                    self.flush()?;
                }
                FrontendMessage::Terminate => return Ok(()),
                other => {
                    let err = PgError::unsupported(format!("the {} message", message_name(&other)));
                    self.queue(BackendMessage::ErrorResponse(error_fields(&err)));
                    self.queue(BackendMessage::ReadyForQuery { tx_status: b'I' });
                    self.flush()?;
                }
            }
        }
    }

    /// queue_outcome queues the messages of one statement's outcome.
    fn queue_outcome(&mut self, outcome: Outcome) {
        match outcome {
            Outcome::Rows { columns, rows, tag } => {
                let fields = columns
                    .into_iter()
                    .map(|c| FieldDescription {
                        name: c.name,
                        data_type_oid: c.type_oid,
                        data_type_size: c.type_size,
                        type_modifier: c.type_modifier,
                        ..FieldDescription::default()
                    })
                    .collect();
                self.queue(BackendMessage::RowDescription { fields });
                for values in rows {
                    self.queue(BackendMessage::DataRow { values });
                }
                self.queue(BackendMessage::CommandComplete { command_tag: tag });
            }
            Outcome::Command { tag } => self.queue(BackendMessage::CommandComplete { command_tag: tag }),
            Outcome::Empty => self.queue(BackendMessage::EmptyQueryResponse),
        }
    }
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
