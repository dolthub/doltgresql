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
fn test_drop_database() {
    run_scripts(&[
        ScriptTest {
            name: "simple create database",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DATABASE testdb",
                    expected: Expected::Tag("CREATE DATABASE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DATABASE testdb",
                    expected: Expected::Tag("DROP DATABASE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "with quotes",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DATABASE testdb",
                    expected: Expected::Tag("CREATE DATABASE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP DATABASE "testdb""#,
                    expected: Expected::Tag("DROP DATABASE"),
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
        ScriptTest {
            name: "drop a database that was previously used in the session",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DATABASE dropdb1",
                    expected: Expected::Tag("CREATE DATABASE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DATABASE dropdb2",
                    expected: Expected::Tag("CREATE DATABASE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE dropdb2",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t1 (a INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE dropdb1",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DATABASE dropdb2",
                    expected: Expected::Tag("DROP DATABASE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "if exists",
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP DATABASE IF EXISTS invalid",
                    expected: Expected::Tag("DROP DATABASE"),
                    notices: &[Diagnostic { code: "00000", message: r#"database "invalid" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
    ]);
}
