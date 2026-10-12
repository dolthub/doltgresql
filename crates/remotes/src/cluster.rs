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

//! Cluster replication over the remotes API: the role and epoch that cluster members send with each request and
//! response, and the signed tokens that a primary sends and its standbys check against the primary's published key,
//! as Dolt's cluster interceptors, RPCCreds, and JWKS handler do.

use std::collections::HashMap;
use std::sync::Mutex;
use std::time::{Duration, SystemTime, UNIX_EPOCH};

use aws_lc_rs::rand::SystemRandom;
use aws_lc_rs::signature::{ED25519, Ed25519KeyPair, KeyPair, UnparsedPublicKey};
use base64::Engine;
use base64::engine::general_purpose::URL_SAFE_NO_PAD;
use sha2::{Digest, Sha512_224};

/// ROLE_HEADER and EPOCH_HEADER carry a cluster member's role and epoch on requests and responses.
pub const ROLE_HEADER: &str = "x-dolt-cluster-role";
pub const EPOCH_HEADER: &str = "x-dolt-cluster-role-epoch";

/// JWKS_PATH is where a cluster member publishes the key that verifies its tokens.
pub const JWKS_PATH: &str = "/.well-known/jwks.json";

/// ISSUER and AUDIENCE are the issuer and audience of the tokens that cluster members send, as Dolt's ClientIssuer
/// and DoltClusterRemoteApiAudience.
const ISSUER: &str = "dolt-client.dolthub.com";
const AUDIENCE: &str = "dolt-cluster-remote-api.dolthub.com";

/// LIFETIME is how long a token is good for, as Dolt's RPCCreds allows.
const LIFETIME: Duration = Duration::from_secs(30);

/// LEEWAY is how far past its expiry a token is still accepted, as go-jose's DefaultLeeway allows.
const LEEWAY: u64 = 60;

/// Member is a cluster member as its remotes API server and clients see it.
pub trait Member: Send + Sync + 'static {
    /// role returns the member's role and epoch.
    fn role(&self) -> (String, i64);

    /// force_role moves the member to a role at an epoch because another member showed that it must, as Dolt's
    /// roleSetter does.
    fn force_role(&self, role: &str, epoch: i64);

    /// credentials returns the member's signing key.
    fn credentials(&self) -> &Credentials;

    /// keys returns the keys of the members that may replicate to this one.
    fn keys(&self) -> &KeySet;

    /// update_users replaces the roles and privileges with a primary's serialized copy.
    fn update_users(&self, contents: &[u8]) -> Result<(), String>;

    /// update_branch_control replaces the branch control tables with a primary's serialized copy.
    fn update_branch_control(&self, contents: &[u8]) -> Result<(), String>;

    /// drop_database drops a database that a primary dropped.
    fn drop_database(&self, name: &str) -> Result<(), String>;
}

/// Credentials is a member's signing key, from which it makes the tokens it sends and the key set it publishes.
pub struct Credentials {
    pair: Ed25519KeyPair,
    kid: String,
}

impl Credentials {
    /// new returns credentials with a new random key, whose key ID is the base32 of the SHA-512/224 of the public key
    /// as Dolt's PubKeyToKID makes it.
    pub fn new() -> Credentials {
        let document = Ed25519KeyPair::generate_pkcs8(&SystemRandom::new()).expect("the system has randomness");
        let pair = Ed25519KeyPair::from_pkcs8(document.as_ref()).expect("a new key parses");
        let kid = base32(&Sha512_224::digest(pair.public_key().as_ref()));
        Credentials { pair, kid }
    }

    /// token returns a bearer token that expires shortly, signed with the key.
    pub fn token(&self) -> String {
        let header = format!(r#"{{"alg":"EdDSA","dolt_token_version":"2023.01","kid":"{}"}}"#, self.kid);
        let expiry = (SystemTime::now() + LIFETIME).duration_since(UNIX_EPOCH).unwrap_or_default().as_secs();
        let claims = format!(r#"{{"aud":["{AUDIENCE}"],"exp":{expiry},"iss":"{ISSUER}"}}"#);
        let signed = format!("{}.{}", URL_SAFE_NO_PAD.encode(header), URL_SAFE_NO_PAD.encode(claims));
        let signature = URL_SAFE_NO_PAD.encode(self.pair.sign(signed.as_bytes()));
        format!("{signed}.{signature}")
    }

    /// jwks returns the JSON web key set that publishes the public key.
    pub fn jwks(&self) -> String {
        let x = URL_SAFE_NO_PAD.encode(self.pair.public_key().as_ref());
        format!(r#"{{"keys":[{{"kty":"OKP","kid":"{}","crv":"Ed25519","x":"{x}"}}]}}"#, self.kid)
    }
}

impl Default for Credentials {
    fn default() -> Credentials {
        Credentials::new()
    }
}

/// base32 encodes bytes in Dolt's base32 alphabet without padding.
fn base32(bytes: &[u8]) -> String {
    const ALPHABET: &[u8] = b"0123456789abcdefghijklmnopqrstuv";
    let mut out = String::new();
    let (mut buffer, mut bits) = (0u32, 0);
    for &byte in bytes {
        buffer = (buffer << 8) | u32::from(byte);
        bits += 8;
        while bits >= 5 {
            bits -= 5;
            out.push(ALPHABET[((buffer >> bits) & 31) as usize] as char);
        }
    }
    if bits > 0 {
        out.push(ALPHABET[((buffer << (5 - bits)) & 31) as usize] as char);
    }
    out
}

/// KeySet is the public keys that other members publish at their key set URLs, fetched when a token names a key
/// that is not yet known, as Dolt's MultiJWKS does.
pub struct KeySet {
    urls: Vec<String>,
    keys: Mutex<HashMap<String, Vec<u8>>>,
    http: reqwest::Client,
}

impl KeySet {
    /// new returns a key set that fetches from the URLs.
    pub fn new(urls: Vec<String>) -> KeySet {
        KeySet { urls, keys: Mutex::default(), http: reqwest::Client::new() }
    }

    /// refresh fetches every URL's keys.
    async fn refresh(&self) {
        for url in &self.urls {
            let Ok(response) = self.http.get(url).send().await else { continue };
            let Ok(body) = response.bytes().await else { continue };
            let Ok(set) = serde_json::from_slice::<serde_json::Value>(&body) else { continue };
            let mut keys = self.keys.lock().unwrap_or_else(|p| p.into_inner());
            for key in set["keys"].as_array().into_iter().flatten() {
                let (Some(kid), Some(x)) = (key["kid"].as_str(), key["x"].as_str()) else { continue };
                if let Ok(public) = URL_SAFE_NO_PAD.decode(x) {
                    keys.insert(kid.to_string(), public);
                }
            }
        }
    }

    /// verify reports whether a bearer token is signed by a known key, is for cluster replication, and has not expired.
    pub async fn verify(&self, token: &str) -> bool {
        let mut parts = token.split('.');
        let (Some(header), Some(claims), Some(signature), None) =
            (parts.next(), parts.next(), parts.next(), parts.next())
        else {
            return false;
        };
        let decode = |part: &str| URL_SAFE_NO_PAD.decode(part).ok();
        let json = |part: &str| decode(part).and_then(|b| serde_json::from_slice::<serde_json::Value>(&b).ok());
        let (Some(header), Some(claims), Some(signature)) = (json(header), json(claims), decode(signature)) else {
            return false;
        };
        let Some(kid) = header["kid"].as_str().filter(|_| header["alg"] == "EdDSA") else { return false };
        let known = |keys: &KeySet| keys.keys.lock().unwrap_or_else(|p| p.into_inner()).get(kid).cloned();
        let public = match known(self) {
            Some(public) => public,
            None => {
                self.refresh().await;
                let Some(public) = known(self) else { return false };
                public
            }
        };
        let signed = &token[..token.rfind('.').unwrap_or_default()];
        if UnparsedPublicKey::new(&ED25519, public).verify(signed.as_bytes(), &signature).is_err() {
            return false;
        }
        let audience = match &claims["aud"] {
            serde_json::Value::String(aud) => aud == AUDIENCE,
            serde_json::Value::Array(auds) => auds.iter().any(|a| a == AUDIENCE),
            _ => false,
        };
        let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default().as_secs();
        let fresh = claims["exp"].as_u64().is_none_or(|exp| now <= exp + LEEWAY);
        claims["iss"] == ISSUER && audience && fresh
    }
}
