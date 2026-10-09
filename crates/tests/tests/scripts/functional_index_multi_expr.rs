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
fn test_functional_index_multi_expr() {
    run_scripts(&[
        ScriptTest {
            name: "mixed expressions and a plain column, filtering across all key parts",
            set_up_script: &[
                "CREATE TABLE t (pk int primary key, name text, age int, c1 int, c2 int);",
                "INSERT INTO t VALUES (1, 'alice', 30, 1, 2), (2, 'bob', 40, 3, 4);",
                "CREATE INDEX idx1 ON t (upper(name), age, (c1 + c2));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE upper(name) = 'ALICE' AND age = 30 AND c1 + c2 = 3;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE upper(name) = 'BOB' AND age = 40 AND c1 + c2 = 7;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx1'::regclass);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("CREATE INDEX idx1 ON public.t USING btree (upper(name), age, ((c1 + c2)))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT indexdef FROM pg_indexes WHERE indexname = 'idx1';",
                    expected: Expected::Rows {
                        columns: &[Column("indexdef", TEXT)],
                        rows: &[
                            &[T("CREATE INDEX idx1 ON public.t USING btree (upper(name), age, ((c1 + c2)))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "expression-first ordering",
            set_up_script: &[
                "CREATE TABLE t (pk int primary key, c1 int, c2 int, c3 int);",
                "INSERT INTO t VALUES (1, 1, 2, 3), (2, 4, 5, 6);",
                "CREATE INDEX idx1 ON t ((c1 + c2), c3, (c1 * c2));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE c1 + c2 = 3 AND c3 = 3 AND c1 * c2 = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "two expressions sharing columns but with different operators",
            set_up_script: &[
                "CREATE TABLE t (pk int primary key, c1 int, c2 int);",
                "INSERT INTO t VALUES (1, 1, 2), (2, 4, 5);",
                "CREATE INDEX idx1 ON t ((c1 + c2), (c1 * c2));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE c1 + c2 = 3;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE c1 * c2 = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE c1 + c2 = 9;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE c1 * c2 = 20;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "composite range scan across a mixed expression/column key",
            set_up_script: &[
                "CREATE TABLE t (pk int primary key, c1 int, c2 int);",
                "INSERT INTO t VALUES (1, 1, 10), (2, 1, 20), (3, 1, 30), (4, 2, 10);",
                "CREATE INDEX idx1 ON t ((c1 * 10), c2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT pk FROM t WHERE c1 * 10 = 10 AND c2 > 15;",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "t", columns: &["!hidden!idx1!0!0", "c2"], ranges: "[{[10, 10], (15, ∞)}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE c1 * 10 = 10 AND c2 > 15 ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UNIQUE constraint enforced across all expressions",
            set_up_script: &[
                "CREATE TABLE t (pk int primary key, c1 int, c2 int, c3 int);",
                "CREATE UNIQUE INDEX idx1 ON t ((c1 * 10), c2, (c3 * 10));",
                "INSERT INTO t VALUES (1, 1, 2, 3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (2, 1, 2, 3);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "idx1""#, detail: "Key ((c1 * 10), c2, (c3 * 10))=(10, 2, 30) already exists.", schema: "public", table: "t", constraint: "idx1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (3, 1, 3, 3);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "expression index on non-doltgres-native functions work correctly",
            set_up_script: &[
                "CREATE TABLE t_expr (id int PRIMARY KEY, col text);",
                "INSERT INTO t_expr VALUES (1, 'before-index');",
                "CREATE INDEX idx_expr_coalesce ON t_expr ((coalesce(col, '')));",
                "INSERT INTO t_expr VALUES (2, 'after-index');",
                "UPDATE t_expr SET col = 'updated' WHERE id = 1;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, col FROM t_expr ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("col", TEXT)],
                        rows: &[
                            &[T("1"), T("updated")],
                            &[T("2"), T("after-index")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coalesce(col, '') FROM t_expr WHERE id = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("coalesce", TEXT)],
                        rows: &[
                            &[T("after-index")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP INDEX removes only its own hidden columns, a second index survives",
            set_up_script: &[
                "CREATE TABLE t (pk int primary key, c1 int, c2 int, c3 int);",
                "INSERT INTO t VALUES (1, 10, 20, 30);",
                "CREATE INDEX idx2 ON t ((c2 * 100));",
                "CREATE INDEX idx1 ON t ((c1 * 10), c2, (c3 * 10));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE c1 * 10 = 100 AND c2 = 20 AND c3 * 10 = 300;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP INDEX idx1;",
                    expected: Expected::Tag("DROP INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk, c2 * 100 FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("?column?", INT4)],
                        rows: &[
                            &[T("1"), T("2000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP INDEX idx2;",
                    expected: Expected::Tag("DROP INDEX"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
