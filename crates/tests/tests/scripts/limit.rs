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
fn test_limit_offset() {
    run_scripts(&[
        ScriptTest {
            name: "basic limit tests",
            set_up_script: &[
                "CREATE TABLE t (i INT PRIMARY KEY, c int)",
                "INSERT INTO t VALUES (1, 1), (2, 2), (3, 3), (4, 4)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t LIMIT 2",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t LIMIT $1",
                    bind_vars: &[BindVar::Int64(2)],
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t LIMIT 2 OFFSET 2",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t order by c asc LIMIT 2 OFFSET 2",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
