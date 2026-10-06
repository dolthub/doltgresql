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
fn test_scalar_subscripts() {
    run_scripts(&[
        ScriptTest {
            name: "reject scalar subscripts",
            set_up_script: &[
                "CREATE TABLE t_scalar (n int);",
                "INSERT INTO t_scalar VALUES (42);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT (42)[1:2];",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "cannot subscript type integer because it does not support subscripting", position: 9, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (NULL::int)[1:2];",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "cannot subscript type integer because it does not support subscripting", position: 9, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t_scalar SET n[1]=7;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "cannot subscript type integer because it does not support subscripting", position: 21, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n[1:2] FROM t_scalar;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "cannot subscript type integer because it does not support subscripting", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ((SELECT n FROM t_scalar))[1:2];",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "cannot subscript type integer because it does not support subscripting", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
