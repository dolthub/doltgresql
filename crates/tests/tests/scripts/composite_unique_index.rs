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
fn test_composite_unique_index_extended_and_adaptive() {
    run_scripts(&[
        ScriptTest {
            name: "keyless: UNIQUE(uuid, text) — insert into empty table",
            set_up_script: &[
                "CREATE TABLE t (iid uuid, slug text, UNIQUE(iid, slug));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('22222222-2222-2222-2222-222222222222', 'hello');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'world');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t_iid_slug_key""#, detail: "Key (iid, slug)=(11111111-1111-1111-1111-111111111111, hello) already exists.", schema: "public", table: "t", constraint: "t_iid_slug_key", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "keyless: UNIQUE(text, uuid) — column order reversed",
            set_up_script: &[
                "CREATE TABLE t (slug text, iid uuid, UNIQUE(slug, iid));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('hello', '11111111-1111-1111-1111-111111111111');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('hello', '11111111-1111-1111-1111-111111111111');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t_slug_iid_key""#, detail: "Key (slug, iid)=(hello, 11111111-1111-1111-1111-111111111111) already exists.", schema: "public", table: "t", constraint: "t_slug_iid_key", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('world', '11111111-1111-1111-1111-111111111111');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "keyless: UNIQUE(uuid, varchar) — unbounded varchar",
            set_up_script: &[
                "CREATE TABLE t (iid uuid, slug varchar, UNIQUE(iid, slug));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t_iid_slug_key""#, detail: "Key (iid, slug)=(11111111-1111-1111-1111-111111111111, hello) already exists.", schema: "public", table: "t", constraint: "t_iid_slug_key", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "keyless: UNIQUE(uuid, text) — NULL uuid skips unique check",
            set_up_script: &[
                "CREATE TABLE t (iid uuid, slug text, UNIQUE(iid, slug));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (NULL, 'hello');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (NULL, 'hello');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "keyed: UNIQUE(uuid, text) on table with primary key",
            set_up_script: &[
                "CREATE TABLE t (pk int primary key, iid uuid, slug text, UNIQUE(iid, slug));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (2, '11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t_iid_slug_key""#, detail: "Key (iid, slug)=(11111111-1111-1111-1111-111111111111, hello) already exists.", schema: "public", table: "t", constraint: "t_iid_slug_key", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (3, '11111111-1111-1111-1111-111111111111', 'world');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "keyed: UNIQUE(uuid, varchar) — unbounded varchar on table with primary key",
            set_up_script: &[
                "CREATE TABLE t (pk int primary key, iid uuid, slug varchar, UNIQUE(iid, slug));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (2, '11111111-1111-1111-1111-111111111111', 'hello');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t_iid_slug_key""#, detail: "Key (iid, slug)=(11111111-1111-1111-1111-111111111111, hello) already exists.", schema: "public", table: "t", constraint: "t_iid_slug_key", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
