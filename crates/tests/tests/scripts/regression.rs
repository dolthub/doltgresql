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
fn test_regressions() {
    run_scripts(&[
        ScriptTest {
            name: "nullif",
            assertions: &[
                ScriptTestAssertion {
                    query: "select nullif(1, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("nullif", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select nullif('', null);",
                    expected: Expected::Rows {
                        columns: &[Column("nullif", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select nullif(10, 'a');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 19, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "coalesce",
            assertions: &[
                ScriptTestAssertion {
                    query: "select coalesce(null + 5, 100);",
                    expected: Expected::Rows {
                        columns: &[Column("coalesce", INT4)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select coalesce(null, null, 'abc');",
                    expected: Expected::Rows {
                        columns: &[Column("coalesce", TEXT)],
                        rows: &[
                            &[T("abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select coalesce(null, null);",
                    expected: Expected::Rows {
                        columns: &[Column("coalesce", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "case / when",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT
  CASE
    WHEN 1 = 1 THEN 'One is equal to One'
    ELSE 'One is not equal to One'
  END AS result;"#,
                    expected: Expected::Rows {
                        columns: &[Column("result", TEXT)],
                        rows: &[
                            &[T("One is equal to One")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT
  CASE
    WHEN NULL IS NULL THEN 'NULL is NULL'
    ELSE 'NULL is not NULL'
  END AS result;"#,
                    expected: Expected::Rows {
                        columns: &[Column("result", TEXT)],
                        rows: &[
                            &[T("NULL is NULL")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALL / DISTINCT in functions",
            set_up_script: &[
                "create table t1 (pk int);",
                "insert into t1 values (1), (2), (3), (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select all count(*) from t1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select all count(distinct pk) from t1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select all count(all pk) from t1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
            name: "cross joins",
            set_up_script: &[
                "create table t1 (pk1 int);",
                "create table t2 (pk2 int);",
                "insert into t1 values (1), (2);",
                "insert into t2 values (3), (4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from t1 cross join t2 order by pk1, pk2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk1", INT4), Column("pk2", INT4)],
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
                    query: "select * from t1, t2 order by pk1, pk2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk1", INT4), Column("pk2", INT4)],
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
            ],
            ..S
        },
        ScriptTest {
            name: "casting null as integer",
            set_up_script: &[
                "CREATE TABLE tab0(pk INTEGER PRIMARY KEY, col0 INTEGER, col1 FLOAT, col2 TEXT, col3 INTEGER, col4 FLOAT, col5 TEXT);",
                "INSERT INTO tab0 VALUES (0,698,169.42,'apdbu',431,316.15,'sqvis'), (1,538,676.36,'fuqeu',514,685.97,'bgwrq');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ALL + 58 FROM tab0 WHERE NULL NOT BETWEEN + 71 * CAST ( NULL AS INTEGER ) AND col4",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "addition expression in prepared statement",
            set_up_script: &[
                "CREATE TABLE t1(x INTEGER);",
                "CREATE TABLE t2(y INTEGER PRIMARY KEY);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1 IN (SELECT x+y FROM t1, t2);",
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
            name: "casting from float64 to int64 and float32",
            set_up_script: &[
                "CREATE TABLE tab0(pk INTEGER PRIMARY KEY, col0 INTEGER, col1 FLOAT, col2 TEXT, col3 INTEGER, col4 FLOAT, col5 TEXT);",
                r#"INSERT INTO tab0 VALUES (0,698,169.42,'apdbu',431,316.15,'sqvis'), (1,538,676.36,'fuqeu',514,685.97,'bgwrq'), (2,90,205.26,'yrrzx',123,836.88,'kpuhc'), 
(3,620,864.8,'myrdv',877,820.98,'oxkuv'), (4,754,677.3,'iofrg',67,665.49,'bzqba'), (5,107,710.19,'lhfro',286,504.28,'kwwsg'), (6,904,193.16,'eozui',48,698.55,'ejyzs'), 
(7,606,650.64,'ovmce',417,962.43,'dvkbh'), (8,535,18.11,'ijika',630,489.63,'hpnyu'), (9,501,776.40,'cvygg',725,75.5,'etlyv');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM tab0 WHERE - - col0 * + - col4 >= ( + CAST ( col1 AS REAL ) );",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("col0", INT4), Column("col1", FLOAT8), Column("col2", TEXT), Column("col3", INT4), Column("col4", FLOAT8), Column("col5", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "typecheck fails to detect doltgres types in GMS ",
            set_up_script: &[
                "CREATE TABLE tab0(col0 INTEGER, col1 INTEGER, col2 INTEGER);",
                "INSERT INTO tab0 VALUES (97,1,99), (15,81,47), (87,21,10);",
                "CREATE TABLE tab1(col0 INTEGER, col1 INTEGER, col2 INTEGER);",
                "INSERT INTO tab1 VALUES (51,14,96), (85,5,59), (91,47,68);",
                "CREATE TABLE tab2(col0 INTEGER, col1 INTEGER, col2 INTEGER);",
                "INSERT INTO tab2 VALUES(64,77,40), (75,67,58), (46,51,23);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ALL + col2 * + ( ( 43 ) ) - 47 + + col1 * CAST ( - ( + 63 ) / col0 AS INTEGER ) AS col0 FROM tab1 AS cor0;",
                    expected: Expected::Rows {
                        columns: &[Column("col0", INT4)],
                        rows: &[
                            &[T("4067")],
                            &[T("2490")],
                            &[T("2877")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT - COUNT ( * ) - 26 * + 96 AS col2 FROM tab0 WHERE + 2 * col0 NOT BETWEEN 33 * CAST ( + 7 / 91 AS REAL ) AND 52;",
                    expected: Expected::Rows {
                        columns: &[Column("col2", INT8)],
                        rows: &[
                            &[T("-2498")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT CAST ( + CAST ( 73 AS REAL ) AS INTEGER ) * 62 FROM tab2 WHERE NULL IS NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("4526")],
                            &[T("4526")],
                            &[T("4526")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SERIAL type column definition",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE sbtest1(id SERIAL, k INTEGER DEFAULT '0' NOT NULL, c CHAR(120) DEFAULT '' NOT NULL, pad CHAR(60) DEFAULT '' NOT NULL, PRIMARY KEY (id));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO sbtest1(k, c, pad) VALUES(4284, '8386864191', '67847967371'),(9261, '339736817', '3861551598704');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "arithmetic op with null casted as integer",
            set_up_script: &[
                "CREATE TABLE tab2(col0 INTEGER, col1 INTEGER, col2 INTEGER);",
                "INSERT INTO tab2 VALUES(7,31,27), (79,17,38), (78,59,26);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ALL - col0 * CAST ( col0 AS REAL ) + col1 + CAST ( NULL AS INTEGER ) AS col0 FROM tab2 AS cor0;",
                    expected: Expected::Rows {
                        columns: &[Column("col0", FLOAT8)],
                        rows: &[
                            &[Null],
                            &[Null],
                            &[Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1.2 + cast ( null as integer );",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", NUMERIC)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "inner join",
            set_up_script: &[
                "CREATE TABLE J1_TBL (i integer, j integer, t text);",
                "CREATE TABLE J2_TBL (i integer, k integer);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM J1_TBL INNER JOIN J2_TBL USING (i);",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("j", INT4), Column("t", TEXT), Column("k", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "use column in function when creating view",
            set_up_script: &[
                "CREATE TABLE base_tbl (a int PRIMARY KEY, b text DEFAULT 'Unspecified');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE VIEW rw_view15 AS SELECT a, upper(b) FROM base_tbl;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "star expression with sql value or names column",
            set_up_script: &[
                "CREATE TABLE test(y INTEGER PRIMARY KEY, z INTEGER, j TEXT);",
                "INSERT INTO test VALUES (1, 2, 'first row'), (3, 4, 'second row');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT *, 1, j FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("y", INT4), Column("z", INT4), Column("j", TEXT), Column("?column?", INT4), Column("j", TEXT)],
                        rows: &[
                            &[T("1"), T("2"), T("first row"), T("1"), T("first row")],
                            &[T("3"), T("4"), T("second row"), T("1"), T("second row")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT j, 11, * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("j", TEXT), Column("?column?", INT4), Column("y", INT4), Column("z", INT4), Column("j", TEXT)],
                        rows: &[
                            &[T("first row"), T("11"), T("1"), T("2"), T("first row")],
                            &[T("second row"), T("11"), T("3"), T("4"), T("second row")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT j, 111, *, j FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("j", TEXT), Column("?column?", INT4), Column("y", INT4), Column("z", INT4), Column("j", TEXT), Column("j", TEXT)],
                        rows: &[
                            &[T("first row"), T("111"), T("1"), T("2"), T("first row"), T("first row")],
                            &[T("second row"), T("111"), T("3"), T("4"), T("second row"), T("second row")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xmin hidden column support",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"select N.oid::bigint as id, N.xmin as state_number, nspname as name, D.description, pg_catalog.pg_get_userbyid(N.nspowner) as "owner" from pg_catalog.pg_namespace N left join pg_catalog.pg_description D on N.oid = D.objoid order by case when nspname = pg_catalog.current_schema() then -1::bigint else N.oid::bigint end;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("state_number", XID), Column("name", NAME), Column("description", TEXT), Column("owner", NAME)],
                        rows: &[
                            &[T("2200"), T("518"), T("public"), T("standard public schema"), T("pg_database_owner")],
                            &[T("11"), T("518"), T("pg_catalog"), T("system catalog schema"), T("postgres")],
                            &[T("99"), T("1"), T("pg_toast"), T("reserved schema for TOAST tables"), T("postgres")],
                            &[T("13679"), T("524"), T("information_schema"), Null, T("postgres")],
                        ],
                        tag: "SELECT 4",
                    },
                    skip: Some("xmin shows the transactions that wrote Postgres' catalog rows, which Doltgres has no row versions for"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tableoid hidden column support in join condition",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
SELECT cls.oid,
       cls.relname AS NAME,
       CASE contype
         WHEN 'p'
         THEN desp.description
         WHEN 'u'
         THEN desp.description
         WHEN 'x'
         THEN desp.description
         ELSE des.description
       END         AS COMMENT
FROM   pg_catalog.pg_index idx
       JOIN pg_catalog.pg_class cls
       ON cls.oid = indexrelid
       LEFT JOIN pg_catalog.pg_depend dep
       ON (dep.classid = cls.tableoid
           AND dep.objid = cls.oid
           AND dep.refobjsubid = '0'
           AND dep.refclassid = (SELECT oid
                                 FROM   pg_catalog.pg_class
                                 WHERE  relname = 'pg_constraint')
           AND dep.deptype = 'i')
       LEFT OUTER JOIN pg_catalog.pg_constraint con
       ON (con.tableoid = dep.refclassid
           AND con.oid = dep.refobjid)
       LEFT OUTER JOIN pg_catalog.pg_description des
       ON (des.objoid = cls.oid
           AND des.classoid = 'pg_class'::REGCLASS)
       LEFT OUTER JOIN pg_catalog.pg_description desp
       ON (desp.objoid = con.oid
           AND desp.objsubid = 0
           AND desp.classoid = 'pg_constraint'::REGCLASS)
WHERE  indrelid = 1397286223::OID
       AND contype = 'p'
"#,
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("name", NAME), Column("comment", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tableoid hidden column support in order by clause",
            assertions: &[
                ScriptTestAssertion {
                    query: "select oid from pg_class where oid = 862653097 order by tableoid",
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
