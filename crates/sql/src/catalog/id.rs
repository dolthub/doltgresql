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

//! Doltgres' internal object IDs: a section byte followed by length-prefixed name segments.

/// SECTION_DATABASE is the ID section of databases.
pub const SECTION_DATABASE: u8 = 6;
/// SECTION_FOREIGN_KEY is the ID section of foreign keys.
pub const SECTION_FOREIGN_KEY: u8 = 11;
/// SECTION_INDEX is the ID section of indexes.
pub const SECTION_INDEX: u8 = 17;
/// SECTION_NAMESPACE is the ID section of schemas.
pub const SECTION_NAMESPACE: u8 = 18;
/// SECTION_SEQUENCE is the ID section of sequences.
pub const SECTION_SEQUENCE: u8 = 27;
/// SECTION_TABLE is the ID section of tables.
pub const SECTION_TABLE: u8 = 29;
/// SECTION_TYPE is the ID section of types.
pub const SECTION_TYPE: u8 = 35;
/// SECTION_USER is the ID section of roles.
pub const SECTION_USER: u8 = 37;
/// SECTION_VIEW is the ID section of views.
pub const SECTION_VIEW: u8 = 38;

/// FORMAT_MASK marks an ID whose segments are separated by NULs because one is too long for a length byte.
const FORMAT_MASK: u8 = 0x80;

/// new returns the ID of an object in a section with name segments, as Go's id.NewId writes it.
pub fn new(section: u8, segments: &[&str]) -> Vec<u8> {
    if segments.iter().all(|s| s.is_empty()) {
        return Vec::new();
    }
    if segments.len() > 255 || segments.iter().any(|s| s.len() > 255) {
        let mut id = vec![section | FORMAT_MASK];
        id.extend(segments.join("\0").into_bytes());
        return id;
    }
    let mut id = vec![section, segments.len() as u8];
    id.extend(segments.iter().map(|s| s.len() as u8));
    for segment in segments {
        id.extend_from_slice(segment.as_bytes());
    }
    id
}

/// segments returns the name segments of an ID.
pub fn segments(id: &[u8]) -> Vec<String> {
    if id.len() <= 1 {
        return Vec::new();
    }
    if id[0] & FORMAT_MASK != 0 {
        return String::from_utf8_lossy(&id[1..]).split('\0').map(str::to_string).collect();
    }
    let count = id[1] as usize;
    let mut start = 2 + count;
    let mut out = Vec::with_capacity(count);
    for i in 0..count {
        let len = id.get(2 + i).copied().unwrap_or(0) as usize;
        let end = (start + len).min(id.len());
        out.push(String::from_utf8_lossy(&id[start.min(end)..end]).into_owned());
        start = end;
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn ids_encode_as_go_writes_them() {
        let id = new(SECTION_SEQUENCE, &["public", "t_id_seq"]);
        assert_eq!(id, b"\x1b\x02\x06\x08publict_id_seq");
        assert_eq!(segments(&id), vec!["public", "t_id_seq"]);
        assert_eq!(new(SECTION_TYPE, &["pg_catalog", "int4"]), b"\x23\x02\x0a\x04pg_catalogint4".to_vec());
        assert_eq!(new(SECTION_TABLE, &["", ""]), Vec::<u8>::new());
    }
}
