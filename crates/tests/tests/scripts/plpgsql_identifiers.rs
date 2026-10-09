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
fn test_plpgsql_identifier_casing() {
    run_scripts(&[
        ScriptTest {
            name: "references fold to the declared name",
            set_up_script: &[
                "CREATE TABLE k (id int, nm text);",
                "INSERT INTO k VALUES (1, 'a');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION r1() RETURNS int LANGUAGE plpgsql AS $$
DECLARE MyVar int := 1; BEGIN RETURN MYVAR; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r1();",
                    expected: Expected::Rows {
                        columns: &[Column("r1", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION r2() RETURNS int LANGUAGE plpgsql AS $$
DECLARE "MyVar" int := 1; BEGIN RETURN "MyVar"; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r2();",
                    expected: Expected::Rows {
                        columns: &[Column("r2", INT4)],
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
            name: "assignment and INTO targets fold too",
            set_up_script: &[
                "CREATE TABLE k (id int, nm text);",
                "INSERT INTO k VALUES (1, 'a');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION r3() RETURNS int LANGUAGE plpgsql AS $$
DECLARE myvar int; BEGIN MyVar := 7; RETURN myvar; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r3();",
                    expected: Expected::Rows {
                        columns: &[Column("r3", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION r4() RETURNS int LANGUAGE plpgsql AS $$
DECLARE "MyVar" int; BEGIN "MyVar" := 7; RETURN "MyVar"; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r4();",
                    expected: Expected::Rows {
                        columns: &[Column("r4", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION r7() RETURNS int LANGUAGE plpgsql AS $$
DECLARE myvar int; BEGIN SELECT id INTO MyVar FROM k; RETURN myvar; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r7();",
                    expected: Expected::Rows {
                        columns: &[Column("r7", INT4)],
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
            name: "parameters and RAISE arguments fold too",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION r5(MyParam int) RETURNS int LANGUAGE plpgsql AS $$
BEGIN RETURN MYPARAM; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r5(7);",
                    expected: Expected::Rows {
                        columns: &[Column("r5", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION r6("MyParam" int) RETURNS int LANGUAGE plpgsql AS $$
BEGIN RETURN "MyParam"; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r6(7);",
                    expected: Expected::Rows {
                        columns: &[Column("r6", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION r8() RETURNS void LANGUAGE plpgsql AS $$
DECLARE myvar int := 3; BEGIN RAISE EXCEPTION 'v=%', MyVar; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r8();",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "v=3", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "a quoted name is a different identifier from an unquoted one",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION s1() RETURNS text LANGUAGE plpgsql AS $$
DECLARE myvar text := 'outer';
BEGIN
	DECLARE "MyVar" text := 'inner-quoted';
	BEGIN
		RETURN MyVar;
	END;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT s1();",
                    expected: Expected::Rows {
                        columns: &[Column("s1", TEXT)],
                        rows: &[
                            &[T("outer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION s2() RETURNS int LANGUAGE plpgsql AS $$
DECLARE "MyVar" int := 1; BEGIN RETURN MyVar; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT s2();",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "myvar" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "FOR loop variables fold like any other name",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION l1() RETURNS int LANGUAGE plpgsql AS $$
DECLARE acc int := 0; BEGIN FOR I IN 1..3 LOOP acc := acc + i; END LOOP; RETURN acc; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1();",
                    expected: Expected::Rows {
                        columns: &[Column("l1", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION l2() RETURNS int LANGUAGE plpgsql AS $$
DECLARE acc int := 0; BEGIN FOR "I" IN 1..3 LOOP acc := acc + "I"; END LOOP; RETURN acc; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2();",
                    expected: Expected::Rows {
                        columns: &[Column("l2", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION l3() RETURNS int LANGUAGE plpgsql AS $$
DECLARE acc int := 0; BEGIN FOR "I" IN 1..3 LOOP acc := acc + i; END LOOP; RETURN acc; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l3();",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "i" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "trigger records fold like any other name",
            set_up_script: &[
                "CREATE TABLE t (id int primary key, v int);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION g() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
	IF new.v < 0 THEN RAISE EXCEPTION 'negative %', new.v; END IF;
	RETURN new;
END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER trg BEFORE INSERT ON t FOR EACH ROW EXECUTE FUNCTION g();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, 5);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (2, -1);",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "negative -1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
    ]);
}
