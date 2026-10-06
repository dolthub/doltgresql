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

use std::collections::{HashSet, VecDeque};

use prolly::{Node, walk_leaves};
use serial::{
    Commit, DoltgresRootValue, MergeState, Message, RebaseState, Stash, StashList, StoreRoot, TableMessage, Tag,
    WorkingSet,
};
use sha2::{Digest, Sha256};
use store::{ChunkReader, Hash, Result};

/// hex renders bytes as lower-case hex.
pub(crate) fn hex(bytes: &[u8]) -> String {
    bytes.iter().map(|b| format!("{b:02x}")).collect()
}

/// hex_list renders byte strings as comma-separated hex.
fn hex_list(values: &[&[u8]]) -> String {
    values.iter().map(|v| hex(v)).collect::<Vec<_>>().join(",")
}

/// address_map returns the entries of a serialized AddressMap in key order.
pub fn address_map(reader: &dyn ChunkReader, bytes: &[u8]) -> Result<Vec<(Vec<u8>, Hash)>> {
    let mut entries = Vec::new();
    walk_leaves(reader, &Node::decode(bytes.to_vec())?, &mut |key, value| {
        entries.push((key.to_vec(), serial::hash(value)?));
        Ok(())
    })?;
    Ok(entries)
}

/// Graph walks the objects reachable from a store's root.
struct Graph<'a> {
    reader: &'a dyn ChunkReader,
    lines: Vec<String>,
    seen: HashSet<Hash>,
    seen_objects: HashSet<Hash>,
    seen_extra: HashSet<Hash>,
    queue: VecDeque<Hash>,
}

impl Graph<'_> {
    fn visit(&mut self, hash: Hash) {
        if !hash.is_empty() && self.seen.insert(hash) {
            self.queue.push_back(hash);
        }
    }

    /// describe_tree renders the digest of the raw items of the tree at the address once.
    fn describe_tree(&mut self, kind: &str, address: Hash) -> Result<()> {
        if address.is_empty() || !self.seen_extra.insert(address) {
            return Ok(());
        }
        let mut hasher = Sha256::new();
        let mut count = 0;
        walk_leaves(self.reader, &Node::load(self.reader, &address)?, &mut |key, value| {
            hasher.update(format!("{} {}\n", hex(key), hex(value)).as_bytes());
            count += 1;
            Ok(())
        })?;
        self.lines.push(format!("tree {address} {kind} count={count} digest={}", hex(&hasher.finalize())));
        Ok(())
    }

    /// describe_foreign_keys renders a foreign key collection once.
    fn describe_foreign_keys(&mut self, address: Hash) -> Result<()> {
        if address.is_empty() || !self.seen_extra.insert(address) {
            return Ok(());
        }
        let chunk = self.reader.require(&address)?;
        let numbers = |values: &[u64]| values.iter().map(u64::to_string).collect::<Vec<_>>().join(",");
        for (i, fk) in serial::foreign_keys(Message(&chunk.data))?.iter().enumerate() {
            self.lines.push(format!(
                "foreignkey {address} {i} name={} child={} childindex={} childcols={} parent={} parentindex={} \
                 parentcols={} onupdate={} ondelete={} unresolvedchild={} unresolvedparent={} childschema={} \
                 parentschema={} notvalid={} match={}",
                hex(fk.name),
                hex(fk.child_table_name),
                hex(fk.child_table_index),
                numbers(&fk.child_table_columns),
                hex(fk.parent_table_name),
                hex(fk.parent_table_index),
                numbers(&fk.parent_table_columns),
                fk.on_update,
                fk.on_delete,
                hex_list(&fk.unresolved_child_columns),
                hex_list(&fk.unresolved_parent_columns),
                hex_list(&fk.child_table_database_schema),
                hex_list(&fk.parent_table_database_schema),
                fk.is_not_valid,
                fk.match_type,
            ));
        }
        Ok(())
    }

    /// describe_merge_state renders a working set's merge and rebase state.
    fn describe_merge_state(&mut self, address: Hash, working_set: &WorkingSet<'_>) -> Result<()> {
        if let Some(table) = working_set.merge_state()? {
            let state = MergeState(table);
            let (pre_working, from) = (serial::hash(state.pre_working_root()?)?, serial::hash(state.from_commit()?)?);
            self.lines.push(format!(
                "mergestate {address} preworking={pre_working} from={from} spec={} unmergable={} cherrypick={} \
                 revert={} premergehead={} pending={}",
                hex(state.from_commit_spec()?),
                hex_list(&state.unmergable_tables()?),
                state.is_cherry_pick()?,
                state.is_revert()?,
                hex(state.pre_merge_head_commit()?),
                hex_list(&state.pending_commit_hashes()?),
            ));
            self.visit(pre_working);
            self.visit(from);
        }
        if let Some(table) = working_set.rebase_state()? {
            let state = RebaseState(table);
            let (pre_working, onto) = (serial::hash(state.pre_working_root()?)?, serial::hash(state.onto_commit()?)?);
            self.lines.push(format!(
                "rebasestate {address} preworking={pre_working} branch={} onto={onto} empty={} becomesempty={} \
                 step={} started={} skip={}",
                hex(state.branch()?),
                state.empty_commit_handling()?,
                state.commit_becomes_empty_handling()?,
                objects::go_float32(state.last_attempted_step()?),
                state.rebasing_started()?,
                state.skip_verification()?,
            ));
            self.visit(pre_working);
            self.visit(onto);
        }
        Ok(())
    }

    /// describe_stash_list renders a stash list and its stashes.
    fn describe_stash_list(&mut self, address: Hash, message: Message<'_>) -> Result<()> {
        let list = StashList::new(message)?;
        for (name, entry) in address_map(self.reader, list.address_map()?)? {
            let chunk = self.reader.require(&entry)?;
            let stash = Stash::new(Message(&chunk.data))?;
            let (root, head) = (serial::hash(stash.root()?)?, serial::hash(stash.head_commit()?)?);
            self.lines.push(format!(
                "stash {address} {} {entry} root={root} head={head} branch={} desc={} stage={}",
                hex(&name),
                hex(stash.branch_name()?),
                hex(stash.description()?),
                hex_list(&stash.tables_to_stage()?),
            ));
            self.visit(root);
            self.visit(head);
        }
        Ok(())
    }

    fn describe(&mut self, address: Hash) -> Result<()> {
        let chunk = self.reader.require(&address)?;
        let message = Message(&chunk.data);
        match message.file_id() {
            serial::COMMIT => {
                let commit = Commit::new(message)?;
                let parents = commit.parents()?;
                for parent in &parents {
                    self.visit(*parent);
                }
                let closure = commit.parent_closure_bytes()?.map(serial::hash).transpose()?;
                if let Some(closure) = closure {
                    self.describe_tree("closure", closure)?;
                }
                self.lines.push(format!(
                    "commit {address} root={} height={} parents={} closure={} name={} email={} desc={} ts={} uts={} \
                     sig={} cname={} cemail={}",
                    commit.root()?,
                    commit.height()?,
                    parents.iter().map(Hash::to_string).collect::<Vec<_>>().join(","),
                    closure.map(|h| h.to_string()).unwrap_or_default(),
                    hex(commit.name()?),
                    hex(commit.email()?),
                    hex(commit.description()?),
                    commit.timestamp_millis()?,
                    commit.user_timestamp_millis()?,
                    hex(commit.signature()?.unwrap_or_default()),
                    hex(commit.committer_name()?.unwrap_or_default()),
                    hex(commit.committer_email()?.unwrap_or_default()),
                ));
                self.visit(commit.root()?);
            }
            serial::TAG => {
                let tag = Tag::new(message)?;
                self.lines.push(format!(
                    "tag {address} commit={} name={} email={} desc={} ts={} uts={}",
                    tag.commit()?,
                    hex(tag.name()?),
                    hex(tag.email()?),
                    hex(tag.description()?),
                    tag.timestamp_millis()?,
                    tag.user_timestamp_millis()?,
                ));
                self.visit(tag.commit()?);
            }
            serial::WORKING_SET => {
                let working_set = WorkingSet::new(message)?;
                let staged = working_set.staged_root()?;
                if let Some(staged) = staged {
                    self.visit(staged);
                }
                self.lines.push(format!(
                    "workingset {address} working={} staged={} name={} email={} desc={} ts={} merge={} rebase={}",
                    working_set.working_root()?,
                    staged.map(|h| h.to_string()).unwrap_or_default(),
                    hex(working_set.name()?),
                    hex(working_set.email()?),
                    hex(working_set.description()?),
                    working_set.timestamp_millis()?,
                    working_set.merge_state()?.is_some(),
                    working_set.rebase_state()?.is_some(),
                ));
                self.visit(working_set.working_root()?);
                self.describe_merge_state(address, &working_set)?;
            }
            serial::DOLTGRES_ROOT_VALUE => {
                let root = DoltgresRootValue::new(message)?;
                self.lines.push(format!(
                    "rootvalue {address} fv={} collation={} fk={} schemas={}",
                    root.feature_version()?,
                    root.collation()?,
                    hex(root.foreign_keys()?.unwrap_or_default()),
                    root.schemas()?.iter().map(|s| hex(s)).collect::<Vec<_>>().join(","),
                ));
                if let Some(fk) = root.foreign_keys()?.filter(|fk| fk.len() == Hash::LEN) {
                    self.describe_foreign_keys(serial::hash(fk)?)?;
                }
                let mut maps = vec![("tables", root.tables()?)];
                maps.extend(root.root_object_maps()?);
                for (name, bytes) in maps {
                    let Some(mut bytes) = bytes.filter(|b| !b.is_empty()).map(<[u8]>::to_vec) else { continue };
                    if name != "tables" {
                        if bytes.len() != Hash::LEN || bytes.iter().all(|&b| b == 0) {
                            continue;
                        }
                        let map = serial::hash(&bytes)?;
                        self.lines.push(format!("rootvalue-map {address} {name} {map}"));
                        bytes = self.reader.require(&map)?.data;
                    }
                    for (key, entry) in address_map(self.reader, &bytes)? {
                        self.lines.push(format!("rootvalue-entry {address} {name} {} {entry}", hex(&key)));
                        if name == "tables" {
                            self.visit(entry);
                        } else if self.seen_objects.insert(entry) {
                            let kind = objects::Kind::from_field(name).unwrap();
                            let object =
                                objects::RootObject::deserialize(kind, &prolly::read_blob(self.reader, &entry)?)?;
                            self.lines.push(format!("object {entry} {name} {}", object.show()));
                        }
                    }
                }
            }
            serial::TABLE => {
                let table = TableMessage::new(message)?;
                let primary = table.primary_index()?;
                let node = Node::decode(primary.to_vec())?;
                self.lines.push(format!(
                    "table {address} schema={} autoinc={} primary={} level={} count={} conflicts={} violations={} \
                     artifacts={}",
                    table.schema()?,
                    table.auto_increment()?,
                    Hash::of(primary),
                    node.level(),
                    node.tree_count(),
                    table.conflicts()?.is_some(),
                    hex(table.violations()?.unwrap_or_default()),
                    hex(table.artifacts()?.unwrap_or_default()),
                ));
                let mut secondary = Vec::new();
                if let Some(bytes) = table.secondary_indexes()?.filter(|b| !b.is_empty()) {
                    secondary = address_map(self.reader, bytes)?;
                    for (key, entry) in &secondary {
                        self.lines.push(format!("table-index {address} {} {entry}", hex(key)));
                    }
                }
                crate::rows::describe_table(
                    self.reader,
                    address,
                    table.schema()?,
                    primary,
                    &secondary,
                    &mut self.lines,
                )?;
                if let Some(artifacts) = table.artifacts()?.filter(|a| a.len() == Hash::LEN) {
                    self.describe_tree("artifacts", serial::hash(artifacts)?)?;
                }
                if let Some(violations) = table.violations()?.filter(|v| v.len() == Hash::LEN) {
                    self.describe_tree("violations", serial::hash(violations)?)?;
                }
            }
            serial::STASH_LIST => self.describe_stash_list(address, message)?,
            other => self.lines.push(format!("other {address} {other}")),
        }
        Ok(())
    }
}

/// dump_graph renders the refs and every object reachable from them, one sorted line each, as the Go graph oracle
/// does.
pub fn dump_graph(reader: &dyn ChunkReader, root: Hash) -> Result<String> {
    let mut graph = Graph {
        reader,
        lines: vec![format!("root {root}")],
        seen: HashSet::new(),
        seen_objects: HashSet::new(),
        seen_extra: HashSet::new(),
        queue: VecDeque::new(),
    };
    let chunk = reader.require(&root)?;
    if let Some(bytes) = StoreRoot::new(Message(&chunk.data))?.address_map()?.filter(|b| !b.is_empty()) {
        for (name, address) in address_map(reader, bytes)? {
            graph.lines.push(format!("ref {} {address}", hex(&name)));
            graph.visit(address);
        }
    }
    while let Some(address) = graph.queue.pop_front() {
        graph.describe(address)?;
    }
    graph.lines.sort();
    Ok(graph.lines.join("\n") + "\n")
}
