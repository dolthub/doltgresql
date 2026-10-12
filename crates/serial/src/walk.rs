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

//! The addresses each message refers to, visited in the order Dolt's SerialMessage.WalkAddrs visits them, which
//! commits check for dangling references and GC follows.

use store::{Error, Hash, Result};

use crate::fb::Table;
use crate::messages::{DoltgresRootValue, hash};
use crate::{
    ADDRESS_MAP, BLOB, COMMIT, COMMIT_CLOSURE, DOLTGRES_ROOT_VALUE, FOREIGN_KEY_COLLECTION, MERGE_ARTIFACTS, Message,
    PROLLY_TREE_NODE, ROOT_VALUE, STASH, STASH_LIST, STATISTIC, STORE_ROOT, TABLE, TABLE_SCHEMA, TAG, TUPLE,
    VECTOR_INDEX_NODE, WORKING_SET,
};

/// AddressVisitor is called with each address a message refers to.
pub type AddressVisitor<'f> = dyn FnMut(Hash) -> Result<()> + 'f;

/// COMMIT_CLOSURE_KEY_LEN is the length of a commit closure key: a u64 height and a commit address.
const COMMIT_CLOSURE_KEY_LEN: usize = 8 + Hash::LEN;

/// field_hash returns the address in a byte vector field, which must hold one.
fn field_hash(table: &Table<'_>, field: usize) -> Result<Hash> {
    hash(table.bytes(field)?.unwrap_or_default())
}

/// visit_nonempty visits the address in a byte vector field unless it is all zero.
fn visit_nonempty(table: &Table<'_>, field: usize, f: &mut AddressVisitor<'_>) -> Result<()> {
    let address = field_hash(table, field)?;
    if address.is_empty() { Ok(()) } else { f(address) }
}

/// visit_array visits each address of a concatenated address array field.
fn visit_array(table: &Table<'_>, field: usize, f: &mut AddressVisitor<'_>) -> Result<()> {
    for address in table.bytes(field)?.unwrap_or_default().as_chunks::<{ Hash::LEN }>().0 {
        f(hash(address)?)?;
    }
    Ok(())
}

/// visit_offsets visits the addresses at the u16 offsets of one field within the items of another.
fn visit_offsets(table: &Table<'_>, offsets: usize, items: usize, f: &mut AddressVisitor<'_>) -> Result<()> {
    let Some(vector) = table.vector(offsets, 2)? else { return Ok(()) };
    let items = table.bytes(items)?.unwrap_or_default();
    for i in 0..vector.len() {
        let offset = vector.u16(i)? as usize;
        let address = items
            .get(offset..offset + Hash::LEN)
            .ok_or_else(|| Error::Corrupt("address offset out of range".into()))?;
        f(hash(address)?)?;
    }
    Ok(())
}

/// walk_inline walks an inline message held in a byte vector field, such as a root value's tables.
fn walk_inline(table: &Table<'_>, field: usize, f: &mut AddressVisitor<'_>) -> Result<()> {
    walk_addrs(Message(table.bytes(field)?.unwrap_or_default()), f)
}

/// walk_addrs calls the function with each address the message refers to, as Dolt's WalkAddrs does.
pub fn walk_addrs(message: Message<'_>, f: &mut AddressVisitor<'_>) -> Result<()> {
    let root = || message.root();
    match message.file_id() {
        STORE_ROOT | STASH_LIST => {
            let t = root()?;
            if t.bytes(0)?.is_some_and(|b| !b.is_empty()) {
                walk_inline(&t, 0, f)?;
            }
        }
        STATISTIC => f(field_hash(&root()?, 0)?)?,
        STASH => {
            let t = root()?;
            f(field_hash(&t, 0)?)?;
            f(field_hash(&t, 1)?)?;
        }
        TAG => f(field_hash(&root()?, 0)?)?,
        WORKING_SET => {
            let t = root()?;
            f(field_hash(&t, 0)?)?;
            if t.bytes(1)?.is_some_and(|b| !b.is_empty()) {
                f(field_hash(&t, 1)?)?;
            }
            if let Some(merge) = t.table(6)? {
                f(field_hash(&merge, 0)?)?;
                f(field_hash(&merge, 1)?)?;
            }
        }
        ROOT_VALUE => {
            let t = root()?;
            walk_inline(&t, 1, f)?;
            visit_nonempty(&t, 2, f)?;
        }
        DOLTGRES_ROOT_VALUE => {
            let t = root()?;
            walk_inline(&t, 1, f)?;
            visit_nonempty(&t, 2, f)?;
            for (_, address) in DoltgresRootValue(t).root_object_maps()? {
                if let Some(address) = address.filter(|a| !a.is_empty()) {
                    let address = hash(address)?;
                    if !address.is_empty() {
                        f(address)?;
                    }
                }
            }
        }
        TABLE => {
            let t = root()?;
            f(field_hash(&t, 0)?)?;
            let conflicts = t.table(4)?.ok_or_else(|| Error::Corrupt("table without conflicts".into()))?;
            for field in 0..4 {
                visit_nonempty(&conflicts, field, f)?;
            }
            visit_nonempty(&t, 5, f)?;
            visit_nonempty(&t, 6, f)?;
            walk_inline(&t, 2, f)?;
            walk_inline(&t, 1, f)?;
        }
        COMMIT => {
            let t = root()?;
            visit_array(&t, 2, f)?;
            f(field_hash(&t, 0)?)?;
            visit_nonempty(&t, 3, f)?;
        }
        TABLE_SCHEMA | FOREIGN_KEY_COLLECTION | TUPLE => {}
        PROLLY_TREE_NODE => {
            let t = root()?;
            visit_array(&t, 7, f)?;
            visit_offsets(&t, 6, 3, f)?;
            visit_offsets(&t, 11, 0, f)?;
        }
        ADDRESS_MAP => visit_array(&root()?, 2, f)?,
        MERGE_ARTIFACTS => {
            let t = root()?;
            visit_array(&t, 5, f)?;
            visit_offsets(&t, 2, 0, f)?;
        }
        COMMIT_CLOSURE => {
            let t = root()?;
            visit_array(&t, 1, f)?;
            if t.u8(4, 0)? == 0 {
                let keys = t.bytes(0)?.unwrap_or_default();
                for key in keys.as_chunks::<COMMIT_CLOSURE_KEY_LEN>().0 {
                    f(hash(&key[8..])?)?;
                }
            }
        }
        BLOB => visit_array(&root()?, 1, f)?,
        VECTOR_INDEX_NODE => visit_array(&root()?, 4, f)?,
        id => return Err(Error::Corrupt(format!("unsupported SerialMessage message with FileID: {id}"))),
    }
    Ok(())
}
