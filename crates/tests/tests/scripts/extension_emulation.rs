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
fn test_extension_emulation() {
    run_scripts(&[
        ScriptTest {
            name: "Declared types are created with their array type",
            set_up_script: &[
                "CREATE EXTENSION doltgres_test;",
            ],
            skip: Some(r#"setup fails on Postgres ("error running setup query: CREATE EXTENSION doltgres_test;: ERROR: extension \"doltgres_test\" is not available (SQLSTATE 0A000)") and on the Go server ("error running setup query: CREATE EXTENSION doltgres_test;: ERROR: extension \"doltgres_test\" is not available (SQLSTATE XX000)")"#),
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT typname, typtype, typcategory, typlen, typinput::text, typoutput::text FROM pg_type WHERE typname = 'dgtest_upper';",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT typname, typtype, typcategory FROM pg_type WHERE typname = '_dgtest_upper';",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc'::dgtest_upper;",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'MiXeD'::dgtest_upper;",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Declared operators resolve for their operand types",
            set_up_script: &[
                "CREATE EXTENSION doltgres_test;",
            ],
            skip: Some(r#"setup fails on Postgres ("error running setup query: CREATE EXTENSION doltgres_test;: ERROR: extension \"doltgres_test\" is not available (SQLSTATE 0A000)") and on the Go server ("error running setup query: CREATE EXTENSION doltgres_test;: ERROR: extension \"doltgres_test\" is not available (SQLSTATE XX000)")"#),
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'abc'::dgtest_upper = 'ABC'::dgtest_upper;",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc'::dgtest_upper = 'xyz'::dgtest_upper;",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT oprname, oprkind, oprcanhash, oprcanmerge FROM pg_operator WHERE oprcode::text = 'dgtest_upper_eq';",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc'::dgtest_upper <-> 'wxyz'::dgtest_upper;",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc'::dgtest_upper <-> 'xyz'::dgtest_upper;",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Declared casts convert to their target type",
            set_up_script: &[
                "CREATE EXTENSION doltgres_test;",
            ],
            skip: Some(r#"setup fails on Postgres ("error running setup query: CREATE EXTENSION doltgres_test;: ERROR: extension \"doltgres_test\" is not available (SQLSTATE 0A000)") and on the Go server ("error running setup query: CREATE EXTENSION doltgres_test;: ERROR: extension \"doltgres_test\" is not available (SQLSTATE XX000)")"#),
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ('abc'::dgtest_upper)::text || 'def';",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(('abcd'::dgtest_upper)::text);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Declared aggregates accumulate through their transition and final functions",
            set_up_script: &[
                "CREATE EXTENSION doltgres_test;",
                "CREATE TABLE t1 (pk INTEGER PRIMARY KEY, v1 TEXT);",
                "INSERT INTO t1 VALUES (1, 'ab'), (2, 'cde'), (3, 'f');",
            ],
            skip: Some(r#"setup fails on Postgres ("error running setup query: CREATE EXTENSION doltgres_test;: ERROR: extension \"doltgres_test\" is not available (SQLSTATE 0A000)") and on the Go server ("error running setup query: CREATE EXTENSION doltgres_test;: ERROR: extension \"doltgres_test\" is not available (SQLSTATE XX000)")"#),
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT dgtest_charcount(v1) FROM t1;",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dgtest_charcount(v1) FROM t1 WHERE pk = 2;",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dgtest_charcount(v1) FROM t1 WHERE pk = 0;",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT aggkind, agginitval, aggcombinefn::text, aggfinalfn::text FROM pg_aggregate WHERE aggtransfn::text = 'dgtest_charcount_transition';",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Objects are created in the extension's target schema",
            set_up_script: &[
                "CREATE EXTENSION doltgres_test;",
            ],
            skip: Some(r#"setup fails on Postgres ("error running setup query: CREATE EXTENSION doltgres_test;: ERROR: extension \"doltgres_test\" is not available (SQLSTATE 0A000)") and on the Go server ("error running setup query: CREATE EXTENSION doltgres_test;: ERROR: extension \"doltgres_test\" is not available (SQLSTATE XX000)")"#),
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT n.nspname FROM pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE t.typname = 'dgtest_upper';",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT public.dgtest_upper_text('abc'::public.dgtest_upper);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Extension objects are not available before the extension is created",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'abc'::dgtest_upper;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "dgtest_upper" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dgtest_charcount('abc');",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function dgtest_charcount(unknown) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                // Doltgres extension: Postgres also names the missing control file, which Doltgres does not have.
                ScriptTestAssertion {
                    query: "CREATE EXTENSION doltgres_test;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"extension "doltgres_test" is not available"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dgtest_charcount('abc');",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function dgtest_charcount(unknown) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
