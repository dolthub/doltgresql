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

//! The binary encoding of Doltgres root objects. Fixed-width integers are big-endian with the sign bit flipped so that
//! they sort as bytes, variable-width integers are LEB128 (zigzag for signed ones), floats are flipped to sort as
//! bytes, and strings and slices start with their length as a variable-width integer.

use std::collections::BTreeMap;

use store::{Error, Result};

/// Reader reads encoded values in order.
pub struct Reader<'a> {
    buf: &'a [u8],
    offset: usize,
}

impl<'a> Reader<'a> {
    pub fn new(buf: &'a [u8]) -> Reader<'a> {
        Reader { buf, offset: 0 }
    }

    /// take returns the next n bytes.
    fn take(&mut self, n: usize) -> Result<&'a [u8]> {
        let end = self.offset.checked_add(n).filter(|&end| end <= self.buf.len());
        let end = end.ok_or_else(|| Error::Corrupt("root object data ends early".to_string()))?;
        let bytes = &self.buf[self.offset..end];
        self.offset = end;
        Ok(bytes)
    }

    pub fn bool(&mut self) -> Result<bool> {
        Ok(self.take(1)?[0] == 1)
    }

    pub fn uint8(&mut self) -> Result<u8> {
        Ok(self.take(1)?[0])
    }

    pub fn uint16(&mut self) -> Result<u16> {
        Ok(u16::from_be_bytes(self.take(2)?.try_into().unwrap()))
    }

    pub fn uint32(&mut self) -> Result<u32> {
        Ok(u32::from_be_bytes(self.take(4)?.try_into().unwrap()))
    }

    pub fn uint64(&mut self) -> Result<u64> {
        Ok(u64::from_be_bytes(self.take(8)?.try_into().unwrap()))
    }

    pub fn int16(&mut self) -> Result<i16> {
        Ok(self.uint16()?.wrapping_sub(1 << 15) as i16)
    }

    pub fn int32(&mut self) -> Result<i32> {
        Ok(self.uint32()?.wrapping_sub(1 << 31) as i32)
    }

    pub fn int64(&mut self) -> Result<i64> {
        Ok(self.uint64()?.wrapping_sub(1 << 63) as i64)
    }

    pub fn float32(&mut self) -> Result<f32> {
        let mut bits = self.uint32()?;
        bits ^= 0x8000_0000;
        bits ^= ((bits as i32 >> 31) as u32) & 0x7FFF_FFFF;
        Ok(f32::from_bits(bits))
    }

    /// variable_uint reads a LEB128 integer of at most ten bytes, returning all ones for a longer one as Go does.
    pub fn variable_uint(&mut self) -> Result<u64> {
        let mut value: u64 = 0;
        for i in 0..10 {
            let byte = self.uint8()? as u64;
            if byte < 0x80 {
                return Ok(value | (byte << (7 * i)));
            }
            value |= (byte & 0x7f) << (7 * i);
        }
        Ok(u64::MAX)
    }

    /// bytes reads a length-prefixed byte string.
    pub fn bytes(&mut self) -> Result<Vec<u8>> {
        let len = self.variable_uint()?;
        let len = usize::try_from(len).map_err(|_| Error::Corrupt("root object string is too long".to_string()))?;
        Ok(self.take(len)?.to_vec())
    }

    /// string reads a string, which Go does not require to be UTF-8.
    pub fn string(&mut self) -> Result<Vec<u8>> {
        self.bytes()
    }

    pub fn string_slice(&mut self) -> Result<Vec<Vec<u8>>> {
        let count = self.variable_uint()?;
        (0..count).map(|_| self.string()).collect()
    }

    pub fn string_map(&mut self) -> Result<BTreeMap<Vec<u8>, Vec<u8>>> {
        let count = self.variable_uint()?;
        let mut map = BTreeMap::new();
        for _ in 0..count {
            let key = self.string()?;
            map.insert(key, self.string()?);
        }
        Ok(map)
    }

    pub fn is_empty(&self) -> bool {
        self.offset >= self.buf.len()
    }
}

/// Writer writes encoded values in order.
#[derive(Default)]
pub struct Writer {
    buf: Vec<u8>,
}

impl Writer {
    pub fn new() -> Writer {
        Writer::default()
    }

    pub fn data(self) -> Vec<u8> {
        self.buf
    }

    pub fn bool(&mut self, value: bool) {
        self.buf.push(value as u8);
    }

    pub fn uint8(&mut self, value: u8) {
        self.buf.push(value);
    }

    pub fn uint16(&mut self, value: u16) {
        self.buf.extend_from_slice(&value.to_be_bytes());
    }

    pub fn uint32(&mut self, value: u32) {
        self.buf.extend_from_slice(&value.to_be_bytes());
    }

    pub fn uint64(&mut self, value: u64) {
        self.buf.extend_from_slice(&value.to_be_bytes());
    }

    pub fn int16(&mut self, value: i16) {
        self.uint16((value as u16).wrapping_add(1 << 15));
    }

    pub fn int32(&mut self, value: i32) {
        self.uint32((value as u32).wrapping_add(1 << 31));
    }

    pub fn int64(&mut self, value: i64) {
        self.uint64((value as u64).wrapping_add(1 << 63));
    }

    pub fn float32(&mut self, value: f32) {
        let mut bits = value.to_bits();
        bits ^= ((bits as i32 >> 31) as u32) & 0x7FFF_FFFF;
        bits ^= 0x8000_0000;
        self.uint32(bits);
    }

    pub fn variable_uint(&mut self, mut value: u64) {
        while value >= 0x80 {
            self.buf.push(value as u8 | 0x80);
            value >>= 7;
        }
        self.buf.push(value as u8);
    }

    pub fn string(&mut self, value: &[u8]) {
        self.variable_uint(value.len() as u64);
        self.buf.extend_from_slice(value);
    }

    pub fn string_slice(&mut self, values: &[Vec<u8>]) {
        self.variable_uint(values.len() as u64);
        for value in values {
            self.string(value);
        }
    }

    pub fn string_map(&mut self, map: &BTreeMap<Vec<u8>, Vec<u8>>) {
        self.variable_uint(map.len() as u64);
        for (key, value) in map {
            self.string(key);
            self.string(value);
        }
    }
}
