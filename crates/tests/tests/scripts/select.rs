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
            name: "column names of schema-qualified functions",
            set_up_script: &[
                "CREATE SCHEMA function_labels;",
                "CREATE FUNCTION public.label_fn(integer) RETURNS integer LANGUAGE SQL AS 'SELECT $1 + 1';",
                "CREATE FUNCTION function_labels.label_fn(integer) RETURNS integer LANGUAGE SQL AS 'SELECT $1 + 2';",
                r#"CREATE FUNCTION function_labels."label.fn"(integer) RETURNS integer LANGUAGE SQL AS 'SELECT $1 + 3';"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT public.label_fn(1), function_labels.label_fn(1);",
                    expected: Expected::Rows {
                        columns: &[Column("label_fn", INT4), Column("label_fn", INT4)],
                        rows: &[
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT function_labels.label_fn(1) AS custom_label;",
                    expected: Expected::Rows {
                        columns: &[Column("custom_label", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT function_labels."label.fn"(1);"#,
                    expected: Expected::Rows {
                        columns: &[Column("label.fn", INT4)],
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

#[test]
fn test_equality_join_rules() {
    run_scripts(&[
        ScriptTest {
            name: "joins on equal values",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE jl (a int, b text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE jr (a int, b text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO jl VALUES (1, 'x'), (2, 'y'), (NULL, 'z'), (2, 'w');",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO jr VALUES (2, 'y'), (1, 'q'), (NULL, 'z'), (3, 'w'), (2, 'y2');",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM jl JOIN jr ON jl.a = jr.a ORDER BY 1, 2, 3, 4;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT), Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("x"), T("1"), T("q")],
                            &[T("2"), T("w"), T("2"), T("y")],
                            &[T("2"), T("w"), T("2"), T("y2")],
                            &[T("2"), T("y"), T("2"), T("y")],
                            &[T("2"), T("y"), T("2"), T("y2")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM jl LEFT JOIN jr ON jl.a = jr.a AND jl.b = jr.b ORDER BY 1, 2, 3, 4;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT), Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("x"), Null, Null],
                            &[T("2"), T("w"), Null, Null],
                            &[T("2"), T("y"), T("2"), T("y")],
                            &[Null, T("z"), Null, Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM jl FULL JOIN jr ON jr.a = jl.a + 0 ORDER BY 1, 2, 3, 4;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT), Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("x"), T("1"), T("q")],
                            &[T("2"), T("w"), T("2"), T("y")],
                            &[T("2"), T("w"), T("2"), T("y2")],
                            &[T("2"), T("y"), T("2"), T("y")],
                            &[T("2"), T("y"), T("2"), T("y2")],
                            &[Null, T("z"), Null, Null],
                            &[Null, Null, T("3"), T("w")],
                            &[Null, Null, Null, T("z")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM jl RIGHT JOIN jr USING (b) ORDER BY 1, 2, 3;",
                    expected: Expected::Rows {
                        columns: &[Column("b", TEXT), Column("a", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("q"), Null, T("1")],
                            &[T("w"), T("2"), T("3")],
                            &[T("y"), T("2"), T("2")],
                            &[T("y2"), Null, T("2")],
                            &[T("z"), Null, Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM jl JOIN jr ON jl.a::bigint = jr.a::smallint;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
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
fn test_select_into() {
    run_scripts(&[
        ScriptTest {
            name: "select into",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE src (f1 int4);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO src VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f1 INTO newt FROM src;",
                    expected: Expected::Tag("SELECT 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM newt;",
                    expected: Expected::Rows {
                        columns: &[Column("f1", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT f1 INTO x FROM src) s;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "SELECT ... INTO is not allowed here", position: 31, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE int4_tbl (f1 int4);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO int4_tbl SELECT 1 INTO f;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "SELECT ... INTO is not allowed here", position: 36, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM int4_tbl;",
                    expected: Expected::Rows {
                        columns: &[Column("f1", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 AS a INTO TEMP tmp1;",
                    expected: Expected::Tag("SELECT 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tmp1;",
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
                    query: "SELECT 2 AS b INTO UNLOGGED u1;",
                    expected: Expected::Tag("SELECT 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * INTO newt FROM src;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "newt" already exists"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_strict_null_constants() {
    run_scripts(&[
        ScriptTest {
            name: "strict operators on NULL constants",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE tab0(pk INTEGER PRIMARY KEY, col0 INTEGER, col3 INTEGER, col4 FLOAT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tab0 VALUES (1, 2000000, 30000, 1.5);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT - 76 AS col4 FROM tab0 AS cor0 WHERE NULL BETWEEN - col0 * + - 73 * 37 * col3 AND NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("col4", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ALL * FROM tab0 WHERE ( NULL ) <> - 39 * - col4 / + 0;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("col0", INT4), Column("col3", INT4), Column("col4", FLOAT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tab0 WHERE NULL = 1/0;",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tab0 WHERE col0 * col3 * 1000 > 0;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "integer out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tab0 WHERE NULL IS DISTINCT FROM col0 / 0;",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_joins_looking_up_primary_keys() {
    run_scripts(&[
        ScriptTest {
            name: "joins that look rows up by primary key",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE jl (id INT PRIMARY KEY, v TEXT, w INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE jr (id INT PRIMARY KEY, lid BIGINT, u TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE js (k SMALLINT PRIMARY KEY, name TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO jl SELECT i, 'v' || i, i % 10 FROM generate_series(1, 64) i;",
                    expected: Expected::Tag("INSERT 0 64"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO jr SELECT i, CASE WHEN i % 9 = 0 THEN NULL ELSE i % 80 END, 'u' || i FROM generate_series(1, 300) i;",
                    expected: Expected::Tag("INSERT 0 300"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO js SELECT i, 'n' || i FROM generate_series(1, 20) i;",
                    expected: Expected::Tag("INSERT 0 20"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), sum(l.w) FROM jr r JOIN jl l ON l.id = r.lid;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("223"), T("965")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r.id, l.v FROM jr r JOIN jl l ON l.id = r.lid WHERE r.id < 20 ORDER BY r.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("v1")],
                            &[T("2"), T("v2")],
                            &[T("3"), T("v3")],
                            &[T("4"), T("v4")],
                            &[T("5"), T("v5")],
                            &[T("6"), T("v6")],
                            &[T("7"), T("v7")],
                            &[T("8"), T("v8")],
                            &[T("10"), T("v10")],
                            &[T("11"), T("v11")],
                            &[T("12"), T("v12")],
                            &[T("13"), T("v13")],
                            &[T("14"), T("v14")],
                            &[T("15"), T("v15")],
                            &[T("16"), T("v16")],
                            &[T("17"), T("v17")],
                            &[T("19"), T("v19")],
                        ],
                        tag: "SELECT 17",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r.id, l.v FROM jr r LEFT JOIN jl l ON l.id = r.lid WHERE r.id BETWEEN 60 AND 75 ORDER BY r.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("60"), T("v60")],
                            &[T("61"), T("v61")],
                            &[T("62"), T("v62")],
                            &[T("63"), Null],
                            &[T("64"), T("v64")],
                            &[T("65"), Null],
                            &[T("66"), Null],
                            &[T("67"), Null],
                            &[T("68"), Null],
                            &[T("69"), Null],
                            &[T("70"), Null],
                            &[T("71"), Null],
                            &[T("72"), Null],
                            &[T("73"), Null],
                            &[T("74"), Null],
                            &[T("75"), Null],
                        ],
                        tag: "SELECT 16",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r.id, l.v FROM jr r JOIN jl l ON l.id = r.lid AND l.w > 5 WHERE r.id < 40 ORDER BY r.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("6"), T("v6")],
                            &[T("7"), T("v7")],
                            &[T("8"), T("v8")],
                            &[T("16"), T("v16")],
                            &[T("17"), T("v17")],
                            &[T("19"), T("v19")],
                            &[T("26"), T("v26")],
                            &[T("28"), T("v28")],
                            &[T("29"), T("v29")],
                            &[T("37"), T("v37")],
                            &[T("38"), T("v38")],
                            &[T("39"), T("v39")],
                        ],
                        tag: "SELECT 12",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM jr r JOIN (SELECT * FROM jl WHERE w = 3) l ON l.id = r.lid;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("25")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM jl l JOIN jr r ON r.lid = l.id WHERE l.id < 5 ORDER BY l.id, r.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("161")],
                            &[T("1"), T("241")],
                            &[T("2"), T("2")],
                            &[T("2"), T("82")],
                            &[T("2"), T("242")],
                            &[T("3"), T("3")],
                            &[T("3"), T("83")],
                            &[T("3"), T("163")],
                            &[T("4"), T("4")],
                            &[T("4"), T("84")],
                            &[T("4"), T("164")],
                            &[T("4"), T("244")],
                        ],
                        tag: "SELECT 13",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM jl l JOIN jr r ON r.lid = l.id;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("223")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT s.name, l.v FROM js s JOIN jl l ON l.id = s.k ORDER BY s.k;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("v", TEXT)],
                        rows: &[
                            &[T("n1"), T("v1")],
                            &[T("n2"), T("v2")],
                            &[T("n3"), T("v3")],
                            &[T("n4"), T("v4")],
                            &[T("n5"), T("v5")],
                            &[T("n6"), T("v6")],
                            &[T("n7"), T("v7")],
                            &[T("n8"), T("v8")],
                            &[T("n9"), T("v9")],
                            &[T("n10"), T("v10")],
                            &[T("n11"), T("v11")],
                            &[T("n12"), T("v12")],
                            &[T("n13"), T("v13")],
                            &[T("n14"), T("v14")],
                            &[T("n15"), T("v15")],
                            &[T("n16"), T("v16")],
                            &[T("n17"), T("v17")],
                            &[T("n18"), T("v18")],
                            &[T("n19"), T("v19")],
                            &[T("n20"), T("v20")],
                        ],
                        tag: "SELECT 20",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT r.id, s.name FROM jr r JOIN js s ON s.k = r.lid WHERE r.id < 30 ORDER BY r.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("n1")],
                            &[T("2"), T("n2")],
                            &[T("3"), T("n3")],
                            &[T("4"), T("n4")],
                            &[T("5"), T("n5")],
                            &[T("6"), T("n6")],
                            &[T("7"), T("n7")],
                            &[T("8"), T("n8")],
                            &[T("10"), T("n10")],
                            &[T("11"), T("n11")],
                            &[T("12"), T("n12")],
                            &[T("13"), T("n13")],
                            &[T("14"), T("n14")],
                            &[T("15"), T("n15")],
                            &[T("16"), T("n16")],
                            &[T("17"), T("n17")],
                            &[T("19"), T("n19")],
                            &[T("20"), T("n20")],
                        ],
                        tag: "SELECT 18",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM jr a JOIN jr b ON a.id = b.id;",
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
                    query: "SELECT count(*) FROM (SELECT a.id, a.u FROM jr a, jr b WHERE a.id = b.id LIMIT 7) s;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM jl a JOIN jl b ON b.id = a.w;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("58")],
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
fn test_alias_column_counts() {
    run_scripts(&[
        ScriptTest {
            name: "aliases that name more columns than a FROM item has",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE alias_t (a INT, b TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO alias_t VALUES (1, 'x');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM alias_t v(p, q);",
                    expected: Expected::Rows {
                        columns: &[Column("p", INT4), Column("q", TEXT)],
                        rows: &[
                            &[T("1"), T("x")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT p FROM alias_t v(p);",
                    expected: Expected::Rows {
                        columns: &[Column("p", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM alias_t v(p, q, r);",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: r#"table "v" has 2 columns available but 3 columns specified"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM generate_series(1, 2) g(n);",
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
                    query: "SELECT * FROM generate_series(1, 2) g(n, m);",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: r#"table "g" has 1 columns available but 2 columns specified"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (VALUES (1, 2)) v(a, b, c);",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: r#"table "v" has 2 columns available but 3 columns specified"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT 1) s(a, b);",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: r#"table "s" has 1 columns available but 2 columns specified"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_outer_join_reduction() {
    run_scripts(&[
        ScriptTest {
            name: "Outer joins that WHERE conditions reduce",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE oj_l (id INT PRIMARY KEY, k INT, t TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE oj_r (id INT PRIMARY KEY, k INT, v INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO oj_l VALUES (1, 1, 'a'), (2, 2, 'b'), (3, NULL, 'c'), (4, 4, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO oj_r VALUES (10, 1, 100), (11, 2, NULL), (12, 5, 50);",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l LEFT JOIN oj_r r ON l.k = r.k WHERE r.v > 10 ORDER BY l.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l LEFT JOIN oj_r r ON l.k = r.k WHERE r.v IS NULL ORDER BY l.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("2"), T("11")],
                            &[T("3"), Null],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l LEFT JOIN oj_r r ON l.k = r.k WHERE r.v + 1 > 10 OR r.id = 11 ORDER BY l.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("11")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l LEFT JOIN oj_r r ON l.k = r.k WHERE r.v > 10 OR l.id = 3 ORDER BY l.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l LEFT JOIN oj_r r ON l.k = r.k WHERE coalesce(r.v, 0) = 0 ORDER BY l.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("2"), T("11")],
                            &[T("3"), Null],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l RIGHT JOIN oj_r r ON l.k = r.k WHERE l.t <> 'z' ORDER BY r.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("11")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l FULL JOIN oj_r r ON l.k = r.k WHERE l.t IS NOT NULL ORDER BY l.id, r.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("11")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l FULL JOIN oj_r r ON l.k = r.k WHERE r.v::text LIKE '1%' ORDER BY l.id, r.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l FULL JOIN oj_r r ON l.k = r.k WHERE l.id < 4 AND r.v >= 0 ORDER BY l.id, r.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l LEFT JOIN oj_r r ON l.k = r.k WHERE (r.v > 10) IS NOT TRUE ORDER BY l.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("2"), T("11")],
                            &[T("3"), Null],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l.id, r.id FROM oj_l l LEFT JOIN oj_r r ON l.k = r.k WHERE NOT (r.v > 10) ORDER BY l.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
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
fn test_join_order_search() {
    run_scripts(&[
        ScriptTest {
            name: "Joins of several tables in any order",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE jo_a (id INT PRIMARY KEY, b_id INT, t TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE jo_b (id INT PRIMARY KEY, c_id INT, v INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE jo_c (id INT PRIMARY KEY, d_id INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE jo_d (id INT PRIMARY KEY, name TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO jo_a SELECT g, g % 30, 't' || g FROM generate_series(1, 300) g;",
                    expected: Expected::Tag("INSERT 0 300"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO jo_b SELECT g, g % 10, g * 2 FROM generate_series(0, 29) g;",
                    expected: Expected::Tag("INSERT 0 30"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO jo_c SELECT g, g % 3 FROM generate_series(0, 9) g;",
                    expected: Expected::Tag("INSERT 0 10"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO jo_d VALUES (0, 'zero'), (1, 'one'), (2, 'two');",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT d.name, count(*) FROM jo_a a JOIN jo_b b ON a.b_id = b.id JOIN jo_c c ON b.c_id = c.id JOIN jo_d d ON c.d_id = d.id GROUP BY d.name ORDER BY d.name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("count", INT8)],
                        rows: &[
                            &[T("one"), T("90")],
                            &[T("two"), T("90")],
                            &[T("zero"), T("120")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.id, d.name FROM jo_d d JOIN jo_c c ON c.d_id = d.id AND d.name = 'two' JOIN jo_b b ON b.c_id = c.id JOIN jo_a a ON a.b_id = b.id AND a.id < 40 ORDER BY a.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("2"), T("two")],
                            &[T("5"), T("two")],
                            &[T("8"), T("two")],
                            &[T("12"), T("two")],
                            &[T("15"), T("two")],
                            &[T("18"), T("two")],
                            &[T("22"), T("two")],
                            &[T("25"), T("two")],
                            &[T("28"), T("two")],
                            &[T("32"), T("two")],
                            &[T("35"), T("two")],
                            &[T("38"), T("two")],
                        ],
                        tag: "SELECT 12",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM jo_a a, jo_b b, jo_c c WHERE a.b_id = b.id AND b.c_id = c.id AND c.d_id = 1 AND a.t LIKE 't1%';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("34")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM jo_a a JOIN jo_b b ON a.b_id = b.id JOIN jo_c c ON true JOIN jo_d d ON 1 = 1 WHERE c.id < 2 AND d.id = 0;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("600")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM jo_a a JOIN jo_b b ON a.b_id = b.id JOIN jo_c c ON false;",
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
                    query: "SELECT a.id, b.id, c.id FROM jo_a a JOIN jo_b b ON a.b_id = b.id JOIN jo_c c ON b.c_id = c.id AND a.id = c.id ORDER BY a.id LIMIT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("1")],
                            &[T("2"), T("2"), T("2")],
                            &[T("3"), T("3"), T("3")],
                            &[T("4"), T("4"), T("4")],
                            &[T("5"), T("5"), T("5")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT x.*, y.* FROM (SELECT 1 AS p) x JOIN jo_d y ON y.id = x.p JOIN jo_c z ON z.d_id = y.id ORDER BY z.id;",
                    expected: Expected::Rows {
                        columns: &[Column("p", INT4), Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("one")],
                            &[T("1"), T("1"), T("one")],
                            &[T("1"), T("1"), T("one")],
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
fn test_implied_equalities() {
    run_scripts(&[
        ScriptTest {
            name: "Equalities implied through joins",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE eq_a (id INT PRIMARY KEY, x INT, t TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE eq_b (id INT PRIMARY KEY, y INT, u TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX eq_a_x ON eq_a (x);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO eq_a SELECT g, g % 20, 'a' || g FROM generate_series(1, 200) g;",
                    expected: Expected::Tag("INSERT 0 200"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO eq_b SELECT g, g % 25, 'b' || (g % 3) FROM generate_series(1, 100) g;",
                    expected: Expected::Tag("INSERT 0 100"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a.id, b.id FROM eq_a a JOIN eq_b b ON a.x = b.y WHERE b.y = 5 ORDER BY a.id, b.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("5"), T("5")],
                            &[T("5"), T("30")],
                            &[T("5"), T("55")],
                            &[T("5"), T("80")],
                            &[T("25"), T("5")],
                            &[T("25"), T("30")],
                            &[T("25"), T("55")],
                            &[T("25"), T("80")],
                            &[T("45"), T("5")],
                            &[T("45"), T("30")],
                            &[T("45"), T("55")],
                            &[T("45"), T("80")],
                            &[T("65"), T("5")],
                            &[T("65"), T("30")],
                            &[T("65"), T("55")],
                            &[T("65"), T("80")],
                            &[T("85"), T("5")],
                            &[T("85"), T("30")],
                            &[T("85"), T("55")],
                            &[T("85"), T("80")],
                            &[T("105"), T("5")],
                            &[T("105"), T("30")],
                            &[T("105"), T("55")],
                            &[T("105"), T("80")],
                            &[T("125"), T("5")],
                            &[T("125"), T("30")],
                            &[T("125"), T("55")],
                            &[T("125"), T("80")],
                            &[T("145"), T("5")],
                            &[T("145"), T("30")],
                            &[T("145"), T("55")],
                            &[T("145"), T("80")],
                            &[T("165"), T("5")],
                            &[T("165"), T("30")],
                            &[T("165"), T("55")],
                            &[T("165"), T("80")],
                            &[T("185"), T("5")],
                            &[T("185"), T("30")],
                            &[T("185"), T("55")],
                            &[T("185"), T("80")],
                        ],
                        tag: "SELECT 40",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM eq_a a, eq_b b WHERE a.x = b.y AND a.x = 7 AND b.y = 7;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("40")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM eq_a a, eq_b b WHERE a.x = b.y AND a.x = 7 AND b.y = 8;",
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
                    query: "SELECT count(*) FROM eq_a a JOIN eq_b b ON a.x = b.y AND b.y = 3 JOIN eq_a c ON c.x = a.x;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("400")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM eq_a a JOIN eq_b b ON a.x = b.y WHERE b.y = NULL;",
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
                    query: "SELECT count(*) FROM eq_a a JOIN eq_b b ON a.t = b.u WHERE b.u = 'b1';",
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
                    query: "SELECT count(*) FROM eq_a a JOIN eq_b b ON a.x = b.y WHERE b.y::bigint = 5;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("40")],
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
