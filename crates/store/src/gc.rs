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

//! Garbage collection's file work: writing the chunks a collection keeps to a new table file, and replacing the files
//! that a noms directory's manifest names.

use std::fs::File;
use std::path::Path;

use crate::chunk::Chunk;
use crate::error::Result;
use crate::hash::Hash;
use crate::journal::JOURNAL_FILE;
use crate::journal_store::write_manifest;
use crate::journal_writer::JOURNAL_INDEX_FILE;
use crate::manifest::{Manifest, TableSpec, lock_hash};
use crate::table::TableWriter;

/// write_table writes chunks to a new table file in a directory, returning the file's spec, or None without chunks.
pub fn write_table(dir: &Path, chunks: &[Chunk]) -> Result<Option<TableSpec>> {
    if chunks.is_empty() {
        return Ok(None);
    }
    std::fs::create_dir_all(dir)?;
    let mut writer = TableWriter::new();
    for chunk in chunks {
        writer.add_chunk(chunk);
    }
    let chunk_count = writer.count() as u32;
    let (name, bytes) = writer.finish();
    let path = dir.join(name.to_string());
    std::fs::write(&path, bytes)?;
    File::open(&path)?.sync_all()?;
    Ok(Some(TableSpec { name, chunk_count }))
}

/// replace_files writes a manifest for a directory that names only the files given, with a new GC generation, and
/// then deletes the table files, archives, and journal that it no longer names.
pub fn replace_files(dir: &Path, root: Hash, format: &str, specs: Vec<TableSpec>) -> Result<Manifest> {
    let manifest = Manifest {
        version: "5".to_string(),
        format: format.to_string(),
        lock: lock_hash(&root, &specs, &[], b""),
        root,
        gc_gen: lock_hash(&root, &specs, &[], b"gc"),
        specs,
    };
    write_manifest(dir, &manifest)?;
    for entry in std::fs::read_dir(dir)? {
        let entry = entry?;
        let name = entry.file_name().to_string_lossy().into_owned();
        let stale = match name.as_str() {
            JOURNAL_FILE | JOURNAL_INDEX_FILE => true,
            _ => Hash::parse(name.strip_suffix(".darc").unwrap_or(&name))
                .is_some_and(|hash| !manifest.specs.iter().any(|spec| spec.name == hash)),
        };
        if stale && entry.file_type()?.is_file() {
            std::fs::remove_file(entry.path())?;
        }
    }
    Ok(manifest)
}
