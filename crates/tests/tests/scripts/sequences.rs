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
fn test_sequences() {
    run_scripts(&[
        ScriptTest {
            name: "Basic CREATE SEQUENCE and DROP SEQUENCE",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test'::regclass);",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('doesnotexist'::regclass);",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "doesnotexist" does not exist"#, position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SEQUENCE test;",
                    expected: Expected::Tag("DROP SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test" does not exist"#, position: 16, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE SEQUENCE IF NOT EXISTS",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "test1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE IF NOT EXISTS test1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "test1" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE IF NOT EXISTS test2;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE IF NOT EXISTS test2;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "test2" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
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
            name: "DROP SEQUENCE IF NOT EXISTS",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test2;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SEQUENCE test1;",
                    expected: Expected::Tag("DROP SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SEQUENCE test1;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"sequence "test1" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SEQUENCE IF EXISTS test1;",
                    expected: Expected::Tag("DROP SEQUENCE"),
                    notices: &[Diagnostic { code: "00000", message: r#"sequence "test1" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test1" does not exist"#, position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SEQUENCE IF EXISTS test2;",
                    expected: Expected::Tag("DROP SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test2" does not exist"#, position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SEQUENCE IF EXISTS test2;",
                    expected: Expected::Tag("DROP SEQUENCE"),
                    notices: &[Diagnostic { code: "00000", message: r#"sequence "test2" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "MINVALUE and MAXVALUE with DATA TYPE",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test1 AS SMALLINT MINVALUE -32768;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test2 AS SMALLINT MINVALUE -32769;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "MINVALUE (-32769) is out of range for sequence data type smallint", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test3 AS SMALLINT MAXVALUE 32767;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test4 AS SMALLINT MINVALUE 32768;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "MINVALUE (32768) is out of range for sequence data type smallint", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test5 AS INTEGER MINVALUE -2147483648;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test6 AS INTEGER MINVALUE -2147483649;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "MINVALUE (-2147483649) is out of range for sequence data type integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test7 AS INTEGER MAXVALUE 2147483647;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test8 AS INTEGER MINVALUE 2147483648;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "MINVALUE (2147483648) is out of range for sequence data type integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test9 AS BIGINT MINVALUE -9223372036854775808;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test10 AS BIGINT MINVALUE -9223372036854775809;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#"value "-9223372036854775809" is out of range for type bigint"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test11 AS BIGINT MAXVALUE 9223372036854775807;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test12 AS BIGINT MINVALUE 9223372036854775808;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#"value "9223372036854775808" is out of range for type bigint"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE SEQUENCE START",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test1 START 39;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("39")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test2 START 0;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "START value (0) cannot be less than MINVALUE (1)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test2 MINVALUE 0 START 0;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test3 MINVALUE -100 START -7;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test3');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("-7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test4 START -5 INCREMENT 1;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "START value (-5) cannot be less than MINVALUE (1)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test4 START -5 INCREMENT -1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test4');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("-5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test5 START 25 INCREMENT -1;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "START value (25) cannot be greater than MAXVALUE (-1)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test5 START 25 MAXVALUE 25 INCREMENT -1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test5');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("25")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test5');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("24")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE and NO CYCLE",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test1 MINVALUE 0 MAXVALUE 3 START 2 INCREMENT 1 NO CYCLE;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Error(Diagnostic { code: "2200H", message: r#"nextval: reached maximum value of sequence "test1" (3)"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test2 MINVALUE 0 MAXVALUE 3 START 2 INCREMENT 1 CYCLE;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test3 MINVALUE 0 MAXVALUE 3 START 1 INCREMENT -1 NO CYCLE;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test3');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test3');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test3');",
                    expected: Expected::Error(Diagnostic { code: "2200H", message: r#"nextval: reached minimum value of sequence "test3" (0)"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test4 MINVALUE 0 MAXVALUE 3 START 1 INCREMENT -1 CYCLE;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test4');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test4');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test4');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test5 MINVALUE 1 MAXVALUE 7 START 1 INCREMENT 5 CYCLE;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test5');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test5');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test5');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test6 MINVALUE 1 MAXVALUE 7 START 6 INCREMENT -5 CYCLE;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test6');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test6');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test6');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test6');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
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
            name: "nextval() over multiple rows/columns",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE test (v1 INTEGER, v2 INTEGER);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE seq1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (nextval('seq1'), 7), (nextval('seq1'), 11), (nextval('seq1'), 17);",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("7")],
                            &[T("2"), T("11")],
                            &[T("3"), T("17")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (nextval('seq1'), nextval('seq1')), (nextval('seq1'), nextval('seq1'));",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("7")],
                            &[T("2"), T("11")],
                            &[T("3"), T("17")],
                            &[T("4"), T("5")],
                            &[T("6"), T("7")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nextval() with double-quoted identifiers",
            set_up_script: &[
                "CREATE SEQUENCE test_sequence;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT nextval('test_sequence');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('public.test_sequence');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT nextval('"test_sequence"');"#,
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT nextval('public."test_sequence"');"#,
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
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
            name: "nextval() in filter",
            set_up_script: &[
                "CREATE TABLE test_serial (v1 SERIAL, v2 INTEGER);",
                "INSERT INTO test_serial (v2) VALUES (4), (5), (6);",
                "CREATE TABLE test_seq (v1 INTEGER, v2 INTEGER);",
                "CREATE SEQUENCE test_sequence OWNED BY test_seq.v1;",
                "INSERT INTO test_seq VALUES (nextval('test_sequence'), 4), (nextval('test_sequence'), 5), (nextval('test_sequence'), 6);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test_serial WHERE nextval('test_serial_v1_seq') = v2 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("4")],
                            &[T("2"), T("5")],
                            &[T("3"), T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test_serial_v1_seq');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_seq WHERE nextval('test_sequence') = v2 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("4")],
                            &[T("2"), T("5")],
                            &[T("3"), T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test_sequence');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
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
            name: "setval()",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test1 MINVALUE 1 MAXVALUE 10 START 5 INCREMENT 1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test2 MINVALUE 1 MAXVALUE 10 START 5 INCREMENT -1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test1', 2);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test1', 10);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Error(Diagnostic { code: "2200H", message: r#"nextval: reached maximum value of sequence "test1" (10)"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test1', 10, false);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test1', 10, true);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Error(Diagnostic { code: "2200H", message: r#"nextval: reached maximum value of sequence "test1" (10)"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test2', 9);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test2', 1);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Error(Diagnostic { code: "2200H", message: r#"nextval: reached minimum value of sequence "test2" (1)"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test2', 1, false);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test2', 1, true);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Error(Diagnostic { code: "2200H", message: r#"nextval: reached minimum value of sequence "test2" (1)"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test3 MINVALUE 3 MAXVALUE 7 START 5 INCREMENT 1 CYCLE;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test4 MINVALUE 3 MAXVALUE 7 START 5 INCREMENT -1 CYCLE;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test3', 7, true);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test3');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test4', 3, true);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test4');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test5;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT setval('public."test5"', 100, true);"#,
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test5');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("101")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SERIAL",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE test (pk SERIAL PRIMARY KEY, v1 INTEGER);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test_small (pk SMALLSERIAL PRIMARY KEY, v1 INTEGER);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test_big (pk BIGSERIAL PRIMARY KEY, v1 INTEGER);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test (v1) VALUES (2), (3), (5), (7), (11);",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test_small (v1) VALUES (2), (3), (5), (7), (11);",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test_big (v1) VALUES (2), (3), (5), (7), (11);",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("3"), T("5")],
                            &[T("4"), T("7")],
                            &[T("5"), T("11")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_small;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT2), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("3"), T("5")],
                            &[T("4"), T("7")],
                            &[T("5"), T("11")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test_big;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT8), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("3"), T("5")],
                            &[T("4"), T("7")],
                            &[T("5"), T("11")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SERIAL type created in table of different schema",
            set_up_script: &[
                "CREATE SCHEMA myschema",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE myschema.test (pk SERIAL PRIMARY KEY, v1 INTEGER);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO myschema.test (v1) VALUES (2), (3), (5), (7), (11);",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM myschema.test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("3"), T("5")],
                            &[T("4"), T("7")],
                            &[T("5"), T("11")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test_pk_seq');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test_pk_seq" does not exist"#, position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('myschema.test_pk_seq');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('postgres.myschema.test_pk_seq');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
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
            name: "Default emulating SERIAL",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE seq1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test (pk INTEGER DEFAULT (nextval('seq1')), v1 INTEGER);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test (v1) VALUES (2), (3), (5), (7), (11);",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("3"), T("5")],
                            &[T("4"), T("7")],
                            &[T("5"), T("11")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Default emulating SERIAL in non default schema",
            set_up_script: &[
                "CREATE SCHEMA myschema",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE myschema.seq1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE myschema.test (pk INTEGER DEFAULT (nextval('seq1')), v1 INTEGER);",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "seq1" does not exist"#, position: 57, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO myschema.test (v1) VALUES (2), (3), (5), (7), (11);",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "myschema.test" does not exist"#, position: 13, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM myschema.test ORDER BY v1;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "myschema.test" does not exist"#, position: 15, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_sequence",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "pg_catalog"."pg_sequence";"#,
                    expected: Expected::Rows {
                        columns: &[Column("seqrelid", OID), Column("seqtypid", OID), Column("seqstart", INT8), Column("seqincrement", INT8), Column("seqmax", INT8), Column("seqmin", INT8), Column("seqcache", INT8), Column("seqcycle", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "PG_catalog"."pg_sequence";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "PG_catalog.pg_sequence" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "pg_catalog"."PG_sequence";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "pg_catalog.PG_sequence" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE some_sequence;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE another_sequence INCREMENT 3 CYCLE;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pg_catalog.pg_sequence ORDER BY seqrelid;",
                    expected: Expected::Rows {
                        columns: &[Column("seqrelid", OID), Column("seqtypid", OID), Column("seqstart", INT8), Column("seqincrement", INT8), Column("seqmax", INT8), Column("seqmin", INT8), Column("seqcache", INT8), Column("seqcycle", BOOL)],
                        rows: &[
                            &[T("16384"), T("20"), T("1"), T("1"), T("9223372036854775807"), T("1"), T("1"), T("f")],
                            &[T("16385"), T("20"), T("1"), T("3"), T("9223372036854775807"), T("1"), T("1"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM PG_catalog.pg_SEQUENCE ORDER BY seqrelid;",
                    expected: Expected::Rows {
                        columns: &[Column("seqrelid", OID), Column("seqtypid", OID), Column("seqstart", INT8), Column("seqincrement", INT8), Column("seqmax", INT8), Column("seqmin", INT8), Column("seqcache", INT8), Column("seqcycle", BOOL)],
                        rows: &[
                            &[T("16384"), T("20"), T("1"), T("1"), T("9223372036854775807"), T("1"), T("1"), T("f")],
                            &[T("16385"), T("20"), T("1"), T("3"), T("9223372036854775807"), T("1"), T("1"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pg_catalog.pg_sequence WHERE seqrelid = 'some_sequence'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("seqrelid", OID), Column("seqtypid", OID), Column("seqstart", INT8), Column("seqincrement", INT8), Column("seqmax", INT8), Column("seqmin", INT8), Column("seqcache", INT8), Column("seqcycle", BOOL)],
                        rows: &[
                            &[T("16384"), T("20"), T("1"), T("1"), T("9223372036854775807"), T("1"), T("1"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pg_catalog.pg_sequence WHERE seqrelid = 'another_sequence'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("seqrelid", OID), Column("seqtypid", OID), Column("seqstart", INT8), Column("seqincrement", INT8), Column("seqmax", INT8), Column("seqmin", INT8), Column("seqcache", INT8), Column("seqcycle", BOOL)],
                        rows: &[
                            &[T("16385"), T("20"), T("1"), T("3"), T("9223372036854775807"), T("1"), T("1"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('another_sequence');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('another_sequence');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
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
            name: "DROP TABLE",
            set_up_script: &[
                "CREATE TABLE test (pk SERIAL PRIMARY KEY, v1 INTEGER);",
                "INSERT INTO test (v1) VALUES (2), (3), (5), (7), (11);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("3"), T("5")],
                            &[T("4"), T("7")],
                            &[T("5"), T("11")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pg_catalog.pg_sequence;",
                    expected: Expected::Rows {
                        columns: &[Column("seqrelid", OID), Column("seqtypid", OID), Column("seqstart", INT8), Column("seqincrement", INT8), Column("seqmax", INT8), Column("seqmin", INT8), Column("seqcache", INT8), Column("seqcycle", BOOL)],
                        rows: &[
                            &[T("16384"), T("23"), T("1"), T("1"), T("2147483647"), T("1"), T("1"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pg_catalog.pg_sequence;",
                    expected: Expected::Rows {
                        columns: &[Column("seqrelid", OID), Column("seqtypid", OID), Column("seqstart", INT8), Column("seqincrement", INT8), Column("seqmax", INT8), Column("seqmin", INT8), Column("seqcache", INT8), Column("seqcycle", BOOL)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "seq name generation",
            set_up_script: &[
                "CREATE SEQUENCE my_table_id_seq;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE my_table (id SERIAL PRIMARY KEY, val INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select nextval('my_table_id_seq1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "Select count(*) from pg_catalog.pg_sequence where seqrelid = 'my_table_id_seq1'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
            name: "drop and create table with same name (issue 659)",
            set_up_script: &[
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
                "create table serial_table (pk serial primary key);",
                "drop table serial_table;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "create table serial_table (pk serial primary key);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "identity generated by default",
            set_up_script: &[
                r#"CREATE TABLE "django_migrations" (
    "id" bigint NOT NULL PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
		"app" varchar(255) NOT NULL,
		"name" varchar(255) NOT NULL,
		"applied" timestamp with time zone NOT NULL)"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"INSERT INTO "django_migrations" ("app", "name", "applied") VALUES ('contenttypes', '0001_initial', '2025-03-25T17:45:54.794344+00:00'::timestamptz) RETURNING "django_migrations"."id""#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO "django_migrations" ("app", "name", "applied") VALUES ('contenttypes', '0001_initial', '2025-03-25T17:45:54.794344+00:00'::timestamptz) RETURNING "django_migrations"."id""#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO "django_migrations" ("id", "app", "name", "applied") VALUES (100, 'contenttypes', '0001_initial', '2025-03-25T17:45:54.794344+00:00'::timestamptz) RETURNING "django_migrations"."id""#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "identity generated by default with sequence options",
            set_up_script: &[
                r#"CREATE TABLE "django_migrations" (
    "id" bigint NOT NULL PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY (START WITH 100 INCREMENT BY 2),
		"app" varchar(255) NOT NULL,
		"name" varchar(255) NOT NULL,
		"applied" timestamp with time zone NOT NULL)"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"INSERT INTO "django_migrations" ("app", "name", "applied") VALUES ('contenttypes', '0001_initial', '2025-03-25T17:45:54.794344+00:00'::timestamptz) RETURNING "django_migrations"."id""#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO "django_migrations" ("app", "name", "applied") VALUES ('contenttypes', '0001_initial', '2025-03-25T17:45:54.794344+00:00'::timestamptz) RETURNING "django_migrations"."id""#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8)],
                        rows: &[
                            &[T("102")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "insert on a different branch",
            set_up_script: &[
                "create table test (pk serial primary key, v1 int);",
                "insert into test (v1) values (2), (3), (5), (7), (11);",
                "call dolt_branch('b1');",
                r#"create table "postgres/b1".public.test2 (pk serial primary key, v1 int);"#,
                r#"insert into "postgres/b1".public.test2 (v1) values (2), (3), (5), (7), (11);"#,
            ],
            skip: Some(r#"setup fails on Postgres ("error running setup query: call dolt_branch('b1');: ERROR: procedure dolt_branch(unknown) does not exist (SQLSTATE 42883)") and on the Go server ("error running setup query: call dolt_branch('b1');: ERROR: Dolt stored procedure may only be invoked using SELECT (SQLSTATE XX000)")"#),
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pk FROM test ORDER BY v1;",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT pk FROM "postgres/b1".public.test2 ORDER BY v1;"#,
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "sequences are globally tracked across dolt_add, dolt_branch, dolt_checkout, dolt_commit, dolt_reset",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test', 10);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("public.test"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('test');",
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
                    query: "SELECT length(dolt_commit('-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_branch('other');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test', 20);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('.');",
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
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("21")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('other');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'other'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("22")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_reset('--hard');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("23")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt_clean",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test2;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test1', 10);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test2', 10);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('test1');",
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
                            &[T("public.test1"), T("t"), T("new table")],
                            &[T("public.test2"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_clean('test2');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clean", INT8)],
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
                            &[T("public.test1"), T("t"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt_merge",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test', 10);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-Am', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_branch('other');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test', 20);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-am', 'next')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('other');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'other'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT setval('test', 30);",
                    expected: Expected::Rows {
                        columns: &[Column("setval", INT8)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-am', 'next2')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("31")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_reset('--hard');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT strpos(dolt_merge('other')::text, 'merge successful') > 32;",
                    expected: Expected::Rows {
                        columns: &[Column("strpos > 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Information Schema & DIFF_STAT regression testing",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE TABLE "user" ("id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY, PRIMARY KEY ("id"));"#,
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE TABLE "call" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "state" character varying NOT NULL,
  "content" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "ended_at" timestamptz NULL,
  "user" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "call_user_user_fk" FOREIGN KEY ("user") REFERENCES "user" ("id") ON DELETE NO ACTION
);"#,
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM information_schema.key_column_usage where constraint_schema <> 'pg_catalog';",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_catalog", NAME), Column("constraint_schema", NAME), Column("constraint_name", NAME), Column("table_catalog", NAME), Column("table_schema", NAME), Column("table_name", NAME), Column("column_name", NAME), Column("ordinal_position", INT4), Column("position_in_unique_constraint", INT4)],
                        rows: &[
                            &[T("postgres"), T("public"), T("user_pkey"), T("postgres"), T("public"), T("user"), T("id"), T("1"), Null],
                            &[T("postgres"), T("public"), T("call_pkey"), T("postgres"), T("public"), T("call"), T("id"), T("1"), Null],
                            &[T("postgres"), T("public"), T("call_user_user_fk"), T("postgres"), T("public"), T("call"), T("user"), T("1"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_DIFF_STAT('HEAD', 'WORKING');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("rows_unmodified", INT8), Column("rows_added", INT8), Column("rows_deleted", INT8), Column("rows_modified", INT8), Column("cells_added", INT8), Column("cells_deleted", INT8), Column("cells_modified", INT8), Column("old_row_count", INT8), Column("new_row_count", INT8), Column("old_cell_count", INT8), Column("new_cell_count", INT8)],
                        rows: &[
                            &[T("public.call_id_seq"), T("0"), T("1"), T("0"), T("0"), T("0"), T("0"), T("0"), T("0"), T("0"), T("0"), T("0")],
                            &[T("public.user_id_seq"), T("0"), T("1"), T("0"), T("0"), T("0"), T("0"), T("0"), T("0"), T("0"), T("0"), T("0")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"SELECT n.nspname as "Schema",
  c.relname as "Name",
  CASE c.relkind WHEN 'r' THEN 'table' WHEN 'v' THEN 'view' WHEN 'm' THEN 'materialized view' WHEN 'i' THEN 'index' WHEN 'S' THEN 'sequence' WHEN 't' THEN 'TOAST table' WHEN 'f' THEN 'foreign table' WHEN 'p' THEN 'partitioned table' WHEN 'I' THEN 'partitioned index' END as "Type",
  pg_catalog.pg_get_userbyid(c.relowner) as "Owner"
FROM pg_catalog.pg_class c
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
     LEFT JOIN pg_catalog.pg_am am ON am.oid = c.relam
WHERE c.relkind IN ('r','p','v','m','S','f','')
      AND n.nspname <> 'pg_catalog'
      AND n.nspname !~ '^pg_toast'
      AND n.nspname <> 'information_schema'
  AND pg_catalog.pg_table_is_visible(c.oid)
ORDER BY 1,2;"#,
                    expected: Expected::Rows {
                        columns: &[Column("Schema", NAME), Column("Name", NAME), Column("Type", TEXT), Column("Owner", TEXT)],
                        rows: &[
                            &[T("public"), T("call"), T("table"), T("postgres")],
                            &[T("public"), T("call_id_seq"), T("sequence"), T("postgres")],
                            &[T("public"), T("user"), T("table"), T("postgres")],
                            &[T("public"), T("user_id_seq"), T("sequence"), T("postgres")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER COLUMN ADD GENERATED BY DEFAULT",
            set_up_script: &[
                "CREATE TABLE public.test1 (id int2 NOT NULL, name character varying(150) NOT NULL);",
                "CREATE TABLE public.test2 (id int4 NOT NULL, name character varying(150) NOT NULL);",
                "CREATE TABLE public.test3 (id int8 NOT NULL, name character varying(150) NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE public.test1 ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (SEQUENCE NAME public.test1_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE public.test2 ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (SEQUENCE NAME public.test2_id_seq START WITH 10 INCREMENT BY 5);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE public.test3 ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (SEQUENCE NAME public.test3_id_seq START WITH 100 INCREMENT BY -1 MAXVALUE 100);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.test1 (name) VALUES ('abc'), ('def');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.test2 (name) VALUES ('abc'), ('def');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.test3 (name) VALUES ('abc'), ('def');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.test1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT2), Column("name", VARCHAR)],
                        rows: &[
                            &[T("1"), T("abc")],
                            &[T("2"), T("def")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.test2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", VARCHAR)],
                        rows: &[
                            &[T("10"), T("abc")],
                            &[T("15"), T("def")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.test3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("name", VARCHAR)],
                        rows: &[
                            &[T("100"), T("abc")],
                            &[T("99"), T("def")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER SEQUENCE OWNED BY",
            set_up_script: &[
                "CREATE SCHEMA other;",
                "CREATE TABLE test (id int4 NOT NULL);",
                "CREATE TABLE other.test (id int4 NOT NULL);",
                "CREATE SEQUENCE seq1;",
                "CREATE SEQUENCE seq2;",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION f_trigger();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT nextval('seq1');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('seq2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE seq1 OWNED BY test.non_existent;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "non_existent" of relation "test" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE public.seq1 OWNED BY other.test.non_existent;",
                    expected: Expected::Error(Diagnostic { code: "55000", message: "sequence must be in same schema as table it is linked to", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE seq1 OWNED BY test;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "invalid OWNED BY option", hint: "Specify OWNED BY table.column or OWNED BY NONE.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE seq1 OWNED BY trig_trigger.trig;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "trig_trigger" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE seq1 OWNED BY test.id;",
                    expected: Expected::Tag("ALTER SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE seq2 OWNED BY test.id;",
                    expected: Expected::Tag("ALTER SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE seq2 OWNED BY NONE;",
                    expected: Expected::Tag("ALTER SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('seq1');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "seq1" does not exist"#, position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('seq2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
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
            name: "sequence collection loaded independently with multiple databases",
            set_up_script: &[
                "CREATE DATABASE testdb2",
                "USE testdb2",
                "CREATE SEQUENCE seq_in_testdb2",
                "USE postgres",
                "CREATE SEQUENCE seq_in_postgres",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('seq_in_postgres'), 'testdb2.public.seq_in_testdb2'::regclass IS NOT NULL",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8), Column("'testdb2.public.seq_in_testdb2'::REGCLASS IS NOT NULL", BOOL)],
                        rows: &[
                            &[T("1"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Sequence names must be unique across all relation types",
            set_up_script: &[
                "CREATE TABLE tbl1 (pk int PRIMARY KEY, v1 int);",
                "CREATE SEQUENCE existing_seq;",
                "CREATE VIEW view1 AS SELECT pk FROM tbl1;",
                "CREATE INDEX idx1 ON tbl1 (v1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE tbl1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "tbl1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE IF NOT EXISTS tbl1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "tbl1" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE existing_seq;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "existing_seq" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE IF NOT EXISTS existing_seq;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "existing_seq" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE view1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "view1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE IF NOT EXISTS view1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "view1" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE idx1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "idx1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE IF NOT EXISTS idx1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "idx1" already exists, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Error when branches contain incompatible sequence definitions",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('-b', 'other');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'other'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE test INCREMENT -1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "unable to advance sequence state, possibly due to having incompatible state on different branches", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
