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

//! The remotes API server, a port of Dolt's remotesrv: the chunk store service over the server's databases, and an
//! HTTP handler on the same port that serves ranges of their table files and accepts uploaded ones.

#![allow(clippy::result_large_err)]

use std::collections::BTreeMap;
use std::io::Read;
use std::sync::Arc;

use axum::body::Bytes;
use axum::http::{HeaderMap, Method, StatusCode, Uri};
use axum::response::{IntoResponse, Response};
use doltdb::database::Database;
use doltdb::handle::Handle;
use md5::{Digest, Md5};
use store::{Hash, TableSpec};
use tokio_stream::StreamExt;
use tonic::{Request, Status};

use crate::cluster::{EPOCH_HEADER, JWKS_PATH, Member, ROLE_HEADER};
use crate::remotesapi::chunk_store_service_server::{ChunkStoreService, ChunkStoreServiceServer};
use crate::remotesapi::{self as api, download_loc, upload_loc};
use crate::replicationapi;
use crate::replicationapi::replication_service_server::{ReplicationService, ReplicationServiceServer};
use crate::sealer::Sealer;

/// Databases gives the server the databases it serves by name.
pub trait Databases: Send + Sync + 'static {
    /// database returns the open database with the name, or None when there is none.
    fn database(&self, name: &str) -> Option<Arc<Handle>>;

    /// committed notes that a commit through the server moved a database's root.
    fn committed(&self, _name: &str) {}
}

/// Shared is what the gRPC service and the HTTP handler share.
struct Shared {
    databases: Arc<dyn Databases>,
    read_only: bool,
    sealer: Sealer,
}

/// serve serves the remotes API on a listener until the listener fails, running its own async runtime, and serves a
/// cluster member's replication to it when given one.
pub fn serve(
    listener: std::net::TcpListener,
    databases: Arc<dyn Databases>,
    read_only: bool,
    member: Option<Arc<dyn Member>>,
) -> std::io::Result<()> {
    let runtime = tokio::runtime::Builder::new_multi_thread().enable_all().build()?;
    runtime.block_on(async move {
        listener.set_nonblocking(true)?;
        let listener = tokio::net::TcpListener::from_std(listener)?;
        let shared = Arc::new(Shared { databases, read_only, sealer: Sealer::new() });
        let service = ChunkStoreServiceServer::new(Service { shared: shared.clone() })
            .max_decoding_message_size(usize::MAX)
            .max_encoding_message_size(usize::MAX);
        let mut routes = tonic::service::Routes::new(service);
        if let Some(member) = &member {
            routes = routes.add_service(ReplicationServiceServer::new(Replication { member: member.clone() }));
        }
        let mut router = routes
            .into_axum_router()
            .fallback(move |method: Method, uri: Uri, headers: HeaderMap, body: Bytes| {
                files(shared.clone(), method, uri, headers, body)
            })
            .layer(axum::middleware::map_request(authority_header));
        if let Some(member) = member {
            router = router.layer(axum::middleware::from_fn_with_state(member, gate));
        }
        axum::serve(listener, router).await
    })
}

/// WRITES are the requests that write to a database, which a request from outside the cluster is told are
/// unimplemented rather than unauthenticated, as Dolt's writeEndpoints are.
const WRITES: [&str; 3] = [
    "/dolt.services.remotesapi.v1alpha1.ChunkStoreService/Commit",
    "/dolt.services.remotesapi.v1alpha1.ChunkStoreService/AddTableFiles",
    "/dolt.services.remotesapi.v1alpha1.ChunkStoreService/GetUploadLocations",
];

/// gate serves a cluster member's published key, and lets through only the requests of an authenticated primary to
/// a standby, answering each with the member's role and epoch, as Dolt's cluster serverinterceptor does.
async fn gate(
    axum::extract::State(member): axum::extract::State<Arc<dyn Member>>,
    request: axum::extract::Request,
    next: axum::middleware::Next,
) -> Response {
    let path = request.uri().path().to_string();
    if path == JWKS_PATH {
        return ([(axum::http::header::CONTENT_TYPE, "application/json")], member.credentials().jwks()).into_response();
    }
    if !path.starts_with("/dolt.services.") {
        return next.run(request).await;
    }
    let [from_role, from_epoch, authorization] = [ROLE_HEADER, EPOCH_HEADER, "authorization"]
        .map(|name| request.headers().get(name).and_then(|v| v.to_str().ok()).map(str::to_string));
    let token = authorization.and_then(|a| a.strip_prefix("Bearer ").map(str::to_string));
    let (Some(from_role), Some(from_epoch)) = (from_role, from_epoch) else {
        return match WRITES.contains(&path.as_str()) {
            true => Status::unimplemented("unimplemented").into_http(),
            false => Status::unauthenticated("unauthenticated").into_http(),
        };
    };
    let (role, epoch) = member.role();
    if let (true, Ok(from_epoch)) = (from_role == "primary", from_epoch.parse::<i64>()) {
        if from_epoch == epoch && role == "primary" {
            member.force_role("detected_broken_config", from_epoch);
        } else if from_epoch > epoch {
            member.force_role("standby", from_epoch);
        }
    }
    if !matches!(&token, Some(token) if member.keys().verify(token).await) {
        return Status::unauthenticated("unauthenticated").into_http();
    }
    let (role, epoch) = member.role();
    let mut response = match role.as_str() {
        "primary" => Status::failed_precondition("this server is a primary and is not currently accepting replication")
            .into_http(),
        "detected_broken_config" => Status::failed_precondition(
            "this server is currently in detected_broken_config and is not currently accepting replication",
        )
        .into_http(),
        _ => next.run(request).await,
    };
    if let (Ok(role), Ok(epoch)) = (role.parse(), epoch.to_string().parse()) {
        response.headers_mut().insert(ROLE_HEADER, role);
        response.headers_mut().insert(EPOCH_HEADER, epoch);
    }
    response
}

/// Replication implements the replication service, through which a primary sends a standby what its databases do not
/// hold, as Dolt's replicationServiceServer does.
struct Replication {
    member: Arc<dyn Member>,
}

#[tonic::async_trait]
impl ReplicationService for Replication {
    async fn update_users_and_grants(
        &self,
        request: Request<replicationapi::UpdateUsersAndGrantsRequest>,
    ) -> Result<tonic::Response<replicationapi::UpdateUsersAndGrantsResponse>, Status> {
        self.member.update_users(&request.into_inner().serialized_contents).map_err(internal)?;
        Ok(tonic::Response::new(replicationapi::UpdateUsersAndGrantsResponse {}))
    }

    async fn update_branch_control(
        &self,
        request: Request<replicationapi::UpdateBranchControlRequest>,
    ) -> Result<tonic::Response<replicationapi::UpdateBranchControlResponse>, Status> {
        self.member.update_branch_control(&request.into_inner().serialized_contents).map_err(internal)?;
        Ok(tonic::Response::new(replicationapi::UpdateBranchControlResponse {}))
    }

    async fn drop_database(
        &self,
        request: Request<replicationapi::DropDatabaseRequest>,
    ) -> Result<tonic::Response<replicationapi::DropDatabaseResponse>, Status> {
        self.member.drop_database(&request.into_inner().name).map_err(internal)?;
        Ok(tonic::Response::new(replicationapi::DropDatabaseResponse {}))
    }
}

/// AUTHORITY_HEADER carries the host and port that a request reached the server at, which gRPC requests only carry
/// in their URI.
const AUTHORITY_HEADER: &str = "x-dolt-authority";

/// authority_header copies a request URI's authority into a header that the gRPC service can read.
async fn authority_header(mut request: axum::extract::Request) -> axum::extract::Request {
    if let Some(value) = request.uri().authority().and_then(|a| a.as_str().parse().ok()) {
        request.headers_mut().insert(AUTHORITY_HEADER, value);
    }
    request
}

/// Service implements the chunk store service.
struct Service {
    shared: Arc<Shared>,
}

/// repo_path returns the repository a request names, by its path or by its organization and name.
fn repo_path(path: &str, id: Option<&api::RepoId>) -> String {
    match id {
        Some(id) if path.is_empty() => format!("{}/{}", id.org, id.repo_name),
        _ => path.to_string(),
    }
}

/// hash reads an address from a request.
fn hash(bytes: &[u8]) -> Result<Hash, Status> {
    let bytes: [u8; 20] = bytes.try_into().map_err(|_| Status::invalid_argument("invalid chunk hash"))?;
    Ok(Hash(bytes))
}

/// internal returns an internal error status.
fn internal(err: impl std::fmt::Display) -> Status {
    Status::internal(err.to_string())
}

/// origin returns the scheme and host that the client reached the server at, which URLs the server hands out use.
fn origin<T>(request: &Request<T>) -> String {
    let metadata = request.metadata();
    let scheme = metadata.get("x-forwarded-proto").and_then(|v| v.to_str().ok()).unwrap_or("http");
    let host = metadata.get(AUTHORITY_HEADER).or_else(|| metadata.get("host")).and_then(|v| v.to_str().ok());
    format!("{scheme}://{}", host.unwrap_or("localhost"))
}

impl Shared {
    /// database returns the database a repository path names.
    fn database(&self, repo: &str) -> Result<Arc<Handle>, Status> {
        self.databases.database(repo).ok_or_else(|| Status::not_found(format!("database not found: {repo}")))
    }

    /// sealed_url returns a sealed URL for a path relative to the server's root.
    fn sealed_url(&self, origin: &str, path: &str, query: &str) -> String {
        let (path, query) = self.sealer.seal(&format!("/{path}"), query);
        format!("{origin}{path}?{query}")
    }

    /// download_locations returns where the chunks are in a repository's files, grouped by file, with sealed URLs.
    fn download_locations(
        &self,
        origin: &str,
        repo: &str,
        hashes: &[Vec<u8>],
    ) -> Result<Vec<api::DownloadLoc>, Status> {
        let database = self.database(repo)?;
        let mut db = database.write();
        let mut by_file: BTreeMap<String, Vec<api::RangeChunk>> = BTreeMap::new();
        for bytes in hashes {
            let hash = hash(bytes)?;
            let Some(location) = db.locate(&hash).map_err(internal)? else { continue };
            let (dictionary_offset, dictionary_length) = location.dictionary.unwrap_or_default();
            by_file.entry(location.file).or_default().push(api::RangeChunk {
                hash: bytes.clone(),
                offset: location.offset,
                length: location.length,
                dictionary_offset,
                dictionary_length,
            });
        }
        Ok(by_file
            .into_iter()
            .map(|(file, ranges)| api::DownloadLoc {
                location: Some(download_loc::Location::HttpGetRange(api::HttpGetRange {
                    url: self.sealed_url(origin, &format!("{repo}/.dolt/noms/{file}"), ""),
                    ranges,
                })),
                ..Default::default()
            })
            .collect())
    }
}

#[tonic::async_trait]
impl ChunkStoreService for Service {
    async fn get_repo_metadata(
        &self,
        request: Request<api::GetRepoMetadataRequest>,
    ) -> Result<tonic::Response<api::GetRepoMetadataResponse>, Status> {
        let req = request.into_inner();
        let database = self.shared.database(&repo_path(&req.repo_path, req.repo_id.as_ref()))?;
        let mut db = database.write();
        let dir = db.noms_dir().map_err(internal)?;
        let storage_size = walk_size(&dir);
        Ok(tonic::Response::new(api::GetRepoMetadataResponse {
            nbf_version: "__DOLT__".to_string(),
            nbs_version: req.client_repo_format.map(|f| f.nbs_version).unwrap_or_default(),
            storage_size,
            ..Default::default()
        }))
    }

    async fn has_chunks(
        &self,
        request: Request<api::HasChunksRequest>,
    ) -> Result<tonic::Response<api::HasChunksResponse>, Status> {
        let req = request.into_inner();
        let database = self.shared.database(&repo_path(&req.repo_path, req.repo_id.as_ref()))?;
        let db = database.write();
        let mut absent = Vec::new();
        for (i, bytes) in req.hashes.iter().enumerate() {
            if !db.has(&hash(bytes)?) {
                absent.push(i as i32);
            }
        }
        Ok(tonic::Response::new(api::HasChunksResponse { absent, ..Default::default() }))
    }

    async fn get_download_locations(
        &self,
        request: Request<api::GetDownloadLocsRequest>,
    ) -> Result<tonic::Response<api::GetDownloadLocsResponse>, Status> {
        let origin = origin(&request);
        let req = request.into_inner();
        let repo = repo_path(&req.repo_path, req.repo_id.as_ref());
        let locs = self.shared.download_locations(&origin, &repo, &req.chunk_hashes)?;
        Ok(tonic::Response::new(api::GetDownloadLocsResponse { locs, ..Default::default() }))
    }

    type StreamDownloadLocationsStream =
        tokio_stream::wrappers::ReceiverStream<Result<api::GetDownloadLocsResponse, Status>>;

    async fn stream_download_locations(
        &self,
        request: Request<tonic::Streaming<api::GetDownloadLocsRequest>>,
    ) -> Result<tonic::Response<Self::StreamDownloadLocationsStream>, Status> {
        let origin = origin(&request);
        let mut requests = request.into_inner();
        let (send, receive) = tokio::sync::mpsc::channel(16);
        let shared = self.shared.clone();
        tokio::spawn(async move {
            while let Some(req) = requests.next().await {
                let response = req.and_then(|req| {
                    let repo = repo_path(&req.repo_path, req.repo_id.as_ref());
                    let locs = shared.download_locations(&origin, &repo, &req.chunk_hashes)?;
                    Ok(api::GetDownloadLocsResponse { locs, ..Default::default() })
                });
                if send.send(response).await.is_err() {
                    break;
                }
            }
        });
        Ok(tonic::Response::new(tokio_stream::wrappers::ReceiverStream::new(receive)))
    }

    type StreamChunkLocationsStream =
        tokio_stream::wrappers::ReceiverStream<Result<api::StreamChunkLocationsResponse, Status>>;

    async fn stream_chunk_locations(
        &self,
        _: Request<tonic::Streaming<api::StreamChunkLocationsRequest>>,
    ) -> Result<tonic::Response<Self::StreamChunkLocationsStream>, Status> {
        Err(Status::unimplemented("StreamChunkLocations is not advertised"))
    }

    async fn get_upload_locations(
        &self,
        request: Request<api::GetUploadLocsRequest>,
    ) -> Result<tonic::Response<api::GetUploadLocsResponse>, Status> {
        let origin = origin(&request);
        let req = request.into_inner();
        let repo = repo_path(&req.repo_path, req.repo_id.as_ref());
        self.shared.database(&repo)?;
        if req.table_file_details.is_empty() {
            return Err(internal("no table file details provided. Your dolt version is pre 1.0. please upgrade."));
        }
        let mut locs = Vec::new();
        for details in &req.table_file_details {
            let id = hash(&details.id)?;
            let query = format!(
                "split_offset={}&num_chunks={}&content_length={}&content_hash={}",
                details.split_offset,
                details.num_chunks,
                details.content_length,
                base64::Engine::encode(&base64::engine::general_purpose::URL_SAFE_NO_PAD, &details.content_hash)
            );
            let url = self.shared.sealed_url(&origin, &format!("{repo}/{id}{}", details.suffix), &query);
            locs.push(api::UploadLoc {
                table_file_hash: details.id.clone(),
                location: Some(upload_loc::Location::HttpPost(api::HttpPostTableFile { url })),
            });
        }
        Ok(tonic::Response::new(api::GetUploadLocsResponse { locs, ..Default::default() }))
    }

    async fn rebase(
        &self,
        request: Request<api::RebaseRequest>,
    ) -> Result<tonic::Response<api::RebaseResponse>, Status> {
        let req = request.into_inner();
        self.shared.database(&repo_path(&req.repo_path, req.repo_id.as_ref()))?;
        Ok(tonic::Response::new(api::RebaseResponse::default()))
    }

    async fn root(&self, request: Request<api::RootRequest>) -> Result<tonic::Response<api::RootResponse>, Status> {
        let req = request.into_inner();
        let database = self.shared.database(&repo_path(&req.repo_path, req.repo_id.as_ref()))?;
        let db = database.write();
        Ok(tonic::Response::new(api::RootResponse { root_hash: db.root().0.to_vec(), ..Default::default() }))
    }

    async fn commit(
        &self,
        request: Request<api::CommitRequest>,
    ) -> Result<tonic::Response<api::CommitResponse>, Status> {
        let req = request.into_inner();
        if self.shared.read_only {
            return Err(Status::permission_denied("this server only allows reads"));
        }
        let database = self.shared.database(&repo_path(&req.repo_path, req.repo_id.as_ref()))?;
        let mut db = database.write();
        add_table_files(&mut db, &req.chunk_table_info)?;
        let success = db.commit_root(hash(&req.current)?, hash(&req.last)?).map_err(|err| match err {
            doltdb::database::Error::Store(store::Error::DanglingRef(_)) => {
                Status::failed_precondition(format!("failed to commit: {err}"))
            }
            err => internal(format!("failed to commit: {err}")),
        })?;
        drop(db);
        if success {
            self.shared.databases.committed(&repo_path(&req.repo_path, req.repo_id.as_ref()));
        }
        Ok(tonic::Response::new(api::CommitResponse { success }))
    }

    async fn list_table_files(
        &self,
        request: Request<api::ListTableFilesRequest>,
    ) -> Result<tonic::Response<api::ListTableFilesResponse>, Status> {
        let origin = origin(&request);
        let req = request.into_inner();
        let repo = repo_path(&req.repo_path, req.repo_id.as_ref());
        let database = self.shared.database(&repo)?;
        let mut db = database.write();
        db.sync().map_err(internal)?;
        let (root, files) = db.table_files().map_err(internal)?;
        let table_file_info = files
            .into_iter()
            .map(|(path, num_chunks)| {
                let file_id = path.rsplit('/').next().unwrap_or_default().trim_end_matches(".darc").to_string();
                api::TableFileInfo {
                    file_id,
                    num_chunks,
                    url: self.shared.sealed_url(&origin, &format!("{repo}/.dolt/noms/{path}"), ""),
                    ..Default::default()
                }
            })
            .collect();
        Ok(tonic::Response::new(api::ListTableFilesResponse {
            root_hash: root.0.to_vec(),
            table_file_info,
            ..Default::default()
        }))
    }

    async fn refresh_table_file_url(
        &self,
        _: Request<api::RefreshTableFileUrlRequest>,
    ) -> Result<tonic::Response<api::RefreshTableFileUrlResponse>, Status> {
        Err(Status::unimplemented("table file URLs need no refreshing"))
    }

    async fn add_table_files(
        &self,
        request: Request<api::AddTableFilesRequest>,
    ) -> Result<tonic::Response<api::AddTableFilesResponse>, Status> {
        let req = request.into_inner();
        if self.shared.read_only {
            return Err(Status::permission_denied("this server only allows reads"));
        }
        let database = self.shared.database(&repo_path(&req.repo_path, req.repo_id.as_ref()))?;
        let mut db = database.write();
        add_table_files(&mut db, &req.chunk_table_info)?;
        Ok(tonic::Response::new(api::AddTableFilesResponse { success: true, ..Default::default() }))
    }
}

/// add_table_files adds the uploaded files that a commit or AddTableFiles request names to the database.
fn add_table_files(db: &mut Database, infos: &[api::ChunkTableInfo]) -> Result<(), Status> {
    let mut specs = Vec::new();
    for info in infos {
        specs.push(TableSpec { name: hash(&info.hash)?, chunk_count: info.chunk_count });
    }
    db.add_table_files(&specs).map_err(|err| Status::failed_precondition(format!("manifest update error: {err}")))
}

/// walk_size returns the total size of the files under a directory.
fn walk_size(dir: &std::path::Path) -> u64 {
    std::fs::read_dir(dir).into_iter().flatten().flatten().fold(0, |total, entry| match entry.metadata() {
        Ok(metadata) if metadata.is_dir() => total + walk_size(&entry.path()),
        Ok(metadata) => total + metadata.len(),
        Err(_) => total,
    })
}

/// files serves a range of a table file to GET and stores an uploaded one from PUT or POST, for sealed URLs only.
async fn files(shared: Arc<Shared>, method: Method, uri: Uri, headers: HeaderMap, body: Bytes) -> Response {
    let Some((path, query)) = shared.sealer.unseal(uri.path(), uri.query().unwrap_or_default()) else {
        return StatusCode::BAD_REQUEST.into_response();
    };
    let path = path.trim_start_matches('/');
    if path.split('/').any(|part| part == ".." || part.is_empty()) {
        return StatusCode::BAD_REQUEST.into_response();
    }
    match method {
        Method::GET => read_file(&shared, path, headers.get("range").and_then(|r| r.to_str().ok())),
        Method::PUT | Method::POST if shared.read_only => StatusCode::FORBIDDEN.into_response(),
        Method::PUT | Method::POST => write_file(&shared, path, &query, &body),
        _ => StatusCode::METHOD_NOT_ALLOWED.into_response(),
    }
}

/// valid_file_name reports whether a name is a table file's or an archive's.
fn valid_file_name(name: &str) -> bool {
    Hash::parse(name.strip_suffix(".darc").unwrap_or(name)).is_some()
}

/// read_file serves a file under a repository's store, or the range of it that a Range header names.
fn read_file(shared: &Shared, path: &str, range: Option<&str>) -> Response {
    let Some((repo, rest)) = path.split_once("/.dolt/noms/") else { return StatusCode::BAD_REQUEST.into_response() };
    let file = rest.rsplit('/').next().unwrap_or_default();
    if !valid_file_name(file) {
        return StatusCode::BAD_REQUEST.into_response();
    }
    let Ok(database) = shared.database(repo) else { return StatusCode::NOT_FOUND.into_response() };
    let dir = match database.write().noms_dir() {
        Ok(dir) => dir,
        _ => return StatusCode::INTERNAL_SERVER_ERROR.into_response(),
    };
    let Ok(mut handle) = std::fs::File::open(dir.join(rest)) else { return StatusCode::NOT_FOUND.into_response() };
    let size = handle.metadata().map(|m| m.len()).unwrap_or_default();
    let Some(range) = range else {
        let mut bytes = Vec::new();
        return match handle.read_to_end(&mut bytes) {
            Ok(_) => (StatusCode::OK, [("accept-ranges", "bytes")], bytes).into_response(),
            Err(_) => StatusCode::INTERNAL_SERVER_ERROR.into_response(),
        };
    };
    let Some((start, end)) = range.strip_prefix("bytes=").and_then(|r| r.split_once('-')) else {
        return StatusCode::BAD_REQUEST.into_response();
    };
    let (Ok(start), Ok(end)) = (start.trim().parse::<u64>(), end.trim().parse::<u64>()) else {
        return StatusCode::BAD_REQUEST.into_response();
    };
    if end < start || end >= size {
        return StatusCode::BAD_REQUEST.into_response();
    }
    let mut bytes = vec![0; (end - start + 1) as usize];
    let read =
        std::io::Seek::seek(&mut handle, std::io::SeekFrom::Start(start)).and_then(|_| handle.read_exact(&mut bytes));
    if read.is_err() {
        return StatusCode::INTERNAL_SERVER_ERROR.into_response();
    }
    let content_range = format!("bytes {start}-{end}/{size}");
    (StatusCode::PARTIAL_CONTENT, [("accept-ranges", "bytes".to_string()), ("content-range", content_range)], bytes)
        .into_response()
}

/// write_file stores an uploaded table file or archive in a repository's store, checking its length and its MD5
/// against the details the upload URL carries.
fn write_file(shared: &Shared, path: &str, query: &str, body: &[u8]) -> Response {
    let Some((repo, file)) = path.rsplit_once('/') else { return StatusCode::NOT_FOUND.into_response() };
    if !valid_file_name(file) {
        return StatusCode::NOT_FOUND.into_response();
    }
    let parameter = |name: &str| {
        query.split('&').find_map(|pair| pair.strip_prefix(name).and_then(|v| v.strip_prefix('='))).unwrap_or_default()
    };
    let Ok(length) = parameter("content_length").parse::<usize>() else {
        return StatusCode::BAD_REQUEST.into_response();
    };
    if length != 0 && length != body.len() {
        return (StatusCode::BAD_REQUEST, "body upload length did not match table file details").into_response();
    }
    let expected = parameter("content_hash");
    if !expected.is_empty() {
        let digest = Md5::digest(body);
        let encoded = base64::Engine::encode(&base64::engine::general_purpose::URL_SAFE_NO_PAD, digest);
        if encoded != expected {
            return (StatusCode::BAD_REQUEST, "body upload hash did not match table file details").into_response();
        }
    }
    let Ok(database) = shared.database(repo) else { return StatusCode::NOT_FOUND.into_response() };
    let dir = match database.write().noms_dir() {
        Ok(dir) => dir,
        _ => return StatusCode::INTERNAL_SERVER_ERROR.into_response(),
    };
    let temp = dir.join(format!(".upload-{file}"));
    let written = std::fs::write(&temp, body).and_then(|_| std::fs::rename(&temp, dir.join(file)));
    match written {
        Ok(()) => StatusCode::OK.into_response(),
        Err(_) => StatusCode::INTERNAL_SERVER_ERROR.into_response(),
    }
}
