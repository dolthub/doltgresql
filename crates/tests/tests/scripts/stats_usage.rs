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
fn test_stats_aggregates() {
    run_scripts(&[
        ScriptTest {
            name: "aggregate functions over dolt_statistics",
            set_up_script: &[
                "CREATE TABLE t (pk int primary key, c1 int);",
                "INSERT INTO t SELECT i, i % 7 FROM generate_series(1, 100) g(i);",
                "CREATE TABLE t2 (pk int primary key, c1 int);",
                "CREATE INDEX t2_c1_idx ON t2(c1);",
                "INSERT INTO t2 SELECT i, i % 3 FROM generate_series(1, 60) g(i);",
                "ANALYZE t;",
                "ANALYZE t2;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT sum(row_count) FROM dolt_statistics WHERE table_name = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT sum(row_count) FROM dolt_statistics WHERE table_name = 't2' AND index_name = 'primary';",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("60")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' numeric formatting.
                ScriptTestAssertion {
                    query: "SELECT avg(row_count) FROM dolt_statistics WHERE table_name = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("avg", NUMERIC)],
                        rows: &[
                            &[T("100.0000000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' numeric formatting.
                ScriptTestAssertion {
                    query: "SELECT avg(null_count) FROM dolt_statistics WHERE table_name = 't2';",
                    expected: Expected::Rows {
                        columns: &[Column("avg", NUMERIC)],
                        rows: &[
                            &[T("0.00000000000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT table_name, index_name, sum(row_count)::int FROM dolt_statistics GROUP BY table_name, index_name ORDER BY table_name, index_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("index_name", TEXT), Column("sum", INT4)],
                        rows: &[
                            &[T("t"), T("primary"), T("100")],
                            &[T("t2"), T("primary"), T("60")],
                            &[T("t2"), T("t2_c1_idx"), T("60")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT max(row_count), min(row_count), count(row_count) FROM dolt_statistics WHERE table_name = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("max", NUMERIC), Column("min", NUMERIC), Column("count", INT8)],
                        rows: &[
                            &[T("100"), T("100"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT sum(row_count) OVER () FROM dolt_statistics WHERE table_name = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT sum(distinct_count) FROM dolt_statistics WHERE table_name = 't' HAVING sum(distinct_count) > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT sum(a.row_count) FROM dolt_statistics a JOIN dolt_statistics b ON a.table_name = b.table_name AND a.index_name = b.index_name WHERE a.table_name = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT sum(row_count + null_count) FROM dolt_statistics WHERE table_name = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT abs(row_count) FROM dolt_statistics WHERE table_name = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("abs", NUMERIC)],
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
    ]);
}

#[test]
fn test_stats_usage() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_statistics contains reasonable values after ANALYZE",
            set_up_script: &[
                "CREATE TABLE big (pk int primary key, lowcard int, highcard int);",
                "CREATE INDEX big_lowcard_idx ON big(lowcard);",
                "CREATE INDEX big_highcard_idx ON big(highcard);",
                "INSERT INTO big SELECT i, i % 10, i FROM generate_series(1, 5000) g(i);",
                "CREATE TABLE small (pk int primary key, c1 int);",
                "INSERT INTO small SELECT i, i % 5 FROM generate_series(1, 10) g(i);",
                "ANALYZE big;",
                "ANALYZE small;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT index_name, sum(row_count)::int FROM dolt_statistics WHERE table_name = 'big' GROUP BY index_name ORDER BY index_name;",
                    expected: Expected::Rows {
                        columns: &[Column("index_name", TEXT), Column("sum", INT4)],
                        rows: &[
                            &[T("big_highcard_idx"), T("5000")],
                            &[T("big_lowcard_idx"), T("5000")],
                            &[T("primary"), T("5000")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT sum(distinct_count)::int FROM dolt_statistics WHERE table_name = 'big' AND index_name = 'primary';",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT4)],
                        rows: &[
                            &[T("5000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT sum(distinct_count)::int FROM dolt_statistics WHERE table_name = 'big' AND index_name = 'big_highcard_idx';",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT4)],
                        rows: &[
                            &[T("5000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT max(distinct_count) <= 10 AND sum(distinct_count) >= 10 FROM dolt_statistics WHERE table_name = 'big' AND index_name = 'big_lowcard_idx';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT sum(null_count)::int FROM dolt_statistics WHERE table_name IN ('big', 'small');",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DISTINCT index_name, columns FROM dolt_statistics WHERE table_name = 'big' ORDER BY index_name;",
                    expected: Expected::Rows {
                        columns: &[Column("index_name", TEXT), Column("columns", TEXT)],
                        rows: &[
                            &[T("big_highcard_idx"), T("highcard")],
                            &[T("big_lowcard_idx"), T("lowcard")],
                            &[T("primary"), T("pk")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT max(upper_bound::int) FROM dolt_statistics WHERE table_name = 'big' AND index_name = 'primary';",
                    expected: Expected::Rows {
                        columns: &[Column("max", INT4)],
                        rows: &[
                            &[T("5000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT index_name, sum(row_count)::int, sum(distinct_count)::int FROM dolt_statistics WHERE table_name = 'small' GROUP BY index_name ORDER BY index_name;",
                    expected: Expected::Rows {
                        columns: &[Column("index_name", TEXT), Column("sum", INT4), Column("sum", INT4)],
                        rows: &[
                            &[T("primary"), T("10"), T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ANALYZE refreshes statistics after data changes",
            set_up_script: &[
                "CREATE TABLE t (pk int primary key, c1 int);",
                "INSERT INTO t SELECT i, i % 7 FROM generate_series(1, 100) g(i);",
                "ANALYZE t;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT sum(row_count)::int FROM dolt_statistics WHERE table_name = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT4)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t SELECT i, i % 7 FROM generate_series(101, 300) g(i);",
                    expected: Expected::Tag("INSERT 0 200"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ANALYZE t;",
                    expected: Expected::Tag("ANALYZE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT sum(row_count)::int FROM dolt_statistics WHERE table_name = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT4)],
                        rows: &[
                            &[T("300")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "join planner puts small table on the scan side of a lookup join",
            set_up_script: &[
                "CREATE TABLE big (pk int primary key, lowcard int, highcard int);",
                "CREATE INDEX big_lowcard_idx ON big(lowcard);",
                "CREATE INDEX big_highcard_idx ON big(highcard);",
                "INSERT INTO big SELECT i, i % 10, i FROM generate_series(1, 5000) g(i);",
                "CREATE TABLE small (pk int primary key, c1 int);",
                "INSERT INTO small SELECT i, i % 5 FROM generate_series(1, 10) g(i);",
                "ANALYZE big;",
                "ANALYZE small;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM small JOIN big ON small.pk = big.highcard;",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "LookupJoin", left: "small", right: "big" }, PlanFact::FullScan { table: "small" }, PlanFact::IndexScan { table: "big", columns: &["highcard"], ranges: "" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM big JOIN small ON small.pk = big.highcard;",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "LookupJoin", left: "small", right: "big" }, PlanFact::FullScan { table: "small" }, PlanFact::IndexScan { table: "big", columns: &["highcard"], ranges: "" }]),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "join planner builds hash table from the small table in a hash join",
            set_up_script: &[
                "CREATE TABLE big (pk int primary key, val int);",
                "INSERT INTO big SELECT i, i % 100 FROM generate_series(1, 5000) g(i);",
                "CREATE TABLE small (pk int primary key, val int);",
                "INSERT INTO small SELECT i, i FROM generate_series(1, 10) g(i);",
                "ANALYZE big;",
                "ANALYZE small;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM big JOIN small ON big.val = small.val;",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "HashJoin", left: "big", right: "small" }, PlanFact::FullScan { table: "big" }, PlanFact::FullScan { table: "small" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM small JOIN big ON big.val = small.val;",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "HashJoin", left: "big", right: "small" }, PlanFact::FullScan { table: "big" }, PlanFact::FullScan { table: "small" }]),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "planner uses cardinality to select more selective index",
            set_up_script: &[
                "SET enable_seqscan = off;",
                "CREATE TABLE t (pk int primary key, lowcard int, highcard int);",
                "CREATE INDEX t_lowcard_idx ON t(lowcard);",
                "CREATE INDEX t_highcard_idx ON t(highcard);",
                "INSERT INTO t SELECT i, i % 10, i FROM generate_series(1, 5000) g(i);",
                "ANALYZE t;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM t WHERE lowcard = 3 AND highcard = 42;",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "t", columns: &["highcard"], ranges: "[{[42, 42]}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM t WHERE lowcard = 3 AND highcard > 0;",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "t", columns: &["lowcard"], ranges: "[{[3, 3]}]" }]),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "join planner uses histograms to reorder a filtered lookup join",
            set_up_script: &[
                "SET enable_seqscan = off;",
                "CREATE TABLE big (pk int primary key, val int, jc int);",
                "CREATE INDEX big_val_idx ON big(val);",
                "CREATE INDEX big_jc_idx ON big(jc);",
                "INSERT INTO big SELECT i, i, i % 1000 + 1 FROM generate_series(1, 5000) g(i);",
                "CREATE TABLE small (pk int primary key, c1 int);",
                "INSERT INTO small SELECT i, i FROM generate_series(1, 1000) g(i);",
                "ANALYZE big;",
                "ANALYZE small;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM small JOIN big ON big.jc = small.pk WHERE big.val > 4950;",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "LookupJoin", left: "big", right: "small" }, PlanFact::IndexScan { table: "big", columns: &["val"], ranges: "[{(4950, ∞)}]" }, PlanFact::IndexScan { table: "small", columns: &["pk"], ranges: "" }]),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
