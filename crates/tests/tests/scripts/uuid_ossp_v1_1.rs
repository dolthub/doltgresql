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
fn test_uuid_ossp() {
    run_scripts(&[
        ScriptTest {
            name: "uuid-ossp",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT uuid_ns_url();",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_ns_url", UUID)],
                        rows: &[
                            &[T("6ba7b811-9dad-11d1-80b4-00c04fd430c8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3('00000000-0000-0000-0000-000000000000'::uuid, 'example text');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v3", UUID)],
                        rows: &[
                            &[T("a55b875a-1bd9-31af-ac66-7d8323785c6e")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3('00000000-0000-0000-0000-000000000001'::uuid, 'example text');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v3", UUID)],
                        rows: &[
                            &[T("a319ab51-8e26-37c6-942f-7dd5fda5c3ef")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3(uuid_ns_url(), 'example text');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v3", UUID)],
                        rows: &[
                            &[T("6541262f-d622-3e35-8873-2b227591bf69")],
                        ],
                        tag: "SELECT 1",
                    },
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
                ScriptTestAssertion {
                    query: "SELECT length(uuid_nil()::text);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("36")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(uuid_generate_v4()::text);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("36")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v4() = uuid_nil();",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH u1 AS (SELECT uuid_nil() AS id), u2 AS (SELECT uuid_nil() AS id) SELECT (SELECT id FROM u1) = (SELECT id FROM u2);",
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
                    query: "WITH u1 AS (SELECT uuid_generate_v4() AS id), u2 AS (SELECT uuid_generate_v4() AS id) SELECT (SELECT id FROM u1) = (SELECT id FROM u2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "uuid-ossp namespace functions",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
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
                ScriptTestAssertion {
                    query: "SELECT uuid_ns_dns();",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_ns_dns", UUID)],
                        rows: &[
                            &[T("6ba7b810-9dad-11d1-80b4-00c04fd430c8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_ns_url();",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_ns_url", UUID)],
                        rows: &[
                            &[T("6ba7b811-9dad-11d1-80b4-00c04fd430c8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_ns_oid();",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_ns_oid", UUID)],
                        rows: &[
                            &[T("6ba7b812-9dad-11d1-80b4-00c04fd430c8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_ns_x500();",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_ns_x500", UUID)],
                        rows: &[
                            &[T("6ba7b814-9dad-11d1-80b4-00c04fd430c8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_ns_dns() = uuid_ns_dns(), uuid_ns_dns() = uuid_ns_url();",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "uuid-ossp uuid_generate_v3",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3(uuid_ns_dns(), 'www.postgresql.org');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v3", UUID)],
                        rows: &[
                            &[T("9a0d5f51-76ff-394e-ba97-b28a9ff12209")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3(uuid_nil(), '');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v3", UUID)],
                        rows: &[
                            &[T("4ae71336-e44b-39bf-b9d2-752e234818a5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3(uuid_ns_dns(), 'héllo wörld');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v3", UUID)],
                        rows: &[
                            &[T("2f301b42-2eaf-3cc3-8646-69e0ffde841f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3(uuid_ns_url(), repeat('a', 1000));",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v3", UUID)],
                        rows: &[
                            &[T("d8f8a14e-39ec-3186-8107-4d4e5a41d2c0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring(uuid_generate_v3(uuid_nil(), 'x')::text, 15, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring(uuid_generate_v3(uuid_nil(), 'x')::text, 20, 1) IN ('8', '9', 'a', 'b');",
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
                    query: "SELECT uuid_generate_v3(uuid_nil(), 'abc') = uuid_generate_v3(uuid_nil(), 'abc');",
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
                    query: "SELECT uuid_generate_v3(uuid_nil(), 'abc') = uuid_generate_v3(uuid_nil(), 'ABC');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3(uuid_ns_dns(), 'abc') = uuid_generate_v3(uuid_ns_url(), 'abc');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v3(NULL, 'abc') IS NULL, uuid_generate_v3(uuid_nil(), NULL) IS NULL;",
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
                    query: "SELECT uuid_generate_v3('not-a-uuid', 'abc');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type uuid: "not-a-uuid""#, position: 25, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "uuid-ossp uuid_generate_v5",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v5(uuid_ns_url(), 'example text');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v5", UUID)],
                        rows: &[
                            &[T("59edfb26-7819-5209-86a3-79a6da9035ba")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v5(uuid_ns_dns(), 'www.postgresql.org');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v5", UUID)],
                        rows: &[
                            &[T("1826c6c4-4d1f-534f-9dcd-7a15978dfeb9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v5(uuid_ns_oid(), 'example text');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v5", UUID)],
                        rows: &[
                            &[T("5758e964-d604-5cde-9b32-57368fc3b1ff")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v5(uuid_ns_x500(), 'example text');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v5", UUID)],
                        rows: &[
                            &[T("1d8aac60-3096-5014-bfd7-1f816c37cf50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v5(uuid_nil(), '');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v5", UUID)],
                        rows: &[
                            &[T("e129f27c-5103-5c5c-844b-cdf0a15e160d")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v5(uuid_ns_dns(), 'héllo wörld');",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v5", UUID)],
                        rows: &[
                            &[T("d24fabfc-fb83-5476-8201-39e27376a62b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v5(uuid_ns_url(), repeat('a', 1000));",
                    expected: Expected::Rows {
                        columns: &[Column("uuid_generate_v5", UUID)],
                        rows: &[
                            &[T("7f46a8f9-f8ba-5a67-983a-ffc2101475df")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring(uuid_generate_v5(uuid_nil(), 'x')::text, 15, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring(uuid_generate_v5(uuid_nil(), 'x')::text, 20, 1) IN ('8', '9', 'a', 'b');",
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
                    query: "SELECT uuid_generate_v5(uuid_nil(), 'abc') = uuid_generate_v5(uuid_nil(), 'abc');",
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
                    query: "SELECT uuid_generate_v3(uuid_nil(), 'abc') = uuid_generate_v5(uuid_nil(), 'abc');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT uuid_generate_v5(NULL, 'abc') IS NULL, uuid_generate_v5(uuid_nil(), NULL) IS NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "uuid-ossp uuid_generate_v1 and uuid_generate_v1mc",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT substring(uuid_generate_v1()::text, 15, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring(uuid_generate_v1mc()::text, 15, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring(uuid_generate_v1()::text, 20, 1) IN ('8', '9', 'a', 'b');",
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
                    query: "SELECT substring(uuid_generate_v1mc()::text, 20, 1) IN ('8', '9', 'a', 'b');",
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
                    query: "SELECT substring(uuid_generate_v1mc()::text, 26, 1) IN ('3', '7', 'b', 'f');",
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
                    query: "SELECT substring(uuid_generate_v1()::text, 25) = substring(uuid_generate_v1mc()::text, 25);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(DISTINCT id::text) FROM (SELECT uuid_generate_v1() AS id FROM generate_series(1, 50)) t;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(DISTINCT id::text) FROM (SELECT uuid_generate_v1mc() AS id FROM generate_series(1, 50)) t;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(uuid_generate_v1()::text), length(uuid_generate_v1mc()::text);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("length", INT4)],
                        rows: &[
                            &[T("36"), T("36")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "uuid-ossp uuid_generate_v4",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT substring(uuid_generate_v4()::text, 15, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring(uuid_generate_v4()::text, 20, 1) IN ('8', '9', 'a', 'b');",
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
                    query: "SELECT count(DISTINCT id::text) FROM (SELECT uuid_generate_v4() AS id FROM generate_series(1, 100)) t;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "uuid-ossp functions used by a table",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
                "CREATE TABLE items (id uuid PRIMARY KEY DEFAULT uuid_generate_v4(), name text NOT NULL);",
                "INSERT INTO items (name) VALUES ('first'), ('second'), ('third');",
                "CREATE TABLE named (id uuid PRIMARY KEY, name text NOT NULL);",
                "INSERT INTO named VALUES (uuid_generate_v5(uuid_ns_url(), 'example text'), 'example');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*), count(DISTINCT id::text) FROM items;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8)],
                        rows: &[
                            &[T("3"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM items ORDER BY name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("first")],
                            &[T("second")],
                            &[T("third")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, name FROM named;",
                    expected: Expected::Rows {
                        columns: &[Column("id", UUID), Column("name", TEXT)],
                        rows: &[
                            &[T("59edfb26-7819-5209-86a3-79a6da9035ba"), T("example")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM named WHERE id = uuid_generate_v5(uuid_ns_url(), 'example text');",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("example")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "uuid-ossp catalog tables",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT extname, extrelocatable, extversion FROM pg_catalog.pg_extension WHERE extname = 'uuid-ossp';",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME), Column("extrelocatable", BOOL), Column("extversion", TEXT)],
                        rows: &[
                            &[T("uuid-ossp"), T("t"), T("1.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, default_version, installed_version, comment FROM pg_catalog.pg_available_extensions WHERE name = 'uuid-ossp';",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("default_version", TEXT), Column("installed_version", TEXT), Column("comment", TEXT)],
                        rows: &[
                            &[T("uuid-ossp"), T("1.1"), T("1.1"), T("generate universally unique identifiers (UUIDs)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, version, installed, superuser, trusted, relocatable, schema, requires FROM pg_catalog.pg_available_extension_versions WHERE name = 'uuid-ossp';",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("version", TEXT), Column("installed", BOOL), Column("superuser", BOOL), Column("trusted", BOOL), Column("relocatable", BOOL), Column("schema", NAME), Column("requires", NAME_ARRAY)],
                        rows: &[
                            &[T("uuid-ossp"), T("1.1"), T("t"), T("t"), T("t"), T("t"), Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pg_catalog.pg_proc WHERE proname LIKE 'uuid_ns_%' OR proname LIKE 'uuid_generate_%' OR proname = 'uuid_nil';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT proname FROM pg_catalog.pg_proc WHERE proname LIKE 'uuid_ns_%' OR proname LIKE 'uuid_generate_%' OR proname = 'uuid_nil' ORDER BY proname;",
                    expected: Expected::Rows {
                        columns: &[Column("proname", NAME)],
                        rows: &[
                            &[T("uuid_generate_v1")],
                            &[T("uuid_generate_v1mc")],
                            &[T("uuid_generate_v3")],
                            &[T("uuid_generate_v4")],
                            &[T("uuid_generate_v5")],
                            &[T("uuid_nil")],
                            &[T("uuid_ns_dns")],
                            &[T("uuid_ns_oid")],
                            &[T("uuid_ns_url")],
                            &[T("uuid_ns_x500")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
