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
fn test_advisory_locks() {
    run_scripts(&[
        ScriptTest {
            name: "basic lock tests",
            set_up_script: &[
                "CREATE USER user1 PASSWORD 'password';",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_advisory_lock(1)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_lock", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_try_advisory_lock(2)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_try_advisory_lock(1)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_try_advisory_lock(2)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_advisory_unlock(1)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_advisory_unlock(2)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_advisory_unlock(3)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    notices: &[Diagnostic { severity: "WARNING", code: "01000", message: "you don't own a lock of type ExclusiveLock", ..E }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "advisory locks are reentrant",
            set_up_script: &[
                "CREATE USER user1 PASSWORD 'password';",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_advisory_lock(10)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_lock", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_advisory_lock(10)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_lock", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_try_advisory_lock(10)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_try_advisory_lock(10)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_advisory_unlock(10)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_advisory_unlock(10)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_try_advisory_lock(10)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_advisory_unlock(10)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_try_advisory_lock(10)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_advisory_locks_2() {
    run_scripts(&[
        ScriptTest {
            name: "transaction advisory locks",
            set_up_script: &[
                "CREATE TABLE lock_commit_test (pk INT PRIMARY KEY)",
                "SELECT DOLT_BRANCH('lock-test-branch')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_xact_lock(20)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_xact_lock", VOID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(20)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(20)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_unlock(20)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_try_advisory_xact_lock(21)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_xact_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(21)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(21)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_unlock(21)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_xact_lock(42)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_xact_lock", VOID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_unlock(42)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(42)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(42)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_unlock(42)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_lock(43)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_xact_lock(43)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_xact_lock", VOID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_unlock(43)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_unlock(43)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(43)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(43)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_unlock(43)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_xact_lock(22)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_xact_lock", VOID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_xact_lock(22)",
                    flow: Flow::Exec,
                    client: "B",
                    expected_blocking: true,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client C */ SELECT pg_try_advisory_lock(22)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "C",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client C */ SELECT pg_advisory_unlock(22)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "C",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_try_advisory_xact_lock(23)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_xact_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(23)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_unlock(23)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_xact_lock(24)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_xact_lock", VOID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ INSERT INTO lock_commit_test VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT DOLT_COMMIT('-Am', 'lock lifecycle test')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(24)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_unlock(24)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_xact_lock(25)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_xact_lock", VOID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ SELECT DOLT_CHECKOUT('lock-test-branch')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(25)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client A */ COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(25)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_unlock(25)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "transaction locks release across session reuse",
            assertions: &[
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_xact_lock(30)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_xact_lock", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(30)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_unlock(30)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_xact_lock(31)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_xact_lock", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_try_advisory_lock(31)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_unlock(31)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_xact_lock(32)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_xact_lock", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client B */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "B",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_xact_lock(32)",
                    flow: Flow::Exec,
                    client: "B",
                    expected_blocking: true,
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client B */ COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "B",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client C */ SELECT pg_try_advisory_lock(32)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "C",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client C */ SELECT pg_advisory_unlock(32)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "C",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "advisory locks release when client disconnects",
            assertions: &[
                ScriptTestAssertion {
                    query: "/* client A */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_lock(40)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_lock", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ SELECT pg_advisory_xact_lock(41)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_xact_lock", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "A",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_lock(40)",
                    flow: Flow::Exec,
                    client: "B",
                    expected_blocking: true,
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client C */ BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "C",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client C */ SELECT pg_advisory_xact_lock(41)",
                    flow: Flow::Exec,
                    client: "C",
                    expected_blocking: true,
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client A */ DISCONNECT",
                    flow: Flow::Exec,
                    client: "A",
                    close_client: true,
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client B */ SELECT pg_advisory_unlock(40)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "B",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client C */ COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "C",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client D */ SELECT pg_try_advisory_lock(40)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "D",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client D */ SELECT pg_advisory_unlock(40)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "D",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client D */ SELECT pg_try_advisory_lock(41)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_try_advisory_lock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "D",
                    ..A
                },
                ScriptTestAssertion {
                    query: "/* client D */ SELECT pg_advisory_unlock(41)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_advisory_unlock", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "D",
                    ..A
                },
            ],
            ..S
        },
    ]);
}
