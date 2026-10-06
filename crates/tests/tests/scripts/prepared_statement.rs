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
fn test_prepared_error_handling() {
    run_scripts_repeated(&[
        ScriptTest {
            name: "error handling doesn't foul session",
            set_up_script: &[
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "insert into test values (1, 1), (2, 2), (3, 3), (4, 4), (5, 5), (6, 6), (7, 7);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select v1 from doesNotExist where pk = 1;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "doesnotexist" does not exist"#, position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v1 from test where pk = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v1 from test where pk = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v1 from test where pk = 3;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v1 from test where pk = 4;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT8)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v1 from test where pk = 5;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v1 from test where pk = 6;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT8)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select v1 from test where pk = 7;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
    ], 20);
}

#[test]
fn test_prepared_pg_catalog() {
    run_scripts(&[
        ScriptTest {
            name: "pg_namespace",
            set_up_script: &[
                "CREATE SCHEMA testschema;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "pg_catalog"."pg_namespace" WHERE nspname=$1;"#,
                    bind_vars: &[BindVar::Str("testschema")],
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("nspname", NAME), Column("nspowner", OID), Column("nspacl", ACLITEM_ARRAY)],
                        rows: &[
                            &[T("16384"), T("testschema"), T("10"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "pg_catalog"."pg_namespace" WHERE oid=$1;"#,
                    bind_vars: &[BindVar::Int(2638679668)],
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("nspname", NAME), Column("nspowner", OID), Column("nspacl", ACLITEM_ARRAY)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_tables",
            set_up_script: &[
                "CREATE TABLE testing (pk INT primary key, v1 INT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "pg_catalog"."pg_tables" WHERE tablename=$1;"#,
                    bind_vars: &[BindVar::Str("testing")],
                    expected: Expected::Rows {
                        columns: &[Column("schemaname", NAME), Column("tablename", NAME), Column("tableowner", NAME), Column("tablespace", NAME), Column("hasindexes", BOOL), Column("hasrules", BOOL), Column("hastriggers", BOOL), Column("rowsecurity", BOOL)],
                        rows: &[
                            &[T("public"), T("testing"), T("postgres"), Null, T("t"), T("f"), T("f"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT count(*) FROM "pg_catalog"."pg_tables" WHERE schemaname=$1;"#,
                    bind_vars: &[BindVar::Str("pg_catalog")],
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("64")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_class",
            set_up_script: &[
                "CREATE SCHEMA testschema;",
                "CREATE TABLE testschema.testtable (id int primary key, v1 text)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select distinct relnamespace from pg_catalog.pg_class c INNER JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid WHERE n.nspname=$1;",
                    bind_vars: &[BindVar::Str("testschema")],
                    expected: Expected::Rows {
                        columns: &[Column("relnamespace", OID)],
                        rows: &[
                            &[T("16384")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT c.oid,pg_catalog.pg_get_expr(c.relpartbound, c.oid) as partition_expr,  pg_catalog.pg_get_partkeydef(c.oid) as partition_key 
FROM pg_catalog.pg_class c
WHERE c.relnamespace=$1 AND c.relkind not in ('i','I','c') and c.oid not in (select oid from pg_catalog.pg_class where left(relname, 5) = 'dolt_');"#,
                    bind_vars: &[BindVar::Int(2638679668)],
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("partition_expr", TEXT), Column("partition_key", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT c.oid,d.description,pg_catalog.pg_get_expr(c.relpartbound, c.oid) as partition_expr,  pg_catalog.pg_get_partkeydef(c.oid) as partition_key 
FROM pg_catalog.pg_class c
LEFT OUTER JOIN pg_catalog.pg_description d ON d.objoid=c.oid AND d.objsubid=0 AND d.classoid='pg_class'::regclass
WHERE c.relnamespace=$1 AND c.relkind not in ('i','I','c') and c.oid not in (select oid from pg_catalog.pg_class where left(relname, 5) = 'dolt_');"#,
                    bind_vars: &[BindVar::Int(2638679668)],
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("description", TEXT), Column("partition_expr", TEXT), Column("partition_key", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT d.description from pg_catalog.pg_description d WHERE d.classoid='pg_class'::regclass",
                    expected: Expected::Rows {
                        columns: &[Column("description", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select c.oid,pg_catalog.pg_total_relation_size(c.oid) as total_rel_size,pg_catalog.pg_relation_size(c.oid) as rel_size FROM pg_class c WHERE c.relnamespace=$1 and c.oid not in (select oid from pg_catalog.pg_class where left(relname, 5) = 'dolt_');",
                    bind_vars: &[BindVar::Int(2638679668)],
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("total_rel_size", INT8), Column("rel_size", INT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT c.relname, a.attrelid, a.attname, a.atttypid, pg_catalog.pg_get_expr(ad.adbin, ad.adrelid, true) as def_value,dsc.description,dep.objid 
FROM pg_catalog.pg_attribute a 
INNER JOIN pg_catalog.pg_class c ON (a.attrelid=c.oid) 
LEFT OUTER JOIN pg_catalog.pg_attrdef ad ON (a.attrelid=ad.adrelid AND a.attnum = ad.adnum) 
LEFT OUTER JOIN pg_catalog.pg_description dsc ON (c.oid=dsc.objoid AND a.attnum = dsc.objsubid) 
LEFT OUTER JOIN pg_depend dep on dep.refobjid = a.attrelid AND dep.deptype = 'i' and dep.refobjsubid = a.attnum and dep.classid = dep.refclassid 
WHERE NOT a.attisdropped AND c.relkind not in ('i','I','c') AND c.oid=$1 ORDER BY a.attnum"#,
                    bind_vars: &[BindVar::Int(1712283605)],
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME), Column("attrelid", OID), Column("attname", NAME), Column("atttypid", OID), Column("def_value", TEXT), Column("description", TEXT), Column("objid", OID)],
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
fn test_prepared_statements() {
    run_scripts(&[
        ScriptTest {
            name: "Expressions without tables",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT CONCAT($1::text, $2::text)",
                    bind_vars: &[BindVar::Str("hello"), BindVar::Str("world")],
                    expected: Expected::Rows {
                        columns: &[Column("concat", TEXT)],
                        rows: &[
                            &[T("helloworld")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT $1::integer + $2::integer",
                    bind_vars: &[BindVar::Int(1), BindVar::Int(2)],
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select $1 as test",
                    bind_vars: &[BindVar::Str("hello")],
                    expected: Expected::Rows {
                        columns: &[Column("test", TEXT)],
                        rows: &[
                            &[T("hello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Expressions with tables",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = $1);",
                    bind_vars: &[BindVar::Str("public")],
                    expected: Expected::Rows {
                        columns: &[Column("exists", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nspname FROM pg_namespace LIMIT $1;",
                    bind_vars: &[BindVar::Int(1)],
                    expected: Expected::Rows {
                        columns: &[Column("nspname", NAME)],
                        rows: &[
                            &[T("pg_toast")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nspname FROM pg_namespace OFFSET $1;",
                    bind_vars: &[BindVar::Int(1)],
                    expected: Expected::Rows {
                        columns: &[Column("nspname", NAME)],
                        rows: &[
                            &[T("pg_catalog")],
                            &[T("public")],
                            &[T("information_schema")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Integer insert with string bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("2"), BindVar::Str("3"), BindVar::Str("4")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("3"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Str("2")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Str("3")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 + $1 = $2;",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("3")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE pk + v1 = $1;",
                    bind_vars: &[BindVar::Str("3")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1::integer + $2::integer;",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("3")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("3"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Integer insert with binary bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Int(2), BindVar::Int(3), BindVar::Int(4)],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("3"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Int(2)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Int(3)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 + $1 = $2;",
                    bind_vars: &[BindVar::Int(1), BindVar::Int(3)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE pk + v1 = $1;",
                    bind_vars: &[BindVar::Int(3)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1::integer + $2::integer;",
                    bind_vars: &[BindVar::Int(1), BindVar::Int(3)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("3"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2) returning *",
                    bind_vars: &[BindVar::Int(2), BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("2"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Integer types",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v2 SMALLINT, v4 INTEGER, v5 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2, $3, $4) returning *",
                    bind_vars: &[BindVar::Int(1), BindVar::Int(10), BindVar::Int(100), BindVar::Int(1000)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v2", INT2), Column("v4", INT4), Column("v5", INT8)],
                        rows: &[
                            &[T("1"), T("10"), T("100"), T("1000")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2, $3, $4) returning *",
                    bind_vars: &[BindVar::Int(2), BindVar::Null, BindVar::Null, BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v2", INT2), Column("v4", INT4), Column("v5", INT8)],
                        rows: &[
                            &[T("2"), Null, Null, Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Integer update",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Int(2), BindVar::Int(3), BindVar::Int(4)],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test set v1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Int(5), BindVar::Int(1)],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Int(5)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Integer delete",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Int(2), BindVar::Int(3), BindVar::Int(4)],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk = $1;",
                    bind_vars: &[BindVar::Int(1)],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("3"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "String insert",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, s character varying(20));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Str("hello"), BindVar::Int(3), BindVar::Str("goodbye")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("s", VARCHAR)],
                        rows: &[
                            &[T("1"), T("hello")],
                            &[T("3"), T("goodbye")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE s = $1;",
                    bind_vars: &[BindVar::Str("hello")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("s", VARCHAR)],
                        rows: &[
                            &[T("1"), T("hello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE s = concat($1::text, $2::text);",
                    bind_vars: &[BindVar::Str("he"), BindVar::Str("llo")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("s", VARCHAR)],
                        rows: &[
                            &[T("1"), T("hello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE concat(s, '!') = $1",
                    bind_vars: &[BindVar::Str("hello!")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("s", VARCHAR)],
                        rows: &[
                            &[T("1"), T("hello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2) returning *",
                    bind_vars: &[BindVar::Int(2), BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("s", VARCHAR)],
                        rows: &[
                            &[T("2"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "String update",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, s character varying(20));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Str("hello"), BindVar::Int(3), BindVar::Str("goodbye")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test set s = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Str("new value"), BindVar::Int(1)],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE s = $1;",
                    bind_vars: &[BindVar::Str("new value")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("s", VARCHAR)],
                        rows: &[
                            &[T("1"), T("new value")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "String delete",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, s character varying(20));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Str("hello"), BindVar::Int(3), BindVar::Str("goodbye")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE s = $1;",
                    bind_vars: &[BindVar::Str("hello")],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("s", VARCHAR)],
                        rows: &[
                            &[T("3"), T("goodbye")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Float insert with string bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, f1 DOUBLE PRECISION);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("1.1"), BindVar::Str("3"), BindVar::Str("3.3")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("f1", FLOAT8)],
                        rows: &[
                            &[T("1"), T("1.1")],
                            &[T("3"), T("3.3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE f1 = $1;",
                    bind_vars: &[BindVar::Str("1.1")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("f1", FLOAT8)],
                        rows: &[
                            &[T("1"), T("1.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE f1 + $1 = $2;",
                    bind_vars: &[BindVar::Str("1.0"), BindVar::Str("2.1")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("f1", FLOAT8)],
                        rows: &[
                            &[T("1"), T("1.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE f1 = $1::decimal + $2::decimal;",
                    bind_vars: &[BindVar::Str("1.0"), BindVar::Str("0.1")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("f1", FLOAT8)],
                        rows: &[
                            &[T("1"), T("1.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Float insert with binary bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, f1 DOUBLE PRECISION);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Float64(1.1), BindVar::Int(3), BindVar::Float64(3.3)],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("f1", FLOAT8)],
                        rows: &[
                            &[T("1"), T("1.1")],
                            &[T("3"), T("3.3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE f1 = $1;",
                    bind_vars: &[BindVar::Float64(1.1)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("f1", FLOAT8)],
                        rows: &[
                            &[T("1"), T("1.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE f1 + $1 = $2;",
                    bind_vars: &[BindVar::Float64(1.0), BindVar::Float64(2.1)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("f1", FLOAT8)],
                        rows: &[
                            &[T("1"), T("1.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE f1 = $1::decimal + $2::decimal;",
                    bind_vars: &[BindVar::Float64(1.0), BindVar::Float64(0.1)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("f1", FLOAT8)],
                        rows: &[
                            &[T("1"), T("1.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Float update",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, f1 DOUBLE PRECISION);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Float64(1.1), BindVar::Int(3), BindVar::Float64(3.3)],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test set f1 = $1 WHERE f1 = $2;",
                    bind_vars: &[BindVar::Float64(2.2), BindVar::Float64(1.1)],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE f1 = $1;",
                    bind_vars: &[BindVar::Float64(2.2)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("f1", FLOAT8)],
                        rows: &[
                            &[T("1"), T("2.2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Float delete",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, f1 DOUBLE PRECISION);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Float64(1.1), BindVar::Int(3), BindVar::Float64(3.3)],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE f1 = $1;",
                    bind_vars: &[BindVar::Float64(1.1)],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("f1", FLOAT8)],
                        rows: &[
                            &[T("3"), T("3.3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Date insert, update, delete with string bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 DATE);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("2022-02-02"), BindVar::Str("3"), BindVar::Str("2024-04-01 -07")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[
                            &[T("1"), T("2022-02-02")],
                            &[T("3"), T("2024-04-01")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Str("2022-02-02")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[
                            &[T("1"), T("2022-02-02")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Str("2022-02-03")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test set v1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Str("2022-02-03"), BindVar::Str("1")],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Str("2022-02-03")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[
                            &[T("1"), T("2022-02-03")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk = $1;",
                    bind_vars: &[BindVar::Str("1")],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[
                            &[T("3"), T("2024-04-01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2) returning *;",
                    bind_vars: &[BindVar::Str("5"), BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[
                            &[T("5"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Date insert, update, delete with binary bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 DATE);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Date(Time { year: 2022, month: 2, day: 2, hour: 0, minute: 0, second: 0, nanosecond: 0, offset_seconds: 0 }), BindVar::Int(3), BindVar::Date(Time { year: 2024, month: 4, day: 1, hour: 0, minute: 0, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[
                            &[T("1"), T("2022-02-02")],
                            &[T("3"), T("2024-04-01")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Date(Time { year: 2022, month: 2, day: 2, hour: 0, minute: 0, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[
                            &[T("1"), T("2022-02-02")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Date(Time { year: 2022, month: 2, day: 3, hour: 0, minute: 0, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test set v1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Date(Time { year: 2022, month: 2, day: 3, hour: 0, minute: 0, second: 0, nanosecond: 0, offset_seconds: 0 }), BindVar::Int(1)],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE v1 = $1;",
                    bind_vars: &[BindVar::Date(Time { year: 2022, month: 2, day: 3, hour: 0, minute: 0, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[
                            &[T("1"), T("2022-02-03")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk = $1;",
                    bind_vars: &[BindVar::Int(1)],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", DATE)],
                        rows: &[
                            &[T("3"), T("2024-04-01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Timestamp insert with string bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, t1 TIMESTAMP);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("2023-01-15 14:30"), BindVar::Str("2"), BindVar::Str("2024-12-25 09:15:30")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIMESTAMP)],
                        rows: &[
                            &[T("1"), T("2023-01-15 14:30:00")],
                            &[T("2"), T("2024-12-25 09:15:30")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Str("2023-01-15 14:30")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIMESTAMP)],
                        rows: &[
                            &[T("1"), T("2023-01-15 14:30:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET t1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Str("2023-01-15 16:45"), BindVar::Str("1")],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Str("2023-01-15 16:45")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIMESTAMP)],
                        rows: &[
                            &[T("1"), T("2023-01-15 16:45:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Str("2023-01-15 16:45")],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIMESTAMP)],
                        rows: &[
                            &[T("2"), T("2024-12-25 09:15:30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2) returning *;",
                    bind_vars: &[BindVar::Str("3"), BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIMESTAMP)],
                        rows: &[
                            &[T("3"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Timestamp insert with binary bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, t1 TIMESTAMP);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Timestamp(Time { year: 2023, month: 1, day: 15, hour: 14, minute: 30, second: 0, nanosecond: 0, offset_seconds: 0 }), BindVar::Int(2), BindVar::Timestamp(Time { year: 2024, month: 12, day: 25, hour: 9, minute: 15, second: 30, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIMESTAMP)],
                        rows: &[
                            &[T("1"), T("2023-01-15 14:30:00")],
                            &[T("2"), T("2024-12-25 09:15:30")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Timestamp(Time { year: 2023, month: 1, day: 15, hour: 14, minute: 30, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIMESTAMP)],
                        rows: &[
                            &[T("1"), T("2023-01-15 14:30:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET t1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Timestamp(Time { year: 2023, month: 1, day: 15, hour: 16, minute: 45, second: 0, nanosecond: 0, offset_seconds: 0 }), BindVar::Int(1)],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Timestamp(Time { year: 2023, month: 1, day: 15, hour: 16, minute: 45, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIMESTAMP)],
                        rows: &[
                            &[T("1"), T("2023-01-15 16:45:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Timestamp(Time { year: 2023, month: 1, day: 15, hour: 16, minute: 45, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIMESTAMP)],
                        rows: &[
                            &[T("2"), T("2024-12-25 09:15:30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Timestamp with timezone insert with string bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, t1 TIMESTAMPTZ);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("2023-01-15 14:30:00+00"), BindVar::Str("2"), BindVar::Str("2024-12-25 09:15:30-05")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Str("2023-01-15 14:30:00+00")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8)],
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
            name: "Timestamp with timezone insert with binary bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, t1 TIMESTAMPTZ);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Time(Time { year: 2023, month: 1, day: 15, hour: 14, minute: 30, second: 0, nanosecond: 0, offset_seconds: 0 }), BindVar::Int(2), BindVar::Time(Time { year: 2024, month: 12, day: 25, hour: 9, minute: 15, second: 30, nanosecond: 0, offset_seconds: -18000 })],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Time(Time { year: 2023, month: 1, day: 15, hour: 14, minute: 30, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8)],
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
            name: "Time insert with string bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, t1 TIME);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("14:30:00"), BindVar::Str("2"), BindVar::Str("09:15:30")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIME)],
                        rows: &[
                            &[T("1"), T("14:30:00")],
                            &[T("2"), T("09:15:30")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Str("14:30:00")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIME)],
                        rows: &[
                            &[T("1"), T("14:30:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET t1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Str("16:45:00"), BindVar::Str("1")],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Str("16:45:00")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIME)],
                        rows: &[
                            &[T("1"), T("16:45:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Str("16:45:00")],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIME)],
                        rows: &[
                            &[T("2"), T("09:15:30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Time insert with binary bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, t1 TIME);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Time(Time { year: 0, month: 1, day: 1, hour: 14, minute: 30, second: 0, nanosecond: 0, offset_seconds: 0 }), BindVar::Int(2), BindVar::Time(Time { year: 0, month: 1, day: 1, hour: 9, minute: 15, second: 30, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIME)],
                        rows: &[
                            &[T("1"), T("14:30:00")],
                            &[T("2"), T("09:15:30")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Time(Time { year: 0, month: 1, day: 1, hour: 14, minute: 30, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIME)],
                        rows: &[
                            &[T("1"), T("14:30:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET t1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Time(Time { year: 0, month: 1, day: 1, hour: 16, minute: 45, second: 0, nanosecond: 0, offset_seconds: 0 }), BindVar::Int(1)],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Time(Time { year: 0, month: 1, day: 1, hour: 16, minute: 45, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIME)],
                        rows: &[
                            &[T("1"), T("16:45:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE t1 = $1;",
                    bind_vars: &[BindVar::Time(Time { year: 0, month: 1, day: 1, hour: 16, minute: 45, second: 0, nanosecond: 0, offset_seconds: 0 })],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("t1", TIME)],
                        rows: &[
                            &[T("2"), T("09:15:30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UUID insert with string bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, u1 UUID);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("550e8400-e29b-41d4-a716-446655440000"), BindVar::Str("2"), BindVar::Str("6ba7b810-9dad-11d1-80b4-00c04fd430c8")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("u1", UUID)],
                        rows: &[
                            &[T("1"), T("550e8400-e29b-41d4-a716-446655440000")],
                            &[T("2"), T("6ba7b810-9dad-11d1-80b4-00c04fd430c8")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE u1 = $1;",
                    bind_vars: &[BindVar::Str("550e8400-e29b-41d4-a716-446655440000")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("u1", UUID)],
                        rows: &[
                            &[T("1"), T("550e8400-e29b-41d4-a716-446655440000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET u1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Str("123e4567-e89b-12d3-a456-426614174000"), BindVar::Str("1")],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE u1 = $1;",
                    bind_vars: &[BindVar::Str("123e4567-e89b-12d3-a456-426614174000")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("u1", UUID)],
                        rows: &[
                            &[T("1"), T("123e4567-e89b-12d3-a456-426614174000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE u1 = $1;",
                    bind_vars: &[BindVar::Str("123e4567-e89b-12d3-a456-426614174000")],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("u1", UUID)],
                        rows: &[
                            &[T("2"), T("6ba7b810-9dad-11d1-80b4-00c04fd430c8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2) returning *;",
                    bind_vars: &[BindVar::Str("3"), BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("u1", UUID)],
                        rows: &[
                            &[T("3"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UUID insert with binary bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, u1 UUID);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4);",
                    bind_vars: &[BindVar::Int(1), BindVar::Uuid([85, 14, 132, 0, 226, 155, 65, 212, 167, 22, 68, 102, 85, 68, 0, 0]), BindVar::Int(2), BindVar::Uuid([107, 167, 184, 16, 157, 173, 17, 209, 128, 180, 0, 192, 79, 212, 48, 200])],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("u1", UUID)],
                        rows: &[
                            &[T("1"), T("550e8400-e29b-41d4-a716-446655440000")],
                            &[T("2"), T("6ba7b810-9dad-11d1-80b4-00c04fd430c8")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE u1 = $1;",
                    bind_vars: &[BindVar::Uuid([85, 14, 132, 0, 226, 155, 65, 212, 167, 22, 68, 102, 85, 68, 0, 0])],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("u1", UUID)],
                        rows: &[
                            &[T("1"), T("550e8400-e29b-41d4-a716-446655440000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET u1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Uuid([18, 62, 69, 103, 232, 155, 18, 211, 164, 86, 66, 102, 20, 23, 64, 0]), BindVar::Int(1)],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE u1 = $1;",
                    bind_vars: &[BindVar::Uuid([18, 62, 69, 103, 232, 155, 18, 211, 164, 86, 66, 102, 20, 23, 64, 0])],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("u1", UUID)],
                        rows: &[
                            &[T("1"), T("123e4567-e89b-12d3-a456-426614174000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE u1 = $1;",
                    bind_vars: &[BindVar::Uuid([18, 62, 69, 103, 232, 155, 18, 211, 164, 86, 66, 102, 20, 23, 64, 0])],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("u1", UUID)],
                        rows: &[
                            &[T("2"), T("6ba7b810-9dad-11d1-80b4-00c04fd430c8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Numeric/Decimal insert with string bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, n1 NUMERIC(10,2), n2 DECIMAL(8,3));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2, $3), ($4, $5, $6);",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("123.45"), BindVar::Str("67.890"), BindVar::Str("2"), BindVar::Str("999.99"), BindVar::Str("12.345")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("n1", NUMERIC), Column("n2", NUMERIC)],
                        rows: &[
                            &[T("1"), T("123.45"), T("67.890")],
                            &[T("2"), T("999.99"), T("12.345")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE n1 = $1;",
                    bind_vars: &[BindVar::Str("123.45")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("n1", NUMERIC), Column("n2", NUMERIC)],
                        rows: &[
                            &[T("1"), T("123.45"), T("67.890")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET n1 = $1, n2 = $2 WHERE pk = $3;",
                    bind_vars: &[BindVar::Str("456.78"), BindVar::Str("98.765"), BindVar::Str("1")],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE n1 = $1;",
                    bind_vars: &[BindVar::Str("456.78")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("n1", NUMERIC), Column("n2", NUMERIC)],
                        rows: &[
                            &[T("1"), T("456.78"), T("98.765")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE n2 = $1;",
                    bind_vars: &[BindVar::Str("98.765")],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("n1", NUMERIC), Column("n2", NUMERIC)],
                        rows: &[
                            &[T("2"), T("999.99"), T("12.345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2, $3) returning *;",
                    bind_vars: &[BindVar::Str("3"), BindVar::Null, BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("n1", NUMERIC), Column("n2", NUMERIC)],
                        rows: &[
                            &[T("3"), Null, Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Numeric/Decimal insert with binary bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, n1 NUMERIC(10,2), n2 DECIMAL(8,3));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2, $3), ($4, $5, $6);",
                    bind_vars: &[BindVar::Int(1), BindVar::Numeric("123.45"), BindVar::Numeric("67.890"), BindVar::Int(2), BindVar::Numeric("999.99"), BindVar::Numeric("12.345")],
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("n1", NUMERIC), Column("n2", NUMERIC)],
                        rows: &[
                            &[T("1"), T("123.45"), T("67.890")],
                            &[T("2"), T("999.99"), T("12.345")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE n1 = $1;",
                    bind_vars: &[BindVar::Numeric("123.45")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("n1", NUMERIC), Column("n2", NUMERIC)],
                        rows: &[
                            &[T("1"), T("123.45"), T("67.890")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET n1 = $1, n2 = $2 WHERE pk = $3;",
                    bind_vars: &[BindVar::Numeric("456.78"), BindVar::Numeric("98.765"), BindVar::Int(1)],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE n1 = $1;",
                    bind_vars: &[BindVar::Numeric("456.78")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("n1", NUMERIC), Column("n2", NUMERIC)],
                        rows: &[
                            &[T("1"), T("456.78"), T("98.765")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE n2 = $1;",
                    bind_vars: &[BindVar::Numeric("98.765")],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("n1", NUMERIC), Column("n2", NUMERIC)],
                        rows: &[
                            &[T("2"), T("999.99"), T("12.345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Boolean insert with string bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, b1 BOOLEAN);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4), ($5, $6);",
                    bind_vars: &[BindVar::Str("1"), BindVar::Str("true"), BindVar::Str("2"), BindVar::Str("false"), BindVar::Str("3"), BindVar::Str("true")],
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("1"), T("t")],
                            &[T("2"), T("f")],
                            &[T("3"), T("t")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE b1 = $1;",
                    bind_vars: &[BindVar::Str("true")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("1"), T("t")],
                            &[T("3"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE b1 = $1;",
                    bind_vars: &[BindVar::Str("false")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("2"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET b1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Str("false"), BindVar::Str("1")],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE b1 = $1 ORDER BY pk;",
                    bind_vars: &[BindVar::Str("false")],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("1"), T("f")],
                            &[T("2"), T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE b1 = $1;",
                    bind_vars: &[BindVar::Str("false")],
                    expected: Expected::Tag("DELETE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("3"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2) returning *;",
                    bind_vars: &[BindVar::Str("4"), BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("4"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Boolean insert with binary bindvars",
            set_up_script: &[
                "drop table if exists test",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, b1 BOOLEAN);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES ($1, $2), ($3, $4), ($5, $6);",
                    bind_vars: &[BindVar::Int(1), BindVar::Bool(true), BindVar::Int(2), BindVar::Bool(false), BindVar::Int(3), BindVar::Bool(true)],
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("1"), T("t")],
                            &[T("2"), T("f")],
                            &[T("3"), T("t")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE b1 = $1;",
                    bind_vars: &[BindVar::Bool(true)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("1"), T("t")],
                            &[T("3"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE b1 = $1;",
                    bind_vars: &[BindVar::Bool(false)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("2"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET b1 = $1 WHERE pk = $2;",
                    bind_vars: &[BindVar::Bool(false), BindVar::Int(1)],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test WHERE b1 = $1 ORDER BY pk;",
                    bind_vars: &[BindVar::Bool(false)],
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("1"), T("f")],
                            &[T("2"), T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE b1 = $1;",
                    bind_vars: &[BindVar::Bool(false)],
                    expected: Expected::Tag("DELETE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("b1", BOOL)],
                        rows: &[
                            &[T("3"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_viewdef function",
            set_up_script: &[
                "CREATE TABLE test (id int, name text)",
                "INSERT INTO test VALUES (1,'desk'), (2,'chair')",
                "CREATE VIEW test_view AS SELECT name FROM test",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select pg_get_viewdef($1::regclass);",
                    bind_vars: &[BindVar::Str("test_view")],
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_viewdef", TEXT)],
                        rows: &[
                            &[T(r#" SELECT test.name
   FROM test;"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "insert returning",
            set_up_script: &[
                "CREATE TABLE test (id serial, name text)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test (name) VALUES ($1) RETURNING id;",
                    bind_vars: &[BindVar::Str("test_name")],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
        ScriptTest {
            name: "define placeholder unordered",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT $3::text, $1::integer + $2::integer",
                    bind_vars: &[BindVar::Int(1), BindVar::Int(3), BindVar::Str("hi")],
                    expected: Expected::Rows {
                        columns: &[Column("text", TEXT), Column("?column?", INT4)],
                        rows: &[
                            &[T("hi"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bytea with binary bindvars",
            set_up_script: &[
                "CREATE TABLE t_bytea (id INTEGER primary key, v1 BYTEA);",
                r#"INSERT INTO t_bytea VALUES (1, E'\\xDEADBEEF'), (2, '\xC0FFEE'), (3, ''), (4, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bytea ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BYTEA)],
                        rows: &[
                            &[T("1"), T(r#"\xdeadbeef"#)],
                            &[T("2"), T(r#"\xc0ffee"#)],
                            &[T("3"), T(r#"\x"#)],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bytea WHERE v1 = $1 ORDER BY id;",
                    bind_vars: &[BindVar::Bytes(&[192, 255, 238])],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BYTEA)],
                        rows: &[
                            &[T("2"), T(r#"\xc0ffee"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t_bytea SET v1 = $1 WHERE id = $2;",
                    bind_vars: &[BindVar::Str(r#"\xC0FFEE"#), BindVar::Int(4)],
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bytea WHERE v1 = $1 ORDER BY id;",
                    bind_vars: &[BindVar::Bytes(&[192, 255, 238])],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BYTEA)],
                        rows: &[
                            &[T("2"), T(r#"\xc0ffee"#)],
                            &[T("4"), T(r#"\xc0ffee"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM t_bytea WHERE v1 = $1;",
                    bind_vars: &[BindVar::Bytes(&[222, 173, 190, 239])],
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_bytea ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BYTEA)],
                        rows: &[
                            &[T("2"), T(r#"\xc0ffee"#)],
                            &[T("3"), T(r#"\x"#)],
                            &[T("4"), T(r#"\xc0ffee"#)],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_bytea VALUES ($1, $2) returning *;",
                    bind_vars: &[BindVar::Int(5), BindVar::Null],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v1", BYTEA)],
                        rows: &[
                            &[T("5"), Null],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Bind parameter to compatible different types",
            set_up_script: &[
                "CREATE TABLE text_test (id text, code varchar(10))",
                "CREATE TABLE num_test(small int2, large int8, other float4)",
                "INSERT INTO text_test values ('foo', 'bar'), ('bar', 'foo')",
                "INSERT INTO num_test values (0,0,0), (1, 2, 1.5)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM text_test where id = any($1) and code = any($1)",
                    bind_vars: &[BindVar::StrArray(&["foo", "bar"])],
                    expected: Expected::Rows {
                        columns: &[Column("id", TEXT), Column("code", VARCHAR)],
                        rows: &[
                            &[T("foo"), T("bar")],
                            &[T("bar"), T("foo")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * from num_test where small = $1 and large = $1 and other = $1",
                    bind_vars: &[BindVar::Int(0)],
                    expected: Expected::Rows {
                        columns: &[Column("small", INT2), Column("large", INT8), Column("other", FLOAT4)],
                        rows: &[
                            &[T("0"), T("0"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * from num_test where small = $1::INTEGER and large = $1::INTEGER and other = $1::INTEGER",
                    bind_vars: &[BindVar::Int(0)],
                    expected: Expected::Rows {
                        columns: &[Column("small", INT2), Column("large", INT8), Column("other", FLOAT4)],
                        rows: &[
                            &[T("0"), T("0"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM num_test where small = $1 or other = $1",
                    bind_vars: &[BindVar::Float64(1.5)],
                    expected: Expected::ClientError("failed to encode args[0]: cannot convert 1.5 to int64"),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Cannot bind parameter to column with incompatible type",
            set_up_script: &[
                "CREATE TABLE text_test (fullname text, id int)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM text_test where fullname = $1 and id = $1",
                    bind_vars: &[BindVar::Int(1)],
                    expected: Expected::Error(Diagnostic { code: "42883", message: "operator does not exist: integer = text", hint: "No operator matches the given name and argument types. You might need to add explicit type casts.", position: 52, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "merge join planning with bindvar equality filter on indexed join column",
            set_up_script: &[
                "CREATE TABLE t1 (pk INT PRIMARY KEY, fk INT, val INT);",
                "CREATE INDEX t1_fk_idx ON t1(fk);",
                "CREATE TABLE t2 (pk INT PRIMARY KEY, val INT);",
                "INSERT INTO t1 VALUES (1, 1, 100), (2, 2, 200), (3, 3, 300);",
                "INSERT INTO t2 VALUES (1, 10), (2, 20), (3, 30);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT t1.val, t2.val FROM t1 JOIN t2 ON t1.fk = t2.pk WHERE t1.fk = $1;",
                    bind_vars: &[BindVar::Int(1)],
                    expected: Expected::Rows {
                        columns: &[Column("val", INT4), Column("val", INT4)],
                        rows: &[
                            &[T("100"), T("10")],
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
