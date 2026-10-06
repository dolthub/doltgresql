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

use std::path::Path;

use crate::error::{Result, corrupt};
use crate::hash::Hash;

/// MANIFEST_FILE is the name of the manifest in a noms directory.
pub const MANIFEST_FILE: &str = "manifest";

/// TableSpec names a table file, archive, or the journal, with its chunk count.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct TableSpec {
    pub name: Hash,
    pub chunk_count: u32,
}

/// Manifest is the persisted state of a chunk store: its format, root, and files.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Manifest {
    /// The manifest version, "4" or "5".
    pub version: String,
    /// The noms binary format, such as "__DOLT__".
    pub format: String,
    /// The hash that writers check before replacing the manifest.
    pub lock: Hash,
    /// The root hash of the store.
    pub root: Hash,
    /// The GC generation, which version 4 manifests lack.
    pub gc_gen: Hash,
    /// The files holding the store's chunks.
    pub specs: Vec<TableSpec>,
}

impl Manifest {
    /// read reads the manifest in a noms directory, returning None when there is none.
    pub fn read(dir: &Path) -> Result<Option<Manifest>> {
        match std::fs::read(dir.join(MANIFEST_FILE)) {
            Ok(bytes) => Manifest::parse(&bytes).map(Some),
            Err(err) if err.kind() == std::io::ErrorKind::NotFound => Ok(None),
            Err(err) => Err(err.into()),
        }
    }

    /// parse parses a version 4 or 5 manifest, whose colon-separated fields are the version, format, lock, root,
    /// the GC generation in version 5, and then a name and chunk count for each file.
    pub fn parse(bytes: &[u8]) -> Result<Manifest> {
        let text = std::str::from_utf8(bytes).map_err(|_| corrupt("manifest is not UTF-8"))?;
        let Some((version, rest)) = text.split_once(':').filter(|(version, _)| version.len() < 8) else {
            return Err(corrupt("corrupt manifest"));
        };
        let fields: Vec<&str> = rest.split(':').collect();
        let fixed = match version {
            "4" => 3,
            "5" => 4,
            _ => {
                return Err(corrupt(format!("Unknown manifest version: {version}. You may need to update your client")));
            }
        };
        if fields.len() < fixed || !(fields.len() - fixed).is_multiple_of(2) {
            return Err(corrupt("corrupt manifest"));
        }
        let lock =
            Hash::parse(fields[1]).ok_or_else(|| corrupt(format!("Could not parse lock hash: {}", fields[1])))?;
        let root = Hash::parse(fields[2]).unwrap_or_default();
        let gc_gen = if version == "5" {
            Hash::parse(fields[3])
                .ok_or_else(|| corrupt(format!("Could not parse GC generation hash: {}", fields[3])))?
        } else {
            Hash::default()
        };
        let mut specs = Vec::new();
        for pair in fields[fixed..].chunks(2) {
            let name = Hash::parse(pair[0]).ok_or_else(|| corrupt(format!("invalid table file name: {}", pair[0])))?;
            let chunk_count = pair[1].parse().map_err(|_| corrupt(format!("invalid chunk count: {}", pair[1])))?;
            specs.push(TableSpec { name, chunk_count });
        }
        Ok(Manifest { version: version.to_string(), format: fields[0].to_string(), lock, root, gc_gen, specs })
    }
}
