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
fn test_alter_statements() {
    run_scripts(&[
        ScriptTest {
            name: "alter database",
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER DATABASE postgres OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter sequence",
            set_up_script: &[
                "CREATE SEQUENCE testseq",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE testseq OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter type",
            set_up_script: &[
                "CREATE TYPE testtype AS ENUM ('a', 'b', 'c')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TYPE testtype OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter function",
            set_up_script: &[
                "CREATE FUNCTION testfunc() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER FUNCTION testfunc() OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter procedure",
            set_up_script: &[
                "CREATE TABLE test (v1 INT8);",
                "CREATE PROCEDURE testproc() AS $$ BEGIN INSERT INTO test VALUES (1); END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER PROCEDURE testproc() OWNER TO foo;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter schema",
            set_up_script: &[
                "CREATE SCHEMA testschema",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER schema testschema OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter view",
            set_up_script: &[
                "CREATE VIEW testview AS SELECT 1",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER VIEW testview OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
