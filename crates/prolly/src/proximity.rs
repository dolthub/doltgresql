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

//! Proximity maps, the trees of Dolt's vector indexes, built as Dolt's ProximityMapBuilder builds them: every key's
//! level comes from its hash, and every key below the root sits under the closest key of the level above.

use std::collections::{BTreeMap, HashMap};

use serial::Builder;
use store::{ChunkReader, Hash};

use crate::Node;
use crate::blob::NodeSink;
use crate::serialize::{encode_counts, write_item_bytes};
use store::{Error, Result};

/// LOG_CHUNK_SIZE is the base-2 log of the average number of keys in a node, which Dolt always uses.
const LOG_CHUNK_SIZE: u8 = 8;

/// Distance is the distance function of a vector index, with Dolt's numbering of its DistanceType.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Distance {
    L2Squared = 1,
    Cosine = 2,
    InnerProduct = 3,
    L1 = 4,
}

impl Distance {
    /// from_stored returns the distance function of Dolt's number, where 0 means the default of squared L2.
    pub fn from_stored(stored: u8) -> Option<Distance> {
        Some(match stored {
            0 | 1 => Distance::L2Squared,
            2 => Distance::Cosine,
            3 => Distance::InnerProduct,
            4 => Distance::L1,
            _ => return None,
        })
    }

    /// eval returns the distance between two vectors, computed as go-mysql-server's distance types compute it.
    pub fn eval(self, left: &[f32], right: &[f32]) -> f64 {
        match self {
            Distance::L2Squared => left.iter().zip(right).map(|(l, r)| (l - r) as f64 * (l - r) as f64).sum(),
            Distance::Cosine => {
                let (mut dot, mut left_squared, mut right_squared) = (0f64, 0f64, 0f64);
                for (l, r) in left.iter().zip(right) {
                    dot += (l * r) as f64;
                    left_squared += (l * l) as f64;
                    right_squared += (r * r) as f64;
                }
                let (left_magnitude, right_magnitude) = (left_squared.sqrt(), right_squared.sqrt());
                if left_magnitude == 0.0 || right_magnitude == 0.0 {
                    return 0.0;
                }
                1.0 - dot / (left_magnitude * right_magnitude)
            }
            Distance::InnerProduct => left.iter().zip(right).fold(0f64, |total, (l, r)| total - *l as f64 * *r as f64),
            Distance::L1 => left.iter().zip(right).map(|(l, r)| (*l as f64 - *r as f64).abs()).sum(),
        }
    }
}

/// level returns the level of a key, from the leading zeros of its hash as Dolt's DeterministicHashLevel finds it.
fn level(key: &[u8]) -> u8 {
    let hash = xxhash_rust::xxh3::xxh3_64_with_seed(key, crate::chunker::level_salt(1)) as u32;
    hash.leading_zeros() as u8 / LOG_CHUNK_SIZE
}

/// Entry is a key of a proximity map, with its value and its vector.
pub struct Entry {
    pub key: Vec<u8>,
    pub value: Vec<u8>,
    pub vector: Vec<f32>,
}

/// write_proximity_map writes the proximity map of the entries, returning the address of its root.
pub fn write_proximity_map(entries: Vec<Entry>, distance: Distance, sink: &mut NodeSink<'_>) -> Result<Hash> {
    let mut write = |keys: &[&[u8]], values: &[&[u8]], subtrees: &[u64], level: u8| -> Result<(Hash, u64)> {
        let bytes = serialize_vector_index_node(keys, values, subtrees, level, distance);
        let hash = Hash::of(&bytes);
        sink(hash, &bytes)?;
        let count = if level == 0 { keys.len() as u64 } else { subtrees.iter().sum() };
        Ok((hash, count))
    };
    let mut leveled: Vec<(u8, usize)> = entries.iter().enumerate().map(|(i, e)| (level(&e.key), i)).collect();
    leveled.sort_by(|a, b| b.0.cmp(&a.0).then_with(|| entries[a.1].key.cmp(&entries[b.1].key)));
    leveled.dedup_by(|a, b| entries[a.1].key == entries[b.1].key);
    let Some(&(max_level, _)) = leveled.first() else { return Ok(write(&[], &[], &[], 0)?.0) };
    if max_level == 0 {
        let keys: Vec<&[u8]> = leveled.iter().map(|&(_, i)| entries[i].key.as_slice()).collect();
        let values: Vec<&[u8]> = leveled.iter().map(|&(_, i)| entries[i].value.as_slice()).collect();
        return Ok(write(&keys, &values, &[], 0)?.0);
    }
    let by_key: HashMap<&[u8], usize> = entries.iter().enumerate().map(|(i, e)| (e.key.as_slice(), i)).collect();
    // path_maps[level] maps the path from the root to each key of the level, ending with the key, to its entry.
    let mut path_maps: Vec<BTreeMap<Vec<&[u8]>, usize>> = vec![BTreeMap::new(); max_level as usize + 1];
    for &(key_level, i) in &leveled {
        let depth = (max_level - key_level) as usize;
        let mut path: Vec<&[u8]> = Vec::with_capacity(max_level as usize + 1);
        for path_depth in 0..depth {
            let lookup = &path_maps[max_level as usize - path_depth];
            let mut closest: Option<(&[u8], f64)> = None;
            for candidate in lookup.range(path.clone()..).take_while(|(p, _)| p[..path_depth] == path[..]) {
                let candidate_key = candidate.0[path_depth];
                let candidate_distance = distance.eval(&entries[i].vector, &entries[by_key[candidate_key]].vector);
                if closest.is_none_or(|(_, d)| candidate_distance < d) {
                    closest = Some((candidate_key, candidate_distance));
                }
            }
            let (closest, _) = closest.ok_or_else(|| Error::Corrupt("a proximity map level without keys".into()))?;
            path.push(closest);
        }
        path.push(&entries[i].key);
        path_maps[key_level as usize].insert(path.clone(), i);
        for child_level in (0..key_level).rev() {
            path.push(&entries[i].key);
            path_maps[child_level as usize].insert(path.clone(), i);
        }
    }
    // Each level's paths, in order, group under the paths of the level above.
    let mut cursors: Vec<Cursor<'_>> = path_maps.into_iter().map(|m| m.into_iter().peekable()).collect();
    let roots: Vec<(Vec<&[u8]>, usize)> = cursors.pop().map(|c| c.collect()).unwrap_or_default();
    let (mut keys, mut values, mut subtrees) = (Vec::new(), Vec::new(), Vec::new());
    for (path, _) in roots {
        let key = path[0];
        let (hash, count) = chunk(&mut cursors, max_level as usize - 1, key, &entries, &mut write)?;
        keys.push(key);
        values.push(hash);
        subtrees.push(count);
    }
    let addresses: Vec<&[u8]> = values.iter().map(|h: &Hash| h.0.as_slice()).collect();
    Ok(write(&keys, &addresses, &subtrees, max_level)?.0)
}

/// Cursor walks the paths of one level of a proximity map in order, with each path's entry.
type Cursor<'e> = std::iter::Peekable<std::collections::btree_map::IntoIter<Vec<&'e [u8]>, usize>>;

/// Write serializes and stores a node, returning its address and the number of keys under it.
type Write<'w> = dyn FnMut(&[&[u8]], &[&[u8]], &[u64], u8) -> Result<(Hash, u64)> + 'w;

/// chunk writes the node of a level that holds the keys whose paths go through a parent key, as Dolt's
/// vectorIndexChunker does, returning its address and the number of keys under it.
fn chunk<'e>(
    cursors: &mut [Cursor<'e>],
    level: usize,
    parent: &[u8],
    entries: &'e [Entry],
    write: &mut Write<'_>,
) -> Result<(Hash, u64)> {
    let (mut keys, mut values, mut subtrees) = (Vec::new(), Vec::new(), Vec::new());
    let mut children = Vec::new();
    while let Some((path, i)) = cursors[level].next_if(|(p, _)| p[p.len() - 2] == parent) {
        let key = path[path.len() - 1];
        if level > 0 {
            let (hash, count) = chunk(cursors, level - 1, key, entries, write)?;
            children.push(hash);
            subtrees.push(count);
        } else {
            values.push(entries[i].value.as_slice());
        }
        keys.push(key);
    }
    if level > 0 {
        let addresses: Vec<&[u8]> = children.iter().map(|h| h.0.as_slice()).collect();
        write(&keys, &addresses, &subtrees, level as u8)
    } else {
        write(&keys, &values, &[], 0)
    }
}

/// write_item_offsets32 writes the u32 offsets of the items, starting at 0 and ending at their total length, into a
/// vector the caller started.
fn write_item_offsets32(b: &mut Builder, items: &[&[u8]]) -> u32 {
    let mut offset: usize = items.iter().map(|item| item.len()).sum();
    for item in items.iter().rev() {
        b.prepend_u32(offset as u32);
        offset -= item.len();
    }
    b.prepend_u32(offset as u32);
    b.end_vector(items.len() + 1)
}

/// serialize_vector_index_node serializes a VectorIndexNode as Dolt's VectorIndexSerializer does.
fn serialize_vector_index_node(
    keys: &[&[u8]],
    values: &[&[u8]],
    subtrees: &[u64],
    level: u8,
    distance: Distance,
) -> Vec<u8> {
    let mut b = Builder::new(0);
    let key_items = write_item_bytes(&mut b, keys);
    b.start_vector(4, keys.len() + 1, 4);
    let key_offsets = write_item_offsets32(&mut b, keys);
    let (mut value_items, mut value_offsets, mut address_array, mut counts) = (0, 0, 0, 0);
    if level == 0 {
        value_items = write_item_bytes(&mut b, values);
        b.start_vector(4, values.len() + 1, 4);
        value_offsets = write_item_offsets32(&mut b, values);
    } else {
        address_array = write_item_bytes(&mut b, values);
        counts = b.create_byte_vector(&encode_counts(subtrees));
    }
    b.start_object(10);
    b.add_offset(0, key_items);
    b.add_offset(1, key_offsets);
    if level == 0 {
        b.add_offset(2, value_items);
        b.add_offset(3, value_offsets);
        b.add_u64(6, keys.len() as u64, 0);
    } else {
        b.add_offset(4, address_array);
        b.add_offset(5, counts);
        b.add_u64(6, subtrees.iter().sum(), 0);
    }
    b.add_u8(7, level, 0);
    b.add_u8(8, LOG_CHUNK_SIZE, 0);
    b.add_u8(9, distance as u8, 0);
    let root = b.end_object();
    b.finish_message(root, serial::VECTOR_INDEX_NODE)
}

/// Candidate is a key of a proximity map, with its value and its distance from a query vector.
struct Candidate {
    key: Vec<u8>,
    value: Vec<u8>,
    distance: f64,
}

/// Candidates holds the closest keys found so far in a min-max heap ordered by distance, laid out as the
/// esote/minmaxheap package lays out Dolt's DistancePriorityHeap, so that ties between equal distances fall as they
/// fall in Dolt.
struct Candidates {
    items: Vec<Candidate>,
    capacity: usize,
}

impl Candidates {
    /// new returns an empty heap that keeps at most `limit` candidates.
    fn new(limit: usize) -> Candidates {
        Candidates { items: Vec::with_capacity(limit + 1), capacity: limit + 1 }
    }

    /// less reports whether the candidate at `i` is closer than the one at `j`.
    fn less(&self, i: usize, j: usize) -> bool {
        self.items[i].distance < self.items[j].distance
    }

    /// is_min_level reports whether a position lies on a level of the heap whose items are smaller than their
    /// descendants.
    fn is_min_level(i: usize) -> bool {
        (usize::BITS - (i + 1).leading_zeros() - 1).is_multiple_of(2)
    }

    /// down moves the item at `i` down among the first `n` items, reporting whether it moved.
    fn down(&mut self, i0: usize, n: usize) -> bool {
        let min = Self::is_min_level(i0);
        let mut i = i0;
        loop {
            let mut m = i;
            let l = i * 2 + 1;
            if l >= n {
                break;
            }
            if self.less(l, m) == min {
                m = l;
            }
            let r = i * 2 + 2;
            if r < n && self.less(r, m) == min {
                m = r;
            }
            let mut g = l * 2 + 1;
            while g < n && g <= r * 2 + 2 {
                if self.less(g, m) == min {
                    m = g;
                }
                g += 1;
            }
            if m == i {
                break;
            }
            self.items.swap(i, m);
            if m == l || m == r {
                break;
            }
            let p = (m - 1) / 2;
            if self.less(p, m) == min {
                self.items.swap(m, p);
            }
            i = m;
        }
        i > i0
    }

    /// up moves the item at `i` up to its place.
    fn up(&mut self, mut i: usize) {
        let mut min = Self::is_min_level(i);
        if i > 0 {
            let p = (i - 1) / 2;
            if self.less(p, i) == min {
                self.items.swap(i, p);
                min = !min;
                i = p;
            }
        }
        while i > 2 {
            let g = ((i - 1) / 2 - 1) / 2;
            if self.less(i, g) != min {
                return;
            }
            self.items.swap(i, g);
            i = g;
        }
    }

    /// insert adds a candidate, dropping the farthest when the heap holds more than its limit.
    fn insert(&mut self, candidate: Candidate) {
        self.items.push(candidate);
        self.up(self.items.len() - 1);
        if self.items.len() == self.capacity {
            self.pop_max();
        }
    }

    /// pop_max removes the farthest candidate.
    fn pop_max(&mut self) {
        let n = self.items.len();
        let mut i = 0;
        if 1 < n && !self.less(1, i) {
            i = 1;
        }
        if 2 < n && !self.less(2, i) {
            i = 2;
        }
        self.items.swap(i, n - 1);
        self.down(i, n - 1);
        self.items.pop();
    }

    /// pop removes the closest candidate.
    fn pop(&mut self) -> Option<Candidate> {
        let n = self.items.len().checked_sub(1)?;
        self.items.swap(0, n);
        self.down(0, n);
        self.items.pop()
    }
}

/// KeyValue is a key of a map with its value.
pub type KeyValue = (Vec<u8>, Vec<u8>);

/// closest returns up to `limit` keys of a proximity map with their values, closest first, by descending from the root
/// through the children of the closest keys of each level, as Dolt's GetClosest does. `distance_of` gives a key's
/// distance from the query vector.
pub fn closest<E: From<Error>>(
    reader: &dyn ChunkReader,
    root: &Node,
    limit: usize,
    distance_of: &mut dyn FnMut(&[u8]) -> std::result::Result<f64, E>,
) -> std::result::Result<Vec<KeyValue>, E> {
    if limit == 0 {
        return Ok(Vec::new());
    }
    let mut add = |candidates: &mut Candidates, node: &Node| -> std::result::Result<(), E> {
        for i in 0..node.count() {
            let key = node.key(i)?;
            let distance = distance_of(key)?;
            candidates.insert(Candidate { key: key.to_vec(), value: node.value(i)?.to_vec(), distance });
        }
        Ok(())
    };
    let mut candidates = Candidates::new(limit);
    add(&mut candidates, root)?;
    for _ in 0..root.level() {
        let mut next = Candidates::new(limit);
        for candidate in &candidates.items {
            add(&mut next, &Node::load(reader, &serial::hash(&candidate.value)?)?)?;
        }
        candidates = next;
    }
    let mut out = Vec::with_capacity(candidates.items.len());
    while let Some(candidate) = candidates.pop() {
        out.push((candidate.key, candidate.value));
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn proximity_maps_match_go() {
        let entries = || {
            (1..=300)
                .map(|id: i32| {
                    let vector = vec![(id % 17) as f32, (id % 23) as f32];
                    let mut field = vec![0];
                    field.extend(vector.iter().flat_map(|v| v.to_le_bytes()));
                    let key = crate::val::build_tuple(&[Some(field.as_slice()), Some(&id.to_le_bytes())]);
                    Entry { key, value: crate::val::build_tuple(&[]), vector }
                })
                .collect()
        };
        let root = |distance| {
            let mut sink = |_: Hash, _: &[u8]| Ok(());
            write_proximity_map(entries(), distance, &mut sink).unwrap().to_string()
        };
        assert_eq!(root(Distance::L2Squared), "4ff5t10qjr16c277clg2uqle2vom3mbs");
        assert_eq!(root(Distance::Cosine), "p9btfg8iejhf7d6sio2q40fetreo0a6e");
    }
}
