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
fn test_info_schema_columns() {
    run_scripts(&[
        ScriptTest {
            name: "information_schema.columns",
            set_up_script: &[
                "create table test_table (id int primary key, col1 varchar(255));",
                "create view test_view as select * from test_table;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT DISTINCT table_schema FROM information_schema.columns ORDER BY table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_schema", NAME)],
                        rows: &[
                            &[T("information_schema")],
                            &[T("pg_catalog")],
                            &[T("public")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema, table_name, column_name FROM information_schema.columns WHERE table_schema='public' ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME), Column("column_name", NAME)],
                        rows: &[
                            &[T("postgres"), T("public"), T("test_table"), T("id")],
                            &[T("postgres"), T("public"), T("test_table"), T("col1")],
                            &[T("postgres"), T("public"), T("test_view"), T("id")],
                            &[T("postgres"), T("public"), T("test_view"), T("col1")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 
						columns.column_name, 
						pg_catalog.col_description(('"' || table_catalog || '"."' || table_schema || '"."' || table_name || '"')::regclass::oid, ordinal_position) AS description, 
						('"' || "udt_schema" || '"."' || "udt_name" || '"')::"regtype" AS "regtype", 
						pg_catalog.format_type("col_attr"."atttypid", "col_attr"."atttypmod") AS "format_type" 
						FROM "information_schema"."columns" 
						LEFT JOIN "pg_catalog"."pg_attribute" AS "col_attr" 
						ON "col_attr"."attname" = "columns"."column_name" AND "col_attr"."attrelid" = ( 
							SELECT "cls"."oid" FROM "pg_catalog"."pg_class" AS "cls" 
							LEFT JOIN "pg_catalog"."pg_namespace" AS "ns" ON "ns"."oid" = "cls"."relnamespace" 
							WHERE "cls"."relname" = "columns"."table_name" AND "ns"."nspname" = "columns"."table_schema" 
						) WHERE ("table_schema" = 'public' AND "table_name" = 'test_table');"#,
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME), Column("description", TEXT), Column("regtype", REGTYPE), Column("format_type", TEXT)],
                        rows: &[
                            &[T("id"), Null, T("integer"), T("integer")],
                            &[T("col1"), Null, T("character varying"), T("character varying(255)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA test_schema;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SEARCH_PATH TO test_schema;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test_table2 (id2 INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT table_schema FROM information_schema.columns order by table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_schema", NAME)],
                        rows: &[
                            &[T("information_schema")],
                            &[T("pg_catalog")],
                            &[T("public")],
                            &[T("test_schema")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema, table_name, column_name FROM information_schema.columns WHERE table_schema='test_schema';",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME), Column("column_name", NAME)],
                        rows: &[
                            &[T("postgres"), T("test_schema"), T("test_table2"), T("id2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM information_schema.columns WHERE table_name='test_table';",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME), Column("column_name", NAME), Column("ordinal_position", INT4), Column("column_default", VARCHAR), Column("is_nullable", VARCHAR), Column("data_type", VARCHAR), Column("character_maximum_length", INT4), Column("character_octet_length", INT4), Column("numeric_precision", INT4), Column("numeric_precision_radix", INT4), Column("numeric_scale", INT4), Column("datetime_precision", INT4), Column("interval_type", VARCHAR), Column("interval_precision", INT4), Column("character_set_catalog", NAME), Column("character_set_schema", NAME), Column("character_set_name", NAME), Column("collation_catalog", NAME), Column("collation_schema", NAME), Column("collation_name", NAME), Column("domain_catalog", NAME), Column("domain_schema", NAME), Column("domain_name", NAME), Column("udt_catalog", NAME), Column("udt_schema", NAME), Column("udt_name", NAME), Column("scope_catalog", NAME), Column("scope_schema", NAME), Column("scope_name", NAME), Column("maximum_cardinality", INT4), Column("dtd_identifier", NAME), Column("is_self_referencing", VARCHAR), Column("is_identity", VARCHAR), Column("identity_generation", VARCHAR), Column("identity_start", VARCHAR), Column("identity_increment", VARCHAR), Column("identity_maximum", VARCHAR), Column("identity_minimum", VARCHAR), Column("identity_cycle", VARCHAR), Column("is_generated", VARCHAR), Column("generation_expression", VARCHAR), Column("is_updatable", VARCHAR)],
                        rows: &[
                            &[T("postgres"), T("public"), T("test_table"), T("id"), T("1"), Null, T("NO"), T("integer"), Null, Null, T("32"), T("2"), T("0"), Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, T("postgres"), T("pg_catalog"), T("int4"), Null, Null, Null, Null, T("1"), T("NO"), T("NO"), Null, Null, Null, Null, Null, T("NO"), T("NEVER"), Null, T("YES")],
                            &[T("postgres"), T("public"), T("test_table"), T("col1"), T("2"), Null, T("YES"), T("character varying"), T("255"), T("1020"), Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, T("postgres"), T("pg_catalog"), T("varchar"), Null, Null, Null, Null, T("2"), T("NO"), T("NO"), Null, Null, Null, Null, Null, T("NO"), T("NEVER"), Null, T("YES")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM information_schema.columns WHERE table_name='test_view';",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME), Column("column_name", NAME), Column("ordinal_position", INT4), Column("column_default", VARCHAR), Column("is_nullable", VARCHAR), Column("data_type", VARCHAR), Column("character_maximum_length", INT4), Column("character_octet_length", INT4), Column("numeric_precision", INT4), Column("numeric_precision_radix", INT4), Column("numeric_scale", INT4), Column("datetime_precision", INT4), Column("interval_type", VARCHAR), Column("interval_precision", INT4), Column("character_set_catalog", NAME), Column("character_set_schema", NAME), Column("character_set_name", NAME), Column("collation_catalog", NAME), Column("collation_schema", NAME), Column("collation_name", NAME), Column("domain_catalog", NAME), Column("domain_schema", NAME), Column("domain_name", NAME), Column("udt_catalog", NAME), Column("udt_schema", NAME), Column("udt_name", NAME), Column("scope_catalog", NAME), Column("scope_schema", NAME), Column("scope_name", NAME), Column("maximum_cardinality", INT4), Column("dtd_identifier", NAME), Column("is_self_referencing", VARCHAR), Column("is_identity", VARCHAR), Column("identity_generation", VARCHAR), Column("identity_start", VARCHAR), Column("identity_increment", VARCHAR), Column("identity_maximum", VARCHAR), Column("identity_minimum", VARCHAR), Column("identity_cycle", VARCHAR), Column("is_generated", VARCHAR), Column("generation_expression", VARCHAR), Column("is_updatable", VARCHAR)],
                        rows: &[
                            &[T("postgres"), T("public"), T("test_view"), T("id"), T("1"), Null, T("YES"), T("integer"), Null, Null, T("32"), T("2"), T("0"), Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, T("postgres"), T("pg_catalog"), T("int4"), Null, Null, Null, Null, T("1"), T("NO"), T("NO"), Null, Null, Null, Null, Null, T("NO"), T("NEVER"), Null, T("YES")],
                            &[T("postgres"), T("public"), T("test_view"), T("col1"), T("2"), Null, T("YES"), T("character varying"), T("255"), T("1020"), Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, Null, T("postgres"), T("pg_catalog"), T("varchar"), Null, Null, Null, Null, T("2"), T("NO"), T("NO"), Null, Null, Null, Null, Null, T("NO"), T("NEVER"), Null, T("YES")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT columns.table_name, columns.column_name from "information_schema"."columns" WHERE table_name='test_table';"#,
                    expected: Expected::Rows {
                        columns: &[Column("table_name", NAME), Column("column_name", NAME)],
                        rows: &[
                            &[T("test_table"), T("id")],
                            &[T("test_table"), T("col1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE testnumtypes (id INT PRIMARY KEY, col1 SMALLINT, col2 BIGINT, col3 REAL, col4 DOUBLE PRECISION, col5 NUMERIC, col6 DECIMAL(10, 2), col7 OID, col8 XID);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT column_name, ordinal_position, data_type, udt_name, numeric_precision, numeric_precision_radix, numeric_scale FROM information_schema.columns WHERE table_name='testnumtypes' ORDER BY ordinal_position ASC;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME), Column("ordinal_position", INT4), Column("data_type", VARCHAR), Column("udt_name", NAME), Column("numeric_precision", INT4), Column("numeric_precision_radix", INT4), Column("numeric_scale", INT4)],
                        rows: &[
                            &[T("id"), T("1"), T("integer"), T("int4"), T("32"), T("2"), T("0")],
                            &[T("col1"), T("2"), T("smallint"), T("int2"), T("16"), T("2"), T("0")],
                            &[T("col2"), T("3"), T("bigint"), T("int8"), T("64"), T("2"), T("0")],
                            &[T("col3"), T("4"), T("real"), T("float4"), T("24"), T("2"), Null],
                            &[T("col4"), T("5"), T("double precision"), T("float8"), T("53"), T("2"), Null],
                            &[T("col5"), T("6"), T("numeric"), T("numeric"), Null, T("10"), Null],
                            &[T("col6"), T("7"), T("numeric"), T("numeric"), T("10"), T("10"), T("2")],
                            &[T("col7"), T("8"), T("oid"), T("oid"), Null, Null, Null],
                            &[T("col8"), T("9"), T("xid"), T("xid"), Null, Null, Null],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE TABLE teststringtypes (id INT PRIMARY KEY, col1 CHAR(10), col2 VARCHAR(10), col3 TEXT, col4 "char", col5 CHARACTER, col6 VARCHAR, col7 UUID);"#,
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT column_name, ordinal_position, data_type, udt_name, character_maximum_length, character_octet_length FROM information_schema.columns WHERE table_name='teststringtypes' ORDER BY ordinal_position ASC;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME), Column("ordinal_position", INT4), Column("data_type", VARCHAR), Column("udt_name", NAME), Column("character_maximum_length", INT4), Column("character_octet_length", INT4)],
                        rows: &[
                            &[T("id"), T("1"), T("integer"), T("int4"), Null, Null],
                            &[T("col1"), T("2"), T("character"), T("bpchar"), T("10"), T("40")],
                            &[T("col2"), T("3"), T("character varying"), T("varchar"), T("10"), T("40")],
                            &[T("col3"), T("4"), T("text"), T("text"), Null, T("1073741824")],
                            &[T("col4"), T("5"), T(r#""char""#), T("char"), Null, Null],
                            &[T("col5"), T("6"), T("character"), T("bpchar"), T("1"), T("4")],
                            &[T("col6"), T("7"), T("character varying"), T("varchar"), Null, T("1073741824")],
                            &[T("col7"), T("8"), T("uuid"), T("uuid"), Null, Null],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE testtimetypes (id INT PRIMARY KEY, col1 DATE, col2 TIME, col3 TIMESTAMP, col4 TIMESTAMPTZ,  col5 TIMETZ);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT column_name, ordinal_position, data_type, datetime_precision FROM information_schema.columns WHERE table_name='testtimetypes' ORDER BY ordinal_position ASC;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME), Column("ordinal_position", INT4), Column("data_type", VARCHAR), Column("datetime_precision", INT4)],
                        rows: &[
                            &[T("id"), T("1"), T("integer"), Null],
                            &[T("col1"), T("2"), T("date"), T("0")],
                            &[T("col2"), T("3"), T("time without time zone"), T("6")],
                            &[T("col3"), T("4"), T("timestamp without time zone"), T("6")],
                            &[T("col4"), T("5"), T("timestamp with time zone"), T("6")],
                            &[T("col5"), T("6"), T("time with time zone"), T("6")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT p.oid AS oid, p.relname AS table_name, n.nspname as table_schema FROM pg_class AS p JOIN pg_namespace AS n ON p.relnamespace=n.oid WHERE (n.nspname='public' AND p.relkind='r') AND left(relname, 5) <> 'dolt_';",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("table_name", NAME), Column("table_schema", NAME)],
                        rows: &[
                            &[Oid(16384), T("test_table"), T("public")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select col_description(2957635223, ordinal_position) as comment from information_schema.columns limit 1;",
                    expected: Expected::Rows {
                        columns: &[Column("comment", TEXT)],
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
            name: "generation_expression",
            set_up_script: &[
                "CREATE TABLE t3328_issue (a INT, b INT GENERATED ALWAYS AS (a + 1) STORED);",
                "INSERT INTO t3328_issue (a) VALUES (1);",
                "CREATE TABLE t3328 (a INT PRIMARY KEY, s TEXT, b INT GENERATED ALWAYS AS (a + 1) STORED, c TEXT GENERATED ALWAYS AS (upper(s)) STORED, e INT GENERATED ALWAYS AS (a) STORED, f TEXT GENERATED ALWAYS AS (s || ')') STORED);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t3328_issue;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT is_generated, generation_expression, column_default FROM information_schema.columns WHERE table_name = 't3328_issue' AND column_name = 'b';",
                    expected: Expected::Rows {
                        columns: &[Column("is_generated", VARCHAR), Column("generation_expression", VARCHAR), Column("column_default", VARCHAR)],
                        rows: &[
                            &[T("ALWAYS"), T("(a + 1)"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT column_name, is_generated, generation_expression, column_default FROM information_schema.columns WHERE table_name = 't3328' ORDER BY ordinal_position;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME), Column("is_generated", VARCHAR), Column("generation_expression", VARCHAR), Column("column_default", VARCHAR)],
                        rows: &[
                            &[T("a"), T("NEVER"), Null, Null],
                            &[T("s"), T("NEVER"), Null, Null],
                            &[T("b"), T("ALWAYS"), T("(a + 1)"), Null],
                            &[T("c"), T("ALWAYS"), T("upper(s)"), Null],
                            &[T("e"), T("ALWAYS"), T("a"), Null],
                            &[T("f"), T("ALWAYS"), T("(s || ')'::text)"), Null],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xml columns",
            set_up_script: &[
                "CREATE TABLE t3337 (id INT PRIMARY KEY, doc pg_catalog.xml, docs xml[]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT column_name, data_type, udt_schema, udt_name, character_maximum_length FROM information_schema.columns WHERE table_name = 't3337' ORDER BY ordinal_position;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME), Column("data_type", VARCHAR), Column("udt_schema", NAME), Column("udt_name", NAME), Column("character_maximum_length", INT4)],
                        rows: &[
                            &[T("id"), T("integer"), T("pg_catalog"), T("int4"), Null],
                            &[T("doc"), T("xml"), T("pg_catalog"), T("xml"), Null],
                            &[T("docs"), T("ARRAY"), T("pg_catalog"), T("_xml"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "literal column defaults",
            set_up_script: &[
                "CREATE TABLE t3432 (i0 INTEGER NOT NULL DEFAULT 0, b0 BIGINT NOT NULL DEFAULT 0, s42 SMALLINT NOT NULL DEFAULT 42, d15 DOUBLE PRECISION NOT NULL DEFAULT 1.5, r0 REAL NOT NULL DEFAULT 0.0, bp BIGINT NOT NULL DEFAULT (0), n NUMERIC DEFAULT 1.5, bo BOOLEAN DEFAULT true, f8i DOUBLE PRECISION DEFAULT 3, t TEXT DEFAULT 'abc', v VARCHAR(10) DEFAULT 'x', i5 INTEGER DEFAULT '5', b5 BIGINT DEFAULT '5', d DATE DEFAULT '2020-01-01', bs BOOLEAN DEFAULT 'true', q TEXT DEFAULT 'it''s', bn BIGINT DEFAULT -5, big BIGINT DEFAULT 5000000000, nn NUMERIC DEFAULT -1.5, ni NUMERIC DEFAULT 7, tn TEXT DEFAULT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT column_name, column_default FROM information_schema.columns WHERE table_name = 't3432' ORDER BY ordinal_position;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME), Column("column_default", VARCHAR)],
                        rows: &[
                            &[T("i0"), T("0")],
                            &[T("b0"), T("0")],
                            &[T("s42"), T("42")],
                            &[T("d15"), T("1.5")],
                            &[T("r0"), T("0.0")],
                            &[T("bp"), T("0")],
                            &[T("n"), T("1.5")],
                            &[T("bo"), T("true")],
                            &[T("f8i"), T("3")],
                            &[T("t"), T("'abc'::text")],
                            &[T("v"), T("'x'::character varying")],
                            &[T("i5"), T("5")],
                            &[T("b5"), T("'5'::bigint")],
                            &[T("d"), T("'2020-01-01'::date")],
                            &[T("bs"), T("true")],
                            &[T("q"), T("'it''s'::text")],
                            &[T("bn"), T("'-5'::integer")],
                            &[T("big"), T("'5000000000'::bigint")],
                            &[T("nn"), T("'-1.5'::numeric")],
                            &[T("ni"), T("7")],
                            &[T("tn"), Null],
                        ],
                        tag: "SELECT 21",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "cast column defaults",
            set_up_script: &[
                "CREATE TABLE t3432_casts (bc0 BIGINT DEFAULT CAST(0 AS BIGINT), bc1 BIGINT DEFAULT 0::bigint, bc2 INTEGER DEFAULT 0::integer, bc3 INTEGER DEFAULT CAST(7 AS BIGINT), tc TEXT DEFAULT 'a'::text, tv VARCHAR(10) DEFAULT CAST('x' AS VARCHAR(10)), nc NUMERIC DEFAULT 1.5::numeric, fc REAL DEFAULT 1.5::real, ec INTEGER DEFAULT (1 + 2)::integer, ng BIGINT DEFAULT (-5)::bigint, nt TEXT DEFAULT (-5)::text, nm NUMERIC(3,1) DEFAULT CAST(1.5 AS NUMERIC(3,1)));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT column_name, column_default FROM information_schema.columns WHERE table_name = 't3432_casts' ORDER BY ordinal_position;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME), Column("column_default", VARCHAR)],
                        rows: &[
                            &[T("bc0"), T("(0)::bigint")],
                            &[T("bc1"), T("(0)::bigint")],
                            &[T("bc2"), T("0")],
                            &[T("bc3"), T("(7)::bigint")],
                            &[T("tc"), T("'a'::text")],
                            &[T("tv"), T("'x'::character varying(10)")],
                            &[T("nc"), T("1.5")],
                            &[T("fc"), T("(1.5)::real")],
                            &[T("ec"), T("(1 + 2)")],
                            &[T("ng"), T("('-5'::integer)::bigint")],
                            &[T("nt"), T("('-5'::integer)::text")],
                            &[T("nm"), T("1.5::numeric(3,1)")],
                        ],
                        tag: "SELECT 12",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3432_casts (bc0) VALUES (DEFAULT);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ng, nt FROM t3432_casts;",
                    expected: Expected::Rows {
                        columns: &[Column("ng", INT8), Column("nt", TEXT)],
                        rows: &[
                            &[T("-5"), T("-5")],
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
fn test_info_schema_pg_char_max_length() {
    run_scripts(&[
        ScriptTest {
            name: "information_schema._pg_char_max_length",
            skip: Some(r#"setup fails on Postgres ("conversion: unsupported bind variable: {\"$type\":\"uint32\",\"$value\":1043}") and on the Go server ("conversion: unsupported bind variable: {\"$type\":\"uint32\",\"$value\":1043}")"#),
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(1043::oid, 14);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(1042::oid, 14);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(1560::oid, 10);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(1562::oid, 10);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_char_max_length(1042::oid, -1),
						information_schema._pg_char_max_length(1043::oid, -1),
						information_schema._pg_char_max_length(1560::oid, -1),
						information_schema._pg_char_max_length(1562::oid, -1);"#,
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(NULL::oid, 14);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(1043::oid, NULL::integer);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(NULL::oid, NULL::integer);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(NULL, NULL);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_char_max_length(25::oid, 14),
						information_schema._pg_char_max_length(23::oid, 14),
						information_schema._pg_char_max_length(18::oid, 14),
						information_schema._pg_char_max_length(19::oid, 14),
						information_schema._pg_char_max_length(1015::oid, 14),
						information_schema._pg_char_max_length(0::oid, 14),
						information_schema._pg_char_max_length(999999::oid, 14);"#,
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_char_max_length(1042::oid, 0),
						information_schema._pg_char_max_length(1043::oid, 4),
						information_schema._pg_char_max_length(1043::oid, -2),
						information_schema._pg_char_max_length(1560::oid, 0),
						information_schema._pg_char_max_length(1562::oid, -2);"#,
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_char_max_length(1043::oid, 2147483647),
						information_schema._pg_char_max_length(1042::oid, '-2147483644'::integer),
						information_schema._pg_char_max_length(1560::oid, '-2147483648'::integer),
						information_schema._pg_char_max_length(1562::oid, 2147483647);"#,
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(1042::oid, '-2147483648'::integer);",
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(1043::oid, '-2147483645'::integer);",
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length(25::oid, '-2147483648'::integer);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_char_max_length('varchar'::regtype::oid, 14);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                // cannot convert: unsupported bind variable: {"$type":"uint32","$value":1043}
            ],
            ..S
        },
        ScriptTest {
            name: "information_schema._pg_char_max_length with catalog inputs",
            set_up_script: &[
                "CREATE DOMAIN char_max_length_domain AS varchar(10);",
                r#"CREATE TABLE char_max_length_columns (
					id integer PRIMARY KEY, c char(10), v varchar(10), txt text,
					default_c character, unlimited_v varchar, b bit(10), vb bit varying(10),
					unlimited_vb bit varying, internal_c "char", va varchar(10)[], d char_max_length_domain
				);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT a.attname, information_schema._pg_char_max_length(a.atttypid, a.atttypmod)
						FROM pg_attribute a
						WHERE a.attrelid = 'char_max_length_columns'::regclass AND a.attnum > 0
						ORDER BY a.attnum;"#,
                    expected: Expected::Rows {
                        columns: &[Column("attname", NAME), Column("_pg_char_max_length", INT4)],
                        rows: &[
                            &[T("id"), Null],
                            &[T("c"), T("10")],
                            &[T("v"), T("10")],
                            &[T("txt"), Null],
                            &[T("default_c"), T("1")],
                            &[T("unlimited_v"), Null],
                            &[T("b"), T("10")],
                            &[T("vb"), T("10")],
                            &[T("unlimited_vb"), Null],
                            &[T("internal_c"), Null],
                            &[T("va"), Null],
                            &[T("d"), Null],
                        ],
                        tag: "SELECT 12",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_info_schema_pg_true_typ_id() {
    run_scripts(&[
        ScriptTest {
            name: "information_schema._pg_truetypid catalog reproduction",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(a.*, t.*)
						FROM pg_catalog.pg_attribute a
						JOIN pg_catalog.pg_type t ON a.atttypid = t.oid
						WHERE a.attrelid = 'pg_catalog.pg_class'::regclass AND a.attname = 'relname';"#,
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[T("19")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(pg_attribute.*, pg_type.*)
						FROM pg_catalog.pg_attribute
						JOIN pg_catalog.pg_type ON pg_attribute.atttypid = pg_type.oid
						WHERE pg_attribute.attrelid = 'pg_catalog.pg_class'::regclass AND pg_attribute.attname = 'relname';"#,
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[T("19")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_truetypid(23::oid, 23::oid);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function information_schema._pg_truetypid(oid, oid) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(t.*, a.*)
						FROM pg_catalog.pg_attribute a CROSS JOIN pg_catalog.pg_type t;"#,
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function information_schema._pg_truetypid(pg_type, pg_attribute) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_truetypid(NULL::pg_catalog.pg_attribute, NULL::pg_catalog.pg_type);",
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(typname) FROM pg_catalog.pg_type typname WHERE oid = 23::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "information_schema._pg_truetypid types and domains",
            set_up_script: &[
                "CREATE TYPE truetypid_mood AS ENUM ('sad', 'happy');",
                "CREATE DOMAIN truetypid_integer_domain AS integer;",
                "CREATE DOMAIN truetypid_varchar_domain AS varchar(10);",
                "CREATE DOMAIN truetypid_array_domain AS integer[];",
                "CREATE DOMAIN truetypid_nested_domain AS truetypid_varchar_domain;",
                r#"CREATE TABLE truetypid_columns (
					id integer PRIMARY KEY, c char(10), v varchar(10), txt text,
					b bit(10), vb bit varying(10), va varchar(10)[], internal_c "char",
					mood truetypid_mood, di truetypid_integer_domain, dv truetypid_varchar_domain,
					da truetypid_array_domain, dn truetypid_nested_domain, ad truetypid_integer_domain[]
				);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT a.attname, information_schema._pg_truetypid(a.*, t.*)
						FROM pg_catalog.pg_attribute a
						JOIN pg_catalog.pg_type t ON a.atttypid = t.oid
						WHERE a.attrelid = 'truetypid_columns'::regclass AND a.attnum BETWEEN 1 AND 8
						ORDER BY a.attnum;"#,
                    expected: Expected::Rows {
                        columns: &[Column("attname", NAME), Column("_pg_truetypid", OID)],
                        rows: &[
                            &[T("id"), T("23")],
                            &[T("c"), T("1042")],
                            &[T("v"), T("1043")],
                            &[T("txt"), T("25")],
                            &[T("b"), T("1560")],
                            &[T("vb"), T("1562")],
                            &[T("va"), T("1015")],
                            &[T("internal_c"), T("18")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(a.*, t.*) = t.oid
						FROM pg_catalog.pg_attribute a
						JOIN pg_catalog.pg_type t ON a.atttypid = t.oid
						WHERE a.attrelid = 'truetypid_columns'::regclass AND a.attname = 'mood';"#,
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
                    query: r#"SELECT a.attname, information_schema._pg_truetypid(a.*, t.*)
						FROM pg_catalog.pg_attribute a
						JOIN pg_catalog.pg_type t ON a.atttypid = t.oid
						WHERE a.attrelid = 'truetypid_columns'::regclass AND a.attname IN ('di', 'dv', 'da')
						ORDER BY a.attnum;"#,
                    expected: Expected::Rows {
                        columns: &[Column("attname", NAME), Column("_pg_truetypid", OID)],
                        rows: &[
                            &[T("di"), T("23")],
                            &[T("dv"), T("1043")],
                            &[T("da"), T("1007")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(a.*, t.*) = (SELECT oid FROM pg_catalog.pg_type WHERE typname = 'truetypid_varchar_domain')
						FROM pg_catalog.pg_attribute a
						JOIN pg_catalog.pg_type t ON a.atttypid = t.oid
						WHERE a.attrelid = 'truetypid_columns'::regclass AND a.attname = 'dn';"#,
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
                    query: r#"SELECT information_schema._pg_truetypid(a.*, t.*) = a.atttypid
						FROM pg_catalog.pg_attribute a
						JOIN pg_catalog.pg_type t ON a.atttypid = t.oid
						WHERE a.attrelid = 'truetypid_columns'::regclass AND a.attname = 'ad';"#,
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
                    query: r#"SELECT information_schema._pg_truetypid(a, t)
						FROM pg_catalog.pg_attribute a
						JOIN pg_catalog.pg_type t ON a.atttypid = t.oid
						WHERE a.attrelid = 'truetypid_columns'::regclass AND a.attname = 'dv';"#,
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[T("1043")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(a.*, t.*)
						FROM pg_catalog.pg_attribute a CROSS JOIN pg_catalog.pg_type t
						WHERE a.attrelid = 'truetypid_columns'::regclass AND a.attname = 'id'
						AND t.oid = 1043::oid;"#,
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[T("23")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(a.*, t.*)
						FROM pg_catalog.pg_attribute a CROSS JOIN pg_catalog.pg_type t
						WHERE a.attrelid = 'truetypid_columns'::regclass AND a.attname = 'id'
						AND t.oid = (SELECT oid FROM pg_catalog.pg_type WHERE typname = 'truetypid_varchar_domain');"#,
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[T("1043")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(NULL, t.*)
						FROM pg_catalog.pg_type t WHERE t.oid = 23::oid;"#,
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(NULL, t.*)
						FROM pg_catalog.pg_type t WHERE t.oid = (SELECT oid FROM pg_catalog.pg_type WHERE typname = 'truetypid_varchar_domain');"#,
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(a.*, NULL)
						FROM pg_catalog.pg_attribute a
						WHERE a.attrelid = 'truetypid_columns'::regclass AND a.attname = 'id';"#,
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT information_schema._pg_truetypid(NULL, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(a.*, t.*)
						FROM pg_catalog.pg_type t LEFT JOIN pg_catalog.pg_attribute a ON false
						WHERE t.typname = 'truetypid_varchar_domain';"#,
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT information_schema._pg_truetypid(a.*, t.*)
						FROM pg_catalog.pg_attribute a LEFT JOIN pg_catalog.pg_type t ON false
						WHERE a.attrelid = 'truetypid_columns'::regclass AND a.attname = 'id';"#,
                    expected: Expected::Rows {
                        columns: &[Column("_pg_truetypid", OID)],
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
fn test_info_schema_revision_db() {
    run_scripts(&[
        ScriptTest {
            name: "info_schema changes with dolt_checkout",
            set_up_script: &[
                "create table t (a int primary key, b int);",
                "select dolt_commit('-Am', 'creating table t');",
                "select dolt_branch('b2');",
                "select dolt_branch('b3');",
                "select dolt_checkout('b2');",
                "alter table t add column c int;",
                "select dolt_commit('-am', 'added column c on branch b2');",
                "select dolt_checkout('b3');",
                "alter table t add column d int;",
                "select dolt_commit('-am', 'added column d on branch b3');",
                "select dolt_checkout('main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select active_branch();",
                    expected: Expected::Rows {
                        columns: &[Column("active_branch", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows the Go server's rows with Postgres' column type.
                ScriptTestAssertion {
                    query: "select column_name from information_schema.columns where table_catalog = 'postgres' and table_name = 't' order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('b2');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select active_branch();",
                    expected: Expected::Rows {
                        columns: &[Column("active_branch", TEXT)],
                        rows: &[
                            &[T("b2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows the Go server's rows with Postgres' column type.
                ScriptTestAssertion {
                    query: "select column_name from information_schema.columns where table_catalog = 'postgres' and table_name = 't' order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                            &[T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('b3');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select active_branch();",
                    expected: Expected::Rows {
                        columns: &[Column("active_branch", TEXT)],
                        rows: &[
                            &[T("b3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows the Go server's rows with Postgres' column type.
                ScriptTestAssertion {
                    query: "select column_name from information_schema.columns where table_catalog = 'postgres' and table_name = 't' order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                            &[T("d")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "info_schema with detached HEAD",
            set_up_script: &[
                "create table t (a int primary key, b int);",
                "select dolt_commit('-Am', 'creating table t');",
                "select dolt_branch('b2');",
                "select dolt_branch('b3');",
                "select dolt_checkout('b2');",
                "alter table t add column c int;",
                "select dolt_commit('-am', 'added column c on branch b2');",
                "select dolt_tag('t2')",
                "select dolt_checkout('b3');",
                "alter table t add column d int;",
                "select dolt_commit('-am', 'added column d on branch b3');",
                "select dolt_tag('t3')",
                "select dolt_checkout('main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select active_branch();",
                    expected: Expected::Rows {
                        columns: &[Column("active_branch", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows the Go server's rows with Postgres' column type.
                ScriptTestAssertion {
                    query: "select column_name from information_schema.columns where table_catalog = 'postgres' and table_name = 't' order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "use postgres/t2;",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select active_branch();",
                    expected: Expected::Rows {
                        columns: &[Column("active_branch", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows the Go server's rows with Postgres' column type.
                ScriptTestAssertion {
                    query: "select column_name from information_schema.columns where table_catalog = 'postgres/t2' and table_name = 't' order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                            &[T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "use postgres/t3;",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select active_branch();",
                    expected: Expected::Rows {
                        columns: &[Column("active_branch", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows the Go server's rows with Postgres' column type.
                ScriptTestAssertion {
                    query: "select column_name from information_schema.columns where table_catalog = 'postgres/t3' and table_name = 't' order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                            &[T("d")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select relname from pg_class where oid = 't'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME)],
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
fn test_info_schema_schemata() {
    run_scripts(&[
        ScriptTest {
            name: "information_schema.schemata",
            database: "newdb",
            set_up_script: &[
                "create schema test_schema",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT catalog_name, schema_name FROM information_schema.schemata order by schema_name;",
                    expected: Expected::Rows {
                        columns: &[Column("catalog_name", NAME), Column("schema_name", NAME)],
                        rows: &[
                            &[T("newdb"), T("information_schema")],
                            &[T("newdb"), T("pg_catalog")],
                            &[T("newdb"), T("pg_toast")],
                            &[T("newdb"), T("public")],
                            &[T("newdb"), T("test_schema")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM information_schema.schemata order by schema_name;",
                    expected: Expected::Rows {
                        columns: &[Column("catalog_name", NAME), Column("schema_name", NAME), Column("schema_owner", NAME), Column("default_character_set_catalog", NAME), Column("default_character_set_schema", NAME), Column("default_character_set_name", NAME), Column("sql_path", VARCHAR)],
                        rows: &[
                            &[T("newdb"), T("information_schema"), T("postgres"), Null, Null, Null, Null],
                            &[T("newdb"), T("pg_catalog"), T("postgres"), Null, Null, Null, Null],
                            &[T("newdb"), T("pg_toast"), T("postgres"), Null, Null, Null, Null],
                            &[T("newdb"), T("public"), T("pg_database_owner"), Null, Null, Null, Null],
                            &[T("newdb"), T("test_schema"), T("postgres"), Null, Null, Null, Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_info_schema_sequences() {
    run_scripts(&[
        ScriptTest {
            name: "information_schema.sequences",
            set_up_script: &[
                "create sequence standard as smallint;",
                "create schema test_schema",
                "create table test_schema.test_table (id serial);",
                "create sequence big increment by 3 start with 10 minvalue 1 cycle;",
                "create sequence negative increment by -1",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from information_schema.sequences where sequence_name = 'standard';",
                    expected: Expected::Rows {
                        columns: &[Column("sequence_catalog", NAME), Column("sequence_schema", NAME), Column("sequence_name", NAME), Column("data_type", VARCHAR), Column("numeric_precision", INT4), Column("numeric_precision_radix", INT4), Column("numeric_scale", INT4), Column("start_value", VARCHAR), Column("minimum_value", VARCHAR), Column("maximum_value", VARCHAR), Column("increment", VARCHAR), Column("cycle_option", VARCHAR)],
                        rows: &[
                            &[T("postgres"), T("public"), T("standard"), T("smallint"), T("16"), T("2"), T("0"), T("1"), T("1"), T("32767"), T("1"), T("NO")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select sequence_schema,sequence_name,data_type,numeric_precision from information_schema.sequences where sequence_name = 'test_table_id_seq';",
                    expected: Expected::Rows {
                        columns: &[Column("sequence_schema", NAME), Column("sequence_name", NAME), Column("data_type", VARCHAR), Column("numeric_precision", INT4)],
                        rows: &[
                            &[T("test_schema"), T("test_table_id_seq"), T("integer"), T("32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select sequence_name,data_type,numeric_precision,minimum_value,increment,cycle_option from information_schema.sequences where sequence_name = 'big';",
                    expected: Expected::Rows {
                        columns: &[Column("sequence_name", NAME), Column("data_type", VARCHAR), Column("numeric_precision", INT4), Column("minimum_value", VARCHAR), Column("increment", VARCHAR), Column("cycle_option", VARCHAR)],
                        rows: &[
                            &[T("big"), T("bigint"), T("64"), T("1"), T("3"), T("YES")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select sequence_name, increment from information_schema.sequences where sequence_name = 'negative';",
                    expected: Expected::Rows {
                        columns: &[Column("sequence_name", NAME), Column("increment", VARCHAR)],
                        rows: &[
                            &[T("negative"), T("-1")],
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
fn test_info_schema_tables() {
    run_scripts(&[
        ScriptTest {
            name: "information_schema.tables",
            set_up_script: &[
                "create table test_table (id int)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM information_schema.tables WHERE table_name='test_table';",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME), Column("table_type", VARCHAR), Column("self_referencing_column_name", NAME), Column("reference_generation", VARCHAR), Column("user_defined_type_catalog", NAME), Column("user_defined_type_schema", NAME), Column("user_defined_type_name", NAME), Column("is_insertable_into", VARCHAR), Column("is_typed", VARCHAR), Column("commit_action", VARCHAR)],
                        rows: &[
                            &[T("postgres"), T("public"), T("test_table"), T("BASE TABLE"), Null, Null, Null, Null, Null, T("YES"), T("NO"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT table_schema FROM information_schema.tables order by table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_schema", NAME)],
                        rows: &[
                            &[T("information_schema")],
                            &[T("pg_catalog")],
                            &[T("public")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema FROM information_schema.tables group by table_catalog, table_schema order by table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME)],
                        rows: &[
                            &[T("postgres"), T("information_schema")],
                            &[T("postgres"), T("pg_catalog")],
                            &[T("postgres"), T("public")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema, table_name FROM information_schema.tables WHERE table_schema='public';",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME)],
                        rows: &[
                            &[T("postgres"), T("public"), T("test_table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA test_schema;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SEARCH_PATH TO test_schema;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test_table2 (id INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT table_schema FROM information_schema.tables order by table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_schema", NAME)],
                        rows: &[
                            &[T("information_schema")],
                            &[T("pg_catalog")],
                            &[T("public")],
                            &[T("test_schema")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema FROM information_schema.tables group by table_catalog, table_schema order by table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME)],
                        rows: &[
                            &[T("postgres"), T("information_schema")],
                            &[T("postgres"), T("pg_catalog")],
                            &[T("postgres"), T("public")],
                            &[T("postgres"), T("test_schema")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema, table_name FROM information_schema.tables WHERE table_schema='test_schema';",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME)],
                        rows: &[
                            &[T("postgres"), T("test_schema"), T("test_table2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema, table_name, table_type FROM information_schema.tables WHERE table_schema = 'test_schema' ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME), Column("table_type", VARCHAR)],
                        rows: &[
                            &[T("postgres"), T("test_schema"), T("test_table2"), T("BASE TABLE")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT "table_schema", "table_name", obj_description(('"' || "table_schema" || '"."' || "table_name" || '"')::regclass, 'pg_class') AS table_comment FROM "information_schema"."tables" WHERE ("table_schema" = 'test_schema' AND "table_name" = 'test_table2')"#,
                    expected: Expected::Rows {
                        columns: &[Column("table_schema", NAME), Column("table_name", NAME), Column("table_comment", TEXT)],
                        rows: &[
                            &[T("test_schema"), T("test_table2"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW test_view AS SELECT * FROM test_table2;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema, table_name, table_type  FROM information_schema.tables WHERE table_schema='test_schema' OR table_schema='public';",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME), Column("table_type", VARCHAR)],
                        rows: &[
                            &[T("postgres"), T("public"), T("test_table"), T("BASE TABLE")],
                            &[T("postgres"), T("test_schema"), T("test_table2"), T("BASE TABLE")],
                            &[T("postgres"), T("test_schema"), T("test_view"), T("VIEW")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_info_schema_triggers() {
    run_scripts(&[
        ScriptTest {
            name: "information_schema.triggers",
            set_up_script: &[
                "CREATE TABLE t3330_issue (a INT);",
                "CREATE FUNCTION report_row() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE NOTICE 'trigger fired for a = %', NEW.a; RETURN NEW; END; $$;",
                "CREATE TRIGGER t_report BEFORE INSERT ON t3330_issue FOR EACH ROW EXECUTE FUNCTION report_row();",
                "INSERT INTO t3330_issue VALUES (1);",
                "CREATE TABLE t3330 (a INT PRIMARY KEY, b INT);",
                "CREATE FUNCTION f3330() RETURNS trigger AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER tr_b BEFORE INSERT OR UPDATE ON t3330 FOR EACH ROW EXECUTE FUNCTION f3330();",
                "CREATE TRIGGER tr_a AFTER INSERT ON t3330 FOR EACH ROW EXECUTE FUNCTION f3330('x', 'y''z');",
                "CREATE TRIGGER tr_c BEFORE INSERT ON t3330 FOR EACH ROW EXECUTE FUNCTION f3330();",
                "CREATE TRIGGER tr_d BEFORE DELETE OR UPDATE ON t3330 FOR EACH ROW WHEN (old.b > 1) EXECUTE FUNCTION f3330();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT trigger_name, event_manipulation, event_object_table FROM information_schema.triggers WHERE event_object_table = 't3330_issue';",
                    expected: Expected::Rows {
                        columns: &[Column("trigger_name", NAME), Column("event_manipulation", VARCHAR), Column("event_object_table", NAME)],
                        rows: &[
                            &[T("t_report"), T("INSERT"), T("t3330_issue")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT trigger_catalog, trigger_schema, trigger_name, event_manipulation, event_object_catalog, event_object_schema, event_object_table, action_order, action_condition, action_statement, action_orientation, action_timing, action_reference_old_table, action_reference_new_table, action_reference_old_row, action_reference_new_row, created FROM information_schema.triggers WHERE event_object_table = 't3330' ORDER BY trigger_name, event_manipulation;",
                    expected: Expected::Rows {
                        columns: &[Column("trigger_catalog", NAME), Column("trigger_schema", NAME), Column("trigger_name", NAME), Column("event_manipulation", VARCHAR), Column("event_object_catalog", NAME), Column("event_object_schema", NAME), Column("event_object_table", NAME), Column("action_order", INT4), Column("action_condition", VARCHAR), Column("action_statement", VARCHAR), Column("action_orientation", VARCHAR), Column("action_timing", VARCHAR), Column("action_reference_old_table", NAME), Column("action_reference_new_table", NAME), Column("action_reference_old_row", NAME), Column("action_reference_new_row", NAME), Column("created", TIMESTAMPTZ)],
                        rows: &[
                            &[T("postgres"), T("public"), T("tr_a"), T("INSERT"), T("postgres"), T("public"), T("t3330"), T("1"), Null, T("EXECUTE FUNCTION f3330('x', 'y''z')"), T("ROW"), T("AFTER"), Null, Null, Null, Null, Null],
                            &[T("postgres"), T("public"), T("tr_b"), T("INSERT"), T("postgres"), T("public"), T("t3330"), T("1"), Null, T("EXECUTE FUNCTION f3330()"), T("ROW"), T("BEFORE"), Null, Null, Null, Null, Null],
                            &[T("postgres"), T("public"), T("tr_b"), T("UPDATE"), T("postgres"), T("public"), T("t3330"), T("1"), Null, T("EXECUTE FUNCTION f3330()"), T("ROW"), T("BEFORE"), Null, Null, Null, Null, Null],
                            &[T("postgres"), T("public"), T("tr_c"), T("INSERT"), T("postgres"), T("public"), T("t3330"), T("2"), Null, T("EXECUTE FUNCTION f3330()"), T("ROW"), T("BEFORE"), Null, Null, Null, Null, Null],
                            &[T("postgres"), T("public"), T("tr_d"), T("DELETE"), T("postgres"), T("public"), T("t3330"), T("1"), T("(old.b > 1)"), T("EXECUTE FUNCTION f3330()"), T("ROW"), T("BEFORE"), Null, Null, Null, Null, Null],
                            &[T("postgres"), T("public"), T("tr_d"), T("UPDATE"), T("postgres"), T("public"), T("t3330"), T("2"), T("(old.b > 1)"), T("EXECUTE FUNCTION f3330()"), T("ROW"), T("BEFORE"), Null, Null, Null, Null, Null],
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
fn test_info_schema_views() {
    run_scripts(&[
        ScriptTest {
            name: "information_schema.views",
            set_up_script: &[
                "create table test_table (id int)",
                "create view test_view as select * from test_table",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM information_schema.views order by table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME), Column("view_definition", VARCHAR), Column("check_option", VARCHAR), Column("is_updatable", VARCHAR), Column("is_insertable_into", VARCHAR), Column("is_trigger_updatable", VARCHAR), Column("is_trigger_deletable", VARCHAR), Column("is_trigger_insertable_into", VARCHAR)],
                        rows: &[
                            &[T("postgres"), T("information_schema"), T("user_mappings"), T(r#" SELECT _pg_user_mappings.authorization_identifier,
    _pg_user_mappings.foreign_server_catalog,
    _pg_user_mappings.foreign_server_name
   FROM information_schema._pg_user_mappings;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("role_column_grants"), T(r#" SELECT column_privileges.grantor,
    column_privileges.grantee,
    column_privileges.table_catalog,
    column_privileges.table_schema,
    column_privileges.table_name,
    column_privileges.column_name,
    column_privileges.privilege_type,
    column_privileges.is_grantable
   FROM information_schema.column_privileges
  WHERE (((column_privileges.grantor)::name IN ( SELECT enabled_roles.role_name
           FROM information_schema.enabled_roles)) OR ((column_privileges.grantee)::name IN ( SELECT enabled_roles.role_name
           FROM information_schema.enabled_roles)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("information_schema_catalog_name"), T(" SELECT (current_database())::information_schema.sql_identifier AS catalog_name;"), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("column_domain_usage"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS domain_catalog,
    (nt.nspname)::information_schema.sql_identifier AS domain_schema,
    (t.typname)::information_schema.sql_identifier AS domain_name,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (nc.nspname)::information_schema.sql_identifier AS table_schema,
    (c.relname)::information_schema.sql_identifier AS table_name,
    (a.attname)::information_schema.sql_identifier AS column_name
   FROM pg_type t,
    pg_namespace nt,
    pg_class c,
    pg_namespace nc,
    pg_attribute a
  WHERE ((t.typnamespace = nt.oid) AND (c.relnamespace = nc.oid) AND (a.attrelid = c.oid) AND (a.atttypid = t.oid) AND (t.typtype = 'd'::"char") AND (c.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"])) AND (a.attnum > 0) AND (NOT a.attisdropped) AND pg_has_role(t.typowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("applicable_roles"), T(r#" SELECT (a.rolname)::information_schema.sql_identifier AS grantee,
    (b.rolname)::information_schema.sql_identifier AS role_name,
    (
        CASE
            WHEN m.admin_option THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_grantable
   FROM ((( SELECT pg_auth_members.member,
            pg_auth_members.roleid,
            pg_auth_members.admin_option
           FROM pg_auth_members
        UNION
         SELECT pg_database.datdba,
            pg_authid.oid,
            false
           FROM pg_database,
            pg_authid
          WHERE ((pg_database.datname = current_database()) AND (pg_authid.rolname = 'pg_database_owner'::name))) m
     JOIN pg_authid a ON ((m.member = a.oid)))
     JOIN pg_authid b ON ((m.roleid = b.oid)))
  WHERE pg_has_role(a.oid, 'USAGE'::text);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("administrable_role_authorizations"), T(r#" SELECT applicable_roles.grantee,
    applicable_roles.role_name,
    applicable_roles.is_grantable
   FROM information_schema.applicable_roles
  WHERE ((applicable_roles.is_grantable)::text = 'YES'::text);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("domain_constraints"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS constraint_catalog,
    (rs.nspname)::information_schema.sql_identifier AS constraint_schema,
    (con.conname)::information_schema.sql_identifier AS constraint_name,
    (current_database())::information_schema.sql_identifier AS domain_catalog,
    (n.nspname)::information_schema.sql_identifier AS domain_schema,
    (t.typname)::information_schema.sql_identifier AS domain_name,
    (
        CASE
            WHEN con.condeferrable THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_deferrable,
    (
        CASE
            WHEN con.condeferred THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS initially_deferred
   FROM pg_namespace rs,
    pg_namespace n,
    pg_constraint con,
    pg_type t
  WHERE ((rs.oid = con.connamespace) AND (n.oid = t.typnamespace) AND (t.oid = con.contypid) AND (pg_has_role(t.typowner, 'USAGE'::text) OR has_type_privilege(t.oid, 'USAGE'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("attributes"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS udt_catalog,
    (nc.nspname)::information_schema.sql_identifier AS udt_schema,
    (c.relname)::information_schema.sql_identifier AS udt_name,
    (a.attname)::information_schema.sql_identifier AS attribute_name,
    (a.attnum)::information_schema.cardinal_number AS ordinal_position,
    (pg_get_expr(ad.adbin, ad.adrelid))::information_schema.character_data AS attribute_default,
    (
        CASE
            WHEN (a.attnotnull OR ((t.typtype = 'd'::"char") AND t.typnotnull)) THEN 'NO'::text
            ELSE 'YES'::text
        END)::information_schema.yes_or_no AS is_nullable,
    (
        CASE
            WHEN ((t.typelem <> (0)::oid) AND (t.typlen = '-1'::integer)) THEN 'ARRAY'::text
            WHEN (nt.nspname = 'pg_catalog'::name) THEN format_type(a.atttypid, NULL::integer)
            ELSE 'USER-DEFINED'::text
        END)::information_schema.character_data AS data_type,
    (information_schema._pg_char_max_length(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS character_maximum_length,
    (information_schema._pg_char_octet_length(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS character_octet_length,
    (NULL::name)::information_schema.sql_identifier AS character_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS character_set_schema,
    (NULL::name)::information_schema.sql_identifier AS character_set_name,
    (
        CASE
            WHEN (nco.nspname IS NOT NULL) THEN current_database()
            ELSE NULL::name
        END)::information_schema.sql_identifier AS collation_catalog,
    (nco.nspname)::information_schema.sql_identifier AS collation_schema,
    (co.collname)::information_schema.sql_identifier AS collation_name,
    (information_schema._pg_numeric_precision(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS numeric_precision,
    (information_schema._pg_numeric_precision_radix(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS numeric_precision_radix,
    (information_schema._pg_numeric_scale(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS numeric_scale,
    (information_schema._pg_datetime_precision(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS datetime_precision,
    (information_schema._pg_interval_type(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.character_data AS interval_type,
    (NULL::integer)::information_schema.cardinal_number AS interval_precision,
    (current_database())::information_schema.sql_identifier AS attribute_udt_catalog,
    (nt.nspname)::information_schema.sql_identifier AS attribute_udt_schema,
    (t.typname)::information_schema.sql_identifier AS attribute_udt_name,
    (NULL::name)::information_schema.sql_identifier AS scope_catalog,
    (NULL::name)::information_schema.sql_identifier AS scope_schema,
    (NULL::name)::information_schema.sql_identifier AS scope_name,
    (NULL::integer)::information_schema.cardinal_number AS maximum_cardinality,
    (a.attnum)::information_schema.sql_identifier AS dtd_identifier,
    ('NO'::character varying)::information_schema.yes_or_no AS is_derived_reference_attribute
   FROM ((((pg_attribute a
     LEFT JOIN pg_attrdef ad ON (((a.attrelid = ad.adrelid) AND (a.attnum = ad.adnum))))
     JOIN (pg_class c
     JOIN pg_namespace nc ON ((c.relnamespace = nc.oid))) ON ((a.attrelid = c.oid)))
     JOIN (pg_type t
     JOIN pg_namespace nt ON ((t.typnamespace = nt.oid))) ON ((a.atttypid = t.oid)))
     LEFT JOIN (pg_collation co
     JOIN pg_namespace nco ON ((co.collnamespace = nco.oid))) ON (((a.attcollation = co.oid) AND ((nco.nspname <> 'pg_catalog'::name) OR (co.collname <> 'default'::name)))))
  WHERE ((a.attnum > 0) AND (NOT a.attisdropped) AND (c.relkind = 'c'::"char") AND (pg_has_role(c.relowner, 'USAGE'::text) OR has_type_privilege(c.reltype, 'USAGE'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("column_privileges"), T(r#" SELECT (u_grantor.rolname)::information_schema.sql_identifier AS grantor,
    (grantee.rolname)::information_schema.sql_identifier AS grantee,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (nc.nspname)::information_schema.sql_identifier AS table_schema,
    (x.relname)::information_schema.sql_identifier AS table_name,
    (x.attname)::information_schema.sql_identifier AS column_name,
    (x.prtype)::information_schema.character_data AS privilege_type,
    (
        CASE
            WHEN (pg_has_role(x.grantee, x.relowner, 'USAGE'::text) OR x.grantable) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_grantable
   FROM ( SELECT pr_c.grantor,
            pr_c.grantee,
            a.attname,
            pr_c.relname,
            pr_c.relnamespace,
            pr_c.prtype,
            pr_c.grantable,
            pr_c.relowner
           FROM ( SELECT pg_class.oid,
                    pg_class.relname,
                    pg_class.relnamespace,
                    pg_class.relowner,
                    (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).grantor AS grantor,
                    (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).grantee AS grantee,
                    (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).privilege_type AS privilege_type,
                    (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).is_grantable AS is_grantable
                   FROM pg_class
                  WHERE (pg_class.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"]))) pr_c(oid, relname, relnamespace, relowner, grantor, grantee, prtype, grantable),
            pg_attribute a
          WHERE ((a.attrelid = pr_c.oid) AND (a.attnum > 0) AND (NOT a.attisdropped))
        UNION
         SELECT pr_a.grantor,
            pr_a.grantee,
            pr_a.attname,
            c.relname,
            c.relnamespace,
            pr_a.prtype,
            pr_a.grantable,
            c.relowner
           FROM ( SELECT a.attrelid,
                    a.attname,
                    (aclexplode(COALESCE(a.attacl, acldefault('c'::"char", cc.relowner)))).grantor AS grantor,
                    (aclexplode(COALESCE(a.attacl, acldefault('c'::"char", cc.relowner)))).grantee AS grantee,
                    (aclexplode(COALESCE(a.attacl, acldefault('c'::"char", cc.relowner)))).privilege_type AS privilege_type,
                    (aclexplode(COALESCE(a.attacl, acldefault('c'::"char", cc.relowner)))).is_grantable AS is_grantable
                   FROM (pg_attribute a
                     JOIN pg_class cc ON ((a.attrelid = cc.oid)))
                  WHERE ((a.attnum > 0) AND (NOT a.attisdropped))) pr_a(attrelid, attname, grantor, grantee, prtype, grantable),
            pg_class c
          WHERE ((pr_a.attrelid = c.oid) AND (c.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"])))) x,
    pg_namespace nc,
    pg_authid u_grantor,
    ( SELECT pg_authid.oid,
            pg_authid.rolname
           FROM pg_authid
        UNION ALL
         SELECT (0)::oid AS oid,
            'PUBLIC'::name) grantee(oid, rolname)
  WHERE ((x.relnamespace = nc.oid) AND (x.grantee = grantee.oid) AND (x.grantor = u_grantor.oid) AND (x.prtype = ANY (ARRAY['INSERT'::text, 'SELECT'::text, 'UPDATE'::text, 'REFERENCES'::text])) AND (pg_has_role(u_grantor.oid, 'USAGE'::text) OR pg_has_role(grantee.oid, 'USAGE'::text) OR (grantee.rolname = 'PUBLIC'::name)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("character_sets"), T(r#" SELECT (NULL::name)::information_schema.sql_identifier AS character_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS character_set_schema,
    (getdatabaseencoding())::information_schema.sql_identifier AS character_set_name,
    (
        CASE
            WHEN (getdatabaseencoding() = 'UTF8'::name) THEN 'UCS'::name
            ELSE getdatabaseencoding()
        END)::information_schema.sql_identifier AS character_repertoire,
    (getdatabaseencoding())::information_schema.sql_identifier AS form_of_use,
    (current_database())::information_schema.sql_identifier AS default_collate_catalog,
    (nc.nspname)::information_schema.sql_identifier AS default_collate_schema,
    (c.collname)::information_schema.sql_identifier AS default_collate_name
   FROM (pg_database d
     LEFT JOIN (pg_collation c
     JOIN pg_namespace nc ON ((c.collnamespace = nc.oid))) ON (((d.datcollate = c.collcollate) AND (d.datctype = c.collctype))))
  WHERE (d.datname = current_database())
  ORDER BY (char_length((c.collname)::text)) DESC, c.collname
 LIMIT 1;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("check_constraint_routine_usage"), T(r#" SELECT DISTINCT (current_database())::information_schema.sql_identifier AS constraint_catalog,
    (nc.nspname)::information_schema.sql_identifier AS constraint_schema,
    (c.conname)::information_schema.sql_identifier AS constraint_name,
    (current_database())::information_schema.sql_identifier AS specific_catalog,
    (np.nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS specific_name
   FROM pg_namespace nc,
    pg_constraint c,
    pg_depend d,
    pg_proc p,
    pg_namespace np
  WHERE ((nc.oid = c.connamespace) AND (c.contype = 'c'::"char") AND (c.oid = d.objid) AND (d.classid = ('pg_constraint'::regclass)::oid) AND (d.refobjid = p.oid) AND (d.refclassid = ('pg_proc'::regclass)::oid) AND (p.pronamespace = np.oid) AND pg_has_role(p.proowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("check_constraints"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS constraint_catalog,
    (rs.nspname)::information_schema.sql_identifier AS constraint_schema,
    (con.conname)::information_schema.sql_identifier AS constraint_name,
    (SUBSTRING(pg_get_constraintdef(con.oid) FROM 7))::information_schema.character_data AS check_clause
   FROM (((pg_constraint con
     LEFT JOIN pg_namespace rs ON ((rs.oid = con.connamespace)))
     LEFT JOIN pg_class c ON ((c.oid = con.conrelid)))
     LEFT JOIN pg_type t ON ((t.oid = con.contypid)))
  WHERE (pg_has_role(COALESCE(c.relowner, t.typowner), 'USAGE'::text) AND (con.contype = 'c'::"char"))
UNION
 SELECT (current_database())::information_schema.sql_identifier AS constraint_catalog,
    (n.nspname)::information_schema.sql_identifier AS constraint_schema,
    (((((((n.oid)::text || '_'::text) || (r.oid)::text) || '_'::text) || (a.attnum)::text) || '_not_null'::text))::information_schema.sql_identifier AS constraint_name,
    (((a.attname)::text || ' IS NOT NULL'::text))::information_schema.character_data AS check_clause
   FROM pg_namespace n,
    pg_class r,
    pg_attribute a
  WHERE ((n.oid = r.relnamespace) AND (r.oid = a.attrelid) AND (a.attnum > 0) AND (NOT a.attisdropped) AND a.attnotnull AND (r.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])) AND pg_has_role(r.relowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("column_udt_usage"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS udt_catalog,
    (COALESCE(nbt.nspname, nt.nspname))::information_schema.sql_identifier AS udt_schema,
    (COALESCE(bt.typname, t.typname))::information_schema.sql_identifier AS udt_name,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (nc.nspname)::information_schema.sql_identifier AS table_schema,
    (c.relname)::information_schema.sql_identifier AS table_name,
    (a.attname)::information_schema.sql_identifier AS column_name
   FROM pg_attribute a,
    pg_class c,
    pg_namespace nc,
    ((pg_type t
     JOIN pg_namespace nt ON ((t.typnamespace = nt.oid)))
     LEFT JOIN (pg_type bt
     JOIN pg_namespace nbt ON ((bt.typnamespace = nbt.oid))) ON (((t.typtype = 'd'::"char") AND (t.typbasetype = bt.oid))))
  WHERE ((a.attrelid = c.oid) AND (a.atttypid = t.oid) AND (nc.oid = c.relnamespace) AND (a.attnum > 0) AND (NOT a.attisdropped) AND (c.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"])) AND pg_has_role(COALESCE(bt.typowner, t.typowner), 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("collations"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS collation_catalog,
    (nc.nspname)::information_schema.sql_identifier AS collation_schema,
    (c.collname)::information_schema.sql_identifier AS collation_name,
    ('NO PAD'::character varying)::information_schema.character_data AS pad_attribute
   FROM pg_collation c,
    pg_namespace nc
  WHERE ((c.collnamespace = nc.oid) AND (c.collencoding = ANY (ARRAY['-1'::integer, ( SELECT pg_database.encoding
           FROM pg_database
          WHERE (pg_database.datname = current_database()))])));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("collation_character_set_applicability"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS collation_catalog,
    (nc.nspname)::information_schema.sql_identifier AS collation_schema,
    (c.collname)::information_schema.sql_identifier AS collation_name,
    (NULL::name)::information_schema.sql_identifier AS character_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS character_set_schema,
    (getdatabaseencoding())::information_schema.sql_identifier AS character_set_name
   FROM pg_collation c,
    pg_namespace nc
  WHERE ((c.collnamespace = nc.oid) AND (c.collencoding = ANY (ARRAY['-1'::integer, ( SELECT pg_database.encoding
           FROM pg_database
          WHERE (pg_database.datname = current_database()))])));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("key_column_usage"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS constraint_catalog,
    (ss.nc_nspname)::information_schema.sql_identifier AS constraint_schema,
    (ss.conname)::information_schema.sql_identifier AS constraint_name,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (ss.nr_nspname)::information_schema.sql_identifier AS table_schema,
    (ss.relname)::information_schema.sql_identifier AS table_name,
    (a.attname)::information_schema.sql_identifier AS column_name,
    ((ss.x).n)::information_schema.cardinal_number AS ordinal_position,
    (
        CASE
            WHEN (ss.contype = 'f'::"char") THEN information_schema._pg_index_position(ss.conindid, ss.confkey[(ss.x).n])
            ELSE NULL::integer
        END)::information_schema.cardinal_number AS position_in_unique_constraint
   FROM pg_attribute a,
    ( SELECT r.oid AS roid,
            r.relname,
            r.relowner,
            nc.nspname AS nc_nspname,
            nr.nspname AS nr_nspname,
            c.oid AS coid,
            c.conname,
            c.contype,
            c.conindid,
            c.confkey,
            c.confrelid,
            information_schema._pg_expandarray(c.conkey) AS x
           FROM pg_namespace nr,
            pg_class r,
            pg_namespace nc,
            pg_constraint c
          WHERE ((nr.oid = r.relnamespace) AND (r.oid = c.conrelid) AND (nc.oid = c.connamespace) AND (c.contype = ANY (ARRAY['p'::"char", 'u'::"char", 'f'::"char"])) AND (r.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])) AND (NOT pg_is_other_temp_schema(nr.oid)))) ss
  WHERE ((ss.roid = a.attrelid) AND (a.attnum = (ss.x).x) AND (NOT a.attisdropped) AND (pg_has_role(ss.relowner, 'USAGE'::text) OR has_column_privilege(ss.roid, a.attnum, 'SELECT, INSERT, UPDATE, REFERENCES'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("column_column_usage"), T(r#" SELECT DISTINCT (current_database())::information_schema.sql_identifier AS table_catalog,
    (n.nspname)::information_schema.sql_identifier AS table_schema,
    (c.relname)::information_schema.sql_identifier AS table_name,
    (ac.attname)::information_schema.sql_identifier AS column_name,
    (ad.attname)::information_schema.sql_identifier AS dependent_column
   FROM pg_namespace n,
    pg_class c,
    pg_depend d,
    pg_attribute ac,
    pg_attribute ad,
    pg_attrdef atd
  WHERE ((n.oid = c.relnamespace) AND (c.oid = ac.attrelid) AND (c.oid = ad.attrelid) AND (ac.attnum <> ad.attnum) AND (ad.attrelid = atd.adrelid) AND (ad.attnum = atd.adnum) AND (d.classid = ('pg_attrdef'::regclass)::oid) AND (d.refclassid = ('pg_class'::regclass)::oid) AND (d.objid = atd.oid) AND (d.refobjid = ac.attrelid) AND (d.refobjsubid = ac.attnum) AND (ad.attgenerated <> ''::"char") AND pg_has_role(c.relowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("columns"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS table_catalog,
    (nc.nspname)::information_schema.sql_identifier AS table_schema,
    (c.relname)::information_schema.sql_identifier AS table_name,
    (a.attname)::information_schema.sql_identifier AS column_name,
    (a.attnum)::information_schema.cardinal_number AS ordinal_position,
    (
        CASE
            WHEN (a.attgenerated = ''::"char") THEN pg_get_expr(ad.adbin, ad.adrelid)
            ELSE NULL::text
        END)::information_schema.character_data AS column_default,
    (
        CASE
            WHEN (a.attnotnull OR ((t.typtype = 'd'::"char") AND t.typnotnull)) THEN 'NO'::text
            ELSE 'YES'::text
        END)::information_schema.yes_or_no AS is_nullable,
    (
        CASE
            WHEN (t.typtype = 'd'::"char") THEN
            CASE
                WHEN ((bt.typelem <> (0)::oid) AND (bt.typlen = '-1'::integer)) THEN 'ARRAY'::text
                WHEN (nbt.nspname = 'pg_catalog'::name) THEN format_type(t.typbasetype, NULL::integer)
                ELSE 'USER-DEFINED'::text
            END
            ELSE
            CASE
                WHEN ((t.typelem <> (0)::oid) AND (t.typlen = '-1'::integer)) THEN 'ARRAY'::text
                WHEN (nt.nspname = 'pg_catalog'::name) THEN format_type(a.atttypid, NULL::integer)
                ELSE 'USER-DEFINED'::text
            END
        END)::information_schema.character_data AS data_type,
    (information_schema._pg_char_max_length(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS character_maximum_length,
    (information_schema._pg_char_octet_length(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS character_octet_length,
    (information_schema._pg_numeric_precision(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS numeric_precision,
    (information_schema._pg_numeric_precision_radix(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS numeric_precision_radix,
    (information_schema._pg_numeric_scale(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS numeric_scale,
    (information_schema._pg_datetime_precision(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.cardinal_number AS datetime_precision,
    (information_schema._pg_interval_type(information_schema._pg_truetypid(a.*, t.*), information_schema._pg_truetypmod(a.*, t.*)))::information_schema.character_data AS interval_type,
    (NULL::integer)::information_schema.cardinal_number AS interval_precision,
    (NULL::name)::information_schema.sql_identifier AS character_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS character_set_schema,
    (NULL::name)::information_schema.sql_identifier AS character_set_name,
    (
        CASE
            WHEN (nco.nspname IS NOT NULL) THEN current_database()
            ELSE NULL::name
        END)::information_schema.sql_identifier AS collation_catalog,
    (nco.nspname)::information_schema.sql_identifier AS collation_schema,
    (co.collname)::information_schema.sql_identifier AS collation_name,
    (
        CASE
            WHEN (t.typtype = 'd'::"char") THEN current_database()
            ELSE NULL::name
        END)::information_schema.sql_identifier AS domain_catalog,
    (
        CASE
            WHEN (t.typtype = 'd'::"char") THEN nt.nspname
            ELSE NULL::name
        END)::information_schema.sql_identifier AS domain_schema,
    (
        CASE
            WHEN (t.typtype = 'd'::"char") THEN t.typname
            ELSE NULL::name
        END)::information_schema.sql_identifier AS domain_name,
    (current_database())::information_schema.sql_identifier AS udt_catalog,
    (COALESCE(nbt.nspname, nt.nspname))::information_schema.sql_identifier AS udt_schema,
    (COALESCE(bt.typname, t.typname))::information_schema.sql_identifier AS udt_name,
    (NULL::name)::information_schema.sql_identifier AS scope_catalog,
    (NULL::name)::information_schema.sql_identifier AS scope_schema,
    (NULL::name)::information_schema.sql_identifier AS scope_name,
    (NULL::integer)::information_schema.cardinal_number AS maximum_cardinality,
    (a.attnum)::information_schema.sql_identifier AS dtd_identifier,
    ('NO'::character varying)::information_schema.yes_or_no AS is_self_referencing,
    (
        CASE
            WHEN (a.attidentity = ANY (ARRAY['a'::"char", 'd'::"char"])) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_identity,
    (
        CASE a.attidentity
            WHEN 'a'::"char" THEN 'ALWAYS'::text
            WHEN 'd'::"char" THEN 'BY DEFAULT'::text
            ELSE NULL::text
        END)::information_schema.character_data AS identity_generation,
    (seq.seqstart)::information_schema.character_data AS identity_start,
    (seq.seqincrement)::information_schema.character_data AS identity_increment,
    (seq.seqmax)::information_schema.character_data AS identity_maximum,
    (seq.seqmin)::information_schema.character_data AS identity_minimum,
    (
        CASE
            WHEN seq.seqcycle THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS identity_cycle,
    (
        CASE
            WHEN (a.attgenerated <> ''::"char") THEN 'ALWAYS'::text
            ELSE 'NEVER'::text
        END)::information_schema.character_data AS is_generated,
    (
        CASE
            WHEN (a.attgenerated <> ''::"char") THEN pg_get_expr(ad.adbin, ad.adrelid)
            ELSE NULL::text
        END)::information_schema.character_data AS generation_expression,
    (
        CASE
            WHEN ((c.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])) OR ((c.relkind = ANY (ARRAY['v'::"char", 'f'::"char"])) AND pg_column_is_updatable((c.oid)::regclass, a.attnum, false))) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_updatable
   FROM ((((((pg_attribute a
     LEFT JOIN pg_attrdef ad ON (((a.attrelid = ad.adrelid) AND (a.attnum = ad.adnum))))
     JOIN (pg_class c
     JOIN pg_namespace nc ON ((c.relnamespace = nc.oid))) ON ((a.attrelid = c.oid)))
     JOIN (pg_type t
     JOIN pg_namespace nt ON ((t.typnamespace = nt.oid))) ON ((a.atttypid = t.oid)))
     LEFT JOIN (pg_type bt
     JOIN pg_namespace nbt ON ((bt.typnamespace = nbt.oid))) ON (((t.typtype = 'd'::"char") AND (t.typbasetype = bt.oid))))
     LEFT JOIN (pg_collation co
     JOIN pg_namespace nco ON ((co.collnamespace = nco.oid))) ON (((a.attcollation = co.oid) AND ((nco.nspname <> 'pg_catalog'::name) OR (co.collname <> 'default'::name)))))
     LEFT JOIN (pg_depend dep
     JOIN pg_sequence seq ON (((dep.classid = ('pg_class'::regclass)::oid) AND (dep.objid = seq.seqrelid) AND (dep.deptype = 'i'::"char")))) ON (((dep.refclassid = ('pg_class'::regclass)::oid) AND (dep.refobjid = c.oid) AND (dep.refobjsubid = a.attnum))))
  WHERE ((NOT pg_is_other_temp_schema(nc.oid)) AND (a.attnum > 0) AND (NOT a.attisdropped) AND (c.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"])) AND (pg_has_role(c.relowner, 'USAGE'::text) OR has_column_privilege(c.oid, a.attnum, 'SELECT, INSERT, UPDATE, REFERENCES'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("domain_udt_usage"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS udt_catalog,
    (nbt.nspname)::information_schema.sql_identifier AS udt_schema,
    (bt.typname)::information_schema.sql_identifier AS udt_name,
    (current_database())::information_schema.sql_identifier AS domain_catalog,
    (nt.nspname)::information_schema.sql_identifier AS domain_schema,
    (t.typname)::information_schema.sql_identifier AS domain_name
   FROM pg_type t,
    pg_namespace nt,
    pg_type bt,
    pg_namespace nbt
  WHERE ((t.typnamespace = nt.oid) AND (t.typbasetype = bt.oid) AND (bt.typnamespace = nbt.oid) AND (t.typtype = 'd'::"char") AND pg_has_role(bt.typowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("constraint_column_usage"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS table_catalog,
    (x.tblschema)::information_schema.sql_identifier AS table_schema,
    (x.tblname)::information_schema.sql_identifier AS table_name,
    (x.colname)::information_schema.sql_identifier AS column_name,
    (current_database())::information_schema.sql_identifier AS constraint_catalog,
    (x.cstrschema)::information_schema.sql_identifier AS constraint_schema,
    (x.cstrname)::information_schema.sql_identifier AS constraint_name
   FROM ( SELECT DISTINCT nr.nspname,
            r.relname,
            r.relowner,
            a.attname,
            nc.nspname,
            c.conname
           FROM pg_namespace nr,
            pg_class r,
            pg_attribute a,
            pg_depend d,
            pg_namespace nc,
            pg_constraint c
          WHERE ((nr.oid = r.relnamespace) AND (r.oid = a.attrelid) AND (d.refclassid = ('pg_class'::regclass)::oid) AND (d.refobjid = r.oid) AND (d.refobjsubid = a.attnum) AND (d.classid = ('pg_constraint'::regclass)::oid) AND (d.objid = c.oid) AND (c.connamespace = nc.oid) AND (c.contype = 'c'::"char") AND (r.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])) AND (NOT a.attisdropped))
        UNION ALL
         SELECT nr.nspname,
            r.relname,
            r.relowner,
            a.attname,
            nc.nspname,
            c.conname
           FROM pg_namespace nr,
            pg_class r,
            pg_attribute a,
            pg_namespace nc,
            pg_constraint c
          WHERE ((nr.oid = r.relnamespace) AND (r.oid = a.attrelid) AND (nc.oid = c.connamespace) AND (r.oid =
                CASE c.contype
                    WHEN 'f'::"char" THEN c.confrelid
                    ELSE c.conrelid
                END) AND (a.attnum = ANY (
                CASE c.contype
                    WHEN 'f'::"char" THEN c.confkey
                    ELSE c.conkey
                END)) AND (NOT a.attisdropped) AND (c.contype = ANY (ARRAY['p'::"char", 'u'::"char", 'f'::"char"])) AND (r.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])))) x(tblschema, tblname, tblowner, colname, cstrschema, cstrname)
  WHERE pg_has_role(x.tblowner, 'USAGE'::text);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("constraint_table_usage"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS table_catalog,
    (nr.nspname)::information_schema.sql_identifier AS table_schema,
    (r.relname)::information_schema.sql_identifier AS table_name,
    (current_database())::information_schema.sql_identifier AS constraint_catalog,
    (nc.nspname)::information_schema.sql_identifier AS constraint_schema,
    (c.conname)::information_schema.sql_identifier AS constraint_name
   FROM pg_constraint c,
    pg_namespace nc,
    pg_class r,
    pg_namespace nr
  WHERE ((c.connamespace = nc.oid) AND (r.relnamespace = nr.oid) AND (((c.contype = 'f'::"char") AND (c.confrelid = r.oid)) OR ((c.contype = ANY (ARRAY['p'::"char", 'u'::"char"])) AND (c.conrelid = r.oid))) AND (r.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])) AND pg_has_role(r.relowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("domains"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS domain_catalog,
    (nt.nspname)::information_schema.sql_identifier AS domain_schema,
    (t.typname)::information_schema.sql_identifier AS domain_name,
    (
        CASE
            WHEN ((t.typelem <> (0)::oid) AND (t.typlen = '-1'::integer)) THEN 'ARRAY'::text
            WHEN (nbt.nspname = 'pg_catalog'::name) THEN format_type(t.typbasetype, NULL::integer)
            ELSE 'USER-DEFINED'::text
        END)::information_schema.character_data AS data_type,
    (information_schema._pg_char_max_length(t.typbasetype, t.typtypmod))::information_schema.cardinal_number AS character_maximum_length,
    (information_schema._pg_char_octet_length(t.typbasetype, t.typtypmod))::information_schema.cardinal_number AS character_octet_length,
    (NULL::name)::information_schema.sql_identifier AS character_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS character_set_schema,
    (NULL::name)::information_schema.sql_identifier AS character_set_name,
    (
        CASE
            WHEN (nco.nspname IS NOT NULL) THEN current_database()
            ELSE NULL::name
        END)::information_schema.sql_identifier AS collation_catalog,
    (nco.nspname)::information_schema.sql_identifier AS collation_schema,
    (co.collname)::information_schema.sql_identifier AS collation_name,
    (information_schema._pg_numeric_precision(t.typbasetype, t.typtypmod))::information_schema.cardinal_number AS numeric_precision,
    (information_schema._pg_numeric_precision_radix(t.typbasetype, t.typtypmod))::information_schema.cardinal_number AS numeric_precision_radix,
    (information_schema._pg_numeric_scale(t.typbasetype, t.typtypmod))::information_schema.cardinal_number AS numeric_scale,
    (information_schema._pg_datetime_precision(t.typbasetype, t.typtypmod))::information_schema.cardinal_number AS datetime_precision,
    (information_schema._pg_interval_type(t.typbasetype, t.typtypmod))::information_schema.character_data AS interval_type,
    (NULL::integer)::information_schema.cardinal_number AS interval_precision,
    (t.typdefault)::information_schema.character_data AS domain_default,
    (current_database())::information_schema.sql_identifier AS udt_catalog,
    (nbt.nspname)::information_schema.sql_identifier AS udt_schema,
    (bt.typname)::information_schema.sql_identifier AS udt_name,
    (NULL::name)::information_schema.sql_identifier AS scope_catalog,
    (NULL::name)::information_schema.sql_identifier AS scope_schema,
    (NULL::name)::information_schema.sql_identifier AS scope_name,
    (NULL::integer)::information_schema.cardinal_number AS maximum_cardinality,
    (1)::information_schema.sql_identifier AS dtd_identifier
   FROM (((pg_type t
     JOIN pg_namespace nt ON ((t.typnamespace = nt.oid)))
     JOIN (pg_type bt
     JOIN pg_namespace nbt ON ((bt.typnamespace = nbt.oid))) ON (((t.typbasetype = bt.oid) AND (t.typtype = 'd'::"char"))))
     LEFT JOIN (pg_collation co
     JOIN pg_namespace nco ON ((co.collnamespace = nco.oid))) ON (((t.typcollation = co.oid) AND ((nco.nspname <> 'pg_catalog'::name) OR (co.collname <> 'default'::name)))))
  WHERE (pg_has_role(t.typowner, 'USAGE'::text) OR has_type_privilege(t.oid, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("referential_constraints"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS constraint_catalog,
    (ncon.nspname)::information_schema.sql_identifier AS constraint_schema,
    (con.conname)::information_schema.sql_identifier AS constraint_name,
    (
        CASE
            WHEN (npkc.nspname IS NULL) THEN NULL::name
            ELSE current_database()
        END)::information_schema.sql_identifier AS unique_constraint_catalog,
    (npkc.nspname)::information_schema.sql_identifier AS unique_constraint_schema,
    (pkc.conname)::information_schema.sql_identifier AS unique_constraint_name,
    (
        CASE con.confmatchtype
            WHEN 'f'::"char" THEN 'FULL'::text
            WHEN 'p'::"char" THEN 'PARTIAL'::text
            WHEN 's'::"char" THEN 'NONE'::text
            ELSE NULL::text
        END)::information_schema.character_data AS match_option,
    (
        CASE con.confupdtype
            WHEN 'c'::"char" THEN 'CASCADE'::text
            WHEN 'n'::"char" THEN 'SET NULL'::text
            WHEN 'd'::"char" THEN 'SET DEFAULT'::text
            WHEN 'r'::"char" THEN 'RESTRICT'::text
            WHEN 'a'::"char" THEN 'NO ACTION'::text
            ELSE NULL::text
        END)::information_schema.character_data AS update_rule,
    (
        CASE con.confdeltype
            WHEN 'c'::"char" THEN 'CASCADE'::text
            WHEN 'n'::"char" THEN 'SET NULL'::text
            WHEN 'd'::"char" THEN 'SET DEFAULT'::text
            WHEN 'r'::"char" THEN 'RESTRICT'::text
            WHEN 'a'::"char" THEN 'NO ACTION'::text
            ELSE NULL::text
        END)::information_schema.character_data AS delete_rule
   FROM ((((((pg_namespace ncon
     JOIN pg_constraint con ON ((ncon.oid = con.connamespace)))
     JOIN pg_class c ON (((con.conrelid = c.oid) AND (con.contype = 'f'::"char"))))
     LEFT JOIN pg_depend d1 ON (((d1.objid = con.oid) AND (d1.classid = ('pg_constraint'::regclass)::oid) AND (d1.refclassid = ('pg_class'::regclass)::oid) AND (d1.refobjsubid = 0))))
     LEFT JOIN pg_depend d2 ON (((d2.refclassid = ('pg_constraint'::regclass)::oid) AND (d2.classid = ('pg_class'::regclass)::oid) AND (d2.objid = d1.refobjid) AND (d2.objsubid = 0) AND (d2.deptype = 'i'::"char"))))
     LEFT JOIN pg_constraint pkc ON (((pkc.oid = d2.refobjid) AND (pkc.contype = ANY (ARRAY['p'::"char", 'u'::"char"])) AND (pkc.conrelid = con.confrelid))))
     LEFT JOIN pg_namespace npkc ON ((pkc.connamespace = npkc.oid)))
  WHERE (pg_has_role(c.relowner, 'USAGE'::text) OR has_table_privilege(c.oid, 'INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER'::text) OR has_any_column_privilege(c.oid, 'INSERT, UPDATE, REFERENCES'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("enabled_roles"), T(r#" SELECT (a.rolname)::information_schema.sql_identifier AS role_name
   FROM pg_authid a
  WHERE pg_has_role(a.oid, 'USAGE'::text);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("parameters"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS specific_catalog,
    (ss.n_nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(ss.proname, ss.p_oid))::information_schema.sql_identifier AS specific_name,
    ((ss.x).n)::information_schema.cardinal_number AS ordinal_position,
    (
        CASE
            WHEN (ss.proargmodes IS NULL) THEN 'IN'::text
            WHEN (ss.proargmodes[(ss.x).n] = 'i'::"char") THEN 'IN'::text
            WHEN (ss.proargmodes[(ss.x).n] = 'o'::"char") THEN 'OUT'::text
            WHEN (ss.proargmodes[(ss.x).n] = 'b'::"char") THEN 'INOUT'::text
            WHEN (ss.proargmodes[(ss.x).n] = 'v'::"char") THEN 'IN'::text
            WHEN (ss.proargmodes[(ss.x).n] = 't'::"char") THEN 'OUT'::text
            ELSE NULL::text
        END)::information_schema.character_data AS parameter_mode,
    ('NO'::character varying)::information_schema.yes_or_no AS is_result,
    ('NO'::character varying)::information_schema.yes_or_no AS as_locator,
    (NULLIF(ss.proargnames[(ss.x).n], ''::text))::information_schema.sql_identifier AS parameter_name,
    (
        CASE
            WHEN ((t.typelem <> (0)::oid) AND (t.typlen = '-1'::integer)) THEN 'ARRAY'::text
            WHEN (nt.nspname = 'pg_catalog'::name) THEN format_type(t.oid, NULL::integer)
            ELSE 'USER-DEFINED'::text
        END)::information_schema.character_data AS data_type,
    (NULL::integer)::information_schema.cardinal_number AS character_maximum_length,
    (NULL::integer)::information_schema.cardinal_number AS character_octet_length,
    (NULL::name)::information_schema.sql_identifier AS character_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS character_set_schema,
    (NULL::name)::information_schema.sql_identifier AS character_set_name,
    (NULL::name)::information_schema.sql_identifier AS collation_catalog,
    (NULL::name)::information_schema.sql_identifier AS collation_schema,
    (NULL::name)::information_schema.sql_identifier AS collation_name,
    (NULL::integer)::information_schema.cardinal_number AS numeric_precision,
    (NULL::integer)::information_schema.cardinal_number AS numeric_precision_radix,
    (NULL::integer)::information_schema.cardinal_number AS numeric_scale,
    (NULL::integer)::information_schema.cardinal_number AS datetime_precision,
    (NULL::character varying)::information_schema.character_data AS interval_type,
    (NULL::integer)::information_schema.cardinal_number AS interval_precision,
    (current_database())::information_schema.sql_identifier AS udt_catalog,
    (nt.nspname)::information_schema.sql_identifier AS udt_schema,
    (t.typname)::information_schema.sql_identifier AS udt_name,
    (NULL::name)::information_schema.sql_identifier AS scope_catalog,
    (NULL::name)::information_schema.sql_identifier AS scope_schema,
    (NULL::name)::information_schema.sql_identifier AS scope_name,
    (NULL::integer)::information_schema.cardinal_number AS maximum_cardinality,
    ((ss.x).n)::information_schema.sql_identifier AS dtd_identifier,
    (
        CASE
            WHEN pg_has_role(ss.proowner, 'USAGE'::text) THEN pg_get_function_arg_default(ss.p_oid, (ss.x).n)
            ELSE NULL::text
        END)::information_schema.character_data AS parameter_default
   FROM pg_type t,
    pg_namespace nt,
    ( SELECT n.nspname AS n_nspname,
            p.proname,
            p.oid AS p_oid,
            p.proowner,
            p.proargnames,
            p.proargmodes,
            information_schema._pg_expandarray(COALESCE(p.proallargtypes, (p.proargtypes)::oid[])) AS x
           FROM pg_namespace n,
            pg_proc p
          WHERE ((n.oid = p.pronamespace) AND (pg_has_role(p.proowner, 'USAGE'::text) OR has_function_privilege(p.oid, 'EXECUTE'::text)))) ss
  WHERE ((t.oid = (ss.x).x) AND (t.typnamespace = nt.oid));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("routine_column_usage"), T(r#" SELECT DISTINCT (current_database())::information_schema.sql_identifier AS specific_catalog,
    (np.nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS specific_name,
    (current_database())::information_schema.sql_identifier AS routine_catalog,
    (np.nspname)::information_schema.sql_identifier AS routine_schema,
    (p.proname)::information_schema.sql_identifier AS routine_name,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (nt.nspname)::information_schema.sql_identifier AS table_schema,
    (t.relname)::information_schema.sql_identifier AS table_name,
    (a.attname)::information_schema.sql_identifier AS column_name
   FROM pg_namespace np,
    pg_proc p,
    pg_depend d,
    pg_class t,
    pg_namespace nt,
    pg_attribute a
  WHERE ((np.oid = p.pronamespace) AND (p.oid = d.objid) AND (d.classid = ('pg_proc'::regclass)::oid) AND (d.refobjid = t.oid) AND (d.refclassid = ('pg_class'::regclass)::oid) AND (t.relnamespace = nt.oid) AND (t.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"])) AND (t.oid = a.attrelid) AND (d.refobjsubid = a.attnum) AND pg_has_role(t.relowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("routine_privileges"), T(r#" SELECT (u_grantor.rolname)::information_schema.sql_identifier AS grantor,
    (grantee.rolname)::information_schema.sql_identifier AS grantee,
    (current_database())::information_schema.sql_identifier AS specific_catalog,
    (n.nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS specific_name,
    (current_database())::information_schema.sql_identifier AS routine_catalog,
    (n.nspname)::information_schema.sql_identifier AS routine_schema,
    (p.proname)::information_schema.sql_identifier AS routine_name,
    ('EXECUTE'::character varying)::information_schema.character_data AS privilege_type,
    (
        CASE
            WHEN (pg_has_role(grantee.oid, p.proowner, 'USAGE'::text) OR p.grantable) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_grantable
   FROM ( SELECT pg_proc.oid,
            pg_proc.proname,
            pg_proc.proowner,
            pg_proc.pronamespace,
            (aclexplode(COALESCE(pg_proc.proacl, acldefault('f'::"char", pg_proc.proowner)))).grantor AS grantor,
            (aclexplode(COALESCE(pg_proc.proacl, acldefault('f'::"char", pg_proc.proowner)))).grantee AS grantee,
            (aclexplode(COALESCE(pg_proc.proacl, acldefault('f'::"char", pg_proc.proowner)))).privilege_type AS privilege_type,
            (aclexplode(COALESCE(pg_proc.proacl, acldefault('f'::"char", pg_proc.proowner)))).is_grantable AS is_grantable
           FROM pg_proc) p(oid, proname, proowner, pronamespace, grantor, grantee, prtype, grantable),
    pg_namespace n,
    pg_authid u_grantor,
    ( SELECT pg_authid.oid,
            pg_authid.rolname
           FROM pg_authid
        UNION ALL
         SELECT (0)::oid AS oid,
            'PUBLIC'::name) grantee(oid, rolname)
  WHERE ((p.pronamespace = n.oid) AND (grantee.oid = p.grantee) AND (u_grantor.oid = p.grantor) AND (p.prtype = 'EXECUTE'::text) AND (pg_has_role(u_grantor.oid, 'USAGE'::text) OR pg_has_role(grantee.oid, 'USAGE'::text) OR (grantee.rolname = 'PUBLIC'::name)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("role_routine_grants"), T(r#" SELECT routine_privileges.grantor,
    routine_privileges.grantee,
    routine_privileges.specific_catalog,
    routine_privileges.specific_schema,
    routine_privileges.specific_name,
    routine_privileges.routine_catalog,
    routine_privileges.routine_schema,
    routine_privileges.routine_name,
    routine_privileges.privilege_type,
    routine_privileges.is_grantable
   FROM information_schema.routine_privileges
  WHERE (((routine_privileges.grantor)::name IN ( SELECT enabled_roles.role_name
           FROM information_schema.enabled_roles)) OR ((routine_privileges.grantee)::name IN ( SELECT enabled_roles.role_name
           FROM information_schema.enabled_roles)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("routine_routine_usage"), T(r#" SELECT DISTINCT (current_database())::information_schema.sql_identifier AS specific_catalog,
    (np.nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS specific_name,
    (current_database())::information_schema.sql_identifier AS routine_catalog,
    (np1.nspname)::information_schema.sql_identifier AS routine_schema,
    (nameconcatoid(p1.proname, p1.oid))::information_schema.sql_identifier AS routine_name
   FROM pg_namespace np,
    pg_proc p,
    pg_depend d,
    pg_proc p1,
    pg_namespace np1
  WHERE ((np.oid = p.pronamespace) AND (p.oid = d.objid) AND (d.classid = ('pg_proc'::regclass)::oid) AND (d.refobjid = p1.oid) AND (d.refclassid = ('pg_proc'::regclass)::oid) AND (p1.pronamespace = np1.oid) AND (p.prokind = ANY (ARRAY['f'::"char", 'p'::"char"])) AND (p1.prokind = ANY (ARRAY['f'::"char", 'p'::"char"])) AND pg_has_role(p1.proowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("table_privileges"), T(r#" SELECT (u_grantor.rolname)::information_schema.sql_identifier AS grantor,
    (grantee.rolname)::information_schema.sql_identifier AS grantee,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (nc.nspname)::information_schema.sql_identifier AS table_schema,
    (c.relname)::information_schema.sql_identifier AS table_name,
    (c.prtype)::information_schema.character_data AS privilege_type,
    (
        CASE
            WHEN (pg_has_role(grantee.oid, c.relowner, 'USAGE'::text) OR c.grantable) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_grantable,
    (
        CASE
            WHEN (c.prtype = 'SELECT'::text) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS with_hierarchy
   FROM ( SELECT pg_class.oid,
            pg_class.relname,
            pg_class.relnamespace,
            pg_class.relkind,
            pg_class.relowner,
            (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).grantor AS grantor,
            (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).grantee AS grantee,
            (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).privilege_type AS privilege_type,
            (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).is_grantable AS is_grantable
           FROM pg_class) c(oid, relname, relnamespace, relkind, relowner, grantor, grantee, prtype, grantable),
    pg_namespace nc,
    pg_authid u_grantor,
    ( SELECT pg_authid.oid,
            pg_authid.rolname
           FROM pg_authid
        UNION ALL
         SELECT (0)::oid AS oid,
            'PUBLIC'::name) grantee(oid, rolname)
  WHERE ((c.relnamespace = nc.oid) AND (c.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"])) AND (c.grantee = grantee.oid) AND (c.grantor = u_grantor.oid) AND (c.prtype = ANY (ARRAY['INSERT'::text, 'SELECT'::text, 'UPDATE'::text, 'DELETE'::text, 'TRUNCATE'::text, 'REFERENCES'::text, 'TRIGGER'::text])) AND (pg_has_role(u_grantor.oid, 'USAGE'::text) OR pg_has_role(grantee.oid, 'USAGE'::text) OR (grantee.rolname = 'PUBLIC'::name)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("routine_sequence_usage"), T(r#" SELECT DISTINCT (current_database())::information_schema.sql_identifier AS specific_catalog,
    (np.nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS specific_name,
    (current_database())::information_schema.sql_identifier AS routine_catalog,
    (np.nspname)::information_schema.sql_identifier AS routine_schema,
    (p.proname)::information_schema.sql_identifier AS routine_name,
    (current_database())::information_schema.sql_identifier AS sequence_catalog,
    (ns.nspname)::information_schema.sql_identifier AS sequence_schema,
    (s.relname)::information_schema.sql_identifier AS sequence_name
   FROM pg_namespace np,
    pg_proc p,
    pg_depend d,
    pg_class s,
    pg_namespace ns
  WHERE ((np.oid = p.pronamespace) AND (p.oid = d.objid) AND (d.classid = ('pg_proc'::regclass)::oid) AND (d.refobjid = s.oid) AND (d.refclassid = ('pg_class'::regclass)::oid) AND (s.relnamespace = ns.oid) AND (s.relkind = 'S'::"char") AND pg_has_role(s.relowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("routine_table_usage"), T(r#" SELECT DISTINCT (current_database())::information_schema.sql_identifier AS specific_catalog,
    (np.nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS specific_name,
    (current_database())::information_schema.sql_identifier AS routine_catalog,
    (np.nspname)::information_schema.sql_identifier AS routine_schema,
    (p.proname)::information_schema.sql_identifier AS routine_name,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (nt.nspname)::information_schema.sql_identifier AS table_schema,
    (t.relname)::information_schema.sql_identifier AS table_name
   FROM pg_namespace np,
    pg_proc p,
    pg_depend d,
    pg_class t,
    pg_namespace nt
  WHERE ((np.oid = p.pronamespace) AND (p.oid = d.objid) AND (d.classid = ('pg_proc'::regclass)::oid) AND (d.refobjid = t.oid) AND (d.refclassid = ('pg_class'::regclass)::oid) AND (t.relnamespace = nt.oid) AND (t.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"])) AND pg_has_role(t.relowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("udt_privileges"), T(r#" SELECT (u_grantor.rolname)::information_schema.sql_identifier AS grantor,
    (grantee.rolname)::information_schema.sql_identifier AS grantee,
    (current_database())::information_schema.sql_identifier AS udt_catalog,
    (n.nspname)::information_schema.sql_identifier AS udt_schema,
    (t.typname)::information_schema.sql_identifier AS udt_name,
    ('TYPE USAGE'::character varying)::information_schema.character_data AS privilege_type,
    (
        CASE
            WHEN (pg_has_role(grantee.oid, t.typowner, 'USAGE'::text) OR t.grantable) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_grantable
   FROM ( SELECT pg_type.oid,
            pg_type.typname,
            pg_type.typnamespace,
            pg_type.typtype,
            pg_type.typowner,
            (aclexplode(COALESCE(pg_type.typacl, acldefault('T'::"char", pg_type.typowner)))).grantor AS grantor,
            (aclexplode(COALESCE(pg_type.typacl, acldefault('T'::"char", pg_type.typowner)))).grantee AS grantee,
            (aclexplode(COALESCE(pg_type.typacl, acldefault('T'::"char", pg_type.typowner)))).privilege_type AS privilege_type,
            (aclexplode(COALESCE(pg_type.typacl, acldefault('T'::"char", pg_type.typowner)))).is_grantable AS is_grantable
           FROM pg_type) t(oid, typname, typnamespace, typtype, typowner, grantor, grantee, prtype, grantable),
    pg_namespace n,
    pg_authid u_grantor,
    ( SELECT pg_authid.oid,
            pg_authid.rolname
           FROM pg_authid
        UNION ALL
         SELECT (0)::oid AS oid,
            'PUBLIC'::name) grantee(oid, rolname)
  WHERE ((t.typnamespace = n.oid) AND (t.typtype = 'c'::"char") AND (t.grantee = grantee.oid) AND (t.grantor = u_grantor.oid) AND (t.prtype = 'USAGE'::text) AND (pg_has_role(u_grantor.oid, 'USAGE'::text) OR pg_has_role(grantee.oid, 'USAGE'::text) OR (grantee.rolname = 'PUBLIC'::name)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("routines"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS specific_catalog,
    (n.nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS specific_name,
    (current_database())::information_schema.sql_identifier AS routine_catalog,
    (n.nspname)::information_schema.sql_identifier AS routine_schema,
    (p.proname)::information_schema.sql_identifier AS routine_name,
    (
        CASE p.prokind
            WHEN 'f'::"char" THEN 'FUNCTION'::text
            WHEN 'p'::"char" THEN 'PROCEDURE'::text
            ELSE NULL::text
        END)::information_schema.character_data AS routine_type,
    (NULL::name)::information_schema.sql_identifier AS module_catalog,
    (NULL::name)::information_schema.sql_identifier AS module_schema,
    (NULL::name)::information_schema.sql_identifier AS module_name,
    (NULL::name)::information_schema.sql_identifier AS udt_catalog,
    (NULL::name)::information_schema.sql_identifier AS udt_schema,
    (NULL::name)::information_schema.sql_identifier AS udt_name,
    (
        CASE
            WHEN (p.prokind = 'p'::"char") THEN NULL::text
            WHEN ((t.typelem <> (0)::oid) AND (t.typlen = '-1'::integer)) THEN 'ARRAY'::text
            WHEN (nt.nspname = 'pg_catalog'::name) THEN format_type(t.oid, NULL::integer)
            ELSE 'USER-DEFINED'::text
        END)::information_schema.character_data AS data_type,
    (NULL::integer)::information_schema.cardinal_number AS character_maximum_length,
    (NULL::integer)::information_schema.cardinal_number AS character_octet_length,
    (NULL::name)::information_schema.sql_identifier AS character_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS character_set_schema,
    (NULL::name)::information_schema.sql_identifier AS character_set_name,
    (NULL::name)::information_schema.sql_identifier AS collation_catalog,
    (NULL::name)::information_schema.sql_identifier AS collation_schema,
    (NULL::name)::information_schema.sql_identifier AS collation_name,
    (NULL::integer)::information_schema.cardinal_number AS numeric_precision,
    (NULL::integer)::information_schema.cardinal_number AS numeric_precision_radix,
    (NULL::integer)::information_schema.cardinal_number AS numeric_scale,
    (NULL::integer)::information_schema.cardinal_number AS datetime_precision,
    (NULL::character varying)::information_schema.character_data AS interval_type,
    (NULL::integer)::information_schema.cardinal_number AS interval_precision,
    (
        CASE
            WHEN (nt.nspname IS NOT NULL) THEN current_database()
            ELSE NULL::name
        END)::information_schema.sql_identifier AS type_udt_catalog,
    (nt.nspname)::information_schema.sql_identifier AS type_udt_schema,
    (t.typname)::information_schema.sql_identifier AS type_udt_name,
    (NULL::name)::information_schema.sql_identifier AS scope_catalog,
    (NULL::name)::information_schema.sql_identifier AS scope_schema,
    (NULL::name)::information_schema.sql_identifier AS scope_name,
    (NULL::integer)::information_schema.cardinal_number AS maximum_cardinality,
    (
        CASE
            WHEN (p.prokind <> 'p'::"char") THEN 0
            ELSE NULL::integer
        END)::information_schema.sql_identifier AS dtd_identifier,
    (
        CASE
            WHEN (l.lanname = 'sql'::name) THEN 'SQL'::text
            ELSE 'EXTERNAL'::text
        END)::information_schema.character_data AS routine_body,
    (
        CASE
            WHEN pg_has_role(p.proowner, 'USAGE'::text) THEN p.prosrc
            ELSE NULL::text
        END)::information_schema.character_data AS routine_definition,
    (
        CASE
            WHEN (l.lanname = 'c'::name) THEN p.prosrc
            ELSE NULL::text
        END)::information_schema.character_data AS external_name,
    (upper((l.lanname)::text))::information_schema.character_data AS external_language,
    ('GENERAL'::character varying)::information_schema.character_data AS parameter_style,
    (
        CASE
            WHEN (p.provolatile = 'i'::"char") THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_deterministic,
    ('MODIFIES'::character varying)::information_schema.character_data AS sql_data_access,
    (
        CASE
            WHEN (p.prokind <> 'p'::"char") THEN
            CASE
                WHEN p.proisstrict THEN 'YES'::text
                ELSE 'NO'::text
            END
            ELSE NULL::text
        END)::information_schema.yes_or_no AS is_null_call,
    (NULL::character varying)::information_schema.character_data AS sql_path,
    ('YES'::character varying)::information_schema.yes_or_no AS schema_level_routine,
    (0)::information_schema.cardinal_number AS max_dynamic_result_sets,
    (NULL::character varying)::information_schema.yes_or_no AS is_user_defined_cast,
    (NULL::character varying)::information_schema.yes_or_no AS is_implicitly_invocable,
    (
        CASE
            WHEN p.prosecdef THEN 'DEFINER'::text
            ELSE 'INVOKER'::text
        END)::information_schema.character_data AS security_type,
    (NULL::name)::information_schema.sql_identifier AS to_sql_specific_catalog,
    (NULL::name)::information_schema.sql_identifier AS to_sql_specific_schema,
    (NULL::name)::information_schema.sql_identifier AS to_sql_specific_name,
    ('NO'::character varying)::information_schema.yes_or_no AS as_locator,
    (NULL::timestamp with time zone)::information_schema.time_stamp AS created,
    (NULL::timestamp with time zone)::information_schema.time_stamp AS last_altered,
    (NULL::character varying)::information_schema.yes_or_no AS new_savepoint_level,
    ('NO'::character varying)::information_schema.yes_or_no AS is_udt_dependent,
    (NULL::character varying)::information_schema.character_data AS result_cast_from_data_type,
    (NULL::character varying)::information_schema.yes_or_no AS result_cast_as_locator,
    (NULL::integer)::information_schema.cardinal_number AS result_cast_char_max_length,
    (NULL::integer)::information_schema.cardinal_number AS result_cast_char_octet_length,
    (NULL::name)::information_schema.sql_identifier AS result_cast_char_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS result_cast_char_set_schema,
    (NULL::name)::information_schema.sql_identifier AS result_cast_char_set_name,
    (NULL::name)::information_schema.sql_identifier AS result_cast_collation_catalog,
    (NULL::name)::information_schema.sql_identifier AS result_cast_collation_schema,
    (NULL::name)::information_schema.sql_identifier AS result_cast_collation_name,
    (NULL::integer)::information_schema.cardinal_number AS result_cast_numeric_precision,
    (NULL::integer)::information_schema.cardinal_number AS result_cast_numeric_precision_radix,
    (NULL::integer)::information_schema.cardinal_number AS result_cast_numeric_scale,
    (NULL::integer)::information_schema.cardinal_number AS result_cast_datetime_precision,
    (NULL::character varying)::information_schema.character_data AS result_cast_interval_type,
    (NULL::integer)::information_schema.cardinal_number AS result_cast_interval_precision,
    (NULL::name)::information_schema.sql_identifier AS result_cast_type_udt_catalog,
    (NULL::name)::information_schema.sql_identifier AS result_cast_type_udt_schema,
    (NULL::name)::information_schema.sql_identifier AS result_cast_type_udt_name,
    (NULL::name)::information_schema.sql_identifier AS result_cast_scope_catalog,
    (NULL::name)::information_schema.sql_identifier AS result_cast_scope_schema,
    (NULL::name)::information_schema.sql_identifier AS result_cast_scope_name,
    (NULL::integer)::information_schema.cardinal_number AS result_cast_maximum_cardinality,
    (NULL::name)::information_schema.sql_identifier AS result_cast_dtd_identifier
   FROM (((pg_namespace n
     JOIN pg_proc p ON ((n.oid = p.pronamespace)))
     JOIN pg_language l ON ((p.prolang = l.oid)))
     LEFT JOIN (pg_type t
     JOIN pg_namespace nt ON ((t.typnamespace = nt.oid))) ON (((p.prorettype = t.oid) AND (p.prokind <> 'p'::"char"))))
  WHERE (pg_has_role(p.proowner, 'USAGE'::text) OR has_function_privilege(p.oid, 'EXECUTE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("role_table_grants"), T(r#" SELECT table_privileges.grantor,
    table_privileges.grantee,
    table_privileges.table_catalog,
    table_privileges.table_schema,
    table_privileges.table_name,
    table_privileges.privilege_type,
    table_privileges.is_grantable,
    table_privileges.with_hierarchy
   FROM information_schema.table_privileges
  WHERE (((table_privileges.grantor)::name IN ( SELECT enabled_roles.role_name
           FROM information_schema.enabled_roles)) OR ((table_privileges.grantee)::name IN ( SELECT enabled_roles.role_name
           FROM information_schema.enabled_roles)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("schemata"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS catalog_name,
    (n.nspname)::information_schema.sql_identifier AS schema_name,
    (u.rolname)::information_schema.sql_identifier AS schema_owner,
    (NULL::name)::information_schema.sql_identifier AS default_character_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS default_character_set_schema,
    (NULL::name)::information_schema.sql_identifier AS default_character_set_name,
    (NULL::character varying)::information_schema.character_data AS sql_path
   FROM pg_namespace n,
    pg_authid u
  WHERE ((n.nspowner = u.oid) AND (pg_has_role(n.nspowner, 'USAGE'::text) OR has_schema_privilege(n.oid, 'CREATE, USAGE'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("sequences"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS sequence_catalog,
    (nc.nspname)::information_schema.sql_identifier AS sequence_schema,
    (c.relname)::information_schema.sql_identifier AS sequence_name,
    (format_type(s.seqtypid, NULL::integer))::information_schema.character_data AS data_type,
    (information_schema._pg_numeric_precision(s.seqtypid, '-1'::integer))::information_schema.cardinal_number AS numeric_precision,
    (2)::information_schema.cardinal_number AS numeric_precision_radix,
    (0)::information_schema.cardinal_number AS numeric_scale,
    (s.seqstart)::information_schema.character_data AS start_value,
    (s.seqmin)::information_schema.character_data AS minimum_value,
    (s.seqmax)::information_schema.character_data AS maximum_value,
    (s.seqincrement)::information_schema.character_data AS increment,
    (
        CASE
            WHEN s.seqcycle THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS cycle_option
   FROM pg_namespace nc,
    pg_class c,
    pg_sequence s
  WHERE ((c.relnamespace = nc.oid) AND (c.relkind = 'S'::"char") AND (NOT (EXISTS ( SELECT 1
           FROM pg_depend
          WHERE ((pg_depend.classid = ('pg_class'::regclass)::oid) AND (pg_depend.objid = c.oid) AND (pg_depend.deptype = 'i'::"char"))))) AND (NOT pg_is_other_temp_schema(nc.oid)) AND (c.oid = s.seqrelid) AND (pg_has_role(c.relowner, 'USAGE'::text) OR has_sequence_privilege(c.oid, 'SELECT, UPDATE, USAGE'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("tables"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS table_catalog,
    (nc.nspname)::information_schema.sql_identifier AS table_schema,
    (c.relname)::information_schema.sql_identifier AS table_name,
    (
        CASE
            WHEN (nc.oid = pg_my_temp_schema()) THEN 'LOCAL TEMPORARY'::text
            WHEN (c.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])) THEN 'BASE TABLE'::text
            WHEN (c.relkind = 'v'::"char") THEN 'VIEW'::text
            WHEN (c.relkind = 'f'::"char") THEN 'FOREIGN'::text
            ELSE NULL::text
        END)::information_schema.character_data AS table_type,
    (NULL::name)::information_schema.sql_identifier AS self_referencing_column_name,
    (NULL::character varying)::information_schema.character_data AS reference_generation,
    (
        CASE
            WHEN (t.typname IS NOT NULL) THEN current_database()
            ELSE NULL::name
        END)::information_schema.sql_identifier AS user_defined_type_catalog,
    (nt.nspname)::information_schema.sql_identifier AS user_defined_type_schema,
    (t.typname)::information_schema.sql_identifier AS user_defined_type_name,
    (
        CASE
            WHEN ((c.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])) OR ((c.relkind = ANY (ARRAY['v'::"char", 'f'::"char"])) AND ((pg_relation_is_updatable((c.oid)::regclass, false) & 8) = 8))) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_insertable_into,
    (
        CASE
            WHEN (t.typname IS NOT NULL) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_typed,
    (NULL::character varying)::information_schema.character_data AS commit_action
   FROM ((pg_namespace nc
     JOIN pg_class c ON ((nc.oid = c.relnamespace)))
     LEFT JOIN (pg_type t
     JOIN pg_namespace nt ON ((t.typnamespace = nt.oid))) ON ((c.reloftype = t.oid)))
  WHERE ((c.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"])) AND (NOT pg_is_other_temp_schema(nc.oid)) AND (pg_has_role(c.relowner, 'USAGE'::text) OR has_table_privilege(c.oid, 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER'::text) OR has_any_column_privilege(c.oid, 'SELECT, INSERT, UPDATE, REFERENCES'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("transforms"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS udt_catalog,
    (nt.nspname)::information_schema.sql_identifier AS udt_schema,
    (t.typname)::information_schema.sql_identifier AS udt_name,
    (current_database())::information_schema.sql_identifier AS specific_catalog,
    (np.nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS specific_name,
    (l.lanname)::information_schema.sql_identifier AS group_name,
    ('FROM SQL'::character varying)::information_schema.character_data AS transform_type
   FROM (((((pg_type t
     JOIN pg_transform x ON ((t.oid = x.trftype)))
     JOIN pg_language l ON ((x.trflang = l.oid)))
     JOIN pg_proc p ON (((x.trffromsql)::oid = p.oid)))
     JOIN pg_namespace nt ON ((t.typnamespace = nt.oid)))
     JOIN pg_namespace np ON ((p.pronamespace = np.oid)))
UNION
 SELECT (current_database())::information_schema.sql_identifier AS udt_catalog,
    (nt.nspname)::information_schema.sql_identifier AS udt_schema,
    (t.typname)::information_schema.sql_identifier AS udt_name,
    (current_database())::information_schema.sql_identifier AS specific_catalog,
    (np.nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS specific_name,
    (l.lanname)::information_schema.sql_identifier AS group_name,
    ('TO SQL'::character varying)::information_schema.character_data AS transform_type
   FROM (((((pg_type t
     JOIN pg_transform x ON ((t.oid = x.trftype)))
     JOIN pg_language l ON ((x.trflang = l.oid)))
     JOIN pg_proc p ON (((x.trftosql)::oid = p.oid)))
     JOIN pg_namespace nt ON ((t.typnamespace = nt.oid)))
     JOIN pg_namespace np ON ((p.pronamespace = np.oid)))
  ORDER BY 1, 2, 3, 7, 8;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("table_constraints"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS constraint_catalog,
    (nc.nspname)::information_schema.sql_identifier AS constraint_schema,
    (c.conname)::information_schema.sql_identifier AS constraint_name,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (nr.nspname)::information_schema.sql_identifier AS table_schema,
    (r.relname)::information_schema.sql_identifier AS table_name,
    (
        CASE c.contype
            WHEN 'c'::"char" THEN 'CHECK'::text
            WHEN 'f'::"char" THEN 'FOREIGN KEY'::text
            WHEN 'p'::"char" THEN 'PRIMARY KEY'::text
            WHEN 'u'::"char" THEN 'UNIQUE'::text
            ELSE NULL::text
        END)::information_schema.character_data AS constraint_type,
    (
        CASE
            WHEN c.condeferrable THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_deferrable,
    (
        CASE
            WHEN c.condeferred THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS initially_deferred,
    ('YES'::character varying)::information_schema.yes_or_no AS enforced,
    (
        CASE
            WHEN (c.contype = 'u'::"char") THEN
            CASE
                WHEN ( SELECT (NOT pg_index.indnullsnotdistinct)
                   FROM pg_index
                  WHERE (pg_index.indexrelid = c.conindid)) THEN 'YES'::text
                ELSE 'NO'::text
            END
            ELSE NULL::text
        END)::information_schema.yes_or_no AS nulls_distinct
   FROM pg_namespace nc,
    pg_namespace nr,
    pg_constraint c,
    pg_class r
  WHERE ((nc.oid = c.connamespace) AND (nr.oid = r.relnamespace) AND (c.conrelid = r.oid) AND (c.contype <> ALL (ARRAY['t'::"char", 'x'::"char"])) AND (r.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])) AND (NOT pg_is_other_temp_schema(nr.oid)) AND (pg_has_role(r.relowner, 'USAGE'::text) OR has_table_privilege(r.oid, 'INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER'::text) OR has_any_column_privilege(r.oid, 'INSERT, UPDATE, REFERENCES'::text)))
UNION ALL
 SELECT (current_database())::information_schema.sql_identifier AS constraint_catalog,
    (nr.nspname)::information_schema.sql_identifier AS constraint_schema,
    (((((((nr.oid)::text || '_'::text) || (r.oid)::text) || '_'::text) || (a.attnum)::text) || '_not_null'::text))::information_schema.sql_identifier AS constraint_name,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (nr.nspname)::information_schema.sql_identifier AS table_schema,
    (r.relname)::information_schema.sql_identifier AS table_name,
    ('CHECK'::character varying)::information_schema.character_data AS constraint_type,
    ('NO'::character varying)::information_schema.yes_or_no AS is_deferrable,
    ('NO'::character varying)::information_schema.yes_or_no AS initially_deferred,
    ('YES'::character varying)::information_schema.yes_or_no AS enforced,
    (NULL::character varying)::information_schema.yes_or_no AS nulls_distinct
   FROM pg_namespace nr,
    pg_class r,
    pg_attribute a
  WHERE ((nr.oid = r.relnamespace) AND (r.oid = a.attrelid) AND a.attnotnull AND (a.attnum > 0) AND (NOT a.attisdropped) AND (r.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])) AND (NOT pg_is_other_temp_schema(nr.oid)) AND (pg_has_role(r.relowner, 'USAGE'::text) OR has_table_privilege(r.oid, 'INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER'::text) OR has_any_column_privilege(r.oid, 'INSERT, UPDATE, REFERENCES'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("view_routine_usage"), T(r#" SELECT DISTINCT (current_database())::information_schema.sql_identifier AS table_catalog,
    (nv.nspname)::information_schema.sql_identifier AS table_schema,
    (v.relname)::information_schema.sql_identifier AS table_name,
    (current_database())::information_schema.sql_identifier AS specific_catalog,
    (np.nspname)::information_schema.sql_identifier AS specific_schema,
    (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS specific_name
   FROM pg_namespace nv,
    pg_class v,
    pg_depend dv,
    pg_depend dp,
    pg_proc p,
    pg_namespace np
  WHERE ((nv.oid = v.relnamespace) AND (v.relkind = 'v'::"char") AND (v.oid = dv.refobjid) AND (dv.refclassid = ('pg_class'::regclass)::oid) AND (dv.classid = ('pg_rewrite'::regclass)::oid) AND (dv.deptype = 'i'::"char") AND (dv.objid = dp.objid) AND (dp.classid = ('pg_rewrite'::regclass)::oid) AND (dp.refclassid = ('pg_proc'::regclass)::oid) AND (dp.refobjid = p.oid) AND (p.pronamespace = np.oid) AND pg_has_role(p.proowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("role_udt_grants"), T(r#" SELECT udt_privileges.grantor,
    udt_privileges.grantee,
    udt_privileges.udt_catalog,
    udt_privileges.udt_schema,
    udt_privileges.udt_name,
    udt_privileges.privilege_type,
    udt_privileges.is_grantable
   FROM information_schema.udt_privileges
  WHERE (((udt_privileges.grantor)::name IN ( SELECT enabled_roles.role_name
           FROM information_schema.enabled_roles)) OR ((udt_privileges.grantee)::name IN ( SELECT enabled_roles.role_name
           FROM information_schema.enabled_roles)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("triggered_update_columns"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS trigger_catalog,
    (n.nspname)::information_schema.sql_identifier AS trigger_schema,
    (t.tgname)::information_schema.sql_identifier AS trigger_name,
    (current_database())::information_schema.sql_identifier AS event_object_catalog,
    (n.nspname)::information_schema.sql_identifier AS event_object_schema,
    (c.relname)::information_schema.sql_identifier AS event_object_table,
    (a.attname)::information_schema.sql_identifier AS event_object_column
   FROM pg_namespace n,
    pg_class c,
    pg_trigger t,
    ( SELECT ta0.tgoid,
            (ta0.tgat).x AS tgattnum,
            (ta0.tgat).n AS tgattpos
           FROM ( SELECT pg_trigger.oid AS tgoid,
                    information_schema._pg_expandarray(pg_trigger.tgattr) AS tgat
                   FROM pg_trigger) ta0) ta,
    pg_attribute a
  WHERE ((n.oid = c.relnamespace) AND (c.oid = t.tgrelid) AND (t.oid = ta.tgoid) AND ((a.attrelid = t.tgrelid) AND (a.attnum = ta.tgattnum)) AND (NOT t.tgisinternal) AND (NOT pg_is_other_temp_schema(n.oid)) AND (pg_has_role(c.relowner, 'USAGE'::text) OR has_column_privilege(c.oid, a.attnum, 'INSERT, UPDATE, REFERENCES'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("triggers"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS trigger_catalog,
    (n.nspname)::information_schema.sql_identifier AS trigger_schema,
    (t.tgname)::information_schema.sql_identifier AS trigger_name,
    (em.text)::information_schema.character_data AS event_manipulation,
    (current_database())::information_schema.sql_identifier AS event_object_catalog,
    (n.nspname)::information_schema.sql_identifier AS event_object_schema,
    (c.relname)::information_schema.sql_identifier AS event_object_table,
    (rank() OVER (PARTITION BY (n.nspname)::information_schema.sql_identifier, (c.relname)::information_schema.sql_identifier, em.num, ((t.tgtype)::integer & 1), ((t.tgtype)::integer & 66) ORDER BY t.tgname))::information_schema.cardinal_number AS action_order,
    (
        CASE
            WHEN pg_has_role(c.relowner, 'USAGE'::text) THEN (regexp_match(pg_get_triggerdef(t.oid), '.{35,} WHEN \((.+)\) EXECUTE FUNCTION'::text))[1]
            ELSE NULL::text
        END)::information_schema.character_data AS action_condition,
    (SUBSTRING(pg_get_triggerdef(t.oid) FROM (POSITION(('EXECUTE FUNCTION'::text) IN (SUBSTRING(pg_get_triggerdef(t.oid) FROM 48))) + 47)))::information_schema.character_data AS action_statement,
    (
        CASE ((t.tgtype)::integer & 1)
            WHEN 1 THEN 'ROW'::text
            ELSE 'STATEMENT'::text
        END)::information_schema.character_data AS action_orientation,
    (
        CASE ((t.tgtype)::integer & 66)
            WHEN 2 THEN 'BEFORE'::text
            WHEN 64 THEN 'INSTEAD OF'::text
            ELSE 'AFTER'::text
        END)::information_schema.character_data AS action_timing,
    (t.tgoldtable)::information_schema.sql_identifier AS action_reference_old_table,
    (t.tgnewtable)::information_schema.sql_identifier AS action_reference_new_table,
    (NULL::name)::information_schema.sql_identifier AS action_reference_old_row,
    (NULL::name)::information_schema.sql_identifier AS action_reference_new_row,
    (NULL::timestamp with time zone)::information_schema.time_stamp AS created
   FROM pg_namespace n,
    pg_class c,
    pg_trigger t,
    ( VALUES (4,'INSERT'::text), (8,'DELETE'::text), (16,'UPDATE'::text)) em(num, text)
  WHERE ((n.oid = c.relnamespace) AND (c.oid = t.tgrelid) AND (((t.tgtype)::integer & em.num) <> 0) AND (NOT t.tgisinternal) AND (NOT pg_is_other_temp_schema(n.oid)) AND (pg_has_role(c.relowner, 'USAGE'::text) OR has_table_privilege(c.oid, 'INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER'::text) OR has_any_column_privilege(c.oid, 'INSERT, UPDATE, REFERENCES'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("user_defined_types"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS user_defined_type_catalog,
    (n.nspname)::information_schema.sql_identifier AS user_defined_type_schema,
    (c.relname)::information_schema.sql_identifier AS user_defined_type_name,
    ('STRUCTURED'::character varying)::information_schema.character_data AS user_defined_type_category,
    ('YES'::character varying)::information_schema.yes_or_no AS is_instantiable,
    (NULL::character varying)::information_schema.yes_or_no AS is_final,
    (NULL::character varying)::information_schema.character_data AS ordering_form,
    (NULL::character varying)::information_schema.character_data AS ordering_category,
    (NULL::name)::information_schema.sql_identifier AS ordering_routine_catalog,
    (NULL::name)::information_schema.sql_identifier AS ordering_routine_schema,
    (NULL::name)::information_schema.sql_identifier AS ordering_routine_name,
    (NULL::character varying)::information_schema.character_data AS reference_type,
    (NULL::character varying)::information_schema.character_data AS data_type,
    (NULL::integer)::information_schema.cardinal_number AS character_maximum_length,
    (NULL::integer)::information_schema.cardinal_number AS character_octet_length,
    (NULL::name)::information_schema.sql_identifier AS character_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS character_set_schema,
    (NULL::name)::information_schema.sql_identifier AS character_set_name,
    (NULL::name)::information_schema.sql_identifier AS collation_catalog,
    (NULL::name)::information_schema.sql_identifier AS collation_schema,
    (NULL::name)::information_schema.sql_identifier AS collation_name,
    (NULL::integer)::information_schema.cardinal_number AS numeric_precision,
    (NULL::integer)::information_schema.cardinal_number AS numeric_precision_radix,
    (NULL::integer)::information_schema.cardinal_number AS numeric_scale,
    (NULL::integer)::information_schema.cardinal_number AS datetime_precision,
    (NULL::character varying)::information_schema.character_data AS interval_type,
    (NULL::integer)::information_schema.cardinal_number AS interval_precision,
    (NULL::name)::information_schema.sql_identifier AS source_dtd_identifier,
    (NULL::name)::information_schema.sql_identifier AS ref_dtd_identifier
   FROM pg_namespace n,
    pg_class c,
    pg_type t
  WHERE ((n.oid = c.relnamespace) AND (t.typrelid = c.oid) AND (c.relkind = 'c'::"char") AND (pg_has_role(t.typowner, 'USAGE'::text) OR has_type_privilege(t.oid, 'USAGE'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("usage_privileges"), T(r#" SELECT (u.rolname)::information_schema.sql_identifier AS grantor,
    ('PUBLIC'::name)::information_schema.sql_identifier AS grantee,
    (current_database())::information_schema.sql_identifier AS object_catalog,
    (n.nspname)::information_schema.sql_identifier AS object_schema,
    (c.collname)::information_schema.sql_identifier AS object_name,
    ('COLLATION'::character varying)::information_schema.character_data AS object_type,
    ('USAGE'::character varying)::information_schema.character_data AS privilege_type,
    ('NO'::character varying)::information_schema.yes_or_no AS is_grantable
   FROM pg_authid u,
    pg_namespace n,
    pg_collation c
  WHERE ((u.oid = c.collowner) AND (c.collnamespace = n.oid) AND (c.collencoding = ANY (ARRAY['-1'::integer, ( SELECT pg_database.encoding
           FROM pg_database
          WHERE (pg_database.datname = current_database()))])))
UNION ALL
 SELECT (u_grantor.rolname)::information_schema.sql_identifier AS grantor,
    (grantee.rolname)::information_schema.sql_identifier AS grantee,
    (current_database())::information_schema.sql_identifier AS object_catalog,
    (n.nspname)::information_schema.sql_identifier AS object_schema,
    (t.typname)::information_schema.sql_identifier AS object_name,
    ('DOMAIN'::character varying)::information_schema.character_data AS object_type,
    ('USAGE'::character varying)::information_schema.character_data AS privilege_type,
    (
        CASE
            WHEN (pg_has_role(grantee.oid, t.typowner, 'USAGE'::text) OR t.grantable) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_grantable
   FROM ( SELECT pg_type.oid,
            pg_type.typname,
            pg_type.typnamespace,
            pg_type.typtype,
            pg_type.typowner,
            (aclexplode(COALESCE(pg_type.typacl, acldefault('T'::"char", pg_type.typowner)))).grantor AS grantor,
            (aclexplode(COALESCE(pg_type.typacl, acldefault('T'::"char", pg_type.typowner)))).grantee AS grantee,
            (aclexplode(COALESCE(pg_type.typacl, acldefault('T'::"char", pg_type.typowner)))).privilege_type AS privilege_type,
            (aclexplode(COALESCE(pg_type.typacl, acldefault('T'::"char", pg_type.typowner)))).is_grantable AS is_grantable
           FROM pg_type) t(oid, typname, typnamespace, typtype, typowner, grantor, grantee, prtype, grantable),
    pg_namespace n,
    pg_authid u_grantor,
    ( SELECT pg_authid.oid,
            pg_authid.rolname
           FROM pg_authid
        UNION ALL
         SELECT (0)::oid AS oid,
            'PUBLIC'::name) grantee(oid, rolname)
  WHERE ((t.typnamespace = n.oid) AND (t.typtype = 'd'::"char") AND (t.grantee = grantee.oid) AND (t.grantor = u_grantor.oid) AND (t.prtype = 'USAGE'::text) AND (pg_has_role(u_grantor.oid, 'USAGE'::text) OR pg_has_role(grantee.oid, 'USAGE'::text) OR (grantee.rolname = 'PUBLIC'::name)))
UNION ALL
 SELECT (u_grantor.rolname)::information_schema.sql_identifier AS grantor,
    (grantee.rolname)::information_schema.sql_identifier AS grantee,
    (current_database())::information_schema.sql_identifier AS object_catalog,
    (''::name)::information_schema.sql_identifier AS object_schema,
    (fdw.fdwname)::information_schema.sql_identifier AS object_name,
    ('FOREIGN DATA WRAPPER'::character varying)::information_schema.character_data AS object_type,
    ('USAGE'::character varying)::information_schema.character_data AS privilege_type,
    (
        CASE
            WHEN (pg_has_role(grantee.oid, fdw.fdwowner, 'USAGE'::text) OR fdw.grantable) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_grantable
   FROM ( SELECT pg_foreign_data_wrapper.fdwname,
            pg_foreign_data_wrapper.fdwowner,
            (aclexplode(COALESCE(pg_foreign_data_wrapper.fdwacl, acldefault('F'::"char", pg_foreign_data_wrapper.fdwowner)))).grantor AS grantor,
            (aclexplode(COALESCE(pg_foreign_data_wrapper.fdwacl, acldefault('F'::"char", pg_foreign_data_wrapper.fdwowner)))).grantee AS grantee,
            (aclexplode(COALESCE(pg_foreign_data_wrapper.fdwacl, acldefault('F'::"char", pg_foreign_data_wrapper.fdwowner)))).privilege_type AS privilege_type,
            (aclexplode(COALESCE(pg_foreign_data_wrapper.fdwacl, acldefault('F'::"char", pg_foreign_data_wrapper.fdwowner)))).is_grantable AS is_grantable
           FROM pg_foreign_data_wrapper) fdw(fdwname, fdwowner, grantor, grantee, prtype, grantable),
    pg_authid u_grantor,
    ( SELECT pg_authid.oid,
            pg_authid.rolname
           FROM pg_authid
        UNION ALL
         SELECT (0)::oid AS oid,
            'PUBLIC'::name) grantee(oid, rolname)
  WHERE ((u_grantor.oid = fdw.grantor) AND (grantee.oid = fdw.grantee) AND (fdw.prtype = 'USAGE'::text) AND (pg_has_role(u_grantor.oid, 'USAGE'::text) OR pg_has_role(grantee.oid, 'USAGE'::text) OR (grantee.rolname = 'PUBLIC'::name)))
UNION ALL
 SELECT (u_grantor.rolname)::information_schema.sql_identifier AS grantor,
    (grantee.rolname)::information_schema.sql_identifier AS grantee,
    (current_database())::information_schema.sql_identifier AS object_catalog,
    (''::name)::information_schema.sql_identifier AS object_schema,
    (srv.srvname)::information_schema.sql_identifier AS object_name,
    ('FOREIGN SERVER'::character varying)::information_schema.character_data AS object_type,
    ('USAGE'::character varying)::information_schema.character_data AS privilege_type,
    (
        CASE
            WHEN (pg_has_role(grantee.oid, srv.srvowner, 'USAGE'::text) OR srv.grantable) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_grantable
   FROM ( SELECT pg_foreign_server.srvname,
            pg_foreign_server.srvowner,
            (aclexplode(COALESCE(pg_foreign_server.srvacl, acldefault('S'::"char", pg_foreign_server.srvowner)))).grantor AS grantor,
            (aclexplode(COALESCE(pg_foreign_server.srvacl, acldefault('S'::"char", pg_foreign_server.srvowner)))).grantee AS grantee,
            (aclexplode(COALESCE(pg_foreign_server.srvacl, acldefault('S'::"char", pg_foreign_server.srvowner)))).privilege_type AS privilege_type,
            (aclexplode(COALESCE(pg_foreign_server.srvacl, acldefault('S'::"char", pg_foreign_server.srvowner)))).is_grantable AS is_grantable
           FROM pg_foreign_server) srv(srvname, srvowner, grantor, grantee, prtype, grantable),
    pg_authid u_grantor,
    ( SELECT pg_authid.oid,
            pg_authid.rolname
           FROM pg_authid
        UNION ALL
         SELECT (0)::oid AS oid,
            'PUBLIC'::name) grantee(oid, rolname)
  WHERE ((u_grantor.oid = srv.grantor) AND (grantee.oid = srv.grantee) AND (srv.prtype = 'USAGE'::text) AND (pg_has_role(u_grantor.oid, 'USAGE'::text) OR pg_has_role(grantee.oid, 'USAGE'::text) OR (grantee.rolname = 'PUBLIC'::name)))
UNION ALL
 SELECT (u_grantor.rolname)::information_schema.sql_identifier AS grantor,
    (grantee.rolname)::information_schema.sql_identifier AS grantee,
    (current_database())::information_schema.sql_identifier AS object_catalog,
    (n.nspname)::information_schema.sql_identifier AS object_schema,
    (c.relname)::information_schema.sql_identifier AS object_name,
    ('SEQUENCE'::character varying)::information_schema.character_data AS object_type,
    ('USAGE'::character varying)::information_schema.character_data AS privilege_type,
    (
        CASE
            WHEN (pg_has_role(grantee.oid, c.relowner, 'USAGE'::text) OR c.grantable) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_grantable
   FROM ( SELECT pg_class.oid,
            pg_class.relname,
            pg_class.relnamespace,
            pg_class.relkind,
            pg_class.relowner,
            (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).grantor AS grantor,
            (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).grantee AS grantee,
            (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).privilege_type AS privilege_type,
            (aclexplode(COALESCE(pg_class.relacl, acldefault('r'::"char", pg_class.relowner)))).is_grantable AS is_grantable
           FROM pg_class) c(oid, relname, relnamespace, relkind, relowner, grantor, grantee, prtype, grantable),
    pg_namespace n,
    pg_authid u_grantor,
    ( SELECT pg_authid.oid,
            pg_authid.rolname
           FROM pg_authid
        UNION ALL
         SELECT (0)::oid AS oid,
            'PUBLIC'::name) grantee(oid, rolname)
  WHERE ((c.relnamespace = n.oid) AND (c.relkind = 'S'::"char") AND (c.grantee = grantee.oid) AND (c.grantor = u_grantor.oid) AND (c.prtype = 'USAGE'::text) AND (pg_has_role(u_grantor.oid, 'USAGE'::text) OR pg_has_role(grantee.oid, 'USAGE'::text) OR (grantee.rolname = 'PUBLIC'::name)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("role_usage_grants"), T(r#" SELECT usage_privileges.grantor,
    usage_privileges.grantee,
    usage_privileges.object_catalog,
    usage_privileges.object_schema,
    usage_privileges.object_name,
    usage_privileges.object_type,
    usage_privileges.privilege_type,
    usage_privileges.is_grantable
   FROM information_schema.usage_privileges
  WHERE (((usage_privileges.grantor)::name IN ( SELECT enabled_roles.role_name
           FROM information_schema.enabled_roles)) OR ((usage_privileges.grantee)::name IN ( SELECT enabled_roles.role_name
           FROM information_schema.enabled_roles)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("view_column_usage"), T(r#" SELECT DISTINCT (current_database())::information_schema.sql_identifier AS view_catalog,
    (nv.nspname)::information_schema.sql_identifier AS view_schema,
    (v.relname)::information_schema.sql_identifier AS view_name,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (nt.nspname)::information_schema.sql_identifier AS table_schema,
    (t.relname)::information_schema.sql_identifier AS table_name,
    (a.attname)::information_schema.sql_identifier AS column_name
   FROM pg_namespace nv,
    pg_class v,
    pg_depend dv,
    pg_depend dt,
    pg_class t,
    pg_namespace nt,
    pg_attribute a
  WHERE ((nv.oid = v.relnamespace) AND (v.relkind = 'v'::"char") AND (v.oid = dv.refobjid) AND (dv.refclassid = ('pg_class'::regclass)::oid) AND (dv.classid = ('pg_rewrite'::regclass)::oid) AND (dv.deptype = 'i'::"char") AND (dv.objid = dt.objid) AND (dv.refobjid <> dt.refobjid) AND (dt.classid = ('pg_rewrite'::regclass)::oid) AND (dt.refclassid = ('pg_class'::regclass)::oid) AND (dt.refobjid = t.oid) AND (t.relnamespace = nt.oid) AND (t.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"])) AND (t.oid = a.attrelid) AND (dt.refobjsubid = a.attnum) AND pg_has_role(t.relowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("view_table_usage"), T(r#" SELECT DISTINCT (current_database())::information_schema.sql_identifier AS view_catalog,
    (nv.nspname)::information_schema.sql_identifier AS view_schema,
    (v.relname)::information_schema.sql_identifier AS view_name,
    (current_database())::information_schema.sql_identifier AS table_catalog,
    (nt.nspname)::information_schema.sql_identifier AS table_schema,
    (t.relname)::information_schema.sql_identifier AS table_name
   FROM pg_namespace nv,
    pg_class v,
    pg_depend dv,
    pg_depend dt,
    pg_class t,
    pg_namespace nt
  WHERE ((nv.oid = v.relnamespace) AND (v.relkind = 'v'::"char") AND (v.oid = dv.refobjid) AND (dv.refclassid = ('pg_class'::regclass)::oid) AND (dv.classid = ('pg_rewrite'::regclass)::oid) AND (dv.deptype = 'i'::"char") AND (dv.objid = dt.objid) AND (dv.refobjid <> dt.refobjid) AND (dt.classid = ('pg_rewrite'::regclass)::oid) AND (dt.refclassid = ('pg_class'::regclass)::oid) AND (dt.refobjid = t.oid) AND (t.relnamespace = nt.oid) AND (t.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'p'::"char"])) AND pg_has_role(t.relowner, 'USAGE'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("views"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS table_catalog,
    (nc.nspname)::information_schema.sql_identifier AS table_schema,
    (c.relname)::information_schema.sql_identifier AS table_name,
    (
        CASE
            WHEN pg_has_role(c.relowner, 'USAGE'::text) THEN pg_get_viewdef(c.oid)
            ELSE NULL::text
        END)::information_schema.character_data AS view_definition,
    (
        CASE
            WHEN ('check_option=cascaded'::text = ANY (c.reloptions)) THEN 'CASCADED'::text
            WHEN ('check_option=local'::text = ANY (c.reloptions)) THEN 'LOCAL'::text
            ELSE 'NONE'::text
        END)::information_schema.character_data AS check_option,
    (
        CASE
            WHEN ((pg_relation_is_updatable((c.oid)::regclass, false) & 20) = 20) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_updatable,
    (
        CASE
            WHEN ((pg_relation_is_updatable((c.oid)::regclass, false) & 8) = 8) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_insertable_into,
    (
        CASE
            WHEN (EXISTS ( SELECT 1
               FROM pg_trigger
              WHERE ((pg_trigger.tgrelid = c.oid) AND (((pg_trigger.tgtype)::integer & 81) = 81)))) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_trigger_updatable,
    (
        CASE
            WHEN (EXISTS ( SELECT 1
               FROM pg_trigger
              WHERE ((pg_trigger.tgrelid = c.oid) AND (((pg_trigger.tgtype)::integer & 73) = 73)))) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_trigger_deletable,
    (
        CASE
            WHEN (EXISTS ( SELECT 1
               FROM pg_trigger
              WHERE ((pg_trigger.tgrelid = c.oid) AND (((pg_trigger.tgtype)::integer & 69) = 69)))) THEN 'YES'::text
            ELSE 'NO'::text
        END)::information_schema.yes_or_no AS is_trigger_insertable_into
   FROM pg_namespace nc,
    pg_class c
  WHERE ((c.relnamespace = nc.oid) AND (c.relkind = 'v'::"char") AND (NOT pg_is_other_temp_schema(nc.oid)) AND (pg_has_role(c.relowner, 'USAGE'::text) OR has_table_privilege(c.oid, 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER'::text) OR has_any_column_privilege(c.oid, 'SELECT, INSERT, UPDATE, REFERENCES'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("data_type_privileges"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS object_catalog,
    x.objschema AS object_schema,
    x.objname AS object_name,
    (x.objtype)::information_schema.character_data AS object_type,
    x.objdtdid AS dtd_identifier
   FROM ( SELECT attributes.udt_schema,
            attributes.udt_name,
            'USER-DEFINED TYPE'::text AS text,
            attributes.dtd_identifier
           FROM information_schema.attributes
        UNION ALL
         SELECT columns.table_schema,
            columns.table_name,
            'TABLE'::text AS text,
            columns.dtd_identifier
           FROM information_schema.columns
        UNION ALL
         SELECT domains.domain_schema,
            domains.domain_name,
            'DOMAIN'::text AS text,
            domains.dtd_identifier
           FROM information_schema.domains
        UNION ALL
         SELECT parameters.specific_schema,
            parameters.specific_name,
            'ROUTINE'::text AS text,
            parameters.dtd_identifier
           FROM information_schema.parameters
        UNION ALL
         SELECT routines.specific_schema,
            routines.specific_name,
            'ROUTINE'::text AS text,
            routines.dtd_identifier
           FROM information_schema.routines) x(objschema, objname, objtype, objdtdid);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("_pg_foreign_table_columns"), T(r#" SELECT n.nspname,
    c.relname,
    a.attname,
    a.attfdwoptions
   FROM pg_foreign_table t,
    pg_authid u,
    pg_namespace n,
    pg_class c,
    pg_attribute a
  WHERE ((u.oid = c.relowner) AND (pg_has_role(c.relowner, 'USAGE'::text) OR has_column_privilege(c.oid, a.attnum, 'SELECT, INSERT, UPDATE, REFERENCES'::text)) AND (n.oid = c.relnamespace) AND (c.oid = t.ftrelid) AND (c.relkind = 'f'::"char") AND (a.attrelid = c.oid) AND (a.attnum > 0));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("element_types"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS object_catalog,
    (n.nspname)::information_schema.sql_identifier AS object_schema,
    x.objname AS object_name,
    (x.objtype)::information_schema.character_data AS object_type,
    (x.objdtdid)::information_schema.sql_identifier AS collection_type_identifier,
    (
        CASE
            WHEN (nbt.nspname = 'pg_catalog'::name) THEN format_type(bt.oid, NULL::integer)
            ELSE 'USER-DEFINED'::text
        END)::information_schema.character_data AS data_type,
    (NULL::integer)::information_schema.cardinal_number AS character_maximum_length,
    (NULL::integer)::information_schema.cardinal_number AS character_octet_length,
    (NULL::name)::information_schema.sql_identifier AS character_set_catalog,
    (NULL::name)::information_schema.sql_identifier AS character_set_schema,
    (NULL::name)::information_schema.sql_identifier AS character_set_name,
    (
        CASE
            WHEN (nco.nspname IS NOT NULL) THEN current_database()
            ELSE NULL::name
        END)::information_schema.sql_identifier AS collation_catalog,
    (nco.nspname)::information_schema.sql_identifier AS collation_schema,
    (co.collname)::information_schema.sql_identifier AS collation_name,
    (NULL::integer)::information_schema.cardinal_number AS numeric_precision,
    (NULL::integer)::information_schema.cardinal_number AS numeric_precision_radix,
    (NULL::integer)::information_schema.cardinal_number AS numeric_scale,
    (NULL::integer)::information_schema.cardinal_number AS datetime_precision,
    (NULL::character varying)::information_schema.character_data AS interval_type,
    (NULL::integer)::information_schema.cardinal_number AS interval_precision,
    (NULL::character varying)::information_schema.character_data AS domain_default,
    (current_database())::information_schema.sql_identifier AS udt_catalog,
    (nbt.nspname)::information_schema.sql_identifier AS udt_schema,
    (bt.typname)::information_schema.sql_identifier AS udt_name,
    (NULL::name)::information_schema.sql_identifier AS scope_catalog,
    (NULL::name)::information_schema.sql_identifier AS scope_schema,
    (NULL::name)::information_schema.sql_identifier AS scope_name,
    (NULL::integer)::information_schema.cardinal_number AS maximum_cardinality,
    (('a'::text || (x.objdtdid)::text))::information_schema.sql_identifier AS dtd_identifier
   FROM pg_namespace n,
    pg_type at,
    pg_namespace nbt,
    pg_type bt,
    (( SELECT c.relnamespace,
            (c.relname)::information_schema.sql_identifier AS relname,
                CASE
                    WHEN (c.relkind = 'c'::"char") THEN 'USER-DEFINED TYPE'::text
                    ELSE 'TABLE'::text
                END AS "case",
            a.attnum,
            a.atttypid,
            a.attcollation
           FROM pg_class c,
            pg_attribute a
          WHERE ((c.oid = a.attrelid) AND (c.relkind = ANY (ARRAY['r'::"char", 'v'::"char", 'f'::"char", 'c'::"char", 'p'::"char"])) AND (a.attnum > 0) AND (NOT a.attisdropped))
        UNION ALL
         SELECT t.typnamespace,
            (t.typname)::information_schema.sql_identifier AS typname,
            'DOMAIN'::text AS text,
            1,
            t.typbasetype,
            t.typcollation
           FROM pg_type t
          WHERE (t.typtype = 'd'::"char")
        UNION ALL
         SELECT ss.pronamespace,
            (nameconcatoid(ss.proname, ss.oid))::information_schema.sql_identifier AS nameconcatoid,
            'ROUTINE'::text AS text,
            (ss.x).n AS n,
            (ss.x).x AS x,
            0
           FROM ( SELECT p.pronamespace,
                    p.proname,
                    p.oid,
                    information_schema._pg_expandarray(COALESCE(p.proallargtypes, (p.proargtypes)::oid[])) AS x
                   FROM pg_proc p) ss
        UNION ALL
         SELECT p.pronamespace,
            (nameconcatoid(p.proname, p.oid))::information_schema.sql_identifier AS nameconcatoid,
            'ROUTINE'::text AS text,
            0,
            p.prorettype,
            0
           FROM pg_proc p) x(objschema, objname, objtype, objdtdid, objtypeid, objcollation)
     LEFT JOIN (pg_collation co
     JOIN pg_namespace nco ON ((co.collnamespace = nco.oid))) ON (((x.objcollation = co.oid) AND ((nco.nspname <> 'pg_catalog'::name) OR (co.collname <> 'default'::name)))))
  WHERE ((n.oid = x.objschema) AND (at.oid = x.objtypeid) AND ((at.typelem <> (0)::oid) AND (at.typlen = '-1'::integer)) AND (at.typelem = bt.oid) AND (nbt.oid = bt.typnamespace) AND ((n.nspname, (x.objname)::name, x.objtype, ((x.objdtdid)::information_schema.sql_identifier)::name) IN ( SELECT data_type_privileges.object_schema,
            data_type_privileges.object_name,
            data_type_privileges.object_type,
            data_type_privileges.dtd_identifier
           FROM information_schema.data_type_privileges)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("column_options"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS table_catalog,
    (c.nspname)::information_schema.sql_identifier AS table_schema,
    (c.relname)::information_schema.sql_identifier AS table_name,
    (c.attname)::information_schema.sql_identifier AS column_name,
    ((pg_options_to_table(c.attfdwoptions)).option_name)::information_schema.sql_identifier AS option_name,
    ((pg_options_to_table(c.attfdwoptions)).option_value)::information_schema.character_data AS option_value
   FROM information_schema._pg_foreign_table_columns c;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("_pg_foreign_data_wrappers"), T(r#" SELECT w.oid,
    w.fdwowner,
    w.fdwoptions,
    (current_database())::information_schema.sql_identifier AS foreign_data_wrapper_catalog,
    (w.fdwname)::information_schema.sql_identifier AS foreign_data_wrapper_name,
    (u.rolname)::information_schema.sql_identifier AS authorization_identifier,
    ('c'::character varying)::information_schema.character_data AS foreign_data_wrapper_language
   FROM pg_foreign_data_wrapper w,
    pg_authid u
  WHERE ((u.oid = w.fdwowner) AND (pg_has_role(w.fdwowner, 'USAGE'::text) OR has_foreign_data_wrapper_privilege(w.oid, 'USAGE'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("foreign_data_wrapper_options"), T(r#" SELECT w.foreign_data_wrapper_catalog,
    w.foreign_data_wrapper_name,
    ((pg_options_to_table(w.fdwoptions)).option_name)::information_schema.sql_identifier AS option_name,
    ((pg_options_to_table(w.fdwoptions)).option_value)::information_schema.character_data AS option_value
   FROM information_schema._pg_foreign_data_wrappers w;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("foreign_data_wrappers"), T(r#" SELECT w.foreign_data_wrapper_catalog,
    w.foreign_data_wrapper_name,
    w.authorization_identifier,
    (NULL::character varying)::information_schema.character_data AS library_name,
    w.foreign_data_wrapper_language
   FROM information_schema._pg_foreign_data_wrappers w;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("_pg_foreign_servers"), T(r#" SELECT s.oid,
    s.srvoptions,
    (current_database())::information_schema.sql_identifier AS foreign_server_catalog,
    (s.srvname)::information_schema.sql_identifier AS foreign_server_name,
    (current_database())::information_schema.sql_identifier AS foreign_data_wrapper_catalog,
    (w.fdwname)::information_schema.sql_identifier AS foreign_data_wrapper_name,
    (s.srvtype)::information_schema.character_data AS foreign_server_type,
    (s.srvversion)::information_schema.character_data AS foreign_server_version,
    (u.rolname)::information_schema.sql_identifier AS authorization_identifier
   FROM pg_foreign_server s,
    pg_foreign_data_wrapper w,
    pg_authid u
  WHERE ((w.oid = s.srvfdw) AND (u.oid = s.srvowner) AND (pg_has_role(s.srvowner, 'USAGE'::text) OR has_server_privilege(s.oid, 'USAGE'::text)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("foreign_server_options"), T(r#" SELECT s.foreign_server_catalog,
    s.foreign_server_name,
    ((pg_options_to_table(s.srvoptions)).option_name)::information_schema.sql_identifier AS option_name,
    ((pg_options_to_table(s.srvoptions)).option_value)::information_schema.character_data AS option_value
   FROM information_schema._pg_foreign_servers s;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("foreign_servers"), T(r#" SELECT _pg_foreign_servers.foreign_server_catalog,
    _pg_foreign_servers.foreign_server_name,
    _pg_foreign_servers.foreign_data_wrapper_catalog,
    _pg_foreign_servers.foreign_data_wrapper_name,
    _pg_foreign_servers.foreign_server_type,
    _pg_foreign_servers.foreign_server_version,
    _pg_foreign_servers.authorization_identifier
   FROM information_schema._pg_foreign_servers;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("_pg_foreign_tables"), T(r#" SELECT (current_database())::information_schema.sql_identifier AS foreign_table_catalog,
    (n.nspname)::information_schema.sql_identifier AS foreign_table_schema,
    (c.relname)::information_schema.sql_identifier AS foreign_table_name,
    t.ftoptions,
    (current_database())::information_schema.sql_identifier AS foreign_server_catalog,
    (s.srvname)::information_schema.sql_identifier AS foreign_server_name,
    (u.rolname)::information_schema.sql_identifier AS authorization_identifier
   FROM pg_foreign_table t,
    pg_foreign_server s,
    pg_foreign_data_wrapper w,
    pg_authid u,
    pg_namespace n,
    pg_class c
  WHERE ((w.oid = s.srvfdw) AND (u.oid = c.relowner) AND (pg_has_role(c.relowner, 'USAGE'::text) OR has_table_privilege(c.oid, 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER'::text) OR has_any_column_privilege(c.oid, 'SELECT, INSERT, UPDATE, REFERENCES'::text)) AND (n.oid = c.relnamespace) AND (c.oid = t.ftrelid) AND (c.relkind = 'f'::"char") AND (s.oid = t.ftserver));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("foreign_table_options"), T(r#" SELECT t.foreign_table_catalog,
    t.foreign_table_schema,
    t.foreign_table_name,
    ((pg_options_to_table(t.ftoptions)).option_name)::information_schema.sql_identifier AS option_name,
    ((pg_options_to_table(t.ftoptions)).option_value)::information_schema.character_data AS option_value
   FROM information_schema._pg_foreign_tables t;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("foreign_tables"), T(r#" SELECT _pg_foreign_tables.foreign_table_catalog,
    _pg_foreign_tables.foreign_table_schema,
    _pg_foreign_tables.foreign_table_name,
    _pg_foreign_tables.foreign_server_catalog,
    _pg_foreign_tables.foreign_server_name
   FROM information_schema._pg_foreign_tables;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("_pg_user_mappings"), T(r#" SELECT um.oid,
    um.umoptions,
    um.umuser,
    (COALESCE(u.rolname, 'PUBLIC'::name))::information_schema.sql_identifier AS authorization_identifier,
    s.foreign_server_catalog,
    s.foreign_server_name,
    s.authorization_identifier AS srvowner
   FROM (pg_user_mapping um
     LEFT JOIN pg_authid u ON ((u.oid = um.umuser))),
    information_schema._pg_foreign_servers s
  WHERE (s.oid = um.umserver);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("information_schema"), T("user_mapping_options"), T(r#" SELECT um.authorization_identifier,
    um.foreign_server_catalog,
    um.foreign_server_name,
    (opts.option_name)::information_schema.sql_identifier AS option_name,
    (
        CASE
            WHEN (((um.umuser <> (0)::oid) AND ((um.authorization_identifier)::name = CURRENT_USER)) OR ((um.umuser = (0)::oid) AND pg_has_role((um.srvowner)::name, 'USAGE'::text)) OR ( SELECT pg_authid.rolsuper
               FROM pg_authid
              WHERE (pg_authid.rolname = CURRENT_USER))) THEN opts.option_value
            ELSE NULL::text
        END)::information_schema.character_data AS option_value
   FROM information_schema._pg_user_mappings um,
    LATERAL pg_options_to_table(um.umoptions) opts(option_name, option_value);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_archiver"), T(r#" SELECT s.archived_count,
    s.last_archived_wal,
    s.last_archived_time,
    s.failed_count,
    s.last_failed_wal,
    s.last_failed_time,
    s.stats_reset
   FROM pg_stat_get_archiver() s(archived_count, last_archived_wal, last_archived_time, failed_count, last_failed_wal, last_failed_time, stats_reset);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_bgwriter"), T(r#" SELECT pg_stat_get_bgwriter_timed_checkpoints() AS checkpoints_timed,
    pg_stat_get_bgwriter_requested_checkpoints() AS checkpoints_req,
    pg_stat_get_checkpoint_write_time() AS checkpoint_write_time,
    pg_stat_get_checkpoint_sync_time() AS checkpoint_sync_time,
    pg_stat_get_bgwriter_buf_written_checkpoints() AS buffers_checkpoint,
    pg_stat_get_bgwriter_buf_written_clean() AS buffers_clean,
    pg_stat_get_bgwriter_maxwritten_clean() AS maxwritten_clean,
    pg_stat_get_buf_written_backend() AS buffers_backend,
    pg_stat_get_buf_fsync_backend() AS buffers_backend_fsync,
    pg_stat_get_buf_alloc() AS buffers_alloc,
    pg_stat_get_bgwriter_stat_reset_time() AS stats_reset;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_wal"), T(r#" SELECT w.wal_records,
    w.wal_fpi,
    w.wal_bytes,
    w.wal_buffers_full,
    w.wal_write,
    w.wal_sync,
    w.wal_write_time,
    w.wal_sync_time,
    w.stats_reset
   FROM pg_stat_get_wal() w(wal_records, wal_fpi, wal_bytes, wal_buffers_full, wal_write, wal_sync, wal_write_time, wal_sync_time, stats_reset);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_progress_analyze"), T(r#" SELECT s.pid,
    s.datid,
    d.datname,
    s.relid,
        CASE s.param1
            WHEN 0 THEN 'initializing'::text
            WHEN 1 THEN 'acquiring sample rows'::text
            WHEN 2 THEN 'acquiring inherited sample rows'::text
            WHEN 3 THEN 'computing statistics'::text
            WHEN 4 THEN 'computing extended statistics'::text
            WHEN 5 THEN 'finalizing analyze'::text
            ELSE NULL::text
        END AS phase,
    s.param2 AS sample_blks_total,
    s.param3 AS sample_blks_scanned,
    s.param4 AS ext_stats_total,
    s.param5 AS ext_stats_computed,
    s.param6 AS child_tables_total,
    s.param7 AS child_tables_done,
    (s.param8)::oid AS current_child_table_relid
   FROM (pg_stat_get_progress_info('ANALYZE'::text) s(pid, datid, relid, param1, param2, param3, param4, param5, param6, param7, param8, param9, param10, param11, param12, param13, param14, param15, param16, param17, param18, param19, param20)
     LEFT JOIN pg_database d ON ((s.datid = d.oid)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_progress_vacuum"), T(r#" SELECT s.pid,
    s.datid,
    d.datname,
    s.relid,
        CASE s.param1
            WHEN 0 THEN 'initializing'::text
            WHEN 1 THEN 'scanning heap'::text
            WHEN 2 THEN 'vacuuming indexes'::text
            WHEN 3 THEN 'vacuuming heap'::text
            WHEN 4 THEN 'cleaning up indexes'::text
            WHEN 5 THEN 'truncating heap'::text
            WHEN 6 THEN 'performing final cleanup'::text
            ELSE NULL::text
        END AS phase,
    s.param2 AS heap_blks_total,
    s.param3 AS heap_blks_scanned,
    s.param4 AS heap_blks_vacuumed,
    s.param5 AS index_vacuum_count,
    s.param6 AS max_dead_tuples,
    s.param7 AS num_dead_tuples
   FROM (pg_stat_get_progress_info('VACUUM'::text) s(pid, datid, relid, param1, param2, param3, param4, param5, param6, param7, param8, param9, param10, param11, param12, param13, param14, param15, param16, param17, param18, param19, param20)
     LEFT JOIN pg_database d ON ((s.datid = d.oid)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_progress_create_index"), T(r#" SELECT s.pid,
    s.datid,
    d.datname,
    s.relid,
    (s.param7)::oid AS index_relid,
        CASE s.param1
            WHEN 1 THEN 'CREATE INDEX'::text
            WHEN 2 THEN 'CREATE INDEX CONCURRENTLY'::text
            WHEN 3 THEN 'REINDEX'::text
            WHEN 4 THEN 'REINDEX CONCURRENTLY'::text
            ELSE NULL::text
        END AS command,
        CASE s.param10
            WHEN 0 THEN 'initializing'::text
            WHEN 1 THEN 'waiting for writers before build'::text
            WHEN 2 THEN ('building index'::text || COALESCE((': '::text || pg_indexam_progress_phasename((s.param9)::oid, s.param11)), ''::text))
            WHEN 3 THEN 'waiting for writers before validation'::text
            WHEN 4 THEN 'index validation: scanning index'::text
            WHEN 5 THEN 'index validation: sorting tuples'::text
            WHEN 6 THEN 'index validation: scanning table'::text
            WHEN 7 THEN 'waiting for old snapshots'::text
            WHEN 8 THEN 'waiting for readers before marking dead'::text
            WHEN 9 THEN 'waiting for readers before dropping'::text
            ELSE NULL::text
        END AS phase,
    s.param4 AS lockers_total,
    s.param5 AS lockers_done,
    s.param6 AS current_locker_pid,
    s.param16 AS blocks_total,
    s.param17 AS blocks_done,
    s.param12 AS tuples_total,
    s.param13 AS tuples_done,
    s.param14 AS partitions_total,
    s.param15 AS partitions_done
   FROM (pg_stat_get_progress_info('CREATE INDEX'::text) s(pid, datid, relid, param1, param2, param3, param4, param5, param6, param7, param8, param9, param10, param11, param12, param13, param14, param15, param16, param17, param18, param19, param20)
     LEFT JOIN pg_database d ON ((s.datid = d.oid)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_progress_basebackup"), T(r#" SELECT s.pid,
        CASE s.param1
            WHEN 0 THEN 'initializing'::text
            WHEN 1 THEN 'waiting for checkpoint to finish'::text
            WHEN 2 THEN 'estimating backup size'::text
            WHEN 3 THEN 'streaming database files'::text
            WHEN 4 THEN 'waiting for wal archiving to finish'::text
            WHEN 5 THEN 'transferring wal files'::text
            ELSE NULL::text
        END AS phase,
        CASE s.param2
            WHEN '-1'::integer THEN NULL::bigint
            ELSE s.param2
        END AS backup_total,
    s.param3 AS backup_streamed,
    s.param4 AS tablespaces_total,
    s.param5 AS tablespaces_streamed
   FROM pg_stat_get_progress_info('BASEBACKUP'::text) s(pid, datid, relid, param1, param2, param3, param4, param5, param6, param7, param8, param9, param10, param11, param12, param13, param14, param15, param16, param17, param18, param19, param20);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_progress_copy"), T(r#" SELECT s.pid,
    s.datid,
    d.datname,
    s.relid,
        CASE s.param5
            WHEN 1 THEN 'COPY FROM'::text
            WHEN 2 THEN 'COPY TO'::text
            ELSE NULL::text
        END AS command,
        CASE s.param6
            WHEN 1 THEN 'FILE'::text
            WHEN 2 THEN 'PROGRAM'::text
            WHEN 3 THEN 'PIPE'::text
            WHEN 4 THEN 'CALLBACK'::text
            ELSE NULL::text
        END AS type,
    s.param1 AS bytes_processed,
    s.param2 AS bytes_total,
    s.param3 AS tuples_processed,
    s.param4 AS tuples_excluded
   FROM (pg_stat_get_progress_info('COPY'::text) s(pid, datid, relid, param1, param2, param3, param4, param5, param6, param7, param8, param9, param10, param11, param12, param13, param14, param15, param16, param17, param18, param19, param20)
     LEFT JOIN pg_database d ON ((s.datid = d.oid)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_user_mappings"), T(r#" SELECT u.oid AS umid,
    s.oid AS srvid,
    s.srvname,
    u.umuser,
        CASE
            WHEN (u.umuser = (0)::oid) THEN 'public'::name
            ELSE a.rolname
        END AS usename,
        CASE
            WHEN (((u.umuser <> (0)::oid) AND (a.rolname = CURRENT_USER) AND (pg_has_role(s.srvowner, 'USAGE'::text) OR has_server_privilege(s.oid, 'USAGE'::text))) OR ((u.umuser = (0)::oid) AND pg_has_role(s.srvowner, 'USAGE'::text)) OR ( SELECT pg_authid.rolsuper
               FROM pg_authid
              WHERE (pg_authid.rolname = CURRENT_USER))) THEN u.umoptions
            ELSE NULL::text[]
        END AS umoptions
   FROM ((pg_user_mapping u
     JOIN pg_foreign_server s ON ((u.umserver = s.oid)))
     LEFT JOIN pg_authid a ON ((a.oid = u.umuser)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_subscription_stats"), T(r#" SELECT ss.subid,
    s.subname,
    ss.apply_error_count,
    ss.sync_error_count,
    ss.stats_reset
   FROM pg_subscription s,
    LATERAL pg_stat_get_subscription_stats(s.oid) ss(subid, apply_error_count, sync_error_count, stats_reset);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_progress_cluster"), T(r#" SELECT s.pid,
    s.datid,
    d.datname,
    s.relid,
        CASE s.param1
            WHEN 1 THEN 'CLUSTER'::text
            WHEN 2 THEN 'VACUUM FULL'::text
            ELSE NULL::text
        END AS command,
        CASE s.param2
            WHEN 0 THEN 'initializing'::text
            WHEN 1 THEN 'seq scanning heap'::text
            WHEN 2 THEN 'index scanning heap'::text
            WHEN 3 THEN 'sorting tuples'::text
            WHEN 4 THEN 'writing new heap'::text
            WHEN 5 THEN 'swapping relation files'::text
            WHEN 6 THEN 'rebuilding index'::text
            WHEN 7 THEN 'performing final cleanup'::text
            ELSE NULL::text
        END AS phase,
    (s.param3)::oid AS cluster_index_relid,
    s.param4 AS heap_tuples_scanned,
    s.param5 AS heap_tuples_written,
    s.param6 AS heap_blks_total,
    s.param7 AS heap_blks_scanned,
    s.param8 AS index_rebuild_count
   FROM (pg_stat_get_progress_info('CLUSTER'::text) s(pid, datid, relid, param1, param2, param3, param4, param5, param6, param7, param8, param9, param10, param11, param12, param13, param14, param15, param16, param17, param18, param19, param20)
     LEFT JOIN pg_database d ON ((s.datid = d.oid)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_shadow"), T(r#" SELECT pg_authid.rolname AS usename,
    pg_authid.oid AS usesysid,
    pg_authid.rolcreatedb AS usecreatedb,
    pg_authid.rolsuper AS usesuper,
    pg_authid.rolreplication AS userepl,
    pg_authid.rolbypassrls AS usebypassrls,
    pg_authid.rolpassword AS passwd,
    pg_authid.rolvaliduntil AS valuntil,
    s.setconfig AS useconfig
   FROM (pg_authid
     LEFT JOIN pg_db_role_setting s ON (((pg_authid.oid = s.setrole) AND (s.setdatabase = (0)::oid))))
  WHERE pg_authid.rolcanlogin;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_roles"), T(r#" SELECT pg_authid.rolname,
    pg_authid.rolsuper,
    pg_authid.rolinherit,
    pg_authid.rolcreaterole,
    pg_authid.rolcreatedb,
    pg_authid.rolcanlogin,
    pg_authid.rolreplication,
    pg_authid.rolconnlimit,
    '********'::text AS rolpassword,
    pg_authid.rolvaliduntil,
    pg_authid.rolbypassrls,
    s.setconfig AS rolconfig,
    pg_authid.oid
   FROM (pg_authid
     LEFT JOIN pg_db_role_setting s ON (((pg_authid.oid = s.setrole) AND (s.setdatabase = (0)::oid))));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_settings"), T(r#" SELECT a.name,
    a.setting,
    a.unit,
    a.category,
    a.short_desc,
    a.extra_desc,
    a.context,
    a.vartype,
    a.source,
    a.min_val,
    a.max_val,
    a.enumvals,
    a.boot_val,
    a.reset_val,
    a.sourcefile,
    a.sourceline,
    a.pending_restart
   FROM pg_show_all_settings() a(name, setting, unit, category, short_desc, extra_desc, context, vartype, source, min_val, max_val, enumvals, boot_val, reset_val, sourcefile, sourceline, pending_restart);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_file_settings"), T(r#" SELECT a.sourcefile,
    a.sourceline,
    a.seqno,
    a.name,
    a.setting,
    a.applied,
    a.error
   FROM pg_show_all_file_settings() a(sourcefile, sourceline, seqno, name, setting, applied, error);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_hba_file_rules"), T(r#" SELECT a.line_number,
    a.type,
    a.database,
    a.user_name,
    a.address,
    a.netmask,
    a.auth_method,
    a.options,
    a.error
   FROM pg_hba_file_rules() a(line_number, type, database, user_name, address, netmask, auth_method, options, error);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_ident_file_mappings"), T(r#" SELECT a.line_number,
    a.map_name,
    a.sys_name,
    a.pg_username,
    a.error
   FROM pg_ident_file_mappings() a(line_number, map_name, sys_name, pg_username, error);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_config"), T(r#" SELECT pg_config.name,
    pg_config.setting
   FROM pg_config() pg_config(name, setting);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_shmem_allocations"), T(r#" SELECT pg_get_shmem_allocations.name,
    pg_get_shmem_allocations.off,
    pg_get_shmem_allocations.size,
    pg_get_shmem_allocations.allocated_size
   FROM pg_get_shmem_allocations() pg_get_shmem_allocations(name, off, size, allocated_size);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_backend_memory_contexts"), T(r#" SELECT pg_get_backend_memory_contexts.name,
    pg_get_backend_memory_contexts.ident,
    pg_get_backend_memory_contexts.parent,
    pg_get_backend_memory_contexts.level,
    pg_get_backend_memory_contexts.total_bytes,
    pg_get_backend_memory_contexts.total_nblocks,
    pg_get_backend_memory_contexts.free_bytes,
    pg_get_backend_memory_contexts.free_chunks,
    pg_get_backend_memory_contexts.used_bytes
   FROM pg_get_backend_memory_contexts() pg_get_backend_memory_contexts(name, ident, parent, level, total_bytes, total_nblocks, free_bytes, free_chunks, used_bytes);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_activity"), T(r#" SELECT s.datid,
    d.datname,
    s.pid,
    s.leader_pid,
    s.usesysid,
    u.rolname AS usename,
    s.application_name,
    s.client_addr,
    s.client_hostname,
    s.client_port,
    s.backend_start,
    s.xact_start,
    s.query_start,
    s.state_change,
    s.wait_event_type,
    s.wait_event,
    s.state,
    s.backend_xid,
    s.backend_xmin,
    s.query_id,
    s.query,
    s.backend_type
   FROM ((pg_stat_get_activity(NULL::integer) s(datid, pid, usesysid, application_name, state, query, wait_event_type, wait_event, xact_start, query_start, backend_start, state_change, client_addr, client_hostname, client_port, backend_xid, backend_xmin, backend_type, ssl, sslversion, sslcipher, sslbits, ssl_client_dn, ssl_client_serial, ssl_issuer_dn, gss_auth, gss_princ, gss_enc, leader_pid, query_id)
     LEFT JOIN pg_database d ON ((s.datid = d.oid)))
     LEFT JOIN pg_authid u ON ((s.usesysid = u.oid)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_replication_origin_status"), T(r#" SELECT pg_show_replication_origin_status.local_id,
    pg_show_replication_origin_status.external_id,
    pg_show_replication_origin_status.remote_lsn,
    pg_show_replication_origin_status.local_lsn
   FROM pg_show_replication_origin_status() pg_show_replication_origin_status(local_id, external_id, remote_lsn, local_lsn);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_replication"), T(r#" SELECT s.pid,
    s.usesysid,
    u.rolname AS usename,
    s.application_name,
    s.client_addr,
    s.client_hostname,
    s.client_port,
    s.backend_start,
    s.backend_xmin,
    w.state,
    w.sent_lsn,
    w.write_lsn,
    w.flush_lsn,
    w.replay_lsn,
    w.write_lag,
    w.flush_lag,
    w.replay_lag,
    w.sync_priority,
    w.sync_state,
    w.reply_time
   FROM ((pg_stat_get_activity(NULL::integer) s(datid, pid, usesysid, application_name, state, query, wait_event_type, wait_event, xact_start, query_start, backend_start, state_change, client_addr, client_hostname, client_port, backend_xid, backend_xmin, backend_type, ssl, sslversion, sslcipher, sslbits, ssl_client_dn, ssl_client_serial, ssl_issuer_dn, gss_auth, gss_princ, gss_enc, leader_pid, query_id)
     JOIN pg_stat_get_wal_senders() w(pid, state, sent_lsn, write_lsn, flush_lsn, replay_lsn, write_lag, flush_lag, replay_lag, sync_priority, sync_state, reply_time) ON ((s.pid = w.pid)))
     LEFT JOIN pg_authid u ON ((s.usesysid = u.oid)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_slru"), T(r#" SELECT s.name,
    s.blks_zeroed,
    s.blks_hit,
    s.blks_read,
    s.blks_written,
    s.blks_exists,
    s.flushes,
    s.truncates,
    s.stats_reset
   FROM pg_stat_get_slru() s(name, blks_zeroed, blks_hit, blks_read, blks_written, blks_exists, flushes, truncates, stats_reset);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_wal_receiver"), T(r#" SELECT s.pid,
    s.status,
    s.receive_start_lsn,
    s.receive_start_tli,
    s.written_lsn,
    s.flushed_lsn,
    s.received_tli,
    s.last_msg_send_time,
    s.last_msg_receipt_time,
    s.latest_end_lsn,
    s.latest_end_time,
    s.slot_name,
    s.sender_host,
    s.sender_port,
    s.conninfo
   FROM pg_stat_get_wal_receiver() s(pid, status, receive_start_lsn, receive_start_tli, written_lsn, flushed_lsn, received_tli, last_msg_send_time, last_msg_receipt_time, latest_end_lsn, latest_end_time, slot_name, sender_host, sender_port, conninfo)
  WHERE (s.pid IS NOT NULL);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_recovery_prefetch"), T(r#" SELECT s.stats_reset,
    s.prefetch,
    s.hit,
    s.skip_init,
    s.skip_new,
    s.skip_fpw,
    s.skip_rep,
    s.wal_distance,
    s.block_distance,
    s.io_depth
   FROM pg_stat_get_recovery_prefetch() s(stats_reset, prefetch, hit, skip_init, skip_new, skip_fpw, skip_rep, wal_distance, block_distance, io_depth);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_group"), T(r#" SELECT pg_authid.rolname AS groname,
    pg_authid.oid AS grosysid,
    ARRAY( SELECT pg_auth_members.member
           FROM pg_auth_members
          WHERE (pg_auth_members.roleid = pg_authid.oid)) AS grolist
   FROM pg_authid
  WHERE (NOT pg_authid.rolcanlogin);"#), T("NONE"), T("YES"), T("YES"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_user"), T(r#" SELECT pg_shadow.usename,
    pg_shadow.usesysid,
    pg_shadow.usecreatedb,
    pg_shadow.usesuper,
    pg_shadow.userepl,
    pg_shadow.usebypassrls,
    '********'::text AS passwd,
    pg_shadow.valuntil,
    pg_shadow.useconfig
   FROM pg_shadow;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_policies"), T(r#" SELECT n.nspname AS schemaname,
    c.relname AS tablename,
    pol.polname AS policyname,
        CASE
            WHEN pol.polpermissive THEN 'PERMISSIVE'::text
            ELSE 'RESTRICTIVE'::text
        END AS permissive,
        CASE
            WHEN (pol.polroles = '{0}'::oid[]) THEN (string_to_array('public'::text, ''::text))::name[]
            ELSE ARRAY( SELECT pg_authid.rolname
               FROM pg_authid
              WHERE (pg_authid.oid = ANY (pol.polroles))
              ORDER BY pg_authid.rolname)
        END AS roles,
        CASE pol.polcmd
            WHEN 'r'::"char" THEN 'SELECT'::text
            WHEN 'a'::"char" THEN 'INSERT'::text
            WHEN 'w'::"char" THEN 'UPDATE'::text
            WHEN 'd'::"char" THEN 'DELETE'::text
            WHEN '*'::"char" THEN 'ALL'::text
            ELSE NULL::text
        END AS cmd,
    pg_get_expr(pol.polqual, pol.polrelid) AS qual,
    pg_get_expr(pol.polwithcheck, pol.polrelid) AS with_check
   FROM ((pg_policy pol
     JOIN pg_class c ON ((c.oid = pol.polrelid)))
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_rules"), T(r#" SELECT n.nspname AS schemaname,
    c.relname AS tablename,
    r.rulename,
    pg_get_ruledef(r.oid) AS definition
   FROM ((pg_rewrite r
     JOIN pg_class c ON ((c.oid = r.ev_class)))
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
  WHERE (r.rulename <> '_RETURN'::name);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_views"), T(r#" SELECT n.nspname AS schemaname,
    c.relname AS viewname,
    pg_get_userbyid(c.relowner) AS viewowner,
    pg_get_viewdef(c.oid) AS definition
   FROM (pg_class c
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
  WHERE (c.relkind = 'v'::"char");"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_tables"), T(r#" SELECT n.nspname AS schemaname,
    c.relname AS tablename,
    pg_get_userbyid(c.relowner) AS tableowner,
    t.spcname AS tablespace,
    c.relhasindex AS hasindexes,
    c.relhasrules AS hasrules,
    c.relhastriggers AS hastriggers,
    c.relrowsecurity AS rowsecurity
   FROM ((pg_class c
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
     LEFT JOIN pg_tablespace t ON ((t.oid = c.reltablespace)))
  WHERE (c.relkind = ANY (ARRAY['r'::"char", 'p'::"char"]));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_matviews"), T(r#" SELECT n.nspname AS schemaname,
    c.relname AS matviewname,
    pg_get_userbyid(c.relowner) AS matviewowner,
    t.spcname AS tablespace,
    c.relhasindex AS hasindexes,
    c.relispopulated AS ispopulated,
    pg_get_viewdef(c.oid) AS definition
   FROM ((pg_class c
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
     LEFT JOIN pg_tablespace t ON ((t.oid = c.reltablespace)))
  WHERE (c.relkind = 'm'::"char");"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_indexes"), T(r#" SELECT n.nspname AS schemaname,
    c.relname AS tablename,
    i.relname AS indexname,
    t.spcname AS tablespace,
    pg_get_indexdef(i.oid) AS indexdef
   FROM ((((pg_index x
     JOIN pg_class c ON ((c.oid = x.indrelid)))
     JOIN pg_class i ON ((i.oid = x.indexrelid)))
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
     LEFT JOIN pg_tablespace t ON ((t.oid = i.reltablespace)))
  WHERE ((c.relkind = ANY (ARRAY['r'::"char", 'm'::"char", 'p'::"char"])) AND (i.relkind = ANY (ARRAY['i'::"char", 'I'::"char"])));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_sequences"), T(r#" SELECT n.nspname AS schemaname,
    c.relname AS sequencename,
    pg_get_userbyid(c.relowner) AS sequenceowner,
    (s.seqtypid)::regtype AS data_type,
    s.seqstart AS start_value,
    s.seqmin AS min_value,
    s.seqmax AS max_value,
    s.seqincrement AS increment_by,
    s.seqcycle AS cycle,
    s.seqcache AS cache_size,
        CASE
            WHEN has_sequence_privilege(c.oid, 'SELECT,USAGE'::text) THEN pg_sequence_last_value((c.oid)::regclass)
            ELSE NULL::bigint
        END AS last_value
   FROM ((pg_sequence s
     JOIN pg_class c ON ((c.oid = s.seqrelid)))
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
  WHERE ((NOT pg_is_other_temp_schema(n.oid)) AND (c.relkind = 'S'::"char"));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stats"), T(r#" SELECT n.nspname AS schemaname,
    c.relname AS tablename,
    a.attname,
    s.stainherit AS inherited,
    s.stanullfrac AS null_frac,
    s.stawidth AS avg_width,
    s.stadistinct AS n_distinct,
        CASE
            WHEN (s.stakind1 = 1) THEN s.stavalues1
            WHEN (s.stakind2 = 1) THEN s.stavalues2
            WHEN (s.stakind3 = 1) THEN s.stavalues3
            WHEN (s.stakind4 = 1) THEN s.stavalues4
            WHEN (s.stakind5 = 1) THEN s.stavalues5
            ELSE NULL::anyarray
        END AS most_common_vals,
        CASE
            WHEN (s.stakind1 = 1) THEN s.stanumbers1
            WHEN (s.stakind2 = 1) THEN s.stanumbers2
            WHEN (s.stakind3 = 1) THEN s.stanumbers3
            WHEN (s.stakind4 = 1) THEN s.stanumbers4
            WHEN (s.stakind5 = 1) THEN s.stanumbers5
            ELSE NULL::real[]
        END AS most_common_freqs,
        CASE
            WHEN (s.stakind1 = 2) THEN s.stavalues1
            WHEN (s.stakind2 = 2) THEN s.stavalues2
            WHEN (s.stakind3 = 2) THEN s.stavalues3
            WHEN (s.stakind4 = 2) THEN s.stavalues4
            WHEN (s.stakind5 = 2) THEN s.stavalues5
            ELSE NULL::anyarray
        END AS histogram_bounds,
        CASE
            WHEN (s.stakind1 = 3) THEN s.stanumbers1[1]
            WHEN (s.stakind2 = 3) THEN s.stanumbers2[1]
            WHEN (s.stakind3 = 3) THEN s.stanumbers3[1]
            WHEN (s.stakind4 = 3) THEN s.stanumbers4[1]
            WHEN (s.stakind5 = 3) THEN s.stanumbers5[1]
            ELSE NULL::real
        END AS correlation,
        CASE
            WHEN (s.stakind1 = 4) THEN s.stavalues1
            WHEN (s.stakind2 = 4) THEN s.stavalues2
            WHEN (s.stakind3 = 4) THEN s.stavalues3
            WHEN (s.stakind4 = 4) THEN s.stavalues4
            WHEN (s.stakind5 = 4) THEN s.stavalues5
            ELSE NULL::anyarray
        END AS most_common_elems,
        CASE
            WHEN (s.stakind1 = 4) THEN s.stanumbers1
            WHEN (s.stakind2 = 4) THEN s.stanumbers2
            WHEN (s.stakind3 = 4) THEN s.stanumbers3
            WHEN (s.stakind4 = 4) THEN s.stanumbers4
            WHEN (s.stakind5 = 4) THEN s.stanumbers5
            ELSE NULL::real[]
        END AS most_common_elem_freqs,
        CASE
            WHEN (s.stakind1 = 5) THEN s.stanumbers1
            WHEN (s.stakind2 = 5) THEN s.stanumbers2
            WHEN (s.stakind3 = 5) THEN s.stanumbers3
            WHEN (s.stakind4 = 5) THEN s.stanumbers4
            WHEN (s.stakind5 = 5) THEN s.stanumbers5
            ELSE NULL::real[]
        END AS elem_count_histogram
   FROM (((pg_statistic s
     JOIN pg_class c ON ((c.oid = s.starelid)))
     JOIN pg_attribute a ON (((c.oid = a.attrelid) AND (a.attnum = s.staattnum))))
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
  WHERE ((NOT a.attisdropped) AND has_column_privilege(c.oid, a.attnum, 'select'::text) AND ((c.relrowsecurity = false) OR (NOT row_security_active(c.oid))));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stats_ext"), T(r#" SELECT cn.nspname AS schemaname,
    c.relname AS tablename,
    sn.nspname AS statistics_schemaname,
    s.stxname AS statistics_name,
    pg_get_userbyid(s.stxowner) AS statistics_owner,
    ( SELECT array_agg(a.attname ORDER BY a.attnum) AS array_agg
           FROM (unnest(s.stxkeys) k(k)
             JOIN pg_attribute a ON (((a.attrelid = s.stxrelid) AND (a.attnum = k.k))))) AS attnames,
    pg_get_statisticsobjdef_expressions(s.oid) AS exprs,
    s.stxkind AS kinds,
    sd.stxdinherit AS inherited,
    sd.stxdndistinct AS n_distinct,
    sd.stxddependencies AS dependencies,
    m.most_common_vals,
    m.most_common_val_nulls,
    m.most_common_freqs,
    m.most_common_base_freqs
   FROM (((((pg_statistic_ext s
     JOIN pg_class c ON ((c.oid = s.stxrelid)))
     JOIN pg_statistic_ext_data sd ON ((s.oid = sd.stxoid)))
     LEFT JOIN pg_namespace cn ON ((cn.oid = c.relnamespace)))
     LEFT JOIN pg_namespace sn ON ((sn.oid = s.stxnamespace)))
     LEFT JOIN LATERAL ( SELECT array_agg(pg_mcv_list_items."values") AS most_common_vals,
            array_agg(pg_mcv_list_items.nulls) AS most_common_val_nulls,
            array_agg(pg_mcv_list_items.frequency) AS most_common_freqs,
            array_agg(pg_mcv_list_items.base_frequency) AS most_common_base_freqs
           FROM pg_mcv_list_items(sd.stxdmcv) pg_mcv_list_items(index, "values", nulls, frequency, base_frequency)) m ON ((sd.stxdmcv IS NOT NULL)))
  WHERE (pg_has_role(c.relowner, 'USAGE'::text) AND ((c.relrowsecurity = false) OR (NOT row_security_active(c.oid))));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stats_ext_exprs"), T(r#" SELECT cn.nspname AS schemaname,
    c.relname AS tablename,
    sn.nspname AS statistics_schemaname,
    s.stxname AS statistics_name,
    pg_get_userbyid(s.stxowner) AS statistics_owner,
    stat.expr,
    sd.stxdinherit AS inherited,
    (stat.a).stanullfrac AS null_frac,
    (stat.a).stawidth AS avg_width,
    (stat.a).stadistinct AS n_distinct,
        CASE
            WHEN ((stat.a).stakind1 = 1) THEN (stat.a).stavalues1
            WHEN ((stat.a).stakind2 = 1) THEN (stat.a).stavalues2
            WHEN ((stat.a).stakind3 = 1) THEN (stat.a).stavalues3
            WHEN ((stat.a).stakind4 = 1) THEN (stat.a).stavalues4
            WHEN ((stat.a).stakind5 = 1) THEN (stat.a).stavalues5
            ELSE NULL::anyarray
        END AS most_common_vals,
        CASE
            WHEN ((stat.a).stakind1 = 1) THEN (stat.a).stanumbers1
            WHEN ((stat.a).stakind2 = 1) THEN (stat.a).stanumbers2
            WHEN ((stat.a).stakind3 = 1) THEN (stat.a).stanumbers3
            WHEN ((stat.a).stakind4 = 1) THEN (stat.a).stanumbers4
            WHEN ((stat.a).stakind5 = 1) THEN (stat.a).stanumbers5
            ELSE NULL::real[]
        END AS most_common_freqs,
        CASE
            WHEN ((stat.a).stakind1 = 2) THEN (stat.a).stavalues1
            WHEN ((stat.a).stakind2 = 2) THEN (stat.a).stavalues2
            WHEN ((stat.a).stakind3 = 2) THEN (stat.a).stavalues3
            WHEN ((stat.a).stakind4 = 2) THEN (stat.a).stavalues4
            WHEN ((stat.a).stakind5 = 2) THEN (stat.a).stavalues5
            ELSE NULL::anyarray
        END AS histogram_bounds,
        CASE
            WHEN ((stat.a).stakind1 = 3) THEN (stat.a).stanumbers1[1]
            WHEN ((stat.a).stakind2 = 3) THEN (stat.a).stanumbers2[1]
            WHEN ((stat.a).stakind3 = 3) THEN (stat.a).stanumbers3[1]
            WHEN ((stat.a).stakind4 = 3) THEN (stat.a).stanumbers4[1]
            WHEN ((stat.a).stakind5 = 3) THEN (stat.a).stanumbers5[1]
            ELSE NULL::real
        END AS correlation,
        CASE
            WHEN ((stat.a).stakind1 = 4) THEN (stat.a).stavalues1
            WHEN ((stat.a).stakind2 = 4) THEN (stat.a).stavalues2
            WHEN ((stat.a).stakind3 = 4) THEN (stat.a).stavalues3
            WHEN ((stat.a).stakind4 = 4) THEN (stat.a).stavalues4
            WHEN ((stat.a).stakind5 = 4) THEN (stat.a).stavalues5
            ELSE NULL::anyarray
        END AS most_common_elems,
        CASE
            WHEN ((stat.a).stakind1 = 4) THEN (stat.a).stanumbers1
            WHEN ((stat.a).stakind2 = 4) THEN (stat.a).stanumbers2
            WHEN ((stat.a).stakind3 = 4) THEN (stat.a).stanumbers3
            WHEN ((stat.a).stakind4 = 4) THEN (stat.a).stanumbers4
            WHEN ((stat.a).stakind5 = 4) THEN (stat.a).stanumbers5
            ELSE NULL::real[]
        END AS most_common_elem_freqs,
        CASE
            WHEN ((stat.a).stakind1 = 5) THEN (stat.a).stanumbers1
            WHEN ((stat.a).stakind2 = 5) THEN (stat.a).stanumbers2
            WHEN ((stat.a).stakind3 = 5) THEN (stat.a).stanumbers3
            WHEN ((stat.a).stakind4 = 5) THEN (stat.a).stanumbers4
            WHEN ((stat.a).stakind5 = 5) THEN (stat.a).stanumbers5
            ELSE NULL::real[]
        END AS elem_count_histogram
   FROM (((((pg_statistic_ext s
     JOIN pg_class c ON ((c.oid = s.stxrelid)))
     LEFT JOIN pg_statistic_ext_data sd ON ((s.oid = sd.stxoid)))
     LEFT JOIN pg_namespace cn ON ((cn.oid = c.relnamespace)))
     LEFT JOIN pg_namespace sn ON ((sn.oid = s.stxnamespace)))
     JOIN LATERAL ( SELECT unnest(pg_get_statisticsobjdef_expressions(s.oid)) AS expr,
            unnest(sd.stxdexpr) AS a) stat ON ((stat.expr IS NOT NULL)))
  WHERE (pg_has_role(c.relowner, 'USAGE'::text) AND ((c.relrowsecurity = false) OR (NOT row_security_active(c.oid))));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_publication_tables"), T(r#" SELECT p.pubname,
    n.nspname AS schemaname,
    c.relname AS tablename,
    ( SELECT array_agg(a.attname ORDER BY a.attnum) AS array_agg
           FROM pg_attribute a
          WHERE ((a.attrelid = gpt.relid) AND (a.attnum > 0) AND (NOT a.attisdropped) AND ((a.attnum = ANY ((gpt.attrs)::smallint[])) OR (gpt.attrs IS NULL)))) AS attnames,
    pg_get_expr(gpt.qual, gpt.relid) AS rowfilter
   FROM pg_publication p,
    LATERAL pg_get_publication_tables((p.pubname)::text) gpt(relid, attrs, qual),
    (pg_class c
     JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
  WHERE (c.oid = gpt.relid);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_locks"), T(r#" SELECT l.locktype,
    l.database,
    l.relation,
    l.page,
    l.tuple,
    l.virtualxid,
    l.transactionid,
    l.classid,
    l.objid,
    l.objsubid,
    l.virtualtransaction,
    l.pid,
    l.mode,
    l.granted,
    l.fastpath,
    l.waitstart
   FROM pg_lock_status() l(locktype, database, relation, page, tuple, virtualxid, transactionid, classid, objid, objsubid, virtualtransaction, pid, mode, granted, fastpath, waitstart);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_cursors"), T(r#" SELECT c.name,
    c.statement,
    c.is_holdable,
    c.is_binary,
    c.is_scrollable,
    c.creation_time
   FROM pg_cursor() c(name, statement, is_holdable, is_binary, is_scrollable, creation_time);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_available_extensions"), T(r#" SELECT e.name,
    e.default_version,
    x.extversion AS installed_version,
    e.comment
   FROM (pg_available_extensions() e(name, default_version, comment)
     LEFT JOIN pg_extension x ON ((e.name = x.extname)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_available_extension_versions"), T(r#" SELECT e.name,
    e.version,
    (x.extname IS NOT NULL) AS installed,
    e.superuser,
    e.trusted,
    e.relocatable,
    e.schema,
    e.requires,
    e.comment
   FROM (pg_available_extension_versions() e(name, version, superuser, trusted, relocatable, schema, requires, comment)
     LEFT JOIN pg_extension x ON (((e.name = x.extname) AND (e.version = x.extversion))));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_prepared_xacts"), T(r#" SELECT p.transaction,
    p.gid,
    p.prepared,
    u.rolname AS owner,
    d.datname AS database
   FROM ((pg_prepared_xact() p(transaction, gid, prepared, ownerid, dbid)
     LEFT JOIN pg_authid u ON ((p.ownerid = u.oid)))
     LEFT JOIN pg_database d ON ((p.dbid = d.oid)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_prepared_statements"), T(r#" SELECT p.name,
    p.statement,
    p.prepare_time,
    p.parameter_types,
    p.from_sql,
    p.generic_plans,
    p.custom_plans
   FROM pg_prepared_statement() p(name, statement, prepare_time, parameter_types, from_sql, generic_plans, custom_plans);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_seclabels"), T(r#" SELECT l.objoid,
    l.classoid,
    l.objsubid,
        CASE
            WHEN (rel.relkind = ANY (ARRAY['r'::"char", 'p'::"char"])) THEN 'table'::text
            WHEN (rel.relkind = 'v'::"char") THEN 'view'::text
            WHEN (rel.relkind = 'm'::"char") THEN 'materialized view'::text
            WHEN (rel.relkind = 'S'::"char") THEN 'sequence'::text
            WHEN (rel.relkind = 'f'::"char") THEN 'foreign table'::text
            ELSE NULL::text
        END AS objtype,
    rel.relnamespace AS objnamespace,
        CASE
            WHEN pg_table_is_visible(rel.oid) THEN quote_ident((rel.relname)::text)
            ELSE ((quote_ident((nsp.nspname)::text) || '.'::text) || quote_ident((rel.relname)::text))
        END AS objname,
    l.provider,
    l.label
   FROM ((pg_seclabel l
     JOIN pg_class rel ON (((l.classoid = rel.tableoid) AND (l.objoid = rel.oid))))
     JOIN pg_namespace nsp ON ((rel.relnamespace = nsp.oid)))
  WHERE (l.objsubid = 0)
UNION ALL
 SELECT l.objoid,
    l.classoid,
    l.objsubid,
    'column'::text AS objtype,
    rel.relnamespace AS objnamespace,
    ((
        CASE
            WHEN pg_table_is_visible(rel.oid) THEN quote_ident((rel.relname)::text)
            ELSE ((quote_ident((nsp.nspname)::text) || '.'::text) || quote_ident((rel.relname)::text))
        END || '.'::text) || (att.attname)::text) AS objname,
    l.provider,
    l.label
   FROM (((pg_seclabel l
     JOIN pg_class rel ON (((l.classoid = rel.tableoid) AND (l.objoid = rel.oid))))
     JOIN pg_attribute att ON (((rel.oid = att.attrelid) AND (l.objsubid = att.attnum))))
     JOIN pg_namespace nsp ON ((rel.relnamespace = nsp.oid)))
  WHERE (l.objsubid <> 0)
UNION ALL
 SELECT l.objoid,
    l.classoid,
    l.objsubid,
        CASE pro.prokind
            WHEN 'a'::"char" THEN 'aggregate'::text
            WHEN 'f'::"char" THEN 'function'::text
            WHEN 'p'::"char" THEN 'procedure'::text
            WHEN 'w'::"char" THEN 'window'::text
            ELSE NULL::text
        END AS objtype,
    pro.pronamespace AS objnamespace,
    (((
        CASE
            WHEN pg_function_is_visible(pro.oid) THEN quote_ident((pro.proname)::text)
            ELSE ((quote_ident((nsp.nspname)::text) || '.'::text) || quote_ident((pro.proname)::text))
        END || '('::text) || pg_get_function_arguments(pro.oid)) || ')'::text) AS objname,
    l.provider,
    l.label
   FROM ((pg_seclabel l
     JOIN pg_proc pro ON (((l.classoid = pro.tableoid) AND (l.objoid = pro.oid))))
     JOIN pg_namespace nsp ON ((pro.pronamespace = nsp.oid)))
  WHERE (l.objsubid = 0)
UNION ALL
 SELECT l.objoid,
    l.classoid,
    l.objsubid,
        CASE
            WHEN (typ.typtype = 'd'::"char") THEN 'domain'::text
            ELSE 'type'::text
        END AS objtype,
    typ.typnamespace AS objnamespace,
        CASE
            WHEN pg_type_is_visible(typ.oid) THEN quote_ident((typ.typname)::text)
            ELSE ((quote_ident((nsp.nspname)::text) || '.'::text) || quote_ident((typ.typname)::text))
        END AS objname,
    l.provider,
    l.label
   FROM ((pg_seclabel l
     JOIN pg_type typ ON (((l.classoid = typ.tableoid) AND (l.objoid = typ.oid))))
     JOIN pg_namespace nsp ON ((typ.typnamespace = nsp.oid)))
  WHERE (l.objsubid = 0)
UNION ALL
 SELECT l.objoid,
    l.classoid,
    l.objsubid,
    'large object'::text AS objtype,
    NULL::oid AS objnamespace,
    (l.objoid)::text AS objname,
    l.provider,
    l.label
   FROM (pg_seclabel l
     JOIN pg_largeobject_metadata lom ON ((l.objoid = lom.oid)))
  WHERE ((l.classoid = ('pg_largeobject'::regclass)::oid) AND (l.objsubid = 0))
UNION ALL
 SELECT l.objoid,
    l.classoid,
    l.objsubid,
    'language'::text AS objtype,
    NULL::oid AS objnamespace,
    quote_ident((lan.lanname)::text) AS objname,
    l.provider,
    l.label
   FROM (pg_seclabel l
     JOIN pg_language lan ON (((l.classoid = lan.tableoid) AND (l.objoid = lan.oid))))
  WHERE (l.objsubid = 0)
UNION ALL
 SELECT l.objoid,
    l.classoid,
    l.objsubid,
    'schema'::text AS objtype,
    nsp.oid AS objnamespace,
    quote_ident((nsp.nspname)::text) AS objname,
    l.provider,
    l.label
   FROM (pg_seclabel l
     JOIN pg_namespace nsp ON (((l.classoid = nsp.tableoid) AND (l.objoid = nsp.oid))))
  WHERE (l.objsubid = 0)
UNION ALL
 SELECT l.objoid,
    l.classoid,
    l.objsubid,
    'event trigger'::text AS objtype,
    NULL::oid AS objnamespace,
    quote_ident((evt.evtname)::text) AS objname,
    l.provider,
    l.label
   FROM (pg_seclabel l
     JOIN pg_event_trigger evt ON (((l.classoid = evt.tableoid) AND (l.objoid = evt.oid))))
  WHERE (l.objsubid = 0)
UNION ALL
 SELECT l.objoid,
    l.classoid,
    l.objsubid,
    'publication'::text AS objtype,
    NULL::oid AS objnamespace,
    quote_ident((p.pubname)::text) AS objname,
    l.provider,
    l.label
   FROM (pg_seclabel l
     JOIN pg_publication p ON (((l.classoid = p.tableoid) AND (l.objoid = p.oid))))
  WHERE (l.objsubid = 0)
UNION ALL
 SELECT l.objoid,
    l.classoid,
    0 AS objsubid,
    'subscription'::text AS objtype,
    NULL::oid AS objnamespace,
    quote_ident((s.subname)::text) AS objname,
    l.provider,
    l.label
   FROM (pg_shseclabel l
     JOIN pg_subscription s ON (((l.classoid = s.tableoid) AND (l.objoid = s.oid))))
UNION ALL
 SELECT l.objoid,
    l.classoid,
    0 AS objsubid,
    'database'::text AS objtype,
    NULL::oid AS objnamespace,
    quote_ident((dat.datname)::text) AS objname,
    l.provider,
    l.label
   FROM (pg_shseclabel l
     JOIN pg_database dat ON (((l.classoid = dat.tableoid) AND (l.objoid = dat.oid))))
UNION ALL
 SELECT l.objoid,
    l.classoid,
    0 AS objsubid,
    'tablespace'::text AS objtype,
    NULL::oid AS objnamespace,
    quote_ident((spc.spcname)::text) AS objname,
    l.provider,
    l.label
   FROM (pg_shseclabel l
     JOIN pg_tablespace spc ON (((l.classoid = spc.tableoid) AND (l.objoid = spc.oid))))
UNION ALL
 SELECT l.objoid,
    l.classoid,
    0 AS objsubid,
    'role'::text AS objtype,
    NULL::oid AS objnamespace,
    quote_ident((rol.rolname)::text) AS objname,
    l.provider,
    l.label
   FROM (pg_shseclabel l
     JOIN pg_authid rol ON (((l.classoid = rol.tableoid) AND (l.objoid = rol.oid))));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_timezone_abbrevs"), T(r#" SELECT pg_timezone_abbrevs.abbrev,
    pg_timezone_abbrevs.utc_offset,
    pg_timezone_abbrevs.is_dst
   FROM pg_timezone_abbrevs() pg_timezone_abbrevs(abbrev, utc_offset, is_dst);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_timezone_names"), T(r#" SELECT pg_timezone_names.name,
    pg_timezone_names.abbrev,
    pg_timezone_names.utc_offset,
    pg_timezone_names.is_dst
   FROM pg_timezone_names() pg_timezone_names(name, abbrev, utc_offset, is_dst);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_sys_tables"), T(r#" SELECT pg_stat_all_tables.relid,
    pg_stat_all_tables.schemaname,
    pg_stat_all_tables.relname,
    pg_stat_all_tables.seq_scan,
    pg_stat_all_tables.seq_tup_read,
    pg_stat_all_tables.idx_scan,
    pg_stat_all_tables.idx_tup_fetch,
    pg_stat_all_tables.n_tup_ins,
    pg_stat_all_tables.n_tup_upd,
    pg_stat_all_tables.n_tup_del,
    pg_stat_all_tables.n_tup_hot_upd,
    pg_stat_all_tables.n_live_tup,
    pg_stat_all_tables.n_dead_tup,
    pg_stat_all_tables.n_mod_since_analyze,
    pg_stat_all_tables.n_ins_since_vacuum,
    pg_stat_all_tables.last_vacuum,
    pg_stat_all_tables.last_autovacuum,
    pg_stat_all_tables.last_analyze,
    pg_stat_all_tables.last_autoanalyze,
    pg_stat_all_tables.vacuum_count,
    pg_stat_all_tables.autovacuum_count,
    pg_stat_all_tables.analyze_count,
    pg_stat_all_tables.autoanalyze_count
   FROM pg_stat_all_tables
  WHERE ((pg_stat_all_tables.schemaname = ANY (ARRAY['pg_catalog'::name, 'information_schema'::name])) OR (pg_stat_all_tables.schemaname ~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_xact_sys_tables"), T(r#" SELECT pg_stat_xact_all_tables.relid,
    pg_stat_xact_all_tables.schemaname,
    pg_stat_xact_all_tables.relname,
    pg_stat_xact_all_tables.seq_scan,
    pg_stat_xact_all_tables.seq_tup_read,
    pg_stat_xact_all_tables.idx_scan,
    pg_stat_xact_all_tables.idx_tup_fetch,
    pg_stat_xact_all_tables.n_tup_ins,
    pg_stat_xact_all_tables.n_tup_upd,
    pg_stat_xact_all_tables.n_tup_del,
    pg_stat_xact_all_tables.n_tup_hot_upd
   FROM pg_stat_xact_all_tables
  WHERE ((pg_stat_xact_all_tables.schemaname = ANY (ARRAY['pg_catalog'::name, 'information_schema'::name])) OR (pg_stat_xact_all_tables.schemaname ~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_user_tables"), T(r#" SELECT pg_stat_all_tables.relid,
    pg_stat_all_tables.schemaname,
    pg_stat_all_tables.relname,
    pg_stat_all_tables.seq_scan,
    pg_stat_all_tables.seq_tup_read,
    pg_stat_all_tables.idx_scan,
    pg_stat_all_tables.idx_tup_fetch,
    pg_stat_all_tables.n_tup_ins,
    pg_stat_all_tables.n_tup_upd,
    pg_stat_all_tables.n_tup_del,
    pg_stat_all_tables.n_tup_hot_upd,
    pg_stat_all_tables.n_live_tup,
    pg_stat_all_tables.n_dead_tup,
    pg_stat_all_tables.n_mod_since_analyze,
    pg_stat_all_tables.n_ins_since_vacuum,
    pg_stat_all_tables.last_vacuum,
    pg_stat_all_tables.last_autovacuum,
    pg_stat_all_tables.last_analyze,
    pg_stat_all_tables.last_autoanalyze,
    pg_stat_all_tables.vacuum_count,
    pg_stat_all_tables.autovacuum_count,
    pg_stat_all_tables.analyze_count,
    pg_stat_all_tables.autoanalyze_count
   FROM pg_stat_all_tables
  WHERE ((pg_stat_all_tables.schemaname <> ALL (ARRAY['pg_catalog'::name, 'information_schema'::name])) AND (pg_stat_all_tables.schemaname !~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_all_tables"), T(r#" SELECT c.oid AS relid,
    n.nspname AS schemaname,
    c.relname,
    pg_stat_get_numscans(c.oid) AS seq_scan,
    pg_stat_get_tuples_returned(c.oid) AS seq_tup_read,
    (sum(pg_stat_get_numscans(i.indexrelid)))::bigint AS idx_scan,
    ((sum(pg_stat_get_tuples_fetched(i.indexrelid)))::bigint + pg_stat_get_tuples_fetched(c.oid)) AS idx_tup_fetch,
    pg_stat_get_tuples_inserted(c.oid) AS n_tup_ins,
    pg_stat_get_tuples_updated(c.oid) AS n_tup_upd,
    pg_stat_get_tuples_deleted(c.oid) AS n_tup_del,
    pg_stat_get_tuples_hot_updated(c.oid) AS n_tup_hot_upd,
    pg_stat_get_live_tuples(c.oid) AS n_live_tup,
    pg_stat_get_dead_tuples(c.oid) AS n_dead_tup,
    pg_stat_get_mod_since_analyze(c.oid) AS n_mod_since_analyze,
    pg_stat_get_ins_since_vacuum(c.oid) AS n_ins_since_vacuum,
    pg_stat_get_last_vacuum_time(c.oid) AS last_vacuum,
    pg_stat_get_last_autovacuum_time(c.oid) AS last_autovacuum,
    pg_stat_get_last_analyze_time(c.oid) AS last_analyze,
    pg_stat_get_last_autoanalyze_time(c.oid) AS last_autoanalyze,
    pg_stat_get_vacuum_count(c.oid) AS vacuum_count,
    pg_stat_get_autovacuum_count(c.oid) AS autovacuum_count,
    pg_stat_get_analyze_count(c.oid) AS analyze_count,
    pg_stat_get_autoanalyze_count(c.oid) AS autoanalyze_count
   FROM ((pg_class c
     LEFT JOIN pg_index i ON ((c.oid = i.indrelid)))
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
  WHERE (c.relkind = ANY (ARRAY['r'::"char", 't'::"char", 'm'::"char", 'p'::"char"]))
  GROUP BY c.oid, n.nspname, c.relname;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_xact_all_tables"), T(r#" SELECT c.oid AS relid,
    n.nspname AS schemaname,
    c.relname,
    pg_stat_get_xact_numscans(c.oid) AS seq_scan,
    pg_stat_get_xact_tuples_returned(c.oid) AS seq_tup_read,
    (sum(pg_stat_get_xact_numscans(i.indexrelid)))::bigint AS idx_scan,
    ((sum(pg_stat_get_xact_tuples_fetched(i.indexrelid)))::bigint + pg_stat_get_xact_tuples_fetched(c.oid)) AS idx_tup_fetch,
    pg_stat_get_xact_tuples_inserted(c.oid) AS n_tup_ins,
    pg_stat_get_xact_tuples_updated(c.oid) AS n_tup_upd,
    pg_stat_get_xact_tuples_deleted(c.oid) AS n_tup_del,
    pg_stat_get_xact_tuples_hot_updated(c.oid) AS n_tup_hot_upd
   FROM ((pg_class c
     LEFT JOIN pg_index i ON ((c.oid = i.indrelid)))
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
  WHERE (c.relkind = ANY (ARRAY['r'::"char", 't'::"char", 'm'::"char", 'p'::"char"]))
  GROUP BY c.oid, n.nspname, c.relname;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_xact_user_tables"), T(r#" SELECT pg_stat_xact_all_tables.relid,
    pg_stat_xact_all_tables.schemaname,
    pg_stat_xact_all_tables.relname,
    pg_stat_xact_all_tables.seq_scan,
    pg_stat_xact_all_tables.seq_tup_read,
    pg_stat_xact_all_tables.idx_scan,
    pg_stat_xact_all_tables.idx_tup_fetch,
    pg_stat_xact_all_tables.n_tup_ins,
    pg_stat_xact_all_tables.n_tup_upd,
    pg_stat_xact_all_tables.n_tup_del,
    pg_stat_xact_all_tables.n_tup_hot_upd
   FROM pg_stat_xact_all_tables
  WHERE ((pg_stat_xact_all_tables.schemaname <> ALL (ARRAY['pg_catalog'::name, 'information_schema'::name])) AND (pg_stat_xact_all_tables.schemaname !~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_statio_all_tables"), T(r#" SELECT c.oid AS relid,
    n.nspname AS schemaname,
    c.relname,
    (pg_stat_get_blocks_fetched(c.oid) - pg_stat_get_blocks_hit(c.oid)) AS heap_blks_read,
    pg_stat_get_blocks_hit(c.oid) AS heap_blks_hit,
    i.idx_blks_read,
    i.idx_blks_hit,
    (pg_stat_get_blocks_fetched(t.oid) - pg_stat_get_blocks_hit(t.oid)) AS toast_blks_read,
    pg_stat_get_blocks_hit(t.oid) AS toast_blks_hit,
    x.idx_blks_read AS tidx_blks_read,
    x.idx_blks_hit AS tidx_blks_hit
   FROM ((((pg_class c
     LEFT JOIN pg_class t ON ((c.reltoastrelid = t.oid)))
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
     LEFT JOIN LATERAL ( SELECT (sum((pg_stat_get_blocks_fetched(pg_index.indexrelid) - pg_stat_get_blocks_hit(pg_index.indexrelid))))::bigint AS idx_blks_read,
            (sum(pg_stat_get_blocks_hit(pg_index.indexrelid)))::bigint AS idx_blks_hit
           FROM pg_index
          WHERE (pg_index.indrelid = c.oid)) i ON (true))
     LEFT JOIN LATERAL ( SELECT (sum((pg_stat_get_blocks_fetched(pg_index.indexrelid) - pg_stat_get_blocks_hit(pg_index.indexrelid))))::bigint AS idx_blks_read,
            (sum(pg_stat_get_blocks_hit(pg_index.indexrelid)))::bigint AS idx_blks_hit
           FROM pg_index
          WHERE (pg_index.indrelid = t.oid)) x ON (true))
  WHERE (c.relkind = ANY (ARRAY['r'::"char", 't'::"char", 'm'::"char"]));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_statio_sys_tables"), T(r#" SELECT pg_statio_all_tables.relid,
    pg_statio_all_tables.schemaname,
    pg_statio_all_tables.relname,
    pg_statio_all_tables.heap_blks_read,
    pg_statio_all_tables.heap_blks_hit,
    pg_statio_all_tables.idx_blks_read,
    pg_statio_all_tables.idx_blks_hit,
    pg_statio_all_tables.toast_blks_read,
    pg_statio_all_tables.toast_blks_hit,
    pg_statio_all_tables.tidx_blks_read,
    pg_statio_all_tables.tidx_blks_hit
   FROM pg_statio_all_tables
  WHERE ((pg_statio_all_tables.schemaname = ANY (ARRAY['pg_catalog'::name, 'information_schema'::name])) OR (pg_statio_all_tables.schemaname ~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_statio_user_tables"), T(r#" SELECT pg_statio_all_tables.relid,
    pg_statio_all_tables.schemaname,
    pg_statio_all_tables.relname,
    pg_statio_all_tables.heap_blks_read,
    pg_statio_all_tables.heap_blks_hit,
    pg_statio_all_tables.idx_blks_read,
    pg_statio_all_tables.idx_blks_hit,
    pg_statio_all_tables.toast_blks_read,
    pg_statio_all_tables.toast_blks_hit,
    pg_statio_all_tables.tidx_blks_read,
    pg_statio_all_tables.tidx_blks_hit
   FROM pg_statio_all_tables
  WHERE ((pg_statio_all_tables.schemaname <> ALL (ARRAY['pg_catalog'::name, 'information_schema'::name])) AND (pg_statio_all_tables.schemaname !~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_all_indexes"), T(r#" SELECT c.oid AS relid,
    i.oid AS indexrelid,
    n.nspname AS schemaname,
    c.relname,
    i.relname AS indexrelname,
    pg_stat_get_numscans(i.oid) AS idx_scan,
    pg_stat_get_tuples_returned(i.oid) AS idx_tup_read,
    pg_stat_get_tuples_fetched(i.oid) AS idx_tup_fetch
   FROM (((pg_class c
     JOIN pg_index x ON ((c.oid = x.indrelid)))
     JOIN pg_class i ON ((i.oid = x.indexrelid)))
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
  WHERE (c.relkind = ANY (ARRAY['r'::"char", 't'::"char", 'm'::"char"]));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_sys_indexes"), T(r#" SELECT pg_stat_all_indexes.relid,
    pg_stat_all_indexes.indexrelid,
    pg_stat_all_indexes.schemaname,
    pg_stat_all_indexes.relname,
    pg_stat_all_indexes.indexrelname,
    pg_stat_all_indexes.idx_scan,
    pg_stat_all_indexes.idx_tup_read,
    pg_stat_all_indexes.idx_tup_fetch
   FROM pg_stat_all_indexes
  WHERE ((pg_stat_all_indexes.schemaname = ANY (ARRAY['pg_catalog'::name, 'information_schema'::name])) OR (pg_stat_all_indexes.schemaname ~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_user_indexes"), T(r#" SELECT pg_stat_all_indexes.relid,
    pg_stat_all_indexes.indexrelid,
    pg_stat_all_indexes.schemaname,
    pg_stat_all_indexes.relname,
    pg_stat_all_indexes.indexrelname,
    pg_stat_all_indexes.idx_scan,
    pg_stat_all_indexes.idx_tup_read,
    pg_stat_all_indexes.idx_tup_fetch
   FROM pg_stat_all_indexes
  WHERE ((pg_stat_all_indexes.schemaname <> ALL (ARRAY['pg_catalog'::name, 'information_schema'::name])) AND (pg_stat_all_indexes.schemaname !~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_statio_all_indexes"), T(r#" SELECT c.oid AS relid,
    i.oid AS indexrelid,
    n.nspname AS schemaname,
    c.relname,
    i.relname AS indexrelname,
    (pg_stat_get_blocks_fetched(i.oid) - pg_stat_get_blocks_hit(i.oid)) AS idx_blks_read,
    pg_stat_get_blocks_hit(i.oid) AS idx_blks_hit
   FROM (((pg_class c
     JOIN pg_index x ON ((c.oid = x.indrelid)))
     JOIN pg_class i ON ((i.oid = x.indexrelid)))
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
  WHERE (c.relkind = ANY (ARRAY['r'::"char", 't'::"char", 'm'::"char"]));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_statio_sys_indexes"), T(r#" SELECT pg_statio_all_indexes.relid,
    pg_statio_all_indexes.indexrelid,
    pg_statio_all_indexes.schemaname,
    pg_statio_all_indexes.relname,
    pg_statio_all_indexes.indexrelname,
    pg_statio_all_indexes.idx_blks_read,
    pg_statio_all_indexes.idx_blks_hit
   FROM pg_statio_all_indexes
  WHERE ((pg_statio_all_indexes.schemaname = ANY (ARRAY['pg_catalog'::name, 'information_schema'::name])) OR (pg_statio_all_indexes.schemaname ~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_statio_user_indexes"), T(r#" SELECT pg_statio_all_indexes.relid,
    pg_statio_all_indexes.indexrelid,
    pg_statio_all_indexes.schemaname,
    pg_statio_all_indexes.relname,
    pg_statio_all_indexes.indexrelname,
    pg_statio_all_indexes.idx_blks_read,
    pg_statio_all_indexes.idx_blks_hit
   FROM pg_statio_all_indexes
  WHERE ((pg_statio_all_indexes.schemaname <> ALL (ARRAY['pg_catalog'::name, 'information_schema'::name])) AND (pg_statio_all_indexes.schemaname !~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_statio_all_sequences"), T(r#" SELECT c.oid AS relid,
    n.nspname AS schemaname,
    c.relname,
    (pg_stat_get_blocks_fetched(c.oid) - pg_stat_get_blocks_hit(c.oid)) AS blks_read,
    pg_stat_get_blocks_hit(c.oid) AS blks_hit
   FROM (pg_class c
     LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
  WHERE (c.relkind = 'S'::"char");"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_statio_sys_sequences"), T(r#" SELECT pg_statio_all_sequences.relid,
    pg_statio_all_sequences.schemaname,
    pg_statio_all_sequences.relname,
    pg_statio_all_sequences.blks_read,
    pg_statio_all_sequences.blks_hit
   FROM pg_statio_all_sequences
  WHERE ((pg_statio_all_sequences.schemaname = ANY (ARRAY['pg_catalog'::name, 'information_schema'::name])) OR (pg_statio_all_sequences.schemaname ~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_statio_user_sequences"), T(r#" SELECT pg_statio_all_sequences.relid,
    pg_statio_all_sequences.schemaname,
    pg_statio_all_sequences.relname,
    pg_statio_all_sequences.blks_read,
    pg_statio_all_sequences.blks_hit
   FROM pg_statio_all_sequences
  WHERE ((pg_statio_all_sequences.schemaname <> ALL (ARRAY['pg_catalog'::name, 'information_schema'::name])) AND (pg_statio_all_sequences.schemaname !~ '^pg_toast'::text));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_subscription"), T(r#" SELECT su.oid AS subid,
    su.subname,
    st.pid,
    st.relid,
    st.received_lsn,
    st.last_msg_send_time,
    st.last_msg_receipt_time,
    st.latest_end_lsn,
    st.latest_end_time
   FROM (pg_subscription su
     LEFT JOIN pg_stat_get_subscription(NULL::oid) st(subid, relid, pid, received_lsn, last_msg_send_time, last_msg_receipt_time, latest_end_lsn, latest_end_time) ON ((st.subid = su.oid)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_ssl"), T(r#" SELECT s.pid,
    s.ssl,
    s.sslversion AS version,
    s.sslcipher AS cipher,
    s.sslbits AS bits,
    s.ssl_client_dn AS client_dn,
    s.ssl_client_serial AS client_serial,
    s.ssl_issuer_dn AS issuer_dn
   FROM pg_stat_get_activity(NULL::integer) s(datid, pid, usesysid, application_name, state, query, wait_event_type, wait_event, xact_start, query_start, backend_start, state_change, client_addr, client_hostname, client_port, backend_xid, backend_xmin, backend_type, ssl, sslversion, sslcipher, sslbits, ssl_client_dn, ssl_client_serial, ssl_issuer_dn, gss_auth, gss_princ, gss_enc, leader_pid, query_id)
  WHERE (s.client_port IS NOT NULL);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_gssapi"), T(r#" SELECT s.pid,
    s.gss_auth AS gss_authenticated,
    s.gss_princ AS principal,
    s.gss_enc AS encrypted
   FROM pg_stat_get_activity(NULL::integer) s(datid, pid, usesysid, application_name, state, query, wait_event_type, wait_event, xact_start, query_start, backend_start, state_change, client_addr, client_hostname, client_port, backend_xid, backend_xmin, backend_type, ssl, sslversion, sslcipher, sslbits, ssl_client_dn, ssl_client_serial, ssl_issuer_dn, gss_auth, gss_princ, gss_enc, leader_pid, query_id)
  WHERE (s.client_port IS NOT NULL);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_replication_slots"), T(r#" SELECT l.slot_name,
    l.plugin,
    l.slot_type,
    l.datoid,
    d.datname AS database,
    l.temporary,
    l.active,
    l.active_pid,
    l.xmin,
    l.catalog_xmin,
    l.restart_lsn,
    l.confirmed_flush_lsn,
    l.wal_status,
    l.safe_wal_size,
    l.two_phase
   FROM (pg_get_replication_slots() l(slot_name, plugin, slot_type, datoid, temporary, active, active_pid, xmin, catalog_xmin, restart_lsn, confirmed_flush_lsn, wal_status, safe_wal_size, two_phase)
     LEFT JOIN pg_database d ON ((l.datoid = d.oid)));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_replication_slots"), T(r#" SELECT s.slot_name,
    s.spill_txns,
    s.spill_count,
    s.spill_bytes,
    s.stream_txns,
    s.stream_count,
    s.stream_bytes,
    s.total_txns,
    s.total_bytes,
    s.stats_reset
   FROM pg_replication_slots r,
    LATERAL pg_stat_get_replication_slot((r.slot_name)::text) s(slot_name, spill_txns, spill_count, spill_bytes, stream_txns, stream_count, stream_bytes, total_txns, total_bytes, stats_reset)
  WHERE (r.datoid IS NOT NULL);"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_database"), T(r#" SELECT d.oid AS datid,
    d.datname,
        CASE
            WHEN (d.oid = (0)::oid) THEN 0
            ELSE pg_stat_get_db_numbackends(d.oid)
        END AS numbackends,
    pg_stat_get_db_xact_commit(d.oid) AS xact_commit,
    pg_stat_get_db_xact_rollback(d.oid) AS xact_rollback,
    (pg_stat_get_db_blocks_fetched(d.oid) - pg_stat_get_db_blocks_hit(d.oid)) AS blks_read,
    pg_stat_get_db_blocks_hit(d.oid) AS blks_hit,
    pg_stat_get_db_tuples_returned(d.oid) AS tup_returned,
    pg_stat_get_db_tuples_fetched(d.oid) AS tup_fetched,
    pg_stat_get_db_tuples_inserted(d.oid) AS tup_inserted,
    pg_stat_get_db_tuples_updated(d.oid) AS tup_updated,
    pg_stat_get_db_tuples_deleted(d.oid) AS tup_deleted,
    pg_stat_get_db_conflict_all(d.oid) AS conflicts,
    pg_stat_get_db_temp_files(d.oid) AS temp_files,
    pg_stat_get_db_temp_bytes(d.oid) AS temp_bytes,
    pg_stat_get_db_deadlocks(d.oid) AS deadlocks,
    pg_stat_get_db_checksum_failures(d.oid) AS checksum_failures,
    pg_stat_get_db_checksum_last_failure(d.oid) AS checksum_last_failure,
    pg_stat_get_db_blk_read_time(d.oid) AS blk_read_time,
    pg_stat_get_db_blk_write_time(d.oid) AS blk_write_time,
    pg_stat_get_db_session_time(d.oid) AS session_time,
    pg_stat_get_db_active_time(d.oid) AS active_time,
    pg_stat_get_db_idle_in_transaction_time(d.oid) AS idle_in_transaction_time,
    pg_stat_get_db_sessions(d.oid) AS sessions,
    pg_stat_get_db_sessions_abandoned(d.oid) AS sessions_abandoned,
    pg_stat_get_db_sessions_fatal(d.oid) AS sessions_fatal,
    pg_stat_get_db_sessions_killed(d.oid) AS sessions_killed,
    pg_stat_get_db_stat_reset_time(d.oid) AS stats_reset
   FROM ( SELECT 0 AS oid,
            NULL::name AS datname
        UNION ALL
         SELECT pg_database.oid,
            pg_database.datname
           FROM pg_database) d;"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_database_conflicts"), T(r#" SELECT d.oid AS datid,
    d.datname,
    pg_stat_get_db_conflict_tablespace(d.oid) AS confl_tablespace,
    pg_stat_get_db_conflict_lock(d.oid) AS confl_lock,
    pg_stat_get_db_conflict_snapshot(d.oid) AS confl_snapshot,
    pg_stat_get_db_conflict_bufferpin(d.oid) AS confl_bufferpin,
    pg_stat_get_db_conflict_startup_deadlock(d.oid) AS confl_deadlock
   FROM pg_database d;"#), T("NONE"), T("YES"), T("YES"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_user_functions"), T(r#" SELECT p.oid AS funcid,
    n.nspname AS schemaname,
    p.proname AS funcname,
    pg_stat_get_function_calls(p.oid) AS calls,
    pg_stat_get_function_total_time(p.oid) AS total_time,
    pg_stat_get_function_self_time(p.oid) AS self_time
   FROM (pg_proc p
     LEFT JOIN pg_namespace n ON ((n.oid = p.pronamespace)))
  WHERE ((p.prolang <> (12)::oid) AND (pg_stat_get_function_calls(p.oid) IS NOT NULL));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("pg_catalog"), T("pg_stat_xact_user_functions"), T(r#" SELECT p.oid AS funcid,
    n.nspname AS schemaname,
    p.proname AS funcname,
    pg_stat_get_xact_function_calls(p.oid) AS calls,
    pg_stat_get_xact_function_total_time(p.oid) AS total_time,
    pg_stat_get_xact_function_self_time(p.oid) AS self_time
   FROM (pg_proc p
     LEFT JOIN pg_namespace n ON ((n.oid = p.pronamespace)))
  WHERE ((p.prolang <> (12)::oid) AND (pg_stat_get_xact_function_calls(p.oid) IS NOT NULL));"#), T("NONE"), T("NO"), T("NO"), T("NO"), T("NO"), T("NO")],
                            &[T("postgres"), T("public"), T("test_view"), T(r#" SELECT test_table.id
   FROM test_table;"#), T("NONE"), T("YES"), T("YES"), T("NO"), T("NO"), T("NO")],
                        ],
                        tag: "SELECT 141",
                    },
                    skip: Some("rows tie on the ORDER BY key, and Postgres' in-memory sort leaves ties in another order"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT table_schema FROM information_schema.views order by table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_schema", NAME)],
                        rows: &[
                            &[T("information_schema")],
                            &[T("pg_catalog")],
                            &[T("public")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema FROM information_schema.views group by table_catalog, table_schema order by table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME)],
                        rows: &[
                            &[T("postgres"), T("information_schema")],
                            &[T("postgres"), T("pg_catalog")],
                            &[T("postgres"), T("public")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema, table_name FROM information_schema.views WHERE table_schema='public';",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME)],
                        rows: &[
                            &[T("postgres"), T("public"), T("test_view")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA test_schema;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SEARCH_PATH TO test_schema;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test_table2 (id int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW test_view2 as select * from test_table2;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT table_schema FROM information_schema.views order by table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_schema", NAME)],
                        rows: &[
                            &[T("information_schema")],
                            &[T("pg_catalog")],
                            &[T("public")],
                            &[T("test_schema")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema FROM information_schema.views group by table_catalog, table_schema order by table_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME)],
                        rows: &[
                            &[T("postgres"), T("information_schema")],
                            &[T("postgres"), T("pg_catalog")],
                            &[T("postgres"), T("public")],
                            &[T("postgres"), T("test_schema")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema, table_name FROM information_schema.views WHERE table_schema='test_schema';",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME)],
                        rows: &[
                            &[T("postgres"), T("test_schema"), T("test_view2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_catalog, table_schema, table_name, view_definition FROM information_schema.views WHERE table_schema = 'test_schema' ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME), Column("view_definition", VARCHAR)],
                        rows: &[
                            &[T("postgres"), T("test_schema"), T("test_view2"), T(r#" SELECT test_table2.id
   FROM test_table2;"#)],
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
fn test_check_constraint_view_rules() {
    run_scripts(&[
        ScriptTest {
            name: "check constraints view",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t3333 (z TEXT PRIMARY KEY CHECK (z ~ '^[0-9]+$'), y TEXT CONSTRAINT y_chk CHECK (regexp_like(y, '^[a-z]+$')), w int NOT NULL CHECK (w > 0 AND w < 10));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT constraint_schema, constraint_name, check_clause FROM information_schema.check_constraints WHERE constraint_schema = 'public' AND constraint_name NOT LIKE '%not_null' ORDER BY 2;",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_schema", NAME), Column("constraint_name", NAME), Column("check_clause", VARCHAR)],
                        rows: &[
                            &[T("public"), T("t3333_w_check"), T("(((w > 0) AND (w < 10)))")],
                            &[T("public"), T("t3333_z_check"), T("((z ~ '^[0-9]+$'::text))")],
                            &[T("public"), T("y_chk"), T("(regexp_like(y, '^[a-z]+$'::text))")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM information_schema.check_constraints WHERE constraint_schema = 'public' AND check_clause = 'w IS NOT NULL';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
            name: "names of scalar subqueries",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT (SELECT 1)::int, (SELECT 'x')::text, (SELECT 1 AS a)::int, (SELECT 1 UNION SELECT 2 LIMIT 1)::int;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4), Column("?column?", TEXT), Column("a", INT4), Column("?column?", INT4)],
                        rows: &[
                            &[T("1"), T("x"), T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN dn AS integer CONSTRAINT dn_check1 CHECK (VALUE > 300) CHECK (VALUE < 400) CHECK (VALUE <> 350);",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname FROM pg_constraint WHERE conname LIKE 'dn_check%' ORDER BY conname;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME)],
                        rows: &[
                            &[T("dn_check")],
                            &[T("dn_check1")],
                            &[T("dn_check2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
