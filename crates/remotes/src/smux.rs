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

//! Version 1 of the smux protocol that Dolt's ssh remotes multiplex their gRPC and HTTP streams with over a single
//! pipe, as github.com/xtaci/smux speaks it.

use std::collections::HashMap;
use std::io;
use std::pin::Pin;
use std::sync::atomic::{AtomicU32, Ordering};
use std::sync::{Arc, Mutex, Weak};
use std::task::{Context, Poll};
use std::time::Duration;

use tokio::io::{AsyncRead, AsyncReadExt, AsyncWrite, AsyncWriteExt, ReadBuf};
use tokio::sync::mpsc::{UnboundedReceiver, UnboundedSender, unbounded_channel};

/// VERSION is the protocol version that every frame carries.
const VERSION: u8 = 1;

/// SYN opens a stream.
const SYN: u8 = 0;

/// FIN closes a stream.
const FIN: u8 = 1;

/// PSH carries a stream's data.
const PSH: u8 = 2;

/// NOP keeps the session alive.
const NOP: u8 = 3;

/// MAX_FRAME is the most data one frame carries, as smux's default MaxFrameSize is.
const MAX_FRAME: usize = 32768;

/// KEEPALIVE is how often the session sends a NOP, which the other side expects at least every 30 seconds.
const KEEPALIVE: Duration = Duration::from_secs(10);

/// Session multiplexes streams over one reader and writer.
#[derive(Clone)]
pub struct Session {
    inner: Arc<Inner>,
}

/// Inner is the state that a session and its streams share.
struct Inner {
    /// The encoded frames that the writer task writes in order, where an empty frame stops it.
    frames: UnboundedSender<Vec<u8>>,
    /// The open streams' data senders by stream ID, which drop when the other side closes them.
    streams: Mutex<HashMap<u32, UnboundedSender<Vec<u8>>>>,
    /// The last stream ID that the session opened.
    next: AtomicU32,
}

/// frame encodes a frame.
fn frame(cmd: u8, sid: u32, data: &[u8]) -> Vec<u8> {
    let mut bytes = Vec::with_capacity(8 + data.len());
    bytes.extend_from_slice(&[VERSION, cmd]);
    bytes.extend_from_slice(&(data.len() as u16).to_le_bytes());
    bytes.extend_from_slice(&sid.to_le_bytes());
    bytes.extend_from_slice(data);
    bytes
}

impl Session {
    /// new starts a client session over a reader and writer on the current runtime.
    pub fn new<R, W>(reader: R, mut writer: W) -> Session
    where
        R: AsyncRead + Unpin + Send + 'static,
        W: AsyncWrite + Unpin + Send + 'static,
    {
        let (frames, mut queued) = unbounded_channel::<Vec<u8>>();
        let inner = Arc::new(Inner { frames, streams: Mutex::new(HashMap::new()), next: AtomicU32::new(1) });
        tokio::spawn(async move {
            while let Some(frame) = queued.recv().await {
                if frame.is_empty() || writer.write_all(&frame).await.is_err() || writer.flush().await.is_err() {
                    break;
                }
            }
            let _ = writer.shutdown().await;
        });
        tokio::spawn(receive(reader, Arc::downgrade(&inner)));
        let weak = Arc::downgrade(&inner);
        tokio::spawn(async move {
            let mut ticks = tokio::time::interval(KEEPALIVE);
            ticks.tick().await;
            loop {
                ticks.tick().await;
                match weak.upgrade() {
                    Some(inner) if inner.frames.send(frame(NOP, 0, &[])).is_ok() => {}
                    _ => break,
                }
            }
        });
        Session { inner }
    }

    /// open opens a stream.
    pub fn open(&self) -> io::Result<Stream> {
        let sid = self.inner.next.fetch_add(2, Ordering::Relaxed) + 2;
        let stream = Stream::new(sid, &self.inner)?;
        self.inner.frames.send(frame(SYN, sid, &[])).map_err(|_| closed())?;
        Ok(stream)
    }

    /// close stops writing, which closes the writer.
    pub fn close(&self) {
        let _ = self.inner.frames.send(Vec::new());
    }
}

/// closed returns the error of a session that has closed.
fn closed() -> io::Error {
    io::Error::new(io::ErrorKind::BrokenPipe, "smux session closed")
}

/// receive reads frames and hands each stream its data until the reader ends, which ends every stream.
async fn receive<R: AsyncRead + Unpin>(mut reader: R, inner: Weak<Inner>) {
    let mut header = [0u8; 8];
    while reader.read_exact(&mut header).await.is_ok() && header[0] == VERSION {
        let length = u16::from_le_bytes([header[2], header[3]]) as usize;
        let sid = u32::from_le_bytes([header[4], header[5], header[6], header[7]]);
        let mut data = vec![0; length];
        if reader.read_exact(&mut data).await.is_err() {
            break;
        }
        let Some(inner) = inner.upgrade() else { break };
        let mut streams = inner.streams.lock().unwrap_or_else(|e| e.into_inner());
        match header[1] {
            FIN => {
                streams.remove(&sid);
            }
            PSH if !data.is_empty() => {
                if let Some(sender) = streams.get(&sid) {
                    let _ = sender.send(data);
                }
            }
            SYN | PSH | NOP => {}
            _ => break,
        }
    }
    if let Some(inner) = inner.upgrade() {
        inner.streams.lock().unwrap_or_else(|e| e.into_inner()).clear();
    }
}

/// Stream is one of a session's streams.
pub struct Stream {
    sid: u32,
    inner: Arc<Inner>,
    data: UnboundedReceiver<Vec<u8>>,
    /// The data received but not yet read, from `at` on.
    buffer: Vec<u8>,
    at: usize,
    finished: bool,
}

impl Stream {
    /// new registers a stream with a session.
    fn new(sid: u32, inner: &Arc<Inner>) -> io::Result<Stream> {
        let (sender, data) = unbounded_channel();
        inner.streams.lock().map_err(|_| closed())?.insert(sid, sender);
        Ok(Stream { sid, inner: inner.clone(), data, buffer: Vec::new(), at: 0, finished: false })
    }

    /// finish tells the other side that the stream is closed, once.
    fn finish(&mut self) {
        if !self.finished {
            self.finished = true;
            let _ = self.inner.frames.send(frame(FIN, self.sid, &[]));
        }
    }
}

impl AsyncRead for Stream {
    fn poll_read(mut self: Pin<&mut Self>, cx: &mut Context<'_>, buf: &mut ReadBuf<'_>) -> Poll<io::Result<()>> {
        if self.at == self.buffer.len() {
            match self.data.poll_recv(cx) {
                Poll::Ready(Some(data)) => {
                    self.buffer = data;
                    self.at = 0;
                }
                Poll::Ready(None) => return Poll::Ready(Ok(())),
                Poll::Pending => return Poll::Pending,
            }
        }
        let count = buf.remaining().min(self.buffer.len() - self.at);
        buf.put_slice(&self.buffer[self.at..self.at + count]);
        self.at += count;
        Poll::Ready(Ok(()))
    }
}

impl AsyncWrite for Stream {
    fn poll_write(self: Pin<&mut Self>, _: &mut Context<'_>, buf: &[u8]) -> Poll<io::Result<usize>> {
        if self.finished {
            return Poll::Ready(Err(closed()));
        }
        for data in buf.chunks(MAX_FRAME) {
            if self.inner.frames.send(frame(PSH, self.sid, data)).is_err() {
                return Poll::Ready(Err(closed()));
            }
        }
        Poll::Ready(Ok(buf.len()))
    }

    fn poll_flush(self: Pin<&mut Self>, _: &mut Context<'_>) -> Poll<io::Result<()>> {
        Poll::Ready(Ok(()))
    }

    fn poll_shutdown(mut self: Pin<&mut Self>, _: &mut Context<'_>) -> Poll<io::Result<()>> {
        self.finish();
        Poll::Ready(Ok(()))
    }
}

impl Drop for Stream {
    fn drop(&mut self) {
        self.finish();
        self.inner.streams.lock().unwrap_or_else(|e| e.into_inner()).remove(&self.sid);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// read_frame reads a frame's version, command, stream ID, and data.
    async fn read_frame(reader: &mut (impl AsyncRead + Unpin)) -> (u8, u8, u32, Vec<u8>) {
        let mut header = [0u8; 8];
        reader.read_exact(&mut header).await.unwrap();
        let mut data = vec![0; u16::from_le_bytes([header[2], header[3]]) as usize];
        reader.read_exact(&mut data).await.unwrap();
        (header[0], header[1], u32::from_le_bytes([header[4], header[5], header[6], header[7]]), data)
    }

    #[test]
    fn streams_speak_smux_version_one() {
        tokio::runtime::Builder::new_current_thread().enable_all().build().unwrap().block_on(exchange());
    }

    /// exchange checks the frames that two streams send and receive.
    async fn exchange() {
        let (local, remote) = tokio::io::duplex(1 << 20);
        let (reader, writer) = tokio::io::split(local);
        let session = Session::new(reader, writer);
        let (mut peer_reader, mut peer_writer) = tokio::io::split(remote);
        let mut first = session.open().unwrap();
        let mut second = session.open().unwrap();
        let payload: Vec<u8> = (0..MAX_FRAME + 10).map(|i| i as u8).collect();
        first.write_all(&payload).await.unwrap();
        second.shutdown().await.unwrap();
        assert_eq!(read_frame(&mut peer_reader).await, (1, SYN, 3, vec![]));
        assert_eq!(read_frame(&mut peer_reader).await, (1, SYN, 5, vec![]));
        assert_eq!(read_frame(&mut peer_reader).await, (1, PSH, 3, payload[..MAX_FRAME].to_vec()));
        assert_eq!(read_frame(&mut peer_reader).await, (1, PSH, 3, payload[MAX_FRAME..].to_vec()));
        assert_eq!(read_frame(&mut peer_reader).await, (1, FIN, 5, vec![]));
        for (cmd, sid, data) in
            [(NOP, 0, &b""[..]), (PSH, 3, b"hello, "), (PSH, 7, b"lost"), (PSH, 3, b"world"), (FIN, 3, b"")]
        {
            peer_writer.write_all(&frame(cmd, sid, data)).await.unwrap();
        }
        let mut received = String::new();
        first.read_to_string(&mut received).await.unwrap();
        assert_eq!(received, "hello, world");
        drop(first);
        assert_eq!(read_frame(&mut peer_reader).await, (1, FIN, 3, vec![]));
        peer_writer.shutdown().await.unwrap();
        let mut rest = Vec::new();
        assert_eq!(second.read_to_end(&mut rest).await.unwrap(), 0);
    }
}
