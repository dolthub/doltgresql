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

use std::path::{Path, PathBuf};

use crate::archive::ArchiveReader;
use crate::chunk::Chunk;
use crate::error::Result;
use crate::hash::Hash;
use crate::journal::{JOURNAL_FILE, Journal};
use crate::manifest::Manifest;
use crate::table::TableReader;

/// ARCHIVE_SUFFIX ends the file name of an archive.
pub(crate) const ARCHIVE_SUFFIX: &str = ".darc";
/// OLDGEN_DIR is the noms subdirectory holding the old generation.
const OLDGEN_DIR: &str = "oldgen";

/// Source is a file holding chunks.
pub(crate) enum Source {
    Table(TableReader),
    Archive(ArchiveReader),
    Journal(Journal),
}

impl Source {
    /// open_file opens the table file or archive with the name in a noms directory.
    pub(crate) fn open_file(dir: &Path, name: &Hash) -> Result<Source> {
        let archive = dir.join(format!("{name}{ARCHIVE_SUFFIX}"));
        if archive.exists() {
            Ok(Source::Archive(ArchiveReader::open(&archive)?))
        } else {
            Ok(Source::Table(TableReader::open(&dir.join(name.to_string()))?))
        }
    }

    pub(crate) fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        match self {
            Source::Table(table) => table.get(hash),
            Source::Archive(archive) => archive.get(hash),
            Source::Journal(journal) => journal.get(hash),
        }
    }

    pub(crate) fn locate(&self, hash: &Hash) -> Option<crate::Location> {
        match self {
            Source::Table(table) => table.locate(hash),
            Source::Archive(archive) => archive.locate(hash),
            Source::Journal(journal) => journal.locate(hash),
        }
    }

    pub(crate) fn has(&self, hash: &Hash) -> bool {
        match self {
            Source::Table(table) => table.has(hash),
            Source::Archive(archive) => archive.has(hash),
            Source::Journal(journal) => journal.has(hash),
        }
    }

    fn for_each(&self, f: &mut dyn FnMut(Chunk) -> Result<()>) -> Result<()> {
        match self {
            Source::Table(table) => table.for_each(f),
            Source::Archive(archive) => archive.for_each(f),
            Source::Journal(journal) => journal.for_each(f),
        }
    }
}

/// BlockStore reads the chunks of one noms directory: the files its manifest names and the chunk journal.
pub struct BlockStore {
    dir: PathBuf,
    manifest: Option<Manifest>,
    sources: Vec<Source>,
    root: Hash,
}

impl BlockStore {
    /// open opens the store in a noms directory without modifying it, taking the root from the journal's last root
    /// hash when it has one and from the manifest otherwise.
    pub fn open(dir: &Path) -> Result<BlockStore> {
        let manifest = Manifest::read(dir)?;
        let mut sources = Vec::new();
        let mut journal_root = Hash::default();
        let mut has_journal = false;
        for spec in manifest.iter().flat_map(|m| m.specs.iter()) {
            let name = spec.name.to_string();
            if name == JOURNAL_FILE {
                let journal = Journal::open(&dir.join(JOURNAL_FILE))?;
                journal_root = journal.root;
                has_journal = true;
                sources.push(Source::Journal(journal));
            } else {
                sources.push(Source::open_file(dir, &spec.name)?);
            }
        }
        if !has_journal && Journal::path(dir).exists() {
            let journal = Journal::open(&Journal::path(dir))?;
            journal_root = journal.root;
            sources.push(Source::Journal(journal));
        }
        let root =
            if journal_root.is_empty() { manifest.as_ref().map(|m| m.root).unwrap_or_default() } else { journal_root };
        Ok(BlockStore { dir: dir.to_path_buf(), manifest, sources, root })
    }

    /// dir returns the store's noms directory.
    pub fn dir(&self) -> &Path {
        &self.dir
    }

    /// manifest returns the store's manifest, which a new store lacks.
    pub fn manifest(&self) -> Option<&Manifest> {
        self.manifest.as_ref()
    }

    /// root returns the store's root hash.
    pub fn root(&self) -> Hash {
        self.root
    }

    /// get returns the chunk when the store holds it.
    pub fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        for source in &self.sources {
            if let Some(chunk) = source.get(hash)? {
                return Ok(Some(chunk));
            }
        }
        Ok(None)
    }

    /// locate returns where the chunk is in the store's files, when the store holds it.
    pub fn locate(&self, hash: &Hash) -> Option<crate::Location> {
        self.sources.iter().find_map(|source| source.locate(hash))
    }

    /// has reports whether the store holds the chunk.
    pub fn has(&self, hash: &Hash) -> bool {
        self.sources.iter().any(|source| source.has(hash))
    }

    /// for_each calls the function with every chunk of every file, so a chunk held by several files is visited once
    /// for each.
    pub fn for_each(&self, f: &mut dyn FnMut(Chunk) -> Result<()>) -> Result<()> {
        for source in &self.sources {
            source.for_each(f)?;
        }
        Ok(())
    }

    /// for_each_record calls the function with the address and compressed record of every chunk in a table file or
    /// the journal, which store chunks as records.
    pub fn for_each_record(&self, f: &mut dyn FnMut(Hash, &[u8]) -> Result<()>) -> Result<()> {
        for source in &self.sources {
            match source {
                Source::Table(table) => table.for_each_record(f)?,
                Source::Journal(journal) => journal.for_each_record(f)?,
                Source::Archive(_) => {}
            }
        }
        Ok(())
    }
}

/// GenerationalStore reads a database's chunks from its new generation, which holds the journal and recent files, and
/// its old generation, which GC moves committed chunks into.
pub struct GenerationalStore {
    pub new_gen: BlockStore,
    pub old_gen: BlockStore,
}

impl GenerationalStore {
    /// open opens the generations in a database's noms directory.
    pub fn open(dir: &Path) -> Result<GenerationalStore> {
        Ok(GenerationalStore { new_gen: BlockStore::open(dir)?, old_gen: BlockStore::open(&dir.join(OLDGEN_DIR))? })
    }

    /// root returns the database's root hash.
    pub fn root(&self) -> Hash {
        self.new_gen.root()
    }

    /// get returns the chunk when either generation holds it.
    pub fn get(&self, hash: &Hash) -> Result<Option<Chunk>> {
        match self.old_gen.get(hash)? {
            Some(chunk) => Ok(Some(chunk)),
            None => self.new_gen.get(hash),
        }
    }

    /// has reports whether either generation holds the chunk.
    pub fn has(&self, hash: &Hash) -> bool {
        self.old_gen.has(hash) || self.new_gen.has(hash)
    }
}
