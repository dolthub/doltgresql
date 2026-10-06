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

#[test]
fn test_invalid_startup_timezone_returns_error() {
    let target = Target::from_env().unwrap_or_else(|err| panic!("{err}"));
    let server = Server::start(&target, "").unwrap();
    let url = format!("postgres://postgres:password@127.0.0.1:{}/postgres", server.port);
    Conn::connect(ConnConfig::parse(&url).unwrap()).unwrap().close();

    let mut config = ConnConfig::parse(&url).unwrap();
    config.runtime_params.push(("timezone".to_string(), "Not/A/Timezone".to_string()));
    let err = Conn::connect(config).err().expect("expected the invalid timezone startup parameter to be rejected");
    let pg = err.pg_error().unwrap_or_else(|| {
        panic!("expected an explicit PostgreSQL error, not a bare connection failure like an unexpected EOF: {err}")
    });
    assert_eq!(pg.fields.code, "22023");
}
