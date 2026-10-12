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
fn test_drop_procedure() {
    run_scripts(&[
        ScriptTest {
            name: "Procedure does not exist",
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP PROCEDURE doesnotexist;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: r#"could not find a procedure named "doesnotexist""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE IF EXISTS doesnotexist;",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    notices: &[Diagnostic { code: "00000", message: "procedure doesnotexist() does not exist, skipping", ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Basic cases",
            set_up_script: &[
                "CREATE PROCEDURE proc1() AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc2(input INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CALL proc1();",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc2(99);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc1;",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc2(INT);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc1();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc1() does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc2(99);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc2(integer) does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Optional type information",
            set_up_script: &[
                "CREATE PROCEDURE proc1() AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc2() AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc3(input INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc4(input INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc5(input INT, foo TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CALL proc1();",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc2();",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc3(1);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc4(2);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc5(3, 'abc');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc1(OUT TEXT);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc2(OUT paramname TEXT);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc3(paramname INT);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc4(IN paramname INT);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc5(IN paramname INT, IN paramname TEXT);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc1();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc1() does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc2();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc2() does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc3(1);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc3(integer) does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc4(2);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc4(integer) does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc5(3, 'abc');",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc5(integer, unknown) does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Qualified names",
            set_up_script: &[
                "CREATE PROCEDURE proc1() AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc2(input TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
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
                    query: "CALL proc1();",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc2('foo');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE public.proc1;",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc1();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc1() does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE postgres.public.proc2(TEXT);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc2('bar');",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc2(unknown) does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Unspecified parameter types",
            set_up_script: &[
                "CREATE PROCEDURE proc1(input1 TEXT, input2 TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc2(input1 TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc2(input1 TEXT, input2 TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc3(input1 TEXT, input2 TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc3() AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc1;",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc2;",
                    expected: Expected::Error(Diagnostic { code: "42725", message: r#"procedure name "proc2" is not unique"#, hint: "Specify the argument list to select the procedure unambiguously.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc3;",
                    expected: Expected::Error(Diagnostic { code: "42725", message: r#"procedure name "proc3" is not unique"#, hint: "Specify the argument list to select the procedure unambiguously.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Overloaded procedures",
            set_up_script: &[
                "CREATE PROCEDURE proc2(input TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE proc2(input INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CALL proc2('foo');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc2(42);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc2(TEXT);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc2('foo'::text);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc2(text) does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc2(42);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE proc2(INT);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL proc2(42);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure proc2(integer) does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
