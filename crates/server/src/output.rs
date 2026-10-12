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

//! What a connection sends: messages encoded into a buffer, which a writer thread can send while a query still runs.

use std::io::Write;
use std::net::TcpStream;
use std::sync::Arc;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::mpsc;

use pgproto::{BackendMessage, ErrorFields, FieldDescription};
use sql::{Column, Outcome, PgError, Value};

use crate::conn::ConnError;

/// BATCH is how many bytes of rows a streaming query collects before handing them to the writer thread.
const BATCH: usize = 64 * 1024;

/// Output is the messages a connection has yet to send, encoded, with the encoding that text values go to the client
/// in.
pub struct Output {
    pub out: Vec<u8>,
    pub client_encoding: sql::encodings::Encoding,
    /// The columns of the rows that a query streams, as its sink began them.
    columns: Vec<Column>,
    /// The thread that writes batches of messages to a plain TCP connection.
    writer: Option<Writer>,
}

/// Writer is a thread that writes batches of messages to a connection, where to hand them to it, and how many it has
/// yet to write.
struct Writer {
    sender: mpsc::Sender<Vec<u8>>,
    thread: std::thread::JoinHandle<()>,
    pending: Arc<AtomicUsize>,
}

impl Default for Output {
    fn default() -> Output {
        Output { out: Vec::new(), client_encoding: sql::encodings::UTF8, columns: Vec::new(), writer: None }
    }
}

impl Output {
    /// queue adds a message to the buffer.
    pub fn queue(&mut self, message: BackendMessage) {
        message.encode(&mut self.out);
    }

    /// start_writer starts the thread that writes the buffer to the connection, so that a statement never waits on
    /// the client while it runs.
    pub fn start_writer(&mut self, mut stream: TcpStream) {
        let (sender, receiver) = mpsc::channel::<Vec<u8>>();
        let pending = Arc::new(AtomicUsize::new(0));
        let written = pending.clone();
        let thread = std::thread::spawn(move || {
            for batch in receiver {
                if stream.write_all(&batch).is_err() {
                    break;
                }
                written.fetch_sub(1, Ordering::Release);
            }
        });
        self.writer = Some(Writer { sender, thread, pending });
    }

    /// send hands the buffer to the writer thread when it is still writing earlier batches, reporting false when the
    /// connection should write the buffer itself, and failing once the writer stopped because the client went away.
    pub fn send(&mut self) -> Result<bool, ConnError> {
        match &self.writer {
            Some(writer) if writer.pending.load(Ordering::Acquire) > 0 => self.hand_off().map(|_| true),
            _ => Ok(false),
        }
    }

    /// hand_off hands the buffer to the writer thread, failing once the writer stopped because the client went away.
    fn hand_off(&mut self) -> Result<(), ConnError> {
        let Some(writer) = &self.writer else { return Ok(()) };
        if !self.out.is_empty() {
            writer.pending.fetch_add(1, Ordering::Release);
            writer.sender.send(std::mem::take(&mut self.out)).map_err(|_| {
                ConnError::Io(std::io::Error::new(std::io::ErrorKind::BrokenPipe, "the connection closed"))
            })?;
        }
        Ok(())
    }

    /// finish waits until the writer thread has written everything handed to it.
    pub fn finish(&mut self) {
        if let Some(writer) = self.writer.take() {
            drop(writer.sender);
            let _ = writer.thread.join();
        }
    }

    /// queue_notices queues a NoticeResponse for each notice.
    pub fn queue_notices(&mut self, notices: Vec<PgError>) {
        for notice in notices {
            self.queue(BackendMessage::NoticeResponse(error_fields(&notice)));
        }
    }

    /// queue_description queues the RowDescription of the columns in the formats, or NoData without columns.
    pub fn queue_description(&mut self, columns: Option<&[Column]>, formats: &[i16]) {
        match columns {
            Some(columns) => self.queue(BackendMessage::RowDescription { fields: fields(columns, formats) }),
            None => self.queue(BackendMessage::NoData),
        }
    }

    /// queue_data_row writes a DataRow message of a row's values in the formats asked for, writing text in UTF-8
    /// straight into the output buffer, and leaves the buffer as it was when a value cannot be sent.
    pub fn queue_data_row(&mut self, row: &[Value], columns: &[Column], formats: &[i16]) -> Result<(), PgError> {
        let start = self.out.len();
        self.out.push(b'D');
        self.out.extend_from_slice(&[0; 4]);
        self.out.extend_from_slice(&(row.len() as u16).to_be_bytes());
        for (i, value) in row.iter().enumerate() {
            let length_at = self.out.len();
            self.out.extend_from_slice(&[0; 4]);
            let length = if format(formats, i) == 0 && self.client_encoding == sql::encodings::UTF8 {
                value.write_text(&mut self.out).then(|| self.out.len() - length_at - 4)
            } else {
                let bytes = match value.encode(columns[i].type_oid, format(formats, i)) {
                    Some(text) if format(formats, i) == 0 => {
                        match self.client_encoding.encode(&String::from_utf8_lossy(&text)) {
                            Ok(encoded) => Some(encoded),
                            Err(err) => {
                                self.out.truncate(start);
                                return Err(err);
                            }
                        }
                    }
                    other => other,
                };
                bytes.map(|bytes| {
                    self.out.extend_from_slice(&bytes);
                    bytes.len()
                })
            };
            let length = length.map_or(-1, |n| n as i32);
            self.out[length_at..length_at + 4].copy_from_slice(&length.to_be_bytes());
        }
        let size = (self.out.len() - start - 1) as u32;
        self.out[start + 1..start + 5].copy_from_slice(&size.to_be_bytes());
        Ok(())
    }

    /// queue_outcome queues the messages of one statement's outcome, with rows in the formats. Without formats, as
    /// for a simple query, it describes the rows first and sends them as text.
    pub fn queue_outcome(&mut self, outcome: Outcome, formats: Option<&[i16]>) {
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
                    if let Err(err) = self.queue_data_row(&row, &columns, formats) {
                        return self.queue(BackendMessage::ErrorResponse(error_fields(&err)));
                    }
                }
                self.queue(BackendMessage::CommandComplete { command_tag: tag });
            }
            Outcome::Command { tag } | Outcome::Streamed { tag } => {
                self.queue(BackendMessage::CommandComplete { command_tag: tag })
            }
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

impl sql::RowSink for Output {
    fn outcome(&mut self, notices: Vec<PgError>, outcome: Outcome) {
        self.queue_notices(notices);
        self.queue_outcome(outcome, None);
    }

    fn begin(&mut self, notices: Vec<PgError>, columns: &[Column]) {
        self.queue_notices(notices);
        self.queue(BackendMessage::RowDescription { fields: fields(columns, &[]) });
        self.columns = columns.to_vec();
    }

    fn row(&mut self, row: &[Value]) -> sql::Result<()> {
        let columns = std::mem::take(&mut self.columns);
        let result = self.queue_data_row(row, &columns, &[]);
        self.columns = columns;
        result?;
        if self.out.len() >= BATCH {
            let _ = self.hand_off();
        }
        Ok(())
    }

    fn notices(&mut self, notices: Vec<PgError>) {
        self.queue_notices(notices);
    }

    fn into_any(self: Box<Self>) -> Box<dyn std::any::Any> {
        self
    }
}

/// error_fields converts an engine error to the fields of an ErrorResponse.
pub fn error_fields(err: &PgError) -> ErrorFields {
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
        internal_query: objects.internal_query.clone().unwrap_or_default(),
        internal_position: objects.internal_position.map_or(0, |p| p as i32),
        ..ErrorFields::default()
    }
}

/// format returns the format code of the value at the index: the only code when there is one, and text when there
/// are none.
pub fn format(codes: &[i16], index: usize) -> i16 {
    match codes {
        [] => 0,
        [code] => *code,
        codes => codes.get(index).copied().unwrap_or(0),
    }
}

/// fields describes the columns, with the result formats.
pub fn fields(columns: &[Column], formats: &[i16]) -> Vec<FieldDescription> {
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
