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

//! The client side of the remotes API: a chunk store over a remote that another server serves, as Dolt's
//! DoltChunkStore is, which reads chunks by downloading the byte ranges the server names and writes them by uploading
//! a table file and committing it with the new root.

use std::collections::HashMap;
use std::sync::Mutex;

use md5::{Digest, Md5};
use store::{Chunk, ChunkReader, ChunkStore, Hash, TableWriter};
use tonic::transport::Channel;

use crate::remotesapi::chunk_store_service_client::ChunkStoreServiceClient;
use crate::remotesapi::{self as api, download_loc, upload_loc};

/// BATCH is how many chunks one download location request asks for.
const BATCH: usize = 4096;

/// RemoteStore is a chunk store on a remote server.
pub struct RemoteStore {
    runtime: tokio::runtime::Runtime,
    client: ChunkStoreServiceClient<Channel>,
    http: reqwest::Client,
    /// The repository path that requests name.
    repo: String,
    root: Hash,
    /// The chunks put since the last commit, which the commit uploads.
    pending: Vec<Chunk>,
    /// The chunks downloaded so far.
    cache: Mutex<HashMap<Hash, Chunk>>,
    /// The decompressed archive dictionaries downloaded so far, by URL and offset.
    dictionaries: Mutex<HashMap<(String, u64), Vec<u8>>>,
}

/// error returns a store error for a failed remote call.
fn error(err: impl std::fmt::Display) -> store::Error {
    store::Error::Io(std::io::Error::other(err.to_string()))
}

impl RemoteStore {
    /// open connects to a remote at a URL such as `http://host:port/repo` and reads its root.
    pub fn open(url: &str) -> Result<RemoteStore, String> {
        let (scheme, rest) = url.split_once("://").ok_or_else(|| format!("invalid remote url: {url}"))?;
        let (host, repo) = rest.split_once('/').unwrap_or((rest, ""));
        let runtime = tokio::runtime::Builder::new_current_thread().enable_all().build().map_err(|e| e.to_string())?;
        let endpoint = format!("{scheme}://{host}");
        let client = runtime
            .block_on(ChunkStoreServiceClient::connect(endpoint))
            .map_err(|err| err.to_string())?
            .max_decoding_message_size(usize::MAX)
            .max_encoding_message_size(usize::MAX);
        let mut store = RemoteStore {
            runtime,
            client,
            http: reqwest::Client::new(),
            repo: repo.trim_end_matches('/').to_string(),
            root: Hash::default(),
            pending: Vec::new(),
            cache: Mutex::new(HashMap::new()),
            dictionaries: Mutex::new(HashMap::new()),
        };
        store.metadata().map_err(|err| err.to_string())?;
        store.root = store.fetch_root().map_err(|err| err.to_string())?;
        Ok(store)
    }

    /// metadata checks that the remote serves the repository in Dolt's storage format.
    fn metadata(&mut self) -> store::Result<()> {
        let request = api::GetRepoMetadataRequest {
            repo_path: self.repo.clone(),
            client_repo_format: Some(api::ClientRepoFormat {
                nbf_version: "__DOLT__".to_string(),
                nbs_version: "5".to_string(),
            }),
            ..Default::default()
        };
        let mut client = self.client.clone();
        let response = self.runtime.block_on(client.get_repo_metadata(request)).map_err(error)?.into_inner();
        if response.nbf_version != "__DOLT__" {
            return Err(error(format!("remote has unsupported storage format {}", response.nbf_version)));
        }
        Ok(())
    }

    /// fetch_root reads the remote's root.
    fn fetch_root(&self) -> store::Result<Hash> {
        let request = api::RootRequest { repo_path: self.repo.clone(), ..Default::default() };
        let mut client = self.client.clone();
        let response = self.runtime.block_on(client.root(request)).map_err(error)?.into_inner();
        Ok(Hash(response.root_hash.as_slice().try_into().map_err(|_| error("invalid root hash"))?))
    }

    /// download fetches chunks the cache lacks by the ranges the server names for them.
    fn download(&self, hashes: &[Hash]) -> store::Result<()> {
        for batch in hashes.chunks(BATCH) {
            let request = api::GetDownloadLocsRequest {
                repo_path: self.repo.clone(),
                chunk_hashes: batch.iter().map(|h| h.0.to_vec()).collect(),
                ..Default::default()
            };
            let mut client = self.client.clone();
            let response = self.runtime.block_on(client.get_download_locations(request)).map_err(error)?.into_inner();
            for location in response.locs {
                let Some(download_loc::Location::HttpGetRange(range)) = location.location else { continue };
                self.fetch_ranges(&range)?;
            }
        }
        Ok(())
    }

    /// get_range downloads a byte range of a URL.
    fn get_range(&self, url: &str, offset: u64, length: u64) -> store::Result<Vec<u8>> {
        let request =
            self.http.get(url).header("range", format!("bytes={}-{}", offset, offset + length.max(1) - 1)).send();
        let response = self.runtime.block_on(request).map_err(error)?;
        if !response.status().is_success() {
            return Err(error(format!("download of {url} failed with {}", response.status())));
        }
        Ok(self.runtime.block_on(response.bytes()).map_err(error)?.to_vec())
    }

    /// fetch_ranges downloads the span that covers a file's ranges and decodes each chunk in it.
    fn fetch_ranges(&self, range: &api::HttpGetRange) -> store::Result<()> {
        let Some(start) = range.ranges.iter().map(|r| r.offset).min() else { return Ok(()) };
        let end = range.ranges.iter().map(|r| r.offset + r.length as u64).max().unwrap_or(start);
        let span = self.get_range(&range.url, start, end - start)?;
        for chunk in &range.ranges {
            let hash = Hash(chunk.hash.as_slice().try_into().map_err(|_| error("invalid chunk hash"))?);
            let at = (chunk.offset - start) as usize;
            let bytes = span.get(at..at + chunk.length as usize).ok_or_else(|| error("short download"))?;
            let chunk = match chunk.dictionary_length {
                0 => Chunk::from_record(hash, bytes)?,
                _ => {
                    let dictionary = self.dictionary(&range.url, chunk.dictionary_offset, chunk.dictionary_length)?;
                    let data = zstd::bulk::Decompressor::with_dictionary(&dictionary)
                        .and_then(|mut d| d.decompress(bytes, 1 << 26))
                        .map_err(error)?;
                    Chunk { hash, data }
                }
            };
            self.cache.lock().map_err(|_| error("a cache lock was poisoned"))?.insert(hash, chunk);
        }
        Ok(())
    }

    /// dictionary returns the decompressed archive dictionary at an offset of a URL, downloading it once.
    fn dictionary(&self, url: &str, offset: u64, length: u32) -> store::Result<Vec<u8>> {
        let path = url.split('?').next().unwrap_or(url).to_string();
        if let Some(dictionary) =
            self.dictionaries.lock().map_err(|_| error("a lock was poisoned"))?.get(&(path.clone(), offset))
        {
            return Ok(dictionary.clone());
        }
        let compressed = self.get_range(url, offset, length as u64)?;
        let dictionary = zstd::stream::decode_all(compressed.as_slice()).map_err(error)?;
        self.dictionaries.lock().map_err(|_| error("a lock was poisoned"))?.insert((path, offset), dictionary.clone());
        Ok(dictionary)
    }

    /// upload writes the pending chunks to a table file and uploads it, returning its name and chunk count.
    fn upload(&mut self) -> store::Result<Option<api::ChunkTableInfo>> {
        if self.pending.is_empty() {
            return Ok(None);
        }
        let mut writer = TableWriter::new();
        for chunk in &self.pending {
            writer.add_chunk(chunk);
        }
        let count = writer.count() as u32;
        let (name, bytes) = writer.finish();
        let details = api::TableFileDetails {
            id: name.0.to_vec(),
            content_length: bytes.len() as u64,
            content_hash: Md5::digest(&bytes).to_vec(),
            num_chunks: count as u64,
            suffix: String::new(),
            split_offset: 0,
        };
        let request = api::GetUploadLocsRequest {
            repo_path: self.repo.clone(),
            table_file_details: vec![details],
            ..Default::default()
        };
        let mut client = self.client.clone();
        let response = self.runtime.block_on(client.get_upload_locations(request)).map_err(error)?.into_inner();
        for location in response.locs {
            let Some(upload_loc::Location::HttpPost(post)) = location.location else { continue };
            let sent = self.runtime.block_on(self.http.put(&post.url).body(bytes.clone()).send()).map_err(error)?;
            if !sent.status().is_success() {
                return Err(error(format!("upload of table file {name} failed with {}", sent.status())));
            }
        }
        Ok(Some(api::ChunkTableInfo { hash: name.0.to_vec(), chunk_count: count }))
    }
}

impl ChunkReader for RemoteStore {
    fn get_many(&self, hashes: &[Hash]) -> store::Result<Vec<Option<Chunk>>> {
        let missing: Vec<Hash> = {
            let cache = self.cache.lock().map_err(|_| error("a cache lock was poisoned"))?;
            hashes
                .iter()
                .filter(|h| !cache.contains_key(h) && !self.pending.iter().any(|c| c.hash == **h))
                .copied()
                .collect()
        };
        self.download(&missing)?;
        hashes.iter().map(|hash| self.get(hash)).collect()
    }

    fn get(&self, hash: &Hash) -> store::Result<Option<Chunk>> {
        if let Some(chunk) = self.pending.iter().find(|c| c.hash == *hash) {
            return Ok(Some(chunk.clone()));
        }
        if let Some(chunk) = self.cache.lock().map_err(|_| error("a cache lock was poisoned"))?.get(hash) {
            return Ok(Some(chunk.clone()));
        }
        self.download(&[*hash])?;
        Ok(self.cache.lock().map_err(|_| error("a cache lock was poisoned"))?.get(hash).cloned())
    }
}

impl ChunkStore for RemoteStore {
    fn has_many(&self, hashes: &[Hash]) -> Vec<bool> {
        let mut held = vec![true; hashes.len()];
        for (start, batch) in hashes.chunks(BATCH).enumerate().map(|(i, b)| (i * BATCH, b)) {
            let request = api::HasChunksRequest {
                repo_path: self.repo.clone(),
                hashes: batch.iter().map(|h| h.0.to_vec()).collect(),
                ..Default::default()
            };
            let mut client = self.client.clone();
            match self.runtime.block_on(client.has_chunks(request)) {
                Ok(response) => {
                    for absent in response.into_inner().absent {
                        held[start + absent as usize] = false;
                    }
                }
                Err(_) => held[start..start + batch.len()].iter_mut().for_each(|h| *h = false),
            }
        }
        for (hash, held) in hashes.iter().zip(held.iter_mut()) {
            *held |= self.pending.iter().any(|c| c.hash == *hash);
        }
        held
    }

    fn has(&self, hash: &Hash) -> bool {
        if self.pending.iter().any(|c| c.hash == *hash) {
            return true;
        }
        let request =
            api::HasChunksRequest { repo_path: self.repo.clone(), hashes: vec![hash.0.to_vec()], ..Default::default() };
        let mut client = self.client.clone();
        match self.runtime.block_on(client.has_chunks(request)) {
            Ok(response) => response.into_inner().absent.is_empty(),
            Err(_) => false,
        }
    }

    fn put(&mut self, chunk: Chunk, _: Vec<Hash>) -> store::Result<()> {
        if !self.pending.iter().any(|c| c.hash == chunk.hash) {
            self.pending.push(chunk);
        }
        Ok(())
    }

    fn commit(&mut self, current: Hash, last: Hash) -> store::Result<bool> {
        let info = self.upload()?;
        let request = api::CommitRequest {
            repo_path: self.repo.clone(),
            current: current.0.to_vec(),
            last: last.0.to_vec(),
            chunk_table_info: info.into_iter().collect(),
            client_repo_format: Some(api::ClientRepoFormat {
                nbf_version: "__DOLT__".to_string(),
                nbs_version: "5".to_string(),
            }),
            ..Default::default()
        };
        let mut client = self.client.clone();
        let success = self.runtime.block_on(client.commit(request)).map_err(error)?.into_inner().success;
        self.pending.clear();
        self.root = if success { current } else { self.fetch_root()? };
        Ok(success)
    }

    fn root(&self) -> Hash {
        self.root
    }
}
