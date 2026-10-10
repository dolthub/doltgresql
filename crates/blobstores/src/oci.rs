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

//! Oracle Cloud Infrastructure Object Storage, with requests signed by the API key of the OCI config file, as Dolt's
//! OCIBlobstore uses it for oci:// URLs.

use std::collections::BTreeMap;
use std::path::PathBuf;

use base64::Engine;
use reqwest::Method;
use reqwest::blocking::Response;
use serde_json::Value as Json;
use store::{Blob, BlobRange, Blobstore, Error, MANIFEST_KEY, Result, not_found};

use crate::http::{client, content_range_size, failure, status_error, uri_encode};

/// invalid returns the error of an invalid configuration or response.
fn invalid(message: impl Into<String>) -> Error {
    Error::Corrupt(message.into())
}

/// env returns an environment variable that is set and not empty.
fn env(name: &str) -> Option<String> {
    std::env::var(name).ok().filter(|v| !v.is_empty())
}

/// Signer signs requests with a user's API key.
struct Signer {
    key_id: String,
    key: aws_lc_rs::signature::RsaKeyPair,
}

/// OciBlobstore keeps blobs as objects under a prefix of an Object Storage bucket in the tenancy's namespace.
pub(crate) struct OciBlobstore {
    signer: Signer,
    endpoint: String,
    namespace: String,
    bucket: String,
    prefix: String,
}

/// expand_home replaces a leading `~` of a path with the home directory.
fn expand_home(path: &str) -> PathBuf {
    match path.strip_prefix("~/") {
        Some(rest) => PathBuf::from(env("HOME").unwrap_or_default()).join(rest),
        None => PathBuf::from(path),
    }
}

/// config returns the settings of the OCI config file's profile, which OCI_CONFIG_FILE and OCI_CONFIG_PROFILE choose,
/// or of the TF_VAR_ environment variables.
fn config() -> Result<BTreeMap<String, String>> {
    let path = env("OCI_CONFIG_FILE").map(|p| expand_home(&p)).unwrap_or_else(|| expand_home("~/.oci/config"));
    let profile = env("OCI_CONFIG_PROFILE").unwrap_or_else(|| "DEFAULT".into());
    let mut settings = BTreeMap::new();
    if let Ok(text) = std::fs::read_to_string(&path) {
        let mut current = false;
        for line in text.lines().map(str::trim).filter(|l| !l.is_empty() && !l.starts_with('#')) {
            if let Some(name) = line.strip_prefix('[').and_then(|l| l.strip_suffix(']')) {
                current = name.trim() == profile;
            } else if current && let Some((key, value)) = line.split_once('=') {
                settings.insert(key.trim().to_string(), value.trim().to_string());
            }
        }
    }
    for (setting, variable) in [
        ("tenancy", "TF_VAR_tenancy_ocid"),
        ("user", "TF_VAR_user_ocid"),
        ("fingerprint", "TF_VAR_fingerprint"),
        ("key_file", "TF_VAR_private_key_path"),
        ("region", "TF_VAR_region"),
    ] {
        if let (false, Some(value)) = (settings.contains_key(setting), env(variable)) {
            settings.insert(setting.to_string(), value);
        }
    }
    if let Some(region) = env("OCI_REGION") {
        settings.insert("region".into(), region);
    }
    Ok(settings)
}

/// realm_domain returns the domain of a region's realm, which OCI_REGION_METADATA names for its region, or which
/// OCI_DEFAULT_REALM sets, and otherwise the commercial realm's.
fn realm_domain(region: &str) -> String {
    if let Some(metadata) = env("OCI_REGION_METADATA").and_then(|m| serde_json::from_str::<Json>(&m).ok())
        && (metadata["regionIdentifier"] == region || metadata["regionKey"] == region)
        && let Some(domain) = metadata["realmDomainComponent"].as_str()
    {
        return domain.to_string();
    }
    env("OCI_DEFAULT_REALM").unwrap_or_else(|| "oraclecloud.com".into())
}

/// load_key reads an unencrypted RSA private key in PKCS#8 or PKCS#1 PEM.
fn load_key(path: &PathBuf) -> Result<aws_lc_rs::signature::RsaKeyPair> {
    let pem = std::fs::read_to_string(path).map_err(|e| invalid(format!("{}: {e}", path.display())))?;
    if pem.contains("ENCRYPTED") {
        return Err(invalid(format!("{}: encrypted private keys are not supported", path.display())));
    }
    let body: String = pem.lines().filter(|l| !l.starts_with("-----")).collect();
    let der = base64::engine::general_purpose::STANDARD
        .decode(body.trim())
        .map_err(|e| invalid(format!("private key: {e}")))?;
    let pair = if pem.contains("BEGIN RSA PRIVATE KEY") {
        aws_lc_rs::signature::RsaKeyPair::from_der(&der)
    } else {
        aws_lc_rs::signature::RsaKeyPair::from_pkcs8(&der)
    };
    pair.map_err(|e| invalid(format!("private key: {e}")))
}

/// open_oci opens the blobstore of an oci://bucket/path URL, reading the tenancy's namespace.
pub(crate) fn open_oci(rest: &str) -> Result<OciBlobstore> {
    let (bucket, prefix) = rest.split_once('/').unwrap_or((rest, ""));
    let settings = config()?;
    let setting = |name: &str| {
        settings.get(name).cloned().ok_or_else(|| {
            invalid(format!("can not create client, bad configuration: did not find a proper configuration for {name}"))
        })
    };
    let key_id = format!("{}/{}/{}", setting("tenancy")?, setting("user")?, setting("fingerprint")?);
    let signer = Signer { key_id, key: load_key(&expand_home(&setting("key_file")?))? };
    let region = setting("region")?;
    let endpoint = format!("https://objectstorage.{region}.{}", realm_domain(&region));
    let mut store = OciBlobstore {
        signer,
        endpoint,
        namespace: String::new(),
        bucket: bucket.to_string(),
        prefix: prefix.trim_start_matches('/').to_string(),
    };
    let response = store.send(Method::GET, "/n/", &[], &[])?;
    if !response.status().is_success() {
        return Err(status_error("GetNamespace", response));
    }
    store.namespace = response.json::<Json>().map_err(failure)?.as_str().unwrap_or_default().to_string();
    Ok(store)
}

impl OciBlobstore {
    /// object returns the path of an object.
    fn object(&self, key: &str) -> String {
        let name = if self.prefix.is_empty() {
            key.to_string()
        } else {
            format!("{}/{key}", self.prefix.trim_end_matches('/'))
        };
        format!("/n/{}/b/{}/o/{}", self.namespace, self.bucket, uri_encode(&name, false))
    }

    /// send signs a request with the date, request target, and host, leaving object bodies unsigned as Object Storage
    /// allows, and sends it.
    fn send(&self, method: Method, path: &str, headers: &[(&str, &str)], body: &[u8]) -> Result<Response> {
        let date = chrono::Utc::now().format("%a, %d %b %Y %H:%M:%S GMT").to_string();
        let host = self.endpoint.split_once("://").map_or(self.endpoint.as_str(), |(_, h)| h);
        let signing =
            format!("date: {date}\n(request-target): {} {path}\nhost: {host}", method.as_str().to_lowercase());
        let mut signature = vec![0; self.signer.key.public_modulus_len()];
        self.signer
            .key
            .sign(
                &aws_lc_rs::signature::RSA_PKCS1_SHA256,
                &aws_lc_rs::rand::SystemRandom::new(),
                signing.as_bytes(),
                &mut signature,
            )
            .map_err(|e| invalid(format!("signing: {e}")))?;
        let authorization = format!(
            "Signature version=\"1\",keyId=\"{}\",algorithm=\"rsa-sha256\",headers=\"date (request-target) host\",signature=\"{}\"",
            self.signer.key_id,
            base64::engine::general_purpose::STANDARD.encode(signature)
        );
        let mut request = client()?
            .request(method, format!("{}{path}", self.endpoint))
            .header("date", date)
            .header("authorization", authorization);
        for (name, value) in headers {
            request = request.header(*name, *value);
        }
        request.body(body.to_vec()).send().map_err(failure)
    }

    /// upload writes an object with the conditions given as headers, returning its ETag, or the response that failed.
    fn upload(
        &self,
        key: &str,
        headers: &[(&str, &str)],
        data: &[u8],
    ) -> Result<std::result::Result<String, Response>> {
        let response = self.send(Method::PUT, &self.object(key), headers, data)?;
        if response.status().is_success() {
            return Ok(Ok(etag(&response)));
        }
        Ok(Err(response))
    }
}

/// etag returns a response's ETag.
fn etag(response: &Response) -> String {
    response.headers().get("etag").and_then(|v| v.to_str().ok()).unwrap_or_default().to_string()
}

impl Blobstore for OciBlobstore {
    fn path(&self) -> String {
        format!("{}/{}", self.bucket, self.prefix).trim_end_matches('/').to_string()
    }

    fn exists(&self, key: &str) -> Result<bool> {
        let response = self.send(Method::HEAD, &self.object(key), &[], &[])?;
        match response.status().as_u16() {
            200 => Ok(true),
            404 => Ok(false),
            _ => Err(status_error("HeadObject", response)),
        }
    }

    fn get(&self, key: &str, range: BlobRange) -> Result<Blob> {
        let header = range.http_header();
        let headers: Vec<(&str, &str)> = header.iter().map(|h| ("range", h.as_str())).collect();
        let response = self.send(Method::GET, &self.object(key), &headers, &[])?;
        match response.status().as_u16() {
            200 | 206 => {}
            404 => return Err(not_found(&format!("oci://{}", self.path()))),
            _ => return Err(status_error("GetObject", response)),
        }
        let mut size = content_range_size(&response);
        if size == 0 && range.is_all() {
            size = response.content_length().unwrap_or(0);
        }
        let version = etag(&response);
        let mut data = response.bytes().map_err(failure)?.to_vec();
        if range.offset < 0 && range.length > 0 {
            data.truncate(range.length as usize);
        }
        Ok(Blob { data, size, version })
    }

    fn put(&self, key: &str, data: &[u8]) -> Result<String> {
        if data.is_empty() {
            return Err(invalid("failed to upload to oci blobstore, no data in reader"));
        }
        self.upload(key, &[], data)?.map_err(|r| status_error("PutObject", r))
    }

    fn check_and_put_manifest(&self, expected: &str, data: &[u8]) -> Result<String> {
        let condition = if expected.is_empty() { ("if-none-match", "*") } else { ("if-match", expected) };
        match self.upload(MANIFEST_KEY, &[condition], data)? {
            Ok(version) => Ok(version),
            Err(response) if response.status().as_u16() == 412 => Err(Error::VersionMismatch {
                key: MANIFEST_KEY.into(),
                expected: expected.into(),
                actual: "unknown (Not supported in OCI implementation)".into(),
            }),
            Err(response) => Err(status_error("PutObject", response)),
        }
    }

    fn concatenate(&self, _: &str, _: &[String]) -> Result<String> {
        Err(invalid("concatenate is unimplemented on the oci blobstore"))
    }
}
