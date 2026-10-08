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

//! Stores on each blobstore: writing, reopening, and losing a commit race. The cloud stores run against the URL that
//! an environment variable names, such as DOLTGRES_TEST_S3_URL, with that service's credentials in the environment.

use std::collections::BTreeMap;

use store::{Chunk, ChunkStore, Hash};

/// unique returns a name that no earlier run used.
fn unique() -> String {
    let nanos = std::time::SystemTime::now().duration_since(std::time::UNIX_EPOCH).unwrap().as_nanos();
    format!("test{}x{nanos}", std::process::id())
}

/// open opens the store at a URL.
fn open(url: &str) -> Box<dyn ChunkStore> {
    blobstores::open(url, &BTreeMap::new()).unwrap().expect("a blobstore scheme")
}

/// check writes chunks to the store at a URL, reads them back from a fresh store, and checks that a store whose view
/// of the root is stale loses a commit race, unless the URL is a mem:// one, whose every store starts empty, or an
/// oss:// one, whose manifest writes have no condition, as in Dolt.
fn check(url: &str) {
    let chunks: Vec<Chunk> = (0..200).map(|i| Chunk::new(format!("chunk {i} ").repeat(40).into_bytes())).collect();
    let mut store = open(url);
    assert_eq!(store.root(), Hash::default());
    for chunk in &chunks[..100] {
        store.put(chunk.clone(), Vec::new()).unwrap();
    }
    assert!(store.commit(chunks[0].hash, Hash::default()).unwrap());
    if url.starts_with("mem://") {
        assert!(store.get(&chunks[99].hash).unwrap().is_some());
        return;
    }
    let mut stale = open(url);
    let fresh = open(url);
    assert_eq!(fresh.root(), chunks[0].hash);
    for chunk in &chunks[..100] {
        assert_eq!(fresh.get(&chunk.hash).unwrap().unwrap().data, chunk.data);
    }
    for chunk in &chunks[100..] {
        store.put(chunk.clone(), Vec::new()).unwrap();
        stale.put(chunk.clone(), Vec::new()).unwrap();
    }
    assert!(store.commit(chunks[100].hash, chunks[0].hash).unwrap());
    if url.starts_with("oss://") {
        return;
    }
    assert!(!stale.commit(chunks[150].hash, chunks[0].hash).unwrap());
    assert_eq!(stale.root(), chunks[100].hash);
    let last = open(url);
    assert_eq!(last.root(), chunks[100].hash);
    assert_eq!(last.get(&chunks[199].hash).unwrap().unwrap().data, chunks[199].data);
}

/// check_env checks the store under the URL that an environment variable names, when it is set.
fn check_env(variable: &str) {
    if let Ok(url) = std::env::var(variable) {
        check(&format!("{}/{}", url.trim_end_matches('/'), unique()));
    }
}

#[test]
fn test_localbs() {
    let dir = std::env::temp_dir().join(unique());
    check(&format!("localbs://{}", dir.display()));
    let _ = std::fs::remove_dir_all(dir);
}

#[test]
fn test_mem() {
    check("mem://test");
}

#[test]
fn test_s3() {
    if let Ok(url) = std::env::var("DOLTGRES_TEST_S3_URL") {
        let (location, query) = url.split_once('?').unwrap_or((&url, ""));
        check(&format!("{}/{}?{query}", location.trim_end_matches('/'), unique()));
    }
}

#[test]
fn test_aws() {
    check_env("DOLTGRES_TEST_AWS_URL");
}

#[test]
fn test_gcs() {
    check_env("DOLTGRES_TEST_GS_URL");
}

#[test]
fn test_azure() {
    check_env("DOLTGRES_TEST_AZ_URL");
}

#[test]
fn test_oci() {
    check_env("DOLTGRES_TEST_OCI_URL");
}

#[test]
fn test_oss() {
    check_env("DOLTGRES_TEST_OSS_URL");
}
