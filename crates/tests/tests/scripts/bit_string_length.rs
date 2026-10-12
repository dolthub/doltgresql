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
fn test_bit_string_length_functions() {
    run_scripts(&[
        ScriptTest {
            name: "length and bit_length for bit strings",
            set_up_script: &[
                "CREATE TABLE bit_lengths (v varbit, b bit(5));",
                "INSERT INTO bit_lengths VALUES (B'101', B'00101'), (B'', B'00000'), (NULL, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT length('101'::varbit), bit_length('101'::varbit);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[T("3"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(B'10101'::bit(5)), bit_length(B'10101'::bit(5));",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[T("5"), T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(''::varbit), bit_length(''::varbit);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[T("0"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(NULL::varbit), bit_length(NULL::varbit);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(v), bit_length(v), length(b), bit_length(b) FROM bit_lengths WHERE v IS NOT NULL ORDER BY v;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("bit_length", INT4), Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[T("0"), T("0"), T("5"), T("5")],
                            &[T("3"), T("3"), T("5"), T("5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(v), bit_length(v), length(b), bit_length(b) FROM bit_lengths WHERE v IS NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("bit_length", INT4), Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[Null, Null, Null, Null],
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
