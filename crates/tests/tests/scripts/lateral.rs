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
fn test_implicit_lateral_join() {
    run_scripts(&[
        ScriptTest {
            name: "Issue #3112: implicit lateral join for set-returning functions",
            set_up_script: &[
                "CREATE TABLE bug15 (a integer, b integer);",
                "CREATE INDEX bug15_ab ON bug15 (a, b);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT k.u FROM pg_index i, unnest(i.indkey) AS k(u) WHERE i.indexrelid = 'bug15_ab'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("u", INT2)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k.u FROM pg_index i, LATERAL unnest(i.indkey) AS k(u) WHERE i.indexrelid = 'bug15_ab'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("u", INT2)],
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
        ScriptTest {
            name: "implicit lateral unnest over user table",
            set_up_script: &[
                "CREATE TABLE t1 (id integer, arr integer[]);",
                "INSERT INTO t1 VALUES (1, ARRAY[10, 20]), (2, ARRAY[30]);",
                "CREATE TABLE t2 (id integer, name text);",
                "INSERT INTO t2 VALUES (1, 'one'), (2, 'two');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, u FROM t1, unnest(t1.arr) AS x(u) ORDER BY id, u;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4)],
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
                    query: "SELECT id, u FROM t1, LATERAL unnest(t1.arr) AS x(u) ORDER BY id, u;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4)],
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
                    query: "SELECT id, u FROM t1, unnest(arr) AS x(u) ORDER BY id, u;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4)],
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
                    query: "SELECT id, x.u FROM t1, unnest(t1.arr) AS x(u) ORDER BY id, x.u;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4)],
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
                    query: "SELECT * FROM t1, unnest(t1.arr) AS x(u) ORDER BY id, u;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("arr", INT4_ARRAY), Column("u", INT4)],
                        rows: &[
                            &[T("1"), T("{10,20}"), T("10")],
                            &[T("1"), T("{10,20}"), T("20")],
                            &[T("2"), T("{30}"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, unnest FROM t1, unnest(t1.arr) ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("unnest", INT4)],
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
                    query: "SELECT t2.name, u FROM t1, t2, unnest(t1.arr) AS x(u) WHERE t1.id = t2.id ORDER BY u;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("u", INT4)],
                        rows: &[
                            &[T("one"), T("10")],
                            &[T("one"), T("20")],
                            &[T("two"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, u FROM t1 CROSS JOIN unnest(t1.arr) AS x(u) ORDER BY id, u;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4)],
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
                    query: "SELECT id, u FROM t1 INNER JOIN unnest(t1.arr) AS x(u) ON true ORDER BY id, u;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4)],
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
                    query: "SELECT id, n FROM t1, generate_series(1, t1.id) AS g(n) ORDER BY id, n;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, u, ord FROM t1, unnest(t1.arr) WITH ORDINALITY AS x(u, ord) ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4), Column("ord", INT8)],
                        rows: &[
                            &[T("1"), T("10"), T("1")],
                            &[T("1"), T("20"), T("2")],
                            &[T("2"), T("30"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, u, ord FROM t1, LATERAL unnest(t1.arr) WITH ORDINALITY AS x(u, ord) ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4), Column("ord", INT8)],
                        rows: &[
                            &[T("1"), T("10"), T("1")],
                            &[T("1"), T("20"), T("2")],
                            &[T("2"), T("30"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, u FROM t1, unnest(ARRAY[7]) AS x(u) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4)],
                        rows: &[
                            &[T("1"), T("7")],
                            &[T("2"), T("7")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t1, (SELECT t1.id) s;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"invalid reference to FROM-clause entry for table "t1""#, detail: r#"There is an entry for table "t1", but it cannot be referenced from this part of the query."#, hint: "To reference that table, you must mark this subquery with LATERAL.", position: 27, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "left lateral join null-extends rows when the function returns no rows",
            set_up_script: &[
                "CREATE TABLE t3 (id integer, vals integer[]);",
                "INSERT INTO t3 VALUES (1, ARRAY[]::integer[]), (2, ARRAY[10, 20]), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT t3.id, u FROM t3 LEFT JOIN LATERAL unnest(t3.vals) AS x(u) ON true ORDER BY t3.id, u NULLS FIRST;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4)],
                        rows: &[
                            &[T("1"), Null],
                            &[T("2"), T("10")],
                            &[T("2"), T("20")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT t3.id, u FROM t3 LEFT JOIN LATERAL unnest(t3.vals) AS x(u) ON u > 10 ORDER BY t3.id, u NULLS FIRST;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4)],
                        rows: &[
                            &[T("1"), Null],
                            &[T("2"), T("20")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "function OUT-parameter column names survive without an explicit alias",
            set_up_script: &[
                "CREATE TABLE t4 (id integer primary key);",
                "INSERT INTO t4 VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT relid FROM pg_partition_ancestors('t4');",
                    expected: Expected::Rows {
                        columns: &[Column("relid", REGCLASS)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT relid FROM t4, pg_partition_ancestors('t4');",
                    expected: Expected::Rows {
                        columns: &[Column("relid", REGCLASS)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_partition_ancestors.relid FROM t4, pg_catalog.pg_partition_ancestors('t4');",
                    expected: Expected::Rows {
                        columns: &[Column("relid", REGCLASS)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT conrelid = 't4'::pg_catalog.regclass AS sametable, conname
  FROM pg_catalog.pg_constraint, pg_catalog.pg_partition_ancestors('t4')
 WHERE conrelid = relid AND contype = 'f' AND conparentid = 0
ORDER BY sametable DESC, conname;"#,
                    expected: Expected::Rows {
                        columns: &[Column("sametable", BOOL), Column("conname", NAME)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_lateral_rules() {
    run_scripts(&[
        ScriptTest {
            name: "lateral joins",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE lt (id INT PRIMARY KEY, n INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO lt VALUES (1, 2), (2, 3), (3, 0);",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lt.id, s.v FROM lt, LATERAL (SELECT lt.n * 10 AS v) s ORDER BY lt.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("20")],
                            &[T("2"), T("30")],
                            &[T("3"), T("0")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lt.id, g FROM lt, generate_series(1, lt.n) g ORDER BY lt.id, g;",
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
                    query: "SELECT lt.id, g FROM lt LEFT JOIN LATERAL generate_series(1, lt.n) g ON true ORDER BY lt.id, g;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("g", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("2")],
                            &[T("2"), T("1")],
                            &[T("2"), T("2")],
                            &[T("2"), T("3")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lt.id, s.c FROM lt, LATERAL (SELECT count(*) AS c FROM lt l2 WHERE l2.id <= lt.id) s ORDER BY lt.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lt.id FROM lt, (SELECT lt.n) s;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"invalid reference to FROM-clause entry for table "lt""#, detail: r#"There is an entry for table "lt", but it cannot be referenced from this part of the query."#, hint: "To reference that table, you must mark this subquery with LATERAL.", position: 31, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.id, b.id FROM lt a JOIN LATERAL (SELECT * FROM lt WHERE lt.id > a.id) b ON true ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("1"), T("3")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_lateral_subqueries() {
    run_scripts(&[
        ScriptTest {
            name: "lateral subqueries that read the left rows in their conditions, limits, and aggregates",
            set_up_script: &["CREATE TABLE lt (id INT PRIMARY KEY, n INT);", "INSERT INTO lt VALUES (1, 2), (2, 3), (3, 0);"],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT lt.id, s.c FROM lt, LATERAL (SELECT max(l2.id) AS c FROM lt l2 WHERE l2.id < lt.id) s ORDER BY lt.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), Null],
                            &[T("2"), T("1")],
                            &[T("3"), T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lt.id, s.id, s.n FROM lt LEFT JOIN LATERAL (SELECT l2.id, l2.n FROM lt l2 WHERE l2.n > lt.n ORDER BY l2.n LIMIT 1) s ON true ORDER BY lt.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("2"), Null, Null],
                            &[T("3"), T("1"), T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lt.id, s.x FROM lt, LATERAL (SELECT l2.id * 10 + lt.n AS x FROM lt l2 WHERE l2.id <> lt.id ORDER BY 1 DESC LIMIT 1) s ORDER BY lt.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("x", INT4)],
                        rows: &[
                            &[T("1"), T("32")],
                            &[T("2"), T("33")],
                            &[T("3"), T("20")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lt.id, s.c FROM lt LEFT JOIN LATERAL (SELECT count(*) AS c FROM lt l2 WHERE l2.n = lt.id) s ON s.c > 0 ORDER BY lt.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", INT8)],
                        rows: &[
                            &[T("1"), Null],
                            &[T("2"), T("1")],
                            &[T("3"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_correlated_in_subqueries() {
    run_scripts(&[
        ScriptTest {
            name: "correlated IN subqueries that become semi joins with lateral subqueries",
            set_up_script: &["CREATE TABLE lt (id INT PRIMARY KEY, n INT);", "INSERT INTO lt VALUES (1, 2), (2, 3), (3, 0);"],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT lt.id FROM lt WHERE lt.n IN (SELECT l2.id FROM lt l2 WHERE l2.id <> lt.id) ORDER BY 1;",
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
                    query: "SELECT lt.id FROM lt WHERE lt.n + 1 IN (SELECT l2.n FROM lt l2 WHERE l2.id > lt.id) ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lt.id FROM lt WHERE lt.n IN (SELECT l2.id FROM lt l2 WHERE l2.n <> lt.n ORDER BY l2.id DESC LIMIT 1) ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
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
fn test_lateral_placeholders() {
    run_scripts(&[
        ScriptTest {
            name: "lateral subqueries whose outputs or conditions read the enclosing row under outer joins",
            set_up_script: &[
                "CREATE TABLE lp1 (a INT, b INT, c TEXT);",
                "CREATE TABLE lp2 (a INT, b INT);",
                "CREATE TABLE lpk (i INT PRIMARY KEY);",
                "INSERT INTO lp1 VALUES (0, 0, 'x'), (1, 1, 'y'), (2, 2, 'z'), (3, 1, 'w');",
                "INSERT INTO lp2 VALUES (1, 1), (2, 3), (3, 2);",
                "INSERT INTO lpk VALUES (1), (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT t1.a, ss.t2a, ss.l FROM lp1 t1 LEFT JOIN LATERAL (SELECT t2.a AS t2a, LEAST(t1.a, t2.a, t3.b) AS l FROM lp1 t2 JOIN lp2 t3 ON t2.a = t3.b) ss ON t1.a = ss.t2a ORDER BY t1.a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("t2a", INT4), Column("l", INT4)],
                        rows: &[
                            &[T("0"), Null, Null],
                            &[T("1"), T("1"), T("1")],
                            &[T("2"), T("2"), T("2")],
                            &[T("3"), T("3"), T("3")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT k FROM (SELECT i, COALESCE(i, j) AS k FROM (SELECT i FROM lpk UNION ALL SELECT 0) u JOIN (SELECT 1 AS j LIMIT 1) l ON i = j) a RIGHT JOIN (SELECT 2 AS x) b ON TRUE JOIN (SELECT 3 AS y) c ON i IS NOT NULL) d, LATERAL (SELECT k AS kl LIMIT 1) e;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("kl", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM lp1 AS i41, LATERAL (SELECT 1 AS x FROM (SELECT i41.a AS lat, i42.a AS loc FROM lp2 AS i81, lp1 AS i42) AS ss1 RIGHT JOIN lp2 AS i43 ON (i43.a > 1) WHERE ss1.loc = ss1.lat) AS ss2 WHERE i41.a > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("18")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT g, ss.a FROM generate_series(1, 3) g, LATERAL (SELECT a FROM lp2 x WHERE g = a UNION ALL SELECT b FROM lp2 y WHERE g = b) ss ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("g", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                            &[T("3"), T("3")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM lp2 c LEFT JOIN (lp2 a LEFT JOIN (SELECT b.a AS q1, COALESCE(b.b, b2.a) AS x FROM lp2 b, lp1 b2 WHERE b.a < b2.a) ss1 ON a.b = ss1.q1 CROSS JOIN LATERAL (SELECT d.a AS q1, COALESCE(ss1.x, d.b) AS y FROM lp2 d) ss2) ON c.b = ss2.q1, LATERAL (SELECT * FROM lp1 i WHERE ss2.y > i.a) ss3;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("21")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT t1.a, s.a FROM lp2 t1 LEFT JOIN lp2 t2 ON t1.a = t2.b JOIN LATERAL (SELECT t1.a AS a UNION ALL SELECT t2.a AS a) s ON TRUE WHERE s.a IS NOT NULL ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("2"), T("3")],
                            &[T("3"), T("2")],
                            &[T("3"), T("3")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM lp1 a FULL JOIN LATERAL generate_series(0, a.a) g ON TRUE;",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: r#"invalid reference to FROM-clause entry for table "a""#, detail: "The combining JOIN type must be INNER or LEFT for a LATERAL reference.", position: 58, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM lp1 a RIGHT JOIN LATERAL (SELECT a.a) g ON TRUE;",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: r#"invalid reference to FROM-clause entry for table "a""#, detail: "The combining JOIN type must be INNER or LEFT for a LATERAL reference.", position: 48, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT g, ss.f1 FROM generate_series(0, 3) g, LATERAL (SELECT a AS f1 FROM lp1 WHERE a = ANY (SELECT b FROM lp2 WHERE lp2.a = g OFFSET 0)) ss ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("g", INT4), Column("f1", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("3")],
                            &[T("3"), T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v.a, s.a FROM lp1 v, LATERAL (SELECT x.a FROM lp2 x WHERE x.a = v.b AND x.b = ANY (SELECT y.b FROM lp2 y WHERE y.a = length(v.c) OFFSET 0)) s ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("3"), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v.a, s.a FROM lp1 v, LATERAL (SELECT x.a FROM lp2 x WHERE x.a = v.b AND EXISTS (SELECT 1 FROM lp2 y WHERE y.b = length(v.c) + 1 OFFSET 0)) s ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
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
    ]);
}
