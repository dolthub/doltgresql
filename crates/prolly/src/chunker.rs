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

//! The chunker that splits sorted items into tree nodes where Dolt's does, so that the same items always build the
//! same nodes.

use sha2::{Digest, Sha512};
use store::{Hash, Result};

/// MIN_CHUNK_SIZE is the size of items below which a node never ends.
const MIN_CHUNK_SIZE: u32 = 1 << 9;

/// MAX_CHUNK_SIZE is the size of items above which a node always ends.
const MAX_CHUNK_SIZE: u32 = 1 << 14;

/// TARGET_SIZE is the scale of the Weibull distribution of node sizes.
const TARGET_SIZE: f64 = 4096.0;

/// MAX_VECTOR_OFFSET is the largest total size of a node's keys and values, whose offsets are u16s.
const MAX_VECTOR_OFFSET: usize = u16::MAX as usize;

/// level_salt returns the hash seed of the splitter at the level, as Dolt's levelSalt does.
fn level_salt(level: u8) -> u64 {
    assert!(level < 15, "prolly trees have at most 15 levels");
    let full = Sha512::digest([level + 1]);
    u64::from_le_bytes(full[..8].try_into().unwrap())
}

/// KeySplitter decides where nodes end from the hash of each key and the size of the items so far, as Dolt's
/// keySplitter does.
struct KeySplitter {
    size: u32,
    crossed_boundary: bool,
    salt: u64,
}

impl KeySplitter {
    fn new(level: u8) -> KeySplitter {
        KeySplitter { size: 0, crossed_boundary: false, salt: level_salt(level) }
    }

    fn append(&mut self, key: &[u8], value: &[u8]) {
        let this_size = (key.len() + value.len()) as u32;
        self.size = self.size.wrapping_add(this_size);
        if self.size < MIN_CHUNK_SIZE {
            return;
        }
        if self.size > MAX_CHUNK_SIZE {
            self.crossed_boundary = true;
            return;
        }
        let hash = xxhash_rust::xxh3::xxh3_64_with_seed(key, self.salt) as u32;
        self.crossed_boundary = weibull_check(self.size, this_size, hash);
    }

    fn reset(&mut self) {
        self.size = 0;
        self.crossed_boundary = false;
    }
}

/// weibull_check reports whether a node ends at an item, with the probability that a Weibull distribution of node
/// sizes ends between the sizes before and after it.
fn weibull_check(size: u32, this_size: u32, hash: u32) -> bool {
    let pow = (size - this_size) as f64 / TARGET_SIZE;
    let start = -expm1(-(pow * pow * pow * pow));
    let pow = size as f64 / TARGET_SIZE;
    let end = -expm1(-(pow * pow * pow * pow));
    let p = hash as f64 / u32::MAX as f64;
    let d = 1.0 - start;
    if d <= 0.0 {
        return true;
    }
    p < (end - start) / d
}

/// mul_add returns `a * b + c`, fused into one rounding on the architectures where Go's compiler fuses it.
#[inline]
fn mul_add(a: f64, b: f64, c: f64) -> f64 {
    if cfg!(target_arch = "aarch64") { a.mul_add(b, c) } else { a * b + c }
}

/// expm1 returns e^x - 1 with Go's math.Expm1 algorithm, fusing the multiply-adds that Go fuses on this
/// architecture, so that it rounds exactly as Go does.
#[allow(clippy::excessive_precision, clippy::approx_constant)]
fn expm1(mut x: f64) -> f64 {
    const OTHRESHOLD: f64 = 7.09782712893383973096e+02;
    const LN2_X56: f64 = 3.88162421113569373274e+01;
    const LN2_HALF_X3: f64 = 1.03972077083991796413e+00;
    const LN2_HALF: f64 = 3.46573590279972654709e-01;
    const LN2_HI: f64 = 6.93147180369123816490e-01;
    const LN2_LO: f64 = 1.90821492927058770002e-10;
    const INV_LN2: f64 = 1.44269504088896338700e+00;
    const TINY: f64 = 1.0 / (1u64 << 54) as f64;
    const Q1: f64 = -3.33333333333331316428e-02;
    const Q2: f64 = 1.58730158725481460165e-03;
    const Q3: f64 = -7.93650757867487942473e-05;
    const Q4: f64 = 4.00821782732936239552e-06;
    const Q5: f64 = -2.01099218183624371326e-07;

    if x.is_nan() || x == f64::INFINITY {
        return x;
    }
    if x == f64::NEG_INFINITY {
        return -1.0;
    }
    let sign = x < 0.0;
    let absx = x.abs();
    if absx >= LN2_X56 {
        if sign {
            return -1.0;
        }
        if absx >= OTHRESHOLD {
            return f64::INFINITY;
        }
    }
    let mut c = 0.0;
    let k: i64;
    if absx > LN2_HALF {
        let (hi, lo);
        if absx < LN2_HALF_X3 {
            if !sign {
                hi = x - LN2_HI;
                lo = LN2_LO;
                k = 1;
            } else {
                hi = x + LN2_HI;
                lo = -LN2_LO;
                k = -1;
            }
        } else {
            k = if !sign { mul_add(INV_LN2, x, 0.5) as i64 } else { mul_add(INV_LN2, x, -0.5) as i64 };
            let t = k as f64;
            hi = mul_add(-t, LN2_HI, x);
            lo = t * LN2_LO;
        }
        x = hi - lo;
        c = (hi - x) - lo;
    } else if absx < TINY {
        return x;
    } else {
        k = 0;
    }
    let hfx = 0.5 * x;
    let hxs = x * hfx;
    let r1 = mul_add(hxs, mul_add(hxs, mul_add(hxs, mul_add(hxs, mul_add(hxs, Q5, Q4), Q3), Q2), Q1), 1.0);
    let t = mul_add(-r1, hfx, 3.0);
    let q = (r1 - t) / mul_add(-x, t, 6.0);
    // Go fuses a product into any sum that uses it, even through a variable such as e or hxs.
    if k == 0 {
        return x - mul_add(-x, hfx, x * (hxs * q));
    }
    let mut e = mul_add(x, mul_add(hxs, q, -c), -c);
    e = mul_add(-x, hfx, e);
    if k == -1 {
        return mul_add(0.5, x - e, -0.5);
    }
    if k == 1 {
        if x < -0.25 {
            return -2.0 * (e - (x + 0.5));
        }
        return mul_add(2.0, x - e, 1.0);
    }
    let add_exponent = |y: f64| f64::from_bits(y.to_bits().wrapping_add((k as u64) << 52));
    if k <= -2 || k > 56 {
        return add_exponent(1.0 - (e - x)) - 1.0;
    }
    if k < 20 {
        let t = f64::from_bits(0x3ff0000000000000 - (0x20000000000000u64 >> k));
        return add_exponent(t - (e - x));
    }
    let t = f64::from_bits(((0x3ff - k) as u64) << 52);
    add_exponent(x - (e + t) + 1.0)
}

/// NodeSerializer serializes the items of a tree node at a level.
pub trait NodeSerializer {
    fn serialize(&self, keys: &[&[u8]], values: &[&[u8]], subtrees: &[u64], level: u8) -> Vec<u8>;
}

/// NodeSink receives each node the chunker writes, with its address.
pub type NodeSink<'a> = dyn FnMut(Hash, &[u8]) -> Result<()> + 'a;

/// Written is the last node a level wrote.
struct Written {
    hash: Hash,
    bytes: Vec<u8>,
    count: usize,
    /// The address of the node's only child, when it is an internal node with one child.
    only_child: Option<Hash>,
}

/// Level is the node a level is building.
struct Level {
    splitter: KeySplitter,
    keys: Vec<Vec<u8>>,
    values: Vec<Vec<u8>>,
    subtrees: Vec<u64>,
    size: usize,
    last_written: Option<Written>,
}

impl Level {
    fn new(level: u8) -> Level {
        Level {
            splitter: KeySplitter::new(level),
            keys: Vec::new(),
            values: Vec::new(),
            subtrees: Vec::new(),
            size: 0,
            last_written: None,
        }
    }
}

/// Chunker builds a tree from items added in key order, writing each node to a sink, as Dolt's chunker does for a
/// tree built from empty.
pub struct Chunker<'a, S: NodeSerializer> {
    serializer: S,
    sink: &'a mut NodeSink<'a>,
    levels: Vec<Level>,
}

impl<'a, S: NodeSerializer> Chunker<'a, S> {
    pub fn new(serializer: S, sink: &'a mut NodeSink<'a>) -> Chunker<'a, S> {
        Chunker { serializer, sink, levels: vec![Level::new(0)] }
    }

    /// add appends a leaf item, which must sort after every item added before it.
    pub fn add(&mut self, key: &[u8], value: &[u8]) -> Result<()> {
        self.append(0, key, value, 1).map(|_| ())
    }

    /// append adds an item to the level, ending the level's node where the splitter or the node's capacity says,
    /// and returns whether it ended.
    fn append(&mut self, level: usize, key: &[u8], value: &[u8], subtree: u64) -> Result<bool> {
        let l = &self.levels[level];
        let degenerate = level > 0 && l.keys.len() == 1;
        let overflow = l.size + key.len() + value.len() > MAX_VECTOR_OFFSET;
        assert!(!(overflow && degenerate), "impossible node");
        if overflow {
            self.handle_boundary(level)?;
        }
        let l = &mut self.levels[level];
        l.keys.push(key.to_vec());
        l.values.push(value.to_vec());
        l.subtrees.push(subtree);
        l.size += key.len() + value.len();
        l.splitter.append(key, value);
        let degenerate = level > 0 && l.keys.len() == 1;
        if l.splitter.crossed_boundary && !degenerate {
            self.handle_boundary(level)?;
            return Ok(true);
        }
        Ok(false)
    }

    /// write_node serializes the level's node, writes it to the sink, and empties the level.
    fn write_node(&mut self, level: usize) -> Result<(Hash, Vec<u8>, u64)> {
        let l = &mut self.levels[level];
        let keys: Vec<&[u8]> = l.keys.iter().map(Vec::as_slice).collect();
        let values: Vec<&[u8]> = l.values.iter().map(Vec::as_slice).collect();
        let bytes = self.serializer.serialize(&keys, &values, &l.subtrees, level as u8);
        let hash = Hash::of(&bytes);
        (self.sink)(hash, &bytes)?;
        let last_key = l.keys.last().cloned().unwrap_or_default();
        let tree_count = if level == 0 { l.keys.len() as u64 } else { l.subtrees.iter().sum() };
        let only_child = if level > 0 && l.keys.len() == 1 { Some(serial::hash(&l.values[0])?) } else { None };
        l.last_written = Some(Written { hash, bytes, count: l.keys.len(), only_child });
        l.keys.clear();
        l.values.clear();
        l.subtrees.clear();
        l.size = 0;
        Ok((hash, last_key, tree_count))
    }

    /// handle_boundary ends the level's node and adds it to the level above.
    fn handle_boundary(&mut self, level: usize) -> Result<()> {
        assert!(!self.levels[level].keys.is_empty(), "in-progress chunk must be non-empty to create chunk boundary");
        let (hash, last_key, tree_count) = self.write_node(level)?;
        if self.levels.len() == level + 1 {
            self.levels.push(Level::new(level as u8 + 1));
        }
        self.append(level + 1, &last_key, &hash.0, tree_count)?;
        self.levels[level].splitter.reset();
        Ok(())
    }

    /// any_pending reports whether the level or one above it holds items.
    fn any_pending(&self, level: usize) -> bool {
        self.levels[level..].iter().any(|l| !l.keys.is_empty())
    }

    /// done finishes the tree and returns its root node's address and bytes.
    pub fn done(mut self) -> Result<(Hash, Vec<u8>)> {
        let mut level = 0;
        while self.levels.len() > level + 1 && self.any_pending(level + 1) {
            if !self.levels[level].keys.is_empty() {
                self.handle_boundary(level)?;
            }
            level += 1;
        }
        if level == 0 || self.levels[level].keys.len() > 1 {
            let (hash, _, _) = self.write_node(level)?;
            let bytes = self.levels[level].last_written.take().unwrap().bytes;
            return Ok((hash, bytes));
        }
        // The root has one child, so the tree's root is the highest node below with more than one item.
        let mut child = serial::hash(&self.levels[level].values[0])?;
        loop {
            level -= 1;
            let written = self.levels[level].last_written.take().expect("a level below the root wrote no node");
            assert_eq!(written.hash, child, "the root's child is not the last node written below it");
            if level == 0 || written.count > 1 {
                return Ok((written.hash, written.bytes));
            }
            child = written.only_child.unwrap();
        }
    }
}

#[cfg(test)]
mod tests {
    use sha2::{Digest, Sha256};

    use super::*;

    #[test]
    fn weibull_expm1_matches_go_for_every_chunk_size() {
        // The SHA-256 of the little-endian bits of -math.Expm1(-(size/4096)^4) for each size from 0 through
        // 16384, from Go 1.26 on each architecture, which differ because Go fuses multiply-adds on arm64.
        let expected = if cfg!(target_arch = "aarch64") {
            "908847d54f378ca4606eb2fed3d9b545a710b0667397458ca114c62e8089324b"
        } else {
            "9955985ff0e5e25d1734ec27d55180b58fbe9c027f0c3ec230741aba4e18bf0b"
        };
        let mut digest = Sha256::new();
        for size in 0..=MAX_CHUNK_SIZE {
            let pow = size as f64 / TARGET_SIZE;
            digest.update((-expm1(-(pow * pow * pow * pow))).to_bits().to_le_bytes());
        }
        let actual: String = digest.finalize().iter().map(|b| format!("{b:02x}")).collect();
        assert_eq!(actual, expected);
    }
}
