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

/// sample returns about a number of the items of the tree at the root, spread evenly over its order, or every item when
/// the tree holds no more, finding each through the subtree counts of the internal nodes above it.
pub fn sample(store: &mut dyn NodeStore, root: Arc<Node>, count: usize) -> Result<Vec<(Vec<u8>, Vec<u8>)>> {
    let total = root.tree_count();
    let mut out = Vec::with_capacity(count.min(total as usize));
    if total <= count as u64 {
        let mut items = Items::first(store, root)?;
        while let Some((key, value)) = items.current()? {
            out.push((key.to_vec(), value.to_vec()));
            items.advance(store)?;
        }
        return Ok(out);
    }
    for i in 0..count as u64 {
        let mut ordinal = (2 * i + 1) * total / (2 * count as u64);
        let mut node = root.clone();
        while !node.is_leaf() {
            let mut child = 0;
            while child + 1 < node.count() && ordinal >= node.subtree_count(child)? {
                ordinal -= node.subtree_count(child)?;
                child += 1;
            }
            node = store.read(&node.child(child)?)?;
        }
        let index = (ordinal as usize).min(node.count().saturating_sub(1));
        out.push((node.key(index)?.to_vec(), node.value(index)?.to_vec()));
    }
    Ok(out)
}

/// Items walks the leaf items of a tree in either direction, reading nodes as it reaches them.
pub struct Items {
    /// The cursor, or None for an empty tree.
    cursor: Option<Cursor>,
}

impl Items {
    /// first returns a walk at the first item of the tree at the root.
    pub fn first(store: &mut dyn NodeStore, root: Arc<Node>) -> Result<Items> {
        Items::at_edge(store, root, false)
    }

    /// last returns a walk at the last item of the tree at the root.
    pub fn last(store: &mut dyn NodeStore, root: Arc<Node>) -> Result<Items> {
        Items::at_edge(store, root, true)
    }

    /// at_key returns a walk at the first item whose key is not less than the key.
    pub fn at_key(store: &mut dyn NodeStore, root: Arc<Node>, key: &[u8], compare: &Compare<'_>) -> Result<Items> {
        if root.count() == 0 {
            return Ok(Items { cursor: None });
        }
        Ok(Items { cursor: Some(Cursor::at_key(store, root, key, compare)?) })
    }

    /// at_edge returns a walk at the first or last item of the tree at the root.
    fn at_edge(store: &mut dyn NodeStore, root: Arc<Node>, last: bool) -> Result<Items> {
        if root.count() == 0 {
            return Ok(Items { cursor: None });
        }
        let mut levels = Vec::new();
        let mut node = root;
        loop {
            let idx = if last { node.count() as isize - 1 } else { 0 };
            let leaf = node.is_leaf();
            levels.push(Position { node: Some(node.clone()), idx });
            if leaf {
                break;
            }
            node = store.read(&node.child(idx as usize)?)?;
        }
        levels.reverse();
        Ok(Items { cursor: Some(Cursor { levels }) })
    }

    /// seek moves the walk to the first item whose key is not less than the key, searching only the nodes around its
    /// position that the key lies outside of.
    pub fn seek(&mut self, store: &mut dyn NodeStore, key: &[u8], compare: &Compare<'_>) -> Result<()> {
        match self.cursor.as_mut() {
            Some(cursor) => cursor.seek(0, key, store, compare),
            None => Ok(()),
        }
    }

    /// current returns the key and value of the item the walk is at, or None once it has passed either end.
    pub fn current(&self) -> Result<Option<(&[u8], &[u8])>> {
        let Some(cursor) = self.cursor.as_ref().filter(|c| c.valid(0)) else { return Ok(None) };
        let (node, idx) = (cursor.node(0), cursor.levels[0].idx as usize);
        Ok(Some((node.key(idx)?, node.value(idx)?)))
    }

    /// leaf returns the leaf node the walk is in and the index of its item there, or None once it has passed either end.
    pub fn leaf(&self) -> Option<(&Arc<Node>, usize)> {
        let cursor = self.cursor.as_ref().filter(|c| c.valid(0))?;
        Some((cursor.node(0), cursor.levels[0].idx as usize))
    }

    /// ordinal returns how many items of the tree come before the walk's position, adding up the subtree counts of
    /// the items before it at each level, which is the tree's count once the walk has passed its end.
    pub fn ordinal(&self) -> Result<u64> {
        let Some(cursor) = &self.cursor else { return Ok(0) };
        let leaf = &cursor.levels[0];
        if leaf.node.as_ref().is_none_or(|node| leaf.idx >= node.count() as isize) {
            return Ok(cursor.levels.last().and_then(|root| root.node.as_ref()).map_or(0, |root| root.tree_count()));
        }
        let mut total = 0;
        for (level, position) in cursor.levels.iter().enumerate() {
            let Some(node) = &position.node else { continue };
            let idx = position.idx.clamp(0, node.count() as isize) as usize;
            if level == 0 {
                total += idx as u64;
            } else {
                for i in 0..idx {
                    total += node.subtree_count(i)?;
                }
            }
        }
        Ok(total)
    }

    /// advance moves the walk to the next item.
    pub fn advance(&mut self, store: &mut dyn NodeStore) -> Result<()> {
        match self.cursor.as_mut() {
            Some(cursor) if cursor.valid(0) => cursor.advance(0, store),
            _ => Ok(()),
        }
    }

    /// retreat moves the walk to the previous item.
    pub fn retreat(&mut self, store: &mut dyn NodeStore) -> Result<()> {
        match self.cursor.as_mut() {
            Some(cursor) if cursor.valid(0) => cursor.retreat(0, store),
            _ => Ok(()),
        }
    }
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

    /// retreat moves the level to its previous item, moving into the previous node through the levels above, and
    /// leaves the level before its node's first item at the start of the tree.
    pub(crate) fn retreat(&mut self, level: usize, store: &mut dyn NodeStore) -> Result<()> {
        if self.levels[level].idx > 0 {
            self.levels[level].idx -= 1;
            return Ok(());
        }
        if !self.has_parent(level) {
            self.levels[level].idx = -1;
            return Ok(());
        }
        self.retreat(level + 1, store)?;
        if self.out_of_bounds(level + 1) {
            self.levels[level].idx = -1;
            return Ok(());
        }
        let (parent, idx) = self.item(level + 1);
        let node = store.read(&parent.child(idx)?)?;
        self.levels[level].idx = node.count() as isize - 1;
        self.levels[level].node = Some(node);
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
