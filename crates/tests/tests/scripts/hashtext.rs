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
fn test_hash_text() {
    run_scripts(&[
        ScriptTest {
            name: "hashtext",
            set_up_script: &[
                "CREATE TABLE hashtext_values (id INT PRIMARY KEY, value TEXT);",
                "INSERT INTO hashtext_values VALUES (1, repeat('x', 20000));",
                "CREATE TABLE hashtext_lengths (length INT PRIMARY KEY);",
                "INSERT INTO hashtext_lengths VALUES (0), (1), (2), (3), (4), (5), (6), (7), (8), (9), (10), (11), (12), (13), (23), (24), (25);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT length, hashtext(repeat('x', length)) FROM hashtext_lengths ORDER BY length;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("hashtext", INT4)],
                        rows: &[
                            &[T("0"), T("-1477818771")],
                            &[T("1"), T("1074944137")],
                            &[T("2"), T("-1086392228")],
                            &[T("3"), T("-1992236649")],
                            &[T("4"), T("-1379736791")],
                            &[T("5"), T("-370454118")],
                            &[T("6"), T("1489915569")],
                            &[T("7"), T("-66683019")],
                            &[T("8"), T("-2126973000")],
                            &[T("9"), T("1651296771")],
                            &[T("10"), T("755764456")],
                            &[T("11"), T("-1494243903")],
                            &[T("12"), T("631527812")],
                            &[T("13"), T("28686851")],
                            &[T("23"), T("597544042")],
                            &[T("24"), T("1380215333")],
                            &[T("25"), T("733930510")],
                        ],
                        tag: "SELECT 17",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hashtext('');",
                    expected: Expected::Rows {
                        columns: &[Column("hashtext", INT4)],
                        rows: &[
                            &[T("-1477818771")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hashtext('abc');",
                    expected: Expected::Rows {
                        columns: &[Column("hashtext", INT4)],
                        rows: &[
                            &[T("-785388649")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hashtext('12345678901'), hashtext('123456789012'), hashtext('1234567890123');",
                    expected: Expected::Rows {
                        columns: &[Column("hashtext", INT4), Column("hashtext", INT4), Column("hashtext", INT4)],
                        rows: &[
                            &[T("1650060602"), T("-2102057603"), T("437480032")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT octet_length('12345678901é'), hashtext('12345678901é');",
                    expected: Expected::Rows {
                        columns: &[Column("octet_length", INT4), Column("hashtext", INT4)],
                        rows: &[
                            &[T("13"), T("-295778157")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hashtext('café'), hashtext('日本');",
                    expected: Expected::Rows {
                        columns: &[Column("hashtext", INT4), Column("hashtext", INT4)],
                        rows: &[
                            &[T("103771354"), T("-1851216170")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hashtext(repeat('x', 1000));",
                    expected: Expected::Rows {
                        columns: &[Column("hashtext", INT4)],
                        rows: &[
                            &[T("-1157355676")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hashtext(value) FROM hashtext_values WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("hashtext", INT4)],
                        rows: &[
                            &[T("-1519670832")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hashtext(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("hashtext", INT4)],
                        rows: &[
                            &[Null],
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
