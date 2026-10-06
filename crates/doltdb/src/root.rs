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

//! Doltgres root values, decoded for reading and editing their tables.

use std::sync::Arc;

use prolly::{AddressMapSerializer, Node, NodeStore, apply_mutations, walk_leaves};
use serial::write::{ROOT_OBJECT_COLLECTIONS, RootValueFields, write_root_value};
use serial::{DoltgresRootValue, Message};
use store::Hash;

use crate::database::{Database, Result};

/// Root is a Doltgres root value: its tables, schemas, foreign keys, and root object collections.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Root {
    pub feature_version: i64,
    pub collation: u16,
    /// The serialized root node of the tables' AddressMap, keyed by `table_key`.
    pub tables: Vec<u8>,
    pub schemas: Vec<Vec<u8>>,
    /// The address of the foreign key collection, all zeros without foreign keys.
    pub foreign_keys: Vec<u8>,
    pub root_objects: [Option<Hash>; ROOT_OBJECT_COLLECTIONS],
    /// The root object collection whose field Doltgres added after the others, which only changes the layout.
    pub added_root_object: Option<usize>,
}

/// table_key returns the key of a table in a root value's tables map.
pub fn table_key(schema: &str, name: &str) -> Vec<u8> {
    let mut key = vec![0];
    key.extend_from_slice(schema.as_bytes());
    key.push(0);
    key.extend_from_slice(name.as_bytes());
    key
}

impl Root {
    /// decode decodes a root value message.
    pub fn decode(bytes: &[u8]) -> Result<Root> {
        let r = DoltgresRootValue::new(Message(bytes))?;
        let mut root_objects = [None; ROOT_OBJECT_COLLECTIONS];
        for (i, (_, address)) in r.root_object_maps()?.into_iter().enumerate() {
            root_objects[i] = address.filter(|a| !a.is_empty()).map(serial::hash).transpose()?;
        }
        let mut root = Root {
            feature_version: r.feature_version()?,
            collation: r.collation()?,
            tables: r.tables()?.unwrap_or_default().to_vec(),
            schemas: r.schemas()?.into_iter().map(<[u8]>::to_vec).collect(),
            foreign_keys: r.foreign_keys()?.unwrap_or_default().to_vec(),
            root_objects,
            added_root_object: None,
        };
        if root.encode() != bytes {
            root.added_root_object = (0..ROOT_OBJECT_COLLECTIONS).find(|&i| {
                root_objects[i].is_some() && Root { added_root_object: Some(i), ..root.clone() }.encode() == bytes
            });
        }
        Ok(root)
    }

    /// encode writes the root value message.
    pub fn encode(&self) -> Vec<u8> {
        write_root_value(&RootValueFields {
            feature_version: self.feature_version,
            collation: self.collation,
            tables: &self.tables,
            schemas: self.schemas.iter().map(Vec::as_slice).collect(),
            foreign_keys: &self.foreign_keys,
            root_objects: self.root_objects,
            added_root_object: self.added_root_object,
        })
    }

    /// tables returns the key and address of every table, in key order.
    pub fn tables(&self, db: &mut Database) -> Result<Vec<(Vec<u8>, Hash)>> {
        let mut tables = Vec::new();
        walk_leaves(db, &Node::decode(self.tables.clone())?, &mut |key, value| {
            tables.push((key.to_vec(), serial::hash(value)?));
            Ok(())
        })?;
        Ok(tables)
    }

    /// table returns the address of a table, if the root value has it.
    pub fn table(&self, db: &mut Database, schema: &str, name: &str) -> Result<Option<Hash>> {
        let key = table_key(schema, name);
        Ok(self.tables(db)?.into_iter().find(|(k, _)| *k == key).map(|(_, address)| address))
    }

    /// put_table sets a table's address, or removes the table without one, writing the new tables map.
    pub fn put_table(&mut self, db: &mut Database, schema: &str, name: &str, address: Option<Hash>) -> Result<()> {
        let node = Arc::new(Node::decode(self.tables.clone())?);
        let edit = (table_key(schema, name), address.map(|a| a.0.to_vec()));
        let (_, node) =
            apply_mutations(db as &mut dyn NodeStore, node, AddressMapSerializer, [edit], &|a, b| a.cmp(b))?;
        self.tables = node.bytes().to_vec();
        Ok(())
    }
}
