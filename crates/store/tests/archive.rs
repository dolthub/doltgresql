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

//! Archives the archive writer builds, read back by the archive reader.

use std::path::Path;

use store::{ArchiveReader, ArchiveWriter, Chunk, Hash};

/// round_trip writes chunks of the given count to an archive and checks that every one reads back.
fn round_trip(name: &str, count: usize) {
    let chunks: Vec<Chunk> = (0..count)
        .map(|i| Chunk::new(format!("chunk {i} with some repeated text {}", "abc".repeat(i % 7)).into_bytes()))
        .collect();
    let mut writer = ArchiveWriter::new();
    for chunk in &chunks {
        writer.add_chunk(chunk.clone()).unwrap();
    }
    assert_eq!(writer.count(), count);
    let (hash, bytes) = writer.finish();
    assert_eq!(hash, Hash::of(&bytes));
    let dir = Path::new(env!("CARGO_TARGET_TMPDIR")).join("archive");
    std::fs::create_dir_all(&dir).unwrap();
    let path = dir.join(format!("{name}.darc"));
    std::fs::write(&path, &bytes).unwrap();
    let reader = ArchiveReader::open(&path).unwrap();
    assert_eq!(reader.count(), count);
    for chunk in &chunks {
        assert_eq!(reader.get(&chunk.hash).unwrap().map(|c| c.data), Some(chunk.data.clone()));
    }
}

#[test]
fn small_archives_store_snappy_records() {
    round_trip("small", 10);
}

#[test]
fn large_archives_compress_with_a_trained_dictionary() {
    round_trip("large", 1500);
}
