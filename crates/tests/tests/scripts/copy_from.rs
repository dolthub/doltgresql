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
fn test_copy() {
    run_scripts(&[
        ScriptTest {
            name: "tab delimited with header",
            set_up_script: &[
                "CREATE TABLE test (pk int primary key);",
                "INSERT INTO test VALUES (0), (1);",
                "CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY test_info FROM STDIN WITH (HEADER);",
                    expected: Expected::Tag("COPY 3"),
                    copy_from_stdin_file: "tab-load-with-header.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_info order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("4"), T("string for 4"), T("1")],
                            &[T("5"), T("string for 5"), T("0")],
                            &[T("6"), T("string for 6"), T("0")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tab delimited with header and column names",
            set_up_script: &[
                "CREATE TABLE test (pk int primary key);",
                "INSERT INTO test VALUES (0), (1);",
                "CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY test_info (id, info, test_pk) FROM STDIN WITH (HEADER);",
                    expected: Expected::Tag("COPY 3"),
                    copy_from_stdin_file: "tab-load-with-header.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_info order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("4"), T("string for 4"), T("1")],
                            &[T("5"), T("string for 5"), T("0")],
                            &[T("6"), T("string for 6"), T("0")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tab delimited with quoted column names",
            set_up_script: &[
                r#"CREATE TABLE Regions (
   "Id" SERIAL UNIQUE NOT NULL,
   "Code" VARCHAR(4) UNIQUE NOT NULL,
   "Capital" VARCHAR(10) NOT NULL,
   "Name" VARCHAR(255) UNIQUE NOT NULL
);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"COPY regions ("Id", "Code", "Capital", "Name") FROM stdin;
"#,
                    expected: Expected::Tag("COPY 3"),
                    copy_from_stdin_file: "tab-load-with-quoted-column-names.sql",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "timestamp columns",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk timestamp primary key, ts timestamp);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 FROM STDIN WITH (HEADER)",
                    expected: Expected::Tag("COPY 3"),
                    copy_from_stdin_file: "tab-load-with-timestamp-col.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", TIMESTAMP), Column("ts", TIMESTAMP)],
                        rows: &[
                            &[T("2020-12-19 19:00:00"), T("2021-04-04 20:00:00")],
                            &[T("2020-12-19 21:36:32.188"), T("2020-12-19 19:00:00")],
                            &[T("2021-04-04 20:00:00"), T("2020-12-19 21:36:32.188")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "basic csv",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 FROM STDIN (FORMAT CSV)",
                    expected: Expected::Tag("COPY 9"),
                    copy_from_stdin_file: "csv-load-basic-cases.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 where pk = 6 order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", VARCHAR), Column("c2", VARCHAR)],
                        rows: &[
                            &[T("6"), T(r#"foo
\\.
bar"#), T("baz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 where pk = 9;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", VARCHAR), Column("c2", VARCHAR)],
                        rows: &[
                            &[T("9"), Null, T("''")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "csv with header",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: " COPY tbl1 FROM STDIN (FORMAT CSV, HEADER TRUE);",
                    expected: Expected::Tag("COPY 9"),
                    copy_from_stdin_file: "csv-load-with-header.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 where pk = 6 order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", VARCHAR), Column("c2", VARCHAR)],
                        rows: &[
                            &[T("6"), T(r#"foo
\\.
bar"#), T("baz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated column",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250), c3 int generated always as (pk + 10) stored);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1, c2) FROM STDIN (FORMAT CSV)",
                    expected: Expected::Tag("COPY 9"),
                    copy_from_stdin_file: "csv-load-basic-cases.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 where pk = 6 order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", VARCHAR), Column("c2", VARCHAR), Column("c3", INT4)],
                        rows: &[
                            &[T("6"), T(r#"foo
\\.
bar"#), T("baz"), T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 where pk = 9;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", VARCHAR), Column("c2", VARCHAR), Column("c3", INT4)],
                        rows: &[
                            &[T("9"), Null, T("''"), T("19")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "load multiple chunks",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 FROM STDIN (FORMAT CSV);",
                    expected: Expected::Tag("COPY 100"),
                    copy_from_stdin_file: "csv-load-multi-chunk.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 where pk = 99 order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", VARCHAR), Column("c2", VARCHAR)],
                        rows: &[
                            &[T("99"), T("foo"), T("barbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbash")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "load psv with headers",
            set_up_script: &[
                "CREATE TABLE test (pk int primary key);",
                "INSERT INTO test VALUES (0), (1);",
                "CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY test_info FROM STDIN (FORMAT CSV, HEADER TRUE, DELIMITER '|');",
                    expected: Expected::Tag("COPY 3"),
                    copy_from_stdin_file: "psv-load.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_info order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("4"), T("string for 4"), T("1")],
                            &[T("5"), T("string for 5"), T("0")],
                            &[T("6"), T("string for 6"), T("0")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "csv from file",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 FROM '{TESTDATA}/csv-load-basic-cases.sql' (FORMAT CSV)",
                    expected: Expected::Tag("COPY 9"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 where pk = 6 order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", VARCHAR), Column("c2", VARCHAR)],
                        rows: &[
                            &[T("6"), T(r#"foo
\\.
bar"#), T("baz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 where pk = 9;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", VARCHAR), Column("c2", VARCHAR)],
                        rows: &[
                            &[T("9"), Null, T("''")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "csv from file with column names",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1, c2) FROM '{TESTDATA}/csv-load-basic-cases.sql' (FORMAT CSV)",
                    expected: Expected::Tag("COPY 9"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 where pk = 6 order by pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", VARCHAR), Column("c2", VARCHAR)],
                        rows: &[
                            &[T("6"), T(r#"foo
\\.
bar"#), T("baz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from tbl1 where pk = 9;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", VARCHAR), Column("c2", VARCHAR)],
                        rows: &[
                            &[T("9"), Null, T("''")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tab delimited with header from file",
            set_up_script: &[
                "CREATE TABLE test (pk int primary key);",
                "INSERT INTO test VALUES (0), (1);",
                "CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY test_info FROM '{TESTDATA}/tab-load-with-header.sql' WITH (HEADER)",
                    expected: Expected::Tag("COPY 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_info order by 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("4"), T("string for 4"), T("1")],
                            &[T("5"), T("string for 5"), T("0")],
                            &[T("6"), T("string for 6"), T("0")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tab delimited with uuid values",
            set_up_script: &[
                r#"CREATE TABLE public.uuid_table (
    id uuid NOT NULL,
    name character varying NOT NULL,
    second_uuid uuid DEFAULT '428d0815-d95b-4cfc-89af-9fca38585dcc'::uuid);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY uuid_table (id, name, second_uuid) FROM STDIN",
                    expected: Expected::Tag("COPY 2"),
                    copy_from_stdin_file: "uuid-table.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM uuid_table order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", UUID), Column("name", VARCHAR), Column("second_uuid", UUID)],
                        rows: &[
                            &[T("1077f506-a6fc-4cb2-aed2-9dea9351ed9c"), T("Company A"), T("428d0815-d95b-4cfc-89af-9fca38585dcc")],
                            &[T("5e080b3a-361f-4e16-b7a4-70d4f175e283"), T("Company B"), T("428d0815-d95b-4cfc-89af-9fca38585dcc")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "binary from stdin",
            set_up_script: &[
                "CREATE TABLE tbl3 (pk int primary key, c1 text, b boolean);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl3 FROM STDIN (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 4"),
                    copy_from_stdin_file: "copy-to-basic.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tbl3 ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", TEXT), Column("b", BOOL)],
                        rows: &[
                            &[T("1"), T("foo"), T("t")],
                            &[T("2"), Null, T("f")],
                            &[T("3"), T(""), Null],
                            &[T("4"), T("héllo"), T("t")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "binary load multiple chunks",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 text);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 FROM STDIN (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 2000"),
                    copy_from_stdin_file: "binary-load-multi-chunk.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), count(c1), sum(length(c1)) FROM tbl1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("2000"), T("1980"), T("202536")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk, length(c1) FROM tbl1 WHERE pk IN (99, 211, 300, 1999) ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("length", INT4)],
                        rows: &[
                            &[T("99"), T("99")],
                            &[T("211"), T("0")],
                            &[T("300"), Null],
                            &[T("1999"), T("100")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "binary from file",
            set_up_script: &[
                "CREATE TABLE tbl3 (pk int primary key, c1 text, b boolean);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl3 FROM '{TESTDATA}/copy-to-basic.bin' (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 4"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tbl3 ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", TEXT), Column("b", BOOL)],
                        rows: &[
                            &[T("1"), T("foo"), T("t")],
                            &[T("2"), Null, T("f")],
                            &[T("3"), T(""), Null],
                            &[T("4"), T("héllo"), T("t")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "malformed binary load does not poison the session",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 text);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 FROM STDIN (FORMAT BINARY);",
                    expected: Expected::Error(Diagnostic { code: "22P04", message: "row field count is 5, expected 2", ..E }),
                    copy_from_stdin_file: "binary-load-malformed.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM tbl1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tbl1 VALUES (100, 'still works');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 FROM STDIN (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 3"),
                    copy_from_stdin_file: "binary-load-2col.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tbl1 ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", TEXT)],
                        rows: &[
                            &[T("1"), T("one")],
                            &[T("2"), Null],
                            &[T("3"), T("three")],
                            &[T("100"), T("still works")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "binary load failing after a successful chunk does not poison the session",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 text);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 FROM STDIN (FORMAT BINARY);",
                    expected: Expected::Error(Diagnostic { code: "22P04", message: "row field count is 5, expected 2", ..E }),
                    copy_from_stdin_file: "binary-load-malformed-late.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM tbl1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 FROM STDIN (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 3"),
                    copy_from_stdin_file: "binary-load-2col.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM tbl1;",
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
            name: "binary load missing its trailer does not poison the session",
            set_up_script: &[
                "CREATE TABLE tbl3 (pk int primary key, c1 text, b boolean);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl3 FROM STDIN (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 4"),
                    flow: Flow::Exec,
                    copy_from_stdin_file: "copy-from-missing-trailer.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM tbl3;",
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
                    query: "COPY tbl3 FROM STDIN (FORMAT BINARY);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "tbl3_pkey""#, detail: "Key (pk)=(1) already exists.", schema: "public", table: "tbl3", constraint: "tbl3_pkey", ..E }),
                    flow: Flow::Query,
                    copy_from_stdin_file: "copy-to-basic.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tbl3 ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", TEXT), Column("b", BOOL)],
                        rows: &[
                            &[T("1"), T("foo"), T("t")],
                            &[T("2"), Null, T("f")],
                            &[T("3"), T(""), Null],
                            &[T("4"), T("héllo"), T("t")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "binary errors",
            set_up_script: &[
                "CREATE TABLE tbl3 (pk int primary key, c1 text, b boolean);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl3 FROM STDIN (FORMAT BINARY, HEADER);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "cannot specify HEADER in BINARY mode", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl3 FROM STDIN (FORMAT BINARY, DELIMITER '|');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "cannot specify DELIMITER in BINARY mode", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "COPY tbl3 FROM '{TESTDATA}/copy-to-basic.txt' (FORMAT BINARY);",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "COPY file signature not recognized", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl3 FROM '{TESTDATA}/copy-from-missing-trailer.bin' (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 4"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "file not found",
            set_up_script: &[
                "CREATE TABLE test (pk int primary key);",
                "INSERT INTO test VALUES (0), (1);",
                "CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "COPY test_info FROM '{TESTDATA}/file-not-found.sql' WITH (HEADER)",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "open {TESTDATA}/file-not-found.sql: no such file or directory", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "wrong columns",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1) FROM '{TESTDATA}/csv-load-basic-cases.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "extra data after last expected column", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1, c3) FROM '{TESTDATA}/csv-load-basic-cases.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "42703", message: "Unknown column 'c3' in 'tbl1'", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "table not found",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "COPY tbl2 (pk, c1) FROM '{TESTDATA}/csv-load-basic-cases.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: tbl2", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "read only table",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "COPY dolt_log FROM '{TESTDATA}/csv-load-basic-cases.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "table doesn't support INSERT INTO", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "bad data rows",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1, c2) FROM '{TESTDATA}/missing-columns.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "record on line 2: wrong number of fields", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select count(*) from tbl1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1, c2) FROM '{TESTDATA}/too-many-columns.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "record on line 6: wrong number of fields", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select count(*) from tbl1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1, c2) FROM '{TESTDATA}/wrong-types.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type int4: "abc""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select count(*) from tbl1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
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
