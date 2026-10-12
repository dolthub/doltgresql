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

use std::collections::BTreeMap;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

use pgproto::{BackendMessage, ErrorFields, FrameReader, FrontendMessage};

use crate::pgx::error::Error;

/// READ_TIMEOUT bounds how long the server may take to answer a request, so that a hung or endlessly streaming
/// server fails the test instead of hanging it.
const READ_TIMEOUT: Duration = Duration::from_secs(120);

/// SharedBuffer is a byte buffer shared between a connection and its recorder.
type SharedBuffer = Arc<Mutex<Vec<u8>>>;

/// Recorder collects the raw bytes that clients send, one entry per connection, in connection order.
#[derive(Clone, Debug, Default)]
pub struct Recorder {
    connections: Arc<Mutex<Vec<SharedBuffer>>>,
}

impl Recorder {
    /// new returns an empty Recorder.
    pub fn new() -> Recorder {
        Recorder::default()
    }

    /// connection registers a new connection and returns the buffer that its sent bytes are appended to.
    fn connection(&self) -> SharedBuffer {
        let buffer = Arc::new(Mutex::new(Vec::new()));
        self.connections.lock().unwrap().push(buffer.clone());
        buffer
    }

    /// take returns the bytes sent on every connection so far, and clears the recorder.
    pub fn take(&self) -> Vec<Vec<u8>> {
        let connections = std::mem::take(&mut *self.connections.lock().unwrap());
        connections.into_iter().map(|buffer| std::mem::take(&mut *buffer.lock().unwrap())).collect()
    }
}

/// Notification is a NOTIFY payload delivered to the connection.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Notification {
    /// The process ID of the notifying backend.
    pub process_id: u32,
    /// The channel name.
    pub channel: String,
    /// The payload.
    pub payload: String,
}

/// Socket is a plaintext or TLS connection.
enum Socket {
    Plain(TcpStream),
    Tls(Box<rustls::StreamOwned<rustls::ClientConnection, TcpStream>>),
}

impl Socket {
    fn tcp(&self) -> &TcpStream {
        match self {
            Socket::Plain(tcp) => tcp,
            Socket::Tls(tls) => tls.get_ref(),
        }
    }

    fn read(&mut self, buffer: &mut [u8]) -> std::io::Result<usize> {
        match self {
            Socket::Plain(tcp) => tcp.read(buffer),
            Socket::Tls(tls) => tls.read(buffer),
        }
    }

    fn write_all(&mut self, bytes: &[u8]) -> std::io::Result<()> {
        match self {
            Socket::Plain(tcp) => tcp.write_all(bytes),
            Socket::Tls(tls) => tls.write_all(bytes).and_then(|_| tls.flush()),
        }
    }
}

/// Stream sends and receives protocol messages over a socket, tracking the state that pgx tracks for every message.
pub(crate) struct Stream {
    socket: Socket,
    reader: FrameReader,
    write_buffer: Vec<u8>,
    recording: Option<SharedBuffer>,
    deadline: Instant,
    /// The transaction status from the most recent ReadyForQuery.
    pub(crate) tx_status: u8,
    /// The most recent value of every runtime parameter that the server reported.
    pub(crate) parameter_statuses: BTreeMap<String, String>,
    /// The notices received since they were last taken.
    pub(crate) notices: Vec<ErrorFields>,
    /// The notifications received since they were last taken.
    pub(crate) notifications: Vec<Notification>,
}

impl Stream {
    /// connect opens a socket to the address, recording sent bytes when a recorder is given.
    pub(crate) fn connect(address: &str, recorder: Option<&Recorder>) -> Result<Stream, std::io::Error> {
        let socket = TcpStream::connect(address)?;
        socket.set_nodelay(true)?;
        socket.set_write_timeout(Some(READ_TIMEOUT))?;
        Ok(Stream {
            socket: Socket::Plain(socket),
            reader: FrameReader::new(),
            write_buffer: Vec::new(),
            recording: recorder.map(Recorder::connection),
            deadline: Instant::now() + READ_TIMEOUT,
            tx_status: 0,
            parameter_statuses: BTreeMap::new(),
            notices: Vec::new(),
            notifications: Vec::new(),
        })
    }

    /// send queues a message to be written by the next flush.
    pub(crate) fn send(&mut self, message: &FrontendMessage) {
        message.encode(&mut self.write_buffer);
    }

    /// flush writes every queued message.
    pub(crate) fn flush(&mut self) -> Result<(), Error> {
        if self.write_buffer.is_empty() {
            return Ok(());
        }
        if let Some(recording) = &self.recording {
            recording.lock().unwrap().extend_from_slice(&self.write_buffer);
        }
        let result = self.socket.write_all(&self.write_buffer);
        self.write_buffer.clear();
        self.deadline = Instant::now() + READ_TIMEOUT;
        result?;
        Ok(())
    }

    /// write_raw writes bytes that are not a typed message, such as an SSLRequest, without queueing them.
    pub(crate) fn write_raw(&mut self, bytes: &[u8]) -> Result<(), Error> {
        if let Some(recording) = &self.recording {
            recording.lock().unwrap().extend_from_slice(bytes);
        }
        self.socket.write_all(bytes)?;
        self.deadline = Instant::now() + READ_TIMEOUT;
        Ok(())
    }

    /// read_byte reads a single unframed byte, which is how the server answers an SSLRequest.
    pub(crate) fn read_byte(&mut self) -> Result<u8, Error> {
        if let Some(byte) = self.reader.next_byte() {
            return Ok(byte);
        }
        let mut byte = [0u8; 1];
        self.socket.tcp().set_read_timeout(Some(READ_TIMEOUT))?;
        let mut read = 0;
        while read == 0 {
            read = self.socket.read(&mut byte)?;
            if read == 0 {
                return Err(Error::Other("unexpected EOF".to_string()));
            }
        }
        Ok(byte[0])
    }

    /// recv returns the next message from the server. Like pgx, it records ReadyForQuery and ParameterStatus state
    /// and collects notices and notifications before returning the message.
    pub(crate) fn recv(&mut self) -> Result<BackendMessage, Error> {
        self.recv_until(self.deadline)?
            .ok_or_else(|| Error::Other(format!("timed out after {READ_TIMEOUT:?} waiting for the server")))
    }

    /// recv_until returns the next message from the server as recv does, or None when none arrives by the deadline.
    pub(crate) fn recv_until(&mut self, deadline: Instant) -> Result<Option<BackendMessage>, Error> {
        let frame = loop {
            if let Some(frame) = self.reader.next_frame()? {
                break frame;
            }
            let remaining = deadline.saturating_duration_since(Instant::now());
            if remaining.is_zero() {
                return Ok(None);
            }
            self.socket.tcp().set_read_timeout(Some(remaining))?;
            let mut buffer = [0u8; 16384];
            let count = match self.socket.read(&mut buffer) {
                Ok(count) => count,
                Err(err) if matches!(err.kind(), std::io::ErrorKind::WouldBlock | std::io::ErrorKind::TimedOut) => {
                    return Ok(None);
                }
                Err(err) => return Err(err.into()),
            };
            if count == 0 {
                return Err(Error::Other("unexpected EOF".to_string()));
            }
            self.reader.extend(&buffer[..count]);
        };
        let message = BackendMessage::decode(frame.tag, &frame.body)?;
        match &message {
            BackendMessage::ReadyForQuery { tx_status } => self.tx_status = *tx_status,
            BackendMessage::ParameterStatus { name, value } => {
                self.parameter_statuses.insert(name.clone(), value.clone());
            }
            BackendMessage::NoticeResponse(fields) => self.notices.push(fields.clone()),
            BackendMessage::NotificationResponse { process_id, channel, payload } => {
                self.notifications.push(Notification {
                    process_id: *process_id,
                    channel: channel.clone(),
                    payload: payload.clone(),
                });
            }
            _ => {}
        }
        Ok(Some(message))
    }

    /// shutdown closes the socket.
    pub(crate) fn shutdown(&mut self) {
        let _ = self.socket.tcp().shutdown(std::net::Shutdown::Both);
    }

    /// start_tls performs a TLS handshake on the socket, which must be plaintext with nothing buffered.
    pub(crate) fn start_tls(
        &mut self,
        config: std::sync::Arc<rustls::ClientConfig>,
        server_name: rustls::pki_types::ServerName<'static>,
    ) -> Result<(), Error> {
        let Socket::Plain(tcp) = &self.socket else {
            return Err(Error::Other("the connection already uses TLS".to_string()));
        };
        let mut tcp = tcp.try_clone()?;
        tcp.set_read_timeout(Some(READ_TIMEOUT))?;
        let mut connection =
            rustls::ClientConnection::new(config, server_name).map_err(|e| Error::Other(e.to_string()))?;
        while connection.is_handshaking() {
            connection.complete_io(&mut tcp)?;
        }
        self.socket = Socket::Tls(Box::new(rustls::StreamOwned::new(connection, tcp)));
        Ok(())
    }
}
