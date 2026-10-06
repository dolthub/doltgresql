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

//! Field encodings, building tuples, and ordering keys as Dolt's val package does.

use std::cmp::Ordering;

use crate::tuple::Tuple;

/// The field encodings of Dolt's serial.Encoding.
pub mod encoding {
    pub const NULL: u8 = 0;
    pub const INT8: u8 = 1;
    pub const UINT8: u8 = 2;
    pub const INT16: u8 = 3;
    pub const UINT16: u8 = 4;
    pub const INT32: u8 = 7;
    pub const UINT32: u8 = 8;
    pub const INT64: u8 = 9;
    pub const UINT64: u8 = 10;
    pub const FLOAT32: u8 = 11;
    pub const FLOAT64: u8 = 12;
    pub const BIT64: u8 = 13;
    pub const HASH128: u8 = 14;
    pub const YEAR: u8 = 15;
    pub const DATE: u8 = 16;
    pub const TIME: u8 = 17;
    pub const DATETIME: u8 = 18;
    pub const ENUM: u8 = 19;
    pub const SET: u8 = 20;
    pub const BYTES_ADDR: u8 = 21;
    pub const COMMIT_ADDR: u8 = 22;
    pub const STRING_ADDR: u8 = 23;
    pub const JSON_ADDR: u8 = 24;
    pub const CELL: u8 = 25;
    pub const GEOM_ADDR: u8 = 26;
    pub const EXTENDED_ADDR: u8 = 27;
    pub const STRING: u8 = 128;
    pub const BYTES: u8 = 129;
    pub const DECIMAL: u8 = 130;
    pub const JSON: u8 = 131;
    pub const GEOMETRY: u8 = 133;
    pub const EXTENDED: u8 = 134;
    pub const STRING_ADAPTIVE: u8 = 135;
    pub const BYTES_ADAPTIVE: u8 = 136;
    pub const EXTENDED_ADAPTIVE: u8 = 137;
    pub const GEOM_ADAPTIVE: u8 = 138;
    pub const JSON_ADAPTIVE: u8 = 139;
}

/// build_tuple packs fields into a tuple, where None is NULL, dropping trailing NULLs as Dolt's NewTuple does.
pub fn build_tuple(fields: &[Option<&[u8]>]) -> Vec<u8> {
    let count = fields.iter().rposition(Option::is_some).map_or(0, |last| last + 1);
    if count == 0 {
        return vec![0, 0];
    }
    let mut tuple = Vec::new();
    let mut offsets = Vec::with_capacity(count);
    for field in &fields[..count] {
        offsets.push(tuple.len() as u16);
        tuple.extend_from_slice(field.unwrap_or_default());
    }
    for offset in &offsets[1..] {
        tuple.extend_from_slice(&offset.to_le_bytes());
    }
    tuple.extend_from_slice(&(count as u16).to_le_bytes());
    tuple
}

/// compare_field orders two fields of a native encoding as Dolt's default comparator does, with NULLs first.
pub fn compare_field(encoding: u8, left: Option<&[u8]>, right: Option<&[u8]>) -> Ordering {
    let (left, right) = match (left, right) {
        (None, None) => return Ordering::Equal,
        (None, Some(_)) => return Ordering::Less,
        (Some(_), None) => return Ordering::Greater,
        (Some(left), Some(right)) => (left, right),
    };
    let le = |bytes: &[u8]| {
        let mut buffer = [0; 8];
        buffer[..bytes.len().min(8)].copy_from_slice(&bytes[..bytes.len().min(8)]);
        u64::from_le_bytes(buffer)
    };
    match encoding {
        encoding::INT8 => (left[0] as i8).cmp(&(right[0] as i8)),
        encoding::INT16 => (le(left) as u16 as i16).cmp(&(le(right) as u16 as i16)),
        encoding::INT32 => (le(left) as u32 as i32).cmp(&(le(right) as u32 as i32)),
        encoding::INT64 => (le(left) as i64).cmp(&(le(right) as i64)),
        encoding::UINT8 | encoding::UINT16 | encoding::UINT32 | encoding::UINT64 | encoding::BIT64 => {
            le(left).cmp(&le(right))
        }
        encoding::FLOAT32 => {
            compare_float(f32::from_bits(le(left) as u32) as f64, f32::from_bits(le(right) as u32) as f64)
        }
        encoding::FLOAT64 => compare_float(f64::from_bits(le(left)), f64::from_bits(le(right))),
        // Strings end with a terminating byte that Dolt drops before comparing.
        encoding::STRING | encoding::BYTES => left[..left.len() - 1].cmp(&right[..right.len() - 1]),
        _ => left.cmp(right),
    }
}

/// compare_float orders floats with NaN after every other value, as Dolt does.
fn compare_float(left: f64, right: f64) -> Ordering {
    left.partial_cmp(&right).unwrap_or_else(|| left.is_nan().cmp(&right.is_nan()))
}

/// compare_tuples orders two tuples field by field with the fields' encodings.
pub fn compare_tuples(encodings: &[u8], left: &[u8], right: &[u8]) -> Ordering {
    let (left, right) = (Tuple(left), Tuple(right));
    for (i, &encoding) in encodings.iter().enumerate() {
        let ordering = compare_field(encoding, left.field(i).ok().flatten(), right.field(i).ok().flatten());
        if ordering != Ordering::Equal {
            return ordering;
        }
    }
    Ordering::Equal
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn tuples_pack_fields_then_offsets_then_count() {
        assert_eq!(build_tuple(&[Some(&1i64.to_le_bytes())]), hex("01000000000000000100"));
        assert_eq!(build_tuple(&[Some(b"ab"), None, Some(b"c"), None]), hex("616263020002000300"));
        assert_eq!(build_tuple(&[None, None]), vec![0, 0]);
    }

    #[test]
    fn keys_order_by_encoding() {
        let int = |i: i64| build_tuple(&[Some(&i.to_le_bytes())]);
        assert_eq!(compare_tuples(&[encoding::INT64], &int(-1), &int(1)), Ordering::Less);
        let float = |f: f64| build_tuple(&[Some(&f.to_le_bytes())]);
        assert_eq!(compare_tuples(&[encoding::FLOAT64], &float(f64::NAN), &float(1e300)), Ordering::Greater);
        assert_eq!(compare_tuples(&[encoding::INT64], &build_tuple(&[None]), &int(i64::MIN)), Ordering::Less);
    }

    /// hex decodes hex text.
    fn hex(text: &str) -> Vec<u8> {
        (0..text.len()).step_by(2).map(|i| u8::from_str_radix(&text[i..i + 2], 16).unwrap()).collect()
    }
}
