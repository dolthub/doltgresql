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

use harness::oid::*;
use harness::pgx::Time;
use harness::plan::PlanFact;
use harness::script::Cell::{Any, Null, Text as T};
use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};
use harness::wire::{Datum, F, Field, Fields, PGX_STARTUP, Receive, Send, Step, W, WireTest, run_wire_tests};

#[test]
fn test_create_database() {
    run_scripts(&[
        ScriptTest {
            name: "simple create database",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DATABASE testdb",
                    expected: Expected::Tag("CREATE DATABASE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE testdb",
                    expected: Expected::Tag("SET"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "encoding",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DATABASE testdb encoding=utf8",
                    expected: Expected::Tag("CREATE DATABASE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE testdb",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE DATABASE testdb2 encoding=latin1",
                    expected: Expected::Tag("CREATE DATABASE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE testdb2",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE DATABASE testdb3 encoding=notexist",
                    expected: Expected::Tag("CREATE DATABASE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "multiple options",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DATABASE testdb OWNER=foo ENCODING=utf8",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, and the failed CREATE DATABASE above creates nothing, as in Postgres.
                ScriptTestAssertion {
                    query: "USE testdb",
                    expected: Expected::Error(Diagnostic { code: "3D000", message: r#"database "testdb" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "with hyphen",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE DATABASE "test-db""#,
                    expected: Expected::Tag("CREATE DATABASE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"USE "test-db""#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
