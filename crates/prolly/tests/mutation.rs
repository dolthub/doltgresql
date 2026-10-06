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

//! Editing a tree in place builds the same tree as building its items from empty, since prolly trees are history
//! independent, which is what lets Go and Rust agree on every hash whichever way they reached a tree.

use std::collections::{BTreeMap, HashMap};
use std::sync::Arc;

use prolly::{AddressMapSerializer, Chunker, Node, NodeStore, apply_mutations};
use store::{Hash, Result};

/// MemoryNodes keeps nodes in memory.
#[derive(Default)]
struct MemoryNodes {
    nodes: HashMap<Hash, Arc<Node>>,
}

impl NodeStore for MemoryNodes {
    fn read(&mut self, hash: &Hash) -> Result<Arc<Node>> {
        Ok(self.nodes.get(hash).expect("node was never written").clone())
    }

    fn write(&mut self, hash: Hash, bytes: Vec<u8>) -> Result<Arc<Node>> {
        let node = Arc::new(Node::decode(bytes)?);
        self.nodes.insert(hash, node.clone());
        Ok(node)
    }
}

/// Random is a xorshift generator, so that each run edits the same way.
struct Random(u64);

impl Random {
    fn next(&mut self) -> u64 {
        self.0 ^= self.0 << 13;
        self.0 ^= self.0 >> 7;
        self.0 ^= self.0 << 17;
        self.0
    }

    fn below(&mut self, n: u64) -> u64 {
        self.next() % n
    }
}

/// key returns the key of an item number, which sorts as the number does.
fn key(n: u64) -> Vec<u8> {
    n.to_be_bytes().to_vec()
}

/// value returns a 20-byte value for an item number and version, as an address map holds.
fn value(n: u64, version: u64) -> Vec<u8> {
    Hash::of(format!("{n} {version}").as_bytes()).0.to_vec()
}

/// build builds a tree of the items from empty.
fn build(nodes: &mut MemoryNodes, items: &BTreeMap<Vec<u8>, Vec<u8>>) -> (Hash, Arc<Node>) {
    let mut chunker = Chunker::new(AddressMapSerializer, nodes);
    for (k, v) in items {
        chunker.add(k, v).unwrap();
    }
    chunker.done().unwrap()
}

/// apply edits the tree at the root, returning the new root.
fn apply(nodes: &mut MemoryNodes, root: Arc<Node>, edits: &BTreeMap<Vec<u8>, Option<Vec<u8>>>) -> (Hash, Arc<Node>) {
    let edits: Vec<(Vec<u8>, Option<Vec<u8>>)> = edits.iter().map(|(k, v)| (k.clone(), v.clone())).collect();
    apply_mutations(nodes, root, AddressMapSerializer, edits, &|a: &[u8], b: &[u8]| a.cmp(b)).unwrap()
}

/// height returns the number of levels of the tree at the root.
fn height(root: &Node) -> u8 {
    root.level() + 1
}

#[test]
fn edits_build_the_trees_that_building_from_empty_does() {
    let mut nodes = MemoryNodes::default();
    let mut random = Random(0x9e37_79b9_7f4a_7c15);
    let mut items: BTreeMap<Vec<u8>, Vec<u8>> = (0..60_000).map(|n| (key(n * 4), value(n * 4, 0))).collect();
    let (_, mut root) = build(&mut nodes, &items);
    assert!(height(&root) >= 3, "the tree has only {} levels", height(&root));
    let mut max_height = 0;
    for round in 1..=60u64 {
        // Rounds edit a few items, many items, runs of items, and the ends of the tree.
        let mut edits: BTreeMap<Vec<u8>, Option<Vec<u8>>> = BTreeMap::new();
        let count = match round % 4 {
            0 => 1,
            1 => 20,
            2 => 2_000,
            _ => 200,
        };
        let span = 260_000;
        let start = if round % 5 == 0 {
            0
        } else if round % 7 == 0 {
            span - 1_000
        } else {
            random.below(span)
        };
        for _ in 0..count {
            let n = if round % 3 == 0 { start + random.below(1_000) } else { random.below(span) };
            let edit = match random.below(3) {
                0 => None,
                _ => Some(value(n, round)),
            };
            edits.insert(key(n), edit);
        }
        for (k, v) in &edits {
            match v {
                Some(v) => items.insert(k.clone(), v.clone()),
                None => items.remove(k),
            };
        }
        let (edited_hash, edited) = apply(&mut nodes, root, &edits);
        let (built_hash, _) = build(&mut nodes, &items);
        assert_eq!(edited_hash, built_hash, "round {round} with {count} edits differs from building from empty");
        max_height = max_height.max(height(&edited));
        root = edited;
    }
    assert!(max_height >= 3, "the trees had at most {max_height} levels");
}

#[test]
fn edits_empty_and_regrow_trees() {
    let mut nodes = MemoryNodes::default();
    let mut items: BTreeMap<Vec<u8>, Vec<u8>> = (0..20_000).map(|n| (key(n), value(n, 0))).collect();
    let (_, root) = build(&mut nodes, &items);
    // Deleting everything leaves an empty leaf.
    let deletes: BTreeMap<Vec<u8>, Option<Vec<u8>>> = items.keys().map(|k| (k.clone(), None)).collect();
    let (empty_hash, empty) = apply(&mut nodes, root, &deletes);
    assert_eq!(empty_hash, build(&mut nodes, &BTreeMap::new()).0);
    assert_eq!(empty.count(), 0);
    // Inserting into an empty tree grows it back level by level.
    let mut root = empty;
    items.clear();
    for size in [1u64, 2, 50, 500, 5_000, 50_000] {
        let edits: BTreeMap<Vec<u8>, Option<Vec<u8>>> =
            (items.len() as u64..size).map(|n| (key(n), Some(value(n, 1)))).collect();
        for (k, v) in &edits {
            items.insert(k.clone(), v.clone().unwrap());
        }
        let (hash, edited) = apply(&mut nodes, root, &edits);
        assert_eq!(hash, build(&mut nodes, &items).0, "growing to {size} items");
        root = edited;
    }
    // Deleting all but one item from either end shrinks it back to a leaf.
    for keep in [0u64, 49_999] {
        let deletes: BTreeMap<Vec<u8>, Option<Vec<u8>>> =
            (0..50_000).filter(|&n| n != keep).map(|n| (key(n), None)).collect();
        let (hash, single) = apply(&mut nodes, root.clone(), &deletes);
        let only: BTreeMap<Vec<u8>, Vec<u8>> = [(key(keep), value(keep, 1))].into();
        assert_eq!(hash, build(&mut nodes, &only).0, "keeping {keep}");
        assert!(single.is_leaf());
    }
    // Edits that change nothing leave the tree as it is.
    let unchanged: BTreeMap<Vec<u8>, Option<Vec<u8>>> = [(key(10), Some(value(10, 1))), (key(1_000_000), None)].into();
    let (hash, _) = apply(&mut nodes, root.clone(), &unchanged);
    assert_eq!(hash, Hash::of(root.bytes()));
}
