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
fn test_array_subscript_update() {
    run_scripts(&[
        ScriptTest {
            name: "array subscript updates",
            set_up_script: &[
                "CREATE TABLE t (id int PRIMARY KEY, a int[]);",
                "INSERT INTO t VALUES (1,ARRAY[[1,2,3],[4,5,6]]),(2,ARRAY[1,2,3]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t SET a[2][3]=60 WHERE id=1 RETURNING a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2,3},{4,5,60}}")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[1:2][2:3]=ARRAY[[20,30],[50,60]] WHERE id=1 RETURNING a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,20,30},{4,50,60}}")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[6]=99 WHERE id=2 RETURNING a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2,3,NULL,NULL,99}")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[1][1]=7 WHERE id=3 RETURNING a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{{7}}")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[:][2:3]=ARRAY[[2,3],[5,6]] WHERE id=1 RETURNING a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2,3},{4,5,6}}")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[3][1]=7 WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "array subscript out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[1:2][1:2]=ARRAY[[9,8]] WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "source array too small", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[NULL]=7 WHERE id=2;",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "array subscript in assignment must not be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a FROM t WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2,3},{4,5,6}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[NULL:2]=NULL WHERE id=1;",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "array subscript in assignment must not be null", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array assignments from row expressions and subqueries",
            set_up_script: &[
                "CREATE TABLE t (id int PRIMARY KEY,a int[]);",
                "INSERT INTO t VALUES (1,ARRAY[[1,2,3],[4,5,6]]),(2,ARRAY[1,2,3,NULL,NULL,99]),(3,ARRAY[[7]]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t SET a[id]=(SELECT 8) WHERE id=2 RETURNING a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,8,3,NULL,NULL,99}")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[(SELECT 1):(SELECT 2)]=(SELECT ARRAY[9,10]) WHERE id=2 RETURNING a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{9,10,3,NULL,NULL,99}")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[1:2]=NULL WHERE id=2 RETURNING a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{9,10,3,NULL,NULL,99}")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[2:1]=NULL WHERE id=2 RETURNING a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{9,10,3,NULL,NULL,99}")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[1:1][1:1]=ARRAY[NULL,99]::int[] WHERE id=1 RETURNING a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{{NULL,2,3},{4,5,6}}")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a FROM t WHERE id=3;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{{7}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[2:1]=ARRAY[1] WHERE id=2;",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "upper bound cannot be less than lower bound", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[id]=(SELECT 8) WHERE id=2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a FROM t WHERE id=2;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{9,8,3,NULL,NULL,99}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "empty array assignments require explicit bounds",
            set_up_script: &[
                "CREATE TABLE t (a int[]);",
                "INSERT INTO t VALUES (ARRAY[]::int[]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t SET a[:]=ARRAY[1];",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "array slice subscript must provide both boundaries", detail: "When assigning to a slice of an empty array value, slice boundaries must be fully specified.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a[1:1]=NULL;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{}")],
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
