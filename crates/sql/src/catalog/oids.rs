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

//! The OIDs that clients see for Doltgres' internal object IDs, which last as long as the server runs, as Go's ID
//! cache hands them out.

use std::collections::HashMap;
use std::sync::{Mutex, OnceLock};

use crate::catalog::id;

/// BUILTIN_OID_LIMIT is the largest OID that Postgres gives built-in objects, which user objects never take.
const BUILTIN_OID_LIMIT: u32 = 65535;

/// Cache maps internal IDs to OIDs and back.
#[derive(Default)]
struct Cache {
    to_oid: HashMap<Vec<u8>, u32>,
    to_id: HashMap<u32, Vec<u8>>,
}

impl Cache {
    /// set records an ID's OID.
    fn set(&mut self, id: Vec<u8>, oid: u32) {
        self.to_id.insert(oid, id.clone());
        self.to_oid.insert(id, oid);
    }
}

/// cache returns the process's cache, with the built-in objects that have fixed OIDs.
fn cache() -> &'static Mutex<Cache> {
    static CACHE: OnceLock<Mutex<Cache>> = OnceLock::new();
    CACHE.get_or_init(|| {
        let mut cache = Cache::default();
        for (name, oid) in [("pg_catalog", 11), ("pg_toast", 99), ("public", 2200), ("information_schema", 13679)] {
            cache.set(id::new(id::SECTION_NAMESPACE, &[name]), oid);
        }
        for (name, oid) in [("template1", 1), ("template0", 4), ("postgres", 5)] {
            cache.set(id::new(id::SECTION_DATABASE, &[name]), oid);
        }
        for ty in crate::catalog::builtin_types() {
            cache.set(ty.definition.id.clone(), ty.oid);
        }
        Mutex::new(cache)
    })
}

/// oid returns the OID of an internal ID, the CRC32C of its bytes unless another ID or a built-in object has it.
pub fn oid(id: &[u8]) -> u32 {
    if id.is_empty() {
        return 0;
    }
    let mut cache = cache().lock().unwrap_or_else(|e| e.into_inner());
    if let Some(&oid) = cache.to_oid.get(id) {
        return oid;
    }
    let mut candidate = crc32c::crc32c(id);
    while candidate <= BUILTIN_OID_LIMIT || cache.to_id.contains_key(&candidate) {
        candidate = candidate.wrapping_sub(1);
    }
    cache.set(id.to_vec(), candidate);
    candidate
}

/// id returns the internal ID that an OID was handed out for.
pub fn id(oid: u32) -> Option<Vec<u8>> {
    cache().lock().unwrap_or_else(|e| e.into_inner()).to_id.get(&oid).cloned()
}
