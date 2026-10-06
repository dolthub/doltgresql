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

/// Check is a check constraint: its name and its expression's SQL text.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Check {
    pub name: String,
    pub expression: String,
}

/// IndexDef is a secondary index of a table.
#[derive(Clone, Debug, PartialEq)]
pub struct IndexDef {
    pub name: String,
    /// The indexed columns, by position in the table's columns.
    pub columns: Vec<usize>,
    pub unique: bool,
    /// Each indexed column's direction and NULLS placement.
    pub descending: Vec<bool>,
    pub nulls_last: Vec<bool>,
    /// Each indexed column's operator class, empty for the default.
    pub op_classes: Vec<String>,
    pub comment: String,
    /// The SQL text of a partial index's predicate, empty for a full index.
    pub predicate: String,
    /// The address of the index's root node.
    pub root: Hash,
    /// Whether Dolt made the index itself, as it does for a foreign key's columns.
    pub system: bool,
}

/// TableDef is a table: its columns, which of them form the primary key, and its storage.
#[derive(Clone, Debug, PartialEq)]
pub struct TableDef {
    pub schema: String,
    pub name: String,
    pub columns: Vec<ColumnDef>,
    pub checks: Vec<Check>,
    pub indexes: Vec<IndexDef>,
    /// The columns of the primary key in key order, empty for a keyless table.
    pub key_columns: Vec<usize>,
    /// The columns stored in the value tuple, in order.
    pub value_columns: Vec<usize>,
    pub table: Table,
}

/// column_type reads a column type from its form in a Dolt schema: a Doltgres type, or one of the MySQL types of
/// Dolt's own tables, such as dolt_schemas.
fn column_type(sql_type: &[u8]) -> Result<ColumnType> {
    let unsupported = || PgError::unsupported(format!("the column type {}", String::from_utf8_lossy(sql_type)));
    let Some(hex) = sql_type.strip_prefix(b"extended_") else {
        let text = String::from_utf8_lossy(sql_type);
        let base = text.split_whitespace().next().unwrap_or_default();
        return Ok(match base {
            "text" | "tinytext" | "mediumtext" | "longtext" => ColumnType { oid: crate::oid::TEXT, modifier: -1 },
            "json" => ColumnType { oid: crate::oid::JSON, modifier: -1 },
            _ => match base.strip_prefix("varchar(").and_then(|n| n.strip_suffix(')')) {
                Some(n) => {
                    ColumnType { oid: crate::oid::VARCHAR, modifier: n.parse::<i32>().map_err(|_| unsupported())? + 4 }
                }
                None => return Err(unsupported()),
            },
        });
    };
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
        // A keyless table's index also holds its hidden hash and cardinality columns, which come after the others.
        let visible = |i: &u16| (*i as usize) < columns.len();
        let key_columns = clustered.key_columns.iter().filter(|i| visible(i)).map(|&i| i as usize).collect();
        let value_columns = clustered.value_columns.iter().filter(|i| visible(i)).map(|&i| i as usize).collect();
        let checks = message
            .checks()?
            .into_iter()
            .map(|c| Check {
                name: String::from_utf8_lossy(c.name).into_owned(),
                expression: String::from_utf8_lossy(c.expression).into_owned(),
            })
            .collect();
        let roots = table.indexes(db)?;
        let lossy = |b: &[u8]| String::from_utf8_lossy(b).into_owned();
        let indexes = message
            .secondary_indexes()?
            .into_iter()
            .map(|index| {
                let name = lossy(index.name);
                let root = roots.iter().find(|(n, _)| *n == name).map(|(_, r)| *r).ok_or_else(missing)?;
                let count = index.index_columns.len();
                Ok(IndexDef {
                    columns: index.index_columns.iter().map(|&i| i as usize).collect(),
                    unique: index.unique_key,
                    descending: (0..count).map(|i| index.descending.get(i).copied().unwrap_or(false)).collect(),
                    nulls_last: (0..count).map(|i| index.nulls_last.get(i).copied().unwrap_or(false)).collect(),
                    op_classes: (0..count)
                        .map(|i| index.op_classes.get(i).map_or(String::new(), |c| lossy(c)))
                        .collect(),
                    comment: lossy(index.comment),
                    predicate: lossy(index.predicate),
                    system: index.system_defined,
                    name,
                    root,
                })
            })
            .collect::<Result<Vec<_>>>()?;
        Ok(TableDef {
            schema: schema.to_string(),
            name: name.to_string(),
            columns,
            checks,
            indexes,
            key_columns,
            value_columns,
            table,
        })
    }

    /// index_key_columns returns the columns of an index's keys: the indexed columns, then the primary key columns
    /// the index lacks, or the hidden row hash of a keyless table, written as the column count.
    pub fn index_key_columns(&self, index: &IndexDef) -> Vec<usize> {
        let mut columns = index.columns.clone();
        if self.keyless() {
            columns.push(self.columns.len());
        } else {
            columns.extend(self.key_columns.iter().filter(|c| !index.columns.contains(c)));
        }
        columns
    }

    /// index_encodings returns the field encodings of an index's keys.
    pub fn index_encodings(&self, index: &IndexDef) -> Vec<u8> {
        self.index_key_columns(index)
            .into_iter()
            .map(|c| self.columns.get(c).map_or(encoding::HASH128, |c| c.encoding))
            .collect()
    }

    /// index_key returns a row's key in an index, given the row's primary key tuple.
    pub fn index_key(&self, db: &mut Database, index: &IndexDef, row: &[Value], primary: &[u8]) -> Result<Vec<u8>> {
        let mut fields = Vec::new();
        for c in self.index_key_columns(index) {
            fields.push(match self.columns.get(c) {
                Some(column) => encode_field(&row[c], column.encoding, column.ty)?,
                None => Tuple(primary).field(0)?.map(<[u8]>::to_vec),
            });
        }
        place_adaptive(db, &mut fields, &self.index_encodings(index), DEFAULT_TARGET_ROW_SIZE as usize)?;
        Ok(build_tuple(&fields.iter().map(Option::as_deref).collect::<Vec<_>>()))
    }

    /// compare_index_keys orders two keys of an index with each indexed column's direction and NULLS placement, as
    /// Dolt's OrderedTupleComparator does.
    pub fn compare_index_keys(&self, index: &IndexDef, left: &[u8], right: &[u8]) -> Ordering {
        self.compare_index_prefix(index, self.index_key_columns(index).len(), left, right)
    }

    /// compare_index_prefix orders the first fields of two keys of an index.
    pub fn compare_index_prefix(&self, index: &IndexDef, fields: usize, left: &[u8], right: &[u8]) -> Ordering {
        let (left, right) = (Tuple(left), Tuple(right));
        for (i, c) in self.index_key_columns(index).into_iter().enumerate().take(fields) {
            let (l, r) = (left.field(i).ok().flatten(), right.field(i).ok().flatten());
            let (encoding, ty) = self
                .columns
                .get(c)
                .map_or((encoding::HASH128, ColumnType { oid: 0, modifier: -1 }), |c| (c.encoding, c.ty));
            let descending = index.descending.get(i).copied().unwrap_or(false);
            let nulls_last = index.nulls_last.get(i).copied().unwrap_or(false);
            let ordering = match (l, r) {
                (None, None) => Ordering::Equal,
                (None, Some(_)) if nulls_last => Ordering::Greater,
                (None, Some(_)) => Ordering::Less,
                (Some(_), None) if nulls_last => Ordering::Less,
                (Some(_), None) => Ordering::Greater,
                (Some(_), Some(_)) => {
                    let ordering = compare_key_field(encoding, ty, l, r);
                    if descending { ordering.reverse() } else { ordering }
                }
            };
            if ordering != Ordering::Equal {
                return ordering;
            }
        }
        Ordering::Equal
    }

    /// keyless reports whether the table has no primary key.
    pub fn keyless(&self) -> bool {
        self.key_columns.is_empty()
    }

    /// schema_message writes the table's Dolt schema.
    pub fn schema_message(&self) -> Result<Vec<u8>> {
        schema_message(&self.columns, &self.key_columns, &self.value_columns, &self.checks, &self.indexes)
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
            let ty = self.key_columns.get(i).map_or(ColumnType { oid: 0, modifier: -1 }, |&c| self.columns[c].ty);
            let ordering =
                compare_key_field(field_encoding, ty, left.field(i).ok().flatten(), right.field(i).ok().flatten());
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
        let field = |i: usize| encode_field(&row[i], self.columns[i].encoding, self.columns[i].ty);
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
            row[i] = decode_field(db, key.field(field)?, self.columns[i].encoding, self.columns[i].ty)?;
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
            row[i] = decode_field(db, value.field(field + offset)?, self.columns[i].encoding, self.columns[i].ty)?;
        }
        Ok((row, cardinality))
    }
}

/// schema_message writes a Dolt schema of the columns, with the key and value columns of its primary index, its check
/// constraints, and its secondary indexes, whose key columns leave out the row hash that ends a keyless table's keys.
pub fn schema_message(
    columns: &[ColumnDef],
    key_columns: &[usize],
    value_columns: &[usize],
    checks: &[Check],
    indexes: &[IndexDef],
) -> Result<Vec<u8>> {
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
            adaptive_encoding: crate::storage::is_adaptive(c.encoding),
            hidden: false,
            hidden_system: false,
        })
        .collect();
    let keyless = key_columns.is_empty();
    let index_fields = indexes
        .iter()
        .map(|index| {
            let mut keys: Vec<u16> = index.columns.iter().map(|&i| i as u16).collect();
            keys.extend(key_columns.iter().filter(|c| !index.columns.contains(c)).map(|&i| i as u16));
            serial::write::IndexFields {
                name: index.name.as_bytes(),
                comment: index.comment.as_bytes(),
                predicate: index.predicate.as_bytes(),
                index_columns: index.columns.iter().map(|&i| i as u16).collect(),
                key_columns: keys,
                prefix_lengths: Vec::new(),
                descending: if index.system { Vec::new() } else { index.descending.clone() },
                nulls_last: if index.system { Vec::new() } else { index.nulls_last.clone() },
                op_classes: if index.op_classes.iter().all(String::is_empty) {
                    Vec::new()
                } else {
                    index.op_classes.iter().map(String::as_bytes).collect()
                },
                unique: index.unique,
                system_defined: index.system,
                spatial: false,
                fulltext: None,
                vector_distance: None,
            }
        })
        .collect();
    // A keyless table's hidden hash and cardinality columns follow the others.
    let (key_columns, value_columns): (Vec<u16>, Vec<u16>) = if keyless {
        let n = columns.len() as u16;
        (vec![n], std::iter::once(n + 1).chain(value_columns.iter().map(|&i| i as u16)).collect())
    } else {
        (key_columns.iter().map(|&i| i as u16).collect(), value_columns.iter().map(|&i| i as u16).collect())
    };
    Ok(write_schema(&SchemaFields {
        columns: fields,
        keyless,
        key_columns,
        value_columns,
        indexes: index_fields,
        checks: checks
            .iter()
            .map(|c| serial::write::CheckFields {
                name: c.name.as_bytes(),
                expression: c.expression.as_bytes(),
                enforced: true,
                is_not_valid: false,
            })
            .collect(),
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
