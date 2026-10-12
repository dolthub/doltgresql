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
fn test_delete() {
    run_scripts(&[
        ScriptTest {
            name: "simple delete",
            set_up_script: &[
                "CREATE TABLE t123 (id int primary key, c1 varchar(100));",
                "INSERT INTO t123 VALUES (1, 'one'), (2, 'two');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DELETE FROM t123 where id = 1;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t123;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c1", VARCHAR)],
                        rows: &[
                            &[T("2"), T("two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "delete returning",
            set_up_script: &[
                "CREATE TABLE t123 (id int primary key, c1 varchar(100));",
                "INSERT INTO t123 VALUES (1, 'one'), (2, 'two'), (3, 'three');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DELETE FROM t123 where id = 1 RETURNING id, c1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c1", VARCHAR)],
                        rows: &[
                            &[T("1"), T("one")],
                        ],
                        tag: "DELETE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM t123 RETURNING *;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c1", VARCHAR)],
                        rows: &[
                            &[T("2"), T("two")],
                            &[T("3"), T("three")],
                        ],
                        tag: "DELETE 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t123;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c1", VARCHAR)],
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
