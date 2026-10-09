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
fn test_array_slices() {
    run_scripts(&[
        ScriptTest {
            name: "array slices",
            set_up_script: &[
                "CREATE TABLE slice_inputs (id int PRIMARY KEY,a int[],lo int,hi int);",
                "INSERT INTO slice_inputs VALUES (1,ARRAY[1,2,3],2,9),(2,ARRAY[]::int[],1,2),(3,NULL,1,2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[[1,2,3],[4,5,6]])[2:2][2:3], (ARRAY[[1,2,3],[4,5,6]])[:][2:2];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY), Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{{5,6}}"), T("{{2},{5}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[[1,2,3],[4,5,6]])[2][2:3], (ARRAY[[1,2,3],[4,5,6]])[2:9][2:9];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY), Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{{2,3},{5,6}}"), T("{{5,6}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[[1,2],[3,4]])[9:10][:], (ARRAY[1,2,3])[NULL:2], (NULL::int[])[1:2];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY), Column("array", INT4_ARRAY), Column("int4", INT4_ARRAY)],
                        rows: &[
                            &[T("{}"), Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[[[1,2]],[[3,4]]])[2:2][:][:], array_ndims((ARRAY[[1,2],[3,4]])[2:2]);",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY), Column("array_ndims", INT4)],
                        rows: &[
                            &[T("{{{3,4}}}"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[1,2,3])[-2:2], (ARRAY[1,2,3])[3:1], (ARRAY[]::int[])[:];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY), Column("array", INT4_ARRAY), Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2}"), T("{}"), T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id,a[lo:hi],a[lo:NULL] FROM slice_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4_ARRAY), Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{2,3}"), Null],
                            &[T("2"), T("{}"), Null],
                            &[T("3"), Null, Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ((SELECT a FROM slice_inputs WHERE id=1))[(SELECT 2):(SELECT 9)];",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[1,2])[1:2][1:1],(ARRAY[[1,2],[3,4]])[1:1];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY), Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{}"), T("{{1,2}}")],
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
