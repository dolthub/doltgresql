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
use harness::script::Cell::{Any, Null, Oid, Text as T};
use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};
use harness::wire::{Datum, F, Field, Fields, PGX_STARTUP, Receive, Send, Step, W, WireTest, run_wire_tests};

#[test]
fn test_composite_types() {
    run_scripts(&[
        ScriptTest {
            name: "composite type as subquery alias",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT 'session_stats' AS chart_name,
       						pg_catalog.Row_to_json(t) AS chart_data FROM (
							SELECT
								 (
									SELECT Count(*)
									FROM   pg_catalog.pg_stat_activity) AS "Total",
								 (
									SELECT Count(*)
									FROM   pg_catalog.pg_stat_activity
									WHERE  state = 'active') AS "Active",
								 (
									SELECT Count(*)
									FROM   pg_catalog.pg_stat_activity
                            		WHERE  state = 'idle') AS "Idle" ) t;"#,
                    expected: Expected::Rows {
                        columns: &[Column("chart_name", TEXT), Column("chart_data", JSON)],
                        rows: &[
                            &[T("session_stats"), T(r#"{"Total":6,"Active":1,"Idle":0}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("Postgres lists its background processes in pg_stat_activity, which Doltgres does not run"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_enum_types() {
    run_scripts(&[
        ScriptTest {
            name: "create enum type",
            set_up_script: &[
                "CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE person (name text, current_mood mood);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO person VALUES ('Moe', 'happy'), ('Larry', 'sad'), ('Curly', 'ok');",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'happy'::mood;",
                    expected: Expected::Rows {
                        columns: &[Column("mood", USER_DEFINED)],
                        rows: &[
                            &[T("happy")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_mood::mood from person where name = 'Moe';",
                    expected: Expected::Rows {
                        columns: &[Column("current_mood", USER_DEFINED)],
                        rows: &[
                            &[T("happy")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM person order by current_mood;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("current_mood", USER_DEFINED)],
                        rows: &[
                            &[T("Larry"), T("sad")],
                            &[T("Curly"), T("ok")],
                            &[T("Moe"), T("happy")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM person order by name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("current_mood", USER_DEFINED)],
                        rows: &[
                            &[T("Curly"), T("ok")],
                            &[T("Larry"), T("sad")],
                            &[T("Moe"), T("happy")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM person;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("current_mood", USER_DEFINED)],
                        rows: &[
                            &[T("Moe"), T("happy")],
                            &[T("Larry"), T("sad")],
                            &[T("Curly"), T("ok")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM person WHERE current_mood = 'happy';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("current_mood", USER_DEFINED)],
                        rows: &[
                            &[T("Moe"), T("happy")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM person WHERE current_mood > 'sad';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("current_mood", USER_DEFINED)],
                        rows: &[
                            &[T("Moe"), T("happy")],
                            &[T("Curly"), T("ok")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM person WHERE current_mood > 'sad' ORDER BY current_mood;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("current_mood", USER_DEFINED)],
                        rows: &[
                            &[T("Curly"), T("ok")],
                            &[T("Moe"), T("happy")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO person VALUES ('Joey', 'invalid');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input value for enum mood: "invalid""#, position: 36, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE failure AS ENUM ('ok','ok');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "pg_enum_typid_label_index""#, detail: "Key (enumtypid, enumlabel)=(16397, ok) already exists.", schema: "pg_catalog", table: "pg_enum", constraint: "pg_enum_typid_label_index", ..E }),
                    skip: Some("the detail names the OID Postgres assigned to the type, while Doltgres derives its OIDs from names"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE empty_mood AS ENUM ();",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop enum type",
            set_up_script: &[
                "CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy')",
                "CREATE TYPE empty_enum AS ENUM ()",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TYPE mood, empty_enum;",
                    expected: Expected::Tag("DROP TYPE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE empty_enum;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "empty_enum" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE empty_enum;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "empty_enum" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE IF EXISTS empty_enum;",
                    expected: Expected::Tag("DROP TYPE"),
                    notices: &[Diagnostic { code: "00000", message: r#"type "empty_enum" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE _mood;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "_mood" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "enum type cast",
            set_up_script: &[
                "CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select 'sad'::mood",
                    expected: Expected::Rows {
                        columns: &[Column("mood", USER_DEFINED)],
                        rows: &[
                            &[T("sad")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 'invalid'::mood",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input value for enum mood: "invalid""#, position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "enum type function",
            set_up_script: &[
                "CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select enum_in('sad'::cstring, 16675);",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "cache lookup failed for type 16675", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "btree index scan on enum column returns correct rows",
            set_up_script: &[
                "CREATE TYPE rainbow AS ENUM ('red', 'orange', 'yellow', 'green', 'blue', 'purple')",
                "CREATE TABLE enumtest (col rainbow)",
                "INSERT INTO enumtest VALUES ('red'), ('orange'), ('yellow'), ('green')",
                "CREATE UNIQUE INDEX enumtest_btree ON enumtest USING btree (col)",
                "SET enable_seqscan = off",
                "SET enable_bitmapscan = off",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM enumtest WHERE col = 'orange'",
                    expected: Expected::Rows {
                        columns: &[Column("col", USER_DEFINED)],
                        rows: &[
                            &[T("orange")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM enumtest WHERE col > 'orange' ORDER BY col",
                    expected: Expected::Rows {
                        columns: &[Column("col", USER_DEFINED)],
                        rows: &[
                            &[T("yellow")],
                            &[T("green")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM enumtest WHERE col < 'orange' ORDER BY col",
                    expected: Expected::Rows {
                        columns: &[Column("col", USER_DEFINED)],
                        rows: &[
                            &[T("red")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "enum array type column",
            set_up_script: &[
                "CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy');",
                "CREATE TABLE t (pk int primary key, v mood[]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, array['sad', 'happy']::mood[]);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (2, '{ok,sad}');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("{sad,happy}")],
                            &[T("2"), T("{ok,sad}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create type with existing array type name updates the name of the array type",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TYPE my_type AS ENUM ();",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE _my_type;",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT typname from pg_type where typname like '%my_type'",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME)],
                        rows: &[
                            &[T("my_type")],
                            &[T("__my_type")],
                            &[T("_my_type")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE my_type;",
                    expected: Expected::Tag("DROP TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE _my_type;",
                    expected: Expected::Tag("DROP TYPE"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_same_types() {
    run_scripts(&[
        ScriptTest {
            name: "Integer types",
            set_up_script: &[
                "CREATE TABLE test1 (v1 SMALLINT, v2 INTEGER, v3 BIGINT);",
                "CREATE TABLE test2 (v1 INT2, v2 INT4, v3 INT8);",
                "INSERT INTO test1 VALUES (1, 2, 3), (4, 5, 6);",
                "INSERT INTO test2 VALUES (1, 2, 3), (4, 5, 6);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test1 ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT2), Column("v2", INT4), Column("v3", INT8)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("4"), T("5"), T("6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test1 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT2), Column("v2", INT4), Column("v3", INT8)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("4"), T("5"), T("6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT2), Column("v2", INT4), Column("v3", INT8)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("4"), T("5"), T("6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT2), Column("v2", INT4), Column("v3", INT8)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("4"), T("5"), T("6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select int2 '2', int4 '3', int8 '4'",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2), Column("int4", INT4), Column("int8", INT8)],
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
            name: "Arbitrary precision types",
            set_up_script: &[
                "CREATE TABLE test (v1 DECIMAL(10, 1), v2 NUMERIC(11, 2));",
                "INSERT INTO test VALUES (14854.5, 2504.25), (566821525.5, 735134574.75), (21525, 134574.7);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", NUMERIC), Column("v2", NUMERIC)],
                        rows: &[
                            &[T("14854.5"), T("2504.25")],
                            &[T("21525.0"), T("134574.70")],
                            &[T("566821525.5"), T("735134574.75")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Floating point types",
            set_up_script: &[
                "CREATE TABLE test1 (v1 REAL, v2 DOUBLE PRECISION);",
                "CREATE TABLE test2 (v1 FLOAT4, v2 FLOAT8);",
                "INSERT INTO test1 VALUES (10.125, 20.4), (40.875, 81.6);",
                "INSERT INTO test2 VALUES (10.125, 20.4), (40.875, 81.6);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test1 ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT4), Column("v2", FLOAT8)],
                        rows: &[
                            &[T("10.125"), T("20.4")],
                            &[T("40.875"), T("81.6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2 ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", FLOAT4), Column("v2", FLOAT8)],
                        rows: &[
                            &[T("10.125"), T("20.4")],
                            &[T("40.875"), T("81.6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Date and time types",
            set_up_script: &[
                "CREATE TABLE test (v1 TIMESTAMP, v2 DATE);",
                "INSERT INTO test VALUES ('1986-08-02 17:04:22', '2023-09-03');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", TIMESTAMP), Column("v2", DATE)],
                        rows: &[
                            &[T("1986-08-02 17:04:22"), T("2023-09-03")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Text types",
            set_up_script: &[
                "CREATE TABLE test (v1 CHARACTER VARYING(255), v2 CHARACTER(3), v3 TEXT);",
                "INSERT INTO test VALUES ('abc', 'def', 'ghi'), ('jkl', 'mno', 'pqr');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", VARCHAR), Column("v2", BPCHAR), Column("v3", TEXT)],
                        rows: &[
                            &[T("abc"), T("def"), T("ghi")],
                            &[T("jkl"), T("mno"), T("pqr")],
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
fn test_shell_types() {
    run_scripts(&[
        ScriptTest {
            name: "shell type use cases",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TYPE undefined_type;",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1::undefined_type;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "undefined_type" is only a shell"#, position: 11, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE undefined_type;",
                    expected: Expected::Tag("DROP TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE IF EXISTS undefined_type;",
                    expected: Expected::Tag("DROP TYPE"),
                    notices: &[Diagnostic { code: "00000", message: r#"type "undefined_type" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_types() {
    run_scripts(&[
        ScriptTest {
            name: "Bigint type",
            set_up_script: &[
                "CREATE TABLE t_bigint (id INTEGER primary key, v1 BIGINT);",
                "INSERT INTO t_bigint VALUES (1, 123456789012345), (2, 987654321098765);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bigint ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("123456789012345")],
                            &[T("2"), T("987654321098765")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1::pg_catalog.int8;",
                    expected: Expected::Rows {
                        columns: &[Column("int8", INT8)],
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
            name: "Bigint key",
            set_up_script: &[
                "CREATE TABLE t_bigint (id BIGINT primary key, v1 BIGINT);",
                "INSERT INTO t_bigint VALUES (1, 123456789012345), (2, 987654321098765);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bigint WHERE id = 1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("123456789012345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bigint array type",
            set_up_script: &[
                "CREATE TABLE t_bigint (id INTEGER primary key, v1 BIGINT[]);",
                "INSERT INTO t_bigint VALUES (1, ARRAY[123456789012345, NULL]), (2, ARRAY[987654321098765, 5]), (3, ARRAY[4, 5]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bigint ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT8_ARRAY)],
                        rows: &[
                            &[T("1"), T("{123456789012345,NULL}")],
                            &[T("2"), T("{987654321098765,5}")],
                            &[T("3"), T("{4,5}")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bit type",
            set_up_script: &[
                "CREATE TABLE t_bit (id INTEGER primary key, v1 BIT(8), v2 BIT(3));",
                "INSERT INTO t_bit VALUES (1, B'11011010', '101'), (2, B'00101011', '000');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bit ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BIT), Column("v2", BIT)],
                        rows: &[
                            &[T("1"), T("11011010"), T("101")],
                            &[T("2"), T("00101011"), T("000")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 0::bit, 1::bit, 2::bit, 3::bit, 4::bit, 5::bit(2), 6::bit(2);",
                    expected: Expected::Rows {
                        columns: &[Column("bit", BIT), Column("bit", BIT), Column("bit", BIT), Column("bit", BIT), Column("bit", BIT), Column("bit", BIT), Column("bit", BIT)],
                        rows: &[
                            &[T("0"), T("1"), T("0"), T("1"), T("0"), T("01"), T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (-1)::bit, (-2)::bit, (-5)::bit(2), (-6::int4)::bit(2);",
                    expected: Expected::Rows {
                        columns: &[Column("bit", BIT), Column("bit", BIT), Column("bit", BIT), Column("bit", BIT)],
                        rows: &[
                            &[T("1"), T("0"), T("11"), T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_bit VALUES (3, B'101', '111');",
                    expected: Expected::Error(Diagnostic { code: "22026", message: "bit string length 3 does not match type bit(8)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_bit VALUES (3, B'1001000110', '111');",
                    expected: Expected::Error(Diagnostic { code: "22026", message: "bit string length 10 does not match type bit(8)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_bit VALUES (3, B'10010001', '11100100');",
                    expected: Expected::Error(Diagnostic { code: "22026", message: "bit string length 8 does not match type bit(3)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_bit VALUES (3, B'10012345', '111');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#""2" is not a valid binary digit"#, position: 30, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_bit VALUES (3, '10012345', '111');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#""2" is not a valid binary digit"#, position: 30, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bit key",
            set_up_script: &[
                "CREATE TABLE t_bit (id BIT(8) primary key, v1 BIT(8));",
                "INSERT INTO t_bit VALUES (B'11011010', B'11011010'), (B'00101011', B'00101011');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bit WHERE id = B'11011010' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", BIT), Column("v1", BIT)],
                        rows: &[
                            &[T("11011010"), T("11011010")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Boolean type",
            set_up_script: &[
                "CREATE TABLE t_boolean (id INTEGER primary key, v1 BOOLEAN);",
                "INSERT INTO t_boolean VALUES (1, true), (2, 'false'), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_boolean ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BOOL)],
                        rows: &[
                            &[T("1"), T("t")],
                            &[T("2"), T("f")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_boolean ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BOOL)],
                        rows: &[
                            &[T("2"), T("f")],
                            &[T("1"), T("t")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_boolean WHERE v1 IS NOT NULL ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BOOL)],
                        rows: &[
                            &[T("1"), T("t")],
                            &[T("2"), T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_boolean WHERE v1 IS NOT NULL ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BOOL)],
                        rows: &[
                            &[T("2"), T("f")],
                            &[T("1"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Boolean key",
            set_up_script: &[
                "CREATE TABLE t_boolean (id boolean primary key, v1 BOOLEAN);",
                "INSERT INTO t_boolean VALUES (true, true), (false, 'false')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_boolean where id ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", BOOL), Column("v1", BOOL)],
                        rows: &[
                            &[T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "boolean indexes",
            set_up_script: &[
                "create table t (b bool);",
                "insert into t values (false);",
                "create table t_idx (b bool);",
                "create index idx on t_idx(b);",
                "insert into t_idx values (false);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from t where (b in (false));",
                    expected: Expected::Rows {
                        columns: &[Column("b", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t_idx where (b in (false));",
                    expected: Expected::Rows {
                        columns: &[Column("b", BOOL)],
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
            name: "Boolean array type",
            set_up_script: &[
                "CREATE TABLE t_boolean_array (id INTEGER primary key, v1 BOOLEAN[]);",
                "INSERT INTO t_boolean_array VALUES (1, ARRAY[true, false]), (2, ARRAY[false, true]), (3, ARRAY[true, true]), (4, ARRAY[false, false]), (5, ARRAY[true]), (6, ARRAY[false]), (7, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_boolean_array ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BOOL_ARRAY)],
                        rows: &[
                            &[T("1"), T("{t,f}")],
                            &[T("2"), T("{f,t}")],
                            &[T("3"), T("{t,t}")],
                            &[T("4"), T("{f,f}")],
                            &[T("5"), T("{t}")],
                            &[T("6"), T("{f}")],
                            &[T("7"), Null],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_boolean_array ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BOOL_ARRAY)],
                        rows: &[
                            &[T("6"), T("{f}")],
                            &[T("4"), T("{f,f}")],
                            &[T("2"), T("{f,t}")],
                            &[T("5"), T("{t}")],
                            &[T("1"), T("{t,f}")],
                            &[T("3"), T("{t,t}")],
                            &[T("7"), Null],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_boolean_array WHERE v1 IS NOT NULL ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BOOL_ARRAY)],
                        rows: &[
                            &[T("1"), T("{t,f}")],
                            &[T("2"), T("{f,t}")],
                            &[T("3"), T("{t,t}")],
                            &[T("4"), T("{f,f}")],
                            &[T("5"), T("{t}")],
                            &[T("6"), T("{f}")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_boolean_array WHERE v1 IS NOT NULL ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BOOL_ARRAY)],
                        rows: &[
                            &[T("6"), T("{f}")],
                            &[T("4"), T("{f,f}")],
                            &[T("2"), T("{f,t}")],
                            &[T("5"), T("{t}")],
                            &[T("1"), T("{t,f}")],
                            &[T("3"), T("{t,t}")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bigserial type",
            set_up_script: &[
                "CREATE TABLE t_bigserial (id INTEGER primary key, v1 BIGSERIAL);",
                "INSERT INTO t_bigserial VALUES (1, 123456789012345), (2, 987654321098765);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bigserial ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("123456789012345")],
                            &[T("2"), T("987654321098765")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bigserial key",
            set_up_script: &[
                "CREATE TABLE t_bigserial (id BIGSERIAL primary key, v1 BIGSERIAL);",
                "INSERT INTO t_bigserial VALUES (123456789012345, 123456789012345), (987654321098765, 987654321098765);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bigserial where ID = 987654321098765 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("987654321098765"), T("987654321098765")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bit varying type",
            set_up_script: &[
                "CREATE TABLE t_bit_varying (id INTEGER primary key, v1 BIT VARYING(16));",
                "INSERT INTO t_bit_varying VALUES (1, B'1101101010101010'), (2, B'0010101101010101');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bit_varying ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", VARBIT)],
                        rows: &[
                            &[T("1"), T("1101101010101010")],
                            &[T("2"), T("0010101101010101")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_bit_varying VALUES (3, B'101010101010101010');",
                    expected: Expected::Error(Diagnostic { code: "22001", message: "bit string too long for type bit varying(16)", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bit varying type, unbounded",
            set_up_script: &[
                "CREATE TABLE t_bit_varying (id INTEGER primary key, v1 BIT VARYING);",
                "INSERT INTO t_bit_varying VALUES (1, B'1101101010101010'), (2, B'0010101101010101');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bit_varying ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", VARBIT)],
                        rows: &[
                            &[T("1"), T("1101101010101010")],
                            &[T("2"), T("0010101101010101")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_bit_varying VALUES (3, B'101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bit_varying WHERE id = 3 order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", VARBIT)],
                        rows: &[
                            &[T("3"), T("101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Box type",
            set_up_script: &[
                "CREATE TABLE t_box (id INTEGER primary key, v1 BOX);",
                "INSERT INTO t_box VALUES (1, '(1,2),(3,4)'), (2, '(5,6),(7,8)');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_box ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BOX)],
                        rows: &[
                            &[T("1"), T("(3,4),(1,2)")],
                            &[T("2"), T("(7,8),(5,6)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bytea type",
            set_up_script: &[
                "CREATE TABLE t_bytea (id INTEGER primary key, v1 BYTEA);",
                r#"INSERT INTO t_bytea VALUES (1, E'\\xDEADBEEF'), (2, '\xC0FFEE'), (3, ''), (4, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bytea ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BYTEA)],
                        rows: &[
                            &[T("1"), T(r#"\xdeadbeef"#)],
                            &[T("2"), T(r#"\xc0ffee"#)],
                            &[T("3"), T(r#"\x"#)],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bytea key",
            set_up_script: &[
                "CREATE TABLE t_bytea (id BYTEA primary key, v1 BYTEA);",
                r#"INSERT INTO t_bytea VALUES (E'\\xCAFEBABE', E'\\xDEADBEEF'), ('\xBADD00D5', '\xC0FFEE');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT * FROM t_bytea WHERE ID = E'\\xCAFEBABE' ORDER BY id;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", BYTEA), Column("v1", BYTEA)],
                        rows: &[
                            &[T(r#"\xcafebabe"#), T(r#"\xdeadbeef"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bytea key with mixed short and long values",
            set_up_script: &[
                "CREATE TABLE t_bytea_keys (id BYTEA primary key, v1 INTEGER);",
                r#"INSERT INTO t_bytea_keys VALUES ('\x11', 1);"#,
                r#"INSERT INTO t_bytea_keys VALUES ('\x22787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878', 2);"#,
                r#"INSERT INTO t_bytea_keys VALUES ('\x33', 3);"#,
                r#"INSERT INTO t_bytea_keys VALUES ('\x44797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979', 4);"#,
                r#"INSERT INTO t_bytea_keys VALUES ('\x55', 5);"#,
                r#"INSERT INTO t_bytea_keys VALUES ('\xff', 6);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_bytea_keys ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_bytea_keys ORDER BY id DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("6")],
                            &[T("5")],
                            &[T("4")],
                            &[T("3")],
                            &[T("2")],
                            &[T("1")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT v1 FROM t_bytea_keys WHERE id = '\x11';"#,
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT v1 FROM t_bytea_keys WHERE id = '\xff';"#,
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT v1 FROM t_bytea_keys WHERE id = '\x22787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878';"#,
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT v1 FROM t_bytea_keys WHERE id = '\x44797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979';"#,
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT v1 FROM t_bytea_keys WHERE id < '\x33'::bytea ORDER BY id;"#,
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT v1 FROM t_bytea_keys WHERE id > '\x44'::bytea ORDER BY id;"#,
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT v1 FROM t_bytea_keys WHERE id > '\x22'::bytea AND id < '\x55'::bytea ORDER BY id;"#,
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select id from t_bytea_keys order by id",
                    expected: Expected::Rows {
                        columns: &[Column("id", BYTEA)],
                        rows: &[
                            &[T(r#"\x11"#)],
                            &[T(r#"\x22787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878787878"#)],
                            &[T(r#"\x33"#)],
                            &[T(r#"\x44797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979797979"#)],
                            &[T(r#"\x55"#)],
                            &[T(r#"\xff"#)],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "bpchar type",
            assertions: &[
                ScriptTestAssertion {
                    query: "create table bptest1 (pk int primary key, c1 bpchar, c2 bpchar(12));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into bptest1 values (1, '1', '1');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from bptest1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", BPCHAR), Column("c2", BPCHAR)],
                        rows: &[
                            &[T("1"), T("1"), T("1           ")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '!'::bpchar;",
                    expected: Expected::Rows {
                        columns: &[Column("bpchar", BPCHAR)],
                        rows: &[
                            &[T("!")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '!'::bpchar(1);",
                    expected: Expected::Rows {
                        columns: &[Column("bpchar", BPCHAR)],
                        rows: &[
                            &[T("!")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '!'::bpchar(2);",
                    expected: Expected::Rows {
                        columns: &[Column("bpchar", BPCHAR)],
                        rows: &[
                            &[T("! ")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Character type",
            set_up_script: &[
                "CREATE TABLE t_character (id INTEGER primary key, v1 CHARACTER(5));",
                "INSERT INTO t_character VALUES (1, 'abcde'), (2, 'vwxyz'), (3, 'ghi'), (4, ''), (5, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_character ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BPCHAR)],
                        rows: &[
                            &[T("1"), T("abcde")],
                            &[T("2"), T("vwxyz")],
                            &[T("3"), T("ghi  ")],
                            &[T("4"), T("     ")],
                            &[T("5"), Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(v1) FROM t_character ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("5")],
                            &[T("5")],
                            &[T("3")],
                            &[T("0")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT char(20) 'characters' || ' and text' AS "Concat char to unknown type";"#,
                    expected: Expected::Rows {
                        columns: &[Column("Concat char to unknown type", TEXT)],
                        rows: &[
                            &[T("characters and text")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT true::char, false::char;",
                    expected: Expected::Rows {
                        columns: &[Column("bpchar", BPCHAR), Column("bpchar", BPCHAR)],
                        rows: &[
                            &[T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT true::character(5), false::character(5);",
                    expected: Expected::Rows {
                        columns: &[Column("bpchar", BPCHAR), Column("bpchar", BPCHAR)],
                        rows: &[
                            &[T("true "), T("false")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT char 'c' = char 'c' AS true;",
                    expected: Expected::Rows {
                        columns: &[Column("true", BOOL)],
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
            name: "Character key",
            set_up_script: &[
                "CREATE TABLE t_character (id CHAR(5) primary key, v1 CHARACTER(5));",
                "INSERT INTO t_character VALUES ('abcde', 'fghjk'), ('vwxyz', '12345'), ('vwxy', '1234')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_character WHERE ID = 'vwxyz' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", BPCHAR), Column("v1", BPCHAR)],
                        rows: &[
                            &[T("vwxyz"), T("12345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(id) FROM t_character;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("5")],
                            &[T("5")],
                            &[T("4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Internal char type",
            set_up_script: &[
                r#"CREATE TABLE t_char (id INTEGER primary key, v1 "char");"#,
                "INSERT INTO t_char VALUES (1, 'abcde'), (2, 'vwxyz'), (3, '123'), (4, ''), (5, NULL), (100, 'こんにちは');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_char ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", CHAR)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("v")],
                            &[T("3"), T("1")],
                            &[T("4"), T("")],
                            &[T("5"), Null],
                            &[T("100"), T(r#"\343"#)],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_char VALUES (6, 7);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v1" is of type "char" but expression is of type integer"#, hint: "You will need to rewrite or cast the expression.", position: 31, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_char VALUES (6, true);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v1" is of type "char" but expression is of type boolean"#, hint: "You will need to rewrite or cast the expression.", position: 31, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT true::"char";"#,
                    expected: Expected::Error(Diagnostic { code: "42846", message: r#"cannot cast type boolean to "char""#, position: 12, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 100000::bigint::"char";"#,
                    expected: Expected::Error(Diagnostic { code: "42846", message: r#"cannot cast type bigint to "char""#, position: 22, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 'abc'::"char", '123'::varchar(3)::"char";"#,
                    expected: Expected::Rows {
                        columns: &[Column("char", CHAR), Column("char", CHAR)],
                        rows: &[
                            &[T("a"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 'def'::name::"char";"#,
                    expected: Expected::Rows {
                        columns: &[Column("char", CHAR)],
                        rows: &[
                            &[T("d")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v1::int, v1::text FROM t_char WHERE id < 10;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("97"), T("a")],
                            &[T("2"), T("118"), T("v")],
                            &[T("3"), T("49"), T("1")],
                            &[T("4"), T("0"), T("")],
                            &[T("5"), Null, Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::int FROM t_char WHERE id = 100;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("-29")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_char VALUES (6, '0123456789012345678901234567890123456789012345678901234567890123456789');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_char WHERE id=6;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", CHAR)],
                        rows: &[
                            &[T("6"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_char VALUES (7, 'abc'::name);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v1" is of type "char" but expression is of type name"#, hint: "You will need to rewrite or cast the expression.", position: 31, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_char VALUES (8, 'def'::text);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_char VALUES (9, 'ghi'::varchar);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_char WHERE id >= 7 AND id < 10 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", CHAR)],
                        rows: &[
                            &[T("8"), T("d")],
                            &[T("9"), T("g")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Character varying type",
            set_up_script: &[
                "CREATE TABLE t_varchar (id INTEGER primary key, v1 CHARACTER VARYING(10));",
                "INSERT INTO t_varchar VALUES (1, 'abcdefghij'), (2, 'klmnopqrst'), (3, ''), (4, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_varchar ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", VARCHAR)],
                        rows: &[
                            &[T("1"), T("abcdefghij")],
                            &[T("2"), T("klmnopqrst")],
                            &[T("3"), T("")],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT true::character varying(10), false::character varying(10);",
                    expected: Expected::Rows {
                        columns: &[Column("varchar", VARCHAR), Column("varchar", VARCHAR)],
                        rows: &[
                            &[T("true"), T("false")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Character varying type as primary key",
            set_up_script: &[
                "CREATE TABLE t_varchar (id INTEGER, v1 CHARACTER VARYING(10) primary key);",
                "INSERT INTO t_varchar VALUES (1, 'abcdefghij'), (2, 'klmnopqrst'), (3, '');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_varchar ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", VARCHAR)],
                        rows: &[
                            &[T("1"), T("abcdefghij")],
                            &[T("2"), T("klmnopqrst")],
                            &[T("3"), T("")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT true::character varying(10), false::character varying(10);",
                    expected: Expected::Rows {
                        columns: &[Column("varchar", VARCHAR), Column("varchar", VARCHAR)],
                        rows: &[
                            &[T("true"), T("false")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Character varying array type, with length",
            set_up_script: &[
                "CREATE TABLE t_varchar1 (v1 CHARACTER VARYING[]);",
                "CREATE TABLE t_varchar2 (v1 CHARACTER VARYING(1)[]);",
                r#"INSERT INTO t_varchar1 VALUES (ARRAY['ab''cdef', 'what', 'is,hi', 'wh"at']);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT v1::varchar(1)[] FROM t_varchar1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{a,w,i,w}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO t_varchar2 VALUES (ARRAY['ab''cdef', 'what', 'is,hi', 'wh"at']);"#,
                    expected: Expected::Error(Diagnostic { code: "22001", message: "value too long for type character varying(1)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_varchar2 VALUES (ARRAY['a', 'w', 'i', 'w']);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_varchar2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{a,w,i,w}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Character varying type, no length",
            set_up_script: &[
                "CREATE TABLE t_varchar (id INTEGER primary key, v1 CHARACTER VARYING);",
                "INSERT INTO t_varchar VALUES (1, 'abcdefghij'), (2, 'klmnopqrst');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_varchar ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", VARCHAR)],
                        rows: &[
                            &[T("1"), T("abcdefghij")],
                            &[T("2"), T("klmnopqrst")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Character varying type, no length, as primary key",
            set_up_script: &[
                "CREATE TABLE t_varchar (id INTEGER, v1 CHARACTER VARYING primary key);",
                "INSERT INTO t_varchar VALUES (1, 'abcdefghij'), (2, 'klmnopqrst');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_varchar ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", VARCHAR)],
                        rows: &[
                            &[T("1"), T("abcdefghij")],
                            &[T("2"), T("klmnopqrst")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Character varying key with mixed short and long values",
            set_up_script: &[
                "CREATE TABLE t_varchar_keys (id VARCHAR primary key, v1 INTEGER);",
                "INSERT INTO t_varchar_keys VALUES ('aa', 1);",
                "INSERT INTO t_varchar_keys VALUES ('bb' || repeat('x', 10500), 2);",
                "INSERT INTO t_varchar_keys VALUES ('cc', 3);",
                "INSERT INTO t_varchar_keys VALUES ('dd' || repeat('y', 10500), 4);",
                "INSERT INTO t_varchar_keys VALUES ('ee', 5);",
                "INSERT INTO t_varchar_keys VALUES ('zz', 6);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_varchar_keys ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_varchar_keys ORDER BY id DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("6")],
                            &[T("5")],
                            &[T("4")],
                            &[T("3")],
                            &[T("2")],
                            &[T("1")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_varchar_keys WHERE id = 'aa';",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_varchar_keys WHERE id = 'zz';",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_varchar_keys WHERE id = 'bb' || repeat('x', 10500);",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_varchar_keys WHERE id = 'dd' || repeat('y', 10500);",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_varchar_keys WHERE id < 'cc' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_varchar_keys WHERE id > 'dd' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_varchar_keys WHERE id > 'bb' AND id < 'ee' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1, length(id) FROM t_varchar_keys WHERE v1 IN (1, 2, 4) ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("length", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("10502")],
                            &[T("4"), T("10502")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Character varying array type, no length",
            set_up_script: &[
                "CREATE TABLE t_varchar (id INTEGER primary key, v1 CHARACTER VARYING[]);",
                r#"INSERT INTO t_varchar VALUES (1, '{abcdefghij, NULL}'), (2, ARRAY['ab''cdef', 'what', 'is,hi', 'wh"at', '}', '{', '{}']);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_varchar ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("1"), T("{abcdefghij,NULL}")],
                            &[T("2"), T(r#"{ab'cdef,what,"is,hi","wh\"at","}","{","{}"}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Array literal parsing preserves internal whitespace in unquoted elements",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '{2026-08-21 12:00:00,2026-08-22 13:30:00}'::timestamp[];",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP_ARRAY)],
                        rows: &[
                            &[T(r#"{"2026-08-21 12:00:00","2026-08-22 13:30:00"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('{2026-08-21 12:00:00+05:00}'::timestamptz[])[1] = '2026-08-21 07:00:00+00'::timestamptz;",
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
                    query: "SELECT '{1 day 2 hours, 3 days}'::interval[];",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL_ARRAY)],
                        rows: &[
                            &[T(r#"{"1 day 02:00:00","3 days"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{ 1 , 2 , 3 }'::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{ }'::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{ NULL , 2 }'::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4_ARRAY)],
                        rows: &[
                            &[T("{NULL,2}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "2D array",
            set_up_script: &[
                "CREATE TABLE t_varchar (id INTEGER primary key, v1 CHARACTER VARYING[][]);",
                r#"INSERT INTO t_varchar VALUES (1, '{{abcdefghij, NULL}, {1234, abc}}'), (2, ARRAY['ab''cdef', 'what', 'is,hi', 'wh"at', '}', '{', '{}']);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_varchar ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("1"), T("{{abcdefghij,NULL},{1234,abc}}")],
                            &[T("2"), T(r#"{ab'cdef,what,"is,hi","wh\"at","}","{","{}"}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Cidr type",
            set_up_script: &[
                "CREATE TABLE t_cidr (id INTEGER primary key, v1 CIDR);",
                "INSERT INTO t_cidr VALUES (1, '192.168.1.0/24'), (2, '10.0.0.0/8');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_cidr ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", CIDR)],
                        rows: &[
                            &[T("1"), T("192.168.1.0/24")],
                            &[T("2"), T("10.0.0.0/8")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Circle type",
            set_up_script: &[
                "CREATE TABLE t_circle (id INTEGER primary key, v1 CIRCLE);",
                "INSERT INTO t_circle VALUES (1, '<(1,2),3>'), (2, '<(4,5),6>');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_circle ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", CIRCLE)],
                        rows: &[
                            &[T("1"), T("<(1,2),3>")],
                            &[T("2"), T("<(4,5),6>")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Date type",
            set_up_script: &[
                "CREATE TABLE t_date (id INTEGER primary key, v1 DATE);",
                "INSERT INTO t_date VALUES (1, '2023-01-01'), (2, '2023-02-02');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_date ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", DATE)],
                        rows: &[
                            &[T("1"), T("2023-01-01")],
                            &[T("2"), T("2023-02-02")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '2022-2-2'",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("2022-02-02")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '2022-02-02'",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("2022-02-02")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '2024-10-31'::date;",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("2024-10-31")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '2024-OCT-31'::date;",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("2024-10-31")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '20241031'::date;",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("2024-10-31")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '2024Oct31'::date;",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("2024-10-31")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '10 31 2024'::date;",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("2024-10-31")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 'Oct 31 2024'::date;",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("2024-10-31")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date 'J2451187';",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("1999-01-08")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '08-Jan-99';",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("1999-01-08")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '2025-07-21' - 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", DATE)],
                        rows: &[
                            &[T("2025-07-20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '2025-07-21' - date '2025-07-18';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '2025-07-21' - interval '2 days';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2025-07-19 00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '1991-02-03' - time '04:05:06';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("1991-02-02 19:54:54")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '2025-07-21' - 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", DATE)],
                        rows: &[
                            &[T("2025-07-20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '1991-02-03' - time '04:05:06';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("1991-02-02 19:54:54")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '2025-07-21' + 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", DATE)],
                        rows: &[
                            &[T("2025-07-22")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '2025-07-21' + interval '2 days';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2025-07-23 00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '2025-07-21' + time '04:05:06';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2025-07-21 04:05:06")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '2025-07-21' + time '04:05:06 UTC';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2025-07-21 04:05:06")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Date key",
            set_up_script: &[
                "CREATE TABLE t_date (id DATE primary key, v1 DATE);",
                "INSERT INTO t_date VALUES ('2025-01-01', '2023-01-01'), ('2026-01-01', '2023-02-02');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_date where Id = '2025-01-01' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", DATE), Column("v1", DATE)],
                        rows: &[
                            &[T("2025-01-01"), T("2023-01-01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Double precision type",
            set_up_script: &[
                "CREATE TABLE t_double_precision (id INTEGER primary key, v1 DOUBLE PRECISION);",
                "INSERT INTO t_double_precision VALUES (1, 123.456), (2, 789.012);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_double_precision ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", FLOAT8)],
                        rows: &[
                            &[T("1"), T("123.456")],
                            &[T("2"), T("789.012")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Double precision key",
            set_up_script: &[
                "CREATE TABLE t_double_precision (id DOUBLE PRECISION primary key, v1 DOUBLE PRECISION);",
                "INSERT INTO t_double_precision VALUES (456.789, 123.456), (123.456, 789.012);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_double_precision WHERE id = 456.789 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", FLOAT8), Column("v1", FLOAT8)],
                        rows: &[
                            &[T("456.789"), T("123.456")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Double precision array type",
            set_up_script: &[
                "CREATE TABLE t_double_precision (id INTEGER primary key, v1 DOUBLE PRECISION[]);",
                "INSERT INTO t_double_precision VALUES (1, ARRAY[123.456, NULL]), (2, ARRAY[789.012, 125.125]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_double_precision ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("1"), T("{123.456,NULL}")],
                            &[T("2"), T("{789.012,125.125}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Inet type",
            set_up_script: &[
                "CREATE TABLE t_inet (id INTEGER primary key, v1 INET);",
                "INSERT INTO t_inet VALUES (1, '192.168.1.1'), (2, '10.0.0.1');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_inet ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INET)],
                        rows: &[
                            &[T("1"), T("192.168.1.1")],
                            &[T("2"), T("10.0.0.1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Integer type",
            set_up_script: &[
                "CREATE TABLE t_integer (id INTEGER primary key, v1 INTEGER);",
                "INSERT INTO t_integer VALUES (1, 123), (2, 456);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_integer ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("123")],
                            &[T("2"), T("456")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Integer array type",
            set_up_script: &[
                "CREATE TABLE t_integer (id INTEGER primary key, v1 INTEGER[]);",
                "INSERT INTO t_integer VALUES (1, ARRAY[123,NULL]), (2, ARRAY[456,823753913]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_integer ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{123,NULL}")],
                            &[T("2"), T("{456,823753913}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Interval type",
            set_up_script: &[
                "CREATE TABLE t_interval (id INTEGER primary key, v1 INTERVAL);",
                "INSERT INTO t_interval VALUES (1, '1 day 3 hours'), (2, '23 hours 30 minutes'), (3, '@ 1 minute');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_interval ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INTERVAL)],
                        rows: &[
                            &[T("1"), T("1 day 03:00:00")],
                            &[T("2"), T("23:30:00")],
                            &[T("3"), T("00:01:00")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_interval ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INTERVAL)],
                        rows: &[
                            &[T("3"), T("00:01:00")],
                            &[T("2"), T("23:30:00")],
                            &[T("1"), T("1 day 03:00:00")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v1::char, v1::name FROM t_interval;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BPCHAR), Column("v1", NAME)],
                        rows: &[
                            &[T("1"), T("1"), T("1 day 03:00:00")],
                            &[T("2"), T("2"), T("23:30:00")],
                            &[T("3"), T("0"), T("00:01:00")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2 years 15 months 100 weeks 99 hours 123456789 milliseconds'::interval;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("3 years 3 mons 700 days 133:17:36.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2 years 15 months 100 weeks 99 hours 123456789 milliseconds'::interval::char;",
                    expected: Expected::Rows {
                        columns: &[Column("bpchar", BPCHAR)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2 years 15 months 100 weeks 99 hours 123456789 milliseconds'::interval::text;",
                    expected: Expected::Rows {
                        columns: &[Column("text", TEXT)],
                        rows: &[
                            &[T("3 years 3 mons 700 days 133:17:36.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2 years 15 months 100 weeks 99 hours 123456789 milliseconds'::char::interval;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("00:00:02")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '13 months'::name::interval;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("1 year 1 mon")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '13 months'::bpchar::interval;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("1 year 1 mon")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '13 months'::varchar::interval;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("1 year 1 mon")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '13 months'::text::interval;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("1 year 1 mon")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '13 months'::char::interval;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("00:00:01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_interval VALUES (3, 7);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v1" is of type interval but expression is of type integer"#, hint: "You will need to rewrite or cast the expression.", position: 35, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_interval VALUES (3, true);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v1" is of type interval but expression is of type boolean"#, hint: "You will need to rewrite or cast the expression.", position: 35, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT CAST(interval '02:03' AS time) AS "02:03:00";"#,
                    expected: Expected::Rows {
                        columns: &[Column("02:03:00", TIME)],
                        rows: &[
                            &[T("02:03:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Interval key",
            set_up_script: &[
                "CREATE TABLE t_interval (id interval primary key, v1 INTERVAL);",
                "INSERT INTO t_interval VALUES ('1 hour', '1 day 3 hours'), ('2 days', '23 hours 30 minutes');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_interval WHERE id = '1 hour' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INTERVAL), Column("v1", INTERVAL)],
                        rows: &[
                            &[T("01:00:00"), T("1 day 03:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Interval array type",
            set_up_script: &[
                "CREATE TABLE t_interval_array (id INTEGER primary key, v1 INTERVAL[]);",
                "INSERT INTO t_interval_array VALUES (1, ARRAY['1 day 3 hours'::interval,'5 days 2 hours'::interval]), (2, ARRAY['3 years 3 mons 700 days 133:17:36.789'::interval,'200 hours'::interval]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_interval_array ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INTERVAL_ARRAY)],
                        rows: &[
                            &[T("1"), T(r#"{"1 day 03:00:00","5 days 02:00:00"}"#)],
                            &[T("2"), T(r#"{"3 years 3 mons 700 days 133:17:36.789",200:00:00}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: json has no btree operator class in Postgres, so it cannot be a primary key; jsonb can.
        ScriptTest {
            name: "JSON key",
            set_up_script: &[
                "CREATE TABLE t_json (id JSONB primary key, v1 JSON);",
                r#"INSERT INTO t_json VALUES ('{"key": "value"}', '{"key": "value"}');"#,
                "INSERT INTO t_json VALUES ('123', '123');",
                "INSERT INTO t_json VALUES ('true', 'true');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT * FROM t_json WHERE id = '{"key": "value"}' ORDER BY id;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", JSONB), Column("v1", JSON)],
                        rows: &[
                            &[T(r#"{"key": "value"}"#), T(r#"{"key": "value"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSON type",
            set_up_script: &[
                "CREATE TABLE t_json (id INTEGER primary key, v1 JSON);",
                r#"INSERT INTO t_json VALUES (1, '{"key1": {"key": "value"}}'), (2, '{"num":42}'), (3, '{"key1": "value1", "key2": "value2"}'), (4, '{"key1": {"key": [2,3]}}');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_json ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", JSON)],
                        rows: &[
                            &[T("1"), T(r#"{"key1": {"key": "value"}}"#)],
                            &[T("2"), T(r#"{"num":42}"#)],
                            &[T("3"), T(r#"{"key1": "value1", "key2": "value2"}"#)],
                            &[T("4"), T(r#"{"key1": {"key": [2,3]}}"#)],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_json ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", JSON)],
                        rows: &[
                            &[T("1"), T(r#"{"key1": {"key": "value"}}"#)],
                            &[T("2"), T(r#"{"num":42}"#)],
                            &[T("3"), T(r#"{"key1": "value1", "key2": "value2"}"#)],
                            &[T("4"), T(r#"{"key1": {"key": [2,3]}}"#)],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "Insert into t_json values (100, null) returning *",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", JSON)],
                        rows: &[
                            &[T("100"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t_json where id = 100",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", JSON)],
                        rows: &[
                            &[T("100"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "Insert into t_json values ($1, $2) returning *",
                    bind_vars: &[BindVar::Str("101"), BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", JSON)],
                        rows: &[
                            &[T("101"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '5'::json;",
                    expected: Expected::Rows {
                        columns: &[Column("json", JSON)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'false'::json;",
                    expected: Expected::Rows {
                        columns: &[Column("json", JSON)],
                        rows: &[
                            &[T("false")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"hi"'::json;"#,
                    expected: Expected::Rows {
                        columns: &[Column("json", JSON)],
                        rows: &[
                            &[T(r#""hi""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"\u0000"'::json"#,
                    expected: Expected::Rows {
                        columns: &[Column("json", JSON)],
                        rows: &[
                            &[T(r#""\u0000""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT null::json;",
                    expected: Expected::Rows {
                        columns: &[Column("json", JSON)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'null'::json;",
                    expected: Expected::Rows {
                        columns: &[Column("json", JSON)],
                        rows: &[
                            &[T("null")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"reading": 1.230e-5}'::json;"#,
                    expected: Expected::Rows {
                        columns: &[Column("json", JSON)],
                        rows: &[
                            &[T(r#"{"reading": 1.230e-5}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"select json '{ "a":  "\ud83d\ude04\ud83d\udc36" }' -> 'a'"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
                        rows: &[
                            &[T(r#""\ud83d\ude04\ud83d\udc36""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{'::json",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "invalid input syntax for type json", detail: "The input string ended unexpectedly.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"key": "value"'::json"#,
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "invalid input syntax for type json", detail: "The input string ended unexpectedly.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSON column default",
            set_up_script: &[
                r#"CREATE TABLE t_json (id INTEGER primary key, v1 JSON DEFAULT '{"num": 42}'::JSON);"#,
                r#"INSERT INTO t_json VALUES (1, '{"key1": {"key": "value"}}');"#,
                "INSERT INTO t_json (id) VALUES (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_json ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", JSON)],
                        rows: &[
                            &[T("1"), T(r#"{"key1": {"key": "value"}}"#)],
                            &[T("2"), T(r#"{"num": 42}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSONB type",
            set_up_script: &[
                "CREATE TABLE t_jsonb (id INTEGER primary key, v1 JSONB);",
                r#"INSERT INTO t_jsonb VALUES (1, '{"key": "value"}'), (2, '{"num": 42}');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_jsonb ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", JSONB)],
                        rows: &[
                            &[T("1"), T(r#"{"key": "value"}"#)],
                            &[T("2"), T(r#"{"num": 42}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t_jsonb values (3, null) returning *",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", JSONB)],
                        rows: &[
                            &[T("3"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t_jsonb values ($1, $2) returning *",
                    bind_vars: &[BindVar::Str("4"), BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", JSONB)],
                        rows: &[
                            &[T("4"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"bar": "baz", "balance": 7.77, "active":false}'::jsonb;"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[T(r#"{"bar": "baz", "active": false, "balance": 7.77}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"active": "baz", "active":false, "balance": 7.77}'::jsonb;"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[T(r#"{"active": false, "balance": 7.77}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"active":false, "balance": 7.77, "bar": "baz"}'::jsonb;"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[T(r#"{"bar": "baz", "active": false, "balance": 7.77}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb '{"a":null, "b":"qq"}' ? 'a';"#,
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
                    query: "SELECT '1.3e100'::jsonb;",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb", JSONB)],
                        rows: &[
                            &[T("13000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '12345.05'::jsonb::int2;",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2)],
                        rows: &[
                            &[T("12345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSONB column default",
            set_up_script: &[
                r#"CREATE TABLE t_json (id INTEGER primary key, v1 JSONB DEFAULT '{"num": 42}'::JSONB);"#,
                r#"INSERT INTO t_json VALUES (1, '{"key1": {"key": "value"}}');"#,
                "INSERT INTO t_json (id) VALUES (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_json ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", JSONB)],
                        rows: &[
                            &[T("1"), T(r#"{"key1": {"key": "value"}}"#)],
                            &[T("2"), T(r#"{"num": 42}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSONB large string",
            set_up_script: &[
                "CREATE TABLE t_jsonl (pk INT4 PRIMARY KEY, v1 JSONB);",
                r#"INSERT INTO t_jsonl VALUES (1, '{"key1": "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"}');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pk, length(v1::TEXT) FROM t_jsonl;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("length", INT4)],
                        rows: &[
                            &[T("1"), T("4112")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSONB GROUP BY with equivalent numbers",
            set_up_script: &[
                "CREATE TABLE t (id SERIAL PRIMARY KEY, doc JSONB);",
                r#"INSERT INTO t (doc) VALUES ('{"age":25}'), ('{"age":25}'), ('{"age":25.0}'), ('{"age":30}');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT doc, COUNT(*) FROM t GROUP BY doc ORDER BY doc;",
                    expected: Expected::Rows {
                        columns: &[Column("doc", JSONB), Column("count", INT8)],
                        rows: &[
                            &[T(r#"{"age": 25}"#), T("3")],
                            &[T(r#"{"age": 30}"#), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSONB int64 boundary values",
            set_up_script: &[
                "CREATE TABLE t (id SERIAL PRIMARY KEY, doc JSONB);",
                "INSERT INTO t (doc) VALUES ('-9223372036854775808'::jsonb), ('9223372036854775807'::jsonb);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT doc FROM t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("doc", JSONB)],
                        rows: &[
                            &[T("-9223372036854775808")],
                            &[T("9223372036854775807")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc::text::bigint FROM t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("doc", INT8)],
                        rows: &[
                            &[T("-9223372036854775808")],
                            &[T("9223372036854775807")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Line type",
            set_up_script: &[
                "CREATE TABLE t_line (id INTEGER primary key, v1 LINE);",
                "INSERT INTO t_line VALUES (1, '{1,2,3}'), (2, '{4,5,6}');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_line ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", LINE)],
                        rows: &[
                            &[T("1"), T("{1,2,3}")],
                            &[T("2"), T("{4,5,6}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Lseg type",
            set_up_script: &[
                "CREATE TABLE t_lseg (id INTEGER primary key, v1 LSEG);",
                "INSERT INTO t_lseg VALUES (1, '((1,2),(3,4))'), (2, '((5,6),(7,8))');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_lseg ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", LSEG)],
                        rows: &[
                            &[T("1"), T("[(1,2),(3,4)]")],
                            &[T("2"), T("[(5,6),(7,8)]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Macaddr type",
            set_up_script: &[
                "CREATE TABLE t_macaddr (id INTEGER primary key, v1 MACADDR);",
                "INSERT INTO t_macaddr VALUES (1, '08:00:2b:01:02:03'), (2, '00:11:22:33:44:55');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_macaddr ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", MACADDR)],
                        rows: &[
                            &[T("1"), T("08:00:2b:01:02:03")],
                            &[T("2"), T("00:11:22:33:44:55")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Money type",
            set_up_script: &[
                "CREATE TABLE t_money (id INTEGER primary key, v1 MONEY);",
                "INSERT INTO t_money VALUES (1, '$100.25'), (2, '$50.50');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_money ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", MONEY)],
                        rows: &[
                            &[T("1"), T("$100.25")],
                            &[T("2"), T("$50.50")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Name type",
            set_up_script: &[
                "CREATE TABLE t_name (id INTEGER primary key, v1 NAME);",
                "INSERT INTO t_name VALUES (1, 'abcdefghij'), (2, 'klmnopqrst');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_name ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", NAME)],
                        rows: &[
                            &[T("1"), T("abcdefghij")],
                            &[T("2"), T("klmnopqrst")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_name ORDER BY v1 DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", NAME)],
                        rows: &[
                            &[T("2"), T("klmnopqrst")],
                            &[T("1"), T("abcdefghij")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::char(1) FROM t_name WHERE v1='klmnopqrst';",
                    expected: Expected::Rows {
                        columns: &[Column("v1", BPCHAR)],
                        rows: &[
                            &[T("k")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t_name SET v1='tuvwxyz' WHERE id=2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM t_name WHERE v1='abcdefghij';",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id::name, v1::text FROM t_name ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", NAME), Column("v1", TEXT)],
                        rows: &[
                            &[T("2"), T("tuvwxyz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_name VALUES (3, '0123456789012345678901234567890123456789012345678901234567890123456789');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_name ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", NAME)],
                        rows: &[
                            &[T("2"), T("tuvwxyz")],
                            &[T("3"), T("012345678901234567890123456789012345678901234567890123456789012")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_name VALUES (4, 12345);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_name ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", NAME)],
                        rows: &[
                            &[T("2"), T("tuvwxyz")],
                            &[T("3"), T("012345678901234567890123456789012345678901234567890123456789012")],
                            &[T("4"), T("12345")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT name 'name string' = name 'name string' AS "True";"#,
                    expected: Expected::Rows {
                        columns: &[Column("True", BOOL)],
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
            name: "Name key",
            set_up_script: &[
                "CREATE TABLE t_name (id NAME primary key, v1 NAME);",
                "INSERT INTO t_name VALUES ('wxyz', 'abcdefghij'), ('abcd', 'klmnopqrst');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_name WHERE id = 'wxyz' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", NAME), Column("v1", NAME)],
                        rows: &[
                            &[T("wxyz"), T("abcdefghij")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Name type, explicit casts",
            set_up_script: &[
                "CREATE TABLE t_name (id INTEGER primary key, v1 NAME);",
                "INSERT INTO t_name VALUES (1, 'abcdefghij'), (2, '12345');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_name ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", NAME)],
                        rows: &[
                            &[T("1"), T("abcdefghij")],
                            &[T("2"), T("12345")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::char(1), v1::varchar(2), v1::text FROM t_name WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", BPCHAR), Column("v1", VARCHAR), Column("v1", TEXT)],
                        rows: &[
                            &[T("a"), T("ab"), T("abcdefghij")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::smallint, v1::integer, v1::bigint, v1::float4, v1::float8, v1::numeric FROM t_name WHERE id=2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT2), Column("v1", INT4), Column("v1", INT8), Column("v1", FLOAT4), Column("v1", FLOAT8), Column("v1", NUMERIC)],
                        rows: &[
                            &[T("12345"), T("12345"), T("12345"), T("12345"), T("12345"), T("12345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::oid, v1::xid FROM t_name WHERE id=2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", OID), Column("v1", XID)],
                        rows: &[
                            &[T("12345"), T("12345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::xid FROM t_name WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", XID)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('0'::name)::boolean, ('1'::name)::boolean;",
                    expected: Expected::Rows {
                        columns: &[Column("bool", BOOL), Column("bool", BOOL)],
                        rows: &[
                            &[T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::smallint FROM t_name WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type smallint: "abcdefghij""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::integer FROM t_name WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "abcdefghij""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::bigint FROM t_name WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type bigint: "abcdefghij""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::float4 FROM t_name WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type real: "abcdefghij""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::float8 FROM t_name WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type double precision: "abcdefghij""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::numeric FROM t_name WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type numeric: "abcdefghij""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::boolean FROM t_name WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type boolean: "abcdefghij""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::oid FROM t_name WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type oid: "abcdefghij""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('abc'::char(3))::name, ('abc'::varchar)::name, ('abc'::text)::name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("name", NAME), Column("name", NAME)],
                        rows: &[
                            &[T("abc"), T("abc"), T("abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (10::int2)::name, (100::int4)::name, (1000::int8)::name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("name", NAME), Column("name", NAME)],
                        rows: &[
                            &[T("10"), T("100"), T("1000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1.1::float4)::name, (10.1::float8)::name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("name", NAME)],
                        rows: &[
                            &[T("1.1"), T("10.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (100.0::numeric)::name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME)],
                        rows: &[
                            &[T("100.0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT false::name, true::name, ('0'::boolean)::name, ('1'::boolean)::name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("name", NAME), Column("name", NAME), Column("name", NAME)],
                        rows: &[
                            &[T("f"), T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('123'::xid)::name, (123::oid)::name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("name", NAME)],
                        rows: &[
                            &[T("123"), T("123")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Name array type",
            set_up_script: &[
                "CREATE TABLE t_namea (id INTEGER primary key, v1 NAME[], v2 CHARACTER(100), v3 BOOLEAN);",
                r#"INSERT INTO t_namea VALUES (1, ARRAY['ab''cdef', 'what', 'is,hi', 'wh"at'], '1234567890123456789012345678901234567890123456789012345678901234567890', true);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT v1::varchar(1)[] FROM t_namea;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{a,w,i,w}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v2::name, v3::name FROM t_namea;",
                    expected: Expected::Rows {
                        columns: &[Column("v2", NAME), Column("v3", NAME)],
                        rows: &[
                            &[T("123456789012345678901234567890123456789012345678901234567890123"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Numeric type",
            set_up_script: &[
                "CREATE TABLE t_numeric (id INTEGER primary key, v1 NUMERIC(5,2));",
                "INSERT INTO t_numeric VALUES (1, 123.45), (2, 67.89), (3, 100.3);",
                "CREATE TABLE fract_only (id int, val numeric(4,4));",
                "CREATE TABLE num_data (id int4, val numeric(210,10));",
                "INSERT INTO num_data VALUES (2, '-34338492.215397047');",
                "CREATE TABLE ceil_floor_round (a numeric);",
                "INSERT INTO ceil_floor_round VALUES ('-0.000001');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_numeric ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", NUMERIC)],
                        rows: &[
                            &[T("1"), T("123.45")],
                            &[T("2"), T("67.89")],
                            &[T("3"), T("100.30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO fract_only VALUES (1, '0.0');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT numeric '10.00';",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("10.00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT numeric '-10.00';",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("-10.00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 0.03::numeric(3,3);",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("0.030")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1.03::numeric(2,2);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "numeric field overflow", detail: "A field with precision 2, scale 2 must round to an absolute value less than 1.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1.03::float4::numeric(2,2);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "numeric field overflow", detail: "A field with precision 2, scale 2 must round to an absolute value less than 1.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'NaN'::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'nan'::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-inf'::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("-Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-infinity'::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("-Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'inf'::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'infinity'::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ' 123'::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("123")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT t1.id, t2.id, round(t1.val * t2.val, 30) FROM num_data t1, num_data t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4), Column("round", NUMERIC)],
                        rows: &[
                            &[T("2"), T("2"), T("1179132047626883.596862135856320209000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select sqrt(1.000000000000004::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("sqrt", NUMERIC)],
                        rows: &[
                            &[T("1.000000000000002")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select ln(5.80397490724e5);",
                    expected: Expected::Rows {
                        columns: &[Column("ln", NUMERIC)],
                        rows: &[
                            &[T("13.271468476626518")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 4770999999999999999999999999999999999999999999999999999999999999999999999999999999999999 * 9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", NUMERIC)],
                        rows: &[
                            &[T("47709999999999999999999999999999999999999999999999999999999999999999999999999999999999985229000000000000000000000000000000000000000000000000000000000000000000000000000000000001")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT floor(-0.000001);",
                    expected: Expected::Rows {
                        columns: &[Column("floor", NUMERIC)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '12345'::jsonb::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("12345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Numeric key",
            set_up_script: &[
                "CREATE TABLE t_numeric (id numeric(5,2) primary key, v1 NUMERIC(5,2));",
                "INSERT INTO t_numeric VALUES (123.45, 67.89), (67.89, 100.3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("id", NUMERIC), Column("v1", NUMERIC)],
                        rows: &[
                            &[T("123.45"), T("67.89")],
                            &[T("67.89"), T("100.30")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_numeric order by id",
                    expected: Expected::Rows {
                        columns: &[Column("id", NUMERIC), Column("v1", NUMERIC)],
                        rows: &[
                            &[T("67.89"), T("100.30")],
                            &[T("123.45"), T("67.89")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_numeric WHERE ID = 123.45 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", NUMERIC), Column("v1", NUMERIC)],
                        rows: &[
                            &[T("123.45"), T("67.89")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Numeric type, no scale or precision",
            set_up_script: &[
                "CREATE TABLE t_numeric (id INTEGER primary key, v1 NUMERIC);",
                "INSERT INTO t_numeric VALUES (1, 123.45), (2, 67.875), (3, 100.3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_numeric ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", NUMERIC)],
                        rows: &[
                            &[T("1"), T("123.45")],
                            &[T("2"), T("67.875")],
                            &[T("3"), T("100.3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Numeric array type, no scale or precision",
            set_up_script: &[
                "CREATE TABLE t_numeric (id INTEGER primary key, v1 NUMERIC[]);",
                "INSERT INTO t_numeric VALUES (1, ARRAY[NULL,123.45]), (2, ARRAY[67.89,572903.1468]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_numeric ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", NUMERIC_ARRAY)],
                        rows: &[
                            &[T("1"), T("{NULL,123.45}")],
                            &[T("2"), T("{67.89,572903.1468}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Oid type",
            set_up_script: &[
                "CREATE TABLE t_oid (id INTEGER primary key, v1 OID);",
                "INSERT INTO t_oid VALUES (1, 1234), (2, 5678);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_oid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", OID)],
                        rows: &[
                            &[T("1"), T("1234")],
                            &[T("2"), T("5678")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_oid ORDER BY v1 DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", OID)],
                        rows: &[
                            &[T("2"), T("5678")],
                            &[T("1"), T("1234")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t_oid SET v1=9012 WHERE id=2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM t_oid WHERE v1=1234;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_oid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", OID)],
                        rows: &[
                            &[T("2"), T("9012")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_oid VALUES (3, '2345');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_oid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", OID)],
                        rows: &[
                            &[T("2"), T("9012")],
                            &[T("3"), T("2345")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_oid VALUES (4, 4294967295);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_oid VALUES (5, 4294967296);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "OID out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_oid VALUES (6, 0);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_oid VALUES (7, -1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_oid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", OID)],
                        rows: &[
                            &[T("2"), T("9012")],
                            &[T("3"), T("2345")],
                            &[T("4"), Oid(4294967295)],
                            &[T("6"), T("0")],
                            &[T("7"), Oid(4294967295)],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select oid '20304';",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID)],
                        rows: &[
                            &[Oid(20304)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Oidvector type",
            set_up_script: &[
                "CREATE TABLE t_oidvector (id INTEGER primary key, v1 oidvector);",
                "INSERT INTO t_oidvector VALUES (1, '1234 5678 9012'), (2, '556 778 223');",
                "CREATE TABLE t_regtype_array (v regtype[]);",
                "INSERT INTO t_regtype_array VALUES (ARRAY['integer'::regtype]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY['character varying'::regtype]::oidvector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", OIDVECTOR)],
                        rows: &[
                            &[T("1043")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY['integer'::regtype, 'text'::regtype]::oidvector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", OIDVECTOR)],
                        rows: &[
                            &[T("23 25")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[23::oid]::oidvector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", OIDVECTOR)],
                        rows: &[
                            &[T("23")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ARRAY['pg_class'::regclass]::oidvector =
					ARRAY['pg_class'::regclass::oid]::oidvector;"#,
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
                    query: "SELECT ARRAY['textin'::regproc]::oidvector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", OIDVECTOR)],
                        rows: &[
                            &[T("46")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL::regtype[]::oidvector;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type regtype[] to oidvector", position: 23, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[]::regtype[]::oidvector;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type regtype[] to oidvector", position: 26, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY['integer'::regtype]::regtype[])::oidvector;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type regtype[] to oidvector", position: 46, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v::oidvector FROM t_regtype_array;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type regtype[] to oidvector", position: 9, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY['integer'::regtype] || ARRAY['text'::regtype])::oidvector;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type regtype[] to oidvector", position: 61, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY['integer'::regtype, NULL]::oidvector;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "array is not a valid oidvector", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[ARRAY[23::oid]]::oidvector;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "array is not a valid oidvector", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_oidvector ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", OIDVECTOR)],
                        rows: &[
                            &[T("1"), T("1234 5678 9012")],
                            &[T("2"), T("556 778 223")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select ('16 17'::oidvector)[1];",
                    expected: Expected::Rows {
                        columns: &[Column("oidvector", OID)],
                        rows: &[
                            &[T("17")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '16 17'::oidvector::oid[];",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID_ARRAY)],
                        rows: &[
                            &[T("[0:1]={16,17}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Oidvector array type",
            set_up_script: &[
                "CREATE TABLE t_oidvector (id INTEGER primary key, v1 oidvector[]);",
                r#"INSERT INTO t_oidvector VALUES (1, '{"1234 5678 9012", "556 778 223"}');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_oidvector ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", OIDVECTOR_ARRAY)],
                        rows: &[
                            &[T("1"), T(r#"{"1234 5678 9012","556 778 223"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Oid type, explicit casts",
            set_up_script: &[
                "CREATE TABLE t_oid (id INTEGER primary key, coid OID);",
                "INSERT INTO t_oid VALUES (1, 1234), (2, 4294967295);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_oid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("coid", OID)],
                        rows: &[
                            &[T("1"), T("1234")],
                            &[T("2"), Oid(4294967295)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::char(1) FROM t_oid WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("coid", BPCHAR)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::varchar(2) FROM t_oid WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("coid", VARCHAR)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::text FROM t_oid WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("coid", TEXT)],
                        rows: &[
                            &[T("1234")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::smallint FROM t_oid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type oid to smallint", position: 12, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::smallint FROM t_oid WHERE id=2;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type oid to smallint", position: 12, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::integer FROM t_oid WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("coid", INT4)],
                        rows: &[
                            &[T("1234")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::integer FROM t_oid WHERE id=2;",
                    expected: Expected::Rows {
                        columns: &[Column("coid", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::bigint FROM t_oid WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("coid", INT8)],
                        rows: &[
                            &[T("1234")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::name FROM t_oid WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("coid", NAME)],
                        rows: &[
                            &[T("1234")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::bigint FROM t_oid WHERE id=2;",
                    expected: Expected::Rows {
                        columns: &[Column("coid", INT8)],
                        rows: &[
                            &[T("4294967295")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::float4 FROM t_oid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type oid to real", position: 12, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::float8 FROM t_oid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type oid to double precision", position: 12, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::numeric FROM t_oid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type oid to numeric", position: 12, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT coid::xid FROM t_oid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type oid to xid", position: 12, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('123'::char(3))::oid, ('123'::varchar)::oid, ('0'::text)::oid, ('400'::name)::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("oid", OID), Column("oid", OID), Column("oid", OID)],
                        rows: &[
                            &[T("123"), T("123"), T("0"), T("400")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('-1'::char(3))::oid, ('-1'::varchar)::oid, ('-1'::text)::oid, ('-1'::name)::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("oid", OID), Column("oid", OID), Column("oid", OID)],
                        rows: &[
                            &[Oid(4294967295), Oid(4294967295), Oid(4294967295), Oid(4294967295)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('-2147483648'::char(11))::oid, ('-2147483648'::varchar)::oid, ('-2147483648'::text)::oid, ('-2147483648'::name)::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("oid", OID), Column("oid", OID), Column("oid", OID)],
                        rows: &[
                            &[Oid(2147483648), Oid(2147483648), Oid(2147483648), Oid(2147483648)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (10::int2)::oid, (10::int4)::oid, (100::int8)::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("oid", OID), Column("oid", OID)],
                        rows: &[
                            &[T("10"), T("10"), T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (-1::int2)::oid, (-1::int4)::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("oid", OID)],
                        rows: &[
                            &[Oid(4294967295), Oid(4294967295)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (-1::int8)::oid;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "OID out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (922337203685477580::int8)::oid;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "OID out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1.1::float4)::oid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type real to oid", position: 21, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1.1::float8)::oid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type double precision to oid", position: 21, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1.1::decimal)::oid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type numeric to oid", position: 22, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('922337203685477580'::text)::oid;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#"value "922337203685477580" is out of range for type oid"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('abc'::char(3))::oid;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type oid: "abc""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('-2147483649'::char(11))::oid;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#"value "-2147483649" is out of range for type oid"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('-2147483649'::varchar)::oid;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#"value "-2147483649" is out of range for type oid"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('-2147483649'::text)::oid;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#"value "-2147483649" is out of range for type oid"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('-2147483649'::name)::oid;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#"value "-2147483649" is out of range for type oid"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Oid array type",
            set_up_script: &[
                "CREATE TABLE t_oid (id INTEGER primary key, v1 OID[], v2 CHARACTER(100), v3 BOOLEAN);",
                "INSERT INTO t_oid VALUES (1, ARRAY[123, 456, 789, 101], '1234567890', true);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT v1::varchar(1)[] FROM t_oid;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{1,4,7,1}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v2::oid, v3::oid FROM t_oid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type boolean to oid", position: 19, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Path type",
            set_up_script: &[
                "CREATE TABLE t_path (id INTEGER primary key, v1 PATH);",
                "INSERT INTO t_path VALUES (1, '((1,2),(3,4),(5,6))'), (2, '((7,8),(9,10),(11,12))');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_path ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", PATH)],
                        rows: &[
                            &[T("1"), T("((1,2),(3,4),(5,6))")],
                            &[T("2"), T("((7,8),(9,10),(11,12))")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Pg_lsn type",
            set_up_script: &[
                "CREATE TABLE t_pg_lsn (id INTEGER primary key, v1 PG_LSN);",
                "INSERT INTO t_pg_lsn VALUES (1, '16/B8E36C60'), (2, '16/B8E36C70');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_pg_lsn ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", PG_LSN)],
                        rows: &[
                            &[T("1"), T("16/B8E36C60")],
                            &[T("2"), T("16/B8E36C70")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Point type",
            set_up_script: &[
                "CREATE TABLE t_point (id INTEGER primary key, v1 POINT);",
                "INSERT INTO t_point VALUES (1, '(1,2)'), (2, '(3,4)');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_point ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", POINT)],
                        rows: &[
                            &[T("1"), T("(1,2)")],
                            &[T("2"), T("(3,4)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Polygon type",
            set_up_script: &[
                "CREATE TABLE t_polygon (id INTEGER primary key, v1 POLYGON);",
                "INSERT INTO t_polygon VALUES (1, '((1,2),(3,4),(5,6))'), (2, '((7,8),(9,10),(11,12))');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_polygon ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", POLYGON)],
                        rows: &[
                            &[T("1"), T("((1,2),(3,4),(5,6))")],
                            &[T("2"), T("((7,8),(9,10),(11,12))")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Real type",
            set_up_script: &[
                "CREATE TABLE t_real (id INTEGER primary key, v1 REAL);",
                "INSERT INTO t_real VALUES (1, 123.875), (2, 67.125);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_real ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", FLOAT4)],
                        rows: &[
                            &[T("1"), T("123.875")],
                            &[T("2"), T("67.125")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_real VALUES (3, 1.0e100);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" is out of range for type real"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_real VALUES (3, 1.0e100::numeric);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" is out of range for type real"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3.4e38::float8::real, (-3.4e38)::float8::real;",
                    expected: Expected::Rows {
                        columns: &[Column("float4", FLOAT4), Column("float4", FLOAT4)],
                        rows: &[
                            &[T("3.4e+38"), T("-3.4e+38")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'Infinity'::numeric::real, '-Infinity'::numeric::real, 'Infinity'::numeric::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("float4", FLOAT4), Column("float4", FLOAT4), Column("float8", FLOAT8)],
                        rows: &[
                            &[T("Infinity"), T("-Infinity"), T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('1' || repeat('0', 320))::numeric::float8;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" is out of range for type double precision"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Real key",
            set_up_script: &[
                "CREATE TABLE t_real (id REAL primary key, v1 REAL);",
                "INSERT INTO t_real VALUES (123.875, 67.125), (67.125, 123.875);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_real WHERE ID = 123.875 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", FLOAT4), Column("v1", FLOAT4)],
                        rows: &[
                            &[T("123.875"), T("67.125")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Real array type",
            set_up_script: &[
                "CREATE TABLE t_real (id INTEGER primary key, v1 REAL[]);",
                "INSERT INTO t_real VALUES (1, ARRAY[NULL,123.875]), (2, ARRAY[67.125, 84256]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_real ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", FLOAT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{NULL,123.875}")],
                            &[T("2"), T("{67.125,84256}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Regclass type",
            set_up_script: &[
                "CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);",
                r#"CREATE TABLE "Testing2" (pk INT primary key, v1 INT);"#,
                "CREATE VIEW testview AS SELECT * FROM testing LIMIT 1;",
                "CREATE SEQUENCE seq1;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'testing'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("regclass", REGCLASS)],
                        rows: &[
                            &[T("testing")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'public.testing'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("regclass", REGCLASS)],
                        rows: &[
                            &[T("testing")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'postgres.public.testing'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("regclass", REGCLASS)],
                        rows: &[
                            &[T("testing")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'doesnotexist.public.testing'::regclass;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cross-database references are not implemented: "doesnotexist.public.testing""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'testview'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("regclass", REGCLASS)],
                        rows: &[
                            &[T("testview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ' testing'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("regclass", REGCLASS)],
                        rows: &[
                            &[T("testing")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'seq1'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("regclass", REGCLASS)],
                        rows: &[
                            &[T("seq1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'Testing2'::regclass;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "testing2" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"Testing2"'::regclass;"#,
                    expected: Expected::Rows {
                        columns: &[Column("regclass", REGCLASS)],
                        rows: &[
                            &[T(r#""Testing2""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4294967295::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("regclass", REGCLASS)],
                        rows: &[
                            &[T("4294967295")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT relname FROM pg_catalog.pg_class WHERE oid = 'testing'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME)],
                        rows: &[
                            &[T("testing")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'public.testing'::regclass, 'public.seq1'::regclass, 'public.testview'::regclass, 'public.testing_pkey'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("regclass", REGCLASS), Column("regclass", REGCLASS), Column("regclass", REGCLASS), Column("regclass", REGCLASS)],
                        rows: &[
                            &[T("testing"), T("seq1"), T("testview"), T("testing_pkey")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = '';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'testing'::regclass;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "testing" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'public.testing'::regclass, 'public.seq1'::regclass, 'public.testview'::regclass, 'public.testing_pkey'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("regclass", REGCLASS), Column("regclass", REGCLASS), Column("regclass", REGCLASS), Column("regclass", REGCLASS)],
                        rows: &[
                            &[T("public.testing"), T("public.seq1"), T("public.testview"), T("public.testing_pkey")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Regproc type",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'acos'::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("acos")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ' acos'::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("acos")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"acos"'::regproc;"#,
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("acos")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (('acos'::regproc)::oid)::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("acos")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ((('acos'::regproc)::oid)::text)::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("acos")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4294967295::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("4294967295")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"Abs"'::regproc;"#,
                    expected: Expected::Error(Diagnostic { code: "42883", message: r#"function ""Abs"" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"acos'::regproc;"#,
                    expected: Expected::Error(Diagnostic { code: "42602", message: "invalid name syntax", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 'acos"'::regproc;"#,
                    expected: Expected::Error(Diagnostic { code: "42883", message: r#"function "acos"" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '""acos'::regproc;"#,
                    expected: Expected::Error(Diagnostic { code: "42602", message: "invalid name syntax", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'pg_catalog.acos'::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("acos")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT typinput = 'pg_catalog.array_in'::regproc FROM pg_catalog.pg_type WHERE typname = 'int4';",
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
                    query: "SELECT typinput = 'pg_catalog.array_in'::regproc FROM pg_catalog.pg_type WHERE typname = '_int4';",
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
                    query: "SELECT 'public.acos'::regproc;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: r#"function "public.acos" does not exist"#, position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Regtype type",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'integer'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'integer'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'integer[]'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("integer[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'int4'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'float8'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("double precision")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'character varying'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("character varying")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"char"'::regtype;"#,
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T(r#""char""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'char'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("character")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'char(10)'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("character")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"char"'::regtype::oid;"#,
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID)],
                        rows: &[
                            &[T("18")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'char'::regtype::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID)],
                        rows: &[
                            &[T("1042")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"char"[]'::regtype;"#,
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T(r#""char"[]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ' integer'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"integer"'::regtype;"#,
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "integer" does not exist"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (('integer'::regtype)::oid)::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ((('integer'::regtype)::oid)::text)::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4294967295::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("4294967295")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"Integer"'::regtype;"#,
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "Integer" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"integer'::regtype;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"unterminated quoted identifier at or near ""integer""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 'integer"'::regtype;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"unterminated quoted identifier at or near """"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '""integer'::regtype;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"zero-length delimited identifier at or near """""#, position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Smallint type",
            set_up_script: &[
                "CREATE TABLE t_smallint (id INTEGER primary key, v1 SMALLINT);",
                "INSERT INTO t_smallint VALUES (1, 42), (2, 99);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_smallint ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT2)],
                        rows: &[
                            &[T("1"), T("42")],
                            &[T("2"), T("99")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Int2vector type",
            set_up_script: &[
                "CREATE TABLE t_int2vector (id INTEGER primary key, v1 int2vector);",
                "INSERT INTO t_int2vector VALUES (1, '1 2 3'), (2, '6 7 8 9');",
                "CREATE TABLE t_int2_array (v int2[]);",
                "INSERT INTO t_int2_array VALUES (ARRAY[1::int2]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1, 2]::int2vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT2VECTOR)],
                        rows: &[
                            &[T("1 2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1::bigint, 2::bigint]::int2vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT2VECTOR)],
                        rows: &[
                            &[T("1 2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1.0::numeric, 2.0::numeric]::int2vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT2VECTOR)],
                        rows: &[
                            &[T("1 2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1, NULL]::int2vector;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "array is not a valid int2vector", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[ARRAY[1]]::int2vector;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "array is not a valid int2vector", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL::int2[]::int2vector;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type smallint[] to int2vector", position: 20, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[]::int2[]::int2vector;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type smallint[] to int2vector", position: 23, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[1::int2]::int2[])::int2vector;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type smallint[] to int2vector", position: 32, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v::int2vector FROM t_int2_array;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type smallint[] to int2vector", position: 9, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[1::int2] || ARRAY[2::int2])::int2vector;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type smallint[] to int2vector", position: 42, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_int2vector ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT2VECTOR)],
                        rows: &[
                            &[T("1"), T("1 2 3")],
                            &[T("2"), T("6 7 8 9")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest(v1) FROM t_int2vector ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT2)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("6")],
                            &[T("7")],
                            &[T("8")],
                            &[T("9")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Int2vector array type",
            set_up_script: &[
                "CREATE TABLE t_int2vector (id INTEGER primary key, v1 int2vector[]);",
                r#"INSERT INTO t_int2vector VALUES (1, '{"1 2", "3 4"}');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_int2vector ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT2VECTOR_ARRAY)],
                        rows: &[
                            &[T("1"), T(r#"{"1 2","3 4"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest(v1) FROM t_int2vector ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT2VECTOR)],
                        rows: &[
                            &[T("1 2")],
                            &[T("3 4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest(unnest(v1)) FROM t_int2vector ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT2)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Domains over vector types",
            set_up_script: &[
                "CREATE DOMAIN two_int2vector AS int2vector CHECK (array_length(VALUE, 1) = 2);",
                "CREATE DOMAIN nonnull_oidvector AS oidvector NOT NULL CHECK (array_length(VALUE, 1) <= 2);",
                "CREATE DOMAIN null_rejecting_int2vector AS int2vector CHECK (VALUE IS NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1, 2]::two_int2vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT2VECTOR)],
                        rows: &[
                            &[T("1 2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1]::two_int2vector;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain two_int2vector violates check constraint "two_int2vector_check""#, schema: "public", data_type: "two_int2vector", constraint: "two_int2vector_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[23::oid, 25::oid]::nonnull_oidvector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", OIDVECTOR)],
                        rows: &[
                            &[T("23 25")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[23::oid, 25::oid, 26::oid]::nonnull_oidvector;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain nonnull_oidvector violates check constraint "nonnull_oidvector_check""#, schema: "public", data_type: "nonnull_oidvector", constraint: "nonnull_oidvector_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL::nonnull_oidvector;",
                    expected: Expected::Error(Diagnostic { code: "23502", message: "domain nonnull_oidvector does not allow null values", schema: "public", data_type: "nonnull_oidvector", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL::null_rejecting_int2vector;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain null_rejecting_int2vector violates check constraint "null_rejecting_int2vector_check""#, schema: "public", data_type: "null_rejecting_int2vector", constraint: "null_rejecting_int2vector_check", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Smallint key",
            set_up_script: &[
                "CREATE TABLE t_smallint (id smallint primary key, v1 SMALLINT);",
                "INSERT INTO t_smallint VALUES (1, 42), (2, 99);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_smallint WHERE ID = 1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT2), Column("v1", INT2)],
                        rows: &[
                            &[T("1"), T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Smallint array type",
            set_up_script: &[
                "CREATE TABLE t_smallint (id INTEGER primary key, v1 SMALLINT[]);",
                "INSERT INTO t_smallint VALUES (1, ARRAY[42,NULL]), (2, ARRAY[99,126]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_smallint ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT2_ARRAY)],
                        rows: &[
                            &[T("1"), T("{42,NULL}")],
                            &[T("2"), T("{99,126}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Smallserial type",
            set_up_script: &[
                "CREATE TABLE t_smallserial (id SERIAL primary key, v1 SMALLSERIAL);",
                "INSERT INTO t_smallserial (v1) VALUES (42), (99);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_smallserial ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT2)],
                        rows: &[
                            &[T("1"), T("42")],
                            &[T("2"), T("99")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Smallserial key",
            set_up_script: &[
                "CREATE TABLE t_smallserial (id smallserial primary key, v1 SMALLSERIAL);",
                "INSERT INTO t_smallserial (v1) VALUES (42), (99);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_smallserial WHERE ID = 1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT2), Column("v1", INT2)],
                        rows: &[
                            &[T("1"), T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Serial type",
            set_up_script: &[
                "CREATE TABLE t_serial (id SERIAL primary key, v1 SERIAL);",
                "INSERT INTO t_serial (v1) VALUES (123), (456);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_serial ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("123")],
                            &[T("2"), T("456")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_serial WHERE ID = 2 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("2"), T("456")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Text type",
            set_up_script: &[
                "CREATE TABLE t_text (id INTEGER primary key, v1 TEXT);",
                "INSERT INTO t_text VALUES (1, 'Hello'), (2, 'World'), (3, ''), (4, NULL);",
                "CREATE TABLE t_text_unique (id INTEGER primary key, v1 TEXT, v2 TEXT NOT NULL UNIQUE);",
                "INSERT INTO t_text_unique VALUES (1, 'Hello', 'Bonjour'), (2, 'World', 'tout le monde'), (3, '', ''), (4, NULL, '!');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT text 'text' || ' and unknown';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("text and unknown")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT text 'this is a text string' = text 'this is a text string' AS true;",
                    expected: Expected::Rows {
                        columns: &[Column("true", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_text ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("Hello")],
                            &[T("2"), T("World")],
                            &[T("3"), T("")],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE UNIQUE INDEX v1_unique ON t_text(v1);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_text WHERE v1 = 'World';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("2"), T("World")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_text VALUES (5, 'World');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "v1_unique""#, detail: "Key (v1)=(World) already exists.", schema: "public", table: "t_text", constraint: "v1_unique", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_text_unique WHERE v2 = '!';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TEXT), Column("v2", TEXT)],
                        rows: &[
                            &[T("4"), Null, T("!")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_text_unique WHERE v2 >= '!' ORDER BY v2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TEXT), Column("v2", TEXT)],
                        rows: &[
                            &[T("4"), Null, T("!")],
                            &[T("1"), T("Hello"), T("Bonjour")],
                            &[T("2"), T("World"), T("tout le monde")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_text_unique ORDER BY v2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TEXT), Column("v2", TEXT)],
                        rows: &[
                            &[T("3"), T(""), T("")],
                            &[T("4"), Null, T("!")],
                            &[T("1"), T("Hello"), T("Bonjour")],
                            &[T("2"), T("World"), T("tout le monde")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_text_unique ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TEXT), Column("v2", TEXT)],
                        rows: &[
                            &[T("1"), T("Hello"), T("Bonjour")],
                            &[T("2"), T("World"), T("tout le monde")],
                            &[T("3"), T(""), T("")],
                            &[T("4"), Null, T("!")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_text_unique VALUES (5, 'Another', 'Bonjour');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t_text_unique_v2_key""#, detail: "Key (v2)=(Bonjour) already exists.", schema: "public", table: "t_text_unique", constraint: "t_text_unique_v2_key", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX on t_text_unique(v1, v2);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t_text_unique WHERE v1='Hello' and v2='Bonjour';",
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
                    query: "CREATE TABLE t2 (pk int primary key, c1 TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx1 ON t2(c1);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (1, 'one'), (2, 'two');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c1 from t2 order by c1;",
                    expected: Expected::Rows {
                        columns: &[Column("c1", TEXT)],
                        rows: &[
                            &[T("one")],
                            &[T("two")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Text key",
            set_up_script: &[
                "CREATE TABLE t_text (id TEXT primary key, v1 TEXT);",
                "INSERT INTO t_text VALUES ('Hello', 'World'), ('goodbye', 'cruel world');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_text where id = 'goodbye' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", TEXT), Column("v1", TEXT)],
                        rows: &[
                            &[T("goodbye"), T("cruel world")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Text key with mixed short and long values",
            set_up_script: &[
                "CREATE TABLE t_text_keys (id TEXT primary key, v1 INTEGER);",
                "INSERT INTO t_text_keys VALUES ('aa', 1);",
                "INSERT INTO t_text_keys VALUES ('bb' || repeat('x', 10500), 2);",
                "INSERT INTO t_text_keys VALUES ('cc', 3);",
                "INSERT INTO t_text_keys VALUES ('dd' || repeat('y', 10500), 4);",
                "INSERT INTO t_text_keys VALUES ('ee', 5);",
                "INSERT INTO t_text_keys VALUES ('zz', 6);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_text_keys ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_text_keys ORDER BY id DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("6")],
                            &[T("5")],
                            &[T("4")],
                            &[T("3")],
                            &[T("2")],
                            &[T("1")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_text_keys WHERE id = 'aa';",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_text_keys WHERE id = 'zz';",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_text_keys WHERE id = 'bb' || repeat('x', 10500);",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_text_keys WHERE id = 'dd' || repeat('y', 10500);",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_text_keys WHERE id < 'cc' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_text_keys WHERE id > 'dd' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1 FROM t_text_keys WHERE id > 'bb' AND id < 'ee' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1, length(id) FROM t_text_keys WHERE v1 IN (1, 2, 4) ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("length", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("10502")],
                            &[T("4"), T("10502")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select id from t_text_keys order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", TEXT)],
                        rows: &[
                            &[T("aa")],
                            &[T("bbxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")],
                            &[T("cc")],
                            &[T("ddyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy")],
                            &[T("ee")],
                            &[T("zz")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Time without time zone type",
            set_up_script: &[
                "CREATE TABLE t_time_without_zone (id INTEGER primary key, v1 TIME);",
                "INSERT INTO t_time_without_zone VALUES (1, '12:34:56'), (2, '23:45:01'), (3, '02:03 EDT');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_time_without_zone ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TIME)],
                        rows: &[
                            &[T("1"), T("12:34:56")],
                            &[T("2"), T("23:45:01")],
                            &[T("3"), T("02:03:00")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::interval FROM t_time_without_zone ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INTERVAL)],
                        rows: &[
                            &[T("12:34:56")],
                            &[T("23:45:01")],
                            &[T("02:03:00")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '00:00:00'::time;",
                    expected: Expected::Rows {
                        columns: &[Column("time", TIME)],
                        rows: &[
                            &[T("00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '23:59:59.999999'::time;",
                    expected: Expected::Rows {
                        columns: &[Column("time", TIME)],
                        rows: &[
                            &[T("23:59:59.999999")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT time without time zone '040506.789+08';",
                    expected: Expected::Rows {
                        columns: &[Column("time", TIME)],
                        rows: &[
                            &[T("04:05:06.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT time '04:05:06' + date '2025-07-21';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2025-07-21 04:05:06")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT time without time zone '04:05:06' + interval '2 minutes';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIME)],
                        rows: &[
                            &[T("04:07:06")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Time without time zone key",
            set_up_script: &[
                "CREATE TABLE t_time_without_zone (id TIME primary key, v1 TIME);",
                "INSERT INTO t_time_without_zone VALUES ('12:34:56', '23:45:01'), ('23:45:01', '12:34:56');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_time_without_zone WHERE ID = '12:34:56' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", TIME), Column("v1", TIME)],
                        rows: &[
                            &[T("12:34:56"), T("23:45:01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Time with time zone type",
            set_up_script: &[
                "CREATE TABLE t_time_with_zone (id INTEGER primary key, v1 TIME WITH TIME ZONE);",
                "INSERT INTO t_time_with_zone VALUES (1, '12:34:56 UTC'), (2, '23:45:01-0200'), (3, '2025-06-03 02:03 EDT');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_time_with_zone ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TIMETZ)],
                        rows: &[
                            &[T("1"), T("12:34:56+00")],
                            &[T("2"), T("23:45:01-02")],
                            &[T("3"), T("02:03:00-04")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET TIMEZONE TO 'UTC';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '00:00:00'::timetz;",
                    expected: Expected::Rows {
                        columns: &[Column("timetz", TIMETZ)],
                        rows: &[
                            &[T("00:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT time with time zone '04:05:06 UTC' + date '2025-07-21';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2025-07-21 04:05:06+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT time with time zone '04:05:06 UTC' + interval '2 minutes';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMETZ)],
                        rows: &[
                            &[T("04:07:06+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET TIMEZONE TO DEFAULT;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '00:00:00-07'::timetz;",
                    expected: Expected::Rows {
                        columns: &[Column("timetz", TIMETZ)],
                        rows: &[
                            &[T("00:00:00-07")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Timestamp without time zone type",
            set_up_script: &[
                "CREATE TABLE t_timestamp_without_zone (id INTEGER primary key, v1 TIMESTAMP);",
                "INSERT INTO t_timestamp_without_zone VALUES (1, '2022-01-01 12:34:56'), (2, '2022-02-01 23:45:01'), (3, 'Feb 10 5:32PM 1997'), (4, 'Feb 10 16:32:05 99');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_timestamp_without_zone ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TIMESTAMP)],
                        rows: &[
                            &[T("1"), T("2022-01-01 12:34:56")],
                            &[T("2"), T("2022-02-01 23:45:01")],
                            &[T("3"), T("1997-02-10 17:32:00")],
                            &[T("4"), T("1999-02-10 16:32:05")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2000-01-01'::timestamp;",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("2000-01-01 00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2000-01-01 00:00:00'::timestamp;",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("2000-01-01 00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp without time zone '2025-07-21 04:05:06' + interval '2 minutes';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2025-07-21 04:07:06")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Timestamp with time zone type",
            set_up_script: &[
                "CREATE TABLE t_timestamp_with_zone (id INTEGER primary key, v1 TIMESTAMP WITH TIME ZONE);",
                "INSERT INTO t_timestamp_with_zone VALUES (1, '2022-01-01 12:34:56 UTC'), (2, '2022-02-01 23:45:01 America/New_York');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SET timezone TO '-04:25'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_timestamp_with_zone ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1"), T("2022-01-01 12:34:56+00")],
                            &[T("2"), T("2022-02-02 04:45:01+00")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2000-01-01'::timestamptz;",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1999-12-31 19:35:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2000-01-01 00:00:00'::timestamptz;",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1999-12-31 19:35:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone TO '-06:00'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_timestamp_with_zone ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1"), T("2022-01-01 12:34:56+00")],
                            &[T("2"), T("2022-02-02 04:45:01+00")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp with time zone '2025-07-21 04:05:06 UTC' + interval '2 minutes';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2025-07-21 04:07:06+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone TO default",
                    expected: Expected::Tag("SET"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Tsquery type",
            set_up_script: &[
                "CREATE TABLE t_tsquery (id INTEGER primary key, v1 TSQUERY);",
                "INSERT INTO t_tsquery VALUES (1, 'word'), (2, 'phrase & (another | term)');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_tsquery ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TSQUERY)],
                        rows: &[
                            &[T("1"), T("'word'")],
                            &[T("2"), T("'phrase' & ( 'another' | 'term' )")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Tsvector type",
            set_up_script: &[
                "CREATE TABLE t_tsvector (id INTEGER primary key, v1 TSVECTOR);",
                "INSERT INTO t_tsvector VALUES (1, 'simple'), (2, 'complex & (query | terms)');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_tsvector ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", TSVECTOR)],
                        rows: &[
                            &[T("1"), T("'simple'")],
                            &[T("2"), T("'&' '(query' 'complex' 'terms)' '|'")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tsvector unsupported error",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t_tsvector (id INTEGER primary key, v1 TSVECTOR);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Uuid type",
            set_up_script: &[
                "CREATE TABLE t_uuid (id INTEGER primary key, v1 UUID);",
                "INSERT INTO t_uuid VALUES (1, 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'), (2, 'f47ac10b58cc4372a567-0e02b2c3d479');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_uuid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", UUID)],
                        rows: &[
                            &[T("1"), T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")],
                            &[T("2"), T("f47ac10b-58cc-4372-a567-0e02b2c3d479")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select uuid 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';",
                    expected: Expected::Rows {
                        columns: &[Column("uuid", UUID)],
                        rows: &[
                            &[T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Uuid default value",
            set_up_script: &[
                "CREATE TABLE t_uuid (id INTEGER primary key, v1 UUID default 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid);",
                "INSERT INTO t_uuid VALUES (1, 'f47ac10b58cc4372a567-0e02b2c3d479');",
                "INSERT INTO t_uuid (id) VALUES (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_uuid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", UUID)],
                        rows: &[
                            &[T("1"), T("f47ac10b-58cc-4372-a567-0e02b2c3d479")],
                            &[T("2"), T("a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Uuid key",
            set_up_script: &[
                "CREATE TABLE t_uuid (id UUID primary key, v1 UUID);",
                "INSERT INTO t_uuid VALUES ('f47ac10b58cc4372a567-0e02b2c3d479', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'), ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'f47ac10b58cc4372a567-0e02b2c3d479');",
                "create table t_uuid2 (id int primary key, v1 uuid, v2 uuid);",
                "create index on t_uuid2(v1, v2);",
                "insert into t_uuid2 values (1, 'f47ac10b58cc4372a567-0e02b2c3d479', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'), (2, 'dcf783c8-49c2-44b4-8b90-34ad8c52ea1e', 'f99802e8-0018-4913-806c-bcad5d246d46');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_uuid WHERE ID = 'f47ac10b58cc4372a567-0e02b2c3d479' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", UUID), Column("v1", UUID)],
                        rows: &[
                            &[T("f47ac10b-58cc-4372-a567-0e02b2c3d479"), T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_uuid2 WHERE v1 = 'f47ac10b58cc4372a567-0e02b2c3d479' and v2 = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", UUID), Column("v2", UUID)],
                        rows: &[
                            &[T("1"), T("f47ac10b-58cc-4372-a567-0e02b2c3d479"), T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_uuid2 WHERE v1 < 'f47ac10b58cc4372a567-0e02b2c3d479' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", UUID), Column("v2", UUID)],
                        rows: &[
                            &[T("2"), T("dcf783c8-49c2-44b4-8b90-34ad8c52ea1e"), T("f99802e8-0018-4913-806c-bcad5d246d46")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Uuid array type",
            set_up_script: &[
                "CREATE TABLE t_uuid (id INTEGER primary key, v1 UUID[]);",
                "INSERT INTO t_uuid VALUES (1, ARRAY['a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid, NULL]), (2, ARRAY[NULL, 'f47ac10b58cc4372a567-0e02b2c3d479'::uuid]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_uuid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", UUID_ARRAY)],
                        rows: &[
                            &[T("1"), T("{a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11,NULL}")],
                            &[T("2"), T("{NULL,f47ac10b-58cc-4372-a567-0e02b2c3d479}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Xid type",
            set_up_script: &[
                "CREATE TABLE t_xid (id INTEGER primary key, v1 XID, v2 VARCHAR(20));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t_xid VALUES (1, 1234, '100');",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v1" is of type xid but expression is of type integer"#, hint: "You will need to rewrite or cast the expression.", position: 30, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xid VALUES (1, 1234::xid, '100');",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type integer to xid", position: 34, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xid VALUES (1, NULL, '100');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_xid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", XID), Column("v2", VARCHAR)],
                        rows: &[
                            &[T("1"), Null, T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xid VALUES (2, '100', '101');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_xid WHERE v1 IS NOT NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", XID), Column("v2", VARCHAR)],
                        rows: &[
                            &[T("2"), T("100"), T("101")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t_xid SET v1='9012' WHERE id=1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM t_xid WHERE v1=100;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_xid ORDER BY v1 DESC;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "could not identify an ordering operator for type xid", hint: "Use an explicit ordering operator or modify the query.", position: 30, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xid VALUES (4, '4294967295', 'a');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xid VALUES (5, '4294967296', 'b');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xid VALUES (6, '0', 'c');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xid VALUES (7, '-1', 'd');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xid VALUES (8, 'abc', 'd');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_xid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", XID), Column("v2", VARCHAR)],
                        rows: &[
                            &[T("1"), T("9012"), T("100")],
                            &[T("4"), T("4294967295"), T("a")],
                            &[T("5"), T("0"), T("b")],
                            &[T("6"), T("0"), T("c")],
                            &[T("7"), T("4294967295"), T("d")],
                            &[T("8"), T("0"), T("d")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Xid type, explicit casts",
            set_up_script: &[
                "CREATE TABLE t_xid (id INTEGER primary key, v1 XID);",
                "INSERT INTO t_xid VALUES (1, '1234'), (2, '4294967295');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_xid ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", XID)],
                        rows: &[
                            &[T("1"), T("1234")],
                            &[T("2"), T("4294967295")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::char(1), v1::varchar(2), v1::text, v1::name FROM t_xid WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", BPCHAR), Column("v1", VARCHAR), Column("v1", TEXT), Column("v1", NAME)],
                        rows: &[
                            &[T("1"), T("12"), T("1234"), T("1234")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::smallint FROM t_xid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type xid to smallint", position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::integer FROM t_xid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type xid to integer", position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::bigint FROM t_xid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type xid to bigint", position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::oid FROM t_xid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type xid to oid", position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::float4 FROM t_xid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type xid to real", position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::float8 FROM t_xid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type xid to double precision", position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::numeric FROM t_xid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type xid to numeric", position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v1::boolean FROM t_xid WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type xid to boolean", position: 10, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('123'::char(3))::xid, ('123'::varchar)::xid, ('0'::text)::xid, ('400'::name)::xid;",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID), Column("xid", XID), Column("xid", XID), Column("xid", XID)],
                        rows: &[
                            &[T("123"), T("123"), T("0"), T("400")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('-1'::char(3))::xid, ('-1'::varchar)::xid, ('-1'::text)::xid, ('-1'::name)::xid;",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID), Column("xid", XID), Column("xid", XID), Column("xid", XID)],
                        rows: &[
                            &[T("4294967295"), T("4294967295"), T("4294967295"), T("4294967295")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('-2147483648'::char(11))::xid, ('-2147483648'::varchar)::xid, ('-2147483648'::text)::xid, ('-2147483648'::name)::xid;",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID), Column("xid", XID), Column("xid", XID), Column("xid", XID)],
                        rows: &[
                            &[T("2147483648"), T("2147483648"), T("2147483648"), T("2147483648")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (10::int2)::xid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type smallint to xid", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (10::boolean)::xid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type boolean to xid", position: 21, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (10::int4)::xid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type integer to xid", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (10::int8)::xid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type bigint to xid", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1.1::float4)::xid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type real to xid", position: 21, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1.1::float8)::xid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type double precision to xid", position: 21, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (1.1::decimal)::xid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type numeric to xid", position: 22, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('4294967295'::text)::xid, ('4294967297'::text)::xid;",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID), Column("xid", XID)],
                        rows: &[
                            &[T("4294967295"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('-4294967295'::text)::xid, ('-4294967297'::text)::xid;",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID), Column("xid", XID)],
                        rows: &[
                            &[T("1"), T("4294967295")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('4294967295'::varchar)::xid, ('4294967296232'::varchar)::xid;",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID), Column("xid", XID)],
                        rows: &[
                            &[T("4294967295"), T("232")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('-4294967295'::varchar)::xid, ('-4294967296232'::varchar)::xid;",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID), Column("xid", XID)],
                        rows: &[
                            &[T("1"), T("4294967064")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('4294967295'::char(11))::xid, ('4294967296'::char(11))::xid;",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID), Column("xid", XID)],
                        rows: &[
                            &[T("4294967295"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('4294967295'::name)::xid, ('4294967296'::name)::xid;",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID), Column("xid", XID)],
                        rows: &[
                            &[T("4294967295"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('abc'::text)::xid, ('abc'::char(3))::xid, ('abc'::varchar)::xid, ('abc'::name)::xid;",
                    expected: Expected::Rows {
                        columns: &[Column("xid", XID), Column("xid", XID), Column("xid", XID), Column("xid", XID)],
                        rows: &[
                            &[T("0"), T("0"), T("0"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Xid array type",
            set_up_script: &[
                "CREATE TABLE t_xid (id INTEGER primary key, v1 XID[], v2 CHARACTER(100), v3 BOOLEAN);",
                "INSERT INTO t_xid VALUES (2, '{123, 456, 789, 101}', '1234567890', true);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT v1::varchar(1)[] FROM t_xid;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{1,4,7,1}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xid VALUES (2, ARRAY[123, 456, 789, 101], '1234567890', true);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v1" is of type xid[] but expression is of type integer[]"#, hint: "You will need to rewrite or cast the expression.", position: 30, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Xml type",
            set_up_script: &[
                "CREATE TABLE t_xml (id INTEGER primary key, v1 XML);",
                "INSERT INTO t_xml VALUES (1, '<note><to>Tove</to><from>Jani</from><body>Don''t forget me this weekend!</body></note>'), (2, '<book><title>Introduction to Golang</title><author>John Doe</author></book>');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_xml ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", XML)],
                        rows: &[
                            &[T("1"), T("<note><to>Tove</to><from>Jani</from><body>Don't forget me this weekend!</body></note>")],
                            &[T("2"), T("<book><title>Introduction to Golang</title><author>John Doe</author></book>")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xml VALUES (3, '<a>');",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Premature end of data in tag a line 1
<a>
   ^"#, position: 30, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xml VALUES (3, 1);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v1" is of type xml but expression is of type integer"#, hint: "You will need to rewrite or cast the expression.", position: 30, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xml VALUES (3, NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, pg_typeof(v1) FROM t_xml WHERE id < 3 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("1"), T("xml")],
                            &[T("2"), T("xml")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT data_type, udt_name FROM information_schema.columns WHERE table_name = 't_xml' AND column_name = 'v1';",
                    expected: Expected::Rows {
                        columns: &[Column("data_type", VARCHAR), Column("udt_name", NAME)],
                        rows: &[
                            &[T("xml"), T("xml")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT typname FROM pg_catalog.pg_type WHERE oid = 142;",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME)],
                        rows: &[
                            &[T("xml")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Polymorphic types",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_append(ARRAY[1], 2);",
                    expected: Expected::Rows {
                        columns: &[Column("array_append", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(ARRAY['abc','def'], 'ghi');",
                    expected: Expected::Rows {
                        columns: &[Column("array_append", TEXT_ARRAY)],
                        rows: &[
                            &[T("{abc,def,ghi}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(ARRAY['abc','def'], null);",
                    expected: Expected::Rows {
                        columns: &[Column("array_append", TEXT_ARRAY)],
                        rows: &[
                            &[T("{abc,def,NULL}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(null, null);",
                    expected: Expected::Rows {
                        columns: &[Column("array_append", TEXT_ARRAY)],
                        rows: &[
                            &[T("{NULL}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(null, 'ghi');",
                    expected: Expected::Rows {
                        columns: &[Column("array_append", TEXT_ARRAY)],
                        rows: &[
                            &[T("{ghi}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(null, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("array_append", INT4_ARRAY)],
                        rows: &[
                            &[T("{3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(1, 2);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_append(integer, integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(1, ARRAY[2]);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_append(integer, integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(ARRAY[1], ARRAY[2]);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_append(integer[], integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Character comparisons ignore trailing spaces",
            set_up_script: &[
                "CREATE TABLE t_bpchar (id INT PRIMARY KEY, c CHAR(3), u CHAR(3) UNIQUE, v INT);",
                "INSERT INTO t_bpchar VALUES (1, 'a', 'x', 10), (2, 'a ', 'y ', 20), (3, 'b', 'z', 30);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, c, COUNT(*) OVER (PARTITION BY c), SUM(v) OVER (PARTITION BY c) FROM t_bpchar ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", BPCHAR), Column("count", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("1"), T("a  "), T("2"), T("30")],
                            &[T("2"), T("a  "), T("2"), T("30")],
                            &[T("3"), T("b  "), T("1"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c, COUNT(*), SUM(v) FROM t_bpchar GROUP BY c ORDER BY c;",
                    expected: Expected::Rows {
                        columns: &[Column("c", BPCHAR), Column("count", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("a  "), T("2"), T("30")],
                            &[T("b  "), T("1"), T("30")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT c FROM t_bpchar ORDER BY c;",
                    expected: Expected::Rows {
                        columns: &[Column("c", BPCHAR)],
                        rows: &[
                            &[T("a  ")],
                            &[T("b  ")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t_bpchar WHERE c = 'a' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t_bpchar WHERE u = 'y';",
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
                    query: "INSERT INTO t_bpchar VALUES (4, 'c', 'y', 40);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t_bpchar_u_key""#, detail: "Key (u)=(y  ) already exists.", schema: "public", table: "t_bpchar", constraint: "t_bpchar_u_key", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c::text, length(c), c || '|' FROM t_bpchar WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("c", TEXT), Column("length", INT4), Column("?column?", TEXT)],
                        rows: &[
                            &[T("a"), T("1"), T("a|")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'ab  '::char(3) = 'ab'::char(3), 'a  '::bpchar = 'a'::bpchar, length('ab  '::char(3));",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("length", INT4)],
                        rows: &[
                            &[T("t"), T("t"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Casting a bpchar value to another string type removes its trailing spaces",
            set_up_script: &[
                "CREATE TABLE t3325 (id INT PRIMARY KEY, c CHAR(2) CHECK (c::text IN ('L', 'R')));",
                "CREATE TABLE t3325_check (c CHARACTER(2), CONSTRAINT t3325_check_check CHECK (c::text IN ('L', 'M', 'H')));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[' || 'L '::character(2)::text || ']' AS as_text, length('L '::character(2)::text) AS length, 'L '::character(2)::text = 'L' AS equals_l;",
                    expected: Expected::Rows {
                        columns: &[Column("as_text", TEXT), Column("length", INT4), Column("equals_l", BOOL)],
                        rows: &[
                            &[T("[L]"), T("1"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3325_check VALUES ('L ');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[' || c::text || ']' AS as_text FROM t3325_check;",
                    expected: Expected::Rows {
                        columns: &[Column("as_text", TEXT)],
                        rows: &[
                            &[T("[L]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3325 VALUES (1, 'L'), (2, 'R ');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, '[' || c || ']', c = 'L' FROM t3325 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("?column?", TEXT), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("[L]"), T("t")],
                            &[T("2"), T("[R]"), T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[' || 'L'::CHAR(2) || ']', length('L'::CHAR(2)), 'L'::CHAR(2) = 'L', 'L '::CHAR(2) = 'L'::CHAR(2), 'L'::CHAR(2)::TEXT = 'L', 'L'::CHAR(2)::VARCHAR = 'L', bpcharcmp('L'::CHAR(2), 'L ');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT), Column("length", INT4), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("bpcharcmp", INT4)],
                        rows: &[
                            &[T("[L]"), T("1"), T("t"), T("t"), T("t"), T("t"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[' || 'L '::char(2) || ']', upper('L '::char(2)) = 'L', 'L '::bpchar = 'L'::bpchar, 'L '::character(2)::name = 'L', 'L '::character(2)::varchar(5) = 'L', length('L '::character(2)::varchar);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("length", INT4)],
                        rows: &[
                            &[T("[L]"), T("t"), T("t"), T("t"), T("t"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '[' || E'L\t'::character(3)::text || ']', length(E'L\t'::character(3)::text), '[' || E'L\n'::character(3)::text || ']', '[' || E'L \t '::character(5)::text || ']', E'L\t'::character(3) = 'L', E'L\t '::character(4) = E'L\t'::character(3), bpcharcmp(E'L\t'::character(3), E'L\t '::character(4)), '[' || E'L\t'::character(3)::varchar || ']', length(E'L\t'::character(3)), E'L\t'::character(3)::text = E'L\t';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT), Column("length", INT4), Column("?column?", TEXT), Column("?column?", TEXT), Column("?column?", BOOL), Column("?column?", BOOL), Column("bpcharcmp", INT4), Column("?column?", TEXT), Column("length", INT4), Column("?column?", BOOL)],
                        rows: &[
                            &[T("[L\t]"), T("2"), T(r#"[L
]"#), T("[L \t]"), T("f"), T("t"), T("0"), T("[L\t]"), T("2"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Xml literals",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '<a>x</a>'::xml AS doc;",
                    expected: Expected::Rows {
                        columns: &[Column("doc", XML)],
                        rows: &[
                            &[T("<a>x</a>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof('<a>x</a>'::xml);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("xml")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 'x'::xml, ''::xml, '<a/><b/>'::xml, '<!-- c --><a/>'::xml, '<a><![CDATA[<x>]]></a>'::xml, '<a xmlns:p="urn:x"><p:b/></a>'::xml;"#,
                    expected: Expected::Rows {
                        columns: &[Column("xml", XML), Column("xml", XML), Column("xml", XML), Column("xml", XML), Column("xml", XML), Column("xml", XML)],
                        rows: &[
                            &[T("x"), T(""), T("<a/><b/>"), T("<!-- c --><a/>"), T("<a><![CDATA[<x>]]></a>"), T(r#"<a xmlns:p="urn:x"><p:b/></a>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '<?xml version="1.0"?><a/>'::xml, '<?xml version="1.0" encoding="UTF-8"?><a/>'::xml, '<?xml version="1.0" standalone="yes"?><a/>'::xml, '<?xml version="1.1"?><a/>'::xml;"#,
                    expected: Expected::Rows {
                        columns: &[Column("xml", XML), Column("xml", XML), Column("xml", XML), Column("xml", XML)],
                        rows: &[
                            &[T("<a/>"), T("<a/>"), T(r#"<?xml version="1.0" standalone="yes"?><a/>"#), T(r#"<?xml version="1.1"?><a/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT E'<?xml version="1.0"?>\n<a/>'::xml, E'<?xml version="1.0"?>\n\n<a/>'::xml, E'<a>\n</a>'::xml;"#,
                    expected: Expected::Rows {
                        columns: &[Column("xml", XML), Column("xml", XML), Column("xml", XML)],
                        rows: &[
                            &[T("<a/>"), T(r#"
<a/>"#), T(r#"<a>
</a>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Premature end of data in tag a line 1
<a>
   ^"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'x<'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: StartTag: invalid element name
x<
  ^"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a>&foo;</a>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Entity 'foo' not defined
<a>&foo;</a>
        ^"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '<a>x</a>'::text::xml, '<a>x</a>'::xml::text, '<a>x</a>'::xml::varchar, '<a>x</a>'::xml::char(5), '<a>x</a>'::varchar::xml, '<?xml version="1.0"?><a/>'::xml::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("xml", XML), Column("text", TEXT), Column("varchar", VARCHAR), Column("bpchar", BPCHAR), Column("xml", XML), Column("text", TEXT)],
                        rows: &[
                            &[T("<a>x</a>"), T("<a>x</a>"), T("<a>x</a>"), T("<a>x<"), T("<a>x</a>"), T(r#"<?xml version="1.0"?><a/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a>'::text::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Premature end of data in tag a line 1
<a>
   ^"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a/>'::xml::int;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type xml to integer", position: 19, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1::xml;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type integer to xml", position: 9, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a/>'::xml = '<a/>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "operator does not exist: xml = xml", hint: "No operator matches the given name and argument types. You might need to add explicit type casts.", position: 20, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Xml document option",
            set_up_script: &[
                "SET xmloption TO document;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '<a/>'::xml, '<?xml version="1.0"?><a/>'::xml, '<!-- c --><a/>'::xml, ' <a/>'::xml;"#,
                    expected: Expected::Rows {
                        columns: &[Column("xml", XML), Column("xml", XML), Column("xml", XML), Column("xml", XML)],
                        rows: &[
                            &[T("<a/>"), T("<a/>"), T("<!-- c --><a/>"), T(" <a/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'x'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "invalid XML document", detail: r#"line 1: Start tag expected, '<' not found
x
^"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a/><b/>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "invalid XML document", detail: r#"line 1: Extra content at the end of the document
<a/><b/>
    ^"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ''::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "invalid XML document", detail: r#"line 1: switching encoding : no input

^
line 1: Document is empty

^"#, position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Xml column default",
            set_up_script: &[
                "CREATE TABLE t_xml (id INTEGER PRIMARY KEY, v1 XML DEFAULT '<d/>'::xml);",
                "INSERT INTO t_xml VALUES (1, '<a>x</a>');",
                "INSERT INTO t_xml (id) VALUES (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_xml ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", XML)],
                        rows: &[
                            &[T("1"), T("<a>x</a>")],
                            &[T("2"), T("<d/>")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Xml array type",
            set_up_script: &[
                "CREATE TABLE t_xml (id INTEGER PRIMARY KEY, v1 XML[]);",
                "INSERT INTO t_xml VALUES (1, ARRAY['<a/>'::xml, '<b>x y</b>']), (2, '{<c/>,NULL}'), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_xml ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", XML_ARRAY)],
                        rows: &[
                            &[T("1"), T(r#"{<a/>,"<b>x y</b>"}"#)],
                            &[T("2"), T("{<c/>,NULL}")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v1[2] FROM t_xml ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", XML)],
                        rows: &[
                            &[T("1"), T("<b>x y</b>")],
                            &[T("2"), Null],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_xml VALUES (4, '{<a>}');",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Premature end of data in tag a line 1
<a>
   ^"#, position: 30, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ARRAY['<a>x</a>'::xml, '<b c="1">y z</b>', 'q,"r"'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("array", XML_ARRAY)],
                        rows: &[
                            &[T(r#"{<a>x</a>,"<b c=\"1\">y z</b>","q,\"r\""}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Xml schema-qualified type",
            set_up_script: &[
                "CREATE TABLE t3337 (id INT PRIMARY KEY, doc pg_catalog.xml, docs xml[]);",
                "INSERT INTO t3337 VALUES (1, '<a>x</a>', ARRAY['<b/>'::xml]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '<a>x</a>'::pg_catalog.xml, pg_typeof('<a>x</a>'::pg_catalog.xml), pg_typeof(ARRAY['<a/>'::xml]);",
                    expected: Expected::Rows {
                        columns: &[Column("xml", XML), Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("<a>x</a>"), T("xml"), T("xml[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, doc, docs, pg_typeof(doc), pg_typeof(docs) FROM t3337;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("doc", XML), Column("docs", XML_ARRAY), Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("1"), T("<a>x</a>"), T("{<b/>}"), T("xml"), T("xml[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Regprocedure type",
            set_up_script: &[
                "CREATE FUNCTION tf() RETURNS trigger AS $$ BEGIN RETURN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f2(a INT, b TEXT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3(INT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3(TEXT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE SCHEMA s;",
                "CREATE FUNCTION s.sf(INT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE p1(INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'tf()'::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("tf()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 's.sf(int)'::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("s.sf(integer)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'f2(int,text)'::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("f2(integer,text)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abs(int)'::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("abs(integer)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'p1(int)'::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("p1(integer)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'tf()'::regprocedure::oid = (SELECT oid FROM pg_proc WHERE proname = 'tf');",
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
                    query: "SELECT 2212::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("regprocedurein(cstring)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'now'::regproc::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("now()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'f2(int,text)'::regprocedure::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("f2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'f2(int,text)'::regprocedure::text;",
                    expected: Expected::Rows {
                        columns: &[Column("text", TEXT)],
                        rows: &[
                            &[T("f2(integer,text)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'f3(int)'::regprocedure::regproc::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("f3(integer)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'f3(bool)'::regprocedure;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: r#"function "f3(bool)" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'nosuch()'::regprocedure;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: r#"function "nosuch()" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'f3'::regprocedure;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "expected a left parenthesis", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'nosuchschema.sf(int)'::regprocedure;",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "nosuchschema" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = s;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 's.sf(int)'::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("sf(integer)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'public.f2(int,text)'::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("public.f2(integer,text)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'array_agg(anynonarray)'::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("array_agg(anynonarray)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'row_number()'::regprocedure;",
                    expected: Expected::Rows {
                        columns: &[Column("regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("row_number()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Regproc user-defined routines",
            set_up_script: &[
                "CREATE FUNCTION tf() RETURNS trigger AS $$ BEGIN RETURN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f2(a INT, b TEXT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3(INT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3(TEXT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE SCHEMA s;",
                "CREATE FUNCTION s.sf(INT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE p1(INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'tf'::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("tf")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regproc('tf');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[T("tf")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regproc('public.tf');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[T("tf")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 's.sf'::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("s.sf")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regproc('s.sf');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[T("s.sf")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regproc('sf');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regproc('p1');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[T("p1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regproc('f3');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regproc('nosuchschema.sf');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'tf'::regproc::oid = (SELECT oid FROM pg_proc WHERE proname = 'tf');",
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
                    query: "SELECT 'abs(int)'::regprocedure::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("pg_catalog.abs")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'f3(int)'::regprocedure::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("public.f3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'f3'::regproc;",
                    expected: Expected::Error(Diagnostic { code: "42725", message: r#"more than one function named "f3""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abs'::regproc;",
                    expected: Expected::Error(Diagnostic { code: "42725", message: r#"more than one function named "abs""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'nosuch'::regproc;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: r#"function "nosuch" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'nosuchschema.sf'::regproc;",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "nosuchschema" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = s;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'sf'::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("sf")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'public.tf'::regproc;",
                    expected: Expected::Rows {
                        columns: &[Column("regproc", REGPROC)],
                        rows: &[
                            &[T("public.tf")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regproc('row_number');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[T("row_number")],
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
fn test_user_type_rules() {
    run_scripts(&[
        ScriptTest {
            name: "enum types",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy');",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE mood AS ENUM ('x');",
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"type "mood" already exists"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT PRIMARY KEY, m mood, ms mood[]);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, 'ok', ARRAY['sad', 'happy']::mood[]), (2, 'happy', NULL), (3, 'sad', '{}');",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (4, 'nope', NULL);",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input value for enum mood: "nope""#, position: 26, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, m, ms FROM t ORDER BY m;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("m", USER_DEFINED), Column("ms", USER_DEFINED)],
                        rows: &[
                            &[T("3"), T("sad"), T("{}")],
                            &[T("1"), T("ok"), T("{sad,happy}")],
                            &[T("2"), T("happy"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT m, m::text, m = 'ok', m < 'happy' FROM t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("m", USER_DEFINED), Column("m", TEXT), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("ok"), T("ok"), T("t"), T("t")],
                            &[T("happy"), T("happy"), T("f"), T("f")],
                            &[T("sad"), T("sad"), T("f"), T("t")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'happy'::mood > 'sad'::mood, enum_first(NULL::mood), enum_last(NULL::mood), enum_range(NULL::mood);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("enum_first", USER_DEFINED), Column("enum_last", USER_DEFINED), Column("enum_range", USER_DEFINED)],
                        rows: &[
                            &[T("t"), T("sad"), T("happy"), T("{sad,ok,happy}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(m), pg_typeof(ms) FROM t WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("mood"), T("mood[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TYPE mood ADD VALUE 'meh' BEFORE 'ok';",
                    expected: Expected::Tag("ALTER TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TYPE mood ADD VALUE 'ok';",
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"enum label "ok" already exists"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TYPE mood ADD VALUE IF NOT EXISTS 'ok';",
                    expected: Expected::Tag("ALTER TYPE"),
                    notices: &[Diagnostic { code: "42710", message: r#"enum label "ok" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TYPE mood ADD VALUE 'great' AFTER 'nope';",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#""nope" is not an existing enum label"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TYPE mood RENAME VALUE 'meh' TO 'fine';",
                    expected: Expected::Tag("ALTER TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT enum_range(NULL::mood), 'fine'::mood < 'ok'::mood;",
                    expected: Expected::Rows {
                        columns: &[Column("enum_range", USER_DEFINED), Column("?column?", BOOL)],
                        rows: &[
                            &[T("{sad,fine,ok,happy}"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT enumlabel, enumsortorder FROM pg_enum e JOIN pg_type t ON t.oid = e.enumtypid WHERE t.typname = 'mood' ORDER BY enumsortorder;",
                    expected: Expected::Rows {
                        columns: &[Column("enumlabel", NAME), Column("enumsortorder", FLOAT4)],
                        rows: &[
                            &[T("sad"), T("1")],
                            &[T("fine"), T("1.5")],
                            &[T("ok"), T("2")],
                            &[T("happy"), T("3")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE mood;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop type mood because other objects depend on it", detail: r#"column ms of table t depends on type mood[]
column m of table t depends on type mood"#, hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE mood;",
                    expected: Expected::Tag("DROP TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE IF EXISTS mood;",
                    expected: Expected::Tag("DROP TYPE"),
                    notices: &[Diagnostic { code: "00000", message: r#"type "mood" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "composite types",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TYPE pair AS (x INT, y TEXT);",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE pair AS (a INT);",
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"type "pair" already exists"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE bad AS (x INT, y nope);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "nope" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE pair (a INT);",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "pair" already exists"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT PRIMARY KEY, p pair);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO t VALUES (1, ROW(1, 'a')), (2, '(2,"b c")'), (3, NULL);"#,
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (4, ROW(1));",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type record to pair", detail: "Input has too few columns.", position: 26, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, p, (p).x, (p).y FROM t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("p", USER_DEFINED), Column("x", INT4), Column("y", TEXT)],
                        rows: &[
                            &[T("1"), T("(1,a)"), T("1"), T("a")],
                            &[T("2"), T(r#"(2,"b c")"#), T("2"), T("b c")],
                            &[T("3"), Null, Null, Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('(5,z)'::pair).y, ROW(7, 'w')::pair, pg_typeof(ROW(7, 'w')::pair);",
                    expected: Expected::Rows {
                        columns: &[Column("y", TEXT), Column("row", USER_DEFINED), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("z"), T("(7,w)"), T("pair")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '(1)'::pair;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed record literal: "(1)""#, detail: "Too few columns.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1,2'::pair;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed record literal: "1,2""#, detail: "Missing left parenthesis.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (p).nope FROM t;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "nope" not found in data type pair"#, position: 9, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_jsonb(p), row_to_json(p) FROM t WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("to_jsonb", JSONB), Column("row_to_json", JSON)],
                        rows: &[
                            &[T(r#"{"x": 1, "y": "a"}"#), T(r#"{"x":1,"y":"a"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT typname, typtype, typcategory, typlen FROM pg_type WHERE typname IN ('pair', '_pair') ORDER BY typname;",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME), Column("typtype", CHAR), Column("typcategory", CHAR), Column("typlen", INT2)],
                        rows: &[
                            &[T("_pair"), T("b"), T("A"), T("-1")],
                            &[T("pair"), T("c"), T("C"), T("-1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT relname, relkind FROM pg_class WHERE relname = 'pair';",
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME), Column("relkind", CHAR)],
                        rows: &[
                            &[T("pair"), T("c")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE pair;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop type pair because other objects depend on it", detail: "column p of table t depends on type pair", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TYPE pair;",
                    expected: Expected::Tag("DROP TYPE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "table row types and whole-row references",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE people (id INT PRIMARY KEY, name TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO people VALUES (1, 'ann'), (2, 'bob');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE holders (id INT PRIMARY KEY, who people);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO holders VALUES (1, ROW(9, 'zed'));",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, who, (who).name FROM holders;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("who", USER_DEFINED), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("(9,zed)"), T("zed")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT p, to_jsonb(p), row(p.*, 42) FROM people p ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("p", USER_DEFINED), Column("to_jsonb", JSONB), Column("row", RECORD)],
                        rows: &[
                            &[T("(1,ann)"), T(r#"{"id": 1, "name": "ann"}"#), T("(1,ann,42)")],
                            &[T("(2,bob)"), T(r#"{"id": 2, "name": "bob"}"#), T("(2,bob,42)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(p) FROM people p WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("people")],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("whole-row references carry their FROM item's alias rather than the table's row type"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "domains",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DOMAIN posint AS INT CHECK (VALUE > 0);",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN d1 AS INT NULL NOT NULL;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "conflicting NULL/NOT NULL constraints", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN d1 AS INT DEFAULT 1 DEFAULT 2;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "multiple default expressions", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN d1 AS INT CHECK (VALUE + 1);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of CHECK must be type boolean, not type integer", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN d1 AS INT CHECK (x > 1);",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "x" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN d1 AS nope;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "nope" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN d1 AS RECORD;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#""record" is not a valid base type for a domain"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN bounded AS INT CONSTRAINT pos CHECK (VALUE > 0) CHECK (VALUE < 100) NOT NULL DEFAULT 7;",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 0::bounded;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain bounded violates check constraint "pos""#, schema: "public", data_type: "bounded", constraint: "pos", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 5::bounded, 5::bounded + 1, pg_typeof(5::bounded), pg_typeof(5::bounded + 1);",
                    expected: Expected::Rows {
                        columns: &[Column("bounded", INT4), Column("?column?", INT4), Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("5"), T("6"), T("bounded"), T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL::bounded;",
                    expected: Expected::Error(Diagnostic { code: "23502", message: "domain bounded does not allow null values", schema: "public", data_type: "bounded", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE td (id INT PRIMARY KEY, v bounded, p posint);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO td (id) VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO td VALUES (2, NULL, 1);",
                    expected: Expected::Error(Diagnostic { code: "23502", message: "domain bounded does not allow null values", schema: "public", data_type: "bounded", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO td VALUES (3, 500, 1);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain bounded violates check constraint "bounded_check""#, schema: "public", data_type: "bounded", constraint: "bounded_check", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO td VALUES (4, 5, -1);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain posint violates check constraint "posint_check""#, schema: "public", data_type: "posint", constraint: "posint_check", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO td VALUES (5, 50, 2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE td SET v = v * 100;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain bounded violates check constraint "bounded_check""#, schema: "public", data_type: "bounded", constraint: "bounded_check", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM td ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4), Column("p", INT4)],
                        rows: &[
                            &[T("1"), T("7"), Null],
                            &[T("5"), T("50"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT typname, typtype, typbasetype::regtype, typnotnull, typdefault FROM pg_type WHERE typname IN ('bounded', 'posint') ORDER BY typname;",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME), Column("typtype", CHAR), Column("typbasetype", REGTYPE), Column("typnotnull", BOOL), Column("typdefault", TEXT)],
                        rows: &[
                            &[T("bounded"), T("d"), T("integer"), T("t"), T("7")],
                            &[T("posint"), T("d"), T("integer"), T("f"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN bounded;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop type bounded because other objects depend on it", detail: "column v of table td depends on type bounded", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN IF EXISTS nope;",
                    expected: Expected::Tag("DROP DOMAIN"),
                    notices: &[Diagnostic { code: "00000", message: r#"type "nope" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE not_domain AS ENUM ('a');",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN not_domain;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#""not_domain" is not a domain"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE td;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN bounded, posint;",
                    expected: Expected::Tag("DROP DOMAIN"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pseudo-types are not column types",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t (pk INT PRIMARY KEY, r RECORD);",
                    expected: Expected::Error(Diagnostic { code: "42P16", message: r#"column "r" has pseudo-type record"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE outer_type AS (id INT, payload RECORD);",
                    expected: Expected::Error(Diagnostic { code: "42P16", message: r#"column "payload" has pseudo-type record"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_binary_type_rules() {
    run_scripts(&[
        ScriptTest {
            name: "bit strings",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '101'::bit, '101'::bit(3), B'101', X'1F', '101'::varbit, '101'::bit varying(2), '10'::bit(3);",
                    expected: Expected::Rows {
                        columns: &[Column("bit", BIT), Column("bit", BIT), Column("?column?", BIT), Column("?column?", BIT), Column("varbit", VARBIT), Column("varbit", VARBIT), Column("bit", BIT)],
                        rows: &[
                            &[T("1"), T("101"), T("101"), T("00011111"), T("101"), T("10"), T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1012'::bit(4);",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#""2" is not a valid binary digit"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'X1G'::bit(8);",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#""G" is not a valid hexadecimal digit"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT B'10012';",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#""2" is not a valid binary digit"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE tb (a BIT(3), b VARBIT(2));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tb VALUES ('10', '1');",
                    expected: Expected::Error(Diagnostic { code: "22026", message: "bit string length 2 does not match type bit(3)", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tb VALUES ('101', '101');",
                    expected: Expected::Error(Diagnostic { code: "22001", message: "bit string too long for type bit varying(2)", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tb VALUES (B'101', B'10');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tb;",
                    expected: Expected::Rows {
                        columns: &[Column("a", BIT), Column("b", VARBIT)],
                        rows: &[
                            &[T("101"), T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 5::bit(4), B'1010'::int, B'101' & B'110', B'101' | B'110', B'101' # B'110', ~B'101', B'1011' << 1, B'1011' >> 2, B'10' || B'01';",
                    expected: Expected::Rows {
                        columns: &[Column("bit", BIT), Column("int4", INT4), Column("?column?", BIT), Column("?column?", BIT), Column("?column?", BIT), Column("?column?", BIT), Column("?column?", BIT), Column("?column?", BIT), Column("?column?", VARBIT)],
                        rows: &[
                            &[T("0101"), T("10"), T("100"), T("111"), T("011"), T("010"), T("0110"), T("0010"), T("1001")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(B'1011'), bit_length(B'1011'), octet_length(B'101010101'), get_bit(B'1011', 1), set_bit(B'1011', 1, 0), substring(B'101101' FROM 2 FOR 3), position(B'01' IN B'1101');",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("bit_length", INT4), Column("octet_length", INT4), Column("get_bit", INT4), Column("set_bit", BIT), Column("substring", BIT), Column("position", INT4)],
                        rows: &[
                            &[T("4"), T("4"), T("2"), T("0"), T("1011"), T("011"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT B'101' & B'1100';",
                    expected: Expected::Error(Diagnostic { code: "22026", message: "cannot AND bit strings of different sizes", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT get_bit(B'101', 3);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "bit index 3 out of valid range (0..2)", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_bit(B'101', 1, 2);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "new bit must be 0 or 1", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT B'1011' << -1, B'1011' >> 5, B'1011' << 0, substring(B'1011' FROM 3), substring(B'1011', 0, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BIT), Column("?column?", BIT), Column("?column?", BIT), Column("substring", BIT), Column("substring", BIT)],
                        rows: &[
                            &[T("0101"), T("0000"), T("1011"), T("11"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT B'101' = '101'::varbit, pg_typeof(B'10' || B'1'), B'101' < B'11', B'1' || '01';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("pg_typeof", REGTYPE), Column("?column?", BOOL), Column("?column?", VARBIT)],
                        rows: &[
                            &[T("t"), T("bit varying"), T("t"), T("101")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT B'101'::varbit(2), B'101'::bit(2), B'101'::bit(5), B'1'::bit(5)::int;",
                    expected: Expected::Rows {
                        columns: &[Column("varbit", VARBIT), Column("bit", BIT), Column("bit", BIT), Column("int4", INT4)],
                        rows: &[
                            &[T("10"), T("10"), T("10100"), T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 300::bit(4), (-1)::bit(4), 5::bigint::bit(70), B'111'::int8;",
                    expected: Expected::Rows {
                        columns: &[Column("bit", BIT), Column("bit", BIT), Column("bit", BIT), Column("int8", INT8)],
                        rows: &[
                            &[T("1100"), T("1111"), T("0000000000000000000000000000000000000000000000000000000000000000000101"), T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "integer bitwise operators",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 5 & 3, 5 | 3, 5 # 3, ~5, 1 << 4, 256 >> 2, 5::int2 & 3::int2, pg_typeof(5::int2 & 3), 5::int8 << 62, 1 << 33, 1::int8 << 65, -16 >> 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4), Column("?column?", INT4), Column("?column?", INT4), Column("?column?", INT4), Column("?column?", INT4), Column("?column?", INT4), Column("?column?", INT2), Column("pg_typeof", REGTYPE), Column("?column?", INT8), Column("?column?", INT4), Column("?column?", INT8), Column("?column?", INT4)],
                        rows: &[
                            &[T("1"), T("7"), T("6"), T("-6"), T("16"), T("64"), T("1"), T("integer"), T("4611686018427387904"), T("2"), T("2"), T("-4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "bytea",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '\x0102ff'::bytea, 'abc\\x'::bytea, E'\\001abc'::bytea, '\x 01 02'::bytea;"#,
                    expected: Expected::Rows {
                        columns: &[Column("bytea", BYTEA), Column("bytea", BYTEA), Column("bytea", BYTEA), Column("bytea", BYTEA)],
                        rows: &[
                            &[T(r#"\x0102ff"#), T(r#"\x6162635c78"#), T(r#"\x01616263"#), T(r#"\x0102"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '\x012'::bytea;"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: "invalid hexadecimal data: odd number of digits", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '\x0g'::bytea;"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid hexadecimal digit: "g""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 'a\b'::bytea;"#,
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "invalid input syntax for type bytea", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '\xdeadbeef'::bytea || '\x01'::bytea, length('\xdead'::bytea), octet_length('\xdead'::bytea), get_byte('\xdead'::bytea, 1), set_byte('\xdead'::bytea, 0, 1), substring('\xdeadbeef'::bytea FROM 2 FOR 2), position('\xbe'::bytea IN '\xdeadbeef'::bytea);"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BYTEA), Column("length", INT4), Column("octet_length", INT4), Column("get_byte", INT4), Column("set_byte", BYTEA), Column("substring", BYTEA), Column("position", INT4)],
                        rows: &[
                            &[T(r#"\xdeadbeef01"#), T("2"), T("2"), T("173"), T(r#"\x01ad"#), T(r#"\xadbe"#), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT get_byte('\xdead'::bytea, 2);"#,
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "index 2 out of valid range, 0..1", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT get_bit('\x80'::bytea, 7), set_bit('\x00'::bytea, 7, 1), substring('\x0102030405'::bytea FROM -1 FOR 3), substr('\x0102'::bytea, 2);"#,
                    expected: Expected::Rows {
                        columns: &[Column("get_bit", INT4), Column("set_bit", BYTEA), Column("substring", BYTEA), Column("substr", BYTEA)],
                        rows: &[
                            &[T("1"), T(r#"\x80"#), T(r#"\x01"#), T(r#"\x02"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT get_bit('\x80'::bytea, 8);"#,
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "index 8 out of valid range, 0..7", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT encode('\xdeadbeef'::bytea, 'hex'), encode('abc'::bytea, 'base64'), encode('\x00ff5c41'::bytea, 'escape'), decode('3q2+7w==', 'base64'), decode('deadbeef', 'hex'), decode('a\001', 'escape');"#,
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT), Column("decode", BYTEA), Column("decode", BYTEA), Column("decode", BYTEA)],
                        rows: &[
                            &[T("deadbeef"), T("YWJj"), T(r#"\000\377\\A"#), T(r#"\xdeadbeef"#), T(r#"\xdeadbeef"#), T(r#"\x6101"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(repeat('x', 60)::bytea, 'base64');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T(r#"eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4
eHh4"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT decode('!!', 'base64');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid symbol "!" found while decoding base64 sequence"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT decode('ab', 'foo');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized encoding: "foo""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT decode('a\x', 'escape');"#,
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "invalid input syntax for type bytea", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT md5('abc'::bytea), sha256('abc'::bytea), sha224('a'::bytea), convert_from('\x616263'::bytea, 'UTF8'), convert_to('abc', 'UTF8');"#,
                    expected: Expected::Rows {
                        columns: &[Column("md5", TEXT), Column("sha256", BYTEA), Column("sha224", BYTEA), Column("convert_from", TEXT), Column("convert_to", BYTEA)],
                        rows: &[
                            &[T("900150983cd24fb0d6963f7d28e17f72"), T(r#"\xba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"#), T(r#"\xabd37534c7d9a2efb9465de931cd7055ffdb8879563ae98078d6d6d5"#), T("abc"), T(r#"\x616263"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\xff'::bytea, 'UTF8');"#,
                    expected: Expected::Error(Diagnostic { code: "22021", message: r#"invalid byte sequence for encoding "UTF8": 0xff"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '\xff'::bytea > '\x01'::bytea, '\x01'::bytea = '\x01'::bytea, 'abc'::bytea::text, 'abc'::text::bytea, '\x01'::bytea || 'ab', pg_typeof('\x01'::bytea || 'ab');"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("text", TEXT), Column("bytea", BYTEA), Column("?column?", BYTEA), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("t"), T("t"), T(r#"\x616263"#), T(r#"\x616263"#), T(r#"\x016162"#), T("bytea")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT length('abc'), length('\x01'::bytea), bit_length('\x0102'::bytea), btrim('\x0001020100'::bytea, '\x00'::bytea);"#,
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("length", INT4), Column("bit_length", INT4), Column("btrim", BYTEA)],
                        rows: &[
                            &[T("3"), T("1"), T("16"), T(r#"\x010201"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET bytea_output = 'escape';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT E'a\\001\\377b\\\\'::bytea;"#,
                    expected: Expected::Rows {
                        columns: &[Column("bytea", BYTEA)],
                        rows: &[
                            &[T(r#"\x6101ff625c"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET bytea_output;",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc'::bytea LIKE 'a%'::bytea;",
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
            name: "uuid",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'A0EEBC99-9C0B-4EF8-BB6D-6BB9BD380A11'::uuid, '{a0eebc999c0b4ef8bb6d6bb9bd380a11}'::uuid, 'a0ee-bc99-9c0b-4ef8-bb6d-6bb9-bd38-0a11'::uuid;",
                    expected: Expected::Rows {
                        columns: &[Column("uuid", UUID), Column("uuid", UUID), Column("uuid", UUID)],
                        rows: &[
                            &[T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a1'::uuid;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type uuid: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a1""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a0eebc99-9c0b4-ef8-bb6d-6bb9bd380a11'::uuid;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type uuid: "a0eebc99-9c0b4-ef8-bb6d-6bb9bd380a11""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid < 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid, 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid::text;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("text", TEXT)],
                        rows: &[
                            &[T("t"), T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(gen_random_uuid()), length(gen_random_uuid()::text);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("length", INT4)],
                        rows: &[
                            &[T("uuid"), T("36")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc'::bytea::uuid;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type bytea to uuid", position: 20, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "stored bytea, uuid, and bit values",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE bt (id INT PRIMARY KEY, b BYTEA, u UUID, x BIT(4), v VARBIT(8), ba BYTEA[], ua UUID[], xa BIT(2)[]);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO bt VALUES (1, '\xdeadbeef', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', B'1010', B'101', ARRAY['\x01'::bytea, '\x'], ARRAY['a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid], ARRAY[B'10', B'01']), (2, '', '00000000-0000-0000-0000-000000000000', '0000', '', NULL, NULL, NULL);"#,
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM bt ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("b", BYTEA), Column("u", UUID), Column("x", BIT), Column("v", VARBIT), Column("ba", BYTEA_ARRAY), Column("ua", UUID_ARRAY), Column("xa", BIT_ARRAY)],
                        rows: &[
                            &[T("1"), T(r#"\xdeadbeef"#), T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), T("1010"), T("101"), T(r#"{"\\x01","\\x"}"#), T("{a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11}"), T("{10,01}")],
                            &[T("2"), T(r#"\x"#), T("00000000-0000-0000-0000-000000000000"), T("0000"), T(""), Null, Null, Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE ku (u UUID PRIMARY KEY, b BYTEA UNIQUE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO ku VALUES ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '\x02'), ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '\x01');"#,
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO ku VALUES ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '\x03');"#,
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "ku_pkey""#, detail: "Key (u)=(b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11) already exists.", schema: "public", table: "ku", constraint: "ku_pkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM ku ORDER BY u;",
                    expected: Expected::Rows {
                        columns: &[Column("u", UUID), Column("b", BYTEA)],
                        rows: &[
                            &[T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), T(r#"\x01"#)],
                            &[T("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), T(r#"\x02"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM ku WHERE u = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';",
                    expected: Expected::Rows {
                        columns: &[Column("u", UUID), Column("b", BYTEA)],
                        rows: &[
                            &[T("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), T(r#"\x01"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM ku WHERE b > '\x01' ORDER BY b DESC;"#,
                    expected: Expected::Rows {
                        columns: &[Column("u", UUID), Column("b", BYTEA)],
                        rows: &[
                            &[T("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), T(r#"\x02"#)],
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
fn test_builtin_base_types() {
    run_scripts(&[
        ScriptTest {
            name: "geometric, network, money, and pg_lsn input and output",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '(1,2)'::point, ' ( 1.5 , -2e3 ) '::point, '1,2'::point;",
                    expected: Expected::Rows {
                        columns: &[Column("point", POINT), Column("point", POINT), Column("point", POINT)],
                        rows: &[
                            &[T("(1,2)"), T("(1.5,-2000)"), T("(1,2)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '((1,2),(3,4))'::lseg, '[(1,2),(3,4)]'::lseg, '1,2,3,4'::lseg;",
                    expected: Expected::Rows {
                        columns: &[Column("lseg", LSEG), Column("lseg", LSEG), Column("lseg", LSEG)],
                        rows: &[
                            &[T("[(1,2),(3,4)]"), T("[(1,2),(3,4)]"), T("[(1,2),(3,4)]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '(1,2),(3,4)'::box, '((5,6),(7,8))'::box, '(3,1),(1,3)'::box;",
                    expected: Expected::Rows {
                        columns: &[Column("box", BOX), Column("box", BOX), Column("box", BOX)],
                        rows: &[
                            &[T("(3,4),(1,2)"), T("(7,8),(5,6)"), T("(3,3),(1,1)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '((1,2),(3,4),(5,6))'::path, '[(1,2),(3,4)]'::path, '(1,2),(3,4)'::path;",
                    expected: Expected::Rows {
                        columns: &[Column("path", PATH), Column("path", PATH), Column("path", PATH)],
                        rows: &[
                            &[T("((1,2),(3,4),(5,6))"), T("[(1,2),(3,4)]"), T("((1,2),(3,4))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '((1,2),(3,4),(5,6))'::polygon, '(1,2),(3,4)'::polygon;",
                    expected: Expected::Rows {
                        columns: &[Column("polygon", POLYGON), Column("polygon", POLYGON)],
                        rows: &[
                            &[T("((1,2),(3,4),(5,6))"), T("((1,2),(3,4))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::line, '[(0,0),(1,1)]'::line, '((0,0),(0,1))'::line, '((0,5),(1,5))'::line;",
                    expected: Expected::Rows {
                        columns: &[Column("line", LINE), Column("line", LINE), Column("line", LINE), Column("line", LINE)],
                        rows: &[
                            &[T("{1,2,3}"), T("{1,-1,0}"), T("{-1,0,0}"), T("{0,-1,5}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<(1,2),3>'::circle, '((1,2),3)'::circle, '(1,2),3'::circle, '1,2,3'::circle;",
                    expected: Expected::Rows {
                        columns: &[Column("circle", CIRCLE), Column("circle", CIRCLE), Column("circle", CIRCLE), Column("circle", CIRCLE)],
                        rows: &[
                            &[T("<(1,2),3>"), T("<(1,2),3>"), T("<(1,2),3>"), T("<(1,2),3>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '192.168.1.1'::inet, '10.0.0.0/8'::inet, '::1'::inet, '2001:db8::1/64'::inet, '::ffff:1.2.3.4'::inet, 'fe80::1:0:0:1'::inet;",
                    expected: Expected::Rows {
                        columns: &[Column("inet", INET), Column("inet", INET), Column("inet", INET), Column("inet", INET), Column("inet", INET), Column("inet", INET)],
                        rows: &[
                            &[T("192.168.1.1"), T("10.0.0.0/8"), T("::1"), T("2001:db8::1/64"), T("::ffff:1.2.3.4"), T("fe80::1:0:0:1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '192.168.1.0/24'::cidr, '10/8'::cidr, '10.1'::cidr, '192.168'::cidr, '2001:db8::/32'::cidr;",
                    expected: Expected::Rows {
                        columns: &[Column("cidr", CIDR), Column("cidr", CIDR), Column("cidr", CIDR), Column("cidr", CIDR), Column("cidr", CIDR)],
                        rows: &[
                            &[T("192.168.1.0/24"), T("10.0.0.0/8"), T("10.1.0.0/16"), T("192.168.0.0/24"), T("2001:db8::/32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '08:00:2b:01:02:03'::macaddr, '08-00-2b-01-02-03'::macaddr, '08002b:010203'::macaddr, '0800.2b01.0203'::macaddr, '08002b010203'::macaddr;",
                    expected: Expected::Rows {
                        columns: &[Column("macaddr", MACADDR), Column("macaddr", MACADDR), Column("macaddr", MACADDR), Column("macaddr", MACADDR), Column("macaddr", MACADDR)],
                        rows: &[
                            &[T("08:00:2b:01:02:03"), T("08:00:2b:01:02:03"), T("08:00:2b:01:02:03"), T("08:00:2b:01:02:03"), T("08:00:2b:01:02:03")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '$100.25'::money, '1234.567'::money, '-$1,234.50'::money, '($5)'::money, '0.01'::money, '12'::money;",
                    expected: Expected::Rows {
                        columns: &[Column("money", MONEY), Column("money", MONEY), Column("money", MONEY), Column("money", MONEY), Column("money", MONEY), Column("money", MONEY)],
                        rows: &[
                            &[T("$100.25"), T("$1,234.57"), T("-$1,234.50"), T("-$5.00"), T("$0.01"), T("$12.00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '16/B8E36C60'::pg_lsn, '0/0'::pg_lsn, 'FFFFFFFF/FFFFFFFF'::pg_lsn;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_lsn", PG_LSN), Column("pg_lsn", PG_LSN), Column("pg_lsn", PG_LSN)],
                        rows: &[
                            &[T("16/B8E36C60"), T("0/0"), T("FFFFFFFF/FFFFFFFF")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'x'::point;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type point: "x""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{0,0,1}'::line;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "invalid line specification: A and B cannot both be zero", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '192.168.1.1/24'::cidr;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid cidr value: "192.168.1.1/24""#, detail: "Value has bits set to right of mask.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1.2.3'::inet;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type inet: "1.2.3""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'zz:00:2b:01:02:03'::macaddr;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type macaddr: "zz:00:2b:01:02:03""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc'::money;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type money: "abc""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '16/'::pg_lsn;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type pg_lsn: "16/""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<(1,2),-3>'::circle;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type circle: "<(1,2),-3>""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE gt (id INT PRIMARY KEY, p POINT, b BOX, i INET, m MONEY, l PG_LSN, c CIDR, mac MACADDR, pa PATH, po POLYGON, ci CIRCLE, li LINE, ls LSEG);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO gt VALUES (1, '(1,2)', '(1,2),(3,4)', '10.0.0.1', '$5.25', '1/2', '10.0.0.0/8', '08:00:2b:01:02:03', '[(1,2),(3,4)]', '((1,2),(3,4),(5,6))', '<(1,2),3>', '{1,2,3}', '[(1,2),(3,4)]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO gt VALUES (2, '(5,6)', '(0,0),(1,1)', '::1', '-$1.00', 'A/B', '2001:db8::/32', '00-11-22-33-44-55', '((0,0),(1,1))', '((0,0),(1,0),(0,1))', '<(0,0),1>', '{0,1,2}', '[(0,0),(0,1)]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM gt ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("p", POINT), Column("b", BOX), Column("i", INET), Column("m", MONEY), Column("l", PG_LSN), Column("c", CIDR), Column("mac", MACADDR), Column("pa", PATH), Column("po", POLYGON), Column("ci", CIRCLE), Column("li", LINE), Column("ls", LSEG)],
                        rows: &[
                            &[T("1"), T("(1,2)"), T("(3,4),(1,2)"), T("10.0.0.1"), T("$5.25"), T("1/2"), T("10.0.0.0/8"), T("08:00:2b:01:02:03"), T("[(1,2),(3,4)]"), T("((1,2),(3,4),(5,6))"), T("<(1,2),3>"), T("{1,2,3}"), T("[(1,2),(3,4)]")],
                            &[T("2"), T("(5,6)"), T("(1,1),(0,0)"), T("::1"), T("-$1.00"), T("A/B"), T("2001:db8::/32"), T("00:11:22:33:44:55"), T("((0,0),(1,1))"), T("((0,0),(1,0),(0,1))"), T("<(0,0),1>"), T("{0,1,2}"), T("[(0,0),(0,1)]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(p), pg_typeof(i), pg_typeof(m) FROM gt ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("point"), T("inet"), T("money")],
                            &[T("point"), T("inet"), T("money")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM gt WHERE i = '10.0.0.1';",
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
                    query: "SELECT m FROM gt ORDER BY m;",
                    expected: Expected::Rows {
                        columns: &[Column("m", MONEY)],
                        rows: &[
                            &[T("-$1.00")],
                            &[T("$5.25")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT i FROM gt ORDER BY i;",
                    expected: Expected::Rows {
                        columns: &[Column("i", INET)],
                        rows: &[
                            &[T("10.0.0.1")],
                            &[T("::1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l FROM gt ORDER BY l DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("l", PG_LSN)],
                        rows: &[
                            &[T("A/B")],
                            &[T("1/2")],
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
fn test_input_edge_cases() {
    run_scripts(&[
        ScriptTest {
            name: "datetime, reg type, and name input, polymorphic arguments, and grouping of equal values",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT TIME WITHOUT TIME ZONE '040506.789+08';",
                    expected: Expected::Rows {
                        columns: &[Column("time", TIME)],
                        rows: &[
                            &[T("04:05:06.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT TIME '0405', TIME '040506', TIMETZ '040506+02';",
                    expected: Expected::Rows {
                        columns: &[Column("time", TIME), Column("time", TIME), Column("timetz", TIMETZ)],
                        rows: &[
                            &[T("04:05:00"), T("04:05:06"), T("04:05:06+02")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT TIMESTAMP 'Feb 10 5:32PM 1997', TIMESTAMP 'Feb 10 16:32:05 99', TIMESTAMP '2022-01-01 10:00am';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP), Column("timestamp", TIMESTAMP), Column("timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("1997-02-10 17:32:00"), T("1999-02-10 16:32:05"), T("2022-01-01 10:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT TIMESTAMPTZ '2022-02-01 23:45:01 America/New_York' = TIMESTAMPTZ '2022-02-02 04:45:01 UTC';",
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
                    query: "SELECT '20220101 040506.5'::TIMESTAMP;",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("2022-01-01 04:05:06.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(ARRAY[1], ARRAY[2]);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_append(integer[], integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(ARRAY[1], 2::BIGINT);",
                    expected: Expected::Rows {
                        columns: &[Column("array_append", INT8_ARRAY)],
                        rows: &[
                            &[T("{1,2}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT x, count(*) FROM (VALUES (25::NUMERIC), (25.0), (25.00), (26)) t(x) GROUP BY x ORDER BY x;",
                    expected: Expected::Rows {
                        columns: &[Column("x", NUMERIC), Column("count", INT8)],
                        rows: &[
                            &[T("25"), T("3")],
                            &[T("26"), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT DISTINCT x FROM (VALUES ('{"a": 1}'::JSONB), ('{"a": 1.0}')) t(x);"#,
                    expected: Expected::Rows {
                        columns: &[Column("x", JSONB)],
                        rows: &[
                            &[T(r#"{"a": 1}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT x FROM (VALUES ('-0'::FLOAT8), (0)) t(x) GROUP BY x;",
                    expected: Expected::Rows {
                        columns: &[Column("x", FLOAT8)],
                        rows: &[
                            &[T("-0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'Testing2'::REGCLASS;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "testing2" does not exist"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a.b.c.d'::REGCLASS;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "improper relation name (too many dotted names): a.b.c.d", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'otherdb.public.t'::REGCLASS;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cross-database references are not implemented: "otherdb.public.t""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"integer"'::REGTYPE;"#,
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "integer" does not exist"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"integer'::REGTYPE;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"unterminated quoted identifier at or near ""integer""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT repeat('x', 70)::NAME, length(repeat('x', 70)::NAME);",
                    expected: Expected::Rows {
                        columns: &[Column("repeat", NAME), Column("length", INT4)],
                        rows: &[
                            &[T("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), T("63")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT true::NAME, false::TEXT;",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("text", TEXT)],
                        rows: &[
                            &[T("t"), T("false")],
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
fn test_interval_modifier_and_precision_rules() {
    run_scripts(&[
        ScriptTest {
            name: "interval type modifiers",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '1 year 2 months 3 days 04:05:06.789'::interval year, '1 year 2 months 3 days 04:05:06.789'::interval day to minute, '1.789'::interval second(1);",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("1 year"), T("1 year 2 mons 3 days 04:05:00"), T("00:00:01.8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE it (a interval(2), b interval hour to second(3), c interval year to month);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type(atttypid, atttypmod), atttypmod FROM pg_attribute WHERE attrelid = 'it'::regclass AND attnum > 0 ORDER BY attnum;",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT), Column("atttypmod", INT4)],
                        rows: &[
                            &[T("interval(2)"), T("2147418114")],
                            &[T("interval hour to second(3)"), T("469762051")],
                            &[T("interval year to month"), T("458751")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO it VALUES ('1 day 01:02:03.4567', '1 day 01:02:03.4567', '1 year 2 months 3 days');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM it;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INTERVAL), Column("b", INTERVAL), Column("c", INTERVAL)],
                        rows: &[
                            &[T("1 day 01:02:03.46"), T("1 day 01:02:03.457"), T("1 year 2 mons")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1'::interval(-1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "-""#, position: 22, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "precision of converted times",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '2020-01-01 10:00:00.555555'::timestamp::timestamptz(1)::text, '10:00:00.555555'::time::interval(1), '2020-01-01 10:00:00.555555'::timestamp::time(2);",
                    expected: Expected::Rows {
                        columns: &[Column("text", TEXT), Column("interval", INTERVAL), Column("time", TIME)],
                        rows: &[
                            &[T("2020-01-01 10:00:00.6-08"), T("10:00:00.6"), T("10:00:00.56")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(current_time(2)::text) < length(current_time(5)::text), length(localtime(0)::text), length(current_timestamp(0)::text) = length(now()::timestamptz(0)::text);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("length", INT4), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("8"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "interval fields read by the modifier",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '1' YEAR, INTERVAL '2' MONTH, INTERVAL '3' DAY, INTERVAL '4' HOUR, INTERVAL '5' MINUTE, INTERVAL '6.5' SECOND, INTERVAL '7' DAY TO HOUR, INTERVAL '8' HOUR TO MINUTE, INTERVAL '1.5' YEAR, '2'::interval minute;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("1 year"), T("2 mons"), T("3 days"), T("04:00:00"), T("00:05:00"), T("00:00:06.5"), T("07:00:00"), T("00:08:00"), T("1 year"), T("00:02:00")],
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
fn test_numeric_datetime_input_rules() {
    run_scripts(&[
        ScriptTest {
            name: "numeric datetime input forms",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET timezone = 'America/Los_Angeles';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date '1999.008';",
                    expected: Expected::Rows {
                        columns: &[Column("date", DATE)],
                        rows: &[
                            &[T("1999-01-08")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '1997.041 17:32:01 UTC';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("1997-02-10 17:32:01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamptz '1997.041 17:32:01 UTC';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1997-02-10 17:32:01+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT time without time zone '040506.789-08';",
                    expected: Expected::Rows {
                        columns: &[Column("time", TIME)],
                        rows: &[
                            &[T("04:05:06.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT time without time zone 'T040506.789+08';",
                    expected: Expected::Rows {
                        columns: &[Column("time", TIME)],
                        rows: &[
                            &[T("04:05:06.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT time without time zone 'T040506.789-08';",
                    expected: Expected::Rows {
                        columns: &[Column("time", TIME)],
                        rows: &[
                            &[T("04:05:06.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp with time zone '20011227 040506-08';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-12-27 12:05:06+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp with time zone '20011227 040506.789-08';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-12-27 12:05:06.789+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp with time zone '20011227T040506-08';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-12-27 12:05:06+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp with time zone '20011227T040506.789+08';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-12-26 20:05:06.789+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp with time zone 'J2452271-08';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-12-27 08:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamptz '1000000312 23:58:48 IST';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("100000-03-12 21:58:48+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("the zone's daylight saving rules are not extended past the last year that tzdata lists"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamptz '100000312 23:58:48 IST';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("10000-03-12 21:58:48+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("the zone's daylight saving rules are not extended past the last year that tzdata lists"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interval '+1 -1:00:00', interval '-1 +1:00:00';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("1 day -01:00:00"), T("-1 days +01:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interval '+1-2 -3 +4:05:06.789', interval '-1-2 +3 -4:05:06.789';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("1 year 2 mons -3 days +04:05:06.789"), T("-1 years -2 mons +3 days -04:05:06.789")],
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
fn test_values_written_into_rows() {
    run_scripts(&[
        ScriptTest {
            name: "values written straight into result rows",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE ov (id INT PRIMARY KEY, f4 FLOAT4, f8 FLOAT8, d DATE, ts TIMESTAMP, n INT8, b BOOLEAN);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ov VALUES (1, '2636433.25', 2636433.25, '2024-02-29', '2024-02-29 13:05:09.120300', -9223372036854775808, true), (2, '-0', '-0', '0001-01-01', '0001-01-01 00:00:00', 0, false), (3, 'NaN', 'Infinity', 'infinity', '-infinity', 9223372036854775807, NULL), (4, '1.5e-07', '1e15', '0044-03-15 BC', '0044-03-15 12:00:00.5 BC', 42, true), (5, '3.4028235e38', '1.7976931348623157e308', '9999-12-31', '12345-12-31 23:59:59.999999', -1, NULL), (6, '0.1', '0.30000000000000004', '1999-01-08', '1999-01-08 04:05:06', 7, false);",
                    expected: Expected::Tag("INSERT 0 6"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM ov ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("f4", FLOAT4), Column("f8", FLOAT8), Column("d", DATE), Column("ts", TIMESTAMP), Column("n", INT8), Column("b", BOOL)],
                        rows: &[
                            &[T("1"), T("2.6364333e+06"), T("2636433.25"), T("2024-02-29"), T("2024-02-29 13:05:09.1203"), T("-9223372036854775808"), T("t")],
                            &[T("2"), T("-0"), T("-0"), T("0001-01-01"), T("0001-01-01 00:00:00"), T("0"), T("f")],
                            &[T("3"), T("NaN"), T("Infinity"), T("infinity"), T("-infinity"), T("9223372036854775807"), Null],
                            &[T("4"), T("1.5e-07"), T("1e+15"), T("0044-03-15 BC"), T("0044-03-15 12:00:00.5 BC"), T("42"), T("t")],
                            &[T("5"), T("3.4028235e+38"), T("1.7976931348623157e+308"), T("9999-12-31"), T("12345-12-31 23:59:59.999999"), T("-1"), Null],
                            &[T("6"), T("0.1"), T("0.30000000000000004"), T("1999-01-08"), T("1999-01-08 04:05:06"), T("7"), T("f")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f4::numeric, f8::numeric FROM ov WHERE id IN (1, 4, 6) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("f4", NUMERIC), Column("f8", NUMERIC)],
                        rows: &[
                            &[T("2636430"), T("2636433.25")],
                            &[T("0.00000015"), T("1000000000000000")],
                            &[T("0.1"), T("0.3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f4 * 3, f8 / 7 FROM ov ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8), Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("7909299.75"), T("376633.3214285714")],
                            &[T("-0"), T("-0")],
                            &[T("NaN"), T("Infinity")],
                            &[T("4.500000159168849e-07"), T("142857142857142.84")],
                            &[T("1.0208470399155866e+39"), T("2.5681330498033083e+307")],
                            &[T("0.30000000447034836"), T("0.042857142857142864")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET DateStyle = 'SQL, DMY';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT d, ts FROM ov ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("d", DATE), Column("ts", TIMESTAMP)],
                        rows: &[
                            &[T("2024-02-29"), T("2024-02-29 13:05:09.1203")],
                            &[T("0001-01-01"), T("0001-01-01 00:00:00")],
                            &[T("infinity"), T("-infinity")],
                            &[T("0044-03-15 BC"), T("0044-03-15 12:00:00.5 BC")],
                            &[T("9999-12-31"), T("12345-12-31 23:59:59.999999")],
                            &[T("1999-01-08"), T("1999-01-08 04:05:06")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET DateStyle = 'German';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT d, ts FROM ov ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("d", DATE), Column("ts", TIMESTAMP)],
                        rows: &[
                            &[T("2024-02-29"), T("2024-02-29 13:05:09.1203")],
                            &[T("0001-01-01"), T("0001-01-01 00:00:00")],
                            &[T("infinity"), T("-infinity")],
                            &[T("0044-03-15 BC"), T("0044-03-15 12:00:00.5 BC")],
                            &[T("9999-12-31"), T("12345-12-31 23:59:59.999999")],
                            &[T("1999-01-08"), T("1999-01-08 04:05:06")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET DateStyle = 'ISO, MDY';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT d, ts FROM ov ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("d", DATE), Column("ts", TIMESTAMP)],
                        rows: &[
                            &[T("2024-02-29"), T("2024-02-29 13:05:09.1203")],
                            &[T("0001-01-01"), T("0001-01-01 00:00:00")],
                            &[T("infinity"), T("-infinity")],
                            &[T("0044-03-15 BC"), T("0044-03-15 12:00:00.5 BC")],
                            &[T("9999-12-31"), T("12345-12-31 23:59:59.999999")],
                            &[T("1999-01-08"), T("1999-01-08 04:05:06")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_text_search_types() {
    run_scripts(&[
        ScriptTest {
            name: "text search types read and write as Postgres does",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'word'::tsquery, 'phrase & (another | term)'::tsquery;",
                    expected: Expected::Rows {
                        columns: &[Column("tsquery", TSQUERY), Column("tsquery", TSQUERY)],
                        rows: &[
                            &[T("'word'"), T("'phrase' & ( 'another' | 'term' )")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'simple'::tsvector, 'complex & (query | terms)'::tsvector;",
                    expected: Expected::Rows {
                        columns: &[Column("tsvector", TSVECTOR), Column("tsvector", TSVECTOR)],
                        rows: &[
                            &[T("'simple'"), T("'&' '(query' 'complex' 'terms)' '|'")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 'a:1,3B b:2A a:3C ''it''''s'' x\\y'::tsvector;"#,
                    expected: Expected::Rows {
                        columns: &[Column("tsvector", TSVECTOR)],
                        rows: &[
                            &[T(r#"'a':1,3B 'b':2A 'it''s' 'x\\y'"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '!a & !(b | c) <-> d <2> e'::tsquery, 'a:*AB | b:C'::tsquery, '(a | b) & c'::tsquery, 'a <-> (b <-> c)'::tsquery, '!!a'::tsquery, '!(a & b)'::tsquery;",
                    expected: Expected::Rows {
                        columns: &[Column("tsquery", TSQUERY), Column("tsquery", TSQUERY), Column("tsquery", TSQUERY), Column("tsquery", TSQUERY), Column("tsquery", TSQUERY), Column("tsquery", TSQUERY)],
                        rows: &[
                            &[T("!'a' & !( 'b' | 'c' ) <-> 'd' <2> 'e'"), T("'a':*AB | 'b':C"), T("( 'a' | 'b' ) & 'c'"), T("'a' <-> ( 'b' <-> 'c' )"), T("!!'a'"), T("!( 'a' & 'b' )")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ''::tsvector, '  '::tsvector, 'a:20000'::tsvector;",
                    expected: Expected::Rows {
                        columns: &[Column("tsvector", TSVECTOR), Column("tsvector", TSVECTOR), Column("tsvector", TSVECTOR)],
                        rows: &[
                            &[T(""), T(""), T("'a':16383")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a &'::tsquery;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"no operand in tsquery: "a &""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a:0'::tsvector;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"wrong position info in tsvector: "a:0""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '''unterminated'::tsvector;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error in tsvector: "'unterminated""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t_ts (id INT PRIMARY KEY, v TSVECTOR, q TSQUERY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_ts VALUES (1, 'b a c', 'a & b'), (2, 'x', 'x | y'), (3, NULL, NULL);",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_ts ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TSVECTOR), Column("q", TSQUERY)],
                        rows: &[
                            &[T("1"), T("'a' 'b' 'c'"), T("'a' & 'b'")],
                            &[T("2"), T("'x'"), T("'x' | 'y'")],
                            &[T("3"), Null, Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v::text, q::text FROM t_ts WHERE v IS NOT NULL ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT), Column("q", TEXT)],
                        rows: &[
                            &[T("1"), T("'a' 'b' 'c'"), T("'a' & 'b'")],
                            &[T("2"), T("'x'"), T("'x' | 'y'")],
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
fn test_range_types() {
    run_scripts(&[
        ScriptTest {
            name: "Range literals, constructors, and canonical forms",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[1,5)'::int4range, '(1,5]'::int4range, '[1,5]'::int4range, '(1,2)'::int4range, 'empty'::int4range, '[,5)'::int4range;",
                    expected: Expected::Rows {
                        columns: &[Column("int4range", INT4RANGE), Column("int4range", INT4RANGE), Column("int4range", INT4RANGE), Column("int4range", INT4RANGE), Column("int4range", INT4RANGE), Column("int4range", INT4RANGE)],
                        rows: &[
                            &[T("[1,5)"), T("[2,6)"), T("[1,6)"), T("empty"), T("empty"), T("(,5)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '[1.5,2.25]'::numrange, '["2020-01-01 10:00","2020-01-02")'::tsrange, '[2020-01-01,2020-01-05]'::daterange;"#,
                    expected: Expected::Rows {
                        columns: &[Column("numrange", NUMRANGE), Column("tsrange", TSRANGE), Column("daterange", DATERANGE)],
                        rows: &[
                            &[T("[1.5,2.25]"), T(r#"["2020-01-01 10:00:00","2020-01-02 00:00:00")"#), T("[2020-01-01,2020-01-06)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '  [ 1 , 5 )  '::int4range;",
                    expected: Expected::Rows {
                        columns: &[Column("int4range", INT4RANGE)],
                        rows: &[
                            &[T("[1,5)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[5,1)'::int4range;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "range lower bound must be less than or equal to range upper bound", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,5'::int4range;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed range literal: "[1,5""#, detail: "Unexpected end of input.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1,5)'::int4range;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed range literal: "1,5)""#, detail: "Missing left parenthesis or bracket.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,5,6)'::int4range;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed range literal: "[1,5,6)""#, detail: "Too many commas.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,5) x'::int4range;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed range literal: "[1,5) x""#, detail: "Junk after right parenthesis or bracket.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT int4range(1,5), int4range(1,5,'[]'), int4range(NULL,5), numrange(1.5, 2), daterange('2020-01-01','2020-02-01','()');",
                    expected: Expected::Rows {
                        columns: &[Column("int4range", INT4RANGE), Column("int4range", INT4RANGE), Column("int4range", INT4RANGE), Column("numrange", NUMRANGE), Column("daterange", DATERANGE)],
                        rows: &[
                            &[T("[1,5)"), T("[1,6)"), T("(,5)"), T("[1.5,2)"), T("[2020-01-02,2020-02-01)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT int4range(5,1);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "range lower bound must be less than or equal to range upper bound", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT int4range(1,5,'x');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "invalid range bound flags", hint: r#"Valid values are "[]", "[)", "(]", and "()"."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Range operators and functions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT int4range(1,5) @> 3, int4range(1,5) @> 5, int4range(1,5) @> int4range(2,3), int4range(1,5) @> '[2,3)', 3 <@ int4range(1,5);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT int4range(1,5) && int4range(4,6), int4range(1,5) << int4range(5,6), int4range(1,5) -|- int4range(5,6), int4range(1,5) &< int4range(2,3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT int4range(1,5) + int4range(5,9), int4range(1,5) * int4range(3,9), int4range(1,9) - int4range(3,5);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "result of range difference would not be contiguous", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT int4range(1,5) + int4range(6,9);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "result of range union would not be contiguous", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT int4range(1,9) - int4range(3,5);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "result of range difference would not be contiguous", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lower(int4range(1,5)), upper(numrange(1.5,NULL)), isempty('empty'::int4range), lower_inc(int4range(1,5)), upper_inf(int4range(1,NULL));",
                    expected: Expected::Rows {
                        columns: &[Column("lower", INT4), Column("upper", NUMERIC), Column("isempty", BOOL), Column("lower_inc", BOOL), Column("upper_inf", BOOL)],
                        rows: &[
                            &[T("1"), Null, T("t"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT range_merge(int4range(1,3), int4range(7,9)), int4range(1,5) = int4range(1,5,'[)'), int4range(1,5) < int4range(1,6);",
                    expected: Expected::Rows {
                        columns: &[Column("range_merge", INT4RANGE), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("[1,9)"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT range_overlaps_multirange(numrange(4.0, 4.2), '{[4.1,5)}'::nummultirange), range_contains_elem(int4range(1,5), 3), range_cmp(int4range(1,5), int4range(1,6)), elem_contained_by_range(3, int4range(1,5));",
                    expected: Expected::Rows {
                        columns: &[Column("range_overlaps_multirange", BOOL), Column("range_contains_elem", BOOL), Column("range_cmp", INT4), Column("elem_contained_by_range", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("-1"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Multiranges",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '{[1,3), [5,7), [2,4)}'::int4multirange, '{}'::int4multirange, '{[1,2), empty}'::int4multirange;",
                    expected: Expected::Rows {
                        columns: &[Column("int4multirange", INT4MULTIRANGE), Column("int4multirange", INT4MULTIRANGE), Column("int4multirange", INT4MULTIRANGE)],
                        rows: &[
                            &[T("{[1,4),[5,7)}"), T("{}"), T("{[1,2)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{[1,3)'::int4multirange;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed multirange literal: "{[1,3)""#, detail: "Unexpected end of input.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT int4multirange(int4range(1,3), int4range(2,6)), int4multirange(), nummultirange(numrange(1,2));",
                    expected: Expected::Rows {
                        columns: &[Column("int4multirange", INT4MULTIRANGE), Column("int4multirange", INT4MULTIRANGE), Column("nummultirange", NUMMULTIRANGE)],
                        rows: &[
                            &[T("{[1,6)}"), T("{}"), T("{[1,2)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{[1,3),[5,7)}'::int4multirange @> 6, '{[1,3),[5,7)}'::int4multirange && int4range(3,5), '{[1,3)}'::int4multirange + '{[3,9)}', '{[1,9)}'::int4multirange - '{[3,4)}';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", INT4MULTIRANGE), Column("?column?", INT4MULTIRANGE)],
                        rows: &[
                            &[T("t"), T("f"), T("{[1,9)}"), T("{[1,3),[4,9)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{[1,3),[5,7)}'::int4multirange -|- int4range(7,9), '{[1,3),[5,7)}'::int4multirange -|- int4range(3,5), int4range(-5,1) -|- '{[1,3),[5,7)}'::int4multirange;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest('{[1,3),[5,7)}'::int4multirange);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4RANGE)],
                        rows: &[
                            &[T("[1,3)")],
                            &[T("[5,7)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT multirange(int4range(1,5)), lower('{[1,3),[5,7)}'::int4multirange), upper('{[1,3),[5,7)}'::int4multirange);",
                    expected: Expected::Rows {
                        columns: &[Column("multirange", INT4MULTIRANGE), Column("lower", INT4), Column("upper", INT4)],
                        rows: &[
                            &[T("{[1,5)}"), T("1"), T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Range columns and aggregates",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE range_cols (id INT PRIMARY KEY, r INT4RANGE, m INT4MULTIRANGE, n NUMRANGE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO range_cols VALUES (1, '[1,10)', '{[1,2),[4,8)}', '(1.5,)'), (2, 'empty', '{}', NULL), (3, '(,5]', '{(,0)}', '[3,3]');",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM range_cols ORDER BY r;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("r", INT4RANGE), Column("m", INT4MULTIRANGE), Column("n", NUMRANGE)],
                        rows: &[
                            &[T("2"), T("empty"), T("{}"), Null],
                            &[T("3"), T("(,6)"), T("{(,0)}"), T("[3,3]")],
                            &[T("1"), T("[1,10)"), T("{[1,2),[4,8)}"), T("(1.5,)")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX range_cols_r ON range_cols USING gist (r);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM range_cols WHERE r @> 4 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT range_agg(r), range_agg(m), range_intersect_agg(r) FROM range_cols WHERE id <> 2;",
                    expected: Expected::Rows {
                        columns: &[Column("range_agg", INT4MULTIRANGE), Column("range_agg", INT4MULTIRANGE), Column("range_intersect_agg", INT4RANGE)],
                        rows: &[
                            &[T("{(,10)}"), T("{(,0),[1,2),[4,8)}"), T("[1,6)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT range_agg(r) FROM range_cols WHERE false;",
                    expected: Expected::Rows {
                        columns: &[Column("range_agg", INT4MULTIRANGE)],
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
        ScriptTest {
            name: "User-defined range types",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE TYPE textrange AS RANGE (subtype = text, collation = "C");"#,
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT textrange('a', 'c'), textrange('a', 'c', '[]'), '[a,b)'::textrange, textmultirange(), textmultirange(textrange('a','c'), textrange('b','f'));",
                    expected: Expected::Rows {
                        columns: &[Column("textrange", USER_DEFINED), Column("textrange", USER_DEFINED), Column("textrange", USER_DEFINED), Column("textmultirange", USER_DEFINED), Column("textmultirange", USER_DEFINED)],
                        rows: &[
                            &[T("[a,c)"), T("[a,c]"), T("[a,b)"), T("{}"), T("{[a,f)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ''::textmultirange;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed multirange literal: """#, detail: "Missing left brace.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE textrange2 AS RANGE (subtype = text, multirange_type_name = multitextrange2);",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT multitextrange2(textrange2('a','z'));",
                    expected: Expected::Rows {
                        columns: &[Column("multitextrange2", USER_DEFINED)],
                        rows: &[
                            &[T("{[a,z)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE bogus AS RANGE (subtype = int4, bogus = 1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"type attribute "bogus" not recognized"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE TYPE bogus2 AS RANGE (collation = "C");"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"type attribute "subtype" is required"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE text_ranges (t textrange, m textmultirange);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO text_ranges VALUES ('[a,m)', '{[a,c),[x,z]}');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM text_ranges WHERE t @> 'b'::text;",
                    expected: Expected::Rows {
                        columns: &[Column("t", USER_DEFINED), Column("m", USER_DEFINED)],
                        rows: &[
                            &[T("[a,m)"), T("{[a,c),[x,z]}")],
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
fn test_text_search_functions() {
    run_scripts(&[
        ScriptTest {
            name: "text search matching",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'a:1 b:2 c:3'::tsvector @@ '!x <-> b', 'a:1 b:2'::tsvector @@ '!a <-> b', 'a:1 b:2'::tsvector @@ 'a <-> !b', 'a:1 c:2'::tsvector @@ 'a <-> !b', 'a:1 b:3'::tsvector @@ '!a <-> !b';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("f"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a:1 b:2 c:3'::tsvector @@ '(a & b) <-> c', 'a:1 b:2 c:3'::tsvector @@ '(a | b) <-> c', 'a:1 b:2 c:3'::tsvector @@ 'a <-> (b & c)', 'a:1 b:2 c:3'::tsvector @@ '(!a | b) <-> c', 'a:2 b:2 c:3'::tsvector @@ '(a & b) <-> c';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), T("t"), T("f"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'wd:1A wd:2'::tsvector @@ 'wd:A', 'wd:2B,3A'::tsvector @@ 'wd:A', 'wd'::tsvector @@ 'wd:A', 'wd'::tsvector @@ '!wd:A';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT strip('wa:1A'::tsvector) @@ 'w:*A'::tsquery, strip('wa:1A'::tsvector) @@ '!w:*A'::tsquery, 'x y'::tsvector @@ '!(z <-> y)', 'x:1 y:2'::tsvector @@ '(z <-> y) | x', 'x:1 y:2 q:3'::tsvector @@ '((z <-> w) | x) <-> y';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a b'::tsquery && 'c', 'a'::tsquery || '!b', 'a'::tsquery <-> 'b', tsquery_phrase('a', 'b', 3), !! 'a & b'::tsquery, numnode('a & !b'), querytree('a & !b'), 'a & b'::tsquery @> 'a', 'a'::tsquery <@ 'a | b';",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error in tsquery: "a b""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tsvector functions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'w:12B w:13* w:12,5,6 a:1,3* a:3 w asd:1dc asd'::tsvector;",
                    expected: Expected::Rows {
                        columns: &[Column("tsvector", TSVECTOR)],
                        rows: &[
                            &[T("'a':1,3A 'asd':1C 'w':5,6,12B,13A")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a:1*A'::tsvector;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error in tsvector: "a:1*A""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a:1 b:2'::tsvector || 'b:1 c:3A', strip('a:1 b:2A'), length('a b c'::tsvector), setweight('a:1 b:2'::tsvector, 'b'), setweight('a asd w:5,6,12B,13A zxc'::tsvector, 'c', ARRAY['a', 'zxc', '', NULL]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TSVECTOR), Column("strip", TSVECTOR), Column("length", INT4), Column("setweight", TSVECTOR), Column("setweight", TSVECTOR)],
                        rows: &[
                            &[T("'a':1 'b':2,3 'c':5A"), T("'a' 'b'"), T("3"), T("'a':1B 'b':2B"), T("'a' 'asd' 'w':5,6,12B,13A 'zxc'")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_delete('base hidden rebel spaceship strike'::tsvector, ARRAY['spaceship','leya','rebel', '', NULL]), ts_delete('a b c'::tsvector, 'b'), ts_filter('a:1A b:2B c:3C'::tsvector, '{a,c}');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_delete", TSVECTOR), Column("ts_delete", TSVECTOR), Column("ts_filter", TSVECTOR)],
                        rows: &[
                            &[T("'base' 'hidden' 'strike'"), T("'a' 'c'"), T("'a':1A 'c':3C")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT tsvector_to_array('b:2 a:1'::tsvector), array_to_tsvector(ARRAY['b', 'a', 'b']);",
                    expected: Expected::Rows {
                        columns: &[Column("tsvector_to_array", TEXT_ARRAY), Column("array_to_tsvector", TSVECTOR)],
                        rows: &[
                            &[T("{a,b}"), T("'a' 'b'")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_tsvector(ARRAY['a', NULL]);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "lexeme array may not contain nulls", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest('a:1A,3 b c:2'::tsvector);",
                    expected: Expected::Rows {
                        columns: &[Column("lexeme", TEXT), Column("positions", INT2_ARRAY), Column("weights", TEXT_ARRAY)],
                        rows: &[
                            &[T("a"), T("{1,3}"), T("{A,D}")],
                            &[T("b"), Null, Null],
                            &[T("c"), T("{2}"), T("{D}")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tsquery ordering",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'a | f' < 'b & c'::tsquery, 'a' < 'b'::tsquery, 'ab' < 'b'::tsquery, 'a & b' < 'a'::tsquery, 'a <-> b' < 'a <2> b'::tsquery, '!a' < 'a'::tsquery, 'a & b' = 'b & a'::tsquery;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("f"), T("f"), T("f"), T("f"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT q FROM (VALUES ('a'::tsquery), ('b'), ('c'), ('foo'), ('a & b'), ('b & a'), ('!x'), ('x <-> y'), ('x <3> y'), ('x | y'), ('qwerty'), ('new <-> york'), ('moscow'), ('a:*'), ('a:AB')) v(q) ORDER BY q;",
                    expected: Expected::Rows {
                        columns: &[Column("q", TSQUERY)],
                        rows: &[
                            &[T("'a'")],
                            &[T("'a':*")],
                            &[T("'a':AB")],
                            &[T("'c'")],
                            &[T("'b'")],
                            &[T("'foo'")],
                            &[T("'qwerty'")],
                            &[T("'moscow'")],
                            &[T("!'x'")],
                            &[T("'x' <3> 'y'")],
                            &[T("'x' <-> 'y'")],
                            &[T("'x' | 'y'")],
                            &[T("'b' & 'a'")],
                            &[T("'a' & 'b'")],
                            &[T("'new' <-> 'york'")],
                        ],
                        tag: "SELECT 15",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "text search ranking",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ts_rank(' a:1 s:2C d g'::tsvector, 'a | s'), ts_rank(' a:1 sa:2C d g'::tsvector, 'a | s:*'), ts_rank(' a:1 s:2B d g'::tsvector, 'a & s'), ts_rank(' a:1 s:2 d g'::tsvector, 'a & s');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rank", FLOAT4), Column("ts_rank", FLOAT4), Column("ts_rank", FLOAT4), Column("ts_rank", FLOAT4)],
                        rows: &[
                            &[T("0.091189064"), T("0.091189064"), T("0.19820644"), T("0.09910322")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rank_cd(' a:1 sa:3C sab:2c d g'::tsvector, 'a | sa:*'), ts_rank_cd(' a:1 s:2,3A d:2A g'::tsvector, 'a <2> s:A'), ts_rank_cd(' a:1 sa:2A sb:2D g'::tsvector, 'a <-> s:* <-> sa:B'), ts_rank_cd(' a:1 s:2 d g'::tsvector, 'a & s');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rank_cd", FLOAT4), Column("ts_rank_cd", FLOAT4), Column("ts_rank_cd", FLOAT4), Column("ts_rank_cd", FLOAT4)],
                        rows: &[
                            &[T("0.5"), T("0.09090909"), T("0"), T("0.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rank('{0.1,0.2,0.3,1}', 'a:1A b:3 c:5,9B'::tsvector, 'a & c', 1), ts_rank('a:1A b:3 c:5,9B'::tsvector, 'a & c', 63), ts_rank_cd('a:1A b:3 c:5,9B a:12'::tsvector, 'a & c', 63), ts_rank_cd('{-1,0.2,0.3,1}', 'a:1A b:3 c:5,9B a:12'::tsvector, 'a <-> c' , 4);",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rank", FLOAT4), Column("ts_rank", FLOAT4), Column("ts_rank_cd", FLOAT4), Column("ts_rank_cd", FLOAT4)],
                        rows: &[
                            &[T("0.23597424"), T("0.010407742"), T("0.00012250624"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rank('{0.1,0.2}', 'a'::tsvector, 'a');",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "array of weight is too short", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rank('{0.1,0.2,0.3,2}', 'a'::tsvector, 'a');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "weight out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rank('{0.1,0.2,0.3,NULL}', 'a'::tsvector, 'a');",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "array of weight must not contain nulls", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rank('a b c'::tsvector, 'a & b'), ts_rank('a:1 b:1'::tsvector, 'a & b'), ts_rank('a:1 ab:3 abc:5'::tsvector, 'a:* & ab'), ts_rank_cd('a:1 b:2 a:3 c:7 b:9 a:20'::tsvector, 'a & b'), ts_rank_cd('a:1 b:2 a:3 c:7 b:9 a:20'::tsvector, '(a | c) & !b');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rank", FLOAT4), Column("ts_rank", FLOAT4), Column("ts_rank", FLOAT4), Column("ts_rank_cd", FLOAT4), Column("ts_rank_cd", FLOAT4)],
                        rows: &[
                            &[T("1e-16"), T("1e-20"), T("0.098500855"), T("0.22575758"), T("0.1")],
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
fn test_text_search_parsing() {
    run_scripts(&[
        ScriptTest {
            name: "to_tsvector and its configurations",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_tsvector('simple', '1 2 3 1'), to_tsvector('english', 'The quick brown foxes jumped over the lazy dogs');",
                    expected: Expected::Rows {
                        columns: &[Column("to_tsvector", TSVECTOR), Column("to_tsvector", TSVECTOR)],
                        rows: &[
                            &[T("'1':1,4 '2':2 '3':3"), T("'brown':3 'dog':9 'fox':4 'jump':5 'lazi':8 'quick':2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_tsvector('english', 'running runs ran easily fairly generously skies dying news only');",
                    expected: Expected::Rows {
                        columns: &[Column("to_tsvector", TSVECTOR)],
                        rows: &[
                            &[T("'die':8 'easili':4 'fair':5 'generous':6 'news':9 'ran':3 'run':1,2 'sky':7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_tsvector('english', 'http://www.google.com/foo.bar.html?a=1 foo@bar.com 1.2.3 -1.5e3 www.example.com:8080/path qwe-rty <b>bold</b> &amp; /usr/local/file.txt 12abc ab12');",
                    expected: Expected::Rows {
                        columns: &[Column("to_tsvector", TSVECTOR)],
                        rows: &[
                            &[T("'-1.5e3':6 '/foo.bar.html?a=1':3 '/path':9 '/usr/local/file.txt':14 '1.2.3':5 '12abc':15 'ab12':16 'bold':13 'foo@bar.com':4 'qwe':11 'qwe-rti':10 'rti':12 'www.example.com:8080':8 'www.example.com:8080/path':7 'www.google.com':2 'www.google.com/foo.bar.html?a=1':1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_tsvector('simple', '1 2 3 1') @@ '1 <2> 3', to_tsvector('simple', 'q x q y') @@ 'q <-> (x & y)';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_text_search_config = 'pg_catalog.simple';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_tsvector('The Cats'), get_current_ts_config();",
                    expected: Expected::Rows {
                        columns: &[Column("to_tsvector", TSVECTOR), Column("get_current_ts_config", REGCONFIG)],
                        rows: &[
                            &[T("'cats':2 'the':1"), T("simple")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET default_text_search_config;",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT get_current_ts_config(), 'english'::regconfig, 'pg_catalog.simple'::regconfig, 'english_stem'::regdictionary;",
                    expected: Expected::Rows {
                        columns: &[Column("get_current_ts_config", REGCONFIG), Column("regconfig", REGCONFIG), Column("regconfig", REGCONFIG), Column("regdictionary", REGDICTIONARY)],
                        rows: &[
                            &[T("english"), T("english"), T("simple"), T("english_stem")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'nonexistent'::regconfig;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"text search configuration "nonexistent" does not exist"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_tsquery and its variants",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_tsquery('english', 'qwe & sKies '), to_tsquery('simple', 'qwe & sKies '), to_tsquery('english', '''the wether'':dc & ''           sKies '':BC ');",
                    expected: Expected::Rows {
                        columns: &[Column("to_tsquery", TSQUERY), Column("to_tsquery", TSQUERY), Column("to_tsquery", TSQUERY)],
                        rows: &[
                            &[T("'qwe' & 'sky'"), T("'qwe' & 'skies'"), T("'wether':CD & 'sky':BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_tsquery('english', 'the');",
                    expected: Expected::Rows {
                        columns: &[Column("to_tsquery", TSQUERY)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    notices: &[Diagnostic { code: "00000", message: "text-search query contains only stop words or doesn't contain lexemes, ignored", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT plainto_tsquery('english', 'the and z 1))& fghj'), phraseto_tsquery('english', 'PostgreSQL can be extended by the user in many ways');",
                    expected: Expected::Rows {
                        columns: &[Column("plainto_tsquery", TSQUERY), Column("phraseto_tsquery", TSQUERY)],
                        rows: &[
                            &[T("'z' & '1' & 'fghj'"), T("'postgresql' <3> 'extend' <3> 'user' <2> 'mani' <-> 'way'")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_tsquery('english', '!(a & !b) & c'), to_tsquery('english', 'foo <-> (a <-> the)'), to_tsquery('english', 'pg_class:*');",
                    expected: Expected::Rows {
                        columns: &[Column("to_tsquery", TSQUERY), Column("to_tsquery", TSQUERY), Column("to_tsquery", TSQUERY)],
                        rows: &[
                            &[T("!!'b' & 'c'"), T("'foo'"), T("'pg':* <-> 'class':*")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT websearch_to_tsquery('simple', 'I have a fat:*ABCD cat'), websearch_to_tsquery('english', '"fat rat" or cat dog'), websearch_to_tsquery('english', 'cat -dog'), websearch_to_tsquery('english', 'or cat');"#,
                    expected: Expected::Rows {
                        columns: &[Column("websearch_to_tsquery", TSQUERY), Column("websearch_to_tsquery", TSQUERY), Column("websearch_to_tsquery", TSQUERY), Column("websearch_to_tsquery", TSQUERY)],
                        rows: &[
                            &[T("'i' & 'have' & 'a' & 'fat' & 'abcd' & 'cat'"), T("'fat' <-> 'rat' | 'cat' & 'dog'"), T("'cat' & !'dog'"), T("'cat'")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_tsquery('english', 'a <99999> b');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "distance in phrase operator must be an integer value between zero and 16384 inclusive", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_tsquery('english', 'a & (b');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error in tsquery: "a & (b""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "text search debugging functions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ts_lexize('english_stem', 'skies'), ts_lexize('english_stem', 'the'), ts_lexize('simple', 'FoO');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_lexize", TEXT_ARRAY), Column("ts_lexize", TEXT_ARRAY), Column("ts_lexize", TEXT_ARRAY)],
                        rows: &[
                            &[T("{sky}"), T("{}"), T("{foo}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM ts_token_type('default');",
                    expected: Expected::Rows {
                        columns: &[Column("tokid", INT4), Column("alias", TEXT), Column("description", TEXT)],
                        rows: &[
                            &[T("1"), T("asciiword"), T("Word, all ASCII")],
                            &[T("2"), T("word"), T("Word, all letters")],
                            &[T("3"), T("numword"), T("Word, letters and digits")],
                            &[T("4"), T("email"), T("Email address")],
                            &[T("5"), T("url"), T("URL")],
                            &[T("6"), T("host"), T("Host")],
                            &[T("7"), T("sfloat"), T("Scientific notation")],
                            &[T("8"), T("version"), T("Version number")],
                            &[T("9"), T("hword_numpart"), T("Hyphenated word part, letters and digits")],
                            &[T("10"), T("hword_part"), T("Hyphenated word part, all letters")],
                            &[T("11"), T("hword_asciipart"), T("Hyphenated word part, all ASCII")],
                            &[T("12"), T("blank"), T("Space symbols")],
                            &[T("13"), T("tag"), T("XML tag")],
                            &[T("14"), T("protocol"), T("Protocol head")],
                            &[T("15"), T("numhword"), T("Hyphenated word, letters and digits")],
                            &[T("16"), T("asciihword"), T("Hyphenated word, all ASCII")],
                            &[T("17"), T("hword"), T("Hyphenated word, all letters")],
                            &[T("18"), T("url_path"), T("URL path")],
                            &[T("19"), T("file"), T("File or path name")],
                            &[T("20"), T("float"), T("Decimal notation")],
                            &[T("21"), T("int"), T("Signed integer")],
                            &[T("22"), T("uint"), T("Unsigned integer")],
                            &[T("23"), T("entity"), T("XML entity")],
                        ],
                        tag: "SELECT 23",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM ts_parse('default', '345 qwe@efd.r http://aew.werc.ewr/?ad=qwe&dw +4.0e-10 234.435 5.005 qwe-wer <fr>qwer 1.2.3 readline-4.2');",
                    expected: Expected::Rows {
                        columns: &[Column("tokid", INT4), Column("token", TEXT)],
                        rows: &[
                            &[T("22"), T("345")],
                            &[T("12"), T(" ")],
                            &[T("1"), T("qwe")],
                            &[T("12"), T("@")],
                            &[T("19"), T("efd.r")],
                            &[T("12"), T(" ")],
                            &[T("14"), T("http://")],
                            &[T("5"), T("aew.werc.ewr/?ad=qwe&dw")],
                            &[T("6"), T("aew.werc.ewr")],
                            &[T("18"), T("/?ad=qwe&dw")],
                            &[T("12"), T(" ")],
                            &[T("7"), T("+4.0e-10")],
                            &[T("12"), T(" ")],
                            &[T("20"), T("234.435")],
                            &[T("12"), T(" ")],
                            &[T("20"), T("5.005")],
                            &[T("12"), T(" ")],
                            &[T("16"), T("qwe-wer")],
                            &[T("11"), T("qwe")],
                            &[T("12"), T("-")],
                            &[T("11"), T("wer")],
                            &[T("12"), T(" ")],
                            &[T("13"), T("<fr>")],
                            &[T("1"), T("qwer")],
                            &[T("12"), T(" ")],
                            &[T("8"), T("1.2.3")],
                            &[T("12"), T(" ")],
                            &[T("1"), T("readline")],
                            &[T("20"), T("-4.2")],
                        ],
                        tag: "SELECT 29",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM ts_debug('english', '<myns:foo-bar_baz.blurfl>abc&nm1;def&#xa9;ghi&#245;jkl</myns:foo-bar_baz.blurfl>');",
                    expected: Expected::Rows {
                        columns: &[Column("alias", TEXT), Column("description", TEXT), Column("token", TEXT), Column("dictionaries", REGDICTIONARY_ARRAY), Column("dictionary", REGDICTIONARY), Column("lexemes", TEXT_ARRAY)],
                        rows: &[
                            &[T("tag"), T("XML tag"), T("<myns:foo-bar_baz.blurfl>"), T("{}"), Null, Null],
                            &[T("asciiword"), T("Word, all ASCII"), T("abc"), T("{english_stem}"), T("english_stem"), T("{abc}")],
                            &[T("entity"), T("XML entity"), T("&nm1;"), T("{}"), Null, Null],
                            &[T("asciiword"), T("Word, all ASCII"), T("def"), T("{english_stem}"), T("english_stem"), T("{def}")],
                            &[T("entity"), T("XML entity"), T("&#xa9;"), T("{}"), Null, Null],
                            &[T("asciiword"), T("Word, all ASCII"), T("ghi"), T("{english_stem}"), T("english_stem"), T("{ghi}")],
                            &[T("entity"), T("XML entity"), T("&#245;"), T("{}"), Null, Null],
                            &[T("asciiword"), T("Word, all ASCII"), T("jkl"), T("{english_stem}"), T("english_stem"), T("{jkl}")],
                            &[T("tag"), T("XML tag"), T("</myns:foo-bar_baz.blurfl>"), T("{}"), Null, Null],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias, token, dictionaries, dictionary, lexemes FROM ts_debug('the cats');",
                    expected: Expected::Rows {
                        columns: &[Column("alias", TEXT), Column("token", TEXT), Column("dictionaries", REGDICTIONARY_ARRAY), Column("dictionary", REGDICTIONARY), Column("lexemes", TEXT_ARRAY)],
                        rows: &[
                            &[T("asciiword"), T("the"), T("{english_stem}"), T("english_stem"), T("{}")],
                            &[T("blank"), T(" "), T("{}"), Null, Null],
                            &[T("asciiword"), T("cats"), T("{english_stem}"), T("english_stem"), T("{cat}")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "json to_tsvector",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT to_tsvector('english', '{"a": "aaa in bbb", "b": 123, "c": 456, "d": true, "f": false, "g": null}'::json), to_tsvector('{"a": "aaa bbb ddd ccc", "b": ["eee fff ggg"], "c": {"d": "hhh iii"}}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_tsvector", TSVECTOR), Column("to_tsvector", TSVECTOR)],
                        rows: &[
                            &[T("'aaa':1 'bbb':3"), T("'aaa':1 'bbb':2 'ccc':4 'ddd':3 'eee':6 'fff':7 'ggg':8 'hhh':10 'iii':11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_to_tsvector('english', '{"a": "aaa in bbb", "b": 123, "c": 456, "d": true, "f": false, "g": null}'::json, '["string", "numeric", "boolean", "key"]'), jsonb_to_tsvector('{"a": "aaa in bbb", "b": 1.50, "d": true}'::jsonb, '"all"');"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_to_tsvector", TSVECTOR), Column("jsonb_to_tsvector", TSVECTOR)],
                        rows: &[
                            &[T("'123':8 '456':12 'aaa':2 'b':6 'bbb':4 'c':10 'd':14 'f':18 'fals':20 'g':22 'true':16"), T("'1.50':8 'aaa':2 'b':6 'bbb':4 'd':10 'true':12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_to_tsvector('english', '{"a": "aaa in bbb"}'::jsonb, '{"a": "all"}');"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: "wrong flag type, only arrays and scalars are allowed", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_to_tsvector('english', '{"a": "aaa in bbb"}'::jsonb, '["foo"]');"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"wrong flag in flag array: "foo""#, hint: r#"Possible values are: "string", "numeric", "boolean", "key", and "all"."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_text_search_rewrite_and_headline() {
    run_scripts(&[
        ScriptTest {
            name: "ts_rewrite",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE test_tsquery (txtkeyword TEXT, txtsample TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test_tsquery VALUES ('New York', 'new & york | big & apple | nyc'), ('Moscow', 'moskva | moscow'), ('''Sanct Peter''', 'Peterburg | peter | ''Sanct Peterburg'''), ('''foo bar qq''', 'foo & (bar | qq) & city'), ('1 & (2 <-> 3)', '2 <-> 4'), ('5 <-> 6', '5 <-> 7');",
                    expected: Expected::Tag("INSERT 0 6"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test_tsquery ADD COLUMN keyword tsquery;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test_tsquery SET keyword = to_tsquery('english', txtkeyword);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error in tsquery: "New York""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test_tsquery ADD COLUMN sample tsquery;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test_tsquery SET sample = to_tsquery('english', txtsample::text);",
                    expected: Expected::Tag("UPDATE 6"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rewrite('foo & bar & qq & new & york', 'new & york'::tsquery, 'big & apple | nyc | new & york & city');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rewrite", TSQUERY)],
                        rows: &[
                            &[T("'foo' & 'bar' & 'qq' & ( 'city' & 'new' & 'york' | 'nyc' | 'big' & 'apple' )")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rewrite(ts_rewrite('new & !york ', 'york', '!jersey'), 'jersey', 'mexico');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rewrite", TSQUERY)],
                        rows: &[
                            &[T("'new' & !!'mexico'")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rewrite('moscow & hotel', 'SELECT keyword, sample FROM test_tsquery'::text);",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rewrite", TSQUERY)],
                        rows: &[
                            &[T("'moscow' & 'hotel'")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rewrite('bar & qq & foo & (new <-> york)', 'SELECT keyword, sample FROM test_tsquery'::text);",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rewrite", TSQUERY)],
                        rows: &[
                            &[T("'foo' & 'bar' & 'qq' & 'new' <-> 'york'")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rewrite('5 <-> (1 & (2 <-> 3))', 'SELECT keyword, sample FROM test_tsquery'::text);",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rewrite", TSQUERY)],
                        rows: &[
                            &[T("'5' <-> ( '1' & '2' <-> '3' )")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rewrite(to_tsquery('5 & (6 | 5)'), to_tsquery('5'), to_tsquery(''));",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rewrite", TSQUERY)],
                        rows: &[
                            &[T("'6'")],
                        ],
                        tag: "SELECT 1",
                    },
                    notices: &[Diagnostic { code: "00000", message: r#"text-search query doesn't contain lexemes: """#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rewrite(to_tsquery('!5'), to_tsquery('5'), to_tsquery(''));",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rewrite", TSQUERY)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    notices: &[Diagnostic { code: "00000", message: r#"text-search query doesn't contain lexemes: """#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rewrite(tsquery_phrase('foo', 'foo'), 'foo', 'bar | baz');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rewrite", TSQUERY)],
                        rows: &[
                            &[T("( 'bar' | 'baz' ) <-> ( 'bar' | 'baz' )")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rewrite('a & b & c & d', 'b & d', 'x'), ts_rewrite('a | b | c', 'c | a', 'z & y'), ts_rewrite('a:A & b', 'a', 'c:*');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_rewrite", TSQUERY), Column("ts_rewrite", TSQUERY), Column("ts_rewrite", TSQUERY)],
                        rows: &[
                            &[T("'c' & 'a' & 'x'"), T("'b' | 'z' & 'y'"), T("'b' & 'c':*")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_rewrite('a', 'SELECT 1, 2');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "ts_rewrite query must return two tsquery columns", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ts_headline",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ts_headline('english', 'Lorem ipsum urna.  Nullam nullam ullamcorper urna.', to_tsquery('english','Lorem') && phraseto_tsquery('english','ullamcorper urna'), 'MaxWords=100, MinWords=1');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_headline", TEXT)],
                        rows: &[
                            &[T("<b>Lorem</b> ipsum <b>urna</b>.  Nullam nullam <b>ullamcorper</b> <b>urna</b>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_headline('simple', '1 2 3 1 3'::text, '1 <-> 3', 'MaxWords=2, MinWords=1'), ts_headline('simple', '1 2 3 1 3'::text, '1 & 3', 'MaxWords=4, MinWords=1');",
                    expected: Expected::Rows {
                        columns: &[Column("ts_headline", TEXT), Column("ts_headline", TEXT)],
                        rows: &[
                            &[T("<b>1</b> <b>3</b>"), T("<b>1</b> 2 <b>3</b>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ts_headline('english', 'Day after day, day after day, We stuck, nor breath nor motion, As idle as a painted Ship Upon a painted Ocean. Water, water, every where And all the boards did shrink; Water, water, every where, Nor any drop to drink.', to_tsquery('english', 'ocean'), 'MaxFragments=2, MaxWords=5, MinWords=2, StartSel=<<, StopSel=>>, FragmentDelimiter=" | "');"#,
                    expected: Expected::Rows {
                        columns: &[Column("ts_headline", TEXT)],
                        rows: &[
                            &[T("painted <<Ocean>>. Water, water, every")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ts_headline('english', '<html><b>Sea</b> <a href="x">view</a> wow</html>', to_tsquery('english', 'sea&view'), 'HighlightAll=true');"#,
                    expected: Expected::Rows {
                        columns: &[Column("ts_headline", TEXT)],
                        rows: &[
                            &[T(r#"<html><b><b>Sea</b></b> <a href="x"><b>view</b></a> wow</html>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_headline('foo bar', 'bar'::tsquery, 'Bogus=1');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized headline parameter: "Bogus""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ts_headline('foo bar', 'bar'::tsquery, 'MinWords=5, MaxWords=4');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "MinWords should be less than MaxWords", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ts_headline('{"a": "aaa bbb", "b": {"c": "ccc ddd fff"}, "d": ["ggg hhh", 1.50, true]}'::json, tsquery('bbb & ddd & hhh')), ts_headline('{"a": "aaa bbb", "b": {"c": "ccc ddd fff"}}'::jsonb, tsquery('bbb & ddd'));"#,
                    expected: Expected::Rows {
                        columns: &[Column("ts_headline", JSON), Column("ts_headline", JSONB)],
                        rows: &[
                            &[T(r#"{"a":"aaa <b>bbb</b>","b":{"c":"ccc <b>ddd</b> fff"},"d":["ggg <b>hhh</b>",1.50,true]}"#), T(r#"{"a": "aaa <b>bbb</b>", "b": {"c": "ccc <b>ddd</b> fff"}}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "text search matching of text",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'the cats sat' @@ 'cat'::tsquery, 'the cats sat' @@ 'cats & sat', ts_match_tt('foo bar', 'bars'), 'a b' @@ to_tsquery('simple', 'a');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("ts_match_tt", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE tv (a tsvector);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tv VALUES ('wr qh'), ('wr'), ('x');",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM tv WHERE a @@ ANY ('{wr,qh}');",
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
fn test_geometric_functions() {
    run_scripts(&[
        ScriptTest {
            name: "geometric operators",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT point(1, 2) + point(3, 4), point(1, 2) * point(3, 4), point(1, 2) / point(3, 4), point(0, 0) <-> point(3, 4), point(1, 1) ~= point(1.0000001, 1), point(1, 2) <@ box '((0,0),(5,5))';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", POINT), Column("?column?", POINT), Column("?column?", POINT), Column("?column?", FLOAT8), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("(4,6)"), T("(-5,10)"), T("(0.44,0.08)"), T("5"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT box '((0,0),(2,2))' && box '((1,1),(3,3))', box '((0,0),(2,2))' @> box '((1,1),(2,2))', box '((0,0),(2,2))' # box '((1,1),(3,3))', @@ box '((0,0),(2,2))', area(box '((0,0),(2,3))'), box '((0,0),(2,2))' << box '((3,3),(4,4))';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOX), Column("?column?", POINT), Column("area", FLOAT8), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("(2,2),(1,1)"), T("(1,1)"), T("6"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lseg '[(0,0),(2,2)]' ?# lseg '[(0,2),(2,0)]', lseg '[(0,0),(2,2)]' # lseg '[(0,2),(2,0)]', point(5, 0) ## lseg '[(0,0),(2,2)]', ?- lseg '[(0,0),(2,0)]', @-@ lseg '[(0,0),(3,4)]', lseg '[(0,0),(1,1)]' <-> lseg '[(3,0),(4,0)]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", POINT), Column("?column?", POINT), Column("?column?", BOOL), Column("?column?", FLOAT8), Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("t"), T("(1,1)"), T("(2,2)"), T("t"), T("5"), T("2.23606797749979")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT line(point(0, 0), point(1, 1)), line '{1,-1,0}' ?# line '{1,1,0}', line '{1,-1,0}' # line '{1,1,0}', point(1, 0) <-> line '{1,-1,0}', line '{0,1,0}' ?|| line '{0,2,5}';",
                    expected: Expected::Rows {
                        columns: &[Column("line", LINE), Column("?column?", BOOL), Column("?column?", POINT), Column("?column?", FLOAT8), Column("?column?", BOOL)],
                        rows: &[
                            &[T("{1,-1,0}"), T("t"), T("(0,0)"), T("0.7071067811865476"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT polygon '((0,0),(4,0),(4,4),(0,4))' @> point(2, 2), polygon '((0,0),(4,0),(4,4),(0,4))' @> polygon '((1,1),(2,1),(2,2))', polygon '((0,0),(4,0),(4,4),(0,4))' && polygon '((3,3),(5,3),(5,5))', # polygon '((0,0),(4,0),(4,4))', @@ polygon '((0,0),(4,0),(4,4),(0,4))';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", INT4), Column("?column?", POINT)],
                        rows: &[
                            &[T("t"), T("t"), T("t"), T("3"), T("(2,2)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT circle(point(0, 0), 2) @> point(1, 1), circle '<(0,0),1>' <-> circle '<(5,0),1>', area(circle '<(0,0),1>'), circle '<(0,0),1>' * point(2, 0), polygon(4, circle '<(0,0),1>');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", FLOAT8), Column("area", FLOAT8), Column("?column?", CIRCLE), Column("polygon", POLYGON)],
                        rows: &[
                            &[T("t"), T("3"), T("3.141592653589793"), T("<(0,0),2>"), T("((-1,0),(-6.123233995736766e-17,1),(1,1.2246467991473532e-16),(1.8369701987210297e-16,-1))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT path '[(0,0),(3,4),(6,0)]' + point(1, 1), length(path '[(0,0),(3,4),(6,0)]'), area(path '((0,0),(4,0),(4,4))'), isopen(path '[(0,0),(1,1)]'), pclose(path '[(0,0),(1,1)]'), npoints(path '[(0,0),(1,1)]');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", PATH), Column("length", FLOAT8), Column("area", FLOAT8), Column("isopen", BOOL), Column("pclose", PATH), Column("npoints", INT4)],
                        rows: &[
                            &[T("[(1,1),(4,5),(7,1)]"), T("10"), T("8"), T("t"), T("((0,0),(1,1))"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT point(1, 2) / point(0, 0);",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "geometric casts, subscripts, and float digits",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '((0,0),(1,1))'::box::polygon, npoints('<(1,1),2>'::circle::polygon), '(1,2)'::point::box, '[(1,2),(3,4)]'::lseg::point, '((0,0),(2,2))'::polygon::box, '((0,0),(2,2))'::polygon::path, '((0,0),(2,2))'::box::circle;",
                    expected: Expected::Rows {
                        columns: &[Column("polygon", POLYGON), Column("npoints", INT4), Column("box", BOX), Column("point", POINT), Column("box", BOX), Column("path", PATH), Column("circle", CIRCLE)],
                        rows: &[
                            &[T("((0,0),(0,1),(1,1),(1,0))"), T("12"), T("(1,2),(1,2)"), T("(2,3)"), T("(2,2),(0,0)"), T("((0,0),(2,2))"), T("<(1,1),1.4142135623730951>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[(1,2),(3,4)]'::path::polygon;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "open path cannot be converted to polygon", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE gc (b box, p point);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO gc VALUES ('(1,2)'::point, '(3,4)');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT b, p[0], p[1], p[2], b[0], b[1] FROM gc;",
                    expected: Expected::Rows {
                        columns: &[Column("b", BOX), Column("p", FLOAT8), Column("p", FLOAT8), Column("p", FLOAT8), Column("b", POINT), Column("b", POINT)],
                        rows: &[
                            &[T("(1,2),(1,2)"), T("3"), T("4"), Null, T("(1,2)"), T("(1,2)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('(1,2)'::point)[0:1];",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "slices of fixed-length arrays not implemented", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX gc_b ON gc USING gist (b);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX gc_p ON gc USING spgist (p);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM gc WHERE p <@ box '((0,0),(5,5))';",
                    expected: Expected::Rows {
                        columns: &[Column("b", BOX), Column("p", POINT)],
                        rows: &[
                            &[T("(1,2),(1,2)"), T("(3,4)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET extra_float_digits = -3;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT point(4.0 / 3, 1e-10), 1.0 / 3::float8, 12345.678::float4, 1e100::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("point", POINT), Column("?column?", FLOAT8), Column("float4", FLOAT4), Column("float8", FLOAT8)],
                        rows: &[
                            &[T("(1.3333333333333333,1e-10)"), T("0.3333333333333333"), T("12345.678"), T("1e+100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET extra_float_digits = 0;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1.0 / 3::float8, 0.1::float4, 1e15::float8, 1e16::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8), Column("float4", FLOAT4), Column("float8", FLOAT8), Column("float8", FLOAT8)],
                        rows: &[
                            &[T("0.3333333333333333"), T("0.1"), T("1e+15"), T("1e+16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET extra_float_digits;",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1.0 / 3::float8, 0.1::float4;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8), Column("float4", FLOAT4)],
                        rows: &[
                            &[T("0.3333333333333333"), T("0.1")],
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
fn test_interval_input() {
    run_scripts(&[
        ScriptTest {
            name: "Interval input",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '1 month - 1 second', INTERVAL '2 days - 12:34:56', INTERVAL '-1 days +02:03';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("1 mon -00:00:01"), T("2 days -12:34:56"), T("-1 days +02:03:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '4 millenniums 5 centuries 4 decades 1 year 4 months 4 days 17 minutes 31 seconds'::INTERVAL;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("4541 years 4 mons 4 days 00:17:31")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '1.5 weeks', INTERVAL '0.7 years', INTERVAL '1.25 months', INTERVAL '10.5 microseconds 1 ms';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("10 days 12:00:00"), T("8 mons"), T("1 mon 7 days 12:00:00"), T("00:00:00.00101")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '-2147483648 months -2147483648 days -9223372036854775808 us';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("-178956970 years -8 mons -2147483648 days -2562047788:00:54.775808")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '2147483647 days 2147483647 months';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("178956970 years 7 mons 2147483647 days")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '2147483648 days';",
                    expected: Expected::Error(Diagnostic { code: "22015", message: r#"interval field value out of range: "2147483648 days""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '-2147483648 years';",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "interval out of range", position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '9223372036854775807 microseconds', INTERVAL '-9223372036854.775808 seconds';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("2562047788:00:54.775807"), T("-2562047788:00:54.775808")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '9223372036854775808 microseconds';",
                    expected: Expected::Error(Diagnostic { code: "22015", message: r#"interval field value out of range: "9223372036854775808 microseconds""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '2562047788.01521550194 hours';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("2562047788:00:54.775807")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '5 ago';",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid input syntax for type interval: "5 ago""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '1 2';",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid input syntax for type interval: "1 2""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '1 day 1 day';",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid input syntax for type interval: "1 day 1 day""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '01:02:03 1 second';",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid input syntax for type interval: "01:02:03 1 second""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '1-13';",
                    expected: Expected::Error(Diagnostic { code: "22015", message: r#"interval field value out of range: "1-13""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '1:61';",
                    expected: Expected::Error(Diagnostic { code: "22015", message: r#"interval field value out of range: "1:61""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'infinity';",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid input syntax for type interval: "infinity""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '1 quarter';",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid input syntax for type interval: "1 quarter""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '1 2' MINUTE TO SECOND;",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid input syntax for type interval: "1 2""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '1:2' MINUTE TO SECOND, INTERVAL '1:2.5' HOUR TO SECOND, INTERVAL '5' YEAR, INTERVAL '5' HOUR;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("00:01:02"), T("00:01:02.5"), T("5 years"), T("05:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{ 0 second, @ 1 hour @ 42 minutes @ 20 seconds }'::INTERVAL[];",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL_ARRAY)],
                        rows: &[
                            &[T("{00:00:00,01:42:20}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ISO 8601 interval input",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'P1Y2M3DT4H5M6S', INTERVAL 'P0.5Y', INTERVAL 'P1.5W', INTERVAL 'PT1.5H';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("1 year 2 mons 3 days 04:05:06"), T("6 mons"), T("10 days 12:00:00"), T("01:30:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'P00021015T103020', INTERVAL 'P0002-10-15T10:30:20';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("2 years 10 mons 15 days 10:30:20"), T("2 years 10 mons 15 days 10:30:20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'P0002', INTERVAL 'P0002-10', INTERVAL 'P0002-10-15', INTERVAL 'P0002T1S', INTERVAL 'P0002-10T1S';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("2 years"), T("2 years 10 mons"), T("2 years 10 mons 15 days"), T("2 years 00:00:01"), T("2 years 10 mons 00:00:01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'P10.5e4Y';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("105000 years")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'PT2562047788H54.775807S', INTERVAL 'PT2562047788:00:54.775807';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("2562047788:00:54.775807"), T("2562047788:00:54.775807")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'PT2562047789';",
                    expected: Expected::Error(Diagnostic { code: "22015", message: r#"interval field value out of range: "PT2562047789""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'P2147483648';",
                    expected: Expected::Error(Diagnostic { code: "22015", message: r#"interval field value out of range: "P2147483648""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'P1-2147483647-2147483647';",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "interval out of range", position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'P0.1Y2147483647M';",
                    expected: Expected::Error(Diagnostic { code: "22015", message: r#"interval field value out of range: "P0.1Y2147483647M""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'p1d';",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid input syntax for type interval: "p1d""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL ' P1D';",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid input syntax for type interval: " P1D""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL 'P1X';",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid input syntax for type interval: "P1X""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SQL standard interval input",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET intervalstyle = sql_standard;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '-1 2:03:04', INTERVAL '-1 +2:03:04', INTERVAL '-1-2';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("-1 days -02:03:04"), T("-1 days +02:03:04"), T("-1 years -2 mons")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET intervalstyle;",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '-1 2:03:04';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("-1 days +02:03:04")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Verbose interval output",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET intervalstyle = postgres_verbose;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '100000000y 10mon -1000000000d -100000h -10min -10.000001s ago'::INTERVAL;",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL)],
                        rows: &[
                            &[T("-100000000 years -10 mons +1000000000 days 100000:10:10.000001")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT INTERVAL '-10 mons -3 days +03:55:06.70', INTERVAL '-1 sec', INTERVAL '1 day -0.5 sec', INTERVAL '0', INTERVAL '-1 day';",
                    expected: Expected::Rows {
                        columns: &[Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL), Column("interval", INTERVAL)],
                        rows: &[
                            &[T("-10 mons -3 days +03:55:06.7"), T("-00:00:01"), T("1 day -00:00:00.5"), T("00:00:00"), T("-1 days")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Extreme stored intervals",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE interval_extremes (f1 INTERVAL);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO interval_extremes VALUES ('2147483647 days -2147483648 months'), ('-2147483648 days -2147483648 months'), ('1 year'), ('-178000000 years');",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM interval_extremes ORDER BY f1;",
                    expected: Expected::Rows {
                        columns: &[Column("f1", INTERVAL)],
                        rows: &[
                            &[T("-178956970 years -8 mons -2147483648 days")],
                            &[T("-178000000 years")],
                            &[T("-178956970 years -8 mons +2147483647 days")],
                            &[T("1 year")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r1.f1, r2.f1 FROM interval_extremes r1, interval_extremes r2 WHERE r1.f1 > r2.f1 ORDER BY r1.f1, r2.f1;",
                    expected: Expected::Rows {
                        columns: &[Column("f1", INTERVAL), Column("f1", INTERVAL)],
                        rows: &[
                            &[T("-178000000 years"), T("-178956970 years -8 mons -2147483648 days")],
                            &[T("-178956970 years -8 mons +2147483647 days"), T("-178956970 years -8 mons -2147483648 days")],
                            &[T("-178956970 years -8 mons +2147483647 days"), T("-178000000 years")],
                            &[T("1 year"), T("-178956970 years -8 mons -2147483648 days")],
                            &[T("1 year"), T("-178000000 years")],
                            &[T("1 year"), T("-178956970 years -8 mons +2147483647 days")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Justified interval overflow",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT justify_interval(INTERVAL '2147483647 days 24 hrs'), justify_interval(INTERVAL '-2147483648 days -24 hrs');",
                    expected: Expected::Rows {
                        columns: &[Column("justify_interval", INTERVAL), Column("justify_interval", INTERVAL)],
                        rows: &[
                            &[T("5965232 years 4 mons 8 days"), T("-5965232 years -4 mons -9 days")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT justify_hours(INTERVAL '2147483647 days 24 hrs');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "interval out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT justify_days(INTERVAL '2147483647 months 30 days');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "interval out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT justify_interval(INTERVAL '2147483647 months 30 days');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "interval out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Interval sum and avg",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT sum(x), avg(x) FROM (VALUES (INTERVAL '1 day'), (INTERVAL '2 hours'), (INTERVAL '1 mon 3 sec'), (NULL)) v(x);",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INTERVAL), Column("avg", INTERVAL)],
                        rows: &[
                            &[T("1 mon 1 day 02:00:03"), T("10 days 08:40:01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(x), avg(x) FROM (VALUES (INTERVAL '1 day')) v(x) WHERE false;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INTERVAL), Column("avg", INTERVAL)],
                        rows: &[
                            &[Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(x) FROM (VALUES (INTERVAL '2147483647 days'), (INTERVAL '1 day')) v(x);",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "interval out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
