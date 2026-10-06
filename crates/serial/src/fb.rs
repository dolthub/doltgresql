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

//! A bounds-checked reader of flatbuffers tables. A table starts with an i32 offset back to its vtable, which lists the
//! u16 vtable length, the u16 table length, and the u16 offset of each field within the table, where 0 means the
//! field is absent. Offset fields hold a u32 forward offset to a vector (a u32 length and its elements), a string (a
//! vector of bytes with a NUL after it), or a nested table. Every integer is little-endian.

use store::Error;

/// Result is a flatbuffers read result.
pub type Result<T> = std::result::Result<T, Error>;

/// out_of_range returns the error for a read past the end of a buffer.
fn out_of_range() -> Error {
    Error::Corrupt("flatbuffer offset out of range".to_string())
}

/// read reads N bytes at the position.
fn read<const N: usize>(buf: &[u8], at: usize) -> Result<[u8; N]> {
    buf.get(at..at.checked_add(N).ok_or_else(out_of_range)?)
        .map(|bytes| bytes.try_into().unwrap())
        .ok_or_else(out_of_range)
}

/// u16_at reads a little-endian u16.
pub fn u16_at(buf: &[u8], at: usize) -> Result<u16> {
    read::<2>(buf, at).map(u16::from_le_bytes)
}

/// u32_at reads a little-endian u32.
pub fn u32_at(buf: &[u8], at: usize) -> Result<u32> {
    read::<4>(buf, at).map(u32::from_le_bytes)
}

/// Table is a flatbuffers table within a buffer.
#[derive(Clone, Copy)]
pub struct Table<'a> {
    buf: &'a [u8],
    pos: usize,
    vtable: usize,
    vtable_len: u16,
}

impl<'a> Table<'a> {
    /// at returns the table at the position.
    pub fn at(buf: &'a [u8], pos: usize) -> Result<Table<'a>> {
        let back = i32::from_le_bytes(read::<4>(buf, pos)?) as i64;
        let vtable = usize::try_from(pos as i64 - back).map_err(|_| out_of_range())?;
        let vtable_len = u16_at(buf, vtable)?;
        if vtable_len < 4 || vtable + vtable_len as usize > buf.len() {
            return Err(out_of_range());
        }
        Ok(Table { buf, pos, vtable, vtable_len })
    }

    /// root returns the root table of a flatbuffer that starts at the offset, whose first u32 points to the root.
    pub fn root(buf: &'a [u8], start: usize) -> Result<Table<'a>> {
        let offset = u32_at(buf, start)? as usize;
        Table::at(buf, start + offset)
    }

    /// buf returns the buffer holding the table.
    pub fn buf(&self) -> &'a [u8] {
        self.buf
    }

    /// field_count returns the number of fields the vtable has slots for.
    pub fn field_count(&self) -> usize {
        (self.vtable_len as usize - 4) / 2
    }

    /// offset returns the position of the field with the index in the buffer, or None when it is absent.
    pub fn offset(&self, field: usize) -> Result<Option<usize>> {
        let slot = 4 + 2 * field;
        if slot + 2 > self.vtable_len as usize {
            return Ok(None);
        }
        match u16_at(self.buf, self.vtable + slot)? {
            0 => Ok(None),
            offset => Ok(Some(self.pos + offset as usize)),
        }
    }

    /// scalar reads a scalar field, returning the default when it is absent.
    fn scalar<const N: usize>(&self, field: usize, default: [u8; N]) -> Result<[u8; N]> {
        match self.offset(field)? {
            Some(at) => read::<N>(self.buf, at),
            None => Ok(default),
        }
    }

    /// u8 reads a u8 field.
    pub fn u8(&self, field: usize, default: u8) -> Result<u8> {
        self.scalar(field, [default]).map(|b| b[0])
    }

    /// bool reads a bool field.
    pub fn bool(&self, field: usize, default: bool) -> Result<bool> {
        self.u8(field, default as u8).map(|b| b != 0)
    }

    /// i16 reads an i16 field.
    pub fn i16(&self, field: usize, default: i16) -> Result<i16> {
        self.scalar(field, default.to_le_bytes()).map(i16::from_le_bytes)
    }

    /// u16 reads a u16 field.
    pub fn u16(&self, field: usize, default: u16) -> Result<u16> {
        self.scalar(field, default.to_le_bytes()).map(u16::from_le_bytes)
    }

    /// i32 reads an i32 field.
    pub fn i32(&self, field: usize, default: i32) -> Result<i32> {
        self.scalar(field, default.to_le_bytes()).map(i32::from_le_bytes)
    }

    /// u32 reads a u32 field.
    pub fn u32(&self, field: usize, default: u32) -> Result<u32> {
        self.scalar(field, default.to_le_bytes()).map(u32::from_le_bytes)
    }

    /// i64 reads an i64 field.
    pub fn i64(&self, field: usize, default: i64) -> Result<i64> {
        self.scalar(field, default.to_le_bytes()).map(i64::from_le_bytes)
    }

    /// u64 reads a u64 field.
    pub fn u64(&self, field: usize, default: u64) -> Result<u64> {
        self.scalar(field, default.to_le_bytes()).map(u64::from_le_bytes)
    }

    /// target follows an offset field to the position it points at.
    fn target(&self, field: usize) -> Result<Option<usize>> {
        match self.offset(field)? {
            Some(at) => Ok(Some(at + u32_at(self.buf, at)? as usize)),
            None => Ok(None),
        }
    }

    /// vector returns the vector in an offset field, with its elements' width in bytes.
    pub fn vector(&self, field: usize, width: usize) -> Result<Option<Vector<'a>>> {
        match self.target(field)? {
            Some(at) => {
                let len = u32_at(self.buf, at)? as usize;
                let end =
                    (at + 4).checked_add(len.checked_mul(width).ok_or_else(out_of_range)?).ok_or_else(out_of_range)?;
                if end > self.buf.len() {
                    return Err(out_of_range());
                }
                Ok(Some(Vector { buf: self.buf, start: at + 4, len, width }))
            }
            None => Ok(None),
        }
    }

    /// bytes returns the bytes of a [ubyte] or [byte] field.
    pub fn bytes(&self, field: usize) -> Result<Option<&'a [u8]>> {
        Ok(self.vector(field, 1)?.map(|v| v.bytes()))
    }

    /// string returns the bytes of a string field, which Go does not require to be UTF-8.
    pub fn string(&self, field: usize) -> Result<Option<&'a [u8]>> {
        self.bytes(field)
    }

    /// table returns a nested table field.
    pub fn table(&self, field: usize) -> Result<Option<Table<'a>>> {
        match self.target(field)? {
            Some(at) => Table::at(self.buf, at).map(Some),
            None => Ok(None),
        }
    }
}

/// Vector is a flatbuffers vector of fixed-width elements.
#[derive(Clone, Copy)]
pub struct Vector<'a> {
    buf: &'a [u8],
    start: usize,
    len: usize,
    width: usize,
}

impl<'a> Vector<'a> {
    /// len returns the number of elements.
    pub fn len(&self) -> usize {
        self.len
    }

    /// is_empty reports whether the vector has no elements.
    pub fn is_empty(&self) -> bool {
        self.len == 0
    }

    /// start returns the position of the first element in the buffer.
    pub fn start(&self) -> usize {
        self.start
    }

    /// bytes returns the elements' bytes.
    pub fn bytes(&self) -> &'a [u8] {
        &self.buf[self.start..self.start + self.len * self.width]
    }

    /// u16 returns the u16 element at the index.
    pub fn u16(&self, index: usize) -> Result<u16> {
        u16_at(self.buf, self.start + index * 2)
    }

    /// table returns the table that the offset element at the index points to.
    pub fn table(&self, index: usize) -> Result<Table<'a>> {
        let at = self.start + index * 4;
        Table::at(self.buf, at + u32_at(self.buf, at)? as usize)
    }

    /// string returns the string that the offset element at the index points to.
    pub fn string(&self, index: usize) -> Result<&'a [u8]> {
        let at = self.start + index * 4;
        let target = at + u32_at(self.buf, at)? as usize;
        let len = u32_at(self.buf, target)? as usize;
        self.buf.get(target + 4..target + 4 + len).ok_or_else(out_of_range)
    }
}
