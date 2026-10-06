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
fn test_create_extension() {
    run_scripts(&[
        ScriptTest {
            name: "create extension uuid-ossp after setting search_path to empty",
            set_up_script: &[
                "SELECT pg_catalog.set_config('search_path', '', false);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;"#,
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_nil();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function uuid_nil() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT public.uuid_nil();",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_nil", UUID)],
                        rows: &[
                            &[T("00000000-0000-0000-0000-000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter table with default expr using extension function when search_path is empty",
            set_up_script: &[
                "SELECT pg_catalog.set_config('search_path', '', false);",
                r#"CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;"#,
                r#"CREATE TABLE public.goals (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    note_id uuid,
    completion_timestamp timestamp without time zone,
    due_date timestamp without time zone
);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE ONLY public.goals ADD CONSTRAINT goals_pkey PRIMARY KEY (id);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "uuid-ossp is not available before it is created",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pg_catalog.pg_extension;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, installed_version FROM pg_catalog.pg_available_extensions WHERE name = 'uuid-ossp';",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("installed_version", TEXT)],
                        rows: &[
                            &[T("uuid-ossp"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_nil();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function uuid_nil() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_nil();",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_nil", UUID)],
                        rows: &[
                            &[T("00000000-0000-0000-0000-000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "only emulated extensions may be created",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "doltgres_no_such_extension";"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"extension "doltgres_no_such_extension" is not available"#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION IF NOT EXISTS "doltgres_no_such_extension";"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"extension "doltgres_no_such_extension" is not available"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "UUID-OSSP";"#,
                    expected: Expected::Tag("CREATE EXTENSION"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "42723", message: r#"function "uuid_nil" already exists with same argument types"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "42723", message: r#"function "uuid_nil" already exists with same argument types"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION IF NOT EXISTS "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "42723", message: r#"function "uuid_nil" already exists with same argument types"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_nil();",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_nil", UUID)],
                        rows: &[
                            &[T("00000000-0000-0000-0000-000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "uuid-ossp options that are not yet supported",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp" VERSION oldversion;"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"extension "uuid-ossp" has no installation script nor update path for version "oldversion""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp" WITH SCHEMA myschema;"#,
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "myschema" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp" CASCADE;"#,
                    expected: Expected::Tag("CREATE EXTENSION"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "uuid-ossp installation participates in branches",
            set_up_script: &[
                "SELECT dolt_commit('--allow-empty', '-m', 'initial commit');",
                "SELECT dolt_checkout('-b', 'ext');",
                r#"CREATE EXTENSION "uuid-ossp";"#,
                "SELECT dolt_commit('-Am', 'create the extension');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT extname FROM pg_catalog.pg_extension;",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME)],
                        rows: &[
                            &[T("uuid-ossp")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('main');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT extname FROM pg_catalog.pg_extension;",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT uuid_nil();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function: 'uuid_nil' not found", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_merge('ext');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT extname, extversion FROM pg_catalog.pg_extension;",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME), Column("extversion", TEXT)],
                        rows: &[
                            &[T("uuid-ossp"), T("1.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT uuid_nil();",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_nil", UUID)],
                        rows: &[
                            &[T("00000000-0000-0000-0000-000000000000")],
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
