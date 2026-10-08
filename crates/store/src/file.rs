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
