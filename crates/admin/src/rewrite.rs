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

use std::collections::HashMap;

use doltdb::table::address_offsets;
use prolly::{Node, NodeSerializer, ProllyMapSerializer};
use sql::PgError;
use sql::integrity::{CacheKey, Scanner, Stats, analyze_internal_keys, analyze_leaf, tree_node};
use store::Hash;

/// LeafTransform returns the replacement of a leaf, or None when the leaf needs no rewrite, given the encodings of
/// its keys and values.
pub type LeafTransform = fn(&mut TreeRewriter<'_>, &Node, &[u8], &[u8]) -> sql::Result<Option<Vec<u8>>>;

/// InternalTransform returns the replacement of an internal node whose children became the new children, or None
/// when the node needs no rewrite, given whether any child changed and the encodings of its keys and values.
pub type InternalTransform =
    fn(&mut TreeRewriter<'_>, &Node, &[Hash], bool, &[u8], &[u8]) -> sql::Result<Option<Vec<u8>>>;

/// TreeRewriter rewrites the tree nodes that fail to record the addresses of their out-of-band values, visiting only
/// the trees that its scanner finds worth rewriting, so a rewrite always agrees with the scan. Leaves are serialized
/// again from their own tuples, which records every address, and the nodes above them then point at the new
/// children. Each tree is rewritten once per set of encodings, however many commits and branches share it.
pub struct TreeRewriter<'d> {
    pub scanner: Scanner<'d>,
    /// The rewrite of a leaf, which tests replace with a corrupting one to simulate old releases.
    pub transform_leaf: LeafTransform,
    /// The rewrite of an internal node.
    pub transform_internal: InternalTransform,
    /// Whether a tree with the scanner's counts holds anything to rewrite, so that other trees are skipped.
    pub should_rewrite: fn(&Stats) -> bool,
    cache: HashMap<CacheKey, Hash>,
    pub leaf_chunks_rewritten: u64,
    pub internal_chunks_rewritten: u64,
}

impl<'d> TreeRewriter<'d> {
    /// new returns a rewriter that repairs what the scanner finds.
    pub fn new(scanner: Scanner<'d>) -> TreeRewriter<'d> {
        TreeRewriter {
            scanner,
            transform_leaf: repair_leaf,
            transform_internal: repair_internal,
            should_rewrite: |stats| stats.corrupt_chunks > 0,
            cache: HashMap::new(),
            leaf_chunks_rewritten: 0,
            internal_chunks_rewritten: 0,
        }
    }

    /// rewrite_map_root rewrites a tree from its root node, which a table message holds rather than the database, and
    /// returns the new root's address, which is the old one when nothing changed.
    pub fn rewrite_map_root(&mut self, root: &[u8], keys: &[u8], values: &[u8]) -> sql::Result<Hash> {
        let address = Hash::of(root);
        let key = (address, keys.to_vec(), values.to_vec());
        if let Some(&rewritten) = self.cache.get(&key) {
            return Ok(rewritten);
        }
        let stats = self.scanner.scan_root_node(root, keys, values)?;
        if !(self.should_rewrite)(&stats) {
            self.cache.insert(key, address);
            return Ok(address);
        }
        self.rewrite_node(key, Node::decode(root.to_vec())?)
    }

    /// rewrite_tree rewrites the tree whose root node is the chunk at the address, and returns the new root's address.
    pub fn rewrite_tree(&mut self, address: Hash, keys: &[u8], values: &[u8]) -> sql::Result<Hash> {
        let key = (address, keys.to_vec(), values.to_vec());
        if let Some(&rewritten) = self.cache.get(&key) {
            return Ok(rewritten);
        }
        let stats = self.scanner.scan_tree(address, keys, values)?;
        if !(self.should_rewrite)(&stats) {
            self.cache.insert(key, address);
            return Ok(address);
        }
        let node = tree_node(self.scanner.db, &address)?;
        self.rewrite_node(key, node)
    }

    /// rewrite_node rewrites a node and the nodes below it, caching the new address under the key.
    fn rewrite_node(&mut self, key: CacheKey, node: Node) -> sql::Result<Hash> {
        let (address, keys, values) = (key.0, key.1.clone(), key.2.clone());
        let rewritten = if node.is_leaf() {
            match (self.transform_leaf)(self, &node, &keys, &values)
                .map_err(|e| PgError::internal(format!("failed to rewrite leaf node {address}: {}", e.message)))?
            {
                Some(bytes) => {
                    self.leaf_chunks_rewritten += 1;
                    self.write_node(bytes)?
                }
                None => address,
            }
        } else {
            let mut children = Vec::with_capacity(node.count());
            let mut changed = false;
            for i in 0..node.count() {
                let child = node.child(i)?;
                let new_child = self.rewrite_tree(child, &keys, &values)?;
                changed |= new_child != child;
                children.push(new_child);
            }
            match (self.transform_internal)(self, &node, &children, changed, &keys, &values)
                .map_err(|e| PgError::internal(format!("failed to rewrite internal node {address}: {}", e.message)))?
            {
                Some(bytes) => {
                    self.internal_chunks_rewritten += 1;
                    self.write_node(bytes)?
                }
                None => address,
            }
        };
        self.cache.insert(key, rewritten);
        Ok(rewritten)
    }

    /// write_node writes a tree node and returns its address.
    fn write_node(&mut self, bytes: Vec<u8>) -> sql::Result<Hash> {
        let file_id = Node::decode(bytes.clone())?.file_id().to_string();
        if file_id != serial::PROLLY_TREE_NODE {
            return Err(PgError::internal(format!("rewritten node has unexpected file ID {file_id}")));
        }
        Ok(self.scanner.db.write_value(bytes)?)
    }
}

/// repair_leaf serializes a corrupt leaf again from its own tuples, which records every address, and leaves a healthy
/// leaf alone.
pub fn repair_leaf(_: &mut TreeRewriter<'_>, node: &Node, keys: &[u8], values: &[u8]) -> sql::Result<Option<Vec<u8>>> {
    if !analyze_leaf(node, keys, values)?.corrupt {
        return Ok(None);
    }
    let bytes = reserialize_leaf(node, keys, values)?;
    let rewritten = analyze_leaf(&Node::decode(bytes.clone())?, keys, values)?;
    if rewritten.corrupt || rewritten.stats.unexpected_offsets != 0 {
        return Err(PgError::internal("rewritten leaf node still has incorrect address offsets"));
    }
    Ok(Some(bytes))
}

/// reserialize_leaf serializes a leaf's tuples again with the encodings, checking that the new leaf holds the same
/// tuples.
pub fn reserialize_leaf(node: &Node, keys: &[u8], values: &[u8]) -> sql::Result<Vec<u8>> {
    let key_items = (0..node.count()).map(|i| node.key(i)).collect::<Result<Vec<_>, _>>()?;
    let value_items = (0..node.count()).map(|i| node.value(i)).collect::<Result<Vec<_>, _>>()?;
    let serializer = ProllyMapSerializer {
        key_addresses: |tuple: &[u8]| address_offsets(tuple, keys),
        value_addresses: |tuple: &[u8]| address_offsets(tuple, values),
    };
    let bytes = serializer.serialize(&key_items, &value_items, &[], 0);
    let rewritten = Node::decode(bytes.clone())?;
    if rewritten.key_items() != node.key_items() || rewritten.value_items() != node.value_items() {
        return Err(PgError::internal("rewritten leaf node has different tuple bytes than the original"));
    }
    if rewritten.tree_count() != node.tree_count() {
        return Err(PgError::internal(format!(
            "rewritten leaf node has tree count {}, expected {}",
            rewritten.tree_count(),
            node.tree_count()
        )));
    }
    Ok(bytes)
}

/// repair_internal serializes an internal node again when its children changed or when it fails to record an
/// address that its keys hold.
pub fn repair_internal(
    _: &mut TreeRewriter<'_>,
    node: &Node,
    children: &[Hash],
    changed: bool,
    keys: &[u8],
    values: &[u8],
) -> sql::Result<Option<Vec<u8>>> {
    if !changed && !analyze_internal_keys(node, keys)?.corrupt {
        return Ok(None);
    }
    let bytes = reserialize_internal(node, children, keys, values)?;
    let rewritten = analyze_internal_keys(&Node::decode(bytes.clone())?, keys)?;
    if rewritten.corrupt || rewritten.unexpected_offsets != 0 {
        return Err(PgError::internal("rewritten internal node still has incorrect key address offsets"));
    }
    Ok(Some(bytes))
}

/// reserialize_internal serializes an internal node's keys again with the encodings and the new children, checking
/// that only the children and the recorded addresses changed.
pub fn reserialize_internal(node: &Node, children: &[Hash], keys: &[u8], values: &[u8]) -> sql::Result<Vec<u8>> {
    if node.count() != children.len() {
        return Err(PgError::internal(format!(
            "internal node has {} keys but {} children",
            node.count(),
            children.len()
        )));
    }
    let key_items = (0..node.count()).map(|i| node.key(i)).collect::<Result<Vec<_>, _>>()?;
    let child_items: Vec<&[u8]> = children.iter().map(|c| &c.0[..]).collect();
    let serializer = ProllyMapSerializer {
        key_addresses: |tuple: &[u8]| address_offsets(tuple, keys),
        value_addresses: |tuple: &[u8]| address_offsets(tuple, values),
    };
    let bytes = serializer.serialize(&key_items, &child_items, &node.subtree_counts()?, node.level());
    let rewritten = Node::decode(bytes.clone())?;
    if rewritten.key_items() != node.key_items() {
        return Err(PgError::internal("rewritten internal node has different key bytes than the original"));
    }
    if rewritten.level() != node.level() || rewritten.tree_count() != node.tree_count() {
        return Err(PgError::internal("rewritten internal node has different level or tree count than the original"));
    }
    if rewritten.count() != children.len() {
        return Err(PgError::internal("rewritten internal node has wrong child count"));
    }
    for (i, child) in children.iter().enumerate() {
        if rewritten.child(i)? != *child {
            return Err(PgError::internal("rewritten internal node has wrong child addresses"));
        }
    }
    Ok(bytes)
}
