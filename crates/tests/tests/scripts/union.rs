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
fn test_except() {
    run_scripts(&[
        ScriptTest {
            name: "except tests",
            set_up_script: &[
                "CREATE TABLE t1 (i INT PRIMARY KEY);",
                "CREATE TABLE t2 (j INT PRIMARY KEY);",
                "INSERT INTO t1 VALUES (1), (2), (3);",
                "INSERT INTO t2 VALUES (2), (3), (4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t1 EXCEPT SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 123 EXCEPT SELECT 456;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("123")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES (123), (456)) a EXCEPT SELECT * FROM (VALUES (456), (789)) b;",
                    expected: Expected::Rows {
                        columns: &[Column("column1", INT4)],
                        rows: &[
                            &[T("123")],
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

#[test]
fn test_intersect() {
    run_scripts(&[
        ScriptTest {
            name: "intersect tests",
            set_up_script: &[
                "CREATE TABLE t1 (i INT PRIMARY KEY);",
                "CREATE TABLE t2 (j INT PRIMARY KEY);",
                "INSERT INTO t1 VALUES (1), (2), (3);",
                "INSERT INTO t2 VALUES (2), (3), (4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t1 INTERSECT SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 123 INTERSECT SELECT 456;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES (123), (456)) a INTERSECT SELECT * FROM (VALUES (456), (789)) b;",
                    expected: Expected::Rows {
                        columns: &[Column("column1", INT4)],
                        rows: &[
                            &[T("456")],
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

#[test]
fn test_union() {
    run_scripts(&[
        ScriptTest {
            name: "union tests",
            set_up_script: &[
                "CREATE TABLE t1 (i INT PRIMARY KEY);",
                "CREATE TABLE t2 (j INT PRIMARY KEY);",
                "INSERT INTO t1 VALUES (1), (2), (3);",
                "INSERT INTO t2 VALUES (2), (3), (4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t1 UNION SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("4")],
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 123 UNION SELECT 456;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("123")],
                            &[T("456")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES (123), (456)) a UNION SELECT * FROM (VALUES (456), (789)) b;",
                    expected: Expected::Rows {
                        columns: &[Column("column1", INT4)],
                        rows: &[
                            &[T("456")],
                            &[T("123")],
                            &[T("789")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
