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
fn test_describe() {
    run_scripts(&[
        ScriptTest {
            name: "describe table",
            set_up_script: &[
                "CREATE TABLE t1 (id INT PRIMARY KEY, name TEXT)",
            ],
            assertions: &[
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "EXPLAIN t1",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), Null, T("not null"), Null],
                            &[T("name"), T("text"), Null, T(""), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESCRIBE t1",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), Null, T("not null"), Null],
                            &[T("name"), T("text"), Null, T(""), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESC t1",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), Null, T("not null"), Null],
                            &[T("name"), T("text"), Null, T(""), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESC public.t1",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), Null, T("not null"), Null],
                            &[T("name"), T("text"), Null, T(""), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESC postgres.public.t1",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), Null, T("not null"), Null],
                            &[T("name"), T("text"), Null, T(""), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "describe table AS OF",
            set_up_script: &[
                "CREATE TABLE t1 (id INT PRIMARY KEY, name TEXT)",
                "select dolt_commit('-Am', 'first commit')",
                "ALTER TABLE t1 ADD COLUMN age INT",
                "select dolt_commit('-am', 'second commit')",
                "ALTER TABLE t1 ADD COLUMN height INT",
            ],
            assertions: &[
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESC t1",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), Null, T("not null"), Null],
                            &[T("name"), T("text"), Null, T(""), Null],
                            &[T("age"), T("integer"), Null, T(""), Null],
                            &[T("height"), T("integer"), Null, T(""), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "EXPLAIN public.t1 AS OF 'HEAD'",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), Null, T("not null"), Null],
                            &[T("name"), T("text"), Null, T(""), Null],
                            &[T("age"), T("integer"), Null, T(""), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESCRIBE postgres.public.t1 AS OF 'HEAD~'",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), Null, T("not null"), Null],
                            &[T("name"), T("text"), Null, T(""), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "describe table in other schema",
            set_up_script: &[
                "CREATE TABLE t1 (a INT PRIMARY KEY, b TEXT)",
                "create schema schema2",
                "CREATE TABLE schema2.t2 (c INT PRIMARY KEY, d TEXT)",
                "create schema schema3",
                "CREATE TABLE schema3.t2 (e INT PRIMARY KEY, f TEXT)",
            ],
            assertions: &[
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESC schema2.t2",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("c"), T("integer"), Null, T("not null"), Null],
                            &[T("d"), T("text"), Null, T(""), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESC postgres.schema2.t2",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("c"), T("integer"), Null, T("not null"), Null],
                            &[T("d"), T("text"), Null, T(""), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESC t2",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "t2" does not exist"#, position: 6, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO 'schema2'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESC t2",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("c"), T("integer"), Null, T("not null"), Null],
                            &[T("d"), T("text"), Null, T(""), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO schema3, schema2",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "DESC t2",
                    expected: Expected::Rows {
                        columns: &[Column("Column", TEXT), Column("Type", TEXT), Column("Collation", TEXT), Column("Nullable", TEXT), Column("Default", TEXT)],
                        rows: &[
                            &[T("e"), T("integer"), Null, T("not null"), Null],
                            &[T("f"), T("text"), Null, T(""), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_show_create_table() {
    run_scripts(&[
        ScriptTest {
            name: "show create table",
            set_up_script: &[
                "CREATE TABLE t1 (a INT PRIMARY KEY, name TEXT)",
                "CREATE TABle t2 (b SERIAL PRIMARY KEY, time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)",
                "CREATE TABLE t3 (a timestamp PRIMARY KEY, name varchar(100))",
            ],
            assertions: &[
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows the CREATE TABLE that DOLT_PATCH writes.
                ScriptTestAssertion {
                    query: "SHOW CREATE TABLE t1",
                    expected: Expected::Rows {
                        columns: &[Column("Table", TEXT), Column("Create Table", TEXT)],
                        rows: &[
                            &[T("t1"), T(r#"CREATE TABLE "t1" (
  "a" integer NOT NULL,
  "name" text,
  PRIMARY KEY ("a")
)"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows the CREATE TABLE that DOLT_PATCH writes.
                ScriptTestAssertion {
                    query: "SHOW CREATE TABLE T2",
                    expected: Expected::Rows {
                        columns: &[Column("Table", TEXT), Column("Create Table", TEXT)],
                        rows: &[
                            &[T("t2"), T(r#"CREATE TABLE "t2" (
  "b" integer NOT NULL DEFAULT (nextval('public.t2_b_seq')),
  "time" timestamp NOT NULL DEFAULT (current_timestamp),
  PRIMARY KEY ("b")
)"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows the CREATE TABLE that DOLT_PATCH writes.
                ScriptTestAssertion {
                    query: "SHOW CREATE TABLE t3",
                    expected: Expected::Rows {
                        columns: &[Column("Table", TEXT), Column("Create Table", TEXT)],
                        rows: &[
                            &[T("t3"), T(r#"CREATE TABLE "t3" (
  "a" timestamp NOT NULL,
  "name" varchar(100),
  PRIMARY KEY ("a")
)"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows the CREATE TABLE that DOLT_PATCH writes.
                ScriptTestAssertion {
                    query: "SHOW CREATE TABLE dne",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dne" does not exist"#, position: 19, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_show_databases_and_schemas() {
    run_scripts(&[
        ScriptTest {
            name: "show databases",
            set_up_script: &[
                "CREATE DATABASE db1",
                "CREATE DATABASE db2",
                "CREATE SCHEMA schema1",
                "CREATE SCHEMA schema2",
            ],
            assertions: &[
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW databases",
                    expected: Expected::Rows {
                        columns: &[Column("Name", NAME), Column("Owner", NAME), Column("Encoding", NAME), Column("Collate", TEXT), Column("Ctype", TEXT), Column("ICU Locale", TEXT), Column("Locale Provider", TEXT), Column("Access privileges", TEXT)],
                        rows: &[
                            &[T("db1"), T("postgres"), T("UTF8"), T("C"), T("C"), Null, T("libc"), Null],
                            &[T("db2"), T("postgres"), T("UTF8"), T("C"), T("C"), Null, T("libc"), Null],
                            &[T("postgres"), T("postgres"), T("UTF8"), T("C"), T("C"), Null, T("libc"), Null],
                            &[T("template0"), T("postgres"), T("UTF8"), T("C"), T("C"), Null, T("libc"), Null],
                            &[T("template1"), T("postgres"), T("UTF8"), T("C"), T("C"), Null, T("libc"), Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "show SCHEMAS",
                    expected: Expected::Rows {
                        columns: &[Column("Name", NAME), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("pg_database_owner")],
                            &[T("schema1"), T("postgres")],
                            &[T("schema2"), T("postgres")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "show SCHEMAS FROM postgres",
                    expected: Expected::Rows {
                        columns: &[Column("Name", NAME), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("pg_database_owner")],
                            &[T("schema1"), T("postgres")],
                            &[T("schema2"), T("postgres")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "show SCHEMAS FROM db1",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cross-database references are not implemented: "db1""#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "show SCHEMAS FROM dne",
                    expected: Expected::Error(Diagnostic { code: "3D000", message: r#"database "dne" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_show_indexes() {
    run_scripts(&[
        ScriptTest {
            name: "show indexes",
            set_up_script: &[
                "CREATE TABLE t1 (a INT PRIMARY KEY, name TEXT, value int)",
                "CREATE INDEX idx_name ON t1(name)",
                "CREATE INDEX idx_value ON t1(value)",
            ],
            assertions: &[
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW indexes FROM t1",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME), Column("Table", NAME)],
                        rows: &[
                            &[T("public"), T("idx_name"), T("index"), T("postgres"), T("t1")],
                            &[T("public"), T("idx_value"), T("index"), T("postgres"), T("t1")],
                            &[T("public"), T("t1_pkey"), T("index"), T("postgres"), T("t1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW indexes FROM dne",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dne" does not exist"#, position: 392, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_show_sequences() {
    run_scripts(&[
        ScriptTest {
            name: "show sequences",
            set_up_script: &[
                "CREATE SEQUENCE seq1",
                "CREATE SEQUENCE seq2",
                "CREATE SCHEMA schema1",
                "CREATE SEQUENCE schema1.seq3",
                "CREATE DATABASE db1",
                "USE db1",
                "CREATE SEQUENCE seq4",
                "CREATE SCHEMA schema2",
                "CREATE SEQUENCE schema2.seq5",
                "use postgres",
            ],
            assertions: &[
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW SEQUENCES",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("seq1"), T("sequence"), T("postgres")],
                            &[T("public"), T("seq2"), T("sequence"), T("postgres")],
                            &[T("schema1"), T("seq3"), T("sequence"), T("postgres")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW SEQUENCES from postgres",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("seq1"), T("sequence"), T("postgres")],
                            &[T("public"), T("seq2"), T("sequence"), T("postgres")],
                            &[T("schema1"), T("seq3"), T("sequence"), T("postgres")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW SEQUENCES FROM db1",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cross-database references are not implemented: "db1""#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW SEQUENCES FROM dne",
                    expected: Expected::Error(Diagnostic { code: "3D000", message: r#"database "dne" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_show_tables() {
    run_scripts(&[
        ScriptTest {
            name: "show tables in single schema",
            set_up_script: &[
                "CREATE SEQUENCE seq1;",
                "CREATE TABLE t1 (a INT PRIMARY KEY, name TEXT)",
                "CREATE TABLE t2 (b INT PRIMARY KEY, name TEXT)",
                "create schema schema2",
                "create database db2",
            ],
            assertions: &[
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("t1"), T("table"), T("postgres")],
                            &[T("public"), T("t2"), T("table"), T("postgres")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from public",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("t1"), T("table"), T("postgres")],
                            &[T("public"), T("t2"), T("table"), T("postgres")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from schema2",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from schema3",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "schema3" does not exist"#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from postgres.public",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("t1"), T("table"), T("postgres")],
                            &[T("public"), T("t2"), T("table"), T("postgres")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from postgres.schema2",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from postgres.schema3",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "schema3" does not exist"#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from db3",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "db3" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "show tables in multiple schemas, dbs",
            set_up_script: &[
                "CREATE TABLE t1 (a INT PRIMARY KEY, name TEXT)",
                "CREATE TABLE t2 (b INT PRIMARY KEY, name TEXT)",
                "create schema schema2",
                "CREATE TABLE schema2.t3 (a INT PRIMARY KEY, name TEXT)",
                "CREATE TABLE schema2.t4 (b INT PRIMARY KEY, name TEXT)",
                "create database db2",
                "use db2",
                "CREATE TABLE t5 (a INT PRIMARY KEY, name TEXT)",
                "create schema schema3",
                "CREATE TABLE schema3.t6 (b INT PRIMARY KEY, name TEXT)",
                "use postgres",
            ],
            assertions: &[
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("t1"), T("table"), T("postgres")],
                            &[T("public"), T("t2"), T("table"), T("postgres")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from public",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("t1"), T("table"), T("postgres")],
                            &[T("public"), T("t2"), T("table"), T("postgres")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from schema2",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("schema2"), T("t3"), T("table"), T("postgres")],
                            &[T("schema2"), T("t4"), T("table"), T("postgres")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from postgres.public",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("public"), T("t1"), T("table"), T("postgres")],
                            &[T("public"), T("t2"), T("table"), T("postgres")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from postgres.schema2",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("schema2"), T("t3"), T("table"), T("postgres")],
                            &[T("schema2"), T("t4"), T("table"), T("postgres")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from db2",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "db2" does not exist"#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from db2.public",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cross-database references are not implemented: "db2.public""#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES from db2.schema3",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cross-database references are not implemented: "db2.schema3""#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path TO 'schema2'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
                        rows: &[
                            &[T("schema2"), T("t3"), T("table"), T("postgres")],
                            &[T("schema2"), T("t4"), T("table"), T("postgres")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path TO 'schema3'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres lacks this statement, so this expectation follows psql's describe commands.
                ScriptTestAssertion {
                    query: "SHOW TABLES",
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", NAME)],
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
