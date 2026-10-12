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
fn test_partial_index_merge_join_does_not_lose_rows() {
    run_scripts(&[
        ScriptTest {
            name: "partial index on join key",
            set_up_script: &[
                "CREATE TABLE profiles (id int PRIMARY KEY);",
                "CREATE TABLE favorites (id int PRIMARY KEY, profile_id int NOT NULL, flag boolean NOT NULL DEFAULT false);",
                "CREATE TABLE profile_ranges (id int PRIMARY KEY, min_id int NOT NULL, max_id int NOT NULL);",
                "CREATE INDEX favorites_partial ON favorites (profile_id) WHERE flag;",
                "CREATE INDEX profile_ranges_min ON profile_ranges (min_id);",
                "CREATE INDEX profile_ranges_max ON profile_ranges (max_id);",
                "INSERT INTO profiles VALUES (1), (2), (3), (4), (5), (6), (7);",
                "INSERT INTO favorites (id, profile_id) SELECT id, id FROM profiles;",
                "UPDATE favorites SET flag = true WHERE profile_id IN (1, 7);",
                "INSERT INTO profile_ranges VALUES (1, 1, 7);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM favorites f JOIN profiles p ON p.id = f.profile_id;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM favorites f LEFT JOIN profiles p ON p.id = f.profile_id;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), count(f.id) FROM profiles p LEFT JOIN favorites f ON f.profile_id = p.id;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8)],
                        rows: &[
                            &[T("7"), T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM profiles p WHERE EXISTS (SELECT 1 FROM favorites f WHERE f.profile_id = p.id);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM profiles p WHERE NOT EXISTS (SELECT 1 FROM favorites f WHERE f.profile_id = p.id);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM favorites f JOIN profile_ranges r ON f.profile_id BETWEEN r.min_id AND r.max_id;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM favorites f JOIN profiles p ON p.id = f.profile_id WHERE f.flag;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "partial index created after rows exist",
            set_up_script: &[
                "CREATE TABLE existing_profiles (id int PRIMARY KEY);",
                "CREATE TABLE existing_favorites (id int PRIMARY KEY, profile_id int NOT NULL, flag boolean NOT NULL DEFAULT false);",
                "INSERT INTO existing_profiles VALUES (1), (2), (3), (4), (5), (6), (7);",
                "INSERT INTO existing_favorites (id, profile_id) SELECT id, id FROM existing_profiles;",
                "UPDATE existing_favorites SET flag = true WHERE profile_id IN (1, 7);",
                "CREATE INDEX existing_favorites_partial ON existing_favorites (profile_id) WHERE flag;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM existing_favorites f JOIN existing_profiles p ON p.id = f.profile_id;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM existing_favorites f JOIN existing_profiles p ON p.id = f.profile_id WHERE f.flag;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
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
