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
use harness::script::Cell::{Any, Null, Oid, Text as T};
use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};
use harness::wire::{Datum, F, Field, Fields, PGX_STARTUP, Receive, Send, Step, W, WireTest, run_wire_tests};

#[test]
fn test_drop_extension() {
    run_scripts(&[
        ScriptTest {
            name: "drop extension that does not exist",
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP EXTENSION IF EXISTS pg_graphql;",
                    expected: Expected::Tag("DROP EXTENSION"),
                    notices: &[Diagnostic { code: "00000", message: r#"extension "pg_graphql" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION IF EXISTS pg_graphql, doltgres_no_such_extension;",
                    expected: Expected::Tag("DROP EXTENSION"),
                    notices: &[Diagnostic { code: "00000", message: r#"extension "pg_graphql" does not exist, skipping"#, ..N }, Diagnostic { code: "00000", message: r#"extension "doltgres_no_such_extension" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION pg_graphql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"extension "pg_graphql" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension removes its functions",
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
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pg_catalog.pg_extension WHERE extname = 'uuid-ossp';",
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
                    query: "SELECT count(*) FROM pg_catalog.pg_proc WHERE proname LIKE 'uuid_ns_%' OR proname LIKE 'uuid_generate_%' OR proname = 'uuid_nil';",
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
            name: "drop extension with if exists drops the extensions that exist",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION IF EXISTS pg_graphql, "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    notices: &[Diagnostic { code: "00000", message: r#"extension "pg_graphql" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pg_catalog.pg_proc WHERE proname = 'uuid_nil';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension drops nothing when a name does not exist",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp", pg_graphql;"#,
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"extension "pg_graphql" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pg_catalog.pg_proc WHERE proname = 'uuid_nil';",
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
            name: "drop extension created in a schema outside the search path",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp" WITH SCHEMA public;"#,
                "SELECT pg_catalog.set_config('search_path', '', false);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp" CASCADE;"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pg_catalog.pg_proc WHERE proname = 'uuid_nil';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension removes its types, operators, casts, and aggregates",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector + '[3,4]'::vector, avg(v) FROM (VALUES ('[1,2]'::vector), ('[3,4]'::vector)) t (v);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED), Column("avg", USER_DEFINED)],
                        rows: &[
                            &[T("[4,6]"), T("[2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pg_catalog.pg_type WHERE typname IN ('vector', '_vector', 'halfvec', '_halfvec', 'sparsevec', '_sparsevec');",
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
                    query: "SELECT '[1,2]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "vector" does not exist"#, position: 17, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE EXTENSION vector;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector + '[3,4]'::vector, avg(v) FROM (VALUES ('[1,2]'::vector), ('[3,4]'::vector)) t (v);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED), Column("avg", USER_DEFINED)],
                        rows: &[
                            &[T("[4,6]"), T("[2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension whose types are used by a table",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
                "CREATE EXTENSION vector;",
                "CREATE TABLE items (id INT PRIMARY KEY, embedding vector(2));",
                "CREATE TABLE item_lists (id INT PRIMARY KEY, embeddings halfvec[]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension vector because other objects depend on it", detail: r#"column embedding of table items depends on type vector
column embeddings of table item_lists depends on type halfvec[]"#, hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp", vector;"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop desired object(s) because other objects depend on them", detail: r#"column embedding of table items depends on type vector
column embeddings of table item_lists depends on type halfvec[]"#, hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
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
                    query: "DROP TABLE items;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector RESTRICT;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension vector because other objects depend on it", detail: "column embeddings of table item_lists depends on type halfvec[]", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE item_lists;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension cascades to the columns that use its types",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE items (id INT PRIMARY KEY, embedding vector(2));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector CASCADE;",
                    expected: Expected::Tag("DROP EXTENSION"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to column embedding of table items", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT column_name FROM information_schema.columns WHERE table_name = 'items';",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME)],
                        rows: &[
                            &[T("id")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension whose functions are used by a column default",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
                "CREATE TABLE goals (id uuid DEFAULT uuid_generate_v4(), other uuid DEFAULT gen_random_uuid());",
                "CREATE TABLE notes (id INT PRIMARY KEY, label text DEFAULT upper(uuid_nil()::text || 'x'));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: r#"default value for column label of table notes depends on function uuid_nil()
default value for column id of table goals depends on function uuid_generate_v4()"#, hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE goals ALTER COLUMN id DROP DEFAULT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: "default value for column label of table notes depends on function uuid_nil()", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE notes ALTER COLUMN label DROP DEFAULT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension cascades to the column defaults that use its functions",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
                "CREATE TABLE goals (id uuid DEFAULT uuid_generate_v4());",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp" CASCADE;"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to default value for column id of table goals", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT column_default FROM information_schema.columns WHERE table_name = 'goals';",
                    expected: Expected::Rows {
                        columns: &[Column("column_default", VARCHAR)],
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
        ScriptTest {
            name: "drop extension whose functions are used by a generated column",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
                "CREATE TABLE gen (id INT PRIMARY KEY, nil_text text GENERATED ALWAYS AS (uuid_nil()::text) STORED);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: "column nil_text of table gen depends on function uuid_nil()", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE gen;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension whose functions are used by a view",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
                "CREATE VIEW plain_view AS SELECT gen_random_uuid() AS n;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE VIEW nil_view AS SELECT uuid_nil() AS n;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: "view nil_view depends on function uuid_nil()", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP VIEW nil_view;",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW ns_view AS SELECT x.n FROM (SELECT uuid_ns_dns() AS n) x UNION SELECT gen_random_uuid();",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: "view ns_view depends on function uuid_ns_dns()", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP VIEW ns_view;",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension whose functions are used by a CHECK constraint",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
                "CREATE TABLE checked (t text CHECK (t <> uuid_nil()::text));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: "constraint checked_t_check on table checked depends on function uuid_nil()", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE checked;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension whose types are used by casts",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE checked (t text CHECK (t::vector IS NOT NULL));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension vector because other objects depend on it", detail: "constraint checked_t_check on table checked depends on type vector", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE checked;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE defaulted (t text DEFAULT ('[1,2]'::vector)::text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension vector because other objects depend on it", detail: "default value for column t of table defaulted depends on type vector", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE defaulted;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW cast_view AS SELECT '[1,2]'::vector AS v;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension vector because other objects depend on it", detail: "view cast_view depends on type vector", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP VIEW cast_view;",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW array_view AS SELECT '{}'::halfvec[] AS v;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension vector because other objects depend on it", detail: "view array_view depends on type halfvec[]", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP VIEW array_view;",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension whose functions are used by a domain",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DOMAIN defaulted_uuid AS uuid DEFAULT uuid_nil();",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: "type defaulted_uuid depends on function uuid_nil()", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN defaulted_uuid;",
                    expected: Expected::Tag("DROP DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN checked_uuid AS uuid CHECK (VALUE <> uuid_nil());",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: "constraint checked_uuid_check depends on function uuid_nil()", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN checked_uuid;",
                    expected: Expected::Tag("DROP DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension whose types are used by a domain",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DOMAIN based AS vector;",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension vector because other objects depend on it", detail: "type based depends on type vector", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN based;",
                    expected: Expected::Tag("DROP DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN casted AS text DEFAULT ('[1,2]'::vector)::text;",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension vector because other objects depend on it", detail: "type casted depends on type vector", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN casted;",
                    expected: Expected::Tag("DROP DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN casted_check AS text CHECK (VALUE::vector IS NOT NULL);",
                    expected: Expected::Tag("CREATE DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension vector because other objects depend on it", detail: "constraint casted_check_check depends on type vector", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DOMAIN casted_check;",
                    expected: Expected::Tag("DROP DOMAIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP EXTENSION vector;",
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop extension whose functions are used by window functions and unions",
            set_up_script: &[
                r#"CREATE EXTENSION "uuid-ossp";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE VIEW window_view AS SELECT row_number() OVER (ORDER BY uuid_nil()) AS n FROM (VALUES (1)) t (i);",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: "view window_view depends on function uuid_nil()", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP VIEW window_view;",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW named_window_view AS SELECT row_number() OVER w AS n FROM (VALUES (1)) t (i) WINDOW w AS (PARTITION BY uuid_ns_url());",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: "view named_window_view depends on function uuid_ns_url()", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP VIEW named_window_view;",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW union_view AS SELECT gen_random_uuid() AS n UNION ALL SELECT uuid_ns_oid();",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop extension uuid-ossp because other objects depend on it", detail: "view union_view depends on function uuid_ns_oid()", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP VIEW union_view;",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DROP EXTENSION "uuid-ossp";"#,
                    expected: Expected::Tag("DROP EXTENSION"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
