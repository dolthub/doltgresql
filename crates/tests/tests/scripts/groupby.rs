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
fn test_group_by() {
    run_scripts(&[
        ScriptTest {
            name: "Basic order by/group by cases",
            set_up_script: &[
                "create table members (id bigint primary key, team text);",
                "insert into members values (3,'red'), (4,'red'),(5,'orange'),(6,'orange'),(7,'orange'),(8,'purple');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select team as f from members order by id, f",
                    expected: Expected::Rows {
                        columns: &[Column("f", TEXT)],
                        rows: &[
                            &[T("red")],
                            &[T("red")],
                            &[T("orange")],
                            &[T("orange")],
                            &[T("orange")],
                            &[T("purple")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT team, COUNT(*) FROM members GROUP BY team ORDER BY 2",
                    expected: Expected::Rows {
                        columns: &[Column("team", TEXT), Column("count", INT8)],
                        rows: &[
                            &[T("purple"), T("1")],
                            &[T("red"), T("2")],
                            &[T("orange"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT team, COUNT(*) FROM members GROUP BY 1 ORDER BY 2",
                    expected: Expected::Rows {
                        columns: &[Column("team", TEXT), Column("count", INT8)],
                        rows: &[
                            &[T("purple"), T("1")],
                            &[T("red"), T("2")],
                            &[T("orange"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT team, COUNT(*) FROM members GROUP BY team ORDER BY columndoesnotexist",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "columndoesnotexist" does not exist"#, position: 59, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT t1.id as id FROM members AS t1 JOIN members AS t2 ON t1.id = t2.id WHERE t2.id > 0 ORDER BY t1.id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8)],
                        rows: &[
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                            &[T("7")],
                            &[T("8")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id as alias1, (SELECT alias1+1 group by alias1 having alias1 > 0) FROM members where id < 6;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "alias1" does not exist"#, position: 30, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, (SELECT UPPER(team) having id > 3) as upper_team FROM members where id < 6;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("upper_team", TEXT)],
                        rows: &[
                            &[T("3"), Null],
                            &[T("4"), T("RED")],
                            &[T("5"), T("ORANGE")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, (SELECT -1 as id having id < 10) as upper_team FROM members where id < 6;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("upper_team", INT4)],
                        rows: &[
                            &[T("3"), T("-1")],
                            &[T("4"), T("-1")],
                            &[T("5"), T("-1")],
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
fn test_outer_aggregate_rules() {
    run_scripts(&[
        ScriptTest {
            name: "aggregates of outer query columns",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE ta (a int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ta VALUES (1),(2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select avg((select avg(a1.a order by (select avg(a2.a) from ta a3)) from ta a1)) from ta a2;",
                    expected: Expected::Error(Diagnostic { code: "42803", message: "aggregate function calls cannot be nested", position: 46, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT max(ta.a)) FROM ta;",
                    expected: Expected::Rows {
                        columns: &[Column("max", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("an aggregate of only outer query columns belongs to the outer query, which Doltgres cannot group yet"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT max(x.a) FROM ta x) FROM ta;",
                    expected: Expected::Rows {
                        columns: &[Column("max", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select max(min(a)) from ta;",
                    expected: Expected::Error(Diagnostic { code: "42803", message: "aggregate function calls cannot be nested", position: 12, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_grouping_sets() {
    run_scripts(&[
        ScriptTest {
            name: "grouping sets, rollup, cube, and grouping",
            set_up_script: &[
                "CREATE TABLE gs (a INT, b INT, v INT);",
                "INSERT INTO gs VALUES (1, 1, 10), (1, 2, 20), (2, 1, 30);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT a, b, grouping(a, b), sum(v) FROM gs GROUP BY ROLLUP (a, b) ORDER BY 3, 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("grouping", INT4), Column("sum", INT8)],
                        rows: &[
                            &[T("1"), T("1"), T("0"), T("10")],
                            &[T("1"), T("2"), T("0"), T("20")],
                            &[T("2"), T("1"), T("0"), T("30")],
                            &[T("1"), Null, T("1"), T("30")],
                            &[T("2"), Null, T("1"), T("30")],
                            &[Null, Null, T("3"), T("60")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b, count(*) FROM gs GROUP BY CUBE (a, b) ORDER BY a NULLS FIRST, b NULLS FIRST;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("count", INT8)],
                        rows: &[
                            &[Null, Null, T("3")],
                            &[Null, T("1"), T("2")],
                            &[Null, T("2"), T("1")],
                            &[T("1"), Null, T("2")],
                            &[T("1"), T("1"), T("1")],
                            &[T("1"), T("2"), T("1")],
                            &[T("2"), Null, T("1")],
                            &[T("2"), T("1"), T("1")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b, sum(v) FROM gs GROUP BY GROUPING SETS ((a), (b), ()) ORDER BY a, b;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("sum", INT8)],
                        rows: &[
                            &[T("1"), Null, T("30")],
                            &[T("2"), Null, T("30")],
                            &[Null, T("1"), T("40")],
                            &[Null, T("2"), T("20")],
                            &[Null, Null, T("60")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM gs WHERE false GROUP BY ();",
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
                    query: "SELECT a, count(*) FROM gs GROUP BY DISTINCT ROLLUP (a), ROLLUP (a) ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("count", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("1")],
                            &[Null, T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT grouping(v) FROM gs GROUP BY a;",
                    expected: Expected::Error(Diagnostic { code: "42803", message: "arguments to GROUPING must be grouping expressions of the associated query level", position: 17, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
