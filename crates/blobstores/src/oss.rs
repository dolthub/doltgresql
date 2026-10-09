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

//! Alibaba Cloud Object Storage Service, with requests signed by OSS's version 1 signature, as Dolt's OSSBlobstore uses
//! it for oss:// URLs.

use std::collections::BTreeMap;

use base64::Engine;
use hmac::{Hmac, Mac};
use reqwest::Method;
use reqwest::blocking::Response;
use serde_json::Value as Json;
use sha1::Sha1;
use store::{Blob, BlobRange, Blobstore, Error, MANIFEST_KEY, Result, not_found};

use crate::http::{client, failure, status_error, uri_encode};

/// invalid returns the error of an invalid configuration or response.
fn invalid(message: impl Into<String>) -> Error {
    Error::Corrupt(message.into())
}

/// OssBlobstore keeps blobs as objects under a prefix of an OSS bucket, whose versions are the objects' version IDs
/// when the bucket keeps versions and are otherwise empty.
pub(crate) struct OssBlobstore {
    scheme: String,
    endpoint: String,
    key_id: String,
    secret: String,
    bucket: String,
    prefix: String,
    versioned: bool,
}

/// credential returns the endpoint and access key of a remote's parameters, from the credentials file that they or the
/// home directory name and then the OSS_ environment variables, as Dolt's ossConfigFromParams finds them.
fn credential(params: &BTreeMap<String, String>) -> Result<(String, String, String)> {
    let file = params
        .get("oss-creds-file")
        .cloned()
        .unwrap_or_else(|| format!("{}/.oss/dolt_oss_credentials", std::env::var("HOME").unwrap_or_default()));
    let profile = params.get("oss-creds-profile").cloned().unwrap_or_else(|| "default".into());
    let credentials: Json =
        std::fs::read_to_string(&file).ok().and_then(|t| serde_json::from_str(&t).ok()).unwrap_or(Json::Null);
    let chosen = match credentials.as_object() {
        Some(all) if all.len() == 1 => all.values().next().cloned().unwrap_or(Json::Null),
        Some(all) => all.get(&profile).cloned().unwrap_or(Json::Null),
        None => Json::Null,
    };
    let field = |name: &str, variable: &str, what: &str| {
        chosen[name]
            .as_str()
            .filter(|v| !v.is_empty())
            .map(str::to_string)
            .or_else(|| std::env::var(variable).ok().filter(|v| !v.is_empty()))
            .ok_or_else(|| {
                invalid(format!("failed to initialize oss err: failed to find {what} from cred file or env {variable}"))
            })
    };
    Ok((
        field("endpoint", "OSS_ENDPOINT", "endpoint")?,
        field("accessKeyID", "OSS_ACCESS_KEY_ID", "accessKeyID")?,
        field("accessKeySecret", "OSS_ACCESS_KEY_SECRET", "accessKeySecret")?,
    ))
}

/// open_oss opens the blobstore of an oss://bucket/path URL, asking whether the bucket keeps versions.
pub(crate) fn open_oss(rest: &str, params: &BTreeMap<String, String>) -> Result<OssBlobstore> {
    let (bucket, prefix) = rest.split_once('/').unwrap_or((rest, ""));
    let (endpoint, key_id, secret) = credential(params)?;
    let (scheme, endpoint) = endpoint.split_once("://").unwrap_or(("http", &endpoint));
    let mut store = OssBlobstore {
        scheme: scheme.to_string(),
        endpoint: endpoint.trim_end_matches('/').to_string(),
        key_id,
        secret,
        bucket: bucket.to_string(),
        prefix: prefix.trim_start_matches('/').to_string(),
        versioned: false,
    };
    let response = store.send(Method::GET, "", "versioning", &[], &[])?;
    if !response.status().is_success() {
        return Err(invalid("failed to initialize oss blob store"));
    }
    store.versioned = response.text().map_err(failure)?.contains("<Status>Enabled</Status>");
    Ok(store)
}

impl OssBlobstore {
    /// name returns an object's name.
    fn name(&self, key: &str) -> String {
        if self.prefix.is_empty() { key.to_string() } else { format!("{}/{key}", self.prefix.trim_end_matches('/')) }
    }

    /// send signs a request for an object, or for the bucket with an empty name, with a subresource, and sends it.
    fn send(
        &self,
        method: Method,
        object: &str,
        subresource: &str,
        headers: &[(&str, &str)],
        body: &[u8],
    ) -> Result<Response> {
        let date = chrono::Utc::now().format("%a, %d %b %Y %H:%M:%S GMT").to_string();
        let content_type = headers.iter().find(|(k, _)| k.eq_ignore_ascii_case("content-type")).map_or("", |(_, v)| *v);
        let mut oss_headers: Vec<(String, &str)> = headers
            .iter()
            .filter(|(k, _)| k.to_lowercase().starts_with("x-oss-"))
            .map(|(k, v)| (k.to_lowercase(), *v))
            .collect();
        oss_headers.sort();
        let canonical_headers: String = oss_headers.iter().map(|(k, v)| format!("{k}:{v}\n")).collect();
        let query = if subresource.is_empty() { String::new() } else { format!("?{subresource}") };
        let resource = format!("/{}/{object}{query}", self.bucket);
        let to_sign = format!("{}\n\n{content_type}\n{date}\n{canonical_headers}{resource}", method.as_str());
        let mut mac = Hmac::<Sha1>::new_from_slice(self.secret.as_bytes()).expect("HMAC takes keys of any length");
        mac.update(to_sign.as_bytes());
        let signature = base64::engine::general_purpose::STANDARD.encode(mac.finalize().into_bytes());
        let url = format!("{}://{}.{}/{}{query}", self.scheme, self.bucket, self.endpoint, uri_encode(object, true));
        let mut request = client()?
            .request(method, url)
            .header("Date", date)
            .header("Authorization", format!("OSS {}:{signature}", self.key_id));
        for (name, value) in headers {
            request = request.header(*name, *value);
        }
        request.body(body.to_vec()).send().map_err(failure)
    }

    /// version returns the version ID of a response when the bucket keeps versions.
    fn version(&self, response: &Response) -> String {
        if !self.versioned {
            return String::new();
        }
        response.headers().get("x-oss-version-id").and_then(|v| v.to_str().ok()).unwrap_or_default().to_string()
    }

    /// size returns an object's size, or None when the bucket lacks it.
    fn size(&self, key: &str) -> Result<Option<(u64, String)>> {
        let response = self.send(Method::HEAD, &self.name(key), "objectMeta", &[], &[])?;
        match response.status().as_u16() {
            200 => {
                let size = response
                    .headers()
                    .get("content-length")
                    .and_then(|v| v.to_str().ok())
                    .and_then(|v| v.parse().ok())
                    .unwrap_or(0);
                Ok(Some((size, self.version(&response))))
            }
            404 => Ok(None),
            _ => Err(status_error("GetObjectMeta", response)),
        }
    }
}

impl Blobstore for OssBlobstore {
    fn path(&self) -> String {
        format!("{}/{}", self.bucket, self.prefix).trim_end_matches('/').to_string()
    }

    fn exists(&self, key: &str) -> Result<bool> {
        Ok(self.size(key)?.is_some())
    }

    fn get(&self, key: &str, range: BlobRange) -> Result<Blob> {
        let name = self.name(key);
        let (size, version) = self.size(key)?.ok_or_else(|| not_found(&format!("oss://{}/{name}", self.bucket)))?;
        let header = (!range.is_all()).then(|| {
            let range = range.positive(size as i64);
            format!("bytes={}-{}", range.offset, range.offset + range.length - 1)
        });
        let headers: Vec<(&str, &str)> = header.iter().map(|h| ("Range", h.as_str())).collect();
        let response = self.send(Method::GET, &name, "", &headers, &[])?;
        if !response.status().is_success() {
            return Err(status_error("GetObject", response));
        }
        Ok(Blob { data: response.bytes().map_err(failure)?.to_vec(), size, version })
    }

    fn put(&self, key: &str, data: &[u8]) -> Result<String> {
        let response = self.send(Method::PUT, &self.name(key), "", &[], data)?;
        if !response.status().is_success() {
            return Err(status_error("PutObject", response));
        }
        Ok(self.version(&response))
    }

    fn check_and_put_manifest(&self, expected: &str, data: &[u8]) -> Result<String> {
        let subresource = if expected.is_empty() { String::new() } else { format!("versionId={expected}") };
        let response = self.send(Method::PUT, &self.name(MANIFEST_KEY), &subresource, &[], data)?;
        if !response.status().is_success() {
            return Err(Error::VersionMismatch {
                key: MANIFEST_KEY.into(),
                expected: expected.into(),
                actual: format!("unknown (OSS error code {})", response.status().as_u16()),
            });
        }
        Ok(self.version(&response))
    }

    fn concatenate(&self, _: &str, _: &[String]) -> Result<String> {
        Err(invalid("Conjoin is not implemented for OSSBlobstore"))
    }
}
