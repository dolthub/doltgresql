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

//! Script tests: setup statements followed by assertions, each run against a fresh server. Every assertion is sent
//! with the same protocol messages that the Go suite's pgx client sends, and its outcome is compared exactly against
//! what Postgres returns.

use std::collections::HashMap;
use std::fmt::Write as _;
use std::path::PathBuf;
use std::sync::mpsc;
use std::time::Duration;

use pgproto::ErrorFields;

use crate::decode::decode_binary;
use crate::pgx::{self, Arg, Conn, ConnConfig, QueryExecMode, Recorder, Time, formats};
use crate::server::{Server, Target};

/// How long a blocking assertion must stay unfinished.
const EXPECTED_BLOCKING_TIMEOUT: Duration = Duration::from_millis(200);
/// How long a blocked assertion may take to finish once it is unblocked.
const UNBLOCK_TIMEOUT: Duration = Duration::from_secs(5);

/// ScriptTest is a set of setup statements and assertions that run against their own fresh server.
#[derive(Clone, Copy, Debug)]
pub struct ScriptTest {
    /// The name of the test.
    pub name: &'static str,
    /// The database to create and use, where empty means "postgres".
    pub database: &'static str,
    /// Statements that run before the assertions, which must succeed and whose results are not checked.
    pub set_up_script: &'static [&'static str],
    /// The assertions, in order.
    pub assertions: &'static [ScriptTestAssertion],
    /// When any test in a run sets this, only those tests run. It must never be committed.
    pub focus: bool,
    /// Skips the whole test, including its setup, for the given reason.
    pub skip: Option<&'static str>,
    /// YAML appended to the doltgres config file.
    pub server_config: &'static str,
}

/// S is a ScriptTest with every field at its default, for use with struct update syntax.
pub const S: ScriptTest = ScriptTest {
    name: "",
    database: "",
    set_up_script: &[],
    assertions: &[],
    focus: false,
    skip: None,
    server_config: "",
};

/// ScriptTestAssertion is a single statement and its expected outcome.
#[derive(Clone, Copy, Debug)]
pub struct ScriptTestAssertion {
    /// The statement to run.
    pub query: &'static str,
    /// The parameter values, which are encoded the way pgx encodes the corresponding Go values.
    pub bind_vars: &'static [BindVar],
    /// The expected outcome.
    pub expected: Expected,
    /// The expected notices, in order.
    pub notices: &'static [Diagnostic],
    /// How the statement is sent, which defaults to how the Go suite sends an assertion with this outcome.
    pub flow: Flow,
    /// Runs the statement as this user on its own connection, where empty means the default connection.
    pub username: &'static str,
    /// The password of the user.
    pub password: &'static str,
    /// The named client that runs the statement in a transaction test.
    pub client: &'static str,
    /// Expects the statement to block until a later assertion on another client unblocks it.
    pub expected_blocking: bool,
    /// Closes the named client instead of running the statement.
    pub close_client: bool,
    /// Sends this testdata file to a COPY FROM STDIN statement.
    pub copy_from_stdin_file: &'static str,
    /// Expects a COPY TO STDOUT statement to send the contents of this testdata file.
    pub copy_to_stdout_file: &'static str,
    /// Pipes the output of a COPY TO STDOUT statement into this COPY FROM STDIN statement.
    pub copy_round_trip_stdin_query: &'static str,
    /// When any assertion in a test sets this, only those assertions run. It must never be committed.
    pub focus: bool,
    /// Skips the assertion for the given reason.
    pub skip: Option<&'static str>,
}

/// A is a ScriptTestAssertion with every field at its default, for use with struct update syntax.
pub const A: ScriptTestAssertion = ScriptTestAssertion {
    query: "",
    bind_vars: &[],
    expected: Expected::Ok,
    notices: &[],
    flow: Flow::Auto,
    username: "",
    password: "",
    client: "",
    expected_blocking: false,
    close_client: false,
    copy_from_stdin_file: "",
    copy_to_stdout_file: "",
    copy_round_trip_stdin_query: "",
    focus: false,
    skip: None,
};

/// Expected is the expected outcome of an assertion.
#[derive(Clone, Copy, Debug)]
pub enum Expected {
    /// The statement succeeds, and nothing else is checked.
    Ok,
    /// The statement returns these columns, rows, and command tag. Rows are compared in order only when the query
    /// contains ORDER BY.
    Rows {
        /// The result columns.
        columns: &'static [Column],
        /// The rows.
        rows: &'static [&'static [Cell]],
        /// The command tag.
        tag: &'static str,
    },
    /// The statement succeeds with this command tag and returns no result columns, such as most DDL and DML.
    Tag(&'static str),
    /// The statement fails with this error.
    Error(Diagnostic),
    /// The client fails before the server can answer, with an error containing this text, such as a parameter that
    /// pgx cannot encode.
    ClientError(&'static str),
}

/// Flow is how an assertion's statement is sent.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Flow {
    /// Query when the statement succeeds with rows or a tag, and Exec when it fails or is unchecked, which is how
    /// the Go suite sends most assertions.
    Auto,
    /// pgx's Exec: the simple protocol without parameters, and a prepared statement with them.
    Exec,
    /// pgx's Query: a described statement executed with pgx's result formats.
    Query,
}

/// Column is an expected result column: its name, and its type OID or USER_DEFINED for any type outside the
/// built-in range.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Column(pub &'static str, pub u32);

/// USER_DEFINED matches any type OID of 16384 or more, since user-defined types get arbitrary OIDs.
pub const USER_DEFINED: u32 = 0;

/// Cell is an expected value.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Cell {
    /// The SQL NULL.
    Null,
    /// A value with this text. Binary values are rendered by the decoder.
    Text(&'static str),
    /// Any value that is not NULL, for values that are inherently arbitrary such as OIDs and process IDs.
    Any,
}

/// Diagnostic is an expected error or notice. Empty strings and zero positions mean the field is absent.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Diagnostic {
    /// The severity, such as ERROR or NOTICE.
    pub severity: &'static str,
    /// The SQLSTATE code.
    pub code: &'static str,
    /// The message.
    pub message: &'static str,
    /// Matches any message that contains the message, as a fallback where an exact match is prohibitive.
    pub message_contains: bool,
    /// The detail.
    pub detail: &'static str,
    /// The hint.
    pub hint: &'static str,
    /// The one-based character position in the query.
    pub position: i32,
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

/// E is an ERROR Diagnostic with every other field at its default, for use with struct update syntax.
pub const E: Diagnostic = Diagnostic {
    severity: "ERROR",
    code: "",
    message: "",
    message_contains: false,
    detail: "",
    hint: "",
    position: 0,
    schema: "",
    table: "",
    column: "",
    data_type: "",
    constraint: "",
};

/// N is a NOTICE Diagnostic with every other field at its default, for use with struct update syntax.
pub const N: Diagnostic = Diagnostic { severity: "NOTICE", ..E };

/// BindVar is a parameter value, mirroring the Go value a test would pass to pgx.
#[derive(Clone, Copy, Debug, PartialEq)]
pub enum BindVar {
    /// An untyped Go nil.
    Null,
    /// A Go int.
    Int(i64),
    /// A Go int32.
    Int32(i32),
    /// A Go int64.
    Int64(i64),
    /// A Go float64.
    Float64(f64),
    /// A Go bool.
    Bool(bool),
    /// A Go string.
    Str(&'static str),
    /// A Go []byte.
    Bytes(&'static [u8]),
    /// A Go time.Time.
    Time(Time),
    /// A pgtype.Date.
    Date(Time),
    /// A pgtype.Timestamp.
    Timestamp(Time),
    /// A pgtype.Numeric scanned from the text.
    Numeric(&'static str),
    /// A pgtype.UUID.
    Uuid([u8; 16]),
    /// A Go []string.
    StrArray(&'static [&'static str]),
    /// A Go []int32.
    Int32Array(&'static [i32]),
}

impl BindVar {
    /// to_arg converts the value to the client's argument type.
    fn to_arg(self) -> Arg {
        match self {
            BindVar::Null => Arg::Null,
            BindVar::Int(v) => Arg::Int(v),
            BindVar::Int32(v) => Arg::Int32(v),
            BindVar::Int64(v) => Arg::Int64(v),
            BindVar::Float64(v) => Arg::Float64(v),
            BindVar::Bool(v) => Arg::Bool(v),
            BindVar::Str(v) => Arg::Str(v.to_string()),
            BindVar::Bytes(v) => Arg::Bytes(v.to_vec()),
            BindVar::Time(v) => Arg::Time(v),
            BindVar::Date(v) => Arg::Date(v),
            BindVar::Timestamp(v) => Arg::Timestamp(v),
            BindVar::Numeric(v) => Arg::Numeric(v.to_string()),
            BindVar::Uuid(v) => Arg::Uuid(v),
            BindVar::StrArray(v) => Arg::StrArray(v.iter().map(|s| s.to_string()).collect()),
            BindVar::Int32Array(v) => Arg::Int32Array(v.to_vec()),
        }
    }
}

/// Observation is everything an assertion returned.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Observation {
    /// The result columns as (name, type OID).
    pub columns: Vec<(String, u32)>,
    /// The rows, rendered as text, where None is NULL.
    pub rows: Vec<Vec<Option<String>>>,
    /// Whether the statement was sent with Query, which is when its rows are known.
    pub queried: bool,
    /// The command tag.
    pub tag: String,
    /// The server's error.
    pub error: Option<ErrorFields>,
    /// An error that did not come from the server, such as a parameter that could not be encoded.
    pub client_error: Option<String>,
    /// The notices received while running the statement.
    pub notices: Vec<ErrorFields>,
    /// The data sent by COPY TO STDOUT.
    pub copy_out: Option<Vec<u8>>,
    /// Whether the assertion was skipped.
    pub skipped: bool,
}

/// testdata_dir returns the directory holding COPY test files: DOLTGRES_TESTDATA when it is set, and otherwise the
/// testdata directory of the crate whose tests are running, which cargo names in CARGO_MANIFEST_DIR.
pub fn testdata_dir() -> PathBuf {
    if let Some(dir) = std::env::var_os("DOLTGRES_TESTDATA") {
        return PathBuf::from(dir);
    }
    let manifest = std::env::var_os("CARGO_MANIFEST_DIR").map(PathBuf::from).unwrap_or_default();
    manifest.join("testdata")
}

/// BlockedResult is the connection and outcome of a blocking statement once it finishes.
type BlockedResult = (Conn, Result<String, pgx::Error>);

/// Session is the server and connections that a script runs on.
pub struct Session {
    /// The target the server was started from.
    pub target: Target,
    /// The server, which stops when the session is dropped.
    pub server: Server,
    database: String,
    default: Conn,
    other: Option<(Conn, String, String)>,
    clients: HashMap<String, Conn>,
    blocked: HashMap<String, mpsc::Receiver<BlockedResult>>,
    recorder: Option<Recorder>,
}

/// The environment variable naming the directory that receives recordings of the bytes clients send.
pub const RECORD_DIR_ENV: &str = "DOLTGRES_RECORD_DIR";

/// sanitize makes a name safe to use in a file name.
fn sanitize(name: &str) -> String {
    name.chars().map(|c| if c.is_ascii_alphanumeric() { c } else { '_' }).collect()
}

impl Session {
    /// start starts a fresh server and connects to the database the way the Go suite does: it creates the database
    /// over a connection without one, then connects to it with describe-exec mode and pings it.
    pub fn start(target: &Target, database: &str, server_config: &str) -> Result<Session, String> {
        let server = Server::start(target, server_config)?;
        let database = if database.is_empty() { "postgres".to_string() } else { database.to_string() };
        let url = |database: &str| {
            format!("postgres://postgres:password@127.0.0.1:{}/{database}?DateStyle=ISO%2C%20MDY", server.port)
        };
        let recorder = std::env::var_os(RECORD_DIR_ENV).map(|_| Recorder::new());
        let mut setup_config = ConnConfig::parse(&url("")).map_err(|err| err.to_string())?;
        setup_config.recorder = recorder.clone();
        let mut setup = connect_with_retries(setup_config)?;
        let create = if target.is_postgres() {
            (database != "postgres").then(|| format!("CREATE DATABASE {database}"))
        } else {
            Some(format!("CREATE DATABASE IF NOT EXISTS {database}"))
        };
        if let Some(create) = create {
            setup.exec(&create, &[]).map_err(|err| format!("{create}: {err}"))?;
        }
        setup.close();
        let mut config = ConnConfig::parse(&url(&database)).map_err(|err| err.to_string())?;
        config.default_query_exec_mode = QueryExecMode::DescribeExec;
        config.recorder = recorder.clone();
        let mut default = Conn::connect(config).map_err(|err| err.to_string())?;
        default.ping().map_err(|err| format!("ping: {err}"))?;
        Ok(Session {
            target: target.clone(),
            server,
            database,
            default,
            other: None,
            clients: HashMap::new(),
            blocked: HashMap::new(),
            recorder,
        })
    }

    /// save_recording writes the bytes every connection sent, one hexadecimal line per connection, to a file in
    /// DOLTGRES_RECORD_DIR named after the running test and script, when recording is enabled.
    pub fn save_recording(&self, script_name: &str) {
        let (Some(recorder), Some(dir)) = (&self.recorder, std::env::var_os(RECORD_DIR_ENV)) else { return };
        let test = std::thread::current().name().unwrap_or("unknown").rsplit("::").next().unwrap_or("unknown").to_string();
        let base = format!("{test}__{}", sanitize(script_name));
        let mut path = PathBuf::from(&dir).join(format!("{base}.hex"));
        let mut counter = 1;
        while path.exists() {
            counter += 1;
            path = PathBuf::from(&dir).join(format!("{base}__{counter}.hex"));
        }
        let text: String = recorder
            .take()
            .iter()
            .map(|bytes| bytes.iter().map(|b| format!("{b:02x}")).collect::<String>() + "\n")
            .collect();
        let _ = std::fs::write(path, text);
    }

    /// set_up runs a setup statement on the default connection.
    pub fn set_up(&mut self, query: &str) -> Result<(), String> {
        self.default.exec(query, &[]).map(|_| ()).map_err(|err| format!("error running setup query: {query}: {err}"))
    }

    /// connection returns the connection for an assertion's user, connecting as Connection.Connect does in the Go
    /// suite: the default connection for no user, and otherwise one connection per user and password.
    fn connection(&mut self, username: &str, password: &str) -> Result<&mut Conn, pgx::Error> {
        if username.is_empty() {
            if let Some((mut other, _, _)) = self.other.take() {
                other.close();
            }
            return Ok(&mut self.default);
        }
        let reuse = matches!(&self.other, Some((_, user, pass)) if user == username && pass == password);
        if !reuse {
            if let Some((mut other, _, _)) = self.other.take() {
                other.close();
            }
            let url = format!("postgres://{username}:{password}@127.0.0.1:{}/{}", self.server.port, self.database);
            let mut config = ConnConfig::parse(&url)?;
            config.recorder = self.recorder.clone();
            let conn = Conn::connect(config)?;
            self.other = Some((conn, username.to_string(), password.to_string()));
        }
        Ok(&mut self.other.as_mut().unwrap().0)
    }

    /// client returns a transaction test's named client, connecting it with the default connection's
    /// configuration the first time.
    fn client(&mut self, name: &str) -> Result<&mut Conn, String> {
        if !self.clients.contains_key(name) {
            let conn = Conn::connect(self.default.config().clone()).map_err(|err| err.to_string())?;
            self.clients.insert(name.to_string(), conn);
        }
        Ok(self.clients.get_mut(name).unwrap())
    }

    /// run runs an assertion and returns what it observed. Transaction-test fields are honored when the assertion
    /// names a client.
    pub fn run(&mut self, assertion: &ScriptTestAssertion) -> Observation {
        if assertion.skip.is_some() {
            return Observation { skipped: true, ..Observation::default() };
        }
        if !assertion.client.is_empty() {
            return self.run_on_client(assertion);
        }
        let conn = match self.connection(assertion.username, assertion.password) {
            Ok(conn) => conn,
            Err(err) => return error_observation(err, Vec::new()),
        };
        conn.take_notices();
        let mut observation = execute(conn, assertion);
        observation.notices.extend(conn.take_notices());
        observation
    }

    /// run_on_client runs an assertion of a transaction test on its named client.
    fn run_on_client(&mut self, assertion: &ScriptTestAssertion) -> Observation {
        let name = assertion.client.to_string();
        if let Some(receiver) = self.blocked.remove(&name) {
            match receiver.recv_timeout(UNBLOCK_TIMEOUT) {
                Ok((conn, result)) => {
                    self.clients.insert(name.clone(), conn);
                    if let Err(err) = result {
                        return Observation {
                            client_error: Some(format!("blocked query for client {name} failed: {err}")),
                            ..Observation::default()
                        };
                    }
                }
                Err(_) => {
                    return Observation {
                        client_error: Some(format!("blocked query for client {name} did not complete")),
                        ..Observation::default()
                    };
                }
            }
        }
        if assertion.close_client {
            return match self.clients.remove(&name) {
                Some(mut conn) => {
                    conn.close();
                    Observation::default()
                }
                None => Observation {
                    client_error: Some(format!("cannot close unknown client {name}")),
                    ..Observation::default()
                },
            };
        }
        if let Err(err) = self.client(&name) {
            return Observation { client_error: Some(err), ..Observation::default() };
        }
        if assertion.expected_blocking {
            let mut conn = self.clients.remove(&name).unwrap();
            let query = assertion.query.to_string();
            let args: Vec<Arg> = assertion.bind_vars.iter().map(|v| v.to_arg()).collect();
            let (sender, receiver) = mpsc::channel();
            std::thread::spawn(move || {
                let result = conn.exec(&query, &args);
                let _ = sender.send((conn, result));
            });
            return match receiver.recv_timeout(EXPECTED_BLOCKING_TIMEOUT) {
                Err(mpsc::RecvTimeoutError::Timeout) => {
                    self.blocked.insert(name, receiver);
                    Observation::default()
                }
                Ok((conn, result)) => {
                    self.clients.insert(name, conn);
                    Observation {
                        client_error: Some(format!("query completed before blocking timeout: {result:?}")),
                        ..Observation::default()
                    }
                }
                Err(mpsc::RecvTimeoutError::Disconnected) => {
                    Observation { client_error: Some("blocked query thread panicked".into()), ..Observation::default() }
                }
            };
        }
        let conn = self.clients.get_mut(&name).unwrap();
        conn.take_notices();
        let mut observation = execute(conn, assertion);
        observation.notices.extend(conn.take_notices());
        observation
    }

    /// finish waits for every query that is still blocked, returning an error for any that fail.
    pub fn finish(&mut self) -> Vec<String> {
        let mut errors = Vec::new();
        for (name, receiver) in self.blocked.drain() {
            match receiver.recv_timeout(UNBLOCK_TIMEOUT) {
                Ok((_, Ok(_))) => {}
                Ok((_, Err(err))) => errors.push(format!("blocked query for client {name} failed: {err}")),
                Err(_) => errors.push(format!("blocked query for client {name} did not complete")),
            }
        }
        errors
    }
}

/// connect_with_retries connects, retrying like the Go suite does in case the server is not ready.
fn connect_with_retries(config: ConnConfig) -> Result<Conn, String> {
    let mut last = String::new();
    for attempt in 0..3 {
        match Conn::connect(config.clone()) {
            Ok(conn) => return Ok(conn),
            Err(err) => last = err.to_string(),
        }
        if attempt < 2 {
            std::thread::sleep(Duration::from_secs(1));
        }
    }
    Err(last)
}

/// error_observation turns a client error into an observation.
fn error_observation(err: pgx::Error, notices: Vec<ErrorFields>) -> Observation {
    match err.pg_error() {
        Some(pg) => Observation { error: Some(pg.fields.clone()), notices, ..Observation::default() },
        None => Observation { client_error: Some(err.to_string()), notices, ..Observation::default() },
    }
}

/// effective_flow returns how an assertion is sent, choosing like the Go suite when the flow is automatic.
pub fn effective_flow(assertion: &ScriptTestAssertion) -> Flow {
    match assertion.flow {
        Flow::Auto => match assertion.expected {
            Expected::Rows { .. } | Expected::Tag(_) => Flow::Query,
            Expected::Ok | Expected::Error(_) | Expected::ClientError(_) => Flow::Exec,
        },
        flow => flow,
    }
}

/// execute sends an assertion's statement on the connection and observes the outcome.
fn execute(conn: &mut Conn, assertion: &ScriptTestAssertion) -> Observation {
    let args: Vec<Arg> = assertion.bind_vars.iter().map(|v| v.to_arg()).collect();
    if !assertion.copy_from_stdin_file.is_empty() {
        let data = match std::fs::read(testdata_dir().join(assertion.copy_from_stdin_file)) {
            Ok(data) => data,
            Err(err) => return Observation { client_error: Some(err.to_string()), ..Observation::default() },
        };
        return match conn.copy_from(assertion.query, &data) {
            Ok(tag) => Observation { tag, ..Observation::default() },
            Err(err) => error_observation(err, Vec::new()),
        };
    }
    if !assertion.copy_round_trip_stdin_query.is_empty() {
        return match conn.copy_to(assertion.query) {
            Ok((data, _)) => match conn.copy_from(assertion.copy_round_trip_stdin_query, &data) {
                Ok(tag) => Observation { tag, copy_out: Some(data), ..Observation::default() },
                Err(err) => error_observation(err, Vec::new()),
            },
            Err(err) => error_observation(err, Vec::new()),
        };
    }
    if !assertion.copy_to_stdout_file.is_empty() {
        return match conn.copy_to(assertion.query) {
            Ok((data, tag)) => Observation { tag, copy_out: Some(data), ..Observation::default() },
            Err(err) => error_observation(err, Vec::new()),
        };
    }
    match effective_flow(assertion) {
        Flow::Query => match conn.query(assertion.query, &args) {
            Ok(result) => {
                let mut observation = Observation {
                    columns: result.fields.iter().map(|f| (f.name.clone(), f.data_type_oid)).collect(),
                    queried: true,
                    tag: result.command_tag,
                    ..Observation::default()
                };
                for row in &result.rows {
                    let mut cells = Vec::with_capacity(row.len());
                    for (value, field) in row.iter().zip(&result.fields) {
                        cells.push(match value {
                            None => None,
                            Some(bytes) => Some(match render(field.data_type_oid, field.format, bytes) {
                                Ok(text) => text,
                                Err(err) => {
                                    observation.client_error = Some(err);
                                    String::new()
                                }
                            }),
                        });
                    }
                    observation.rows.push(cells);
                }
                if let Some(err) = result.error {
                    let error = error_observation(err, Vec::new());
                    observation.error = error.error;
                    observation.client_error = observation.client_error.or(error.client_error);
                }
                observation
            }
            Err(err) => error_observation(err, Vec::new()),
        },
        _ => match conn.exec(assertion.query, &args) {
            Ok(tag) => Observation { tag, ..Observation::default() },
            Err(err) => error_observation(err, Vec::new()),
        },
    }
}

/// render renders a result value as text.
fn render(oid: u32, format: i16, bytes: &[u8]) -> Result<String, String> {
    if format == formats::BINARY {
        decode_binary(oid, bytes).map_err(|err| format!("cannot decode a value of type {oid}: {err}"))
    } else {
        String::from_utf8(bytes.to_vec()).map_err(|err| format!("a value of type {oid} is not UTF-8: {err}"))
    }
}

/// is_ordered reports whether a query's rows are compared in order, which the Go suite decides by whether the
/// query contains ORDER BY.
pub fn is_ordered(query: &str) -> bool {
    query.to_lowercase().contains("order by")
}

/// check compares an observation against an assertion's expectations, returning every difference.
pub fn check(assertion: &ScriptTestAssertion, observation: &Observation) -> Vec<String> {
    let mut problems = Vec::new();
    if let Expected::ClientError(expected) = &assertion.expected {
        match &observation.client_error {
            Some(actual) if actual.contains(expected) => {}
            actual => problems.push(format!("expected a client error containing {expected:?}, got {actual:?}")),
        }
        return problems;
    }
    if let Some(err) = &observation.client_error {
        problems.push(format!("client error: {err}"));
    }
    match &assertion.expected {
        Expected::Error(expected) => match &observation.error {
            Some(actual) => check_diagnostic("error", expected, actual, &mut problems),
            None => problems.push(format!("expected error {}, but the statement succeeded", expected.code)),
        },
        expected => {
            if let Some(actual) = &observation.error {
                problems.push(format!("unexpected error: {} ({}) {}", actual.message, actual.code, actual.detail));
            }
            match expected {
                Expected::Rows { columns, rows, tag } => {
                    check_columns(columns, &observation.columns, &mut problems);
                    check_rows(rows, &observation.rows, is_ordered(assertion.query), &mut problems);
                    if *tag != observation.tag {
                        problems.push(format!("expected tag {tag:?}, got {:?}", observation.tag));
                    }
                }
                Expected::Tag(tag) => {
                    if *tag != observation.tag {
                        problems.push(format!("expected tag {tag:?}, got {:?}", observation.tag));
                    }
                    if !observation.columns.is_empty() {
                        problems.push(format!("expected no result columns, got {:?}", observation.columns));
                    }
                }
                Expected::Ok | Expected::Error(_) | Expected::ClientError(_) => {}
            }
        }
    }
    if assertion.notices.len() != observation.notices.len() {
        problems.push(format!(
            "expected {} notices, got {}: {:?}",
            assertion.notices.len(),
            observation.notices.len(),
            observation.notices.iter().map(|n| n.message.as_str()).collect::<Vec<_>>()
        ));
    } else {
        for (expected, actual) in assertion.notices.iter().zip(&observation.notices) {
            check_diagnostic("notice", expected, actual, &mut problems);
        }
    }
    if !assertion.copy_to_stdout_file.is_empty() {
        match std::fs::read(testdata_dir().join(assertion.copy_to_stdout_file)) {
            Ok(expected) => {
                if observation.copy_out.as_deref() != Some(expected.as_slice()) {
                    problems.push(format!(
                        "COPY output differs from {}: got {:?}",
                        assertion.copy_to_stdout_file,
                        observation.copy_out.as_ref().map(|data| String::from_utf8_lossy(data).into_owned())
                    ));
                }
            }
            Err(err) => problems.push(format!("cannot read {}: {err}", assertion.copy_to_stdout_file)),
        }
    }
    problems
}

/// check_other_rows compares the rows a query returned against expected rows in any order, for queries that only
/// observe state, such as a wire test's checks on a separate connection.
pub fn check_other_rows(expected: &[&[Cell]], observation: &Observation) -> Vec<String> {
    let mut problems = Vec::new();
    if let Some(err) = &observation.client_error {
        problems.push(format!("client error: {err}"));
    }
    if let Some(err) = &observation.error {
        problems.push(format!("unexpected error: {} ({})", err.message, err.code));
    }
    check_rows(expected, &observation.rows, false, &mut problems);
    problems
}

/// check_diagnostic compares an error or notice field by field.
fn check_diagnostic(kind: &str, expected: &Diagnostic, actual: &ErrorFields, problems: &mut Vec<String>) {
    let mut differences = Vec::new();
    if expected.message_contains && !actual.message.contains(expected.message) {
        differences.push(format!("message: expected to contain {:?}, got {:?}", expected.message, actual.message));
    }
    let mut compare = |field: &str, expected: &str, actual: &str| {
        if expected != actual {
            differences.push(format!("{field}: expected {expected:?}, got {actual:?}"));
        }
    };
    compare("severity", expected.severity, &actual.severity);
    compare("code", expected.code, &actual.code);
    if !expected.message_contains {
        compare("message", expected.message, &actual.message);
    }
    compare("detail", expected.detail, &actual.detail);
    compare("hint", expected.hint, &actual.hint);
    compare("position", &expected.position.to_string(), &actual.position.to_string());
    compare("schema", expected.schema, &actual.schema_name);
    compare("table", expected.table, &actual.table_name);
    compare("column", expected.column, &actual.column_name);
    compare("data type", expected.data_type, &actual.data_type_name);
    compare("constraint", expected.constraint, &actual.constraint_name);
    if !differences.is_empty() {
        problems.push(format!("{kind} differs: {}", differences.join("; ")));
    }
}

/// check_columns compares result columns.
fn check_columns(expected: &[Column], actual: &[(String, u32)], problems: &mut Vec<String>) {
    let matches = expected.len() == actual.len()
        && expected
            .iter()
            .zip(actual)
            .all(|(e, (name, oid))| e.0 == name && (e.1 == *oid || (e.1 == USER_DEFINED && *oid >= 16384)));
    if !matches {
        problems.push(format!(
            "expected columns {:?}, got {:?}",
            expected.iter().map(|c| (c.0, c.1)).collect::<Vec<_>>(),
            actual
        ));
    }
}

/// cell_matches reports whether a value matches an expected cell.
fn cell_matches(expected: &Cell, actual: &Option<String>) -> bool {
    match (expected, actual) {
        (Cell::Null, None) => true,
        (Cell::Text(text), Some(value)) => text == value,
        (Cell::Any, Some(_)) => true,
        _ => false,
    }
}

/// row_matches reports whether a row matches an expected row.
fn row_matches(expected: &[Cell], actual: &[Option<String>]) -> bool {
    expected.len() == actual.len() && expected.iter().zip(actual).all(|(e, a)| cell_matches(e, a))
}

/// check_rows compares rows, in order or as a multiset.
fn check_rows(expected: &[&[Cell]], actual: &[Vec<Option<String>>], ordered: bool, problems: &mut Vec<String>) {
    let matches = if ordered {
        expected.len() == actual.len() && expected.iter().zip(actual).all(|(e, a)| row_matches(e, a))
    } else {
        let mut unmatched: Vec<&Vec<Option<String>>> = actual.iter().collect();
        expected.len() == actual.len()
            && expected.iter().all(|e| match unmatched.iter().position(|a| row_matches(e, a)) {
                Some(index) => {
                    unmatched.swap_remove(index);
                    true
                }
                None => false,
            })
    };
    if !matches {
        let mut text = format!("rows differ ({}):\n  expected:\n", if ordered { "in order" } else { "any order" });
        for row in expected {
            let _ = writeln!(text, "    {row:?}");
        }
        text.push_str("  actual:\n");
        for row in actual {
            let _ = writeln!(text, "    {row:?}");
        }
        problems.push(text);
    }
}

/// focused returns the items to run: the focused ones when any are focused, and otherwise all of them. Focus fails
/// the run in CI, where it must never be committed.
fn focused<T>(items: &[T], is_focused: impl Fn(&T) -> bool, describe: impl Fn(&T) -> String) -> Vec<&T> {
    let focus: Vec<&T> = items.iter().filter(|item| is_focused(item)).collect();
    if focus.is_empty() {
        return items.iter().collect();
    }
    if std::env::var_os("GITHUB_ACTION").is_some() {
        panic!("{} has focus set, which CI does not allow", describe(focus[0]));
    }
    focus
}

/// Capture is everything a script returned.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Capture {
    /// The error that stopped the script before its assertions ran, such as a failed setup statement.
    pub setup_error: Option<String>,
    /// The index of the setup statement that failed.
    pub setup_failed_at: Option<usize>,
    /// The observation of each assertion run, in order, repeated for each repetition.
    pub observations: Vec<Observation>,
    /// Errors from blocked statements that failed to finish.
    pub finish_errors: Vec<String>,
}

/// capture_script runs a script against a fresh server and records what it returned, without checking it. Every
/// assertion runs, including focused ones' neighbors, since a capture describes the whole script.
pub fn capture_script(target: &Target, script: &ScriptTest, repetitions: usize) -> Capture {
    let mut capture = Capture::default();
    let mut session = match Session::start(target, script.database, script.server_config) {
        Ok(session) => session,
        Err(err) => {
            capture.setup_error = Some(format!("cannot start: {err}"));
            return capture;
        }
    };
    for (index, query) in script.set_up_script.iter().enumerate() {
        if let Err(err) = session.set_up(query) {
            capture.setup_error = Some(err);
            capture.setup_failed_at = Some(index);
            return capture;
        }
    }
    for _ in 0..repetitions {
        for assertion in script.assertions {
            capture.observations.push(session.run(assertion));
        }
    }
    capture.finish_errors = session.finish();
    capture
}

/// run_script runs a single script against a fresh server, returning the failures.
pub fn run_script(target: &Target, script: &ScriptTest, repetitions: usize) -> Vec<String> {
    let mut failures = Vec::new();
    if script.skip.is_some() {
        return failures;
    }
    let mut session = match Session::start(target, script.database, script.server_config) {
        Ok(session) => session,
        Err(err) => return vec![format!("{}: cannot start: {err}", script.name)],
    };
    for query in script.set_up_script {
        if let Err(err) = session.set_up(query) {
            return vec![format!("{}: {err}", script.name)];
        }
    }
    let assertions = focused(script.assertions, |a| a.focus, |a| format!("the assertion {:?}", a.query));
    for _ in 0..repetitions {
        for assertion in &assertions {
            let observation = session.run(assertion);
            if observation.skipped {
                continue;
            }
            let problems = check(assertion, &observation);
            if !problems.is_empty() {
                failures.push(format!("{} / {}\n  {}", script.name, assertion.query, problems.join("\n  ")));
            }
        }
    }
    for err in session.finish() {
        failures.push(format!("{}: {err}", script.name));
    }
    session.save_recording(script.name);
    failures
}

/// run_scripts runs scripts against the target in DOLTGRES_TEST_TARGET, panicking with every failure.
pub fn run_scripts(scripts: &[ScriptTest]) {
    run_scripts_repeated(scripts, 1);
}

/// run_scripts_repeated runs each script's assertions the given number of times on the same connection.
pub fn run_scripts_repeated(scripts: &[ScriptTest], repetitions: usize) {
    let target = Target::from_env().unwrap_or_else(|err| panic!("{err}"));
    let mut failures = Vec::new();
    for script in focused(scripts, |s| s.focus, |s| format!("the script {:?}", s.name)) {
        failures.extend(run_script(&target, script, repetitions));
    }
    if !failures.is_empty() {
        panic!("{} failures:\n\n{}", failures.len(), failures.join("\n\n"));
    }
}
