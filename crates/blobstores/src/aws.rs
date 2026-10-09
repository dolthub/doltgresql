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

//! Amazon S3 and DynamoDB: the credentials of the AWS SDK's default chain, Signature Version 4, the s3:// blobstore,
//! and the aws:// store that keeps its table files in S3 and its manifest in DynamoDB.

use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

use base64::Engine;
use hmac::{Hmac, Mac};
use reqwest::Method;
use reqwest::blocking::Response;
use serde_json::{Value as Json, json};
use sha2::{Digest, Sha256};
use store::{Blob, BlobRange, Blobstore, Error, Hash, MANIFEST_KEY, Manifest, Result, TableSpec, not_found};

use crate::http::{client, content_range_size, failure, hex, status_error, uri_encode};

/// Credentials are an AWS access key, with a session token for temporary ones.
#[derive(Clone, Debug)]
pub(crate) struct Credentials {
    access_key: String,
    secret_key: String,
    token: Option<String>,
}

/// AwsConfig is a region and the credentials that sign requests in it.
#[derive(Clone, Debug)]
pub(crate) struct AwsConfig {
    region: String,
    credentials: Credentials,
}

/// invalid returns the error of an invalid URL or setting.
fn invalid(message: impl Into<String>) -> Error {
    Error::Corrupt(message.into())
}

/// env returns an environment variable that is set and not empty.
fn env(name: &str) -> Option<String> {
    std::env::var(name).ok().filter(|v| !v.is_empty())
}

/// home returns the user's home directory.
fn home() -> PathBuf {
    PathBuf::from(env("HOME").unwrap_or_default())
}

/// ini_section returns the keys of a section of an INI file, as AWS's shared files hold them.
fn ini_section(path: &Path, section: &str) -> Option<BTreeMap<String, String>> {
    let text = std::fs::read_to_string(path).ok()?;
    let mut found = None;
    let mut current = false;
    for line in text.lines().map(str::trim) {
        if line.starts_with('#') || line.starts_with(';') || line.is_empty() {
            continue;
        }
        if let Some(name) = line.strip_prefix('[').and_then(|l| l.strip_suffix(']')) {
            current = name.trim() == section;
            if current {
                found.get_or_insert_with(BTreeMap::new);
            }
            continue;
        }
        if current && let Some((key, value)) = line.split_once('=') {
            found.get_or_insert_with(BTreeMap::new).insert(key.trim().to_lowercase(), value.trim().to_string());
        }
    }
    found
}

/// profile_name returns the profile that a parameter, AWS_PROFILE, or AWS_DEFAULT_PROFILE names, or `default`.
fn profile_name(profile: Option<&str>) -> String {
    profile
        .map(str::to_string)
        .or_else(|| env("AWS_PROFILE"))
        .or_else(|| env("AWS_DEFAULT_PROFILE"))
        .unwrap_or("default".into())
}

/// config_file returns the shared config file.
fn config_file() -> PathBuf {
    env("AWS_CONFIG_FILE").map(PathBuf::from).unwrap_or_else(|| home().join(".aws").join("config"))
}

/// credentials_file returns the shared credentials file.
fn credentials_file() -> PathBuf {
    env("AWS_SHARED_CREDENTIALS_FILE").map(PathBuf::from).unwrap_or_else(|| home().join(".aws").join("credentials"))
}

/// profile_settings returns a profile's settings from the shared credentials file and then the shared config file.
fn profile_settings(profile: &str) -> BTreeMap<String, String> {
    let mut settings = BTreeMap::new();
    let section = if profile == "default" { "default".to_string() } else { format!("profile {profile}") };
    if let Some(config) = ini_section(&config_file(), &section) {
        settings.extend(config);
    }
    if let Some(credentials) = ini_section(&credentials_file(), profile) {
        settings.extend(credentials);
    }
    settings
}

/// env_credentials returns the credentials of AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.
fn env_credentials() -> Option<Credentials> {
    Some(Credentials {
        access_key: env("AWS_ACCESS_KEY_ID")?,
        secret_key: env("AWS_SECRET_ACCESS_KEY")?,
        token: env("AWS_SESSION_TOKEN"),
    })
}

/// file_credentials returns the static credentials of a profile of an INI file.
fn file_credentials(path: &Path, profile: &str) -> Option<Credentials> {
    let section = ini_section(path, profile)?;
    Some(Credentials {
        access_key: section.get("aws_access_key_id")?.clone(),
        secret_key: section.get("aws_secret_access_key")?.clone(),
        token: section.get("aws_session_token").cloned(),
    })
}

/// json_credentials reads credentials from the JSON that a credential process, container endpoint, or instance
/// metadata returns.
fn json_credentials(json: &Json) -> Option<Credentials> {
    let text = |names: &[&str]| names.iter().find_map(|n| json[*n].as_str()).map(str::to_string);
    Some(Credentials {
        access_key: text(&["AccessKeyId"])?,
        secret_key: text(&["SecretAccessKey"])?,
        token: text(&["SessionToken", "Token"]),
    })
}

/// process_credentials runs a profile's credential_process.
fn process_credentials(command: &str) -> Result<Credentials> {
    let output = std::process::Command::new("sh").arg("-c").arg(command).output()?;
    let json: Json = serde_json::from_slice(&output.stdout).map_err(|e| invalid(format!("credential_process: {e}")))?;
    json_credentials(&json).ok_or_else(|| invalid("credential_process returned no credentials"))
}

/// web_identity_credentials exchanges a web identity token for a role's credentials, as the SDK does in Kubernetes.
fn web_identity_credentials(role: &str, token_file: &str, region: &str) -> Result<Credentials> {
    let token = std::fs::read_to_string(token_file)?;
    let session = env("AWS_ROLE_SESSION_NAME").unwrap_or_else(|| "doltgres".into());
    let host = if region.is_empty() { "sts.amazonaws.com".to_string() } else { format!("sts.{region}.amazonaws.com") };
    let url = format!(
        "https://{host}/?Action=AssumeRoleWithWebIdentity&Version=2011-06-15&RoleArn={}&RoleSessionName={}&WebIdentityToken={}",
        uri_encode(role, false),
        uri_encode(&session, false),
        uri_encode(token.trim(), false)
    );
    let response = client()?.get(url).header("Accept", "application/json").send().map_err(failure)?;
    if !response.status().is_success() {
        return Err(status_error("AssumeRoleWithWebIdentity", response));
    }
    let json: Json = response.json().map_err(failure)?;
    json_credentials(&json["AssumeRoleWithWebIdentityResponse"]["AssumeRoleWithWebIdentityResult"]["Credentials"])
        .ok_or_else(|| invalid("AssumeRoleWithWebIdentity returned no credentials"))
}

/// container_credentials reads the credentials that an ECS or EKS container endpoint serves.
fn container_credentials() -> Option<Result<Credentials>> {
    let url = match (env("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"), env("AWS_CONTAINER_CREDENTIALS_FULL_URI")) {
        (Some(relative), _) => format!("http://169.254.170.2{relative}"),
        (None, Some(full)) => full,
        _ => return None,
    };
    Some((|| {
        let mut request = client()?.get(url);
        let token = env("AWS_CONTAINER_AUTHORIZATION_TOKEN")
            .or_else(|| env("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE").and_then(|f| std::fs::read_to_string(f).ok()));
        if let Some(token) = token {
            request = request.header("Authorization", token.trim());
        }
        let json: Json = request.send().map_err(failure)?.json().map_err(failure)?;
        json_credentials(&json).ok_or_else(|| invalid("the container endpoint returned no credentials"))
    })())
}

/// instance_credentials reads an EC2 instance role's credentials from instance metadata, with IMDSv2's session token.
fn instance_credentials() -> Result<Credentials> {
    let base = env("AWS_EC2_METADATA_SERVICE_ENDPOINT").unwrap_or_else(|| "http://169.254.169.254".into());
    let http =
        reqwest::blocking::Client::builder().timeout(std::time::Duration::from_secs(2)).build().map_err(failure)?;
    let token = http
        .put(format!("{base}/latest/api/token"))
        .header("X-aws-ec2-metadata-token-ttl-seconds", "21600")
        .send()
        .and_then(|r| r.text())
        .map_err(failure)?;
    let get = |path: &str| {
        http.get(format!("{base}{path}")).header("X-aws-ec2-metadata-token", &token).send().and_then(|r| r.text())
    };
    let role = get("/latest/meta-data/iam/security-credentials/").map_err(failure)?;
    let role = role.lines().next().unwrap_or_default().to_string();
    let json: Json =
        serde_json::from_str(&get(&format!("/latest/meta-data/iam/security-credentials/{role}")).map_err(failure)?)
            .map_err(|e| invalid(format!("instance metadata: {e}")))?;
    json_credentials(&json).ok_or_else(|| invalid("instance metadata returned no credentials"))
}

/// default_credentials finds credentials as the SDK's default chain does: the environment, the profile's shared files
/// and its credential process or web identity role, a web identity role from the environment, a container endpoint,
/// and instance metadata.
fn default_credentials(profile: &str, region: &str) -> Result<Credentials> {
    if let Some(credentials) = env_credentials() {
        return Ok(credentials);
    }
    let settings = profile_settings(profile);
    if let (Some(access_key), Some(secret_key)) =
        (settings.get("aws_access_key_id"), settings.get("aws_secret_access_key"))
    {
        let token = settings.get("aws_session_token").cloned();
        return Ok(Credentials { access_key: access_key.clone(), secret_key: secret_key.clone(), token });
    }
    if let Some(command) = settings.get("credential_process") {
        return process_credentials(command);
    }
    if let (Some(role), Some(file)) = (settings.get("role_arn"), settings.get("web_identity_token_file")) {
        return web_identity_credentials(role, file, region);
    }
    if let (Some(role), Some(file)) = (env("AWS_ROLE_ARN"), env("AWS_WEB_IDENTITY_TOKEN_FILE")) {
        return web_identity_credentials(&role, &file, region);
    }
    if let Some(credentials) = container_credentials() {
        return credentials;
    }
    instance_credentials()
        .map_err(|err| invalid(format!("failed to refresh cached credentials, no EC2 IMDS role found, {err}")))
}

/// region returns the region that a setting, AWS_REGION, AWS_DEFAULT_REGION, or the profile names, or us-east-1.
fn region(setting: Option<&str>, profile: &str) -> String {
    setting
        .map(str::to_string)
        .or_else(|| env("AWS_REGION"))
        .or_else(|| env("AWS_DEFAULT_REGION"))
        .or_else(|| profile_settings(profile).get("region").cloned())
        .unwrap_or_else(|| "us-east-1".into())
}

/// config_from_params returns the region and credentials of an aws:// remote's parameters, as Dolt's
/// awsConfigFromParams chooses them: from the default chain, the environment, or a credentials file.
pub(crate) fn config_from_params(params: &BTreeMap<String, String>) -> Result<AwsConfig> {
    let profile = profile_name(params.get("aws-creds-profile").map(String::as_str));
    let region = region(params.get("aws-region").map(String::as_str), &profile);
    let file = params.get("aws-creds-file").filter(|f| !f.is_empty());
    let mut source = match params.get("aws-creds-type").map(|t| t.trim().to_lowercase()) {
        None => "role".to_string(),
        Some(t) if t.is_empty() || t == "auto" => "auto".to_string(),
        Some(t) if ["role", "env", "file"].contains(&t.as_str()) => t,
        Some(_) => return Err(invalid("invalid value for aws-creds-source")),
    };
    if file.is_some() && source == "role" {
        source = "file".into();
    }
    let credentials = match source.as_str() {
        "env" => env_credentials().ok_or_else(|| {
            invalid("error loading env creds; did not find AWS_ACCESS_KEY_ID or AWS_SECRET_ACCESS_KEY environment variable.")
        })?,
        "file" => {
            let file = file.ok_or_else(|| invalid("no such file or directory"))?;
            file_credentials(Path::new(file), &profile).ok_or_else(|| invalid(format!("no credentials in {file}")))?
        }
        "auto" => match env_credentials().or_else(|| file.and_then(|f| file_credentials(Path::new(f), &profile))) {
            Some(credentials) => credentials,
            None => default_credentials(&profile, &region)?,
        },
        _ => default_credentials(&profile, &region)?,
    };
    Ok(AwsConfig { region, credentials })
}

/// default_config returns the region and credentials of the SDK's default chain, with a region that overrides it.
pub(crate) fn default_config(region_setting: Option<&str>) -> Result<AwsConfig> {
    let profile = profile_name(None);
    let region = region(region_setting, &profile);
    Ok(AwsConfig { credentials: default_credentials(&profile, &region)?, region })
}

/// endpoint returns the endpoint that AWS_ENDPOINT_URL_<SERVICE> or AWS_ENDPOINT_URL overrides a service's with.
fn endpoint_override(service: &str) -> Option<String> {
    env(&format!("AWS_ENDPOINT_URL_{}", service.to_uppercase())).or_else(|| env("AWS_ENDPOINT_URL"))
}

/// hmac_sha256 returns the HMAC-SHA256 of data.
fn hmac_sha256(key: &[u8], data: &[u8]) -> Vec<u8> {
    let mut mac = Hmac::<Sha256>::new_from_slice(key).expect("HMAC takes keys of any length");
    mac.update(data);
    mac.finalize().into_bytes().to_vec()
}

/// Request is an AWS request to sign and send.
struct Request<'a> {
    method: Method,
    /// The URL's scheme and authority, such as https://s3.us-east-1.amazonaws.com.
    base: String,
    /// The path, already encoded.
    path: String,
    query: Vec<(String, String)>,
    headers: Vec<(String, String)>,
    body: &'a [u8],
}

impl AwsConfig {
    /// send signs a request for a service with Signature Version 4 and sends it.
    fn send(&self, service: &str, request: Request<'_>) -> Result<Response> {
        let now = chrono::Utc::now();
        let (amz_date, date) = (now.format("%Y%m%dT%H%M%SZ").to_string(), now.format("%Y%m%d").to_string());
        let host = request.base.split_once("://").map_or(request.base.as_str(), |(_, h)| h).trim_end_matches('/');
        let payload = hex(&Sha256::digest(request.body));
        let mut headers = request.headers;
        headers.push(("host".into(), host.to_string()));
        headers.push(("x-amz-date".into(), amz_date.clone()));
        headers.push(("x-amz-content-sha256".into(), payload.clone()));
        if let Some(token) = &self.credentials.token {
            headers.push(("x-amz-security-token".into(), token.clone()));
        }
        let mut canonical_headers: Vec<(String, String)> =
            headers.iter().map(|(k, v)| (k.to_lowercase(), v.trim().to_string())).collect();
        canonical_headers.sort();
        let signed: Vec<&str> = canonical_headers.iter().map(|(k, _)| k.as_str()).collect();
        let signed = signed.join(";");
        let mut query: Vec<(String, String)> =
            request.query.iter().map(|(k, v)| (uri_encode(k, false), uri_encode(v, false))).collect();
        query.sort();
        let query = query.iter().map(|(k, v)| format!("{k}={v}")).collect::<Vec<_>>().join("&");
        let canonical = format!(
            "{}\n{}\n{query}\n{}\n{signed}\n{payload}",
            request.method,
            request.path,
            canonical_headers.iter().map(|(k, v)| format!("{k}:{v}\n")).collect::<String>()
        );
        let scope = format!("{date}/{}/{service}/aws4_request", self.region);
        let to_sign = format!("AWS4-HMAC-SHA256\n{amz_date}\n{scope}\n{}", hex(&Sha256::digest(canonical.as_bytes())));
        let mut key = hmac_sha256(format!("AWS4{}", self.credentials.secret_key).as_bytes(), date.as_bytes());
        for part in [self.region.as_str(), service, "aws4_request"] {
            key = hmac_sha256(&key, part.as_bytes());
        }
        let signature = hex(&hmac_sha256(&key, to_sign.as_bytes()));
        let authorization = format!(
            "AWS4-HMAC-SHA256 Credential={}/{scope}, SignedHeaders={signed}, Signature={signature}",
            self.credentials.access_key
        );
        let url = if query.is_empty() {
            format!("{}{}", request.base.trim_end_matches('/'), request.path)
        } else {
            format!("{}{}?{query}", request.base.trim_end_matches('/'), request.path)
        };
        let mut builder = client()?.request(request.method, url).header("Authorization", authorization);
        for (k, v) in headers.iter().filter(|(k, _)| k != "host") {
            builder = builder.header(k, v);
        }
        builder.body(request.body.to_vec()).send().map_err(failure)
    }
}

/// S3 is a bucket of S3 or of a service compatible with it.
#[derive(Clone)]
pub(crate) struct S3 {
    config: AwsConfig,
    /// The endpoint that replaces AWS's, from the URL or AWS_ENDPOINT_URL_S3.
    endpoint: Option<String>,
    path_style: bool,
    bucket: String,
}

impl S3 {
    /// object returns the base URL and the encoded path of an object.
    fn object(&self, key: &str) -> (String, String) {
        let key = uri_encode(key, true);
        match &self.endpoint {
            Some(endpoint) if self.path_style => (endpoint.clone(), format!("/{}/{key}", self.bucket)),
            Some(endpoint) => {
                let (scheme, host) = endpoint.split_once("://").unwrap_or(("https", endpoint));
                (format!("{scheme}://{}.{host}", self.bucket), format!("/{key}"))
            }
            None if self.path_style => {
                (format!("https://s3.{}.amazonaws.com", self.config.region), format!("/{}/{key}", self.bucket))
            }
            None => (format!("https://{}.s3.{}.amazonaws.com", self.bucket, self.config.region), format!("/{key}")),
        }
    }

    /// send sends a request about an object.
    fn send(&self, method: Method, key: &str, headers: Vec<(String, String)>, body: &[u8]) -> Result<Response> {
        let (base, path) = self.object(key);
        self.config.send("s3", Request { method, base, path, query: Vec::new(), headers, body })
    }

    /// exists reports whether the bucket holds an object.
    fn exists(&self, key: &str) -> Result<bool> {
        let response = self.send(Method::HEAD, key, Vec::new(), &[])?;
        match response.status().as_u16() {
            200 => Ok(true),
            404 => Ok(false),
            _ => Err(status_error("HeadObject", response)),
        }
    }

    /// get reads a range of an object.
    fn get(&self, key: &str, range: BlobRange) -> Result<Blob> {
        let headers = range.http_header().map(|r| vec![("range".to_string(), r)]).unwrap_or_default();
        let response = self.send(Method::GET, key, headers, &[])?;
        match response.status().as_u16() {
            200 | 206 => {}
            404 => return Err(not_found(&format!("s3://{}/{key}", self.bucket))),
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

    /// put writes an object, returning its ETag.
    fn put(
        &self,
        key: &str,
        headers: Vec<(String, String)>,
        data: &[u8],
    ) -> Result<std::result::Result<String, Response>> {
        let response = self.send(Method::PUT, key, headers, data)?;
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

/// S3Blobstore keeps blobs as objects under a prefix of an S3 bucket, as Dolt's S3Blobstore does for s3:// URLs.
pub(crate) struct S3Blobstore {
    s3: S3,
    prefix: String,
}

/// join joins a prefix and a key as Go's path.Join does for them.
fn join(prefix: &str, key: &str) -> String {
    let joined = format!("{}/{key}", prefix.trim_matches('/'));
    joined.trim_start_matches('/').to_string()
}

impl Blobstore for S3Blobstore {
    fn path(&self) -> String {
        join(&self.s3.bucket, &self.prefix)
    }

    fn exists(&self, key: &str) -> Result<bool> {
        self.s3.exists(&join(&self.prefix, key))
    }

    fn get(&self, key: &str, range: BlobRange) -> Result<Blob> {
        self.s3.get(&join(&self.prefix, key), range)
    }

    fn put(&self, key: &str, data: &[u8]) -> Result<String> {
        self.s3.put(&join(&self.prefix, key), Vec::new(), data)?.map_err(|r| status_error("PutObject", r))
    }

    fn check_and_put_manifest(&self, expected: &str, data: &[u8]) -> Result<String> {
        let condition = match expected {
            "" => ("if-none-match".to_string(), "*".to_string()),
            version => ("if-match".to_string(), version.to_string()),
        };
        match self.s3.put(&join(&self.prefix, MANIFEST_KEY), vec![condition], data)? {
            Ok(version) => Ok(version),
            Err(response) => {
                let status = response.status().as_u16();
                let body = response.text().unwrap_or_default();
                if status == 412
                    || body.contains("ConditionalRequestConflict")
                    || (status == 404 && !expected.is_empty())
                {
                    let actual = if status == 404 { String::new() } else { "unknown".into() };
                    return Err(Error::VersionMismatch { key: MANIFEST_KEY.into(), expected: expected.into(), actual });
                }
                Err(Error::Io(std::io::Error::other(format!("PutObject: {status}: {}", body.trim()))))
            }
        }
    }

    fn concatenate(&self, _: &str, _: &[String]) -> Result<String> {
        Err(invalid("concatenate is unimplemented on the s3 blobstore"))
    }
}

/// open_s3 opens the blobstore of an s3://bucket/path URL, whose query may name an endpoint, a region, and path-style
/// addressing, as Dolt's S3Factory parses it.
pub(crate) fn open_s3(rest: &str) -> Result<S3Blobstore> {
    let (location, query) = rest.split_once('?').unwrap_or((rest, ""));
    let (bucket, prefix) = location.split_once('/').unwrap_or((location, ""));
    if bucket.is_empty() {
        return Err(invalid("s3 url must be of the form s3://bucket/path"));
    }
    if bucket.contains('@') {
        return Err(invalid(
            "s3 urls must not embed credentials: they would be stored in plaintext with the remote. Use the standard AWS credential chain instead",
        ));
    }
    let (mut endpoint, mut region, mut path_style) = (None, None, false);
    for pair in query.split('&').filter(|p| !p.is_empty()) {
        let (key, value) = pair.split_once('=').unwrap_or((pair, ""));
        if value.is_empty() {
            return Err(invalid(format!("s3 url parameter {key:?} needs exactly one non-empty value")));
        }
        let value = crate::percent_decode(value);
        match key {
            "endpoint" => endpoint = Some(value),
            "region" => region = Some(value),
            "path-style" => {
                path_style = match value.to_lowercase().as_str() {
                    "1" | "t" | "true" => true,
                    "0" | "f" | "false" => false,
                    _ => return Err(invalid(format!("s3 url parameter {key:?} must be true or false, got {value:?}"))),
                }
            }
            _ => {
                return Err(invalid(format!(
                    "unknown s3 url parameter {key:?}; supported parameters are \"endpoint\", \"region\" and \"path-style\""
                )));
            }
        }
    }
    let config = default_config(region.as_deref())?;
    let endpoint = endpoint.or_else(|| endpoint_override("s3")).map(|e| e.trim_end_matches('/').to_string());
    let s3 = S3 { config, endpoint, path_style, bucket: bucket.to_string() };
    Ok(S3Blobstore { s3, prefix: format!("/{prefix}").trim_start_matches('/').to_string() })
}

/// AWS_STORAGE_VERSION is the version of the DynamoDB manifest items that Dolt writes.
const AWS_STORAGE_VERSION: &str = "4";

/// AwsBlobstore keeps a store's table files as objects of an S3 bucket under the database's name, and its manifest as
/// an item of a DynamoDB table keyed by the database's name, whose lock is its version, as Dolt's AWS store does.
pub(crate) struct AwsBlobstore {
    s3: S3,
    table: String,
    db: String,
    dynamo_endpoint: String,
}

impl AwsBlobstore {
    /// dynamo calls a DynamoDB operation.
    fn dynamo(&self, operation: &str, body: &Json) -> Result<Response> {
        let body = body.to_string();
        let headers = vec![
            ("content-type".to_string(), "application/x-amz-json-1.0".to_string()),
            ("x-amz-target".to_string(), format!("DynamoDB_20120810.{operation}")),
        ];
        let request = Request {
            method: Method::POST,
            base: self.dynamo_endpoint.clone(),
            path: "/".into(),
            query: Vec::new(),
            headers,
            body: body.as_bytes(),
        };
        self.s3.config.send("dynamodb", request)
    }

    /// key returns the S3 key of a blob.
    fn key(&self, key: &str) -> String {
        format!("{}/{key}", self.db)
    }

    /// manifest reads the manifest item, returning it with its lock.
    fn manifest(&self) -> Result<Option<(Manifest, Hash)>> {
        let body = json!({"TableName": self.table, "ConsistentRead": true, "Key": {"db": {"S": self.db}}});
        let response = self.dynamo("GetItem", &body)?;
        if !response.status().is_success() {
            let status = response.status();
            let text = response.text().unwrap_or_default();
            return Err(invalid(format!("failed to get dynamo table: '{}' - {status}: {}", self.table, text.trim())));
        }
        let json: Json = response.json().map_err(failure)?;
        let item = &json["Item"];
        if !item.is_object() {
            return Ok(None);
        }
        let corrupt = || invalid("manifest corrupt");
        let text = |name: &str| item[name]["S"].as_str().map(str::to_string);
        let bytes = |name: &str| -> Option<Hash> {
            let data = base64::engine::general_purpose::STANDARD.decode(item[name]["B"].as_str()?).ok()?;
            Some(Hash(data.try_into().ok()?))
        };
        if text("nbsVers").as_deref() != Some(AWS_STORAGE_VERSION) {
            return Err(corrupt());
        }
        let (format, root, lock) =
            (text("vers").ok_or_else(corrupt)?, bytes("root").ok_or_else(corrupt)?, bytes("lck").ok_or_else(corrupt)?);
        let mut specs = Vec::new();
        if let Some(listed) = text("specs") {
            let fields: Vec<&str> = listed.split(':').collect();
            for pair in fields.chunks(2) {
                let [name, count] = pair else { return Err(corrupt()) };
                let name = Hash::parse(name).ok_or_else(corrupt)?;
                specs.push(TableSpec { name, chunk_count: count.parse().map_err(|_| corrupt())? });
            }
        }
        let manifest = Manifest { version: "5".into(), format, lock, root, gc_gen: Hash::default(), specs };
        Ok(Some((manifest, lock)))
    }
}

impl Blobstore for AwsBlobstore {
    fn path(&self) -> String {
        format!("{}{}", self.table, self.db)
    }

    fn exists(&self, key: &str) -> Result<bool> {
        match key {
            MANIFEST_KEY => Ok(self.manifest()?.is_some()),
            key => self.s3.exists(&self.key(key)),
        }
    }

    fn get(&self, key: &str, range: BlobRange) -> Result<Blob> {
        if key != MANIFEST_KEY {
            return self.s3.get(&self.key(key), range);
        }
        let (manifest, lock) = self.manifest()?.ok_or_else(|| not_found(MANIFEST_KEY))?;
        let data = manifest.format().into_bytes();
        Ok(Blob { size: data.len() as u64, data, version: lock.to_string() })
    }

    fn put(&self, key: &str, data: &[u8]) -> Result<String> {
        self.s3.put(&self.key(key), Vec::new(), data)?.map_err(|r| status_error("PutObject", r))
    }

    fn check_and_put_manifest(&self, expected: &str, data: &[u8]) -> Result<String> {
        let manifest = Manifest::parse(data)?;
        let b64 = |hash: &Hash| base64::engine::general_purpose::STANDARD.encode(hash.0);
        let mut item = json!({
            "db": {"S": self.db},
            "nbsVers": {"S": AWS_STORAGE_VERSION},
            "vers": {"S": manifest.format},
            "root": {"B": b64(&manifest.root)},
            "lck": {"B": b64(&manifest.lock)},
        });
        if !manifest.specs.is_empty() {
            let specs: Vec<String> =
                manifest.specs.iter().flat_map(|s| [s.name.to_string(), s.chunk_count.to_string()]).collect();
            item["specs"] = json!({"S": specs.join(":")});
        }
        let previous = match expected {
            "" => Hash::default(),
            lock => Hash::parse(lock).ok_or_else(|| invalid(format!("invalid manifest version {lock}")))?,
        };
        let condition = "(lck = :prev) and (vers = :vers)";
        let condition = if expected.is_empty() {
            format!("attribute_not_exists(lck) or {condition}")
        } else {
            condition.to_string()
        };
        let body = json!({
            "TableName": self.table,
            "Item": item,
            "ConditionExpression": condition,
            "ExpressionAttributeValues": {":prev": {"B": b64(&previous)}, ":vers": {"S": manifest.format}},
        });
        let response = self.dynamo("PutItem", &body)?;
        if response.status().is_success() {
            return Ok(manifest.lock.to_string());
        }
        let status = response.status();
        let text = response.text().unwrap_or_default();
        if text.contains("ConditionalCheckFailedException") {
            return Err(Error::VersionMismatch {
                key: MANIFEST_KEY.into(),
                expected: expected.into(),
                actual: "unknown".into(),
            });
        }
        Err(Error::Io(std::io::Error::other(format!("PutItem: {status}: {}", text.trim()))))
    }

    fn concatenate(&self, _: &str, _: &[String]) -> Result<String> {
        Err(invalid("concatenate is unimplemented on the aws store"))
    }
}

/// open_aws opens the store of an aws://[table:bucket]/database URL with a remote's parameters.
pub(crate) fn open_aws(rest: &str, params: &BTreeMap<String, String>) -> Result<AwsBlobstore> {
    let (host, path) = rest.split_once('/').unwrap_or((rest, ""));
    let host = host.trim_start_matches('[').trim_end_matches(']');
    let Some((table, bucket)) = host.split_once(':') else { return Err(invalid("aws url has an invalid format")) };
    let db = path.trim_matches('/');
    if db.is_empty() {
        return Err(invalid("invalid database name"));
    }
    let config = config_from_params(params)?;
    let dynamo_endpoint =
        endpoint_override("dynamodb").unwrap_or_else(|| format!("https://dynamodb.{}.amazonaws.com", config.region));
    let endpoint = endpoint_override("s3").map(|e| e.trim_end_matches('/').to_string());
    let path_style = endpoint.is_some();
    let s3 = S3 { config, endpoint, path_style, bucket: bucket.to_string() };
    Ok(AwsBlobstore { s3, table: table.to_string(), db: db.to_string(), dynamo_endpoint })
}
