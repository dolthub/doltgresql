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
fn test_all_quantifier() {
    run_scripts(&[
        ScriptTest {
            name: "ALL quantifier over arrays",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1 = ALL(ARRAY[1, 1, 1]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 = ALL(ARRAY[1, 2, 1]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 = ALL(ARRAY[]::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 = ALL(ARRAY[1, NULL]::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 = ALL(ARRAY[2, NULL]::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 5 != ALL(ARRAY[1, 2, 3]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 != ALL(ARRAY[1, 2, 3]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALL quantifier over subqueries",
            set_up_script: &[
                "create table t (i int primary key);",
                "insert into t values (1), (2), (3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 0 < ALL(SELECT i FROM t);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2 < ALL(SELECT i FROM t);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 = ALL(SELECT i FROM t WHERE i = 4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
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
            name: "ALL quantifier with a NULL left-hand operand",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT NULL = ALL(ARRAY[1, 2]::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL = ALL(ARRAY[]::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL != ALL(ARRAY[]::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
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
            name: "ALL quantifier over a subquery containing NULL rows",
            set_up_script: &[
                "create table t3 (id int primary key, v int);",
                "insert into t3 values (1, 1), (2, NULL), (3, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1 = ALL(SELECT v FROM t3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 = ALL(SELECT v FROM t3 WHERE id IN (1, 2));",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3 != ALL(SELECT v FROM t3 WHERE id = 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
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
            name: "ALL quantifier filtering table rows",
            set_up_script: &[
                "create table t2 (id int primary key, vals int[]);",
                "insert into t2 values (1, ARRAY[1,2,3]), (2, ARRAY[4,5,6]), (3, ARRAY[1,7,8]), (4, ARRAY[9,9,9]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM t2 WHERE 9 = ALL(vals) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t2 WHERE 1 != ALL(vals) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t2 WHERE NOT (1 != ALL(vals)) ORDER BY id;",
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
            name: "ALL quantifier with an untyped bind variable array",
            set_up_script: &[
                "create table t4 (id int primary key);",
                "insert into t4 values (1), (2), (3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM t4 WHERE id != ALL($1) ORDER BY id;",
                    bind_vars: &[BindVar::Int32Array(&[1, 2])],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "!= works with all three quantifiers",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1 != ANY(ARRAY[1, 2, 3]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 != SOME(ARRAY[1, 1, 1]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 != ALL(ARRAY[1, 1, 1]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALL quantifier in a recursive CTE cycle check",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE r(id, path) AS (
					SELECT 'a'::text, ARRAY['a'::text]
					UNION ALL
					SELECT r.id || 'x', r.path || (r.id || 'x')
					FROM r
					WHERE length(r.id) < 3 AND (r.id || 'x') != ALL(r.path)
				)
				SELECT count(*) FROM r;"#,
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "any expression with array in string format",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'i' = ANY('{information_schema, something}');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'somedb' = ANY('{information_schema, somedb}');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'somedb' = SOME('{information_schema, somedb}');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'somedb' = ALL('{information_schema, somedb}');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT nsp.nspname AS schema_name,
       (nsp.nspname = 'pg_catalog'
        AND EXISTS (SELECT 1
                    FROM   pg_catalog.pg_class
                    WHERE  relname = 'pg_class'
                           AND relnamespace = nsp.oid LIMIT 1))
       OR (nsp.nspname = 'pgagent'
           AND EXISTS (SELECT 1
                       FROM   pg_catalog.pg_class
                       WHERE  relname = 'pga_job'
                              AND relnamespace = nsp.oid LIMIT 1))
       OR (nsp.nspname = 'information_schema'
           AND EXISTS (SELECT 1
                       FROM   pg_catalog.pg_class
                       WHERE  relname = 'tables'
                              AND relnamespace = nsp.oid LIMIT 1)) AS is_catalog,
       CASE
         WHEN nsp.nspname = ANY('{information_schema}')
         THEN FALSE
         ELSE TRUE
       END AS db_support
FROM   pg_catalog.pg_namespace nsp
WHERE  nsp.oid = 2200::OID;"#,
                    expected: Expected::Rows {
                        columns: &[Column("schema_name", NAME), Column("is_catalog", BOOL), Column("db_support", BOOL)],
                        rows: &[
                            &[T("public"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select oid from pg_class join pg_index on pg_class.oid = ANY(indclass);",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID)],
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
fn test_any_all_serialization() {
    run_scripts(&[
        ScriptTest {
            name: "ANY/ALL check constraint round trip",
            set_up_script: &[
                "CREATE TABLE t_all (id int, v text, CONSTRAINT ck_all CHECK (v <> ALL (ARRAY['no','nope'])));",
                "CREATE TABLE t_any (id int, v text, CONSTRAINT ck_any CHECK (v <> ANY (ARRAY['no','nope'])));",
                "CREATE TABLE t_some (id int, v int, CONSTRAINT ck_some CHECK (v > SOME (ARRAY[1,2])));",
                "CREATE TABLE t_eq (id int, v text, CONSTRAINT ck_eq CHECK (v = ANY (ARRAY['yes','yep'])));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_constraintdef(oid) FROM pg_catalog.pg_constraint WHERE conname = 'ck_all';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_constraintdef", TEXT)],
                        rows: &[
                            &[T("CHECK ((v <> ALL (ARRAY['no'::text, 'nope'::text])))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_constraintdef(oid) FROM pg_catalog.pg_constraint WHERE conname = 'ck_any';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_constraintdef", TEXT)],
                        rows: &[
                            &[T("CHECK ((v <> ANY (ARRAY['no'::text, 'nope'::text])))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_constraintdef(oid) FROM pg_catalog.pg_constraint WHERE conname = 'ck_some';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_constraintdef", TEXT)],
                        rows: &[
                            &[T("CHECK ((v > ANY (ARRAY[1, 2])))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_constraintdef(oid) FROM pg_catalog.pg_constraint WHERE conname = 'ck_eq';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_constraintdef", TEXT)],
                        rows: &[
                            &[T("CHECK ((v = ANY (ARRAY['yes'::text, 'yep'::text])))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_all VALUES (1, 'ok');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_all VALUES (2, 'no');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t_all" violates check constraint "ck_all""#, detail: "Failing row contains (2, no).", schema: "public", table: "t_all", constraint: "ck_all", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_any VALUES (1, 'no');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_some VALUES (1, 2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_some VALUES (2, 1);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t_some" violates check constraint "ck_some""#, detail: "Failing row contains (2, 1).", schema: "public", table: "t_some", constraint: "ck_some", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_eq VALUES (1, 'yes');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_eq VALUES (2, 'nope');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t_eq" violates check constraint "ck_eq""#, detail: "Failing row contains (2, nope).", schema: "public", table: "t_eq", constraint: "ck_eq", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
