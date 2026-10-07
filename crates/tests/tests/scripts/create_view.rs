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
fn test_create_view_statements() {
    run_scripts(&[
        ScriptTest {
            name: "basic create view statements",
            set_up_script: &[
                "create table t1 (pk int);",
                "insert into t1 values (1), (2), (3), (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "create view v as select * from t1 order by pk;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from v order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "views on different schemas",
            set_up_script: &[
                "CREATE SCHEMA testschema;",
                "SET search_path TO testschema;",
                "CREATE TABLE testing (pk INT primary key, v2 TEXT);",
                "INSERT INTO testing VALUES (1,'a'), (2,'b'), (3,'c');",
                "CREATE VIEW testview AS SELECT * FROM testing;",
                "CREATE SCHEMA myschema;",
                "SET search_path TO myschema;",
                "CREATE TABLE mytable (pk INT primary key, v1 INT);",
                "INSERT INTO mytable VALUES (1,4), (2,5), (3,6);",
                "CREATE VIEW myview AS SELECT * FROM mytable;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW search_path;",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("myschema")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v1 from myview order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v2 from testview order by pk;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "testview" does not exist"#, position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v2 from testschema.testview order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("v2", TEXT)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                            &[T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt_schemas;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("myview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'testschema';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so this expectation follows Postgres' behavior.
                ScriptTestAssertion {
                    query: "SHOW search_path;",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("testschema")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "select * from myview order by pk; /* err */",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "relation \"myview\" does not exist", position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v1 from myschema.myview order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v2 from testview order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("v2", TEXT)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                            &[T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt_schemas;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("testview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = testschema, myschema;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so this expectation follows Postgres' behavior.
                ScriptTestAssertion {
                    query: "SHOW search_path;",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("testschema, myschema")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so this expectation follows Postgres' behavior.
                ScriptTestAssertion {
                    query: "select v1 from myview order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v2 from testview order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("v2", TEXT)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                            &[T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt_schemas;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("testview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create view from view",
            set_up_script: &[
                "create table t1 (pk int);",
                "insert into t1 values (1), (2), (3), (1);",
                "create view v as select * from t1 where pk > 1;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "create view v1 as select * from v order by pk;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from v1 order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("2")],
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
            name: "view with expression name",
            set_up_script: &[
                "create view v as select 2+2",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * from v;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
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
            name: "view with column names",
            set_up_script: &[
                "CREATE TABLE xy (x int primary key, y int);",
                "insert into xy values (1, 4), (4, 9)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "create view v_today(today) as select 2",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW xyv (u,v) AS SELECT * from xy",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v from xyv;",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("9")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT today from v_today;",
                    expected: Expected::Rows {
                        columns: &[Column("today", INT4)],
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
        ScriptTest {
            name: "nested view",
            set_up_script: &[
                "create table t1 (pk int);",
                "insert into t1 values (1), (2), (3), (4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "create view unionView as (select * from t1 order by pk desc limit 1) union all (select * from t1 order by pk limit 1)",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from unionView order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "cast (postgres-specific syntax)",
            set_up_script: &[
                "create table t1 (pk int);",
                "insert into t1 values (1), (2), (3), (4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE VIEW v AS SELECT pk::INT2 FROM t1 ORDER BY pk;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from v order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT2)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v_text AS SELECT pk::int2, (pk)::text AS pk_text FROM t1;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select pk_text from v_text order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk_text", TEXT)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "not yet supported create view queries",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TEMPORARY VIEW v AS SELECT 1;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE RECURSIVE VIEW v AS SELECT 1;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "AS""#, position: 25, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v AS SELECT 1 WITH LOCAL CHECK OPTION;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "WITH CHECK OPTION is supported only on automatically updatable views", hint: "Views that do not select from a single table or view are not automatically updatable.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v WITH (check_option = 'local') AS SELECT 1;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "WITH CHECK OPTION is supported only on automatically updatable views", hint: "Views that do not select from a single table or view are not automatically updatable.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v WITH (security_barrier = true) AS SELECT 1;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create view with CTE",
            set_up_script: &[
                "CREATE TABLE public.t1 (id integer NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "create view public.v1 as with table1 as (select * from t1) select id from table1;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create view with custom type in its select statement",
            set_up_script: &[
                "CREATE TYPE e AS ENUM ('sched', 'busy', 'final', 'help');",
                "CREATE TABLE t (id integer NOT NULL, t e);",
                "INSERT INTO t VALUES (1, 'busy'), (2, 'final'), (3, 'busy'), (4, 'help');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "create view v as select * from t where (t = 'busy'::e);",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from v;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("busy")],
                            &[T("3"), T("busy")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "View names must be unique across all relation types",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int PRIMARY KEY, v1 int);",
                "CREATE SEQUENCE seq1;",
                "CREATE VIEW existing_view AS SELECT pk FROM tbl1;",
                "CREATE INDEX idx1 ON tbl1 (v1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE VIEW tbl1 AS SELECT pk FROM tbl1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "tbl1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OR REPLACE VIEW tbl1 AS SELECT pk FROM tbl1;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#""tbl1" is not a view"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW seq1 AS SELECT pk FROM tbl1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "seq1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OR REPLACE VIEW seq1 AS SELECT pk FROM tbl1;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#""seq1" is not a view"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW existing_view AS SELECT pk FROM tbl1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "existing_view" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OR REPLACE VIEW existing_view AS SELECT pk FROM tbl1;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW idx1 AS SELECT pk FROM tbl1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "idx1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OR REPLACE VIEW idx1 AS SELECT pk FROM tbl1;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#""idx1" is not a view"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_view_definitions() {
    run_scripts(&[
        ScriptTest {
            name: "pg_get_viewdef and pg_views definitions",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE a (id int PRIMARY KEY, name text, d date, n numeric);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b (id int, a_id int, x varchar(10));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v1 AS SELECT name FROM a;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v2 AS SELECT * FROM a WHERE id > 1 AND name LIKE 'x%' ORDER BY name DESC, id LIMIT 10 OFFSET 2;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v3 AS SELECT a.id, b.x AS bx, a.id + 1 AS next, upper(a.name), count(*) OVER () FROM a JOIN b ON a.id = b.a_id LEFT JOIN b b2 ON b2.id = b.id;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v4 AS SELECT a_id, count(*), sum(id) AS total FROM b GROUP BY a_id HAVING count(*) > 1;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v5 AS SELECT id FROM a UNION SELECT id FROM b UNION ALL SELECT 3;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v6 AS SELECT id FROM a WHERE EXISTS (SELECT 1 FROM b WHERE b.a_id = a.id) AND id IN (SELECT a_id FROM b) AND name = (SELECT max(x) FROM b);",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v7 (c1, c2) AS SELECT DISTINCT id, CASE WHEN id > 1 THEN 'big' ELSE 'small' END FROM a;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v8 AS WITH t AS (SELECT id FROM a) SELECT t.id FROM t, b WHERE t.id = b.id;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v9 AS SELECT s.x FROM (SELECT id AS x FROM a) s, generate_series(1, 3) g(n) WHERE s.x = g.n;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v10 AS SELECT 1 AS one, 'a'::text, now(), current_date, coalesce(name, 'z'), id::text, n::int FROM a;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v11 AS SELECT id, d FROM a WHERE d > '2020-01-01' AND id BETWEEN 1 AND 5 AND name IS NOT NULL AND id IN (1, 2, 3);",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v12 AS SELECT * FROM (VALUES (1, 'one'), (2, 'two')) AS v(num, word);",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v13 AS SELECT extract(year FROM d), substring(name FROM 2 FOR 3), trim(name), position('a' IN name) FROM a;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v14 AS SELECT b.id FROM a CROSS JOIN b;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v15 AS SELECT id FROM a INTERSECT SELECT id FROM b EXCEPT SELECT 5;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v16 AS SELECT a.id FROM a JOIN b USING (id);",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v17 AS SELECT DISTINCT ON (a_id) a_id, x FROM b ORDER BY a_id, x DESC NULLS LAST;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT viewname, definition FROM pg_views WHERE schemaname = 'public' ORDER BY viewname;",
                    expected: Expected::Rows {
                        columns: &[Column("viewname", NAME), Column("definition", TEXT)],
                        rows: &[
                            &[T("v1"), T(r#" SELECT a.name
   FROM a;"#)],
                            &[T("v10"), T(r#" SELECT 1 AS one,
    'a'::text AS text,
    now() AS now,
    CURRENT_DATE AS "current_date",
    COALESCE(a.name, 'z'::text) AS "coalesce",
    (a.id)::text AS id,
    (a.n)::integer AS n
   FROM a;"#)],
                            &[T("v11"), T(r#" SELECT a.id,
    a.d
   FROM a
  WHERE ((a.d > '2020-01-01'::date) AND ((a.id >= 1) AND (a.id <= 5)) AND (a.name IS NOT NULL) AND (a.id = ANY (ARRAY[1, 2, 3])));"#)],
                            &[T("v12"), T(r#" SELECT v.num,
    v.word
   FROM ( VALUES (1,'one'::text), (2,'two'::text)) v(num, word);"#)],
                            &[T("v13"), T(r#" SELECT EXTRACT(year FROM a.d) AS "extract",
    SUBSTRING(a.name FROM 2 FOR 3) AS "substring",
    TRIM(BOTH FROM a.name) AS btrim,
    POSITION(('a'::text) IN (a.name)) AS "position"
   FROM a;"#)],
                            &[T("v14"), T(r#" SELECT b.id
   FROM (a
     CROSS JOIN b);"#)],
                            &[T("v15"), T(r#"(
         SELECT a.id
           FROM a
        INTERSECT
         SELECT b.id
           FROM b
) EXCEPT
 SELECT 5 AS id;"#)],
                            &[T("v16"), T(r#" SELECT a.id
   FROM (a
     JOIN b USING (id));"#)],
                            &[T("v17"), T(r#" SELECT DISTINCT ON (b.a_id) b.a_id,
    b.x
   FROM b
  ORDER BY b.a_id, b.x DESC NULLS LAST;"#)],
                            &[T("v2"), T(r#" SELECT a.id,
    a.name,
    a.d,
    a.n
   FROM a
  WHERE ((a.id > 1) AND (a.name ~~ 'x%'::text))
  ORDER BY a.name DESC, a.id
 OFFSET 2
 LIMIT 10;"#)],
                            &[T("v3"), T(r#" SELECT a.id,
    b.x AS bx,
    (a.id + 1) AS next,
    upper(a.name) AS upper,
    count(*) OVER () AS count
   FROM ((a
     JOIN b ON ((a.id = b.a_id)))
     LEFT JOIN b b2 ON ((b2.id = b.id)));"#)],
                            &[T("v4"), T(r#" SELECT b.a_id,
    count(*) AS count,
    sum(b.id) AS total
   FROM b
  GROUP BY b.a_id
 HAVING (count(*) > 1);"#)],
                            &[T("v5"), T(r#"(
         SELECT a.id
           FROM a
        UNION
         SELECT b.id
           FROM b
) UNION ALL
 SELECT 3 AS id;"#)],
                            &[T("v6"), T(r#" SELECT a.id
   FROM a
  WHERE ((EXISTS ( SELECT 1
           FROM b
          WHERE (b.a_id = a.id))) AND (a.id IN ( SELECT b.a_id
           FROM b)) AND (a.name = ( SELECT max((b.x)::text) AS max
           FROM b)));"#)],
                            &[T("v7"), T(r#" SELECT DISTINCT a.id AS c1,
        CASE
            WHEN (a.id > 1) THEN 'big'::text
            ELSE 'small'::text
        END AS c2
   FROM a;"#)],
                            &[T("v8"), T(r#" WITH t AS (
         SELECT a.id
           FROM a
        )
 SELECT t.id
   FROM t,
    b
  WHERE (t.id = b.id);"#)],
                            &[T("v9"), T(r#" SELECT s.x
   FROM ( SELECT a.id AS x
           FROM a) s,
    generate_series(1, 3) g(n)
  WHERE (s.x = g.n);"#)],
                        ],
                        tag: "SELECT 17",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_viewdef('v3'::regclass, true), pg_get_viewdef('v6', true), pg_get_viewdef('v1'::regclass::oid, 10);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_viewdef", TEXT), Column("pg_get_viewdef", TEXT), Column("pg_get_viewdef", TEXT)],
                        rows: &[
                            &[T(r#" SELECT a.id,
    b.x AS bx,
    a.id + 1 AS next,
    upper(a.name) AS upper,
    count(*) OVER () AS count
   FROM a
     JOIN b ON a.id = b.a_id
     LEFT JOIN b b2 ON b2.id = b.id;"#), T(r#" SELECT a.id
   FROM a
  WHERE (EXISTS ( SELECT 1
           FROM b
          WHERE b.a_id = a.id)) AND (a.id IN ( SELECT b.a_id
           FROM b)) AND a.name = (( SELECT max(b.x::text) AS max
           FROM b));"#), T(r#" SELECT a.name
   FROM a;"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_viewdef with windows, aggregates, functions, and pretty-printing",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT string_agg(relname, ',') FROM (SELECT relname FROM pg_class WHERE relkind='v' AND relnamespace='information_schema'::regnamespace LIMIT 6) s; SELECT count(*) FROM information_schema.views; SELECT count(*) FROM pg_class WHERE relkind='v';",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "cannot insert multiple commands into a prepared statement", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_viewdef('pg_catalog.pg_roles'::regclass) = definition FROM pg_views WHERE viewname = 'pg_roles';",
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
    ]);
}

#[test]
fn test_view_and_routine_rules() {
    run_scripts(&[
        ScriptTest {
            name: "WITH queries in INSERT, UPDATE, and DELETE",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE w1 (a INT PRIMARY KEY, b TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH src AS (SELECT 1 AS a, 'x' AS b UNION ALL SELECT 2, 'y') INSERT INTO w1 SELECT * FROM src;",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH v AS (SELECT 5 AS n) INSERT INTO w1 VALUES ((SELECT n FROM v), 'z') RETURNING *;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("5"), T("z")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH t AS (SELECT 2 AS a) UPDATE w1 SET b = 'updated' WHERE a IN (SELECT a FROM t) RETURNING *;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("2"), T("updated")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH t AS (SELECT 1 AS a) DELETE FROM w1 USING t WHERE w1.a = t.a RETURNING w1.*;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("x")],
                        ],
                        tag: "DELETE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH t AS (SELECT 1 AS a), t AS (SELECT 2) INSERT INTO w1 VALUES (9, 'q');",
                    expected: Expected::Error(Diagnostic { code: "42712", message: r#"WITH query name "t" specified more than once"#, position: 28, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM w1 ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("2"), T("updated")],
                            &[T("5"), T("z")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Views over other relations and check options",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE vt (pk INT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE vs;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OR REPLACE VIEW vt AS SELECT 1;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#""vt" is not a view"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE OR REPLACE VIEW vs AS SELECT 1;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#""vs" is not a view"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW vc AS SELECT 1 WITH LOCAL CHECK OPTION;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "WITH CHECK OPTION is supported only on automatically updatable views", hint: "Views that do not select from a single table or view are not automatically updatable.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW vc WITH (check_option = 'cascaded') AS SELECT 1;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "WITH CHECK OPTION is supported only on automatically updatable views", hint: "Views that do not select from a single table or view are not automatically updatable.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW vc WITH (security_barrier = true) AS SELECT 1;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW vd AS SELECT pk FROM vt WITH CHECK OPTION;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM vc;",
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
            name: "Composite routine parameters and results",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SCHEMA rsch;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE rsch.pair AS (id INT, label TEXT);",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE goods (id INT PRIMARY KEY, name TEXT NOT NULL, qty INT NOT NULL, price REAL NOT NULL);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO goods VALUES (1, 'apple', 3, 2.5), (2, 'banana', 5, 1.2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION total(g goods) RETURNS REAL AS $$ BEGIN RETURN g.qty * g.price; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT total(g) FROM goods AS g ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("total", FLOAT4)],
                        rows: &[
                            &[T("6")],
                            &[T("7.5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION pairs() RETURNS TABLE(p rsch.pair) LANGUAGE plpgsql AS $$ BEGIN RETURN QUERY SELECT 1, 'one'; RETURN QUERY SELECT 2, 'two'; END; $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pairs();",
                    expected: Expected::Rows {
                        columns: &[Column("pairs", USER_DEFINED)],
                        rows: &[
                            &[T("(1,one)")],
                            &[T("(2,two)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pairs();",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("label", TEXT)],
                        rows: &[
                            &[T("1"), T("one")],
                            &[T("2"), T("two")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION pairs_sql() RETURNS TABLE(p rsch.pair) LANGUAGE sql AS $$ SELECT 3, 'three' $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pairs_sql();",
                    expected: Expected::Rows {
                        columns: &[Column("pairs_sql", USER_DEFINED)],
                        rows: &[
                            &[T("(3,three)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION nested_do() RETURNS void AS $f$ BEGIN DO $b$ BEGIN INSERT INTO goods VALUES (3, 'cherry', 1, 9); END $b$; END; $f$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nested_do();",
                    expected: Expected::Rows {
                        columns: &[Column("nested_do", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM goods WHERE id = 3;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("cherry")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Arrays of arrays",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[ARRAY[]::int[]];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[ARRAY[1,2]::int[], ARRAY[3,4]::int[]];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2},{3,4}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(ARRAY[ARRAY[1]::int[]]);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("integer[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Recreated serial sequences start over",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE rs (pk SERIAL PRIMARY KEY, v TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rs (v) VALUES ('a'), ('b');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE rs;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE rs (pk SERIAL PRIMARY KEY, v TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rs (v) VALUES ('c') RETURNING pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
