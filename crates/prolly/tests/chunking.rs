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

//! Every tree node in the store crate's fixtures, which Go wrote, is rebuilt by the chunker from the leaf items under
//! it, since a node's items always split where they did when Go built it.

use std::collections::{HashMap, HashSet};
use std::path::Path;
use std::sync::Arc;

use prolly::{
    AddressMapSerializer, Chunker, CommitClosureSerializer, MergeArtifactsSerializer, Node, NodeSerializer, NodeStore,
    ProllyMapSerializer,
};
use serial::Message;
use store::{GenerationalStore, Hash};

/// RelativeOffsets maps an item to the positions of addresses within it.
type RelativeOffsets = HashMap<Vec<u8>, Vec<u16>>;

/// collect_leaves calls the function with each leaf node under the node, in key order.
fn collect_leaves(store: &GenerationalStore, node: &Node, f: &mut dyn FnMut(&Node)) {
    if node.is_leaf() {
        f(node);
        return;
    }
    for i in 0..node.count() {
        collect_leaves(store, &Node::load(store, &node.child(i).unwrap()).unwrap(), f);
    }
}

/// record_offsets records the positions of addresses within each item of a leaf, from the node's vector of
/// positions within the concatenated items.
fn record_offsets(node: &Node, field: usize, items: &[&[u8]], into: &mut RelativeOffsets) {
    let Some(vector) = Message(node.bytes()).root().unwrap().vector(field, 2).unwrap() else {
        for item in items {
            into.entry(item.to_vec()).or_default();
        }
        return;
    };
    let offsets: Vec<u16> = (0..vector.len()).map(|i| vector.u16(i).unwrap()).collect();
    let mut start = 0;
    for item in items {
        let end = start + item.len();
        let relative = offsets.iter().filter(|&&o| (start..end).contains(&(o as usize))).map(|&o| o - start as u16);
        into.insert(item.to_vec(), relative.collect());
        start = end;
    }
}

/// rebuild builds a tree from the items under the node with a serializer for its kind, returning the root's address
/// and the number of nodes written that the store lacks.
fn rebuild(store: &GenerationalStore, node: &Node) -> (Hash, usize) {
    let mut items: Vec<(Vec<u8>, Vec<u8>)> = Vec::new();
    let (mut key_offsets, mut value_offsets) = (RelativeOffsets::new(), RelativeOffsets::new());
    let file_id = node.file_id().to_string();
    collect_leaves(store, node, &mut |leaf| {
        let keys: Vec<&[u8]> = (0..leaf.count()).map(|i| leaf.key(i).unwrap()).collect();
        let values: Vec<&[u8]> = (0..leaf.count()).map(|i| leaf.value(i).unwrap()).collect();
        if file_id == serial::PROLLY_TREE_NODE {
            record_offsets(leaf, 11, &keys, &mut key_offsets);
            record_offsets(leaf, 6, &values, &mut value_offsets);
        } else if file_id == serial::MERGE_ARTIFACTS {
            record_offsets(leaf, 2, &keys, &mut key_offsets);
        }
        items.extend(keys.iter().zip(&values).map(|(k, v)| (k.to_vec(), v.to_vec())));
    });
    let key_addresses = |key: &[u8]| key_offsets[key].clone();
    let value_addresses = |value: &[u8]| value_offsets[value].clone();
    match file_id.as_str() {
        serial::ADDRESS_MAP => build(store, AddressMapSerializer, &items),
        serial::COMMIT_CLOSURE => build(store, CommitClosureSerializer, &items),
        serial::PROLLY_TREE_NODE => build(store, ProllyMapSerializer { key_addresses, value_addresses }, &items),
        _ => build(store, MergeArtifactsSerializer { key_addresses }, &items),
    }
}

/// build adds the items to a chunker with the serializer.
fn build<S: NodeSerializer>(store: &GenerationalStore, serializer: S, items: &[(Vec<u8>, Vec<u8>)]) -> (Hash, usize) {
    let mut nodes = FixtureNodes { store, written: HashMap::new(), missing: 0 };
    let mut chunker = Chunker::new(serializer, &mut nodes);
    for (key, value) in items {
        chunker.add(key, value).unwrap();
    }
    let (root, _) = chunker.done().unwrap();
    (root, nodes.missing)
}

/// FixtureNodes reads a fixture's nodes and keeps the nodes a chunker writes, counting those the fixture lacks.
struct FixtureNodes<'s> {
    store: &'s GenerationalStore,
    written: HashMap<Hash, Arc<Node>>,
    missing: usize,
}

impl NodeStore for FixtureNodes<'_> {
    fn read(&mut self, hash: &Hash) -> store::Result<Arc<Node>> {
        match self.written.get(hash) {
            Some(node) => Ok(node.clone()),
            None => Ok(Arc::new(Node::load(self.store, hash)?)),
        }
    }

    fn write(&mut self, hash: Hash, bytes: Vec<u8>) -> store::Result<Arc<Node>> {
        if self.store.get(&hash)?.is_none() {
            self.missing += 1;
        }
        let node = Arc::new(Node::decode(bytes)?);
        self.written.insert(hash, node.clone());
        Ok(node)
    }
}

#[test]
fn chunker_rebuilds_every_tree_node_go_wrote() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("../store/tests/fixtures");
    let kinds = [serial::PROLLY_TREE_NODE, serial::ADDRESS_MAP, serial::COMMIT_CLOSURE, serial::MERGE_ARTIFACTS];
    let mut checked: HashMap<String, usize> = HashMap::new();
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
            let mut nodes = Vec::new();
            for generation in [&store.new_gen, &store.old_gen] {
                generation
                    .for_each(&mut |chunk| {
                        if kinds.contains(&Message(&chunk.data).file_id()) {
                            nodes.push(chunk);
                        }
                        Ok(())
                    })
                    .unwrap();
            }
            for chunk in nodes {
                let node = Node::decode(chunk.data.clone()).unwrap();
                // An internal node with one child is only ever the last node of its level, whose rebuilt tree is
                // the child.
                if !node.is_leaf() && node.count() == 1 {
                    continue;
                }
                // Older Dolt versions wrote an empty vector of key addresses in artifact nodes, which Dolt no longer
                // writes.
                let message = Message(&chunk.data);
                if message.file_id() == serial::MERGE_ARTIFACTS
                    && message.root().unwrap().vector(2, 2).unwrap().is_some_and(|v| v.is_empty())
                {
                    continue;
                }
                let (root, missing) = rebuild(&store, &node);
                *checked.entry(format!("{} level {}", node.file_id(), node.level())).or_default() += 1;
                if root != chunk.hash || missing > 0 {
                    failures.push(format!(
                        "{}: {} {} level {} rebuilt as {root} with {missing} unknown nodes",
                        noms.display(),
                        chunk.hash,
                        node.file_id(),
                        node.level()
                    ));
                }
            }
        }
    }
    assert!(checked.values().sum::<usize>() > 1000, "only {checked:?} nodes were checked");
    assert!(failures.is_empty(), "{} differ ({checked:?}):\n{}", failures.len(), failures.join("\n"));
}

#[test]
fn blob_builder_rebuilds_every_blob_go_wrote() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("../store/tests/fixtures");
    let mut checked: HashMap<u8, usize> = HashMap::new();
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
            let mut blobs = HashMap::new();
            let mut children = HashSet::new();
            for generation in [&store.new_gen, &store.old_gen] {
                generation
                    .for_each(&mut |chunk| {
                        let message = Message(&chunk.data);
                        if message.file_id() == serial::BLOB {
                            let blob = serial::Blob::new(message).unwrap();
                            let addresses = blob.address_array().unwrap().unwrap_or_default();
                            children.extend(serial::hashes(addresses).unwrap());
                            blobs.insert(chunk.hash, (blob.tree_level().unwrap(), addresses.is_empty()));
                        }
                        Ok(())
                    })
                    .unwrap();
            }
            for (hash, (level, empty)) in blobs {
                // Go also writes an empty internal node that nothing uses when a blob fills its levels exactly.
                if children.contains(&hash) || (level > 0 && empty) {
                    continue;
                }
                let data = prolly::read_blob(&store, &hash).unwrap();
                let mut missing = 0;
                let mut sink = |written: Hash, _: &[u8]| {
                    if store.get(&written)?.is_none() {
                        missing += 1;
                    }
                    Ok(())
                };
                let (root, _) = prolly::write_blob(&data, &mut sink).unwrap().unwrap();
                *checked.entry(level).or_default() += 1;
                if root != hash || missing > 0 {
                    failures.push(format!(
                        "{}: {hash} level {level} of {} bytes rebuilt as {root} with {missing} unknown nodes",
                        noms.display(),
                        data.len()
                    ));
                }
            }
        }
    }
    assert!(checked.contains_key(&2), "no blobs of three levels were checked: {checked:?}");
    assert!(failures.is_empty(), "{} differ ({checked:?}):\n{}", failures.len(), failures.join("\n"));
}
