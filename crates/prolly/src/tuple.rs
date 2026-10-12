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

//! Tuples, the keys and values of index rows. A tuple is its field values packed together, then the u16 offset of
//! every field after the first, then the u16 field count, all little-endian. An empty field, or one past the count, is
//! NULL.

use store::{Error, Result};

/// Tuple is an encoded tuple.
#[derive(Clone, Copy, Debug)]
pub struct Tuple<'a>(pub &'a [u8]);

impl<'a> Tuple<'a> {
    /// count returns the number of encoded fields, which omits trailing NULLs.
    pub fn count(&self) -> Result<usize> {
        let len = self.0.len();
        if len == 0 {
            return Ok(0);
        }
        if len < 2 {
            return Err(Error::Corrupt("tuple is too short for its field count".to_string()));
        }
        Ok(u16::from_le_bytes([self.0[len - 2], self.0[len - 1]]) as usize)
    }

    /// field returns the bytes of the field at the index, or None when it is NULL.
    pub fn field(&self, index: usize) -> Result<Option<&'a [u8]>> {
        Ok(self.field_range(index)?.map(|(start, end)| &self.0[start..end]))
    }

    /// field_range returns where the field at the index lies in the tuple, or None when it is NULL.
    pub fn field_range(&self, index: usize) -> Result<Option<(usize, usize)>> {
        let count = self.count()?;
        if index >= count {
            return Ok(None);
        }
        let len = self.0.len();
        let offsets_at = len
            .checked_sub(2 + (count - 1) * 2)
            .ok_or_else(|| Error::Corrupt("tuple is too short for its offsets".to_string()))?;
        let offset = |i: usize| -> usize {
            if i == 0 {
                0
            } else {
                u16::from_le_bytes([self.0[offsets_at + (i - 1) * 2], self.0[offsets_at + (i - 1) * 2 + 1]]) as usize
            }
        };
        let start = offset(index);
        let end = if index + 1 < count { offset(index + 1) } else { offsets_at };
        if start > end || end > offsets_at {
            return Err(Error::Corrupt("tuple field offsets out of range".to_string()));
        }
        Ok(if start == end { None } else { Some((start, end)) })
    }
}
