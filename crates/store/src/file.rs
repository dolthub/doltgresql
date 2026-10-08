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

use std::fs::File;

use crate::error::Result;

/// ReadAt reads bytes at offsets of a file or of a blob that a remote store holds.
pub trait ReadAt: Send + Sync {
    /// read_at reads exactly `len` bytes at the offset.
    fn read_at(&self, offset: u64, len: usize) -> Result<Vec<u8>>;

    /// size returns the number of bytes.
    fn size(&self) -> Result<u64>;
}

impl ReadAt for File {
    fn read_at(&self, offset: u64, len: usize) -> Result<Vec<u8>> {
        read_at(self, offset, len)
    }

    fn size(&self) -> Result<u64> {
        Ok(self.metadata()?.len())
    }
}

/// read_at reads exactly `len` bytes at the offset without moving the file's cursor, so readers can share the file.
pub fn read_at(file: &File, offset: u64, len: usize) -> Result<Vec<u8>> {
    let mut buffer = vec![0; len];
    #[cfg(unix)]
    {
        use std::os::unix::fs::FileExt;
        file.read_exact_at(&mut buffer, offset)?;
    }
    #[cfg(windows)]
    {
        use std::os::windows::fs::FileExt;
        let mut done = 0;
        while done < len {
            let read = file.seek_read(&mut buffer[done..], offset + done as u64)?;
            if read == 0 {
                return Err(std::io::Error::from(std::io::ErrorKind::UnexpectedEof).into());
            }
            done += read;
        }
    }
    Ok(buffer)
}

/// write_at writes all of the bytes at the offset without moving the file's cursor.
pub fn write_at(file: &File, offset: u64, bytes: &[u8]) -> Result<()> {
    #[cfg(unix)]
    {
        use std::os::unix::fs::FileExt;
        file.write_all_at(bytes, offset)?;
    }
    #[cfg(windows)]
    {
        use std::os::windows::fs::FileExt;
        let mut done = 0;
        while done < bytes.len() {
            done += file.seek_write(&bytes[done..], offset + done as u64)?;
        }
    }
    Ok(())
}

/// be_u32 reads a big-endian u32 at the offset of the bytes.
pub fn be_u32(bytes: &[u8], offset: usize) -> u32 {
    u32::from_be_bytes(bytes[offset..offset + 4].try_into().unwrap())
}

/// be_u64 reads a big-endian u64 at the offset of the bytes.
pub fn be_u64(bytes: &[u8], offset: usize) -> u64 {
    u64::from_be_bytes(bytes[offset..offset + 8].try_into().unwrap())
}

/// SPILL_LEN is how many bytes a spilling writer holds before moving them to its file.
pub(crate) const SPILL_LEN: usize = 8 << 20;

/// NEXT_SPILL numbers the spill files of the process, so that writers running at once never share one.
static NEXT_SPILL: std::sync::atomic::AtomicU64 = std::sync::atomic::AtomicU64::new(0);

/// Spill is a temporary file in a directory that a writer moves its bytes to as they pile up, hashing them on the way,
/// until it renames the file to its final name. A spill dropped before then deletes its file.
pub(crate) struct Spill {
    file: std::io::BufWriter<File>,
    path: Option<std::path::PathBuf>,
    hasher: sha2::Sha512,
}

impl Spill {
    /// create starts a spill file in a directory.
    pub(crate) fn create(dir: &std::path::Path) -> Result<Spill> {
        std::fs::create_dir_all(dir)?;
        let number = NEXT_SPILL.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
        let path = dir.join(format!(".spill-{}-{number}.tmp", std::process::id()));
        let file = std::io::BufWriter::new(File::create(&path)?);
        Ok(Spill { file, path: Some(path), hasher: sha2::Digest::new() })
    }

    /// write moves bytes to the file, emptying the buffer they were in.
    pub(crate) fn write(&mut self, bytes: &mut Vec<u8>) -> Result<()> {
        std::io::Write::write_all(&mut self.file, bytes)?;
        sha2::Digest::update(&mut self.hasher, &bytes[..]);
        bytes.clear();
        Ok(())
    }

    /// finish syncs the file and renames it to the path that the hash of everything written to it gives.
    pub(crate) fn finish(mut self, to: impl FnOnce(crate::hash::Hash) -> std::path::PathBuf) -> Result<()> {
        std::io::Write::flush(&mut self.file)?;
        self.file.get_ref().sync_all()?;
        let digest = sha2::Digest::finalize(std::mem::take(&mut self.hasher));
        let mut bytes = [0; crate::hash::Hash::LEN];
        bytes.copy_from_slice(&digest[..crate::hash::Hash::LEN]);
        let from = self.path.take().expect("an unfinished spill has its path");
        std::fs::rename(&from, to(crate::hash::Hash(bytes)))?;
        Ok(())
    }
}

impl Drop for Spill {
    fn drop(&mut self) {
        if let Some(path) = self.path.take() {
            let _ = std::fs::remove_file(path);
        }
    }
}
