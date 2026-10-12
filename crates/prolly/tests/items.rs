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

//! Walking a tree's items forward and backward from either end or from a key visits every item in order, crossing
//! node boundaries at every level.

use std::collections::HashMap;
use std::sync::Arc;

use prolly::{AddressMapSerializer, Chunker, Items, Node, NodeStore};
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

/// key returns the key of an item number, which sorts as the number does.
fn key(n: u64) -> Vec<u8> {
    n.to_be_bytes().to_vec()
}

/// build builds a tree of the even numbers below twice the count, so that odd numbers fall between keys.
fn build(nodes: &mut MemoryNodes, count: u64) -> Arc<Node> {
    let mut chunker = Chunker::new(AddressMapSerializer, nodes);
    for n in 0..count {
        chunker.add(&key(n * 2), &Hash::of(&key(n)).0).unwrap();
    }
    chunker.done().unwrap().1
}

/// number returns the item number of the walk's current key.
fn number(items: &Items) -> Option<u64> {
    items.current().unwrap().map(|(k, _)| u64::from_be_bytes(k.try_into().unwrap()))
}

#[test]
fn walks_visit_every_item_in_both_directions() {
    let mut nodes = MemoryNodes::default();
    let count = 20_000;
    let root = build(&mut nodes, count);
    assert!(root.level() > 1);
    let mut items = Items::first(&mut nodes, root.clone()).unwrap();
    for n in 0..count {
        assert_eq!(number(&items), Some(n * 2));
        items.advance(&mut nodes).unwrap();
    }
    assert_eq!(number(&items), None);
    let mut items = Items::last(&mut nodes, root.clone()).unwrap();
    for n in (0..count).rev() {
        assert_eq!(number(&items), Some(n * 2));
        items.retreat(&mut nodes).unwrap();
    }
    assert_eq!(number(&items), None);
}

#[test]
fn walks_from_keys_start_at_the_first_key_not_less() {
    let mut nodes = MemoryNodes::default();
    let root = build(&mut nodes, 20_000);
    let compare = |a: &[u8], b: &[u8]| a.cmp(b);
    for target in [0, 1, 2, 777, 12_345, 39_998, 39_999] {
        let mut items = Items::at_key(&mut nodes, root.clone(), &key(target), &compare).unwrap();
        let expected = target.next_multiple_of(2);
        assert_eq!(number(&items), (expected < 40_000).then_some(expected));
        if expected > 0 && expected < 40_000 {
            items.retreat(&mut nodes).unwrap();
            assert_eq!(number(&items), Some(expected - 2));
            items.advance(&mut nodes).unwrap();
            items.advance(&mut nodes).unwrap();
            assert_eq!(number(&items), (expected + 2 < 40_000).then_some(expected + 2));
        }
    }
    let mut empty = MemoryNodes::default();
    let root = build(&mut empty, 0);
    assert_eq!(number(&Items::first(&mut empty, root.clone()).unwrap()), None);
    assert_eq!(number(&Items::last(&mut empty, root.clone()).unwrap()), None);
    assert_eq!(number(&Items::at_key(&mut empty, root, &key(5), &compare).unwrap()), None);
}

#[test]
fn seeks_move_walks_to_keys_in_either_direction() {
    let mut nodes = MemoryNodes::default();
    let root = build(&mut nodes, 20_000);
    let compare = |a: &[u8], b: &[u8]| a.cmp(b);
    let mut items = Items::first(&mut nodes, root).unwrap();
    for target in [3, 4, 5_000, 5_001, 39_998, 12, 0, 39_999, 20_000, 7] {
        items.seek(&mut nodes, &key(target), &compare).unwrap();
        let expected = target.next_multiple_of(2);
        assert_eq!(number(&items), (expected < 40_000).then_some(expected), "seeking {target}");
    }
}

#[test]
fn ordinals_count_the_items_before_walks() {
    let mut nodes = MemoryNodes::default();
    let root = build(&mut nodes, 20_000);
    let compare = |a: &[u8], b: &[u8]| a.cmp(b);
    for target in [0, 1, 2, 999, 1000, 25_001, 39_998, 39_999, 50_000] {
        let items = Items::at_key(&mut nodes, root.clone(), &key(target), &compare).unwrap();
        assert_eq!(items.ordinal().unwrap(), target.next_multiple_of(2).min(40_000) / 2, "at {target}");
    }
    let mut items = Items::last(&mut nodes, root.clone()).unwrap();
    assert_eq!(items.ordinal().unwrap(), 19_999);
    items.advance(&mut nodes).unwrap();
    assert_eq!(items.ordinal().unwrap(), 20_000);
}
