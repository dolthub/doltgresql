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
fn test_array_dimension_limit() {
    run_scripts(&[
        ScriptTest {
            name: "array dimension limits",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_ndims(ARRAY[[[[[[1]]]]]]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_ndims", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[[[[[[[1]]]]]]];",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "number of array dimensions (7) exceeds the maximum allowed (6)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{{{{{{{1}}}}}}}'::int[];",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "number of array dimensions (7) exceeds the maximum allowed (6)", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_ndims(ARRAY[(SELECT ARRAY[[[[[1]]]]])]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_ndims", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[(SELECT ARRAY[[[[[[1]]]]]])];",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "number of array dimensions (7) exceeds the maximum allowed (6)", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
