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
