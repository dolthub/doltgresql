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
fn test_array_regtype() {
    run_scripts(&[
        ScriptTest {
            name: "array dimensionality does not change regtype",
            set_up_script: &[
                "CREATE TABLE array_type_names (id int PRIMARY KEY,name text);",
                "INSERT INTO array_type_names VALUES (1,'integer[][]'),(2,'varchar[][][]');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'integer[]'::regtype='integer[][]'::regtype,'varchar[][][]'::regtype='varchar[]'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'pg_catalog.int4[][]'::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("regtype", REGTYPE)],
                        rows: &[
                            &[T("integer[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name::regtype FROM array_type_names ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("name", REGTYPE)],
                        rows: &[
                            &[T("integer[]")],
                            &[T("character varying[]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT 'pg_catalog.int4[][]')::regtype;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", REGTYPE)],
                        rows: &[
                            &[T("integer[]")],
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
