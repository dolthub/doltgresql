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

//! Every version-control message in the store crate's fixtures, which Go wrote, writes back to the same bytes.

use std::collections::BTreeMap;
use std::path::Path;

use serial::write::{
    CommitFields, MergeStateFields, Meta, RebaseStateFields, StashFields, WorkingSetFields, write_commit, write_stash,
    write_stash_list, write_store_root, write_tag, write_working_set,
};
use serial::{Commit, MergeState, Message, RebaseState, Stash, StashList, StoreRoot, Tag, WorkingSet};
use store::{Chunk, GenerationalStore};

/// owned copies a list of byte strings.
fn owned(values: Vec<&[u8]>) -> Vec<Vec<u8>> {
    values.into_iter().map(<[u8]>::to_vec).collect()
}

/// rewrite writes a version-control message again, returning None for other messages.
fn rewrite(chunk: &Chunk) -> Option<Vec<u8>> {
    let message = Message(&chunk.data);
    Some(match message.file_id() {
        serial::STORE_ROOT => {
            write_store_root(StoreRoot::new(message).unwrap().address_map().unwrap().unwrap_or_default())
        }
        serial::COMMIT => {
            let c = Commit::new(message).unwrap();
            let closure = c.parent_closure_bytes().unwrap().map(|b| serial::hash(b).unwrap()).unwrap_or_default();
            write_commit(&CommitFields {
                root: c.root().unwrap(),
                height: c.height().unwrap(),
                parents: c.parents().unwrap(),
                parent_closure: closure,
                name: c.name().unwrap().to_vec(),
                email: c.email().unwrap().to_vec(),
                description: c.description().unwrap().to_vec(),
                timestamp_millis: c.timestamp_millis().unwrap(),
                user_timestamp_millis: c.user_timestamp_millis().unwrap(),
                signature: c.signature().unwrap().unwrap_or_default().to_vec(),
                committer_name: c.committer_name().unwrap().map(<[u8]>::to_vec),
                committer_email: c.committer_email().unwrap().map(<[u8]>::to_vec),
            })
        }
        serial::TAG => {
            let t = Tag::new(message).unwrap();
            let meta = Meta {
                name: t.name().unwrap().to_vec(),
                email: t.email().unwrap().to_vec(),
                description: t.description().unwrap().to_vec(),
                timestamp_millis: t.timestamp_millis().unwrap(),
                user_timestamp_millis: t.user_timestamp_millis().unwrap(),
            };
            write_tag(t.commit().unwrap(), Some(&meta))
        }
        serial::WORKING_SET => {
            let w = WorkingSet::new(message).unwrap();
            let merge_state = w.merge_state().unwrap().map(|t| {
                let m = MergeState(t);
                MergeStateFields {
                    pre_working_root: serial::hash(m.pre_working_root().unwrap()).unwrap(),
                    from_commit: serial::hash(m.from_commit().unwrap()).unwrap(),
                    from_commit_spec: m.from_commit_spec().unwrap().to_vec(),
                    unmergable_tables: owned(m.unmergable_tables().unwrap()),
                    is_cherry_pick: m.is_cherry_pick().unwrap(),
                    is_revert: m.is_revert().unwrap(),
                    pre_merge_head_commit: Some(m.pre_merge_head_commit().unwrap())
                        .filter(|h| !h.is_empty())
                        .map(|h| serial::hash(h).unwrap()),
                    pending_commit_hashes: owned(m.pending_commit_hashes().unwrap()),
                }
            });
            let rebase_state = w.rebase_state().unwrap().map(|t| {
                let r = RebaseState(t);
                RebaseStateFields {
                    pre_working_root: serial::hash(r.pre_working_root().unwrap()).unwrap(),
                    onto_commit: serial::hash(r.onto_commit().unwrap()).unwrap(),
                    branch: r.branch().unwrap().to_vec(),
                    commit_becomes_empty_handling: r.commit_becomes_empty_handling().unwrap(),
                    empty_commit_handling: r.empty_commit_handling().unwrap(),
                    last_attempted_step: r.last_attempted_step().unwrap(),
                    rebasing_started: r.rebasing_started().unwrap(),
                    skip_verification: r.skip_verification().unwrap(),
                }
            });
            write_working_set(&WorkingSetFields {
                working_root: w.working_root().unwrap(),
                staged_root: w.staged_root().unwrap(),
                merge_state,
                rebase_state,
                meta: Some(Meta {
                    name: w.name().unwrap().to_vec(),
                    email: w.email().unwrap().to_vec(),
                    description: w.description().unwrap().to_vec(),
                    timestamp_millis: w.timestamp_millis().unwrap(),
                    user_timestamp_millis: 0,
                }),
            })
        }
        serial::STASH => {
            let s = Stash::new(message).unwrap();
            let tables = s.0.vector(4, 4).unwrap().map(|_| owned(s.tables_to_stage().unwrap()));
            write_stash(&StashFields {
                root: serial::hash(s.root().unwrap()).unwrap(),
                head_commit: serial::hash(s.head_commit().unwrap()).unwrap(),
                branch_name: s.branch_name().unwrap().to_vec(),
                description: s.description().unwrap().to_vec(),
                tables_to_stage: tables,
            })
        }
        serial::STASH_LIST => write_stash_list(StashList::new(message).unwrap().address_map().unwrap()),
        _ => return None,
    })
}

#[test]
fn version_control_messages_write_the_bytes_go_wrote() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("../store/tests/fixtures");
    let mut checked: BTreeMap<String, usize> = BTreeMap::new();
    let mut failures = Vec::new();
    for fixture in std::fs::read_dir(&fixtures).unwrap() {
        let fixture = fixture.unwrap().path();
        if !fixture.is_dir() {
            continue;
        }
        for database in std::fs::read_dir(&fixture).unwrap() {
            let noms = database.unwrap().path().join(".dolt/noms");
            if !noms.is_dir() {
                continue;
            }
            let store = GenerationalStore::open(&noms).unwrap();
            for generation in [&store.new_gen, &store.old_gen] {
                generation
                    .for_each(&mut |chunk| {
                        if let Some(bytes) = rewrite(&chunk) {
                            let file_id = Message(&chunk.data).file_id().to_string();
                            *checked.entry(file_id.clone()).or_default() += 1;
                            if bytes != chunk.data {
                                failures.push(format!("{}: {} {file_id}", noms.display(), chunk.hash));
                            }
                        }
                        Ok(())
                    })
                    .unwrap();
            }
        }
    }
    for kind in
        [serial::STORE_ROOT, serial::COMMIT, serial::TAG, serial::WORKING_SET, serial::STASH, serial::STASH_LIST]
    {
        assert!(checked.contains_key(kind), "no {kind} messages were checked: {checked:?}");
    }
    assert!(failures.is_empty(), "{} differ ({checked:?}):\n{}", failures.len(), failures.join("\n"));
}
