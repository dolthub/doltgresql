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

use std::fmt;
use std::hash::{BuildHasherDefault, Hasher};

use sha2::{Digest, Sha512};

/// AddrHasher hashes chunk addresses by folding their bytes together, which suffices because addresses are already
/// uniformly random.
#[derive(Default)]
pub struct AddrHasher(u64);

impl Hasher for AddrHasher {
    fn finish(&self) -> u64 {
        self.0
    }

    fn write(&mut self, bytes: &[u8]) {
        for part in bytes.chunks(8) {
            let mut word = [0; 8];
            word[..part.len()].copy_from_slice(part);
            self.0 = self.0.rotate_left(29) ^ u64::from_le_bytes(word);
        }
    }
}

/// BuildAddrHasher builds the `AddrHasher` that maps and sets keyed by chunk addresses use.
pub type BuildAddrHasher = BuildHasherDefault<AddrHasher>;

/// ALPHABET is the base32 alphabet of hash strings, which is base32hex in lower case.
const ALPHABET: &[u8; 32] = b"0123456789abcdefghijklmnopqrstuv";

/// Hash is a chunk address: the first 20 bytes of the SHA-512 of the chunk's data.
#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Default)]
pub struct Hash(pub [u8; Hash::LEN]);

impl Hash {
    /// LEN is the number of bytes in a hash.
    pub const LEN: usize = 20;
    /// PREFIX_LEN is the number of bytes in a hash's prefix, which indexes store as a big-endian u64.
    pub const PREFIX_LEN: usize = 8;
    /// SUFFIX_LEN is the number of bytes after the prefix.
    pub const SUFFIX_LEN: usize = Hash::LEN - Hash::PREFIX_LEN;
    /// STRING_LEN is the number of characters in a hash string.
    pub const STRING_LEN: usize = 32;

    /// of returns the address of the data.
    pub fn of(data: &[u8]) -> Hash {
        let digest = Sha512::digest(data);
        let mut bytes = [0; Hash::LEN];
        bytes.copy_from_slice(&digest[..Hash::LEN]);
        Hash(bytes)
    }

    /// from_parts joins a prefix and a suffix into a hash.
    pub fn from_parts(prefix: u64, suffix: &[u8]) -> Hash {
        let mut bytes = [0; Hash::LEN];
        bytes[..Hash::PREFIX_LEN].copy_from_slice(&prefix.to_be_bytes());
        bytes[Hash::PREFIX_LEN..].copy_from_slice(suffix);
        Hash(bytes)
    }

    /// prefix returns the first 8 bytes as a big-endian integer.
    pub fn prefix(&self) -> u64 {
        u64::from_be_bytes(self.0[..Hash::PREFIX_LEN].try_into().unwrap())
    }

    /// suffix returns the bytes after the prefix.
    pub fn suffix(&self) -> &[u8] {
        &self.0[Hash::PREFIX_LEN..]
    }

    /// is_empty reports whether every byte is zero.
    pub fn is_empty(&self) -> bool {
        self.0 == [0; Hash::LEN]
    }

    /// parse parses a hash string, returning None when it is not 32 characters of the alphabet.
    pub fn parse(text: &str) -> Option<Hash> {
        let text = text.as_bytes();
        if text.len() != Hash::STRING_LEN {
            return None;
        }
        let mut bytes = [0; Hash::LEN];
        let mut buffer: u64 = 0;
        let mut bits = 0;
        let mut out = 0;
        for &c in text {
            let value = ALPHABET.iter().position(|&a| a == c)? as u64;
            buffer = (buffer << 5) | value;
            bits += 5;
            if bits >= 8 {
                bits -= 8;
                bytes[out] = (buffer >> bits) as u8;
                out += 1;
            }
        }
        Some(Hash(bytes))
    }
}

impl fmt::Display for Hash {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let mut text = String::with_capacity(Hash::STRING_LEN);
        let mut buffer: u64 = 0;
        let mut bits = 0;
        for &byte in &self.0 {
            buffer = (buffer << 8) | byte as u64;
            bits += 8;
            while bits >= 5 {
                bits -= 5;
                text.push(ALPHABET[((buffer >> bits) & 31) as usize] as char);
            }
        }
        f.write_str(&text)
    }
}

impl fmt::Debug for Hash {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "Hash({self})")
    }
}
