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

//! Writers of the version-control messages, which build their fields in the same order as Dolt so that they write the
//! same bytes.

use store::Hash;

use crate::Builder;

/// string_vector writes a vector of strings, creating the strings from last to first as Dolt's
/// SerializeStringVector does.
pub fn string_vector(b: &mut Builder, strings: &[Vec<u8>]) -> u32 {
    let mut offsets = vec![0; strings.len()];
    for i in (0..strings.len()).rev() {
        offsets[i] = b.create_string(&strings[i]);
    }
    b.create_vector_of_tables(&offsets)
}

/// write_store_root writes a StoreRoot around a serialized AddressMap of refs.
pub fn write_store_root(address_map: &[u8]) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let map = b.create_byte_vector(address_map);
    b.start_object(1);
    b.add_offset(0, map);
    let root = b.end_object();
    b.finish_message(root, crate::STORE_ROOT)
}

/// CommitFields is the contents of a Commit message.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CommitFields {
    pub root: Hash,
    pub height: u64,
    pub parents: Vec<Hash>,
    /// The address of the parents' commit closure, all zero for a commit without parents.
    pub parent_closure: Hash,
    pub name: Vec<u8>,
    pub email: Vec<u8>,
    pub description: Vec<u8>,
    pub timestamp_millis: u64,
    pub user_timestamp_millis: i64,
    /// The signature, which is only written when it is not empty.
    pub signature: Vec<u8>,
    /// The committer's name, which is only written when it differs from the author's.
    pub committer_name: Option<Vec<u8>>,
    /// The committer's email, which is only written when it differs from the author's.
    pub committer_email: Option<Vec<u8>>,
}

/// write_commit writes a Commit as Dolt's commit_flatbuffer does.
pub fn write_commit(c: &CommitFields) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let root = b.create_byte_vector(&c.root.0);
    let parents: Vec<u8> = c.parents.iter().flat_map(|p| p.0).collect();
    b.prep(4, parents.len());
    let parents = b.create_byte_vector(&parents);
    let closure = b.create_byte_vector(&c.parent_closure.0);
    let name = b.create_string(&c.name);
    let email = b.create_string(&c.email);
    let description = b.create_string(&c.description);
    let signature = if c.signature.is_empty() { 0 } else { b.create_string(&c.signature) };
    let committer_name = c.committer_name.as_ref().map_or(0, |n| b.create_string(n));
    let committer_email = c.committer_email.as_ref().map_or(0, |e| b.create_string(e));
    b.start_object(12);
    b.add_offset(0, root);
    b.add_u64(1, c.height, 0);
    b.add_offset(2, parents);
    b.add_offset(3, closure);
    b.add_offset(4, name);
    b.add_offset(5, email);
    b.add_offset(6, description);
    b.add_u64(7, c.timestamp_millis, 0);
    b.add_i64(8, c.user_timestamp_millis, 0);
    b.add_offset(9, signature);
    b.add_offset(10, committer_name);
    b.add_offset(11, committer_email);
    let message = b.end_object();
    b.finish_message(message, crate::COMMIT)
}

/// Meta is the author and time of a tag or working set.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Meta {
    pub name: Vec<u8>,
    pub email: Vec<u8>,
    pub description: Vec<u8>,
    pub timestamp_millis: u64,
    pub user_timestamp_millis: i64,
}

/// write_tag writes a Tag as Dolt's tag_flatbuffer does.
pub fn write_tag(commit: Hash, meta: Option<&Meta>) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let address = b.create_byte_vector(&commit.0);
    let strings = meta.map(|m| (b.create_string(&m.name), b.create_string(&m.email), b.create_string(&m.description)));
    b.start_object(6);
    b.add_offset(0, address);
    if let (Some(m), Some((name, email, description))) = (meta, strings) {
        b.add_offset(1, name);
        b.add_offset(2, email);
        b.add_offset(3, description);
        b.add_u64(4, m.timestamp_millis, 0);
        b.add_i64(5, m.user_timestamp_millis, 0);
    }
    let root = b.end_object();
    b.finish_message(root, crate::TAG)
}

/// MergeStateFields is a working set's merge in progress.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct MergeStateFields {
    pub pre_working_root: Hash,
    pub from_commit: Hash,
    pub from_commit_spec: Vec<u8>,
    pub unmergable_tables: Vec<Vec<u8>>,
    pub is_cherry_pick: bool,
    pub is_revert: bool,
    pub pre_merge_head_commit: Option<Hash>,
    /// The hashes of pending revert commits, which are only written when there are some.
    pub pending_commit_hashes: Vec<Vec<u8>>,
}

/// RebaseStateFields is a working set's rebase in progress.
#[derive(Clone, Debug, PartialEq)]
pub struct RebaseStateFields {
    pub pre_working_root: Hash,
    pub onto_commit: Hash,
    pub branch: Vec<u8>,
    pub commit_becomes_empty_handling: u8,
    pub empty_commit_handling: u8,
    pub last_attempted_step: f32,
    pub rebasing_started: bool,
    pub skip_verification: bool,
}

/// WorkingSetFields is the contents of a WorkingSet message.
#[derive(Clone, Debug, PartialEq)]
pub struct WorkingSetFields {
    pub working_root: Hash,
    pub staged_root: Option<Hash>,
    pub merge_state: Option<MergeStateFields>,
    pub rebase_state: Option<RebaseStateFields>,
    /// The author, which Dolt writes without the user timestamp.
    pub meta: Option<Meta>,
}

/// write_working_set writes a WorkingSet as Dolt's workingset_flatbuffer does.
pub fn write_working_set(w: &WorkingSetFields) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let working = b.create_byte_vector(&w.working_root.0);
    let staged = w.staged_root.map_or(0, |s| b.create_byte_vector(&s.0));
    let mut merge_state = 0;
    if let Some(m) = &w.merge_state {
        let pre_working = b.create_byte_vector(&m.pre_working_root.0);
        let from = b.create_byte_vector(&m.from_commit.0);
        let spec = b.create_string(&m.from_commit_spec);
        let unmergable = string_vector(&mut b, &m.unmergable_tables);
        let head = m.pre_merge_head_commit.map_or(0, |h| b.create_byte_vector(&h.0));
        let pending =
            if m.pending_commit_hashes.is_empty() { 0 } else { string_vector(&mut b, &m.pending_commit_hashes) };
        b.start_object(8);
        b.add_offset(0, pre_working);
        b.add_offset(1, from);
        b.add_offset(2, spec);
        b.add_offset(3, unmergable);
        b.add_bool(4, m.is_cherry_pick, false);
        b.add_bool(5, m.is_revert, false);
        b.add_offset(6, head);
        b.add_offset(7, pending);
        merge_state = b.end_object();
    }
    let mut rebase_state = 0;
    if let Some(r) = &w.rebase_state {
        let pre_working = b.create_byte_vector(&r.pre_working_root.0);
        let onto = b.create_byte_vector(&r.onto_commit.0);
        let branch = b.create_string(&r.branch);
        b.start_object(8);
        b.add_offset(0, pre_working);
        b.add_offset(1, branch);
        b.add_offset(2, onto);
        b.add_u8(4, r.commit_becomes_empty_handling, 0);
        b.add_u8(3, r.empty_commit_handling, 0);
        b.add_f32(5, r.last_attempted_step, 0.0);
        b.add_bool(6, r.rebasing_started, false);
        b.add_bool(7, r.skip_verification, false);
        rebase_state = b.end_object();
    }
    let strings =
        w.meta.as_ref().map(|m| (b.create_string(&m.name), b.create_string(&m.email), b.create_string(&m.description)));
    b.start_object(8);
    b.add_offset(0, working);
    b.add_offset(1, staged);
    b.add_offset(6, merge_state);
    b.add_offset(7, rebase_state);
    if let (Some(m), Some((name, email, description))) = (&w.meta, strings) {
        b.add_offset(2, name);
        b.add_offset(3, email);
        b.add_offset(4, description);
        b.add_u64(5, m.timestamp_millis, 0);
    }
    let root = b.end_object();
    b.finish_message(root, crate::WORKING_SET)
}

/// StashFields is the contents of a Stash message.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct StashFields {
    pub root: Hash,
    pub head_commit: Hash,
    pub branch_name: Vec<u8>,
    pub description: Vec<u8>,
    pub tables_to_stage: Option<Vec<Vec<u8>>>,
}

/// write_stash writes a Stash as Dolt's stash_flatbuffer does.
pub fn write_stash(s: &StashFields) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let root = b.create_byte_vector(&s.root.0);
    let head = b.create_byte_vector(&s.head_commit.0);
    let branch = b.create_string(&s.branch_name);
    let description = b.create_string(&s.description);
    let tables = s.tables_to_stage.as_ref().map_or(0, |t| string_vector(&mut b, t));
    b.start_object(5);
    b.add_offset(0, root);
    b.add_offset(1, head);
    b.add_offset(2, branch);
    b.add_offset(3, description);
    b.add_offset(4, tables);
    let message = b.end_object();
    b.finish_message(message, crate::STASH)
}

/// write_stash_list writes a StashList around a serialized AddressMap of stashes.
pub fn write_stash_list(address_map: &[u8]) -> Vec<u8> {
    let mut b = Builder::new(1024);
    let map = b.create_byte_vector(address_map);
    b.start_object(1);
    b.add_offset(0, map);
    let root = b.end_object();
    b.finish_message(root, crate::STASH_LIST)
}
