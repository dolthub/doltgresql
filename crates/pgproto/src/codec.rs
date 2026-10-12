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

use std::fmt;

/// DecodeError is returned when bytes do not form a valid message.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct DecodeError {
    /// A description of what was malformed.
    pub message: String,
}

impl DecodeError {
    /// new returns a DecodeError with the given description.
    pub fn new(message: impl Into<String>) -> DecodeError {
        DecodeError { message: message.into() }
    }
}

impl fmt::Display for DecodeError {
    /// fmt implements the interface Display.
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.message)
    }
}

impl std::error::Error for DecodeError {}

/// Frame is a single typed message read from a stream: its type byte and its body without the length prefix.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Frame {
    /// The message type byte.
    pub tag: u8,
    /// The message body, which excludes the type byte and the length.
    pub body: Vec<u8>,
}

/// FrameReader accumulates bytes from a stream and splits them into frames.
#[derive(Debug, Default)]
pub struct FrameReader {
    buffer: Vec<u8>,
    start: usize,
}

impl FrameReader {
    /// new returns an empty FrameReader.
    pub fn new() -> FrameReader {
        FrameReader::default()
    }

    /// extend appends bytes that were read from the stream.
    pub fn extend(&mut self, data: &[u8]) {
        if self.start > 0 && self.start == self.buffer.len() {
            self.buffer.clear();
            self.start = 0;
        }
        self.buffer.extend_from_slice(data);
    }

    /// buffered returns the bytes that have not been consumed yet.
    pub fn buffered(&self) -> &[u8] {
        &self.buffer[self.start..]
    }

    /// next_byte returns the next unframed byte, such as the answer to an SSLRequest, or None when none is buffered.
    pub fn next_byte(&mut self) -> Option<u8> {
        let byte = *self.buffer.get(self.start)?;
        self.consume(1);
        Some(byte)
    }

    /// next_frame returns the next complete typed frame, or None when more bytes are needed.
    pub fn next_frame(&mut self) -> Result<Option<Frame>, DecodeError> {
        let available = &self.buffer[self.start..];
        if available.len() < 5 {
            return Ok(None);
        }
        let length = i32::from_be_bytes([available[1], available[2], available[3], available[4]]);
        if length < 4 {
            return Err(DecodeError::new(format!("invalid message length: {length}")));
        }
        let total = 1 + length as usize;
        if available.len() < total {
            return Ok(None);
        }
        let frame = Frame { tag: available[0], body: available[5..total].to_vec() };
        self.consume(total);
        Ok(Some(frame))
    }

    /// next_untyped_frame returns the next complete frame that has no type byte, such as a StartupMessage. The
    /// returned body excludes the length.
    pub fn next_untyped_frame(&mut self) -> Result<Option<Vec<u8>>, DecodeError> {
        let available = &self.buffer[self.start..];
        if available.len() < 4 {
            return Ok(None);
        }
        let length = i32::from_be_bytes([available[0], available[1], available[2], available[3]]);
        if length < 4 {
            return Err(DecodeError::new(format!("invalid message length: {length}")));
        }
        let total = length as usize;
        if available.len() < total {
            return Ok(None);
        }
        let body = available[4..total].to_vec();
        self.consume(total);
        Ok(Some(body))
    }

    /// consume marks the given number of buffered bytes as read.
    fn consume(&mut self, count: usize) {
        self.start += count;
        if self.start == self.buffer.len() {
            self.buffer.clear();
            self.start = 0;
        }
    }
}

/// Reader reads protocol primitives from a message body.
pub(crate) struct Reader<'a> {
    data: &'a [u8],
    position: usize,
}

impl<'a> Reader<'a> {
    /// new returns a Reader over the given body.
    pub(crate) fn new(data: &'a [u8]) -> Reader<'a> {
        Reader { data, position: 0 }
    }

    /// remaining returns the number of unread bytes.
    pub(crate) fn remaining(&self) -> usize {
        self.data.len() - self.position
    }

    /// take returns the next count bytes.
    pub(crate) fn take(&mut self, count: usize) -> Result<&'a [u8], DecodeError> {
        if self.remaining() < count {
            return Err(DecodeError::new("unexpected end of message"));
        }
        let bytes = &self.data[self.position..self.position + count];
        self.position += count;
        Ok(bytes)
    }

    /// rest returns all unread bytes.
    pub(crate) fn rest(&mut self) -> &'a [u8] {
        let bytes = &self.data[self.position..];
        self.position = self.data.len();
        bytes
    }

    /// u8 reads a single byte.
    pub(crate) fn u8(&mut self) -> Result<u8, DecodeError> {
        Ok(self.take(1)?[0])
    }

    /// i16 reads a big-endian 16-bit integer.
    pub(crate) fn i16(&mut self) -> Result<i16, DecodeError> {
        let bytes = self.take(2)?;
        Ok(i16::from_be_bytes([bytes[0], bytes[1]]))
    }

    /// u16 reads a big-endian unsigned 16-bit integer.
    pub(crate) fn u16(&mut self) -> Result<u16, DecodeError> {
        let bytes = self.take(2)?;
        Ok(u16::from_be_bytes([bytes[0], bytes[1]]))
    }

    /// i32 reads a big-endian 32-bit integer.
    pub(crate) fn i32(&mut self) -> Result<i32, DecodeError> {
        let bytes = self.take(4)?;
        Ok(i32::from_be_bytes([bytes[0], bytes[1], bytes[2], bytes[3]]))
    }

    /// u32 reads a big-endian unsigned 32-bit integer.
    pub(crate) fn u32(&mut self) -> Result<u32, DecodeError> {
        let bytes = self.take(4)?;
        Ok(u32::from_be_bytes([bytes[0], bytes[1], bytes[2], bytes[3]]))
    }

    /// cstring reads a null-terminated string.
    pub(crate) fn cstring(&mut self) -> Result<String, DecodeError> {
        let rest = &self.data[self.position..];
        let Some(end) = rest.iter().position(|&b| b == 0) else {
            return Err(DecodeError::new("string is missing its null terminator"));
        };
        let value = String::from_utf8_lossy(&rest[..end]).into_owned();
        self.position += end + 1;
        Ok(value)
    }

    /// nullable_bytes reads a length-prefixed byte string where a length of -1 means NULL.
    pub(crate) fn nullable_bytes(&mut self) -> Result<Option<Vec<u8>>, DecodeError> {
        let length = self.i32()?;
        if length == -1 {
            return Ok(None);
        }
        if length < 0 {
            return Err(DecodeError::new(format!("invalid value length: {length}")));
        }
        Ok(Some(self.take(length as usize)?.to_vec()))
    }
}

/// Writer appends protocol primitives to a buffer.
pub(crate) struct Writer<'a> {
    buffer: &'a mut Vec<u8>,
    length_at: usize,
}

impl<'a> Writer<'a> {
    /// typed starts a message with the given type byte, reserving space for its length.
    pub(crate) fn typed(buffer: &'a mut Vec<u8>, tag: u8) -> Writer<'a> {
        buffer.push(tag);
        Writer::untyped(buffer)
    }

    /// untyped starts a message that has no type byte, reserving space for its length.
    pub(crate) fn untyped(buffer: &'a mut Vec<u8>) -> Writer<'a> {
        let length_at = buffer.len();
        buffer.extend_from_slice(&[0, 0, 0, 0]);
        Writer { buffer, length_at }
    }

    /// u8 writes a single byte.
    pub(crate) fn u8(&mut self, value: u8) -> &mut Self {
        self.buffer.push(value);
        self
    }

    /// i16 writes a big-endian 16-bit integer.
    pub(crate) fn i16(&mut self, value: i16) -> &mut Self {
        self.buffer.extend_from_slice(&value.to_be_bytes());
        self
    }

    /// u16 writes a big-endian unsigned 16-bit integer.
    pub(crate) fn u16(&mut self, value: u16) -> &mut Self {
        self.buffer.extend_from_slice(&value.to_be_bytes());
        self
    }

    /// i32 writes a big-endian 32-bit integer.
    pub(crate) fn i32(&mut self, value: i32) -> &mut Self {
        self.buffer.extend_from_slice(&value.to_be_bytes());
        self
    }

    /// u32 writes a big-endian unsigned 32-bit integer.
    pub(crate) fn u32(&mut self, value: u32) -> &mut Self {
        self.buffer.extend_from_slice(&value.to_be_bytes());
        self
    }

    /// bytes writes raw bytes.
    pub(crate) fn bytes(&mut self, value: &[u8]) -> &mut Self {
        self.buffer.extend_from_slice(value);
        self
    }

    /// cstring writes a null-terminated string.
    pub(crate) fn cstring(&mut self, value: &str) -> &mut Self {
        self.buffer.extend_from_slice(value.as_bytes());
        self.buffer.push(0);
        self
    }

    /// nullable_bytes writes a length-prefixed byte string, using a length of -1 for NULL.
    pub(crate) fn nullable_bytes(&mut self, value: &Option<Vec<u8>>) -> &mut Self {
        match value {
            Some(bytes) => {
                self.i32(bytes.len() as i32);
                self.bytes(bytes)
            }
            None => self.i32(-1),
        }
    }

    /// finish fills in the message length.
    pub(crate) fn finish(self) {
        let length = (self.buffer.len() - self.length_at) as i32;
        self.buffer[self.length_at..self.length_at + 4].copy_from_slice(&length.to_be_bytes());
    }
}
