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
fn test_create_functions_language_sql() {
    run_scripts(&[
        ScriptTest {
            name: "unnamed parameter",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alt_func1(int) RETURNS int LANGUAGE sql AS 'SELECT $1 + 1';",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alt_func1(3);",
                    expected: Expected::Rows {
                        columns: &[Column("alt_func1", INT4)],
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
            name: "default on input parameters",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alt_func1(int = 2, int) RETURNS int LANGUAGE sql AS 'SELECT $1 + $2';",
                    expected: Expected::Error(Diagnostic { code: "42P13", message: "input parameters after one with a default value must also have defaults", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alt_func1(int = 2, int = 3) RETURNS int LANGUAGE sql AS 'SELECT $1 + $2';",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alt_func1();",
                    expected: Expected::Rows {
                        columns: &[Column("alt_func1", INT4)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alt_func1(12);",
                    expected: Expected::Rows {
                        columns: &[Column("alt_func1", INT4)],
                        rows: &[
                            &[T("15")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "named parameter",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alt_func1(x int) RETURNS int LANGUAGE sql AS 'SELECT x + 1';",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alt_func1(3);",
                    expected: Expected::Rows {
                        columns: &[Column("alt_func1", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION sub_numbers(x int, y int) RETURNS int LANGUAGE sql AS 'SELECT y - x';",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sub_numbers(1, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("sub_numbers", INT4)],
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
            name: "nil default on input parameters",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"create function dfunc(a varchar = 'def a', out _a varchar, c numeric = NULL, out _c numeric)
							returns record as $$ select $1, $2; $$ language sql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select dfunc('Hello');",
                    expected: Expected::Rows {
                        columns: &[Column("dfunc", RECORD)],
                        rows: &[
                            &[T("(Hello,)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from dfunc('Hello');",
                    expected: Expected::Rows {
                        columns: &[Column("_a", VARCHAR), Column("_c", NUMERIC)],
                        rows: &[
                            &[T("Hello"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "unknown functions",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION get_grade_description(score INT)
							RETURNS TEXT
							LANGUAGE SQL
							AS $$
								SELECT
									CASE
										WHEN score >= 90 THEN 'Excellent'
										WHEN score >= 75 THEN 'Good'
										WHEN score >= 50 THEN 'Average'
									ELSE 'Fail'
									END;
							$$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT get_grade_description(92);",
                    expected: Expected::Rows {
                        columns: &[Column("get_grade_description", TEXT)],
                        rows: &[
                            &[T("Excellent")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT get_grade_description(65);",
                    expected: Expected::Rows {
                        columns: &[Column("get_grade_description", TEXT)],
                        rows: &[
                            &[T("Average")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nested functions",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION calculate_double_sum(x INT, y INT)
							RETURNS INT
							LANGUAGE SQL
							AS $$
								SELECT add_numbers(x, y) * 2;
							$$;"#,
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function add_numbers(integer, integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 119, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION add_numbers(int, int) RETURNS int LANGUAGE sql AS 'SELECT $1 + $2';",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION calculate_double_sum(x INT, y INT)
							RETURNS INT
							LANGUAGE SQL
							AS $$
								SELECT add_numbers(x, y) * 2;
							$$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT calculate_double_sum(1, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("calculate_double_sum", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "function returning multiple rows",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION gen(a int) RETURNS SETOF INT LANGUAGE SQL AS $$ SELECT generate_series(1, a) $$ STABLE;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM gen(3);",
                    expected: Expected::Rows {
                        columns: &[Column("gen", INT4)],
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
            name: "function with create or replace view",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION public.sp_build_view_bathymetry_layer() RETURNS void
							LANGUAGE sql
							AS $$
								CREATE OR REPLACE VIEW public.view_bathymetry_layer AS
								SELECT 1;
							$$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT public.sp_build_view_bathymetry_layer()",
                    expected: Expected::Rows {
                        columns: &[Column("sp_build_view_bathymetry_layer", VOID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * from view_bathymetry_layer",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT public.sp_build_view_bathymetry_layer()",
                    expected: Expected::Rows {
                        columns: &[Column("sp_build_view_bathymetry_layer", VOID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * from view_bathymetry_layer",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
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
            name: "function with update ... returning",
            set_up_script: &[
                r#"CREATE TYPE public.tax_job_state AS ENUM (
					'sched',
					'busy',
					'final',
					'help'
				);"#,
                r#"CREATE TABLE public.tax_job (
					id bigint NOT NULL,
					state public.tax_job_state NOT NULL,
					created timestamp NOT NULL,
					modified timestamp NOT NULL,
					scheduled timestamp,
					worker text,
					processor text,
					ext_id text,
					data jsonb,
					gross integer,
					notes text[],
					ops jsonb,
					CONSTRAINT tax_job_check CHECK ((NOT ((state = 'sched'::public.tax_job_state) AND (scheduled IS NULL)))),
					CONSTRAINT tax_job_check1 CHECK ((NOT ((state = 'busy'::public.tax_job_state) AND (worker IS NULL))))
				);"#,
                "INSERT INTO tax_job (id, state, created, modified, scheduled, worker, processor, ext_id, data) VALUES (1, 'sched', '2025-05-05 05:05:05', '2025-05-05 05:05:05', '2025-05-05 05:05:05', 'worker', 'processor', 'ext_id', NULL)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION public.tax_job_take(arg_worker text) RETURNS SETOF public.tax_job
								LANGUAGE sql
								AS '
								UPDATE
									tax_job
								SET
									state = ''busy'',
									worker = arg_worker
								WHERE
									state = ''sched''
									AND scheduled <= CURRENT_TIMESTAMP
								RETURNING
									*;
							';"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT public.tax_job_take('worker')",
                    expected: Expected::Rows {
                        columns: &[Column("tax_job_take", USER_DEFINED)],
                        rows: &[
                            &[T(r#"(1,busy,"2025-05-05 05:05:05","2025-05-05 05:05:05","2025-05-05 05:05:05",worker,processor,ext_id,,,,)"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tax_job (id, state, created, modified, scheduled, worker, processor, ext_id, data) VALUES (2, 'sched', '2025-05-05 05:05:06', '2025-05-05 05:05:06', '2025-05-05 05:05:06', 'worker', 'processor', 'ext_id', NULL), (3, 'sched', '2025-05-05 05:05:07', '2025-05-05 05:05:07', '2025-05-05 05:05:07', 'worker', 'processor', 'ext_id', NULL)",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT public.tax_job_take('worker')",
                    expected: Expected::Rows {
                        columns: &[Column("tax_job_take", USER_DEFINED)],
                        rows: &[
                            &[T(r#"(2,busy,"2025-05-05 05:05:06","2025-05-05 05:05:06","2025-05-05 05:05:06",worker,processor,ext_id,,,,)"#)],
                            &[T(r#"(3,busy,"2025-05-05 05:05:07","2025-05-05 05:05:07","2025-05-05 05:05:07",worker,processor,ext_id,,,,)"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "function with delete",
            set_up_script: &[
                "CREATE TABLE test (id bigint NOT NULL, state text NOT NULL);",
                "INSERT INTO test VALUES (1, 'sched'), (2, 'busy');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION d(w text) RETURNS bigint
								LANGUAGE sql
								AS '
								DELETE FROM test
								WHERE
									state = w
								RETURNING
									id;
							';"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("state", TEXT)],
                        rows: &[
                            &[T("1"), T("sched")],
                            &[T("2"), T("busy")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT d('sched');",
                    expected: Expected::Rows {
                        columns: &[Column("d", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("state", TEXT)],
                        rows: &[
                            &[T("2"), T("busy")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "multiple statements in function",
            set_up_script: &[
                "CREATE TABLE test (id int);",
                "INSERT INTO test VALUES (1), (2), (3);",
                "CREATE VIEW test1 AS SELECT * FROM test WHERE id = 1;",
                "CREATE VIEW test2 AS SELECT * FROM test WHERE id = 2;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION drop_views() RETURNS void
								LANGUAGE sql
								AS $$
							DROP VIEW test1;
							DROP VIEW test2;
							$$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test1",
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
                    query: "SELECT * FROM test2",
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
                    query: "SELECT drop_views();",
                    expected: Expected::Rows {
                        columns: &[Column("drop_views", VOID)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test1",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test1" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test2" does not exist"#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "function with default expression in parameter",
            set_up_script: &[
                "CREATE TABLE cp_test (a int, b text);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE FUNCTION dfunc(e int, d text, f int default 100)
							 RETURNS int LANGUAGE SQL
							AS $$
								INSERT INTO cp_test VALUES(e+f, d);
								SELECT a FROM cp_test WHERE b = d;
							$$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE FUNCTION dfunc(e int, f int default 100)
							 RETURNS int LANGUAGE SQL
							AS $$
								INSERT INTO cp_test VALUES(e+f, 'seconddfunc');
								SELECT a FROM cp_test WHERE b = 'seconddfunc';
							$$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM dfunc(10, 'Hello', 20);",
                    expected: Expected::Rows {
                        columns: &[Column("dfunc", INT4)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM cp_test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("30"), T("Hello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM dfunc(50, 'Bye');",
                    expected: Expected::Rows {
                        columns: &[Column("dfunc", INT4)],
                        rows: &[
                            &[T("150")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM cp_test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("30"), T("Hello")],
                            &[T("150"), T("Bye")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dfunc(2, 'After');",
                    expected: Expected::Rows {
                        columns: &[Column("dfunc", INT4)],
                        rows: &[
                            &[T("102")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM cp_test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("30"), T("Hello")],
                            &[T("150"), T("Bye")],
                            &[T("102"), T("After")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE FUNCTION dfunc(e int, f text default '100')
							 RETURNS int LANGUAGE SQL
							AS $$
								INSERT INTO cp_test VALUES(e, f);
								SELECT a FROM cp_test WHERE b = f;
							$$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dfunc(50);",
                    expected: Expected::Error(Diagnostic { code: "42725", message: "function dfunc(integer) is not unique", hint: "Could not choose a best candidate function. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "use sql statements in BEGIN ATOMIC ... END in sql_body",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION match_default() RETURNS jsonb
            LANGUAGE sql
            BEGIN ATOMIC 
				SELECT jsonb_build_object('k', 6, 'm', 2048, 'include_original', true, 'tokenizer', json_build_object('kind', 'ngram', 'token_length', 3), 'token_filters', json_build_array(json_build_object('kind', 'downcase'))) AS jsonb_build_object; 
			END;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT public.match_default();",
                    expected: Expected::Rows {
                        columns: &[Column("match_default", JSONB)],
                        rows: &[
                            &[T(r#"{"k": 6, "m": 2048, "tokenizer": {"kind": "ngram", "token_length": 3}, "token_filters": [{"kind": "downcase"}], "include_original": true}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION select1() RETURNS int
            LANGUAGE sql
            BEGIN ATOMIC 
				SELECT 1; 
			END;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT select1();",
                    expected: Expected::Rows {
                        columns: &[Column("select1", INT4)],
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
            name: "use RETURN in BEGIN ATOMIC ... END in sql_body",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION return1() RETURNS text
            LANGUAGE sql
            BEGIN ATOMIC 
				RETURN 1::text || 'one'; 
			END;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT return1();",
                    expected: Expected::Rows {
                        columns: &[Column("return1", TEXT)],
                        rows: &[
                            &[T("1one")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "user-defined composite array parameter and return type",
            set_up_script: &[
                "CREATE TYPE aggtype AS (a integer, b integer, c text)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION aggf_trans(aggtype[],integer,integer,text) RETURNS aggtype[] AS 'select array_append($1,ROW($2,$3,$4)::aggtype)' LANGUAGE sql STRICT IMMUTABLE",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT aggf_trans(ARRAY[]::aggtype[], 1, 2, 'hello')",
                    expected: Expected::Rows {
                        columns: &[Column("aggf_trans", USER_DEFINED)],
                        rows: &[
                            &[T(r#"{"(1,2,hello)"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT aggf_trans(ARRAY[ROW(1,2,'hello')::aggtype], 3, 4, 'world')",
                    expected: Expected::Rows {
                        columns: &[Column("aggf_trans", USER_DEFINED)],
                        rows: &[
                            &[T(r#"{"(1,2,hello)","(3,4,world)"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "table function in FROM returning a table",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION figures() RETURNS TABLE(shape TEXT, sides INT) LANGUAGE SQL AS $$ SELECT 'triangle', 3 UNION ALL SELECT 'square', 4 $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM figures();",
                    expected: Expected::Rows {
                        columns: &[Column("shape", TEXT), Column("sides", INT4)],
                        rows: &[
                            &[T("triangle"), T("3")],
                            &[T("square"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT shape, sides FROM figures();",
                    expected: Expected::Rows {
                        columns: &[Column("shape", TEXT), Column("sides", INT4)],
                        rows: &[
                            &[T("triangle"), T("3")],
                            &[T("square"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT shape FROM figures() WHERE sides = 4;",
                    expected: Expected::Rows {
                        columns: &[Column("shape", TEXT)],
                        rows: &[
                            &[T("square")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.figures();",
                    expected: Expected::Rows {
                        columns: &[Column("shape", TEXT), Column("sides", INT4)],
                        rows: &[
                            &[T("triangle"), T("3")],
                            &[T("square"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM figures() AS f(name, edges);",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("edges", INT4)],
                        rows: &[
                            &[T("triangle"), T("3")],
                            &[T("square"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT figures();",
                    expected: Expected::Rows {
                        columns: &[Column("figures", RECORD)],
                        rows: &[
                            &[T("(triangle,3)")],
                            &[T("(square,4)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "table function in FROM returning SETOF a table type",
            set_up_script: &[
                "CREATE TABLE shapes (shape TEXT, sides INT);",
                "INSERT INTO shapes VALUES ('triangle', 3), ('square', 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION with_sides(n INT) RETURNS SETOF shapes LANGUAGE SQL AS $$ SELECT * FROM shapes WHERE sides >= n $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM with_sides(4);",
                    expected: Expected::Rows {
                        columns: &[Column("shape", TEXT), Column("sides", INT4)],
                        rows: &[
                            &[T("square"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT s.shape FROM shapes s JOIN with_sides(3) w ON s.shape = w.shape ORDER BY s.shape;",
                    expected: Expected::Rows {
                        columns: &[Column("shape", TEXT)],
                        rows: &[
                            &[T("square")],
                            &[T("triangle")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT with_sides(4);",
                    expected: Expected::Rows {
                        columns: &[Column("with_sides", USER_DEFINED)],
                        rows: &[
                            &[T("(square,4)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "table function in FROM returning a composite type",
            set_up_script: &[
                "CREATE TYPE figure AS (shape TEXT, sides INT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION a_figure() RETURNS figure LANGUAGE SQL AS $$ SELECT ROW('triangle', 3)::figure $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM a_figure();",
                    expected: Expected::Rows {
                        columns: &[Column("shape", TEXT), Column("sides", INT4)],
                        rows: &[
                            &[T("triangle"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a_figure();",
                    expected: Expected::Rows {
                        columns: &[Column("a_figure", USER_DEFINED)],
                        rows: &[
                            &[T("(triangle,3)")],
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
