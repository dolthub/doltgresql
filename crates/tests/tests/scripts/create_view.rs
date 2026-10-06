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
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW search_path;",
                    expected: Expected::Rows {
                        columns: &[Column("@@session.search_path", TEXT)],
                        rows: &[
                            &[T("testschema")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from myview order by pk; /* err */",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: myview", ..E }),
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
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW search_path;",
                    expected: Expected::Rows {
                        columns: &[Column("@@session.search_path", TEXT)],
                        rows: &[
                            &[T("testschema, myschema")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select v1 from myview order by pk;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: myview", ..E }),
                    flow: Flow::Query,
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
