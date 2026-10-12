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
    File::options().write(true).open(&path)?.sync_all()?;
    Ok(Some(TableSpec { name, chunk_count }))
}

/// ABORT_ENV makes garbage collection fail after it writes each incremental file, as Dolt's tests use to check that
/// a collection resumes.
const ABORT_ENV: &str = "DOLT_TEST_ABORT_GC_AFTER_INCREMENTAL_FILE_WRITE";

/// FileWriter builds a table file or an archive in a directory, moving its bytes to a spill file there as they
/// pile up, so that a collection never holds a whole file in memory.
struct FileWriter {
    writer: Writer,
    spill: crate::file::Spill,
}

/// Writer is the builder a file writer uses.
enum Writer {
    Table(TableWriter),
    Archive(ArchiveWriter),
}

impl FileWriter {
    /// new returns a writer of an archive in a directory, or of a table file at archive level 0.
    fn new(dir: &Path, archive: bool) -> Result<FileWriter> {
        let writer = match archive {
            true => Writer::Archive(ArchiveWriter::new()),
            false => Writer::Table(TableWriter::new()),
        };
        Ok(FileWriter { writer, spill: crate::file::Spill::create(dir)? })
    }

    /// add adds a chunk, which an archive takes over as another archive stores it when it comes with that form.
    fn add(&mut self, chunk: Chunk, stored: Option<crate::Stored>) -> Result<()> {
        let buffered = match &mut self.writer {
            Writer::Table(writer) => {
                writer.add_chunk(&chunk);
                writer.buffered()
            }
            Writer::Archive(writer) => {
                match stored {
                    Some(stored) => writer.add_stored(chunk.hash, stored),
                    None => writer.add_chunk(chunk)?,
                }
                writer.buffered()
            }
        };
        if buffered >= crate::file::SPILL_LEN {
            self.spill_buffer()?;
        }
        Ok(())
    }

    /// spill_buffer moves what the writer holds in memory to the spill file.
    fn spill_buffer(&mut self) -> Result<()> {
        match &mut self.writer {
            Writer::Table(writer) => writer.spill(&mut self.spill),
            Writer::Archive(writer) => writer.spill(&mut self.spill),
        }
    }

    /// count returns the number of chunks added.
    fn count(&self) -> usize {
        match &self.writer {
            Writer::Table(writer) => writer.count(),
            Writer::Archive(writer) => writer.count(),
        }
    }

    /// write finishes the file in its directory and returns its spec, or None without chunks.
    fn write(mut self, dir: &Path) -> Result<Option<TableSpec>> {
        let chunk_count = self.count() as u32;
        if chunk_count == 0 {
            return Ok(None);
        }
        self.spill_buffer()?;
        let FileWriter { writer, mut spill } = self;
        let name = match writer {
            Writer::Table(writer) => {
                let (name, mut tail) = writer.finish();
                spill.write(&mut tail)?;
                spill.finish(|_| dir.join(name.to_string()))?;
                name
            }
            Writer::Archive(writer) => {
                let (_, mut tail) = writer.finish();
                spill.write(&mut tail)?;
                let mut named = Hash::default();
                spill.finish(|hash| {
                    named = hash;
                    dir.join(format!("{hash}.darc"))
                })?;
                named
            }
        };
        Ok(Some(TableSpec { name, chunk_count }))
    }
}

/// write_files writes the chunks a collection keeps to new files in a directory, as `GcWriter` does. Each chunk comes
/// with whether it is a leaf.
pub fn write_files(
    dir: &Path,
    chunks: Vec<(Chunk, bool)>,
    archive: bool,
    incremental_file_size: u64,
    written: &mut dyn FnMut(&TableSpec) -> Result<()>,
) -> Result<Vec<TableSpec>> {
    let mut writer = GcWriter::new(dir, archive, incremental_file_size)?;
    for (chunk, leaf) in chunks {
        writer.add(chunk, None, leaf, written)?;
    }
    writer.finish(written)
}

/// GcWriter writes the chunks a collection keeps to new files in a directory as they come, as Dolt's GC copiers do:
/// archives, or table files at archive level 0, and with an incremental file size, the leaf chunks in files of about
/// that many compressed bytes, each passed to `written` once it is on disk, and the other chunks in one more file.
pub struct GcWriter {
    dir: std::path::PathBuf,
    archive: bool,
    incremental_file_size: u64,
    leaves: FileWriter,
    others: FileWriter,
    leaf_bytes: u64,
    specs: Vec<TableSpec>,
}

impl GcWriter {
    /// new returns a writer of a collection's files in a directory.
    pub fn new(dir: &Path, archive: bool, incremental_file_size: u64) -> Result<GcWriter> {
        Ok(GcWriter {
            dir: dir.to_path_buf(),
            archive,
            incremental_file_size,
            leaves: FileWriter::new(dir, archive)?,
            others: FileWriter::new(dir, archive)?,
            leaf_bytes: 0,
            specs: Vec::new(),
        })
    }

    /// add adds a chunk, with its stored form in an archive when it has one and whether it is a leaf, which refers to
    /// no other chunk.
    pub fn add(
        &mut self,
        chunk: Chunk,
        stored: Option<crate::Stored>,
        leaf: bool,
        written: &mut dyn FnMut(&TableSpec) -> Result<()>,
    ) -> Result<()> {
        if !leaf || self.incremental_file_size == 0 {
            return self.others.add(chunk, stored);
        }
        self.leaf_bytes += chunk.to_record().len() as u64;
        self.leaves.add(chunk, stored)?;
        if self.leaf_bytes >= self.incremental_file_size {
            let full = std::mem::replace(&mut self.leaves, FileWriter::new(&self.dir, self.archive)?);
            self.specs.extend(finish_incremental(&self.dir, full, written)?);
            self.leaf_bytes = 0;
        }
        Ok(())
    }

    /// finish writes the files still open and returns the specs of every file written.
    pub fn finish(mut self, written: &mut dyn FnMut(&TableSpec) -> Result<()>) -> Result<Vec<TableSpec>> {
        if self.incremental_file_size != 0 {
            self.specs.extend(finish_incremental(&self.dir, self.leaves, written)?);
        }
        self.specs.extend(self.others.write(&self.dir)?);
        Ok(self.specs)
    }
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
