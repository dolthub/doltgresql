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
fn test_exist_subquery() {
    run_scripts(&[
        ScriptTest {
            name: "basic case",
            set_up_script: &[
                "CREATE TABLE test (id INT PRIMARY KEY);",
                "INSERT INTO test VALUES (1), (3), (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE EXISTS (SELECT 123);",
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
                    query: "SELECT * FROM test WHERE NOT EXISTS (SELECT 123);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 123 WHERE EXISTS (SELECT * FROM test);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("123")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 123 WHERE EXISTS (SELECT * FROM test WHERE id > 10);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 123 WHERE NOT EXISTS (SELECT * FROM test);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 123 WHERE NOT EXISTS (SELECT * FROM test WHERE id > 10);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("123")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "correlated EXISTS decorrelates into a semi join",
            set_up_script: &[
                "CREATE TABLE a (id INT PRIMARY KEY, x INT);",
                "CREATE TABLE b (id INT PRIMARY KEY, x INT);",
                "INSERT INTO a VALUES (1,1),(2,2),(3,3);",
                "INSERT INTO b VALUES (1,1),(2,2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM a WHERE EXISTS (SELECT 1 FROM b WHERE a.x = b.x);",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "SemiJoin", left: "a", right: "b" }, PlanFact::FullScan { table: "a" }, PlanFact::FullScan { table: "b" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM a WHERE EXISTS (SELECT 1 FROM b WHERE a.x = b.x) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("x", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM a WHERE NOT EXISTS (SELECT 1 FROM b WHERE a.x = b.x);",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "LeftOuterJoin", left: "a", right: "b" }, PlanFact::FullScan { table: "a" }, PlanFact::FullScan { table: "b" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM a WHERE NOT EXISTS (SELECT 1 FROM b WHERE a.x = b.x) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("x", INT4)],
                        rows: &[
                            &[T("3"), T("3")],
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
fn test_subqueries() {
    run_scripts(&[
        ScriptTest {
            name: "Subselect",
            set_up_script: &[
                "CREATE TABLE test (id INT);",
                "INSERT INTO test VALUES (1), (3), (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id = (SELECT 2);",
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
                    query: "SELECT *, (SELECT id from test where id = 2) FROM test order by id;",
                    expected: Expected::Error(Diagnostic { code: "42702", message: r#"ORDER BY "id" is ambiguous"#, position: 65, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT *, (SELECT id from test t2 where t2.id = test.id) FROM test order by id;",
                    expected: Expected::Error(Diagnostic { code: "42702", message: r#"ORDER BY "id" is ambiguous"#, position: 77, ..E }),
                    flow: Flow::Query,
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
                    query: "SELECT * FROM test WHERE id IN (SELECT * FROM test WHERE id = 2);",
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
                    query: "SELECT * FROM test WHERE id IN (SELECT id FROM test WHERE id = 3);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id IN (SELECT * FROM test WHERE id > 0);",
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
                    query: "SELECT * FROM test2 WHERE test_id IN (SELECT * FROM test WHERE id > 0);",
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
                    query: "SELECT id FROM test2 WHERE (2, 10) IN (SELECT id, test_id FROM test2 WHERE id > 0);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM test2 WHERE (id, test_id) IN (SELECT id, test_id FROM test2 WHERE id > 0);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "subquery equality",
            set_up_script: &[
                "CREATE TABLE test (id INT, c varchar);",
                "INSERT INTO test VALUES (1, 'a'), (2, 'b'), (3, 'b');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE id = (SELECT id from test where id = 2);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", VARCHAR)],
                        rows: &[
                            &[T("2"), T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT id from test where id = 2) = (SELECT id from test where id = 2);",
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
                    query: "SELECT (SELECT c from test where id = 2) = (SELECT c from test where id = 3);",
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
                    query: "SELECT (SELECT c from test where id = 1) = (SELECT c from test where id = 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array flatten",
            set_up_script: &[
                "CREATE TABLE test (id INT, c varchar);",
                "INSERT INTO test VALUES (1, 'a'), (2, 'b'), (3, 'c');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY(SELECT id FROM test order by 1);",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY(SELECT c FROM test order by id limit 1);",
                    expected: Expected::Rows {
                        columns: &[Column("array", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{a}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY(SELECT c FROM test order by id desc);",
                    expected: Expected::Rows {
                        columns: &[Column("array", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{c,b,a}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY(SELECT id, id FROM test order by 1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "subquery must return only one column", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_string(ARRAY(SELECT id FROM test order by 1), ',')",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("1,2,3")],
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
fn test_subquery_joins() {
    run_scripts(&[
        ScriptTest {
            name: "subquery join",
            set_up_script: &[
                "CREATE TABLE t1 (a int primary key);",
                "CREATE TABLE t2 (b int primary key);",
                "INSERT INTO t1 VALUES (1), (2), (3);",
                "INSERT INTO t2 VALUES (2), (3), (4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT
s1.a FROM (SELECT a from t1) s1
INNER JOIN t2 q1
ON q1.b = s1.a
ORDER BY 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
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
            name: "subquery join with aliased column",
            set_up_script: &[
                "CREATE TABLE t1 (a int primary key);",
                "CREATE TABLE t2 (b int primary key);",
                "INSERT INTO t1 VALUES (1), (2), (3);",
                "INSERT INTO t2 VALUES (2), (3), (4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT
s1.c FROM (SELECT a as c from t1) s1
INNER JOIN t2 q1
ON q1.b = s1.c
ORDER BY 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("c", INT4)],
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
            name: "subquery join with column renames",
            set_up_script: &[
                "CREATE TABLE t1 (a int primary key, b int);",
                "CREATE TABLE t2 (c int primary key);",
                "INSERT INTO t1 VALUES (1,10), (2,20), (3,30);",
                "INSERT INTO t2 VALUES (2), (3), (4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT
s1.d FROM (SELECT b as f, a as g from t1) s1(d,e)
INNER JOIN t2 q1
ON q1.c = s1.e
ORDER BY 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("d", INT4)],
                        rows: &[
                            &[T("20")],
                            &[T("30")],
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
