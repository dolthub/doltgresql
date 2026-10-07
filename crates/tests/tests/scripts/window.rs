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
fn test_window_functions() {
    run_scripts(&[
        ScriptTest {
            name: "native sum and row_number as window functions",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, grp INT, amt INT);",
                "INSERT INTO t VALUES (1, 1, 10), (2, 1, 20), (3, 1, 30), (4, 2, 5), (5, 2, 15);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, sum(amt) OVER (PARTITION BY grp ORDER BY id) FROM t ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("sum", INT8)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("30")],
                            &[T("3"), T("60")],
                            &[T("4"), T("5")],
                            &[T("5"), T("20")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, row_number() OVER (PARTITION BY grp ORDER BY id) FROM t ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("row_number", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                            &[T("4"), T("1")],
                            &[T("5"), T("2")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT grp, sum(amt) FROM t GROUP BY grp ORDER BY grp",
                    expected: Expected::Rows {
                        columns: &[Column("grp", INT4), Column("sum", INT8)],
                        rows: &[
                            &[T("1"), T("60")],
                            &[T("2"), T("20")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "distinct window aggregates are unsupported",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(DISTINCT *) OVER () FROM (VALUES (1)) AS t(v)",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "*""#, position: 23, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(DISTINCT v) OVER (ORDER BY id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) FROM (VALUES (1, 1), (2, 1), (3, 2)) AS t(id, v)",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DISTINCT is not implemented for window functions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(DISTINCT v) OVER (ORDER BY id ROWS BETWEEN 1 PRECEDING AND CURRENT ROW) FROM (VALUES (1, 1), (2, 1), (3, 2)) AS t(id, v)",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DISTINCT is not implemented for window functions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(DISTINCT v) OVER (ORDER BY id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) FROM (VALUES (1, 1.00::numeric), (2, 1.00::numeric), (3, 4.00::numeric)) AS t(id, v)",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DISTINCT is not implemented for window functions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT min(DISTINCT v) OVER (ORDER BY id ROWS BETWEEN 1 PRECEDING AND CURRENT ROW) FROM (VALUES (1, 1), (2, 1), (3, 2)) AS t(id, v)",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DISTINCT is not implemented for window functions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT max(DISTINCT v) OVER (ORDER BY id ROWS BETWEEN 1 PRECEDING AND CURRENT ROW) FROM (VALUES (1, 1), (2, 1), (3, 2)) AS t(id, v)",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DISTINCT is not implemented for window functions", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "basic window functions",
            set_up_script: &[
                "CREATE TABLE c (c_id INT PRIMARY KEY, bill TEXT);",
                "CREATE TABLE o (o_id INT PRIMARY KEY, c_id INT, ship TEXT);",
                "INSERT INTO c VALUES (1, 'CA'), (2, 'TX'), (3, 'MA'), (4, 'TX'), (5, NULL), (6, 'FL');",
                "INSERT INTO o VALUES (10, 1, 'CA'), (20, 1, 'CA'), (30, 1, 'CA'), (40, 2, 'CA'), (50, 2, 'TX'), (60, 2, NULL), (70, 4, 'WY'), (80, 4, NULL), (90, 6, 'WA');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT row_number() OVER () AS rn FROM o WHERE c_id=-999",
                    expected: Expected::Rows {
                        columns: &[Column("rn", INT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_number() OVER () AS rn FROM o WHERE c_id=1",
                    expected: Expected::Rows {
                        columns: &[Column("rn", INT8)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT rank() OVER () AS rnk FROM o WHERE c_id=-999",
                    expected: Expected::Rows {
                        columns: &[Column("rnk", INT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT o_id, c_id, rank() OVER (ORDER BY o_id) AS rnk FROM o WHERE c_id=1",
                    expected: Expected::Rows {
                        columns: &[Column("o_id", INT4), Column("c_id", INT4), Column("rnk", INT8)],
                        rows: &[
                            &[T("10"), T("1"), T("1")],
                            &[T("20"), T("1"), T("2")],
                            &[T("30"), T("1"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dense_rank() OVER () AS drnk FROM o WHERE c_id=-999",
                    expected: Expected::Rows {
                        columns: &[Column("drnk", INT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ship, dense_rank() OVER (ORDER BY ship) AS drnk FROM o WHERE c_id IN (1, 2) ORDER BY ship",
                    expected: Expected::Rows {
                        columns: &[Column("ship", TEXT), Column("drnk", INT8)],
                        rows: &[
                            &[T("CA"), T("1")],
                            &[T("CA"), T("1")],
                            &[T("CA"), T("1")],
                            &[T("CA"), T("1")],
                            &[T("TX"), T("2")],
                            &[Null, T("3")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT c_id AS c_c_id, bill FROM c) sq1, LATERAL (SELECT row_number() OVER () AS rownum FROM o WHERE c_id = c_c_id) sq2 ORDER BY c_c_id, bill, rownum",
                    expected: Expected::Rows {
                        columns: &[Column("c_c_id", INT4), Column("bill", TEXT), Column("rownum", INT8)],
                        rows: &[
                            &[T("1"), T("CA"), T("1")],
                            &[T("1"), T("CA"), T("2")],
                            &[T("1"), T("CA"), T("3")],
                            &[T("2"), T("TX"), T("1")],
                            &[T("2"), T("TX"), T("2")],
                            &[T("2"), T("TX"), T("3")],
                            &[T("4"), T("TX"), T("1")],
                            &[T("4"), T("TX"), T("2")],
                            &[T("6"), T("FL"), T("1")],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c_id, rank() OVER (ORDER BY c_id) AS rnk FROM c ORDER BY rnk DESC LIMIT 3",
                    expected: Expected::Rows {
                        columns: &[Column("c_id", INT4), Column("rnk", INT8)],
                        rows: &[
                            &[T("6"), T("6")],
                            &[T("5"), T("5")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c_id, rank() OVER (ORDER BY c_id) AS r FROM c ORDER BY r",
                    expected: Expected::Rows {
                        columns: &[Column("c_id", INT4), Column("r", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                            &[T("5"), T("5")],
                            &[T("6"), T("6")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT c_id, rank() OVER (ORDER BY c_id) AS r FROM c ORDER BY r",
                    expected: Expected::Rows {
                        columns: &[Column("c_id", INT4), Column("r", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                            &[T("5"), T("5")],
                            &[T("6"), T("6")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(CASE WHEN r > 0 THEN 1 ELSE 0 END) FROM (SELECT rank() OVER (ORDER BY c_id) AS r FROM c) t",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT8)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c_id, SUM(o_id) OVER (PARTITION BY c_id) AS s FROM o WHERE c_id = 1 ORDER BY o_id",
                    expected: Expected::Rows {
                        columns: &[Column("c_id", INT4), Column("s", INT8)],
                        rows: &[
                            &[T("1"), T("60")],
                            &[T("1"), T("60")],
                            &[T("1"), T("60")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "named window reference honors ORDER BY from the WINDOW clause",
            set_up_script: &[
                "CREATE TABLE t_named(id int, grp int, amt int);",
                "INSERT INTO t_named VALUES (1,1,10),(2,1,20),(3,2,5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, SUM(amt) OVER w AS s FROM t_named WINDOW w AS (PARTITION BY grp ORDER BY id) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("s", INT8)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("30")],
                            &[T("3"), T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, SUM(amt) OVER (PARTITION BY grp ORDER BY id) AS s FROM t_named ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("s", INT8)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("30")],
                            &[T("3"), T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "named window inheritance chain honors ORDER BY added by the child",
            set_up_script: &[
                "CREATE TABLE t_inherit(id int, grp int, amt int);",
                "INSERT INTO t_inherit VALUES (1,1,10),(2,1,20),(3,1,30),(4,2,5),(5,2,15);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, SUM(amt) OVER w2 AS s FROM t_inherit WINDOW w1 AS (PARTITION BY grp), w2 AS (w1 ORDER BY id) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("s", INT8)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("30")],
                            &[T("3"), T("60")],
                            &[T("4"), T("5")],
                            &[T("5"), T("20")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, SUM(amt) OVER (w1 ORDER BY id) AS s FROM t_inherit WINDOW w1 AS (PARTITION BY grp) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("s", INT8)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("30")],
                            &[T("3"), T("60")],
                            &[T("4"), T("5")],
                            &[T("5"), T("20")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, SUM(amt) OVER (PARTITION BY grp ORDER BY id) AS s FROM t_inherit ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("s", INT8)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("30")],
                            &[T("3"), T("60")],
                            &[T("4"), T("5")],
                            &[T("5"), T("20")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "window SUM/AVG wrapped in a subquery projection",
            set_up_script: &[
                "CREATE TABLE wrapper_probe (grp INT, val INT);",
                "INSERT INTO wrapper_probe VALUES (1, 10), (1, 20), (2, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT grp, val, grp_total FROM (SELECT grp, val, SUM(val) OVER (PARTITION BY grp) AS grp_total FROM wrapper_probe) sub ORDER BY grp, val;",
                    expected: Expected::Rows {
                        columns: &[Column("grp", INT4), Column("val", INT4), Column("grp_total", INT8)],
                        rows: &[
                            &[T("1"), T("10"), T("30")],
                            &[T("1"), T("20"), T("30")],
                            &[T("2"), T("5"), T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT grp, val, grp_avg FROM (SELECT grp, val, AVG(val) OVER (PARTITION BY grp) AS grp_avg FROM wrapper_probe) sub ORDER BY grp, val;",
                    expected: Expected::Rows {
                        columns: &[Column("grp", INT4), Column("val", INT4), Column("grp_avg", NUMERIC)],
                        rows: &[
                            &[T("1"), T("10"), T("15.0000000000000000")],
                            &[T("1"), T("20"), T("15.0000000000000000")],
                            &[T("2"), T("5"), T("5.0000000000000000")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RANGE frame with INTERVAL month boundary is calendar-correct",
            set_up_script: &[
                "CREATE TABLE month_edge (d DATE, v INT);",
                "INSERT INTO month_edge VALUES ('2022-01-31', 1), ('2022-02-28', 2), ('2022-03-01', 3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT sum(v) OVER (ORDER BY d RANGE BETWEEN UNBOUNDED PRECEDING AND INTERVAL '1' MONTH FOLLOWING) FROM month_edge ORDER BY d",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT8)],
                        rows: &[
                            &[T("3")],
                            &[T("6")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ntile and cume_dist ignore ties/frame and operate over the whole partition",
            set_up_script: &[
                "CREATE TABLE rank_ext (id INT PRIMARY KEY, grp INT, val INT);",
                "INSERT INTO rank_ext VALUES (1,1,10),(2,1,10),(3,1,20),(4,1,30),(5,2,5),(6,2,15);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, ntile(2) OVER (PARTITION BY grp ORDER BY val) FROM rank_ext ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("ntile", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("1")],
                            &[T("3"), T("2")],
                            &[T("4"), T("2")],
                            &[T("5"), T("1")],
                            &[T("6"), T("2")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, cume_dist() OVER (PARTITION BY grp ORDER BY val) FROM rank_ext ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("cume_dist", FLOAT8)],
                        rows: &[
                            &[T("1"), T("0.5")],
                            &[T("2"), T("0.5")],
                            &[T("3"), T("0.75")],
                            &[T("4"), T("1")],
                            &[T("5"), T("0.5")],
                            &[T("6"), T("1")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cume_dist() FROM rank_ext",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "window function cume_dist requires an OVER clause", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_number() FROM rank_ext",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "window function row_number requires an OVER clause", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT rank() FROM rank_ext",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "window function rank requires an OVER clause", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "multiple differently-framed numeric RANGE windows in one SELECT don't collide",
            set_up_script: &[
                "CREATE TABLE boundary_2 (id INT PRIMARY KEY, val INT);",
                "INSERT INTO boundary_2 VALUES (1,10),(2,20),(3,30);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT
					  sum(val) over (order by id range between 0 preceding and 0 following) as r0,
					  sum(val) over (order by id range between current row and 1 following) as r1foll,
					  sum(val) over (order by id range between unbounded preceding and current row) as runbndprec,
					  sum(val) over (order by id range between current row and unbounded following) as runbndfoll
					FROM boundary_2 ORDER BY id"#,
                    expected: Expected::Rows {
                        columns: &[Column("r0", INT8), Column("r1foll", INT8), Column("runbndprec", INT8), Column("runbndfoll", INT8)],
                        rows: &[
                            &[T("10"), T("30"), T("10"), T("60")],
                            &[T("20"), T("50"), T("30"), T("50")],
                            &[T("30"), T("30"), T("60"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_number() over (order by id) as rn1, row_number() over (order by id desc) as rn2 FROM boundary_2 ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("rn1", INT8), Column("rn2", INT8)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("2"), T("2")],
                            &[T("3"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nth_value respects the window's frame and returns the polymorphic argument type",
            set_up_script: &[
                "CREATE TABLE nv (id INT PRIMARY KEY, grp INT, val INT);",
                "INSERT INTO nv VALUES (1,1,100),(2,1,200),(3,1,300),(4,2,999);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, nth_value(val, 2) OVER (PARTITION BY grp ORDER BY id) FROM nv ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("nth_value", INT4)],
                        rows: &[
                            &[T("1"), Null],
                            &[T("2"), T("200")],
                            &[T("3"), T("200")],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, nth_value(val, 2) OVER (PARTITION BY grp ORDER BY id ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) FROM nv ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("nth_value", INT4)],
                        rows: &[
                            &[T("1"), T("200")],
                            &[T("2"), T("200")],
                            &[T("3"), T("200")],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "variance/stddev window functions over an int column",
            set_up_script: &[
                "CREATE TABLE t3038 (id BIGINT PRIMARY KEY, grp VARCHAR(10), val INT);",
                "INSERT INTO t3038 VALUES (1,'a',10), (2,'a',20), (3,'b',30), (4,'b',5), (5,'c',15), (6,'c',25);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, STDDEV_POP(val) OVER (ORDER BY grp) FROM t3038 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("stddev_pop", NUMERIC)],
                        rows: &[
                            &[T("1"), T("5.0000000000000000")],
                            &[T("2"), T("5.0000000000000000")],
                            &[T("3"), T("9.6014321848357602")],
                            &[T("4"), T("9.6014321848357602")],
                            &[T("5"), T("8.5391256382996653")],
                            &[T("6"), T("8.5391256382996653")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, STDDEV_SAMP(val) OVER (ORDER BY grp) FROM t3038 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("stddev_samp", NUMERIC)],
                        rows: &[
                            &[T("1"), T("7.0710678118654752")],
                            &[T("2"), T("7.0710678118654752")],
                            &[T("3"), T("11.0867789130417256")],
                            &[T("4"), T("11.0867789130417256")],
                            &[T("5"), T("9.3541434669348535")],
                            &[T("6"), T("9.3541434669348535")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, VAR_POP(val) OVER (ORDER BY grp) FROM t3038 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("var_pop", NUMERIC)],
                        rows: &[
                            &[T("1"), T("25.0000000000000000")],
                            &[T("2"), T("25.0000000000000000")],
                            &[T("3"), T("92.1875000000000000")],
                            &[T("4"), T("92.1875000000000000")],
                            &[T("5"), T("72.9166666666666667")],
                            &[T("6"), T("72.9166666666666667")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, VAR_SAMP(val) OVER (ORDER BY grp) FROM t3038 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("var_samp", NUMERIC)],
                        rows: &[
                            &[T("1"), T("50.0000000000000000")],
                            &[T("2"), T("50.0000000000000000")],
                            &[T("3"), T("122.9166666666666667")],
                            &[T("4"), T("122.9166666666666667")],
                            &[T("5"), T("87.5000000000000000")],
                            &[T("6"), T("87.5000000000000000")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "variance/stddev as GROUP BY aggregates, single-row, float8, and aliases",
            set_up_script: &[
                "CREATE TABLE t3038b (grp VARCHAR(10), val INT);",
                "INSERT INTO t3038b VALUES ('a',10), ('a',20), ('b',30), ('b',5), ('c',15), ('c',25);",
                "CREATE TABLE t3038_one (val INT);",
                "INSERT INTO t3038_one VALUES (42);",
                "CREATE TABLE t3038_f (val DOUBLE PRECISION);",
                "INSERT INTO t3038_f VALUES (10.0), (20.0);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT grp, VAR_POP(val), VAR_SAMP(val), STDDEV_POP(val), STDDEV_SAMP(val) FROM t3038b GROUP BY grp ORDER BY grp;",
                    expected: Expected::Rows {
                        columns: &[Column("grp", VARCHAR), Column("var_pop", NUMERIC), Column("var_samp", NUMERIC), Column("stddev_pop", NUMERIC), Column("stddev_samp", NUMERIC)],
                        rows: &[
                            &[T("a"), T("25.0000000000000000"), T("50.0000000000000000"), T("5.0000000000000000"), T("7.0710678118654752")],
                            &[T("b"), T("156.2500000000000000"), T("312.5000000000000000"), T("12.5000000000000000"), T("17.6776695296636881")],
                            &[T("c"), T("25.0000000000000000"), T("50.0000000000000000"), T("5.0000000000000000"), T("7.0710678118654752")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT VAR_POP(val), VAR_SAMP(val), STDDEV_POP(val), STDDEV_SAMP(val) FROM t3038_one;",
                    expected: Expected::Rows {
                        columns: &[Column("var_pop", NUMERIC), Column("var_samp", NUMERIC), Column("stddev_pop", NUMERIC), Column("stddev_samp", NUMERIC)],
                        rows: &[
                            &[T("0"), Null, T("0"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT VAR_POP(val), VAR_SAMP(val), STDDEV_POP(val), STDDEV_SAMP(val) FROM t3038_f;",
                    expected: Expected::Rows {
                        columns: &[Column("var_pop", FLOAT8), Column("var_samp", FLOAT8), Column("stddev_pop", FLOAT8), Column("stddev_samp", FLOAT8)],
                        rows: &[
                            &[T("25"), T("50"), T("5"), T("7.0710678118654755")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT VARIANCE(val), STDDEV(val) FROM t3038b WHERE grp = 'a';",
                    expected: Expected::Rows {
                        columns: &[Column("variance", NUMERIC), Column("stddev", NUMERIC)],
                        rows: &[
                            &[T("50.0000000000000000"), T("7.0710678118654752")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "variance/stddev over float8 avoid cancellation for large nearly-equal values",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT avg(x::float8), var_pop(x::float8), var_samp(x::float8), stddev_pop(x::float8), stddev_samp(x::float8) FROM (VALUES (100000003), (100000004), (100000006), (100000007)) v(x);",
                    expected: Expected::Rows {
                        columns: &[Column("avg", FLOAT8), Column("var_pop", FLOAT8), Column("var_samp", FLOAT8), Column("stddev_pop", FLOAT8), Column("stddev_samp", FLOAT8)],
                        rows: &[
                            &[T("100000005"), T("2.5"), T("3.3333333333333335"), T("1.5811388300841898"), T("1.8257418583505538")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(x::float8), var_pop(x::float8), var_samp(x::float8), stddev_pop(x::float8), stddev_samp(x::float8) FROM (VALUES (7000000000005), (7000000000007)) v(x);",
                    expected: Expected::Rows {
                        columns: &[Column("avg", FLOAT8), Column("var_pop", FLOAT8), Column("var_samp", FLOAT8), Column("stddev_pop", FLOAT8), Column("stddev_samp", FLOAT8)],
                        rows: &[
                            &[T("7000000000006"), T("1"), T("2"), T("1"), T("1.4142135623730951")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "variance/stddev over a real column, as GROUP BY aggregates and window functions",
            set_up_script: &[
                "CREATE TABLE t3038r (id INT PRIMARY KEY, grp VARCHAR(10), val REAL);",
                "INSERT INTO t3038r VALUES (1,'a',10), (2,'a',20), (3,'b',30), (4,'b',5), (5,'c',15), (6,'c',25);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT grp, VAR_POP(val), VAR_SAMP(val), STDDEV_POP(val), STDDEV_SAMP(val) FROM t3038r GROUP BY grp ORDER BY grp;",
                    expected: Expected::Rows {
                        columns: &[Column("grp", VARCHAR), Column("var_pop", FLOAT8), Column("var_samp", FLOAT8), Column("stddev_pop", FLOAT8), Column("stddev_samp", FLOAT8)],
                        rows: &[
                            &[T("a"), T("25"), T("50"), T("5"), T("7.0710678118654755")],
                            &[T("b"), T("156.25"), T("312.5"), T("12.5"), T("17.67766952966369")],
                            &[T("c"), T("25"), T("50"), T("5"), T("7.0710678118654755")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, VAR_POP(val) OVER (ORDER BY grp) FROM t3038r ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("var_pop", FLOAT8)],
                        rows: &[
                            &[T("1"), T("25")],
                            &[T("2"), T("25")],
                            &[T("3"), T("92.1875")],
                            &[T("4"), T("92.1875")],
                            &[T("5"), T("72.91666666666667")],
                            &[T("6"), T("72.91666666666667")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "a star argument is not a value",
            set_up_script: &[
                "CREATE TABLE tstar (id INT PRIMARY KEY, val INT);",
                "INSERT INTO tstar VALUES (1, 10), (2, 20);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT FIRST_VALUE(*) OVER (ORDER BY id) FROM tstar;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function first_value() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) OVER (ORDER BY id) FROM tstar;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
fn test_range_offset_and_window_chain_rules() {
    run_scripts(&[
        ScriptTest {
            name: "RANGE frames with offsets",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE rg (d date, n int, f float8, t text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rg VALUES ('2020-01-31', 1, 1.5, 'a'), ('2020-02-29', 2, 2.5, 'b'), ('2020-03-31', 4, NULL, 'c'), (NULL, NULL, 4.0, 'd'), ('2020-03-01', 7, 7.5, 'e');",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT d, n, sum(n) OVER (ORDER BY d RANGE BETWEEN INTERVAL '1 month' PRECEDING AND CURRENT ROW) FROM rg ORDER BY d;",
                    expected: Expected::Rows {
                        columns: &[Column("d", DATE), Column("n", INT4), Column("sum", INT8)],
                        rows: &[
                            &[T("2020-01-31"), T("1"), T("1")],
                            &[T("2020-02-29"), T("2"), T("3")],
                            &[T("2020-03-01"), T("7"), T("9")],
                            &[T("2020-03-31"), T("4"), T("13")],
                            &[Null, Null, Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n, sum(n) OVER (ORDER BY n RANGE BETWEEN 2 PRECEDING AND 1 FOLLOWING), sum(n) OVER (ORDER BY n DESC RANGE BETWEEN 2 PRECEDING AND 1 FOLLOWING) FROM rg ORDER BY n;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("sum", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("1"), T("3"), T("3")],
                            &[T("2"), T("3"), T("7")],
                            &[T("4"), T("6"), T("4")],
                            &[T("7"), T("7"), T("7")],
                            &[Null, Null, Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f, sum(n) OVER (ORDER BY f RANGE BETWEEN 1.5 PRECEDING AND 1 FOLLOWING) FROM rg ORDER BY f;",
                    expected: Expected::Rows {
                        columns: &[Column("f", FLOAT8), Column("sum", INT8)],
                        rows: &[
                            &[T("1.5"), T("3")],
                            &[T("2.5"), T("3")],
                            &[T("4"), T("2")],
                            &[T("7.5"), T("7")],
                            &[Null, T("4")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n, sum(n) OVER (ORDER BY n NULLS FIRST RANGE BETWEEN 1 FOLLOWING AND 3 FOLLOWING) FROM rg ORDER BY n;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("sum", INT8)],
                        rows: &[
                            &[T("1"), T("6")],
                            &[T("2"), T("4")],
                            &[T("4"), T("7")],
                            &[T("7"), Null],
                            &[Null, Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(n) OVER (ORDER BY n, f RANGE 1 PRECEDING) FROM rg;",
                    expected: Expected::Error(Diagnostic { code: "42P20", message: "RANGE with offset PRECEDING/FOLLOWING requires exactly one ORDER BY column", position: 20, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(n) OVER (ORDER BY t RANGE 1 PRECEDING) FROM rg;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "RANGE with offset PRECEDING/FOLLOWING is not supported for column type text", position: 38, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(n) OVER (ORDER BY n RANGE -1 PRECEDING) FROM rg;",
                    expected: Expected::Error(Diagnostic { code: "22013", message: "invalid preceding or following size in window function", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(n) OVER (ORDER BY n RANGE NULL PRECEDING) FROM rg;",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "frame starting offset must not be null", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(n) OVER (ORDER BY d RANGE 1 PRECEDING) FROM rg;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "RANGE with offset PRECEDING/FOLLOWING is not supported for column type date and offset type integer", hint: "Cast the offset value to an appropriate type.", position: 38, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(n) OVER (RANGE 1 PRECEDING) FROM rg;",
                    expected: Expected::Error(Diagnostic { code: "42P20", message: "RANGE with offset PRECEDING/FOLLOWING requires exactly one ORDER BY column", position: 20, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "named window chains",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE nw (id int, grp int, amt int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO nw VALUES (1, 1, 10), (2, 1, 20), (3, 1, 30), (4, 2, 5), (5, 2, 15);",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, sum(amt) OVER w3 FROM nw WINDOW w1 AS (PARTITION BY grp), w2 AS (w1 ORDER BY id), w3 AS (w2) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("sum", INT8)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("30")],
                            &[T("3"), T("60")],
                            &[T("4"), T("5")],
                            &[T("5"), T("20")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, sum(amt) OVER (w2 ROWS 1 PRECEDING) FROM nw WINDOW w1 AS (PARTITION BY grp), w2 AS (w1 ORDER BY id) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("sum", INT8)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("30")],
                            &[T("3"), T("50")],
                            &[T("4"), T("5")],
                            &[T("5"), T("20")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
