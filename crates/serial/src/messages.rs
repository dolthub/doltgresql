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

//! Typed views of the message tables, whose field indexes follow the order of the fields in Dolt's and Doltgres'
//! flatbuffers schemas.

use store::Hash;

use crate::fb::{Result, Table, Vector};
use crate::{Message, *};

/// missing returns the error for a required field that is absent.
fn missing(field: &str) -> store::Error {
    store::Error::Corrupt(format!("required field {field} is missing"))
}

/// hash converts bytes into an address.
pub fn hash(bytes: &[u8]) -> Result<Hash> {
    bytes.try_into().map(Hash).map_err(|_| store::Error::Corrupt(format!("address has {} bytes", bytes.len())))
}

/// hashes splits concatenated addresses.
pub fn hashes(bytes: &[u8]) -> Result<Vec<Hash>> {
    if !bytes.len().is_multiple_of(Hash::LEN) {
        return Err(store::Error::Corrupt(format!("address array has {} bytes", bytes.len())));
    }
    bytes.chunks(Hash::LEN).map(hash).collect()
}

/// StoreRoot is the root of a database's chunk graph: an address map from ref names to commits, tags, and working sets.
pub struct StoreRoot<'a>(pub Table<'a>);

impl<'a> StoreRoot<'a> {
    pub fn new(message: Message<'a>) -> Result<StoreRoot<'a>> {
        message.expect(STORE_ROOT).map(StoreRoot)
    }

    /// address_map returns the serialized AddressMap of refs, which is absent in an empty database.
    pub fn address_map(&self) -> Result<Option<&'a [u8]>> {
        self.0.bytes(0)
    }
}

/// TreeNode is the shared shape of the tree messages: key items and either value items or child addresses.
pub struct TreeNode<'a> {
    pub table: Table<'a>,
    pub key_items: &'a [u8],
    /// The offsets of the key items, one more than the item count, which commit closures lack.
    pub key_offsets: Option<Vector<'a>>,
    pub value_items: Option<&'a [u8]>,
    pub value_offsets: Option<Vector<'a>>,
    pub address_array: Option<&'a [u8]>,
    pub subtree_counts: Option<&'a [u8]>,
    pub tree_count: u64,
    pub tree_level: u8,
    /// The width in bytes of each item offset: 4 in a vector index node, otherwise 2.
    pub offset_width: usize,
}

impl<'a> TreeNode<'a> {
    /// new reads a ProllyTreeNode, AddressMap, CommitClosure, MergeArtifacts, or VectorIndexNode message as a tree
    /// node.
    pub fn new(message: Message<'a>) -> Result<TreeNode<'a>> {
        let table = message.root()?;
        match message.file_id() {
            PROLLY_TREE_NODE => Ok(TreeNode {
                table,
                key_items: table.bytes(0)?.ok_or_else(|| missing("key_items"))?,
                key_offsets: Some(table.vector(1, 2)?.ok_or_else(|| missing("key_offsets"))?),
                value_items: table.bytes(3)?,
                value_offsets: table.vector(4, 2)?,
                address_array: table.bytes(7)?,
                subtree_counts: table.bytes(8)?,
                tree_count: table.u64(9, 0)?,
                tree_level: table.u8(10, 0)?,
                offset_width: 2,
            }),
            ADDRESS_MAP => Ok(TreeNode {
                table,
                key_items: table.bytes(0)?.ok_or_else(|| missing("key_items"))?,
                key_offsets: Some(table.vector(1, 2)?.ok_or_else(|| missing("key_offsets"))?),
                value_items: None,
                value_offsets: None,
                address_array: Some(table.bytes(2)?.ok_or_else(|| missing("address_array"))?),
                subtree_counts: table.bytes(3)?,
                tree_count: table.u64(4, 0)?,
                tree_level: table.u8(5, 0)?,
                offset_width: 2,
            }),
            MERGE_ARTIFACTS => Ok(TreeNode {
                table,
                key_items: table.bytes(0)?.ok_or_else(|| missing("key_items"))?,
                key_offsets: Some(table.vector(1, 2)?.ok_or_else(|| missing("key_offsets"))?),
                value_items: table.bytes(3)?,
                value_offsets: table.vector(4, 2)?,
                address_array: table.bytes(5)?,
                subtree_counts: table.bytes(6)?,
                tree_count: table.u64(7, 0)?,
                tree_level: table.u8(8, 0)?,
                offset_width: 2,
            }),
            COMMIT_CLOSURE => Ok(TreeNode {
                table,
                key_items: table.bytes(0)?.ok_or_else(|| missing("key_items"))?,
                key_offsets: None,
                value_items: None,
                value_offsets: None,
                address_array: table.bytes(1)?,
                subtree_counts: table.bytes(2)?,
                tree_count: table.u64(3, 0)?,
                tree_level: table.u8(4, 0)?,
                offset_width: 2,
            }),
            VECTOR_INDEX_NODE => Ok(TreeNode {
                table,
                key_items: table.bytes(0)?.ok_or_else(|| missing("key_items"))?,
                key_offsets: Some(table.vector(1, 4)?.ok_or_else(|| missing("key_offsets"))?),
                value_items: table.bytes(2)?,
                value_offsets: table.vector(3, 4)?,
                address_array: table.bytes(4)?,
                subtree_counts: table.bytes(5)?,
                tree_count: table.u64(6, 0)?,
                tree_level: table.u8(7, 0)?,
                offset_width: 4,
            }),
            other => Err(store::Error::Corrupt(format!("{other:?} is not a tree node message"))),
        }
    }
}

/// Commit is a commit: its root value, parents, and metadata.
pub struct Commit<'a>(pub Table<'a>);

impl<'a> Commit<'a> {
    pub fn new(message: Message<'a>) -> Result<Commit<'a>> {
        message.expect(COMMIT).map(Commit)
    }

    pub fn root(&self) -> Result<Hash> {
        hash(self.0.bytes(0)?.ok_or_else(|| missing("root"))?)
    }

    pub fn height(&self) -> Result<u64> {
        self.0.u64(1, 0)
    }

    pub fn parents(&self) -> Result<Vec<Hash>> {
        hashes(self.0.bytes(2)?.ok_or_else(|| missing("parent_addrs"))?)
    }

    /// parent_closure_bytes returns the address bytes of the commit closure of the parents, which are absent or empty
    /// when there is none, and all zero for a commit without parents.
    pub fn parent_closure_bytes(&self) -> Result<Option<&'a [u8]>> {
        Ok(self.0.bytes(3)?.filter(|bytes| !bytes.is_empty()))
    }

    pub fn name(&self) -> Result<&'a [u8]> {
        self.0.string(4)?.ok_or_else(|| missing("name"))
    }

    pub fn email(&self) -> Result<&'a [u8]> {
        self.0.string(5)?.ok_or_else(|| missing("email"))
    }

    pub fn description(&self) -> Result<&'a [u8]> {
        self.0.string(6)?.ok_or_else(|| missing("description"))
    }

    pub fn timestamp_millis(&self) -> Result<u64> {
        self.0.u64(7, 0)
    }

    pub fn user_timestamp_millis(&self) -> Result<i64> {
        self.0.i64(8, 0)
    }

    pub fn signature(&self) -> Result<Option<&'a [u8]>> {
        self.0.string(9)
    }

    pub fn committer_name(&self) -> Result<Option<&'a [u8]>> {
        self.0.string(10)
    }

    pub fn committer_email(&self) -> Result<Option<&'a [u8]>> {
        self.0.string(11)
    }
}

/// Tag is a tag: the commit it names and its metadata.
pub struct Tag<'a>(pub Table<'a>);

impl<'a> Tag<'a> {
    pub fn new(message: Message<'a>) -> Result<Tag<'a>> {
        message.expect(TAG).map(Tag)
    }

    pub fn commit(&self) -> Result<Hash> {
        hash(self.0.bytes(0)?.ok_or_else(|| missing("commit_addr"))?)
    }

    pub fn name(&self) -> Result<&'a [u8]> {
        self.0.string(1)?.ok_or_else(|| missing("name"))
    }

    pub fn email(&self) -> Result<&'a [u8]> {
        self.0.string(2)?.ok_or_else(|| missing("email"))
    }

    pub fn description(&self) -> Result<&'a [u8]> {
        self.0.string(3)?.ok_or_else(|| missing("desc"))
    }

    pub fn timestamp_millis(&self) -> Result<u64> {
        self.0.u64(4, 0)
    }

    pub fn user_timestamp_millis(&self) -> Result<i64> {
        self.0.i64(5, 0)
    }
}

/// WorkingSet is a branch's working and staged roots, with any merge or rebase in progress.
pub struct WorkingSet<'a>(pub Table<'a>);

impl<'a> WorkingSet<'a> {
    pub fn new(message: Message<'a>) -> Result<WorkingSet<'a>> {
        message.expect(WORKING_SET).map(WorkingSet)
    }

    pub fn working_root(&self) -> Result<Hash> {
        hash(self.0.bytes(0)?.ok_or_else(|| missing("working_root_addr"))?)
    }

    pub fn staged_root(&self) -> Result<Option<Hash>> {
        self.0.bytes(1)?.map(hash).transpose()
    }

    pub fn name(&self) -> Result<&'a [u8]> {
        self.0.string(2)?.ok_or_else(|| missing("name"))
    }

    pub fn email(&self) -> Result<&'a [u8]> {
        self.0.string(3)?.ok_or_else(|| missing("email"))
    }

    pub fn description(&self) -> Result<&'a [u8]> {
        self.0.string(4)?.ok_or_else(|| missing("desc"))
    }

    pub fn timestamp_millis(&self) -> Result<u64> {
        self.0.u64(5, 0)
    }

    pub fn merge_state(&self) -> Result<Option<Table<'a>>> {
        self.0.table(6)
    }

    pub fn rebase_state(&self) -> Result<Option<Table<'a>>> {
        self.0.table(7)
    }
}

/// RootObjectMaps is the address bytes of each root object collection's AddressMap, by field name.
pub type RootObjectMaps<'a> = Vec<(&'static str, Option<&'a [u8]>)>;

/// DoltgresRootValue is a Doltgres root value: the address maps of its tables and root objects, and its schemas.
pub struct DoltgresRootValue<'a>(pub Table<'a>);

impl<'a> DoltgresRootValue<'a> {
    pub fn new(message: Message<'a>) -> Result<DoltgresRootValue<'a>> {
        message.expect(DOLTGRES_ROOT_VALUE).map(DoltgresRootValue)
    }

    pub fn feature_version(&self) -> Result<i64> {
        self.0.i64(0, 0)
    }

    /// tables returns the serialized AddressMap of tables.
    pub fn tables(&self) -> Result<Option<&'a [u8]>> {
        self.0.bytes(1)
    }

    pub fn foreign_keys(&self) -> Result<Option<&'a [u8]>> {
        self.0.bytes(2)
    }

    pub fn collation(&self) -> Result<u16> {
        self.0.u16(3, 0)
    }

    /// schemas returns the names of the database's schemas.
    pub fn schemas(&self) -> Result<Vec<&'a [u8]>> {
        let Some(vector) = self.0.vector(4, 4)? else { return Ok(Vec::new()) };
        (0..vector.len()).map(|i| Ok(vector.table(i)?.string(0)?.unwrap_or_default())).collect()
    }

    /// root_object_maps returns the address bytes of the AddressMap of each root object collection, by field name,
    /// which are empty or all zero for an empty collection.
    pub fn root_object_maps(&self) -> Result<RootObjectMaps<'a>> {
        const FIELDS: [(&str, usize); 10] = [
            ("sequences", 5),
            ("types", 6),
            ("functions", 7),
            ("triggers", 8),
            ("extensions", 9),
            ("conflicts", 10),
            ("procedures", 11),
            ("casts", 12),
            ("operators", 13),
            ("aggregates", 14),
        ];
        FIELDS.iter().map(|&(name, field)| Ok((name, self.0.bytes(field)?))).collect()
    }
}

/// TableMessage is a table: its schema, primary index, secondary indexes, and conflict and violation state.
pub struct TableMessage<'a>(pub Table<'a>);

impl<'a> TableMessage<'a> {
    pub fn new(message: Message<'a>) -> Result<TableMessage<'a>> {
        message.expect(TABLE).map(TableMessage)
    }

    pub fn schema(&self) -> Result<Hash> {
        hash(self.0.bytes(0)?.ok_or_else(|| missing("schema"))?)
    }

    /// primary_index returns the serialized root node of the primary index.
    pub fn primary_index(&self) -> Result<&'a [u8]> {
        self.0.bytes(1)?.ok_or_else(|| missing("primary_index"))
    }

    /// secondary_indexes returns the serialized AddressMap of secondary indexes.
    pub fn secondary_indexes(&self) -> Result<Option<&'a [u8]>> {
        self.0.bytes(2)
    }

    pub fn auto_increment(&self) -> Result<u64> {
        self.0.u64(3, 0)
    }

    pub fn conflicts(&self) -> Result<Option<Table<'a>>> {
        self.0.table(4)
    }

    pub fn violations(&self) -> Result<Option<&'a [u8]>> {
        self.0.bytes(5)
    }

    pub fn artifacts(&self) -> Result<Option<&'a [u8]>> {
        self.0.bytes(6)
    }
}

/// u16_list reads a [uint16] field.
fn u16_list(table: &Table<'_>, field: usize) -> Result<Vec<u16>> {
    match table.vector(field, 2)? {
        Some(vector) => (0..vector.len()).map(|i| vector.u16(i)).collect(),
        None => Ok(Vec::new()),
    }
}

/// bool_list reads a [bool] field.
fn bool_list(table: &Table<'_>, field: usize) -> Result<Vec<bool>> {
    Ok(table.bytes(field)?.unwrap_or_default().iter().map(|&b| b != 0).collect())
}

/// TableSchema is a table's columns, indexes, and checks.
pub struct TableSchema<'a>(pub Table<'a>);

/// Column is a column of a table schema.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Column<'a> {
    pub name: &'a [u8],
    pub sql_type: &'a [u8],
    pub default_value: &'a [u8],
    pub comment: &'a [u8],
    pub display_order: i16,
    pub tag: u64,
    pub encoding: u8,
    pub primary_key: bool,
    pub nullable: bool,
    pub auto_increment: bool,
    pub hidden: bool,
    pub generated: bool,
    pub is_virtual: bool,
    pub on_update_value: &'a [u8],
    pub uses_adaptive_encoding: bool,
    pub hidden_system: bool,
    pub adaptive_encoding_breaking_change: bool,
}

/// Index is the clustered index or a secondary index of a table schema.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Index<'a> {
    pub name: &'a [u8],
    pub comment: &'a [u8],
    pub index_columns: Vec<u16>,
    pub key_columns: Vec<u16>,
    pub value_columns: Vec<u16>,
    pub primary_key: bool,
    pub unique_key: bool,
    pub system_defined: bool,
    pub prefix_lengths: Vec<u16>,
    pub spatial_key: bool,
    pub fulltext_key: bool,
    pub vector_key: bool,
    pub predicate: &'a [u8],
    pub descending: Vec<bool>,
    pub nulls_last: Vec<bool>,
    pub op_classes: Vec<&'a [u8]>,
    /// Whether the unique or primary key constraint is DEFERRABLE, a Doltgres field that Dolt lacks.
    pub deferrable: bool,
    /// Whether the unique or primary key constraint is INITIALLY DEFERRED, a Doltgres field that Dolt lacks.
    pub initially_deferred: bool,
    pub fulltext_info: Option<FulltextInfo<'a>>,
    /// The distance type of a vector index.
    pub vector_distance: Option<u8>,
}

/// FulltextInfo is the pseudo-index tables and key of a full-text index.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct FulltextInfo<'a> {
    pub config_table: &'a [u8],
    pub position_table: &'a [u8],
    pub doc_count_table: &'a [u8],
    pub global_count_table: &'a [u8],
    pub row_count_table: &'a [u8],
    pub key_type: u8,
    pub key_name: &'a [u8],
    pub key_positions: Vec<u16>,
}

/// CheckConstraint is a check constraint of a table schema.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CheckConstraint<'a> {
    pub name: &'a [u8],
    pub expression: &'a [u8],
    pub enforced: bool,
    pub is_not_valid: bool,
}

impl<'a> TableSchema<'a> {
    pub fn new(message: Message<'a>) -> Result<TableSchema<'a>> {
        message.expect(TABLE_SCHEMA).map(TableSchema)
    }

    pub fn columns(&self) -> Result<Vec<Column<'a>>> {
        let vector = self.0.vector(0, 4)?.ok_or_else(|| missing("columns"))?;
        (0..vector.len())
            .map(|i| {
                let t = vector.table(i)?;
                Ok(Column {
                    name: t.string(0)?.ok_or_else(|| missing("name"))?,
                    sql_type: t.string(1)?.unwrap_or_default(),
                    default_value: t.string(2)?.unwrap_or_default(),
                    comment: t.string(3)?.unwrap_or_default(),
                    display_order: t.i16(4, 0)?,
                    tag: t.u64(5, 0)?,
                    encoding: t.u8(6, 0)?,
                    primary_key: t.bool(7, false)?,
                    nullable: t.bool(8, false)?,
                    auto_increment: t.bool(9, false)?,
                    hidden: t.bool(10, false)?,
                    generated: t.bool(11, false)?,
                    is_virtual: t.bool(12, false)?,
                    on_update_value: t.string(13)?.unwrap_or_default(),
                    uses_adaptive_encoding: t.bool(14, false)?,
                    hidden_system: t.bool(15, false)?,
                    adaptive_encoding_breaking_change: t.bool(16, false)?,
                })
            })
            .collect()
    }

    /// index decodes an Index table.
    fn index(t: Table<'a>) -> Result<Index<'a>> {
        let op_classes = match t.vector(17, 4)? {
            Some(vector) => (0..vector.len()).map(|i| vector.string(i)).collect::<Result<Vec<_>>>()?,
            None => Vec::new(),
        };
        Ok(Index {
            name: t.string(0)?.unwrap_or_default(),
            comment: t.string(1)?.unwrap_or_default(),
            index_columns: u16_list(&t, 2)?,
            key_columns: u16_list(&t, 3)?,
            value_columns: u16_list(&t, 4)?,
            primary_key: t.bool(5, false)?,
            unique_key: t.bool(6, false)?,
            system_defined: t.bool(7, false)?,
            prefix_lengths: u16_list(&t, 8)?,
            spatial_key: t.bool(9, false)?,
            fulltext_key: t.bool(10, false)?,
            vector_key: t.bool(12, false)?,
            predicate: t.string(14)?.unwrap_or_default(),
            descending: bool_list(&t, 15)?,
            nulls_last: bool_list(&t, 16)?,
            op_classes,
            deferrable: t.bool(18, false)?,
            initially_deferred: t.bool(19, false)?,
            fulltext_info: match t.table(11)? {
                Some(f) => Some(FulltextInfo {
                    config_table: f.string(0)?.unwrap_or_default(),
                    position_table: f.string(1)?.unwrap_or_default(),
                    doc_count_table: f.string(2)?.unwrap_or_default(),
                    global_count_table: f.string(3)?.unwrap_or_default(),
                    row_count_table: f.string(4)?.unwrap_or_default(),
                    key_type: f.u8(5, 0)?,
                    key_name: f.string(6)?.unwrap_or_default(),
                    key_positions: u16_list(&f, 7)?,
                }),
                None => None,
            },
            vector_distance: t.table(13)?.map(|v| v.u8(0, 0)).transpose()?,
        })
    }

    pub fn clustered_index(&self) -> Result<Index<'a>> {
        TableSchema::index(self.0.table(1)?.ok_or_else(|| missing("clustered_index"))?)
    }

    pub fn secondary_indexes(&self) -> Result<Vec<Index<'a>>> {
        let Some(vector) = self.0.vector(2, 4)? else { return Ok(Vec::new()) };
        (0..vector.len()).map(|i| TableSchema::index(vector.table(i)?)).collect()
    }

    pub fn checks(&self) -> Result<Vec<CheckConstraint<'a>>> {
        let Some(vector) = self.0.vector(3, 4)? else { return Ok(Vec::new()) };
        (0..vector.len())
            .map(|i| {
                let t = vector.table(i)?;
                Ok(CheckConstraint {
                    name: t.string(0)?.unwrap_or_default(),
                    expression: t.string(1)?.unwrap_or_default(),
                    enforced: t.bool(2, false)?,
                    is_not_valid: t.bool(3, false)?,
                })
            })
            .collect()
    }

    pub fn collation(&self) -> Result<u16> {
        self.0.u16(4, 0)
    }

    pub fn comment(&self) -> Result<&'a [u8]> {
        Ok(self.0.string(6)?.unwrap_or_default())
    }

    pub fn target_row_size(&self) -> Result<u16> {
        self.0.u16(7, 2048)
    }
}

/// string_list reads a [string] field.
pub fn string_list<'a>(table: &Table<'a>, field: usize) -> Result<Vec<&'a [u8]>> {
    match table.vector(field, 4)? {
        Some(vector) => (0..vector.len()).map(|i| vector.string(i)).collect(),
        None => Ok(Vec::new()),
    }
}

/// u64_list reads a [uint64] field.
fn u64_list(table: &Table<'_>, field: usize) -> Result<Vec<u64>> {
    match table.vector(field, 8)? {
        Some(vector) => (0..vector.len()).map(|i| vector.u64(i)).collect(),
        None => Ok(Vec::new()),
    }
}

/// ForeignKey is a foreign key of a root value's foreign key collection.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ForeignKey<'a> {
    pub name: &'a [u8],
    pub child_table_name: &'a [u8],
    pub child_table_index: &'a [u8],
    pub child_table_columns: Vec<u64>,
    pub parent_table_name: &'a [u8],
    pub parent_table_index: &'a [u8],
    pub parent_table_columns: Vec<u64>,
    pub on_update: u8,
    pub on_delete: u8,
    pub unresolved_child_columns: Vec<&'a [u8]>,
    pub unresolved_parent_columns: Vec<&'a [u8]>,
    pub child_table_database_schema: Vec<&'a [u8]>,
    pub parent_table_database_schema: Vec<&'a [u8]>,
    pub is_not_valid: bool,
    pub match_type: u8,
    /// Whether the foreign key is DEFERRABLE, a Doltgres field that Dolt lacks.
    pub deferrable: bool,
    /// Whether the foreign key is INITIALLY DEFERRED, a Doltgres field that Dolt lacks.
    pub initially_deferred: bool,
}

/// foreign_keys decodes a ForeignKeyCollection message.
pub fn foreign_keys(message: Message<'_>) -> Result<Vec<ForeignKey<'_>>> {
    let collection = message.expect(FOREIGN_KEY_COLLECTION)?;
    let Some(vector) = collection.vector(0, 4)? else { return Ok(Vec::new()) };
    (0..vector.len())
        .map(|i| {
            let t = vector.table(i)?;
            Ok(ForeignKey {
                name: t.string(0)?.unwrap_or_default(),
                child_table_name: t.string(1)?.unwrap_or_default(),
                child_table_index: t.string(2)?.unwrap_or_default(),
                child_table_columns: u64_list(&t, 3)?,
                parent_table_name: t.string(4)?.unwrap_or_default(),
                parent_table_index: t.string(5)?.unwrap_or_default(),
                parent_table_columns: u64_list(&t, 6)?,
                on_update: t.u8(7, 0)?,
                on_delete: t.u8(8, 0)?,
                unresolved_child_columns: string_list(&t, 9)?,
                unresolved_parent_columns: string_list(&t, 10)?,
                child_table_database_schema: string_list(&t, 11)?,
                parent_table_database_schema: string_list(&t, 12)?,
                is_not_valid: t.bool(13, false)?,
                match_type: t.u8(14, 0)?,
                deferrable: t.bool(15, false)?,
                initially_deferred: t.bool(16, false)?,
            })
        })
        .collect()
}

/// MergeState is a working set's merge in progress.
pub struct MergeState<'a>(pub Table<'a>);

impl<'a> MergeState<'a> {
    pub fn pre_working_root(&self) -> Result<&'a [u8]> {
        Ok(self.0.bytes(0)?.unwrap_or_default())
    }

    pub fn from_commit(&self) -> Result<&'a [u8]> {
        Ok(self.0.bytes(1)?.unwrap_or_default())
    }

    pub fn from_commit_spec(&self) -> Result<&'a [u8]> {
        Ok(self.0.string(2)?.unwrap_or_default())
    }

    pub fn unmergable_tables(&self) -> Result<Vec<&'a [u8]>> {
        string_list(&self.0, 3)
    }

    pub fn is_cherry_pick(&self) -> Result<bool> {
        self.0.bool(4, false)
    }

    pub fn is_revert(&self) -> Result<bool> {
        self.0.bool(5, false)
    }

    pub fn pre_merge_head_commit(&self) -> Result<&'a [u8]> {
        Ok(self.0.bytes(6)?.unwrap_or_default())
    }

    pub fn pending_commit_hashes(&self) -> Result<Vec<&'a [u8]>> {
        string_list(&self.0, 7)
    }
}

/// RebaseState is a working set's rebase in progress.
pub struct RebaseState<'a>(pub Table<'a>);

impl<'a> RebaseState<'a> {
    pub fn pre_working_root(&self) -> Result<&'a [u8]> {
        Ok(self.0.bytes(0)?.unwrap_or_default())
    }

    pub fn branch(&self) -> Result<&'a [u8]> {
        Ok(self.0.bytes(1)?.unwrap_or_default())
    }

    pub fn onto_commit(&self) -> Result<&'a [u8]> {
        Ok(self.0.bytes(2)?.unwrap_or_default())
    }

    pub fn empty_commit_handling(&self) -> Result<u8> {
        self.0.u8(3, 0)
    }

    pub fn commit_becomes_empty_handling(&self) -> Result<u8> {
        self.0.u8(4, 0)
    }

    pub fn last_attempted_step(&self) -> Result<f32> {
        self.0.u32(5, 0).map(f32::from_bits)
    }

    pub fn rebasing_started(&self) -> Result<bool> {
        self.0.bool(6, false)
    }

    pub fn skip_verification(&self) -> Result<bool> {
        self.0.bool(7, false)
    }
}

/// StashList is the address map of a database's stashes.
pub struct StashList<'a>(pub Table<'a>);

impl<'a> StashList<'a> {
    pub fn new(message: Message<'a>) -> Result<StashList<'a>> {
        message.expect(STASH_LIST).map(StashList)
    }

    pub fn address_map(&self) -> Result<&'a [u8]> {
        Ok(self.0.bytes(0)?.unwrap_or_default())
    }
}

/// Stash is a stash: its root, the commit it was made on, and its metadata.
pub struct Stash<'a>(pub Table<'a>);

impl<'a> Stash<'a> {
    pub fn new(message: Message<'a>) -> Result<Stash<'a>> {
        message.expect(STASH).map(Stash)
    }

    pub fn root(&self) -> Result<&'a [u8]> {
        Ok(self.0.bytes(0)?.unwrap_or_default())
    }

    pub fn head_commit(&self) -> Result<&'a [u8]> {
        Ok(self.0.bytes(1)?.unwrap_or_default())
    }

    pub fn branch_name(&self) -> Result<&'a [u8]> {
        Ok(self.0.string(2)?.unwrap_or_default())
    }

    pub fn description(&self) -> Result<&'a [u8]> {
        Ok(self.0.string(3)?.unwrap_or_default())
    }

    pub fn tables_to_stage(&self) -> Result<Vec<&'a [u8]>> {
        string_list(&self.0, 4)
    }
}

/// Blob is a node of a blob tree: payload bytes at a leaf, and child addresses with their sizes above.
pub struct Blob<'a>(pub Table<'a>);

impl<'a> Blob<'a> {
    pub fn new(message: Message<'a>) -> Result<Blob<'a>> {
        message.expect(BLOB).map(Blob)
    }

    pub fn payload(&self) -> Result<Option<&'a [u8]>> {
        self.0.bytes(0)
    }

    pub fn address_array(&self) -> Result<Option<&'a [u8]>> {
        self.0.bytes(1)
    }

    pub fn subtree_sizes(&self) -> Result<Option<&'a [u8]>> {
        self.0.bytes(2)
    }

    pub fn tree_size(&self) -> Result<u64> {
        self.0.u64(3, 0)
    }

    pub fn tree_level(&self) -> Result<u8> {
        self.0.u8(4, 0)
    }
}
