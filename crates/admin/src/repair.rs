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

use std::collections::{HashMap, HashSet};

use doltdb::create::working_set_ref;
use doltdb::database::{CommitMeta, Database};
use doltdb::root::Root;
use serial::write::{Meta, write_tag};
use serial::{Commit, Message, Tag};
use sql::PgError;
use sql::integrity::{Scanner, tables_for_root};
use sql::txn::{read, read_working_set};
use store::Hash;

use crate::log;
use crate::rewrite::TreeRewriter;

/// RepairSummary counts what a repair of one database rewrote.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub struct RepairSummary {
    pub commits_examined: u64,
    pub commits_rewritten: u64,
    pub branches_updated: u64,
    pub tags_updated: u64,
    pub working_sets_fixed: u64,
    pub leaf_chunks_rewritten: u64,
    pub internal_chunks_rewritten: u64,
}

/// Repairer rewrites the corruption that its rewriter's scanner finds across a database's whole commit graph, moving
/// branches, tags, and working sets to the rewritten history. A commit whose contents and ancestors need no repair
/// keeps its hash.
pub struct Repairer<'d> {
    pub rewriter: TreeRewriter<'d>,
    pub summary: RepairSummary,
    verbose: bool,
}

impl<'d> Repairer<'d> {
    /// new returns a repairer built on the scanner, which shares what an earlier scan with it found.
    pub fn new(scanner: Scanner<'d>, verbose: bool) -> Repairer<'d> {
        Repairer { rewriter: TreeRewriter::new(scanner), summary: RepairSummary::default(), verbose }
    }

    /// db returns the database being repaired.
    fn db(&mut self) -> &mut Database {
        self.rewriter.scanner.db
    }

    /// repair_database rewrites every commit that a branch or tag reaches, then moves the branches, tags, and working
    /// sets to the rewritten commits.
    pub fn repair_database(&mut self) -> sql::Result<RepairSummary> {
        let datasets = self.db().datasets()?;
        let mut visited = HashMap::new();
        for (dataset, head) in &datasets {
            let Some(branch) = dataset.strip_prefix("refs/heads/") else { continue };
            let original = match self.db().head(&working_set_ref(branch))? {
                Some(address) => {
                    let working_set = read_working_set(self.db(), &address)?;
                    Some((working_set.working_root, working_set.staged_root))
                }
                None => None,
            };
            let new_head = self.repair_commit(*head, &mut visited).map_err(|e| {
                PgError::internal(format!("failed to repair history of branch {dataset}: {}", e.message))
            })?;
            if new_head != *head {
                sql::dolt::procedures::new_branch(self.db(), branch, new_head)
                    .map_err(|e| PgError::internal(format!("failed to update branch {dataset}: {}", e.message)))?;
                self.summary.branches_updated += 1;
                if self.verbose {
                    log(&format!("updated branch {dataset} to repaired commit"));
                }
            }
            if let Some((working, staged)) = original {
                self.restore_working_set(branch, working, staged).map_err(|e| {
                    PgError::internal(format!("failed to repair working set of branch {dataset}: {}", e.message))
                })?;
            }
        }
        for (dataset, address) in &datasets {
            if !dataset.starts_with("refs/tags/") {
                continue;
            }
            let data = read(self.db(), address)?;
            let tag = Tag::new(Message(&data))?;
            let commit = tag.commit()?;
            let meta = Meta {
                name: tag.name()?.to_vec(),
                email: tag.email()?.to_vec(),
                description: tag.description()?.to_vec(),
                timestamp_millis: tag.timestamp_millis()?,
                user_timestamp_millis: tag.user_timestamp_millis()?,
            };
            let new_commit = self
                .repair_commit(commit, &mut visited)
                .map_err(|e| PgError::internal(format!("failed to repair history of tag {dataset}: {}", e.message)))?;
            if new_commit != commit {
                let tag = self.db().write_value(write_tag(new_commit, Some(&meta)))?;
                self.db().set_head(dataset, tag)?;
                self.summary.tags_updated += 1;
                if self.verbose {
                    log(&format!("updated tag {dataset} to repaired commit"));
                }
            }
        }
        self.summary.leaf_chunks_rewritten = self.rewriter.leaf_chunks_rewritten;
        self.summary.internal_chunks_rewritten = self.rewriter.internal_chunks_rewritten;
        Ok(self.summary)
    }

    /// restore_working_set writes a branch's repaired original working and staged roots into its working set when they
    /// differ from what it holds, which repairs uncommitted data and restores the changes that moving the branch reset.
    fn restore_working_set(&mut self, branch: &str, working: Hash, staged: Option<Hash>) -> sql::Result<()> {
        let dataset = working_set_ref(branch);
        let Some(address) = self.db().head(&dataset)? else { return Ok(()) };
        let mut working_set = read_working_set(self.db(), &address)?;
        let mut changed = false;
        let repaired = self.repair_root_value(working)?.0;
        if repaired != working_set.working_root {
            working_set.working_root = repaired;
            changed = true;
        }
        if let Some(staged) = staged {
            let repaired = self.repair_root_value(staged)?.0;
            if Some(repaired) != working_set.staged_root {
                working_set.staged_root = Some(repaired);
                changed = true;
            }
        }
        if !changed {
            return Ok(());
        }
        self.db().update_working_set(&dataset, &working_set, address)?;
        self.summary.working_sets_fixed += 1;
        if self.verbose {
            log(&format!("repaired working set {dataset}"));
        }
        Ok(())
    }

    /// repair_commit repairs a commit after its ancestors, returning the repaired commit, which is the commit itself
    /// when neither its root value nor an ancestor changed.
    fn repair_commit(&mut self, start: Hash, visited: &mut HashMap<Hash, Hash>) -> sql::Result<Hash> {
        let mut examined = HashSet::new();
        let mut stack = vec![start];
        while let Some(&hash) = stack.last() {
            if visited.contains_key(&hash) {
                stack.pop();
                continue;
            }
            let data = read(self.db(), &hash)?;
            let commit = Commit::new(Message(&data))?;
            let parents = commit.parents()?;
            if examined.insert(hash) {
                self.summary.commits_examined += 1;
                for parent in parents.iter().rev() {
                    if self.db().read_value(parent)?.is_none() {
                        return Err(PgError::internal(format!(
                            "commit {hash} has a ghost parent: cannot repair shallow clones"
                        )));
                    }
                    if !visited.contains_key(parent) {
                        stack.push(*parent);
                    }
                }
                continue;
            }
            stack.pop();
            let new_parents: Vec<Hash> = parents.iter().map(|p| visited[p]).collect();
            let (root, root_changed) = self.repair_root_value(commit.root()?).map_err(|e| {
                PgError::internal(format!("failed to repair root value of commit {hash}: {}", e.message))
            })?;
            if !root_changed && new_parents == parents {
                visited.insert(hash, hash);
                continue;
            }
            let text = |bytes: &[u8]| String::from_utf8_lossy(bytes).into_owned();
            let meta = CommitMeta {
                name: text(commit.name()?),
                email: text(commit.email()?),
                description: text(commit.description()?),
                author_millis: commit.user_timestamp_millis()?,
                committer_millis: commit.timestamp_millis()?,
                signature: commit.signature()?.map(text).unwrap_or_default(),
                committer_name: commit.committer_name()?.map(text),
                committer_email: commit.committer_email()?.map(text),
            };
            let root_value = read(self.db(), &root)?;
            let rewritten = self.db().build_commit(None, root_value, new_parents, &meta)?;
            self.db().write_value(rewritten.bytes.clone())?;
            self.summary.commits_rewritten += 1;
            if self.verbose {
                log(&format!("rewrote commit {hash} (root changed: {root_changed})"));
            }
            visited.insert(hash, rewritten.hash);
        }
        Ok(visited[&start])
    }

    /// repair_root_value repairs the rows of every table of a root value that holds adaptive values, returning the
    /// repaired root value's address and whether it changed.
    fn repair_root_value(&mut self, address: Hash) -> sql::Result<(Hash, bool)> {
        let mut root = Root::decode(&read(self.db(), &address)?)?;
        let mut changed = false;
        for table in tables_for_root(self.db(), &root)? {
            if !table.values_impacted() && !table.keys_impacted() {
                continue;
            }
            let rows = &table.def.table.primary_index;
            let repaired = self
                .rewriter
                .rewrite_map_root(rows, &table.keys, &table.values)
                .map_err(|e| PgError::internal(format!("failed to repair table {}: {}", table.shown(), e.message)))?;
            if repaired == Hash::of(rows) {
                continue;
            }
            let mut message = table.def.table.clone();
            message.primary_index = read(self.db(), &repaired)?;
            let written = message.write(self.db())?;
            root.put_table(self.db(), &table.schema, &table.name, Some(written))?;
            changed = true;
        }
        if !changed {
            return Ok((address, false));
        }
        Ok((self.db().write_value(root.encode())?, true))
    }
}
