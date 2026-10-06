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

//! Tables: their messages, and the row maps of their primary indexes.

use std::sync::Arc;

use prolly::{Node, NodeStore, ProllyNode, apply_mutations, serialize_prolly_node};
use serial::write::{TableFields, write_table};
use serial::{Message, TableMessage};
use store::Hash;

use crate::database::{Database, Result};

/// Table is a table message, decoded for editing.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Table {
    pub schema: Hash,
    /// The serialized root node of the primary index.
    pub primary_index: Vec<u8>,
    /// The serialized AddressMap of secondary indexes.
    pub secondary_indexes: Vec<u8>,
    pub auto_increment: u64,
    pub conflicts: [Vec<u8>; 4],
    pub violations: Vec<u8>,
    pub artifacts: Vec<u8>,
}

/// empty_rows returns the serialized empty leaf of a row map.
pub fn empty_rows() -> Vec<u8> {
    serialize_prolly_node(&ProllyNode {
        keys: Vec::new(),
        values: Vec::new(),
        subtrees: Vec::new(),
        level: 0,
        key_address_offsets: Vec::new(),
        value_address_offsets: Vec::new(),
    })
}

/// AddressOffsets returns the positions of addresses within a tuple.
type AddressOffsets = fn(&[u8]) -> Vec<u16>;

/// rows_serializer returns the serializer of row maps whose tuples hold no addresses.
pub fn rows_serializer() -> prolly::ProllyMapSerializer<AddressOffsets, AddressOffsets> {
    prolly::ProllyMapSerializer { key_addresses: |_| Vec::new(), value_addresses: |_| Vec::new() }
}

impl Table {
    /// create writes a new empty table with the schema message: the empty row map, then the schema, then the table,
    /// whose conflict, violation, and artifact addresses are empty.
    pub fn create(db: &mut Database, schema: Vec<u8>) -> Result<(Hash, Table)> {
        let rows = empty_rows();
        db.write(Hash::of(&rows), rows.clone())?;
        let schema = db.write_value(schema)?;
        let table = Table {
            schema,
            primary_index: rows,
            secondary_indexes: prolly::serialize_address_map(&[], &[], &[], 0),
            auto_increment: 0,
            conflicts: std::array::from_fn(|_| vec![0; Hash::LEN]),
            violations: vec![0; Hash::LEN],
            artifacts: vec![0; Hash::LEN],
        };
        Ok((table.write(db)?, table))
    }

    /// decode decodes a table message.
    pub fn decode(bytes: &[u8]) -> Result<Table> {
        let t = TableMessage::new(Message(bytes))?;
        let conflicts = t.conflicts()?;
        let conflict = |i| -> Result<Vec<u8>> {
            Ok(conflicts.as_ref().map(|c| c.bytes(i)).transpose()?.flatten().unwrap_or_default().to_vec())
        };
        Ok(Table {
            schema: t.schema()?,
            primary_index: t.primary_index()?.to_vec(),
            secondary_indexes: t.secondary_indexes()?.unwrap_or_default().to_vec(),
            auto_increment: t.auto_increment()?,
            conflicts: [conflict(0)?, conflict(1)?, conflict(2)?, conflict(3)?],
            violations: t.violations()?.unwrap_or_default().to_vec(),
            artifacts: t.artifacts()?.unwrap_or_default().to_vec(),
        })
    }

    /// write writes the table message and returns its address.
    pub fn write(&self, db: &mut Database) -> Result<Hash> {
        db.write_value(write_table(&TableFields {
            schema: self.schema,
            primary_index: &self.primary_index,
            secondary_indexes: &self.secondary_indexes,
            auto_increment: self.auto_increment,
            conflicts_data: &self.conflicts[0],
            conflicts_ours: &self.conflicts[1],
            conflicts_theirs: &self.conflicts[2],
            conflicts_ancestor: &self.conflicts[3],
            violations: &self.violations,
            artifacts: &self.artifacts,
        }))
    }

    /// edit_rows applies edits, sorted by the comparison, to the primary index, writing the changed nodes.
    pub fn edit_rows(
        &mut self,
        db: &mut Database,
        edits: Vec<(Vec<u8>, Option<Vec<u8>>)>,
        compare: &prolly::Compare<'_>,
    ) -> Result<()> {
        let node = Arc::new(Node::decode(self.primary_index.clone())?);
        let (_, node) = apply_mutations(db as &mut dyn NodeStore, node, rows_serializer(), edits, compare)?;
        self.primary_index = node.bytes().to_vec();
        Ok(())
    }
}
