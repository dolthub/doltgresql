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

use std::collections::HashMap;
use std::path::PathBuf;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Condvar, Mutex};
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

use harness::pgx::{Conn, ConnConfig};
use pgproto::{BackendMessage, FrontendMessage};

use crate::log;

/// OUTPUT_PLUGIN is the logical decoding plugin that the replication slot uses.
const OUTPUT_PLUGIN: &str = "pgoutput";

/// MAX_CONSECUTIVE_FAILURES is how many errors in a row stop replication.
const MAX_CONSECUTIVE_FAILURES: usize = 10;

/// STANDBY_MESSAGE_TIMEOUT is how often the replicator reports its position to the primary.
const STANDBY_MESSAGE_TIMEOUT: Duration = Duration::from_secs(10);

/// STOP_CHECK_INTERVAL is how long the replicator waits for a message before it checks whether to stop.
const STOP_CHECK_INTERVAL: Duration = Duration::from_millis(100);

/// POSTGRES_EPOCH_SECONDS is the Unix time of 2000-01-01, which replication messages count time from.
const POSTGRES_EPOCH_SECONDS: u64 = 946_684_800;

/// LogicalReplicator replicates the changes that a Postgres primary publishes to a replication slot into a Doltgres
/// server, as SQL statements over a connection to it, as Doltgres' logrepl.LogicalReplicator does.
pub struct LogicalReplicator {
    primary: String,
    replica: String,
    wal_file_path: PathBuf,
    status: Mutex<Status>,
    stopped: Condvar,
    stop_requested: AtomicBool,
}

/// Status is whether replication is running, and whether it received a message since it started.
#[derive(Default)]
struct Status {
    running: bool,
    message_received: bool,
}

/// Lsn is a position in the primary's write-ahead log.
type Lsn = u64;

/// Column is a column of a replicated table, as a Relation message describes it.
struct Column {
    /// Whether the column is part of the key that identifies rows.
    key: bool,
    name: String,
}

/// Relation is a replicated table, as a Relation message describes it.
struct Relation {
    namespace: String,
    name: String,
    columns: Vec<Column>,
}

/// State is what replication tracks while it runs.
struct State {
    replica: Conn,
    /// The commit position of the last transaction that the replica committed.
    last_written: Lsn,
    /// The last position that the primary sent.
    last_received: Lsn,
    /// The commit position of the transaction being applied.
    current_transaction: Lsn,
    /// Whether the messages of the current transaction apply, which they do unless the replica already committed it.
    process_messages: bool,
    relations: HashMap<u32, Relation>,
}

impl LogicalReplicator {
    /// new returns a replicator from the primary to the replica, connection URLs of the form
    /// postgres://user:password@host:port/database, that records its position in the WAL file.
    pub fn new(wal_file_path: PathBuf, primary: String, replica: String) -> LogicalReplicator {
        LogicalReplicator {
            primary,
            replica,
            wal_file_path,
            status: Mutex::new(Status::default()),
            stopped: Condvar::new(),
            stop_requested: AtomicBool::new(false),
        }
    }

    /// primary_dsn returns the URL of the primary for ordinary queries.
    pub fn primary_dsn(&self) -> &str {
        &self.primary
    }

    /// replication_dsn returns the URL of the primary for replication commands.
    pub fn replication_dsn(&self) -> String {
        let separator = if self.primary.contains('?') { '&' } else { '?' };
        format!("{}{separator}replication=database", self.primary)
    }

    /// status locks the replicator's status.
    fn status(&self) -> std::sync::MutexGuard<'_, Status> {
        self.status.lock().unwrap_or_else(|poisoned| poisoned.into_inner())
    }

    /// running reports whether replication is running.
    pub fn running(&self) -> bool {
        self.status().running
    }

    /// stop stops replication and waits until it has stopped.
    pub fn stop(&self) {
        let mut status = self.status();
        if !status.running {
            return;
        }
        log("stopping replication...");
        self.stop_requested.store(true, Ordering::SeqCst);
        while status.running {
            status = self.stopped.wait(status).unwrap_or_else(|poisoned| poisoned.into_inner());
        }
    }

    /// caught_up reports whether the primary is fewer than the threshold of bytes ahead of what the replica wrote,
    /// which only works with a single replication slot on the primary, so it suits tests.
    pub fn caught_up(&self, threshold: i64) -> Result<bool, String> {
        if !self.status().message_received {
            return Ok(false);
        }
        let mut conn = connect(&self.primary)?;
        let result = conn
            .simple_query_rows(
                "SELECT pg_wal_lsn_diff(write_lsn, sent_lsn) AS replication_lag FROM pg_stat_replication",
            )
            .map_err(|e| e.to_string())?;
        if let Some(err) = result.error {
            return Err(err.to_string());
        }
        for row in &result.rows {
            match row.first().and_then(Option::as_ref).and_then(|v| String::from_utf8_lossy(v).parse::<f64>().ok()) {
                Some(lag) => {
                    log(&format!("Current replication lag: {lag}"));
                    return Ok((lag.abs() as i64) < threshold);
                }
                None => log("Replication lag unknown"),
            }
        }
        Ok(true)
    }

    /// start_replication replicates from the slot until stop is called or an error stops it.
    pub fn start_replication(&self, slot: &str) -> Result<(), String> {
        let last_written = self.read_wal_position()?;
        let replica = connect(&self.replica)?;
        let mut state = State {
            replica,
            last_written,
            last_received: 0,
            current_transaction: 0,
            process_messages: false,
            relations: HashMap::new(),
        };
        log("Starting replicator");
        self.stop_requested.store(false, Ordering::SeqCst);
        *self.status() = Status { running: true, message_received: false };
        let result = self.replicate(slot, &mut state);
        if let Err(err) = &result {
            log(&format!("Error during replication: {err}"));
        }
        log("shutting down replicator");
        self.status().running = false;
        self.stopped.notify_all();
        result
    }

    /// replicate runs the replication loop, reconnecting to the primary after errors until too many happen in a row.
    fn replicate(&self, slot: &str, state: &mut State) -> Result<(), String> {
        let mut primary: Option<Conn> = None;
        let mut next_standby_message = Instant::now() + STANDBY_MESSAGE_TIMEOUT;
        let mut failures = 0;
        loop {
            if self.stop_requested.load(Ordering::SeqCst) {
                return Ok(());
            }
            let Some(conn) = primary.as_mut() else {
                match self.begin_replication(slot, state.last_written) {
                    Ok(conn) => primary = Some(conn),
                    Err(err) => {
                        std::thread::sleep(Duration::from_secs(3));
                        retry(err, true, &mut failures, &mut primary)?;
                    }
                }
                continue;
            };
            if Instant::now() > next_standby_message && state.last_received > 0 {
                if let Err(err) = send_standby_status_update(conn, state) {
                    retry(err, false, &mut failures, &mut primary)?;
                    continue;
                }
                next_standby_message = Instant::now() + STANDBY_MESSAGE_TIMEOUT;
            }
            let deadline = next_standby_message.min(Instant::now() + STOP_CHECK_INTERVAL);
            let message = match conn.receive_message(deadline) {
                Ok(Some(message)) => message,
                Ok(None) => continue,
                Err(err) => {
                    retry(err.to_string(), true, &mut failures, &mut primary)?;
                    continue;
                }
            };
            self.status().message_received = true;
            let data = match message {
                BackendMessage::ErrorResponse(fields) => {
                    return Err(format!("received Postgres WAL error: {}", harness::pgx::Error::pg(fields)));
                }
                BackendMessage::CopyData { data } => data,
                other => {
                    log(&format!("Received unexpected message: {other:?}"));
                    continue;
                }
            };
            match data.first() {
                Some(b'k') => {
                    let mut reader = Reader(&data[1..]);
                    let wal_end = reader.u64()?;
                    let _server_time = reader.u64()?;
                    let reply_requested = reader.u8()? != 0;
                    state.last_received = wal_end;
                    if reply_requested {
                        next_standby_message = Instant::now();
                    }
                }
                Some(b'w') => {
                    let mut reader = Reader(&data[1..]);
                    let _wal_start = reader.u64()?;
                    let wal_end = reader.u64()?;
                    let _server_time = reader.u64()?;
                    state.last_received = wal_end;
                    match process_message(reader.0, state) {
                        Ok(true) => {
                            state.last_written = state.current_transaction;
                            self.write_wal_position(state.last_written)?;
                            failures = 0;
                        }
                        Ok(false) => failures = 0,
                        Err(err) => {
                            let _ = execute(&mut state.replica, "ROLLBACK");
                            state.process_messages = false;
                            retry(err, true, &mut failures, &mut primary)?;
                            continue;
                        }
                    }
                    if let Err(err) = send_standby_status_update(conn, state) {
                        retry(err, false, &mut failures, &mut primary)?;
                        continue;
                    }
                    next_standby_message = Instant::now() + STANDBY_MESSAGE_TIMEOUT;
                }
                _ => log("Received unexpected message"),
            }
        }
    }

    /// begin_replication connects to the primary and starts streaming the slot's changes after the last position
    /// that the replica committed.
    fn begin_replication(&self, slot: &str, last_flushed: Lsn) -> Result<Conn, String> {
        let mut conn = connect(&self.replication_dsn())?;
        let start = last_flushed + 1;
        log(&format!("Starting logical replication on slot {slot} at WAL location {}", lsn_text(start)));
        let query = format!(
            "START_REPLICATION SLOT {slot} LOGICAL {} (proto_version '2', publication_names '{slot}', messages 'true')",
            lsn_text(start)
        );
        conn.send_message(&FrontendMessage::Query { query }).map_err(|e| e.to_string())?;
        loop {
            match conn.receive_message(Instant::now() + Duration::from_secs(30)).map_err(|e| e.to_string())? {
                Some(BackendMessage::CopyBothResponse { .. }) => break,
                Some(BackendMessage::ErrorResponse(fields)) => return Err(harness::pgx::Error::pg(fields).to_string()),
                Some(BackendMessage::NoticeResponse(_) | BackendMessage::ParameterStatus { .. }) => {}
                Some(other) => return Err(format!("unexpected message starting replication: {other:?}")),
                None => return Err("timed out starting replication".to_string()),
            }
        }
        log(&format!("Logical replication started on slot {slot}"));
        Ok(conn)
    }

    /// drop_replication_slot drops the slot, ignoring an error from it not existing.
    pub fn drop_replication_slot(&self, slot: &str) -> Result<(), String> {
        let mut conn = connect(&self.replication_dsn())?;
        let _ = conn.simple_query(&format!("DROP_REPLICATION_SLOT {slot}"));
        Ok(())
    }

    /// create_replication_slot_if_necessary creates the slot unless it exists.
    pub fn create_replication_slot_if_necessary(&self, slot: &str) -> Result<(), String> {
        let mut conn = connect(&self.primary)?;
        let query = format!("SELECT 1 FROM pg_replication_slots WHERE slot_name = {}", literal(slot.as_bytes()));
        let result = conn.simple_query_rows(&query).map_err(|e| e.to_string())?;
        if let Some(err) = result.error {
            return Err(err.to_string());
        }
        if !result.rows.is_empty() {
            return Ok(());
        }
        let mut conn = connect(&self.replication_dsn())?;
        let result = conn
            .simple_query(&format!("CREATE_REPLICATION_SLOT {slot} LOGICAL {OUTPUT_PLUGIN}"))
            .map_err(|e| e.to_string())?;
        match result.error {
            Some(harness::pgx::Error::Pg(err)) if err.fields.code == "42710" => {}
            Some(err) => return Err(err.to_string()),
            None => {}
        }
        log(&format!("Created replication slot: {slot}"));
        Ok(())
    }

    /// read_wal_position reads the position that the WAL file records, which is 0 without the file.
    fn read_wal_position(&self) -> Result<Lsn, String> {
        match std::fs::read_to_string(&self.wal_file_path) {
            Ok(text) => parse_lsn(&text),
            Err(err) if err.kind() == std::io::ErrorKind::NotFound => Ok(0),
            Err(err) => Err(err.to_string()),
        }
    }

    /// write_wal_position records the position in the WAL file.
    fn write_wal_position(&self, lsn: Lsn) -> Result<(), String> {
        std::fs::write(&self.wal_file_path, lsn_text(lsn)).map_err(|e| e.to_string())
    }
}

/// retry drops the connection to the primary after an error, so that replication reconnects, unless too many errors
/// happened in a row, when it returns the error.
fn retry(err: String, count: bool, failures: &mut usize, primary: &mut Option<Conn>) -> Result<(), String> {
    if count {
        *failures += 1;
    }
    if *failures >= MAX_CONSECUTIVE_FAILURES {
        return Err(err);
    }
    log(&format!("Error: {err}. Retrying"));
    *primary = None;
    Ok(())
}

/// create_publication creates a publication of every table on the primary, which tests use, since users publish the
/// tables they want themselves.
pub fn create_publication(primary: &str, slot: &str) -> Result<(), String> {
    execute(&mut connect(primary)?, &format!("CREATE PUBLICATION {slot} FOR ALL TABLES;"))
}

/// drop_publication drops the publication when it exists.
pub fn drop_publication(primary: &str, slot: &str) -> Result<(), String> {
    execute(&mut connect(primary)?, &format!("DROP PUBLICATION IF EXISTS {slot};"))
}

/// connect connects to a server by its URL.
fn connect(url: &str) -> Result<Conn, String> {
    Conn::connect(ConnConfig::parse(url).map_err(|e| e.to_string())?).map_err(|e| e.to_string())
}

/// execute runs a statement, failing with its error.
fn execute(conn: &mut Conn, sql: &str) -> Result<(), String> {
    match conn.simple_query(sql).map_err(|e| e.to_string())?.error {
        Some(err) => Err(err.to_string()),
        None => Ok(()),
    }
}

/// send_standby_status_update reports to the primary what the replica wrote and what it received.
fn send_standby_status_update(conn: &mut Conn, state: &State) -> Result<(), String> {
    let since_epoch = SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default();
    let clock = since_epoch.as_micros() as u64 - POSTGRES_EPOCH_SECONDS * 1_000_000;
    let mut data = vec![b'r'];
    data.extend_from_slice(&(state.last_written + 1).to_be_bytes());
    data.extend_from_slice(&(state.last_written + 1).to_be_bytes());
    data.extend_from_slice(&(state.last_received + 1).to_be_bytes());
    data.extend_from_slice(&clock.to_be_bytes());
    data.push(0);
    conn.send_message(&FrontendMessage::CopyData { data }).map_err(|e| e.to_string())
}

/// process_message applies one pgoutput message to the replica, returning whether it committed a transaction.
fn process_message(data: &[u8], state: &mut State) -> Result<bool, String> {
    let mut r = Reader(data);
    match r.u8()? {
        b'R' => {
            let id = r.u32()?;
            let namespace = r.string()?;
            let name = r.string()?;
            let _replica_identity = r.u8()?;
            let mut columns = Vec::new();
            for _ in 0..r.u16()? {
                let flags = r.u8()?;
                let column = r.string()?;
                let _type_oid = r.u32()?;
                let _type_modifier = r.u32()?;
                columns.push(Column { key: flags & 1 != 0, name: column });
            }
            state.relations.insert(id, Relation { namespace, name, columns });
        }
        b'B' => {
            let final_lsn = r.u64()?;
            if state.last_written > final_lsn {
                log(&format!(
                    "Received stale message, ignoring. Last written LSN: {} Message LSN: {}",
                    lsn_text(state.last_written),
                    lsn_text(final_lsn)
                ));
                state.process_messages = false;
                return Ok(false);
            }
            state.process_messages = true;
            state.current_transaction = final_lsn;
            execute(&mut state.replica, "START TRANSACTION")?;
        }
        b'C' => {
            if !state.process_messages {
                return Ok(false);
            }
            execute(&mut state.replica, "COMMIT")?;
            state.process_messages = false;
            return Ok(true);
        }
        b'I' if state.process_messages => {
            let relation = relation(state, r.u32()?)?;
            r.u8()?;
            let values = r.tuple()?;
            let columns: Vec<String> = relation.columns.iter().map(|c| identifier(&c.name)).collect();
            let values: Vec<String> = values.iter().map(|v| v.clone().unwrap_or_else(|| "NULL".to_string())).collect();
            let statement =
                format!("INSERT INTO {} ({}) VALUES ({})", table(relation), columns.join(", "), values.join(", "));
            execute(&mut state.replica, &statement)?;
        }
        b'U' if state.process_messages => {
            let relation = relation(state, r.u32()?)?;
            let mut old = None;
            let mut kind = r.u8()?;
            if kind == b'K' || kind == b'O' {
                old = Some(r.tuple()?);
                kind = r.u8()?;
            }
            if kind != b'N' {
                return Err(format!("unexpected tuple kind {}", kind as char));
            }
            let new = r.tuple()?;
            let set: Vec<String> = relation
                .columns
                .iter()
                .zip(&new)
                .filter(|(_, value)| value.as_deref() != Some(UNCHANGED))
                .map(|(c, v)| format!("{} = {}", identifier(&c.name), v.as_deref().unwrap_or("NULL")))
                .collect();
            let statement = format!(
                "UPDATE {} SET {}{}",
                table(relation),
                set.join(", "),
                where_clause(relation, old.as_ref().unwrap_or(&new))
            );
            execute(&mut state.replica, &statement)?;
        }
        b'D' if state.process_messages => {
            let relation = relation(state, r.u32()?)?;
            r.u8()?;
            let old = r.tuple()?;
            let statement = format!("DELETE FROM {}{}", table(relation), where_clause(relation, &old));
            execute(&mut state.replica, &statement)?;
        }
        b'T' if state.process_messages => {
            let count = r.u32()?;
            let options = r.u8()?;
            let mut tables = Vec::new();
            for _ in 0..count {
                tables.push(table(relation(state, r.u32()?)?));
            }
            let restart = if options & 2 != 0 { " RESTART IDENTITY" } else { "" };
            let cascade = if options & 1 != 0 { " CASCADE" } else { "" };
            execute(&mut state.replica, &format!("TRUNCATE {}{restart}{cascade}", tables.join(", ")))?;
        }
        b'I' | b'U' | b'D' | b'T' => {
            log(&format!("Received stale message, ignoring. Last written LSN: {}", lsn_text(state.last_written)));
        }
        b'Y' | b'O' | b'M' => {}
        other => log(&format!("Unknown message type in pgoutput stream: {}", other as char)),
    }
    Ok(false)
}

/// UNCHANGED stands for a TOAST value that an update left alone, which the primary does not send.
const UNCHANGED: &str = "\0unchanged";

/// relation returns a table that a Relation message described.
fn relation(state: &State, id: u32) -> Result<&Relation, String> {
    state.relations.get(&id).ok_or_else(|| format!("unknown relation ID {id}"))
}

/// table returns a table's qualified name.
fn table(relation: &Relation) -> String {
    format!("{}.{}", identifier(&relation.namespace), identifier(&relation.name))
}

/// where_clause returns the condition that matches a row by its key columns.
fn where_clause(relation: &Relation, values: &[Option<String>]) -> String {
    let conditions: Vec<String> = relation
        .columns
        .iter()
        .zip(values)
        .filter(|(c, _)| c.key)
        .map(|(c, v)| match v {
            Some(v) => format!("{} = {v}", identifier(&c.name)),
            None => format!("{} IS NULL", identifier(&c.name)),
        })
        .collect();
    if conditions.is_empty() { String::new() } else { format!(" WHERE {}", conditions.join(" AND ")) }
}

/// identifier quotes an identifier.
fn identifier(name: &str) -> String {
    format!("\"{}\"", name.replace('"', "\"\""))
}

/// literal quotes text as a string literal, which the replica casts to its column's type.
fn literal(text: &[u8]) -> String {
    format!("'{}'", String::from_utf8_lossy(text).replace('\'', "''"))
}

/// lsn_text formats a position as Postgres does, as two hexadecimal halves.
fn lsn_text(lsn: Lsn) -> String {
    format!("{:X}/{:X}", lsn >> 32, lsn as u32)
}

/// parse_lsn reads a position that lsn_text formatted.
fn parse_lsn(text: &str) -> Result<Lsn, String> {
    let invalid = || format!("failed to parse LSN: {text}");
    let (high, low) = text.trim().split_once('/').ok_or_else(invalid)?;
    let high = u64::from_str_radix(high, 16).map_err(|_| invalid())?;
    let low = u64::from_str_radix(low, 16).map_err(|_| invalid())?;
    Ok((high << 32) | low)
}

/// Reader reads the fields of a replication message.
struct Reader<'a>(&'a [u8]);

impl Reader<'_> {
    /// take reads the next bytes.
    fn take(&mut self, count: usize) -> Result<&[u8], String> {
        if self.0.len() < count {
            return Err("replication message is too short".to_string());
        }
        let (head, rest) = self.0.split_at(count);
        self.0 = rest;
        Ok(head)
    }

    /// u8 reads a byte.
    fn u8(&mut self) -> Result<u8, String> {
        Ok(self.take(1)?[0])
    }

    /// u16 reads a big-endian 16-bit integer.
    fn u16(&mut self) -> Result<u16, String> {
        let bytes = self.take(2)?;
        Ok(u16::from_be_bytes([bytes[0], bytes[1]]))
    }

    /// u32 reads a big-endian 32-bit integer.
    fn u32(&mut self) -> Result<u32, String> {
        let bytes = self.take(4)?;
        Ok(u32::from_be_bytes([bytes[0], bytes[1], bytes[2], bytes[3]]))
    }

    /// u64 reads a big-endian 64-bit integer.
    fn u64(&mut self) -> Result<u64, String> {
        Ok((u64::from(self.u32()?) << 32) | u64::from(self.u32()?))
    }

    /// string reads a null-terminated string.
    fn string(&mut self) -> Result<String, String> {
        let end = self.0.iter().position(|&b| b == 0).ok_or("replication message has an unterminated string")?;
        let text = String::from_utf8_lossy(&self.0[..end]).into_owned();
        self.0 = &self.0[end + 1..];
        Ok(text)
    }

    /// tuple reads a row's values as SQL literals, None for NULL and UNCHANGED for an unchanged TOAST value.
    fn tuple(&mut self) -> Result<Vec<Option<String>>, String> {
        let mut values = Vec::new();
        for _ in 0..self.u16()? {
            values.push(match self.u8()? {
                b'n' => None,
                b'u' => Some(UNCHANGED.to_string()),
                b't' => {
                    let length = self.u32()? as usize;
                    Some(literal(self.take(length)?))
                }
                kind => return Err(format!("unknown column data type: {}", kind as char)),
            });
        }
        Ok(values)
    }
}
