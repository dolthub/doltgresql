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

//! Sealed URLs: the table file URLs the server hands out carry a signature and an expiry, so that its HTTP handler
//! serves only the files and ranges it offered, as Dolt's singleSymmetricKeySealer does with its own encoding.

use std::time::{Duration, SystemTime, UNIX_EPOCH};

use base64::Engine;
use base64::engine::general_purpose::URL_SAFE_NO_PAD;
use hmac::{Hmac, Mac};
use rand::RngCore;
use sha2::Sha256;

/// PREFIX begins the path of every sealed URL.
const PREFIX: &str = "/sealed";

/// LIFETIME is how long a sealed URL works, as Dolt's sealer allows.
const LIFETIME: Duration = Duration::from_secs(15 * 60);

/// Sealer signs URLs with a key that only this server process knows.
pub struct Sealer {
    key: [u8; 32],
}

impl Sealer {
    /// new returns a sealer with a random key.
    pub fn new() -> Sealer {
        let mut key = [0; 32];
        rand::thread_rng().fill_bytes(&mut key);
        Sealer { key }
    }

    /// mac returns the message authentication code over a path, query, and expiry, ready to sign or verify.
    fn mac(&self, path: &str, query: &str, expiry: u128) -> Hmac<Sha256> {
        let mut mac = Hmac::<Sha256>::new_from_slice(&self.key).expect("HMAC takes keys of any length");
        mac.update(format!("{path}?{query}&exp={expiry}").as_bytes());
        mac
    }

    /// seal returns the path and query of a URL with its signature and expiry added.
    pub fn seal(&self, path: &str, query: &str) -> (String, String) {
        let expiry = (SystemTime::now() + LIFETIME).duration_since(UNIX_EPOCH).unwrap_or_default().as_millis();
        let signature = URL_SAFE_NO_PAD.encode(self.mac(path, query, expiry).finalize().into_bytes());
        let separator = if query.is_empty() { "" } else { "&" };
        (format!("{PREFIX}{path}"), format!("{query}{separator}exp={expiry}&sig={signature}"))
    }

    /// unseal returns the path and query that a sealed URL carries, or None when its signature is wrong or it expired.
    pub fn unseal(&self, path: &str, query: &str) -> Option<(String, String)> {
        let path = path.strip_prefix(PREFIX)?;
        let (rest, signature) = query.rsplit_once("&sig=")?;
        let (query, expiry) = match rest.rsplit_once("&exp=") {
            Some((query, expiry)) => (query, expiry),
            None => ("", rest.strip_prefix("exp=")?),
        };
        let expiry: u128 = expiry.parse().ok()?;
        let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default().as_millis();
        let signature = URL_SAFE_NO_PAD.decode(signature).ok()?;
        let valid = self.mac(path, query, expiry).verify_slice(&signature).is_ok();
        (expiry > now && valid).then(|| (path.to_string(), query.to_string()))
    }
}
