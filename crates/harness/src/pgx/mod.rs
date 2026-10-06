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

//! A client that sends exactly the protocol messages that pgx v5.9.2 sends for the operations the Go test suite
//! uses. Each method names the pgx method it mirrors.

mod args;
mod auth;
mod error;
pub mod formats;
mod stmtcache;
mod stream;

use std::collections::HashMap;

use pgproto::{BackendMessage, FieldDescription, FrontendMessage, PROTOCOL_VERSION_3, SSL_REQUEST_CODE};

pub use args::{Arg, Time};
pub use auth::ScramClient;
pub use error::{ConnectAttemptError, Error, PgError};
pub use stream::{Notification, Recorder};

use args::encode_arg;
use auth::{SCRAM_SHA_256, md5_password};
use stmtcache::{LruCache, statement_name};
use stream::Stream;

/// The capacity of pgx's statement cache.
const STATEMENT_CACHE_CAPACITY: usize = 512;
/// The size of pgx's COPY FROM buffer, less the five bytes of the CopyData header.
const COPY_CHUNK_SIZE: usize = 65536 - 5;

/// QueryExecMode is pgx's QueryExecMode, limited to the modes that the Go test suite uses.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum QueryExecMode {
    /// Prepares and caches a named statement per query, which is pgx's default.
    CacheStatement,
    /// Describes the unnamed statement before every execution.
    DescribeExec,
}

/// ConnConfig is the configuration of a connection.
#[derive(Clone, Debug)]
pub struct ConnConfig {
    /// The host to connect to.
    pub host: String,
    /// The port to connect to.
    pub port: u16,
    /// The user name.
    pub user: String,
    /// The password.
    pub password: String,
    /// The database, where empty means the server's default.
    pub database: String,
    /// The runtime parameters sent in the StartupMessage, besides the user and database.
    pub runtime_params: Vec<(String, String)>,
    /// The mode used by query and exec.
    pub default_query_exec_mode: QueryExecMode,
    /// Whether to first try TLS and fall back to plaintext, which is pgx's default sslmode of "prefer".
    pub prefer_tls: bool,
    /// Records the bytes this connection sends, when set.
    pub recorder: Option<Recorder>,
}

impl ConnConfig {
    /// parse parses a connection URL of the form postgres://user:password@host:port/database?name=value the way
    /// pgx.ParseConfig does for the URLs the Go test suite uses. Query parameters become runtime parameters, except
    /// sslmode and default_query_exec_mode which configure the client.
    pub fn parse(url: &str) -> Result<ConnConfig, Error> {
        let invalid = || Error::Other(format!("cannot parse `{url}`"));
        let rest = url.strip_prefix("postgres://").or_else(|| url.strip_prefix("postgresql://")).ok_or_else(invalid)?;
        let (rest, query) = match rest.split_once('?') {
            Some((rest, query)) => (rest, query),
            None => (rest, ""),
        };
        let (authority, database) = match rest.split_once('/') {
            Some((authority, database)) => (authority, database),
            None => (rest, ""),
        };
        let (credentials, host_port) = match authority.rsplit_once('@') {
            Some((credentials, host_port)) => (credentials, host_port),
            None => ("", authority),
        };
        let (user, password) = match credentials.split_once(':') {
            Some((user, password)) => (percent_decode(user), percent_decode(password)),
            None => (percent_decode(credentials), String::new()),
        };
        let (host, port) = match host_port.rsplit_once(':') {
            Some((host, port)) => (host.to_string(), port.parse().map_err(|_| invalid())?),
            None => (host_port.to_string(), 5432),
        };
        let mut config = ConnConfig {
            host,
            port,
            user,
            password,
            database: percent_decode(database),
            runtime_params: Vec::new(),
            default_query_exec_mode: QueryExecMode::CacheStatement,
            prefer_tls: true,
            recorder: None,
        };
        for pair in query.split('&').filter(|pair| !pair.is_empty()) {
            let (name, value) = pair.split_once('=').unwrap_or((pair, ""));
            let (name, value) = (percent_decode(name), percent_decode(value));
            match name.as_str() {
                "sslmode" => config.prefer_tls = value != "disable",
                "default_query_exec_mode" => {
                    config.default_query_exec_mode = match value.as_str() {
                        "cache_statement" => QueryExecMode::CacheStatement,
                        "describe_exec" => QueryExecMode::DescribeExec,
                        _ => return Err(Error::Other(format!("unsupported default_query_exec_mode: {value}"))),
                    }
                }
                _ => config.runtime_params.push((name, value)),
            }
        }
        Ok(config)
    }
}

/// percent_decode decodes percent escapes in a URL component.
fn percent_decode(text: &str) -> String {
    let bytes = text.as_bytes();
    let mut decoded = Vec::with_capacity(bytes.len());
    let mut index = 0;
    while index < bytes.len() {
        if bytes[index] == b'%'
            && let Some(byte) = text.get(index + 1..index + 3).and_then(|hex| u8::from_str_radix(hex, 16).ok())
        {
            decoded.push(byte);
            index += 3;
            continue;
        }
        decoded.push(bytes[index]);
        index += 1;
    }
    String::from_utf8_lossy(&decoded).into_owned()
}

/// StatementDescription describes a prepared statement.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct StatementDescription {
    /// The statement name, where empty is the unnamed statement.
    pub name: String,
    /// The statement's SQL.
    pub sql: String,
    /// The parameter types.
    pub param_oids: Vec<u32>,
    /// The result columns, with the text format that statement descriptions report.
    pub fields: Vec<FieldDescription>,
}

/// QueryResult is everything a query returned.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct QueryResult {
    /// The result columns, with the formats that were requested.
    pub fields: Vec<FieldDescription>,
    /// The rows, where None is NULL.
    pub rows: Vec<Vec<Option<Vec<u8>>>>,
    /// The command tag.
    pub command_tag: String,
    /// The error that ended the query, when it failed after it started.
    pub error: Option<Error>,
}

/// SimpleResult is everything a simple-protocol Exec returned.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct SimpleResult {
    /// The command tag of the last result.
    pub command_tag: String,
    /// The first error.
    pub error: Option<Error>,
}

/// Conn mirrors pgx.Conn.
pub struct Conn {
    stream: Stream,
    config: ConnConfig,
    process_id: u32,
    secret_key: Vec<u8>,
    prepared_statements: HashMap<String, StatementDescription>,
    statement_cache: LruCache,
    failed_describe_statement: String,
    closed: bool,
}

impl Conn {
    /// connect mirrors pgx.ConnectConfig. With TLS preferred, pgx first sends an SSLRequest, and when the server
    /// refuses, it opens a second connection in plaintext.
    pub fn connect(config: ConnConfig) -> Result<Conn, Error> {
        let address = format!("{}:{}", config.host, config.port);
        let mut attempts = Vec::new();
        let attempt_error = |stage: &str, message: String, pg_error: Option<Box<PgError>>| ConnectAttemptError {
            address: address.clone(),
            host: config.host.clone(),
            stage: stage.to_string(),
            pg_error,
            message,
        };
        if config.prefer_tls {
            match Stream::connect(&address, config.recorder.as_ref()) {
                Err(err) => attempts.push(attempt_error("dial error", err.to_string(), None)),
                Ok(mut stream) => {
                    let mut request = Vec::new();
                    request.extend_from_slice(&8i32.to_be_bytes());
                    request.extend_from_slice(&SSL_REQUEST_CODE.to_be_bytes());
                    let answer = stream.write_raw(&request).and_then(|_| stream.read_byte());
                    stream.shutdown();
                    match answer {
                        Ok(b'S') => {
                            return Err(Error::Other(
                                "the server accepted TLS, which the harness does not support".into(),
                            ));
                        }
                        Ok(_) => {
                            attempts.push(attempt_error("tls error", "server refused TLS connection".into(), None))
                        }
                        Err(err) => attempts.push(attempt_error("tls error", err.to_string(), None)),
                    }
                }
            }
        }
        match Conn::connect_one(&config, &address) {
            Ok(conn) => Ok(conn),
            Err(err) => {
                attempts.push(err);
                Err(Error::Connect { user: config.user.clone(), database: config.database.clone(), attempts })
            }
        }
    }

    /// connect_one mirrors pgconn's connectOne for a plaintext connection.
    fn connect_one(config: &ConnConfig, address: &str) -> Result<Conn, ConnectAttemptError> {
        let attempt_error = |stage: &str, err: Error| ConnectAttemptError {
            address: address.to_string(),
            host: config.host.clone(),
            stage: stage.to_string(),
            pg_error: match &err {
                Error::Pg(pg) => Some(pg.clone()),
                _ => None,
            },
            message: err.to_string(),
        };
        let mut stream = Stream::connect(address, config.recorder.as_ref())
            .map_err(|err| attempt_error("dial error", Error::from(err)))?;
        let mut parameters = config.runtime_params.clone();
        parameters.push(("user".to_string(), config.user.clone()));
        if !config.database.is_empty() {
            parameters.push(("database".to_string(), config.database.clone()));
        }
        stream.send(&FrontendMessage::StartupMessage { protocol_version: PROTOCOL_VERSION_3 as u32, parameters });
        stream.flush().map_err(|err| attempt_error("failed to write startup message", err))?;

        let mut process_id = 0;
        let mut secret_key = Vec::new();
        loop {
            let message = stream.recv().map_err(|err| attempt_error("failed to receive message", err))?;
            match message {
                BackendMessage::BackendKeyData { process_id: pid, secret_key: key } => {
                    process_id = pid;
                    secret_key = key;
                }
                BackendMessage::AuthenticationOk => {}
                BackendMessage::AuthenticationCleartextPassword => {
                    stream.send(&FrontendMessage::PasswordMessage { password: config.password.clone() });
                    stream.flush().map_err(|err| attempt_error("failed to write password message", err))?;
                }
                BackendMessage::AuthenticationMD5Password { salt } => {
                    let password = md5_password(&config.user, &config.password, &salt);
                    stream.send(&FrontendMessage::PasswordMessage { password });
                    stream.flush().map_err(|err| attempt_error("failed to write password message", err))?;
                }
                BackendMessage::AuthenticationSASL { mechanisms } => {
                    scram_auth(&mut stream, &config.password, &mechanisms)
                        .map_err(|err| attempt_error("failed SASL auth", err))?;
                }
                BackendMessage::ReadyForQuery { .. } => {
                    return Ok(Conn {
                        stream,
                        config: config.clone(),
                        process_id,
                        secret_key,
                        prepared_statements: HashMap::new(),
                        statement_cache: LruCache::new(STATEMENT_CACHE_CAPACITY),
                        failed_describe_statement: String::new(),
                        closed: false,
                    });
                }
                BackendMessage::ParameterStatus { .. }
                | BackendMessage::NoticeResponse(_)
                | BackendMessage::NegotiateProtocolVersion { .. } => {}
                BackendMessage::ErrorResponse(fields) => {
                    stream.shutdown();
                    return Err(attempt_error("server error", Error::pg(fields)));
                }
                other => {
                    stream.shutdown();
                    return Err(attempt_error(
                        "received unexpected message",
                        Error::Other(format!("unexpected message: {other:?}")),
                    ));
                }
            }
        }
    }

    /// config returns the configuration this connection was made with.
    pub fn config(&self) -> &ConnConfig {
        &self.config
    }

    /// process_id returns the backend process ID from BackendKeyData.
    pub fn process_id(&self) -> u32 {
        self.process_id
    }

    /// secret_key returns the cancellation key from BackendKeyData.
    pub fn secret_key(&self) -> &[u8] {
        &self.secret_key
    }

    /// tx_status returns the transaction status from the most recent ReadyForQuery.
    pub fn tx_status(&self) -> u8 {
        self.stream.tx_status
    }

    /// parameter_status returns the most recent value the server reported for a runtime parameter.
    pub fn parameter_status(&self, name: &str) -> Option<&str> {
        self.stream.parameter_statuses.get(name).map(String::as_str)
    }

    /// take_notices returns the notices received since the last call.
    pub fn take_notices(&mut self) -> Vec<pgproto::ErrorFields> {
        std::mem::take(&mut self.stream.notices)
    }

    /// take_notifications returns the notifications received since the last call.
    pub fn take_notifications(&mut self) -> Vec<Notification> {
        std::mem::take(&mut self.stream.notifications)
    }

    /// is_closed reports whether the connection was closed.
    pub fn is_closed(&self) -> bool {
        self.closed
    }

    /// close mirrors pgx's Conn.Close: it sends Terminate and closes the socket.
    pub fn close(&mut self) {
        if self.closed {
            return;
        }
        self.closed = true;
        self.stream.send(&FrontendMessage::Terminate);
        let _ = self.stream.flush();
        self.stream.shutdown();
    }

    /// ping mirrors pgconn's Ping, which runs an empty simple query.
    pub fn ping(&mut self) -> Result<(), Error> {
        match self.simple_query("-- ping")?.error {
            Some(err) => Err(err),
            None => Ok(()),
        }
    }

    /// exec mirrors pgx's Conn.Exec, returning the command tag.
    pub fn exec(&mut self, sql: &str, args: &[Arg]) -> Result<String, Error> {
        self.deallocate_invalidated_cached_statements()?;
        let result = self.exec_inner(sql, args);
        if result.is_err() {
            self.statement_cache.invalidate(sql);
        }
        result
    }

    /// exec_inner is the body of exec, without its cache maintenance.
    fn exec_inner(&mut self, sql: &str, args: &[Arg]) -> Result<String, Error> {
        if let Some(sd) = self.prepared_statements.get(sql).cloned() {
            return self.exec_prepared(&sd, args);
        }
        if args.is_empty() {
            let result = self.simple_query(sql)?;
            return match result.error {
                Some(err) => Err(err),
                None => Ok(result.command_tag),
            };
        }
        let sd = self.statement_description(sql)?;
        self.exec_prepared(&sd, args)
    }

    /// exec_prepared mirrors pgx's execPrepared.
    fn exec_prepared(&mut self, sd: &StatementDescription, args: &[Arg]) -> Result<String, Error> {
        let result = self.exec_statement(sd, args)?;
        match result.error {
            Some(err) => Err(err),
            None => Ok(result.command_tag),
        }
    }

    /// query mirrors pgx's Conn.Query followed by reading every row and closing the rows. Errors that happen before
    /// the query starts are returned as Err, and errors after it starts are in the result.
    pub fn query(&mut self, sql: &str, args: &[Arg]) -> Result<QueryResult, Error> {
        self.deallocate_invalidated_cached_statements()?;
        let result = self.query_inner(sql, args);
        let failed = match &result {
            Ok(result) => result.error.is_some(),
            Err(_) => true,
        };
        if failed && !sql.is_empty() {
            self.statement_cache.invalidate(sql);
        }
        result
    }

    /// query_inner is the body of query, without its cache maintenance.
    fn query_inner(&mut self, sql: &str, args: &[Arg]) -> Result<QueryResult, Error> {
        let explicit = self.prepared_statements.get(sql).cloned();
        if explicit.is_none() && sql.is_empty() {
            let result = self.simple_query(sql)?;
            return Ok(QueryResult { command_tag: result.command_tag, error: result.error, ..QueryResult::default() });
        }
        let sd = match explicit {
            Some(sd) => sd,
            None => self.statement_description(sql)?,
        };
        if sd.param_oids.len() != args.len() {
            return Err(Error::Other(format!("expected {} arguments, got {}", sd.param_oids.len(), args.len())));
        }
        self.exec_statement(&sd, args)
    }

    /// statement_description mirrors pgx's getStatementDescription for the configured mode.
    fn statement_description(&mut self, sql: &str) -> Result<StatementDescription, Error> {
        match self.config.default_query_exec_mode {
            QueryExecMode::CacheStatement => {
                if let Some(sd) = self.statement_cache.get(sql) {
                    return Ok(sd);
                }
                let sd = self.prepare(&statement_name(sql), sql)?;
                self.statement_cache.put(sd.clone());
                Ok(sd)
            }
            QueryExecMode::DescribeExec => self.prepare("", sql),
        }
    }

    /// prepare mirrors pgx's Conn.Prepare.
    pub fn prepare(&mut self, name: &str, sql: &str) -> Result<StatementDescription, Error> {
        if !self.failed_describe_statement.is_empty() {
            let failed = std::mem::take(&mut self.failed_describe_statement);
            if let Err(err) = self.deallocate(&failed) {
                self.failed_describe_statement = failed.clone();
                return Err(Error::Other(format!(
                    "failed to deallocate previously failed statement {failed:?}: {err}"
                )));
            }
        }
        if !name.is_empty()
            && let Some(sd) = self.prepared_statements.get(name)
            && sd.sql == sql
        {
            return Ok(sd.clone());
        }
        let (statement_name, key) = if name == sql {
            let digest = <sha2::Sha256 as sha2::Digest>::digest(sql.as_bytes());
            (format!("stmt_{}", auth::hex(&digest[..24])), sql.to_string())
        } else {
            (name.to_string(), name.to_string())
        };
        match self.pg_prepare(&statement_name, sql) {
            Ok(sd) => {
                if !key.is_empty() {
                    self.prepared_statements.insert(key, sd.clone());
                }
                Ok(sd)
            }
            Err(err) => {
                if matches!(err, Error::Pg(_)) {
                    self.failed_describe_statement = key;
                }
                Err(err)
            }
        }
    }

    /// pg_prepare mirrors pgconn's Prepare: Parse, Describe the statement, and Sync.
    fn pg_prepare(&mut self, name: &str, sql: &str) -> Result<StatementDescription, Error> {
        self.stream.send(&FrontendMessage::Parse {
            name: name.to_string(),
            query: sql.to_string(),
            parameter_oids: Vec::new(),
        });
        self.stream.send(&FrontendMessage::Describe { object_type: b'S', name: name.to_string() });
        self.stream.send(&FrontendMessage::Sync);
        self.stream.flush()?;
        let mut sd = StatementDescription {
            name: name.to_string(),
            sql: sql.to_string(),
            param_oids: Vec::new(),
            fields: Vec::new(),
        };
        let mut pg_error = None;
        loop {
            match self.stream.recv()? {
                BackendMessage::ParameterDescription { parameter_oids } => sd.param_oids = parameter_oids,
                BackendMessage::RowDescription { fields } => sd.fields = fields,
                BackendMessage::ErrorResponse(fields) => pg_error = Some(PgError { fields }),
                BackendMessage::ReadyForQuery { .. } => break,
                _ => {}
            }
        }
        match pg_error {
            Some(err) => Err(Error::Pg(Box::new(err))),
            None => Ok(sd),
        }
    }

    /// deallocate mirrors pgx's Conn.Deallocate. Like pgconn, it returns as soon as an error arrives, leaving the
    /// rest of the response unread.
    pub fn deallocate(&mut self, name: &str) -> Result<(), Error> {
        let statement_name = self.prepared_statements.get(name).map(|sd| sd.name.clone());
        let existed = statement_name.is_some();
        let statement_name = statement_name.unwrap_or_else(|| name.to_string());
        self.stream.send(&FrontendMessage::Close { object_type: b'S', name: statement_name });
        self.stream.send(&FrontendMessage::Sync);
        self.stream.flush()?;
        loop {
            match self.stream.recv()? {
                BackendMessage::ErrorResponse(fields) => return Err(Error::pg(fields)),
                BackendMessage::ReadyForQuery { .. } => break,
                _ => {}
            }
        }
        if existed {
            self.prepared_statements.remove(name);
        }
        Ok(())
    }

    /// deallocate_invalidated_cached_statements mirrors pgx's method of the same name, which closes evicted and
    /// invalidated cached statements in a pipeline before the next query.
    fn deallocate_invalidated_cached_statements(&mut self) -> Result<(), Error> {
        let tx_status = self.stream.tx_status;
        if tx_status != b'I' && tx_status != b'T' {
            return Ok(());
        }
        let invalidated = self.statement_cache.invalidated().to_vec();
        if invalidated.is_empty() {
            return Ok(());
        }
        for sd in &invalidated {
            self.stream.send(&FrontendMessage::Close { object_type: b'S', name: sd.name.clone() });
        }
        self.stream.send(&FrontendMessage::Sync);
        self.stream.flush()?;
        let mut pg_error = None;
        loop {
            match self.stream.recv()? {
                BackendMessage::ErrorResponse(fields) => {
                    if pg_error.is_none() {
                        pg_error = Some(PgError { fields });
                    }
                }
                BackendMessage::ReadyForQuery { .. } => break,
                _ => {}
            }
        }
        if let Some(err) = pg_error {
            return Err(Error::Other(format!("failed to deallocate cached statement(s): {err}")));
        }
        self.statement_cache.remove_invalidated();
        for sd in &invalidated {
            self.prepared_statements.remove(&sd.name);
        }
        Ok(())
    }

    /// query_with_result_format describes the unnamed statement and runs it with every result column in the given
    /// format. pgx never does this; it exists to compare the text and binary renderings of the same values.
    pub fn query_with_result_format(&mut self, sql: &str, format: i16) -> Result<QueryResult, Error> {
        let sd = self.pg_prepare("", sql)?;
        self.stream.send(&FrontendMessage::Bind {
            destination_portal: String::new(),
            prepared_statement: String::new(),
            parameter_format_codes: Vec::new(),
            parameters: Vec::new(),
            result_format_codes: vec![format],
        });
        self.stream.send(&FrontendMessage::Execute { portal: String::new(), max_rows: 0 });
        self.stream.send(&FrontendMessage::Sync);
        self.stream.flush()?;
        let mut result = QueryResult {
            fields: sd.fields.iter().map(|field| FieldDescription { format, ..field.clone() }).collect(),
            ..QueryResult::default()
        };
        self.read_execution(&mut result)?;
        Ok(result)
    }

    /// read_execution reads the rows, command tag, and first error of an execution through ReadyForQuery.
    fn read_execution(&mut self, result: &mut QueryResult) -> Result<(), Error> {
        loop {
            match self.stream.recv()? {
                BackendMessage::DataRow { values } => result.rows.push(values),
                BackendMessage::CommandComplete { command_tag } => result.command_tag = command_tag,
                BackendMessage::ErrorResponse(fields) => {
                    if result.error.is_none() {
                        result.error = Some(Error::pg(fields));
                    }
                }
                BackendMessage::ReadyForQuery { .. } => return Ok(()),
                _ => {}
            }
        }
    }

    /// exec_statement mirrors pgconn's ExecStatement followed by reading the whole result: Bind, Execute, and Sync,
    /// without describing the portal since the statement description is known.
    fn exec_statement(&mut self, sd: &StatementDescription, args: &[Arg]) -> Result<QueryResult, Error> {
        if sd.param_oids.len() != args.len() {
            return Err(Error::Other("mismatched param and argument count".to_string()));
        }
        let mut parameter_format_codes = Vec::with_capacity(args.len());
        let mut parameters = Vec::with_capacity(args.len());
        for (index, (oid, arg)) in sd.param_oids.iter().zip(args).enumerate() {
            let encoded =
                encode_arg(*oid, arg).map_err(|err| Error::Other(format!("failed to encode args[{index}]: {err}")))?;
            parameter_format_codes.push(encoded.format);
            parameters.push(encoded.value);
        }
        let result_format_codes: Vec<i16> =
            sd.fields.iter().map(|field| formats::format_code_for_oid(field.data_type_oid)).collect();
        self.stream.send(&FrontendMessage::Bind {
            destination_portal: String::new(),
            prepared_statement: sd.name.clone(),
            parameter_format_codes,
            parameters,
            result_format_codes: result_format_codes.clone(),
        });
        self.stream.send(&FrontendMessage::Execute { portal: String::new(), max_rows: 0 });
        self.stream.send(&FrontendMessage::Sync);
        self.stream.flush()?;

        let mut result = QueryResult {
            fields: sd
                .fields
                .iter()
                .zip(&result_format_codes)
                .map(|(field, format)| FieldDescription { format: *format, ..field.clone() })
                .collect(),
            ..QueryResult::default()
        };
        self.read_execution(&mut result)?;
        Ok(result)
    }

    /// simple_query mirrors pgconn's Exec read to completion: a simple Query whose results are each closed in
    /// turn, keeping the last result's command tag and the first error.
    pub fn simple_query(&mut self, sql: &str) -> Result<SimpleResult, Error> {
        self.stream.send(&FrontendMessage::Query { query: sql.to_string() });
        self.stream.flush()?;
        let mut result = SimpleResult::default();
        loop {
            match self.stream.recv()? {
                BackendMessage::RowDescription { .. } => result.command_tag.clear(),
                BackendMessage::CommandComplete { command_tag } => {
                    if result.error.is_none() {
                        result.command_tag = command_tag;
                    }
                }
                BackendMessage::EmptyQueryResponse => {
                    if result.error.is_none() {
                        result.command_tag.clear();
                    }
                }
                BackendMessage::ErrorResponse(fields) => {
                    if result.error.is_none() {
                        result.error = Some(Error::pg(fields));
                        result.command_tag.clear();
                    }
                }
                BackendMessage::ReadyForQuery { .. } => break,
                _ => {}
            }
        }
        Ok(result)
    }

    /// copy_from mirrors pgconn's CopyFrom. pgx streams the data while it watches for an early error, so the number
    /// of CopyData messages it sends after an error is timing dependent; this sends all of the data, then CopyDone.
    pub fn copy_from(&mut self, sql: &str, data: &[u8]) -> Result<String, Error> {
        self.stream.send(&FrontendMessage::Query { query: sql.to_string() });
        self.stream.flush()?;
        for chunk in data.chunks(COPY_CHUNK_SIZE) {
            self.stream.send(&FrontendMessage::CopyData { data: chunk.to_vec() });
            self.stream.flush()?;
        }
        self.stream.send(&FrontendMessage::CopyDone);
        self.stream.flush()?;
        let mut command_tag = String::new();
        let mut pg_error = None;
        loop {
            match self.stream.recv()? {
                BackendMessage::CommandComplete { command_tag: tag } => command_tag = tag,
                BackendMessage::ErrorResponse(fields) => pg_error = Some(PgError { fields }),
                BackendMessage::ReadyForQuery { .. } => break,
                _ => {}
            }
        }
        match pg_error {
            Some(err) => Err(Error::Pg(Box::new(err))),
            None => Ok(command_tag),
        }
    }

    /// copy_to mirrors pgconn's CopyTo, returning the copied data and the command tag.
    pub fn copy_to(&mut self, sql: &str) -> Result<(Vec<u8>, String), Error> {
        self.stream.send(&FrontendMessage::Query { query: sql.to_string() });
        self.stream.flush()?;
        let mut data = Vec::new();
        let mut command_tag = String::new();
        let mut pg_error = None;
        loop {
            match self.stream.recv()? {
                BackendMessage::CopyData { data: chunk } => data.extend_from_slice(&chunk),
                BackendMessage::CommandComplete { command_tag: tag } => command_tag = tag,
                BackendMessage::ErrorResponse(fields) => pg_error = Some(PgError { fields }),
                BackendMessage::ReadyForQuery { .. } => break,
                _ => {}
            }
        }
        match pg_error {
            Some(err) => Err(Error::Pg(Box::new(err))),
            None => Ok((data, command_tag)),
        }
    }
}

impl Drop for Conn {
    /// drop implements the interface Drop by closing the connection.
    fn drop(&mut self) {
        self.close();
    }
}

/// scram_auth mirrors pgconn's scramAuth over a plaintext connection.
fn scram_auth(stream: &mut Stream, password: &str, mechanisms: &[String]) -> Result<(), Error> {
    if !mechanisms.iter().any(|mechanism| mechanism == SCRAM_SHA_256) {
        return Err(Error::Other("server does not support SCRAM-SHA-256".to_string()));
    }
    let mut client = ScramClient::new(password);
    stream.send(&FrontendMessage::SASLInitialResponse {
        auth_mechanism: SCRAM_SHA_256.to_string(),
        data: Some(client.client_first_message()),
    });
    stream.flush()?;
    let server_first = match stream.recv()? {
        BackendMessage::AuthenticationSASLContinue { data } => data,
        BackendMessage::ErrorResponse(fields) => return Err(Error::pg(fields)),
        other => {
            return Err(Error::Other(format!(
                "expected AuthenticationSASLContinue message but received unexpected message {other:?}"
            )));
        }
    };
    let client_final = client.client_final_message(&server_first).map_err(Error::Other)?;
    stream.send(&FrontendMessage::SASLResponse { data: client_final });
    stream.flush()?;
    let server_final = match stream.recv()? {
        BackendMessage::AuthenticationSASLFinal { data } => data,
        BackendMessage::ErrorResponse(fields) => return Err(Error::pg(fields)),
        other => {
            return Err(Error::Other(format!(
                "expected AuthenticationSASLFinal message but received unexpected message {other:?}"
            )));
        }
    };
    client.verify_server_final_message(&server_final).map_err(Error::Other)
}
