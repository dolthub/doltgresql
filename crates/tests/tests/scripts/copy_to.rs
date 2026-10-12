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
fn test_copy_to() {
    run_scripts(&[
        ScriptTest {
            name: "tab delimited to stdout",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
                r#"INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40), (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 TO STDOUT;",
                    expected: Expected::Tag("COPY 7"),
                    copy_to_stdout_file: "copy-to-basic.txt",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tab delimited with header to stdout",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
                r#"INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40), (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 TO STDOUT WITH (HEADER);",
                    expected: Expected::Tag("COPY 7"),
                    copy_to_stdout_file: "copy-to-header.txt",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "tab delimited with column names",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
                r#"INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40), (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 (c1, pk) TO STDOUT;",
                    expected: Expected::Tag("COPY 7"),
                    copy_to_stdout_file: "copy-to-columns.txt",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "csv with header to stdout",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
                r#"INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40), (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 TO STDOUT (FORMAT CSV, HEADER);",
                    expected: Expected::Tag("COPY 7"),
                    copy_to_stdout_file: "copy-to-basic.csv",
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 TO STDOUT (FORMAT 'csv', HEADER);",
                    expected: Expected::Tag("COPY 7"),
                    copy_to_stdout_file: "copy-to-basic.csv",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "psv with header to stdout",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
                r#"INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40), (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl1 TO STDOUT (FORMAT CSV, HEADER, DELIMITER '|');",
                    expected: Expected::Tag("COPY 7"),
                    copy_to_stdout_file: "copy-to-header.psv",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "query to stdout",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
                r#"INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40), (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY (SELECT pk, c1 FROM tbl1 WHERE pk < 3 ORDER BY pk DESC) TO STDOUT;",
                    expected: Expected::Tag("COPY 2"),
                    copy_to_stdout_file: "copy-to-query.txt",
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY (SELECT 1 AS x UNION ALL SELECT 2 ORDER BY x) TO STDOUT;",
                    expected: Expected::Tag("COPY 2"),
                    copy_to_stdout_file: "copy-to-union.txt",
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY (SELECT pk FROM tbl1 WHERE pk > 100) TO STDOUT;",
                    expected: Expected::Tag("COPY 0"),
                    copy_to_stdout_file: "copy-to-empty.txt",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "boolean values to stdout",
            set_up_script: &[
                "CREATE TABLE tbl2 (pk int primary key, b boolean);",
                "INSERT INTO tbl2 VALUES (1, true), (2, false), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl2 TO STDOUT;",
                    expected: Expected::Tag("COPY 3"),
                    copy_to_stdout_file: "copy-to-bool.txt",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "schema qualified table to stdout",
            set_up_script: &[
                "CREATE SCHEMA s1;",
                "CREATE TABLE s1.tbl2 (pk int primary key, b boolean);",
                "INSERT INTO s1.tbl2 VALUES (1, true), (2, false), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY s1.tbl2 TO STDOUT;",
                    expected: Expected::Tag("COPY 3"),
                    copy_to_stdout_file: "copy-to-bool.txt",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "binary to stdout",
            set_up_script: &[
                "CREATE TABLE tbl3 (pk int primary key, c1 text, b boolean);",
                "INSERT INTO tbl3 VALUES (1, 'foo', true), (2, NULL, false), (3, '', NULL), (4, 'héllo', true);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl3 TO STDOUT (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 4"),
                    copy_to_stdout_file: "copy-to-basic.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl3 TO STDOUT BINARY;",
                    expected: Expected::Tag("COPY 4"),
                    copy_to_stdout_file: "copy-to-basic.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY (SELECT * FROM tbl3 ORDER BY pk) TO STDOUT (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 4"),
                    copy_to_stdout_file: "copy-to-basic.bin",
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"COPY (SELECT * FROM tbl3 ORDER BY pk) TO STDOUT (FORMAT "binary");"#,
                    expected: Expected::Tag("COPY 4"),
                    copy_to_stdout_file: "copy-to-basic.bin",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "csv round trip",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
                r#"INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40), (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE tbl1_copy (pk int primary key, c1 varchar(100), c2 int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 TO STDOUT (FORMAT CSV, HEADER);",
                    expected: Expected::Tag("COPY 7"),
                    copy_round_trip_stdin_query: "COPY tbl1_copy FROM STDIN (FORMAT CSV, HEADER);",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM tbl1 t1 JOIN tbl1_copy t2 ON t1.pk = t2.pk WHERE t1.c1 IS NOT DISTINCT FROM t2.c1 AND t1.c2 IS NOT DISTINCT FROM t2.c2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
        ScriptTest {
            name: "tab delimited round trip",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
                r#"INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40), (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE tbl1_copy (pk int primary key, c1 varchar(100), c2 int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 TO STDOUT;",
                    expected: Expected::Tag("COPY 7"),
                    copy_round_trip_stdin_query: "COPY tbl1_copy FROM STDIN;",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM tbl1 t1 JOIN tbl1_copy t2 ON t1.pk = t2.pk WHERE t1.c1 IS NOT DISTINCT FROM t2.c1 AND t1.c2 IS NOT DISTINCT FROM t2.c2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
        ScriptTest {
            name: "tab delimited escape characters round trip",
            set_up_script: &[
                "CREATE TABLE esc (pk int primary key, c1 text);",
                r#"INSERT INTO esc VALUES
					(1, E'tab\tseparated'),
					(2, E'new\nline'),
					(3, E'carriage\rreturn'),
					(4, E'back\\slash'),
					(5, E'\b\f\v'),
					(6, E'\\N'),
					(7, E'ends with backslash\\'),
					(8, E'\\.'),
					(9, E'mixed\t\\.\nescapes\\\r'),
					(10, 'pipe|delimiter');"#,
                "CREATE TABLE esc_copy (pk int primary key, c1 text);",
                "CREATE TABLE esc_copy2 (pk int primary key, c1 text);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY esc TO STDOUT;",
                    expected: Expected::Tag("COPY 10"),
                    copy_round_trip_stdin_query: "COPY esc_copy FROM STDIN;",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM esc t1 JOIN esc_copy t2 ON t1.pk = t2.pk WHERE t1.c1 = t2.c1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY esc TO STDOUT (FORMAT TEXT, DELIMITER '|');",
                    expected: Expected::Tag("COPY 10"),
                    copy_round_trip_stdin_query: "COPY esc_copy2 FROM STDIN (FORMAT TEXT, DELIMITER '|');",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM esc t1 JOIN esc_copy2 t2 ON t1.pk = t2.pk WHERE t1.c1 = t2.c1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "single column round trip with NULL",
            set_up_script: &[
                "CREATE TABLE single (c1 text);",
                "INSERT INTO single VALUES ('foo'), (NULL), ('');",
                "CREATE TABLE single_copy_text (c1 text);",
                "CREATE TABLE single_copy_csv (c1 text);",
                "CREATE TABLE single_copy_bin (c1 text);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY single TO STDOUT;",
                    expected: Expected::Tag("COPY 3"),
                    copy_round_trip_stdin_query: "COPY single_copy_text FROM STDIN;",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), count(c1), count(CASE WHEN c1 = '' THEN 1 END) FROM single_copy_text;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8), Column("count", INT8)],
                        rows: &[
                            &[T("3"), T("2"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY single TO STDOUT (FORMAT CSV);",
                    expected: Expected::Tag("COPY 3"),
                    copy_round_trip_stdin_query: "COPY single_copy_csv FROM STDIN (FORMAT CSV);",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), count(c1), count(CASE WHEN c1 = '' THEN 1 END) FROM single_copy_csv;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8), Column("count", INT8)],
                        rows: &[
                            &[T("3"), T("2"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY single TO STDOUT (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 3"),
                    copy_round_trip_stdin_query: "COPY single_copy_bin FROM STDIN (FORMAT BINARY);",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), count(c1), count(CASE WHEN c1 = '' THEN 1 END) FROM single_copy_bin;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8), Column("count", INT8)],
                        rows: &[
                            &[T("3"), T("2"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "binary round trip",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
                r#"INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40), (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE tbl1_copy (pk int primary key, c1 varchar(100), c2 int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 TO STDOUT (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 7"),
                    copy_round_trip_stdin_query: "COPY tbl1_copy FROM STDIN (FORMAT BINARY);",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM tbl1 t1 JOIN tbl1_copy t2 ON t1.pk = t2.pk WHERE t1.c1 IS NOT DISTINCT FROM t2.c1 AND t1.c2 IS NOT DISTINCT FROM t2.c2;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
        ScriptTest {
            name: "binary round trip over multiple chunks",
            set_up_script: &[
                "CREATE TABLE big (pk int primary key, c1 text);",
                "INSERT INTO big SELECT i, repeat('x', 200) || i FROM generate_series(1, 1000) g(i);",
                "CREATE TABLE big_copy (pk int primary key, c1 text);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY big TO STDOUT (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 1000"),
                    copy_round_trip_stdin_query: "COPY big_copy FROM STDIN (FORMAT BINARY);",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM big t1 JOIN big_copy t2 ON t1.pk = t2.pk WHERE t1.c1 = t2.c1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "binary round trip with various types",
            set_up_script: &[
                r#"CREATE TABLE typed (pk int primary key, i2 int2, i8 int8, f4 float4, f8 float8, n numeric(10,2),
					d date, ts timestamp, u uuid, by bytea, b boolean);"#,
                r#"INSERT INTO typed VALUES
					(1, 32767, 9223372036854775807, 1.5, -2.25, 12345.67, '2025-01-01', '2025-01-01 12:34:56',
					 '1077f506-a6fc-4cb2-aed2-9dea9351ed9c', '\xdeadbeef', true),
					(2, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
					(3, -32768, -9223372036854775808, 0.0, 1e10, -0.01, '1999-12-31', '1999-12-31 23:59:59',
					 '428d0815-d95b-4cfc-89af-9fca38585dcc', '\x00', false);"#,
                r#"CREATE TABLE typed_copy (pk int primary key, i2 int2, i8 int8, f4 float4, f8 float8, n numeric(10,2),
					d date, ts timestamp, u uuid, by bytea, b boolean);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY typed TO STDOUT (FORMAT BINARY);",
                    expected: Expected::Tag("COPY 3"),
                    copy_round_trip_stdin_query: "COPY typed_copy FROM STDIN (FORMAT BINARY);",
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT count(*) FROM typed t1 JOIN typed_copy t2 ON t1.pk = t2.pk
						WHERE t1.i2 IS NOT DISTINCT FROM t2.i2 AND t1.i8 IS NOT DISTINCT FROM t2.i8
						AND t1.f4 IS NOT DISTINCT FROM t2.f4 AND t1.f8 IS NOT DISTINCT FROM t2.f8
						AND t1.n IS NOT DISTINCT FROM t2.n AND t1.d IS NOT DISTINCT FROM t2.d
						AND t1.ts IS NOT DISTINCT FROM t2.ts AND t1.u IS NOT DISTINCT FROM t2.u
						AND t1.by IS NOT DISTINCT FROM t2.by AND t1.b IS NOT DISTINCT FROM t2.b;"#,
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
            name: "errors",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
                r#"INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40), (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY tbl2 TO STDOUT;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "tbl2" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 (pk, c3) TO STDOUT;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "c3" of relation "tbl1" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 TO STDOUT (FORMAT BINARY, HEADER);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "cannot specify HEADER in BINARY mode", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 TO STDOUT (FORMAT BINARY, DELIMITER '|');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "cannot specify DELIMITER in BINARY mode", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"COPY tbl1 TO STDOUT (FORMAT "nonsense");"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"COPY format "nonsense" not recognized"#, position: 22, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY tbl1 TO '{TEMPDIR}/copy-to-out.csv' (FORMAT CSV);",
                    expected: Expected::Tag("COPY 7"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
