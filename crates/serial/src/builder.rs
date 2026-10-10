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

//! A flatbuffers builder that lays out bytes exactly as Dolt's fork of the Go builder does, so that messages hash the
//! same. It builds from the end of the buffer toward the front, aligning every value relative to the end.

/// VTABLE_METADATA_FIELDS is the number of u16 fields before a vtable's field offsets.
const VTABLE_METADATA_FIELDS: usize = 2;

/// MESSAGE_TYPES_KIND is the noms kind byte that starts a serialized message.
const MESSAGE_TYPES_KIND: u8 = 27;

/// Builder builds a flatbuffer back to front.
pub struct Builder {
    /// The buffer, whose data is everything from `head` to the end.
    bytes: Vec<u8>,
    head: usize,
    minalign: usize,
    /// The positions of the current object's fields, as offsets from the end, where 0 means absent.
    vtable: Vec<u32>,
    object_end: u32,
    /// The positions of the written vtables, as offsets from the end.
    vtables: Vec<u32>,
    nested: bool,
}

impl Builder {
    /// new returns a builder whose buffer starts with the capacity.
    pub fn new(capacity: usize) -> Builder {
        Builder {
            bytes: vec![0; capacity],
            head: capacity,
            minalign: 1,
            vtable: Vec::new(),
            object_end: 0,
            vtables: Vec::new(),
            nested: false,
        }
    }

    /// offset returns the size of the data, which is the position of the front of the data from the end.
    pub fn offset(&self) -> u32 {
        (self.bytes.len() - self.head) as u32
    }

    /// grow doubles the buffer, moving the data to the end.
    fn grow(&mut self) {
        let old = self.bytes.len();
        let new = if old == 0 { 1 } else { old * 2 };
        let mut bytes = vec![0; new];
        bytes[new - old..].copy_from_slice(&self.bytes);
        self.bytes = bytes;
        self.head += new - old;
    }

    /// pad places zero bytes.
    fn pad(&mut self, n: usize) {
        for _ in 0..n {
            self.place(&[0]);
        }
    }

    /// prep aligns for a value of the size after the additional bytes, growing the buffer as needed.
    pub fn prep(&mut self, size: usize, additional: usize) {
        if size > self.minalign {
            self.minalign = size;
        }
        let align = (!(self.offset() as usize + additional)).wrapping_add(1) & (size - 1);
        while self.head <= align + size + additional {
            self.grow();
        }
        self.pad(align);
    }

    /// place writes bytes in front of the data without aligning.
    fn place(&mut self, bytes: &[u8]) {
        self.head -= bytes.len();
        self.bytes[self.head..self.head + bytes.len()].copy_from_slice(bytes);
    }

    /// prepend aligns for the bytes and writes them in front of the data.
    fn prepend(&mut self, bytes: &[u8]) {
        self.prep(bytes.len(), 0);
        self.place(bytes);
    }

    pub fn prepend_u8(&mut self, value: u8) {
        self.prepend(&[value]);
    }

    pub fn prepend_u16(&mut self, value: u16) {
        self.prepend(&value.to_le_bytes());
    }

    pub fn prepend_u32(&mut self, value: u32) {
        self.prepend(&value.to_le_bytes());
    }

    pub fn prepend_u64(&mut self, value: u64) {
        self.prepend(&value.to_le_bytes());
    }

    /// prepend_offset writes the offset to a position, relative to where it is written.
    pub fn prepend_offset(&mut self, target: u32) {
        self.prep(4, 0);
        assert!(target <= self.offset(), "unreachable: off <= b.Offset()");
        let relative = self.offset() - target + 4;
        self.place(&relative.to_le_bytes());
    }

    /// start_object starts a table with the number of fields.
    pub fn start_object(&mut self, fields: usize) {
        assert!(!self.nested, "Incorrect creation order: object must not be nested.");
        self.nested = true;
        self.vtable.clear();
        self.vtable.resize(fields, 0);
        self.object_end = self.offset();
    }

    /// slot records that the field with the index is at the front of the data.
    fn slot(&mut self, field: usize) {
        self.vtable[field] = self.offset();
    }

    pub fn add_u8(&mut self, field: usize, value: u8, default: u8) {
        if value != default {
            self.prepend_u8(value);
            self.slot(field);
        }
    }

    pub fn add_bool(&mut self, field: usize, value: bool, default: bool) {
        self.add_u8(field, value as u8, default as u8);
    }

    pub fn add_u16(&mut self, field: usize, value: u16, default: u16) {
        if value != default {
            self.prepend_u16(value);
            self.slot(field);
        }
    }

    pub fn add_i16(&mut self, field: usize, value: i16, default: i16) {
        self.add_u16(field, value as u16, default as u16);
    }

    pub fn add_u32(&mut self, field: usize, value: u32, default: u32) {
        if value != default {
            self.prepend_u32(value);
            self.slot(field);
        }
    }

    pub fn add_i32(&mut self, field: usize, value: i32, default: i32) {
        self.add_u32(field, value as u32, default as u32);
    }

    pub fn add_f32(&mut self, field: usize, value: f32, default: f32) {
        if value != default {
            self.prepend_u32(value.to_bits());
            self.slot(field);
        }
    }

    pub fn add_u64(&mut self, field: usize, value: u64, default: u64) {
        if value != default {
            self.prepend_u64(value);
            self.slot(field);
        }
    }

    pub fn add_i64(&mut self, field: usize, value: i64, default: i64) {
        self.add_u64(field, value as u64, default as u64);
    }

    /// add_offset adds an offset field, where 0 means absent as in Go's generated code.
    pub fn add_offset(&mut self, field: usize, target: u32) {
        if target != 0 {
            self.prepend_offset(target);
            self.slot(field);
        }
    }

    /// end_object writes the current table's vtable, reusing an equal one already written, and returns the table.
    pub fn end_object(&mut self) -> u32 {
        assert!(self.nested, "Incorrect creation order: must be inside object.");
        self.prep(4, 0);
        self.place(&0i32.to_le_bytes());
        let object = self.offset();
        while self.vtable.last() == Some(&0) {
            self.vtable.pop();
        }
        let existing = self.vtables.iter().rev().copied().find(|&written| self.vtable_equal(written, object));
        match existing {
            None => {
                for i in (0..self.vtable.len()).rev() {
                    let field = if self.vtable[i] != 0 { object - self.vtable[i] } else { 0 };
                    self.prep(2, 0);
                    self.place(&(field as u16).to_le_bytes());
                }
                self.prep(2, 0);
                self.place(&((object - self.object_end) as u16).to_le_bytes());
                self.prep(2, 0);
                self.place(&(((self.vtable.len() + VTABLE_METADATA_FIELDS) * 2) as u16).to_le_bytes());
                let at = self.bytes.len() - object as usize;
                let back = self.offset() as i32 - object as i32;
                self.bytes[at..at + 4].copy_from_slice(&back.to_le_bytes());
                self.vtables.push(self.offset());
            }
            Some(written) => {
                self.head = self.bytes.len() - object as usize;
                let back = written as i32 - object as i32;
                self.bytes[self.head..self.head + 4].copy_from_slice(&back.to_le_bytes());
            }
        }
        self.vtable.clear();
        self.nested = false;
        object
    }

    /// vtable_equal reports whether the written vtable describes the current object's fields.
    fn vtable_equal(&self, written: u32, object: u32) -> bool {
        let start = self.bytes.len() - written as usize;
        let len = u16::from_le_bytes([self.bytes[start], self.bytes[start + 1]]) as usize;
        let fields = &self.bytes[start + VTABLE_METADATA_FIELDS * 2..start + len];
        if self.vtable.len() * 2 != fields.len() {
            return false;
        }
        self.vtable.iter().enumerate().all(|(i, &position)| {
            let x = u16::from_le_bytes([fields[i * 2], fields[i * 2 + 1]]);
            (x == 0 && position == 0) || x as i32 == object as i32 - position as i32
        })
    }

    /// start_vector starts a vector of elements of the size, aligned to the alignment.
    pub fn start_vector(&mut self, element_size: usize, count: usize, alignment: usize) -> u32 {
        assert!(!self.nested, "Incorrect creation order: object must not be nested.");
        self.nested = true;
        self.prep(4, element_size * count);
        self.prep(alignment, element_size * count);
        self.offset()
    }

    /// end_vector writes the vector's length and returns the vector.
    pub fn end_vector(&mut self, count: usize) -> u32 {
        assert!(self.nested, "Incorrect creation order: must be inside object.");
        self.place(&(count as u32).to_le_bytes());
        self.nested = false;
        self.offset()
    }

    /// create_byte_vector writes a vector of bytes.
    pub fn create_byte_vector(&mut self, bytes: &[u8]) -> u32 {
        assert!(!self.nested, "Incorrect creation order: object must not be nested.");
        self.nested = true;
        self.prep(4, bytes.len());
        self.place(bytes);
        self.end_vector(bytes.len())
    }

    /// create_string writes a NUL-terminated string.
    pub fn create_string(&mut self, bytes: &[u8]) -> u32 {
        assert!(!self.nested, "Incorrect creation order: object must not be nested.");
        self.nested = true;
        self.prep(4, bytes.len() + 1);
        self.place(&[0]);
        self.place(bytes);
        self.end_vector(bytes.len())
    }

    /// create_vector_of_tables writes a vector of offsets to tables.
    pub fn create_vector_of_tables(&mut self, tables: &[u32]) -> u32 {
        self.start_vector(4, tables.len(), 4);
        for &table in tables.iter().rev() {
            self.prepend_offset(table);
        }
        self.end_vector(tables.len())
    }

    /// finish_message finishes the root table with the file identifier and returns the message: the noms kind byte,
    /// the big-endian 24-bit size, and the flatbuffer.
    pub fn finish_message(mut self, root: u32, file_id: &str) -> Vec<u8> {
        self.prep(1, 4 + 4 + crate::PREFIX_LEN);
        self.prep(self.minalign, 4 + 4);
        for &byte in file_id.as_bytes().iter().rev() {
            self.place(&[byte]);
        }
        assert!(!self.nested, "Incorrect creation order: object must not be nested.");
        self.prep(self.minalign, 4);
        self.prepend_offset(root);
        let size = (self.bytes.len() - self.head) as u32;
        assert!(size < 1 << 24, "message is too large to be encoded");
        let mut message = Vec::with_capacity(crate::PREFIX_LEN + size as usize);
        message.push(MESSAGE_TYPES_KIND);
        message.extend_from_slice(&size.to_be_bytes()[1..]);
        message.extend_from_slice(&self.bytes[self.head..]);
        message
    }
}
