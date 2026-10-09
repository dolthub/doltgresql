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
                    expected: Expected::Plan(&[PlanFact::Join { kind: "AntiJoin", left: "a", right: "b" }, PlanFact::FullScan { table: "a" }, PlanFact::FullScan { table: "b" }]),
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

#[test]
fn test_array_subqueries() {
    run_scripts(&[
        ScriptTest {
            name: "ARRAY subqueries",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t (a int, b text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, 'x'), (2, 'y'), (3, NULL);",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY(SELECT a FROM t ORDER BY a DESC), ARRAY(SELECT b FROM t WHERE a > 5), ARRAY(SELECT b FROM t ORDER BY a);",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY), Column("array", TEXT_ARRAY), Column("array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{3,2,1}"), T("{}"), T("{x,y,NULL}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, ARRAY(SELECT t2.a FROM t t2 WHERE t2.a < t.a) FROM t ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{}")],
                            &[T("2"), T("{1}")],
                            &[T("3"), T("{1,2}")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY(SELECT a, b FROM t);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "subquery must return only one column", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(ARRAY(SELECT b FROM t));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("text[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW av AS SELECT ARRAY(SELECT 1) AS arr;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_viewdef('av');",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_viewdef", TEXT)],
                        rows: &[
                            &[T(" SELECT ARRAY( SELECT 1) AS arr;")],
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
fn test_subquery_evaluation_rules() {
    run_scripts(&[
        ScriptTest {
            name: "uncorrelated and correlated subqueries",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE sa (a int, b text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE sb (a int, b text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO sa VALUES (1, 'x'), (2, 'y'), (3, NULL), (NULL, 'w');",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO sb VALUES (1, 'x'), (2, 'q'), (2, 'r'), (5, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, a IN (SELECT a FROM sb), a NOT IN (SELECT a FROM sb WHERE a IS NOT NULL) FROM sa ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("t"), T("f")],
                            &[T("2"), T("t"), T("f")],
                            &[T("3"), T("f"), T("t")],
                            &[Null, Null, Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO sb VALUES (NULL, 'n');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, a IN (SELECT a FROM sb), a NOT IN (SELECT a FROM sb) FROM sa ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("t"), T("f")],
                            &[T("2"), T("t"), T("f")],
                            &[T("3"), Null, Null],
                            &[Null, Null, Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b IN (SELECT b FROM sb) FROM sa ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("t")],
                            &[T("2"), Null],
                            &[T("3"), Null],
                            &[Null, Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a FROM sa WHERE a IN (SELECT a FROM sb WHERE false) ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, EXISTS (SELECT 1 FROM sb WHERE sb.a = sa.a), (SELECT count(*) FROM sb WHERE sb.a = sa.a) FROM sa ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("exists", BOOL), Column("count", INT8)],
                        rows: &[
                            &[T("1"), T("t"), T("1")],
                            &[T("2"), T("t"), T("2")],
                            &[T("3"), T("f"), T("0")],
                            &[Null, T("f"), T("0")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a FROM sa WHERE EXISTS (SELECT 1 FROM sb WHERE sb.a = sa.a AND sb.b <> 'q') ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a FROM sa WHERE NOT EXISTS (SELECT 1 FROM sb WHERE sb.a = sa.a + 0 AND NOT EXISTS (SELECT 1 FROM sb c WHERE c.b = sa.b)) ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, ARRAY(SELECT b FROM sb WHERE sb.a = sa.a ORDER BY b) FROM sa ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("array", TEXT_ARRAY)],
                        rows: &[
                            &[T("1"), T("{x}")],
                            &[T("2"), T("{q,r}")],
                            &[T("3"), T("{}")],
                            &[Null, T("{}")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT max(a) FROM sb), a FROM sa WHERE a < (SELECT max(a) FROM sb) ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("max", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("5"), T("1")],
                            &[T("5"), T("2")],
                            &[T("5"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, (SELECT b FROM sb WHERE sb.a = sa.a) FROM sa ORDER BY a;",
                    expected: Expected::Error(Diagnostic { code: "21000", message: "more than one row returned by a subquery used as an expression", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_regtype syntax errors",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT to_regtype('integer"');"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"unterminated quoted identifier at or near """"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('23');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "23""#, position: 1, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('int4 x');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "x""#, position: 6, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regclass('23'), to_regproc('23'), to_regrole('23'), to_regnamespace('23');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regclass", REGCLASS), Column("to_regproc", REGPROC), Column("to_regrole", REGROLE), Column("to_regnamespace", REGNAMESPACE)],
                        rows: &[
                            &[Null, Null, Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('23');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "expected a left parenthesis", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_hashed_correlated_subqueries() {
    run_scripts(&[
        ScriptTest {
            name: "correlated subqueries over hashed scans",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE so (id INT PRIMARY KEY, c INT, amt INT, note TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE sc (id INT PRIMARY KEY, name TEXT, lim INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO so SELECT i, i % 20, i % 50, 'n' || i FROM generate_series(1, 400) i;",
                    expected: Expected::Tag("INSERT 0 400"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO sc SELECT i, 'c' || i, i * 2 FROM generate_series(0, 24) i;",
                    expected: Expected::Tag("INSERT 0 25"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM sc WHERE EXISTS (SELECT 1 FROM so WHERE so.c = sc.id AND so.amt > 45);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM sc WHERE NOT EXISTS (SELECT 1 FROM so WHERE so.c = sc.id) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("20")],
                            &[T("21")],
                            &[T("22")],
                            &[T("23")],
                            &[T("24")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, (SELECT count(*) FROM so WHERE so.c = sc.id AND so.amt < sc.lim) FROM sc ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("count", INT8)],
                        rows: &[
                            &[T("0"), T("0")],
                            &[T("1"), T("4")],
                            &[T("2"), T("4")],
                            &[T("3"), T("4")],
                            &[T("4"), T("4")],
                            &[T("5"), T("4")],
                            &[T("6"), T("4")],
                            &[T("7"), T("4")],
                            &[T("8"), T("4")],
                            &[T("9"), T("4")],
                            &[T("10"), T("8")],
                            &[T("11"), T("12")],
                            &[T("12"), T("12")],
                            &[T("13"), T("12")],
                            &[T("14"), T("12")],
                            &[T("15"), T("12")],
                            &[T("16"), T("12")],
                            &[T("17"), T("12")],
                            &[T("18"), T("12")],
                            &[T("19"), T("12")],
                            &[T("20"), T("0")],
                            &[T("21"), T("0")],
                            &[T("22"), T("0")],
                            &[T("23"), T("0")],
                            &[T("24"), T("0")],
                        ],
                        tag: "SELECT 25",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, name FROM sc WHERE EXISTS (SELECT 1 FROM so WHERE so.c = sc.id AND so.note LIKE 'n1%' AND so.amt = sc.lim) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("0"), T("c0")],
                            &[T("10"), T("c10")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sc.id, (SELECT max(amt) FROM so WHERE so.c = sc.id + 0 AND so.amt * 2 > sc.lim) FROM sc WHERE sc.id < 6 ORDER BY sc.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("max", INT4)],
                        rows: &[
                            &[T("0"), T("40")],
                            &[T("1"), T("41")],
                            &[T("2"), T("42")],
                            &[T("3"), T("43")],
                            &[T("4"), T("44")],
                            &[T("5"), T("45")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM so WHERE so.amt IN (SELECT lim FROM sc WHERE sc.id = so.c);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("8")],
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
fn test_exists_pull_up() {
    run_scripts(&[
        ScriptTest {
            name: "Ordered scans of indexes that read most rows",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE ordered_big (id INT PRIMARY KEY, v INT, pad TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ordered_big SELECT g, (g * 7919) % 2000, repeat('x', 60) FROM generate_series(1, 2000) g;",
                    expected: Expected::Tag("INSERT 0 2000"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ordered_big_v ON ordered_big (v);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v FROM ordered_big WHERE v >= 0 ORDER BY v LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("2000"), T("0")],
                            &[T("1679"), T("1")],
                            &[T("1358"), T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v FROM ordered_big WHERE v >= 0 ORDER BY v DESC LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("321"), T("1999")],
                            &[T("642"), T("1998")],
                            &[T("963"), T("1997")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), sum(v) FROM ordered_big WHERE v > 10;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("1989"), T("1998945")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM ordered_big WHERE v > 10 AND pad LIKE 'x%';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1989")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "EXISTS pulled up into joins",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE pull_a (id INT PRIMARY KEY, k INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE pull_b (id INT PRIMARY KEY, k INT, g INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE pull_c (id INT PRIMARY KEY, g INT, k INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO pull_a SELECT g, g % 10 FROM generate_series(1, 50) g;",
                    expected: Expected::Tag("INSERT 0 50"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO pull_b SELECT g, g % 10, g % 4 FROM generate_series(1, 40) g;",
                    expected: Expected::Tag("INSERT 0 40"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO pull_c SELECT g, g % 4, g % 7 FROM generate_series(1, 20) g;",
                    expected: Expected::Tag("INSERT 0 20"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.id FROM pull_a a WHERE EXISTS (SELECT 1 FROM pull_b b JOIN pull_c c ON b.g = c.g WHERE b.k = a.k AND c.k = 3) ORDER BY a.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                            &[T("7")],
                            &[T("8")],
                            &[T("9")],
                            &[T("10")],
                            &[T("11")],
                            &[T("12")],
                            &[T("13")],
                            &[T("14")],
                            &[T("15")],
                            &[T("16")],
                            &[T("17")],
                            &[T("18")],
                            &[T("19")],
                            &[T("20")],
                            &[T("21")],
                            &[T("22")],
                            &[T("23")],
                            &[T("24")],
                            &[T("25")],
                            &[T("26")],
                            &[T("27")],
                            &[T("28")],
                            &[T("29")],
                            &[T("30")],
                            &[T("31")],
                            &[T("32")],
                            &[T("33")],
                            &[T("34")],
                            &[T("35")],
                            &[T("36")],
                            &[T("37")],
                            &[T("38")],
                            &[T("39")],
                            &[T("40")],
                            &[T("41")],
                            &[T("42")],
                            &[T("43")],
                            &[T("44")],
                            &[T("45")],
                            &[T("46")],
                            &[T("47")],
                            &[T("48")],
                            &[T("49")],
                            &[T("50")],
                        ],
                        tag: "SELECT 50",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.id FROM pull_a a WHERE NOT EXISTS (SELECT 1 FROM pull_b b WHERE b.k = a.k AND b.id > 35) ORDER BY a.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("11")],
                            &[T("12")],
                            &[T("13")],
                            &[T("14")],
                            &[T("15")],
                            &[T("21")],
                            &[T("22")],
                            &[T("23")],
                            &[T("24")],
                            &[T("25")],
                            &[T("31")],
                            &[T("32")],
                            &[T("33")],
                            &[T("34")],
                            &[T("35")],
                            &[T("41")],
                            &[T("42")],
                            &[T("43")],
                            &[T("44")],
                            &[T("45")],
                        ],
                        tag: "SELECT 25",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.id, b.id FROM pull_a a, pull_b b WHERE a.k = b.k AND a.id < 12 AND EXISTS (SELECT 1 FROM pull_c c WHERE c.g = b.g AND NOT EXISTS (SELECT 1 FROM pull_c d WHERE d.k = a.k AND d.id = c.id)) ORDER BY a.id, b.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("11")],
                            &[T("1"), T("21")],
                            &[T("1"), T("31")],
                            &[T("2"), T("2")],
                            &[T("2"), T("12")],
                            &[T("2"), T("22")],
                            &[T("2"), T("32")],
                            &[T("3"), T("3")],
                            &[T("3"), T("13")],
                            &[T("3"), T("23")],
                            &[T("3"), T("33")],
                            &[T("4"), T("4")],
                            &[T("4"), T("14")],
                            &[T("4"), T("24")],
                            &[T("4"), T("34")],
                            &[T("5"), T("5")],
                            &[T("5"), T("15")],
                            &[T("5"), T("25")],
                            &[T("5"), T("35")],
                            &[T("6"), T("6")],
                            &[T("6"), T("16")],
                            &[T("6"), T("26")],
                            &[T("6"), T("36")],
                            &[T("7"), T("7")],
                            &[T("7"), T("17")],
                            &[T("7"), T("27")],
                            &[T("7"), T("37")],
                            &[T("8"), T("8")],
                            &[T("8"), T("18")],
                            &[T("8"), T("28")],
                            &[T("8"), T("38")],
                            &[T("9"), T("9")],
                            &[T("9"), T("19")],
                            &[T("9"), T("29")],
                            &[T("9"), T("39")],
                            &[T("10"), T("10")],
                            &[T("10"), T("20")],
                            &[T("10"), T("30")],
                            &[T("10"), T("40")],
                            &[T("11"), T("1")],
                            &[T("11"), T("11")],
                            &[T("11"), T("21")],
                            &[T("11"), T("31")],
                        ],
                        tag: "SELECT 44",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.id FROM pull_a a WHERE EXISTS (SELECT 1 FROM pull_b b WHERE b.k = a.k AND b.g > a.id % 3 AND b.id IN (SELECT id FROM pull_c)) ORDER BY a.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                            &[T("7")],
                            &[T("9")],
                            &[T("10")],
                            &[T("11")],
                            &[T("12")],
                            &[T("13")],
                            &[T("15")],
                            &[T("16")],
                            &[T("17")],
                            &[T("18")],
                            &[T("19")],
                            &[T("21")],
                            &[T("22")],
                            &[T("23")],
                            &[T("24")],
                            &[T("25")],
                            &[T("27")],
                            &[T("28")],
                            &[T("29")],
                            &[T("30")],
                            &[T("31")],
                            &[T("33")],
                            &[T("34")],
                            &[T("35")],
                            &[T("36")],
                            &[T("37")],
                            &[T("39")],
                            &[T("40")],
                            &[T("41")],
                            &[T("42")],
                            &[T("43")],
                            &[T("45")],
                            &[T("46")],
                            &[T("47")],
                            &[T("48")],
                            &[T("49")],
                        ],
                        tag: "SELECT 41",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.id FROM pull_a a WHERE EXISTS (SELECT 1 FROM pull_b b WHERE b.k = a.k AND (SELECT count(*) FROM pull_c c WHERE c.g = b.g) > 4) ORDER BY a.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                            &[T("7")],
                            &[T("8")],
                            &[T("9")],
                            &[T("10")],
                            &[T("11")],
                            &[T("12")],
                            &[T("13")],
                            &[T("14")],
                            &[T("15")],
                            &[T("16")],
                            &[T("17")],
                            &[T("18")],
                            &[T("19")],
                            &[T("20")],
                            &[T("21")],
                            &[T("22")],
                            &[T("23")],
                            &[T("24")],
                            &[T("25")],
                            &[T("26")],
                            &[T("27")],
                            &[T("28")],
                            &[T("29")],
                            &[T("30")],
                            &[T("31")],
                            &[T("32")],
                            &[T("33")],
                            &[T("34")],
                            &[T("35")],
                            &[T("36")],
                            &[T("37")],
                            &[T("38")],
                            &[T("39")],
                            &[T("40")],
                            &[T("41")],
                            &[T("42")],
                            &[T("43")],
                            &[T("44")],
                            &[T("45")],
                            &[T("46")],
                            &[T("47")],
                            &[T("48")],
                            &[T("49")],
                            &[T("50")],
                        ],
                        tag: "SELECT 50",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pull_a a WHERE EXISTS (SELECT 1 FROM pull_b b WHERE b.k = a.k + 100);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pull_a a WHERE NOT EXISTS (SELECT 1 FROM pull_b b WHERE b.k IS NOT DISTINCT FROM NULL AND b.id = a.id);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("50")],
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
