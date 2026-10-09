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
use crate::error::{PgError, Result, code};
use crate::storage::{compare_key_field, decode_field_into, encode_field, place_adaptive};
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
    /// The default expression as stored, which is empty without one, or the expression of a generated column.
    pub default: String,
    /// Whether the column is generated from its expression, as a stored generated column is.
    pub generated: bool,
    /// The MySQL type that Dolt gives a column of one of its own tables, as its schema names it, which is empty for a
    /// Doltgres type.
    pub mysql_type: String,
    /// The comment that COMMENT ON COLUMN gives the column, which is empty without one.
    pub comment: String,
    /// Whether the column is an identity column, `a` for GENERATED ALWAYS and `d` for BY DEFAULT, or 0, as
    /// pg_attribute's attidentity says.
    pub identity: u8,
    /// Whether the column's array type has the version that older versions wrote, which cannot hold multidimensional
    /// values until ALTER COLUMN TYPE upgrades it.
    pub legacy_array: bool,
}

/// Check is a check constraint: its name and its expression's SQL text.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Check {
    pub name: String,
    pub expression: String,
}

/// HIDDEN_BASE is the position an index gives a table's first hidden expression column, past any table's columns.
pub const HIDDEN_BASE: usize = 1 << 20;

/// KEYLESS_HASH is the position an index gives the row hash that ends a keyless table's index keys.
pub const KEYLESS_HASH: usize = usize::MAX;

/// IndexDef is a secondary index of a table.
#[derive(Clone, Debug, PartialEq)]
pub struct IndexDef {
    pub name: String,
    /// The indexed columns, by position in the table's columns, or past `HIDDEN_BASE` for a hidden expression column.
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
    /// The distance of a vector index, which Dolt stores as a proximity map rather than a prolly map.
    pub vector: Option<prolly::Distance>,
    /// Whether a unique index's constraint is DEFERRABLE, and whether it is INITIALLY DEFERRED.
    pub deferrable: bool,
    pub initially_deferred: bool,
    /// Whether a unique index came from CREATE UNIQUE INDEX rather than a UNIQUE constraint, which leaves it backing
    /// no constraint.
    pub plain: bool,
}

impl IndexDef {
    /// empty_root writes the index's root for when it holds no rows, which a vector index stores as a proximity map,
    /// and returns its address.
    pub fn empty_root(&self, db: &mut Database) -> Result<Hash> {
        match self.vector {
            Some(distance) => {
                let mut sink = |_: Hash, bytes: &[u8]| {
                    db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
                };
                Ok(prolly::write_proximity_map(Vec::new(), distance, &mut sink)?)
            }
            None => Ok(db.write_value(doltdb::table::empty_rows())?),
        }
    }
}

/// TableDef is a table: its columns, which of them form the primary key, and its storage.
#[derive(Clone, Debug, PartialEq)]
pub struct TableDef {
    pub schema: String,
    pub name: String,
    pub primary: Primary,
    pub columns: Vec<ColumnDef>,
    /// The hidden virtual columns that hold the expressions of expression indexes, as Dolt stores them.
    pub hidden: Vec<ColumnDef>,
    pub checks: Vec<Check>,
    pub indexes: Vec<IndexDef>,
    /// The columns of the primary key in key order, empty for a keyless table.
    pub key_columns: Vec<usize>,
    /// The columns stored in the value tuple, in order.
    pub value_columns: Vec<usize>,
    /// The comment that COMMENT ON TABLE gives the table, which is empty without one.
    pub comment: String,
    pub table: Table,
    /// The name that a query's FROM gives the table, which EXPLAIN shows, or None without one.
    pub alias: Option<String>,
}

/// Primary is a primary key constraint's name, empty for the default `<table>_pkey`, and whether it is DEFERRABLE
/// and INITIALLY DEFERRED.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Primary {
    pub name: String,
    pub deferrable: bool,
    pub initially_deferred: bool,
}

/// column_type reads a column type from its form in a Dolt schema: a Doltgres type, or one of the MySQL types of
/// Dolt's own tables, such as dolt_schemas, with whether it is an array type of the version older versions wrote.
fn column_type(sql_type: &[u8], user_types: &mut Vec<objects::SerializedType>) -> Result<(ColumnType, bool)> {
    let unsupported = || PgError::unsupported(format!("the column type {}", String::from_utf8_lossy(sql_type)));
    let Some(hex) = sql_type.strip_prefix(b"extended_") else {
        let text = String::from_utf8_lossy(sql_type);
        let base = text.split_whitespace().next().unwrap_or_default();
        let oid = match base {
            "text" | "tinytext" | "mediumtext" | "longtext" => crate::oid::TEXT,
            "json" => crate::oid::JSON,
            _ if base.starts_with("varchar(") => crate::oid::TEXT,
            _ => return Err(unsupported()),
        };
        return Ok((ColumnType { oid, modifier: -1 }, false));
    };
    let bytes: Vec<u8> = hex
        .chunks(2)
        .map(|pair| std::str::from_utf8(pair).ok().and_then(|p| u8::from_str_radix(p, 16).ok()))
        .collect::<Option<_>>()
        .ok_or_else(unsupported)?;
    let definition = objects::SerializedType::deserialize(&bytes)?;
    let (modifier, legacy_array) = (
        definition.att_typ_mod,
        definition.version == 0 && definition.typ_category == b"A" && definition.typ_type == b"b",
    );
    let oid = match builtin_type_by_id(&definition.id) {
        Some(builtin) => builtin.oid,
        None => {
            user_types.push(definition.clone());
            crate::usertypes::register(definition)
        }
    };
    Ok((ColumnType { oid, modifier }, legacy_array))
}

/// CACHE_LIMIT is how many tables and schemas a thread keeps decoded before it starts over.
const CACHE_LIMIT: usize = 1024;

/// Decoded is a table definition that a thread decoded, with the user types its columns registered.
type Decoded = (std::sync::Arc<TableDef>, Vec<objects::SerializedType>);

thread_local! {
    /// TABLES holds the definitions of the tables this thread decoded by their address, and SCHEMAS by their schema's
    /// address, which never change since chunks are addressed by their content.
    static TABLES: std::cell::RefCell<std::collections::HashMap<Hash, Decoded>> = Default::default();
    static SCHEMAS: std::cell::RefCell<std::collections::HashMap<Hash, Decoded>> = Default::default();
}

/// remember adds a decoded definition to a cache, emptying the cache first when it is full.
fn remember(
    cache: &'static std::thread::LocalKey<std::cell::RefCell<std::collections::HashMap<Hash, Decoded>>>,
    address: Hash,
    decoded: Decoded,
) {
    cache.with(|c| {
        let mut c = c.borrow_mut();
        if c.len() >= CACHE_LIMIT {
            c.clear();
        }
        c.insert(address, decoded);
    });
}

impl TableDef {
    /// load reads the table at the address, reusing what this thread decoded of the table or its schema before.
    pub fn load(db: &mut Database, schema: &str, name: &str, address: Hash) -> Result<TableDef> {
        let mut table = TableDef::clone(&*TableDef::shared(db, schema, name, address)?);
        table.schema = schema.to_string();
        table.name = name.to_string();
        Ok(table)
    }

    /// shared reads the table at the address as this thread decoded it before, without copying it, where the schema
    /// and name may be another table's with the same contents.
    pub fn shared(db: &mut Database, schema: &str, name: &str, address: Hash) -> Result<std::sync::Arc<TableDef>> {
        let shared = |(table, user_types): Decoded| {
            for definition in user_types {
                crate::usertypes::register(definition);
            }
            table
        };
        if let Some(decoded) = TABLES.with(|c| c.borrow().get(&address).cloned()) {
            return Ok(shared(decoded));
        }
        let missing = || PgError::internal(format!("missing chunk for table {schema}.{name}"));
        let table = Table::decode(&db.read_value(&address)?.ok_or_else(missing)?)?;
        let roots = table.indexes(db)?;
        let decoded = match SCHEMAS.with(|c| c.borrow().get(&table.schema).cloned()) {
            Some((shape, user_types)) => {
                let mut def = TableDef::clone(&shape);
                for index in &mut def.indexes {
                    index.root = roots.iter().find(|(n, _)| *n == index.name).map(|(_, r)| *r).ok_or_else(missing)?;
                }
                def.table = table;
                (std::sync::Arc::new(def), user_types)
            }
            None => {
                let mut user_types = Vec::new();
                let def = TableDef::decode(db, schema, name, table, &roots, &mut user_types)?;
                let decoded = (std::sync::Arc::new(def), user_types);
                remember(&SCHEMAS, decoded.0.table.schema, decoded.clone());
                decoded
            }
        };
        remember(&TABLES, address, decoded.clone());
        Ok(shared(decoded))
    }

    /// decode decodes a table and the definition its schema gives, with the roots of its secondary indexes, listing
    /// the user types that its columns register.
    fn decode(
        db: &mut Database,
        schema: &str,
        name: &str,
        table: Table,
        roots: &[(String, Hash)],
        user_types: &mut Vec<objects::SerializedType>,
    ) -> Result<TableDef> {
        let missing = || PgError::internal(format!("missing chunk for table {schema}.{name}"));
        let message = db.read_value(&table.schema)?.ok_or_else(missing)?;
        let message = TableSchema::new(Message(&message))?;
        let (mut columns, mut hidden, mut positions) = (Vec::new(), Vec::new(), Vec::new());
        for c in message.columns()? {
            if c.hidden && !(c.hidden_system && c.is_virtual) {
                positions.push(None);
                continue;
            }
            let (ty, legacy_array) = column_type(c.sql_type, user_types)?;
            let column = ColumnDef {
                name: String::from_utf8_lossy(c.name).into_owned(),
                ty,
                tag: c.tag,
                encoding: c.encoding,
                nullable: c.nullable,
                primary_key: c.primary_key,
                default: String::from_utf8_lossy(c.default_value).into_owned(),
                generated: c.generated,
                mysql_type: match c.sql_type.starts_with(b"extended_") {
                    true => String::new(),
                    false => String::from_utf8_lossy(c.sql_type).into_owned(),
                },
                comment: String::from_utf8_lossy(c.comment).into_owned(),
                identity: c.identity,
                legacy_array,
            };
            if c.hidden_system && c.is_virtual {
                positions.push(Some(HIDDEN_BASE + hidden.len()));
                hidden.push(column);
            } else {
                positions.push(Some(columns.len()));
                columns.push(column);
            }
        }
        let clustered = message.clustered_index()?;
        // A keyless table's index also holds its hidden hash and cardinality columns, which come after the others.
        let position = |i: &u16| positions.get(*i as usize).copied().flatten();
        let stored = |i: &u16| position(i).filter(|&p| p < HIDDEN_BASE);
        let key_columns = clustered.key_columns.iter().filter_map(stored).collect();
        let value_columns = clustered.value_columns.iter().filter_map(stored).collect();
        let checks = message
            .checks()?
            .into_iter()
            .map(|c| Check {
                name: String::from_utf8_lossy(c.name).into_owned(),
                expression: String::from_utf8_lossy(c.expression).into_owned(),
            })
            .collect();
        let lossy = |b: &[u8]| String::from_utf8_lossy(b).into_owned();
        let indexes = message
            .secondary_indexes()?
            .into_iter()
            .map(|index| {
                let name = lossy(index.name);
                let root = roots.iter().find(|(n, _)| *n == name).map(|(_, r)| *r).ok_or_else(missing)?;
                let count = index.index_columns.len();
                Ok(IndexDef {
                    columns: index.index_columns.iter().map(|i| position(i).unwrap_or(KEYLESS_HASH)).collect(),
                    unique: index.unique_key,
                    descending: (0..count).map(|i| index.descending.get(i).copied().unwrap_or(false)).collect(),
                    nulls_last: (0..count).map(|i| index.nulls_last.get(i).copied().unwrap_or(false)).collect(),
                    op_classes: (0..count)
                        .map(|i| index.op_classes.get(i).map_or(String::new(), |c| lossy(c)))
                        .collect(),
                    comment: lossy(index.comment),
                    predicate: lossy(index.predicate),
                    system: index.system_defined,
                    vector: if index.vector_key {
                        Some(prolly::Distance::from_stored(index.vector_distance.unwrap_or(0)).ok_or_else(missing)?)
                    } else {
                        None
                    },
                    name,
                    root,
                    deferrable: index.deferrable,
                    initially_deferred: index.initially_deferred,
                    plain: index.plain,
                })
            })
            .collect::<Result<Vec<_>>>()?;
        let primary = Primary {
            name: lossy(clustered.name),
            deferrable: clustered.deferrable,
            initially_deferred: clustered.initially_deferred,
        };
        Ok(TableDef {
            alias: None,
            schema: schema.to_string(),
            name: name.to_string(),
            primary,
            columns,
            hidden,
            checks,
            indexes,
            key_columns,
            value_columns,
            comment: lossy(message.comment()?),
            table,
        })
    }

    /// index_column returns the column at a position an index gives, a table column or a hidden expression column,
    /// or None for a keyless table's row hash.
    pub fn index_column(&self, position: usize) -> Option<&ColumnDef> {
        match position.checked_sub(HIDDEN_BASE) {
            Some(k) => self.hidden.get(k),
            None => self.columns.get(position),
        }
    }

    /// drop_index removes an index with the hidden expression columns that only it reads.
    pub fn drop_index(&mut self, name: &str) {
        let Some(i) = self.indexes.iter().position(|index| index.name == name) else { return };
        let index = self.indexes.remove(i);
        let mut doomed: Vec<usize> = index.columns.iter().filter_map(|c| c.checked_sub(HIDDEN_BASE)).collect();
        doomed.sort_unstable();
        for &k in doomed.iter().rev() {
            self.hidden.remove(k);
        }
        for column in self.indexes.iter_mut().flat_map(|index| &mut index.columns) {
            if let Some(k) = column.checked_sub(HIDDEN_BASE) {
                *column -= doomed.iter().filter(|&&d| d < k).count();
            }
        }
    }

    /// row_position returns where the value at a position an index gives sits in a row that the hidden columns'
    /// values extend.
    pub fn row_position(&self, position: usize) -> usize {
        match position.checked_sub(HIDDEN_BASE) {
            Some(k) => self.columns.len() + k,
            None => position,
        }
    }

    /// index_key_columns returns the columns of an index's keys: the indexed columns, then the primary key columns
    /// the index lacks, or the hidden row hash of a keyless table, written as the column count.
    pub fn index_key_columns(&self, index: &IndexDef) -> Vec<usize> {
        let mut columns = index.columns.clone();
        if self.keyless() {
            columns.push(KEYLESS_HASH);
        } else {
            columns.extend(self.key_columns.iter().filter(|c| !index.columns.contains(c)));
        }
        columns
    }

    /// index_encodings returns the field encodings of an index's keys.
    pub fn index_encodings(&self, index: &IndexDef) -> Vec<u8> {
        self.index_key_columns(index)
            .into_iter()
            .map(|c| self.index_column(c).map_or(encoding::HASH128, |c| c.encoding))
            .collect()
    }

    /// index_key returns a row's key in an index, given the row's primary key tuple.
    pub fn index_key(&self, db: &mut Database, index: &IndexDef, row: &[Value], primary: &[u8]) -> Result<Vec<u8>> {
        let mut fields = Vec::new();
        for c in self.index_key_columns(index) {
            fields.push(match self.index_column(c) {
                Some(column) => encode_field(&row[self.row_position(c)], column.encoding, column.ty)?,
                None => Tuple(primary).field(0)?.map(<[u8]>::to_vec),
            });
        }
        place_adaptive(db, &mut fields, &self.index_encodings(index), DEFAULT_TARGET_ROW_SIZE as usize)?;
        Ok(build_tuple(&fields.iter().map(Option::as_deref).collect::<Vec<_>>()))
    }

    /// write_vector_index writes a vector index of the table's rows as Dolt's proximity map, leaving out the rows whose
    /// vector is NULL, and returns the address of its root.
    pub fn write_vector_index(&self, db: &mut Database, index: &IndexDef, distance: prolly::Distance) -> Result<Hash> {
        let mut entries = Vec::new();
        for row in crate::query::scan(db, self)? {
            let Value::Base(base) = &row[index.columns[0]] else { continue };
            let Some(vector) = crate::types::base_type(base.type_oid).and_then(|t| t.vector) else { continue };
            let (primary, _) = self.encode_row(db, &row)?;
            let key = self.index_key(db, index, &row, &primary)?;
            entries.push(prolly::Entry { key, value: build_tuple(&[]), vector: vector(&base.data) });
        }
        let mut sink = |_: Hash, bytes: &[u8]| {
            db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
        };
        Ok(prolly::write_proximity_map(entries, distance, &mut sink)?)
    }

    /// compare_index_keys orders two keys of an index with each indexed column's direction and NULLS placement, as
    /// Dolt's OrderedTupleComparator does.
    pub fn compare_index_keys(&self, index: &IndexDef, left: &[u8], right: &[u8]) -> Ordering {
        self.compare_index_prefix(index, self.index_key_columns(index).len(), left, right)
    }

    /// compare_index_prefix orders the first fields of two keys of an index.
    pub fn compare_index_prefix(&self, index: &IndexDef, fields: usize, left: &[u8], right: &[u8]) -> Ordering {
        let (left, right) = (Tuple(left), Tuple(right));
        let key_columns = index
            .columns
            .iter()
            .copied()
            .chain(self.keyless().then_some(KEYLESS_HASH))
            .chain(self.key_columns.iter().copied().filter(|c| !index.columns.contains(c)));
        for (i, c) in key_columns.enumerate().take(fields) {
            let (l, r) = (left.field(i).ok().flatten(), right.field(i).ok().flatten());
            let (encoding, ty) = self
                .index_column(c)
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

    /// primary_name returns the name of the table's primary key constraint and of its index.
    pub fn primary_name(&self) -> String {
        if self.primary.name.is_empty() { format!("{}_pkey", self.name) } else { self.primary.name.clone() }
    }

    /// schema_message writes the table's Dolt schema.
    pub fn schema_message(&self) -> Result<Vec<u8>> {
        schema_message(
            &self.columns,
            &self.hidden,
            (&self.key_columns, &self.value_columns),
            &self.checks,
            &self.indexes,
            &self.primary,
            &self.comment,
        )
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
        // A key of one integer column is the integer's bytes and then the field count, which compare directly.
        if let [c] = self.key_columns.as_slice() {
            let int = |b: &[u8]| -> i64 {
                match b.len() {
                    4 => i16::from_le_bytes([b[0], b[1]]) as i64,
                    6 => i32::from_le_bytes([b[0], b[1], b[2], b[3]]) as i64,
                    _ => i64::from_le_bytes([b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7]]),
                }
            };
            let width = match self.columns[*c].encoding {
                encoding::INT16 => 4,
                encoding::INT32 => 6,
                encoding::INT64 => 10,
                _ => 0,
            };
            if width > 0 && left.len() == width && right.len() == width {
                return int(left).cmp(&int(right));
            }
        }
        let (left, right) = (Tuple(left), Tuple(right));
        if self.keyless() {
            let ty = ColumnType { oid: 0, modifier: -1 };
            return compare_key_field(
                encoding::HASH128,
                ty,
                left.field(0).ok().flatten(),
                right.field(0).ok().flatten(),
            );
        }
        for (i, &c) in self.key_columns.iter().enumerate() {
            let column = &self.columns[c];
            let ordering = compare_key_field(
                column.encoding,
                column.ty,
                left.field(i).ok().flatten(),
                right.field(i).ok().flatten(),
            );
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
        let field = |i: usize| {
            if let (true, Value::Array(array)) = (self.columns[i].legacy_array, &row[i])
                && array.dims.len() > 1
            {
                return Err(PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    "multidimensional arrays are not supported by the column's type version, alter the column's type \
                     to upgrade it",
                ));
            }
            encode_field(&row[i], self.columns[i].encoding, self.columns[i].ty)
        };
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
        self.decode_columns(db, key, value, None)
    }

    /// decode_columns decodes the columns of a primary index entry that the mask selects, or all of them without a
    /// mask, leaving the others NULL, with the row's cardinality.
    pub fn decode_columns(
        &self,
        db: &Database,
        key: &[u8],
        value: &[u8],
        needed: Option<&[bool]>,
    ) -> Result<(Vec<Value>, u64)> {
        let mut row = Vec::new();
        let cardinality = self.decode_columns_into(db, key, value, needed, &mut row)?;
        Ok((row, cardinality))
    }

    /// decode_columns_into is `decode_columns` into a buffer that holds nothing or a row it decoded before with the
    /// same mask, whose other columns are still NULL, returning the row's cardinality.
    pub fn decode_columns_into(
        &self,
        db: &Database,
        key: &[u8],
        value: &[u8],
        needed: Option<&[bool]>,
        row: &mut Vec<Value>,
    ) -> Result<u64> {
        if row.len() != self.columns.len() {
            *row = Value::nulls(self.columns.len());
        }
        let (key, value) = (Tuple(key), Tuple(value));
        let wanted = |i: usize| needed.is_none_or(|n| n[i]);
        for (field, &i) in self.key_columns.iter().enumerate() {
            if wanted(i) {
                decode_field_into(db, key.field(field)?, self.columns[i].encoding, self.columns[i].ty, &mut row[i])?;
            }
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
            if wanted(i) {
                let field = value.field(field + offset)?;
                decode_field_into(db, field, self.columns[i].encoding, self.columns[i].ty, &mut row[i])?;
            }
        }
        Ok(cardinality)
    }
}

/// schema_message writes a Dolt schema of the columns, then the hidden expression columns as Dolt's virtual hidden
/// system columns, with the key and value columns of its primary index, its check constraints, its secondary indexes,
/// whose key columns leave out the row hash that ends a keyless table's keys, and its primary key constraint.
pub fn schema_message(
    columns: &[ColumnDef],
    hidden: &[ColumnDef],
    (key_columns, value_columns): (&[usize], &[usize]),
    checks: &[Check],
    indexes: &[IndexDef],
    primary: &Primary,
    comment: &str,
) -> Result<Vec<u8>> {
    let all: Vec<(&ColumnDef, bool)> =
        columns.iter().map(|c| (c, false)).chain(hidden.iter().map(|c| (c, true))).collect();
    let types: Vec<Vec<u8>> = all
        .iter()
        .map(|(c, _)| match c.mysql_type.is_empty() {
            true => c.ty.serialized(c.legacy_array).map(String::into_bytes),
            false => Ok(c.mysql_type.clone().into_bytes()),
        })
        .collect::<Result<_>>()?;
    let fields = all
        .iter()
        .zip(&types)
        .map(|(&(c, is_hidden), sql_type)| ColumnFields {
            name: c.name.as_bytes(),
            sql_type,
            default_value: c.default.as_bytes(),
            comment: c.comment.as_bytes(),
            on_update: b"",
            tag: c.tag,
            encoding: c.encoding,
            primary_key: c.primary_key,
            auto_increment: false,
            nullable: c.nullable,
            generated: c.generated,
            is_virtual: is_hidden,
            adaptive_encoding: crate::storage::marks_adaptive(c.encoding),
            hidden: false,
            hidden_system: is_hidden,
            identity: c.identity,
        })
        .collect();
    let stored = |i: usize| i.checked_sub(HIDDEN_BASE).map_or(i, |k| columns.len() + k) as u16;
    let keyless = key_columns.is_empty();
    let mut sorted: Vec<&IndexDef> = indexes.iter().collect();
    sorted.sort_by(|a, b| a.name.cmp(&b.name));
    let index_fields = sorted
        .into_iter()
        .map(|index| {
            let mut keys: Vec<u16> = index.columns.iter().map(|&i| stored(i)).collect();
            keys.extend(key_columns.iter().filter(|c| !index.columns.contains(c)).map(|&i| i as u16));
            let ordered = !index.system && index.descending.iter().chain(&index.nulls_last).any(|&o| o);
            serial::write::IndexFields {
                name: index.name.as_bytes(),
                comment: index.comment.as_bytes(),
                predicate: index.predicate.as_bytes(),
                index_columns: index.columns.iter().map(|&i| stored(i)).collect(),
                key_columns: keys,
                prefix_lengths: Vec::new(),
                descending: if ordered { index.descending.clone() } else { Vec::new() },
                nulls_last: if ordered { index.nulls_last.clone() } else { Vec::new() },
                op_classes: if index.op_classes.iter().all(String::is_empty) {
                    Vec::new()
                } else {
                    index.op_classes.iter().map(String::as_bytes).collect()
                },
                unique: index.unique,
                deferrable: index.deferrable,
                initially_deferred: index.initially_deferred,
                plain: index.plain,
                system_defined: index.system,
                spatial: false,
                fulltext: None,
                vector_distance: index.vector.map(|d| d as u8),
            }
        })
        .collect();
    // A keyless table's hidden hash and cardinality columns follow the others.
    let virtual_columns = (columns.len()..columns.len() + hidden.len()).map(|i| i as u16);
    let (key_columns, value_columns): (Vec<u16>, Vec<u16>) = if keyless {
        let n = (columns.len() + hidden.len()) as u16;
        let values = std::iter::once(n + 1).chain(value_columns.iter().map(|&i| i as u16)).chain(virtual_columns);
        (vec![n], values.collect())
    } else {
        let values = value_columns.iter().map(|&i| i as u16).chain(virtual_columns);
        (key_columns.iter().map(|&i| i as u16).collect(), values.collect())
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
        primary_key_name: primary.name.as_bytes(),
        primary_deferrable: primary.deferrable,
        primary_initially_deferred: primary.initially_deferred,
        collation: COLLATION,
        comment: comment.as_bytes(),
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
