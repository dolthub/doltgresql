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

/// Reader reads big-endian primitives from a binary value.
pub(crate) struct Reader<'a> {
    data: &'a [u8],
    position: usize,
}

impl<'a> Reader<'a> {
    /// new returns a Reader over the bytes.
    pub(crate) fn new(data: &'a [u8]) -> Reader<'a> {
        Reader { data, position: 0 }
    }

    /// take returns the next count bytes.
    pub(crate) fn take(&mut self, count: usize) -> Result<&'a [u8], String> {
        if self.data.len() - self.position < count {
            return Err(format!("value ended after {} bytes", self.data.len()));
        }
        let bytes = &self.data[self.position..self.position + count];
        self.position += count;
        Ok(bytes)
    }

    /// rest returns every unread byte.
    pub(crate) fn rest(&mut self) -> &'a [u8] {
        let bytes = &self.data[self.position..];
        self.position = self.data.len();
        bytes
    }

    /// is_empty reports whether every byte was read.
    pub(crate) fn is_empty(&self) -> bool {
        self.position == self.data.len()
    }

    /// expect_end returns an error when bytes remain unread.
    pub(crate) fn expect_end(&self, oid: u32) -> Result<(), String> {
        if !self.is_empty() {
            return Err(format!("type {oid} has {} trailing bytes", self.data.len() - self.position));
        }
        Ok(())
    }

    /// u8 reads one byte.
    pub(crate) fn u8(&mut self) -> Result<u8, String> {
        Ok(self.take(1)?[0])
    }

    /// i16 reads a 16-bit integer.
    pub(crate) fn i16(&mut self) -> Result<i16, String> {
        Ok(i16::from_be_bytes(self.take(2)?.try_into().unwrap()))
    }

    /// u16 reads an unsigned 16-bit integer.
    pub(crate) fn u16(&mut self) -> Result<u16, String> {
        Ok(u16::from_be_bytes(self.take(2)?.try_into().unwrap()))
    }

    /// i32 reads a 32-bit integer.
    pub(crate) fn i32(&mut self) -> Result<i32, String> {
        Ok(i32::from_be_bytes(self.take(4)?.try_into().unwrap()))
    }

    /// u32 reads an unsigned 32-bit integer.
    pub(crate) fn u32(&mut self) -> Result<u32, String> {
        Ok(u32::from_be_bytes(self.take(4)?.try_into().unwrap()))
    }

    /// i64 reads a 64-bit integer.
    pub(crate) fn i64(&mut self) -> Result<i64, String> {
        Ok(i64::from_be_bytes(self.take(8)?.try_into().unwrap()))
    }

    /// u64 reads an unsigned 64-bit integer.
    pub(crate) fn u64(&mut self) -> Result<u64, String> {
        Ok(u64::from_be_bytes(self.take(8)?.try_into().unwrap()))
    }

    /// f64 reads a 64-bit float.
    pub(crate) fn f64(&mut self) -> Result<f64, String> {
        Ok(f64::from_bits(self.u64()?))
    }

    /// length_prefixed reads a 32-bit length followed by that many bytes, where -1 is NULL.
    pub(crate) fn length_prefixed(&mut self) -> Result<Option<&'a [u8]>, String> {
        let length = self.i32()?;
        if length == -1 {
            return Ok(None);
        }
        if length < 0 {
            return Err(format!("invalid length {length}"));
        }
        Ok(Some(self.take(length as usize)?))
    }
}
