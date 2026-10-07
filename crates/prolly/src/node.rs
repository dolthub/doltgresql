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

use std::ops::Range;

use serial::{Message, TreeNode};
use store::{ChunkReader, Error, Hash, Result};

/// COMMIT_CLOSURE_KEY_LEN is the width of a commit closure key: a u64 height and a commit address.
const COMMIT_CLOSURE_KEY_LEN: usize = 8 + Hash::LEN;

/// Values is where a node's values are.
#[derive(Clone, Debug)]
enum Values {
    /// Variable-width value items with their offsets.
    Items { items: Range<usize>, offsets: Range<usize> },
    /// Fixed-width child or value addresses.
    Addresses(Range<usize>),
    /// No values.
    None,
}

/// Node is a decoded tree node that owns its message bytes.
#[derive(Clone, Debug)]
pub struct Node {
    bytes: Vec<u8>,
    file_id: String,
    key_items: Range<usize>,
    /// The position of the key offsets, or None for fixed-width keys.
    key_offsets: Option<usize>,
    /// The width in bytes of each item offset.
    offset_width: usize,
    count: usize,
    values: Values,
    level: u8,
    tree_count: u64,
    /// The leaf item count under each child of an internal node.
    subtree_counts: Vec<u64>,
}

/// range_of returns the range of a slice within the bytes it was taken from.
fn range_of(bytes: &[u8], slice: &[u8]) -> Range<usize> {
    let start = slice.as_ptr() as usize - bytes.as_ptr() as usize;
    start..start + slice.len()
}

/// corrupt returns a Corrupt error.
fn corrupt(message: impl Into<String>) -> Error {
    Error::Corrupt(message.into())
}

impl Node {
    /// decode decodes a tree node message.
    pub fn decode(bytes: Vec<u8>) -> Result<Node> {
        let message = Message(&bytes);
        let file_id = message.file_id().to_string();
        let tree = TreeNode::new(message)?;
        let key_items = range_of(&bytes, tree.key_items);
        let (key_offsets, count) = match tree.key_offsets {
            Some(offsets) => (Some(offsets.start()), offsets.len().saturating_sub(1)),
            None if tree.key_items.len() % COMMIT_CLOSURE_KEY_LEN == 0 => {
                (None, tree.key_items.len() / COMMIT_CLOSURE_KEY_LEN)
            }
            None => return Err(corrupt("commit closure keys are not a whole number of items")),
        };
        let values = match (tree.value_items, tree.value_offsets, tree.address_array) {
            (Some(items), Some(offsets), _) => Values::Items {
                items: range_of(&bytes, items),
                offsets: offsets.start()..offsets.start() + offsets.len() * tree.offset_width,
            },
            (_, _, Some(addresses)) => Values::Addresses(range_of(&bytes, addresses)),
            _ => Values::None,
        };
        let subtree_counts = match tree.subtree_counts {
            Some(counts) => decode_counts(counts, count)?,
            None => Vec::new(),
        };
        let (level, tree_count, offset_width) = (tree.tree_level, tree.tree_count, tree.offset_width);
        Ok(Node {
            bytes,
            file_id,
            key_items,
            key_offsets,
            offset_width,
            count,
            values,
            level,
            tree_count,
            subtree_counts,
        })
    }

    /// load reads and decodes the node at the address.
    pub fn load(reader: &dyn ChunkReader, hash: &Hash) -> Result<Node> {
        Node::decode(reader.require(hash)?.data)
    }

    /// bytes returns the node's message bytes.
    pub fn bytes(&self) -> &[u8] {
        &self.bytes
    }

    /// file_id returns the node's message type.
    pub fn file_id(&self) -> &str {
        &self.file_id
    }

    /// count returns the number of items.
    pub fn count(&self) -> usize {
        self.count
    }

    /// level returns the node's height above the leaves, which is 0 for a leaf.
    pub fn level(&self) -> u8 {
        self.level
    }

    /// is_leaf reports whether the node is a leaf.
    pub fn is_leaf(&self) -> bool {
        self.level == 0
    }

    /// tree_count returns the number of leaf items under the node.
    pub fn tree_count(&self) -> u64 {
        self.tree_count
    }

    /// offset reads the offset at the index of the offsets starting at the position.
    fn offset(&self, start: usize, index: usize) -> Result<usize> {
        match self.offset_width {
            4 => serial::fb::u32_at(&self.bytes, start + index * 4).map(|offset| offset as usize),
            _ => serial::fb::u16_at(&self.bytes, start + index * 2).map(|offset| offset as usize),
        }
    }

    /// item returns the item at the index of items with offsets.
    fn item(&self, items: &Range<usize>, offsets: usize, index: usize) -> Result<&[u8]> {
        let (start, end) = (self.offset(offsets, index)?, self.offset(offsets, index + 1)?);
        if start > end || items.start + end > items.end {
            return Err(corrupt("tree node item offsets out of range"));
        }
        Ok(&self.bytes[items.start + start..items.start + end])
    }

    /// key returns the key at the index.
    pub fn key(&self, index: usize) -> Result<&[u8]> {
        match self.key_offsets {
            Some(offsets) => self.item(&self.key_items, offsets, index),
            None => {
                let start = self.key_items.start + index * COMMIT_CLOSURE_KEY_LEN;
                Ok(&self.bytes[start..start + COMMIT_CLOSURE_KEY_LEN])
            }
        }
    }

    /// value returns the value at the index: a value item, an address in an address array, or nothing for a node
    /// without values, such as a commit closure leaf.
    pub fn value(&self, index: usize) -> Result<&[u8]> {
        match &self.values {
            Values::Items { items, offsets } => self.item(items, offsets.start, index),
            Values::Addresses(addresses) => {
                let start = addresses.start + index * Hash::LEN;
                if start + Hash::LEN > addresses.end {
                    return Err(corrupt("tree node address out of range"));
                }
                Ok(&self.bytes[start..start + Hash::LEN])
            }
            Values::None => Ok(&[]),
        }
    }

    /// child returns the address of the child at the index of an internal node.
    pub fn child(&self, index: usize) -> Result<Hash> {
        serial::hash(self.value(index)?)
    }

    /// subtree_counts returns the leaf item count under each child of an internal node.
    pub fn subtree_counts(&self) -> Result<Vec<u64>> {
        Ok(self.subtree_counts.clone())
    }

    /// subtree_count returns the leaf item count under the child at the index of an internal node.
    pub fn subtree_count(&self, index: usize) -> Result<u64> {
        self.subtree_counts.get(index).copied().ok_or_else(|| corrupt("subtree count out of range"))
    }
}

/// decode_counts decodes the zigzag varint deltas of a node's subtree counts.
fn decode_counts(mut bytes: &[u8], count: usize) -> Result<Vec<u64>> {
    let mut counts = Vec::with_capacity(count);
    let mut previous: i64 = 0;
    for _ in 0..count {
        let (delta, len) = varint(bytes)?;
        bytes = &bytes[len..];
        previous += delta;
        counts.push(previous as u64);
    }
    if !bytes.is_empty() {
        return Err(corrupt("extra bytes after decoding varints"));
    }
    Ok(counts)
}

/// varint decodes a Go zigzag varint, returning it and its length.
fn varint(bytes: &[u8]) -> Result<(i64, usize)> {
    let mut value: u64 = 0;
    for (i, &byte) in bytes.iter().enumerate().take(10) {
        value |= ((byte & 0x7f) as u64) << (7 * i);
        if byte < 0x80 {
            let signed = (value >> 1) as i64 ^ -((value & 1) as i64);
            return Ok((signed, i + 1));
        }
    }
    Err(corrupt("invalid varint"))
}

/// ItemVisitor is called with each key and value.
pub type ItemVisitor<'f> = dyn FnMut(&[u8], &[u8]) -> Result<()> + 'f;

/// walk_leaves calls the function with every leaf item under the node, in key order.
pub fn walk_leaves(reader: &dyn ChunkReader, node: &Node, f: &mut ItemVisitor<'_>) -> Result<()> {
    if node.is_leaf() {
        for i in 0..node.count() {
            f(node.key(i)?, node.value(i)?)?;
        }
        return Ok(());
    }
    for i in 0..node.count() {
        walk_leaves(reader, &Node::load(reader, &node.child(i)?)?, f)?;
    }
    Ok(())
}
