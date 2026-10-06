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
fn test_copy_from_stdin_extended_protocol() {
    run_wire_tests(&[
        WireTest {
            name: "extended copy completes at sync",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "", query: "COPY test3 FROM STDIN", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::CopyDone,
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COPY 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "extended copy failure discards until sync",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "", query: "COPY test3 FROM STDIN", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyFail("abort extended copy"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "57014", message: "COPY from stdin failed: abort extended copy", where_: "COPY test3, line 1", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
    ]);
}

#[test]
fn test_copy_from_stdin_in_multi_statement_simple_query() {
    run_wire_tests(&[
        WireTest {
            name: "empty copy initializes and completes the transfer",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("COPY test3 FROM STDIN; SELECT count(*) FROM test3;"),
                ]),
                Step::Receive(&[
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyDone,
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COPY 0"),
                    Receive::RowDescription(&[Field { name: "count", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("0")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "copy fail before data aborts cleanly",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("COPY test3 FROM STDIN; SELECT 1;"),
                ]),
                Step::Receive(&[
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyFail("client aborted empty copy"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "57014", message: "COPY from stdin failed: client aborted empty copy", where_: "COPY test3, line 1", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
        WireTest {
            name: "flush and sync are ignored during copy",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("COPY test3 FROM STDIN; SELECT count(*) FROM test3;"),
                ]),
                Step::Receive(&[
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::Flush,
                    Send::Sync,
                    Send::CopyData(b"1\n"),
                    Send::CopyDone,
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COPY 1"),
                    Receive::RowDescription(&[Field { name: "count", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "multiple copy inputs preserve statement order",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("SELECT 0; COPY test3 FROM STDIN; COPY test3 FROM STDIN; SELECT 1;"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("0")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::CopyDone,
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COPY 1"),
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"2\n"),
                    Send::CopyDone,
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COPY 1"),
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT * FROM test3 ORDER BY c;"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "c", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::DataRow(&[Datum::Text("2")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("DROP TABLE test3;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("DROP TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "copy first rolls back when a later statement fails",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("COPY test3 FROM STDIN; SELECT * FROM missing_table;"),
                ]),
                Step::Receive(&[
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::CopyDone,
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COPY 1"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "missing_table" does not exist"#, position: 38, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
        WireTest {
            name: "copy fail rolls back compound query",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO test3 VALUES (0); COPY test3 FROM STDIN; SELECT 1;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::CopyData(b"2\n"),
                    Send::CopyFail("client aborted copy"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "57014", message: "COPY from stdin failed: client aborted copy", where_: "COPY test3, line 3", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
        WireTest {
            name: "copy fail marks explicit transaction failed",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN; COPY test3 FROM STDIN; SELECT 1;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::CopyFail("client aborted copy"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "57014", message: "COPY from stdin failed: client aborted copy", where_: "COPY test3, line 2", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Query("ROLLBACK;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
        WireTest {
            name: "explicit transaction commits copy and surrounding statements",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN; INSERT INTO test3 VALUES (0); COPY test3 FROM STDIN; INSERT INTO test3 VALUES (2); COMMIT;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::CopyDone,
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COPY 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("COMMIT"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3 ORDER BY c;", rows: &[&[T("0")], &[T("1")], &[T("2")], ] },
            ],
            ..W
        },
        WireTest {
            name: "explicit transaction rolls back copy and surrounding statements",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN; INSERT INTO test3 VALUES (0); COPY test3 FROM STDIN; INSERT INTO test3 VALUES (2); ROLLBACK;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::CopyDone,
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COPY 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
        WireTest {
            name: "copy compound query continues an existing explicit transaction",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("COPY test3 FROM STDIN; INSERT INTO test3 VALUES (2);"),
                ]),
                Step::Receive(&[
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::CopyDone,
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COPY 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
                Step::Send(&[
                    Send::Query("ROLLBACK;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
    ]);
}

#[test]
fn test_copy_unexpected_message_fatal() {
    run_wire_tests(&[
        WireTest {
            name: "simple origin rejects *pgproto3.Query",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("COPY test3 FROM STDIN"),
                ]),
                Step::Receive(&[
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::Query("SELECT 2"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "08P01", message: "unexpected message type 0x51 during COPY from stdin", where_: "COPY test3, line 2", ..F }),
                    Receive::Error(Fields { severity: "FATAL", severity_unlocalized: "FATAL", code: "08P01", message: "terminating connection because protocol synchronization was lost", ..F }),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
    ]);
}

#[test]
fn test_copy_unexpected_message_fatal_2() {
    run_wire_tests(&[
        WireTest {
            name: "simple origin rejects *pgproto3.Parse",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("COPY test3 FROM STDIN"),
                ]),
                Step::Receive(&[
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::Parse { name: "", query: "SELECT 2", parameter_oids: &[] },
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "08P01", message: "unexpected message type 0x50 during COPY from stdin", where_: "COPY test3, line 2", ..F }),
                    Receive::Error(Fields { severity: "FATAL", severity_unlocalized: "FATAL", code: "08P01", message: "terminating connection because protocol synchronization was lost", ..F }),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
    ]);
}

#[test]
fn test_copy_unexpected_message_fatal_3() {
    run_wire_tests(&[
        WireTest {
            name: "extended origin rejects *pgproto3.Query",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "", query: "COPY test3 FROM STDIN", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::Query("SELECT 2"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "08P01", message: "unexpected message type 0x51 during COPY from stdin", where_: "COPY test3, line 2", ..F }),
                    Receive::Error(Fields { severity: "FATAL", severity_unlocalized: "FATAL", code: "08P01", message: "terminating connection because protocol synchronization was lost", ..F }),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
    ]);
}

#[test]
fn test_copy_unexpected_message_fatal_4() {
    run_wire_tests(&[
        WireTest {
            name: "extended origin rejects *pgproto3.Parse",
            set_up_script: &[
                "CREATE TABLE test3 (c int);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "", query: "COPY test3 FROM STDIN", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CopyInResponse(0, &[0]),
                ]),
                Step::Send(&[
                    Send::CopyData(b"1\n"),
                    Send::Parse { name: "", query: "SELECT 2", parameter_oids: &[] },
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "08P01", message: "unexpected message type 0x50 during COPY from stdin", where_: "COPY test3, line 2", ..F }),
                    Receive::Error(Fields { severity: "FATAL", severity_unlocalized: "FATAL", code: "08P01", message: "terminating connection because protocol synchronization was lost", ..F }),
                ]),
                Step::OtherQuery { query: "SELECT * FROM test3;", rows: &[] },
            ],
            ..W
        },
    ]);
}

#[test]
fn test_discard_all_clears_protocol_prepared_statements() {
    run_wire_tests(&[
        WireTest {
            name: "DISCARD ALL removes named protocol prepared statements",
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "saved", query: "SELECT 1", parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("DISCARD ALL"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("DISCARD ALL"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "saved", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "26000", message: r#"prepared statement "saved" does not exist"#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 2"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("2")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_extended_protocol_transitions() {
    run_wire_tests(&[
        WireTest {
            name: "Bind without Parse discards later messages until Sync",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Bind { portal: "", statement: "missing", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Parse { name: "skipped", query: "INSERT INTO mytable VALUES (1)", parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "26000", message: r#"prepared statement "missing" does not exist"#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 1"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "Describe unknown statement recovers at Sync",
            steps: &[
                Step::Send(&[
                    Send::Describe(b'S', "missing"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "26000", message: r#"prepared statement "missing" does not exist"#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "select", query: "SELECT 2", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "select", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("2")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "Execute unknown portal recovers at Sync",
            steps: &[
                Step::Send(&[
                    Send::Execute("missing", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "34000", message: r#"portal "missing" does not exist"#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 3"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("3")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "named statement supports Describe before Bind and reuse after Sync",
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "saved", query: "SELECT 4", parameter_oids: &[] },
                    Send::Describe(b'S', "saved"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "saved", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("4")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "named statements and portals support phase-grouped pipelining",
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "first_statement", query: "SELECT 11", parameter_oids: &[] },
                    Send::Parse { name: "second_statement", query: "SELECT 22", parameter_oids: &[] },
                    Send::Describe(b'S', "first_statement"),
                    Send::Describe(b'S', "second_statement"),
                    Send::Bind { portal: "first_portal", statement: "first_statement", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Bind { portal: "second_portal", statement: "second_statement", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("first_portal", 0),
                    Send::Execute("second_portal", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::BindComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("11")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::DataRow(&[Datum::Text("22")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "duplicate named statement discards later messages and preserves original",
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "saved", query: "SELECT 5", parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "saved", query: "SELECT 6", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "saved", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P05", message: r#"prepared statement "saved" already exists"#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "saved", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("5")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "duplicate named portal discards later messages until Sync",
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "saved", query: "SELECT 6", parameter_oids: &[] },
                    Send::Bind { portal: "portal", statement: "saved", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Bind { portal: "portal", statement: "saved", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("portal", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P03", message: r#"cursor "portal" already exists"#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 6"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("6")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "invalid Describe subtype discards later messages until Sync",
            steps: &[
                Step::Send(&[
                    Send::Describe(b'X', "bad"),
                    Send::Parse { name: "skipped", query: "SELECT 9", parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "08P01", message: "invalid DESCRIBE message subtype 88", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 9"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("9")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "invalid Close subtype discards later messages until Sync",
            steps: &[
                Step::Send(&[
                    Send::Close(b'X', "bad"),
                    Send::Parse { name: "skipped", query: "SELECT 10", parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "08P01", message: "invalid CLOSE message subtype 88", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 10"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("10")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "Flush delivers responses without ending the extended batch",
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "saved", query: "SELECT 7", parameter_oids: &[] },
                    Send::Flush,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                ]),
                Step::Send(&[
                    Send::Describe(b'S', "saved"),
                    Send::Bind { portal: "", statement: "saved", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("7")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "Sync is accepted while ready",
            steps: &[
                Step::Send(&[
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 8"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("8")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_implicit_transactions_extended_protocol() {
    run_wire_tests(&[
        WireTest {
            name: "failed DO statement rolls back at Sync",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "do", query: "DO $$ BEGIN INSERT INTO mytable VALUES (1); RAISE EXCEPTION 'forced failure'; END $$", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "do", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "P0001", message: "forced failure", where_: "PL/pgSQL function inline_code_block line 1 at RAISE", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "statements in a batch commit as a single implicit transaction at Sync",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "ins1", query: "INSERT INTO mytable VALUES (1)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins1", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "ins2", query: "INSERT INTO mytable VALUES (2)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins2", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "sel", query: "SELECT * FROM mytable ORDER BY i", parameter_oids: &[] },
                    Send::Describe(b'S', "sel"),
                    Send::Bind { portal: "", statement: "sel", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "i", attnum: 1, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::DataRow(&[Datum::Text("2")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], &[T("2")], ] },
            ],
            ..W
        },
        WireTest {
            name: "work is not visible to other connections until Sync commits it",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "ins1", query: "INSERT INTO mytable VALUES (1)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins1", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Flush,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
                Step::Send(&[
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "error rolls back the batch at Sync and later messages are skipped",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "ins1", query: "INSERT INTO mytable VALUES (1)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins1", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "boom", query: "SELECT 1/0", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "boom", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "ins2", query: "INSERT INTO mytable VALUES (2)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins2", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ParseComplete,
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
                Step::Send(&[
                    Send::Parse { name: "ins3", query: "INSERT INTO mytable VALUES (3)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins3", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("3")], ] },
            ],
            ..W
        },
        WireTest {
            name: "parse error rolls back earlier statements in the batch",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "ins1", query: "INSERT INTO mytable VALUES (1)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins1", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "bad", query: "SELCT 1", parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42601", message: r#"syntax error at or near "SELCT""#, position: 1, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "Sync does not close an explicit transaction block",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "begin", query: "BEGIN", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "begin", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("BEGIN"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "ins1", query: "INSERT INTO mytable VALUES (1)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins1", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
                Step::Send(&[
                    Send::Parse { name: "commit", query: "COMMIT", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "commit", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Flush,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("COMMIT"),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
                Step::Send(&[
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "error inside an explicit transaction block leaves a failed transaction at Sync",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "begin", query: "BEGIN", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "begin", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "ins1", query: "INSERT INTO mytable VALUES (1)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins1", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "boom", query: "SELECT 1/0", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "boom", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("BEGIN"),
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ParseComplete,
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "sel", query: "SELECT 1", parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "25P02", message: "current transaction is aborted, commands ignored until end of transaction block", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "rb", query: "ROLLBACK", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "rb", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "error on the final Execute of a batch rolls back the whole batch",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "ins1", query: "INSERT INTO mytable VALUES (1)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins1", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "ins2", query: "INSERT INTO mytable VALUES (2)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins2", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "boom", query: "SELECT 1/0", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "boom", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ParseComplete,
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "explicit COMMIT as the final statement of a batch commits before Sync",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "ins1", query: "INSERT INTO mytable VALUES (1)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins1", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "commit", query: "COMMIT", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "commit", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Flush,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::Notice(Fields { severity: "WARNING", severity_unlocalized: "WARNING", code: "25P01", message: "there is no transaction in progress", ..F }),
                    Receive::CommandComplete("COMMIT"),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
                Step::Send(&[
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "explicit ROLLBACK as the final statement of a batch discards it",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "ins1", query: "INSERT INTO mytable VALUES (1)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ins1", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Parse { name: "rb", query: "ROLLBACK", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "rb", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::Notice(Fields { severity: "WARNING", severity_unlocalized: "WARNING", code: "25P01", message: "there is no transaction in progress", ..F }),
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "a batch handled entirely outside the engine does not disable autocommit for later statements",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "da", query: "DEALLOCATE ALL", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "da", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("DEALLOCATE ALL"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "DDL committed by the engine mid-batch does not disable autocommit for later statements",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "ct", query: "CREATE TABLE other (j BIGINT)", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "ct", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
    ]);
}

#[test]
fn test_implicit_transactions_simple_protocol() {
    run_wire_tests(&[
        WireTest {
            name: "successful DO statement commits its writes",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("DO $$ BEGIN INSERT INTO mytable VALUES (1); END $$;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("DO"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "multiple statements commit as a single implicit transaction",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); INSERT INTO mytable VALUES (2); SELECT * FROM mytable ORDER BY i;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::RowDescription(&[Field { name: "i", attnum: 1, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::DataRow(&[Datum::Text("2")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], &[T("2")], ] },
            ],
            ..W
        },
        WireTest {
            name: "error rolls back the entire implicit transaction",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); SELECT 1/0; INSERT INTO mytable VALUES (2);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (3);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("3")], ] },
            ],
            ..W
        },
        WireTest {
            name: "failed DO statement rolls back its writes",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("DO $$ BEGIN INSERT INTO mytable VALUES (1); RAISE EXCEPTION 'forced failure'; END $$;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "P0001", message: "forced failure", where_: "PL/pgSQL function inline_code_block line 1 at RAISE", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (2);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "failed dynamic SQL in DO rolls back its writes",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("DO $$ BEGIN INSERT INTO mytable VALUES (1); EXECUTE 'SELECT * FROM missing_dynamic_relation'; END $$;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "missing_dynamic_relation" does not exist"#, internal_query: "SELECT * FROM missing_dynamic_relation", where_: "PL/pgSQL function inline_code_block line 1 at EXECUTE", internal_position: 15, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (2);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "DO statement participates in an explicit transaction",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("DO $$ BEGIN INSERT INTO mytable VALUES (1); END $$;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("DO"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
                Step::Send(&[
                    Send::Query("ROLLBACK;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "failed DO statement aborts an explicit transaction",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("DO $$ BEGIN INSERT INTO mytable VALUES (1); RAISE EXCEPTION 'forced failure'; END $$;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "P0001", message: "forced failure", where_: "PL/pgSQL function inline_code_block line 1 at RAISE", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 1;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "25P02", message: "current transaction is aborted, commands ignored until end of transaction block", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Query("ROLLBACK;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "DO statement participates in a multi-statement implicit transaction",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("DO $$ BEGIN INSERT INTO mytable VALUES (1); END $$; SELECT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("DO"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "explicit COMMIT inside the message commits preceding statements only",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN; INSERT INTO mytable VALUES (1); COMMIT; INSERT INTO mytable VALUES (2); SELECT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("COMMIT"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "COMMIT closes an implicit transaction block and starts a new one",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); COMMIT; INSERT INTO mytable VALUES (2); SELECT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Notice(Fields { severity: "WARNING", severity_unlocalized: "WARNING", code: "25P01", message: "there is no transaction in progress", ..F }),
                    Receive::CommandComplete("COMMIT"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "ROLLBACK closes an implicit transaction block and discards its work",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); ROLLBACK; INSERT INTO mytable VALUES (2);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Notice(Fields { severity: "WARNING", severity_unlocalized: "WARNING", code: "25P01", message: "there is no transaction in progress", ..F }),
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("2")], ] },
            ],
            ..W
        },
        WireTest {
            name: "BEGIN converts an implicit transaction block into an explicit one",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); BEGIN; INSERT INTO mytable VALUES (2);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("BEGIN"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
                Step::Send(&[
                    Send::Query("ROLLBACK;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "transaction block left open by a Query message continues across messages",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN; INSERT INTO mytable VALUES (1);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT * FROM mytable ORDER BY i;"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "i", attnum: 1, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (2);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
                Step::Send(&[
                    Send::Query("COMMIT;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COMMIT"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], &[T("2")], ] },
            ],
            ..W
        },
        WireTest {
            name: "COMMIT of a block opened in an earlier message starts an implicit block for remaining statements",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); COMMIT; INSERT INTO mytable VALUES (2); SELECT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("COMMIT"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "error in an explicit transaction block leaves the session in a failed transaction",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN; SELECT 1/0; ROLLBACK;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 1;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "25P02", message: "current transaction is aborted, commands ignored until end of transaction block", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Query("ROLLBACK;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "syntax error anywhere in the message prevents any statement from executing",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN; INSERT INTO mytable VALUES (1); COMMIT; INSERT INTO mytable VALUES (2); SELCT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42601", message: r#"syntax error at or near "SELCT""#, position: 80, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "ROLLBACK TO SAVEPOINT recovers a failed transaction block",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN; INSERT INTO mytable VALUES (1); SAVEPOINT sp1;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("SAVEPOINT"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Query("ROLLBACK TO sp1;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT * FROM mytable ORDER BY i;"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "i", attnum: 1, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("COMMIT;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COMMIT"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "savepoints are not allowed in an implicit transaction block",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); SAVEPOINT sp1;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "25P01", message: "SAVEPOINT can only be used in transaction blocks", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "error on the final statement rolls back the entire implicit transaction",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); INSERT INTO mytable VALUES (2); SELECT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "error part way through the final statement's results rolls back the entire implicit transaction",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); INSERT INTO mytable VALUES (2); SELECT i / (i - 2) FROM mytable ORDER BY i;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("-1")]),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "explicit COMMIT as the final statement commits the implicit transaction",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); INSERT INTO mytable VALUES (2); COMMIT;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Notice(Fields { severity: "WARNING", severity_unlocalized: "WARNING", code: "25P01", message: "there is no transaction in progress", ..F }),
                    Receive::CommandComplete("COMMIT"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], &[T("2")], ] },
            ],
            ..W
        },
        WireTest {
            name: "explicit ROLLBACK as the final statement discards the implicit transaction",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); INSERT INTO mytable VALUES (2); ROLLBACK;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Notice(Fields { severity: "WARNING", severity_unlocalized: "WARNING", code: "25P01", message: "there is no transaction in progress", ..F }),
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
        WireTest {
            name: "error on the final statement of an explicit transaction block preserves the block",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("BEGIN; INSERT INTO mytable VALUES (1); SAVEPOINT sp1; SELECT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("SAVEPOINT"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Query("ROLLBACK TO sp1;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT * FROM mytable ORDER BY i;"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "i", attnum: 1, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("COMMIT;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COMMIT"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "statement handled outside the engine as the final statement still commits the block",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1); DEALLOCATE ALL;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("DEALLOCATE ALL"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
            ],
            ..W
        },
        WireTest {
            name: "errors do not poison the session outside of explicit transaction blocks",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("SELCT 1;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42601", message: r#"syntax error at or near "SELCT""#, position: 1, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
                Step::Send(&[
                    Send::Query("SELECT * FROM doesnotexist;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "doesnotexist" does not exist"#, position: 15, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (2);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], &[T("2")], ] },
                Step::Send(&[
                    Send::Query("SELECT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1);"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "23505", message: r#"duplicate key value violates unique constraint "mytable_pkey""#, detail: "Key (i)=(1) already exists.", schema: "public", table: "mytable", constraint: "mytable_pkey", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELCT 1;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42601", message: r#"syntax error at or near "SELCT""#, position: 1, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (3);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], &[T("2")], &[T("3")], ] },
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (4); SELECT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT * FROM mytable ORDER BY i;"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "i", attnum: 1, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::DataRow(&[Datum::Text("2")]),
                    Receive::DataRow(&[Datum::Text("3")]),
                    Receive::CommandComplete("SELECT 3"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "errors in a transaction block do not poison the session after the block is ended",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("begin;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT * FROM doesnotexist;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "doesnotexist" does not exist"#, position: 15, ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 1;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "25P02", message: "current transaction is aborted, commands ignored until end of transaction block", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Query("rollback;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
                Step::Send(&[
                    Send::Query("begin work;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT 1/0;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "22012", message: "division by zero", ..F }),
                    Receive::ReadyForQuery(b'E'),
                ]),
                Step::Send(&[
                    Send::Query("commit;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (2);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], &[T("2")], ] },
            ],
            ..W
        },
        WireTest {
            name: "single-statement Query messages commit immediately and are visible to other connections",
            set_up_script: &[
                "CREATE TABLE mytable (i BIGINT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("INSERT INTO mytable VALUES (1);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("1")], ] },
                Step::Send(&[
                    Send::Query("UPDATE mytable SET i = 2;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("UPDATE 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT * FROM mytable ORDER BY i;", rows: &[&[T("2")], ] },
                Step::Send(&[
                    Send::Query("DELETE FROM mytable;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("DELETE 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::OtherQuery { query: "SELECT count(*) FROM mytable;", rows: &[&[T("0")], ] },
            ],
            ..W
        },
    ]);
}

#[test]
fn test_issue3116_wire_format() {
    run_wire_tests(&[
        WireTest {
            name: "Issue #3116: dolt system table booleans over the wire",
            set_up_script: &[
                "CREATE TABLE t3116 (id INT PRIMARY KEY);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("SELECT dirty FROM dolt.branches;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "dolt.branches" does not exist"#, position: 19, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT dirty FROM dolt.branches WHERE name = 'main';"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "dolt.branches" does not exist"#, position: 19, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT dirty FROM dolt.branches ORDER BY name;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "dolt.branches" does not exist"#, position: 19, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT dirty FROM dolt.branches WHERE dirty = true;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "dolt.branches" does not exist"#, position: 19, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT staged FROM dolt.status ORDER BY table_name;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "dolt.status" does not exist"#, position: 20, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT data_change, schema_change FROM dolt.diff ORDER BY table_name;"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "dolt.diff" does not exist"#, position: 40, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}
