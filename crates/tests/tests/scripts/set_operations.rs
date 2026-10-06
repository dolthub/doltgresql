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
fn test_set_operations() {
    run_scripts(&[
        ScriptTest {
            name: "Test intersect",
            set_up_script: &[
                "create table b (m int, n int);",
                "insert into b values (1,2), (1,3), (3,4);",
                "create table c (m int, n int);",
                "insert into c values (1,3), (1,3), (3,4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from b intersect select * from c order by 1,2;",
                    expected: Expected::Rows {
                        columns: &[Column("m", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("3"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "(table b order by m limit 1 offset 1) intersect (table c order by m limit 1);",
                    expected: Expected::Rows {
                        columns: &[Column("m", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Test union",
            set_up_script: &[
                "create table b (m int, n int);",
                "insert into b values (1,2), (1,3), (3,4);",
                "create table c (m int, n int);",
                "insert into c values (1,3), (1,3), (3,4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from b union select * from c order by 1,2;",
                    expected: Expected::Rows {
                        columns: &[Column("m", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("1"), T("3")],
                            &[T("3"), T("4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from b union all select * from c order by 1,2;",
                    expected: Expected::Rows {
                        columns: &[Column("m", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("1"), T("3")],
                            &[T("1"), T("3")],
                            &[T("1"), T("3")],
                            &[T("3"), T("4")],
                            &[T("3"), T("4")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "(table b order by m limit 1 offset 1) union (table c order by m limit 1);",
                    expected: Expected::Rows {
                        columns: &[Column("m", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "(table b order by m limit 1 offset 1) union all (table c order by m limit 1);",
                    expected: Expected::Rows {
                        columns: &[Column("m", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("1"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Test except",
            set_up_script: &[
                "create table b (m int, n int);",
                "insert into b values (1,2), (1,3), (3,4);",
                "create table c (m int, n int);",
                "insert into c values (1,3), (1,3), (3,4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from b except select * from c order by 1,2;",
                    expected: Expected::Rows {
                        columns: &[Column("m", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "(table b order by m limit 1 offset 1) except (table c order by m limit 1);",
                    expected: Expected::Rows {
                        columns: &[Column("m", INT4), Column("n", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
