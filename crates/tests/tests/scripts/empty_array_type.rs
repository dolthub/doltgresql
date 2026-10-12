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
fn test_empty_array_type() {
    run_scripts(&[
        ScriptTest {
            name: "empty array element types",
            set_up_script: &[
                "CREATE TABLE empty_inputs (id int PRIMARY KEY,a int[]);",
                "INSERT INTO empty_inputs VALUES (1,ARRAY[]::int[]),(2,ARRAY[1]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[];",
                    expected: Expected::Error(Diagnostic { code: "42P18", message: "cannot determine type of empty array", hint: "Explicitly cast to the desired type, for example ARRAY[]::integer[].", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(ARRAY[]);",
                    expected: Expected::Error(Diagnostic { code: "42P18", message: "cannot determine type of empty array", hint: "Explicitly cast to the desired type, for example ARRAY[]::integer[].", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[]::int[],pg_typeof(ARRAY[]::int[]),ARRAY[ARRAY[]]::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY), Column("pg_typeof", REGTYPE), Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{}"), T("integer[]"), T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[ARRAY[]::int[]];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT ARRAY[]::text[]),ARRAY[ARRAY[]::text[],ARRAY[]::text[]];",
                    expected: Expected::Rows {
                        columns: &[Column("array", TEXT_ARRAY), Column("array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{}"), T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id,ARRAY[a] FROM empty_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{}")],
                            &[T("2"), T("{{1}}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(ARRAY[[NULL,NULL],[1,NULL]]::int[]),ARRAY[[NULL,NULL],[1,NULL]]::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("integer[]"), T("{{NULL,NULL},{1,NULL}}")],
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
