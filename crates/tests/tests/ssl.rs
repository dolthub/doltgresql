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

use std::path::PathBuf;
use std::time::Duration;

use harness::pgx::{Conn, ConnConfig};
use harness::script::{A, Cell, Column, Expected, ScriptTestAssertion, run_assertion};
use harness::server::{Server, Target};

/// write_test_certificate writes a self-signed certificate and its private key, readable only by their owner as
/// Postgres requires, into a new directory.
fn write_test_certificate() -> (PathBuf, PathBuf, PathBuf) {
    let dir = std::env::temp_dir().join(format!("ssl-test-{}", std::process::id()));
    std::fs::create_dir_all(&dir).unwrap();
    let certified = rcgen::generate_simple_self_signed(vec!["localhost".to_string()]).unwrap();
    let (cert, key) = (dir.join("cert.pem"), dir.join("key.pem"));
    std::fs::write(&cert, certified.cert.pem()).unwrap();
    std::fs::write(&key, certified.signing_key.serialize_pem()).unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        std::fs::set_permissions(&key, std::fs::Permissions::from_mode(0o600)).unwrap();
    }
    (dir, cert, key)
}

#[test]
fn test_ssl() {
    let target = Target::from_env().unwrap_or_else(|err| panic!("{err}"));
    let (dir, cert, key) = write_test_certificate();
    let server = Server::start_tls(&target, &cert, &key).unwrap();
    let url =
        |database: &str| format!("postgres://postgres:password@127.0.0.1:{}/{database}?sslmode=require", server.port);

    let mut result = Err(String::new());
    for _ in 0..3 {
        result = ConnConfig::parse(&url("")).and_then(Conn::connect).map_err(|err| err.to_string());
        if result.is_ok() {
            break;
        }
        std::thread::sleep(Duration::from_secs(1));
    }
    result.unwrap().close();

    let mut conn = Conn::connect(ConnConfig::parse(&url("postgres")).unwrap()).unwrap();
    let mut failures = Vec::new();
    failures.extend(run_assertion(
        &mut conn,
        &ScriptTestAssertion { query: "CREATE TABLE test (pk INT8 PRIMARY KEY, v1 int4);", ..A },
    ));
    failures.extend(run_assertion(
        &mut conn,
        &ScriptTestAssertion { query: "INSERT INTO test VALUES (3645, 37643);", ..A },
    ));
    let select = ScriptTestAssertion {
        query: "SELECT * FROM test;",
        expected: Expected::Rows {
            columns: &[Column("pk", 20), Column("v1", 23)],
            rows: &[&[Cell::Text("3645"), Cell::Text("37643")]],
            tag: "SELECT 1",
        },
        ..A
    };
    failures.extend(run_assertion(&mut conn, &select));
    conn.close();
    drop(server);
    let _ = std::fs::remove_dir_all(dir);
    assert!(failures.is_empty(), "{}", failures.join("\n"));
}
