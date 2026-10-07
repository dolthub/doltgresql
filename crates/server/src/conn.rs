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

use pgproto::{BackendMessage, ErrorFields, FieldDescription, Frame, FrameReader, FrontendMessage, PasswordKind};
use sql::{Column, Outcome, PgError, Prepared, Results, Session, Value, code};

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
    let objects = err.objects.as_deref().cloned().unwrap_or_default();
    ErrorFields {
        severity: err.severity.to_string(),
        severity_unlocalized: err.severity.to_string(),
        code: err.code.to_string(),
        message: err.message.clone(),
        detail: err.detail.clone().unwrap_or_default(),
        hint: err.hint.clone().unwrap_or_default(),
        position: err.position.map_or(0, |p| p as i32),
        where_: objects.where_.clone().unwrap_or_default(),
        schema_name: objects.schema.clone().unwrap_or_default(),
        table_name: objects.table.clone().unwrap_or_default(),
        column_name: objects.column.clone().unwrap_or_default(),
        data_type_name: objects.data_type.clone().unwrap_or_default(),
        constraint_name: objects.constraint.clone().unwrap_or_default(),
        ..ErrorFields::default()
    }
}

/// Stream is a client's connection, which TLS encrypts once the client asks for it.
enum Stream {
    Plain(TcpStream),
    Tls(Box<rustls::StreamOwned<rustls::ServerConnection, TcpStream>>),
}

impl Stream {
    /// tcp returns the TCP connection beneath the stream.
    fn tcp(&self) -> &TcpStream {
        match self {
            Stream::Plain(stream) => stream,
            Stream::Tls(stream) => &stream.sock,
        }
    }
}

impl Read for Stream {
    fn read(&mut self, buffer: &mut [u8]) -> std::io::Result<usize> {
        match self {
            Stream::Plain(stream) => stream.read(buffer),
            Stream::Tls(stream) => stream.read(buffer),
        }
    }
}

impl Write for Stream {
    fn write(&mut self, buffer: &[u8]) -> std::io::Result<usize> {
        match self {
            Stream::Plain(stream) => stream.write(buffer),
            Stream::Tls(stream) => stream.write(buffer),
        }
    }

    fn flush(&mut self) -> std::io::Result<()> {
        match self {
            Stream::Plain(stream) => stream.flush(),
            Stream::Tls(stream) => stream.flush(),
        }
    }
}

/// Conn is a client connection.
pub struct Conn {
    stream: Stream,
    frames: FrameReader,
    out: Vec<u8>,
    server: Arc<Server>,
    /// The encoding that text values go to the client in.
    client_encoding: sql::encodings::Encoding,
}

impl Conn {
    pub fn new(stream: TcpStream, server: Arc<Server>) -> Conn {
        let client_encoding = sql::encodings::UTF8;
        Conn { stream: Stream::Plain(stream), frames: FrameReader::new(), out: Vec::new(), server, client_encoding }
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
        let frame = self.read_frame()?;
        FrontendMessage::decode(frame.tag, &frame.body, password_kind)
            .map_err(|err| ConnError::Protocol(err.to_string()))
    }

    /// read_frame reads the frame of a typed message.
    fn read_frame(&mut self) -> Result<Frame, ConnError> {
        loop {
            match self.frames.next_frame() {
                Ok(Some(frame)) => return Ok(frame),
                Ok(None) => self.fill()?,
                Err(err) => return Err(ConnError::Protocol(err.to_string())),
            }
        }
    }

    /// start_tls answers an SSLRequest, encrypting the connection when the server has a certificate.
    fn start_tls(&mut self) -> Result<(), ConnError> {
        let (Some(config), Stream::Plain(stream)) = (self.server.tls.clone(), &self.stream) else {
            self.stream.write_all(b"N")?;
            return Ok(());
        };
        let stream = stream.try_clone()?;
        self.stream.write_all(b"S")?;
        let connection = rustls::ServerConnection::new(config).map_err(|err| ConnError::Protocol(err.to_string()))?;
        self.stream = Stream::Tls(Box::new(rustls::StreamOwned::new(connection, stream)));
        Ok(())
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
                FrontendMessage::SSLRequest => self.start_tls()?,
                FrontendMessage::GSSEncRequest => self.stream.write_all(b"N")?,
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
        if self.server.engine.login(&user).is_some_and(|(_, login)| !login) {
            return self.fatal(PgError::fatal("28000", format!("role \"{user}\" is not permitted to log in")));
        }
        let database = parameter("database").filter(|d| !d.is_empty()).unwrap_or_else(|| user.clone());
        let host = self.stream.tcp().peer_addr().map(|addr| addr.ip().to_string()).unwrap_or_default();
        let mut startup: Vec<(String, String)> = parameters
            .iter()
            .filter(|(name, _)| !matches!(name.as_str(), "user" | "database" | "options" | "replication"))
            .cloned()
            .collect();
        if let Some(options) = parameter("options") {
            startup.extend(command_line_settings(&options));
        }
        if !startup.iter().any(|(name, _)| name.eq_ignore_ascii_case("DateStyle")) {
            startup.push(("DateStyle".into(), "ISO, MDY".into()));
        }
        let mut session = match self.server.engine.session(&user, &host, &database, &startup) {
            Ok(session) => session,
            Err(err) => return self.fatal(err),
        };
        self.queue(BackendMessage::AuthenticationOk);
        self.queue_parameters(&mut session);
        self.queue(BackendMessage::BackendKeyData { process_id: session.state.id as u32, secret_key: vec![0; 4] });
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
                    session.prepared.remove("");
                    extended.portals.remove("");
                    let (mut outcomes, mut error) = session.execute(&query);
                    loop {
                        let copy = match outcomes.last() {
                            Some((_, Outcome::CopyIn { binary, .. })) => Some(*binary),
                            _ => None,
                        };
                        for (notices, outcome) in outcomes {
                            self.queue_notices(notices);
                            self.queue_outcome(outcome, None);
                        }
                        self.queue_notices(session.take_notices());
                        if let Some(err) = error {
                            self.queue(BackendMessage::ErrorResponse(error_fields(&err)));
                        }
                        let Some(binary) = copy else { break };
                        (outcomes, error) = self.copy_in(session, binary)?;
                    }
                    self.queue_parameters(session);
                    self.queue(BackendMessage::ReadyForQuery { tx_status: session.tx_status() });
                    self.flush()?;
                }
                FrontendMessage::Sync => {
                    if let Err(err) = session.sync() {
                        self.queue(BackendMessage::ErrorResponse(error_fields(&err)));
                    }
                    extended.failed = false;
                    self.queue_parameters(session);
                    self.queue(BackendMessage::ReadyForQuery { tx_status: session.tx_status() });
                    self.flush()?;
                }
                FrontendMessage::Flush => self.flush()?,
                FrontendMessage::Terminate => return Ok(()),
                message => {
                    let result = self.extended_message(session, &mut extended, message);
                    self.queue_notices(session.take_notices());
                    if let Err(err) = result {
                        let err = session.abort(err);
                        self.queue(BackendMessage::ErrorResponse(error_fields(&err)));
                        extended.failed = true;
                    }
                    if let Some(binary) = extended.copying.take() {
                        let (outcomes, error) = self.copy_in(session, binary)?;
                        for (notices, outcome) in outcomes {
                            self.queue_notices(notices);
                            self.queue_outcome(outcome, Some(&[]));
                        }
                        self.queue_notices(session.take_notices());
                        if let Some(err) = error {
                            self.queue(BackendMessage::ErrorResponse(error_fields(&err)));
                            extended.failed = true;
                        }
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
                if !name.is_empty() && session.prepared.contains_key(&name) {
                    return Err(PgError::new(
                        code::DUPLICATE_PREPARED_STATEMENT,
                        format!("prepared statement \"{name}\" already exists"),
                    ));
                }
                if name.is_empty() {
                    session.prepared.remove("");
                }
                let prepared = session.prepare(&query, &parameter_oids)?;
                session.prepared.insert(name, Arc::new(prepared));
                self.queue(BackendMessage::ParseComplete);
            }
            FrontendMessage::Bind {
                destination_portal,
                prepared_statement,
                parameter_format_codes,
                parameters,
                result_format_codes,
            } => {
                let prepared = session.statement(&prepared_statement)?;
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
                if !destination_portal.is_empty() && extended.portals.contains_key(&destination_portal) {
                    return Err(PgError::new(
                        code::DUPLICATE_CURSOR,
                        format!("cursor \"{destination_portal}\" already exists"),
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
                let prepared = session.statement(&name)?;
                self.queue(BackendMessage::ParameterDescription { parameter_oids: prepared.parameter_types.clone() });
                self.queue_description(prepared.columns.as_deref(), &[]);
            }
            FrontendMessage::Describe { object_type: b'P', name } => {
                let portal = extended.portal(&name)?;
                let (prepared, formats) = (portal.prepared.clone(), portal.result_formats.clone());
                self.queue_description(prepared.columns.as_deref(), &formats);
            }
            FrontendMessage::Describe { object_type, .. } => {
                return Err(PgError::new(
                    code::PROTOCOL_VIOLATION,
                    format!("invalid DESCRIBE message subtype {object_type}"),
                ));
            }
            FrontendMessage::Execute { portal, .. } => {
                let portal = extended.portal(&portal)?;
                let (prepared, formats) = (portal.prepared.clone(), portal.result_formats.clone());
                let outcome = session.execute_prepared(&prepared, &portal.parameters.clone())?;
                self.queue_notices(session.take_notices());
                if let Outcome::CopyIn { binary, .. } = outcome {
                    extended.copying = Some(binary);
                }
                self.queue_outcome(outcome, Some(&formats));
            }
            FrontendMessage::Close { object_type, name } => {
                match object_type {
                    b'S' => {
                        session.prepared.remove(&name);
                    }
                    b'P' => {
                        extended.portals.remove(&name);
                    }
                    _ => {
                        return Err(PgError::new(
                            code::PROTOCOL_VIOLATION,
                            format!("invalid CLOSE message subtype {object_type}"),
                        ));
                    }
                }
                self.queue(BackendMessage::CloseComplete);
            }
            other => return Err(PgError::unsupported(format!("the {} message", message_name(&other)))),
        }
        Ok(())
    }

    /// copy_in receives the data of a COPY FROM STDIN and finishes the copy, returning what it and the rest of its
    /// query produced, or ends the connection when the client leaves the copy protocol, as Postgres does.
    fn copy_in(&mut self, session: &mut Session, binary: bool) -> Result<(Results, Option<PgError>), ConnError> {
        self.flush()?;
        let mut data = Vec::new();
        let line = |data: &[u8]| data.iter().filter(|&&b| b == b'\n').count() + 1;
        loop {
            let frame = self.read_frame()?;
            match frame.tag {
                b'd' => data.extend_from_slice(&frame.body),
                b'c' => break,
                b'f' => {
                    let message = String::from_utf8_lossy(frame.body.strip_suffix(&[0]).unwrap_or(&frame.body));
                    let err = PgError::new(code::QUERY_CANCELED, format!("COPY from stdin failed: {message}"));
                    return Ok((Vec::new(), Some(session.abort_copy(err, line(&data)))));
                }
                b'S' | b'H' => {}
                tag => {
                    let message = format!("unexpected message type 0x{tag:02X} during COPY from stdin");
                    let err = session.abort_copy(PgError::new(code::PROTOCOL_VIOLATION, message), line(&data));
                    self.queue(BackendMessage::ErrorResponse(error_fields(&err)));
                    self.fatal(PgError::new(
                        code::PROTOCOL_VIOLATION,
                        "terminating connection because protocol synchronization was lost",
                    ))?;
                    return Err(ConnError::Protocol(format!("unexpected message type 0x{tag:02X} during COPY")));
                }
            }
        }
        if !binary && self.client_encoding != sql::encodings::UTF8 {
            match self.client_encoding.decode(&data) {
                Ok(text) => data = text.into_bytes(),
                Err(err) => return Ok((Vec::new(), Some(session.abort_copy(err, line(&data))))),
            }
        }
        Ok(session.copy_data(&data))
    }

    /// queue_parameters queues a ParameterStatus for each reported parameter that changed.
    fn queue_parameters(&mut self, session: &mut Session) {
        for (name, value) in session.parameter_changes() {
            if name == "client_encoding"
                && let Some(encoding) = sql::encodings::Encoding::lookup(&value)
            {
                self.client_encoding = encoding;
            }
            self.queue(BackendMessage::ParameterStatus { name, value });
        }
    }

    /// queue_notices queues a NoticeResponse for each notice.
    fn queue_notices(&mut self, notices: Vec<PgError>) {
        for notice in notices {
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
                    let values: Result<Vec<Option<Vec<u8>>>, PgError> = row
                        .iter()
                        .enumerate()
                        .map(|(i, value)| {
                            let bytes = value.encode(columns[i].type_oid, format(formats, i));
                            match bytes {
                                Some(text)
                                    if format(formats, i) == 0 && self.client_encoding != sql::encodings::UTF8 =>
                                {
                                    let text = String::from_utf8_lossy(&text);
                                    self.client_encoding.encode(&text).map(Some)
                                }
                                other => Ok(other),
                            }
                        })
                        .collect();
                    match values {
                        Ok(values) => self.queue(BackendMessage::DataRow { values }),
                        Err(err) => return self.queue(BackendMessage::ErrorResponse(error_fields(&err))),
                    }
                }
                self.queue(BackendMessage::CommandComplete { command_tag: tag });
            }
            Outcome::Command { tag } => self.queue(BackendMessage::CommandComplete { command_tag: tag }),
            Outcome::Empty => self.queue(BackendMessage::EmptyQueryResponse),
            Outcome::CopyIn { binary, columns } => {
                let format = u8::from(binary);
                let column_format_codes = vec![format as u16; columns];
                self.queue(BackendMessage::CopyInResponse { overall_format: format, column_format_codes });
            }
            Outcome::CopyOut { binary, columns, chunks, tag } => {
                let format = u8::from(binary);
                let column_format_codes = vec![format as u16; columns];
                self.queue(BackendMessage::CopyOutResponse { overall_format: format, column_format_codes });
                for chunk in chunks {
                    let data = if binary || self.client_encoding == sql::encodings::UTF8 {
                        chunk
                    } else {
                        match self.client_encoding.encode(&String::from_utf8_lossy(&chunk)) {
                            Ok(encoded) => encoded,
                            Err(err) => return self.queue(BackendMessage::ErrorResponse(error_fields(&err))),
                        }
                    };
                    self.queue(BackendMessage::CopyData { data });
                }
                self.queue(BackendMessage::CopyDone);
                self.queue(BackendMessage::CommandComplete { command_tag: tag });
            }
        }
    }
}

/// Extended is the state of the extended query protocol: the portals, and whether an error is discarding messages
/// until the next Sync.
#[derive(Default)]
struct Extended {
    portals: HashMap<String, Portal>,
    failed: bool,
    /// Whether an executed COPY FROM STDIN waits for its data, in the binary format or as text.
    copying: Option<bool>,
}

impl Extended {
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
            table_oid: c.origin.0,
            table_attribute_number: c.origin.1,
            format: format(formats, i),
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

/// command_line_settings returns the settings that a startup packet's options give as `-c name=value` or
/// `--name=value`, split at spaces that no backslash escapes, as Postgres' pg_split_opts and process_postgres_switches
/// read them, with dashes in names read as underscores.
fn command_line_settings(options: &str) -> Vec<(String, String)> {
    let mut words = Vec::new();
    let mut word = String::new();
    let mut chars = options.chars();
    while let Some(c) = chars.next() {
        match c {
            '\\' => word.extend(chars.next()),
            c if c.is_ascii_whitespace() => {
                if !word.is_empty() {
                    words.push(std::mem::take(&mut word));
                }
            }
            c => word.push(c),
        }
    }
    if !word.is_empty() {
        words.push(word);
    }
    let mut out = Vec::new();
    let mut words = words.into_iter();
    while let Some(word) = words.next() {
        let setting = match word.strip_prefix("--") {
            Some(setting) => Some(setting.to_string()),
            None if word == "-c" => words.next(),
            None => word.strip_prefix("-c").map(str::to_string),
        };
        if let Some((name, value)) = setting.as_deref().and_then(|s| s.split_once('=')) {
            out.push((name.replace('-', "_"), value.to_string()));
        }
    }
    out
}
