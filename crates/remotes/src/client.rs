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
use std::sync::{Arc, Mutex};

use md5::{Digest, Md5};
use store::{Chunk, ChunkReader, ChunkStore, Hash, TableWriter};
use tonic::transport::Channel;

use crate::cluster::{EPOCH_HEADER, Member, ROLE_HEADER};
use crate::remotesapi::chunk_store_service_client::ChunkStoreServiceClient;
use crate::remotesapi::{self as api, download_loc, upload_loc};
use crate::replicationapi;
use crate::replicationapi::replication_service_client::ReplicationServiceClient;
use crate::ssh::Ssh;

/// BATCH is how many chunks one download location request asks for.
const BATCH: usize = 4096;

/// Client is the gRPC client of a remote's chunk store service.
type Client = ChunkStoreServiceClient<Channel>;

/// RemoteStore is a chunk store on a remote server.
pub struct RemoteStore {
    runtime: tokio::runtime::Runtime,
    client: Client,
    transport: Transport,
    /// The repository path that requests name.
    repo: String,
    root: Hash,
    /// The chunks put since the last commit, which the commit uploads.
    pending: Vec<Chunk>,
    /// The chunks downloaded so far.
    cache: Mutex<HashMap<Hash, Chunk>>,
    /// The decompressed archive dictionaries downloaded so far, by URL and offset.
    dictionaries: Mutex<HashMap<(String, u64), Vec<u8>>>,
    /// The cluster member that replicates through the store, if any.
    member: Option<Arc<dyn Member>>,
}

/// Transport carries the table file downloads and uploads.
enum Transport {
    /// Requests go to the hosts that their URLs name.
    Http(reqwest::Client),
    /// Requests go over an ssh remote's session.
    Ssh(Ssh),
}

/// error returns a store error for a failed remote call.
fn error(err: impl std::fmt::Display) -> store::Error {
    store::Error::Io(std::io::Error::other(err.to_string()))
}

impl RemoteStore {
    /// open connects to a remote at a URL such as `http://host:port/repo` or `ssh://user@host/path` and reads its root.
    pub fn open(url: &str) -> Result<RemoteStore, String> {
        if url.starts_with("ssh://") {
            return RemoteStore::open_ssh(url);
        }
        RemoteStore::open_as(url, None)
    }

    /// open_ssh runs `dolt transfer` on an ssh remote's host and reads the root of the repository it serves.
    fn open_ssh(url: &str) -> Result<RemoteStore, String> {
        let runtime = tokio::runtime::Builder::new_multi_thread()
            .worker_threads(1)
            .enable_all()
            .build()
            .map_err(|e| e.to_string())?;
        let mut ssh = runtime.block_on(async { Ssh::start(url) })?;
        let client = match runtime.block_on(ssh.channel()) {
            Ok(channel) => ChunkStoreServiceClient::new(channel)
                .max_decoding_message_size(usize::MAX)
                .max_encoding_message_size(usize::MAX),
            Err(err) => return Err(ssh.failure("failed to create gRPC client", err)),
        };
        let repo = url.split_once("://").and_then(|(_, rest)| rest.split_once('/')).map_or("", |(_, path)| path);
        RemoteStore::start(runtime, client, Transport::Ssh(ssh), repo, None).map_err(|(err, transport)| match transport
        {
            Transport::Ssh(mut ssh) => ssh.failure("failed to create chunk store", err),
            Transport::Http(_) => err,
        })
    }

    /// open_as connects to a remote as open does, sending a cluster member's role, epoch, and token with each request
    /// when given one.
    pub fn open_as(url: &str, member: Option<Arc<dyn Member>>) -> Result<RemoteStore, String> {
        let (scheme, rest) = url.split_once("://").ok_or_else(|| format!("invalid remote url: {url}"))?;
        let (host, repo) = rest.split_once('/').unwrap_or((rest, ""));
        let runtime = tokio::runtime::Builder::new_current_thread().enable_all().build().map_err(|e| e.to_string())?;
        let endpoint = format!("{scheme}://{host}");
        let client = runtime
            .block_on(ChunkStoreServiceClient::connect(endpoint))
            .map_err(|err| err.to_string())?
            .max_decoding_message_size(usize::MAX)
            .max_encoding_message_size(usize::MAX);
        RemoteStore::start(runtime, client, Transport::Http(reqwest::Client::new()), repo, member)
            .map_err(|(err, _)| err)
    }

    /// start checks the remote's storage format and reads its root, handing back the transport when either fails.
    fn start(
        runtime: tokio::runtime::Runtime,
        client: Client,
        transport: Transport,
        repo: &str,
        member: Option<Arc<dyn Member>>,
    ) -> Result<RemoteStore, (String, Transport)> {
        let mut store = RemoteStore {
            runtime,
            client,
            transport,
            repo: repo.trim_matches('/').to_string(),
            root: Hash::default(),
            pending: Vec::new(),
            cache: Mutex::new(HashMap::new()),
            dictionaries: Mutex::new(HashMap::new()),
            member,
        };
        match store.metadata().and_then(|_| store.fetch_root()) {
            Ok(root) => {
                store.root = root;
                Ok(store)
            }
            Err(err) => Err((err.to_string(), store.transport)),
        }
    }

    /// call sends a request through exchange.
    fn call<T, R, F>(&self, message: T, send: impl FnOnce(Client, tonic::Request<T>) -> F) -> store::Result<R>
    where
        F: std::future::Future<Output = Result<tonic::Response<R>, tonic::Status>>,
    {
        exchange(&self.runtime, self.client.clone(), self.member.as_deref(), message, send).map_err(error)
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
        let response = self.call(request, |mut c, r| async move { c.get_repo_metadata(r).await })?;
        if response.nbf_version != "__DOLT__" {
            return Err(error(format!("remote has unsupported storage format {}", response.nbf_version)));
        }
        Ok(())
    }

    /// fetch_root reads the remote's root.
    fn fetch_root(&self) -> store::Result<Hash> {
        let request = api::RootRequest { repo_path: self.repo.clone(), ..Default::default() };
        let response = self.call(request, |mut c, r| async move { c.root(r).await })?;
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
            let response = self.call(request, |mut c, r| async move { c.get_download_locations(r).await })?;
            for location in response.locs {
                let Some(download_loc::Location::HttpGetRange(range)) = location.location else { continue };
                self.fetch_ranges(&range)?;
            }
        }
        Ok(())
    }

    /// get_range downloads a byte range of a URL.
    fn get_range(&self, url: &str, offset: u64, length: u64) -> store::Result<Vec<u8>> {
        let range = format!("bytes={}-{}", offset, offset + length.max(1) - 1);
        let (status, body) = match &self.transport {
            Transport::Http(http) => self
                .runtime
                .block_on(async {
                    let response = http.get(url).header("range", range).send().await?;
                    Ok::<_, reqwest::Error>((response.status(), response.bytes().await?))
                })
                .map_err(error)?,
            Transport::Ssh(ssh) => {
                let request = axum::http::Request::get(url).header("range", range).body(Default::default());
                self.runtime.block_on(ssh.send(request.map_err(error)?)).map_err(error)?
            }
        };
        if !status.is_success() {
            return Err(error(format!("download of {url} failed with {status}")));
        }
        Ok(body.to_vec())
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
        let response = self.call(request, |mut c, r| async move { c.get_upload_locations(r).await })?;
        for location in response.locs {
            let Some(upload_loc::Location::HttpPost(post)) = location.location else { continue };
            let status = match &self.transport {
                Transport::Http(http) => {
                    self.runtime.block_on(http.put(&post.url).body(bytes.clone()).send()).map_err(error)?.status()
                }
                Transport::Ssh(ssh) => {
                    let request = axum::http::Request::put(&post.url).body(bytes.clone().into()).map_err(error)?;
                    self.runtime.block_on(ssh.send(request)).map_err(error)?.0
                }
            };
            if !status.is_success() {
                return Err(error(format!("upload of table file {name} failed with {status}")));
            }
        }
        Ok(Some(api::ChunkTableInfo { hash: name.0.to_vec(), chunk_count: count }))
    }
}

/// exchange sends a request, with the cluster member's role, epoch, and token when there is one, and lets the member
/// learn from the role and epoch of the response, as Dolt's cluster clientinterceptor does.
fn exchange<C, T, R, F>(
    runtime: &tokio::runtime::Runtime,
    client: C,
    member: Option<&dyn Member>,
    message: T,
    send: impl FnOnce(C, tonic::Request<T>) -> F,
) -> Result<R, String>
where
    F: std::future::Future<Output = Result<tonic::Response<R>, tonic::Status>>,
{
    let mut request = tonic::Request::new(message);
    if let Some(member) = member {
        let (role, epoch) = member.role();
        let state = match role.as_str() {
            "primary" => None,
            "standby" => Some("a standby"),
            _ => Some("in detected_broken_config"),
        };
        if let Some(state) = state {
            return Err(format!(
                "cluster: clientinterceptor: this server is {state} and is not currently replicating to its standby"
            ));
        }
        let metadata = request.metadata_mut();
        let mut insert = |name: &'static str, value: String| {
            if let Ok(value) = value.parse() {
                metadata.insert(name, value);
            }
        };
        insert(ROLE_HEADER, role);
        insert(EPOCH_HEADER, epoch.to_string());
        insert("authorization", format!("Bearer {}", member.credentials().token()));
    }
    let result = runtime.block_on(send(client, request));
    if let Some(member) = member {
        let metadata = match &result {
            Ok(response) => response.metadata(),
            Err(status) => status.metadata(),
        };
        learn(member, metadata);
    }
    result
        .map(tonic::Response::into_inner)
        .map_err(|status| format!("rpc error: code = {:?} desc = {}", status.code(), status.message()))
}

/// learn moves a primary to the role that a standby's response shows it must take, as Dolt's
/// handleResponseHeaders does.
fn learn(member: &dyn Member, metadata: &tonic::metadata::MetadataMap) {
    let (role, epoch) = member.role();
    let header = |name: &str| metadata.get(name).and_then(|v| v.to_str().ok());
    let (Some(from_role), Some(Ok(from_epoch))) = (header(ROLE_HEADER), header(EPOCH_HEADER).map(str::parse::<i64>))
    else {
        return;
    };
    if role != "primary" {
        return;
    }
    if from_role == "primary" && from_epoch == epoch {
        member.force_role("detected_broken_config", from_epoch);
    } else if from_role == "primary" && from_epoch > epoch {
        member.force_role("standby", from_epoch);
    } else if from_role == "detected_broken_config" && from_epoch >= epoch {
        member.force_role("detected_broken_config", from_epoch);
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
            match self.call(request, |mut c, r| async move { c.has_chunks(r).await }) {
                Ok(response) => {
                    for absent in response.absent {
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
        match self.call(request, |mut c, r| async move { c.has_chunks(r).await }) {
            Ok(response) => response.absent.is_empty(),
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
        let success = self.call(request, |mut c, r| async move { c.commit(r).await })?.success;
        self.pending.clear();
        self.root = if success { current } else { self.fetch_root()? };
        Ok(success)
    }

    fn root(&self) -> Hash {
        self.root
    }
}

/// Replica sends a standby what a primary replicates outside of its databases: roles and privileges, branch control,
/// and dropped databases, as Dolt's replicationServiceClient does.
pub struct Replica {
    runtime: tokio::runtime::Runtime,
    client: ReplicationServiceClient<Channel>,
    member: Arc<dyn Member>,
}

impl Replica {
    /// connect connects to the standby at the host and port of a URL such as `http://host:port/`.
    pub fn connect(url: &str, member: Arc<dyn Member>) -> Result<Replica, String> {
        let (scheme, rest) = url.split_once("://").ok_or_else(|| format!("invalid remote url: {url}"))?;
        let host = rest.split('/').next().unwrap_or(rest);
        let runtime = tokio::runtime::Builder::new_current_thread().enable_all().build().map_err(|e| e.to_string())?;
        let client = runtime
            .block_on(ReplicationServiceClient::connect(format!("{scheme}://{host}")))
            .map_err(|err| err.to_string())?
            .max_decoding_message_size(usize::MAX)
            .max_encoding_message_size(usize::MAX);
        Ok(Replica { runtime, client, member })
    }

    /// update_users sends the serialized roles and privileges.
    pub fn update_users(&self, contents: Vec<u8>) -> Result<(), String> {
        let request = replicationapi::UpdateUsersAndGrantsRequest { serialized_contents: contents };
        let member = Some(self.member.as_ref());
        exchange(&self.runtime, self.client.clone(), member, request, |mut c, r| async move {
            c.update_users_and_grants(r).await
        })
        .map(|_| ())
    }

    /// update_branch_control sends the serialized branch control tables.
    pub fn update_branch_control(&self, contents: Vec<u8>) -> Result<(), String> {
        let request = replicationapi::UpdateBranchControlRequest { serialized_contents: contents };
        let member = Some(self.member.as_ref());
        exchange(&self.runtime, self.client.clone(), member, request, |mut c, r| async move {
            c.update_branch_control(r).await
        })
        .map(|_| ())
    }

    /// drop_database tells the standby to drop a database.
    pub fn drop_database(&self, name: &str) -> Result<(), String> {
        let request = replicationapi::DropDatabaseRequest { name: name.to_string() };
        let member = Some(self.member.as_ref());
        exchange(
            &self.runtime,
            self.client.clone(),
            member,
            request,
            |mut c, r| async move { c.drop_database(r).await },
        )
        .map(|_| ())
    }
}
