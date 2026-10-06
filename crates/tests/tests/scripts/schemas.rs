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
fn test_schemas() {
    run_scripts(&[
        ScriptTest {
            name: "implicit schema with index gets created in public schema",
            set_up_script: &[
                "create table employees (id int, last_name varchar(255), first_name varchar(255), primary key(id));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO employees VALUES (1, 'John', 'Doe'), (2, 'Jane', 'Doe');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM employees;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("last_name", VARCHAR), Column("first_name", VARCHAR)],
                        rows: &[
                            &[T("1"), T("John"), T("Doe")],
                            &[T("2"), T("Jane"), T("Doe")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.employees;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("last_name", VARCHAR), Column("first_name", VARCHAR)],
                        rows: &[
                            &[T("1"), T("John"), T("Doe")],
                            &[T("2"), T("Jane"), T("Doe")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "table gets created in public schema by default",
            set_up_script: &[
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO public.test VALUES (1, 1), (2, 2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "table creation respects search_path",
            set_up_script: &[
                "create schema postgres",
                "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO postgres.test VALUES (1, 1), (2, 2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM postgres.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop table",
            set_up_script: &[
                "CREATE TABLE t1 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE t2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE t3 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE t4 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (1, 1), (2, 2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (3, 3), (4, 4);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3 VALUES (1, 1), (2, 2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t4 VALUES (3, 3), (4, 4);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "drop table t1",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "drop table public.t2",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "drop table if exists t3",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "drop table if exists public.t4",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (1, 1), (2, 2);",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "t1" does not exist"#, position: 13, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (3, 3), (4, 4);",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "t2" does not exist"#, position: 13, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3 VALUES (1, 1), (2, 2);",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "t3" does not exist"#, position: 13, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t4 VALUES (3, 3), (4, 4);",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "t4" does not exist"#, position: 13, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "drop table if exists t1",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"table "t1" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "drop table t1",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"table "t1" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter table",
            set_up_script: &[
                "CREATE TABLE t1 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE t2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "alter table t1 add column v2 BIGINT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table public.t2 add column v2 BIGINT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (5, 5, 5), (6, 6, 6);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.t2 VALUES (7, 7, 7), (8, 8, 8);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "table creation fails with no schema available",
            set_up_script: &[
                "set search_path to ''",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: "no schema has been selected to create in", position: 14, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"set search_path to "$user""#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: "no schema has been selected to create in", position: 14, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "create schema postgres",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "qualified type in a column default with an empty search path",
            set_up_script: &[
                "SELECT pg_catalog.set_config('search_path', '', false);",
                "CREATE TYPE public.registration_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED');",
                r#"CREATE TABLE public.registration_request (
				id int NOT NULL,
				status public.registration_status DEFAULT 'PENDING'::public.registration_status NOT NULL
			);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE ONLY public.registration_request ADD CONSTRAINT registration_request_pkey PRIMARY KEY (id);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.registration_request (id) VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, status FROM public.registration_request;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("status", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("PENDING")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "qualified type in a column default when the type name is ambiguous",
            set_up_script: &[
                "CREATE SCHEMA schema1;",
                "CREATE SCHEMA schema2;",
                "CREATE TYPE schema1.status AS ENUM ('one');",
                "CREATE TYPE schema2.status AS ENUM ('two');",
                "CREATE TABLE public.t1 (id int NOT NULL, status schema2.status DEFAULT 'two'::schema2.status NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE ONLY public.t1 ADD CONSTRAINT t1_pkey PRIMARY KEY (id);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.t1 (id) VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, status FROM public.t1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("status", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "search path returns user table before public table",
            set_up_script: &[
                "create schema postgres",
                "CREATE TABLE public.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO public.test VALUES (1, 1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE postgres.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO postgres.test VALUES (2, 2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "empty search path will not resolve tables",
            set_up_script: &[
                "CREATE TABLE public.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "set search_path to ''",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO public.test VALUES (1, 1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "public schema qualifier",
            set_up_script: &[
                "CREATE TABLE public.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO public.test VALUES (1, 1), (2, 2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "public schema qualifier, multiple tables",
            set_up_script: &[
                "CREATE TABLE public.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE public.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO public.test VALUES (1, 1), (2, 2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.test2 VALUES (3, 3), (4, 4);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "public db and schema qualifier",
            set_up_script: &[
                "CREATE TABLE postgres.public.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO postgres.public.test VALUES (1, 1), (2, 2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM postgres.public.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create new schema",
            assertions: &[
                ScriptTestAssertion {
                    query: "create schema mySchema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "create schema otherSchema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mySchema.test values (1,1), (2,2)",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into otherSchema.test values (3,3), (4,4)",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mySchema.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM otherSchema.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Create schema authorized role",
            assertions: &[
                ScriptTestAssertion {
                    query: "create schema AUTHORIZATION myUser",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "myuser" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE myUser.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "myuser" does not exist"#, position: 14, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into myUser.test values (1,1), (2,2)",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "myuser.test" does not exist"#, position: 13, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA myuser",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create schema invalid names",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"create schema """#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"zero-length delimited identifier at or near """""#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "create schema dolt_123",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "schema already exists",
            assertions: &[
                ScriptTestAssertion {
                    query: "create schema mySchema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "create schema MYSCHEMA",
                    expected: Expected::Error(Diagnostic { code: "42P06", message: r#"schema "myschema" already exists"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create new schema (diff order)",
            assertions: &[
                ScriptTestAssertion {
                    query: "create schema mySchema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "create schema otherSchema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mySchema.test values (1,1), (2,2)",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mySchema.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into otherSchema.test values (3,3), (4,4)",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM otherSchema.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "insert, update, delete with schema",
            assertions: &[
                ScriptTestAssertion {
                    query: "create schema mySchema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "create schema otherSchema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mySchema.test values (1,1), (2,2)",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into otherSchema.test values (3,3), (4,4)",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "update mySchema.test set v1 = 3 where pk = 1",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "update otherSchema.test set v1 = 4 where pk = 3",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "delete from mySchema.test where pk = 2",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "delete from otherSchema.test where pk = 4",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mySchema.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM otherSchema.test;",
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
            name: "schema does not exist",
            assertions: &[
                ScriptTestAssertion {
                    query: "create schema mySchema",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE mySchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE otherSchema.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "otherschema" does not exist"#, position: 14, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mySchema.test values (1,1), (2,2)",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into otherSchema.test values (3,3), (4,4)",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "otherschema.test" does not exist"#, position: 13, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mySchema.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM otherSchema.test;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "otherschema.test" does not exist"#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create new database and new schema",
            set_up_script: &[
                "CREATE DATABASE db2;",
                "USE db2;",
                "create schema schema2;",
                "use postgres",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE db2.schema2.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO db2.schema2.test VALUES (1, 1), (2, 2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM db2.schema2.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "non default schema qualifier",
            set_up_script: &[
                "CREATE SCHEMA myschema",
                "SET search_path = 'myschema'",
                "CREATE TABLE mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mytbl VALUES (1, 1), (2, 2)",
                "CREATE VIEW myvw AS SELECT pk+3, v1 from mytbl",
                "SET search_path TO DEFAULT",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""$user", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mytbl;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "mytbl" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM myvw;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "myvw" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM myschema.mytbl;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM myschema.myvw;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("4"), T("1")],
                            &[T("5"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "add new table in new schema, commit, status",
            set_up_script: &[
                "CREATE SCHEMA myschema",
                "Create table myschema.mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("myschema.mytbl"), T("f"), T("new table")],
                            &[T("myschema"), T("f"), T("new schema")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_add('.')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("myschema.mytbl"), T("t"), T("new table")],
                            &[T("myschema"), T("t"), T("new schema")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_commit('-m', 'new table in new schema')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select message from dolt.log order by date desc limit 1",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("new table in new schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "add new table in new schema, commit -Am",
            set_up_script: &[
                "CREATE SCHEMA myschema",
                "Create table myschema.mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("myschema.mytbl"), T("f"), T("new table")],
                            &[T("myschema"), T("f"), T("new schema")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_commit('-Am', 'new table in new schema')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select message from dolt.log order by date desc limit 1",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("new table in new schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "merge new table in new schema",
            set_up_script: &[
                "select dolt_checkout('-b', 'branch1')",
                "CREATE SCHEMA branchschema",
                "Create table branchschema.mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO branchschema.mytbl VALUES (1, 1), (2, 2)",
                "Create table branchschema.mytbl2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO branchschema.mytbl2 VALUES (3, 3), (4, 4)",
                "select dolt_commit('-Am', 'new table in new schema')",
                "select dolt_checkout('main')",
                "create schema mainschema",
                "create table mainschema.maintable (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "insert into mainschema.maintable values (5, 5), (6, 6)",
                "select dolt_commit('-Am', 'new table in main')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_merge('branch1')::text)",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("57")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * from mainschema.maintable",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("5"), T("5")],
                            &[T("6"), T("6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * from branchschema.mytbl",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * from branchschema.mytbl2",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[
                            &[T("3"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create new schema with no tables, add and commit",
            set_up_script: &[
                "CREATE SCHEMA myschema",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("myschema"), T("f"), T("new schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_add('.')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("myschema"), T("t"), T("new schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_commit('-m', 'new schema')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select message from dolt.log order by date desc limit 1",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("new schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "USE branches",
            set_up_script: &[
                r#"USE "postgres/main""#,
                "CREATE SCHEMA myschema",
                "SET search_path = 'myschema'",
                "CREATE TABLE mytbl (pk BIGINT PRIMARY KEY, v1 BIGINT);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT active_branch();",
                    expected: Expected::Rows {
                        columns: &[Column("active_branch", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT current_schemas(true);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{pg_catalog,myschema}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT catalog_name, schema_name FROM information_schema.schemata;",
                    expected: Expected::Rows {
                        columns: &[Column("catalog_name", VARCHAR), Column("schema_name", VARCHAR)],
                        rows: &[
                            &[T("postgres"), T("dolt")],
                            &[T("postgres"), T("myschema")],
                            &[T("postgres"), T("pg_catalog")],
                            &[T("postgres"), T("public")],
                            &[T("postgres"), T("information_schema")],
                            &[T("postgres/main"), T("dolt")],
                            &[T("postgres/main"), T("myschema")],
                            &[T("postgres/main"), T("pg_catalog")],
                            &[T("postgres/main"), T("public")],
                            &[T("postgres/main"), T("information_schema")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres/main';",
                    expected: Expected::Rows {
                        columns: &[Column("schema_name", VARCHAR)],
                        rows: &[
                            &[T("dolt")],
                            &[T("myschema")],
                            &[T("pg_catalog")],
                            &[T("public")],
                            &[T("information_schema")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres';",
                    expected: Expected::Rows {
                        columns: &[Column("schema_name", VARCHAR)],
                        rows: &[
                            &[T("dolt")],
                            &[T("myschema")],
                            &[T("pg_catalog")],
                            &[T("public")],
                            &[T("information_schema")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM mytbl;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT schemaname, tablename FROM pg_catalog.pg_tables WHERE schemaname not in ('pg_catalog', 'information_schema', 'dolt') and left(tablename, 5) <> 'dolt_';",
                    expected: Expected::Rows {
                        columns: &[Column("schemaname", NAME), Column("tablename", NAME)],
                        rows: &[
                            &[T("myschema"), T("mytbl")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("myschema.mytbl"), T("f"), T("new table")],
                            &[T("myschema"), T("f"), T("new schema")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-A', '-m', 'Add mytbl');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_branch('newbranch')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE 'postgres/newbranch'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT current_schemas(true);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{pg_catalog,myschema}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newbranchschema;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path TO 'newbranchschema'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT current_schemas(true);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{pg_catalog,newbranchschema}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE mytbl2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres/newbranch';",
                    expected: Expected::Rows {
                        columns: &[Column("schema_name", VARCHAR)],
                        rows: &[
                            &[T("dolt")],
                            &[T("myschema")],
                            &[T("newbranchschema")],
                            &[T("pg_catalog")],
                            &[T("public")],
                            &[T("information_schema")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT schema_name FROM information_schema.schemata WHERE catalog_name = 'postgres';",
                    expected: Expected::Rows {
                        columns: &[Column("schema_name", VARCHAR)],
                        rows: &[
                            &[T("dolt")],
                            &[T("myschema")],
                            &[T("pg_catalog")],
                            &[T("public")],
                            &[T("information_schema")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT schemaname, tablename FROM pg_catalog.pg_tables WHERE schemaname not in ('pg_catalog', 'information_schema', 'dolt') and left(tablename, 5) <> 'dolt_';",
                    expected: Expected::Rows {
                        columns: &[Column("schemaname", NAME), Column("tablename", NAME)],
                        rows: &[
                            &[T("myschema"), T("mytbl")],
                            &[T("newbranchschema"), T("mytbl2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: the schema is created with a quoted mixed-case name, so Postgres needs it quoted to find it.
        ScriptTest {
            name: "drop schema",
            set_up_script: &[
                "CREATE SCHEMA dropme",
                r#"CREATE schema "hasTables""#,
                r#"CREATE TABLE "hasTables".t1 (pk BIGINT PRIMARY KEY, v1 BIGINT);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "Show schemas",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "schemas""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SCHEMA dropme;",
                    expected: Expected::Tag("DROP SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "Show schemas",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "schemas""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SCHEMA dropme;",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "dropme" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "drop schema if exists dropme;",
                    expected: Expected::Tag("DROP SCHEMA"),
                    notices: &[Diagnostic { code: "00000", message: r#"schema "dropme" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SCHEMA hasTables;",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "hastables" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "drop schema hasTables cascade;",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "hastables" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "create schema hastype;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "create type hastype.mytype as enum('a', 'b', 'c');",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SCHEMA hastype;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop schema hastype because other objects depend on it", detail: "type hastype.mytype depends on schema hastype", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "create schema hassequence;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "create sequence hassequence.myseq start 1 increment 1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SCHEMA hassequence;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop schema hassequence because other objects depend on it", detail: "sequence hassequence.myseq depends on schema hassequence", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
