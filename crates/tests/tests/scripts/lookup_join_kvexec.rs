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
fn test_lookup_join_extended_key_storage_position() {
    run_scripts(&[
        ScriptTest {
            name: "join key is the second non-PK column",
            set_up_script: &[
                "CREATE TABLE c (id uuid PRIMARY KEY, name text);",
                "CREATE TABLE m (seq int PRIMARY KEY, filler text, company_id uuid, label text);",
                "INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'acme');",
                "INSERT INTO m VALUES (1, 'f', '11111111-1111-1111-1111-111111111111', 'L');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select /*+ lookup_join(m, c) */ HINT count(*) from m join c on c.id = m.company_id;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "join key is the first non-PK column",
            set_up_script: &[
                "CREATE TABLE c (id uuid PRIMARY KEY, name text);",
                "CREATE TABLE m (seq int PRIMARY KEY, company_id uuid, label text);",
                "INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'acme');",
                "INSERT INTO m VALUES (1, '11111111-1111-1111-1111-111111111111', 'L');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select /*+ lookup_join(m, c) */ HINT count(*) from m join c on c.id = m.company_id;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "join key is the third non-PK column",
            set_up_script: &[
                "CREATE TABLE c (id uuid PRIMARY KEY, name text);",
                "CREATE TABLE m (seq int PRIMARY KEY, filler1 text, filler2 text, company_id uuid, label text);",
                "INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'acme');",
                "INSERT INTO m VALUES (1, 'f1', 'f2', '11111111-1111-1111-1111-111111111111', 'L');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select /*+ lookup_join(m, c) */ HINT count(*) from m join c on c.id = m.company_id;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "composite key with both columns at non-zero storage positions",
            set_up_script: &[
                "CREATE TABLE c (id uuid PRIMARY KEY, tag text, code uuid, label text);",
                "CREATE TABLE m (seq int PRIMARY KEY, filler text, ref_code uuid, ref_label text);",
                "INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'c1', '22222222-2222-2222-2222-222222222222', 'match');",
                "INSERT INTO c VALUES ('33333333-3333-3333-3333-333333333333', 'c2', '22222222-2222-2222-2222-222222222222', 'other');",
                "INSERT INTO m VALUES (1, 'f', '22222222-2222-2222-2222-222222222222', 'match');",
                "INSERT INTO m VALUES (2, 'f', '22222222-2222-2222-2222-222222222222', 'other');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select /*+ lookup_join(m, c) */ HINT m.seq, c.tag from m join c on c.code = m.ref_code and c.label = m.ref_label order by m.seq;",
                    expected: Expected::Rows {
                        columns: &[Column("seq", INT4), Column("tag", TEXT)],
                        rows: &[
                            &[T("1"), T("c1")],
                            &[T("2"), T("c2")],
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
