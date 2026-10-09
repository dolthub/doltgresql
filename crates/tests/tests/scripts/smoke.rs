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
fn test_create_table_as() {
    run_scripts(&[
        ScriptTest {
            name: "create table as select",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, k INT)",
                "INSERT INTO t VALUES (1, 10), (2, 99)",
                "CREATE TABLE u (k INT PRIMARY KEY)",
                "INSERT INTO u VALUES (10)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE c AS SELECT id, k FROM t WHERE NOT EXISTS(SELECT COUNT(*) FROM u WHERE u.k = t.k)",
                    expected: Expected::Tag("SELECT 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, k FROM c ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("k", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE rolled_back AS SELECT 1 AS v",
                    expected: Expected::Tag("SELECT 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regclass('rolled_back')",
                    expected: Expected::Rows {
                        columns: &[Column("to_regclass", REGCLASS)],
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
fn test_empty_query() {
    run_scripts(&[
        ScriptTest {
            name: "Empty query test",
            assertions: &[
                ScriptTestAssertion {
                    query: ";",
                    expected: Expected::Tag(""),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: " ",
                    expected: Expected::Tag(""),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_smoke_tests() {
    run_scripts(&[
        ScriptTest {
            name: "Simple statements",
            set_up_script: &[
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 1), (2, 2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test2 VALUES (3, 3), (4, 4);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT test2.pk FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8)],
                        rows: &[
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY 1 LIMIT 1 OFFSET 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL = NULL",
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
                    query: ";",
                    expected: Expected::Tag(""),
                    ..A
                },
                ScriptTestAssertion {
                    query: " ; ",
                    expected: Expected::Tag(""),
                    ..A
                },
                ScriptTestAssertion {
                    query: "-- this is only a comment",
                    expected: Expected::Tag(""),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Insert statements",
            set_up_script: &[
                "CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 2, 3);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test (v1, pk) VALUES (5, 4);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test (pk, v2) SELECT pk + 5, v2 + 10 FROM test WHERE v2 IS NOT NULL;",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT4), Column("v2", INT2)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("4"), T("5"), Null],
                            &[T("6"), Null, T("13")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Update statements",
            set_up_script: &[
                "CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
                "INSERT INTO test VALUES (1, 2, 3), (4, 5, 6), (7, 8, 9);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test SET v2 = 10;",
                    expected: Expected::Tag("UPDATE 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET v1 = pk + v2;",
                    expected: Expected::Tag("UPDATE 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT4), Column("v2", INT2)],
                        rows: &[
                            &[T("1"), T("11"), T("10")],
                            &[T("4"), T("14"), T("10")],
                            &[T("7"), T("17"), T("10")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET pk = subquery.val FROM (SELECT 22 as val) AS subquery WHERE pk >= 7;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT4), Column("v2", INT2)],
                        rows: &[
                            &[T("1"), T("11"), T("10")],
                            &[T("4"), T("14"), T("10")],
                            &[T("22"), T("17"), T("10")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Delete statements",
            set_up_script: &[
                "CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
                "INSERT INTO test VALUES (1, 1, 1), (2, 3, 4), (5, 7, 9);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE v2 = 9;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE v1 = pk;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT4), Column("v2", INT2)],
                        rows: &[
                            &[T("2"), T("3"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "USE statements",
            set_up_script: &[
                "CREATE DATABASE test",
                "USE test",
                "CREATE TABLE t1 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO t1 VALUES (1, 1), (2, 2);",
                "select dolt_commit('-Am', 'initial commit');",
                "select dolt_branch('b1');",
                "select dolt_checkout('b1');",
                "INSERT INTO t1 VALUES (3, 3), (4, 4);",
                "select dolt_commit('-Am', 'commit b1');",
                "select dolt_tag('tag1')",
                "INSERT INTO t1 VALUES (5, 5), (6, 6);",
                "select dolt_checkout('main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from t1 order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE test/b1",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from t1 order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                            &[T("5"), T("5")],
                            &[T("6"), T("6")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"USE "test/main""#,
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from t1 order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE 'test/tag1'",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from t1 order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Boolean results",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1 IN (2);",
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
                    query: "SELECT 2 IN (2);",
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
            name: "Commit and diff across branches",
            set_up_script: &[
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO test VALUES (1, 1), (2, 2);",
                "SELECT DOLT_ADD('-A');",
                "SELECT DOLT_COMMIT('-m', 'initial commit');",
                "SELECT DOLT_BRANCH('other');",
                "UPDATE test SET v1 = 3;",
                "SELECT DOLT_ADD('-A');",
                "SELECT DOLT_COMMIT('-m', 'commit main');",
                "SELECT DOLT_CHECKOUT('other');",
                "UPDATE test SET v1 = 4 WHERE pk = 2;",
                "SELECT DOLT_ADD('-A');",
                "SELECT DOLT_COMMIT('-m', 'commit other');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('other');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'other'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_pk, to_pk, from_v1, to_v1 FROM dolt_diff_test;",
                    expected: Expected::Rows {
                        columns: &[Column("from_pk", INT8), Column("to_pk", INT8), Column("from_v1", INT8), Column("to_v1", INT8)],
                        rows: &[
                            &[T("2"), T("2"), T("2"), T("4")],
                            &[Null, T("1"), Null, T("1")],
                            &[Null, T("2"), Null, T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ARRAY expression",
            set_up_script: &[
                "CREATE TABLE test1 (id INTEGER primary key, v1 BOOLEAN);",
                "INSERT INTO test1 VALUES (1, 'true'), (2, 'false');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[v1]::boolean[] FROM test1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("array", BOOL_ARRAY)],
                        rows: &[
                            &[T("{t}")],
                            &[T("{f}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[v1] FROM test1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("array", BOOL_ARRAY)],
                        rows: &[
                            &[T("{t}")],
                            &[T("{f}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[v1, true, v1] FROM test1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("array", BOOL_ARRAY)],
                        rows: &[
                            &[T("{t,t,t}")],
                            &[T("{f,t,f}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1::float8, 2::numeric];",
                    expected: Expected::Rows {
                        columns: &[Column("array", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("{1,2}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1::float8, NULL];",
                    expected: Expected::Rows {
                        columns: &[Column("array", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("{1,NULL}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1::int2, 2::int4, 3::int8]::varchar[];",
                    expected: Expected::Rows {
                        columns: &[Column("array", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1::int8]::int;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type bigint[] to integer", position: 22, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1::int8, 2::varchar];",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "ARRAY types bigint and character varying cannot be matched", position: 23, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Array casting",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '{true,false,true}'::boolean[];",
                    expected: Expected::Rows {
                        columns: &[Column("bool", BOOL_ARRAY)],
                        rows: &[
                            &[T("{t,f,t}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"\x68656c6c6f", "\x776f726c64", "\x6578616d706c65"}'::bytea[]::text[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("text", TEXT_ARRAY)],
                        rows: &[
                            &[T(r#"{"\\x7836383635366336633666","\\x7837373666373236633634","\\x783635373836313664373036633635"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"\\x68656c6c6f", "\\x776f726c64", "\\x6578616d706c65"}'::bytea[]::text[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("text", TEXT_ARRAY)],
                        rows: &[
                            &[T(r#"{"\\x68656c6c6f","\\x776f726c64","\\x6578616d706c65"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"abcd", "efgh", "ijkl"}'::char(3)[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("bpchar", BPCHAR_ARRAY)],
                        rows: &[
                            &[T("{abc,efg,ijk}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"2020-02-03", "2020-04-05", "2020-06-06"}'::date[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE_ARRAY)],
                        rows: &[
                            &[T("{2020-02-03,2020-04-05,2020-06-06}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1.25,2.5,3.75}'::float4[];",
                    expected: Expected::Rows {
                        columns: &[Column("float4", FLOAT4_ARRAY)],
                        rows: &[
                            &[T("{1.25,2.5,3.75}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{4.25,5.5,6.75}'::float8[];",
                    expected: Expected::Rows {
                        columns: &[Column("float8", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("{4.25,5.5,6.75}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::int2[];",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2_ARRAY)],
                        rows: &[
                            &[T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{4,5,6}'::int4[];",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4_ARRAY)],
                        rows: &[
                            &[T("{4,5,6}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{7,8,9}'::int8[];",
                    expected: Expected::Rows {
                        columns: &[Column("int8", INT8_ARRAY)],
                        rows: &[
                            &[T("{7,8,9}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"{\"a\":\"val1\"}", "{\"b\":\"value2\"}", "{\"c\": \"object_value3\"}"}'::json[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("json", JSON_ARRAY)],
                        rows: &[
                            &[T(r#"{"{\"a\":\"val1\"}","{\"b\":\"value2\"}","{\"c\": \"object_value3\"}"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"{\"d\":\"val1\"}", "{\"e\":\"value2\"}", "{\"f\": \"object_value3\"}"}'::jsonb[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB_ARRAY)],
                        rows: &[
                            &[T(r#"{"{\"d\": \"val1\"}","{\"e\": \"value2\"}","{\"f\": \"object_value3\"}"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"the", "legendary", "formula"}'::name[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME_ARRAY)],
                        rows: &[
                            &[T("{the,legendary,formula}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{10.01,20.02,30.03}'::numeric[];",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC_ARRAY)],
                        rows: &[
                            &[T("{10.01,20.02,30.03}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,10,100}'::oid[];",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID_ARRAY)],
                        rows: &[
                            &[T("{1,10,100}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"this", "is", "some", "text"}'::text[], '{text,without,quotes}'::text[], '{null,NULL,"NULL","quoted"}'::text[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("text", TEXT_ARRAY), Column("text", TEXT_ARRAY), Column("text", TEXT_ARRAY)],
                        rows: &[
                            &[T("{this,is,some,text}"), T("{text,without,quotes}"), T(r#"{NULL,NULL,"NULL",quoted}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"12:12:13", "14:14:15", "16:16:17"}'::time[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("time", TIME_ARRAY)],
                        rows: &[
                            &[T("{12:12:13,14:14:15,16:16:17}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"2020-02-03 12:13:14", "2020-04-05 15:16:17", "2020-06-06 18:19:20"}'::timestamp[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP_ARRAY)],
                        rows: &[
                            &[T(r#"{"2020-02-03 12:13:14","2020-04-05 15:16:17","2020-06-06 18:19:20"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"3920fd79-7b53-437c-b647-d450b58b4532", "a594c217-4c63-4669-96ec-40eed180b7cf", "4367b70d-8d8b-4969-a1aa-bf59536455fb"}'::uuid[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("uuid", UUID_ARRAY)],
                        rows: &[
                            &[T("{3920fd79-7b53-437c-b647-d450b58b4532,a594c217-4c63-4669-96ec-40eed180b7cf,4367b70d-8d8b-4969-a1aa-bf59536455fb}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"somewhere", "over", "the", "rainbow"}'::varchar(5)[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("varchar", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{somew,over,the,rainb}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::xid[];",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID_ARRAY)],
                        rows: &[
                            &[T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"abc""","def"}'::text[];"#,
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "{"abc""","def"}""#, detail: "Unexpected array element.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{a,b,c'::text[];",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "{a,b,c""#, detail: "Unexpected end of input.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a,b,c}'::text[];",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "a,b,c}""#, detail: r#"Array value must start with "{" or dimension information."#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a,b,c}'::text[];"#,
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "{"a,b,c}""#, detail: "Unexpected end of input.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{a",b,c}'::text[];"#,
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "{a",b,c}""#, detail: "Unexpected array element.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{a,b,"c}'::text[];"#,
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "{a,b,"c}""#, detail: "Unexpected end of input.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{a,b,c"}'::text[];"#,
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "{a,b,c"}""#, detail: "Unexpected array element.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BETWEEN",
            set_up_script: &[
                "CREATE TABLE test (v1 FLOAT8);",
                "INSERT INTO test VALUES (1), (3), (7);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 BETWEEN 1 AND 4 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 BETWEEN 2 AND 4 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 BETWEEN 4 AND 2 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 1 AND 4 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 2 AND 4 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 4 AND 2 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 NOT BETWEEN 1 AND 4 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 NOT BETWEEN 2 AND 4 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("1")],
                            &[T("7")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 NOT BETWEEN 4 AND 2 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("7")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 1 AND 4 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 2 AND 4 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("1")],
                            &[T("7")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 4 AND 2 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT8)],
                        rows: &[
                            &[T("1")],
                            &[T("7")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "IN",
            set_up_script: &[
                "CREATE TABLE test(v1 INT4, v2 INT4);",
                "INSERT INTO test VALUES (1, 1), (2, 2), (3, 3), (4, 4), (5, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 IN (2, '3', 4) ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX v2_idx ON test(v2);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v2 IN (2, '3', 4) ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SUM",
            set_up_script: &[
                "CREATE TABLE test(pk SERIAL PRIMARY KEY, v1 INT4);",
                "INSERT INTO test (v1) VALUES (1), (2), (3), (4), (5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT SUM(v1) FROM test WHERE v1 BETWEEN 3 AND 5;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT8)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX v1_idx ON test(v1);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT SUM(v1) FROM test WHERE v1 BETWEEN 3 AND 5;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT8)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ANY ROW",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL::int4) = ROW(NULL::int4);",
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
                    query: "SELECT ROW(NULL::int4) = ANY(ARRAY[ROW(NULL::int4)]);",
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
            name: "Empty statement",
            assertions: &[
                ScriptTestAssertion {
                    query: ";",
                    expected: Expected::Tag(""),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Unsupported MySQL statements",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW CREATE TABLE;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "CREATE""#, position: 6, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "querying tables with same name as pg_catalog tables",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT attname FROM pg_catalog.pg_attribute ORDER BY attname LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("attname", NAME)],
                        rows: &[
                            &[T("abbrev")],
                            &[T("abbrev")],
                            &[T("action_condition")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT attname FROM pg_attribute ORDER BY attname LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("attname", NAME)],
                        rows: &[
                            &[T("abbrev")],
                            &[T("abbrev")],
                            &[T("action_condition")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE pg_attribute (id INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into pg_attribute values (1);",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"null value in column "attname" of relation "pg_attribute" violates not-null constraint"#, detail: "Failing row contains (1, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null).", schema: "pg_catalog", table: "pg_attribute", column: "attname", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into public.pg_attribute values (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT attname FROM pg_attribute ORDER BY attname LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("attname", NAME)],
                        rows: &[
                            &[T("abbrev")],
                            &[T("abbrev")],
                            &[T("action_condition")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.pg_attribute;",
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
                    query: "drop table pg_attribute;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: r#"permission denied: "pg_attribute" is a system catalog"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "drop table public.pg_attribute;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.pg_attribute;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "public.pg_attribute" does not exist"#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "200 Row Test",
            set_up_script: &[
                "CREATE TABLE test (pk INT8 PRIMARY KEY);",
                "INSERT INTO test VALUES (1),   (2),   (3),   (4),   (5),   (6),   (7),   (8),   (9),   (10),(11),  (12),  (13),  (14),  (15),  (16),  (17),  (18),  (19),  (20),(21),  (22),  (23),  (24),  (25),  (26),  (27),  (28),  (29),  (30),(31),  (32),  (33),  (34),  (35),  (36),  (37),  (38),  (39),  (40),(41),  (42),  (43),  (44),  (45),  (46),  (47),  (48),  (49),  (50),(51),  (52),  (53),  (54),  (55),  (56),  (57),  (58),  (59),  (60),(61),  (62),  (63),  (64),  (65),  (66),  (67),  (68),  (69),  (70),(71),  (72),  (73),  (74),  (75),  (76),  (77),  (78),  (79),  (80),(81),  (82),  (83),  (84),  (85),  (86),  (87),  (88),  (89),  (90),(91),  (92),  (93),  (94),  (95),  (96),  (97),  (98),  (99),  (100),(101), (102), (103), (104), (105), (106), (107), (108), (109), (110),(111), (112), (113), (114), (115), (116), (117), (118), (119), (120),(121), (122), (123), (124), (125), (126), (127), (128), (129), (130),(131), (132), (133), (134), (135), (136), (137), (138), (139), (140),(141), (142), (143), (144), (145), (146), (147), (148), (149), (150),(151), (152), (153), (154), (155), (156), (157), (158), (159), (160),(161), (162), (163), (164), (165), (166), (167), (168), (169), (170),(171), (172), (173), (174), (175), (176), (177), (178), (179), (180),(181), (182), (183), (184), (185), (186), (187), (188), (189), (190),(191), (192), (193), (194), (195), (196), (197), (198), (199), (200);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8)],
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
                            &[T("51")],
                            &[T("52")],
                            &[T("53")],
                            &[T("54")],
                            &[T("55")],
                            &[T("56")],
                            &[T("57")],
                            &[T("58")],
                            &[T("59")],
                            &[T("60")],
                            &[T("61")],
                            &[T("62")],
                            &[T("63")],
                            &[T("64")],
                            &[T("65")],
                            &[T("66")],
                            &[T("67")],
                            &[T("68")],
                            &[T("69")],
                            &[T("70")],
                            &[T("71")],
                            &[T("72")],
                            &[T("73")],
                            &[T("74")],
                            &[T("75")],
                            &[T("76")],
                            &[T("77")],
                            &[T("78")],
                            &[T("79")],
                            &[T("80")],
                            &[T("81")],
                            &[T("82")],
                            &[T("83")],
                            &[T("84")],
                            &[T("85")],
                            &[T("86")],
                            &[T("87")],
                            &[T("88")],
                            &[T("89")],
                            &[T("90")],
                            &[T("91")],
                            &[T("92")],
                            &[T("93")],
                            &[T("94")],
                            &[T("95")],
                            &[T("96")],
                            &[T("97")],
                            &[T("98")],
                            &[T("99")],
                            &[T("100")],
                            &[T("101")],
                            &[T("102")],
                            &[T("103")],
                            &[T("104")],
                            &[T("105")],
                            &[T("106")],
                            &[T("107")],
                            &[T("108")],
                            &[T("109")],
                            &[T("110")],
                            &[T("111")],
                            &[T("112")],
                            &[T("113")],
                            &[T("114")],
                            &[T("115")],
                            &[T("116")],
                            &[T("117")],
                            &[T("118")],
                            &[T("119")],
                            &[T("120")],
                            &[T("121")],
                            &[T("122")],
                            &[T("123")],
                            &[T("124")],
                            &[T("125")],
                            &[T("126")],
                            &[T("127")],
                            &[T("128")],
                            &[T("129")],
                            &[T("130")],
                            &[T("131")],
                            &[T("132")],
                            &[T("133")],
                            &[T("134")],
                            &[T("135")],
                            &[T("136")],
                            &[T("137")],
                            &[T("138")],
                            &[T("139")],
                            &[T("140")],
                            &[T("141")],
                            &[T("142")],
                            &[T("143")],
                            &[T("144")],
                            &[T("145")],
                            &[T("146")],
                            &[T("147")],
                            &[T("148")],
                            &[T("149")],
                            &[T("150")],
                            &[T("151")],
                            &[T("152")],
                            &[T("153")],
                            &[T("154")],
                            &[T("155")],
                            &[T("156")],
                            &[T("157")],
                            &[T("158")],
                            &[T("159")],
                            &[T("160")],
                            &[T("161")],
                            &[T("162")],
                            &[T("163")],
                            &[T("164")],
                            &[T("165")],
                            &[T("166")],
                            &[T("167")],
                            &[T("168")],
                            &[T("169")],
                            &[T("170")],
                            &[T("171")],
                            &[T("172")],
                            &[T("173")],
                            &[T("174")],
                            &[T("175")],
                            &[T("176")],
                            &[T("177")],
                            &[T("178")],
                            &[T("179")],
                            &[T("180")],
                            &[T("181")],
                            &[T("182")],
                            &[T("183")],
                            &[T("184")],
                            &[T("185")],
                            &[T("186")],
                            &[T("187")],
                            &[T("188")],
                            &[T("189")],
                            &[T("190")],
                            &[T("191")],
                            &[T("192")],
                            &[T("193")],
                            &[T("194")],
                            &[T("195")],
                            &[T("196")],
                            &[T("197")],
                            &[T("198")],
                            &[T("199")],
                            &[T("200")],
                        ],
                        tag: "SELECT 200",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "INDEX as column name",
            set_up_script: &[
                "CREATE TABLE test1 (index INT4, CONSTRAINT index_constraint1 CHECK ((index >= 0)));",
                r#"CREATE TABLE test2 ("IndeX" INT4, CONSTRAINT index_constraint2 CHECK (("IndeX" >= 0)));"#,
                "INSERT INTO test1 VALUES (1);",
                "INSERT INTO test2 VALUES (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test1;",
                    expected: Expected::Rows {
                        columns: &[Column("index", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("IndeX", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test1 VALUES (-1);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "test1" violates check constraint "index_constraint1""#, detail: "Failing row contains (-1).", schema: "public", table: "test1", constraint: "index_constraint1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test2 VALUES (-1);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "test2" violates check constraint "index_constraint2""#, detail: "Failing row contains (-1).", schema: "public", table: "test2", constraint: "index_constraint2", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
