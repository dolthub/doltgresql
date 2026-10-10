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

#![forbid(unsafe_code)]

//! Prints the refs and every object reachable from them in a database's noms directory, as the Go graph oracle does.

use std::path::Path;

fn main() {
    let args: Vec<String> = std::env::args().collect();
    if args.len() != 2 {
        eprintln!("usage: graphdump <noms dir>");
        std::process::exit(2);
    }
    let result =
        store::GenerationalStore::open(Path::new(&args[1])).and_then(|store| doltdb::dump_graph(&store, store.root()));
    match result {
        Ok(text) => print!("{text}"),
        Err(err) => {
            eprintln!("{err}");
            std::process::exit(1);
        }
    }
}
