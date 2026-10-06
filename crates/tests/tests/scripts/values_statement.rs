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
fn test_values_statement() {
    run_scripts(&[
        ScriptTest {
            name: "basic values statements",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES (1), (2), (3)) sqa;",
                    expected: Expected::Rows {
                        columns: &[Column("column1", INT4)],
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
                    query: "SELECT * FROM (VALUES (1, 2), (3, 4)) sqa;",
                    expected: Expected::Rows {
                        columns: &[Column("column1", INT4), Column("column2", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("3"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT i * 10, j * 100 FROM (VALUES (1, 2), (3, 4)) sqa(i, j);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4), Column("?column?", INT4)],
                        rows: &[
                            &[T("10"), T("200")],
                            &[T("30"), T("400")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with mixed int and decimal",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1),(2.01),(3)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1")],
                            &[T("2.01")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1.01),(2),(3)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1.01")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT SUM(n) FROM (VALUES(1),(2.01),(3)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("6.01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT SUM(n::numeric) FROM (VALUES(1),(2.01),(3)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("6.01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT SUM(n::numeric) FROM (VALUES(1.01),(2),(3)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("6.01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with multiple columns mixed types",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1, 'a'), (2.5, 'b')) v(num, str);",
                    expected: Expected::Rows {
                        columns: &[Column("num", NUMERIC), Column("str", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2.5"), T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with GROUP BY",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT n, COUNT(*) FROM (VALUES(1),(2.5),(1),(3.5),(2.5)) v(n) GROUP BY n ORDER BY n;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC), Column("count", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2.5"), T("2")],
                            &[T("3.5"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT category, SUM(amount) FROM (VALUES('a', 1),('b', 2.5),('a', 3),('b', 4.5)) v(category, amount) GROUP BY category ORDER BY category;",
                    expected: Expected::Rows {
                        columns: &[Column("category", TEXT), Column("sum", NUMERIC)],
                        rows: &[
                            &[T("a"), T("4")],
                            &[T("b"), T("7.0")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with DISTINCT",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT DISTINCT n FROM (VALUES(1),(2.5),(1),(2.5),(3)) v(n) ORDER BY n;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1")],
                            &[T("2.5")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with LIMIT and OFFSET",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1")],
                            &[T("2.5")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) LIMIT 2 OFFSET 2;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("3")],
                            &[T("4.5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with ORDER BY",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(3),(1.5),(2),(4.5)) v(n) ORDER BY n;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1.5")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4.5")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(3),(1.5),(2),(4.5)) v(n) ORDER BY n DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("4.5")],
                            &[T("3")],
                            &[T("2")],
                            &[T("1.5")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES in subquery",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT n * 2 AS doubled FROM (VALUES(1),(2.5),(3)) v(n)) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("doubled", NUMERIC)],
                        rows: &[
                            &[T("2")],
                            &[T("5.0")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES(1),(2.5),(3),(4.5)) v(n) LIMIT 2) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1")],
                            &[T("2.5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES(3),(1.5),(2)) v(n) ORDER BY n) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1.5")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with WHERE clause (Filter node)",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) WHERE n > 2;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("2.5")],
                            &[T("3")],
                            &[T("4.5")],
                            &[T("5")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) WHERE n > 1 AND n < 4.5;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("2.5")],
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
            name: "VALUES with aggregate functions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT AVG(n) FROM (VALUES(1),(2),(3),(4)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("avg", NUMERIC)],
                        rows: &[
                            &[T("2.5000000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT MIN(n), MAX(n) FROM (VALUES(1),(2.5),(3),(0.5)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("min", NUMERIC), Column("max", NUMERIC)],
                        rows: &[
                            &[T("0.5"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES combined operations",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT n, COUNT(*) as cnt FROM (VALUES(1),(2.5),(1),(2.5),(3),(1)) v(n) GROUP BY n ORDER BY cnt DESC LIMIT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC), Column("cnt", INT8)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("2.5"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT n FROM (VALUES(1),(2.5),(1),(3),(2.5),(4)) v(n) ORDER BY n DESC LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("4")],
                            &[T("3")],
                            &[T("2.5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) WHERE n > 1 ORDER BY n DESC LIMIT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("5")],
                            &[T("4.5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with single row (no type unification needed)",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(42)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(3.14)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("3.14")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with NULL values",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1),(NULL),(3)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("1")],
                            &[Null],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1.5),(NULL),(3.5)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1.5")],
                            &[Null],
                            &[T("3.5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1),(NULL),(2.5)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1")],
                            &[Null],
                            &[T("2.5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(NULL),(NULL)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", TEXT)],
                        rows: &[
                            &[Null],
                            &[Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES type mismatch errors",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1),('text'),(3)) v(n);",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "text""#, position: 27, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(true),(1),(false)) v(n);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "VALUES types boolean and integer cannot be matched", position: 30, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with all unknown types (string literals)",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES('a'),('b'),('c')) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", TEXT)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                            &[T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n || '!' FROM (VALUES('hello'),('world')) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("hello!")],
                            &[T("world!")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with array types",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(ARRAY[1,2]),(ARRAY[3,4])) v(arr);",
                    expected: Expected::Rows {
                        columns: &[Column("arr", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2}")],
                            &[T("{3,4}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(ARRAY['a','b']),(ARRAY['c','d'])) v(arr);",
                    expected: Expected::Rows {
                        columns: &[Column("arr", TEXT_ARRAY)],
                        rows: &[
                            &[T("{a,b}")],
                            &[T("{c,d}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with all same type multi-row (no casts needed)",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1),(2),(3)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
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
                    query: "SELECT * FROM (VALUES(1.5),(2.5),(3.5)) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1.5")],
                            &[T("2.5")],
                            &[T("3.5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES('x'),('y'),('z')) v(n);",
                    expected: Expected::Rows {
                        columns: &[Column("n", TEXT)],
                        rows: &[
                            &[T("x")],
                            &[T("y")],
                            &[T("z")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with multi-column partial cast",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1, 'a'),(2.5, 'b'),(3, 'c')) v(num, str);",
                    expected: Expected::Rows {
                        columns: &[Column("num", NUMERIC), Column("str", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2.5"), T("b")],
                            &[T("3"), T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(1, 10),(2, 20.5),(3, 30)) v(a, b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", NUMERIC)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("20.5")],
                            &[T("3"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES in CTE (WITH clause)",
            assertions: &[
                ScriptTestAssertion {
                    query: "WITH nums AS (SELECT * FROM (VALUES(1),(2.5),(3)) v(n)) SELECT * FROM nums;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1")],
                            &[T("2.5")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH nums AS (SELECT * FROM (VALUES(1),(2.5),(3)) v(n)) SELECT SUM(n) FROM nums;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("6.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with JOIN",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT a.n, b.label FROM (VALUES(1),(2),(3)) a(n) JOIN (VALUES(1, 'one'),(2, 'two'),(3, 'three')) b(id, label) ON a.n = b.id;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("label", TEXT)],
                        rows: &[
                            &[T("1"), T("one")],
                            &[T("2"), T("two")],
                            &[T("3"), T("three")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.n, b.label FROM (VALUES(1),(2.5),(3)) a(n) JOIN (VALUES(1, 'one'),(3, 'three')) b(id, label) ON a.n = b.id;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC), Column("label", TEXT)],
                        rows: &[
                            &[T("1"), T("one")],
                            &[T("3"), T("three")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with same-type booleans",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(true),(false),(true)) v(b);",
                    expected: Expected::Rows {
                        columns: &[Column("b", BOOL)],
                        rows: &[
                            &[T("t")],
                            &[T("f")],
                            &[T("t")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES(true),(false),(true),(false)) v(b) WHERE b = true;",
                    expected: Expected::Rows {
                        columns: &[Column("b", BOOL)],
                        rows: &[
                            &[T("t")],
                            &[T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with case-sensitive quoted column names",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT "ColA", "colb" FROM (VALUES(1, 2),(3.5, 4.5)) v("ColA", "colb");"#,
                    expected: Expected::Rows {
                        columns: &[Column("ColA", NUMERIC), Column("colb", NUMERIC)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("3.5"), T("4.5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT "MixedCase", plain FROM (VALUES(1, 'a'),(2.5, 'b')) v("MixedCase", plain);"#,
                    expected: Expected::Rows {
                        columns: &[Column("MixedCase", NUMERIC), Column("plain", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2.5"), T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT SUM("Val") FROM (VALUES(1),(2.5),(3)) v("Val");"#,
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC)],
                        rows: &[
                            &[T("6.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES with case-differing quoted columns and aggregates",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT SUM("Val"), SUM("val") FROM (VALUES(1, 10),(2.5, 20)) v("Val", "val");"#,
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC), Column("sum", INT8)],
                        rows: &[
                            &[T("3.5"), T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values inside subquery preserves projections",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT n * 2 AS doubled FROM (VALUES (1), (2), (3)) v(n)) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("doubled", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("4")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT n * 2 AS doubled FROM (VALUES (1), (2.5), (3)) v(n)) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("doubled", NUMERIC)],
                        rows: &[
                            &[T("2")],
                            &[T("5.0")],
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
            name: "values inside subquery preserves LIMIT",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (1), (2), (3), (4)) v(n) LIMIT 2) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (1), (2.5), (3), (4.5)) v(n) LIMIT 2) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1")],
                            &[T("2.5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values inside subquery preserves ORDER BY",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (3), (1), (2)) v(n) ORDER BY n) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
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
                    query: "SELECT * FROM (SELECT * FROM (VALUES (3), (1.5), (2)) v(n) ORDER BY n) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1.5")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values inside subquery preserves DISTINCT",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT DISTINCT * FROM (VALUES (1), (1), (2), (2), (3)) v(n)) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT DISTINCT * FROM (VALUES (1), (1), (2.5), (2.5), (3)) v(n)) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("2.5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values inside subquery preserves WHERE",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (1), (2), (3), (4), (5)) v(n) WHERE n > 3) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (1), (2.5), (3), (4.5), (5)) v(n) WHERE n > 3) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("4.5")],
                            &[T("5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values inside subquery preserves OFFSET",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (10), (20), (30)) v(n) OFFSET 1) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("20")],
                            &[T("30")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (10), (20), (30), (40), (50)) v(n) LIMIT 2 OFFSET 1) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("20")],
                            &[T("30")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (10), (20.5), (30)) v(n) OFFSET 1) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("20.5")],
                            &[T("30")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (10), (20.5), (30), (40.5), (50)) v(n) LIMIT 2 OFFSET 1) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("20.5")],
                            &[T("30")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values inside subquery preserves ORDER BY with LIMIT",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (5), (3), (1), (4), (2)) v(n) ORDER BY n LIMIT 3) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
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
                    query: "SELECT * FROM (SELECT * FROM (VALUES (5), (3.5), (1), (4), (2.5)) v(n) ORDER BY n LIMIT 3) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1")],
                            &[T("2.5")],
                            &[T("3.5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values inside subquery preserves GROUP BY with aggregate",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT n, count(*) AS cnt FROM (VALUES (1), (1), (2), (2), (2), (3)) v(n) GROUP BY n ORDER BY n) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("cnt", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("3"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT n, count(*) AS cnt FROM (VALUES (1), (1), (2.5), (2.5), (2.5), (3)) v(n) GROUP BY n ORDER BY n) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC), Column("cnt", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2.5"), T("3")],
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
            name: "values inside subquery preserves HAVING",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT n, count(*) AS cnt FROM (VALUES (1), (1), (2), (2), (2), (3)) v(n) GROUP BY n HAVING count(*) > 1 ORDER BY n) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("cnt", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT n, count(*) AS cnt FROM (VALUES (1), (1), (2.5), (2.5), (2.5), (3)) v(n) GROUP BY n HAVING count(*) > 1 ORDER BY n) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC), Column("cnt", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2.5"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values inside subquery preserves column aliasing",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT n AS val, n * 10 AS tenfold FROM (VALUES (1), (2), (3)) v(n)) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("val", INT4), Column("tenfold", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("20")],
                            &[T("3"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT n AS val, n * 10 AS tenfold FROM (VALUES (1), (2.5), (3)) v(n)) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("val", NUMERIC), Column("tenfold", NUMERIC)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2.5"), T("25.0")],
                            &[T("3"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values inside subquery preserves column subset selection",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT a FROM (VALUES (1, 10), (2, 20), (3, 30)) v(a, b)) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values subquery trivial SELECT * still unwraps correctly",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM (VALUES (1), (2), (3)) v(n)) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "values inside subquery with multiple combined clauses",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT DISTINCT n FROM (VALUES (3), (1), (1), (2), (2), (3)) v(n) ORDER BY n LIMIT 2) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT DISTINCT n FROM (VALUES (3), (1.5), (1.5), (2), (2), (3)) v(n) ORDER BY n LIMIT 2) sub;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC)],
                        rows: &[
                            &[T("1.5")],
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
            name: "values in JOIN preserves inner subquery semantics",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT a.n, b.m FROM (VALUES (1), (2)) a(n) JOIN (SELECT m * 10 AS m FROM (VALUES (1), (2)) v(m)) b ON a.n = b.m / 10;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("m", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("20")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.n, b.m FROM (VALUES (1), (2.5)) a(n) JOIN (SELECT m * 10 AS m FROM (VALUES (1), (2.5)) v(m)) b ON a.n = b.m / 10;",
                    expected: Expected::Rows {
                        columns: &[Column("n", NUMERIC), Column("m", NUMERIC)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2.5"), T("25.0")],
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
