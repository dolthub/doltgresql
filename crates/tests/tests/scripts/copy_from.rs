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
                ScriptTestAssertion {
                    query: "COPY tbl3 FROM '{TESTDATA}/copy-to-basic.txt' (FORMAT BINARY);",
                    expected: Expected::Error(Diagnostic { code: "22P04", message: "COPY file signature not recognized", ..E }),
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
                ScriptTestAssertion {
                    query: "COPY test_info FROM '{TESTDATA}/file-not-found.sql' WITH (HEADER)",
                    expected: Expected::Error(Diagnostic { code: "58P01", message: r#"could not open file "{TESTDATA}/file-not-found.sql" for reading: No such file or directory"#, hint: r#"COPY FROM instructs the PostgreSQL server process to read a file. You may want a client-side facility such as psql's \copy."#, ..E }),
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
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1) FROM '{TESTDATA}/csv-load-basic-cases.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "22P04", message: "extra data after last expected column", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1, c3) FROM '{TESTDATA}/csv-load-basic-cases.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "c3" of relation "tbl1" does not exist"#, ..E }),
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
                ScriptTestAssertion {
                    query: "COPY tbl2 (pk, c1) FROM '{TESTDATA}/csv-load-basic-cases.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "tbl2" does not exist"#, ..E }),
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
                    skip: Some("Dolt system tables do not resolve as the targets of writes yet"),
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
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1, c2) FROM '{TESTDATA}/missing-columns.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "22P04", message: r#"missing data for column "c2""#, ..E }),
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
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1, c2) FROM '{TESTDATA}/too-many-columns.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "22P04", message: "extra data after last expected column", ..E }),
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
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c1, c2) FROM '{TESTDATA}/wrong-types.sql' (FORMAT CSV)",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "abc""#, ..E }),
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

#[test]
fn test_copy_rules() {
    run_scripts(&[
        ScriptTest {
            name: "copy option errors",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t (pk INT PRIMARY KEY, c1 TEXT, c2 TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (FORMAT CSV, QUOTE 'ab');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "COPY quote must be a single one-byte character", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (QUOTE 'a');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "COPY quote available only in CSV mode", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (ESCAPE 'a');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "COPY escape available only in CSV mode", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (bogus 1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"option "bogus" not recognized"#, position: 19, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"COPY t TO STDOUT (FORMAT CSV, DELIMITER '"');"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: "COPY delimiter and quote must be different", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (DELIMITER 'ab');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "COPY delimiter must be a single one-byte character", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (DELIMITER 'a');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"COPY delimiter cannot be "a""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (FORMAT BINARY, NULL 'x');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "cannot specify NULL in BINARY mode", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (FORMAT CSV, HEADER 'match');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cannot use "match" with HEADER in COPY TO"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (FORMAT CSV, HEADER 'maybe');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"header requires a Boolean value or "match""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (FORCE_QUOTE *);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "COPY force quote available only in CSV mode", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t FROM STDIN (FORMAT CSV, FORCE_QUOTE (c1));",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "COPY force quote only available using COPY TO", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t FROM STDIN (FORCE_NOT_NULL (c1));",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "COPY force not null available only in CSV mode", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (FORMAT CSV, FORCE_NOT_NULL (c1));",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "COPY force not null only available using COPY FROM", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (FORMAT CSV, FORCE_NULL (c1));",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "COPY force null only available using COPY FROM", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (FORMAT CSV, NULL 'a,b');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "COPY delimiter must not appear in the NULL specification", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"COPY t TO STDOUT (FORMAT CSV, NULL 'a"b');"#,
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "CSV quote character must not appear in the NULL specification", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t TO STDOUT (FORMAT CSV, FORMAT TEXT);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "conflicting or redundant options", position: 31, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t (pk, pk) TO STDOUT;",
                    expected: Expected::Error(Diagnostic { code: "42701", message: r#"column "pk" specified more than once"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY t (pk, c3) TO STDOUT;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "c3" of relation "t" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY missing TO STDOUT;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "missing" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "copy round trips through files",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE src (pk INT PRIMARY KEY, c1 TEXT, c2 TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO src VALUES (1, 'a,b', NULL), (2, '', 'x"y'), (3, E'tab\there\\back', E'nl\nx'), (4, '\.', 'q|r');"#,
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"COPY src TO '{TEMPDIR}/doltgres-kept-copy.csv' (FORMAT CSV, HEADER, FORCE_QUOTE (c2), ESCAPE '\');"#,
                    expected: Expected::Tag("COPY 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE dst1 (pk INT PRIMARY KEY, c1 TEXT, c2 TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"COPY dst1 FROM '{TEMPDIR}/doltgres-kept-copy.csv' (FORMAT CSV, HEADER, ESCAPE '\');"#,
                    expected: Expected::Tag("COPY 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM dst1 ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", TEXT), Column("c2", TEXT)],
                        rows: &[
                            &[T("1"), T("a,b"), Null],
                            &[T("2"), T(""), T(r#"x"y"#)],
                            &[T("3"), T(r#"tab	here\back"#), T(r#"nl
x"#)],
                            &[T("4"), T(r#"\."#), T("q|r")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY src TO '{TEMPDIR}/doltgres-kept-copy.txt' (DELIMITER '|', NULL 'NULL');",
                    expected: Expected::Tag("COPY 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE dst2 (pk INT PRIMARY KEY, c1 TEXT, c2 TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY dst2 FROM '{TEMPDIR}/doltgres-kept-copy.txt' (DELIMITER '|', NULL 'NULL');",
                    expected: Expected::Tag("COPY 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM dst2 ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", TEXT), Column("c2", TEXT)],
                        rows: &[
                            &[T("1"), T("a,b"), Null],
                            &[T("2"), T(""), T(r#"x"y"#)],
                            &[T("3"), T(r#"tab	here\back"#), T(r#"nl
x"#)],
                            &[T("4"), T(r#"\."#), T("q|r")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY (SELECT pk, c1 FROM src WHERE pk > 1 ORDER BY pk) TO '{TEMPDIR}/doltgres-kept-copy.bin' (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE dst3 (pk INT, c1 TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY dst3 FROM '{TEMPDIR}/doltgres-kept-copy.bin' (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM dst3 ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", TEXT)],
                        rows: &[
                            &[T("2"), T("")],
                            &[T("3"), T(r#"tab	here\back"#)],
                            &[T("4"), T(r#"\."#)],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY dst3 FROM '{TEMPDIR}/doltgres-kept-copy.csv' (FORMAT BINARY);",
                    expected: Expected::Error(Diagnostic { code: "22P04", message: "COPY file signature not recognized", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY dst1 FROM '{TEMPDIR}/doltgres-kept-copy.bin' (FORMAT BINARY);",
                    expected: Expected::Error(Diagnostic { code: "22P04", message: "row field count is 2, expected 3", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "copy null handling in CSV",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE n (pk INT PRIMARY KEY, a TEXT, b TEXT, c TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY (SELECT 1, '', NULL::TEXT, 'x' UNION ALL SELECT 2, NULL, '', 'y') TO '{TEMPDIR}/doltgres-kept-nulls.csv' (FORMAT CSV);",
                    expected: Expected::Tag("COPY 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY n FROM '{TEMPDIR}/doltgres-kept-nulls.csv' (FORMAT CSV);",
                    expected: Expected::Tag("COPY 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk, a IS NULL, a, b IS NULL, b, c FROM n ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("?column?", BOOL), Column("a", TEXT), Column("?column?", BOOL), Column("b", TEXT), Column("c", TEXT)],
                        rows: &[
                            &[T("1"), T("f"), T(""), T("t"), Null, T("x")],
                            &[T("2"), T("t"), Null, T("f"), T(""), T("y")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM n;",
                    expected: Expected::Tag("DELETE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY n FROM '{TEMPDIR}/doltgres-kept-nulls.csv' (FORMAT CSV, FORCE_NOT_NULL (b), FORCE_NULL (a));",
                    expected: Expected::Tag("COPY 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk, a IS NULL, a, b IS NULL, b, c FROM n ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("?column?", BOOL), Column("a", TEXT), Column("?column?", BOOL), Column("b", TEXT), Column("c", TEXT)],
                        rows: &[
                            &[T("1"), T("t"), Null, T("f"), T(""), T("x")],
                            &[T("2"), T("t"), Null, T("f"), T(""), T("y")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "copy of generated columns and views",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE g (a INT PRIMARY KEY, b INT GENERATED ALWAYS AS (a * 2) STORED);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO g (a) VALUES (1), (2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY g (b) TO STDOUT;",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: r#"column "b" is a generated column"#, detail: "Generated columns cannot be used in COPY.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY g TO '{TEMPDIR}/doltgres-kept-generated.txt';",
                    expected: Expected::Tag("COPY 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM g;",
                    expected: Expected::Tag("DELETE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY g FROM '{TEMPDIR}/doltgres-kept-generated.txt';",
                    expected: Expected::Tag("COPY 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM g ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v AS SELECT 1 AS x;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY v TO STDOUT;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#"cannot copy from view "v""#, hint: "Try the COPY (SELECT ...) TO variant.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY v FROM STDIN;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#"cannot copy to view "v""#, hint: "To enable copying to a view, provide an INSTEAD OF INSERT trigger.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "copy errors in the data",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE e (pk INT PRIMARY KEY, c TEXT NOT NULL);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY (SELECT 'abc', 'x') TO '{TEMPDIR}/doltgres-kept-bad.txt';",
                    expected: Expected::Tag("COPY 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY e FROM '{TEMPDIR}/doltgres-kept-bad.txt';",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "abc""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY (SELECT 1, 'x', 'y') TO '{TEMPDIR}/doltgres-kept-bad.txt';",
                    expected: Expected::Tag("COPY 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY e FROM '{TEMPDIR}/doltgres-kept-bad.txt';",
                    expected: Expected::Error(Diagnostic { code: "22P04", message: "extra data after last expected column", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY (SELECT 1) TO '{TEMPDIR}/doltgres-kept-bad.txt';",
                    expected: Expected::Tag("COPY 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY e FROM '{TEMPDIR}/doltgres-kept-bad.txt';",
                    expected: Expected::Error(Diagnostic { code: "22P04", message: r#"missing data for column "c""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY (SELECT 1, NULL) TO '{TEMPDIR}/doltgres-kept-bad.txt';",
                    expected: Expected::Tag("COPY 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY e FROM '{TEMPDIR}/doltgres-kept-bad.txt';",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"null value in column "c" of relation "e" violates not-null constraint"#, detail: "Failing row contains (1, null).", schema: "public", table: "e", column: "c", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY (SELECT 1, 'x' UNION ALL SELECT 1, 'y') TO '{TEMPDIR}/doltgres-kept-bad.txt';",
                    expected: Expected::Tag("COPY 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY e FROM '{TEMPDIR}/doltgres-kept-bad.txt';",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "e_pkey""#, detail: "Key (pk)=(1) already exists.", schema: "public", table: "e", constraint: "e_pkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM e;",
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
                    query: "COPY e FROM '/tmp/doltgres-kept-missing.txt';",
                    expected: Expected::Error(Diagnostic { code: "58P01", message: r#"could not open file "/tmp/doltgres-kept-missing.txt" for reading: No such file or directory"#, hint: r#"COPY FROM instructs the PostgreSQL server process to read a file. You may want a client-side facility such as psql's \copy."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
