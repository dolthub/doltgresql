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

//! Every tree node in the store crate's fixtures, which Go wrote, serializes back to the same bytes.

use std::path::Path;

use prolly::{
    Node, ProllyNode, serialize_address_map, serialize_blob, serialize_commit_closure, serialize_merge_artifacts,
    serialize_prolly_node,
};
use serial::{Blob, Message};
use store::{Chunk, GenerationalStore};

/// u16s reads a [uint16] field of a message's root table.
fn u16s(message: Message<'_>, field: usize) -> Vec<u16> {
    match message.root().unwrap().vector(field, 2).unwrap() {
        Some(vector) => (0..vector.len()).map(|i| vector.u16(i).unwrap()).collect(),
        None => Vec::new(),
    }
}

/// reserialize serializes a tree node chunk again, returning None for other messages.
fn reserialize(chunk: &Chunk) -> Option<Vec<u8>> {
    let message = Message(&chunk.data);
    let file_id = message.file_id();
    if file_id == serial::BLOB {
        let blob = Blob::new(message).unwrap();
        let level = blob.tree_level().unwrap();
        return Some(if level == 0 {
            let payload = blob.payload().unwrap().unwrap_or_default();
            serialize_blob(&[payload], &[blob.tree_size().unwrap()], 0)
        } else {
            let addresses: Vec<&[u8]> = blob.address_array().unwrap().unwrap_or_default().chunks(20).collect();
            let sizes = blob_sizes(blob.subtree_sizes().unwrap().unwrap_or_default(), addresses.len());
            serialize_blob(&addresses, &sizes, level)
        });
    }
    if ![serial::PROLLY_TREE_NODE, serial::ADDRESS_MAP, serial::COMMIT_CLOSURE, serial::MERGE_ARTIFACTS]
        .contains(&file_id)
    {
        return None;
    }
    let node = Node::decode(chunk.data.clone()).unwrap();
    let keys: Vec<&[u8]> = (0..node.count()).map(|i| node.key(i).unwrap()).collect();
    let values: Vec<&[u8]> = (0..node.count()).map(|i| node.value(i).unwrap()).collect();
    let subtrees = node.subtree_counts().unwrap();
    Some(match file_id {
        serial::PROLLY_TREE_NODE => serialize_prolly_node(&ProllyNode {
            keys,
            values,
            subtrees,
            level: node.level(),
            key_address_offsets: u16s(message, 11),
            value_address_offsets: u16s(message, 6),
        }),
        serial::ADDRESS_MAP => serialize_address_map(&keys, &values, &subtrees, node.level()),
        serial::MERGE_ARTIFACTS => {
            let present = message.root().unwrap().vector(2, 2).unwrap().is_some();
            let key_address_offsets = u16s(message, 2);
            serialize_merge_artifacts(
                &keys,
                &values,
                &subtrees,
                node.level(),
                present.then_some(&key_address_offsets[..]),
            )
        }
        _ => serialize_commit_closure(&keys, &values, &subtrees, node.level()),
    })
}

/// blob_sizes decodes the zigzag varint deltas of a blob's subtree sizes.
fn blob_sizes(mut bytes: &[u8], count: usize) -> Vec<u64> {
    let mut sizes = Vec::with_capacity(count);
    let mut previous: i64 = 0;
    for _ in 0..count {
        let mut value: u64 = 0;
        let mut shift = 0;
        loop {
            let byte = bytes[0];
            bytes = &bytes[1..];
            value |= ((byte & 0x7f) as u64) << shift;
            shift += 7;
            if byte < 0x80 {
                break;
            }
        }
        previous += (value >> 1) as i64 ^ -((value & 1) as i64);
        sizes.push(previous as u64);
    }
    sizes
}

#[test]
fn tree_nodes_serialize_to_the_bytes_go_wrote() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("../store/tests/fixtures");
    let mut checked = 0;
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
                        if let Some(bytes) = reserialize(&chunk) {
                            checked += 1;
                            if bytes != chunk.data
                                && std::env::var_os("ROUNDTRIP_DEBUG").is_some()
                                && failures.len() < 3
                            {
                                let hex = |b: &[u8]| b.iter().map(|x| format!("{x:02x}")).collect::<String>();
                                eprintln!("go   {}\nrust {}", hex(&chunk.data), hex(&bytes));
                            }
                            if bytes != chunk.data {
                                failures.push(format!(
                                    "{}: {} {}",
                                    noms.display(),
                                    chunk.hash,
                                    Message(&chunk.data).file_id()
                                ));
                            }
                        }
                        Ok(())
                    })
                    .unwrap();
            }
        }
    }
    assert!(checked > 1000, "only {checked} tree nodes were checked");
    assert!(failures.is_empty(), "{} of {checked} differ:\n{}", failures.len(), failures.join("\n"));
}
