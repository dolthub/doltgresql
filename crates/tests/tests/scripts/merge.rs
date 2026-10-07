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
fn test_merge() {
    run_scripts(&[
        ScriptTest {
            name: "simple merge",
            set_up_script: &[
                "CREATE TABLE t1 (a INT, b INT, c INT, PRIMARY KEY (a))",
                "CREATE TABLE t2 (a INT, b INT, c INT, PRIMARY KEY (a))",
                "INSERT INTO t1 VALUES (1, 2, 3), (4, 5, 6), (7, 8, 9)",
                "INSERT INTO t2 VALUES (1, 2, 3), (4, 5, 6), (7, 8, 9)",
                "SELECT DOLT_COMMIT('-Am', 'intial commit')",
                "SELECT DOLT_BRANCH('branch1')",
                "SELECT DOLT_BRANCH('branch2')",
                "SELECT DOLT_CHECKOUT('branch1')",
                "INSERT INTO t1 VALUES (10, 11, 12)",
                "INSERT INTO t2 VALUES (10, 11, 12)",
                "SELECT DOLT_COMMIT('-Am', 'added 10')",
                "SELECT DOLT_CHECKOUT('branch2')",
                "INSERT INTO t1 VALUES (20, 21, 22)",
                "INSERT INTO t2 VALUES (20, 21, 22)",
                "SELECT DOLT_COMMIT('-Am', 'added 20')",
                "SELECT DOLT_CHECKOUT('main')",
                "SELECT DOLT_MERGE('branch1')",
                "SELECT DOLT_MERGE('branch2')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t1",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("4"), T("5"), T("6")],
                            &[T("7"), T("8"), T("9")],
                            &[T("10"), T("11"), T("12")],
                            &[T("20"), T("21"), T("22")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t2",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("4"), T("5"), T("6")],
                            &[T("7"), T("8"), T("9")],
                            &[T("10"), T("11"), T("12")],
                            &[T("20"), T("21"), T("22")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "merge independent JSON document changes",
            set_up_script: &[
                "CREATE TABLE documents_json (id INT PRIMARY KEY, doc JSON)",
                "CREATE TABLE documents_jsonb (id INT PRIMARY KEY, doc JSONB)",
                r#"INSERT INTO documents_json VALUES (1, '{"a": 1, "b": 2}')"#,
                r#"INSERT INTO documents_jsonb VALUES (1, '{"a": 1, "b": 2}')"#,
                "SELECT DOLT_COMMIT('-Am', 'base')",
                "SELECT DOLT_BRANCH('left')",
                "SELECT DOLT_BRANCH('right')",
                "SELECT DOLT_CHECKOUT('left')",
                r#"UPDATE documents_json SET doc = '{"a": 100, "b": 2}' WHERE id = 1"#,
                r#"UPDATE documents_jsonb SET doc = doc || '{"a": 100}' WHERE id = 1"#,
                "SELECT DOLT_COMMIT('-am', 'left')",
                "SELECT DOLT_CHECKOUT('right')",
                r#"UPDATE documents_json SET doc = '{"a": 1, "b": 200}' WHERE id = 1"#,
                r#"UPDATE documents_jsonb SET doc = doc || '{"b": 200}' WHERE id = 1"#,
                "SELECT DOLT_COMMIT('-am', 'right')",
                "SELECT DOLT_CHECKOUT('left')",
                "SELECT DOLT_MERGE('right')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT doc::TEXT FROM documents_json",
                    expected: Expected::Rows {
                        columns: &[Column("doc", TEXT)],
                        rows: &[
                            &[T(r#"{"a":100,"b":200}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT doc::TEXT FROM documents_jsonb",
                    expected: Expected::Rows {
                        columns: &[Column("doc", TEXT)],
                        rows: &[
                            &[T(r#"{"a": 100, "b": 200}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_conflicts",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "merge with check expressions and column defaults",
            set_up_script: &[
                "SET timezone TO 'UTC';",
                "CREATE TABLE t1 (a INT, b timestamptz default '2020-01-01 00:00:00'::timestamptz, PRIMARY KEY (a))",
                "ALTER TABLE t1 ADD CONSTRAINT check_b CHECK (b >= '2020-01-01 00:00:00'::timestamptz)",
                "INSERT INTO t1 VALUES (1, '2020-01-02 00:00:00'), (2, '2020-01-03 00:00:00')",
                "SELECT DOLT_COMMIT('-Am', 'intial commit')",
                "SELECT DOLT_BRANCH('branch1')",
                "SELECT DOLT_BRANCH('branch2')",
                "SELECT DOLT_CHECKOUT('branch1')",
                "INSERT INTO t1 VALUES (3, '2020-01-04 00:00:00')",
                "SELECT DOLT_COMMIT('-Am', 'added 3')",
                "SELECT DOLT_CHECKOUT('branch2')",
                "INSERT INTO t1 VALUES (4, '2020-01-05 00:00:00')",
                "SELECT DOLT_COMMIT('-Am', 'added 4')",
                "SELECT DOLT_CHECKOUT('main')",
                "SELECT DOLT_MERGE('branch1')",
                "SELECT DOLT_MERGE('branch2')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t1 order by a",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1"), T("2020-01-02 00:00:00+00")],
                            &[T("2"), T("2020-01-03 00:00:00+00")],
                            &[T("3"), T("2020-01-04 00:00:00+00")],
                            &[T("4"), T("2020-01-05 00:00:00+00")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Postgres reports the violated constraint as it does outside a merge.
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (5, '2019-12-31 00:00:00')",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t1" violates check constraint "check_b""#, detail: "Failing row contains (5, 2019-12-31 00:00:00+00).", schema: "public", table: "t1", constraint: "check_b", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "merge with unique constraints and foreign keys",
            set_up_script: &[
                "CREATE TABLE t1 (a INT, b INT, PRIMARY KEY (a), unique (b))",
                "CREATE TABLE t2 (a INT, b INT, PRIMARY KEY (a), foreign key (b) references t1(b))",
                "INSERT INTO t1 VALUES (1, 2), (4, 5), (7, 8)",
                "INSERT INTO t2 VALUES (1, 2), (4, 5), (7, 8)",
                "SELECT DOLT_COMMIT('-Am', 'intial commit')",
                "SELECT DOLT_BRANCH('branch1')",
                "SELECT DOLT_BRANCH('branch2')",
                "SELECT DOLT_CHECKOUT('branch1')",
                "INSERT INTO t1 VALUES (10, 11)",
                "INSERT INTO t2 VALUES (10, 11)",
                "SELECT DOLT_COMMIT('-Am', 'added 10')",
                "SELECT DOLT_CHECKOUT('branch2')",
                "INSERT INTO t1 VALUES (20, 21)",
                "INSERT INTO t2 VALUES (20, 21)",
                "SELECT DOLT_COMMIT('-Am', 'added 20')",
                "SELECT DOLT_CHECKOUT('main')",
                "SELECT DOLT_MERGE('branch1')",
                "SELECT DOLT_MERGE('branch2')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t1 order by a",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("4"), T("5")],
                            &[T("7"), T("8")],
                            &[T("10"), T("11")],
                            &[T("20"), T("21")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                // Postgres reports the violated constraint as it does outside a merge.
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (100, 2)",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t1_b_key""#, detail: "Key (b)=(2) already exists.", schema: "public", table: "t1", constraint: "t1_b_key", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (100, 200)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "t2" violates foreign key constraint "t2_b_fkey""#, detail: r#"Key (b)=(200) is not present in table "t1"."#, schema: "public", table: "t2", constraint: "t2_b_fkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "merge a branch that created a type",
            set_up_script: &[
                "CREATE TABLE t1 (a INT PRIMARY KEY)",
                "SELECT DOLT_COMMIT('-Am', 'initial commit')",
                "SELECT DOLT_CHECKOUT('-b', 'branch1')",
                "CREATE TYPE type1 AS ENUM ('a', 'b')",
                "CREATE TABLE t2 (a INT PRIMARY KEY, b type1)",
                "INSERT INTO t2 VALUES (1, 'a')",
                "SELECT DOLT_COMMIT('-Am', 'added type1')",
                "SELECT DOLT_CHECKOUT('main')",
                "INSERT INTO t1 VALUES (1)",
                "SELECT DOLT_COMMIT('-Am', 'added 1')",
                "SELECT DOLT_MERGE('branch1')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t2",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT 'b'::type1",
                    expected: Expected::Rows {
                        columns: &[Column("type1", USER_DEFINED)],
                        rows: &[
                            &[T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "merge columns that each side added",
            set_up_script: &[
                "CREATE TABLE a (pk INT PRIMARY KEY, x INT)",
                "INSERT INTO a VALUES (1, 1)",
                "SELECT DOLT_COMMIT('-Am', 'a')",
                "SELECT DOLT_BRANCH('r2')",
                "ALTER TABLE a ADD COLUMN l TEXT",
                "INSERT INTO a VALUES (2, 2, 'left')",
                "SELECT DOLT_COMMIT('-am', 'left a')",
                "SELECT DOLT_CHECKOUT('r2')",
                "ALTER TABLE a ADD COLUMN r INT",
                "UPDATE a SET x = 10, r = 5 WHERE pk = 1",
                "SELECT DOLT_COMMIT('-am', 'right a')",
                "SELECT DOLT_CHECKOUT('main')",
                "SELECT DOLT_MERGE('r2')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM a ORDER BY pk",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("x", INT4), Column("l", TEXT), Column("r", INT4)],
                        rows: &[
                            &[T("1"), T("10"), Null, T("5")],
                            &[T("2"), T("2"), T("left"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
