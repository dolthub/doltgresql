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
fn test_domain() {
    run_scripts(&[
        ScriptTest {
            name: "create domain",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DOMAIN year AS integer CONSTRAINT not_null_c NOT NULL CONSTRAINT null_c  NULL;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "conflicting NULL/NOT NULL constraints", position: 62, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN year AS integer NULL NOT NULL;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "conflicting NULL/NOT NULL constraints", position: 36, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN year AS integer DEFAULT 1999 NOT NULL CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN year AS integer CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));",
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"type "year" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN year_with_check AS integer CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN year_with_two_checks AS integer CONSTRAINT year_check_min CHECK (VALUE >= 1901) CONSTRAINT year_check_max CHECK (VALUE <= 2155);",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test_table (id int primary key, v non_existing_domain);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "non_existing_domain" does not exist"#, position: 48, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname, contype, conrelid, contypid from pg_constraint WHERE conname IN ('year_check', 'year_check_min', 'year_check_max') ORDER BY conname;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("contype", CHAR), Column("conrelid", OID), Column("contypid", OID)],
                        rows: &[
                            &[T("year_check"), T("c"), T("0"), Oid(16385)],
                            &[T("year_check"), T("c"), T("0"), Oid(16389)],
                            &[T("year_check_max"), T("c"), T("0"), Oid(16392)],
                            &[T("year_check_min"), T("c"), T("0"), Oid(16392)],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "multiple checks",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DOMAIN d1 AS integer CONSTRAINT check1 CHECK (VALUE > 100) CONSTRAINT check2 CHECK (VALUE < 200);",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname, contype, conrelid, contypid from pg_constraint WHERE conname IN ('check1', 'check2') ORDER BY conname;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("contype", CHAR), Column("conrelid", OID), Column("contypid", OID)],
                        rows: &[
                            &[T("check1"), T("c"), T("0"), Oid(16385)],
                            &[T("check2"), T("c"), T("0"), Oid(16385)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "create table t1 (pk int primary key, v d1);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t1 values (1, 50);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain d1 violates check constraint "check1""#, schema: "public", data_type: "d1", constraint: "check1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t1 values (2, 150);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t1 values (3, 250);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain d1 violates check constraint "check2""#, schema: "public", data_type: "d1", constraint: "check2", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN d2 AS integer CHECK (VALUE > 300) CHECK (VALUE < 400);",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname, contype, conrelid, contypid from pg_constraint WHERE conname IN ('d2_check', 'd2_check1') ORDER BY conname;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("contype", CHAR), Column("conrelid", OID), Column("contypid", OID)],
                        rows: &[
                            &[T("d2_check"), T("c"), T("0"), Oid(16394)],
                            &[T("d2_check1"), T("c"), T("0"), Oid(16394)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN d3 AS integer CONSTRAINT d3_check1 CHECK (VALUE > 300) CHECK (VALUE < 400);",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname, contype, conrelid, contypid from pg_constraint WHERE conname IN ('d3_check1', 'd3_check') ORDER BY conname;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("contype", CHAR), Column("conrelid", OID), Column("contypid", OID)],
                        rows: &[
                            &[T("d3_check"), T("c"), T("0"), Oid(16398)],
                            &[T("d3_check1"), T("c"), T("0"), Oid(16398)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create table with domain type",
            set_up_script: &[
                "CREATE DOMAIN year AS integer CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE table_with_domain (pk int primary key, y year);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO table_with_domain VALUES (1, 1999)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO table_with_domain VALUES (2, 1899)",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain year violates check constraint "year_check""#, schema: "public", data_type: "year", constraint: "year_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM table_with_domain",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("y", INT4)],
                        rows: &[
                            &[T("1"), T("1999")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create table with domain type with default value",
            set_up_script: &[
                "CREATE DOMAIN year AS integer DEFAULT 2000;",
                "CREATE TABLE table_with_domain_with_default (pk int primary key, y year);",
                "INSERT INTO table_with_domain_with_default VALUES (1, 1999)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO table_with_domain_with_default(pk) VALUES (2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM table_with_domain_with_default",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("y", INT4)],
                        rows: &[
                            &[T("1"), T("1999")],
                            &[T("2"), T("2000")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create table with domain type with not null constraint",
            set_up_script: &[
                "CREATE DOMAIN year AS integer NOT NULL;",
                "CREATE TABLE tbl_not_null (pk int primary key, y year);",
                "INSERT INTO tbl_not_null VALUES (1, 1999)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO tbl_not_null VALUES (2, null)",
                    expected: Expected::Error(Diagnostic { code: "23502", message: "domain year does not allow null values", schema: "public", data_type: "year", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tbl_not_null(pk) VALUES (2)",
                    expected: Expected::Error(Diagnostic { code: "23502", message: "domain year does not allow null values", schema: "public", data_type: "year", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tbl_not_null",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("y", INT4)],
                        rows: &[
                            &[T("1"), T("1999")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "update on table with domain type",
            set_up_script: &[
                "CREATE DOMAIN year AS integer NOT NULL CONSTRAINT year_check_min CHECK (VALUE >= 1901) CONSTRAINT year_check_max CHECK (VALUE <= 2155);",
                "CREATE TABLE test_table (pk int primary key, y year);",
                "INSERT INTO test_table VALUES (1, 1999), (2, 2000)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test_table SET y = 1902 WHERE pk = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test_table SET y = 1900 WHERE pk = 1;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain year violates check constraint "year_check_min""#, schema: "public", data_type: "year", constraint: "year_check_min", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test_table SET y = null WHERE pk = 1;",
                    expected: Expected::Error(Diagnostic { code: "23502", message: "domain year does not allow null values", schema: "public", data_type: "year", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_table",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("y", INT4)],
                        rows: &[
                            &[T("2"), T("2000")],
                            &[T("1"), T("1902")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "domain type as text type",
            set_up_script: &[
                "CREATE DOMAIN non_empty_string AS text NULL CONSTRAINT name_check CHECK (VALUE <> '');",
                "CREATE TABLE non_empty_string_t (id int primary key, first_name non_empty_string, last_name non_empty_string);",
                "INSERT INTO non_empty_string_t VALUES (1, 'John', 'Doe')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO non_empty_string_t VALUES (2, 'Jane', 'Doe')",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE non_empty_string_t SET last_name = '' WHERE first_name = 'Jane'",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain non_empty_string violates check constraint "name_check""#, schema: "public", data_type: "non_empty_string", constraint: "name_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE non_empty_string_t SET last_name = NULL WHERE first_name = 'Jane'",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM non_empty_string_t",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("first_name", TEXT), Column("last_name", TEXT)],
                        rows: &[
                            &[T("1"), T("John"), T("Doe")],
                            &[T("2"), T("Jane"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop domain",
            set_up_script: &[
                "CREATE DOMAIN year AS integer CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));",
                "CREATE TABLE table_with_domain (pk int primary key, y year);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP DOMAIN year;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop type year because other objects depend on it", detail: "column y of table table_with_domain depends on type year", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE table_with_domain;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN year;",
                    expected: Expected::Tag("DROP DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN IF EXISTS year;",
                    expected: Expected::Tag("DROP DOMAIN"),
                    notices: &[Diagnostic { code: "00000", message: r#"type "year" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN IF EXISTS postgres.public.year;",
                    expected: Expected::Tag("DROP DOMAIN"),
                    notices: &[Diagnostic { code: "00000", message: r#"type "postgres.public.year" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN IF EXISTS mydb.public.year;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "cross-database references are not implemented: mydb.public.year", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN non_existing_domain;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "non_existing_domain" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "domain array type column",
            set_up_script: &[
                "CREATE DOMAIN vc4 AS varchar(4);",
                "CREATE TABLE t (pk int primary key, v vc4[]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, array['ab', 'cd']::vc4[]);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (2, '{ef,gh}');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("{ab,cd}")],
                            &[T("2"), T("{ef,gh}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "composite type with a domain attribute",
            set_up_script: &[
                "CREATE DOMAIN ct_posint AS int4 CHECK (VALUE > 0);",
                "CREATE TYPE ct_comp AS (f1 ct_posint, f2 text);",
                "CREATE TABLE ct_test (id int PRIMARY KEY, v ct_comp);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO ct_test VALUES (1, '(5,abc)');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v FROM ct_test ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(5,abc)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (v).f1, (v).f2 FROM ct_test ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("f1", INT4), Column("f2", TEXT)],
                        rows: &[
                            &[T("5"), T("abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "explicit cast to domain type",
            set_up_script: &[
                "CREATE DOMAIN year_not_null AS integer NOT NULL CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));",
                "CREATE TABLE test_table (year integer);",
                "INSERT INTO test_table VALUES (2000), (2024);",
                "CREATE TABLE my_table (id integer);",
                "INSERT INTO my_table VALUES (2000), (2002);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1903::year_not_null;",
                    expected: Expected::Rows {
                        columns: &[Column("year_not_null", INT4)],
                        rows: &[
                            &[T("1903")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1903::year_not_null::text;",
                    expected: Expected::Rows {
                        columns: &[Column("text", TEXT)],
                        rows: &[
                            &[T("1903")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1900::year_not_null;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain year_not_null violates check constraint "year_check""#, schema: "public", data_type: "year_not_null", constraint: "year_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL::year_not_null;",
                    expected: Expected::Error(Diagnostic { code: "23502", message: "domain year_not_null does not allow null values", schema: "public", data_type: "year_not_null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT year::year_not_null from test_table order by year;",
                    expected: Expected::Rows {
                        columns: &[Column("year", INT4)],
                        rows: &[
                            &[T("2000")],
                            &[T("2024")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test_table VALUES (null);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT year::year_not_null from test_table;",
                    expected: Expected::Error(Diagnostic { code: "23502", message: "domain year_not_null does not allow null values", schema: "public", data_type: "year_not_null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id::year_not_null from my_table order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2000")],
                            &[T("2002")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO my_table VALUES (2156);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id::year_not_null from my_table;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain year_not_null violates check constraint "year_check""#, schema: "public", data_type: "year_not_null", constraint: "year_check", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "domains over array types",
            set_up_script: &[
                "CREATE DOMAIN domint4arr AS INT4[];",
                "CREATE DOMAIN domvarchar4arr AS VARCHAR(4)[2][3];",
                "CREATE TABLE domarr (pk INT PRIMARY KEY, i domint4arr, v domvarchar4arr);",
                "INSERT INTO domarr VALUES (1, '{3,4}', '{{a,b},{c,d}}'), (2, NULL, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '{1,2}'::domint4arr, '{{a},{b}}'::domvarchar4arr;",
                    expected: Expected::Rows {
                        columns: &[Column("domint4arr", INT4_ARRAY), Column("domvarchar4arr", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{1,2}"), T("{{a},{b}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM domarr ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("i", INT4_ARRAY), Column("v", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("1"), T("{3,4}"), T("{{a,b},{c,d}}")],
                            &[T("2"), Null, Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT i[2], v[2][1], pg_typeof(i[2]) FROM domarr ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("v", VARCHAR), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("4"), T("c"), T("integer")],
                            &[Null, Null, T("integer")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE domarr SET i = '{{1,2},{3,4}}' WHERE pk = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT i, array_ndims(i) FROM domarr ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4_ARRAY), Column("array_ndims", INT4)],
                        rows: &[
                            &[T("{3,4}"), T("1")],
                            &[T("{{1,2},{3,4}}"), T("2")],
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
