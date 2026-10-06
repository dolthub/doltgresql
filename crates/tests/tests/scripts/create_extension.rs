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
                // Doltgres extension: Postgres also names the missing control file, which Doltgres does not have.
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "doltgres_no_such_extension";"#,
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"extension "doltgres_no_such_extension" is not available"#, ..E }),
                    ..A
                },
                // Doltgres extension: Postgres also names the missing control file, which Doltgres does not have.
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION IF NOT EXISTS "doltgres_no_such_extension";"#,
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"extension "doltgres_no_such_extension" is not available"#, ..E }),
                    ..A
                },
                // Changed from the recording: a case-insensitive file system let Postgres open uuid-ossp's control file for
                // "UUID-OSSP", which extension names otherwise never match.
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "UUID-OSSP";"#,
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"extension "UUID-OSSP" is not available"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("CREATE EXTENSION"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"extension "uuid-ossp" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION IF NOT EXISTS "uuid-ossp";"#,
                    expected: Expected::Tag("CREATE EXTENSION"),
                    notices: &[Diagnostic { code: "42710", message: r#"extension "uuid-ossp" already exists, skipping"#, ..N }],
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
                // Doltgres extension: dolt_checkout and dolt_merge change which extensions exist, and Postgres always lists plpgsql.
                ScriptTestAssertion {
                    query: "SELECT extname FROM pg_catalog.pg_extension;",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME)],
                        rows: &[
                            &[T("plpgsql")],
                            &[T("uuid-ossp")],
                        ],
                        tag: "SELECT 2",
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
                // Doltgres extension: dolt_checkout and dolt_merge change which extensions exist, and Postgres always lists plpgsql.
                ScriptTestAssertion {
                    query: "SELECT extname FROM pg_catalog.pg_extension;",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME)],
                        rows: &[
                            &[T("plpgsql")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres extension: dolt_checkout changed which functions exist, and Postgres reports the missing function.
                ScriptTestAssertion {
                    query: "SELECT uuid_nil();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function uuid_nil() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_merge('ext');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres extension: dolt_checkout and dolt_merge change which extensions exist, and Postgres always lists plpgsql.
                ScriptTestAssertion {
                    query: "SELECT extname, extversion FROM pg_catalog.pg_extension;",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME), Column("extversion", TEXT)],
                        rows: &[
                            &[T("plpgsql"), T("1.0")],
                            &[T("uuid-ossp"), T("1.1")],
                        ],
                        tag: "SELECT 2",
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

#[test]
fn test_extension_rules() {
    run_scripts(&[
        ScriptTest {
            name: "uuid-ossp functions",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_nil(), uuid_ns_dns(), uuid_ns_url(), uuid_ns_oid(), uuid_ns_x500();",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_nil", UUID), Column("uuid_ns_dns", UUID), Column("uuid_ns_url", UUID), Column("uuid_ns_oid", UUID), Column("uuid_ns_x500", UUID)],
                        rows: &[
                            &[T("00000000-0000-0000-0000-000000000000"), T("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), T("6ba7b811-9dad-11d1-80b4-00c04fd430c8"), T("6ba7b812-9dad-11d1-80b4-00c04fd430c8"), T("6ba7b814-9dad-11d1-80b4-00c04fd430c8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3(uuid_ns_dns(), 'www.example.com'), uuid_generate_v5(uuid_ns_url(), 'http://www.postgresql.org');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v3", UUID), Column("uuid_generate_v5", UUID)],
                        rows: &[
                            &[T("5df41881-3aed-3515-88a7-2f4a814cf09e"), T("e1ee1ad4-cd4e-5889-962a-4f605a68d94e")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring(uuid_generate_v4()::text, 15, 1), substring(uuid_generate_v1()::text, 15, 1), substring(uuid_generate_v1mc()::text, 26, 1) IN ('3', '7', 'b', 'f');",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT), Column("substring", TEXT), Column("?column?", BOOL)],
                        rows: &[
                            &[T("4"), T("1"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3(NULL, 'x') IS NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"extension "uuid-ossp" already exists"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION IF NOT EXISTS "uuid-ossp";"#,
                    expected: Expected::Tag("CREATE EXTENSION"),
                    notices: &[Diagnostic { code: "42710", message: r#"extension "uuid-ossp" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE EXTENSION "uuid-ossp" VERSION '9.9';"#,
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"extension "uuid-ossp" already exists"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres extension: Postgres also names the missing control file, which Doltgres does not have.
                ScriptTestAssertion {
                    query: "CREATE EXTENSION no_such_extension_here;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"extension "no_such_extension_here" is not available"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT extname, extversion, extrelocatable FROM pg_extension ORDER BY extname;",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME), Column("extversion", TEXT), Column("extrelocatable", BOOL)],
                        rows: &[
                            &[T("plpgsql"), T("1.0"), T("f")],
                            &[T("uuid-ossp"), T("1.1"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_nil();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function uuid_nil() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"extension "uuid-ossp" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION IF EXISTS "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    notices: &[Diagnostic { code: "00000", message: r#"extension "uuid-ossp" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "vector input and output",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE EXTENSION vector;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1000000, 0.0000001, 1.5e20, -0, 1e-45]'::vector, '[1e-50]'::vector, '[ 1 , 2 ]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED), Column("vector", USER_DEFINED), Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1e+06,1e-07,1.5e+20,-0,1e-45]"), T("[0]"), T("[1,2]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e39]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""1e39" is out of range for type vector"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[nan]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "NaN not allowed in vector", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[1,2""#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1,2'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "1,2""#, detail: r#"Vector contents must start with "["."#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2] x'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[1,2] x""#, detail: "Junk after closing right brace.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec, '[65519]'::halfvec, '[1.0001, 0.00001]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED), Column("halfvec", USER_DEFINED), Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2]"), T("[65504]"), T("[1,1.001358e-05]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[65520]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""65520" is out of range for type halfvec"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:0,3:2}/3'::sparsevec, '{}/2'::sparsevec, ' { 3 : 1 , 1:2 } / 4 '::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED), Column("sparsevec", USER_DEFINED), Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{3:2}/3"), T("{}/2"), T("{1:2,3:1}/4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1}/3 x'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{1:1}/3 x""#, detail: "Junk after closing.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{4:1}/3'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec index out of bounds", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,1:2}/3'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec indices must not contain duplicates", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1}/0'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec must have at least 1 dimension", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1e-50}/3'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""1e-50" is out of range for type sparsevec"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{}""#, detail: "Unexpected end of input.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector(3);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 3 dimensions, not 2", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector(2)::vector(3);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 3 dimensions, not 2", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector(0);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "dimensions for type vector must be at least 1", position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector(1,2);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "invalid type modifier", position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector('a');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 17, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof('[1,2]'::vector), format_type('vector'::regtype, 3), format_type('_vector'::regtype, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("format_type", TEXT), Column("format_type", TEXT)],
                        rows: &[
                            &[T("vector"), T("vector(3)"), T("vector(3)[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "vector functions and operators",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE EXTENSION vector;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_dims('[1,2,3]'::vector), vector_norm('[3,4]'), l2_normalize('[3,4]'::vector), l2_normalize('[0,0]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("vector_dims", INT4), Column("vector_norm", FLOAT8), Column("l2_normalize", USER_DEFINED), Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("3"), T("5"), T("[0.6,0.8]"), T("[0,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector <-> '[3,4]', '[1,2]'::vector <#> '[3,4]', '[1,2]'::vector <=> '[3,4]', '[1,2]'::vector <+> '[3,4]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8), Column("?column?", FLOAT8), Column("?column?", FLOAT8), Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("2.8284271247461903"), T("-11"), T("0.01613008990009257"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector + '[3,4]', '[1,2]'::vector - '[3,4]', '[1,2]'::vector * '[3,4]', '[1,2]'::vector || '[3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED), Column("?column?", USER_DEFINED), Column("?column?", USER_DEFINED), Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[4,6]"), T("[-2,-2]"), T("[3,8]"), T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector + '[3]';",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different vector dimensions 2 and 1", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e38]'::vector * '[10]';",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: overflow", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e-30]'::vector * '[1e-30]';",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: underflow", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector < '[1,2,3]', '[1,2]'::vector = '[1,2]', '[2]'::vector > '[1,9]', '{1:1}/2'::sparsevec < '{2:1}/2', '{1:-1}/2'::sparsevec < '{2:1}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4]'::vector, 2, 2), subvector('[1,2,3]'::vector, 0, 2), subvector('[1,2,3]'::vector, 3, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED), Column("subvector", USER_DEFINED), Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[2,3]"), T("[1]"), T("[3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3]'::vector, 4, 1);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT binary_quantize('[1,-1,0,2]'::vector), hamming_distance(B'101', B'011'), jaccard_distance(B'101', B'011'), jaccard_distance(B'000', B'000');",
                    expected: Expected::Rows {
                        columns: &[Column("binary_quantize", BIT), Column("hamming_distance", FLOAT8), Column("jaccard_distance", FLOAT8), Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("1001"), T("2"), T("0.6666666666666667"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance(B'101', B'01');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different bit lengths 3 and 2", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[1,2]'::halfvec, '[3,4]'), inner_product('{1:2,3:4}/3'::sparsevec, '{3:1}/3'), cosine_distance('{1:1}/2'::sparsevec, '{2:1}/2'), l2_norm('{}/2'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8), Column("inner_product", FLOAT8), Column("cosine_distance", FLOAT8), Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("2.8284271247461903"), T("4"), T("1"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[0,0]'::vector, '[1,1]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_accum('{0}', '[1,2]'), vector_avg('{2,4,6}'), vector_combine('{0,1}', '{1,2,3}');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_accum", FLOAT8_ARRAY), Column("vector_avg", USER_DEFINED), Column("vector_combine", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("{1,1,2}"), T("[2,3]"), T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine('{0}', '{0}');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v), sum(v) FROM (VALUES ('[1,2]'::vector), ('[3,5]'), (NULL)) t(v);",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED), Column("sum", USER_DEFINED)],
                        rows: &[
                            &[T("[2,3.5]"), T("[4,7]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v), sum(v) FROM (VALUES ('[1,2]'::vector)) t(v) WHERE false;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED), Column("sum", USER_DEFINED)],
                        rows: &[
                            &[Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM (VALUES ('[1,2]'::vector), ('[3]')) t(v);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 1", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT unnest('{"[1,2,3]", "[4,5,6]"}'::vector[]);"#,
                    expected: Expected::Rows {
                        columns: &[Column("unnest", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[4,5,6]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "vector casts",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE EXTENSION vector;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,2,3]::vector, '{1.5,2}'::real[]::vector, ARRAY[1,2]::numeric[]::halfvec, '{1,0,2}'::float8[]::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED), Column("vector", USER_DEFINED), Column("array", USER_DEFINED), Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]"), T("[1.5,2]"), T("[1,2]"), T("{1:1,3:2}/3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector::real[], '[1,2]'::vector::halfvec::sparsevec, '{1:1,3:2}/3'::sparsevec::vector, '[1.5,2]'::halfvec::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("float4", FLOAT4_ARRAY), Column("sparsevec", USER_DEFINED), Column("vector", USER_DEFINED), Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("{1,2}"), T("{1:1,2:2}/2"), T("[1,0,2]"), T("[1.5,2]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::real[]::vector(2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[[1,2],[3,4]]::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "array must be 1-D", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,NULL]::vector;",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "array must not contain nulls", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}'::real[]::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[65520]'::vector::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""65520" is out of range for type halfvec"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:65520}/1'::sparsevec::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""65520" is out of range for type halfvec"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector::halfvec(2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc'::bytea::vector;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type bytea to vector", position: 20, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "vector columns",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE EXTENSION vector;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE items (id INT PRIMARY KEY, v vector(3), h halfvec(2), s sparsevec(5), vs vector[]);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO items VALUES (1, '[1,2,3]', '[1.5,2]', '{1:1,3:2}/5', ARRAY['[1,2]'::vector]), (2, '[4,5,6]', '[0,1]', '{}/5', NULL);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO items (id, v) VALUES (3, '[1,2]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 3 dimensions, not 2", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO items (id, v) VALUES (3, ARRAY[1,2,3]);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO items (id, h) VALUES (4, '[1,2]'::vector);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM items ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", USER_DEFINED), Column("h", USER_DEFINED), Column("s", USER_DEFINED), Column("vs", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("[1,2,3]"), T("[1.5,2]"), T("{1:1,3:2}/5"), T(r#"{"[1,2]"}"#)],
                            &[T("2"), T("[4,5,6]"), T("[0,1]"), T("{}/5"), Null],
                            &[T("3"), T("[1,2,3]"), Null, Null, Null],
                            &[T("4"), Null, T("[1,2]"), Null, Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v <-> '[1,1,1]' AS d FROM items ORDER BY d LIMIT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("d", FLOAT8)],
                        rows: &[
                            &[T("1"), T("2.23606797749979")],
                            &[T("3"), T("2.23606797749979")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM items ORDER BY v DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("2")],
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v), sum(h) FROM items;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED), Column("sum", USER_DEFINED)],
                        rows: &[
                            &[T("[2,3,4]"), T("[2.5,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE vk (v vector(2) PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO vk VALUES ('[2,1]'), ('[1,2]');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO vk VALUES ('[1,2]');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "vk_pkey""#, detail: "Key (v)=([1,2]) already exists.", schema: "public", table: "vk", constraint: "vk_pkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM vk ORDER BY v;",
                    expected: Expected::Rows {
                        columns: &[Column("v", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2]")],
                            &[T("[2,1]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT attname, atttypmod, format_type(atttypid, atttypmod) FROM pg_attribute WHERE attrelid = 'items'::regclass AND attnum > 0 ORDER BY attnum;",
                    expected: Expected::Rows {
                        columns: &[Column("attname", NAME), Column("atttypmod", INT4), Column("format_type", TEXT)],
                        rows: &[
                            &[T("id"), T("-1"), T("integer")],
                            &[T("v"), T("3"), T("vector(3)")],
                            &[T("h"), T("2"), T("halfvec(2)")],
                            &[T("s"), T("5"), T("sparsevec(5)")],
                            &[T("vs"), T("-1"), T("vector[]")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension vector because other objects depend on it", detail: r#"column vs of table items depends on type vector[]
column v of table items depends on type vector
column v of table vk depends on type vector
column h of table items depends on type halfvec
column s of table items depends on type sparsevec"#, hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "extension catalogs",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM pg_available_extension_versions WHERE name IN ('plpgsql', 'uuid-ossp', 'vector') ORDER BY name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("version", TEXT), Column("installed", BOOL), Column("superuser", BOOL), Column("trusted", BOOL), Column("relocatable", BOOL), Column("schema", NAME), Column("requires", NAME_ARRAY), Column("comment", TEXT)],
                        rows: &[
                            &[T("plpgsql"), T("1.0"), T("t"), T("t"), T("t"), T("f"), T("pg_catalog"), Null, T("PL/pgSQL procedural language")],
                            &[T("uuid-ossp"), T("1.1"), T("f"), T("t"), T("t"), T("t"), Null, Null, T("generate universally unique identifiers (UUIDs)")],
                            &[T("vector"), T("0.8.6"), T("f"), T("t"), T("f"), T("t"), Null, Null, T("vector data type and ivfflat and hnsw access methods")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE EXTENSION vector;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, installed_version FROM pg_available_extensions WHERE name IN ('plpgsql', 'uuid-ossp', 'vector') ORDER BY name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("installed_version", TEXT)],
                        rows: &[
                            &[T("plpgsql"), T("1.0")],
                            &[T("uuid-ossp"), Null],
                            &[T("vector"), T("0.8.6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT amname, amtype, amhandler::text FROM pg_am ORDER BY amname;",
                    expected: Expected::Rows {
                        columns: &[Column("amname", NAME), Column("amtype", CHAR), Column("amhandler", TEXT)],
                        rows: &[
                            &[T("brin"), T("i"), T("brinhandler")],
                            &[T("btree"), T("i"), T("bthandler")],
                            &[T("gin"), T("i"), T("ginhandler")],
                            &[T("gist"), T("i"), T("gisthandler")],
                            &[T("hash"), T("i"), T("hashhandler")],
                            &[T("heap"), T("t"), T("heap_tableam_handler")],
                            &[T("hnsw"), T("i"), T("hnswhandler")],
                            &[T("ivfflat"), T("i"), T("ivfflathandler")],
                            &[T("spgist"), T("i"), T("spghandler")],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT o.opcname, a.amname, o.opcdefault FROM pg_opclass o JOIN pg_am a ON a.oid = o.opcmethod WHERE a.amname IN ('hnsw', 'ivfflat') ORDER BY a.amname, o.opcname;",
                    expected: Expected::Rows {
                        columns: &[Column("opcname", NAME), Column("amname", NAME), Column("opcdefault", BOOL)],
                        rows: &[
                            &[T("bit_hamming_ops"), T("hnsw"), T("f")],
                            &[T("bit_jaccard_ops"), T("hnsw"), T("f")],
                            &[T("halfvec_cosine_ops"), T("hnsw"), T("f")],
                            &[T("halfvec_ip_ops"), T("hnsw"), T("f")],
                            &[T("halfvec_l1_ops"), T("hnsw"), T("f")],
                            &[T("halfvec_l2_ops"), T("hnsw"), T("f")],
                            &[T("sparsevec_cosine_ops"), T("hnsw"), T("f")],
                            &[T("sparsevec_ip_ops"), T("hnsw"), T("f")],
                            &[T("sparsevec_l1_ops"), T("hnsw"), T("f")],
                            &[T("sparsevec_l2_ops"), T("hnsw"), T("f")],
                            &[T("vector_cosine_ops"), T("hnsw"), T("f")],
                            &[T("vector_ip_ops"), T("hnsw"), T("f")],
                            &[T("vector_l1_ops"), T("hnsw"), T("f")],
                            &[T("vector_l2_ops"), T("hnsw"), T("f")],
                            &[T("bit_hamming_ops"), T("ivfflat"), T("f")],
                            &[T("halfvec_cosine_ops"), T("ivfflat"), T("f")],
                            &[T("halfvec_ip_ops"), T("ivfflat"), T("f")],
                            &[T("halfvec_l2_ops"), T("ivfflat"), T("f")],
                            &[T("vector_cosine_ops"), T("ivfflat"), T("f")],
                            &[T("vector_ip_ops"), T("ivfflat"), T("f")],
                            &[T("vector_l2_ops"), T("ivfflat"), T("t")],
                        ],
                        tag: "SELECT 21",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT typname, typtype, typcategory, typlen, typstorage FROM pg_type WHERE typname IN ('vector', '_vector', 'halfvec', 'sparsevec') ORDER BY typname;",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME), Column("typtype", CHAR), Column("typcategory", CHAR), Column("typlen", INT2), Column("typstorage", CHAR)],
                        rows: &[
                            &[T("_vector"), T("b"), T("A"), T("-1"), T("x")],
                            &[T("halfvec"), T("b"), T("U"), T("-1"), T("e")],
                            &[T("sparsevec"), T("b"), T("U"), T("-1"), T("e")],
                            &[T("vector"), T("b"), T("U"), T("-1"), T("e")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
