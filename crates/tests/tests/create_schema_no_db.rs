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

use harness::pgx::{Conn, ConnConfig};
use harness::server::{Server, Target};

/// start starts a fresh server, which scripts cannot be used for since they create their database first.
fn start() -> Server {
    let target = Target::from_env().unwrap_or_else(|err| panic!("{err}"));
    Server::start(&target, "").unwrap()
}

/// connect connects to the database on the server without TLS.
fn connect(server: &Server, database: &str) -> Result<Conn, harness::pgx::Error> {
    let url = format!("postgres://postgres:password@127.0.0.1:{}/{database}?sslmode=disable", server.port);
    Conn::connect(ConnConfig::parse(&url)?)
}

#[test]
fn test_create_schema_with_non_existent_database_connection_to_non_existent_database_fails() {
    let server = start();
    let err = connect(&server, "nonexistent_db").err().expect("connection should fail when database doesn't exist");
    assert!(err.to_string().contains("does not exist"), "expected 'does not exist' error, got: {err}");
}

#[test]
fn test_create_schema_with_non_existent_database_connection_to_existing_database_succeeds() {
    let server = start();
    let mut conn = connect(&server, "postgres").unwrap();
    conn.exec("CREATE SCHEMA test_schema_1863", &[]).unwrap();
    let result = conn
        .query("SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'test_schema_1863'", &[])
        .unwrap();
    assert!(result.error.is_none(), "{:?}", result.error);
    assert!(!result.rows.is_empty(), "expected to find test_schema_1863");
    conn.close();
}
