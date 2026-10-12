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
fn test_with_statements() {
    run_scripts(&[
        ScriptTest {
            name: "basic values statements",
            set_up_script: &[
                "create table t (i int primary key);",
                "insert into t values (1), (2), (3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "with cte as (select 1) select * from cte;",
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
                    query: "with cte as (select 1, 2, 3 union select 4, 5, 6) select * from cte;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4), Column("?column?", INT4), Column("?column?", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("4"), T("5"), T("6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "with cte as (values (1)) select * from cte;",
                    expected: Expected::Rows {
                        columns: &[Column("column1", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "with cte as (values (1, 2, 3) union values (4, 5, 6)) select * from cte;",
                    expected: Expected::Rows {
                        columns: &[Column("column1", INT4), Column("column2", INT4), Column("column3", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("4"), T("5"), T("6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "with cte as (select 1, 2, 3 union values (4, 5, 6)) select * from cte;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4), Column("?column?", INT4), Column("?column?", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("4"), T("5"), T("6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "with recursive cte(x) as (select 1 union all select x + 1 from cte) select * from cte limit 5;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4)],
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
            ],
            ..S
        },
    ]);
}

#[test]
fn test_recursive_query_form_rules() {
    run_scripts(&[
        ScriptTest {
            name: "malformed recursive queries",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"WITH x(n, b) AS (SELECT 1)
SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P10", message: r#"WITH query "x" has 1 columns available but 2 columns specified"#, position: 6, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT 1 INTERSECT SELECT n+1 FROM x)
	SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive query "x" does not have the form non-recursive-term UNION [ALL] recursive-term"#, position: 16, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT 1 INTERSECT ALL SELECT n+1 FROM x)
	SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive query "x" does not have the form non-recursive-term UNION [ALL] recursive-term"#, position: 16, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT 1 EXCEPT SELECT n+1 FROM x)
	SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive query "x" does not have the form non-recursive-term UNION [ALL] recursive-term"#, position: 16, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT 1 EXCEPT ALL SELECT n+1 FROM x)
	SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive query "x" does not have the form non-recursive-term UNION [ALL] recursive-term"#, position: 16, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT n FROM x)
	SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive query "x" does not have the form non-recursive-term UNION [ALL] recursive-term"#, position: 16, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT n FROM x UNION ALL SELECT 1)
	SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive reference to query "x" must not appear within its non-recursive term"#, position: 39, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE y (a INTEGER);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO y SELECT generate_series(1, 10);",
                    expected: Expected::Tag("INSERT 0 10"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT a FROM y WHERE a = 1
	UNION ALL
	SELECT x.n+1 FROM y LEFT JOIN x ON x.n = y.a WHERE n < 10)
SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive reference to query "x" must not appear within an outer join"#, position: 95, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT a FROM y WHERE a = 1
	UNION ALL
	SELECT x.n+1 FROM x RIGHT JOIN y ON x.n = y.a WHERE n < 10)
SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive reference to query "x" must not appear within an outer join"#, position: 83, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT a FROM y WHERE a = 1
	UNION ALL
	SELECT x.n+1 FROM x FULL JOIN y ON x.n = y.a WHERE n < 10)
SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive reference to query "x" must not appear within an outer join"#, position: 83, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM x
                          WHERE n IN (SELECT * FROM x))
  SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive reference to query "x" must not appear within a subquery"#, position: 114, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT 1 UNION ALL SELECT count(*) FROM x)
  SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: "aggregate functions are not allowed in a recursive query's recursive term", position: 51, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT 1 UNION ALL SELECT sum(n) FROM x)
  SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: "aggregate functions are not allowed in a recursive query's recursive term", position: 51, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM x ORDER BY 1)
  SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "ORDER BY in a recursive query is not implemented", position: 71, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM x LIMIT 10 OFFSET 1)
  SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "OFFSET in a recursive query is not implemented", position: 78, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM x FOR UPDATE)
  SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "FOR UPDATE/SHARE in a recursive query is not implemented", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE x(id) AS (values (1)
    UNION ALL
    SELECT (SELECT * FROM x) FROM x WHERE id < 5
) SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive reference to query "x" must not appear within a subquery"#, position: 77, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE
  x (id) AS (SELECT 1 UNION ALL SELECT id+1 FROM y WHERE id < 5),
  y (id) AS (SELECT 1 UNION ALL SELECT id+1 FROM x WHERE id < 5)
SELECT * FROM x;"#,
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "mutual recursion between WITH items is not implemented", position: 18, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE foo(i) AS
    (values (1)
    UNION ALL
       (SELECT i+1 FROM foo WHERE i < 10
          UNION ALL
       SELECT i+1 FROM foo WHERE i < 5)
) SELECT * FROM foo;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive reference to query "foo" must not appear more than once"#, position: 140, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE foo(i) AS
    (values (1)
    UNION ALL
	   SELECT * FROM
       (SELECT i+1 FROM foo WHERE i < 10
          UNION ALL
       SELECT i+1 FROM foo WHERE i < 5) AS t
) SELECT * FROM foo;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive reference to query "foo" must not appear more than once"#, position: 158, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE foo(i) AS
    (values (1)
    UNION ALL
       (SELECT i+1 FROM foo WHERE i < 10
          EXCEPT
       SELECT i+1 FROM foo WHERE i < 5)
) SELECT * FROM foo;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive reference to query "foo" must not appear within EXCEPT"#, position: 137, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE foo(i) AS
    (values (1)
    UNION ALL
       (SELECT i+1 FROM foo WHERE i < 10
          INTERSECT
       SELECT i+1 FROM foo WHERE i < 5)
) SELECT * FROM foo;"#,
                    expected: Expected::Error(Diagnostic { code: "42P19", message: r#"recursive reference to query "foo" must not appear more than once"#, position: 140, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE foo(i) AS
   (SELECT i FROM (VALUES(1),(2)) t(i)
   UNION ALL
   SELECT (i+1)::numeric(10,0) FROM foo WHERE i < 10)
SELECT * FROM foo;"#,
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"recursive query "foo" column 1 has type integer in non-recursive term but type numeric overall"#, hint: "Cast the output of the non-recursive term to the correct type.", position: 37, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE foo(i) AS
   (SELECT i::numeric(3,0) FROM (VALUES(1),(2)) t(i)
   UNION ALL
   SELECT (i+1)::numeric(10,0) FROM foo WHERE i < 10)
SELECT * FROM foo;"#,
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"recursive query "foo" column 1 has type numeric(3,0) in non-recursive term but type numeric overall"#, hint: "Cast the output of the non-recursive term to the correct type.", position: 37, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_recursive_union_of_unhashable_types() {
    run_scripts(&[
        ScriptTest {
            name: "a recursive UNION over a type without a hash function",
            assertions: &[
                ScriptTestAssertion {
                    query: "WITH RECURSIVE t(n) AS (VALUES ('01'::varbit) UNION SELECT n || '10'::varbit FROM t WHERE n < '100'::varbit) SELECT n FROM t;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "could not implement recursive UNION", detail: "All column datatypes must be hashable.", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
