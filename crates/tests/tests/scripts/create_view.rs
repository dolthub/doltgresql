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
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "select * from myview order by pk; /* err */",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "myview" does not exist"#, position: 15, ..E }),
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from myschema.dolt_schemas;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("myview")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "view lookup follows relation order on search_path",
            set_up_script: &[
                "CREATE SCHEMA first_schema",
                "CREATE SCHEMA second_schema",
                "CREATE TABLE second_schema.source (v INT)",
                "INSERT INTO second_schema.source VALUES (42)",
                "CREATE VIEW second_schema.later_view AS SELECT v FROM second_schema.source",
                "CREATE TABLE first_schema.shadow (v INT)",
                "INSERT INTO first_schema.shadow VALUES (10)",
                "CREATE VIEW second_schema.shadow AS SELECT 20 AS v",
                "CREATE VIEW first_schema.first_view AS SELECT 30 AS v",
                "CREATE TABLE second_schema.first_view (v INT)",
                "INSERT INTO second_schema.first_view VALUES (40)",
                "CREATE VIEW first_schema.same_view AS SELECT 1 AS v",
                "CREATE VIEW second_schema.same_view AS SELECT 2 AS v",
                "SET search_path TO first_schema, second_schema",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM later_view",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM second_schema.later_view",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM shadow",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM first_view",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM same_view",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO missing_schema, second_schema, first_schema",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM later_view",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM shadow",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM first_view",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("40")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM same_view",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
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
            name: "view in later search_path schema reaches privilege checks",
            set_up_script: &[
                "CREATE SCHEMA empty_schema",
                "CREATE SCHEMA protected_schema",
                "CREATE VIEW protected_schema.target_view AS SELECT 42 AS v",
                "CREATE ROLE allowed_reader LOGIN PASSWORD 'password'",
                "CREATE ROLE denied_reader LOGIN PASSWORD 'password'",
                "GRANT USAGE ON SCHEMA empty_schema, protected_schema TO allowed_reader, denied_reader",
                "GRANT SELECT ON protected_schema.target_view TO allowed_reader",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SET search_path TO empty_schema, protected_schema",
                    expected: Expected::Tag("SET"),
                    username: "allowed_reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM target_view",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "allowed_reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO empty_schema, protected_schema",
                    expected: Expected::Tag("SET"),
                    username: "denied_reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM target_view",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for view target_view", ..E }),
                    username: "denied_reader",
                    password: "password",
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
        ScriptTest {
            name: "Views that reference each other",
            set_up_script: &[
                "CREATE TABLE cyc_t1 (a BIGINT);",
                "CREATE TABLE cyc_t2 (b BIGINT);",
                "CREATE VIEW cyc_v2 AS SELECT * FROM cyc_t1, cyc_t2;",
                "CREATE VIEW cyc_v3 AS SELECT * FROM cyc_v2;",
                "CREATE OR REPLACE VIEW cyc_v2 AS SELECT * FROM cyc_v3;",
                "CREATE VIEW cyc_self AS SELECT 1 AS x;",
                "CREATE OR REPLACE VIEW cyc_self AS SELECT * FROM cyc_self;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM cyc_v2;",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: r#"infinite recursion detected in rules for relation "cyc_v2""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM cyc_v3;",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: r#"infinite recursion detected in rules for relation "cyc_v3""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT * FROM cyc_v3) sq;",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: r#"infinite recursion detected in rules for relation "cyc_v3""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM cyc_self;",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: r#"infinite recursion detected in rules for relation "cyc_self""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM cyc_v2;",
                    expected: Expected::Error(Diagnostic { code: "42P17", message: r#"infinite recursion detected in rules for relation "cyc_v2""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c.relname, a.attname FROM pg_catalog.pg_class c JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid WHERE c.relname IN ('cyc_t1', 'cyc_t2') AND a.attnum > 0 ORDER BY c.relname, a.attnum;",
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME), Column("attname", NAME)],
                        rows: &[
                            &[T("cyc_t1"), T("a")],
                            &[T("cyc_t2"), T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c.relname, a.attname FROM pg_catalog.pg_class c JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid WHERE c.relname IN ('cyc_v2', 'cyc_v3') ORDER BY c.relname, a.attnum;",
                    skip: Some("views store their query rather than their columns, so a view in a cycle has no columns to report"),
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME), Column("attname", NAME)],
                        rows: &[
                            &[T("cyc_v2"), T("a")],
                            &[T("cyc_v2"), T("b")],
                            &[T("cyc_v3"), T("a")],
                            &[T("cyc_v3"), T("b")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW cyc_v4 AS SELECT * FROM cyc_v2;",
                    skip: Some("views store their query rather than their columns, so a view in a cycle has no columns to report"),
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Views with the same name in different schemas",
            set_up_script: &[
                "CREATE SCHEMA cyc_s1;",
                "CREATE SCHEMA cyc_s2;",
                "CREATE TABLE cyc_s2.t (a INT);",
                "INSERT INTO cyc_s2.t VALUES (1);",
                "CREATE VIEW cyc_s2.v AS SELECT a FROM cyc_s2.t;",
                "CREATE VIEW cyc_s1.v AS SELECT a FROM cyc_s2.v;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM cyc_s1.v;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
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
                            &[T("v1"), T(r#" SELECT name
   FROM a;"#)],
                            &[T("v10"), T(r#" SELECT 1 AS one,
    'a'::text AS text,
    now() AS now,
    CURRENT_DATE AS "current_date",
    COALESCE(name, 'z'::text) AS "coalesce",
    (id)::text AS id,
    (n)::integer AS n
   FROM a;"#)],
                            &[T("v11"), T(r#" SELECT id,
    d
   FROM a
  WHERE ((d > '2020-01-01'::date) AND ((id >= 1) AND (id <= 5)) AND (name IS NOT NULL) AND (id = ANY (ARRAY[1, 2, 3])));"#)],
                            &[T("v12"), T(r#" SELECT num,
    word
   FROM ( VALUES (1,'one'::text), (2,'two'::text)) v(num, word);"#)],
                            &[T("v13"), T(r#" SELECT EXTRACT(year FROM d) AS "extract",
    SUBSTRING(name FROM 2 FOR 3) AS "substring",
    TRIM(BOTH FROM name) AS btrim,
    POSITION(('a'::text) IN (name)) AS "position"
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
                            &[T("v17"), T(r#" SELECT DISTINCT ON (a_id) a_id,
    x
   FROM b
  ORDER BY a_id, x DESC NULLS LAST;"#)],
                            &[T("v2"), T(r#" SELECT id,
    name,
    d,
    n
   FROM a
  WHERE ((id > 1) AND (name ~~ 'x%'::text))
  ORDER BY name DESC, id
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
                            &[T("v4"), T(r#" SELECT a_id,
    count(*) AS count,
    sum(id) AS total
   FROM b
  GROUP BY a_id
 HAVING (count(*) > 1);"#)],
                            &[T("v5"), T(r#"(
         SELECT a.id
           FROM a
        UNION
         SELECT b.id
           FROM b
) UNION ALL
 SELECT 3 AS id;"#)],
                            &[T("v6"), T(r#" SELECT id
   FROM a
  WHERE ((EXISTS ( SELECT 1
           FROM b
          WHERE (b.a_id = a.id))) AND (id IN ( SELECT b.a_id
           FROM b)) AND (name = ( SELECT max((b.x)::text) AS max
           FROM b)));"#)],
                            &[T("v7"), T(r#" SELECT DISTINCT id AS c1,
        CASE
            WHEN (id > 1) THEN 'big'::text
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
     LEFT JOIN b b2 ON b2.id = b.id;"#), T(r#" SELECT id
   FROM a
  WHERE (EXISTS ( SELECT 1
           FROM b
          WHERE b.a_id = a.id)) AND (id IN ( SELECT b.a_id
           FROM b)) AND name = (( SELECT max(b.x::text) AS max
           FROM b));"#), T(r#" SELECT name
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

#[test]
fn test_updatable_views() {
    run_scripts(&[
        ScriptTest {
            name: "Automatically updatable views",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE uv_base (a INT PRIMARY KEY, b TEXT DEFAULT 'Unspecified');",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uv_base SELECT i, 'Row ' || i FROM generate_series(-2, 2) g(i);",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW uv_rw1 AS SELECT * FROM uv_base WHERE a > 0;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW uv_rw2 AS SELECT b AS bb, a AS aa, a + 1 AS ac FROM uv_base t WHERE t.a < 2;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW uv_ro1 AS SELECT DISTINCT a, b FROM uv_base;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW uv_ro2 AS SELECT count(*) FROM uv_base;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uv_rw1 VALUES (3, 'Row 3');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uv_rw1 (a) VALUES (4);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE uv_rw1 SET a = 5 WHERE a = 4;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM uv_rw1 WHERE b = 'Row 2';",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM uv_base ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("-2"), T("Row -2")],
                            &[T("-1"), T("Row -1")],
                            &[T("0"), T("Row 0")],
                            &[T("1"), T("Row 1")],
                            &[T("3"), T("Row 3")],
                            &[T("5"), T("Unspecified")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE uv_rw2 SET bb = 'x' || bb WHERE aa < 1 RETURNING *;",
                    expected: Expected::Rows {
                        columns: &[Column("bb", TEXT), Column("aa", INT4), Column("ac", INT4)],
                        rows: &[
                            &[T("xRow -2"), T("-2"), T("-1")],
                            &[T("xRow -1"), T("-1"), T("0")],
                            &[T("xRow 0"), T("0"), T("1")],
                        ],
                        tag: "UPDATE 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE uv_rw2 SET ac = 1;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cannot update column "ac" of view "uv_rw2""#, detail: "View columns that are not columns of their base relation are not updatable.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uv_rw2 (aa, bb) VALUES (10, 'ten') RETURNING aa, ac;",
                    expected: Expected::Rows {
                        columns: &[Column("aa", INT4), Column("ac", INT4)],
                        rows: &[
                            &[T("10"), T("11")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uv_rw2 VALUES ('q', 11, 12);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cannot insert into column "ac" of view "uv_rw2""#, detail: "View columns that are not columns of their base relation are not updatable.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM uv_rw2 v WHERE v.aa = -2 RETURNING v.*;",
                    expected: Expected::Rows {
                        columns: &[Column("bb", TEXT), Column("aa", INT4), Column("ac", INT4)],
                        rows: &[
                            &[T("xRow -2"), T("-2"), T("-1")],
                        ],
                        tag: "DELETE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uv_ro1 VALUES (1, 'x');",
                    expected: Expected::Error(Diagnostic { code: "55000", message: r#"cannot insert into view "uv_ro1""#, detail: "Views containing DISTINCT are not automatically updatable.", hint: "To enable inserting into the view, provide an INSTEAD OF INSERT trigger or an unconditional ON INSERT DO INSTEAD rule.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE uv_ro2 SET count = 1;",
                    expected: Expected::Error(Diagnostic { code: "55000", message: r#"cannot update view "uv_ro2""#, detail: "Views that return aggregate functions are not automatically updatable.", hint: "To enable updating the view, provide an INSTEAD OF UPDATE trigger or an unconditional ON UPDATE DO INSTEAD rule.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM uv_ro1;",
                    expected: Expected::Error(Diagnostic { code: "55000", message: r#"cannot delete from view "uv_ro1""#, detail: "Views containing DISTINCT are not automatically updatable.", hint: "To enable deleting from the view, provide an INSTEAD OF DELETE trigger or an unconditional ON DELETE DO INSTEAD rule.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM uv_base ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("-1"), T("xRow -1")],
                            &[T("0"), T("xRow 0")],
                            &[T("1"), T("Row 1")],
                            &[T("3"), T("Row 3")],
                            &[T("5"), T("Unspecified")],
                            &[T("10"), T("ten")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Updatable view check options and information_schema",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE uvc_base (a INT PRIMARY KEY, b TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW uvc_rw1 AS SELECT * FROM uvc_base WHERE a > 0;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW uvc_rw3 AS SELECT * FROM uvc_rw1 WHERE a < 10 WITH CHECK OPTION;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uvc_rw3 VALUES (6, 'six');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uvc_rw3 VALUES (20, 'twenty');",
                    expected: Expected::Error(Diagnostic { code: "44000", message: r#"new row violates check option for view "uvc_rw3""#, detail: "Failing row contains (20, twenty).", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uvc_rw3 VALUES (-5, 'neg');",
                    expected: Expected::Error(Diagnostic { code: "44000", message: r#"new row violates check option for view "uvc_rw1""#, detail: "Failing row contains (-5, neg).", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE uvc_rw3 SET a = a + 100 WHERE a = 6;",
                    expected: Expected::Error(Diagnostic { code: "44000", message: r#"new row violates check option for view "uvc_rw3""#, detail: "Failing row contains (106, six).", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW uvc_rw4 AS SELECT * FROM uvc_rw1 WHERE a < 10 WITH LOCAL CHECK OPTION;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uvc_rw4 VALUES (-6, 'neg');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM uvc_base ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("-6"), T("neg")],
                            &[T("6"), T("six")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW uvc_ro3 AS SELECT 1 AS one, a + 1 AS c FROM uvc_base;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_name, is_insertable_into FROM information_schema.tables WHERE table_name LIKE 'uvc_%' ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", NAME), Column("is_insertable_into", VARCHAR)],
                        rows: &[
                            &[T("uvc_base"), T("YES")],
                            &[T("uvc_ro3"), T("NO")],
                            &[T("uvc_rw1"), T("YES")],
                            &[T("uvc_rw3"), T("YES")],
                            &[T("uvc_rw4"), T("YES")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_name, is_updatable, is_insertable_into, check_option FROM information_schema.views WHERE table_name LIKE 'uvc_%' ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", NAME), Column("is_updatable", VARCHAR), Column("is_insertable_into", VARCHAR), Column("check_option", VARCHAR)],
                        rows: &[
                            &[T("uvc_ro3"), T("NO"), T("NO"), T("NONE")],
                            &[T("uvc_rw1"), T("YES"), T("YES"), T("NONE")],
                            &[T("uvc_rw3"), T("YES"), T("YES"), T("CASCADED")],
                            &[T("uvc_rw4"), T("YES"), T("YES"), T("LOCAL")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT table_name, column_name, is_updatable FROM information_schema.columns WHERE table_name LIKE 'uvc_r%' ORDER BY table_name, ordinal_position;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", NAME), Column("column_name", NAME), Column("is_updatable", VARCHAR)],
                        rows: &[
                            &[T("uvc_ro3"), T("one"), T("NO")],
                            &[T("uvc_ro3"), T("c"), T("NO")],
                            &[T("uvc_rw1"), T("a"), T("YES")],
                            &[T("uvc_rw1"), T("b"), T("YES")],
                            &[T("uvc_rw3"), T("a"), T("YES")],
                            &[T("uvc_rw3"), T("b"), T("YES")],
                            &[T("uvc_rw4"), T("a"), T("YES")],
                            &[T("uvc_rw4"), T("b"), T("YES")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_view_privileges() {
    run_scripts(&[
        ScriptTest {
            name: "CREATE VIEW checks table privileges only when the view is read",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER regress_cv_user;",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA cv_schema;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT ALL ON SCHEMA cv_schema TO regress_cv_user;",
                    expected: Expected::Tag("GRANT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE cv_schema.cv_base (a INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION regress_cv_user;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW cv_schema.cv_view AS SELECT * FROM cv_schema.cv_base;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM cv_schema.cv_view;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table cv_base", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET SESSION AUTHORIZATION;",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
