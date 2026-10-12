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

//! Azure Blob Storage, through its REST API with the credentials of Azure's DefaultAzureCredential, as Dolt's
//! AzureBlobstore uses it for az:// URLs.

use std::sync::Mutex;
use std::time::{Duration, Instant};

use base64::Engine;
use reqwest::blocking::{RequestBuilder, Response};
use serde_json::Value as Json;
use store::{Blob, BlobRange, Blobstore, Error, MANIFEST_KEY, Result, not_found};

use crate::http::{client, content_range_size, failure, status_error, uri_encode};

/// API_VERSION is the version of the Blob service API that requests ask for.
const API_VERSION: &str = "2023-11-03";

/// RESOURCE is the resource that storage tokens are for.
const RESOURCE: &str = "https://storage.azure.com";

/// invalid returns the error of invalid credentials or responses.
fn invalid(message: impl Into<String>) -> Error {
    Error::Corrupt(message.into())
}

/// env returns an environment variable that is set and not empty.
fn env(name: &str) -> Option<String> {
    std::env::var(name).ok().filter(|v| !v.is_empty())
}

/// Token is an access token with when it expires.
struct Token {
    value: String,
    expires: Instant,
}

/// AzureBlobstore keeps blobs as block blobs under a prefix of a container of a storage account.
pub(crate) struct AzureBlobstore {
    service: String,
    container: String,
    prefix: String,
    token: Mutex<Option<Token>>,
}

/// open_azure opens the blobstore of an az://account-host/container/path URL.
pub(crate) fn open_azure(rest: &str) -> Result<AzureBlobstore> {
    let (host, path) = rest.split_once('/').unwrap_or((rest, ""));
    let (container, prefix) = path.split_once('/').unwrap_or((path, ""));
    if container.is_empty() {
        return Err(invalid("azure url must include container name in path"));
    }
    Ok(AzureBlobstore {
        service: format!("https://{host}"),
        container: container.to_string(),
        prefix: prefix.trim_start_matches('/').to_string(),
        token: Mutex::new(None),
    })
}

/// authority returns the Microsoft Entra authority host, which AZURE_AUTHORITY_HOST overrides.
fn authority() -> String {
    env("AZURE_AUTHORITY_HOST")
        .unwrap_or_else(|| "https://login.microsoftonline.com".into())
        .trim_end_matches('/')
        .to_string()
}

/// client_token requests a token for an application from the tenant's token endpoint.
fn client_token(tenant: &str, form: &[(&str, &str)]) -> Result<Json> {
    let mut form = form.to_vec();
    let scope = format!("{RESOURCE}/.default");
    form.push(("scope", &scope));
    form.push(("grant_type", "client_credentials"));
    token_response(client()?.post(format!("{}/{tenant}/oauth2/v2.0/token", authority())).form(&form))
}

/// token_response sends a token request and returns its JSON.
fn token_response(request: RequestBuilder) -> Result<Json> {
    let response = request.send().map_err(failure)?;
    if !response.status().is_success() {
        return Err(status_error("token request", response));
    }
    response.json().map_err(failure)
}

/// default_token finds a token as DefaultAzureCredential does: a service principal's secret from the environment, a
/// workload identity, a managed identity, and then the Azure CLI's login.
fn default_token() -> Result<(String, u64)> {
    let json = if let (Some(tenant), Some(id), Some(secret)) =
        (env("AZURE_TENANT_ID"), env("AZURE_CLIENT_ID"), env("AZURE_CLIENT_SECRET"))
    {
        client_token(&tenant, &[("client_id", &id), ("client_secret", &secret)])?
    } else if let (Some(tenant), Some(id), Some(file)) =
        (env("AZURE_TENANT_ID"), env("AZURE_CLIENT_ID"), env("AZURE_FEDERATED_TOKEN_FILE"))
    {
        let assertion = std::fs::read_to_string(file)?;
        let assertion_type = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer";
        client_token(
            &tenant,
            &[("client_id", &id), ("client_assertion", assertion.trim()), ("client_assertion_type", assertion_type)],
        )?
    } else if let Some(token) = managed_identity_token() {
        token?
    } else {
        let output = std::process::Command::new("az")
            .args(["account", "get-access-token", "--output", "json", "--resource", RESOURCE])
            .output()
            .map_err(|e| invalid(format!("DefaultAzureCredential: failed to acquire a token: {e}")))?;
        if !output.status.success() {
            return Err(invalid(format!(
                "DefaultAzureCredential: failed to acquire a token: {}",
                String::from_utf8_lossy(&output.stderr).trim()
            )));
        }
        let json: Json = serde_json::from_slice(&output.stdout).map_err(|e| invalid(format!("az: {e}")))?;
        return Ok((json["accessToken"].as_str().unwrap_or_default().to_string(), 3000));
    };
    let value = json["access_token"].as_str().ok_or_else(|| invalid("token response without access_token"))?;
    let lifetime = match &json["expires_in"] {
        Json::String(s) => s.parse().unwrap_or(3600),
        other => other.as_u64().unwrap_or(3600),
    };
    Ok((value.to_string(), lifetime))
}

/// managed_identity_token asks App Service's identity endpoint or the instance metadata service for a managed
/// identity's token, or returns None when neither answers.
fn managed_identity_token() -> Option<Result<Json>> {
    let client_id =
        env("AZURE_CLIENT_ID").map(|id| format!("&client_id={}", uri_encode(&id, false))).unwrap_or_default();
    if let (Some(endpoint), Some(header)) = (env("IDENTITY_ENDPOINT"), env("IDENTITY_HEADER")) {
        let url = format!("{endpoint}?api-version=2019-08-01&resource={RESOURCE}{client_id}");
        return Some(client().and_then(|c| token_response(c.get(url).header("X-IDENTITY-HEADER", header))));
    }
    let probe = reqwest::blocking::Client::builder().timeout(Duration::from_secs(1)).build().ok()?;
    let url = format!(
        "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource={RESOURCE}{client_id}"
    );
    let response = probe.get(url).header("Metadata", "true").send().ok()?;
    if !response.status().is_success() {
        return None;
    }
    Some(response.json().map_err(failure))
}

impl AzureBlobstore {
    /// token returns an access token, fetching a new one when the last expired.
    fn token(&self) -> Result<String> {
        let mut token = self.token.lock().unwrap_or_else(|p| p.into_inner());
        if token.as_ref().is_none_or(|t| t.expires <= Instant::now()) {
            let (value, lifetime) = default_token()?;
            let expires = Instant::now() + Duration::from_secs(lifetime.saturating_sub(60));
            *token = Some(Token { value, expires });
        }
        Ok(token.as_ref().expect("a token").value.clone())
    }

    /// url returns the URL of a blob.
    fn url(&self, key: &str) -> String {
        let name = if self.prefix.is_empty() {
            key.to_string()
        } else {
            format!("{}/{key}", self.prefix.trim_end_matches('/'))
        };
        format!("{}/{}/{}", self.service, self.container, uri_encode(&name, true))
    }

    /// send sends a request with the API version and a token.
    fn send(&self, request: RequestBuilder) -> Result<Response> {
        request.header("x-ms-version", API_VERSION).bearer_auth(self.token()?).send().map_err(failure)
    }

    /// properties returns a blob's size and ETag, or None when the container lacks it.
    fn properties(&self, key: &str) -> Result<Option<(u64, String)>> {
        let response = self.send(client()?.head(self.url(key)))?;
        match response.status().as_u16() {
            200 => Ok(Some((response.content_length().unwrap_or(0), etag(&response)))),
            404 => Ok(None),
            _ => Err(status_error("GetProperties", response)),
        }
    }

    /// put_blob writes a block blob with the conditions given as headers, returning the response.
    fn put_blob(&self, key: &str, headers: &[(&str, &str)], data: &[u8]) -> Result<Response> {
        let mut request = client()?.put(self.url(key)).header("x-ms-blob-type", "BlockBlob").body(data.to_vec());
        for (name, value) in headers {
            request = request.header(*name, *value);
        }
        self.send(request)
    }
}

/// etag returns a response's ETag.
fn etag(response: &Response) -> String {
    response.headers().get("etag").and_then(|v| v.to_str().ok()).unwrap_or_default().to_string()
}

/// block_id returns the ID of the block at an index, as Dolt names blocks.
fn block_id(index: usize) -> String {
    base64::engine::general_purpose::STANDARD.encode(format!("{index:064}"))
}

impl Blobstore for AzureBlobstore {
    fn path(&self) -> String {
        format!("{}/{}", self.container, self.prefix).trim_end_matches('/').to_string()
    }

    fn exists(&self, key: &str) -> Result<bool> {
        Ok(self.properties(key)?.is_some())
    }

    fn get(&self, key: &str, range: BlobRange) -> Result<Blob> {
        let missing = || not_found(&format!("{}/{key}", self.path()));
        let mut size = 0;
        let mut request = client()?.get(self.url(key));
        if !range.is_all() {
            let range = if range.offset < 0 {
                let (blob_size, _) = self.properties(key)?.ok_or_else(missing)?;
                size = blob_size;
                range.positive(blob_size as i64)
            } else {
                range
            };
            let header = if range.length == 0 {
                format!("bytes={}-", range.offset)
            } else {
                format!("bytes={}-{}", range.offset, range.offset + range.length - 1)
            };
            request = request.header("x-ms-range", header);
        }
        let response = self.send(request)?;
        match response.status().as_u16() {
            200 | 206 => {}
            404 => return Err(missing()),
            _ => return Err(status_error("Download", response)),
        }
        if size == 0 {
            size = content_range_size(&response);
        }
        if size == 0 && range.is_all() {
            size = response.content_length().unwrap_or(0);
        }
        let version = etag(&response);
        Ok(Blob { data: response.bytes().map_err(failure)?.to_vec(), size, version })
    }

    fn put(&self, key: &str, data: &[u8]) -> Result<String> {
        let response = self.put_blob(key, &[], data)?;
        if !response.status().is_success() {
            return Err(status_error("PutBlob", response));
        }
        Ok(etag(&response))
    }

    fn check_and_put_manifest(&self, expected: &str, data: &[u8]) -> Result<String> {
        let condition = if expected.is_empty() { ("If-None-Match", "*") } else { ("If-Match", expected) };
        let response = self.put_blob(MANIFEST_KEY, &[condition], data)?;
        let status = response.status().as_u16();
        if (expected.is_empty() && status == 409) || (!expected.is_empty() && status == 412) {
            let actual = self.properties(MANIFEST_KEY).ok().flatten().map_or("unknown".into(), |(_, etag)| etag);
            return Err(Error::VersionMismatch { key: MANIFEST_KEY.into(), expected: expected.into(), actual });
        }
        if !response.status().is_success() {
            return Err(status_error("PutBlob", response));
        }
        Ok(etag(&response))
    }

    fn concatenate(&self, key: &str, sources: &[String]) -> Result<String> {
        let token = self.token()?;
        let mut blocks = String::from("<?xml version=\"1.0\" encoding=\"utf-8\"?><BlockList>");
        for (i, source) in sources.iter().enumerate() {
            let id = block_id(i);
            let url = format!("{}?comp=block&blockid={}", self.url(key), uri_encode(&id, false));
            let request = client()?
                .put(url)
                .header("x-ms-copy-source", self.url(source))
                .header("x-ms-copy-source-authorization", format!("Bearer {token}"))
                .header("Content-Length", "0");
            let response = self.send(request)?;
            if !response.status().is_success() {
                return Err(status_error(&format!("failed to stage block from URL for source {source}"), response));
            }
            blocks.push_str(&format!("<Latest>{id}</Latest>"));
        }
        blocks.push_str("</BlockList>");
        let response = self.send(client()?.put(format!("{}?comp=blocklist", self.url(key))).body(blocks))?;
        if !response.status().is_success() {
            return Err(status_error("failed to commit block list", response));
        }
        Ok(etag(&response))
    }
}
