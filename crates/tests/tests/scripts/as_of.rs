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
fn test_as_of() {
    run_scripts(&[
        ScriptTest {
            name: "Single table",
            set_up_script: &[
                "CREATE TABLE test (a INT)",
                "INSERT INTO test VALUES (1)",
                "SELECT DOLT_COMMIT('-Am', 'new table')",
                "INSERT INTO test VALUES (2)",
                "SELECT DOLT_COMMIT('-am', 'new row')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test AS OF 'HEAD^' as t1",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test AS OF 'HEAD'",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Join",
            set_up_script: &[
                "CREATE TABLE test (a INT)",
                "INSERT INTO test VALUES (1)",
                "SELECT DOLT_COMMIT('-Am', 'new table')",
                "INSERT INTO test VALUES (2)",
                "SELECT DOLT_COMMIT('-am', 'new row')",
                "CREATE TABLE test2 (b INT)",
                "INSERT INTO test2 VALUES (1)",
                "SELECT DOLT_COMMIT('-Am', 'new table')",
                "INSERT INTO test2 VALUES (2)",
                "SELECT DOLT_COMMIT('-am', 'new row')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test AS OF 'HEAD~3' t1 join test2 AS OF 'HEAD~' t2 on t1.a = t2.b",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test AS OF 'HEAD~3' t1 join test2 AS t2 on t1.a = t2.b",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test t1 join test2 AS OF 'HEAD~' AS t2 on t1.a = t2.b",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test AS OF 'HEAD~3' t1 cross join test2 AS OF 'HEAD~' t2",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test AS OF 'HEAD' t1 cross join test2 AS OF 'HEAD~' t2",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), T("1")],
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Syntax variations",
            set_up_script: &[
                "CREATE TABLE test (a INT)",
                "INSERT INTO test VALUES (1)",
                "SELECT DOLT_COMMIT('-Am', 'new table')",
                "INSERT INTO test VALUES (2)",
                "SELECT DOLT_COMMIT('-am', 'new row')",
                "CREATE TABLE test2 (b INT)",
                "INSERT INTO test2 VALUES (1)",
                "SELECT DOLT_COMMIT('-Am', 'new table')",
                "INSERT INTO test2 VALUES (2)",
                "SELECT DOLT_COMMIT('-am', 'new row')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test AS OF 'HEAD~3' AS t1 join test2 AS OF 'HEAD' AS t2 on t1.a = t2.b",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test AS OF 'HEAD~3' t1 join test2 AS OF 'HEAD' t2 on t1.a = t2.b",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test AS t1 join test2 AS OF 'HEAD~' t2 on t1.a = t2.b",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test AS OF SYSTEM TIME 'HEAD~3' join test2 AS t2 on test.a = t2.b",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
