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

mod common;

use std::io::{BufRead, BufReader, Read, Write};
use std::net::{TcpListener, TcpStream};
use std::path::PathBuf;
use std::sync::OnceLock;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

use base64::Engine;
use base64::engine::general_purpose::URL_SAFE_NO_PAD;
use common::*;
use driver::model::Server;
use driver::resources::Resources;

/// testdata returns the path of a file in the Go driver tests' testdata directory.
fn testdata(name: &str) -> PathBuf {
    std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("../../integration-tests/go-sql-server-driver/testdata")
        .join(name)
}

/// read_request reads an HTTP request's head, returning its path.
fn read_request(stream: &mut TcpStream) -> std::io::Result<String> {
    let mut reader = BufReader::new(stream);
    let mut request_line = String::new();
    reader.read_line(&mut request_line)?;
    let mut line = String::new();
    while reader.read_line(&mut line)? > 2 {
        line.clear();
    }
    Ok(request_line.split_whitespace().nth(1).unwrap_or("").to_string())
}

/// jwks_port serves the test JWKS at /jwks.json for the rest of the process, returning its port.
fn jwks_port() -> u16 {
    static PORT: OnceLock<u16> = OnceLock::new();
    *PORT.get_or_init(|| {
        let port = Box::leak(Box::new(Resources::default())).port("jwks").unwrap();
        let data = std::fs::read(testdata("test_jwks.json")).unwrap();
        let listener = TcpListener::bind(("127.0.0.1", port)).unwrap();
        std::thread::spawn(move || {
            for mut stream in listener.incoming().flatten() {
                let Ok(path) = read_request(&mut stream) else { continue };
                let (status, content_type, body) = if path == "/jwks.json" {
                    ("200 OK", "application/json", data.clone())
                } else {
                    ("404 Not Found", "text/plain; charset=utf-8", b"404 page not found\n".to_vec())
                };
                let head = format!(
                    "HTTP/1.1 {status}\r\nContent-Type: {content_type}\r\nContent-Length: {}\r\nConnection: close\r\n\r\n",
                    body.len()
                );
                let _ = stream.write_all(head.as_bytes()).and_then(|_| stream.write_all(&body));
            }
        });
        port
    })
}

/// der_length appends a DER length.
fn der_length(len: usize, out: &mut Vec<u8>) {
    if len < 128 {
        out.push(len as u8);
        return;
    }
    let bytes = len.to_be_bytes();
    let bytes = &bytes[bytes.iter().position(|b| *b != 0).unwrap()..];
    out.push(0x80 | bytes.len() as u8);
    out.extend_from_slice(bytes);
}

/// der_integer appends a non-negative DER INTEGER from its big-endian bytes.
fn der_integer(value: &[u8], out: &mut Vec<u8>) {
    let start = value.iter().position(|b| *b != 0).unwrap_or(value.len().saturating_sub(1));
    let mut content = value[start..].to_vec();
    if content.first().is_none_or(|b| b & 0x80 != 0) {
        content.insert(0, 0);
    }
    out.push(0x02);
    der_length(content.len(), out);
    out.extend(content);
}

/// rsa_private_key_der builds a PKCS #1 RSAPrivateKey from a private JWK.
fn rsa_private_key_der(jwk: &serde_json::Value) -> Vec<u8> {
    let mut content = Vec::new();
    der_integer(&[0], &mut content);
    for field in ["n", "e", "d", "p", "q", "dp", "dq", "qi"] {
        der_integer(&URL_SAFE_NO_PAD.decode(jwk[field].as_str().unwrap()).unwrap(), &mut content);
    }
    let mut der = vec![0x30];
    der_length(content.len(), &mut der);
    der.extend(content);
    der
}

/// create_jwt signs a JWT with RS256 using the test JWKS's private key, like go-jose.
fn create_jwt(issuer: &str, audience: &str, subject: &str) -> String {
    const KID: &str = "749df841-6e38-48f1-a178-20ecdd0b09f7";
    let jwks: serde_json::Value =
        serde_json::from_slice(&std::fs::read(testdata("test_jwks_private.json")).unwrap()).unwrap();
    let keys = jwks["keys"].as_array().unwrap();
    assert!(!keys.is_empty());
    let jwk = keys.iter().find(|k| k["kid"] == KID).expect("no key with the kid");
    assert!(jwk.get("d").is_some(), "the key must be private");
    let key = aws_lc_rs::signature::RsaKeyPair::from_der(&rsa_private_key_der(jwk)).unwrap();

    let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() as i64;
    let header = serde_json::json!({"alg": "RS256", "kid": KID, "typ": "JWT"});
    let claims = serde_json::json!({"aud": audience, "exp": now + 3600, "iat": now, "iss": issuer, "sub": subject});
    let signing_input = format!(
        "{}.{}",
        URL_SAFE_NO_PAD.encode(serde_json::to_vec(&header).unwrap()),
        URL_SAFE_NO_PAD.encode(serde_json::to_vec(&claims).unwrap())
    );
    let mut signature = vec![0; key.public_modulus_len()];
    let rng = aws_lc_rs::rand::SystemRandom::new();
    key.sign(&aws_lc_rs::signature::RSA_PKCS1_SHA256, &rng, signing_input.as_bytes(), &mut signature).unwrap();
    format!("{signing_input}.{}", URL_SAFE_NO_PAD.encode(signature))
}

/// metrics_status gets /metrics with the bearer token when there is one, returning the response's status code.
fn metrics_status(metrics_port: u16, bearer_token: &str) -> std::io::Result<u16> {
    let mut stream = TcpStream::connect(("127.0.0.1", metrics_port))?;
    let mut request =
        format!("GET /metrics HTTP/1.1\r\nHost: 127.0.0.1:{metrics_port}\r\nUser-Agent: Go-http-client/1.1\r\n");
    if !bearer_token.is_empty() {
        request.push_str(&format!("Authorization: Bearer {bearer_token}\r\n"));
    }
    request.push_str("Accept-Encoding: gzip\r\nConnection: close\r\n\r\n");
    stream.write_all(request.as_bytes())?;
    let mut response = Vec::new();
    stream.read_to_end(&mut response)?;
    let status_line = String::from_utf8_lossy(&response).lines().next().unwrap_or("").to_string();
    status_line
        .split_whitespace()
        .nth(1)
        .and_then(|code| code.parse().ok())
        .ok_or_else(|| std::io::Error::other(format!("malformed status line {status_line:?}")))
}

/// start_server_with_metrics starts a server with the config and waits for its metrics listener.
fn start_server_with_metrics(get_config: impl Fn(u16, u16) -> String) -> (driver::runner::Env, u16) {
    let mut env = env();
    let server_port = env.resources.port("server").unwrap();
    let metrics_port = env.resources.port("metrics").unwrap();
    let config = get_config(server_port, metrics_port);
    let repo = make_repo(&mut env, "metrics_auth_test");
    let path = repo.store.user.dir.join("config.yaml");
    std::fs::write(&path, &config).unwrap();
    let server = Server {
        args: vec!["--config".into(), path.to_string_lossy().into_owned()],
        dynamic_port: "server".into(),
        ..Server::default()
    };
    println!("Starting server with config:\n{config}");
    assert!(env.start_server(SERVER, &repo.store, &server, None).unwrap());
    let deadline = Instant::now() + Duration::from_secs(5);
    while metrics_status(metrics_port, "").is_err() {
        assert!(Instant::now() < deadline, "the metrics listener did not start");
        std::thread::sleep(Duration::from_millis(50));
    }
    (env, metrics_port)
}

/// jwks_config returns a config whose metrics listener requires JWTs checked against the test JWKS, with the
/// algorithm claim when given.
fn jwks_config(server_port: u16, metrics_port: u16, alg: bool) -> String {
    let alg = if alg { "      alg: RS256\n" } else { "" };
    format!(
        "listener:\n  host: localhost\n  port: {server_port}\n\nmetrics:\n  host: localhost\n  port: {metrics_port}\n  jwks:\n    \
         name: jwksname\n    location_url: http://127.0.0.1:{}/jwks.json\n    claims:\n{alg}      iss: dolthub.com\n      \
         sub: test_sub\n      aud: test_aud\n  jwt_required_for_localhost: true\n",
        jwks_port()
    )
}

/// run_metrics_auth requests metrics with the token that make_token returns and checks the status.
fn run_metrics_auth(alg: bool, token: Option<(&str, &str, &str)>, expected: u16) {
    let (env, metrics_port) = start_server_with_metrics(|server, metrics| jwks_config(server, metrics, alg));
    let token = token.map(|(iss, aud, sub)| create_jwt(iss, aud, sub)).unwrap_or_default();
    assert_eq!(metrics_status(metrics_port, &token).unwrap(), expected);
    finish(env);
}

#[test]
#[ignore = "metrics JWKS / JWT-auth configuration (metrics.jwks, jwt_required_for_localhost) is unsupported in Doltgres"]
fn test_metrics_auth_no_metrics_auth() {
    let (env, metrics_port) = start_server_with_metrics(|server_port, metrics_port| {
        format!(
            "\nlistener:\n  host: localhost\n  port: {server_port}\n\nmetrics:\n  host: localhost\n  port: {metrics_port}\n  \
             jwt_required_for_localhost: true\n"
        )
    });
    assert_eq!(metrics_status(metrics_port, "").unwrap(), 200);
    finish(env);
}

#[test]
#[ignore = "metrics JWKS / JWT-auth configuration (metrics.jwks, jwt_required_for_localhost) is unsupported in Doltgres"]
fn test_metrics_auth_missing_metrics_auth() {
    run_metrics_auth(true, None, 401);
}

#[test]
#[ignore = "metrics JWKS / JWT-auth configuration (metrics.jwks, jwt_required_for_localhost) is unsupported in Doltgres"]
fn test_metrics_auth_valid_metrics_auth() {
    run_metrics_auth(false, Some(("dolthub.com", "test_aud", "test_sub")), 200);
}

#[test]
#[ignore = "metrics JWKS / JWT-auth configuration (metrics.jwks, jwt_required_for_localhost) is unsupported in Doltgres"]
fn test_metrics_auth_bad_audience_claim() {
    run_metrics_auth(false, Some(("dolthub.com", "bad_aud", "test_sub")), 401);
}

#[test]
#[ignore = "metrics JWKS / JWT-auth configuration (metrics.jwks, jwt_required_for_localhost) is unsupported in Doltgres"]
fn test_metrics_auth_bad_issuer_claim() {
    run_metrics_auth(false, Some(("badissuer.com", "test_aud", "test_sub")), 401);
}

#[test]
#[ignore = "metrics JWKS / JWT-auth configuration (metrics.jwks, jwt_required_for_localhost) is unsupported in Doltgres"]
fn test_metrics_auth_bad_subject_claim() {
    run_metrics_auth(false, Some(("dolthub.com", "test_aud", "bad_sub")), 401);
}

#[test]
fn jwt_signature_verifies_with_the_public_jwks() {
    let token = create_jwt("dolthub.com", "test_aud", "test_sub");
    let (signing_input, signature) = token.rsplit_once('.').unwrap();
    let jwks: serde_json::Value = serde_json::from_slice(&std::fs::read(testdata("test_jwks.json")).unwrap()).unwrap();
    let jwk = &jwks["keys"][0];
    let component = |name: &str| URL_SAFE_NO_PAD.decode(jwk[name].as_str().unwrap()).unwrap();
    let public = aws_lc_rs::signature::RsaPublicKeyComponents { n: component("n"), e: component("e") };
    public
        .verify(
            &aws_lc_rs::signature::RSA_PKCS1_2048_8192_SHA256,
            signing_input.as_bytes(),
            &URL_SAFE_NO_PAD.decode(signature).unwrap(),
        )
        .unwrap();
    let header = URL_SAFE_NO_PAD.decode(signing_input.split('.').next().unwrap()).unwrap();
    assert_eq!(
        String::from_utf8(header).unwrap(),
        r#"{"alg":"RS256","kid":"749df841-6e38-48f1-a178-20ecdd0b09f7","typ":"JWT"}"#
    );
}
