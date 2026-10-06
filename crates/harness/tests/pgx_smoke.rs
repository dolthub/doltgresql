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

use harness::pgx::{Arg, Conn, ConnConfig, QueryExecMode};

/// connect returns a connection to the server in PGX_SMOKE_URL, or None when it is unset.
fn connect(mode: QueryExecMode) -> Option<Conn> {
    let url = std::env::var("PGX_SMOKE_URL").ok()?;
    let mut config = ConnConfig::parse(&url).unwrap();
    config.default_query_exec_mode = mode;
    Some(Conn::connect(config).unwrap())
}

#[test]
fn queries_round_trip() {
    for mode in [QueryExecMode::DescribeExec, QueryExecMode::CacheStatement] {
        let Some(mut conn) = connect(mode) else { return };
        conn.ping().unwrap();
        assert_eq!(
            conn.exec("CREATE TEMPORARY TABLE smoke (pk INT4 PRIMARY KEY, v TEXT);", &[]).unwrap(),
            "CREATE TABLE"
        );
        assert_eq!(
            conn.exec("INSERT INTO smoke VALUES ($1, $2);", &[Arg::Int(1), Arg::Str("one".into())]).unwrap(),
            "INSERT 0 1"
        );
        let result = conn.query("SELECT pk, v, 1.50::numeric FROM smoke WHERE pk = $1;", &[Arg::Int(1)]).unwrap();
        assert_eq!(result.error, None);
        assert_eq!(result.command_tag, "SELECT 1");
        assert_eq!(
            result.fields.iter().map(|f| (f.data_type_oid, f.format)).collect::<Vec<_>>(),
            [(23, 1), (25, 0), (1700, 1)]
        );
        assert_eq!(
            result.rows,
            vec![vec![
                Some(vec![0, 0, 0, 1]),
                Some(b"one".to_vec()),
                Some(vec![0, 2, 0, 0, 0, 0, 0, 2, 0, 1, 0x13, 0x88])
            ]]
        );
        let err = conn.exec("SELECT * FROM missing;", &[]).unwrap_err();
        assert_eq!(err.pg_error().unwrap().fields.code, "42P01");
    }
}
