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

#![forbid(unsafe_code)]

//! Dolt's prolly trees, which store sorted key-value items in content-addressed nodes. A leaf node holds items, and
//! an internal node holds the last key of each child with the child's address and item count.

mod blob;
mod chunker;
mod cursor;
mod node;
mod serialize;
mod tuple;

pub use blob::{BLOB_CHUNK_SIZE, NodeSink, read_blob, write_blob};
pub use chunker::{Chunker, NodeSerializer, apply_mutations};
pub use cursor::{Compare, NodeStore};
pub use node::{ItemVisitor, Node, walk_leaves};
pub use serialize::{
    AddressMapSerializer, CommitClosureSerializer, MergeArtifactsSerializer, ProllyMapSerializer, ProllyNode,
    serialize_address_map, serialize_blob, serialize_commit_closure, serialize_merge_artifacts, serialize_prolly_node,
};
pub use tuple::Tuple;
