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
fn test_binary_logic() {
    run_scripts(&[
        ScriptTest {
            name: "AND",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1 = 1 AND 2 = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1 = 1 AND 2 = 2) AND (false);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1 > 1 AND 2 = 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1 = 1 AND 2 = 2) AND (false);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1 = 1 AND 2 = 2) AND (true);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "OR",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1 = 1 OR 2 = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1 = 1 AND 2 = 2) OR (false);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1 > 1 OR 2 = 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1 > 1 OR 2 > 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1 > 1 OR 2 > 2) OR (true);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1 = 1 AND 2 = 2) OR (true);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "IS DISTINCT FROM",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1 IS DISTINCT FROM 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS DISTINCT FROM 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT null IS DISTINCT FROM 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT null IS DISTINCT FROM null;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS DISTINCT FROM null;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS DISTINCT FROM 2.5;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS DISTINCT FROM '2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS DISTINCT FROM 'a';",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 27, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "IS NOT DISTINCT FROM",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1 IS NOT DISTINCT FROM 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS NOT DISTINCT FROM 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT null IS NOT DISTINCT FROM 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT null IS NOT DISTINCT FROM null;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS NOT DISTINCT FROM null;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS NOT DISTINCT FROM 2.5;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS NOT DISTINCT FROM '2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS NOT DISTINCT FROM 'a';",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 31, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "IS DISTINCT FROM and IS NOT DISTINCT FROM inside a subquery",
            set_up_script: &[
                "CREATE TABLE t_indf (id INT PRIMARY KEY, v TEXT);",
                "INSERT INTO t_indf VALUES (1, 'a'), (2, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM t_indf o WHERE EXISTS (SELECT 1 FROM t_indf i WHERE i.v IS NOT DISTINCT FROM o.v);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM t_indf o WHERE EXISTS (SELECT 1 FROM t_indf i WHERE i.v IS DISTINCT FROM o.v);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
    ]);
}

#[test]
fn test_case() {
    run_scripts(&[
        ScriptTest {
            name: "CASE with mixed numeric column and integer literal branches",
            set_up_script: &[
                "CREATE TABLE t (status text, price numeric(10,2));",
                "INSERT INTO t VALUES ('confirmed', 100.00), ('confirmed', 20.50), ('pending', 7.00);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(CASE WHEN status='confirmed' THEN price ELSE 0 END)::text FROM t LIMIT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", TEXT)],
                        rows: &[
                            &[T("numeric")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT SUM(CASE WHEN status='confirmed' THEN price ELSE 0 END) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("120.50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT SUM(CASE WHEN status='confirmed' THEN price ELSE 0::numeric END) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("120.50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT SUM(CASE WHEN status='confirmed' THEN price ELSE CAST(0 AS NUMERIC(10,2)) END) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("120.50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT SUM(CASE WHEN status='confirmed' THEN price END) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("120.50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT MAX(CASE WHEN status='confirmed' THEN price ELSE 0 END) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("max", NUMERIC)],
                        rows: &[
                            &[T("100.00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT MIN(CASE WHEN status='confirmed' THEN price ELSE 0 END) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("min", NUMERIC)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT SUM(CASE WHEN status='confirmed' THEN 1 ELSE 0 END) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT8)],
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
    ]);
}

#[test]
fn test_coalesce() {
    run_scripts(&[
        ScriptTest {
            name: "COALESCE(NULL, col) in UPDATE",
            set_up_script: &[
                "CREATE TABLE t (id UUID PRIMARY KEY, val INTEGER NOT NULL DEFAULT 0, d DATE)",
                "INSERT INTO t VALUES ('00000000-0000-0000-0000-000000000001', 42, '2026-01-01')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t SET val = COALESCE(NULL, val) WHERE id = '00000000-0000-0000-0000-000000000001'",
                    expected: Expected::Tag("UPDATE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t WHERE id = '00000000-0000-0000-0000-000000000001'",
                    expected: Expected::Rows {
                        columns: &[Column("val", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET d = COALESCE(NULL, d) WHERE id = '00000000-0000-0000-0000-000000000001'",
                    expected: Expected::Tag("UPDATE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT d FROM t WHERE id = '00000000-0000-0000-0000-000000000001'",
                    expected: Expected::Rows {
                        columns: &[Column("d", DATE)],
                        rows: &[
                            &[T("2026-01-01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "COALESCE type resolution in SELECT",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT COALESCE(NULL, 42)",
                    expected: Expected::Rows {
                        columns: &[Column("coalesce", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COALESCE(NULL, NULL)",
                    expected: Expected::Rows {
                        columns: &[Column("coalesce", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COALESCE(NULL, NULL, 'hello')",
                    expected: Expected::Rows {
                        columns: &[Column("coalesce", TEXT)],
                        rows: &[
                            &[T("hello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COALESCE(1, 2, 3)",
                    expected: Expected::Rows {
                        columns: &[Column("coalesce", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COALESCE(NULL, 2, 3)",
                    expected: Expected::Rows {
                        columns: &[Column("coalesce", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COALESCE(NULL::integer, 42)",
                    expected: Expected::Rows {
                        columns: &[Column("coalesce", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "COALESCE with mixed types",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT COALESCE('a'::TEXT, 1::INTEGER)",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "COALESCE types text and integer cannot be matched", position: 28, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COALESCE(NULL, 1::SMALLINT, 2::BIGINT) AS v;",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT8)],
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
    ]);
}

#[test]
fn test_in() {
    run_scripts(&[
        ScriptTest {
            name: "ANY",
            set_up_script: &[
                "CREATE TABLE test (id INT);",
                "INSERT INTO test VALUES (1), (3), (2);",
                "CREATE TABLE test2 (id INT PRIMARY KEY, test_id INT, txt text);",
                "INSERT INTO test2 VALUES (1, 1, 'foo'), (2, 10, 'bar'), (3, 2, 'baz');",
                "CREATE TABLE test3 (id INT PRIMARY KEY, carr smallint[]);",
                "INSERT INTO test3 VALUES (1, ARRAY[1, 2, 3]), (2, ARRAY[4, 5, 6]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 3 = ANY (ARRAY[1, 2, 3, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3 = ANY (ARRAY[1, 2, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a' = ANY (ARRAY['c', 'a', 't']);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a' = ANY (ARRAY['c', 'at', 't']);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3 = ANY (ARRAY[1.0, 2.1, 3.0, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6 > ANY (ARRAY[1, 2, 3, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6 < ANY (ARRAY[1, 2, 3, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6 <= ANY (ARRAY[1, 2, 3, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6 >= ANY (ARRAY[1, 2, 3, 6, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id = ANY(ARRAY[2, 3, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id = ANY(ARRAY[4, 3, 2, 1, 0]);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id = ANY(ARRAY[4, 5, 6]);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM test3 WHERE 4 = ANY(carr);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id = ANY(SELECT * FROM test WHERE id = 2);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("3"), T("2"), T("baz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id = ANY(SELECT * FROM test WHERE id = 10);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id = ANY(SELECT * FROM test WHERE id > 1) AND txt = 'baz';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("3"), T("2"), T("baz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id > ANY(SELECT * FROM test);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("2"), T("10"), T("bar")],
                            &[T("3"), T("2"), T("baz")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id = ANY(SELECT * FROM test WHERE id > 0);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("foo")],
                            &[T("3"), T("2"), T("baz")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT "ns"."nspname" AS "table_schema",
       "t"."relname" AS "table_name",
       "cnst"."conname" AS "constraint_name",
       pg_get_constraintdef("cnst"."oid") AS "expression",
       CASE "cnst"."contype" 
           WHEN 'p' THEN 'PRIMARY'
           WHEN 'u' THEN 'UNIQUE'
           WHEN 'c' THEN 'CHECK'
           WHEN 'x' THEN 'EXCLUDE'
           END AS "constraint_type", 
    "a"."attname" AS "column_name" 
FROM "pg_catalog"."pg_constraint" "cnst" 
    INNER JOIN "pg_catalog"."pg_class" "t" ON "t"."oid" = "cnst"."conrelid"
    INNER JOIN "pg_catalog"."pg_namespace" "ns" ON "ns"."oid" = "cnst"."connamespace"
    LEFT JOIN "pg_catalog"."pg_attribute" "a" ON "a"."attrelid" = "cnst"."conrelid" AND "a"."attnum" = ANY ("cnst"."conkey")
WHERE "t"."relkind" IN ('r', 'p') AND (("ns"."nspname" = 'public' AND "t"."relname" = 'test2'));"#,
                    expected: Expected::Rows {
                        columns: &[Column("table_schema", NAME), Column("table_name", NAME), Column("constraint_name", NAME), Column("expression", TEXT), Column("constraint_type", TEXT), Column("column_name", NAME)],
                        rows: &[
                            &[T("public"), T("test2"), T("test2_pkey"), T("PRIMARY KEY (id)"), T("PRIMARY"), T("id")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SOME",
            set_up_script: &[
                "CREATE TABLE test (id INT);",
                "INSERT INTO test VALUES (1), (3), (2);",
                "CREATE TABLE test2 (id INT PRIMARY KEY, test_id INT, txt text);",
                "INSERT INTO test2 VALUES (1, 1, 'foo'), (2, 10, 'bar'), (3, 2, 'baz');",
                "CREATE TABLE test3 (id INT PRIMARY KEY, carr smallint[]);",
                "INSERT INTO test3 VALUES (1, ARRAY[1, 2, 3]), (2, ARRAY[4, 5, 6]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 3 = SOME (ARRAY[1, 2, 3, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3 = SOME (ARRAY[1, 2, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a' = SOME (ARRAY['c', 'a', 't']);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a' = SOME (ARRAY['c', 'at', 't']);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3 = SOME (ARRAY[1.0, 2.1, 3.0, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6 > SOME (ARRAY[1, 2, 3, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6 < SOME (ARRAY[1, 2, 3, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6 <= SOME (ARRAY[1, 2, 3, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6 >= SOME (ARRAY[1, 2, 3, 6, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id = SOME(ARRAY[2, 3, 4, 5]);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id = SOME(ARRAY[4, 3, 2, 1, 0]);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id = SOME(ARRAY[4, 5, 6]);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM test3 WHERE 4 = SOME(carr);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id = SOME(SELECT * FROM test WHERE id = 2);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("3"), T("2"), T("baz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id = SOME(SELECT * FROM test WHERE id = 10);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id = SOME(SELECT * FROM test WHERE id > 1) AND txt = 'baz';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("3"), T("2"), T("baz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id > SOME(SELECT * FROM test);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("2"), T("10"), T("bar")],
                            &[T("3"), T("2"), T("baz")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id = SOME(SELECT * FROM test WHERE id > 0);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("foo")],
                            &[T("3"), T("2"), T("baz")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT "ns"."nspname" AS "table_schema",
       "t"."relname" AS "table_name",
       "cnst"."conname" AS "constraint_name",
       pg_get_constraintdef("cnst"."oid") AS "expression",
       CASE "cnst"."contype" 
           WHEN 'p' THEN 'PRIMARY'
           WHEN 'u' THEN 'UNIQUE'
           WHEN 'c' THEN 'CHECK'
           WHEN 'x' THEN 'EXCLUDE'
           END AS "constraint_type", 
    "a"."attname" AS "column_name" 
FROM "pg_catalog"."pg_constraint" "cnst" 
    INNER JOIN "pg_catalog"."pg_class" "t" ON "t"."oid" = "cnst"."conrelid"
    INNER JOIN "pg_catalog"."pg_namespace" "ns" ON "ns"."oid" = "cnst"."connamespace"
    LEFT JOIN "pg_catalog"."pg_attribute" "a" ON "a"."attrelid" = "cnst"."conrelid" AND "a"."attnum" = SOME ("cnst"."conkey")
WHERE "t"."relkind" IN ('r', 'p') AND (("ns"."nspname" = 'public' AND "t"."relname" = 'test2'));"#,
                    expected: Expected::Rows {
                        columns: &[Column("table_schema", NAME), Column("table_name", NAME), Column("constraint_name", NAME), Column("expression", TEXT), Column("constraint_type", TEXT), Column("column_name", NAME)],
                        rows: &[
                            &[T("public"), T("test2"), T("test2_pkey"), T("PRIMARY KEY (id)"), T("PRIMARY"), T("id")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "IN",
            set_up_script: &[
                "CREATE TABLE test (id INT);",
                "INSERT INTO test VALUES (1), (3), (2);",
                "CREATE TABLE test2 (id INT, test_id INT, txt text);",
                "INSERT INTO test2 VALUES (1, 1, 'foo'), (2, 10, 'bar'), (3, 2, 'baz');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id IN (2, 3, 4, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id IN (4, 3, 2, 1, 0);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id IN (SELECT * FROM test WHERE id = 2);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("3"), T("2"), T("baz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id IN(SELECT * FROM test WHERE id > 0);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("foo")],
                            &[T("3"), T("2"), T("baz")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4 IN (null, 1, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4 IN (null, 1, 2, 3, 4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL IN (null, 1, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4 IN (1, 2, 3, null::int4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4 IN (1, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4 IN (1, 2, 3, 4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT concat('a', 'b') in ('a', 'b', 'ab');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT concat('a', 'b') in ('a', 'b');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT concat('a', 'b') in ('a', NULL, 'b');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT concat('a', NULL) in ('a', 'b', 'ab');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT concat('a', NULL) in ('a', NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "NOT IN",
            set_up_script: &[
                "CREATE TABLE test (id INT);",
                "INSERT INTO test VALUES (1), (3), (2);",
                "CREATE TABLE test2 (id INT, test_id INT, txt text);",
                "INSERT INTO test2 VALUES (1, 1, 'foo'), (2, 10, 'bar'), (3, 2, 'baz');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id NOT IN (2, 3, 4, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id NOT IN (SELECT * FROM test WHERE id = 2);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("foo")],
                            &[T("2"), T("10"), T("bar")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 WHERE test_id NOT IN (SELECT * FROM test WHERE id > 0);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("test_id", INT4), Column("txt", TEXT)],
                        rows: &[
                            &[T("2"), T("10"), T("bar")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4 NOT IN (null, 1, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL NOT IN (null, 1, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4 NOT IN (1, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4 NOT IN (1, 2, 3, 4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT concat('a', 'b') NOT in ('a', 'b');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
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

#[test]
fn test_subscript() {
    run_scripts(&[
        ScriptTest {
            name: "array literal",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1, 2, 3][1];",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "[""#, position: 22, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[1, 2, 3])[3];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[1, 2, 3])[1+1];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1, 2, 3][0];",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "[""#, position: 22, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1, 2, 3][4];",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "[""#, position: 22, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1, 2, 3][null];",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "[""#, position: 22, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY['a', 'b', 'c'][2];",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "[""#, position: 28, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1, 2, 3][1:3];",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "[""#, position: 22, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1, 2, 3]['abc'];",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "[""#, position: 22, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array column",
            set_up_script: &[
                "CREATE TABLE test (id INT, arr INT[]);",
                "INSERT INTO test VALUES (1, ARRAY[1, 2, 3]), (2, ARRAY[4, 5, 6]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT arr[2] FROM test order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("arr", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array subquery",
            set_up_script: &[
                "CREATE TABLE test (id INT);",
                "INSERT INTO test VALUES (1), (2), (3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT (array(select id from test order by 1))[2]",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4)],
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
            name: "null index on a non-array value",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ('123'::jsonb)[NULL];",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ('{"a": 1}'::jsonb)[NULL];"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[Null],
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

#[test]
fn test_jsonb_subscript_rules() {
    run_scripts(&[
        ScriptTest {
            name: "jsonb subscripts",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ('123'::jsonb)[NULL];",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ('{"a": {"b": [10, 20]}}'::jsonb)['a']['b'][1];"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ('{"a": {"b": [10, 20]}}'::jsonb)['a']['b']['1'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('[10, 20]'::jsonb)[-1];",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ('{"1": 5}'::jsonb)[1];"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ('{"a": 1}'::jsonb)['a':'b'];"#,
                    expected: Expected::Error(Diagnostic { code: "42804", message: "jsonb subscript does not support slices", position: 32, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ('{"a": 1}'::jsonb)[1.5];"#,
                    expected: Expected::Error(Diagnostic { code: "42804", message: "subscript type numeric is not supported", hint: "jsonb subscript must be coercible to either integer or text.", position: 28, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ('{"a": 1}'::jsonb)[true];"#,
                    expected: Expected::Error(Diagnostic { code: "42804", message: "subscript type boolean is not supported", hint: "jsonb subscript must be coercible to either integer or text.", position: 28, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT pg_typeof(('{"a": 1}'::jsonb)['a']);"#,
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("jsonb")],
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

#[test]
fn test_unique_clause_and_rename_rules() {
    run_scripts(&[
        ScriptTest {
            name: "Keyless unique indexes and repeated rows",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t (iid uuid, slug text, UNIQUE(iid, slug));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t_iid_slug_key""#, detail: "Key (iid, slug)=(11111111-1111-1111-1111-111111111111, hello) already exists.", schema: "public", table: "t", constraint: "t_iid_slug_key", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t2 (iid uuid, slug text, UNIQUE(iid, slug));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES ('22222222-2222-2222-2222-222222222222', 'hello');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t2_iid_slug_key""#, detail: "Key (iid, slug)=(11111111-1111-1111-1111-111111111111, hello) already exists.", schema: "public", table: "t2", constraint: "t2_iid_slug_key", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t3 (a int, slug text, UNIQUE(a, slug));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3 VALUES (1, 'hello'), (2, 'hello');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3 VALUES (1, 'hello');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t3_a_slug_key""#, detail: "Key (a, slug)=(1, hello) already exists.", schema: "public", table: "t3", constraint: "t3_a_slug_key", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Clause, domain, and subscript rules",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'a' || '2020-01-01 00:00:00'::timestamp;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("a2020-01-01 00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT now()::date || 'x';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("2026-10-07x")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'x' || 1.5::numeric || 'y' || true;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("x1.5ytrue")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'nan'::numeric / '0'::numeric, 'nan'::numeric % '0'::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", NUMERIC), Column("?column?", NUMERIC)],
                        rows: &[
                            &[T("NaN"), T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM current_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM current_user AS u;",
                    expected: Expected::Rows {
                        columns: &[Column("u", NAME)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT group_concat(1 ORDER BY 1);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function group_concat(integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lower('a' ORDER BY 1);",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "ORDER BY specified, but lower is not an aggregate function", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lower(DISTINCT 'a');",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "DISTINCT specified, but lower is not an aggregate function", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lower('a') FILTER (WHERE true);",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "FILTER specified, but lower is not an aggregate function", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t2 (c1 int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t2 ALTER COLUMN c1 TYPE RECORD;",
                    expected: Expected::Error(Diagnostic { code: "42P16", message: r#"column "c1" has pseudo-type record"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION IF EXISTS public.tax_job_trans(t public.trans);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    notices: &[Diagnostic { code: "00000", message: r#"type "public.trans" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN domint4arr AS int4[];",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE domarr (pk int primary key, i domint4arr);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO domarr VALUES (1, '{1,2,3}');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT i[2], pg_typeof(i[2]), i[1:2] FROM domarr;",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("pg_typeof", REGTYPE), Column("i", INT4_ARRAY)],
                        rows: &[
                            &[T("2"), T("integer"), T("{1,2}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t_scalar (n int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t_scalar SET n[1]=7;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "cannot subscript type integer because it does not support subscripting", position: 21, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE jorder (val jsonb);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO jorder VALUES ('null'), ('[]'), ('{}'), ('"a"'), ('1'), ('true'), ('[1]'), ('false'), ('{"a":1}');"#,
                    expected: Expected::Tag("INSERT 0 9"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM jorder ORDER BY val;",
                    expected: Expected::Rows {
                        columns: &[Column("val", JSONB)],
                        rows: &[
                            &[T("[]")],
                            &[T("null")],
                            &[T(r#""a""#)],
                            &[T("1")],
                            &[T("false")],
                            &[T("true")],
                            &[T("[1]")],
                            &[T("{}")],
                            &[T(r#"{"a": 1}"#)],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Renamed columns of row types in use",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t1a (a INT4, b VARCHAR(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t2 (id SERIAL, t1a t1a);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 (t1a) VALUES (ROW(1, 'abc'));",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a RENAME COLUMN a TO z;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (t1a).a FROM t2;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "a" not found in data type t1a"#, position: 9, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (t1a).z, t1a FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("z", INT4), Column("t1a", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(1,abc)")],
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

#[test]
fn test_literal_error_position_rules() {
    run_scripts(&[
        ScriptTest {
            name: "positions of unreadable literals",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 2 = 'a';",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 12, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a' = 2;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IS DISTINCT FROM 'a';",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 27, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 IN (1, 'a');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nullif(2, 'a');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 18, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 + 'a';",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 12, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = ROW(1, 'two');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "two""#, position: 27, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2020-13-01'::date = current_date;",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2020-13-01""#, hint: r#"Perhaps you need a different "datestyle" setting."#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_date = '2020-13-01';",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2020-13-01""#, hint: r#"Perhaps you need a different "datestyle" setting."#, position: 23, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
