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

use crate::archive::ArchiveWriter;
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

/// ABORT_ENV makes garbage collection fail after it writes each incremental file, as Dolt's tests use to check that
/// a collection resumes.
const ABORT_ENV: &str = "DOLT_TEST_ABORT_GC_AFTER_INCREMENTAL_FILE_WRITE";

/// FileWriter builds a table file or an archive.
enum FileWriter {
    Table(TableWriter),
    Archive(ArchiveWriter),
}

impl FileWriter {
    /// new returns a writer of archives, or of table files at archive level 0.
    fn new(archive: bool) -> FileWriter {
        match archive {
            true => FileWriter::Archive(ArchiveWriter::new()),
            false => FileWriter::Table(TableWriter::new()),
        }
    }

    /// add adds a chunk.
    fn add(&mut self, chunk: Chunk) -> Result<()> {
        match self {
            FileWriter::Table(writer) => writer.add_chunk(&chunk),
            FileWriter::Archive(writer) => writer.add_chunk(chunk)?,
        }
        Ok(())
    }

    /// count returns the number of chunks added.
    fn count(&self) -> usize {
        match self {
            FileWriter::Table(writer) => writer.count(),
            FileWriter::Archive(writer) => writer.count(),
        }
    }

    /// write writes the file to a directory and returns its spec, or None without chunks.
    fn write(self, dir: &Path) -> Result<Option<TableSpec>> {
        let chunk_count = self.count() as u32;
        if chunk_count == 0 {
            return Ok(None);
        }
        std::fs::create_dir_all(dir)?;
        let (name, path, bytes) = match self {
            FileWriter::Table(writer) => {
                let (name, bytes) = writer.finish();
                (name, dir.join(name.to_string()), bytes)
            }
            FileWriter::Archive(writer) => {
                let (name, bytes) = writer.finish();
                (name, dir.join(format!("{name}.darc")), bytes)
            }
        };
        std::fs::write(&path, bytes)?;
        File::open(&path)?.sync_all()?;
        Ok(Some(TableSpec { name, chunk_count }))
    }
}

/// write_files writes the chunks a collection keeps to new files in a directory, as Dolt's GC copiers do: archives,
/// or table files at archive level 0, and with an incremental file size, the leaf chunks in files of about that many
/// compressed bytes, each passed to `written` once it is on disk, and the other chunks in one more file. Each chunk
/// comes with whether it is a leaf.
pub fn write_files(
    dir: &Path,
    chunks: Vec<(Chunk, bool)>,
    archive: bool,
    incremental_file_size: u64,
    written: &mut dyn FnMut(&TableSpec) -> Result<()>,
) -> Result<Vec<TableSpec>> {
    let mut specs = Vec::new();
    let (mut leaves, mut others) = (FileWriter::new(archive), FileWriter::new(archive));
    let mut leaf_bytes = 0;
    for (chunk, leaf) in chunks {
        if !leaf || incremental_file_size == 0 {
            others.add(chunk)?;
            continue;
        }
        leaf_bytes += chunk.to_record().len() as u64;
        leaves.add(chunk)?;
        if leaf_bytes >= incremental_file_size {
            let full = std::mem::replace(&mut leaves, FileWriter::new(archive));
            specs.extend(finish_incremental(dir, full, written)?);
            leaf_bytes = 0;
        }
    }
    if incremental_file_size != 0 {
        specs.extend(finish_incremental(dir, leaves, written)?);
    }
    specs.extend(others.write(dir)?);
    Ok(specs)
}

/// finish_incremental writes a file of leaf chunks and passes it on, failing afterwards when a test asks for that.
fn finish_incremental(
    dir: &Path,
    writer: FileWriter,
    written: &mut dyn FnMut(&TableSpec) -> Result<()>,
) -> Result<Option<TableSpec>> {
    let spec = writer.write(dir)?;
    if let Some(spec) = &spec {
        written(spec)?;
    }
    if std::env::var_os(ABORT_ENV).is_some() {
        return Err(std::io::Error::other("GC aborting after writing incremental table file").into());
    }
    Ok(spec)
}

/// add_to_manifest adds a file that it does not already name to a directory's manifest, writing a new manifest with
/// the root when there is none, as Dolt's addTableFilesToManifest does.
pub fn add_to_manifest(dir: &Path, root: Hash, format: &str, spec: &TableSpec) -> Result<()> {
    let mut specs = Manifest::read(dir)?.map(|m| m.specs).unwrap_or_default();
    if specs.iter().any(|s| s.name == spec.name) {
        return Ok(());
    }
    specs.push(*spec);
    let manifest = Manifest {
        version: "5".to_string(),
        format: format.to_string(),
        lock: lock_hash(&root, &specs, &[], b""),
        root,
        gc_gen: lock_hash(&root, &specs, &[], b"gc"),
        specs,
    };
    write_manifest(dir, &manifest)
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
