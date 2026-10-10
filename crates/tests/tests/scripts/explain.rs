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
fn test_explain() {
    run_scripts(&[
        ScriptTest {
            name: "basic explain tests",
            set_up_script: &[
                "CREATE TABLE t (i INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM T;",
                    skip: Some("the Go test expects Postgres plan text, whose costs are implementation details"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"
EXPLAIN 
(
	ANALYZE, 
	VERBOSE, 
	COSTS, 
	SETTINGS,
	BUFFERS,
	WAL,
	TIMING,
	SUMMARY,
	FORMAT TEXT
) 
	SELECT * FROM t;
"#,
                    skip: Some("the Go test expects Postgres plan text, whose costs are implementation details"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"
EXPLAIN 
(
	ANALYZE ON, 
	VERBOSE OFF, 
	COSTS TRUE, 
	SETTINGS FALSE,
	BUFFERS,
	WAL,
	TIMING,
	SUMMARY,
	FORMAT TEXT
) 
	SELECT * FROM t;
"#,
                    skip: Some("the Go test expects Postgres plan text, whose costs are implementation details"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"
EXPLAIN 
(
	NOTAVALIDOPTION
) 
	SELECT * FROM t;
"#,
                    skip: Some("the Go test expects Postgres plan text, whose costs are implementation details"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_planned_joins() {
    run_scripts(&[
        ScriptTest {
            name: "lookup joins through secondary indexes",
            set_up_script: &["SET enable_seqscan = off;"],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE lj_big (id INT PRIMARY KEY, k SMALLINT, v INT, label TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX lj_big_k ON lj_big (k, v);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO lj_big SELECT i, (i % 500)::SMALLINT, i % 7, 'b' || i FROM generate_series(1, 3000) i;",
                    expected: Expected::Tag("INSERT 0 3000"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO lj_big VALUES (3001, NULL, 1, 'null key');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE lj_small (id INT PRIMARY KEY, k BIGINT, w INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO lj_small VALUES (1, 3, 1), (2, 40000, 2), (3, NULL, 3), (4, 499, 4), (5, 3, 5), (6, 7, 6);",
                    expected: Expected::Tag("INSERT 0 6"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT s.id, b.id, b.k, b.v FROM lj_small s JOIN lj_big b ON b.k = s.k ORDER BY s.id, b.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4), Column("k", INT2), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("3"), T("3"), T("3")],
                            &[T("1"), T("503"), T("3"), T("6")],
                            &[T("1"), T("1003"), T("3"), T("2")],
                            &[T("1"), T("1503"), T("3"), T("5")],
                            &[T("1"), T("2003"), T("3"), T("1")],
                            &[T("1"), T("2503"), T("3"), T("4")],
                            &[T("4"), T("499"), T("499"), T("2")],
                            &[T("4"), T("999"), T("499"), T("5")],
                            &[T("4"), T("1499"), T("499"), T("1")],
                            &[T("4"), T("1999"), T("499"), T("4")],
                            &[T("4"), T("2499"), T("499"), T("0")],
                            &[T("4"), T("2999"), T("499"), T("3")],
                            &[T("5"), T("3"), T("3"), T("3")],
                            &[T("5"), T("503"), T("3"), T("6")],
                            &[T("5"), T("1003"), T("3"), T("2")],
                            &[T("5"), T("1503"), T("3"), T("5")],
                            &[T("5"), T("2003"), T("3"), T("1")],
                            &[T("5"), T("2503"), T("3"), T("4")],
                            &[T("6"), T("7"), T("7"), T("0")],
                            &[T("6"), T("507"), T("7"), T("3")],
                            &[T("6"), T("1007"), T("7"), T("6")],
                            &[T("6"), T("1507"), T("7"), T("2")],
                            &[T("6"), T("2007"), T("7"), T("5")],
                            &[T("6"), T("2507"), T("7"), T("1")],
                        ],
                        tag: "SELECT 24",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT s.id, b.id FROM lj_small s JOIN lj_big b ON b.k = s.k AND b.v = s.w ORDER BY s.id, b.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("2003")],
                            &[T("4"), T("1999")],
                            &[T("5"), T("1503")],
                            &[T("6"), T("1007")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT s.id, b.id FROM lj_small s LEFT JOIN lj_big b ON b.k = s.k AND b.v = 2 ORDER BY s.id, b.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("1003")],
                            &[T("2"), Null],
                            &[T("3"), Null],
                            &[T("4"), T("499")],
                            &[T("5"), T("1003")],
                            &[T("6"), T("1507")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT s.id, b.id, b.label FROM lj_small s JOIN lj_big b ON b.k = s.k WHERE b.label LIKE 'b2%' ORDER BY s.id, b.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4), Column("label", TEXT)],
                        rows: &[
                            &[T("1"), T("2003"), T("b2003")],
                            &[T("1"), T("2503"), T("b2503")],
                            &[T("4"), T("2499"), T("b2499")],
                            &[T("4"), T("2999"), T("b2999")],
                            &[T("5"), T("2003"), T("b2003")],
                            &[T("5"), T("2503"), T("b2503")],
                            &[T("6"), T("2007"), T("b2007")],
                            &[T("6"), T("2507"), T("b2507")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT s.id, b.id FROM lj_big b JOIN lj_small s ON b.k = s.k WHERE b.id > 2000 ORDER BY s.id, b.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("2003")],
                            &[T("1"), T("2503")],
                            &[T("4"), T("2499")],
                            &[T("4"), T("2999")],
                            &[T("5"), T("2003")],
                            &[T("5"), T("2503")],
                            &[T("6"), T("2007")],
                            &[T("6"), T("2507")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT s.id, b.id FROM lj_small s JOIN lj_big b ON b.k = s.k;",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "LookupJoin", left: "lj_small", right: "lj_big" }, PlanFact::FullScan { table: "lj_small" }, PlanFact::IndexScan { table: "lj_big", columns: &["k", "v"], ranges: "" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM lj_small s JOIN lj_big b ON b.k = s.k;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("24")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "semi and anti joins from EXISTS",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE ex_a (id INT PRIMARY KEY, x INT, y INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE ex_b (id INT PRIMARY KEY, x INT, y INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ex_a VALUES (1, 1, 1), (2, 2, 2), (3, 3, 3), (4, NULL, 4), (5, 2, 5);",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ex_b VALUES (1, 1, 1), (2, 2, 9), (3, 2, 2), (4, NULL, 4);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ex_a WHERE EXISTS (SELECT 1 FROM ex_b WHERE ex_b.x = ex_a.x) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM ex_a WHERE EXISTS (SELECT 1 FROM ex_b WHERE ex_b.x = ex_a.x);",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "SemiJoin", left: "ex_a", right: "ex_b" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ex_a WHERE NOT EXISTS (SELECT 1 FROM ex_b WHERE ex_b.x = ex_a.x) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ex_a WHERE EXISTS (SELECT 1 FROM ex_b WHERE ex_b.x = ex_a.x AND ex_b.y = ex_a.y) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ex_a WHERE EXISTS (SELECT 1 FROM ex_b WHERE ex_b.x = ex_a.x AND ex_b.id > 2) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ex_a WHERE id > 1 AND NOT EXISTS (SELECT 1 FROM ex_b WHERE ex_b.id = ex_a.id) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ex_a WHERE EXISTS (SELECT 1 FROM ex_b WHERE ex_b.x = ex_a.x OR ex_b.y = ex_a.y) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("4")],
                            &[T("5")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ex_a WHERE EXISTS (SELECT 1 FROM ex_b WHERE ex_b.x = ex_a.x) AND NOT EXISTS (SELECT 1 FROM ex_b WHERE ex_b.y = ex_a.y) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM ex_b;",
                    expected: Expected::Tag("DELETE 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ex_a WHERE NOT EXISTS (SELECT 1 FROM ex_b WHERE ex_b.x = ex_a.x) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ex_a WHERE EXISTS (SELECT 1 FROM ex_b WHERE ex_b.x = ex_a.x) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "catalog index scans and lookups",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE cat_t1 (id INT PRIMARY KEY, v TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE cat_t2 (id INT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT relname FROM pg_catalog.pg_class WHERE oid = 'cat_t1'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME)],
                        rows: &[
                            &[T("cat_t1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT relname FROM pg_catalog.pg_class WHERE oid IN ('cat_t1'::regclass, 'cat_t2'::regclass) ORDER BY relname;",
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME)],
                        rows: &[
                            &[T("cat_t1")],
                            &[T("cat_t2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT typname FROM pg_catalog.pg_type WHERE oid IN (23, 25) ORDER BY typname;",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME)],
                        rows: &[
                            &[T("int4")],
                            &[T("text")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT typname FROM pg_catalog.pg_type WHERE oid > 22 AND oid < 26 ORDER BY oid;",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME)],
                        rows: &[
                            &[T("int4")],
                            &[T("regproc")],
                            &[T("text")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = 'public';",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID)],
                        rows: &[
                            &[T("2200")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c.relname, n.nspname FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid WHERE c.relname = 'cat_t1';",
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME), Column("nspname", NAME)],
                        rows: &[
                            &[T("cat_t1"), T("public")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c.relname, a.attname FROM pg_catalog.pg_class c JOIN pg_catalog.pg_attribute a ON c.oid = a.attrelid WHERE c.relname = 'cat_t1' AND a.attnum > 0 ORDER BY a.attnum;",
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME), Column("attname", NAME)],
                        rows: &[
                            &[T("cat_t1"), T("id")],
                            &[T("cat_t1"), T("v")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT t.typname, n.nspname FROM pg_catalog.pg_type t JOIN pg_catalog.pg_namespace n ON t.typnamespace = n.oid WHERE t.typname = 'int4';",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME), Column("nspname", NAME)],
                        rows: &[
                            &[T("int4"), T("pg_catalog")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pg_catalog.pg_index WHERE indrelid = 'cat_t1'::regclass;",
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
    ]);
}

#[test]
fn test_limited_joins() {
    run_scripts(&[
        ScriptTest {
            name: "a LIMIT over an unordered join looks rows up instead of hashing a whole input",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE lim_a (id INT PRIMARY KEY, v INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE lim_b (id INT PRIMARY KEY, a_id INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO lim_a SELECT i, i % 10 FROM generate_series(1, 5000) i;",
                    expected: Expected::Tag("INSERT 0 5000"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO lim_b SELECT i, 5001 - i FROM generate_series(1, 5000) i;",
                    expected: Expected::Tag("INSERT 0 5000"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT a.id, a.v FROM lim_a a, lim_a b WHERE a.id = b.id LIMIT 50;",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "LookupJoin", left: "lim_a", right: "lim_a" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT a.id, b.id FROM lim_a a, lim_b b WHERE a.id = b.a_id LIMIT 50;",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "LookupJoin", left: "lim_b", right: "lim_a" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), count(DISTINCT a_id) FROM (SELECT a.id, b.a_id FROM lim_a a, lim_b b WHERE a.id = b.a_id LIMIT 50) s WHERE id = a_id;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8)],
                        rows: &[
                            &[T("50"), T("50")],
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
fn test_empty_and_full_joins() {
    run_scripts(&[
        ScriptTest {
            name: "joins that find no rows and FULL JOIN conditions",
            set_up_script: &[
                "CREATE TABLE fj_a (x INT);",
                "CREATE TABLE fj_b (y INT);",
                "INSERT INTO fj_a VALUES (1), (2);",
                "INSERT INTO fj_b VALUES (1), (3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM fj_a JOIN fj_b ON false;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4), Column("y", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM fj_a LEFT JOIN fj_b ON fj_a.x = fj_b.y WHERE false;",
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
                    query: "SELECT * FROM fj_a FULL JOIN fj_b ON fj_a.x = fj_b.y ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4), Column("y", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), Null],
                            &[Null, T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM fj_a FULL JOIN fj_b ON fj_a.x < fj_b.y;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "FULL JOIN is only supported with merge-joinable or hash-joinable join conditions", ..E }),
                    skip: Some("the default planner runs a FULL JOIN on any condition"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_pulled_up_subqueries() {
    run_scripts(&[
        ScriptTest {
            name: "subqueries of join conditions and of subqueries' conditions",
            set_up_script: &[
                "CREATE TABLE sl_a (x INT PRIMARY KEY, v INT);",
                "CREATE TABLE sl_b (y INT, w INT);",
                "CREATE TABLE sl_c (z INT, u INT);",
                "INSERT INTO sl_a SELECT i, i % 7 FROM generate_series(1, 200) i;",
                "INSERT INTO sl_b SELECT i % 50, i FROM generate_series(1, 300) i;",
                "INSERT INTO sl_c SELECT i % 20, i % 3 FROM generate_series(1, 100) i;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM sl_a JOIN sl_b ON sl_a.x = sl_b.y AND sl_b.w IN (SELECT z FROM sl_c);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("19")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM sl_a LEFT JOIN sl_b ON sl_a.x = sl_b.y AND EXISTS (SELECT 1 FROM sl_c WHERE sl_c.z = sl_b.w);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("200")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM sl_a LEFT JOIN sl_b ON sl_a.x = sl_b.y AND EXISTS (SELECT 1 FROM sl_c WHERE sl_c.z = sl_a.v);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("445")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM sl_a WHERE EXISTS (SELECT 1 FROM sl_b WHERE sl_b.y = sl_a.x AND sl_b.w IN (SELECT z FROM sl_c));",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("19")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM sl_a WHERE EXISTS (SELECT 1 FROM sl_b WHERE sl_b.y = sl_a.x AND EXISTS (SELECT 1 FROM sl_c WHERE sl_c.z = sl_b.w AND sl_c.u = sl_a.v));",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM sl_a WHERE NOT EXISTS (SELECT 1 FROM sl_b WHERE sl_b.y = sl_a.x AND NOT EXISTS (SELECT 1 FROM sl_c WHERE sl_c.z = sl_b.w));",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("151")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM sl_a WHERE EXISTS (SELECT 1 FROM sl_b WHERE sl_b.y = sl_a.x AND sl_b.w > (SELECT max(u) FROM sl_c WHERE sl_c.z = sl_b.y));",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("19")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM sl_a RIGHT JOIN sl_b ON sl_a.x = sl_b.y AND sl_a.v IN (SELECT u FROM sl_c);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("300")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM sl_a FULL JOIN sl_b ON sl_a.x = sl_b.y AND sl_a.v IN (SELECT u FROM sl_c);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("479")],
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
fn test_group_by_unique_columns() {
    run_scripts(&[
        ScriptTest {
            name: "GROUP BY columns that a unique index's columns determine",
            set_up_script: &[
                "CREATE TABLE gk_t (id INT PRIMARY KEY, a INT, b TEXT);",
                "CREATE TABLE gk_u (k INT NOT NULL, j INT NOT NULL, c INT, UNIQUE (k, j));",
                "INSERT INTO gk_t SELECT i, i % 5, 'x' || (i % 3) FROM generate_series(1, 50) i;",
                "INSERT INTO gk_u SELECT i, i % 4, i % 6 FROM generate_series(1, 40) i;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, a, b, count(*) FROM gk_t GROUP BY id, a, b ORDER BY id LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4), Column("b", TEXT), Column("count", INT8)],
                        rows: &[
                            &[T("1"), T("1"), T("x1"), T("1")],
                            &[T("2"), T("2"), T("x2"), T("1")],
                            &[T("3"), T("3"), T("x0"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k, j, c, count(*) FROM gk_u GROUP BY k, j, c ORDER BY k LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("j", INT4), Column("c", INT4), Column("count", INT8)],
                        rows: &[
                            &[T("1"), T("1"), T("1"), T("1")],
                            &[T("2"), T("2"), T("2"), T("1")],
                            &[T("3"), T("3"), T("3"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b, count(*) FROM gk_t GROUP BY a, b ORDER BY a, b LIMIT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT), Column("count", INT8)],
                        rows: &[
                            &[T("0"), T("x0"), T("3")],
                            &[T("0"), T("x1"), T("3")],
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

#[test]
fn test_window_run_conditions() {
    run_scripts(&[
        ScriptTest {
            name: "restrictions on window functions of a subquery",
            set_up_script: &[
                "CREATE TABLE wr_t (id INT PRIMARY KEY, g INT, v INT);",
                "INSERT INTO wr_t SELECT i, i % 3, i * 7 % 11 FROM generate_series(1, 30) i;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT id, row_number() OVER (ORDER BY v, id) rn FROM wr_t) s WHERE rn <= 3 ORDER BY rn;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("rn", INT8)],
                        rows: &[
                            &[T("11"), T("1")],
                            &[T("22"), T("2")],
                            &[T("8"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT id, g, rank() OVER (PARTITION BY g ORDER BY v) r FROM wr_t) s WHERE r < 2 AND id > 3 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("g", INT4), Column("r", INT8)],
                        rows: &[
                            &[T("11"), T("2"), T("1")],
                            &[T("22"), T("1"), T("1")],
                            &[T("30"), T("0"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT id, count(*) OVER (ORDER BY id) c FROM wr_t) s WHERE 5 >= c ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                            &[T("5"), T("5")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT id, row_number() OVER (ORDER BY id) rn FROM wr_t) s WHERE rn = 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("rn", INT8)],
                        rows: &[
                            &[T("4"), T("4")],
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
fn test_planned_views_and_inlined_functions() {
    run_scripts(&[
        ScriptTest {
            name: "views whose plans test a condition once, and SQL functions of one expression",
            set_up_script: &[
                "CREATE TABLE pv_a (a INT);",
                "CREATE TABLE pv_b (b INT);",
                "CREATE TABLE pv_c (c INT);",
                "CREATE VIEW pv_v AS SELECT a FROM pv_a WHERE EXISTS (SELECT 1 FROM pv_b JOIN pv_c ON b = c);",
                "CREATE FUNCTION pv_plus1(x INT) RETURNS INT LANGUAGE SQL IMMUTABLE AS 'SELECT x + 1';",
                "CREATE FUNCTION pv_add(a INT, b INT) RETURNS INT LANGUAGE SQL STRICT IMMUTABLE RETURN a + b;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM pv_v;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO pv_a VALUES (1); INSERT INTO pv_b VALUES (2); INSERT INTO pv_c VALUES (2);",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pv_v;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pv_plus1(pv_plus1(a)), pv_add(a, NULL), pv_add(a, 2) FROM pv_a WHERE a = pv_plus1(0);",
                    expected: Expected::Rows {
                        columns: &[Column("pv_plus1", INT4), Column("pv_add", INT4), Column("pv_add", INT4)],
                        rows: &[
                            &[T("3"), Null, T("3")],
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
fn test_window_clause_frames() {
    run_scripts(&[
        ScriptTest {
            name: "ranking functions and aggregates over windows of different frames",
            set_up_script: &[
                "CREATE TABLE wf_t (id INT PRIMARY KEY, p INT, o INT);",
                "INSERT INTO wf_t SELECT i, i % 3, i % 7 FROM generate_series(1, 20) i;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, row_number() OVER (ORDER BY o, id), sum(o) OVER (ORDER BY o, id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW), rank() OVER (ORDER BY o) FROM wf_t ORDER BY id LIMIT 4;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("row_number", INT8), Column("sum", INT8), Column("rank", INT8)],
                        rows: &[
                            &[T("1"), T("3"), T("1"), T("3")],
                            &[T("2"), T("6"), T("5"), T("6")],
                            &[T("3"), T("9"), T("12"), T("9")],
                            &[T("4"), T("12"), T("22"), T("12")],
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

#[test]
fn test_lateral_functions() {
    run_scripts(&[
        ScriptTest {
            name: "functions and VALUES lists that read the FROM items before them",
            set_up_script: &[
                "CREATE TABLE lat_t (id INT PRIMARY KEY, n INT, arr INT[]);",
                "INSERT INTO lat_t VALUES (1, 2, '{10,20}'), (2, 3, '{30}'), (3, 0, '{}');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, g FROM lat_t, generate_series(1, lat_t.n) g ORDER BY id, g;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("g", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("2")],
                            &[T("2"), T("1")],
                            &[T("2"), T("2")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, x FROM lat_t, unnest(lat_t.arr) x ORDER BY id, x;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("x", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("1"), T("20")],
                            &[T("2"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lat_t.id, v.a FROM lat_t, LATERAL (VALUES (lat_t.id * 10), (lat_t.n)) v(a) ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("1"), T("10")],
                            &[T("2"), T("3")],
                            &[T("2"), T("20")],
                            &[T("3"), T("0")],
                            &[T("3"), T("30")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.id, b.id, g FROM lat_t a JOIN lat_t b ON a.id = b.id, generate_series(a.n, b.n + 1) g ORDER BY 1, 3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4), Column("g", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("2")],
                            &[T("1"), T("1"), T("3")],
                            &[T("2"), T("2"), T("3")],
                            &[T("2"), T("2"), T("4")],
                            &[T("3"), T("3"), T("0")],
                            &[T("3"), T("3"), T("1")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_lateral_chains_and_correlated_ctes() {
    run_scripts(&[
        ScriptTest {
            name: "functions that read several FROM items before them, and WITH queries that read an enclosing row",
            set_up_script: &[
                "CREATE TABLE lc_t (f1 INT);",
                "INSERT INTO lc_t VALUES (1), (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*), sum(r3) FROM generate_series(1, 2) r1, generate_series(r1, 3) r2, generate_series(r1 + r2, 5) r3;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("12"), T("49")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM generate_series(1, 2) r1, generate_series(r1, 3) r2, ROWS FROM (generate_series(10 + r1, 11), generate_series(10 + r2, 11)) ORDER BY 1, 2, 3, 4;",
                    expected: Expected::Rows {
                        columns: &[Column("r1", INT4), Column("r2", INT4), Column("generate_series", INT4), Column("generate_series", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("11"), T("11")],
                            &[T("1"), T("2"), T("11"), Null],
                            &[T("1"), T("3"), T("11"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (WITH cte(foo) AS (VALUES (f1)) SELECT (SELECT foo FROM cte)) FROM lc_t ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("foo", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
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

#[test]
fn test_memoized_lookup_filters() {
    run_scripts(&[
        ScriptTest {
            name: "a cached index lookup that filters the rows it finds",
            set_up_script: &[
                "CREATE TABLE mm_t (u INT, h INT, tw INT);",
                "CREATE INDEX mm_t_u ON mm_t (u);",
                "CREATE INDEX mm_t_h ON mm_t (h);",
                "INSERT INTO mm_t SELECT i, i % 100, i % 20 FROM generate_series(0, 9999) i;",
                "ANALYZE mm_t;",
                "SET enable_hashjoin = off;",
                "SET enable_mergejoin = off;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM mm_t t1 JOIN mm_t t2 ON t1.u = t2.h WHERE t1.tw = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("500")],
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
fn test_right_semi_join_order() {
    run_scripts(&[
        ScriptTest {
            name: "ordering by the column that an IN subquery matches, above a right semi join",
            set_up_script: &[
                "CREATE TABLE rs_a (a INT, b INT, c VARCHAR);",
                "CREATE TABLE rs_b (a INT, b INT, c VARCHAR);",
                "CREATE TABLE rs_c (a INT, b INT, c INT);",
                "INSERT INTO rs_a SELECT i, i % 25, to_char(i, 'FM0000') FROM generate_series(0, 599) i WHERE i % 2 = 0;",
                "INSERT INTO rs_b SELECT i % 25, i, to_char(i, 'FM0000') FROM generate_series(0, 599) i WHERE i % 3 = 0;",
                "INSERT INTO rs_c SELECT i, i, i % 25 FROM generate_series(0, 599, 2) i;",
                "ANALYZE rs_a;",
                "ANALYZE rs_b;",
                "ANALYZE rs_c;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT t1.* FROM rs_a t1 WHERE t1.a IN (SELECT t1.b FROM rs_b t1, rs_c t2 WHERE t1.a = 0 AND t1.b = (t2.a + t2.b)/2) AND t1.b = 0 ORDER BY t1.a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", VARCHAR)],
                        rows: &[
                            &[T("0"), T("0"), T("0000")],
                            &[T("150"), T("0"), T("0150")],
                            &[T("300"), T("0"), T("0300")],
                            &[T("450"), T("0"), T("0450")],
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
