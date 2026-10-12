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
fn test_pgvector_catalog() {
    run_scripts(&[
        ScriptTest {
            name: "dimensioned types expose their modifiers through pg_attribute",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE vector_typmod_probe (v vector(1024), h halfvec(7), s sparsevec(99), unbounded vector, vectors vector(3)[]);",
                "CREATE VIEW vector_typmod_view AS SELECT v FROM vector_typmod_probe;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT attribute.attname, attribute.atttypmod, format_type(attribute.atttypid, attribute.atttypmod)
FROM pg_attribute AS attribute
JOIN pg_class AS relation ON relation.oid = attribute.attrelid
WHERE relation.relname = 'vector_typmod_probe' AND attribute.attnum > 0 AND NOT attribute.attisdropped
ORDER BY attribute.attnum;"#,
                    expected: Expected::Rows {
                        columns: &[Column("attname", NAME), Column("atttypmod", INT4), Column("format_type", TEXT)],
                        rows: &[
                            &[T("v"), T("1024"), T("vector(1024)")],
                            &[T("h"), T("7"), T("halfvec(7)")],
                            &[T("s"), T("99"), T("sparsevec(99)")],
                            &[T("unbounded"), T("-1"), T("vector")],
                            &[T("vectors"), T("3"), T("vector(3)[]")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT attribute.atttypmod, format_type(attribute.atttypid, attribute.atttypmod)
FROM pg_attribute AS attribute
JOIN pg_class AS relation ON relation.oid = attribute.attrelid
WHERE relation.relname = 'vector_typmod_view' AND attribute.attname = 'v';"#,
                    expected: Expected::Rows {
                        columns: &[Column("atttypmod", INT4), Column("format_type", TEXT)],
                        rows: &[
                            &[T("1024"), T("vector(1024)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "user-defined type names are quoted and qualified according to visibility",
            set_up_script: &[
                "CREATE SCHEMA first_schema;",
                "CREATE SCHEMA second_schema;",
                r#"CREATE TYPE first_schema."Mixed Type" AS ENUM ('first');"#,
                r#"CREATE TYPE second_schema."Mixed Type" AS ENUM ('second');"#,
                "SET search_path TO first_schema, second_schema, public;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT namespace.nspname, format_type(type.oid, NULL)
FROM pg_type AS type
JOIN pg_namespace AS namespace ON namespace.oid = type.typnamespace
WHERE type.typname = 'Mixed Type'
ORDER BY namespace.nspname;"#,
                    expected: Expected::Rows {
                        columns: &[Column("nspname", NAME), Column("format_type", TEXT)],
                        rows: &[
                            &[T("first_schema"), T(r#""Mixed Type""#)],
                            &[T("second_schema"), T(r#"second_schema."Mixed Type""#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_am lists the extension access methods once installed",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT amname, amtype FROM pg_catalog.pg_am ORDER BY amname;",
                    expected: Expected::Rows {
                        columns: &[Column("amname", NAME), Column("amtype", CHAR)],
                        rows: &[
                            &[T("brin"), T("i")],
                            &[T("btree"), T("i")],
                            &[T("gin"), T("i")],
                            &[T("gist"), T("i")],
                            &[T("hash"), T("i")],
                            &[T("heap"), T("t")],
                            &[T("spgist"), T("i")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE EXTENSION vector;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT amname, amtype FROM pg_catalog.pg_am ORDER BY amname;",
                    expected: Expected::Rows {
                        columns: &[Column("amname", NAME), Column("amtype", CHAR)],
                        rows: &[
                            &[T("brin"), T("i")],
                            &[T("btree"), T("i")],
                            &[T("gin"), T("i")],
                            &[T("gist"), T("i")],
                            &[T("hash"), T("i")],
                            &[T("heap"), T("t")],
                            &[T("hnsw"), T("i")],
                            &[T("ivfflat"), T("i")],
                            &[T("spgist"), T("i")],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT amname, amhandler::text FROM pg_catalog.pg_am WHERE amname IN ('hnsw', 'ivfflat') ORDER BY amname;",
                    expected: Expected::Rows {
                        columns: &[Column("amname", NAME), Column("amhandler", TEXT)],
                        rows: &[
                            &[T("hnsw"), T("hnswhandler")],
                            &[T("ivfflat"), T("ivfflathandler")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_opclass lists the extension operator classes once installed",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pg_catalog.pg_opclass o JOIN pg_catalog.pg_am a ON a.oid = o.opcmethod WHERE a.amname IN ('hnsw', 'ivfflat');",
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
                    query: "CREATE EXTENSION vector;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT o.opcname, a.amname, o.opcdefault FROM pg_catalog.pg_opclass o JOIN pg_catalog.pg_am a ON a.oid = o.opcmethod WHERE a.amname = 'hnsw' ORDER BY o.opcname;",
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
                        ],
                        tag: "SELECT 14",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT o.opcname, a.amname, o.opcdefault FROM pg_catalog.pg_opclass o JOIN pg_catalog.pg_am a ON a.oid = o.opcmethod WHERE a.amname = 'ivfflat' ORDER BY o.opcname;",
                    expected: Expected::Rows {
                        columns: &[Column("opcname", NAME), Column("amname", NAME), Column("opcdefault", BOOL)],
                        rows: &[
                            &[T("bit_hamming_ops"), T("ivfflat"), T("f")],
                            &[T("halfvec_cosine_ops"), T("ivfflat"), T("f")],
                            &[T("halfvec_ip_ops"), T("ivfflat"), T("f")],
                            &[T("halfvec_l2_ops"), T("ivfflat"), T("f")],
                            &[T("vector_cosine_ops"), T("ivfflat"), T("f")],
                            &[T("vector_ip_ops"), T("ivfflat"), T("f")],
                            &[T("vector_l2_ops"), T("ivfflat"), T("t")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT o.opcname, t.typname FROM pg_catalog.pg_opclass o JOIN pg_catalog.pg_type t ON t.oid = o.opcintype JOIN pg_catalog.pg_am a ON a.oid = o.opcmethod WHERE a.amname = 'hnsw' ORDER BY o.opcname;",
                    expected: Expected::Rows {
                        columns: &[Column("opcname", NAME), Column("typname", NAME)],
                        rows: &[
                            &[T("bit_hamming_ops"), T("bit")],
                            &[T("bit_jaccard_ops"), T("bit")],
                            &[T("halfvec_cosine_ops"), T("halfvec")],
                            &[T("halfvec_ip_ops"), T("halfvec")],
                            &[T("halfvec_l1_ops"), T("halfvec")],
                            &[T("halfvec_l2_ops"), T("halfvec")],
                            &[T("sparsevec_cosine_ops"), T("sparsevec")],
                            &[T("sparsevec_ip_ops"), T("sparsevec")],
                            &[T("sparsevec_l1_ops"), T("sparsevec")],
                            &[T("sparsevec_l2_ops"), T("sparsevec")],
                            &[T("vector_cosine_ops"), T("vector")],
                            &[T("vector_ip_ops"), T("vector")],
                            &[T("vector_l1_ops"), T("vector")],
                            &[T("vector_l2_ops"), T("vector")],
                        ],
                        tag: "SELECT 14",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT o.opcname, n.nspname FROM pg_catalog.pg_opclass o JOIN pg_catalog.pg_namespace n ON n.oid = o.opcnamespace JOIN pg_catalog.pg_am a ON a.oid = o.opcmethod WHERE a.amname = 'hnsw' AND o.opcname LIKE 'vector%' ORDER BY o.opcname;",
                    expected: Expected::Rows {
                        columns: &[Column("opcname", NAME), Column("nspname", NAME)],
                        rows: &[
                            &[T("vector_cosine_ops"), T("public")],
                            &[T("vector_ip_ops"), T("public")],
                            &[T("vector_l1_ops"), T("public")],
                            &[T("vector_l2_ops"), T("public")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "vector index definitions render with the hnsw method and operator class",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE items (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL);",
                "CREATE INDEX idx_cos ON items USING hnsw (v vector_cosine_ops);",
                "CREATE INDEX idx_ip ON items USING hnsw (v vector_ip_ops);",
                "CREATE INDEX idx_l1 ON items USING hnsw (v vector_l1_ops);",
                "CREATE INDEX idx_half ON items USING ivfflat (h halfvec_l2_ops);",
                "CREATE INDEX idx_def ON items USING ivfflat (v);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'items' ORDER BY indexname;",
                    expected: Expected::Rows {
                        columns: &[Column("indexname", NAME), Column("indexdef", TEXT)],
                        rows: &[
                            &[T("idx_cos"), T("CREATE INDEX idx_cos ON public.items USING hnsw (v vector_cosine_ops)")],
                            &[T("idx_def"), T("CREATE INDEX idx_def ON public.items USING ivfflat (v)")],
                            &[T("idx_half"), T("CREATE INDEX idx_half ON public.items USING ivfflat (h halfvec_l2_ops)")],
                            &[T("idx_ip"), T("CREATE INDEX idx_ip ON public.items USING hnsw (v vector_ip_ops)")],
                            &[T("idx_l1"), T("CREATE INDEX idx_l1 ON public.items USING hnsw (v vector_l1_ops)")],
                            &[T("items_pkey"), T("CREATE UNIQUE INDEX items_pkey ON public.items USING btree (id)")],
                        ],
                        tag: "SELECT 6",
                    },
                    skip: Some("Dolt stores only a vector index's distance, so every vector index renders with hnsw and its operator class"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_cos'::regclass);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("CREATE INDEX idx_cos ON public.items USING hnsw (v vector_cosine_ops)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_half'::regclass);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("CREATE INDEX idx_half ON public.items USING ivfflat (h halfvec_l2_ops)")],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("Dolt stores only a vector index's distance, so every vector index renders with hnsw and its operator class"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
