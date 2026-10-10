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

//! Every version-control message in the store crate's fixtures, which Go wrote, writes back to the same bytes.

use std::collections::{BTreeMap, BTreeSet};
use std::path::Path;

use objects::{Kind, RootObject};
use serial::write::{
    CheckFields, ColumnFields, CommitFields, ForeignKeyFields, IndexFields, KEYLESS_CARDINALITY_TAG,
    KEYLESS_ROW_ID_TAG, MergeStateFields, Meta, ROOT_OBJECT_COLLECTIONS, RebaseStateFields, RootValueFields,
    SchemaFields, StashFields, TableFields, WorkingSetFields, mutate_root_scalar, write_commit, write_foreign_keys,
    write_root_value, write_schema, write_stash, write_stash_list, write_store_root, write_table, write_tag,
    write_working_set,
};
use serial::{
    Commit, DoltgresRootValue, MergeState, Message, RebaseState, Stash, StashList, StoreRoot, TableMessage,
    TableSchema, Tag, WorkingSet,
};
use store::{Chunk, ChunkReader, GenerationalStore};

/// owned copies a list of byte strings.
fn owned(values: Vec<&[u8]>) -> Vec<Vec<u8>> {
    values.into_iter().map(<[u8]>::to_vec).collect()
}

/// rewrite writes a version-control message again, returning None for other messages.
fn rewrite(chunk: &Chunk) -> Option<Vec<u8>> {
    let message = Message(&chunk.data);
    Some(match message.file_id() {
        serial::STORE_ROOT => {
            write_store_root(StoreRoot::new(message).unwrap().address_map().unwrap().unwrap_or_default())
        }
        serial::COMMIT => {
            let c = Commit::new(message).unwrap();
            let closure = c.parent_closure_bytes().unwrap().map(|b| serial::hash(b).unwrap()).unwrap_or_default();
            write_commit(&CommitFields {
                root: c.root().unwrap(),
                height: c.height().unwrap(),
                parents: c.parents().unwrap(),
                parent_closure: closure,
                name: c.name().unwrap().to_vec(),
                email: c.email().unwrap().to_vec(),
                description: c.description().unwrap().to_vec(),
                timestamp_millis: c.timestamp_millis().unwrap(),
                user_timestamp_millis: c.user_timestamp_millis().unwrap(),
                signature: c.signature().unwrap().unwrap_or_default().to_vec(),
                committer_name: c.committer_name().unwrap().map(<[u8]>::to_vec),
                committer_email: c.committer_email().unwrap().map(<[u8]>::to_vec),
            })
        }
        serial::TAG => {
            let t = Tag::new(message).unwrap();
            let meta = Meta {
                name: t.name().unwrap().to_vec(),
                email: t.email().unwrap().to_vec(),
                description: t.description().unwrap().to_vec(),
                timestamp_millis: t.timestamp_millis().unwrap(),
                user_timestamp_millis: t.user_timestamp_millis().unwrap(),
            };
            write_tag(t.commit().unwrap(), Some(&meta))
        }
        serial::WORKING_SET => {
            let w = WorkingSet::new(message).unwrap();
            let merge_state = w.merge_state().unwrap().map(|t| {
                let m = MergeState(t);
                MergeStateFields {
                    pre_working_root: serial::hash(m.pre_working_root().unwrap()).unwrap(),
                    from_commit: serial::hash(m.from_commit().unwrap()).unwrap(),
                    from_commit_spec: m.from_commit_spec().unwrap().to_vec(),
                    unmergable_tables: owned(m.unmergable_tables().unwrap()),
                    is_cherry_pick: m.is_cherry_pick().unwrap(),
                    is_revert: m.is_revert().unwrap(),
                    pre_merge_head_commit: Some(m.pre_merge_head_commit().unwrap())
                        .filter(|h| !h.is_empty())
                        .map(|h| serial::hash(h).unwrap()),
                    pending_commit_hashes: owned(m.pending_commit_hashes().unwrap()),
                }
            });
            let rebase_state = w.rebase_state().unwrap().map(|t| {
                let r = RebaseState(t);
                RebaseStateFields {
                    pre_working_root: serial::hash(r.pre_working_root().unwrap()).unwrap(),
                    onto_commit: serial::hash(r.onto_commit().unwrap()).unwrap(),
                    branch: r.branch().unwrap().to_vec(),
                    commit_becomes_empty_handling: r.commit_becomes_empty_handling().unwrap(),
                    empty_commit_handling: r.empty_commit_handling().unwrap(),
                    last_attempted_step: r.last_attempted_step().unwrap(),
                    rebasing_started: r.rebasing_started().unwrap(),
                    skip_verification: r.skip_verification().unwrap(),
                }
            });
            write_working_set(&WorkingSetFields {
                working_root: w.working_root().unwrap(),
                staged_root: w.staged_root().unwrap(),
                merge_state,
                rebase_state,
                meta: Some(Meta {
                    name: w.name().unwrap().to_vec(),
                    email: w.email().unwrap().to_vec(),
                    description: w.description().unwrap().to_vec(),
                    timestamp_millis: w.timestamp_millis().unwrap(),
                    user_timestamp_millis: 0,
                }),
            })
        }
        serial::STASH => {
            let s = Stash::new(message).unwrap();
            let tables = s.0.vector(4, 4).unwrap().map(|_| owned(s.tables_to_stage().unwrap()));
            write_stash(&StashFields {
                root: serial::hash(s.root().unwrap()).unwrap(),
                head_commit: serial::hash(s.head_commit().unwrap()).unwrap(),
                branch_name: s.branch_name().unwrap().to_vec(),
                description: s.description().unwrap().to_vec(),
                tables_to_stage: tables,
            })
        }
        serial::STASH_LIST => write_stash_list(StashList::new(message).unwrap().address_map().unwrap()),
        serial::TABLE => rewrite_table(message),
        serial::FOREIGN_KEY_COLLECTION => rewrite_foreign_keys(message),
        serial::TABLE_SCHEMA => rewrite_schema(message),
        serial::DOLTGRES_ROOT_VALUE => rewrite_root_value(message, &chunk.data),
        _ => return None,
    })
}

/// rewrite_table writes a Table again, setting an auto-increment value of zero in place as Dolt does when the field
/// already exists.
fn rewrite_table(message: Message<'_>) -> Vec<u8> {
    let t = TableMessage::new(message).unwrap();
    let conflicts = t.conflicts().unwrap().unwrap();
    let auto_increment = t.auto_increment().unwrap();
    let zeroed = auto_increment == 0 && t.0.offset(3).unwrap().is_some();
    let mut bytes = write_table(&TableFields {
        schema: t.schema().unwrap(),
        primary_index: t.primary_index().unwrap(),
        secondary_indexes: t.secondary_indexes().unwrap().unwrap(),
        auto_increment: if zeroed { 1 } else { auto_increment },
        conflicts_data: conflicts.bytes(0).unwrap().unwrap(),
        conflicts_ours: conflicts.bytes(1).unwrap().unwrap(),
        conflicts_theirs: conflicts.bytes(2).unwrap().unwrap(),
        conflicts_ancestor: conflicts.bytes(3).unwrap().unwrap(),
        violations: t.violations().unwrap().unwrap(),
        artifacts: t.artifacts().unwrap().unwrap(),
    });
    if zeroed {
        assert!(mutate_root_scalar(&mut bytes, 3, &0u64.to_le_bytes()).unwrap());
    }
    bytes
}

/// rewrite_foreign_keys writes a ForeignKeyCollection again.
fn rewrite_foreign_keys(message: Message<'_>) -> Vec<u8> {
    let tables = message.root().unwrap().vector(0, 4).unwrap();
    let keys = serial::foreign_keys(message).unwrap();
    let fields: Vec<ForeignKeyFields<'_>> = keys
        .iter()
        .enumerate()
        .map(|(i, fk)| {
            let t = tables.as_ref().unwrap().table(i).unwrap();
            ForeignKeyFields {
                name: fk.name,
                child_table_name: fk.child_table_name,
                child_table_index: fk.child_table_index,
                child_table_columns: fk.child_table_columns.clone(),
                parent_table_name: fk.parent_table_name,
                parent_table_index: fk.parent_table_index,
                parent_table_columns: fk.parent_table_columns.clone(),
                on_update: fk.on_update,
                on_delete: fk.on_delete,
                unresolved_child_columns: t.vector(9, 4).unwrap().map(|_| owned(fk.unresolved_child_columns.clone())),
                unresolved_parent_columns: t
                    .vector(10, 4)
                    .unwrap()
                    .map(|_| owned(fk.unresolved_parent_columns.clone())),
                is_not_valid: fk.is_not_valid,
                match_type: fk.match_type,
                deferrable: fk.deferrable,
                initially_deferred: fk.initially_deferred,
            }
        })
        .collect();
    write_foreign_keys(&fields)
}

/// rewrite_schema writes a TableSchema again.
fn rewrite_schema(message: Message<'_>) -> Vec<u8> {
    let s = TableSchema::new(message).unwrap();
    let mut columns = s.columns().unwrap();
    let count = columns.len();
    let keyless =
        count >= 2 && columns[count - 2].tag == KEYLESS_ROW_ID_TAG && columns[count - 1].tag == KEYLESS_CARDINALITY_TAG;
    if keyless {
        columns.truncate(count - 2);
    }
    let clustered = s.clustered_index().unwrap();
    write_schema(&SchemaFields {
        columns: columns
            .iter()
            .map(|c| {
                assert_eq!(c.uses_adaptive_encoding, c.adaptive_encoding_breaking_change);
                ColumnFields {
                    name: c.name,
                    sql_type: c.sql_type,
                    default_value: c.default_value,
                    comment: c.comment,
                    on_update: c.on_update_value,
                    tag: c.tag,
                    encoding: c.encoding,
                    primary_key: c.primary_key,
                    auto_increment: c.auto_increment,
                    nullable: c.nullable,
                    generated: c.generated,
                    is_virtual: c.is_virtual,
                    adaptive_encoding: c.uses_adaptive_encoding,
                    hidden: c.hidden,
                    hidden_system: c.hidden_system,
                    identity: c.identity,
                    not_null_name: c.not_null_name,
                }
            })
            .collect(),
        keyless,
        key_columns: clustered.key_columns,
        value_columns: clustered.value_columns,
        indexes: s
            .secondary_indexes()
            .unwrap()
            .into_iter()
            .map(|i| IndexFields {
                name: i.name,
                comment: i.comment,
                predicate: i.predicate,
                index_columns: i.index_columns,
                key_columns: i.key_columns,
                prefix_lengths: i.prefix_lengths,
                descending: i.descending,
                nulls_last: i.nulls_last,
                op_classes: i.op_classes,
                unique: i.unique_key,
                system_defined: i.system_defined,
                spatial: i.spatial_key,
                fulltext: i.fulltext_info,
                vector_distance: i.vector_distance,
                deferrable: i.deferrable,
                initially_deferred: i.initially_deferred,
                plain: i.plain,
            })
            .collect(),
        checks: s
            .checks()
            .unwrap()
            .into_iter()
            .map(|c| CheckFields {
                name: c.name,
                expression: c.expression,
                enforced: c.enforced,
                is_not_valid: c.is_not_valid,
            })
            .collect(),
        collation: s.collation().unwrap(),
        comment: s.comment().unwrap(),
        target_row_size: s.target_row_size().unwrap(),
        primary_key_name: clustered.name,
        primary_deferrable: clustered.deferrable,
        primary_initially_deferred: clustered.initially_deferred,
    })
}

/// rewrite_root_value writes a Doltgres root value again, also trying each root object collection as the one added
/// after the others, and returns the first rewrite that matches or else the plain one.
fn rewrite_root_value(message: Message<'_>, original: &[u8]) -> Vec<u8> {
    let r = DoltgresRootValue::new(message).unwrap();
    let mut root_objects = [None; ROOT_OBJECT_COLLECTIONS];
    for (i, (_, address)) in r.root_object_maps().unwrap().into_iter().enumerate() {
        root_objects[i] = address.map(|a| serial::hash(a).unwrap());
    }
    let mut fields = RootValueFields {
        feature_version: r.feature_version().unwrap(),
        collation: r.collation().unwrap(),
        tables: r.tables().unwrap().unwrap(),
        schemas: r.schemas().unwrap(),
        foreign_keys: r.foreign_keys().unwrap().unwrap(),
        root_objects,
        added_root_object: None,
    };
    let plain = write_root_value(&fields);
    if plain == original {
        return plain;
    }
    for i in (0..ROOT_OBJECT_COLLECTIONS).filter(|&i| root_objects[i].is_some()) {
        fields.added_root_object = Some(i);
        let bytes = write_root_value(&fields);
        if bytes == original {
            return bytes;
        }
    }
    plain
}

#[test]
fn version_control_messages_write_the_bytes_go_wrote() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("../store/tests/fixtures");
    let mut checked: BTreeMap<String, usize> = BTreeMap::new();
    let mut failures = Vec::new();
    for fixture in std::fs::read_dir(&fixtures).unwrap() {
        let fixture = fixture.unwrap().path();
        if !fixture.is_dir() {
            continue;
        }
        for database in std::fs::read_dir(&fixture).unwrap() {
            let noms = database.unwrap().path().join(".dolt/noms");
            if !noms.is_dir() {
                continue;
            }
            let store = GenerationalStore::open(&noms).unwrap();
            for generation in [&store.new_gen, &store.old_gen] {
                generation
                    .for_each(&mut |chunk| {
                        if let Some(bytes) = rewrite(&chunk) {
                            let file_id = Message(&chunk.data).file_id().to_string();
                            *checked.entry(file_id.clone()).or_default() += 1;
                            if bytes != chunk.data {
                                failures.push(format!("{}: {} {file_id}", noms.display(), chunk.hash));
                            }
                        }
                        Ok(())
                    })
                    .unwrap();
            }
        }
    }
    for kind in [
        serial::STORE_ROOT,
        serial::COMMIT,
        serial::TAG,
        serial::WORKING_SET,
        serial::STASH,
        serial::STASH_LIST,
        serial::TABLE,
        serial::FOREIGN_KEY_COLLECTION,
        serial::TABLE_SCHEMA,
        serial::DOLTGRES_ROOT_VALUE,
    ] {
        assert!(checked.contains_key(kind), "no {kind} messages were checked: {checked:?}");
    }
    assert!(failures.is_empty(), "{} differ ({checked:?}):\n{}", failures.len(), failures.join("\n"));
}

/// CURRENT_VERSIONS is the serialization version that Go writes for each root object collection, by the root value
/// field that holds it, where a type keeps the version it was read with and so never upgrades.
const CURRENT_VERSIONS: [(&str, u8); 10] = [
    ("sequences", 1),
    ("types", 0),
    ("functions", 4),
    ("triggers", 0),
    ("extensions", 1),
    ("conflicts", 0),
    ("procedures", 1),
    ("casts", 0),
    ("operators", 0),
    ("aggregates", 0),
];

#[test]
fn root_objects_serialize_to_the_bytes_go_wrote() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("../store/tests/fixtures");
    let mut checked: BTreeMap<&str, usize> = BTreeMap::new();
    let mut upgraded: BTreeMap<&str, usize> = BTreeMap::new();
    let mut failures = Vec::new();
    for fixture in std::fs::read_dir(&fixtures).unwrap() {
        let fixture = fixture.unwrap().path();
        if !fixture.is_dir() {
            continue;
        }
        for database in std::fs::read_dir(&fixture).unwrap() {
            let noms = database.unwrap().path().join(".dolt/noms");
            if !noms.is_dir() {
                continue;
            }
            let store = GenerationalStore::open(&noms).unwrap();
            let mut maps = BTreeSet::new();
            for generation in [&store.new_gen, &store.old_gen] {
                generation
                    .for_each(&mut |chunk| {
                        let message = Message(&chunk.data);
                        if message.file_id() == serial::DOLTGRES_ROOT_VALUE {
                            let root = DoltgresRootValue::new(message).unwrap();
                            for (name, address) in root.root_object_maps().unwrap() {
                                if let Some(address) = address.filter(|a| a.iter().any(|&b| b != 0)) {
                                    maps.insert((name, serial::hash(address).unwrap()));
                                }
                            }
                        }
                        Ok(())
                    })
                    .unwrap();
            }
            let mut seen = BTreeSet::new();
            for (name, map) in maps {
                let map = store.require(&map).unwrap().data;
                for (_, entry) in doltdb::address_map(&store, &map).unwrap() {
                    if !seen.insert(entry) {
                        continue;
                    }
                    let data = prolly::read_blob(&store, &entry).unwrap();
                    let kind = Kind::from_field(name).unwrap();
                    let object = RootObject::deserialize(kind, &data).unwrap();
                    let bytes = object.serialize();
                    *checked.entry(name).or_default() += 1;
                    if bytes == data {
                        continue;
                    }
                    let current = CURRENT_VERSIONS.iter().find(|(n, _)| *n == name).unwrap().1;
                    if data[0] < current && RootObject::deserialize(kind, &bytes).unwrap() == object {
                        *upgraded.entry(name).or_default() += 1;
                    } else {
                        failures.push(format!("{}: {entry} {name}", noms.display()));
                    }
                }
            }
        }
    }
    for (name, _) in CURRENT_VERSIONS {
        assert!(checked.contains_key(name), "no {name} were checked: {checked:?}");
    }
    assert!(
        failures.is_empty(),
        "{} differ ({checked:?}, upgraded {upgraded:?}):\n{}",
        failures.len(),
        failures.join("\n")
    );
}
