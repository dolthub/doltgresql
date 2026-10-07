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
fn test_select() {
    run_scripts(&[
        ScriptTest {
            name: "SELECT empty",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT;",
                    expected: Expected::Rows {
                        columns: &[],
                        rows: &[
                            &[],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SELECT DISTINCT ON",
            set_up_script: &[
                "CREATE TABLE test (v1 INT4, v2 INT4);",
                "INSERT INTO test VALUES (1, 3), (1, 4), (2, 3), (2, 4);",
                "CREATE TABLE test2 (v1 INT4, v2 INT4, v3 INT4);",
                "INSERT INTO test2 VALUES (1, 3, 5), (2, 3, 5), (1, 4, 5), (2, 4, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY v1, v2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("1"), T("4")],
                            &[T("2"), T("3")],
                            &[T("2"), T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT * FROM test ORDER BY v1, v2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("1"), T("4")],
                            &[T("2"), T("3")],
                            &[T("2"), T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v1) * FROM test ORDER BY v1, v2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v2) * FROM test ORDER BY v2, v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("1"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v1) * FROM test ORDER BY v2, v1;",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: "SELECT DISTINCT ON expressions must match initial ORDER BY expressions", position: 20, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v2) * FROM test ORDER BY v2 DESC, v1 DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("2"), T("4")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v2, v1) * FROM test2 ORDER BY v1, v2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4), Column("v3", INT4)],
                        rows: &[
                            &[T("1"), T("3"), T("5")],
                            &[T("1"), T("4"), T("5")],
                            &[T("2"), T("3"), T("5")],
                            &[T("2"), T("4"), T("5")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v2, v1) * FROM test2 ORDER BY v1, v2 DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4), Column("v3", INT4)],
                        rows: &[
                            &[T("1"), T("4"), T("5")],
                            &[T("1"), T("3"), T("5")],
                            &[T("2"), T("4"), T("5")],
                            &[T("2"), T("3"), T("5")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v2, v1) * FROM test2 ORDER BY v1, v2 LIMIT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4), Column("v3", INT4)],
                        rows: &[
                            &[T("1"), T("3"), T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v2, v1) * FROM test2 ORDER BY v1, v2 DESC LIMIT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4), Column("v3", INT4)],
                        rows: &[
                            &[T("1"), T("4"), T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v1, v2, v3) * FROM test2 ORDER BY v1, v2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4), Column("v3", INT4)],
                        rows: &[
                            &[T("1"), T("3"), T("5")],
                            &[T("1"), T("4"), T("5")],
                            &[T("2"), T("3"), T("5")],
                            &[T("2"), T("4"), T("5")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v3) v1 FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("keyless rows come back in Dolt's row hash order rather than Postgres' insertion order"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v1, v3) * FROM test2 ORDER BY v1, v2;",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: "SELECT DISTINCT ON expressions must match initial ORDER BY expressions", position: 24, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON(v2) * FROM test2 ORDER BY v1, v2;",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: "SELECT DISTINCT ON expressions must match initial ORDER BY expressions", position: 20, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALUES NULL type inference with ORDER BY",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT v FROM (VALUES (NULL), (2), (1)) AS t(v) ORDER BY v + 0 ASC NULLS FIRST;",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[Null],
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v FROM (VALUES (NULL), (2), (1)) AS t(v) ORDER BY v + 0 ASC NULLS LAST;",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v FROM (VALUES (NULL), (2), (1)) AS t(v) ORDER BY v + 0 DESC NULLS FIRST;",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[Null],
                            &[T("2")],
                            &[T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v FROM (VALUES (NULL), (2), (1)) AS t(v) ORDER BY v + 0 DESC NULLS LAST;",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("1")],
                            &[Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ORDER BY NULL ordering",
            set_up_script: &[
                "CREATE TABLE null_ordering (id INT4 PRIMARY KEY, a INT4, b INT4);",
                "INSERT INTO null_ordering VALUES (1, NULL, 1), (2, NULL, NULL), (3, 1, 1), (4, 1, NULL), (5, 2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM null_ordering ORDER BY a, id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM null_ordering ORDER BY a ASC NULLS FIRST, id;",
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
                    query: "SELECT id FROM null_ordering ORDER BY a ASC NULLS LAST, id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM null_ordering ORDER BY a NULLS FIRST, id;",
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
                    query: "SELECT id FROM null_ordering ORDER BY a NULLS LAST, id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("4")],
                            &[T("5")],
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM null_ordering ORDER BY a DESC, id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("5")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM null_ordering ORDER BY a DESC NULLS FIRST, id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("5")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM null_ordering ORDER BY a DESC NULLS LAST, id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("5")],
                            &[T("3")],
                            &[T("4")],
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM null_ordering ORDER BY a ASC NULLS LAST, b DESC NULLS FIRST;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("3")],
                            &[T("5")],
                            &[T("2")],
                            &[T("1")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, row_number() OVER (ORDER BY a ASC NULLS LAST, id) FROM null_ordering ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("row_number", INT8)],
                        rows: &[
                            &[T("1"), T("4")],
                            &[T("2"), T("5")],
                            &[T("3"), T("1")],
                            &[T("4"), T("2")],
                            &[T("5"), T("3")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(id ORDER BY a DESC NULLS FIRST, id) FROM null_ordering;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2,5,3,4}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "select large limit",
            assertions: &[
                ScriptTestAssertion {
                    query: "select 1 limit 18446744073709551615",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "bigint out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 limit -18446744073709551616",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "bigint out of range", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "select values",
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from (values(1,'峰哥',18),(2,'王哥',20),(3,'张哥',22));",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "VALUES in FROM must have an alias", hint: "For example, FROM (VALUES ...) [AS] foo.", position: 15, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from (values(1,'峰哥',18),(2,'王哥',20),(3,'张哥',22)) x(id,name,age);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT), Column("age", INT4)],
                        rows: &[
                            &[T("1"), T("峰哥"), T("18")],
                            &[T("2"), T("王哥"), T("20")],
                            &[T("3"), T("张哥"), T("22")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from (values(1,'峰哥',18),(2,'王哥',20),(3,'张哥',22)) x(id,name,age) limit $1;",
                    bind_vars: &[BindVar::Int(2)],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT), Column("age", INT4)],
                        rows: &[
                            &[T("1"), T("峰哥"), T("18")],
                            &[T("2"), T("王哥"), T("20")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SELECT with no select expressions",
            set_up_script: &[
                "CREATE TABLE mytable (pk int primary key);",
                "INSERT INTO mytable VALUES (1), (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select from mytable;",
                    expected: Expected::Rows {
                        columns: &[],
                        rows: &[
                            &[],
                            &[],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXISTS (SELECT FROM mytable where pk > 0);",
                    expected: Expected::Rows {
                        columns: &[Column("exists", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "String literal alias preserved in column name",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 42 as z, 'bob' as bob;",
                    expected: Expected::Rows {
                        columns: &[Column("z", INT4), Column("bob", TEXT)],
                        rows: &[
                            &[T("42"), T("bob")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'hello';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("hello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'a', 'b';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT), Column("?column?", TEXT)],
                        rows: &[
                            &[T("a"), T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"select 'select 1 as ones', 'select x.y, x.y*2 as double from generate_series(1,4) as x(y)'
union all
select 'drop table gexec_test', NULL
union all
select 'drop table gexec_test', 'select ''2000-01-01''::date as party_over'"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT), Column("?column?", TEXT)],
                        rows: &[
                            &[T("select 1 as ones"), T("select x.y, x.y*2 as double from generate_series(1,4) as x(y)")],
                            &[T("drop table gexec_test"), Null],
                            &[T("drop table gexec_test"), T("select '2000-01-01'::date as party_over")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "column names of functions called in FROM",
            set_up_script: &[
                "CREATE TABLE ft (id integer primary key);",
                "INSERT INTO ft VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT relid FROM pg_partition_ancestors('ft');",
                    expected: Expected::Rows {
                        columns: &[Column("relid", REGCLASS)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_partition_ancestors.relid FROM pg_partition_ancestors('ft');",
                    expected: Expected::Rows {
                        columns: &[Column("relid", REGCLASS)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT x FROM pg_partition_ancestors('ft') AS x;",
                    expected: Expected::Rows {
                        columns: &[Column("x", REGCLASS)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT relid FROM ft, pg_partition_ancestors('ft');",
                    expected: Expected::Rows {
                        columns: &[Column("relid", REGCLASS)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest FROM unnest(ARRAY[1]);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4)],
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
            name: "derived table with duplicate column names",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT 1 AS a, 'x' AS a) t;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("a", TEXT)],
                        rows: &[
                            &[T("1"), T("x")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT *, ROW_NUMBER() OVER () AS n FROM (SELECT 1 AS a, 'x' AS a) t;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("a", TEXT), Column("n", INT8)],
                        rows: &[
                            &[T("1"), T("x"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT 1, 2) t(a, a);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
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
fn test_collate_clauses() {
    run_scripts(&[
        ScriptTest {
            name: "COLLATE clauses",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT 'a' COLLATE "C", 'b' COLLATE pg_catalog.default, 'c' COLLATE "POSIX", pg_typeof('a' COLLATE "C");"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT), Column("?column?", TEXT), Column("?column?", TEXT), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("a"), T("b"), T("c"), T("unknown")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 'a' COLLATE "nope";"#,
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"collation "nope" for encoding "UTF8" does not exist"#, position: 12, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 1 COLLATE "C";"#,
                    expected: Expected::Error(Diagnostic { code: "42804", message: "collations are not supported by type integer", position: 10, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 'b' < 'a' COLLATE "C", 'B' < 'a' COLLATE "C";"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT x FROM (VALUES ('b'), ('A'), ('a')) v(x) ORDER BY x COLLATE "C";"#,
                    expected: Expected::Rows {
                        columns: &[Column("x", TEXT)],
                        rows: &[
                            &[T("A")],
                            &[T("a")],
                            &[T("b")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 'a' COLLATE ucs_basic, 'a'::name COLLATE "C", ARRAY['a'] COLLATE "C";"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT), Column("name", NAME), Column("array", TEXT_ARRAY)],
                        rows: &[
                            &[T("a"), T("a"), T("{a}")],
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
fn test_limit_alias_and_distinct_on_rules() {
    run_scripts(&[
        ScriptTest {
            name: "LIMIT and OFFSET convert their arguments by assignment",
            assertions: &[
                ScriptTestAssertion {
                    query: "select 1 limit 18446744073709551615;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "bigint out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 limit -18446744073709551616;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "bigint out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 limit 2.5;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 limit 'x';",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type bigint: "x""#, position: 16, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 limit '2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 limit true;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of LIMIT must be type bigint, not type boolean", position: 16, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 offset 1.5;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 limit null;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
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
            name: "subqueries and VALUES in FROM need aliases",
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from (values(1,'a',18),(2,'b',20));",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "VALUES in FROM must have an alias", hint: "For example, FROM (VALUES ...) [AS] foo.", position: 15, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from (select 1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "subquery in FROM must have an alias", hint: "For example, FROM (SELECT ...) [AS] foo.", position: 15, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from ( SELECT 1 ), (values (1));",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "subquery in FROM must have an alias", hint: "For example, FROM (SELECT ...) [AS] foo.", position: 15, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DISTINCT ON matches the ORDER BY prefix in any order",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE d (v1 INT4 PRIMARY KEY, v2 INT4, v3 INT4);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO d VALUES (1, 3, 5), (2, 3, 6), (3, 4, 5), (4, 4, 6);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON (v2, v3) * FROM d ORDER BY v3, v2, v1 DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4), Column("v3", INT4)],
                        rows: &[
                            &[T("1"), T("3"), T("5")],
                            &[T("3"), T("4"), T("5")],
                            &[T("2"), T("3"), T("6")],
                            &[T("4"), T("4"), T("6")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON (v3) * FROM d ORDER BY v2, v3;",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: "SELECT DISTINCT ON expressions must match initial ORDER BY expressions", position: 21, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON (v2, v1) * FROM d ORDER BY v2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4), Column("v3", INT4)],
                        rows: &[
                            &[T("1"), T("3"), T("5")],
                            &[T("2"), T("3"), T("6")],
                            &[T("3"), T("4"), T("5")],
                            &[T("4"), T("4"), T("6")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT DISTINCT ON (v2, v1) * FROM d ORDER BY v3, v2;",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: "SELECT DISTINCT ON expressions must match initial ORDER BY expressions", position: 21, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
