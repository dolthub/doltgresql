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

//! Google Cloud Storage, through its JSON API with Application Default Credentials, as Dolt's GCSBlobstore uses it for
//! gs:// URLs.

use std::sync::Mutex;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

use base64::Engine;
use reqwest::blocking::{RequestBuilder, Response};
use serde_json::{Value as Json, json};
use store::{Blob, BlobRange, Blobstore, Error, MANIFEST_KEY, Result, not_found};

use crate::http::{client, content_range_size, failure, status_error, uri_encode};

/// SCOPE is the OAuth scope that reads and writes objects.
const SCOPE: &str = "https://www.googleapis.com/auth/devstorage.read_write";

/// COMPOSE_BATCH is how many objects one compose request joins.
const COMPOSE_BATCH: usize = 32;

/// invalid returns the error of invalid credentials or responses.
fn invalid(message: impl Into<String>) -> Error {
    Error::Corrupt(message.into())
}

/// Token is an OAuth access token with when it expires.
struct Token {
    value: String,
    expires: Instant,
}

/// GcsBlobstore keeps blobs as objects under a prefix of a Cloud Storage bucket.
pub(crate) struct GcsBlobstore {
    base: String,
    bucket: String,
    prefix: String,
    /// The credentials file's contents, or None for the metadata server, unused against an emulator.
    credentials: Option<Json>,
    emulated: bool,
    token: Mutex<Option<Token>>,
}

/// open_gcs opens the blobstore of a gs://bucket/path URL.
pub(crate) fn open_gcs(rest: &str) -> Result<GcsBlobstore> {
    let (bucket, prefix) = rest.split_once('/').unwrap_or((rest, ""));
    let emulator = std::env::var("STORAGE_EMULATOR_HOST").ok().filter(|h| !h.is_empty());
    let (base, credentials) = match &emulator {
        Some(host) if host.contains("://") => (host.trim_end_matches('/').to_string(), None),
        Some(host) => (format!("http://{host}"), None),
        None => ("https://storage.googleapis.com".to_string(), application_credentials()?),
    };
    Ok(GcsBlobstore {
        base,
        bucket: bucket.to_string(),
        prefix: prefix.trim_start_matches('/').to_string(),
        credentials,
        emulated: emulator.is_some(),
        token: Mutex::new(None),
    })
}

/// application_credentials reads the file that GOOGLE_APPLICATION_CREDENTIALS names, or gcloud's application default
/// credentials, or returns None to use the metadata server.
fn application_credentials() -> Result<Option<Json>> {
    let path = match std::env::var("GOOGLE_APPLICATION_CREDENTIALS").ok().filter(|p| !p.is_empty()) {
        Some(path) => std::path::PathBuf::from(path),
        None => {
            let config = std::env::var("CLOUDSDK_CONFIG").map(std::path::PathBuf::from).unwrap_or_else(|_| {
                std::path::PathBuf::from(std::env::var("HOME").unwrap_or_default()).join(".config").join("gcloud")
            });
            let path = config.join("application_default_credentials.json");
            if !path.exists() {
                return Ok(None);
            }
            path
        }
    };
    let text = std::fs::read_to_string(&path)?;
    serde_json::from_str(&text).map(Some).map_err(|e| invalid(format!("{}: {e}", path.display())))
}

/// pem_der decodes the body of a PEM block.
fn pem_der(pem: &str) -> Result<Vec<u8>> {
    let body: String = pem.lines().filter(|l| !l.starts_with("-----")).collect();
    base64::engine::general_purpose::STANDARD.decode(body.trim()).map_err(|e| invalid(format!("private key: {e}")))
}

/// b64url encodes bytes as unpadded URL-safe base64.
fn b64url(bytes: &[u8]) -> String {
    base64::engine::general_purpose::URL_SAFE_NO_PAD.encode(bytes)
}

/// service_account_token exchanges a JWT that a service account's key signs for an access token.
fn service_account_token(credentials: &Json) -> Result<Json> {
    let email = credentials["client_email"].as_str().ok_or_else(|| invalid("service account without client_email"))?;
    let key = credentials["private_key"].as_str().ok_or_else(|| invalid("service account without private_key"))?;
    let token_uri = credentials["token_uri"].as_str().unwrap_or("https://oauth2.googleapis.com/token");
    let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default().as_secs();
    let header = b64url(json!({"alg": "RS256", "typ": "JWT"}).to_string().as_bytes());
    let claims = json!({"iss": email, "scope": SCOPE, "aud": token_uri, "iat": now, "exp": now + 3600});
    let message = format!("{header}.{}", b64url(claims.to_string().as_bytes()));
    let pair = aws_lc_rs::signature::RsaKeyPair::from_pkcs8(&pem_der(key)?)
        .map_err(|e| invalid(format!("private key: {e}")))?;
    let mut signature = vec![0; pair.public_modulus_len()];
    pair.sign(
        &aws_lc_rs::signature::RSA_PKCS1_SHA256,
        &aws_lc_rs::rand::SystemRandom::new(),
        message.as_bytes(),
        &mut signature,
    )
    .map_err(|e| invalid(format!("signing: {e}")))?;
    let assertion = format!("{message}.{}", b64url(&signature));
    let form = [("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer"), ("assertion", assertion.as_str())];
    token_response(client()?.post(token_uri).form(&form))
}

/// user_token exchanges an authorized user's refresh token for an access token.
fn user_token(credentials: &Json) -> Result<Json> {
    let field = |name: &str| credentials[name].as_str().ok_or_else(|| invalid(format!("credentials without {name}")));
    let form = [
        ("grant_type", "refresh_token"),
        ("client_id", field("client_id")?),
        ("client_secret", field("client_secret")?),
        ("refresh_token", field("refresh_token")?),
    ];
    token_response(client()?.post("https://oauth2.googleapis.com/token").form(&form))
}

/// metadata_token reads an access token from the Compute Engine metadata server.
fn metadata_token() -> Result<Json> {
    let host = std::env::var("GCE_METADATA_HOST").unwrap_or_else(|_| "metadata.google.internal".into());
    let url = format!("http://{host}/computeMetadata/v1/instance/service-accounts/default/token?scopes={SCOPE}");
    token_response(client()?.get(url).header("Metadata-Flavor", "Google")).map_err(|err| {
        invalid(format!("google: could not find default credentials, and the metadata server failed: {err}"))
    })
}

/// token_response sends a token request and returns its JSON.
fn token_response(request: RequestBuilder) -> Result<Json> {
    let response = request.send().map_err(failure)?;
    if !response.status().is_success() {
        return Err(status_error("oauth2 token", response));
    }
    response.json().map_err(failure)
}

impl GcsBlobstore {
    /// authorize adds an access token to a request, fetching a new one when the last expired.
    fn authorize(&self, request: RequestBuilder) -> Result<RequestBuilder> {
        if self.emulated {
            return Ok(request);
        }
        let mut token = self.token.lock().unwrap_or_else(|p| p.into_inner());
        if token.as_ref().is_none_or(|t| t.expires <= Instant::now()) {
            let json = match &self.credentials {
                Some(c) if c["type"] == "service_account" => service_account_token(c)?,
                Some(c) if c["type"] == "authorized_user" => user_token(c)?,
                Some(c) => return Err(invalid(format!("unsupported credentials type {}", c["type"]))),
                None => metadata_token()?,
            };
            let value = json["access_token"].as_str().ok_or_else(|| invalid("token response without access_token"))?;
            let lifetime = json["expires_in"].as_u64().unwrap_or(3600).saturating_sub(60);
            *token = Some(Token { value: value.to_string(), expires: Instant::now() + Duration::from_secs(lifetime) });
        }
        Ok(request.bearer_auth(&token.as_ref().expect("a token").value))
    }

    /// name returns an object's name.
    fn name(&self, key: &str) -> String {
        if self.prefix.is_empty() { key.to_string() } else { format!("{}/{key}", self.prefix.trim_end_matches('/')) }
    }

    /// object returns the JSON API URL of an object.
    fn object(&self, key: &str) -> String {
        format!("{}/storage/v1/b/{}/o/{}", self.base, self.bucket, uri_encode(&self.name(key), false))
    }

    /// send sends a request.
    fn send(&self, request: RequestBuilder) -> Result<Response> {
        self.authorize(request)?.send().map_err(failure)
    }

    /// upload writes an object, with a generation it must have, 0 for none, returning its generation.
    fn upload(&self, key: &str, data: &[u8], if_generation: Option<i64>) -> Result<std::result::Result<String, u16>> {
        let mut url = format!(
            "{}/upload/storage/v1/b/{}/o?uploadType=media&name={}",
            self.base,
            self.bucket,
            uri_encode(&self.name(key), false)
        );
        if let Some(generation) = if_generation {
            url.push_str(&format!("&ifGenerationMatch={generation}"));
        }
        let response =
            self.send(client()?.post(url).header("Content-Type", "application/octet-stream").body(data.to_vec()))?;
        match response.status().as_u16() {
            200 => Ok(Ok(generation(&response.json().map_err(failure)?))),
            412 => Ok(Err(412)),
            _ => Err(status_error("upload", response)),
        }
    }

    /// compose joins up to COMPOSE_BATCH objects into one, returning its generation.
    fn compose(&self, key: &str, sources: &[String]) -> Result<String> {
        let objects: Vec<Json> = sources.iter().map(|s| json!({"name": self.name(s)})).collect();
        let body = json!({"sourceObjects": objects, "destination": {"contentType": "application/octet-stream"}});
        let response = self.send(client()?.post(format!("{}/compose", self.object(key))).json(&body))?;
        if !response.status().is_success() {
            return Err(status_error("compose", response));
        }
        Ok(generation(&response.json().map_err(failure)?))
    }
}

/// generation returns an object's generation as Dolt formats it, in hexadecimal.
fn generation(object: &Json) -> String {
    let value = match &object["generation"] {
        Json::String(s) => s.parse::<i64>().unwrap_or(0),
        other => other.as_i64().unwrap_or(0),
    };
    format!("{value:x}")
}

impl Blobstore for GcsBlobstore {
    fn path(&self) -> String {
        format!("{}/{}", self.bucket, self.prefix).trim_end_matches('/').to_string()
    }

    fn exists(&self, key: &str) -> Result<bool> {
        let response = self.send(client()?.get(self.object(key)))?;
        match response.status().as_u16() {
            200 => Ok(true),
            404 => Ok(false),
            _ => Err(status_error("object attributes", response)),
        }
    }

    fn get(&self, key: &str, range: BlobRange) -> Result<Blob> {
        let mut request = client()?.get(format!("{}?alt=media", self.object(key)));
        if let Some(header) = range.http_header() {
            request = request.header("Range", header);
        }
        let response = self.send(request)?;
        match response.status().as_u16() {
            200 | 206 => {}
            404 => return Err(not_found(&format!("gs://{}/{}", self.bucket, self.name(key)))),
            _ => return Err(status_error("read object", response)),
        }
        let header = |name: &str| response.headers().get(name).and_then(|v| v.to_str().ok()).map(str::to_string);
        let mut size = content_range_size(&response);
        if size == 0 {
            size = header("x-goog-stored-content-length").and_then(|s| s.parse().ok()).unwrap_or(0);
        }
        if size == 0 && range.is_all() {
            size = response.content_length().unwrap_or(0);
        }
        let version = match header("x-goog-generation") {
            Some(g) => format!("{:x}", g.parse::<i64>().unwrap_or(0)),
            None => {
                let attributes = self.send(client()?.get(self.object(key)))?;
                generation(&attributes.json().map_err(failure)?)
            }
        };
        let mut data = response.bytes().map_err(failure)?.to_vec();
        if range.offset < 0 && range.length > 0 {
            data.truncate(range.length as usize);
        }
        Ok(Blob { data, size, version })
    }

    fn put(&self, key: &str, data: &[u8]) -> Result<String> {
        self.upload(key, data, None)?.map_err(|_| invalid("unexpected precondition failure"))
    }

    fn check_and_put_manifest(&self, expected: &str, data: &[u8]) -> Result<String> {
        let generation = match expected {
            "" => 0,
            version => {
                i64::from_str_radix(version, 16).map_err(|_| invalid(format!("Invalid expected Version {version}")))?
            }
        };
        self.upload(MANIFEST_KEY, data, Some(generation))?.map_err(|_| Error::VersionMismatch {
            key: MANIFEST_KEY.into(),
            expected: expected.into(),
            actual: "unknown (Not supported in GCS implementation)".into(),
        })
    }

    fn concatenate(&self, key: &str, sources: &[String]) -> Result<String> {
        let mut sources = sources.to_vec();
        let mut round = 0;
        while sources.len() > COMPOSE_BATCH {
            let mut next = Vec::new();
            for (i, batch) in sources.chunks(COMPOSE_BATCH).enumerate() {
                let name = format!("{key}.compose.{round}.{i}");
                self.compose(&name, batch)?;
                next.push(name);
            }
            sources = next;
            round += 1;
        }
        self.compose(key, &sources)
    }
}
