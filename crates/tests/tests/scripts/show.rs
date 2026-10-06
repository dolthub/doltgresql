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
                ScriptTestAssertion {
                    query: "EXPLAIN t1",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "t1""#, position: 9, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DESCRIBE t1",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "DESCRIBE""#, position: 1, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DESC t1",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "DESC""#, position: 1, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DESC public.t1",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "DESC""#, position: 1, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DESC postgres.public.t1",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "DESC""#, position: 1, ..E }),
                    flow: Flow::Query,
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DESC t1",
                    expected: Expected::Rows {
                        columns: &[Column("Field", TEXT), Column("Type", TEXT), Column("Null", TEXT), Column("Key", TEXT), Column("Default", TEXT), Column("Extra", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), T("NO"), T("PRI"), Null, T("")],
                            &[T("name"), T("text"), T("YES"), T(""), Null, T("")],
                            &[T("age"), T("integer"), T("YES"), T(""), Null, T("")],
                            &[T("height"), T("integer"), T("YES"), T(""), Null, T("")],
                        ],
                        tag: "EXPLAIN",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "EXPLAIN public.t1 AS OF 'HEAD'",
                    expected: Expected::Rows {
                        columns: &[Column("Field", TEXT), Column("Type", TEXT), Column("Null", TEXT), Column("Key", TEXT), Column("Default", TEXT), Column("Extra", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), T("NO"), T("PRI"), Null, T("")],
                            &[T("name"), T("text"), T("YES"), T(""), Null, T("")],
                            &[T("age"), T("integer"), T("YES"), T(""), Null, T("")],
                        ],
                        tag: "EXPLAIN",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DESCRIBE postgres.public.t1 AS OF 'HEAD~'",
                    expected: Expected::Rows {
                        columns: &[Column("Field", TEXT), Column("Type", TEXT), Column("Null", TEXT), Column("Key", TEXT), Column("Default", TEXT), Column("Extra", TEXT)],
                        rows: &[
                            &[T("id"), T("integer"), T("NO"), T("PRI"), Null, T("")],
                            &[T("name"), T("text"), T("YES"), T(""), Null, T("")],
                        ],
                        tag: "EXPLAIN",
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
                ScriptTestAssertion {
                    query: "DESC schema2.t2",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "DESC""#, position: 1, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DESC postgres.schema2.t2",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "DESC""#, position: 1, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DESC t2",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "DESC""#, position: 1, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO 'schema2'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DESC t2",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "DESC""#, position: 1, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO schema3, schema2",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DESC t2",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "DESC""#, position: 1, ..E }),
                    flow: Flow::Query,
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
                ScriptTestAssertion {
                    query: "SHOW CREATE TABLE t1",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "CREATE""#, position: 6, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW CREATE TABLE T2",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "CREATE""#, position: 6, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW CREATE TABLE t3",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "CREATE""#, position: 6, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW CREATE TABLE dne",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "CREATE""#, position: 6, ..E }),
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
                ScriptTestAssertion {
                    query: "SHOW databases",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "databases""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "show SCHEMAS",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "schemas""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "show SCHEMAS FROM postgres",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "FROM""#, position: 14, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "show SCHEMAS FROM db1",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "FROM""#, position: 14, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "show SCHEMAS FROM dne",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "FROM""#, position: 14, ..E }),
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
                ScriptTestAssertion {
                    query: "SHOW indexes FROM t1",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "FROM""#, position: 14, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW indexes FROM dne",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "FROM""#, position: 14, ..E }),
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW SEQUENCES",
                    expected: Expected::Rows {
                        columns: &[Column("sequence_schema", TEXT), Column("sequence_name", TEXT)],
                        rows: &[
                            &[T("public"), T("seq2")],
                            &[T("schema1"), T("seq3")],
                            &[T("public"), T("seq1")],
                        ],
                        tag: "SHOW SCHEMAS",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW SEQUENCES from postgres",
                    expected: Expected::Rows {
                        columns: &[Column("sequence_schema", TEXT), Column("sequence_name", TEXT)],
                        rows: &[
                            &[T("public"), Any],
                            &[Any, Any],
                            &[Any, Any],
                        ],
                        tag: "SHOW SCHEMAS",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW SEQUENCES FROM db1",
                    expected: Expected::Rows {
                        columns: &[Column("sequence_schema", TEXT), Column("sequence_name", TEXT)],
                        rows: &[
                            &[T("public"), T("seq4")],
                            &[T("schema2"), T("seq5")],
                        ],
                        tag: "SHOW SCHEMAS",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW SEQUENCES FROM dne",
                    expected: Expected::Error(Diagnostic { code: "3D000", message: "database not found: dne", ..E }),
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
                ScriptTestAssertion {
                    query: "SHOW TABLES",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "tables""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW TABLES from public",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "from""#, position: 13, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW TABLES from schema2",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "from""#, position: 13, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW TABLES from schema3",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "from""#, position: 13, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW TABLES from postgres.public",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "from""#, position: 13, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW TABLES from postgres.schema2",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "from""#, position: 13, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW TABLES from postgres.schema3",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "from""#, position: 13, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW TABLES from db3",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "from""#, position: 13, ..E }),
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW TABLES",
                    expected: Expected::Rows {
                        columns: &[Column("Tables_in_postgres", TEXT)],
                        rows: &[
                            &[T("t1")],
                            &[T("t2")],
                        ],
                        tag: "SHOW TABLES",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW TABLES from public",
                    expected: Expected::Rows {
                        columns: &[Column("Tables_in_postgres", TEXT)],
                        rows: &[
                            &[T("t1")],
                            &[T("t2")],
                        ],
                        tag: "SHOW TABLES",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW TABLES from schema2",
                    expected: Expected::Rows {
                        columns: &[Column("Tables_in_postgres", TEXT)],
                        rows: &[
                            &[T("t3")],
                            &[T("t4")],
                        ],
                        tag: "SHOW TABLES",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW TABLES from postgres.public",
                    expected: Expected::Rows {
                        columns: &[Column("Tables_in_postgres", TEXT)],
                        rows: &[
                            &[T("t1")],
                            &[T("t2")],
                        ],
                        tag: "SHOW TABLES",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW TABLES from postgres.schema2",
                    expected: Expected::Rows {
                        columns: &[Column("Tables_in_postgres", TEXT)],
                        rows: &[
                            &[T("t3")],
                            &[T("t4")],
                        ],
                        tag: "SHOW TABLES",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW TABLES from db2",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: "database schema not found: db2", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW TABLES from db2.public",
                    expected: Expected::Rows {
                        columns: &[Column("Tables_in_db2", TEXT)],
                        rows: &[
                            &[T("t5")],
                        ],
                        tag: "SHOW TABLES",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW TABLES from db2.schema3",
                    expected: Expected::Rows {
                        columns: &[Column("Tables_in_db2", TEXT)],
                        rows: &[
                            &[T("t6")],
                        ],
                        tag: "SHOW TABLES",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path TO 'schema2'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW TABLES",
                    expected: Expected::Rows {
                        columns: &[Column("Tables_in_postgres", TEXT)],
                        rows: &[
                            &[T("t3")],
                            &[T("t4")],
                        ],
                        tag: "SHOW TABLES",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path TO 'schema3'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SHOW TABLES",
                    expected: Expected::Rows {
                        columns: &[Column("Tables_in_postgres", TEXT)],
                        rows: &[],
                        tag: "SHOW TABLES",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
