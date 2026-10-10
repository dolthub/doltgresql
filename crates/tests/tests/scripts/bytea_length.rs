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
fn test_bytea_length_functions() {
    run_scripts(&[
        ScriptTest {
            name: "octet_length, length and bit_length for bytea",
            set_up_script: &[
                "CREATE TABLE byteas (id int primary key, v bytea);",
                r#"INSERT INTO byteas VALUES (1, '\x00010203'), (2, ''), (3, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT octet_length('x'::bytea), length('x'::bytea), bit_length('x'::bytea);",
                    expected: Expected::Rows {
                        columns: &[Column("octet_length", INT4), Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT octet_length('\x00010203'::bytea), length('\x00010203'::bytea), bit_length('\x00010203'::bytea);"#,
                    expected: Expected::Rows {
                        columns: &[Column("octet_length", INT4), Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[T("4"), T("4"), T("32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length('héllo'::bytea), length('héllo'::text), octet_length('héllo'::text);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("length", INT4), Column("octet_length", INT4)],
                        rows: &[
                            &[T("6"), T("5"), T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT octet_length(''::bytea), length(''::bytea), bit_length(''::bytea);",
                    expected: Expected::Rows {
                        columns: &[Column("octet_length", INT4), Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[T("0"), T("0"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT octet_length(NULL::bytea), length(NULL::bytea), bit_length(NULL::bytea);",
                    expected: Expected::Rows {
                        columns: &[Column("octet_length", INT4), Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[Null, Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT octet_length('abc'), length('abc'), bit_length('abc');",
                    expected: Expected::Rows {
                        columns: &[Column("octet_length", INT4), Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[T("3"), T("3"), T("24")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, octet_length(v), length(v), bit_length(v) FROM byteas ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("octet_length", INT4), Column("length", INT4), Column("bit_length", INT4)],
                        rows: &[
                            &[T("1"), T("4"), T("4"), T("32")],
                            &[T("2"), T("0"), T("0"), T("0")],
                            &[T("3"), Null, Null, Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "octet_length in a bytea CHECK constraint",
            set_up_script: &[
                "CREATE TABLE artifacts (id int primary key, artifact bytea NOT NULL, CONSTRAINT artifact_not_empty CHECK (octet_length(artifact) > 0));",
                r#"INSERT INTO artifacts VALUES (1, '\x0001');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, octet_length(artifact) FROM artifacts;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("octet_length", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO artifacts VALUES (2, '');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "artifacts" violates check constraint "artifact_not_empty""#, detail: r#"Failing row contains (2, \x)."#, schema: "public", table: "artifacts", constraint: "artifact_not_empty", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
