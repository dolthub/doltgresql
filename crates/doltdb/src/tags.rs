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

//! Column tags, which Dolt chooses deterministically so that the same changes made on two branches give the same
//! schemas.

use std::collections::HashSet;

use sha2::{Digest, Sha512};

use crate::gorand::GoRand;

/// RESERVED_TAG_MIN is the first tag Dolt reserves for its own columns.
pub const RESERVED_TAG_MIN: u64 = 1 << 50;

/// EXTENDED_KIND is the Noms kind of every Doltgres column.
pub const EXTENDED_KIND: u8 = 32;

/// STRING_KIND is the Noms kind of a MySQL string column, as Dolt's own tables have.
pub const STRING_KIND: u8 = 2;

/// simple_string lowercases the text and drops every character other than an ASCII letter or digit.
fn simple_string(text: &str) -> String {
    text.chars().filter(char::is_ascii_alphanumeric).map(|c| c.to_ascii_lowercase()).collect()
}

/// auto_generate_tag chooses the tag of a new column as Dolt's AutoGenerateTag does: a random tag outside the
/// existing ones, from a generator seeded by the kinds of the table's columns and the table and column names.
pub fn auto_generate_tag(
    existing_tags: &HashSet<u64>,
    table: &str,
    existing_kinds: &[u8],
    column: &str,
    kind: u8,
) -> u64 {
    let mut max_tag: u64 = 128 * 128;
    while max_tag / 2 < existing_tags.len() as u64 {
        if max_tag >= RESERVED_TAG_MIN - 1 {
            panic!("too many columns to choose a tag");
        } else if max_tag.wrapping_mul(128) < max_tag {
            max_tag = RESERVED_TAG_MIN - 1;
            break;
        } else {
            max_tag *= 128;
        }
    }
    let mut seed = existing_kinds.to_vec();
    seed.push(kind);
    seed.extend_from_slice(simple_string(table).as_bytes());
    seed.extend_from_slice(simple_string(column).as_bytes());
    let hash = Sha512::digest(&seed);
    let mut rng = GoRand::new(i64::from_le_bytes(hash[..8].try_into().unwrap()));
    loop {
        let tag = rng.int63n(max_tag as i64) as u64;
        if !existing_tags.contains(&tag) {
            return tag;
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn tags_match_the_ones_go_chose() {
        let mut tags = HashSet::new();
        let pk = auto_generate_tag(&tags, "test", &[], "pk", EXTENDED_KIND);
        tags.insert(pk);
        let v1 = auto_generate_tag(&tags, "test", &[EXTENDED_KIND], "v1", EXTENDED_KIND);
        assert_eq!((pk, v1), (14384, 15991));
    }
}
