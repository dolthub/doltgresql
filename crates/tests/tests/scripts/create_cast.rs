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
fn test_casts() {
    run_scripts(&[
        ScriptTest {
            name: "CREATE CAST with function creates explicit-only cast",
            set_up_script: &[
                "CREATE TABLE cast_explicit_src (v text);",
                "CREATE TABLE cast_explicit_dst (v text, tag text);",
                r#"CREATE FUNCTION cast_explicit_fn(src cast_explicit_src) RETURNS cast_explicit_dst
					AS $$ SELECT ROW((src).v, 'explicit')::cast_explicit_dst $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_explicit_accept(dst cast_explicit_dst) RETURNS text
					AS $$ SELECT (dst).v || ':' || (dst).tag $$ LANGUAGE SQL;"#,
                "CREATE TABLE cast_explicit_holder (v cast_explicit_dst);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_explicit_src AS cast_explicit_dst) WITH FUNCTION cast_explicit_fn(cast_explicit_src);",
                    expected: Expected::Tag("CREATE CAST"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_explicit_accept((ROW('one')::cast_explicit_src)::cast_explicit_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_explicit_accept", TEXT)],
                        rows: &[
                            &[T("one:explicit")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_explicit_accept(ROW('one')::cast_explicit_src);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function cast_explicit_accept(cast_explicit_src) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO cast_explicit_holder VALUES (ROW('one')::cast_explicit_src);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v" is of type cast_explicit_dst but expression is of type cast_explicit_src"#, hint: "You will need to rewrite or cast the expression.", position: 52, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_explicit_src'
						  AND dst.typname = 'cast_explicit_dst';"#,
                    expected: Expected::Rows {
                        columns: &[Column("castcontext", TEXT), Column("castmethod", TEXT)],
                        rows: &[
                            &[T("e"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_explicit_src AS cast_explicit_dst) WITH FUNCTION cast_explicit_fn(cast_explicit_src);",
                    expected: Expected::Error(Diagnostic { code: "42710", message: "cast from type cast_explicit_src to type cast_explicit_dst already exists", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE CAST with PL/pgSQL function creates explicit-only cast",
            set_up_script: &[
                "CREATE TABLE cast_explicit_src (v text);",
                "CREATE TABLE cast_explicit_dst (v text, tag text);",
                r#"CREATE FUNCTION cast_explicit_fn(src cast_explicit_src) RETURNS cast_explicit_dst AS $$ BEGIN
						RETURN ROW((src).v, 'explicit')::cast_explicit_dst;
					END; $$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION cast_explicit_accept(dst cast_explicit_dst) RETURNS text AS $$ BEGIN
						RETURN (dst).v || ':' || (dst).tag;
					END; $$ LANGUAGE plpgsql;"#,
                "CREATE TABLE cast_explicit_holder (v cast_explicit_dst);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_explicit_src AS cast_explicit_dst) WITH FUNCTION cast_explicit_fn(cast_explicit_src);",
                    expected: Expected::Tag("CREATE CAST"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_explicit_accept((ROW('one')::cast_explicit_src)::cast_explicit_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_explicit_accept", TEXT)],
                        rows: &[
                            &[T("one:explicit")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_explicit_accept(ROW('one')::cast_explicit_src);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function cast_explicit_accept(cast_explicit_src) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO cast_explicit_holder VALUES (ROW('one')::cast_explicit_src);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"column "v" is of type cast_explicit_dst but expression is of type cast_explicit_src"#, hint: "You will need to rewrite or cast the expression.", position: 52, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_explicit_src'
						  AND dst.typname = 'cast_explicit_dst';"#,
                    expected: Expected::Rows {
                        columns: &[Column("castcontext", TEXT), Column("castmethod", TEXT)],
                        rows: &[
                            &[T("e"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_explicit_src AS cast_explicit_dst) WITH FUNCTION cast_explicit_fn(cast_explicit_src);",
                    expected: Expected::Error(Diagnostic { code: "42710", message: "cast from type cast_explicit_src to type cast_explicit_dst already exists", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE CAST AS ASSIGNMENT works for assignment contexts only",
            set_up_script: &[
                "CREATE TABLE cast_assignment_src (v text);",
                "CREATE TABLE cast_assignment_dst (v text, tag text);",
                r#"CREATE FUNCTION cast_assignment_fn(cast_assignment_src) RETURNS cast_assignment_dst
					AS $$ SELECT ROW(($1).v, 'assignment')::cast_assignment_dst $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_assignment_accept(cast_assignment_dst) RETURNS text
					AS $$ SELECT ($1).v || ':' || ($1).tag $$ LANGUAGE SQL;"#,
                "CREATE TABLE cast_assignment_holder (v cast_assignment_dst);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_assignment_src AS cast_assignment_dst) WITH FUNCTION cast_assignment_fn(cast_assignment_src) AS ASSIGNMENT;",
                    expected: Expected::Tag("CREATE CAST"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_assignment_accept((ROW('on')::cast_assignment_src)::cast_assignment_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_assignment_accept", TEXT)],
                        rows: &[
                            &[T("on:assignment")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO cast_assignment_holder VALUES (ROW('on')::cast_assignment_src);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_assignment_accept(v) FROM cast_assignment_holder;",
                    expected: Expected::Rows {
                        columns: &[Column("cast_assignment_accept", TEXT)],
                        rows: &[
                            &[T("on:assignment")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_assignment_accept(ROW('off')::cast_assignment_src);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function cast_assignment_accept(cast_assignment_src) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_assignment_src'
						  AND dst.typname = 'cast_assignment_dst';"#,
                    expected: Expected::Rows {
                        columns: &[Column("castcontext", TEXT), Column("castmethod", TEXT)],
                        rows: &[
                            &[T("a"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE CAST AS IMPLICIT works for function resolution",
            set_up_script: &[
                "CREATE TABLE cast_implicit_src (v text);",
                "CREATE TABLE cast_implicit_dst (v text, tag text);",
                r#"CREATE FUNCTION cast_implicit_fn(cast_implicit_src) RETURNS cast_implicit_dst
					AS $$ SELECT ROW(($1).v, 'implicit')::cast_implicit_dst $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_implicit_accept(cast_implicit_dst) RETURNS text
					AS $$ SELECT ($1).v || ':' || ($1).tag $$ LANGUAGE SQL;"#,
                "CREATE TABLE cast_implicit_holder (v cast_implicit_dst);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_implicit_src AS cast_implicit_dst) WITH FUNCTION cast_implicit_fn(cast_implicit_src) AS IMPLICIT;",
                    expected: Expected::Tag("CREATE CAST"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_implicit_accept((ROW('x')::cast_implicit_src)::cast_implicit_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_implicit_accept", TEXT)],
                        rows: &[
                            &[T("x:implicit")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_implicit_accept(ROW('y')::cast_implicit_src);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_implicit_accept", TEXT)],
                        rows: &[
                            &[T("y:implicit")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO cast_implicit_holder VALUES (ROW('z')::cast_implicit_src);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_implicit_accept(v) FROM cast_implicit_holder;",
                    expected: Expected::Rows {
                        columns: &[Column("cast_implicit_accept", TEXT)],
                        rows: &[
                            &[T("z:implicit")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_implicit_src'
						  AND dst.typname = 'cast_implicit_dst';"#,
                    expected: Expected::Rows {
                        columns: &[Column("castcontext", TEXT), Column("castmethod", TEXT)],
                        rows: &[
                            &[T("i"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE CAST WITH INOUT",
            set_up_script: &[
                "CREATE TABLE cast_inout_src (v text);",
                "CREATE TABLE cast_inout_dst (v int);",
                r#"CREATE FUNCTION cast_inout_accept(cast_inout_dst) RETURNS text
					AS $$ SELECT (($1).v)::text $$ LANGUAGE SQL;"#,
                "CREATE TABLE cast_inout_holder (v cast_inout_dst);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_inout_src AS cast_inout_dst) WITH INOUT AS ASSIGNMENT;",
                    expected: Expected::Tag("CREATE CAST"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_inout_accept((ROW('42')::cast_inout_src)::cast_inout_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_inout_accept", TEXT)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO cast_inout_holder VALUES (ROW('99')::cast_inout_src);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_inout_accept(v) FROM cast_inout_holder;",
                    expected: Expected::Rows {
                        columns: &[Column("cast_inout_accept", TEXT)],
                        rows: &[
                            &[T("99")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_inout_accept((ROW('not_an_int')::cast_inout_src)::cast_inout_dst);",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "not_an_int""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_inout_src'
						  AND dst.typname = 'cast_inout_dst';"#,
                    expected: Expected::Rows {
                        columns: &[Column("castcontext", TEXT), Column("castmethod", TEXT)],
                        rows: &[
                            &[T("a"), T("i")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE CAST with three-argument function receives explicit flag",
            set_up_script: &[
                "CREATE TABLE cast_three_arg_src (v text);",
                "CREATE TABLE cast_three_arg_dst (v text, tag text);",
                r#"CREATE FUNCTION cast_three_arg_fn(cast_three_arg_src, integer, boolean) RETURNS cast_three_arg_dst
					AS $$
						SELECT ROW(
							($1).v,
							CASE WHEN $3 THEN 'explicit' ELSE 'implicit_or_assignment' END
						)::cast_three_arg_dst
					$$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_three_arg_accept(cast_three_arg_dst) RETURNS text
					AS $$ SELECT ($1).v || ':' || ($1).tag $$ LANGUAGE SQL;"#,
                "CREATE TABLE cast_three_arg_holder (v cast_three_arg_dst);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_three_arg_src AS cast_three_arg_dst) WITH FUNCTION cast_three_arg_fn(cast_three_arg_src, integer, boolean) AS IMPLICIT;",
                    expected: Expected::Tag("CREATE CAST"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_three_arg_accept((ROW('a')::cast_three_arg_src)::cast_three_arg_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_three_arg_accept", TEXT)],
                        rows: &[
                            &[T("a:explicit")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_three_arg_accept(ROW('b')::cast_three_arg_src);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_three_arg_accept", TEXT)],
                        rows: &[
                            &[T("b:implicit_or_assignment")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO cast_three_arg_holder VALUES (ROW('c')::cast_three_arg_src);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_three_arg_accept(v) FROM cast_three_arg_holder;",
                    expected: Expected::Rows {
                        columns: &[Column("cast_three_arg_accept", TEXT)],
                        rows: &[
                            &[T("c:implicit_or_assignment")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_three_arg_src'
						  AND dst.typname = 'cast_three_arg_dst';"#,
                    expected: Expected::Rows {
                        columns: &[Column("castcontext", TEXT), Column("castmethod", TEXT)],
                        rows: &[
                            &[T("i"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP CAST removes catalog entry and allows recreation",
            set_up_script: &[
                "CREATE TABLE cast_drop_src (v text);",
                "CREATE TABLE cast_drop_dst (v text, tag text);",
                r#"CREATE FUNCTION cast_drop_fn(cast_drop_src) RETURNS cast_drop_dst
					AS $$ SELECT ROW(($1).v, 'drop')::cast_drop_dst $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_drop_accept(cast_drop_dst) RETURNS text
					AS $$ SELECT ($1).v || ':' || ($1).tag $$ LANGUAGE SQL;"#,
                "CREATE CAST (cast_drop_src AS cast_drop_dst) WITH FUNCTION cast_drop_fn(cast_drop_src);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT cast_drop_accept((ROW('before')::cast_drop_src)::cast_drop_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_drop_accept", TEXT)],
                        rows: &[
                            &[T("before:drop")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT EXISTS (
						SELECT 1
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_drop_src'
						  AND dst.typname = 'cast_drop_dst'
					);"#,
                    expected: Expected::Rows {
                        columns: &[Column("exists", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP CAST (cast_drop_src AS cast_drop_dst);",
                    expected: Expected::Tag("DROP CAST"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT EXISTS (
						SELECT 1
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_drop_src'
						  AND dst.typname = 'cast_drop_dst'
					);"#,
                    expected: Expected::Rows {
                        columns: &[Column("exists", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_drop_accept((ROW('after')::cast_drop_src)::cast_drop_dst);",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type cast_drop_src to cast_drop_dst", position: 54, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_drop_src AS cast_drop_dst) WITH FUNCTION cast_drop_fn(cast_drop_src);",
                    expected: Expected::Tag("CREATE CAST"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_drop_accept((ROW('after')::cast_drop_src)::cast_drop_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_drop_accept", TEXT)],
                        rows: &[
                            &[T("after:drop")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE CAST function is invoked for NULL when function is not STRICT",
            set_up_script: &[
                "CREATE TABLE cast_null_src (v text);",
                "CREATE TABLE cast_null_dst (v text, tag text);",
                r#"CREATE FUNCTION cast_null_fn(cast_null_src) RETURNS cast_null_dst
					AS $$ SELECT CASE
						WHEN $1 IS NULL THEN ROW('saw_null', 'called')::cast_null_dst
						ELSE ROW('hi', 'nonnull')::cast_null_dst
					END $$ LANGUAGE SQL;"#,
                "CREATE CAST (cast_null_src AS cast_null_dst) WITH FUNCTION cast_null_fn(cast_null_src);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ((NULL::cast_null_src)::cast_null_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_null_dst", USER_DEFINED)],
                        rows: &[
                            &[T("(saw_null,called)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ((ROW('hi')::cast_null_src)::cast_null_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("row", USER_DEFINED)],
                        rows: &[
                            &[T("(hi,nonnull)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE CAST STRICT function is not invoked for NULL",
            set_up_script: &[
                "CREATE TABLE cast_strict_null_src (v text);",
                "CREATE TABLE cast_strict_null_dst (v text, tag text);",
                r#"CREATE FUNCTION cast_strict_null_fn(cast_strict_null_src) RETURNS cast_strict_null_dst
					AS $$ SELECT ROW('bad', 'called')::cast_strict_null_dst $$ LANGUAGE SQL STRICT;"#,
                "CREATE CAST (cast_strict_null_src AS cast_strict_null_dst) WITH FUNCTION cast_strict_null_fn(cast_strict_null_src);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ((NULL::cast_strict_null_src)::cast_strict_null_dst) IS NULL;",
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
                    query: "SELECT ((ROW('there')::cast_strict_null_src)::cast_strict_null_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("row", USER_DEFINED)],
                        rows: &[
                            &[T("(bad,called)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE CAST validation",
            set_up_script: &[
                "CREATE TABLE cast_bad_src (v text);",
                "CREATE TABLE cast_bad_dst (v text, tag text);",
                r#"CREATE FUNCTION cast_bad_wrong_return(cast_bad_src) RETURNS text
					AS $$ SELECT ($1).v $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_bad_wrong_source(text) RETURNS cast_bad_dst
					AS $$ SELECT ROW($1, 'wrong_source')::cast_bad_dst $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_bad_wrong_second(cast_bad_src, text) RETURNS cast_bad_dst
					AS $$ SELECT ROW(($1).v, 'wrong_second')::cast_bad_dst $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_bad_wrong_third(cast_bad_src, integer, integer) RETURNS cast_bad_dst
					AS $$ SELECT ROW(($1).v, 'wrong_third')::cast_bad_dst $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_bad_too_many(cast_bad_src, integer, boolean, integer) RETURNS cast_bad_dst
					AS $$ SELECT ROW(($1).v, 'too_many')::cast_bad_dst $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_good(cast_bad_src) RETURNS cast_bad_dst
					AS $$ SELECT ROW(($1).v, 'good')::cast_bad_dst $$ LANGUAGE SQL;"#,
                "CREATE TABLE cast_without_src (v text);",
                "CREATE TABLE cast_without_dst (v text, tag text);",
                "CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_good(cast_bad_src);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_missing(cast_bad_src);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function cast_bad_missing(cast_bad_src) does not exist", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_wrong_return(cast_bad_src);",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: "return data type of cast function must match or be binary-coercible to target data type", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_wrong_source(text);",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: "argument of cast function must match or be binary-coercible from source data type", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_wrong_second(cast_bad_src, text);",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: "second argument of cast function must be type integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_wrong_third(cast_bad_src, integer, integer);",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: "third argument of cast function must be type boolean", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_too_many(cast_bad_src, integer, boolean, integer);",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: "cast function must take one to three arguments", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_without_src AS cast_without_dst) WITHOUT FUNCTION;",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: "composite data types are not binary-compatible", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (int4 AS float8) WITHOUT FUNCTION;",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: "source and target data types are not physically compatible", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_bad_src AS cast_bad_src) WITH INOUT;",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: "source data type and target data type are the same", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_good(cast_bad_src);",
                    expected: Expected::Error(Diagnostic { code: "42710", message: "cast from type cast_bad_src to type cast_bad_dst already exists", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
