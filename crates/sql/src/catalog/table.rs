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

//! Table definitions, read from and written to Dolt schemas, and rows as index tuples.

use std::cmp::Ordering;

use doltdb::database::Database;
use doltdb::table::Table;
use prolly::Tuple;
use prolly::val::{build_tuple, encoding};
use serial::write::{ColumnFields, DEFAULT_TARGET_ROW_SIZE, SchemaFields, write_schema};
use serial::{Message, TableSchema};
use store::Hash;

use super::{ColumnType, builtin_type_by_id};
use crate::error::{PgError, Result};
use crate::storage::{compare_key_field, decode_field, encode_field, place_adaptive};
use crate::types::Value;

/// COLLATION is the collation of every Doltgres schema, utf8mb4_0900_bin.
const COLLATION: u16 = 309;

/// ColumnDef is a column of a table.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ColumnDef {
    pub name: String,
    pub ty: ColumnType,
    pub tag: u64,
    pub encoding: u8,
    pub nullable: bool,
    pub primary_key: bool,
    /// The default expression as stored, which is empty without one.
    pub default: String,
}

/// TableDef is a table: its columns, which of them form the primary key, and its storage.
#[derive(Clone, Debug, PartialEq)]
pub struct TableDef {
    pub schema: String,
    pub name: String,
    pub columns: Vec<ColumnDef>,
    /// The columns of the primary key in key order, empty for a keyless table.
    pub key_columns: Vec<usize>,
    /// The columns stored in the value tuple, in order.
    pub value_columns: Vec<usize>,
    pub table: Table,
}

/// column_type reads a column type from its form in a Dolt schema.
fn column_type(sql_type: &[u8]) -> Result<ColumnType> {
    let unsupported = || PgError::unsupported(format!("the column type {}", String::from_utf8_lossy(sql_type)));
    let hex = sql_type.strip_prefix(b"extended_").ok_or_else(unsupported)?;
    let bytes: Vec<u8> = hex
        .chunks(2)
        .map(|pair| std::str::from_utf8(pair).ok().and_then(|p| u8::from_str_radix(p, 16).ok()))
        .collect::<Option<_>>()
        .ok_or_else(unsupported)?;
    let definition = objects::SerializedType::deserialize(&bytes)?;
    let builtin = builtin_type_by_id(&definition.id).ok_or_else(unsupported)?;
    Ok(ColumnType { oid: builtin.oid, modifier: definition.att_typ_mod })
}

impl TableDef {
    /// load reads the table at the address.
    pub fn load(db: &mut Database, schema: &str, name: &str, address: Hash) -> Result<TableDef> {
        let missing = || PgError::internal(format!("missing chunk for table {schema}.{name}"));
        let table = Table::decode(&db.read_value(&address)?.ok_or_else(missing)?)?;
        let message = db.read_value(&table.schema)?.ok_or_else(missing)?;
        let message = TableSchema::new(Message(&message))?;
        let columns = message
            .columns()?
            .into_iter()
            .filter(|c| !c.hidden)
            .map(|c| {
                Ok(ColumnDef {
                    name: String::from_utf8_lossy(c.name).into_owned(),
                    ty: column_type(c.sql_type)?,
                    tag: c.tag,
                    encoding: c.encoding,
                    nullable: c.nullable,
                    primary_key: c.primary_key,
                    default: String::from_utf8_lossy(c.default_value).into_owned(),
                })
            })
            .collect::<Result<Vec<_>>>()?;
        let clustered = message.clustered_index()?;
        let key_columns = clustered.key_columns.iter().map(|&i| i as usize).collect();
        let value_columns = clustered.value_columns.iter().map(|&i| i as usize).collect();
        Ok(TableDef { schema: schema.to_string(), name: name.to_string(), columns, key_columns, value_columns, table })
    }

    /// keyless reports whether the table has no primary key.
    pub fn keyless(&self) -> bool {
        self.key_columns.is_empty()
    }

    /// schema_message writes the table's Dolt schema.
    pub fn schema_message(&self) -> Result<Vec<u8>> {
        schema_message(&self.columns, &self.key_columns, &self.value_columns)
    }

    /// key_encodings returns the field encodings of the primary index's keys.
    pub fn key_encodings(&self) -> Vec<u8> {
        if self.keyless() {
            return vec![encoding::HASH128];
        }
        self.key_columns.iter().map(|&i| self.columns[i].encoding).collect()
    }

    /// compare_keys orders two keys of the primary index.
    pub fn compare_keys(&self, left: &[u8], right: &[u8]) -> Ordering {
        let (left, right) = (Tuple(left), Tuple(right));
        for (i, field_encoding) in self.key_encodings().into_iter().enumerate() {
            let ordering =
                compare_key_field(field_encoding, left.field(i).ok().flatten(), right.field(i).ok().flatten());
            if ordering != Ordering::Equal {
                return ordering;
            }
        }
        Ordering::Equal
    }

    /// value_encodings returns the field encodings of the primary index's values.
    pub fn value_encodings(&self) -> Vec<u8> {
        let mut encodings = if self.keyless() { vec![encoding::UINT64] } else { Vec::new() };
        encodings.extend(self.value_columns.iter().map(|&i| self.columns[i].encoding));
        encodings
    }

    /// encode_row returns a row's key and value tuples. A keyless row's value starts with its cardinality, and its key
    /// is a hash of the rest of the value, as Dolt's keyless tables store them.
    pub fn encode_row(&self, db: &mut Database, row: &[Value]) -> Result<(Vec<u8>, Vec<u8>)> {
        let field = |i: usize| encode_field(&row[i], self.columns[i].encoding);
        let mut values: Vec<Option<Vec<u8>>> = Vec::with_capacity(self.value_columns.len() + 1);
        if self.keyless() {
            values.push(Some(1u64.to_le_bytes().to_vec()));
        }
        for &i in &self.value_columns {
            values.push(field(i)?);
        }
        place_adaptive(db, &mut values, &self.value_encodings(), DEFAULT_TARGET_ROW_SIZE as usize)?;
        let value = build_tuple(&values.iter().map(Option::as_deref).collect::<Vec<_>>());
        if self.keyless() {
            return Ok((keyless_key(&value), value));
        }
        let mut keys = self.key_columns.iter().map(|&i| field(i)).collect::<Result<Vec<_>>>()?;
        place_adaptive(db, &mut keys, &self.key_encodings(), DEFAULT_TARGET_ROW_SIZE as usize)?;
        Ok((build_tuple(&keys.iter().map(Option::as_deref).collect::<Vec<_>>()), value))
    }

    /// decode_row reads a row from its key and value tuples, with its cardinality, which is 1 for a keyed table.
    pub fn decode_row(&self, db: &Database, key: &[u8], value: &[u8]) -> Result<(Vec<Value>, u64)> {
        let mut row = vec![Value::Null; self.columns.len()];
        let (key, value) = (Tuple(key), Tuple(value));
        for (field, &i) in self.key_columns.iter().enumerate() {
            row[i] = decode_field(db, key.field(field)?, self.columns[i].encoding)?;
        }
        let mut cardinality = 1;
        let offset = if self.keyless() {
            let field = value.field(0)?.unwrap_or_default();
            cardinality = u64::from_le_bytes(field.try_into().map_err(|_| PgError::internal("bad cardinality"))?);
            1
        } else {
            0
        };
        for (field, &i) in self.value_columns.iter().enumerate() {
            row[i] = decode_field(db, value.field(field + offset)?, self.columns[i].encoding)?;
        }
        Ok((row, cardinality))
    }
}

/// schema_message writes a Dolt schema of the columns, with the key and value columns of its primary index.
pub fn schema_message(columns: &[ColumnDef], key_columns: &[usize], value_columns: &[usize]) -> Result<Vec<u8>> {
    let types: Vec<Vec<u8>> =
        columns.iter().map(|c| c.ty.serialized().map(String::into_bytes)).collect::<Result<_>>()?;
    let fields = columns
        .iter()
        .zip(&types)
        .map(|(c, sql_type)| ColumnFields {
            name: c.name.as_bytes(),
            sql_type,
            default_value: c.default.as_bytes(),
            comment: b"",
            on_update: b"",
            tag: c.tag,
            encoding: c.encoding,
            primary_key: c.primary_key,
            auto_increment: false,
            nullable: c.nullable,
            generated: false,
            is_virtual: false,
            adaptive_encoding: false,
            hidden: false,
            hidden_system: false,
        })
        .collect();
    Ok(write_schema(&SchemaFields {
        columns: fields,
        keyless: key_columns.is_empty(),
        key_columns: key_columns.iter().map(|&i| i as u16).collect(),
        value_columns: value_columns.iter().map(|&i| i as u16).collect(),
        indexes: Vec::new(),
        checks: Vec::new(),
        collation: COLLATION,
        comment: b"",
        target_row_size: DEFAULT_TARGET_ROW_SIZE,
    }))
}

/// keyless_key returns the key of a keyless row: the xxh3 hash of its value after the cardinality, as Dolt's
/// HashTupleFromValue computes it.
pub fn keyless_key(value: &[u8]) -> Vec<u8> {
    let hash = xxhash_rust::xxh3::xxh3_128(&value[8..]);
    let mut key = Vec::with_capacity(18);
    key.extend_from_slice(&(hash as u64).to_le_bytes());
    key.extend_from_slice(&((hash >> 64) as u64).to_le_bytes());
    key.extend_from_slice(&1u16.to_le_bytes());
    key
}
