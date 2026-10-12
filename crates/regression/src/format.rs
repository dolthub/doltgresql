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

/// Reader reads values in order from a byte buffer.
pub struct Reader<'a> {
    buf: &'a [u8],
    offset: usize,
}

impl<'a> Reader<'a> {
    /// new returns a reader at the start of the buffer.
    pub fn new(buf: &'a [u8]) -> Reader<'a> {
        Reader { buf, offset: 0 }
    }

    /// is_empty returns whether every byte has been read.
    pub fn is_empty(&self) -> bool {
        self.offset >= self.buf.len()
    }

    fn take(&mut self, count: usize) -> Result<&'a [u8], String> {
        let end = self.offset.checked_add(count).filter(|end| *end <= self.buf.len());
        let end = end.ok_or_else(|| format!("unexpected end of data at offset {}", self.offset))?;
        let bytes = &self.buf[self.offset..end];
        self.offset = end;
        Ok(bytes)
    }

    /// u16 reads a big-endian u16.
    pub fn u16(&mut self) -> Result<u16, String> {
        Ok(u16::from_be_bytes(self.take(2)?.try_into().unwrap()))
    }

    /// u32 reads a big-endian u32.
    pub fn u32(&mut self) -> Result<u32, String> {
        Ok(u32::from_be_bytes(self.take(4)?.try_into().unwrap()))
    }

    /// varint reads an unsigned LEB128 integer.
    pub fn varint(&mut self) -> Result<u64, String> {
        let mut value = 0u64;
        for shift in (0..64).step_by(7) {
            let byte = self.take(1)?[0];
            value |= u64::from(byte & 0x7f) << shift;
            if byte < 0x80 {
                return Ok(value);
            }
        }
        Err("varint is too long".to_string())
    }

    /// bytes reads a varint length followed by that many bytes.
    pub fn bytes(&mut self) -> Result<&'a [u8], String> {
        let length = self.varint()?;
        self.take(usize::try_from(length).map_err(|e| e.to_string())?)
    }

    /// string reads a varint length followed by that many bytes of text.
    pub fn string(&mut self) -> Result<String, String> {
        Ok(String::from_utf8_lossy(self.bytes()?).into_owned())
    }

    /// strings reads a varint count followed by that many strings.
    pub fn strings(&mut self) -> Result<Vec<String>, String> {
        (0..self.varint()?).map(|_| self.string()).collect()
    }
}

/// Writer appends values to a byte buffer, mirroring Reader.
#[derive(Default)]
pub struct Writer {
    buf: Vec<u8>,
}

impl Writer {
    /// u16 writes a big-endian u16.
    pub fn u16(&mut self, value: u16) {
        self.buf.extend_from_slice(&value.to_be_bytes());
    }

    /// u32 writes a big-endian u32.
    pub fn u32(&mut self, value: u32) {
        self.buf.extend_from_slice(&value.to_be_bytes());
    }

    /// varint writes an unsigned LEB128 integer.
    pub fn varint(&mut self, mut value: u64) {
        while value >= 0x80 {
            self.buf.push((value as u8) | 0x80);
            value >>= 7;
        }
        self.buf.push(value as u8);
    }

    /// bytes writes a varint length followed by the bytes.
    pub fn bytes(&mut self, value: &[u8]) {
        self.varint(value.len() as u64);
        self.buf.extend_from_slice(value);
    }

    /// string writes a varint length followed by the text.
    pub fn string(&mut self, value: &str) {
        self.bytes(value.as_bytes());
    }

    /// strings writes a varint count followed by each string.
    pub fn strings(&mut self, values: &[String]) {
        self.varint(values.len() as u64);
        for value in values {
            self.string(value);
        }
    }

    /// finish returns the written bytes.
    pub fn finish(self) -> Vec<u8> {
        self.buf
    }
}
