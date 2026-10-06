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

//! Reading Dolt's commit graph: commits, refs, and the specs that name commits.

use std::collections::HashSet;

use doltdb::database::Database;
use serial::{Commit, Message, Tag};
use store::Hash;

use crate::dolt::args::error;
use crate::error::{PgError, Result};
use crate::txn::read;

/// CommitInfo is a commit's contents.
#[derive(Clone, Debug)]
pub struct CommitInfo {
    pub hash: Hash,
    pub root: Hash,
    pub parents: Vec<Hash>,
    pub height: u64,
    /// The author's name and email.
    pub name: String,
    pub email: String,
    pub description: String,
    /// The committer's time and the author's time, in Unix milliseconds.
    pub committer_millis: u64,
    pub author_millis: i64,
    pub signature: String,
    pub committer_name: String,
    pub committer_email: String,
}

/// text decodes stored text.
fn text(bytes: &[u8]) -> String {
    String::from_utf8_lossy(bytes).into_owned()
}

/// load reads the commit at the address.
pub fn load(db: &Database, hash: Hash) -> Result<CommitInfo> {
    let data = read(db, &hash)?;
    let c = Commit::new(Message(&data))?;
    let name = text(c.name()?);
    let email = text(c.email()?);
    Ok(CommitInfo {
        hash,
        root: c.root()?,
        parents: c.parents()?,
        height: c.height()?,
        committer_name: c.committer_name()?.map_or_else(|| name.clone(), text),
        committer_email: c.committer_email()?.map_or_else(|| email.clone(), text),
        name,
        email,
        description: text(c.description()?),
        committer_millis: c.timestamp_millis()?,
        author_millis: c.user_timestamp_millis()?,
        signature: c.signature()?.map(text).unwrap_or_default(),
    })
}

/// is_hash reports whether text looks like a commit hash.
pub fn is_hash(text: &str) -> bool {
    text.len() == 32 && text.bytes().all(|b| b.is_ascii_digit() || (b'a'..=b'v').contains(&b))
}

/// valid_ref_name reports whether a name is a valid dataset name, as Dolt's ValidateDatasetId checks.
fn valid_ref_name(name: &str) -> bool {
    if name.is_empty() || name == "@" || name.ends_with('/') || name.ends_with('.') {
        return false;
    }
    name.split('/').all(|component| {
        !component.is_empty()
            && !component.starts_with('.')
            && !component.ends_with(".lock")
            && !component.contains("..")
            && !component.contains("@{")
            && component.bytes().all(|b| b.is_ascii() && b > 0x1f && b != 0x7f && !b":?[\\^~ *".contains(&b))
    })
}

/// valid_branch_name reports whether a name can name a branch.
pub fn valid_branch_name(name: &str) -> bool {
    !matches!(name, "" | "HEAD" | "-") && !is_hash(name) && !name.contains("//") && valid_ref_name(name)
}

/// valid_tag_name reports whether a name can name a tag.
pub fn valid_tag_name(name: &str) -> bool {
    valid_branch_name(name)
}

/// refs returns the name and address of every ref under the prefix, such as `refs/heads/`, in name order.
pub fn refs(db: &mut Database, prefix: &str) -> Result<Vec<(String, Hash)>> {
    Ok(db
        .datasets()?
        .into_iter()
        .filter_map(|(name, hash)| name.strip_prefix(prefix).map(|n| (n.to_string(), hash)))
        .collect())
}

/// commit_of returns the commit a ref's address names, following a tag to its commit.
pub fn commit_of(db: &Database, address: Hash) -> Result<Hash> {
    let data = read(db, &address)?;
    match Tag::new(Message(&data)) {
        Ok(tag) => Ok(tag.commit()?),
        Err(_) => Ok(address),
    }
}

/// ancestor_instructions parses an ancestor spec such as `~2^2` into the parent index to follow at each step.
fn ancestor_instructions(spec: &str) -> Result<Vec<usize>> {
    let bytes = spec.as_bytes();
    let mut instructions = Vec::new();
    let mut i = 0;
    while i < bytes.len() {
        let op = bytes[i];
        let start = i;
        while i + 1 < bytes.len() && bytes[i + 1].is_ascii_digit() {
            i += 1;
        }
        let num =
            if start == i { 1 } else { spec[start + 1..=i].parse::<usize>().map_err(|e| error(e.to_string()))? };
        match op {
            b'^' if num == 1 || num == 2 => instructions.push(num - 1),
            b'^' => return Err(error("invalid ancestor spec")),
            b'~' => instructions.extend(std::iter::repeat_n(0, num)),
            _ => return Err(error(format!("Invalid HEAD spec: {spec}"))),
        }
        i += 1;
    }
    Ok(instructions)
}

/// resolve returns the commit that a commit spec names: HEAD, a commit hash, or a branch, tag, or remote ref, with
/// an optional ancestor spec, as Dolt's Resolve does.
pub fn resolve(db: &mut Database, head: Hash, spec: &str) -> Result<Hash> {
    let spec = spec.trim();
    let split = spec.find(['^', '~']).unwrap_or(spec.len());
    let (name, ancestry) = spec.split_at(split);
    let instructions = ancestor_instructions(ancestry)?;
    let mut commit = if name.eq_ignore_ascii_case("head") {
        head
    } else if is_hash(name) {
        let hash = Hash::parse(name).ok_or_else(|| error(format!("invalid hash: {name}")))?;
        if db.read_value(&hash)?.is_none() {
            return Err(error("target commit not found"));
        }
        hash
    } else {
        if !valid_branch_name(name) {
            return Err(error("string is not a valid branch or hash"));
        }
        let mut candidates = vec![
            format!("refs/{name}"),
            format!("refs/heads/{name}"),
            format!("refs/tags/{name}"),
            format!("refs/remotes/{name}"),
        ];
        if name.starts_with("refs/") {
            candidates.insert(0, name.to_string());
        }
        let datasets = db.datasets()?;
        let found = candidates.iter().find_map(|c| datasets.iter().find(|(n, _)| n == c).map(|(_, h)| *h));
        let address = found.ok_or_else(|| error(format!("branch not found: {name}")))?;
        commit_of(db, address)?
    };
    for index in instructions {
        let parents = load(db, commit)?.parents;
        commit = *parents.get(index).ok_or_else(|| error("invalid ancestor spec"))?;
    }
    Ok(commit)
}

/// log returns the commits reachable from the starting commits, newest first, in Dolt's topological order: greater
/// heights first, then newer author times.
pub fn log(db: &Database, starts: &[Hash]) -> Result<Vec<CommitInfo>> {
    let mut seen: HashSet<Hash> = HashSet::new();
    let mut pending: Vec<CommitInfo> = Vec::new();
    for &start in starts {
        if seen.insert(start) {
            pending.push(load(db, start)?);
        }
    }
    let mut out = Vec::new();
    while !pending.is_empty() {
        let best = (0..pending.len())
            .max_by(|&a, &b| {
                let (a, b) = (&pending[a], &pending[b]);
                a.height.cmp(&b.height).then(a.author_millis.cmp(&b.author_millis))
            })
            .unwrap_or(0);
        let commit = pending.swap_remove(best);
        for &parent in &commit.parents {
            if seen.insert(parent) {
                pending.push(load(db, parent)?);
            }
        }
        out.push(commit);
    }
    Ok(out)
}

/// is_ancestor reports whether a commit is an ancestor of, or the same as, another.
pub fn is_ancestor(db: &Database, ancestor: Hash, of: Hash) -> Result<bool> {
    if ancestor == of {
        return Ok(true);
    }
    let target = load(db, ancestor)?.height;
    let mut stack = vec![of];
    let mut seen = HashSet::new();
    while let Some(hash) = stack.pop() {
        if hash == ancestor {
            return Ok(true);
        }
        if !seen.insert(hash) {
            continue;
        }
        let commit = load(db, hash)?;
        if commit.height > target {
            stack.extend(commit.parents);
        }
    }
    Ok(false)
}

/// merge_base returns the best common ancestor of two commits, the one with the greatest height.
pub fn merge_base(db: &Database, left: Hash, right: Hash) -> Result<Option<Hash>> {
    let mut left_ancestors = HashSet::new();
    let mut stack = vec![left];
    while let Some(hash) = stack.pop() {
        if left_ancestors.insert(hash) {
            stack.extend(load(db, hash)?.parents);
        }
    }
    let mut best: Option<CommitInfo> = None;
    let mut stack = vec![right];
    let mut seen = HashSet::new();
    while let Some(hash) = stack.pop() {
        if !seen.insert(hash) {
            continue;
        }
        let commit = load(db, hash)?;
        if left_ancestors.contains(&hash) {
            if best.as_ref().is_none_or(|b| commit.height > b.height) {
                best = Some(commit);
            }
            continue;
        }
        stack.extend(commit.parents);
    }
    Ok(best.map(|c| c.hash))
}

/// branch_not_found returns Dolt's error for a missing branch.
pub fn branch_not_found(name: &str) -> PgError {
    error(format!("branch not found: {name}"))
}
