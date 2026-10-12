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
fn test_plpgsql_select_into() {
    run_scripts(&[
        ScriptTest {
            name: "INTO targets when the query matches no rows",
            set_up_script: &[
                "CREATE TABLE k (id int, nm text);",
                "INSERT INTO k VALUES (1, 'a'), (2, 'b');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_scalar_miss() RETURNS int LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 99;
	RETURN coalesce(v, -1);
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_scalar_miss();",
                    expected: Expected::Rows {
                        columns: &[Column("f_scalar_miss", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_two_miss() RETURNS text LANGUAGE plpgsql AS $$
DECLARE a int; b text;
BEGIN
	SELECT id, nm INTO a, b FROM k WHERE id = 99;
	RETURN coalesce(a::text, 'NULLa') || '/' || coalesce(b, 'NULLb');
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_two_miss();",
                    expected: Expected::Rows {
                        columns: &[Column("f_two_miss", TEXT)],
                        rows: &[
                            &[T("NULLa/NULLb")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "INTO with several targets assigns every one of them",
            set_up_script: &[
                "CREATE TABLE k (id int, nm text);",
                "INSERT INTO k VALUES (1, 'a'), (2, 'b');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_two() RETURNS text LANGUAGE plpgsql AS $$
DECLARE a int; b text;
BEGIN
	SELECT id, nm INTO a, b FROM k WHERE id = 1;
	RETURN coalesce(a::text, 'NULLa') || '/' || coalesce(b, 'NULLb');
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_two();",
                    expected: Expected::Rows {
                        columns: &[Column("f_two", TEXT)],
                        rows: &[
                            &[T("1/a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "INTO keeps the first row when the query matches several",
            set_up_script: &[
                "CREATE TABLE k (id int, nm text);",
                "INSERT INTO k VALUES (1, 'a'), (2, 'b');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_scalar_multi() RETURNS int LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k ORDER BY id;
	RETURN v;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_scalar_multi();",
                    expected: Expected::Rows {
                        columns: &[Column("f_scalar_multi", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION f_two_multi() RETURNS text LANGUAGE plpgsql AS $$
DECLARE a int; b text;
BEGIN
	SELECT id, nm INTO a, b FROM k ORDER BY id;
	RETURN coalesce(a::text, 'NULLa') || '/' || coalesce(b, 'NULLb');
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f_two_multi();",
                    expected: Expected::Rows {
                        columns: &[Column("f_two_multi", TEXT)],
                        rows: &[
                            &[T("1/a")],
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
