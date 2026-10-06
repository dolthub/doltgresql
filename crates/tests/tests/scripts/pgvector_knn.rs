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
fn test_pgvector_knn() {
    run_scripts(&[
        ScriptTest {
            name: "each vector metric is served by its matching index",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE knn (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL);",
                r#"INSERT INTO knn VALUES
			(1, '[1.2,0.3,2.1]', '[1.2,0.3,2.1]'),
			(2, '[2.5,1.1,0.4]', '[2.5,1.1,0.4]'),
			(3, '[0.9,2.2,1.7]', '[0.9,2.2,1.7]'),
			(4, '[-1.4,2.8,0.6]', '[-1.4,2.8,0.6]'),
			(5, '[3.1,-0.7,1.9]', '[3.1,-0.7,1.9]'),
			(6, '[0.2,0.8,-1.3]', '[0.2,0.8,-1.3]'),
			(7, '[4.6,3.2,2.4]', '[4.6,3.2,2.4]'),
			(8, '[-2.2,-1.1,3.3]', '[-2.2,-1.1,3.3]'),
			(9, '[1.8,1.6,0.9]', '[1.8,1.6,0.9]'),
			(10, '[0.4,-2.6,2.2]', '[0.4,-2.6,2.2]');"#,
                "CREATE INDEX idx_l2 ON knn USING hnsw (v vector_l2_ops);",
                "CREATE INDEX idx_ip ON knn USING hnsw (v vector_ip_ops);",
                "CREATE INDEX idx_cos ON knn USING hnsw (v vector_cosine_ops);",
                "CREATE INDEX idx_l1 ON knn USING hnsw (v vector_l1_ops);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "knn", columns: &["v"], ranges: "" }, PlanFact::NoSort]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 10;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                            &[T("5")],
                            &[T("6")],
                            &[T("4")],
                            &[T("10")],
                            &[T("7")],
                            &[T("8")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <#> '[1.5,1,2]' LIMIT 10;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("7")],
                            &[T("5")],
                            &[T("3")],
                            &[T("1")],
                            &[T("9")],
                            &[T("2")],
                            &[T("10")],
                            &[T("8")],
                            &[T("4")],
                            &[T("6")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <=> '[1.5,1,2]' LIMIT 10;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("7")],
                            &[T("3")],
                            &[T("9")],
                            &[T("5")],
                            &[T("2")],
                            &[T("10")],
                            &[T("4")],
                            &[T("8")],
                            &[T("6")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <+> '[1.5,1,2]' LIMIT 10;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                            &[T("5")],
                            &[T("6")],
                            &[T("10")],
                            &[T("7")],
                            &[T("4")],
                            &[T("8")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY '[1.5,1,2]' <-> v LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]'::vector LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4 OFFSET 3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("5")],
                            &[T("6")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "each halfvec metric is served by its matching index",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE knn (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL);",
                r#"INSERT INTO knn VALUES
			(1, '[1.2,0.3,2.1]', '[1.2,0.3,2.1]'),
			(2, '[2.5,1.1,0.4]', '[2.5,1.1,0.4]'),
			(3, '[0.9,2.2,1.7]', '[0.9,2.2,1.7]'),
			(4, '[-1.4,2.8,0.6]', '[-1.4,2.8,0.6]'),
			(5, '[3.1,-0.7,1.9]', '[3.1,-0.7,1.9]'),
			(6, '[0.2,0.8,-1.3]', '[0.2,0.8,-1.3]'),
			(7, '[4.6,3.2,2.4]', '[4.6,3.2,2.4]'),
			(8, '[-2.2,-1.1,3.3]', '[-2.2,-1.1,3.3]'),
			(9, '[1.8,1.6,0.9]', '[1.8,1.6,0.9]'),
			(10, '[0.4,-2.6,2.2]', '[0.4,-2.6,2.2]');"#,
                "CREATE INDEX idx_l2 ON knn USING hnsw (h halfvec_l2_ops);",
                "CREATE INDEX idx_ip ON knn USING hnsw (h halfvec_ip_ops);",
                "CREATE INDEX idx_cos ON knn USING hnsw (h halfvec_cosine_ops);",
                "CREATE INDEX idx_l1 ON knn USING hnsw (h halfvec_l1_ops);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM knn ORDER BY h <-> '[1.5,1,2]' LIMIT 4;",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "knn", columns: &["h"], ranges: "" }, PlanFact::NoSort]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY h <-> '[1.5,1,2]' LIMIT 10;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                            &[T("5")],
                            &[T("6")],
                            &[T("4")],
                            &[T("10")],
                            &[T("7")],
                            &[T("8")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY h <#> '[1.5,1,2]' LIMIT 10;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("7")],
                            &[T("5")],
                            &[T("3")],
                            &[T("1")],
                            &[T("9")],
                            &[T("2")],
                            &[T("10")],
                            &[T("8")],
                            &[T("4")],
                            &[T("6")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY h <=> '[1.5,1,2]' LIMIT 10;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("7")],
                            &[T("3")],
                            &[T("9")],
                            &[T("5")],
                            &[T("2")],
                            &[T("10")],
                            &[T("4")],
                            &[T("8")],
                            &[T("6")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY h <+> '[1.5,1,2]' LIMIT 10;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                            &[T("5")],
                            &[T("6")],
                            &[T("10")],
                            &[T("7")],
                            &[T("4")],
                            &[T("8")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ivfflat serves knn through the same native index",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE knn (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL);",
                r#"INSERT INTO knn VALUES
			(1, '[1.2,0.3,2.1]', '[1.2,0.3,2.1]'),
			(2, '[2.5,1.1,0.4]', '[2.5,1.1,0.4]'),
			(3, '[0.9,2.2,1.7]', '[0.9,2.2,1.7]'),
			(4, '[-1.4,2.8,0.6]', '[-1.4,2.8,0.6]'),
			(5, '[3.1,-0.7,1.9]', '[3.1,-0.7,1.9]'),
			(6, '[0.2,0.8,-1.3]', '[0.2,0.8,-1.3]'),
			(7, '[4.6,3.2,2.4]', '[4.6,3.2,2.4]'),
			(8, '[-2.2,-1.1,3.3]', '[-2.2,-1.1,3.3]'),
			(9, '[1.8,1.6,0.9]', '[1.8,1.6,0.9]'),
			(10, '[0.4,-2.6,2.2]', '[0.4,-2.6,2.2]');"#,
                "CREATE INDEX ON knn USING ivfflat (v);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "knn", columns: &["v"], ranges: "" }, PlanFact::NoSort]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "queries that cannot use the index fall back to an exact scan",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE knn (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL);",
                r#"INSERT INTO knn VALUES
			(1, '[1.2,0.3,2.1]', '[1.2,0.3,2.1]'),
			(2, '[2.5,1.1,0.4]', '[2.5,1.1,0.4]'),
			(3, '[0.9,2.2,1.7]', '[0.9,2.2,1.7]'),
			(4, '[-1.4,2.8,0.6]', '[-1.4,2.8,0.6]'),
			(5, '[3.1,-0.7,1.9]', '[3.1,-0.7,1.9]'),
			(6, '[0.2,0.8,-1.3]', '[0.2,0.8,-1.3]'),
			(7, '[4.6,3.2,2.4]', '[4.6,3.2,2.4]'),
			(8, '[-2.2,-1.1,3.3]', '[-2.2,-1.1,3.3]'),
			(9, '[1.8,1.6,0.9]', '[1.8,1.6,0.9]'),
			(10, '[0.4,-2.6,2.2]', '[0.4,-2.6,2.2]');"#,
                "CREATE INDEX idx_l2 ON knn USING hnsw (v vector_l2_ops);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' DESC LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("8")],
                            &[T("7")],
                            &[T("10")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn WHERE id > 3 ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("9")],
                            &[T("5")],
                            &[T("6")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM knn ORDER BY l2_distance(v, '[1.5,1,2]') LIMIT 4;",
                    expected: Expected::Plan(&[PlanFact::Sort, PlanFact::FullScan { table: "knn" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY l2_distance(v, '[1.5,1,2]') LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <=> '[1.5,1,2]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("7")],
                            &[T("3")],
                            &[T("9")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM (SELECT id FROM knn ORDER BY v <-> NULL LIMIT 4) sq;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM (SELECT id FROM knn ORDER BY v <-> NULL::vector LIMIT 4) sq;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> NULL, id LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM (SELECT id FROM knn ORDER BY v <-> v LIMIT 4) sq;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "bound parameter query vectors use the index",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE knn (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL);",
                r#"INSERT INTO knn VALUES
			(1, '[1.2,0.3,2.1]', '[1.2,0.3,2.1]'),
			(2, '[2.5,1.1,0.4]', '[2.5,1.1,0.4]'),
			(3, '[0.9,2.2,1.7]', '[0.9,2.2,1.7]'),
			(4, '[-1.4,2.8,0.6]', '[-1.4,2.8,0.6]'),
			(5, '[3.1,-0.7,1.9]', '[3.1,-0.7,1.9]'),
			(6, '[0.2,0.8,-1.3]', '[0.2,0.8,-1.3]'),
			(7, '[4.6,3.2,2.4]', '[4.6,3.2,2.4]'),
			(8, '[-2.2,-1.1,3.3]', '[-2.2,-1.1,3.3]'),
			(9, '[1.8,1.6,0.9]', '[1.8,1.6,0.9]'),
			(10, '[0.4,-2.6,2.2]', '[0.4,-2.6,2.2]');"#,
                "CREATE INDEX idx_l2 ON knn USING hnsw (v vector_l2_ops);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> $1 LIMIT 4;",
                    bind_vars: &[BindVar::Str("[1.5,1,2]")],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> $1 LIMIT 10;",
                    bind_vars: &[BindVar::Str("[1.5,1,2]")],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                            &[T("5")],
                            &[T("6")],
                            &[T("4")],
                            &[T("10")],
                            &[T("7")],
                            &[T("8")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "index maintenance is reflected in knn results",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE knn (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL);",
                r#"INSERT INTO knn VALUES
			(1, '[1.2,0.3,2.1]', '[1.2,0.3,2.1]'),
			(2, '[2.5,1.1,0.4]', '[2.5,1.1,0.4]'),
			(3, '[0.9,2.2,1.7]', '[0.9,2.2,1.7]'),
			(4, '[-1.4,2.8,0.6]', '[-1.4,2.8,0.6]'),
			(5, '[3.1,-0.7,1.9]', '[3.1,-0.7,1.9]'),
			(6, '[0.2,0.8,-1.3]', '[0.2,0.8,-1.3]'),
			(7, '[4.6,3.2,2.4]', '[4.6,3.2,2.4]'),
			(8, '[-2.2,-1.1,3.3]', '[-2.2,-1.1,3.3]'),
			(9, '[1.8,1.6,0.9]', '[1.8,1.6,0.9]'),
			(10, '[0.4,-2.6,2.2]', '[0.4,-2.6,2.2]');"#,
                "CREATE INDEX idx_l2 ON knn USING hnsw (v vector_l2_ops);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO knn VALUES (11, '[1.5,1,2]', '[1.5,1,2]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("11")],
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE knn SET v = '[100,100,100]', h = '[100,100,100]' WHERE id = 11;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' DESC LIMIT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM knn WHERE id = 11;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "knn across branches, merges, and AS OF",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE knn (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL);",
                r#"INSERT INTO knn VALUES
			(1, '[1.2,0.3,2.1]', '[1.2,0.3,2.1]'),
			(2, '[2.5,1.1,0.4]', '[2.5,1.1,0.4]'),
			(3, '[0.9,2.2,1.7]', '[0.9,2.2,1.7]'),
			(4, '[-1.4,2.8,0.6]', '[-1.4,2.8,0.6]'),
			(5, '[3.1,-0.7,1.9]', '[3.1,-0.7,1.9]'),
			(6, '[0.2,0.8,-1.3]', '[0.2,0.8,-1.3]'),
			(7, '[4.6,3.2,2.4]', '[4.6,3.2,2.4]'),
			(8, '[-2.2,-1.1,3.3]', '[-2.2,-1.1,3.3]'),
			(9, '[1.8,1.6,0.9]', '[1.8,1.6,0.9]'),
			(10, '[0.4,-2.6,2.2]', '[0.4,-2.6,2.2]');"#,
                "CREATE INDEX idx_l2 ON knn USING hnsw (v vector_l2_ops);",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_branch('other');",
                "INSERT INTO knn VALUES (11, '[1.5,1,2]', '[1.5,1,2]');",
                "SELECT dolt_commit('-am', 'main adds 11');",
                "SELECT dolt_checkout('other');",
                "INSERT INTO knn VALUES (12, '[1.4,1,2]', '[1.4,1,2]');",
                "SELECT dolt_commit('-am', 'other adds 12');",
                "SELECT dolt_checkout('main');",
                "SELECT dolt_merge('other');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("11")],
                            &[T("12")],
                            &[T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM knn AS OF 'HEAD~2' ORDER BY v <-> '[1.5,1,2]' LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("9")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM knn AS OF 'other' ORDER BY v <-> '[1.5,1,2]' LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("12")],
                            &[T("1")],
                            &[T("9")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nullable vector columns",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE nulls (id INT4 PRIMARY KEY, v vector(3));",
                "INSERT INTO nulls VALUES (1, '[1.2,0.3,2.1]'), (2, NULL), (3, '[0.9,2.2,1.7]'), (4, NULL), (5, '[3.1,-0.7,1.9]');",
                "CREATE INDEX idx_l2 ON nulls USING hnsw (v vector_l2_ops);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("5")],
                            &[T("2")],
                            &[T("4")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]', id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("5")],
                            &[T("2")],
                            &[T("4")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE nulls SET v = '[1.5,1,2]' WHERE id = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("1")],
                            &[T("3")],
                            &[T("5")],
                            &[T("4")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE nulls SET v = NULL WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                            &[T("5")],
                            &[T("4")],
                            &[T("1")],
                        ],
                        tag: "SELECT 5",
                    },
                    skip: Some("rows tied at a NULL distance follow Postgres' physical row order"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM nulls WHERE id = 4;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                            &[T("5")],
                            &[T("1")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nullable vector columns across branches, merges, and AS OF",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE nulls (id INT4 PRIMARY KEY, v vector(3));",
                "INSERT INTO nulls VALUES (1, '[1.2,0.3,2.1]'), (2, NULL), (3, '[0.9,2.2,1.7]'), (4, NULL), (5, '[3.1,-0.7,1.9]');",
                "CREATE INDEX idx_l2 ON nulls USING hnsw (v vector_l2_ops);",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_branch('other');",
                "UPDATE nulls SET v = '[1.5,1,2]' WHERE id = 2;",
                "SELECT dolt_commit('-am', 'main fills 2');",
                "SELECT dolt_checkout('other');",
                "UPDATE nulls SET v = NULL WHERE id = 5;",
                "SELECT dolt_commit('-am', 'other nulls 5');",
                "SELECT dolt_checkout('main');",
                "SELECT dolt_merge('other');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]', id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("1")],
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM nulls AS OF 'HEAD~2' ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM nulls AS OF 'other' ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nullable halfvec columns",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE hnulls (id INT4 PRIMARY KEY, h halfvec(3));",
                "INSERT INTO hnulls VALUES (1, '[1.2,0.3,2.1]'), (2, NULL), (3, '[0.9,2.2,1.7]'), (4, NULL), (5, '[3.1,-0.7,1.9]');",
                "CREATE INDEX idx_l2 ON hnulls USING hnsw (h halfvec_l2_ops);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM hnulls ORDER BY h <-> '[1.5,1,2]' LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("5")],
                            &[T("2")],
                            &[T("4")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE hnulls SET h = '[1.5,1,2]' WHERE id = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM hnulls ORDER BY h <-> '[1.5,1,2]' LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("1")],
                            &[T("3")],
                            &[T("5")],
                            &[T("4")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE hnulls SET h = NULL WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM hnulls ORDER BY h <-> '[1.5,1,2]' LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                            &[T("5")],
                            &[T("4")],
                            &[T("1")],
                        ],
                        tag: "SELECT 5",
                    },
                    skip: Some("rows tied at a NULL distance follow Postgres' physical row order"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "recall sanity on a multi-thousand-row table",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE big (id INT4 PRIMARY KEY, v vector(8) NOT NULL);",
                "INSERT INTO big VALUES (1, '[1,3,7,11,13,17,19,23]'), (2, '[2,6,14,22,26,34,38,46]'), (3, '[3,9,21,33,39,1,7,19]'), (4, '[4,12,28,44,2,18,26,42]'), (5, '[5,15,35,5,15,35,45,15]'), (6, '[6,18,42,16,28,2,14,38]'), (7, '[7,21,49,27,41,19,33,11]'), (8, '[8,24,6,38,4,36,2,34]'), (9, '[9,27,13,49,17,3,21,7]'), (10, '[10,30,20,10,30,20,40,30]'), (11, '[11,33,27,21,43,37,9,3]'), (12, '[12,36,34,32,6,4,28,26]'), (13, '[13,39,41,43,19,21,47,49]'), (14, '[14,42,48,4,32,38,16,22]'), (15, '[15,45,5,15,45,5,35,45]'), (16, '[16,48,12,26,8,22,4,18]'), (17, '[17,1,19,37,21,39,23,41]'), (18, '[18,4,26,48,34,6,42,14]'), (19, '[19,7,33,9,47,23,11,37]'), (20, '[20,10,40,20,10,40,30,10]'), (21, '[21,13,47,31,23,7,49,33]'), (22, '[22,16,4,42,36,24,18,6]'), (23, '[23,19,11,3,49,41,37,29]'), (24, '[24,22,18,14,12,8,6,2]'), (25, '[25,25,25,25,25,25,25,25]'), (26, '[26,28,32,36,38,42,44,48]'), (27, '[27,31,39,47,1,9,13,21]'), (28, '[28,34,46,8,14,26,32,44]'), (29, '[29,37,3,19,27,43,1,17]'), (30, '[30,40,10,30,40,10,20,40]'), (31, '[31,43,17,41,3,27,39,13]'), (32, '[32,46,24,2,16,44,8,36]'), (33, '[33,49,31,13,29,11,27,9]'), (34, '[34,2,38,24,42,28,46,32]'), (35, '[35,5,45,35,5,45,15,5]'), (36, '[36,8,2,46,18,12,34,28]'), (37, '[37,11,9,7,31,29,3,1]'), (38, '[38,14,16,18,44,46,22,24]'), (39, '[39,17,23,29,7,13,41,47]'), (40, '[40,20,30,40,20,30,10,20]'), (41, '[41,23,37,1,33,47,29,43]'), (42, '[42,26,44,12,46,14,48,16]'), (43, '[43,29,1,23,9,31,17,39]'), (44, '[44,32,8,34,22,48,36,12]'), (45, '[45,35,15,45,35,15,5,35]'), (46, '[46,38,22,6,48,32,24,8]'), (47, '[47,41,29,17,11,49,43,31]'), (48, '[48,44,36,28,24,16,12,4]'), (49, '[49,47,43,39,37,33,31,27]'), (50, '[50,0,0,0,0,0,0,0]'), (51, '[51,3,7,11,13,17,19,23]'), (52, '[52,6,14,22,26,34,38,46]'), (53, '[53,9,21,33,39,1,7,19]'), (54, '[54,12,28,44,2,18,26,42]'), (55, '[55,15,35,5,15,35,45,15]'), (56, '[56,18,42,16,28,2,14,38]'), (57, '[57,21,49,27,41,19,33,11]'), (58, '[58,24,6,38,4,36,2,34]'), (59, '[59,27,13,49,17,3,21,7]'), (60, '[60,30,20,10,30,20,40,30]'), (61, '[61,33,27,21,43,37,9,3]'), (62, '[62,36,34,32,6,4,28,26]'), (63, '[63,39,41,43,19,21,47,49]'), (64, '[64,42,48,4,32,38,16,22]'), (65, '[65,45,5,15,45,5,35,45]'), (66, '[66,48,12,26,8,22,4,18]'), (67, '[67,1,19,37,21,39,23,41]'), (68, '[68,4,26,48,34,6,42,14]'), (69, '[69,7,33,9,47,23,11,37]'), (70, '[70,10,40,20,10,40,30,10]'), (71, '[71,13,47,31,23,7,49,33]'), (72, '[72,16,4,42,36,24,18,6]'), (73, '[73,19,11,3,49,41,37,29]'), (74, '[74,22,18,14,12,8,6,2]'), (75, '[75,25,25,25,25,25,25,25]'), (76, '[76,28,32,36,38,42,44,48]'), (77, '[77,31,39,47,1,9,13,21]'), (78, '[78,34,46,8,14,26,32,44]'), (79, '[79,37,3,19,27,43,1,17]'), (80, '[80,40,10,30,40,10,20,40]'), (81, '[81,43,17,41,3,27,39,13]'), (82, '[82,46,24,2,16,44,8,36]'), (83, '[83,49,31,13,29,11,27,9]'), (84, '[84,2,38,24,42,28,46,32]'), (85, '[85,5,45,35,5,45,15,5]'), (86, '[86,8,2,46,18,12,34,28]'), (87, '[87,11,9,7,31,29,3,1]'), (88, '[88,14,16,18,44,46,22,24]'), (89, '[89,17,23,29,7,13,41,47]'), (90, '[90,20,30,40,20,30,10,20]'), (91, '[91,23,37,1,33,47,29,43]'), (92, '[92,26,44,12,46,14,48,16]'), (93, '[93,29,1,23,9,31,17,39]'), (94, '[94,32,8,34,22,48,36,12]'), (95, '[95,35,15,45,35,15,5,35]'), (96, '[96,38,22,6,48,32,24,8]'), (97, '[97,41,29,17,11,49,43,31]'), (98, '[98,44,36,28,24,16,12,4]'), (99, '[99,47,43,39,37,33,31,27]'), (100, '[100,0,0,0,0,0,0,0]'), (101, '[101,3,7,11,13,17,19,23]'), (102, '[102,6,14,22,26,34,38,46]'), (103, '[103,9,21,33,39,1,7,19]'), (104, '[104,12,28,44,2,18,26,42]'), (105, '[105,15,35,5,15,35,45,15]'), (106, '[106,18,42,16,28,2,14,38]'), (107, '[107,21,49,27,41,19,33,11]'), (108, '[108,24,6,38,4,36,2,34]'), (109, '[109,27,13,49,17,3,21,7]'), (110, '[110,30,20,10,30,20,40,30]'), (111, '[111,33,27,21,43,37,9,3]'), (112, '[112,36,34,32,6,4,28,26]'), (113, '[113,39,41,43,19,21,47,49]'), (114, '[114,42,48,4,32,38,16,22]'), (115, '[115,45,5,15,45,5,35,45]'), (116, '[116,48,12,26,8,22,4,18]'), (117, '[117,1,19,37,21,39,23,41]'), (118, '[118,4,26,48,34,6,42,14]'), (119, '[119,7,33,9,47,23,11,37]'), (120, '[120,10,40,20,10,40,30,10]'), (121, '[121,13,47,31,23,7,49,33]'), (122, '[122,16,4,42,36,24,18,6]'), (123, '[123,19,11,3,49,41,37,29]'), (124, '[124,22,18,14,12,8,6,2]'), (125, '[125,25,25,25,25,25,25,25]'), (126, '[126,28,32,36,38,42,44,48]'), (127, '[127,31,39,47,1,9,13,21]'), (128, '[128,34,46,8,14,26,32,44]'), (129, '[129,37,3,19,27,43,1,17]'), (130, '[130,40,10,30,40,10,20,40]'), (131, '[131,43,17,41,3,27,39,13]'), (132, '[132,46,24,2,16,44,8,36]'), (133, '[133,49,31,13,29,11,27,9]'), (134, '[134,2,38,24,42,28,46,32]'), (135, '[135,5,45,35,5,45,15,5]'), (136, '[136,8,2,46,18,12,34,28]'), (137, '[137,11,9,7,31,29,3,1]'), (138, '[138,14,16,18,44,46,22,24]'), (139, '[139,17,23,29,7,13,41,47]'), (140, '[140,20,30,40,20,30,10,20]'), (141, '[141,23,37,1,33,47,29,43]'), (142, '[142,26,44,12,46,14,48,16]'), (143, '[143,29,1,23,9,31,17,39]'), (144, '[144,32,8,34,22,48,36,12]'), (145, '[145,35,15,45,35,15,5,35]'), (146, '[146,38,22,6,48,32,24,8]'), (147, '[147,41,29,17,11,49,43,31]'), (148, '[148,44,36,28,24,16,12,4]'), (149, '[149,47,43,39,37,33,31,27]'), (150, '[150,0,0,0,0,0,0,0]'), (151, '[151,3,7,11,13,17,19,23]'), (152, '[152,6,14,22,26,34,38,46]'), (153, '[153,9,21,33,39,1,7,19]'), (154, '[154,12,28,44,2,18,26,42]'), (155, '[155,15,35,5,15,35,45,15]'), (156, '[156,18,42,16,28,2,14,38]'), (157, '[157,21,49,27,41,19,33,11]'), (158, '[158,24,6,38,4,36,2,34]'), (159, '[159,27,13,49,17,3,21,7]'), (160, '[160,30,20,10,30,20,40,30]'), (161, '[161,33,27,21,43,37,9,3]'), (162, '[162,36,34,32,6,4,28,26]'), (163, '[163,39,41,43,19,21,47,49]'), (164, '[164,42,48,4,32,38,16,22]'), (165, '[165,45,5,15,45,5,35,45]'), (166, '[166,48,12,26,8,22,4,18]'), (167, '[167,1,19,37,21,39,23,41]'), (168, '[168,4,26,48,34,6,42,14]'), (169, '[169,7,33,9,47,23,11,37]'), (170, '[170,10,40,20,10,40,30,10]'), (171, '[171,13,47,31,23,7,49,33]'), (172, '[172,16,4,42,36,24,18,6]'), (173, '[173,19,11,3,49,41,37,29]'), (174, '[174,22,18,14,12,8,6,2]'), (175, '[175,25,25,25,25,25,25,25]'), (176, '[176,28,32,36,38,42,44,48]'), (177, '[177,31,39,47,1,9,13,21]'), (178, '[178,34,46,8,14,26,32,44]'), (179, '[179,37,3,19,27,43,1,17]'), (180, '[180,40,10,30,40,10,20,40]'), (181, '[181,43,17,41,3,27,39,13]'), (182, '[182,46,24,2,16,44,8,36]'), (183, '[183,49,31,13,29,11,27,9]'), (184, '[184,2,38,24,42,28,46,32]'), (185, '[185,5,45,35,5,45,15,5]'), (186, '[186,8,2,46,18,12,34,28]'), (187, '[187,11,9,7,31,29,3,1]'), (188, '[188,14,16,18,44,46,22,24]'), (189, '[189,17,23,29,7,13,41,47]'), (190, '[190,20,30,40,20,30,10,20]'), (191, '[191,23,37,1,33,47,29,43]'), (192, '[192,26,44,12,46,14,48,16]'), (193, '[193,29,1,23,9,31,17,39]'), (194, '[194,32,8,34,22,48,36,12]'), (195, '[195,35,15,45,35,15,5,35]'), (196, '[196,38,22,6,48,32,24,8]'), (197, '[197,41,29,17,11,49,43,31]'), (198, '[198,44,36,28,24,16,12,4]'), (199, '[199,47,43,39,37,33,31,27]'), (200, '[200,0,0,0,0,0,0,0]'), (201, '[201,3,7,11,13,17,19,23]'), (202, '[202,6,14,22,26,34,38,46]'), (203, '[203,9,21,33,39,1,7,19]'), (204, '[204,12,28,44,2,18,26,42]'), (205, '[205,15,35,5,15,35,45,15]'), (206, '[206,18,42,16,28,2,14,38]'), (207, '[207,21,49,27,41,19,33,11]'), (208, '[208,24,6,38,4,36,2,34]'), (209, '[209,27,13,49,17,3,21,7]'), (210, '[210,30,20,10,30,20,40,30]'), (211, '[211,33,27,21,43,37,9,3]'), (212, '[212,36,34,32,6,4,28,26]'), (213, '[213,39,41,43,19,21,47,49]'), (214, '[214,42,48,4,32,38,16,22]'), (215, '[215,45,5,15,45,5,35,45]'), (216, '[216,48,12,26,8,22,4,18]'), (217, '[217,1,19,37,21,39,23,41]'), (218, '[218,4,26,48,34,6,42,14]'), (219, '[219,7,33,9,47,23,11,37]'), (220, '[220,10,40,20,10,40,30,10]'), (221, '[221,13,47,31,23,7,49,33]'), (222, '[222,16,4,42,36,24,18,6]'), (223, '[223,19,11,3,49,41,37,29]'), (224, '[224,22,18,14,12,8,6,2]'), (225, '[225,25,25,25,25,25,25,25]'), (226, '[226,28,32,36,38,42,44,48]'), (227, '[227,31,39,47,1,9,13,21]'), (228, '[228,34,46,8,14,26,32,44]'), (229, '[229,37,3,19,27,43,1,17]'), (230, '[230,40,10,30,40,10,20,40]'), (231, '[231,43,17,41,3,27,39,13]'), (232, '[232,46,24,2,16,44,8,36]'), (233, '[233,49,31,13,29,11,27,9]'), (234, '[234,2,38,24,42,28,46,32]'), (235, '[235,5,45,35,5,45,15,5]'), (236, '[236,8,2,46,18,12,34,28]'), (237, '[237,11,9,7,31,29,3,1]'), (238, '[238,14,16,18,44,46,22,24]'), (239, '[239,17,23,29,7,13,41,47]'), (240, '[240,20,30,40,20,30,10,20]'), (241, '[241,23,37,1,33,47,29,43]'), (242, '[242,26,44,12,46,14,48,16]'), (243, '[243,29,1,23,9,31,17,39]'), (244, '[244,32,8,34,22,48,36,12]'), (245, '[245,35,15,45,35,15,5,35]'), (246, '[246,38,22,6,48,32,24,8]'), (247, '[247,41,29,17,11,49,43,31]'), (248, '[248,44,36,28,24,16,12,4]'), (249, '[249,47,43,39,37,33,31,27]'), (250, '[250,0,0,0,0,0,0,0]'), (251, '[251,3,7,11,13,17,19,23]'), (252, '[252,6,14,22,26,34,38,46]'), (253, '[253,9,21,33,39,1,7,19]'), (254, '[254,12,28,44,2,18,26,42]'), (255, '[255,15,35,5,15,35,45,15]'), (256, '[256,18,42,16,28,2,14,38]'), (257, '[257,21,49,27,41,19,33,11]'), (258, '[258,24,6,38,4,36,2,34]'), (259, '[259,27,13,49,17,3,21,7]'), (260, '[260,30,20,10,30,20,40,30]'), (261, '[261,33,27,21,43,37,9,3]'), (262, '[262,36,34,32,6,4,28,26]'), (263, '[263,39,41,43,19,21,47,49]'), (264, '[264,42,48,4,32,38,16,22]'), (265, '[265,45,5,15,45,5,35,45]'), (266, '[266,48,12,26,8,22,4,18]'), (267, '[267,1,19,37,21,39,23,41]'), (268, '[268,4,26,48,34,6,42,14]'), (269, '[269,7,33,9,47,23,11,37]'), (270, '[270,10,40,20,10,40,30,10]'), (271, '[271,13,47,31,23,7,49,33]'), (272, '[272,16,4,42,36,24,18,6]'), (273, '[273,19,11,3,49,41,37,29]'), (274, '[274,22,18,14,12,8,6,2]'), (275, '[275,25,25,25,25,25,25,25]'), (276, '[276,28,32,36,38,42,44,48]'), (277, '[277,31,39,47,1,9,13,21]'), (278, '[278,34,46,8,14,26,32,44]'), (279, '[279,37,3,19,27,43,1,17]'), (280, '[280,40,10,30,40,10,20,40]'), (281, '[281,43,17,41,3,27,39,13]'), (282, '[282,46,24,2,16,44,8,36]'), (283, '[283,49,31,13,29,11,27,9]'), (284, '[284,2,38,24,42,28,46,32]'), (285, '[285,5,45,35,5,45,15,5]'), (286, '[286,8,2,46,18,12,34,28]'), (287, '[287,11,9,7,31,29,3,1]'), (288, '[288,14,16,18,44,46,22,24]'), (289, '[289,17,23,29,7,13,41,47]'), (290, '[290,20,30,40,20,30,10,20]'), (291, '[291,23,37,1,33,47,29,43]'), (292, '[292,26,44,12,46,14,48,16]'), (293, '[293,29,1,23,9,31,17,39]'), (294, '[294,32,8,34,22,48,36,12]'), (295, '[295,35,15,45,35,15,5,35]'), (296, '[296,38,22,6,48,32,24,8]'), (297, '[297,41,29,17,11,49,43,31]'), (298, '[298,44,36,28,24,16,12,4]'), (299, '[299,47,43,39,37,33,31,27]'), (300, '[300,0,0,0,0,0,0,0]'), (301, '[301,3,7,11,13,17,19,23]'), (302, '[302,6,14,22,26,34,38,46]'), (303, '[303,9,21,33,39,1,7,19]'), (304, '[304,12,28,44,2,18,26,42]'), (305, '[305,15,35,5,15,35,45,15]'), (306, '[306,18,42,16,28,2,14,38]'), (307, '[307,21,49,27,41,19,33,11]'), (308, '[308,24,6,38,4,36,2,34]'), (309, '[309,27,13,49,17,3,21,7]'), (310, '[310,30,20,10,30,20,40,30]'), (311, '[311,33,27,21,43,37,9,3]'), (312, '[312,36,34,32,6,4,28,26]'), (313, '[313,39,41,43,19,21,47,49]'), (314, '[314,42,48,4,32,38,16,22]'), (315, '[315,45,5,15,45,5,35,45]'), (316, '[316,48,12,26,8,22,4,18]'), (317, '[317,1,19,37,21,39,23,41]'), (318, '[318,4,26,48,34,6,42,14]'), (319, '[319,7,33,9,47,23,11,37]'), (320, '[320,10,40,20,10,40,30,10]'), (321, '[321,13,47,31,23,7,49,33]'), (322, '[322,16,4,42,36,24,18,6]'), (323, '[323,19,11,3,49,41,37,29]'), (324, '[324,22,18,14,12,8,6,2]'), (325, '[325,25,25,25,25,25,25,25]'), (326, '[326,28,32,36,38,42,44,48]'), (327, '[327,31,39,47,1,9,13,21]'), (328, '[328,34,46,8,14,26,32,44]'), (329, '[329,37,3,19,27,43,1,17]'), (330, '[330,40,10,30,40,10,20,40]'), (331, '[331,43,17,41,3,27,39,13]'), (332, '[332,46,24,2,16,44,8,36]'), (333, '[333,49,31,13,29,11,27,9]'), (334, '[334,2,38,24,42,28,46,32]'), (335, '[335,5,45,35,5,45,15,5]'), (336, '[336,8,2,46,18,12,34,28]'), (337, '[337,11,9,7,31,29,3,1]'), (338, '[338,14,16,18,44,46,22,24]'), (339, '[339,17,23,29,7,13,41,47]'), (340, '[340,20,30,40,20,30,10,20]'), (341, '[341,23,37,1,33,47,29,43]'), (342, '[342,26,44,12,46,14,48,16]'), (343, '[343,29,1,23,9,31,17,39]'), (344, '[344,32,8,34,22,48,36,12]'), (345, '[345,35,15,45,35,15,5,35]'), (346, '[346,38,22,6,48,32,24,8]'), (347, '[347,41,29,17,11,49,43,31]'), (348, '[348,44,36,28,24,16,12,4]'), (349, '[349,47,43,39,37,33,31,27]'), (350, '[350,0,0,0,0,0,0,0]'), (351, '[351,3,7,11,13,17,19,23]'), (352, '[352,6,14,22,26,34,38,46]'), (353, '[353,9,21,33,39,1,7,19]'), (354, '[354,12,28,44,2,18,26,42]'), (355, '[355,15,35,5,15,35,45,15]'), (356, '[356,18,42,16,28,2,14,38]'), (357, '[357,21,49,27,41,19,33,11]'), (358, '[358,24,6,38,4,36,2,34]'), (359, '[359,27,13,49,17,3,21,7]'), (360, '[360,30,20,10,30,20,40,30]'), (361, '[361,33,27,21,43,37,9,3]'), (362, '[362,36,34,32,6,4,28,26]'), (363, '[363,39,41,43,19,21,47,49]'), (364, '[364,42,48,4,32,38,16,22]'), (365, '[365,45,5,15,45,5,35,45]'), (366, '[366,48,12,26,8,22,4,18]'), (367, '[367,1,19,37,21,39,23,41]'), (368, '[368,4,26,48,34,6,42,14]'), (369, '[369,7,33,9,47,23,11,37]'), (370, '[370,10,40,20,10,40,30,10]'), (371, '[371,13,47,31,23,7,49,33]'), (372, '[372,16,4,42,36,24,18,6]'), (373, '[373,19,11,3,49,41,37,29]'), (374, '[374,22,18,14,12,8,6,2]'), (375, '[375,25,25,25,25,25,25,25]'), (376, '[376,28,32,36,38,42,44,48]'), (377, '[377,31,39,47,1,9,13,21]'), (378, '[378,34,46,8,14,26,32,44]'), (379, '[379,37,3,19,27,43,1,17]'), (380, '[380,40,10,30,40,10,20,40]'), (381, '[381,43,17,41,3,27,39,13]'), (382, '[382,46,24,2,16,44,8,36]'), (383, '[383,49,31,13,29,11,27,9]'), (384, '[384,2,38,24,42,28,46,32]'), (385, '[385,5,45,35,5,45,15,5]'), (386, '[386,8,2,46,18,12,34,28]'), (387, '[387,11,9,7,31,29,3,1]'), (388, '[388,14,16,18,44,46,22,24]'), (389, '[389,17,23,29,7,13,41,47]'), (390, '[390,20,30,40,20,30,10,20]'), (391, '[391,23,37,1,33,47,29,43]'), (392, '[392,26,44,12,46,14,48,16]'), (393, '[393,29,1,23,9,31,17,39]'), (394, '[394,32,8,34,22,48,36,12]'), (395, '[395,35,15,45,35,15,5,35]'), (396, '[396,38,22,6,48,32,24,8]'), (397, '[397,41,29,17,11,49,43,31]'), (398, '[398,44,36,28,24,16,12,4]'), (399, '[399,47,43,39,37,33,31,27]'), (400, '[400,0,0,0,0,0,0,0]'), (401, '[401,3,7,11,13,17,19,23]'), (402, '[402,6,14,22,26,34,38,46]'), (403, '[403,9,21,33,39,1,7,19]'), (404, '[404,12,28,44,2,18,26,42]'), (405, '[405,15,35,5,15,35,45,15]'), (406, '[406,18,42,16,28,2,14,38]'), (407, '[407,21,49,27,41,19,33,11]'), (408, '[408,24,6,38,4,36,2,34]'), (409, '[409,27,13,49,17,3,21,7]'), (410, '[410,30,20,10,30,20,40,30]'), (411, '[411,33,27,21,43,37,9,3]'), (412, '[412,36,34,32,6,4,28,26]'), (413, '[413,39,41,43,19,21,47,49]'), (414, '[414,42,48,4,32,38,16,22]'), (415, '[415,45,5,15,45,5,35,45]'), (416, '[416,48,12,26,8,22,4,18]'), (417, '[417,1,19,37,21,39,23,41]'), (418, '[418,4,26,48,34,6,42,14]'), (419, '[419,7,33,9,47,23,11,37]'), (420, '[420,10,40,20,10,40,30,10]'), (421, '[421,13,47,31,23,7,49,33]'), (422, '[422,16,4,42,36,24,18,6]'), (423, '[423,19,11,3,49,41,37,29]'), (424, '[424,22,18,14,12,8,6,2]'), (425, '[425,25,25,25,25,25,25,25]'), (426, '[426,28,32,36,38,42,44,48]'), (427, '[427,31,39,47,1,9,13,21]'), (428, '[428,34,46,8,14,26,32,44]'), (429, '[429,37,3,19,27,43,1,17]'), (430, '[430,40,10,30,40,10,20,40]'), (431, '[431,43,17,41,3,27,39,13]'), (432, '[432,46,24,2,16,44,8,36]'), (433, '[433,49,31,13,29,11,27,9]'), (434, '[434,2,38,24,42,28,46,32]'), (435, '[435,5,45,35,5,45,15,5]'), (436, '[436,8,2,46,18,12,34,28]'), (437, '[437,11,9,7,31,29,3,1]'), (438, '[438,14,16,18,44,46,22,24]'), (439, '[439,17,23,29,7,13,41,47]'), (440, '[440,20,30,40,20,30,10,20]'), (441, '[441,23,37,1,33,47,29,43]'), (442, '[442,26,44,12,46,14,48,16]'), (443, '[443,29,1,23,9,31,17,39]'), (444, '[444,32,8,34,22,48,36,12]'), (445, '[445,35,15,45,35,15,5,35]'), (446, '[446,38,22,6,48,32,24,8]'), (447, '[447,41,29,17,11,49,43,31]'), (448, '[448,44,36,28,24,16,12,4]'), (449, '[449,47,43,39,37,33,31,27]'), (450, '[450,0,0,0,0,0,0,0]'), (451, '[451,3,7,11,13,17,19,23]'), (452, '[452,6,14,22,26,34,38,46]'), (453, '[453,9,21,33,39,1,7,19]'), (454, '[454,12,28,44,2,18,26,42]'), (455, '[455,15,35,5,15,35,45,15]'), (456, '[456,18,42,16,28,2,14,38]'), (457, '[457,21,49,27,41,19,33,11]'), (458, '[458,24,6,38,4,36,2,34]'), (459, '[459,27,13,49,17,3,21,7]'), (460, '[460,30,20,10,30,20,40,30]'), (461, '[461,33,27,21,43,37,9,3]'), (462, '[462,36,34,32,6,4,28,26]'), (463, '[463,39,41,43,19,21,47,49]'), (464, '[464,42,48,4,32,38,16,22]'), (465, '[465,45,5,15,45,5,35,45]'), (466, '[466,48,12,26,8,22,4,18]'), (467, '[467,1,19,37,21,39,23,41]'), (468, '[468,4,26,48,34,6,42,14]'), (469, '[469,7,33,9,47,23,11,37]'), (470, '[470,10,40,20,10,40,30,10]'), (471, '[471,13,47,31,23,7,49,33]'), (472, '[472,16,4,42,36,24,18,6]'), (473, '[473,19,11,3,49,41,37,29]'), (474, '[474,22,18,14,12,8,6,2]'), (475, '[475,25,25,25,25,25,25,25]'), (476, '[476,28,32,36,38,42,44,48]'), (477, '[477,31,39,47,1,9,13,21]'), (478, '[478,34,46,8,14,26,32,44]'), (479, '[479,37,3,19,27,43,1,17]'), (480, '[480,40,10,30,40,10,20,40]'), (481, '[481,43,17,41,3,27,39,13]'), (482, '[482,46,24,2,16,44,8,36]'), (483, '[483,49,31,13,29,11,27,9]'), (484, '[484,2,38,24,42,28,46,32]'), (485, '[485,5,45,35,5,45,15,5]'), (486, '[486,8,2,46,18,12,34,28]'), (487, '[487,11,9,7,31,29,3,1]'), (488, '[488,14,16,18,44,46,22,24]'), (489, '[489,17,23,29,7,13,41,47]'), (490, '[490,20,30,40,20,30,10,20]'), (491, '[491,23,37,1,33,47,29,43]'), (492, '[492,26,44,12,46,14,48,16]'), (493, '[493,29,1,23,9,31,17,39]'), (494, '[494,32,8,34,22,48,36,12]'), (495, '[495,35,15,45,35,15,5,35]'), (496, '[496,38,22,6,48,32,24,8]'), (497, '[497,41,29,17,11,49,43,31]'), (498, '[498,44,36,28,24,16,12,4]'), (499, '[499,47,43,39,37,33,31,27]'), (500, '[500,0,0,0,0,0,0,0]'), (501, '[501,3,7,11,13,17,19,23]'), (502, '[502,6,14,22,26,34,38,46]'), (503, '[503,9,21,33,39,1,7,19]'), (504, '[504,12,28,44,2,18,26,42]'), (505, '[505,15,35,5,15,35,45,15]'), (506, '[506,18,42,16,28,2,14,38]'), (507, '[507,21,49,27,41,19,33,11]'), (508, '[508,24,6,38,4,36,2,34]'), (509, '[509,27,13,49,17,3,21,7]'), (510, '[510,30,20,10,30,20,40,30]'), (511, '[511,33,27,21,43,37,9,3]'), (512, '[512,36,34,32,6,4,28,26]'), (513, '[513,39,41,43,19,21,47,49]'), (514, '[514,42,48,4,32,38,16,22]'), (515, '[515,45,5,15,45,5,35,45]'), (516, '[516,48,12,26,8,22,4,18]'), (517, '[517,1,19,37,21,39,23,41]'), (518, '[518,4,26,48,34,6,42,14]'), (519, '[519,7,33,9,47,23,11,37]'), (520, '[520,10,40,20,10,40,30,10]'), (521, '[521,13,47,31,23,7,49,33]'), (522, '[522,16,4,42,36,24,18,6]'), (523, '[523,19,11,3,49,41,37,29]'), (524, '[524,22,18,14,12,8,6,2]'), (525, '[525,25,25,25,25,25,25,25]'), (526, '[526,28,32,36,38,42,44,48]'), (527, '[527,31,39,47,1,9,13,21]'), (528, '[528,34,46,8,14,26,32,44]'), (529, '[529,37,3,19,27,43,1,17]'), (530, '[530,40,10,30,40,10,20,40]'), (531, '[531,43,17,41,3,27,39,13]'), (532, '[532,46,24,2,16,44,8,36]'), (533, '[533,49,31,13,29,11,27,9]'), (534, '[534,2,38,24,42,28,46,32]'), (535, '[535,5,45,35,5,45,15,5]'), (536, '[536,8,2,46,18,12,34,28]'), (537, '[537,11,9,7,31,29,3,1]'), (538, '[538,14,16,18,44,46,22,24]'), (539, '[539,17,23,29,7,13,41,47]'), (540, '[540,20,30,40,20,30,10,20]'), (541, '[541,23,37,1,33,47,29,43]'), (542, '[542,26,44,12,46,14,48,16]'), (543, '[543,29,1,23,9,31,17,39]'), (544, '[544,32,8,34,22,48,36,12]'), (545, '[545,35,15,45,35,15,5,35]'), (546, '[546,38,22,6,48,32,24,8]'), (547, '[547,41,29,17,11,49,43,31]'), (548, '[548,44,36,28,24,16,12,4]'), (549, '[549,47,43,39,37,33,31,27]'), (550, '[550,0,0,0,0,0,0,0]'), (551, '[551,3,7,11,13,17,19,23]'), (552, '[552,6,14,22,26,34,38,46]'), (553, '[553,9,21,33,39,1,7,19]'), (554, '[554,12,28,44,2,18,26,42]'), (555, '[555,15,35,5,15,35,45,15]'), (556, '[556,18,42,16,28,2,14,38]'), (557, '[557,21,49,27,41,19,33,11]'), (558, '[558,24,6,38,4,36,2,34]'), (559, '[559,27,13,49,17,3,21,7]'), (560, '[560,30,20,10,30,20,40,30]'), (561, '[561,33,27,21,43,37,9,3]'), (562, '[562,36,34,32,6,4,28,26]'), (563, '[563,39,41,43,19,21,47,49]'), (564, '[564,42,48,4,32,38,16,22]'), (565, '[565,45,5,15,45,5,35,45]'), (566, '[566,48,12,26,8,22,4,18]'), (567, '[567,1,19,37,21,39,23,41]'), (568, '[568,4,26,48,34,6,42,14]'), (569, '[569,7,33,9,47,23,11,37]'), (570, '[570,10,40,20,10,40,30,10]'), (571, '[571,13,47,31,23,7,49,33]'), (572, '[572,16,4,42,36,24,18,6]'), (573, '[573,19,11,3,49,41,37,29]'), (574, '[574,22,18,14,12,8,6,2]'), (575, '[575,25,25,25,25,25,25,25]'), (576, '[576,28,32,36,38,42,44,48]'), (577, '[577,31,39,47,1,9,13,21]'), (578, '[578,34,46,8,14,26,32,44]'), (579, '[579,37,3,19,27,43,1,17]'), (580, '[580,40,10,30,40,10,20,40]'), (581, '[581,43,17,41,3,27,39,13]'), (582, '[582,46,24,2,16,44,8,36]'), (583, '[583,49,31,13,29,11,27,9]'), (584, '[584,2,38,24,42,28,46,32]'), (585, '[585,5,45,35,5,45,15,5]'), (586, '[586,8,2,46,18,12,34,28]'), (587, '[587,11,9,7,31,29,3,1]'), (588, '[588,14,16,18,44,46,22,24]'), (589, '[589,17,23,29,7,13,41,47]'), (590, '[590,20,30,40,20,30,10,20]'), (591, '[591,23,37,1,33,47,29,43]'), (592, '[592,26,44,12,46,14,48,16]'), (593, '[593,29,1,23,9,31,17,39]'), (594, '[594,32,8,34,22,48,36,12]'), (595, '[595,35,15,45,35,15,5,35]'), (596, '[596,38,22,6,48,32,24,8]'), (597, '[597,41,29,17,11,49,43,31]'), (598, '[598,44,36,28,24,16,12,4]'), (599, '[599,47,43,39,37,33,31,27]'), (600, '[600,0,0,0,0,0,0,0]'), (601, '[601,3,7,11,13,17,19,23]'), (602, '[602,6,14,22,26,34,38,46]'), (603, '[603,9,21,33,39,1,7,19]'), (604, '[604,12,28,44,2,18,26,42]'), (605, '[605,15,35,5,15,35,45,15]'), (606, '[606,18,42,16,28,2,14,38]'), (607, '[607,21,49,27,41,19,33,11]'), (608, '[608,24,6,38,4,36,2,34]'), (609, '[609,27,13,49,17,3,21,7]'), (610, '[610,30,20,10,30,20,40,30]'), (611, '[611,33,27,21,43,37,9,3]'), (612, '[612,36,34,32,6,4,28,26]'), (613, '[613,39,41,43,19,21,47,49]'), (614, '[614,42,48,4,32,38,16,22]'), (615, '[615,45,5,15,45,5,35,45]'), (616, '[616,48,12,26,8,22,4,18]'), (617, '[617,1,19,37,21,39,23,41]'), (618, '[618,4,26,48,34,6,42,14]'), (619, '[619,7,33,9,47,23,11,37]'), (620, '[620,10,40,20,10,40,30,10]'), (621, '[621,13,47,31,23,7,49,33]'), (622, '[622,16,4,42,36,24,18,6]'), (623, '[623,19,11,3,49,41,37,29]'), (624, '[624,22,18,14,12,8,6,2]'), (625, '[625,25,25,25,25,25,25,25]'), (626, '[626,28,32,36,38,42,44,48]'), (627, '[627,31,39,47,1,9,13,21]'), (628, '[628,34,46,8,14,26,32,44]'), (629, '[629,37,3,19,27,43,1,17]'), (630, '[630,40,10,30,40,10,20,40]'), (631, '[631,43,17,41,3,27,39,13]'), (632, '[632,46,24,2,16,44,8,36]'), (633, '[633,49,31,13,29,11,27,9]'), (634, '[634,2,38,24,42,28,46,32]'), (635, '[635,5,45,35,5,45,15,5]'), (636, '[636,8,2,46,18,12,34,28]'), (637, '[637,11,9,7,31,29,3,1]'), (638, '[638,14,16,18,44,46,22,24]'), (639, '[639,17,23,29,7,13,41,47]'), (640, '[640,20,30,40,20,30,10,20]'), (641, '[641,23,37,1,33,47,29,43]'), (642, '[642,26,44,12,46,14,48,16]'), (643, '[643,29,1,23,9,31,17,39]'), (644, '[644,32,8,34,22,48,36,12]'), (645, '[645,35,15,45,35,15,5,35]'), (646, '[646,38,22,6,48,32,24,8]'), (647, '[647,41,29,17,11,49,43,31]'), (648, '[648,44,36,28,24,16,12,4]'), (649, '[649,47,43,39,37,33,31,27]'), (650, '[650,0,0,0,0,0,0,0]'), (651, '[651,3,7,11,13,17,19,23]'), (652, '[652,6,14,22,26,34,38,46]'), (653, '[653,9,21,33,39,1,7,19]'), (654, '[654,12,28,44,2,18,26,42]'), (655, '[655,15,35,5,15,35,45,15]'), (656, '[656,18,42,16,28,2,14,38]'), (657, '[657,21,49,27,41,19,33,11]'), (658, '[658,24,6,38,4,36,2,34]'), (659, '[659,27,13,49,17,3,21,7]'), (660, '[660,30,20,10,30,20,40,30]'), (661, '[661,33,27,21,43,37,9,3]'), (662, '[662,36,34,32,6,4,28,26]'), (663, '[663,39,41,43,19,21,47,49]'), (664, '[664,42,48,4,32,38,16,22]'), (665, '[665,45,5,15,45,5,35,45]'), (666, '[666,48,12,26,8,22,4,18]'), (667, '[667,1,19,37,21,39,23,41]'), (668, '[668,4,26,48,34,6,42,14]'), (669, '[669,7,33,9,47,23,11,37]'), (670, '[670,10,40,20,10,40,30,10]'), (671, '[671,13,47,31,23,7,49,33]'), (672, '[672,16,4,42,36,24,18,6]'), (673, '[673,19,11,3,49,41,37,29]'), (674, '[674,22,18,14,12,8,6,2]'), (675, '[675,25,25,25,25,25,25,25]'), (676, '[676,28,32,36,38,42,44,48]'), (677, '[677,31,39,47,1,9,13,21]'), (678, '[678,34,46,8,14,26,32,44]'), (679, '[679,37,3,19,27,43,1,17]'), (680, '[680,40,10,30,40,10,20,40]'), (681, '[681,43,17,41,3,27,39,13]'), (682, '[682,46,24,2,16,44,8,36]'), (683, '[683,49,31,13,29,11,27,9]'), (684, '[684,2,38,24,42,28,46,32]'), (685, '[685,5,45,35,5,45,15,5]'), (686, '[686,8,2,46,18,12,34,28]'), (687, '[687,11,9,7,31,29,3,1]'), (688, '[688,14,16,18,44,46,22,24]'), (689, '[689,17,23,29,7,13,41,47]'), (690, '[690,20,30,40,20,30,10,20]'), (691, '[691,23,37,1,33,47,29,43]'), (692, '[692,26,44,12,46,14,48,16]'), (693, '[693,29,1,23,9,31,17,39]'), (694, '[694,32,8,34,22,48,36,12]'), (695, '[695,35,15,45,35,15,5,35]'), (696, '[696,38,22,6,48,32,24,8]'), (697, '[697,41,29,17,11,49,43,31]'), (698, '[698,44,36,28,24,16,12,4]'), (699, '[699,47,43,39,37,33,31,27]'), (700, '[700,0,0,0,0,0,0,0]'), (701, '[701,3,7,11,13,17,19,23]'), (702, '[702,6,14,22,26,34,38,46]'), (703, '[703,9,21,33,39,1,7,19]'), (704, '[704,12,28,44,2,18,26,42]'), (705, '[705,15,35,5,15,35,45,15]'), (706, '[706,18,42,16,28,2,14,38]'), (707, '[707,21,49,27,41,19,33,11]'), (708, '[708,24,6,38,4,36,2,34]'), (709, '[709,27,13,49,17,3,21,7]'), (710, '[710,30,20,10,30,20,40,30]'), (711, '[711,33,27,21,43,37,9,3]'), (712, '[712,36,34,32,6,4,28,26]'), (713, '[713,39,41,43,19,21,47,49]'), (714, '[714,42,48,4,32,38,16,22]'), (715, '[715,45,5,15,45,5,35,45]'), (716, '[716,48,12,26,8,22,4,18]'), (717, '[717,1,19,37,21,39,23,41]'), (718, '[718,4,26,48,34,6,42,14]'), (719, '[719,7,33,9,47,23,11,37]'), (720, '[720,10,40,20,10,40,30,10]'), (721, '[721,13,47,31,23,7,49,33]'), (722, '[722,16,4,42,36,24,18,6]'), (723, '[723,19,11,3,49,41,37,29]'), (724, '[724,22,18,14,12,8,6,2]'), (725, '[725,25,25,25,25,25,25,25]'), (726, '[726,28,32,36,38,42,44,48]'), (727, '[727,31,39,47,1,9,13,21]'), (728, '[728,34,46,8,14,26,32,44]'), (729, '[729,37,3,19,27,43,1,17]'), (730, '[730,40,10,30,40,10,20,40]'), (731, '[731,43,17,41,3,27,39,13]'), (732, '[732,46,24,2,16,44,8,36]'), (733, '[733,49,31,13,29,11,27,9]'), (734, '[734,2,38,24,42,28,46,32]'), (735, '[735,5,45,35,5,45,15,5]'), (736, '[736,8,2,46,18,12,34,28]'), (737, '[737,11,9,7,31,29,3,1]'), (738, '[738,14,16,18,44,46,22,24]'), (739, '[739,17,23,29,7,13,41,47]'), (740, '[740,20,30,40,20,30,10,20]'), (741, '[741,23,37,1,33,47,29,43]'), (742, '[742,26,44,12,46,14,48,16]'), (743, '[743,29,1,23,9,31,17,39]'), (744, '[744,32,8,34,22,48,36,12]'), (745, '[745,35,15,45,35,15,5,35]'), (746, '[746,38,22,6,48,32,24,8]'), (747, '[747,41,29,17,11,49,43,31]'), (748, '[748,44,36,28,24,16,12,4]'), (749, '[749,47,43,39,37,33,31,27]'), (750, '[750,0,0,0,0,0,0,0]'), (751, '[751,3,7,11,13,17,19,23]'), (752, '[752,6,14,22,26,34,38,46]'), (753, '[753,9,21,33,39,1,7,19]'), (754, '[754,12,28,44,2,18,26,42]'), (755, '[755,15,35,5,15,35,45,15]'), (756, '[756,18,42,16,28,2,14,38]'), (757, '[757,21,49,27,41,19,33,11]'), (758, '[758,24,6,38,4,36,2,34]'), (759, '[759,27,13,49,17,3,21,7]'), (760, '[760,30,20,10,30,20,40,30]'), (761, '[761,33,27,21,43,37,9,3]'), (762, '[762,36,34,32,6,4,28,26]'), (763, '[763,39,41,43,19,21,47,49]'), (764, '[764,42,48,4,32,38,16,22]'), (765, '[765,45,5,15,45,5,35,45]'), (766, '[766,48,12,26,8,22,4,18]'), (767, '[767,1,19,37,21,39,23,41]'), (768, '[768,4,26,48,34,6,42,14]'), (769, '[769,7,33,9,47,23,11,37]'), (770, '[770,10,40,20,10,40,30,10]'), (771, '[771,13,47,31,23,7,49,33]'), (772, '[772,16,4,42,36,24,18,6]'), (773, '[773,19,11,3,49,41,37,29]'), (774, '[774,22,18,14,12,8,6,2]'), (775, '[775,25,25,25,25,25,25,25]'), (776, '[776,28,32,36,38,42,44,48]'), (777, '[777,31,39,47,1,9,13,21]'), (778, '[778,34,46,8,14,26,32,44]'), (779, '[779,37,3,19,27,43,1,17]'), (780, '[780,40,10,30,40,10,20,40]'), (781, '[781,43,17,41,3,27,39,13]'), (782, '[782,46,24,2,16,44,8,36]'), (783, '[783,49,31,13,29,11,27,9]'), (784, '[784,2,38,24,42,28,46,32]'), (785, '[785,5,45,35,5,45,15,5]'), (786, '[786,8,2,46,18,12,34,28]'), (787, '[787,11,9,7,31,29,3,1]'), (788, '[788,14,16,18,44,46,22,24]'), (789, '[789,17,23,29,7,13,41,47]'), (790, '[790,20,30,40,20,30,10,20]'), (791, '[791,23,37,1,33,47,29,43]'), (792, '[792,26,44,12,46,14,48,16]'), (793, '[793,29,1,23,9,31,17,39]'), (794, '[794,32,8,34,22,48,36,12]'), (795, '[795,35,15,45,35,15,5,35]'), (796, '[796,38,22,6,48,32,24,8]'), (797, '[797,41,29,17,11,49,43,31]'), (798, '[798,44,36,28,24,16,12,4]'), (799, '[799,47,43,39,37,33,31,27]'), (800, '[800,0,0,0,0,0,0,0]'), (801, '[801,3,7,11,13,17,19,23]'), (802, '[802,6,14,22,26,34,38,46]'), (803, '[803,9,21,33,39,1,7,19]'), (804, '[804,12,28,44,2,18,26,42]'), (805, '[805,15,35,5,15,35,45,15]'), (806, '[806,18,42,16,28,2,14,38]'), (807, '[807,21,49,27,41,19,33,11]'), (808, '[808,24,6,38,4,36,2,34]'), (809, '[809,27,13,49,17,3,21,7]'), (810, '[810,30,20,10,30,20,40,30]'), (811, '[811,33,27,21,43,37,9,3]'), (812, '[812,36,34,32,6,4,28,26]'), (813, '[813,39,41,43,19,21,47,49]'), (814, '[814,42,48,4,32,38,16,22]'), (815, '[815,45,5,15,45,5,35,45]'), (816, '[816,48,12,26,8,22,4,18]'), (817, '[817,1,19,37,21,39,23,41]'), (818, '[818,4,26,48,34,6,42,14]'), (819, '[819,7,33,9,47,23,11,37]'), (820, '[820,10,40,20,10,40,30,10]'), (821, '[821,13,47,31,23,7,49,33]'), (822, '[822,16,4,42,36,24,18,6]'), (823, '[823,19,11,3,49,41,37,29]'), (824, '[824,22,18,14,12,8,6,2]'), (825, '[825,25,25,25,25,25,25,25]'), (826, '[826,28,32,36,38,42,44,48]'), (827, '[827,31,39,47,1,9,13,21]'), (828, '[828,34,46,8,14,26,32,44]'), (829, '[829,37,3,19,27,43,1,17]'), (830, '[830,40,10,30,40,10,20,40]'), (831, '[831,43,17,41,3,27,39,13]'), (832, '[832,46,24,2,16,44,8,36]'), (833, '[833,49,31,13,29,11,27,9]'), (834, '[834,2,38,24,42,28,46,32]'), (835, '[835,5,45,35,5,45,15,5]'), (836, '[836,8,2,46,18,12,34,28]'), (837, '[837,11,9,7,31,29,3,1]'), (838, '[838,14,16,18,44,46,22,24]'), (839, '[839,17,23,29,7,13,41,47]'), (840, '[840,20,30,40,20,30,10,20]'), (841, '[841,23,37,1,33,47,29,43]'), (842, '[842,26,44,12,46,14,48,16]'), (843, '[843,29,1,23,9,31,17,39]'), (844, '[844,32,8,34,22,48,36,12]'), (845, '[845,35,15,45,35,15,5,35]'), (846, '[846,38,22,6,48,32,24,8]'), (847, '[847,41,29,17,11,49,43,31]'), (848, '[848,44,36,28,24,16,12,4]'), (849, '[849,47,43,39,37,33,31,27]'), (850, '[850,0,0,0,0,0,0,0]'), (851, '[851,3,7,11,13,17,19,23]'), (852, '[852,6,14,22,26,34,38,46]'), (853, '[853,9,21,33,39,1,7,19]'), (854, '[854,12,28,44,2,18,26,42]'), (855, '[855,15,35,5,15,35,45,15]'), (856, '[856,18,42,16,28,2,14,38]'), (857, '[857,21,49,27,41,19,33,11]'), (858, '[858,24,6,38,4,36,2,34]'), (859, '[859,27,13,49,17,3,21,7]'), (860, '[860,30,20,10,30,20,40,30]'), (861, '[861,33,27,21,43,37,9,3]'), (862, '[862,36,34,32,6,4,28,26]'), (863, '[863,39,41,43,19,21,47,49]'), (864, '[864,42,48,4,32,38,16,22]'), (865, '[865,45,5,15,45,5,35,45]'), (866, '[866,48,12,26,8,22,4,18]'), (867, '[867,1,19,37,21,39,23,41]'), (868, '[868,4,26,48,34,6,42,14]'), (869, '[869,7,33,9,47,23,11,37]'), (870, '[870,10,40,20,10,40,30,10]'), (871, '[871,13,47,31,23,7,49,33]'), (872, '[872,16,4,42,36,24,18,6]'), (873, '[873,19,11,3,49,41,37,29]'), (874, '[874,22,18,14,12,8,6,2]'), (875, '[875,25,25,25,25,25,25,25]'), (876, '[876,28,32,36,38,42,44,48]'), (877, '[877,31,39,47,1,9,13,21]'), (878, '[878,34,46,8,14,26,32,44]'), (879, '[879,37,3,19,27,43,1,17]'), (880, '[880,40,10,30,40,10,20,40]'), (881, '[881,43,17,41,3,27,39,13]'), (882, '[882,46,24,2,16,44,8,36]'), (883, '[883,49,31,13,29,11,27,9]'), (884, '[884,2,38,24,42,28,46,32]'), (885, '[885,5,45,35,5,45,15,5]'), (886, '[886,8,2,46,18,12,34,28]'), (887, '[887,11,9,7,31,29,3,1]'), (888, '[888,14,16,18,44,46,22,24]'), (889, '[889,17,23,29,7,13,41,47]'), (890, '[890,20,30,40,20,30,10,20]'), (891, '[891,23,37,1,33,47,29,43]'), (892, '[892,26,44,12,46,14,48,16]'), (893, '[893,29,1,23,9,31,17,39]'), (894, '[894,32,8,34,22,48,36,12]'), (895, '[895,35,15,45,35,15,5,35]'), (896, '[896,38,22,6,48,32,24,8]'), (897, '[897,41,29,17,11,49,43,31]'), (898, '[898,44,36,28,24,16,12,4]'), (899, '[899,47,43,39,37,33,31,27]'), (900, '[900,0,0,0,0,0,0,0]'), (901, '[901,3,7,11,13,17,19,23]'), (902, '[902,6,14,22,26,34,38,46]'), (903, '[903,9,21,33,39,1,7,19]'), (904, '[904,12,28,44,2,18,26,42]'), (905, '[905,15,35,5,15,35,45,15]'), (906, '[906,18,42,16,28,2,14,38]'), (907, '[907,21,49,27,41,19,33,11]'), (908, '[908,24,6,38,4,36,2,34]'), (909, '[909,27,13,49,17,3,21,7]'), (910, '[910,30,20,10,30,20,40,30]'), (911, '[911,33,27,21,43,37,9,3]'), (912, '[912,36,34,32,6,4,28,26]'), (913, '[913,39,41,43,19,21,47,49]'), (914, '[914,42,48,4,32,38,16,22]'), (915, '[915,45,5,15,45,5,35,45]'), (916, '[916,48,12,26,8,22,4,18]'), (917, '[917,1,19,37,21,39,23,41]'), (918, '[918,4,26,48,34,6,42,14]'), (919, '[919,7,33,9,47,23,11,37]'), (920, '[920,10,40,20,10,40,30,10]'), (921, '[921,13,47,31,23,7,49,33]'), (922, '[922,16,4,42,36,24,18,6]'), (923, '[923,19,11,3,49,41,37,29]'), (924, '[924,22,18,14,12,8,6,2]'), (925, '[925,25,25,25,25,25,25,25]'), (926, '[926,28,32,36,38,42,44,48]'), (927, '[927,31,39,47,1,9,13,21]'), (928, '[928,34,46,8,14,26,32,44]'), (929, '[929,37,3,19,27,43,1,17]'), (930, '[930,40,10,30,40,10,20,40]'), (931, '[931,43,17,41,3,27,39,13]'), (932, '[932,46,24,2,16,44,8,36]'), (933, '[933,49,31,13,29,11,27,9]'), (934, '[934,2,38,24,42,28,46,32]'), (935, '[935,5,45,35,5,45,15,5]'), (936, '[936,8,2,46,18,12,34,28]'), (937, '[937,11,9,7,31,29,3,1]'), (938, '[938,14,16,18,44,46,22,24]'), (939, '[939,17,23,29,7,13,41,47]'), (940, '[940,20,30,40,20,30,10,20]'), (941, '[941,23,37,1,33,47,29,43]'), (942, '[942,26,44,12,46,14,48,16]'), (943, '[943,29,1,23,9,31,17,39]'), (944, '[944,32,8,34,22,48,36,12]'), (945, '[945,35,15,45,35,15,5,35]'), (946, '[946,38,22,6,48,32,24,8]'), (947, '[947,41,29,17,11,49,43,31]'), (948, '[948,44,36,28,24,16,12,4]'), (949, '[949,47,43,39,37,33,31,27]'), (950, '[950,0,0,0,0,0,0,0]'), (951, '[951,3,7,11,13,17,19,23]'), (952, '[952,6,14,22,26,34,38,46]'), (953, '[953,9,21,33,39,1,7,19]'), (954, '[954,12,28,44,2,18,26,42]'), (955, '[955,15,35,5,15,35,45,15]'), (956, '[956,18,42,16,28,2,14,38]'), (957, '[957,21,49,27,41,19,33,11]'), (958, '[958,24,6,38,4,36,2,34]'), (959, '[959,27,13,49,17,3,21,7]'), (960, '[960,30,20,10,30,20,40,30]'), (961, '[961,33,27,21,43,37,9,3]'), (962, '[962,36,34,32,6,4,28,26]'), (963, '[963,39,41,43,19,21,47,49]'), (964, '[964,42,48,4,32,38,16,22]'), (965, '[965,45,5,15,45,5,35,45]'), (966, '[966,48,12,26,8,22,4,18]'), (967, '[967,1,19,37,21,39,23,41]'), (968, '[968,4,26,48,34,6,42,14]'), (969, '[969,7,33,9,47,23,11,37]'), (970, '[970,10,40,20,10,40,30,10]'), (971, '[971,13,47,31,23,7,49,33]'), (972, '[972,16,4,42,36,24,18,6]'), (973, '[973,19,11,3,49,41,37,29]'), (974, '[974,22,18,14,12,8,6,2]'), (975, '[975,25,25,25,25,25,25,25]'), (976, '[976,28,32,36,38,42,44,48]'), (977, '[977,31,39,47,1,9,13,21]'), (978, '[978,34,46,8,14,26,32,44]'), (979, '[979,37,3,19,27,43,1,17]'), (980, '[980,40,10,30,40,10,20,40]'), (981, '[981,43,17,41,3,27,39,13]'), (982, '[982,46,24,2,16,44,8,36]'), (983, '[983,49,31,13,29,11,27,9]'), (984, '[984,2,38,24,42,28,46,32]'), (985, '[985,5,45,35,5,45,15,5]'), (986, '[986,8,2,46,18,12,34,28]'), (987, '[987,11,9,7,31,29,3,1]'), (988, '[988,14,16,18,44,46,22,24]'), (989, '[989,17,23,29,7,13,41,47]'), (990, '[990,20,30,40,20,30,10,20]'), (991, '[991,23,37,1,33,47,29,43]'), (992, '[992,26,44,12,46,14,48,16]'), (993, '[993,29,1,23,9,31,17,39]'), (994, '[994,32,8,34,22,48,36,12]'), (995, '[995,35,15,45,35,15,5,35]'), (996, '[996,38,22,6,48,32,24,8]'), (997, '[997,41,29,17,11,49,43,31]'), (998, '[998,44,36,28,24,16,12,4]'), (999, '[999,47,43,39,37,33,31,27]'), (1000, '[1000,0,0,0,0,0,0,0]'), (1001, '[1001,3,7,11,13,17,19,23]'), (1002, '[1002,6,14,22,26,34,38,46]'), (1003, '[1003,9,21,33,39,1,7,19]'), (1004, '[1004,12,28,44,2,18,26,42]'), (1005, '[1005,15,35,5,15,35,45,15]'), (1006, '[1006,18,42,16,28,2,14,38]'), (1007, '[1007,21,49,27,41,19,33,11]'), (1008, '[1008,24,6,38,4,36,2,34]'), (1009, '[1009,27,13,49,17,3,21,7]'), (1010, '[1010,30,20,10,30,20,40,30]'), (1011, '[1011,33,27,21,43,37,9,3]'), (1012, '[1012,36,34,32,6,4,28,26]'), (1013, '[1013,39,41,43,19,21,47,49]'), (1014, '[1014,42,48,4,32,38,16,22]'), (1015, '[1015,45,5,15,45,5,35,45]'), (1016, '[1016,48,12,26,8,22,4,18]'), (1017, '[1017,1,19,37,21,39,23,41]'), (1018, '[1018,4,26,48,34,6,42,14]'), (1019, '[1019,7,33,9,47,23,11,37]'), (1020, '[1020,10,40,20,10,40,30,10]'), (1021, '[1021,13,47,31,23,7,49,33]'), (1022, '[1022,16,4,42,36,24,18,6]'), (1023, '[1023,19,11,3,49,41,37,29]'), (1024, '[1024,22,18,14,12,8,6,2]'), (1025, '[1025,25,25,25,25,25,25,25]'), (1026, '[1026,28,32,36,38,42,44,48]'), (1027, '[1027,31,39,47,1,9,13,21]'), (1028, '[1028,34,46,8,14,26,32,44]'), (1029, '[1029,37,3,19,27,43,1,17]'), (1030, '[1030,40,10,30,40,10,20,40]'), (1031, '[1031,43,17,41,3,27,39,13]'), (1032, '[1032,46,24,2,16,44,8,36]'), (1033, '[1033,49,31,13,29,11,27,9]'), (1034, '[1034,2,38,24,42,28,46,32]'), (1035, '[1035,5,45,35,5,45,15,5]'), (1036, '[1036,8,2,46,18,12,34,28]'), (1037, '[1037,11,9,7,31,29,3,1]'), (1038, '[1038,14,16,18,44,46,22,24]'), (1039, '[1039,17,23,29,7,13,41,47]'), (1040, '[1040,20,30,40,20,30,10,20]'), (1041, '[1041,23,37,1,33,47,29,43]'), (1042, '[1042,26,44,12,46,14,48,16]'), (1043, '[1043,29,1,23,9,31,17,39]'), (1044, '[1044,32,8,34,22,48,36,12]'), (1045, '[1045,35,15,45,35,15,5,35]'), (1046, '[1046,38,22,6,48,32,24,8]'), (1047, '[1047,41,29,17,11,49,43,31]'), (1048, '[1048,44,36,28,24,16,12,4]'), (1049, '[1049,47,43,39,37,33,31,27]'), (1050, '[1050,0,0,0,0,0,0,0]'), (1051, '[1051,3,7,11,13,17,19,23]'), (1052, '[1052,6,14,22,26,34,38,46]'), (1053, '[1053,9,21,33,39,1,7,19]'), (1054, '[1054,12,28,44,2,18,26,42]'), (1055, '[1055,15,35,5,15,35,45,15]'), (1056, '[1056,18,42,16,28,2,14,38]'), (1057, '[1057,21,49,27,41,19,33,11]'), (1058, '[1058,24,6,38,4,36,2,34]'), (1059, '[1059,27,13,49,17,3,21,7]'), (1060, '[1060,30,20,10,30,20,40,30]'), (1061, '[1061,33,27,21,43,37,9,3]'), (1062, '[1062,36,34,32,6,4,28,26]'), (1063, '[1063,39,41,43,19,21,47,49]'), (1064, '[1064,42,48,4,32,38,16,22]'), (1065, '[1065,45,5,15,45,5,35,45]'), (1066, '[1066,48,12,26,8,22,4,18]'), (1067, '[1067,1,19,37,21,39,23,41]'), (1068, '[1068,4,26,48,34,6,42,14]'), (1069, '[1069,7,33,9,47,23,11,37]'), (1070, '[1070,10,40,20,10,40,30,10]'), (1071, '[1071,13,47,31,23,7,49,33]'), (1072, '[1072,16,4,42,36,24,18,6]'), (1073, '[1073,19,11,3,49,41,37,29]'), (1074, '[1074,22,18,14,12,8,6,2]'), (1075, '[1075,25,25,25,25,25,25,25]'), (1076, '[1076,28,32,36,38,42,44,48]'), (1077, '[1077,31,39,47,1,9,13,21]'), (1078, '[1078,34,46,8,14,26,32,44]'), (1079, '[1079,37,3,19,27,43,1,17]'), (1080, '[1080,40,10,30,40,10,20,40]'), (1081, '[1081,43,17,41,3,27,39,13]'), (1082, '[1082,46,24,2,16,44,8,36]'), (1083, '[1083,49,31,13,29,11,27,9]'), (1084, '[1084,2,38,24,42,28,46,32]'), (1085, '[1085,5,45,35,5,45,15,5]'), (1086, '[1086,8,2,46,18,12,34,28]'), (1087, '[1087,11,9,7,31,29,3,1]'), (1088, '[1088,14,16,18,44,46,22,24]'), (1089, '[1089,17,23,29,7,13,41,47]'), (1090, '[1090,20,30,40,20,30,10,20]'), (1091, '[1091,23,37,1,33,47,29,43]'), (1092, '[1092,26,44,12,46,14,48,16]'), (1093, '[1093,29,1,23,9,31,17,39]'), (1094, '[1094,32,8,34,22,48,36,12]'), (1095, '[1095,35,15,45,35,15,5,35]'), (1096, '[1096,38,22,6,48,32,24,8]'), (1097, '[1097,41,29,17,11,49,43,31]'), (1098, '[1098,44,36,28,24,16,12,4]'), (1099, '[1099,47,43,39,37,33,31,27]'), (1100, '[1100,0,0,0,0,0,0,0]'), (1101, '[1101,3,7,11,13,17,19,23]'), (1102, '[1102,6,14,22,26,34,38,46]'), (1103, '[1103,9,21,33,39,1,7,19]'), (1104, '[1104,12,28,44,2,18,26,42]'), (1105, '[1105,15,35,5,15,35,45,15]'), (1106, '[1106,18,42,16,28,2,14,38]'), (1107, '[1107,21,49,27,41,19,33,11]'), (1108, '[1108,24,6,38,4,36,2,34]'), (1109, '[1109,27,13,49,17,3,21,7]'), (1110, '[1110,30,20,10,30,20,40,30]'), (1111, '[1111,33,27,21,43,37,9,3]'), (1112, '[1112,36,34,32,6,4,28,26]'), (1113, '[1113,39,41,43,19,21,47,49]'), (1114, '[1114,42,48,4,32,38,16,22]'), (1115, '[1115,45,5,15,45,5,35,45]'), (1116, '[1116,48,12,26,8,22,4,18]'), (1117, '[1117,1,19,37,21,39,23,41]'), (1118, '[1118,4,26,48,34,6,42,14]'), (1119, '[1119,7,33,9,47,23,11,37]'), (1120, '[1120,10,40,20,10,40,30,10]'), (1121, '[1121,13,47,31,23,7,49,33]'), (1122, '[1122,16,4,42,36,24,18,6]'), (1123, '[1123,19,11,3,49,41,37,29]'), (1124, '[1124,22,18,14,12,8,6,2]'), (1125, '[1125,25,25,25,25,25,25,25]'), (1126, '[1126,28,32,36,38,42,44,48]'), (1127, '[1127,31,39,47,1,9,13,21]'), (1128, '[1128,34,46,8,14,26,32,44]'), (1129, '[1129,37,3,19,27,43,1,17]'), (1130, '[1130,40,10,30,40,10,20,40]'), (1131, '[1131,43,17,41,3,27,39,13]'), (1132, '[1132,46,24,2,16,44,8,36]'), (1133, '[1133,49,31,13,29,11,27,9]'), (1134, '[1134,2,38,24,42,28,46,32]'), (1135, '[1135,5,45,35,5,45,15,5]'), (1136, '[1136,8,2,46,18,12,34,28]'), (1137, '[1137,11,9,7,31,29,3,1]'), (1138, '[1138,14,16,18,44,46,22,24]'), (1139, '[1139,17,23,29,7,13,41,47]'), (1140, '[1140,20,30,40,20,30,10,20]'), (1141, '[1141,23,37,1,33,47,29,43]'), (1142, '[1142,26,44,12,46,14,48,16]'), (1143, '[1143,29,1,23,9,31,17,39]'), (1144, '[1144,32,8,34,22,48,36,12]'), (1145, '[1145,35,15,45,35,15,5,35]'), (1146, '[1146,38,22,6,48,32,24,8]'), (1147, '[1147,41,29,17,11,49,43,31]'), (1148, '[1148,44,36,28,24,16,12,4]'), (1149, '[1149,47,43,39,37,33,31,27]'), (1150, '[1150,0,0,0,0,0,0,0]'), (1151, '[1151,3,7,11,13,17,19,23]'), (1152, '[1152,6,14,22,26,34,38,46]'), (1153, '[1153,9,21,33,39,1,7,19]'), (1154, '[1154,12,28,44,2,18,26,42]'), (1155, '[1155,15,35,5,15,35,45,15]'), (1156, '[1156,18,42,16,28,2,14,38]'), (1157, '[1157,21,49,27,41,19,33,11]'), (1158, '[1158,24,6,38,4,36,2,34]'), (1159, '[1159,27,13,49,17,3,21,7]'), (1160, '[1160,30,20,10,30,20,40,30]'), (1161, '[1161,33,27,21,43,37,9,3]'), (1162, '[1162,36,34,32,6,4,28,26]'), (1163, '[1163,39,41,43,19,21,47,49]'), (1164, '[1164,42,48,4,32,38,16,22]'), (1165, '[1165,45,5,15,45,5,35,45]'), (1166, '[1166,48,12,26,8,22,4,18]'), (1167, '[1167,1,19,37,21,39,23,41]'), (1168, '[1168,4,26,48,34,6,42,14]'), (1169, '[1169,7,33,9,47,23,11,37]'), (1170, '[1170,10,40,20,10,40,30,10]'), (1171, '[1171,13,47,31,23,7,49,33]'), (1172, '[1172,16,4,42,36,24,18,6]'), (1173, '[1173,19,11,3,49,41,37,29]'), (1174, '[1174,22,18,14,12,8,6,2]'), (1175, '[1175,25,25,25,25,25,25,25]'), (1176, '[1176,28,32,36,38,42,44,48]'), (1177, '[1177,31,39,47,1,9,13,21]'), (1178, '[1178,34,46,8,14,26,32,44]'), (1179, '[1179,37,3,19,27,43,1,17]'), (1180, '[1180,40,10,30,40,10,20,40]'), (1181, '[1181,43,17,41,3,27,39,13]'), (1182, '[1182,46,24,2,16,44,8,36]'), (1183, '[1183,49,31,13,29,11,27,9]'), (1184, '[1184,2,38,24,42,28,46,32]'), (1185, '[1185,5,45,35,5,45,15,5]'), (1186, '[1186,8,2,46,18,12,34,28]'), (1187, '[1187,11,9,7,31,29,3,1]'), (1188, '[1188,14,16,18,44,46,22,24]'), (1189, '[1189,17,23,29,7,13,41,47]'), (1190, '[1190,20,30,40,20,30,10,20]'), (1191, '[1191,23,37,1,33,47,29,43]'), (1192, '[1192,26,44,12,46,14,48,16]'), (1193, '[1193,29,1,23,9,31,17,39]'), (1194, '[1194,32,8,34,22,48,36,12]'), (1195, '[1195,35,15,45,35,15,5,35]'), (1196, '[1196,38,22,6,48,32,24,8]'), (1197, '[1197,41,29,17,11,49,43,31]'), (1198, '[1198,44,36,28,24,16,12,4]'), (1199, '[1199,47,43,39,37,33,31,27]'), (1200, '[1200,0,0,0,0,0,0,0]'), (1201, '[1201,3,7,11,13,17,19,23]'), (1202, '[1202,6,14,22,26,34,38,46]'), (1203, '[1203,9,21,33,39,1,7,19]'), (1204, '[1204,12,28,44,2,18,26,42]'), (1205, '[1205,15,35,5,15,35,45,15]'), (1206, '[1206,18,42,16,28,2,14,38]'), (1207, '[1207,21,49,27,41,19,33,11]'), (1208, '[1208,24,6,38,4,36,2,34]'), (1209, '[1209,27,13,49,17,3,21,7]'), (1210, '[1210,30,20,10,30,20,40,30]'), (1211, '[1211,33,27,21,43,37,9,3]'), (1212, '[1212,36,34,32,6,4,28,26]'), (1213, '[1213,39,41,43,19,21,47,49]'), (1214, '[1214,42,48,4,32,38,16,22]'), (1215, '[1215,45,5,15,45,5,35,45]'), (1216, '[1216,48,12,26,8,22,4,18]'), (1217, '[1217,1,19,37,21,39,23,41]'), (1218, '[1218,4,26,48,34,6,42,14]'), (1219, '[1219,7,33,9,47,23,11,37]'), (1220, '[1220,10,40,20,10,40,30,10]'), (1221, '[1221,13,47,31,23,7,49,33]'), (1222, '[1222,16,4,42,36,24,18,6]'), (1223, '[1223,19,11,3,49,41,37,29]'), (1224, '[1224,22,18,14,12,8,6,2]'), (1225, '[1225,25,25,25,25,25,25,25]'), (1226, '[1226,28,32,36,38,42,44,48]'), (1227, '[1227,31,39,47,1,9,13,21]'), (1228, '[1228,34,46,8,14,26,32,44]'), (1229, '[1229,37,3,19,27,43,1,17]'), (1230, '[1230,40,10,30,40,10,20,40]'), (1231, '[1231,43,17,41,3,27,39,13]'), (1232, '[1232,46,24,2,16,44,8,36]'), (1233, '[1233,49,31,13,29,11,27,9]'), (1234, '[1234,2,38,24,42,28,46,32]'), (1235, '[1235,5,45,35,5,45,15,5]'), (1236, '[1236,8,2,46,18,12,34,28]'), (1237, '[1237,11,9,7,31,29,3,1]'), (1238, '[1238,14,16,18,44,46,22,24]'), (1239, '[1239,17,23,29,7,13,41,47]'), (1240, '[1240,20,30,40,20,30,10,20]'), (1241, '[1241,23,37,1,33,47,29,43]'), (1242, '[1242,26,44,12,46,14,48,16]'), (1243, '[1243,29,1,23,9,31,17,39]'), (1244, '[1244,32,8,34,22,48,36,12]'), (1245, '[1245,35,15,45,35,15,5,35]'), (1246, '[1246,38,22,6,48,32,24,8]'), (1247, '[1247,41,29,17,11,49,43,31]'), (1248, '[1248,44,36,28,24,16,12,4]'), (1249, '[1249,47,43,39,37,33,31,27]'), (1250, '[1250,0,0,0,0,0,0,0]'), (1251, '[1251,3,7,11,13,17,19,23]'), (1252, '[1252,6,14,22,26,34,38,46]'), (1253, '[1253,9,21,33,39,1,7,19]'), (1254, '[1254,12,28,44,2,18,26,42]'), (1255, '[1255,15,35,5,15,35,45,15]'), (1256, '[1256,18,42,16,28,2,14,38]'), (1257, '[1257,21,49,27,41,19,33,11]'), (1258, '[1258,24,6,38,4,36,2,34]'), (1259, '[1259,27,13,49,17,3,21,7]'), (1260, '[1260,30,20,10,30,20,40,30]'), (1261, '[1261,33,27,21,43,37,9,3]'), (1262, '[1262,36,34,32,6,4,28,26]'), (1263, '[1263,39,41,43,19,21,47,49]'), (1264, '[1264,42,48,4,32,38,16,22]'), (1265, '[1265,45,5,15,45,5,35,45]'), (1266, '[1266,48,12,26,8,22,4,18]'), (1267, '[1267,1,19,37,21,39,23,41]'), (1268, '[1268,4,26,48,34,6,42,14]'), (1269, '[1269,7,33,9,47,23,11,37]'), (1270, '[1270,10,40,20,10,40,30,10]'), (1271, '[1271,13,47,31,23,7,49,33]'), (1272, '[1272,16,4,42,36,24,18,6]'), (1273, '[1273,19,11,3,49,41,37,29]'), (1274, '[1274,22,18,14,12,8,6,2]'), (1275, '[1275,25,25,25,25,25,25,25]'), (1276, '[1276,28,32,36,38,42,44,48]'), (1277, '[1277,31,39,47,1,9,13,21]'), (1278, '[1278,34,46,8,14,26,32,44]'), (1279, '[1279,37,3,19,27,43,1,17]'), (1280, '[1280,40,10,30,40,10,20,40]'), (1281, '[1281,43,17,41,3,27,39,13]'), (1282, '[1282,46,24,2,16,44,8,36]'), (1283, '[1283,49,31,13,29,11,27,9]'), (1284, '[1284,2,38,24,42,28,46,32]'), (1285, '[1285,5,45,35,5,45,15,5]'), (1286, '[1286,8,2,46,18,12,34,28]'), (1287, '[1287,11,9,7,31,29,3,1]'), (1288, '[1288,14,16,18,44,46,22,24]'), (1289, '[1289,17,23,29,7,13,41,47]'), (1290, '[1290,20,30,40,20,30,10,20]'), (1291, '[1291,23,37,1,33,47,29,43]'), (1292, '[1292,26,44,12,46,14,48,16]'), (1293, '[1293,29,1,23,9,31,17,39]'), (1294, '[1294,32,8,34,22,48,36,12]'), (1295, '[1295,35,15,45,35,15,5,35]'), (1296, '[1296,38,22,6,48,32,24,8]'), (1297, '[1297,41,29,17,11,49,43,31]'), (1298, '[1298,44,36,28,24,16,12,4]'), (1299, '[1299,47,43,39,37,33,31,27]'), (1300, '[1300,0,0,0,0,0,0,0]'), (1301, '[1301,3,7,11,13,17,19,23]'), (1302, '[1302,6,14,22,26,34,38,46]'), (1303, '[1303,9,21,33,39,1,7,19]'), (1304, '[1304,12,28,44,2,18,26,42]'), (1305, '[1305,15,35,5,15,35,45,15]'), (1306, '[1306,18,42,16,28,2,14,38]'), (1307, '[1307,21,49,27,41,19,33,11]'), (1308, '[1308,24,6,38,4,36,2,34]'), (1309, '[1309,27,13,49,17,3,21,7]'), (1310, '[1310,30,20,10,30,20,40,30]'), (1311, '[1311,33,27,21,43,37,9,3]'), (1312, '[1312,36,34,32,6,4,28,26]'), (1313, '[1313,39,41,43,19,21,47,49]'), (1314, '[1314,42,48,4,32,38,16,22]'), (1315, '[1315,45,5,15,45,5,35,45]'), (1316, '[1316,48,12,26,8,22,4,18]'), (1317, '[1317,1,19,37,21,39,23,41]'), (1318, '[1318,4,26,48,34,6,42,14]'), (1319, '[1319,7,33,9,47,23,11,37]'), (1320, '[1320,10,40,20,10,40,30,10]'), (1321, '[1321,13,47,31,23,7,49,33]'), (1322, '[1322,16,4,42,36,24,18,6]'), (1323, '[1323,19,11,3,49,41,37,29]'), (1324, '[1324,22,18,14,12,8,6,2]'), (1325, '[1325,25,25,25,25,25,25,25]'), (1326, '[1326,28,32,36,38,42,44,48]'), (1327, '[1327,31,39,47,1,9,13,21]'), (1328, '[1328,34,46,8,14,26,32,44]'), (1329, '[1329,37,3,19,27,43,1,17]'), (1330, '[1330,40,10,30,40,10,20,40]'), (1331, '[1331,43,17,41,3,27,39,13]'), (1332, '[1332,46,24,2,16,44,8,36]'), (1333, '[1333,49,31,13,29,11,27,9]'), (1334, '[1334,2,38,24,42,28,46,32]'), (1335, '[1335,5,45,35,5,45,15,5]'), (1336, '[1336,8,2,46,18,12,34,28]'), (1337, '[1337,11,9,7,31,29,3,1]'), (1338, '[1338,14,16,18,44,46,22,24]'), (1339, '[1339,17,23,29,7,13,41,47]'), (1340, '[1340,20,30,40,20,30,10,20]'), (1341, '[1341,23,37,1,33,47,29,43]'), (1342, '[1342,26,44,12,46,14,48,16]'), (1343, '[1343,29,1,23,9,31,17,39]'), (1344, '[1344,32,8,34,22,48,36,12]'), (1345, '[1345,35,15,45,35,15,5,35]'), (1346, '[1346,38,22,6,48,32,24,8]'), (1347, '[1347,41,29,17,11,49,43,31]'), (1348, '[1348,44,36,28,24,16,12,4]'), (1349, '[1349,47,43,39,37,33,31,27]'), (1350, '[1350,0,0,0,0,0,0,0]'), (1351, '[1351,3,7,11,13,17,19,23]'), (1352, '[1352,6,14,22,26,34,38,46]'), (1353, '[1353,9,21,33,39,1,7,19]'), (1354, '[1354,12,28,44,2,18,26,42]'), (1355, '[1355,15,35,5,15,35,45,15]'), (1356, '[1356,18,42,16,28,2,14,38]'), (1357, '[1357,21,49,27,41,19,33,11]'), (1358, '[1358,24,6,38,4,36,2,34]'), (1359, '[1359,27,13,49,17,3,21,7]'), (1360, '[1360,30,20,10,30,20,40,30]'), (1361, '[1361,33,27,21,43,37,9,3]'), (1362, '[1362,36,34,32,6,4,28,26]'), (1363, '[1363,39,41,43,19,21,47,49]'), (1364, '[1364,42,48,4,32,38,16,22]'), (1365, '[1365,45,5,15,45,5,35,45]'), (1366, '[1366,48,12,26,8,22,4,18]'), (1367, '[1367,1,19,37,21,39,23,41]'), (1368, '[1368,4,26,48,34,6,42,14]'), (1369, '[1369,7,33,9,47,23,11,37]'), (1370, '[1370,10,40,20,10,40,30,10]'), (1371, '[1371,13,47,31,23,7,49,33]'), (1372, '[1372,16,4,42,36,24,18,6]'), (1373, '[1373,19,11,3,49,41,37,29]'), (1374, '[1374,22,18,14,12,8,6,2]'), (1375, '[1375,25,25,25,25,25,25,25]'), (1376, '[1376,28,32,36,38,42,44,48]'), (1377, '[1377,31,39,47,1,9,13,21]'), (1378, '[1378,34,46,8,14,26,32,44]'), (1379, '[1379,37,3,19,27,43,1,17]'), (1380, '[1380,40,10,30,40,10,20,40]'), (1381, '[1381,43,17,41,3,27,39,13]'), (1382, '[1382,46,24,2,16,44,8,36]'), (1383, '[1383,49,31,13,29,11,27,9]'), (1384, '[1384,2,38,24,42,28,46,32]'), (1385, '[1385,5,45,35,5,45,15,5]'), (1386, '[1386,8,2,46,18,12,34,28]'), (1387, '[1387,11,9,7,31,29,3,1]'), (1388, '[1388,14,16,18,44,46,22,24]'), (1389, '[1389,17,23,29,7,13,41,47]'), (1390, '[1390,20,30,40,20,30,10,20]'), (1391, '[1391,23,37,1,33,47,29,43]'), (1392, '[1392,26,44,12,46,14,48,16]'), (1393, '[1393,29,1,23,9,31,17,39]'), (1394, '[1394,32,8,34,22,48,36,12]'), (1395, '[1395,35,15,45,35,15,5,35]'), (1396, '[1396,38,22,6,48,32,24,8]'), (1397, '[1397,41,29,17,11,49,43,31]'), (1398, '[1398,44,36,28,24,16,12,4]'), (1399, '[1399,47,43,39,37,33,31,27]'), (1400, '[1400,0,0,0,0,0,0,0]'), (1401, '[1401,3,7,11,13,17,19,23]'), (1402, '[1402,6,14,22,26,34,38,46]'), (1403, '[1403,9,21,33,39,1,7,19]'), (1404, '[1404,12,28,44,2,18,26,42]'), (1405, '[1405,15,35,5,15,35,45,15]'), (1406, '[1406,18,42,16,28,2,14,38]'), (1407, '[1407,21,49,27,41,19,33,11]'), (1408, '[1408,24,6,38,4,36,2,34]'), (1409, '[1409,27,13,49,17,3,21,7]'), (1410, '[1410,30,20,10,30,20,40,30]'), (1411, '[1411,33,27,21,43,37,9,3]'), (1412, '[1412,36,34,32,6,4,28,26]'), (1413, '[1413,39,41,43,19,21,47,49]'), (1414, '[1414,42,48,4,32,38,16,22]'), (1415, '[1415,45,5,15,45,5,35,45]'), (1416, '[1416,48,12,26,8,22,4,18]'), (1417, '[1417,1,19,37,21,39,23,41]'), (1418, '[1418,4,26,48,34,6,42,14]'), (1419, '[1419,7,33,9,47,23,11,37]'), (1420, '[1420,10,40,20,10,40,30,10]'), (1421, '[1421,13,47,31,23,7,49,33]'), (1422, '[1422,16,4,42,36,24,18,6]'), (1423, '[1423,19,11,3,49,41,37,29]'), (1424, '[1424,22,18,14,12,8,6,2]'), (1425, '[1425,25,25,25,25,25,25,25]'), (1426, '[1426,28,32,36,38,42,44,48]'), (1427, '[1427,31,39,47,1,9,13,21]'), (1428, '[1428,34,46,8,14,26,32,44]'), (1429, '[1429,37,3,19,27,43,1,17]'), (1430, '[1430,40,10,30,40,10,20,40]'), (1431, '[1431,43,17,41,3,27,39,13]'), (1432, '[1432,46,24,2,16,44,8,36]'), (1433, '[1433,49,31,13,29,11,27,9]'), (1434, '[1434,2,38,24,42,28,46,32]'), (1435, '[1435,5,45,35,5,45,15,5]'), (1436, '[1436,8,2,46,18,12,34,28]'), (1437, '[1437,11,9,7,31,29,3,1]'), (1438, '[1438,14,16,18,44,46,22,24]'), (1439, '[1439,17,23,29,7,13,41,47]'), (1440, '[1440,20,30,40,20,30,10,20]'), (1441, '[1441,23,37,1,33,47,29,43]'), (1442, '[1442,26,44,12,46,14,48,16]'), (1443, '[1443,29,1,23,9,31,17,39]'), (1444, '[1444,32,8,34,22,48,36,12]'), (1445, '[1445,35,15,45,35,15,5,35]'), (1446, '[1446,38,22,6,48,32,24,8]'), (1447, '[1447,41,29,17,11,49,43,31]'), (1448, '[1448,44,36,28,24,16,12,4]'), (1449, '[1449,47,43,39,37,33,31,27]'), (1450, '[1450,0,0,0,0,0,0,0]'), (1451, '[1451,3,7,11,13,17,19,23]'), (1452, '[1452,6,14,22,26,34,38,46]'), (1453, '[1453,9,21,33,39,1,7,19]'), (1454, '[1454,12,28,44,2,18,26,42]'), (1455, '[1455,15,35,5,15,35,45,15]'), (1456, '[1456,18,42,16,28,2,14,38]'), (1457, '[1457,21,49,27,41,19,33,11]'), (1458, '[1458,24,6,38,4,36,2,34]'), (1459, '[1459,27,13,49,17,3,21,7]'), (1460, '[1460,30,20,10,30,20,40,30]'), (1461, '[1461,33,27,21,43,37,9,3]'), (1462, '[1462,36,34,32,6,4,28,26]'), (1463, '[1463,39,41,43,19,21,47,49]'), (1464, '[1464,42,48,4,32,38,16,22]'), (1465, '[1465,45,5,15,45,5,35,45]'), (1466, '[1466,48,12,26,8,22,4,18]'), (1467, '[1467,1,19,37,21,39,23,41]'), (1468, '[1468,4,26,48,34,6,42,14]'), (1469, '[1469,7,33,9,47,23,11,37]'), (1470, '[1470,10,40,20,10,40,30,10]'), (1471, '[1471,13,47,31,23,7,49,33]'), (1472, '[1472,16,4,42,36,24,18,6]'), (1473, '[1473,19,11,3,49,41,37,29]'), (1474, '[1474,22,18,14,12,8,6,2]'), (1475, '[1475,25,25,25,25,25,25,25]'), (1476, '[1476,28,32,36,38,42,44,48]'), (1477, '[1477,31,39,47,1,9,13,21]'), (1478, '[1478,34,46,8,14,26,32,44]'), (1479, '[1479,37,3,19,27,43,1,17]'), (1480, '[1480,40,10,30,40,10,20,40]'), (1481, '[1481,43,17,41,3,27,39,13]'), (1482, '[1482,46,24,2,16,44,8,36]'), (1483, '[1483,49,31,13,29,11,27,9]'), (1484, '[1484,2,38,24,42,28,46,32]'), (1485, '[1485,5,45,35,5,45,15,5]'), (1486, '[1486,8,2,46,18,12,34,28]'), (1487, '[1487,11,9,7,31,29,3,1]'), (1488, '[1488,14,16,18,44,46,22,24]'), (1489, '[1489,17,23,29,7,13,41,47]'), (1490, '[1490,20,30,40,20,30,10,20]'), (1491, '[1491,23,37,1,33,47,29,43]'), (1492, '[1492,26,44,12,46,14,48,16]'), (1493, '[1493,29,1,23,9,31,17,39]'), (1494, '[1494,32,8,34,22,48,36,12]'), (1495, '[1495,35,15,45,35,15,5,35]'), (1496, '[1496,38,22,6,48,32,24,8]'), (1497, '[1497,41,29,17,11,49,43,31]'), (1498, '[1498,44,36,28,24,16,12,4]'), (1499, '[1499,47,43,39,37,33,31,27]'), (1500, '[1500,0,0,0,0,0,0,0]'), (1501, '[1501,3,7,11,13,17,19,23]'), (1502, '[1502,6,14,22,26,34,38,46]'), (1503, '[1503,9,21,33,39,1,7,19]'), (1504, '[1504,12,28,44,2,18,26,42]'), (1505, '[1505,15,35,5,15,35,45,15]'), (1506, '[1506,18,42,16,28,2,14,38]'), (1507, '[1507,21,49,27,41,19,33,11]'), (1508, '[1508,24,6,38,4,36,2,34]'), (1509, '[1509,27,13,49,17,3,21,7]'), (1510, '[1510,30,20,10,30,20,40,30]'), (1511, '[1511,33,27,21,43,37,9,3]'), (1512, '[1512,36,34,32,6,4,28,26]'), (1513, '[1513,39,41,43,19,21,47,49]'), (1514, '[1514,42,48,4,32,38,16,22]'), (1515, '[1515,45,5,15,45,5,35,45]'), (1516, '[1516,48,12,26,8,22,4,18]'), (1517, '[1517,1,19,37,21,39,23,41]'), (1518, '[1518,4,26,48,34,6,42,14]'), (1519, '[1519,7,33,9,47,23,11,37]'), (1520, '[1520,10,40,20,10,40,30,10]'), (1521, '[1521,13,47,31,23,7,49,33]'), (1522, '[1522,16,4,42,36,24,18,6]'), (1523, '[1523,19,11,3,49,41,37,29]'), (1524, '[1524,22,18,14,12,8,6,2]'), (1525, '[1525,25,25,25,25,25,25,25]'), (1526, '[1526,28,32,36,38,42,44,48]'), (1527, '[1527,31,39,47,1,9,13,21]'), (1528, '[1528,34,46,8,14,26,32,44]'), (1529, '[1529,37,3,19,27,43,1,17]'), (1530, '[1530,40,10,30,40,10,20,40]'), (1531, '[1531,43,17,41,3,27,39,13]'), (1532, '[1532,46,24,2,16,44,8,36]'), (1533, '[1533,49,31,13,29,11,27,9]'), (1534, '[1534,2,38,24,42,28,46,32]'), (1535, '[1535,5,45,35,5,45,15,5]'), (1536, '[1536,8,2,46,18,12,34,28]'), (1537, '[1537,11,9,7,31,29,3,1]'), (1538, '[1538,14,16,18,44,46,22,24]'), (1539, '[1539,17,23,29,7,13,41,47]'), (1540, '[1540,20,30,40,20,30,10,20]'), (1541, '[1541,23,37,1,33,47,29,43]'), (1542, '[1542,26,44,12,46,14,48,16]'), (1543, '[1543,29,1,23,9,31,17,39]'), (1544, '[1544,32,8,34,22,48,36,12]'), (1545, '[1545,35,15,45,35,15,5,35]'), (1546, '[1546,38,22,6,48,32,24,8]'), (1547, '[1547,41,29,17,11,49,43,31]'), (1548, '[1548,44,36,28,24,16,12,4]'), (1549, '[1549,47,43,39,37,33,31,27]'), (1550, '[1550,0,0,0,0,0,0,0]'), (1551, '[1551,3,7,11,13,17,19,23]'), (1552, '[1552,6,14,22,26,34,38,46]'), (1553, '[1553,9,21,33,39,1,7,19]'), (1554, '[1554,12,28,44,2,18,26,42]'), (1555, '[1555,15,35,5,15,35,45,15]'), (1556, '[1556,18,42,16,28,2,14,38]'), (1557, '[1557,21,49,27,41,19,33,11]'), (1558, '[1558,24,6,38,4,36,2,34]'), (1559, '[1559,27,13,49,17,3,21,7]'), (1560, '[1560,30,20,10,30,20,40,30]'), (1561, '[1561,33,27,21,43,37,9,3]'), (1562, '[1562,36,34,32,6,4,28,26]'), (1563, '[1563,39,41,43,19,21,47,49]'), (1564, '[1564,42,48,4,32,38,16,22]'), (1565, '[1565,45,5,15,45,5,35,45]'), (1566, '[1566,48,12,26,8,22,4,18]'), (1567, '[1567,1,19,37,21,39,23,41]'), (1568, '[1568,4,26,48,34,6,42,14]'), (1569, '[1569,7,33,9,47,23,11,37]'), (1570, '[1570,10,40,20,10,40,30,10]'), (1571, '[1571,13,47,31,23,7,49,33]'), (1572, '[1572,16,4,42,36,24,18,6]'), (1573, '[1573,19,11,3,49,41,37,29]'), (1574, '[1574,22,18,14,12,8,6,2]'), (1575, '[1575,25,25,25,25,25,25,25]'), (1576, '[1576,28,32,36,38,42,44,48]'), (1577, '[1577,31,39,47,1,9,13,21]'), (1578, '[1578,34,46,8,14,26,32,44]'), (1579, '[1579,37,3,19,27,43,1,17]'), (1580, '[1580,40,10,30,40,10,20,40]'), (1581, '[1581,43,17,41,3,27,39,13]'), (1582, '[1582,46,24,2,16,44,8,36]'), (1583, '[1583,49,31,13,29,11,27,9]'), (1584, '[1584,2,38,24,42,28,46,32]'), (1585, '[1585,5,45,35,5,45,15,5]'), (1586, '[1586,8,2,46,18,12,34,28]'), (1587, '[1587,11,9,7,31,29,3,1]'), (1588, '[1588,14,16,18,44,46,22,24]'), (1589, '[1589,17,23,29,7,13,41,47]'), (1590, '[1590,20,30,40,20,30,10,20]'), (1591, '[1591,23,37,1,33,47,29,43]'), (1592, '[1592,26,44,12,46,14,48,16]'), (1593, '[1593,29,1,23,9,31,17,39]'), (1594, '[1594,32,8,34,22,48,36,12]'), (1595, '[1595,35,15,45,35,15,5,35]'), (1596, '[1596,38,22,6,48,32,24,8]'), (1597, '[1597,41,29,17,11,49,43,31]'), (1598, '[1598,44,36,28,24,16,12,4]'), (1599, '[1599,47,43,39,37,33,31,27]'), (1600, '[1600,0,0,0,0,0,0,0]'), (1601, '[1601,3,7,11,13,17,19,23]'), (1602, '[1602,6,14,22,26,34,38,46]'), (1603, '[1603,9,21,33,39,1,7,19]'), (1604, '[1604,12,28,44,2,18,26,42]'), (1605, '[1605,15,35,5,15,35,45,15]'), (1606, '[1606,18,42,16,28,2,14,38]'), (1607, '[1607,21,49,27,41,19,33,11]'), (1608, '[1608,24,6,38,4,36,2,34]'), (1609, '[1609,27,13,49,17,3,21,7]'), (1610, '[1610,30,20,10,30,20,40,30]'), (1611, '[1611,33,27,21,43,37,9,3]'), (1612, '[1612,36,34,32,6,4,28,26]'), (1613, '[1613,39,41,43,19,21,47,49]'), (1614, '[1614,42,48,4,32,38,16,22]'), (1615, '[1615,45,5,15,45,5,35,45]'), (1616, '[1616,48,12,26,8,22,4,18]'), (1617, '[1617,1,19,37,21,39,23,41]'), (1618, '[1618,4,26,48,34,6,42,14]'), (1619, '[1619,7,33,9,47,23,11,37]'), (1620, '[1620,10,40,20,10,40,30,10]'), (1621, '[1621,13,47,31,23,7,49,33]'), (1622, '[1622,16,4,42,36,24,18,6]'), (1623, '[1623,19,11,3,49,41,37,29]'), (1624, '[1624,22,18,14,12,8,6,2]'), (1625, '[1625,25,25,25,25,25,25,25]'), (1626, '[1626,28,32,36,38,42,44,48]'), (1627, '[1627,31,39,47,1,9,13,21]'), (1628, '[1628,34,46,8,14,26,32,44]'), (1629, '[1629,37,3,19,27,43,1,17]'), (1630, '[1630,40,10,30,40,10,20,40]'), (1631, '[1631,43,17,41,3,27,39,13]'), (1632, '[1632,46,24,2,16,44,8,36]'), (1633, '[1633,49,31,13,29,11,27,9]'), (1634, '[1634,2,38,24,42,28,46,32]'), (1635, '[1635,5,45,35,5,45,15,5]'), (1636, '[1636,8,2,46,18,12,34,28]'), (1637, '[1637,11,9,7,31,29,3,1]'), (1638, '[1638,14,16,18,44,46,22,24]'), (1639, '[1639,17,23,29,7,13,41,47]'), (1640, '[1640,20,30,40,20,30,10,20]'), (1641, '[1641,23,37,1,33,47,29,43]'), (1642, '[1642,26,44,12,46,14,48,16]'), (1643, '[1643,29,1,23,9,31,17,39]'), (1644, '[1644,32,8,34,22,48,36,12]'), (1645, '[1645,35,15,45,35,15,5,35]'), (1646, '[1646,38,22,6,48,32,24,8]'), (1647, '[1647,41,29,17,11,49,43,31]'), (1648, '[1648,44,36,28,24,16,12,4]'), (1649, '[1649,47,43,39,37,33,31,27]'), (1650, '[1650,0,0,0,0,0,0,0]'), (1651, '[1651,3,7,11,13,17,19,23]'), (1652, '[1652,6,14,22,26,34,38,46]'), (1653, '[1653,9,21,33,39,1,7,19]'), (1654, '[1654,12,28,44,2,18,26,42]'), (1655, '[1655,15,35,5,15,35,45,15]'), (1656, '[1656,18,42,16,28,2,14,38]'), (1657, '[1657,21,49,27,41,19,33,11]'), (1658, '[1658,24,6,38,4,36,2,34]'), (1659, '[1659,27,13,49,17,3,21,7]'), (1660, '[1660,30,20,10,30,20,40,30]'), (1661, '[1661,33,27,21,43,37,9,3]'), (1662, '[1662,36,34,32,6,4,28,26]'), (1663, '[1663,39,41,43,19,21,47,49]'), (1664, '[1664,42,48,4,32,38,16,22]'), (1665, '[1665,45,5,15,45,5,35,45]'), (1666, '[1666,48,12,26,8,22,4,18]'), (1667, '[1667,1,19,37,21,39,23,41]'), (1668, '[1668,4,26,48,34,6,42,14]'), (1669, '[1669,7,33,9,47,23,11,37]'), (1670, '[1670,10,40,20,10,40,30,10]'), (1671, '[1671,13,47,31,23,7,49,33]'), (1672, '[1672,16,4,42,36,24,18,6]'), (1673, '[1673,19,11,3,49,41,37,29]'), (1674, '[1674,22,18,14,12,8,6,2]'), (1675, '[1675,25,25,25,25,25,25,25]'), (1676, '[1676,28,32,36,38,42,44,48]'), (1677, '[1677,31,39,47,1,9,13,21]'), (1678, '[1678,34,46,8,14,26,32,44]'), (1679, '[1679,37,3,19,27,43,1,17]'), (1680, '[1680,40,10,30,40,10,20,40]'), (1681, '[1681,43,17,41,3,27,39,13]'), (1682, '[1682,46,24,2,16,44,8,36]'), (1683, '[1683,49,31,13,29,11,27,9]'), (1684, '[1684,2,38,24,42,28,46,32]'), (1685, '[1685,5,45,35,5,45,15,5]'), (1686, '[1686,8,2,46,18,12,34,28]'), (1687, '[1687,11,9,7,31,29,3,1]'), (1688, '[1688,14,16,18,44,46,22,24]'), (1689, '[1689,17,23,29,7,13,41,47]'), (1690, '[1690,20,30,40,20,30,10,20]'), (1691, '[1691,23,37,1,33,47,29,43]'), (1692, '[1692,26,44,12,46,14,48,16]'), (1693, '[1693,29,1,23,9,31,17,39]'), (1694, '[1694,32,8,34,22,48,36,12]'), (1695, '[1695,35,15,45,35,15,5,35]'), (1696, '[1696,38,22,6,48,32,24,8]'), (1697, '[1697,41,29,17,11,49,43,31]'), (1698, '[1698,44,36,28,24,16,12,4]'), (1699, '[1699,47,43,39,37,33,31,27]'), (1700, '[1700,0,0,0,0,0,0,0]'), (1701, '[1701,3,7,11,13,17,19,23]'), (1702, '[1702,6,14,22,26,34,38,46]'), (1703, '[1703,9,21,33,39,1,7,19]'), (1704, '[1704,12,28,44,2,18,26,42]'), (1705, '[1705,15,35,5,15,35,45,15]'), (1706, '[1706,18,42,16,28,2,14,38]'), (1707, '[1707,21,49,27,41,19,33,11]'), (1708, '[1708,24,6,38,4,36,2,34]'), (1709, '[1709,27,13,49,17,3,21,7]'), (1710, '[1710,30,20,10,30,20,40,30]'), (1711, '[1711,33,27,21,43,37,9,3]'), (1712, '[1712,36,34,32,6,4,28,26]'), (1713, '[1713,39,41,43,19,21,47,49]'), (1714, '[1714,42,48,4,32,38,16,22]'), (1715, '[1715,45,5,15,45,5,35,45]'), (1716, '[1716,48,12,26,8,22,4,18]'), (1717, '[1717,1,19,37,21,39,23,41]'), (1718, '[1718,4,26,48,34,6,42,14]'), (1719, '[1719,7,33,9,47,23,11,37]'), (1720, '[1720,10,40,20,10,40,30,10]'), (1721, '[1721,13,47,31,23,7,49,33]'), (1722, '[1722,16,4,42,36,24,18,6]'), (1723, '[1723,19,11,3,49,41,37,29]'), (1724, '[1724,22,18,14,12,8,6,2]'), (1725, '[1725,25,25,25,25,25,25,25]'), (1726, '[1726,28,32,36,38,42,44,48]'), (1727, '[1727,31,39,47,1,9,13,21]'), (1728, '[1728,34,46,8,14,26,32,44]'), (1729, '[1729,37,3,19,27,43,1,17]'), (1730, '[1730,40,10,30,40,10,20,40]'), (1731, '[1731,43,17,41,3,27,39,13]'), (1732, '[1732,46,24,2,16,44,8,36]'), (1733, '[1733,49,31,13,29,11,27,9]'), (1734, '[1734,2,38,24,42,28,46,32]'), (1735, '[1735,5,45,35,5,45,15,5]'), (1736, '[1736,8,2,46,18,12,34,28]'), (1737, '[1737,11,9,7,31,29,3,1]'), (1738, '[1738,14,16,18,44,46,22,24]'), (1739, '[1739,17,23,29,7,13,41,47]'), (1740, '[1740,20,30,40,20,30,10,20]'), (1741, '[1741,23,37,1,33,47,29,43]'), (1742, '[1742,26,44,12,46,14,48,16]'), (1743, '[1743,29,1,23,9,31,17,39]'), (1744, '[1744,32,8,34,22,48,36,12]'), (1745, '[1745,35,15,45,35,15,5,35]'), (1746, '[1746,38,22,6,48,32,24,8]'), (1747, '[1747,41,29,17,11,49,43,31]'), (1748, '[1748,44,36,28,24,16,12,4]'), (1749, '[1749,47,43,39,37,33,31,27]'), (1750, '[1750,0,0,0,0,0,0,0]'), (1751, '[1751,3,7,11,13,17,19,23]'), (1752, '[1752,6,14,22,26,34,38,46]'), (1753, '[1753,9,21,33,39,1,7,19]'), (1754, '[1754,12,28,44,2,18,26,42]'), (1755, '[1755,15,35,5,15,35,45,15]'), (1756, '[1756,18,42,16,28,2,14,38]'), (1757, '[1757,21,49,27,41,19,33,11]'), (1758, '[1758,24,6,38,4,36,2,34]'), (1759, '[1759,27,13,49,17,3,21,7]'), (1760, '[1760,30,20,10,30,20,40,30]'), (1761, '[1761,33,27,21,43,37,9,3]'), (1762, '[1762,36,34,32,6,4,28,26]'), (1763, '[1763,39,41,43,19,21,47,49]'), (1764, '[1764,42,48,4,32,38,16,22]'), (1765, '[1765,45,5,15,45,5,35,45]'), (1766, '[1766,48,12,26,8,22,4,18]'), (1767, '[1767,1,19,37,21,39,23,41]'), (1768, '[1768,4,26,48,34,6,42,14]'), (1769, '[1769,7,33,9,47,23,11,37]'), (1770, '[1770,10,40,20,10,40,30,10]'), (1771, '[1771,13,47,31,23,7,49,33]'), (1772, '[1772,16,4,42,36,24,18,6]'), (1773, '[1773,19,11,3,49,41,37,29]'), (1774, '[1774,22,18,14,12,8,6,2]'), (1775, '[1775,25,25,25,25,25,25,25]'), (1776, '[1776,28,32,36,38,42,44,48]'), (1777, '[1777,31,39,47,1,9,13,21]'), (1778, '[1778,34,46,8,14,26,32,44]'), (1779, '[1779,37,3,19,27,43,1,17]'), (1780, '[1780,40,10,30,40,10,20,40]'), (1781, '[1781,43,17,41,3,27,39,13]'), (1782, '[1782,46,24,2,16,44,8,36]'), (1783, '[1783,49,31,13,29,11,27,9]'), (1784, '[1784,2,38,24,42,28,46,32]'), (1785, '[1785,5,45,35,5,45,15,5]'), (1786, '[1786,8,2,46,18,12,34,28]'), (1787, '[1787,11,9,7,31,29,3,1]'), (1788, '[1788,14,16,18,44,46,22,24]'), (1789, '[1789,17,23,29,7,13,41,47]'), (1790, '[1790,20,30,40,20,30,10,20]'), (1791, '[1791,23,37,1,33,47,29,43]'), (1792, '[1792,26,44,12,46,14,48,16]'), (1793, '[1793,29,1,23,9,31,17,39]'), (1794, '[1794,32,8,34,22,48,36,12]'), (1795, '[1795,35,15,45,35,15,5,35]'), (1796, '[1796,38,22,6,48,32,24,8]'), (1797, '[1797,41,29,17,11,49,43,31]'), (1798, '[1798,44,36,28,24,16,12,4]'), (1799, '[1799,47,43,39,37,33,31,27]'), (1800, '[1800,0,0,0,0,0,0,0]'), (1801, '[1801,3,7,11,13,17,19,23]'), (1802, '[1802,6,14,22,26,34,38,46]'), (1803, '[1803,9,21,33,39,1,7,19]'), (1804, '[1804,12,28,44,2,18,26,42]'), (1805, '[1805,15,35,5,15,35,45,15]'), (1806, '[1806,18,42,16,28,2,14,38]'), (1807, '[1807,21,49,27,41,19,33,11]'), (1808, '[1808,24,6,38,4,36,2,34]'), (1809, '[1809,27,13,49,17,3,21,7]'), (1810, '[1810,30,20,10,30,20,40,30]'), (1811, '[1811,33,27,21,43,37,9,3]'), (1812, '[1812,36,34,32,6,4,28,26]'), (1813, '[1813,39,41,43,19,21,47,49]'), (1814, '[1814,42,48,4,32,38,16,22]'), (1815, '[1815,45,5,15,45,5,35,45]'), (1816, '[1816,48,12,26,8,22,4,18]'), (1817, '[1817,1,19,37,21,39,23,41]'), (1818, '[1818,4,26,48,34,6,42,14]'), (1819, '[1819,7,33,9,47,23,11,37]'), (1820, '[1820,10,40,20,10,40,30,10]'), (1821, '[1821,13,47,31,23,7,49,33]'), (1822, '[1822,16,4,42,36,24,18,6]'), (1823, '[1823,19,11,3,49,41,37,29]'), (1824, '[1824,22,18,14,12,8,6,2]'), (1825, '[1825,25,25,25,25,25,25,25]'), (1826, '[1826,28,32,36,38,42,44,48]'), (1827, '[1827,31,39,47,1,9,13,21]'), (1828, '[1828,34,46,8,14,26,32,44]'), (1829, '[1829,37,3,19,27,43,1,17]'), (1830, '[1830,40,10,30,40,10,20,40]'), (1831, '[1831,43,17,41,3,27,39,13]'), (1832, '[1832,46,24,2,16,44,8,36]'), (1833, '[1833,49,31,13,29,11,27,9]'), (1834, '[1834,2,38,24,42,28,46,32]'), (1835, '[1835,5,45,35,5,45,15,5]'), (1836, '[1836,8,2,46,18,12,34,28]'), (1837, '[1837,11,9,7,31,29,3,1]'), (1838, '[1838,14,16,18,44,46,22,24]'), (1839, '[1839,17,23,29,7,13,41,47]'), (1840, '[1840,20,30,40,20,30,10,20]'), (1841, '[1841,23,37,1,33,47,29,43]'), (1842, '[1842,26,44,12,46,14,48,16]'), (1843, '[1843,29,1,23,9,31,17,39]'), (1844, '[1844,32,8,34,22,48,36,12]'), (1845, '[1845,35,15,45,35,15,5,35]'), (1846, '[1846,38,22,6,48,32,24,8]'), (1847, '[1847,41,29,17,11,49,43,31]'), (1848, '[1848,44,36,28,24,16,12,4]'), (1849, '[1849,47,43,39,37,33,31,27]'), (1850, '[1850,0,0,0,0,0,0,0]'), (1851, '[1851,3,7,11,13,17,19,23]'), (1852, '[1852,6,14,22,26,34,38,46]'), (1853, '[1853,9,21,33,39,1,7,19]'), (1854, '[1854,12,28,44,2,18,26,42]'), (1855, '[1855,15,35,5,15,35,45,15]'), (1856, '[1856,18,42,16,28,2,14,38]'), (1857, '[1857,21,49,27,41,19,33,11]'), (1858, '[1858,24,6,38,4,36,2,34]'), (1859, '[1859,27,13,49,17,3,21,7]'), (1860, '[1860,30,20,10,30,20,40,30]'), (1861, '[1861,33,27,21,43,37,9,3]'), (1862, '[1862,36,34,32,6,4,28,26]'), (1863, '[1863,39,41,43,19,21,47,49]'), (1864, '[1864,42,48,4,32,38,16,22]'), (1865, '[1865,45,5,15,45,5,35,45]'), (1866, '[1866,48,12,26,8,22,4,18]'), (1867, '[1867,1,19,37,21,39,23,41]'), (1868, '[1868,4,26,48,34,6,42,14]'), (1869, '[1869,7,33,9,47,23,11,37]'), (1870, '[1870,10,40,20,10,40,30,10]'), (1871, '[1871,13,47,31,23,7,49,33]'), (1872, '[1872,16,4,42,36,24,18,6]'), (1873, '[1873,19,11,3,49,41,37,29]'), (1874, '[1874,22,18,14,12,8,6,2]'), (1875, '[1875,25,25,25,25,25,25,25]'), (1876, '[1876,28,32,36,38,42,44,48]'), (1877, '[1877,31,39,47,1,9,13,21]'), (1878, '[1878,34,46,8,14,26,32,44]'), (1879, '[1879,37,3,19,27,43,1,17]'), (1880, '[1880,40,10,30,40,10,20,40]'), (1881, '[1881,43,17,41,3,27,39,13]'), (1882, '[1882,46,24,2,16,44,8,36]'), (1883, '[1883,49,31,13,29,11,27,9]'), (1884, '[1884,2,38,24,42,28,46,32]'), (1885, '[1885,5,45,35,5,45,15,5]'), (1886, '[1886,8,2,46,18,12,34,28]'), (1887, '[1887,11,9,7,31,29,3,1]'), (1888, '[1888,14,16,18,44,46,22,24]'), (1889, '[1889,17,23,29,7,13,41,47]'), (1890, '[1890,20,30,40,20,30,10,20]'), (1891, '[1891,23,37,1,33,47,29,43]'), (1892, '[1892,26,44,12,46,14,48,16]'), (1893, '[1893,29,1,23,9,31,17,39]'), (1894, '[1894,32,8,34,22,48,36,12]'), (1895, '[1895,35,15,45,35,15,5,35]'), (1896, '[1896,38,22,6,48,32,24,8]'), (1897, '[1897,41,29,17,11,49,43,31]'), (1898, '[1898,44,36,28,24,16,12,4]'), (1899, '[1899,47,43,39,37,33,31,27]'), (1900, '[1900,0,0,0,0,0,0,0]'), (1901, '[1901,3,7,11,13,17,19,23]'), (1902, '[1902,6,14,22,26,34,38,46]'), (1903, '[1903,9,21,33,39,1,7,19]'), (1904, '[1904,12,28,44,2,18,26,42]'), (1905, '[1905,15,35,5,15,35,45,15]'), (1906, '[1906,18,42,16,28,2,14,38]'), (1907, '[1907,21,49,27,41,19,33,11]'), (1908, '[1908,24,6,38,4,36,2,34]'), (1909, '[1909,27,13,49,17,3,21,7]'), (1910, '[1910,30,20,10,30,20,40,30]'), (1911, '[1911,33,27,21,43,37,9,3]'), (1912, '[1912,36,34,32,6,4,28,26]'), (1913, '[1913,39,41,43,19,21,47,49]'), (1914, '[1914,42,48,4,32,38,16,22]'), (1915, '[1915,45,5,15,45,5,35,45]'), (1916, '[1916,48,12,26,8,22,4,18]'), (1917, '[1917,1,19,37,21,39,23,41]'), (1918, '[1918,4,26,48,34,6,42,14]'), (1919, '[1919,7,33,9,47,23,11,37]'), (1920, '[1920,10,40,20,10,40,30,10]'), (1921, '[1921,13,47,31,23,7,49,33]'), (1922, '[1922,16,4,42,36,24,18,6]'), (1923, '[1923,19,11,3,49,41,37,29]'), (1924, '[1924,22,18,14,12,8,6,2]'), (1925, '[1925,25,25,25,25,25,25,25]'), (1926, '[1926,28,32,36,38,42,44,48]'), (1927, '[1927,31,39,47,1,9,13,21]'), (1928, '[1928,34,46,8,14,26,32,44]'), (1929, '[1929,37,3,19,27,43,1,17]'), (1930, '[1930,40,10,30,40,10,20,40]'), (1931, '[1931,43,17,41,3,27,39,13]'), (1932, '[1932,46,24,2,16,44,8,36]'), (1933, '[1933,49,31,13,29,11,27,9]'), (1934, '[1934,2,38,24,42,28,46,32]'), (1935, '[1935,5,45,35,5,45,15,5]'), (1936, '[1936,8,2,46,18,12,34,28]'), (1937, '[1937,11,9,7,31,29,3,1]'), (1938, '[1938,14,16,18,44,46,22,24]'), (1939, '[1939,17,23,29,7,13,41,47]'), (1940, '[1940,20,30,40,20,30,10,20]'), (1941, '[1941,23,37,1,33,47,29,43]'), (1942, '[1942,26,44,12,46,14,48,16]'), (1943, '[1943,29,1,23,9,31,17,39]'), (1944, '[1944,32,8,34,22,48,36,12]'), (1945, '[1945,35,15,45,35,15,5,35]'), (1946, '[1946,38,22,6,48,32,24,8]'), (1947, '[1947,41,29,17,11,49,43,31]'), (1948, '[1948,44,36,28,24,16,12,4]'), (1949, '[1949,47,43,39,37,33,31,27]'), (1950, '[1950,0,0,0,0,0,0,0]'), (1951, '[1951,3,7,11,13,17,19,23]'), (1952, '[1952,6,14,22,26,34,38,46]'), (1953, '[1953,9,21,33,39,1,7,19]'), (1954, '[1954,12,28,44,2,18,26,42]'), (1955, '[1955,15,35,5,15,35,45,15]'), (1956, '[1956,18,42,16,28,2,14,38]'), (1957, '[1957,21,49,27,41,19,33,11]'), (1958, '[1958,24,6,38,4,36,2,34]'), (1959, '[1959,27,13,49,17,3,21,7]'), (1960, '[1960,30,20,10,30,20,40,30]'), (1961, '[1961,33,27,21,43,37,9,3]'), (1962, '[1962,36,34,32,6,4,28,26]'), (1963, '[1963,39,41,43,19,21,47,49]'), (1964, '[1964,42,48,4,32,38,16,22]'), (1965, '[1965,45,5,15,45,5,35,45]'), (1966, '[1966,48,12,26,8,22,4,18]'), (1967, '[1967,1,19,37,21,39,23,41]'), (1968, '[1968,4,26,48,34,6,42,14]'), (1969, '[1969,7,33,9,47,23,11,37]'), (1970, '[1970,10,40,20,10,40,30,10]'), (1971, '[1971,13,47,31,23,7,49,33]'), (1972, '[1972,16,4,42,36,24,18,6]'), (1973, '[1973,19,11,3,49,41,37,29]'), (1974, '[1974,22,18,14,12,8,6,2]'), (1975, '[1975,25,25,25,25,25,25,25]'), (1976, '[1976,28,32,36,38,42,44,48]'), (1977, '[1977,31,39,47,1,9,13,21]'), (1978, '[1978,34,46,8,14,26,32,44]'), (1979, '[1979,37,3,19,27,43,1,17]'), (1980, '[1980,40,10,30,40,10,20,40]'), (1981, '[1981,43,17,41,3,27,39,13]'), (1982, '[1982,46,24,2,16,44,8,36]'), (1983, '[1983,49,31,13,29,11,27,9]'), (1984, '[1984,2,38,24,42,28,46,32]'), (1985, '[1985,5,45,35,5,45,15,5]'), (1986, '[1986,8,2,46,18,12,34,28]'), (1987, '[1987,11,9,7,31,29,3,1]'), (1988, '[1988,14,16,18,44,46,22,24]'), (1989, '[1989,17,23,29,7,13,41,47]'), (1990, '[1990,20,30,40,20,30,10,20]'), (1991, '[1991,23,37,1,33,47,29,43]'), (1992, '[1992,26,44,12,46,14,48,16]'), (1993, '[1993,29,1,23,9,31,17,39]'), (1994, '[1994,32,8,34,22,48,36,12]'), (1995, '[1995,35,15,45,35,15,5,35]'), (1996, '[1996,38,22,6,48,32,24,8]'), (1997, '[1997,41,29,17,11,49,43,31]'), (1998, '[1998,44,36,28,24,16,12,4]'), (1999, '[1999,47,43,39,37,33,31,27]'), (2000, '[2000,0,0,0,0,0,0,0]');",
                "CREATE INDEX ON big USING hnsw (v vector_l2_ops);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM big ORDER BY v <-> '[1234,2,38,24,42,28,46,32]' LIMIT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1234")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM (SELECT id FROM big ORDER BY v <-> '[1234,2,38,24,42,28,46,32]' LIMIT 100) sq;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("40")],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("Doltgres answers nearest-neighbor queries exactly, while hnsw stops after ef_search candidates"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET hnsw.ef_search = 100;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM big ORDER BY v <-> '[1234,2,38,24,42,28,46,32]' LIMIT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1234")],
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
