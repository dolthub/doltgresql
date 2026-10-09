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

//! The chunk stores of the remote URL schemes that keep a store's files as blobs, as Dolt's dbfactory opens them.

mod aws;
mod azure;
mod gcs;
mod git;
mod http;
mod memory;
mod oci;
mod oss;

use std::collections::BTreeMap;
use std::sync::Arc;

use store::{BlobChunkStore, Blobstore, ChunkStore, LocalBlobstore, Result};

pub use memory::MemoryBlobstore;

/// FORMAT is the storage format that new stores record.
const FORMAT: &str = "__DOLT__";

/// open returns the chunk store at a URL whose scheme keeps blobs, given the remote's parameters, or None for a scheme
/// that does not.
pub fn open(url: &str, params: &BTreeMap<String, String>) -> Result<Option<Box<dyn ChunkStore>>> {
    let Some((scheme, rest)) = url.split_once("://") else { return Ok(None) };
    let (blobs, concatenates): (Arc<dyn Blobstore>, bool) = match scheme.to_ascii_lowercase().as_str() {
        "localbs" => (Arc::new(LocalBlobstore::new(std::path::absolute(rest)?)), true),
        "mem" => (Arc::new(MemoryBlobstore::default()), true),
        "s3" => (Arc::new(aws::open_s3(rest)?), false),
        "aws" => (Arc::new(aws::open_aws(rest, params)?), false),
        "gs" => (Arc::new(gcs::open_gcs(rest)?), true),
        "az" => (Arc::new(azure::open_azure(rest)?), true),
        "oci" => (Arc::new(oci::open_oci(rest)?), false),
        "oss" => (Arc::new(oss::open_oss(rest, params)?), false),
        "git+file" | "git+http" | "git+https" | "git+ssh" => (Arc::new(git::open(url, params)?), false),
        _ => return Ok(None),
    };
    Ok(Some(Box::new(BlobChunkStore::open(blobs, FORMAT, concatenates)?)))
}

/// percent_decode decodes the percent escapes of a URL's query value.
pub(crate) fn percent_decode(text: &str) -> String {
    let bytes = text.as_bytes();
    let mut out = Vec::with_capacity(bytes.len());
    let mut i = 0;
    while i < bytes.len() {
        match (
            bytes[i],
            bytes.get(i + 1..i + 3).and_then(|h| u8::from_str_radix(std::str::from_utf8(h).ok()?, 16).ok()),
        ) {
            (b'%', Some(byte)) => {
                out.push(byte);
                i += 3;
            }
            (b'+', _) => {
                out.push(b' ');
                i += 1;
            }
            (byte, _) => {
                out.push(byte);
                i += 1;
            }
        }
    }
    String::from_utf8_lossy(&out).into_owned()
}
