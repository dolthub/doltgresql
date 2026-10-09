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
fn test_drop_function() {
    run_scripts(&[
        ScriptTest {
            name: "Function does not exist",
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP FUNCTION doesnotexist;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: r#"could not find a function named "doesnotexist""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION IF EXISTS doesnotexist;",
                    expected: Expected::Tag("DROP FUNCTION"),
                    notices: &[Diagnostic { code: "00000", message: "function doesnotexist() does not exist, skipping", ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Basic cases",
            set_up_script: &[
                r#"
CREATE FUNCTION func1() RETURNS TEXT AS $$
BEGIN RETURN 'func1'; END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func2(input INT) RETURNS TEXT AS $$
BEGIN RETURN 'func2(INT)'; END;
$$ LANGUAGE plpgsql;"#,
                "CREATE FUNCTION alt_func1(int = 2, int = 3) RETURNS int LANGUAGE sql AS 'SELECT $1 + $2';",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func1(), func2(99);",
                    expected: Expected::Rows {
                        columns: &[Column("func1", TEXT), Column("func2", TEXT)],
                        rows: &[
                            &[T("func1"), T("func2(INT)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func1;",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func2(INT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func1();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function func1() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(99);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function func2(integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION alt_func1();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alt_func1() does not exist", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION alt_func1;",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Optional type information",
            set_up_script: &[
                r#"
CREATE FUNCTION func1() RETURNS TEXT AS $$
BEGIN RETURN 'func1'; END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func2() RETURNS TEXT AS $$
BEGIN RETURN 'func2'; END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func3(input INT) RETURNS TEXT AS $$
BEGIN RETURN 'func3(INT)'; END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func4(input INT) RETURNS TEXT AS $$
BEGIN RETURN 'func4(INT)'; END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func5(input INT, foo TEXT) RETURNS TEXT AS $$
BEGIN RETURN 'func5(INT, TEXT)'; END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func1(), func2(), func3(1), func4(2);",
                    expected: Expected::Rows {
                        columns: &[Column("func1", TEXT), Column("func2", TEXT), Column("func3", TEXT), Column("func4", TEXT)],
                        rows: &[
                            &[T("func1"), T("func2"), T("func3(INT)"), T("func4(INT)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func1(OUT TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func2(OUT paramname TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func3(paramname INT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func4(IN paramname INT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func5(IN paramname INT, IN paramname TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Qualified names",
            set_up_script: &[
                r#"
CREATE FUNCTION func1() RETURNS TEXT AS $$
BEGIN RETURN 'func1'; END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func2(input TEXT) RETURNS TEXT AS $$
BEGIN RETURN 'func2(TEXT)'; END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT current_schema(), current_database();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME), Column("current_database", NAME)],
                        rows: &[
                            &[T("public"), T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func1(), func2('foo');",
                    expected: Expected::Rows {
                        columns: &[Column("func1", TEXT), Column("func2", TEXT)],
                        rows: &[
                            &[T("func1"), T("func2(TEXT)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION public.func1;",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func1();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function func1() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION postgres.public.func2(TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2('w00t');",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function func2(unknown) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Unspecified parameter types",
            set_up_script: &[
                r#"
CREATE FUNCTION func1(input1 TEXT, input2 TEXT) RETURNS int AS $$
BEGIN RETURN 42; END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func2(input1 TEXT) RETURNS int AS $$
BEGIN RETURN 42; END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func2(input1 TEXT, input2 TEXT) RETURNS int AS $$
BEGIN RETURN 42; END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func3(input1 TEXT, input2 TEXT) RETURNS int AS $$
BEGIN RETURN 42; END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func3() RETURNS int AS $$
BEGIN RETURN 42; END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP FUNCTION func1;",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func2;",
                    expected: Expected::Error(Diagnostic { code: "42725", message: r#"function name "func2" is not unique"#, hint: "Specify the argument list to select the function unambiguously.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func3;",
                    expected: Expected::Error(Diagnostic { code: "42725", message: r#"function name "func3" is not unique"#, hint: "Specify the argument list to select the function unambiguously.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Multiple functions",
            set_up_script: &[
                r#"
CREATE FUNCTION func1() RETURNS TEXT AS $$
BEGIN
	RETURN 'func1';
END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func2(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN 'func2(TEXT)';
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func1(), func2('foo');",
                    expected: Expected::Rows {
                        columns: &[Column("func1", TEXT), Column("func2", TEXT)],
                        rows: &[
                            &[T("func1"), T("func2(TEXT)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FUNCTION func1, func2(TExT);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "FUNCTION""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Overloaded functions",
            set_up_script: &[
                r#"
CREATE FUNCTION func2(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN 'func2(TEXT)';
END;
$$ LANGUAGE plpgsql;"#,
                r#"
CREATE FUNCTION func2(input INT) RETURNS TEXT AS $$
BEGIN
	RETURN 'func2(INT)';
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func2('foo'), func2(42);",
                    expected: Expected::Rows {
                        columns: &[Column("func2", TEXT), Column("func2", TEXT)],
                        rows: &[
                            &[T("func2(TEXT)"), T("func2(INT)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func2(TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2('foo'::text);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function func2(text) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(42);",
                    expected: Expected::Rows {
                        columns: &[Column("func2", TEXT)],
                        rows: &[
                            &[T("func2(INT)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION func2(INT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(42);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function func2(integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop function with empty search_path",
            set_up_script: &[
                "SELECT pg_catalog.set_config('search_path', '', false);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP FUNCTION IF EXISTS public.vmstate(s integer);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    notices: &[Diagnostic { code: "00000", message: "function public.vmstate(pg_catalog.int4) does not exist, skipping", ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "user defined type used as parameter",
            set_up_script: &[
                "CREATE TABLE public.trans (vmid integer NOT NULL);",
                r#"CREATE FUNCTION public.tax_job_trans(t public.trans) RETURNS public.trans
    LANGUAGE plpgsql
    AS '
BEGIN
    SELECT * FROM public.trans;
END;
';"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP FUNCTION IF EXISTS public.tax_job_trans(t public.trans);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop non existing function with non existing type",
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP FUNCTION IF EXISTS public.tax_job_trans(t public.trans);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    notices: &[Diagnostic { code: "00000", message: r#"type "public.trans" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
    ]);
}
