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

//! The server side of SCRAM-SHA-256 (RFC 5802 and RFC 7677) as Postgres speaks it, without channel binding.

use base64::Engine as _;
use base64::engine::general_purpose::STANDARD;
use hmac::{Hmac, Mac};
use rand::RngCore;
use sha2::{Digest, Sha256};

/// ITERATIONS is the PBKDF2 iteration count of stored passwords, Postgres' default.
pub const ITERATIONS: u32 = 4096;

/// Verifier is what the server keeps of a password: its salt and iteration count with the derived keys.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Verifier {
    pub salt: Vec<u8>,
    pub iterations: u32,
    pub stored_key: [u8; 32],
    pub server_key: [u8; 32],
}

/// hmac returns the HMAC-SHA-256 of the message with the key.
fn hmac(key: &[u8], message: &[u8]) -> [u8; 32] {
    let mut mac = Hmac::<Sha256>::new_from_slice(key).expect("HMAC takes keys of any length");
    mac.update(message);
    mac.finalize().into_bytes().into()
}

impl Verifier {
    /// new derives the verifier of a password with a random salt.
    pub fn new(password: &str) -> Verifier {
        let mut salt = vec![0; 16];
        rand::thread_rng().fill_bytes(&mut salt);
        Verifier::with_salt(password, salt, ITERATIONS)
    }

    /// with_salt derives the verifier of a password with the salt and iteration count.
    pub fn with_salt(password: &str, salt: Vec<u8>, iterations: u32) -> Verifier {
        let mut salted = [0; 32];
        pbkdf2::pbkdf2_hmac::<Sha256>(password.as_bytes(), &salt, iterations, &mut salted);
        let client_key = hmac(&salted, b"Client Key");
        let stored_key: [u8; 32] = Sha256::digest(client_key).into();
        let server_key = hmac(&salted, b"Server Key");
        Verifier { salt, iterations, stored_key, server_key }
    }
}

/// Exchange is a SCRAM exchange in progress, after the server's first message.
pub struct Exchange {
    client_first_bare: String,
    server_first: String,
    nonce: String,
}

/// attribute returns the value of the attribute with the name in a comma-separated SCRAM message.
fn attribute(message: &str, name: char) -> Option<&str> {
    message.split(',').find_map(|part| part.strip_prefix(name)?.strip_prefix('='))
}

impl Exchange {
    /// start reads the client's first message and returns the exchange with the server's first message, or None
    /// when the message is malformed or asks for channel binding.
    pub fn start(client_first: &[u8], verifier: &Verifier) -> Option<(Exchange, String)> {
        let client_first = std::str::from_utf8(client_first).ok()?;
        // The GS2 header: no channel binding ("n") or a client that supports it but thinks the server does not ("y").
        let rest = client_first.strip_prefix("n,").or_else(|| client_first.strip_prefix("y,"))?;
        let (_authzid, client_first_bare) = rest.split_once(',')?;
        let client_nonce = attribute(client_first_bare, 'r')?;
        let mut server_nonce = [0; 18];
        rand::thread_rng().fill_bytes(&mut server_nonce);
        let nonce = format!("{client_nonce}{}", STANDARD.encode(server_nonce));
        let server_first = format!("r={nonce},s={},i={}", STANDARD.encode(&verifier.salt), verifier.iterations);
        let exchange =
            Exchange { client_first_bare: client_first_bare.to_string(), server_first: server_first.clone(), nonce };
        Some((exchange, server_first))
    }

    /// finish checks the client's final message, returning the server's final message when the client proved it
    /// knows the password.
    pub fn finish(self, client_final: &[u8], verifier: &Verifier) -> Option<String> {
        let client_final = std::str::from_utf8(client_final).ok()?;
        let (without_proof, proof) = client_final.rsplit_once(",p=")?;
        if attribute(without_proof, 'r')? != self.nonce {
            return None;
        }
        let proof = STANDARD.decode(proof).ok()?;
        let auth_message = format!("{},{},{without_proof}", self.client_first_bare, self.server_first);
        let client_signature = hmac(&verifier.stored_key, auth_message.as_bytes());
        if proof.len() != client_signature.len() {
            return None;
        }
        let client_key: Vec<u8> = proof.iter().zip(client_signature).map(|(p, s)| p ^ s).collect();
        let stored_key: [u8; 32] = Sha256::digest(&client_key).into();
        if stored_key != verifier.stored_key {
            return None;
        }
        let server_signature = hmac(&verifier.server_key, auth_message.as_bytes());
        Some(format!("v={}", STANDARD.encode(server_signature)))
    }
}
