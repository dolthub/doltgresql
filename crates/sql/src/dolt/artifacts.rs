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

//! Merge artifacts: the conflicts and constraint violations that a merge leaves on a table's rows, stored in the
//! table's artifact map as Dolt stores them.

use std::cmp::Ordering;
use std::sync::Arc;

use base64::Engine;
use doltdb::database::Database;
use prolly::val::{build_tuple, encoding};
use prolly::{MergeArtifactsSerializer, Node, NodeStore, Tuple, apply_mutations, walk_leaves};
use store::Hash;

use crate::catalog::table::TableDef;
use crate::error::Result;

/// CONFLICT is the kind of a row that both sides of a merge changed differently.
pub const CONFLICT: u8 = 1;

/// FOREIGN_KEY is the kind of a row that violates a foreign key after a merge.
pub const FOREIGN_KEY: u8 = 2;

/// UNIQUE is the kind of a row that violates a unique index after a merge.
pub const UNIQUE: u8 = 3;

/// CHECK is the kind of a row that violates a check constraint after a merge.
pub const CHECK: u8 = 4;

/// NOT_NULL is the kind of a row that violates a NOT NULL constraint after a merge.
pub const NOT_NULL: u8 = 5;

/// INFO_HASH_SEED seeds the second hash of a violation's information, as Dolt's violationInfoHashSeed does.
const INFO_HASH_SEED: u64 = 0x9e37_79b9_7f4a_7c15;

/// Artifact is a conflict or constraint violation on a row: the row's key, the commit whose changes the merge
/// brought in, its kind, a hash that tells apart violations of one row, and its JSON metadata.
#[derive(Clone, Debug, PartialEq)]
pub struct Artifact {
    pub key: Vec<u8>,
    pub rootish: Hash,
    pub kind: u8,
    pub info_hash: Vec<u8>,
    pub meta: Vec<u8>,
}

/// info_hash returns the hash of a violation's information that tells apart violations of one row, as Dolt's
/// ConstraintViolationInfoHash computes it.
pub fn info_hash(info: &[u8]) -> Vec<u8> {
    let mut out = xxhash_rust::xxh3::xxh3_128(info).to_be_bytes().to_vec();
    out.extend_from_slice(&xxhash_rust::xxh3::xxh3_64_with_seed(info, INFO_HASH_SEED).to_be_bytes()[4..8]);
    out
}

/// conflict_meta returns the metadata of a conflict, which names the merge base, written as Go writes the array of a
/// hash in JSON.
pub fn conflict_meta(base: Hash) -> Vec<u8> {
    let bytes: Vec<String> = base.0.iter().map(u8::to_string).collect();
    format!("{{\"bc\":[{}]}}", bytes.join(",")).into_bytes()
}

/// conflict_base reads the merge base that a conflict's metadata names.
pub fn conflict_base(meta: &[u8]) -> Option<Hash> {
    let parsed: serde_json::Value = serde_json::from_slice(meta).ok()?;
    let bytes: Vec<u8> = parsed.get("bc")?.as_array()?.iter().filter_map(|b| b.as_u64().map(|b| b as u8)).collect();
    Some(Hash(bytes.try_into().ok()?))
}

/// violation_meta returns the metadata of a constraint violation: its information as JSON and the value of the
/// violating row, both as base64, as Go writes their bytes in JSON.
pub fn violation_meta(info: &[u8], value: &[u8]) -> Vec<u8> {
    let engine = base64::engine::general_purpose::STANDARD;
    format!("{{\"v_info\":\"{}\",\"value\":\"{}\"}}", engine.encode(info), engine.encode(value)).into_bytes()
}

/// violation_parts reads the information and the row value of a constraint violation's metadata.
pub fn violation_parts(meta: &[u8]) -> Option<(Vec<u8>, Vec<u8>)> {
    let parsed: serde_json::Value = serde_json::from_slice(meta).ok()?;
    let engine = base64::engine::general_purpose::STANDARD;
    let field = |name: &str| parsed.get(name).and_then(|v| v.as_str()).and_then(|s| engine.decode(s).ok());
    Some((field("v_info")?, field("value").unwrap_or_default()))
}

/// key_fields returns how many fields a table's row keys have.
fn key_fields(table: &TableDef) -> usize {
    table.key_encodings().len()
}

/// artifact_address_offsets returns the positions of the addresses within an artifact key whose fields have the
/// encodings, which include the commit, as Dolt's serializer finds them.
fn artifact_address_offsets(key: &[u8], encodings: &[u8]) -> Vec<u16> {
    let tuple = Tuple(key);
    let mut offsets = Vec::new();
    for (i, &field_encoding) in encodings.iter().enumerate() {
        let Ok(Some((start, end))) = tuple.field_range(i) else { continue };
        let address = (21..=27).contains(&field_encoding) && field_encoding != 25;
        if address && end > start && key[start..end].iter().any(|&b| b != 0) {
            offsets.push(start as u16);
        } else if (135..=139).contains(&field_encoding) && key[start] != 0 && end - start >= Hash::LEN {
            offsets.push((end - Hash::LEN) as u16);
        }
    }
    offsets
}

/// read returns a table's artifacts in key order.
pub fn read(db: &mut Database, table: &TableDef) -> Result<Vec<Artifact>> {
    let address = Hash(table.table.artifacts.as_slice().try_into().unwrap_or([0; Hash::LEN]));
    if address.0 == [0; Hash::LEN] {
        return Ok(Vec::new());
    }
    let node = Node::load(db, &address)?;
    let fields = key_fields(table);
    let mut artifacts = Vec::new();
    let mut failure = None;
    walk_leaves(db, &node, &mut |key, value| {
        let tuple = Tuple(key);
        let parse = || -> Result<Artifact> {
            let mut src: Vec<Option<&[u8]>> = Vec::with_capacity(fields);
            for i in 0..fields {
                src.push(tuple.field(i)?);
            }
            let rootish = tuple.field(fields)?.unwrap_or_default();
            Ok(Artifact {
                key: build_tuple(&src),
                rootish: Hash(rootish.try_into().unwrap_or([0; Hash::LEN])),
                kind: tuple.field(fields + 1)?.and_then(|b| b.first().copied()).unwrap_or_default(),
                info_hash: tuple.field(fields + 2)?.unwrap_or_default().to_vec(),
                meta: Tuple(value).field(0)?.unwrap_or_default().to_vec(),
            })
        };
        match parse() {
            Ok(artifact) => artifacts.push(artifact),
            Err(err) => failure = Some(err),
        }
        Ok(())
    })?;
    match failure {
        Some(err) => Err(err),
        None => Ok(artifacts),
    }
}

/// compare orders artifacts as Dolt's artifact map does: by row key, then commit, kind, and information hash.
fn compare(table: &TableDef, a: &Artifact, b: &Artifact) -> Ordering {
    table
        .compare_keys(&a.key, &b.key)
        .then_with(|| a.rootish.0.cmp(&b.rootish.0))
        .then_with(|| a.kind.cmp(&b.kind))
        .then_with(|| a.info_hash.cmp(&b.info_hash))
}

/// write writes a table's artifacts, returning the address to store in the table, which is zero without artifacts.
pub fn write(db: &mut Database, table: &TableDef, mut artifacts: Vec<Artifact>) -> Result<Vec<u8>> {
    if artifacts.is_empty() {
        return Ok(vec![0; Hash::LEN]);
    }
    artifacts.sort_by(|a, b| compare(table, a, b));
    artifacts.dedup_by(|a, b| compare(table, a, b) == Ordering::Equal);
    let fields = key_fields(table);
    let mut encodings = table.key_encodings();
    encodings.extend([encoding::COMMIT_ADDR, encoding::UINT8, encoding::BYTES]);
    let edits: Vec<(Vec<u8>, Option<Vec<u8>>)> = artifacts
        .iter()
        .map(|a| {
            let src = Tuple(&a.key);
            let mut parts: Vec<Option<&[u8]>> = (0..fields).map(|i| src.field(i).ok().flatten()).collect();
            let kind = [a.kind];
            parts.extend([Some(&a.rootish.0[..]), Some(&kind[..]), Some(&a.info_hash[..])]);
            (build_tuple(&parts), Some(build_tuple(&[Some(&a.meta[..])])))
        })
        .collect();
    let empty = prolly::serialize_merge_artifacts(&[], &[], &[], 0, None);
    let root = Arc::new(Node::decode(empty)?);
    let serializer = MergeArtifactsSerializer { key_addresses: |key: &[u8]| artifact_address_offsets(key, &encodings) };
    let compare_tuples = |a: &[u8], b: &[u8]| {
        let (ta, tb) = (Tuple(a), Tuple(b));
        table.compare_keys(a, b).then_with(|| {
            (fields..fields + 3)
                .map(|i| ta.field(i).ok().flatten().cmp(&tb.field(i).ok().flatten()))
                .find(|o| *o != Ordering::Equal)
                .unwrap_or(Ordering::Equal)
        })
    };
    let (address, _) = apply_mutations(db as &mut dyn NodeStore, root, serializer, edits, &compare_tuples)?;
    Ok(address.0.to_vec())
}
