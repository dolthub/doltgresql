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

//! Cursors into trees, which hold a position at every level from a leaf up to the root, as Dolt's cursor does with
//! its chain of parents. A position at a level is that level's cursor, and the levels above it are its parents.

use std::cmp::Ordering;
use std::sync::Arc;

use store::{Hash, Result};

use crate::Node;

/// NodeStore reads the nodes of trees and stores the nodes that chunkers build.
pub trait NodeStore {
    /// read returns the node at the address.
    fn read(&mut self, hash: &Hash) -> Result<Arc<Node>>;

    /// write stores a node that a chunker built, returning it decoded.
    fn write(&mut self, hash: Hash, bytes: Vec<u8>) -> Result<Arc<Node>>;
}

/// Compare orders two keys.
pub type Compare<'a> = dyn Fn(&[u8], &[u8]) -> Ordering + 'a;

/// Visit receives an item and returns whether to continue.
pub type Visit<'a> = dyn FnMut(&[u8], &[u8]) -> Result<bool> + 'a;

/// Position is a cursor's node and item index at one level, where an index outside the node is out of bounds.
#[derive(Clone)]
pub(crate) struct Position {
    /// The node, which a finished cursor drops.
    pub(crate) node: Option<Arc<Node>>,
    pub(crate) idx: isize,
}

/// Cursor is a position at each level of a tree, with the leaf level first.
#[derive(Clone)]
pub(crate) struct Cursor {
    pub(crate) levels: Vec<Position>,
}

/// search returns the index of the first key of the node that is not less than the key, as Dolt's searchForKey does.
fn search(node: &Node, key: &[u8], compare: &Compare<'_>) -> Result<isize> {
    let (mut i, mut j) = (0, node.count());
    while i < j {
        let h = (i + j) / 2;
        if compare(key, node.key(h)?) == Ordering::Greater {
            i = h + 1;
        } else {
            j = h;
        }
    }
    Ok(i as isize)
}

/// get returns the value of a key in the tree at the root, if the tree has the key.
pub fn get(store: &mut dyn NodeStore, root: Arc<Node>, key: &[u8], compare: &Compare<'_>) -> Result<Option<Vec<u8>>> {
    if root.count() == 0 {
        return Ok(None);
    }
    let cursor = Cursor::at_key(store, root, key, compare)?;
    if !cursor.valid(0) {
        return Ok(None);
    }
    let (node, idx) = cursor.item(0);
    if compare(node.key(idx)?, key) != Ordering::Equal {
        return Ok(None);
    }
    Ok(Some(node.value(idx)?.to_vec()))
}

/// scan_from visits the items of the tree at the root in order, from the first key not less than the start, until the
/// visitor returns false.
pub fn scan_from(
    store: &mut dyn NodeStore,
    root: Arc<Node>,
    start: &[u8],
    compare: &Compare<'_>,
    visit: &mut Visit<'_>,
) -> Result<()> {
    if root.count() == 0 {
        return Ok(());
    }
    let mut cursor = Cursor::at_key(store, root, start, compare)?;
    while cursor.valid(0) {
        let (node, idx) = cursor.item(0);
        if !visit(node.key(idx)?, node.value(idx)?)? {
            break;
        }
        cursor.advance(0, store)?;
    }
    Ok(())
}

impl Cursor {
    /// at_key returns a cursor at the first key not less than the key, as Dolt's newCursorAtKey does.
    pub(crate) fn at_key(
        store: &mut dyn NodeStore,
        root: Arc<Node>,
        key: &[u8],
        compare: &Compare<'_>,
    ) -> Result<Cursor> {
        let mut levels = Vec::new();
        let mut node = root;
        loop {
            let idx = search(&node, key, compare)?;
            let leaf = node.is_leaf();
            levels.push(Position { node: Some(node.clone()), idx });
            if leaf {
                break;
            }
            let position = levels.last_mut().unwrap();
            keep_in_bounds(position);
            let child = node.child(position.idx as usize)?;
            node = store.read(&child)?;
        }
        levels.reverse();
        Ok(Cursor { levels })
    }

    /// has_parent reports whether a level has a level above it.
    pub(crate) fn has_parent(&self, level: usize) -> bool {
        level + 1 < self.levels.len()
    }

    /// node returns the node at the level.
    fn node(&self, level: usize) -> &Arc<Node> {
        self.levels[level].node.as_ref().expect("cursor has no node")
    }

    /// valid reports whether the level is at an item of its node.
    pub(crate) fn valid(&self, level: usize) -> bool {
        let position = &self.levels[level];
        position
            .node
            .as_ref()
            .is_some_and(|node| node.count() != 0 && position.idx >= 0 && (position.idx as usize) < node.count())
    }

    /// item returns the node and index of the level's current item.
    pub(crate) fn item(&self, level: usize) -> (Arc<Node>, usize) {
        (self.node(level).clone(), self.levels[level].idx as usize)
    }

    /// subtree_size returns the number of leaf items under the level's current item.
    pub(crate) fn subtree_size(&self, level: usize) -> Result<u64> {
        let node = self.node(level);
        if node.is_leaf() { Ok(1) } else { node.subtree_count(self.levels[level].idx as usize) }
    }

    /// at_node_end reports whether the level is at the last item of its node.
    pub(crate) fn at_node_end(&self, level: usize) -> bool {
        self.levels[level].idx == self.node(level).count() as isize - 1
    }

    pub(crate) fn skip_to_node_start(&mut self, level: usize) {
        self.levels[level].idx = 0;
    }

    /// invalidate_at_end moves the level just past the end of its node.
    pub(crate) fn invalidate_at_end(&mut self, level: usize) {
        self.levels[level].idx = self.node(level).count() as isize;
    }

    /// finish drops the level's node, which leaves it invalid.
    pub(crate) fn finish(&mut self, level: usize) {
        self.levels[level].node = None;
    }

    /// out_of_bounds reports whether the level's index is outside its node.
    fn out_of_bounds(&self, level: usize) -> bool {
        let idx = self.levels[level].idx;
        idx < 0 || idx >= self.node(level).count() as isize
    }

    /// advance moves the level to its next item, moving into the next node through the levels above, as Dolt's
    /// cursor.advance does.
    pub(crate) fn advance(&mut self, level: usize, store: &mut dyn NodeStore) -> Result<()> {
        if self.levels[level].idx < self.node(level).count() as isize - 1 {
            self.levels[level].idx += 1;
            return Ok(());
        }
        if !self.has_parent(level) {
            self.invalidate_at_end(level);
            return Ok(());
        }
        self.advance(level + 1, store)?;
        if self.out_of_bounds(level + 1) {
            self.invalidate_at_end(level);
            return Ok(());
        }
        let (parent, idx) = self.item(level + 1);
        self.levels[level].node = Some(store.read(&parent.child(idx)?)?);
        self.skip_to_node_start(level);
        Ok(())
    }

    /// compare compares the cursors from the level up, by the difference of their indexes at the highest level
    /// where they differ, as Dolt's compareCursors does.
    pub(crate) fn compare(&self, other: &Cursor, level: usize) -> isize {
        let mut diff = 0;
        let mut level = level;
        loop {
            let d = self.levels[level].idx - other.levels[level].idx;
            if d != 0 {
                diff = d;
            }
            if !self.has_parent(level) || !other.has_parent(level) {
                return diff;
            }
            level += 1;
        }
    }

    /// copy_from copies the other cursor's positions from the level up.
    pub(crate) fn copy_from(&mut self, other: &Cursor, level: usize) {
        assert_eq!(self.levels.len(), other.levels.len(), "cursors must be of equal height to call copy()");
        self.levels[level..].clone_from_slice(&other.levels[level..]);
    }

    /// seek moves the cursor to the first key not less than the key, from the level up as far as the key is outside
    /// the level's node, as Dolt's Seek does.
    pub(crate) fn seek(
        &mut self,
        level: usize,
        key: &[u8],
        store: &mut dyn NodeStore,
        compare: &Compare<'_>,
    ) -> Result<()> {
        let mut in_bounds = true;
        if self.has_parent(level) {
            let node = self.node(level);
            let first = compare(key, node.key(0)?);
            let last = compare(key, node.key(node.count() - 1)?);
            in_bounds = first != Ordering::Less && last != Ordering::Greater;
        }
        if !in_bounds {
            self.seek(level + 1, key, store, compare)?;
            keep_in_bounds(&mut self.levels[level + 1]);
            let (parent, idx) = self.item(level + 1);
            self.levels[level].node = Some(store.read(&parent.child(idx)?)?);
        }
        self.levels[level].idx = search(self.node(level), key, compare)?;
        Ok(())
    }
}

/// keep_in_bounds moves an index outside its node to the node's nearest item.
fn keep_in_bounds(position: &mut Position) {
    let count = position.node.as_ref().map_or(0, |node| node.count()) as isize;
    if position.idx < 0 {
        position.idx = 0;
    }
    if position.idx > count - 1 {
        position.idx = count - 1;
    }
}
