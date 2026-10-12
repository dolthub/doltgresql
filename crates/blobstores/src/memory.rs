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
use std::sync::Mutex;

use store::{Blob, BlobRange, Blobstore, Error, MANIFEST_KEY, Result, not_found};

/// MemoryBlobstore keeps blobs in memory, each with a version that counts its writes, as Dolt's InMemoryBlobstore
/// does for mem:// URLs.
#[derive(Default)]
pub struct MemoryBlobstore {
    blobs: Mutex<HashMap<String, (Vec<u8>, u64)>>,
}

impl MemoryBlobstore {
    /// store stores a blob, returning its version.
    fn store(blobs: &mut HashMap<String, (Vec<u8>, u64)>, key: &str, data: Vec<u8>) -> String {
        let version = blobs.get(key).map_or(1, |(_, v)| v + 1);
        blobs.insert(key.to_string(), (data, version));
        version.to_string()
    }

    /// lock locks the blobs.
    fn lock(&self) -> std::sync::MutexGuard<'_, HashMap<String, (Vec<u8>, u64)>> {
        self.blobs.lock().unwrap_or_else(|poisoned| poisoned.into_inner())
    }
}

impl Blobstore for MemoryBlobstore {
    fn path(&self) -> String {
        String::new()
    }

    fn exists(&self, key: &str) -> Result<bool> {
        Ok(self.lock().contains_key(key))
    }

    fn get(&self, key: &str, range: BlobRange) -> Result<Blob> {
        let blobs = self.lock();
        let (data, version) = blobs.get(key).ok_or_else(|| not_found(key))?;
        let range = range.positive(data.len() as i64);
        let (start, end) = (range.offset.max(0) as usize, (range.offset + range.length).max(0) as usize);
        Ok(Blob {
            data: data[start..end.min(data.len())].to_vec(),
            size: data.len() as u64,
            version: version.to_string(),
        })
    }

    fn put(&self, key: &str, data: &[u8]) -> Result<String> {
        Ok(MemoryBlobstore::store(&mut self.lock(), key, data.to_vec()))
    }

    fn check_and_put_manifest(&self, expected: &str, data: &[u8]) -> Result<String> {
        let mut blobs = self.lock();
        let actual = blobs.get(MANIFEST_KEY).map(|(_, v)| v.to_string()).unwrap_or_default();
        if actual != expected {
            return Err(Error::VersionMismatch { key: MANIFEST_KEY.into(), expected: expected.into(), actual });
        }
        Ok(MemoryBlobstore::store(&mut blobs, MANIFEST_KEY, data.to_vec()))
    }

    fn concatenate(&self, key: &str, sources: &[String]) -> Result<String> {
        let mut blobs = self.lock();
        let mut data = Vec::new();
        for source in sources {
            data.extend_from_slice(&blobs.get(source).ok_or_else(|| not_found(source))?.0);
        }
        Ok(MemoryBlobstore::store(&mut blobs, key, data))
    }
}
