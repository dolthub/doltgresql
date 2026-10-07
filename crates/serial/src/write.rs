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

//! Writers of the version-control messages, which build their fields in the same order as Dolt so that they write the
//! same bytes.

use store::Hash;

use crate::Builder;

/// string_vector writes a vector of strings, creating the strings from last to first as Dolt's
/// SerializeStringVector does.
pub fn string_vector(b: &mut Builder, strings: &[Vec<u8>]) -> u32 {
    let mut offsets = vec![0; strings.len()];
    for i in (0..strings.len()).rev() {
        offsets[i] = b.create_string(&strings[i]);
    }
    b.create_vector_of_tables(&offsets)
}

/// write_store_root writes a StoreRoot around a serialized AddressMap of refs.
pub fn write_store_root(address_map: &[u8]) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let map = b.create_byte_vector(address_map);
    b.start_object(1);
    b.add_offset(0, map);
    let root = b.end_object();
    b.finish_message(root, crate::STORE_ROOT)
}

/// CommitFields is the contents of a Commit message.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CommitFields {
    pub root: Hash,
    pub height: u64,
    pub parents: Vec<Hash>,
    /// The address of the parents' commit closure, all zero for a commit without parents.
    pub parent_closure: Hash,
    pub name: Vec<u8>,
    pub email: Vec<u8>,
    pub description: Vec<u8>,
    pub timestamp_millis: u64,
    pub user_timestamp_millis: i64,
    /// The signature, which is only written when it is not empty.
    pub signature: Vec<u8>,
    /// The committer's name, which is only written when it differs from the author's.
    pub committer_name: Option<Vec<u8>>,
    /// The committer's email, which is only written when it differs from the author's.
    pub committer_email: Option<Vec<u8>>,
}

/// write_commit writes a Commit as Dolt's commit_flatbuffer does.
pub fn write_commit(c: &CommitFields) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let root = b.create_byte_vector(&c.root.0);
    let parents: Vec<u8> = c.parents.iter().flat_map(|p| p.0).collect();
    b.prep(4, parents.len());
    let parents = b.create_byte_vector(&parents);
    let closure = b.create_byte_vector(&c.parent_closure.0);
    let name = b.create_string(&c.name);
    let email = b.create_string(&c.email);
    let description = b.create_string(&c.description);
    let signature = if c.signature.is_empty() { 0 } else { b.create_string(&c.signature) };
    let committer_name = c.committer_name.as_ref().map_or(0, |n| b.create_string(n));
    let committer_email = c.committer_email.as_ref().map_or(0, |e| b.create_string(e));
    b.start_object(12);
    b.add_offset(0, root);
    b.add_u64(1, c.height, 0);
    b.add_offset(2, parents);
    b.add_offset(3, closure);
    b.add_offset(4, name);
    b.add_offset(5, email);
    b.add_offset(6, description);
    b.add_u64(7, c.timestamp_millis, 0);
    b.add_i64(8, c.user_timestamp_millis, 0);
    b.add_offset(9, signature);
    b.add_offset(10, committer_name);
    b.add_offset(11, committer_email);
    let message = b.end_object();
    b.finish_message(message, crate::COMMIT)
}

/// Meta is the author and time of a tag or working set.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Meta {
    pub name: Vec<u8>,
    pub email: Vec<u8>,
    pub description: Vec<u8>,
    pub timestamp_millis: u64,
    pub user_timestamp_millis: i64,
}

/// write_tag writes a Tag as Dolt's tag_flatbuffer does.
pub fn write_tag(commit: Hash, meta: Option<&Meta>) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let address = b.create_byte_vector(&commit.0);
    let strings = meta.map(|m| (b.create_string(&m.name), b.create_string(&m.email), b.create_string(&m.description)));
    b.start_object(6);
    b.add_offset(0, address);
    if let (Some(m), Some((name, email, description))) = (meta, strings) {
        b.add_offset(1, name);
        b.add_offset(2, email);
        b.add_offset(3, description);
        b.add_u64(4, m.timestamp_millis, 0);
        b.add_i64(5, m.user_timestamp_millis, 0);
    }
    let root = b.end_object();
    b.finish_message(root, crate::TAG)
}

/// MergeStateFields is a working set's merge in progress.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct MergeStateFields {
    pub pre_working_root: Hash,
    pub from_commit: Hash,
    pub from_commit_spec: Vec<u8>,
    pub unmergable_tables: Vec<Vec<u8>>,
    pub is_cherry_pick: bool,
    pub is_revert: bool,
    pub pre_merge_head_commit: Option<Hash>,
    /// The hashes of pending revert commits, which are only written when there are some.
    pub pending_commit_hashes: Vec<Vec<u8>>,
}

/// RebaseStateFields is a working set's rebase in progress.
#[derive(Clone, Debug, PartialEq)]
pub struct RebaseStateFields {
    pub pre_working_root: Hash,
    pub onto_commit: Hash,
    pub branch: Vec<u8>,
    pub commit_becomes_empty_handling: u8,
    pub empty_commit_handling: u8,
    pub last_attempted_step: f32,
    pub rebasing_started: bool,
    pub skip_verification: bool,
}

/// WorkingSetFields is the contents of a WorkingSet message.
#[derive(Clone, Debug, PartialEq)]
pub struct WorkingSetFields {
    pub working_root: Hash,
    pub staged_root: Option<Hash>,
    pub merge_state: Option<MergeStateFields>,
    pub rebase_state: Option<RebaseStateFields>,
    /// The author, which Dolt writes without the user timestamp.
    pub meta: Option<Meta>,
}

/// write_working_set writes a WorkingSet as Dolt's workingset_flatbuffer does.
pub fn write_working_set(w: &WorkingSetFields) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let working = b.create_byte_vector(&w.working_root.0);
    let staged = w.staged_root.map_or(0, |s| b.create_byte_vector(&s.0));
    let mut merge_state = 0;
    if let Some(m) = &w.merge_state {
        let pre_working = b.create_byte_vector(&m.pre_working_root.0);
        let from = b.create_byte_vector(&m.from_commit.0);
        let spec = b.create_string(&m.from_commit_spec);
        let unmergable = string_vector(&mut b, &m.unmergable_tables);
        let head = m.pre_merge_head_commit.map_or(0, |h| b.create_byte_vector(&h.0));
        let pending =
            if m.pending_commit_hashes.is_empty() { 0 } else { string_vector(&mut b, &m.pending_commit_hashes) };
        b.start_object(8);
        b.add_offset(0, pre_working);
        b.add_offset(1, from);
        b.add_offset(2, spec);
        b.add_offset(3, unmergable);
        b.add_bool(4, m.is_cherry_pick, false);
        b.add_bool(5, m.is_revert, false);
        b.add_offset(6, head);
        b.add_offset(7, pending);
        merge_state = b.end_object();
    }
    let mut rebase_state = 0;
    if let Some(r) = &w.rebase_state {
        let pre_working = b.create_byte_vector(&r.pre_working_root.0);
        let onto = b.create_byte_vector(&r.onto_commit.0);
        let branch = b.create_string(&r.branch);
        b.start_object(8);
        b.add_offset(0, pre_working);
        b.add_offset(1, branch);
        b.add_offset(2, onto);
        b.add_u8(4, r.commit_becomes_empty_handling, 0);
        b.add_u8(3, r.empty_commit_handling, 0);
        b.add_f32(5, r.last_attempted_step, 0.0);
        b.add_bool(6, r.rebasing_started, false);
        b.add_bool(7, r.skip_verification, false);
        rebase_state = b.end_object();
    }
    let strings =
        w.meta.as_ref().map(|m| (b.create_string(&m.name), b.create_string(&m.email), b.create_string(&m.description)));
    b.start_object(8);
    b.add_offset(0, working);
    b.add_offset(1, staged);
    b.add_offset(6, merge_state);
    b.add_offset(7, rebase_state);
    if let (Some(m), Some((name, email, description))) = (&w.meta, strings) {
        b.add_offset(2, name);
        b.add_offset(3, email);
        b.add_offset(4, description);
        b.add_u64(5, m.timestamp_millis, 0);
    }
    let root = b.end_object();
    b.finish_message(root, crate::WORKING_SET)
}

/// StashFields is the contents of a Stash message.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct StashFields {
    pub root: Hash,
    pub head_commit: Hash,
    pub branch_name: Vec<u8>,
    pub description: Vec<u8>,
    pub tables_to_stage: Option<Vec<Vec<u8>>>,
}

/// write_stash writes a Stash as Dolt's stash_flatbuffer does.
pub fn write_stash(s: &StashFields) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let root = b.create_byte_vector(&s.root.0);
    let head = b.create_byte_vector(&s.head_commit.0);
    let branch = b.create_string(&s.branch_name);
    let description = b.create_string(&s.description);
    let tables = s.tables_to_stage.as_ref().map_or(0, |t| string_vector(&mut b, t));
    b.start_object(5);
    b.add_offset(0, root);
    b.add_offset(1, head);
    b.add_offset(2, branch);
    b.add_offset(3, description);
    b.add_offset(4, tables);
    let message = b.end_object();
    b.finish_message(message, crate::STASH)
}

/// write_stash_list writes a StashList around a serialized AddressMap of stashes.
pub fn write_stash_list(address_map: &[u8]) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let map = b.create_byte_vector(address_map);
    b.start_object(1);
    b.add_offset(0, map);
    let root = b.end_object();
    b.finish_message(root, crate::STASH_LIST)
}

/// mutate_root_scalar overwrites a scalar field of a message's root table in place as Go's Mutate methods do,
/// returning false when the field is absent.
pub fn mutate_root_scalar(message: &mut [u8], field: usize, value: &[u8]) -> crate::Result<bool> {
    let Some(at) = crate::Message(message).root()?.offset(field)? else { return Ok(false) };
    message[at..at + value.len()].copy_from_slice(value);
    Ok(true)
}

/// mutate_root_bytes copies bytes over a byte vector field of a message's root table in place, as Dolt does when it
/// changes an address, returning false when the field is absent or has another length.
pub fn mutate_root_bytes(message: &mut [u8], field: usize, value: &[u8]) -> crate::Result<bool> {
    let Some(vector) = crate::Message(message).root()?.vector(field, 1)? else { return Ok(false) };
    if vector.len() != value.len() {
        return Ok(false);
    }
    let start = vector.start();
    message[start..start + value.len()].copy_from_slice(value);
    Ok(true)
}

/// TableFields is the contents of a Table message.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct TableFields<'a> {
    pub schema: Hash,
    /// The serialized root node of the primary index.
    pub primary_index: &'a [u8],
    /// The serialized AddressMap of secondary indexes.
    pub secondary_indexes: &'a [u8],
    pub auto_increment: u64,
    pub conflicts_data: &'a [u8],
    pub conflicts_ours: &'a [u8],
    pub conflicts_theirs: &'a [u8],
    pub conflicts_ancestor: &'a [u8],
    pub violations: &'a [u8],
    pub artifacts: &'a [u8],
}

/// write_table writes a Table as Dolt's serialTableFields.write does.
pub fn write_table(t: &TableFields<'_>) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let schema = b.create_byte_vector(&t.schema.0);
    let rows = b.create_byte_vector(t.primary_index);
    let indexes = b.create_byte_vector(t.secondary_indexes);
    let data = b.create_byte_vector(t.conflicts_data);
    let ours = b.create_byte_vector(t.conflicts_ours);
    let theirs = b.create_byte_vector(t.conflicts_theirs);
    let ancestor = b.create_byte_vector(t.conflicts_ancestor);
    b.start_object(4);
    b.add_offset(0, data);
    b.add_offset(1, ours);
    b.add_offset(2, theirs);
    b.add_offset(3, ancestor);
    let conflicts = b.end_object();
    let violations = b.create_byte_vector(t.violations);
    let artifacts = b.create_byte_vector(t.artifacts);
    b.start_object(7);
    b.add_offset(0, schema);
    b.add_offset(1, rows);
    b.add_offset(2, indexes);
    b.add_u64(3, t.auto_increment, 0);
    b.add_offset(4, conflicts);
    b.add_offset(5, violations);
    b.add_offset(6, artifacts);
    let root = b.end_object();
    b.finish_message(root, crate::TABLE)
}

/// ForeignKeyFields is a foreign key of a ForeignKeyCollection message.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ForeignKeyFields<'a> {
    pub name: &'a [u8],
    pub child_table_name: &'a [u8],
    pub child_table_index: &'a [u8],
    pub child_table_columns: Vec<u64>,
    pub parent_table_name: &'a [u8],
    pub parent_table_index: &'a [u8],
    pub parent_table_columns: Vec<u64>,
    pub on_update: u8,
    pub on_delete: u8,
    pub unresolved_child_columns: Option<Vec<Vec<u8>>>,
    pub unresolved_parent_columns: Option<Vec<Vec<u8>>>,
    pub is_not_valid: bool,
    pub match_type: u8,
    pub deferrable: bool,
    pub initially_deferred: bool,
}

/// u64_vector writes a vector of u64 values in order.
fn u64_vector(b: &mut Builder, values: &[u64]) -> u32 {
    b.start_vector(8, values.len(), 8);
    for &value in values.iter().rev() {
        b.prepend_u64(value);
    }
    b.end_vector(values.len())
}

/// u16_vector writes a vector of u16 values in order.
fn u16_vector(b: &mut Builder, values: &[u16]) -> u32 {
    b.start_vector(2, values.len(), 2);
    for &value in values.iter().rev() {
        b.prepend_u16(value);
    }
    b.end_vector(values.len())
}

/// bool_vector writes a vector of bools in order.
fn bool_vector(b: &mut Builder, values: &[bool]) -> u32 {
    b.start_vector(1, values.len(), 1);
    for &value in values.iter().rev() {
        b.prepend_u8(value as u8);
    }
    b.end_vector(values.len())
}

/// write_foreign_keys writes a ForeignKeyCollection as Dolt's serializeFlatbufferForeignKeys does.
pub fn write_foreign_keys(foreign_keys: &[ForeignKeyFields<'_>]) -> Vec<u8> {
    let mut b = Builder::new(2048);
    let mut offsets = vec![0; foreign_keys.len()];
    for (i, fk) in foreign_keys.iter().enumerate().rev() {
        let unresolved_parent = fk.unresolved_parent_columns.as_ref().map_or(0, |c| string_vector(&mut b, c));
        let unresolved_child = fk.unresolved_child_columns.as_ref().map_or(0, |c| string_vector(&mut b, c));
        let parent_columns = u64_vector(&mut b, &fk.parent_table_columns);
        let child_columns = u64_vector(&mut b, &fk.child_table_columns);
        let parent_table = b.create_string(fk.parent_table_name);
        let parent_index = b.create_string(fk.parent_table_index);
        let child_table = b.create_string(fk.child_table_name);
        let child_index = b.create_string(fk.child_table_index);
        let name = b.create_string(fk.name);
        b.start_object(17);
        b.add_offset(0, name);
        b.add_offset(1, child_table);
        b.add_offset(2, child_index);
        b.add_offset(3, child_columns);
        b.add_offset(4, parent_table);
        b.add_offset(5, parent_index);
        b.add_offset(6, parent_columns);
        b.add_offset(9, unresolved_child);
        b.add_offset(10, unresolved_parent);
        b.add_u8(7, fk.on_update, 0);
        b.add_u8(8, fk.on_delete, 0);
        b.add_bool(13, fk.is_not_valid, false);
        b.add_u8(14, fk.match_type, 0);
        b.add_bool(15, fk.deferrable, false);
        b.add_bool(16, fk.initially_deferred, false);
        offsets[i] = b.end_object();
    }
    let vector = b.create_vector_of_tables(&offsets);
    b.start_object(1);
    b.add_offset(0, vector);
    let root = b.end_object();
    b.finish_message(root, crate::FOREIGN_KEY_COLLECTION)
}

/// ColumnFields is a column of a TableSchema message, whose display order is its position.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ColumnFields<'a> {
    pub name: &'a [u8],
    pub sql_type: &'a [u8],
    /// The default value, or the generated expression of a generated column.
    pub default_value: &'a [u8],
    pub comment: &'a [u8],
    pub on_update: &'a [u8],
    pub tag: u64,
    pub encoding: u8,
    pub primary_key: bool,
    pub auto_increment: bool,
    pub nullable: bool,
    pub generated: bool,
    pub is_virtual: bool,
    pub adaptive_encoding: bool,
    pub hidden: bool,
    pub hidden_system: bool,
}

/// IndexFields is a secondary index of a TableSchema message, whose columns are positions in the columns.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct IndexFields<'a> {
    pub name: &'a [u8],
    pub comment: &'a [u8],
    pub predicate: &'a [u8],
    pub index_columns: Vec<u16>,
    pub key_columns: Vec<u16>,
    pub prefix_lengths: Vec<u16>,
    /// The descending flag of each column, which is written with nulls_last only when there are some.
    pub descending: Vec<bool>,
    pub nulls_last: Vec<bool>,
    pub op_classes: Vec<&'a [u8]>,
    pub unique: bool,
    pub deferrable: bool,
    pub initially_deferred: bool,
    pub system_defined: bool,
    pub spatial: bool,
    pub fulltext: Option<crate::FulltextInfo<'a>>,
    pub vector_distance: Option<u8>,
}

/// CheckFields is a check constraint of a TableSchema message.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CheckFields<'a> {
    pub name: &'a [u8],
    pub expression: &'a [u8],
    pub enforced: bool,
    pub is_not_valid: bool,
}

/// SchemaFields is the contents of a TableSchema message.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SchemaFields<'a> {
    pub columns: Vec<ColumnFields<'a>>,
    /// Whether the table has no primary key, which adds the hidden row ID and cardinality columns.
    pub keyless: bool,
    /// The clustered index's key columns, as positions in the columns.
    pub key_columns: Vec<u16>,
    /// The clustered index's value columns, as positions in the columns.
    pub value_columns: Vec<u16>,
    pub indexes: Vec<IndexFields<'a>>,
    pub checks: Vec<CheckFields<'a>>,
    /// The primary key constraint's name, a Doltgres field that Dolt lacks, empty for the default name.
    pub primary_key_name: &'a [u8],
    pub primary_deferrable: bool,
    pub primary_initially_deferred: bool,
    pub collation: u16,
    pub comment: &'a [u8],
    pub target_row_size: u16,
}

/// KEYLESS_ROW_ID_TAG is the tag of a keyless table's hidden row ID column.
pub const KEYLESS_ROW_ID_TAG: u64 = (1 << 51) + 5000;

/// KEYLESS_CARDINALITY_TAG is the tag of a keyless table's hidden cardinality column.
pub const KEYLESS_CARDINALITY_TAG: u64 = KEYLESS_ROW_ID_TAG + 1;

/// DEFAULT_TARGET_ROW_SIZE is the target row size that a schema omits.
pub const DEFAULT_TARGET_ROW_SIZE: u16 = 2048;

/// write_hidden_column writes one of a keyless table's hidden columns.
fn write_hidden_column(b: &mut Builder, name: &[u8], tag: u64, encoding: u8) -> u32 {
    let name = b.create_string(name);
    b.start_object(17);
    b.add_offset(0, name);
    b.add_i16(4, -1, 0);
    b.add_u64(5, tag, 0);
    b.add_u8(6, encoding, 0);
    b.add_bool(11, true, false);
    b.add_bool(10, true, false);
    b.end_object()
}

/// write_columns writes the columns vector as Dolt's serializeSchemaColumns does.
fn write_columns(b: &mut Builder, s: &SchemaFields<'_>) -> u32 {
    let mut offsets = vec![0; s.columns.len()];
    if s.keyless {
        let cardinality = write_hidden_column(b, b"keyless_cardinality", KEYLESS_CARDINALITY_TAG, 10);
        let id = write_hidden_column(b, b"keyless_hash_id", KEYLESS_ROW_ID_TAG, 14);
        offsets.extend([id, cardinality]);
    }
    for (i, c) in s.columns.iter().enumerate().rev() {
        let comment = b.create_string(c.comment);
        let default = b.create_string(c.default_value);
        let on_update = b.create_string(c.on_update);
        let sql_type = b.create_string(c.sql_type);
        let name = b.create_string(c.name);
        b.start_object(17);
        b.add_offset(0, name);
        b.add_offset(1, sql_type);
        b.add_offset(2, default);
        b.add_offset(3, comment);
        b.add_i16(4, i as i16, 0);
        b.add_u64(5, c.tag, 0);
        b.add_u8(6, c.encoding, 0);
        b.add_bool(7, c.primary_key, false);
        b.add_bool(9, c.auto_increment, false);
        b.add_bool(8, c.nullable, false);
        b.add_bool(11, c.generated, false);
        b.add_bool(12, c.is_virtual, false);
        if !c.on_update.is_empty() {
            b.add_offset(13, on_update);
        }
        if c.adaptive_encoding {
            b.add_bool(14, true, false);
            b.add_bool(16, true, false);
        }
        b.add_bool(10, c.hidden, false);
        b.add_bool(15, c.hidden_system, false);
        offsets[i] = b.end_object();
    }
    b.create_vector_of_tables(&offsets)
}

/// write_clustered_index writes the clustered index as Dolt's serializeClusteredIndex does.
fn write_clustered_index(b: &mut Builder, s: &SchemaFields<'_>) -> u32 {
    let name = if s.primary_key_name.is_empty() { 0 } else { b.create_string(s.primary_key_name) };
    let keys = u16_vector(b, &s.key_columns);
    let values = u16_vector(b, &s.value_columns);
    b.start_object(20);
    b.add_offset(0, name);
    b.add_offset(2, keys);
    b.add_offset(3, keys);
    b.add_offset(4, values);
    b.add_bool(5, true, false);
    b.add_bool(6, true, false);
    b.add_bool(9, false, false);
    b.add_bool(7, false, false);
    b.add_bool(18, s.primary_deferrable, false);
    b.add_bool(19, s.primary_initially_deferred, false);
    b.end_object()
}

/// write_fulltext_info writes a full-text index's info as Dolt's serializeFullTextInfo does.
fn write_fulltext_info(b: &mut Builder, f: &crate::FulltextInfo<'_>) -> u32 {
    let config = b.create_string(f.config_table);
    let position = b.create_string(f.position_table);
    let doc_count = b.create_string(f.doc_count_table);
    let global_count = b.create_string(f.global_count_table);
    let row_count = b.create_string(f.row_count_table);
    let key_name = b.create_string(f.key_name);
    let key_positions = u16_vector(b, &f.key_positions);
    b.start_object(8);
    b.add_offset(0, config);
    b.add_offset(1, position);
    b.add_offset(2, doc_count);
    b.add_offset(3, global_count);
    b.add_offset(4, row_count);
    b.add_u8(5, f.key_type, 0);
    b.add_offset(6, key_name);
    b.add_offset(7, key_positions);
    b.end_object()
}

/// write_secondary_indexes writes the secondary indexes vector as Dolt's serializeSecondaryIndexes does.
fn write_secondary_indexes(b: &mut Builder, indexes: &[IndexFields<'_>]) -> u32 {
    let mut offsets = vec![0; indexes.len()];
    for (i, index) in indexes.iter().enumerate().rev() {
        let name = b.create_string(index.name);
        let comment = b.create_string(index.comment);
        let predicate = if index.predicate.is_empty() { 0 } else { b.create_string(index.predicate) };
        let index_columns = u16_vector(b, &index.index_columns);
        let key_columns = u16_vector(b, &index.key_columns);
        let prefix_lengths = u16_vector(b, &index.prefix_lengths);
        let (mut descending, mut nulls_last) = (0, 0);
        if !index.descending.is_empty() {
            descending = bool_vector(b, &index.descending);
            nulls_last = bool_vector(b, &index.nulls_last);
        }
        let mut op_classes = 0;
        if !index.op_classes.is_empty() {
            let strings: Vec<u32> = index.op_classes.iter().map(|c| b.create_string(c)).collect();
            op_classes = b.create_vector_of_tables(&strings);
        }
        let fulltext = index.fulltext.as_ref().map_or(0, |f| write_fulltext_info(b, f));
        let vector = index.vector_distance.map_or(0, |distance| {
            b.start_object(1);
            b.add_u8(0, distance, 0);
            b.end_object()
        });
        b.start_object(20);
        b.add_offset(0, name);
        b.add_offset(1, comment);
        b.add_offset(2, index_columns);
        b.add_offset(3, key_columns);
        b.add_bool(5, false, false);
        b.add_bool(6, index.unique, false);
        b.add_bool(7, index.system_defined, false);
        b.add_offset(8, prefix_lengths);
        b.add_bool(9, index.spatial, false);
        b.add_bool(10, index.fulltext.is_some(), false);
        b.add_offset(11, fulltext);
        if index.vector_distance.is_some() {
            b.add_bool(12, true, false);
            b.add_offset(13, vector);
        }
        b.add_offset(14, predicate);
        b.add_offset(15, descending);
        b.add_offset(16, nulls_last);
        b.add_offset(17, op_classes);
        b.add_bool(18, index.deferrable, false);
        b.add_bool(19, index.initially_deferred, false);
        offsets[i] = b.end_object();
    }
    b.create_vector_of_tables(&offsets)
}

/// write_checks writes the checks vector as Dolt's serializeChecks does.
fn write_checks(b: &mut Builder, checks: &[CheckFields<'_>]) -> u32 {
    let mut offsets = vec![0; checks.len()];
    for (i, check) in checks.iter().enumerate().rev() {
        let expression = b.create_string(check.expression);
        let name = b.create_string(check.name);
        b.start_object(4);
        b.add_bool(2, check.enforced, false);
        b.add_offset(1, expression);
        b.add_offset(0, name);
        b.add_bool(3, check.is_not_valid, false);
        offsets[i] = b.end_object();
    }
    b.create_vector_of_tables(&offsets)
}

/// write_schema writes a TableSchema as Dolt's serializeSchemaAsFlatbuffer does.
pub fn write_schema(s: &SchemaFields<'_>) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let columns = write_columns(&mut b, s);
    let clustered = write_clustered_index(&mut b, s);
    let indexes = write_secondary_indexes(&mut b, &s.indexes);
    let checks = write_checks(&mut b, &s.checks);
    let comment = b.create_string(s.comment);
    let mut has_features = s.columns.iter().any(|c| !c.on_update.is_empty());
    b.start_object(8);
    b.add_offset(1, clustered);
    b.add_offset(0, columns);
    b.add_offset(2, indexes);
    b.add_offset(3, checks);
    b.add_u16(4, s.collation, 0);
    if !s.comment.is_empty() {
        b.add_offset(6, comment);
        has_features = true;
    }
    if s.target_row_size != DEFAULT_TARGET_ROW_SIZE {
        b.add_u16(7, s.target_row_size, DEFAULT_TARGET_ROW_SIZE);
        has_features = true;
    }
    b.add_bool(5, has_features, false);
    let root = b.end_object();
    b.finish_message(root, crate::TABLE_SCHEMA)
}

/// ROOT_OBJECT_COLLECTIONS is the number of Doltgres root object collections, whose address fields follow the
/// schemas field in ID order.
pub const ROOT_OBJECT_COLLECTIONS: usize = 10;

/// RootValueFields is the contents of a Doltgres root value.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct RootValueFields<'a> {
    pub feature_version: i64,
    pub collation: u16,
    /// The serialized AddressMap of tables.
    pub tables: &'a [u8],
    pub schemas: Vec<&'a [u8]>,
    pub foreign_keys: &'a [u8],
    /// The address of each root object collection's AddressMap by collection ID minus one, written when present.
    pub root_objects: [Option<Hash>; ROOT_OBJECT_COLLECTIONS],
    /// The root object collection whose field is added after the others, as Doltgres's SetRootObjectHash does when
    /// the message lacks the field.
    pub added_root_object: Option<usize>,
}

/// write_root_value writes a Doltgres root value as Doltgres's serializeRootValue does.
pub fn write_root_value(r: &RootValueFields<'_>) -> Vec<u8> {
    let mut b = Builder::new(80);
    let tables = b.create_byte_vector(r.tables);
    let mut schemas = 0;
    if !r.schemas.is_empty() {
        let mut offsets = vec![0; r.schemas.len()];
        for (i, schema) in r.schemas.iter().enumerate().rev() {
            let name = b.create_string(schema);
            b.start_object(1);
            b.add_offset(0, name);
            offsets[i] = b.end_object();
        }
        schemas = b.create_vector_of_tables(&offsets);
    }
    let foreign_keys = b.create_byte_vector(r.foreign_keys);
    let mut root_objects = [0; ROOT_OBJECT_COLLECTIONS];
    for (i, address) in r.root_objects.iter().enumerate() {
        if let Some(address) = address
            && r.added_root_object != Some(i)
        {
            root_objects[i] = b.create_byte_vector(&address.0);
        }
    }
    let added = r.added_root_object.map(|i| (i, b.create_byte_vector(&r.root_objects[i].unwrap_or_default().0)));
    b.start_object(5 + ROOT_OBJECT_COLLECTIONS);
    b.add_i64(0, r.feature_version, 0);
    b.add_u16(3, r.collation, 0);
    b.add_offset(1, tables);
    b.add_offset(2, foreign_keys);
    for (i, &offset) in root_objects.iter().enumerate() {
        b.add_offset(5 + i, offset);
    }
    if let Some((i, offset)) = added {
        b.add_offset(5 + i, offset);
    }
    b.add_offset(4, schemas);
    let root = b.end_object();
    b.finish_message(root, crate::DOLTGRES_ROOT_VALUE)
}
