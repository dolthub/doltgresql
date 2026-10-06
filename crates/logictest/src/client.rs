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
use std::sync::atomic::{AtomicU64, Ordering};
use std::time::Instant;

use harness::pgx::ScramClient;
use pgproto::{BackendMessage, ErrorFields, FieldDescription, FrameReader, FrontendMessage};

/// CACHE_CAPACITY is the size of pgx v4's prepared statement cache.
const CACHE_CAPACITY: usize = 512;

/// BINARY_RESULTS lists the types whose results database/sql asks pgx v4 for in binary.
const BINARY_RESULTS: [u32; 13] = [16, 17, 29, 1082, 700, 701, 21, 23, 20, 26, 1114, 1184, 28];

static CACHE_COUNT: AtomicU64 = AtomicU64::new(0);

/// Error is a failed request, formatted like pgconn's errors.
#[derive(Debug)]
pub enum Error {
    Pg(Box<ErrorFields>),
    Other(String),
    Timeout,
}

impl std::fmt::Display for Error {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Error::Pg(e) => write!(f, "{}: {} (SQLSTATE {})", e.severity, e.message, e.code),
            Error::Other(message) => f.write_str(message),
            Error::Timeout => f.write_str("timeout: context deadline exceeded"),
        }
    }
}

/// Statement is a cached prepared statement.
struct Statement {
    name: String,
    sql: String,
    parameters: usize,
    fields: Vec<FieldDescription>,
}

/// QueryResult is a query's result columns and rows, where None is NULL.
pub struct QueryResult {
    pub fields: Vec<FieldDescription>,
    pub rows: Vec<Vec<Option<Vec<u8>>>>,
}

/// Client is a connection that sends what pgx v4's database/sql driver sends.
pub struct Client {
    stream: TcpStream,
    frames: FrameReader,
    cache: VecDeque<Statement>,
    prefix: String,
    prepare_count: u64,
    pub closed: bool,
    deadline: Option<Instant>,
}

impl Client {
    /// connect connects as postgres to the database, or to the default database when it is empty.
    pub fn connect(port: u16, database: &str) -> Result<Client, Error> {
        let stream = TcpStream::connect(("127.0.0.1", port)).map_err(|e| Error::Other(e.to_string()))?;
        stream.set_nodelay(true).map_err(|e| Error::Other(e.to_string()))?;
        let n = CACHE_COUNT.fetch_add(1, Ordering::Relaxed) + 1;
        let mut client = Client {
            stream,
            frames: FrameReader::new(),
            cache: VecDeque::new(),
            prefix: format!("lrupsc_{n}"),
            prepare_count: 0,
            closed: false,
            deadline: None,
        };
        let mut parameters = vec![("user".to_string(), "postgres".to_string())];
        if !database.is_empty() {
            parameters.push(("database".to_string(), database.to_string()));
        }
        client.send(&[FrontendMessage::StartupMessage {
            protocol_version: pgproto::PROTOCOL_VERSION_3 as u32,
            parameters,
        }])?;
        let mut scram = ScramClient::with_user("password", "postgres");
        loop {
            match client.receive()? {
                BackendMessage::AuthenticationSASL { .. } => client.send(&[FrontendMessage::SASLInitialResponse {
                    auth_mechanism: "SCRAM-SHA-256".to_string(),
                    data: Some(scram.client_first_message()),
                }])?,
                BackendMessage::AuthenticationSASLContinue { data } => {
                    let data = scram.client_final_message(&data).map_err(Error::Other)?;
                    client.send(&[FrontendMessage::SASLResponse { data }])?;
                }
                BackendMessage::AuthenticationSASLFinal { data } => {
                    scram.verify_server_final_message(&data).map_err(Error::Other)?
                }
                BackendMessage::ErrorResponse(e) => return Err(Error::Pg(Box::new(e))),
                BackendMessage::ReadyForQuery { .. } => return Ok(client),
                _ => {}
            }
        }
    }

    /// set_deadline bounds every later request, until it is cleared.
    pub fn set_deadline(&mut self, deadline: Option<Instant>) {
        self.deadline = deadline;
    }

    fn send(&mut self, messages: &[FrontendMessage]) -> Result<(), Error> {
        let mut buffer = Vec::new();
        for message in messages {
            message.encode(&mut buffer);
        }
        self.stream.write_all(&buffer).map_err(|e| self.broken(e.to_string()))
    }

    fn broken(&mut self, message: String) -> Error {
        self.closed = true;
        let _ = self.stream.shutdown(Shutdown::Both);
        Error::Other(message)
    }

    fn receive(&mut self) -> Result<BackendMessage, Error> {
        loop {
            match self.frames.next_frame() {
                Ok(Some(frame)) => {
                    return BackendMessage::decode(frame.tag, &frame.body).map_err(|e| self.broken(e.to_string()));
                }
                Ok(None) => {}
                Err(e) => return Err(self.broken(e.to_string())),
            }
            let timeout = match self.deadline {
                Some(deadline) => {
                    let remaining = deadline.saturating_duration_since(Instant::now());
                    if remaining.is_zero() {
                        self.broken(String::new());
                        return Err(Error::Timeout);
                    }
                    Some(remaining)
                }
                None => None,
            };
            let _ = self.stream.set_read_timeout(timeout);
            let mut buffer = [0u8; 65536];
            match self.stream.read(&mut buffer) {
                Ok(0) => return Err(self.broken("unexpected EOF".to_string())),
                Ok(n) => self.frames.extend(&buffer[..n]),
                Err(e) if matches!(e.kind(), std::io::ErrorKind::WouldBlock | std::io::ErrorKind::TimedOut) => {
                    self.broken(String::new());
                    return Err(Error::Timeout);
                }
                Err(e) => return Err(self.broken(e.to_string())),
            }
        }
    }

    /// exec runs SQL over the simple protocol and returns its first error.
    pub fn exec(&mut self, sql: &str) -> Result<(), Error> {
        self.send(&[FrontendMessage::Query { query: sql.to_string() }])?;
        let mut error = None;
        loop {
            match self.receive()? {
                BackendMessage::ErrorResponse(e) => {
                    error.get_or_insert(Error::Pg(Box::new(e)));
                }
                BackendMessage::ReadyForQuery { .. } => return error.map_or(Ok(()), Err),
                BackendMessage::CopyInResponse { .. } => {
                    self.send(&[FrontendMessage::CopyFail { message: "unexpected COPY".to_string() }])?
                }
                _ => {}
            }
        }
    }

    /// ping matches pgx v4's Ping, which runs an empty statement.
    pub fn ping(&mut self) -> Result<(), Error> {
        self.exec(";")
    }

    /// prepare moves the SQL's statement to the front of the cache, preparing it when it is missing.
    fn prepare(&mut self, sql: &str) -> Result<usize, Error> {
        if let Some(index) = self.cache.iter().position(|s| s.sql == sql) {
            let statement = self.cache.remove(index).unwrap();
            self.cache.push_front(statement);
            return Ok(0);
        }
        if self.cache.len() == CACHE_CAPACITY {
            let oldest = self.cache.pop_back().unwrap();
            self.exec(&format!("deallocate {}", oldest.name))?;
        }
        let name = format!("{}_{}", self.prefix, self.prepare_count);
        self.prepare_count += 1;
        self.send(&[
            FrontendMessage::Parse { name: name.clone(), query: sql.to_string(), parameter_oids: Vec::new() },
            FrontendMessage::Describe { object_type: b'S', name: name.clone() },
            FrontendMessage::Sync,
        ])?;
        let mut error = None;
        let mut parameters = 0;
        let mut fields = Vec::new();
        loop {
            match self.receive()? {
                BackendMessage::ParameterDescription { parameter_oids } => parameters = parameter_oids.len(),
                BackendMessage::RowDescription { fields: f } => fields = f,
                BackendMessage::ErrorResponse(e) => {
                    error.get_or_insert(Error::Pg(Box::new(e)));
                }
                BackendMessage::ReadyForQuery { .. } => break,
                _ => {}
            }
        }
        if let Some(error) = error {
            return Err(error);
        }
        self.cache.push_front(Statement { name, sql: sql.to_string(), parameters, fields });
        Ok(0)
    }

    /// query runs a query as a cached prepared statement, asking for binary results where database/sql does.
    pub fn query(&mut self, sql: &str) -> Result<QueryResult, Error> {
        let index = self.prepare(sql)?;
        let statement = &self.cache[index];
        if statement.parameters != 0 {
            return Err(Error::Other(format!("expected {} arguments, got 0", statement.parameters)));
        }
        let formats: Vec<i16> =
            statement.fields.iter().map(|f| i16::from(BINARY_RESULTS.contains(&f.data_type_oid))).collect();
        let name = statement.name.clone();
        self.send(&[
            FrontendMessage::Bind {
                destination_portal: String::new(),
                prepared_statement: name,
                parameter_format_codes: Vec::new(),
                parameters: Vec::new(),
                result_format_codes: formats,
            },
            FrontendMessage::Describe { object_type: b'P', name: String::new() },
            FrontendMessage::Execute { portal: String::new(), max_rows: 0 },
            FrontendMessage::Sync,
        ])?;
        let mut error = None;
        let mut result = QueryResult { fields: Vec::new(), rows: Vec::new() };
        loop {
            match self.receive()? {
                BackendMessage::RowDescription { fields } => result.fields = fields,
                BackendMessage::DataRow { values } => result.rows.push(values),
                BackendMessage::ErrorResponse(e) => {
                    error.get_or_insert(Error::Pg(Box::new(e)));
                }
                BackendMessage::ReadyForQuery { .. } => break,
                _ => {}
            }
        }
        error.map_or(Ok(result), Err)
    }
}
