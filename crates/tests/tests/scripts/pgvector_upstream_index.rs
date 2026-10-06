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
fn test_pgvector_upstream_index() {
    run_scripts(&[
        ScriptTest {
            name: "vector indexes require a primary key",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE t (val vector(3));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING hnsw (val vector_l2_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "hnsw vector",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING hnsw (val vector_l2_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (5, '[1,2,4]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <-> (SELECT NULL::vector)) t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "TRUNCATE t;",
                    expected: Expected::Tag("TRUNCATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING hnsw (val vector_ip_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (5, '[1,2,4]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <#> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,4]")],
                            &[T("[1,2,3]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <#> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,4]")],
                            &[T("[1,2,3]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <#> (SELECT NULL::vector)) t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING hnsw (val vector_cosine_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (5, '[1,2,4]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <=> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,1,1]")],
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[0,0,0]")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <=> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,1,1]")],
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[0,0,0]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <=> '[0,0,0]') t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <=> (SELECT NULL::vector)) t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT t.val, t2.val FROM t CROSS JOIN LATERAL (SELECT val FROM t t3 ORDER BY val <=> t.val LIMIT 1) t2 WHERE t.val != '[0,0,0]' ORDER BY t.val;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED), Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,1,1]"), T("[1,1,1]")],
                            &[T("[1,2,3]"), T("[1,2,3]")],
                            &[T("[1,2,4]"), T("[1,2,4]")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING hnsw (val vector_l1_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (5, '[1,2,4]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <+> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <+> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <+> (SELECT NULL::vector)) t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "hnsw halfvec",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING hnsw (val halfvec_l2_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (5, '[1,2,4]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <-> (SELECT NULL::halfvec)) t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "TRUNCATE t;",
                    expected: Expected::Tag("TRUNCATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING hnsw (val halfvec_ip_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (5, '[1,2,4]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <#> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,4]")],
                            &[T("[1,2,3]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <#> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,4]")],
                            &[T("[1,2,3]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <#> (SELECT NULL::halfvec)) t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING hnsw (val halfvec_cosine_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (5, '[1,2,4]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <=> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,1,1]")],
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[0,0,0]")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <=> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,1,1]")],
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[0,0,0]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <=> '[0,0,0]') t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <=> (SELECT NULL::halfvec)) t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING hnsw (val halfvec_l1_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (5, '[1,2,4]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <+> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <+> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <+> (SELECT NULL::halfvec)) t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ivfflat vector",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING ivfflat (val vector_l2_ops) WITH (lists = 1);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (5, '[1,2,4]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <-> (SELECT NULL::vector)) t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres extension: Dolt stores only a vector index's distance, so TRUNCATE cannot tell that it rebuilt an
                // ivfflat index and warn about its recall.
                ScriptTestAssertion {
                    query: "TRUNCATE t;",
                    expected: Expected::Tag("TRUNCATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ivfflat halfvec",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(3));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON t USING ivfflat (val halfvec_l2_ops) WITH (lists = 1);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (5, '[1,2,4]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                            &[Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[1,2,4]")],
                            &[T("[1,1,1]")],
                            &[T("[0,0,0]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <-> (SELECT NULL::halfvec)) t2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
