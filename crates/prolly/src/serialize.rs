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

//! Serializers of tree nodes that write the same bytes as Dolt's, field for field in the same order.

use serial::Builder;

/// ITEM_TYPE_TUPLE_FORMAT_ALPHA is the item type of tuples.
const ITEM_TYPE_TUPLE_FORMAT_ALPHA: u8 = 1;

/// write_item_bytes writes the items concatenated as a byte vector.
fn write_item_bytes(b: &mut Builder, items: &[&[u8]]) -> u32 {
    let total: usize = items.iter().map(|item| item.len()).sum();
    b.prep(4, total);
    b.create_byte_vector(&items.concat())
}

/// write_item_offsets writes the u16 offsets of the items, starting at 0 and ending at their total length, into a
/// vector the caller started.
fn write_item_offsets(b: &mut Builder, items: &[&[u8]]) -> u32 {
    let mut offset: usize = items.iter().map(|item| item.len()).sum();
    for item in items.iter().rev() {
        b.prepend_u16(offset as u16);
        offset -= item.len();
    }
    b.prepend_u16(offset as u16);
    b.end_vector(items.len() + 1)
}

/// write_u16_vector writes a vector of u16 values in order.
fn write_u16_vector(b: &mut Builder, values: &[u16]) -> u32 {
    b.start_vector(2, values.len(), 2);
    for &value in values.iter().rev() {
        b.prepend_u16(value);
    }
    b.end_vector(values.len())
}

/// encode_counts encodes subtree counts as Go zigzag varint deltas.
fn encode_counts(counts: &[u64]) -> Vec<u8> {
    let mut out = Vec::new();
    let mut previous: i64 = 0;
    for &count in counts {
        let delta = (count as i64).wrapping_sub(previous);
        previous = count as i64;
        let mut zigzag = ((delta << 1) ^ (delta >> 63)) as u64;
        while zigzag >= 0x80 {
            out.push(zigzag as u8 | 0x80);
            zigzag >>= 7;
        }
        out.push(zigzag as u8);
    }
    out
}

/// ProllyNode is the contents of a ProllyTreeNode message.
pub struct ProllyNode<'a> {
    pub keys: Vec<&'a [u8]>,
    /// The value tuples of a leaf, or the child addresses of an internal node.
    pub values: Vec<&'a [u8]>,
    /// The leaf item count under each child of an internal node.
    pub subtrees: Vec<u64>,
    pub level: u8,
    /// The positions of addresses within the key items, in the order Dolt writes them.
    pub key_address_offsets: Vec<u16>,
    /// The positions of addresses within the value items of a leaf, in the order Dolt writes them.
    pub value_address_offsets: Vec<u16>,
}

/// serialize_prolly_node serializes a ProllyTreeNode as Dolt's ProllyMapSerializer does.
pub fn serialize_prolly_node(node: &ProllyNode<'_>) -> Vec<u8> {
    let mut b = Builder::new(0);
    let key_items = write_item_bytes(&mut b, &node.keys);
    b.start_vector(2, node.keys.len() + 1, 2);
    let key_offsets = write_item_offsets(&mut b, &node.keys);
    let leaf = node.level == 0;
    let (mut value_items, mut value_offsets, mut value_addresses, mut address_array, mut counts) = (0, 0, 0, 0, 0);
    let key_addresses;
    if leaf {
        value_items = write_item_bytes(&mut b, &node.values);
        b.start_vector(2, node.values.len() + 1, 2);
        value_offsets = write_item_offsets(&mut b, &node.values);
        key_addresses =
            if node.key_address_offsets.is_empty() { 0 } else { write_u16_vector(&mut b, &node.key_address_offsets) };
        if !node.value_address_offsets.is_empty() {
            value_addresses = write_u16_vector(&mut b, &node.value_address_offsets);
        }
    } else {
        address_array = write_item_bytes(&mut b, &node.values);
        counts = b.create_byte_vector(&encode_counts(&node.subtrees));
        key_addresses =
            if node.key_address_offsets.is_empty() { 0 } else { write_u16_vector(&mut b, &node.key_address_offsets) };
    }
    b.start_object(12);
    b.add_offset(0, key_items);
    b.add_offset(1, key_offsets);
    if leaf {
        b.add_offset(3, value_items);
        b.add_offset(4, value_offsets);
        b.add_u64(9, node.keys.len() as u64, 0);
        b.add_offset(6, value_addresses);
    } else {
        b.add_offset(7, address_array);
        b.add_offset(8, counts);
        b.add_u64(9, node.subtrees.iter().sum(), 0);
    }
    b.add_offset(11, key_addresses);
    b.add_u8(2, ITEM_TYPE_TUPLE_FORMAT_ALPHA, 0);
    b.add_u8(5, ITEM_TYPE_TUPLE_FORMAT_ALPHA, 0);
    b.add_u8(10, node.level, 0);
    let root = b.end_object();
    b.finish_message(root, serial::PROLLY_TREE_NODE)
}

/// serialize_address_map serializes an AddressMap node as Dolt's AddressMapSerializer does.
pub fn serialize_address_map(keys: &[&[u8]], addresses: &[&[u8]], subtrees: &[u64], level: u8) -> Vec<u8> {
    let mut b = Builder::new(0);
    let key_items = write_item_bytes(&mut b, keys);
    b.start_vector(2, keys.len() + 1, 2);
    let key_offsets = write_item_offsets(&mut b, keys);
    let address_array = write_item_bytes(&mut b, addresses);
    let counts = if level > 0 { b.create_byte_vector(&encode_counts(subtrees)) } else { 0 };
    b.start_object(6);
    b.add_offset(0, key_items);
    b.add_offset(1, key_offsets);
    b.add_offset(2, address_array);
    if level > 0 {
        b.add_offset(3, counts);
        b.add_u64(4, subtrees.iter().sum(), 0);
    } else {
        b.add_u64(4, keys.len() as u64, 0);
    }
    b.add_u8(5, level, 0);
    let root = b.end_object();
    b.finish_message(root, serial::ADDRESS_MAP)
}

/// serialize_blob serializes a Blob node as Dolt's BlobSerializer does: a payload at a leaf, and child addresses
/// with their sizes above.
pub fn serialize_blob(values: &[&[u8]], subtrees: &[u64], level: u8) -> Vec<u8> {
    let mut b = Builder::new(0);
    if level == 0 {
        let payload = b.create_byte_vector(values[0]);
        b.start_object(5);
        b.add_offset(0, payload);
    } else {
        let addresses = write_item_bytes(&mut b, values);
        let sizes = b.create_byte_vector(&encode_counts(subtrees));
        b.start_object(5);
        b.add_offset(1, addresses);
        b.add_offset(2, sizes);
    }
    b.add_u64(3, subtrees.iter().sum(), 0);
    b.add_u8(4, level, 0);
    let root = b.end_object();
    b.finish_message(root, serial::BLOB)
}

/// serialize_commit_closure serializes a CommitClosure node as Dolt's CommitClosureSerializer does.
pub fn serialize_commit_closure(keys: &[&[u8]], addresses: &[&[u8]], subtrees: &[u64], level: u8) -> Vec<u8> {
    let mut b = Builder::new(0);
    let key_items = write_item_bytes(&mut b, keys);
    let (mut address_array, mut counts) = (0, 0);
    if level > 0 {
        address_array = write_item_bytes(&mut b, addresses);
        counts = b.create_byte_vector(&encode_counts(subtrees));
    }
    b.start_object(5);
    b.add_offset(0, key_items);
    if level > 0 {
        b.add_offset(1, address_array);
        b.add_offset(2, counts);
        b.add_u64(3, subtrees.iter().sum(), 0);
    } else {
        b.add_u64(3, keys.len() as u64, 0);
    }
    b.add_u8(4, level, 0);
    let root = b.end_object();
    b.finish_message(root, serial::COMMIT_CLOSURE)
}
