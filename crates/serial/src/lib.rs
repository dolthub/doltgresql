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

#![forbid(unsafe_code)]

//! Dolt's serialized messages. A message is a flatbuffer after a 4-byte prefix of the kind byte and a big-endian
//! 24-bit size, with a 4-byte file identifier after the flatbuffer's root offset naming its type.

pub mod fb;
mod messages;

pub use fb::{Result, Table, Vector};
pub use messages::*;

/// PREFIX_LEN is the length of the prefix before a message's flatbuffer.
pub const PREFIX_LEN: usize = 4;

/// Message is a serialized message.
#[derive(Clone, Copy)]
pub struct Message<'a>(pub &'a [u8]);

impl<'a> Message<'a> {
    /// file_id returns the message's file identifier, or an empty string when it is too short to have one.
    pub fn file_id(&self) -> &'a str {
        self.0.get(PREFIX_LEN + 4..PREFIX_LEN + 8).and_then(|id| std::str::from_utf8(id).ok()).unwrap_or("")
    }

    /// root returns the message's root table.
    pub fn root(&self) -> Result<Table<'a>> {
        Table::root(self.0, PREFIX_LEN)
    }

    /// expect returns the root table, failing unless the message has the file identifier.
    pub fn expect(&self, file_id: &str) -> Result<Table<'a>> {
        if self.file_id() != file_id {
            return Err(store::Error::Corrupt(format!("expected a {file_id} message, got {:?}", self.file_id())));
        }
        self.root()
    }
}

/// File identifiers of the message types.
pub const STORE_ROOT: &str = "STRT";
pub const TAG: &str = "DTAG";
pub const WORKING_SET: &str = "WRST";
pub const COMMIT: &str = "DCMT";
pub const ROOT_VALUE: &str = "RTVL";
pub const TABLE: &str = "DTBL";
pub const PROLLY_TREE_NODE: &str = "TUPM";
pub const ADDRESS_MAP: &str = "ADRM";
pub const COMMIT_CLOSURE: &str = "CMCL";
pub const TABLE_SCHEMA: &str = "DSCH";
pub const FOREIGN_KEY_COLLECTION: &str = "DFKC";
pub const MERGE_ARTIFACTS: &str = "ARTM";
pub const BLOB: &str = "BLOB";
pub const BRANCH_CONTROL: &str = "BRCL";
pub const STASH_LIST: &str = "SLST";
pub const STASH: &str = "STSH";
pub const STATISTIC: &str = "STAT";
pub const DOLTGRES_ROOT_VALUE: &str = "DGRV";
pub const TUPLE: &str = "TUPL";
pub const VECTOR_INDEX_NODE: &str = "IVFF";
