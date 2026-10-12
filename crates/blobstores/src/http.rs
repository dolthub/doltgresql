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

use std::sync::OnceLock;
use std::time::Duration;

use reqwest::blocking::{Client, Response};
use store::{Error, Result};

/// client returns the HTTP client that every blobstore shares, which also trusts the certificates of the file that
/// SSL_CERT_FILE names.
pub(crate) fn client() -> Result<Client> {
    static CLIENT: OnceLock<std::result::Result<Client, String>> = OnceLock::new();
    CLIENT
        .get_or_init(|| {
            let _ = rustls::crypto::aws_lc_rs::default_provider().install_default();
            let mut builder = Client::builder().timeout(Duration::from_secs(300));
            if let Some(file) = std::env::var_os("SSL_CERT_FILE") {
                let pem = std::fs::read(&file).map_err(|e| format!("{}: {e}", file.to_string_lossy()))?;
                for certificate in reqwest::Certificate::from_pem_bundle(&pem).map_err(|e| e.to_string())? {
                    builder = builder.add_root_certificate(certificate);
                }
            }
            builder.build().map_err(|e| e.to_string())
        })
        .clone()
        .map_err(Error::Corrupt)
}

/// failure returns the error of an HTTP request that failed.
pub(crate) fn failure(err: reqwest::Error) -> Error {
    Error::Io(std::io::Error::other(err.to_string()))
}

/// status_error returns the error of a response with an unexpected status, with its body.
pub(crate) fn status_error(what: &str, response: Response) -> Error {
    let status = response.status();
    let body = response.text().unwrap_or_default();
    Error::Io(std::io::Error::other(format!("{what}: {status}: {}", body.trim())))
}

/// hex returns bytes in lower-case hexadecimal.
pub(crate) fn hex(bytes: &[u8]) -> String {
    bytes.iter().map(|b| format!("{b:02x}")).collect()
}

/// content_range_size returns the size of the whole object that a Content-Range header names, or 0 without one.
pub(crate) fn content_range_size(response: &Response) -> u64 {
    response
        .headers()
        .get(reqwest::header::CONTENT_RANGE)
        .and_then(|v| v.to_str().ok())
        .and_then(|v| v.rsplit_once('/'))
        .and_then(|(_, size)| size.parse().ok())
        .unwrap_or(0)
}

/// uri_encode percent-encodes text as AWS and other signers do, keeping slashes when asked.
pub(crate) fn uri_encode(text: &str, keep_slash: bool) -> String {
    let mut out = String::with_capacity(text.len());
    for b in text.bytes() {
        match b {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'~' => out.push(b as char),
            b'/' if keep_slash => out.push('/'),
            _ => out.push_str(&format!("%{b:02X}")),
        }
    }
    out
}
