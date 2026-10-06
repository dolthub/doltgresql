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

use crate::error::{Result, corrupt};
use crate::hash::Hash;

/// CASTAGNOLI computes the CRC-32C checksums of chunk records.
const CASTAGNOLI: crc::Crc<u32> = crc::Crc::<u32>::new(&crc::CRC_32_ISCSI);

/// CHECKSUM_LEN is the length of the checksum that ends a compressed chunk record.
pub const CHECKSUM_LEN: usize = 4;

/// crc returns the CRC-32C checksum of the bytes.
pub fn crc(bytes: &[u8]) -> u32 {
    CASTAGNOLI.checksum(bytes)
}

/// Chunk is a chunk's data and its address.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Chunk {
    pub hash: Hash,
    pub data: Vec<u8>,
}

impl Chunk {
    /// new returns a chunk of the data, computing its address.
    pub fn new(data: Vec<u8>) -> Chunk {
        Chunk { hash: Hash::of(&data), data }
    }

    /// from_record decodes a compressed chunk record, which is snappy-compressed data followed by the big-endian
    /// CRC-32C of the compressed data, as table files, the journal, and snappy archive chunks store it.
    pub fn from_record(hash: Hash, record: &[u8]) -> Result<Chunk> {
        if record.len() < CHECKSUM_LEN {
            return Err(corrupt(format!("chunk record for {hash} is too short")));
        }
        let (compressed, checksum) = record.split_at(record.len() - CHECKSUM_LEN);
        if u32::from_be_bytes(checksum.try_into().unwrap()) != crc(compressed) {
            return Err(corrupt("checksum error"));
        }
        let data = snap::raw::Decoder::new()
            .decompress_vec(compressed)
            .map_err(|err| corrupt(format!("cannot decompress chunk {hash}: {err}")))?;
        Ok(Chunk { hash, data })
    }
    /// to_record encodes the chunk as a compressed chunk record.
    pub fn to_record(&self) -> Vec<u8> {
        let mut record = snap::raw::Encoder::new().compress_vec(&self.data).expect("chunks fit in snappy's limit");
        let checksum = crc(&record);
        record.extend_from_slice(&checksum.to_be_bytes());
        record
    }
}
