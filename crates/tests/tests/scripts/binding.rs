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
fn test_binding_explicit_cast_to_timestamptz_preserves_offset() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingExplicitCastToTimestamptzPreservesOffset",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE t9 (id INT PRIMARY KEY, ts timestamptz);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "INSERT INTO t9 VALUES (1, $1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TIMESTAMPTZ]),
                    Receive::NoData,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 2, 252, 135, 244, 32, 220, 0])], result_formats: &[] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT $1::timestamptz", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TIMESTAMPTZ]),
                    Receive::RowDescription(&[Field { name: "timestamptz", attnum: 0, type_oid: TIMESTAMPTZ, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("2026-08-21 12:00:00+05:00")], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 2, 252, 135, 244, 32, 220, 0])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT count(*) FROM t9 WHERE ts = $1::timestamptz", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TIMESTAMPTZ]),
                    Receive::RowDescription(&[Field { name: "count", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("2026-08-21 12:00:00+05:00")], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\0\0\0\0\u{1}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_float4_to_double_with_untyped_null() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingFloat4ToDoubleWithUntypedNull",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE t4 (a DOUBLE PRECISION, f BOOLEAN);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "INSERT INTO t4 (a, f) VALUES ($1, $2);", parameter_oids: &[FLOAT4, 0] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[1, 1], parameters: &[Datum::Text("@`\0\0"), Datum::Null], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::NoData,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT a, f FROM t4", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "a", attnum: 1, type_oid: FLOAT8, size: 8, typmod: -1, format: 0 }, Field { name: "f", attnum: 2, type_oid: BOOL, size: 1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1, 1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("@\u{c}\0\0\0\0\0\0"), Datum::Null]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_insert_select_with_oid_zero() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingInsertSelectWithOidZero",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE t (id INT PRIMARY KEY, v BIGINT, label TEXT);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "insert_select", query: "INSERT INTO t (id, v, label) SELECT $1, $2, $3", parameter_oids: &[] },
                    Send::Describe(b'S', "insert_select"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[INT4, INT8, TEXT]),
                    Receive::NoData,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "INSERT INTO t (id, v, label) SELECT $1, $2, $3", parameter_oids: &[0, 0, 0] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0, 0, 0], parameters: &[Datum::Text("12"), Datum::Text("9223372036854775806"), Datum::Text("inserted")], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::NoData,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT id, v, label FROM t", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "id", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "v", attnum: 2, type_oid: INT8, size: 8, typmod: -1, format: 0 }, Field { name: "label", attnum: 3, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1, 1, 0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{c}"), Datum::Bytes(&[127, 255, 255, 255, 255, 255, 255, 254]), Datum::Text("inserted")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_int4_to_bigint_with_untyped_null() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingInt4ToBigintWithUntypedNull",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE t1 (a BIGINT, b INT, f BOOLEAN);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "INSERT INTO t1 (a, b, f) VALUES ($1, $2, $3);", parameter_oids: &[INT4, INT4, 0] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[1, 1, 1], parameters: &[Datum::Text("\0\0\0\u{1}"), Datum::Text("\0\0\0\u{2}"), Datum::Null], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::NoData,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT a, b, f FROM t1", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "a", attnum: 1, type_oid: INT8, size: 8, typmod: -1, format: 0 }, Field { name: "b", attnum: 2, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "f", attnum: 3, type_oid: BOOL, size: 1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1, 1, 1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\0\0\0\0\u{1}"), Datum::Text("\0\0\0\u{2}"), Datum::Null]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_int8_to_int_column_with_untyped_null() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingInt8ToIntColumnWithUntypedNull",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE t5 (a INT, f BOOLEAN);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "INSERT INTO t5 (a, f) VALUES ($1, $2);", parameter_oids: &[INT8, 0] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[1, 1], parameters: &[Datum::Text("\0\0\0\0\0\0\0\u{7}"), Datum::Null], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::NoData,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT a, f FROM t5", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "a", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "f", attnum: 2, type_oid: BOOL, size: 1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1, 1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{7}"), Datum::Null]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_jsonb_extract_path_text_with_untyped_param() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingJSONBExtractPathTextWithUntypedParam",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE t8 (id TEXT PRIMARY KEY, props JSONB);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query(r#"INSERT INTO t8 VALUES ('a', '{"name":"Alice"}');"#),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT count(*) FROM t8 WHERE props #>> array['name']::text[] = $1", parameter_oids: &[0] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("Alice")], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::RowDescription(&[Field { name: "count", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_multiple_inferred_and_explicit_oids_interleaved() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingMultipleInferredAndExplicitOIDsInterleaved",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE t7 (a INT, b BIGINT, c INT, d DOUBLE PRECISION, e BOOLEAN);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "INSERT INTO t7 (a, b, c, d, e) VALUES ($1, $2, $3, $4, $5);", parameter_oids: &[0, INT4, 0, FLOAT4, 0] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0, 1, 0, 1, 1], parameters: &[Datum::Text("11"), Datum::Text("\0\0\0\u{16}"), Datum::Text("33"), Datum::Bytes(&[64, 144, 0, 0]), Datum::Null], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::NoData,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT a, b, c, d, e FROM t7", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "a", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "b", attnum: 2, type_oid: INT8, size: 8, typmod: -1, format: 0 }, Field { name: "c", attnum: 3, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "d", attnum: 4, type_oid: FLOAT8, size: 8, typmod: -1, format: 0 }, Field { name: "e", attnum: 5, type_oid: BOOL, size: 1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1, 1, 1, 1, 1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{b}"), Datum::Text("\0\0\0\0\0\0\0\u{16}"), Datum::Text("\0\0\0!"), Datum::Text("@\u{12}\0\0\0\0\0\0"), Datum::Null]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_on_conflict_do_update_with_untyped_null() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingOnConflictDoUpdateWithUntypedNull",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE tconf (id INT PRIMARY KEY, a BIGINT, f BOOLEAN);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO tconf VALUES (1, 0, true);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "INSERT INTO tconf (id, a) VALUES (1, 0) ON CONFLICT (id) DO UPDATE SET a = $1, f = $2;", parameter_oids: &[INT4, 0] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[1, 1], parameters: &[Datum::Text("\0\0\0*"), Datum::Null], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::NoData,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT a, f FROM tconf WHERE id = 1", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "a", attnum: 2, type_oid: INT8, size: 8, typmod: -1, format: 0 }, Field { name: "f", attnum: 3, type_oid: BOOL, size: 1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1, 1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\0\0\0\0*"), Datum::Null]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_short_parameter_oids_array() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingShortParameterOIDsArray",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE t6 (a INT, b INT);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "INSERT INTO t6 (a, b) VALUES ($1, $2);", parameter_oids: &[INT4] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0, 0], parameters: &[Datum::Text("1"), Datum::Text("2")], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::NoData,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT a, b FROM t6", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "a", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "b", attnum: 2, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1, 1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1}"), Datum::Text("\0\0\0\u{2}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_update_int4_to_bigint_with_untyped_null() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingUpdateInt4ToBigintWithUntypedNull",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE t3 (id INT PRIMARY KEY, a BIGINT, f BOOLEAN);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO t3 VALUES (1, 0, true);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "UPDATE t3 SET a = $1, f = $2 WHERE id = 1;", parameter_oids: &[INT4, 0] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[1, 1], parameters: &[Datum::Text("\0\0\0*"), Datum::Null], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::NoData,
                    Receive::CommandComplete("UPDATE 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT a, f FROM t3 WHERE id = 1", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "a", attnum: 2, type_oid: INT8, size: 8, typmod: -1, format: 0 }, Field { name: "f", attnum: 3, type_oid: BOOL, size: 1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1, 1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\0\0\0\0*"), Datum::Null]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_where_clause_int4_to_bigint_with_untyped_null() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingWhereClauseInt4ToBigintWithUntypedNull",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE t2 (a BIGINT, f BOOLEAN);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO t2 VALUES (1, true), (2, false);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 2"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT a, f FROM t2 WHERE a = $1 AND (f = $2 OR f IS NOT NULL);", parameter_oids: &[INT4, 0] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[1, 1], parameters: &[Datum::Text("\0\0\0\u{1}"), Datum::Null], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::RowDescription(&[Field { name: "a", attnum: 1, type_oid: INT8, size: 8, typmod: -1, format: 0 }, Field { name: "f", attnum: 2, type_oid: BOOL, size: 1, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1"), Datum::Text("t")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_with_oid_zero() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingWithOidZero",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE my_table (id INT, name varchar(100));"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "INSERT INTO my_table (id, name) VALUES ($1, $2);", parameter_oids: &[0, TEXT] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0, 0], parameters: &[Datum::Text("42"), Datum::Text("Alice")], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::NoData,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT id, name FROM my_table", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "id", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "name", attnum: 2, type_oid: VARCHAR, size: -1, typmod: 104, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1, 0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0*"), Datum::Text("Alice")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_binding_with_text_array() {
    run_wire_tests(&[
        WireTest {
            name: "TestBindingWithTextArray",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT $1::text[]", parameter_oids: &[TEXT_ARRAY] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[1], parameters: &[Datum::Text("\0\0\0\u{1}\0\0\0\0\0\0\0\u{19}\0\0\0\u{2}\0\0\0\u{1}\0\0\0\u{3}foo\0\0\0\u{3}bar")], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::RowDescription(&[Field { name: "text", attnum: 0, type_oid: TEXT_ARRAY, size: -1, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("{foo,bar}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_issue2386() {
    run_wire_tests(&[
        WireTest {
            name: "TestIssue2386",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("CREATE TABLE users (id INT PRIMARY KEY, name TEXT NOT NULL);"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO users VALUES (1, 'alice'), (2, 'bob'), (3, 'carol'), (4, 'dave');"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 4"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT id, name FROM users WHERE id = ANY($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[INT4_ARRAY]),
                    Receive::RowDescription(&[Field { name: "id", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "name", attnum: 2, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[1], parameters: &[Datum::Text("\0\0\0\u{1}\0\0\0\0\0\0\0\u{17}\0\0\0\u{2}\0\0\0\u{1}\0\0\0\u{4}\0\0\0\u{1}\0\0\0\u{4}\0\0\0\u{3}")], result_formats: &[1, 0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1}"), Datum::Text("alice")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{3}"), Datum::Text("carol")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}
