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

use base64::Engine;
use base64::engine::general_purpose::{STANDARD, STANDARD_NO_PAD};
use hmac::{Hmac, Mac};
use md5::{Digest, Md5};
use rand::RngCore;
use sha2::Sha256;

/// The mechanism name of SCRAM-SHA-256 without channel binding.
pub(crate) const SCRAM_SHA_256: &str = "SCRAM-SHA-256";

/// md5_password returns the password message for MD5 authentication.
pub(crate) fn md5_password(user: &str, password: &str, salt: &[u8; 4]) -> String {
    let inner = hex(&Md5::digest(format!("{password}{user}").as_bytes()));
    let mut outer = Md5::new();
    outer.update(inner.as_bytes());
    outer.update(salt);
    format!("md5{}", hex(&outer.finalize()))
}

/// hex renders bytes as lowercase hexadecimal.
pub(crate) fn hex(bytes: &[u8]) -> String {
    bytes.iter().map(|b| format!("{b:02x}")).collect()
}

/// ScramClient performs the client side of SCRAM-SHA-256 without channel binding, as pgx does over a plaintext
/// connection.
pub(crate) struct ScramClient {
    password: String,
    client_nonce: String,
    client_first_message_bare: String,
    salted_password: Vec<u8>,
    auth_message: String,
}

impl ScramClient {
    /// new returns a client with a fresh random nonce. Like pgx, the nonce is 18 random bytes encoded as unpadded
    /// base64. Passwords are used as given, which matches pgx for every password that SASLprep leaves unchanged.
    pub(crate) fn new(password: &str) -> ScramClient {
        let mut nonce = [0u8; 18];
        rand::thread_rng().fill_bytes(&mut nonce);
        ScramClient {
            password: password.to_string(),
            client_nonce: STANDARD_NO_PAD.encode(nonce),
            client_first_message_bare: String::new(),
            salted_password: Vec::new(),
            auth_message: String::new(),
        }
    }

    /// client_first_message returns the data of the SASLInitialResponse.
    pub(crate) fn client_first_message(&mut self) -> Vec<u8> {
        self.client_first_message_bare = format!("n=,r={}", self.client_nonce);
        format!("n,,{}", self.client_first_message_bare).into_bytes()
    }

    /// client_final_message consumes the server-first-message and returns the data of the SASLResponse.
    pub(crate) fn client_final_message(&mut self, server_first_message: &[u8]) -> Result<Vec<u8>, String> {
        let server_first = String::from_utf8_lossy(server_first_message).into_owned();
        let rest = server_first
            .strip_prefix("r=")
            .ok_or("invalid SCRAM server-first-message received from server: did not include r=")?;
        let (nonce, rest) = rest
            .split_once(',')
            .ok_or("invalid SCRAM server-first-message received from server: did not include s=")?;
        let rest = rest
            .strip_prefix("s=")
            .ok_or("invalid SCRAM server-first-message received from server: did not include s=")?;
        let (salt, rest) = rest
            .split_once(',')
            .ok_or("invalid SCRAM server-first-message received from server: did not include i=")?;
        let iterations = rest
            .strip_prefix("i=")
            .ok_or("invalid SCRAM server-first-message received from server: did not include i=")?;
        let salt = STANDARD.decode(salt).map_err(|err| format!("invalid SCRAM salt received from server: {err}"))?;
        let iterations: u32 =
            iterations.parse().map_err(|err| format!("invalid SCRAM iteration count received from server: {err}"))?;
        if !nonce.starts_with(&self.client_nonce) {
            return Err("invalid SCRAM nonce: did not start with client nonce".to_string());
        }

        let mut salted_password = [0u8; 32];
        pbkdf2::pbkdf2_hmac::<Sha256>(self.password.as_bytes(), &salt, iterations, &mut salted_password);
        let client_final_without_proof = format!("c={},r={nonce}", STANDARD.encode("n,,"));
        self.auth_message = format!("{},{server_first},{client_final_without_proof}", self.client_first_message_bare);
        self.salted_password = salted_password.to_vec();

        let client_key = hmac_sha256(&salted_password, b"Client Key");
        let stored_key = sha2::Sha256::digest(&client_key);
        let client_signature = hmac_sha256(&stored_key, self.auth_message.as_bytes());
        let proof: Vec<u8> = client_key.iter().zip(client_signature.iter()).map(|(a, b)| a ^ b).collect();
        Ok(format!("{client_final_without_proof},p={}", STANDARD.encode(proof)).into_bytes())
    }

    /// verify_server_final_message checks the server's signature.
    pub(crate) fn verify_server_final_message(&self, server_final_message: &[u8]) -> Result<(), String> {
        let signature = server_final_message
            .strip_prefix(b"v=")
            .ok_or("invalid SCRAM server-final-message received from server")?;
        let server_key = hmac_sha256(&self.salted_password, b"Server Key");
        let expected = STANDARD.encode(hmac_sha256(&server_key, self.auth_message.as_bytes()));
        if signature != expected.as_bytes() {
            return Err("invalid SCRAM ServerSignature received from server".to_string());
        }
        Ok(())
    }
}

/// hmac_sha256 returns the HMAC-SHA-256 of the message with the key.
fn hmac_sha256(key: &[u8], message: &[u8]) -> Vec<u8> {
    let mut mac = Hmac::<Sha256>::new_from_slice(key).expect("HMAC accepts keys of any length");
    mac.update(message);
    mac.finalize().into_bytes().to_vec()
}
