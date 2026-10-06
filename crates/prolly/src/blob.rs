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

use serial::{Blob, Message};
use store::{ChunkReader, Hash, Result};

use crate::serialize_blob;

/// NodeSink receives each node a blob builder writes, with its address.
pub type NodeSink<'a> = dyn FnMut(Hash, &[u8]) -> Result<()> + 'a;

/// read_blob returns the bytes of the blob tree at the address: the payloads of its leaves in order.
pub fn read_blob(reader: &dyn ChunkReader, hash: &Hash) -> Result<Vec<u8>> {
    let mut out = Vec::new();
    append_blob(reader, hash, &mut out)?;
    Ok(out)
}

/// append_blob appends the payloads of the blob tree's leaves.
fn append_blob(reader: &dyn ChunkReader, hash: &Hash, out: &mut Vec<u8>) -> Result<()> {
    let chunk = reader.require(hash)?;
    let blob = Blob::new(Message(&chunk.data))?;
    if blob.tree_level()? == 0 {
        out.extend_from_slice(blob.payload()?.unwrap_or_default());
        return Ok(());
    }
    for child in serial::hashes(blob.address_array()?.unwrap_or_default())? {
        append_blob(reader, &child, out)?;
    }
    Ok(())
}

/// BLOB_CHUNK_SIZE is the size of a blob leaf's payload, as Dolt's DefaultFixedChunkLength.
pub const BLOB_CHUNK_SIZE: usize = 4000;

/// BLOB_FANOUT is the number of children of an internal blob node.
const BLOB_FANOUT: usize = BLOB_CHUNK_SIZE / Hash::LEN;

/// write_blob writes the bytes as a blob tree of fixed-size leaves to the sink, as Dolt's BlobBuilder does, and
/// returns the root's address and bytes, or None for no bytes.
pub fn write_blob(data: &[u8], sink: &mut NodeSink<'_>) -> Result<Option<(Hash, Vec<u8>)>> {
    if data.is_empty() {
        return Ok(None);
    }
    let mut top_level = 0;
    if data.len() > BLOB_CHUNK_SIZE {
        let mut size = data.len() / BLOB_CHUNK_SIZE;
        while size > 0 {
            size /= BLOB_FANOUT;
            top_level += 1;
        }
    }
    let mut position = 0;
    let written = write_blob_level(data, &mut position, top_level, sink)?;
    Ok(written.node)
}

/// BlobLevelWrite is what writing one node at a level of a blob produced.
struct BlobLevelWrite {
    /// The node's address and bytes, which a leaf lacks when no bytes remained.
    node: Option<(Hash, Vec<u8>)>,
    /// The number of leaves under the node.
    leaves: u64,
    /// Whether the bytes ran out, as the reader's io.EOF does in Go.
    end: bool,
}

/// write_blob_level writes the next node of the level from the bytes at the position, where an internal node takes
/// up to BLOB_FANOUT children and is written even when it ends up empty, as Go's blobLevelWriter is.
fn write_blob_level(data: &[u8], position: &mut usize, level: u8, sink: &mut NodeSink<'_>) -> Result<BlobLevelWrite> {
    if level == 0 {
        if *position == data.len() {
            return Ok(BlobLevelWrite { node: None, leaves: 0, end: true });
        }
        let end = (*position + BLOB_CHUNK_SIZE).min(data.len());
        let bytes = serialize_blob(&[&data[*position..end]], &[1], 0);
        *position = end;
        let hash = Hash::of(&bytes);
        sink(hash, &bytes)?;
        return Ok(BlobLevelWrite { node: Some((hash, bytes)), leaves: 1, end: false });
    }
    let (mut addresses, mut counts) = (Vec::new(), Vec::new());
    loop {
        let child = write_blob_level(data, position, level - 1, sink)?;
        if child.leaves != 0 {
            addresses.push(child.node.unwrap().0);
            counts.push(child.leaves);
        }
        if addresses.len() >= BLOB_FANOUT || child.end {
            let values: Vec<&[u8]> = addresses.iter().map(|a| &a.0[..]).collect();
            let bytes = serialize_blob(&values, &counts, level);
            let hash = Hash::of(&bytes);
            sink(hash, &bytes)?;
            return Ok(BlobLevelWrite { node: Some((hash, bytes)), leaves: counts.iter().sum(), end: child.end });
        }
    }
}
