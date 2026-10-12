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

//! Every chunk of the store crate's fixtures refers to the addresses that Dolt's WalkAddrs visits, in the same order,
//! as `refs.txt` beside each fixture database records them.

use std::collections::BTreeSet;
use std::path::Path;

use serial::Message;
use serial::walk::walk_addrs;
use store::GenerationalStore;

#[test]
fn messages_refer_to_the_addresses_go_walks() {
    let fixtures = Path::new(env!("CARGO_MANIFEST_DIR")).join("../store/tests/fixtures");
    let mut checked = 0;
    for fixture in std::fs::read_dir(&fixtures).unwrap() {
        let fixture = fixture.unwrap().path();
        if !fixture.is_dir() {
            continue;
        }
        for database in std::fs::read_dir(&fixture).unwrap() {
            let database = database.unwrap().path();
            let Ok(expected) = std::fs::read_to_string(database.join("refs.txt")) else { continue };
            let store = GenerationalStore::open(&database.join(".dolt/noms")).unwrap();
            let mut lines = BTreeSet::new();
            for generation in [&store.new_gen, &store.old_gen] {
                generation
                    .for_each(&mut |chunk| {
                        let mut refs = vec![chunk.hash.to_string()];
                        walk_addrs(Message(&chunk.data), &mut |address| {
                            refs.push(address.to_string());
                            Ok(())
                        })?;
                        lines.insert(refs.join(" "));
                        Ok(())
                    })
                    .unwrap();
            }
            let actual: String = lines.into_iter().map(|line| line + "\n").collect();
            checked += 1;
            assert!(actual == expected, "{} differs from Go:\n{actual}", database.display());
        }
    }
    assert!(checked > 10, "only {checked} fixtures were checked");
}
