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
